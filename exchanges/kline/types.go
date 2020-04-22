package kline

import (
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// Consts here define basic time intervals
const (
	OneMin     = Interval(time.Minute)
	ThreeMin   = Interval(3 * time.Minute)
	FiveMin    = Interval(5 * time.Minute)
	FifteenMin = Interval(15 * time.Minute)
	ThirtyMin  = Interval(30 * time.Minute)
	OneHour    = Interval(1 * time.Hour)
	TwoHour    = Interval(2 * time.Hour)
	FourHour   = Interval(4 * time.Hour)
	SixHour    = Interval(6 * time.Hour)
	TwelveHour = Interval(12 * time.Hour)
	OneDay     = Interval(24 * time.Hour)
	ThreeDay   = Interval(72 * time.Hour)
	OneWeek    = Interval(168 * time.Hour)
)

// const (
// 	OneMinStr     string = "onemin"
// 	ThreeMinStr   string = "threemin"
// 	FiveMinStr    string = "fivemin"
// 	FifteenMinStr string = "fifteenmin"
// 	ThirtyMinStr  string = "thirtymin"
// 	OneHourStr    string = "onehour"
// 	TwoHourStr    string = "twohour"
// 	FourHourStr   string = "fourhour"
// 	SixhourStr    string = "sixhour"
// 	TwelveHourStr string = "twelvehour"
// 	OneDayStr     string = "oneday"
// 	ThreeDayStr   string = "threeday"
// 	OneWeekStr    string = "oneweeks"
// )

const ErrUnsupportedInterval = "%s interval unsupported by exchange"

// Item holds all the relevant information for internal kline elements
type Item struct {
	Exchange string
	Pair     currency.Pair
	Asset    asset.Item
	Interval Interval
	Candles  []Candle
}

// Candle holds historic rate information.
type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// ExchangeCapabilities all kline related exchane supported options
type ExchangeCapabilities struct {
	SupportsIntervals bool
	Intervals         map[string]bool `json:"intervals,omitempty"`
	SupportsDateRange bool
	Limit             uint32
}

type Interval time.Duration

func (k Interval) String() string {
	return k.Duration().String()
}

func (k Interval) Word() string {
	return DurationToWord(k)
}

func (k Interval) Duration() time.Duration {
	return time.Duration(k)
}

func (k Interval) Short() string {
	s := k.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}