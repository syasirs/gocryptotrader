package candle

import (
	"errors"
	"time"
)

var (
	errInvalidInput = errors.New("exchange, base , quote, interval, start & end cannot be empty")
)

// Item generic candle holder for modelPSQL & modelSQLite
type Item struct {
	ID         string
	ExchangeID string
	Base       string
	Quote      string
	Interval   string
	Asset      string
	Candles    []Candle
}

// Candle holds each interval
type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}
