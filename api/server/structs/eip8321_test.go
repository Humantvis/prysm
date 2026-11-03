package structs

import (
	"encoding/json"
	"testing"

	"github.com/OffchainLabs/prysm/v7/consensus-types/blocks"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestEip8321BlockJSON(t *testing.T) {
	pb := util.NewBeaconBlockEip8321()
	pb.Block.Slot = 32
	pb.Block.Body.HashChainReveal[0] = 42
	pb.Block.Body.RandaoCommitmentRegistrations = []*ethpb.SignedRandaoCommitmentRegistration{{
		Message: &ethpb.RandaoCommitmentRegistration{ValidatorIndex: 7, Commitment: make([]byte, 32)}, Signature: make([]byte, 96),
	}}
	blk, err := blocks.NewSignedBeaconBlock(pb)
	require.NoError(t, err)
	converted, err := SignedBeaconBlockMessageJsoner(blk)
	require.NoError(t, err)
	data, err := converted.MessageRawJson()
	require.NoError(t, err)
	var decoded BeaconBlockEip8321
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, "32", decoded.Slot)
	require.Equal(t, hexutil.Encode(pb.Block.Body.HashChainReveal), decoded.Body.HashChainReveal)
	require.Equal(t, 1, len(decoded.Body.RandaoCommitmentRegistrations))
	require.Equal(t, "7", decoded.Body.RandaoCommitmentRegistrations[0].Message.ValidatorIndex)
	require.Equal(t, hexutil.Encode(pb.Block.Body.SignedExecutionPayloadBid.Message.InclusionListBits), decoded.Body.SignedExecutionPayloadBid.Message.InclusionListBits)
}
