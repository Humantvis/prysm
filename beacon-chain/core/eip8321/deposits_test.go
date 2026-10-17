package eip8321_test

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/electra"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/config/params"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

func TestNewValidatorStartsWithZeroRandaoCommitment(t *testing.T) {
	st, err := state_native.InitializeFromProtoUnsafeEip8321(&ethpb.BeaconStateEip8321{})
	require.NoError(t, err)
	credentials := make([]byte, 32)
	credentials[0] = 1
	require.NoError(t, electra.AddValidatorToRegistry(st, make([]byte, 48), credentials, params.BeaconConfig().MinActivationBalance))
	commitments, err := st.RandaoCommitments()
	require.NoError(t, err)
	require.Equal(t, st.NumValidators(), len(commitments))
	require.DeepEqual(t, [][]byte{make([]byte, 32)}, commitments)
}
