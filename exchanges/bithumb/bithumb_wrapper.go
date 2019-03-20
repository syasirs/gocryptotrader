package bithumb

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// Start starts the bithumb go routine
func (b *Bithumb) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the bithumb wrapper
func (b *Bithumb) Run() {
	if b.Verbose {
		log.Debugf("%s Websocket: %s. (url: %s).\n", b.GetName(), common.IsEnabled(b.Websocket.IsEnabled()), b.WebsocketURL)
		log.Debugf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay)
		log.Debugf("%s %d currencies enabled: %s.\n", b.GetName(), len(b.EnabledPairs), b.EnabledPairs)
	}

	exchangeProducts, err := b.GetTradingPairs()
	if err != nil {
		log.Errorf("%s Failed to get available symbols.\n", b.GetName())
	} else {
		var newExchangeProducts currency.Pairs
		for _, p := range exchangeProducts {
			newExchangeProducts = append(newExchangeProducts,
				currency.NewPairFromString(p))
		}

		err = b.UpdateCurrencies(newExchangeProducts, false, false)
		if err != nil {
			log.Errorf("%s Failed to update available symbols.\n", b.GetName())
		}
	}
}

// GetTradingPairs gets the available trading currencies
func (b *Bithumb) GetTradingPairs() ([]string, error) {
	currencies, err := b.GetTradablePairs()
	if err != nil {
		return nil, err
	}

	for x := range currencies {
		currencies[x] += "KRW"
	}

	return currencies, nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bithumb) UpdateTicker(p currency.Pair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price

	tickers, err := b.GetAllTickers()
	if err != nil {
		return tickerPrice, err
	}

	for _, x := range b.GetEnabledCurrencies() {
		currency := x.Base.String()
		var tp ticker.Price
		tp.Pair = x
		tp.Ask = tickers[currency].SellPrice
		tp.Bid = tickers[currency].BuyPrice
		tp.Low = tickers[currency].MinPrice
		tp.Last = tickers[currency].ClosingPrice
		tp.Volume = tickers[currency].Volume1Day
		tp.High = tickers[currency].MaxPrice

		err = ticker.ProcessTicker(b.Name, tp, assetType)
		if err != nil {
			return tickerPrice, err
		}
	}
	return ticker.GetTicker(b.Name, p, assetType)
}

// GetTickerPrice returns the ticker for a currency pair
func (b *Bithumb) GetTickerPrice(p currency.Pair, assetType string) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (b *Bithumb) GetOrderbookEx(currency currency.Pair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.Get(b.GetName(), currency, assetType)
	if err != nil {
		return b.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bithumb) UpdateOrderbook(p currency.Pair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base
	currency := p.Base.String()

	orderbookNew, err := b.GetOrderBook(currency)
	if err != nil {
		return orderBook, err
	}

	for _, bids := range orderbookNew.Data.Bids {
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: bids.Quantity, Price: bids.Price})
	}

	for _, asks := range orderbookNew.Data.Asks {
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: asks.Quantity, Price: asks.Price})
	}

	orderBook.Pair = p
	orderBook.ExchangeName = b.GetName()
	orderBook.AssetType = assetType

	err = orderBook.Process()
	if err != nil {
		return orderBook, err
	}

	return orderbook.Get(b.Name, p, assetType)
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// Bithumb exchange
func (b *Bithumb) GetAccountInfo() (exchange.AccountInfo, error) {
	var info exchange.AccountInfo
	bal, err := b.GetAccountBalance("ALL")
	if err != nil {
		return info, err
	}

	var exchangeBalances []exchange.AccountCurrencyInfo
	for key, totalAmount := range bal.Total {
		hold, ok := bal.InUse[key]
		if !ok {
			return info, fmt.Errorf("getAccountInfo error - in use item not found for currency %s",
				key)
		}

		exchangeBalances = append(exchangeBalances, exchange.AccountCurrencyInfo{
			CurrencyName: currency.NewCode(key),
			TotalValue:   totalAmount,
			Hold:         hold,
		})
	}

	info.Accounts = append(info.Accounts, exchange.Account{
		Currencies: exchangeBalances,
	})

	info.Exchange = b.GetName()
	return info, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bithumb) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetPlatformHistory returns historic platform trade data since exchange
