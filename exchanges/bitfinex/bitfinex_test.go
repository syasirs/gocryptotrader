package bitfinex

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency"
)

var ACCOUNT_LIVE_TEST bool = false //Supply correct API keys in testdata/configtest.dat before changing this.

func TestSetDefaults(t *testing.T) {
	t.Parallel()

	setDefaults := Bitfinex{}
	setDefaults.SetDefaults()

	if setDefaults.Name != "Bitfinex" || setDefaults.Enabled != false || setDefaults.Verbose != false ||
		setDefaults.Websocket != false || setDefaults.RESTPollingDelay != 10 {
		t.Error("Test Failed - Bitfinex SetDefaults values not set correctly")
	}
	if reflect.TypeOf(setDefaults.WebsocketSubdChannels).String() != "map[int]bitfinex.BitfinexWebsocketChanInfo" {
		t.Error("Test Failed - Bitfinex SetDefaults value: MAP not set correctly")
	}
}

func TestSetup(t *testing.T) {
	t.Parallel()

	testConfig := config.ExchangeConfig{
		Enabled:                 true,
		AuthenticatedAPISupport: true,
		APIKey:                  "lamb",
		APISecret:               "cutlets",
		RESTPollingDelay:        time.Duration(10),
		Verbose:                 true,
		Websocket:               true,
		BaseCurrencies:          currency.DefaultCurrencies,
		AvailablePairs:          currency.MakecurrencyPairs(currency.DefaultCurrencies),
		EnabledPairs:            currency.MakecurrencyPairs(currency.DefaultCurrencies),
	}
	setup := Bitfinex{}
	setup.Setup(testConfig)

	if !setup.Enabled || !setup.AuthenticatedAPISupport || setup.APIKey != "lamb" ||
		setup.APISecret != "cutlets" || setup.RESTPollingDelay != time.Duration(10) ||
		!setup.Verbose || !setup.Websocket || len(setup.BaseCurrencies) < 1 ||
		len(setup.AvailablePairs) < 1 || len(setup.EnabledPairs) < 1 {

		t.Error("Test Failed - Bitfinex Setup values not set correctly")
	}
}

//Live Testing
func TestGetTicker(t *testing.T) {
	t.Parallel()
	bitfinex := Bitfinex{}

	response, err := bitfinex.GetTicker("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetTicker init error: ", err)
	}
	if reflect.ValueOf(response).NumField() != 8 {
		t.Error("BitfinexGetTicker struct change/or updated")
	}
	if reflect.TypeOf(response.Timestamp).String() != "string" {
		t.Error("Bitfinex ticker.Timestamp value is not a string variable")
	}
	if reflect.TypeOf(response.Ask).String() != "float64" {
		t.Error("Bitfinex ticker.Ask value is not a float64 variable")
	}
	if reflect.TypeOf(response.Bid).String() != "float64" {
		t.Error("Bitfinex ticker.Bid value is not a float64 variable")
	}
	if reflect.TypeOf(response.High).String() != "float64" {
		t.Error("Bitfinex ticker.High value is not a float64 variable")
	}
	if reflect.TypeOf(response.Last).String() != "float64" {
		t.Error("Bitfinex ticker.Last value is not a float64 variable")
	}
	if reflect.TypeOf(response.Low).String() != "float64" {
		t.Error("Bitfinex ticker.Low value is not a float64 variable")
	}
	if reflect.TypeOf(response.Mid).String() != "float64" {
		t.Error("Bitfinex ticker.Mid value is not a float64 variable")
	}
	if reflect.TypeOf(response.Volume).String() != "float64" {
		t.Error("Bitfinex ticker.Volume value is not a float64 variable")
	}

	responseTimestamp, err := strconv.ParseFloat(response.Timestamp, 64)
	if err != nil {
		t.Error("ticker.Timestamp value cannot be converted to a float64")
	}
	if responseTimestamp <= 0 {
		t.Error("ticker.Timestamp value is negative or 0")
	}
	if response.Ask < 0 {
		t.Error("ticker.Ask value is negative")
	}
	if response.Bid < 0 {
		t.Error("ticker.Bid value is negative")
	}
	if response.High < 0 {
		t.Error("ticker.High value is negative")
	}
	if response.Last < 0 {
		t.Error("ticker.Last value is negative")
	}
	if response.Low < 0 {
		t.Error("ticker.Low value is negative")
	}
	if response.Mid < 0 {
		t.Error("ticker.Mid value is negative")
	}
	if response.Volume < 0 {
		t.Error("ticker.ask value is negative")
	}
}

//Live Testing
func TestGetStats(t *testing.T) {
	t.Parallel()
	BitfinexGetStatsTest := Bitfinex{}

	response, err := BitfinexGetStatsTest.GetStats("BTCUSD")
	if err != nil {
		t.Error("BitfinexGetStatsTest init error: ", err)
	}
	if reflect.ValueOf(response[0]).NumField() != 2 {
		t.Error("BitfinexGetTicker []struct change/or updated")
	}
	if reflect.TypeOf(response[0].Period).String() != "int64" {
		t.Error("Bitfinex Getstats.Period is not an int64")
	}
	if reflect.TypeOf(response[0].Volume).String() != "float64" {
		t.Error("Bitfiniex Getstats.Volume is not a float64")
	}

	for _, explicitResponse := range response {
		if explicitResponse.Period <= 0 {
			t.Error("response.Period value is negative or zero")
		}
		if explicitResponse.Volume < 0 {
			t.Error("response.Volume value is negative")
		}
	}
}

