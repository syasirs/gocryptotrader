package portfolio

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/exchange"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	holdings2 "github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/risk"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/settings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/event"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/fill"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/order"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/signal"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding/holdings"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// Setup creates a portfolio manager instance and sets private fields
func Setup(sh SizeHandler, r risk.Handler, riskFreeRate decimal.Decimal) (*Portfolio, error) {
	if sh == nil {
		return nil, errSizeManagerUnset
	}
	if riskFreeRate.IsNegative() {
		return nil, errNegativeRiskFreeRate
	}
	if r == nil {
		return nil, errRiskManagerUnset
	}
	p := &Portfolio{}
	p.sizeManager = sh
	p.riskManager = r
	p.riskFreeRate = riskFreeRate

	return p, nil
}

// Reset returns the portfolio manager to its default state
func (p *Portfolio) Reset() {
	p.exchangeAssetPairSettings = nil
}

// OnSignal receives the event from the strategy on whether it has signalled to buy, do nothing or sell
// on buy/sell, the portfolio manager will size the order and assess the risk of the order
// if successful, it will pass on an order.Order to be used by the exchange event handler to place an order based on
// the portfolio manager's recommendations
func (p *Portfolio) OnSignal(s signal.Event, cs *exchange.Settings, funds funding.IPairReserver) (*order.Order, error) {
	if s == nil || cs == nil {
		return nil, common.ErrNilArguments
	}
	if p.sizeManager == nil {
		return nil, errSizeManagerUnset
	}
	if p.riskManager == nil {
		return nil, errRiskManagerUnset
	}

	o := &order.Order{
		Base: event.Base{
			Offset:       s.GetOffset(),
			Exchange:     s.GetExchange(),
			Time:         s.GetTime(),
			CurrencyPair: s.Pair(),
			AssetType:    s.GetAssetType(),
			Interval:     s.GetInterval(),
			Reason:       s.GetReason(),
		},
		Direction: s.GetDirection(),
	}
	if s.GetDirection() == "" {
		return o, errInvalidDirection
	}

	lookup := p.exchangeAssetPairSettings[s.GetExchange()][s.GetAssetType()][s.Pair()]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v",
			errNoPortfolioSettings,
			s.GetExchange(),
			s.GetAssetType(),
			s.Pair())
	}

	prevHolding := lookup.GetLatestHoldings()
	if p.iteration.IsZero() {
		prevHolding.InitialFunds = funds.QuoteInitialFunds()
		prevHolding.RemainingFunds = funds.QuoteInitialFunds()
		prevHolding.Exchange = s.GetExchange()
		prevHolding.Pair = s.Pair()
		prevHolding.Asset = s.GetAssetType()
		prevHolding.Timestamp = s.GetTime()
	}
	p.iteration = p.iteration.Add(decimal.NewFromInt(1))

	if s.GetDirection() == common.DoNothing || s.GetDirection() == common.MissingData || s.GetDirection() == "" {
		return o, nil
	}

	if !funds.CanPlaceOrder(s.GetDirection()) {
		if s.GetDirection() == gctorder.Sell {
			o.AppendReason("no holdings to sell")
			o.SetDirection(common.CouldNotSell)
		} else if s.GetDirection() == gctorder.Buy {
			o.AppendReason("not enough funds to buy")
			o.SetDirection(common.CouldNotBuy)
		}
		s.SetDirection(o.Direction)
		return o, nil
	}

	o.Price = s.GetPrice()
	o.OrderType = gctorder.Market
	o.BuyLimit = s.GetBuyLimit()
	o.SellLimit = s.GetSellLimit()
	var sizingFunds decimal.Decimal
	if s.GetDirection() == gctorder.Sell {
		sizingFunds = funds.BaseAvailable()
	} else {
		sizingFunds = funds.QuoteAvailable()
	}
	sizedOrder := p.sizeOrder(s, cs, o, sizingFunds, funds)

	return p.evaluateOrder(s, o, sizedOrder)
}

