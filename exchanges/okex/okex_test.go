package okex

import (
	"fmt"
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/okgroup"
)

// Please supply you own test keys here for due diligence testing.
const (
	apiKey                  = ""
	apiSecret               = ""
	passphrase              = ""
	OKGroupExchange         = "OKEX"
	canManipulateRealOrders = false
)

var o = OKEX{}

func TestSetDefaults(t *testing.T) {
	if o.Name != OKGroupExchange {
		o.SetDefaults()
	}
	if o.GetName() != OKGroupExchange {
		t.Errorf("Test Failed - %v - SetDefaults() error", OKGroupExchange)
	}
	t.Parallel()
	TestSetup(t)
}

func TestSetRealOrderDefaults(t *testing.T) {
	TestSetDefaults(t)
	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}
}

func TestSetup(t *testing.T) {
	if o.APIKey == apiKey && o.APISecret == apiSecret &&
		o.ClientID == passphrase {
		return
	}
	o.ExchangeName = OKGroupExchange
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")

	okexConfig, err := cfg.GetExchangeConfig(OKGroupExchange)
	if err != nil {
		t.Errorf("Test Failed - %v Setup() init error", OKGroupExchange)
	}

	okexConfig.AuthenticatedAPISupport = true
	okexConfig.APIKey = apiKey
	okexConfig.APISecret = apiSecret
	okexConfig.ClientID = passphrase
	okexConfig.Verbose = true
	okexConfig.WebsocketURL = o.WebsocketURL
	o.Setup(okexConfig)
}

func areTestAPIKeysSet() bool {
	if o.APIKey != "" && o.APIKey != "Key" &&
		o.APISecret != "" && o.APISecret != "Secret" {
		return true
	}
	return false
}

func testStandardErrorHandling(t *testing.T, err error) {
	if !areTestAPIKeysSet() && err == nil {
		t.Fatal("Expecting an error when no keys are set")
	}
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Encountered error: %v", err)
	}
}

// setupWSConnection Connect to WS, but pass back error so test can handle it if needed
func setupWSConnection() error {
	o.Enabled = true
	err := o.WebsocketSetup(o.WsConnect,
		o.Name,
		true,
		o.WebsocketURL,
		o.WebsocketURL)
	o.Websocket.DataHandler = make(chan interface{}, 500)
	if err != nil {
		return err
	}
	o.Websocket.SetWsStatusAndConnection(true)
	return nil
}

func connectToWs() error {
	err := o.Websocket.Connect()
	if err != nil {
		return err
	}
	return nil
}

// disconnectFromWS disconnect to WS, but pass back error so test can handle it if needed
func disconnectFromWS() error {
	err := o.Websocket.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

// TestGetAccountCurrencies API endpoint test
func TestGetAccountCurrencies(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountCurrencies()
	testStandardErrorHandling(t, err)
}

// TestGetAccountWalletInformation API endpoint test
func TestGetAccountWalletInformation(t *testing.T) {
	TestSetDefaults(t)
	resp, err := o.GetAccountWalletInformation("")
	testStandardErrorHandling(t, err)

	if areTestAPIKeysSet() && len(resp) == 0 {
		t.Error("No wallets returned")
	}
}

// TestGetAccountWalletInformationForCurrency API endpoint test
func TestGetAccountWalletInformationForCurrency(t *testing.T) {
	TestSetDefaults(t)
	resp, err := o.GetAccountWalletInformation(symbol.BTC)
	testStandardErrorHandling(t, err)

	if areTestAPIKeysSet() && len(resp) != 1 {
		t.Errorf("Error receiving wallet information for currency: %v", symbol.BTC)
	}
}

// TestTransferAccountFunds API endpoint test
func TestTransferAccountFunds(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.TransferAccountFundsRequest{
		Amount:   10,
		Currency: symbol.BTC,
		From:     6,
		To:       1,
	}

	_, err := o.TransferAccountFunds(request)
	testStandardErrorHandling(t, err)
}

// TestBaseWithdraw API endpoint test
func TestAccountWithdrawRequest(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.AccountWithdrawRequest{
		Amount:      10,
		Currency:    symbol.BTC,
		TradePwd:    "1234",
		Destination: 4,
		ToAddress:   "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		Fee:         1,
	}

	_, err := o.AccountWithdraw(request)
	testStandardErrorHandling(t, err)
}

// TestGetAccountWithdrawalFee API endpoint test
func TestGetAccountWithdrawalFee(t *testing.T) {
	TestSetDefaults(t)
	resp, err := o.GetAccountWithdrawalFee("")
	testStandardErrorHandling(t, err)

	if areTestAPIKeysSet() && len(resp) == 0 {
		t.Error("Expected fees")
	}
}

// TestGetWithdrawalFeeForCurrency API endpoint test
func TestGetAccountWithdrawalFeeForCurrency(t *testing.T) {
	TestSetDefaults(t)
	resp, err := o.GetAccountWithdrawalFee(symbol.BTC)
	testStandardErrorHandling(t, err)

	if areTestAPIKeysSet() && len(resp) != 1 {
		t.Error("Expected fee for one currency")
	}
}

// TestGetAccountWithdrawalHistory API endpoint test
func TestGetAccountWithdrawalHistory(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountWithdrawalHistory("")
	testStandardErrorHandling(t, err)
}

// TestGetAccountWithdrawalHistoryForCurrency API endpoint test
func TestGetAccountWithdrawalHistoryForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountWithdrawalHistory(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetAccountBillDetails API endpoint test
func TestGetAccountBillDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountBillDetails(okgroup.GetAccountBillDetailsRequest{})
	testStandardErrorHandling(t, err)
}

// TestGetAccountDepositAddressForCurrency API endpoint test
func TestGetAccountDepositAddressForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountDepositAddressForCurrency(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetAccountDepositHistory API endpoint test
func TestGetAccountDepositHistory(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountDepositHistory("")
	testStandardErrorHandling(t, err)
}

// TestGetAccountDepositHistoryForCurrency API endpoint test
func TestGetAccountDepositHistoryForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAccountDepositHistory(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetSpotTradingAccounts API endpoint test
func TestGetSpotTradingAccounts(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSpotTradingAccounts()
	testStandardErrorHandling(t, err)
}

// TestGetSpotTradingAccountsForCurrency API endpoint test
func TestGetSpotTradingAccountsForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSpotTradingAccountForCurrency(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetSpotBillDetailsForCurrency API endpoint test
func TestGetSpotBillDetailsForCurrency(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotBillDetailsForCurrencyRequest{
		Currency: symbol.BTC,
		Limit:    100,
	}

	_, err := o.GetSpotBillDetailsForCurrency(request)
	testStandardErrorHandling(t, err)

	request.Limit = -1
	_, err = o.GetSpotBillDetailsForCurrency(request)
	if areTestAPIKeysSet() && err == nil {
		t.Errorf("Expecting an error when invalid request sent")
	}

}

// TestPlaceSpotOrderLimit API endpoint test
func TestPlaceSpotOrderLimit(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "limit",
		Side:          "buy",
		MarginTrading: "1",
		Price:         "100",
		Size:          "100",
	}

	_, err := o.PlaceSpotOrder(request)
	testStandardErrorHandling(t, err)
}

// TestPlaceSpotOrderMarket API endpoint test
func TestPlaceSpotOrderMarket(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	_, err := o.PlaceSpotOrder(request)
	testStandardErrorHandling(t, err)
}

// TestPlaceMultipleSpotOrders API endpoint test
func TestPlaceMultipleSpotOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
	}

	_, errs := o.PlaceMultipleSpotOrders(request)
	if len(errs) > 0 {
		testStandardErrorHandling(t, errs[0])
	}
}

// TestPlaceMultipleSpotOrdersOverCurrencyLimits API logic test
func TestPlaceMultipleSpotOrdersOverCurrencyLimits(t *testing.T) {
	TestSetDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
		order,
		order,
		order,
		order,
	}

	_, errs := o.PlaceMultipleSpotOrders(request)
	if errs[0].Error() != "maximum 4 orders for each pair" {
		t.Error("Expecting an error when more than 4 orders for a pair supplied", errs[0])
	}
}

// TestPlaceMultipleSpotOrdersOverPairLimits API logic test
func TestPlaceMultipleSpotOrdersOverPairLimits(t *testing.T) {
	TestSetDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
	}

	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.LTC, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.DOGE, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.XMR, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.BCH, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)

	_, errs := o.PlaceMultipleSpotOrders(request)
	if errs[0].Error() != "up to 4 trading pairs" {
		t.Error("Expecting an error when more than 4 trading pairs supplied", errs[0])
	}
}

