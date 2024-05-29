package emailer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

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

type Service interface {
	SendConfirmationOfRegistration(ctx context.Context, email, fullName, code string) error
}

type service struct {
	smtpAddr           string
	smtpAuth           smtp.Auth
	senderEmailAddress string

	logger *zap.Logger
}

func New(config *Config, logger *zap.Logger) Service {
	return &service{
		smtpAddr: net.JoinHostPort(config.SMTPHost, config.SMTPPort),
		smtpAuth: smtp.PlainAuth(
			"",
			config.SenderUsername,
			config.SenderPassword,
			config.SMTPHost,
		),
		senderEmailAddress: config.SenderEmailAddress,

		logger: logger,
	}
}

func (s *service) SendConfirmationOfRegistration(ctx context.Context, email, fullName, code string) error {
	return s.sendTemplate(
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

func (s *service) sendTemplate(name TemplateName, to, cc, files []string, data interface{}) error {
	l := s.logger.With(
		zap.Strings("to", to),
		zap.String("template-name", string(name)),
	)
	message, err := s.formMessage(name, s.senderEmailAddress, to, cc, data)
	if err != nil {
		l.Info("failed to compose email message", zap.Error(err))
		return err
	}
	if err := smtp.SendMail(s.smtpAddr, s.smtpAuth, s.senderEmailAddress, to, message); err != nil {
		l.Info("failed to send email", zap.Error(err))
		return err
	}
	l.Info("email was sent successfully")

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