// initial operations
func (b *Bithumb) GetPlatformHistory(p currency.Pair, assetType string, timestampStart time.Time, tradeID string) ([]exchange.PlatformTrade, error) {
	var resp []exchange.PlatformTrade
	ID, err := strconv.ParseInt(tradeID, 10, 64)
	if err != nil && tradeID != "" {
		return nil, err
	}

	t, err := b.GetTransactionHistory(p.Base.String(), ID)
	if err != nil {
		return resp, err
	}

	for i := range t.Data {
		orderID := strconv.FormatInt(t.Data[i].ContNumber, 10)
		resp = append(resp, exchange.PlatformTrade{
			Timestamp: ConvertToRFC3339(t.Data[i].TransactionDate),
			TID:       orderID,
			Price:     t.Data[i].Price,
			Amount:    t.Data[i].UnitsTraded,
			Exchange:  b.GetName(),
			Type:      t.Data[i].Type,
		})
	}
	return resp, nil
}

// ConvertToRFC3339 converts string from bithumb to a RFC3339 format
func ConvertToRFC3339(t string) time.Time {
	split := common.SplitStrings(t, " ")
	join := common.JoinStrings(split, "T")
	join += "Z"
	newTime, _ := time.Parse(time.RFC3339, join)
	return newTime
}

// SubmitOrder submits a new order
// TODO: Fill this out to support limit orders
func (b *Bithumb) SubmitOrder(p currency.Pair, side exchange.OrderSide, _ exchange.OrderType, amount, _ float64, _ string) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse
	var err error
	var orderID string
	if side == exchange.BuyOrderSide {
		var result MarketBuy
		result, err = b.MarketBuyOrder(p.Base.String(), amount)
		orderID = result.OrderID
	} else if side == exchange.SellOrderSide {
		var result MarketSell
		result, err = b.MarketSellOrder(p.Base.String(), amount)
		orderID = result.OrderID
	}

	if orderID != "" {
		submitOrderResponse.OrderID = fmt.Sprintf("%v", orderID)
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bithumb) ModifyOrder(action exchange.ModifyOrder) (string, error) {
	order, err := b.ModifyTrade(action.OrderID,
		action.CurrencyPair.Base.String(),
		common.StringToLower(action.OrderSide.ToString()),
		action.Amount,
		int64(action.Price))

	if err != nil {
		return "", err
	}

	return order.Data[0].ContID, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bithumb) CancelOrder(order exchange.OrderCancellation) error {
	_, err := b.CancelTrade(order.Side.ToString(),
		order.OrderID,
		order.CurrencyPair.Base.String())
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bithumb) CancelAllOrders(orderCancellation exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}
	var allOrders []OrderData

	for _, currency := range b.GetEnabledCurrencies() {
		orders, err := b.GetOrders("",
			orderCancellation.Side.ToString(),
			"100",
			"",
			currency.Base.String())
		if err != nil {
			return cancelAllOrdersResponse, err
		}
		allOrders = append(allOrders, orders.Data...)
	}

	for _, order := range allOrders {
		_, err := b.CancelTrade(orderCancellation.Side.ToString(),
			order.OrderID,
			orderCancellation.CurrencyPair.Base.String())
		if err != nil {
			cancelAllOrdersResponse.OrderStatus[order.OrderID] = err.Error()
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (b *Bithumb) GetOrderInfo(orderID string) (exchange.OrderDetail, error) {
	var orderDetail exchange.OrderDetail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *Bithumb) GetDepositAddress(cryptocurrency currency.Code, _ string) (string, error) {
	addr, err := b.GetWalletAddress(cryptocurrency.String())
	if err != nil {
		return "", err
	}

	return addr.Data.WalletAddress, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bithumb) WithdrawCryptocurrencyFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	_, err := b.WithdrawCrypto(withdrawRequest.Address, withdrawRequest.AddressTag, withdrawRequest.Currency.String(), withdrawRequest.Amount)
	return "", err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bithumb) WithdrawFiatFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	if math.Mod(withdrawRequest.Amount, 1) != 0 {
		return "", errors.New("currency KRW does not support decimal places")
	}
	if withdrawRequest.Currency != currency.KRW {
		return "", errors.New("only KRW is supported")
	}
	bankDetails := fmt.Sprintf("%v_%v", withdrawRequest.BankCode, withdrawRequest.BankName)
	bankAccountNumber := strconv.FormatFloat(withdrawRequest.BankAccountNumber, 'f', -1, 64)
	withdrawAmountInt := int64(withdrawRequest.Amount)
	resp, err := b.RequestKRWWithdraw(bankDetails, bankAccountNumber, withdrawAmountInt)
	if err != nil {
		return "", err
	}
	if resp.Status != "0000" {
		return "", errors.New(resp.Message)
	}

	return resp.Message, nil
}

