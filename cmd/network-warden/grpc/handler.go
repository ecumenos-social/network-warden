package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/pkg/grpcutils"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/holders"
	"github.com/ecumenos-social/network-warden/services/jwt"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkit/validators"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pbv1.NetworkWardenServiceServer

	jwt       jwt.Service
	hs        holders.Service
	auth      auth.Service
	emailer   emailer.Service
	smsSender smssender.Service
	logger    *zap.Logger
}

var _ pbv1.NetworkWardenServiceServer = (*Handler)(nil)

type handlerParams struct {
	fx.In

	HoldersService        holders.Service
	HolderSessionsService auth.Service
	AuthService           jwt.Service
	EmailerService        emailer.Service
	SMSSenderService      smssender.Service
	Logger                *zap.Logger
}

func NewHandler(params handlerParams) *Handler {
	return &Handler{
		hs:        params.HoldersService,
		auth:      params.HolderSessionsService,
		jwt:       params.AuthService,
		emailer:   params.EmailerService,
		smsSender: params.SMSSenderService,
		logger:    params.Logger,
	}
}

func (h *Handler) customizeLogger(ctx context.Context, operationName string) *zap.Logger {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return h.logger
	}

	l := h.logger.With(zap.String("operation-name", operationName))
	if corrID := md.Get("correlation-id"); len(corrID) > 0 {
		l = l.With(zap.String("correlation-id", corrID[0]))
	}
	if ip := grpcutils.ExtractRemoteIPAddress(ctx); ip != "" {
		l = l.With(zap.String("remote-ip-address", ip))
	}

	return l
}

func (h *Handler) CheckEmails(ctx context.Context, req *pbv1.CheckEmailsRequest) (*pbv1.CheckEmailsResponse, error) {
	logger := h.customizeLogger(ctx, "CheckEmails")
	defer logger.Info("request processed")

	errs := make([]string, 0, len(req.Emails))
	for _, email := range req.Emails {
		if err := validators.ValidateEmail(ctx, email); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := h.hs.CheckEmailsUsage(ctx, logger, req.Emails); err != nil {
		errs = append(errs, err.Error())
	}

	return &pbv1.CheckEmailsResponse{
		Valid:   len(errs) == 0,
		Results: errs,
	}, nil
}

func (h *Handler) CheckPhoneNumbers(ctx context.Context, req *pbv1.CheckPhoneNumbersRequest) (*pbv1.CheckPhoneNumbersResponse, error) {
	logger := h.customizeLogger(ctx, "CheckPhoneNumbers")
	defer logger.Info("request processed")

	errs := make([]string, 0, len(req.PhoneNumbers))
	for _, phoneNumber := range req.PhoneNumbers {
		if err := validators.ValidatePhoneNumber(ctx, phoneNumber); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := h.hs.CheckPhoneNumbersUsage(ctx, logger, req.PhoneNumbers); err != nil {
		errs = append(errs, err.Error())
	}

	return &pbv1.CheckPhoneNumbersResponse{
		Valid:   len(errs) == 0,
		Results: errs,
	}, nil
}

func (h *Handler) RegisterHolder(ctx context.Context, req *pbv1.RegisterHolderRequest) (*pbv1.RegisterHolderResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterHolder")
	defer logger.Info("request processed")

	if err := h.validateRegisterHolderRequest(ctx, logger, req); err != nil {
		return nil, err
	}

	params := &holders.InsertParams{
		Emails:       req.Emails,
		PhoneNumbers: req.PhoneNumbers,
		Countries:    req.Countries,
		Languages:    req.Languages,
		Password:     req.Password,
	}
	if req.AvatarImageUrl != "" {
		params.AvatarImageURL = &req.AvatarImageUrl
	}
	holder, err := h.hs.Insert(ctx, logger, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create holder entity (error = %v)", err.Error())
	}

	token, refreshToken, err := h.createSession(ctx, logger, holder.ID, grpcutils.ExtractRemoteIPAddress(ctx), req.RemoteMacAddress)
	if err != nil {
		return nil, err
	}
	approach, err := h.sendConfirmationMessage(ctx, logger, holder)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed send confirmation (error = %v)", err.Error())
	}

	return &pbv1.RegisterHolderResponse{
		Token:                token,
		RefreshToken:         refreshToken,
		ConfirmationApproach: approach,
	}, nil
}

func (h *Handler) createSession(ctx context.Context, logger *zap.Logger, holderID int64, ip, mac string) (string, string, error) {
	token, refreshToken, err := h.jwt.CreateTokens(ctx, logger, fmt.Sprint(holderID))
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "failed create tokens (error = %v)", err.Error())
	}

	_, err = h.auth.Insert(ctx, logger, &auth.InsertParams{
		HolderID:         holderID,
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

func (h *Handler) validateRegisterHolderRequest(ctx context.Context, logger *zap.Logger, req *pbv1.RegisterHolderRequest) error {
	if err := req.ValidateAll(); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid request (error = %v)", err.Error())
	}
	if len(req.Emails) == 0 && len(req.PhoneNumbers) == 0 {
		return status.Error(codes.InvalidArgument, "invalid request (you should set at least one email address or phone number)")
	}
	if len(req.Emails) > 0 {
		for _, email := range req.Emails {
			if err := validators.ValidateEmail(ctx, email); err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid request, invalid email (email: %v, error = %v)", email, err.Error())
			}
		}
		if err := h.hs.CheckEmailsUsage(ctx, logger, req.Emails); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid request, email in use (error = %v)", err.Error())
		}
	}
	if len(req.PhoneNumbers) > 0 {
		for _, phoneNumber := range req.PhoneNumbers {
			if err := validators.ValidatePhoneNumber(ctx, phoneNumber); err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid request, invalid phone number (phone_number: %v, error = %v)", phoneNumber, err.Error())
			}
		}
		if err := h.hs.CheckPhoneNumbersUsage(ctx, logger, req.PhoneNumbers); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid request (error = %v)", err.Error())
		}
	}
	for _, cc := range req.Countries {
		if err := validators.ValidateCountryCode(ctx, cc); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid request, invalid country code (country_code: %v, error = %v)", cc, err.Error())
		}
	}
	for _, lc := range req.Languages {
		if err := validators.ValidateLanguageCode(ctx, lc); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid request, invalid language code (language_code: %v, error = %v)", lc, err.Error())
		}
	}

	return nil
}

