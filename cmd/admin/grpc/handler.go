package grpc

import (
	"context"
	"fmt"

	grpcutils "github.com/ecumenos-social/grpc-utils"
	"github.com/ecumenos-social/network-warden/services/adminauth"
	"github.com/ecumenos-social/network-warden/services/admins"
	"github.com/ecumenos-social/network-warden/services/jwt"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pbv1.AdminServiceServer

	jwt    jwt.Service
	admins admins.Service
	auth   adminauth.Service
	logger *zap.Logger
}

var _ pbv1.AdminServiceServer = (*Handler)(nil)

type handlerParams struct {
	fx.In

	AdminsService        admins.Service
	AdminSessionsService adminauth.Service
	AuthService          jwt.Service
	Logger               *zap.Logger
}

func NewHandler(params handlerParams) *Handler {
	return &Handler{
		jwt:    params.AuthService,
		admins: params.AdminsService,
		auth:   params.AdminSessionsService,

		logger: params.Logger,
	}
}

func (h *Handler) customizeLogger(ctx context.Context, operationName string) *zap.Logger {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return h.logger
	}

	l := h.logger.With(
		zap.String("operation-name", operationName),
	)
	if corrID := md.Get("correlation-id"); len(corrID) > 0 {
		l = l.With(zap.String("correlation-id", corrID[0]))
	}
	if ip := grpcutils.ExtractRemoteIPAddress(ctx); ip != "" {
		l = l.With(zap.String("remote-ip-address", ip))
	}

	return l
}

func (h *Handler) createSession(ctx context.Context, logger *zap.Logger, adminID int64, ip string, mac *string) (string, string, error) {
	token, refreshToken, err := h.jwt.CreateTokens(ctx, logger, fmt.Sprint(adminID))
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "failed create tokens (error = %v)", err.Error())
	}

	_, err = h.auth.Insert(ctx, logger, &adminauth.InsertParams{
		AdminID:          adminID,
		Token:            token,
		RefreshToken:     refreshToken,
		RemoteIPAddress:  ip,
		RemoteMACAddress: mac,
	})
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "failed create session (error = %v)", err.Error())
	}

	return token, refreshToken, nil
}

func (h *Handler) LoginAdmin(ctx context.Context, req *pbv1.AdminServiceLoginAdminRequest) (*pbv1.AdminServiceLoginAdminResponse, error) {
	logger := h.customizeLogger(ctx, "LoginAdmin")
	defer logger.Info("request processed")

	admin, err := h.admins.GetHolderByEmailOrPhoneNumber(ctx, logger, req.Email, req.PhoneNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get admin, err=%v", err.Error())
	}
	if admin == nil {
		return nil, status.Error(codes.InvalidArgument, "admin not found")
	}
	logger = logger.With(zap.Int64("admin-id", admin.ID))

	if err := h.admins.ValidatePassword(ctx, logger, admin, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	token, refreshToken, err := h.createSession(ctx, logger, admin.ID, grpcutils.ExtractRemoteIPAddress(ctx), lo.ToPtr(req.RemoteMacAddress))
	if err != nil {
		return nil, err
	}

	return &pbv1.AdminServiceLoginAdminResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) RefreshAdminToken(ctx context.Context, _ *pbv1.AdminServiceRefreshAdminTokenRequest) (*pbv1.AdminServiceRefreshAdminTokenResponse, error) {
	logger := h.customizeLogger(ctx, "RefreshAdminToken")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) LogoutAdmin(ctx context.Context, _ *pbv1.AdminServiceLogoutAdminRequest) (*pbv1.AdminServiceLogoutAdminResponse, error) {
	logger := h.customizeLogger(ctx, "LogoutAdmin")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) ChangeAdminPassword(ctx context.Context, _ *pbv1.AdminServiceChangeAdminPasswordRequest) (*pbv1.AdminServiceChangeAdminPasswordResponse, error) {
	logger := h.customizeLogger(ctx, "ChangeAdminPassword")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetPersonalDataNodesList(ctx context.Context, _ *pbv1.AdminServiceGetPersonalDataNodesListRequest) (*pbv1.AdminServiceGetPersonalDataNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodesList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetPersonalDataNodeByID(ctx context.Context, _ *pbv1.AdminServiceGetPersonalDataNodeByIDRequest) (*pbv1.AdminServiceGetPersonalDataNodeByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodeByID")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) SetPersonalDataNodeStatus(ctx context.Context, _ *pbv1.AdminServiceSetPersonalDataNodeStatusRequest) (*pbv1.AdminServiceSetPersonalDataNodeStatusResponse, error) {
	logger := h.customizeLogger(ctx, "SetPersonalDataNodeStatus")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkNodesList(ctx context.Context, _ *pbv1.AdminServiceGetNetworkNodesListRequest) (*pbv1.AdminServiceGetNetworkNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodesList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkNodeByID(ctx context.Context, _ *pbv1.AdminServiceGetNetworkNodeByIDRequest) (*pbv1.AdminServiceGetNetworkNodeByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodeByID")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) SetNetworkNodeStatus(ctx context.Context, _ *pbv1.AdminServiceSetNetworkNodeStatusRequest) (*pbv1.AdminServiceSetNetworkNodeStatusResponse, error) {
	logger := h.customizeLogger(ctx, "SetNetworkNodeStatus")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkWardensList(ctx context.Context, _ *pbv1.AdminServiceGetNetworkWardensListRequest) (*pbv1.AdminServiceGetNetworkWardensListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardensList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkWardenByID(ctx context.Context, _ *pbv1.AdminServiceGetNetworkWardenByIDRequest) (*pbv1.AdminServiceGetNetworkWardenByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardenByID")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}
