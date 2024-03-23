package cryptodotcom

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/core"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// Please supply your own keys here to do authenticated endpoint testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

var cr = &Cryptodotcom{}

func TestMain(m *testing.M) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal(err)
	}
	exchCfg, err := cfg.GetExchangeConfig("Cryptodotcom")
	if err != nil {
		log.Fatal(err)
	}
	exchCfg.API.Credentials.Key = apiKey
	exchCfg.API.Credentials.Secret = apiSecret
	cr.SetDefaults()
	if apiKey != "" && apiSecret != "" {
		exchCfg.API.AuthenticatedSupport = true
		exchCfg.API.AuthenticatedWebsocketSupport = true
	}
	cr.Websocket = sharedtestvalues.NewTestWebsocket()
	err = cr.Setup(exchCfg)
	if err != nil {
		log.Fatal(err)
	}
	if apiKey != "" && apiSecret != "" {
		cr.Websocket.SetCanUseAuthenticatedEndpoints(true)
	}
	cr.Websocket.DataHandler = sharedtestvalues.GetWebsocketInterfaceChannelOverride()
	cr.Websocket.TrafficAlert = sharedtestvalues.GetWebsocketStructChannelOverride()
	setupWS()
	os.Exit(m.Run())
}

func TestGetSymbols(t *testing.T) {
	t.Parallel()
	_, err := cr.GetInstruments(context.Background())
	assert.NoError(t, err)
}

func TestGetOrderbook(t *testing.T) {
	t.Parallel()
	_, err := cr.GetOrderbook(context.Background(), "BTC_USDT", 0)
	assert.NoError(t, err)
}

func TestGetCandlestickDetail(t *testing.T) {
	t.Parallel()
	_, err := cr.GetCandlestickDetail(context.Background(), "BTC_USDT", kline.FiveMin)
	assert.NoError(t, err)
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := cr.GetTicker(context.Background(), "BTC_USDT")
	assert.NoError(t, err)
	_, err = cr.GetTicker(context.Background(), "")
	assert.NoError(t, err)
}

func TestGetTrades(t *testing.T) {
	t.Parallel()
	_, err := cr.GetTrades(context.Background(), "BTC_USDT")
	assert.NoError(t, err)
}

func TestWithdrawFunds(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	_, err := cr.WithdrawFunds(context.Background(), currency.BTC, 10, core.BitcoinDonationAddress, "", "", "")
	assert.NoError(t, err)
	_, err = cr.WsCreateWithdrawal(currency.BTC, 10, core.BitcoinDonationAddress, "", "", "")
	assert.NoError(t, err)
}

func TestGetCurrencyNetworks(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetCurrencyNetworks(context.Background())
	assert.NoError(t, err)
}

func TestGetWithdrawalHistory(t *testing.T) {
	t.Parallel()
	var resp *WithdrawalResponse
	err := json.Unmarshal([]byte(getWithdrawalHistory), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetWithdrawalHistory(context.Background())
	assert.NoError(t, err)
	_, err = cr.WsRetriveWithdrawalHistory()
	assert.NoError(t, err)
}

func TestGetDepositHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetDepositHistory(context.Background(), currency.EMPTYCODE, time.Time{}, time.Time{}, 20, 0, 0)
	assert.NoError(t, err)
}

func TestGetPersonalDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetPersonalDepositAddress(context.Background(), currency.BTC)
	assert.NoError(t, err)
}

func TestGetAccountSummary(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetAccountSummary(context.Background(), currency.USDT)
	assert.NoError(t, err)
	_, err = cr.WsRetriveAccountSummary(currency.BTC)
	assert.NoError(t, err)
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	arg := &CreateOrderParam{InstrumentName: "BTC_USDT", Side: order.Buy, OrderType: orderTypeToString(order.Limit), Price: 123, Quantity: 12}
	_, err := cr.CreateOrder(context.Background(), arg)
	assert.NoError(t, err)
	_, err = cr.WsPlaceOrder(arg)
	assert.NoError(t, err)
}

func TestCancelExistingOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	err := cr.CancelExistingOrder(context.Background(), "BTC_USDT", "1232412")
	assert.NoError(t, err)
	err = cr.WsCancelExistingOrder("BTC_USDT", "1232412")
	assert.NoError(t, err)
}

func TestGetPrivateTrades(t *testing.T) {
	t.Parallel()
	var resp *PersonalTrades
	err := json.Unmarshal([]byte(getPrivateTrades), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetPrivateTrades(context.Background(), "", time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
	_, err = cr.WsRetrivePrivateTrades("", time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

func TestGetOrderDetail(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetOrderDetail(context.Background(), "1234")
	assert.NoError(t, err)
	_, err = cr.WsRetriveOrderDetail("1234")
	assert.NoError(t, err)
}

func TestGetPersonalOpenOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetPersonalOpenOrders(context.Background(), "", 0, 0)
	assert.NoError(t, err)
	_, err = cr.WsRetrivePersonalOpenOrders("", 0, 0)
	assert.NoError(t, err)
}

func TestGetPersonalOrderHistory(t *testing.T) {
	t.Parallel()
	var resp *PersonalOrdersResponse
	err := json.Unmarshal([]byte(getPersonalOrderHistory), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	resp, err = cr.GetPersonalOrderHistory(context.Background(), "", time.Time{}, time.Time{}, 0, 20)
	assert.NoError(t, err)
	_, err = cr.WsRetrivePersonalOrderHistory("", time.Time{}, time.Time{}, 0, 20)
	assert.NoError(t, err)
}

func TestCreateOrderList(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	_, err := cr.CreateOrderList(context.Background(), "LIST", []CreateOrderParam{
		{
			InstrumentName: "BTC_USDT", ClientOrderID: "", TimeInForce: "", Side: order.Buy, OrderType: orderTypeToString(order.Limit), PostOnly: false, TriggerPrice: 0, Price: 123, Quantity: 12, Notional: 0,
		}})
	assert.NoError(t, err)
	_, err = cr.WsCreateOrderList("LIST", []CreateOrderParam{
		{
			InstrumentName: "BTC_USDT", ClientOrderID: "", TimeInForce: "", Side: order.Buy, OrderType: orderTypeToString(order.Limit), PostOnly: false, TriggerPrice: 0, Price: 123, Quantity: 12, Notional: 0,
		}})
	assert.NoError(t, err)
}

func TestCancelOrderList(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	_, err := cr.CancelOrderList(context.Background(), []CancelOrderParam{
		{InstrumentName: "BTC_USDT", OrderID: "1234567"}, {InstrumentName: "BTC_USDT",
			OrderID: "123450067"}})
	assert.NoError(t, err)
	_, err = cr.WsCancelOrderList([]CancelOrderParam{
		{InstrumentName: "BTC_USDT", OrderID: "1234567"}, {InstrumentName: "BTC_USDT",
			OrderID: "123450067"}})
	assert.NoError(t, err)
}

func TestCancelAllPersonalOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	err = cr.CancelAllPersonalOrders(context.Background(), enabledPairs[0].String())
	assert.NoError(t, err)
	err = cr.WsCancelAllPersonalOrders(enabledPairs[0].String())
	assert.NoError(t, err)
}

func TestGetAccounts(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetAccounts(context.Background())
	assert.NoError(t, err)
}

func TestGetTransactions(t *testing.T) {
	t.Parallel()
	var resp *TransactionResponse
	err := json.Unmarshal([]byte(getTransactions), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetTransactions(context.Background(), "BTCUSD-PERP", "", time.Time{}, time.Time{}, 0)
	assert.NoError(t, err)
}

func TestCreateSubAccountTransfer(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	err := cr.CreateSubAccountTransfer(context.Background(), "destination_address", core.BitcoinDonationAddress, currency.USDT, 1232)
	assert.NoError(t, err)
}

func TestGetOTCUser(t *testing.T) {
	t.Parallel()
	var resp *OTCTrade
	err := json.Unmarshal([]byte(getOtcUser), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetOTCUser(context.Background())
	assert.NoError(t, err)
}

func TestGetOTCInstruments(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetOTCInstruments(context.Background())
	assert.NoError(t, err)
}

func TestRequestOTCQuote(t *testing.T) {
	t.Parallel()
	var resp *OTCQuoteResponse
	err := json.Unmarshal([]byte(requestOTCQuote), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.RequestOTCQuote(context.Background(), currency.NewPair(currency.BTC, currency.USDT), .001, 232, "BUY")
	assert.NoError(t, err)
}

func TestAcceptOTCQuote(t *testing.T) {
	t.Parallel()
	var resp *AcceptQuoteResponse
	err := json.Unmarshal([]byte(acceptOTCQuote), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.AcceptOTCQuote(context.Background(), "12323123", "")
	assert.NoError(t, err)
}

func TestGetOTCQuoteHistory(t *testing.T) {
	t.Parallel()
	var resp *QuoteHistoryResponse
	err := json.Unmarshal([]byte(getOTCQuoteHistory), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetOTCQuoteHistory(context.Background(), currency.EMPTYPAIR, time.Time{}, time.Time{}, 0, 10)
	assert.NoError(t, err)
}

func TestGetOTCTradeHistory(t *testing.T) {
	t.Parallel()
	var resp *OTCTradeHistoryResponse
	err := json.Unmarshal([]byte(getOTCTradeHistory), &resp)
	require.NoError(t, err)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err = cr.GetOTCTradeHistory(context.Background(), currency.NewPair(currency.BTC, currency.USDT), time.Time{}, time.Time{}, 0, 0)
	assert.NoError(t, err)
}

// wrapper test functions

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	_, err := cr.FetchTradablePairs(context.Background(), asset.Spot)
	assert.NoError(t, err)
}

func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	_, err = cr.UpdateTicker(context.Background(), enabledPairs[0], asset.Spot)
	assert.NoError(t, err)
}

func TestUpdateTickers(t *testing.T) {
	t.Parallel()
	err := cr.UpdateTickers(context.Background(), asset.Spot)
	assert.NoError(t, err)
}

func TestFetchTicker(t *testing.T) {
	t.Parallel()
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	_, err = cr.FetchTicker(context.Background(), enabledPairs[0], asset.Spot)
	assert.NoError(t, err)
}

func TestFetchOrderbook(t *testing.T) {
	t.Parallel()
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	_, err = cr.FetchOrderbook(context.Background(), enabledPairs[1], asset.Spot)
	assert.NoError(t, err)
}

func TestUpdateOrderbook(t *testing.T) {
	t.Parallel()
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	_, err = cr.UpdateOrderbook(context.Background(), enabledPairs[1], asset.Spot)
	assert.NoError(t, err)
}

func TestUpdateAccountInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.UpdateAccountInfo(context.Background(), asset.Spot)
	assert.NoError(t, err)
}

func TestGetWithdrawalsHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetWithdrawalsHistory(context.Background(), currency.BTC, asset.Spot)
	assert.NoError(t, err)
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	_, err := cr.GetRecentTrades(context.Background(), currency.NewPair(currency.BTC, currency.USDT), asset.Spot)
	assert.NoError(t, err)
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	_, err := cr.GetHistoricTrades(context.Background(), currency.NewPair(currency.BTC, currency.USDT), asset.Spot, time.Now().Add(-time.Hour*4), time.Now())
	assert.NoError(t, err)
}

func TestGetFundingHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetAccountFundingHistory(context.Background())
	assert.NoError(t, err)
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	require.NoError(t, err)
	startTime := time.Now().Add(-time.Minute * 40)
	endTime := time.Now()
	_, err = cr.GetHistoricCandles(context.Background(), enabledPairs[0], asset.Spot, kline.OneDay, startTime, endTime)
	require.NoError(t, err)
	_, err = cr.GetHistoricCandles(context.Background(), enabledPairs[0], asset.Spot, kline.FiveMin, startTime, endTime)
	assert.NoError(t, err)
}

func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	enabledPairs, err := cr.GetEnabledPairs(asset.Spot)
	assert.NoError(t, err)
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.Limit,
		Pairs:     currency.Pairs{enabledPairs[0], currency.NewPair(currency.USDT, currency.USD), currency.NewPair(currency.USD, currency.LTC)},
		AssetType: asset.Spot,
		Side:      order.Buy,
	}
	_, err = cr.GetActiveOrders(context.Background(), &getOrdersRequest)
	assert.NoError(t, err)
}

func TestGetOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.Buy,
	}
	_, err := cr.GetOrderHistory(context.Background(), &getOrdersRequest)
	assert.NoError(t, err)
	getOrdersRequest.Pairs = []currency.Pair{currency.NewPair(currency.LTC, currency.BTC)}
	_, err = cr.GetOrderHistory(context.Background(), &getOrdersRequest)
	assert.NoError(t, err)
}

func TestSubmitOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	var orderSubmission = &order.Submit{
		Pair: currency.Pair{
			Base:  currency.LTC,
			Quote: currency.BTC,
		},
		Exchange:  cr.Name,
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1000000000,
		ClientID:  "myOwnOrder",
		AssetType: asset.Spot,
	}
	_, err := cr.SubmitOrder(context.Background(), orderSubmission)
	assert.NoError(t, err)
	orderSubmission = &order.Submit{
		Pair: currency.Pair{
			Base:  currency.LTC,
			Quote: currency.BTC,
		},
		Exchange:  cr.Name,
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1000000000,
		ClientID:  "myOwnOrder",
		AssetType: asset.Spot,
	}
	_, err = cr.SubmitOrder(context.Background(), orderSubmission)
	assert.NoError(t, err)
}
func TestCancelOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	var orderCancellation = &order.Cancel{
		OrderID:   "1",
		Pair:      currency.NewPair(currency.LTC, currency.BTC),
		AssetType: asset.Spot,
	}
	err := cr.CancelOrder(context.Background(), orderCancellation)
	assert.NoError(t, err)
}

func TestCancelBatchOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	var orderCancellationParams = []order.Cancel{
		{
			OrderID: "1",
			Pair:    currency.NewPair(currency.LTC, currency.BTC),
		},
		{
			OrderID: "1",
			Pair:    currency.NewPair(currency.LTC, currency.BTC),
		},
	}
	_, err := cr.CancelBatchOrders(context.Background(), orderCancellationParams)
	assert.NoError(t, err)
}

func TestCancelAllOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	_, err := cr.CancelAllOrders(context.Background(), &order.Cancel{})
	assert.NoError(t, err)
}

func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	enabled, err := cr.GetEnabledPairs(asset.Spot)
	assert.NoError(t, err)
	if len(enabled) == 0 {
		t.SkipNow()
	}
	_, err = cr.GetOrderInfo(context.Background(),
		"123", enabled[0], asset.Spot)
	assert.NoError(t, err)
}

func TestGetDepositAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.GetDepositAddress(context.Background(), currency.ETH, "", "")
	assert.NoError(t, err)
}

func TestWithdrawCryptocurrencyFunds(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.WithdrawCryptocurrencyFunds(context.Background(), &withdraw.Request{
		Amount:   10,
		Currency: currency.BTC,
		Crypto: withdraw.CryptoRequest{
			Chain:      currency.BTC.String(),
			Address:    core.BitcoinDonationAddress,
			AddressTag: "",
		}})
	require.NoError(t, err)
}

func setupWS() {
	if !cr.Websocket.IsEnabled() {
		return
	}
	if !sharedtestvalues.AreAPICredentialsSet(cr) {
		cr.Websocket.SetCanUseAuthenticatedEndpoints(false)
	}
	err := cr.WsConnect()
	if err != nil {
		log.Fatal(err)
	}
}

func TestGenerateDefaultSubscriptions(t *testing.T) {
	t.Parallel()
	_, err := cr.GenerateDefaultSubscriptions()
	assert.NoError(t, err)
}

func TestWsRetriveCancelOnDisconnect(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr)
	_, err := cr.WsRetriveCancelOnDisconnect()
	assert.NoError(t, err)
}
func TestWsSetCancelOnDisconnect(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, cr, canManipulateRealOrders)
	_, err := cr.WsSetCancelOnDisconnect("ACCOUNT")
	assert.NoError(t, err)
}

