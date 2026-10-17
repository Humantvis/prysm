package operations

import (
	"context"
	"path"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/blocks"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	"github.com/OffchainLabs/prysm/v7/testing/require"
)

// RunRandaoTest executes "operations/randao" tests (EIP-8321 process_randao).
func RunRandaoTest(t *testing.T, config string) {
	testFolders, testsFolderPath := testCases(t, config, "operations/randao/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			blockSSZ, ok := readSnappy(t, path.Join(folderPath, "block.ssz_snappy"))
			if !ok {
				// Multi-block cases (hash_chain_reveal_consumes_one_link_per_block)
				// produce only a manifest in the operations format.
				t.Skipf("no block.ssz_snappy in %s", folder.Name())
			}
			blk, err := sszToBlock(blockSSZ)
			require.NoError(t, err)
			runBlockOperation(t, folderPath, blk, func(ctx context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				return blocks.ProcessRandao(ctx, s, b)
			})
		})
	}
}
