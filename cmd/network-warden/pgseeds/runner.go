package pgseeds

import (
	"context"
	"database/sql"

	"github.com/ecumenos-social/network-warden/cmd/network-warden/pgseeds/data"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/holders"
	"go.uber.org/zap"
)

type Runner interface {
	Run() error
}

type runner struct {
	logger      *zap.Logger
	holdersRepo holders.Repository
}

func New(logger *zap.Logger, holdersRepo holders.Repository) Runner {
	return &runner{
		logger:      logger,
		holdersRepo: holdersRepo,
	}
}

func (r *runner) Run() error {
	ctx := context.Background()

	return r.seedHolders(ctx)
}

func (r *runner) seedHolders(ctx context.Context) error {
	logger := r.logger.With(zap.String("seed-round", "holders"))
	logger.Info("started seeding")
	defer logger.Info("finished seeding")
	hs, err := data.LoadHolders(ctx)
	if err != nil {
		return err
	}
	logger.Info("loaded holders", zap.Int("holders-length", len(hs)))
	for _, h := range hs {
		avatarImageURL := sql.NullString{}
		if h.AvatarImageURL != nil {
			avatarImageURL.Valid = true
			avatarImageURL.String = *h.AvatarImageURL
		}
		passwordHash, err := holders.HashPassword(h.Password)
		if err != nil {
			return err
		}

		if err := r.holdersRepo.InsertHolder(ctx, &models.Holder{
			ID:               h.ID,
			CreatedAt:        h.CreatedAt,
			LastModifiedAt:   h.LastModifiedAt,
			Emails:           h.Emails,
			PhoneNumbers:     h.PhoneNumbers,
			AvatarImageURL:   avatarImageURL,
			Countries:        h.Countries,
			Languages:        h.Languages,
			PasswordHash:     passwordHash,
			Confirmed:        h.Confirmed,
			ConfirmationCode: h.ConfirmationCode,
		}); err != nil {
			return err
		}
	}

	return nil
}
