package eip8321

import "errors"

var (
	// ErrRegistrationIndexOutOfRange: validator_index is not in the validator registry.
	ErrRegistrationIndexOutOfRange = errors.New("randao commitment registration: validator index out of range")
	// ErrZeroRandaoCommitment: the registered commitment equals UNSET_RANDAO_COMMITMENT.
	ErrZeroRandaoCommitment = errors.New("randao commitment registration: commitment is zero")
	// ErrAlreadyRegistered: the validator already has an active commitment.
	ErrAlreadyRegistered = errors.New("randao commitment registration: validator already registered")
	// ErrRegistrationAlreadyPending: the validator already has a commitment waiting in the queue.
	ErrRegistrationAlreadyPending = errors.New("randao commitment registration: validator already has a pending registration")
)

// VerifyRandaoCommitmentRegistration runs the non-signature checks of
// process_randao_commitment_registration, in the same order as the spec.
// The caller reads the values from the block and the state.
//
// Spec (beacon-chain.md, EIP-8321):
//
//	assert index < len(state.validators)
//	assert registration.commitment != UNSET_RANDAO_COMMITMENT
//	assert state.randao_commitments[index] == UNSET_RANDAO_COMMITMENT
//	assert all(pending.validator_index != index for pending in state.pending_randao_commitments)
func VerifyRandaoCommitmentRegistration(
	index uint64,
	numValidators uint64,
	commitment [32]byte,
	storedCommitment [32]byte,
	pendingIndices []uint64,
) error {
	if index >= numValidators {
		return ErrRegistrationIndexOutOfRange
	}
	if commitment == ([32]byte{}) {
		return ErrZeroRandaoCommitment
	}
	if storedCommitment != ([32]byte{}) {
		return ErrAlreadyRegistered
	}
	for _, pending := range pendingIndices {
		if pending == index {
			return ErrRegistrationAlreadyPending
		}
	}
	return nil
}

// RegistrationActivationEpoch returns the epoch at which a new registration
// becomes active:
//
//	activation_epoch = get_current_epoch(state) + COMMITMENT_REGISTRATION_DELAY
func RegistrationActivationEpoch(currentEpoch, delay uint64) uint64 {
	return currentEpoch + delay
}
