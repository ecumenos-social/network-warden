package emailer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/toolkit/slices"
	"go.uber.org/zap"
)

type Config struct {
	SMTPHost           string
	SMTPPort           string
	SenderUsername     string
	SenderEmailAddress string
	SenderPassword     string

	ConfirmationOfRegistration *RateLimit
}

type Repository interface {
	InsertSentEmail(ctx context.Context, se *models.SentEmail) error
	ModifySentEmail(ctx context.Context, id int64, se *models.SentEmail) error
	GetSentEmails(ctx context.Context, sender, receiver, templateName string) ([]*models.SentEmail, error)
}

type Service interface {
	SendConfirmationOfRegistration(ctx context.Context, logger *zap.Logger, email, fullName, code string) error
	CanSendConfirmationOfRegistration(ctx context.Context, logger *zap.Logger, email string) (bool, error)
}

type service struct {
	smtpAddr           string
	smtpAuth           smtp.Auth
	senderEmailAddress string
	rateLimits         map[TemplateName]*RateLimit

	repo        Repository
	idgenerator idgenerators.SentEmailsIDGenerator
}

func New(config *Config, repo Repository, g idgenerators.SentEmailsIDGenerator) Service {
	return &service{
		smtpAddr: net.JoinHostPort(config.SMTPHost, config.SMTPPort),
		smtpAuth: smtp.PlainAuth(
			"",
			config.SenderUsername,
			config.SenderPassword,
			config.SMTPHost,
		),
		senderEmailAddress: config.SenderEmailAddress,
		rateLimits: map[TemplateName]*RateLimit{
			TemplateNameConfirmHolderRegistration: config.ConfirmationOfRegistration,
		},

		repo:        repo,
		idgenerator: g,
	}
}

func (s *service) SendConfirmationOfRegistration(ctx context.Context, logger *zap.Logger, email, fullName, code string) error {
	return s.sendTemplate(
		ctx,
		logger,
		TemplateNameConfirmHolderRegistration,
		[]string{email},
		[]string{},
		[]string{},
		struct{ FullName, ConfirmationCode, CurrentYear string }{
			FullName:         fullName,
			ConfirmationCode: code,
			CurrentYear:      fmt.Sprint(time.Now().Year()),
		},
		s.rateLimits[TemplateNameConfirmHolderRegistration],
	)
}

func (s *service) sendTemplate(ctx context.Context, logger *zap.Logger, name TemplateName, to, cc, files []string, data interface{}, rl *RateLimit) error {
	logger = logger.With(
		zap.Strings("to", to),
		zap.Strings("cc", cc),
		zap.String("template-name", string(name)),
	)
	canSend, err := s.canSendTemplate(ctx, logger, name, to, cc, rl)
	if err != nil {
		return err
	}
	if !canSend {
		logger.Error("too many send email requests", zap.Int64("max-requests", rl.MaxRequests), zap.Duration("interval", rl.Interval))
		return errorwrapper.New("too many send email requests")
	}

	message, err := s.formMessage(name, s.senderEmailAddress, to, cc, data)
	if err != nil {
		logger.Info("failed to compose email message", zap.Error(err))
		return err
	}
	if err := smtp.SendMail(s.smtpAddr, s.smtpAuth, s.senderEmailAddress, to, message); err != nil {
		logger.Info("failed to send email", zap.Error(err))
		return err
	}
	logger.Info("email was sent successfully")

	for _, receiverEmail := range slices.Merge(to, cc) {
		m := &models.SentEmail{
			ID:             s.idgenerator.Generate().Int64(),
			CreatedAt:      time.Now(),
			LastModifiedAt: time.Now(),
			SenderEmail:    s.senderEmailAddress,
			ReceiverEmail:  receiverEmail,
			TemplateName:   name.String(),
		}
		if err := s.repo.InsertSentEmail(ctx, m); err != nil {
			logger.Info("failed to insert sent email entity database", zap.Error(err), zap.String("receiver-email", receiverEmail))
			return err
		}
	}

	return nil
}

func (s *service) CanSendConfirmationOfRegistration(ctx context.Context, logger *zap.Logger, email string) (bool, error) {
	return s.canSendTemplate(ctx, logger, TemplateNameConfirmHolderRegistration, []string{email}, []string{}, s.rateLimits[TemplateNameConfirmHolderRegistration])
}

func (s *service) canSendTemplate(ctx context.Context, logger *zap.Logger, name TemplateName, to, cc []string, rl *RateLimit) (bool, error) {
	receivers := slices.Merge(to, cc)
	ses := make([]*models.SentEmail, 0, len(receivers))
	for _, r := range receivers {
		sentEmails, err := s.repo.GetSentEmails(ctx, s.senderEmailAddress, r, name.String())
		if err != nil {
			logger.Error("can not get sent emails", zap.Error(err))
			return false, errorwrapper.WrapMessage(err, "can not get sent emails")
		}
		ses = append(ses, sentEmails...)
	}

	return !rl.Exceed(ses), nil
}

func (s *service) formMessage(name TemplateName, from string, to, cc []string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(slices.Merge(to, cc), ",")))
	if len(cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ",")))
	}
	subject, err := takeSubject(TemplateNameConfirmHolderRegistration)
	if err != nil {
		return nil, err
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\";\r\n")

	t, err := takeTemplate(name)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
