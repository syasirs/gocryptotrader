package kraken

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/key"
	"github.com/thrasher-corp/gocryptotrader/core"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fundingrate"
	"github.com/thrasher-corp/gocryptotrader/exchanges/futures"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	testexch "github.com/thrasher-corp/gocryptotrader/internal/testing/exchange"
	testsubs "github.com/thrasher-corp/gocryptotrader/internal/testing/subscriptions"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

var k *Kraken

// Please add your own APIkeys to do correct due diligence testing.
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

var (
	xbtusdPair  = currency.NewPair(currency.XBT, currency.USD)
	bcheurPair  = currency.NewPair(currency.BCH, currency.EUR)
	fxbtusdPair = currency.NewPairWithDelimiter("pi", "xbtusd", "_")
)

func TestMain(m *testing.M) {
	k = new(Kraken)
	if err := testexch.TestInstance(k); err != nil {
		log.Fatal(err)
	}
	if apiKey != "" && apiSecret != "" {
		k.API.AuthenticatedSupport = true
		k.SetCredentials(apiKey, apiSecret, "", "", "", "")
	}
	os.Exit(m.Run())
}

func TestUpdateTradablePairs(t *testing.T) {
	t.Parallel()
	testexch.UpdatePairsOnce(t, k)
}

func TestGetCurrentServerTime(t *testing.T) {
	t.Parallel()
	_, err := k.GetCurrentServerTime(context.Background())
	assert.NoError(t, err, "GetCurrentServerTime should not error")
}

func TestWrapperGetServerTime(t *testing.T) {
	t.Parallel()
	st, err := k.GetServerTime(context.Background(), asset.Spot)
	require.NoError(t, err, "GetServerTime should not error")
	assert.WithinRange(t, st, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), "ServerTime should be within a day of now")
}

// TestUpdateOrderExecutionLimits exercises UpdateOrderExecutionLimits and GetOrderExecutionLimits
func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()

	err := k.UpdateOrderExecutionLimits(context.Background(), asset.Spot)
	require.NoError(t, err, "UpdateOrderExecutionLimits must not error")
	for _, p := range []currency.Pair{
		currency.NewPair(currency.ETH, currency.USDT),
		currency.NewPair(currency.XBT, currency.USDT),
	} {
		limits, err := k.GetOrderExecutionLimits(asset.Spot, p)
		require.NoErrorf(t, err, "%s GetOrderExecutionLimits must not error", p)
		assert.Positivef(t, limits.PriceStepIncrementSize, "%s PriceStepIncrementSize should be positive", p)
		assert.Positivef(t, limits.MinimumBaseAmount, "%s MinimumBaseAmount should be positive", p)
	}
}

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	_, err := k.FetchTradablePairs(context.Background(), asset.Futures)
	assert.NoError(t, err, "FetchTradablePairs should not error")
}

func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	testexch.UpdatePairsOnce(t, k)

	_, err := k.UpdateTicker(context.Background(), xbtusdPair, asset.Spot)
	assert.NoError(t, err, "UpdateTicker should not error")

	_, err = k.UpdateTicker(context.Background(), fxbtusdPair, asset.Futures)
	assert.NoError(t, err, "UpdateTicker should not error")
}

func TestUpdateTickers(t *testing.T) {
	t.Parallel()

	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")

	testexch.UpdatePairsOnce(t, k)

	ap, err := k.GetAvailablePairs(asset.Spot)
	require.NoError(t, err, "GetAvailablePairs must not error")
	err = k.CurrencyPairs.StorePairs(asset.Spot, ap, true)
	require.NoError(t, err, "StorePairs must not error")

	err = k.UpdateTickers(context.Background(), asset.Spot)
	require.NoError(t, err, "Update Tickers must not error")

	for i := range ap {
		_, err = ticker.GetTicker(k.Name, ap[i], asset.Spot)
		assert.NoErrorf(t, err, "GetTicker should not error for %s", ap[i])
	}

	ap, err = k.GetAvailablePairs(asset.Futures)
	require.NoError(t, err, "GetAvailablePairs must not error")
	err = k.CurrencyPairs.StorePairs(asset.Futures, ap, true)
	require.NoError(t, err, "StorePairs must not error")

	err = k.UpdateTickers(context.Background(), asset.Futures)
	require.NoError(t, err, "Update Tickers must not error")

	for i := range ap {
		_, err = ticker.GetTicker(k.Name, ap[i], asset.Futures)
		assert.NoErrorf(t, err, "GetTicker should not error for %s", ap[i])
	}

	err = k.UpdateTickers(context.Background(), asset.Index)
	assert.ErrorIs(t, err, asset.ErrNotSupported, "UpdateTickers should error correctly for asset.Index")
}

func TestUpdateOrderbook(t *testing.T) {
	t.Parallel()
	sp, err := currency.NewPairFromString("BTCEUR")
	require.NoError(t, err, "NewPairFromString must not error")

	_, err = k.UpdateOrderbook(context.Background(), sp, asset.Spot)
	assert.NoError(t, err, "UpdateOrderbook should not error")

	_, err = k.UpdateOrderbook(context.Background(), fxbtusdPair, asset.Futures)
	assert.NoError(t, err, "UpdateOrderbook should not error")
}

func TestFuturesBatchOrder(t *testing.T) {
	t.Parallel()
	var data []PlaceBatchOrderData
	var tempData PlaceBatchOrderData
	tempData.PlaceOrderType = "meow"
	tempData.OrderID = "test123"
	tempData.Symbol = "pi_xbtusd"
	data = append(data, tempData)
	_, err := k.FuturesBatchOrder(context.Background(), data)
	assert.ErrorIs(t, err, errInvalidBatchOrderType, "FuturesBatchOrder should error correctly")

	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	data[0].PlaceOrderType = "cancel"
	_, err = k.FuturesBatchOrder(context.Background(), data)
	assert.NoError(t, err, "FuturesBatchOrder should not error")
}

func TestFuturesEditOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesEditOrder(context.Background(), "test123", "", 5.2, 1, 0)
	assert.NoError(t, err, "FuturesEditOrder should not error")
}

func TestFuturesSendOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesSendOrder(context.Background(), order.Limit, fxbtusdPair, "buy", "", "", "", true, 1, 1, 0.9)
	assert.NoError(t, err, "FuturesSendOrder should not error")
}

func TestFuturesCancelOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesCancelOrder(context.Background(), "test123", "")
	assert.NoError(t, err, "FuturesCancelOrder should not error")
}

func TestFuturesGetFills(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesGetFills(context.Background(), time.Now().Add(-time.Hour*24))
	assert.NoError(t, err, "FuturesGetFills should not error")
}

func TestFuturesTransfer(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesTransfer(context.Background(), "cash", "futures", "btc", 2)
	assert.NoError(t, err, "FuturesTransfer should not error")
}

func TestFuturesGetOpenPositions(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesGetOpenPositions(context.Background())
	assert.NoError(t, err, "FuturesGetOpenPositions should not error")
}

func TestFuturesNotifications(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesNotifications(context.Background())
	assert.NoError(t, err, "FuturesNotifications should not error")
}

func TestFuturesCancelAllOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesCancelAllOrders(context.Background(), fxbtusdPair)
	assert.NoError(t, err, "FuturesCancelAllOrders should not error")
}

func TestGetFuturesAccountData(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.GetFuturesAccountData(context.Background())
	assert.NoError(t, err, "GetFuturesAccountData should not error")
}

func TestFuturesCancelAllOrdersAfter(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesCancelAllOrdersAfter(context.Background(), 50)
	assert.NoError(t, err, "FuturesCancelAllOrdersAfter should not error")
}

func TestFuturesOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesOpenOrders(context.Background())
	assert.NoError(t, err, "FuturesOpenOrders should not error")
}

func TestFuturesRecentOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.FuturesRecentOrders(context.Background(), fxbtusdPair)
	assert.NoError(t, err, "FuturesRecentOrders should not error")
}

