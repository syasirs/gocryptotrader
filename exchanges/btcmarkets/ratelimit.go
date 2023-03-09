package btcmarkets

import (
	"context"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"golang.org/x/time/rate"
)

// BTCMarkets Rate limit consts
const (
	btcmarketsRateInterval         = time.Second * 10
	btcmarketsAuthLimit            = 50
	btcmarketsUnauthLimit          = 50
	btcmarketsOrderLimit           = 30
	btcmarketsBatchOrderLimit      = 5
	btcmarketsWithdrawLimit        = 10
	btcmarketsCreateNewReportLimit = 1

	// Used to match endpoints to rate limits
	orderFunc request.EndpointLimit = iota
	batchFunc
	withdrawFunc
	newReportFunc
)

// RateLimit implements the request.Limiter interface
type RateLimit struct {
	Auth            *rate.Limiter
	UnAuth          *rate.Limiter
	OrderPlacement  *rate.Limiter
	BatchOrders     *rate.Limiter
	WithdrawRequest *rate.Limiter
	CreateNewReport *rate.Limiter
}

// Limit limits the outbound requests
func (r *RateLimit) Limit(ctx context.Context, f request.EndpointLimit) (*rate.Limiter, int, error) {
	switch f {
	case request.Auth:
		return r.Auth, 1, nil
	case orderFunc:
		return r.OrderPlacement, 1, nil
	case batchFunc:
		return r.BatchOrders, 1, nil
	case withdrawFunc:
		return r.WithdrawRequest, 1, nil
	case newReportFunc:
		return r.CreateNewReport, 1, nil
	default:
		return r.UnAuth, 1, nil
	}
}

// SetRateLimit returns the rate limit for the exchange
func SetRateLimit() *RateLimit {
	return &RateLimit{
		Auth:            request.NewRateLimit(btcmarketsRateInterval, btcmarketsAuthLimit),
		UnAuth:          request.NewRateLimit(btcmarketsRateInterval, btcmarketsUnauthLimit),
		OrderPlacement:  request.NewRateLimit(btcmarketsRateInterval, btcmarketsOrderLimit),
		BatchOrders:     request.NewRateLimit(btcmarketsRateInterval, btcmarketsBatchOrderLimit),
		WithdrawRequest: request.NewRateLimit(btcmarketsRateInterval, btcmarketsWithdrawLimit),
		CreateNewReport: request.NewRateLimit(btcmarketsRateInterval, btcmarketsCreateNewReportLimit),
	}
}
