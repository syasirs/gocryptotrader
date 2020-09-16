package trade

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"
	sqltrade "github.com/thrasher-corp/gocryptotrader/database/repository/trade"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

func TestAddTradesToBuffer(t *testing.T) {
	dbConf := database.Config{
		Enabled: true,
		Driver:  database.DBSQLite3,
		ConnectionDetails: drivers.ConnectionDetails{
			Host:     "localhost",
			Database: "./rpctestdb",
		},
	}
	database.DB.Config = &dbConf
	cp, _ := currency.NewPairFromString("BTC-USD")
	err := AddTradesToBuffer("test!", []Data{
		{
			Timestamp:    time.Now(),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        1337,
			Amount:       1337,
			Side:         order.Buy,
		},
	}...)
	if err != nil {
		t.Error(err)
	}
	if atomic.AddInt32(&processor.started, 0) == 0 {
		t.Error("expected the processor to have started")
	}

	err = AddTradesToBuffer("test!", []Data{
		{
			Timestamp:    time.Now(),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        0,
			Amount:       0,
			Side:         order.Buy,
		},
	}...)
	if err == nil {
		t.Error("expected error")
	}
	buffer = nil
	err = AddTradesToBuffer("test!", []Data{
		{
			Timestamp:    time.Now(),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        -1,
			Amount:       -1,
			Side:         "",
		},
	}...)
	if err != nil {
		t.Error(err)
	}
	if buffer[0].Amount != 1 {
		t.Error("expected positive amount")
	}
	if buffer[0].Side != order.UnknownSide {
		t.Error("expected unknown side")
	}
}

func TestSqlDataToTrade(t *testing.T) {
	t.Parallel()
	uuiderino, _ := uuid.NewV4()
	data, err := SQLDataToTrade(sqltrade.Data{
		ID:        uuiderino.String(),
		Timestamp: time.Time{},
		Exchange:  "hello",
		Base:      currency.BTC.String(),
		Quote:     currency.USD.String(),
		AssetType: "spot",
		Price:     1337,
		Amount:    1337,
		Side:      "buy",
	})
	if err != nil {
		t.Error(err)
	}
	if len(data) != 1 {
		t.Fatal("unexpected scenario")
	}
	if data[0].Side != order.Buy {
		t.Error("expected buy side")
	}
	if data[0].CurrencyPair.String() != "BTCUSD" {
		t.Errorf("expected \"BTCUSD\", got %v", data[0].CurrencyPair)
	}
	if data[0].AssetType != asset.Spot {
		t.Error("expected spot")
	}
}

func TestTradeToSQLData(t *testing.T) {
	t.Parallel()
	cp := currency.NewPair(currency.BTC, currency.USD)
	sqlData, err := tradeToSQLData(Data{
		Timestamp:    time.Now(),
		Exchange:     "test!",
		CurrencyPair: cp,
		AssetType:    asset.Spot,
		Price:        1337,
		Amount:       1337,
		Side:         order.Buy,
	})
	if err != nil {
		t.Error(err)
	}
	if len(sqlData) != 1 {
		t.Fatal("unexpected result")
	}
	if sqlData[0].Base != cp.Base.String() {
		t.Errorf("expected \"BTC\", got %v", sqlData[0].Base)
	}
	if sqlData[0].AssetType != asset.Spot.String() {
		t.Error("expected spot")
	}
}

func TestConvertTradesToCandles(t *testing.T) {
	t.Parallel()
	cp, _ := currency.NewPairFromString("BTC-USD")
	candles, err := ConvertTradesToCandles(kline.FifteenSecond, []Data{
		{
			Timestamp:    time.Now(),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        1337,
			Amount:       1337,
			Side:         order.Buy,
		},
		{
			Timestamp:    time.Now().Add(time.Second),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        1337,
			Amount:       1337,
			Side:         order.Buy,
		},
		{
			Timestamp:    time.Now().Add(time.Minute),
			Exchange:     "test!",
			CurrencyPair: cp,
			AssetType:    asset.Spot,
			Price:        -1337,
			Amount:       -1337,
			Side:         order.Buy,
		},
	}...)
	if err != nil {
		t.Fatal(err)
	}
	if len(candles.Candles) != 2 {
		t.Fatal("unexpected candle amount")
	}
	if candles.Interval != kline.FifteenSecond {
		t.Error("expected fifteen seconds")
	}
}

func TestShutdown(t *testing.T) {
	var p Processor
	p.mutex.Lock()
	buffer = nil
	bufferProcessorInterval = time.Second
	p.mutex.Unlock()
	go p.Run()
	time.Sleep(time.Millisecond)
	if atomic.LoadInt32(&p.started) != 1 {
		t.Error("expected it to start running")
	}
	time.Sleep(time.Second * 2)
	if atomic.LoadInt32(&p.started) != 0 {
		t.Error("expected it to stop running")
	}
}

func TestFilterTradesByTime(t *testing.T) {
	t.Parallel()
	trades := []Data{
		{
			Exchange:  "test",
			Timestamp: time.Now().Add(-time.Second),
		},
	}
	trades = FilterTradesByTime(trades, time.Now().Add(-time.Minute), time.Now())
	if len(trades) != 1 {
		t.Error("failed to filter")
	}
	trades = FilterTradesByTime(trades, time.Now().Add(-time.Millisecond), time.Now())
	if len(trades) != 0 {
		t.Error("failed to filter")
	}
}

func TestSaveTradesToDatabase(t *testing.T) {
	t.Parallel()
	err := SaveTradesToDatabase(Data{})
	if err != nil && err.Error() != "exchange name/uuid not set, cannot insert" {
		t.Error(err)
	}
}
