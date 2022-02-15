package portfolio

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/exchange"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/event"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/fill"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/order"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/signal"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// OnSignal receives the event from the strategy on whether it has signalled to buy, do nothing or sell
// on buy/sell, the portfolio manager will size the order and assess the risk of the order
// if successful, it will pass on an order.Order to be used by the exchange event handler to place an order based on
// the portfolio manager's recommendations
func (p *Portfolio) OnSignal(ev signal.Event, cs *exchange.Settings, funds funding.IFundReserver) (*order.Order, error) {
	if ev == nil || cs == nil {
		return nil, common.ErrNilArguments
	}
	if p.sizeManager == nil {
		return nil, errSizeManagerUnset
	}
	if p.riskManager == nil {
		return nil, errRiskManagerUnset
	}
	if funds == nil {
		return nil, funding.ErrFundsNotFound
	}

	o := &order.Order{
		Base: event.Base{
			Offset:       ev.GetOffset(),
			Exchange:     ev.GetExchange(),
			Time:         ev.GetTime(),
			CurrencyPair: ev.Pair(),
			AssetType:    ev.GetAssetType(),
			Interval:     ev.GetInterval(),
			Reason:       ev.GetReason(),
		},
		Direction:          ev.GetDirection(),
		FillDependentEvent: ev.GetFillDependentEvent(),
	}
	if ev.GetDirection() == "" {
		return o, errInvalidDirection
	}

	lookup := p.exchangeAssetPairSettings[ev.GetExchange()][ev.GetAssetType()][ev.Pair()]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v",
			errNoPortfolioSettings,
			ev.GetExchange(),
			ev.GetAssetType(),
			ev.Pair())
	}

	if ev.GetDirection() == common.DoNothing ||
		ev.GetDirection() == common.MissingData ||
		ev.GetDirection() == common.TransferredFunds ||
		ev.GetDirection() == "" {
		return o, nil
	}
	dir := ev.GetDirection()
	if !funds.CanPlaceOrder(dir) {
		return cannotPurchase(ev, o, dir)
	}

	o.Price = ev.GetPrice()
	o.OrderType = gctorder.Market
	o.BuyLimit = ev.GetBuyLimit()
	o.SellLimit = ev.GetSellLimit()
	var sizingFunds decimal.Decimal
	if ev.GetAssetType() == asset.Spot {
		pReader, err := funds.GetPairReader()
		if err != nil {
			return nil, err
		}
		if ev.GetDirection() == gctorder.Sell {
			sizingFunds = pReader.BaseAvailable()
		} else {
			sizingFunds = pReader.QuoteAvailable()
		}
	} else if ev.GetAssetType().IsFutures() {
		collateralFunds, err := funds.GetCollateralReader()
		if err != nil {
			return nil, err
		}

		sizingFunds = collateralFunds.AvailableFunds()
	}
	if sizingFunds.LessThanOrEqual(decimal.Zero) {
		return cannotPurchase(ev, o, dir)
	}
	sizedOrder := p.sizeOrder(ev, cs, o, sizingFunds, funds)

	return p.evaluateOrder(ev, o, sizedOrder)
}

func cannotPurchase(ev signal.Event, o *order.Order, dir gctorder.Side) (*order.Order, error) {
	o.AppendReason(notEnoughFundsTo + " " + dir.Lower())
	switch ev.GetDirection() {
	case gctorder.Sell:
		o.SetDirection(common.CouldNotSell)
	case gctorder.Buy:
		o.SetDirection(common.CouldNotBuy)
	case gctorder.Short:
		o.SetDirection(common.CouldNotShort)
	case gctorder.Long:
		o.SetDirection(common.CouldNotLong)
	}
	ev.SetDirection(o.Direction)
	return o, nil
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
		case gctorder.Short:
			originalOrderSignal.Direction = common.CouldNotShort
		case gctorder.Long:
			originalOrderSignal.Direction = common.CouldNotLong
		default:
			originalOrderSignal.Direction = common.DoNothing
		}
		d.SetDirection(originalOrderSignal.Direction)
		return originalOrderSignal, nil
	}

	return evaluatedOrder, nil
}

