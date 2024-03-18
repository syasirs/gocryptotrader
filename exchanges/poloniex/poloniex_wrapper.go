package poloniex

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fundingrate"
	"github.com/thrasher-corp/gocryptotrader/exchanges/futures"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream/buffer"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// GetDefaultConfig returns a default exchange config
func (p *Poloniex) GetDefaultConfig(ctx context.Context) (*config.Exchange, error) {
	p.SetDefaults()
	exchCfg, err := p.GetStandardConfig()
	if err != nil {
		return nil, err
	}

	err = p.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}

	if p.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err = p.UpdateTradablePairs(ctx, true)
		if err != nil {
			return nil, err
		}
	}

	return exchCfg, nil
}

// SetDefaults sets default settings for poloniex
func (p *Poloniex) SetDefaults() {
	p.Name = "Poloniex"
	p.Enabled = true
	p.Verbose = true
	p.API.CredentialsValidator.RequiresKey = true
	p.API.CredentialsValidator.RequiresSecret = true

	err := p.StoreAssetPairFormat(asset.Spot, currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true, Delimiter: currency.UnderscoreDelimiter},
		ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: currency.UnderscoreDelimiter},
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	err = p.StoreAssetPairFormat(asset.Futures, currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true},
		ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: currency.UnderscoreDelimiter},
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	p.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerBatching:        true,
				TickerFetching:        true,
				KlineFetching:         true,
				TradeFetching:         true,
				OrderbookFetching:     true,
				AutoPairUpdates:       true,
				AccountInfo:           true,
				GetOrder:              true,
				GetOrders:             true,
				CancelOrder:           true,
				CancelOrders:          true,
				SubmitOrder:           true,
				DepositHistory:        true,
				WithdrawalHistory:     true,
				UserTradeHistory:      true,
				CryptoDeposit:         true,
				CryptoWithdrawal:      true,
				TradeFee:              true,
				CryptoWithdrawalFee:   true,
				MultiChainDeposits:    true,
				MultiChainWithdrawals: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:         true,
				TradeFetching:          true,
				OrderbookFetching:      true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				GetOrders:              true,
				GetOrder:               true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCryptoWithAPIPermission |
				exchange.NoFiatWithdrawals,
			Kline: kline.ExchangeCapabilitiesSupported{
				Intervals: true,
			},
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: kline.DeployExchangeIntervals(
					kline.IntervalCapacity{Interval: kline.OneMin},
					kline.IntervalCapacity{Interval: kline.FiveMin},
					kline.IntervalCapacity{Interval: kline.TenMin},
					kline.IntervalCapacity{Interval: kline.FifteenMin},
					kline.IntervalCapacity{Interval: kline.ThirtyMin},
					kline.IntervalCapacity{Interval: kline.OneHour},
					kline.IntervalCapacity{Interval: kline.TwoHour},
					kline.IntervalCapacity{Interval: kline.FourHour},
					kline.IntervalCapacity{Interval: kline.SixHour},
					kline.IntervalCapacity{Interval: kline.TwelveHour},
					kline.IntervalCapacity{Interval: kline.OneDay},
					kline.IntervalCapacity{Interval: kline.ThreeDay},
					kline.IntervalCapacity{Interval: kline.OneWeek},
					kline.IntervalCapacity{Interval: kline.OneMonth},
				),
				GlobalResultLimit: 500,
			},
		},
	}

	p.Requester, err = request.New(p.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	p.API.Endpoints = p.NewEndpoints()
	err = p.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      poloniexAPIURL,
		exchange.WebsocketSpot: poloniexWebsocketAddress,
		exchange.RestFutures:   poloniexFuturesAPIURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	p.Websocket = stream.NewWebsocket()
	p.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	p.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	p.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup sets user exchange configuration settings
func (p *Poloniex) Setup(exch *config.Exchange) error {
	err := exch.Validate()
	if err != nil {
		return err
	}
	if !exch.Enabled {
		p.SetEnabled(false)
		return nil
	}
	err = p.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsRunningURL, err := p.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = p.Websocket.Setup(&stream.WebsocketSetup{
		ExchangeConfig:        exch,
		DefaultURL:            poloniexWebsocketAddress,
		RunningURL:            wsRunningURL,
		Connector:             p.WsConnect,
		Subscriber:            p.Subscribe,
		Unsubscriber:          p.Unsubscribe,
		GenerateSubscriptions: p.GenerateDefaultSubscriptions,
		Features:              &p.Features.Supports.WebsocketCapabilities,
		OrderbookBufferConfig: buffer.Config{
			SortBuffer:            true,
			SortBufferByUpdateIDs: true,
		},
	})
	if err != nil {
		return err
	}

	err = p.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
		URL:                  poloniexWebsocketAddress,
		RateLimit:            500,
	})
	if err != nil {
		return err
	}
	return p.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
		URL:                  poloniexPrivateWebsocketAddress,
		RateLimit:            500,
		Authenticated:        true,
	})
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (p *Poloniex) FetchTradablePairs(ctx context.Context, _ asset.Item) (currency.Pairs, error) {
	// TODO: Upgrade to new API version for fetching operational pairs.
	resp, err := p.GetSymbolInformation(ctx, currency.EMPTYPAIR)
	if err != nil {
		return nil, err
	}

	pairs := make([]currency.Pair, 0, len(resp))
	for x := range resp {
		if strings.EqualFold(resp[x].State, "PAUSE") {
			continue
		}
		var pair currency.Pair
		pair, err = currency.NewPairFromString(resp[x].Symbol)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (p *Poloniex) UpdateTradablePairs(ctx context.Context, forceUpgrade bool) error {
	pairs, err := p.FetchTradablePairs(ctx, asset.Spot)
	if err != nil {
		return err
	}
	err = p.UpdatePairs(pairs, asset.Spot, false, forceUpgrade)
	if err != nil {
		return err
	}
	return p.EnsureOnePairEnabled()
}

// UpdateTickers updates the ticker for all currency pairs of a given asset type
func (p *Poloniex) UpdateTickers(ctx context.Context, a asset.Item) error {
	ticks, err := p.GetTickers(ctx)
	if err != nil {
		return err
	}

	enabledPairs, err := p.GetEnabledPairs(a)
	if err != nil {
		return err
	}

	for i := range ticks {
		pair, err := currency.NewPairFromString(ticks[i].Symbol)
		if err != nil {
			return err
		}

		if !enabledPairs.Contains(pair, true) {
			continue
		}

		err = ticker.ProcessTicker(&ticker.Price{
			AssetType:    a,
			Pair:         pair,
			ExchangeName: p.Name,
			Last:         ticks[i].MarkPrice.Float64(),
			Low:          ticks[i].Low.Float64(),
			Ask:          ticks[i].Ask.Float64(),
			Bid:          ticks[i].Bid.Float64(),
			High:         ticks[i].High.Float64(),
			QuoteVolume:  ticks[i].Amount.Float64(),
			Volume:       ticks[i].Quantity.Float64(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (p *Poloniex) UpdateTicker(ctx context.Context, currencyPair currency.Pair, a asset.Item) (*ticker.Price, error) {
	if err := p.UpdateTickers(ctx, a); err != nil {
		return nil, err
	}
	return ticker.GetTicker(p.Name, currencyPair, a)
}

// FetchTicker returns the ticker for a currency pair
func (p *Poloniex) FetchTicker(ctx context.Context, currencyPair currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(p.Name, currencyPair, assetType)
	if err != nil {
		return p.UpdateTicker(ctx, currencyPair, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (p *Poloniex) FetchOrderbook(ctx context.Context, currencyPair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(p.Name, currencyPair, assetType)
	if err != nil {
		return p.UpdateOrderbook(ctx, currencyPair, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (p *Poloniex) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	if pair.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	err := p.CurrencyPairs.IsAssetEnabled(assetType)
	if err != nil {
		return nil, err
	}
	pair, err = p.FormatExchangeCurrency(pair, assetType)
	if err != nil {
		return nil, err
	}
	orderbookNew, err := p.GetOrderbook(ctx, pair, 0, 0)
	if err != nil {
		return nil, err
	}

	book := &orderbook.Base{
		Exchange:        p.Name,
		Pair:            pair,
		Asset:           assetType,
		VerifyOrderbook: p.CanVerifyOrderbook,
	}

	book.Bids = make(orderbook.Items, len(orderbookNew.Bids)/2)
	for y := range book.Bids {
		book.Bids[y].Price = orderbookNew.Bids[y*2].Float64()
		book.Bids[y].Amount = orderbookNew.Bids[y*2+1].Float64()
	}

	book.Asks = make(orderbook.Items, len(orderbookNew.Asks)/2)
	for y := range book.Asks {
		book.Asks[y].Price = orderbookNew.Asks[y*2].Float64()
		book.Asks[y].Amount = orderbookNew.Asks[y*2+1].Float64()
	}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(p.Name, pair, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies for the
// Poloniex exchange
func (p *Poloniex) UpdateAccountInfo(ctx context.Context, _ asset.Item) (account.Holdings, error) {
	var response account.Holdings
	accountBalance, err := p.GetSubAccountBalances(ctx)
	if err != nil {
		return response, err
	}

	subAccounts := make([]account.SubAccount, len(accountBalance))
	for i := range accountBalance {
		subAccount := account.SubAccount{
			ID:        accountBalance[i].AccountID,
			AssetType: stringToAccountType(accountBalance[i].AccountType),
		}
		currencyBalances := make([]account.Balance, len(accountBalance[i].Balances))
		for x := range accountBalance[i].Balances {
			currencyBalances[x] = account.Balance{
				Currency:               currency.NewCode(accountBalance[i].Balances[x].Currency),
				Total:                  accountBalance[i].Balances[x].AvailableBalance.Float64(),
				Hold:                   accountBalance[i].Balances[x].Hold.Float64(),
				Free:                   accountBalance[i].Balances[x].Available.Float64(),
				AvailableWithoutBorrow: accountBalance[i].Balances[x].AvailableBalance.Float64(),
			}
		}
		subAccounts[i] = subAccount
	}
	response = account.Holdings{
		Exchange: p.Name,
		Accounts: subAccounts,
	}
	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	err = account.Process(&response, creds)
	if err != nil {
		return account.Holdings{}, err
	}
	return response, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (p *Poloniex) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	acc, err := account.GetHoldings(p.Name, creds, assetType)
	if err != nil {
		return p.UpdateAccountInfo(ctx, assetType)
	}
	return acc, nil
}

// GetAccountFundingHistory returns funding history, deposits and
// withdrawals
func (p *Poloniex) GetAccountFundingHistory(ctx context.Context) ([]exchange.FundingHistory, error) {
	end := time.Now()
	walletActivity, err := p.WalletActivity(ctx, end.Add(-time.Hour*24*365), end, "")
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.FundingHistory, len(walletActivity.Deposits))
	for i := range walletActivity.Deposits {
		resp[i] = exchange.FundingHistory{
			ExchangeName:    p.Name,
			Status:          walletActivity.Deposits[i].Status,
			Timestamp:       walletActivity.Deposits[i].Timestamp.Time(),
			Currency:        walletActivity.Deposits[i].Currency,
			Amount:          walletActivity.Deposits[i].Amount.Float64(),
			CryptoToAddress: walletActivity.Deposits[i].Address,
			CryptoTxID:      walletActivity.Deposits[i].TransactionID,
		}
	}
	for i := range walletActivity.Withdrawals {
		resp[i] = exchange.FundingHistory{
			ExchangeName:    p.Name,
			Status:          walletActivity.Withdrawals[i].Status,
			Timestamp:       walletActivity.Withdrawals[i].Timestamp.Time(),
			Currency:        walletActivity.Withdrawals[i].Currency,
			Amount:          walletActivity.Withdrawals[i].Amount.Float64(),
			Fee:             walletActivity.Withdrawals[i].Fee.Float64(),
			CryptoToAddress: walletActivity.Withdrawals[i].Address,
			CryptoTxID:      walletActivity.Withdrawals[i].TransactionID,
		}
	}
	return resp, nil
}

// GetWithdrawalsHistory returns previous withdrawals data
func (p *Poloniex) GetWithdrawalsHistory(ctx context.Context, c currency.Code, _ asset.Item) ([]exchange.WithdrawalHistory, error) {
	end := time.Now()
	withdrawals, err := p.WalletActivity(ctx, end.Add(-time.Hour*24*365), end, "withdrawals")
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.WithdrawalHistory, 0, len(withdrawals.Withdrawals))
	for i := range withdrawals.Withdrawals {
		if !c.Equal(currency.NewCode(withdrawals.Withdrawals[i].Currency)) {
			continue
		}
		resp[i] = exchange.WithdrawalHistory{
			Status:          withdrawals.Withdrawals[i].Status,
			Timestamp:       withdrawals.Withdrawals[i].Timestamp.Time(),
			Currency:        withdrawals.Withdrawals[i].Currency,
			Amount:          withdrawals.Withdrawals[i].Amount.Float64(),
			Fee:             withdrawals.Withdrawals[i].Fee.Float64(),
			CryptoToAddress: withdrawals.Withdrawals[i].Address,
			CryptoTxID:      withdrawals.Withdrawals[i].TransactionID,
		}
	}
	return resp, nil
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (p *Poloniex) GetRecentTrades(ctx context.Context, pair currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return p.GetHistoricTrades(ctx, pair, assetType, time.Now().Add(-time.Minute*15), time.Now())
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (p *Poloniex) GetHistoricTrades(ctx context.Context, pair currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	if err := common.StartEndTimeCheck(timestampStart, timestampEnd); err != nil {
		return nil, fmt.Errorf("invalid time range supplied. Start: %v End %v %w", timestampStart, timestampEnd, err)
	}
	var err error
	pair, err = p.FormatExchangeCurrency(pair, assetType)
	if err != nil {
		return nil, err
	}

	var resp []trade.Data
	ts := timestampStart
allTrades:
	for {
		var tradeData []TradeHistoryItem
		tradeData, err = p.GetTradeHistory(ctx, currency.Pairs{pair}, "", 0, 0, ts, timestampEnd)
		if err != nil {
			return nil, err
		}
		for i := range tradeData {
			var tt time.Time
			if (tradeData[i].CreateTime.Time().Before(timestampStart) && !timestampStart.IsZero()) || (tradeData[i].CreateTime.Time().After(timestampEnd) && !timestampEnd.IsZero()) {
				break allTrades
			}
			var side order.Side
			side, err = order.StringToOrderSide(tradeData[i].Type)
			if err != nil {
				return nil, err
			}
			resp = append(resp, trade.Data{
				Exchange:     p.Name,
				CurrencyPair: pair,
				AssetType:    assetType,
				Side:         side,
				Price:        tradeData[i].Price.Float64(),
				Amount:       tradeData[i].Amount.Float64(),
				Timestamp:    tt,
			})
			if i == len(tradeData)-1 {
				if ts.Equal(tt) {
					// reached end of trades to crawl
					break allTrades
				}
				if timestampStart.IsZero() {
					break allTrades
				}
				ts = tt
			}
		}
	}

	err = p.AddTradesToBuffer(resp...)
	if err != nil {
		return nil, err
	}
	resp = trade.FilterTradesByTime(resp, timestampStart, timestampEnd)
	sort.Sort(trade.ByDate(resp))
	return resp, nil
}

// SubmitOrder submits a new order
func (p *Poloniex) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	if s == nil {
		return nil, common.ErrNilPointer
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}

	fPair, err := p.FormatExchangeCurrency(s.Pair, s.AssetType)
	if err != nil {
		return nil, err
	}
	var smartOrder bool
	var response *PlaceOrderResponse
	switch s.Type {
	case order.Stop, order.StopLimit:
		smartOrder = true
	case order.Limit, order.Market, order.LimitMaker:
	default:
		return nil, fmt.Errorf("%v order type %v is not supported", order.ErrTypeIsInvalid, s.Type)
	}
	if smartOrder {
		var sOrder *PlaceOrderResponse
		sOrder, err = p.CreateSmartOrder(ctx, &SmartOrderRequestParam{
			Symbol:        fPair,
			Side:          orderSideString(s.Side),
			Type:          orderTypeString(s.Type),
			AccountType:   accountTypeString(s.AssetType),
			Price:         s.Price,
			StopPrice:     s.TriggerPrice,
			Quantity:      s.Amount,
			ClientOrderID: s.ClientOrderID,
		})
		if err != nil {
			return nil, err
		}
		return s.DeriveSubmitResponse(sOrder.ID)
	}
	if p.Websocket.IsConnected() && p.Websocket.CanUseAuthenticatedEndpoints() && p.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		response, err = p.WsCreateOrder(&PlaceOrderParams{
			Symbol:      fPair,
			Price:       s.Price,
			Amount:      s.Amount,
			AllowBorrow: false,
			Type:        s.Type.String(),
			Side:        s.Side.String(),
		})
	} else {
		response, err = p.PlaceOrder(ctx, &PlaceOrderParams{
			Symbol:      fPair,
			Price:       s.Price,
			Amount:      s.Amount,
			AllowBorrow: false,
			Type:        s.Type.String(),
			Side:        s.Side.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	return s.DeriveSubmitResponse(response.ID)
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (p *Poloniex) ModifyOrder(ctx context.Context, action *order.Modify) (*order.ModifyResponse, error) {
	if action == nil {
		return nil, common.ErrNilPointer
	}
	if err := action.Validate(); err != nil {
		return nil, err
	}

	var orderID string
	resp, err := p.CancelReplaceOrder(ctx, &CancelReplaceOrderParam{
		orderID:       action.OrderID,
		ClientOrderID: action.ClientOrderID,
		Price:         action.Price,
		Quantity:      action.Amount,
		AmendedType:   action.Type.String(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "Couldn't locate order") {
			var smartOResponse *CancelReplaceSmartOrderResponse
			smartOResponse, err = p.CancelReplaceSmartOrder(ctx, &CancelReplaceSmartOrderParam{
				orderID:       action.OrderID,
				ClientOrderID: action.ClientOrderID,
				Price:         action.Price,
				StopPrice:     action.TriggerPrice,
				Quantity:      action.Amount,
				AmendedType:   orderTypeString(action.Type),
			})
			if err != nil {
				return nil, err
			}
			orderID = smartOResponse.ID
		} else {
			return nil, err
		}
	} else {
		orderID = resp.ID
	}
	modResp, err := action.DeriveModifyResponse()
	if err != nil {
		return nil, err
	}
	modResp.OrderID = orderID
	return modResp, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (p *Poloniex) CancelOrder(ctx context.Context, o *order.Cancel) error {
	if err := o.Validate(o.StandardCancel()); err != nil {
		return err
	}
	_, err := p.CancelOrderByID(ctx, o.OrderID)
	return err
}

// CancelBatchOrders cancels an orders by their corresponding ID numbers
func (p *Poloniex) CancelBatchOrders(ctx context.Context, o []order.Cancel) (*order.CancelBatchResponse, error) {
	if len(o) == 0 {
		return nil, order.ErrCancelOrderIsNil
	}
	orderIDs := make([]string, 0, len(o))
	clientOrderIDs := make([]string, 0, len(o))
	for i := range o {
		switch {
		case o[i].ClientOrderID != "":
			clientOrderIDs = append(clientOrderIDs, o[i].ClientOrderID)
		case o[i].OrderID != "":
			orderIDs = append(orderIDs, o[i].OrderID)
		default:
			return nil, order.ErrOrderIDNotSet
		}
	}
	resp := &order.CancelBatchResponse{
		Status: make(map[string]string),
	}
	if p.Websocket.IsConnected() && p.Websocket.CanUseAuthenticatedEndpoints() && p.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		wsCancelledOrders, err := p.WsCancelMultipleOrdersByIDs(&OrderCancellationParams{OrderIds: orderIDs, ClientOrderIds: clientOrderIDs})
		if err != nil {
			return nil, err
		}
		for i := range wsCancelledOrders {
			if wsCancelledOrders[i].ClientOrderID != "" {
				resp.Status[wsCancelledOrders[i].ClientOrderID] = wsCancelledOrders[i].State + " " + wsCancelledOrders[i].Message
				continue
			}
			orderID := strconv.FormatInt(wsCancelledOrders[i].OrderID, 10)
			resp.Status[orderID] = wsCancelledOrders[i].State + " " + wsCancelledOrders[i].Message
		}
	} else {
		cancelledOrders, err := p.CancelMultipleOrdersByIDs(ctx, &OrderCancellationParams{OrderIds: orderIDs, ClientOrderIds: clientOrderIDs})
		if err != nil {
			return nil, err
		}
		for i := range cancelledOrders {
			if cancelledOrders[i].ClientOrderID != "" {
				resp.Status[cancelledOrders[i].ClientOrderID] = cancelledOrders[i].State + " " + cancelledOrders[i].Message
				continue
			}
			resp.Status[cancelledOrders[i].OrderID] = cancelledOrders[i].State + " " + cancelledOrders[i].Message
		}
	}
	return resp, nil
}

// CancelAllOrders cancels all orders associated with a currency pair
func (p *Poloniex) CancelAllOrders(ctx context.Context, cancelOrd *order.Cancel) (order.CancelAllResponse, error) {
	cancelAllOrdersResponse := order.CancelAllResponse{
		Status: make(map[string]string),
	}
	if cancelOrd == nil {
		return cancelAllOrdersResponse, common.ErrNilPointer
	}
	var err error
	var pairs currency.Pairs
	if !cancelOrd.Pair.IsEmpty() {
		pairs = append(pairs, cancelOrd.Pair)
	}
	var resp []CancelOrderResponse
	if p.Websocket.IsConnected() && p.Websocket.CanUseAuthenticatedEndpoints() && p.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		var wsResponse []WsCancelOrderResponse
		wsResponse, err = p.WsCancelAllTradeOrders(pairs.Strings(), []string{accountTypeString(cancelOrd.AssetType)})
		if err != nil {
			return cancelAllOrdersResponse, err
		}
		for x := range wsResponse {
			cancelAllOrdersResponse.Status[strconv.FormatInt(wsResponse[x].OrderID, 10)] = wsResponse[x].State
		}
	} else {
		resp, err = p.CancelAllTradeOrders(ctx, pairs.Strings(), []string{accountTypeString(cancelOrd.AssetType)})
		if err != nil {
			return cancelAllOrdersResponse, err
		}
		for x := range resp {
			cancelAllOrdersResponse.Status[resp[x].OrderID] = resp[x].State
		}
	}
	resp, err = p.CancelAllSmartOrders(ctx, pairs.Strings(), []string{accountTypeString(cancelOrd.AssetType)})
	if err != nil {
		return cancelAllOrdersResponse, err
	}
	for x := range resp {
		cancelAllOrdersResponse.Status[resp[x].OrderID] = resp[x].State
	}
	return cancelAllOrdersResponse, nil
}

// GetOrderInfo returns order information based on order ID
func (p *Poloniex) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, _ asset.Item) (*order.Detail, error) {
	if pair.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	trades, err := p.GetTradesByOrderID(ctx, orderID)
	if err != nil && !strings.Contains(err.Error(), "Order not found") {
		return nil, err
	}
	orderTrades := make([]order.TradeHistory, len(trades))
	var oType order.Type
	var oSide order.Side
	for i := range trades {
		oType, err = order.StringToOrderType(trades[i].Type)
		if err != nil {
			return nil, err
		}
		oSide, err = order.StringToOrderSide(trades[i].Side)
		if err != nil {
			return nil, err
		}
		orderTrades[i] = order.TradeHistory{
			Price:     trades[i].Price.Float64(),
			Amount:    trades[i].Quantity.Float64(),
			Fee:       trades[i].FeeAmount.Float64(),
			Exchange:  p.Name,
			TID:       trades[i].ID,
			Type:      oType,
			Side:      oSide,
			Timestamp: trades[i].CreateTime.Time(),
			FeeAsset:  trades[i].FeeCurrency,
			Total:     trades[i].Amount.Float64(),
		}
	}
	var smartOrders []SmartOrderDetail
	resp, err := p.GetOrderDetail(ctx, orderID, "")
	if err != nil {
		smartOrders, err = p.GetSmartOrderDetail(ctx, orderID, "")
		if err != nil {
			return nil, err
		} else if len(smartOrders) == 0 {
			return nil, order.ErrOrderNotFound
		}
	}

	var dPair currency.Pair
	var oStatus order.Status
	if len(smartOrders) > 0 {
		dPair, err = currency.NewPairFromString(smartOrders[0].Symbol)
		if err != nil {
			return nil, err
		} else if !pair.IsEmpty() && !dPair.Equal(pair) {
			return nil, fmt.Errorf("order with ID %s expected a symbol %v, but got %v", orderID, pair, dPair)
		}
		oType, err = order.StringToOrderType(smartOrders[0].Type)
		if err != nil {
			return nil, err
		}
		oStatus, err = order.StringToOrderStatus(smartOrders[0].State)
		if err != nil {
			return nil, err
		}
		oSide, err = order.StringToOrderSide(smartOrders[0].Side)
		if err != nil {
			return nil, err
		}
		return &order.Detail{
			Price:         smartOrders[0].Price.Float64(),
			Amount:        smartOrders[0].Quantity.Float64(),
			QuoteAmount:   smartOrders[0].Amount.Float64(),
			Exchange:      p.Name,
			OrderID:       smartOrders[0].ID,
			ClientOrderID: smartOrders[0].ClientOrderID,
			Type:          oType,
			Side:          oSide,
			Status:        oStatus,
			AssetType:     stringToAccountType(smartOrders[0].AccountType),
			Date:          smartOrders[0].CreateTime.Time(),
			LastUpdated:   smartOrders[0].UpdateTime.Time(),
			Pair:          dPair,
			Trades:        orderTrades,
		}, nil
	}
	dPair, err = currency.NewPairFromString(resp.Symbol)
	if err != nil {
		return nil, err
	} else if !pair.IsEmpty() && !dPair.Equal(pair) {
		return nil, fmt.Errorf("order with ID %s expected a symbol %v, but got %v", orderID, pair, dPair)
	}
	oType, err = order.StringToOrderType(resp.Type)
	if err != nil {
		return nil, err
	}
	oStatus, err = order.StringToOrderStatus(resp.State)
	if err != nil {
		return nil, err
	}
	oSide, err = order.StringToOrderSide(resp.Side)
	if err != nil {
		return nil, err
	}
	return &order.Detail{
		Price:                resp.Price.Float64(),
		Amount:               resp.Quantity.Float64(),
		AverageExecutedPrice: resp.AvgPrice.Float64(),
		QuoteAmount:          resp.Amount.Float64(),
		ExecutedAmount:       resp.FilledQuantity.Float64(),
		RemainingAmount:      resp.Quantity.Float64() - resp.FilledAmount.Float64(),
		Cost:                 resp.FilledQuantity.Float64() * resp.AvgPrice.Float64(),
		Exchange:             p.Name,
		OrderID:              resp.ID,
		ClientOrderID:        resp.ClientOrderID,
		Type:                 oType,
		Side:                 oSide,
		Status:               oStatus,
		AssetType:            stringToAccountType(resp.AccountType),
		Date:                 resp.CreateTime.Time(),
		LastUpdated:          resp.UpdateTime.Time(),
		Pair:                 dPair,
		Trades:               orderTrades,
	}, nil
}

// GetDepositAddress returns a deposit address for a specified currency
func (p *Poloniex) GetDepositAddress(ctx context.Context, cryptocurrency currency.Code, _, chain string) (*deposit.Address, error) {
	depositAddrs, err := p.GetDepositAddresses(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}
	// Some coins use a main address, so we must use this in conjunction with the returned
	// deposit address to produce the full deposit address and payment-id
	currencies, err := p.GetCurrencyInformation(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	coinParams, ok := currencies[cryptocurrency.Upper().String()]
	if !ok {
		return nil, fmt.Errorf("unable to find currency %s in map", cryptocurrency)
	}

	var address, paymentID string
	if coinParams.Type == "address-payment-id" && coinParams.DepositAddress != "" {
		paymentID, ok = (*depositAddrs)[cryptocurrency.Upper().String()]
		if !ok {
			newAddr, err := p.NewCurrencyDepositAddress(ctx, cryptocurrency)
			if err != nil {
				return nil, err
			}
			paymentID = newAddr
		}
		return &deposit.Address{
			Address: coinParams.DepositAddress,
			Tag:     paymentID,
			Chain:   coinParams.ParentChain,
		}, nil
	}

	address, ok = (*depositAddrs)[cryptocurrency.Upper().String()]
	if !ok {
		if len(coinParams.ChildChains) > 1 && chain != "" && !common.StringDataCompare(coinParams.ChildChains, chain) {
			return nil, fmt.Errorf("currency %s has %v chains available, one of these must be specified",
				cryptocurrency,
				coinParams.ChildChains)
		}

		coinParams, ok = currencies[cryptocurrency.Upper().String()]
		if !ok {
			return nil, fmt.Errorf("unable to find currency %s in map", cryptocurrency)
		}
		if coinParams.WalletDepositState != "ENABLED" {
			return nil, fmt.Errorf("deposits and withdrawals for %v are currently disabled", cryptocurrency.Upper().String())
		}

		newAddr, err := p.NewCurrencyDepositAddress(ctx, cryptocurrency)
		if err != nil {
			return nil, err
		}
		address = newAddr
	}
	return &deposit.Address{
		Address: address,
	}, nil
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (p *Poloniex) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if withdrawRequest == nil {
		return nil, common.ErrNilPointer
	}
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}
	v, err := p.WithdrawCurrency(ctx, &WithdrawCurrencyParam{
		Currency: withdrawRequest.Currency,
		Address:  withdrawRequest.Crypto.Address,
		Amount:   withdrawRequest.Amount})
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		Name: p.Name,
		ID:   strconv.FormatInt(v.WithdrawRequestID, 10),
	}, err
}

// WithdrawFiatFunds returns a withdrawal ID when a
// withdrawal is submitted
func (p *Poloniex) WithdrawFiatFunds(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a
// withdrawal is submitted
func (p *Poloniex) WithdrawFiatFundsToInternationalBank(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetFeeByType returns an estimate of fee based on type of transaction
func (p *Poloniex) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	if feeBuilder == nil {
		return 0, common.ErrNilPointer
	}
	if feeBuilder == nil {
		return 0, fmt.Errorf("%T %w", feeBuilder, common.ErrNilPointer)
	}
	if (!p.AreCredentialsValid(ctx) || p.SkipAuthCheck) && // Todo check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return p.GetFee(ctx, feeBuilder)
}

// GetActiveOrders retrieves any orders that are active/open
func (p *Poloniex) GetActiveOrders(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	if req == nil {
		return nil, common.ErrNilPointer
	}
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	var samplePair currency.Pair
	if len(req.Pairs) == 1 {
		samplePair = req.Pairs[0]
	}
	resp, err := p.GetOpenOrders(ctx, samplePair, orderSideString(req.Side), "", req.FromOrderID, 0)
	if err != nil {
		return nil, err
	}
	orders := make([]order.Detail, 0, len(resp))
	for a := range resp {
		var symbol currency.Pair
		symbol, err = currency.NewPairFromString(resp[a].Symbol)
		if err != nil {
			return nil, err
		}
		if len(req.Pairs) != 0 && req.Pairs.Contains(symbol, true) {
			continue
		}
		var orderSide order.Side
		orderSide, err = order.StringToOrderSide(resp[a].Side)
		if err != nil {
			return nil, err
		}
		oType, err := order.StringToOrderType(resp[a].Type)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order.Detail{
			Type:     oType,
			OrderID:  resp[a].ID,
			Side:     orderSide,
			Amount:   resp[a].Amount.Float64(),
			Date:     resp[a].CreateTime.Time(),
			Price:    resp[a].Price.Float64(),
			Pair:     symbol,
			Exchange: p.Name,
		})
	}
	return req.Filter(p.Name, orders), nil
}

func accountTypeString(assetType asset.Item) string {
	switch assetType {
	case asset.Spot:
		return "SPOT"
	case asset.Futures:
		return "FUTURE"
	default:
		return ""
	}
}

func stringToAccountType(assetType string) asset.Item {
	switch assetType {
	case "SPOT":
		return asset.Spot
	case "FUTURE":
		return asset.Futures
	default:
		return asset.Empty
	}
}

func orderSideString(oSide order.Side) string {
	switch oSide {
	case order.Buy, order.Sell:
		return oSide.String()
	default:
		return ""
	}
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (p *Poloniex) GetOrderHistory(ctx context.Context, req *order.MultiOrderRequest) (order.FilteredOrders, error) {
	if req == nil {
		return nil, common.ErrNilPointer
	}
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	resp, err := p.GetOrdersHistory(ctx, currency.EMPTYPAIR, accountTypeString(req.AssetType), orderTypeString(req.Type), orderSideString(req.Side), "", "", 0, 100, req.StartTime, req.EndTime, false)
	if err != nil {
		return nil, err
	}

	var oSide order.Side
	var oType order.Type
	orders := make([]order.Detail, 0, len(resp))
	for i := range resp {
		var pair currency.Pair
		pair, err = currency.NewPairFromString(resp[i].Symbol)
		if err != nil {
			return nil, err
		}
		if len(req.Pairs) != 0 && !req.Pairs.Contains(pair, true) {
			continue
		}
		oSide, err = order.StringToOrderSide(resp[i].Side)
		if err != nil {
			return nil, err
		}
		oType, err = order.StringToOrderType(resp[i].Type)
		if err != nil {
			return nil, err
		}
		var assetType asset.Item
		assetType, err = asset.New(resp[i].AccountType)
		if err != nil {
			return nil, err
		}
		detail := order.Detail{
			Side:                 oSide,
			Amount:               resp[i].Amount.Float64(),
			ExecutedAmount:       resp[i].FilledAmount.Float64(),
			Price:                resp[i].Price.Float64(),
			AverageExecutedPrice: resp[i].AvgPrice.Float64(),
			Pair:                 pair,
			Type:                 oType,
			Exchange:             p.Name,
			QuoteAmount:          resp[i].Amount.Float64() * resp[i].AvgPrice.Float64(),
			RemainingAmount:      resp[i].Quantity.Float64() - resp[i].FilledQuantity.Float64(),
			OrderID:              resp[i].ID,
			ClientOrderID:        resp[i].ClientOrderID,
			Status:               order.Filled,
			AssetType:            assetType,
			Date:                 resp[i].CreateTime.Time(),
			LastUpdated:          resp[i].UpdateTime.Time(),
		}
		detail.InferCostsAndTimes()
		orders = append(orders, detail)
	}
	smartOrders, err := p.GetSmartOrderHistory(ctx, currency.EMPTYPAIR, accountTypeString(req.AssetType),
		orderTypeString(req.Type), orderSideString(req.Side), "", "", 0, 100, req.StartTime, req.EndTime, false)
	if err != nil {
		return nil, err
	}
	for i := range smartOrders {
		var pair currency.Pair
		pair, err = currency.NewPairFromString(smartOrders[i].Symbol)
		if err != nil {
			return nil, err
		}
		if len(req.Pairs) != 0 && !req.Pairs.Contains(pair, true) {
			continue
		}
		oSide, err = order.StringToOrderSide(smartOrders[i].Side)
		if err != nil {
			return nil, err
		}
		oType, err = order.StringToOrderType(smartOrders[i].Type)
		if err != nil {
			return nil, err
		}
		assetType, err := asset.New(smartOrders[i].AccountType)
		if err != nil {
			return nil, err
		}
		detail := order.Detail{
			Side:          oSide,
			Amount:        smartOrders[i].Amount.Float64(),
			Price:         smartOrders[i].Price.Float64(),
			TriggerPrice:  smartOrders[i].StopPrice.Float64(),
			Pair:          pair,
			Type:          oType,
			Exchange:      p.Name,
			OrderID:       smartOrders[i].ID,
			ClientOrderID: smartOrders[i].ClientOrderID,
			Status:        order.Filled,
			AssetType:     assetType,
			Date:          smartOrders[i].CreateTime.Time(),
			LastUpdated:   smartOrders[i].UpdateTime.Time(),
		}
		detail.InferCostsAndTimes()
		orders = append(orders, detail)
	}
	return req.Filter(p.Name, orders), nil
}

// ValidateAPICredentials validates current credentials used for wrapper
// functionality
func (p *Poloniex) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	_, err := p.UpdateAccountInfo(ctx, assetType)
	return p.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (p *Poloniex) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := p.GetKlineRequest(pair, a, interval, start, end, false)
	if err != nil {
		return nil, err
	}
	resp, err := p.GetCandlesticks(ctx, req.RequestFormatted, req.ExchangeInterval, req.Start, req.End, req.RequestLimit)
	if err != nil {
		return nil, err
	}
	timeSeries := make([]kline.Candle, len(resp))
	for x := range resp {
		timeSeries[x] = kline.Candle{
			Time:   resp[x].StartTime,
			Open:   resp[x].Open,
			High:   resp[x].High,
			Low:    resp[x].Low,
			Close:  resp[x].Close,
			Volume: resp[x].Quantity,
		}
	}
	return req.ProcessResponse(timeSeries)
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (p *Poloniex) GetHistoricCandlesExtended(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := p.GetKlineExtendedRequest(pair, a, interval, start, end)
	if err != nil {
		return nil, err
	}

	timeSeries := make([]kline.Candle, 0, req.Size())
	for i := range req.RangeHolder.Ranges {
		resp, err := p.GetCandlesticks(ctx,
			req.RequestFormatted,
			req.ExchangeInterval,
			req.RangeHolder.Ranges[i].Start.Time,
			req.RangeHolder.Ranges[i].End.Time,
			req.RequestLimit,
		)
		if err != nil {
			return nil, err
		}
		for x := range resp {
			timeSeries = append(timeSeries, kline.Candle{
				Time:   resp[x].StartTime,
				Open:   resp[x].Open,
				High:   resp[x].High,
				Low:    resp[x].Low,
				Close:  resp[x].Close,
				Volume: resp[x].Quantity,
			})
		}
	}
	return req.ProcessResponse(timeSeries)
}

// GetAvailableTransferChains returns the available transfer blockchains for the specific
// cryptocurrency
func (p *Poloniex) GetAvailableTransferChains(ctx context.Context, cryptocurrency currency.Code) ([]string, error) {
	if cryptocurrency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	currencies, err := p.GetCurrencyInformations(ctx)
	if err != nil {
		return nil, err
	}
	for a := range currencies {
		curr, ok := currencies[a][cryptocurrency.Upper().String()]
		if !ok {
			continue
		}
		return curr.ChildChains, nil
	}
	return nil, fmt.Errorf("%w for currency %v", errChainsNotFound, cryptocurrency)
}

// GetServerTime returns the current exchange server time.
func (p *Poloniex) GetServerTime(ctx context.Context, _ asset.Item) (time.Time, error) {
	sysServerTime, err := p.GetSystemTimestamp(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return sysServerTime.ServerTime.Time(), nil
}

// GetFuturesContractDetails returns all contracts from the exchange by asset type
func (p *Poloniex) GetFuturesContractDetails(context.Context, asset.Item) ([]futures.Contract, error) {
	// TODO: implement with API upgrade
	return nil, common.ErrFunctionNotSupported
}

// GetLatestFundingRates returns the latest funding rates data
func (p *Poloniex) GetLatestFundingRates(context.Context, *fundingrate.LatestRateRequest) ([]fundingrate.LatestRateResponse, error) {
	// TODO: implement with API upgrade
	return nil, common.ErrFunctionNotSupported
}

// UpdateOrderExecutionLimits updates order execution limits
func (p *Poloniex) UpdateOrderExecutionLimits(ctx context.Context, a asset.Item) error {
	if !p.SupportsAsset(a) {
		return fmt.Errorf("%w asset: %v", asset.ErrNotSupported, a)
	}
	instruments, err := p.GetSymbolInformation(ctx, currency.EMPTYPAIR)
	if err != nil {
		return err
	}
	limits := make([]order.MinMaxLevel, len(instruments))
	for x := range instruments {
		pair, err := currency.NewPairFromString(instruments[x].Symbol)
		if err != nil {
			return err
		}

		limits[x] = order.MinMaxLevel{
			Pair:                    pair,
			Asset:                   a,
			PriceStepIncrementSize:  instruments[x].SymbolTradeLimit.PriceScale,
			MinimumBaseAmount:       instruments[x].SymbolTradeLimit.MinQuantity.Float64(),
			MinimumQuoteAmount:      instruments[x].SymbolTradeLimit.MinAmount.Float64(),
			AmountStepIncrementSize: instruments[x].SymbolTradeLimit.AmountScale,
			QuoteStepIncrementSize:  instruments[x].SymbolTradeLimit.QuantityScale,
		}
	}
	return p.LoadLimits(limits)
}
