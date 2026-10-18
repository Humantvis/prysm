package sync

import (
    "context"
    "fmt"

    "github.com/OffchainLabs/prysm/v7/beacon-chain/core/feed/operation"
    ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
    "google.golang.org/protobuf/proto"
    log "github.com/sirupsen/logrus"
)

// randaoCommitmentRegistrationSubscriber inserts a gossip message that has already
// passed validateRandaoCommitmentRegistration into the operations pool, where the
// block proposer can pick it up for inclusion.
func (s *Service) randaoCommitmentRegistrationSubscriber(ctx context.Context, msg proto.Message) error {
    signed, ok := msg.(*ethpb.SignedRandaoCommitmentRegistration)
    if !ok {
        return fmt.Errorf("message was not type *ethpb.SignedRandaoCommitmentRegistration, was type %T", msg)
    }

    // Expected when gossip redelivers or two messages race validation; the pool
    // itself also dedupes, so this is just an early-out.
    if s.cfg.randaoCommitmentPool.ValidatorExists(signed.Message.ValidatorIndex) {
        return nil
    }

    s.cfg.randaoCommitmentPool.InsertRandaoCommitment(signed)

    log.WithField("validatorIndex", signed.Message.ValidatorIndex).
        Debug("Randao commitment registration inserted into pool")

    //.cfg.operationNotifier.OperationFeed().Send(&operation.RandaoCommitmentRegistrationReceived{
    //   Registration: signed,
    //)
    return nil
}