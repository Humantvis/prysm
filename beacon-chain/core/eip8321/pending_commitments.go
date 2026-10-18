package eip8321

type PendingCommitment struct {
	ValidatorIndex  uint64
	Commitment      [32]byte
	ActivationEpoch uint64
}

func SplitPendingRandaoCommitments(pending []PendingCommitment, currentEpoch uint64) (activate, remaining []PendingCommitment) {
	nextEpoch := currentEpoch + 1
	count := 0
	for _, p := range pending {
		if p.ActivationEpoch > nextEpoch {
			break
		}
		count++
	}
	return pending[:count], pending[count:]
}
