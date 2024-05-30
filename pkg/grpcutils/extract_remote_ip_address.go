package grpcutils

import (
	"context"
	"strings"

	"google.golang.org/grpc/peer"
)

func ExtractRemoteIPAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	ip := p.Addr.String()
	if parts := strings.Split(ip, ":"); len(parts) > 0 {
		return parts[0]
	}

	return ip
}
