package grpcutils

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthHandler struct {
	*health.Server
	GServer *Server
	svrMap  map[string]*grpc.ClientConn
	logger  *zap.Logger
}

func NewHealthServer(svrName string, addr string, logger *zap.Logger, healthAddr string) (healthHandler *HealthHandler) {
	healthHandler = &HealthHandler{
		Server:  health.NewServer(),
		GServer: NewServer("Health", healthAddr),
		logger:  logger,
		svrMap:  make(map[string]*grpc.ClientConn),
	}
	healthHandler.GServer.Init()
	healthPb.RegisterHealthServer(healthHandler.GServer.Server, healthHandler)
	healthHandler.AddService(svrName, addr)
	return
}

func (s *HealthHandler) AddConn(svrName string, conn *ClientConnection) {
	s.svrMap[svrName] = conn.Connection
}

func (s *HealthHandler) AddService(svrName string, addr string) {
	s.connect(svrName, addr)
}

func (s *HealthHandler) Check(ctx context.Context, in *healthPb.HealthCheckRequest) (*healthPb.HealthCheckResponse, error) {
	if in.Service == "" {
		return s.checkOverallStatus(ctx)
	}
	return s.Server.Check(ctx, in)
}

func (s *HealthHandler) CheckConnection(ctx context.Context) error {
	s.initStatus()
	for svrName, conn := range s.svrMap {
		go s.connTracker(ctx, svrName, conn)
	}
	<-ctx.Done()
	return nil
}

func (s *HealthHandler) connTracker(ctx context.Context, svrName string, conn *grpc.ClientConn) {
	lastState := conn.GetState()
	s.handleState(svrName, lastState, conn)
	for conn.WaitForStateChange(ctx, lastState) {
		lastState = conn.GetState()
		s.handleState(svrName, lastState, conn)
	}
}

func (s *HealthHandler) handleState(svrName string, state connectivity.State, conn *grpc.ClientConn) {
	switch state {
	case connectivity.Ready:
		s.logger.Debug("setting service to Serving")
		s.SetServingStatus(svrName, healthPb.HealthCheckResponse_SERVING)
	case connectivity.Idle:
		conn.Connect()
		s.logger.Debug("idle, forcing reconnect")
		s.SetServingStatus(svrName, healthPb.HealthCheckResponse_NOT_SERVING)
	default:
		s.logger.Debug("setting service to Not-Serving")
		s.SetServingStatus(svrName, healthPb.HealthCheckResponse_NOT_SERVING)
	}
}

func (s *HealthHandler) initStatus() {
	for svrName := range s.svrMap {
		s.SetServingStatus(svrName, healthPb.HealthCheckResponse_NOT_SERVING)
	}
}

func (s *HealthHandler) checkOverallStatus(ctx context.Context) (*healthPb.HealthCheckResponse, error) {
	for svrName := range s.svrMap {
		resp, err := s.Server.Check(ctx, &healthPb.HealthCheckRequest{Service: svrName})
		if err != nil || resp.Status != healthPb.HealthCheckResponse_SERVING {
			return resp, err
		}
	}
	return &healthPb.HealthCheckResponse{Status: healthPb.HealthCheckResponse_SERVING}, nil
}

func (s *HealthHandler) connect(svrName string, addr string) {
	s.svrMap[svrName], _ = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
