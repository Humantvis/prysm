package eth_test

import (
	stdbytes "bytes"
	"context"
	"net"
	"testing"
	"time"

	ethpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1"
	validatorpb "github.com/OffchainLabs/prysm/v7/proto/prysm/v1alpha1/validator-client"
	"github.com/OffchainLabs/prysm/v7/testing/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type randaoRPCServer struct {
	ethpb.UnimplementedBeaconNodeValidatorServer
	registration *ethpb.SignedRandaoCommitmentRegistration
	publicKey    []byte
}

func (s *randaoRPCServer) SubmitRandaoCommitmentRegistration(_ context.Context, req *ethpb.SignedRandaoCommitmentRegistration) (*emptypb.Empty, error) {
	s.registration = req
	return &emptypb.Empty{}, nil
}

func (s *randaoRPCServer) RandaoCommitment(_ context.Context, req *ethpb.ValidatorIndexRequest) (*ethpb.RandaoCommitmentResponse, error) {
	s.publicKey = req.PublicKey
	return &ethpb.RandaoCommitmentResponse{Commitment: s.registration.Message.Commitment}, nil
}

func TestRandaoCommitmentRPCWireRoundTrip(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	implementation := &randaoRPCServer{}
	ethpb.RegisterBeaconNodeValidatorServer(server, implementation)
	t.Cleanup(server.Stop)
	go func() { _ = server.Serve(listener) }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.NewClient("passthrough:///randao", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	client := ethpb.NewBeaconNodeValidatorClient(conn)
	registration := &ethpb.SignedRandaoCommitmentRegistration{
		Message:   &ethpb.RandaoCommitmentRegistration{ValidatorIndex: 42, Commitment: stdbytes.Repeat([]byte{7}, 32)},
		Signature: stdbytes.Repeat([]byte{9}, 96),
	}
	_, err = client.SubmitRandaoCommitmentRegistration(ctx, registration)
	require.NoError(t, err)
	publicKey := stdbytes.Repeat([]byte{3}, 48)
	response, err := client.RandaoCommitment(ctx, &ethpb.ValidatorIndexRequest{PublicKey: publicKey})
	require.NoError(t, err)
	require.Equal(t, true, proto.Equal(registration, implementation.registration))
	require.DeepEqual(t, publicKey, implementation.publicKey)
	require.DeepEqual(t, registration.Message.Commitment, response.Commitment)
}

func TestRandaoCommitmentSigningWireRoundTrip(t *testing.T) {
	registration := &ethpb.RandaoCommitmentRegistration{ValidatorIndex: 42, Commitment: stdbytes.Repeat([]byte{7}, 32)}
	request := &validatorpb.SignRequest{Object: &validatorpb.SignRequest_RandaoCommitment{RandaoCommitment: registration}}
	encoded, err := proto.Marshal(request)
	require.NoError(t, err)
	decoded := new(validatorpb.SignRequest)
	require.NoError(t, proto.Unmarshal(encoded, decoded))
	require.Equal(t, true, proto.Equal(registration, decoded.GetRandaoCommitment()))
	fields := request.ProtoReflect().Descriptor().Fields()
	require.Equal(t, int32(131), int32(fields.ByName("randao_commitment").Number()))
	require.Equal(t, int32(130), int32(fields.ByName("block_eip8321").Number()))
}
