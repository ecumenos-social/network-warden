package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	grpcutils "github.com/ecumenos-social/grpc-utils"
	"github.com/ecumenos-social/network-warden/converters"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/holders"
	"github.com/ecumenos-social/network-warden/services/jwt"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	networkwardens "github.com/ecumenos-social/network-warden/services/network-wardens"
	personaldatanodes "github.com/ecumenos-social/network-warden/services/personal-data-nodes"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkit/types"
	"github.com/ecumenos-social/toolkit/validators"
	"github.com/ecumenos-social/toolkitfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pbv1.NetworkWardenServiceServer

	jwt                      jwt.Service
	hs                       holders.Service
	auth                     auth.Service
	emailer                  emailer.Service
	smsSender                smssender.Service
	networkNodesService      networknodes.Service
	personalDataNodesService personaldatanodes.Service
	networkWardensService    networkwardens.Service
	logger                   *zap.Logger

	networkWardenID int64
}

var _ pbv1.NetworkWardenServiceServer = (*Handler)(nil)

type handlerParams struct {
	fx.In

	AppConfig                *toolkitfx.GenericAppConfig
	HoldersService           holders.Service
	HolderSessionsService    auth.Service
	AuthService              jwt.Service
	EmailerService           emailer.Service
	SMSSenderService         smssender.Service
	NetworkNodesService      networknodes.Service
	PersonalDataNodesService personaldatanodes.Service
	NetworkWardensService    networkwardens.Service
	Logger                   *zap.Logger
}

func NewHandler(params handlerParams) *Handler {
	return &Handler{
		hs:                       params.HoldersService,
		auth:                     params.HolderSessionsService,
		jwt:                      params.AuthService,
		emailer:                  params.EmailerService,
		smsSender:                params.SMSSenderService,
		networkNodesService:      params.NetworkNodesService,
		personalDataNodesService: params.PersonalDataNodesService,
		networkWardensService:    params.NetworkWardensService,
		logger:                   params.Logger,

		networkWardenID: params.AppConfig.ID,
	}
}

func (h *Handler) customizeLogger(ctx context.Context, operationName string) *zap.Logger {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return h.logger
	}

	l := h.logger.With(
		zap.String("operation-name", operationName),
		zap.Int64("network-warden-id", h.networkWardenID),
	)
	if corrID := md.Get("correlation-id"); len(corrID) > 0 {
		l = l.With(zap.String("correlation-id", corrID[0]))
	}
	if ip := grpcutils.ExtractRemoteIPAddress(ctx); ip != "" {
		l = l.With(zap.String("remote-ip-address", ip))
	}

	return l
}

func (h *Handler) CheckEmails(ctx context.Context, req *pbv1.NetworkWardenServiceCheckEmailsRequest) (*pbv1.NetworkWardenServiceCheckEmailsResponse, error) {
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

	return &pbv1.NetworkWardenServiceCheckEmailsResponse{
		Valid:   len(errs) == 0,
		Results: errs,
	}, nil
}

func (h *Handler) CheckPhoneNumbers(ctx context.Context, req *pbv1.NetworkWardenServiceCheckPhoneNumbersRequest) (*pbv1.NetworkWardenServiceCheckPhoneNumbersResponse, error) {
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

	return &pbv1.NetworkWardenServiceCheckPhoneNumbersResponse{
		Valid:   len(errs) == 0,
		Results: errs,
	}, nil
}

