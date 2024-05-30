package auth

import (
	"context"
	"database/sql"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/jwt"
	"go.uber.org/zap"
)

type Config struct {
	SessionAge time.Duration
}

type Repository interface {
	InsertHolderSession(ctx context.Context, holderSession *models.HolderSession) error
	GetHolderSessionByRefreshToken(ctx context.Context, refToken string) (*models.HolderSession, error)
	GetHolderSessionByToken(ctx context.Context, token string) (*models.HolderSession, error)
	ModifyHolderSession(ctx context.Context, id int64, holderSession *models.HolderSession) error
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.HolderSession, error)
	GetHolderSessionByToken(ctx context.Context, logger *zap.Logger, token string, scope jwt.TokenScope) (*models.HolderSession, error)
	GetExpiredAtForHolderSession() time.Time
	ModifyHolderSession(ctx context.Context, logger *zap.Logger, id int64, holderSession *models.HolderSession) error
	MakeHolderSessionExpired(ctx context.Context, logger *zap.Logger, id int64, holderSession *models.HolderSession) error
}

type service struct {
	sessionAge  time.Duration
	repo        Repository
	idgenerator idgenerator.Generator
}

func New(config *Config, repo Repository, g idgenerator.Generator) Service {
	return &service{
		sessionAge:  config.SessionAge,
		repo:        repo,
		idgenerator: g,
	}
}

type InsertParams struct {
	HolderID         int64
	Token            string
	RefreshToken     string
	RemoteIPAddress  string
	RemoteMACAddress string
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.HolderSession, error) {
	id := s.idgenerator.Generate().Int64()
	logger = logger.With(
		zap.Int64("holder-session-id", id),
		zap.Int64("holder-id", params.HolderID),
	)
	hs := &models.HolderSession{
		ID:             id,
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		HolderID:       params.HolderID,
		Token:          params.Token,
		RefreshToken:   params.RefreshToken,
		ExpiredAt: sql.NullTime{
			Time:  time.Now().Add(s.sessionAge),
			Valid: true,
		},
		RemoteIPAddress: sql.NullString{
			String: params.RemoteIPAddress,
			Valid:  true,
		},
		RemoteMACAddress: sql.NullString{
			String: params.RemoteMACAddress,
			Valid:  true,
		},
	}

	if err := s.repo.InsertHolderSession(ctx, hs); err != nil {
		logger.Info("failed to insert holder session", zap.Error(err))
		return nil, err
	}

	return hs, nil
}

func (s *service) GetHolderSessionByToken(ctx context.Context, logger *zap.Logger, token string, scope jwt.TokenScope) (*models.HolderSession, error) {
	var hs *models.HolderSession
	var err error
	switch scope {
	case jwt.TokenScopeAccess:
		hs, err = s.repo.GetHolderSessionByToken(ctx, token)
	case jwt.TokenScopeRefresh:
		hs, err = s.repo.GetHolderSessionByRefreshToken(ctx, token)
	}
	if err != nil {
		logger.Info("failed to get holder session by token", zap.String("token-scope", scope.String()))
		return nil, errorwrapper.WrapMessage(err, "failed to get session")
	}
	if hs == nil {
		logger.Info("session is not found")
		return nil, errorwrapper.New("session is not found")
	}
	if hs.ExpiredAt.Valid && hs.ExpiredAt.Time.Before(time.Now()) {
		logger.Info("token was expired", zap.Time("expired-at", hs.ExpiredAt.Time))
		return nil, errorwrapper.New("token was expired")
	}

	return hs, nil
}

func (s *service) GetExpiredAtForHolderSession() time.Time {
	return time.Now().Add(s.sessionAge)
}

func (s *service) ModifyHolderSession(ctx context.Context, logger *zap.Logger, id int64, holderSession *models.HolderSession) error {
	logger = logger.With(
		zap.Int64("holder-session-id", id),
		zap.Int64("holder-id", holderSession.HolderID),
	)
	holderSession.LastModifiedAt = time.Now()

	if err := s.repo.ModifyHolderSession(ctx, id, holderSession); err != nil {
		logger.Error("failed to modify holder session", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) MakeHolderSessionExpired(ctx context.Context, logger *zap.Logger, id int64, holderSession *models.HolderSession) error {
	holderSession.LastModifiedAt = time.Now()
	holderSession.ExpiredAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return s.ModifyHolderSession(ctx, logger, id, holderSession)
}
