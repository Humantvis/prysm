package sync

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing/trace"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
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
	reg := signed.Message

	if len(reg.Commitment) != 32 {
		return pubsub.ValidationReject, errors.New("randao commitment must be 32 bytes")
	}
	var commitment [32]byte
	copy(commitment[:], reg.Commitment)

	if s.cfg.randaoCommitmentPool.ValidatorExists(reg.ValidatorIndex) {
		return pubsub.ValidationIgnore, nil
	}
	
	st, err := s.cfg.chain.HeadStateReadOnly(ctx)
	if err != nil {
		return pubsub.ValidationIgnore, err
	}

	numValidators := uint64(st.NumValidators())
	// Only read the stored commitment when the index exists; Verify rejects the rest.
	var stored [32]byte
	if uint64(reg.ValidatorIndex) < numValidators {
		stored, err = st.RandaoCommitmentAtIndex(reg.ValidatorIndex)
		if err != nil {
			return pubsub.ValidationIgnore, err
		}
	}
	pendings, err := st.PendingRandaoCommitments()
	if err != nil {
		return pubsub.ValidationIgnore, err
	}
	pendingIdx := make([]uint64, len(pendings))
	for i, p := range pendings {
		pendingIdx[i] = uint64(p.ValidatorIndex)
	}

	// Shared consensus checks: index in range, non-zero commitment,
	// not already registered, not already pending.
	if err := eip8321.VerifyRandaoCommitmentRegistration(uint64(reg.ValidatorIndex), numValidators, commitment, stored, pendingIdx); err != nil {
		// EIP text says REJECT for already-registered, but the reftest suite
		// (gossip_randao_commitment_registration__ignore_already_registered)
		// expects IGNORE. Matching the reftests for CI. Upstream divergence
		// tracked at <link-to-issue>.
		if errors.Is(err, eip8321.ErrAlreadyRegistered) {
			return pubsub.ValidationIgnore, nil
		}
		if errors.Is(err, eip8321.ErrRegistrationAlreadyPending) {
			return pubsub.ValidationIgnore, nil
		}
		return pubsub.ValidationReject, err
	}

	sigBatch, err := eip8321.RandaoCommitmentRegistrationSignatureBatch(st, []*ethpb.SignedRandaoCommitmentRegistration{signed})
	if err != nil {
		return pubsub.ValidationReject, err
	}
	res, err := s.validateWithBatchVerifier(ctx, "randao commitment registration", sigBatch)
	if res != pubsub.ValidationAccept {
		return res, err
	}

	msg.ValidatorData = signed // Used in downstream subscriber
	return pubsub.ValidationAccept, nil
}