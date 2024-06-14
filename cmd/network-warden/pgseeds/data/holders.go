package data

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"
)

//go:embed holders.json
var holdersFile []byte

type Holder struct {
	ID               int64     `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	LastModifiedAt   time.Time `json:"last_modified_at"`
	Emails           []string  `json:"emails"`
	PhoneNumbers     []string  `json:"phone_numbers"`
	AvatarImageURL   *string   `json:"avatar_image_url"`
	Countries        []string  `json:"countries"`
	Languages        []string  `json:"languages"`
	Password         string    `json:"password"`
	Confirmed        bool      `json:"confirmed"`
	ConfirmationCode string    `json:"confirmation_code"`
}

func LoadHolders(ctx context.Context) ([]Holder, error) {
	var out []Holder
	if err := json.Unmarshal(holdersFile, &out); err != nil {
		return nil, err
	}

	return out, nil
}
