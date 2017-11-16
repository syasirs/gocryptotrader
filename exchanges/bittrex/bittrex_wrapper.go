package bittrex

import (
	"log"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start starts the Bittrex go routine
func (b *Bittrex) Start() {
	go b.Run()
}

// Run implements the Bittrex wrapper
func (b *Bittrex) Run() {
	if b.Verbose {
		log.Printf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", b.GetName(), len(b.EnabledPairs), b.EnabledPairs)
	}

	exchangeProducts, err := b.GetMarkets()
	if err != nil {
		log.Printf("%s Failed to get available symbols.\n", b.GetName())
	} else {
		forceUpgrade := false
		if !common.DataContains(b.EnabledPairs, "-") || !common.DataContains(b.AvailablePairs, "-") {
			forceUpgrade = true
		}
		var currencies []string
		for x := range exchangeProducts {
			if !exchangeProducts[x].IsActive || exchangeProducts[x].MarketName == "" {
				continue
			}
			currencies = append(currencies, exchangeProducts[x].MarketName)
		}

		if forceUpgrade {
			enabledPairs := []string{"USDT-BTC"}
			log.Println("WARNING: Available pairs for Bittrex reset due to config upgrade, please enable the ones you would like again")

			err = b.UpdateEnabledCurrencies(enabledPairs, true)
			if err != nil {
				log.Printf("%s Failed to get config.\n", b.GetName())
			}
		}
		err = b.UpdateAvailableCurrencies(currencies, forceUpgrade)
		if err != nil {
			log.Printf("%s Failed to get config.\n", b.GetName())
		}
	}
}

// GetExchangeAccountInfo Retrieves balances for all enabled currencies for the
// Bittrex exchange
func (b *Bittrex) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = b.GetName()
	accountBalance, err := b.GetAccountBalances()
	if err != nil {
		return response, err
	}

	for i := 0; i < len(accountBalance); i++ {
		var exchangeCurrency exchange.AccountCurrencyInfo
		exchangeCurrency.CurrencyName = accountBalance[i].Currency
		exchangeCurrency.TotalValue = accountBalance[i].Balance
		exchangeCurrency.Hold = accountBalance[i].Balance - accountBalance[i].Available
		response.Currencies = append(response.Currencies, exchangeCurrency)
	}
	return response, nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bittrex) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	tick, err := b.GetMarketSummary(exchange.FormatExchangeCurrency(b.GetName(), p).String())
	if err != nil {
		return tickerPrice, err
	}
	tickerPrice.Pair = p
	tickerPrice.Ask = tick[0].Ask
	tickerPrice.Bid = tick[0].Bid
	tickerPrice.Last = tick[0].Last
	tickerPrice.Volume = tick[0].Volume
	ticker.ProcessTicker(b.GetName(), p, tickerPrice, assetType)
	return ticker.GetTicker(b.Name, p, assetType)
}

// GetTickerPrice returns the ticker for a currency pair
func (b *Bittrex) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tick, err := ticker.GetTicker(b.GetName(), p, ticker.Spot)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tick, nil
}

// GetOrderbookEx returns the orderbook for a currency pair
func (b *Bittrex) GetOrderbookEx(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := b.Orderbooks.GetOrderbook(b.GetName(), p, assetType)
	if err == nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bittrex) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := b.GetOrderbook(exchange.FormatExchangeCurrency(b.GetName(), p).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Buy {
		orderBook.Bids = append(orderBook.Bids,
			orderbook.Item{
				Amount: orderbookNew.Buy[x].Quantity,
				Price:  orderbookNew.Buy[x].Rate,
			},
		)
	}

	for x := range orderbookNew.Sell {
		orderBook.Asks = append(orderBook.Asks,
			orderbook.Item{
				Amount: orderbookNew.Sell[x].Quantity,
				Price:  orderbookNew.Sell[x].Rate,
			},
		)
	}

	b.Orderbooks.ProcessOrderbook(b.GetName(), p, orderBook, assetType)
	return b.Orderbooks.GetOrderbook(b.Name, p, assetType)
}