func (p *Portfolio) evaluateOrder(d common.Directioner, originalOrderSignal, sizedOrder *order.Order) (*order.Order, error) {
	var evaluatedOrder *order.Order
	cm, err := p.GetComplianceManager(originalOrderSignal.GetExchange(), originalOrderSignal.GetAssetType(), originalOrderSignal.Pair())
	if err != nil {
		return nil, err
	}

	evaluatedOrder, err = p.riskManager.EvaluateOrder(sizedOrder, p.GetLatestHoldingsForAllCurrencies(), cm.GetLatestSnapshot())
	if err != nil {
		originalOrderSignal.AppendReason(err.Error())
		switch d.GetDirection() {
		case gctorder.Buy:
			originalOrderSignal.Direction = common.CouldNotBuy
		case gctorder.Sell:
			originalOrderSignal.Direction = common.CouldNotSell
		case common.CouldNotBuy, common.CouldNotSell:
		default:
			originalOrderSignal.Direction = common.DoNothing
		}
		d.SetDirection(originalOrderSignal.Direction)
		return originalOrderSignal, nil
	}

	return evaluatedOrder, nil
}

func (p *Portfolio) sizeOrder(d common.Directioner, cs *exchange.Settings, originalOrderSignal *order.Order, sizingFunds decimal.Decimal, funds funding.IPairReserver) *order.Order {
	sizedOrder, err := p.sizeManager.SizeOrder(originalOrderSignal, sizingFunds, cs)
	if err != nil {
		originalOrderSignal.AppendReason(err.Error())
		switch originalOrderSignal.Direction {
		case gctorder.Buy:
			originalOrderSignal.Direction = common.CouldNotBuy
		case gctorder.Sell:
			originalOrderSignal.Direction = common.CouldNotSell
		default:
			originalOrderSignal.Direction = common.DoNothing
		}
		d.SetDirection(originalOrderSignal.Direction)
		return originalOrderSignal
	}

	if sizedOrder.Amount.IsZero() {
		switch originalOrderSignal.Direction {
		case gctorder.Buy:
			originalOrderSignal.Direction = common.CouldNotBuy
		case gctorder.Sell:
			originalOrderSignal.Direction = common.CouldNotSell
		default:
			originalOrderSignal.Direction = common.DoNothing
		}
		d.SetDirection(originalOrderSignal.Direction)
		originalOrderSignal.AppendReason("sized order to 0")
	}
	if d.GetDirection() == gctorder.Sell {
		err = funds.Reserve(sizedOrder.Amount, gctorder.Sell)
		sizedOrder.AllocatedFunds = sizedOrder.Amount
	} else {
		err = funds.Reserve(sizedOrder.Amount.Mul(sizedOrder.Price), gctorder.Buy)
		sizedOrder.AllocatedFunds = sizedOrder.Amount.Mul(sizedOrder.Price)
	}
	if err != nil {
		sizedOrder.Direction = common.DoNothing
		sizedOrder.AppendReason(err.Error())
	}
	return sizedOrder
}

// OnFill processes the event after an order has been placed by the exchange. Its purpose is to track holdings for future portfolio decisions.
func (p *Portfolio) OnFill(f fill.Event, funds funding.IPairReader) (*fill.Fill, error) {
	if f == nil {
		return nil, common.ErrNilEvent
	}
	lookup := p.exchangeAssetPairSettings[f.GetExchange()][f.GetAssetType()][f.Pair()]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v", errNoPortfolioSettings, f.GetExchange(), f.GetAssetType(), f.Pair())
	}
	var err error
	// Get the holding from the previous iteration, create it if it doesn't yet have a timestamp
	h := lookup.GetHoldingsForTime(f.GetTime().Add(-f.GetInterval().Duration()))
	if !h.Timestamp.IsZero() {
		h.Update(f)
	} else {
		h = lookup.GetLatestHoldings()
		if !h.Timestamp.IsZero() {
			h.Update(f)
		} else {
			h, err = holdings.CreatePairHolding(f, funds.QuoteInitialFunds(), p.riskFreeRate)
			if err != nil {
				return nil, err
			}
		}
	}
	err = p.setHoldingsForOffset(f.GetExchange(), f.GetAssetType(), f.Pair(), &h, true)
	if errors.Is(err, errNoHoldings) {
		err = p.setHoldingsForOffset(f.GetExchange(), f.GetAssetType(), f.Pair(), &h, false)
	}
	if err != nil {
		log.Error(log.BackTester, err)
	}

	err = p.addComplianceSnapshot(f)
	if err != nil {
		log.Error(log.BackTester, err)
	}

	direction := f.GetDirection()
	if direction == common.DoNothing ||
		direction == common.CouldNotBuy ||
		direction == common.CouldNotSell ||
		direction == common.MissingData ||
		direction == "" {
		fe := f.(*fill.Fill)
		fe.ExchangeFee = decimal.Zero
		return fe, nil
	}

	return f.(*fill.Fill), nil
}