func TestFuturesWithdrawToSpotWallet(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesWithdrawToSpotWallet(context.Background(), "xbt", 5)
	assert.NoError(t, err, "FuturesWithdrawToSpotWallet should not error")
}

func TestFuturesGetTransfers(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	_, err := k.FuturesGetTransfers(context.Background(), time.Now().Add(-time.Hour*24))
	assert.NoError(t, err, "FuturesGetTransfers should not error")
}

func TestGetFuturesOrderbook(t *testing.T) {
	t.Parallel()
	cp, err := currency.NewPairFromString("FI_xbtusd_200925")
	require.NoError(t, err, "NewPairFromString must not error")

	_, err = k.GetFuturesOrderbook(context.Background(), cp)
	assert.NoError(t, err, "GetFuturesOrderbook should not error")
}

func TestGetFuturesMarkets(t *testing.T) {
	t.Parallel()
	_, err := k.GetInstruments(context.Background())
	assert.NoError(t, err, "GetInstruments should not error")
}

func TestGetFuturesTickers(t *testing.T) {
	t.Parallel()
	_, err := k.GetFuturesTickers(context.Background())
	assert.NoError(t, err, "GetFuturesTickers should not error")
}

func TestGetFuturesTradeHistory(t *testing.T) {
	t.Parallel()
	_, err := k.GetFuturesTradeHistory(context.Background(), fxbtusdPair, time.Now().Add(-time.Hour*24))
	assert.NoError(t, err, "GetFuturesTradeHistory should not error")
}

// TestGetAssets API endpoint test
func TestGetAssets(t *testing.T) {
	t.Parallel()
	_, err := k.GetAssets(context.Background())
	assert.NoError(t, err, "GetAssets should not error")
}

func TestSeedAssetTranslator(t *testing.T) {
	t.Parallel()

	err := k.SeedAssets(context.TODO())
	require.NoError(t, err, "SeedAssets must not error")

	for from, to := range map[string]string{"XBTUSD": "XXBTZUSD", "USD": "ZUSD", "XBT": "XXBT"} {
		assert.Equal(t, from, assetTranslator.LookupAltName(to), "LookupAltName should return the correct value")
		assert.Equal(t, to, assetTranslator.LookupCurrency(from), "LookupCurrency should return the correct value")
	}
}

func TestSeedAssets(t *testing.T) {
	t.Parallel()
	var a assetTranslatorStore
	assert.Empty(t, a.LookupAltName("ZUSD"), "LookupAltName on unseeded store should return empty")
	a.Seed("ZUSD", "USD")
	assert.Equal(t, "USD", a.LookupAltName("ZUSD"), "LookupAltName should return the correct value")
	a.Seed("ZUSD", "BLA")
	assert.Equal(t, "USD", a.LookupAltName("ZUSD"), "Store should ignore second reseed of existing currency")
}

func TestLookupCurrency(t *testing.T) {
	t.Parallel()
	var a assetTranslatorStore
	assert.Empty(t, a.LookupCurrency("USD"), "LookupCurrency on unseeded store should return empty")
	a.Seed("ZUSD", "USD")
	assert.Equal(t, "ZUSD", a.LookupCurrency("USD"), "LookupCurrency should return the correct value")
	assert.Empty(t, a.LookupCurrency("EUR"), "LookupCurrency should still not return an unseeded key")
}

// TestGetAssetPairs API endpoint test
func TestGetAssetPairs(t *testing.T) {
	t.Parallel()
	for _, v := range []string{"fees", "leverage", "margin", ""} {
		_, err := k.GetAssetPairs(context.Background(), []string{}, v)
		require.NoErrorf(t, err, "GetAssetPairs %s must not error", v)
	}
}

// TestGetTicker API endpoint test
func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := k.GetTicker(context.Background(), bcheurPair)
	assert.NoError(t, err, "GetTicker should not error")
}

// TestGetTickers API endpoint test
func TestGetTickers(t *testing.T) {
	t.Parallel()
	_, err := k.GetTickers(context.Background(), "LTCUSD,ETCUSD")
	assert.NoError(t, err, "GetTickers should not error")
}

// TestGetOHLC API endpoint test
func TestGetOHLC(t *testing.T) {
	t.Parallel()
	_, err := k.GetOHLC(context.Background(), currency.NewPairWithDelimiter("XXBT", "ZUSD", ""), "1440")
	assert.NoError(t, err, "GetOHLC should not error")
}

// TestGetDepth API endpoint test
func TestGetDepth(t *testing.T) {
	t.Parallel()
	_, err := k.GetDepth(context.Background(), bcheurPair)
	assert.NoError(t, err, "GetDepth should not error")
}

// TestGetTrades API endpoint test
func TestGetTrades(t *testing.T) {
	t.Parallel()
	testexch.UpdatePairsOnce(t, k)
	_, err := k.GetTrades(context.Background(), bcheurPair)
	assert.NoError(t, err, "GetTrades should not error")

	_, err = k.GetTrades(context.Background(), currency.NewPairWithDelimiter("XXX", "XXX", ""))
	assert.ErrorContains(t, err, "Unknown asset pair", "GetDepth should error correctly")
}

// TestGetSpread API endpoint test
func TestGetSpread(t *testing.T) {
	t.Parallel()
	_, err := k.GetSpread(context.Background(), bcheurPair)
	assert.NoError(t, err, "GetSpread should not error")
}

// TestGetBalance API endpoint test
func TestGetBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.GetBalance(context.Background())
	assert.NoError(t, err, "GetBalance should not error")
}

// TestGetTradeBalance API endpoint test
func TestGetDepositMethods(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.GetDepositMethods(context.Background(), "USDT")
	assert.NoError(t, err, "GetDepositMethods should not error")
}

// TestGetTradeBalance API endpoint test
func TestGetTradeBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	args := TradeBalanceOptions{Asset: "ZEUR"}
	_, err := k.GetTradeBalance(context.Background(), args)
	assert.NoError(t, err)
}

// TestGetOpenOrders API endpoint test
func TestGetOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	args := OrderInfoOptions{Trades: true}
	_, err := k.GetOpenOrders(context.Background(), args)
	assert.NoError(t, err)
}

// TestGetClosedOrders API endpoint test
func TestGetClosedOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	args := GetClosedOrdersOptions{Trades: true, Start: "OE4KV4-4FVQ5-V7XGPU"}
	_, err := k.GetClosedOrders(context.Background(), args)
	assert.NoError(t, err)
}

// TestQueryOrdersInfo API endpoint test
func TestQueryOrdersInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	args := OrderInfoOptions{Trades: true}
	_, err := k.QueryOrdersInfo(context.Background(), args, "OR6ZFV-AA6TT-CKFFIW", "OAMUAJ-HLVKG-D3QJ5F")
	assert.NoError(t, err)
}

// TestGetTradesHistory API endpoint test
func TestGetTradesHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	args := GetTradesHistoryOptions{Trades: true, Start: "TMZEDR-VBJN2-NGY6DX", End: "TVRXG2-R62VE-RWP3UW"}
	_, err := k.GetTradesHistory(context.Background(), args)
	assert.NoError(t, err)
}

// TestQueryTrades API endpoint test
func TestQueryTrades(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.QueryTrades(context.Background(), true, "TMZEDR-VBJN2-NGY6DX", "TFLWIB-KTT7L-4TWR3L", "TDVRAH-2H6OS-SLSXRX")
	assert.NoError(t, err)
}

// TestOpenPositions API endpoint test
func TestOpenPositions(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.OpenPositions(context.Background(), false)
	assert.NoError(t, err)
}

