package eip8321

import (
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/signing"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/crypto/bls"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/pkg/errors"
)

// RandaoCommitmentRegistrationSignatureBatch builds the BLS signature batch
// for verifying one or more randao commitment registrations.
//
// Domain is anchored to the genesis fork version, matching
// BLSToExecutionChange, so registrations stay valid across forks.
//
// Shared by gossip validation (beacon-chain/sync) and RPC submission
// (beacon-chain/rpc) so the domain and signing-root computation can never
// drift between the two paths.
func RandaoCommitmentRegistrationSignatureBatch(
	st state.ReadOnlyBeaconState,
	signed []*ethpb.SignedRandaoCommitmentRegistration,
) (*bls.SignatureBatch, error) {
	cfg := params.BeaconConfig()
	domain, err := signing.ComputeDomain(
		cfg.DomainRandaoCommitmentRegistration,
		cfg.GenesisForkVersion,
		st.GenesisValidatorsRoot(),
	)
	if err != nil {
		return nil, err
	}

	batch := &bls.SignatureBatch{
		Signatures:   make([][]byte, 0, len(signed)),
		PublicKeys:   make([]bls.PublicKey, 0, len(signed)),
		Messages:     make([][32]byte, 0, len(signed)),
		Descriptions: make([]string, 0, len(signed)),
	}
	for _, s := range signed {
		if s == nil || s.Message == nil {
			return nil, errors.New("nil registration")
		}
		root, err := signing.ComputeSigningRoot(s.Message, domain)
		if err != nil {
			return nil, err
		}
		pk := st.PubkeyAtIndex(s.Message.ValidatorIndex)
		pub, err := bls.PublicKeyFromBytes(pk[:])
		if err != nil {
			return nil, err
		}
		batch.Signatures = append(batch.Signatures, s.Signature)
		batch.PublicKeys = append(batch.PublicKeys, pub)
		batch.Messages = append(batch.Messages, root)
		batch.Descriptions = append(batch.Descriptions, "randao commitment registration signature")
	}
	return batch, nil
}