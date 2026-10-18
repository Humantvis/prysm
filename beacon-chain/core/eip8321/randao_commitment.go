package eip8321

import "errors"

var (
	ErrRegistrationIndexOutOfRange = errors.New("randao commitment registration: validator index out of range")
	ErrZeroRandaoCommitment        = errors.New("randao commitment registration: commitment is zero")
	ErrAlreadyRegistered           = errors.New("randao commitment registration: validator already registered")
	ErrRegistrationAlreadyPending  = errors.New("randao commitment registration: validator already has a pending registration")
)

// Does validator exist, is commitment not zero, not already registered, not already waiting in queue
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

// Registration activates only after a delay
func RegistrationActivationEpoch(currentEpoch, delay uint64) uint64 {
	return currentEpoch + delay
}
