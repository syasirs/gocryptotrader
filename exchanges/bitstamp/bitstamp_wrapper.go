package bitstamp

import (
	"errors"
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
func (b *Bitstamp) GetDefaultConfig() (*config.ExchangeConfig, error) {
	b.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = b.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = b.BaseCurrencies

	err := b.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if b.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = b.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets default for Bitstamp
func (b *Bitstamp) SetDefaults() {
	b.Name = "Bitstamp"
	b.Enabled = true
	b.Verbose = true
	b.API.CredentialsValidator.RequiresKey = true
	b.API.CredentialsValidator.RequiresSecret = true
	b.API.CredentialsValidator.RequiresClientID = true

	b.CurrencyPairs = currency.PairsManager{
		AssetTypes: asset.Items{
			asset.Spot,
		},
		UseGlobalFormat: true,
		RequestFormat: &currency.PairFormat{
			Uppercase: true,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
		},
	}

	b.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	b.Requester = request.New(b.Name,
		request.NewRateLimit(time.Minute*10, bitstampAuthRate),
		request.NewRateLimit(time.Minute*10, bitstampUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	b.API.Endpoints.URLDefault = bitstampAPIURL
	b.API.Endpoints.URL = b.API.Endpoints.URLDefault
	b.API.Endpoints.WebsocketURL = bitstampWSURL
	b.Websocket = wshandler.New()
	b.Websocket.Functionality = wshandler.WebsocketOrderbookSupported |
		wshandler.WebsocketTradeDataSupported |
		wshandler.WebsocketSubscribeSupported |
		wshandler.WebsocketUnsubscribeSupported
	b.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	b.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	b.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup sets configuration values to bitstamp
func (b *Bitstamp) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	err := b.SetupDefaults(exch)
	if err != nil {
		return err
	}

	err = b.Websocket.Setup(
		&wshandler.WebsocketSetup{
			Enabled:                          exch.Features.Enabled.Websocket,
			Verbose:                          exch.Verbose,
			AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
			WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
			DefaultURL:                       bitstampWSURL,
			ExchangeName:                     exch.Name,
			RunningURL:                       exch.API.Endpoints.WebsocketURL,
			Connector:                        b.WsConnect,
			Subscriber:                       b.Subscribe,
			UnSubscriber:                     b.Unsubscribe,
		})
	if err != nil {
		return err
	}

	b.WebsocketConn = &wshandler.WebsocketConnection{
		ExchangeName:         b.Name,
		URL:                  b.Websocket.GetWebsocketURL(),
		ProxyURL:             b.Websocket.GetProxyAddress(),
		Verbose:              b.Verbose,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	}

	b.Websocket.Orderbook.Setup(
		exch.WebsocketOrderbookBufferLimit,
		true,
		true,
		true,
		false,
		exch.Name)
	return nil
}

// Start starts the Bitstamp go routine
func (b *Bitstamp) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the Bitstamp wrapper
func (b *Bitstamp) Run() {
	if b.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			b.GetName(),
			common.IsEnabled(b.Websocket.IsEnabled()))
		b.PrintEnabledPairs()
	}

	if !b.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := b.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			b.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *Bitstamp) FetchTradablePairs(asset asset.Item) ([]string, error) {
	pairs, err := b.GetTradingPairs()
	if err != nil {
		return nil, err
	}

	var products []string
	for x := range pairs {
		if pairs[x].Trading != "Enabled" {
			continue
		}

		pair := strings.Split(pairs[x].Name, "/")
		products = append(products, pair[0]+pair[1])
	}

	return products, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *Bitstamp) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := b.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	return b.UpdatePairs(currency.NewPairsFromStrings(pairs), asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bitstamp) UpdateTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	var tickerPrice ticker.Price
	tick, err := b.GetTicker(p.String(), false)
	if err != nil {
		return tickerPrice, err
	}

	tickerPrice = ticker.Price{
		Last:        tick.Last,
		High:        tick.High,
		Low:         tick.Low,
		Bid:         tick.Bid,
		Ask:         tick.Ask,
		Volume:      tick.Volume,
		Open:        tick.Open,
		Pair:        p,
		LastUpdated: time.Unix(tick.Timestamp, 0),
	}

	err = ticker.ProcessTicker(b.GetName(), &tickerPrice, assetType)
	if err != nil {
		return tickerPrice, err
	}

	return ticker.GetTicker(b.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *Bitstamp) FetchTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	tick, err := ticker.GetTicker(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tick, nil
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bitstamp) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if (!b.AllowAuthenticatedRequest() || b.SkipAuthCheck) && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)

}

