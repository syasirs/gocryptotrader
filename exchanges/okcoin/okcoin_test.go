package okcoin

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
)

// Please supply you own test keys here for due diligence testing.
const (
	apiKey                  = ""
	apiSecret               = ""
	passphrase              = ""
	canManipulateRealOrders = false
)

var (
	o                    = &OKCoin{}
	spotCurrency         = currency.NewPairWithDelimiter(currency.BTC.String(), currency.USD.String(), "-")
	spotCurrencyLowerStr = spotCurrency.Lower().String()
	spotCurrencyUpperStr = spotCurrency.Upper().String()
)

func TestMain(m *testing.M) {
	o.SetDefaults()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal("Okcoin load config error", err)
	}
	okcoinConfig, err := cfg.GetExchangeConfig(o.Name)
	if err != nil {
		log.Fatalf("%v Setup() init error", o.Name)
	}

	okcoinConfig.API.AuthenticatedSupport = true
	okcoinConfig.API.AuthenticatedWebsocketSupport = true
	okcoinConfig.API.Credentials.Key = apiKey
	okcoinConfig.API.Credentials.Secret = apiSecret
	okcoinConfig.API.Credentials.ClientID = passphrase
	o.Websocket = sharedtestvalues.NewTestWebsocket()
	err = o.Setup(okcoinConfig)
	if err != nil {
		log.Fatal("OKCoin setup error", err)
	}
	os.Exit(m.Run())
}

func TestStart(t *testing.T) {
	t.Parallel()
	err := o.Start(context.Background(), nil)
	if !errors.Is(err, common.ErrNilPointer) {
		t.Fatalf("received: '%v' but expected: '%v'", err, common.ErrNilPointer)
	}
	var testWg sync.WaitGroup
	err = o.Start(context.Background(), &testWg)
	if err != nil {
		t.Fatal(err)
	}
	testWg.Wait()
}

