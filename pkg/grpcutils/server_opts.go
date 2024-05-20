package grpcutils

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

var DefaultServerOptions = func(logger *zap.Logger) []interface{} {
	return []interface{}{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 1 * time.Minute, PermitWithoutStream: true}),
		CtxTagsServerOption(),
		ValidatorServerOption(),
		RecoveryServerOption(),
	}
}

var DefaultLoggingServerOptions = func(logger *zap.Logger) []interface{} {
	return []interface{}{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 1 * time.Minute, PermitWithoutStream: true}),
		CtxTagsServerOption(),
		ValidatorServerOption(),
		RecoveryServerOption(),
	}
}

var TestSvrOpts = func(logger *zap.Logger) []interface{} {
	return []interface{}{
		CtxTagsServerOption,
		ValidatorServerOption,
		RecoveryServerOption,
	}
}

var DefaultCodeToLevel = func(code codes.Code) zapcore.Level {
	switch code {
	case codes.Unknown, codes.DeadlineExceeded, codes.PermissionDenied, codes.Unauthenticated,
		codes.Unimplemented, codes.Internal, codes.DataLoss, codes.Unavailable, codes.OutOfRange,
		codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted:
		return zap.InfoLevel
	default:
		return zap.DebugLevel
	}
}

var DebugCodeToLevel = func(code codes.Code) zapcore.Level {
	return zap.DebugLevel
}

var ServerPayloadLoggingDecider = func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return false }

func CtxTagsServerOption() ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors,
			grpcCtxTags.UnaryServerInterceptor(grpcCtxTags.WithFieldExtractor(grpcCtxTags.CodeGenRequestFieldExtractor)))
		s.streamInterceptors = append(s.streamInterceptors,
			grpcCtxTags.StreamServerInterceptor(grpcCtxTags.WithFieldExtractor(grpcCtxTags.CodeGenRequestFieldExtractor)))
	}
}

func ValidatorServerOption() ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors, grpcValidator.UnaryServerInterceptor())
		s.streamInterceptors = append(s.streamInterceptors, grpcValidator.StreamServerInterceptor())
	}
}

func RecoveryServerOption() ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors, grpcRecovery.UnaryServerInterceptor())
		s.streamInterceptors = append(s.streamInterceptors, grpcRecovery.StreamServerInterceptor())
	}
}
