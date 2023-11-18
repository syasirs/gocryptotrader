package bitstamp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/core"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/portfolio/banking"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// Please add your private keys and customerID for better tests
const (
	apiKey                  = ""
	apiSecret               = ""
	customerID              = "" // This is the customer id you use to log in
	canManipulateRealOrders = true
)

var b = &Bitstamp{}

func setFeeBuilder() *exchange.FeeBuilder {
	return &exchange.FeeBuilder{
		Amount:        5,
		FeeType:       exchange.CryptocurrencyTradeFee,
		Pair:          currency.NewPair(currency.LTC, currency.BTC),
		PurchasePrice: 1800,
	}
}

// TestGetFeeByTypeOfflineTradeFee logic test
func TestGetFeeByTypeOfflineTradeFee(t *testing.T) {
	t.Parallel()

	var feeBuilder = setFeeBuilder()
	_, err := b.GetFeeByType(context.Background(), feeBuilder)
	if err != nil {
		t.Fatal(err)
	}
	if !sharedtestvalues.AreAPICredentialsSet(b) {
		if feeBuilder.FeeType != exchange.OfflineTradeFee {
			t.Errorf("Expected %v, received %v",
				exchange.OfflineTradeFee,
				feeBuilder.FeeType)
		}
	} else {
		if feeBuilder.FeeType != exchange.CryptocurrencyTradeFee {
			t.Errorf("Expected %v, received %v",
				exchange.CryptocurrencyTradeFee,
				feeBuilder.FeeType)
		}
	}
}

func TestGetFee(t *testing.T) {
	t.Parallel()

	var feeBuilder = setFeeBuilder()

	// CryptocurrencyTradeFee Basic
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// CryptocurrencyTradeFee High quantity
	feeBuilder = setFeeBuilder()
	feeBuilder.Amount = 1000
	feeBuilder.PurchasePrice = 1000
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// CryptocurrencyTradeFee IsMaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = true
	fee, err := b.GetFee(context.Background(), feeBuilder)
	fee, err = b.GetFee(context.Background(), feeBuilder)
	if assert.NoError(t, err, "GetFee should not error") {
		assert.NotEmpty(t, feeBuilder.Pair, "Pair should not be empty")
		if mockTests {
			assert.Positive(t, fee, "Maker fee should be positive")
		}
	}

	// CryptocurrencyTradeFee IsTaker
	feeBuilder = setFeeBuilder()
	feeBuilder.IsMaker = false
	fee, err = b.GetFee(context.Background(), feeBuilder)
	if assert.NoError(t, err, "GetFee should not error") {
		assert.NotEmpty(t, feeBuilder.Pair, "Pair should not be empty")
		if mockTests {
			assert.Positive(t, fee, "Taker fee should be positive")
		}
	}

	// CryptocurrencyTradeFee Negative purchase price
	feeBuilder = setFeeBuilder()
	feeBuilder.PurchasePrice = -1000
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// CryptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyDepositFee
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.FiatCurrency = currency.HKD
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.FiatCurrency = currency.HKD
	if _, err := b.GetFee(context.Background(), feeBuilder); err != nil {
		t.Error(err)
	}
}

func TestGetAccountTradingFee(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}

	fee, err := b.GetAccountTradingFee(context.Background(), currency.NewPair(currency.LTC, currency.BTC))
	if assert.NoError(t, err, "GetAccountTradingFee should not error") {
		if mockTests {
			assert.Positive(t, fee.Fees.Maker, "Maker should be positive")
			assert.Positive(t, fee.Fees.Taker, "Taker should be positive")
		}
		assert.NotEmpty(t, fee.Symbol, "Symbol should not be empty")
		assert.Equal(t, fee.Symbol, "ltcbtc", "Symbol should be correct")
	}

	_, err = b.GetAccountTradingFee(context.Background(), currency.EMPTYPAIR)
	assert.ErrorIs(t, err, currency.ErrCurrencyPairEmpty, "Should get back the right error")
}

func TestGetAccountTradingFees(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}

	fees, err := b.GetAccountTradingFees(context.Background())
	if assert.NoError(t, err, "GetAccountTradingFee should not error") {
		if assert.Positive(t, len(fees), "Should get back multiple fees") {
			fee := fees[0]
			assert.NotEmpty(t, fee.Symbol, "Should get back a symbol")
			if mockTests {
				assert.Positive(t, fee.Fees.Maker, "Maker should be positive")
				assert.Positive(t, fee.Fees.Taker, "Taker should be positive")
			}
		}
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()

	_, err := b.GetTicker(context.Background(),
		currency.BTC.String()+currency.USD.String(), false)
	if err != nil {
		t.Error("GetTicker() error", err)
	}
}

func TestGetOrderbook(t *testing.T) {
	t.Parallel()
	_, err := b.GetOrderbook(context.Background(),
		currency.BTC.String()+currency.USD.String())
	if err != nil {
		t.Error("GetOrderbook() error", err)
	}
}

func TestGetTradingPairs(t *testing.T) {
	t.Parallel()

	_, err := b.GetTradingPairs(context.Background())
	if err != nil {
		t.Error("GetTradingPairs() error", err)
	}
}

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	pairs, err := b.FetchTradablePairs(context.Background(), asset.Spot)
	if err != nil {
		t.Fatal(err)
	}
	if !pairs.Contains(currency.NewPair(currency.COMP, currency.USD), false) {
		t.Error("expected pair COMP/USD")
	}
	if !pairs.Contains(currency.NewPair(currency.BTC, currency.USD), false) {
		t.Error("expected pair BTC/USD")
	}
	if !pairs.Contains(currency.NewPair(currency.USDC, currency.USDT), false) {
		t.Error("expected pair USDC/USDT")
	}
}

