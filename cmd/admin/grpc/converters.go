package grpc

import (
	"fmt"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/schemas/formats"
	v1 "github.com/ecumenos-social/schemas/proto/gen/common/v1"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkit/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

func convertProtoPaginationToPagination(p *v1.Pagination) *types.Pagination {
	if p == nil {
		return types.NewPagination(nil, nil)
	}

	return types.NewPagination(p.Limit, p.Offset)
}

func convertProtoPersonalDataNodeStatusToPersonalDataNodeStatus(status pbv1.PersonalDataNode_Status) models.PersonalDataNodeStatus {
	switch status {
	case pbv1.PersonalDataNode_STATUS_APPROVED:
		return models.PersonalDataNodeStatusApproved
	case pbv1.PersonalDataNode_STATUS_PENDING:
		return models.PersonalDataNodeStatusPending
	case pbv1.PersonalDataNode_STATUS_REJECTED:
		return models.PersonalDataNodeStatusRejected
	}

	return ""
}

func convertPersonalDataNodeToProtoPersonalDataNode(pdn *models.PersonalDataNode) *pbv1.PersonalDataNode {
	var lastPingedAt string
	if pdn.LastPingedAt.Valid {
		lastPingedAt = formats.FormatDateTime(pdn.LastPingedAt.Time)
	}
	var status pbv1.PersonalDataNode_Status
	switch pdn.Status {
	case models.PersonalDataNodeStatusApproved:
		status = pbv1.PersonalDataNode_STATUS_APPROVED
	case models.PersonalDataNodeStatusPending:
		status = pbv1.PersonalDataNode_STATUS_PENDING
	case models.PersonalDataNodeStatusRejected:
		status = pbv1.PersonalDataNode_STATUS_REJECTED
	}

	return &pbv1.PersonalDataNode{
		Id:                   fmt.Sprint(pdn.ID),
		CreatedAt:            formats.FormatDateTime(pdn.CreatedAt),
		LastModifiedAt:       formats.FormatDateTime(pdn.LastModifiedAt),
		NwId:                 fmt.Sprint(pdn.NetworkWardenID),
		Address:              pdn.Address,
		Label:                pdn.Label,
		Name:                 pdn.Name,
		Description:          pdn.Description,
		Location:             convertLocationToProtoLocation(pdn.Location),
		AccountsCapacity:     pdn.AccountsCapacity,
		Alive:                pdn.Alive,
		LastPingedAt:         lastPingedAt,
		IsOpen:               pdn.IsOpen,
		IsInviteCodeRequired: pdn.IsInviteCodeRequired,
		OwnerHolderId:        fmt.Sprint(pdn.HolderID),
		Url:                  pdn.URL,
		Version:              pdn.Version,
		RateLimit: convertRateLimitToProtoRateLimit(&types.RateLimit{
			MaxRequests: pdn.RateLimitMaxRequests,
			Interval:    pdn.RateLimitInterval,
		}),
		CrawlRateLimit: convertRateLimitToProtoRateLimit(&types.RateLimit{
			MaxRequests: pdn.CrawlRateLimitMaxRequests,
			Interval:    pdn.CrawlRateLimitInterval,
		}),
		Status: status,
	}
}

func convertLocationToProtoLocation(l *models.Location) *v1.Geolocation {
	if l == nil {
		return nil
	}
	return &v1.Geolocation{
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
	}
}

func convertRateLimitToProtoRateLimit(rl *types.RateLimit) *v1.RateLimit {
	if rl == nil {
		return nil
	}
	return &v1.RateLimit{
		MaxRequests: rl.MaxRequests,
		Interval:    durationpb.New(rl.Interval),
	}
}
