package grpcutils

import (
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var DefaultDialOpts = func(logger *zap.Logger) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 60 * time.Second, Timeout: time.Minute * 2, PermitWithoutStream: true}),
	}
}

var DefaultLoggingDialOpts = func(logger *zap.Logger) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 60 * time.Second, Timeout: time.Minute * 2, PermitWithoutStream: true}),
	}
}

var ClientPayloadLoggingDecider = func(ctx context.Context, fullMethodName string) bool { return false }
