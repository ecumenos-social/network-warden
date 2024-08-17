package models

import (
	"database/sql"
	"time"
)

type NetworkNode struct {
	ID                        int64             `json:"id"`
	CreatedAt                 time.Time         `json:"created_at"`
	LastModifiedAt            time.Time         `json:"last_modified_at"`
	NetworkWardenID           int64             `json:"network_warden_id"`
	HolderID                  int64             `json:"holder_id"`
	Name                      string            `json:"name"`
	Description               string            `json:"description"`
	DomainName                string            `json:"domain_name"`
	Location                  *Location         `json:"location"`
	AccountsCapacity          int64             `json:"accounts_capacity"`
	Alive                     bool              `json:"alive"`
	LastPingedAt              sql.NullTime      `json:"last_pinged_at"`
	IsOpen                    bool              `json:"is_open"`
	IsInviteCodeRequired      bool              `json:"is_invite_code_required"`
	URL                       string            `json:"url"`
	APIKeyHash                string            `json:"api_key_hash"`
	Version                   string            `json:"version"`
	RateLimitMaxRequests      int64             `json:"rate_limit_max_requests"`
	RateLimitInterval         time.Duration     `json:"rate_limit_interval"`
	CrawlRateLimitMaxRequests int64             `json:"crawl_rate_limit_max_requests"`
	CrawlRateLimitInterval    time.Duration     `json:"crawl_rate_limit_interval"`
	Status                    NetworkNodeStatus `json:"status"`
	IDGenNode                 int64             `json:"id_gen_node"`
}

type NetworkNodeStatus string

const (
	NetworkNodeStatusApproved NetworkNodeStatus = "approved"
	NetworkNodeStatusPending  NetworkNodeStatus = "pending"
	NetworkNodeStatusRejected NetworkNodeStatus = "rejected"
)
