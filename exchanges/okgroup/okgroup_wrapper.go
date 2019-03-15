package okgroup

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// Note: GoCryptoTrader wrapper funcs currently only support SPOT trades.
// Therefore this OKGroup_Wrapper can be shared between OKEX and OKCoin.
// When circumstances change, wrapper funcs can be split appropriately

// Start starts the OKGroup go routine
func (o *OKGroup) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		o.Run()
		wg.Done()
	}()
}

// Run implements the OKEX wrapper
func (o *OKGroup) Run() {
	if o.Verbose {
		log.Debugf("%s Websocket: %s. (url: %s).\n", o.GetName(), common.IsEnabled(o.Websocket.IsEnabled()), o.WebsocketURL)
		log.Debugf("%s polling delay: %ds.\n", o.GetName(), o.RESTPollingDelay)
		log.Debugf("%s %d currencies enabled: %s.\n", o.GetName(), len(o.EnabledPairs), o.EnabledPairs)
	}

	prods, err := o.GetSpotTokenPairDetails()
	if err != nil {
		log.Errorf("%v failed to obtain available spot instruments. Err: %d", o.Name, err)
		return
	}

	var pairs []string
	for x := range prods {
		pairs = append(pairs, prods[x].BaseCurrency+"_"+prods[x].QuoteCurrency)
	}

	err = o.UpdateCurrencies(pairs, false, false)
	if err != nil {
		log.Errorf("%v failed to update available currencies. Err: %s", o.Name, err)
		return
	}
}

// UpdateTicker updates and returns the ticker for a currency pair
func (o *OKGroup) UpdateTicker(p pair.CurrencyPair, assetType string) (tickerData ticker.Price, err error) {
	resp, err := o.GetSpotAllTokenPairsInformationForCurrency(exchange.FormatExchangeCurrency(o.Name, p).String())
	if err != nil {
		return
	}
	respTime, err := time.Parse(time.RFC3339Nano, resp.Timestamp)
	if err != nil {
		log.Warnf("Exchange %v Func %v Currency %v Could not parse date to unix with value of %v",
			o.Name, "UpdateTicker", exchange.FormatExchangeCurrency(o.Name, p).String(), resp.Timestamp)
	}
	tickerData = ticker.Price{
		Ask:          resp.BestAsk,
		Bid:          resp.BestBid,
		CurrencyPair: exchange.FormatExchangeCurrency(o.Name, p).String(),
		High:         resp.High24h,
		Last:         resp.Last,
		LastUpdated:  respTime,
		Low:          resp.Low24h,
		Pair:         p,
		Volume:       resp.BaseVolume24h,
	}

	ticker.ProcessTicker(o.Name, p, tickerData, assetType)

	return
}

// GetTickerPrice returns the ticker for a currency pair
func (o *OKGroup) GetTickerPrice(p pair.CurrencyPair, assetType string) (tickerData ticker.Price, err error) {
	tickerData, err = ticker.GetTicker(o.GetName(), p, assetType)
	if err != nil {
		return o.UpdateTicker(p, assetType)
	}
	return
}

