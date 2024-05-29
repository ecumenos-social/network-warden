package holders

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/toolkit/hash"
	"github.com/ecumenos-social/toolkit/random"
)

type Repository interface {
	GetHoldersByEmails(ctx context.Context, emails []string) ([]*models.Holder, error)
	GetHoldersByPhoneNumbers(ctx context.Context, phoneNumbers []string) ([]*models.Holder, error)
	InsertHolder(ctx context.Context, holder *models.Holder) error
	GetHolderByEmail(ctx context.Context, email string) (*models.Holder, error)
	GetHolderByPhoneNumber(ctx context.Context, phoneNumber string) (*models.Holder, error)
}

type Service interface {
	CheckEmailsUsage(ctx context.Context, emails []string) error
	CheckPhoneNumbersUsage(ctx context.Context, phoneNumbers []string) error
	Insert(ctx context.Context, params *InsertParams) (*models.Holder, error)
	GetHolderByEmailOrPhoneNumber(ctx context.Context, email, phoneNumber string) (*models.Holder, error)
	ValidatePassword(ctx context.Context, holder *models.Holder, password string) error
}

type service struct {
	repo        Repository
	idgenerator idgenerator.Generator
}

func New(repo Repository, g idgenerator.Generator) Service {
	return &service{
		repo:        repo,
		idgenerator: g,
	}
}

func (s *service) CheckEmailsUsage(ctx context.Context, emails []string) error {
	entities, err := s.repo.GetHoldersByEmails(ctx, emails)
	if err != nil {
		return err
	}
	if len(entities) > 0 {
		return errorwrapper.New(fmt.Sprintf("some email from emails list ([%s]) is in use", strings.Join(emails, ", ")))
	}

	return nil
}

func (s *service) CheckPhoneNumbersUsage(ctx context.Context, phoneNumbers []string) error {
	entities, err := s.repo.GetHoldersByPhoneNumbers(ctx, phoneNumbers)
	if err != nil {
		return err
	}
	if len(entities) > 0 {
		return errorwrapper.New(fmt.Sprintf("some email from phone numbers list ([%s]) is in use", strings.Join(phoneNumbers, ", ")))
	}

	return nil
}

type InsertParams struct {
	Emails         []string
	PhoneNumber    []string
	AvatarImageURL *string
	Countries      []string
	Languages      []string
	Password       string
}

func (s *service) Insert(ctx context.Context, params *InsertParams) (*models.Holder, error) {
	passwordHash, err := hash.Hash(params.Password)
	if err != nil {
		return nil, err
	}

	h := &models.Holder{
		ID:               s.idgenerator.Generate().Int64(),
		CreatedAt:        time.Now(),
		LastModifiedAt:   time.Now(),
		Emails:           params.Emails,
		PhoneNumbers:     params.PhoneNumber,
		Countries:        params.Countries,
		Languages:        params.Languages,
		PasswordHash:     passwordHash,
		Confirmed:        false,
		ConfirmationCode: generateConfirmationCode(),
	}
	if params.AvatarImageURL != nil {
		h.AvatarImageURL = sql.NullString{
			String: *params.AvatarImageURL,
			Valid:  true,
		}
	}

	if err := s.repo.InsertHolder(ctx, h); err != nil {
		return nil, err
	}

	return h, nil
}

func generateConfirmationCode() string {
	return random.GenNumericString(10)
}

func (s *service) GetHolderByEmailOrPhoneNumber(ctx context.Context, email, phoneNumber string) (*models.Holder, error) {
	if email == "" && phoneNumber == "" {
		return nil, errorwrapper.New("can not query holder if email address is empty and phone number is empty")
	}

	if email != "" {
		return s.repo.GetHolderByEmail(ctx, email)
	}

	return s.repo.GetHolderByPhoneNumber(ctx, phoneNumber)
}

func (s *service) ValidatePassword(_ context.Context, holder *models.Holder, password string) error {
	if hash.CompareHash(password, holder.PasswordHash) {
		return nil
	}

	return errorwrapper.New("invalid password")
}
