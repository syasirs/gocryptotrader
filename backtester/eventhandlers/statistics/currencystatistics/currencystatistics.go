package currencystatistics

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	gctcommon "github.com/thrasher-corp/gocryptotrader/common"
	gctmath "github.com/thrasher-corp/gocryptotrader/common/math"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctkline "github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// CalculateResults calculates all statistics for the exchange, asset, currency pair
func (c *CurrencyStatistic) CalculateResults(f funding.IPairReader) error {
	var errs gctcommon.Errors
	var err error
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
		if c.LowestClosePrice.IsZero() || price.LessThan(c.LowestClosePrice) {
			c.LowestClosePrice = price
		}
		if price.GreaterThan(c.HighestClosePrice) {
			c.HighestClosePrice = price
		}
	}

	oneHundred := decimal.NewFromInt(100)
	c.MarketMovement = lastPrice.Sub(firstPrice).Div(firstPrice).Mul(oneHundred)
	if f.QuoteInitialFunds().GreaterThan(decimal.Zero) {
		c.StrategyMovement = last.Holdings.TotalValue.Sub(f.QuoteInitialFunds()).Div(f.QuoteInitialFunds()).Mul(oneHundred)
	}
	c.calculateHighestCommittedFunds()
	c.RiskFreeRate = last.Holdings.RiskFreeRate.Mul(oneHundred)
	returnPerCandle := make([]float64, len(c.Events))
	benchmarkRates := make([]float64, len(c.Events))

	var allDataEvents []common.DataEventHandler
	for i := range c.Events {
		returnPerCandle[i], _ = c.Events[i].Holdings.ChangeInTotalValuePercent.Float64()
		allDataEvents = append(allDataEvents, c.Events[i].DataEvent)
		if i == 0 {
			continue
		}
		if c.Events[i].SignalEvent != nil && c.Events[i].SignalEvent.GetDirection() == common.MissingData {
			c.ShowMissingDataWarning = true
		}
		benchmarkRates[i], _ = c.Events[i].DataEvent.ClosePrice().Sub(c.Events[i-1].DataEvent.ClosePrice()).Div(c.Events[i-1].DataEvent.ClosePrice()).Float64()
	}

	// remove the first entry as its zero and impacts
	// ratio calculations as no movement has been made
	benchmarkRates = benchmarkRates[1:]
	returnPerCandle = returnPerCandle[1:]

	var arithmeticBenchmarkAverage, geometricBenchmarkAverage float64
	arithmeticBenchmarkAverage, err = gctmath.ArithmeticMean(benchmarkRates)
	if err != nil {
		errs = append(errs, err)
	}
	geometricBenchmarkAverage, err = gctmath.FinancialGeometricMean(benchmarkRates)
	if err != nil {
		errs = append(errs, err)
	}

	c.MaxDrawdown = calculateMaxDrawdown(allDataEvents)
	interval := first.DataEvent.GetInterval()
	intervalsPerYear := interval.IntervalsPerYear()

	riskFreeRatePerCandle, _ := first.Holdings.RiskFreeRate.Div(decimal.NewFromFloat(intervalsPerYear)).Float64()
	riskFreeRateForPeriod := riskFreeRatePerCandle * float64(len(benchmarkRates))

	var arithmeticReturnsPerCandle, geometricReturnsPerCandle, arithmeticSharpe, arithmeticSortino,
		arithmeticInformation, arithmeticCalmar, geomSharpe, geomSortino, geomInformation, geomCalmar float64

	arithmeticReturnsPerCandle, err = gctmath.ArithmeticMean(returnPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geometricReturnsPerCandle, err = gctmath.FinancialGeometricMean(returnPerCandle)
	if err != nil {
		errs = append(errs, err)
	}

	arithmeticSharpe, err = gctmath.SharpeRatio(returnPerCandle, riskFreeRatePerCandle, arithmeticReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	arithmeticSortino, err = gctmath.SortinoRatio(returnPerCandle, riskFreeRatePerCandle, arithmeticReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	arithmeticInformation, err = gctmath.InformationRatio(returnPerCandle, benchmarkRates, arithmeticReturnsPerCandle, arithmeticBenchmarkAverage)
	if err != nil {
		errs = append(errs, err)
	}
	mxhp, _ := c.MaxDrawdown.Highest.Price.Float64()
	mdlp, _ := c.MaxDrawdown.Lowest.Price.Float64()
	arithmeticCalmar, err = gctmath.CalmarRatio(mxhp, mdlp, arithmeticReturnsPerCandle, riskFreeRateForPeriod)
	if err != nil {
		errs = append(errs, err)
	}

	c.ArithmeticRatios = Ratios{}
	if !math.IsNaN(arithmeticSharpe) {
		c.ArithmeticRatios.SharpeRatio = decimal.NewFromFloat(arithmeticSharpe)
	}
	if !math.IsNaN(arithmeticSortino) {
		c.ArithmeticRatios.SortinoRatio = decimal.NewFromFloat(arithmeticSortino)
	}
	if !math.IsNaN(arithmeticInformation) {
		c.ArithmeticRatios.InformationRatio = decimal.NewFromFloat(arithmeticInformation)
	}
	if !math.IsNaN(arithmeticCalmar) {
		c.ArithmeticRatios.CalmarRatio = decimal.NewFromFloat(arithmeticCalmar)
	}

	geomSharpe, err = gctmath.SharpeRatio(returnPerCandle, riskFreeRatePerCandle, geometricReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geomSortino, err = gctmath.SortinoRatio(returnPerCandle, riskFreeRatePerCandle, geometricReturnsPerCandle)
	if err != nil {
		errs = append(errs, err)
	}
	geomInformation, err = gctmath.InformationRatio(returnPerCandle, benchmarkRates, geometricReturnsPerCandle, geometricBenchmarkAverage)
	if err != nil {
		errs = append(errs, err)
	}
	geomCalmar, err = gctmath.CalmarRatio(mxhp, mdlp, geometricReturnsPerCandle, riskFreeRateForPeriod)
	if err != nil {
		errs = append(errs, err)
	}
	c.GeometricRatios = Ratios{}
	if !math.IsNaN(arithmeticSharpe) {
		c.GeometricRatios.SharpeRatio = decimal.NewFromFloat(geomSharpe)
	}
	if !math.IsNaN(arithmeticSortino) {
		c.GeometricRatios.SortinoRatio = decimal.NewFromFloat(geomSortino)
	}
	if !math.IsNaN(arithmeticInformation) {
		c.GeometricRatios.InformationRatio = decimal.NewFromFloat(geomInformation)
	}
	if !math.IsNaN(arithmeticCalmar) {
		c.GeometricRatios.CalmarRatio = decimal.NewFromFloat(geomCalmar)
	}

	lastInitial, _ := last.Holdings.QuoteInitialFunds.Float64()
	lastTotal, _ := last.Holdings.TotalValue.Float64()
	if lastInitial > 0 {
		cagr, err := gctmath.CompoundAnnualGrowthRate(
			lastInitial,
			lastTotal,
			intervalsPerYear,
			float64(len(c.Events)),
		)
		if err != nil {
			errs = append(errs, err)
		}
		if !math.IsNaN(cagr) {
			c.CompoundAnnualGrowthRate = decimal.NewFromFloat(cagr)
		}
	}
	c.IsStrategyProfitable = last.Holdings.TotalValue.GreaterThan(first.Holdings.TotalValue)
	c.DoesPerformanceBeatTheMarket = c.StrategyMovement.GreaterThan(c.MarketMovement)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// PrintResults outputs all calculated statistics to the command line
func (c *CurrencyStatistic) PrintResults(e string, a asset.Item, p currency.Pair, f funding.IPairReader) {
	var errs gctcommon.Errors
	sort.Slice(c.Events, func(i, j int) bool {
		return c.Events[i].DataEvent.GetTime().Before(c.Events[j].DataEvent.GetTime())
	})
	last := c.Events[len(c.Events)-1]
	first := c.Events[0]
	c.StartingClosePrice = first.DataEvent.ClosePrice()
	c.EndingClosePrice = last.DataEvent.ClosePrice()
	c.TotalOrders = c.BuyOrders + c.SellOrders
	last.Holdings.TotalValueLost = last.Holdings.TotalValueLostToSlippage.Add(last.Holdings.TotalValueLostToVolumeSizing)
	currStr := fmt.Sprintf("------------------Stats for %v %v %v------------------------------------------", e, a, p)
	sep := fmt.Sprintf("%v %v %v |\t", e, a, p)
	log.Infof(log.BackTester, currStr[:61])
	log.Infof(log.BackTester, "%s Initial funds: $%v", sep, f.QuoteInitialFunds())
	log.Infof(log.BackTester, "%s Highest committed funds: $%v at %v\n\n", sep, c.HighestCommittedFunds.Value.Round(roundTo8), c.HighestCommittedFunds.Time)

	log.Infof(log.BackTester, "%s Buy orders: %d", sep, c.BuyOrders)
	log.Infof(log.BackTester, "%s Buy value: $%v", sep, last.Holdings.BoughtValue.Round(roundTo8))
	log.Infof(log.BackTester, "%s Buy amount: %v %v", sep, last.Holdings.BoughtAmount.Round(roundTo8), last.Holdings.Pair.Base)
	log.Infof(log.BackTester, "%s Sell orders: %d", sep, c.SellOrders)
	log.Infof(log.BackTester, "%s Sell value: $%v", sep, last.Holdings.SoldValue.Round(roundTo8))
	log.Infof(log.BackTester, "%s Sell amount: %v %v", sep, last.Holdings.SoldAmount.Round(roundTo8), last.Holdings.Pair.Base)
	log.Infof(log.BackTester, "%s Total orders: %d\n\n", sep, c.TotalOrders)

	log.Info(log.BackTester, "------------------Max Drawdown-------------------------------")
	log.Infof(log.BackTester, "%s Highest Price of drawdown: $%v", sep, c.MaxDrawdown.Highest.Price.Round(roundTo8))
	log.Infof(log.BackTester, "%s Time of highest price of drawdown: %v", sep, c.MaxDrawdown.Highest.Time)
	log.Infof(log.BackTester, "%s Lowest Price of drawdown: $%v", sep, c.MaxDrawdown.Lowest.Price.Round(roundTo8))
	log.Infof(log.BackTester, "%s Time of lowest price of drawdown: %v", sep, c.MaxDrawdown.Lowest.Time)
	log.Infof(log.BackTester, "%s Calculated Drawdown: %v%%", sep, c.MaxDrawdown.DrawdownPercent.Round(roundTo2))
	log.Infof(log.BackTester, "%s Difference: $%v", sep, c.MaxDrawdown.Highest.Price.Sub(c.MaxDrawdown.Lowest.Price).Round(roundTo2))
	log.Infof(log.BackTester, "%s Drawdown length: %d\n\n", sep, c.MaxDrawdown.IntervalDuration)

	log.Info(log.BackTester, "------------------Rates-------------------------------------------------")
	log.Infof(log.BackTester, "%s Risk free rate: %v%%", sep, c.RiskFreeRate.Round(roundTo2))
	log.Infof(log.BackTester, "%s Compound Annual Growth Rate: %v\n\n", sep, c.CompoundAnnualGrowthRate.Round(roundTo2))

	log.Info(log.BackTester, "------------------Arithmetic Ratios-------------------------------------")
	if c.ShowMissingDataWarning {
		log.Infoln(log.BackTester, "Missing data was detected during this backtesting run")
		log.Infoln(log.BackTester, "Ratio calculations will be skewed")
	}
	log.Infof(log.BackTester, "%s Sharpe ratio: %v", sep, c.ArithmeticRatios.SharpeRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Sortino ratio: %v", sep, c.ArithmeticRatios.SortinoRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Information ratio: %v", sep, c.ArithmeticRatios.InformationRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Calmar ratio: %v\n\n", sep, c.ArithmeticRatios.CalmarRatio.Round(roundTo2))

	log.Info(log.BackTester, "------------------Geometric Ratios-------------------------------------")
	if c.ShowMissingDataWarning {
		log.Infoln(log.BackTester, "Missing data was detected during this backtesting run")
		log.Infoln(log.BackTester, "Ratio calculations will be skewed")
	}
	log.Infof(log.BackTester, "%s Sharpe ratio: %v", sep, c.GeometricRatios.SharpeRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Sortino ratio: %v", sep, c.GeometricRatios.SortinoRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Information ratio: %v", sep, c.GeometricRatios.InformationRatio.Round(roundTo2))
	log.Infof(log.BackTester, "%s Calmar ratio: %v\n\n", sep, c.GeometricRatios.CalmarRatio.Round(roundTo2))

	log.Info(log.BackTester, "------------------Results------------------------------------")
	log.Infof(log.BackTester, "%s Starting Close Price: $%v", sep, c.StartingClosePrice.Round(roundTo8))
	log.Infof(log.BackTester, "%s Finishing Close Price: $%v", sep, c.EndingClosePrice.Round(roundTo8))
	log.Infof(log.BackTester, "%s Lowest Close Price: $%v", sep, c.LowestClosePrice.Round(roundTo8))
	log.Infof(log.BackTester, "%s Highest Close Price: $%v", sep, c.HighestClosePrice.Round(roundTo8))

	log.Infof(log.BackTester, "%s Market movement: %v%%", sep, c.MarketMovement.Round(roundTo2))
	log.Infof(log.BackTester, "%s Strategy movement: %v%%", sep, c.StrategyMovement.Round(roundTo2))
	log.Infof(log.BackTester, "%s Did it beat the market: %v", sep, c.StrategyMovement.GreaterThan(c.MarketMovement))

	log.Infof(log.BackTester, "%s Value lost to volume sizing: $%v", sep, last.Holdings.TotalValueLostToVolumeSizing.Round(roundTo2))
	log.Infof(log.BackTester, "%s Value lost to slippage: $%v", sep, last.Holdings.TotalValueLostToSlippage.Round(roundTo2))
	log.Infof(log.BackTester, "%s Total Value lost: $%v", sep, last.Holdings.TotalValueLost.Round(roundTo2))
	log.Infof(log.BackTester, "%s Total Fees: $%v\n\n", sep, last.Holdings.TotalFees.Round(roundTo8))

	log.Infof(log.BackTester, "%s Final funds: $%v", sep, last.Holdings.QuoteSize.Round(roundTo8))
	log.Infof(log.BackTester, "%s Final holdings: %v", sep, last.Holdings.BaseSize.Round(roundTo8))
	log.Infof(log.BackTester, "%s Final holdings value: $%v", sep, last.Holdings.BaseValue.Round(roundTo8))
	log.Infof(log.BackTester, "%s Final total value: $%v\n\n", sep, last.Holdings.TotalValue.Round(roundTo8))

	if len(errs) > 0 {
		log.Info(log.BackTester, "------------------Errors-------------------------------------")
		for i := range errs {
			log.Info(log.BackTester, errs[i].Error())
		}
	}
}

func calculateMaxDrawdown(closePrices []common.DataEventHandler) Swing {
	var lowestPrice, highestPrice decimal.Decimal
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
		if lowestPrice.GreaterThan(currLow) && !currLow.IsZero() {
			lowestPrice = currLow
			lowestTime = currTime
		}
		if highestPrice.LessThan(currHigh) && highestPrice.IsPositive() {
			if lowestTime.Equal(highestTime) {
				// create distinction if the greatest drawdown occurs within the same candle
				lowestTime = lowestTime.Add((time.Hour * 23) + (time.Minute * 59) + (time.Second * 59))
			}
			intervals, err := gctkline.CalculateCandleDateRanges(highestTime, lowestTime, closePrices[i].GetInterval(), 0)
			if err != nil {
				log.Error(log.BackTester, err)
				continue
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
				DrawdownPercent:  lowestPrice.Sub(highestPrice).Div(highestPrice).Mul(decimal.NewFromInt(100)),
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
		if lowestTime.Equal(highestTime) {
			// create distinction if the greatest drawdown occurs within the same candle
			lowestTime = lowestTime.Add((time.Hour * 23) + (time.Minute * 59) + (time.Second * 59))
		}
		intervals, err := gctkline.CalculateCandleDateRanges(highestTime, lowestTime, closePrices[0].GetInterval(), 0)
		if err != nil {
			log.Error(log.BackTester, err)
		}
		drawdownPercent := decimal.Zero
		if highestPrice.GreaterThan(decimal.Zero) {
			drawdownPercent = lowestPrice.Sub(highestPrice).Div(highestPrice).Mul(decimal.NewFromInt(100))
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
		if swings[i].DrawdownPercent.LessThan(maxDrawdown.DrawdownPercent) {
			// drawdowns are negative
			maxDrawdown = swings[i]
		}
	}

	return maxDrawdown
}

func (c *CurrencyStatistic) calculateHighestCommittedFunds() {
	for i := range c.Events {
		if c.Events[i].Holdings.BaseValue.GreaterThan(c.HighestCommittedFunds.Value) {
			c.HighestCommittedFunds.Value = c.Events[i].Holdings.BaseValue
			c.HighestCommittedFunds.Time = c.Events[i].Holdings.Timestamp
		}
	}
}
