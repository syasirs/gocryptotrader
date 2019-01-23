package bitmex

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// Start starts the Bitmex go routine
func (b *Bitmex) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the Bitmex wrapper
func (b *Bitmex) Run() {
	if b.Verbose {
		log.Debugf("%s Websocket: %s. (url: %s).\n", b.GetName(), common.IsEnabled(b.Websocket.IsEnabled()), b.WebsocketURL)
		log.Debugf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay)
		log.Debugf("%s %d currencies enabled: %s.\n", b.GetName(), len(b.EnabledPairs), b.EnabledPairs)
	}

	marketInfo, err := b.GetActiveInstruments(GenericRequestParams{})
	if err != nil {
		log.Errorf("%s Failed to get available symbols.\n", b.GetName())

	} else {
		var exchangeProducts []string
		for _, info := range marketInfo {
			exchangeProducts = append(exchangeProducts, info.Symbol)
		}

		err = b.UpdateCurrencies(exchangeProducts, false, false)
		if err != nil {
			log.Errorf("%s Failed to update available currencies.\n", b.GetName())
		}
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bitmex) UpdateTicker(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	var tickerPrice ticker.Price
	currency := exchange.FormatExchangeCurrency(b.Name, p)

	tick, err := b.GetTrade(GenericRequestParams{
		Symbol:    currency.String(),
		StartTime: time.Now().Format(time.RFC3339),
		Reverse:   true,
		Count:     1})
	if err != nil {
		return tickerPrice, err
	}

	if len(tick) == 0 {
		return tickerPrice, errors.New("Bitmex REST error: no ticker return")
	}

	tickerPrice.Pair = p
	tickerPrice.LastUpdated = time.Now()
	tickerPrice.CurrencyPair = tick[0].Symbol
	tickerPrice.Last = tick[0].Price
	tickerPrice.Volume = float64(tick[0].Size)

	ticker.ProcessTicker(b.Name, p, tickerPrice, assetType)

	return tickerPrice, nil
}

