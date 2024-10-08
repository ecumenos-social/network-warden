package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	grpcutils "github.com/ecumenos-social/grpc-utils"
	"github.com/ecumenos-social/network-warden/converters"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/adminauth"
	"github.com/ecumenos-social/network-warden/services/admins"
	"github.com/ecumenos-social/network-warden/services/jwt"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	networkwardens "github.com/ecumenos-social/network-warden/services/network-wardens"
	personaldatanodes "github.com/ecumenos-social/network-warden/services/personal-data-nodes"
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

	jwt                      jwt.Service
	admins                   admins.Service
	auth                     adminauth.Service
	logger                   *zap.Logger
	personalDataNodesService personaldatanodes.Service
	networkNodesService      networknodes.Service
	networkWardenService     networkwardens.Service
}

var _ pbv1.AdminServiceServer = (*Handler)(nil)

type handlerParams struct {
	fx.In

	AdminsService            admins.Service
	AdminSessionsService     adminauth.Service
	AuthService              jwt.Service
	Logger                   *zap.Logger
	PersonalDataNodesService personaldatanodes.Service
	NetworkNodesService      networknodes.Service
	NetworkWardenService     networkwardens.Service
}

func NewHandler(params handlerParams) *Handler {
	return &Handler{
		jwt:                      params.AuthService,
		admins:                   params.AdminsService,
		auth:                     params.AdminSessionsService,
		personalDataNodesService: params.PersonalDataNodesService,
		networkNodesService:      params.NetworkNodesService,
		networkWardenService:     params.NetworkWardenService,

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

func (h *Handler) parseToken(ctx context.Context, logger *zap.Logger, token string, remoteMacAddress *string, scope jwt.TokenScope) (*models.AdminSession, error) {
	t, err := h.jwt.DecodeToken(logger, token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	anyAdminID, ok := t.Get("sub")
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, it doesn't contain admin information")
	}

	if _, err := strconv.ParseInt(fmt.Sprint(anyAdminID), 10, 64); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, admin information is formatted incorrectly")
	}

	as, err := h.auth.GetAdminSessionByToken(ctx, logger, token, scope)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if as.RemoteIPAddress.Valid && as.RemoteIPAddress.String != grpcutils.ExtractRemoteIPAddress(ctx) {
		logger.Error("remote IP address doesn't match with session's remote IP address", zap.String("session-remote-ip-address", as.RemoteIPAddress.String), zap.String("incoming-remote-ip-address", grpcutils.ExtractRemoteIPAddress(ctx)))
		return nil, status.Error(codes.PermissionDenied, "no permissions")
	}
	if remoteMacAddress != nil {
		logger = logger.With(zap.String("incoming-remote-mac-address", *remoteMacAddress))
		if as.RemoteMACAddress.Valid && as.RemoteMACAddress.String != *remoteMacAddress {
			logger.Error("remote MAC address doesn't match with session's remote MAC address", zap.String("session-remote-mac-address", as.RemoteMACAddress.String))
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		}
	}

	return as, nil
}

func (h *Handler) LoginAdmin(ctx context.Context, req *pbv1.AdminServiceLoginAdminRequest) (*pbv1.AdminServiceLoginAdminResponse, error) {
	logger := h.customizeLogger(ctx, "LoginAdmin")
	defer logger.Info("request processed")

	admin, err := h.admins.GetAdminByEmailOrPhoneNumber(ctx, logger, req.Email, req.PhoneNumber)
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

func (h *Handler) RefreshAdminToken(ctx context.Context, req *pbv1.AdminServiceRefreshAdminTokenRequest) (*pbv1.AdminServiceRefreshAdminTokenResponse, error) {
	logger := h.customizeLogger(ctx, "RefreshAdminToken")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.RefreshToken, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeRefresh)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))

	token, refreshToken, err := h.jwt.CreateTokens(ctx, logger, fmt.Sprint(as.AdminID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create tokens (error = %v)", err.Error())
	}

	as.Token = token
	as.RefreshToken = refreshToken
	as.ExpiredAt = sql.NullTime{
		Time:  h.auth.GetExpiredAtForAdminSession(),
		Valid: true,
	}
	if err := h.auth.ModifyAdminSession(ctx, logger, as.AdminID, as); err != nil {
		return nil, status.Errorf(codes.Internal, "failed modify session (error = %v)", err.Error())
	}

	return &pbv1.AdminServiceRefreshAdminTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) LogoutAdmin(ctx context.Context, req *pbv1.AdminServiceLogoutAdminRequest) (*pbv1.AdminServiceLogoutAdminResponse, error) {
	logger := h.customizeLogger(ctx, "LogoutAdmin")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))

	if err := h.auth.MakeAdminSessionExpired(ctx, logger, as.AdminID, as); err != nil {
		return nil, status.Errorf(codes.Internal, "failed modify session (error = %v)", err.Error())
	}

	return &pbv1.AdminServiceLogoutAdminResponse{Success: true}, nil
}