func (h *Handler) RegisterHolder(ctx context.Context, req *pbv1.NetworkWardenServiceRegisterHolderRequest) (*pbv1.NetworkWardenServiceRegisterHolderResponse, error) {
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
	if req.AvatarImageUrl != nil {
		params.AvatarImageURL = req.AvatarImageUrl
	}
	holder, err := h.hs.Insert(ctx, logger, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create holder entity (error = %v)", err.Error())
	}

	token, refreshToken, err := h.createSession(ctx, logger, holder.ID, grpcutils.ExtractRemoteIPAddress(ctx), req.RemoteMacAddress)
	if err != nil {
		return nil, err
	}

	approach := pbv1.NetworkWardenServiceConfirmationApproach_NETWORK_WARDEN_SERVICE_CONFIRMATION_APPROACH_PHONE_NUMBER
	if len(holder.Emails) > 0 {
		approach = pbv1.NetworkWardenServiceConfirmationApproach_NETWORK_WARDEN_SERVICE_CONFIRMATION_APPROACH_EMAIL
	}
	if err := h.sendConfirmationMessage(ctx, logger, approach, holder); err != nil {
		return nil, status.Errorf(codes.Internal, "failed send confirmation (error = %v)", err.Error())
	}

	return &pbv1.NetworkWardenServiceRegisterHolderResponse{
		Token:                token,
		RefreshToken:         refreshToken,
		ConfirmationApproach: approach,
	}, nil
}