func TestUpdateTradablePairs(t *testing.T) {
	t.Parallel()
	if err := b.UpdateTradablePairs(context.Background(), true); err != nil {
		t.Error("Bitstamp UpdateTradablePairs() error", err)
	}
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	t.Parallel()
	type limitTest struct {
		pair currency.Pair
		step float64
		min  float64
	}

	tests := map[asset.Item][]limitTest{
		asset.Spot: {
			{currency.NewPair(currency.ETH, currency.USDT), 0.01, 20},
			{currency.NewPair(currency.BTC, currency.USDT), 0.01, 20},
		},
	}
	for assetItem, limitTests := range tests {
		if err := b.UpdateOrderExecutionLimits(context.Background(), assetItem); err != nil {
			t.Errorf("Error fetching %s pairs for test: %v", assetItem, err)
		}
		for _, limitTest := range limitTests {
			limits, err := b.GetOrderExecutionLimits(assetItem, limitTest.pair)
			if err != nil {
				t.Errorf("Bitstamp GetOrderExecutionLimits() error during TestExecutionLimits; Asset: %s Pair: %s Err: %v", assetItem, limitTest.pair, err)
				continue
			}
			assert.NotEmpty(t, limits.Pair, "Pair should not be empty")
			assert.Positive(t, limits.PriceStepIncrementSize, "PriceStepIncrementSize should be positive")
			assert.Positive(t, limits.AmountStepIncrementSize, "AmountStepIncrementSize should be positive")
			assert.Positive(t, limits.MinimumQuoteAmount, "MinAmount should be positive")
			if mockTests {
				if got := limits.PriceStepIncrementSize; got != limitTest.step {
					t.Errorf("Bitstamp UpdateOrderExecutionLimits wrong PriceStepIncrementSize; Asset: %s Pair: %s Expected: %v Got: %v", assetItem, limitTest.pair, limitTest.step, got)
				}
				if got := limits.MinimumQuoteAmount; got != limitTest.min {
					t.Errorf("Bitstamp UpdateOrderExecutionLimits wrong MinAmount; Pair: %s Expected: %v Got: %v", limitTest.pair, limitTest.min, got)
				}
			}
		}
	}
}

func TestGetTransactions(t *testing.T) {
	t.Parallel()
	_, err := b.GetTransactions(context.Background(),
		currency.BTC.String()+currency.USD.String(), "hour")
	if err != nil {
		t.Error("GetTransactions() error", err)
	}
}

func TestGetEURUSDConversionRate(t *testing.T) {
	t.Parallel()

	_, err := b.GetEURUSDConversionRate(context.Background())
	if err != nil {
		t.Error("GetEURUSDConversionRate() error", err)
	}
}

func TestGetBalance(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	bal, err := b.GetBalance(context.Background())
	if mockTests {
		for k, e := range map[string]Balance{
			"USDT": {
				Available:     42.42,
				Balance:       1337.42,
				Reserved:      1295.00,
				WithdrawalFee: 5.0,
			},
			"BTC": {
				Available:     9.1,
				Balance:       11.2,
				Reserved:      2.1,
				WithdrawalFee: 0.00050000,
			},
		} {
			assert.Equal(t, e.Available, bal[k].Available, "Available balance should match")
			assert.Equal(t, e.Balance, bal[k].Balance, "Balance should match")
			assert.Equal(t, e.Reserved, bal[k].Reserved, "Reserved balance should match")
			assert.Equal(t, e.WithdrawalFee, bal[k].WithdrawalFee, "WithdraawlFee should match")
		}
	}
	assert.NoError(t, err, "TestGetBalance should not error")
}

func TestGetUserTransactions(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetUserTransactions(context.Background(), "btcusd")
	assert.NoError(t, err, "TestGetUserTransactions should not error")
}

func TestGetOpenOrders(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetOpenOrders(context.Background(), "btcusd")
	assert.NoError(t, err, "TestGetOpenOrders should not error")
}

// Failing Test MatchAndGetResponse
func TestGetOrderStatus(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	o, err := b.GetOrderStatus(context.Background(), 1458532827766784)
	if !mockTests {
		assert.ErrorContains(t, err, "Order not found")
	} else {
		assert.NoError(t, err, "TestGetOrderStatus should not error")
		assert.Equal(t, "2022-01-31 14:43:15", o.DateTime, "order date should match")
		assert.Equal(t, 1458532827766784, o.ID, "order ID should match")
		assert.Equal(t, "100.00", o.Amount, "amount should match")
		assert.Equal(t, "100.00", o.Price, "amount should match")
		assert.Equal(t, "0", o.Type, "amount should match")
	}
}

func TestGetWithdrawalRequests(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	r, err := b.GetWithdrawalRequests(context.Background(), 1)
	assert.NoError(t, err, "GetWithdrawalRequests should not error")
	if mockTests {
		assert.NotEmpty(t, r, "GetWithdrawalRequests should return a withdrawal request")
		for _, req := range r {
			assert.Equal(t, int64(1), req.OrderID, "OrderId should match")
			assert.Equal(t, "aMDHooGmAkyrsaQiKhAORhSNTmoRzxqWIO", req.Address, "Address should match")
			assert.Equal(t, "2022-01-31 16:07:32", req.Date, "Date should match")
			assert.Equal(t, currency.BTC, req.Currency, "Currency should match")
			assert.Equal(t, 0.00006000, req.Amount, "Amount should match")
			assert.Equal(t, "NsOeFbQhRnpGzNIThWGBTkQwRJqTNOGPVhYavrVyMfkAyMUmIlUpFIwGTzSvpeOP", req.TransactionID, "TransactionID should match")
			assert.Equal(t, int64(2), req.Status, "Status should match")
			assert.Equal(t, int64(0), req.Type, "Type should match")
		}
	}
}

func TestGetUnconfirmedBitcoinDeposits(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetUnconfirmedBitcoinDeposits(context.Background())
	assert.NoError(t, err, "TestGetUnconfirmedBitcoinDeposits should not error")

}

// Failing no documentation on what data to be used
func TestTransferAccountBalance(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	err := b.TransferAccountBalance(context.Background(),
		0.01, "btc", "testAccount", true)
	assert.NoError(t, err, "TestTransferAccountBalance should not error")
}

func TestFormatWithdrawPermissions(t *testing.T) {
	t.Parallel()

	expectedResult := exchange.AutoWithdrawCryptoText +
		" & " +
		exchange.AutoWithdrawFiatText
	withdrawPermissions := b.FormatWithdrawPermissions()
	assert.Equal(t, expectedResult, withdrawPermissions, "Permissions should be the same")
}

func TestGetActiveOrders(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetActiveOrders(context.Background(), &order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.AnySide,
	})
	assert.NoError(t, err, "TestGetActiveOrders should not error")
}

