package client

import (
	"context"
	"errors"
	"testing"

	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/encoding/bytesutil"
	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/mock/gomock"
)

func TestAutomaticRandaoRegistration(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.Eip8321ForkEpoch = 4
	params.OverrideBeaconConfig(cfg)
	for _, minimal := range []bool{false, true} {
		t.Run(map[bool]string{false: "full DB", true: "minimal DB"}[minimal], func(t *testing.T) {
			v, m, key, finish := setup(t, minimal)
			defer finish()
			ctx := t.Context()
			pub := bytesutil.ToBytes48(key.PublicKey().Marshal())
			require.NoError(t, v.db.SaveGenesisValidatorsRoot(ctx, make([]byte, 32)))
			v.registerUncommittedRandao(ctx, 3)
			m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), gomock.Any()).Return(&ethpb.RandaoCommitmentResponse{Commitment: make([]byte, 32)}, nil).Times(2)
			m.validatorClient.EXPECT().ValidatorIndex(gomock.Any(), gomock.Any()).Return(&ethpb.ValidatorIndexResponse{Index: 1}, nil).Times(2)
			var submitted []byte
			m.validatorClient.EXPECT().SubmitRandaoCommitmentRegistration(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *ethpb.SignedRandaoCommitmentRegistration) (*empty.Empty, error) {
				submitted = append([]byte(nil), r.Message.Commitment...)
				return nil, errors.New("temporary outage")
			})
			v.registerUncommittedRandao(ctx, 4)
			seed, length, ok, err := v.db.RandaoHashChain(ctx, pub)
			require.NoError(t, err)
			require.Equal(t, true, ok)
			m.validatorClient.EXPECT().SubmitRandaoCommitmentRegistration(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *ethpb.SignedRandaoCommitmentRegistration) (*empty.Empty, error) {
				require.DeepEqual(t, submitted, r.Message.Commitment)
				return &empty.Empty{}, nil
			})
			v.registerUncommittedRandao(ctx, 5)
			gotSeed, gotLength, _, err := v.db.RandaoHashChain(ctx, pub)
			require.NoError(t, err)
			require.Equal(t, seed, gotSeed)
			require.Equal(t, length, gotLength)
			m.validatorClient.EXPECT().RandaoCommitment(gomock.Any(), gomock.Any()).Return(&ethpb.RandaoCommitmentResponse{Commitment: submitted}, nil)
			v.registerUncommittedRandao(ctx, 8)
		})
	}
}