// addComplianceSnapshot gets the previous snapshot of compliance events, updates with the latest fillevent
// then saves the snapshot to the c
func (p *Portfolio) addComplianceSnapshot(fillEvent fill.Event) error {
	if fillEvent == nil {
		return common.ErrNilEvent
	}
	complianceManager, err := p.GetComplianceManager(fillEvent.GetExchange(), fillEvent.GetAssetType(), fillEvent.Pair())
	if err != nil {
		return err
	}
	prevSnap := complianceManager.GetLatestSnapshot()
	fo := fillEvent.GetOrder()
	if fo != nil {
		price := decimal.NewFromFloat(fo.Price)
		amount := decimal.NewFromFloat(fo.Amount)
		fee := decimal.NewFromFloat(fo.Fee)
		snapOrder := compliance.SnapshotOrder{
			ClosePrice:          fillEvent.GetClosePrice(),
			VolumeAdjustedPrice: fillEvent.GetVolumeAdjustedPrice(),
			SlippageRate:        fillEvent.GetSlippageRate(),
			Detail:              fo,
			CostBasis:           price.Mul(amount).Add(fee),
		}
		prevSnap.Orders = append(prevSnap.Orders, snapOrder)
	}
	return complianceManager.AddSnapshot(prevSnap.Orders, fillEvent.GetTime(), fillEvent.GetOffset(), false)
}

// GetComplianceManager returns the order snapshots for a given exchange, asset, pair
func (p *Portfolio) GetComplianceManager(exchangeName string, a asset.Item, cp currency.Pair) (*compliance.Manager, error) {
	lookup := p.exchangeAssetPairSettings[exchangeName][a][cp]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v could not retrieve compliance manager", errNoPortfolioSettings, exchangeName, a, cp)
	}
	return &lookup.ComplianceManager, nil
}

// SetFee sets the fee rate
func (p *Portfolio) SetFee(exch string, a asset.Item, cp currency.Pair, fee decimal.Decimal) {
	lookup := p.exchangeAssetPairSettings[exch][a][cp]
	lookup.Fee = fee
}

// GetFee can panic for bad requests, but why are you getting things that don't exist?
func (p *Portfolio) GetFee(exchangeName string, a asset.Item, cp currency.Pair) decimal.Decimal {
	if p.exchangeAssetPairSettings == nil {
		return decimal.Zero
	}
	lookup := p.exchangeAssetPairSettings[exchangeName][a][cp]
	if lookup == nil {
		return decimal.Zero
	}
	return lookup.Fee
}

// IsInvested determines if there are any holdings for a given exchange, asset, pair
func (p *Portfolio) IsInvested(exchangeName string, a asset.Item, cp currency.Pair) (holdings2.Holding, bool) {
	s := p.exchangeAssetPairSettings[exchangeName][a][cp]
	if s == nil {
		return holdings2.Holding{}, false
	}
	h := s.GetLatestHoldings()
	if h.PositionsSize.GreaterThan(decimal.Zero) {
		return h, true
	}
	return h, false
}

// Update updates the portfolio holdings for the data event
func (p *Portfolio) Update(d common.DataEventHandler) error {
	if d == nil {
		return common.ErrNilEvent
	}
	h, ok := p.IsInvested(d.GetExchange(), d.GetAssetType(), d.Pair())
	if !ok {
		return nil
	}
	h.UpdateValue(d)
	err := p.setHoldingsForOffset(d.GetExchange(), d.GetAssetType(), d.Pair(), &h, true)
	if errors.Is(err, errNoHoldings) {
		err = p.setHoldingsForOffset(d.GetExchange(), d.GetAssetType(), d.Pair(), &h, false)
	}
	return err
}