// TestGetLedgers API endpoint test
// TODO: Needs a positive test
func TestGetLedgers(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	args := GetLedgersOptions{Start: "LRUHXI-IWECY-K4JYGO", End: "L5NIY7-JZQJD-3J4M2V", Ofs: 15}
	_, err := k.GetLedgers(context.Background(), args)
	assert.ErrorContains(t, err, "EQuery:Unknown asset pair", "GetLedger should error on imaginary ledgers")
}

// TestQueryLedgers API endpoint test
func TestQueryLedgers(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.QueryLedgers(context.Background(), "LVTSFS-NHZVM-EXNZ5M")
	assert.NoError(t, err)
}

// TestGetTradeVolume API endpoint test
func TestGetTradeVolume(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.GetTradeVolume(context.Background(), true, bcheurPair)
	assert.NoError(t, err, "GetTradeVolume should not error")
}

// TestOrders Tests AddOrder and CancelExistingOrder
func TestOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)

	// EGeneral:Invalid arguments:volume
	t.Skip("AddOrder test is a known failure due to volume parameter")

	args := AddOrderOptions{OrderFlags: "fcib"}
	cp, err := currency.NewPairFromString("XXBTZUSD")
	assert.NoError(t, err, "NewPairFromString should not error")
	var resp AddOrderResponse
	resp, err = k.AddOrder(context.Background(),
		cp,
		order.Sell.Lower(), order.Limit.Lower(),
		0.00000001, 0, 0, 0, &args)

	if assert.NoError(t, err, "AddOrder should not error") {
		if assert.Len(t, resp.TransactionIDs, 1, "One TransactionId should be returned") {
			id := resp.TransactionIDs[0]
			_, err = k.CancelExistingOrder(context.Background(), id)
			assert.NoErrorf(t, err, "CancelExistingOrder should not error, Please ensure order %s is cancelled manually", id)
		}
	}
}

// TestCancelExistingOrder API endpoint test
func TestCancelExistingOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)
	_, err := k.CancelExistingOrder(context.Background(), "OAVY7T-MV5VK-KHDF5X")
	if assert.Error(t, err, "Cancel with imaginary order-id should error") {
		assert.ErrorContains(t, err, "EOrder:Unknown order", "Cancel with imaginary order-id should error Unknown Order")
	}
}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:              1,
		FeeType:             exchange.CryptocurrencyTradeFee,
		Pair:                currency.NewPair(currency.XXBT, currency.ZUSD),
		PurchasePrice:       1,
		FiatCurrency:        currency.USD,
		BankTransactionType: exchange.WireTransfer,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	t.Parallel()
	var feeBuilder = setFeeBuilder()
	f, err := k.GetFeeByType(context.Background(), feeBuilder)
	require.NoError(t, err, "GetFeeByType must not error")
	assert.Positive(t, f, "GetFeeByType should return a positive value")
	if !sharedtestvalues.AreAPICredentialsSet(k) {
		assert.Equal(t, exchange.OfflineTradeFee, feeBuilder.FeeType, "GetFeeByType should set FeeType correctly")
	} else {
		assert.Equal(t, exchange.CryptocurrencyTradeFee, feeBuilder.FeeType, "GetFeeByType should set FeeType correctly")
	}
}

// TestGetFee exercises GetFee
func TestGetFee(t *testing.T) {
	t.Parallel()
	var feeBuilder = setFeeBuilder()

	if sharedtestvalues.AreAPICredentialsSet(k) {
		_, err := k.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err, "CryptocurrencyTradeFee Basic GetFee should not error")

		feeBuilder = setFeeBuilder()
		feeBuilder.Amount = 1000
		feeBuilder.PurchasePrice = 1000
		_, err = k.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err, "CryptocurrencyTradeFee High quantity GetFee should not error")

		feeBuilder = setFeeBuilder()
		feeBuilder.IsMaker = true
		_, err = k.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err, "CryptocurrencyTradeFee IsMaker GetFee should not error")

		feeBuilder = setFeeBuilder()
		feeBuilder.PurchasePrice = -1000
		_, err = k.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err, "CryptocurrencyTradeFee Negative purchase price GetFee should not error")

		feeBuilder = setFeeBuilder()
		feeBuilder.FeeType = exchange.InternationalBankDepositFee
		_, err = k.GetFee(context.Background(), feeBuilder)
		assert.NoError(t, err, "InternationalBankDepositFee Basic GetFee should not error")
	}

	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyDepositFee
	feeBuilder.Pair.Base = currency.XXBT
	_, err := k.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err, "CryptocurrencyDepositFee Basic GetFee should not error")

	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	_, err = k.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err, "CryptocurrencyWithdrawalFee Basic GetFee should not error")

	feeBuilder = setFeeBuilder()
	feeBuilder.Pair.Base = currency.NewCode("hello")
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	_, err = k.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err, "CryptocurrencyWithdrawalFee Invalid currency GetFee should not error")

	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	_, err = k.GetFee(context.Background(), feeBuilder)
	assert.NoError(t, err, "InternationalBankWithdrawalFee Basic GetFee should not error")
}

// TestFormatWithdrawPermissions logic test
func TestFormatWithdrawPermissions(t *testing.T) {
	t.Parallel()
	exp := exchange.AutoWithdrawCryptoWithSetupText + " & " + exchange.WithdrawCryptoWith2FAText + " & " + exchange.AutoWithdrawFiatWithSetupText + " & " + exchange.WithdrawFiatWith2FAText
	withdrawPermissions := k.FormatWithdrawPermissions()
	assert.Equal(t, exp, withdrawPermissions, "FormatWithdrawPermissions should return correct value")
}

// TestGetActiveOrders wrapper test
func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	pair, err := currency.NewPairFromString("LTC_USDT")
	assert.NoError(t, err, "NewPairFromString should not error")
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Pairs:     currency.Pairs{pair},
		Side:      order.AnySide,
	}

	_, err = k.GetActiveOrders(context.Background(), &getOrdersRequest)
	assert.NoError(t, err, "GetActiveOrders should not error")
}

// TestGetOrderHistory wrapper test
func TestGetOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.AnySide,
	}

	_, err := k.GetOrderHistory(context.Background(), &getOrdersRequest)
	assert.NoError(t, err)
}

// TestGetOrderInfo exercises GetOrderInfo
func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	_, err := k.GetOrderInfo(context.Background(), "OZPTPJ-HVYHF-EDIGXS", currency.EMPTYPAIR, asset.Spot)
	assert.ErrorContains(t, err, "order OZPTPJ-HVYHF-EDIGXS not found in response", "Should error that order was not found in response")
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------

// TestSubmitOrder wrapper test
func TestSubmitOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	var orderSubmission = &order.Submit{
		Exchange:  k.Name,
		Pair:      xbtusdPair,
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1,
		ClientID:  "meowOrder",
		AssetType: asset.Spot,
	}
	response, err := k.SubmitOrder(context.Background(), orderSubmission)
	if sharedtestvalues.AreAPICredentialsSet(k) {
		assert.NoError(t, err, "SubmitOrder should not error")
		assert.Equal(t, order.New, response.Status, "SubmitOrder should return a New order status")
	} else {
		assert.ErrorIs(t, err, exchange.ErrAuthenticationSupportNotEnabled, "SubmitOrder should error correctly")
	}
}

// TestCancelExchangeOrder wrapper test
func TestCancelExchangeOrder(t *testing.T) {
	t.Parallel()

	err := k.CancelOrder(context.Background(), &order.Cancel{
		AssetType: asset.Options,
		OrderID:   "1337",
	})
	assert.ErrorIs(t, err, asset.ErrNotSupported, "CancelOrder should error on Options asset")

	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	var orderCancellation = &order.Cancel{
		OrderID:   "OGEX6P-B5Q74-IGZ72R",
		AssetType: asset.Spot,
	}

	err = k.CancelOrder(context.Background(), orderCancellation)
	if sharedtestvalues.AreAPICredentialsSet(k) {
		assert.NoError(t, err, "CancelOrder should not error")
	} else {
		assert.ErrorIs(t, err, exchange.ErrAuthenticationSupportNotEnabled, "CancelOrder should error correctly")
	}
}

