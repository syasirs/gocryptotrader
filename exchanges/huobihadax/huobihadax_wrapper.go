package huobihadax

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
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// GetDefaultConfig returns a default exchange config
func (h *HUOBIHADAX) GetDefaultConfig() (*config.ExchangeConfig, error) {
	h.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = h.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = h.BaseCurrencies

	err := h.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if h.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = h.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets default values for the exchange
func (h *HUOBIHADAX) SetDefaults() {
	h.Name = "HuobiHadax"
	h.Enabled = true
	h.Verbose = true
	h.API.CredentialsValidator.RequiresKey = true
	h.API.CredentialsValidator.RequiresSecret = true

	h.CurrencyPairs = currency.PairsManager{
		AssetTypes: asset.Items{
			asset.Spot,
		},

		UseGlobalFormat: true,
		RequestFormat: &currency.PairFormat{
			Uppercase: false,
		},
		ConfigFormat: &currency.PairFormat{
			Delimiter: "-",
			Uppercase: true,
		},
	}

	h.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
				TickerBatching:  true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCryptoWithSetup |
				exchange.NoFiatWithdrawals,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	h.Requester = request.New(h.Name,
		request.NewRateLimit(time.Second*10, huobihadaxAuthRate),
		request.NewRateLimit(time.Second*10, huobihadaxUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	h.API.Endpoints.URLDefault = huobihadaxAPIURL
	h.API.Endpoints.URL = h.API.Endpoints.URLDefault
	h.Websocket = wshandler.New()
	h.Websocket.Functionality = wshandler.WebsocketKlineSupported |
		wshandler.WebsocketTradeDataSupported |
		wshandler.WebsocketOrderbookSupported |
		wshandler.WebsocketSubscribeSupported |
		wshandler.WebsocketUnsubscribeSupported |
		wshandler.WebsocketAuthenticatedEndpointsSupported |
		wshandler.WebsocketAccountDataSupported |
		wshandler.WebsocketMessageCorrelationSupported
	h.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	h.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	h.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup sets user configuration
func (h *HUOBIHADAX) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		h.SetEnabled(false)
		return nil
	}

	err := h.SetupDefaults(exch)
	if err != nil {
		return err
	}

	err = h.Websocket.Setup(h.WsConnect,
		h.Subscribe,
		h.Unsubscribe,
		exch.Name,
		exch.Features.Enabled.Websocket,
		exch.Verbose,
		HuobiHadaxSocketIOAddress,
		exch.API.Endpoints.WebsocketURL,
		exch.API.AuthenticatedWebsocketSupport)
	if err != nil {
		return err
	}

	h.WebsocketConn = &wshandler.WebsocketConnection{
		ExchangeName:         h.Name,
		URL:                  HuobiHadaxSocketIOAddress,
		ProxyURL:             h.Websocket.GetProxyAddress(),
		Verbose:              h.Verbose,
		RateLimit:            rateLimit,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	}
	h.AuthenticatedWebsocketConn = &wshandler.WebsocketConnection{
		ExchangeName:         h.Name,
		URL:                  wsAccountsOrdersURL,
		ProxyURL:             h.Websocket.GetProxyAddress(),
		Verbose:              h.Verbose,
		RateLimit:            rateLimit,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	}

	h.Websocket.Orderbook.Setup(
		exch.WebsocketOrderbookBufferLimit,
		false,
		false,
		false,
		false,
		exch.Name)
	return nil
}

// Start starts the HUOBIHADAX go routine
func (h *HUOBIHADAX) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		h.Run()
		wg.Done()
	}()
}

