package liqui

import (
	"errors"
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/stats"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

func (l *Liqui) Start() {
	go l.Run()
}

func (l *Liqui) Run() {
	if l.Verbose {
		log.Printf("%s polling delay: %ds.\n", l.GetName(), l.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", l.GetName(), len(l.EnabledPairs), l.EnabledPairs)
	}

	var err error
	l.Info, err = l.GetInfo()
	if err != nil {
		log.Printf("%s Unable to fetch info.\n", l.GetName())
	} else {
		exchangeProducts := l.GetAvailablePairs(true)
		err = l.UpdateAvailableCurrencies(exchangeProducts)
		if err != nil {
			log.Printf("%s Failed to get config.\n", l.GetName())
		}
	}

	pairs := []string{}
	for _, x := range l.EnabledPairs {
		currencies := common.SplitStrings(x, "_")
		x = common.StringToLower(currencies[0]) + "_" + common.StringToLower(currencies[1])
		pairs = append(pairs, x)
	}
	pairsString := common.JoinStrings(pairs, "-")

	for l.Enabled {
		go func() {
			ticker, err := l.GetTicker(pairsString)
			if err != nil {
				log.Println(err)
				return
			}
			for x, y := range ticker {
				currency := pair.NewCurrencyPairDelimiter(common.StringToUpper(x), "_")
				log.Printf("Liqui %s: Last %f High %f Low %f Volume %f\n", currency.Pair().String(), y.Last, y.High, y.Low, y.Vol_cur)
				l.Ticker[x] = y
				stats.AddExchangeInfo(l.GetName(), currency.GetFirstCurrency().String(), currency.GetSecondCurrency().String(), y.Last, y.Vol_cur)
			}
		}()
		time.Sleep(time.Second * l.RESTPollingDelay)
	}
}

func (l *Liqui) GetTickerPrice(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	var tickerPrice ticker.TickerPrice
	tick, ok := l.Ticker[p.Pair().Lower().String()]
	if !ok {
		return tickerPrice, errors.New("Unable to get currency.")
	}
	tickerPrice.Pair = p
	tickerPrice.Ask = tick.Buy
	tickerPrice.Bid = tick.Sell
	tickerPrice.Low = tick.Low
	tickerPrice.Last = tick.Last
	tickerPrice.Volume = tick.Vol_cur
	tickerPrice.High = tick.High
	ticker.ProcessTicker(l.GetName(), p, tickerPrice)
	return tickerPrice, nil
}

func (l *Liqui) GetOrderbookEx(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(l.GetName(), p)
	if err == nil {
		return ob, nil
	}

	var orderBook orderbook.OrderbookBase
	orderbookNew, err := l.GetDepth(p.Pair().Lower().String())
	if err != nil {
		return orderBook, err
	}

	for x, _ := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.OrderbookItem{Amount: data[1], Price: data[0]})
	}

	for x, _ := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.OrderbookItem{Amount: data[1], Price: data[0]})
	}
	orderBook.Pair = p
	orderbook.ProcessOrderbook(l.GetName(), p, orderBook)
	return orderBook, nil
}

//GetExchangeAccountInfo : Retrieves balances for all enabled currencies for the Liqui exchange
func (e *Liqui) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = e.GetName()
	accountBalance, err := e.GetAccountInfo()
	if err != nil {
		return response, err
	}

	for x, y := range accountBalance.Funds {
		var exchangeCurrency exchange.AccountCurrencyInfo
		exchangeCurrency.CurrencyName = common.StringToUpper(x)
		exchangeCurrency.TotalValue = y
		exchangeCurrency.Hold = 0
		response.Currencies = append(response.Currencies, exchangeCurrency)
	}

	return response, nil
}