// TestCancelExchangeOrder wrapper test
func TestCancelBatchExchangeOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	pair := currency.Pair{
		Delimiter: "/",
		Base:      currency.BTC,
		Quote:     currency.USD,
	}

	var ordersCancellation []order.Cancel
	ordersCancellation = append(ordersCancellation, order.Cancel{
		Pair:      pair,
		OrderID:   "OGEX6P-B5Q74-IGZ72R,OGEX6P-B5Q74-IGZ722",
		AssetType: asset.Spot,
	})

	_, err := k.CancelBatchOrders(context.Background(), ordersCancellation)
	if sharedtestvalues.AreAPICredentialsSet(k) {
		assert.NoError(t, err, "CancelBatchOrder should not error")
	} else {
		assert.ErrorIs(t, err, exchange.ErrAuthenticationSupportNotEnabled, "CancelBatchOrders should error correctly")
	}
}

// TestCancelAllExchangeOrders wrapper test
func TestCancelAllExchangeOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	resp, err := k.CancelAllOrders(context.Background(), &order.Cancel{AssetType: asset.Spot})

	if sharedtestvalues.AreAPICredentialsSet(k) {
		assert.NoError(t, err, "CancelAllOrders should not error")
	} else {
		assert.ErrorIs(t, err, exchange.ErrAuthenticationSupportNotEnabled, "CancelBatchOrders should error correctly")
	}

	assert.Empty(t, resp.Status, "CancelAllOrders Status should not contain any failed order errors")
}

// TestUpdateAccountInfo exercises UpdateAccountInfo
func TestUpdateAccountInfo(t *testing.T) {
	t.Parallel()

	for _, a := range []asset.Item{asset.Spot, asset.Futures} {
		_, err := k.UpdateAccountInfo(context.Background(), a)

		if sharedtestvalues.AreAPICredentialsSet(k) {
			assert.NoErrorf(t, err, "UpdateAccountInfo should not error for asset %s", a) // Note Well: Spot and Futures have separate api keys
		} else {
			assert.ErrorIsf(t, err, exchange.ErrAuthenticationSupportNotEnabled, "UpdateAccountInfo should error correctly for asset %s", a)
		}
	}
}

// TestModifyOrder wrapper test
func TestModifyOrder(t *testing.T) {
	t.Parallel()

	_, err := k.ModifyOrder(context.Background(), &order.Modify{AssetType: asset.Spot})
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported, "ModifyOrder should error correctly")
}

// TestWithdraw wrapper test
func TestWithdraw(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	withdrawCryptoRequest := withdraw.Request{
		Exchange: k.Name,
		Crypto: withdraw.CryptoRequest{
			Address: core.BitcoinDonationAddress,
		},
		Amount:        -1,
		Currency:      currency.XXBT,
		Description:   "WITHDRAW IT ALL",
		TradePassword: "Key",
	}

	_, err := k.WithdrawCryptocurrencyFunds(context.Background(),
		&withdrawCryptoRequest)
	if !sharedtestvalues.AreAPICredentialsSet(k) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(k) && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

// TestWithdrawFiat wrapper test
func TestWithdrawFiat(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	var withdrawFiatRequest = withdraw.Request{
		Amount:        -1,
		Currency:      currency.EUR,
		Description:   "WITHDRAW IT ALL",
		TradePassword: "someBank",
	}

	_, err := k.WithdrawFiatFunds(context.Background(), &withdrawFiatRequest)
	if !sharedtestvalues.AreAPICredentialsSet(k) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(k) && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

// TestWithdrawInternationalBank wrapper test
func TestWithdrawInternationalBank(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, k, canManipulateRealOrders)

	var withdrawFiatRequest = withdraw.Request{
		Amount:        -1,
		Currency:      currency.EUR,
		Description:   "WITHDRAW IT ALL",
		TradePassword: "someBank",
	}

	_, err := k.WithdrawFiatFundsToInternationalBank(context.Background(),
		&withdrawFiatRequest)
	if !sharedtestvalues.AreAPICredentialsSet(k) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(k) && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

func TestGetCryptoDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	_, err := k.GetCryptoDepositAddress(context.Background(), "Bitcoin", "XBT", false)
	if err != nil {
		t.Error(err)
	}
	if !canManipulateRealOrders {
		t.Skip("canManipulateRealOrders not set, skipping test")
	}
	_, err = k.GetCryptoDepositAddress(context.Background(), "Bitcoin", "XBT", true)
	if err != nil {
		t.Error(err)
	}
}

// TestGetDepositAddress wrapper test
func TestGetDepositAddress(t *testing.T) {
	t.Parallel()
	if sharedtestvalues.AreAPICredentialsSet(k) {
		_, err := k.GetDepositAddress(context.Background(), currency.USDT, "", "")
		if err != nil {
			t.Error("GetDepositAddress() error", err)
		}
	} else {
		_, err := k.GetDepositAddress(context.Background(), currency.BTC, "", "")
		if err == nil {
			t.Error("GetDepositAddress() error can not be nil")
		}
	}
}

// TestWithdrawStatus wrapper test
func TestWithdrawStatus(t *testing.T) {
	t.Parallel()
	if sharedtestvalues.AreAPICredentialsSet(k) {
		_, err := k.WithdrawStatus(context.Background(), currency.BTC, "")
		if err != nil {
			t.Error("WithdrawStatus() error", err)
		}
	} else {
		_, err := k.WithdrawStatus(context.Background(), currency.BTC, "")
		if err == nil {
			t.Error("GetDepositAddress() error can not be nil")
		}
	}
}

// TestWithdrawCancel wrapper test
func TestWithdrawCancel(t *testing.T) {
	t.Parallel()
	_, err := k.WithdrawCancel(context.Background(), currency.BTC, "")
	if sharedtestvalues.AreAPICredentialsSet(k) && err == nil {
		t.Error("WithdrawCancel() error cannot be nil")
	} else if !sharedtestvalues.AreAPICredentialsSet(k) && err == nil {
		t.Errorf("WithdrawCancel() error - expecting an error when no keys are set but received nil")
	}
}

// ---------------------------- Websocket tests -----------------------------------------

// TestWsSubscribe tests unauthenticated websocket subscriptions
// Specifically looking to ensure multiple errors are collected and returned and ws.Subscriptions Added/Removed in cases of:
// single pass, single fail, mixed fail, multiple pass, all fail
// No objection to this becoming a fixture test, so long as it integrates through Un/Subscribe roundtrip
func TestWsSubscribe(t *testing.T) {
	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	testexch.SetupWs(t, k)

	err := k.Subscribe(subscription.List{{Channel: subscription.TickerChannel, Pairs: currency.Pairs{xbtusdPair}}})
	assert.NoError(t, err, "Simple subscription should not error")
	subs := k.Websocket.GetSubscriptions()
	require.Len(t, subs, 1, "Should add 1 Subscription")
	assert.Equal(t, subscription.SubscribedState, subs[0].State(), "Subscription should be subscribed state")

	err = k.Subscribe(subscription.List{{Channel: subscription.TickerChannel, Pairs: currency.Pairs{xbtusdPair}}})
	assert.ErrorIs(t, err, subscription.ErrDuplicate, "Resubscribing to the same channel should error with SubscribedAlready")
	subs = k.Websocket.GetSubscriptions()
	require.Len(t, subs, 1, "Should not add a subscription on error")
	assert.Equal(t, subscription.SubscribedState, subs[0].State(), "Existing subscription state should not change")

	err = k.Subscribe(subscription.List{{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "HOBBIT", "/")}}})
	assert.ErrorContains(t, err, "Currency pair not supported; Channel: ticker Pairs: DWARF/HOBBIT", "Subscribing to an invalid pair should error correctly")
	require.Len(t, k.Websocket.GetSubscriptions(), 1, "Should not add a subscription on error")

	// Mix success and failure
	err = k.Subscribe(subscription.List{
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("ETH", "USD", "/")}},
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "HOBBIT", "/")}},
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "ELF", "/")}},
	})
	assert.ErrorContains(t, err, "Currency pair not supported; Channel: ticker", "Subscribing to an invalid pair should error correctly")
	assert.ErrorContains(t, err, "DWARF/ELF", "Subscribing to an invalid pair should error correctly")
	assert.ErrorContains(t, err, "DWARF/HOBBIT", "Subscribing to an invalid pair should error correctly")
	require.Len(t, k.Websocket.GetSubscriptions(), 2, "Should have 2 subscriptions after mixed success/failures")

	// Just failures
	err = k.Subscribe(subscription.List{
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "HOBBIT", "/")}},
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "GOBLIN", "/")}},
	})
	assert.ErrorContains(t, err, "Currency pair not supported; Channel: ticker", "Subscribing to an invalid pair should error correctly")
	assert.ErrorContains(t, err, "DWARF/GOBLIN", "Subscribing to an invalid pair should error correctly")
	assert.ErrorContains(t, err, "DWARF/HOBBIT", "Subscribing to an invalid pair should error correctly")
	require.Len(t, k.Websocket.GetSubscriptions(), 2, "Should have 2 subscriptions after mixed success/failures")

	// Just success
	err = k.Subscribe(subscription.List{
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("ETH", "XBT", "/")}},
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("LTC", "ETH", "/")}},
	})
	assert.NoError(t, err, "Multiple successful subscriptions should not error")

	subs = k.Websocket.GetSubscriptions()
	assert.Len(t, subs, 4, "Should have correct number of subscriptions")

	err = k.Unsubscribe(subs[:1])
	assert.NoError(t, err, "Simple Unsubscribe should succeed")
	assert.Len(t, k.Websocket.GetSubscriptions(), 3, "Should have removed 1 channel")

	err = k.Unsubscribe(subscription.List{{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "WIZARD", "/")}, Key: 1337}})
	assert.ErrorIs(t, err, subscription.ErrNotFound, "Simple failing Unsubscribe should error NotFound")
	assert.ErrorContains(t, err, "DWARF/WIZARD", "Unsubscribing from an invalid pair should error correctly")
	assert.Len(t, k.Websocket.GetSubscriptions(), 3, "Should not have removed any channels")

	err = k.Unsubscribe(subscription.List{
		subs[1],
		{Channel: subscription.TickerChannel, Pairs: currency.Pairs{currency.NewPairWithDelimiter("DWARF", "EAGLE", "/")}, Key: 1338},
	})
	assert.ErrorIs(t, err, subscription.ErrNotFound, "Mixed failing Unsubscribe should error NotFound")
	assert.ErrorContains(t, err, "Channel: ticker Pairs: DWARF/EAGLE", "Unsubscribing from an invalid pair should error correctly")

	subs = k.Websocket.GetSubscriptions()
	assert.Len(t, subs, 2, "Should have removed only 1 more channel")

	err = k.Unsubscribe(subs)
	assert.NoError(t, err, "Unsubscribe multiple passing subscriptions should not error")
	assert.Empty(t, k.Websocket.GetSubscriptions(), "Should have successfully removed all channels")

	for _, c := range []string{"ohlc", "ohlc-5"} {
		err = k.Subscribe(subscription.List{{
			Channel: c,
			Pairs:   currency.Pairs{xbtusdPair},
		}})
		assert.ErrorIs(t, err, subscription.ErrPrivateChannelName, "Must error when trying to use a private channel name")
		assert.ErrorContains(t, err, c+" => subscription.CandlesChannel", "Must error when trying to use a private channel name")
	}
}

