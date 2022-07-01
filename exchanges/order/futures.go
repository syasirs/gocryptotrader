package order

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// SetupPositionController creates a position controller
// to track futures orders
func SetupPositionController() *PositionController {
	return &PositionController{
		multiPositionTrackers: make(map[string]map[asset.Item]map[currency.Pair]*MultiPositionTracker),
	}
}

// TrackNewOrder sets up the maps to then create a
// multi position tracker which funnels down into the
// position tracker, to then track an order's pnl
func (c *PositionController) TrackNewOrder(d *Detail) error {
	if d == nil {
		return errNilOrder
	}
	if !d.AssetType.IsFutures() {
		return fmt.Errorf("order %v %v %v %v %w",
			d.Exchange, d.AssetType, d.Pair, d.OrderID, ErrNotFuturesAsset)
	}
	if c == nil {
		return fmt.Errorf("position controller %w", common.ErrNilPointer)
	}
	c.m.Lock()
	defer c.m.Unlock()
	exchM, ok := c.multiPositionTrackers[strings.ToLower(d.Exchange)]
	if !ok {
		exchM = make(map[asset.Item]map[currency.Pair]*MultiPositionTracker)
		c.multiPositionTrackers[strings.ToLower(d.Exchange)] = exchM
	}
	itemM, ok := exchM[d.AssetType]
	if !ok {
		itemM = make(map[currency.Pair]*MultiPositionTracker)
		exchM[d.AssetType] = itemM
	}
	var err error
	multiPositionTracker, ok := itemM[d.Pair]
	if !ok {
		multiPositionTracker, err = SetupMultiPositionTracker(&MultiPositionTrackerSetup{
			Exchange:   strings.ToLower(d.Exchange),
			Asset:      d.AssetType,
			Pair:       d.Pair,
			Underlying: d.Pair.Base,
		})
		if err != nil {
			return err
		}
		itemM[d.Pair] = multiPositionTracker
	}
	return multiPositionTracker.TrackNewOrder(d)
}

// SetCollateralCurrency allows the setting of a collateral currency to all child trackers
// when using position controller for futures orders tracking
func (c *PositionController) SetCollateralCurrency(exch string, item asset.Item, pair currency.Pair, collateralCurrency currency.Code) error {
	if c == nil {
		return fmt.Errorf("position controller %w", common.ErrNilPointer)
	}
	c.m.Lock()
	defer c.m.Unlock()
	if !item.IsFutures() {
		return fmt.Errorf("%v %v %v %w", exch, item, pair, ErrNotFuturesAsset)
	}

	exchM, ok := c.multiPositionTrackers[strings.ToLower(exch)]
	if !ok {
		return fmt.Errorf("cannot set collateral %v for %v %v %v %w", collateralCurrency, exch, item, pair, ErrPositionsNotLoadedForExchange)
	}
	itemM, ok := exchM[item]
	if !ok {
		return fmt.Errorf("cannot set collateral %v for %v %v %v %w", collateralCurrency, exch, item, pair, ErrPositionsNotLoadedForAsset)
	}
	multiPositionTracker, ok := itemM[pair]
	if !ok {
		return fmt.Errorf("cannot set collateral %v for %v %v %v %w", collateralCurrency, exch, item, pair, ErrPositionsNotLoadedForPair)
	}
	if multiPositionTracker == nil {
		return fmt.Errorf("cannot set collateral %v for %v %v %v %w", collateralCurrency, exch, item, pair, common.ErrNilPointer)
	}
	multiPositionTracker.m.Lock()
	multiPositionTracker.collateralCurrency = collateralCurrency
	for i := range multiPositionTracker.positions {
		multiPositionTracker.positions[i].m.Lock()
		multiPositionTracker.positions[i].collateralCurrency = collateralCurrency
		multiPositionTracker.positions[i].m.Unlock()
	}
	multiPositionTracker.m.Unlock()
	return nil
}

