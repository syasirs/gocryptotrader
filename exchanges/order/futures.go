package order

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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
	if _, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)]; !ok {
		c.positionTrackerControllers[strings.ToLower(d.Exchange)] = make(map[asset.Item]map[currency.Pair]*MultiPositionTracker)
	}
	if _, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType]; !ok {
		c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType] = make(map[currency.Pair]*MultiPositionTracker)
	}
	if _, ok := c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType][d.Pair]; !ok {
		ptc, err := SetupMultiPositionTracker(&PositionControllerSetup{
			Exchange:   strings.ToLower(d.Exchange),
			Asset:      d.AssetType,
			Pair:       d.Pair,
			Underlying: d.Pair.Base,
		})
		if err != nil {
			return err
		}
		c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType][d.Pair] = ptc
	}
	return c.positionTrackerControllers[strings.ToLower(d.Exchange)][d.AssetType][d.Pair].TrackNewOrder(d)
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
	return e.positions
}

// TrackNewOrder upserts an order to the tracker and updates position
// status and exposure. PNL is calculated separately as it requires mark prices
func (e *MultiPositionTracker) TrackNewOrder(d *Detail) error {
	if d == nil {
		return ErrSubmissionIsNil
	}
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
			if err != nil && !errors.Is(err, errPositionClosed) {
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
		resp.PNLCalculation = resp
	} else {
		resp.PNLCalculation = e.exchangePNLCalculation
	}
	return resp, nil
}

// TrackPNLByTime calculates the PNL based on a position tracker's exposure
// and current pricing. Adds the entry to PNL history to track over time
func (p *PositionTracker) TrackPNLByTime(t time.Time, currentPrice float64) error {
	defer func() {
		p.latestPrice = decimal.NewFromFloat(currentPrice)
	}()
	pnl, err := p.PNLCalculation.CalculatePNL(&PNLCalculator{
		TimeBasedCalculation: &TimeBasedCalculation{
			Time:         t,
			CurrentPrice: currentPrice,
		},
	})
	if err != nil {
		return err
	}
	return p.UpsertPNLEntry(&PNLResult{
		Time:                  pnl.Time,
		UnrealisedPNL:         pnl.UnrealisedPNL,
		RealisedPNLBeforeFees: pnl.RealisedPNLBeforeFees,
	})
}

