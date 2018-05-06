package wex

import (
	"errors"
	"log"
	"sync"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start starts the WEX go routine
func (w *WEX) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		w.Run()
		wg.Done()
	}()
}

// Run implements the WEX wrapper
func (w *WEX) Run() {
	if w.Verbose {
		log.Printf("%s Websocket: %s.", w.GetName(), common.IsEnabled(w.Websocket))
		log.Printf("%s polling delay: %ds.\n", w.GetName(), w.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", w.GetName(), len(w.EnabledPairs), w.EnabledPairs)
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (w *WEX) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	pairsCollated, err := exchange.GetAndFormatExchangeCurrencies(w.Name, w.GetEnabledCurrencies())
	if err != nil {
		return tickerPrice, err
	}

	result, err := w.GetTicker(pairsCollated.String())
	if err != nil {
		return tickerPrice, err
	}

	for _, x := range w.GetEnabledCurrencies() {
		currency := exchange.FormatExchangeCurrency(w.Name, x).Lower().String()
		var tp ticker.Price
		tp.Pair = x
		tp.Last = result[currency].Last
		tp.Ask = result[currency].Sell
		tp.Bid = result[currency].Buy
		tp.Last = result[currency].Last
		tp.Low = result[currency].Low
		tp.Volume = result[currency].VolumeCurrent
		ticker.ProcessTicker(w.Name, x, tp, assetType)
	}
	return ticker.GetTicker(w.Name, p, assetType)
}

// GetTickerPrice returns the ticker for a currency pair
func (w *WEX) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tick, err := ticker.GetTicker(w.GetName(), p, assetType)
	if err != nil {
		return w.UpdateTicker(p, assetType)
	}
	return tick, nil
}

// GetOrderbookEx returns the orderbook for a currency pair
func (w *WEX) GetOrderbookEx(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(w.GetName(), p, assetType)
	if err != nil {
		return w.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (w *WEX) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := w.GetDepth(exchange.FormatExchangeCurrency(w.Name, p).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Price: data[0], Amount: data[1]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Price: data[0], Amount: data[1]})
	}

	orderbook.ProcessOrderbook(w.GetName(), p, orderBook, assetType)
	return orderbook.GetOrderbook(w.Name, p, assetType)
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
// WEX exchange
func (w *WEX) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = w.GetName()
	accountBalance, err := w.GetAccountInfo()
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

// GetExchangeHistory returns historic trade data since exchange opening.
func (w *WEX) GetExchangeHistory(p pair.CurrencyPair, assetType string) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, errors.New("trade history not yet implemented")
}

// SubmitExchangeOrder submits a new order
func (w *WEX) SubmitExchangeOrder(p pair.CurrencyPair, side string, orderType int, amount, price float64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// ModifyExchangeOrder will allow of changing orderbook placement and limit to
// market conversion
func (w *WEX) ModifyExchangeOrder(p pair.CurrencyPair, orderID, action int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelExchangeOrder cancels an order by its corresponding ID number
func (w *WEX) CancelExchangeOrder(p pair.CurrencyPair, orderID int64) (int64, error) {
	return 0, errors.New("not yet implemented")
}

// CancelAllExchangeOrders cancels all orders associated with a currency pair
func (w *WEX) CancelAllExchangeOrders(p pair.CurrencyPair) error {
	return errors.New("not yet implemented")
}

// GetExchangeOrderInfo returns information on a current open order
func (w *WEX) GetExchangeOrderInfo(orderID int64) (float64, error) {
	return 0, errors.New("not yet implemented")
}

// GetExchangeDepositAddress returns a deposit address for a specified currency
func (w *WEX) GetExchangeDepositAddress(p pair.CurrencyPair) (string, error) {
	return "", errors.New("not yet implemented")
}

// WithdrawExchangeFunds returns a withdrawal ID when a withdrawal is submitted
func (w *WEX) WithdrawExchangeFunds(address string, p pair.CurrencyPair, amount float64) (string, error) {
	return "", errors.New("not yet implemented")
}