// GetPositionsForExchange returns all positions for an
// exchange, asset pair that is stored in the position controller
func (c *PositionController) GetPositionsForExchange(exch string, item asset.Item, pair currency.Pair) ([]PositionStats, error) {
	if c == nil {
		return nil, fmt.Errorf("position controller %w", common.ErrNilPointer)
	}
	c.m.Lock()
	defer c.m.Unlock()
	if !item.IsFutures() {
		return nil, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrNotFuturesAsset)
	}
	exchM, ok := c.multiPositionTrackers[strings.ToLower(exch)]
	if !ok {
		return nil, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForExchange)
	}
	itemM, ok := exchM[item]
	if !ok {
		return nil, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForAsset)
	}
	multiPositionTracker, ok := itemM[pair]
	if !ok {
		return nil, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForPair)
	}

	return multiPositionTracker.GetPositions(), nil
}

// UpdateOpenPositionUnrealisedPNL finds an open position from
// an exchange asset pair, then calculates the unrealisedPNL
// using the latest ticker data
func (c *PositionController) UpdateOpenPositionUnrealisedPNL(exch string, item asset.Item, pair currency.Pair, last float64, updated time.Time) (decimal.Decimal, error) {
	if c == nil {
		return decimal.Zero, fmt.Errorf("position controller %w", common.ErrNilPointer)
	}
	if !item.IsFutures() {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrNotFuturesAsset)
	}

	c.m.Lock()
	defer c.m.Unlock()
	exchM, ok := c.multiPositionTrackers[strings.ToLower(exch)]
	if !ok {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForExchange)
	}
	itemM, ok := exchM[item]
	if !ok {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForAsset)
	}
	multiPositionTracker, ok := itemM[pair]
	if !ok {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForPair)
	}

	multiPositionTracker.m.Lock()
	defer multiPositionTracker.m.Unlock()
	pos := multiPositionTracker.positions
	if len(pos) == 0 {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForPair)
	}
	latestPos := pos[len(pos)-1]
	if latestPos.status != Open {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionClosed)
	}
	err := latestPos.TrackPNLByTime(updated, last)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w for position %v %v %v", err, exch, item, pair)
	}
	latestPos.m.Lock()
	defer latestPos.m.Unlock()
	return latestPos.unrealisedPNL, nil
}

// SetupMultiPositionTracker creates a futures order tracker for a specific exchange
func SetupMultiPositionTracker(setup *MultiPositionTrackerSetup) (*MultiPositionTracker, error) {
	if setup == nil {
		return nil, errNilSetup
	}
	if setup.Exchange == "" {
		return nil, errExchangeNameEmpty
	}
	if !setup.Asset.IsValid() || !setup.Asset.IsFutures() {
		return nil, ErrNotFuturesAsset
	}
	if setup.Pair.IsEmpty() {
		return nil, ErrPairIsEmpty
	}
	if setup.Underlying.IsEmpty() {
		return nil, errEmptyUnderlying
	}
	if setup.ExchangePNLCalculation == nil && setup.UseExchangePNLCalculation {
		return nil, errMissingPNLCalculationFunctions
	}
	return &MultiPositionTracker{
		exchange:                   strings.ToLower(setup.Exchange),
		asset:                      setup.Asset,
		pair:                       setup.Pair,
		underlying:                 setup.Underlying,
		offlinePNLCalculation:      setup.OfflineCalculation,
		orderPositions:             make(map[string]*PositionTracker),
		useExchangePNLCalculations: setup.UseExchangePNLCalculation,
		exchangePNLCalculation:     setup.ExchangePNLCalculation,
		collateralCurrency:         setup.CollateralCurrency,
	}, nil
}

