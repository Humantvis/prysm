package sync

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing"
	"github.com/OffchainLabs/prysm/v7/monitoring/tracing/trace"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (s *Service) validateRandaoCommitmentRegistration(ctx context.Context, pid peer.ID, msg *pubsub.Message) (pubsub.ValidationResult, error) {
	epoch := s.cfg.chain.CurrentSlot().ToEpoch()
	// Ignore pre EIP-8321 registrations.
	if !isEip8321Active(epoch) {
		return pubsub.ValidationIgnore, nil
	}

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

	res, err := s.validateRandaoCommitmentRegistrationAtEpoch(ctx, epoch, signed)
	if res == pubsub.ValidationAccept {
		msg.ValidatorData = signed // Used in downstream subscriber
	}
	return res, err
}

// validateRandaoCommitmentRegistrationAtEpoch is the runner-facing core of
// gossip validation for SignedRandaoCommitmentRegistration. It must not
// consult the wall clock — the caller supplies epoch, so reftests can drive
// pre-fork and post-fork vectors deterministically. It must not touch
// *pubsub.Message — the runner has already decoded the typed object.
func (s *Service) validateRandaoCommitmentRegistrationAtEpoch(
	ctx context.Context,
	epoch primitives.Epoch,
	signed *ethpb.SignedRandaoCommitmentRegistration,
) (pubsub.ValidationResult, error) {
	// Redundant with the outer wrapper, but required so the runner can test
	// the pre-fork vector without faking the chain clock.
	if !isEip8321Active(epoch) {
		return pubsub.ValidationIgnore, nil
	}

	if signed == nil || signed.Message == nil {
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

	if st.Version() < version.Eip8321 {
		return pubsub.ValidationIgnore, nil
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
	verified, err := sigBatch.Verify()
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "could not verify randao commitment registration")
	}
	if !verified {
		return pubsub.ValidationReject, errors.New("randao commitment registration signature verification failed")
	}

	return pubsub.ValidationAccept, nil
}

// isEip8321Active reports whether EIP-8321 is active at epoch. FarFutureEpoch
// is the "not scheduled" sentinel and is never active (see config_test.go).
func isEip8321Active(epoch primitives.Epoch) bool {
	forkEpoch := params.BeaconConfig().Eip8321ForkEpoch
	return forkEpoch != params.BeaconConfig().FarFutureEpoch && epoch >= forkEpoch
}