// TestCancelSpotOrder API endpoint test
func TestCancelSpotOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelSpotOrderRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderID:      1234,
	}

	_, err := o.CancelSpotOrder(request)
	testStandardErrorHandling(t, err)
}

// TestCancelMultipleSpotOrders API endpoint test
func TestCancelMultipleSpotOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelMultipleSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderIDs:     []int64{1, 2, 3, 4},
	}

	cancellations, err := o.CancelMultipleSpotOrders(request)
	testStandardErrorHandling(t, err)
	for _, cancellationsPerCurrency := range cancellations {
		for _, cancellation := range cancellationsPerCurrency {
			if !cancellation.Result {
				t.Error(cancellation.Error)
			}
		}
	}
}

// TestCancelMultipleSpotOrdersOverCurrencyLimits API logic test
func TestCancelMultipleSpotOrdersOverCurrencyLimits(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelMultipleSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderIDs:     []int64{1, 2, 3, 4, 5},
	}

	_, err := o.CancelMultipleSpotOrders(request)
	if err.Error() != "maximum 4 order cancellations for each pair" {
		t.Error("Expecting an error when more than 4 orders for a pair supplied", err)
	}
}

// TestGetSpotOrders API endpoint test
func TestGetSpotOrders(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Status:       "all",
		Limit:        1,
	}
	_, err := o.GetSpotOrders(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotOpenOrders API endpoint test
func TestGetSpotOpenOrders(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOpenOrdersRequest{}
	_, err := o.GetSpotOpenOrders(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotOrder API endpoint test
func TestGetSpotOrder(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOrderRequest{
		OrderID:      -1234,
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Upper().String(),
	}
	_, err := o.GetSpotOrder(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotTransactionDetails API endpoint test
func TestGetSpotTransactionDetails(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotTransactionDetailsRequest{
		OrderID:      1234,
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
	}
	_, err := o.GetSpotTransactionDetails(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotTokenPairDetails API endpoint test
func TestGetSpotTokenPairDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSpotTokenPairDetails()
	testStandardErrorHandling(t, err)
}

// TestGetSpotOrderBook API endpoint test
func TestGetSpotOrderBook(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOrderBookRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
	}
	_, err := o.GetSpotOrderBook(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotAllTokenPairsInformation API endpoint test
func TestGetSpotAllTokenPairsInformation(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSpotAllTokenPairsInformation()
	testStandardErrorHandling(t, err)
}

// TestGetSpotAllTokenPairsInformationForCurrency API endpoint test
func TestGetSpotAllTokenPairsInformationForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSpotAllTokenPairsInformationForCurrency(pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String())
	testStandardErrorHandling(t, err)
}

// TestGetSpotFilledOrdersInformation API endpoint test
func TestGetSpotFilledOrdersInformation(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotFilledOrdersInformationRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
	}
	_, err := o.GetSpotFilledOrdersInformation(request)
	testStandardErrorHandling(t, err)
}

// TestGetSpotMarketData API endpoint test
func TestGetSpotMarketData(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotMarketDataRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Granularity:  604800,
	}
	_, err := o.GetSpotMarketData(request)
	testStandardErrorHandling(t, err)
}

// TestGetMarginTradingAccounts API endpoint test
func TestGetMarginTradingAccounts(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetMarginTradingAccounts()
	testStandardErrorHandling(t, err)
}

// TestGetMarginTradingAccountsForCurrency API endpoint test
func TestGetMarginTradingAccountsForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetMarginTradingAccountsForCurrency(pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String())
	testStandardErrorHandling(t, err)
}

// TestGetMarginBillDetails API endpoint test
func TestGetMarginBillDetails(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetMarginBillDetailsRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Limit:        100,
	}

	_, err := o.GetMarginBillDetails(request)
	testStandardErrorHandling(t, err)
}

// TestGetMarginAccountSettings API endpoint test
func TestGetMarginAccountSettings(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetMarginAccountSettings("")
	testStandardErrorHandling(t, err)
}

// TestGetMarginAccountSettingsForCurrency API endpoint test
func TestGetMarginAccountSettingsForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetMarginAccountSettings(pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String())
	testStandardErrorHandling(t, err)
}

// TestOpenMarginLoan API endpoint test
func TestOpenMarginLoan(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.OpenMarginLoanRequest{
		Amount:        100,
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		QuoteCurrency: symbol.USDT,
	}

	_, err := o.OpenMarginLoan(request)
	testStandardErrorHandling(t, err)
}

// TestRepayMarginLoan API endpoint test
func TestRepayMarginLoan(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.RepayMarginLoanRequest{
		Amount:        100,
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		QuoteCurrency: symbol.USDT,
		BorrowID:      1,
	}

	_, err := o.RepayMarginLoan(request)
	testStandardErrorHandling(t, err)
}

// TestPlaceMarginOrderLimit API endpoint test
func TestPlaceMarginOrderLimit(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "limit",
		Side:          "buy",
		MarginTrading: "2",
		Price:         "100",
		Size:          "100",
	}

	_, err := o.PlaceMarginOrder(request)
	testStandardErrorHandling(t, err)
}

// TestPlaceMarginOrderMarket API endpoint test
func TestPlaceMarginOrderMarket(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "2",
		Size:          "100",
		Notional:      "100",
	}

	_, err := o.PlaceMarginOrder(request)
	testStandardErrorHandling(t, err)
}

// TestPlaceMultipleMarginOrders API endpoint test
func TestPlaceMultipleMarginOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
	}

	_, errs := o.PlaceMultipleMarginOrders(request)
	if len(errs) > 0 {
		testStandardErrorHandling(t, errs[0])
	}
}

// TestPlaceMultipleMarginOrdersOverCurrencyLimits API logic test

func TestPlaceMultipleMarginOrdersOverCurrencyLimits(t *testing.T) {
	TestSetDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
		order,
		order,
		order,
		order,
	}

	_, errs := o.PlaceMultipleMarginOrders(request)
	if errs[0].Error() != "maximum 4 orders for each pair" {
		t.Error("Expecting an error when more than 4 orders for a pair supplied", errs[0])
	}
}

// TestPlaceMultipleMarginOrdersOverPairLimits API logic test
func TestPlaceMultipleMarginOrdersOverPairLimits(t *testing.T) {
	TestSetDefaults(t)
	order := okgroup.PlaceSpotOrderRequest{
		InstrumentID:  pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          "market",
		Side:          "buy",
		MarginTrading: "1",
		Size:          "100",
		Notional:      "100",
	}

	request := []okgroup.PlaceSpotOrderRequest{
		order,
	}

	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.LTC, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.DOGE, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.XMR, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)
	order.InstrumentID = pair.NewCurrencyPairWithDelimiter(symbol.BCH, symbol.USDT, "-").Pair().Lower().String()
	request = append(request, order)

	_, errs := o.PlaceMultipleMarginOrders(request)
	if errs[0].Error() != "up to 4 trading pairs" {
		t.Error("Expecting an error when more than 4 trading pairs supplied", errs[0])
	}
}

// TestCancelMarginOrder API endpoint test
func TestCancelMarginOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelSpotOrderRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderID:      1234,
	}

	_, err := o.CancelMarginOrder(request)
	testStandardErrorHandling(t, err)
}

// TestCancelMultipleMarginOrders API endpoint test
func TestCancelMultipleMarginOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelMultipleSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderIDs:     []int64{1, 2, 3, 4},
	}

	_, errs := o.CancelMultipleMarginOrders(request)
	if len(errs) > 0 {
		testStandardErrorHandling(t, errs[0])
	}
}

// TestCancelMultipleMarginOrdersOverCurrencyLimits API logic test
func TestCancelMultipleMarginOrdersOverCurrencyLimits(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelMultipleSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		OrderIDs:     []int64{1, 2, 3, 4, 5},
	}

	_, errs := o.CancelMultipleMarginOrders(request)
	if errs[0].Error() != "maximum 4 order cancellations for each pair" {
		t.Error("Expecting an error when more than 4 orders for a pair supplied", errs[0])
	}
}

// TestGetMarginOrders API endpoint test
func TestGetMarginOrders(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOrdersRequest{
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Status:       "all",
	}
	_, err := o.GetMarginOrders(request)
	testStandardErrorHandling(t, err)
}