// SetupPositionTracker creates a new position tracker to track n futures orders
// until the position(s) are closed
func (m *MultiPositionTracker) SetupPositionTracker(setup *PositionTrackerSetup) (*PositionTracker, error) {
	if m == nil {
		return nil, fmt.Errorf("multi-position tracker %w", common.ErrNilPointer)
	}
	if m.exchange == "" {
		return nil, errExchangeNameEmpty
	}
	if setup == nil {
		return nil, errNilSetup
	}
	if !setup.Asset.IsValid() || !setup.Asset.IsFutures() {
		return nil, ErrNotFuturesAsset
	}
	if setup.Pair.IsEmpty() {
		return nil, ErrPairIsEmpty
	}

	resp := &PositionTracker{
		exchange:                  strings.ToLower(m.exchange),
		asset:                     setup.Asset,
		contractPair:              setup.Pair,
		underlyingAsset:           setup.Underlying,
		status:                    Open,
		entryPrice:                setup.EntryPrice,
		currentDirection:          setup.Side,
		openingDirection:          setup.Side,
		useExchangePNLCalculation: setup.UseExchangePNLCalculation,
		collateralCurrency:        setup.CollateralCurrency,
		offlinePNLCalculation:     m.offlinePNLCalculation,
	}
	if !setup.UseExchangePNLCalculation {
		// use position tracker's pnl calculation by default
		resp.PNLCalculation = &PNLCalculator{}
	} else {
		if m.exchangePNLCalculation == nil {
			return nil, ErrNilPNLCalculator
		}
		resp.PNLCalculation = m.exchangePNLCalculation
	}
	return resp, nil
}

// UpdateOpenPositionUnrealisedPNL updates the pnl for the latest open position
// based on the last price and the time
func (m *MultiPositionTracker) UpdateOpenPositionUnrealisedPNL(last float64, updated time.Time) (decimal.Decimal, error) {
	m.m.Lock()
	defer m.m.Unlock()
	pos := m.positions
	if len(pos) == 0 {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", m.exchange, m.asset, m.pair, ErrPositionsNotLoadedForPair)
	}
	latestPos := pos[len(pos)-1]
	if latestPos.status.IsInactive() {
		return decimal.Zero, fmt.Errorf("%v %v %v %w", m.exchange, m.asset, m.pair, ErrPositionClosed)
	}
	err := latestPos.TrackPNLByTime(updated, last)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w for position %v %v %v", err, m.exchange, m.asset, m.pair)
	}
	latestPos.m.Lock()
	defer latestPos.m.Unlock()
	return latestPos.unrealisedPNL, nil
}

// ClearPositionsForExchange resets positions for an
// exchange, asset, pair that has been stored
func (c *PositionController) ClearPositionsForExchange(exch string, item asset.Item, pair currency.Pair) error {
	if c == nil {
		return fmt.Errorf("position controller %w", common.ErrNilPointer)
	}
	c.m.Lock()
	defer c.m.Unlock()
	if !item.IsFutures() {
		return fmt.Errorf("%v %v %v %w", exch, item, pair, ErrNotFuturesAsset)
	}
	exchM, ok := c.multiPositionTrackers[strings.ToLower(exch)]
	if !ok {
		return fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForExchange)
	}
	itemM, ok := exchM[item]
	if !ok {
		return fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForAsset)
	}
	multiPositionTracker, ok := itemM[pair]
	if !ok {
		return fmt.Errorf("%v %v %v %w", exch, item, pair, ErrPositionsNotLoadedForPair)
	}
	newMPT, err := SetupMultiPositionTracker(&MultiPositionTrackerSetup{
		Exchange:                  exch,
		Asset:                     item,
		Pair:                      pair,
		Underlying:                multiPositionTracker.underlying,
		OfflineCalculation:        multiPositionTracker.offlinePNLCalculation,
		UseExchangePNLCalculation: multiPositionTracker.useExchangePNLCalculations,
		ExchangePNLCalculation:    multiPositionTracker.exchangePNLCalculation,
		CollateralCurrency:        multiPositionTracker.collateralCurrency,
	})
	if err != nil {
		return err
	}
	itemM[pair] = newMPT
	return nil
}

