package hitbtc

import (
	"context"
	"errors"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"golang.org/x/time/rate"
)

const (
	hitbtcRateInterval      = time.Second
	hitbtcMarketDataReqRate = 100
	hitbtcTradingReqRate    = 300
	hitbtcAllOthers         = 10

	marketRequests request.EndpointLimit = iota
	tradingRequests
	otherRequests
)

// RateLimit implements the request.Limiter interface
type RateLimit struct {
	MarketData *rate.Limiter
	Trading    *rate.Limiter
	Other      *rate.Limiter
}

// Limit limits outbound requests
func (r *RateLimit) Limit(ctx context.Context, f request.EndpointLimit) (*rate.Limiter, request.Tokens, error) {
	switch f {
	case marketRequests:
		return r.MarketData, 1, nil
	case tradingRequests:
		return r.Trading, 1, nil
	case otherRequests:
		return r.Other, 1, nil
	default:
		return nil, 0, errors.New("functionality not found")
	}
}

// SetRateLimit returns the rate limit for the exchange
func SetRateLimit() *RateLimit {
	return &RateLimit{
		MarketData: request.NewRateLimit(hitbtcRateInterval, hitbtcMarketDataReqRate),
		Trading:    request.NewRateLimit(hitbtcRateInterval, hitbtcTradingReqRate),
		Other:      request.NewRateLimit(hitbtcRateInterval, hitbtcAllOthers),
	}
}
