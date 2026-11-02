package util

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
)

func NewBeaconStateEip8321(options ...func(*ethpb.BeaconStateHeze) error) (state.BeaconState, error) {
	st, err := NewBeaconStateHeze(options...)
	if err != nil {
		return nil, err
	}
	return eip8321.UpgradeToEip8321(st)
}

func NewBeaconBlockEip8321() *ethpb.SignedBeaconBlockEip8321 {
	h := NewBeaconBlockHeze()
	b := h.Block.Body
	return &ethpb.SignedBeaconBlockEip8321{
		Signature: h.Signature,
		Block: &ethpb.BeaconBlockEip8321{
			ParentRoot: h.Block.ParentRoot, StateRoot: h.Block.StateRoot,
			Body: &ethpb.BeaconBlockBodyEip8321{
				RandaoReveal: b.RandaoReveal, Graffiti: b.Graffiti,
				SyncAggregate: b.SyncAggregate, SignedExecutionPayloadBid: b.SignedExecutionPayloadBid,
				ParentExecutionRequests: b.ParentExecutionRequests, HashChainReveal: make([]byte, 32),
			},
		},
	}
}
