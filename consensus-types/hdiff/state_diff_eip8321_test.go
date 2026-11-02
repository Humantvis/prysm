package hdiff

import (
	"encoding/binary"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/heze"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

func TestEip8321Diff(t *testing.T) {
	gloas, err := gloasState(t, 64)
	require.NoError(t, err)
	require.NoError(t, gloas.SetDepositRequestsStartIndex(gloas.Eth1DepositIndex()))
	hezeState, err := heze.UpgradeToHeze(gloas)
	require.NoError(t, err)
	feature, err := eip8321.UpgradeToEip8321(hezeState.Copy())
	require.NoError(t, err)
	require.NoError(t, feature.UpdateRandaoCommitmentAtIndex(0, [32]byte{42}))
	require.NoError(t, feature.AppendPendingRandaoCommitment(&ethpb.PendingRandaoCommitment{ValidatorIndex: 1, Commitment: make([]byte, 32), ActivationEpoch: 7}))
	roundTrip := func(t *testing.T, source, target state.BeaconState) {
		t.Helper()
		encoded, err := Diff(source, target)
		require.NoError(t, err)
		got, err := ApplyDiff(t.Context(), source.Copy(), encoded)
		require.NoError(t, err)
		requireEqualState(t, target, got)
	}
	t.Run("Heze to feature fork retains registrations", func(t *testing.T) { roundTrip(t, hezeState, feature) })
	t.Run("changed commitment and drained pending queue", func(t *testing.T) {
		target := feature.Copy()
		require.NoError(t, target.UpdateRandaoCommitmentAtIndex(0, [32]byte{99}))
		require.NoError(t, target.SetPendingRandaoCommitments(nil))
		roundTrip(t, feature, target)
	})
	t.Run("validator registry growth precedes commitment restoration", func(t *testing.T) {
		target := feature.Copy()
		require.NoError(t, target.AppendValidator(&ethpb.Validator{PublicKey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32)}))
		require.NoError(t, target.AppendBalance(32000000000))
		commitments, err := target.RandaoCommitments()
		require.NoError(t, err)
		commitments = append(commitments, make([]byte, 32))
		require.NoError(t, target.SetRandaoCommitments(commitments))
		roundTrip(t, feature, target)
	})
}

func TestEip8321DiffRejectsTruncatedLists(t *testing.T) {
	for _, input := range [][]byte{nil, make([]byte, 8), binary.LittleEndian.AppendUint64(nil, ^uint64(0)), binary.LittleEndian.AppendUint64(make([]byte, 8), 1)} {
		data := input
		s := &stateDiff{}
		require.NotNil(t, s.readEip8321Fields(&data))
	}
}
