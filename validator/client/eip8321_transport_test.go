package client

import (
	"testing"

	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	validatormock "github.com/OffchainLabs/prysm/v7/testing/validator-mock"
	"go.uber.org/mock/gomock"
)

func TestEip8321Transport(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.Eip8321ForkEpoch = 4
	params.OverrideBeaconConfig(cfg)
	ctrl := gomock.NewController(t)
	rest := validatormock.NewMockValidatorClient(ctrl)
	grpc := validatormock.NewMockValidatorClient(ctrl)
	c := &eip8321ValidatorClient{ValidatorClient: rest, grpcClient: grpc}
	ctx := t.Context()
	pre := &ethpb.BlockRequest{Slot: primitives.Slot(4*cfg.SlotsPerEpoch - 1)}
	post := &ethpb.BlockRequest{Slot: primitives.Slot(4 * cfg.SlotsPerEpoch)}
	rest.EXPECT().BeaconBlock(ctx, pre).Return(&ethpb.GenericBeaconBlock{}, nil)
	grpc.EXPECT().BeaconBlock(ctx, post).Return(&ethpb.GenericBeaconBlock{}, nil)
	_, err := c.BeaconBlock(ctx, pre)
	require.NoError(t, err)
	_, err = c.BeaconBlock(ctx, post)
	require.NoError(t, err)
	heze := &ethpb.GenericSignedBeaconBlock{Block: &ethpb.GenericSignedBeaconBlock_Heze{Heze: &ethpb.SignedBeaconBlockHeze{}}}
	eip := &ethpb.GenericSignedBeaconBlock{Block: &ethpb.GenericSignedBeaconBlock_Eip8321{Eip8321: &ethpb.SignedBeaconBlockEip8321{}}}
	rest.EXPECT().ProposeBeaconBlock(ctx, heze).Return(nil, nil)
	grpc.EXPECT().ProposeBeaconBlock(ctx, eip).Return(nil, nil)
	_, err = c.ProposeBeaconBlock(ctx, heze)
	require.NoError(t, err)
	_, err = c.ProposeBeaconBlock(ctx, eip)
	require.NoError(t, err)
	query := &ethpb.ValidatorIndexRequest{PublicKey: make([]byte, 48)}
	grpc.EXPECT().RandaoCommitment(ctx, query).Return(nil, nil)
	_, err = c.RandaoCommitment(ctx, query)
	require.NoError(t, err)
	reg := &ethpb.SignedRandaoCommitmentRegistration{}
	grpc.EXPECT().SubmitRandaoCommitmentRegistration(ctx, reg).Return(nil, nil)
	_, err = c.SubmitRandaoCommitmentRegistration(ctx, reg)
	require.NoError(t, err)
	// Embedded duties and connection tracking must still use the REST client.
	rest.EXPECT().ConnectionGeneration().Return(uint64(7))
	require.Equal(t, uint64(7), c.ConnectionGeneration())
}
