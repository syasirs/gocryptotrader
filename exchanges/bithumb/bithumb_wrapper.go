package bithumb

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// GetDefaultConfig returns a default exchange config
func (b *Bithumb) GetDefaultConfig() (*config.ExchangeConfig, error) {
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

// SetDefaults sets the basic defaults for Bithumb
func (b *Bithumb) SetDefaults() {
	b.Name = "Bithumb"
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
			Uppercase: true,
		},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Index:     "KRW",
		},
	}

	b.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST: true,
			RESTCapabilities: protocol.Features{
				TickerBatching:      true,
				TickerFetching:      true,
				TradeFetching:       true,
				OrderbookFetching:   true,
				AutoPairUpdates:     true,
				AccountInfo:         true,
				CryptoWithdrawal:    true,
				FiatDeposit:         true,
				FiatWithdraw:        true,
				GetOrder:            true,
				CancelOrder:         true,
				SubmitOrder:         true,
				ModifyOrder:         true,
				DepositHistory:      true,
				WithdrawalHistory:   true,
				UserTradeHistory:    true,
				TradeFee:            true,
				FiatWithdrawalFee:   true,
				CryptoDepositFee:    true,
				CryptoWithdrawalFee: true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}

	b.Requester = request.New(b.Name,
		request.NewRateLimit(time.Second, bithumbAuthRate),
		request.NewRateLimit(time.Second, bithumbUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	b.API.Endpoints.URLDefault = apiURL
	b.API.Endpoints.URL = b.API.Endpoints.URLDefault
}

// Setup takes in the supplied exchange configuration details and sets params
func (b *Bithumb) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		b.SetEnabled(false)
		return nil
	}

	return b.SetupDefaults(exch)
}

// Start starts the Bithumb go routine
func (b *Bithumb) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		b.Run()
		wg.Done()
	}()
}

