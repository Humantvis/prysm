package gloas

import (
	"bytes"
	"errors"

	"github.com/OffchainLabs/prysm/v7/crypto/bls/common"
	"lukechampine.com/blake3"
)

var (
	// ErrZeroHashChainReveal: a registered proposer sent an empty hash_chain_reveal.
	ErrZeroHashChainReveal = errors.New("hash chain reveal is zero")
	// ErrRandaoRevealNotInfinity: a registered proposer must send the G2 point at infinity as randao_reveal.
	ErrRandaoRevealNotInfinity = errors.New("randao reveal must be the G2 point at infinity for a registered proposer")
	// ErrHashChainRevealMismatch: blake3(DST + reveal) does not equal the stored commitment.
	ErrHashChainRevealMismatch = errors.New("hash chain reveal is not the preimage of the stored commitment")
	// ErrUnexpectedHashChainReveal: an unregistered proposer sent a non-empty hash_chain_reveal.
	ErrUnexpectedHashChainReveal = errors.New("unregistered proposer must not send a hash chain reveal")
)

// VerifyHashChainReveal checks the hash-chain part of verify_hash_chain_reveal,
// in the same order as the spec. The caller reads the values from the block
// body and the state.
//
// Spec (beacon-chain.md, EIP-8321):
//
//	def verify_hash_chain_reveal(state, body, proposer_index) -> None:
//	    assert body.hash_chain_reveal != Bytes32()
//	    assert body.randao_reveal == G2_POINT_AT_INFINITY
//	    commitment = state.randao_commitments[proposer_index]
//	    assert blake3(HASH_CHAIN_RANDAO_DST + body.hash_chain_reveal) == commitment
func VerifyHashChainReveal(hashChainReveal [32]byte, randaoReveal []byte, commitment [32]byte) error {
	if hashChainReveal == ([32]byte{}) {
		return ErrZeroHashChainReveal
	}
	if !bytes.Equal(randaoReveal, common.InfiniteSignature[:]) {
		return ErrRandaoRevealNotInfinity
	}
	if !VerifyHashChainLink(hashChainReveal, commitment) {
		return ErrHashChainRevealMismatch
	}
	return nil
}

// VerifyNoHashChainReveal checks the first assertion of verify_bls_randao_reveal:
// an unregistered proposer must leave hash_chain_reveal empty. The BLS
// signature itself is checked by Prysm's existing randao code.
//
//	assert body.hash_chain_reveal == Bytes32()
func VerifyNoHashChainReveal(hashChainReveal [32]byte) error {
	if hashChainReveal != ([32]byte{}) {
		return ErrUnexpectedHashChainReveal
	}
	return nil
}

// HashChainRandaoMix computes the new RANDAO mix on the hash-chain path:
//
//	mix = blake3(get_randao_mix(state, epoch) + body.hash_chain_reveal)
//
// Note: this replaces the legacy xor(mix, sha256(randao_reveal)).
func HashChainRandaoMix(currentMix []byte, hashChainReveal [32]byte) [32]byte {
	buf := make([]byte, 0, len(currentMix)+32)
	buf = append(buf, currentMix...)
	buf = append(buf, hashChainReveal[:]...)
	return blake3.Sum256(buf)
}