func (h *Handler) createSession(ctx context.Context, logger *zap.Logger, holderID int64, ip string, mac *string) (string, string, error) {
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

func (h *Handler) validateRegisterHolderRequest(ctx context.Context, logger *zap.Logger, req *pbv1.NetworkWardenServiceRegisterHolderRequest) error {
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

func (h *Handler) sendConfirmationMessage(ctx context.Context, logger *zap.Logger, approach pbv1.NetworkWardenServiceConfirmationApproach, holder *models.Holder) error {
	if approach == pbv1.NetworkWardenServiceConfirmationApproach_NETWORK_WARDEN_SERVICE_CONFIRMATION_APPROACH_EMAIL {
		if err := h.emailer.SendConfirmationOfRegistration(ctx, logger, holder.Emails[0], holder.Emails[0], holder.ConfirmationCode); err != nil {
			return err
		}
		return nil
	}
	if approach == pbv1.NetworkWardenServiceConfirmationApproach_NETWORK_WARDEN_SERVICE_CONFIRMATION_APPROACH_PHONE_NUMBER {
		if err := h.smsSender.Send(ctx); err != nil {
			return err
		}
	}

	return errorwrapper.New("unknown approach for sending confirmation of registration code")
}

func (h *Handler) parseToken(ctx context.Context, logger *zap.Logger, token string, remoteMacAddress *string, scope jwt.TokenScope) (*models.HolderSession, error) {
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
	if remoteMacAddress != nil {
		logger = logger.With(zap.String("incoming-remote-mac-address", *remoteMacAddress))
		if hs.RemoteMACAddress.Valid && hs.RemoteMACAddress.String != *remoteMacAddress {
			logger.Error("remote MAC address doesn't match with session's remote MAC address", zap.String("session-remote-mac-address", hs.RemoteMACAddress.String))
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		}
	}

	return hs, nil
}

func (h *Handler) ConfirmHolderRegistration(ctx context.Context, req *pbv1.NetworkWardenServiceConfirmHolderRegistrationRequest) (*pbv1.NetworkWardenServiceConfirmHolderRegistrationResponse, error) {
	logger := h.customizeLogger(ctx, "ConfirmHolderRegistration")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	logger = logger.With(zap.Int64("session-holder-id", hs.HolderID))
	if _, err := h.hs.Confirm(ctx, logger, hs.HolderID, req.ConfirmationCode); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to confirm holder registration (err = %v)", err.Error())
	}

	return &pbv1.NetworkWardenServiceConfirmHolderRegistrationResponse{Success: true}, nil
}

func (h *Handler) ResendConfirmationCode(ctx context.Context, req *pbv1.NetworkWardenServiceResendConfirmationCodeRequest) (*pbv1.NetworkWardenServiceResendConfirmationCodeResponse, error) {
	logger := h.customizeLogger(ctx, "ResendConfirmationCode")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	holder, err := h.hs.GetHolderByID(ctx, logger, hs.HolderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get holder for resending confirmation code, err = %v", err.Error())
	}
	if holder == nil {
		return nil, status.Error(codes.InvalidArgument, "can not found holder by token's information")
	}
	canSend, err := h.emailer.CanSendConfirmationOfRegistration(ctx, logger, holder.Emails[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify if service can send confirmation code, err = %v", err.Error())
	}
	if !canSend {
		return nil, status.Error(codes.InvalidArgument, "rate limit of sent emails was exceeded. please, try next time")
	}

	holder, err = h.hs.RegenerateConfirmationCode(ctx, logger, hs.HolderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to regenerate confirmation code, err = %v", err.Error())
	}
	if err := h.sendConfirmationMessage(ctx, logger, req.ConfirmationApproach, holder); err != nil {
		return nil, status.Errorf(codes.Internal, "failed send confirmation (error = %v)", err.Error())
	}

	return &pbv1.NetworkWardenServiceResendConfirmationCodeResponse{Success: true}, nil
}

func (h *Handler) LoginHolder(ctx context.Context, req *pbv1.NetworkWardenServiceLoginHolderRequest) (*pbv1.NetworkWardenServiceLoginHolderResponse, error) {
	logger := h.customizeLogger(ctx, "LoginHolder")
	defer logger.Info("request processed")

	holder, err := h.hs.GetHolderByEmailOrPhoneNumber(ctx, logger, req.Email, req.PhoneNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}
	logger = logger.With(zap.Int64("holder-id", holder.ID))

	if err := h.hs.ValidatePassword(ctx, logger, holder, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	token, refreshToken, err := h.createSession(ctx, logger, holder.ID, grpcutils.ExtractRemoteIPAddress(ctx), req.RemoteMacAddress)
	if err != nil {
		return nil, err
	}

	return &pbv1.NetworkWardenServiceLoginHolderResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) LogoutHolder(ctx context.Context, req *pbv1.NetworkWardenServiceLogoutHolderRequest) (*pbv1.NetworkWardenServiceLogoutHolderResponse, error) {
	logger := h.customizeLogger(ctx, "LogoutHolder")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	if err := h.auth.MakeHolderSessionExpired(ctx, logger, hs.HolderID, hs); err != nil {
		return nil, status.Errorf(codes.Internal, "failed modify session (error = %v)", err.Error())
	}

	return &pbv1.NetworkWardenServiceLogoutHolderResponse{Success: true}, nil
}

func (h *Handler) RefreshHolderToken(ctx context.Context, req *pbv1.NetworkWardenServiceRefreshHolderTokenRequest) (*pbv1.NetworkWardenServiceRefreshHolderTokenResponse, error) {
	logger := h.customizeLogger(ctx, "RefreshHolderToken")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.RefreshToken, req.RemoteMacAddress, jwt.TokenScopeRefresh)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

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

	return &pbv1.NetworkWardenServiceRefreshHolderTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) ChangeHolderPassword(ctx context.Context, req *pbv1.NetworkWardenServiceChangeHolderPasswordRequest) (*pbv1.NetworkWardenServiceChangeHolderPasswordResponse, error) {
	logger := h.customizeLogger(ctx, "ChangeHolderPassword")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	holder, err := h.hs.GetHolderByID(ctx, logger, hs.HolderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		logger.Error("holder not found")
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}
	if err := h.hs.ValidatePassword(ctx, logger, holder, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}
	if err := h.hs.ChangePassword(ctx, logger, holder, req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to change holder's password")
	}

	return &pbv1.NetworkWardenServiceChangeHolderPasswordResponse{Success: true}, nil
}

func (h *Handler) ModifyHolder(ctx context.Context, req *pbv1.NetworkWardenServiceModifyHolderRequest) (*pbv1.NetworkWardenServiceModifyHolderResponse, error) {
	logger := h.customizeLogger(ctx, "ModifyHolder")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	holder, err := h.hs.GetHolderByID(ctx, logger, hs.HolderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		logger.Error("holder not found")
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}

	params := &holders.ModifyParams{
		AvatarImageURL: req.AvatarImageUrl,
		Countries:      req.Countries,
		Languages:      req.Languages,
	}
	if _, err := h.hs.Modify(ctx, logger, holder, params); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to modify holder")
	}

	return &pbv1.NetworkWardenServiceModifyHolderResponse{Success: true}, nil
}

func (h *Handler) GetHolder(ctx context.Context, req *pbv1.NetworkWardenServiceGetHolderRequest) (*pbv1.NetworkWardenServiceGetHolderResponse, error) {
	logger := h.customizeLogger(ctx, "GetHolder")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	holderID, err := strconv.ParseInt(req.HolderId, 10, 64)
	if err != nil {
		logger.Error("invalid ID", zap.Error(err), zap.String("incoming-holder-id", req.HolderId))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}
	holder, err := h.hs.GetHolderByID(ctx, logger, holderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		logger.Error("holder not found")
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}

	return &pbv1.NetworkWardenServiceGetHolderResponse{Data: converters.ConvertHolderToProtoHolder(holder)}, nil
}

func (h *Handler) DeleteHolder(ctx context.Context, req *pbv1.NetworkWardenServiceDeleteHolderRequest) (*pbv1.NetworkWardenServiceDeleteHolderResponse, error) {
	logger := h.customizeLogger(ctx, "DeleteHolder")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))

	holder, err := h.hs.GetHolderByID(ctx, logger, hs.HolderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get holder, err=%v", err.Error())
	}
	if holder == nil {
		logger.Error("holder not found")
		return nil, status.Error(codes.InvalidArgument, "holder not found")
	}
	if err := h.hs.ValidatePassword(ctx, logger, holder, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	if err := h.hs.Delete(ctx, logger, holder.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete holder, err=%v", err.Error())
	}

	return &pbv1.NetworkWardenServiceDeleteHolderResponse{Success: true}, nil
}

