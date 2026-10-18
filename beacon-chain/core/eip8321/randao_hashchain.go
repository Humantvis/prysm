package eip8321

import (
	"bytes"
	"errors"

	"github.com/OffchainLabs/prysm/v7/crypto/bls/common"
	"lukechampine.com/blake3"
)

var (
	ErrZeroHashChainReveal       = errors.New("hash chain reveal is zero")
	ErrRandaoRevealNotInfinity   = errors.New("randao reveal must be the G2 point at infinity for a registered proposer")
	ErrHashChainRevealMismatch   = errors.New("hash chain reveal is not the preimage of the stored commitment")
	ErrUnexpectedHashChainReveal = errors.New("unregistered proposer must not send a hash chain reveal")
)

// Hash chain reveal not zero, randao reveal infinity, reveal hashes to the commitment
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

// unregistered proposer - hash_chain_reveal is made of all zeros
func VerifyNoHashChainReveal(hashChainReveal [32]byte) error {
	if hashChainReveal != ([32]byte{}) {
		return ErrUnexpectedHashChainReveal
	}
	return nil
}

// Create new mix
func HashChainRandaoMix(currentMix []byte, hashChainReveal [32]byte) [32]byte {
	buf := make([]byte, 0, len(currentMix)+32)
	buf = append(buf, currentMix...)
	buf = append(buf, hashChainReveal[:]...)
	return blake3.Sum256(buf)
}
