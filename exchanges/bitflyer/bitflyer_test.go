package bitflyer

import (
	"testing"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// Please supply your own keys here for due diligence testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

var b Bitflyer

func TestSetDefaults(t *testing.T) {
	b.SetDefaults()
}

func TestSetup(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		t.Fatal("Bitflyer load config error", err)
	}
	bitflyerConfig, err := cfg.GetExchangeConfig("Bitflyer")
	if err != nil {
		t.Error("bitflyer Setup() init error")
	}

	bitflyerConfig.API.AuthenticatedSupport = true
	bitflyerConfig.API.Credentials.Key = apiKey
	bitflyerConfig.API.Credentials.Secret = apiSecret

	err = b.Setup(bitflyerConfig)
	if err != nil {
		t.Fatal("Bitflyer setup error", err)
	}
}

func TestGetLatestBlockCA(t *testing.T) {
	t.Parallel()
	_, err := b.GetLatestBlockCA()
	if err != nil {
		t.Error("Bitflyer - GetLatestBlockCA() error:", err)
	}
}

func TestGetBlockCA(t *testing.T) {
	t.Parallel()
	_, err := b.GetBlockCA("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
	if err != nil {
		t.Error("Bitflyer - GetBlockCA() error:", err)
	}
}

func TestGetBlockbyHeightCA(t *testing.T) {
	t.Parallel()
	_, err := b.GetBlockbyHeightCA(0)
	if err != nil {
		t.Error("Bitflyer - GetBlockbyHeightCA() error:", err)
	}
}

func TestGetTransactionByHashCA(t *testing.T) {
	t.Parallel()
	_, err := b.GetTransactionByHashCA("0562d1f063cd4127053d838b165630445af5e480ceb24e1fd9ecea52903cb772")
	if err != nil {
		t.Error("Bitflyer - GetTransactionByHashCA() error:", err)
	}
}

func TestGetAddressInfoCA(t *testing.T) {
	t.Parallel()
	v, err := b.GetAddressInfoCA("1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB")
	if err != nil {
		t.Error("Bitflyer - GetAddressInfoCA() error:", err)
	}
	if v.UnconfirmedBalance == 0 || v.ConfirmedBalance == 0 {
		log.Warn(log.ExchangeSys, "Donation wallet is empty :( - please consider donating")
	}
}

func TestGetMarkets(t *testing.T) {
	t.Parallel()
	_, err := b.GetMarkets()
	if err != nil {
		t.Error("Bitflyer - GetMarkets() error:", err)
	}
}

func TestGetOrderBook(t *testing.T) {
	t.Parallel()
	_, err := b.GetOrderBook("BTC_JPY")
	if err != nil {
		t.Error("Bitflyer - GetOrderBook() error:", err)
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := b.GetTicker("BTC_JPY")
	if err != nil {
		t.Error("Bitflyer - GetTicker() error:", err)
	}
}

func TestGetExecutionHistory(t *testing.T) {
	t.Parallel()
	_, err := b.GetExecutionHistory("BTC_JPY")
	if err != nil {
		t.Error("Bitflyer - GetExecutionHistory() error:", err)
	}
}

func TestGetExchangeStatus(t *testing.T) {
	t.Parallel()
	_, err := b.GetExchangeStatus()
	if err != nil {
		t.Error("Bitflyer - GetExchangeStatus() error:", err)
	}
}

func TestCheckFXString(t *testing.T) {
	t.Parallel()
	p := currency.NewPairDelimiter("FXBTC_JPY", "_")
	p = b.CheckFXString(p)
	if p.Base.String() != "FX_BTC" {
		t.Error("Bitflyer - CheckFXString() error")
	}
}

func TestFetchTicker(t *testing.T) {
	t.Parallel()
	var p currency.Pair

	currencies := b.GetAvailablePairs(asset.Spot)
	for _, pair := range currencies {
		if pair.String() == "FXBTC_JPY" {
			p = pair
			break
		}
	}

	_, err := b.FetchTicker(p, asset.Spot)
	if err != nil {
		t.Error("Bitflyer - FetchTicker() error", err)
	}
}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:              1,
		FeeType:             exchange.CryptocurrencyTradeFee,
		Pair:                currency.NewPair(currency.BTC, currency.LTC),
		PurchasePrice:       1,
		FiatCurrency:        currency.JPY,
		BankTransactionType: exchange.WireTransfer,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	var feeBuilder = setFeeBuilder()
	b.GetFeeByType(feeBuilder)
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
	b.SetDefaults()
	TestSetup(t)
	var feeBuilder = setFeeBuilder()

	if apiKey != "" || apiSecret != "" {
		// CryptocurrencyTradeFee Basic
		if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Error(err)
			t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		}

		// CryptocurrencyTradeFee High quantity
		feeBuilder = setFeeBuilder()
		feeBuilder.Amount = 1000
		feeBuilder.PurchasePrice = 1000
		if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee IsMaker
		feeBuilder = setFeeBuilder()
		feeBuilder.IsMaker = true
		if resp, err := b.GetFee(feeBuilder); resp != float64(0.1) || err != nil {
			t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0.1), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee Negative purchase price
		feeBuilder = setFeeBuilder()
		feeBuilder.PurchasePrice = -1000
		if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
			t.Error(err)
		}

		// CryptocurrencyWithdrawalFee Basic
		feeBuilder = setFeeBuilder()
		feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
		if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
			t.Error(err)
		}
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.FiatCurrency = currency.JPY
	if resp, err := b.GetFee(feeBuilder); resp != float64(324) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(324), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.JPY
	if resp, err := b.GetFee(feeBuilder); resp != float64(540) || err != nil {
		t.Errorf("GetFee() error. Expected: %f, Received: %f", float64(540), resp)
		t.Error(err)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	b.SetDefaults()
	expectedResult := exchange.AutoWithdrawFiatText + " & " + exchange.WithdrawCryptoViaWebsiteOnlyText

	withdrawPermissions := b.FormatWithdrawPermissions()

	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Received: %s", expectedResult, withdrawPermissions)
	}
}