// TestGetMarginOpenOrders API endpoint test
func TestGetMarginOpenOrders(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOpenOrdersRequest{}
	_, err := o.GetMarginOpenOrders(request)
	testStandardErrorHandling(t, err)
}

// TestGetMarginOrder API endpoint test
func TestGetMarginOrder(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotOrderRequest{
		OrderID:      1234,
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Upper().String(),
	}
	_, err := o.GetMarginOrder(request)
	testStandardErrorHandling(t, err)
}

// TestGetMarginTransactionDetails API endpoint test
func TestGetMarginTransactionDetails(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSpotTransactionDetailsRequest{
		OrderID:      1234,
		InstrumentID: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
	}
	_, err := o.GetMarginTransactionDetails(request)
	testStandardErrorHandling(t, err)
}

var genericFutureInstrumentID string

// getFutureInstrumentID Future contract ids are date based without an easy way to calculate the closest valid date
// This retrieves the value and stores it if running all tests so only one call is made
func getFutureInstrumentID() string {
	if genericFutureInstrumentID != "" {
		return genericFutureInstrumentID
	}
	resp, err := o.GetFuturesContractInformation()
	if err != nil {
		// No error handling here because we're not testing this
		return err.Error()
	}
	genericFutureInstrumentID = resp[0].InstrumentID
	return genericFutureInstrumentID
}

// TestGetFuturesPostions API endpoint test
func TestGetFuturesPostions(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesPostions()
	testStandardErrorHandling(t, err)
}

// TestGetFuturesPostionsForCurrency API endpoint test
func TestGetFuturesPostionsForCurrency(t *testing.T) {
	TestSetDefaults(t)
	currencyContract := getFutureInstrumentID()
	_, err := o.GetFuturesPostionsForCurrency(currencyContract)
	testStandardErrorHandling(t, err)
}

// TestGetFuturesAccountOfAllCurrencies API endpoint test
func TestGetFuturesAccountOfAllCurrencies(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesAccountOfAllCurrencies()
	testStandardErrorHandling(t, err)
}

// TestGetFuturesAccountOfACurrency API endpoint test
func TestGetFuturesAccountOfACurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesAccountOfACurrency(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetFuturesLeverage API endpoint test
func TestGetFuturesLeverage(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesLeverage(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestSetFuturesLeverage API endpoint test
func TestSetFuturesLeverage(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.SetFuturesLeverageRequest{
		Currency:     symbol.BTC,
		InstrumentID: getFutureInstrumentID(),
		Leverage:     10,
		Direction:    "Long",
	}
	_, err := o.SetFuturesLeverage(request)
	testStandardErrorHandling(t, err)
}

// TestGetFuturesBillDetails API endpoint test
func TestGetFuturesBillDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesBillDetails(okgroup.GetSpotBillDetailsForCurrencyRequest{
		Currency: symbol.BTC,
	})
	testStandardErrorHandling(t, err)
}

// TestPlaceFuturesOrder API endpoint test
func TestPlaceFuturesOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.PlaceFuturesOrder(okgroup.PlaceFuturesOrderRequest{
		InstrumentID: getFutureInstrumentID(),
		Leverage:     10,
		Type:         1,
		Size:         2,
		Price:        432.11,
		ClientOid:    "12233456",
	})
	testStandardErrorHandling(t, err)
}

// TestPlaceFuturesOrderBatch API endpoint test
func TestPlaceFuturesOrderBatch(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.PlaceFuturesOrderBatch(okgroup.PlaceFuturesOrderBatchRequest{
		InstrumentID: getFutureInstrumentID(),
		Leverage:     10,
		OrdersData: []okgroup.PlaceFuturesOrderBatchRequestDetails{
			okgroup.PlaceFuturesOrderBatchRequestDetails{
				ClientOid:  "1",
				MatchPrice: "0",
				Price:      "100",
				Size:       "100",
				Type:       "1",
			},
		},
	})
	testStandardErrorHandling(t, err)
}

// TestCancelFuturesOrder API endpoint test
func TestCancelFuturesOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.CancelFuturesOrder(okgroup.CancelFuturesOrderRequest{
		InstrumentID: getFutureInstrumentID(),
		OrderID:      "1",
	})
	testStandardErrorHandling(t, err)
}

// TestCancelMultipleSpotOrders API endpoint test
func TestCancelMultipleFuturesOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.CancelMultipleSpotOrdersRequest{
		InstrumentID: getFutureInstrumentID(),
		OrderIDs:     []int64{1, 2, 3, 4},
	}

	_, err := o.CancelFuturesOrderBatch(request)
	testStandardErrorHandling(t, err)
}

// TestGetFuturesOrderList API endpoint test
func TestGetFuturesOrderList(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesOrderList(okgroup.GetFuturesOrdersListRequest{
		InstrumentID: getFutureInstrumentID(),
		Status:       6,
	})
	testStandardErrorHandling(t, err)
}

// TestGetFuturesOrderDetails API endpoint test
func TestGetFuturesOrderDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesOrderDetails(okgroup.GetFuturesOrderDetailsRequest{
		InstrumentID: getFutureInstrumentID(),
		OrderID:      1,
	})
	testStandardErrorHandling(t, err)
}

// TestGetFuturesTransactionDetails API endpoint test
func TestGetFuturesTransactionDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesTransactionDetails(okgroup.GetFuturesTransactionDetailsRequest{
		InstrumentID: getFutureInstrumentID(),
		OrderID:      1,
	})
	testStandardErrorHandling(t, err)
}

// TestGetFuturesContractInformation API endpoint test
func TestGetFuturesContractInformation(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesContractInformation()
	if err != nil {
		t.Error(err)
	}
}

// TestGetFuturesContractInformation API endpoint test
func TestGetFuturesOrderBook(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesOrderBook(okgroup.GetFuturesOrderBookRequest{
		InstrumentID: getFutureInstrumentID(),
		Size:         10,
	})
	testStandardErrorHandling(t, err)
}

// TestGetAllFuturesTokenInfo API endpoint test
func TestGetAllFuturesTokenInfo(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAllFuturesTokenInfo()
	testStandardErrorHandling(t, err)
}

