package pgseeds

import (
	"context"
	"database/sql"

	"github.com/ecumenos-social/network-warden/cmd/admin/pgseeds/data"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/admins"
	"go.uber.org/zap"
)

type Runner interface {
	Run() error
}

type runner struct {
	logger     *zap.Logger
	adminsRepo admins.Repository
}

func New(logger *zap.Logger, adminsRepo admins.Repository) Runner {
	return &runner{
		logger:     logger,
		adminsRepo: adminsRepo,
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
	as, err := data.LoadAdmins(ctx)
	if err != nil {
		return err
	}
	logger.Info("loaded admins", zap.Int("admins-length", len(as)))
	for _, a := range as {
		avatarImageURL := sql.NullString{}
		if a.AvatarImageURL != nil {
			avatarImageURL.Valid = true
			avatarImageURL.String = *a.AvatarImageURL
		}
		passwordHash, err := admins.HashPassword(a.Password)
		if err != nil {
			return err
		}

		if err := r.adminsRepo.InsertAdmin(ctx, &models.Admin{
			ID:             a.ID,
			CreatedAt:      a.CreatedAt,
			LastModifiedAt: a.LastModifiedAt,
			Emails:         a.Emails,
			PhoneNumbers:   a.PhoneNumbers,
			AvatarImageURL: avatarImageURL,
			Countries:      a.Countries,
			Languages:      a.Languages,
			PasswordHash:   passwordHash,
		}); err != nil {
			return err
		}
	}

	return nil
}
