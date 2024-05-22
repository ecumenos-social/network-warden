package repository

import "github.com/ecumenos-social/network-warden/pkg/fxpostgres"

type Repository struct {
	driver fxpostgres.Driver
}

func New(driver fxpostgres.Driver) *Repository {
	return &Repository{driver: driver}
}
