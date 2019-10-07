package coinut

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// GetDefaultConfig returns a default exchange config
func (c *COINUT) GetDefaultConfig() (*config.ExchangeConfig, error) {
	c.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = c.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = c.BaseCurrencies

	err := c.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if c.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = c.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets current default values
func (c *COINUT) SetDefaults() {
	c.Name = "COINUT"
	c.Enabled = true
	c.Verbose = true
	c.API.CredentialsValidator.RequiresKey = true
	c.API.CredentialsValidator.RequiresClientID = true

	c.CurrencyPairs = currency.PairsManager{
		AssetTypes: asset.Items{
			asset.Spot,
		},
		UseGlobalFormat: true,
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: "-",
		},
	}

	c.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
				TickerBatching:  false,
			},
			WithdrawPermissions: exchange.WithdrawCryptoViaWebsiteOnly |
				exchange.WithdrawFiatViaWebsiteOnly,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	c.Requester = request.New(c.Name,
		request.NewRateLimit(time.Second, coinutAuthRate),
		request.NewRateLimit(time.Second, coinutUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	c.API.Endpoints.URLDefault = coinutAPIURL
	c.API.Endpoints.URL = c.API.Endpoints.URLDefault
	c.API.Endpoints.WebsocketURL = coinutWebsocketURL
	c.Websocket = wshandler.New()
	c.Websocket.Functionality = wshandler.WebsocketTickerSupported |
		wshandler.WebsocketOrderbookSupported |
		wshandler.WebsocketTradeDataSupported |
		wshandler.WebsocketSubscribeSupported |
		wshandler.WebsocketUnsubscribeSupported |
		wshandler.WebsocketAuthenticatedEndpointsSupported |
		wshandler.WebsocketSubmitOrderSupported |
		wshandler.WebsocketCancelOrderSupported |
		wshandler.WebsocketMessageCorrelationSupported
	c.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	c.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	c.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup sets the current exchange configuration
func (c *COINUT) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		c.SetEnabled(false)
		return nil
	}

	err := c.SetupDefaults(exch)
	if err != nil {
		return err
	}

	err = c.Websocket.Setup(
		&wshandler.WebsocketSetup{
			Enabled:                          exch.Features.Enabled.Websocket,
			Verbose:                          exch.Verbose,
			AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
			WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
			DefaultURL:                       coinutWebsocketURL,
			ExchangeName:                     exch.Name,
			RunningURL:                       exch.API.Endpoints.WebsocketURL,
			Connector:                        c.WsConnect,
			Subscriber:                       c.Subscribe,
			UnSubscriber:                     c.Unsubscribe,
		})
	if err != nil {
		return err
	}

	c.WebsocketConn = &wshandler.WebsocketConnection{
		ExchangeName:         c.Name,
		URL:                  c.Websocket.GetWebsocketURL(),
		ProxyURL:             c.Websocket.GetProxyAddress(),
		Verbose:              c.Verbose,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	}

	c.Websocket.Orderbook.Setup(
		exch.WebsocketOrderbookBufferLimit,
		true,
		true,
		true,
		false,
		exch.Name)
	return nil
}

// Start starts the COINUT go routine
func (c *COINUT) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		c.Run()
		wg.Done()
	}()
}