func (h *Handler) sendConfirmationMessage(ctx context.Context, logger *zap.Logger, holder *models.Holder) (pbv1.ConfirmationApproach, error) {
	if len(holder.Emails) > 0 {
		if err := h.emailer.SendConfirmationOfRegistration(ctx, logger, holder.Emails[0], holder.Emails[0], holder.ConfirmationCode); err != nil {
			return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_UNKNOWN_UNSPECIFIED, err
		}
		return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_EMAIL, nil
	}

	if err := h.smsSender.Send(ctx); err != nil {
		return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_UNKNOWN_UNSPECIFIED, err
	}

	return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_PHONE_NUMBER, nil
}

func (h *Handler) parseToken(ctx context.Context, logger *zap.Logger, token, remoteMacAddress string, scope jwt.TokenScope) (*models.HolderSession, error) {
	t, err := h.jwt.DecodeToken(logger, token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	anyHolderID, ok := t.Get("sub")
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, it doesn't contain holder information")
	}

	if _, err := strconv.ParseInt(fmt.Sprint(anyHolderID), 10, 64); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, holder information is formatted incorrectly")
	}

	hs, err := h.auth.GetHolderSessionByToken(ctx, logger, token, scope)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if hs.RemoteIPAddress.Valid && hs.RemoteIPAddress.String != grpcutils.ExtractRemoteIPAddress(ctx) {
		logger.Error("remote IP address doesn't match with session's remote IP address", zap.String("session-remote-ip-address", hs.RemoteIPAddress.String), zap.String("incoming-remote-ip-address", grpcutils.ExtractRemoteIPAddress(ctx)))
		return nil, status.Error(codes.PermissionDenied, "no permissions")
	}
	if hs.RemoteMACAddress.Valid && hs.RemoteMACAddress.String != remoteMacAddress {
		logger.Error("remote MAC address doesn't match with session's remote MAC address", zap.String("session-remote-mac-address", hs.RemoteMACAddress.String), zap.String("incoming-remote-mac-address", remoteMacAddress))
		return nil, status.Error(codes.PermissionDenied, "no permissions")
	}

	return hs, nil
}

func (h *Handler) ConfirmHolderRegistration(ctx context.Context, req *pbv1.ConfirmHolderRegistrationRequest) (*pbv1.ConfirmHolderRegistrationResponse, error) {
	logger := h.customizeLogger(ctx, "ConfirmHolderRegistration")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("session-holder-id", hs.HolderID))
	if _, err := h.hs.Confirm(ctx, logger, hs.HolderID, req.ConfirmationCode); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to confirm holder registration (err = %v)", err.Error())
	}

	return &pbv1.ConfirmHolderRegistrationResponse{Success: true}, nil
}

