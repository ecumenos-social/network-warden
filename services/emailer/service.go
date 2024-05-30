package emailer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/pkg/sliceutils"
	"go.uber.org/zap"
)

type Config struct {
	SMTPHost           string
	SMTPPort           string
	SenderUsername     string
	SenderEmailAddress string
	SenderPassword     string
}

type Repository interface {
	InsertSentEmail(ctx context.Context, sentEmail *models.SentEmail) error
}

type Service interface {
	SendConfirmationOfRegistration(ctx context.Context, logger *zap.Logger, email, fullName, code string) error
}

type service struct {
	smtpAddr           string
	smtpAuth           smtp.Auth
	senderEmailAddress string

	repo        Repository
	idgenerator idgenerator.Generator
}

func New(config *Config, repo Repository, g idgenerator.Generator) Service {
	return &service{
		smtpAddr: net.JoinHostPort(config.SMTPHost, config.SMTPPort),
		smtpAuth: smtp.PlainAuth(
			"",
			config.SenderUsername,
			config.SenderPassword,
			config.SMTPHost,
		),
		senderEmailAddress: config.SenderEmailAddress,

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
	)
}

func (s *service) sendTemplate(ctx context.Context, logger *zap.Logger, name TemplateName, to, cc, files []string, data interface{}) error {
	logger = logger.With(
		zap.Strings("to", to),
		zap.Strings("cc", cc),
		zap.String("template-name", string(name)),
	)
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

	for _, receiverEmail := range sliceutils.Merge(to, cc) {
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

func (s *service) formMessage(name TemplateName, from string, to, cc []string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(sliceutils.Merge(to, cc), ",")))
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
