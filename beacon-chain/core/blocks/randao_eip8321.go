package blocks

import (
	"context"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/helpers"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
	"github.com/OffchainLabs/prysm/v7/time/slots"
	"github.com/pkg/errors"
)

func ProcessRandaoEip8321(
	ctx context.Context,
	st state.BeaconState,
	body interfaces.ReadOnlyBeaconBlockBody,
	verifyBLS bool,
) (state.BeaconState, error) {
	// Who and are they registered
	proposerIdx, err := helpers.BeaconProposerIndex(ctx, st)
	if err != nil {
		return nil, errors.Wrap(err, "could not get beacon proposer index")
	}
	commitment, err := st.RandaoCommitmentAtIndex(proposerIdx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get proposer randao commitment")
	}
	// Read reveals
	hashChainReveal, err := body.HashChainReveal()
	if err != nil {
		return nil, errors.Wrap(err, "could not get hash chain reveal")
	}
	randaoReveal := body.RandaoReveal()

	if commitment != ([32]byte{}) {
		// Runs checks from randao_hashchain.go
		if err := eip8321.VerifyHashChainReveal(hashChainReveal, randaoReveal[:], commitment); err != nil {
			return nil, errors.Wrap(err, "could not verify hash chain reveal")
		}
		// Slot number to an epoch
		epoch := slots.ToEpoch(st.Slot())
		mixIndex := uint64(epoch % params.BeaconConfig().EpochsPerHistoricalVector)
		currentMix, err := st.RandaoMixAtIndex(mixIndex)
		if err != nil {
			return nil, err
		}
		// Compute the new mix
		if err := st.UpdateRandaoMixesAtIndex(mixIndex, eip8321.HashChainRandaoMix(currentMix, hashChainReveal)); err != nil {
			return nil, err
		}
		if err := st.UpdateRandaoCommitmentAtIndex(proposerIdx, hashChainReveal); err != nil {
			return nil, err
		}
		// Reveal is the new commitment
		return st, nil
	}

	// Runs checks from randao_hashchain.go
	if err := eip8321.VerifyNoHashChainReveal(hashChainReveal); err != nil {
		return nil, err
	}
	if verifyBLS {
		buf, proposerPub, domain, err := randaoSigningData(ctx, st)
		if err != nil {
			return nil, err
		}
		if err := verifySignature(buf, proposerPub, randaoReveal[:], domain); err != nil {
			return nil, errors.Wrap(err, "could not verify block randao")
		}
	}
	return ProcessRandaoNoVerify(st, randaoReveal[:])
}

// Does this block use the hash chain
func RandaoUsesHashChain(body interfaces.ReadOnlyBeaconBlockBody) (bool, error) {
	if body.Version() < version.Eip8321 {
		return false, nil
	}
	reveal, err := body.HashChainReveal()
	if err != nil {
		return false, err
	}
	return reveal != ([32]byte{}), nil
}
