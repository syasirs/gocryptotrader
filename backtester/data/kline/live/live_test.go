package live

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctkline "github.com/thrasher-corp/gocryptotrader/exchanges/kline"
)

const testExchange = "binance"

func TestLoadCandles(t *testing.T) {
	t.Parallel()
	interval := gctkline.OneHour
	cp1 := currency.NewPair(currency.BTC, currency.USDT)
	a := asset.Spot
	em := engine.SetupExchangeManager()
	exch, err := em.NewExchangeByName(testExchange)
	if err != nil {
		t.Fatal(err)
	}
	pFormat := &currency.PairFormat{Uppercase: true}
	b := exch.GetBase()
	exch.SetDefaults()
	b.CurrencyPairs.Pairs = make(map[asset.Item]*currency.PairStore)
	b.CurrencyPairs.Pairs[asset.Spot] = &currency.PairStore{
		Available:     currency.Pairs{cp1},
		Enabled:       currency.Pairs{cp1},
		AssetEnabled:  convert.BoolPtr(true),
		RequestFormat: pFormat,
		ConfigFormat:  pFormat,
	}
	var data *gctkline.Item
	data, err = LoadData(context.Background(), time.Now(), exch, common.DataCandle, interval.Duration(), cp1, currency.EMPTYPAIR, a)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Candles) == 0 {
		t.Error("expected candles")
	}
	_, err = LoadData(context.Background(), time.Now(), exch, -1, interval.Duration(), cp1, currency.EMPTYPAIR, a)
	if !errors.Is(err, common.ErrInvalidDataType) {
		t.Errorf("received: %v, expected: %v", err, common.ErrInvalidDataType)
	}
}

func TestLoadTrades(t *testing.T) {
	t.Parallel()
	interval := gctkline.OneMin
	cp1 := currency.NewPair(currency.BTC, currency.USDT)
	a := asset.Spot
	em := engine.SetupExchangeManager()
	exch, err := em.NewExchangeByName(testExchange)
	if err != nil {
		t.Fatal(err)
	}
	pFormat := &currency.PairFormat{Uppercase: true}
	b := exch.GetBase()
	exch.SetDefaults()
	b.CurrencyPairs.Pairs = make(map[asset.Item]*currency.PairStore)
	b.CurrencyPairs.Pairs[asset.Spot] = &currency.PairStore{
		Available:     currency.Pairs{cp1},
		Enabled:       currency.Pairs{cp1},
		AssetEnabled:  convert.BoolPtr(true),
		RequestFormat: pFormat,
		ConfigFormat:  pFormat,
	}
	var data *gctkline.Item
	data, err = LoadData(context.Background(), time.Now(), exch, common.DataTrade, interval.Duration(), cp1, currency.EMPTYPAIR, a)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Candles) == 0 {
		t.Error("expected candles")
	}
}

type hello struct {
	shutdown chan struct{}
	updated  chan struct{}
	timer    *time.Timer
}

func butts(yo *hello) {
	for {
		select {
		case <-yo.shutdown:
			return
		case <-yo.timer.C:
			yo.timer.Reset(time.Second)
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func TestButts(t *testing.T) {
	hi := hello{
		shutdown: make(chan struct{}),
		updated:  make(chan struct{}),
		timer:    time.NewTimer(0),
	}

}
