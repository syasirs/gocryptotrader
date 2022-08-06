package data

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/event"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/order"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
)

const testExchange = "binance"

type fakeDataHandler struct {
	time int
}

func TestLatest(t *testing.T) {
	t.Parallel()
	var d Base
	d.AppendStream(&fakeDataHandler{time: 1})
	if latest := d.Latest(); latest != d.stream[d.offset] {
		t.Error("expected latest to match offset")
	}
}

func TestBaseDataFunctions(t *testing.T) {
	t.Parallel()
	var d Base

	d.Next()
	o := d.Offset()
	if o != 0 {
		t.Error("expected 0")
	}
	d.AppendStream(nil)
	if d.IsLastEvent() {
		t.Error("no")
	}
	d.AppendStream(nil)
	if len(d.stream) != 0 {
		t.Error("expected 0")
	}
	d.AppendStream(&fakeDataHandler{time: 1})
	d.AppendStream(&fakeDataHandler{time: 2})
	d.AppendStream(&fakeDataHandler{time: 3})
	d.AppendStream(&fakeDataHandler{time: 4})
	d.Next()

	d.Next()
	if list := d.List(); len(list) != 2 {
		t.Errorf("expected 2 received %v", len(list))
	}
	d.Next()
	d.Next()
	if !d.IsLastEvent() {
		t.Error("expected last event")
	}
	o = d.Offset()
	if o != 4 {
		t.Error("expected 4")
	}
	if list := d.List(); len(list) != 0 {
		t.Error("expected 0")
	}
	if history := d.History(); len(history) != 4 {
		t.Errorf("expected 4 received %v", len(history))
	}

	d.SetStream(nil)
	if st := d.GetStream(); st != nil {
		t.Error("expected nil")
	}
	d.Reset()
	d.GetStream()
	d.SortStream()
}

func TestSetup(t *testing.T) {
	t.Parallel()
	d := HandlerPerCurrency{}
	d.Setup()
	if d.data == nil {
		t.Error("expected not nil")
	}
}

func TestSetDataForCurrency(t *testing.T) {
	t.Parallel()
	d := HandlerPerCurrency{}
	exch := testExchange
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d.SetDataForCurrency(exch, a, p, nil)
	if d.data == nil {
		t.Error("expected not nil")
	}
	if d.data[exch][a][p] != nil {
		t.Error("expected nil")
	}
}

func TestGetAllData(t *testing.T) {
	t.Parallel()
	d := HandlerPerCurrency{}
	exch := testExchange
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d.SetDataForCurrency(exch, a, p, nil)
	d.SetDataForCurrency(exch, a, currency.NewPair(currency.BTC, currency.DOGE), nil)
	result := d.GetAllData()
	if len(result) != 1 {
		t.Error("expected 1")
	}
	if len(result[exch][a]) != 2 {
		t.Error("expected 2")
	}
}

func TestGetDataForCurrency(t *testing.T) {
	t.Parallel()
	d := HandlerPerCurrency{}
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d.SetDataForCurrency(testExchange, a, p, nil)
	d.SetDataForCurrency(testExchange, a, currency.NewPair(currency.BTC, currency.DOGE), nil)
	ev := &order.Order{Base: &event.Base{
		Exchange:     testExchange,
		AssetType:    a,
		CurrencyPair: p,
	}}
	result, err := d.GetDataForCurrency(ev)
	if err != nil {
		t.Error(err)
	}
	if result != nil {
		t.Error("expected nil")
	}

	_, err = d.GetDataForCurrency(nil)
	if !errors.Is(err, common.ErrNilEvent) {
		t.Errorf("received '%v' expected '%v'", err, common.ErrNilEvent)
	}

	_, err = d.GetDataForCurrency(&order.Order{Base: &event.Base{
		Exchange:     "lol",
		AssetType:    asset.USDTMarginedFutures,
		CurrencyPair: currency.NewPair(currency.EMB, currency.DOGE),
	}})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Errorf("received '%v' expected '%v'", err, ErrHandlerNotFound)
	}

	_, err = d.GetDataForCurrency(&order.Order{Base: &event.Base{
		Exchange:     testExchange,
		AssetType:    asset.USDTMarginedFutures,
		CurrencyPair: currency.NewPair(currency.EMB, currency.DOGE),
	}})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Errorf("received '%v' expected '%v'", err, ErrHandlerNotFound)
	}
}

func TestReset(t *testing.T) {
	t.Parallel()
	d := HandlerPerCurrency{}
	exch := testExchange
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d.SetDataForCurrency(exch, a, p, nil)
	d.SetDataForCurrency(exch, a, currency.NewPair(currency.BTC, currency.DOGE), nil)
	d.Reset()
	if d.data != nil {
		t.Error("expected nil")
	}
}

// methods that satisfy the common.DataEventHandler interface
func (f fakeDataHandler) GetOffset() int64 {
	return 4
}

func (f fakeDataHandler) SetOffset(int64) {
}

func (f fakeDataHandler) IsEvent() bool {
	return false
}

func (f fakeDataHandler) GetTime() time.Time {
	return time.Now().Add(time.Hour * time.Duration(f.time))
}

func (f fakeDataHandler) Pair() currency.Pair {
	return currency.NewPair(currency.BTC, currency.USD)
}

func (f fakeDataHandler) GetExchange() string {
	return "fake"
}

func (f fakeDataHandler) GetInterval() kline.Interval {
	return kline.Interval(time.Minute)
}

func (f fakeDataHandler) GetAssetType() asset.Item {
	return asset.Spot
}

func (f fakeDataHandler) GetReason() string {
	return "fake"
}

func (f fakeDataHandler) AppendReason(string) {
}

func (f fakeDataHandler) GetClosePrice() decimal.Decimal {
	return decimal.Zero
}

func (f fakeDataHandler) GetHighPrice() decimal.Decimal {
	return decimal.Zero
}

func (f fakeDataHandler) GetLowPrice() decimal.Decimal {
	return decimal.Zero
}

func (f fakeDataHandler) GetOpenPrice() decimal.Decimal {
	return decimal.Zero
}

func (f fakeDataHandler) GetUnderlyingPair() currency.Pair {
	return f.Pair()
}

func (f fakeDataHandler) AppendReasonf(s string, i ...interface{}) {}

func (f fakeDataHandler) GetBase() *event.Base {
	return &event.Base{}
}

func (f fakeDataHandler) GetConcatReasons() string {
	return ""
}

func (f fakeDataHandler) GetReasons() []string {
	return nil
}