func (h *Handler) ChangeAdminPassword(ctx context.Context, req *pbv1.AdminServiceChangeAdminPasswordRequest) (*pbv1.AdminServiceChangeAdminPasswordResponse, error) {
	logger := h.customizeLogger(ctx, "ChangeAdminPassword")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))

	admin, err := h.admins.GetAdminByID(ctx, logger, as.AdminID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get admin, err=%v", err.Error())
	}
	if admin == nil {
		logger.Error("admin not found")
		return nil, status.Error(codes.InvalidArgument, "admin not found")
	}
	if err := h.admins.ValidatePassword(ctx, logger, admin, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	if err := h.admins.ChangePassword(ctx, logger, admin, req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to change admin's password")
	}

	return &pbv1.AdminServiceChangeAdminPasswordResponse{Success: true}, nil
}

func (h *Handler) GetPersonalDataNodesList(ctx context.Context, req *pbv1.AdminServiceGetPersonalDataNodesListRequest) (*pbv1.AdminServiceGetPersonalDataNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodesList")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	pdns, err := h.personalDataNodesService.GetList(ctx, logger, 0, converters.ConvertProtoPaginationToPagination(req.Pagination), false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get PDNs list, err=%v", err.Error())
	}
	data := make([]*pbv1.PersonalDataNode, 0, len(pdns))
	for _, pdn := range pdns {
		data = append(data, converters.ConvertPersonalDataNodeToProtoPersonalDataNode(pdn))
	}

	return &pbv1.AdminServiceGetPersonalDataNodesListResponse{
		Data: data,
	}, nil
}

func (h *Handler) GetPersonalDataNodeByID(ctx context.Context, req *pbv1.AdminServiceGetPersonalDataNodeByIDRequest) (*pbv1.AdminServiceGetPersonalDataNodeByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodeByID")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	pdnID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-personal-data-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	pdn, err := h.personalDataNodesService.GetByID(ctx, logger, pdnID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get PDN, err=%v", err.Error())
	}
	if pdn == nil {
		return &pbv1.AdminServiceGetPersonalDataNodeByIDResponse{
			Data: nil,
		}, nil
	}

	return &pbv1.AdminServiceGetPersonalDataNodeByIDResponse{
		Data: converters.ConvertPersonalDataNodeToProtoPersonalDataNode(pdn),
	}, nil
}

func (h *Handler) SetPersonalDataNodeStatus(ctx context.Context, req *pbv1.AdminServiceSetPersonalDataNodeStatusRequest) (*pbv1.AdminServiceSetPersonalDataNodeStatusResponse, error) {
	logger := h.customizeLogger(ctx, "SetPersonalDataNodeStatus")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	pdnID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-personal-data-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	if err := h.personalDataNodesService.SetStatusByID(ctx, logger, pdnID, converters.ConvertProtoPersonalDataNodeStatusToPersonalDataNodeStatus(req.Status)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to update personal data node status")
	}

	return &pbv1.AdminServiceSetPersonalDataNodeStatusResponse{Success: true}, nil
}

