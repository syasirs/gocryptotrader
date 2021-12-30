package order

import (
	"errors"
	"fmt"
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
		positionTrackerControllers: make(map[string]map[asset.Item]map[currency.Pair]*MultiPositionTracker),
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
		return fmt.Errorf("order %v %v %v %v %w", d.Exchange, d.AssetType, d.Pair, d.ID, errNotFutureAsset)
	}
	if c == nil {
		return common.ErrNilPointer
	}
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)]; !ok {
		c.positionTrackerControllers[strings.ToLower(d.Exchange)] = make(map[asset.Item]map[currency.Pair]*MultiPositionTracker)
	}
	if _, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType]; !ok {
		c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType] = make(map[currency.Pair]*MultiPositionTracker)
	}
	var err error
	mpt, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType][d.Pair]
	if !ok {
		mpt, err = SetupMultiPositionTracker(&PositionControllerSetup{
			Exchange:   strings.ToLower(d.Exchange),
			Asset:      d.AssetType,
			Pair:       d.Pair,
			Underlying: d.Pair.Base,
		})
		if err != nil {
			return err
		}
		c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType][d.Pair] = mpt
	}
	return mpt.TrackNewOrder(d)
}

func (c *PositionController) GetPositionsForExchange(exch string, item asset.Item, pair currency.Pair) ([]*PositionTracker, error) {
	if c == nil {
		return nil, common.ErrNilPointer
	}
	c.m.Lock()
	defer c.m.Unlock()
	exchM, ok := c.positionTrackerControllers[exch]
	if !ok {
		return nil, ErrPositionsNotLoadedForExchange
	}
	itemM, ok := exchM[item]
	if !ok {
		return nil, ErrPositionsNotLoadedForAsset
	}
	multiPositionTracker, ok := itemM[pair]
	if !ok {
		return nil, ErrPositionsNotLoadedForPair
	}
	return multiPositionTracker.GetPositions(), nil
}

// SetupMultiPositionTracker creates a futures order tracker for a specific exchange
func SetupMultiPositionTracker(setup *PositionControllerSetup) (*MultiPositionTracker, error) {
	if setup == nil {
		return nil, errNilSetup
	}
	if setup.Exchange == "" {
		return nil, errExchangeNameEmpty
	}
	if !setup.Asset.IsValid() || !setup.Asset.IsFutures() {
		return nil, errNotFutureAsset
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
	}, nil
}

// GetPositions returns all positions
func (e *MultiPositionTracker) GetPositions() []*PositionTracker {
	if e == nil {
		return nil
	}
	e.m.Lock()
	defer e.m.Unlock()
	return e.positions
}

// TrackNewOrder upserts an order to the tracker and updates position
// status and exposure. PNL is calculated separately as it requires mark prices
func (e *MultiPositionTracker) TrackNewOrder(d *Detail) error {
	if e == nil {
		return common.ErrNilPointer
	}
	if d == nil {
		return ErrSubmissionIsNil
	}
	e.m.Lock()
	defer e.m.Unlock()
	if d.AssetType != e.asset {
		return errAssetMismatch
	}
	if tracker, ok := e.orderPositions[d.ID]; ok {
		// this has already been associated
		// update the tracker
		return tracker.TrackNewOrder(d)
	}
	if len(e.positions) > 0 {
		for i := range e.positions {
			if e.positions[i].status == Open && i != len(e.positions)-1 {
				return fmt.Errorf("%w %v at position %v/%v", errPositionDiscrepancy, e.positions[i], i, len(e.positions)-1)
			}
		}
		if e.positions[len(e.positions)-1].status == Open {
			err := e.positions[len(e.positions)-1].TrackNewOrder(d)
			if err != nil && !errors.Is(err, ErrPositionClosed) {
				return err
			}
			e.orderPositions[d.ID] = e.positions[len(e.positions)-1]
			return nil
		}
	}
	setup := &PositionTrackerSetup{
		Pair:                      d.Pair,
		EntryPrice:                d.Price,
		Underlying:                d.Pair.Base,
		Asset:                     d.AssetType,
		Side:                      d.Side,
		UseExchangePNLCalculation: e.useExchangePNLCalculations,
	}
	tracker, err := e.SetupPositionTracker(setup)
	if err != nil {
		return err
	}
	e.positions = append(e.positions, tracker)
	err = tracker.TrackNewOrder(d)
	if err != nil {
		return err
	}
	e.orderPositions[d.ID] = tracker
	return nil
}

