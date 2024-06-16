package grpc

import (
	"fmt"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/schemas/formats"
	v1 "github.com/ecumenos-social/schemas/proto/gen/common/v1"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkit/types"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/durationpb"
)

func convertHolderToProtoHolder(holder *models.Holder) *pbv1.Holder {
	var avatarImageURL *string
	if holder.AvatarImageURL.Valid {
		avatarImageURL = lo.ToPtr(holder.AvatarImageURL.String)
	}

	return &pbv1.Holder{
		Id:             fmt.Sprint(holder.ID),
		Emails:         holder.Emails,
		PhoneNumbers:   holder.PhoneNumbers,
		AvatarImageUrl: avatarImageURL,
		Countries:      holder.Countries,
		Languages:      holder.Languages,
	}
}

func convertProtoPaginationToPagination(p *v1.Pagination) *types.Pagination {
	if p == nil {
		return types.NewPagination(nil, nil)
	}

	return types.NewPagination(p.Limit, p.Offset)
}

func convertNetworkNodeToProtoNetworkNode(nn *models.NetworkNode) *pbv1.NetworkNode {
	var lastPingedAt *string
	if nn.LastPingedAt.Valid {
		lastPingedAt = lo.ToPtr(formats.FormatDateTime(nn.LastPingedAt.Time))
	}
	var status pbv1.NetworkNode_Status
	switch nn.Status {
	case models.NetworkNodeStatusApproved:
		status = pbv1.NetworkNode_STATUS_APPROVED
	case models.NetworkNodeStatusPending:
		status = pbv1.NetworkNode_STATUS_PENDING
	case models.NetworkNodeStatusRejected:
		status = pbv1.NetworkNode_STATUS_REJECTED
	}

	return &pbv1.NetworkNode{
		Id:               fmt.Sprint(nn.ID),
		CreatedAt:        formats.FormatDateTime(nn.CreatedAt),
		LastModifiedAt:   formats.FormatDateTime(nn.LastModifiedAt),
		NwId:             fmt.Sprint(nn.NetworkWardenID),
		Name:             nn.Name,
		DomainName:       nn.DomainName,
		Location:         convertLocationToProtoLocation(nn.Location),
		AccountsCapacity: nn.AccountsCapacity,
		Alive:            nn.Alive,
		LastPingedAt:     lastPingedAt,
		IsOpen:           nn.IsOpen,
		OwnerHolderId:    fmt.Sprint(nn.HolderID),
		Url:              nn.URL,
		Version:          nn.Version,
		RateLimit: convertRateLimitToProtoRateLimit(&types.RateLimit{
			MaxRequests: nn.RateLimitMaxRequests,
			Interval:    nn.RateLimitInterval,
		}),
		CrawlRateLimit: convertRateLimitToProtoRateLimit(&types.RateLimit{
			MaxRequests: nn.CrawlRateLimitMaxRequests,
			Interval:    nn.CrawlRateLimitInterval,
		}),
		Status: status,
	}
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
		Id:               fmt.Sprint(pdn.ID),
		CreatedAt:        formats.FormatDateTime(pdn.CreatedAt),
		LastModifiedAt:   formats.FormatDateTime(pdn.LastModifiedAt),
		NwId:             fmt.Sprint(pdn.NetworkWardenID),
		Address:          pdn.Address,
		Label:            pdn.Label,
		Name:             pdn.Name,
		Description:      pdn.Description,
		Location:         convertLocationToProtoLocation(pdn.Location),
		AccountsCapacity: pdn.AccountsCapacity,
		Alive:            pdn.Alive,
		LastPingedAt:     lastPingedAt,
		IsOpen:           pdn.IsOpen,
		OwnerHolderId:    fmt.Sprint(pdn.HolderID),
		Url:              pdn.URL,
		Version:          pdn.Version,
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
