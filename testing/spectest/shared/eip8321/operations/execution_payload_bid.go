package operations

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/consensus-types/blocks"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	common "github.com/OffchainLabs/prysm/v7/testing/spectest/shared/common/operations"
	"github.com/OffchainLabs/prysm/v7/testing/util"
)

// EIP-8321 does not change the bid type (it is the Heze one), so a Heze block
// is enough to carry it. The vector checks that prev_randao in the bid matches
// the RANDAO mix after the hash-chain update.
func blockWithSignedExecutionPayloadBid(bidSSZ []byte) (interfaces.SignedBeaconBlock, error) {
	signedBid := &ethpb.SignedExecutionPayloadBidHeze{}
	if err := signedBid.UnmarshalSSZ(bidSSZ); err != nil {
		return nil, err
	}
	blk := util.NewBeaconBlockHeze()
	blk.Block.Slot = signedBid.Message.Slot
	blk.Block.ParentRoot = signedBid.Message.ParentBlockRoot
	blk.Block.Body.SignedExecutionPayloadBid = signedBid
	return blocks.NewSignedBeaconBlock(blk)
}

// RunExecutionPayloadBidTest runs operations/execution_payload_bid for EIP-8321.
func RunExecutionPayloadBidTest(t *testing.T, config string) {
	common.RunExecutionPayloadBidTest(t, config, "eip8321", blockWithSignedExecutionPayloadBid, sszToState)
}
