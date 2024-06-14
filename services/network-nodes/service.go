package networknodes

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkit/hash"
	"github.com/ecumenos-social/toolkit/random"
	"github.com/ecumenos-social/toolkit/types"
	"github.com/ecumenos-social/toolkitfx"
	"go.uber.org/zap"
)

type Repository interface {
	InsertNetworkNode(ctx context.Context, nn *models.NetworkNode) error
	GetNetworkNodeByDomainName(ctx context.Context, domainName string) (*models.NetworkNode, error)
	GetNetworkNodeByID(ctx context.Context, id int64) (*models.NetworkNode, error)
	GetNetworkNodeByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.NetworkNode, error)
	ModifyNetworkNode(ctx context.Context, id int64, nn *models.NetworkNode) error
	GetNetworkNodesList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.NetworkNode, error)
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.NetworkNode, error)
	Activate(ctx context.Context, logger *zap.Logger, holderID, id int64) (nn *models.NetworkNode, apiKey string, err error)
	GetList(ctx context.Context, logger *zap.Logger, holderID int64, pagination *types.Pagination, onlyMy bool) ([]*models.NetworkNode, error)
	Initiate(ctx context.Context, logger *zap.Logger, apiKey string, params *InitiateParams) error
}

type service struct {
	repo            Repository
	idgenerator     idgenerators.NetworkNodesIDGenerator
	networkWardenID int64
}

func New(config *toolkitfx.NetworkWardenAppConfig, repo Repository, g idgenerators.NetworkNodesIDGenerator) Service {
	return &service{
		repo:            repo,
		idgenerator:     g,
		networkWardenID: config.ID,
	}
}

type InsertParams struct {
	HolderID        int64
	NetworkWardenID int64
	Name            string
	Description     string
	DomainName      string
	Location        *models.Location
	URL             string
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.NetworkNode, error) {
	logger = logger.With(zap.String("network-node-domain-name", params.DomainName), zap.String("network-node-name", params.Name))
	if nn, err := s.repo.GetNetworkNodeByDomainName(ctx, params.DomainName); err != nil || nn != nil {
		if err != nil {
			logger.Error("failed to get network node by domain name", zap.Error(err))
			return nil, err
		}
		if nn != nil {
			logger.Error(
				"network node with domain name already exists",
				zap.Error(err),
				zap.Int64("existing-network-node-id", nn.ID),
			)
			return nil, errorwrapper.New(fmt.Sprintf("network node already exists, domain name = %s", params.DomainName))
		}
	}

	nn := &models.NetworkNode{
		ID:                        s.idgenerator.Generate().Int64(),
		CreatedAt:                 time.Now(),
		LastModifiedAt:            time.Now(),
		NetworkWardenID:           params.NetworkWardenID,
		HolderID:                  params.HolderID,
		Name:                      params.Name,
		Description:               params.Description,
		DomainName:                params.DomainName,
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
		Status:                    models.NetworkNodeStatusPending,
		IDGenNode:                 -1,
	}

	if err := s.repo.InsertNetworkNode(ctx, nn); err != nil {
		logger.Error(
			"failed to insert network node",
			zap.Error(err),
			zap.Int64("network-node-id", nn.ID),
		)
		return nil, err
	}

	return nn, nil
}

func HashAPIKey(apiKey string) string {
	return hash.SHA1(apiKey)
}

func (s *service) Activate(ctx context.Context, logger *zap.Logger, holderID, id int64) (*models.NetworkNode, string, error) {
	nn, err := s.repo.GetNetworkNodeByID(ctx, id)
	if err != nil {
		logger.Error("failed to get network node by id", zap.Error(err))
		return nil, "", err
	}
	if nn == nil {
		logger.Error("network node is not found")
		return nil, "", errorwrapper.New("network node is not found")
	}
	if nn.HolderID != holderID {
		logger.Error("have no permissions for confirm network node")
		return nil, "", errorwrapper.New("have no permissions for confirm network node")
	}
	if nn.Status != models.NetworkNodeStatusApproved {
		logger.Error("network node is not approved")
		return nil, "", errorwrapper.New("network node is not approved")
	}

	apiKey, err := random.GenAPIKey("nn", fmt.Sprint(s.networkWardenID))
	if err != nil {
		logger.Error("failed to generate API key", zap.Error(err))
		return nil, "", err
	}
	nn.APIKeyHash = HashAPIKey(apiKey)

	if err := s.repo.ModifyNetworkNode(ctx, id, nn); err != nil {
		logger.Error("failed to modify network node", zap.Error(err))
		return nil, "", errorwrapper.New("failed to modify network node")
	}

	return nn, apiKey, nil
}

type InitiateParams struct {
	AccountsCapacity int64
	IsOpen           bool
	Version          string
	RateLimit        *types.RateLimit
	CrawlRateLimit   *types.RateLimit
	IDGenNode        int64
}

func (s *service) Initiate(ctx context.Context, logger *zap.Logger, apiKey string, params *InitiateParams) error {
	apiKeyHash := HashAPIKey(apiKey)
	nn, err := s.repo.GetNetworkNodeByAPIKeyHash(ctx, apiKeyHash)
	if err != nil {
		logger.Error("failed to get network node by api key", zap.Error(err), zap.String("api-key", apiKey))
		return err
	}
	if nn == nil {
		logger.Error("network node is not found")
		return errorwrapper.New("network node is not found")
	}
	if nn.Status != models.NetworkNodeStatusApproved {
		logger.Error("network node is not approved", zap.Int64("network-node-id", nn.ID), zap.String("network-node-status", string(nn.Status)))
		return errorwrapper.New("network node must be approved")
	}

	nn.AccountsCapacity = params.AccountsCapacity
	nn.IsOpen = params.IsOpen
	nn.Version = params.Version
	nn.RateLimitInterval = params.RateLimit.Interval
	nn.RateLimitMaxRequests = params.RateLimit.MaxRequests
	nn.CrawlRateLimitInterval = params.CrawlRateLimit.Interval
	nn.CrawlRateLimitMaxRequests = params.CrawlRateLimit.MaxRequests
	nn.IDGenNode = params.IDGenNode
	if err := s.repo.ModifyNetworkNode(ctx, nn.ID, nn); err != nil {
		logger.Error("failed to modify network node", zap.Error(err), zap.Int64("network-node-id", nn.ID))
		return errorwrapper.New("failed to modify network node")
	}

	return nil
}

func (s *service) GetList(ctx context.Context, logger *zap.Logger, holderID int64, pagination *types.Pagination, onlyMy bool) ([]*models.NetworkNode, error) {
	filters := make(map[string]interface{}, 1)
	if onlyMy {
		filters["holder_id"] = holderID
	}

	return s.repo.GetNetworkNodesList(ctx, filters, pagination)
}