func (p *Portfolio) sizeOrder(d common.Directioner, cs *exchange.Settings, originalOrderSignal *order.Order, sizingFunds decimal.Decimal, funds funding.IFundReserver) *order.Order {
	sizedOrder, err := p.sizeManager.SizeOrder(originalOrderSignal, sizingFunds, cs)
	if err != nil {
		originalOrderSignal.AppendReason(err.Error())
		switch originalOrderSignal.Direction {
		case gctorder.Buy:
			originalOrderSignal.Direction = common.CouldNotBuy
		case gctorder.Sell:
			originalOrderSignal.Direction = common.CouldNotSell
		case gctorder.Long:
			originalOrderSignal.Direction = common.CouldNotLong
		case gctorder.Short:
			originalOrderSignal.Direction = common.CouldNotShort
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
		case gctorder.Long:
			originalOrderSignal.Direction = common.CouldNotLong
		case gctorder.Short:
			originalOrderSignal.Direction = common.CouldNotShort
		default:
			originalOrderSignal.Direction = common.DoNothing
		}
		d.SetDirection(originalOrderSignal.Direction)
		originalOrderSignal.AppendReason("sized order to 0")
	}
	switch d.GetDirection() {
	case gctorder.Sell:
		err = funds.Reserve(sizedOrder.Amount, gctorder.Sell)
		sizedOrder.AllocatedFunds = sizedOrder.Amount
	case gctorder.Short, gctorder.Long:
		err = funds.Reserve(sizedOrder.Amount, d.GetDirection())
		sizedOrder.AllocatedFunds = sizedOrder.Amount.Div(sizedOrder.Price)
	default:
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
func (p *Portfolio) OnFill(ev fill.Event, funding funding.IFundReleaser) (fill.Event, error) {
	if ev == nil {
		return nil, common.ErrNilEvent
	}
	lookup := p.exchangeAssetPairSettings[ev.GetExchange()][ev.GetAssetType()][ev.Pair()]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v", errNoPortfolioSettings, ev.GetExchange(), ev.GetAssetType(), ev.Pair())
	}
	var err error

	if ev.GetAssetType() == asset.Spot {
		fp, err := funding.GetPairReleaser()
		if err != nil {
			return nil, err
		}
		// Get the holding from the previous iteration, create it if it doesn't yet have a timestamp
		h := lookup.GetHoldingsForTime(ev.GetTime().Add(-ev.GetInterval().Duration()))
		if !h.Timestamp.IsZero() {
			h.Update(ev, fp)
		} else {
			h = lookup.GetLatestHoldings()
			if h.Timestamp.IsZero() {
				h, err = holdings.Create(ev, funding)
				if err != nil {
					return nil, err
				}
			} else {
				h.Update(ev, fp)
			}
		}
		err = p.setHoldingsForOffset(&h, true)
		if errors.Is(err, errNoHoldings) {
			err = p.setHoldingsForOffset(&h, false)
		}
		if err != nil {
			log.Error(log.BackTester, err)
		}
	}

	err = p.addComplianceSnapshot(ev)
	if err != nil {
		log.Error(log.BackTester, err)
	}
	ev.SetExchangeFee(decimal.Zero)

	return ev, nil
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
	if fo := fillEvent.GetOrder(); fo != nil {
		price := decimal.NewFromFloat(fo.Price)
		amount := decimal.NewFromFloat(fo.Amount)
		fee := decimal.NewFromFloat(fo.Fee)
		snapOrder := compliance.SnapshotOrder{
			ClosePrice:          fillEvent.GetClosePrice(),
			VolumeAdjustedPrice: fillEvent.GetVolumeAdjustedPrice(),
			SlippageRate:        fillEvent.GetSlippageRate(),
			CostBasis:           price.Mul(amount).Add(fee),
		}
		snapOrder.Order = fo
		prevSnap.Orders = append(prevSnap.Orders, snapOrder)
	}
	snap := &compliance.Snapshot{
		Offset:    fillEvent.GetOffset(),
		Timestamp: fillEvent.GetTime(),
		Orders:    prevSnap.Orders,
	}
	return complianceManager.AddSnapshot(snap, false)
}

func (p *Portfolio) setHoldingsForOffset(h *holdings.Holding, overwriteExisting bool) error {
	if h.Timestamp.IsZero() {
		return errHoldingsNoTimestamp
	}
	lookup, ok := p.exchangeAssetPairSettings[h.Exchange][h.Asset][h.Pair]
	if !ok {
		return fmt.Errorf("%w for %v %v %v", errNoPortfolioSettings, h.Exchange, h.Asset, h.Pair)
	}

	if overwriteExisting && len(lookup.HoldingsSnapshots) == 0 {
		return errNoHoldings
	}
	for i := len(lookup.HoldingsSnapshots) - 1; i >= 0; i-- {
		if lookup.HoldingsSnapshots[i].Offset == h.Offset {
			if overwriteExisting {
				lookup.HoldingsSnapshots[i] = *h
				return nil
			}
			return errHoldingsAlreadySet
		}
	}
	if overwriteExisting {
		return fmt.Errorf("%w at %v", errNoHoldings, h.Timestamp)
	}

	lookup.HoldingsSnapshots = append(lookup.HoldingsSnapshots, *h)
	return nil
}

// GetLatestOrderSnapshotForEvent gets orders related to the event
func (p *Portfolio) GetLatestOrderSnapshotForEvent(e common.EventHandler) (compliance.Snapshot, error) {
	eapSettings, ok := p.exchangeAssetPairSettings[e.GetExchange()][e.GetAssetType()][e.Pair()]
	if !ok {
		return compliance.Snapshot{}, fmt.Errorf("%w for %v %v %v", errNoPortfolioSettings, e.GetExchange(), e.GetAssetType(), e.Pair())
	}
	return eapSettings.ComplianceManager.GetLatestSnapshot(), nil
}

// GetLatestOrderSnapshots returns the latest snapshots from all stored pair data
func (p *Portfolio) GetLatestOrderSnapshots() ([]compliance.Snapshot, error) {
	var resp []compliance.Snapshot
	for _, exchangeMap := range p.exchangeAssetPairSettings {
		for _, assetMap := range exchangeMap {
			for _, pairMap := range assetMap {
				resp = append(resp, pairMap.ComplianceManager.GetLatestSnapshot())
			}
		}
	}
	if len(resp) == 0 {
		return nil, errNoPortfolioSettings
	}
	return resp, nil
}

// GetComplianceManager returns the order snapshots for a given exchange, asset, pair
func (p *Portfolio) GetComplianceManager(exchangeName string, a asset.Item, cp currency.Pair) (*compliance.Manager, error) {
	lookup := p.exchangeAssetPairSettings[exchangeName][a][cp]
	if lookup == nil {
		return nil, fmt.Errorf("%w for %v %v %v could not retrieve compliance manager", errNoPortfolioSettings, exchangeName, a, cp)
	}
	return &lookup.ComplianceManager, nil
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

// UpdateHoldings updates the portfolio holdings for the data event
func (p *Portfolio) UpdateHoldings(e common.DataEventHandler, funds funding.IFundReleaser) error {
	if e == nil {
		return common.ErrNilEvent
	}
	if funds == nil {
		return funding.ErrFundsNotFound
	}
	settings, err := p.getSettings(e.GetExchange(), e.GetAssetType(), e.Pair())
	if err != nil {
		return fmt.Errorf("%v %v %v %w", e.GetExchange(), e.GetAssetType(), e.Pair(), err)
	}
	h := settings.GetLatestHoldings()
	if h.Timestamp.IsZero() {
		h, err = holdings.Create(e, funds)
		if err != nil {
			return err
		}
	}
	h.UpdateValue(e)
	err = p.setHoldingsForOffset(&h, true)
	if errors.Is(err, errNoHoldings) {
		err = p.setHoldingsForOffset(&h, false)
	}
	return err
}

// GetLatestHoldingsForAllCurrencies will return the current holdings for all loaded currencies
// this is useful to assess the position of your entire portfolio in order to help with risk decisions
func (p *Portfolio) GetLatestHoldingsForAllCurrencies() []holdings.Holding {
	var resp []holdings.Holding
	for _, x := range p.exchangeAssetPairSettings {
		for _, y := range x {
			for _, z := range y {
				holds := z.GetLatestHoldings()
				if !holds.Timestamp.IsZero() {
					resp = append(resp, holds)
				}
			}
		}
	}
	return resp
}

// ViewHoldingAtTimePeriod retrieves a snapshot of holdings at a specific time period,
// returning empty when not found
func (p *Portfolio) ViewHoldingAtTimePeriod(ev common.EventHandler) (*holdings.Holding, error) {
	exchangeAssetPairSettings := p.exchangeAssetPairSettings[ev.GetExchange()][ev.GetAssetType()][ev.Pair()]
	if exchangeAssetPairSettings == nil {
		return nil, fmt.Errorf("%w for %v %v %v", errNoHoldings, ev.GetExchange(), ev.GetAssetType(), ev.Pair())
	}

	for i := len(exchangeAssetPairSettings.HoldingsSnapshots) - 1; i >= 0; i-- {
		if ev.GetTime().Equal(exchangeAssetPairSettings.HoldingsSnapshots[i].Timestamp) {
			return &exchangeAssetPairSettings.HoldingsSnapshots[i], nil
		}
	}

	return nil, fmt.Errorf("%w for %v %v %v at %v", errNoHoldings, ev.GetExchange(), ev.GetAssetType(), ev.Pair(), ev.GetTime())
}

// GetLatestHoldings returns the latest holdings after being sorted by time
func (e *Settings) GetLatestHoldings() holdings.Holding {
	if len(e.HoldingsSnapshots) == 0 {
		return holdings.Holding{}
	}

	return e.HoldingsSnapshots[len(e.HoldingsSnapshots)-1]
}

// GetHoldingsForTime returns the holdings for a time period, or an empty holding if not found
func (e *Settings) GetHoldingsForTime(t time.Time) holdings.Holding {
	if e.HoldingsSnapshots == nil {
		// no holdings yet
		return holdings.Holding{}
	}
	for i := len(e.HoldingsSnapshots) - 1; i >= 0; i-- {
		if e.HoldingsSnapshots[i].Timestamp.Equal(t) {
			return e.HoldingsSnapshots[i]
		}
	}
	return holdings.Holding{}
}

// GetPositions returns all futures positions for an event's exchange, asset, pair
func (p *Portfolio) GetPositions(e common.EventHandler) ([]gctorder.PositionStats, error) {
	if !e.GetAssetType().IsFutures() {
		return nil, errors.New("not a future")
	}
	settings, err := p.getSettings(e.GetExchange(), e.GetAssetType(), e.Pair())
	if err != nil {
		return nil, fmt.Errorf("%v %v %v %w", e.GetExchange(), e.GetAssetType(), e.Pair(), err)
	}
	if settings.FuturesTracker == nil {
		return nil, errors.New("no futures tracker")
	}
	return settings.FuturesTracker.GetPositions(), nil
}

// UpdatePNL will analyse any futures orders that have been placed over the backtesting run
// that are not closed and calculate their PNL
func (p *Portfolio) UpdatePNL(e common.EventHandler, closePrice decimal.Decimal) error {
	if !e.GetAssetType().IsFutures() {
		return fmt.Errorf("%s %w", e.GetAssetType(), gctorder.ErrNotFutureAsset)
	}
	settings, err := p.getSettings(e.GetExchange(), e.GetAssetType(), e.Pair())
	if err != nil {
		return fmt.Errorf("%v %v %v %w", e.GetExchange(), e.GetAssetType(), e.Pair(), err)
	}

	_, err = settings.FuturesTracker.UpdateOpenPositionUnrealisedPNL(closePrice.InexactFloat64(), e.GetTime())
	if err != nil && !errors.Is(err, gctorder.ErrPositionClosed) {
		return err
	}

	return nil
}

// TrackFuturesOrder updates the futures tracker with a new order
// from a fill event
func (p *Portfolio) TrackFuturesOrder(detail *gctorder.Detail, fund funding.IFundReleaser) error {
	if detail == nil {
		return gctorder.ErrSubmissionIsNil
	}
	if !detail.AssetType.IsFutures() {
		return fmt.Errorf("order '%v' %w", detail.ID, gctorder.ErrNotFuturesAsset)
	}
	collateralReleaser, err := fund.GetCollateralReleaser()
	if err != nil {
		return fmt.Errorf("%v %v %v %w", detail.Exchange, detail.AssetType, detail.Pair, err)
	}
	settings, err := p.getSettings(detail.Exchange, detail.AssetType, detail.Pair)
	if err != nil {
		return fmt.Errorf("%v %v %v %w", detail.Exchange, detail.AssetType, detail.Pair, err)
	}

	err = settings.FuturesTracker.TrackNewOrder(detail)
	if err != nil {
		return err
	}

	pos := settings.FuturesTracker.GetPositions()
	if len(pos) == 0 {
		return fmt.Errorf("%w should not happen", errNoHoldings)
	}
	if pos[len(pos)-1].Status == gctorder.Closed {
		amount := decimal.NewFromFloat(detail.Amount)
		err = collateralReleaser.TakeProfit(amount, pos[len(pos)-1].EntryAmount, pos[len(pos)-1].RealisedPNL)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLatestPNLForEvent takes in an event and returns the latest PNL data
// if it exists
func (p *Portfolio) GetLatestPNLForEvent(e common.EventHandler) (*PNLSummary, error) {
	if !e.GetAssetType().IsFutures() {
		return nil, errors.New("not a future")
	}
	settings, err := p.getSettings(e.GetExchange(), e.GetAssetType(), e.Pair())
	if err != nil {
		return nil, fmt.Errorf("%v %v %v %w", e.GetExchange(), e.GetAssetType(), e.Pair(), err)
	}

	if settings.FuturesTracker == nil {
		return nil, errors.New("no futures tracker")
	}

	response := &PNLSummary{
		Exchange: e.GetExchange(),
		Item:     e.GetAssetType(),
		Pair:     e.Pair(),
		Offset:   e.GetOffset(),
	}
	positions := settings.FuturesTracker.GetPositions()
	if len(positions) == 0 {
		return response, nil
	}
	pnlHistory := positions[len(positions)-1].PNLHistory
	if len(pnlHistory) == 0 {
		return response, nil
	}
	response.PNL = pnlHistory[len(pnlHistory)-1]
	return response, nil
}

func (p *Portfolio) getSettings(exch string, item asset.Item, pair currency.Pair) (*Settings, error) {
	exchMap, ok := p.exchangeAssetPairSettings[strings.ToLower(exch)]
	if !ok {
		return nil, errExchangeUnset
	}
	itemMap, ok := exchMap[item]
	if !ok {
		return nil, errAssetUnset
	}
	pairMap, ok := itemMap[pair]
	if !ok {
		return nil, errCurrencyPairUnset
	}

	return pairMap, nil
}

// GetLatestPNLs returns all PNL details in one array
func (p *Portfolio) GetLatestPNLs() []PNLSummary {
	var result []PNLSummary
	for exchK := range p.exchangeAssetPairSettings {
		for assetK := range p.exchangeAssetPairSettings[exchK] {
			if !assetK.IsFutures() {
				continue
			}
			for pairK, settings := range p.exchangeAssetPairSettings[exchK][assetK] {
				if settings == nil {
					continue
				}
				if settings.FuturesTracker == nil {
					continue
				}
				summary := PNLSummary{
					Exchange: exchK,
					Item:     assetK,
					Pair:     pairK,
				}
				positions := settings.FuturesTracker.GetPositions()
				if len(positions) > 0 {
					pnlHistory := positions[len(positions)-1].PNLHistory
					if len(pnlHistory) > 0 {
						summary.PNL = pnlHistory[len(pnlHistory)-1]
					}
				}

				result = append(result, summary)
			}
		}
	}
	return result
}