func TestGetCreateParamMap(t *testing.T) {
	t.Parallel()
	arg := &CreateOrderParam{InstrumentName: "", OrderType: orderTypeToString(order.Limit), Price: 123, Quantity: 12}
	_, err := arg.getCreateParamMap()
	require.ErrorIs(t, err, errSymbolIsRequired)
	var newone *CreateOrderParam
	_, err = newone.getCreateParamMap()
	assert.ErrorIs(t, err, common.ErrNilPointer)
	arg.InstrumentName = "BTC_USDT"
	_, err = arg.getCreateParamMap()
	assert.ErrorIs(t, err, order.ErrSideIsInvalid)
	arg.Side = order.Buy
	_, err = arg.getCreateParamMap()
	assert.NoError(t, err)
	arg.OrderType = orderTypeToString(order.Market)
	_, err = arg.getCreateParamMap()
	assert.NoError(t, err)
	arg.OrderType = orderTypeToString(order.TakeProfit)
	arg.Notional = 12
	_, err = arg.getCreateParamMap()
	assert.ErrorIs(t, err, errTriggerPriceRequired)
	arg.OrderType = orderTypeToString(order.UnknownType)
	_, err = arg.getCreateParamMap()
	assert.ErrorIs(t, err, order.ErrTypeIsInvalid)
	arg.OrderType = orderTypeToString(order.StopLimit)
	_, err = arg.getCreateParamMap()
	assert.ErrorIs(t, err, errTriggerPriceRequired)
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	feeBuilder := &exchange.FeeBuilder{
		FeeType:       exchange.CryptocurrencyTradeFee,
		Pair:          currency.NewPair(currency.BTC, currency.USD),
		IsMaker:       true,
		Amount:        1,
		PurchasePrice: 1000,
	}
	_, err := cr.GetFeeByType(context.Background(), feeBuilder)
	require.NoError(t, err)
	if !sharedtestvalues.AreAPICredentialsSet(cr) {
		assert.Equal(t, exchange.OfflineTradeFee, feeBuilder.FeeType)
	} else {
		assert.Equal(t, exchange.CryptocurrencyTradeFee, feeBuilder.FeeType)
	}
}

