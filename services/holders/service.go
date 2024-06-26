package holders

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkit/hash"
	"github.com/ecumenos-social/toolkit/random"
	"go.uber.org/zap"
)

type Repository interface {
	GetHoldersByEmails(ctx context.Context, emails []string) ([]*models.Holder, error)
	GetHoldersByPhoneNumbers(ctx context.Context, phoneNumbers []string) ([]*models.Holder, error)
	InsertHolder(ctx context.Context, holder *models.Holder) error
	GetHolderByEmail(ctx context.Context, email string) (*models.Holder, error)
	GetHolderByPhoneNumber(ctx context.Context, phoneNumber string) (*models.Holder, error)
	GetHolderByID(ctx context.Context, id int64) (*models.Holder, error)
	ModifyHolder(ctx context.Context, id int64, holder *models.Holder) error
	DeleteHolder(ctx context.Context, id int64) error
}

type Service interface {
	CheckEmailsUsage(ctx context.Context, logger *zap.Logger, emails []string) error
	CheckPhoneNumbersUsage(ctx context.Context, logger *zap.Logger, phoneNumbers []string) error
	Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.Holder, error)
	GetHolderByEmailOrPhoneNumber(ctx context.Context, logger *zap.Logger, email, phoneNumber string) (*models.Holder, error)
	ValidatePassword(ctx context.Context, logger *zap.Logger, holder *models.Holder, password string) error
	GetHolderByID(ctx context.Context, logger *zap.Logger, id int64) (*models.Holder, error)
	Confirm(ctx context.Context, logger *zap.Logger, id int64, confirmationCode string) (*models.Holder, error)
	RegenerateConfirmationCode(ctx context.Context, logger *zap.Logger, id int64) (*models.Holder, error)
	ChangePassword(ctx context.Context, logger *zap.Logger, holder *models.Holder, password string) error
	Modify(ctx context.Context, logger *zap.Logger, holder *models.Holder, params *ModifyParams) (*models.Holder, error)
	Delete(ctx context.Context, logger *zap.Logger, id int64) error
}

type service struct {
	repo        Repository
	idgenerator idgenerators.HoldersIDGenerator
}

func New(repo Repository, g idgenerators.HoldersIDGenerator) Service {
	return &service{
		repo:        repo,
		idgenerator: g,
	}
}

func (s *service) CheckEmailsUsage(ctx context.Context, logger *zap.Logger, emails []string) error {
	entities, err := s.repo.GetHoldersByEmails(ctx, emails)
	if err != nil {
		logger.Error("failed to get holders by emails", zap.Error(err))
		return err
	}
	if len(entities) > 0 {
		logger.Error("some email from emails list is in use", zap.Strings("emails", emails))
		return errorwrapper.New(fmt.Sprintf("some email from emails list ([%s]) is in use", strings.Join(emails, ", ")))
	}

	return nil
}

func (s *service) CheckPhoneNumbersUsage(ctx context.Context, logger *zap.Logger, phoneNumbers []string) error {
	entities, err := s.repo.GetHoldersByPhoneNumbers(ctx, phoneNumbers)
	if err != nil {
		logger.Error("failed to get holders by phone numbers", zap.Error(err))
		return err
	}
	if len(entities) > 0 {
		logger.Error("some phone number from phone numbers list is in use", zap.Strings("phone-numbers", phoneNumbers))
		return errorwrapper.New(fmt.Sprintf("some phone number from phone numbers list ([%s]) is in use", strings.Join(phoneNumbers, ", ")))
	}

	return nil
}

type InsertParams struct {
	Emails         []string
	PhoneNumbers   []string
	AvatarImageURL *string
	Countries      []string
	Languages      []string
	Password       string
}

func HashPassword(password string) (string, error) {
	return hash.Hash(password)
}

func (s *service) Insert(ctx context.Context, logger *zap.Logger, params *InsertParams) (*models.Holder, error) {
	passwordHash, err := HashPassword(params.Password)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	id := s.idgenerator.Generate().Int64()
	h := &models.Holder{
		ID:               id,
		CreatedAt:        time.Now(),
		LastModifiedAt:   time.Now(),
		Emails:           params.Emails,
		PhoneNumbers:     params.PhoneNumbers,
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
		logger.Error(
			"failed to insert holder",
			zap.Error(err),
			zap.Int64("holder-id", id),
			zap.Strings("emails", params.Emails),
			zap.Strings("phone_numbers", params.PhoneNumbers),
			zap.Strings("countries", params.Countries),
			zap.Strings("languages", params.Languages),
		)
		return nil, err
	}

	return h, nil
}