func TestFetchTradablePair(t *testing.T) {
	t.Parallel()
	_, err := o.GetInstruments(context.Background(), "SPOT", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSystemStatus(t *testing.T) {
	t.Parallel()
	o.Verbose = true
	// allowed state value: ongoing, scheduled, processing, pre_open, completed, canceled
	_, err := o.GetSystemStatus(context.Background(), "scheduled")
	if err != nil {
		t.Fatal(err)
	}
	_, err = o.GetSystemStatus(context.Background(), "ongoing")
	if err != nil {
		t.Fatal(err)
	}
	_, err = o.GetSystemStatus(context.Background(), "processing")
	if err != nil {
		t.Fatal(err)
	}
	_, err = o.GetSystemStatus(context.Background(), "pre_open")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSystemTime(t *testing.T) {
	t.Parallel()
	_, err := o.GetSystemTime(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTickers(t *testing.T) {
	t.Parallel()
	_, err := o.GetTickers(context.Background(), "SPOT")
	if err != nil {
		t.Error(err)
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := o.GetTicker(context.Background(), "USDT-USD")
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderbooks(t *testing.T) {
	t.Parallel()
	_, err := o.GetOrderbook(context.Background(), "BTC-USD", 200)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLiteOrderbook(t *testing.T) {
	t.Parallel()
	_, err := o.GetOrderbookLitebook(context.Background(), "BTC-USD")
	if err != nil {
		t.Error(err)
	}
}

func TestGetCandlestick(t *testing.T) {
	t.Parallel()
	_, err := o.GetCandlesticks(context.Background(), "BTC-USD", kline.FiveMin, time.Now().Add(-time.Hour*3), time.Now(), 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCandlestickHistory(t *testing.T) {
	t.Parallel()
	_, err := o.GetCandlestickHistory(context.Background(), "BTC-USD", time.Now().Add(-time.Minute*30), time.Now(), kline.FiveMin, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()
	_, err := o.GetTrades(context.Background(), "BTC-USD", 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTradeHistory(t *testing.T) {
	t.Parallel()
	_, err := o.GetTradeHistory(context.Background(), "BTC-USD", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGet24HourTradingVolume(t *testing.T) {
	t.Parallel()
	_, err := o.Get24HourTradingVolume(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetOracle(t *testing.T) {
	t.Parallel()
	_, err := o.GetOracle(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetExchangeRate(t *testing.T) {
	t.Parallel()
	o.Verbose = true
	_, err := o.GetExchangeRate(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGenerateDefaultSubscriptions(t *testing.T) {
	t.Parallel()
	_, err := o.GenerateDefaultSubscriptions()
	if err != nil {
		t.Error(err)
	}
}

func TestWsConnect(t *testing.T) {
	err := o.WsConnect()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 25)
}

func TestGetCurrencies(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetCurrencies(context.Background(), currency.EMPTYCODE)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetBalance(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAccountAssetValuation(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetAccountAssetValuation(context.Background(), currency.EMPTYCODE)
	if err != nil {
		t.Error(err)
	}
}

func TestFundsTransfer(t *testing.T) {
	t.Parallel()
	_, err := o.FundsTransfer(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.FundsTransfer(context.Background(), &FundingTransferRequest{
		Currency: currency.EMPTYCODE,
	})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("found %v, but expected %v", err, currency.ErrCurrencyCodeEmpty)
	}
	_, err = o.FundsTransfer(context.Background(), &FundingTransferRequest{
		Currency: currency.BTC,
	})
	if !errors.Is(err, errInvalidAmount) { // "From" address
		t.Errorf("found %v, but expected %v", err, errInvalidAmount)
	}
	_, err = o.FundsTransfer(context.Background(), &FundingTransferRequest{
		Currency: currency.BTC,
		Amount:   1,
		From:     "abcde",
	})
	if !errors.Is(err, errAddressMustNotBeEmptyString) { // 'To' address
		t.Errorf("found %v, but expected %v", err, errAddressMustNotBeEmptyString)
	}
	_, err = o.FundsTransfer(context.Background(), &FundingTransferRequest{
		Currency:     currency.BTC,
		Amount:       1,
		From:         "abcdef",
		To:           "ghijklmnopqrstu",
		TransferType: 2,
	})
	if !errors.Is(err, errSubAccountNameRequired) {
		t.Errorf("found %v, but expected %v", err, errSubAccountNameRequired)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err = o.FundsTransfer(context.Background(), &FundingTransferRequest{
		Currency: currency.BTC,
		Amount:   1,
		From:     "1",
		To:       "6",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetFundsTransferState(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetFundsTransferState(context.Background(), "", "", "")
	if !errors.Is(err, errTransferIDOrClientIDRequred) {
		t.Error(err)
	}
	_, err = o.GetFundsTransferState(context.Background(), "1", "", "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetAssetBillType(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetAssetBilsDetail(context.Background(), currency.BTC, "2", "", time.Now().Add(-time.Minute), time.Now().Add(-time.Hour), 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLightningDeposits(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetLightningDeposits(context.Background(), currency.BTC, 0.001, "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetCurrencyDepositAddresses(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetCurrencyDepositAddresses(context.Background(), currency.EMPTYCODE)
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Errorf("found %v, expected %v", err, currency.ErrCurrencyCodeEmpty)
	}
	_, err = o.GetCurrencyDepositAddresses(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetDepositHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetDepositHistory(context.Background(), currency.BTC, "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawal(t *testing.T) {
	t.Parallel()
	_, err := o.Withdrawal(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Fatalf("found %v, expected %v", err, errNilArgument)
	}
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{})
	if !errors.Is(err, currency.ErrCurrencyCodeEmpty) {
		t.Fatalf("found %v, expected %v", err, currency.ErrCurrencyCodeEmpty)
	}
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{
		Ccy: currency.BTC,
	})
	if !errors.Is(err, errInvalidAmount) {
		t.Fatalf("found %v, expected %v", err, errInvalidAmount)
	}
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{
		Ccy:    currency.BTC,
		Amount: 1,
	})
	if !errors.Is(err, errInvalidWithdrawalMethod) {
		t.Fatalf("found %v, expected %v", err, errInvalidWithdrawalMethod)
	}
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{Amount: 1, Ccy: currency.BTC, WithdrawalMethod: "1"})
	if !errors.Is(err, errAddressMustNotBeEmptyString) {
		t.Fatalf("found %v, expected %v", err, errInvalidWithdrawalMethod)
	}
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{Amount: 1, Ccy: currency.BTC, WithdrawalMethod: "1", ToAddress: "abcdefg"})
	if !errors.Is(err, errInvalidTrasactionFeeValue) {
		t.Fatalf("found %v, expected %v", err, errAddressMustNotBeEmptyString)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err = o.Withdrawal(context.Background(), &WithdrawalRequest{Amount: 1, Ccy: currency.BTC, WithdrawalMethod: "1", ToAddress: "abcdefg", TransactionFee: 0.004})
	if err != nil {
		t.Error(err)
	}
}

func TestLightningWithdrawals(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.SubmitLightningWithdrawals(context.Background(), &LightningWithdrawalsRequest{
		Ccy:     currency.BTC,
		Invoice: "something",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelWithdrawal(t *testing.T) {
	t.Parallel()
	// sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.CancelWithdrawal(context.Background(), &WithdrawalCancelation{
		WithdrawalID: "1123456",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetWithdrawalHistory(t *testing.T) {
	t.Parallel()
	_, err := o.GetWithdrawalHistory(context.Background(), currency.BTC, "", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAccountBalance(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetAccountBalance(context.Background(), currency.BTC, currency.USDT)
	if err != nil {
		t.Error(err)
	}
}

func TestGetBillsDetails(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetBillsDetails(context.Background(), currency.BTC, "", "", "", "", "", time.Now().Add(-time.Hour*30), time.Now(), 0)
	if err != nil {
		t.Error(err)
	}
}
func TestGetBillsDetailsFor3Months(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetBillsDetailsFor3Months(context.Background(), currency.BTC, "", "", "", "", "", time.Now().Add(-time.Hour*30), time.Now(), 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAccountConfigurations(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetAccountConfigurations(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestGetMaximumBuySellOrOpenAmount(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetMaximumBuySellOrOpenAmount(context.Background(), "BTC-USD", "cash", 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMaximumAvailableTradableAmount(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetMaximumAvailableTradableAmount(context.Background(), "cash", "BTC-USD")
	if err != nil {
		t.Error(err)
	}
}

func TestGetFeeRates(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetFeeRates(context.Background(), "SPOT", "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetMaximumWithdrawals(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetMaximumWithdrawals(context.Background(), currency.BTC)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAvailableRFQPairs(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetAvailableRFQPairs(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestRequestQuote(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.RequestQuote(context.Background(), &QuoteRequestArg{
		BaseCurrency:  currency.BTC,
		QuoteCurrency: currency.USD,
		Side:          "sell",
		RfqSize:       1000,
		RfqSzCurrency: currency.USD,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestPlaceRFQOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.PlaceRFQOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("expected %v, but found %v", errNilArgument, err)
	}
}

func TestGetRFQOrderDetails(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetRFQOrderDetails(context.Background(), "", "1234")
	if err != nil {
		t.Error(err)
	}
}

func TestGetRFQOrderHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetRFQOrderHistory(context.Background(), time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestDeposit(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.Deposit(context.Background(), &FiatDepositRequestArg{
		ChannelID:         "28",
		BankAccountNumber: "1000221891299",
		Amount:            100,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelFiatDeposit(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.CancelFiatDeposit(context.Background(), "1234")
	if err != nil {
		t.Error(err)
	}
}

func TestGetFiatDepositHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetFiatDepositHistory(context.Background(), currency.BTC, "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestFiatWithdrawal(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.FiatWithdrawal(context.Background(), &FiatWithdrawalParam{
		ChannelID:      "3",
		BankAcctNumber: "100221891299",
		Amount:         12,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestFiatCancelWithdrawal(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.FiatCancelWithdrawal(context.Background(), "1234")
	if err != nil {
		t.Error(err)
	}
}

func TestGetFiatWithdrawalHistory(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetFiatWithdrawalHistory(context.Background(), currency.BTC, "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestGetChannelInfo(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetChannelInfo(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
}

func TestPlaceOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.PlaceOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Errorf("found %v, but expected %v", err, currency.ErrCurrencyPairEmpty)
	}
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{InstrumentID: currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD}})
	if !errors.Is(err, errTradeModeIsRequired) {
		t.Errorf("found %v, but expected %v", err, errTradeModeIsRequired)
	}
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{InstrumentID: currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
		TradeMode: "cash",
	})
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{InstrumentID: currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
		TradeMode: "cash",
		Side:      "buy",
	})
	if !errors.Is(err, order.ErrTypeIsInvalid) {
		t.Errorf("found %v, but expected %v", err, order.ErrTypeIsInvalid)
	}
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{InstrumentID: currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
		TradeMode: "cash",
		Side:      "buy",
		OrderType: "limit",
	})
	if !errors.Is(err, errInvalidAmount) {
		t.Errorf("found %v, but expected %v", err, errInvalidAmount)
	}
	_, err = o.PlaceOrder(context.Background(), &PlaceTradeOrderParam{
		InstrumentID:  currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
		TradeMode:     "cash",
		ClientOrderID: "12345",
		Side:          "buy",
		OrderType:     "limit",
		Price:         2.15,
		Size:          2,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestPlaceMultipleOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	o.Verbose = true
	_, err := o.PlaceMultipleOrder(context.Background(), []PlaceTradeOrderParam{
		{
			InstrumentID:  currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
			TradeMode:     "cash",
			ClientOrderID: "1",
			Side:          "buy",
			OrderType:     "limit",
			Price:         2.15,
			Size:          2,
		},
		{
			InstrumentID:  currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
			TradeMode:     "cash",
			ClientOrderID: "12",
			Side:          "buy",
			OrderType:     "limit",
			Price:         2.15,
			Size:          1.5,
		},
		{
			InstrumentID:  currency.Pair{Base: currency.BTC, Delimiter: "-", Quote: currency.USD},
			TradeMode:     "cash",
			ClientOrderID: "123",
			Side:          "buy",
			OrderType:     "limit",
			Price:         2.15,
			Size:          1,
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelTradeOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	o.Verbose = true
	_, err := o.CancelTradeOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.CancelTradeOrder(context.Background(), &CancelTradeOrderRequest{})
	if !errors.Is(err, errMissingInstrumentID) {
		t.Errorf("found %v, but expected %v", err, errMissingInstrumentID)
	}
	_, err = o.CancelTradeOrder(context.Background(), &CancelTradeOrderRequest{
		InstrumentID: "BTC-USD",
	})
	if !errors.Is(err, errOrderIDOrClientOrderIDRequired) {
		t.Errorf("found %v, but expected %v", err, errOrderIDOrClientOrderIDRequired)
	}
	_, err = o.CancelTradeOrder(context.Background(), &CancelTradeOrderRequest{
		InstrumentID:  "BTC-USD",
		ClientOrderID: "123",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelMultipleOrders(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.CancelMultipleOrders(context.Background(), []CancelTradeOrderRequest{
		{
			InstrumentID:  "BTC-USD",
			ClientOrderID: "123",
		},
		{
			InstrumentID:  "BTC-USD",
			ClientOrderID: "abcdefg",
		},
		{
			InstrumentID:  "ETH-USD",
			ClientOrderID: "1234",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestAmendOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.AmendOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.AmendOrder(context.Background(), &AmendTradeOrderRequestParam{})
	if !errors.Is(err, errMissingInstrumentID) {
		t.Errorf("found %v, expected %v", err, errMissingInstrumentID)
	}
	_, err = o.AmendOrder(context.Background(), &AmendTradeOrderRequestParam{
		InstrumentID: "BTC-USD"})
	if !errors.Is(err, errOrderIDOrClientOrderIDRequired) {
		t.Errorf("found %v, expected %v", err, errOrderIDOrClientOrderIDRequired)
	}
	_, err = o.AmendOrder(context.Background(), &AmendTradeOrderRequestParam{
		InstrumentID: "BTC-USD",
		OrderID:      "1234"})
	if !errors.Is(err, errSizeOrPriceRequired) {
		t.Errorf("found %v, expected %v", err, errSizeOrPriceRequired)
	}
	_, err = o.AmendOrder(context.Background(), &AmendTradeOrderRequestParam{
		InstrumentID: "BTC-USD",
		OrderID:      "1234",
		NewSize:      5,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestAmendMultipleOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.AmendMultipleOrder(context.Background(), []AmendTradeOrderRequestParam{
		{
			InstrumentID: "BTC-USD",
			OrderID:      "1234",
			NewSize:      5,
		},
		{
			InstrumentID:  "BTC-USD",
			ClientOrderID: "abe",
			NewPrice:      100,
		},
		{
			InstrumentID:    "BTC-USD",
			OrderID:         "3452",
			ClientRequestID: "9879",
			NewSize:         2,
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderDetail(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetPersonalOrderDetail(context.Background(), "BTC-USD", "1243", "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetPersonalOrderList(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetPersonalOrderList(context.Background(), "SPOT", "BTC-USD", "", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderHistory7Days(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetOrderHistory7Days(context.Background(), "SPOT", "BTC-USD", "", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetOrderHistory3MonthsDays(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetOrderHistory3MonthsDays(context.Background(), "SPOT", "BTC-USD", "", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Error(err)
	}
}

func TestGetRecentTransactionDetail(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetRecentTransactionDetail(context.Background(), "SPOT", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}
func TestGetTransactionDetails3Months(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o)
	_, err := o.GetTransactionDetails3Months(context.Background(), "SPOT", "", "", "", "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Error(err)
	}
}

func TestPlaceAlgoOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.PlaceAlgoOrder(context.Background(), nil)
	if !errors.Is(err, errNilArgument) {
		t.Error(err)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{})
	if !errors.Is(err, errMissingInstrumentID) {
		t.Errorf("found %v, but expected %v", err, errMissingInstrumentID)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
	})
	if !errors.Is(err, errTradeModeIsRequired) {
		t.Errorf("found %v, but expected %v", err, errTradeModeIsRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
	})
	if !errors.Is(err, order.ErrSideIsInvalid) {
		t.Errorf("found %v, but expected %v", err, order.ErrSideIsInvalid)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
	})
	if !errors.Is(err, order.ErrTypeIsInvalid) {
		t.Errorf("found %v, but expected %v", err, order.ErrTypeIsInvalid)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		OrderType:    "oco",
	})
	if !errors.Is(err, errInvalidAmount) {
		t.Errorf("found %v, but expected %v", err, errInvalidAmount)
	}

	// Stop loss
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "conditional",
	})
	if !errors.Is(err, errStopLossOrTakeProfitOrderPriceRequired) {
		t.Errorf("found %v, but expected %v", err, errStopLossOrTakeProfitOrderPriceRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "conditional",
	})
	if !errors.Is(err, errStopLossOrTakeProfitOrderPriceRequired) {
		t.Errorf("found %v, but expected %v", err, errStopLossOrTakeProfitOrderPriceRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "conditional",
		TpOrderPrice: 1,
	})
	if !errors.Is(err, errTakeProfitOrderPriceRequired) {
		t.Errorf("found %v, but expected %v", err, errTakeProfitOrderPriceRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:   "BTC-USD",
		TradeMode:      "cash",
		Side:           "buy",
		Size:           2,
		OrderType:      "conditional",
		TpOrderPrice:   1,
		TpTriggerPrice: 1,
	})
	if !errors.Is(err, errTpTriggerOrderPriceTypeRequired) {
		t.Errorf("found %v, but expected %v", err, errTpTriggerOrderPriceTypeRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:            "BTC-USD",
		TradeMode:               "cash",
		Side:                    "buy",
		Size:                    2,
		OrderType:               "conditional",
		TpOrderPrice:            1,
		TpTriggerOrderPriceType: "last",
		TpTriggerPrice:          1,
	})
	if err != nil {
		t.Error(err)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:       "BTC-USD",
		TradeMode:          "cash",
		Side:               "buy",
		Size:               2,
		OrderType:          "conditional",
		StopLossOrderPrice: 1,
	})
	if !errors.Is(err, errStopLossTriggerPriceRequired) {
		t.Errorf("found %v, but expected %v", err, errStopLossTriggerPriceRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:         "BTC-USD",
		TradeMode:            "cash",
		Side:                 "buy",
		Size:                 2,
		OrderType:            "conditional",
		StopLossOrderPrice:   1,
		StopLossTriggerPrice: 2,
	})
	if !errors.Is(err, errStopLossTriggerPriceTypeRequired) {
		t.Errorf("found %v, but expected %v", err, errStopLossTriggerPriceTypeRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:             "BTC-USD",
		TradeMode:                "cash",
		Side:                     "buy",
		Size:                     2,
		OrderType:                "conditional",
		StopLossOrderPrice:       5000,
		StopLossTriggerPrice:     2,
		StopLossTriggerPriceType: "last",
	})
	if err != nil {
		t.Error(err)
	}
	//  Trigger order
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "trigger",
		TriggerPrice: 123,
	})
	if !errors.Is(err, errInvalidPrice) {
		t.Errorf("found %v, but expected %v", err, errInvalidPrice)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "trigger",
		TriggerPrice: 123,
	})
	if !errors.Is(err, errInvalidPrice) {
		t.Errorf("found %v, but expected %v", err, errInvalidPrice)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "trigger",
		TriggerPrice: 123,
		OrderPrice:   234,
	})
	if err != nil {
		t.Error(err)
	}

	// OCO Orders
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "oco",
	})
	if !errors.Is(err, errTakeProfitOrderPriceRequired) {
		t.Errorf("found %v, but expected %v", err, errTakeProfitOrderPriceRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "oco",
		TpOrderPrice: 123,
	})
	if !errors.Is(err, errTpTriggerOrderPriceTypeRequired) {
		t.Errorf("found %v, but expected %v", err, errTpTriggerOrderPriceTypeRequired)
	}

	// Iceberg order
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "move_order_stop",
	})
	if !errors.Is(err, errCallbackRatioOrCallbackSpeedRequired) {
		t.Errorf("found %v, but expected %v", err, errCallbackRatioOrCallbackSpeedRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID:  "BTC-USD",
		TradeMode:     "cash",
		Side:          "buy",
		Size:          2,
		OrderType:     "move_order_stop",
		CallbackRatio: 0.002,
	})
	if err != nil {
		t.Error(err)
	}
	// Twap Order
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
	})
	if !errors.Is(err, errPriceRatioOrPriveSpreadRequired) {
		t.Errorf("found %v, but expected %v", err, errPriceRatioOrPriveSpreadRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
		PriceRatio:   0.02,
	})
	if !errors.Is(err, errSizeLimitRequired) {
		t.Errorf("found %v, but expected %v", err, errSizeLimitRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
		PriceRatio:   0.02,
		PriceLimit:   1234,
	})
	if !errors.Is(err, errSizeLimitRequired) {
		t.Errorf("found %v, but expected %v", err, errSizeLimitRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
		PriceRatio:   0.02,
		SizeLimit:    1234,
		TimeInterval: "1m",
	})
	if !errors.Is(err, errPriceLimitRequired) {
		t.Errorf("found %v, but expected %v", err, errPriceLimitRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
		PriceRatio:   0.02,
		PriceLimit:   1234,
		SizeLimit:    1234,
	})
	if !errors.Is(err, errTimeIntervlaInformationRequired) {
		t.Errorf("found %v, but expected %v", err, errTimeIntervlaInformationRequired)
	}
	_, err = o.PlaceAlgoOrder(context.Background(), &AlgoOrderRequestParam{
		InstrumentID: "BTC-USD",
		TradeMode:    "cash",
		Side:         "buy",
		Size:         2,
		OrderType:    "twap",
		PriceRatio:   0.02,
		PriceLimit:   1234,
		SizeLimit:    1234,
		TimeInterval: "1m",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCancelAlgoOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.CancelAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{})
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.CancelAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			InstrumentID: "BTC-USD",
		},
	})
	if !errors.Is(err, errAlgoIDRequired) {
		t.Errorf("found %v, but expected %v", err, errAlgoIDRequired)
	}
	_, err = o.CancelAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			AlgoOrderID: "1234",
		},
	})
	if !errors.Is(err, errMissingInstrumentID) {
		t.Errorf("found %v, but expected %v", err, errMissingInstrumentID)
	}
	_, err = o.CancelAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			InstrumentID: "BTC-USD",
			AlgoOrderID:  "2234",
		},
	})
	if err != nil {
		t.Error(err)
	}
}
func TestCancelAdvancedAlgoOrder(t *testing.T) {
	t.Parallel()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, o, canManipulateRealOrders)
	_, err := o.CancelAdvancedAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{})
	if !errors.Is(err, errNilArgument) {
		t.Errorf("found %v, but expected %v", err, errNilArgument)
	}
	_, err = o.CancelAdvancedAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			InstrumentID: "BTC-USD",
		},
	})
	if !errors.Is(err, errAlgoIDRequired) {
		t.Errorf("found %v, but expected %v", err, errAlgoIDRequired)
	}
	_, err = o.CancelAdvancedAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			AlgoOrderID: "1234",
		},
	})
	if !errors.Is(err, errMissingInstrumentID) {
		t.Errorf("found %v, but expected %v", err, errMissingInstrumentID)
	}
	_, err = o.CancelAdvancedAlgoOrder(context.Background(), []CancelAlgoOrderRequestParam{
		{
			InstrumentID: "BTC-USD",
			AlgoOrderID:  "2234",
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestFetchTradablePairs(t *testing.T) {
	t.Parallel()
	_, err := o.FetchTradablePairs(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTradablePairs(t *testing.T) {
	t.Parallel()
	err := o.UpdateTradablePairs(context.Background(), true)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTickers(t *testing.T) {
	t.Parallel()
	err := o.UpdateTickers(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateTicker(t *testing.T) {
	t.Parallel()
	_, err := o.UpdateTicker(context.Background(), currency.Pair{Base: currency.BTC, Delimiter: currency.DashDelimiter, Quote: currency.USD}, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestFetchTicker(t *testing.T) {
	t.Parallel()
	_, err := o.FetchTicker(context.Background(), currency.Pair{Base: currency.BTC, Delimiter: currency.DashDelimiter, Quote: currency.USD}, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetRecentTrades(t *testing.T) {
	t.Parallel()
	_, err := o.GetRecentTrades(context.Background(), currency.Pair{Base: currency.BTC, Delimiter: currency.DashDelimiter, Quote: currency.USD}, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}
func TestFetchOrderbook(t *testing.T) {
	t.Parallel()
	_, err := o.FetchOrderbook(context.Background(), spotCurrency, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateOrderbook(t *testing.T) {
	t.Parallel()
	_, err := o.UpdateOrderbook(context.Background(), spotCurrency, asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateAccountInfo(t *testing.T) {
	t.Parallel()
	_, err := o.UpdateAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestFetchAccountInfo(t *testing.T) {
	t.Parallel()
	_, err := o.FetchAccountInfo(context.Background(), asset.Spot)
	if err != nil {
		t.Error(err)
	}
}

func TestGetHistoricCandles(t *testing.T) {
	t.Parallel()
	_, err := o.GetHistoricCandles(context.Background(), spotCurrency, asset.Spot, kline.OneDay, time.Now().Add(-5*time.Hour), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	_, err = o.GetHistoricCandles(context.Background(), spotCurrency, asset.Spot, kline.Interval(time.Hour*4), time.Now().Add(-5*time.Hour), time.Now())
	if !errors.Is(err, kline.ErrRequestExceedsExchangeLimits) {
		t.Errorf("received: '%v' but expected: '%v'", err, kline.ErrRequestExceedsExchangeLimits)
	}
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	t.Parallel()
	_, err := o.GetHistoricCandlesExtended(context.Background(), spotCurrency, asset.Spot, kline.OneMin, time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Errorf("%s GetHistoricCandlesExtended() error: %v", o.Name, err)
	}
}

func TestGetHistoricTrades(t *testing.T) {
	t.Parallel()
	if _, err := o.GetHistoricTrades(context.Background(), currency.NewPair(currency.BTC, currency.USDT), asset.Spot, time.Now().Add(-time.Minute*4), time.Now().Add(-time.Minute*2)); err != nil {
		t.Errorf("%s GetHistoricTrades() error %v", o.Name, err)
	}
}
