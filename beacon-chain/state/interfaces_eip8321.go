package state

import (
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
)

type readOnlyEip8321Fields interface {
	RandaoCommitmentAtIndex(primitives.ValidatorIndex) ([32]byte, error)
	RandaoCommitments() ([][]byte, error)
	PendingRandaoCommitments() ([]*ethpb.PendingRandaoCommitment, error)
}
type writeOnlyEip8321Fields interface {
	SetRandaoCommitments([][]byte) error
	UpdateRandaoCommitmentAtIndex(primitives.ValidatorIndex, [32]byte) error
	AppendRandaoCommitment([32]byte) error
	SetPendingRandaoCommitments([]*ethpb.PendingRandaoCommitment) error
	AppendPendingRandaoCommitment(*ethpb.PendingRandaoCommitment) error
}