func generateConfirmationCode() string {
	return random.GenNumericString(10)
}

func (s *service) GetHolderByEmailOrPhoneNumber(ctx context.Context, logger *zap.Logger, email, phoneNumber string) (h *models.Holder, err error) {
	if email == "" && phoneNumber == "" {
		logger.Error("failed to hash password")
		return nil, errorwrapper.New("can not query holder if email address is empty and phone number is empty")
	}

	if email != "" {
		h, err = s.repo.GetHolderByEmail(ctx, email)
	} else {
		h, err = s.repo.GetHolderByPhoneNumber(ctx, phoneNumber)
	}
	if err != nil {
		logger.Error(
			"failed to get holder",
			zap.Error(err),
			zap.String("email", email),
			zap.String("phone-number", phoneNumber),
		)
		return nil, err
	}

	return h, nil
}

func (s *service) ValidatePassword(_ context.Context, logger *zap.Logger, holder *models.Holder, password string) error {
	if hash.CompareHash(password, holder.PasswordHash) {
		return nil
	}
	logger.Error("invalid password")

	return errorwrapper.New("invalid password")
}

func (s *service) GetHolderByID(ctx context.Context, logger *zap.Logger, id int64) (*models.Holder, error) {
	h, err := s.repo.GetHolderByID(ctx, id)
	if err != nil {
		logger.Error("failed get holder by id", zap.Error(err))
		return nil, err
	}

	return h, nil
}

func (s *service) Confirm(ctx context.Context, logger *zap.Logger, id int64, confirmationCode string) (*models.Holder, error) {
	holder, err := s.repo.GetHolderByID(ctx, id)
	if err != nil {
		logger.Error("failed to get holder by id", zap.Error(err))
		return nil, err
	}
	if holder == nil {
		logger.Error("holder is not found")
		return nil, errorwrapper.New("holder is not found")
	}
	if holder.ConfirmationCode != confirmationCode {
		logger.Error("invalid confirmation code")
		return nil, errorwrapper.New("invalid confirmation code")
	}
	holder.Confirmed = true
	if err := s.repo.ModifyHolder(ctx, id, holder); err != nil {
		logger.Error("failed to modify holder to make confirm=true", zap.Error(err))
		return nil, err
	}

	return holder, nil
}

func (s *service) RegenerateConfirmationCode(ctx context.Context, logger *zap.Logger, id int64) (*models.Holder, error) {
	holder, err := s.repo.GetHolderByID(ctx, id)
	if err != nil {
		logger.Error("failed to get holder by id", zap.Error(err))
		return nil, err
	}

	holder.ConfirmationCode = generateConfirmationCode()
	if err := s.repo.ModifyHolder(ctx, id, holder); err != nil {
		logger.Error("failed to modify holder to make confirmation_code={{new_confirmation_code}}", zap.Error(err))
		return nil, err
	}

	return holder, nil
}

func (s *service) ChangePassword(ctx context.Context, logger *zap.Logger, holder *models.Holder, password string) error {
	passwordHash, err := hash.Hash(password)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	holder.PasswordHash = passwordHash
	if err := s.repo.ModifyHolder(ctx, holder.ID, holder); err != nil {
		logger.Error("failed to modify holder to change password", zap.Error(err))
		return err
	}

	return nil
}

type ModifyParams struct {
	AvatarImageURL *string
	Countries      []string
	Languages      []string
}

func (s *service) Modify(ctx context.Context, logger *zap.Logger, holder *models.Holder, params *ModifyParams) (*models.Holder, error) {
	if params.AvatarImageURL != nil {
		holder.AvatarImageURL = sql.NullString{
			String: *params.AvatarImageURL,
			Valid:  true,
		}
	}
	if params.Countries != nil {
		holder.Countries = params.Countries
	}
	if params.Languages != nil {
		holder.Languages = params.Languages
	}
	if err := s.repo.ModifyHolder(ctx, holder.ID, holder); err != nil {
		logger.Error("failed to modify holder", zap.Error(err))
		return nil, err
	}

	return holder, nil
}

func (s *service) Delete(ctx context.Context, logger *zap.Logger, id int64) error {
	if err := s.repo.DeleteHolder(ctx, id); err != nil {
		logger.Error("failed to delete holder", zap.Error(err))
		return errorwrapper.WrapMessage(err, "failed to delete holder")
	}

	return nil
}