func (h *Handler) isHolderConfirmed(ctx context.Context, logger *zap.Logger, holderID int64) (bool, error) {
	holder, err := h.hs.GetHolderByID(ctx, logger, holderID)
	if err != nil {
		logger.Error("failed to get holder by ID", zap.Error(err))
		return false, err
	}
	if holder == nil {
		logger.Error("holder is not found by ID")
		return false, errorwrapper.New("holder is not found")
	}

	return holder.Confirmed, nil
}

func (h *Handler) GetPersonalDataNodesList(ctx context.Context, req *pbv1.NetworkWardenServiceGetPersonalDataNodesListRequest) (*pbv1.NetworkWardenServiceGetPersonalDataNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetPersonalDataNodesList")
	defer logger.Info("request processed")

	if req.Token == nil && req.OnlyMy != nil && *req.OnlyMy {
		return nil, status.Error(codes.InvalidArgument, "token is required if only_my filter is used")
	}
	var holderID int64
	if req.Token != nil {
		hs, err := h.parseToken(ctx, logger, *req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
		if err != nil {
			return nil, err
		}
		logger = logger.With(zap.Int64("holder-id", hs.HolderID))
		holderID = hs.HolderID
	}

	pdns, err := h.personalDataNodesService.GetList(
		ctx,
		logger,
		holderID,
		converters.ConvertProtoPaginationToPagination(req.Pagination),
		req.OnlyMy != nil && *req.OnlyMy,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get personal data nodes list")
	}
	data := make([]*pbv1.PersonalDataNode, 0, len(pdns))
	for _, pdn := range pdns {
		data = append(data, converters.ConvertPersonalDataNodeToProtoPersonalDataNode(pdn))
	}

	return &pbv1.NetworkWardenServiceGetPersonalDataNodesListResponse{
		Data: data,
	}, nil
}

func (h *Handler) JoinPersonalDataNodeRegistrationWaitlist(ctx context.Context, req *pbv1.NetworkWardenServiceJoinPersonalDataNodeRegistrationWaitlistRequest) (*pbv1.NetworkWardenServiceJoinPersonalDataNodeRegistrationWaitlistResponse, error) {
	logger := h.customizeLogger(ctx, "JoinPersonalDataNodeRegistrationWaitlist")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, &req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))
	if isConfirmed, err := h.isHolderConfirmed(ctx, logger, hs.HolderID); !isConfirmed || err != nil {
		if !isConfirmed {
			return nil, status.Error(codes.PermissionDenied, "holder is not confirmed")
		}
		return nil, status.Errorf(codes.Internal, "failed to validate holder if holder is confirmed, err = %v", err.Error())
	}

	pdn, err := h.personalDataNodesService.Insert(ctx, logger, &personaldatanodes.InsertParams{
		HolderID:        hs.HolderID,
		NetworkWardenID: h.networkWardenID,
		Name:            req.Name,
		Description:     req.Description,
		Label:           req.Label,
		Location: &models.Location{
			Longitude: req.Location.Longitude,
			Latitude:  req.Location.Latitude,
		},
		URL: req.Url,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert personal data node, err = %v", err.Error())
	}

	return &pbv1.NetworkWardenServiceJoinPersonalDataNodeRegistrationWaitlistResponse{
		Success: true,
		Id:      fmt.Sprint(pdn.ID),
	}, nil
}