// Run implements the HUOBIHADAX wrapper
func (h *HUOBIHADAX) Run() {
	if h.Verbose {
		h.PrintEnabledPairs()
	}

	if !h.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := h.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", h.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (h *HUOBIHADAX) FetchTradablePairs(asset asset.Item) ([]string, error) {
	symbols, err := h.GetSymbols()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for x := range symbols {
		pairs = append(pairs, fmt.Sprintf("%v%v%v", symbols[x].BaseCurrency,
			h.GetPairFormat(asset, false).Delimiter, symbols[x].QuoteCurrency))
	}

	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (h *HUOBIHADAX) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := h.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	return h.UpdatePairs(currency.NewPairsFromStrings(pairs), asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (h *HUOBIHADAX) UpdateTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	var tickerPrice ticker.Price
	tickers, err := h.GetTickers()
	if err != nil {
		return tickerPrice, err
	}
	pairs := h.GetEnabledPairs(assetType)
	for i := range pairs {
		for j := range tickers.Data {
			pairFmt := h.FormatExchangeCurrency(pairs[i], assetType).String()
			if pairFmt != tickers.Data[j].Symbol {
				continue
			}
			tickerPrice := ticker.Price{
				High:   tickers.Data[j].High,
				Low:    tickers.Data[j].Low,
				Volume: tickers.Data[j].Volume,
				Open:   tickers.Data[j].Open,
				Close:  tickers.Data[j].Close,
				Pair:   pairs[i],
			}
			err = ticker.ProcessTicker(h.GetName(), &tickerPrice, assetType)
			if err != nil {
				log.Error(log.Ticker, err)
			}
		}
	}

	return ticker.GetTicker(h.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (h *HUOBIHADAX) FetchTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(h.GetName(), p, assetType)
	if err != nil {
		return h.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (h *HUOBIHADAX) FetchOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	ob, err := orderbook.Get(h.GetName(), p, assetType)
	if err != nil {
		return h.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (h *HUOBIHADAX) UpdateOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := h.GetDepth(h.FormatExchangeCurrency(p, assetType).String(), "step1")
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: data[1], Price: data[0]})
	}

	orderBook.Pair = p
	orderBook.ExchangeName = h.GetName()
	orderBook.AssetType = assetType

	err = orderBook.Process()
	if err != nil {
		return orderBook, err
	}

	return orderbook.Get(h.Name, p, assetType)
}

// GetAccountID returns the account info
func (h *HUOBIHADAX) GetAccountID() ([]Account, error) {
	acc, err := h.GetAccounts()
	if err != nil {
		return nil, err
	}

	if len(acc) < 1 {
		return nil, errors.New("no account returned")
	}

	return acc, nil
}

// GetAccountInfo retrieves balances for all enabled currencies for the
// HUOBIHADAX exchange
func (h *HUOBIHADAX) GetAccountInfo() (exchange.AccountInfo, error) {
	var info exchange.AccountInfo
	info.Exchange = h.GetName()

	accounts, err := h.GetAccountID()
	if err != nil {
		return info, err
	}

	for i := range accounts {
		var acc exchange.Account
		acc.ID = strconv.FormatInt(accounts[i].ID, 10)
		balances, err := h.GetAccountBalance(acc.ID)
		if err != nil {
			return info, err
		}

		var currencyDetails []exchange.AccountCurrencyInfo
		for j := range balances {
			var frozen bool
			if balances[j].Type == "frozen" {
				frozen = true
			}

			var updated bool
			for i := range currencyDetails {
				if currencyDetails[i].CurrencyName == currency.NewCode(balances[j].Currency) {
					if frozen {
						currencyDetails[i].Hold = balances[j].Balance
					} else {
						currencyDetails[i].TotalValue = balances[j].Balance
					}
					updated = true
				}
			}

			if updated {
				continue
			}

			if frozen {
				currencyDetails = append(currencyDetails,
					exchange.AccountCurrencyInfo{
						CurrencyName: currency.NewCode(balances[j].Currency),
						Hold:         balances[j].Balance,
					})
			} else {
				currencyDetails = append(currencyDetails,
					exchange.AccountCurrencyInfo{
						CurrencyName: currency.NewCode(balances[j].Currency),
						TotalValue:   balances[j].Balance,
					})
			}
		}

		acc.Currencies = currencyDetails
		info.Accounts = append(info.Accounts, acc)
	}

	return info, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (h *HUOBIHADAX) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (h *HUOBIHADAX) GetExchangeHistory(p currency.Pair, assetType asset.Item) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (h *HUOBIHADAX) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}

	accountID, err := strconv.ParseInt(s.ClientID, 10, 64)
	if err != nil {
		return submitOrderResponse, err
	}

	var formattedType SpotNewOrderRequestParamsType
	var params = SpotNewOrderRequestParams{
		Amount:    s.Amount,
		Source:    "api",
		Symbol:    strings.ToLower(s.Pair.String()),
		AccountID: int(accountID),
	}

	switch {
	case s.OrderSide == order.Buy && s.OrderType == order.Market:
		formattedType = SpotNewOrderRequestTypeBuyMarket
	case s.OrderSide == order.Sell && s.OrderType == order.Market:
		formattedType = SpotNewOrderRequestTypeSellMarket
	case s.OrderSide == order.Buy && s.OrderType == order.Limit:
		formattedType = SpotNewOrderRequestTypeBuyLimit
		params.Price = s.Price
	case s.OrderSide == order.Sell && s.OrderType == order.Limit:
		formattedType = SpotNewOrderRequestTypeSellLimit
		params.Price = s.Price
	}

	params.Type = formattedType
	response, err := h.SpotNewOrder(params)
	if response > 0 {
		submitOrderResponse.OrderID = strconv.FormatInt(response, 10)
	}
	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}
	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (h *HUOBIHADAX) ModifyOrder(action *order.Modify) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (h *HUOBIHADAX) CancelOrder(order *order.Cancellation) error {
	orderIDInt, err := strconv.ParseInt(order.OrderID, 10, 64)
	if err != nil {
		return err
	}

	_, err = h.CancelExistingOrder(orderIDInt)
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (h *HUOBIHADAX) CancelAllOrders(orderCancellation *order.Cancellation) (order.CancelAllResponse, error) {
	var cancelAllOrdersResponse order.CancelAllResponse
	for _, currency := range h.GetEnabledPairs(asset.Spot) {
		resp, err := h.CancelOpenOrdersBatch(orderCancellation.AccountID,
			h.FormatExchangeCurrency(currency, asset.Spot).String())
		if err != nil {
			return cancelAllOrdersResponse, err
		}

		if resp.Data.FailedCount > 0 {
			return cancelAllOrdersResponse,
				fmt.Errorf("%v orders failed to cancel",
					resp.Data.FailedCount)
		}

		if resp.Status == "error" {
			return cancelAllOrdersResponse, errors.New(resp.ErrorMessage)
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (h *HUOBIHADAX) GetOrderInfo(orderID string) (order.Detail, error) {
	var orderDetail order.Detail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (h *HUOBIHADAX) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (h *HUOBIHADAX) WithdrawCryptocurrencyFunds(withdrawRequest *exchange.CryptoWithdrawRequest) (string, error) {
	resp, err := h.Withdraw(withdrawRequest.Currency,
		withdrawRequest.Address,
		withdrawRequest.AddressTag,
		withdrawRequest.Amount,
		withdrawRequest.FeeAmount)
	return strconv.FormatInt(resp, 10), err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (h *HUOBIHADAX) WithdrawFiatFunds(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (h *HUOBIHADAX) WithdrawFiatFundsToInternationalBank(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (h *HUOBIHADAX) GetWebsocket() (*wshandler.Websocket, error) {
	return h.Websocket, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (h *HUOBIHADAX) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !h.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return h.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (h *HUOBIHADAX) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if len(req.Currencies) == 0 {
		return nil, errors.New("currency must be supplied")
	}

	side := ""
	if req.OrderSide == order.AnySide || req.OrderSide == "" {
		side = ""
	} else if req.OrderSide == order.Sell {
		side = strings.ToLower(string(req.OrderSide))
	}

	var allOrders []OrderInfo
	for _, currency := range req.Currencies {
		resp, err := h.GetOpenOrders(h.API.Credentials.ClientID,
			currency.Lower().String(),
			side,
			500)
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, resp...)
	}

	var orders []order.Detail
	for i := range allOrders {
		symbol := currency.NewPairDelimiter(allOrders[i].Symbol,
			h.GetPairFormat(asset.Spot, false).Delimiter)
		orderDate := time.Unix(0, allOrders[i].CreatedAt*int64(time.Millisecond))

		orders = append(orders, order.Detail{
			ID:              strconv.FormatInt(int64(allOrders[i].ID), 10),
			Exchange:        h.Name,
			Amount:          allOrders[i].Amount,
			Price:           allOrders[i].Price,
			OrderDate:       orderDate,
			ExecutedAmount:  allOrders[i].FilledAmount,
			RemainingAmount: (allOrders[i].Amount - allOrders[i].FilledAmount),
			CurrencyPair:    symbol,
		})
	}
	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (h *HUOBIHADAX) GetOrderHistory(req *order.GetOrdersRequest) ([]order.Detail, error) {
	if len(req.Currencies) == 0 {
		return nil, errors.New("currency must be supplied")
	}

	states := "partial-canceled,filled,canceled"
	var allOrders []OrderInfo
	for _, currency := range req.Currencies {
		resp, err := h.GetOrders(currency.Lower().String(),
			"",
			"",
			"",
			states,
			"",
			"",
			"")
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, resp...)
	}

	var orders []order.Detail
	for i := range allOrders {
		symbol := currency.NewPairDelimiter(allOrders[i].Symbol,
			h.GetPairFormat(asset.Spot, false).Delimiter)
		orderDate := time.Unix(0, allOrders[i].CreatedAt*int64(time.Millisecond))

		orders = append(orders, order.Detail{
			ID:           strconv.FormatInt(int64(allOrders[i].ID), 10),
			Exchange:     h.Name,
			Amount:       allOrders[i].Amount,
			Price:        allOrders[i].Price,
			OrderDate:    orderDate,
			CurrencyPair: symbol,
		})
	}

	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	return orders, nil
}

// SubscribeToWebsocketChannels appends to ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle subscribing
func (h *HUOBIHADAX) SubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	h.Websocket.SubscribeToChannels(channels)
	return nil
}

// UnsubscribeToWebsocketChannels removes from ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle unsubscribing
func (h *HUOBIHADAX) UnsubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	h.Websocket.RemoveSubscribedChannels(channels)
	return nil
}

// GetSubscriptions returns a copied list of subscriptions
func (h *HUOBIHADAX) GetSubscriptions() ([]wshandler.WebsocketChannelSubscription, error) {
	return h.Websocket.GetSubscriptions(), nil
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (h *HUOBIHADAX) AuthenticateWebsocket() error {
	return h.wsLogin()
}