//Live Testing
func TestGetLendbook(t *testing.T) {
	t.Parallel()
	BitfinexGetLendbook := Bitfinex{}

	response, err := BitfinexGetLendbook.GetLendbook("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetLendbook init error: ", err)
	}
	if reflect.ValueOf(response).NumField() != 2 {
		t.Error("BitfinexGetLendbook struct change/or updated")
	}
	if reflect.ValueOf(response.Asks[0]).NumField() != 5 {
		t.Error("BitfinexGetLendbook GetLendbook.Asks []struct change/or updated")
	}
	if reflect.TypeOf(response.Asks[0].Amount).String() != "float64" {
		t.Error("Bitfinex GetLendbook.Asks.Amount is not a float64")
	}
	if reflect.TypeOf(response.Asks[0].FlashReturnRate).String() != "string" {
		t.Error("Bitfinex GetLendbook.Asks.FlashReturnRate is not a string")
	}
	if reflect.TypeOf(response.Asks[0].Period).String() != "int" {
		t.Error("Bitfinex GetLendbook.Asks.Period is not an int")
	}
	if reflect.TypeOf(response.Asks[0].Rate).String() != "float64" {
		t.Error("Bitfinex GetLendbook.Asks.Rate is not a float64")
	}
	if reflect.ValueOf(response.Bids[0]).NumField() != 5 {
		t.Error("BitfinexGetLendbook GetLendbook.Bids []struct change/or updated")
	}
	if reflect.TypeOf(response.Bids[0].Amount).String() != "float64" {
		t.Error("Bitfinex GetLendbook.Bids.Amount is not a float64")
	}
	if reflect.TypeOf(response.Bids[0].FlashReturnRate).String() != "string" {
		t.Error("Bitfinex GetLendbook.Bids.FlashReturnRate is not a string")
	}
	if reflect.TypeOf(response.Bids[0].Period).String() != "int" {
		t.Error("Bitfinex GetLendbook.Bids.Period is not an int")
	}
	if reflect.TypeOf(response.Bids[0].Rate).String() != "float64" {
		t.Error("Bitfinex GetLendbook.Bids.Rate is not a float64")
	}

	for _, asks := range response.Asks {
		responseTimestamp, err := strconv.ParseFloat(asks.Timestamp, 64)
		if err != nil {
			t.Error("Could not convert Bitfinex GetLendbook.Asks.Timestamp into float64")
		}
		if asks.Amount <= 0 {
			t.Error("Bitfinex GetLendbook.Asks.Amount is negative or 0")
		}
		if asks.FlashReturnRate != "No" && asks.FlashReturnRate != "Yes" {
			t.Error("Bitfinex GetLendbook.Bids.FlashReturnRate incorrect string")
		}
		if asks.Period <= 0 {
			t.Error("Bitfinex GetLendbook.Asks.Period is negative or 0")
		}
		if asks.Rate <= 0 {
			t.Error("Bitfinex GetLendbook.Asks.Rate is negative or 0")
		}
		if responseTimestamp <= 0 {
			t.Error("Bitfinex GetLendbook.Asks.Timestamp is negative or 0")
		}
	}

	for _, bids := range response.Bids {
		responseTimetamp, err := strconv.ParseFloat(bids.Timestamp, 64)
		if err != nil {
			t.Error("Could not convert Bitfinex GetLendbook.Bids.Timestamp into float64")
		}
		if bids.Amount <= 0 {
			t.Error("Bitfinex GetLendbook.Bids.Amount is negative or 0")
		}
		if bids.FlashReturnRate == "no" || bids.FlashReturnRate == "yes" {
			t.Error("Bitfinex GetLendbook.Bids.FlashReturnRate incorrect string")
		}
		if bids.Period <= 0 {
			t.Error("Bitfinex GetLendbook.Bids.Period is negative or 0")
		}
		if bids.Rate <= 0 {
			t.Error("Bitfinex GetLendbook.Bids.Rate is negative or 0")
		}
		if responseTimetamp <= 0 {
			t.Error("Bitfinex GetLendbook.Bids.Timestamp is negative or 0")
		}
	}
}

//Live Testing
func TestGetOrderbook(t *testing.T) {
	t.Parallel()
	BitfinexGetOrderbook := Bitfinex{}

	orderBook, err := BitfinexGetOrderbook.GetOrderbook("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetOrderbook init error: ", err)
	}
	if reflect.ValueOf(orderBook).NumField() != 2 {
		t.Error("BitfinexGetOrderbook struct change/or updated")
	}
	if reflect.ValueOf(orderBook.Asks[0]).NumField() != 3 {
		t.Error("BitfinexGetOrderbook []struct change/or updated")
	}
	if reflect.ValueOf(orderBook.Bids[0]).NumField() != 3 {
		t.Error("BitfinexGetOrderbook []struct change/or updated")
	}
	if reflect.TypeOf(orderBook.Asks[0].Amount).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}
	if reflect.TypeOf(orderBook.Asks[0].Price).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}
	if reflect.TypeOf(orderBook.Asks[0].Timestamp).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}
	if reflect.TypeOf(orderBook.Bids[0].Amount).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}
	if reflect.TypeOf(orderBook.Bids[0].Price).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}
	if reflect.TypeOf(orderBook.Bids[0].Timestamp).String() != "string" {
		t.Error("Bitfinex GetOrderbook.Bids.Amount is not a string")
	}

	for _, asks := range orderBook.Asks {
		amount, err := strconv.ParseFloat(asks.Amount, 64)
		if err != nil {
			t.Error("Cannot convert Bitfinex Orderbook.Asks.Amount into a float64")
		}
		if amount < 0 {
			t.Error("Bitfinex Orderbook.Asks.Amount is negative")
		}
		price, err2 := strconv.ParseFloat(asks.Price, 64)
		if err2 != nil {
			t.Error("Cannot convert Bitfinex Orderbook.Asks.Price into a float64")
		}
		if price < 0 {
			t.Error("Bitfinex Orderbook.Asks.Price is negative")
		}
		timestamp, err3 := strconv.ParseFloat(asks.Timestamp, 64)
		if err3 != nil {
			t.Error("Cannot convert Bitfinex Orderbook.Asks.timestamp into a float64")
		}
		if timestamp <= 0 {
			t.Error("Bitfinex Orderbook.Asks.Amount is negative or 0")
		}
	}

	for _, bids := range orderBook.Bids {
		amount, err := strconv.ParseFloat(bids.Amount, 64)
		if err != nil {
			t.Error("Cannot convert Bitfinex Orderbook.bids.Amount into a float64")
		}
		if amount < 0 {
			t.Error("Bitfinex Orderbook.bids.Amount is negative")
		}
		price, err2 := strconv.ParseFloat(bids.Price, 64)
		if err2 != nil {
			t.Error("Cannot convert Bitfinex Orderbook.bids.Price into a float64")
		}
		if price < 0 {
			t.Error("Bitfinex Orderbook.bids.Price is negative")
		}
		timestamp, err3 := strconv.ParseFloat(bids.Timestamp, 64)
		if err3 != nil {
			t.Error("Cannot convert Bitfinex Orderbook.bids.timestamp into a float64")
		}
		if timestamp <= 0 {
			t.Error("Bitfinex Orderbook.bids.Amount is negative or 0")
		}
	}
}