func TestGetActiveOrders(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	var getOrdersRequest = exchange.GetOrdersRequest{
		OrderType: exchange.AnyOrderType,
	}

	_, err := b.GetActiveOrders(&getOrdersRequest)
	if areTestAPIKeysSet() && err != nil {
		t.Errorf("Could not get open orders: %s", err)
	} else if !areTestAPIKeysSet() && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

func TestGetOrderHistory(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	var getOrdersRequest = exchange.GetOrdersRequest{
		OrderType: exchange.AnyOrderType,
	}

	_, err := b.GetOrderHistory(&getOrdersRequest)
	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected '%v', received '%v'", common.ErrNotYetImplemented, err)
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------
func areTestAPIKeysSet() bool {
	return b.ValidateAPICredentials()
}

func TestSubmitOrder(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var orderSubmission = &exchange.OrderSubmission{
		Pair: currency.Pair{
			Base:  currency.BTC,
			Quote: currency.LTC,
		},
		OrderSide: exchange.BuyOrderSide,
		OrderType: exchange.LimitOrderType,
		Price:     1,
		Amount:    1,
		ClientID:  "meowOrder",
	}
	_, err := b.SubmitOrder(orderSubmission)
	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected 'Not Yet Implemented', received %v", err)
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)
	var orderCancellation = &exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	err := b.CancelOrder(orderCancellation)

	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected 'Not Yet Implemented', received %v", err)
	}
}

func TestCancelAllExchangeOrders(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)
	var orderCancellation = &exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	_, err := b.CancelAllOrders(orderCancellation)

	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected 'Not Yet Implemented', received %v", err)
	}
}

func TestWithdraw(t *testing.T) {
	b.SetDefaults()
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

	_, err := b.WithdrawCryptocurrencyFunds(&withdrawCryptoRequest)
	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected 'Not Yet Implemented', received %v", err)
	}
}

func TestModifyOrder(t *testing.T) {
	_, err := b.ModifyOrder(&exchange.ModifyOrder{})
	if err == nil {
		t.Error("ModifyOrder() Expected error")
	}
}

func TestWithdrawFiat(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = exchange.FiatWithdrawRequest{}

	_, err := b.WithdrawFiatFunds(&withdrawFiatRequest)
	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected '%v', received: '%v'", common.ErrNotYetImplemented, err)
	}
}

func TestWithdrawInternationalBank(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if areTestAPIKeysSet() && !canManipulateRealOrders {
		t.Skip("API keys set, canManipulateRealOrders false, skipping test")
	}

	var withdrawFiatRequest = exchange.FiatWithdrawRequest{}

	_, err := b.WithdrawFiatFundsToInternationalBank(&withdrawFiatRequest)
	if err != common.ErrNotYetImplemented {
		t.Errorf("Expected '%v', received: '%v'", common.ErrNotYetImplemented, err)
	}
}