var pushDataMap = map[string]string{
	"orderbookPushData":   `{ "id": -1, "code": 0, "method": "subscribe", "result": { "channel": "book", "subscription": "book.RSR_USDT", "instrument_name": "RSR_USDT", "depth": 150, "data": [ { "asks": [ [ "0.0041045", "164840", "1" ], [ "0.0041057", "273330", "1" ], [ "0.0041116", "6440", "1" ], [ "0.0041159", "29490", "1" ], [ "0.0041185", "21940", "1" ], [ "0.0041238", "191790", "2" ], [ "0.0041317", "495840", "2" ], [ "0.0041396", "1117990", "1" ], [ "0.0041475", "1430830", "1" ], [ "0.0041528", "785220", "1" ], [ "0.0041554", "1409330", "1" ], [ "0.0041633", "1710820", "1" ], [ "0.0041712", "2399680", "1" ], [ "0.0041791", "2355400", "1" ], [ "0.0042500", "1500", "1" ], [ "0.0044000", "1000", "1" ], [ "0.0045000", "1000", "1" ], [ "0.0046600", "85770", "1" ], [ "0.0049230", "20660", "1" ], [ "0.0049380", "88520", "2" ], [ "0.0050000", "1120", "1" ], [ "0.0050203", "304960", "2" ], [ "0.0051026", "509200", "2" ], [ "0.0051849", "3452290", "1" ], [ "0.0052672", "10928750", "1" ], [ "0.0206000", "730", "1" ], [ "0.0406000", "370", "1" ] ], "bids": [ [ "0.0041013", "273330", "1" ], [ "0.0040975", "3750", "1" ], [ "0.0040974", "174120", "1" ], [ "0.0040934", "6440", "1" ], [ "0.0040922", "32200", "1" ], [ "0.0040862", "21940", "1" ], [ "0.0040843", "187900", "2" ], [ "0.0040764", "483650", "3" ], [ "0.0040686", "12280", "1" ], [ "0.0040685", "813180", "3" ], [ "0.0040607", "16020", "1" ], [ "0.0040606", "1123210", "3" ], [ "0.0040527", "1432240", "3" ], [ "0.0040482", "642210", "1" ], [ "0.0040448", "1441580", "2" ], [ "0.0040369", "2071370", "2" ], [ "0.0040290", "1453600", "1" ], [ "0.0037500", "29390", "1" ], [ "0.0033776", "80", "1" ], [ "0.0033740", "29630", "1" ], [ "0.0033000", "50", "1" ], [ "0.0032797", "30990", "1" ], [ "0.0032097", "175720", "2" ], [ "0.0032000", "50", "1" ], [ "0.0031274", "511460", "2" ], [ "0.0031000", "50", "1" ], [ "0.0030451", "793150", "2" ], [ "0.0030400", "750000", "1" ], [ "0.0030000", "100", "1" ], [ "0.0029628", "5620050", "2" ], [ "0.0029000", "50", "1" ], [ "0.0028805", "20567780", "2" ], [ "0.0018000", "500", "1" ], [ "0.0014500", "500", "1" ] ], "t": 1679082891435, "tt": 1679082890266, "u": 27043535761920, "cs": 723295208 } ] } }`,
	"tickerPushData":      `{ "id": -1, "code": 0, "method": "subscribe", "result": { "channel": "ticker", "instrument_name": "RSR_USDT", "subscription": "ticker.RSR_USDT", "id": -1, "data": [ { "h": "0.0041622", "l": "0.0037959", "a": "0.0040738", "c": "0.0721", "b": "0.0040738", "bs": "3680", "k": "0.0040796", "ks": "179780", "i": "RSR_USDT", "v": "45133400", "vv": "181223.95", "oi": "0","t": 1679087156318}]}}`,
	"tradePushData":       `{"id": 140466243, "code": 0, "method": "subscribe", "result": { "channel": "trade", "subscription": "trade.RSR_USDT", "instrument_name": "RSR_USDT", "data": [ { "d": "4611686018428182866", "t": 1679085786004, "p": "0.0040604", "q": "10", "s": "BUY", "i": "RSR_USDT" }, { "d": "4611686018428182865", "t": 1679085717204, "p": "0.0040671", "q": "10", "s": "BUY", "i": "RSR_USDT" }, { "d": "4611686018428182864", "t": 1679085672504, "p": "0.0040664", "q": "10", "s": "BUY", "i": "RSR_USDT" }, { "d": "4611686018428182863", "t": 1679085638806, "p": "0.0040674", "q": "10", "s": "BUY", "i": "RSR_USDT" }, { "d": "4611686018428182862", "t": 1679085568762, "p": "0.0040689", "q": "20", "s": "BUY", "i": "RSR_USDT" } ] } }`,
	"candlestickPushData": `{"id": -1, "code": 0, "method": "subscribe", "result": { "channel": "candlestick", "instrument_name": "RSR_USDT", "subscription": "candlestick.5m.RSR_USDT", "interval": "5m", "data": [ { "o": "0.0040838", "h": "0.0040920", "l": "0.0040838", "c": "0.0040920", "v": "60.0000", "t": 1679087700000, "ut": 1679087959106 } ] } }`,
	"userBalancePushData": `{"id":3397447550047468012,"method":"subscribe","code":0,"result":{"subscription":"user.balance","channel":"user.balance","data":[{"stake":0,"balance":7.26648846,"available":7.26648846,"currency":"BOSON","order":0},{"stake":0,"balance":15.2782122,"available":15.2782122,"currency":"EFI","order":0},{"stake":0,"balance":90.63857968,"available":90.63857968,"currency":"ZIL","order":0},{"stake":0,"balance":16790279.87929312,"available":16790279.87929312,"currency":"SHIB","order":0},{"stake":0,"balance":1.79673318,"available":1.79673318,"currency":"NEAR","order":0},{"stake":0,"balance":307.29679422,"available":307.29679422,"currency":"DOGE","order":0},{"stake":0,"balance":0.00109125,"available":0.00109125,"currency":"BTC","order":0},{"stake":0,"balance":18634.17320776,"available":18634.17320776,"currency":"CRO-STAKE","order":0},{"stake":0,"balance":0.4312475,"available":0.4312475,"currency":"DOT","order":0},{"stake":0,"balance":924.07197632,"available":924.07197632,"currency":"CRO","order":0}]}}`,
	"userOrderPushData":   `{"method": "subscribe", "result": { "instrument_name": "ETH_CRO", "subscription": "user.order.ETH_CRO", "channel": "user.order", "data": [ { "status": "ACTIVE", "side": "BUY", "price": 1, "quantity": 1, "order_id": "366455245775097673", "client_oid": "my_order_0002", "create_time": 1588758017375, "update_time": 1588758017411, "type": "LIMIT", "instrument_name": "ETH_CRO", "cumulative_quantity": 0, "cumulative_value": 0, "avg_price": 0, "fee_currency": "CRO", "time_in_force":"GOOD_TILL_CANCEL" } ], "channel": "user.order.ETH_CRO" } }`,
	"userTradePushData":   `{"method": "subscribe", "code": 0, "result": { "instrument_name": "ETH_CRO", "subscription": "user.trade.ETH_CRO", "channel": "user.trade", "data": [ { "side": "SELL", "instrument_name": "ETH_CRO", "fee": 0.014, "trade_id": "367107655537806900", "create_time": "1588777459755", "traded_price": 7, "traded_quantity": 1, "fee_currency": "CRO", "order_id": "367107623521528450" } ], "channel": "user.trade.ETH_CRO" } }`,
}

