package epoch_processing

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/electra"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
)

// RunRandaoMixesResetTests executes "epoch_processing/randao_mixes_reset" tests.
func RunRandaoMixesResetTests(t *testing.T, config string) {
	runEpochHandler(t, config, "epoch_processing/randao_mixes_reset/pyspec_tests", func(_ *testing.T, st state.BeaconState) (state.BeaconState, error) {
		return electra.ProcessRandaoMixesReset(st)
	})
}
