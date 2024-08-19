package networkwardens

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkit/types"
	"github.com/ecumenos-social/toolkitfx"
	"go.uber.org/zap"
)

type Repository interface {
	InsertNetworkWarden(ctx context.Context, nw *models.NetworkWarden) error
	GetNetworkWardensList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.NetworkWarden, error)
	GetNetworkWardenByLabel(ctx context.Context, label string) (*models.NetworkWarden, error)
	GetNetworkWardenByID(ctx context.Context, id int64) (*models.NetworkWarden, error)
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.NetworkWarden, error)
	GetList(ctx context.Context, logger *zap.Logger, pagination *types.Pagination) ([]*models.NetworkWarden, error)
}

type service struct {
	repo            Repository
	idgenerator     idgenerators.NetworkWardensIDGenerator
	networkWardenID int64
}

func New(config *toolkitfx.GenericAppConfig, repo Repository, g idgenerators.NetworkWardensIDGenerator) Service {
	return &service{
		repo:            repo,
		idgenerator:     g,
		networkWardenID: config.ID,
	}
}

func (s *service) buildAddress(label string) string {
	return fmt.Sprintf("::%s", label)
}

type InsertParams struct {
	Name        string
	Description string
	Label       string
	Location    *models.Location
	IsOpen      bool
	Version     string
	URL         string
	PDNCapacity int64
	NNCapacity  int64
	RateLimit   *types.RateLimit
	IDGenNode   int64
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.NetworkWarden, error) {
	logger = logger.With(zap.String("network-warden-label", params.Label), zap.String("network-warden-name", params.Name))
	if nw, err := s.repo.GetNetworkWardenByLabel(ctx, params.Label); err != nil || nw != nil {
		if err != nil {
			logger.Error("failed to get network warden by label", zap.Error(err))
			return nil, err
		}
		if nw != nil {
			logger.Error(
				"network warden with label already exists",
				zap.Error(err),
				zap.Int64("existing-network-warden-id", nw.ID),
			)
			return nil, errorwrapper.New(fmt.Sprintf("network warden already exists, label = %s", params.Label))
		}
	}

	nw := &models.NetworkWarden{
		ID:                   s.idgenerator.Generate().Int64(),
		CreatedAt:            time.Now(),
		LastModifiedAt:       time.Now(),
		Label:                params.Label,
		Address:              s.buildAddress(params.Label),
		Name:                 params.Name,
		Description:          params.Description,
		Location:             params.Location,
		PDNCapacity:          params.PDNCapacity,
		NNCapacity:           params.NNCapacity,
		Alive:                false,
		LastPingedAt:         sql.NullTime{},
		IsOpen:               params.IsOpen,
		URL:                  params.URL,
		Version:              params.Version,
		RateLimitMaxRequests: params.RateLimit.MaxRequests,
		RateLimitInterval:    params.RateLimit.Interval,
		IDGenNode:            params.IDGenNode,
	}

	if err := s.repo.InsertNetworkWarden(ctx, nw); err != nil {
		logger.Error(
			"failed to insert network warden",
			zap.Error(err),
			zap.Int64("network-warden-id", nw.ID),
		)
		return nil, err
	}

	return nw, nil
}

func (s *service) GetList(ctx context.Context, logger *zap.Logger, pagination *types.Pagination) ([]*models.NetworkWarden, error) {
	return s.repo.GetNetworkWardensList(ctx, map[string]interface{}{}, pagination)
}