func (h *Handler) ResendConfirmationCode(ctx context.Context, _ *pbv1.ResendConfirmationCodeRequest) (*pbv1.ResendConfirmationCodeResponse, error) {
	logger := h.customizeLogger(ctx, "ResendConfirmationCode")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) LoginHolder(ctx context.Context, req *pbv1.LoginHolderRequest) (*pbv1.LoginHolderResponse, error) {
	logger := h.customizeLogger(ctx, "LoginHolder")
	defer logger.Info("request processed")

	holder, err := h.hs.GetHolderByEmailOrPhoneNumber(ctx, logger, req.Email, req.PhoneNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}
	if err := h.hs.ValidatePassword(ctx, logger, holder, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	token, refreshToken, err := h.createSession(ctx, logger, holder.ID, grpcutils.ExtractRemoteIPAddress(ctx), req.RemoteMacAddress)
	if err != nil {
		return nil, err
	}

	return &pbv1.LoginHolderResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) LogoutHolder(ctx context.Context, req *pbv1.LogoutHolderRequest) (*pbv1.LogoutHolderResponse, error) {
	logger := h.customizeLogger(ctx, "LogoutHolder")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	if err := h.auth.MakeHolderSessionExpired(ctx, logger, hs.HolderID, hs); err != nil {
		return nil, status.Errorf(codes.Internal, "failed modify session (error = %v)", err.Error())
	}

	return &pbv1.LogoutHolderResponse{Success: true}, nil
}

func (h *Handler) RefreshHolderToken(ctx context.Context, req *pbv1.RefreshHolderTokenRequest) (*pbv1.RefreshHolderTokenResponse, error) {
	logger := h.customizeLogger(ctx, "RefreshHolderToken")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.RefreshToken, req.RemoteMacAddress, jwt.TokenScopeRefresh)
	if err != nil {
		return nil, err
	}

	token, refreshToken, err := h.jwt.CreateTokens(ctx, logger, fmt.Sprint(hs.HolderID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create tokens (error = %v)", err.Error())
	}

	hs.Token = token
	hs.RefreshToken = refreshToken
	hs.ExpiredAt = sql.NullTime{
		Time:  h.auth.GetExpiredAtForHolderSession(),
		Valid: true,
	}
	if err := h.auth.ModifyHolderSession(ctx, logger, hs.HolderID, hs); err != nil {
		return nil, status.Errorf(codes.Internal, "failed modify session (error = %v)", err.Error())
	}

	return &pbv1.RefreshHolderTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) ChangeHolderPassword(ctx context.Context, _ *pbv1.ChangeHolderPasswordRequest) (*pbv1.ChangeHolderPasswordResponse, error) {
	logger := h.customizeLogger(ctx, "ChangeHolderPassword")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) ModifyHolder(ctx context.Context, _ *pbv1.ModifyHolderRequest) (*pbv1.ModifyHolderResponse, error) {
	logger := h.customizeLogger(ctx, "ModifyHolder")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetHolder(ctx context.Context, _ *pbv1.GetHolderRequest) (*pbv1.GetHolderResponse, error) {
	logger := h.customizeLogger(ctx, "GetHolder")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) DeleteHolder(ctx context.Context, _ *pbv1.DeleteHolderRequest) (*pbv1.DeleteHolderResponse, error) {
	logger := h.customizeLogger(ctx, "DeleteHolder")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetPersonalDataNodesList(ctx context.Context, _ *pbv1.GetPersonalDataNodesListRequest) (*pbv1.GetPersonalDataNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodesList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) JoinPersonalDataNodeRegistrationWaitlist(ctx context.Context, _ *pbv1.JoinPersonalDataNodeRegistrationWaitlistRequest) (*pbv1.JoinPersonalDataNodeRegistrationWaitlistResponse, error) {
	logger := h.customizeLogger(ctx, "JoinPersonalDataNodeRegistrationWaitlist")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterPersonalDataNode(ctx context.Context, _ *pbv1.RegisterPersonalDataNodeRequest) (*pbv1.RegisterPersonalDataNodeResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterPersonalDataNode")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkNodesList(ctx context.Context, _ *pbv1.GetNetworkNodesListRequest) (*pbv1.GetNetworkNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodesList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) JoinNetworkNodeRegistrationWaitlist(ctx context.Context, _ *pbv1.JoinNetworkNodeRegistrationWaitlistRequest) (*pbv1.JoinNetworkNodeRegistrationWaitlistResponse, error) {
	logger := h.customizeLogger(ctx, "JoinNetworkNodeRegistrationWaitlist")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterNetworkNode(ctx context.Context, _ *pbv1.RegisterNetworkNodeRequest) (*pbv1.RegisterNetworkNodeResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterNetworkNode")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkWardensList(ctx context.Context, _ *pbv1.GetNetworkWardensListRequest) (*pbv1.GetNetworkWardensListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardensList")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterNetworkWarden(ctx context.Context, _ *pbv1.RegisterNetworkWardenRequest) (*pbv1.RegisterNetworkWardenResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterNetworkWarden")
	defer logger.Info("request processed")

	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}
