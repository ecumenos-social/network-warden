package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Config struct {
	JWTSigningKey   string
	TokenAge        time.Duration
	RefreshTokenAge time.Duration
}

type Service interface {
	CreateTokens(ctx context.Context, subj string) (string, string, error)
	DecodeToken(token string) (jwt.Token, error)
}

type service struct {
	jwtSigningKey   []byte
	tokenAge        time.Duration
	refreshTokenAge time.Duration
}

func New(config *Config) Service {
	return &service{
		jwtSigningKey:   []byte(config.JWTSigningKey),
		tokenAge:        config.TokenAge,
		refreshTokenAge: config.RefreshTokenAge,
	}
}

func makeToken(subject string, scope string, exp time.Time) jwt.Token {
	tok := jwt.New()
	tok.Set("scope", scope)
	tok.Set("sub", subject)
	tok.Set("iat", time.Now().Unix())
	tok.Set("exp", exp.Unix())

	return tok
}

func (s *service) CreateTokens(ctx context.Context, subj string) (string, string, error) {
	tokExp := time.Now().Add(s.tokenAge)
	refTokExp := time.Now().Add(s.refreshTokenAge)
	accessTok := makeToken(subj, "access", tokExp)
	refreshTok := makeToken(subj, "refresh", refTokExp)

	rVal := make([]byte, 10)
	rand.Read(rVal)
	refreshTok.Set("jti", base64.StdEncoding.EncodeToString(rVal))

	accSig, err := jwt.Sign(accessTok, jwt.WithKey(jwa.HS256, s.jwtSigningKey))
	if err != nil {
		return "", "", fmt.Errorf("signing access token: %w", err)
	}

	refSig, err := jwt.Sign(refreshTok, jwt.WithKey(jwa.HS256, s.jwtSigningKey))
	if err != nil {
		return "", "", fmt.Errorf("signing refresh token: %w", err)
	}

	return string(accSig), string(refSig), nil
}

func (s *service) DecodeToken(token string) (jwt.Token, error) {
	t, err := jwt.ParseString(token, jwt.WithVerify(false), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("decode token err: %w", err)
	}
	return t, nil
}
