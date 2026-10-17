package state_native

import (
	"encoding/binary"
	"fmt"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native/types"
	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	"github.com/OffchainLabs/prysm/v7/encoding/ssz"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
)

func (b *BeaconState) RandaoCommitmentAtIndex(index primitives.ValidatorIndex) ([32]byte, error) {
	if b.version < version.Eip8321 {
		return [32]byte{}, errNotSupported("RandaoCommitmentAtIndex", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	if uint64(index) >= uint64(len(b.randaoCommitments)) {
		return [32]byte{}, fmt.Errorf("randao commitment index %d out of range", index)
	}
	var result [32]byte
	copy(result[:], b.randaoCommitments[index])
	return result, nil
}

func (b *BeaconState) RandaoCommitments() ([][]byte, error) {
	if b.version < version.Eip8321 {
		return nil, errNotSupported("RandaoCommitments", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.randaoCommitmentsVal(), nil
}

func (b *BeaconState) PendingRandaoCommitments() ([]*ethpb.PendingRandaoCommitment, error) {
	if b.version < version.Eip8321 {
		return nil, errNotSupported("PendingRandaoCommitments", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	return ethpb.CopySlice(b.pendingRandaoCommitments), nil
}

func (b *BeaconState) SetRandaoCommitments(commitments [][]byte) error {
	if b.version < version.Eip8321 {
		return errNotSupported("SetRandaoCommitments", b.version)
	}
	if uint64(len(commitments)) > fieldparams.ValidatorRegistryLimit {
		return fmt.Errorf("too many randao commitments")
	}
	for _, commitment := range commitments {
		if len(commitment) != 32 {
			return fmt.Errorf("randao commitment must be 32 bytes")
		}
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if len(commitments) != b.validatorsLen() {
		return fmt.Errorf("randao commitments length does not match validators")
	}
	b.randaoCommitments = bytesutil.SafeCopy2dBytes(commitments)
	b.markFieldAsDirty(types.RandaoCommitments)
	return nil
}

func (b *BeaconState) UpdateRandaoCommitmentAtIndex(index primitives.ValidatorIndex, commitment [32]byte) error {
	if b.version < version.Eip8321 {
		return errNotSupported("UpdateRandaoCommitmentAtIndex", b.version)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if uint64(index) >= uint64(len(b.randaoCommitments)) {
		return fmt.Errorf("randao commitment index %d out of range", index)
	}
	b.randaoCommitments[index] = bytesutil.SafeCopyBytes(commitment[:])
	b.markFieldAsDirty(types.RandaoCommitments)
	return nil
}

func (b *BeaconState) AppendRandaoCommitment(commitment [32]byte) error {
	if b.version < version.Eip8321 {
		return errNotSupported("AppendRandaoCommitment", b.version)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if uint64(len(b.randaoCommitments)) >= fieldparams.ValidatorRegistryLimit {
		return fmt.Errorf("too many randao commitments")
	}
	b.randaoCommitments = append(b.randaoCommitments, bytesutil.SafeCopyBytes(commitment[:]))
	b.markFieldAsDirty(types.RandaoCommitments)
	return nil
}

func (b *BeaconState) SetPendingRandaoCommitments(pending []*ethpb.PendingRandaoCommitment) error {
	if b.version < version.Eip8321 {
		return errNotSupported("SetPendingRandaoCommitments", b.version)
	}
	for _, entry := range pending {
		if entry == nil || len(entry.Commitment) != 32 {
			return fmt.Errorf("invalid pending randao commitment")
		}
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.pendingRandaoCommitments = ethpb.CopySlice(pending)
	b.markFieldAsDirty(types.PendingRandaoCommitments)
	return nil
}

func (b *BeaconState) AppendPendingRandaoCommitment(pending *ethpb.PendingRandaoCommitment) error {
	if b.version < version.Eip8321 {
		return errNotSupported("AppendPendingRandaoCommitment", b.version)
	}
	if pending == nil || len(pending.Commitment) != 32 {
		return fmt.Errorf("invalid pending randao commitment")
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.pendingRandaoCommitments = append(b.pendingRandaoCommitments, pending.Copy())
	b.markFieldAsDirty(types.PendingRandaoCommitments)
	return nil
}

func (b *BeaconState) randaoCommitmentsVal() [][]byte {
	return bytesutil.SafeCopy2dBytes(b.randaoCommitments)
}

func (b *BeaconState) randaoCommitmentsRoot() ([32]byte, error) {
	chunks := make([][32]byte, len(b.randaoCommitments))
	for i, c := range b.randaoCommitments {
		if len(c) != 32 {
			return [32]byte{}, fmt.Errorf("randao commitment %d must be 32 bytes", i)
		}
		copy(chunks[i][:], c)
	}
	root := ssz.MerkleizeProgressiveChunks(chunks)
	var length [32]byte
	binary.LittleEndian.PutUint64(length[:], uint64(len(chunks)))
	return ssz.MixInLength(root, length[:]), nil
}