// TestWsOrderbookSub tests orderbook subscriptions for MaxDepth params
func TestWsOrderbookSub(t *testing.T) {
	t.Parallel()

	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	testexch.SetupWs(t, k)

	err := k.Subscribe(subscription.List{{
		Channel: subscription.OrderbookChannel,
		Pairs:   currency.Pairs{xbtusdPair},
		Levels:  25,
	}})
	require.NoError(t, err, "Simple subscription should not error")

	subs := k.Websocket.GetSubscriptions()
	require.Equal(t, 1, len(subs), "Should have 1 subscription channel")

	err = k.Unsubscribe(subs)
	assert.NoError(t, err, "Unsubscribe should not error")
	assert.Empty(t, k.Websocket.GetSubscriptions(), "Should have successfully removed all channels")

	err = k.Subscribe(subscription.List{{
		Channel: subscription.OrderbookChannel,
		Pairs:   currency.Pairs{xbtusdPair},
		Levels:  42,
	}})
	assert.ErrorContains(t, err, "Subscription depth not supported", "Bad subscription should error about depth")
}

// TestWsCandlesSub tests candles subscription for Timeframe params
func TestWsCandlesSub(t *testing.T) {
	t.Parallel()

	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	testexch.SetupWs(t, k)

	err := k.Subscribe(subscription.List{{
		Channel:  subscription.CandlesChannel,
		Pairs:    currency.Pairs{xbtusdPair},
		Interval: kline.OneHour,
	}})
	require.NoError(t, err, "Simple subscription should not error")

	subs := k.Websocket.GetSubscriptions()
	require.Equal(t, 1, len(subs), "Should add 1 Subscription")

	err = k.Unsubscribe(subs)
	assert.NoError(t, err, "Unsubscribe should not error")
	assert.Empty(t, k.Websocket.GetSubscriptions(), "Should have successfully removed all channels")

	err = k.Subscribe(subscription.List{{
		Channel:  subscription.CandlesChannel,
		Pairs:    currency.Pairs{xbtusdPair},
		Interval: kline.Interval(time.Minute * time.Duration(127)),
	}})
	assert.ErrorContains(t, err, "Subscription ohlc interval not supported", "Bad subscription should error about interval")
}

// TestWsOwnTradesSub tests the authenticated WS subscription channel for trades
func TestWsOwnTradesSub(t *testing.T) {
	t.Parallel()

	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)

	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	testexch.SetupWs(t, k)

	err := k.Subscribe(subscription.List{{Channel: subscription.MyTradesChannel, Authenticated: true}})
	assert.NoError(t, err, "Subsrcibing to ownTrades should not error")

	subs := k.Websocket.GetSubscriptions()
	assert.Len(t, subs, 1, "Should add 1 Subscription")

	err = k.Unsubscribe(subs)
	assert.NoError(t, err, "Unsubscribing an auth channel should not error")
	assert.Empty(t, k.Websocket.GetSubscriptions(), "Should have successfully removed channel")
}

// TestGenerateSubscriptions tests the subscriptions generated from configuration
func TestGenerateSubscriptions(t *testing.T) {
	t.Parallel()

	subs, err := k.generateSubscriptions()
	require.NoError(t, err, "generateSubscriptions should not error")
	pairs, err := k.GetEnabledPairs(asset.Spot)
	require.NoError(t, err, "GetEnabledPairs must not error")
	pairs = pairs.Format(currency.PairFormat{Uppercase: true, Delimiter: "/"})
	require.False(t, k.Websocket.CanUseAuthenticatedEndpoints(), "Websocket must not be authenticated by default")
	expected := subscription.List{}
	for _, exp := range k.Features.Subscriptions {
		if exp.Authenticated {
			continue
		}
		s := exp.Clone()
		s.Asset = asset.Spot
		s.Pairs = pairs
		expected = append(expected, s)
	}
	testsubs.Equal(t, expected, subs)
}

func TestGetWSToken(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k)
	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	testexch.SetupWs(t, k)

	resp, err := k.GetWebsocketToken(context.Background())
	require.NoError(t, err, "GetWebsocketToken must not error")
	assert.NotEmpty(t, resp, "Token should not be empty")
}

