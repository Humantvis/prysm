package validator

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/feed"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/feed/operation"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/p2p"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
    fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
)

// SubmitRandaoCommitmentRegistration validates a signed randao commitment
// registration, inserts it into the operations pool, and broadcasts it to
// the randao_commitment_registration gossip topic.
//
// Validation reuses the same helpers as gossip (eip8321.VerifyRandaoCommitment
// Registration and eip8321.RandaoCommitmentRegistrationSignatureBatch), so the
// two paths cannot drift.
func (vs *Server) SubmitRandaoCommitmentRegistration(
	ctx context.Context,
	req *ethpb.SignedRandaoCommitmentRegistration,
) (*emptypb.Empty, error) {
	if req == nil || req.Message == nil {
		return nil, status.Error(codes.InvalidArgument, "nil registration")
	}
	reg := req.Message

	// Cheap structural checks first.
	if len(reg.Commitment) != 32 {
		return nil, status.Error(codes.InvalidArgument, "commitment must be 32 bytes")
	}
	var commitment [32]byte
	copy(commitment[:], reg.Commitment)
	if commitment == ([32]byte{}) {
		return nil, status.Error(codes.InvalidArgument, "commitment must be non-zero")
	}
	if len(req.Signature) != 96 {
		return nil, status.Error(codes.InvalidArgument, "signature must be 96 bytes")
	}

	// Idempotent: already waiting in pool.
	if vs.RandaoCommitmentPool.ValidatorExists(reg.ValidatorIndex) {
		return &emptypb.Empty{}, nil
	}

	st, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get head state: %v", err)
	}

	// Only accept after the fork has activated.
	if st.Version() < version.Eip8321 {
		return nil, status.Error(codes.FailedPrecondition, "EIP-8321 is not active on this chain")
	}

	numValidators := uint64(st.NumValidators())
	var stored [32]byte
	if uint64(reg.ValidatorIndex) < numValidators {
		stored, err = st.RandaoCommitmentAtIndex(reg.ValidatorIndex)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not read randao commitment: %v", err)
		}
	}
	pendings, err := st.PendingRandaoCommitments()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not read pending commitments: %v", err)
	}
	pendingIdx := make([]uint64, len(pendings))
	for i, p := range pendings {
		pendingIdx[i] = uint64(p.ValidatorIndex)
	}

	// Consensus checks — same helper gossip uses.
	if err := eip8321.VerifyRandaoCommitmentRegistration(
		uint64(reg.ValidatorIndex), numValidators, commitment, stored, pendingIdx,
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid registration: %v", err)
	}

	// Signature verification via the shared batch helper (batch of 1).
	sigBatch, err := eip8321.RandaoCommitmentRegistrationSignatureBatch(st, []*ethpb.SignedRandaoCommitmentRegistration{req})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not build signature batch: %v", err)
	}
	verified, err := sigBatch.Verify()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "signature verification error: %v", err)
	}
	if !verified {
		return nil, status.Error(codes.InvalidArgument, "invalid registration signature")
	}

	// Insert + broadcast. Insert first so the pool is ready; a failed
	// broadcast is not fatal (the VC can retry, and the message may reach
	// peers via other paths).
	vs.RandaoCommitmentPool.InsertRandaoCommitment(req)
	vs.OperationNotifier.OperationFeed().Send(&feed.Event{
		Type: operation.RandaoCommitmentRegistrationReceived,
		Data: &operation.RandaoCommitmentRegistrationReceivedData{
			Registration: req,
		},
	})
	if err := vs.Broadcaster.Broadcast(ctx, p2p.RandaoCommitmentRegistrationTopicFormat, req); err != nil {
		logrus.WithError(err).WithField("validatorIndex", reg.ValidatorIndex).
			Warn("Could not broadcast randao commitment registration")
	}
	return &emptypb.Empty{}, nil
}

func (vs *Server) RandaoCommitment(
	ctx context.Context,
	req *ethpb.ValidatorIndexRequest,
) (*ethpb.RandaoCommitmentResponse, error) {
	if req == nil || len(req.PublicKey) != fieldparams.BLSPubkeyLength {
		return nil, status.Error(codes.InvalidArgument, "invalid public key")
	}
	st, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get head state: %v", err)
	}

	// Pre-EIP-8321: no randao commitments exist, return 32 zero bytes
	// (= unregistered) so the VC takes the BLS reveal path. Avoids
	// depending on state-native getters working before the fork.
	if st.Version() < version.Eip8321 {
		return &ethpb.RandaoCommitmentResponse{Commitment: make([]byte, 32)}, nil
	}

	var pk [fieldparams.BLSPubkeyLength]byte
	copy(pk[:], req.PublicKey)
	idx, ok := st.ValidatorIndexByPubkey(pk)
	if !ok {
		return nil, status.Error(codes.NotFound, "validator not found")
	}
	commitment, err := st.RandaoCommitmentAtIndex(idx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not read randao commitment: %v", err)
	}
	return &ethpb.RandaoCommitmentResponse{Commitment: commitment[:]}, nil
}