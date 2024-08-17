package admins

import (
	"context"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkit/hash"
	"go.uber.org/zap"
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
	GetHolderByEmailOrPhoneNumber(ctx context.Context, logger *zap.Logger, email, phoneNumber string) (a *models.Admin, err error)
	ValidatePassword(_ context.Context, logger *zap.Logger, a *models.Admin, password string) error
}

type service struct {
	repo        Repository
	idgenerator idgenerators.AdminsIDGenerator
}

func New(repo Repository, idgenerator idgenerators.AdminsIDGenerator) Service {
	return &service{
		repo:        repo,
		idgenerator: idgenerator,
	}
}

func (s *service) GetHolderByEmailOrPhoneNumber(ctx context.Context, logger *zap.Logger, email, phoneNumber string) (a *models.Admin, err error) {
	if email == "" && phoneNumber == "" {
		logger.Error("failed to hash password")
		return nil, errorwrapper.New("can not query admin if email address is empty and phone number is empty")
	}

	if email != "" {
		a, err = s.repo.GetAdminByEmail(ctx, email)
	} else {
		a, err = s.repo.GetAdminByPhoneNumber(ctx, phoneNumber)
	}
	if err != nil {
		logger.Error(
			"failed to get admin",
			zap.Error(err),
			zap.String("email", email),
			zap.String("phone-number", phoneNumber),
		)
		return nil, err
	}

	return a, nil
}

func HashPassword(password string) (string, error) {
	return hash.Hash(password)
}

func (s *service) ValidatePassword(_ context.Context, logger *zap.Logger, a *models.Admin, password string) error {
	if hash.CompareHash(password, a.PasswordHash) {
		return nil
	}
	logger.Error("invalid password")

	return errorwrapper.New("invalid password")
}