// SetupPositionTracker creates a new position tracker to track n futures orders
// until the position(s) are closed
func (e *MultiPositionTracker) SetupPositionTracker(setup *PositionTrackerSetup) (*PositionTracker, error) {
	if e == nil {
		return nil, common.ErrNilPointer
	}
	if e.exchange == "" {
		return nil, errExchangeNameEmpty
	}
	if setup == nil {
		return nil, errNilSetup
	}
	if !setup.Asset.IsValid() || !setup.Asset.IsFutures() {
		return nil, errNotFutureAsset
	}
	if setup.Pair.IsEmpty() {
		return nil, ErrPairIsEmpty
	}

	resp := &PositionTracker{
		exchange:                  strings.ToLower(e.exchange),
		asset:                     setup.Asset,
		contractPair:              setup.Pair,
		underlyingAsset:           setup.Underlying,
		status:                    Open,
		entryPrice:                decimal.NewFromFloat(setup.EntryPrice),
		currentDirection:          setup.Side,
		openingDirection:          setup.Side,
		useExchangePNLCalculation: e.useExchangePNLCalculations,
	}
	if !e.useExchangePNLCalculations {
		// use position tracker's pnl calculation by default
		resp.PNLCalculation = &PNLCalculator{}
	} else {
		resp.PNLCalculation = e.exchangePNLCalculation
	}
	return resp, nil
}

func (p *PositionTracker) GetStats() PositionStats {
	if p == nil {
		return PositionStats{}
	}
	p.m.Lock()
	defer p.m.Unlock()
	return PositionStats{
		Exchange:         p.exchange,
		Asset:            p.asset,
		Pair:             p.contractPair,
		Underlying:       p.underlyingAsset,
		Status:           p.status,
		Orders:           append(p.longPositions, p.shortPositions...),
		RealisedPNL:      p.realisedPNL,
		UnrealisedPNL:    p.unrealisedPNL,
		LatestDirection:  p.currentDirection,
		OpeningDirection: p.openingDirection,
		OpeningPrice:     p.entryPrice,
		LatestPrice:      p.latestPrice,
		PNLHistory:       p.pnlHistory,
	}
}

// TrackPNLByTime calculates the PNL based on a position tracker's exposure
// and current pricing. Adds the entry to PNL history to track over time
func (p *PositionTracker) TrackPNLByTime(t time.Time, currentPrice float64) error {
	if p == nil {
		return common.ErrNilPointer
	}
	p.m.Lock()
	defer func() {
		p.latestPrice = decimal.NewFromFloat(currentPrice)
		p.m.Unlock()
	}()
	price := decimal.NewFromFloat(currentPrice)
	result := &PNLResult{
		Time:  t,
		Price: price,
	}
	diff := price.Sub(p.entryPrice)
	result.UnrealisedPNL = p.exposure.Mul(diff)
	result.Price = price
	if len(p.pnlHistory) > 0 {
		result.RealisedPNLBeforeFees = p.pnlHistory[len(p.pnlHistory)-1].RealisedPNLBeforeFees
		result.Exposure = p.pnlHistory[len(p.pnlHistory)-1].Exposure
	}
	var err error
	p.pnlHistory, err = upsertPNLEntry(p.pnlHistory, result)
	return err
}