const (
	getWithdrawalHistory    = `{ "withdrawal_list": [ { "currency": "XRP", "client_wid": "my_withdrawal_002", "fee": 1.0, "create_time": 1607063412000, "id": "2220", "update_time": 1607063460000, "amount": 100, "address": "2NBqqD5GRJ8wHy1PYyCXTe9ke5226FhavBf?1234567890", "status": "1", "txid": "", "network_id": null }]}`
	getPrivateTrades        = `{ "trade_list": [ { "side": "SELL", "instrument_name": "ETH_CRO", "fee": 0.014, "trade_id": "367107655537806900", "create_time": 1588777459755, "traded_price": 7, "traded_quantity": 1, "fee_currency": "CRO", "order_id": "367107623521528450"}]}`
	getPersonalOrderHistory = `{ "order_list": [ { "status": "FILLED", "side": "SELL", "price": 1, "quantity": 1, "order_id": "367107623521528457", "client_oid": "my_order_0002", "create_time": 1588777459755, "update_time": 1588777460700, "type": "LIMIT", "instrument_name": "ETH_CRO", "cumulative_quantity": 1, "cumulative_value": 1, "avg_price": 1, "fee_currency": "CRO", "time_in_force": "GOOD_TILL_CANCEL" }, { "status": "FILLED", "side": "SELL", "price": 1, "quantity": 1, "order_id": "367063282527104905", "client_oid": "my_order_0002", "create_time": 1588776138290, "update_time": 1588776138679, "type": "LIMIT", "instrument_name": "ETH_CRO", "cumulative_quantity": 1, "cumulative_value": 1, "avg_price": 1, "fee_currency": "CRO", "time_in_force": "GOOD_TILL_CANCEL"}]}`
	getTransactions         = `{ "data": [ { "account_id": "88888888-8888-8888-8888-000000000007", "event_date": "2021-02-18", "journal_type": "TRADING", "journal_id": "187078", "transaction_qty": "-0.0005", "transaction_cost": "-24.500000", "realized_pnl": "-0.006125", "order_id": "72062", "trade_id": "71497", "trade_match_id": "8625", "event_timestamp_ms": 1613640752166, "event_timestamp_ns": "1613640752166234567", "client_oid": "6ac2421d-5078-4ef6-a9d5-9680602ce123", "taker_side": "MAKER", "side": "SELL", "instrument_name": "BTCUSD-PERP" }, { "account_id": "9c72d8f1-583d-4b9d-b27c-55e695a2d116", "event_date": "2021-02-18", "journal_type": "SESSION_SETTLE", "journal_id": "186959", "transaction_qty": "0", "transaction_cost": "0.000000", "realized_pnl": "-0.007800", "trade_match_id": "0", "event_timestamp_ms": 1613638800001, "event_timestamp_ns": "1613638800001124563", "client_oid": "", "taker_side": "", "instrument_name": "BTCUSD-PERP" }]}`
	getOtcUser              = `{ "account_uuid": "00000000-00000000-00000000-00000000", "requests_per_minute": 30, "max_trade_value_usd": "5000000", "min_trade_value_usd": "50000", "accept_otc_tc_datetime": 1636512069509 }`
	requestOTCQuote         = `{"quote_id": "2412548678404715041", "quote_status": "ACTIVE", "quote_direction": "BUY", "base_currency": "BTC", "quote_currency": "USDT", "base_currency_size": null, "quote_currency_size": "100000.00", "quote_buy": "39708.24", "quote_buy_quantity": "2.51836898", "quote_buy_value": "100000.00", "quote_sell": "39677.18", "quote_sell_quantity": "2.52034040", "quote_sell_value": "100000.00", "quote_duration": 2, "quote_time": 1649736353489, "quote_expiry_time": 1649736363578 }`
	acceptOTCQuote          = `{"quote_id": "2412548678404715041", "quote_status": "FILLED", "quote_direction": "BUY", "base_currency": "BTC", "quote_currency": "USDT", "base_currency_size": null, "quote_currency_size": "100000.00", "quote_buy": "39708.24", "quote_sell": null, "quote_duration": 2, "quote_time": 1649743710146, "quote_expiry_time": 1649743720231, "trade_direction": "BUY", "trade_price": "39708.24", "trade_quantity": "2.51836898", "trade_value": "100000.00", "trade_time": 1649743718963 }`
	getOTCQuoteHistory      = `{"count": 1, "quote_list": [ { "quote_id": "2412795526826582752", "quote_status": "EXPIRED", "quote_direction": "BUY", "base_currency": "BTC", "quote_currency": "USDT", "base_currency_size": null, "quote_currency_size": "100000.00", "quote_buy": "39708.24", "quote_sell": null, "quote_duration": 2, "quote_time": 1649743710146, "quote_expiry_time": 1649743720231, "trade_direction": null, "trade_price": null, "trade_quantity": null, "trade_value": null, "trade_time": null } ] }`
	getOTCTradeHistory      = `{"count": 1, "trade_list": [ { "quote_id": "2412795526826582752", "quote_status": "FILLED", "quote_direction": "BUY", "base_currency": "BTC", "quote_currency": "USDT", "base_currency_size": null, "quote_currency_size": "100000.00", "quote_buy": "39708.24", "quote_sell": null, "quote_duration": 10, "quote_time": 1649743710146, "quote_expiry_time": 1649743720231, "trade_direction": "BUY", "trade_price": "39708.24", "trade_quantity": "2.51836898", "trade_value": "100000.00", "trade_time": 1649743718963 } ] }`
)

func TestPushData(t *testing.T) {
	t.Parallel()
	for x := range pushDataMap {
		err := cr.WsHandleData([]byte(pushDataMap[x]), true)
		require.NoError(t, err)
	}
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()
	err := cr.UpdateOrderExecutionLimits(context.Background(), asset.Binary)
	require.ErrorIs(t, err, asset.ErrNotSupported)
	err = cr.UpdateOrderExecutionLimits(context.Background(), asset.Spot)
	require.NoError(t, err)

	pairs, err := cr.FetchTradablePairs(context.Background(), asset.Spot)
	require.NoError(t, err)

	for y := range pairs {
		lim, err := cr.GetOrderExecutionLimits(asset.Spot, pairs[y])
		require.NoErrorf(t, err, "%v %s %v", err, pairs[y], asset.Spot)
		assert.NotEmpty(t, lim, "limit cannot be empty")
	}
}

func TestSafeNumberUnmarshal(t *testing.T) {
	t.Parallel()
	result := []byte(`{"value": null}`)
	resp := &struct {
		Value SafeNumber `json:"value"`
	}{}
	err := json.Unmarshal(result, &resp)
	require.NoError(t, err)
}
