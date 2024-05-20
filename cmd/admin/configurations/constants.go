package configurations

import "time"

var (
	ServiceName    = "network-warden-admin"
	ServiceVersion = "v0.0.0"
	StartTimeout   = 5 * time.Minute
)