// TestWsAddOrder exercises roundtrip of wsAddOrder; See also: mockWsAddOrder
func TestWsAddOrder(t *testing.T) {
	t.Parallel()

	k := testexch.MockWsInstance[Kraken](t, curryWsMockUpgrader(t, mockWsServer)) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.True(t, k.IsWebsocketAuthenticationSupported(), "WS must be authenticated")
	id, err := k.wsAddOrder(&WsAddOrderRequest{
		OrderType: order.Limit.Lower(),
		OrderSide: order.Buy.Lower(),
		Pair:      "XBT/USD",
		Price:     80000,
	})
	require.NoError(t, err, "wsAddOrder must not error")
	assert.Equal(t, "ONPNXH-KMKMU-F4MR5V", id, "wsAddOrder should return correct order ID")
}

// TestWsCancelOrders exercises roundtrip of wsCancelOrders; See also: mockWsCancelOrders
func TestWsCancelOrders(t *testing.T) {
	t.Parallel()

	k := testexch.MockWsInstance[Kraken](t, curryWsMockUpgrader(t, mockWsServer)) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.True(t, k.IsWebsocketAuthenticationSupported(), "WS must be authenticated")

	err := k.wsCancelOrders([]string{"RABBIT", "BATFISH", "SQUIRREL", "CATFISH", "MOUSE"})
	assert.ErrorIs(t, err, errCancellingOrder, "Should error cancelling order")
	assert.ErrorContains(t, err, "BATFISH", "Should error containing txn id")
	assert.ErrorContains(t, err, "CATFISH", "Should error containing txn id")
	assert.ErrorContains(t, err, "[EOrder:Unknown order]", "Should error containing server error")

	err = k.wsCancelOrders([]string{"RABBIT", "SQUIRREL", "MOUSE"})
	assert.NoError(t, err, "Should not error with valid ids")
}

func TestWsCancelAllOrders(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, k, canManipulateRealOrders)
	testexch.SetupWs(t, k)
	_, err := k.wsCancelAllOrders()
	require.NoError(t, err, "wsCancelAllOrders must not error")
}

func TestWsHandleData(t *testing.T) {
	t.Parallel()
	base := k
	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	for _, l := range []int{10, 100} {
		err := k.Websocket.AddSuccessfulSubscriptions(&subscription.Subscription{
			Channel: subscription.OrderbookChannel,
			Pairs:   currency.Pairs{xbtusdPair},
			Asset:   asset.Spot,
			Levels:  l,
		})
		require.NoError(t, err, "AddSuccessfulSubscriptions must not error")
	}
	sharedtestvalues.TestFixtureToDataHandler(t, base, k, "testdata/wsHandleData.json", k.wsHandleData)
}

func TestWsOpenOrders(t *testing.T) {
	t.Parallel()
	n := new(Kraken)
	sharedtestvalues.TestFixtureToDataHandler(t, k, n, "testdata/wsOpenTrades.json", n.wsHandleData)
	seen := 0
	for reading := true; reading; {
		select {
		default:
			reading = false
		case resp := <-n.Websocket.DataHandler:
			seen++
			switch v := resp.(type) {
			case *order.Detail:
				switch seen {
				case 1:
					assert.Equal(t, "OGTT3Y-C6I3P-XRI6HR", v.OrderID, "OrderID")
					assert.Equal(t, order.Limit, v.Type, "order type")
					assert.Equal(t, order.Sell, v.Side, "order side")
					assert.Equal(t, order.Open, v.Status, "order status")
					assert.Equal(t, 34.5, v.Price, "price")
					assert.Equal(t, 10.00345345, v.Amount, "amount")
				case 2:
					assert.Equal(t, "OKB55A-UEMMN-YUXM2A", v.OrderID, "OrderID")
					assert.Equal(t, order.Market, v.Type, "order type")
					assert.Equal(t, order.Buy, v.Side, "order side")
					assert.Equal(t, order.Pending, v.Status, "order status")
					assert.Equal(t, 0.0, v.Price, "price")
					assert.Equal(t, 0.0001, v.Amount, "amount")
					assert.Equal(t, time.UnixMicro(1692851641361371).UTC(), v.Date, "Date")
				case 3:
					assert.Equal(t, "OKB55A-UEMMN-YUXM2A", v.OrderID, "OrderID")
					assert.Equal(t, order.Open, v.Status, "order status")
				case 4:
					assert.Equal(t, "OKB55A-UEMMN-YUXM2A", v.OrderID, "OrderID")
					assert.Equal(t, order.UnknownStatus, v.Status, "order status")
					assert.Equal(t, 26425.2, v.AverageExecutedPrice, "AverageExecutedPrice")
					assert.Equal(t, 0.0001, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 0.0, v.RemainingAmount, "RemainingAmount") // Not in the message; Testing regression to bad derivation
					assert.Equal(t, 0.00687, v.Fee, "Fee")
				case 5:
					assert.Equal(t, "OKB55A-UEMMN-YUXM2A", v.OrderID, "OrderID")
					assert.Equal(t, order.Closed, v.Status, "order status")
					assert.Equal(t, 0.0001, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 26425.2, v.AverageExecutedPrice, "AverageExecutedPrice")
					assert.Equal(t, 0.00687, v.Fee, "Fee")
					assert.Equal(t, time.UnixMicro(1692851641361447).UTC(), v.LastUpdated, "LastUpdated")
				case 6:
					assert.Equal(t, "OGTT3Y-C6I3P-XRI6HR", v.OrderID, "OrderID")
					assert.Equal(t, order.UnknownStatus, v.Status, "order status")
					assert.Equal(t, 10.00345345, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 0.001, v.Fee, "Fee")
					assert.Equal(t, 34.5, v.AverageExecutedPrice, "AverageExecutedPrice")
				case 7:
					assert.Equal(t, "OGTT3Y-C6I3P-XRI6HR", v.OrderID, "OrderID")
					assert.Equal(t, order.Closed, v.Status, "order status")
					assert.Equal(t, time.UnixMicro(1692675961789052).UTC(), v.LastUpdated, "LastUpdated")
					assert.Equal(t, 10.00345345, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 0.001, v.Fee, "Fee")
					assert.Equal(t, 34.5, v.AverageExecutedPrice, "AverageExecutedPrice")
					reading = false
				}
			default:
				t.Errorf("Unexpected type in DataHandler: %T (%s)", v, v)
			}
		}
	}

	assert.Equal(t, 7, seen, "number of DataHandler emissions")
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()

	testexch.UpdatePairsOnce(t, k)

	_, err := k.GetHistoricCandles(context.Background(), xbtusdPair, asset.Spot, kline.OneHour, time.Now().Add(-time.Hour*12), time.Now())
	assert.NoError(t, err, "GetHistoricCandles should not error")

	err = k.CurrencyPairs.EnablePair(asset.Futures, fxbtusdPair)
	require.NoError(t, err, "EnablePair must not error")

	_, err = k.GetHistoricCandles(context.Background(), fxbtusdPair, asset.Futures, kline.OneHour, time.Now().Add(-time.Hour*12), time.Now())
	assert.ErrorIs(t, err, asset.ErrNotSupported, "GetHistoricCandles should error correctly on Futures")
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	t.Parallel()
	_, err := k.GetHistoricCandlesExtended(context.Background(), fxbtusdPair, asset.Spot, kline.OneMin, time.Now().Add(-time.Minute*3), time.Now())
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported, "GetHistoricCandlesExtended should error correctly")
}

