package currencystatistics

import (
	"fmt"
	"sort"
	"time"

	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	gctcommon "github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/math"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctkline "github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// CalculateResults calculates all statistics for the exchange, asset, currency pair
func (c *CurrencyStatistic) CalculateResults() error {
	var errs gctcommon.Errors
	first := c.Events[0]
	firstPrice := first.DataEvent.ClosePrice()
	last := c.Events[len(c.Events)-1]
	lastPrice := last.DataEvent.ClosePrice()
	for i := range last.Transactions.Orders {
		if last.Transactions.Orders[i].Side == gctorder.Buy {
			c.BuyOrders++
		} else if last.Transactions.Orders[i].Side == gctorder.Sell {
			c.SellOrders++
		}
	}
	for i := range c.Events {
		price := c.Events[i].DataEvent.ClosePrice()
		if c.LowestClosePrice == 0 || price < c.LowestClosePrice {
			c.LowestClosePrice = price
		}
		if price > c.HighestClosePrice {
			c.HighestClosePrice = price
		}
	}
	c.MarketMovement = ((lastPrice - firstPrice) / firstPrice) * 100
	c.StrategyMovement = ((last.Holdings.TotalValue - last.Holdings.InitialFunds) / last.Holdings.InitialFunds) * 100
	c.calculateHighestCommittedFunds()
	c.RiskFreeRate = last.Holdings.RiskFreeRate * 100
	returnPerCandle := make([]float64, len(c.Events))
	benchmarkRates := make([]float64, len(c.Events))

	var allDataEvents []common.DataEventHandler
	for i := range c.Events {
		returnPerCandle[i] = c.Events[i].Holdings.ChangeInTotalValuePercent
		allDataEvents = append(allDataEvents, c.Events[i].DataEvent)
		if i == 0 {
			continue
		}
		if c.Events[i].SignalEvent != nil && c.Events[i].SignalEvent.GetDirection() == common.MissingData {
			c.ShowMissingDataWarning = true
		}
		benchmarkRates[i] = (c.Events[i].DataEvent.ClosePrice() - c.Events[i-1].DataEvent.ClosePrice()) / c.Events[i-1].DataEvent.ClosePrice()
	}

	// remove the first entry as its zero and impacts
	// ratio calculations as no movement has been made
	benchmarkRates = benchmarkRates[1:]
	returnPerCandle = returnPerCandle[1:]

	var arithmeticBenchmarkAverage, geometricBenchmarkAverage float64
	var err error
	arithmeticBenchmarkAverage, err = math.ArithmeticMean(benchmarkRates)
	if err != nil {
		errs = append(errs, err)
	}
	geometricBenchmarkAverage, err = math.FinancialGeometricMean(benchmarkRates)
	if err != nil {
		errs = append(errs, err)
	}

	c.MaxDrawdown = calculateMaxDrawdown(allDataEvents)
	interval := first.DataEvent.GetInterval()
	intervalsPerYear := interval.IntervalsPerYear()

	riskFreeRatePerCandle := first.Holdings.RiskFreeRate / intervalsPerYear
	riskFreeRateForPeriod := riskFreeRatePerCandle * float64(len(benchmarkRates))

	var arithmeticReturnsPerCandle, geometricReturnsPerCandle, arithmeticSharpe, arithmeticSortino,
		arithmeticInformation, arithmeticCalmar, geomSharpe, geomSortino, geomInformation, geomCalmar float64

	arithmeticReturnsPerCandle, err = math.ArithmeticMean(returnPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geometricReturnsPerCandle, err = math.FinancialGeometricMean(returnPerCandle)
	if err != nil {
		errs = append(errs, err)
	}

	arithmeticSharpe, err = math.SharpeRatio(returnPerCandle, riskFreeRatePerCandle, arithmeticReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	arithmeticSortino, err = math.SortinoRatio(returnPerCandle, riskFreeRatePerCandle, arithmeticReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	arithmeticInformation, err = math.InformationRatio(returnPerCandle, benchmarkRates, arithmeticReturnsPerCandle, arithmeticBenchmarkAverage)
	if err != nil {
		errs = append(errs, err)
	}
	arithmeticCalmar, err = math.CalmarRatio(c.MaxDrawdown.Highest.Price, c.MaxDrawdown.Lowest.Price, arithmeticReturnsPerCandle, riskFreeRateForPeriod)
	if err != nil {
		errs = append(errs, err)
	}
	c.ArithmeticRatios = Ratios{
		SharpeRatio:      arithmeticSharpe,
		SortinoRatio:     arithmeticSortino,
		InformationRatio: arithmeticInformation,
		CalmarRatio:      arithmeticCalmar,
	}

	geomSharpe, err = math.SharpeRatio(returnPerCandle, riskFreeRatePerCandle, geometricReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geomSortino, err = math.SortinoRatio(returnPerCandle, riskFreeRatePerCandle, geometricReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geomInformation, err = math.InformationRatio(returnPerCandle, benchmarkRates, geometricReturnsPerCandle, geometricBenchmarkAverage)
	if err != nil {
		errs = append(errs, err)
	}
	geomCalmar, err = math.CalmarRatio(c.MaxDrawdown.Highest.Price, c.MaxDrawdown.Lowest.Price, geometricReturnsPerCandle, riskFreeRateForPeriod)
	if err != nil {
		errs = append(errs, err)
	}
	c.GeometricRatios = Ratios{
		SharpeRatio:      geomSharpe,
		SortinoRatio:     geomSortino,
		InformationRatio: geomInformation,
		CalmarRatio:      geomCalmar,
	}

	c.CompoundAnnualGrowthRate, err = math.CompoundAnnualGrowthRate(
		last.Holdings.InitialFunds,
		last.Holdings.TotalValue,
		intervalsPerYear,
		float64(len(c.Events)))
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// PrintResults outputs all calculated statistics to the command line
func (c *CurrencyStatistic) PrintResults(e string, a asset.Item, p currency.Pair) {
	var errs gctcommon.Errors
	sort.Slice(c.Events, func(i, j int) bool {
		return c.Events[i].DataEvent.GetTime().Before(c.Events[j].DataEvent.GetTime())
	})
	last := c.Events[len(c.Events)-1]
	first := c.Events[0]
	c.StartingClosePrice = first.DataEvent.ClosePrice()
	c.EndingClosePrice = last.DataEvent.ClosePrice()
	c.TotalOrders = c.BuyOrders + c.SellOrders
	last.Holdings.TotalValueLost = last.Holdings.TotalValueLostToSlippage + last.Holdings.TotalValueLostToVolumeSizing
	currStr := fmt.Sprintf("------------------Stats for %v %v %v------------------------------------------", e, a, p)

	log.BackTester.Infof(currStr[:61])
	log.BackTester.Infof("Initial funds: $%.2f", last.Holdings.InitialFunds)
	log.BackTester.Infof("Highest committed funds: $%.2f at %v\n\n", c.HighestCommittedFunds.Value, c.HighestCommittedFunds.Time)

	log.BackTester.Infof("Buy orders: %d", c.BuyOrders)
	log.BackTester.Infof("Buy value: $%.2f", last.Holdings.BoughtValue)
	log.BackTester.Infof("Buy amount: %.2f %v", last.Holdings.BoughtAmount, last.Holdings.Pair.Base)
	log.BackTester.Infof("Sell orders: %d", c.SellOrders)
	log.BackTester.Infof("Sell value: $%.2f", last.Holdings.SoldValue)
	log.BackTester.Infof("Sell amount: %.2f %v", last.Holdings.SoldAmount, last.Holdings.Pair.Base)
	log.BackTester.Infof("Total orders: %d\n\n", c.TotalOrders)

	log.BackTester.Info("------------------Max Drawdown-------------------------------")
	log.BackTester.Infof("Highest Price of drawdown: $%.2f", c.MaxDrawdown.Highest.Price)
	log.BackTester.Infof("Time of highest price of drawdown: %v", c.MaxDrawdown.Highest.Time)
	log.BackTester.Infof("Lowest Price of drawdown: $%.2f", c.MaxDrawdown.Lowest.Price)
	log.BackTester.Infof("Time of lowest price of drawdown: %v", c.MaxDrawdown.Lowest.Time)
	log.BackTester.Infof("Calculated Drawdown: %.2f%%", c.MaxDrawdown.DrawdownPercent)
	log.BackTester.Infof("Difference: $%.2f", c.MaxDrawdown.Highest.Price-c.MaxDrawdown.Lowest.Price)
	log.BackTester.Infof("Drawdown length: %d\n\n", c.MaxDrawdown.IntervalDuration)

	log.BackTester.Info("------------------Rates-------------------------------------------------")
	log.BackTester.Infof("Risk free rate: %.3f%%", c.RiskFreeRate)
	log.BackTester.Infof("Compound Annual Growth Rate: %.2f\n\n", c.CompoundAnnualGrowthRate)

	log.BackTester.Info("------------------Arithmetic Ratios-------------------------------------")
	if c.ShowMissingDataWarning {
		log.Infoln(log.BackTester, "Missing data was detected during this backtesting run")
		log.Infoln(log.BackTester, "Ratio calculations will be skewed")
	}
	log.BackTester.Infof("Sharpe ratio: %.2f", c.ArithmeticRatios.SharpeRatio)
	log.BackTester.Infof("Sortino ratio: %.2f", c.ArithmeticRatios.SortinoRatio)
	log.BackTester.Infof("Information ratio: %.2f", c.ArithmeticRatios.InformationRatio)
	log.BackTester.Infof("Calmar ratio: %.2f\n\n", c.ArithmeticRatios.CalmarRatio)

	log.BackTester.Info("------------------Geometric Ratios-------------------------------------")
	if c.ShowMissingDataWarning {
		log.Infoln(log.BackTester, "Missing data was detected during this backtesting run")
		log.Infoln(log.BackTester, "Ratio calculations will be skewed")
	}
	log.BackTester.Infof("Sharpe ratio: %.2f", c.GeometricRatios.SharpeRatio)
	log.BackTester.Infof("Sortino ratio: %.2f", c.GeometricRatios.SortinoRatio)
	log.BackTester.Infof("Information ratio: %.2f", c.GeometricRatios.InformationRatio)
	log.BackTester.Infof("Calmar ratio: %.2f\n\n", c.GeometricRatios.CalmarRatio)

	log.BackTester.Info("------------------Results------------------------------------")
	log.BackTester.Infof("Starting Close Price: $%.2f", c.StartingClosePrice)
	log.BackTester.Infof("Finishing Close Price: $%.2f", c.EndingClosePrice)
	log.BackTester.Infof("Lowest Close Price: $%.2f", c.LowestClosePrice)
	log.BackTester.Infof("Highest Close Price: $%.2f", c.HighestClosePrice)

	log.BackTester.Infof("Market movement: %.4f%%", c.MarketMovement)
	log.BackTester.Infof("Strategy movement: %.4f%%", c.StrategyMovement)
	log.BackTester.Infof("Did it beat the market: %v", c.StrategyMovement > c.MarketMovement)

	log.BackTester.Infof("Value lost to volume sizing: $%.2f", last.Holdings.TotalValueLostToVolumeSizing)
	log.BackTester.Infof("Value lost to slippage: $%.2f", last.Holdings.TotalValueLostToSlippage)
	log.BackTester.Infof("Total Value lost: $%.2f", last.Holdings.TotalValueLost)
	log.BackTester.Infof("Total Fees: $%.2f\n\n", last.Holdings.TotalFees)

	log.BackTester.Infof("Final funds: $%.2f", last.Holdings.RemainingFunds)
	log.BackTester.Infof("Final holdings: %.2f", last.Holdings.PositionsSize)
	log.BackTester.Infof("Final holdings value: $%.2f", last.Holdings.PositionsValue)
	log.BackTester.Infof("Final total value: $%.2f\n\n", last.Holdings.TotalValue)

	if len(errs) > 0 {
		log.BackTester.Info("------------------Errors-------------------------------------")
		for i := range errs {
			log.BackTester.Info(errs[i].Error())
		}
	}
}

func calculateMaxDrawdown(closePrices []common.DataEventHandler) Swing {
	var lowestPrice, highestPrice float64
	var lowestTime, highestTime time.Time
	var swings []Swing
	if len(closePrices) > 0 {
		lowestPrice = closePrices[0].LowPrice()
		highestPrice = closePrices[0].HighPrice()
		lowestTime = closePrices[0].GetTime()
		highestTime = closePrices[0].GetTime()
	}
	for i := range closePrices {
		currHigh := closePrices[i].HighPrice()
		currLow := closePrices[i].LowPrice()
		currTime := closePrices[i].GetTime()
		if lowestPrice > currLow && currLow != 0 {
			lowestPrice = currLow
			lowestTime = currTime
		}
		if highestPrice < currHigh && highestPrice > 0 {
			intervals := gctkline.CalculateCandleDateRanges(highestTime, lowestTime, closePrices[i].GetInterval(), 0)
			if lowestTime.Equal(highestTime) {
				// create distinction if the greatest drawdown occurs within the same candle
				lowestTime = lowestTime.Add((time.Hour * 23) + (time.Minute * 59) + (time.Second * 59))
			}
			swings = append(swings, Swing{
				Highest: Iteration{
					Time:  highestTime,
					Price: highestPrice,
				},
				Lowest: Iteration{
					Time:  lowestTime,
					Price: lowestPrice,
				},
				DrawdownPercent:  ((lowestPrice - highestPrice) / highestPrice) * 100,
				IntervalDuration: int64(len(intervals.Ranges[0].Intervals)),
			})
			// reset the drawdown
			highestPrice = currHigh
			highestTime = currTime
			lowestPrice = currLow
			lowestTime = currTime
		}
	}
	if (len(swings) > 0 && swings[len(swings)-1].Lowest.Price != closePrices[len(closePrices)-1].LowPrice()) || swings == nil {
		// need to close out the final drawdown
		intervals := gctkline.CalculateCandleDateRanges(highestTime, lowestTime, closePrices[0].GetInterval(), 0)
		drawdownPercent := 0.0
		if highestPrice > 0 {
			drawdownPercent = ((lowestPrice - highestPrice) / highestPrice) * 100
		}
		if lowestTime.Equal(highestTime) {
			// create distinction if the greatest drawdown occurs within the same candle
			lowestTime = lowestTime.Add((time.Hour * 23) + (time.Minute * 59) + (time.Second * 59))
		}
		swings = append(swings, Swing{
			Highest: Iteration{
				Time:  highestTime,
				Price: highestPrice,
			},
			Lowest: Iteration{
				Time:  lowestTime,
				Price: lowestPrice,
			},
			DrawdownPercent:  drawdownPercent,
			IntervalDuration: int64(len(intervals.Ranges[0].Intervals)),
		})
	}

	var maxDrawdown Swing
	if len(swings) > 0 {
		maxDrawdown = swings[0]
	}
	for i := range swings {
		if swings[i].DrawdownPercent < maxDrawdown.DrawdownPercent {
			// drawdowns are negative
			maxDrawdown = swings[i]
		}
	}

	return maxDrawdown
}

func (c *CurrencyStatistic) calculateHighestCommittedFunds() {
	for i := range c.Events {
		if c.Events[i].Holdings.CommittedFunds > c.HighestCommittedFunds.Value {
			c.HighestCommittedFunds.Value = c.Events[i].Holdings.CommittedFunds
			c.HighestCommittedFunds.Time = c.Events[i].Holdings.Timestamp
		}
	}
}
