package grpc

import (
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
)

type Handler struct {
	pbv1.AdminServiceServer
}

var _ pbv1.AdminServiceServer = (*Handler)(nil)

func NewHandler() *Handler {
	return &Handler{}
}
