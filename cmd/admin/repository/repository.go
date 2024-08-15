package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/toolkit/primitives"
	"github.com/ecumenos-social/toolkitfx/fxpostgres"
	"github.com/jackc/pgx/v4"
)

type Repository struct {
	driver fxpostgres.Driver
}

func New(driver fxpostgres.Driver) *Repository {
	return &Repository{driver: driver}
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *Repository) scanAdmin(rows scanner) (*models.Admin, error) {
	var a models.Admin
	err := rows.Scan(
		&a.ID,
		&a.CreatedAt,
		&a.LastModifiedAt,
		&a.Emails,
		&a.PhoneNumbers,
		&a.AvatarImageURL,
		&a.Countries,
		&a.Languages,
		&a.PasswordHash,
	)
	return &a, err
}

func (r *Repository) GetAdminsByEmails(ctx context.Context, emails []string) ([]*models.Admin, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash
    from public.admins
    where emails && array[%s]::text[];`, "'"+strings.Join(emails, "', '")+"'")
	rows, err := r.driver.QueryRows(ctx, q)
	if err != nil {
		return nil, err
	}
	var out []*models.Admin

	for rows.Next() {
		a, err := r.scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetAdminsByPhoneNumbers(ctx context.Context, phoneNumbers []string) ([]*models.Admin, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash
    from public.admins
    where phone_numbers && array[%s]::text[];`, "'"+strings.Join(phoneNumbers, "', '")+"'")
	rows, err := r.driver.QueryRows(ctx, q)
	if err != nil {
		return nil, err
	}
	var out []*models.Admin

	for rows.Next() {
		a, err := r.scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash
    from public.admins
    where emails && array['%s']::text[];`, email)
	row, err := r.driver.QueryRow(ctx, q)
	if err != nil {
		return nil, err
	}

	a, err := r.scanAdmin(row)
	if err == nil {
		return a, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetAdminByPhoneNumber(ctx context.Context, phoneNumber string) (*models.Admin, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash
    from public.admins
    where phone_numbers && array['%s']::text[];`, phoneNumber)
	row, err := r.driver.QueryRow(ctx, q)
	if err != nil {
		return nil, err
	}

	a, err := r.scanAdmin(row)
	if err == nil {
		return a, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetAdminByID(ctx context.Context, id int64) (*models.Admin, error) {
	q := `
  select
    id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
    countries, languages, password_hash
  from public.admins
  where id=$1;`
	row, err := r.driver.QueryRow(ctx, q, id)
	if err != nil {
		return nil, err
	}

	a, err := r.scanAdmin(row)
	if err == nil {
		return a, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) InsertAdmin(ctx context.Context, holder *models.Admin) error {
	query := `insert into public.admins
  (id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url, countries, languages, password_hash)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	params := []interface{}{
		holder.ID, holder.CreatedAt, holder.LastModifiedAt,
		holder.Emails, holder.PhoneNumbers, holder.AvatarImageURL, holder.Countries, holder.Languages,
		holder.PasswordHash,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}
