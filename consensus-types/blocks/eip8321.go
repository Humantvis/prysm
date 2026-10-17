package blocks

import (
	"fmt"

	consensus_types "github.com/OffchainLabs/prysm/v7/consensus-types"
	eth "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/runtime/version"
)

func (b *BeaconBlockBody) HashChainReveal() ([32]byte, error) {
	if b.version < version.Eip8321 {
		return [32]byte{}, consensus_types.ErrNotSupported("HashChainReveal", b.version)
	}
	return b.hashChainReveal, nil
}

func (b *BeaconBlockBody) RandaoCommitmentRegistrations() ([]*eth.SignedRandaoCommitmentRegistration, error) {
	if b.version < version.Eip8321 {
		return nil, consensus_types.ErrNotSupported("RandaoCommitmentRegistrations", b.version)
	}
	return b.randaoCommitmentRegistrations, nil
}

func (b *SignedBeaconBlock) SetHashChainReveal(reveal []byte) error {
	if b.version < version.Eip8321 {
		return consensus_types.ErrNotSupported("SetHashChainReveal", b.version)
	}
	if len(reveal) != 32 {
		return fmt.Errorf("hash chain reveal must be 32 bytes, got %d", len(reveal))
	}
	copy(b.block.body.hashChainReveal[:], reveal)
	return nil
}

func (b *SignedBeaconBlock) SetRandaoCommitmentRegistrations(registrations []*eth.SignedRandaoCommitmentRegistration) error {
	if b.version < version.Eip8321 {
		return consensus_types.ErrNotSupported("SetRandaoCommitmentRegistrations", b.version)
	}
	b.block.body.randaoCommitmentRegistrations = registrations
	return nil
}
