package blocks

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/container/trie"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	"github.com/OffchainLabs/prysm/v7/encoding/ssz"
	eth "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

func TestEip8321BlockProgressiveSSZ(t *testing.T) {
	h := hydrateBeaconBlockBodyHeze()
	body := &eth.BeaconBlockBodyEip8321{
		RandaoReveal: h.RandaoReveal, Graffiti: h.Graffiti, SyncAggregate: h.SyncAggregate,
		SignedExecutionPayloadBid: h.SignedExecutionPayloadBid, ParentExecutionRequests: h.ParentExecutionRequests,
		HashChainReveal: make([]byte, 32),
	}
	// The processing limit is separate from ProgressiveList serialization.
	for i := 0; i < 129; i++ {
		body.RandaoCommitmentRegistrations = append(body.RandaoCommitmentRegistrations, &eth.SignedRandaoCommitmentRegistration{
			Message: &eth.RandaoCommitmentRegistration{Commitment: make([]byte, 32)}, Signature: make([]byte, 96),
		})
	}
	body.HashChainReveal[0] = 42
	pb := &eth.SignedBeaconBlockEip8321{Block: &eth.BeaconBlockEip8321{
		ParentRoot: make([]byte, 32), StateRoot: make([]byte, 32), Body: body,
	}, Signature: make([]byte, 96)}
	block, err := NewSignedBeaconBlock(pb)
	require.NoError(t, err)
	require.Equal(t, version.Eip8321, block.Version())
	encoded, err := block.MarshalSSZ()
	require.NoError(t, err)
	decoded := &eth.SignedBeaconBlockEip8321{}
	require.NoError(t, decoded.UnmarshalSSZ(encoded))
	require.Equal(t, 129, len(decoded.Block.Body.RandaoCommitmentRegistrations))
	bodyRoot, err := body.HashTreeRoot()
	require.NoError(t, err)
	roots, err := ComputeBlockBodyFieldRoots(t.Context(), block.Block().Body().(*BeaconBlockBody))
	require.NoError(t, err)
	require.Equal(t, 15, len(roots))
	require.DeepEqual(t, body.HashChainReveal, roots[13])
	active, err := progressiveContainerActiveFields(version.Eip8321, beaconBlockBodyProgressiveContainer)
	require.NoError(t, err)
	chunks := make([][32]byte, 0, len(roots))
	for i, fieldRoot := range roots {
		if active[i] {
			chunks = append(chunks, bytesutil.ToBytes32(fieldRoot))
		}
	}
	root, err := ssz.ContainerRootProgressive(chunks, active)
	require.NoError(t, err)
	require.Equal(t, bodyRoot, root)
	proof, err := ExecutionBlockHashProof(t.Context(), block.Block())
	require.NoError(t, err)
	require.Equal(t, true, trie.VerifyMerkleProof(bodyRoot[:], h.SignedExecutionPayloadBid.Message.ParentBlockHash, 2856, proof))
	require.ErrorContains(t, "32 bytes", block.SetHashChainReveal([]byte{1}))
	require.NoError(t, block.SetHashChainReveal(make([]byte, 32)))
	copyBlock, err := block.Copy()
	require.NoError(t, err)
	require.Equal(t, version.Eip8321, copyBlock.Version())
}
