package kv

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/consensus-types/blocks"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/util"
	"github.com/golang/snappy"
)

func TestEip8321Persistence(t *testing.T) {
	store := setupDB(t)
	ctx := t.Context()
	pb := util.NewBeaconBlockEip8321()
	pb.Block.Body.HashChainReveal[0] = 42
	blk, err := blocks.NewSignedBeaconBlock(pb)
	require.NoError(t, err)
	root, err := blk.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, store.SaveBlock(ctx, blk))
	got, err := store.Block(ctx, root)
	require.NoError(t, err)
	require.Equal(t, version.Eip8321, got.Version())
	wantBytes, err := blk.MarshalSSZ()
	require.NoError(t, err)
	gotBytes, err := got.MarshalSSZ()
	require.NoError(t, err)
	require.DeepEqual(t, wantBytes, gotBytes)

	st, err := util.NewBeaconStateEip8321(func(s *ethpb.BeaconStateHeze) error {
		s.Validators = []*ethpb.Validator{{PublicKey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32)}}
		s.Balances = []uint64{32000000000}
		s.InactivityScores = []uint64{0}
		s.PreviousEpochParticipation = []byte{0}
		s.CurrentEpochParticipation = []byte{0}
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, st.UpdateRandaoCommitmentAtIndex(0, [32]byte{42}))
	require.NoError(t, st.AppendPendingRandaoCommitment(&ethpb.PendingRandaoCommitment{ValidatorIndex: 0, Commitment: make([]byte, 32), ActivationEpoch: 7}))
	require.NoError(t, store.SaveState(ctx, st, root))
	loaded, err := store.State(ctx, root)
	require.NoError(t, err)
	require.Equal(t, version.Eip8321, loaded.Version())
	wantBytes, err = st.MarshalSSZ()
	require.NoError(t, err)
	gotBytes, err = loaded.MarshalSSZ()
	require.NoError(t, err)
	require.DeepEqual(t, wantBytes, gotBytes)
	encoded, err := encodeStateWithKey(st)
	require.NoError(t, err)
	encoded, err = snappy.Decode(nil, encoded)
	require.NoError(t, err)
	snapshot, err := decodeStateSnapshot(encoded)
	require.NoError(t, err)
	gotBytes, err = snapshot.MarshalSSZ()
	require.NoError(t, err)
	require.DeepEqual(t, wantBytes, gotBytes)
}
