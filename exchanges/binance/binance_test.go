package binance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/key"
	"github.com/thrasher-corp/gocryptotrader/core"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/collateral"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fundingrate"
	"github.com/thrasher-corp/gocryptotrader/exchanges/futures"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/margin"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	testexch "github.com/thrasher-corp/gocryptotrader/internal/testing/exchange"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// Please supply your own keys here for due diligence testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
	useTestNet              = false

	apiStreamingIsNotConnected = "API streaming is not connected"
)

var (
	b = &Binance{}
	// this lock guards against orderbook tests race
	binanceOrderBookLock = &sync.Mutex{}
	// this pair is used to ensure that endpoints match it correctly
	testPairMapping = currency.NewPair(currency.DOGE, currency.USDT)
)

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:        1,
		FeeType:       exchange.CryptocurrencyTradeFee,
		Pair:          currency.NewPair(currency.BTC, currency.LTC),
		PurchasePrice: 1,
	}
}

// getTime returns a static time for mocking endpoints, if mock is not enabled
// this will default to time now with a window size of 30 days.
// Mock details are unix seconds; start = 1577836800 and end = 1580515200
func getTime() (start, end time.Time) {
	if mockTests {
		return time.Unix(1577836800, 0), time.Unix(1580515200, 0)
	}

	tn := time.Now()
	offset := time.Hour * 24 * 6
	return tn.Add(-offset), tn
}

func TestUServerTime(t *testing.T) {
	t.Parallel()
	_, err := b.UServerTime(context.Background())
	assert.NoError(t, err)
}

func TestWrapperGetServerTime(t *testing.T) {
	t.Parallel()
	_, err := b.GetServerTime(context.Background(), asset.Empty)
	require.ErrorIs(t, err, asset.ErrNotSupported)

	st, err := b.GetServerTime(context.Background(), asset.Spot)
	assert.NoError(t, err)
	require.False(t, st.IsZero(), "expected a time")

	st, err = b.GetServerTime(context.Background(), asset.USDTMarginedFutures)
	require.NoErrorf(t, err, "received: '%v' but expected: '%v'", err, nil)

	require.False(t, st.IsZero(), "expected a time")

	st, err = b.GetServerTime(context.Background(), asset.CoinMarginedFutures)
	require.NoError(t, err)

	assert.False(t, st.IsZero(), "expected a time")
}

func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	r, err := b.UpdateTicker(context.Background(), testPairMapping, asset.Spot)
	require.NoError(t, err)
	assert.False(t, r.Pair.Base != currency.DOGE && r.Pair.Quote != currency.USDT, "invalid pair values")
	tradablePairs, err := b.FetchTradablePairs(context.Background(), asset.CoinMarginedFutures)
	require.NoError(t, err)
	require.NotEmpty(t, tradablePairs, "no tradable pairs")
	_, err = b.UpdateTicker(context.Background(), tradablePairs[0], asset.CoinMarginedFutures)
	assert.NoError(t, err)

	usdtMarginedPairs, err := b.FetchTradablePairs(context.Background(), asset.USDTMarginedFutures)
	assert.NoError(t, err)
	assert.NotEmpty(t, usdtMarginedPairs, "no pairs are enabled")
	_, err = b.UpdateTicker(context.Background(), usdtMarginedPairs[0], asset.USDTMarginedFutures)
	assert.NoError(t, err)
}

func TestUpdateTickers(t *testing.T) {
	err := b.UpdateTickers(context.Background(), asset.Spot)
	assert.NoError(t, err)
	err = b.UpdateTickers(context.Background(), asset.CoinMarginedFutures)
	assert.NoError(t, err)
	err = b.UpdateTickers(context.Background(), asset.USDTMarginedFutures)
	assert.NoError(t, err)
}

func TestUpdateOrderbook(t *testing.T) {
	t.Parallel()
	cp, err := currency.NewPairFromString("BTCUSDT")
	assert.NoError(t, err)
	_, err = b.UpdateOrderbook(context.Background(), cp, asset.Spot)
	assert.NoError(t, err)
	_, err = b.UpdateOrderbook(context.Background(), cp, asset.Margin)
	assert.NoError(t, err)
	_, err = b.UpdateOrderbook(context.Background(), cp, asset.USDTMarginedFutures)
	assert.NoError(t, err)
	cp2, err := currency.NewPairFromString("BTCUSD_PERP")
	assert.NoError(t, err)
	_, err = b.UpdateOrderbook(context.Background(), cp2, asset.CoinMarginedFutures)
	assert.NoError(t, err)
}

// USDT Margined Futures

func TestUExchangeInfo(t *testing.T) {
	t.Parallel()
	_, err := b.UExchangeInfo(context.Background())
	assert.NoError(t, err)
}

func TestUFuturesOrderbook(t *testing.T) {
	t.Parallel()
	_, err := b.UFuturesOrderbook(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 1000)
	assert.NoError(t, err)
}

