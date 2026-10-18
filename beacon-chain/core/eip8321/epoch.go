package eip8321

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/time"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
)

func ProcessPendingRandaoCommitments(st state.BeaconState) error {
	// Read queue
	pending, err := st.PendingRandaoCommitments()
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}

	// Convert from typed numbers and byte slices to plain numbers and fixed arrays
	queue := make([]PendingCommitment, len(pending))
	for i, p := range pending {
		var commitment [32]byte
		copy(commitment[:], p.Commitment)
		queue[i] = PendingCommitment{
			ValidatorIndex:  uint64(p.ValidatorIndex),
			Commitment:      commitment,
			ActivationEpoch: uint64(p.ActivationEpoch),
		}
	}

	// Calls rule from pending_commitments.go
	activate, _ := SplitPendingRandaoCommitments(queue, uint64(time.CurrentEpoch(st)))
	if len(activate) == 0 {
		return nil
	}

	// Activation (from the next block, this validator has to use the hash chain)
	for _, p := range activate {
		if err := st.UpdateRandaoCommitmentAtIndex(primitives.ValidatorIndex(p.ValidatorIndex), p.Commitment); err != nil {
			return err
		}
	}
	return st.SetPendingRandaoCommitments(pending[len(activate):])
}
