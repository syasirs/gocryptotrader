package kline

import (
	"errors"
)

var errNoDataData = errors.New("no candle data")

// GetAveragePrice returns the average price from the open, high, low and close
func (c *Candle) GetAveragePrice() float64 {
	return (c.Open + c.High + c.Low + c.Close) / 4
}

// GetTWAP returns the time weighted average price for the specified period.
// NOTE: This assumes the most recent price is at the tail end of the slice.
// Based off: https://blog.quantinsti.com/twap/
// Only returns one item as all other items are just the average price.
func (k *Item) GetTWAP() (float64, error) {
	if len(k.Candles) == 0 {
		return 0, errNoDataData
	}
	var cumAveragePrice float64
	for x := range k.Candles {
		cumAveragePrice += k.Candles[x].GetAveragePrice()
	}
	return cumAveragePrice / float64(len(k.Candles)), nil
}

// GetTypicalPrice returns the typical average price from the high, low and
// close values.
func (c *Candle) GetTypicalPrice() float64 {
	return (c.High + c.Low + c.Close) / 3
}

// GetVWAPs returns the Volume Weighted Averages prices which are the cumulative
// average price with respect to the volume.
// NOTE: This assumes the most recent price is at the tail end of the slice.
// Based off: https://blog.quantinsti.com/vwap-strategy/
func (k *Item) GetVWAPs() ([]float64, error) {
	if len(k.Candles) == 0 {
		return nil, errNoDataData
	}
	store := make([]float64, len(k.Candles))
	var cumTotal, cumVolume float64
	for x := range k.Candles {
		cumTotal += k.Candles[x].GetTypicalPrice() * k.Candles[x].Volume
		cumVolume += k.Candles[x].Volume
		store[x] = cumTotal / cumVolume
	}
	return store, nil
}
