package eip8321_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
)

func mustHex32(t *testing.T, s string) [32]byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 32 {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	var out [32]byte
	copy(out[:], b)
	return out
}

func seed(b byte) [32]byte {
	var s [32]byte
	for i := range s {
		s[i] = b
	}
	return s
}

func TestHashChainStep_KnownAnswer(t *testing.T) {
	c1 := eip8321.HashChainStep(seed(0x01))
	want := mustHex32(t, "b16fcaaa6cf97c506a9ca4c4e04e766cec67d5e1bd308152be59d082bb52f187")
	if c1 != want {
		t.Fatalf("c1 = %x, want %x", c1, want)
	}
}

func TestComputeHashChain_KnownAnswer(t *testing.T) {
	chain, err := eip8321.ComputeHashChain(seed(0x01), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 4 {
		t.Fatalf("len = %d, want 4", len(chain))
	}
	if chain[0] != seed(0x01) {
		t.Fatalf("chain[0] must be the secret")
	}
	want := []string{
		"b16fcaaa6cf97c506a9ca4c4e04e766cec67d5e1bd308152be59d082bb52f187",
		"9462feb62862042eefd812dec8f9904403985ae44a3780f9e88624c66b8d7d71",
		"a059408453c3255c6f700113dc68413ef6354c03dcd8af69452bee1f2c35ffef",
	}
	for i, w := range want {
		if chain[i+1] != mustHex32(t, w) {
			t.Fatalf("chain[%d] = %x, want %s", i+1, chain[i+1], w)
		}
	}
}

func TestVerifyHashChainLink(t *testing.T) {
	chain, err := eip8321.ComputeHashChain(seed(0x42), 16)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(chain); i++ {
		if !eip8321.VerifyHashChainLink(chain[i-1], chain[i]) {
			t.Fatalf("link %d does not verify against %d", i-1, i)
		}
	}
	commitment := chain[len(chain)-1]
	if eip8321.VerifyHashChainLink(chain[len(chain)-3], commitment) {
		t.Fatal("skipped link must not verify")
	}
	if eip8321.VerifyHashChainLink(commitment, commitment) {
		t.Fatal("commitment is not its own preimage")
	}
	if eip8321.VerifyHashChainLink([32]byte{}, commitment) {
		t.Fatal("zero reveal must not verify")
	}
}

func TestComputeHashChain_LengthZero(t *testing.T) {
	chain, err := eip8321.ComputeHashChain(seed(0x07), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 1 || chain[0] != seed(0x07) {
		t.Fatal("length 0 must return just the secret")
	}
}

func TestComputeHashChain_ZeroSecretRejected(t *testing.T) {
	_, err := eip8321.ComputeHashChain([32]byte{}, 4)
	if !errors.Is(err, eip8321.ErrZeroHashChainLink) {
		t.Fatalf("want ErrZeroHashChainLink, got %v", err)
	}
}