//Live Testing
func TestGetTrades(t *testing.T) {
	t.Parallel()
	BitfinexGetTrades := Bitfinex{}

	trades, err := BitfinexGetTrades.GetTrades("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetTrades init error: ", err)
	}
	if reflect.ValueOf(trades[0]).NumField() != 6 {
		t.Error("BitfinexGetTrades struct change/or updated")
	}
	if reflect.TypeOf(trades[0].Amount).String() != "string" {
		t.Error("Bitfinex GetGetTrades.Amount is not a string")
	}
	if reflect.TypeOf(trades[0].Exchange).String() != "string" {
		t.Error("Bitfinex GetGetTrades.Exchange is not a string")
	}
	if reflect.TypeOf(trades[0].Price).String() != "string" {
		t.Error("Bitfinex GetGetTrades.Price is not a string")
	}
	if reflect.TypeOf(trades[0].Tid).String() != "int64" {
		t.Error("Bitfinex GetGetTrades.Tid is not a int64")
	}
	if reflect.TypeOf(trades[0].Timestamp).String() != "int64" {
		t.Error("Bitfinex GetGetTrades.Timestamp is not a int64")
	}
	if reflect.TypeOf(trades[0].Type).String() != "string" {
		t.Error("Bitfinex GetGetTrades.Type is not a string")
	}

	for _, explicitTrades := range trades {
		amount, err := strconv.ParseFloat(explicitTrades.Amount, 64)
		if err != nil {
			t.Error("Cannot convert Bitfinex GetTrades.Amount into a float64")
		}
		if amount <= 0 {
			t.Error("Bitfinex GetTrades.Amount is negative or 0")
		}
		if explicitTrades.Exchange != "bitfinex" {
			t.Error("Bitfinex GetTrades.Exchange incorrect name")
		}
		price, err2 := strconv.ParseFloat(explicitTrades.Price, 64)
		if err2 != nil {
			t.Error("Cannot convert Bitfinex GetTrades.Price into a float64")
		}
		if price <= 0 {
			t.Error("Bitfinex GetTrades.Price is negative or 0")
		}
		if explicitTrades.Tid <= 0 {
			t.Error("Bitfinex GetTrades.Tid is negative or 0")
		}
		if explicitTrades.Timestamp <= 0 {
			t.Error("Bitfinex GetTrades.Timestamp is negative or 0")
		}
		if explicitTrades.Type != "buy" && explicitTrades.Type != "sell" {
			t.Error("Bitfinex GetTrades.Type is wrong")
		}
	}
}

//Live Testing
func TestGetLends(t *testing.T) {
	t.Parallel()
	BitfinexGetLends := Bitfinex{}

	lends, err := BitfinexGetLends.GetLends("BTC", url.Values{})
	if err != nil {
		t.Error("BitfinexGetLends init error: ", err)
	}
	if reflect.ValueOf(lends[0]).NumField() != 4 {
		t.Error("BitfinexGetLends struct change/or updated")
	}
	if reflect.TypeOf(lends[0].AmountLent).String() != "float64" {
		t.Error("Bitfinex GetGetLends.AmountLent is not a float64")
	}
	if reflect.TypeOf(lends[0].AmountUsed).String() != "float64" {
		t.Error("Bitfinex GetGetLends.AmountUsed is not a float64")
	}
	if reflect.TypeOf(lends[0].Rate).String() != "float64" {
		t.Error("Bitfinex GetGetLends.Rate is not a float64")
	}
	if reflect.TypeOf(lends[0].Timestamp).String() != "int64" {
		t.Error("Bitfinex GetGetLends.Timestamp is not a int64")
	}

	for _, explicitLends := range lends {
		if explicitLends.AmountLent <= 0 {
			t.Error("Bitfinex GetLends.AmountLent is negative or 0")
		}
		if explicitLends.AmountUsed <= 0 {
			t.Error("Bitfinex GetLends.AmountUsed is negative or 0")
		}
		if explicitLends.Rate <= 0 {
			t.Error("Bitfinex GetLends.Rate is negative or 0")
		}
		if explicitLends.Timestamp <= 0 {
			t.Error("Bitfinex GetLends.Timestamp is negative or 0")
		}
	}
}

//Live Testing
func TestGetSymbols(t *testing.T) {
	t.Parallel()
	BitfinexGetSymbols := Bitfinex{}

	symbols, err := BitfinexGetSymbols.GetSymbols()
	if err != nil {
		t.Error("BitfinexGetSymbols init error: ", err)
	}
	if reflect.TypeOf(symbols[0]).String() != "string" {
		t.Error("Bitfinex GetSymbols is not a string")
	}

	expectedCurrencies := []string{
		"rrtbtc",
		"zecusd",
		"zecbtc",
		"xmrusd",
		"xmrbtc",
		"dshusd",
		"dshbtc",
		"bccbtc",
		"bcubtc",
		"bccusd",
		"bcuusd",
		"btcusd",
		"ltcusd",
		"ltcbtc",
		"ethusd",
		"ethbtc",
		"etcbtc",
		"etcusd",
		"bfxusd",
		"bfxbtc",
		"rrtusd",
	}
	if len(expectedCurrencies) <= len(symbols) {

		for _, explicitSymbol := range expectedCurrencies {
			if common.DataContains(expectedCurrencies, explicitSymbol) {
				break
			} else {
				t.Error("BitfinexGetSymbols currency mismatch with: ", explicitSymbol)
			}
		}
	} else {
		t.Error("BitfinexGetSymbols currency mismatch, Expected Currencies < Exchange Currencies")
	}
}

