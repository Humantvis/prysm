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

// ProcessRandaoEip8321 is process_randao for the EIP-8321 fork.
//
// A proposer with a registered commitment reveals the next link of its hash
// chain; everyone else keeps using the legacy BLS reveal. When verifyBLS is
// false, the legacy BLS signature is not checked here because the caller
// verifies it later in a signature batch (see RandaoSignatureBatch).
//
// Spec (beacon-chain.md, EIP-8321):
//
//	def process_randao(state: BeaconState, body: BeaconBlockBody) -> None:
//	    epoch = get_current_epoch(state)
//	    proposer_index = get_beacon_proposer_index(state)
//	    if state.randao_commitments[proposer_index] != UNSET_RANDAO_COMMITMENT:
//	        verify_hash_chain_reveal(state, body, proposer_index)
//	        mix = blake3(get_randao_mix(state, epoch) + body.hash_chain_reveal)
//	        state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR] = mix
//	        state.randao_commitments[proposer_index] = body.hash_chain_reveal
//	    else:
//	        verify_bls_randao_reveal(state, body, proposer_index)
//	        mix = xor(get_randao_mix(state, epoch), sha256(body.randao_reveal))
//	        state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR] = mix
func ProcessRandaoEip8321(
	ctx context.Context,
	st state.BeaconState,
	body interfaces.ReadOnlyBeaconBlockBody,
	verifyBLS bool,
) (state.BeaconState, error) {
	proposerIdx, err := helpers.BeaconProposerIndex(ctx, st)
	if err != nil {
		return nil, errors.Wrap(err, "could not get beacon proposer index")
	}
	commitment, err := st.RandaoCommitmentAtIndex(proposerIdx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get proposer randao commitment")
	}
	hashChainReveal, err := body.HashChainReveal()
	if err != nil {
		return nil, errors.Wrap(err, "could not get hash chain reveal")
	}
	randaoReveal := body.RandaoReveal()

	// Registered proposer: hash-chain path.
	if commitment != ([32]byte{}) {
		if err := eip8321.VerifyHashChainReveal(hashChainReveal, randaoReveal[:], commitment); err != nil {
			return nil, errors.Wrap(err, "could not verify hash chain reveal")
		}
		epoch := slots.ToEpoch(st.Slot())
		mixIndex := uint64(epoch % params.BeaconConfig().EpochsPerHistoricalVector)
		currentMix, err := st.RandaoMixAtIndex(mixIndex)
		if err != nil {
			return nil, err
		}
		if err := st.UpdateRandaoMixesAtIndex(mixIndex, eip8321.HashChainRandaoMix(currentMix, hashChainReveal)); err != nil {
			return nil, err
		}
		// The reveal becomes the new commitment, so the next proposal must reveal one link deeper.
		if err := st.UpdateRandaoCommitmentAtIndex(proposerIdx, hashChainReveal); err != nil {
			return nil, err
		}
		return st, nil
	}

	// Unregistered proposer: legacy BLS path.
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

// RandaoUsesHashChain reports whether the block revealed its RANDAO through the
// hash chain. Such blocks carry the G2 point at infinity as randao_reveal, so
// they must not be added to the BLS signature batch.
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
