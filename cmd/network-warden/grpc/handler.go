package grpc

import (
	"context"

	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pbv1.NetworkWardenServiceServer
}

var _ pbv1.NetworkWardenServiceServer = (*Handler)(nil)

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterHolder(context.Context, *pbv1.RegisterHolderRequest) (*pbv1.RegisterHolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method is not implemented")
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
