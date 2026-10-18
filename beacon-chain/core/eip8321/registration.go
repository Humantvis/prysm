package eip8321

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/signing"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/time"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/pkg/errors"
)

var ErrTooManyRegistrations = errors.New("too many randao commitment registrations in block")

func ProcessRandaoCommitmentRegistrations(st state.BeaconState, registrations []*ethpb.SignedRandaoCommitmentRegistration) error {
	if uint64(len(registrations)) > params.BeaconConfig().MaxRandaoCommitmentRegistrations {
		return ErrTooManyRegistrations
	}
	for i, registration := range registrations {
		if err := ProcessRandaoCommitmentRegistration(st, registration); err != nil {
			return errors.Wrapf(err, "could not process randao commitment registration %d", i)
		}
	}
	return nil
}

func ProcessRandaoCommitmentRegistration(st state.BeaconState, signed *ethpb.SignedRandaoCommitmentRegistration) error {
	if signed == nil || signed.Message == nil {
		return errors.New("nil randao commitment registration")
	}
	registration := signed.Message
	if len(registration.Commitment) != 32 {
		return errors.New("randao commitment must be 32 bytes")
	}
	var commitment [32]byte
	copy(commitment[:], registration.Commitment)
	index := registration.ValidatorIndex
	numValidators := uint64(st.NumValidators())

	var stored [32]byte
	if uint64(index) < numValidators {
		var err error
		stored, err = st.RandaoCommitmentAtIndex(index)
		if err != nil {
			return err
		}
	}
	pending, err := st.PendingRandaoCommitments()
	if err != nil {
		return err
	}
	pendingIndices := make([]uint64, len(pending))
	for i, p := range pending {
		pendingIndices[i] = uint64(p.ValidatorIndex)
	}
	if err := VerifyRandaoCommitmentRegistration(uint64(index), numValidators, commitment, stored, pendingIndices); err != nil {
		return err
	}

	cfg := params.BeaconConfig()
	domain, err := signing.ComputeDomain(cfg.DomainRandaoCommitmentRegistration, cfg.GenesisForkVersion, st.GenesisValidatorsRoot())
	if err != nil {
		return err
	}
	pubkey := st.PubkeyAtIndex(index)
	if err := signing.VerifySigningRoot(registration, pubkey[:], signed.Signature, domain); err != nil {
		return errors.Wrap(err, "could not verify randao commitment registration signature")
	}

	activationEpoch := RegistrationActivationEpoch(uint64(time.CurrentEpoch(st)), uint64(cfg.CommitmentRegistrationDelay))
	return st.AppendPendingRandaoCommitment(&ethpb.PendingRandaoCommitment{
		ValidatorIndex:  index,
		Commitment:      commitment[:],
		ActivationEpoch: primitives.Epoch(activationEpoch),
	})
}
