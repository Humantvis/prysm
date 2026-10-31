package eip8321

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/epoch_processing"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/fork"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/networking"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/operations"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/shared/eip8321/sanity"
)

// Official EIP-8321 test vectors. Generate them with consensus-specs
// (make test reftests=true fork=eip8321) and expose them under
// third_party/testdata/tests/minimal/eip8321 (see docs/eip8321-spectests.md).

func TestMinimal_Eip8321_Operations_Randao(t *testing.T) {
	operations.RunRandaoTest(t, "minimal")
}

func TestMinimal_Eip8321_Operations_RandaoCommitmentRegistration(t *testing.T) {
	operations.RunRandaoCommitmentRegistrationTest(t, "minimal")
}

func TestMinimal_Eip8321_EpochProcessing_PendingRandaoCommitments(t *testing.T) {
	epoch_processing.RunPendingRandaoCommitmentsTests(t, "minimal")
}

func TestMinimal_Eip8321_EpochProcessing_RandaoMixesReset(t *testing.T) {
	epoch_processing.RunRandaoMixesResetTests(t, "minimal")
}

func TestMinimal_Eip8321_Fork_UpgradeToEip8321(t *testing.T) {
	fork.RunUpgradeToEip8321(t, "minimal")
}

// The mainnet lifecycle vector is not generated upstream (it needs 256
// registrations in one block, above MAX_RANDAO_COMMITMENT_REGISTRATIONS = 128).
func TestMinimal_Eip8321_Sanity_Blocks(t *testing.T) {
	sanity.RunBlockProcessingTest(t, "minimal", "sanity/blocks/pyspec_tests")
}

func TestMinimal_Eip8321_Operations_ExecutionPayloadBid(t *testing.T) {
	operations.RunExecutionPayloadBidTest(t, "minimal")
}

func TestMinimal_Eip8321_Networking_GossipRandaoCommitmentRegistration(t *testing.T) {
	networking.RunGossipRandaoCommitmentRegistrationTest(t, "minimal")
}
