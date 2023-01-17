package kline

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

func TestCreateKlineRequest(t *testing.T) {
	t.Parallel()
	_, err := CreateKlineRequest("", currency.EMPTYPAIR, currency.EMPTYPAIR, 0, 0, 0, time.Time{}, time.Time{})
	if !errors.Is(err, ErrUnsetName) {
		t.Fatalf("received: '%v', but expected '%v'", err, ErrUnsetName)
	}

	_, err = CreateKlineRequest("name", currency.EMPTYPAIR, currency.EMPTYPAIR, 0, 0, 0, time.Time{}, time.Time{})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Fatalf("received: '%v', but expected '%v'", err, currency.ErrCurrencyPairEmpty)
	}

	pair := currency.NewPair(currency.BTC, currency.USDT)
	_, err = CreateKlineRequest("name", pair, currency.EMPTYPAIR, 0, 0, 0, time.Time{}, time.Time{})
	if !errors.Is(err, currency.ErrCurrencyPairEmpty) {
		t.Fatalf("received: '%v', but expected '%v'", err, currency.ErrCurrencyPairEmpty)
	}

	pair2 := pair.Upper()
	_, err = CreateKlineRequest("name", pair, pair2, 0, 0, 0, time.Time{}, time.Time{})
	if !errors.Is(err, asset.ErrNotSupported) {
		t.Fatalf("received: '%v', but expected '%v'", err, asset.ErrNotSupported)
	}

	_, err = CreateKlineRequest("name", pair, pair2, asset.Spot, 0, 0, time.Time{}, time.Time{})
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("received: '%v', but expected '%v'", err, ErrInvalidInterval)
	}

	_, err = CreateKlineRequest("name", pair, pair2, asset.Spot, OneHour, 0, time.Time{}, time.Time{})
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("received: '%v', but expected '%v'", err, ErrInvalidInterval)
	}

	_, err = CreateKlineRequest("name", pair, pair2, asset.Spot, OneHour, OneMin, time.Time{}, time.Time{})
	if !errors.Is(err, common.ErrDateUnset) {
		t.Fatalf("received: '%v', but expected '%v'", err, common.ErrDateUnset)
	}

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = CreateKlineRequest("name", pair, pair2, asset.Spot, OneHour, OneMin, start, time.Time{})
	if !errors.Is(err, common.ErrDateUnset) {
		t.Fatalf("received: '%v', but expected '%v'", err, common.ErrDateUnset)
	}

	end := start.AddDate(0, 0, 1)
	r, err := CreateKlineRequest("name", pair, pair2, asset.Spot, OneHour, OneMin, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if r.Exchange != "name" {
		t.Fatalf("received: '%v' but expected: '%v'", r.Exchange, "name")
	}

	if !r.Pair.Equal(pair) {
		t.Fatalf("received: '%v' but expected: '%v'", r.Pair, pair)
	}

	if r.Asset != asset.Spot {
		t.Fatalf("received: '%v' but expected: '%v'", r.Asset, asset.Spot)
	}

	if r.ExchangeInterval != OneMin {
		t.Fatalf("received: '%v' but expected: '%v'", r.ExchangeInterval, OneMin)
	}

	if r.ClientRequired != OneHour {
		t.Fatalf("received: '%v' but expected: '%v'", r.ClientRequired, OneHour)
	}

	if r.Start != start {
		t.Fatalf("received: '%v' but expected: '%v'", r.Start, start)
	}

	if r.End != end {
		t.Fatalf("received: '%v' but expected: '%v'", r.End, end)
	}

	if r.RequestFormatted.String() != "BTCUSDT" {
		t.Fatalf("received: '%v' but expected: '%v'", r.RequestFormatted.String(), "BTCUSDT")
	}

	// Check end date/time shift if the request time is mid candle and not
	// aligned correctly.
	end = end.Round(0)
	end = end.Add(time.Second * 30)
	r, err = CreateKlineRequest("name", pair, pair2, asset.Spot, OneHour, OneMin, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if !r.End.Equal(end.Add(OneHour.Duration() - (time.Second * 30))) {
		t.Fatalf("received: '%v', but expected '%v'", r.End, end.Add(OneHour.Duration()-(time.Second*30)))
	}
}

func TestGetRanges(t *testing.T) {
	t.Parallel()

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	pair := currency.NewPair(currency.BTC, currency.USDT)

	var r *Request
	_, err := r.GetRanges(100)
	if !errors.Is(err, errNilRequest) {
		t.Fatalf("received: '%v', but expected '%v'", err, errNilRequest)
	}

	r, err = CreateKlineRequest("name", pair, pair, asset.Spot, OneHour, OneMin, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	holder, err := r.GetRanges(100)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if len(holder.Ranges) != 15 {
		t.Fatalf("received: '%v', but expected '%v'", len(holder.Ranges), 15)
	}
}

var protecThyCandles sync.Mutex

func getOneMinute() []Candle {
	protecThyCandles.Lock()
	candles := make([]Candle, len(oneMinuteCandles))
	copy(candles, oneMinuteCandles)
	protecThyCandles.Unlock()
	return candles
}

var oneMinuteCandles = func() []Candle {
	var candles []Candle
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for x := 0; x < 1442; x++ { // two extra candles.
		candles = append(candles, Candle{
			Time:   start,
			Volume: 1,
			Open:   1,
			High:   float64(1 + x),
			Low:    float64(-(1 + x)),
			Close:  1,
		})
		start = start.Add(time.Minute)
	}
	return candles
}()

func getOneHour() []Candle {
	protecThyCandles.Lock()
	candles := make([]Candle, len(oneHourCandles))
	copy(candles, oneHourCandles)
	protecThyCandles.Unlock()
	return candles
}

var oneHourCandles = func() []Candle {
	var candles []Candle
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for x := 0; x < 24; x++ {
		candles = append(candles, Candle{
			Time:   start,
			Volume: 1,
			Open:   1,
			High:   float64(1 + x),
			Low:    float64(-(1 + x)),
			Close:  1,
		})
		start = start.Add(time.Hour)
	}
	return candles
}()

func TestRequest_ProcessResponse(t *testing.T) {
	t.Parallel()

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	pair := currency.NewPair(currency.BTC, currency.USDT)

	var r *Request
	_, err := r.ProcessResponse(nil)
	if !errors.Is(err, errNilRequest) {
		t.Fatalf("received: '%v', but expected '%v'", err, errNilRequest)
	}

	r = &Request{}
	_, err = r.ProcessResponse(nil)
	if !errors.Is(err, errNoTimeSeriesDataToConvert) {
		t.Fatalf("received: '%v', but expected '%v'", err, errNoTimeSeriesDataToConvert)
	}

	// no conversion
	r, err = CreateKlineRequest("name", pair, pair, asset.Spot, OneHour, OneHour, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	holder, err := r.ProcessResponse(getOneHour())
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if len(holder.Candles) != 24 {
		t.Fatalf("received: '%v', but expected '%v'", len(holder.Candles), 24)
	}

	// with conversion
	r, err = CreateKlineRequest("name", pair, pair, asset.Spot, OneHour, OneMin, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	holder, err = r.ProcessResponse(getOneMinute())
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if len(holder.Candles) != 24 {
		t.Fatalf("received: '%v', but expected '%v'", len(holder.Candles), 24)
	}

	// Potential partial candle
	end = time.Now().UTC()
	fmt.Println("END:", end)
	start = end.AddDate(0, 0, -5).Truncate(time.Duration(OneDay))
	r, err = CreateKlineRequest("name", pair, pair, asset.Spot, OneDay, OneDay, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if !r.PartialCandle {
		t.Fatalf("received: '%v', but expected '%v'", r.PartialCandle, true)
	}

	hasIncomplete := []Candle{
		{Time: start, Close: 1},
		{Time: start.Add(OneDay.Duration()), Close: 2},
		{Time: start.Add(OneDay.Duration() * 2), Close: 3},
		{Time: start.Add(OneDay.Duration() * 3), Close: 4},
		{Time: start.Add(OneDay.Duration() * 4), Close: 5},
		{Time: start.Add(OneDay.Duration() * 5), Close: 5.5},
	}

	fmt.Println("start", start)

	fmt.Printf("hasIncomplete: %+v\n", hasIncomplete)

	sweetItem, err := r.ProcessResponse(hasIncomplete)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if sweetItem.Candles[len(sweetItem.Candles)-1].ValidationIssues != PartialCandle {
		t.Fatalf("received: '%v', but expected '%v'", "no issues", PartialCandle)
	}

	missingIncomplete := []Candle{
		{Time: start, Close: 1},
		{Time: start.Add(OneDay.Duration()), Close: 2},
		{Time: start.Add(OneDay.Duration() * 2), Close: 3},
		{Time: start.Add(OneDay.Duration() * 3), Close: 4},
		{Time: start.Add(OneDay.Duration() * 4), Close: 5},
	}

	sweetItem, err = r.ProcessResponse(missingIncomplete)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if sweetItem.Candles[len(sweetItem.Candles)-1].ValidationIssues == PartialCandle {
		t.Fatalf("received: '%v', but expected '%v'", sweetItem.Candles[len(sweetItem.Candles)-1].ValidationIssues, "no issues")
	}
}

func TestExtendedRequest_ProcessResponse(t *testing.T) {
	t.Parallel()

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	pair := currency.NewPair(currency.BTC, currency.USDT)

	var rExt *ExtendedRequest
	_, err := rExt.ProcessResponse(nil)
	if !errors.Is(err, errNilRequest) {
		t.Fatalf("received: '%v', but expected '%v'", err, errNilRequest)
	}

	rExt = &ExtendedRequest{}
	_, err = rExt.ProcessResponse(nil)
	if !errors.Is(err, errNoTimeSeriesDataToConvert) {
		t.Fatalf("received: '%v', but expected '%v'", err, errNoTimeSeriesDataToConvert)
	}

	// no conversion
	r, err := CreateKlineRequest("name", pair, pair, asset.Spot, OneHour, OneHour, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	dates, err := r.GetRanges(100)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	rExt = &ExtendedRequest{r, dates}

	holder, err := rExt.ProcessResponse(getOneHour())
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if len(holder.Candles) != 24 {
		t.Fatalf("received: '%v', but expected '%v'", len(holder.Candles), 24)
	}

	// with conversion
	r, err = CreateKlineRequest("name", pair, pair, asset.Spot, OneHour, OneMin, start, end)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	dates, err = r.GetRanges(100)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	rExt = &ExtendedRequest{r, dates}

	holder, err = rExt.ProcessResponse(getOneMinute())
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v', but expected '%v'", err, nil)
	}

	if len(holder.Candles) != 24 {
		t.Fatalf("received: '%v', but expected '%v'", len(holder.Candles), 24)
	}
}

func TestExtendedRequest_Size(t *testing.T) {
	t.Parallel()

	var rExt *ExtendedRequest
	if rExt.Size() != 0 {
		t.Fatalf("received: '%v', but expected '%v'", rExt.Size(), 0)
	}

	rExt = &ExtendedRequest{IntervalRangeHolder: &IntervalRangeHolder{Limit: 100, Ranges: []IntervalRange{{}, {}}}}
	if rExt.Size() != 200 {
		t.Fatalf("received: '%v', but expected '%v'", rExt.Size(), 200)
	}
}