func TestURecentTrades(t *testing.T) {
	t.Parallel()
	_, err := b.URecentTrades(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", 1000)
	assert.NoError(t, err)
}

func TestUCompressedTrades(t *testing.T) {
	t.Parallel()
	_, err := b.UCompressedTrades(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UCompressedTrades(context.Background(), currency.NewPair(currency.LTC, currency.USDT), "", 0, start, end)
	assert.NoError(t, err)
}

func TestUKlineData(t *testing.T) {
	t.Parallel()
	_, err := b.UKlineData(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "1d", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UKlineData(context.Background(), currency.NewPair(currency.LTC, currency.USDT), "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestUGetMarkPrice(t *testing.T) {
	t.Parallel()
	_, err := b.UGetMarkPrice(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	_, err = b.UGetMarkPrice(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestUGetFundingHistory(t *testing.T) {
	t.Parallel()
	_, err := b.UGetFundingHistory(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 1, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UGetFundingHistory(context.Background(), currency.NewPair(currency.LTC, currency.USDT), 1, start, end)
	assert.NoError(t, err)
}

func TestU24HTickerPriceChangeStats(t *testing.T) {
	t.Parallel()
	_, err := b.U24HTickerPriceChangeStats(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	_, err = b.U24HTickerPriceChangeStats(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestUSymbolPriceTicker(t *testing.T) {
	t.Parallel()
	_, err := b.USymbolPriceTicker(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	_, err = b.USymbolPriceTicker(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestUSymbolOrderbookTicker(t *testing.T) {
	t.Parallel()
	_, err := b.USymbolOrderbookTicker(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	_, err = b.USymbolOrderbookTicker(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestUOpenInterest(t *testing.T) {
	t.Parallel()
	_, err := b.UOpenInterest(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUOpenInterestStats(t *testing.T) {
	t.Parallel()
	_, err := b.UOpenInterestStats(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 1, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UOpenInterestStats(context.Background(), currency.NewPair(currency.LTC, currency.USDT), "1d", 10, start, end)
	assert.NoError(t, err)
}

func TestUTopAcccountsLongShortRatio(t *testing.T) {
	t.Parallel()
	_, err := b.UTopAcccountsLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 2, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UTopAcccountsLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 2, start, end)
	assert.NoError(t, err)
}

func TestUTopPostionsLongShortRatio(t *testing.T) {
	t.Parallel()
	_, err := b.UTopPostionsLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 3, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UTopPostionsLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "1d", 0, start, end)
	assert.NoError(t, err)
}

func TestUGlobalLongShortRatio(t *testing.T) {
	t.Parallel()
	_, err := b.UGlobalLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 3, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.UGlobalLongShortRatio(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "4h", 0, start, end)
	assert.NoError(t, err)
}

func TestUTakerBuySellVol(t *testing.T) {
	t.Parallel()
	start, end := getTime()
	_, err := b.UTakerBuySellVol(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "5m", 10, start, end)
	assert.NoError(t, err)
}

func TestUCompositeIndexInfo(t *testing.T) {
	t.Parallel()
	cp, err := currency.NewPairFromString("DEFI-USDT")
	assert.NoError(t, err)
	_, err = b.UCompositeIndexInfo(context.Background(), cp)
	assert.NoError(t, err)
	_, err = b.UCompositeIndexInfo(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestUFuturesNewOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UFuturesNewOrder(context.Background(),
		&UFuturesNewOrderRequest{
			Symbol:      currency.NewPair(currency.BTC, currency.USDT),
			Side:        "BUY",
			OrderType:   "LIMIT",
			TimeInForce: "GTC",
			Quantity:    1,
			Price:       1,
		},
	)
	assert.NoError(t, err)
}

func TestUPlaceBatchOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	var data []PlaceBatchOrderData
	var tempData PlaceBatchOrderData
	tempData.Symbol = "BTCUSDT"
	tempData.Side = "BUY"
	tempData.OrderType = "LIMIT"
	tempData.Quantity = 4
	tempData.Price = 1
	tempData.TimeInForce = "GTC"
	data = append(data, tempData)
	_, err := b.UPlaceBatchOrders(context.Background(), data)
	assert.NoError(t, err)
}

func TestUGetOrderData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UGetOrderData(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "123", "")
	assert.NoError(t, err)
}

func TestUCancelOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UCancelOrder(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "123", "")
	assert.NoError(t, err)
}

func TestUCancelAllOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UCancelAllOpenOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUCancelBatchOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UCancelBatchOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT), []string{"123"}, []string{})
	assert.NoError(t, err)
}

func TestUAutoCancelAllOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UAutoCancelAllOpenOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 30)
	assert.NoError(t, err)
}

func TestUFetchOpenOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UFetchOpenOrder(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "123", "")
	assert.NoError(t, err)
}

func TestUAllAccountOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAllAccountOpenOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUAllAccountOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAllAccountOrders(context.Background(), currency.EMPTYPAIR, 0, 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	_, err = b.UAllAccountOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 0, 5, time.Now().Add(-time.Hour*4), time.Now())
	assert.NoError(t, err)
}

func TestUAccountBalanceV2(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAccountBalanceV2(context.Background())
	assert.NoError(t, err)
}

func TestUAccountInformationV2(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAccountInformationV2(context.Background())
	assert.NoError(t, err)
}

func TestUChangeInitialLeverageRequest(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UChangeInitialLeverageRequest(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 2)
	assert.NoError(t, err)
}

func TestUChangeInitialMarginType(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.UChangeInitialMarginType(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "ISOLATED")
	assert.NoError(t, err)
}

func TestUModifyIsolatedPositionMarginReq(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UModifyIsolatedPositionMarginReq(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "LONG", "add", 5)
	assert.NoError(t, err)
}

func TestUPositionMarginChangeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UPositionMarginChangeHistory(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "add", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestUPositionsInfoV2(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UPositionsInfoV2(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUAccountTradesHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAccountTradesHistory(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestUAccountIncomeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAccountIncomeHistory(context.Background(), currency.EMPTYPAIR, "", 5, time.Now().Add(-time.Hour*48), time.Now())
	assert.NoError(t, err)
}

func TestUGetNotionalAndLeverageBrackets(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UGetNotionalAndLeverageBrackets(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUPositionsADLEstimate(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UPositionsADLEstimate(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestUAccountForcedOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UAccountForcedOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "ADL", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

// Coin Margined Futures

func TestGetFuturesExchangeInfo(t *testing.T) {
	t.Parallel()
	_, err := b.FuturesExchangeInfo(context.Background())
	assert.NoError(t, err)
}

func TestGetFuturesOrderbook(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesOrderbook(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 1000)
	assert.NoError(t, err)
}

func TestGetFuturesPublicTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesPublicTrades(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 5)
	assert.NoError(t, err)
}

func TestGetPastPublicTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetPastPublicTrades(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 5, 0)
	assert.NoError(t, err)
}

func TestGetAggregatedTradesList(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesAggregatedTradesList(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 0, 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestGetPerpsExchangeInfo(t *testing.T) {
	t.Parallel()
	_, err := b.GetPerpMarkets(context.Background())
	assert.NoError(t, err)
}

func TestGetIndexAndMarkPrice(t *testing.T) {
	t.Parallel()
	_, err := b.GetIndexAndMarkPrice(context.Background(), "", "BTCUSD")
	assert.NoError(t, err)
}

func TestGetFuturesKlineData(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesKlineData(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "1M", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)

	start, end := getTime()
	_, err = b.GetFuturesKlineData(context.Background(), currency.NewPairWithDelimiter("LTCUSD", "PERP", "_"), "5m", 5, start, end)
	assert.NoError(t, err)
}

func TestGetContinuousKlineData(t *testing.T) {
	t.Parallel()
	_, err := b.GetContinuousKlineData(context.Background(), "BTCUSD", "CURRENT_QUARTER", "1M", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetContinuousKlineData(context.Background(), "BTCUSD", "CURRENT_QUARTER", "1M", 5, start, end)
	assert.NoError(t, err)
}

func TestGetIndexPriceKlines(t *testing.T) {
	t.Parallel()
	_, err := b.GetIndexPriceKlines(context.Background(), "BTCUSD", "1M", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetIndexPriceKlines(context.Background(), "BTCUSD", "1M", 5, start, end)
	assert.NoError(t, err)
}

func TestGetFuturesSwapTickerChangeStats(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesSwapTickerChangeStats(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "")
	assert.NoError(t, err)
	_, err = b.GetFuturesSwapTickerChangeStats(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "")
	assert.NoError(t, err)
	_, err = b.GetFuturesSwapTickerChangeStats(context.Background(), currency.EMPTYPAIR, "")
	assert.NoError(t, err)
}

func TestFuturesGetFundingHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesGetFundingHistory(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.FuturesGetFundingHistory(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 50, start, end)
	assert.NoError(t, err)
}

func TestGetFuturesHistoricalTrades(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFuturesHistoricalTrades(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "", 5)
	assert.NoError(t, err)
	_, err = b.GetFuturesHistoricalTrades(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "", 0)
	assert.NoError(t, err)
}

func TestGetFuturesSymbolPriceTicker(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesSymbolPriceTicker(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "")
	assert.NoError(t, err)
}

func TestGetFuturesOrderbookTicker(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesOrderbookTicker(context.Background(), currency.EMPTYPAIR, "")
	assert.NoError(t, err)
	_, err = b.GetFuturesOrderbookTicker(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "")
	assert.NoError(t, err)
}

func TestOpenInterest(t *testing.T) {
	t.Parallel()
	_, err := b.OpenInterest(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"))
	assert.NoError(t, err)
}

func TestGetOpenInterestStats(t *testing.T) {
	t.Parallel()
	_, err := b.GetOpenInterestStats(context.Background(), "BTCUSD", "CURRENT_QUARTER", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetOpenInterestStats(context.Background(), "BTCUSD", "CURRENT_QUARTER", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestGetTraderFuturesAccountRatio(t *testing.T) {
	t.Parallel()
	_, err := b.GetTraderFuturesAccountRatio(context.Background(), "BTCUSD", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetTraderFuturesAccountRatio(context.Background(), "BTCUSD", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestGetTraderFuturesPositionsRatio(t *testing.T) {
	t.Parallel()
	_, err := b.GetTraderFuturesPositionsRatio(context.Background(), "BTCUSD", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetTraderFuturesPositionsRatio(context.Background(), "BTCUSD", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestGetMarketRatio(t *testing.T) {
	t.Parallel()
	_, err := b.GetMarketRatio(context.Background(), "BTCUSD", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetMarketRatio(context.Background(), "BTCUSD", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestGetFuturesTakerVolume(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesTakerVolume(context.Background(), "BTCUSD", "ALL", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetFuturesTakerVolume(context.Background(), "BTCUSD", "ALL", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestFuturesBasisData(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesBasisData(context.Background(), "BTCUSD", "CURRENT_QUARTER", "5m", 0, time.Time{}, time.Time{})
	assert.NoError(t, err)
	start, end := getTime()
	_, err = b.GetFuturesBasisData(context.Background(), "BTCUSD", "CURRENT_QUARTER", "5m", 0, start, end)
	assert.NoError(t, err)
}

func TestFuturesNewOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.FuturesNewOrder(
		context.Background(),
		&FuturesNewOrderRequest{
			Symbol:      currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"),
			Side:        "BUY",
			OrderType:   "LIMIT",
			TimeInForce: BinanceRequestParamsTimeGTC,
			Quantity:    1,
			Price:       1,
		},
	)
	assert.NoError(t, err)
}

func TestFuturesBatchOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	var data []PlaceBatchOrderData
	var tempData PlaceBatchOrderData
	tempData.Symbol = "BTCUSD_PERP"
	tempData.Side = "BUY"
	tempData.OrderType = "LIMIT"
	tempData.Quantity = 1
	tempData.Price = 1
	tempData.TimeInForce = "GTC"

	data = append(data, tempData)
	_, err := b.FuturesBatchOrder(context.Background(), data)
	assert.NoError(t, err)
}

func TestFuturesBatchCancelOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.FuturesBatchCancelOrders(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), []string{"123"}, []string{})
	assert.NoError(t, err)
}

func TestFuturesGetOrderData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesGetOrderData(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "123", "")
	assert.NoError(t, err)
}

func TestCancelAllOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.FuturesCancelAllOpenOrders(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"))
	assert.NoError(t, err)
}

func TestAutoCancelAllOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.AutoCancelAllOpenOrders(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 30000)
	assert.NoError(t, err)
}

func TestFuturesOpenOrderData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesOpenOrderData(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", "")
	assert.NoError(t, err)
}

func TestGetFuturesAllOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFuturesAllOpenOrders(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "")
	assert.NoError(t, err)
}

func TestGetAllFuturesOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAllFuturesOrders(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), currency.EMPTYPAIR, time.Time{}, time.Time{}, 0, 2)
	assert.NoError(t, err)
}

func TestFuturesChangeMarginType(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.FuturesChangeMarginType(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "ISOLATED")
	assert.NoError(t, err)
}

func TestGetFuturesAccountBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFuturesAccountBalance(context.Background())
	assert.NoError(t, err)
}

func TestGetFuturesAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFuturesAccountInfo(context.Background())
	assert.NoError(t, err)
}

func TestFuturesChangeInitialLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.FuturesChangeInitialLeverage(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), 5)
	assert.NoError(t, err)
}

func TestModifyIsolatedPositionMargin(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ModifyIsolatedPositionMargin(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "BOTH", "add", 5)
	assert.NoError(t, err)
}

func TestFuturesMarginChangeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesMarginChangeHistory(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "add", time.Time{}, time.Time{}, 10)
	assert.NoError(t, err)
}

func TestFuturesPositionsInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesPositionsInfo(context.Background(), "BTCUSD", "")
	assert.NoError(t, err)
}

func TestFuturesTradeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesTradeHistory(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "", time.Time{}, time.Time{}, 5, 0)
	assert.NoError(t, err)
}

func TestFuturesIncomeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesIncomeHistory(context.Background(), currency.EMPTYPAIR, "TRANSFER", time.Time{}, time.Time{}, 5)
	assert.NoError(t, err)
}

func TestFuturesForceOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesForceOrders(context.Background(), currency.EMPTYPAIR, "ADL", time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestUGetNotionalLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesNotionalBracket(context.Background(), "BTCUSD")
	assert.NoError(t, err)
	_, err = b.FuturesNotionalBracket(context.Background(), "")
	assert.NoError(t, err)
}

func TestFuturesPositionsADLEstimate(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FuturesPositionsADLEstimate(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)
}

func TestGetMarkPriceKline(t *testing.T) {
	t.Parallel()
	_, err := b.GetMarkPriceKline(context.Background(), currency.NewPairWithDelimiter("BTCUSD", "PERP", "_"), "1M", 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestGetExchangeInfo(t *testing.T) {
	t.Parallel()
	info, err := b.GetExchangeInfo(context.Background())
	require.NoError(t, err, "GetExchangeInfo must not error")
	if mockTests {
		exp := time.Date(2022, 2, 25, 3, 50, 40, int(601*time.Millisecond), time.UTC)
		assert.True(t, info.ServerTime.Time().Equal(exp), "ServerTime should be correct")
	} else {
		assert.WithinRange(t, info.ServerTime.Time(), time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), "ServerTime should be within a day of now")
	}
}

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	results, err := b.FetchTradablePairs(context.Background(), asset.Spot)
	require.NoError(t, err)
	require.NotNil(t, results)
	results, err = b.FetchTradablePairs(context.Background(), asset.CoinMarginedFutures)
	require.NoError(t, err)
	require.NotNil(t, results)
	results, err = b.FetchTradablePairs(context.Background(), asset.USDTMarginedFutures)
	require.NoError(t, err)
	require.NotNil(t, results)
	results, err = b.FetchTradablePairs(context.Background(), asset.Options)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestGetOrderBook(t *testing.T) {
	t.Parallel()
	_, err := b.GetOrderBook(context.Background(),
		OrderBookDataRequestParams{
			Symbol: currency.NewPair(currency.BTC, currency.USDT),
			Limit:  1000,
		})

	assert.NoError(t, err)
}

func TestGetMostRecentTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetMostRecentTrades(context.Background(),
		RecentTradeRequestParams{
			Symbol: currency.NewPair(currency.BTC, currency.USDT),
			Limit:  15,
		})

	assert.NoError(t, err)
}

func TestGetHistoricalTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetHistoricalTrades(context.Background(), "BTCUSDT", 5, -1)
	require.False(t, !mockTests && err == nil, "expected error")
	assert.False(t, mockTests && err != nil, err)
}

func TestGetAggregatedTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetAggregatedTrades(context.Background(),
		&AggregatedTradeRequestParams{
			Symbol: "BTCUSDT",
			Limit:  5,
		})
	assert.NoError(t, err)
}

func TestGetSpotKline(t *testing.T) {
	t.Parallel()
	start, end := getTime()
	_, err := b.GetSpotKline(context.Background(),
		&KlinesRequestParams{
			Symbol:    currency.NewPair(currency.BTC, currency.USDT),
			Interval:  kline.FiveMin.Short(),
			Limit:     24,
			StartTime: start,
			EndTime:   end,
		})
	assert.NoError(t, err)
}

func TestGetAveragePrice(t *testing.T) {
	t.Parallel()

	_, err := b.GetAveragePrice(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestGetPriceChangeStats(t *testing.T) {
	t.Parallel()

	_, err := b.GetPriceChangeStats(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestGetTickers(t *testing.T) {
	t.Parallel()

	_, err := b.GetTickers(context.Background())
	assert.NoError(t, err)
}

func TestGetLatestSpotPrice(t *testing.T) {
	t.Parallel()

	_, err := b.GetLatestSpotPrice(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestGetBestPrice(t *testing.T) {
	t.Parallel()

	_, err := b.GetBestPrice(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestQueryOrder(t *testing.T) {
	t.Parallel()

	_, err := b.QueryOrder(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", 1337)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

func TestOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.OpenOrders(context.Background(), currency.EMPTYPAIR)
	assert.NoError(t, err)

	p := currency.NewPair(currency.BTC, currency.USDT)
	_, err = b.OpenOrders(context.Background(), p)
	assert.NoError(t, err)
}

func TestAllOrders(t *testing.T) {
	t.Parallel()
	_, err := b.AllOrders(context.Background(), currency.NewPair(currency.BTC, currency.USDT), "", "")
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	t.Parallel()

	var feeBuilder = setFeeBuilder()
	_, err := b.GetFeeByType(context.Background(), feeBuilder)
	require.NoError(t, err)
	if !sharedtestvalues.AreAPICredentialsSet(b) || mockTests {
		require.Equal(t, exchange.OfflineTradeFee, feeBuilder.FeeType)
	} else {
		require.Equal(t, exchange.CryptocurrencyTradeFee, feeBuilder.FeeType)
	}
}

func TestGetFee(t *testing.T) {
	t.Parallel()

	var feeBuilder = setFeeBuilder()

	if sharedtestvalues.AreAPICredentialsSet(b) && mockTests {
		// CryptocurrencyTradeFee Basic
		_, err := b.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err)

		// CryptocurrencyTradeFee High quantity
		feeBuilder = setFeeBuilder()
		feeBuilder.Amount = 1000
		feeBuilder.PurchasePrice = 1000
		_, err = b.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err)

		// CryptocurrencyTradeFee IsMaker
		feeBuilder = setFeeBuilder()
		feeBuilder.IsMaker = true
		_, err = b.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err)

		// CryptocurrencyTradeFee Negative purchase price
		feeBuilder = setFeeBuilder()
		feeBuilder.PurchasePrice = -1000
		_, err = b.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err)
	}

	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	_, err := b.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err)

	// CryptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyDepositFee
	_, err = b.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err)

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.FiatCurrency = currency.HKD
	_, err = b.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err)

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.HKD
	_, err = b.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err)
}

func TestFormatWithdrawPermissions(t *testing.T) {
	t.Parallel()

	expectedResult := exchange.AutoWithdrawCryptoText + " & " + exchange.NoFiatWithdrawalsText
	withdrawPermissions := b.FormatWithdrawPermissions()
	require.Equal(t, expectedResult, withdrawPermissions)
}

func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	pair, err := currency.NewPairFromString("BTC_USDT")
	assert.NoError(t, err)
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		Pairs:     currency.Pairs{pair},
		AssetType: asset.Spot,
		Side:      order.AnySide,
	}

	_, err = b.GetActiveOrders(context.Background(), &getOrdersRequest)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

func TestGetOrderHistory(t *testing.T) {
	t.Parallel()

	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.AnySide,
	}

	_, err := b.GetOrderHistory(context.Background(), &getOrdersRequest)
	assert.NotEmpty(t, err, "Expected: 'At least one currency is required to fetch order history'. received nil")

	getOrdersRequest.Pairs = []currency.Pair{
		currency.NewPair(currency.LTC,
			currency.BTC)}

	_, err = b.GetOrderHistory(context.Background(), &getOrdersRequest)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

func TestNewOrderTest(t *testing.T) {
	t.Parallel()

	req := &NewOrderRequest{
		Symbol:      currency.NewPair(currency.LTC, currency.BTC),
		Side:        order.Buy.String(),
		TradeType:   BinanceRequestParamsOrderLimit,
		Price:       0.0025,
		Quantity:    100000,
		TimeInForce: BinanceRequestParamsTimeGTC,
	}

	err := b.NewOrderTest(context.Background(), req)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)

	req = &NewOrderRequest{
		Symbol:        currency.NewPair(currency.LTC, currency.BTC),
		Side:          order.Sell.String(),
		TradeType:     BinanceRequestParamsOrderMarket,
		Price:         0.0045,
		QuoteOrderQty: 10,
	}

	err = b.NewOrderTest(context.Background(), req)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	p := currency.NewPair(currency.BTC, currency.USDT)
	start := time.Unix(1577977445, 0)  // 2020-01-02 15:04:05
	end := start.Add(15 * time.Minute) // 2020-01-02 15:19:05
	if b.IsAPIStreamConnected() {
		start = time.Now().Add(-time.Hour * 10)
		end = time.Now().Add(-time.Hour)
	}
	result, err := b.GetHistoricTrades(context.Background(), p, asset.Spot, start, end)
	assert.NoError(t, err)
	expected := 2134
	if b.IsAPIStreamConnected() {
		expected = len(result)
	} else if mockTests {
		expected = 1002
	}
	assert.Equal(t, expected, len(result), "GetHistoricTrades should return correct number of entries") //nolint:testifylint // assert.Len doesn't produce clear messages on result
	for _, r := range result {
		if !assert.WithinRange(t, r.Timestamp, start, end, "All trades should be within time range") {
			break
		}
	}
}

func TestGetAggregatedTradesBatched(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("BTCUSDT")
	require.NoError(t, err)
	start, err := time.Parse(time.RFC3339, "2020-01-02T15:04:05Z")
	require.NoError(t, err)
	expectTime, err := time.Parse(time.RFC3339Nano, "2020-01-02T16:19:04.831Z")
	require.NoError(t, err)
	tests := []struct {
		name string
		// mock test or live test
		mock         bool
		args         *AggregatedTradeRequestParams
		numExpected  int
		lastExpected time.Time
	}{
		{
			name: "mock batch with timerange",
			mock: true,
			args: &AggregatedTradeRequestParams{
				Symbol:    currencyPair.String(),
				StartTime: start,
				EndTime:   start.Add(75 * time.Minute),
			},
			numExpected:  1012,
			lastExpected: time.Date(2020, 1, 2, 16, 18, 31, int(919*time.Millisecond), time.UTC),
		},
		{
			name: "batch with timerange",
			args: &AggregatedTradeRequestParams{
				Symbol:    currencyPair.String(),
				StartTime: start,
				EndTime:   start.Add(75 * time.Minute),
			},
			numExpected:  12130,
			lastExpected: expectTime,
		},
		{
			name: "mock custom limit with start time set, no end time",
			mock: true,
			args: &AggregatedTradeRequestParams{
				Symbol:    "BTCUSDT",
				StartTime: start,
				Limit:     1001,
			},
			numExpected:  1001,
			lastExpected: time.Date(2020, 1, 2, 15, 18, 39, int(226*time.Millisecond), time.UTC),
		},
		{
			name: "custom limit with start time set, no end time",
			args: &AggregatedTradeRequestParams{
				Symbol:    "BTCUSDT",
				StartTime: time.Date(2020, 11, 18, 23, 0, 28, 921, time.UTC),
				Limit:     1001,
			},
			numExpected:  1001,
			lastExpected: time.Date(2020, 11, 18, 23, 1, 33, int(62*time.Millisecond*10), time.UTC),
		},
		{
			name: "mock recent trades",
			mock: true,
			args: &AggregatedTradeRequestParams{
				Symbol: "BTCUSDT",
				Limit:  3,
			},
			numExpected:  3,
			lastExpected: time.Date(2020, 1, 2, 16, 19, 5, int(200*time.Millisecond), time.UTC),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.mock != mockTests {
				t.Skip("mock mismatch, skipping")
			}
			result, err := b.GetAggregatedTrades(context.Background(), tt.args)
			assert.NoError(t, err)
			assert.Lenf(t, result, tt.numExpected, "expected %v entries, got %v", tt.numExpected, len(result))
			lastTradeTime := result[len(result)-1].TimeStamp
			require.True(t, lastTradeTime.Time().Equal(tt.lastExpected), "last trade expected %v, got %v", tt.lastExpected.UTC(), lastTradeTime.Time().UTC())
		})
	}
}

func TestGetAggregatedTradesErrors(t *testing.T) {
	t.Parallel()
	start, err := time.Parse(time.RFC3339, "2020-01-02T15:04:05Z")
	require.NoError(t, err)
	tests := []struct {
		name string
		args *AggregatedTradeRequestParams
	}{
		{
			name: "get recent trades does not support custom limit",
			args: &AggregatedTradeRequestParams{
				Symbol: "BTCUSDT",
				Limit:  1001,
			},
		},
		{
			name: "start time and fromId cannot be both set",
			args: &AggregatedTradeRequestParams{
				Symbol:    "BTCUSDT",
				StartTime: start,
				EndTime:   start.Add(75 * time.Minute),
				FromID:    2,
			},
		},
		{
			name: "can't get most recent 5000 (more than 1000 not allowed)",
			args: &AggregatedTradeRequestParams{
				Symbol: "BTCUSDT",
				Limit:  5000,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := b.GetAggregatedTrades(context.Background(), tt.args)
			require.Error(t, err)
		})
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// -----------------------------------------------------------------------------------------------------------------------------

func TestSubmitOrder(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}

	var orderSubmission = &order.Submit{
		Exchange: b.Name,
		Pair: currency.Pair{
			Delimiter: "_",
			Base:      currency.LTC,
			Quote:     currency.BTC,
		},
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1000000000,
		ClientID:  "meowOrder",
		AssetType: asset.Spot,
	}

	_, err := b.SubmitOrder(context.Background(), orderSubmission)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	assert.False(t, mockTests && err != nil, err)
}

func TestCancelExchangeOrder(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}
	var orderCancellation = &order.Cancel{
		OrderID:       "1",
		WalletAddress: core.BitcoinDonationAddress,
		AccountID:     "1",
		Pair:          currency.NewPair(currency.LTC, currency.BTC),
		AssetType:     asset.Spot,
	}

	err := b.CancelOrder(context.Background(), orderCancellation)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	require.False(t, mockTests && err != nil, err)
}

func TestCancelAllExchangeOrders(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}
	var orderCancellation = &order.Cancel{
		OrderID:       "1",
		WalletAddress: core.BitcoinDonationAddress,
		AccountID:     "1",
		Pair:          currency.NewPair(currency.LTC, currency.BTC),
		AssetType:     asset.Spot,
	}

	_, err := b.CancelAllOrders(context.Background(), orderCancellation)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
	assert.False(t, mockTests && err != nil, err)
}

func TestGetAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	items := asset.Items{
		asset.CoinMarginedFutures,
		asset.USDTMarginedFutures,
		asset.Spot,
		asset.Margin,
	}
	for i := range items {
		assetType := items[i]
		t.Run(fmt.Sprintf("Update info of account [%s]", assetType.String()), func(t *testing.T) {
			t.Parallel()
			_, err := b.UpdateAccountInfo(context.Background(), assetType)
			assert.NoError(t, err)
		})
	}
}

func TestWrapperGetActiveOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	p, err := currency.NewPairFromString("EOS-USDT")
	assert.NoError(t, err)
	_, err = b.GetActiveOrders(context.Background(), &order.MultiOrderRequest{
		Type:      order.AnyType,
		Side:      order.AnySide,
		Pairs:     currency.Pairs{p},
		AssetType: asset.CoinMarginedFutures,
	})
	assert.NoError(t, err)

	p2, err := currency.NewPairFromString("BTCUSDT")
	assert.NoError(t, err)
	_, err = b.GetActiveOrders(context.Background(), &order.MultiOrderRequest{
		Type:      order.AnyType,
		Side:      order.AnySide,
		Pairs:     currency.Pairs{p2},
		AssetType: asset.USDTMarginedFutures,
	})
	assert.NoError(t, err)
}

func TestWrapperGetOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	p, err := currency.NewPairFromString("EOSUSD_PERP")
	assert.NoError(t, err)
	_, err = b.GetOrderHistory(context.Background(), &order.MultiOrderRequest{
		Type:        order.AnyType,
		Side:        order.AnySide,
		FromOrderID: "123",
		Pairs:       currency.Pairs{p},
		AssetType:   asset.CoinMarginedFutures,
	})
	assert.NoError(t, err)

	p2, err := currency.NewPairFromString("BTCUSDT")
	assert.NoError(t, err)
	_, err = b.GetOrderHistory(context.Background(), &order.MultiOrderRequest{
		Type:        order.AnyType,
		Side:        order.AnySide,
		FromOrderID: "123",
		Pairs:       currency.Pairs{p2},
		AssetType:   asset.USDTMarginedFutures,
	})
	assert.NoError(t, err)

	_, err = b.GetOrderHistory(context.Background(), &order.MultiOrderRequest{
		AssetType: asset.USDTMarginedFutures,
	})
	assert.Error(t, err, "expecting an error since invalid param combination is given. Got err: %v", err)
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	p, err := currency.NewPairFromString("EOS-USDT")
	assert.NoError(t, err)
	fPair, err := b.FormatExchangeCurrency(p, asset.CoinMarginedFutures)
	assert.NoError(t, err)
	err = b.CancelOrder(context.Background(), &order.Cancel{
		AssetType: asset.CoinMarginedFutures,
		Pair:      fPair,
		OrderID:   "1234",
	})
	assert.NoError(t, err)

	p2, err := currency.NewPairFromString("BTC-USDT")
	assert.NoError(t, err)
	fpair2, err := b.FormatExchangeCurrency(p2, asset.USDTMarginedFutures)
	assert.NoError(t, err)
	err = b.CancelOrder(context.Background(), &order.Cancel{
		AssetType: asset.USDTMarginedFutures,
		Pair:      fpair2,
		OrderID:   "1234",
	})
	assert.NoError(t, err)
}

func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	tradablePairs, err := b.FetchTradablePairs(context.Background(),
		asset.CoinMarginedFutures)
	assert.NoError(t, err)
	require.NotEmpty(t, tradablePairs, "no tradable pairs")
	_, err = b.GetOrderInfo(context.Background(),
		"123", tradablePairs[0], asset.CoinMarginedFutures)
	assert.NoError(t, err)
}

func TestGetAllCoinsInfo(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetAllCoinsInfo(context.Background())
	assert.NoError(t, err)
}

func TestWithdraw(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}

	withdrawCryptoRequest := withdraw.Request{
		Exchange:    b.Name,
		Amount:      -1,
		Currency:    currency.BTC,
		Description: "WITHDRAW IT ALL",
		Crypto: withdraw.CryptoRequest{
			Address: core.BitcoinDonationAddress,
		},
	}

	_, err := b.WithdrawCryptocurrencyFunds(context.Background(),
		&withdrawCryptoRequest)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	assert.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
}

func TestDepositHistory(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}
	_, err := b.DepositHistory(context.Background(), currency.ETH, "", time.Time{}, time.Time{}, 0, 10000)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	assert.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
}

func TestWithdrawHistory(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)
	}
	_, err := b.GetWithdrawalsHistory(context.Background(), currency.ETH, asset.Spot)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	assert.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "expecting an error when no keys are set")
}

func TestWithdrawFiat(t *testing.T) {
	t.Parallel()
	_, err := b.WithdrawFiatFunds(context.Background(),
		&withdraw.Request{})
	assert.Equal(t, err, common.ErrFunctionNotSupported)
}

func TestWithdrawInternationalBank(t *testing.T) {
	t.Parallel()
	_, err := b.WithdrawFiatFundsToInternationalBank(context.Background(),
		&withdraw.Request{})
	require.Equal(t, err, common.ErrFunctionNotSupported)
}

func TestGetDepositAddress(t *testing.T) {
	t.Parallel()
	_, err := b.GetDepositAddress(context.Background(), currency.USDT, "", currency.BNB.String())
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "error cannot be nil")
	require.False(t, mockTests && err != nil, err)
}

func BenchmarkWsHandleData(bb *testing.B) {
	bb.ReportAllocs()
	ap, err := b.CurrencyPairs.GetPairs(asset.Spot, false)
	require.NoError(bb, err)
	err = b.CurrencyPairs.StorePairs(asset.Spot, ap, true)
	require.NoError(bb, err)

	data, err := os.ReadFile("testdata/wsHandleData.json")
	require.NoError(bb, err)
	lines := bytes.Split(data, []byte("\n"))
	require.Len(bb, lines, 8)
	go func() {
		for {
			<-b.Websocket.DataHandler
		}
	}()
	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		for x := range lines {
			require.NoError(bb, b.wsHandleData(lines[x]))
		}
	}
}

func TestSubscribe(t *testing.T) {
	t.Parallel()
	b := b
	channels := []subscription.Subscription{
		{Channel: "btcusdt@ticker"},
		{Channel: "btcusdt@trade"},
	}
	if mockTests {
		mock := func(msg []byte, w *websocket.Conn) error {
			var req WsPayload
			err := json.Unmarshal(msg, &req)
			require.NoError(t, err, "unmarshal should not error")
			require.Len(t, req.Params, len(channels), "Params should only have 2 channel") // Failure might mean mockWSInstance default Subs is not empty
			assert.Equal(t, req.Params[0], channels[0].Channel, "Channel name should be correct")
			assert.Equal(t, req.Params[1], channels[1].Channel, "Channel name should be correct")
			return w.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"result":null,"id":%d}`, req.ID)))
		}
		b = testexch.MockWsInstance[Binance](t, testexch.CurryWsMockUpgrader(t, mock))
	} else {
		testexch.SetupWs(t, b)
	}
	err := b.Subscribe(channels)
	require.NoError(t, err, "Subscribe should not error")
	err = b.Unsubscribe(channels)
	require.NoError(t, err, "Unsubscribe should not error")
}

func TestSubscribeBadResp(t *testing.T) {
	t.Parallel()
	channels := []subscription.Subscription{
		{Channel: "moons@ticker"},
	}
	mock := func(msg []byte, w *websocket.Conn) error {
		var req WsPayload
		err := json.Unmarshal(msg, &req)
		require.NoError(t, err, "Unmarshal should not error")
		return w.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"result":{"error":"carrots"},"id":%d}`, req.ID)))
	}
	b := testexch.MockWsInstance[Binance](t, testexch.CurryWsMockUpgrader(t, mock)) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	err := b.Subscribe(channels)
	assert.ErrorIs(t, err, stream.ErrSubscriptionFailure, "Subscribe should error ErrSubscriptionFailure")
	assert.ErrorIs(t, err, errUnknownError, "Subscribe should error errUnknownError")
	assert.ErrorContains(t, err, "carrots", "Subscribe should error containing the carrots")
}

