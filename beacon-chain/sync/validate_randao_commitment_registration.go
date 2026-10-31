package sync

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing/trace"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (s *Service) validateRandaoCommitmentRegistration(ctx context.Context, pid peer.ID, msg *pubsub.Message) (pubsub.ValidationResult, error) {
	// Validation runs on publish too, so accept our own messages.
	if pid == s.cfg.p2p.PeerID() {
		return pubsub.ValidationAccept, nil
	}

	// The head state will be too far behind to validate while syncing.
	if s.cfg.initialSync.Syncing() {
		return pubsub.ValidationIgnore, nil
	}

	ctx, span := trace.StartSpan(ctx, "sync.validateRandaoCommitmentRegistration")
	defer span.End()

	m, err := s.decodePubsubMessage(msg)
	if err != nil {
		tracing.AnnotateError(span, err)
		return pubsub.ValidationReject, err
	}

	signed, ok := m.(*ethpb.SignedRandaoCommitmentRegistration)
	if !ok {
		return pubsub.ValidationReject, errWrongMessage
	}
	if signed.Message == nil {
		return pubsub.ValidationReject, errNilMessage
	}
	st, err := s.cfg.chain.HeadStateReadOnly(ctx)
	if err != nil {
		return pubsub.ValidationIgnore, err
	}

	// All spec checks except BLS live in core/eip8321 so the spec tests can run them.
	result, sigBatch, err := eip8321.ValidateRandaoCommitmentRegistrationGossip(st, s.cfg.randaoCommitmentPool.ValidatorExists, signed)
	switch result {
	case eip8321.GossipIgnore:
		return pubsub.ValidationIgnore, nil
	case eip8321.GossipReject:
		tracing.AnnotateError(span, err)
		return pubsub.ValidationReject, err
	}

	res, err := s.validateWithBatchVerifier(ctx, "randao commitment registration", sigBatch)
	if res != pubsub.ValidationAccept {
		return res, err
	}

	msg.ValidatorData = signed // Used in downstream subscriber
	return pubsub.ValidationAccept, nil
}
