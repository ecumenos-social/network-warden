package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	"github.com/jackc/pgx/v4"
)

type Repository struct {
	driver fxpostgres.Driver
}

func New(driver fxpostgres.Driver) *Repository {
	return &Repository{driver: driver}
}

func (r *Repository) scanHolder(rows pgx.Rows) (*models.Holder, error) {
	var h models.Holder
	err := rows.Scan(
		&h.ID,
		&h.CreatedAt,
		&h.LastModifiedAt,
		&h.Emails,
		&h.PhoneNumbers,
		&h.AvatarImageURL,
		&h.Countries,
		&h.Languages,
		&h.PasswordHash,
		&h.Confirmed,
		&h.ConfirmationCode,
	)
	return &h, err
}

func (r *Repository) GetHoldersByEmails(ctx context.Context, emails []string) ([]*models.Holder, error) {
	q := fmt.Sprintf(`
    select
        id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
        countries, languages, password_hash, confirmed, confirmation_code
    from public.holders
    where emails && array[%s]::text[];`, "'"+strings.Join(emails, "', '")+"'")
	rows, err := r.driver.QueryRows(ctx, q)
	if err != nil {
		return nil, err
	}
	var out []*models.Holder

	for rows.Next() {
		h, err := r.scanHolder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetHoldersByPhoneNumbers(ctx context.Context, phoneNumbers []string) ([]*models.Holder, error) {
	q := fmt.Sprintf(`
    select
        id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
        countries, languages, password_hash, confirmed, confirmation_code
    from public.holders
    where phone_numbers && array[%s]::text[];`, "'"+strings.Join(phoneNumbers, "', '")+"'")
	rows, err := r.driver.QueryRows(ctx, q)
	if err != nil {
		return nil, err
	}
	var out []*models.Holder

	for rows.Next() {
		h, err := r.scanHolder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) InsertHolder(ctx context.Context, holder *models.Holder) error {
	query := `insert into public.holders
  (id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url, countries, languages, password_hash, confirmed, confirmation_code)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`
	params := []interface{}{
		holder.ID, holder.CreatedAt, holder.LastModifiedAt,
		holder.Emails, holder.PhoneNumbers, holder.AvatarImageURL, holder.Countries, holder.Languages,
		holder.PasswordHash, holder.Confirmed, holder.ConfirmationCode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) InsertHolderSession(ctx context.Context, holderSession *models.HolderSession) error {
	query := `insert into public.holder_sessions
  (id, created_at, last_modified_at, holder_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	params := []interface{}{
		holderSession.ID, holderSession.CreatedAt, holderSession.LastModifiedAt,
		holderSession.HolderID, holderSession.Token, holderSession.RefreshToken, holderSession.ExpiredAt,
		holderSession.RemoteIPAddress, holderSession.RemoteMACAddress,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}
