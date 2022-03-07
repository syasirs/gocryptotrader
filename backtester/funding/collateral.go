package funding

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

var (
	// ErrNotCollateral is returned when a user requests collateral from a non-collateral pair
	ErrNotCollateral = errors.New("not a collateral pair")
	ErrNilPair       = errors.New("nil pair")
)

// TODO consider moving futures tracking to funding
// we're already passing around funding items, it can then also have all the lovely tracking attached?

func (c *Collateral) CanPlaceOrder(_ order.Side) bool {
	return c.Collateral.CanPlaceOrder()
}

func (c *Collateral) TakeProfit(contracts, positionReturns decimal.Decimal) error {
	err := c.Contract.ReduceContracts(contracts)
	if err != nil {
		return err
	}
	return c.Collateral.TakeProfit(positionReturns)
}

func (c *Collateral) ContractCurrency() currency.Code {
	return c.Contract.currency
}

func (c *Collateral) UnderlyingAsset() currency.Code {
	// somehow get the underlying
	return c.Contract.currency
}

func (c *Collateral) CollateralCurrency() currency.Code {
	return c.Collateral.currency
}

func (c *Collateral) InitialFunds() decimal.Decimal {
	return c.Collateral.initialFunds
}

func (c *Collateral) AvailableFunds() decimal.Decimal {
	return c.Collateral.available
}

func (c *Collateral) GetPairReader() (IPairReader, error) {
	return nil, fmt.Errorf("could not return pair reader for %v %v %v %v %w", c.Contract.exchange, c.Collateral.asset, c.ContractCurrency(), c.CollateralCurrency(), ErrNotPair)
}

func (c *Collateral) GetCollateralReader() (ICollateralReader, error) {
	return c, nil
}

func (c *Collateral) UpdateCollateral(amount decimal.Decimal) error {
	return c.Collateral.TakeProfit(amount)
}

func (c *Collateral) UpdateContracts(s order.Side, amount decimal.Decimal) error {
	switch {
	case c.currentDirection == nil:
		c.currentDirection = &s
		return c.Contract.AddContracts(amount)
	case *c.currentDirection == s:
		return c.Contract.AddContracts(amount)
	case *c.currentDirection != s:
		return c.Contract.ReduceContracts(amount)
	default:
		return errors.New("woah nelly")
	}
}

func (i *Item) TakeProfit(amount decimal.Decimal) error {
	if !i.asset.IsFutures() {
		return fmt.Errorf("%v %v %v %w", i.exchange, i.asset, i.currency, errNotFutures)
	}
	i.available = i.available.Add(amount)
	return nil
}

// AddContracts allocates an amount of funds to be used at a later time
// it prevents multiple events from claiming the same resource
func (i *Item) AddContracts(amount decimal.Decimal) error {
	if !i.asset.IsFutures() {
		return fmt.Errorf("%v %v %v %w", i.exchange, i.asset, i.currency, errNotFutures)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errZeroAmountReceived
	}
	i.available = i.available.Add(amount)
	return nil
}

// ReduceContracts allocates an amount of funds to be used at a later time
// it prevents multiple events from claiming the same resource
func (i *Item) ReduceContracts(amount decimal.Decimal) error {
	if !i.asset.IsFutures() {
		return fmt.Errorf("%v %v %v %w", i.exchange, i.asset, i.currency, errNotFutures)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errZeroAmountReceived
	}
	if amount.GreaterThan(i.available) {
		return fmt.Errorf("%w for %v %v %v. Requested %v Reserved: %v",
			errCannotAllocate,
			i.exchange,
			i.asset,
			i.currency,
			amount,
			i.reserved)
	}
	i.available = i.available.Sub(amount)
	return nil
}

func (c *Collateral) ReleaseContracts(amount decimal.Decimal) error {
	// turn this into a protected func
	c.Contract.available = c.Contract.available.Sub(amount)
	return nil
}

// FundReader
func (c *Collateral) FundReader() IFundReader {
	return c
}

// FundReserver
func (c *Collateral) FundReserver() IFundReserver {
	return c
}

// GetPairReleaser
func (c *Collateral) GetPairReleaser() (IPairReleaser, error) {
	return nil, fmt.Errorf("could not get pair releaser for %v %v %v %v %w", c.Contract.exchange, c.Collateral.asset, c.ContractCurrency(), c.CollateralCurrency(), ErrNotPair)
}

func (c *Collateral) Reserve(amount decimal.Decimal, side order.Side) error {
	switch side {
	case order.Long, order.Short:
		return c.Collateral.Reserve(amount)
	case common.ClosePosition:
		return c.Collateral.Release(amount, amount)
	default:
		return fmt.Errorf("%w for %v %v %v. Unknown side %v",
			errCannotAllocate,
			c.Collateral.exchange,
			c.Collateral.asset,
			c.Collateral.currency,
			side)
	}
}

// GetCollateralReleaser
func (c *Collateral) GetCollateralReleaser() (ICollateralReleaser, error) {
	return c, nil
}

// FundReleaser
func (c *Collateral) FundReleaser() IFundReleaser {
	return c
}

func (c *Collateral) Liquidate() {
	c.Collateral.available = decimal.Zero
	c.Contract.available = decimal.Zero
}

func (c *Collateral) CurrentHoldings() decimal.Decimal {
	return c.Contract.available
}
