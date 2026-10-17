package epoch_processing

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
)

// RunPendingRandaoCommitmentsTests executes "epoch_processing/pending_randao_commitments" tests.
func RunPendingRandaoCommitmentsTests(t *testing.T, config string) {
	runEpochHandler(t, config, "epoch_processing/pending_randao_commitments/pyspec_tests", func(_ *testing.T, st state.BeaconState) (state.BeaconState, error) {
		if err := eip8321.ProcessPendingRandaoCommitments(st); err != nil {
			return nil, err
		}
		return st, nil
	})
}