// GetRealisedPNL returns the realised pnl if the order
// is closed
func (p *PositionTracker) GetRealisedPNL() decimal.Decimal {
	return p.calculateRealisedPNL()
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
	if p.status == Closed {
		return errPositionClosed
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
	if p.useExchangePNLCalculation {
		cal := &ExchangeBasedCalculation{
			Underlying:   p.underlyingAsset,
			Asset:        p.asset,
			Side:         d.Side,
			Leverage:     d.Leverage,
			EntryPrice:   p.entryPrice.InexactFloat64(),
			Amount:       d.Amount,
			CurrentPrice: d.Price,
			Pair:         p.contractPair,
			Time:         d.Date,
		}
		if len(p.pnlHistory) != 0 {
			cal.PreviousPrice = p.pnlHistory[len(p.pnlHistory)-1].Price.InexactFloat64()
		}
		result, err = p.PNLCalculation.CalculatePNL(&PNLCalculator{ExchangeBasedCalculation: cal})
	} else {
		result, err = p.PNLCalculation.CalculatePNL(&PNLCalculator{OrderBasedCalculation: d})
	}
	if err != nil {
		if !errors.Is(err, ErrPositionLiquidated) {
			return err
		}
		result.UnrealisedPNL = decimal.Zero
		result.RealisedPNLBeforeFees = decimal.Zero
		p.status = Closed
	}
	upsertErr := p.UpsertPNLEntry(result)
	if upsertErr != nil {
		return upsertErr
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
		p.realisedPNL = p.calculateRealisedPNL()
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

func (p *PositionTracker) calculateRealisedPNL() decimal.Decimal {
	var realisedPNL, totalFees decimal.Decimal
	for i := range p.pnlHistory {
		realisedPNL = realisedPNL.Add(p.pnlHistory[i].RealisedPNLBeforeFees)
		totalFees = totalFees.Add(p.pnlHistory[i].Fee)
	}
	if realisedPNL.IsZero() {
		return decimal.Zero
	}
	fullyDone := realisedPNL.Sub(totalFees)
	return fullyDone
}

// CalculatePNL this is a localised generic way of calculating open
// positions' worth
func (p *PositionTracker) CalculatePNL(calc *PNLCalculator) (*PNLResult, error) {
	if calc == nil {
		return nil, ErrNilPNLCalculator
	}
	result := &PNLResult{}
	var price, amount decimal.Decimal
	var err error
	if calc.OrderBasedCalculation != nil {
		price = decimal.NewFromFloat(calc.OrderBasedCalculation.Price)
		amount = decimal.NewFromFloat(calc.OrderBasedCalculation.Amount)
		fee := decimal.NewFromFloat(calc.OrderBasedCalculation.Fee)
		if (p.currentDirection.IsShort() && calc.OrderBasedCalculation.Side.IsLong() || p.currentDirection.IsLong() && calc.OrderBasedCalculation.Side.IsShort()) &&
			p.exposure.LessThan(amount) {

			// latest order swaps directions!
			first := amount.Sub(p.exposure)
			second := p.exposure.Sub(amount).Abs()
			fee = fee.Div(decimal.NewFromInt(2))
			result, err = p.calculatePNL(calc.OrderBasedCalculation.Date, calc.OrderBasedCalculation.Side, first, price, fee)
			if err != nil {
				return nil, err
			}
			err = p.UpsertPNLEntry(result)
			if err != nil {
				return nil, err
			}

			if calc.OrderBasedCalculation.Side.IsLong() {
				calc.OrderBasedCalculation.Side = Short
			} else if calc.OrderBasedCalculation.Side.IsShort() {
				calc.OrderBasedCalculation.Side = Long
			}
			if p.openingDirection.IsLong() {
				p.openingDirection = Short
			} else if p.openingDirection.IsShort() {
				p.openingDirection = Long
			}

			p.entryPrice = price
			result, err = p.calculatePNL(calc.OrderBasedCalculation.Date.Add(1), calc.OrderBasedCalculation.Side, second, price, fee)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		result, err = p.calculatePNL(calc.OrderBasedCalculation.Date, calc.OrderBasedCalculation.Side, amount, price, fee)
		if err != nil {
			return nil, err
		}

		return result, nil
	} else if calc.TimeBasedCalculation != nil {
		result.Time = calc.TimeBasedCalculation.Time
		price = decimal.NewFromFloat(calc.TimeBasedCalculation.CurrentPrice)
		diff := price.Sub(p.entryPrice)
		result.UnrealisedPNL = p.exposure.Mul(diff)
		result.Price = price
		return result, nil
	}

	return nil, errMissingPNLCalculationFunctions
}

func (p *PositionTracker) calculatePNL(t time.Time, side Side, amount, price, fee decimal.Decimal) (*PNLResult, error) {
	var previousPNL *PNLResult
	if len(p.pnlHistory) > 0 {
		previousPNL = &p.pnlHistory[len(p.pnlHistory)-1]
	}
	var prevExposure decimal.Decimal
	if previousPNL != nil {
		prevExposure = previousPNL.Exposure
	}
	var currentExposure, realisedPNL, unrealisedPNL, first, second decimal.Decimal
	if p.openingDirection.IsLong() {
		first = price
		if previousPNL != nil {
			second = previousPNL.Price
		}
	} else if p.openingDirection.IsShort() {
		if previousPNL != nil {
			first = previousPNL.Price
			second = price
		}
	}
	switch {
	case p.currentDirection.IsShort() && side.IsShort(),
		p.currentDirection.IsLong() && side.IsLong():
		// appending to your position
		currentExposure = prevExposure.Add(amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
	case p.currentDirection.IsShort() && side.IsLong(),
		p.currentDirection.IsLong() && side.IsShort():
		// selling/closing your position by "amount"
		currentExposure = prevExposure.Sub(amount)
		unrealisedPNL = currentExposure.Mul(first.Sub(second))
		step1 := first.Sub(second)
		realisedPNL = amount.Mul(step1)
	default:
		return nil, fmt.Errorf("%v %v %v %w", p.currentDirection, side, currentExposure, errCannotCalculateUnrealisedPNL)
	}
	totalFees := fee
	for i := range p.pnlHistory {
		totalFees = totalFees.Add(p.pnlHistory[i].Fee)
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

// UpsertPNLEntry upserts an entry to PNLHistory field
// with some basic checks
func (p *PositionTracker) UpsertPNLEntry(entry *PNLResult) error {
	if entry.Time.IsZero() {
		return errTimeUnset
	}
	for i := range p.pnlHistory {
		if entry.Time.Equal(p.pnlHistory[i].Time) {
			p.pnlHistory[i] = *entry
			return nil
		}
	}
	p.pnlHistory = append(p.pnlHistory, *entry)
	return nil
}

// IsShort returns if the side is short
func (s Side) IsShort() bool {
	return s == Short || s == Sell
}

// IsLong returns if the side is long
func (s Side) IsLong() bool {
	return s == Long || s == Buy
}