func Test_FormatExchangeKlineInterval(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		interval kline.Interval
		exp      string
	}{
		{kline.OneMin, "1"},
		{kline.OneDay, "1440"},
	} {
		assert.Equalf(t, tt.exp, k.FormatExchangeKlineInterval(tt.interval), "FormatExchangeKlineInterval should return correct output for %s", tt.interval.Short())
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()

	testexch.UpdatePairsOnce(t, k)

	_, err := k.GetRecentTrades(context.Background(), xbtusdPair, asset.Spot)
	assert.NoError(t, err, "GetRecentTrades should not error")

	_, err = k.GetRecentTrades(context.Background(), currency.NewPairWithDelimiter("PI", "BCHUSD", "_"), asset.Futures)
	assert.NoError(t, err, "GetRecentTrades should not error")
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("XBTUSD")
	if err != nil {
		t.Fatal(err)
	}
	_, err = k.GetHistoricTrades(context.Background(), currencyPair, asset.Spot, time.Now().Add(-time.Minute*15), time.Now())
	if err != nil && err != common.ErrFunctionNotSupported {
		t.Error(err)
	}
}

var testOb = orderbook.Base{
	Asks: []orderbook.Item{
		// NOTE: 0.00000500 float64 == 0.000005
		{Price: 0.05005, StrPrice: "0.05005", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05010, StrPrice: "0.05010", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05015, StrPrice: "0.05015", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05020, StrPrice: "0.05020", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05025, StrPrice: "0.05025", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05030, StrPrice: "0.05030", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05035, StrPrice: "0.05035", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05040, StrPrice: "0.05040", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05045, StrPrice: "0.05045", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.05050, StrPrice: "0.05050", Amount: 0.00000500, StrAmount: "0.00000500"},
	},
	Bids: []orderbook.Item{
		{Price: 0.05000, StrPrice: "0.05000", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04995, StrPrice: "0.04995", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04990, StrPrice: "0.04990", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04980, StrPrice: "0.04980", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04975, StrPrice: "0.04975", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04970, StrPrice: "0.04970", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04965, StrPrice: "0.04965", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04960, StrPrice: "0.04960", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04955, StrPrice: "0.04955", Amount: 0.00000500, StrAmount: "0.00000500"},
		{Price: 0.04950, StrPrice: "0.04950", Amount: 0.00000500, StrAmount: "0.00000500"},
	},
}

const krakenAPIDocChecksum = 974947235

func TestChecksumCalculation(t *testing.T) {
	t.Parallel()
	expected := "5005"
	if v := trim("0.05005"); v != expected {
		t.Errorf("expected %s but received %s", expected, v)
	}

	expected = "500"
	if v := trim("0.00000500"); v != expected {
		t.Errorf("expected %s but received %s", expected, v)
	}

	err := validateCRC32(&testOb, krakenAPIDocChecksum)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCharts(t *testing.T) {
	t.Parallel()
	resp, err := k.GetFuturesCharts(context.Background(), "1d", "spot", fxbtusdPair, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}

	end := time.UnixMilli(resp.Candles[0].Time)
	_, err = k.GetFuturesCharts(context.Background(), "1d", "spot", fxbtusdPair, end.Add(-time.Hour*24*7), end)
	if err != nil {
		t.Error(err)
	}
}

func TestGetFuturesTrades(t *testing.T) {
	t.Parallel()
	cp, err := currency.NewPairFromStrings("PI", "BCHUSD")
	if err != nil {
		t.Error(err)
	}
	cp.Delimiter = "_"
	_, err = k.GetFuturesTrades(context.Background(), cp, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}

	_, err = k.GetFuturesTrades(context.Background(), cp, time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Error(err)
	}
}

var websocketXDGUSDOrderbookUpdates = []string{
	`[2304,{"as":[["0.074602700","278.39626342","1690246067.832139"],["0.074611000","555.65134028","1690246086.243668"],["0.074613300","524.87121572","1690245901.574881"],["0.074624600","77.57180740","1690246060.668500"],["0.074632500","620.64648404","1690246010.904883"],["0.074698400","409.57419037","1690246041.269821"],["0.074700000","61067.71115772","1690246089.485595"],["0.074723200","4394.01869240","1690246087.557913"],["0.074725200","4229.57885125","1690246082.911452"],["0.074738400","212.25501214","1690246089.421559"]],"bs":[["0.074597400","53591.43163675","1690246089.451762"],["0.074596700","33594.18269213","1690246089.514152"],["0.074596600","53598.60351469","1690246089.340781"],["0.074594800","5358.57247081","1690246089.347962"],["0.074594200","30168.21074680","1690246089.345112"],["0.074590900","7089.69894583","1690246088.212880"],["0.074586700","46925.20182082","1690246089.074618"],["0.074577200","5500.00000000","1690246087.568856"],["0.074569600","8132.49888631","1690246086.841219"],["0.074562900","8413.11098009","1690246087.024863"]]},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074700000","0.00000000","1690246089.516119"],["0.074738500","125000.00000000","1690246063.352141","r"]],"c":"2219685759"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074678800","33476.70673703","1690246089.570183"]],"c":"1897176819"},"book-10","XDG/USD"]`,
	`[2304,{"b":[["0.074562900","0.00000000","1690246089.570206"],["0.074559600","4000.00000000","1690246086.478591","r"]],"c":"2498018751"},"book-10","XDG/USD"]`,
	`[2304,{"b":[["0.074577300","125000.00000000","1690246089.577140"]],"c":"155006629"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074678800","0.00000000","1690246089.584498"],["0.074738500","125000.00000000","1690246063.352141","r"]],"c":"3703147735"},"book-10","XDG/USD"]`,
	`[2304,{"b":[["0.074597500","10000.00000000","1690246089.602477"]],"c":"2989534775"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074738500","0.00000000","1690246089.608769"],["0.074750800","51369.02100000","1690246089.495500","r"]],"c":"1842075082"},"book-10","XDG/USD"]`,
	`[2304,{"b":[["0.074583500","8413.11098009","1690246089.612144"]],"c":"710274752"},"book-10","XDG/USD"]`,
	`[2304,{"b":[["0.074578500","9966.55841398","1690246089.634739"]],"c":"1646135532"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074738400","0.00000000","1690246089.638648"],["0.074751500","80499.09450000","1690246086.679402","r"]],"c":"2509689626"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074750700","290.96851266","1690246089.638754"]],"c":"3981738175"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074720000","61067.71115772","1690246089.662102"]],"c":"1591820326"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074602700","0.00000000","1690246089.670911"],["0.074750800","51369.02100000","1690246089.495500","r"]],"c":"3838272404"},"book-10","XDG/USD"]`,
	`[2304,{"a":[["0.074611000","0.00000000","1690246089.680343"],["0.074758500","159144.39750000","1690246035.158327","r"]],"c":"4241552383"},"book-10","XDG/USD"]	`,
}

var websocketLUNAEUROrderbookUpdates = []string{
	`[9536,{"as":[["0.000074650000","147354.32016076","1690249755.076929"],["0.000074710000","5084881.40000000","1690250711.359411"],["0.000074760000","9700502.70476704","1690250743.279490"],["0.000074990000","2933380.23886300","1690249596.627969"],["0.000075000000","433333.33333333","1690245575.626780"],["0.000075020000","152914.84493416","1690243661.232520"],["0.000075070000","146529.90542161","1690249048.358424"],["0.000075250000","737072.85720004","1690211553.549248"],["0.000075400000","670061.64567140","1690250769.261196"],["0.000075460000","980226.63603417","1690250769.627523"]],"bs":[["0.000074590000","71029.87806720","1690250763.012724"],["0.000074580000","15935576.86404000","1690250763.012710"],["0.000074520000","33758611.79634000","1690250718.290955"],["0.000074350000","3156650.58590277","1690250766.499648"],["0.000074340000","301727260.79999999","1690250766.490238"],["0.000074320000","64611496.53837000","1690250742.680258"],["0.000074310000","104228596.60000000","1690250744.679121"],["0.000074300000","40366046.10582000","1690250762.685914"],["0.000074200000","3690216.57320475","1690250645.311465"],["0.000074060000","1337170.52532521","1690250742.012527"]]},"book-10","LUNA/EUR"]`,
	`[9536,{"b":[["0.000074060000","0.00000000","1690250770.616604"],["0.000074050000","16742421.17790510","1690250710.867730","r"]],"c":"418307145"},"book-10","LUNA/EUR"]`,
}

var websocketGSTEUROrderbookUpdates = []string{
	`[8912,{"as":[["0.01300","850.00000000","1690230914.230506"],["0.01400","323483.99590510","1690256356.615823"],["0.01500","100287.34442717","1690219133.193345"],["0.01600","67995.78441017","1690118389.451216"],["0.01700","41776.38397740","1689676303.381189"],["0.01800","11785.76177777","1688631951.812452"],["0.01900","23700.00000000","1686935422.319042"],["0.02000","3941.17000000","1689415829.176481"],["0.02100","16598.69173066","1689420942.541943"],["0.02200","17572.51572836","1689851425.907427"]],"bs":[["0.01200","14220.66466572","1690256540.842831"],["0.01100","160223.61546438","1690256401.072463"],["0.01000","63083.48958963","1690256604.037673"],["0.00900","6750.00000000","1690252470.633938"],["0.00800","213059.49706376","1690256360.386301"],["0.00700","1000.00000000","1689869458.464975"],["0.00600","4000.00000000","1690221333.528698"],["0.00100","245000.00000000","1690051368.753455"]]},"book-10","GST/EUR"]`,
	`[8912,{"b":[["0.01000","60583.48958963","1690256620.206768"],["0.01000","63083.48958963","1690256620.206783"]],"c":"69619317"},"book-10","GST/EUR"]`,
}

func TestWsOrderbookMax10Depth(t *testing.T) {
	t.Parallel()
	k := new(Kraken) //nolint:govet // Intentional shadow to avoid future copy/paste mistakes
	require.NoError(t, testexch.TestInstance(k), "TestInstance must not error")
	pairs := currency.Pairs{
		currency.NewPairWithDelimiter("XDG", "USD", "/"),
		currency.NewPairWithDelimiter("LUNA", "EUR", "/"),
		currency.NewPairWithDelimiter("GST", "EUR", "/"),
	}
	for _, p := range pairs {
		err := k.Websocket.AddSuccessfulSubscriptions(&subscription.Subscription{
			Channel: subscription.OrderbookChannel,
			Pairs:   currency.Pairs{p},
			Asset:   asset.Spot,
			Levels:  10,
		})
		require.NoError(t, err, "AddSuccessfulSubscriptions must not error")
	}

	for x := range websocketXDGUSDOrderbookUpdates {
		err := k.wsHandleData([]byte(websocketXDGUSDOrderbookUpdates[x]))
		require.NoError(t, err, "wsHandleData should not error")
	}

	for x := range websocketLUNAEUROrderbookUpdates {
		err := k.wsHandleData([]byte(websocketLUNAEUROrderbookUpdates[x]))
		// TODO: Known issue with LUNA pairs and big number float precision
		// storage and checksum calc. Might need to store raw strings as fields
		// in the orderbook.Item struct.
		// Required checksum: 7465000014735432016076747100005084881400000007476000097005027047670474990000293338023886300750000004333333333333375020000152914844934167507000014652990542161752500007370728572000475400000670061645671407546000098022663603417745900007102987806720745800001593557686404000745200003375861179634000743500003156650585902777434000030172726079999999743200006461149653837000743100001042285966000000074300000403660461058200074200000369021657320475740500001674242117790510
		if x != len(websocketLUNAEUROrderbookUpdates)-1 {
			require.NoError(t, err, "wsHandleData should not error")
		}
	}

	// This has less than 10 bids and still needs a checksum calc.
	for x := range websocketGSTEUROrderbookUpdates {
		err := k.wsHandleData([]byte(websocketGSTEUROrderbookUpdates[x]))
		require.NoError(t, err, "wsHandleData should not error")
	}
}

func TestGetFuturesContractDetails(t *testing.T) {
	t.Parallel()
	_, err := k.GetFuturesContractDetails(context.Background(), asset.Spot)
	if !errors.Is(err, futures.ErrNotFuturesAsset) {
		t.Error(err)
	}
	_, err = k.GetFuturesContractDetails(context.Background(), asset.USDTMarginedFutures)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Error(err)
	}

	_, err = k.GetFuturesContractDetails(context.Background(), asset.Futures)
	assert.NoError(t, err, "GetFuturesContractDetails should not error")
}

