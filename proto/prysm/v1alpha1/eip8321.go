package eth

import (
	"slices"

	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
)

func CopySignedBeaconBlockEip8321(sb *SignedBeaconBlockEip8321) *SignedBeaconBlockEip8321 {
	if sb == nil {
		return nil
	}
	return &SignedBeaconBlockEip8321{
		Block:     copyBeaconBlockEip8321(sb.Block),
		Signature: bytesutil.SafeCopyBytes(sb.Signature),
	}
}

func copyBeaconBlockEip8321(b *BeaconBlockEip8321) *BeaconBlockEip8321 {
	if b == nil {
		return nil
	}
	return &BeaconBlockEip8321{
		Slot:          b.Slot,
		ProposerIndex: b.ProposerIndex,
		ParentRoot:    bytesutil.SafeCopyBytes(b.ParentRoot),
		StateRoot:     bytesutil.SafeCopyBytes(b.StateRoot),
		Body:          copyBeaconBlockBodyEip8321(b.Body),
	}
}

func copyBeaconBlockBodyEip8321(body *BeaconBlockBodyEip8321) *BeaconBlockBodyEip8321 {
	if body == nil {
		return nil
	}
	copied := &BeaconBlockBodyEip8321{
		HashChainReveal:               bytesutil.SafeCopyBytes(body.HashChainReveal),
		RandaoCommitmentRegistrations: CopySlice(body.RandaoCommitmentRegistrations),
		RandaoReveal:                  bytesutil.SafeCopyBytes(body.RandaoReveal),
		Graffiti:                      bytesutil.SafeCopyBytes(body.Graffiti),
		ProposerSlashings:             CopySlice(body.ProposerSlashings),
		AttesterSlashings:             CopySlice(body.AttesterSlashings),
		Attestations:                  CopySlice(body.Attestations),
		VoluntaryExits:                CopySlice(body.VoluntaryExits),
		BlsToExecutionChanges:         CopySlice(body.BlsToExecutionChanges),
		SignedExecutionPayloadBid:     body.SignedExecutionPayloadBid.Copy(),
		PayloadAttestations:           copyPayloadAttestations(body.PayloadAttestations),
		ParentExecutionRequests:       CopyExecutionRequestsGloas(body.ParentExecutionRequests),
	}
	if body.SyncAggregate != nil {
		copied.SyncAggregate = body.SyncAggregate.Copy()
	}
	return copied
}

func (r *RandaoCommitmentRegistration) Copy() *RandaoCommitmentRegistration {
	if r == nil {
		return nil
	}
	return &RandaoCommitmentRegistration{ValidatorIndex: r.ValidatorIndex, Commitment: bytesutil.SafeCopyBytes(r.Commitment)}
}

func (r *SignedRandaoCommitmentRegistration) Copy() *SignedRandaoCommitmentRegistration {
	if r == nil {
		return nil
	}
	return &SignedRandaoCommitmentRegistration{Message: r.Message.Copy(), Signature: bytesutil.SafeCopyBytes(r.Signature)}
}

func (p *PendingRandaoCommitment) Copy() *PendingRandaoCommitment {
	if p == nil {
		return nil
	}
	return &PendingRandaoCommitment{ValidatorIndex: p.ValidatorIndex, Commitment: bytesutil.SafeCopyBytes(p.Commitment), ActivationEpoch: p.ActivationEpoch}
}

func (st *BeaconStateEip8321) Copy() *BeaconStateEip8321 {
	if st == nil {
		return nil
	}
	cp := *st
	cp.GenesisValidatorsRoot = bytesutil.SafeCopyBytes(st.GenesisValidatorsRoot)
	cp.Fork = st.Fork.Copy()
	cp.LatestBlockHeader = st.LatestBlockHeader.Copy()
	cp.BlockRoots = bytesutil.SafeCopy2dBytes(st.BlockRoots)
	cp.StateRoots = bytesutil.SafeCopy2dBytes(st.StateRoots)
	cp.HistoricalRoots = bytesutil.SafeCopy2dBytes(st.HistoricalRoots)
	cp.Validators = CopySlice(st.Validators)
	cp.Balances = slices.Clone(st.Balances)
	cp.RandaoMixes = bytesutil.SafeCopy2dBytes(st.RandaoMixes)
	cp.Slashings = slices.Clone(st.Slashings)
	cp.PreviousEpochParticipation = bytesutil.SafeCopyBytes(st.PreviousEpochParticipation)
	cp.CurrentEpochParticipation = bytesutil.SafeCopyBytes(st.CurrentEpochParticipation)
	cp.JustificationBits = bytesutil.SafeCopyBytes(st.JustificationBits)
	cp.PreviousJustifiedCheckpoint = st.PreviousJustifiedCheckpoint.Copy()
	cp.CurrentJustifiedCheckpoint = st.CurrentJustifiedCheckpoint.Copy()
	cp.FinalizedCheckpoint = st.FinalizedCheckpoint.Copy()
	cp.InactivityScores = slices.Clone(st.InactivityScores)
	cp.CurrentSyncCommittee = st.CurrentSyncCommittee.Copy()
	cp.NextSyncCommittee = st.NextSyncCommittee.Copy()
	cp.LatestBlockHash = bytesutil.SafeCopyBytes(st.LatestBlockHash)
	cp.HistoricalSummaries = CopySlice(st.HistoricalSummaries)
	cp.PendingDeposits = CopySlice(st.PendingDeposits)
	cp.PendingPartialWithdrawals = CopySlice(st.PendingPartialWithdrawals)
	cp.PendingConsolidations = CopySlice(st.PendingConsolidations)
	cp.ProposerLookahead = slices.Clone(st.ProposerLookahead)
	cp.Builders = CopySlice(st.Builders)
	cp.ExecutionPayloadAvailability = bytesutil.SafeCopyBytes(st.ExecutionPayloadAvailability)
	cp.BuilderPendingPayments = CopySlice(st.BuilderPendingPayments)
	cp.BuilderPendingWithdrawals = CopySlice(st.BuilderPendingWithdrawals)
	cp.LatestExecutionPayloadBid = st.LatestExecutionPayloadBid.Copy()
	cp.PayloadExpectedWithdrawals = CopySlice(st.PayloadExpectedWithdrawals)
	cp.PtcWindow = CopyPTCWindow(st.PtcWindow)
	cp.RandaoCommitments = bytesutil.SafeCopy2dBytes(st.RandaoCommitments)
	cp.PendingRandaoCommitments = CopySlice(st.PendingRandaoCommitments)
	return &cp
}
