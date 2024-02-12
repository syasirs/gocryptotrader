package poloniex

import (
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

const (
	poloniexRateInterval = time.Second
	poloniexAuthRate     = 6
	poloniexUnauthRate   = 6
)

// SetRateLimit returns the rate limit for the exchange
// If your account's volume is over $5 million in 30 day volume,
// you may be eligible for an API rate limit increase.
// Please email poloniex@circle.com.
// As per https://docs.poloniex.com/#http-api
func SetRateLimit() request.RateLimitDefinitions {
	return request.RateLimitDefinitions{
		request.Auth:   request.NewRateLimit(poloniexRateInterval, poloniexAuthRate, 1),
		request.UnAuth: request.NewRateLimit(poloniexRateInterval, poloniexUnauthRate, 1),
	}
}
