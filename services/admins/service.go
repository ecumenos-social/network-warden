package admins

import (
	"context"

	"github.com/ecumenos-social/network-warden/models"
)

type Repository interface {
	GetAdminsByEmails(ctx context.Context, emails []string) ([]*models.Admin, error)
	GetAdminsByPhoneNumbers(ctx context.Context, phoneNumbers []string) ([]*models.Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error)
	GetAdminByPhoneNumber(ctx context.Context, phoneNumber string) (*models.Admin, error)
	GetAdminByID(ctx context.Context, id int64) (*models.Admin, error)
	InsertAdmin(ctx context.Context, admin *models.Admin) error
}

type Service interface {
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{
		repo: repo,
	}
}