// FetchOrderbook returns the orderbook for a currency pair
func (b *Bitstamp) FetchOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	ob, err := orderbook.Get(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bitstamp) UpdateOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := b.GetOrderbook(p.String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.Item{Amount: data.Amount, Price: data.Price})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.Item{Amount: data.Amount, Price: data.Price})
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
// Bitstamp exchange
func (b *Bitstamp) GetAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.Exchange = b.GetName()
	accountBalance, err := b.GetBalance()
	if err != nil {
		return response, err
	}

	var currencies = []exchange.AccountCurrencyInfo{
		{
			CurrencyName: currency.BTC,
			TotalValue:   accountBalance.BTCAvailable,
			Hold:         accountBalance.BTCReserved,
		},
		{
			CurrencyName: currency.XRP,
			TotalValue:   accountBalance.XRPAvailable,
			Hold:         accountBalance.XRPReserved,
		},
		{
			CurrencyName: currency.USD,
			TotalValue:   accountBalance.USDAvailable,
			Hold:         accountBalance.USDReserved,
		},
		{
			CurrencyName: currency.EUR,
			TotalValue:   accountBalance.EURAvailable,
			Hold:         accountBalance.EURReserved,
		},
	}
	response.Accounts = append(response.Accounts, exchange.Account{
		Currencies: currencies,
	})

	return response, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bitstamp) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (b *Bitstamp) GetExchangeHistory(p currency.Pair, assetType asset.Item) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (b *Bitstamp) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}

	buy := s.OrderSide == order.Buy
	market := s.OrderType == order.Market
	response, err := b.PlaceOrder(s.Pair.String(),
		s.Price,
		s.Amount,
		buy,
		market)

	if response.ID > 0 {
		submitOrderResponse.OrderID = strconv.FormatInt(response.ID, 10)
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bitstamp) ModifyOrder(action *order.Modify) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bitstamp) CancelOrder(order *order.Cancellation) error {
	orderIDInt, err := strconv.ParseInt(order.OrderID, 10, 64)
	if err != nil {
		return err
	}
	_, err = b.CancelExistingOrder(orderIDInt)
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bitstamp) CancelAllOrders(_ *order.Cancellation) (order.CancelAllResponse, error) {
	success, err := b.CancelAllExistingOrders()
	if err != nil {
		return order.CancelAllResponse{}, err
	}
	if !success {
		err = errors.New("cancel all orders failed. Bitstamp provides no further information. Check order status to verify")
	}

	return order.CancelAllResponse{}, err
}

// GetOrderInfo returns information on a current open order
func (b *Bitstamp) GetOrderInfo(orderID string) (order.Detail, error) {
	var orderDetail order.Detail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *Bitstamp) GetDepositAddress(cryptocurrency currency.Code, _ string) (string, error) {
	return b.GetCryptoDepositAddress(cryptocurrency)
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bitstamp) WithdrawCryptocurrencyFunds(withdrawRequest *exchange.CryptoWithdrawRequest) (string, error) {
	resp, err := b.CryptoWithdrawal(withdrawRequest.Amount,
		withdrawRequest.Address,
		withdrawRequest.Currency.String(),
		withdrawRequest.AddressTag,
		true)
	if err != nil {
		return "", err
	}
	if len(resp.Error) != 0 {
		var details string
		for _, v := range resp.Error {
			details += strings.Join(v, "")
		}
		return "", errors.New(details)
	}

	return resp.ID, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bitstamp) WithdrawFiatFunds(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	resp, err := b.OpenBankWithdrawal(withdrawRequest.Amount,
		withdrawRequest.Currency.String(),
		withdrawRequest.BankAccountName,
		withdrawRequest.IBAN,
		withdrawRequest.SwiftCode,
		withdrawRequest.BankAddress,
		withdrawRequest.BankPostalCode,
		withdrawRequest.BankCity,
		withdrawRequest.BankCountry,
		withdrawRequest.Description,
		sepaWithdrawal)
	if err != nil {
		return "", err
	}
	if resp.Status == errStr {
		var details string
		for _, v := range resp.Reason {
			details += strings.Join(v, "")
		}
		return "", errors.New(details)
	}

	return resp.ID, nil
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bitstamp) WithdrawFiatFundsToInternationalBank(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	resp, err := b.OpenInternationalBankWithdrawal(withdrawRequest.Amount,
		withdrawRequest.Currency.String(),
		withdrawRequest.BankAccountName,
		withdrawRequest.IBAN,
		withdrawRequest.SwiftCode,
		withdrawRequest.BankAddress,
		withdrawRequest.BankPostalCode,
		withdrawRequest.BankCity,
		withdrawRequest.BankCountry,
		withdrawRequest.IntermediaryBankName,
		withdrawRequest.IntermediaryBankAddress,
		withdrawRequest.IntermediaryBankPostalCode,
		withdrawRequest.IntermediaryBankCity,
		withdrawRequest.IntermediaryBankCountry,
		withdrawRequest.WireCurrency,
		withdrawRequest.Description,
		internationalWithdrawal)
	if err != nil {
		return "", err
	}
	if resp.Status == errStr {
		var details string
		for _, v := range resp.Reason {
			details += strings.Join(v, "")
		}
		return "", errors.New(details)
	}

	return resp.ID, nil
}