func TestGetOrderHistory(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b)
	}
	_, err := b.GetOrderHistory(context.Background(), &order.MultiOrderRequest{
		Type:      order.AnyType,
		AssetType: asset.Spot,
		Side:      order.AnySide,
	})
	assert.NoError(t, err, "TestGetOrderHistory should not error")
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------
// Failing Test MatchAndGetResponse
func TestSubmitOrder(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}
	_, err := b.SubmitOrder(context.Background(), &order.Submit{
		Exchange: b.Name,
		Pair: currency.Pair{
			Base:  currency.BTC,
			Quote: currency.USD,
		},
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     2211.00,
		Amount:    45,
		ClientID:  "123456789",
		AssetType: asset.Spot,
	})
	if !mockTests {
		assert.ErrorContains(t, err, "You need 10 USD to open that order. You have only 0 USD available. Check your account balance for details.")
	} else {
		assert.NoError(t, err, "TestSubmitOrder should not error")
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}
	err := b.CancelOrder(context.Background(), &order.Cancel{
		OrderID: "1453282316578816",
	})
	if !mockTests {
		assert.ErrorContains(t, err, "Order not found")
	} else {
		assert.NoError(t, err, "TestCancelExchangeOrder should not error")
	}
}

func TestCancelAllExchangeOrders(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}
	resp, err := b.CancelAllOrders(context.Background(),
		&order.Cancel{AssetType: asset.Spot})
	assert.NoError(t, err, "TestCancelAllExchangeOrders should not error")
	if len(resp.Status) > 0 {
		t.Errorf("%v orders failed to cancel", len(resp.Status))
	}
}

func TestModifyOrder(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}
	_, err := b.ModifyOrder(context.Background(), &order.Modify{AssetType: asset.Spot})
	assert.ErrorIs(t, err, common.ErrFunctionNotSupported)
}

// TODO Live test fix
func TestWithdraw(t *testing.T) {
	t.Parallel()
	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}
	_, err := b.WithdrawCryptocurrencyFunds(context.Background(), &withdraw.Request{
		Exchange:    b.Name,
		Amount:      6,
		Currency:    currency.BTC,
		Description: "WITHDRAW IT ALL",
		Crypto: withdraw.CryptoRequest{
			Address: core.BitcoinDonationAddress,
		},
	})
	assert.NoError(t, err, "TestWithdraw should not error")

}

func TestWithdrawFiat(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}

	var withdrawFiatRequest = withdraw.Request{
		Fiat: withdraw.FiatRequest{
			Bank: banking.Account{
				SupportedExchanges:  b.Name,
				Enabled:             true,
				AccountName:         "Satoshi Nakamoto",
				AccountNumber:       "12345",
				BankAddress:         "123 Fake St",
				BankPostalCity:      "Tarry Town",
				BankCountry:         "AU",
				BankName:            "Federal Reserve Bank",
				SWIFTCode:           "CTBAAU2S",
				BankPostalCode:      "2088",
				IBAN:                "IT60X0542811101000000123456",
				SupportedCurrencies: "USD",
			},
			WireCurrency:             currency.USD.String(),
			RequiresIntermediaryBank: false,
			IsExpressWire:            false,
		},
		Amount:      10,
		Currency:    currency.USD,
		Description: "WITHDRAW IT ALL",
	}

	_, err := b.WithdrawFiatFunds(context.Background(), &withdrawFiatRequest)
	if mockTests {
		assert.Equal(t, float64(10), withdrawFiatRequest.Amount, "Amount should match")
		assert.Equal(t, currency.Code(currency.USD), withdrawFiatRequest.Currency, "Currency should match")
		assert.NotEmpty(t, withdrawFiatRequest.Fiat.Bank, "Bank details must not be empty")
		assert.NotEmpty(t, withdrawFiatRequest.Fiat.WireCurrency, "WireCurrency details should not be empty")
		assert.Equal(t, currency.USD.String(), withdrawFiatRequest.Fiat.WireCurrency, "WireCurrency should match")
	} else {
		assert.ErrorContains(t, err, "amount: [You have only 0.00 USD available. Check your account balance for details.]")
	}
}

func TestWithdrawInternationalBank(t *testing.T) {
	t.Parallel()

	if !mockTests {
		sharedtestvalues.SkipTestIfCredentialsUnset(t, b, canManipulateRealOrders)
	}

	var withdrawFiatRequest = withdraw.Request{
		Fiat: withdraw.FiatRequest{
			Bank: banking.Account{
				SupportedExchanges:  b.Name,
				Enabled:             true,
				AccountName:         "Satoshi Nakamoto",
				AccountNumber:       "12345",
				BankAddress:         "123 Fake St",
				BankPostalCity:      "Tarry Town",
				BankCountry:         "AU",
				BankName:            "Federal Reserve Bank",
				SWIFTCode:           "CTBAAU2S",
				BankPostalCode:      "2088",
				IBAN:                "IT60X0542811101000000123456",
				SupportedCurrencies: "USD",
			},
			WireCurrency:                  currency.USD.String(),
			RequiresIntermediaryBank:      false,
			IsExpressWire:                 false,
			IntermediaryBankAccountNumber: 12345,
			IntermediaryBankAddress:       "123 Fake St",
			IntermediaryBankCity:          "Tarry Town",
			IntermediaryBankCountry:       "AU",
			IntermediaryBankName:          "Federal Reserve Bank",
			IntermediaryBankPostalCode:    "2088",
		},
		Amount:      50,
		Currency:    currency.USD,
		Description: "WITHDRAW IT ALL",
	}

	_, err := b.WithdrawFiatFundsToInternationalBank(context.Background(),
		&withdrawFiatRequest)
	if mockTests {
		assert.Equal(t, float64(50), withdrawFiatRequest.Amount, "Amount should match")
		assert.Equal(t, currency.USD, withdrawFiatRequest.Currency, "Currency should match")
		assert.NotEmpty(t, withdrawFiatRequest.Fiat.Bank, "Bank details must not be empty")
		assert.NotEmpty(t, withdrawFiatRequest.Fiat.WireCurrency, "WireCurrency details should not be empty")
	}
	assert.ErrorContains(t, err, "amount: [You have only 0.00 USD available. Check your account balance for details.]")
}

func TestGetDepositAddress(t *testing.T) {
	t.Parallel()

	_, err := b.GetDepositAddress(context.Background(), currency.XRP, "", "")
	switch {
	case sharedtestvalues.AreAPICredentialsSet(b) && customerID != "" && err != nil && !mockTests:
		t.Error("GetDepositAddress error", err)
	case !sharedtestvalues.AreAPICredentialsSet(b) && err == nil && !mockTests:
		t.Error("GetDepositAddress error cannot be nil")
	case mockTests && err != nil:
		t.Error("GetDepositAddress error", err)
	}
}

