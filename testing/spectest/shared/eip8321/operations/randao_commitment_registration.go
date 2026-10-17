package operations

import (
	"context"
	"path"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/consensus-types/blocks"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

// RunRandaoCommitmentRegistrationTest executes "operations/randao_commitment_registration" tests.
func RunRandaoCommitmentRegistrationTest(t *testing.T, config string) {
	testFolders, testsFolderPath := testCases(t, config, "operations/randao_commitment_registration/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			regSSZ, ok := readSnappy(t, path.Join(folderPath, "randao_commitment_registration.ssz_snappy"))
			require.Equal(t, true, ok, "missing randao_commitment_registration.ssz_snappy")
			registration := &ethpb.SignedRandaoCommitmentRegistration{}
			require.NoError(t, registration.UnmarshalSSZ(regSSZ))
			blk, err := blocks.NewSignedBeaconBlock(&ethpb.SignedBeaconBlockEip8321{
				Block: &ethpb.BeaconBlockEip8321{
					Body: &ethpb.BeaconBlockBodyEip8321{
						RandaoCommitmentRegistrations: []*ethpb.SignedRandaoCommitmentRegistration{registration},
					},
				},
			})
			require.NoError(t, err)
			runBlockOperation(t, folderPath, blk, func(_ context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				registrations, err := b.Block().Body().RandaoCommitmentRegistrations()
				if err != nil {
					return nil, err
				}
				if err := eip8321.ProcessRandaoCommitmentRegistrations(s, registrations); err != nil {
					return nil, err
				}
				return s, nil
			})
		})
	}
}
