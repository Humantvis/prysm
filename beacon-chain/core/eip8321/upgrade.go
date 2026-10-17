package eip8321

import (
	"fmt"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/time"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/config/params"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
)

// UpgradeToEip8321 upgrades a Heze state to the EIP-8321 feature fork.
func UpgradeToEip8321(st state.BeaconState) (state.BeaconState, error) {
	if st == nil || st.IsNil() || st.Version() != version.Heze {
		return nil, fmt.Errorf("EIP-8321 upgrade requires a Heze state")
	}
	pre, err := state_native.ProtobufBeaconStateHeze(st.ToProto())
	if err != nil {
		return nil, err
	}
	commitments := make([][]byte, len(pre.Validators))
	for i := range commitments {
		commitments[i] = make([]byte, 32)
	}
	post := &ethpb.BeaconStateEip8321{
		GenesisTime:                   pre.GenesisTime,
		GenesisValidatorsRoot:         pre.GenesisValidatorsRoot,
		Slot:                          pre.Slot,
		LatestBlockHeader:             pre.LatestBlockHeader,
		BlockRoots:                    pre.BlockRoots,
		StateRoots:                    pre.StateRoots,
		HistoricalRoots:               pre.HistoricalRoots,
		Validators:                    pre.Validators,
		Balances:                      pre.Balances,
		RandaoMixes:                   pre.RandaoMixes,
		Slashings:                     pre.Slashings,
		PreviousEpochParticipation:    pre.PreviousEpochParticipation,
		CurrentEpochParticipation:     pre.CurrentEpochParticipation,
		JustificationBits:             pre.JustificationBits,
		PreviousJustifiedCheckpoint:   pre.PreviousJustifiedCheckpoint,
		CurrentJustifiedCheckpoint:    pre.CurrentJustifiedCheckpoint,
		FinalizedCheckpoint:           pre.FinalizedCheckpoint,
		InactivityScores:              pre.InactivityScores,
		CurrentSyncCommittee:          pre.CurrentSyncCommittee,
		NextSyncCommittee:             pre.NextSyncCommittee,
		LatestBlockHash:               pre.LatestBlockHash,
		NextWithdrawalIndex:           pre.NextWithdrawalIndex,
		NextWithdrawalValidatorIndex:  pre.NextWithdrawalValidatorIndex,
		HistoricalSummaries:           pre.HistoricalSummaries,
		DepositBalanceToConsume:       pre.DepositBalanceToConsume,
		ExitBalanceToConsume:          pre.ExitBalanceToConsume,
		EarliestExitEpoch:             pre.EarliestExitEpoch,
		ConsolidationBalanceToConsume: pre.ConsolidationBalanceToConsume,
		EarliestConsolidationEpoch:    pre.EarliestConsolidationEpoch,
		PendingDeposits:               pre.PendingDeposits,
		PendingPartialWithdrawals:     pre.PendingPartialWithdrawals,
		PendingConsolidations:         pre.PendingConsolidations,
		ProposerLookahead:             pre.ProposerLookahead,
		Builders:                      pre.Builders,
		NextWithdrawalBuilderIndex:    pre.NextWithdrawalBuilderIndex,
		ExecutionPayloadAvailability:  pre.ExecutionPayloadAvailability,
		BuilderPendingPayments:        pre.BuilderPendingPayments,
		BuilderPendingWithdrawals:     pre.BuilderPendingWithdrawals,
		LatestExecutionPayloadBid:     pre.LatestExecutionPayloadBid,
		PayloadExpectedWithdrawals:    pre.PayloadExpectedWithdrawals,
		PtcWindow:                     pre.PtcWindow,
		Fork:                          &ethpb.Fork{PreviousVersion: pre.Fork.CurrentVersion, CurrentVersion: params.BeaconConfig().Eip8321ForkVersion, Epoch: time.CurrentEpoch(st)},
		RandaoCommitments:             commitments,
		PendingRandaoCommitments:      []*ethpb.PendingRandaoCommitment{},
	}
	return state_native.InitializeFromProtoUnsafeEip8321(post)
}
