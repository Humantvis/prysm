package eip8321

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/time"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
)

// ProcessPendingRandaoCommitments implements process_pending_randao_commitments.
// It runs in process_epoch right after process_registry_updates: every queued
// registration whose activation epoch has been reached is written into
// randao_commitments, and those entries are removed from the front of the queue.
// The decision itself is made by SplitPendingRandaoCommitments.
func ProcessPendingRandaoCommitments(st state.BeaconState) error {
	pending, err := st.PendingRandaoCommitments()
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
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
	activate, _ := SplitPendingRandaoCommitments(queue, uint64(time.CurrentEpoch(st)))
	if len(activate) == 0 {
		return nil
	}
	for _, p := range activate {
		if err := st.UpdateRandaoCommitmentAtIndex(primitives.ValidatorIndex(p.ValidatorIndex), p.Commitment); err != nil {
			return err
		}
	}
	return st.SetPendingRandaoCommitments(pending[len(activate):])
}
