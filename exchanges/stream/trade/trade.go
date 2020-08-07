package trade

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/log"
)


// Setup creates the trade processor if trading is supported
func (t *Traderino) Setup(name string) {
	t.Name = name
	go t.Processor()
}

// Shutdown kills the lingering processor
func (t *Traderino) Shutdown() {
	close(t.shutdown)
}

// Process will push trade data onto the buffer
func (t *Traderino) Process(data ...Data) {
	t.mutex.Lock()
	for i := range data {
		buffer = append(buffer, data[i])
	}
	t.mutex.Unlock()
}

// Processor will convert buffered trade data into candles
// then stores the candles and clears the buffer to allow
// more allocations
func (t *Traderino) Processor() {
	if atomic.AddInt32(&t.started, 1) != 1 {
		log.Error(log.WebsocketMgr, "websocket trade processor already started")
	}
	defer func() {
		atomic.CompareAndSwapInt32(&t.started, 1, 0)
	}()
	log.Debugln(log.OrderBook, "Order manager starting...")
	timer := time.NewTicker(time.Minute)
	for {
		select {
		case <-t.shutdown:
			return
		case <-timer.C:
			log.Debug(log.WebsocketMgr, "TRADE PROCESSOR STARTING")
			t.mutex.Lock()
			sort.Sort(ByDate(buffer))
			groupedData := convertTradeDatasToCandles(kline.FifteenSecond, buffer...)
			var candles []CandleHolder
			for k, v := range groupedData {
				candles = append(candles, classifyOHLCV(time.Unix(k, 0), v...))
			}
			for i := range candles {
				for j := range t.previousCandles {
					if candles[i].candle.Time.Equal(t.previousCandles[j].candle.Time) {
						t.previousCandles[j].amendCandle(candles[i].trades...)
						candles[i].candle = t.previousCandles[j].candle
					}
				}
			}
			// now save previous candles
			err := t.SaveCandlesToButt(t.previousCandles)
			if err != nil {
				log.Error(log.WebsocketMgr,"Processor SaveCandlesToButt ", err)
				t.mutex.Unlock()
				continue
			}

			// now set current candles to previous for the next run
			t.previousCandles = candles
			buffer = nil
			t.mutex.Unlock()
		}
	}
}

func (t *Traderino) SaveCandlesToButt(candles []CandleHolder) error {

}

func convertTradeDatasToCandles(interval kline.Interval, times ...Data) map[int64][]Data {
	groupedData := make(map[int64][]Data)
	for i:= range times {
		nearestInterval := getNearestInterval(times[i].Timestamp, interval)
		groupedData[nearestInterval] = append(
			groupedData[nearestInterval],
			times[i],
		)
	}
	return groupedData
}

func getNearestInterval(t time.Time, interval kline.Interval) int64 {
	return t.Truncate(interval.Duration()).Unix()
}

func (c *CandleHolder) amendCandle(datas ...Data) {
	sort.Sort(ByDate(datas))
	c.trades = append(c.trades, datas...)
	sort.Sort(ByDate(c.trades))
	c.candle.Open = c.trades[0].Price
	c.candle.Close = c.trades[len(c.trades)-1].Price
	for i := range datas {
		c.candle.Volume += datas[i].Amount
	}
	for i := range c.trades {
		// some exchanges will send it as negative for sells
		// do they though?
		if c.trades[i].Price < 0 {
			log.Debug(log.WebsocketMgr, "NEGATIVE TRADE")
			c.trades[i].Price = c.trades[i].Price * -1
		}
		if c.trades[i].Amount < 0 {
			log.Debug(log.WebsocketMgr, "NEGATIVE TRADE")
			c.trades[i].Amount = c.trades[i].Amount * -1
		}
		if c.trades[i].Price < c.candle.Low || c.candle.Low == 0 {
			c.candle.Low = c.trades[i].Price
		}
		if c.trades[i].Price > c.candle.High {
			c.candle.High = c.trades[i].Price
		}
	}
}

func classifyOHLCV (t time.Time, datas ...Data) (c CandleHolder) {
	sort.Sort(ByDate(datas))
	c.candle.Open = datas[0].Price
	c.candle.Close = datas[len(datas)-1].Price
	c.trades = datas
	for i := range datas {
		// some exchanges will send it as negative for sells
		// do they though?
		if datas[i].Price < 0 {
			log.Debug(log.WebsocketMgr, "NEGATIVE TRADE")
			datas[i].Price = datas[i].Price * -1
		}
		if datas[i].Amount < 0 {
			log.Debug(log.WebsocketMgr, "NEGATIVE TRADE")
			datas[i].Amount = datas[i].Amount * -1
		}
		if datas[i].Price < c.candle.Low || c.candle.Low == 0 {
			c.candle.Low = datas[i].Price
		}
		if datas[i].Price > c.candle.High {
			c.candle.High = datas[i].Price
		}
		c.candle.Volume += datas[i].Amount
	}
	c.candle.Time = t
	return
}