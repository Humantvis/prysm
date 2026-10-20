package validator

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
)

// setRandaoCommitmentRegistrations fills the block body with valid pending
// randao commitment registrations from the operations pool.
func (vs *Server) setRandaoCommitmentRegistrations(blk interfaces.SignedBeaconBlock, headState state.BeaconState) {
	if blk.Version() < version.Eip8321 {
		return
	}
	// Start from an empty list so the body is valid even if the pool read fails.
	if err := blk.SetRandaoCommitmentRegistrations([]*ethpb.SignedRandaoCommitmentRegistration{}); err != nil {
		log.WithError(err).Error("Could not initialize randao commitment registrations in block")
		return
	}
	regs, err := vs.RandaoCommitmentPool.RandaoCommitmentsForInclusion(headState)
	if err != nil {
		log.WithError(err).Error("Could not get randao commitment registrations")
		return
	}
	if err := blk.SetRandaoCommitmentRegistrations(regs); err != nil {
		log.WithError(err).Error("Could not set randao commitment registrations")
	}
}