package jwt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

type Config struct {
	SigningKey      string
	TokenAge        time.Duration
	RefreshTokenAge time.Duration
}

type Service interface {
	CreateTokens(ctx context.Context, logger *zap.Logger, subj string) (string, string, error)
	DecodeToken(logger *zap.Logger, token string) (jwt.Token, error)
}

type service struct {
	jwtSigningKey   []byte
	tokenAge        time.Duration
	refreshTokenAge time.Duration
}

func New(config *Config) Service {
	return &service{
		jwtSigningKey:   []byte(config.SigningKey),
		tokenAge:        config.TokenAge,
		refreshTokenAge: config.RefreshTokenAge,
	}
}

func makeToken(subject string, scope TokenScope, exp time.Time) jwt.Token {
	tok := jwt.New()
	tok.Set("scope", scope)
	tok.Set("sub", subject)
	tok.Set("iat", time.Now().Unix())
	tok.Set("exp", exp.Unix())

	return tok
}

type TokenScope string

func (s TokenScope) String() string {
	return string(s)
}

const TokenScopeAccess TokenScope = "access"
const TokenScopeRefresh TokenScope = "refresh"

func (s *service) CreateTokens(ctx context.Context, logger *zap.Logger, subj string) (string, string, error) {
	tokExp := time.Now().Add(s.tokenAge)
	refTokExp := time.Now().Add(s.refreshTokenAge)
	accessTok := makeToken(subj, TokenScopeAccess, tokExp)
	refreshTok := makeToken(subj, TokenScopeRefresh, refTokExp)

	rVal := make([]byte, 10)
	rand.Read(rVal)
	refreshTok.Set("jti", base64.StdEncoding.EncodeToString(rVal))

	accSig, err := jwt.Sign(accessTok, jwt.WithKey(jwa.HS256, s.jwtSigningKey))
	if err != nil {
		logger.Error("signing access token error", zap.Error(err))
		return "", "", errorwrapper.WrapMessage(err, "signing access token")
	}

	refSig, err := jwt.Sign(refreshTok, jwt.WithKey(jwa.HS256, s.jwtSigningKey))
	if err != nil {
		logger.Error("signing refresh token error", zap.Error(err))
		return "", "", errorwrapper.WrapMessage(err, "signing refresh token")
	}

	return string(accSig), string(refSig), nil
}

func (s *service) DecodeToken(logger *zap.Logger, token string) (jwt.Token, error) {
	t, err := jwt.ParseString(token, jwt.WithVerify(false), jwt.WithValidate(true))
	if err != nil {
		logger.Error("decode token error", zap.Error(err))
		return nil, errorwrapper.WrapMessage(err, "decode token error")
	}

	return t, nil
}