// GetOrderbookEx returns orderbook base on the currency pair
func (o *OKGroup) GetOrderbookEx(p pair.CurrencyPair, assetType string) (resp orderbook.Base, err error) {
	ob, err := orderbook.GetOrderbook(o.GetName(), p, assetType)
	if err != nil {
		return o.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (o *OKGroup) UpdateOrderbook(p pair.CurrencyPair, assetType string) (resp orderbook.Base, err error) {
	orderbookNew, err := o.GetSpotOrderBook(GetSpotOrderBookRequest{
		InstrumentID: exchange.FormatExchangeCurrency(o.Name, p).String(),
	})
	if err != nil {
		return
	}

	for x := range orderbookNew.Bids {
		amount, err := strconv.ParseFloat(orderbookNew.Bids[x][2], 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", orderbookNew.Bids[x][2])
		}
		price, err := strconv.ParseFloat(orderbookNew.Asks[x][0], 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", orderbookNew.Bids[x][2])
		}
		resp.Bids = append(resp.Bids, orderbook.Item{
			Amount: amount,
			Price:  price,
		})
	}

	for x := range orderbookNew.Asks {
		amount, err := strconv.ParseFloat(orderbookNew.Asks[x][2], 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", orderbookNew.Asks[x][2])
		}
		price, err := strconv.ParseFloat(orderbookNew.Asks[x][0], 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", orderbookNew.Asks[x][2])
		}
		resp.Asks = append(resp.Asks, orderbook.Item{
			Amount: amount,
			Price:  price,
		})
	}

	orderbook.ProcessOrderbook(o.GetName(), p, resp, assetType)
	return orderbook.GetOrderbook(o.Name, p, assetType)
}

// GetAccountInfo retrieves balances for all enabled currencies
func (o *OKGroup) GetAccountInfo() (resp exchange.AccountInfo, err error) {
	resp.Exchange = o.Name
	currencies, err := o.GetSpotTradingAccounts()
	currencyAccount := exchange.Account{}

	for _, curr := range currencies {
		hold, err := strconv.ParseFloat(curr.Hold, 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", curr.Hold)
		}
		totalValue, err := strconv.ParseFloat(curr.Balance, 64)
		if err != nil {
			log.Errorf("Could not convert %v to float64", curr.Balance)
		}
		currencyAccount.Currencies = append(currencyAccount.Currencies, exchange.AccountCurrencyInfo{
			CurrencyName: curr.ID,
			Hold:         hold,
			TotalValue:   totalValue,
		})
	}

	resp.Accounts = append(resp.Accounts, currencyAccount)
	return
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (o *OKGroup) GetFundingHistory() (resp []exchange.FundHistory, err error) {
	accountDepositHistory, err := o.GetAccountDepositHistory("")
	for _, deposit := range accountDepositHistory {
		timeStamp, err := time.Parse(time.RFC3339Nano, deposit.Timestamp)
		if err != nil {
			log.Warnf("Exchange %v Func %v Could not parse date to unix with value of %v",
				o.Name, "GetFundingHistory", deposit.Timestamp)
		}
		orderStatus := ""
		switch deposit.Status {
		case 0:
			orderStatus = "waiting"
		case 1:
			orderStatus = "confirmation account"
		case 2:
			orderStatus = "recharge success"
		}

		resp = append(resp, exchange.FundHistory{
			Amount:       deposit.Amount,
			Currency:     deposit.Currency,
			ExchangeName: o.Name,
			Status:       orderStatus,
			Timestamp:    timeStamp,
			TransferID:   deposit.TransactionID,
			TransferType: "deposit",
		})
	}
	accountWithdrawlHistory, err := o.GetAccountWithdrawalHistory("")
	for _, withdrawal := range accountWithdrawlHistory {
		timeStamp, err := time.Parse(time.RFC3339Nano, withdrawal.Timestamp)
		if err != nil {
			log.Warnf("Exchange %v Func %v Could not parse date to unix with value of %v",
				o.Name, "GetFundingHistory", withdrawal.Timestamp)
		}
		resp = append(resp, exchange.FundHistory{
			Amount:       withdrawal.Amount,
			Currency:     withdrawal.Currency,
			ExchangeName: o.Name,
			Status:       OrderStatus[withdrawal.Status],
			Timestamp:    timeStamp,
			TransferID:   withdrawal.Txid,
			TransferType: "withdrawal",
		})
	}
	return
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (o *OKGroup) GetExchangeHistory(p pair.CurrencyPair, assetType string) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (o *OKGroup) SubmitOrder(p pair.CurrencyPair, side exchange.OrderSide, orderType exchange.OrderType, amount, price float64, clientID string) (resp exchange.SubmitOrderResponse, err error) {
	request := PlaceSpotOrderRequest{
		ClientOID:    clientID,
		InstrumentID: exchange.FormatExchangeCurrency(o.Name, p).String(),
		Side:         strings.ToLower(side.ToString()),
		Type:         strings.ToLower(orderType.ToString()),
		Size:         strconv.FormatFloat(amount, 'f', -1, 64),
	}
	if orderType == exchange.LimitOrderType {
		request.Price = strconv.FormatFloat(price, 'f', -1, 64)
	}

	orderResponse, err := o.PlaceSpotOrder(request)
	if err != nil {
		return
	}
	resp.IsOrderPlaced = orderResponse.Result
	resp.OrderID = orderResponse.OrderID

	return
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (o *OKGroup) ModifyOrder(action exchange.ModifyOrder) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (o *OKGroup) CancelOrder(orderCancellation exchange.OrderCancellation) (err error) {
	orderID, err := strconv.ParseInt(orderCancellation.OrderID, 10, 64)
	if err != nil {
		return
	}
	orderCancellationResponse, err := o.CancelSpotOrder(CancelSpotOrderRequest{
		InstrumentID: exchange.FormatExchangeCurrency(o.Name, orderCancellation.CurrencyPair).String(),
		OrderID:      orderID,
	})
	if !orderCancellationResponse.Result {
		err = fmt.Errorf("Order %v failed to be cancelled", orderCancellationResponse.OrderID)
	}

	return
}

// CancelAllOrders cancels all orders associated with a currency pair
func (o *OKGroup) CancelAllOrders(orderCancellation exchange.OrderCancellation) (resp exchange.CancelAllOrdersResponse, err error) {
	orderIDs := strings.Split(orderCancellation.OrderID, ",")
	var orderIDNumbers []int64
	for _, i := range orderIDs {
		orderIDNumber, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return resp, err
		}
		orderIDNumbers = append(orderIDNumbers, orderIDNumber)
	}

	cancelOrdersResponse, err := o.CancelMultipleSpotOrders(CancelMultipleSpotOrdersRequest{
		InstrumentID: exchange.FormatExchangeCurrency(o.Name, orderCancellation.CurrencyPair).String(),
		OrderIDs:     orderIDNumbers,
	})
	if err != nil {
		return
	}

	for _, orderMap := range cancelOrdersResponse {
		for _, cancelledOrder := range orderMap {
			resp.OrderStatus[fmt.Sprintf("%v", cancelledOrder.OrderID)] = fmt.Sprintf("%v", cancelledOrder.Result)
		}
	}

	return
}

// GetOrderInfo returns information on a current open order
func (o *OKGroup) GetOrderInfo(orderID int64) (resp exchange.OrderDetail, err error) {
	order, err := o.GetSpotOrder(GetSpotOrderRequest{OrderID: orderID})
	if err != nil {
		return
	}
	timestamp, err := time.Parse(time.RFC3339Nano, order.Timestamp)
	if err != nil {
		log.Warnf("Exchange %v Func %v Could not parse date to unix with value of %v",
			o.Name, "GetOrderInfo", order.Timestamp)
	}
	resp = exchange.OrderDetail{
		Amount:         order.Size,
		CurrencyPair:   pair.NewCurrencyPairDelimiter(order.InstrumentID, o.ConfigCurrencyPairFormat.Delimiter),
		Exchange:       o.Name,
		OrderDate:      timestamp,
		ExecutedAmount: order.FilledSize,
		Status:         order.Status,
		OrderSide:      exchange.OrderSide(order.Side),
	}
	return
}

// GetDepositAddress returns a deposit address for a specified currency
func (o *OKGroup) GetDepositAddress(p pair.CurrencyItem, accountID string) (_ string, err error) {
	wallet, err := o.GetAccountDepositAddressForCurrency(p.Lower().String())
	if err != nil {
		return
	}
	return wallet[0].Address, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (o *OKGroup) WithdrawCryptocurrencyFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	withdrawal, err := o.AccountWithdraw(AccountWithdrawRequest{
		Amount:      withdrawRequest.Amount,
		Currency:    withdrawRequest.Currency.Lower().String(),
		Destination: 4,
		Fee:         withdrawRequest.FeeAmount,
		ToAddress:   withdrawRequest.Address,
		TradePwd:    withdrawRequest.TradePassword,
	})
	if err != nil {
		return "", err
	}
	if !withdrawal.Result {
		return fmt.Sprintf("%v", withdrawal.WithdrawalID), fmt.Errorf("Could not withdraw currency %v to %v, no error specified", withdrawRequest.Currency.String(), withdrawRequest.Address)
	}

	return fmt.Sprintf("%v", withdrawal.WithdrawalID), nil
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (o *OKGroup) WithdrawFiatFunds(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (o *OKGroup) WithdrawFiatFundsToInternationalBank(withdrawRequest exchange.WithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetActiveOrders retrieves any orders that are active/open
func (o *OKGroup) GetActiveOrders(getOrdersRequest exchange.GetOrdersRequest) (resp []exchange.OrderDetail, err error) {
	for _, currency := range getOrdersRequest.Currencies {
		spotOpenOrders, err := o.GetSpotOpenOrders(GetSpotOpenOrdersRequest{
			InstrumentID: exchange.FormatExchangeCurrency(o.Name, currency).String(),
		})
		if err != nil {
			return resp, err
		}
		for _, openOrder := range spotOpenOrders {
			timestamp, err := time.Parse(time.RFC3339Nano, openOrder.Timestamp)
			if err != nil {
				log.Warnf("Exchange %v Func %v Could not parse date to unix with value of %v",
					o.Name, "GetOrderInfo", openOrder.Timestamp)
			}
			resp = append(resp, exchange.OrderDetail{
				Amount:         openOrder.Size,
				CurrencyPair:   currency,
				Exchange:       o.Name,
				ExecutedAmount: openOrder.FilledSize,
				OrderDate:      timestamp,
				Status:         openOrder.Status,
			})
		}
	}

	return
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (o *OKGroup) GetOrderHistory(getOrdersRequest exchange.GetOrdersRequest) (resp []exchange.OrderDetail, err error) {
	for _, currency := range getOrdersRequest.Currencies {
		spotOpenOrders, err := o.GetSpotOrders(GetSpotOrdersRequest{
			InstrumentID: exchange.FormatExchangeCurrency(o.Name, currency).String(),
		})
		if err != nil {
			return resp, err
		}
		for _, openOrder := range spotOpenOrders {
			timestamp, err := time.Parse(time.RFC3339Nano, openOrder.Timestamp)
			if err != nil {
				log.Warnf("Exchange %v Func %v Could not parse date to unix with value of %v",
					o.Name, "GetOrderInfo", openOrder.Timestamp)
			}
			resp = append(resp, exchange.OrderDetail{
				Amount:         openOrder.Size,
				CurrencyPair:   currency,
				Exchange:       o.Name,
				ExecutedAmount: openOrder.FilledSize,
				OrderDate:      timestamp,
				Status:         openOrder.Status,
			})
		}
	}

	return
}

// GetWebsocket returns a pointer to the exchange websocket
func (o *OKGroup) GetWebsocket() (*exchange.Websocket, error) {
	return o.Websocket, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (o *OKGroup) GetFeeByType(feeBuilder exchange.FeeBuilder) (float64, error) {
	return o.GetFee(feeBuilder)
}

// GetWithdrawCapabilities returns the types of withdrawal methods permitted by the exchange
func (o *OKGroup) GetWithdrawCapabilities() uint32 {
	return o.GetWithdrawPermissions()
}