// TestGetAllFuturesTokenInfo API endpoint test
func TestGetFuturesTokenInfoForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesTokenInfoForCurrency(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesFilledOrder API endpoint test
func TestGetFuturesFilledOrder(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesFilledOrder(okgroup.GetFuturesFilledOrderRequest{
		InstrumentID: getFutureInstrumentID(),
	})
	testStandardErrorHandling(t, err)
}

// TestGetFuturesHoldAmount API endpoint test
func TestGetFuturesHoldAmount(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesHoldAmount(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesHoldAmount API endpoint test
func TestGetFuturesIndices(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesIndices(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesHoldAmount API endpoint test
func TestGetFuturesExchangeRates(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesExchangeRates()
	if err != nil {
		t.Errorf("Encountered error: %v", err)
	}
}

// TestGetFuturesHoldAmount API endpoint test
func TestGetFuturesEstimatedDeliveryPrice(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesEstimatedDeliveryPrice(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesOpenInterests API endpoint test
func TestGetFuturesOpenInterests(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesOpenInterests(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesOpenInterests API endpoint test
func TestGetFuturesCurrentPriceLimit(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesCurrentPriceLimit(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesCurrentMarkPrice API endpoint test
func TestGetFuturesCurrentMarkPrice(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesCurrentMarkPrice(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetFuturesForceLiquidatedOrders API endpoint test
func TestGetFuturesForceLiquidatedOrders(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesForceLiquidatedOrders(okgroup.GetFuturesForceLiquidatedOrdersRequest{
		InstrumentID: getFutureInstrumentID(),
		Status:       "1",
	})
	testStandardErrorHandling(t, err)
}

// TestGetFuturesTagPrice API endpoint test
func TestGetFuturesTagPrice(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetFuturesTagPrice(getFutureInstrumentID())
	testStandardErrorHandling(t, err)
}

// TestGetSwapPostions API endpoint test
func TestGetSwapPostions(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapPostions()
	testStandardErrorHandling(t, err)
}

func TestGetSwapPostionsForContract(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapPostionsForContract(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapAccountOfAllCurrency API endpoint test
func TestGetSwapAccountOfAllCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapAccountOfAllCurrency()
	testStandardErrorHandling(t, err)
}

// TestGetSwapAccountSettingsOfAContract API endpoint test
func TestGetSwapAccountSettingsOfAContract(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapAccountSettingsOfAContract(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestSetSwapLeverageLevelOfAContract API endpoint test
func TestSetSwapLeverageLevelOfAContract(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.SetSwapLeverageLevelOfAContract(okgroup.SetSwapLeverageLevelOfAContractRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Leverage:     10,
		Side:         1,
	})

	testStandardErrorHandling(t, err)
}

// TestGetSwapAccountSettingsOfAContract API endpoint test
func TestGetSwapBillDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapBillDetails(okgroup.GetSpotBillDetailsForCurrencyRequest{
		Currency: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Limit:    100,
	})
	testStandardErrorHandling(t, err)
}

// TestPlaceSwapOrder API endpoint test
func TestPlaceSwapOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.PlaceSwapOrder(okgroup.PlaceSwapOrderRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Size:         1,
		Type:         1,
		Price:        1,
	})
	testStandardErrorHandling(t, err)
}

// TestPlaceMultipleSwapOrders API endpoint test
func TestPlaceMultipleSwapOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.PlaceMultipleSwapOrders(okgroup.PlaceMultipleSwapOrdersRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Leverage:     10,
		OrdersData: []okgroup.PlaceMultipleSwapOrderData{
			okgroup.PlaceMultipleSwapOrderData{
				ClientOID:  "hello",
				MatchPrice: "0",
				Price:      "10",
				Size:       "1",
				Type:       "1",
			}, okgroup.PlaceMultipleSwapOrderData{
				ClientOID:  "hello2",
				MatchPrice: "0",
				Price:      "10",
				Size:       "1",
				Type:       "1",
			}},
	})
	testStandardErrorHandling(t, err)
}

// TestCancelSwapOrder API endpoint test
func TestCancelSwapOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.CancelSwapOrder(okgroup.CancelSwapOrderRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		OrderID:      "64-2a-26132f931-3",
	})
	testStandardErrorHandling(t, err)
}

// TestCancelMultipleSwapOrders API endpoint test
func TestCancelMultipleSwapOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.CancelMultipleSwapOrders(okgroup.CancelMultipleSwapOrdersRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		OrderIDs:     []int64{1, 2, 3, 4},
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapOrderList API endpoint test
func TestGetSwapOrderList(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapOrderList(okgroup.GetSwapOrderListRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Status:       6,
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapOrderDetails API endpoint test
func TestGetSwapOrderDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapOrderDetails(okgroup.GetSwapOrderDetailsRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		OrderID:      "64-2a-26132f931-3",
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapTransactionDetails API endpoint test
func TestGetSwapTransactionDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapTransactionDetails(okgroup.GetSwapTransactionDetailsRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		OrderID:      "64-2a-26132f931-3",
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapContractInformation API endpoint test
func TestGetSwapContractInformation(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapContractInformation()
	testStandardErrorHandling(t, err)
}

// TestGetSwapOrderBook API endpoint test
func TestGetSwapOrderBook(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapOrderBook(okgroup.GetSwapOrderBookRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Size:         200,
	})

	testStandardErrorHandling(t, err)
}

// TestGetAllSwapTokensInformation API endpoint test
func TestGetAllSwapTokensInformation(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetAllSwapTokensInformation()
	testStandardErrorHandling(t, err)
}

// TestGetSwapTokensInformationForCurrency API endpoint test
func TestGetSwapTokensInformationForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapTokensInformationForCurrency(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapFilledOrdersData API endpoint test
func TestGetSwapFilledOrdersData(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapFilledOrdersData(&okgroup.GetSwapFilledOrdersDataRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Limit:        100,
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapMarketData API endpoint test
func TestGetSwapMarketData(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetSwapMarketDataRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Granularity:  604800,
	}
	_, err := o.GetSwapMarketData(request)
	testStandardErrorHandling(t, err)
}

// TestGetSwapIndeces API endpoint test
func TestGetSwapIndeces(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapIndeces(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapExchangeRates API endpoint test
func TestGetSwapExchangeRates(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapExchangeRates()
	testStandardErrorHandling(t, err)
}

// TestGetSwapOpenInterest API endpoint test
func TestGetSwapOpenInterest(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapOpenInterest(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapCurrentPriceLimits API endpoint test
func TestGetSwapCurrentPriceLimits(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapCurrentPriceLimits(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapForceLiquidatedOrders API endpoint test
func TestGetSwapForceLiquidatedOrders(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapForceLiquidatedOrders(okgroup.GetSwapForceLiquidatedOrdersRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Status:       "0",
	})
	testStandardErrorHandling(t, err)
}

// TestGetSwapOnHoldAmountForOpenOrders API endpoint test
func TestGetSwapOnHoldAmountForOpenOrders(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapOnHoldAmountForOpenOrders(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapNextSettlementTime API endpoint test
func TestGetSwapNextSettlementTime(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapNextSettlementTime(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapMarkPrice API endpoint test
func TestGetSwapMarkPrice(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapMarkPrice(fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD))
	testStandardErrorHandling(t, err)
}

// TestGetSwapFundingRateHistory API endpoint test
func TestGetSwapFundingRateHistory(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetSwapFundingRateHistory(okgroup.GetSwapFundingRateHistoryRequest{
		InstrumentID: fmt.Sprintf("%v-%v-SWAP", symbol.BTC, symbol.USD),
		Limit:        100,
	})
	testStandardErrorHandling(t, err)
}

// TestGetETT API endpoint test
func TestGetETT(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETT()
	testStandardErrorHandling(t, err)
}

// TestGetETTAccountInformationForCurrency API endpoint test
func TestGetETTAccountInformationForCurrency(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETTBillsDetails(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestGetETTBillsDetails API endpoint test
func TestGetETTBillsDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETTBillsDetails(symbol.BTC)
	testStandardErrorHandling(t, err)
}

// TestPlaceETTOrder API endpoint test
func TestPlaceETTOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	request := okgroup.PlaceETTOrderRequest{
		QuoteCurrency: pair.NewCurrencyPairWithDelimiter(symbol.BTC, symbol.USDT, "-").Pair().Lower().String(),
		Type:          0,
		Size:          "100",
		Amount:        1,
		ETT:           "OK06",
	}

	_, err := o.PlaceETTOrder(request)
	testStandardErrorHandling(t, err)
}

// TestCancelETTOrder API endpoint test
func TestCancelETTOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.CancelETTOrder("888845120785408")
	testStandardErrorHandling(t, err)
}

// TestGetETTOrderList API endpoint test
// This results in a 500 error when its a request object
// Or when it is submitted as URL params
// Unsure how to fix
func TestGetETTOrderList(t *testing.T) {
	TestSetDefaults(t)
	request := okgroup.GetETTOrderListRequest{
		Type:   1,
		ETT:    "OK06ETT",
		Status: 0,
	}

	_, err := o.GetETTOrderList(request)
	testStandardErrorHandling(t, err)
}

// TestGetETTOrderDetails API endpoint test
func TestGetETTOrderDetails(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETTOrderDetails("888845020785408")
	testStandardErrorHandling(t, err)
}

// TestGetETTConstituents API endpoint test
func TestGetETTConstituents(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETTConstituents("OK06ETT")
	testStandardErrorHandling(t, err)
}

// TestGetETTSettlementPriceHistory API endpoint test
func TestGetETTSettlementPriceHistory(t *testing.T) {
	TestSetDefaults(t)
	_, err := o.GetETTSettlementPriceHistory("OK06ETT")
	testStandardErrorHandling(t, err)
}

// Websocket tests ----------------------------------------------------------------------------------------------

// TestSubscribeToPewDiePie API endpoint test
func TestWSSetup(t *testing.T) {
	defer disconnectFromWS()
	TestSetDefaults(t)
	err := setupWSConnection()
	if err != nil {
		t.Error(err)
	}
}

// TestSubscribeToPewDiePie API endpoint test
func TestSubscribeToPewDiePie(t *testing.T) {
	defer disconnectFromWS()
	TestSetDefaults(t)
	err := setupWSConnection()
	testStandardErrorHandling(t, err)
	err = o.WsSubscribeToChannel("Pewdiepie")
	testStandardErrorHandling(t, err)
	err = o.WsUnsubscribeToChannel("T-Series")
	testStandardErrorHandling(t, err)
}

// TestWsLogin API endpoint test
func TestWsLogin(t *testing.T) {
	defer disconnectFromWS()
	TestSetDefaults(t)
	err := setupWSConnection()
	testStandardErrorHandling(t, err)
	err = o.WsLogin()
	time.Sleep(5 * time.Second)
	testStandardErrorHandling(t, err)
}

//TestGetAssetTypeFromTableName logic test
func TestGetAssetTypeFromTableName(t *testing.T) {
	str := "spot/candle300s:BTC-USDT"
	spot := o.GetAssetTypeFromTableName(str)
	if spot != "SPOT" {
		t.Errorf("Error, expected 'SPOT', received: '%v'", spot)
	}
}

// TestOrderBookUpdateChecksumCalculator logic test
func TestOrderBookUpdateChecksumCalculator(t *testing.T) {
	original := `{"table":"spot/depth","action":"partial","data":[{"instrument_id":"BTC-USDT","asks":[["3864.6786","0.145",1],["3864.7682","0.005",1],["3864.9851","0.57",1],["3864.9852","0.30137754",1],["3864.9986","2.81818419",1],["3864.9995","0.002",1],["3865","0.0597",1],["3865.0309","0.4",1],["3865.1995","0.004",1],["3865.3995","0.004",1],["3865.5995","0.004",1],["3865.7995","0.004",1],["3865.9995","0.004",1],["3866.0961","0.25865886",1],["3866.1995","0.004",1],["3866.3995","0.004",1],["3866.4004","0.3243",2],["3866.5995","0.004",1],["3866.7633","0.44247086",1],["3866.7995","0.004",1],["3866.9197","0.511",1],["3867.256","0.51716256",1],["3867.3951","0.02588112",1],["3867.4014","0.025",1],["3867.4566","0.02499999",1],["3867.4675","4.01155057",5],["3867.5515","1.1",1],["3867.6113","0.009",1],["3867.7349","0.026",1],["3867.7781","0.03738652",1],["3867.9163","0.0521",1],["3868.0381","0.34354941",1],["3868.0436","0.051",1],["3868.0657","0.90552172",3],["3868.1819","0.03863346",1],["3868.2013","0.194",1],["3868.346","0.051",1],["3868.3863","0.01155",1],["3868.7716","0.009",1],["3868.947","0.025",1],["3868.98","0.001",1],["3869.0764","1.03487931",1],["3869.2773","0.07724578",1],["3869.4039","0.025",1],["3869.4068","1.03",1],["3869.7068","2.06976398",1],["3870","0.5",1],["3870.0465","0.01",1],["3870.7042","0.02099651",1],["3870.9451","2.07047375",1],["3871.5254","1.2",1],["3871.5596","0.001",1],["3871.6605","0.01035032",1],["3871.7179","2.07047375",1],["3871.8816","0.51751625",1],["3872.1","0.75",1],["3872.2464","0.0646",1],["3872.3747","0.283",1],["3872.4039","0.2",1],["3872.7655","0.23179307",1],["3872.8005","2.06976398",1],["3873.1509","2",1],["3873.3215","0.26",1],["3874.1392","0.001",1],["3874.1487","3.88224364",4],["3874.1685","1.8",1],["3874.5571","0.08974762",1],["3874.734","2.06976398",1],["3874.99","0.3",1],["3875","1.001",2],["3875.0041","1.03505051",1],["3875.45","0.3",1],["3875.4766","0.15",1],["3875.7057","0.51751625",1],["3876","0.001",1],["3876.68","0.3",1],["3876.7188","0.001",1],["3877","0.75",1],["3877.31","0.035",1],["3877.38","0.3",1],["3877.7","0.3",1],["3877.88","0.3",1],["3878.0364","0.34770122",1],["3878.4525","0.48579748",1],["3878.4955","0.02812511",1],["3878.8855","0.00258579",1],["3878.9605","0.895",1],["3879","0.001",1],["3879.2984","0.002",2],["3879.432","0.001",1],["3879.6313","6",1],["3879.9999","0.002",2],["3880","1.25132834",5],["3880.2526","0.04075162",1],["3880.7145","0.0647",1],["3881.2469","1.883",1],["3881.878","0.002",2],["3884.4576","0.002",2],["3885","0.002",2],["3885.2233","0.28304103",1],["3885.7416","18",1],["3886","0.001",1],["3886.1554","5.4",1],["3887","0.001",1],["3887.0372","0.002",2],["3887.2559","0.05214011",1],["3887.9238","0.0019",1],["3888","0.15810538",4],["3889","0.001",1],["3889.5175","0.50510653",1],["3889.6168","0.002",2],["3889.9999","0.001",1],["3890","2.34968109",4],["3890.5222","0.00257806",1],["3891.2659","5",1],["3891.9999","0.00893897",1],["3892.1964","0.002",2],["3892.4358","0.0176",1],["3893.1388","1.4279",1],["3894","0.0026321",1],["3894.776","0.001",1],["3895","1.501",2],["3895.379","0.25881288",1],["3897","0.05",1],["3897.3556","0.001",1],["3897.8432","0.73708079",1],["3898","3.31353018",7],["3898.4462","4.757",1],["3898.6","0.47159638",1],["3898.8769","0.0129",1],["3899","6",2],["3899.6516","0.025",1],["3899.9352","0.001",1],["3899.9999","0.013",2],["3900","22.37447743",24],["3900.9999","0.07763916",1],["3901","0.10192487",1],["3902.1937","0.00257034",1],["3902.3991","1.5532141",1],["3902.5148","0.001",1],["3904","1.49331984",1],["3904.9999","0.95905447",1],["3905","0.501",2],["3905.0944","0.001",1],["3905.61","0.099",1],["3905.6801","0.54343686",1],["3906.2901","0.0258",1],["3907.674","0.001",1],["3907.85","1.35778084",1],["3908","0.03846153",1],["3908.23","1.95189531",1],["3908.906","0.03148978",1],["3909","0.001",1],["3909.9999","0.01398721",2],["3910","0.016",2],["3910.2536","0.001",1],["3912.5406","0.88270517",1],["3912.8332","0.001",1],["3913","1.2640608",1],["3913.87","1.69114184",1],["3913.9003","0.00256266",1],["3914","1.21766411",1],["3915","0.001",1],["3915.4128","0.001",1],["3915.7425","6.848",1],["3916","0.0050949",1],["3917.36","1.28658296",1],["3917.9924","0.001",1],["3919","0.001",1],["3919.9999","0.001",1],["3920","1.21171832",3],["3920.0002","0.20217038",1],["3920.572","0.001",1],["3921","0.128",1],["3923.0756","0.00148064",1],["3923.1516","0.001",1],["3923.86","1.38831714",1],["3925","0.01867801",2],["3925.642","0.00255499",1],["3925.7312","0.001",1],["3926","0.04290757",1],["3927","0.023",1],["3927.3175","0.01212865",1],["3927.65","1.51375612",1],["3928","0.5",1],["3928.3108","0.001",1],["3929","0.001",1],["3929.9999","0.01519338",2],["3930","0.0174985",3],["3930.21","1.49335799",1],["3930.8904","0.001",1],["3932.2999","0.01953",1],["3932.8962","7.96",1],["3933.0387","11.808",1],["3933.47","0.001",1],["3934","1.40839932",1],["3935","0.001",1],["3936.8","0.62879518",1],["3937.23","1.56977841",1],["3937.4189","0.00254735",1]],"bids":[["3864.5217","0.00540709",1],["3864.5216","0.14068758",2],["3864.2275","0.01033576",1],["3864.0989","0.00825047",1],["3864.0273","0.38",1],["3864.0272","0.4",1],["3863.9957","0.01083539",1],["3863.9184","0.01653723",1],["3863.8282","0.25588165",1],["3863.8153","0.154",1],["3863.7791","1.14122492",1],["3863.6866","0.01733662",1],["3863.6093","0.02645958",1],["3863.3775","0.02773862",1],["3863.0297","0.513",1],["3863.0286","1.1028564",2],["3862.8489","0.01",1],["3862.5972","0.01890179",1],["3862.3431","0.01152944",1],["3862.313","0.009",1],["3862.2445","0.90551002",3],["3862.0734","0.014",1],["3862.0539","0.64976067",1],["3861.8586","0.025",1],["3861.7888","0.025",1],["3861.7673","0.008",1],["3861.5785","0.01",1],["3861.3895","0.005",1],["3861.3338","0.25875855",1],["3861.161","0.01",1],["3861.1111","0.03863352",1],["3861.0732","0.51703882",1],["3860.9116","0.17754895",1],["3860.75","0.19",1],["3860.6554","0.015",1],["3860.6172","0.005",1],["3860.6088","0.008",1],["3860.4724","0.12940042",1],["3860.4424","0.25880084",1],["3860.42","0.01",1],["3860.3725","0.51760102",1],["3859.8449","0.005",1],["3859.8285","0.03738652",1],["3859.7638","0.07726703",1],["3859.4502","0.008",1],["3859.3772","0.05173471",1],["3859.3409","0.194",1],["3859","5",1],["3858.827","0.0521",1],["3858.8208","0.001",1],["3858.679","0.26",1],["3858.4814","0.07477305",1],["3858.1669","1.03503422",1],["3857.6005","0.006",1],["3857.4005","0.004",1],["3857.2005","0.004",1],["3857.1871","1.218",1],["3857.0005","0.004",1],["3856.8135","0.0646",1],["3856.8005","0.004",1],["3856.2412","0.001",1],["3856.2349","1.03503422",1],["3856.0197","0.01037339",1],["3855.8781","0.23178117",1],["3855.8005","0.004",1],["3855.7165","0.00259355",1],["3855.4858","0.25875855",1],["3854.4584","0.01",1],["3853.6616","0.001",1],["3853.1373","0.92",1],["3852.5072","0.48599702",1],["3851.3926","0.13008333",1],["3851.082","0.001",1],["3850.9317","2",1],["3850.6359","0.34770165",1],["3850.2058","0.51751624",1],["3850.0823","0.15",1],["3850.0042","0.5175171",1],["3850","0.001",1],["3849.6325","1.8",1],["3849.41","0.3",1],["3848.9686","1.85",1],["3848.7426","0.18511466",1],["3848.52","0.3",1],["3848.5024","0.001",1],["3848.42","0.3",1],["3848.1618","2.204",1],["3847.77","0.3",1],["3847.48","0.3",1],["3847.3581","2.05",1],["3846.8259","0.0646",1],["3846.59","0.3",1],["3846.49","0.3",1],["3845.9228","0.001",1],["3844.184","0.00260133",1],["3844.0092","6.3",1],["3843.3432","0.001",1],["3841","0.06300963",1],["3840.7636","0.001",1],["3840","0.201",3],["3839.7681","18",1],["3839.5328","0.05214011",1],["3838.184","0.001",1],["3837.2344","0.27589557",1],["3836.6479","5.2",1],["3836","2.37196773",3],["3835.6044","0.001",1],["3833.6053","0.25873556",1],["3833.0248","0.001",1],["3833","0.8726502",1],["3832.6859","0.00260913",1],["3832","0.007",1],["3831.637","6",1],["3831.0602","0.001",1],["3830.4452","0.001",1],["3830","0.20375718",4],["3829.7125","0.07833486",1],["3829.6283","0.3519681",1],["3829","0.0039261",1],["3827.8656","0.001",1],["3826.0001","0.53251232",1],["3826","0.0509",1],["3825.7834","0.00698562",1],["3825.286","0.001",1],["3823.0001","0.03010127",1],["3822.8014","0.00261588",1],["3822.7064","0.001",1],["3822.2","1",1],["3822.1121","0.35994101",1],["3821.2222","0.00261696",1],["3821","0.001",1],["3820.1268","0.001",1],["3820","1.12992803",4],["3819","0.01331195",2],["3817.5472","0.001",1],["3816","1.13807184",2],["3815.8343","0.32463428",1],["3815.7834","0.00525295",1],["3815","28.99386799",4],["3814.9676","0.001",1],["3813","0.91303023",4],["3812.388","0.002",2],["3811.2257","0.07",1],["3810","0.32573997",2],["3809.8084","0.001",1],["3809.7928","0.00262481",1],["3807.2288","0.001",1],["3806.8421","0.07003461",1],["3806","0.19",1],["3805.8041","0.05678805",1],["3805","1.01",2],["3804.6492","0.001",1],["3804.3551","0.1",1],["3803","0.005",1],["3802.22","2.05042631",1],["3802.0696","0.001",1],["3802","1.63290092",1],["3801.2257","0.07",1],["3801","57.4",3],["3800.9853","0.02492278",1],["3800.8421","0.06503533",1],["3800.7844","0.02812628",1],["3800.0001","0.00409473",1],["3800","17.91401074",15],["3799.49","0.001",1],["3799","0.1",1],["3796.9104","0.001",1],["3796","9.00128053",2],["3795.5441","0.0028",1],["3794.3308","0.001",1],["3791","55",1],["3790.7777","0.07",1],["3790","12.03238184",7],["3789","1",1],["3788","0.21110454",2],["3787.2959","9",1],["3786.592","0.001",1],["3786","9.01916822",2],["3785","12.87914268",5],["3784.0124","0.001",1],["3781.4328","0.002",2],["3781","56.3",2],["3780.7777","0.07",1],["3780","23.41537654",10],["3778.8532","0.002",2],["3776","9",1],["3774","0.003",1],["3772.2481","0.06901672",1],["3771","55.1",2],["3770.7777","0.07",1],["3770","7.30268416",5],["3769","0.25",1],["3768","1.3725",3],["3766.66","0.02",1],["3766","7.64837924",2],["3765.58","1.22775492",1],["3762.58","1.22873383",1],["3761","51.68262164",1],["3760.8031","0.0399",1],["3760.7777","0.07",1]],"timestamp":"2019-03-06T23:19:17.705Z","checksum":-1785549915}]}`
	update := `{"table":"spot/depth","action":"update","data":[{"instrument_id":"BTC-USDT","asks":[["3864.6786","0",0],["3864.9852","0",0],["3865.9994","0.48402971",1],["3866.4004","0.001",1],["3866.7995","0.3273",2],["3867.4566","0",0],["3867.7031","0.025",1],["3868.0436","0",0],["3868.346","0",0],["3868.3695","0.051",1],["3870.9243","0.642",1],["3874.9942","0.51751796",1],["3875.7057","0",0],["3939","0.001",1]],"bids":[["3864.55","0.0565449",1],["3863.8282","0",0],["3863.8153","0",0],["3863.7898","0.01320077",1],["3863.4807","0.02112123",1],["3863.3002","0.04233533",1],["3863.1717","0.03379397",1],["3863.0685","0.04438179",1],["3863.0286","0.7362564",1],["3862.9912","0.06773651",1],["3862.8626","0.05407035",1],["3862.7595","0.07101087",1],["3862.313","0.3756",2],["3862.1848","0.012",1],["3862.0734","0",0],["3861.8391","0.025",1],["3861.7888","0",0],["3856.6716","0.38893641",1],["3768","0",0],["3766.66","0",0],["3766","0",0],["3765.58","0",0],["3762.58","0",0],["3761","0",0],["3760.8031","0",0],["3760.7777","0",0]],"timestamp":"2019-03-06T23:19:18.239Z","checksum":-1587788848}]}`
	TestWSSetup(t)
	var dataResponse okgroup.WebsocketDataResponse
	err := common.JSONDecode([]byte(original), &dataResponse)
	if err != nil {
		t.Error(err)
	}
	err = o.WsProcessOrderBook(dataResponse)
	if err != nil {
		t.Error(err)
		return
	}
	var updateResponse okgroup.WebsocketDataResponse
	err = common.JSONDecode([]byte(update), &updateResponse)
	if err != nil {
		t.Error(err)
	}
	err = o.WsProcessOrderBook(updateResponse)
	if err != nil {
		t.Error(err)
	}
}

// TestOrderBookPartialChecksumCalculator logic test
func TestOrderBookPartialChecksumCalculator(t *testing.T) {
	orderbookPartialJSON := `{"table":"spot/depth","action":"partial","data":[{"instrument_id":"EOS-USDT","asks":[["3.5196","0.1077",1],["3.5198","21.71",1],["3.5199","51.1805",1],["3.5208","75.09",1],["3.521","196.3333",1],["3.5213","0.1",1],["3.5218","39.276",2],["3.5219","395.6334",1],["3.522","27.956",1],["3.5222","404.9595",1],["3.5225","300",1],["3.5227","143.5442",2],["3.523","42.4746",1],["3.5231","852.64",2],["3.5235","34.9602",1],["3.5237","442.0918",2],["3.5238","352.8404",2],["3.5239","341.6759",2],["3.524","84.9493",1],["3.5241","148.4882",1],["3.5242","261.64",1],["3.5243","142.045",1],["3.5246","10",1],["3.5247","284.0788",1],["3.5248","720",1],["3.5249","89.2518",2],["3.5251","1201.8965",2],["3.5254","426.2938",1],["3.5255","213.0863",1],["3.5257","568.1576",1],["3.5258","0.3",1],["3.5259","34.4602",1],["3.526","0.1",1],["3.5263","850.771",1],["3.5265","5.9",1],["3.5268","10.5064",2],["3.5272","1136.8965",1],["3.5274","255.1481",1],["3.5276","29.5374",1],["3.5278","50",1],["3.5282","284.1797",1],["3.5283","1136.8965",1],["3.5284","0.4275",1],["3.5285","100",1],["3.5292","90.9",1],["3.5298","0.2",1],["3.5303","568.1576",1],["3.5305","279.9999",1],["3.532","0.409",1],["3.5321","568.1576",1],["3.5326","6016.8756",1],["3.5328","4.9849",1],["3.533","92.88",2],["3.5343","1200.2383",2],["3.5344","100",1],["3.535","359.7047",1],["3.5354","100",1],["3.5355","100",1],["3.5356","10",1],["3.5358","200",2],["3.5362","435.139",1],["3.5365","2152",1],["3.5366","284.1756",1],["3.5367","568.4644",1],["3.5369","33.9878",1],["3.537","337.1191",2],["3.5373","0.4045",1],["3.5383","1136.7188",1],["3.5386","12.1614",1],["3.5387","90.89",1],["3.54","4.54",1],["3.5423","90.8",1],["3.5436","0.1",1],["3.5454","853.4156",1],["3.5468","142.0656",1],["3.5491","0.0008",1],["3.55","14478.8206",6],["3.5537","21521",1],["3.5555","11.53",1],["3.5573","50.6001",1],["3.5599","4591.4221",1],["3.56","1227.0002",4],["3.5603","2670",1],["3.5608","58.6638",1],["3.5613","0.1",1],["3.5621","45.9473",1],["3.57","2141.7274",3],["3.5712","2956.9816",1],["3.5717","27.9978",1],["3.5718","0.9285",1],["3.5739","299.73",1],["3.5761","864",1],["3.579","22.5225",1],["3.5791","38.26",2],["3.58","7618.4634",5],["3.5801","457.2184",1],["3.582","24.5",1],["3.5822","1572.6425",1],["3.5845","14.1438",1],["3.585","527.169",1],["3.5865","20",1],["3.5867","4490",1],["3.5876","39.0493",1],["3.5879","392.9083",1],["3.5888","436.42",2],["3.5896","50",1],["3.59","2608.9128",8],["3.5913","19.5246",1],["3.5938","7082",1],["3.597","0.1",1],["3.5979","399",1],["3.5995","315.1509",1],["3.5999","2566.2648",1],["3.6","18511.2292",35],["3.603","22.3379",2],["3.605","499.5",1],["3.6055","100",1],["3.6058","499.5",1],["3.608","1021.1485",1],["3.61","11755.4596",13],["3.611","42.8571",1],["3.6131","6690",1],["3.6157","19.5247",1],["3.618","2500",1],["3.6197","525.7146",1],["3.6198","0.4455",1],["3.62","6440.6295",8],["3.6219","0.4175",1],["3.6237","168",1],["3.6265","0.1001",1],["3.628","64.9345",1],["3.63","4435.4985",6],["3.6308","1.7815",1],["3.6331","0.1",1],["3.6338","355.527",2],["3.6358","50",1],["3.6363","2074.7096",1],["3.6376","4000",1],["3.6396","11090",1],["3.6399","0.4055",1],["3.64","4161.9805",4],["3.6437","117.6524",1],["3.648","190",1],["3.6488","200",1],["3.65","11740.5045",25],["3.6512","0.1",1],["3.6521","728",1],["3.6555","100",1],["3.6598","36.6914",1],["3.66","4331.2148",6],["3.6638","200",1],["3.6673","100",1],["3.6679","38",1],["3.6688","2",1],["3.6695","0.1",1],["3.67","7984.698",6],["3.672","300",1],["3.6777","257.8247",1],["3.6789","393.4217",2],["3.68","9202.3222",11],["3.6818","500",1],["3.6823","299.7",1],["3.6839","422.3748",1],["3.685","100",1],["3.6878","0.1",1],["3.6888","72.0958",2],["3.6889","2876",1],["3.689","28",1],["3.6891","28",1],["3.6892","28",1],["3.6895","28",1],["3.6898","28",1],["3.69","643.96",7],["3.6908","118",2],["3.691","28",1],["3.6916","28",1],["3.6918","28",1],["3.6926","28",1],["3.6928","28",1],["3.6932","28",1],["3.6933","200",1],["3.6935","28",1],["3.6936","28",1],["3.6938","28",1],["3.694","28",1],["3.698","1498.5",1],["3.6988","2014.2004",2],["3.7","21904.2689",22],["3.7029","71.95",1],["3.704","3690.1362",1],["3.7055","100",1],["3.7063","0.1",1],["3.71","4421.3468",4],["3.719","17.3491",1],["3.72","1304.5995",3],["3.7211","10",1],["3.7248","0.1",1],["3.725","1900",1],["3.73","31.1785",2],["3.7375","38",1]],"bids":[["3.5182","151.5343",6],["3.5181","0.3691",1],["3.518","271.3967",2],["3.5179","257.8352",1],["3.5178","12.3811",1],["3.5173","34.1921",2],["3.5171","1013.8256",2],["3.517","272.1119",2],["3.5168","395.3376",1],["3.5166","317.1756",2],["3.5165","348.302",3],["3.5164","142.0414",1],["3.5163","96.8933",2],["3.516","600.1034",3],["3.5159","27.481",1],["3.5158","27.33",1],["3.5157","583.1898",2],["3.5156","24.6819",2],["3.5154","25",1],["3.5153","0.429",1],["3.5152","453.9204",3],["3.5151","2131.592",4],["3.515","335",3],["3.5149","37.1586",1],["3.5147","41.6759",1],["3.5146","54.569",1],["3.5145","70.3515",1],["3.5143","68.206",3],["3.5142","359.4538",2],["3.5139","45.4123",2],["3.5137","71.673",2],["3.5136","25",1],["3.5135","300",1],["3.5134","442.57",2],["3.5132","83.3518",1],["3.513","1245.2529",3],["3.5127","20",1],["3.512","284.1353",1],["3.5119","1136.8319",1],["3.5113","56.9351",1],["3.5111","588.1898",2],["3.5109","255.0946",1],["3.5105","48.65",1],["3.5103","50.2",1],["3.5098","720",1],["3.5096","148.95",1],["3.5094","570.5758",2],["3.509","2.386",1],["3.5089","0.4065",1],["3.5087","282.3859",2],["3.5086","145.036",2],["3.5084","2.386",1],["3.5082","90.98",1],["3.5081","2.386",1],["3.5079","2.386",1],["3.5078","857.6229",2],["3.5075","2.386",1],["3.5074","284.1877",1],["3.5073","100",1],["3.5071","100",1],["3.507","768.4159",3],["3.5069","313.0863",2],["3.5068","426.2938",1],["3.5066","568.3594",1],["3.5063","1136.6865",1],["3.5059","0.3",1],["3.5054","9.9999",1],["3.5053","0.2",1],["3.5051","392.428",1],["3.505","13.79",1],["3.5048","99.5497",2],["3.5047","78.5331",2],["3.5046","2153",1],["3.5041","5983.999",1],["3.5037","668.5682",1],["3.5036","160.5948",1],["3.5024","534.8075",1],["3.5014","28.5604",1],["3.5011","91",1],["3.5","1058.8771",2],["3.4997","50.2",1],["3.4985","3430.0414",1],["3.4949","232.0591",1],["3.4942","21521",1],["3.493","2",1],["3.4928","2",1],["3.4925","0.44",1],["3.4917","142.0656",1],["3.49","2051.8826",4],["3.488","280.7459",1],["3.4852","643.4038",1],["3.4851","86.0807",1],["3.485","213.2436",1],["3.484","0.1",1],["3.4811","144.3399",1],["3.4808","89",1],["3.4803","12.1999",1],["3.4801","2390",1],["3.48","930.8453",9],["3.4791","310",1],["3.4768","206",1],["3.4767","0.9415",1],["3.4754","1.4387",1],["3.4728","20",1],["3.4701","1219.2873",1],["3.47","1904.3139",7],["3.468","0.4035",1],["3.4667","0.1",1],["3.4666","3020.0101",1],["3.465","10",1],["3.464","0.4485",1],["3.462","2119.6556",1],["3.46","1305.6113",8],["3.4589","8.0228",1],["3.457","100",1],["3.456","70.3859",2],["3.4538","20",1],["3.4536","4323.9486",2],["3.4531","827.0427",1],["3.4528","0.439",1],["3.4522","8.0381",1],["3.4513","441.1873",1],["3.4512","50.707",1],["3.451","87.0902",1],["3.4509","200",1],["3.4506","100",1],["3.4505","86.4045",2],["3.45","12409.4595",28],["3.4494","0.5365",2],["3.449","10761",1],["3.4482","8.0476",1],["3.4469","0.449",1],["3.445","2000",1],["3.4427","14",1],["3.4421","100",1],["3.4416","8.0631",1],["3.4404","1",1],["3.44","4580.733",11],["3.4388","1868.2085",1],["3.438","937.7246",2],["3.4367","1500",1],["3.4366","62",1],["3.436","29.8743",1],["3.4356","25.4801",1],["3.4349","4.3086",1],["3.4343","43.2402",1],["3.433","2.0688",1],["3.4322","2.7335",2],["3.432","93.3233",1],["3.4302","328.8301",2],["3.43","4440.8158",11],["3.4288","754.574",2],["3.4283","125.7043",2],["3.428","744.3154",2],["3.4273","5460",1],["3.4258","50",1],["3.4255","109.005",1],["3.4248","100",1],["3.4241","129.2048",2],["3.4233","5.3598",1],["3.4228","4498.866",1],["3.4222","3.5435",1],["3.4217","404.3252",2],["3.4211","1000",1],["3.4208","31",1],["3.42","1834.024",9],["3.4175","300",1],["3.4162","400",1],["3.4152","0.1",1],["3.4151","4.3336",1],["3.415","1.5974",1],["3.414","1146",1],["3.4134","306.4246",1],["3.4129","7.5556",1],["3.4111","198.5188",1],["3.4109","500",1],["3.4106","4305",1],["3.41","2150.7635",13],["3.4085","4.342",1],["3.4054","5.6985",1],["3.4019","5.438",1],["3.4015","1010.846",1],["3.4009","8610",1],["3.4005","1.9122",1],["3.4004","1",1],["3.4","27081.1806",67],["3.3955","3.2682",1],["3.3953","5.4486",1],["3.3937","1591.3805",1],["3.39","3221.4155",8],["3.3899","3.2736",1],["3.3888","1500",2],["3.3887","5.4592",1],["3.385","117.0969",2],["3.3821","5.4699",1],["3.382","100.0529",1],["3.3818","172.0164",1],["3.3815","165.6288",1],["3.381","887.3115",1],["3.3808","100",1]],"timestamp":"2019-03-04T00:15:04.155Z","checksum":-2036653089}]}`
	var dataResponse okgroup.WebsocketDataResponse
	err := common.JSONDecode([]byte(orderbookPartialJSON), &dataResponse)
	if err != nil {
		t.Error(err)
	}
	calculatedChecksum := o.CalculatePartialOrderbookChecksum(dataResponse.Data[0])

	if calculatedChecksum != dataResponse.Data[0].Checksum {
		t.Errorf("Expected %v, Receieved %v", dataResponse.Data[0].Checksum, calculatedChecksum)
	}
}

// Function tests ----------------------------------------------------------------------------------------------
func setFeeBuilder() exchange.FeeBuilder {
	return exchange.FeeBuilder{
		Amount:              1,
		Delimiter:           "-",
		FeeType:             exchange.CryptocurrencyTradeFee,
		FirstCurrency:       symbol.LTC,
		SecondCurrency:      symbol.BTC,
		IsMaker:             false,
		PurchasePrice:       1,
		CurrencyItem:        symbol.USD,
		BankTransactionType: exchange.WireTransfer,
	}
}

// TestGetFee fee calcuation test
func TestGetFee(t *testing.T) {
	TestSetDefaults(t)
	var feeBuilder = setFeeBuilder()
	// CryptocurrencyTradeFee Basic
	if resp, err := o.GetFee(feeBuilder); resp != float64(0.0015) || err != nil {
		t.Error(err)
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.0015), resp)
	}

	// CryptocurrencyTradeFee High quantity
	feeBuilder = setFeeBuilder()
	feeBuilder.Amount = 1000
	feeBuilder.PurchasePrice = 1000
	if resp, err := o.GetFee(feeBuilder); resp != float64(1500) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(1500), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee IsMaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = true
	if resp, err := o.GetFee(feeBuilder); resp != float64(0.001) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.001), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee Negative purchase price
	feeBuilder = setFeeBuilder()
	feeBuilder.PurchasePrice = -1000
	if resp, err := o.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := o.GetFee(feeBuilder); resp != float64(0.001) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.001), resp)
		t.Error(err)
	}

	// CryptocurrencyWithdrawalFee Invalid currency
	feeBuilder = setFeeBuilder()
	feeBuilder.FirstCurrency = "hello"
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := o.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := o.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	if resp, err := o.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.CurrencyItem = symbol.USD
	if resp, err := o.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
}

// TestFormatWithdrawPermissions helper test
func TestFormatWithdrawPermissions(t *testing.T) {
	TestSetDefaults(t)
	expectedResult := exchange.AutoWithdrawCryptoText + " & " + exchange.NoFiatWithdrawalsText
	withdrawPermissions := o.FormatWithdrawPermissions()
	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Received: %s", expectedResult, withdrawPermissions)
	}
}

// Wrapper tests --------------------------------------------------------------------------------------------------

// TestSubmitOrder Wrapper test
func TestSubmitOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	var p = pair.CurrencyPair{
		Delimiter:      "",
		FirstCurrency:  symbol.BTC,
		SecondCurrency: symbol.USDT,
	}
	response, err := o.SubmitOrder(p, exchange.BuyOrderSide, exchange.MarketOrderType, 1, 10, "hi")
	if areTestAPIKeysSet() && (err != nil || !response.IsOrderPlaced) {
		t.Errorf("Order failed to be placed: %v", err)
	} else if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

// TestCancelExchangeOrder Wrapper test
func TestCancelExchangeOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	currencyPair := pair.NewCurrencyPair(symbol.LTC, symbol.BTC)
	var orderCancellation = exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	err := o.CancelOrder(orderCancellation)
	testStandardErrorHandling(t, err)

}

// TestCancelAllExchangeOrders Wrapper test
func TestCancelAllExchangeOrders(t *testing.T) {
	TestSetRealOrderDefaults(t)
	currencyPair := pair.NewCurrencyPair(symbol.LTC, symbol.BTC)
	var orderCancellation = exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	resp, err := o.CancelAllOrders(orderCancellation)
	testStandardErrorHandling(t, err)

	if len(resp.OrderStatus) > 0 {
		t.Errorf("%v orders failed to cancel", len(resp.OrderStatus))
	}
}

// TestGetAccountInfo Wrapper test
func TestGetAccountInfo(t *testing.T) {
	_, err := o.GetAccountInfo()
	testStandardErrorHandling(t, err)
}

// TestModifyOrder Wrapper test
func TestModifyOrder(t *testing.T) {
	TestSetRealOrderDefaults(t)
	_, err := o.ModifyOrder(exchange.ModifyOrder{})
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}

// TestWithdraw Wrapper test
func TestWithdraw(t *testing.T) {
	TestSetRealOrderDefaults(t)
	var withdrawCryptoRequest = exchange.WithdrawRequest{
		Amount:        100,
		Currency:      symbol.BTC,
		Address:       "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		Description:   "WITHDRAW IT ALL",
		TradePassword: "Password",
		FeeAmount:     1,
	}
	_, err := o.WithdrawCryptocurrencyFunds(withdrawCryptoRequest)
	testStandardErrorHandling(t, err)
}

// TestWithdrawFiat Wrapper test
func TestWithdrawFiat(t *testing.T) {
	TestSetRealOrderDefaults(t)
	var withdrawFiatRequest = exchange.WithdrawRequest{}
	_, err := o.WithdrawFiatFunds(withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}

// TestSubmitOrder Wrapper test
func TestWithdrawInternationalBank(t *testing.T) {
	TestSetRealOrderDefaults(t)
	var withdrawFiatRequest = exchange.WithdrawRequest{}
	_, err := o.WithdrawFiatFundsToInternationalBank(withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}
