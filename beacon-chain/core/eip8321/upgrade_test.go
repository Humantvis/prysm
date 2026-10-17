package eip8321_test

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/heze"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	"github.com/OffchainLabs/prysm/v7/config/params"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/util"
)

func TestUpgradeToEip8321(t *testing.T) {
	setupEip8321Preset(t)
	pre, err := util.NewBeaconStateGloas(func(s *ethpb.BeaconStateGloas) error {
		s.CurrentSyncCommittee.Pubkeys = s.CurrentSyncCommittee.Pubkeys[:fieldparams.SyncCommitteeLength]
		s.NextSyncCommittee.Pubkeys = s.NextSyncCommittee.Pubkeys[:fieldparams.SyncCommitteeLength]
		s.ProposerLookahead = s.ProposerLookahead[:2*fieldparams.SlotsPerEpoch]
		s.BuilderPendingPayments = s.BuilderPendingPayments[:2*fieldparams.SlotsPerEpoch]
		s.ExecutionPayloadAvailability = s.ExecutionPayloadAvailability[:fieldparams.BlockRootsLength/8]
		s.Slot = 64
		s.Validators = []*ethpb.Validator{{PublicKey: make([]byte, 48), WithdrawalCredentials: make([]byte, 32)}}
		s.Balances = []uint64{32_000_000_000}
		s.InactivityScores = []uint64{0}
		s.PreviousEpochParticipation = []byte{0}
		s.CurrentEpochParticipation = []byte{0}
		return nil
	})
	require.NoError(t, err)
	pre, err = heze.UpgradeToHeze(pre)
	require.NoError(t, err)
	preBytes, err := pre.MarshalSSZ()
	require.NoError(t, err)
	post, err := eip8321.UpgradeToEip8321(pre)
	require.NoError(t, err)
	require.Equal(t, version.Eip8321, post.Version())
	pb, err := state_native.ProtobufBeaconStateEip8321(post.ToProto())
	require.NoError(t, err)
	require.DeepEqual(t, params.BeaconConfig().HezeForkVersion, pb.Fork.PreviousVersion)
	require.DeepEqual(t, params.BeaconConfig().Eip8321ForkVersion, pb.Fork.CurrentVersion)
	require.DeepEqual(t, [][]byte{make([]byte, 32)}, pb.RandaoCommitments)
	require.Equal(t, 0, len(pb.PendingRandaoCommitments))
	want, err := pb.HashTreeRoot()
	require.NoError(t, err)
	got, err := post.HashTreeRoot(t.Context())
	require.NoError(t, err)
	require.Equal(t, want, got)

	// New fields must be independent in copies and invalidate the cached state root.
	clone := post.Copy()
	commitment := [32]byte{42}
	require.NoError(t, clone.UpdateRandaoCommitmentAtIndex(0, commitment))
	require.NoError(t, clone.AppendPendingRandaoCommitment(&ethpb.PendingRandaoCommitment{Commitment: commitment[:]}))
	unchanged, err := post.RandaoCommitmentAtIndex(0)
	require.NoError(t, err)
	require.Equal(t, [32]byte{}, unchanged)
	clonedProto, err := state_native.ProtobufBeaconStateEip8321(clone.ToProto())
	require.NoError(t, err)
	want, err = clonedProto.HashTreeRoot()
	require.NoError(t, err)
	got, err = clone.HashTreeRoot(t.Context())
	require.NoError(t, err)
	require.Equal(t, want, got)
	after, err := pre.MarshalSSZ()
	require.NoError(t, err)
	require.DeepEqual(t, preBytes, after)
	_, err = pre.RandaoCommitments()
	require.ErrorContains(t, "not supported", err)
}

func setupEip8321Preset(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	if fieldparams.Preset == "minimal" {
		params.OverrideBeaconConfig(params.MinimalSpecConfig())
	} else {
		params.OverrideBeaconConfig(params.MainnetConfig())
	}
}

func TestUpgradeToEip8321RejectsOtherForks(t *testing.T) {
	setupEip8321Preset(t)
	_, err := eip8321.UpgradeToEip8321(nil)
	require.ErrorContains(t, "requires a Heze state", err)
	st, err := util.NewBeaconStateGloas()
	require.NoError(t, err)
	_, err = eip8321.UpgradeToEip8321(st)
	require.ErrorContains(t, "requires a Heze state", err)
}
