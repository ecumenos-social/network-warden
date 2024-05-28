package emailer

import "context"

type Config struct {
}

type Service interface {
	Send(ctx context.Context) error
}

type service struct {
}

func New(*Config) Service {
	return &service{}
}

func (s *service) Send(ctx context.Context) error {
	// TODO: add functionality here: https://ecumenos.atlassian.net/browse/NNW-26
	return nil
}
