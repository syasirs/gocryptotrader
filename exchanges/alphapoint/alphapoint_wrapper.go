package alphapoint

import (
	"log"

	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

//GetExchangeAccountInfo : Retrieves balances for all enabled currencies for the Alphapoint exchange
func (e *Alphapoint) GetExchangeAccountInfo() (exchange.ExchangeAccountInfo, error) {
	var response exchange.ExchangeAccountInfo
	response.ExchangeName = e.GetName()
	account, err := e.GetAccountInfo()
	if err != nil {
		return response, err
	}
	for i := 0; i < len(account.Currencies); i++ {
		var exchangeCurrency exchange.ExchangeAccountCurrencyInfo
		exchangeCurrency.CurrencyName = account.Currencies[i].Name
		exchangeCurrency.TotalValue = float64(account.Currencies[i].Balance)
		exchangeCurrency.Hold = float64(account.Currencies[i].Hold)

		response.Currencies = append(response.Currencies, exchangeCurrency)
	}
	//If it all works out
	return response, nil
}

func (a *Alphapoint) GetTickerPrice(currency string) ticker.TickerPrice {
	var tickerPrice ticker.TickerPrice
	tick, err := a.GetTicker(currency)
	if err != nil {
		log.Println(err)
		return ticker.TickerPrice{}
	}
	tickerPrice.Ask = tick.Ask
	tickerPrice.Bid = tick.Bid

	return tickerPrice
}

func (a *Alphapoint) GetOrderbookEx(currency string) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(a.GetName(), currency[0:3], currency[3:])
	if err == nil {
		return ob, nil
	}

	var orderBook orderbook.OrderbookBase
	orderbookNew, err := a.GetOrderbook(currency)
	if err != nil {
		return orderBook, err
	}

	for x, _ := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.OrderbookItem{Amount: data.Quantity, Price: data.Price})
	}

	for x, _ := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.OrderbookItem{Amount: data.Quantity, Price: data.Price})
	}

	orderBook.FirstCurrency = currency[0:3]
	orderBook.SecondCurrency = currency[3:]
	orderbook.ProcessOrderbook(a.GetName(), orderBook.FirstCurrency, orderBook.SecondCurrency, orderBook)
	return orderBook, nil
}
