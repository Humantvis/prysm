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

// ErrTooManyRegistrations: a block carries more than MAX_RANDAO_COMMITMENT_REGISTRATIONS registrations.
var ErrTooManyRegistrations = errors.New("too many randao commitment registrations in block")

// ProcessRandaoCommitmentRegistrations processes every registration of a block body,
// in order. It is called from process_operations, after the payload attestations.
//
//	assert len(body.randao_commitment_registrations) <= MAX_RANDAO_COMMITMENT_REGISTRATIONS
//	for_ops(body.randao_commitment_registrations, process_randao_commitment_registration)
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

// ProcessRandaoCommitmentRegistration implements process_randao_commitment_registration.
//
// Spec (beacon-chain.md, EIP-8321):
//
//	registration = signed_registration.message
//	index = registration.validator_index
//	assert index < len(state.validators)
//	assert registration.commitment != UNSET_RANDAO_COMMITMENT
//	assert state.randao_commitments[index] == UNSET_RANDAO_COMMITMENT
//	assert all(pending.validator_index != index for pending in state.pending_randao_commitments)
//	domain = compute_domain(DOMAIN_RANDAO_COMMITMENT_REGISTRATION, genesis_validators_root=state.genesis_validators_root)
//	signing_root = compute_signing_root(registration, domain)
//	assert bls.Verify(state.validators[index].pubkey, signing_root, signed_registration.signature)
//	state.pending_randao_commitments.append(PendingRandaoCommitment(
//	    validator_index=index,
//	    commitment=registration.commitment,
//	    activation_epoch=get_current_epoch(state) + COMMITMENT_REGISTRATION_DELAY,
//	))
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

	// Only read the stored commitment when the index exists; the range check below rejects the rest.
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

	// Fork-agnostic domain: GENESIS_FORK_VERSION, so a registration stays valid across forks.
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
