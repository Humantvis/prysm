package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/core/eip8321"
	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/consensus-types/primitives"
	blscommon "github.com/OffchainLabs/prysm/v7/crypto/bls/common"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	validatormock "github.com/OffchainLabs/prysm/v7/testing/validator-mock"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"go.uber.org/mock/gomock"
)

// expectUnregisteredRandao expects one commitment lookup returning "unregistered"
// when the EIP-8321 fork is scheduled at genesis. When it is not active (the
// default test config has it at FAR_FUTURE) ProposeBlock must not look the
// commitment up at all, so no call is expected and gomock fails on any call.
func expectUnregisteredRandao(client *validatormock.MockValidatorClient, pubKey [fieldparams.BLSPubkeyLength]byte) {
	if params.BeaconConfig().Eip8321ForkEpoch != 0 {
		return
	}
	client.EXPECT().RandaoCommitment(gomock.Any(), &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]}).
		Return(&ethpb.RandaoCommitmentResponse{Commitment: make([]byte, fieldparams.RootLength)}, nil).Times(1)
}

func setEip8321ForkEpoch(t *testing.T, epoch primitives.Epoch) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.Eip8321ForkEpoch = epoch
	params.OverrideBeaconConfig(cfg)
}

func TestProposeBlock_RandaoCommitmentFailure(t *testing.T) {
	setEip8321ForkEpoch(t, 0)
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
	setEip8321ForkEpoch(t, 0)
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

// expectBLSRevealBlockRequest expects a BLS randao reveal to be signed and the
// block to be requested with it and without a hash-chain reveal.
func expectBLSRevealBlockRequest(t *testing.T, client *validatormock.MockValidatorClient, slot primitives.Slot) {
	client.EXPECT().DomainData(gomock.Any(), gomock.Any()).
		Return(&ethpb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil).Times(1)
	client.EXPECT().BeaconBlock(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *ethpb.BlockRequest) (*ethpb.GenericBeaconBlock, error) {
		require.Equal(t, slot, req.Slot)
		require.Equal(t, 0, len(req.HashChainReveal))
		require.Equal(t, fieldparams.BLSSignatureLength, len(req.RandaoReveal))
		require.NotEqual(t, true, string(blscommon.InfiniteSignature[:]) == string(req.RandaoReveal))
		return nil, errors.New("stop after block request")
	}).Times(1)
}

// With the fork unscheduled, ProposeBlock must not query the hash-chain
// commitment (the REST client does not support it) and must use a BLS reveal.
func TestProposeBlock_ForkUnscheduled_UsesBLSReveal(t *testing.T) {
	setEip8321ForkEpoch(t, params.BeaconConfig().FarFutureEpoch)
	for _, minimalDB := range []bool{false, true} {
		t.Run(fmt.Sprintf("minimalDB=%v", minimalDB), func(t *testing.T) {
			hook := logTest.NewGlobal()
			v, m, key, finish := setup(t, minimalDB)
			defer finish()
			pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
			m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), gomock.Any()).Times(0)
			expectBLSRevealBlockRequest(t, m.validatorClient, 1)
			v.ProposeBlock(t.Context(), 1, pubKey)
			require.LogsDoNotContain(t, hook, "Failed to determine hash-chain registration")
		})
	}
}

func TestProposeBlock_ForkScheduled_RevealByEpoch(t *testing.T) {
	const forkEpoch = primitives.Epoch(2)
	setEip8321ForkEpoch(t, forkEpoch)
	spe := params.BeaconConfig().SlotsPerEpoch

	t.Run("epoch before fork uses BLS reveal", func(t *testing.T) {
		v, m, key, finish := setup(t, false)
		defer finish()
		pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
		slot := spe.Mul(uint64(forkEpoch-1)) + 1
		m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), gomock.Any()).Times(0)
		expectBLSRevealBlockRequest(t, m.validatorClient, slot)
		v.ProposeBlock(t.Context(), slot, pubKey)
	})

	t.Run("fork epoch unregistered queries commitment then uses BLS reveal", func(t *testing.T) {
		v, m, key, finish := setup(t, false)
		defer finish()
		pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
		slot := spe.Mul(uint64(forkEpoch))
		m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), &ethpb.ValidatorIndexRequest{PublicKey: pubKey[:]}).
			Return(&ethpb.RandaoCommitmentResponse{Commitment: make([]byte, fieldparams.RootLength)}, nil).Times(1)
		expectBLSRevealBlockRequest(t, m.validatorClient, slot)
		v.ProposeBlock(t.Context(), slot, pubKey)
	})

	t.Run("fork epoch registered uses hash-chain reveal", func(t *testing.T) {
		v, m, key, finish := setup(t, false)
		defer finish()
		pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
		slot := spe.Mul(uint64(forkEpoch))
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
		v.ProposeBlock(t.Context(), slot, pubKey)
	})

	t.Run("fork epoch lookup error fails closed", func(t *testing.T) {
		hook := logTest.NewGlobal()
		v, m, key, finish := setup(t, false)
		defer finish()
		pubKey := bytesutil.ToBytes48(key.PublicKey().Marshal())
		slot := spe.Mul(uint64(forkEpoch))
		m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), gomock.Any()).Return(nil, errors.New("endpoint not supported")).Times(1)
		v.ProposeBlock(t.Context(), slot, pubKey)
		require.LogsContain(t, hook, "Failed to determine hash-chain registration")
	})
}
