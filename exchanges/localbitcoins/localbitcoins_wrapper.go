package localbitcoins

import (
	"log"

	"github.com/thrasher-/gocryptotrader/common"

	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start starts the LocalBitcoins go routine
func (l *LocalBitcoins) Start() {
	go l.Run()
}

// Run implements the LocalBitcoins wrapper
func (l *LocalBitcoins) Run() {
	if l.Verbose {
		log.Printf("%s polling delay: %ds.\n", l.GetName(), l.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", l.GetName(), len(l.EnabledPairs), l.EnabledPairs)
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (l *LocalBitcoins) UpdateTicker(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	var tickerPrice ticker.TickerPrice
	tick, err := l.GetTicker()
	if err != nil {
		return tickerPrice, err
	}

	for key, value := range tick {
		currency := pair.NewCurrencyPair("BTC", common.StringToUpper(key))
		var tp ticker.TickerPrice
		tp.Pair = currency
		tp.Last = value.Rates.Last
		tp.Volume = value.VolumeBTC
		ticker.ProcessTicker(l.GetName(), currency, tp)
	}

	return ticker.GetTicker(l.GetName(), p)
}

// GetTickerPrice returns the ticker for a currency pair
func (l *LocalBitcoins) GetTickerPrice(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	tickerNew, err := ticker.GetTicker(l.GetName(), p)
	if err == nil {
		return l.UpdateTicker(p)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (l *LocalBitcoins) GetOrderbookEx(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(l.GetName(), p)
	if err == nil {
		return l.UpdateOrderbook(p)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (l *LocalBitcoins) UpdateOrderbook(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	var orderBook orderbook.OrderbookBase
	orderbookNew, err := l.GetOrderbook(p.GetSecondCurrency().String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.OrderbookItem{Amount: data.Amount, Price: data.Price})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Bids = append(orderBook.Asks, orderbook.OrderbookItem{Amount: data.Amount, Price: data.Price})
	}

	orderBook.Pair = p
	orderbook.ProcessOrderbook(l.GetName(), p, orderBook)
	return orderBook, nil
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
// LocalBitcoins exchange
func (l *LocalBitcoins) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = l.GetName()
	accountBalance, err := l.GetWalletBalance()
	if err != nil {
		return response, err
	}
	var exchangeCurrency exchange.AccountCurrencyInfo
	exchangeCurrency.CurrencyName = "BTC"
	exchangeCurrency.TotalValue = accountBalance.Total.Balance

	response.Currencies = append(response.Currencies, exchangeCurrency)
	return response, nil
}
