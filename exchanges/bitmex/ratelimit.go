package bitmex

import (
	"context"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"golang.org/x/time/rate"
)

// Bitmex rate limits
const (
	bitmexRateInterval = time.Minute
	bitmexUnauthRate   = 30
	bitmexAuthRate     = 60
)

// RateLimit implements the request.Limiter interface
type RateLimit struct {
	Auth   *rate.Limiter
	UnAuth *rate.Limiter
}

// Limit limits outbound calls
func (r *RateLimit) Limit(ctx context.Context, f request.EndpointLimit) (*rate.Limiter, int, error) {
	if f == request.Auth {
		return r.Auth, 1, nil
	}
	return r.UnAuth, 1, nil
}

// SetRateLimit returns the rate limit for the exchange
func SetRateLimit() *RateLimit {
	return &RateLimit{
		Auth:   request.NewRateLimit(bitmexRateInterval, bitmexAuthRate),
		UnAuth: request.NewRateLimit(bitmexRateInterval, bitmexUnauthRate),
	}
}
