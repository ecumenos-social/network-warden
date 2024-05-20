package grpcutils

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DialOption struct {
	grpc.EmptyDialOption
	function func(conn *ClientConnection)
}

func NewDialOption(f func(conn *ClientConnection)) DialOption {
	return DialOption{EmptyDialOption: grpc.EmptyDialOption{}, function: f}
}

type ClientConnection struct {
	Connection         *grpc.ClientConn
	addr               string
	logger             *zap.Logger
	dialOpts           []grpc.DialOption
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	cleanFuncs         []func(c *ClientConnection) error
}

func NewClientConnection(addr string) *ClientConnection {
	return &ClientConnection{addr: addr}
}

func (c *ClientConnection) Dial(opts ...grpc.DialOption) error {
	return c.DialContext(context.Background(), opts...)
}

func (c *ClientConnection) DialContext(ctx context.Context, opts ...grpc.DialOption) error {
	for _, opt := range opts {
		if o, ok := opt.(DialOption); ok {
			o.function(c)
		} else {
			c.dialOpts = append(c.dialOpts, opt)
		}
	}
	if c.logger == nil {
		c.logger = zap.L()
	}
	c.logger = c.logger.With(zap.String("addr", c.addr))

	c.dialOpts = append(c.dialOpts,
		grpc.WithChainUnaryInterceptor(c.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(c.streamInterceptors...),
	)

	c.logger.Info("dialing gRPC client connection...")
	conn, err := grpc.NewClient(c.addr, c.dialOpts...)
	if err != nil {
		return err
	}

	c.Connection = conn
	c.cleanFuncs = append(c.cleanFuncs, func(c *ClientConnection) error {
		return c.Connection.Close()
	})

	return nil
}

func (c *ClientConnection) CleanUp() error {
	var err *multierror.Error
	for _, clean := range c.cleanFuncs {
		err = multierror.Append(err, clean(c))
	}

	return err.ErrorOrNil()
}