//Live Testing
func TestGetSymbolsDetails(t *testing.T) {
	t.Parallel()
	BitfinexGetSymbolsDetails := Bitfinex{}

	symbolDetails, err := BitfinexGetSymbolsDetails.GetSymbolsDetails()
	if err != nil {
		t.Error("BitfinexGetSymbolsDetails init error: ", err)
	}
	if reflect.ValueOf(symbolDetails[0]).NumField() != 7 {
		t.Error("BitfinexGetSymbolsDetails struct change/or updated")
	}
	if reflect.TypeOf(symbolDetails[0].Expiration).String() != "string" {
		t.Error("Bitfinex GetSymbolsDetails.Expiration is not a string")
	}
	if reflect.TypeOf(symbolDetails[0].InitialMargin).String() != "float64" {
		t.Error("Bitfinex GetSymbolsDetails.InitialMargin is not a float64")
	}
	if reflect.TypeOf(symbolDetails[0].MaximumOrderSize).String() != "float64" {
		t.Error("Bitfinex GetSymbolsDetails.MaximumOrderSize is not a float64")
	}
	if reflect.TypeOf(symbolDetails[0].MinimumMargin).String() != "float64" {
		t.Error("Bitfinex GetSymbolsDetails.MinimumMargin is not a float64")
	}
	if reflect.TypeOf(symbolDetails[0].MinimumOrderSize).String() != "float64" {
		t.Error("Bitfinex GetSymbolsDetails.MinimumOrderSize is not a float64")
	}
	if reflect.TypeOf(symbolDetails[0].Pair).String() != "string" {
		t.Error("Bitfinex GetSymbolsDetails.Pair is not a string")
	}
	if reflect.TypeOf(symbolDetails[0].PricePrecision).String() != "int" {
		t.Error("Bitfinex GetSymbolsDetails.PricePrecision is not a int")
	}

	for _, explicitDetails := range symbolDetails {
		if explicitDetails.Expiration != "NA" {
			expiration, err := strconv.ParseFloat(explicitDetails.Expiration, 64)
			if err != nil {
				t.Error("Cannot convert Bitfinex GetSymbolsDetails.Expiration into a float64")
			}
			if expiration < 0 {
				t.Error("Bitfinex GetSymbolsDetails.Expiration is negative")
			}
		}
		if explicitDetails.InitialMargin <= 0 {
			t.Error("Bitfinex GetSymbolsDetails.InitialMargin is negative or 0")
		}
		if explicitDetails.MaximumOrderSize <= 0 {
			t.Error("Bitfinex GetSymbolsDetails.MaximumOrderSize is negative or 0")
		}
		if explicitDetails.MinimumMargin <= 0 {
			t.Error("Bitfinex GetSymbolsDetails.MinimumMargin is negative or 0")
		}
		if explicitDetails.MinimumOrderSize <= 0 {
			t.Error("Bitfinex GetSymbolsDetails.MinimumOrderSize is negative or 0")
		}
		if len(explicitDetails.Pair) != 6 {
			t.Error("Bitfinex GetSymbolsDetails.Pair incorrect length")
		}
		if explicitDetails.PricePrecision <= 0 {
			t.Error("Bitfinex GetSymbolsDetails.PricePrecision is negative or 0")
		}
	}
}

//Hybrid Testing
func TestGetAccountInfo(t *testing.T) {
	expectedCryptoCurrencies := []string{
		"BTC",
		"LTC",
		"ETH",
		"ETC",
		"ZEC",
		"XMR",
		"DSH",
	}

	if ACCOUNT_LIVE_TEST { //Live Test
		newConfig := config.Config{}

		err := newConfig.LoadConfig("../../testdata/configtest.dat")
		if err != nil {
			t.Errorf("Test Failed - New Order init error: %s\n", err)
		}
		exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
		if err != nil {
			t.Errorf("Test Failed - New Order init error: %s\n", err)
		}

		BitfinexGetAccountInfo := Bitfinex{}
		BitfinexGetAccountInfo.Setup(exchangeConfig)

		response, err := BitfinexGetAccountInfo.GetAccountInfo()
		if err != nil {
			newErrString := fmt.Sprintf("TestGetAccountInfo: \nError: %s\n", err)
			t.Error(newErrString)
			response = append(response, BitfinexAccountInfo{})
		}

		if reflect.ValueOf(response[0]).NumField() != 3 {
			t.Error("BitfinexGetAccountInfo struct change/or updated")
		}
		if reflect.TypeOf(response[0].MakerFees).String() != "string" {
			t.Error("Bitfinex GetAccountInfo.MakerFees is not a string")
		}
		if reflect.TypeOf(response[0].TakerFees).String() != "string" {
			t.Error("Bitfinex GetAccountInfo.TakerFees is not a string")
		}

		if len(expectedCryptoCurrencies) == len(response[0].Fees) {
			if !common.DataContains(expectedCryptoCurrencies, response[0].Fees[0].Pairs) {
				t.Error("Bitfinex GetAccountInfo currency mismatch")
			}
		} else if len(expectedCryptoCurrencies) > len(response[0].Fees) {
			t.Error("BitfinexGetSymbols currency mismatch, Expected Currencies > Exchange Currencies")
		} else {
			t.Error("BitfinexGetSymbols currency mismatch, Expected Currencies < Exchange Currencies")
		}

		if len(response[0].Fees) != 7 {
			t.Error("Bitfinex GetAccountInfo.Fees incorrect length")
		}

		for _, explicitAI := range response {
			makerFees, err := strconv.ParseFloat(explicitAI.MakerFees, 64)
			if err != nil {
				t.Error("Cannot convert Bitfinex GetAccountInfo.MakerFees into float64")
			}
			if makerFees < 0 {
				t.Error("Bitfinex GetAccountInfo.MakerFees is negative")
			}

			takerFees, err := strconv.ParseFloat(explicitAI.TakerFees, 64)
			if err != nil {
				t.Error("Cannot convert Bitfinex GetAccountInfo.TakerFees into float64")
			}
			if takerFees < 0 {
				t.Error("Bitfinex GetAccountInfo.TakerFees is negative")
			}

			for _, fees := range explicitAI.Fees {
				MakerFees, err := strconv.ParseFloat(fees.MakerFees, 64)
				if err != nil {
					t.Error("Cannot convert Bitfinex GetAccountInfo.Fees.MakerFees into float64")
				}
				if MakerFees < 0 {
					t.Error("Bitfinex GetAccountInfo.Fees.MakerFees is negative")
				}
				TakerFees, err := strconv.ParseFloat(fees.TakerFees, 64)
				if err != nil {
					t.Error("Cannot convert Bitfinex GetAccountInfo.Fees.TakerFees into float64")
				}
				if TakerFees < 0 {
					t.Error("Bitfinex GetAccountInfo.Fees.TakerFees is negative")
				}
			}
		}

	} else { //Non-Live Test
		type Fees struct {
			Pairs     string `json:"pairs"`
			MakerFees string `json:"maker_fees"`
			TakerFees string `json:"taker_fees"`
		}
		accountInfoNonLive := [1]BitfinexAccountInfo{}
		accountInfoNonLive[0].MakerFees = "0.1"
		accountInfoNonLive[0].TakerFees = "0.2"
		nonLiveFees := Fees{}
		nonLiveFees.MakerFees = "0.1"
		nonLiveFees.Pairs = "BTC"
		nonLiveFees.TakerFees = "0.2"
		accountInfoNonLive[0].Fees = append(accountInfoNonLive[0].Fees, nonLiveFees)

		if reflect.ValueOf(accountInfoNonLive[0]).NumField() != 3 {
			t.Error("BitfinexGetAccountInfo struct change/or updated")
		}
		if reflect.TypeOf(accountInfoNonLive[0].MakerFees).String() != "string" {
			t.Error("Bitfinex GetAccountInfo.MakerFees is not a string")
		}
		if reflect.TypeOf(accountInfoNonLive[0].TakerFees).String() != "string" {
			t.Error("Bitfinex GetAccountInfo.TakerFees is not a string")
		}

		for _, explicitAI := range accountInfoNonLive {
			makerFees, err := strconv.ParseFloat(explicitAI.MakerFees, 64)
			if err != nil {
				t.Error("Cannot convert Bitfinex GetAccountInfo.MakerFees into float64")
			}
			if makerFees < 0 {
				t.Error("Bitfinex GetAccountInfo.MakerFees is negative")
			}

			takerFees, err := strconv.ParseFloat(explicitAI.TakerFees, 64)
			if err != nil {
				t.Error("Cannot convert Bitfinex GetAccountInfo.TakerFees into float64")
			}
			if takerFees < 0 {
				t.Error("Bitfinex GetAccountInfo.TakerFees is negative")
			}
			if len(explicitAI.Fees) != 1 {
				t.Error("Bitfinex GetAccountInfo.Fees.Pairs incorrect length")
			}

			for _, fees := range explicitAI.Fees {
				MakerFees, err := strconv.ParseFloat(fees.MakerFees, 64)
				if err != nil {
					t.Error("Cannot convert Bitfinex GetAccountInfo.Fees.MakerFees into float64")
				}
				if MakerFees < 0 {
					t.Error("Bitfinex GetAccountInfo.Fees.MakerFees is negative")
				}
				TakerFees, err := strconv.ParseFloat(fees.TakerFees, 64)
				if err != nil {
					t.Error("Cannot convert Bitfinex GetAccountInfo.Fees.TakerFees into float64")
				}
				if TakerFees < 0 {
					t.Error("Bitfinex GetAccountInfo.Fees.TakerFees is negative")
				}
			}
		}
	}
}

