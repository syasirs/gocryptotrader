package bithumb

import (
	"context"
	"errors"
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

// Please supply your own keys here for due diligence testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = true
	testCurrency            = "btc"
)

var b = &Bithumb{}

func TestMain(m *testing.M) {
	b.SetDefaults()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal("Bithumb load config error", err)
	}
	bitConfig, err := cfg.GetExchangeConfig("Bithumb")
	if err != nil {
		log.Fatal("Bithumb Setup() init error")
	}

	bitConfig.API.AuthenticatedSupport = true
	if apiKey != "" {
		bitConfig.API.Credentials.Key = apiKey
	}
	if apiSecret != "" {
		bitConfig.API.Credentials.Secret = apiSecret
	}

	err = b.Setup(bitConfig)
	if err != nil {
		log.Fatal("Bithumb setup error", err)
	}

	err = b.UpdateTradablePairs(context.Background(), false)
	if err != nil {
		log.Fatal("Bithumb Setup() init error", err)
	}

	os.Exit(m.Run())
}

func TestGetTradablePairs(t *testing.T) {
	t.Parallel()
	_, err := b.GetTradablePairs(context.Background())
	require.NoError(t, err, "GetTradablePairs must not error")
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	tick, err := b.GetTicker(context.Background(), testCurrency)
	require.NoError(t, err, "GetTicker must not error")
	assert.Positive(t, tick.OpeningPrice, "OpeningPrice should be positive")
	assert.Positive(t, tick.ClosingPrice, "ClosingPrice should be positive")
	assert.Positive(t, tick.MinPrice, "MinPrice should be positive")
	assert.Positive(t, tick.MaxPrice, "MaxPrice should be positive")
	assert.Positive(t, tick.UnitsTraded, "UnitsTraded should be positive")
	assert.Positive(t, tick.AccumulatedTradeValue, "AccumulatedTradeValue should be positive")
	assert.Positive(t, tick.PreviousClosingPrice, "PreviousClosingPrice should be positive")
	assert.Positive(t, tick.UnitsTraded24Hr, "UnitsTraded24Hr should be positive")
	assert.Positive(t, tick.AccumulatedTradeValue24hr, "AccumulatedTradeValue24hr should be positive")
	assert.NotEmpty(t, tick.Fluctuate24Hr, "Fluctuate24Hr should not be empty")
	assert.NotEmpty(t, tick.FluctuateRate24hr, "FluctuateRate24hr should not be empty")
	assert.Positive(t, tick.Date, "Date should be positive")
}

// not all currencies have dates and fluctuation rates
func TestGetAllTickers(t *testing.T) {
	t.Parallel()
	tick, err := b.GetAllTickers(context.Background())
	require.NoError(t, err, "GetAllTickers must not error")
	for _, res := range tick {
		assert.Positive(t, res.OpeningPrice, "OpeningPrice should be positive")
		assert.Positive(t, res.ClosingPrice, "ClosingPrice should be positive")
		assert.Positive(t, res.MinPrice, "MinPrice should be positive")
		assert.Positive(t, res.MaxPrice, "MaxPrice should be positive")
		assert.Positive(t, res.UnitsTraded, "UnitsTraded should be positive")
		assert.Positive(t, res.AccumulatedTradeValue, "AccumulatedTradeValue should be positive")
		assert.Positive(t, res.PreviousClosingPrice, "PreviousClosingPrice should be positive")
		assert.Positive(t, res.UnitsTraded24Hr, "UnitsTraded24Hr should be positive")
		assert.Positive(t, res.AccumulatedTradeValue24hr, "AccumulatedTradeValue24hr should be positive")
	}
}

func TestGetOrderBook(t *testing.T) {
	t.Parallel()
	ob, err := b.GetOrderBook(context.Background(), testCurrency)
	require.NoError(t, err, "GetOrderBook must not error")
	assert.NotEmpty(t, ob.Status, "Status should not be empty")
	assert.NotEmpty(t, ob.Data.Timestamp, "Timestamp should not be empty")
	assert.NotEmpty(t, ob.Data.OrderCurrency, "OrderCurrency should not be empty")
	assert.NotEmpty(t, ob.Data.PaymentCurrency, "PaymentCurrency should not be empty")
	for _, a := range ob.Data.Asks {
		assert.Positive(t, a.Price, "Price should be positive")
		assert.Positive(t, a.Quantity, "Quantity should be positive")
	}
	for _, b := range ob.Data.Bids {
		assert.Positive(t, b.Price, "Price should be positive")
		assert.Positive(t, b.Quantity, "Quantity should be positive")
	}
}

func TestGetTransactionHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	th, err := b.GetTransactionHistory(context.Background(), testCurrency)
	require.NoError(t, err, "GetTransactionHistory must not error")
	assert.NotEmpty(t, th.Status, "Status should not be empty")
	for _, res := range th.Data {
		assert.Positive(t, res.ContNumber, "ContNumber should be positive")
		assert.NotEmpty(t, res.TransactionDate, "TransactionDate should not be empty")
		assert.Positive(t, res.ContNumber, "ContNumber should be positive")
		assert.NotEmpty(t, res.Type, "Type should not be empty")
		assert.Positive(t, res.UnitsTraded, "UnitsTraded should be positive")
		assert.Positive(t, res.Price, "Price should be positive")
		assert.Positive(t, res.Total, "Total should be positive")
	}
}

func TestGetAccountInformation(t *testing.T) {
	t.Parallel()

	// Offline test
	_, err := b.GetAccountInformation(context.Background(), "", "")
	assert.Error(t, err, "expected error when no order currency is specified")

	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	a, err := b.GetAccountInformation(context.Background(),
		testCurrency,
		currency.KRW.String())
	require.NoError(t, err, "GetAccountInformation should not error")
	assert.NotEmpty(t, a.Status, "Status should not be empty")
	assert.Positive(t, a.Data.Created, "Created should be positive")
	assert.NotEmpty(t, a.Data.AccountID, "AccountID should not be empty")
	assert.Positive(t, a.Data.TradeFee, "TradeFee should be positive")
	assert.Positive(t, a.Data.Balance, "Balance should be positive")
}

func TestGetAccountBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	b, err := b.GetAccountBalance(context.Background(), testCurrency)
	require.NoError(t, err, "GetAccountBalance must not error")
	assert.NotEmpty(t, b.Available, "Available should not be empty")
	assert.NotEmpty(t, b.InUse, "InUse should not be empty")
	assert.NotEmpty(t, b.Misu, "Misu should not be empty")
	assert.NotEmpty(t, b.Total, "Total should not be empty")
	assert.NotEmpty(t, b.Xcoin, "Xcoin should not be empty")
}

func TestGetWalletAddress(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	a, err := b.GetWalletAddress(context.Background(), currency.BTC)
	require.NoError(t, err, "GetWalletAddress must not error")
	assert.NotEmpty(t, a.Data.Currency, "Currency should not be empty")
	assert.NotEmpty(t, a.Data.Tag, "Tag should not be empty")
	assert.NotEmpty(t, a.Data.WalletAddress, "WalletAddress should not be empty")
}

func TestGetLastTransaction(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	tr, err := b.GetLastTransaction(context.Background())
	require.NoError(t, err, "GetLastTransaction must not error")
	assert.Positive(t, tr.Data.AveragePrice, "AveragePrice should not be positive")
	assert.Positive(t, tr.Data.BuyPrice, "BuyPrice should not be positive")
	assert.Positive(t, tr.Data.ClosingPrice, "ClosingPrice should not be positive")
	assert.NotEmpty(t, tr.Data.Date, "ClosingPrice should not be empty")
	assert.Positive(t, tr.Data.MaxPrice, "MaxPrice should not be positive")
	assert.Positive(t, tr.Data.MinPrice, "MinPrice should not be positive")
	assert.Positive(t, tr.Data.OpeningPrice, "OpeningPrice should not be positive")
	assert.Positive(t, tr.Data.SellPrice, "SellPrice should not be positive")
}

func TestGetOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetOrders(context.Background(),
		"1337", order.Bid.Lower(), 100, time.Time{}, currency.BTC, currency.KRW)
	require.NoError(t, err, "GetOrders must not error")
}

func TestGetUserTransactions(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetUserTransactions(context.Background(), 0, 0, 0, currency.EMPTYCODE, currency.EMPTYCODE)
	require.NoError(t, err, "GetUserTransactions must not error")
}

func TestPlaceTrade(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	_, err := b.PlaceTrade(context.Background(),
		testCurrency, order.Bid.Lower(), 0, 0)
	require.NoError(t, err, "PlaceTrade must not error")
}

func TestGetOrderDetails(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetOrderDetails(context.Background(),
		"1337", order.Bid.Lower(), testCurrency)
	require.NoError(t, err, "GetOrderDetails must not error")
}

