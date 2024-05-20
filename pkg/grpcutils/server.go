package grpcutils

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServerOption func(s *Server)

type Server struct {
	Server             *grpc.Server
	serviceName        string
	addr               string
	logger             *zap.Logger
	serverOpts         []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	cleanFuncs         []func(s *Server) error
}

func NewServer(name, addr string) *Server {
	return &Server{serviceName: name, addr: addr}
}

func (s *Server) Init(opts ...interface{}) error {
	for _, opt := range opts {
		switch o := opt.(type) {
		case ServerOption:
			o(s)
		case grpc.ServerOption:
			s.serverOpts = append(s.serverOpts, o)
		default:
			return fmt.Errorf("invalid server option: %v", opt)
		}
	}
	if s.logger == nil {
		s.logger = zap.L()
	}
	s.logger = s.logger.With(zap.String("addr", s.addr))

	s.serverOpts = append(s.serverOpts,
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.ChainStreamInterceptor(s.streamInterceptors...),
	)
	s.Server = grpc.NewServer(s.serverOpts...)

	return nil
}

func (s *Server) CleanUp() error {
	var err *multierror.Error
	for _, clean := range s.cleanFuncs {
		err = multierror.Append(err, clean(s))
	}

	return err.ErrorOrNil()
}

func (s *Server) Serve() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.cleanFuncs = append([]func(s *Server) error{
		func(srv *Server) error {
			if srv.Server != nil {
				srv.Server.GracefulStop()
			}
			return nil
		},
	}, s.cleanFuncs...)

	return s.Server.Serve(listener)
}

func (s *Server) GetHttpProxy(httpAddr string, opts []grpc.DialOption, register func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) (*http.Server, error) {
	mux := runtime.NewServeMux()
	conn := NewClientConnection(s.addr)
	if err := conn.Dial(opts...); err != nil {
		return nil, err
	}
	s.cleanFuncs = append(s.cleanFuncs, func(*Server) error {
		return conn.CleanUp()
	})

	ctx := context.Background()
	if err := register(ctx, mux, conn.Connection); err != nil {
		return nil, err
	}

	return &http.Server{Addr: httpAddr, Handler: mux}, nil
}
