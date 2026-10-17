package state_native

import (
	"testing"

	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

func TestEip8321CommitmentValidationAndOwnership(t *testing.T) {
	st, err := InitializeFromProtoUnsafeEip8321(&ethpb.BeaconStateEip8321{Validators: []*ethpb.Validator{{}}})
	require.NoError(t, err)
	require.ErrorContains(t, "length does not match", st.SetRandaoCommitments(nil))
	require.ErrorContains(t, "32 bytes", st.SetRandaoCommitments([][]byte{{1}}))
	input := [][]byte{make([]byte, 32)}
	input[0][0] = 7
	require.NoError(t, st.SetRandaoCommitments(input))
	input[0][0] = 9
	got, err := st.RandaoCommitments()
	require.NoError(t, err)
	require.Equal(t, byte(7), got[0][0])
	got[0][0] = 12
	value, err := st.RandaoCommitmentAtIndex(0)
	require.NoError(t, err)
	require.Equal(t, byte(7), value[0])
	_, err = st.RandaoCommitmentAtIndex(1)
	require.ErrorContains(t, "out of range", err)
	require.ErrorContains(t, "invalid pending", st.AppendPendingRandaoCommitment(nil))
	pending := &ethpb.PendingRandaoCommitment{Commitment: make([]byte, 32)}
	require.NoError(t, st.AppendPendingRandaoCommitment(pending))
	pending.Commitment[0] = 5
	queue, err := st.PendingRandaoCommitments()
	require.NoError(t, err)
	require.Equal(t, byte(0), queue[0].Commitment[0])
	queue[0].Commitment[0] = 6
	queue, err = st.PendingRandaoCommitments()
	require.NoError(t, err)
	require.Equal(t, byte(0), queue[0].Commitment[0])
}
