package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	holdersessions "github.com/ecumenos-social/network-warden/services/holder-sessions"
	"github.com/ecumenos-social/network-warden/services/holders"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkit/validators"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pbv1.NetworkWardenServiceServer

	auth      auth.Service
	hs        holders.Service
	hss       holdersessions.Service
	emailer   emailer.Service
	smsSender smssender.Service
	logger    *zap.Logger
}

var _ pbv1.NetworkWardenServiceServer = (*Handler)(nil)

type handlerParams struct {
	fx.In

	HoldersService        holders.Service
	HolderSessionsService holdersessions.Service
	AuthService           auth.Service
	EmailerService        emailer.Service
	SMSSenderService      smssender.Service
	Logger                *zap.Logger
}

func NewHandler(params handlerParams) *Handler {
	return &Handler{
		hs:        params.HoldersService,
		hss:       params.HolderSessionsService,
		auth:      params.AuthService,
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
	if ip := extractRemoteIPAddress(ctx); ip != "" {
		l = l.With(zap.String("remote-ip-address", ip))
	}

	return l
}

func (h *Handler) CheckEmails(ctx context.Context, req *pbv1.CheckEmailsRequest) (*pbv1.CheckEmailsResponse, error) {
	errs := make([]string, 0, len(req.Emails))
	for _, email := range req.Emails {
		if err := validators.ValidateEmail(ctx, email); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if err := h.hs.CheckEmailsUsage(ctx, req.Emails); err != nil {
		errs = append(errs, err.Error())
	}
	var result = "ok"
	if len(errs) > 0 {
		result = fmt.Sprintf("errors: %s", strings.Join(errs, ", "))
	}

	return &pbv1.CheckEmailsResponse{
		Result: result,
	}, nil
}

func (h *Handler) RegisterHolder(ctx context.Context, req *pbv1.RegisterHolderRequest) (*pbv1.RegisterHolderResponse, error) {
	logger := h.customizeLogger(ctx, "RegisterHolder")
	if err := h.validateRegisterHolderRequest(ctx, req); err != nil {
		logger.Error("validation error", zap.Error(err))
		return nil, err
	}

	params := &holders.InsertParams{
		Emails:      req.Emails,
		PhoneNumber: req.PhoneNumbers,
		Countries:   req.Countries,
		Languages:   req.Languages,
		Password:    req.Password,
	}
	if req.AvatarImageUrl != "" {
		params.AvatarImageURL = &req.AvatarImageUrl
	}
	holder, err := h.hs.Insert(ctx, params)
	if err != nil {
		logger.Error("insert holder error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed create holder entity (error = %v)", err.Error())
	}

	token, refreshToken, err := h.auth.CreateTokens(ctx, fmt.Sprint(holder.ID))
	if err != nil {
		logger.Error("create token pair error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed create tokens (error = %v)", err.Error())
	}
	if _, err := h.hss.Insert(ctx, &holdersessions.InsertParams{
		HolderID:         holder.ID,
		Token:            token,
		RefreshToken:     refreshToken,
		RemoteIPAddress:  extractRemoteIPAddress(ctx),
		RemoteMACAddress: req.RemoteMacAddress,
	}); err != nil {
		logger.Error("insert holder session error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed create session (error = %v)", err.Error())
	}
	approach, err := h.sendConfirmationMessage(ctx, holder)
	if err != nil {
		logger.Error("send confirmation message error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed send confirmation (error = %v)", err.Error())
	}

	return &pbv1.RegisterHolderResponse{
		Token:                token,
		RefreshToken:         refreshToken,
		ConfirmationApproach: approach,
	}, nil
}

func extractRemoteIPAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if ok {
		return p.Addr.String()
	}

	return ""
}

func (h *Handler) validateRegisterHolderRequest(ctx context.Context, req *pbv1.RegisterHolderRequest) error {
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
		if err := h.hs.CheckEmailsUsage(ctx, req.Emails); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid request, email in use (error = %v)", err.Error())
		}
	}
	if len(req.PhoneNumbers) > 0 {
		for _, phoneNumber := range req.PhoneNumbers {
			if err := validators.ValidatePhoneNumber(ctx, phoneNumber); err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid request, invalid phone number (phone_number: %v, error = %v)", phoneNumber, err.Error())
			}
		}
		if err := h.hs.CheckPhoneNumbersUsage(ctx, req.PhoneNumbers); err != nil {
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

func (h *Handler) sendConfirmationMessage(ctx context.Context, holder *models.Holder) (pbv1.ConfirmationApproach, error) {
	if len(holder.Emails) > 0 {
		if err := h.emailer.Send(ctx); err != nil {
			return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_UNKNOWN_UNSPECIFIED, err
		}
		return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_EMAIL, nil
	}

	if err := h.smsSender.Send(ctx); err != nil {
		return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_UNKNOWN_UNSPECIFIED, err
	}

	return pbv1.ConfirmationApproach_CONFIRMATION_APPROACH_PHONE_NUMBER, nil
}

func (h *Handler) ConfirmHolderRegistration(context.Context, *pbv1.ConfirmHolderRegistrationRequest) (*pbv1.ConfirmHolderRegistrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) ResendConfirmationCode(context.Context, *pbv1.ResendConfirmationCodeRequest) (*pbv1.ResendConfirmationCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) LoginHolder(context.Context, *pbv1.LoginHolderRequest) (*pbv1.LoginHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) LogoutHolder(context.Context, *pbv1.LogoutHolderRequest) (*pbv1.LogoutHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RefreshHolderToken(context.Context, *pbv1.RefreshHolderTokenRequest) (*pbv1.RefreshHolderTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) ChangeHolderPassword(context.Context, *pbv1.ChangeHolderPasswordRequest) (*pbv1.ChangeHolderPasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) ModifyHolder(context.Context, *pbv1.ModifyHolderRequest) (*pbv1.ModifyHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetHolder(context.Context, *pbv1.GetHolderRequest) (*pbv1.GetHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) DeleteHolder(context.Context, *pbv1.DeleteHolderRequest) (*pbv1.DeleteHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetPersonalDataNodesList(context.Context, *pbv1.GetPersonalDataNodesListRequest) (*pbv1.GetPersonalDataNodesListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) JoinPersonalDataNodeRegistrationWaitlist(context.Context, *pbv1.JoinPersonalDataNodeRegistrationWaitlistRequest) (*pbv1.JoinPersonalDataNodeRegistrationWaitlistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterPersonalDataNode(context.Context, *pbv1.RegisterPersonalDataNodeRequest) (*pbv1.RegisterPersonalDataNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkNodesList(context.Context, *pbv1.GetNetworkNodesListRequest) (*pbv1.GetNetworkNodesListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) JoinNetworkNodeRegistrationWaitlist(context.Context, *pbv1.JoinNetworkNodeRegistrationWaitlistRequest) (*pbv1.JoinNetworkNodeRegistrationWaitlistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterNetworkNode(context.Context, *pbv1.RegisterNetworkNodeRequest) (*pbv1.RegisterNetworkNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) GetNetworkWardensList(context.Context, *pbv1.GetNetworkWardensListRequest) (*pbv1.GetNetworkWardensListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}

func (h *Handler) RegisterNetworkWarden(context.Context, *pbv1.RegisterNetworkWardenRequest) (*pbv1.RegisterNetworkWardenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
}