func TestWsTickerUpdate(t *testing.T) {
	t.Parallel()
	pressXToJSON := []byte(`{"stream":"btcusdt@ticker","data":{"e":"24hrTicker","E":1580254809477,"s":"BTCUSDT","p":"420.97000000","P":"4.720","w":"9058.27981278","x":"8917.98000000","c":"9338.96000000","Q":"0.17246300","b":"9338.03000000","B":"0.18234600","a":"9339.70000000","A":"0.14097600","o":"8917.99000000","h":"9373.19000000","l":"8862.40000000","v":"72229.53692000","q":"654275356.16896672","O":1580168409456,"C":1580254809456,"F":235294268,"L":235894703,"n":600436}}`)
	err := b.wsHandleData(pressXToJSON)
	assert.NoError(t, err)
}

func TestWsKlineUpdate(t *testing.T) {
	t.Parallel()
	pressXToJSON := []byte(`{"stream":"btcusdt@kline_1m","data":{
	  "e": "kline",
	  "E": 123456789,   
	  "s": "BTCUSDT",    
	  "k": {
		"t": 123400000, 
		"T": 123460000, 
		"s": "BTCUSDT",  
		"i": "1m",      
		"f": 100,       
		"L": 200,       
		"o": "0.0010",  
		"c": "0.0020",  
		"h": "0.0025",  
		"l": "0.0015",  
		"v": "1000",    
		"n": 100,       
		"x": false,     
		"q": "1.0000",  
		"V": "500",     
		"Q": "0.500",   
		"B": "123456"   
	  }
	}}`)
	err := b.wsHandleData(pressXToJSON)
	assert.NoError(t, err)
}

func TestWsTradeUpdate(t *testing.T) {
	t.Parallel()
	b.SetSaveTradeDataStatus(true)
	pressXToJSON := []byte(`{"stream":"btcusdt@trade","data":{
	  "e": "trade",     
	  "E": 123456789,   
	  "s": "BTCUSDT",    
	  "t": 12345,       
	  "p": "0.001",     
	  "q": "100",       
	  "b": 88,          
	  "a": 50,          
	  "T": 123456785,   
	  "m": true,        
	  "M": true         
	}}`)
	err := b.wsHandleData(pressXToJSON)
	assert.NoError(t, err)
}