func TestParseTime(t *testing.T) {
	t.Parallel()

	tm, err := parseTime("2019-10-18 01:55:14")
	if err != nil {
		t.Error(err)
	}

	if tm.Year() != 2019 ||
		tm.Month() != 10 ||
		tm.Day() != 18 ||
		tm.Hour() != 1 ||
		tm.Minute() != 55 ||
		tm.Second() != 14 {
		t.Error("invalid time values")
	}
}

func TestWsSubscription(t *testing.T) {
	pressXToJSON := []byte(`{
		"event": "bts:subscribe",
		"data": {
			"channel": "[channel_name]"
		}
	}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsUnsubscribe(t *testing.T) {
	pressXToJSON := []byte(`{
		"event": "bts:subscribe",
		"data": {
			"channel": "[channel_name]"
		}
	}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsTrade(t *testing.T) {
	pressXToJSON := []byte(`{"data": {"microtimestamp": "1580336751488517", "amount": 0.00598803, "buy_order_id": 4621328909, "sell_order_id": 4621329035, "amount_str": "0.00598803", "price_str": "9334.73", "timestamp": "1580336751", "price": 9334.73, "type": 1, "id": 104007706}, "event": "trade", "channel": "live_trades_btcusd"}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsOrderbook(t *testing.T) {
	pressXToJSON := []byte(`{"data": {"timestamp": "1580336834", "microtimestamp": "1580336834607546", "bids": [["9328.28", "0.05925332"], ["9327.34", "0.43120000"], ["9327.29", "0.63470860"], ["9326.59", "0.41114619"], ["9326.38", "1.06910000"], ["9323.91", "2.67930000"], ["9322.69", "0.80000000"], ["9322.57", "0.03000000"], ["9322.31", "1.36010820"], ["9319.54", "0.03090000"], ["9318.97", "0.28000000"], ["9317.61", "0.02910000"], ["9316.39", "1.08000000"], ["9316.20", "2.00000000"], ["9315.48", "1.00000000"], ["9314.72", "0.11197459"], ["9314.47", "0.32207398"], ["9312.53", "0.03961501"], ["9312.29", "1.00000000"], ["9311.78", "0.03060000"], ["9311.69", "0.32217221"], ["9310.98", "3.29000000"], ["9310.18", "0.01304192"], ["9310.13", "0.02500000"], ["9309.04", "1.00000000"], ["9309.00", "0.05000000"], ["9308.96", "0.03030000"], ["9308.91", "0.32227154"], ["9307.52", "0.32191362"], ["9307.25", "2.44280000"], ["9305.92", "3.00000000"], ["9305.62", "2.37600000"], ["9305.60", "0.21815312"], ["9305.54", "2.80000000"], ["9305.13", "0.05000000"], ["9305.02", "2.90917302"], ["9303.68", "0.02316372"], ["9303.53", "12.55000000"], ["9303.00", "0.02191430"], ["9302.94", "2.38250000"], ["9302.37", "0.01000000"], ["9301.85", "2.50000000"], ["9300.89", "0.02000000"], ["9300.40", "4.10000000"], ["9300.00", "0.33936139"], ["9298.48", "1.45200000"], ["9297.80", "0.42380000"], ["9295.44", "4.54689328"], ["9295.43", "3.20000000"], ["9295.00", "0.28669566"], ["9291.66", "14.09931321"], ["9290.13", "2.87254900"], ["9290.00", "0.67530840"], ["9285.37", "0.38033002"], ["9285.15", "5.37993528"], ["9285.00", "0.09419278"], ["9283.71", "0.15679830"], ["9280.33", "12.55000000"], ["9280.13", "3.20310000"], ["9280.00", "1.36477909"], ["9276.01", "0.00707488"], ["9275.75", "0.56974291"], ["9275.00", "5.88000000"], ["9274.00", "0.00754205"], ["9271.68", "0.01400000"], ["9271.11", "15.37188500"], ["9270.00", "0.06674325"], ["9268.79", "24.54320000"], ["9257.18", "12.55000000"], ["9256.30", "0.17876365"], ["9255.71", "13.82642967"], ["9254.79", "0.96329407"], ["9250.00", "0.78214958"], ["9245.34", "4.90200000"], ["9245.13", "0.10000000"], ["9240.00", "0.44383459"], ["9238.84", "13.16615207"], ["9234.11", "0.43317656"], ["9234.10", "12.55000000"], ["9231.28", "11.79290000"], ["9230.09", "4.15059441"], ["9227.69", "0.00791097"], ["9225.00", "0.44768346"], ["9224.49", "0.85857203"], ["9223.50", "5.61001041"], ["9216.01", "0.03222653"], ["9216.00", "0.05000000"], ["9213.54", "0.71253866"], ["9212.50", "2.86768195"], ["9211.07", "12.55000000"], ["9210.00", "0.54288817"], ["9208.00", "1.00000000"], ["9206.06", "2.62587578"], ["9205.98", "15.40000000"], ["9205.52", "0.01710603"], ["9205.37", "0.03524953"], ["9205.11", "0.15000000"], ["9205.00", "0.01534763"], ["9204.76", "7.00600000"], ["9203.00", "0.01090000"]], "asks": [["9337.10", "0.03000000"], ["9340.85", "2.67820000"], ["9340.95", "0.02900000"], ["9341.17", "1.00000000"], ["9341.41", "2.13966390"], ["9341.61", "0.20000000"], ["9341.97", "0.11199911"], ["9341.98", "3.00000000"], ["9342.26", "0.32112762"], ["9343.87", "1.00000000"], ["9344.17", "3.57250000"], ["9345.04", "0.32103450"], ["9345.41", "4.90000000"], ["9345.69", "1.03000000"], ["9345.80", "0.03000000"], ["9346.00", "0.10200000"], ["9346.69", "0.02397394"], ["9347.41", "1.00000000"], ["9347.82", "0.32094177"], ["9348.23", "0.02880000"], ["9348.62", "11.96287551"], ["9349.31", "2.44270000"], ["9349.47", "0.96000000"], ["9349.86", "4.50000000"], ["9350.37", "0.03300000"], ["9350.57", "0.34682266"], ["9350.60", "0.32085527"], ["9351.45", "0.31147923"], ["9352.31", "0.28000000"], ["9352.86", "9.80000000"], ["9353.73", "0.02360739"], ["9354.00", "0.45000000"], ["9354.12", "0.03000000"], ["9354.29", "3.82446861"], ["9356.20", "0.64000000"], ["9356.90", "0.02316372"], ["9357.30", "2.50000000"], ["9357.70", "2.38240000"], ["9358.92", "6.00000000"], ["9359.97", "0.34898075"], ["9359.98", "2.30000000"], ["9362.56", "2.37600000"], ["9365.00", "0.64000000"], ["9365.16", "1.70030306"], ["9365.27", "3.03000000"], ["9369.99", "2.47102665"], ["9370.00", "3.15688574"], ["9370.21", "2.32720000"], ["9371.78", "13.20000000"], ["9371.89", "0.96293482"], ["9375.08", "4.74762500"], ["9384.34", "1.45200000"], ["9384.49", "16.42310000"], ["9385.66", "0.34382112"], ["9388.19", "0.00268265"], ["9392.20", "0.20980000"], ["9392.40", "0.10320000"], ["9393.00", "0.20980000"], ["9395.40", "0.40000000"], ["9398.86", "24.54310000"], ["9400.00", "0.05489988"], ["9400.33", "0.00495100"], ["9400.45", "0.00484700"], ["9402.92", "17.20000000"], ["9404.18", "10.00000000"], ["9418.89", "16.38000000"], ["9419.41", "3.06700000"], ["9420.40", "12.50000000"], ["9421.11", "0.10500000"], ["9434.47", "0.03215805"], ["9434.48", "0.28285714"], ["9434.49", "15.83000000"], ["9435.13", "0.15000000"], ["9438.93", "0.00368800"], ["9439.19", "0.69343985"], ["9442.86", "0.10000000"], ["9443.96", "12.50000000"], ["9444.00", "0.06004471"], ["9444.97", "0.01494896"], ["9447.00", "0.01234000"], ["9448.97", "0.14500000"], ["9449.00", "0.05000000"], ["9450.00", "11.13426018"], ["9451.87", "15.90000000"], ["9452.00", "0.20000000"], ["9454.25", "0.01100000"], ["9454.51", "0.02409062"], ["9455.05", "0.00600063"], ["9456.00", "0.27965118"], ["9456.10", "0.17000000"], ["9459.00", "0.00320000"], ["9459.98", "0.02460685"], ["9459.99", "8.11000000"], ["9460.00", "0.08500000"], ["9464.36", "0.56957951"], ["9464.54", "0.69158059"], ["9465.00", "21.00002015"], ["9467.57", "12.50000000"], ["9468.00", "0.08800000"], ["9469.09", "13.94000000"]]}, "event": "data", "channel": "order_book_btcusd"}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
	pressXToJSON = []byte(`{"data": {"timestamp": "1580336834", "microtimestamp": "1580336834607546", "bids": [["9328.28", "0.05925332"], ["9327.34", "0.43120000"], ["9327.29", "0.63470860"], ["9326.59", "0.41114619"], ["9326.38", "1.06910000"], ["9323.91", "2.67930000"], ["9322.69", "0.80000000"], ["9322.57", "0.03000000"], ["9322.31", "1.36010820"], ["9319.54", "0.03090000"], ["9318.97", "0.28000000"], ["9317.61", "0.02910000"], ["9316.39", "1.08000000"], ["9316.20", "2.00000000"], ["9315.48", "1.00000000"], ["9314.72", "0.11197459"], ["9314.47", "0.32207398"], ["9312.53", "0.03961501"], ["9312.29", "1.00000000"], ["9311.78", "0.03060000"], ["9311.69", "0.32217221"], ["9310.98", "3.29000000"], ["9310.18", "0.01304192"], ["9310.13", "0.02500000"], ["9309.04", "1.00000000"], ["9309.00", "0.05000000"], ["9308.96", "0.03030000"], ["9308.91", "0.32227154"], ["9307.52", "0.32191362"], ["9307.25", "2.44280000"], ["9305.92", "3.00000000"], ["9305.62", "2.37600000"], ["9305.60", "0.21815312"], ["9305.54", "2.80000000"], ["9305.13", "0.05000000"], ["9305.02", "2.90917302"], ["9303.68", "0.02316372"], ["9303.53", "12.55000000"], ["9303.00", "0.02191430"], ["9302.94", "2.38250000"], ["9302.37", "0.01000000"], ["9301.85", "2.50000000"], ["9300.89", "0.02000000"], ["9300.40", "4.10000000"], ["9300.00", "0.33936139"], ["9298.48", "1.45200000"], ["9297.80", "0.42380000"], ["9295.44", "4.54689328"], ["9295.43", "3.20000000"], ["9295.00", "0.28669566"], ["9291.66", "14.09931321"], ["9290.13", "2.87254900"], ["9290.00", "0.67530840"], ["9285.37", "0.38033002"], ["9285.15", "5.37993528"], ["9285.00", "0.09419278"], ["9283.71", "0.15679830"], ["9280.33", "12.55000000"], ["9280.13", "3.20310000"], ["9280.00", "1.36477909"], ["9276.01", "0.00707488"], ["9275.75", "0.56974291"], ["9275.00", "5.88000000"], ["9274.00", "0.00754205"], ["9271.68", "0.01400000"], ["9271.11", "15.37188500"], ["9270.00", "0.06674325"], ["9268.79", "24.54320000"], ["9257.18", "12.55000000"], ["9256.30", "0.17876365"], ["9255.71", "13.82642967"], ["9254.79", "0.96329407"], ["9250.00", "0.78214958"], ["9245.34", "4.90200000"], ["9245.13", "0.10000000"], ["9240.00", "0.44383459"], ["9238.84", "13.16615207"], ["9234.11", "0.43317656"], ["9234.10", "12.55000000"], ["9231.28", "11.79290000"], ["9230.09", "4.15059441"], ["9227.69", "0.00791097"], ["9225.00", "0.44768346"], ["9224.49", "0.85857203"], ["9223.50", "5.61001041"], ["9216.01", "0.03222653"], ["9216.00", "0.05000000"], ["9213.54", "0.71253866"], ["9212.50", "2.86768195"], ["9211.07", "12.55000000"], ["9210.00", "0.54288817"], ["9208.00", "1.00000000"], ["9206.06", "2.62587578"], ["9205.98", "15.40000000"], ["9205.52", "0.01710603"], ["9205.37", "0.03524953"], ["9205.11", "0.15000000"], ["9205.00", "0.01534763"], ["9204.76", "7.00600000"], ["9203.00", "0.01090000"]], "asks": [["9337.10", "0.03000000"], ["9340.85", "2.67820000"], ["9340.95", "0.02900000"], ["9341.17", "1.00000000"], ["9341.41", "2.13966390"], ["9341.61", "0.20000000"], ["9341.97", "0.11199911"], ["9341.98", "3.00000000"], ["9342.26", "0.32112762"], ["9343.87", "1.00000000"], ["9344.17", "3.57250000"], ["9345.04", "0.32103450"], ["9345.41", "4.90000000"], ["9345.69", "1.03000000"], ["9345.80", "0.03000000"], ["9346.00", "0.10200000"], ["9346.69", "0.02397394"], ["9347.41", "1.00000000"], ["9347.82", "0.32094177"], ["9348.23", "0.02880000"], ["9348.62", "11.96287551"], ["9349.31", "2.44270000"], ["9349.47", "0.96000000"], ["9349.86", "4.50000000"], ["9350.37", "0.03300000"], ["9350.57", "0.34682266"], ["9350.60", "0.32085527"], ["9351.45", "0.31147923"], ["9352.31", "0.28000000"], ["9352.86", "9.80000000"], ["9353.73", "0.02360739"], ["9354.00", "0.45000000"], ["9354.12", "0.03000000"], ["9354.29", "3.82446861"], ["9356.20", "0.64000000"], ["9356.90", "0.02316372"], ["9357.30", "2.50000000"], ["9357.70", "2.38240000"], ["9358.92", "6.00000000"], ["9359.97", "0.34898075"], ["9359.98", "2.30000000"], ["9362.56", "2.37600000"], ["9365.00", "0.64000000"], ["9365.16", "1.70030306"], ["9365.27", "3.03000000"], ["9369.99", "2.47102665"], ["9370.00", "3.15688574"], ["9370.21", "2.32720000"], ["9371.78", "13.20000000"], ["9371.89", "0.96293482"], ["9375.08", "4.74762500"], ["9384.34", "1.45200000"], ["9384.49", "16.42310000"], ["9385.66", "0.34382112"], ["9388.19", "0.00268265"], ["9392.20", "0.20980000"], ["9392.40", "0.10320000"], ["9393.00", "0.20980000"], ["9395.40", "0.40000000"], ["9398.86", "24.54310000"], ["9400.00", "0.05489988"], ["9400.33", "0.00495100"], ["9400.45", "0.00484700"], ["9402.92", "17.20000000"], ["9404.18", "10.00000000"], ["9418.89", "16.38000000"], ["9419.41", "3.06700000"], ["9420.40", "12.50000000"], ["9421.11", "0.10500000"], ["9434.47", "0.03215805"], ["9434.48", "0.28285714"], ["9434.49", "15.83000000"], ["9435.13", "0.15000000"], ["9438.93", "0.00368800"], ["9439.19", "0.69343985"], ["9442.86", "0.10000000"], ["9443.96", "12.50000000"], ["9444.00", "0.06004471"], ["9444.97", "0.01494896"], ["9447.00", "0.01234000"], ["9448.97", "0.14500000"], ["9449.00", "0.05000000"], ["9450.00", "11.13426018"], ["9451.87", "15.90000000"], ["9452.00", "0.20000000"], ["9454.25", "0.01100000"], ["9454.51", "0.02409062"], ["9455.05", "0.00600063"], ["9456.00", "0.27965118"], ["9456.10", "0.17000000"], ["9459.00", "0.00320000"], ["9459.98", "0.02460685"], ["9459.99", "8.11000000"], ["9460.00", "0.08500000"], ["9464.36", "0.56957951"], ["9464.54", "0.69158059"], ["9465.00", "21.00002015"], ["9467.57", "12.50000000"], ["9468.00", "0.08800000"], ["9469.09", "13.94000000"]]}, "event": "data", "channel": ""}`)
	if err = b.wsHandleData(pressXToJSON); !errors.Is(err, errWSPairParsingError) {
		t.Errorf("expected %s, got %s", errWSPairParsingError, err)
	}
}

func TestWsOrderbook2(t *testing.T) {
	pressXToJSON := []byte(`{"data":{"timestamp":"1606965727","microtimestamp":"1606965727403931","bids":[["19133.97","0.01000000"],["19131.58","0.39200000"],["19131.18","0.69581810"],["19131.17","0.48139054"],["19129.72","0.48164130"],["19129.71","0.65400000"],["19128.80","1.04500000"],["19128.59","0.65400000"],["19128.12","0.00259236"],["19127.81","0.19784245"],["19126.66","1.04500000"],["19125.74","0.26020000"],["19124.68","0.22000000"],["19122.01","0.39777840"],["19122.00","1.04600000"],["19121.27","0.16741000"],["19121.10","1.56390000"],["19119.90","1.60000000"],["19119.58","0.15593238"],["19117.70","1.14600000"],["19115.36","2.61300000"],["19114.60","1.19570000"],["19113.88","0.07500000"],["19113.86","0.15668522"],["19113.70","1.00000000"],["19113.69","1.60000000"],["19112.27","0.00166667"],["19111.00","0.15464628"],["19108.80","0.70000000"],["19108.77","0.16300000"],["19108.38","1.10000000"],["19107.53","0.10000000"],["19106.83","0.21377991"],["19106.78","3.45938881"],["19104.24","1.30000000"],["19100.81","0.00166667"],["19100.21","0.49770000"],["19099.54","2.40971961"],["19099.53","0.51223189"],["19097.40","1.55000000"],["19095.55","2.61300000"],["19092.94","0.27402906"],["19092.20","1.60000000"],["19089.36","0.00166667"],["19086.32","1.62000000"],["19085.23","1.65670000"],["19080.88","1.40000000"],["19075.45","1.16000000"],["19071.24","1.20000000"],["19065.09","1.51000000"],["19059.38","1.57000000"],["19058.11","0.37393556"],["19052.98","0.01000000"],["19052.90","0.33000000"],["19049.55","6.89000000"],["19047.61","6.03623432"],["19030.16","16.60260000"],["19026.76","23.90800000"],["19024.78","2.16656212"],["19022.11","0.02628500"],["19020.37","6.03000000"],["19000.00","0.00132020"],["18993.52","2.22000000"],["18979.21","6.03240000"],["18970.20","0.01500000"],["18969.14","7.42000000"],["18956.46","6.03240000"],["18950.22","42.37500000"],["18950.00","0.00132019"],["18949.94","0.52650000"],["18946.00","0.00791700"],["18933.74","6.03240000"],["18932.21","8.21000000"],["18926.99","0.00150000"],["18926.98","0.02641500"],["18925.00","0.02000000"],["18909.99","0.00133000"],["18908.47","7.15000000"],["18905.99","0.00133000"],["18905.20","0.00190000"],["18901.00","0.10000000"],["18900.67","0.24430000"],["18900.00","7.56529933"],["18895.99","0.00178450"],["18890.00","0.10000000"],["18889.90","0.10580000"],["18888.00","0.00362564"],["18887.00","4.00000000"],["18881.62","0.20583403"],["18880.08","5.72198740"],["18880.05","8.33480000"],["18879.09","7.33000000"],["18875.99","0.00132450"],["18875.00","0.02000000"],["18873.47","0.25934200"],["18871.99","0.00132600"],["18870.93","0.36463225"],["18864.10","43.56800000"],["18853.11","0.00540000"],["18850.01","0.38925549"]],"asks":[["19141.75","0.39300000"],["19141.78","0.10204700"],["19143.05","1.99685100"],["19143.08","0.05777900"],["19143.09","1.60700800"],["19143.10","0.48282909"],["19143.36","0.11250000"],["19144.06","0.26040000"],["19145.97","0.65400000"],["19146.02","0.22000000"],["19146.56","0.45061841"],["19147.45","0.15877831"],["19148.92","0.70431840"],["19148.93","0.78400000"],["19150.32","0.78400000"],["19151.55","0.07500000"],["19152.64","3.11400000"],["19153.32","1.04600000"],["19153.84","0.15626630"],["19155.57","3.10000000"],["19156.40","0.13438213"],["19156.92","0.16300000"],["19157.54","1.38970000"],["19158.18","0.00166667"],["19158.41","0.15317000"],["19158.78","0.15888798"],["19160.14","0.10000000"],["19160.34","1.60000000"],["19160.70","1.21590000"],["19162.17","0.00352761"],["19162.67","1.04500000"],["19163.61","0.15000000"],["19163.80","1.18050000"],["19164.62","0.86919692"],["19165.36","0.15674424"],["19166.75","1.40000000"],["19167.47","2.61300000"],["19169.68","0.00166667"],["19171.08","0.15452025"],["19171.69","0.54308236"],["19172.12","0.49000000"],["19173.47","1.34000000"],["19174.49","1.07436448"],["19175.37","0.01200000"],["19178.25","1.50000000"],["19178.80","0.49770000"],["19181.18","0.00166667"],["19182.75","1.77297176"],["19182.76","2.61099999"],["19183.03","1.20000000"],["19185.17","6.00352761"],["19189.56","0.05797137"],["19189.72","1.17000000"],["19193.94","1.60000000"],["19197.15","0.26961100"],["19200.00","0.03107838"],["19200.06","1.29000000"],["19202.73","1.65670000"],["19206.06","1.30000000"],["19208.19","6.00352761"],["19209.00","0.00132021"],["19210.70","1.20000000"],["19213.77","0.02615500"],["19217.40","8.50000000"],["19217.57","1.29000000"],["19222.61","1.19000000"],["19230.00","0.00193480"],["19231.24","6.00000000"],["19237.91","6.89152278"],["19240.13","6.90000000"],["19242.16","0.00336000"],["19243.38","0.00299103"],["19244.48","14.79300000"],["19248.25","0.01300000"],["19250.00","1.95802492"],["19251.00","0.45000000"],["19254.20","0.00366102"],["19254.32","6.00000000"],["19259.00","0.00131022"],["19266.43","0.00917191"],["19267.63","0.05000000"],["19267.79","7.10000000"],["19268.72","16.60260000"],["19277.42","6.00000000"],["19286.64","0.00916230"],["19295.49","7.77000000"],["19300.00","0.19668172"],["19306.00","0.06000000"],["19307.00","3.00000000"],["19307.40","0.19000000"],["19309.00","0.00262046"],["19310.33","0.02602500"],["19319.33","0.00213688"],["19320.00","0.00171242"],["19321.02","48.47300000"],["19322.74","0.00250000"],["19324.00","0.36983571"],["19325.54","0.02314521"],["19325.73","7.22000000"],["19326.50","0.00915272"]]},"channel":"order_book_btcusd","event":"data"}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestWsOrderUpdate(t *testing.T) {
	t.Parallel()
	n := new(Bitstamp)
	sharedtestvalues.TestFixtureToDataHandler(t, b, n, "testdata/wsMyOrders.json", n.wsHandleData)
	seen := 0
	for reading := true; reading; {
		select {
		default:
			reading = false
		case resp := <-n.GetBase().Websocket.DataHandler:
			seen++
			switch v := resp.(type) {
			case *order.Detail:
				switch seen {
				case 1:
					assert.Equal(t, "1658864794234880", v.OrderID, "OrderID")
					assert.Equal(t, time.UnixMicro(1693831262313000), v.Date, "Date")
					assert.Equal(t, "test_market_buy", v.ClientOrderID, "ClientOrderID")
					assert.Equal(t, order.New, v.Status, "Status")
					assert.Equal(t, order.Buy, v.Side, "Side")
					assert.Equal(t, asset.Spot, v.AssetType, "AssetType")
					assert.Equal(t, currency.NewPairWithDelimiter("BTC", "USD", "/"), v.Pair, "Pair")
					assert.Equal(t, 0.0, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 999999999.0, v.Price, "Price") // Market Buy Price
					// Note: Amount is 0 for market order create messages, oddly
				case 2:
					assert.Equal(t, "1658864794234880", v.OrderID, "OrderID")
					assert.Equal(t, order.PartiallyFilled, v.Status, "Status")
					assert.Equal(t, 0.00038667, v.Amount, "Amount")
					assert.Equal(t, 0.00000001, v.RemainingAmount, "RemainingAmount") // During live tests we consistently got back this Sat remaining
					assert.Equal(t, 0.00038666, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 25862.0, v.Price, "Price")
				case 3:
					assert.Equal(t, "1658864794234880", v.OrderID, "OrderID")
					assert.Equal(t, order.Cancelled, v.Status, "Status") // Even though they probably consider it filled, Deleted + PartialFill = Cancelled
					assert.Equal(t, 0.00038667, v.Amount, "Amount")
					assert.Equal(t, 0.00000001, v.RemainingAmount, "RemainingAmount")
					assert.Equal(t, 0.00038666, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 25862.0, v.Price, "Price")
				case 4:
					assert.Equal(t, "1658870500933632", v.OrderID, "OrderID")
					assert.Equal(t, order.New, v.Status, "Status")
					assert.Equal(t, order.Sell, v.Side, "Side")
					assert.Equal(t, 0.0, v.Price, "Price") // Market Sell Price
				case 5:
					assert.Equal(t, "1658870500933632", v.OrderID, "OrderID")
					assert.Equal(t, order.PartiallyFilled, v.Status, "Status")
					assert.Equal(t, 0.00038679, v.Amount, "Amount")
					assert.Equal(t, 0.00000001, v.RemainingAmount, "RemainingAmount")
					assert.Equal(t, 0.00038678, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 25854.0, v.Price, "Price")
				case 6:
					assert.Equal(t, "1658870500933632", v.OrderID, "OrderID")
					assert.Equal(t, order.Cancelled, v.Status, "Status")
					assert.Equal(t, 0.00038679, v.Amount, "Amount")
					assert.Equal(t, 0.00000001, v.RemainingAmount, "RemainingAmount")
					assert.Equal(t, 0.00038678, v.ExecutedAmount, "ExecutedAmount")
					assert.Equal(t, 25854.0, v.Price, "Price")
				case 7:
					assert.Equal(t, "1658869033291777", v.OrderID, "OrderID")
					assert.Equal(t, order.New, v.Status, "Status")
					assert.Equal(t, order.Sell, v.Side, "Side")
					assert.Equal(t, 25845.0, v.Price, "Price")
					assert.Equal(t, 0.00038692, v.Amount, "Amount")
				case 8:
					assert.Equal(t, "1658869033291777", v.OrderID, "OrderID")
					assert.Equal(t, order.Filled, v.Status, "Status")
					assert.Equal(t, 25845.0, v.Price, "Price")
					assert.Equal(t, 0.00038692, v.Amount, "Amount")
					assert.Equal(t, 0.0, v.RemainingAmount, "RemainingAmount")
					assert.Equal(t, 0.00038692, v.ExecutedAmount, "ExecutedAmount")
				}
			case error:
				t.Error(v)
			default:
				t.Errorf("Got unexpected data: %T %v", v, v)
			}
		}
	}
	assert.Equal(t, 8, seen, "Number of messages")
}

func TestWsRequestReconnect(t *testing.T) {
	pressXToJSON := []byte(`{
		"event": "bts:request_reconnect",
		"channel": "",
		"data": ""
	}`)
	err := b.wsHandleData(pressXToJSON)
	if err != nil {
		t.Error(err)
	}
}

func TestBitstamp_OHLC(t *testing.T) {
	start := time.Unix(1546300800, 0)
	end := time.Unix(1577836799, 0)
	_, err := b.OHLC(context.Background(), "btcusd", start, end, "60", "10")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBitstamp_GetHistoricCandles(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(1546300800, 0)
	end := time.Unix(1577836799, 0)
	_, err = b.GetHistoricCandles(context.Background(), pair, asset.Spot, kline.OneDay, start, end)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBitstamp_GetHistoricCandlesExtended(t *testing.T) {
	pair, err := currency.NewPairFromString("BTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(1546300800, 0)
	end := time.Unix(1577836799, 0)

	_, err = b.GetHistoricCandlesExtended(context.Background(), pair, asset.Spot, kline.OneDay, start, end)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("LTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetRecentTrades(context.Background(), currencyPair, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	currencyPair, err := currency.NewPairFromString("LTCUSD")
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.GetHistoricTrades(context.Background(),
		currencyPair, asset.Spot, time.Now().Add(-time.Minute*15), time.Now())
	if err != nil && err != common.ErrFunctionNotSupported {
		t.Error(err)
	}
}

func TestOrderbookZeroBidPrice(t *testing.T) {
	t.Parallel()
	ob := &orderbook.Base{
		Exchange: "Bitstamp",
		Pair:     currency.NewPair(currency.BTC, currency.USD),
		Asset:    asset.Spot,
	}

	filterOrderbookZeroBidPrice(ob)

	ob.Bids = orderbook.Items{
		{Price: 69, Amount: 1337},
		{Price: 0, Amount: 69},
	}

	filterOrderbookZeroBidPrice(ob)

	if ob.Bids[0].Price != 69 || ob.Bids[0].Amount != 1337 || len(ob.Bids) != 1 {
		t.Error("invalid orderbook bid values")
	}

	ob.Bids = orderbook.Items{
		{Price: 59, Amount: 1337},
		{Price: 42, Amount: 8595},
	}

	filterOrderbookZeroBidPrice(ob)

	if ob.Bids[0].Price != 59 || ob.Bids[0].Amount != 1337 ||
		ob.Bids[1].Price != 42 || ob.Bids[1].Amount != 8595 || len(ob.Bids) != 2 {
		t.Error("invalid orderbook bid values")
	}
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
	o, err := b.GetOrderInfo(context.Background(), "1458532827766784", currency.NewPair(currency.BTC, currency.USD), asset.Spot)
	if mockTests {
		assert.NoError(t, err, "GetOrderInfo should not error")
		assert.Equal(t, "2022-01-31 14:43:15", o.Date, "order date should match")
		assert.Equal(t, 1458532827766784, o.OrderID, "order ID should match")
		assert.Equal(t, "100.00", o.Amount, "amount should match")
		assert.Equal(t, "100.00", o.Price, "amount should match")
	}
	assert.ErrorContains(t, err, "authenticated request failed Order not found")
}

func TestFetchWSAuth(t *testing.T) {
	t.Parallel()
	resp, err := b.FetchWSAuth(context.TODO())
	assert.NoError(t, err, "FetchWSAuth should not error")
	assert.NotNil(t, resp, "resp should not be nil")
	assert.Positive(t, resp.UserID, "UserID should be positive")
	assert.Len(t, resp.Token, 32, "Token should be 32 chars")
	assert.Positive(t, resp.ValidSecs, "ValidSecs should be positive")
}