// GetLatestHoldingsForAllCurrencies will return the current holdings for all loaded currencies
// this is useful to assess the position of your entire portfolio in order to help with risk decisions
func (p *Portfolio) GetLatestHoldingsForAllCurrencies() []holdings2.Holding {
	var resp []holdings2.Holding
	for _, x := range p.exchangeAssetPairSettings {
		for _, y := range x {
			for _, z := range y {
				holds := z.GetLatestHoldings()
				if holds.Offset != 0 {
					resp = append(resp, holds)
				}
			}
		}
	}
	return resp
}

func (p *Portfolio) setHoldingsForOffset(exch string, a asset.Item, cp currency.Pair, h *holdings2.Holding, overwriteExisting bool) error {
	if h.Timestamp.IsZero() {
		return errHoldingsNoTimestamp
	}
	lookup := p.exchangeAssetPairSettings[exch][a][cp]
	if lookup == nil {
		var err error
		lookup, err = p.SetupCurrencySettingsMap(exch, a, cp)
		if err != nil {
			return err
		}
	}
	if overwriteExisting && len(lookup.HoldingHolder) == 0 {
		return errNoHoldings
	}
	for i := len(lookup.HoldingHolder) - 1; i >= 0; i-- {
		if lookup.HoldingHolder[i].Offset == h.Offset {
			if overwriteExisting {
				lookup.HoldingHolder[i] = *h
				return nil
			}
			return errHoldingsAlreadySet
		}
	}
	if overwriteExisting {
		return fmt.Errorf("%w at %v", errNoHoldings, h.Timestamp)
	}

	lookup.HoldingHolder = append(lookup.HoldingHolder, *h)
	return nil
}

// ViewHoldingAtTimePeriod retrieves a snapshot of holdings at a specific time period,
// returning empty when not found
func (p *Portfolio) ViewHoldingAtTimePeriod(exch string, a asset.Item, cp currency.Pair, t time.Time) (holdings2.Holding, error) {
	exchangeAssetPairSettings := p.exchangeAssetPairSettings[exch][a][cp]
	if exchangeAssetPairSettings == nil {
		return holdings2.Holding{}, fmt.Errorf("%w for %v %v %v", errNoHoldings, exch, a, cp)
	}

	for i := len(exchangeAssetPairSettings.HoldingHolder) - 1; i >= 0; i-- {
		if t.Equal(exchangeAssetPairSettings.HoldingHolder[i].Timestamp) {
			return exchangeAssetPairSettings.HoldingHolder[i], nil
		}
	}

	return holdings2.Holding{}, fmt.Errorf("%w for %v %v %v at %v", errNoHoldings, exch, a, cp, t)
}

// SetupCurrencySettingsMap ensures a map is created and no panics happen
func (p *Portfolio) SetupCurrencySettingsMap(exch string, a asset.Item, cp currency.Pair) (*settings.Settings, error) {
	if exch == "" {
		return nil, errExchangeUnset
	}
	if a == "" {
		return nil, errAssetUnset
	}
	if cp.IsEmpty() {
		return nil, errCurrencyPairUnset
	}
	if p.exchangeAssetPairSettings == nil {
		p.exchangeAssetPairSettings = make(map[string]map[asset.Item]map[currency.Pair]*settings.Settings)
	}
	if p.exchangeAssetPairSettings[exch] == nil {
		p.exchangeAssetPairSettings[exch] = make(map[asset.Item]map[currency.Pair]*settings.Settings)
	}
	if p.exchangeAssetPairSettings[exch][a] == nil {
		p.exchangeAssetPairSettings[exch][a] = make(map[currency.Pair]*settings.Settings)
	}
	if _, ok := p.exchangeAssetPairSettings[exch][a][cp]; !ok {
		p.exchangeAssetPairSettings[exch][a][cp] = &settings.Settings{}
	}

	return p.exchangeAssetPairSettings[exch][a][cp], nil
}