//Hybrid Testing
func TestNewDeposit(t *testing.T) {

	applicableMethods := []string{
		"bitcoin_address",
		"litecoin_address",
		"ethereum_address",
		"mastercoin_address", //Requires verified account
		"ethereumc_address",
		"zcash_address",
		"monero_address",
	}
	expectedCryptoCurrencies := []string{
		"BTC",
		"LTC",
		"ETH",
		"ETC",
		"ZEC",
		"XMR",
		"DSH",
	}

	if ACCOUNT_LIVE_TEST { //Live Test
		newConfig := config.Config{}

		err := newConfig.LoadConfig("../../testdata/configtest.dat")
		if err != nil {
			t.Errorf("Test Failed - New Order init error: %s\n", err)
		}
		exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
		if err != nil {
			t.Errorf("Test Failed - New Order init error: %s\n", err)
		}

		BitfinexNewDeposit := Bitfinex{}
		BitfinexNewDeposit.Setup(exchangeConfig)

		liveResponse, err := BitfinexNewDeposit.NewDeposit("bitcoin", "deposit", 0)
		if err != nil {
			t.Error("BitfinexNewDeposit init error: ", err)
		}

		if reflect.ValueOf(liveResponse).NumField() != 4 {
			t.Error("BitfinexNewDeposit struct change/or updated")
		}
		if reflect.TypeOf(liveResponse.Address).String() != "string" {
			t.Error("Bitfinex NewDeposit.Address is not a string")
		}
		if reflect.TypeOf(liveResponse.Currency).String() != "string" {
			t.Error("Bitfinex NewDeposit.Currency is not a string")
		}
		if reflect.TypeOf(liveResponse.Method).String() != "string" {
			t.Error("Bitfinex NewDeposit.Method) is not a string")
		}
		if reflect.TypeOf(liveResponse.Result).String() != "string" {
			t.Error("Bitfinex NewDeposit.Result is not a string")
		}
		if len(liveResponse.Address) != 34 {
			t.Error("Bitfinex NewDeposit.Address is incorrect")
		}
		if !common.DataContains(expectedCryptoCurrencies, liveResponse.Currency) {
			t.Error("Bitfinex NewDeposit.Currency currency mismatch" + liveResponse.Currency)
		}
		if !common.DataContains(applicableMethods, liveResponse.Method) {
			t.Error("Bitfinex NewDeposit.Method method mismatch")
		}
		if liveResponse.Result != "" && liveResponse.Result != "success" {
			t.Error("Bitfinex NewDeposit.Result " + liveResponse.Result)
		}

	} else { //Non-Live Test
		nonLiveResponse := BitfinexDepositResponse{}
		nonLiveResponse.Address = "1DPUgBaZoKbL38BEC1A3exPKCDZjQpnBa1"
		nonLiveResponse.Currency = "BTC"
		nonLiveResponse.Method = "bitcoin_address"
		nonLiveResponse.Result = ""

		if reflect.ValueOf(nonLiveResponse).NumField() != 4 {
			t.Error("BitfinexNewDeposit struct change/or updated")
		}
		if reflect.TypeOf(nonLiveResponse.Address).String() != "string" {
			t.Error("Bitfinex NewDeposit.Address is not a string")
		}
		if reflect.TypeOf(nonLiveResponse.Currency).String() != "string" {
			t.Error("Bitfinex NewDeposit.Currency is not a string")
		}
		if reflect.TypeOf(nonLiveResponse.Method).String() != "string" {
			t.Error("Bitfinex NewDeposit.Method) is not a string")
		}
		if reflect.TypeOf(nonLiveResponse.Result).String() != "string" {
			t.Error("Bitfinex NewDeposit.Result is not a string")
		}

		if len(nonLiveResponse.Address) != 34 {
			t.Error("Bitfinex NewDeposit.Address is incorrect")
		}
		if !common.DataContains(expectedCryptoCurrencies, nonLiveResponse.Currency) {
			t.Error("Bitfinex NewDeposit.Currency currency mismatch")
		}
		if !common.DataContains(applicableMethods, nonLiveResponse.Method) {
			t.Error("Bitfinex NewDeposit.Method method mismatch")
		}
		if nonLiveResponse.Result != "" && nonLiveResponse.Result != "success" {
			t.Error("Bitfinex NewDeposit.Result " + nonLiveResponse.Result)
		}
	}
}

