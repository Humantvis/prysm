package eip8321_test

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
)

// Test names follow the official EIP-8321 test vectors in
// reftests/<preset>/eip8321/epoch_processing/pending_randao_commitments/pyspec_tests/.

const currentEpoch = 5 // so next epoch = 6

// applyPending simulates what the state code will do: write activated
// commitments into randao_commitments and keep the rest in the queue.
func applyPending(commitments [][32]byte, pending []eip8321.PendingCommitment) []eip8321.PendingCommitment {
	activate, remaining := eip8321.SplitPendingRandaoCommitments(pending, currentEpoch)
	for _, p := range activate {
		commitments[p.ValidatorIndex] = p.Commitment
	}
	return remaining
}

func TestPendingRandaoCommitments_NoPendingCommitments(t *testing.T) {
	commitments := make([][32]byte, 8)
	remaining := applyPending(commitments, nil)
	if len(remaining) != 0 {
		t.Fatalf("queue must stay empty, got %d", len(remaining))
	}
	for i, c := range commitments {
		if c != ([32]byte{}) {
			t.Fatalf("commitment %d changed", i)
		}
	}
}

func TestPendingRandaoCommitments_ActivateAtNextEpoch(t *testing.T) {
	commitments := make([][32]byte, 8)
	pending := []eip8321.PendingCommitment{{ValidatorIndex: 0, Commitment: seed(0x01), ActivationEpoch: currentEpoch + 1}}
	remaining := applyPending(commitments, pending)
	if commitments[0] != seed(0x01) {
		t.Fatal("commitment was not activated")
	}
	if len(remaining) != 0 {
		t.Fatalf("queue must be empty, got %d", len(remaining))
	}
}

func TestPendingRandaoCommitments_NotYetActivated(t *testing.T) {
	commitments := make([][32]byte, 8)
	pending := []eip8321.PendingCommitment{{ValidatorIndex: 0, Commitment: seed(0x01), ActivationEpoch: currentEpoch + 2}}
	remaining := applyPending(commitments, pending)
	if commitments[0] != ([32]byte{}) {
		t.Fatal("commitment must not be activated yet")
	}
	if len(remaining) != 1 {
		t.Fatalf("queue must still hold 1 entry, got %d", len(remaining))
	}
}

func TestPendingRandaoCommitments_ActivateOverdue(t *testing.T) {
	// An entry whose activation epoch already passed is applied immediately.
	commitments := make([][32]byte, 8)
	pending := []eip8321.PendingCommitment{{ValidatorIndex: 0, Commitment: seed(0x01), ActivationEpoch: 0}}
	remaining := applyPending(commitments, pending)
	if commitments[0] != seed(0x01) {
		t.Fatal("overdue commitment was not activated")
	}
	if len(remaining) != 0 {
		t.Fatalf("queue must be empty, got %d", len(remaining))
	}
}

func TestPendingRandaoCommitments_DrainPrefixOnly(t *testing.T) {
	commitments := make([][32]byte, 8)
	pending := []eip8321.PendingCommitment{
		{ValidatorIndex: 0, Commitment: seed(0x10), ActivationEpoch: currentEpoch + 1},
		{ValidatorIndex: 1, Commitment: seed(0x11), ActivationEpoch: currentEpoch + 1},
		{ValidatorIndex: 2, Commitment: seed(0x12), ActivationEpoch: currentEpoch + 1},
		{ValidatorIndex: 3, Commitment: seed(0x13), ActivationEpoch: currentEpoch + 3},
	}
	remaining := applyPending(commitments, pending)
	for i := 0; i < 3; i++ {
		if commitments[i] != seed(byte(0x10+i)) {
			t.Fatalf("commitment %d was not activated", i)
		}
	}
	if commitments[3] != ([32]byte{}) {
		t.Fatal("commitment 3 must not be activated yet")
	}
	if len(remaining) != 1 || remaining[0].Commitment != seed(0x13) {
		t.Fatalf("queue must hold only validator 3, got %+v", remaining)
	}
}
