package kv

import (
    "context"
    "encoding/binary"
    "fmt"

    fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
    "github.com/OffchainLabs/prysm/v7/monitoring/tracing/trace"
    "github.com/pkg/errors"
    bolt "go.etcd.io/bbolt"
)

var (
    // ErrRandaoHashChainAlreadySet is returned when SaveRandaoHashChain is
    // called with a different (seed, length) than the one already stored for
    // the pubkey. The stored seed anchors an on-chain commitment.
    ErrRandaoHashChainAlreadySet = errors.New("randao hash chain already set for pubkey with different seed or length")

    // ErrInvalidRandaoHashChain is returned for length < 1 or a zero seed.
    ErrInvalidRandaoHashChain = errors.New("invalid randao hash chain parameters")
)

// randaoRecordSize is seed (32) plus a big-endian uint64 length (8).
const randaoRecordSize = 32 + 8

// SaveRandaoHashChain persists the randao hash chain seed and length for a
// validator pubkey.
//
// Write-once semantics:
//   - no existing record           -> write, return nil
//   - existing record, identical   -> no-op, return nil (idempotent)
//   - existing record, different   -> ErrRandaoHashChainAlreadySet
//
// A rotated seed requires an explicit DeleteRandaoHashChain first.
func (s *Store) SaveRandaoHashChain(_ context.Context, pubKey [fieldparams.BLSPubkeyLength]byte, seed [32]byte, length uint64) error {
    if length < 1 {
        return fmt.Errorf("%w: length must be >= 1, got %d", ErrInvalidRandaoHashChain, length)
    }
    if seed == ([32]byte{}) {
        return fmt.Errorf("%w: seed must be non-zero", ErrInvalidRandaoHashChain)
    }
    rec := encodeRandaoRecord(seed, length)
    return s.db.Update(func(tx *bolt.Tx) error {
        bkt, err := tx.CreateBucketIfNotExists(randaoHashChainBucket)
        if err != nil {
            return err
        }
        if existing := bkt.Get(pubKey[:]); existing != nil {
            eSeed, eLen, err := decodeRandaoRecord(existing)
            if err != nil {
                return err
            }
            if eSeed == seed && eLen == length {
                return nil // idempotent
            }
            return ErrRandaoHashChainAlreadySet
        }
        return bkt.Put(pubKey[:], rec)
    })
}

// RandaoHashChain returns the stored seed and length for a pubkey, or
// ok=false if none is stored. Not-found is not an error.
func (s *Store) RandaoHashChain(_ context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) ([32]byte, uint64, bool, error) {
    var (
        seed   [32]byte
        length uint64
        ok     bool
    )
    err := s.db.View(func(tx *bolt.Tx) error {
        bkt := tx.Bucket(randaoHashChainBucket)
        if bkt == nil {
            return nil
        }
        raw := bkt.Get(pubKey[:])
        if raw == nil {
            return nil
        }
        var decErr error
        seed, length, decErr = decodeRandaoRecord(raw)
        if decErr != nil {
            return decErr
        }
        ok = true
        return nil
    })
    return seed, length, ok, err
}

// DeleteRandaoHashChain removes the stored randao hash chain for a pubkey.
// Idempotent: returns nil if no entry exists.
func (s *Store) DeleteRandaoHashChain(ctx context.Context, pubKey [fieldparams.BLSPubkeyLength]byte) error {
    _, span := trace.StartSpan(ctx, "validator.db.DeleteRandaoHashChain")
    defer span.End()
    return s.db.Update(func(tx *bolt.Tx) error {
        bkt := tx.Bucket(randaoHashChainBucket)
        if bkt == nil {
            return nil
        }
        return bkt.Delete(pubKey[:])
    })
}

func encodeRandaoRecord(seed [32]byte, length uint64) []byte {
    rec := make([]byte, randaoRecordSize)
    copy(rec[:32], seed[:])
    binary.BigEndian.PutUint64(rec[32:], length)
    return rec
}

func decodeRandaoRecord(rec []byte) ([32]byte, uint64, error) {
    var seed [32]byte
    if len(rec) != randaoRecordSize {
        return seed, 0, fmt.Errorf("randao hash chain record has length %d, want %d", len(rec), randaoRecordSize)
    }
    copy(seed[:], rec[:32])
    return seed, binary.BigEndian.Uint64(rec[32:]), nil
}