func (h *Handler) ActivatePersonalDataNode(ctx context.Context, req *pbv1.NetworkWardenServiceActivatePersonalDataNodeRequest) (*pbv1.NetworkWardenServiceActivatePersonalDataNodeResponse, error) {
	logger := h.customizeLogger(ctx, "ActivatePersonalDataNode")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, &req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid personal data node ID", zap.Error(err), zap.String("incoming-personal-data-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid personal data node ID")
	}

	_, apiKey, err := h.personalDataNodesService.Activate(ctx, logger, hs.HolderID, id)
	if err != nil {
		logger.Error("failed to confirm", zap.Error(err), zap.String("incoming-personal-data-node-id", req.Id))
		return nil, status.Error(codes.Internal, "failed to confirm")
	}

	return &pbv1.NetworkWardenServiceActivatePersonalDataNodeResponse{
		Success: true,
		ApiKey:  apiKey,
	}, nil
}

func (h *Handler) InitiatePersonalDataNode(ctx context.Context, req *pbv1.NetworkWardenServiceInitiatePersonalDataNodeRequest) (*pbv1.NetworkWardenServiceInitiatePersonalDataNodeResponse, error) {
	logger := h.customizeLogger(ctx, "InitiatePersonalDataNode")
	defer logger.Info("request processed")

	params := &personaldatanodes.InitiateParams{
		AccountsCapacity:     req.AccountsCapacity,
		IsOpen:               req.IsOpen,
		IsInviteCodeRequired: req.IsInviteCodeRequired,
		Version:              req.Version,
		RateLimit: &types.RateLimit{
			MaxRequests: req.RateLimit.MaxRequests,
			Interval:    req.RateLimit.Interval.AsDuration(),
		},
		CrawlRateLimit: &types.RateLimit{
			MaxRequests: req.CrawlRateLimit.MaxRequests,
			Interval:    req.CrawlRateLimit.Interval.AsDuration(),
		},
		IDGenNode: req.IdGenNode,
	}
	if err := h.personalDataNodesService.Initiate(ctx, logger, req.ApiKey, params); err != nil {
		logger.Error("failed to initiate", zap.Error(err), zap.String("incoming-personal-data-node-api-key", req.ApiKey))
		return nil, status.Error(codes.Internal, "failed to confirm")
	}

	return &pbv1.NetworkWardenServiceInitiatePersonalDataNodeResponse{Success: true}, nil
}

