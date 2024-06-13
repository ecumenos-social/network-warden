package personaldatanodes

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkitfx"
	"go.uber.org/zap"
)

type Repository interface {
	InsertPersonalDataNode(ctx context.Context, pdn *models.PersonalDataNode) error
	GetPersonalDataNodeByLabel(ctx context.Context, label string) (*models.PersonalDataNode, error)
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.PersonalDataNode, error)
}

type service struct {
	repo                       Repository
	idgenerator                idgenerators.PersonalDataNodesIDGenerator
	networkWardenID            int64
	networkWardenAddressSuffix string
}

func New(config *toolkitfx.NetworkWardenAppConfig, repo Repository, g idgenerators.PersonalDataNodesIDGenerator) Service {
	return &service{
		repo:                       repo,
		idgenerator:                g,
		networkWardenID:            config.ID,
		networkWardenAddressSuffix: config.AddressSuffix,
	}
}

type InsertParams struct {
	HolderID        int64
	NetworkWardenID int64
	Name            string
	Description     string
	Label           string
	Location        *models.Location
	URL             string
}

func (s *service) buildAddress(label string) string {
	return fmt.Sprintf("#%s::%s", label, s.networkWardenAddressSuffix)
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.PersonalDataNode, error) {
	address := s.buildAddress(params.Label)
	logger = logger.With(
		zap.String("personal-data-node-label", params.Label),
		zap.String("personal-data-node-address", address),
		zap.String("personal-data-node-name", params.Name),
	)
	if pdn, err := s.repo.GetPersonalDataNodeByLabel(ctx, params.Label); err != nil || pdn != nil {
		if err != nil {
			logger.Error("failed to get personal data node by label", zap.Error(err))
			return nil, err
		}
		if pdn != nil {
			logger.Error(
				"personal data node with label already exists",
				zap.Error(err),
				zap.Int64("existing-personal-data-node-id", pdn.ID),
			)
			return nil, errorwrapper.New(fmt.Sprintf("personal data node already exists, label = %s", params.Label))
		}
	}

	pdn := &models.PersonalDataNode{
		ID:                        s.idgenerator.Generate().Int64(),
		CreatedAt:                 time.Now(),
		LastModifiedAt:            time.Now(),
		NetworkWardenID:           params.NetworkWardenID,
		HolderID:                  params.HolderID,
		Label:                     params.Label,
		Address:                   address,
		Name:                      params.Name,
		Description:               params.Description,
		Location:                  params.Location,
		AccountsCapacity:          0,
		Alive:                     false,
		LastPingedAt:              sql.NullTime{},
		IsOpen:                    false,
		URL:                       params.URL,
		APIKeyHash:                "",
		Version:                   "",
		RateLimitMaxRequests:      0,
		RateLimitInterval:         0,
		CrawlRateLimitMaxRequests: 0,
		CrawlRateLimitInterval:    0,
		Status:                    models.PersonalDataNodeStatusPending,
		IDGenNode:                 -1,
	}

	if err := s.repo.InsertPersonalDataNode(ctx, pdn); err != nil {
		logger.Error(
			"failed to insert personal data node",
			zap.Error(err),
			zap.Int64("personal-data-node-id", pdn.ID),
		)
		return nil, err
	}

	return pdn, nil
}
