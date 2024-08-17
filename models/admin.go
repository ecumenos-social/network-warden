package models

import (
	"database/sql"
	"time"
)

type Admin struct {
	ID             int64          `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	LastModifiedAt time.Time      `json:"last_modified_at"`
	Emails         []string       `json:"emails"`
	PhoneNumbers   []string       `json:"phone_numbers"`
	AvatarImageURL sql.NullString `json:"avatar_image_url"`
	Countries      []string       `json:"countries"`
	Languages      []string       `json:"languages"`
	PasswordHash   string         `json:"password_hash"`
}
