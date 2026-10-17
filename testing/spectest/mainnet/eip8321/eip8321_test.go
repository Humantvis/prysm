package eip8321

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/epoch_processing"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/fork"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/operations"
)

// Official EIP-8321 test vectors. Generate them with consensus-specs
// (make test reftests=true fork=eip8321) and expose them under
// third_party/testdata/tests/mainnet/eip8321 (see docs/eip8321-spectests.md).

func TestMainnet_Eip8321_Operations_Randao(t *testing.T) {
	operations.RunRandaoTest(t, "mainnet")
}

func TestMainnet_Eip8321_Operations_RandaoCommitmentRegistration(t *testing.T) {
	operations.RunRandaoCommitmentRegistrationTest(t, "mainnet")
}

func TestMainnet_Eip8321_EpochProcessing_PendingRandaoCommitments(t *testing.T) {
	epoch_processing.RunPendingRandaoCommitmentsTests(t, "mainnet")
}

func TestMainnet_Eip8321_EpochProcessing_RandaoMixesReset(t *testing.T) {
	epoch_processing.RunRandaoMixesResetTests(t, "mainnet")
}

func TestMainnet_Eip8321_Fork_UpgradeToEip8321(t *testing.T) {
	fork.RunUpgradeToEip8321(t, "mainnet")
}
