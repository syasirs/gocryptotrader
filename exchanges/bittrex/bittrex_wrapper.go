package bittrex

import (
	"errors"
	"fmt"
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
func (b *Bittrex) GetDefaultConfig() (*config.ExchangeConfig, error) {
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

// SetDefaults method assignes the default values for Bittrex
func (b *Bittrex) SetDefaults() {
	b.Name = "Bittrex"
	b.Enabled = true
	b.Verbose = true
	b.API.CredentialsValidator.RequiresKey = true
	b.API.CredentialsValidator.RequiresSecret = true

	b.CurrencyPairs = currency.PairsManager{
		AssetTypes: asset.Items{
			asset.Spot,
		},
		UseGlobalFormat: true,
		RequestFormat: &currency.PairFormat{
			Delimiter: "-",
			Uppercase: true,
		},
		ConfigFormat: &currency.PairFormat{
			Delimiter: "-",
			Uppercase: true,
		},
	}

	b.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: false,
			RESTCapabilities: exchange.ProtocolFeatures{
				AutoPairUpdates: true,
				TickerBatching:  true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCryptoWithAPIPermission |
				exchange.NoFiatWithdrawals,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	b.Requester = request.New(b.Name,
		request.NewRateLimit(time.Second, bittrexAuthRate),
		request.NewRateLimit(time.Second, bittrexUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	b.API.Endpoints.URLDefault = bittrexAPIURL
	b.API.Endpoints.URL = b.API.Endpoints.URLDefault
}

// Setup method sets current configuration details if enabled
func (b *Bittrex) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	return b.SetupDefaults(exch)
}

// Start starts the Bittrex go routine
func (b *Bittrex) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the Bittrex wrapper
func (b *Bittrex) Run() {
	if b.Verbose {
		b.PrintEnabledPairs()
	}

	forceUpdate := false
	if !common.StringDataContains(b.GetEnabledPairs(asset.Spot).Strings(), "-") ||
		!common.StringDataContains(b.GetAvailablePairs(asset.Spot).Strings(), "-") {
		forceUpdate = true
		enabledPairs := []string{"USDT-BTC"}
		log.Warn(log.ExchangeSys, "Available pairs for Bittrex reset due to config upgrade, please enable the ones you would like again")

		err := b.UpdatePairs(currency.NewPairsFromStrings(enabledPairs), asset.Spot, true, true)
		if err != nil {
			log.Errorf(log.ExchangeSys, "%s failed to update currencies. Err: %s\n", b.Name, err)
		}
	}

	if !b.GetEnabledFeatures().AutoPairUpdates && !forceUpdate {
		return
	}

	err := b.UpdateTradablePairs(forceUpdate)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", b.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *Bittrex) FetchTradablePairs(asset asset.Item) ([]string, error) {
	markets, err := b.GetMarkets()
	if err != nil {
		return nil, err
	}

	var pairs []string
	for x := range markets.Result {
		if !markets.Result[x].IsActive || markets.Result[x].MarketName == "" {
			continue
		}
		pairs = append(pairs, markets.Result[x].MarketName)
	}

	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *Bittrex) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := b.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	return b.UpdatePairs(currency.NewPairsFromStrings(pairs), asset.Spot, false, forceUpdate)
}

// GetAccountInfo Retrieves balances for all enabled currencies for the
// Bittrex exchange
func (b *Bittrex) GetAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.Exchange = b.GetName()
	accountBalance, err := b.GetAccountBalances()
	if err != nil {
		return response, err
	}

	var currencies []exchange.AccountCurrencyInfo
	for i := 0; i < len(accountBalance.Result); i++ {
		var exchangeCurrency exchange.AccountCurrencyInfo
		exchangeCurrency.CurrencyName = currency.NewCode(accountBalance.Result[i].Currency)
		exchangeCurrency.TotalValue = accountBalance.Result[i].Balance
		exchangeCurrency.Hold = accountBalance.Result[i].Balance - accountBalance.Result[i].Available
		currencies = append(currencies, exchangeCurrency)
	}

	response.Accounts = append(response.Accounts, exchange.Account{
		Currencies: currencies,
	})

	return response, nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bittrex) UpdateTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	var tickerPrice ticker.Price
	ticks, err := b.GetMarketSummaries()
	if err != nil {
		return tickerPrice, err
	}
	pairs := b.GetEnabledPairs(assetType)
	for i := range pairs {
		for j := range ticks.Result {
			if !strings.EqualFold(ticks.Result[j].MarketName, b.FormatExchangeCurrency(pairs[i], assetType).String()) {
				continue
			}
			tickerPrice = ticker.Price{
				Last:        ticks.Result[j].Last,
				High:        ticks.Result[j].High,
				Low:         ticks.Result[j].Low,
				Bid:         ticks.Result[j].Bid,
				Ask:         ticks.Result[j].Ask,
				Volume:      ticks.Result[j].BaseVolume,
				QuoteVolume: ticks.Result[j].Volume,
				Close:       ticks.Result[j].PrevDay,
				Pair:        pairs[i],
				LastUpdated: ticks.Result[j].TimeStamp,
			}
			err = ticker.ProcessTicker(b.GetName(), &tickerPrice, assetType)
			if err != nil {
				return tickerPrice, err
			}
		}
	}

	return ticker.GetTicker(b.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *Bittrex) FetchTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	tick, err := ticker.GetTicker(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tick, nil
}

// FetchOrderbook returns the orderbook for a currency pair
func (b *Bittrex) FetchOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	ob, err := orderbook.Get(b.GetName(), p, assetType)
	if err != nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bittrex) UpdateOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	var orderBook orderbook.Base
	orderbookNew, err := b.GetOrderbook(b.FormatExchangeCurrency(p, assetType).String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Result.Buy {
		orderBook.Bids = append(orderBook.Bids,
			orderbook.Item{
				Amount: orderbookNew.Result.Buy[x].Quantity,
				Price:  orderbookNew.Result.Buy[x].Rate,
			},
		)
	}

	for x := range orderbookNew.Result.Sell {
		orderBook.Asks = append(orderBook.Asks,
			orderbook.Item{
				Amount: orderbookNew.Result.Sell[x].Quantity,
				Price:  orderbookNew.Result.Sell[x].Rate,
			},
		)
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

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bittrex) GetFundingHistory() ([]exchange.FundHistory, error) {
	var fundHistory []exchange.FundHistory
	return fundHistory, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (b *Bittrex) GetExchangeHistory(p currency.Pair, assetType asset.Item) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (b *Bittrex) SubmitOrder(order *exchange.OrderSubmission) (exchange.SubmitOrderResponse, error) {
	var submitOrderResponse exchange.SubmitOrderResponse
	if order == nil {
		return submitOrderResponse, exchange.ErrOrderSubmissionIsNil
	}

	if err := order.Validate(); err != nil {
		return submitOrderResponse, err
	}

	buy := order.OrderSide == exchange.BuyOrderSide
	var response UUID
	var err error

	if order.OrderType != exchange.LimitOrderType {
		return submitOrderResponse, errors.New("limit order not supported on exchange")
	}

	if buy {
		response, err = b.PlaceBuyLimit(order.Pair.String(), order.Amount,
			order.Price)
	} else {
		response, err = b.PlaceSellLimit(order.Pair.String(), order.Amount,
			order.Price)
	}

	if response.Result.ID != "" {
		submitOrderResponse.OrderID = response.Result.ID
	}

	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}

	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bittrex) ModifyOrder(action *exchange.ModifyOrder) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bittrex) CancelOrder(order *exchange.OrderCancellation) error {
	_, err := b.CancelExistingOrder(order.OrderID)

	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bittrex) CancelAllOrders(_ *exchange.OrderCancellation) (exchange.CancelAllOrdersResponse, error) {
	cancelAllOrdersResponse := exchange.CancelAllOrdersResponse{
		OrderStatus: make(map[string]string),
	}
	openOrders, err := b.GetOpenOrders("")
	if err != nil {
		return cancelAllOrdersResponse, err
	}

	for i := range openOrders.Result {
		_, err := b.CancelExistingOrder(openOrders.Result[i].OrderUUID)
		if err != nil {
			cancelAllOrdersResponse.OrderStatus[openOrders.Result[i].OrderUUID] = err.Error()
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (b *Bittrex) GetOrderInfo(orderID string) (exchange.OrderDetail, error) {
	var orderDetail exchange.OrderDetail
	return orderDetail, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (b *Bittrex) GetDepositAddress(cryptocurrency currency.Code, _ string) (string, error) {
	depositAddr, err := b.GetCryptoDepositAddress(cryptocurrency.String())
	if err != nil {
		return "", err
	}

	return depositAddr.Result.Address, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (b *Bittrex) WithdrawCryptocurrencyFunds(withdrawRequest *exchange.CryptoWithdrawRequest) (string, error) {
	uuid, err := b.Withdraw(withdrawRequest.Currency.String(), withdrawRequest.AddressTag, withdrawRequest.Address, withdrawRequest.Amount)
	return fmt.Sprintf("%v", uuid), err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bittrex) WithdrawFiatFunds(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bittrex) WithdrawFiatFundsToInternationalBank(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (b *Bittrex) GetWebsocket() (*wshandler.Websocket, error) {
	return nil, common.ErrNotYetImplemented
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bittrex) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !b.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)

}

// GetActiveOrders retrieves any orders that are active/open
func (b *Bittrex) GetActiveOrders(getOrdersRequest *exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var currPair string
	if len(getOrdersRequest.Currencies) == 1 {
		currPair = getOrdersRequest.Currencies[0].String()
	}

	resp, err := b.GetOpenOrders(currPair)
	if err != nil {
		return nil, err
	}

	var orders []exchange.OrderDetail
	for i := range resp.Result {
		orderDate, err := time.Parse(time.RFC3339, resp.Result[i].Opened)
		if err != nil {
			log.Warnf(log.ExchangeSys, "Exchange %v Func %v Order %v Could not parse date to unix with value of %v",
				b.Name, "GetActiveOrders", resp.Result[i].OrderUUID, resp.Result[i].Opened)
		}

		pair := currency.NewPairDelimiter(resp.Result[i].Exchange,
			b.GetPairFormat(asset.Spot, false).Delimiter)
		orderType := exchange.OrderType(strings.ToUpper(resp.Result[i].Type))

		orders = append(orders, exchange.OrderDetail{
			Amount:          resp.Result[i].Quantity,
			RemainingAmount: resp.Result[i].QuantityRemaining,
			Price:           resp.Result[i].Price,
			OrderDate:       orderDate,
			ID:              resp.Result[i].OrderUUID,
			Exchange:        b.Name,
			OrderType:       orderType,
			CurrencyPair:    pair,
		})
	}

	exchange.FilterOrdersByType(&orders, getOrdersRequest.OrderType)
	exchange.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks,
		getOrdersRequest.EndTicks)
	exchange.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *Bittrex) GetOrderHistory(getOrdersRequest *exchange.GetOrdersRequest) ([]exchange.OrderDetail, error) {
	var currPair string
	if len(getOrdersRequest.Currencies) == 1 {
		currPair = getOrdersRequest.Currencies[0].String()
	}

	resp, err := b.GetOrderHistoryForCurrency(currPair)
	if err != nil {
		return nil, err
	}

	var orders []exchange.OrderDetail
	for i := range resp.Result {
		orderDate, err := time.Parse(time.RFC3339, resp.Result[i].TimeStamp)
		if err != nil {
			log.Warnf(log.ExchangeSys, "Exchange %v Func %v Order %v Could not parse date to unix with value of %v",
				b.Name, "GetActiveOrders", resp.Result[i].OrderUUID, resp.Result[i].Opened)
		}

		pair := currency.NewPairDelimiter(resp.Result[i].Exchange,
			b.GetPairFormat(asset.Spot, false).Delimiter)
		orderType := exchange.OrderType(strings.ToUpper(resp.Result[i].Type))

		orders = append(orders, exchange.OrderDetail{
			Amount:          resp.Result[i].Quantity,
			RemainingAmount: resp.Result[i].QuantityRemaining,
			Price:           resp.Result[i].Price,
			OrderDate:       orderDate,
			ID:              resp.Result[i].OrderUUID,
			Exchange:        b.Name,
			OrderType:       orderType,
			Fee:             resp.Result[i].Commission,
			CurrencyPair:    pair,
		})
	}

	exchange.FilterOrdersByType(&orders, getOrdersRequest.OrderType)
	exchange.FilterOrdersByTickRange(&orders, getOrdersRequest.StartTicks,
		getOrdersRequest.EndTicks)
	exchange.FilterOrdersByCurrencies(&orders, getOrdersRequest.Currencies)
	return orders, nil
}

// SubscribeToWebsocketChannels appends to ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle subscribing
func (b *Bittrex) SubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	return common.ErrFunctionNotSupported
}

// UnsubscribeToWebsocketChannels removes from ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle unsubscribing
func (b *Bittrex) UnsubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	return common.ErrFunctionNotSupported
}

// GetSubscriptions returns a copied list of subscriptions
func (b *Bittrex) GetSubscriptions() ([]wshandler.WebsocketChannelSubscription, error) {
	return nil, common.ErrFunctionNotSupported
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (b *Bittrex) AuthenticateWebsocket() error {
	return common.ErrFunctionNotSupported
}