//Non-Live Testing
func TestNewOrder(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - New Order init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - New Order init error: %s\n", err)
	}

	BitfinexNewOrder := Bitfinex{}
	BitfinexNewOrder.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		response, errLive := BitfinexNewOrder.NewOrder("BTCUSD", 0, 0, true, "test", false)
		if errLive == nil {
			t.Errorf("Test Failed - BitfinexNewOrder - Error: Expected Error Status: %t", response.IsLive)
		}
	}

	nonLiveResponse := BitfinexOrder{}
	nonLiveResponse.AverageExecutionPrice = 0.0
	nonLiveResponse.Exchange = "bitfinex"
	nonLiveResponse.ExecutedAmount = 0.0
	nonLiveResponse.ID = 448364249
	nonLiveResponse.IsCancelled = false
	nonLiveResponse.IsHidden = false
	nonLiveResponse.IsLive = true
	nonLiveResponse.OrderID = 448364249
	nonLiveResponse.OriginalAmount = 0.01
	nonLiveResponse.Price = 0.01
	nonLiveResponse.RemainingAmount = 0.01
	nonLiveResponse.Side = "buy"
	nonLiveResponse.Symbol = "btcusd"
	nonLiveResponse.Timestamp = "1444272165.252370982"
	nonLiveResponse.Type = "exchange limit"
	nonLiveResponse.WasForced = false

	if reflect.ValueOf(nonLiveResponse).NumField() != 16 {
		t.Error("BitfinexNewDeposit struct change/or updated")
	}
	if reflect.TypeOf(nonLiveResponse.AverageExecutionPrice).String() != "float64" {
		t.Error("Bitfinex NewOrder.AverageExecutionPrice is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Exchange).String() != "string" {
		t.Error("Bitfinex NewOrder.Exchange is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.ExecutedAmount).String() != "float64" {
		t.Error("Bitfinex NewOrder.ExecutedAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.OrderID).String() != "int64" {
		t.Error("Bitfinex NewOrder.ID is not an int64")
	}
	if reflect.TypeOf(nonLiveResponse.IsCancelled).String() != "bool" {
		t.Error("Bitfinex NewOrder.IsCancelled is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.IsHidden).String() != "bool" {
		t.Error("Bitfinex NewOrder.IsHidden is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.IsLive).String() != "bool" {
		t.Error("Bitfinex NewOrder.IsLive is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.OrderID).String() != "int64" {
		t.Error("Bitfinex NewOrder.OrderID is not an int64")
	}
	if reflect.TypeOf(nonLiveResponse.OriginalAmount).String() != "float64" {
		t.Error("Bitfinex NewOrder.OriginalAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Price).String() != "float64" {
		t.Error("Bitfinex NewOrder.Price is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.RemainingAmount).String() != "float64" {
		t.Error("Bitfinex NewOrder.RemainingAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Side).String() != "string" {
		t.Error("Bitfinex NewOrder.Side is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Symbol).String() != "string" {
		t.Error("Bitfinex NewOrder.Address is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Timestamp).String() != "string" {
		t.Error("Bitfinex NewOrder.Timestamp is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Type).String() != "string" {
		t.Error("Bitfinex NewOrder.Type is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.WasForced).String() != "bool" {
		t.Error("Bitfinex NewOrder.WasForced is not a bool")
	}

	if nonLiveResponse.AverageExecutionPrice < 0 {
		t.Error("Bitfinex NewOrder.AverageExecutionPrice is negative")
	}
	if nonLiveResponse.Exchange != "bitfinex" {
		t.Error("Bitfinex NewOrder.AverageExecutionPrice wrong exchange name")
	}
	if nonLiveResponse.ExecutedAmount < 0 {
		t.Error("Bitfinex NewOrder.ExecutedAmount is negative or 0")
	}
	if nonLiveResponse.ID <= 0 {
		t.Error("Bitfinex NewOrder.ID is negative or 0")
	}
	if nonLiveResponse.OrderID <= 0 {
		t.Error("Bitfinex NewOrder.OrderID is negative or 0")
	}
	if nonLiveResponse.OriginalAmount <= 0 {
		t.Error("Bitfinex NewOrder.OriginalAmount is negative or 0")
	}
	if nonLiveResponse.Price <= 0 {
		t.Error("Bitfinex NewOrder.Price is negative or 0")
	}
	if nonLiveResponse.RemainingAmount <= 0 {
		t.Error("Bitfinex NewOrder.RemainingAmount is negative or 0")
	}
	nonLiveTimestamp, err := strconv.ParseFloat(nonLiveResponse.Timestamp, 64)
	if err != nil {
		t.Error("Bitfinex NewOrder.Timestamp cannot convert to float64")
	}
	if nonLiveTimestamp <= 0 {
		t.Error("Bitfinex NewOrder.Timestamp is negative or 0")
	}
}

//Non-Live Testing
func TestNewOrderMulti(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - New Order init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - New Order init error: %s\n", err)
	}

	BitfinexNewOrderMulti := Bitfinex{}
	BitfinexNewOrderMulti.Setup(exchangeConfig)

	var orders []BitfinexPlaceOrder
	order := BitfinexPlaceOrder{}
	order.Amount = 0.0
	order.Exchange = "bitfinex"
	order.Price = 0.0
	order.Side = "test"
	order.Symbol = "BTCUSD"
	order.Type = "test"
	orders = append(orders, order)

	if ACCOUNT_LIVE_TEST {
		response, err := BitfinexNewOrderMulti.NewOrderMulti(orders)
		if err == nil {
			newErrString := fmt.Sprintf("BitfinexNewOrderMulti - Error: Expected Error Status: %s\n", response.Status)
			t.Error(newErrString)
		}
	}

	nonLiveResponse := BitfinexOrderMultiResponse{}
	nonLiveResponse.Status = "success"

	orderTest := BitfinexOrder{}
	orderTest.AverageExecutionPrice = 0.0
	orderTest.Exchange = "bitfinex"
	orderTest.ExecutedAmount = 0.0
	orderTest.ID = 448364249
	orderTest.IsCancelled = false
	orderTest.IsHidden = false
	orderTest.IsLive = true
	orderTest.OrderID = 448364249
	orderTest.OriginalAmount = 0.01
	orderTest.Price = 0.01
	orderTest.RemainingAmount = 0.01
	orderTest.Side = "buy"
	orderTest.Symbol = "btcusd"
	orderTest.Timestamp = "1444272165.252370982"
	orderTest.Type = "exchange limit"
	orderTest.WasForced = false

	nonLiveResponse.Orders = append(nonLiveResponse.Orders, orderTest)

	if reflect.ValueOf(nonLiveResponse).NumField() != 2 {
		t.Error("Bitfinex NewOrderMulti struct change/or updated")
	}
	if reflect.TypeOf(nonLiveResponse.Status).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Status is not a string")
	}
	if reflect.ValueOf(nonLiveResponse.Orders[0]).NumField() != 16 {
		t.Error("Bitfinex NewOrderMulti struct change/or updated")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].AverageExecutionPrice).String() != "float64" {
		t.Error("Bitfinex NewOrderMulti.AverageExecutionPrice is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Exchange).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Exchange is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].ExecutedAmount).String() != "float64" {
		t.Error("Bitfinex NewOrderMulti.ExecutedAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].OrderID).String() != "int64" {
		t.Error("Bitfinex NewOrderMulti.ID is not an int64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].IsCancelled).String() != "bool" {
		t.Error("Bitfinex NewOrderMulti.IsCancelled is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].IsHidden).String() != "bool" {
		t.Error("Bitfinex NewOrderMulti.IsHidden is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].IsLive).String() != "bool" {
		t.Error("Bitfinex NewOrderMulti.IsLive is not a bool")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].OrderID).String() != "int64" {
		t.Error("Bitfinex NewOrderMulti.OrderID is not an int64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].OriginalAmount).String() != "float64" {
		t.Error("Bitfinex NewOrderMulti.OriginalAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Price).String() != "float64" {
		t.Error("Bitfinex NewOrderMulti.Price is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].RemainingAmount).String() != "float64" {
		t.Error("Bitfinex NewOrderMulti.RemainingAmount is not a float64")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Side).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Side is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Symbol).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Address is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Timestamp).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Timestamp is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].Type).String() != "string" {
		t.Error("Bitfinex NewOrderMulti.Type is not a string")
	}
	if reflect.TypeOf(nonLiveResponse.Orders[0].WasForced).String() != "bool" {
		t.Error("Bitfinex NewOrderMulti.WasForced is not a bool")
	}
}

func TestCancelOrder(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelOrder init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelOrder init error: %s\n", err)
	}

	BitfinexCancelOrder := Bitfinex{}
	BitfinexCancelOrder.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexCancelOrder.CancelOrder(1337)
		if err == nil {
			t.Errorf("Test Failed - Bitfinex CancelOrder - Error: %s", err)
		}
	}
}

func TestCancelMultipleOrders(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelMultipleOrders init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelMultipleOrders init error: %s\n", err)
	}

	BitfinexMultipleOrders := Bitfinex{}
	BitfinexMultipleOrders.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		orders := []int64{1336, 1337}
		response, err := BitfinexMultipleOrders.CancelMultipleOrders(orders)
		if err != nil || response != "" {
			t.Errorf("Test Failed - Bitfinex CancelMultipleOrders - Error: %s", err)
		}
	}
}

func TestCancelAllOrders(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelAllOrders init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CancelAllOrders init error: %s\n", err)
	}

	BitfinexCancelAllOrders := Bitfinex{}
	BitfinexCancelAllOrders.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		response, err := BitfinexCancelAllOrders.CancelAllOrders()
		if err != nil || response != "" {
			t.Errorf("Test Failed - Bitfinex CancelAllOrders - Error: %s", err)
		}
	}
}

func TestReplaceOrder(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex ReplaceOrder init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex ReplaceOrder init error: %s\n", err)
	}

	BitfinexReplaceOrder := Bitfinex{}
	BitfinexReplaceOrder.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexReplaceOrder.ReplaceOrder(1337, "BTC", 0, 0, true, "exchange limit", false)
		if err == nil {
			t.Error("Test Failed - Bitfinex ReplaceOrder - Expected Error")
		}
	}
}

func TestGetOrderStatus(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetOrderStatus init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetOrderStatus init error: %s\n", err)
	}

	BitfinexGetOrderStatus := Bitfinex{}
	BitfinexGetOrderStatus.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetOrderStatus.GetOrderStatus(1337)
		if err == nil {
			t.Error("Test Failed - Bitfinex GetOrderStatus - Expected Error")
		}
	}
}

func TestGetActiveOrders(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveOrders init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveOrders init error: %s\n", err)
	}

	BitfinexGetActiveOrders := Bitfinex{}
	BitfinexGetActiveOrders.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetActiveOrders.GetActiveOrders()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetActiveOrders - Expected Error")
		}
	}
}

