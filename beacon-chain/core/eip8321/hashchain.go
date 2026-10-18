package eip8321

import (
	"errors"

	"lukechampine.com/blake3"
)

var HashChainRandaoDST = []byte("HASH_CHAIN_RANDAO")

var ErrZeroHashChainLink = errors.New("hash chain contains a zero link")

// Hash one step of the hash chain
func HashChainStep(prev [32]byte) [32]byte {
	buf := make([]byte, 0, len(HashChainRandaoDST)+32)
	buf = append(buf, HashChainRandaoDST...)
	buf = append(buf, prev[:]...)
	return blake3.Sum256(buf)
}

// Build the whole chain
func ComputeHashChain(chainSecret [32]byte, length uint64) ([][32]byte, error) {
	chain := make([][32]byte, 0, length+1)
	chain = append(chain, chainSecret)
	for i := uint64(0); i < length; i++ {
		chain = append(chain, HashChainStep(chain[len(chain)-1]))
	}
	for _, link := range chain {
		if link == ([32]byte{}) {
			return nil, ErrZeroHashChainLink
		}
	}
	return chain, nil
}

// Check if hashing the reveal gives us the commitment
func VerifyHashChainLink(reveal, commitment [32]byte) bool {
	return HashChainStep(reveal) == commitment
}
