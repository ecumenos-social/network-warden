package personaldatanodes

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
	InsertPersonalDataNode(ctx context.Context, pdn *models.PersonalDataNode) error
	GetPersonalDataNodeByLabel(ctx context.Context, label string) (*models.PersonalDataNode, error)
	GetPersonalDataNodeByID(ctx context.Context, id int64) (*models.PersonalDataNode, error)
	GetPersonalDataNodeByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.PersonalDataNode, error)
	GetPersonalDataNodesList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.PersonalDataNode, error)
	ModifyPersonalDataNode(ctx context.Context, id int64, pdn *models.PersonalDataNode) error
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.PersonalDataNode, error)
	Activate(ctx context.Context, logger *zap.Logger, holderID, id int64) (*models.PersonalDataNode, string, error)
	Initiate(ctx context.Context, logger *zap.Logger, apiKey string, params *InitiateParams) error
	GetList(ctx context.Context, logger *zap.Logger, holderID int64, pagination *types.Pagination, onlyMy bool) ([]*models.PersonalDataNode, error)
	GetByID(ctx context.Context, logger *zap.Logger, id int64) (*models.PersonalDataNode, error)
	SetStatusByID(ctx context.Context, logger *zap.Logger, id int64, status models.PersonalDataNodeStatus) error
}

type service struct {
	repo                       Repository
	idgenerator                idgenerators.PersonalDataNodesIDGenerator
	networkWardenID            int64
	networkWardenAddressSuffix string
}

func New(cfg *toolkitfx.GenericAppConfig, spcCfg *toolkitfx.NetworkWardenAppConfig, repo Repository, g idgenerators.PersonalDataNodesIDGenerator) Service {
	return &service{
		repo:                       repo,
		idgenerator:                g,
		networkWardenID:            cfg.ID,
		networkWardenAddressSuffix: spcCfg.AddressSuffix,
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
		IsInviteCodeRequired:      false,
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

func HashAPIKey(apiKey string) string {
	return hash.SHA1(apiKey)
}

func (s *service) Activate(ctx context.Context, logger *zap.Logger, holderID, id int64) (*models.PersonalDataNode, string, error) {
	pdn, err := s.repo.GetPersonalDataNodeByID(ctx, id)
	if err != nil {
		logger.Error("failed to get personal data node by id", zap.Error(err))
		return nil, "", err
	}
	if pdn == nil {
		logger.Error("network node is not found")
		return nil, "", errorwrapper.New("personal data node is not found")
	}
	if pdn.HolderID != holderID {
		logger.Error("have no permissions for confirm personal data node")
		return nil, "", errorwrapper.New("have no permissions for confirm personal data node")
	}
	if pdn.Status != models.PersonalDataNodeStatusApproved {
		logger.Error("personal data node is not approved")
		return nil, "", errorwrapper.New("personal data node is not approved")
	}

	apiKey, err := random.GenAPIKey("nn", fmt.Sprint(s.networkWardenID))
	if err != nil {
		logger.Error("failed to generate API key", zap.Error(err))
		return nil, "", err
	}
	pdn.APIKeyHash = HashAPIKey(apiKey)

	if err := s.repo.ModifyPersonalDataNode(ctx, id, pdn); err != nil {
		logger.Error("failed to modify personal data node", zap.Error(err))
		return nil, "", errorwrapper.New("failed to modify personal data node")
	}

	return pdn, apiKey, nil
}

type InitiateParams struct {
	AccountsCapacity     int64
	IsOpen               bool
	IsInviteCodeRequired bool
	Version              string
	RateLimit            *types.RateLimit
	CrawlRateLimit       *types.RateLimit
	IDGenNode            int64
}

func (s *service) Initiate(ctx context.Context, logger *zap.Logger, apiKey string, params *InitiateParams) error {
	apiKeyHash := HashAPIKey(apiKey)
	pdn, err := s.repo.GetPersonalDataNodeByAPIKeyHash(ctx, apiKeyHash)
	if err != nil {
		logger.Error("failed to get personal data node by api key", zap.Error(err), zap.String("api-key", apiKey))
		return err
	}
	if pdn == nil {
		logger.Error("personal data node is not found")
		return errorwrapper.New("personal data node is not found")
	}
	if pdn.Status != models.PersonalDataNodeStatusApproved {
		logger.Error("personal data node is not approved", zap.Int64("personal-data-node-id", pdn.ID), zap.String("personal-data-node-status", string(pdn.Status)))
		return errorwrapper.New("personal data node must be approved")
	}

	pdn.AccountsCapacity = params.AccountsCapacity
	pdn.IsOpen = params.IsOpen
	pdn.IsInviteCodeRequired = params.IsInviteCodeRequired
	pdn.Version = params.Version
	pdn.RateLimitInterval = params.RateLimit.Interval
	pdn.RateLimitMaxRequests = params.RateLimit.MaxRequests
	pdn.CrawlRateLimitInterval = params.CrawlRateLimit.Interval
	pdn.CrawlRateLimitMaxRequests = params.CrawlRateLimit.MaxRequests
	pdn.IDGenNode = params.IDGenNode
	if err := s.repo.ModifyPersonalDataNode(ctx, pdn.ID, pdn); err != nil {
		logger.Error("failed to modify personal data node", zap.Error(err), zap.Int64("personal-data-node-id", pdn.ID))
		return errorwrapper.New("failed to modify personal data node")
	}

	return nil
}

func (s *service) GetList(ctx context.Context, logger *zap.Logger, holderID int64, pagination *types.Pagination, onlyMy bool) ([]*models.PersonalDataNode, error) {
	filters := make(map[string]interface{}, 1)
	if onlyMy {
		filters["holder_id"] = holderID
	}

	return s.repo.GetPersonalDataNodesList(ctx, filters, pagination)
}

func (s *service) GetByID(ctx context.Context, logger *zap.Logger, id int64) (*models.PersonalDataNode, error) {
	return s.repo.GetPersonalDataNodeByID(ctx, id)
}

func (s *service) SetStatusByID(ctx context.Context, logger *zap.Logger, id int64, status models.PersonalDataNodeStatus) error {
	pdn, err := s.repo.GetPersonalDataNodeByID(ctx, id)
	if err != nil {
		logger.Error("failed to get personal data node by id", zap.Error(err), zap.Int64("personal-data-node-id", id))
		return err
	}
	pdn.Status = status
	if err := s.repo.ModifyPersonalDataNode(ctx, pdn.ID, pdn); err != nil {
		logger.Error("failed to modify personal data node", zap.Error(err), zap.Int64("personal-data-node-id", pdn.ID))
		return errorwrapper.New("failed to modify personal data node")
	}

	return nil
}
