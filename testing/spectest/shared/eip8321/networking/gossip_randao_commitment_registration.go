package networking

import (
	"path"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/helpers"
	state_native "github.com/OffchainLabs/prysm/v7/beacon-chain/state/state-native"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/OffchainLabs/prysm/v7/testing/spectest/utils"
	"github.com/OffchainLabs/prysm/v7/testing/util"
	"github.com/golang/snappy"
)

type gossipMeta struct {
	Topic  string `spec-name:"topic"`
	Blocks []struct {
		Block string `spec-name:"block"`
	} `spec-name:"blocks"`
	Messages []struct {
		Message  string `spec-name:"message"`
		Expected string `spec-name:"expected"`
		Reason   string `spec-name:"reason"`
	} `spec-name:"messages"`
}

// skipCases cannot be reproduced from the vector files alone.
var skipCases = map[string]string{
	// The generator primes seen.randao_commitment_registration_indices with
	// index 0 before validating, but meta.yaml does not export it, so a client
	// sees an empty seen set and (correctly) accepts the message.
	"gossip_randao_commitment_registration__ignore_already_seen": "seen set is primed in the generator but not exported in meta.yaml",
}

// RunGossipRandaoCommitmentRegistrationTest runs
// networking/gossip_randao_commitment_registration: every message is validated
// against the anchor state, sharing one seen set, and the result
// (valid/ignore/reject) is compared with the expected one.
func RunGossipRandaoCommitmentRegistrationTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, "eip8321", "networking/gossip_randao_commitment_registration/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			if reason, ok := skipCases[folder.Name()]; ok {
				t.Skip(reason)
			}
			helpers.ClearCache()
			folderPath := path.Join(testsFolderPath, folder.Name())

			metaFile, err := util.BazelFileBytes(path.Join(folderPath, "meta.yaml"))
			require.NoError(t, err)
			var meta gossipMeta
			require.NoError(t, utils.UnmarshalYaml(metaFile, &meta))
			require.Equal(t, "randao_commitment_registration", meta.Topic)
			// Every case so far has only the anchor block, so the head state is
			// the anchor state and no block needs to be replayed.
			if len(meta.Blocks) > 1 {
				t.Skip("replaying blocks before validation is not implemented")
			}

			stateSSZ := readSnappy(t, path.Join(folderPath, "state.ssz_snappy"))
			base := &ethpb.BeaconStateEip8321{}
			require.NoError(t, base.UnmarshalSSZ(stateSSZ))
			st, err := state_native.InitializeFromProtoUnsafeEip8321(base)
			require.NoError(t, err)

			seen := make(map[primitives.ValidatorIndex]bool)
			seenFn := func(idx primitives.ValidatorIndex) bool { return seen[idx] }

			for _, m := range meta.Messages {
				msgSSZ := readSnappy(t, path.Join(folderPath, m.Message+".ssz_snappy"))
				signed := &ethpb.SignedRandaoCommitmentRegistration{}
				require.NoError(t, signed.UnmarshalSSZ(msgSSZ))

				result, batch, err := eip8321.ValidateRandaoCommitmentRegistrationGossip(st, seenFn, signed)
				got := "reject"
				switch result {
				case eip8321.GossipIgnore:
					got = "ignore"
				case eip8321.GossipAccept:
					ok, verr := batch.Verify()
					if verr == nil && ok {
						got = "valid"
						seen[signed.Message.ValidatorIndex] = true
					} else {
						err = verr
					}
				}
				require.Equal(t, m.Expected, got, "message %s (expected reason: %q, got error: %v)", m.Message, m.Reason, err)
			}
		})
	}
}

func readSnappy(t *testing.T, file string) []byte {
	compressed, err := util.BazelFileBytes(file)
	require.NoError(t, err)
	data, err := snappy.Decode(nil /* dst */, compressed)
	require.NoError(t, err, "Failed to decompress")
	return data
}
