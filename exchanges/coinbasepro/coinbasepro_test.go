package coinbasepro

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
)

var c CoinbasePro

// Please supply your APIKeys here for better testing
const (
	apiKey    = ""
	apiSecret = ""
	clientID  = "" //passphrase you made at API CREATION
)

func TestSetDefaults(t *testing.T) {
	c.SetDefaults()
	c.Requester.SetRateLimit(false, time.Second, 1)
}

func TestSetup(t *testing.T) {
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	gdxConfig, err := cfg.GetExchangeConfig("CoinbasePro")
	if err != nil {
		t.Error("Test Failed - coinbasepro Setup() init error")
	}

	c.Setup(gdxConfig)
}

func TestGetFee(t *testing.T) {
	if c.GetFee(true).Equal(decimal.NewFromFloat(0.25)) {
		t.Error("Test failed - GetFee() error")
	}
	if common.EqualZero(c.GetFee(false)) {
		t.Error("Test failed - GetFee() error")
	}
}

func TestGetProducts(t *testing.T) {
	_, err := c.GetProducts()
	if err != nil {
		t.Error("Test failed - GetProducts() error")
	}
}

func TestGetTicker(t *testing.T) {
	_, err := c.GetTicker("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetTicker() error", err)
	}
}

func TestGetTrades(t *testing.T) {
	_, err := c.GetTrades("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetTrades() error", err)
	}
}

func TestGetHistoricRates(t *testing.T) {
	_, err := c.GetHistoricRates("BTC-USD", 0, 0, 0)
	if err != nil {
		t.Error("Test failed - GetHistoricRates() error", err)
	}
}

func TestGetStats(t *testing.T) {
	_, err := c.GetStats("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetStats() error", err)
	}
}

func TestGetCurrencies(t *testing.T) {
	_, err := c.GetCurrencies()
	if err != nil {
		t.Error("Test failed - GetCurrencies() error", err)
	}
}

func TestGetServerTime(t *testing.T) {
	_, err := c.GetServerTime()
	if err != nil {
		t.Error("Test failed - GetServerTime() error", err)
	}
}

func TestAuthRequests(t *testing.T) {

	if c.APIKey != "" && c.APISecret != "" && c.ClientID != "" {

		_, err := c.GetAccounts()
		if err == nil {
			t.Error("Test failed - GetAccounts() error", err)
		}

		_, err = c.GetAccount("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetAccount() error", err)
		}

		_, err = c.GetAccountHistory("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetAccountHistory() error", err)
		}

		_, err = c.GetHolds("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetHolds() error", err)
		}

		_, err = c.PlaceLimitOrder("", decimal.Zero, decimal.Zero, "buy", "", "", "BTC-USD", "", false)
		if err == nil {
			t.Error("Test failed - PlaceLimitOrder() error", err)
		}

		_, err = c.PlaceMarketOrder("", common.One, decimal.Zero, "buy", "BTC-USD", "")
		if err == nil {
			t.Error("Test failed - PlaceMarketOrder() error", err)
		}

		err = c.CancelOrder("1337")
		if err == nil {
			t.Error("Test failed - CancelOrder() error", err)
		}

		_, err = c.CancelAllOrders("BTC-USD")
		if err == nil {
			t.Error("Test failed - CancelAllOrders() error", err)
		}

		_, err = c.GetOrders([]string{"open", "done"}, "BTC-USD")
		if err == nil {
			t.Error("Test failed - GetOrders() error", err)
		}

		_, err = c.GetOrder("1337")
		if err == nil {
			t.Error("Test failed - GetOrders() error", err)
		}

		_, err = c.GetFills("1337", "BTC-USD")
		if err == nil {
			t.Error("Test failed - GetFills() error", err)
		}
		_, err = c.GetFills("", "")
		if err == nil {
			t.Error("Test failed - GetFills() error", err)
		}

		_, err = c.GetFundingRecords("rejected")
		if err == nil {
			t.Error("Test failed - GetFundingRecords() error", err)
		}

		// 	_, err := c.RepayFunding("1", "BTC")
		// 	if err != nil {
		// 		t.Error("Test failed - RepayFunding() error", err)
		// 	}

		_, err = c.MarginTransfer(common.One, "withdraw", "45fa9e3b-00ba-4631-b907-8a98cbdf21be", "BTC")
		if err == nil {
			t.Error("Test failed - MarginTransfer() error", err)
		}

		_, err = c.GetPosition()
		if err == nil {
			t.Error("Test failed - GetPosition() error", err)
		}

		_, err = c.ClosePosition(false)
		if err == nil {
			t.Error("Test failed - ClosePosition() error", err)
		}

		_, err = c.GetPayMethods()
		if err == nil {
			t.Error("Test failed - GetPayMethods() error", err)
		}

		_, err = c.DepositViaPaymentMethod(common.One, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - DepositViaPaymentMethod() error", err)
		}

		_, err = c.DepositViaCoinbase(common.One, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - DepositViaCoinbase() error", err)
		}

		_, err = c.WithdrawViaPaymentMethod(common.One, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - WithdrawViaPaymentMethod() error", err)
		}

		// 	_, err := c.WithdrawViaCoinbase(1, "BTC", "c13cd0fc-72ca-55e9-843b-b84ef628c198")
		// 	if err != nil {
		// 		t.Error("Test failed - WithdrawViaCoinbase() error", err)
		// 	}

		_, err = c.WithdrawCrypto(common.One, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - WithdrawViaCoinbase() error", err)
		}

		_, err = c.GetCoinbaseAccounts()
		if err == nil {
			t.Error("Test failed - GetCoinbaseAccounts() error", err)
		}

		_, err = c.GetReportStatus("1337")
		if err == nil {
			t.Error("Test failed - GetReportStatus() error", err)
		}

		_, err = c.GetTrailingVolume()
		if err == nil {
			t.Error("Test failed - GetTrailingVolume() error", err)
		}
	}
}