// Run implements the Bithumb wrapper
func (b *Bithumb) Run() {
	if b.Verbose {
		b.PrintEnabledPairs()
	}

	if !b.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := b.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys, "%s failed to update tradable pairs. Err: %s", b.Name, err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (b *Bithumb) FetchTradablePairs(asset asset.Item) ([]string, error) {
	currencies, err := b.GetTradablePairs()
	if err != nil {
		return nil, err
	}

	for x := range currencies {
		currencies[x] += "KRW"
	}

	return currencies, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (b *Bithumb) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := b.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	return b.UpdatePairs(currency.NewPairsFromStrings(pairs), asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (b *Bithumb) UpdateTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	var tickerPrice ticker.Price
	tickers, err := b.GetAllTickers()
	if err != nil {
		return tickerPrice, err
	}
	pairs := b.GetEnabledPairs(assetType)
	for i := range pairs {
		curr := pairs[i].Base.String()
		t, ok := tickers[curr]
		if !ok {
			continue
		}
		tp := ticker.Price{
			High:   t.MaxPrice,
			Low:    t.MinPrice,
			Volume: t.UnitsTraded24Hr,
			Open:   t.OpeningPrice,
			Close:  t.ClosingPrice,
			Pair:   pairs[i],
		}
		err = ticker.ProcessTicker(b.Name, &tp, assetType)
		if err != nil {
			log.Error(log.Ticker, err)
		}
	}
	return ticker.GetTicker(b.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (b *Bithumb) FetchTicker(p currency.Pair, assetType asset.Item) (ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(b.Name, p, assetType)
	if err != nil {
		return b.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (b *Bithumb) FetchOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	ob, err := orderbook.Get(b.Name, p, assetType)
	if err != nil {
		return b.UpdateOrderbook(p, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (b *Bithumb) UpdateOrderbook(p currency.Pair, assetType asset.Item) (orderbook.Base, error) {
	var orderBook orderbook.Base
	curr := p.Base.String()

	orderbookNew, err := b.GetOrderBook(curr)
	if err != nil {
		return orderBook, err
	}

	for i := range orderbookNew.Data.Bids {
		orderBook.Bids = append(orderBook.Bids,
			orderbook.Item{
				Amount: orderbookNew.Data.Bids[i].Quantity,
				Price:  orderbookNew.Data.Bids[i].Price,
			})
	}

	for i := range orderbookNew.Data.Asks {
		orderBook.Asks = append(orderBook.Asks,
			orderbook.Item{
				Amount: orderbookNew.Data.Asks[i].Quantity,
				Price:  orderbookNew.Data.Asks[i].Price,
			})
	}

	orderBook.Pair = p
	orderBook.ExchangeName = b.Name
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

	info.Exchange = b.Name
	return info, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (b *Bithumb) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetExchangeHistory returns historic trade data since exchange opening.
func (b *Bithumb) GetExchangeHistory(p currency.Pair, assetType asset.Item) ([]exchange.TradeHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
// TODO: Fill this out to support limit orders
func (b *Bithumb) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}

	var orderID string
	var err error
	if s.OrderSide == order.Buy {
		var result MarketBuy
		result, err = b.MarketBuyOrder(s.Pair.Base.String(), s.Amount)
		orderID = result.OrderID
	} else if s.OrderSide == order.Sell {
		var result MarketSell
		result, err = b.MarketSellOrder(s.Pair.Base.String(), s.Amount)
		orderID = result.OrderID
	}

	if orderID != "" {
		submitOrderResponse.OrderID = orderID
	}
	if err == nil {
		submitOrderResponse.IsOrderPlaced = true
	}
	return submitOrderResponse, err
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (b *Bithumb) ModifyOrder(action *order.Modify) (string, error) {
	order, err := b.ModifyTrade(action.OrderID,
		action.CurrencyPair.Base.String(),
		action.Side.Lower(),
		action.Amount,
		int64(action.Price))

	if err != nil {
		return "", err
	}

	return order.Data[0].ContID, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (b *Bithumb) CancelOrder(order *order.Cancel) error {
	_, err := b.CancelTrade(order.Side.String(),
		order.OrderID,
		order.CurrencyPair.Base.String())
	return err
}

// CancelAllOrders cancels all orders associated with a currency pair
func (b *Bithumb) CancelAllOrders(orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	cancelAllOrdersResponse := order.CancelAllResponse{
		Status: make(map[string]string),
	}

	var allOrders []OrderData
	currs := b.GetEnabledPairs(asset.Spot)
	for i := range currs {
		orders, err := b.GetOrders("",
			orderCancellation.Side.String(),
			"100",
			"",
			currs[i].Base.String())
		if err != nil {
			return cancelAllOrdersResponse, err
		}
		allOrders = append(allOrders, orders.Data...)
	}

	for i := range allOrders {
		_, err := b.CancelTrade(orderCancellation.Side.String(),
			allOrders[i].OrderID,
			orderCancellation.CurrencyPair.Base.String())
		if err != nil {
			cancelAllOrdersResponse.Status[allOrders[i].OrderID] = err.Error()
		}
	}

	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns information on a current open order
func (b *Bithumb) GetOrderInfo(orderID string) (order.Detail, error) {
	var orderDetail order.Detail
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
func (b *Bithumb) WithdrawCryptocurrencyFunds(withdrawRequest *exchange.CryptoWithdrawRequest) (string, error) {
	_, err := b.WithdrawCrypto(withdrawRequest.Address,
		withdrawRequest.AddressTag,
		withdrawRequest.Currency.String(),
		withdrawRequest.Amount)
	return "", err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (b *Bithumb) WithdrawFiatFunds(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	if math.Mod(withdrawRequest.Amount, 1) != 0 {
		return "", errors.New("currency KRW does not support decimal places")
	}
	if withdrawRequest.Currency != currency.KRW {
		return "", errors.New("only KRW is supported")
	}
	bankDetails := fmt.Sprintf("%v_%v", withdrawRequest.BankCode, withdrawRequest.BankName)
	resp, err := b.RequestKRWWithdraw(bankDetails, withdrawRequest.BankAccountNumber, int64(withdrawRequest.Amount))
	if err != nil {
		return "", err
	}
	if resp.Status != "0000" {
		return "", errors.New(resp.Message)
	}

	return resp.Message, nil
}

// WithdrawFiatFundsToInternationalBank is not supported as Bithumb only withdraws KRW to South Korean banks
func (b *Bithumb) WithdrawFiatFundsToInternationalBank(withdrawRequest *exchange.FiatWithdrawRequest) (string, error) {
	return "", common.ErrFunctionNotSupported
}

// GetWebsocket returns a pointer to the exchange websocket
func (b *Bithumb) GetWebsocket() (*wshandler.Websocket, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (b *Bithumb) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	if !b.AllowAuthenticatedRequest() && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return b.GetFee(feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (b *Bithumb) GetActiveOrders(req *order.GetOrdersRequest) ([]order.Detail, error) {
	var orders []order.Detail
	resp, err := b.GetOrders("", "", "1000", "", "")
	if err != nil {
		return nil, err
	}

	for i := range resp.Data {
		if resp.Data[i].Status != "placed" {
			continue
		}

		orderDate := time.Unix(resp.Data[i].OrderDate, 0)
		orderDetail := order.Detail{
			Amount:          resp.Data[i].Units,
			Exchange:        b.Name,
			ID:              resp.Data[i].OrderID,
			OrderDate:       orderDate,
			Price:           resp.Data[i].Price,
			RemainingAmount: resp.Data[i].UnitsRemaining,
			Status:          order.Active,
			CurrencyPair: currency.NewPairWithDelimiter(resp.Data[i].OrderCurrency,
				resp.Data[i].PaymentCurrency,
				b.GetPairFormat(asset.Spot, false).Delimiter),
		}

		if resp.Data[i].Type == "bid" {
			orderDetail.OrderSide = order.Buy
		} else if resp.Data[i].Type == "ask" {
			orderDetail.OrderSide = order.Sell
		}

		orders = append(orders, orderDetail)
	}

	order.FilterOrdersBySide(&orders, req.OrderSide)
	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	order.FilterOrdersByCurrencies(&orders, req.Currencies)
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (b *Bithumb) GetOrderHistory(req *order.GetOrdersRequest) ([]order.Detail, error) {
	var orders []order.Detail
	resp, err := b.GetOrders("", "", "1000", "", "")
	if err != nil {
		return nil, err
	}

	for i := range resp.Data {
		if resp.Data[i].Status == "placed" {
			continue
		}

		orderDate := time.Unix(resp.Data[i].OrderDate, 0)
		orderDetail := order.Detail{
			Amount:          resp.Data[i].Units,
			Exchange:        b.Name,
			ID:              resp.Data[i].OrderID,
			OrderDate:       orderDate,
			Price:           resp.Data[i].Price,
			RemainingAmount: resp.Data[i].UnitsRemaining,
			CurrencyPair: currency.NewPairWithDelimiter(resp.Data[i].OrderCurrency,
				resp.Data[i].PaymentCurrency,
				b.GetPairFormat(asset.Spot, false).Delimiter),
		}

		if resp.Data[i].Type == "bid" {
			orderDetail.OrderSide = order.Buy
		} else if resp.Data[i].Type == "ask" {
			orderDetail.OrderSide = order.Sell
		}

		orders = append(orders, orderDetail)
	}

	order.FilterOrdersBySide(&orders, req.OrderSide)
	order.FilterOrdersByTickRange(&orders, req.StartTicks, req.EndTicks)
	order.FilterOrdersByCurrencies(&orders, req.Currencies)
	return orders, nil
}

// SubscribeToWebsocketChannels appends to ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle subscribing
func (b *Bithumb) SubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	return common.ErrFunctionNotSupported
}

// UnsubscribeToWebsocketChannels removes from ChannelsToSubscribe
// which lets websocket.manageSubscriptions handle unsubscribing
func (b *Bithumb) UnsubscribeToWebsocketChannels(channels []wshandler.WebsocketChannelSubscription) error {
	return common.ErrFunctionNotSupported
}

// GetSubscriptions returns a copied list of subscriptions
func (b *Bithumb) GetSubscriptions() ([]wshandler.WebsocketChannelSubscription, error) {
	return nil, common.ErrFunctionNotSupported
}

// AuthenticateWebsocket sends an authentication message to the websocket
func (b *Bithumb) AuthenticateWebsocket() error {
	return common.ErrFunctionNotSupported
}
