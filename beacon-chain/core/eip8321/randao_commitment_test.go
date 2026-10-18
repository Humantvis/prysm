package eip8321_test

import (
	"errors"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
)

func TestVerifyRandaoCommitmentRegistration(t *testing.T) {
	const numValidators = 64
	commitment := seed(0x22)
	unset := [32]byte{}

	tests := []struct {
		name           string
		index          uint64
		commitment     [32]byte
		stored         [32]byte
		pendingIndices []uint64
		wantErr        error
	}{
		{"success", 0, commitment, unset, nil, nil},
		{"success_last_validator", numValidators - 1, commitment, unset, nil, nil},
		{"success_exited_validator", 5, commitment, unset, nil, nil},
		{"success_second_validator_with_one_pending", 2, commitment, unset, []uint64{1}, nil},
		{"invalid_validator_index_out_of_range", numValidators, commitment, unset, nil, eip8321.ErrRegistrationIndexOutOfRange},
		{"invalid_zero_commitment", 0, unset, unset, nil, eip8321.ErrZeroRandaoCommitment},
		{"invalid_already_registered", 0, commitment, seed(0x33), nil, eip8321.ErrAlreadyRegistered},
		{"invalid_registration_already_pending", 1, commitment, unset, []uint64{1}, eip8321.ErrRegistrationAlreadyPending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eip8321.VerifyRandaoCommitmentRegistration(tt.index, numValidators, tt.commitment, tt.stored, tt.pendingIndices)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyRandaoCommitmentRegistration_CheckOrder(t *testing.T) {
	err := eip8321.VerifyRandaoCommitmentRegistration(10, 10, [32]byte{}, [32]byte{}, nil)
	if !errors.Is(err, eip8321.ErrRegistrationIndexOutOfRange) {
		t.Fatalf("got %v", err)
	}
	err = eip8321.VerifyRandaoCommitmentRegistration(3, 10, seed(0x22), seed(0x33), []uint64{3})
	if !errors.Is(err, eip8321.ErrAlreadyRegistered) {
		t.Fatalf("got %v", err)
	}
}

func TestRegistrationActivationEpoch(t *testing.T) {
	if got := eip8321.RegistrationActivationEpoch(10, 3); got != 13 {
		t.Fatalf("got %d, want 13", got)
	}
	if got := eip8321.RegistrationActivationEpoch(0, 3); got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
}
