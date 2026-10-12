package gloas

import (
	"errors"

	"lukechampine.com/blake3"
)

// HashChainRandaoDST is the domain separation tag prepended to every link
// before hashing (spec constant HASH_CHAIN_RANDAO_DST, EIP-8321).
var HashChainRandaoDST = []byte("HASH_CHAIN_RANDAO")

// ErrZeroHashChainLink is returned when a generated link equals the all-zero
// value, which the protocol reserves as UNSET_RANDAO_COMMITMENT.
var ErrZeroHashChainLink = errors.New("hash chain contains a zero link")

// HashChainStep computes the next link of a hash chain:
//
//	blake3(HASH_CHAIN_RANDAO_DST + prev)
func HashChainStep(prev [32]byte) [32]byte {
	buf := make([]byte, 0, len(HashChainRandaoDST)+32)
	buf = append(buf, HashChainRandaoDST...)
	buf = append(buf, prev[:]...)
	return blake3.Sum256(buf)
}

// ComputeHashChain returns the chain [c_0, ..., c_length] generated from
// chainSecret, where c_0 is the secret itself and c_length is the commitment.
//
// Spec (validator.md, EIP-8321):
//
//	def compute_hash_chain(chain_secret: Bytes32, length: Uint64) -> Sequence[Bytes32]:
//	    chain = [chain_secret]
//	    for _ in range(length):
//	        chain.append(blake3(HASH_CHAIN_RANDAO_DST + chain[-1]))
//	    assert all(value != UNSET_RANDAO_COMMITMENT for value in chain)
//	    return chain
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

// VerifyHashChainLink reports whether reveal is the preimage of commitment,
// i.e. blake3(HASH_CHAIN_RANDAO_DST + reveal) == commitment.
func VerifyHashChainLink(reveal, commitment [32]byte) bool {
	return HashChainStep(reveal) == commitment
}