// GetPositions returns all positions
func (m *MultiPositionTracker) GetPositions() []PositionStats {
	if m == nil {
		return nil
	}
	m.m.Lock()
	defer m.m.Unlock()
	resp := make([]PositionStats, len(m.positions))
	for i := range m.positions {
		resp[i] = m.positions[i].GetStats()
	}
	return resp
}

// TrackNewOrder upserts an order to the tracker and updates position
// status and exposure. PNL is calculated separately as it requires mark prices
func (m *MultiPositionTracker) TrackNewOrder(d *Detail) error {
	if m == nil {
		return fmt.Errorf("multi-position tracker %w", common.ErrNilPointer)
	}
	if d == nil {
		return ErrSubmissionIsNil
	}
	m.m.Lock()
	defer m.m.Unlock()
	if d.AssetType != m.asset {
		return errAssetMismatch
	}
	if tracker, ok := m.orderPositions[d.OrderID]; ok {
		// this has already been associated
		// update the tracker
		return tracker.TrackNewOrder(d, false)
	}
	if len(m.positions) > 0 {
		for i := range m.positions {
			if m.positions[i].status == Open && i != len(m.positions)-1 {
				return fmt.Errorf("%w %v at position %v/%v", errPositionDiscrepancy, m.positions[i], i, len(m.positions)-1)
			}
		}
		if m.positions[len(m.positions)-1].status == Open {
			err := m.positions[len(m.positions)-1].TrackNewOrder(d, false)
			if err != nil && !errors.Is(err, ErrPositionClosed) {
				return err
			}
			m.orderPositions[d.OrderID] = m.positions[len(m.positions)-1]
			return nil
		}
	}
	setup := &PositionTrackerSetup{
		Pair:                      d.Pair,
		EntryPrice:                decimal.NewFromFloat(d.Price),
		Underlying:                d.Pair.Base,
		Asset:                     d.AssetType,
		Side:                      d.Side,
		UseExchangePNLCalculation: m.useExchangePNLCalculations,
		CollateralCurrency:        m.collateralCurrency,
	}
	tracker, err := m.SetupPositionTracker(setup)
	if err != nil {
		return err
	}
	m.positions = append(m.positions, tracker)
	err = tracker.TrackNewOrder(d, true)
	if err != nil {
		return err
	}
	m.orderPositions[d.OrderID] = tracker
	return nil
}

// Liquidate will update the latest open position's
// to reflect its liquidated status
func (m *MultiPositionTracker) Liquidate(price decimal.Decimal, t time.Time) error {
	if m == nil {
		return fmt.Errorf("multi-position tracker %w", common.ErrNilPointer)
	}
	m.m.Lock()
	defer m.m.Unlock()
	if len(m.positions) == 0 {
		return fmt.Errorf("%v %v %v %w", m.exchange, m.asset, m.pair, ErrPositionsNotLoadedForPair)
	}
	return m.positions[len(m.positions)-1].Liquidate(price, t)
}

// GetStats returns a summary of a future position
func (p *PositionTracker) GetStats() PositionStats {
	if p == nil {
		return PositionStats{}
	}
	p.m.Lock()
	defer p.m.Unlock()
	var orders []Detail
	orders = append(orders, p.longPositions...)
	orders = append(orders, p.shortPositions...)

	return PositionStats{
		Exchange:           p.exchange,
		Asset:              p.asset,
		Pair:               p.contractPair,
		Underlying:         p.underlyingAsset,
		CollateralCurrency: p.collateralCurrency,
		Status:             p.status,
		Orders:             orders,
		RealisedPNL:        p.realisedPNL,
		UnrealisedPNL:      p.unrealisedPNL,
		LatestDirection:    p.currentDirection,
		OpeningDirection:   p.openingDirection,
		OpeningPrice:       p.entryPrice,
		LatestPrice:        p.latestPrice,
		PNLHistory:         p.pnlHistory,
		Exposure:           p.exposure,
	}
}

