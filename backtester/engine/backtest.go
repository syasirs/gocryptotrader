package engine

import (
	"errors"
	"fmt"

	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/data"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/exchange"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/statistics"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/strategies/base"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/fill"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/kline"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/order"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/signal"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	gctcommon "github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	gctexchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// Reset BackTest values to default
func (bt *BackTest) Reset() {
	if bt == nil {
		return
	}
	err := bt.orderManager.Stop()
	if err != nil {
		log.Error(common.Backtester, err)
	}
	err = bt.databaseManager.Stop()
	if err != nil {
		log.Error(common.Backtester, err)
	}
	bt.EventQueue.Reset()
	bt.DataHolder.Reset()
	bt.Portfolio.Reset()
	bt.Statistic.Reset()
	bt.Exchange.Reset()
	bt.Funding.Reset()
	bt.exchangeManager = nil
	bt.orderManager = nil
	bt.databaseManager = nil
}

// RunLive is a proof of concept function that does not yet support multi currency usage
// It runs by constantly checking for new live datas and running through the list of events
// once new data is processed. It will run until application close event has been received
func (bt *BackTest) RunLive() error {
	if bt.LiveDataHandler == nil {
		return errLiveOnly
	}
	log.Info(common.Livetester, "running backtester against live data")
	for {
		select {
		case <-bt.shutdown:
			return bt.LiveDataHandler.Stop()
		case <-bt.LiveDataHandler.Updated():
			bt.Run()
		case <-bt.LiveDataHandler.HasShutdown():
			return nil
		}
	}
}

// Run will iterate over loaded data events
// save them and then handle the event based on its type
func (bt *BackTest) Run() {
	var doubleNil bool
dataLoadingIssue:
	for ev := bt.EventQueue.NextEvent(); ; ev = bt.EventQueue.NextEvent() {
		if ev == nil {
			if doubleNil {
				if bt.verbose {
					log.Info(common.Backtester, "no new data on second check")
				}
				break dataLoadingIssue
			}
			doubleNil = true
			dataHandlerMap := bt.DataHolder.GetAllData()
			for exchangeName, exchangeMap := range dataHandlerMap {
				for assetItem, assetMap := range exchangeMap {
					for baseCurrency, baseMap := range assetMap {
						for quoteCurrency, dataHandler := range baseMap {
							d := dataHandler.Next()
							if d == nil {
								if !bt.hasProcessedAnEvent && bt.LiveDataHandler == nil {
									log.Errorf(common.Backtester, "Unable to perform `Next` for %v %v %v-%v", exchangeName, assetItem, baseCurrency, quoteCurrency)
								}
								break dataLoadingIssue
							}
							o := d.GetOffset()
							if bt.Strategy.UsingSimultaneousProcessing() && bt.hasProcessedDataAtOffset[o] {
								// only append one event, as simultaneous processing
								// will retrieve all relevant events to process under
								// processSimultaneousDataEvents()
								continue
							}
							bt.EventQueue.AppendEvent(d)
							if !bt.hasProcessedDataAtOffset[o] {
								bt.hasProcessedDataAtOffset[o] = true
							}
						}
					}
				}
			}
		} else {
			doubleNil = false
			err := bt.handleEvent(ev)
			if err != nil {
				log.Error(common.Backtester, err)
			}
			if !bt.hasProcessedAnEvent {
				bt.hasProcessedAnEvent = true
			}
		}
	}
}

// handleEvent is the main processor of data for the backtester
// after data has been loaded and Run has appended a data event to the queue,
// handle event will process events and add further events to the queue if they
// are required
func (bt *BackTest) handleEvent(ev common.Event) error {
	if ev == nil {
		return fmt.Errorf("cannot handle event %w", errNilData)
	}
	funds, err := bt.Funding.GetFundingForEvent(ev)
	if err != nil {
		return err
	}

	if bt.Funding.HasFutures() {
		err = bt.Funding.UpdateCollateral(ev)
		if err != nil {
			return err
		}
	}

	switch eType := ev.(type) {
	case kline.Event:
		// using kline.Event as signal.Event also matches data.Event
		if bt.Strategy.UsingSimultaneousProcessing() {
			err = bt.processSimultaneousDataEvents()
		} else {
			err = bt.processSingleDataEvent(eType, funds.FundReleaser())
		}
	case signal.Event:
		err = bt.processSignalEvent(eType, funds.FundReserver())
	case order.Event:
		err = bt.processOrderEvent(eType, funds.FundReleaser())
	case fill.Event:
		err = bt.processFillEvent(eType, funds.FundReleaser())
		if bt.LiveDataHandler != nil {
			// output log data per interval instead of at the end
			var result string
			result, err = bt.Statistic.CreateLog(eType)
			if err != nil {
				log.Error(common.Livetester, err)
			} else {
				log.Info(common.Livetester, result)
			}
		}
	default:
		err = fmt.Errorf("handleEvent %w %T received, could not process",
			errUnhandledDatatype,
			ev)
	}
	if err != nil {
		return err
	}

	bt.Funding.CreateSnapshot(ev.GetTime())
	return nil
}

// processSingleDataEvent will pass the event to the strategy and determine how it should be handled
func (bt *BackTest) processSingleDataEvent(ev data.Event, funds funding.IFundReleaser) error {
	err := bt.updateStatsForDataEvent(ev, funds)
	if err != nil {
		return err
	}
	d, err := bt.DataHolder.GetDataForCurrency(ev)
	if err != nil {
		return err
	}
	s, err := bt.Strategy.OnSignal(d, bt.Funding, bt.Portfolio)
	if err != nil {
		if errors.Is(err, base.ErrTooMuchBadData) {
			// too much bad data is a severe error and backtesting must cease
			return err
		}
		log.Errorf(common.Backtester, "OnSignal %v", err)
		return nil
	}
	err = bt.Statistic.SetEventForOffset(s)
	if err != nil {
		log.Errorf(common.Backtester, "SetEventForOffset %v", err)
	}
	bt.EventQueue.AppendEvent(s)

	return nil
}

// processSimultaneousDataEvents determines what signal events are generated and appended
// to the event queue. It will pass all currency events to the strategy to determine what
// currencies to act upon
func (bt *BackTest) processSimultaneousDataEvents() error {
	var dataEvents []data.Handler
	dataHandlerMap := bt.DataHolder.GetAllData()
	for _, exchangeMap := range dataHandlerMap {
		for _, assetMap := range exchangeMap {
			for _, baseMap := range assetMap {
				for _, dataHandler := range baseMap {
					latestData := dataHandler.Latest()
					funds, err := bt.Funding.GetFundingForEvent(latestData)
					if err != nil {
						return err
					}
					err = bt.updateStatsForDataEvent(latestData, funds.FundReleaser())
					if err != nil {
						switch {
						case errors.Is(err, statistics.ErrAlreadyProcessed):
							log.Warnf(common.Livetester, "%v %v", latestData.GetOffset(), err)
							continue
						case errors.Is(err, gctorder.ErrPositionLiquidated):
							return nil
						default:
							log.Error(common.Backtester, err)
						}
					}
					dataEvents = append(dataEvents, dataHandler)
				}
			}
		}
	}
	signals, err := bt.Strategy.OnSimultaneousSignals(dataEvents, bt.Funding, bt.Portfolio)
	if err != nil {
		if errors.Is(err, base.ErrTooMuchBadData) {
			// too much bad data is a severe error and backtesting must cease
			return err
		}
		log.Errorf(common.Backtester, "OnSimultaneousSignals %v", err)
		return nil
	}
	for i := range signals {
		err = bt.Statistic.SetEventForOffset(signals[i])
		if err != nil {
			log.Errorf(common.Backtester, "SetEventForOffset %v %v %v %v", signals[i].GetExchange(), signals[i].GetAssetType(), signals[i].Pair(), err)
		}
		bt.EventQueue.AppendEvent(signals[i])
	}
	return nil
}

// updateStatsForDataEvent makes various systems aware of price movements from
// data events
func (bt *BackTest) updateStatsForDataEvent(ev data.Event, funds funding.IFundReleaser) error {
	if ev == nil {
		return common.ErrNilEvent
	}
	if funds == nil {
		return fmt.Errorf("%v %v %v %w missing fund releaser", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), common.ErrNilArguments)
	}
	// update statistics with the latest price
	err := bt.Statistic.SetEventForOffset(ev)
	if err != nil {
		if errors.Is(err, statistics.ErrAlreadyProcessed) {
			return err
		}
		log.Errorf(common.Backtester, "SetEventForOffset %v", err)
	}
	// update portfolio manager with the latest price
	err = bt.Portfolio.UpdateHoldings(ev, funds)
	if err != nil {
		log.Errorf(common.Backtester, "UpdateHoldings %v", err)
	}

	if ev.GetAssetType().IsFutures() {
		var cr funding.ICollateralReleaser
		cr, err = funds.CollateralReleaser()
		if err != nil {
			return err
		}

		err = bt.Portfolio.UpdatePNL(ev, ev.GetClosePrice())
		if err != nil {
			if errors.Is(err, gctorder.ErrPositionsNotLoadedForPair) {
				// if there is no position yet, there's nothing to update
				return nil
			}
			if !errors.Is(err, gctorder.ErrPositionLiquidated) {
				return fmt.Errorf("UpdatePNL %v", err)
			}
		}
		var pnl *portfolio.PNLSummary
		pnl, err = bt.Portfolio.GetLatestPNLForEvent(ev)
		if err != nil {
			return err
		}

		if pnl.Result.IsLiquidated {
			return nil
		}
		err = bt.Portfolio.CheckLiquidationStatus(ev, cr, pnl)
		if err != nil {
			if errors.Is(err, gctorder.ErrPositionLiquidated) {
				liquidErr := bt.triggerLiquidationsForExchange(ev, pnl)
				if liquidErr != nil {
					return liquidErr
				}
			}
			return err
		}

		return bt.Statistic.AddPNLForTime(pnl)
	}

	return nil
}

