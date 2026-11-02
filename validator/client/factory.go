package client

import (
	"github.com/OffchainLabs/prysm/v7/config/features"
	"github.com/OffchainLabs/prysm/v7/config/params"
	beaconApi "github.com/OffchainLabs/prysm/v7/validator/client/beacon-api"
	grpcApi "github.com/OffchainLabs/prysm/v7/validator/client/grpc-api"
	"github.com/OffchainLabs/prysm/v7/validator/client/iface"
	validatorHelpers "github.com/OffchainLabs/prysm/v7/validator/helpers"
)

func NewNodeClient(validatorConn *validatorHelpers.NodeConnection) iface.NodeClient {
	grpcClient := grpcApi.NewNodeClient(validatorConn)
	if features.Get().EnableBeaconRESTApi && validatorConn.GetRestConnectionProvider() != nil {
		restProvider := validatorConn.GetRestConnectionProvider()
		return beaconApi.NewNodeClientWithFallback(restProvider, grpcClient)
	}
	return grpcClient
}

func NewChainClient(validatorConn *validatorHelpers.NodeConnection) iface.ChainClient {
	grpcClient := grpcApi.NewGrpcChainClient(validatorConn)
	if features.Get().EnableBeaconRESTApi && validatorConn.GetRestConnectionProvider() != nil {
		return beaconApi.NewBeaconApiChainClientWithFallback(validatorConn.GetRestConnectionProvider(), grpcClient)
	}
	return grpcClient
}

func NewValidatorClient(validatorConn *validatorHelpers.NodeConnection, opts ...iface.Option) iface.ValidatorClient {
	if features.Get().EnableBeaconRESTApi && validatorConn.GetRestConnectionProvider() != nil {
		restClient := beaconApi.NewBeaconApiValidatorClient(validatorConn.GetRestConnectionProvider(), opts...)
		if params.BeaconConfig().Eip8321ForkEpoch < params.BeaconConfig().FarFutureEpoch && validatorConn.GetGrpcConnectionProvider() != nil {
			return &eip8321ValidatorClient{ValidatorClient: restClient, grpcClient: grpcApi.NewGrpcValidatorClient(validatorConn, opts...)}
		}
		return restClient
	}
	return grpcApi.NewGrpcValidatorClient(validatorConn, opts...)
}
