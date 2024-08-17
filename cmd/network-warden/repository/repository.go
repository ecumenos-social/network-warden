package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecumenos-social/network-warden/models"
	"github.com/ecumenos-social/toolkit/primitives"
	"github.com/ecumenos-social/toolkit/types"
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

func (r *Repository) scanHolder(rows scanner) (*models.Holder, error) {
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

func (r *Repository) GetHolderByEmail(ctx context.Context, email string) (*models.Holder, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash, confirmed, confirmation_code
    from public.holders
    where emails && array['%s']::text[];`, email)
	row, err := r.driver.QueryRow(ctx, q)
	if err != nil {
		return nil, err
	}

	h, err := r.scanHolder(row)
	if err == nil {
		return h, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetHolderByPhoneNumber(ctx context.Context, phoneNumber string) (*models.Holder, error) {
	q := fmt.Sprintf(`
    select
      id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
      countries, languages, password_hash, confirmed, confirmation_code
    from public.holders
    where phone_numbers && array['%s']::text[];`, phoneNumber)
	row, err := r.driver.QueryRow(ctx, q)
	if err != nil {
		return nil, err
	}

	h, err := r.scanHolder(row)
	if err == nil {
		return h, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetHolderByID(ctx context.Context, id int64) (*models.Holder, error) {
	q := `
  select
    id, created_at, last_modified_at, emails, phone_numbers, avatar_image_url,
    countries, languages, password_hash, confirmed, confirmation_code
  from public.holders
  where id=$1;`
	row, err := r.driver.QueryRow(ctx, q, id)
	if err != nil {
		return nil, err
	}

	h, err := r.scanHolder(row)
	if err == nil {
		return h, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
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

func (r *Repository) ModifyHolder(ctx context.Context, id int64, holder *models.Holder) error {
	query := `update public.holders
  set created_at=$2, last_modified_at=$3, emails=$4, phone_numbers=$5, avatar_image_url=$6, countries=$7, languages=$8, password_hash=$9, confirmed=$10, confirmation_code=$11
  where id=$1;`
	params := []interface{}{
		holder.ID, holder.CreatedAt, holder.LastModifiedAt,
		holder.Emails, holder.PhoneNumbers, holder.AvatarImageURL, holder.Countries,
		holder.Languages, holder.PasswordHash, holder.Confirmed, holder.ConfirmationCode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) DeleteHolder(ctx context.Context, id int64) error {
	query := "delete from public.holders cascade where id=$1;"
	err := r.driver.ExecuteQuery(ctx, query, id)
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

func (r *Repository) scanHolderSession(rows scanner) (*models.HolderSession, error) {
	var hs models.HolderSession
	err := rows.Scan(
		&hs.ID,
		&hs.CreatedAt,
		&hs.LastModifiedAt,
		&hs.HolderID,
		&hs.Token,
		&hs.RefreshToken,
		&hs.ExpiredAt,
		&hs.RemoteIPAddress,
		&hs.RemoteMACAddress,
	)
	return &hs, err
}

func (r *Repository) GetHolderSessionByRefreshToken(ctx context.Context, refToken string) (*models.HolderSession, error) {
	q := `
  select
  id, created_at, last_modified_at, holder_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address
  from public.holder_sessions
  where refresh_token=$1;`
	row, err := r.driver.QueryRow(ctx, q, refToken)
	if err != nil {
		return nil, err
	}

	hs, err := r.scanHolderSession(row)
	if err == nil {
		return hs, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetHolderSessionByToken(ctx context.Context, token string) (*models.HolderSession, error) {
	q := `
  select
  id, created_at, last_modified_at, holder_id, token, refresh_token, expired_at, remote_ip_address, remote_mac_address
  from public.holder_sessions
  where token=$1;`
	row, err := r.driver.QueryRow(ctx, q, token)
	if err != nil {
		return nil, err
	}

	hs, err := r.scanHolderSession(row)
	if err == nil {
		return hs, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) ModifyHolderSession(ctx context.Context, id int64, holderSession *models.HolderSession) error {
	query := `update public.holder_sessions
  set created_at=$2, last_modified_at=$3, holder_id=$4, token=$5, refresh_token=$6, expired_at=$7, remote_ip_address=$8, remote_mac_address=$9
  where id=$1;`
	params := []interface{}{
		holderSession.ID, holderSession.CreatedAt, holderSession.LastModifiedAt,
		holderSession.HolderID, holderSession.Token, holderSession.RefreshToken, holderSession.ExpiredAt,
		holderSession.RemoteIPAddress, holderSession.RemoteMACAddress,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) InsertSentEmail(ctx context.Context, se *models.SentEmail) error {
	query := `insert into public.sent_emails
  (id, created_at, last_modified_at, sender_email, receiver_email, template_name)
  values ($1, $2, $3, $4, $5, $6);`
	params := []interface{}{se.ID, se.CreatedAt, se.LastModifiedAt, se.SenderEmail, se.ReceiverEmail, se.TemplateName}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) ModifySentEmail(ctx context.Context, id int64, se *models.SentEmail) error {
	query := `update public.sent_emails
  set created_at=$2, last_modified_at=$3, sender_email=$4, receiver_email=$5, template_name=$6
  where id=$1;`
	params := []interface{}{se.ID, se.CreatedAt, se.LastModifiedAt, se.SenderEmail, se.ReceiverEmail, se.TemplateName}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) scanSentEmail(rows scanner) (*models.SentEmail, error) {
	var se models.SentEmail
	err := rows.Scan(
		&se.ID,
		&se.CreatedAt,
		&se.LastModifiedAt,
		&se.SenderEmail,
		&se.ReceiverEmail,
		&se.TemplateName,
	)
	return &se, err
}

func (r *Repository) GetSentEmails(ctx context.Context, sender, receiver, templateName string) ([]*models.SentEmail, error) {
	q := `
  select
    id, created_at, last_modified_at, sender_email, receiver_email, template_name
  from public.sent_emails
  where sender_email=$1 and receiver_email=$2 and template_name=$3;`
	rows, err := r.driver.QueryRows(ctx, q, sender, receiver, templateName)
	if err != nil {
		return nil, err
	}
	var out []*models.SentEmail

	for rows.Next() {
		se, err := r.scanSentEmail(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, se)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) InsertNetworkNode(ctx context.Context, nn *models.NetworkNode) error {
	query := `insert into public.network_nodes
  (id, created_at, last_modified_at, network_warden_id, holder_id, name, description, domain_name, location,
   accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
   rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node)
  values ($1, $2, $3, $4, $5, $6, $7, $8, ST_SetSRID(ST_MakePoint($9, $10), 4326), $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);`
	params := []interface{}{
		nn.ID, nn.CreatedAt, nn.LastModifiedAt, nn.NetworkWardenID, nn.HolderID, nn.Name, nn.Description, nn.DomainName, nn.Location.Longitude, nn.Location.Latitude,
		nn.AccountsCapacity, nn.Alive, nn.LastPingedAt, nn.IsOpen, nn.URL, nn.APIKeyHash, nn.Version,
		nn.RateLimitMaxRequests, nn.RateLimitInterval, nn.CrawlRateLimitMaxRequests, nn.CrawlRateLimitInterval, nn.Status, nn.IDGenNode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) ModifyNetworkNode(ctx context.Context, id int64, nn *models.NetworkNode) error {
	query := `update public.network_nodes
  set created_at=$2, last_modified_at=$3, network_warden_id=$4, holder_id=$5, name=$6, description=$7, domain_name=$8, location=ST_SetSRID(ST_MakePoint($9, $10), 4326),
  accounts_capacity=$11, alive=$12, last_pinged_at=$13, is_open=$14, url=$15, api_key_hash=$16, version=$17,
  rate_limit_max_requests=$18, rate_limit_interval=$19, crawl_rate_limit_max_requests=$20, crawl_rate_limit_interval=$21, status=$22, id_gen_node=$23
  where id=$1;`
	params := []interface{}{
		nn.ID, nn.CreatedAt, nn.LastModifiedAt, nn.NetworkWardenID, nn.HolderID, nn.Name, nn.Description, nn.DomainName, nn.Location.Longitude, nn.Location.Latitude,
		nn.AccountsCapacity, nn.Alive, nn.LastPingedAt, nn.IsOpen, nn.URL, nn.APIKeyHash, nn.Version,
		nn.RateLimitMaxRequests, nn.RateLimitInterval, nn.CrawlRateLimitMaxRequests, nn.CrawlRateLimitInterval, nn.Status, nn.IDGenNode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) scanNetworkNode(rows scanner) (*models.NetworkNode, error) {
	var (
		nn       models.NetworkNode
		location models.Location
	)
	err := rows.Scan(
		&nn.ID,
		&nn.CreatedAt,
		&nn.LastModifiedAt,
		&nn.NetworkWardenID,
		&nn.HolderID,
		&nn.Name,
		&nn.Description,
		&nn.DomainName,
		&location.Longitude,
		&location.Latitude,
		&nn.AccountsCapacity,
		&nn.Alive,
		&nn.LastPingedAt,
		&nn.IsOpen,
		&nn.URL,
		&nn.APIKeyHash,
		&nn.Version,
		&nn.RateLimitMaxRequests,
		&nn.RateLimitInterval,
		&nn.CrawlRateLimitMaxRequests,
		&nn.CrawlRateLimitInterval,
		&nn.Status,
		&nn.IDGenNode,
	)
	if err != nil {
		return nil, err
	}
	nn.Location = &location

	return &nn, nil
}

func (r *Repository) GetNetworkNodesList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.NetworkNode, error) {
	var (
		whereStatements = make([]string, 0, len(filters))
		args            = make([]interface{}, 0, len(filters)+2)
	)
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	for field, value := range filters {
		args = append(args, value)
		whereStatements = append(whereStatements, field+"=$"+fmt.Sprint(len(args)))
	}
	var whereStatement string
	if len(whereStatements) > 0 {
		whereStatement = "where " + strings.Join(whereStatements, ", ")
	}

	q := fmt.Sprintf(`
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, name, description, domain_name, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.network_nodes %s limit $1 offset $2;`, whereStatement)
	rows, err := r.driver.QueryRows(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var out []*models.NetworkNode

	for rows.Next() {
		nn, err := r.scanNetworkNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, nn)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetNetworkNodeByDomainName(ctx context.Context, domainName string) (*models.NetworkNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, name, description, domain_name, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.network_nodes
  where domain_name=$1;`
	row, err := r.driver.QueryRow(ctx, q, domainName)
	if err != nil {
		return nil, err
	}

	nn, err := r.scanNetworkNode(row)
	if err == nil {
		return nn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetNetworkNodeByID(ctx context.Context, id int64) (*models.NetworkNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, name, description, domain_name, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.network_nodes
  where id=$1;`
	row, err := r.driver.QueryRow(ctx, q, id)
	if err != nil {
		return nil, err
	}

	nn, err := r.scanNetworkNode(row)
	if err == nil {
		return nn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetNetworkNodeByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.NetworkNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, name, description, domain_name, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.network_nodes
  where api_key_hash=$1;`
	row, err := r.driver.QueryRow(ctx, q, apiKeyHash)
	if err != nil {
		return nil, err
	}

	nn, err := r.scanNetworkNode(row)
	if err == nil {
		return nn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) InsertPersonalDataNode(ctx context.Context, pdn *models.PersonalDataNode) error {
	query := `insert into public.personal_data_nodes
  (id, created_at, last_modified_at, network_warden_id, holder_id, label, address, name, description, location,
   accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
   rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9, ST_SetSRID(ST_MakePoint($10, $11), 4326), $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24);`
	params := []interface{}{
		pdn.ID, pdn.CreatedAt, pdn.LastModifiedAt, pdn.NetworkWardenID, pdn.HolderID, pdn.Label, pdn.Address, pdn.Name, pdn.Description, pdn.Location.Longitude, pdn.Location.Latitude,
		pdn.AccountsCapacity, pdn.Alive, pdn.LastPingedAt, pdn.IsOpen, pdn.URL, pdn.APIKeyHash, pdn.Version,
		pdn.RateLimitMaxRequests, pdn.RateLimitInterval, pdn.CrawlRateLimitMaxRequests, pdn.CrawlRateLimitInterval, pdn.Status, pdn.IDGenNode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) ModifyPersonalDataNode(ctx context.Context, id int64, pdn *models.PersonalDataNode) error {
	query := `update public.personal_data_nodes
  set created_at=$2, last_modified_at=$3, network_warden_id=$4, holder_id=$5, label=$6, address=$7, name=$8, description=$9, location=ST_SetSRID(ST_MakePoint($10, $11), 4326),
  accounts_capacity=$12, alive=$13, last_pinged_at=$14, is_open=$15, url=$16, api_key_hash=$17, version=$18,
  rate_limit_max_requests=$19, rate_limit_interval=$20, crawl_rate_limit_max_requests=$21, crawl_rate_limit_interval=$22, status=$23, id_gen_node=$24
  where id=$1;`
	params := []interface{}{
		pdn.ID, pdn.CreatedAt, pdn.LastModifiedAt, pdn.NetworkWardenID, pdn.HolderID, pdn.Label, pdn.Address, pdn.Name, pdn.Description, pdn.Location.Longitude, pdn.Location.Latitude,
		pdn.AccountsCapacity, pdn.Alive, pdn.LastPingedAt, pdn.IsOpen, pdn.URL, pdn.APIKeyHash, pdn.Version,
		pdn.RateLimitMaxRequests, pdn.RateLimitInterval, pdn.CrawlRateLimitMaxRequests, pdn.CrawlRateLimitInterval, pdn.Status, pdn.IDGenNode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) scanPersonalDataNode(rows scanner) (*models.PersonalDataNode, error) {
	var (
		pdn      models.PersonalDataNode
		location models.Location
	)
	err := rows.Scan(
		&pdn.ID,
		&pdn.CreatedAt,
		&pdn.LastModifiedAt,
		&pdn.NetworkWardenID,
		&pdn.HolderID,
		&pdn.Label,
		&pdn.Address,
		&pdn.Name,
		&pdn.Description,
		&location.Longitude,
		&location.Latitude,
		&pdn.AccountsCapacity,
		&pdn.Alive,
		&pdn.LastPingedAt,
		&pdn.IsOpen,
		&pdn.URL,
		&pdn.APIKeyHash,
		&pdn.Version,
		&pdn.RateLimitMaxRequests,
		&pdn.RateLimitInterval,
		&pdn.CrawlRateLimitMaxRequests,
		&pdn.CrawlRateLimitInterval,
		&pdn.Status,
		&pdn.IDGenNode,
	)
	if err != nil {
		return nil, err
	}
	pdn.Location = &location

	return &pdn, nil
}

func (r *Repository) GetPersonalDataNodesList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.PersonalDataNode, error) {
	var (
		whereStatements = make([]string, 0, len(filters))
		args            = make([]interface{}, 0, len(filters)+2)
	)
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	for field, value := range filters {
		args = append(args, value)
		whereStatements = append(whereStatements, field+"=$"+fmt.Sprint(len(args)))
	}
	var whereStatement string
	if len(whereStatements) > 0 {
		whereStatement = "where " + strings.Join(whereStatements, ", ")
	}

	q := fmt.Sprintf(`
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.personal_data_nodes %s limit $1 offset $2;`, whereStatement)
	rows, err := r.driver.QueryRows(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var out []*models.PersonalDataNode

	for rows.Next() {
		pdn, err := r.scanPersonalDataNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, pdn)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetPersonalDataNodeByLabel(ctx context.Context, label string) (*models.PersonalDataNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.personal_data_nodes
  where label=$1;`
	row, err := r.driver.QueryRow(ctx, q, label)
	if err != nil {
		return nil, err
	}

	pdn, err := r.scanPersonalDataNode(row)
	if err == nil {
		return pdn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetPersonalDataNodeByID(ctx context.Context, id int64) (*models.PersonalDataNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.personal_data_nodes
  where id=$1;`
	row, err := r.driver.QueryRow(ctx, q, id)
	if err != nil {
		return nil, err
	}

	pdn, err := r.scanPersonalDataNode(row)
	if err == nil {
		return pdn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetPersonalDataNodeByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.PersonalDataNode, error) {
	q := `
  select
    id, created_at, last_modified_at, network_warden_id, holder_id, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    accounts_capacity, alive, last_pinged_at, is_open, url, api_key_hash, version,
    rate_limit_max_requests, rate_limit_interval, crawl_rate_limit_max_requests, crawl_rate_limit_interval, status, id_gen_node
  from public.personal_data_nodes
  where api_key_hash=$1;`
	row, err := r.driver.QueryRow(ctx, q, apiKeyHash)
	if err != nil {
		return nil, err
	}

	pdn, err := r.scanPersonalDataNode(row)
	if err == nil {
		return pdn, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) InsertNetworkWarden(ctx context.Context, nw *models.NetworkWarden) error {
	query := `insert into public.network_wardens
  (id, created_at, last_modified_at, label, address, name, description, location,
   pdn_capacity, nn_capacity, alive, last_pinged_at, is_open, url, version,
   rate_limit_max_requests, rate_limit_interval, id_gen_node)
  values ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($8, $9), 4326), $10, $11, $12, $13, $14, $15, $16, $17, $18, $19);`
	params := []interface{}{
		nw.ID, nw.CreatedAt, nw.LastModifiedAt, nw.Label, nw.Address, nw.Name, nw.Description, nw.Location.Longitude, nw.Location.Latitude,
		nw.PDNCapacity, nw.NNCapacity, nw.Alive, nw.LastPingedAt, nw.IsOpen, nw.URL, nw.Version,
		nw.RateLimitMaxRequests, nw.RateLimitInterval, nw.IDGenNode,
	}
	err := r.driver.ExecuteQuery(ctx, query, params...)
	return err
}

func (r *Repository) scanNetworkWarden(rows scanner) (*models.NetworkWarden, error) {
	var (
		nw       models.NetworkWarden
		location models.Location
	)
	err := rows.Scan(
		&nw.ID,
		&nw.CreatedAt,
		&nw.LastModifiedAt,
		&nw.Label,
		&nw.Address,
		&nw.Name,
		&nw.Description,
		&location.Longitude,
		&location.Latitude,
		&nw.PDNCapacity,
		&nw.NNCapacity,
		&nw.Alive,
		&nw.LastPingedAt,
		&nw.IsOpen,
		&nw.URL,
		&nw.Version,
		&nw.RateLimitMaxRequests,
		&nw.RateLimitInterval,
		&nw.IDGenNode,
	)
	if err != nil {
		return nil, err
	}
	nw.Location = &location

	return &nw, nil
}

func (r *Repository) GetNetworkWardensList(ctx context.Context, filters map[string]interface{}, pagination *types.Pagination) ([]*models.NetworkWarden, error) {
	var (
		whereStatements = make([]string, 0, len(filters))
		args            = make([]interface{}, 0, len(filters)+2)
	)
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	for field, value := range filters {
		args = append(args, value)
		whereStatements = append(whereStatements, field+"=$"+fmt.Sprint(len(args)))
	}
	var whereStatement string
	if len(whereStatements) > 0 {
		whereStatement = "where " + strings.Join(whereStatements, ", ")
	}

	q := fmt.Sprintf(`
  select
    id, created_at, last_modified_at, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    pdn_capacity, nn_capacity, alive, last_pinged_at, is_open, url, version,
    rate_limit_max_requests, rate_limit_interval, id_gen_node
  from public.network_wardens %s limit $1 offset $2;`, whereStatement)
	rows, err := r.driver.QueryRows(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var out []*models.NetworkWarden

	for rows.Next() {
		nw, err := r.scanNetworkWarden(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, nw)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) GetNetworkWardenByLabel(ctx context.Context, label string) (*models.NetworkWarden, error) {
	q := `
  select
    id, created_at, last_modified_at, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    pdn_capacity, nn_capacity, alive, last_pinged_at, is_open, url, version,
    rate_limit_max_requests, rate_limit_interval, id_gen_node
  from public.network_wardens
  where label=$1;`
	row, err := r.driver.QueryRow(ctx, q, label)
	if err != nil {
		return nil, err
	}

	nw, err := r.scanNetworkWarden(row)
	if err == nil {
		return nw, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func (r *Repository) GetNetworkWardenByID(ctx context.Context, id int64) (*models.NetworkWarden, error) {
	q := `
  select
    id, created_at, last_modified_at, label, address, name, description, ST_X(location::geometry), ST_Y(location::geometry),
    pdn_capacity, nn_capacity, alive, last_pinged_at, is_open, url, version,
    rate_limit_max_requests, rate_limit_interval, id_gen_node
  from public.network_wardens
  where id=$1;`
	row, err := r.driver.QueryRow(ctx, q, id)
	if err != nil {
		return nil, err
	}

	nw, err := r.scanNetworkWarden(row)
	if err == nil {
		return nw, nil
	}

	if primitives.IsSameError(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}
