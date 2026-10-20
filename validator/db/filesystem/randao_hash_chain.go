package filesystem

import (
	"bytes"
	"context"
	"fmt"

	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

var (
	// ErrRandaoHashChainAlreadySet is returned when SaveRandaoHashChain is
	// called with a different (seed, length) than the one already stored for
	// the pubkey. The stored seed anchors an on-chain commitment; silently
	// overwriting it would surface as client.ErrCommitmentMismatch at the
	// next proposal.
	//
	// Mirrors kv.ErrRandaoHashChainAlreadySet.
	ErrRandaoHashChainAlreadySet = errors.New("randao hash chain already set for pubkey with different seed or length")

	// ErrInvalidRandaoHashChain is returned for length < 1 or a zero seed.
	// Mirrors kv.ErrInvalidRandaoHashChain.
	ErrInvalidRandaoHashChain = errors.New("invalid randao hash chain parameters")
)

// SaveRandaoHashChain persists the randao hash chain seed and length for a
// validator pubkey.
//
// Write-once semantics:
//   - no existing entry           -> write, return nil
//   - existing entry, identical   -> no-op, return nil (idempotent)
//   - existing entry, different   -> ErrRandaoHashChainAlreadySet
//
// A rotated seed requires an explicit DeleteRandaoHashChain first (only safe
// while the validator is unregistered on-chain).
func (s *Store) SaveRandaoHashChain(_ context.Context, pubKey [fieldparams.BLSPubkeyLength]byte, seed [32]byte, length uint64) error {
	if length < 1 {
		return fmt.Errorf("%w: length must be >= 1, got %d", ErrInvalidRandaoHashChain, length)
	}
	if seed == ([32]byte{}) {
		return fmt.Errorf("%w: seed must be non-zero", ErrInvalidRandaoHashChain)
	}

	configuration, err := s.configuration()
	if err != nil {
		return errors.Wrap(err, "could not get configuration")
	}
	if configuration == nil {
		configuration = &Configuration{}
	}
	if configuration.RandaoHashChains == nil {
		configuration.RandaoHashChains = make(map[string]*RandaoHashChainEntry)
	}

	key := hexutil.Encode(pubKey[:])
	if existing, ok := configuration.RandaoHashChains[key]; ok && existing != nil {
		existingSeed, err := hexutil.Decode(existing.Seed)
		if err != nil {
			return errors.Wrap(err, "could not decode existing randao hash chain seed")
		}
		if bytes.Equal(existingSeed, seed[:]) && existing.Length == length {
			return nil // idempotent
		}
		return ErrRandaoHashChainAlreadySet
	}

	configuration.RandaoHashChains[key] = &RandaoHashChainEntry{
		Seed:   hexutil.Encode(seed[:]),
		Length: length,
	}
	if err := s.saveConfiguration(configuration); err != nil {
		return errors.Wrap(err, "could not save configuration")
	}
	return nil
}

// RandaoHashChain returns the stored seed and length for a pubkey, or
// ok=false if none is stored. Not-found is not an error.
func (s *Store) RandaoHashChain(_ context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) (seed [32]byte, length uint64, ok bool, err error) {
	configuration, err := s.configuration()
	if err != nil {
		return seed, 0, false, errors.Wrap(err, "could not get configuration")
	}
	if configuration == nil || configuration.RandaoHashChains == nil {
		return seed, 0, false, nil
	}

	entry, found := configuration.RandaoHashChains[hexutil.Encode(pubKey[:])]
	if !found || entry == nil {
		return seed, 0, false, nil
	}

	seedBytes, err := hexutil.Decode(entry.Seed)
	if err != nil {
		return seed, 0, false, errors.Wrap(err, "could not decode randao hash chain seed")
	}
	if len(seedBytes) != 32 {
		return seed, 0, false, fmt.Errorf("randao hash chain seed has length %d, want 32", len(seedBytes))
	}
	copy(seed[:], seedBytes)
	return seed, entry.Length, true, nil
}

// DeleteRandaoHashChain removes the stored randao hash chain for a pubkey.
//
// Idempotent: returns nil if no entry exists. Call only after confirming the
// validator is unregistered on-chain — this is the recovery path out of
// client.ErrHashChainExhausted, which otherwise makes re-registration fail
// with ErrRandaoHashChainAlreadySet.
func (s *Store) DeleteRandaoHashChain(_ context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) error {
	configuration, err := s.configuration()
	if err != nil {
		return errors.Wrap(err, "could not get configuration")
	}
	if configuration == nil || configuration.RandaoHashChains == nil {
		return nil
	}

	key := hexutil.Encode(pubKey[:])
	if _, ok := configuration.RandaoHashChains[key]; !ok {
		return nil
	}

	delete(configuration.RandaoHashChains, key)
	// Collapse to nil so yaml.v3's omitempty drops the key entirely rather
	// than emitting `randaoHashChains: {}`.
	if len(configuration.RandaoHashChains) == 0 {
		configuration.RandaoHashChains = nil
	}

	if err := s.saveConfiguration(configuration); err != nil {
		return errors.Wrap(err, "could not save configuration")
	}
	return nil
}