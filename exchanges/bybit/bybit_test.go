package bybit

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// Please supply your own keys here to do authenticated endpoint testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

var b = &Bybit{}

var spotTradablePair, linearTradablePair, inverseTradablePair, optionsTradablePair currency.Pair

func TestMain(m *testing.M) {
	b.SetDefaults()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal(err)
	}

	exchCfg, err := cfg.GetExchangeConfig("Bybit")
	if err != nil {
		log.Fatal(err)
	}

	exchCfg.Enabled = true
	exchCfg.API.AuthenticatedSupport = true
	exchCfg.API.AuthenticatedWebsocketSupport = true
	exchCfg.API.Credentials.Key = apiKey
	exchCfg.API.Credentials.Secret = apiSecret
	b.Websocket = sharedtestvalues.NewTestWebsocket()
	request.MaxRequestJobs = 100
	err = b.Setup(exchCfg)
	if err != nil {
		log.Fatal(err)
	}

	// Turn on all pairs for testing
	supportedAssets := b.GetAssetTypes(false)
	for x := range supportedAssets {
		var avail currency.Pairs
		avail, err = b.GetAvailablePairs(supportedAssets[x])
		if err != nil {
			log.Fatal(err)
		}

		err = b.CurrencyPairs.StorePairs(supportedAssets[x], avail, true)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = instantiateTradablePairs()
	if err != nil {
		log.Fatalf("%s %v", b.Name, err)
	}
	os.Exit(m.Run())
}

func TestStart(t *testing.T) {
	t.Parallel()
	err := b.Start(context.Background(), nil)
	if !errors.Is(err, common.ErrNilPointer) {
		t.Fatalf("received: '%v' but expected: '%v'", err, common.ErrNilPointer)
	}
	var testWg sync.WaitGroup
	err = b.Start(context.Background(), &testWg)
	if err != nil {
		t.Fatal(err)
	}
	testWg.Wait()
}

// test cases for SPOT