// GetTickerPrice returns the ticker for a currency pair
func (b *Bitmex) GetTickerPrice(p pair.CurrencyPair, assetType string) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (b *Bitmex) GetOrderbookEx(currency pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	ob, err := orderbook.GetOrderbook(b.GetName(), currency, assetType)
	if err != nil {
		return b.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bitmex) UpdateOrderbook(p pair.CurrencyPair, assetType string) (orderbook.Base, error) {
	var orderBook orderbook.Base

	orderbookNew, err := b.GetOrderbook(OrderBookGetL2Params{
		Symbol: exchange.FormatExchangeCurrency(b.Name, p).String(),
		Depth:  500})
	if err != nil {
		return orderBook, err
	}

	for _, ob := range orderbookNew {
		if ob.Side == exchange.Sell.ToString() {
			orderBook.Asks = append(orderBook.Asks,
				orderbook.Item{Amount: float64(ob.Size), Price: ob.Price})
			continue
		}
		if ob.Side == exchange.Buy.ToString() {
			orderBook.Bids = append(orderBook.Bids,
				orderbook.Item{Amount: float64(ob.Size), Price: ob.Price})
			continue
		}
	}
	orderbook.ProcessOrderbook(b.GetName(), p, orderBook, assetType)

	return orderbook.GetOrderbook(b.Name, p, assetType)
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// Bitmex exchange
func (b *Bitmex) GetAccountInfo() (exchange.AccountInfo, error) {
	var info exchange.AccountInfo

	bal, err := b.GetAllUserMargin()
	if err != nil {
		return info, err
	}

	// Need to update to add Margin/Liquidity availibilty
	var balances []exchange.AccountCurrencyInfo
	for _, data := range bal {
		balances = append(balances, exchange.AccountCurrencyInfo{
			CurrencyName: data.Currency,
			TotalValue:   float64(data.WalletBalance),
		})
	}

	info.Exchange = b.GetName()
	info.Accounts = append(info.Accounts, exchange.Account{
		Currencies: balances,
	})

	return info, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bitmex) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	// b.GetFullFundingHistory()
	return fundHistory, common.ErrNotYetImplemented
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (b *Bitmex) GetExchangeHistory(p pair.CurrencyPair, assetType string) ([]exchange.TradeHistory, error) {
	var resp []exchange.TradeHistory

	return resp, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (b *Bitmex) SubmitOrder(p pair.CurrencyPair, side exchange.OrderSide, orderType exchange.OrderType, amount, price float64, clientID string) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse

	if math.Mod(amount, 1) != 0 {
		return submitOrderResponse,
			errors.New("contract amount can not have decimals")
	}

	var orderNewParams = OrderNewParams{
		OrdType:  side.ToString(),
		Symbol:   p.Pair().String(),
		OrderQty: amount,
		Side:     side.ToString(),
	}

	if orderType == exchange.LimitOrderType {
		orderNewParams.Price = price
	}

	response, err := b.CreateOrder(orderNewParams)
	if response.OrderID != "" {
		submitOrderResponse.OrderID = response.OrderID
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bitmex) ModifyOrder(action exchange.ModifyOrder) (string, error) {
	var params OrderAmendParams

	if math.Mod(action.Amount, 1) != 0 {
		return "", errors.New("contract amount can not have decimals")
	}

	params.OrderID = action.OrderID
	params.OrderQty = int32(action.Amount)
	params.Price = action.Price

	order, err := b.AmendOrder(params)
	if err != nil {
		return "", err
	}

	return order.OrderID, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bitmex) CancelOrder(order exchange.OrderCancellation) error {
	var params = OrderCancelParams{
		OrderID: order.OrderID,
	}
	_, err := b.CancelOrders(params)

	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bitmex) CancelAllOrders(orderCancellation exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}
	var emptyParams OrderCancelAllParams
	orders, err := b.CancelAllExistingOrders(emptyParams)
	if err != nil {
		return cancelAllOrdersResponse, err
	}

	for _, order := range orders {
		cancelAllOrdersResponse.OrderStatus[order.OrderID] = order.OrdRejReason
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (b *Bitmex) GetOrderInfo(orderID int64) (exchange.OrderDetail, error) {
	var orderDetail exchange.OrderDetail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *Bitmex) GetDepositAddress(cryptocurrency pair.CurrencyItem, accountID string) (string, error) {
	return b.GetCryptoDepositAddress(cryptocurrency.String())
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawCryptocurrencyFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	var request = UserRequestWithdrawalParams{
		Address:  withdrawRequest.Address,
		Amount:   withdrawRequest.Amount,
		Currency: withdrawRequest.Currency.String(),
		OtpToken: withdrawRequest.OneTimePassword,
	}
	if withdrawRequest.FeeAmount > 0 {
		request.Fee = withdrawRequest.FeeAmount
	}

	resp, err := b.UserRequestWithdrawal(request)
	if err != nil {
		return "", err
	}

	return resp.TransactID, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawFiatFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitmex) WithdrawFiatFundsToInternationalBank(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (b *Bitmex) GetWebsocket() (*exchange.Websocket, error) {
	return b.Websocket, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bitmex) GetFeeByType(feeBuilder exchange.FeeBuilder) (float64, error) {
	return b.GetFee(feeBuilder)
}

// GetWithdrawCapabilities returns the types of withdrawal methods permitted by the exchange
func (b *Bitmex) GetWithdrawCapabilities() uint32 {
	return b.GetWithdrawPermissions()
}

// GetActiveOrders retrieves any orders that are active/open
func (b *Bitmex) GetActiveOrders(getOrdersRequest exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var orders []exchange.OrderDetail
	params := OrdersRequest{}
	params.Filter = "{\"open\":true}"

	resp, err := b.GetOrders(params)
	if err != nil {
		return nil, err
	}

	for _, order := range resp {
		var orderType, orderSide string
		switch order.OrdType {
		case "1":
			orderType = string(exchange.MarketOrderType)
		case "2":
			orderType = string(exchange.LimitOrderType)
		case "3":
			orderType = string(exchange.StopOrderType)
		case "7":
			orderType = string(exchange.TrailingStopOrderType)
		default:
			log.Warnf("Uknown order type: '%v' returned for orderID '%v'. Leaving blank", order.OrdType, order.OrderID)
			orderType = string(exchange.UnknownOrderType)
		}

		if order.Side == "1" {
			orderSide = string(exchange.BuyOrderSide)
		} else if order.Side == "1" {
			orderSide = string(exchange.SellOrderSide)
		}

		orderDetail := exchange.OrderDetail{
			Price:         order.Price,
			Amount:        float64(order.OrderQty),
			BaseCurrency:  order.Symbol,
			QuoteCurrency: order.SettlCurrency,
			Exchange:      b.Name,
			ID:            order.OrderID,
			OrderSide:     orderSide,
			OrderType:     orderType,
			Status:        order.OrdStatus,
		}

		orders = append(orders, orderDetail)
	}

	b.FilterOrdersByType(&orders, getOrdersRequest.OrderType)
	b.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks, getOrdersRequest.EndTicks)
	b.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)

	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *Bitmex) GetOrderHistory(getOrdersRequest exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var orders []exchange.OrderDetail
	params := OrdersRequest{}
	resp, err := b.GetOrders(params)
	if err != nil {
		return nil, err
	}

	for _, order := range resp {
		var orderType, orderSide string
		switch order.OrdType {
		case "1":
			orderType = string(exchange.MarketOrderType)
		case "2":
			orderType = string(exchange.LimitOrderType)
		case "3":
			orderType = string(exchange.StopOrderType)
		case "7":
			orderType = string(exchange.TrailingStopOrderType)
		default:
			log.Warnf("Uknown order type: '%v' returned for orderID '%v'. Leaving blank", order.OrdType, order.OrderID)
			orderType = string(exchange.UnknownOrderType)
		}

		if order.Side == "1" {
			orderSide = string(exchange.BuyOrderSide)
		} else if order.Side == "1" {
			orderSide = string(exchange.SellOrderSide)
		}

		orderDetail := exchange.OrderDetail{
			Price:         order.Price,
			Amount:        float64(order.OrderQty),
			BaseCurrency:  order.Symbol,
			QuoteCurrency: order.SettlCurrency,
			Exchange:      b.Name,
			ID:            order.OrderID,
			OrderSide:     orderSide,
			OrderType:     orderType,
			Status:        order.OrdStatus,
		}

		orders = append(orders, orderDetail)
	}

	b.FilterOrdersByType(&orders, getOrdersRequest.OrderType)
	b.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks, getOrdersRequest.EndTicks)
	b.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)

	return orders, nil
}
