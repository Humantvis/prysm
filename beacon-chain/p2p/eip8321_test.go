package p2p

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/OffchainLabs/prysm/v7/config/params"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

func TestEip8321Gossip(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.HezeForkEpoch = 3
	cfg.Eip8321ForkEpoch = 4
	params.OverrideBeaconConfig(cfg)
	require.Equal(t, reflect.TypeOf(&ethpb.SignedBeaconBlockHeze{}), reflect.TypeOf(GossipTopicMappings(BlockSubnetTopicFormat, 3)))
	require.Equal(t, reflect.TypeOf(&ethpb.SignedBeaconBlockEip8321{}), reflect.TypeOf(GossipTopicMappings(BlockSubnetTopicFormat, 4)))
	require.Equal(t, BlockSubnetTopicFormat, GossipTypeMapping[reflect.TypeOf(&ethpb.SignedBeaconBlockEip8321{})])
	s := &Service{}
	score, err := s.topicScoreParams(fmt.Sprintf(RandaoCommitmentRegistrationTopicFormat, []byte{1, 2, 3, 4}))
	require.NoError(t, err)
	require.NotNil(t, score)
	require.Equal(t, true, score.InvalidMessageDeliveriesWeight < 0)
}
