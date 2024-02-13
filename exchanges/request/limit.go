package request

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Defines rate limiting errors
var (
	ErrRateLimiterAlreadyDisabled = errors.New("rate limiter already disabled")
	ErrRateLimiterAlreadyEnabled  = errors.New("rate limiter already enabled")

	errLimiterSystemIsNil       = errors.New("limiter system is nil")
	errInvalidTokenCount        = errors.New("invalid token count must equal or greater than 1")
	errSpecificRateLimiterIsNil = errors.New("specific rate limiter is nil")
)

// Const here define individual functionality sub types for rate limiting
const (
	Unset EndpointLimit = iota
	Auth
	UnAuth
)

// BasicLimit denotes basic rate limit that implements the Limiter interface
// does not need to set endpoint functionality.
type BasicLimit struct {
	r *RateLimiterWithToken
}

// Limit executes a single rate limit set by NewRateLimit
func (b *BasicLimit) Limit(context.Context, EndpointLimit) (*RateLimiterWithToken, error) {
	return b.r, nil
}

// EndpointLimit defines individual endpoint rate limits that are set when
// New is called.
type EndpointLimit uint16

// Tokens defines the number of tokens to be consumed. This is a generalised
// weight for rate limiting. e.g. n token = n request. i.e. 50 tokens = 50
// requests.
type Tokens uint8

// RateLimitDefinitions is a map of endpoint limits to rate limiters
type RateLimitDefinitions map[interface{}]*RateLimiterWithToken

// RateLimiterWithToken is a rate limiter coupled with a token count which
// refers to the number or weighting of the request. This is used to define
// the rate limit for a specific endpoint.
type RateLimiterWithToken struct {
	*rate.Limiter
	Tokens
}

// Reservations is a slice of rate reservations
type Reservations []*rate.Reservation

// CancelAll cancels all potential reservations to free up rate limiter for
// context cancellations and deadline exceeded cases.
func (r Reservations) CancelAll() {
	for x := range r {
		r[x].Cancel()
	}
}

// NewRateLimit creates a new RateLimit based of time interval and how many
// actions allowed and breaks it down to an actions-per-second basis -- Burst
// rate is kept as one as this is not supported for out-bound requests.
func NewRateLimit(interval time.Duration, actions int) *rate.Limiter {
	if actions <= 0 || interval <= 0 {
		// Returns an un-restricted rate limiter
		return rate.NewLimiter(rate.Inf, 1)
	}

	i := 1 / interval.Seconds()
	rps := i * float64(actions)
	return rate.NewLimiter(rate.Limit(rps), 1)
}

// NewRateLimitWithToken creates a new RateLimit based of time interval and how
// many actions allowed. This also has a token count which refers to the number
// or weighting of the request. This is used to define the rate limit for a
// specific endpoint.
func NewRateLimitWithToken(interval time.Duration, actions int, tokens Tokens) *RateLimiterWithToken {
	return GetRateLimiterWithToken(NewRateLimit(interval, actions), tokens)
}

func GetRateLimiterWithToken(l *rate.Limiter, t Tokens) *RateLimiterWithToken {
	return &RateLimiterWithToken{l, t}
}

// NewBasicRateLimit returns an object that implements the limiter interface
// for basic rate limit
func NewBasicRateLimit(interval time.Duration, actions int, tokens Tokens) RateLimitDefinitions {
	rl := NewRateLimitWithToken(interval, actions, tokens)
	return RateLimitDefinitions{Unset: rl, Auth: rl, UnAuth: rl}
}

// InitiateRateLimit sleeps for designated end point rate limits
func (r *Requester) InitiateRateLimit(ctx context.Context, e EndpointLimit) error {
	if r == nil {
		return ErrRequestSystemIsNil
	}
	if atomic.LoadInt32(&r.disableRateLimiter) == 1 {
		return nil
	}
	if r.limiter == nil {
		return fmt.Errorf("cannot rate limit request %w", errLimiterSystemIsNil)
	}

	rateLimiter := r.limiter[e]

	if rateLimiter == nil {
		return fmt.Errorf("cannot rate limit request %w for endpoint %d", errSpecificRateLimiterIsNil, e)
	}

	if rateLimiter.Tokens <= 0 {
		return fmt.Errorf("cannot rate limit request %w for endpoint %d", errInvalidTokenCount, e)
	}

	var finalDelay time.Duration
	var reservations = make(Reservations, rateLimiter.Tokens)
	for i := Tokens(0); i < rateLimiter.Tokens; i++ {
		// Consume tokens 1 at a time as this avoids needing burst capacity in the limiter,
		// which would otherwise allow the rate limit to be exceeded over short periods
		reservations[i] = rateLimiter.Reserve()
		finalDelay = reservations[i].Delay()
	}

	if dl, ok := ctx.Deadline(); ok && dl.Before(time.Now().Add(finalDelay)) {
		reservations.CancelAll()
		return fmt.Errorf("rate limit delay of %s will exceed deadline: %w",
			finalDelay,
			context.DeadlineExceeded)
	}

	tick := time.NewTimer(finalDelay)
	select {
	case <-tick.C:
		return nil
	case <-ctx.Done():
		tick.Stop()
		reservations.CancelAll()
		return ctx.Err()
	}
	// TODO: Shutdown case
}

// DisableRateLimiter disables the rate limiting system for the exchange
func (r *Requester) DisableRateLimiter() error {
	if r == nil {
		return ErrRequestSystemIsNil
	}
	if !atomic.CompareAndSwapInt32(&r.disableRateLimiter, 0, 1) {
		return fmt.Errorf("%s %w", r.name, ErrRateLimiterAlreadyDisabled)
	}
	return nil
}

// EnableRateLimiter enables the rate limiting system for the exchange
func (r *Requester) EnableRateLimiter() error {
	if r == nil {
		return ErrRequestSystemIsNil
	}
	if !atomic.CompareAndSwapInt32(&r.disableRateLimiter, 1, 0) {
		return fmt.Errorf("%s %w", r.name, ErrRateLimiterAlreadyEnabled)
	}
	return nil
}