// WithdrawFiatFundsToInternationalBank is not supported as Bithumb only withdraws KRW to South Korean banks
func (b *Bithumb) WithdrawFiatFundsToInternationalBank(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (b *Bithumb) GetWebsocket() (*exchange.Websocket, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bithumb) GetFeeByType(feeBuilder exchange.FeeBuilder) (float64, error) {
	return b.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (b *Bithumb) GetActiveOrders(getOrdersRequest exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var orders []exchange.OrderDetail
	resp, err := b.GetOrders("", "", "1000", "", "")
	if err != nil {
		return nil, err
	}

	for _, order := range resp.Data {
		if order.Status != "placed" {
			continue
		}

		orderDate := time.Unix(order.OrderDate, 0)
		orderDetail := exchange.OrderDetail{
			Amount:          order.Units,
			Exchange:        b.Name,
			ID:              order.OrderID,
			OrderDate:       orderDate,
			Price:           order.Price,
			RemainingAmount: order.UnitsRemaining,
			Status:          string(exchange.ActiveOrderStatus),
			CurrencyPair: currency.NewPairWithDelimiter(order.OrderCurrency,
				order.PaymentCurrency,
				b.ConfigCurrencyPairFormat.Delimiter),
		}

		if order.Type == "bid" {
			orderDetail.OrderSide = exchange.BuyOrderSide
		} else if order.Type == "ask" {
			orderDetail.OrderSide = exchange.SellOrderSide
		}

		orders = append(orders, orderDetail)
	}

	exchange.FilterOrdersBySide(&orders, getOrdersRequest.OrderSide)
	exchange.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks,
		getOrdersRequest.EndTicks)
	exchange.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)

	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *Bithumb) GetOrderHistory(getOrdersRequest exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var orders []exchange.OrderDetail
	resp, err := b.GetOrders("", "", "1000", "", "")
	if err != nil {
		return nil, err
	}

	for _, order := range resp.Data {
		if order.Status == "placed" {
			continue
		}

		orderDate := time.Unix(order.OrderDate, 0)
		orderDetail := exchange.OrderDetail{
			Amount:          order.Units,
			Exchange:        b.Name,
			ID:              order.OrderID,
			OrderDate:       orderDate,
			Price:           order.Price,
			RemainingAmount: order.UnitsRemaining,
			CurrencyPair: currency.NewPairWithDelimiter(order.OrderCurrency,
				order.PaymentCurrency,
				b.ConfigCurrencyPairFormat.Delimiter),
		}

		if order.Type == "bid" {
			orderDetail.OrderSide = exchange.BuyOrderSide
		} else if order.Type == "ask" {
			orderDetail.OrderSide = exchange.SellOrderSide
		}

		orders = append(orders, orderDetail)
	}

	exchange.FilterOrdersBySide(&orders, getOrdersRequest.OrderSide)
	exchange.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks,
		getOrdersRequest.EndTicks)
	exchange.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)

	return orders, nil
}
