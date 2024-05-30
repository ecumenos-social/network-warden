package auth

import (
	"context"
	"database/sql"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/jwt"
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
	Insert(ctx context.Context, params *InsertParams) (*models.HolderSession, error)
	GetHolderSessionByToken(ctx context.Context, token string, scope jwt.TokenScope) (*models.HolderSession, error)
	ModifyHolderSession(ctx context.Context, id int64, holderSession *models.HolderSession) error
	MakeHolderSessionExpired(ctx context.Context, id int64, holderSession *models.HolderSession) error
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

func (s *service) Insert(ctx context.Context, params *InsertParams) (*models.HolderSession, error) {
	hs := &models.HolderSession{
		ID:             s.idgenerator.Generate().Int64(),
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
		return nil, err
	}

	return hs, nil
}

func (s *service) GetHolderSessionByToken(ctx context.Context, token string, scope jwt.TokenScope) (*models.HolderSession, error) {
	var hs *models.HolderSession
	var err error
	switch scope {
	case jwt.TokenScopeAccess:
		hs, err = s.repo.GetHolderSessionByToken(ctx, token)
	case jwt.TokenScopeRefresh:
		hs, err = s.repo.GetHolderSessionByRefreshToken(ctx, token)
	}
	if err != nil {
		return nil, errorwrapper.WrapMessage(err, "failed to get session")
	}
	if hs == nil {
		return nil, errorwrapper.New("session is not found")
	}
	if hs.ExpiredAt.Valid && hs.ExpiredAt.Time.Before(time.Now()) {
		return nil, errorwrapper.New("token was expired")
	}

	return hs, nil
}

func (s *service) ModifyHolderSession(ctx context.Context, id int64, holderSession *models.HolderSession) error {
	holderSession.LastModifiedAt = time.Now()
	holderSession.ExpiredAt = sql.NullTime{
		Time:  time.Now().Add(s.sessionAge),
		Valid: true,
	}

	return s.repo.ModifyHolderSession(ctx, id, holderSession)
}

func (s *service) MakeHolderSessionExpired(ctx context.Context, id int64, holderSession *models.HolderSession) error {
	holderSession.LastModifiedAt = time.Now()
	holderSession.ExpiredAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	return s.repo.ModifyHolderSession(ctx, id, holderSession)
}