// TrackPNLByTime calculates the PNL based on a position tracker's exposure
// and current pricing. Adds the entry to PNL history to track over time
func (p *PositionTracker) TrackPNLByTime(t time.Time, currentPrice float64) error {
	if p == nil {
		return fmt.Errorf("position tracker %w", common.ErrNilPointer)
	}
	p.m.Lock()
	defer func() {
		p.latestPrice = decimal.NewFromFloat(currentPrice)
		p.m.Unlock()
	}()
	price := decimal.NewFromFloat(currentPrice)
	result := &PNLResult{
		Time:   t,
		Price:  price,
		Status: p.status,
	}
	if p.currentDirection.IsLong() {
		diff := price.Sub(p.entryPrice)
		result.UnrealisedPNL = p.exposure.Mul(diff)
	} else if p.currentDirection.IsShort() {
		diff := p.entryPrice.Sub(price)
		result.UnrealisedPNL = p.exposure.Mul(diff)
	}
	if len(p.pnlHistory) > 0 {
		latest := p.pnlHistory[len(p.pnlHistory)-1]
		result.RealisedPNLBeforeFees = latest.RealisedPNLBeforeFees
		result.Exposure = latest.Exposure
		result.Direction = latest.Direction
		result.RealisedPNL = latest.RealisedPNL
		result.IsLiquidated = latest.IsLiquidated
	}
	var err error
	p.pnlHistory, err = upsertPNLEntry(p.pnlHistory, result)
	p.unrealisedPNL = result.UnrealisedPNL
	return err
}

// GetRealisedPNL returns the realised pnl if the order
// is closed
func (p *PositionTracker) GetRealisedPNL() decimal.Decimal {
	if p == nil {
		return decimal.Zero
	}
	p.m.Lock()
	defer p.m.Unlock()
	return calculateRealisedPNL(p.pnlHistory)
}

// Liquidate will update the positions stats to reflect its liquidation
func (p *PositionTracker) Liquidate(price decimal.Decimal, t time.Time) error {
	if p == nil {
		return fmt.Errorf("position tracker %w", common.ErrNilPointer)
	}
	p.m.Lock()
	defer p.m.Unlock()
	latest, err := p.GetLatestPNLSnapshot()
	if err != nil {
		return err
	}
	if !latest.Time.Equal(t) {
		return fmt.Errorf("%w cannot liquidate from a different time. PNL snapshot %v. Liquidation request on %v Status: %v", errCannotLiquidate, latest.Time, t, p.status)
	}
	p.status = Liquidated
	p.currentDirection = ClosePosition
	p.exposure = decimal.Zero
	p.realisedPNL = decimal.Zero
	p.unrealisedPNL = decimal.Zero
	_, err = upsertPNLEntry(p.pnlHistory, &PNLResult{
		Time:         t,
		Price:        price,
		Direction:    ClosePosition,
		IsLiquidated: true,
		IsOrder:      true,
		Status:       p.status,
	})

	return err
}

// GetLatestPNLSnapshot takes the latest pnl history value
// and returns it
func (p *PositionTracker) GetLatestPNLSnapshot() (PNLResult, error) {
	if len(p.pnlHistory) == 0 {
		return PNLResult{}, fmt.Errorf("%v %v %v %w", p.exchange, p.asset, p.contractPair, errNoPNLHistory)
	}
	return p.pnlHistory[len(p.pnlHistory)-1], nil
}

