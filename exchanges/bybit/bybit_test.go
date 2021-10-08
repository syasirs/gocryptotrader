package bybit

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
)

// Please supply your own keys here to do authenticated endpoint testing
const (
	apiKey    = ""
	apiSecret = ""
)

var by Bybit

func TestMain(m *testing.M) {
	by.SetDefaults()
	by.Verbose = true
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal(err)
	}

	exchCfg, err := cfg.GetExchangeConfig("Bybit")
	if err != nil {
		log.Fatal(err)
	}

	exchCfg.API.AuthenticatedSupport = true
	exchCfg.API.AuthenticatedWebsocketSupport = true
	exchCfg.API.Credentials.Key = apiKey
	exchCfg.API.Credentials.Secret = apiSecret
	by.Websocket = sharedtestvalues.NewTestWebsocket()
	err = by.Setup(exchCfg)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// Ensures that this exchange package is compatible with IBotExchange
func TestInterface(t *testing.T) {
	var e exchange.IBotExchange
	if e = new(Bybit); e == nil {
		t.Fatal("unable to allocate exchange")
	}
}

func areTestAPIKeysSet() bool {
	return by.ValidateAPICredentials()
}

func TestGetAllPairs(t *testing.T) {
	t.Parallel()

	_, err := by.GetAllPairs()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetOrderBook(t *testing.T) {
	t.Parallel()

	_, err := by.GetOrderBook("BTCUSDT", 100)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()

	_, err := by.GetTrades("BTCUSDT", 100)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetKlines(t *testing.T) {
	t.Parallel()

	_, err := by.GetKlines("BTCUSDT", "5m", 2000, time.Now().Add(-time.Hour*1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet24HrsChange(t *testing.T) {
	t.Parallel()

	_, err := by.Get24HrsChange("BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.Get24HrsChange("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLastTradedPrice(t *testing.T) {
	t.Parallel()

	_, err := by.GetLastTradedPrice("BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetLastTradedPrice("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetBestBidAskPrice(t *testing.T) {
	t.Parallel()

	_, err := by.GetBestBidAskPrice("BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetBestBidAskPrice("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreatePostOrder(t *testing.T) {
	t.Parallel()

	r, err := by.CreatePostOrder(&PlaceOrderRequest{
		Symbol:      "BTCUSDT",
		Quantity:    1,
		Side:        "BUY",
		TradeType:   "LIMIT",
		TimeInForce: "GTC",
		Price:       100,
		OrderLinkID: "linkID",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestQueryOrder(t *testing.T) {
	t.Parallel()

	r, err := by.QueryOrder("0", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestCancelExistingOrder(t *testing.T) {
	t.Parallel()

	r, err := by.CancelExistingOrder("", "linkID")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestBatchCancelOrder(t *testing.T) {
	t.Parallel()

	r, err := by.BatchCancelOrder("", "BUY", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestListOpenOrders(t *testing.T) {
	t.Parallel()

	r, err := by.ListOpenOrders("", "BUY", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestListPastOrders(t *testing.T) {
	t.Parallel()

	r, err := by.ListPastOrders("", "BUY", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestGetTradeHistory(t *testing.T) {
	t.Parallel()

	r, err := by.GetTradeHistory("", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestGetWalletBalance(t *testing.T) {
	t.Parallel()

	r, err := by.GetWalletBalance()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestWsSubscription(t *testing.T) {
	t.Parallel()

	pressXToJSON := []byte(`{
		"symbol": "BTCUSDT",
		"event": "sub",
		"topic": "trade",
		"params": {
			"binary": false
		}
	}`)
	err := by.wsHandleData(pressXToJSON)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWsUnsubscribe(t *testing.T) {
	t.Parallel()

	pressXToJSON := []byte(`{
		"symbol":"BTCUSDT",
		"event": "cancel",
		"topic":"trade",
		"params": {
			"binary": false
		}
	}`)
	err := by.wsHandleData(pressXToJSON)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWsTrade(t *testing.T) {
	by.SetSaveTradeDataStatus(true)
	t.Parallel()

	pressXToJSON := []byte(`{
		"topic": "trade",
		"params": {
			"symbol": "BTCUSDT",
			"binary": false,
			"symbolName": "BTCUSDT"
		},
		"data": {
			"v": "564265886622695424",
			"t": 1582001735462,
			"p": "9787.5",
			"q": "0.195009",
			"m": true
		}
	}`)
	err := by.wsHandleData(pressXToJSON)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWsOrderbook(t *testing.T) {
	pressXToJSON := []byte(`{
		"topic": "depth",
		"params": {
		  "symbol": "BTCUSDT",
		  "binary": false,
		  "symbolName": "BTCUSDT"
		},
		"data": {
			"s": "BTCUSDT",
			"t": 1582001376853,
			"v": "13850022_2",
			"b": [
				[
					"9780.79",
					"0.01"
				],
				[
					"9780.5",
					"0.1"
				],
				[
					"9780.4",
					"0.517813"
				]
			],
			"a": [
				[
					"9781.21",
					"0.042842"
				],
				[
					"9782",
					"0.3"
				],
				[
					"9782.1",
					"0.226"
				]
			]
		}
	}`)
	err := by.wsHandleData(pressXToJSON)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWsTicker(t *testing.T) {
	t.Parallel()

	pressXToJSON := []byte(`{
		"topic": "bookTicker",
		"params": {
			"symbol": "BTCUSDT",
			"binary": false,
			"symbolName": "BTCUSDT"
		},
		"data": {
			"symbol": "BTCUSDT",
			"bidPrice": "9797.79",
			"bidQty": "0.177976",
			"askPrice": "9799",
			"askQty": "0.65",
			"time": 1582001830346
		}
	}`)
	err := by.wsHandleData(pressXToJSON)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFuturesOrderbook(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	_, err = by.GetFuturesOrderbook(pair)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetFuturesKlineData(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	_, err = by.GetFuturesKlineData(pair, "M", 5, time.Time{})
	if err != nil {
		t.Error(err)
	}

	_, err = by.GetFuturesKlineData(pair, "60", 5, time.Unix(1577836800, 0))
	if err != nil {
		t.Error(err)
	}
}

func TestGetFuturesSymbolPriceTicker(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetFuturesSymbolPriceTicker(pair)
	if err != nil {
		t.Error(err)
	}
}

func TestGetPublicTrades(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetPublicTrades(pair, 0, 0)
	if err != nil {
		t.Error(err)
	}

	_, err = by.GetPublicTrades(pair, 0, 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetSymbolsInfo(t *testing.T) {
	_, err := by.GetSymbolsInfo()
	if err != nil {
		t.Error(err)
	}
}

func TestGetMarkPriceKline(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetMarkPriceKline(pair, "D", 0, time.Unix(1577836800, 0))
	if err != nil {
		t.Error(err)
	}
}

func TestGetIndexPriceKline(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetIndexPriceKline(pair, "D", 0, time.Unix(1577836800, 0))
	if err != nil {
		t.Error(err)
	}
}

func TestGetPremiumIndexPriceKline(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetPremiumIndexPriceKline(pair, "D", 0, time.Unix(1577836800, 0))
	if err != nil {
		t.Error(err)
	}
}

func TestGetOpenInterest(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetOpenInterest(pair, "5min", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLatestBigDeal(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetLatestBigDeal(pair, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAccountRatio(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetAccountRatio(pair, "1d", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetRiskLimit(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetRiskLimit(pair)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLastFundingRate(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetLastFundingRate(pair)
	if err != nil {
		t.Error(err)
	}
}

func TestGetServerTime(t *testing.T) {
	_, err := by.GetServerTime()
	if err != nil {
		t.Error(err)
	}
}

func TestGetAnnouncement(t *testing.T) {
	_, err := by.GetAnnouncement()
	if err != nil {
		t.Error(err)
	}
}

func TestCreateFuturesOrder(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.CreateFuturesOrder(pair, "Buy", "Limit", "GTC", "", "", "", 1, 0, 0, 0, false, false)
	if err != nil {
		t.Error(err)
	}
}

func TestGetActiveFuturesOrders(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.GetActiveFuturesOrders(pair, "", "", "", 0)
	if err != nil {
		t.Error(err)
	}

	_, err = by.GetActiveFuturesOrders(pair, "Filled", "", "", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestCancelActiveFuturesOrders(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = by.CancelActiveFuturesOrders(pair, "", "")
	if err != nil {
		t.Error(err)
	}
}
