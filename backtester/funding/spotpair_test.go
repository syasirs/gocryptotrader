package funding

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

func TestBaseInitialFunds(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	funds := pairItems.BaseInitialFunds()
	if !funds.IsZero() {
		t.Errorf("received '%v' expected '%v'", funds, baseItem.available)
	}
}

func TestQuoteInitialFunds(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	funds := pairItems.QuoteInitialFunds()
	if !funds.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", funds, elite)
	}
}

func TestBaseAvailable(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	funds := pairItems.BaseAvailable()
	if !funds.IsZero() {
		t.Errorf("received '%v' expected '%v'", funds, baseItem.available)
	}
}

func TestQuoteAvailable(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	funds := pairItems.QuoteAvailable()
	if !funds.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", funds, elite)
	}
}

func TestReservePair(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	err = pairItems.Reserve(decimal.Zero, gctorder.Buy)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
	err = pairItems.Reserve(elite, gctorder.Buy)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	err = pairItems.Reserve(decimal.Zero, gctorder.Sell)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
	err = pairItems.Reserve(elite, gctorder.Sell)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}
	err = pairItems.Reserve(elite, common.DoNothing)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}
}

func TestReleasePair(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	err = pairItems.Reserve(decimal.Zero, gctorder.Buy)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
	err = pairItems.Reserve(elite, gctorder.Buy)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	err = pairItems.Reserve(decimal.Zero, gctorder.Sell)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
	err = pairItems.Reserve(elite, gctorder.Sell)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}

	err = pairItems.Release(decimal.Zero, decimal.Zero, gctorder.Buy)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
	err = pairItems.Release(elite, decimal.Zero, gctorder.Buy)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	err = pairItems.Release(elite, decimal.Zero, gctorder.Buy)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}

	err = pairItems.Release(elite, decimal.Zero, common.DoNothing)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}

	err = pairItems.Release(elite, decimal.Zero, gctorder.Sell)
	if !errors.Is(err, errCannotAllocate) {
		t.Errorf("received '%v' expected '%v'", err, errCannotAllocate)
	}
	err = pairItems.Release(decimal.Zero, decimal.Zero, gctorder.Sell)
	if !errors.Is(err, errZeroAmountReceived) {
		t.Errorf("received '%v' expected '%v'", err, errZeroAmountReceived)
	}
}

func TestIncreaseAvailablePair(t *testing.T) {
	t.Parallel()
	baseItem, err := CreateItem(exchName, a, pair.Base, decimal.Zero, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	quoteItem, err := CreateItem(exchName, a, pair.Quote, elite, decimal.Zero)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	baseItem.pairedWith = quoteItem
	quoteItem.pairedWith = baseItem
	pairItems := SpotPair{base: baseItem, quote: quoteItem}
	pairItems.IncreaseAvailable(decimal.Zero, gctorder.Buy)
	if !pairItems.quote.available.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", elite, pairItems.quote.available)
	}
	pairItems.IncreaseAvailable(decimal.Zero, gctorder.Sell)
	if !pairItems.base.available.IsZero() {
		t.Errorf("received '%v' expected '%v'", decimal.Zero, pairItems.base.available)
	}

	pairItems.IncreaseAvailable(elite.Neg(), gctorder.Sell)
	if !pairItems.quote.available.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", elite, pairItems.quote.available)
	}
	pairItems.IncreaseAvailable(elite, gctorder.Buy)
	if !pairItems.base.available.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", elite, pairItems.base.available)
	}

	pairItems.IncreaseAvailable(elite, common.DoNothing)
	if !pairItems.base.available.Equal(elite) {
		t.Errorf("received '%v' expected '%v'", elite, pairItems.base.available)
	}
}

func TestCanPlaceOrderPair(t *testing.T) {
	t.Parallel()
	p := SpotPair{
		base:  &Item{},
		quote: &Item{},
	}
	if p.CanPlaceOrder(common.DoNothing) {
		t.Error("expected false")
	}
	if p.CanPlaceOrder(gctorder.Buy) {
		t.Error("expected false")
	}
	if p.CanPlaceOrder(gctorder.Sell) {
		t.Error("expected false")
	}

	p.quote.available = decimal.NewFromInt(32)
	if !p.CanPlaceOrder(gctorder.Buy) {
		t.Error("expected true")
	}
	p.base.available = decimal.NewFromInt(32)
	if !p.CanPlaceOrder(gctorder.Sell) {
		t.Error("expected true")
	}
}

func TestGetPairReader(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	var expectedError error
	ip, err := p.GetPairReader()
	if !errors.Is(err, expectedError) {
		t.Errorf("recevied '%v' expected '%v'", err, expectedError)
	}
	if ip != p {
		t.Error("expected the same thing")
	}
}

func TestGetCollateralReader(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	expectedError := ErrNotCollateral
	_, err := p.GetCollateralReader()
	if !errors.Is(err, expectedError) {
		t.Errorf("recevied '%v' expected '%v'", err, expectedError)
	}
}

func TestFundReader(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	if p.FundReader() != p {
		t.Error("expected the same thing")
	}
}

func TestFundReserver(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	if p.FundReserver() != p {
		t.Error("expected the same thing")
	}
}

func TestFundReleaser(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	if p.FundReleaser() != p {
		t.Error("expected the same thing")
	}
}

func TestPairReleaser(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	if _, err := p.PairReleaser(); !errors.Is(err, nil) {
		t.Errorf("recevied '%v' expected '%v'", err, nil)
	}
}

func TestCollateralReleaser(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{exchange: "hello"},
	}
	if _, err := p.CollateralReleaser(); !errors.Is(err, ErrNotCollateral) {
		t.Errorf("recevied '%v' expected '%v'", err, ErrNotCollateral)
	}
}

func TestLiquidate(t *testing.T) {
	t.Parallel()
	p := &SpotPair{
		base: &Item{
			available: decimal.NewFromInt(1337),
		},
		quote: &Item{
			available: decimal.NewFromInt(1337),
		},
	}
	p.Liquidate()
	if !p.base.available.IsZero() {
		t.Errorf("received '%v' expected '%v'", p.base.available, "0")
	}
	if !p.quote.available.IsZero() {
		t.Errorf("received '%v' expected '%v'", p.quote.available, "0")
	}
}
