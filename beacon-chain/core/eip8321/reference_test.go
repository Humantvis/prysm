package eip8321_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/config/params"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/golang/snappy"
)

// Generated locally with consensus-specs make test reftests=true fork=eip8321.
func TestEip8321ReferenceFork(t *testing.T) {
	setupEip8321Preset(t)
	root := os.Getenv("EIP8321_SPEC_TESTS_DIR")
	if root == "" {
		t.Skip("set EIP8321_SPEC_TESTS_DIR to the generated reftests directory")
	}
	pattern := filepath.Join(root, params.BeaconConfig().PresetBase, "eip8321", "fork", "*", "*", "*", "pre.ssz_snappy")
	cases, err := filepath.Glob(pattern)
	require.NoError(t, err)
	if len(cases) == 0 {
		t.Fatalf("no EIP-8321 fork vectors matched %s", pattern)
	}
	decode := func(t *testing.T, path string) []byte {
		t.Helper()
		compressed, err := os.ReadFile(path)
		require.NoError(t, err)
		data, err := snappy.Decode(nil, compressed)
		require.NoError(t, err)
		return data
	}
	for _, path := range cases {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			pre := &ethpb.BeaconStateHeze{}
			require.NoError(t, pre.UnmarshalSSZ(decode(t, path)))
			st, err := state_native.InitializeFromProtoHeze(pre)
			require.NoError(t, err)
			post, err := eip8321.UpgradeToEip8321(st)
			require.NoError(t, err)
			wantBytes := decode(t, filepath.Join(filepath.Dir(path), "post.ssz_snappy"))
			gotBytes, err := post.MarshalSSZ()
			require.NoError(t, err)
			require.DeepEqual(t, wantBytes, gotBytes)
			want := &ethpb.BeaconStateEip8321{}
			require.NoError(t, want.UnmarshalSSZ(wantBytes))
			wantRoot, err := want.HashTreeRoot()
			require.NoError(t, err)
			gotRoot, err := post.HashTreeRoot(t.Context())
			require.NoError(t, err)
			require.Equal(t, wantRoot, gotRoot)
		})
	}
}
