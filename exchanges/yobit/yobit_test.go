package yobit

import (
	"math"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

var y Yobit

// Please supply your own keys for better unit testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

func TestSetDefaults(t *testing.T) {
	y.SetDefaults()
}

func TestSetup(t *testing.T) {
	yobitConfig := config.GetConfig()
	err := yobitConfig.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		t.Fatal("Test Failed - Yobit load config error", err)
	}
	conf, err := yobitConfig.GetExchangeConfig("Yobit")
	if err != nil {
		t.Fatal("Test Failed - Yobit init error", err)
	}
	conf.API.Credentials.Key = apiKey
	conf.API.Credentials.Secret = apiSecret
	conf.API.AuthenticatedSupport = true

	err = y.Setup(conf)
	if err != nil {
		t.Fatal("Test Failed - Yobit setup error", err)
	}
}

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	_, err := y.FetchTradablePairs(asset.Spot)
	if err != nil {
		t.Errorf("Test failed. FetchTradablePairs err: %s", err)
	}
}

func TestGetInfo(t *testing.T) {
	t.Parallel()
	_, err := y.GetInfo()
	if err != nil {
		t.Error("Test Failed - GetInfo() error")
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := y.GetTicker("btc_usd")
	if err != nil {
		t.Error("Test Failed - GetTicker() error", err)
	}
}

func TestGetDepth(t *testing.T) {
	t.Parallel()
	_, err := y.GetDepth("btc_usd")
	if err != nil {
		t.Error("Test Failed - GetDepth() error", err)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()
	_, err := y.GetTrades("btc_usd")
	if err != nil {
		t.Error("Test Failed - GetTrades() error", err)
	}
}

func TestGetAccountInfo(t *testing.T) {
	t.Parallel()
	_, err := y.GetAccountInfo()
	if err == nil {
		t.Error("Test Failed - GetAccountInfo() error", err)
	}
}

func TestGetOpenOrders(t *testing.T) {
	t.Parallel()
	_, err := y.GetOpenOrders("")
	if err == nil {
		t.Error("Test Failed - GetOpenOrders() error", err)
	}
}

func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	_, err := y.GetOrderInfo("6196974")
	if err == nil {
		t.Error("Test Failed - GetOrderInfo() error", err)
	}
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	err := y.CancelExistingOrder(1337)
	if err == nil {
		t.Error("Test Failed - CancelOrder() error", err)
	}
}

func TestTrade(t *testing.T) {
	t.Parallel()
	_, err := y.Trade("", order.Buy.String(), 0, 0)
	if err == nil {
		t.Error("Test Failed - Trade() error", err)
	}
}

func TestWithdrawCoinsToAddress(t *testing.T) {
	t.Parallel()
	_, err := y.WithdrawCoinsToAddress("", 0, "")
	if err == nil {
		t.Error("Test Failed - WithdrawCoinsToAddress() error", err)
	}
}

func TestCreateYobicode(t *testing.T) {
	t.Parallel()
	_, err := y.CreateCoupon("bla", 0)
	if err == nil {
		t.Error("Test Failed - CreateYobicode() error", err)
	}
}

func TestRedeemYobicode(t *testing.T) {
	t.Parallel()
	_, err := y.RedeemCoupon("bla2")
	if err == nil {
		t.Error("Test Failed - RedeemYobicode() error", err)
	}
}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:  1,
		FeeType: exchange.CryptocurrencyTradeFee,
		Pair: currency.NewPairWithDelimiter(currency.LTC.String(),
			currency.BTC.String(),
			"-"),
		PurchasePrice:       1,
		FiatCurrency:        currency.USD,
		BankTransactionType: exchange.WireTransfer,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	var feeBuilder = setFeeBuilder()
	y.GetFeeByType(feeBuilder)
	if apiKey == "" || apiSecret == "" {
		if feeBuilder.FeeType != exchange.OfflineTradeFee {
			t.Errorf("Expected %v, received %v", exchange.OfflineTradeFee, feeBuilder.FeeType)
		}
	} else {
		if feeBuilder.FeeType != exchange.CryptocurrencyTradeFee {
			t.Errorf("Expected %v, received %v", exchange.CryptocurrencyTradeFee, feeBuilder.FeeType)
		}
	}
}

