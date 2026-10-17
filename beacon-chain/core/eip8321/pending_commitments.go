package eip8321

// PendingCommitment is a fork-agnostic copy of the spec's PendingRandaoCommitment.
// It lets this package work without depending on the generated proto types.
type PendingCommitment struct {
	ValidatorIndex  uint64
	Commitment      [32]byte
	ActivationEpoch uint64
}

// SplitPendingRandaoCommitments decides which queued registrations become
// active in this epoch transition. It returns the entries to activate (in
// queue order) and the entries that stay in the queue.
//
// The caller then writes each activated commitment into
// state.randao_commitments[ValidatorIndex] and stores `remaining` as the new
// pending queue.
//
// Spec (beacon-chain.md, EIP-8321):
//
//	def process_pending_randao_commitments(state: BeaconState) -> None:
//	    next_epoch = get_current_epoch(state) + 1
//	    next_pending_commitment = 0
//	    for pending_commitment in state.pending_randao_commitments:
//	        if pending_commitment.activation_epoch > next_epoch:
//	            break
//	        index = pending_commitment.validator_index
//	        state.randao_commitments[index] = pending_commitment.commitment
//	        next_pending_commitment += 1
//	    state.pending_randao_commitments = state.pending_randao_commitments[next_pending_commitment:]
func SplitPendingRandaoCommitments(pending []PendingCommitment, currentEpoch uint64) (activate, remaining []PendingCommitment) {
	nextEpoch := currentEpoch + 1
	count := 0
	for _, p := range pending {
		// The queue is in registration order, so activation epochs never
		// decrease. The first entry that is not due yet stops the drain.
		if p.ActivationEpoch > nextEpoch {
			break
		}
		count++
	}
	return pending[:count], pending[count:]
}