// GetRealisedPNL returns the realised pnl if the order
// is closed
func (p *PositionTracker) GetRealisedPNL() decimal.Decimal {
	p.m.Lock()
	defer p.m.Unlock()
	return p.calculateRealisedPNL(p.pnlHistory)
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
func (p *PositionTracker) TrackNewOrder(d *Detail) error {
	if p == nil {
		return common.ErrNilPointer
	}
	p.m.Lock()
	defer p.m.Unlock()
	if p.status == Closed {
		return ErrPositionClosed
	}
	if d == nil {
		return ErrSubmissionIsNil
	}
	if !p.contractPair.Equal(d.Pair) {
		return fmt.Errorf("%w pair '%v' received: '%v'", errOrderNotEqualToTracker, d.Pair, p.contractPair)
	}
	if p.exchange != strings.ToLower(d.Exchange) {
		return fmt.Errorf("%w exchange '%v' received: '%v'", errOrderNotEqualToTracker, d.Exchange, p.exchange)
	}
	if p.asset != d.AssetType {
		return fmt.Errorf("%w asset '%v' received: '%v'", errOrderNotEqualToTracker, d.AssetType, p.asset)
	}
	if d.Side == "" {
		return ErrSideIsInvalid
	}
	if d.ID == "" {
		return ErrOrderIDNotSet
	}
	if d.Date.IsZero() {
		return fmt.Errorf("%w for %v %v %v order ID: %v unset", errTimeUnset, d.Exchange, d.AssetType, d.Pair, d.ID)
	}
	if len(p.shortPositions) == 0 && len(p.longPositions) == 0 {
		p.entryPrice = decimal.NewFromFloat(d.Price)
	}

	for i := range p.shortPositions {
		if p.shortPositions[i].ID == d.ID {
			ord := p.shortPositions[i].Copy()
			ord.UpdateOrderFromDetail(d)
			p.shortPositions[i] = ord
			break
		}
	}
	for i := range p.longPositions {
		if p.longPositions[i].ID == d.ID {
			ord := p.longPositions[i].Copy()
			ord.UpdateOrderFromDetail(d)
			p.longPositions[i] = ord
			break
		}
	}

	if d.Side.IsShort() {
		p.shortPositions = append(p.shortPositions, d.Copy())
	} else {
		p.longPositions = append(p.longPositions, d.Copy())
	}
	var shortSide, longSide, averageLeverage decimal.Decimal

	for i := range p.shortPositions {
		shortSide = shortSide.Add(decimal.NewFromFloat(p.shortPositions[i].Amount))
		averageLeverage = decimal.NewFromFloat(p.shortPositions[i].Leverage)
	}
	for i := range p.longPositions {
		longSide = longSide.Add(decimal.NewFromFloat(p.longPositions[i].Amount))
		averageLeverage = decimal.NewFromFloat(p.longPositions[i].Leverage)
	}

	averageLeverage.Div(decimal.NewFromInt(int64(len(p.shortPositions))).Add(decimal.NewFromInt(int64(len(p.longPositions)))))
	if p.currentDirection == "" {
		p.currentDirection = d.Side
	}

	var result *PNLResult
	var err error
	cal := &PNLCalculatorRequest{
		Underlying:       p.underlyingAsset,
		Asset:            p.asset,
		OrderDirection:   d.Side,
		Leverage:         d.Leverage,
		EntryPrice:       p.entryPrice.InexactFloat64(),
		Amount:           d.Amount,
		CurrentPrice:     d.Price,
		Pair:             p.contractPair,
		Time:             d.Date,
		OpeningDirection: p.openingDirection,
		CurrentDirection: p.currentDirection,
		PNLHistory:       p.pnlHistory,
		Exposure:         p.exposure,
		Fee:              decimal.NewFromFloat(d.Fee),
	}
	if len(p.pnlHistory) != 0 {
		cal.PreviousPrice = p.pnlHistory[len(p.pnlHistory)-1].Price.InexactFloat64()
	}
	amount := decimal.NewFromFloat(d.Amount)
	if (cal.OrderDirection.IsShort() && cal.CurrentDirection.IsLong() || cal.OrderDirection.IsLong() && cal.CurrentDirection.IsShort()) &&
		cal.Exposure.LessThan(amount) {
		// latest order swaps directions!
		// split the order to calculate PNL from each direction
		first := amount.Sub(cal.Exposure)
		second := cal.Exposure.Sub(amount).Abs()
		cal.Fee = cal.Fee.Div(decimal.NewFromInt(2))
		result, err = createPNLResult(
			cal.Time,
			cal.OrderDirection,
			first,
			decimal.NewFromFloat(d.Price),
			cal.Fee,
			cal.OpeningDirection,
			cal.CurrentDirection,
			cal.PNLHistory)
		if err != nil {
			return err
		}
		p.pnlHistory, err = upsertPNLEntry(cal.PNLHistory, result)
		if err != nil {
			return err
		}

		if cal.OrderDirection.IsLong() {
			cal.OrderDirection = Short
		} else if cal.OrderDirection.IsShort() {
			cal.OrderDirection = Long
		}
		if cal.OpeningDirection.IsLong() {
			cal.OpeningDirection = Short
		} else if cal.OpeningDirection.IsShort() {
			cal.OpeningDirection = Long
		}

		cal.EntryPrice = d.Price
		result, err = createPNLResult(
			cal.Time.Add(1),
			cal.OrderDirection,
			second,
			decimal.NewFromFloat(d.Price),
			cal.Fee,
			cal.OpeningDirection,
			cal.CurrentDirection,
			cal.PNLHistory)
	} else {
		result, err = p.PNLCalculation.CalculatePNL(cal)
		if err != nil {
			if !errors.Is(err, ErrPositionLiquidated) {
				return err
			}
			result.UnrealisedPNL = decimal.Zero
			result.RealisedPNLBeforeFees = decimal.Zero
			p.status = Closed
		}
	}
	p.pnlHistory, err = upsertPNLEntry(p.pnlHistory, result)
	if err != nil {
		return err
	}
	p.unrealisedPNL = result.UnrealisedPNL

	if longSide.GreaterThan(shortSide) {
		p.currentDirection = Long
	} else if shortSide.GreaterThan(longSide) {
		p.currentDirection = Short
	} else {
		p.currentDirection = UnknownSide
	}
	if p.currentDirection.IsLong() {
		p.exposure = longSide.Sub(shortSide)
	} else {
		p.exposure = shortSide.Sub(longSide)
	}
	if p.exposure.Equal(decimal.Zero) {
		p.status = Closed
		p.closingPrice = decimal.NewFromFloat(d.Price)
		p.realisedPNL = p.calculateRealisedPNL(p.pnlHistory)
		p.unrealisedPNL = decimal.Zero
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

// CalculatePNL this is a localised generic way of calculating open
// positions' worth, it is an implementation of the PNLCalculation interface
//
// do not use any properties of p, use calc, otherwise there will be
// sync issues
func (p *PNLCalculator) CalculatePNL(calc *PNLCalculatorRequest) (*PNLResult, error) {
	if calc == nil {
		return nil, ErrNilPNLCalculator
	}
	var price, amount decimal.Decimal
	price = decimal.NewFromFloat(calc.CurrentPrice)
	amount = decimal.NewFromFloat(calc.Amount)
	result, err := createPNLResult(
		calc.Time,
		calc.OrderDirection,
		amount,
		price,
		calc.Fee,
		calc.OpeningDirection,
		calc.CurrentDirection,
		calc.PNLHistory)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createPNLResult(t time.Time, side Side, amount, price, fee decimal.Decimal, openingDirection, currentDirection Side, pnlHistory []PNLResult) (*PNLResult, error) {
	var previousPNL *PNLResult
	if len(pnlHistory) > 0 {
		previousPNL = &pnlHistory[len(pnlHistory)-1]
	}
	var prevExposure decimal.Decimal
	if previousPNL != nil {
		prevExposure = previousPNL.Exposure
	}
	var currentExposure, realisedPNL, unrealisedPNL, first, second decimal.Decimal
	if openingDirection.IsLong() {
		first = price
		if previousPNL != nil {
			second = previousPNL.Price
		}
	} else if openingDirection.IsShort() {
		if previousPNL != nil {
			first = previousPNL.Price
			second = price
		}
	}
	switch {
	case currentDirection.IsShort() && side.IsShort(),
		currentDirection.IsLong() && side.IsLong():
		// appending to your position
		currentExposure = prevExposure.Add(amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
	case currentDirection.IsShort() && side.IsLong(),
		currentDirection.IsLong() && side.IsShort():
		// selling/closing your position by "amount"
		currentExposure = prevExposure.Sub(amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
		step1 := first.Sub(second)
		realisedPNL = amount.Mul(step1)
	default:
		return nil, fmt.Errorf("%v %v %v %w", currentDirection, side, currentExposure, errCannotCalculateUnrealisedPNL)
	}
	totalFees := fee
	for i := range pnlHistory {
		totalFees = totalFees.Add(pnlHistory[i].Fee)
	}
	if !unrealisedPNL.IsZero() {
		unrealisedPNL = unrealisedPNL.Sub(totalFees)
	}

	response := &PNLResult{
		Time:                  t,
		UnrealisedPNL:         unrealisedPNL,
		RealisedPNLBeforeFees: realisedPNL,
		Price:                 price,
		Exposure:              currentExposure,
		Fee:                   fee,
	}
	return response, nil
}

func (p *PositionTracker) calculateRealisedPNL(pnlHistory []PNLResult) decimal.Decimal {
	var realisedPNL, totalFees decimal.Decimal
	for i := range pnlHistory {
		realisedPNL = realisedPNL.Add(pnlHistory[i].RealisedPNLBeforeFees)
		totalFees = totalFees.Add(pnlHistory[i].Fee)
	}
	if realisedPNL.IsZero() {
		return decimal.Zero
	}
	fullyDone := realisedPNL.Sub(totalFees)
	return fullyDone
}

// upsertPNLEntry upserts an entry to PNLHistory field
// with some basic checks
func upsertPNLEntry(pnlHistory []PNLResult, entry *PNLResult) ([]PNLResult, error) {
	if entry.Time.IsZero() {
		return nil, errTimeUnset
	}
	for i := range pnlHistory {
		if entry.Time.Equal(pnlHistory[i].Time) {
			pnlHistory[i] = *entry
			return pnlHistory, nil
		}
	}
	pnlHistory = append(pnlHistory, *entry)
	return pnlHistory, nil
}