func TestGetFee(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)
	var feeBuilder = setFeeBuilder()

	// CryptocurrencyTradeFee Basic
	if resp, err := y.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
		t.Error(err)
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.0015), resp)
	}

	// CryptocurrencyTradeFee High quantity
	feeBuilder = setFeeBuilder()
	feeBuilder.Amount = 1000
	feeBuilder.PurchasePrice = 1000
	if resp, err := y.GetFee(feeBuilder); resp != float64(2000) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(2000), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee IsMaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = true
	if resp, err := y.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.002), resp)
		t.Error(err)
	}

	// CryptocurrencyTradeFee Negative purchase price
	feeBuilder = setFeeBuilder()
	feeBuilder.PurchasePrice = -1000
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := y.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.002), resp)
		t.Error(err)
	}

	// CryptocurrencyWithdrawalFee Invalid currency
	feeBuilder = setFeeBuilder()
	feeBuilder.Pair.Base = currency.NewCode("hello")
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee QIWI
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	feeBuilder.BankTransactionType = exchange.Qiwi
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Wire
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	feeBuilder.BankTransactionType = exchange.WireTransfer
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Payeer
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	feeBuilder.BankTransactionType = exchange.Payeer
	if resp, err := y.GetFee(feeBuilder); resp != float64(0.03) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.03), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Capitalist
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.RUR
	feeBuilder.BankTransactionType = exchange.Capitalist
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee AdvCash
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.USD
	feeBuilder.BankTransactionType = exchange.AdvCash
	if resp, err := y.GetFee(feeBuilder); resp != float64(0.04) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0.04), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee PerfectMoney
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.RUR
	feeBuilder.BankTransactionType = exchange.PerfectMoney
	if resp, err := y.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	y.SetDefaults()
	expectedResult := exchange.AutoWithdrawCryptoWithAPIPermissionText + " & " + exchange.WithdrawFiatViaWebsiteOnlyText

	withdrawPermissions := y.FormatWithdrawPermissions()

	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Received: %s", expectedResult, withdrawPermissions)
	}
}

func TestGetActiveOrders(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	var getOrdersRequest = order.GetOrdersRequest{
		OrderType: order.AnyType,
		Currencies: []currency.Pair{currency.NewPair(currency.LTC,
			currency.BTC)},
	}

	_, err := y.GetActiveOrders(&getOrdersRequest)
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Could not get open orders: %s", err)
	} else if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

func TestGetOrderHistory(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	var getOrdersRequest = order.GetOrdersRequest{
		OrderType: order.AnyType,
		Currencies: []currency.Pair{currency.NewPair(currency.LTC,
			currency.BTC)},
		StartTicks: time.Unix(0, 0),
		EndTicks:   time.Unix(math.MaxInt64, 0),
	}

	_, err := y.GetOrderHistory(&getOrdersRequest)
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Could not get order history: %s", err)
	} else if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------
func areTestAPIKeysSet() bool {
	return y.ValidateAPICredentials()
}

func TestSubmitOrder(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var orderSubmission = &order.Submit{
		Pair: currency.Pair{
			Delimiter: "_",
			Base:      currency.BTC,
			Quote:     currency.USD,
		},
		OrderSide: order.Buy,
		OrderType: order.Limit,
		Price:     1,
		Amount:    1,
		ClientID:  "meowOrder",
	}
	response, err := y.SubmitOrder(orderSubmission)
	if areTestAPIKeysSet() && (err != nil || !response.IsOrderPlaced) {
		t.Errorf("Order failed to be placed: %v", err)
	} else if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)

	var orderCancellation = &order.Cancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	err := y.CancelOrder(orderCancellation)
	if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Could not cancel orders: %v", err)
	}
}

func TestCancelAllExchangeOrders(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)

	var orderCancellation = &order.Cancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	resp, err := y.CancelAllOrders(orderCancellation)

	if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Could not cancel orders: %v", err)
	}

	if len(resp.Status) > 0 {
		t.Errorf("%v orders failed to cancel", len(resp.Status))
	}
}

func TestModifyOrder(t *testing.T) {
	_, err := y.ModifyOrder(&order.Modify{})
	if err == nil {
		t.Error("Test failed - ModifyOrder() error")
	}
}

func TestWithdraw(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)
	withdrawCryptoRequest := exchange.CryptoWithdrawRequest{
		GenericWithdrawRequestInfo: exchange.GenericWithdrawRequestInfo{
			Amount:      -1,
			Currency:    currency.BTC,
			Description: "WITHDRAW IT ALL",
		},
		Address: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
	}

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	_, err := y.WithdrawCryptocurrencyFunds(&withdrawCryptoRequest)
	if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

func TestWithdrawFiat(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = exchange.FiatWithdrawRequest{}
	_, err := y.WithdrawFiatFunds(&withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'",
			common.ErrFunctionNotSupported,
			err)
	}
}

func TestWithdrawInternationalBank(t *testing.T) {
	y.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = exchange.FiatWithdrawRequest{}
	_, err := y.WithdrawFiatFundsToInternationalBank(&withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'",
			common.ErrFunctionNotSupported,
			err)
	}
}

func TestGetDepositAddress(t *testing.T) {
	if apiKey != "" || apiSecret != "" {
		_, err := y.GetDepositAddress(currency.BTC, "")
		if err != nil {
			t.Error("Test Failed - GetDepositAddress() error", err)
		}
	} else {
		_, err := y.GetDepositAddress(currency.BTC, "")
		if err == nil {
			t.Error("Test Failed - GetDepositAddress() error")
		}
	}
}
