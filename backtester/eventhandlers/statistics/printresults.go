package statistics

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio"
	gctcommon "github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	limit12 = 12
	limit14 = 14
	limit10 = 10
)

// addReason basic helper to append event reason if one is there
func addReason(reason, msg string) string {
	if reason != "" {
		msg += "\tReason: " + reason
	}
	return msg
}

// PrintTotalResults outputs all results to the CMD
func (s *Statistic) PrintTotalResults() {
	log.Info(common.Statistics, common.ColourH1+"------------------Strategy-----------------------------------"+common.ColourDefault)
	log.Infof(common.Statistics, "Strategy Name: %v", s.StrategyName)
	log.Infof(common.Statistics, "Strategy Nickname: %v", s.StrategyNickname)
	log.Infof(common.Statistics, "Strategy Goal: %v\n\n", s.StrategyGoal)

	log.Info(common.Statistics, common.ColourH2+"------------------Total Results------------------------------"+common.ColourDefault)
	log.Info(common.Statistics, common.ColourH3+"------------------Orders-------------------------------------"+common.ColourDefault)
	log.Infof(common.Statistics, "Total buy orders: %v", convert.IntToHumanFriendlyString(s.TotalBuyOrders, ","))
	log.Infof(common.Statistics, "Total sell orders: %v", convert.IntToHumanFriendlyString(s.TotalSellOrders, ","))
	log.Infof(common.Statistics, "Total long orders: %v", convert.IntToHumanFriendlyString(s.TotalLongOrders, ","))
	log.Infof(common.Statistics, "Total short orders: %v", convert.IntToHumanFriendlyString(s.TotalShortOrders, ","))
	log.Infof(common.Statistics, "Total orders: %v\n\n", convert.IntToHumanFriendlyString(s.TotalOrders, ","))

	if s.BiggestDrawdown != nil {
		log.Info(common.Statistics, common.ColourH3+"------------------Biggest Drawdown-----------------------"+common.ColourDefault)
		log.Infof(common.Statistics, "Exchange: %v Asset: %v Currency: %v", s.BiggestDrawdown.Exchange, s.BiggestDrawdown.Asset, s.BiggestDrawdown.Pair)
		log.Infof(common.Statistics, "Highest Price: %s", convert.DecimalToHumanFriendlyString(s.BiggestDrawdown.MaxDrawdown.Highest.Value, 8, ".", ","))
		log.Infof(common.Statistics, "Highest Price Time: %v", s.BiggestDrawdown.MaxDrawdown.Highest.Time)
		log.Infof(common.Statistics, "Lowest Price: %s", convert.DecimalToHumanFriendlyString(s.BiggestDrawdown.MaxDrawdown.Lowest.Value, 8, ".", ","))
		log.Infof(common.Statistics, "Lowest Price Time: %v", s.BiggestDrawdown.MaxDrawdown.Lowest.Time)
		log.Infof(common.Statistics, "Calculated Drawdown: %s%%", convert.DecimalToHumanFriendlyString(s.BiggestDrawdown.MaxDrawdown.DrawdownPercent, 2, ".", ","))
		log.Infof(common.Statistics, "Difference: %s", convert.DecimalToHumanFriendlyString(s.BiggestDrawdown.MaxDrawdown.Highest.Value.Sub(s.BiggestDrawdown.MaxDrawdown.Lowest.Value), 8, ".", ","))
		log.Infof(common.Statistics, "Drawdown length: %v candles\n\n", convert.IntToHumanFriendlyString(s.BiggestDrawdown.MaxDrawdown.IntervalDuration, ","))
	}
	if s.BestMarketMovement != nil && s.BestStrategyResults != nil {
		log.Info(common.Statistics, common.ColourH4+"------------------Orders----------------------------------"+common.ColourDefault)
		log.Infof(common.Statistics, "Best performing market movement: %v %v %v %v%%", s.BestMarketMovement.Exchange, s.BestMarketMovement.Asset, s.BestMarketMovement.Pair, convert.DecimalToHumanFriendlyString(s.BestMarketMovement.MarketMovement, 2, ".", ","))
		log.Infof(common.Statistics, "Best performing strategy movement: %v %v %v %v%%\n\n", s.BestStrategyResults.Exchange, s.BestStrategyResults.Asset, s.BestStrategyResults.Pair, convert.DecimalToHumanFriendlyString(s.BestStrategyResults.StrategyMovement, 2, ".", ","))
	}
}