// TrackNewOrder knows how things are going for a given
// futures contract
func (p *PositionTracker) TrackNewOrder(d *Detail, isInitialOrder bool) error {
	if p == nil {
		return fmt.Errorf("position tracker %w", common.ErrNilPointer)
	}
	p.m.Lock()
	defer p.m.Unlock()
	if isInitialOrder && len(p.pnlHistory) > 0 {
		return fmt.Errorf("%w received isInitialOrder = true with existing position", errCannotTrackInvalidParams)
	}
	if p.status.IsInactive() {
		return ErrPositionClosed
	}
	if d == nil {
		return ErrSubmissionIsNil
	}
	if !p.contractPair.Equal(d.Pair) {
		return fmt.Errorf("%w pair '%v' received: '%v'",
			errOrderNotEqualToTracker, d.Pair, p.contractPair)
	}
	if !strings.EqualFold(p.exchange, d.Exchange) {
		return fmt.Errorf("%w exchange '%v' received: '%v'",
			errOrderNotEqualToTracker, d.Exchange, p.exchange)
	}
	if p.asset != d.AssetType {
		return fmt.Errorf("%w asset '%v' received: '%v'",
			errOrderNotEqualToTracker, d.AssetType, p.asset)
	}

	if d.Side == UnknownSide {
		return ErrSideIsInvalid
	}
	if d.OrderID == "" {
		return ErrOrderIDNotSet
	}
	if d.Date.IsZero() {
		return fmt.Errorf("%w for %v %v %v order ID: %v unset",
			errTimeUnset, d.Exchange, d.AssetType, d.Pair, d.OrderID)
	}
	if len(p.shortPositions) == 0 && len(p.longPositions) == 0 {
		p.entryPrice = decimal.NewFromFloat(d.Price)
	}

	var updated bool
	for i := range p.shortPositions {
		if p.shortPositions[i].OrderID != d.OrderID {
			continue
		}
		ord := p.shortPositions[i].Copy()
		err := ord.UpdateOrderFromDetail(d)
		if err != nil {
			return err
		}
		p.shortPositions[i] = ord
		updated = true
		break
	}
	for i := range p.longPositions {
		if p.longPositions[i].OrderID != d.OrderID {
			continue
		}
		ord := p.longPositions[i].Copy()
		err := ord.UpdateOrderFromDetail(d)
		if err != nil {
			return err
		}
		p.longPositions[i] = ord
		updated = true
		break
	}

	if !updated {
		if d.Side.IsShort() {
			p.shortPositions = append(p.shortPositions, d.Copy())
		} else {
			p.longPositions = append(p.longPositions, d.Copy())
		}
	}
	var shortSide, longSide decimal.Decimal

	for i := range p.shortPositions {
		shortSide = shortSide.Add(decimal.NewFromFloat(p.shortPositions[i].Amount))
	}
	for i := range p.longPositions {
		longSide = longSide.Add(decimal.NewFromFloat(p.longPositions[i].Amount))
	}

	if isInitialOrder {
		p.openingDirection = d.Side
		p.currentDirection = d.Side
	}

	var result *PNLResult
	var err error
	var price, amount, leverage decimal.Decimal
	price = decimal.NewFromFloat(d.Price)
	amount = decimal.NewFromFloat(d.Amount)
	leverage = decimal.NewFromFloat(d.Leverage)
	cal := &PNLCalculatorRequest{
		Underlying:       p.underlyingAsset,
		Asset:            p.asset,
		OrderDirection:   d.Side,
		Leverage:         leverage,
		EntryPrice:       p.entryPrice,
		Amount:           amount,
		CurrentPrice:     price,
		Pair:             p.contractPair,
		Time:             d.Date,
		OpeningDirection: p.openingDirection,
		CurrentDirection: p.currentDirection,
		PNLHistory:       p.pnlHistory,
		Exposure:         p.exposure,
		Fee:              decimal.NewFromFloat(d.Fee),
		CalculateOffline: p.offlinePNLCalculation,
	}
	if len(p.pnlHistory) != 0 {
		cal.PreviousPrice = p.pnlHistory[len(p.pnlHistory)-1].Price
	}
	switch {
	case isInitialOrder:
		result = &PNLResult{
			IsOrder:       true,
			Time:          cal.Time,
			Price:         cal.CurrentPrice,
			Exposure:      cal.Amount,
			Fee:           cal.Fee,
			Direction:     cal.OpeningDirection,
			UnrealisedPNL: cal.Fee.Neg(),
		}
	case (cal.OrderDirection.IsShort() && cal.CurrentDirection.IsLong() || cal.OrderDirection.IsLong() && cal.CurrentDirection.IsShort()) && cal.Exposure.LessThan(amount):
		// latest order swaps directions!
		// split the order to calculate PNL from each direction
		first := cal.Exposure
		second := amount.Sub(cal.Exposure)
		baseFee := cal.Fee.Div(amount)
		cal.Fee = baseFee.Mul(first)
		cal.Amount = first
		result, err = p.PNLCalculation.CalculatePNL(context.TODO(), cal)
		if err != nil {
			return err
		}
		result.Status = p.status
		p.pnlHistory, err = upsertPNLEntry(cal.PNLHistory, result)
		if err != nil {
			return err
		}
		if cal.OrderDirection.IsLong() {
			cal.OrderDirection = Short
		} else if cal.OrderDirection.IsShort() {
			cal.OrderDirection = Long
		}
		if p.openingDirection.IsLong() {
			p.openingDirection = Short
		} else if p.openingDirection.IsShort() {
			p.openingDirection = Long
		}

		cal.Fee = baseFee.Mul(second)
		cal.Amount = second
		cal.EntryPrice = price
		cal.Time = cal.Time.Add(1)
		cal.PNLHistory = p.pnlHistory
		result, err = p.PNLCalculation.CalculatePNL(context.TODO(), cal)
	default:
		result, err = p.PNLCalculation.CalculatePNL(context.TODO(), cal)
	}
	if err != nil {
		if !errors.Is(err, ErrPositionLiquidated) {
			return err
		}
		result.UnrealisedPNL = decimal.Zero
		result.RealisedPNLBeforeFees = decimal.Zero
		p.status = Closed
	}
	result.Status = p.status
	p.pnlHistory, err = upsertPNLEntry(p.pnlHistory, result)
	if err != nil {
		return err
	}
	p.unrealisedPNL = result.UnrealisedPNL

	switch {
	case longSide.GreaterThan(shortSide):
		p.currentDirection = Long
	case shortSide.GreaterThan(longSide):
		p.currentDirection = Short
	default:
		p.currentDirection = ClosePosition
	}

	if p.currentDirection.IsLong() {
		p.exposure = longSide.Sub(shortSide)
	} else {
		p.exposure = shortSide.Sub(longSide)
	}

	if p.exposure.Equal(decimal.Zero) {
		p.status = Closed
		p.closingPrice = decimal.NewFromFloat(d.Price)
		p.realisedPNL = calculateRealisedPNL(p.pnlHistory)
		p.unrealisedPNL = decimal.Zero
		p.pnlHistory[len(p.pnlHistory)-1].RealisedPNL = p.realisedPNL
		p.pnlHistory[len(p.pnlHistory)-1].UnrealisedPNL = p.unrealisedPNL
		p.pnlHistory[len(p.pnlHistory)-1].Direction = p.currentDirection
	} else if p.exposure.IsNegative() {
		if p.currentDirection.IsLong() {
			p.currentDirection = Short
		} else {
			p.currentDirection = Long
		}
		p.exposure = p.exposure.Abs()
	}
	return nil
}

