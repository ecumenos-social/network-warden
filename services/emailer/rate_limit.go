package emailer

import (
	"time"

	"github.com/ecumenos-social/network-warden/models"
)

type RateLimit struct {
	MaxRequests int64
	Interval    time.Duration
}

func (rl *RateLimit) Exceed(ms []*models.SentEmail) bool {
	var (
		startTime = time.Now().Add(-rl.Interval)
		count     int64
	)
	for _, m := range ms {
		if m == nil {
			continue
		}
		if m.CreatedAt.After(startTime) {
			count++
		}
	}

	return count > rl.MaxRequests
}
