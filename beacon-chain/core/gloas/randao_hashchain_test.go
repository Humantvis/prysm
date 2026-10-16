package gloas_test

import (
	"errors"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/gloas"
	"github.com/OffchainLabs/prysm/v7/crypto/bls/common"
)

// Test names follow the official EIP-8321 test vectors in
// reftests/<preset>/eip8321/operations/randao/pyspec_tests/.

func testChain(t *testing.T) [][32]byte {
	t.Helper()
	chain, err := gloas.ComputeHashChain(seed(0x11), 8)
	if err != nil {
		t.Fatal(err)
	}
	return chain
}

func TestVerifyHashChainReveal(t *testing.T) {
	chain := testChain(t)
	commitment := chain[8]
	infinity := common.InfiniteSignature[:]
	nonEmptyBLS := make([]byte, 96)
	nonEmptyBLS[0] = 0xa0

	tests := []struct {
		name         string
		reveal       [32]byte
		randaoReveal []byte
		commitment   [32]byte
		wantErr      error
	}{
		{"hash_chain_reveal", chain[7], infinity, commitment, nil},
		{"invalid_hash_chain_reveal_not_preimage", seed(0x99), infinity, commitment, gloas.ErrHashChainRevealMismatch},
		{"invalid_hash_chain_reveal_skips_a_link", chain[6], infinity, commitment, gloas.ErrHashChainRevealMismatch},
		{"invalid_zero_hash_chain_reveal", [32]byte{}, infinity, commitment, gloas.ErrZeroHashChainReveal},
		{"invalid_registered_proposer_with_nonempty_bls_reveal", chain[7], nonEmptyBLS, commitment, gloas.ErrRandaoRevealNotInfinity},
		{"invalid_registered_proposer_with_empty_bytes_bls_reveal", chain[7], make([]byte, 96), commitment, gloas.ErrRandaoRevealNotInfinity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gloas.VerifyHashChainReveal(tt.reveal, tt.randaoReveal, tt.commitment)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// hash_chain_reveal_consumes_one_link_per_block: after each accepted reveal the
// reveal becomes the new commitment, so the next block must reveal one link deeper.
func TestVerifyHashChainReveal_ConsumesOneLinkPerBlock(t *testing.T) {
	chain := testChain(t)
	infinity := common.InfiniteSignature[:]
	commitment := chain[8]
	for i := 7; i >= 1; i-- {
		if err := gloas.VerifyHashChainReveal(chain[i], infinity, commitment); err != nil {
			t.Fatalf("reveal of link %d: %v", i, err)
		}
		// Replaying the same reveal after it was consumed must fail.
		if err := gloas.VerifyHashChainReveal(chain[i], infinity, chain[i]); !errors.Is(err, gloas.ErrHashChainRevealMismatch) {
			t.Fatalf("replay of link %d: got %v", i, err)
		}
		commitment = chain[i] // what process_randao stores
	}
}

func TestVerifyNoHashChainReveal(t *testing.T) {
	// unregistered_proposer_uses_bls_reveal
	if err := gloas.VerifyNoHashChainReveal([32]byte{}); err != nil {
		t.Fatalf("empty reveal must pass, got %v", err)
	}
	// invalid_unregistered_proposer_with_hash_chain_reveal
	if err := gloas.VerifyNoHashChainReveal(seed(0x01)); !errors.Is(err, gloas.ErrUnexpectedHashChainReveal) {
		t.Fatalf("got %v", err)
	}
}

func TestHashChainRandaoMix_KnownAnswer(t *testing.T) {
	// blake3(0xaa*32 + c1), computed with the Python blake3 package.
	mix := make([]byte, 32)
	for i := range mix {
		mix[i] = 0xaa
	}
	c1 := mustHex32(t, "b16fcaaa6cf97c506a9ca4c4e04e766cec67d5e1bd308152be59d082bb52f187")
	got := gloas.HashChainRandaoMix(mix, c1)
	want := mustHex32(t, "a3d075e7c1d7667b90b0c0e8ee7eaeae75f631c152df22382c56e4312e66fcaf")
	if got != want {
		t.Fatalf("mix = %x, want %x", got, want)
	}
}