// Run implements the COINUT wrapper
func (c *COINUT) Run() {
	if c.Verbose {
		log.Debugf(log.ExchangeSys, "%s Websocket: %s. (url: %s).\n", c.GetName(), common.IsEnabled(c.Websocket.IsEnabled()), coinutWebsocketURL)
		c.PrintEnabledPairs()
	}

	forceUpdate := false
	delim := c.GetPairFormat(asset.Spot, false).Delimiter
	if !common.StringDataContains(c.CurrencyPairs.GetPairs(asset.Spot,
		true).Strings(), delim) ||
		!common.StringDataContains(c.CurrencyPairs.GetPairs(asset.Spot,
			false).Strings(), delim) {
		enabledPairs := currency.NewPairsFromStrings(
			[]string{fmt.Sprintf("LTC%sUSDT", delim)},
		)
		log.Warn(log.ExchangeSys,
			"Enabled pairs for Coinut reset due to config upgrade, please enable the ones you would like to use again")
		forceUpdate = true

		err := c.UpdatePairs(enabledPairs, asset.Spot, true, true)
		if err != nil {
			log.Errorf(log.ExchangeSys, "%s failed to update currencies. Err: %s\n", c.Name, err)
		}
	}

	if !c.GetEnabledFeatures().AutoPairUpdates && !forceUpdate {
		return
	}

	err := c.UpdateTradablePairs(forceUpdate)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", c.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (c *COINUT) FetchTradablePairs(asset asset.Item) ([]string, error) {
	i, err := c.GetInstruments()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for _, y := range i.Instruments {
		c.instrumentMap.Seed(y[0].Base+y[0].Quote, y[0].InstID)
		p := y[0].Base + c.GetPairFormat(asset, false).Delimiter + y[0].Quote
		pairs = append(pairs, p)
	}

	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (c *COINUT) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := c.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	return c.UpdatePairs(currency.NewPairsFromStrings(pairs),
		asset.Spot, false, forceUpdate)
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// COINUT exchange
func (c *COINUT) GetAccountInfo() (exchange.AccountInfo, error) {
	var info exchange.AccountInfo
	bal, err := c.GetUserBalance()
	if err != nil {
		return info, err
	}

	var balances = []exchange.AccountCurrencyInfo{
		{
			CurrencyName: currency.BCH,
			TotalValue:   bal.BCH,
		},
		{
			CurrencyName: currency.BTC,
			TotalValue:   bal.BTC,
		},
		{
			CurrencyName: currency.BTG,
			TotalValue:   bal.BTG,
		},
		{
			CurrencyName: currency.CAD,
			TotalValue:   bal.CAD,
		},
		{
			CurrencyName: currency.ETC,
			TotalValue:   bal.ETC,
		},
		{
			CurrencyName: currency.ETH,
			TotalValue:   bal.ETH,
		},
		{
			CurrencyName: currency.LCH,
			TotalValue:   bal.LCH,
		},
		{
			CurrencyName: currency.LTC,
			TotalValue:   bal.LTC,
		},
		{
			CurrencyName: currency.MYR,
			TotalValue:   bal.MYR,
		},
		{
			CurrencyName: currency.SGD,
			TotalValue:   bal.SGD,
		},
		{
			CurrencyName: currency.USD,
			TotalValue:   bal.USD,
		},
		{
			CurrencyName: currency.USDT,
			TotalValue:   bal.USDT,
		},
		{
			CurrencyName: currency.XMR,
			TotalValue:   bal.XMR,
		},
		{
			CurrencyName: currency.ZEC,
			TotalValue:   bal.ZEC,
		},
	}
	info.Exchange = c.GetName()
	info.Accounts = append(info.Accounts, exchange.Account{
		Currencies: balances,
	})

	return info, nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (c *COINUT) UpdateTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	var tickerPrice ticker.Price

	if !c.instrumentMap.IsLoaded() {
		err := c.SeedInstruments()
		if err != nil {
			return tickerPrice, err
		}
	}

	instID := c.instrumentMap.LookupID(c.FormatExchangeCurrency(p,
		assetType).String())
	if instID == 0 {
		return tickerPrice, errors.New("unable to lookup instrument ID")
	}

	tick, err := c.GetInstrumentTicker(instID)
	if err != nil {
		return tickerPrice, err
	}
	tickerPrice = ticker.Price{
		Last:        tick.Last,
		High:        tick.HighestBuy,
		Low:         tick.LowestSell,
		Volume:      tick.Volume24,
		Pair:        p,
		LastUpdated: time.Unix(0, tick.Timestamp),
	}
	err = ticker.ProcessTicker(c.GetName(), &tickerPrice, assetType)
	if err != nil {
		return tickerPrice, err
	}

	return ticker.GetTicker(c.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (c *COINUT) FetchTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(c.GetName(), p, assetType)
	if err != nil {
		return c.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (c *COINUT) FetchOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	ob, err := orderbook.Get(c.GetName(), p, assetType)
	if err != nil {
		return c.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (c *COINUT) UpdateOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	var orderBook orderbook.Base

	if !c.instrumentMap.IsLoaded() {
		err := c.SeedInstruments()
		if err != nil {
			return orderBook, err
		}
	}

	instID := c.instrumentMap.LookupID(c.FormatExchangeCurrency(p,
		assetType).String())
	if instID == 0 {
		return orderBook, errLookupInstrumentID
	}

	orderbookNew, err := c.GetInstrumentOrderbook(instID, 200)
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Buy {
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: orderbookNew.Buy[x].Quantity, Price: orderbookNew.Buy[x].Price})
	}

	for x := range orderbookNew.Sell {
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: orderbookNew.Sell[x].Quantity, Price: orderbookNew.Sell[x].Price})
	}

	orderBook.Pair = p
	orderBook.ExchangeName = c.GetName()
	orderBook.AssetType = assetType

	err = orderBook.Process()
	if err != nil {
		return orderBook, err
	}

	return orderbook.Get(c.Name, p, assetType)
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (c *COINUT) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory

	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (c *COINUT) GetExchangeHistory(p currency.Pair, assetType asset.Item) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (c *COINUT) SubmitOrder(order *exchange.OrderSubmission) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse
	if order == nil {
		return submitOrderResponse, exchange.ErrOrderSubmissionIsNil
	}

	if err := order.Validate(); err != nil {
		return submitOrderResponse, err
	}

	var APIresponse interface{}
	isBuyOrder := order.OrderSide == exchange.BuyOrderSide
	clientIDInt, err := strconv.ParseUint(order.ClientID, 0, 32)
	if err != nil {
		return submitOrderResponse, err
	}

	clientIDUint := uint32(clientIDInt)

	if !c.instrumentMap.IsLoaded() {
		err = c.SeedInstruments()
		if err != nil {
			return submitOrderResponse, err
		}
	}

	currencyID := c.instrumentMap.LookupID(c.FormatExchangeCurrency(order.Pair,
		asset.Spot).String())
	if currencyID == 0 {
		return submitOrderResponse, errLookupInstrumentID
	}

	switch order.OrderType {
	case exchange.LimitOrderType:
		APIresponse, err = c.NewOrder(currencyID, order.Amount, order.Price,
			isBuyOrder, clientIDUint)
	case exchange.MarketOrderType:
		APIresponse, err = c.NewOrder(currencyID, order.Amount, 0, isBuyOrder,
			clientIDUint)
	}

	switch apiResp := APIresponse.(type) {
	case OrdersBase:
		orderResult := apiResp
		submitOrderResponse.OrderID = fmt.Sprintf("%v", orderResult.OrderID)
	case OrderFilledResponse:
		orderResult := apiResp
		submitOrderResponse.OrderID = fmt.Sprintf("%v", orderResult.Order.OrderID)
	case OrderRejectResponse:
		orderResult := apiResp
		submitOrderResponse.OrderID = fmt.Sprintf("%v", orderResult.OrderID)
		err = fmt.Errorf("orderID: %v was rejected: %v", orderResult.OrderID, orderResult.Reasons)
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (c *COINUT) ModifyOrder(action *exchange.ModifyOrder) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (c *COINUT) CancelOrder(order *exchange.OrderCancellation) error {
	orderIDInt, err := strconv.ParseInt(order.OrderID, 10, 64)
	if err != nil {
		return err
	}

	if !c.instrumentMap.IsLoaded() {
		err = c.SeedInstruments()
		if err != nil {
			return err
		}
	}

	currencyID := c.instrumentMap.LookupID(c.FormatExchangeCurrency(
		order.CurrencyPair,
		asset.Spot).String(),
	)
	if currencyID == 0 {
		return errLookupInstrumentID
	}
	_, err = c.CancelExistingOrder(currencyID, orderIDInt)
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (c *COINUT) CancelAllOrders(_ *exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	// TODO, this is a terrible implementation. Requires DB to improve
	// Coinut provides no way of retrieving orders without a currency
	// So we need to retrieve all currencies, then retrieve orders for each currency
	// Then cancel. Advisable to never use this until DB due to performance
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}

	if !c.instrumentMap.IsLoaded() {
		err := c.SeedInstruments()
		if err != nil {
			return cancelAllOrdersResponse, err
		}
	}

	var allTheOrders []OrderResponse
	ids := c.instrumentMap.GetInstrumentIDs()
	for x := range ids {
		openOrders, err := c.GetOpenOrders(ids[x])
		if err != nil {
			return cancelAllOrdersResponse, err
		}
		allTheOrders = append(allTheOrders, openOrders.Orders...)
	}

	var allTheOrdersToCancel []CancelOrders
	for _, orderToCancel := range allTheOrders {
		cancelOrder := CancelOrders{
			InstrumentID: orderToCancel.InstrumentID,
			OrderID:      orderToCancel.OrderID,
		}
		allTheOrdersToCancel = append(allTheOrdersToCancel, cancelOrder)
	}

	if len(allTheOrdersToCancel) > 0 {
		resp, err := c.CancelOrders(allTheOrdersToCancel)
		if err != nil {
			return cancelAllOrdersResponse, err
		}

		for _, order := range resp.Results {
			if order.Status != "OK" {
				cancelAllOrdersResponse.OrderStatus[strconv.FormatInt(order.OrderID, 10)] = order.Status
			}
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (c *COINUT) GetOrderInfo(orderID string) (exchange.OrderDetail, error) {
	return exchange.OrderDetail{}, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (c *COINUT) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (c *COINUT) WithdrawCryptocurrencyFunds(withdrawRequest *exchange.CryptoWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (c *COINUT) WithdrawFiatFunds(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (c *COINUT) WithdrawFiatFundsToInternationalBank(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (c *COINUT) GetWebsocket() (*wshandler.Websocket, error) {
	return c.Websocket, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (c *COINUT) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !c.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return c.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (c *COINUT) GetActiveOrders(getOrdersRequest *exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	if !c.instrumentMap.IsLoaded() {
		err := c.SeedInstruments()
		if err != nil {
			return nil, err
		}
	}

	var instrumentsToUse []int64
	if len(getOrdersRequest.Currencies) > 0 {
		for x := range getOrdersRequest.Currencies {
			currency := c.FormatExchangeCurrency(getOrdersRequest.Currencies[x],
				asset.Spot).String()
			instrumentsToUse = append(instrumentsToUse,
				c.instrumentMap.LookupID(currency))
		}
	} else {
		instrumentsToUse = c.instrumentMap.GetInstrumentIDs()
	}

	if len(instrumentsToUse) == 0 {
		return nil, errors.New("no instrument IDs to use")
	}

	var orders []exchange.OrderDetail
	for x := range instrumentsToUse {
		openOrders, err := c.GetOpenOrders(instrumentsToUse[x])
		if err != nil {
			return nil, err
		}
		for y := range openOrders.Orders {
			curr := c.instrumentMap.LookupInstrument(instrumentsToUse[x])
			p := currency.NewPairFromFormattedPairs(curr,
				c.GetEnabledPairs(asset.Spot),
				c.GetPairFormat(asset.Spot, true))
			orderSide := exchange.OrderSide(strings.ToUpper(openOrders.Orders[y].Side))
			orderDate := time.Unix(openOrders.Orders[y].Timestamp, 0)
			orders = append(orders, exchange.OrderDetail{
				ID:           strconv.FormatInt(openOrders.Orders[y].OrderID, 10),
				Amount:       openOrders.Orders[y].Quantity,
				Price:        openOrders.Orders[y].Price,
				Exchange:     c.Name,
				OrderSide:    orderSide,
				OrderDate:    orderDate,
				CurrencyPair: p,
			})
		}
	}

	exchange.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks, getOrdersRequest.EndTicks)
	exchange.FilterOrdersBySide(&orders, getOrdersRequest.OrderSide)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (c *COINUT) GetOrderHistory(getOrdersRequest *exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	if !c.instrumentMap.IsLoaded() {
		err := c.SeedInstruments()
		if err != nil {
			return nil, err
		}
	}

	var instrumentsToUse []int64
	if len(getOrdersRequest.Currencies) > 0 {
		for x := range getOrdersRequest.Currencies {
			currency := c.FormatExchangeCurrency(getOrdersRequest.Currencies[x],
				asset.Spot).String()
			instrumentsToUse = append(instrumentsToUse,
				c.instrumentMap.LookupID(currency))
		}
	} else {
		instrumentsToUse = c.instrumentMap.GetInstrumentIDs()
	}

	if len(instrumentsToUse) == 0 {
		return nil, errors.New("no instrument IDs to use")
	}

	var allOrders []exchange.OrderDetail
	for x := range instrumentsToUse {
		orders, err := c.GetTradeHistory(instrumentsToUse[x], -1, -1)
		if err != nil {
			return nil, err
		}
		for y := range orders.Trades {
			curr := c.instrumentMap.LookupInstrument(instrumentsToUse[x])
			p := currency.NewPairFromFormattedPairs(curr,
				c.GetEnabledPairs(asset.Spot),
				c.GetPairFormat(asset.Spot, true))
			orderSide := exchange.OrderSide(
				strings.ToUpper(orders.Trades[y].Order.Side))
			orderDate := time.Unix(orders.Trades[y].Order.Timestamp, 0)
			allOrders = append(allOrders, exchange.OrderDetail{
				ID:           strconv.FormatInt(orders.Trades[y].Order.OrderID, 10),
				Amount:       orders.Trades[y].Order.Quantity,
				Price:        orders.Trades[y].Order.Price,
				Exchange:     c.Name,
				OrderSide:    orderSide,
				OrderDate:    orderDate,
				CurrencyPair: p,
			})
		}
	}

	exchange.FilterOrdersByTickRange(&allOrders, getOrdersRequest.StartTicks,
		getOrdersRequest.EndTicks)
	exchange.FilterOrdersBySide(&allOrders, getOrdersRequest.OrderSide)
	return allOrders, nil
}

// SubscribeToWebsocketChannels appends to ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle subscribing
func (c *COINUT) SubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	c.Websocket.SubscribeToChannels(channels)
	return nil
}

// UnsubscribeToWebsocketChannels removes from ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle unsubscribing
func (c *COINUT) UnsubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	c.Websocket.RemoveSubscribedChannels(channels)
	return nil
}

// GetSubscriptions returns a copied list of subscriptions
func (c *COINUT) GetSubscriptions() ([]wshandler.WebsocketChannelSubscription, error) {
	return c.Websocket.GetSubscriptions(), nil
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (c *COINUT) AuthenticateWebsocket() error {
	return c.wsAuthenticate()
}
