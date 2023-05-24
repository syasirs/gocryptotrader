package margin

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

var (
	// ErrInvalidMarginType returned when the margin type is invalid
	ErrInvalidMarginType = errors.New("invalid margin type")
	// ErrMarginTypeUnsupported returned when the margin type is unsupported
	ErrMarginTypeUnsupported = errors.New("unsupported margin type")
	// ErrNewAllocatedMarginRequired returned when the new allocated margin is missing
	ErrNewAllocatedMarginRequired = errors.New("new allocated margin required")
	// ErrOriginalPositionMarginRequired
	ErrOriginalPositionMarginRequired = errors.New("original allocated margin required")
)

// RateHistoryRequest is used to request a funding rate
type RateHistoryRequest struct {
	Exchange           string
	Asset              asset.Item
	Currency           currency.Code
	StartDate          time.Time
	EndDate            time.Time
	GetPredictedRate   bool
	GetLendingPayments bool
	GetBorrowRates     bool
	GetBorrowCosts     bool

	// CalculateOffline allows for the borrow rate, lending payment amount
	// and borrow costs to be calculated offline. It requires the takerfeerate
	// and existing rates
	CalculateOffline bool
	TakeFeeRate      decimal.Decimal
	// Rates is used when calculating offline and determiningPayments
	// Each Rate must have the Rate and Size fields populated
	Rates []Rate
}

type PositionChangeRequest struct {
	// Required fields
	Exchange string
	Pair     currency.Pair
	Asset    asset.Item
	// Optional fields depending on desired outcome
	MarginType              Type
	OriginalAllocatedMargin float64
	NewAllocatedMargin      float64
	MarginSide              string
}

type PositionChangeResponse struct {
	Exchange        string
	Pair            currency.Pair
	Asset           asset.Item
	AllocatedMargin float64
	MarginType      Type
}

// Type defines the different margin types supported by exchanges
type Type uint8

// Margin types
const (
	// Unset is the default value
	Unset = Type(0)
	// Isolated means a margin trade is isolated from other margin trades
	Isolated Type = 1 << (iota - 1)
	// Multi means a margin trade is not isolated from other margin trades
	Multi
	// Unknown is an unknown margin type but is not unset
	Unknown
)

var supported = Isolated | Multi

const (
	unsetStr    = "unset"
	isolatedStr = "isolated"
	multiStr    = "multi"
	crossedStr  = "crossed"
	unknownStr  = "unknown"
)

// RateHistoryResponse has the funding rate details
type RateHistoryResponse struct {
	Rates              []Rate
	SumBorrowCosts     decimal.Decimal
	AverageBorrowSize  decimal.Decimal
	SumLendingPayments decimal.Decimal
	AverageLendingSize decimal.Decimal
	PredictedRate      Rate
	TakerFeeRate       decimal.Decimal
}

// Rate has the funding rate details
// and optionally the borrow rate
type Rate struct {
	Time             time.Time
	MarketBorrowSize decimal.Decimal
	HourlyRate       decimal.Decimal
	YearlyRate       decimal.Decimal
	HourlyBorrowRate decimal.Decimal
	YearlyBorrowRate decimal.Decimal
	LendingPayment   LendingPayment
	BorrowCost       BorrowCost
}

// LendingPayment contains a lending rate payment
type LendingPayment struct {
	Payment decimal.Decimal
	Size    decimal.Decimal
}

// BorrowCost contains the borrow rate costs
type BorrowCost struct {
	Cost decimal.Decimal
	Size decimal.Decimal
}
