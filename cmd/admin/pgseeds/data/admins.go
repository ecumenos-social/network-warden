package data

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"
)

//go:embed admins.json
var holdersFile []byte

type Admin struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	Emails         []string  `json:"emails"`
	PhoneNumbers   []string  `json:"phone_numbers"`
	AvatarImageURL *string   `json:"avatar_image_url"`
	Countries      []string  `json:"countries"`
	Languages      []string  `json:"languages"`
	Password       string    `json:"password"`
}

func LoadAdmins(ctx context.Context) ([]Admin, error) {
	var out []Admin
	if err := json.Unmarshal(holdersFile, &out); err != nil {
		return nil, err
	}

	return out, nil
}
