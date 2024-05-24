package models

import (
	"database/sql"
	"time"
)

type HolderSession struct {
	ID               int64          `json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	LastModifiedAt   time.Time      `json:"last_modified_at"`
	HolderID         int64          `json:"holder_id"`
	Token            string         `json:"token"`
	RefreshToken     string         `json:"refresh_token"`
	ExpiredAt        sql.NullTime   `json:"expired_at"`
	RemoteIPAddress  sql.NullString `json:"remote_ip_address"`
	RemoteMACAddress sql.NullString `json:"remote_mac_address"`
}
