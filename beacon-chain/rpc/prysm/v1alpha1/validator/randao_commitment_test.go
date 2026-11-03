//go:build minimal

package validator

import (
	"testing"

	mockchain "github.com/OffchainLabs/prysm/v7/beacon-chain/blockchain/testing"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/heze"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/util"
)

func TestRandaoCommitmentEpochBoundary(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.OverrideBeaconConfig(params.MinimalSpecConfig())
	st, keys := util.DeterministicGenesisStateGloas(t, 64)
	st, err := heze.UpgradeToHeze(st)
	require.NoError(t, err)
	st, err = eip8321.UpgradeToEip8321(st)
	require.NoError(t, err)
	target := primitives.Slot(2 * params.BeaconConfig().SlotsPerEpoch)
	require.NoError(t, st.SetSlot(target-1))
	commitment := make([]byte, 32)
	commitment[0] = 42
	require.NoError(t, st.AppendPendingRandaoCommitment(&ethpb.PendingRandaoCommitment{
		ValidatorIndex: 0, Commitment: commitment, ActivationEpoch: 2,
	}))
	chain := &mockchain.ChainService{State: st, Slot: &target}
	vs := &Server{HeadFetcher: chain, TimeFetcher: chain}
	response, err := vs.RandaoCommitment(t.Context(), &ethpb.ValidatorIndexRequest{PublicKey: keys[0].PublicKey().Marshal()})
	require.NoError(t, err)
	require.DeepEqual(t, commitment, response.Commitment)
	// The query must never mutate the canonical head while advancing a copy.
	stored, err := st.RandaoCommitmentAtIndex(0)
	require.NoError(t, err)
	require.Equal(t, [32]byte{}, stored)
	require.Equal(t, target-1, st.Slot())
}
