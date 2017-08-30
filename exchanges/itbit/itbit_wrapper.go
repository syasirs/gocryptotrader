package itbit

import (
	"log"
	"strconv"

	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start starts the ItBit go routine
func (i *ItBit) Start() {
	go i.Run()
}

// Run implements the ItBit wrapper
func (i *ItBit) Run() {
	if i.Verbose {
		log.Printf("%s polling delay: %ds.\n", i.GetName(), i.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", i.GetName(), len(i.EnabledPairs), i.EnabledPairs)
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (i *ItBit) UpdateTicker(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	var tickerPrice ticker.TickerPrice
	tick, err := i.GetTicker(p.Pair().String())
	if err != nil {
		return tickerPrice, err
	}

	tickerPrice.Pair = p
	tickerPrice.Ask = tick.Ask
	tickerPrice.Bid = tick.Bid
	tickerPrice.Last = tick.LastPrice
	tickerPrice.High = tick.High24h
	tickerPrice.Low = tick.Low24h
	tickerPrice.Volume = tick.Volume24h
	ticker.ProcessTicker(i.GetName(), p, tickerPrice)
	return tickerPrice, nil
}

// GetTickerPrice returns the ticker for a currency pair
func (i *ItBit) GetTickerPrice(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	tickerNew, err := ticker.GetTicker(i.GetName(), p)
	if err != nil {
		return i.UpdateTicker(p)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (i *ItBit) GetOrderbookEx(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(i.GetName(), p)
	if err == nil {
		return i.UpdateOrderbook(p)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (i *ItBit) UpdateOrderbook(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	var orderBook orderbook.OrderbookBase
	orderbookNew, err := i.GetOrderbook(p.Pair().String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			log.Println(err)
		}
		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			log.Println(err)
		}
		orderBook.Bids = append(orderBook.Bids, orderbook.OrderbookItem{Amount: amount, Price: price})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			log.Println(err)
		}
		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			log.Println(err)
		}
		orderBook.Asks = append(orderBook.Asks, orderbook.OrderbookItem{Amount: amount, Price: price})
	}

	orderBook.Pair = p
	orderbook.ProcessOrderbook(i.GetName(), p, orderBook)
	return orderBook, nil
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
//ItBit exchange - to-do
func (i *ItBit) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = i.GetName()
	return response, nil
}