func TestGetInstrumentInfo(t *testing.T) {
	t.Parallel()
	_, err := b.GetInstruments(context.Background(), "spot", "", "", "", "", 0)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetInstruments(context.Background(), "linear", "", "", "", "", 0)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetInstruments(context.Background(), "inverse", "", "", "", "", 0)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetInstruments(context.Background(), "option", "", "", "", "", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetKlines(t *testing.T) {
	t.Parallel()
	_, err := b.GetKlines(context.Background(), "spot", spotTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetKlines(context.Background(), "linear", linearTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetKlines(context.Background(), "inverse", inverseTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetKlines(context.Background(), "option", optionsTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err == nil {
		t.Fatalf("expected 'params error: Category is invalid', but found nil")
	}
}

func TestGetMarkPriceKline(t *testing.T) {
	t.Parallel()
	b.Verbose = true
	_, err := b.GetMarkPriceKline(context.Background(), "linear", linearTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetMarkPriceKline(context.Background(), "inverse", inverseTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetMarkPriceKline(context.Background(), "option", optionsTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err == nil {
		t.Fatalf("expected 'params error: Category is invalid', but found nil")
	}
}

func TestGetIndexPriceKline(t *testing.T) {
	t.Parallel()
	_, err := b.GetIndexPriceKline(context.Background(), "linear", linearTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetIndexPriceKline(context.Background(), "inverse", inverseTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetIndexPriceKline(context.Background(), "option", optionsTradablePair.String(), kline.FiveMin, time.Now().Add(-time.Hour*1), time.Now(), 5)
	if err == nil {
		t.Fatalf("expected 'params error: Category is invalid', but found nil")
	}
}

func TestGetOrderBook(t *testing.T) {
	t.Parallel()
	_, err := b.GetOrderBook(context.Background(), "spot", spotTradablePair.String(), 100)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetOrderBook(context.Background(), "linear", linearTradablePair.String(), 100)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetOrderBook(context.Background(), "inverse", inverseTradablePair.String(), 100)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetOrderBook(context.Background(), "option", optionsTradablePair.String(), 100)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRiskLimit(t *testing.T) {
	t.Parallel()
	_, err := b.GetRiskLimit(context.Background(), "linear", linearTradablePair.String())
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRiskLimit(context.Background(), "inverse", inverseTradablePair.String())
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRiskLimit(context.Background(), "option", optionsTradablePair.String())
	if !errors.Is(err, errInvalidCategory) {
		t.Error(err)
	}
	_, err = b.GetRiskLimit(context.Background(), "spot", spotTradablePair.String())
	if !errors.Is(err, errInvalidCategory) {
		t.Error(err)
	}
}

// test cases for Wrapper
func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	_, err := b.UpdateTicker(context.Background(), spotTradablePair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
	_, err = b.UpdateTicker(context.Background(), linearTradablePair, asset.Linear)
	if err != nil {
		t.Error(err)
	}
	_, err = b.UpdateTicker(context.Background(), inverseTradablePair, asset.Inverse)
	if err != nil {
		t.Error(err)
	}
	_, err = b.UpdateTicker(context.Background(), optionsTradablePair, asset.Options)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateOrderbook(t *testing.T) {
	t.Parallel()
	_, err := b.UpdateOrderbook(context.Background(), spotTradablePair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
	_, err = b.UpdateOrderbook(context.Background(), linearTradablePair, asset.Linear)
	if err != nil {
		t.Error(err)
	}

	_, err = b.UpdateOrderbook(context.Background(), inverseTradablePair, asset.Inverse)
	if err != nil {
		t.Error(err)
	}

	_, err = b.UpdateOrderbook(context.Background(), optionsTradablePair, asset.Options)
	if err != nil {
		t.Error(err)
	}
}

func TestSubmitOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	var orderSubmission = &order.Submit{
		Exchange:      b.GetName(),
		Pair:          spotTradablePair,
		Side:          order.Buy,
		Type:          order.Limit,
		Price:         1,
		Amount:        1,
		ClientOrderID: "1234",
		AssetType:     asset.Spot,
	}
	_, err := b.SubmitOrder(context.Background(), orderSubmission)
	if err != nil {
		t.Error(err)
	}
	_, err = b.SubmitOrder(context.Background(), &order.Submit{
		Exchange:      b.GetName(),
		AssetType:     asset.Options,
		Pair:          optionsTradablePair,
		Side:          order.Sell,
		Type:          order.Market,
		Price:         1,
		Amount:        1,
		Leverage:      1234,
		ClientOrderID: "1234",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestModifyOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ModifyOrder(context.Background(), &order.Modify{
		OrderID:      "1234",
		Type:         order.Limit,
		Side:         order.Buy,
		AssetType:    asset.Options,
		Pair:         spotTradablePair,
		PostOnly:     true,
		Price:        1234,
		Amount:       0.15,
		TriggerPrice: 1145,
		RiskManagementModes: order.RiskManagementModes{
			StopLoss: order.RiskManagement{
				Price: 0,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()
	end := time.Now()
	start := end.AddDate(0, 0, -3)
	_, err := b.GetHistoricCandles(context.Background(), spotTradablePair, asset.Spot, kline.OneDay, start, end)
	if err != nil {
		t.Error(err)
	}

	_, err = b.GetHistoricCandles(context.Background(), linearTradablePair, asset.Linear, kline.OneDay, start, end)
	if err != nil {
		t.Error(err)
	}

	_, err = b.GetHistoricCandles(context.Background(), inverseTradablePair, asset.Inverse, kline.OneHour, start, end)
	if err != nil {
		t.Error(err)
	}

	_, err = b.GetHistoricCandles(context.Background(), optionsTradablePair, asset.Options, kline.OneHour, start, end)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Errorf("expected %v, got %v", err, asset.ErrNotSupported)
	}
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	t.Parallel()
	startTime := time.Now().Add(-time.Hour * 24 * 3)
	end := time.Now().Add(-time.Hour * 1)
	_, err := b.GetHistoricCandlesExtended(context.Background(), spotTradablePair, asset.Spot, kline.OneMin, startTime, end)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricCandlesExtended(context.Background(), inverseTradablePair, asset.Inverse, kline.OneHour, startTime, end)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricCandlesExtended(context.Background(), linearTradablePair, asset.Linear, kline.OneDay, startTime, end)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricCandlesExtended(context.Background(), optionsTradablePair, asset.Options, kline.FiveMin, startTime, end)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Errorf("found %v, expected %v", err, asset.ErrNotSupported)
	}
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.CancelOrder(context.Background(), &order.Cancel{
		Exchange:  b.Name,
		AssetType: asset.Spot,
		Pair:      spotTradablePair,
		OrderID:   "1234"})
	if err != nil {
		t.Error(err)
	}
	err = b.CancelOrder(context.Background(), &order.Cancel{
		Exchange:  b.Name,
		AssetType: asset.Linear,
		Pair:      linearTradablePair,
		OrderID:   "1234"})
	if err != nil {
		t.Error(err)
	}

	err = b.CancelOrder(context.Background(), &order.Cancel{
		Exchange:  b.Name,
		AssetType: asset.Inverse,
		Pair:      inverseTradablePair,
		OrderID:   "1234"})
	if err != nil {
		t.Error(err)
	}
	err = b.CancelOrder(context.Background(), &order.Cancel{
		Exchange:  b.Name,
		AssetType: asset.Options,
		Pair:      optionsTradablePair,
		OrderID:   "1234"})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelAllOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CancelAllOrders(context.Background(), &order.Cancel{AssetType: asset.Spot})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CancelAllOrders(context.Background(), &order.Cancel{Exchange: b.Name, AssetType: asset.Linear, Pair: linearTradablePair})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CancelAllOrders(context.Background(), &order.Cancel{Exchange: b.Name, AssetType: asset.Inverse, Pair: inverseTradablePair})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CancelAllOrders(context.Background(), &order.Cancel{Exchange: b.Name, AssetType: asset.Options, Pair: optionsTradablePair})
	if err != nil {
		t.Error(err)
	}
	_, err = b.CancelAllOrders(context.Background(), &order.Cancel{Exchange: b.Name, AssetType: asset.Futures})
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Errorf("expected %v, but found %v", asset.ErrNotSupported, err)
	}
}

func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetOrderInfo(context.Background(),
		"12234", spotTradablePair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetOrderInfo(context.Background(),
		"12234", linearTradablePair, asset.Linear)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetOrderInfo(context.Background(),
		"12234", inverseTradablePair, asset.Inverse)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetOrderInfo(context.Background(),
		"12234", optionsTradablePair, asset.Options)
	if err != nil {
		t.Error(err)
	}
}

func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	var getOrdersRequestSpot = order.MultiOrderRequest{
		Pairs:     currency.Pairs{spotTradablePair},
		AssetType: asset.Spot,
		Side:      order.AnySide,
		Type:      order.AnyType,
	}
	_, err := b.GetActiveOrders(context.Background(), &getOrdersRequestSpot)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestLinear = order.MultiOrderRequest{Pairs: currency.Pairs{linearTradablePair}, AssetType: asset.Linear, Side: order.AnySide, Type: order.AnyType}
	_, err = b.GetActiveOrders(context.Background(), &getOrdersRequestLinear)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestInverse = order.MultiOrderRequest{Pairs: currency.Pairs{inverseTradablePair}, AssetType: asset.Inverse, Side: order.AnySide, Type: order.AnyType}
	_, err = b.GetActiveOrders(context.Background(), &getOrdersRequestInverse)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestFutures = order.MultiOrderRequest{Pairs: currency.Pairs{optionsTradablePair}, AssetType: asset.Options, Side: order.AnySide, Type: order.AnyType}
	_, err = b.GetActiveOrders(context.Background(), &getOrdersRequestFutures)
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	var getOrdersRequestSpot = order.MultiOrderRequest{
		Pairs:     currency.Pairs{spotTradablePair},
		AssetType: asset.Spot,
		Type:      order.AnyType,
		Side:      order.AnySide,
	}
	_, err := b.GetOrderHistory(context.Background(), &getOrdersRequestSpot)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestUMF = order.MultiOrderRequest{
		Pairs:     currency.Pairs{linearTradablePair},
		AssetType: asset.Linear,
		Type:      order.AnyType,
		Side:      order.AnySide,
	}
	_, err = b.GetOrderHistory(context.Background(), &getOrdersRequestUMF)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestCMF = order.MultiOrderRequest{
		Pairs:     currency.Pairs{inverseTradablePair},
		AssetType: asset.Inverse,
		Type:      order.AnyType,
		Side:      order.AnySide,
	}
	_, err = b.GetOrderHistory(context.Background(), &getOrdersRequestCMF)
	if err != nil {
		t.Error(err)
	}
	var getOrdersRequestFutures = order.MultiOrderRequest{
		Pairs:     currency.Pairs{optionsTradablePair},
		AssetType: asset.Options,
		Type:      order.AnyType,
		Side:      order.AnySide,
	}
	_, err = b.GetOrderHistory(context.Background(), &getOrdersRequestFutures)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetDepositAddress(context.Background(), currency.USDT, "", currency.ETH.String())
	if err != nil {
		t.Error(err)
	}
}

func TestGetAvailableTransferChains(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAvailableTransferChains(context.Background(), currency.USDT)
	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawCryptocurrencyFunds(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WithdrawCryptocurrencyFunds(context.Background(), &withdraw.Request{
		Exchange: "Bybit",
		Amount:   10,
		Currency: currency.LTC,
		Crypto: withdraw.CryptoRequest{
			Chain:      currency.LTC.String(),
			Address:    "3CDJNfdWX8m2NwuGUV3nhXHXEeLygMXoAj",
			AddressTag: "",
		}})
	if err != nil && err.Error() != "Withdraw address chain or destination tag are not equal" {
		t.Fatal(err)
	}
}

func TestUpdateTickers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	err := b.UpdateTickers(ctx, asset.Spot)
	if err != nil {
		t.Fatalf("%v %v\n", asset.Spot, err)
	}
	err = b.UpdateTickers(ctx, asset.Linear)
	if err != nil {
		t.Fatalf("%v %v\n", asset.Linear, err)
	}
	err = b.UpdateTickers(ctx, asset.Inverse)
	if err != nil {
		t.Fatalf("%v %v\n", asset.Inverse, err)
	}
	err = b.UpdateTickers(ctx, asset.Options)
	if err != nil {
		t.Fatalf("%v %v\n", asset.Options, err)
	}
}

func TestGetTickersV5(t *testing.T) {
	t.Parallel()
	_, err := b.GetTickers(context.Background(), "bruh", "", "", time.Time{})
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetTickers(context.Background(), "option", optionsTradablePair.String(), "", time.Time{})
	if !errors.Is(err, errBaseNotSet) {
		t.Fatalf("expected: %v, received: %v", errBaseNotSet, err)
	}
	_, err = b.GetTickers(context.Background(), "spot", "", "", time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetTickers(context.Background(), "option", "", "BTC", time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetTickers(context.Background(), "inverse", "", "", time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetTickers(context.Background(), "linear", "", "", time.Time{})
	if err != nil {
		t.Error(err)
	}
}

func TestGetFundingRateHistory(t *testing.T) {
	t.Parallel()
	_, err := b.GetFundingRateHistory(context.Background(), "bruh", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetFundingRateHistory(context.Background(), "spot", spotTradablePair.String(), time.Time{}, time.Time{}, 100)
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetFundingRateHistory(context.Background(), "linear", linearTradablePair.String(), time.Time{}, time.Time{}, 100)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetFundingRateHistory(context.Background(), "inverse", inverseTradablePair.String(), time.Time{}, time.Time{}, 100)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetFundingRateHistory(context.Background(), "option", optionsTradablePair.String(), time.Time{}, time.Time{}, 100)
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
}

func TestGetPublicTradingHistory(t *testing.T) {
	t.Parallel()
	_, err := b.GetPublicTradingHistory(context.Background(), "spot", spotTradablePair.String(), "", "", 30)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetPublicTradingHistory(context.Background(), "linear", linearTradablePair.String(), "", "", 30)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetPublicTradingHistory(context.Background(), "inverse", inverseTradablePair.String(), "", "", 30)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetPublicTradingHistory(context.Background(), "option", optionsTradablePair.String(), "BTC", "", 30)
	if err != nil {
		t.Error(err)
	}
}

func TestGetOpenInterest(t *testing.T) {
	t.Parallel()
	_, err := b.GetOpenInterest(context.Background(), "spot", spotTradablePair.String(), "5min", time.Time{}, time.Time{}, 0, "")
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetOpenInterest(context.Background(), "linear", linearTradablePair.String(), "5min", time.Time{}, time.Time{}, 0, "")
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetOpenInterest(context.Background(), "inverse", inverseTradablePair.String(), "5min", time.Time{}, time.Time{}, 0, "")
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetOpenInterest(context.Background(), "option", optionsTradablePair.String(), "5min", time.Time{}, time.Time{}, 0, "")
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
}

func TestGetHistoricalValatility(t *testing.T) {
	t.Parallel()
	_, err := b.GetHistoricalValatility(context.Background(), "option", "", 123, time.Now().Add(-time.Hour*30*24), time.Now())
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricalValatility(context.Background(), "spot", "", 123, time.Now().Add(-time.Hour*30*24), time.Now())
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, but found %v", errInvalidCategory, err)
	}
}

func TestGetInsurance(t *testing.T) {
	t.Parallel()
	_, err := b.GetInsurance(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetDeliveryPrice(t *testing.T) {
	t.Parallel()
	_, err := b.GetDeliveryPrice(context.Background(), "spot", spotTradablePair.String(), "", "", 200)
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, but found %v", errInvalidCategory, err)
	}
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()
	err := b.UpdateOrderExecutionLimits(context.Background(), asset.USDCMarginedFutures)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Fatalf("received: %v expected: %v", err, asset.ErrNotSupported)
	}
	err = b.UpdateOrderExecutionLimits(context.Background(), asset.Spot)
	if err != nil {
		t.Error("Bybit UpdateOrderExecutionLimits() error", err)
	}
	avail, err := b.GetAvailablePairs(asset.Spot)
	if err != nil {
		t.Fatal("Bybit GetAvailablePairs() error", err)
	}
	for x := range avail {
		limits, err := b.GetOrderExecutionLimits(asset.Spot, avail[x])
		if err != nil {
			t.Fatal("Bybit GetOrderExecutionLimits() error", err)
		}
		if limits == (order.MinMaxLevel{}) {
			t.Fatal("Bybit GetOrderExecutionLimits() error cannot be nil")
		}
	}
}

func TestPlaceOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	ctx := context.Background()
	_, err := b.PlaceOrder(ctx, nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category: "my-category",
	})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category: "spot",
	})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyPairEmpty, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category: "spot",
		Symbol:   currency.Pair{Delimiter: "", Base: currency.BTC, Quote: currency.USDT},
	})
	if !errors.Is(err, order.ErrSideIsInvalid) {
		t.Fatalf("expected %v, got %v", order.ErrSideIsInvalid, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category: "spot",
		Symbol:   spotTradablePair,
		Side:     "buy",
	})
	if !errors.Is(err, order.ErrTypeIsInvalid) {
		t.Fatalf("expected %v, got %v", order.ErrTypeIsInvalid, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category:  "spot",
		Symbol:    spotTradablePair,
		Side:      "buy",
		OrderType: "limit",
	})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Fatalf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.PlaceOrder(ctx, &PlaceOrderParams{
		Category:         "spot",
		Symbol:           spotTradablePair,
		Side:             "buy",
		OrderType:        "limit",
		OrderQuantity:    1,
		TriggerDirection: 3,
	})
	if !errors.Is(err, errInvalidTriggerDirection) {
		t.Fatalf("expected %v, got %v", errInvalidTriggerDirection, err)
	}
	_, err = b.PlaceOrder(context.Background(), &PlaceOrderParams{
		Category:         "spot",
		Symbol:           spotTradablePair,
		Side:             "buy",
		OrderType:        "limit",
		OrderQuantity:    1,
		Price:            31431.48,
		TriggerDirection: 2,
	})
	if err != nil {
		t.Error(err)
	}
	// Spot post only normal order
	arg := &PlaceOrderParams{Category: "spot", Symbol: spotTradablePair, Side: "Buy", OrderType: "Limit", OrderQuantity: 0.1, Price: 15600, TimeInForce: "PostOnly", OrderLinkID: "spot-test-01", IsLeverage: 0, OrderFilter: "Order"}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
	// Spot TP/SL order
	arg = &PlaceOrderParams{Category: "spot",
		Symbol: spotTradablePair,
		Side:   "Buy", OrderType: "Limit",
		OrderQuantity: 0.1, Price: 15600, TriggerPrice: 15000,
		TimeInForce: "GTC", OrderLinkID: "spot-test-02", IsLeverage: 0, OrderFilter: "tpslOrder"}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
	// Spot margin normal order (UTA)
	arg = &PlaceOrderParams{Category: "spot", Symbol: spotTradablePair, Side: "Buy", OrderType: "Limit",
		OrderQuantity: 0.1, Price: 15600, TimeInForce: "IOC", OrderLinkID: "spot-test-limit", IsLeverage: 1, OrderFilter: "Order"}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
	arg = &PlaceOrderParams{Category: "spot",
		Symbol: spotTradablePair,
		Side:   "Buy", OrderType: "Market", OrderQuantity: 200,
		TimeInForce: "IOC", OrderLinkID: "spot-test-04",
		IsLeverage: 0, OrderFilter: "Order"}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
	// USDT Perp open long position (one-way mode)
	arg = &PlaceOrderParams{Category: "linear",
		Symbol: linearTradablePair, Side: "Buy", OrderType: "Limit", OrderQuantity: 1, Price: 25000, TimeInForce: "GTC", PositionIdx: 0, OrderLinkID: "usdt-test-01", ReduceOnly: false, TakeProfitPrice: 28000, StopLossPrice: 20000, TpslMode: "Partial", TpOrderType: "Limit", SlOrderType: "Limit", TpLimitPrice: 27500, SlLimitPrice: 20500}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
	// USDT Perp close long position (one-way mode)
	arg = &PlaceOrderParams{Category: "linear", Symbol: linearTradablePair, Side: "Sell",
		OrderType: "Limit", OrderQuantity: 1, Price: 3000, TimeInForce: "GTC", PositionIdx: 0, OrderLinkID: "usdt-test-02", ReduceOnly: true}
	_, err = b.PlaceOrder(context.Background(), arg)
	if err != nil {
		t.Error(err)
	}
}

func TestAmendOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.AmendOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.AmendOrder(context.Background(), &AmendOrderParams{})
	if !errors.Is(err, errEitherOrderIDOROrderLinkIDRequired) {
		t.Fatalf("expected %v, got %v", errEitherOrderIDOROrderLinkIDRequired, err)
	}
	_, err = b.AmendOrder(context.Background(), &AmendOrderParams{
		OrderID: "c6f055d9-7f21-4079-913d-e6523a9cfffa",
	})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.AmendOrder(context.Background(), &AmendOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "mycat"})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.AmendOrder(context.Background(), &AmendOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "option"})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyPairEmpty, err)
	}
	_, err = b.AmendOrder(context.Background(), &AmendOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "spot", Symbol: spotTradablePair,
		TriggerPrice:    1145,
		OrderQuantity:   0.15,
		Price:           1050,
		TakeProfitPrice: 0,
		StopLossPrice:   0})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelTradeOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CancelTradeOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CancelTradeOrder(context.Background(), &CancelOrderParams{})
	if !errors.Is(err, errEitherOrderIDOROrderLinkIDRequired) {
		t.Fatalf("expected %v, got %v", errEitherOrderIDOROrderLinkIDRequired, err)
	}
	_, err = b.CancelTradeOrder(context.Background(), &CancelOrderParams{
		OrderID: "c6f055d9-7f21-4079-913d-e6523a9cfffa",
	})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.CancelTradeOrder(context.Background(), &CancelOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "mycat"})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.CancelTradeOrder(context.Background(), &CancelOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "option"})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyPairEmpty, err)
	}
	_, err = b.CancelTradeOrder(context.Background(), &CancelOrderParams{
		OrderID:  "c6f055d9-7f21-4079-913d-e6523a9cfffa",
		Category: "option",
		Symbol:   optionsTradablePair,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetOpenOrders(context.Background(), "", "", "", "", "", "", "", "", 0, 100)
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.GetOpenOrders(context.Background(), "spot", "", "", "", "", "", "", "", 0, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestCancelAllTradeOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CancelAllTradeOrders(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CancelAllTradeOrders(context.Background(), &CancelAllOrdersParam{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.CancelAllTradeOrders(context.Background(), &CancelAllOrdersParam{Category: "option"})
	if err != nil {
		t.Error(err)
	}
}

func TestGetTradeOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetTradeOrderHistory(context.Background(), "", "", "", "", "", "", "", "", "", time.Now().Add(-time.Hour*24*20), time.Now(), 100)
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.GetTradeOrderHistory(context.Background(), "spot", spotTradablePair.String(), "", "", "BTC", "", "StopOrder", "", "", time.Now().Add(-time.Hour*24*20), time.Now(), 100)
	if err != nil {
		t.Error(err)
	}
}

func TestPlaceBatchOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.PlaceBatchOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.PlaceBatchOrder(context.Background(), &PlaceBatchOrderParam{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.PlaceBatchOrder(context.Background(), &PlaceBatchOrderParam{
		Category: "spot",
	})
	if !errors.Is(err, errNoOrderPassed) {
		t.Fatalf("expected %v, got %v", errNoOrderPassed, err)
	}
	_, err = b.PlaceBatchOrder(context.Background(), &PlaceBatchOrderParam{
		Category: "option",
		Request: []BatchOrderItemParam{
			{
				Symbol:        optionsTradablePair,
				OrderType:     "Limit",
				Side:          "Buy",
				OrderQuantity: 1,
				OrderIv:       6,
				TimeInForce:   "GTC",
				OrderLinkID:   "option-test-001",
				Mmp:           false,
				ReduceOnly:    false,
			},
			{
				Symbol:        optionsTradablePair,
				OrderType:     "Limit",
				Side:          "Sell",
				OrderQuantity: 2,
				Price:         700,
				TimeInForce:   "GTC",
				OrderLinkID:   "option-test-001",
				Mmp:           false,
				ReduceOnly:    false,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBatchAmendOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.BatchAmendOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.BatchAmendOrder(context.Background(), &BatchAmendOrderParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.BatchAmendOrder(context.Background(), &BatchAmendOrderParams{Category: "spot"})
	if !errors.Is(err, errNoOrderPassed) {
		t.Fatalf("expected %v, got %v", errNoOrderPassed, err)
	}
	_, err = b.BatchAmendOrder(context.Background(), &BatchAmendOrderParams{
		Category: "option",
		Request: []BatchAmendOrderParamItem{
			{
				Symbol:                 optionsTradablePair,
				OrderImpliedVolatility: "6.8",
				OrderID:                "b551f227-7059-4fb5-a6a6-699c04dbd2f2",
			},
			{
				Symbol:  optionsTradablePair,
				Price:   650,
				OrderID: "fa6a595f-1a57-483f-b9d3-30e9c8235a52",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelBatchOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CancelBatchOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CancelBatchOrder(context.Background(), &CancelBatchOrder{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.CancelBatchOrder(context.Background(), &CancelBatchOrder{Category: "spot"})
	if !errors.Is(err, errNoOrderPassed) {
		t.Fatalf("expected %v, got %v", errNoOrderPassed, err)
	}
	_, err = b.CancelBatchOrder(context.Background(), &CancelBatchOrder{
		Category: "option",
		Request: []CancelOrderParams{
			{
				Symbol:  optionsTradablePair,
				OrderID: "b551f227-7059-4fb5-a6a6-699c04dbd2f2",
			},
			{
				Symbol:  optionsTradablePair,
				OrderID: "fa6a595f-1a57-483f-b9d3-30e9c8235a52",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetBorrowQuota(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetBorrowQuota(context.Background(), "", "BTCUSDT", "Buy")
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.GetBorrowQuota(context.Background(), "spot", "", "Buy")
	if !errors.Is(err, errSymbolMissing) {
		t.Fatalf("expected %v, got %v", errSymbolMissing, err)
	}
	_, err = b.GetBorrowQuota(context.Background(), "spot", spotTradablePair.String(), "")
	if !errors.Is(err, order.ErrSideIsInvalid) {
		t.Error(err)
	}
	_, err = b.GetBorrowQuota(context.Background(), "spot", spotTradablePair.String(), "Buy")
	if err != nil {
		t.Error(err)
	}
}

func TestSetDisconnectCancelAll(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetDisconnectCancelAll(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	err = b.SetDisconnectCancelAll(context.Background(), &SetDCPParams{TimeWindow: 300})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetPositionInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPositionInfo(context.Background(), "", "", "", "", "", 20)
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.GetPositionInfo(context.Background(), "spot", "", "", "", "", 20)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetPositionInfo(context.Background(), "linear", linearTradablePair.String(), "", "", "", 20)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetPositionInfo(context.Background(), "option", optionsTradablePair.String(), "BTC", "", "", 20)
	if err != nil {
		t.Error(err)
	}
}
func TestSetLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetLeverage(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	err = b.SetLeverage(context.Background(), &SetLeverageParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	err = b.SetLeverage(context.Background(), &SetLeverageParams{Category: "spot"})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	err = b.SetLeverage(context.Background(), &SetLeverageParams{Category: "linear"})
	if !errors.Is(err, errSymbolMissing) {
		t.Fatalf("expected %v, got %v", errSymbolMissing, err)
	}
	err = b.SetLeverage(context.Background(), &SetLeverageParams{Category: "linear", Symbol: "BTCUSDT"})
	if !errors.Is(err, errInvalidLeverage) {
		t.Fatalf("expected %v, got %v", errInvalidLeverage, err)
	}
	err = b.SetLeverage(context.Background(), &SetLeverageParams{Category: "linear", Symbol: "BTCUSDT", SellLeverage: 3, BuyLeverage: 3})
	if err != nil {
		t.Error(err)
	}
}

func TestSwitchTradeMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SwitchTradeMode(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{Category: "spot"})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{Category: "linear"})
	if !errors.Is(err, errSymbolMissing) {
		t.Fatalf("expected %v, got %v", errSymbolMissing, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{Category: "linear", Symbol: linearTradablePair.String()})
	if !errors.Is(err, errInvalidLeverage) {
		t.Fatalf("expected %v, got %v", errInvalidLeverage, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{Category: "linear", Symbol: linearTradablePair.String(), SellLeverage: 3, BuyLeverage: 3, TradeMode: 2})
	if !errors.Is(err, errInvalidTradeModeValue) {
		t.Fatalf("expected %v, got %v", errInvalidTradeModeValue, err)
	}
	err = b.SwitchTradeMode(context.Background(), &SwitchTradeModeParams{Category: "linear", Symbol: linearTradablePair.String(), SellLeverage: 3, BuyLeverage: 3, TradeMode: 1})
	if err != nil {
		t.Error(err)
	}
}

func TestSetTakeProfitStopLossMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.SetTakeProfitStopLossMode(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{
		Category: "spot",
	})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{Category: "spot"})
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{Category: "linear"})
	if !errors.Is(err, errSymbolMissing) {
		t.Fatalf("expected %v, got %v", errSymbolMissing, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{Category: "linear", Symbol: "BTCUSDT"})
	if !errors.Is(err, errTakeProfitOrStopLossModeMissing) {
		t.Fatalf("expected %v, got %v", errTakeProfitOrStopLossModeMissing, err)
	}
	_, err = b.SetTakeProfitStopLossMode(context.Background(), &TPSLModeParams{Category: "linear", Symbol: "BTCUSDT", TpslMode: "Partial"})
	if err != nil {
		t.Error(err)
	}
}

func TestSwitchPositionMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SwitchPositionMode(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	err = b.SwitchPositionMode(context.Background(), &SwitchPositionModeParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	err = b.SwitchPositionMode(context.Background(), &SwitchPositionModeParams{Category: "linear"})
	if !errors.Is(err, errEitherSymbolOrCoinRequired) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	err = b.SwitchPositionMode(context.Background(), &SwitchPositionModeParams{Category: "linear", Symbol: linearTradablePair, PositionMode: 3})
	if err != nil {
		t.Error(err)
	}
}

func TestSetRiskLimit(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.SetRiskLimit(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.SetRiskLimit(context.Background(), &SetRiskLimitParam{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Errorf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.SetRiskLimit(context.Background(), &SetRiskLimitParam{Category: "linear", PositionMode: -2})
	if !errors.Is(err, errInvalidPositionMode) {
		t.Errorf("expected %v, got %v", errInvalidPositionMode, err)
	}
	_, err = b.SetRiskLimit(context.Background(), &SetRiskLimitParam{Category: "linear"})
	if !errors.Is(err, errSymbolMissing) {
		t.Errorf("expected %v, got %v", errSymbolMissing, err)
	}
	_, err = b.SetRiskLimit(context.Background(), &SetRiskLimitParam{
		Category:     "linear",
		RiskID:       1234,
		Symbol:       linearTradablePair,
		PositionMode: 0,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSetTradingStop(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetTradingStop(context.Background(), &TradingStopParams{})
	if !errors.Is(err, errCategoryNotSet) {
		t.Errorf("expected %v, got %v", errCategoryNotSet, err)
	}
	err = b.SetTradingStop(context.Background(), &TradingStopParams{Category: "spot"})
	if !errors.Is(err, errInvalidCategory) {
		t.Errorf("expected %v, got %v", errInvalidCategory, err)
	}
	err = b.SetTradingStop(context.Background(), &TradingStopParams{
		Category:                 "linear",
		Symbol:                   linearTradablePair,
		TakeProfit:               "0.5",
		StopLoss:                 "0.2",
		TakeProfitTriggerType:    "MarkPrice",
		StopLossTriggerType:      "IndexPrice",
		TakeProfitOrStopLossMode: "Partial",
		TakeProfitOrderType:      "Limit",
		StopLossOrderType:        "Limit",
		TakeProfitSize:           50,
		StopLossSize:             50,
		TakeProfitLimitPrice:     0.49,
		StopLossLimitPrice:       0.21,
		PositionIndex:            0,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSetAutoAddMargin(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetAutoAddMargin(context.Background(), &AddRemoveMarginParams{
		Category:      "inverse",
		Symbol:        inverseTradablePair,
		AutoAddmargin: 0,
		PositionMode:  2,
	})
	if err != nil {
		t.Error(err)
	}
}
func TestAddOrReduceMargin(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.AddOrReduceMargin(context.Background(), &AddRemoveMarginParams{
		Category:      "inverse",
		Symbol:        inverseTradablePair,
		AutoAddmargin: 0,
		PositionMode:  2,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetExecution(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetExecution(context.Background(), "spot", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetClosedPnL(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetClosedPnL(context.Background(), "spot", "", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", err, errInvalidCategory)
	}
	_, err = b.GetClosedPnL(context.Background(), "linear", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetPreUpgradeOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeOrderHistory(context.Background(), "", "", "", "", "", "", "", "", time.Time{}, time.Time{}, 100)
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("expected %v, got %v", errCategoryNotSet, err)
	}
	_, err = b.GetPreUpgradeOrderHistory(context.Background(), "option", "", "", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errBaseNotSet) {
		t.Fatalf("expected %v, got %v", errBaseNotSet, err)
	}
	_, err = b.GetPreUpgradeOrderHistory(context.Background(), "linear", "", "", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPreUpgradeTradeHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeTradeHistory(context.Background(), "", "", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errCategoryNotSet) {
		t.Fatalf("found %v, expected %v", err, errCategoryNotSet)
	}
	_, err = b.GetPreUpgradeTradeHistory(context.Background(), "option", "", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("found %v, expected %v", err, errInvalidCategory)
	}
	_, err = b.GetPreUpgradeTradeHistory(context.Background(), "linear", "", "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPreUpgradeClosedPnL(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeClosedPnL(context.Background(), "option", "BTCUSDT", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetPreUpgradeClosedPnL(context.Background(), "linear", "BTCUSDT", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPreUpgradeTransactionLog(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeTransactionLog(context.Background(), "option", "", "", "", time.Time{}, time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("found %v, expected %v", err, errInvalidCategory)
	}
	_, err = b.GetPreUpgradeTransactionLog(context.Background(), "linear", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPreUpgradeOptionDeliveryRecord(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeOptionDeliveryRecord(context.Background(), "linear", "", "", time.Time{}, 0)
	if !errors.Is(err, errInvalidCategory) {
		t.Error(err)
	}
	_, err = b.GetPreUpgradeOptionDeliveryRecord(context.Background(), "option", "", "", time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPreUpgradeUSDCSessionSettlement(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPreUpgradeUSDCSessionSettlement(context.Background(), "option", "", "", 10)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetPreUpgradeUSDCSessionSettlement(context.Background(), "linear", "", "", 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetWalletBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetWalletBalance(context.Background(), "UNIFIED", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeToUnifiedAccount(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.UpgradeToUnifiedAccount(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetBorrowHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetBorrowHistory(context.Background(), "BTC", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCollateralInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetCollateralInfo(context.Background(), "BTC")
	if err != nil {
		t.Error(err)
	}
}

func TestGetCoinGreeks(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetCoinGreeks(context.Background(), "BTC")
	if err != nil {
		t.Error(err)
	}
}

func TestGetFeeRate(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetFeeRate(context.Background(), "something", "", "BTC")
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetFeeRate(context.Background(), "linear", "", "BTC")
	if err != nil {
		t.Error(err)
	}
}

func TestGetAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAccountInfo(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetTransactionLog(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetTransactionLog(context.Background(), "option", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetTransactionLog(context.Background(), "linear", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestSetMarginMode(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.SetMarginMode(context.Background(), "PORTFOLIO_MARGIN")
	if err != nil {
		t.Error(err)
	}
}

func TestSetMMP(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetMMP(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("found %v, expected %v", err, errNilArgument)
	}
	b.Verbose = true
	err = b.SetMMP(context.Background(), &MMPRequestParam{
		BaseCoin:           "ETH",
		TimeWindowMS:       5000,
		FrozenPeriod:       100000,
		TradeQuantityLimit: 50,
		DeltaLimit:         20,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestResetMMP(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.ResetMMP(context.Background(), "USDT")
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("found %v, expected %v", err, errNilArgument)
	}
	err = b.ResetMMP(context.Background(), "BTC")
	if err != nil {
		t.Error(err)
	}
}

func TestGetMMPState(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	b.Verbose = true
	_, err := b.GetMMPState(context.Background(), "BTC")
	if err != nil {
		t.Error(err)
	}
}

func TestGetCoinExchangeRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetCoinExchangeRecords(context.Background(), "", "", "", 20)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetDeliveryRecord(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetDeliveryRecord(context.Background(), "spot", "", "", time.Now().Add(time.Hour*40), 20)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatal(err)
	}
	_, err = b.GetDeliveryRecord(context.Background(), "linear", "", "", time.Now().Add(time.Hour*40), 20)
	if err != nil {
		t.Error(err)
	}
}
func TestGetUSDCSessionSettlement(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetUSDCSessionSettlement(context.Background(), "option", "", "", 10)
	if !errors.Is(err, errInvalidCategory) {
		t.Fatalf("expected %v, got %v", errInvalidCategory, err)
	}
	_, err = b.GetUSDCSessionSettlement(context.Background(), "linear", "", "", 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAssetInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAssetInfo(context.Background(), "", "BTC")
	if !errors.Is(err, errMissingAccountType) {
		t.Fatal(err)
	}
	_, err = b.GetAssetInfo(context.Background(), "SPOT", "BTC")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAllCoinBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAllCoinBalance(context.Background(), "", "", "", 0)
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.GetAllCoinBalance(context.Background(), "FUND", "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSingleCoinBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetSingleCoinBalance(context.Background(), "", "", "", 0, 0)
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.GetSingleCoinBalance(context.Background(), "SPOT", currency.BTC.String(), "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTransferableCoin(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetTransferableCoin(context.Background(), "SPOT", "OPTION")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateInternalTransfer(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CreateInternalTransfer(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{})
	if !errors.Is(err, errMissingTransferID) {
		t.Fatalf("expected %v, got %v", errMissingTransferID, err)
	}
	transferID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{TransferID: transferID})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{
		TransferID: transferID,
		Coin:       currency.BTC,
	})
	if !errors.Is(err, order.ErrAmountIsInvalid) {
		t.Fatalf("expected %v, got %v", order.ErrAmountIsInvalid, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{
		TransferID: transferID,
		Coin:       currency.BTC,
		Amount:     123.456,
	})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456, FromAccountType: "UNIFIED"})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateInternalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456,
		ToAccountType:   "CONTRACT",
		FromAccountType: "UNIFIED"})
	if err != nil {
		t.Error(err)
	}
}

func TestGetInternalTransferRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	transferID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetInternalTransferRecords(context.Background(), transferID.String(), currency.BTC.String(), "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetSubUID(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetSubUID(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestEnableUniversalTransferForSubUID(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	err := b.EnableUniversalTransferForSubUID(context.Background())
	if !errors.Is(err, errMembersIDsNotSet) {
		t.Fatalf("expected %v, got %v", errMembersIDsNotSet, err)
	}
	transferID1, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	transferID2, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	err = b.EnableUniversalTransferForSubUID(context.Background(), transferID1.String(), transferID2.String())
	if err != nil {
		t.Error(err)
	}
}

func TestCreateUniversalTransfer(t *testing.T) {
	t.Parallel()
	_, err := b.CreateUniversalTransfer(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{})
	if !errors.Is(err, errMissingTransferID) {
		t.Fatalf("expected %v, got %v", errMissingTransferID, err)
	}
	transferID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{TransferID: transferID})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{
		TransferID: transferID,
		Coin:       currency.BTC,
	})
	if !errors.Is(err, order.ErrAmountIsInvalid) {
		t.Fatalf("expected %v, got %v", order.ErrAmountIsInvalid, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{
		TransferID: transferID,
		Coin:       currency.BTC,
		Amount:     123.456,
	})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456, FromAccountType: "UNIFIED"})
	if !errors.Is(err, errMissingAccountType) {
		t.Fatalf("expected %v, got %v", errMissingAccountType, err)
	}
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{TransferID: transferID,
		Coin: currency.BTC, Amount: 123.456,
		ToAccountType:   "CONTRACT",
		FromAccountType: "UNIFIED"})
	if !errors.Is(err, errMemberIDRequired) {
		t.Fatalf("expected %v, got %v", errMemberIDRequired, err)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CreateUniversalTransfer(context.Background(), &TransferParams{
		TransferID: transferID,
		Coin:       currency.BTC, Amount: 123.456,
		ToAccountType:   "CONTRACT",
		FromAccountType: "UNIFIED",
		FromMemberID:    123,
		ToMemberID:      456,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetUniversalTransferRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	transferID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetUniversalTransferRecords(context.Background(), transferID.String(), currency.BTC.String(), "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAllowedDepositCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAllowedDepositCoinInfo(context.Background(), "BTC", "", "", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestSetDepositAccount(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.SetDepositAccount(context.Background(), "FUND")
	if err != nil {
		t.Error(err)
	}
}

func TestGetDepositRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetDepositRecords(context.Background(), "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetSubDepositRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetSubDepositRecords(context.Background(), "12345", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestInternalDepositRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetInternalDepositRecordsOffChain(context.Background(), currency.ETH.String(), "", time.Time{}, time.Time{}, 8)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMasterDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetMasterDepositAddress(context.Background(), currency.LTC, "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetSubDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetSubDepositAddress(context.Background(), currency.LTC, "LTC", "12345")
	if err != nil {
		t.Error(err)
	}
}

func TestGetCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetCoinInfo(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetWithdrawalRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetWithdrawalRecords(context.Background(), currency.LTC, "", "", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetWithdrawableAmount(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetWithdrawableAmount(context.Background(), currency.LTC)
	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawCurrency(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.WithdrawCurrency(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.WithdrawCurrency(context.Background(), &WithdrawalParam{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.WithdrawCurrency(context.Background(), &WithdrawalParam{Coin: currency.BTC})
	if !errors.Is(err, errMissingChainInformation) {
		t.Fatalf("expected %v, got %v", errMissingChainInformation, err)
	}
	_, err = b.WithdrawCurrency(context.Background(), &WithdrawalParam{Coin: currency.LTC, Chain: "LTC"})
	if !errors.Is(err, errMissingAddressInfo) {
		t.Fatalf("expected %v, got %v", errMissingAddressInfo, err)
	}
	_, err = b.WithdrawCurrency(context.Background(), &WithdrawalParam{Coin: currency.LTC, Chain: "LTC", Address: "234234234"})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Fatalf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.WithdrawCurrency(context.Background(), &WithdrawalParam{Coin: currency.LTC, Chain: "LTC", Address: "234234234", Amount: 123})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelWithdrawal(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.CancelWithdrawal(context.Background(), "")
	if !errors.Is(err, errMissingWithdrawalID) {
		t.Fatalf("expected %v, got %v", errMissingWithdrawalID, err)
	}
	_, err = b.CancelWithdrawal(context.Background(), "12314")
	if err != nil {
		t.Error(err)
	}
}

func TestCreateNewSubUserID(t *testing.T) {
	t.Parallel()
	_, err := b.CreateNewSubUserID(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CreateNewSubUserID(context.Background(), &CreateSubUserParams{MemberType: 1, Switch: 1, Note: "test"})
	if !errors.Is(err, errMissingusername) {
		t.Fatalf("expected %v, got %v", errMissingusername, err)
	}
	_, err = b.CreateNewSubUserID(context.Background(), &CreateSubUserParams{Username: "Sami", Switch: 1, Note: "test"})
	if !errors.Is(err, errInvalidMemberType) {
		t.Fatalf("expected %v, got %v", errInvalidMemberType, err)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CreateNewSubUserID(context.Background(), &CreateSubUserParams{Username: "sami", MemberType: 1, Switch: 1, Note: "test"})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateSubUIDAPIKey(t *testing.T) {
	t.Parallel()
	_, err := b.CreateSubUIDAPIKey(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.CreateSubUIDAPIKey(context.Background(), &SubUIDAPIKeyParam{})
	if !errors.Is(err, errMissingUserID) {
		t.Fatalf("expected %v, got %v", errMissingUserID, err)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err = b.CreateSubUIDAPIKey(context.Background(), &SubUIDAPIKeyParam{
		Subuid:      53888000,
		Note:        "testxxx",
		ReadOnly:    0,
		Permissions: map[string][]string{"Wallet": {"AccountTransfer"}},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetSubUIDList(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetSubUIDList(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestFreezeSubUID(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.FreezeSubUID(context.Background(), "1234", true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAPIKeyInformation(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAPIKeyInformation(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetUIDWalletType(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetUIDWalletType(context.Background(), "234234")
	if err != nil {
		t.Error(err)
	}
}

func TestModifyMasterAPIKey(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ModifyMasterAPIKey(context.Background(), &SubUIDAPIKeyUpdateParam{})
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.ModifyMasterAPIKey(context.Background(), &SubUIDAPIKeyUpdateParam{
		ReadOnly: 0,
		IPs:      []string{"*"},
		Permissions: map[string][]string{
			"ContractTrade": {"Order", "Position"},
			"Spot":          {"SpotTrade"},
			"Wallet":        {"AccountTransfer", "SubMemberTransfer"},
			"Options":       {"OptionsTrade"},
			"Derivatives":   {"DerivativesTrade"},
			"CopyTrading":   {"CopyTrading"},
			"BlockTrade":    {},
			"Exchange":      {"ExchangeHistory"},
			"NFT":           {"NFTQueryProductList"}},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestModifySubAPIKey(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ModifySubAPIKey(context.Background(), &SubUIDAPIKeyUpdateParam{})
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.ModifySubAPIKey(context.Background(), &SubUIDAPIKeyUpdateParam{
		ReadOnly: 0,
		IPs:      []string{"*"},
		Permissions: map[string][]string{
			"ContractTrade": {},
			"Spot":          {"SpotTrade"},
			"Wallet":        {"AccountTransfer"},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteMasterAPIKey(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.DeleteMasterAPIKey(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteSubAPIKey(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.DeleteSubAccountAPIKey(context.Background(), "12434")
	if err != nil {
		t.Error(err)
	}
}

func TestGetAffiliateUserInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetAffiliateUserInfo(context.Background(), "1234")
	if err != nil {
		t.Error(err)
	}
}

func TestGetLeverageTokenInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLeverageTokenInfo(context.Background(), currency.NewCode("BTC3L"))
	if err != nil {
		t.Error(err)
	}
}

func TestGetLeveragedTokenMarket(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLeveragedTokenMarket(context.Background(), currency.EMPTYCODE)
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Fatalf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.GetLeveragedTokenMarket(context.Background(), currency.NewCode("BTC3L"))
	if err != nil {
		t.Error(err)
	}
}

func TestPurchaseLeverageToken(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.PurchaseLeverageToken(context.Background(), currency.BTC3L, 100, "")
	if err != nil {
		t.Error(err)
	}
}

func TestRedeemLeverageToken(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.RedeemLeverageToken(context.Background(), currency.BTC3L, 100, "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetPurchaseAndRedemptionRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetPurchaseAndRedemptionRecords(context.Background(), currency.EMPTYCODE, "", "", time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestToggleMarginTrade(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.ToggleMarginTrade(context.Background(), true)
	if err != nil {
		t.Error(err)
	}
}

func TestSetSpotMarginTradeLeverage(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	err := b.SetSpotMarginTradeLeverage(context.Background(), 3)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
}

func TestGetMarginCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetMarginCoinInfo(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBorrowableCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetBorrowableCoinInfo(context.Background(), currency.EMPTYCODE)
	if err != nil {
		t.Error(err)
	}
}

func TestGetInterestAndQuota(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetInterestAndQuota(context.Background(), currency.EMPTYCODE)
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.GetInterestAndQuota(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLoanAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLoanAccountInfo(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestBorrow(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.Borrow(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.Borrow(context.Background(), &LendArgument{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.Borrow(context.Background(), &LendArgument{Coin: currency.BTC})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Errorf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.Borrow(context.Background(), &LendArgument{Coin: currency.BTC, AmountToBorrow: 0.1})
	if err != nil {
		t.Error(err)
	}
}
func TestRepay(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.Repay(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, got %v", errNilArgument, err)
	}
	_, err = b.Repay(context.Background(), &LendArgument{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.Repay(context.Background(), &LendArgument{Coin: currency.BTC})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Errorf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.Repay(context.Background(), &LendArgument{Coin: currency.BTC, AmountToBorrow: 0.1})
	if err != nil {
		t.Error(err)
	}
}

func TestGetBorrowOrderDetail(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetBorrowOrderDetail(context.Background(), time.Time{}, time.Time{}, currency.BTC, 0, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetRepaymentOrderDetail(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetRepaymentOrderDetail(context.Background(), time.Time{}, time.Time{}, currency.BTC, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestToggleMarginTradeNormal(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.ToggleMarginTradeNormal(context.Background(), true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetProductInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetProductInfo(context.Background(), "78")
	if err != nil {
		t.Error(err)
	}
}

func TestGetInstitutionalLengingMarginCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetInstitutionalLengingMarginCoinInfo(context.Background(), "123")
	if err != nil {
		t.Error(err)
	}
}

func TestGetInstitutionalLoanOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetInstitutionalLoanOrders(context.Background(), "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetInstitutionalRepayOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetInstitutionalRepayOrders(context.Background(), time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLTV(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetLTV(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetC2CLendingCoinInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetC2CLendingCoinInfo(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestC2CDepositFunds(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.C2CDepositFunds(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Error(err)
	}
	_, err = b.C2CDepositFunds(context.Background(), &C2CLendingFundsParams{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.C2CDepositFunds(context.Background(), &C2CLendingFundsParams{Coin: currency.BTC})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Errorf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.C2CDepositFunds(context.Background(), &C2CLendingFundsParams{Coin: currency.BTC, Quantity: 1232})
	if err != nil {
		t.Error(err)
	}
}

func TestC2CRedeemFunds(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	_, err := b.C2CRedeemFunds(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Error(err)
	}
	_, err = b.C2CRedeemFunds(context.Background(), &C2CLendingFundsParams{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("expected %v, got %v", currency.ErrCurrencyCodeEmpty, err)
	}
	_, err = b.C2CRedeemFunds(context.Background(), &C2CLendingFundsParams{Coin: currency.BTC})
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Errorf("expected %v, got %v", order.ErrAmountBelowMin, err)
	}
	_, err = b.C2CRedeemFunds(context.Background(), &C2CLendingFundsParams{Coin: currency.BTC, Quantity: 1232})
	if err != nil {
		t.Error(err)
	}
}

func TestGetC2CLendingOrderRecords(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetC2CLendingOrderRecords(context.Background(), currency.EMPTYCODE, "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetC2CLendingAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetC2CLendingAccountInfo(context.Background(), currency.LTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBrokerEarning(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.GetBrokerEarning(context.Background(), "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func instantiateTradablePairs() error {
	err := b.UpdateTradablePairs(context.Background(), true)
	if err != nil {
		return err
	}
	tradables, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		return err
	}
	format, err := b.GetPairFormat(asset.Spot, true)
	if err != nil {
		return err
	}
	spotTradablePair = tradables[0].Format(format)
	tradables, err = b.GetEnabledPairs(asset.Linear)
	if err != nil {
		return err
	}
	format, err = b.GetPairFormat(asset.Linear, true)
	if err != nil {
		return err
	}
	linearTradablePair = tradables[0].Format(format)
	tradables, err = b.GetEnabledPairs(asset.Inverse)
	if err != nil {
		return err
	}
	format, err = b.GetPairFormat(asset.Inverse, true)
	if err != nil {
		return err
	}
	inverseTradablePair = tradables[0].Format(format)
	tradables, err = b.GetEnabledPairs(asset.Options)
	if err != nil {
		return err
	}
	format, err = b.GetPairFormat(asset.Options, true)
	if err != nil {
		return err
	}
	optionsTradablePair = tradables[0].Format(format)
	return nil
}

func TestUpdateAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.UpdateAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestFetchAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetWithdrawalsHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetWithdrawalsHistory(context.Background(), currency.BTC, asset.CoinMarginedFutures)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Errorf("expected %v, got %v", asset.ErrNotSupported, err)
	}
	_, err = b.GetWithdrawalsHistory(context.Background(), currency.BTC, asset.Spot)
	if err != nil {
		t.Error("GetWithdrawalsHistory()", err)
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetRecentTrades(context.Background(), spotTradablePair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRecentTrades(context.Background(), optionsTradablePair, asset.Options)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRecentTrades(context.Background(), inverseTradablePair, asset.Inverse)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRecentTrades(context.Background(), linearTradablePair, asset.Linear)
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetRecentTrades(context.Background(), spotTradablePair, asset.CoinMarginedFutures)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Error(err)
	}
}

func TestGetBybitServerTime(t *testing.T) {
	t.Parallel()
	_, err := b.GetBybitServerTime(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetServerTime(t *testing.T) {
	t.Parallel()
	_, err := b.GetServerTime(context.Background(), asset.Empty)
	if err != nil {
		t.Error(err)
	}
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	_, err := b.GetHistoricTrades(context.Background(), spotTradablePair, asset.Spot, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricTrades(context.Background(), linearTradablePair, asset.Linear, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricTrades(context.Background(), inverseTradablePair, asset.Inverse, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}
	_, err = b.GetHistoricTrades(context.Background(), optionsTradablePair, asset.Options, time.Time{}, time.Time{})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelBatchOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	var orderCancellationParams = []order.Cancel{{
		OrderID:   "1",
		Pair:      spotTradablePair,
		AssetType: asset.Spot}, {
		OrderID:   "1",
		Pair:      linearTradablePair,
		AssetType: asset.Linear}}
	_, err := b.CancelBatchOrders(context.Background(), orderCancellationParams)
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Errorf("expected %v, got %v", asset.ErrNotSupported, err)
	}
	orderCancellationParams = []order.Cancel{{
		OrderID:   "1",
		AccountID: "1",
		Pair:      optionsTradablePair,
		AssetType: asset.Options}, {
		OrderID:   "2",
		Pair:      optionsTradablePair,
		AssetType: asset.Options}}
	_, err = b.CancelBatchOrders(context.Background(), orderCancellationParams)
	if err != nil {
		t.Error(err)
	}
}

func TestWsConnect(t *testing.T) {
	t.Parallel()
	err := b.WsConnect()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 23)
}
func TestWsLinearConnect(t *testing.T) {
	t.Parallel()
	err := b.WsLinearConnect()
	if err != nil && !errors.Is(err, errWebsocketNotEnabled) {
		t.Error(err)
	}
}
func TestWsInverseConnect(t *testing.T) {
	t.Parallel()
	err := b.WsInverseConnect()
	if err != nil && !errors.Is(err, errWebsocketNotEnabled) {
		t.Error(err)
	}
}
func TestWsOptionsConnect(t *testing.T) {
	t.Parallel()
	err := b.WsOptionsConnect()
	if err != nil && !errors.Is(err, errWebsocketNotEnabled) {
		t.Error(err)
	}
	time.Sleep(time.Second * 23)
}

var pushDataMap = map[string]string{
	"Orderbook Snapshot":   `{"topic":"orderbook.50.BTCUSDT","ts":1690719970602,"type":"snapshot","data":{"s":"BTCUSDT","b":[["29328.25","3.911681"],["29328.21","0.117584"],["29328.19","0.511493"],["29328.16","0.013639"],["29328","0.1646"],["29327.99","1"],["29327.98","0.681309"],["29327.53","0.001"],["29327.46","0.000048"],["29327","0.046517"],["29326.99","0.077528"],["29326.55","0.026808"],["29326.48","0.03"],["29326","0.1646"],["29325.99","0.00075"],["29325.93","0.409862"],["29325.92","0.745"],["29325.87","0.511533"],["29325.85","0.00018"],["29325.42","0.001023"],["29325.41","0.68199"],["29325.36","0.006309"],["29325.35","0.0153"],["29324.97","0.903728"],["29324.96","1.506212"],["29324.49","0.016966"],["29324.38","0.0341"],["29324.17","1.4535"],["29324","0.1646"],["29323.99","0.00075"],["29323.92","0.050492"],["29323.77","1.023141"],["29323.72","0.12"],["29323.48","0.0153"],["29323.26","0.001362"],["29322.78","0.464948"],["29322.77","0.745"],["29322.76","0.0153"],["29322.73","0.013633"],["29322.67","0.53"],["29322.62","0.01"],["29322.04","0.97036"],["29322","0.1656"],["29321.99","0.00075"],["29321.56","0.0341"],["29321.52","0.613945"],["29321.51","0.13"],["29321.4","0.002"],["29321.18","0.196788"],["29321.13","0.34104"]],"a":[["29328.26","1.256884"],["29328.36","0.013639"],["29328.97","0.51148"],["29329","0.002046"],["29329.2","0.035597"],["29329.27","0.001"],["29329.44","0.03523"],["29329.99","0.791676"],["29330","0.546264"],["29330.28","0.001"],["29330.35","0.767184"],["29330.5","0.002725"],["29330.51","0.0341"],["29330.79","0.03"],["29330.81","0.158412"],["29330.93","0.68199"],["29330.95","0.282036"],["29331","0.041"],["29331.13","0.0003"],["29331.19","0.01"],["29331.53","0.050164"],["29331.54","0.008573"],["29331.99","0.26305"],["29332.11","0.008124"],["29332.21","0.8721"],["29332.22","1.4535"],["29332.41","0.157"],["29332.58","0.001023"],["29332.59","0.0153"],["29332.84","0.679527"],["29332.85","1.022812"],["29332.98","0.200071"],["29333.01","1.13254"],["29333.24","0.0153"],["29333.25","0.001362"],["29333.35","0.625"],["29333.37","0.01"],["29333.56","0.0341"],["29333.68","0.21795"],["29333.85","0.182562"],["29333.98","0.0003"],["29333.99","0.00105"],["29334.16","0.009132"],["29334.29","0.0003"],["29334.48","0.029675"],["29334.7","0.00086"],["29334.99","0.006838"],["29335","0.002177"],["29335.18","0.013622"],["29335.32","0.034099"]],"u":51668654,"seq":10194901787}}`,
	"Orderbook Update":     `{"topic":"orderbook.50.ACAUSDT","ts":1690719548494,"type":"snapshot","data":{"s":"ACAUSDT","b":[["0.0657","5363.66"],["0.0646","7910.21"],["0.0645","1435.73"],["0.0644","1552.8"],["0.0642","6904.01"],["0.064","3232.64"],["0.0639","106"],["0.0637","100"],["0.0636","25.62"],["0.0635","209.43"],["0.0631","237.47"],["0.063","258.13"],["0.0627","318.97"],["0.0625","10066.99"],["0.0624","16.1"],["0.0623","41.72"],["0.0622","1624.59"],["0.0621","402.57"],["0.0616","10.65"],["0.0613","652"],["0.061","1081.97"],["0.0604","413.91"],["0.06","1471.82"],["0.0597","15000"],["0.0595","15000"],["0.0593","608.77"],["0.0591","430.79"],["0.059","444"],["0.0586","4536.97"],["0.0584","1533.58"],["0.0583","3764.43"],["0.0581","3072.34"],["0.058","2654.9"],["0.0579","1022.23"],["0.0576","1931.71"],["0.0574","2545.88"],["0.0573","821.27"],["0.0571","2957"],["0.0568","1483.57"],["0.0561","392.24"],["0.0555","900.9"],["0.055","322.15"],["0.0549","182"],["0.0545","30"],["0.0536","24.24"],["0.0535","1869.15"],["0.053","40"],["0.0529","189"],["0.0525","701.66"],["0.0521","1122.64"]],"a":[["0.0661","3320.27"],["0.0662","8667.02"],["0.0663","6087.91"],["0.0664","6060.61"],["0.0684","591.31"],["0.0689","155.77"],["0.069","1148.02"],["0.0694","2421.86"],["0.0699","155.77"],["0.07","445.87"],["0.0701","142.65"],["0.071","2131.4"],["0.0718","1447.83"],["0.072","420.62"],["0.0743","1399.15"],["0.0745","1481.62"],["0.0747","32.97"],["0.0748","900.38"],["0.0749","209.44"],["0.075","124.49"],["0.0757","41.9"],["0.0762","657.43"],["0.077","48.77"],["0.0779","96.26"],["0.078","12305.94"],["0.079","29.77"],["0.0797","512.26"],["0.0799","743.29"],["0.08","5050.7"],["0.0814","11.71"],["0.0815","75.93"],["0.0817","403"],["0.082","817.43"],["0.0825","768.47"],["0.0828","388.77"],["0.083","150.53"],["0.0835","18"],["0.084","10776.95"],["0.0841","1465.17"],["0.0848","15000"],["0.085","16976.73"],["0.0853","798.45"],["0.0856","5239.19"],["0.0857","5134.18"],["0.0858","3885.13"],["0.0859","3691.71"],["0.086","16847.35"],["0.0862","898.68"],["0.0863","994.24"],["0.0865","1251.56"]],"u":4694899,"seq":12206894097}}`,
	"Public Trade":         `{"topic":"publicTrade.ATOM2SUSDT","ts":1690720953113,"type":"snapshot","data":[{"i":"2200000000067341890","T":1690720953111,"p":"3.6279","v":"1.3637","S":"Sell","s":"ATOM2SUSDT","BT":false}]}`,
	"Public Linear Ticker": `{ "topic": "tickers.BTCUSDT", "type": "snapshot", "data": { "symbol": "BTCUSDT", "tickDirection": "PlusTick", "price24hPcnt": "0.017103", "lastPrice": "17216.00", "prevPrice24h": "16926.50", "highPrice24h": "17281.50", "lowPrice24h": "16915.00", "prevPrice1h": "17238.00", "markPrice": "17217.33", "indexPrice": "17227.36", "openInterest": "68744.761", "openInterestValue": "1183601235.91", "turnover24h": "1570383121.943499", "volume24h": "91705.276", "nextFundingTime": "1673280000000", "fundingRate": "-0.000212", "bid1Price": "17215.50", "bid1Size": "84.489", "ask1Price": "17216.00", "ask1Size": "83.020" }, "cs": 24987956059, "ts": 1673272861686 }`,
	"Public Option Ticker": `{ "id": "tickers.BTC-6JAN23-17500-C-2480334983-1672917511074", "topic": "tickers.BTC-6JAN23-17500-C", "ts": 1672917511074, "data": { "symbol": "BTC-6JAN23-17500-C", "bidPrice": "0", "bidSize": "0", "bidIv": "0", "askPrice": "10", "askSize": "5.1", "askIv": "0.514", "lastPrice": "10", "highPrice24h": "25", "lowPrice24h": "5", "markPrice": "7.86976724", "indexPrice": "16823.73", "markPriceIv": "0.4896", "underlyingPrice": "16815.1", "openInterest": "49.85", "turnover24h": "446802.8473", "volume24h": "26.55", "totalVolume": "86", "totalTurnover": "1437431", "delta": "0.047831", "gamma": "0.00021453", "vega": "0.81351067", "theta": "-19.9115368", "predictedDeliveryPrice": "0", "change24h": "-0.33333334" }, "type": "snapshot" }`,
	"Public Ticker":        `{"topic":"tickers.APTUSDC","ts":1690724804979,"type":"snapshot","cs":11505608330,"data":{"symbol":"APTUSDC","lastPrice":"7.0884","highPrice24h":"7.19","lowPrice24h":"7.0666","prevPrice24h":"7.0767","volume24h":"642.45","turnover24h":"4568.920448","price24hPcnt":"0.0017","usdIndexPrice":"7.07930012"}}`,
	"Public Kline":         `{ "topic": "kline.5.BTCUSDT", "data": [ { "start": 1672324800000, "end": 1672325099999, "interval": "5", "open": "16649.5", "close": "16677", "high": "16677", "low": "16608", "volume": "2.081", "turnover": "34666.4005", "confirm": false, "timestamp": 1672324988882 } ], "ts": 1672324988882,"type": "snapshot"}`,
	"Public Liquidiation":  `{ "data": { "price": "0.03803", "side": "Buy", "size": "1637", "symbol": "GALAUSDT", "updatedTime": 1673251091822 }, "topic": "liquidation.GALAUSDT", "ts": 1673251091822, "type": "snapshot" }`,
	"Public LT Kline":      `{ "type": "snapshot", "topic": "kline_lt.5.EOS3LUSDT", "data": [ { "start": 1672325100000, "end": 1672325399999, "interval": "5", "open": "0.416039541212402799", "close": "0.41477848043290448", "high": "0.416039541212402799", "low": "0.409734237314911206", "confirm": false, "timestamp": 1672325322393 } ], "ts": 1672325322393 }`,
	"Public LT Ticker":     `{ "topic": "tickers_lt.EOS3LUSDT", "ts": 1672325446847, "type": "snapshot", "data": { "symbol": "EOS3LUSDT", "lastPrice": "0.41477848043290448", "highPrice24h": "0.435285472510871305", "lowPrice24h": "0.394601507960931382", "prevPrice24h": "0.431502290172376349", "price24hPcnt": "-0.0388" } }`,
	"Public LT Navigation": `{ "topic": "lt.EOS3LUSDT", "ts": 1672325564669, "type": "snapshot", "data": { "symbol": "EOS3LUSDT", "time": 1672325564554, "nav": "0.413517419653406162", "basketPosition": "1.261060779498318641", "leverage": "2.656197506416192150", "basketLoan": "-0.684866519289629374", "circulation": "72767.309468460367138199", "basket": "91764.000000292013277472" } }`,
	"Private Position":     `{"id": "59232430b58efe-5fc5-4470-9337-4ce293b68edd", "topic": "position", "creationTime": 1672364174455, "data": [ { "positionIdx": 0, "tradeMode": 0, "riskId": 41, "riskLimitValue": "200000", "symbol": "XRPUSDT", "side": "Buy", "size": "75", "entryPrice": "0.3615", "leverage": "10", "positionValue": "27.1125", "positionBalance": "0", "markPrice": "0.3374", "positionIM": "2.72589075", "positionMM": "0.28576575", "takeProfit": "0", "stopLoss": "0", "trailingStop": "0", "unrealisedPnl": "-1.8075", "cumRealisedPnl": "0.64782276", "createdTime": "1672121182216", "updatedTime": "1672364174449", "tpslMode": "Full", "liqPrice": "", "bustPrice": "", "category": "linear","positionStatus":"Normal","adlRankIndicator":2}]}`,
	"Private Order":        `{ "id": "5923240c6880ab-c59f-420b-9adb-3639adc9dd90", "topic": "order", "creationTime": 1672364262474, "data": [ { "symbol": "ETH-30DEC22-1400-C", "orderId": "5cf98598-39a7-459e-97bf-76ca765ee020", "side": "Sell", "orderType": "Market", "cancelType": "UNKNOWN", "price": "72.5", "qty": "1", "orderIv": "", "timeInForce": "IOC", "orderStatus": "Filled", "orderLinkId": "", "lastPriceOnCreated": "", "reduceOnly": false, "leavesQty": "", "leavesValue": "", "cumExecQty": "1", "cumExecValue": "75", "avgPrice": "75", "blockTradeId": "", "positionIdx": 0, "cumExecFee": "0.358635", "createdTime": "1672364262444", "updatedTime": "1672364262457", "rejectReason": "EC_NoError", "stopOrderType": "", "tpslMode": "", "triggerPrice": "", "takeProfit": "", "stopLoss": "", "tpTriggerBy": "", "slTriggerBy": "", "tpLimitPrice": "", "slLimitPrice": "", "triggerDirection": 0, "triggerBy": "", "closeOnTrigger": false, "category": "option", "placeType": "price", "smpType": "None", "smpGroup": 0, "smpOrderId": "" } ] }`,
	"Private Wallet":       `{ "id": "5923242c464be9-25ca-483d-a743-c60101fc656f", "topic": "wallet", "creationTime": 1672364262482, "data": [ { "accountIMRate": "0.016", "accountMMRate": "0.003", "totalEquity": "12837.78330098", "totalWalletBalance": "12840.4045924", "totalMarginBalance": "12837.78330188", "totalAvailableBalance": "12632.05767702", "totalPerpUPL": "-2.62129051", "totalInitialMargin": "205.72562486", "totalMaintenanceMargin": "39.42876721", "coin": [ { "coin": "USDC", "equity": "200.62572554", "usdValue": "200.62572554", "walletBalance": "201.34882644", "availableToWithdraw": "0", "availableToBorrow": "1500000", "borrowAmount": "0", "accruedInterest": "0", "totalOrderIM": "0", "totalPositionIM": "202.99874213", "totalPositionMM": "39.14289747", "unrealisedPnl": "74.2768991", "cumRealisedPnl": "-209.1544627", "bonus": "0" }, { "coin": "BTC", "equity": "0.06488393", "usdValue": "1023.08402268", "walletBalance": "0.06488393", "availableToWithdraw": "0.06488393", "availableToBorrow": "2.5", "borrowAmount": "0", "accruedInterest": "0", "totalOrderIM": "0", "totalPositionIM": "0", "totalPositionMM": "0", "unrealisedPnl": "0", "cumRealisedPnl": "0", "bonus": "0" }, { "coin": "ETH", "equity": "0", "usdValue": "0", "walletBalance": "0", "availableToWithdraw": "0", "availableToBorrow": "26", "borrowAmount": "0", "accruedInterest": "0", "totalOrderIM": "0", "totalPositionIM": "0", "totalPositionMM": "0", "unrealisedPnl": "0", "cumRealisedPnl": "0", "bonus": "0" }, { "coin": "USDT", "equity": "11726.64664904", "usdValue": "11613.58597018", "walletBalance": "11728.54414904", "availableToWithdraw": "11723.92075829", "availableToBorrow": "2500000", "borrowAmount": "0", "accruedInterest": "0", "totalOrderIM": "0", "totalPositionIM": "2.72589075", "totalPositionMM": "0.28576575", "unrealisedPnl": "-1.8975", "cumRealisedPnl": "0.64782276", "bonus": "0" }, { "coin": "EOS3L", "equity": "215.0570412", "usdValue": "0", "walletBalance": "215.0570412", "availableToWithdraw": "215.0570412", "availableToBorrow": "0", "borrowAmount": "0", "accruedInterest": "", "totalOrderIM": "0", "totalPositionIM": "0", "totalPositionMM": "0", "unrealisedPnl": "0", "cumRealisedPnl": "0", "bonus": "0" }, { "coin": "BIT", "equity": "1.82", "usdValue": "0.48758257", "walletBalance": "1.82", "availableToWithdraw": "1.82", "availableToBorrow": "0", "borrowAmount": "0", "accruedInterest": "", "totalOrderIM": "0", "totalPositionIM": "0", "totalPositionMM": "0", "unrealisedPnl": "0", "cumRealisedPnl": "0", "bonus": "0" } ], "accountType": "UNIFIED", "accountLTV": "0.017" } ] }`,
	"Private Greek":        `{ "id": "592324fa945a30-2603-49a5-b865-21668c29f2a6", "topic": "greeks", "creationTime": 1672364262482, "data": [ { "baseCoin": "ETH", "totalDelta": "0.06999986", "totalGamma": "-0.00000001", "totalVega": "-0.00000024", "totalTheta": "0.00001314" } ] }`,
}

func TestPushData(t *testing.T) {
	t.Parallel()
	for x := range pushDataMap {
		err := b.wsHandleData(asset.Spot, []byte(pushDataMap[x]))
		if err != nil {
			t.Errorf("%s: %v", x, err)
		}
	}
}
