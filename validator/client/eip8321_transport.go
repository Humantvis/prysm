package client

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/time/slots"
	"github.com/OffchainLabs/prysm/v7/validator/client/iface"
	"github.com/golang/protobuf/ptypes/empty"
)

// EIP-8321 block types use gRPC while Heze duties keep their REST transport.
type eip8321ValidatorClient struct {
	iface.ValidatorClient
	grpcClient iface.ValidatorClient
}

func eip8321Slot(slot primitives.Slot) bool {
	return slots.ToEpoch(slot) >= params.BeaconConfig().Eip8321ForkEpoch
}

func (c *eip8321ValidatorClient) RandaoCommitment(ctx context.Context, req *ethpb.ValidatorIndexRequest) (*ethpb.RandaoCommitmentResponse, error) {
	return c.grpcClient.RandaoCommitment(ctx, req)
}

func (c *eip8321ValidatorClient) SubmitRandaoCommitmentRegistration(ctx context.Context, req *ethpb.SignedRandaoCommitmentRegistration) (*empty.Empty, error) {
	return c.grpcClient.SubmitRandaoCommitmentRegistration(ctx, req)
}

func (c *eip8321ValidatorClient) BeaconBlock(ctx context.Context, req *ethpb.BlockRequest) (*ethpb.GenericBeaconBlock, error) {
	if req != nil && eip8321Slot(req.Slot) {
		return c.grpcClient.BeaconBlock(ctx, req)
	}
	return c.ValidatorClient.BeaconBlock(ctx, req)
}

func (c *eip8321ValidatorClient) ProposeBeaconBlock(ctx context.Context, req *ethpb.GenericSignedBeaconBlock) (*ethpb.ProposeResponse, error) {
	if req.GetEip8321() != nil {
		return c.grpcClient.ProposeBeaconBlock(ctx, req)
	}
	return c.ValidatorClient.ProposeBeaconBlock(ctx, req)
}

func (c *eip8321ValidatorClient) GetExecutionPayloadEnvelope(ctx context.Context, slot primitives.Slot, root [32]byte) (*ethpb.ExecutionPayloadEnvelope, error) {
	if eip8321Slot(slot) {
		return c.grpcClient.GetExecutionPayloadEnvelope(ctx, slot, root)
	}
	return c.ValidatorClient.GetExecutionPayloadEnvelope(ctx, slot, root)
}

func (c *eip8321ValidatorClient) PublishExecutionPayloadEnvelope(ctx context.Context, req *ethpb.SignedExecutionPayloadEnvelope) (*empty.Empty, error) {
	if payload := req.GetMessage().GetPayload(); payload != nil && eip8321Slot(primitives.Slot(payload.SlotNumber)) {
		return c.grpcClient.PublishExecutionPayloadEnvelope(ctx, req)
	}
	return c.ValidatorClient.PublishExecutionPayloadEnvelope(ctx, req)
}
