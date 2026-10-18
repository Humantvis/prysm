package client

import (
	"context"
	"fmt"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/signing"
	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v7/crypto/rand"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	validatorpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1/validator-client"
	"github.com/OffchainLabs/prysm/v7/time/slots"
	"github.com/pkg/errors"
)

// DefaultRandaoHashChainLength is the chain length used when creating a new
// randao hash chain. 2^16 is the EIP-recommended minimum.
const DefaultRandaoHashChainLength = 1 << 16

var (
	ErrHashChainExhausted = errors.New("randao hash chain exhausted")
	ErrCommitmentMismatch = errors.New("on-chain commitment not found in local randao hash chain")
)

// nextHashChainReveal returns the preimage of the validator's current on-chain
// commitment c_k. See the original doc comment — unchanged.
func nextHashChainReveal(seed [32]byte, length uint64, onChain [32]byte) ([32]byte, error) {
	chain, err := eip8321.ComputeHashChain(seed, length)
	if err != nil {
		return [32]byte{}, err
	}
	for i := len(chain) - 1; i >= 1; i-- {
		if chain[i] == onChain {
			return chain[i-1], nil
		}
	}
	if chain[0] == onChain {
		return [32]byte{}, ErrHashChainExhausted
	}
	return [32]byte{}, ErrCommitmentMismatch
}

func (v *validator) randaoChainLength() uint64 {
	return DefaultRandaoHashChainLength
}

// ensureRandaoChain loads stored {seed, length}, or creates + persists a fresh
// chain on first call. Idempotent.
func (v *validator) ensureRandaoChain(ctx context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) (seed [32]byte, length uint64, err error) {
	seed, length, ok, err := v.db.RandaoHashChain(ctx, pubKey)
	if err != nil {
		return seed, length, err
	}
	if ok {
		return seed, length, nil
	}

	var newSeed [32]byte
	copy(newSeed[:], rand.Bytes(32))
	if newSeed == ([32]byte{}) {
		return newSeed, 0, errors.New("generated zero randao hash chain seed")
	}
	newLength := v.randaoChainLength()
	if _, err := eip8321.ComputeHashChain(newSeed, newLength); err != nil {
		return newSeed, newLength, fmt.Errorf("invalid randao hash chain params: %w", err)
	}
	if err := v.db.SaveRandaoHashChain(ctx, pubKey, newSeed, newLength); err != nil {
		return newSeed, newLength, err
	}
	return newSeed, newLength, nil
}

// onChainRandaoCommitment reads the validator's current on-chain commitment.
// Zero return = unregistered.
func (v *validator) onChainRandaoCommitment(ctx context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) ([32]byte, error) {
	resp, err := v.validatorClient.RandaoCommitment(ctx, &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]})
	if err != nil {
		return [32]byte{}, err
	}
	if resp == nil || len(resp.Commitment) != 32 {
		return [32]byte{}, errors.New("invalid randao commitment response")
	}
	var c [32]byte
	copy(c[:], resp.Commitment)
	return c, nil
}

// nextHashChainRevealFor is called from ProposeBlock. Returns (reveal,
// registered, error). Fails closed on any error.
func (v *validator) nextHashChainRevealFor(ctx context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) ([32]byte, bool, error) {
	onChain, err := v.onChainRandaoCommitment(ctx, pubKey)
	if err != nil {
		return [32]byte{}, false, err
	}
	if onChain == ([32]byte{}) {
		return [32]byte{}, false, nil
	}

	seed, length, ok, err := v.db.RandaoHashChain(ctx, pubKey)
	if err != nil {
		return [32]byte{}, false, err
	}
	if !ok {
		return [32]byte{}, false, ErrCommitmentMismatch
	}

	reveal, err := nextHashChainReveal(seed, length, onChain)
	if err != nil {
		return [32]byte{}, false, err
	}
	return reveal, true, nil
}

// RegisterRandaoCommitment signs and submits a RandaoCommitmentRegistration.
// Requires validatorpb.SignRequest_RandaoCommitment (from keymanager.proto
// regeneration) and validatorClient.SubmitRandaoCommitmentRegistration.
func (v *validator) RegisterRandaoCommitment(
	ctx context.Context,
	pubKey [fieldparams.BLSPubkeyLength]byte,
	epoch primitives.Epoch,
) error {
	seed, length, err := v.ensureRandaoChain(ctx, pubKey)
	if err != nil {
		return errors.Wrap(err, "ensure randao chain")
	}

	chain, err := eip8321.ComputeHashChain(seed, length)
	if err != nil {
		return errors.Wrap(err, "compute hash chain")
	}
	commitment := chain[length]
	if commitment == ([32]byte{}) {
		return errors.New("computed zero commitment (impossible)")
	}

	idxResp, err := v.validatorClient.ValidatorIndex(ctx, &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]})
	if err != nil {
		return errors.Wrap(err, "validator index lookup")
	}

	reg := &ethpb.RandaoCommitmentRegistration{
		ValidatorIndex: idxResp.Index,
		Commitment:     commitment[:],
	}

	// EIP-8321: the registration domain is genesis-fork-anchored (like
	// BLS-to-execution-change), so signatures stay valid across forks.
	// Compute it locally — the DomainData RPC returns a per-epoch fork
	// version, which would produce the wrong domain.
	gvr, err := v.db.GenesisValidatorsRoot(ctx)
	if err != nil {
		return errors.Wrap(err, "genesis validators root")
	}
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainRandaoCommitmentRegistration,
		params.BeaconConfig().GenesisForkVersion,
		gvr,
	)
	if err != nil {
		return errors.Wrap(err, "compute domain")
	}

	root, err := signing.ComputeSigningRoot(reg, domain)
	if err != nil {
		return errors.Wrap(err, signingRootErr)
	}

	signingSlot, err := slots.EpochStart(epoch)
	if err != nil {
		return errors.Wrap(err, "epoch start")
	}
	sig, err := v.km.Sign(ctx, &validatorpb.SignRequest{
		PublicKey:       pubKey[:],
		SigningRoot:     root[:],
		SignatureDomain: domain,
		Object:          &validatorpb.SignRequest_RandaoCommitment{RandaoCommitment: reg},
		SigningSlot:     signingSlot,
	})
	if err != nil {
		return errors.Wrap(err, "sign randao commitment registration")
	}

	signed := &ethpb.SignedRandaoCommitmentRegistration{
		Message:   reg,
		Signature: sig.Marshal(),
	}
	if _, err := v.validatorClient.SubmitRandaoCommitmentRegistration(ctx, signed); err != nil {
		return errors.Wrap(err, "submit randao commitment registration")
	}
	return nil
}