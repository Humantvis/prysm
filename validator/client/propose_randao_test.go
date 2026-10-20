package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	blscommon "github.com/OffchainLabs/prysm/v7/crypto/bls/common"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	validatormock "github.com/OffchainLabs/prysm/v7/testing/validator-mock"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
)

func expectUnregisteredRandao(client *validatormock.MockValidatorClient, pubKey [fieldparams.BLSPubkeyLength]byte) {
	client.EXPECT().RandaoCommitment(gomock.Any(), &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]}).
		Return(&ethpb.RandaoCommitmentResponse{Commitment: make([]byte, fieldparams.RootLength)}, nil).Times(1)
}

func TestProposeBlock_RandaoCommitmentFailure(t *testing.T) {
	commitment := make([]byte, fieldparams.RootLength)
	commitment[0] = 1
	tests := []struct {
		name     string
		response *ethpb.RandaoCommitmentResponse
		err      error
		wantLog  string
	}{
		{name: "RPC error", err: errors.New("lookup failed"), wantLog: "Failed to determine hash-chain registration"},
		{name: "nil response", wantLog: "Failed to determine hash-chain registration"},
		{name: "invalid commitment length", response: &ethpb.RandaoCommitmentResponse{Commitment: make([]byte, fieldparams.RootLength-1)}, wantLog: "Failed to determine hash-chain registration"},
		{name: "missing local seed", response: &ethpb.RandaoCommitmentResponse{Commitment: commitment}, wantLog: "On-chain commitment not found in local chain"},
	}
	for _, tt := range tests {
		for _, minimalDB := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/minimalDB=%v", tt.name, minimalDB), func(t *testing.T) {
				hook := logTest.NewGlobal()
				v, m, key, finish := setup(t, minimalDB)
				defer finish()
				pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
				m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]}).Return(tt.response, tt.err).Times(1)
				v.ProposeBlock(t.Context(), 1, pubKey)
				require.LogsContain(t, hook, tt.wantLog)
			})
		}
	}
}

func TestProposeBlock_HashChainReveal(t *testing.T) {
	for _, minimalDB := range []bool{false, true} {
		t.Run(fmt.Sprintf("minimalDB=%v", minimalDB), func(t *testing.T) {
			v, m, key, finish := setup(t, minimalDB)
			defer finish()
			pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
			seed := [32]byte{1}
			chain, err := eip8321.ComputeHashChain(seed, 3)
			require.NoError(t, err)
			require.NoError(t, v.db.SaveRandaoHashChain(t.Context(), pubKey, seed, 3))
			m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]}).
				Return(&ethpb.RandaoCommitmentResponse{Commitment: chain[3][:]}, nil).Times(1)
			m.validatorClient.EXPECT().BeaconBlock(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *ethpb.BlockRequest) (*ethpb.GenericBeaconBlock, error) {
				require.DeepEqual(t, chain[2][:], req.HashChainReveal)
				require.DeepEqual(t, blscommon.InfiniteSignature[:], req.RandaoReveal)
				return nil, errors.New("stop after block request")
			}).Times(1)
			v.ProposeBlock(t.Context(), 1, pubKey)
		})
	}
}
