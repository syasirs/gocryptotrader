package kraken

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start starts the Kraken go routine
func (k *Kraken) Start() {
	go k.Run()
}

// Run implements the Kraken wrapper
func (k *Kraken) Run() {
	if k.Verbose {
		log.Printf("%s polling delay: %ds.\n", k.GetName(), k.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", k.GetName(), len(k.EnabledPairs), k.EnabledPairs)
	}

	assetPairs, err := k.GetAssetPairs()
	if err != nil {
		log.Printf("%s Failed to get available symbols.\n", k.GetName())
	} else {
		var exchangeProducts []string
		for _, v := range assetPairs {
			exchangeProducts = append(exchangeProducts, v.Altname)
		}
		err = k.UpdateAvailableCurrencies(exchangeProducts, false)
		if err != nil {
			log.Printf("%s Failed to get config.\n", k.GetName())
		}
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (k *Kraken) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	pairs := k.GetEnabledCurrencies()
	pairsCollated, err := exchange.GetAndFormatExchangeCurrencies(k.Name, pairs)
	if err != nil {
		return tickerPrice, err
	}
	err = k.SetTicker(pairsCollated.String())
	if err != nil {
		return tickerPrice, err
	}

	for _, x := range pairs {
		var tp ticker.Price
		tick, ok := k.Ticker[x.Pair().String()]
		if !ok {
			continue
		}

		tp.Pair = x
		tp.Last = tick.Last
		tp.Ask = tick.Ask
		tp.Bid = tick.Bid
		tp.High = tick.High
		tp.Low = tick.Low
		tp.Volume = tick.Volume
		ticker.ProcessTicker(k.GetName(), x, tp, assetType)
	}
	return ticker.GetTicker(k.GetName(), p, assetType)
}

// SetTicker sets ticker information from kraken
func (k *Kraken) SetTicker(symbol string) error {
	values := url.Values{}
	values.Set("pair", symbol)

	type Response struct {
		Error []interface{}             `json:"error"`
		Data  map[string]TickerResponse `json:"result"`
	}

	resp := Response{}
	path := fmt.Sprintf("%s/%s/public/%s?%s", krakenAPIURL, krakenAPIVersion, krakenTicker, values.Encode())

	err := common.SendHTTPGetRequest(path, true, k.Verbose, &resp)
	if err != nil {
		return err
	}

	if len(resp.Error) > 0 {
		return fmt.Errorf("Kraken error: %s", resp.Error)
	}

	for x, y := range resp.Data {
		ticker := Ticker{}
		ticker.Ask, _ = strconv.ParseFloat(y.Ask[0], 64)
		ticker.Bid, _ = strconv.ParseFloat(y.Bid[0], 64)
		ticker.Last, _ = strconv.ParseFloat(y.Last[0], 64)
		ticker.Volume, _ = strconv.ParseFloat(y.Volume[1], 64)
		ticker.VWAP, _ = strconv.ParseFloat(y.VWAP[1], 64)
		ticker.Trades = y.Trades[1]
		ticker.Low, _ = strconv.ParseFloat(y.Low[1], 64)
		ticker.High, _ = strconv.ParseFloat(y.High[1], 64)
		ticker.Open, _ = strconv.ParseFloat(y.Open, 64)
		k.Ticker[x] = ticker
	}
	return nil
}

// GetTickerPrice returns the ticker for a currency pair
func (k *Kraken) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(k.GetName(), p, assetType)
	if err != nil {
		return k.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (k *Kraken) GetOrderbookEx(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(k.GetName(), p, assetType)
	if err == nil {
		return k.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (k *Kraken) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := k.GetDepth(exchange.FormatExchangeCurrency(k.GetName(), p).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: orderbookNew.Bids[x].Amount, Price: orderbookNew.Bids[x].Price})
	}

	for x := range orderbookNew.Asks {
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: orderbookNew.Asks[x].Amount, Price: orderbookNew.Asks[x].Price})
	}

	orderbook.ProcessOrderbook(k.GetName(), p, orderBook, assetType)
	return orderbook.GetOrderbook(k.Name, p, assetType)
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
// Kraken exchange - to-do
func (k *Kraken) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = k.GetName()
	return response, nil
}