func TestCancelTrade(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	_, err := b.CancelTrade(context.Background(), "", "", "")
	require.NoError(t, err, "CancelTrade must not error")
}

func TestWithdrawCrypto(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	_, err := b.WithdrawCrypto(context.Background(),
		"LQxiDhKU7idKiWQhx4ALKYkBx8xKEQVxJR", "", "ltc", 0)
	require.NoError(t, err, "WithdrawCrypto must not error")
}

func TestRequestKRWDepositDetails(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	_, err := b.RequestKRWDepositDetails(context.Background())
	require.NoError(t, err, "RequestKRWDepositDetails must not error")
}

func TestRequestKRWWithdraw(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	_, err := b.RequestKRWWithdraw(context.Background(),
		"102_bank", "1337", 1000)
	require.NoError(t, err, "RequestKRWWithdraw must not error")
}

func TestMarketBuyOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	p := currency.NewPair(currency.BTC, currency.KRW)
	_, err := b.MarketBuyOrder(context.Background(), p, 0)
	require.NoError(t, err, "MarketBuyOrder must not error")
}

func TestMarketSellOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	p := currency.NewPair(currency.BTC, currency.KRW)
	_, err := b.MarketSellOrder(context.Background(), p, 0)
	require.NoError(t, err, "MarketSellOrder must not error")
}

func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	cp := currency.NewPair(currency.QTUM, currency.KRW)
	_, err := b.UpdateTicker(context.Background(), cp, asset.Spot)
	require.NoError(t, err, "UpdateTicker must not error")
}

func TestUpdateTickers(t *testing.T) {
	t.Parallel()
	err := b.UpdateTickers(context.Background(), asset.Spot)
	require.NoError(t, err, "UpdateTickers must not error")
}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:        1,
		FeeType:       exchange.CryptocurrencyTradeFee,
		Pair:          currency.NewPair(currency.BTC, currency.LTC),
		PurchasePrice: 1,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	t.Parallel()
	var feeBuilder = setFeeBuilder()
	_, err := b.GetFeeByType(context.Background(), feeBuilder)
	require.NoError(t, err, "GetFeeByType must not error")

	if !sharedtestvalues.AreAPICredentialsSet(b) {
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
	t.Parallel()
	var feeBuilder = setFeeBuilder()
	// CryptocurrencyTradeFee Basic
	_, err := b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// CryptocurrencyTradeFee High quantity
	feeBuilder = setFeeBuilder()
	feeBuilder.Amount = 1000
	feeBuilder.PurchasePrice = 1000
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// CryptocurrencyTradeFee IsMaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = true
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// CryptocurrencyTradeFee Negative purchase price
	feeBuilder = setFeeBuilder()
	feeBuilder.PurchasePrice = -1000
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// CryptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyDepositFee
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.FiatCurrency = currency.HKD
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.HKD
	_, err = b.GetFee(feeBuilder)
	require.NoError(t, err, "GetFee must not error")
}

func TestFormatWithdrawPermissions(t *testing.T) {
	t.Parallel()
	expectedResult := exchange.AutoWithdrawCryptoText + " & " + exchange.AutoWithdrawFiatText
	withdrawPermissions := b.FormatWithdrawPermissions()
	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Received: %s", expectedResult, withdrawPermissions)
	}
}

func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		Side:      order.Sell,
		AssetType: asset.Spot,
	}

	_, err := b.GetActiveOrders(context.Background(), &getOrdersRequest)
	require.NoError(t, err, "GetActiveOrders must not error")
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Could not get open orders: %s", err)
	} else if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

func TestGetOrderHistory(t *testing.T) {
	t.Parallel()
	var getOrdersRequest = order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.AnySide,
		Pairs:     currency.Pairs{currency.NewPair(currency.BTC, currency.KRW)},
	}

	_, err := b.GetOrderHistory(context.Background(), &getOrdersRequest)
	require.NoError(t, err, "GetOrderHistory must not error")
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Could not get order history: %s", err)
	} else if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------

func TestSubmitOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

	var orderSubmission = &order.Submit{
		Exchange: b.Name,
		Pair: currency.Pair{
			Base:  currency.BTC,
			Quote: currency.LTC,
		},
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     1,
		Amount:    1,
		ClientID:  "meowOrder",
		AssetType: asset.Spot,
	}
	response, err := b.SubmitOrder(context.Background(), orderSubmission)
	require.NoError(t, err, "SubmitOrder must not error")
	if sharedtestvalues.AreAPICredentialsSet(b) && (err != nil || response.Status != order.New) {
		t.Errorf("Order failed to be placed: %v", err)
	} else if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)
	var orderCancellation = &order.Cancel{
		OrderID:       "1",
		WalletAddress: core.BitcoinDonationAddress,
		AccountID:     "1",
		Pair:          currencyPair,
		AssetType:     asset.Spot,
	}

	err := b.CancelOrder(context.Background(), orderCancellation)
	require.NoError(t, err, "CancelOrder must not error")
	if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Could not cancel order: %v", err)
	}
}

func TestCancelAllExchangeOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

	currencyPair := currency.NewPair(currency.LTC, currency.BTC)
	var orderCancellation = &order.Cancel{
		OrderID:       "1",
		WalletAddress: core.BitcoinDonationAddress,
		AccountID:     "1",
		Pair:          currencyPair,
		AssetType:     asset.Spot,
	}

	resp, err := b.CancelAllOrders(context.Background(), orderCancellation)
	require.NoError(t, err, "CancelAllOrders must not error")
	if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Could not cancel order: %v", err)
	}

	if len(resp.Status) > 0 {
		t.Errorf("%v orders failed to cancel", len(resp.Status))
	}
}

func TestGetAccountInfo(t *testing.T) {
	t.Parallel()
	if sharedtestvalues.AreAPICredentialsSet(b) {
		_, err := b.UpdateAccountInfo(context.Background(), asset.Spot)
		require.NoError(t, err, "UpdateAccountInfo must not error")
	} else {
		_, err := b.UpdateAccountInfo(context.Background(), asset.Spot)
		if err == nil {
			t.Error("Bithumb GetAccountInfo() Expected error")
		}
	}
}

func TestModifyOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)

	curr, err := currency.NewPairFromString("BTCUSD")
	require.NoError(t, err, "Must set the currency")

	_, err = b.ModifyOrder(context.Background(), &order.Modify{
		OrderID:   "1337",
		Price:     100,
		Amount:    1000,
		Side:      order.Sell,
		Pair:      curr,
		AssetType: asset.Spot,
	})
	require.NoError(t, err, "ModifyOrder must not error")
}

