package epoch_processing

import (
	"os"
	"path"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/build/bazel"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/utils"
	"github.com/golang/snappy"
)

type epochOperation func(*testing.T, state.BeaconState) (state.BeaconState, error)

// runEpochHandler runs every case of an epoch_processing handler. It uses absolute
// paths so a missing post state never triggers a download of the upstream archive.
func runEpochHandler(t *testing.T, config, handler string, operationFn epochOperation) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, rel := utils.TestFolders(t, config, "eip8321", handler)
	testsFolderPath, err := bazel.Runfile(rel)
	require.NoError(t, err)
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			RunEpochOperationTest(t, path.Join(testsFolderPath, folder.Name()), operationFn)
		})
	}
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

// RunEpochOperationTest takes in the prestate and processes it through the
// passed in epoch operation function and checks the post state with the expected post state.
func RunEpochOperationTest(t *testing.T, testFolderPath string, operationFn epochOperation) {
	preSSZ, ok := readSnappy(t, path.Join(testFolderPath, "pre.ssz_snappy"))
	require.Equal(t, true, ok, "missing pre.ssz_snappy")
	preBase := &ethpb.BeaconStateEip8321{}
	require.NoError(t, preBase.UnmarshalSSZ(preSSZ), "Failed to unmarshal")
	preState, err := state_native.InitializeFromProtoUnsafeEip8321(preBase)
	require.NoError(t, err)

	beaconState, err := operationFn(t, preState)

	postSSZ, postExists := readSnappy(t, path.Join(testFolderPath, "post.ssz_snappy"))
	if !postExists {
		if err == nil {
			t.Fatal("Did not fail when expected")
		}
		t.Logf("Expected failure; failure reason = %v", err)
		return
	}
	require.NoError(t, err)
	postState := &ethpb.BeaconStateEip8321{}
	require.NoError(t, postState.UnmarshalSSZ(postSSZ), "Failed to unmarshal")
	pbState, err := state_native.ProtobufBeaconStateEip8321(beaconState.ToProtoUnsafe())
	require.NoError(t, err)
	require.DeepSSZEqual(t, pbState, postState, "Post state does not match expected")
}