func TestGetActivePositions(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActivePositions init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActivePositions init error: %s\n", err)
	}

	BitfinexGetActivePositions := Bitfinex{}
	BitfinexGetActivePositions.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetActivePositions.GetActivePositions()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetActivePositions - Expected Error")
		}
	}
}

func TestClaimPosition(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex ClaimPosition init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex ClaimPosition init error: %s\n", err)
	}

	BitfinexClaimPosition := Bitfinex{}
	BitfinexClaimPosition.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexClaimPosition.ClaimPosition(1337)
		if err == nil {
			t.Error("Test Failed - Bitfinex ClaimPosition - Expected Error")
		}
	}
}

func TestGetBalanceHistory(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetBalanceHistory init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetBalanceHistory init error: %s\n", err)
	}

	BitfinexGetBalanceHistory := Bitfinex{}
	BitfinexGetBalanceHistory.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetBalanceHistory.GetBalanceHistory("BTC", time.Now(), time.Now(), 1, "testWallet")
		if err == nil {
			t.Error("Test Failed - Bitfinex GetBalanceHistory - Expected Error")
		}
	}
}

func TestGetMovementHistory(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMovementHistory init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMovementHistory init error: %s\n", err)
	}

	BitfinexGetMovementHistory := Bitfinex{}
	BitfinexGetMovementHistory.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetMovementHistory.GetMovementHistory("BTC", "BITCOIN", time.Now(), time.Now(), 1)
		if err == nil {
			t.Error("Test Failed - Bitfinex GetMovementHistory - Expected Error")
		}
	}
}