func (h *Handler) GetNetworkNodesList(ctx context.Context, req *pbv1.NetworkWardenServiceGetNetworkNodesListRequest) (*pbv1.NetworkWardenServiceGetNetworkNodesListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkNodesList")
	defer logger.Info("request processed")

	if req.Token == nil && req.OnlyMy != nil && *req.OnlyMy {
		return nil, status.Error(codes.InvalidArgument, "token is required if only_my filter is used")
	}
	var holderID int64
	if req.Token != nil {
		hs, err := h.parseToken(ctx, logger, *req.Token, req.RemoteMacAddress, jwt.TokenScopeAccess)
		if err != nil {
			return nil, err
		}
		logger = logger.With(zap.Int64("holder-id", hs.HolderID))
		holderID = hs.HolderID
	}

	nns, err := h.networkNodesService.GetList(
		ctx,
		logger,
		holderID,
		converters.ConvertProtoPaginationToPagination(req.Pagination),
		req.OnlyMy != nil && *req.OnlyMy,
	)
	if err != nil {
		logger.Error("failed to get list", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get network nodes list")
	}
	data := make([]*pbv1.NetworkNode, 0, len(nns))
	for _, nn := range nns {
		data = append(data, converters.ConvertNetworkNodeToProtoNetworkNode(nn))
	}

	return &pbv1.NetworkWardenServiceGetNetworkNodesListResponse{
		Data: data,
	}, nil
}

func (h *Handler) JoinNetworkNodeRegistrationWaitlist(ctx context.Context, req *pbv1.NetworkWardenServiceJoinNetworkNodeRegistrationWaitlistRequest) (*pbv1.NetworkWardenServiceJoinNetworkNodeRegistrationWaitlistResponse, error) {
	logger := h.customizeLogger(ctx, "JoinNetworkNodeRegistrationWaitlist")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, &req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))
	if isConfirmed, err := h.isHolderConfirmed(ctx, logger, hs.HolderID); !isConfirmed || err != nil {
		if !isConfirmed {
			return nil, status.Error(codes.PermissionDenied, "holder is not confirmed")
		}
		return nil, status.Errorf(codes.Internal, "failed to validate holder if holder is confirmed, err = %v", err.Error())
	}

	nn, err := h.networkNodesService.Insert(ctx, logger, &networknodes.InsertParams{
		HolderID:        hs.HolderID,
		NetworkWardenID: h.networkWardenID,
		Name:            req.Name,
		Description:     req.Description,
		DomainName:      req.DomainName,
		Location: &models.Location{
			Longitude: req.Location.Longitude,
			Latitude:  req.Location.Latitude,
		},
		URL: req.Url,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert network node, err = %v", err.Error())
	}

	return &pbv1.NetworkWardenServiceJoinNetworkNodeRegistrationWaitlistResponse{
		Success: true,
		Id:      fmt.Sprint(nn.ID),
	}, nil
}

func (h *Handler) ActivateNetworkNode(ctx context.Context, req *pbv1.NetworkWardenServiceActivateNetworkNodeRequest) (*pbv1.NetworkWardenServiceActivateNetworkNodeResponse, error) {
	logger := h.customizeLogger(ctx, "ActivateNetworkNode")
	defer logger.Info("request processed")

	hs, err := h.parseToken(ctx, logger, req.Token, &req.RemoteMacAddress, jwt.TokenScopeAccess)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.Int64("holder-id", hs.HolderID))
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		logger.Error("invalid network node ID", zap.Error(err), zap.String("incoming-network-node-id", req.Id))
		return nil, status.Error(codes.InvalidArgument, "invalid network node ID")
	}

	_, apiKey, err := h.networkNodesService.Activate(ctx, logger, hs.HolderID, id)
	if err != nil {
		logger.Error("failed to confirm", zap.Error(err), zap.String("incoming-network-node-id", req.Id))
		return nil, status.Error(codes.Internal, "failed to confirm")
	}

	return &pbv1.NetworkWardenServiceActivateNetworkNodeResponse{
		Success: true,
		ApiKey:  apiKey,
	}, nil
}

