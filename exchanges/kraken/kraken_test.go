package kraken

import (
	"testing"

	"github.com/thrasher-/gocryptotrader/config"
)

var k Kraken

// Please add your own APIkeys to do correct due diligence testing.
const (
	apiKey    = ""
	apiSecret = ""
	clientID  = ""
)

func TestSetDefaults(t *testing.T) {
	k.SetDefaults()
}

func TestSetup(t *testing.T) {
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	krakenConfig, err := cfg.GetExchangeConfig("Kraken")
	if err != nil {
		t.Error("Test Failed - kraken Setup() init error", err)
	}

	krakenConfig.AuthenticatedAPISupport = true
	krakenConfig.APIKey = apiKey
	krakenConfig.APISecret = apiSecret
	krakenConfig.ClientID = clientID

	k.Setup(krakenConfig)
}

func TestGetFee(t *testing.T) {
	t.Parallel()
	if k.GetFee(true) != 0.1 {
		t.Error("Test Failed - kraken GetFee() error")
	}
	if k.GetFee(false) != 0.35 {
		t.Error("Test Failed - kraken GetFee() error")
	}
}

func TestGetServerTime(t *testing.T) {
	t.Parallel()
	err := k.GetServerTime(false)
	if err != nil {
		t.Error("Test Failed - GetServerTime() error", err)
	}
	err = k.GetServerTime(true)
	if err != nil {
		t.Error("Test Failed - GetServerTime() error", err)
	}
}

func TestGetAssets(t *testing.T) {
	t.Parallel()
	err := k.GetAssets()
	if err != nil {
		t.Error("Test Failed - GetAssets() error", err)
	}
}

func TestGetAssetPairs(t *testing.T) {
	t.Parallel()
	err := k.GetAssetPairs(nil)
	if err != nil {
		t.Error("Test Failed - GetAssetPairs() error", err)
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := k.GetTicker("BCHEUR")
	if err != nil {
		t.Error("Test Failed - GetTicker() error", err)
	}
}

func TestGetOHLC(t *testing.T) {
	t.Parallel()
	_, err := k.GetOHLC("BCHEUR")
	if err != nil {
		t.Error("Test Failed - GetOHLC() error", err)
	}
}

func TestGetDepth(t *testing.T) {
	t.Parallel()
	_, err := k.GetDepth("BCHEUR")
	if err != nil {
		t.Error("Test Failed - GetDepth() error", err)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()
	_, err := k.GetTrades("BCHEUR")
	if err != nil {
		t.Error("Test Failed - GetTrades() error", err)
	}
}

func TestGetSpread(t *testing.T) {
	t.Parallel()
	_, err := k.GetSpread("BCHEUR")
	if err != nil {
		t.Error("Test Failed - GetSpread() error", err)
	}
}

func TestGetBalance(t *testing.T) {
	t.Parallel()
	err := k.GetBalance()
	if err == nil {
		t.Error("Test Failed - GetBalance() error", err)
	}
}

func TestGetTradeBalance(t *testing.T) {
	t.Parallel()
	err := k.GetTradeBalance("", "")
	if err == nil {
		t.Error("Test Failed - GetTradeBalance() error", err)
	}
}

func TestGetOpenOrders(t *testing.T) {
	t.Parallel()
	err := k.GetOpenOrders(true, 0)
	if err == nil {
		t.Error("Test Failed - GetOpenOrders() error", err)
	}
}

func TestGetClosedOrders(t *testing.T) {
	t.Parallel()
	err := k.GetClosedOrders(true, 0, 0, 0, 0, "")
	if err == nil {
		t.Error("Test Failed - GetClosedOrders() error", err)
	}
}

func TestQueryOrdersInfo(t *testing.T) {
	t.Parallel()
	err := k.QueryOrdersInfo(false, 0, 0)
	if err == nil {
		t.Error("Test Failed - QueryOrdersInfo() error", err)
	}
}

func TestGetTradesHistory(t *testing.T) {
	t.Parallel()
	args := GetTradesHistoryOptions{Trades: true, Start: "TMZEDR-VBJN2-NGY6DX", End: "TVRXG2-R62VE-RWP3UW"}
	_, err := k.GetTradesHistory(args)
	if err == nil {
		t.Error("Test Failed - GetTradesHistory() error", err)
	}
}

func TestQueryTrades(t *testing.T) {
	t.Parallel()
	err := k.QueryTrades(0, false)
	if err == nil {
		t.Error("Test Failed - QueryTrades() error", err)
	}
}

func TestOpenPositions(t *testing.T) {
	t.Parallel()
	_, err := k.OpenPositions(false)
	if err == nil {
		t.Error("Test Failed - OpenPositions() error", err)
	}
}

func TestGetLedgers(t *testing.T) {
	t.Parallel()
	args := GetLedgersOptions{Start: "LRUHXI-IWECY-K4JYGO", End: "L5NIY7-JZQJD-3J4M2V", Ofs: 15}
	_, err := k.GetLedgers(args)
	if err == nil {
		t.Error("Test Failed - GetLedgers() error", err)
	}
}

func TestQueryLedgers(t *testing.T) {
	t.Parallel()
	_, err := k.QueryLedgers("LVTSFS-NHZVM-EXNZ5M")
	if err == nil {
		t.Error("Test Failed - QueryLedgers() error", err)
	}
}

func TestGetTradeVolume(t *testing.T) {
	t.Parallel()
	_, err := k.GetTradeVolume(true, "OAVY7T-MV5VK-KHDF5X")
	if err == nil {
		t.Error("Test Failed - GetTradeVolume() error", err)
	}
}

func TestAddOrder(t *testing.T) {
	t.Parallel()
	_, err := k.AddOrder("XXBTZUSD", "sell", "market", 0.00000001, 0, 0, 0, AddOrderOptions{Oflags: "fcib"})
	if err == nil {
		t.Error("Test Failed - AddOrder() error", err)
	}
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	err := k.CancelOrder("OAVY7T-MV5VK-KHDF5X")
	if err == nil {
		t.Error("Test Failed - CancelOrder() error", err)
	}
}
