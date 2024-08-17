package adminauth

import (
	"context"
	"database/sql"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/network-warden/services/jwt"
	"go.uber.org/zap"
)

type Config struct {
	SessionAge time.Duration
}

type Repository interface {
	InsertAdminSession(ctx context.Context, adminSession *models.AdminSession) error
	GetAdminSessionByRefreshToken(ctx context.Context, refToken string) (*models.AdminSession, error)
	GetAdminSessionByToken(ctx context.Context, token string) (*models.AdminSession, error)
	ModifyAdminSession(ctx context.Context, id int64, adminSession *models.AdminSession) error
}

type Service interface {
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.AdminSession, error)
	GetAdminSessionByToken(ctx context.Context, logger *zap.Logger, token string, scope jwt.TokenScope) (*models.AdminSession, error)
	GetExpiredAtForAdminSession() time.Time
	ModifyAdminSession(ctx context.Context, logger *zap.Logger, id int64, adminSession *models.AdminSession) error
	MakeAdminSessionExpired(ctx context.Context, logger *zap.Logger, id int64, adminSession *models.AdminSession) error
}

type service struct {
	sessionAge  time.Duration
	repo        Repository
	idgenerator idgenerators.AdminSessionsIDGenerator
}

func New(config *Config, repo Repository, g idgenerators.AdminSessionsIDGenerator) Service {
	return &service{
		sessionAge:  config.SessionAge,
		repo:        repo,
		idgenerator: g,
	}
}

type InsertParams struct {
	AdminID          int64
	Token            string
	RefreshToken     string
	RemoteIPAddress  string
	RemoteMACAddress *string
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.AdminSession, error) {
	id := s.idgenerator.Generate().Int64()
	logger = logger.With(
		zap.Int64("admin-session-id", id),
		zap.Int64("admin-id", params.AdminID),
	)
	remoteMACAddress := sql.NullString{}
	if params.RemoteMACAddress != nil {
		remoteMACAddress.String = *params.RemoteMACAddress
		remoteMACAddress.Valid = true
	}
	hs := &models.AdminSession{
		ID:             id,
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		AdminID:        params.AdminID,
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
		RemoteMACAddress: remoteMACAddress,
	}

	if err := s.repo.InsertAdminSession(ctx, hs); err != nil {
		logger.Info("failed to insert admin session", zap.Error(err))
		return nil, err
	}

	return hs, nil
}

func (s *service) GetAdminSessionByToken(ctx context.Context, logger *zap.Logger, token string, scope jwt.TokenScope) (*models.AdminSession, error) {
	var as *models.AdminSession
	var err error
	switch scope {
	case jwt.TokenScopeAccess:
		as, err = s.repo.GetAdminSessionByToken(ctx, token)
	case jwt.TokenScopeRefresh:
		as, err = s.repo.GetAdminSessionByRefreshToken(ctx, token)
	}
	if err != nil {
		logger.Info("failed to get admin session by token", zap.String("token-scope", scope.String()))
		return nil, errorwrapper.WrapMessage(err, "failed to get session")
	}
	if as == nil {
		logger.Info("session is not found")
		return nil, errorwrapper.New("session is not found")
	}
	if as.ExpiredAt.Valid && as.ExpiredAt.Time.Before(time.Now()) {
		logger.Info("token was expired", zap.Time("expired-at", as.ExpiredAt.Time))
		return nil, errorwrapper.New("token was expired")
	}

	return as, nil
}

func (s *service) GetExpiredAtForAdminSession() time.Time {
	return time.Now().Add(s.sessionAge)
}

func (s *service) ModifyAdminSession(ctx context.Context, logger *zap.Logger, id int64, adminSession *models.AdminSession) error {
	logger = logger.With(
		zap.Int64("admin-session-id", id),
		zap.Int64("admin-id", adminSession.AdminID),
	)
	adminSession.LastModifiedAt = time.Now()

	if err := s.repo.ModifyAdminSession(ctx, id, adminSession); err != nil {
		logger.Error("failed to modify admin session", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) MakeAdminSessionExpired(ctx context.Context, logger *zap.Logger, id int64, adminSession *models.AdminSession) error {
	adminSession.LastModifiedAt = time.Now()
	adminSession.ExpiredAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return s.ModifyAdminSession(ctx, logger, id, adminSession)
}