func (h *Handler) GetNetworkNodesList(ctx context.Context, req *pbv1.AdminServiceGetNetworkNodesListRequest) (*pbv1.AdminServiceGetNetworkNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodesList")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	nns, err := h.networkNodesService.GetList(ctx, logger, 0, converters.ConvertProtoPaginationToPagination(req.Pagination), false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get PDNs list, err=%v", err.Error())
	}
	data := make([]*pbv1.NetworkNode, 0, len(nns))
	for _, nn := range nns {
		data = append(data, converters.ConvertNetworkNodeToProtoNetworkNode(nn))
	}

	return &pbv1.AdminServiceGetNetworkNodesListResponse{
		Data: data,
	}, nil
}

func (h *Handler) GetNetworkNodeByID(ctx context.Context, req *pbv1.AdminServiceGetNetworkNodeByIDRequest) (*pbv1.AdminServiceGetNetworkNodeByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodeByID")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	nnID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-network-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	nn, err := h.networkNodesService.GetByID(ctx, logger, nnID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get NN, err=%v", err.Error())
	}
	if nn == nil {
		return &pbv1.AdminServiceGetNetworkNodeByIDResponse{
			Data: nil,
		}, nil
	}

	return &pbv1.AdminServiceGetNetworkNodeByIDResponse{
		Data: converters.ConvertNetworkNodeToProtoNetworkNode(nn),
	}, nil
}

func (h *Handler) SetNetworkNodeStatus(ctx context.Context, req *pbv1.AdminServiceSetNetworkNodeStatusRequest) (*pbv1.AdminServiceSetNetworkNodeStatusResponse, error) {
	logger := h.customizeLogger(ctx, "SetNetworkNodeStatus")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	pdnID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-network-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	if err := h.networkNodesService.SetStatusByID(ctx, logger, pdnID, converters.ConvertProtoNetworkNodeStatusToNetworkNodeStatus(req.Status)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to update network node status")
	}

	return &pbv1.AdminServiceSetNetworkNodeStatusResponse{Success: true}, nil
}

func (h *Handler) GetNetworkWardensList(ctx context.Context, req *pbv1.AdminServiceGetNetworkWardensListRequest) (*pbv1.AdminServiceGetNetworkWardensListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardensList")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	nws, err := h.networkWardenService.GetList(ctx, logger, converters.ConvertProtoPaginationToPagination(req.Pagination))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get NWs list, err=%v", err.Error())
	}
	data := make([]*pbv1.NetworkWarden, 0, len(nws))
	for _, nw := range nws {
		data = append(data, converters.ConvertNetworkWardenToProtoNetworkWarden(nw))
	}

	return &pbv1.AdminServiceGetNetworkWardensListResponse{
		Data: data,
	}, nil
}

func (h *Handler) GetNetworkWardenByID(ctx context.Context, req *pbv1.AdminServiceGetNetworkWardenByIDRequest) (*pbv1.AdminServiceGetNetworkWardenByIDResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardenByID")
	defer logger.Info("request processed")

	as, err := h.parseToken(ctx, logger, req.Token, lo.ToPtr(req.RemoteMacAddress), jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("admin-id", as.AdminID))
	nwID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-network-warden-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	nw, err := h.networkWardenService.GetByID(ctx, logger, nwID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get NW, err=%v", err.Error())
	}
	if nw == nil {
		return &pbv1.AdminServiceGetNetworkWardenByIDResponse{
			Data: nil,
		}, nil
	}

	return &pbv1.AdminServiceGetNetworkWardenByIDResponse{
		Data: converters.ConvertNetworkWardenToProtoNetworkWarden(nw),
	}, nil
}