// GetWebsocket returns a pointer to the exchange websocket
func (b *Bitstamp) GetWebsocket() (*wshandler.Websocket, error) {
	return b.Websocket, nil
}

// GetActiveOrders retrieves any orders that are active/open
func (b *Bitstamp) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	var orders []order.Detail
	var currPair string
	if len(req.Currencies) != 1 {
		currPair = "all"
	} else {
		currPair = req.Currencies[0].String()
	}

	resp, err := b.GetOpenOrders(currPair)
	if err != nil {
		return nil, err
	}

	for i := range resp {
		orderDate := time.Unix(resp[i].Date, 0)
		orders = append(orders, order.Detail{
			Amount:    resp[i].Amount,
			ID:        strconv.FormatInt(resp[i].ID, 10),
			Price:     resp[i].Price,
			OrderDate: orderDate,
			CurrencyPair: currency.NewPairFromStrings(resp[i].Currency[0:3],
				resp[i].Currency[len(resp[i].Currency)-3:]),
			Exchange: b.Name,
		})
	}

	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	order.FilterOrdersByCurrencies(&orders, req.Currencies)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *Bitstamp) GetOrderHistory(req *order.GetOrdersRequest) ([]order.Detail, error) {
	var currPair string
	if len(req.Currencies) == 1 {
		currPair = req.Currencies[0].String()
	}
	resp, err := b.GetUserTransactions(currPair)
	if err != nil {
		return nil, err
	}

	var orders []order.Detail
	for i := range resp {
		if resp[i].Type != 2 {
			continue
		}
		var quoteCurrency, baseCurrency currency.Code

		switch {
		case resp[i].BTC > 0:
			baseCurrency = currency.BTC
		case resp[i].XRP > 0:
			baseCurrency = currency.XRP
		default:
			log.Warnf(log.ExchangeSys,
				"no base currency found for OrderID '%d'",
				resp[i].OrderID)
		}

		switch {
		case resp[i].USD > 0:
			quoteCurrency = currency.USD
		case resp[i].EUR > 0:
			quoteCurrency = currency.EUR
		default:
			log.Warnf(log.ExchangeSys,
				"no quote currency found for orderID '%d'",
				resp[i].OrderID)
		}

		var currPair currency.Pair
		if quoteCurrency.String() != "" && baseCurrency.String() != "" {
			currPair = currency.NewPairWithDelimiter(baseCurrency.String(),
				quoteCurrency.String(),
				b.GetPairFormat(asset.Spot, false).Delimiter)
		}

		orderDate, err := time.Parse("2006-01-02 15:04:05", resp[i].Date)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order.Detail{
			ID:           strconv.FormatInt(resp[i].OrderID, 10),
			OrderDate:    orderDate,
			Exchange:     b.Name,
			CurrencyPair: currPair,
		})
	}

	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	order.FilterOrdersByCurrencies(&orders, req.Currencies)
	return orders, nil
}

// SubscribeToWebsocketChannels appends to ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle subscribing
func (b *Bitstamp) SubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	b.Websocket.SubscribeToChannels(channels)
	return nil
}

// UnsubscribeToWebsocketChannels removes from ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle unsubscribing
func (b *Bitstamp) UnsubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	b.Websocket.RemoveSubscribedChannels(channels)
	return nil
}

// GetSubscriptions returns a copied list of subscriptions
func (b *Bitstamp) GetSubscriptions() ([]wshandler.WebsocketChannelSubscription, error) {
	return b.Websocket.GetSubscriptions(), nil
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (b *Bitstamp) AuthenticateWebsocket() error {
	return common.ErrFunctionNotSupported
}