func TestWsDepthUpdate(t *testing.T) {
	binanceOrderBookLock.Lock()
	defer binanceOrderBookLock.Unlock()
	b.setupOrderbookManager()
	seedLastUpdateID := int64(161)
	book := OrderBook{
		Asks: []OrderbookItem{
			{Price: 6621.80000000, Quantity: 0.00198100},
			{Price: 6622.14000000, Quantity: 4.00000000},
			{Price: 6622.46000000, Quantity: 2.30000000},
			{Price: 6622.47000000, Quantity: 1.18633300},
			{Price: 6622.64000000, Quantity: 4.00000000},
			{Price: 6622.73000000, Quantity: 0.02900000},
			{Price: 6622.76000000, Quantity: 0.12557700},
			{Price: 6622.81000000, Quantity: 2.08994200},
			{Price: 6622.82000000, Quantity: 0.01500000},
			{Price: 6623.17000000, Quantity: 0.16831300},
		},
		Bids: []OrderbookItem{
			{Price: 6621.55000000, Quantity: 0.16356700},
			{Price: 6621.45000000, Quantity: 0.16352600},
			{Price: 6621.41000000, Quantity: 0.86091200},
			{Price: 6621.25000000, Quantity: 0.16914100},
			{Price: 6621.23000000, Quantity: 0.09193600},
			{Price: 6621.22000000, Quantity: 0.00755100},
			{Price: 6621.13000000, Quantity: 0.08432000},
			{Price: 6621.03000000, Quantity: 0.00172000},
			{Price: 6620.94000000, Quantity: 0.30506700},
			{Price: 6620.93000000, Quantity: 0.00200000},
		},
		LastUpdateID: seedLastUpdateID,
	}

	update1 := []byte(`{"stream":"btcusdt@depth","data":{
	  "e": "depthUpdate", 
	  "E": 123456788,     
	  "s": "BTCUSDT",      
	  "U": 157,           
	  "u": 160,           
	  "b": [              
		["6621.45", "0.3"]
	  ],
	  "a": [              
		["6622.46", "1.5"]
	  ]
	}}`)

	p := currency.NewPairWithDelimiter("BTC", "USDT", "-")
	err := b.SeedLocalCacheWithBook(p, &book)
	require.NoError(t, err)

	err = b.wsHandleData(update1)
	assert.NoError(t, err)

	b.obm.state[currency.BTC][currency.USDT][asset.Spot].fetchingBook = false

	ob, err := b.Websocket.Orderbook.GetOrderbook(p, asset.Spot)
	require.NoError(t, err)

	exp, got := seedLastUpdateID, ob.LastUpdateID
	require.Equalf(t, exp, got, "Last update id of orderbook for old update. Exp: %d, got: %d", exp, got)
	expAmnt, gotAmnt := 2.3, ob.Asks[2].Amount
	require.Equalf(t, expAmnt, gotAmnt, "Ask altered by outdated update. Exp: %f, got %f", expAmnt, gotAmnt)
	expAmnt, gotAmnt = 0.163526, ob.Bids[1].Amount
	assert.Equalf(t, expAmnt, gotAmnt, "Bid altered by outdated update. Exp: %f, got %f", expAmnt, gotAmnt)

	update2 := []byte(`{"stream":"btcusdt@depth","data":{
	  "e": "depthUpdate", 
	  "E": 123456789,     
	  "s": "BTCUSDT",      
	  "U": 161,           
	  "u": 165,           
	  "b": [           
		["6621.45", "0.163526"]
	  ],
	  "a": [             
		["6622.46", "2.3"], 
		["6622.47", "1.9"]
	  ]
	}}`)

	err = b.wsHandleData(update2)
	assert.NoError(t, err)

	ob, err = b.Websocket.Orderbook.GetOrderbook(p, asset.Spot)
	require.NoError(t, err)
	exp, got = int64(165), ob.LastUpdateID
	require.Equalf(t, exp, got, "Unexpected Last update id of orderbook for new update. Exp: %d, got: %d", exp, got)
	expAmnt, gotAmnt = 2.3, ob.Asks[2].Amount
	require.Equalf(t, expAmnt, gotAmnt, "Unexpected Ask amount. Exp: %f, got %f", expAmnt, gotAmnt)
	expAmnt, gotAmnt = 1.9, ob.Asks[3].Amount
	require.Equal(t, expAmnt, gotAmnt, "Unexpected Ask amount. Exp: %f, got %f", exp, got)
	expAmnt, gotAmnt = 0.163526, ob.Bids[1].Amount
	require.Equal(t, expAmnt, gotAmnt, "Unexpected Bid amount. Exp: %f, got %f", exp, got)

	// reset order book sync status
	b.obm.state[currency.BTC][currency.USDT][asset.Spot].lastUpdateID = 0
}

func TestWsBalanceUpdate(t *testing.T) {
	t.Parallel()
	pressXToJSON := []byte(`{"stream":"jTfvpakT2yT0hVIo5gYWVihZhdM2PrBgJUZ5PyfZ4EVpCkx4Uoxk5timcrQc","data":{
	"e": "balanceUpdate",         
	"E": 1573200697110,           
	"a": "BTC",                   
	"d": "100.00000000",          
	"T": 1573200697068            
	}}`)
	err := b.wsHandleData(pressXToJSON)
	assert.NoError(t, err)
}

func TestWsOCO(t *testing.T) {
	t.Parallel()
	pressXToJSON := []byte(`{"stream":"jTfvpakT2yT0hVIo5gYWVihZhdM2PrBgJUZ5PyfZ4EVpCkx4Uoxk5timcrQc","data":{
	"e": "listStatus",                
	"E": 1564035303637,               
	"s": "ETHBTC",                    
	"g": 2,                           
	"c": "OCO",                       
	"l": "EXEC_STARTED",              
	"L": "EXECUTING",                 
	"r": "NONE",                      
	"C": "F4QN4G8DlFATFlIUQ0cjdD",    
	"T": 1564035303625,               
	"O": [                            
		{
		"s": "ETHBTC",                
		"i": 17,                      
		"c": "AJYsMjErWJesZvqlJCTUgL" 
		},
		{
		"s": "ETHBTC",
		"i": 18,
		"c": "bfYPSQdLoqAJeNrOr9adzq"
		}
	]
	}}`)
	err := b.wsHandleData(pressXToJSON)
	assert.NoError(t, err)
}

func TestGetWsAuthStreamKey(t *testing.T) {
	key, err := b.GetWsAuthStreamKey(context.Background())
	require.False(t, mockTests && err != nil,
		!mockTests && sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !mockTests && !sharedtestvalues.AreAPICredentialsSet(b) && err == nil, "expected error")
	assert.False(t, key == "" && (sharedtestvalues.AreAPICredentialsSet(b) || mockTests), "expected key")
}