func TestGetTradeHistory(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetTradeHistory init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetTradeHistory init error: %s\n", err)
	}

	BitfinexGetTradeHistory := Bitfinex{}
	BitfinexGetTradeHistory.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetTradeHistory.GetTradeHistory("BTC", time.Now(), time.Now(), 1, 0)
		if err == nil {
			t.Error("Test Failed - Bitfinex GetTradeHistory - Expected Error")
		}
	}
}

func TestNewOffer(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex NewOffer init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex NewOffer init error: %s\n", err)
	}

	BitfinexNewOffer := Bitfinex{}
	BitfinexNewOffer.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		response := BitfinexNewOffer.NewOffer("BTC", 1, 2, 2, "buy")
		if response != 0 {
			t.Error("Test Failed - Bitfinex NewOffer - Expected Error")
		}
	}
}

func TestGetOfferStatus(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetOfferStatus init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetOfferStatus init error: %s\n", err)
	}

	BitfinexGetOfferStatus := Bitfinex{}
	BitfinexGetOfferStatus.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetOfferStatus.GetOfferStatus(1337)
		if err == nil {
			t.Error("Test Failed - Bitfinex GetOfferStatus - Expected Error")
		}
	}
}

func TestGetActiveOffers(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveOffers init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveOffers init error: %s\n", err)
	}

	BitfinexGetActiveOffers := Bitfinex{}
	BitfinexGetActiveOffers.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetActiveOffers.GetActiveOffers()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetActiveOffers - Expected Error")
		}
	}
}

func TestGetActiveMarginFunding(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveMarginFunding init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetActiveMarginFunding init error: %s\n", err)
	}

	BitfinexGetActiveMarginFunding := Bitfinex{}
	BitfinexGetActiveMarginFunding.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetActiveMarginFunding.GetActiveMarginFunding()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetActiveMarginFunding - Expected Error")
		}
	}
}

func TestGetMarginTotalTakenFunds(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMarginTotalTakenFunds init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMarginTotalTakenFunds init error: %s\n", err)
	}

	BitfinexGetMarginTotalTakenFunds := Bitfinex{}
	BitfinexGetMarginTotalTakenFunds.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetMarginTotalTakenFunds.GetMarginTotalTakenFunds()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetMarginTotalTakenFunds - Expected Error")
		}
	}
}

func TestCloseMarginFunding(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CloseMarginFunding init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex CloseMarginFunding init error: %s\n", err)
	}

	BitfinexCloseMarginFunding := Bitfinex{}
	BitfinexCloseMarginFunding.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexCloseMarginFunding.CloseMarginFunding(1337)
		if err == nil {
			t.Error("Test Failed - Bitfinex CloseMarginFunding - Expected Error")
		}
	}
}

func TestGetAccountBalance(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetAccountBalance init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetAccountBalance init error: %s\n", err)
	}

	BitfinexGetAccountBalance := Bitfinex{}
	BitfinexGetAccountBalance.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetAccountBalance.GetAccountBalance()
		if err != nil {
			t.Error("Test Failed - Bitfinex GetAccountBalance - Expected Error")
		}
	}
}

func TestGetMarginInfo(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMarginInfo init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex GetMarginInfo init error: %s\n", err)
	}

	BitfinexGetMarginInfo := Bitfinex{}
	BitfinexGetMarginInfo.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexGetMarginInfo.GetMarginInfo()
		if err == nil {
			t.Error("Test Failed - Bitfinex GetMarginInfo - Expected Error")
		}
	}
}

func TestWalletTransfer(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex WalletTransfer init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex WalletTransfer init error: %s\n", err)
	}

	BitfinexWalletTransfer := Bitfinex{}
	BitfinexWalletTransfer.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexWalletTransfer.WalletTransfer(100, "BTC", "somewallet", "someotherwallet")
		if err == nil {
			t.Error("Test Failed - Bitfinex WalletTransfer - Expected Error")
		}
	}
}

func TestWithdrawal(t *testing.T) {
	newConfig := config.Config{}

	err := newConfig.LoadConfig("../../testdata/configtest.dat")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex Withdrawal init error: %s\n", err)
	}
	exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Errorf("Test Failed - Bitfinex Withdrawal init error: %s\n", err)
	}

	BitfinexWithdrawal := Bitfinex{}
	BitfinexWithdrawal.Setup(exchangeConfig)

	if ACCOUNT_LIVE_TEST {
		_, err := BitfinexWithdrawal.Withdrawal("BITCOIN", "somewallet", "123A87612376", 100)
		if err == nil {
			t.Error("Test Failed - Bitfinex Withdrawal - Expected Error")
		}
	}
}

func TestSendAuthenticatedHTTPRequest(t *testing.T) {
	if ACCOUNT_LIVE_TEST {
		newConfig := config.Config{}

		err := newConfig.LoadConfig("../../testdata/configtest.dat")
		if err != nil {
			t.Errorf("Test Failed - SendAuthenticatedHTTPRequest init error: %s\n", err)
		}
		exchangeConfig, err := newConfig.GetExchangeConfig("Bitfinex")
		if err != nil {
			t.Errorf("Test Failed - SendAuthenticatedHTTPRequest init error: %s\n", err)
		}

		sendAuthHTTPRequest := Bitfinex{}
		sendAuthHTTPRequest.Setup(exchangeConfig)
		result := []BitfinexAccountInfo{}

		err = sendAuthHTTPRequest.SendAuthenticatedHTTPRequest("POST", BITFINEX_ACCOUNT_INFO, nil, &result)
		if err != nil {
			t.Errorf("Test Failed - Bitfinex SendAuthenticatedHTTPRequest() error: %s", err)
		}

		if len(result) < 1 {
			t.Error("Test Failed - Bitfinex SendAuthenticatedHTTPRequest() incorrect length")
		}
	}
}