func TestWithdraw(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

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
	require.NoError(t, err, "WithdrawCryptocurrencyFunds must not error")

	if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

func TestWithdrawFiat(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

	var withdrawFiatRequest = withdraw.Request{
		Fiat: withdraw.FiatRequest{
			WireCurrency:             currency.KRW.String(),
			RequiresIntermediaryBank: false,
			IsExpressWire:            false,
		},
		Amount:      -1,
		Currency:    currency.USD,
		Description: "WITHDRAW IT ALL",
	}

	_, err := b.WithdrawFiatFunds(context.Background(), &withdrawFiatRequest)
	require.NoError(t, err, "WithdrawFiatFunds must not error")
	if !sharedtestvalues.AreAPICredentialsSet(b) && err == nil {
		t.Error("Expecting an error when no keys are set")
	}
	if sharedtestvalues.AreAPICredentialsSet(b) && err != nil {
		t.Errorf("Withdraw failed to be placed: %v", err)
	}
}

func TestWithdrawInternationalBank(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCannotManipulateOrders(t, b, canManipulateRealOrders)

	var withdrawFiatRequest = withdraw.Request{}
	_, err := b.WithdrawFiatFundsToInternationalBank(context.Background(),
		&withdrawFiatRequest)
	if err != common.ErrFunctionNotSupported {
		t.Errorf("Expected '%v', received: '%v'", common.ErrFunctionNotSupported, err)
	}
}

func TestGetDepositAddress(t *testing.T) {
	t.Parallel()
	if sharedtestvalues.AreAPICredentialsSet(b) {
		_, err := b.GetDepositAddress(context.Background(), currency.BTC, "", "")
		require.NoError(t, err, "GetDepositAddress must not error")
	} else {
		_, err := b.GetDepositAddress(context.Background(), currency.BTC, "", "")
		if err == nil {
			t.Error("GetDepositAddress() error cannot be nil")
		}
	}
}

func TestGetCandleStick(t *testing.T) {
	t.Parallel()
	_, err := b.GetCandleStick(context.Background(), "BTC_KRW", "1m")
	require.NoError(t, err, "GetCandleStick must not error")
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()
	pair, err := currency.NewPairFromString("BTCKRW")
	require.NoError(t, err, "Must set the currency")
	startTime := time.Now().AddDate(0, -1, 0)
	_, err = b.GetHistoricCandles(context.Background(), pair, asset.Spot, kline.OneDay, startTime, time.Now())
	require.NoError(t, err, "GetHistoricCandles must not error")
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	t.Parallel()
	pair, err := currency.NewPairFromString("BTCKRW")
	require.NoError(t, err, "Must set the currency")

	startTime := time.Now().Add(-time.Hour * 24)
	_, err = b.GetHistoricCandlesExtended(context.Background(), pair, asset.Spot, kline.OneDay, startTime, time.Now())
	if !errors.Is(err, common.ErrFunctionNotSupported) {
		t.Fatal(err)
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("BTC_KRW")
	require.NoError(t, err, "Must set the currency")

	_, err = b.GetRecentTrades(context.Background(), currencyPair, asset.Spot)
	require.NoError(t, err, "GetRecentTrades must not error")
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("BTC_KRW")
	require.NoError(t, err, "Must set the currency")

	_, err = b.GetHistoricTrades(context.Background(),
		currencyPair, asset.Spot, time.Now().Add(-time.Minute*15), time.Now())
	if err != nil && err != common.ErrFunctionNotSupported {
		t.Error(err)
	}
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()
	err := b.UpdateOrderExecutionLimits(context.Background(), asset.Empty)
	require.NoError(t, err, "UpdateOrderExecutionLimits must not error")

	cp := currency.NewPair(currency.BTC, currency.KRW)
	limit, err := b.GetOrderExecutionLimits(asset.Spot, cp)
	require.NoError(t, err, "GetOrderExecutionLimits must not error")

	err = limit.Conforms(46241000, 0.00001, order.Limit)
	if !errors.Is(err, order.ErrAmountBelowMin) {
		t.Fatalf("expected error %v but received %v",
			order.ErrAmountBelowMin,
			err)
	}

	err = limit.Conforms(46241000, 0.0001, order.Limit)
	if !errors.Is(err, nil) {
		t.Fatalf("expected error %v but received %v",
			nil,
			err)
	}
}

func TestGetAmountMinimum(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		unitprice float64
		expected  float64
	}{
		{
			name:      "ETH-KRW",
			unitprice: 2638000.0,
			expected:  0.0002,
		},
		{
			name:      "DOGE-KRW",
			unitprice: 236.5,
			expected:  2.1142,
		},
		{
			name:      "XRP-KRW",
			unitprice: 818.8,
			expected:  0.6107,
		},
		{
			name:      "LTC-KRW",
			unitprice: 160100,
			expected:  0.0032,
		},
		{
			name:      "BTC-KRW",
			unitprice: 46079000,
			expected:  0.0001,
		},
		{
			name:      "nonsense",
			unitprice: 0,
			expected:  0,
		},
	}

	for i := range testCases {
		tt := &testCases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			minAmount := getAmountMinimum(tt.unitprice)
			if minAmount != tt.expected {
				t.Fatalf("expected: %f but received: %f for unit price: %f",
					tt.expected,
					minAmount,
					tt.unitprice)
			}
		})
	}
}

func TestGetAssetStatus(t *testing.T) {
	t.Parallel()
	_, err := b.GetAssetStatus(context.Background(), "")
	if !errors.Is(err, errSymbolIsEmpty) {
		t.Fatalf("received: %v but expected: %v", err, errSymbolIsEmpty)
	}

	_, err = b.GetAssetStatus(context.Background(), "sol")
	require.NoError(t, err, "GetAssetStatus must not error")
}

func TestGetAssetStatusAll(t *testing.T) {
	t.Parallel()
	_, err := b.GetAssetStatusAll(context.Background())
	require.NoError(t, err, "GetAssetStatusAll must not error")
}

func TestUpdateCurrencyStates(t *testing.T) {
	t.Parallel()
	err := b.UpdateCurrencyStates(context.Background(), asset.Spot)
	require.NoError(t, err, "UpdateCurrencyStates must not error")
}

func TestGetWithdrawalsHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetWithdrawalsHistory(context.Background(), currency.BTC, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetOrderInfo(context.Background(), "1234", currency.NewPair(currency.BTC, currency.USDT), asset.Spot)
	require.NoError(t, err, "GetOrderInfo must not error")
}

func TestGetWithdrawalHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, b)

	_, err := b.GetWithdrawalsHistory(context.Background(), currency.BTC, asset.Spot)
	require.NoError(t, err, "GetWithdrawalsHistory must not error")
}