func TestMaintainWsAuthStreamKey(t *testing.T) {
	err := b.MaintainWsAuthStreamKey(context.Background())
	require.False(t, mockTests && err != nil,
		!mockTests && sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	assert.False(t, !mockTests && !sharedtestvalues.AreAPICredentialsSet(b) && err == nil, "expected error")
}

func TestExecutionTypeToOrderStatus(t *testing.T) {
	type TestCases struct {
		Case   string
		Result order.Status
	}
	testCases := []TestCases{
		{Case: "NEW", Result: order.New},
		{Case: "PARTIALLY_FILLED", Result: order.PartiallyFilled},
		{Case: "FILLED", Result: order.Filled},
		{Case: "CANCELED", Result: order.Cancelled},
		{Case: "PENDING_CANCEL", Result: order.PendingCancel},
		{Case: "REJECTED", Result: order.Rejected},
		{Case: "EXPIRED", Result: order.Expired},
		{Case: "LOL", Result: order.UnknownStatus},
	}
	for i := range testCases {
		result, _ := stringToOrderStatus(testCases[i].Case)
		require.Equal(t, result, testCases[i].Result, "Expected: %v, received: %v", testCases[i].Result, result)
	}
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := startTime.Add(time.Hour * 24 * 7)
	bAssets := b.GetAssetTypes(false)
	for i := range bAssets {
		cps, err := b.GetAvailablePairs(bAssets[i])
		assert.NoError(t, err)
		err = b.CurrencyPairs.EnablePair(bAssets[i], cps[0])
		require.False(t, err != nil && !errors.Is(err, currency.ErrPairAlreadyEnabled), err)
		_, err = b.GetHistoricCandles(context.Background(), cps[0], bAssets[i], kline.OneDay, startTime, end)
		assert.NoError(t, err)
	}

	pair, err := currency.NewPairFromString("BTC-USDT")
	require.NoError(t, err)
	startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = b.GetHistoricCandles(context.Background(), pair, asset.Spot, kline.Interval(time.Hour*7), startTime, end)
	require.ErrorIs(t, err, kline.ErrRequestExceedsExchangeLimits)
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	t.Parallel()
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := startTime.Add(time.Hour * 24 * 7)
	bAssets := b.GetAssetTypes(false)
	for i := range bAssets {
		cps, err := b.GetAvailablePairs(bAssets[i])
		require.NoError(t, err)
		err = b.CurrencyPairs.EnablePair(bAssets[i], cps[0])
		require.True(t, err == nil || errors.Is(err, currency.ErrPairAlreadyEnabled), err)
		_, err = b.GetHistoricCandlesExtended(context.Background(), cps[0], bAssets[i], kline.OneDay, startTime, end)
		require.NoError(t, err)
	}
}

func TestBinance_FormatExchangeKlineInterval(t *testing.T) {
	testCases := []struct {
		name     string
		interval kline.Interval
		output   string
	}{
		{
			"OneMin",
			kline.OneMin,
			"1m",
		},
		{
			"OneDay",
			kline.OneDay,
			"1d",
		},
		{
			"OneWeek",
			kline.OneWeek,
			"1w",
		},
		{
			"OneMonth",
			kline.OneMonth,
			"1M",
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			ret := b.FormatExchangeKlineInterval(test.interval)
			require.Equal(t, ret, test.output, "unexpected result return expected: %v received: %v", test.output, ret)
		})
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	pair := currency.NewPair(currency.BTC, currency.USDT)
	_, err := b.GetRecentTrades(context.Background(),
		pair, asset.Spot)
	assert.NoError(t, err)
	_, err = b.GetRecentTrades(context.Background(),
		pair, asset.USDTMarginedFutures)
	assert.NoError(t, err)
	pair.Base = currency.NewCode("BTCUSD")
	pair.Quote = currency.PERP
	_, err = b.GetRecentTrades(context.Background(),
		pair, asset.CoinMarginedFutures)
	assert.NoError(t, err)
}

func TestGetAvailableTransferChains(t *testing.T) {
	t.Parallel()
	_, err := b.GetAvailableTransferChains(context.Background(), currency.BTC)
	require.False(t, sharedtestvalues.AreAPICredentialsSet(b) && err != nil, err)
	require.False(t, !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests, "error cannot be nil")
	require.False(t, mockTests && err != nil, err)
}

func TestSeedLocalCache(t *testing.T) {
	t.Parallel()
	err := b.SeedLocalCache(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	require.NoError(t, err)
}

func TestGenerateSubscriptions(t *testing.T) {
	t.Parallel()
	expected := []subscription.Subscription{}
	pairs, err := b.GetEnabledPairs(asset.Spot)
	assert.NoError(t, err)
	for _, p := range pairs {
		for _, c := range []string{"kline_1m", "depth@100ms", "ticker", "trade"} {
			expected = append(expected, subscription.Subscription{
				Channel: p.Format(currency.PairFormat{Delimiter: "", Uppercase: false}).String() + "@" + c,
				Pair:    p,
				Asset:   asset.Spot,
			})
		}
	}
	subs, err := b.GenerateSubscriptions()
	assert.NoError(t, err)
	if assert.Len(t, subs, len(expected), "Should have the correct number of subs") {
		assert.ElementsMatch(t, subs, expected, "Should get the correct subscriptions")
	}
}

func TestChannelName(t *testing.T) {
	_, err := channelName(&subscription.Subscription{Channel: "Wobbegongs"})
	assert.ErrorIs(t, err, stream.ErrSubscriptionNotSupported, "Invalid channel name should return ErrSubNotSupported")
	assert.ErrorContains(t, err, "Wobbegong", "Invalid channel name error should contain at least one shark")

	n, err := channelName(&subscription.Subscription{Channel: subscription.TickerChannel})
	assert.NoError(t, err)
	assert.Equal(t, "ticker", n, "Ticker channel name should be correct")

	n, err = channelName(&subscription.Subscription{Channel: subscription.AllTradesChannel})
	assert.NoError(t, err)
	assert.Equal(t, "trade", n, "Trades channel name should be correct")

	n, err = channelName(&subscription.Subscription{Channel: subscription.OrderbookChannel})
	assert.NoError(t, err)
	assert.Equal(t, "depth@0s", n, "Orderbook with no update rate should return 0s") // It's not channelName's job to supply defaults

	n, err = channelName(&subscription.Subscription{Channel: subscription.OrderbookChannel, Interval: kline.Interval(time.Second)})
	assert.NoError(t, err)
	assert.Equal(t, "depth@1000ms", n, "Orderbook with 1s update rate should 1000ms")

	n, err = channelName(&subscription.Subscription{Channel: subscription.OrderbookChannel, Interval: kline.HundredMilliseconds})
	assert.NoError(t, err)
	assert.Equal(t, "depth@100ms", n, "Orderbook with update rate should return it in the depth channel name")

	n, err = channelName(&subscription.Subscription{Channel: subscription.OrderbookChannel, Interval: kline.HundredMilliseconds, Levels: 5})
	assert.NoError(t, err)
	assert.Equal(t, "depth@5@100ms", n, "Orderbook with Level should return it in the depth channel name")

	n, err = channelName(&subscription.Subscription{Channel: subscription.CandlesChannel, Interval: kline.FifteenMin})
	assert.NoError(t, err)
	assert.Equal(t, "kline_15m", n, "Candles with interval should return it in the depth channel name")

	n, err = channelName(&subscription.Subscription{Channel: subscription.CandlesChannel})
	assert.NoError(t, err)
	assert.Equal(t, "kline_0s", n, "Candles with no interval should return 0s") // It's not channelName's job to supply defaults
}

var websocketDepthUpdate = []byte(`{"E":1608001030784,"U":7145637266,"a":[["19455.19000000","0.59490200"],["19455.37000000","0.00000000"],["19456.11000000","0.00000000"],["19456.16000000","0.00000000"],["19458.67000000","0.06400000"],["19460.73000000","0.05139800"],["19461.43000000","0.00000000"],["19464.59000000","0.00000000"],["19466.03000000","0.45000000"],["19466.36000000","0.00000000"],["19508.67000000","0.00000000"],["19572.96000000","0.00217200"],["24386.00000000","0.00256600"]],"b":[["19455.18000000","2.94649200"],["19453.15000000","0.01233600"],["19451.18000000","0.00000000"],["19446.85000000","0.11427900"],["19446.74000000","0.00000000"],["19446.73000000","0.00000000"],["19444.45000000","0.14937800"],["19426.75000000","0.00000000"],["19416.36000000","0.36052100"]],"e":"depthUpdate","s":"BTCUSDT","u":7145637297}`)

func TestProcessUpdate(t *testing.T) {
	t.Parallel()
	binanceOrderBookLock.Lock()
	defer binanceOrderBookLock.Unlock()
	p := currency.NewPair(currency.BTC, currency.USDT)
	var depth WebsocketDepthStream
	err := json.Unmarshal(websocketDepthUpdate, &depth)
	require.NoError(t, err)

	err = b.obm.stageWsUpdate(&depth, p, asset.Spot)
	require.NoError(t, err)

	err = b.obm.fetchBookViaREST(p)
	require.NoError(t, err)

	err = b.obm.cleanup(p)
	require.NoError(t, err)

	// reset order book sync status
	b.obm.state[currency.BTC][currency.USDT][asset.Spot].lastUpdateID = 0
}

func TestUFuturesHistoricalTrades(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	cp, err := currency.NewPairFromString("BTCUSDT")
	assert.NoError(t, err)
	_, err = b.UFuturesHistoricalTrades(context.Background(), cp, "", 5)
	assert.NoError(t, err)
	_, err = b.UFuturesHistoricalTrades(context.Background(), cp, "", 0)
	assert.NoError(t, err)
}

func TestSetExchangeOrderExecutionLimits(t *testing.T) {
	t.Parallel()
	err := b.UpdateOrderExecutionLimits(context.Background(), asset.Spot)
	require.NoError(t, err)
	err = b.UpdateOrderExecutionLimits(context.Background(), asset.CoinMarginedFutures)
	require.NoError(t, err)

	err = b.UpdateOrderExecutionLimits(context.Background(), asset.USDTMarginedFutures)
	require.NoError(t, err)

	err = b.UpdateOrderExecutionLimits(context.Background(), asset.Binary)
	require.NotEmpty(t, err, "expected unhandled case")

	cmfCP, err := currency.NewPairFromStrings("BTCUSD", "PERP")
	require.NoError(t, err)

	limit, err := b.GetOrderExecutionLimits(asset.CoinMarginedFutures, cmfCP)
	require.NoError(t, err)

	require.Empty(t, limit, "exchange limit should be loaded")

	err = limit.Conforms(0.000001, 0.1, order.Limit)
	require.ErrorIs(t, err, order.ErrAmountBelowMin)

	err = limit.Conforms(0.01, 1, order.Limit)
	assert.ErrorIs(t, err, order.ErrPriceBelowMin)
}

func TestWsOrderExecutionReport(t *testing.T) {
	// cannot run in parallel due to inspecting the DataHandler result
	payload := []byte(`{"stream":"jTfvpakT2yT0hVIo5gYWVihZhdM2PrBgJUZ5PyfZ4EVpCkx4Uoxk5timcrQc","data":{"e":"executionReport","E":1616627567900,"s":"BTCUSDT","c":"c4wyKsIhoAaittTYlIVLqk","S":"BUY","o":"LIMIT","f":"GTC","q":"0.00028400","p":"52789.10000000","P":"0.00000000","F":"0.00000000","g":-1,"C":"","x":"NEW","X":"NEW","r":"NONE","i":5340845958,"l":"0.00000000","z":"0.00000000","L":"0.00000000","n":"0","N":"BTC","T":1616627567900,"t":-1,"I":11388173160,"w":true,"m":false,"M":false,"O":1616627567900,"Z":"0.00000000","Y":"0.00000000","Q":"0.00000000","W":1616627567900}}`)
	// this is a buy BTC order, normally commission is charged in BTC, vice versa.
	expectedResult := order.Detail{
		Price:                52789.1,
		Amount:               0.00028400,
		AverageExecutedPrice: 0,
		QuoteAmount:          0,
		ExecutedAmount:       0,
		RemainingAmount:      0.00028400,
		Cost:                 0,
		CostAsset:            currency.USDT,
		Fee:                  0,
		FeeAsset:             currency.BTC,
		Exchange:             "Binance",
		OrderID:              "5340845958",
		ClientOrderID:        "c4wyKsIhoAaittTYlIVLqk",
		Type:                 order.Limit,
		Side:                 order.Buy,
		Status:               order.New,
		AssetType:            asset.Spot,
		Date:                 time.UnixMilli(1616627567900),
		LastUpdated:          time.UnixMilli(1616627567900),
		Pair:                 currency.NewPair(currency.BTC, currency.USDT),
	}
	// empty the channel. otherwise mock_test will fail
	for len(b.Websocket.DataHandler) > 0 {
		<-b.Websocket.DataHandler
	}

	err := b.wsHandleData(payload)
	require.NoError(t, err)
	res := <-b.Websocket.DataHandler
	switch r := res.(type) {
	case *order.Detail:
		assert.True(t, reflect.DeepEqual(expectedResult, *r), "results do not match:\nexpected: %v\nreceived: %v", expectedResult, *r)
	default:
		t.Fatalf("expected type order.Detail, found %T", res)
	}

	payload = []byte(`{"stream":"jTfvpakT2yT0hVIo5gYWVihZhdM2PrBgJUZ5PyfZ4EVpCkx4Uoxk5timcrQc","data":{"e":"executionReport","E":1616633041556,"s":"BTCUSDT","c":"YeULctvPAnHj5HXCQo9Mob","S":"BUY","o":"LIMIT","f":"GTC","q":"0.00028600","p":"52436.85000000","P":"0.00000000","F":"0.00000000","g":-1,"C":"","x":"TRADE","X":"FILLED","r":"NONE","i":5341783271,"l":"0.00028600","z":"0.00028600","L":"52436.85000000","n":"0.00000029","N":"BTC","T":1616633041555,"t":726946523,"I":11390206312,"w":false,"m":false,"M":true,"O":1616633041555,"Z":"14.99693910","Y":"14.99693910","Q":"0.00000000","W":1616633041555}}`)
	err = b.wsHandleData(payload)
	require.NoError(t, err)
}

func TestWsOutboundAccountPosition(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"stream":"jTfvpakT2yT0hVIo5gYWVihZhdM2PrBgJUZ5PyfZ4EVpCkx4Uoxk5timcrQc","data":{"e":"outboundAccountPosition","E":1616628815745,"u":1616628815745,"B":[{"a":"BTC","f":"0.00225109","l":"0.00123000"},{"a":"BNB","f":"0.00000000","l":"0.00000000"},{"a":"USDT","f":"54.43390661","l":"0.00000000"}]}}`)
	err := b.wsHandleData(payload)
	require.NoError(t, err)
}

func TestFormatExchangeCurrency(t *testing.T) {
	t.Parallel()
	type testos struct {
		name              string
		pair              currency.Pair
		asset             asset.Item
		expectedDelimiter string
	}
	testerinos := []testos{
		{
			name:              "spot-btcusdt",
			pair:              currency.NewPairWithDelimiter("BTC", "USDT", currency.UnderscoreDelimiter),
			asset:             asset.Spot,
			expectedDelimiter: "",
		},
		{
			name:              "coinmarginedfutures-btcusd_perp",
			pair:              currency.NewPairWithDelimiter("BTCUSD", "PERP", currency.DashDelimiter),
			asset:             asset.CoinMarginedFutures,
			expectedDelimiter: currency.UnderscoreDelimiter,
		},
		{
			name:              "coinmarginedfutures-btcusd_211231",
			pair:              currency.NewPairWithDelimiter("BTCUSD", "211231", currency.DashDelimiter),
			asset:             asset.CoinMarginedFutures,
			expectedDelimiter: currency.UnderscoreDelimiter,
		},
		{
			name:              "margin-ltousdt",
			pair:              currency.NewPairWithDelimiter("LTO", "USDT", currency.UnderscoreDelimiter),
			asset:             asset.Margin,
			expectedDelimiter: "",
		},
		{
			name:              "usdtmarginedfutures-btcusdt",
			pair:              currency.NewPairWithDelimiter("btc", "usdt", currency.DashDelimiter),
			asset:             asset.USDTMarginedFutures,
			expectedDelimiter: "",
		},
		{
			name:              "usdtmarginedfutures-btcusdt_211231",
			pair:              currency.NewPairWithDelimiter("btcusdt", "211231", currency.UnderscoreDelimiter),
			asset:             asset.USDTMarginedFutures,
			expectedDelimiter: currency.UnderscoreDelimiter,
		},
	}
	for i := range testerinos {
		tt := testerinos[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := b.FormatExchangeCurrency(tt.pair, tt.asset)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedDelimiter, result.Delimiter)
		})
	}
}

func TestFormatSymbol(t *testing.T) {
	t.Parallel()
	type testos struct {
		name           string
		pair           currency.Pair
		asset          asset.Item
		expectedString string
	}
	testerinos := []testos{
		{
			name:           "spot-BTCUSDT",
			pair:           currency.NewPairWithDelimiter("BTC", "USDT", currency.UnderscoreDelimiter),
			asset:          asset.Spot,
			expectedString: "BTCUSDT",
		},
		{
			name:           "coinmarginedfutures-btcusdperp",
			pair:           currency.NewPairWithDelimiter("BTCUSD", "PERP", currency.DashDelimiter),
			asset:          asset.CoinMarginedFutures,
			expectedString: "BTCUSD_PERP",
		},
		{
			name:           "coinmarginedfutures-BTCUSD_211231",
			pair:           currency.NewPairWithDelimiter("BTCUSD", "211231", currency.DashDelimiter),
			asset:          asset.CoinMarginedFutures,
			expectedString: "BTCUSD_211231",
		},
		{
			name:           "margin-LTOUSDT",
			pair:           currency.NewPairWithDelimiter("LTO", "USDT", currency.UnderscoreDelimiter),
			asset:          asset.Margin,
			expectedString: "LTOUSDT",
		},
		{
			name:           "usdtmarginedfutures-BTCUSDT",
			pair:           currency.NewPairWithDelimiter("btc", "usdt", currency.DashDelimiter),
			asset:          asset.USDTMarginedFutures,
			expectedString: "BTCUSDT",
		},
		{
			name:           "usdtmarginedfutures-BTCUSDT_211231",
			pair:           currency.NewPairWithDelimiter("btcusdt", "211231", currency.UnderscoreDelimiter),
			asset:          asset.USDTMarginedFutures,
			expectedString: "BTCUSDT_211231",
		},
	}
	for i := range testerinos {
		tt := testerinos[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := b.FormatSymbol(tt.pair, tt.asset)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestFormatUSDTMarginedFuturesPair(t *testing.T) {
	t.Parallel()
	pairFormat := currency.PairFormat{Uppercase: true}
	resp := b.formatUSDTMarginedFuturesPair(currency.NewPair(currency.DOGE, currency.USDT), pairFormat)
	assert.Equal(t, "DOGEUSDT", resp.String())

	resp = b.formatUSDTMarginedFuturesPair(currency.NewPair(currency.DOGE, currency.NewCode("1234567890")), pairFormat)
	assert.Equal(t, "DOGE_1234567890", resp.String())
}

func TestFetchExchangeLimits(t *testing.T) {
	t.Parallel()
	limits, err := b.FetchExchangeLimits(context.Background(), asset.Spot)
	require.NoError(t, err)
	require.NotEmpty(t, limits, "Should get some limits back")

	limits, err = b.FetchExchangeLimits(context.Background(), asset.Margin)
	require.NoError(t, err)
	require.NotEmpty(t, limits, "Should get some limits back")

	_, err = b.FetchExchangeLimits(context.Background(), asset.Futures)
	assert.ErrorIs(t, err, asset.ErrNotSupported, "FetchExchangeLimits should error on other asset types")
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()

	tests := map[asset.Item]currency.Pair{
		asset.Spot:   currency.NewPair(currency.BTC, currency.USDT),
		asset.Margin: currency.NewPair(currency.ETH, currency.BTC),
	}
	for _, a := range []asset.Item{asset.CoinMarginedFutures, asset.USDTMarginedFutures} {
		pairs, err := b.FetchTradablePairs(context.Background(), a)
		require.NoErrorf(t, err, "FetchTradablePairs should not error for %s", a)
		require.NotEmptyf(t, pairs, "Should get some pairs for %s", a)
		tests[a] = pairs[0]
	}

	for _, a := range b.GetAssetTypes(false) {
		err := b.UpdateOrderExecutionLimits(context.Background(), a)
		require.NoError(t, err, "UpdateOrderExecutionLimits should not error")

		p := tests[a]
		limits, err := b.GetOrderExecutionLimits(a, p)
		require.NoErrorf(t, err, "GetOrderExecutionLimits should not error for %s pair %s", a, p)
		require.Positivef(t, limits.MinPrice, "MinPrice must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.MaxPrice, "MaxPrice must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.PriceStepIncrementSize, "PriceStepIncrementSize must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.MinimumBaseAmount, "MinimumBaseAmount must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.MaximumBaseAmount, "MaximumBaseAmount must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.AmountStepIncrementSize, "AmountStepIncrementSize must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.MarketMaxQty, "MarketMaxQty must be positive for %s pair %s", a, p)
		require.Positivef(t, limits.MaxTotalOrders, "MaxTotalOrders must be positive for %s pair %s", a, p)
		switch a {
		case asset.Spot, asset.Margin:
			assert.Positivef(t, limits.MaxIcebergParts, "MaxIcebergParts must be positive for %s pair %s", a, p)
		case asset.USDTMarginedFutures:
			assert.Positivef(t, limits.MinNotional, "MinNotional must be positive for %s pair %s", a, p)
			fallthrough
		case asset.CoinMarginedFutures:
			assert.Positivef(t, limits.MultiplierUp, "MultiplierUp must be positive for %s pair %s", a, p)
			assert.Positivef(t, limits.MultiplierDown, "MultiplierDown must be positive for %s pair %s", a, p)
			assert.Positivef(t, limits.MarketMinQty, "MarketMinQty must be positive for %s pair %s", a, p)
			assert.Positivef(t, limits.MarketStepIncrementSize, "MarketStepIncrementSize must be positive for %s pair %s", a, p)
			assert.Positivef(t, limits.MaxAlgoOrders, "MaxAlgoOrders must be positive for %s pair %s", a, p)
		}
	}
}

func TestGetFundingRates(t *testing.T) {
	t.Parallel()
	s, e := getTime()
	_, err := b.GetHistoricalFundingRates(context.Background(), &fundingrate.HistoricalRatesRequest{
		Asset:                asset.USDTMarginedFutures,
		Pair:                 currency.NewPair(currency.BTC, currency.USDT),
		StartDate:            s,
		EndDate:              e,
		IncludePayments:      true,
		IncludePredictedRate: true,
	})
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported)

	_, err = b.GetHistoricalFundingRates(context.Background(), &fundingrate.HistoricalRatesRequest{
		Asset:           asset.USDTMarginedFutures,
		Pair:            currency.NewPair(currency.BTC, currency.USDT),
		StartDate:       s,
		EndDate:         e,
		PaymentCurrency: currency.DOGE,
	})
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported)

	r := &fundingrate.HistoricalRatesRequest{
		Asset:     asset.USDTMarginedFutures,
		Pair:      currency.NewPair(currency.BTC, currency.USDT),
		StartDate: s,
		EndDate:   e,
	}
	if sharedtestvalues.AreAPICredentialsSet(b) {
		r.IncludePayments = true
	}
	_, err = b.GetHistoricalFundingRates(context.Background(), r)
	assert.NoError(t, err)

	r.Asset = asset.CoinMarginedFutures
	r.Pair, err = currency.NewPairFromString("BTCUSD_PERP")
	require.NoError(t, err)
	_, err = b.GetHistoricalFundingRates(context.Background(), r)
	assert.NoError(t, err)
}

func TestGetLatestFundingRates(t *testing.T) {
	t.Parallel()
	cp := currency.NewPair(currency.BTC, currency.USDT)
	_, err := b.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset:                asset.USDTMarginedFutures,
		Pair:                 cp,
		IncludePredictedRate: true,
	})
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported)
	err = b.CurrencyPairs.EnablePair(asset.USDTMarginedFutures, cp)
	require.True(t, err == nil || errors.Is(err, currency.ErrPairAlreadyEnabled), err)
	_, err = b.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset: asset.USDTMarginedFutures,
		Pair:  cp,
	})
	assert.NoError(t, err)
	_, err = b.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset: asset.CoinMarginedFutures,
	})
	assert.NoError(t, err)
}

func TestIsPerpetualFutureCurrency(t *testing.T) {
	t.Parallel()
	is, err := b.IsPerpetualFutureCurrency(asset.Binary, currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	assert.False(t, is)

	is, err = b.IsPerpetualFutureCurrency(asset.CoinMarginedFutures, currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	assert.False(t, is)
	is, err = b.IsPerpetualFutureCurrency(asset.CoinMarginedFutures, currency.NewPair(currency.BTC, currency.PERP))
	assert.NoError(t, err)
	assert.True(t, is)

	is, err = b.IsPerpetualFutureCurrency(asset.USDTMarginedFutures, currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
	assert.True(t, is)
	is, err = b.IsPerpetualFutureCurrency(asset.USDTMarginedFutures, currency.NewPair(currency.BTC, currency.PERP))
	assert.NoError(t, err)
	assert.False(t, is)
}

func TestGetUserMarginInterestHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetUserMarginInterestHistory(context.Background(), currency.USDT, currency.NewPair(currency.BTC, currency.USDT), time.Now().Add(-time.Hour*24), time.Now(), 1, 10, false)
	assert.NoError(t, err)
}

func TestSetAssetsMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	is, err := b.GetAssetsMode(context.Background())
	assert.NoError(t, err)

	err = b.SetAssetsMode(context.Background(), !is)
	assert.NoError(t, err)

	err = b.SetAssetsMode(context.Background(), is)
	assert.NoError(t, err)
}

func TestGetAssetsMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAssetsMode(context.Background())
	assert.NoError(t, err)
}

func TestGetCollateralMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.GetCollateralMode(context.Background(), asset.Spot)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	_, err = b.GetCollateralMode(context.Background(), asset.CoinMarginedFutures)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	_, err = b.GetCollateralMode(context.Background(), asset.USDTMarginedFutures)
	assert.NoError(t, err)
}

func TestSetCollateralMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetCollateralMode(context.Background(), asset.Spot, collateral.SingleMode)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	err = b.SetCollateralMode(context.Background(), asset.CoinMarginedFutures, collateral.SingleMode)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	err = b.SetCollateralMode(context.Background(), asset.USDTMarginedFutures, collateral.MultiMode)
	assert.NoError(t, err)
	err = b.SetCollateralMode(context.Background(), asset.USDTMarginedFutures, collateral.PortfolioMode)
	assert.ErrorIs(t, err, order.ErrCollateralInvalid)
}

func TestChangePositionMargin(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ChangePositionMargin(context.Background(), &margin.PositionChangeRequest{
		Pair:                    currency.NewBTCUSDT(),
		Asset:                   asset.USDTMarginedFutures,
		MarginType:              margin.Isolated,
		OriginalAllocatedMargin: 1337,
		NewAllocatedMargin:      1333337,
	})
	assert.NoError(t, err)
}

func TestGetPositionSummary(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	bb := currency.NewBTCUSDT()
	_, err := b.GetFuturesPositionSummary(context.Background(), &futures.PositionSummaryRequest{
		Asset: asset.USDTMarginedFutures,
		Pair:  bb,
	})
	assert.NoError(t, err)

	bb.Quote = currency.BUSD
	_, err = b.GetFuturesPositionSummary(context.Background(), &futures.PositionSummaryRequest{
		Asset: asset.USDTMarginedFutures,
		Pair:  bb,
	})
	assert.NoError(t, err)

	p, err := currency.NewPairFromString("BTCUSD_PERP")
	require.NoError(t, err)
	bb.Quote = currency.USD
	_, err = b.GetFuturesPositionSummary(context.Background(), &futures.PositionSummaryRequest{
		Asset:          asset.CoinMarginedFutures,
		Pair:           p,
		UnderlyingPair: bb,
	})
	assert.NoError(t, err)

	_, err = b.GetFuturesPositionSummary(context.Background(), &futures.PositionSummaryRequest{
		Asset:          asset.Spot,
		Pair:           p,
		UnderlyingPair: bb,
	})
	assert.ErrorIs(t, err, asset.ErrNotSupported)
}

func TestGetFuturesPositionOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFuturesPositionOrders(context.Background(), &futures.PositionsRequest{
		Asset:                     asset.USDTMarginedFutures,
		Pairs:                     []currency.Pair{currency.NewBTCUSDT()},
		StartDate:                 time.Now().Add(-time.Hour * 24 * 70),
		RespectOrderHistoryLimits: true,
	})
	assert.NoError(t, err)

	p, err := currency.NewPairFromString("ADAUSD_PERP")
	require.NoError(t, err)
	_, err = b.GetFuturesPositionOrders(context.Background(), &futures.PositionsRequest{
		Asset:                     asset.CoinMarginedFutures,
		Pairs:                     []currency.Pair{p},
		StartDate:                 time.Now().Add(time.Hour * 24 * -70),
		RespectOrderHistoryLimits: true,
	})
	assert.NoError(t, err)
}

func TestSetMarginType(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	err := b.SetMarginType(context.Background(), asset.USDTMarginedFutures, currency.NewPair(currency.BTC, currency.USDT), margin.Isolated)
	assert.NoError(t, err)

	p, err := currency.NewPairFromString("BTCUSD_PERP")
	require.NoError(t, err)
	err = b.SetMarginType(context.Background(), asset.CoinMarginedFutures, p, margin.Isolated)
	assert.NoError(t, err)

	err = b.SetMarginType(context.Background(), asset.Spot, currency.NewPair(currency.BTC, currency.USDT), margin.Isolated)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
}

func TestGetLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLeverage(context.Background(), asset.USDTMarginedFutures, currency.NewBTCUSDT(), 0, order.UnknownSide)
	assert.NoError(t, err)

	p, err := currency.NewPairFromString("BTCUSD_PERP")
	require.NoError(t, err)
	_, err = b.GetLeverage(context.Background(), asset.CoinMarginedFutures, p, 0, order.UnknownSide)
	assert.NoError(t, err)
	_, err = b.GetLeverage(context.Background(), asset.Spot, currency.NewBTCUSDT(), 0, order.UnknownSide)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
}

func TestSetLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetLeverage(context.Background(), asset.USDTMarginedFutures, currency.NewBTCUSDT(), margin.Multi, 5, order.UnknownSide)
	assert.NoError(t, err)

	p, err := currency.NewPairFromString("BTCUSD_PERP")
	require.NoError(t, err)
	err = b.SetLeverage(context.Background(), asset.CoinMarginedFutures, p, margin.Multi, 5, order.UnknownSide)
	assert.NoError(t, err)
	err = b.SetLeverage(context.Background(), asset.Spot, p, margin.Multi, 5, order.UnknownSide)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
}

func TestGetCryptoLoansIncomeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanIncomeHistory(context.Background(), currency.USDT, "", time.Time{}, time.Time{}, 100)
	assert.NoError(t, err)
}

func TestCryptoLoanBorrow(t *testing.T) {
	t.Parallel()
	_, err := b.CryptoLoanBorrow(context.Background(), currency.EMPTYCODE, 1000, currency.BTC, 1, 7)
	require.ErrorIs(t, err, errLoanCoinMustBeSet)
	_, err = b.CryptoLoanBorrow(context.Background(), currency.USDT, 1000, currency.EMPTYCODE, 1, 7)
	assert.ErrorIs(t, err, errCollateralCoinMustBeSet)
	_, err = b.CryptoLoanBorrow(context.Background(), currency.USDT, 0, currency.BTC, 1, 0)
	require.ErrorIs(t, err, errLoanTermMustBeSet)
	_, err = b.CryptoLoanBorrow(context.Background(), currency.USDT, 0, currency.BTC, 0, 7)
	assert.ErrorIs(t, err, errEitherLoanOrCollateralAmountsMustBeSet)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CryptoLoanBorrow(context.Background(), currency.USDT, 1000, currency.BTC, 1, 7)
	assert.NoError(t, err)
}

func TestCryptoLoanBorrowHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanBorrowHistory(context.Background(), 0, currency.USDT, currency.BTC, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanOngoingOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanOngoingOrders(context.Background(), 0, currency.USDT, currency.BTC, 0, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanRepay(t *testing.T) {
	t.Parallel()
	_, err := b.CryptoLoanRepay(context.Background(), 0, 1000, 1, false)
	require.ErrorIs(t, err, errOrderIDMustBeSet)
	_, err = b.CryptoLoanRepay(context.Background(), 42069, 0, 1, false)
	assert.ErrorIs(t, err, errAmountMustBeSet)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CryptoLoanRepay(context.Background(), 42069, 1000, 1, false)
	assert.NoError(t, err)
}

func TestCryptoLoanRepaymentHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanRepaymentHistory(context.Background(), 0, currency.USDT, currency.BTC, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanAdjustLTV(t *testing.T) {
	t.Parallel()
	_, err := b.CryptoLoanAdjustLTV(context.Background(), 0, true, 1)
	require.ErrorIs(t, err, errOrderIDMustBeSet)
	_, err = b.CryptoLoanAdjustLTV(context.Background(), 42069, true, 0)
	require.ErrorIs(t, err, errAmountMustBeSet)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CryptoLoanAdjustLTV(context.Background(), 42069, true, 1)
	assert.NoError(t, err)
}

func TestCryptoLoanLTVAdjustmentHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanLTVAdjustmentHistory(context.Background(), 0, currency.USDT, currency.BTC, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanAssetsData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanAssetsData(context.Background(), currency.EMPTYCODE, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanCollateralAssetsData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.CryptoLoanCollateralAssetsData(context.Background(), currency.EMPTYCODE, 0)
	assert.NoError(t, err)
}

func TestCryptoLoanCheckCollateralRepayRate(t *testing.T) {
	t.Parallel()
	_, err := b.CryptoLoanCheckCollateralRepayRate(context.Background(), currency.EMPTYCODE, currency.BNB, 69)
	assert.ErrorIs(t, err, errLoanCoinMustBeSet)
	_, err = b.CryptoLoanCheckCollateralRepayRate(context.Background(), currency.BUSD, currency.EMPTYCODE, 69)
	assert.ErrorIs(t, err, errCollateralCoinMustBeSet)
	_, err = b.CryptoLoanCheckCollateralRepayRate(context.Background(), currency.BUSD, currency.BNB, 0)
	assert.ErrorIs(t, err, errAmountMustBeSet)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err = b.CryptoLoanCheckCollateralRepayRate(context.Background(), currency.BUSD, currency.BNB, 69)
	assert.NoError(t, err)
}

func TestCryptoLoanCustomiseMarginCall(t *testing.T) {
	t.Parallel()
	_, err := b.CryptoLoanCustomiseMarginCall(context.Background(), 0, currency.BTC, 0)
	assert.NotEmpty(t, err)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CryptoLoanCustomiseMarginCall(context.Background(), 1337, currency.BTC, .70)
	assert.NoError(t, err)
}

func TestFlexibleLoanBorrow(t *testing.T) {
	t.Parallel()
	_, err := b.FlexibleLoanBorrow(context.Background(), currency.EMPTYCODE, currency.USDC, 1, 0)
	require.ErrorIs(t, err, errLoanCoinMustBeSet)
	_, err = b.FlexibleLoanBorrow(context.Background(), currency.ATOM, currency.EMPTYCODE, 1, 0)
	require.ErrorIs(t, err, errCollateralCoinMustBeSet)
	_, err = b.FlexibleLoanBorrow(context.Background(), currency.ATOM, currency.USDC, 0, 0)
	require.ErrorIs(t, err, errEitherLoanOrCollateralAmountsMustBeSet)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.FlexibleLoanBorrow(context.Background(), currency.ATOM, currency.USDC, 1, 0)
	assert.NoError(t, err)
}

func TestFlexibleLoanOngoingOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleLoanOngoingOrders(context.Background(), currency.EMPTYCODE, currency.EMPTYCODE, 0, 0)
	assert.NoError(t, err)
}

func TestFlexibleLoanBorrowHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleLoanBorrowHistory(context.Background(), currency.EMPTYCODE, currency.EMPTYCODE, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestFlexibleLoanRepay(t *testing.T) {
	t.Parallel()

	_, err := b.FlexibleLoanRepay(context.Background(), currency.EMPTYCODE, currency.BTC, 1, false, false)
	require.ErrorIs(t, err, errLoanCoinMustBeSet)
	_, err = b.FlexibleLoanRepay(context.Background(), currency.USDT, currency.EMPTYCODE, 1, false, false)
	require.ErrorIs(t, err, errCollateralCoinMustBeSet)
	_, err = b.FlexibleLoanRepay(context.Background(), currency.USDT, currency.BTC, 0, false, false)
	require.ErrorIs(t, err, errAmountMustBeSet)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.FlexibleLoanRepay(context.Background(), currency.ATOM, currency.USDC, 1, false, false)
	assert.NoError(t, err)
}

func TestFlexibleLoanRepayHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleLoanRepayHistory(context.Background(), currency.EMPTYCODE, currency.EMPTYCODE, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestFlexibleLoanAdjustLTV(t *testing.T) {
	t.Parallel()
	_, err := b.FlexibleLoanAdjustLTV(context.Background(), currency.EMPTYCODE, currency.BTC, 1, true)
	require.ErrorIs(t, err, errLoanCoinMustBeSet)
	_, err = b.FlexibleLoanAdjustLTV(context.Background(), currency.USDT, currency.EMPTYCODE, 1, true)
	require.ErrorIs(t, err, errCollateralCoinMustBeSet)

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.FlexibleLoanAdjustLTV(context.Background(), currency.USDT, currency.BTC, 1, true)
	assert.NoError(t, err)
}

func TestFlexibleLoanLTVAdjustmentHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleLoanLTVAdjustmentHistory(context.Background(), currency.EMPTYCODE, currency.EMPTYCODE, time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestFlexibleLoanAssetsData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleLoanAssetsData(context.Background(), currency.EMPTYCODE)
	assert.NoError(t, err)
}

func TestFlexibleCollateralAssetsData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FlexibleCollateralAssetsData(context.Background(), currency.EMPTYCODE)
	assert.NoError(t, err)
}

func TestGetFuturesContractDetails(t *testing.T) {
	t.Parallel()
	_, err := b.GetFuturesContractDetails(context.Background(), asset.Spot)
	assert.ErrorIs(t, err, futures.ErrNotFuturesAsset)
	_, err = b.GetFuturesContractDetails(context.Background(), asset.Futures)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	_, err = b.GetFuturesContractDetails(context.Background(), asset.USDTMarginedFutures)
	assert.NoError(t, err)
	_, err = b.GetFuturesContractDetails(context.Background(), asset.CoinMarginedFutures)
	assert.NoError(t, err)
}

func TestGetFundingRateInfo(t *testing.T) {
	t.Parallel()
	_, err := b.GetFundingRateInfo(context.Background())
	assert.NoError(t, err)
}

func TestUGetFundingRateInfo(t *testing.T) {
	t.Parallel()
	_, err := b.UGetFundingRateInfo(context.Background())
	assert.NoError(t, err)
}

func TestWsUFuturesConnect(t *testing.T) {
	t.Parallel()
	err := b.WsUFuturesConnect()
	require.NoError(t, err)
}

var messageMap = map[string]string{
	"Asset Index":                   `{"stream": "!assetIndex@arr", "data": [{ "e":"assetIndexUpdate", "E":1686749230000, "s":"ADAUSD", "i":"0.27462452", "b":"0.10000000", "a":"0.10000000", "B":"0.24716207", "A":"0.30208698", "q":"0.05000000", "g":"0.05000000", "Q":"0.26089330", "G":"0.28835575" }, { "e":"assetIndexUpdate", "E":1686749230000, "s":"USDTUSD", "i":"0.99987691", "b":"0.00010000", "a":"0.00010000", "B":"0.99977692", "A":"0.99997689", "q":"0.00010000", "g":"0.00010000", "Q":"0.99977692", "G":"0.99997689" } ]}`,
	"Contract Info":                 `{"stream": "!contractInfo", "data": {"e":"contractInfo", "E":1669356423908, "s":"IOTAUSDT", "ps":"IOTAUSDT", "ct":"PERPETUAL", "dt":4133404800000, "ot":1569398400000, "cs":"TRADING", "bks":[ { "bs":1, "bnf":0, "bnc":5000, "mmr":0.01, "cf":0, "mi":21, "ma":50 }, { "bs":2, "bnf":5000, "bnc":25000, "mmr":0.025, "cf":75, "mi":11, "ma":20 } ] }}`,
	"Force Order":                   `{"stream": "!forceOrder@arr", "data": {"e":"forceOrder", "E":1568014460893, "o":{ "s":"BTCUSDT", "S":"SELL", "o":"LIMIT", "f":"IOC", "q":"0.014", "p":"9910", "ap":"9910", "X":"FILLED", "l":"0.014", "z":"0.014", "T":1568014460893 }}}`,
	"All BookTicker":                `{"stream": "!bookTicker","data":{"e":"bookTicker","u":3682854202063,"s":"NEARUSDT","b":"2.4380","B":"20391","a":"2.4390","A":"271","T":1703015198639,"E":1703015198640}}`,
	"Multiple Market Ticker":        `{"stream": "!ticker@arr", "data": [{"e":"24hrTicker","E":1703018247910,"s":"ICPUSDT","p":"-0.540000","P":"-5.395","w":"9.906194","c":"9.470000","Q":"1","o":"10.010000","h":"10.956000","l":"9.236000","v":"34347035","q":"340248403.001000","O":1702931820000,"C":1703018247909,"F":78723309,"L":80207941,"n":1484628},{"e":"24hrTicker","E":1703018247476,"s":"MEMEUSDT","p":"0.0020900","P":"7.331","w":"0.0300554","c":"0.0305980","Q":"7568","o":"0.0285080","h":"0.0312730","l":"0.0284120","v":"5643663185","q":"169622568.3721920","O":1702931820000,"C":1703018247475,"F":88665791,"L":89517438,"n":851643},{"e":"24hrTicker","E":1703018247822,"s":"SOLUSDT","p":"0.8680","P":"1.192","w":"74.4933","c":"73.6900","Q":"21","o":"72.8220","h":"76.3840","l":"71.8000","v":"26283647","q":"1957955612.4830","O":1702931820000,"C":1703018247820,"F":1126774871,"L":1129007642,"n":2232761},{"e":"24hrTicker","E":1703018247254,"s":"IMXUSDT","p":"0.0801","P":"3.932","w":"2.1518","c":"2.1171","Q":"225","o":"2.0370","h":"2.2360","l":"2.0319","v":"59587050","q":"128216496.4538","O":1702931820000,"C":1703018247252,"F":169814879,"L":170587124,"n":772246},{"e":"24hrTicker","E":1703018247309,"s":"DYDXUSDT","p":"-0.036","P":"-1.255","w":"2.896","c":"2.832","Q":"169.6","o":"2.868","h":"2.987","l":"2.782","v":"81690098.5","q":"236599791.383","O":1702931820000,"C":1703018247308,"F":385238821,"L":385888621,"n":649799},{"e":"24hrTicker","E":1703018247240,"s":"ONTUSDT","p":"0.0022","P":"1.011","w":"0.2213","c":"0.2197","Q":"45.7","o":"0.2175","h":"0.2251","l":"0.2157","v":"60880132.6","q":"13471239.8637","O":1702931820000,"C":1703018247238,"F":186008331,"L":186088275,"n":79945},{"e":"24hrTicker","E":1703018247658,"s":"AAVEUSDT","p":"4.660","P":"4.778","w":"102.969","c":"102.190","Q":"0.4","o":"97.530","h":"108.000","l":"97.370","v":"1205430.6","q":"124121750.870","O":1702931820000,"C":1703018247657,"F":343017862,"L":343487276,"n":469414},{"e":"24hrTicker","E":1703018247545,"s":"USTCUSDT","p":"0.0018500","P":"5.628","w":"0.0348991","c":"0.0347200","Q":"2316","o":"0.0328700","h":"0.0371100","l":"0.0328000","v":"2486985654","q":"86793545.3903700","O":1702931820000,"C":1703018247544,"F":32136013,"L":32601947,"n":465935},{"e":"24hrTicker","E":1703018247997,"s":"FTMUSDT","p":"-0.005000","P":"-1.221","w":"0.409721","c":"0.404400","Q":"1421","o":"0.409400","h":"0.421200","l":"0.392100","v":"471077518","q":"193010517.884400","O":1702931820000,"C":1703018247996,"F":716077491,"L":716712548,"n":635055},{"e":"24hrTicker","E":1703018247338,"s":"LRCUSDT","p":"-0.00290","P":"-1.104","w":"0.26531","c":"0.25980","Q":"113","o":"0.26270","h":"0.27190","l":"0.25590","v":"142488749","q":"37803477.10260","O":1702931820000,"C":1703018247336,"F":318115460,"L":318317340,"n":201880},{"e":"24hrTicker","E":1703018247776,"s":"TRBUSDT","p":"25.037","P":"21.840","w":"131.860","c":"139.677","Q":"0.3","o":"114.640","h":"143.900","l":"113.600","v":"3955845.0","q":"521616257.947","O":1702931820000,"C":1703018247775,"F":417041483,"L":419226886,"n":2185249},{"e":"24hrTicker","E":1703018247513,"s":"ACEUSDT","p":"0.108200","P":"0.826","w":"13.544944","c":"13.211400","Q":"14.37","o":"13.103200","h":"15.131200","l":"12.402900","v":"41359842.25","q":"560216757.038015","O":1702931820000,"C":1703018247512,"F":2261106,"L":4779982,"n":2518828},{"e":"24hrTicker","E":1703018247995,"s":"KEYUSDT","p":"0.0000270","P":"0.506","w":"0.0054583","c":"0.0053660","Q":"3540","o":"0.0053390","h":"0.0056230","l":"0.0053220","v":"1658962254","q":"9055176.9144700","O":1702931820000,"C":1703018247993,"F":32127330,"L":32236546,"n":109217},{"e":"24hrTicker","E":1703018247825,"s":"SUIUSDT","p":"0.094400","P":"15.783","w":"0.658766","c":"0.692500","Q":"157.6","o":"0.598100","h":"0.719600","l":"0.596400","v":"538807943.2","q":"354948524.988570","O":1702931820000,"C":1703018247824,"F":129572611,"L":130637476,"n":1064863},{"e":"24hrTicker","E":1703018247328,"s":"AGLDUSDT","p":"0.0738000","P":"7.016","w":"1.1222224","c":"1.1257000","Q":"49","o":"1.0519000","h":"1.1936000","l":"1.0471000","v":"63230369","q":"70958539.3508000","O":1702931820000,"C":1703018247327,"F":40498492,"L":41170995,"n":672503},{"e":"24hrTicker","E":1703018247882,"s":"BTCUSDT","p":"412.30","P":"0.986","w":"42651.76","c":"42247.00","Q":"0.003","o":"41834.70","h":"43550.00","l":"41792.00","v":"366582.423","q":"15635385730.76","O":1702931820000,"C":1703018247880,"F":4392041494,"L":4395950440,"n":3908934},{"e":"24hrTicker","E":1703018247531,"s":"WLDUSDT","p":"-0.0475000","P":"-1.232","w":"3.9879959","c":"3.8089000","Q":"50","o":"3.8564000","h":"4.3320000","l":"3.7237000","v":"119350666","q":"475969966.2747000","O":1702931820000,"C":1703018247530,"F":183723717,"L":186154953,"n":2431230},{"e":"24hrTicker","E":1703018247595,"s":"WAVESUSDT","p":"0.1108","P":"4.876","w":"2.4490","c":"2.3833","Q":"8.1","o":"2.2725","h":"2.5775","l":"2.2658","v":"54051344.0","q":"132369622.6356","O":1702931820000,"C":1703018247593,"F":503343992,"L":504167968,"n":823975},{"e":"24hrTicker","E":1703018247943,"s":"BLZUSDT","p":"0.00441","P":"1.274","w":"0.34477","c":"0.35043","Q":"35","o":"0.34602","h":"0.35844","l":"0.33146","v":"224686045","q":"77465133.09517","O":1702931820000,"C":1703018247942,"F":301286442,"L":301919432,"n":632991},{"e":"24hrTicker","E":1703018248027,"s":"ALGOUSDT","p":"0.0044","P":"2.329","w":"0.1982","c":"0.1933","Q":"1724.4","o":"0.1889","h":"0.2053","l":"0.1883","v":"418107041.7","q":"82860752.3534","O":1702931820000,"C":1703018248025,"F":317274252,"L":317530189,"n":255937},{"e":"24hrTicker","E":1703018247795,"s":"LUNA2USDT","p":"0.0849000","P":"9.610","w":"0.9622720","c":"0.9684000","Q":"91","o":"0.8835000","h":"1.0234000","l":"0.8800000","v":"132211955","q":"127223857.1990000","O":1702931820000,"C":1703018247793,"F":143814989,"L":144504341,"n":689350},{"e":"24hrTicker","E":1703018247557,"s":"DOGEUSDT","p":"-0.000290","P":"-0.320","w":"0.091710","c":"0.090210","Q":"1211","o":"0.090500","h":"0.093550","l":"0.089300","v":"4695249277","q":"430603554.425970","O":1702931820000,"C":1703018247556,"F":1408300026,"L":1409042131,"n":742103},{"e":"24hrTicker","E":1703018247578,"s":"SUSHIUSDT","p":"0.0024","P":"0.217","w":"1.1263","c":"1.1097","Q":"34","o":"1.1073","h":"1.1479","l":"1.0921","v":"34830643","q":"39229338.9293","O":1702931820000,"C":1703018247576,"F":389676753,"L":389892337,"n":215584},{"e":"24hrTicker","E":1703018247636,"s":"ROSEUSDT","p":"0.00859","P":"9.826","w":"0.09344","c":"0.09601","Q":"300","o":"0.08742","h":"0.09842","l":"0.08724","v":"768803655","q":"71837497.60153","O":1702931820000,"C":1703018247635,"F":145874088,"L":146347778,"n":473689},{"e":"24hrTicker","E":1703018247446,"s":"CTKUSDT","p":"0.05240","P":"6.933","w":"0.76993","c":"0.80820","Q":"16","o":"0.75580","h":"0.81760","l":"0.73560","v":"39275735","q":"30239750.04250","O":1702931820000,"C":1703018247445,"F":129601557,"L":129911270,"n":309714},{"e":"24hrTicker","E":1703018247083,"s":"MATICUSDT","p":"-0.02260","P":"-2.883","w":"0.78657","c":"0.76130","Q":"11","o":"0.78390","h":"0.82380","l":"0.74930","v":"510723474","q":"401719478.20480","O":1702931820000,"C":1703018247081,"F":899425701,"L":900164133,"n":738432},{"e":"24hrTicker","E":1703018247954,"s":"INJUSDT","p":"3.554000","P":"10.740","w":"37.577625","c":"36.646000","Q":"9.3","o":"33.092000","h":"39.988000","l":"32.803000","v":"30119373.7","q":"1131814520.584100","O":1702931820000,"C":1703018247953,"F":210846748,"L":214612851,"n":3766053},{"e":"24hrTicker","E":1703018247559,"s":"OCEANUSDT","p":"0.00890","P":"1.805","w":"0.50791","c":"0.50200","Q":"147","o":"0.49310","h":"0.52090","l":"0.49170","v":"42754656","q":"21715597.51239","O":1702931820000,"C":1703018247557,"F":243729859,"L":243879437,"n":149578},{"e":"24hrTicker","E":1703018247779,"s":"UNIUSDT","p":"0.0220","P":"0.378","w":"5.9288","c":"5.8470","Q":"10","o":"5.8250","h":"6.0440","l":"5.7520","v":"11324960","q":"67143423.4300","O":1702931820000,"C":1703018247778,"F":356204442,"L":356430119,"n":225678},{"e":"24hrTicker","E":1703018247999,"s":"1000BONKUSDT","p":"-0.0004410","P":"-2.245","w":"0.0205588","c":"0.0192000","Q":"1562","o":"0.0196410","h":"0.0231060","l":"0.0188770","v":"30632634003","q":"629769968.4590968","O":1702931820000,"C":1703018247998,"F":75958362,"L":80131721,"n":4173351},{"e":"24hrTicker","E":1703018247559,"s":"ARUSDT","p":"-0.382","P":"-4.176","w":"9.030","c":"8.765","Q":"4.5","o":"9.147","h":"9.467","l":"8.571","v":"3178087.5","q":"28698158.147","O":1702931820000,"C":1703018247557,"F":143756455,"L":143985699,"n":229244},{"e":"24hrTicker","E":1703018247344,"s":"AUCTIONUSDT","p":"9.690000","P":"31.369","w":"38.302392","c":"40.580000","Q":"0.65","o":"30.890000","h":"43.400000","l":"30.650000","v":"15656989.13","q":"599700134.856300","O":1702931820000,"C":1703018247343,"F":2451094,"L":5013398,"n":2561767},{"e":"24hrTicker","E":1703018247959,"s":"XRPUSDT","p":"-0.0021","P":"-0.346","w":"0.6083","c":"0.6045","Q":"396.3","o":"0.6066","h":"0.6170","l":"0.5973","v":"744301855.7","q":"452752948.7478","O":1702931820000,"C":1703018247957,"F":1344388341,"L":1344913573,"n":525224},{"e":"24hrTicker","E":1703018247813,"s":"EGLDUSDT","p":"-0.130","P":"-0.223","w":"58.569","c":"58.070","Q":"1.1","o":"58.200","h":"60.240","l":"56.670","v":"802381.7","q":"46994956.463","O":1702931820000,"C":1703018247811,"F":235206699,"L":235456030,"n":249331},{"e":"24hrTicker","E":1703018247990,"s":"ETHUSDT","p":"-10.21","P":"-0.468","w":"2206.89","c":"2170.39","Q":"0.060","o":"2180.60","h":"2256.64","l":"2135.03","v":"3187161.031","q":"7033700225.77","O":1702931820000,"C":1703018247988,"F":3443398114,"L":3446512406,"n":3114283},{"e":"24hrTicker","E":1703018247096,"s":"PENDLEUSDT","p":"-0.0059000","P":"-0.569","w":"1.0590403","c":"1.0319000","Q":"12","o":"1.0378000","h":"1.0960000","l":"1.0120000","v":"7593669","q":"8042001.5937000","O":1702931820000,"C":1703018247095,"F":16663914,"L":16782530,"n":118617}]}`,
	"Single Market Ticker":          `{"stream": "<symbol>@ticker", "data": { "e": "24hrTicker", "E": 123456789, "s": "BTCUSDT", "p": "0.0015", "P": "250.00", "w": "0.0018", "c": "0.0025", "Q": "10", "o": "0.0010", "h": "0.0025", "l": "0.0010", "v": "10000", "q": "18", "O": 0, "C": 86400000, "F": 0, "L": 18150, "n": 18151 } }`,
	"Multiple Mini Tickers":         `{"stream": "!miniTicker@arr","data":[{"e":"24hrMiniTicker","E":1703019429455,"s":"BICOUSDT","c":"0.3667000","o":"0.3792000","h":"0.3892000","l":"0.3639000","v":"28768370","q":"10779000.9922000"},{"e":"24hrMiniTicker","E":1703019429985,"s":"API3USDT","c":"1.6834","o":"1.7326","h":"1.8406","l":"1.6699","v":"12371516.4","q":"21642153.0574"},{"e":"24hrMiniTicker","E":1703019429111,"s":"ICPUSDT","c":"9.414000","o":"10.126000","h":"10.956000","l":"9.236000","v":"34262192","q":"339148145.539000"},{"e":"24hrMiniTicker","E":1703019429945,"s":"SOLUSDT","c":"73.0930","o":"73.2180","h":"76.3840","l":"71.8000","v":"26319095","q":"1960871540.2620"}]}`,
	"Multi Asset Mode Asset":        `{"stream": "!assetIndex@arr", "data":[{ "e":"assetIndexUpdate", "E":1686749230000, "s":"ADAUSD","i":"0.27462452","b":"0.10000000","a":"0.10000000","B":"0.24716207","A":"0.30208698","q":"0.05000000","g":"0.05000000","Q":"0.26089330","G":"0.28835575"}, { "e":"assetIndexUpdate", "E":1686749230000, "s":"USDTUSD", "i":"0.99987691", "b":"0.00010000", "a":"0.00010000", "B":"0.99977692", "A":"0.99997689", "q":"0.00010000", "g":"0.00010000", "Q":"0.99977692", "G":"0.99997689" }]}`,
	"Composite Index Symbol":        `{"stream": "<symbol>@compositeIndex", "data":{ "e":"compositeIndex", "E":1602310596000, "s":"DEFIUSDT", "p":"554.41604065", "C":"baseAsset", "c":[ { "b":"BAL", "q":"USDT", "w":"1.04884844", "W":"0.01457800", "i":"24.33521021" }, { "b":"BAND", "q":"USDT" , "w":"3.53782729", "W":"0.03935200", "i":"7.26420084" } ] } }`,
	"Diff Book Depth Stream":        `{"stream": "<symbol>@depth@500ms", "data": { "e": "depthUpdate", "E": 123456789, "T": 123456788, "s": "BTCUSDT", "U": 157, "u": 160, "pu": 149, "b": [ [ "0.0024", "10" ] ], "a": [ [ "0.0026", "100" ] ] } }`,
	"Partial Book Depth Stream":     `{"stream": "<symbol>@depth<levels>", "data":{ "e": "depthUpdate", "E": 1571889248277, "T": 1571889248276, "s": "BTCUSDT", "U": 390497796, "u": 390497878, "pu": 390497794, "b": [ [ "7403.89", "0.002" ], [ "7403.90", "3.906" ], [ "7404.00", "1.428" ], [ "7404.85", "5.239" ], [ "7405.43", "2.562" ] ], "a": [ [ "7405.96", "3.340" ], [ "7406.63", "4.525" ], [ "7407.08", "2.475" ], [ "7407.15", "4.800" ], [ "7407.20","0.175"]]}}`,
	"Individual Symbol Mini Ticker": `{"stream": "<symbol>@miniTicker", "data": { "e": "24hrMiniTicker", "E": 123456789, "s": "BTCUSDT", "c": "0.0025", "o": "0.0010", "h": "0.0025", "l": "0.0010", "v": "10000", "q": "18"}}`,
}

func TestHandleData(t *testing.T) {
	t.Parallel()
	for x := range messageMap {
		err := b.wsHandleFuturesData([]byte(messageMap[x]), asset.USDTMarginedFutures)
		require.NoError(t, err)
	}
}

func TestListSubscriptions(t *testing.T) {
	t.Parallel()
	if !b.Websocket.IsConnected() {
		err := b.WsUFuturesConnect()
		require.NoError(t, err)
	}
	_, err := b.ListSubscriptions()
	require.NoError(t, err)
}

func TestSetProperty(t *testing.T) {
	t.Parallel()
	if !b.Websocket.IsConnected() {
		err := b.WsUFuturesConnect()
		require.NoError(t, err)
	}
	err := b.SetProperty("combined", true)
	require.NoError(t, err)
}

func TestWsConnect(t *testing.T) {
	t.Parallel()
	err := b.WsConnect()
	require.NoError(t, err)
}

func TestGetWsOrderbook(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsOrderbook(&OrderBookDataRequestParams{Symbol: currency.NewPair(currency.BTC, currency.USDT), Limit: 1000})
	assert.NoError(t, err)
}

func TestGetWsMostRecentTrades(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsMostRecentTrades(&RecentTradeRequestParams{
		Symbol: currency.NewPair(currency.BTC, currency.USDT),
		Limit:  15,
	})
	assert.NoError(t, err)
}

func TestGetWsAggregatedTrades(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsAggregatedTrades(&WsAggregateTradeRequestParams{
		Symbol: "BTCUSDT",
		Limit:  5,
	})
	require.NoError(t, err)
}

func TestGetWsKlines(t *testing.T) {
	t.Parallel()
	start, end := getTime()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsCandlestick(&KlinesRequestParams{
		Symbol:    currency.NewPair(currency.BTC, currency.USDT),
		Interval:  kline.FiveMin.Short(),
		Limit:     24,
		StartTime: start,
		EndTime:   end,
	})
	assert.NoError(t, err)
}

func TestGetWsOptimizedCandlestick(t *testing.T) {
	t.Parallel()
	start, end := getTime()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsOptimizedCandlestick(&KlinesRequestParams{
		Symbol:    currency.NewPair(currency.BTC, currency.USDT),
		Interval:  kline.FiveMin.Short(),
		Limit:     24,
		StartTime: start,
		EndTime:   end,
	})
	assert.NoError(t, err)
}

func setupWs() {
	err := b.WsConnect()
	if err != nil {
		log.Fatal(err)
	}
}

func TestGetCurrenctAveragePrice(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsCurrenctAveragePrice(currency.NewPair(currency.BTC, currency.USDT))
	assert.NoError(t, err)
}

func TestGetWs24HourPriceChanges(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWs24HourPriceChanges(&PriceChangeRequestParam{
		Symbols: []currency.Pair{currency.NewPair(currency.BTC, currency.USDT)},
	})
	assert.NoError(t, err)
}

func TestGetWsTradingDayTickers(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsTradingDayTickers(&PriceChangeRequestParam{
		Symbols: []currency.Pair{currency.NewPair(currency.BTC, currency.USDT)},
	})
	assert.NoError(t, err)
}

func TestGetSymbolPriceTicker(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetSymbolPriceTicker(currency.NewPair(currency.BTC, currency.USDT))
	require.NoError(t, err)
}

func TestGetWsSymbolOrderbookTicker(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, err := b.GetWsSymbolOrderbookTicker([]currency.Pair{currency.NewPair(currency.BTC, currency.USDT)})
	require.NoError(t, err)
	_, err = b.GetWsSymbolOrderbookTicker([]currency.Pair{
		currency.NewPair(currency.BTC, currency.USDT),
		currency.NewPair(currency.ETH, currency.USDT),
	})
	require.NoError(t, err)
}

func TestGetQuerySessionStatus(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetQuerySessionStatus()
	require.NoError(t, err)
}

func TestGetLogOutOfSession(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLogOutOfSession()
	require.NoError(t, err)
}

func TestPlaceNewOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsPlaceNewOrder(&TradeOrderRequestParam{
		Symbol:      "BTCUSDT",
		Side:        "SELL",
		OrderType:   "LIMIT",
		TimeInForce: "GTC",
		Price:       1234,
		Quantity:    1,
	})
	require.NoError(t, err)
}

func TestValidatePlaceNewOrderRequest(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.ValidatePlaceNewOrderRequest(&TradeOrderRequestParam{
		Symbol:      "BTCUSDT",
		Side:        "SELL",
		OrderType:   "LIMIT",
		TimeInForce: "GTC",
		Price:       1234,
		Quantity:    1,
	})
	require.NoError(t, err)
}

func TestWsQueryOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsQueryOrder(&QueryOrderParam{
		Symbol:  "BTCUSDT",
		OrderID: 12345,
	})
	require.NoError(t, err)
}

func TestSignRequest(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	_, signature, err := b.SignRequest(map[string]interface{}{
		"name": "nameValue",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, signature, "unexpected signature")
}

func TestWsCancelAndReplaceTradeOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsCancelAndReplaceTradeOrder(&WsCancelAndReplaceParam{
		Symbol:                    "BTCUSDT",
		CancelReplaceMode:         "ALLOW_FAILURE",
		CancelOriginClientOrderID: "4d96324ff9d44481926157",
		Side:                      "SELL",
		OrderType:                 "LIMIT",
		TimeInForce:               "GTC",
		Price:                     23416.10000000,
		Quantity:                  0.00847000,
	})
	require.NoError(t, err)
}

func TestWsCurrentOpenOrders(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsCurrentOpenOrders(currency.NewPair(currency.BTC, currency.USDT), 6000)
	require.NoError(t, err)
}

func TestWsCancelOpenOrders(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsCancelOpenOrders(currency.NewPair(currency.BTC, currency.USDT), 6000)
	require.NoError(t, err)
}

func TestWsPlaceOCOOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsPlaceOCOOrder(&PlaceOCOOrderParam{
		Symbol:               "BTCUSDT",
		Side:                 "SELL",
		Price:                23420.00000000,
		Quantity:             0.00650000,
		StopPrice:            23410.00000000,
		StopLimitPrice:       23405.00000000,
		StopLimitTimeInForce: "GTC",
		NewOrderRespType:     "RESULT",
	})
	require.NoError(t, err)
}

func TestWsQueryOCOOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsQueryOCOOrder("123456788", 0, 0)
	require.NoError(t, err)
}

func TestWsCancelOCOOrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsCancelOCOOrder(
		currency.NewPair(currency.BTC, currency.USDT), "someID", "12354", "", 0)
	require.NoError(t, err)
}

func TestWsCurrentOpenOCOOrders(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsCurrentOpenOCOOrders(0)
	require.NoError(t, err)
}

func TestWsPlaceNewSOROrder(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WsPlaceNewSOROrder(&WsOSRPlaceOrderParams{
		Symbol:      "BTCUSDT",
		Side:        "BUY",
		OrderType:   "LIMIT",
		Quantity:    0.5,
		TimeInForce: "GTC",
		Price:       31000,
	})
	require.NoError(t, err)
}

func TestWsTestNewOrderUsingSOR(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.WsTestNewOrderUsingSOR(&WsOSRPlaceOrderParams{
		Symbol:      "BTCUSDT",
		Side:        "BUY",
		OrderType:   "LIMIT",
		Quantity:    0.5,
		TimeInForce: "GTC",
		Price:       31000,
	})
	require.NoError(t, err)
}

func TestToMap(t *testing.T) {
	t.Parallel()
	input := &struct {
		Zebiba bool   `json:"zebiba"`
		Value  int64  `json:"value"`
		Abebe  string `json:"abebe"`
		Name   string `json:"name"`
	}{
		Name:  "theName",
		Value: 347,
	}
	_, err := b.ToMap(input)
	assert.NoError(t, err)
}

func TestSortingTest(t *testing.T) {
	params := map[string]interface{}{"apiKey": "wwhj3r3amR", "signature": "f89c6e5c0b", "timestamp": 1704873175325, "symbol": "BTCUSDT", "startTime": 1704009175325, "endTime": 1704873175325, "limit": 5}
	sortedKeys := []string{"apiKey", "endTime", "limit", "signature", "startTime", "symbol", "timestamp"}
	keys := SortMap(params)
	require.Len(t, keys, len(sortedKeys), "unexptected keys length")
	for a := range keys {
		require.Equal(t, keys[a], sortedKeys[a])
	}
}

func TestGetAccountInformation(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetWsAccountInfo(0)
	require.NoError(t, err)
}

func TestWsQueryAccountOrderRateLimits(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsQueryAccountOrderRateLimits(0)
	require.NoError(t, err)
}

func TestWsQueryAccountOrderHistory(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsQueryAccountOrderHistory(&AccountOrderRequestParam{
		Symbol:    "BTCUSDT",
		StartTime: time.Now().Add(-time.Hour * 24 * 10).UnixMilli(),
		EndTime:   time.Now().Add(-time.Hour * 6).UnixMilli(),
		Limit:     5,
	})
	require.NoError(t, err)
}

func TestWsQueryAccountOCOOrderHistory(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsQueryAccountOCOOrderHistory(0, 0, 0, time.Time{}, time.Time{})
	require.NoError(t, err)
}

func TestWsAccountTradeHistory(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsAccountTradeHistory(&AccountOrderRequestParam{
		Symbol:  "BTCUSDT",
		OrderID: 1234,
	})
	require.NoError(t, err)
}

func TestWsAccountPreventedMatches(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsAccountPreventedMatches(currency.NewPair(currency.BTC, currency.USDT), 1223456, 0, 0, 0, 0)
	require.NoError(t, err)
}

func TestWsAccountAllocation(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsAccountAllocation(currency.NewPair(currency.BTC, currency.USDT), time.Time{}, time.Now(), 0, 0, 0, 19)
	require.NoError(t, err)
}

func TestWsAccountCommissionRates(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsAccountCommissionRates(currency.NewPair(currency.BTC, currency.USDT))
	require.NoError(t, err)
}

func TestWsStartUserDataStream(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.WsStartUserDataStream()
	require.NoError(t, err)
}

func TestWsPingUserDataStream(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	err := b.WsPingUserDataStream("xs0mRXdAKlIPDRFrlPcw0qI41Eh3ixNntmymGyhrhgqo7L6FuLaWArTD7RLP")
	require.NoError(t, err)
}

func TestWsStopUserDataStream(t *testing.T) {
	t.Parallel()
	if !b.IsAPIStreamConnected() {
		t.Skip(apiStreamingIsNotConnected)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.WsStopUserDataStream("xs0mRXdAKlIPDRFrlPcw0qI41Eh3ixNntmymGyhrhgqo7L6FuLaWArTD7RLP")
	require.NoError(t, err)
}
func TestGetOpenInterest(t *testing.T) {
	t.Parallel()
	resp, err := b.GetOpenInterest(context.Background(), key.PairAsset{
		Base:  currency.BTC.Item,
		Quote: currency.USDT.Item,
		Asset: asset.USDTMarginedFutures,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	resp, err = b.GetOpenInterest(context.Background(), key.PairAsset{
		Base:  currency.NewCode("BTCUSD").Item,
		Quote: currency.PERP.Item,
		Asset: asset.CoinMarginedFutures,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	_, err = b.GetOpenInterest(context.Background(), key.PairAsset{
		Base:  currency.BTC.Item,
		Quote: currency.USDT.Item,
		Asset: asset.Spot,
	})
	assert.ErrorIs(t, err, asset.ErrNotSupported)
}

func TestWsOptionsConnect(t *testing.T) {
	t.Parallel()
	err := b.WsOptionsConnect()
	require.NoError(t, err)
	time.Sleep(time.Second * 23)
}

func TestGetOptionsExchangeInformation(t *testing.T) {
	t.Parallel()
	exchangeinformation, err := b.GetOptionsExchangeInformation(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, exchangeinformation)
}