// PrintAllEventsChronologically outputs all event details in the CMD
// rather than separated by exchange, asset and currency pair, it's
// grouped by time to allow a clearer picture of events
func (s *Statistic) PrintAllEventsChronologically() {
	var results []eventOutputHolder
	log.Info(common.Statistics, common.ColourH1+"------------------Events-------------------------------------"+common.ColourDefault)
	var errs gctcommon.Errors
	colour := common.ColourDefault
	for exch, x := range s.ExchangeAssetPairStatistics {
		for a, y := range x {
			for pair, currencyStatistic := range y {
				for i := range currencyStatistic.Events {
					switch {
					case currencyStatistic.Events[i].FillEvent != nil:
						direction := currencyStatistic.Events[i].FillEvent.GetDirection()
						if direction == order.CouldNotBuy ||
							direction == order.CouldNotSell ||
							direction == order.MissingData ||
							direction == order.DoNothing ||
							direction == order.TransferredFunds ||
							direction == order.UnknownSide {
							if direction == order.DoNothing {
								colour = common.ColourDarkGrey
							}
							msg := fmt.Sprintf(colour+
								"%v %v%v%v| Price: %v\tDirection: %v",
								currencyStatistic.Events[i].FillEvent.GetTime().Format(gctcommon.SimpleTimeFormat),
								fSIL(exch, limit12),
								fSIL(a.String(), limit10),
								fSIL(currencyStatistic.Events[i].FillEvent.Pair().String(), limit14),
								currencyStatistic.Events[i].FillEvent.GetClosePrice().Round(8),
								currencyStatistic.Events[i].FillEvent.GetDirection())
							msg = addReason(currencyStatistic.Events[i].FillEvent.GetConcatReasons(), msg)
							msg += common.ColourDefault
							results = addEventOutputToTime(results, currencyStatistic.Events[i].FillEvent.GetTime(), msg)
						} else {
							// successful order!
							colour = common.ColourSuccess
							if currencyStatistic.Events[i].FillEvent.IsLiquidated() {
								colour = common.ColourError
							}
							msg := fmt.Sprintf(colour+
								"%v %v%v%v| Price: %v\tDirection %v\tOrder placed: Amount: %v\tFee: %v\tTotal: %v",
								currencyStatistic.Events[i].FillEvent.GetTime().Format(gctcommon.SimpleTimeFormat),
								fSIL(exch, limit12),
								fSIL(a.String(), limit10),
								fSIL(currencyStatistic.Events[i].FillEvent.Pair().String(), limit14),
								currencyStatistic.Events[i].FillEvent.GetPurchasePrice().Round(8),
								currencyStatistic.Events[i].FillEvent.GetDirection(),
								currencyStatistic.Events[i].FillEvent.GetAmount().Round(8),
								currencyStatistic.Events[i].FillEvent.GetExchangeFee(),
								currencyStatistic.Events[i].FillEvent.GetTotal().Round(8))
							msg = addReason(currencyStatistic.Events[i].FillEvent.GetConcatReasons(), msg)
							msg += common.ColourDefault
							results = addEventOutputToTime(results, currencyStatistic.Events[i].FillEvent.GetTime(), msg)
						}
					case currencyStatistic.Events[i].SignalEvent != nil:
						msg := fmt.Sprintf("%v %v%v%v| Price: $%v",
							currencyStatistic.Events[i].SignalEvent.GetTime().Format(gctcommon.SimpleTimeFormat),
							fSIL(exch, limit12),
							fSIL(a.String(), limit10),
							fSIL(currencyStatistic.Events[i].SignalEvent.Pair().String(), limit14),
							currencyStatistic.Events[i].SignalEvent.GetClosePrice().Round(8))
						msg = addReason(currencyStatistic.Events[i].SignalEvent.GetConcatReasons(), msg)
						msg += common.ColourDefault
						results = addEventOutputToTime(results, currencyStatistic.Events[i].SignalEvent.GetTime(), msg)
					case currencyStatistic.Events[i].DataEvent != nil:
						msg := fmt.Sprintf("%v %v%v%v| Price: $%v",
							currencyStatistic.Events[i].DataEvent.GetTime().Format(gctcommon.SimpleTimeFormat),
							fSIL(exch, limit12),
							fSIL(a.String(), limit10),
							fSIL(currencyStatistic.Events[i].DataEvent.Pair().String(), limit14),
							currencyStatistic.Events[i].DataEvent.GetClosePrice().Round(8))
						msg = addReason(currencyStatistic.Events[i].DataEvent.GetConcatReasons(), msg)
						msg += common.ColourDefault
						results = addEventOutputToTime(results, currencyStatistic.Events[i].DataEvent.GetTime(), msg)
					default:
						errs = append(errs, fmt.Errorf(common.ColourError+"%v%v%v unexpected data received %+v"+common.ColourDefault, exch, a, fSIL(pair.String(), limit14), currencyStatistic.Events[i]))
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		b1 := results[i]
		b2 := results[j]
		return b1.Time.Before(b2.Time)
	})
	for i := range results {
		for j := range results[i].Events {
			log.Info(common.Statistics, results[i].Events[j])
		}
	}
	if len(errs) > 0 {
		log.Info(common.Statistics, common.ColourError+"------------------Errors-------------------------------------"+common.ColourDefault)
		for i := range errs {
			log.Error(common.Statistics, errs[i].Error())
		}
	}
}

// PrintResults outputs all calculated statistics to the command line
func (c *CurrencyPairStatistic) PrintResults(e string, a asset.Item, p currency.Pair, usingExchangeLevelFunding bool) {
	var errs gctcommon.Errors
	sort.Slice(c.Events, func(i, j int) bool {
		return c.Events[i].Time.Before(c.Events[j].Time)
	})
	last := c.Events[len(c.Events)-1]
	first := c.Events[0]
	c.StartingClosePrice.Value = first.DataEvent.GetClosePrice()
	c.StartingClosePrice.Time = first.Time
	c.EndingClosePrice.Value = last.DataEvent.GetClosePrice()
	c.EndingClosePrice.Time = last.Time
	c.TotalOrders = c.BuyOrders + c.SellOrders + c.ShortOrders + c.LongOrders
	last.Holdings.TotalValueLost = last.Holdings.TotalValueLostToSlippage.Add(last.Holdings.TotalValueLostToVolumeSizing)
	sep := fmt.Sprintf("%v %v %v |\t", fSIL(e, limit12), fSIL(a.String(), limit10), fSIL(p.String(), limit14))
	currStr := fmt.Sprintf(common.ColourH1+"------------------Stats for %v %v %v------------------------------------------------------"+common.ColourDefault, e, a, p)
	log.Infof(common.CurrencyStatistics, currStr[:70])
	if a.IsFutures() {
		log.Infof(common.CurrencyStatistics, "%s Long orders: %s", sep, convert.IntToHumanFriendlyString(c.LongOrders, ","))
		log.Infof(common.CurrencyStatistics, "%s Short orders: %s", sep, convert.IntToHumanFriendlyString(c.ShortOrders, ","))
		log.Infof(common.CurrencyStatistics, "%s Highest Unrealised PNL: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.HighestUnrealisedPNL.Value, 8, ".", ","), c.HighestUnrealisedPNL.Time)
		log.Infof(common.CurrencyStatistics, "%s Lowest Unrealised PNL: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.LowestUnrealisedPNL.Value, 8, ".", ","), c.LowestUnrealisedPNL.Time)
		log.Infof(common.CurrencyStatistics, "%s Highest Realised PNL: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.HighestRealisedPNL.Value, 8, ".", ","), c.HighestRealisedPNL.Time)
		log.Infof(common.CurrencyStatistics, "%s Lowest Realised PNL: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.LowestRealisedPNL.Value, 8, ".", ","), c.LowestRealisedPNL.Time)
		log.Infof(common.CurrencyStatistics, "%s Highest committed funds: %s %s at %v", sep, convert.DecimalToHumanFriendlyString(c.HighestCommittedFunds.Value, 8, ".", ","), c.UnderlyingPair.Quote, c.HighestCommittedFunds.Time)
	} else {
		log.Infof(common.CurrencyStatistics, "%s Buy orders: %s", sep, convert.IntToHumanFriendlyString(c.BuyOrders, ","))
		log.Infof(common.CurrencyStatistics, "%s Buy amount: %s %s", sep, convert.DecimalToHumanFriendlyString(last.Holdings.BoughtAmount, 8, ".", ","), last.Holdings.Pair.Base)
		log.Infof(common.CurrencyStatistics, "%s Sell orders: %s", sep, convert.IntToHumanFriendlyString(c.SellOrders, ","))
		log.Infof(common.CurrencyStatistics, "%s Sell amount: %s %s", sep, convert.DecimalToHumanFriendlyString(last.Holdings.SoldAmount, 8, ".", ","), last.Holdings.Pair.Base)
		log.Infof(common.CurrencyStatistics, "%s Highest committed funds: %s %s at %v", sep, convert.DecimalToHumanFriendlyString(c.HighestCommittedFunds.Value, 8, ".", ","), last.Holdings.Pair.Quote, c.HighestCommittedFunds.Time)
	}

	log.Infof(common.CurrencyStatistics, "%s Total orders: %s", sep, convert.IntToHumanFriendlyString(c.TotalOrders, ","))

	log.Info(common.CurrencyStatistics, common.ColourH2+"------------------Max Drawdown-------------------------------"+common.ColourDefault)
	log.Infof(common.CurrencyStatistics, "%s Highest Price of drawdown: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.MaxDrawdown.Highest.Value, 8, ".", ","), c.MaxDrawdown.Highest.Time)
	log.Infof(common.CurrencyStatistics, "%s Lowest Price of drawdown: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.MaxDrawdown.Lowest.Value, 8, ".", ","), c.MaxDrawdown.Lowest.Time)
	log.Infof(common.CurrencyStatistics, "%s Calculated Drawdown: %s%%", sep, convert.DecimalToHumanFriendlyString(c.MaxDrawdown.DrawdownPercent, 8, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Difference: %s", sep, convert.DecimalToHumanFriendlyString(c.MaxDrawdown.Highest.Value.Sub(c.MaxDrawdown.Lowest.Value), 2, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Drawdown length: %s", sep, convert.IntToHumanFriendlyString(c.MaxDrawdown.IntervalDuration, ","))
	if !usingExchangeLevelFunding {
		log.Info(common.CurrencyStatistics, common.ColourH2+"------------------Ratios------------------------------------------------"+common.ColourDefault)
		log.Info(common.CurrencyStatistics, common.ColourH3+"------------------Rates-------------------------------------------------"+common.ColourDefault)
		log.Infof(common.CurrencyStatistics, "%s Compound Annual Growth Rate: %s", sep, convert.DecimalToHumanFriendlyString(c.CompoundAnnualGrowthRate, 2, ".", ","))
		log.Info(common.CurrencyStatistics, common.ColourH4+"------------------Arithmetic--------------------------------------------"+common.ColourDefault)
		if c.ShowMissingDataWarning {
			log.Infoln(common.CurrencyStatistics, "Missing data was detected during this backtesting run")
			log.Infoln(common.CurrencyStatistics, "Ratio calculations will be skewed")
		}
		log.Infof(common.CurrencyStatistics, "%s Sharpe ratio: %v", sep, c.ArithmeticRatios.SharpeRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Sortino ratio: %v", sep, c.ArithmeticRatios.SortinoRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Information ratio: %v", sep, c.ArithmeticRatios.InformationRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Calmar ratio: %v", sep, c.ArithmeticRatios.CalmarRatio.Round(4))

		log.Info(common.CurrencyStatistics, common.ColourH4+"------------------Geometric--------------------------------------------"+common.ColourDefault)
		if c.ShowMissingDataWarning {
			log.Infoln(common.CurrencyStatistics, "Missing data was detected during this backtesting run")
			log.Infoln(common.CurrencyStatistics, "Ratio calculations will be skewed")
		}
		log.Infof(common.CurrencyStatistics, "%s Sharpe ratio: %v", sep, c.GeometricRatios.SharpeRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Sortino ratio: %v", sep, c.GeometricRatios.SortinoRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Information ratio: %v", sep, c.GeometricRatios.InformationRatio.Round(4))
		log.Infof(common.CurrencyStatistics, "%s Calmar ratio: %v", sep, c.GeometricRatios.CalmarRatio.Round(4))
	}

	log.Info(common.CurrencyStatistics, common.ColourH2+"------------------Results------------------------------------"+common.ColourDefault)
	log.Infof(common.CurrencyStatistics, "%s Starting Close Price: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.StartingClosePrice.Value, 8, ".", ","), c.StartingClosePrice.Time)
	log.Infof(common.CurrencyStatistics, "%s Finishing Close Price: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.EndingClosePrice.Value, 8, ".", ","), c.EndingClosePrice.Time)
	log.Infof(common.CurrencyStatistics, "%s Lowest Close Price: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.LowestClosePrice.Value, 8, ".", ","), c.LowestClosePrice.Time)
	log.Infof(common.CurrencyStatistics, "%s Highest Close Price: %s at %v", sep, convert.DecimalToHumanFriendlyString(c.HighestClosePrice.Value, 8, ".", ","), c.HighestClosePrice.Time)

	log.Infof(common.CurrencyStatistics, "%s Market movement: %s%%", sep, convert.DecimalToHumanFriendlyString(c.MarketMovement, 2, ".", ","))
	if !usingExchangeLevelFunding {
		log.Infof(common.CurrencyStatistics, "%s Strategy movement: %s%%", sep, convert.DecimalToHumanFriendlyString(c.StrategyMovement, 2, ".", ","))
		log.Infof(common.CurrencyStatistics, "%s Did it beat the market: %v", sep, c.StrategyMovement.GreaterThan(c.MarketMovement))
	}

	log.Infof(common.CurrencyStatistics, "%s Value lost to volume sizing: %s", sep, convert.DecimalToHumanFriendlyString(c.TotalValueLostToVolumeSizing, 2, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Value lost to slippage: %s", sep, convert.DecimalToHumanFriendlyString(c.TotalValueLostToSlippage, 2, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Total Value lost: %s", sep, convert.DecimalToHumanFriendlyString(c.TotalValueLost, 2, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Total Fees: %s", sep, convert.DecimalToHumanFriendlyString(c.TotalFees, 8, ".", ","))
	log.Infof(common.CurrencyStatistics, "%s Final holdings value: %s", sep, convert.DecimalToHumanFriendlyString(c.TotalAssetValue, 8, ".", ","))
	if !usingExchangeLevelFunding {
		// the following have no direct translation to individual exchange level funds as they
		// combine base and quote values
		log.Infof(common.CurrencyStatistics, "%s Final funds: %s", sep, convert.DecimalToHumanFriendlyString(last.Holdings.QuoteSize, 8, ".", ","))
		log.Infof(common.CurrencyStatistics, "%s Final holdings: %s", sep, convert.DecimalToHumanFriendlyString(last.Holdings.BaseSize, 8, ".", ","))
		log.Infof(common.CurrencyStatistics, "%s Final total value: %s", sep, convert.DecimalToHumanFriendlyString(last.Holdings.TotalValue, 8, ".", ","))
	}

	if last.PNL != nil {
		var unrealised, realised portfolio.BasicPNLResult
		unrealised = last.PNL.GetUnrealisedPNL()
		realised = last.PNL.GetRealisedPNL()
		log.Infof(common.CurrencyStatistics, "%s Final Unrealised PNL: %s", sep, convert.DecimalToHumanFriendlyString(unrealised.PNL, 8, ".", ","))
		log.Infof(common.CurrencyStatistics, "%s Final Realised PNL: %s", sep, convert.DecimalToHumanFriendlyString(realised.PNL, 8, ".", ","))
	}
	if len(errs) > 0 {
		log.Info(common.CurrencyStatistics, common.ColourError+"------------------Errors-------------------------------------"+common.ColourDefault)
		for i := range errs {
			log.Error(common.CurrencyStatistics, errs[i].Error())
		}
	}
}

// PrintResults outputs all calculated funding statistics to the command line
func (f *FundingStatistics) PrintResults(wasAnyDataMissing bool) error {
	if f.Report == nil {
		return fmt.Errorf("%w requires report to be generated", common.ErrNilArguments)
	}
	var spotResults, futuresResults []FundingItemStatistics
	for i := range f.Items {
		if f.Items[i].ReportItem.Asset.IsFutures() {
			futuresResults = append(futuresResults, f.Items[i])
		} else {
			spotResults = append(spotResults, f.Items[i])
		}
	}
	if len(spotResults) > 0 || len(futuresResults) > 0 {
		log.Info(common.FundingStatistics, common.ColourH1+"------------------Funding------------------------------------"+common.ColourDefault)
	}
	if len(spotResults) > 0 {
		log.Info(common.FundingStatistics, common.ColourH2+"------------------Funding Spot Item Results------------------"+common.ColourDefault)
		for i := range spotResults {
			sep := fmt.Sprintf("%v%v%v| ", fSIL(spotResults[i].ReportItem.Exchange, limit12), fSIL(spotResults[i].ReportItem.Asset.String(), limit10), fSIL(spotResults[i].ReportItem.Currency.String(), limit14))
			if !spotResults[i].ReportItem.PairedWith.IsEmpty() {
				log.Infof(common.FundingStatistics, "%s Paired with: %v", sep, spotResults[i].ReportItem.PairedWith)
			}
			log.Infof(common.FundingStatistics, "%s Initial funds: %s", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.InitialFunds, 8, ".", ","))
			log.Infof(common.FundingStatistics, "%s Final funds: %s", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.FinalFunds, 8, ".", ","))

			if !f.Report.DisableUSDTracking && f.Report.UsingExchangeLevelFunding {
				log.Infof(common.FundingStatistics, "%s Initial funds in USD: $%s", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.USDInitialFunds, 2, ".", ","))
				log.Infof(common.FundingStatistics, "%s Final funds in USD: $%s", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.USDFinalFunds, 2, ".", ","))
			}
			if spotResults[i].ReportItem.ShowInfinite {
				log.Infof(common.FundingStatistics, "%s Difference: ∞%%", sep)
			} else {
				log.Infof(common.FundingStatistics, "%s Difference: %s%%", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.Difference, 8, ".", ","))
			}
			if spotResults[i].ReportItem.TransferFee.GreaterThan(decimal.Zero) {
				log.Infof(common.FundingStatistics, "%s Transfer fee: %s", sep, convert.DecimalToHumanFriendlyString(spotResults[i].ReportItem.TransferFee, 8, ".", ","))
			}
			if i != len(spotResults)-1 {
				log.Info(common.FundingStatistics, "")
			}
		}
	}
	if len(futuresResults) > 0 {
		log.Info(common.FundingStatistics, common.ColourH2+"------------------Funding Futures Item Results---------------"+common.ColourDefault)
		for i := range futuresResults {
			sep := fmt.Sprintf("%v%v%v| ", fSIL(futuresResults[i].ReportItem.Exchange, limit12), fSIL(futuresResults[i].ReportItem.Asset.String(), limit10), fSIL(futuresResults[i].ReportItem.Currency.String(), limit14))
			log.Infof(common.FundingStatistics, "%s Is Collateral: %v", sep, futuresResults[i].IsCollateral)
			if futuresResults[i].IsCollateral {
				log.Infof(common.FundingStatistics, "%s Initial Collateral: %v %v at %v", sep, futuresResults[i].InitialCollateral.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].InitialCollateral.Time)
				log.Infof(common.FundingStatistics, "%s Final Collateral: %v %v at %v", sep, futuresResults[i].FinalCollateral.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].FinalCollateral.Time)
				log.Infof(common.FundingStatistics, "%s Lowest Collateral: %v %v at %v", sep, futuresResults[i].LowestCollateral.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].LowestCollateral.Time)
				log.Infof(common.FundingStatistics, "%s Highest Collateral: %v %v at %v", sep, futuresResults[i].HighestCollateral.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].HighestCollateral.Time)
			} else {
				if !futuresResults[i].ReportItem.PairedWith.IsEmpty() {
					log.Infof(common.FundingStatistics, "%s Collateral currency: %v", sep, futuresResults[i].ReportItem.PairedWith)
				}
				log.Infof(common.FundingStatistics, "%s Lowest Contract Holdings: %v %v at %v", sep, futuresResults[i].LowestHoldings.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].LowestHoldings.Time)
				log.Infof(common.FundingStatistics, "%s Highest Contract Holdings: %v %v at %v", sep, futuresResults[i].HighestHoldings.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].HighestHoldings.Time)
				log.Infof(common.FundingStatistics, "%s Initial Contract Holdings: %v %v at %v", sep, futuresResults[i].InitialHoldings.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].InitialHoldings.Time)
				log.Infof(common.FundingStatistics, "%s Final Contract Holdings: %v %v at %v", sep, futuresResults[i].FinalHoldings.Value, futuresResults[i].ReportItem.Currency, futuresResults[i].FinalHoldings.Time)
			}
			if i != len(futuresResults)-1 {
				log.Info(common.FundingStatistics, "")
			}
		}
	}
	if f.Report.DisableUSDTracking {
		return nil
	}
	log.Info(common.FundingStatistics, common.ColourH2+"------------------USD Tracking Totals------------------------"+common.ColourDefault)
	sep := "USD Tracking Total |\t"

	log.Infof(common.FundingStatistics, "%s Initial value: $%s", sep, convert.DecimalToHumanFriendlyString(f.Report.InitialFunds, 8, ".", ","))
	log.Infof(common.FundingStatistics, "%s Final value: $%s", sep, convert.DecimalToHumanFriendlyString(f.Report.FinalFunds, 8, ".", ","))
	log.Infof(common.FundingStatistics, "%s Benchmark Market Movement: %s%%", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.BenchmarkMarketMovement, 8, ".", ","))
	log.Infof(common.FundingStatistics, "%s Strategy Movement: %s%%", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.StrategyMovement, 8, ".", ","))
	log.Infof(common.FundingStatistics, "%s Did strategy make a profit: %v", sep, f.TotalUSDStatistics.DidStrategyMakeProfit)
	log.Infof(common.FundingStatistics, "%s Did strategy beat the benchmark: %v", sep, f.TotalUSDStatistics.DidStrategyBeatTheMarket)
	log.Infof(common.FundingStatistics, "%s Highest funds: $%s at %v", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.HighestHoldingValue.Value, 8, ".", ","), f.TotalUSDStatistics.HighestHoldingValue.Time)
	log.Infof(common.FundingStatistics, "%s Lowest funds: $%s at %v", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.LowestHoldingValue.Value, 8, ".", ","), f.TotalUSDStatistics.LowestHoldingValue.Time)

	log.Info(common.FundingStatistics, common.ColourH3+"------------------Ratios------------------------------------------------"+common.ColourDefault)
	log.Info(common.FundingStatistics, common.ColourH4+"------------------Rates-------------------------------------------------"+common.ColourDefault)
	log.Infof(common.FundingStatistics, "%s Risk free rate: %s%%", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.RiskFreeRate.Mul(decimal.NewFromInt(100)), 2, ".", ","))
	log.Infof(common.FundingStatistics, "%s Compound Annual Growth Rate: %v%%", sep, convert.DecimalToHumanFriendlyString(f.TotalUSDStatistics.CompoundAnnualGrowthRate, 8, ".", ","))
	if f.TotalUSDStatistics.ArithmeticRatios == nil || f.TotalUSDStatistics.GeometricRatios == nil {
		return fmt.Errorf("%w missing ratio calculations", common.ErrNilArguments)
	}
	log.Info(common.FundingStatistics, common.ColourH4+"------------------Arithmetic--------------------------------------------"+common.ColourDefault)
	if wasAnyDataMissing {
		log.Infoln(common.FundingStatistics, "Missing data was detected during this backtesting run")
		log.Infoln(common.FundingStatistics, "Ratio calculations will be skewed")
	}
	log.Infof(common.FundingStatistics, "%s Sharpe ratio: %v", sep, f.TotalUSDStatistics.ArithmeticRatios.SharpeRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Sortino ratio: %v", sep, f.TotalUSDStatistics.ArithmeticRatios.SortinoRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Information ratio: %v", sep, f.TotalUSDStatistics.ArithmeticRatios.InformationRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Calmar ratio: %v", sep, f.TotalUSDStatistics.ArithmeticRatios.CalmarRatio.Round(4))

	log.Info(common.FundingStatistics, common.ColourH4+"------------------Geometric--------------------------------------------"+common.ColourDefault)
	if wasAnyDataMissing {
		log.Infoln(common.FundingStatistics, "Missing data was detected during this backtesting run")
		log.Infoln(common.FundingStatistics, "Ratio calculations will be skewed")
	}
	log.Infof(common.FundingStatistics, "%s Sharpe ratio: %v", sep, f.TotalUSDStatistics.GeometricRatios.SharpeRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Sortino ratio: %v", sep, f.TotalUSDStatistics.GeometricRatios.SortinoRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Information ratio: %v", sep, f.TotalUSDStatistics.GeometricRatios.InformationRatio.Round(4))
	log.Infof(common.FundingStatistics, "%s Calmar ratio: %v\n\n", sep, f.TotalUSDStatistics.GeometricRatios.CalmarRatio.Round(4))

	return nil
}
