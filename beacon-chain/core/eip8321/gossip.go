package eip8321

import (
	"bytes"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v7/crypto/bls"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/pkg/errors"
)

// GossipResult is the outcome of gossip validation, independent of libp2p.
type GossipResult int

const (
	GossipAccept GossipResult = iota
	GossipIgnore
	GossipReject
)

var (
	ErrRegistrationAlreadySeen = errors.New("already seen RANDAO commitment registration for this validator")
	ErrHeadStatePreEip8321     = errors.New("head state is pre-eip8321")
)

// ValidateRandaoCommitmentRegistrationGossip runs the checks of
// validate_randao_commitment_registration_gossip (p2p-interface.md) except
// the BLS check: on accept it returns the signature batch, which the caller
// verifies. seen reports whether a registration for the index was already
// accepted; it may be nil.
func ValidateRandaoCommitmentRegistrationGossip(
	st state.ReadOnlyBeaconState,
	seen func(primitives.ValidatorIndex) bool,
	signed *ethpb.SignedRandaoCommitmentRegistration,
) (GossipResult, *bls.SignatureBatch, error) {
	if signed == nil || signed.Message == nil {
		return GossipReject, nil, errors.New("nil registration")
	}
	reg := signed.Message
	if len(reg.Commitment) != 32 {
		return GossipReject, nil, errors.New("randao commitment must be 32 bytes")
	}

	// [IGNORE] First valid registration received for the validator.
	if seen != nil && seen(reg.ValidatorIndex) {
		return GossipIgnore, nil, ErrRegistrationAlreadySeen
	}

	// [IGNORE] The head state has upgraded to EIP-8321. The spec compares the
	// fork version, so check it as well as the state type.
	if st.Version() < version.Eip8321 ||
		bytes.Compare(st.Fork().CurrentVersion, params.BeaconConfig().Eip8321ForkVersion) < 0 {
		return GossipIgnore, nil, ErrHeadStatePreEip8321
	}

	var commitment [32]byte
	copy(commitment[:], reg.Commitment)

	numValidators := uint64(st.NumValidators())
	var stored [32]byte
	if uint64(reg.ValidatorIndex) < numValidators {
		var err error
		stored, err = st.RandaoCommitmentAtIndex(reg.ValidatorIndex)
		if err != nil {
			return GossipIgnore, nil, err
		}
	}
	pendings, err := st.PendingRandaoCommitments()
	if err != nil {
		return GossipIgnore, nil, err
	}
	pendingIdx := make([]uint64, len(pendings))
	for i, p := range pendings {
		pendingIdx[i] = uint64(p.ValidatorIndex)
	}

	// Index in range and non-zero commitment are REJECT;
	// already registered and already pending are IGNORE.
	if err := VerifyRandaoCommitmentRegistration(uint64(reg.ValidatorIndex), numValidators, commitment, stored, pendingIdx); err != nil {
		if errors.Is(err, ErrAlreadyRegistered) || errors.Is(err, ErrRegistrationAlreadyPending) {
			return GossipIgnore, nil, err
		}
		return GossipReject, nil, err
	}

	// [REJECT] The signature is valid: verified by the caller.
	batch, err := RandaoCommitmentRegistrationSignatureBatch(st, []*ethpb.SignedRandaoCommitmentRegistration{signed})
	if err != nil {
		return GossipReject, nil, err
	}
	return GossipAccept, batch, nil
}
