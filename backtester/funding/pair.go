package funding

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

var (
	// ErrNotPair is returned when a user requests funding pair details when it is a collateral pair
	ErrNotPair = errors.New("not a funding pair")
)

// BaseInitialFunds returns the initial funds
// from the base in a currency pair
func (p *Pair) BaseInitialFunds() decimal.Decimal {
	return p.base.initialFunds
}

// QuoteInitialFunds returns the initial funds
// from the quote in a currency pair
func (p *Pair) QuoteInitialFunds() decimal.Decimal {
	return p.quote.initialFunds
}

// BaseAvailable returns the available funds
// from the base in a currency pair
func (p *Pair) BaseAvailable() decimal.Decimal {
	return p.base.available
}

// QuoteAvailable returns the available funds
// from the quote in a currency pair
func (p *Pair) QuoteAvailable() decimal.Decimal {
	return p.quote.available
}

func (p *Pair) GetPairReader() (IPairReader, error) {
	return p, nil
}

func (p *Pair) GetCollateralReader() (ICollateralReader, error) {
	return nil, ErrNotCollateral
}

// Reserve allocates an amount of funds to be used at a later time
// it prevents multiple events from claiming the same resource
// changes which currency to affect based on the order side
func (p *Pair) Reserve(amount decimal.Decimal, side order.Side) error {
	switch side {
	case order.Buy:
		return p.quote.Reserve(amount)
	case order.Sell:
		return p.base.Reserve(amount)
	default:
		return fmt.Errorf("%w for %v %v %v. Unknown side %v",
			errCannotAllocate,
			p.base.exchange,
			p.base.asset,
			p.base.currency,
			side)
	}
}

// Release reduces the amount of funding reserved and adds any difference
// back to the available amount
// changes which currency to affect based on the order side
func (p *Pair) Release(amount, diff decimal.Decimal, side order.Side) error {
	switch side {
	case order.Buy:
		return p.quote.Release(amount, diff)
	case order.Sell:
		return p.base.Release(amount, diff)
	default:
		return fmt.Errorf("%w for %v %v %v. Unknown side %v",
			errCannotAllocate,
			p.base.exchange,
			p.base.asset,
			p.base.currency,
			side)
	}
}

// IncreaseAvailable adds funding to the available amount
// changes which currency to affect based on the order side
func (p *Pair) IncreaseAvailable(amount decimal.Decimal, side order.Side) {
	switch side {
	case order.Buy:
		p.base.IncreaseAvailable(amount)
	case order.Sell:
		p.quote.IncreaseAvailable(amount)
	}
}

// CanPlaceOrder does a > 0 check to see if there are any funds
// to place an order with
// changes which currency to affect based on the order side
func (p *Pair) CanPlaceOrder(side order.Side) bool {
	switch side {
	case order.Buy:
		return p.quote.CanPlaceOrder()
	case order.Sell:
		return p.base.CanPlaceOrder()
	}
	return false
}

// FundReader returns a fund reader interface of the pair
func (p *Pair) FundReader() IFundReader {
	return p
}

// FundReserver returns a fund reserver interface of the pair
func (p *Pair) FundReserver() IFundReserver {
	return p
}

// PairReleaser returns a pair releaser interface of the pair
func (p *Pair) PairReleaser() (IPairReleaser, error) {
	if p == nil {
		return nil, ErrNilPair
	}
	return p, nil
}

// CollateralReleaser returns an error because a pair is not collateral
func (p *Pair) CollateralReleaser() (ICollateralReleaser, error) {
	return nil, ErrNotCollateral
}

// FundReleaser returns a pair releaser interface of the pair
func (p *Pair) FundReleaser() IFundReleaser {
	return p
}

// Liquidate basic liquidation response to remove
// all asset value
func (p *Pair) Liquidate() {
	p.base.available = decimal.Zero
	p.base.reserved = decimal.Zero
	p.quote.available = decimal.Zero
	p.quote.reserved = decimal.Zero
}
