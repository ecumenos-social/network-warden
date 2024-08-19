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

func (r *Repository) InsertAdmin(ctx context.Context, admin *models.Admin) error {
	query := `insert into public.admins
  (id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url, countries, languages, password_hash)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	params := []interface{}{
		admin.ID, admin.CreatedAt, admin.LastModifiedAt,
		admin.Emails, admin.PhoneNumbers, admin.AvatarImageURL, admin.Countries, admin.Languages,
		admin.PasswordHash,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) ModifyAdmin(ctx context.Context, id int64, admin *models.Admin) error {
	query := `update public.admins
  set created_at=$2, last_modified_at=$3, emails=$4, phone_numbers=$5, avatar_image_url=$6, countries=$7, languages=$8, password_hash=$9
  where id=$1;`
	params := []interface{}{
		admin.ID, admin.CreatedAt, admin.LastModifiedAt,
		admin.Emails, admin.PhoneNumbers, admin.AvatarImageURL, admin.Countries, admin.Languages,
		admin.PasswordHash,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) InsertAdminSession(ctx context.Context, adminSession *models.AdminSession) error {
	query := `insert into public.admin_sessions
  (id, created_at, last_modified_at, admin_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	params := []interface{}{
		adminSession.ID, adminSession.CreatedAt, adminSession.LastModifiedAt,
		adminSession.AdminID, adminSession.Token, adminSession.RefreshToken, adminSession.ExpiredAt,
		adminSession.RemoteIPAddress, adminSession.RemoteMACAddress,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) scanAdminSession(rows scanner) (*models.AdminSession, error) {
	var hs models.AdminSession
	err := rows.Scan(
		&hs.ID,
		&hs.CreatedAt,
		&hs.LastModifiedAt,
		&hs.AdminID,
		&hs.Token,
		&hs.RefreshToken,
		&hs.ExpiredAt,
		&hs.RemoteIPAddress,
		&hs.RemoteMACAddress,
	)
	return &hs, err
}

func (r *Repository) GetAdminSessionByRefreshToken(ctx context.Context, refToken string) (*models.AdminSession, error) {
	q := `
  select
  id, created_at, last_modified_at, admin_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address
  from public.admin_sessions
  where refresh_token=$1;`
	row, err := r.driver.QueryRow(ctx, q, refToken)
	if err != nil {
		return nil, err
	}

	hs, err := r.scanAdminSession(row)
	if err == nil {
		return hs, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetAdminSessionByToken(ctx context.Context, token string) (*models.AdminSession, error) {
	q := `
  select
  id, created_at, last_modified_at, admin_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address
  from public.admin_sessions
  where token=$1;`
	row, err := r.driver.QueryRow(ctx, q, token)
	if err != nil {
		return nil, err
	}

	hs, err := r.scanAdminSession(row)
	if err == nil {
		return hs, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) ModifyAdminSession(ctx context.Context, id int64, adminSession *models.AdminSession) error {
	query := `update public.admin_sessions
  set created_at=$2, last_modified_at=$3, admin_id=$4, token=$5, refresh_token=$6, expired_at=$7, remote_ip_address=$8, remote_mac_address=$9
  where id=$1;`
	params := []interface{}{
		adminSession.ID, adminSession.CreatedAt, adminSession.LastModifiedAt,
		adminSession.AdminID, adminSession.Token, adminSession.RefreshToken, adminSession.ExpiredAt,
		adminSession.RemoteIPAddress, adminSession.RemoteMACAddress,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}