func (h *Handler) InitiateNetworkNode(ctx context.Context, req *pbv1.NetworkWardenServiceInitiateNetworkNodeRequest) (*pbv1.NetworkWardenServiceInitiateNetworkNodeResponse, error) {
	logger := h.customizeLogger(ctx, "InitiateNetworkNode")
	defer logger.Info("request processed")

	params := &networknodes.InitiateParams{
		AccountsCapacity:     req.AccountsCapacity,
		IsOpen:               req.IsOpen,
		IsInviteCodeRequired: req.IsInviteCodeRequired,
		Version:              req.Version,
		RateLimit: &types.RateLimit{
			MaxRequests: req.RateLimit.MaxRequests,
			Interval:    req.RateLimit.Interval.AsDuration(),
		},
		CrawlRateLimit: &types.RateLimit{
			MaxRequests: req.CrawlRateLimit.MaxRequests,
			Interval:    req.CrawlRateLimit.Interval.AsDuration(),
		},
		IDGenNode: req.IdGenNode,
	}
	if err := h.networkNodesService.Initiate(ctx, logger, req.ApiKey, params); err != nil {
		logger.Error("failed to initiate", zap.Error(err), zap.String("incoming-network-node-api-key", req.ApiKey))
		return nil, status.Error(codes.Internal, "failed to confirm")
	}

	return &pbv1.NetworkWardenServiceInitiateNetworkNodeResponse{Success: true}, nil
}

func (h *Handler) GetNetworkWardensList(ctx context.Context, req *pbv1.NetworkWardenServiceGetNetworkWardensListRequest) (*pbv1.NetworkWardenServiceGetNetworkWardensListResponse, error) {
	logger := h.customizeLogger(ctx, "GetNetworkWardensList")
	defer logger.Info("request processed")

	nws, err := h.networkWardensService.GetList(
		ctx,
		logger,
		converters.ConvertProtoPaginationToPagination(req.Pagination),
	)
	if err != nil {
		logger.Error("failed to get list", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get network wardens list")
	}
	data := make([]*pbv1.NetworkWarden, 0, len(nws))
	for _, nw := range nws {
		data = append(data, converters.ConvertNetworkWardenToProtoNetworkWarden(nw))
	}

	return &pbv1.NetworkWardenServiceGetNetworkWardensListResponse{
		Data: data,
	}, nil
}

func (h *Handler) RegisterNetworkWarden(ctx context.Context, req *pbv1.NetworkWardenServiceRegisterNetworkWardenRequest) (*pbv1.NetworkWardenServiceRegisterNetworkWardenResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterNetworkWarden")
	defer logger.Info("request processed")

	params := &networkwardens.InsertParams{
		Name:        req.Name,
		Description: req.Description,
		Label:       req.Label,
		Location: &models.Location{
			Longitude: req.Location.Longitude,
			Latitude:  req.Location.Latitude,
		},
		IsOpen:      req.IsOpen,
		Version:     req.Version,
		URL:         req.Url,
		PDNCapacity: int64(req.PdnCapacity),
		NNCapacity:  int64(req.NnCapacity),
		RateLimit: &types.RateLimit{
			MaxRequests: req.RateLimit.MaxRequests,
			Interval:    req.RateLimit.Interval.AsDuration(),
		},
		IDGenNode: req.IdGenNode,
	}
	if _, err := h.networkWardensService.Insert(ctx, logger, params); err != nil {
		logger.Error("failed to register", zap.Error(err), zap.Any("request-body", req))
		return nil, status.Error(codes.Internal, "failed to register")
	}

	return &pbv1.NetworkWardenServiceRegisterNetworkWardenResponse{Success: true}, nil
}
