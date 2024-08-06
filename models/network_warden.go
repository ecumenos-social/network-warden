package models

import (
	"database/sql"
	"time"
)

type NetworkWarden struct {
	ID                        int64               `json:"id"`
	CreatedAt                 time.Time           `json:"created_at"`
	LastModifiedAt            time.Time           `json:"last_modified_at"`
	Label                     string              `json:"label"`
	Address                   string              `json:"address"`
	Name                      string              `json:"name"`
	Description               string              `json:"description"`
	Location                  *Location           `json:"location"`
	PDNCapacity               int64               `json:"pdn_capacity"`
	NNCapacity                int64               `json:"nn_capacity"`
	Alive                     bool                `json:"alive"`
	LastPingedAt              sql.NullTime        `json:"last_pinged_at"`
	IsOpen                    bool                `json:"is_open"`
	URL                       string              `json:"url"`
	Version                   string              `json:"version"`
	RateLimitMaxRequests      int64               `json:"rate_limit_max_requests"`
	RateLimitInterval         time.Duration       `json:"rate_limit_interval"`
	CrawlRateLimitMaxRequests int64               `json:"crawl_rate_limit_max_requests"`
	CrawlRateLimitInterval    time.Duration       `json:"crawl_rate_limit_interval"`
	Status                    NetworkWardenStatus `json:"status"`
	IDGenNode                 int64               `json:"id_gen_node"`
}

type NetworkWardenStatus string

const (
	NetworkWardenStatusApproved NetworkWardenStatus = "approved"
	NetworkWardenStatusPending  NetworkWardenStatus = "pending"
	NetworkWardenStatusRejected NetworkWardenStatus = "rejected"
)