// GetCurrencyForRealisedPNL is a generic handling of determining the asset
// to assign realised PNL into, which is just itself
func (p *PNLCalculator) GetCurrencyForRealisedPNL(realisedAsset asset.Item, realisedPair currency.Pair) (currency.Code, asset.Item, error) {
	return realisedPair.Base, realisedAsset, nil
}

// CalculatePNL this is a localised generic way of calculating open
// positions' worth, it is an implementation of the PNLCalculation interface
func (p *PNLCalculator) CalculatePNL(_ context.Context, calc *PNLCalculatorRequest) (*PNLResult, error) {
	if calc == nil {
		return nil, ErrNilPNLCalculator
	}
	var previousPNL *PNLResult
	if len(calc.PNLHistory) > 0 {
		for i := len(calc.PNLHistory) - 1; i >= 0; i-- {
			if calc.PNLHistory[i].Time.Equal(calc.Time) || !calc.PNLHistory[i].IsOrder {
				continue
			}
			previousPNL = &calc.PNLHistory[i]
			break
		}
	}
	var prevExposure decimal.Decimal
	if previousPNL != nil {
		prevExposure = previousPNL.Exposure
	}
	var currentExposure, realisedPNL, unrealisedPNL, first, second decimal.Decimal
	if calc.OpeningDirection.IsLong() {
		first = calc.CurrentPrice
		if previousPNL != nil {
			second = previousPNL.Price
		}
	} else if calc.OpeningDirection.IsShort() {
		if previousPNL != nil {
			first = previousPNL.Price
		}
		second = calc.CurrentPrice
	}
	switch {
	case calc.OpeningDirection.IsShort() && calc.OrderDirection.IsShort(),
		calc.OpeningDirection.IsLong() && calc.OrderDirection.IsLong():
		// appending to your position
		currentExposure = prevExposure.Add(calc.Amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
	case calc.OpeningDirection.IsShort() && calc.OrderDirection.IsLong(),
		calc.OpeningDirection.IsLong() && calc.OrderDirection.IsShort():
		// selling/closing your position by "amount"
		currentExposure = prevExposure.Sub(calc.Amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
		realisedPNL = calc.Amount.Mul(first.Sub(second))
	default:
		return nil, fmt.Errorf("%w openinig direction: '%v' order direction: '%v' exposure: '%v'", errCannotCalculateUnrealisedPNL, calc.OpeningDirection, calc.OrderDirection, currentExposure)
	}
	totalFees := calc.Fee
	for i := range calc.PNLHistory {
		totalFees = totalFees.Add(calc.PNLHistory[i].Fee)
	}
	if !unrealisedPNL.IsZero() {
		unrealisedPNL = unrealisedPNL.Sub(totalFees)
	}

	response := &PNLResult{
		IsOrder:               true,
		Time:                  calc.Time,
		UnrealisedPNL:         unrealisedPNL,
		RealisedPNLBeforeFees: realisedPNL,
		Price:                 calc.CurrentPrice,
		Exposure:              currentExposure,
		Fee:                   calc.Fee,
		Direction:             calc.CurrentDirection,
	}

	return response, nil
}

// calculateRealisedPNL calculates the total realised PNL
// based on PNL history, minus fees
func calculateRealisedPNL(pnlHistory []PNLResult) decimal.Decimal {
	var realisedPNL, totalFees decimal.Decimal
	for i := range pnlHistory {
		if !pnlHistory[i].IsOrder {
			continue
		}
		realisedPNL = realisedPNL.Add(pnlHistory[i].RealisedPNLBeforeFees)
		totalFees = totalFees.Add(pnlHistory[i].Fee)
	}
	return realisedPNL.Sub(totalFees)
}

// upsertPNLEntry upserts an entry to PNLHistory field
// with some basic checks
func upsertPNLEntry(pnlHistory []PNLResult, entry *PNLResult) ([]PNLResult, error) {
	if entry.Time.IsZero() {
		return nil, errTimeUnset
	}
	for i := range pnlHistory {
		if !entry.Time.Equal(pnlHistory[i].Time) {
			continue
		}
		pnlHistory[i].UnrealisedPNL = entry.UnrealisedPNL
		pnlHistory[i].RealisedPNL = entry.RealisedPNL
		pnlHistory[i].RealisedPNLBeforeFees = entry.RealisedPNLBeforeFees
		pnlHistory[i].Exposure = entry.Exposure
		pnlHistory[i].Direction = entry.Direction
		pnlHistory[i].Price = entry.Price
		pnlHistory[i].Status = entry.Status
		pnlHistory[i].Fee = entry.Fee
		if entry.IsOrder {
			pnlHistory[i].IsOrder = true
		}
		if entry.IsLiquidated {
			pnlHistory[i].IsLiquidated = true
		}
		return pnlHistory, nil
	}
	pnlHistory = append(pnlHistory, *entry)
	sort.Slice(pnlHistory, func(i, j int) bool {
		return pnlHistory[i].Time.Before(pnlHistory[j].Time)
	})
	return pnlHistory, nil
}