func (bt *BackTest) triggerLiquidationsForExchange(ev data.Event, pnl *portfolio.PNLSummary) error {
	if ev == nil {
		return common.ErrNilEvent
	}
	if pnl == nil {
		return fmt.Errorf("%w pnl summary", common.ErrNilArguments)
	}
	orders, err := bt.Portfolio.CreateLiquidationOrdersForExchange(ev, bt.Funding)
	if err != nil {
		return err
	}
	for i := range orders {
		// these orders are raising events for event offsets
		// which may not have been processed yet
		// this will create and store stats for each order
		// then liquidate it at the funding level
		var datas data.Handler
		datas, err = bt.DataHolder.GetDataForCurrency(orders[i])
		if err != nil {
			return err
		}
		latest := datas.Latest()
		err = bt.Statistic.SetEventForOffset(latest)
		if err != nil && !errors.Is(err, statistics.ErrAlreadyProcessed) {
			return err
		}
		bt.EventQueue.AppendEvent(orders[i])
		err = bt.Statistic.SetEventForOffset(orders[i])
		if err != nil {
			log.Errorf(common.Backtester, "SetEventForOffset %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		}
		bt.Funding.Liquidate(orders[i])
	}
	pnl.Result.IsLiquidated = true
	pnl.Result.Status = gctorder.Liquidated
	return bt.Statistic.AddPNLForTime(pnl)
}

// processSignalEvent receives an event from the strategy for processing under the portfolio
func (bt *BackTest) processSignalEvent(ev signal.Event, funds funding.IFundReserver) error {
	if ev == nil {
		return common.ErrNilEvent
	}
	if funds == nil {
		return fmt.Errorf("%w funds", common.ErrNilArguments)
	}
	cs, err := bt.Exchange.GetCurrencySettings(ev.GetExchange(), ev.GetAssetType(), ev.Pair())
	if err != nil {
		log.Errorf(common.Backtester, "GetCurrencySettings %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		return fmt.Errorf("GetCurrencySettings %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}
	var o *order.Order
	o, err = bt.Portfolio.OnSignal(ev, &cs, funds)
	if err != nil {
		log.Errorf(common.Backtester, "OnSignal %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		return fmt.Errorf("OnSignal %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}
	err = bt.Statistic.SetEventForOffset(o)
	if err != nil {
		return fmt.Errorf("SetEventForOffset %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}

	bt.EventQueue.AppendEvent(o)
	return nil
}

func (bt *BackTest) processOrderEvent(ev order.Event, funds funding.IFundReleaser) error {
	if ev == nil {
		return common.ErrNilEvent
	}
	if funds == nil {
		return fmt.Errorf("%w funds", common.ErrNilArguments)
	}
	d, err := bt.DataHolder.GetDataForCurrency(ev)
	if err != nil {
		return err
	}
	f, err := bt.Exchange.ExecuteOrder(ev, d, bt.orderManager, funds)
	if err != nil {
		if f == nil {
			log.Errorf(common.Backtester, "ExecuteOrder fill event should always be returned, please fix, %v", err)
			return fmt.Errorf("ExecuteOrder fill event should always be returned, please fix, %v", err)
		}
		if !errors.Is(err, exchange.ErrCannotTransact) {
			log.Errorf(common.Backtester, "ExecuteOrder %v %v %v %v", f.GetExchange(), f.GetAssetType(), f.Pair(), err)
		}
	}
	err = bt.Statistic.SetEventForOffset(f)
	if err != nil {
		log.Errorf(common.Backtester, "SetEventForOffset %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}
	bt.EventQueue.AppendEvent(f)
	return nil
}

func (bt *BackTest) processFillEvent(ev fill.Event, funds funding.IFundReleaser) error {
	_, err := bt.Portfolio.OnFill(ev, funds)
	if err != nil {
		return fmt.Errorf("OnFill %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}

	var holding *holdings.Holding
	holding, err = bt.Portfolio.ViewHoldingAtTimePeriod(ev)
	if err != nil {
		log.Error(common.Backtester, err)
	}
	if holding == nil {
		log.Error(common.Backtester, "ViewHoldingAtTimePeriod why is holdings nil?")
	} else {
		err = bt.Statistic.AddHoldingsForTime(holding)
		if err != nil {
			log.Errorf(common.Backtester, "AddHoldingsForTime %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		}
	}

	var cp *compliance.Manager
	cp, err = bt.Portfolio.GetComplianceManager(ev.GetExchange(), ev.GetAssetType(), ev.Pair())
	if err != nil {
		log.Errorf(common.Backtester, "GetComplianceManager %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}

	snap := cp.GetLatestSnapshot()
	err = bt.Statistic.AddComplianceSnapshotForTime(snap, ev)
	if err != nil {
		log.Errorf(common.Backtester, "AddComplianceSnapshotForTime %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}

	fde := ev.GetFillDependentEvent()
	if fde != nil && !fde.IsNil() {
		// some events can only be triggered on a successful fill event
		fde.SetOffset(ev.GetOffset())
		err = bt.Statistic.SetEventForOffset(fde)
		if err != nil {
			log.Errorf(common.Backtester, "SetEventForOffset %v %v %v %v", fde.GetExchange(), fde.GetAssetType(), fde.Pair(), err)
		}
		od := ev.GetOrder()
		if fde.MatchOrderAmount() && od != nil {
			fde.SetAmount(ev.GetAmount())
		}
		fde.AppendReasonf("raising event after %v %v %v fill", ev.GetExchange(), ev.GetAssetType(), ev.Pair())
		bt.EventQueue.AppendEvent(fde)
	}
	if ev.GetAssetType().IsFutures() {
		return bt.processFuturesFillEvent(ev, funds)
	}
	return nil
}

func (bt *BackTest) processFuturesFillEvent(ev fill.Event, funds funding.IFundReleaser) error {
	if ev.GetOrder() != nil {
		pnl, err := bt.Portfolio.TrackFuturesOrder(ev, funds)
		if err != nil && !errors.Is(err, gctorder.ErrSubmissionIsNil) {
			return fmt.Errorf("TrackFuturesOrder %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		}

		var exch gctexchange.IBotExchange
		exch, err = bt.exchangeManager.GetExchangeByName(ev.GetExchange())
		if err != nil {
			return fmt.Errorf("GetExchangeByName %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		}

		rPNL := pnl.GetRealisedPNL()
		if !rPNL.PNL.IsZero() {
			var receivingCurrency currency.Code
			var receivingAsset asset.Item
			receivingCurrency, receivingAsset, err = exch.GetCurrencyForRealisedPNL(ev.GetAssetType(), ev.Pair())
			if err != nil {
				return fmt.Errorf("GetCurrencyForRealisedPNL %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
			}
			err = bt.Funding.RealisePNL(ev.GetExchange(), receivingAsset, receivingCurrency, rPNL.PNL)
			if err != nil {
				return fmt.Errorf("RealisePNL %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
			}
		}

		err = bt.Statistic.AddPNLForTime(pnl)
		if err != nil {
			log.Errorf(common.Backtester, "AddHoldingsForTime %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
		}
	}
	err := bt.Funding.UpdateCollateral(ev)
	if err != nil {
		return fmt.Errorf("UpdateCollateral %v %v %v %v", ev.GetExchange(), ev.GetAssetType(), ev.Pair(), err)
	}
	return nil
}

// Stop shuts down the live data loop
func (bt *BackTest) Stop() {
	close(bt.shutdown)
}

// CloseAllPositions will close sell any positions held on closure
// can only be with live testing and where a strategy supports it
func (bt *BackTest) CloseAllPositions() error {
	if bt.LiveDataHandler == nil {
		return errLiveOnly
	}
	allData := bt.DataHolder.GetAllData()
	var latestPrices []data.Event
	for _, exchangeMap := range allData {
		for _, assetMap := range exchangeMap {
			for _, baseMap := range assetMap {
				for _, handler := range baseMap {
					latestPrices = append(latestPrices, handler.Latest())
				}
			}
		}
	}

	events, err := bt.Strategy.CloseAllPositions(bt.Portfolio.GetLatestHoldingsForAllCurrencies(), latestPrices)
	if err != nil {
		if errors.Is(err, gctcommon.ErrFunctionNotSupported) {
			log.Warnf(common.Livetester, "closing all positions is not supported by strategy %v", bt.Strategy.Name())
			return nil
		}
		return err
	}
	err = bt.LiveDataHandler.SetDataForClosingAllPositions(events...)
	if err != nil {
		return err
	}

	for i := range events {
		k := events[i].ToKline()
		err = bt.Statistic.SetEventForOffset(k)
		if err != nil {
			log.Errorf(common.Backtester, "SetEventForOffset kline %v %v %v %v", events[i].GetExchange(), events[i].GetAssetType(), events[i].Pair(), err)
		}
		err = bt.Statistic.SetEventForOffset(events[i])
		if err != nil {
			log.Errorf(common.Backtester, "SetEventForOffset signal %v %v %v %v", events[i].GetExchange(), events[i].GetAssetType(), events[i].Pair(), err)
		}
		bt.EventQueue.AppendEvent(events[i])
	}
	// signal.Event run
	bt.Run()
	// order.Event run
	bt.Run()
	// fill.Event run
	bt.Run()

	return nil
}
