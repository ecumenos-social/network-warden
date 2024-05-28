package holdersessions

import (
	"context"
	"database/sql"
	"time"

	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/network-warden/models"
)

type Config struct {
	Age time.Duration
}

type Repository interface {
	InsertHolderSession(ctx context.Context, holderSession *models.HolderSession) error
}

type Service interface {
	Insert(ctx context.Context, params *InsertParams) (*models.HolderSession, error)
}

type service struct {
	sessionAge  time.Duration
	repo        Repository
	idgenerator idgenerator.Generator
}

func New(config *Config, repo Repository, g idgenerator.Generator) Service {
	return &service{
		sessionAge:  config.Age,
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
