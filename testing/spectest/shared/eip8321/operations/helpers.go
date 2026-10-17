package operations

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/helpers"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/build/bazel"
	"github.com/OffchainLabs/prysm/v7/consensus-types/blocks"
	"github.com/OffchainLabs/prysm/v7/consensus-types/interfaces"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/utils"
	"github.com/golang/snappy"
)

type blockOperation func(context.Context, state.BeaconState, interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error)

func sszToState(b []byte) (state.BeaconState, error) {
	base := &ethpb.BeaconStateEip8321{}
	if err := base.UnmarshalSSZ(b); err != nil {
		return nil, err
	}
	return state_native.InitializeFromProtoUnsafeEip8321(base)
}

func sszToBlock(b []byte) (interfaces.SignedBeaconBlock, error) {
	base := &ethpb.BeaconBlockEip8321{}
	if err := base.UnmarshalSSZ(b); err != nil {
		return nil, err
	}
	return blocks.NewSignedBeaconBlock(&ethpb.SignedBeaconBlockEip8321{Block: base})
}

// testCases returns the case folders of a handler and the absolute path of the
// handler directory. Absolute paths matter outside Bazel: a missing relative
// path would make the test-data loader try to download the upstream archive.
func testCases(t *testing.T, config, handler string) ([]os.DirEntry, string) {
	require.NoError(t, utils.SetConfig(t, config))
	folders, rel := utils.TestFolders(t, config, "eip8321", handler)
	abs, err := bazel.Runfile(rel)
	require.NoError(t, err)
	return folders, abs
}

func readSnappy(t *testing.T, file string) ([]byte, bool) {
	compressed, err := os.ReadFile(file) // #nosec G304
	if os.IsNotExist(err) {
		return nil, false
	}
	require.NoError(t, err)
	data, err := snappy.Decode(nil /* dst */, compressed)
	require.NoError(t, err, "Failed to decompress")
	return data, true
}

// runBlockOperation loads pre.ssz_snappy, applies the operation and compares the
// result with post.ssz_snappy. A missing post state means the operation must fail.
func runBlockOperation(t *testing.T, folderPath string, blk interfaces.SignedBeaconBlock, op blockOperation) {
	preSSZ, ok := readSnappy(t, path.Join(folderPath, "pre.ssz_snappy"))
	require.Equal(t, true, ok, "missing pre.ssz_snappy")
	preState, err := sszToState(preSSZ)
	require.NoError(t, err)

	helpers.ClearCache()
	_, err = op(context.Background(), preState, blk)

	postSSZ, postExists := readSnappy(t, path.Join(folderPath, "post.ssz_snappy"))
	if !postExists {
		if err == nil {
			t.Fatal("Did not fail when expected")
		}
		t.Logf("Expected failure; failure reason = %v", err)
		return
	}
	require.NoError(t, err)
	postState, err := sszToState(postSSZ)
	require.NoError(t, err)
	require.DeepSSZEqual(t, postState.ToProtoUnsafe(), preState.ToProtoUnsafe(), "Post state does not match expected")
}