func TestGetLatestFundingRates(t *testing.T) {
	t.Parallel()
	_, err := k.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset:                asset.USDTMarginedFutures,
		Pair:                 currency.NewPair(currency.BTC, currency.USD),
		IncludePredictedRate: true,
	})
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Error(err)
	}

	_, err = k.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset: asset.Futures,
	})
	if !errors.Is(err, nil) {
		t.Error(err)
	}

	cp := currency.NewPair(currency.PF, currency.NewCode("XBTUSD"))
	cp.Delimiter = "_"
	err = k.CurrencyPairs.EnablePair(asset.Futures, cp)
	if err != nil && !errors.Is(err, currency.ErrPairAlreadyEnabled) {
		t.Fatal(err)
	}
	_, err = k.GetLatestFundingRates(context.Background(), &fundingrate.LatestRateRequest{
		Asset:                asset.Futures,
		Pair:                 cp,
		IncludePredictedRate: true,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestIsPerpetualFutureCurrency(t *testing.T) {
	t.Parallel()
	is, err := k.IsPerpetualFutureCurrency(asset.Binary, currency.NewPair(currency.BTC, currency.USDT))
	if err != nil {
		t.Error(err)
	}
	if is {
		t.Error("expected false")
	}

	is, err = k.IsPerpetualFutureCurrency(asset.Futures, currency.NewPair(currency.BTC, currency.USDT))
	if err != nil {
		t.Error(err)
	}
	if is {
		t.Error("expected false")
	}

	is, err = k.IsPerpetualFutureCurrency(asset.Futures, currency.NewPair(currency.PF, currency.NewCode("XBTUSD")))
	if err != nil {
		t.Error(err)
	}
	if !is {
		t.Error("expected true")
	}
}

func TestGetOpenInterest(t *testing.T) {
	t.Parallel()
	_, err := k.GetOpenInterest(context.Background(), key.PairAsset{
		Base:  currency.ETH.Item,
		Quote: currency.USDT.Item,
		Asset: asset.USDTMarginedFutures,
	})
	assert.ErrorIs(t, err, asset.ErrNotSupported)

	cp1 := currency.NewPair(currency.PF, currency.NewCode("ETHUSD"))
	cp2 := currency.NewPair(currency.PF, currency.NewCode("XBTUSD"))
	sharedtestvalues.SetupCurrencyPairsForExchangeAsset(t, k, asset.Futures, cp1, cp2)

	resp, err := k.GetOpenInterest(context.Background(), key.PairAsset{
		Base:  cp1.Base.Item,
		Quote: cp1.Quote.Item,
		Asset: asset.Futures,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	resp, err = k.GetOpenInterest(context.Background(),
		key.PairAsset{
			Base:  cp1.Base.Item,
			Quote: cp1.Quote.Item,
			Asset: asset.Futures,
		},
		key.PairAsset{
			Base:  cp2.Base.Item,
			Quote: cp2.Quote.Item,
			Asset: asset.Futures,
		})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	resp, err = k.GetOpenInterest(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

// curryWsMockUpgrader handles Kraken specific http auth token responses prior to handling off to standard Websocket upgrader
func curryWsMockUpgrader(tb testing.TB, h testexch.WsMockFunc) http.HandlerFunc {
	tb.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "GetWebSocketsToken") {
			_, err := w.Write([]byte(`{"result":{"token":"mockAuth"}}`))
			require.NoError(tb, err, "Write should not error")
			return
		}
		testexch.WsMockUpgrader(tb, w, r, h)
	}
}
