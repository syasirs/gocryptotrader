package coinbaseinternational

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
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
func (co *CoinbaseInternational) GetDefaultConfig(ctx context.Context) (*config.Exchange, error) {
	co.SetDefaults()
	exchCfg := new(config.Exchange)
	exchCfg.Name = co.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = co.BaseCurrencies

	co.SetupDefaults(exchCfg)

	if co.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := co.UpdateTradablePairs(ctx, true)
		if err != nil {
			return nil, err
		}
	}
	return exchCfg, nil
}

// SetDefaults sets the basic defaults for CoinbaseInternational
func (co *CoinbaseInternational) SetDefaults() {
	co.Name = "CoinbaseInternational"
	co.Enabled = true
	co.Verbose = true
	co.API.CredentialsValidator.RequiresKey = true
	co.API.CredentialsValidator.RequiresClientID = true
	co.API.CredentialsValidator.RequiresSecret = true
	co.API.CredentialsValidator.RequiresBase64DecodeSecret = true
	requestFmt := &currency.PairFormat{Uppercase: true, Delimiter: ":"}
	configFmt := &currency.PairFormat{}
	err := co.SetGlobalPairsManager(requestFmt, configFmt)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	fmt := currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true, Delimiter: currency.DashDelimiter},
		ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: currency.DashDelimiter},
	}

	err = co.StoreAssetPairFormat(asset.Spot, fmt)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	co.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}
	co.Requester, err = request.New(co.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	co.API.Endpoints = co.NewEndpoints()
	co.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      coinbaseInternationalAPIURL,
		exchange.WebsocketSpot: coinbaseinternationalWSAPIURL,
	})
	co.Websocket = stream.New()
	co.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	co.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	co.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (co *CoinbaseInternational) Setup(exch *config.Exchange) error {
	err := exch.Validate()
	if err != nil {
		return err
	}
	if !exch.Enabled {
		co.SetEnabled(false)
		return nil
	}
	err = co.SetupDefaults(exch)
	if err != nil {
		return err
	}
	wsRunningEndpoint, err := co.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	err = co.Websocket.Setup(&stream.WebsocketSetup{
		ExchangeConfig:        exch,
		DefaultURL:            coinbaseinternationalWSAPIURL,
		RunningURL:            wsRunningEndpoint,
		Connector:             co.WsConnect,
		Subscriber:            co.Subscribe,
		Unsubscriber:          co.Unsubscribe,
		GenerateSubscriptions: co.GenerateDefaultSubscriptions,
		Features:              &co.Features.Supports.WebsocketCapabilities,
		OrderbookBufferConfig: buffer.Config{
			SortBuffer:            true,
			SortBufferByUpdateIDs: true,
		},
	})
	if err != nil {
		return err
	}
	return co.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the CoinbaseInternational go routine
func (co *CoinbaseInternational) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return fmt.Errorf("%T %w", wg, common.ErrNilPointer)
	}
	wg.Add(1)
	go func() {
		co.Run(ctx)
		wg.Done()
	}()
	return nil
}

// Run implements the CoinbaseInternational wrapper
func (co *CoinbaseInternational) Run(ctx context.Context) {
	if co.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			co.Name,
			common.IsEnabled(co.Websocket.IsEnabled()))
		co.PrintEnabledPairs()
	}

	if !co.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := co.UpdateTradablePairs(ctx, false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			co.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (co *CoinbaseInternational) FetchTradablePairs(ctx context.Context, a asset.Item) (currency.Pairs, error) {
	if !co.SupportsAsset(a) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, a)
	}
	instruments, err := co.GetInstruments(ctx)
	if err != nil {
		return nil, err
	}
	pairs := make([]currency.Pair, 0, len(instruments))
	for x := range instruments {
		if instruments[x].TradingState != "TRADING" {
			continue
		}
		cp, err := currency.NewPairFromString(instruments[x].Symbol)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, cp)
	}
	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (co *CoinbaseInternational) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	assetTypes := co.GetAssetTypes(false)
	for x := range assetTypes {
		pairs, err := co.FetchTradablePairs(ctx, assetTypes[x])
		if err != nil {
			return err
		}

		err = co.UpdatePairs(pairs, assetTypes[x], false, forceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTicker updates and returns the ticker for a currency pair
func (co *CoinbaseInternational) UpdateTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	if assetType != asset.Spot {
		return nil, fmt.Errorf("%w asset type %v", asset.ErrNotSupported, asset.Spot)
	}
	format, err := co.GetPairFormat(asset.Spot, true)
	if err != nil {
		return nil, err
	}
	if p.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	tick, err := co.GetQuotePerInstrument(ctx, p.String(), "", "")
	if err != nil {
		return nil, err
	}
	err = ticker.ProcessTicker(&ticker.Price{
		High:         tick.LimitUp.Float64(),
		Low:          tick.LimitDown.Float64(),
		Bid:          tick.BestBidPrice.Float64(),
		BidSize:      tick.BestBidSize.Float64(),
		Ask:          tick.BestAskPrice.Float64(),
		AskSize:      tick.BestAskSize.Float64(),
		Open:         tick.MarkPrice.Float64(),
		Close:        tick.SettlementPrice.Float64(),
		LastUpdated:  tick.Timestamp,
		Volume:       tick.TradeQty.Float64() / tick.TradePrice.Float64(), // TODO: if the volume is representing the quote volume,  then the base quentity is the quote volume divided by the trade price.
		QuoteVolume:  tick.TradeQty.Float64(),
		ExchangeName: co.Name,
		AssetType:    asset.Spot,
		Pair:         p.Format(format),
	})
	if err != nil {
		return nil, err
	}
	return ticker.GetTicker(co.Name, p, asset.Spot)
}

// UpdateTickers updates all currency pairs of a given asset type
func (co *CoinbaseInternational) UpdateTickers(ctx context.Context, assetType asset.Item) error {
	if assetType != asset.Spot {
		return fmt.Errorf("%w asset type %v", asset.ErrNotSupported, assetType)
	}
	var tick *QuoteInformation
	enabledPairs, err := co.GetEnabledPairs(asset.Spot)
	if err != nil {
		return err
	}
	for x := range enabledPairs {
		tick, err = co.GetQuotePerInstrument(ctx, enabledPairs[x].String(), "", "")
		if err != nil {
			return err
		}
		err = ticker.ProcessTicker(&ticker.Price{
			High:         tick.LimitUp.Float64(),
			Low:          tick.LimitDown.Float64(),
			Bid:          tick.BestBidPrice.Float64(),
			BidSize:      tick.BestBidSize.Float64(),
			Ask:          tick.BestAskPrice.Float64(),
			AskSize:      tick.BestAskSize.Float64(),
			Open:         tick.MarkPrice.Float64(),
			Close:        tick.SettlementPrice.Float64(),
			LastUpdated:  tick.Timestamp,
			Volume:       tick.TradeQty.Float64() / tick.TradePrice.Float64(), // TODO: if the volume is representing the quote volume,  then the base quentity is the quote volume divided by the trade price.
			QuoteVolume:  tick.TradeQty.Float64(),
			Pair:         enabledPairs[x],
			AssetType:    asset.Spot,
			ExchangeName: co.Name,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchTicker returns the ticker for a currency pair
func (co *CoinbaseInternational) FetchTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(co.Name, p, assetType)
	if err != nil {
		return co.UpdateTicker(ctx, p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (co *CoinbaseInternational) FetchOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(co.Name, pair, assetType)
	if err != nil {
		return co.UpdateOrderbook(ctx, pair, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (co *CoinbaseInternational) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        co.Name,
		Pair:            pair,
		Asset:           assetType,
		VerifyOrderbook: co.CanVerifyOrderbook,
	}
	format, err := co.GetPairFormat(asset.Spot, true)
	if err != nil {
		return nil, err
	}
	orderbookNew, err := co.GetQuotePerInstrument(ctx, pair.Format(format).String(), "", "")
	if err != nil {
		return book, err
	}
	book.Bids = []orderbook.Item{{
		Amount: orderbookNew.BestBidSize.Float64(),
		Price:  orderbookNew.BestBidPrice.Float64(),
	}}
	book.Asks = []orderbook.Item{{
		Amount: orderbookNew.BestAskSize.Float64(),
		Price:  orderbookNew.BestAskPrice.Float64(),
	}}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(co.Name, pair, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (co *CoinbaseInternational) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	if assetType != asset.Spot {
		return account.Holdings{}, asset.ErrNotSupported
	}
	portfolios, err := co.GetAllUserPortfolios(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	holdings := account.Holdings{
		Exchange: co.Name,
		Accounts: make([]account.SubAccount, len(portfolios)),
	}
	var balances []PortfolioBalance
	for p := range portfolios {
		balances, err = co.ListPortfolioBalances(ctx, portfolios[p].PortfolioUUID, portfolios[p].PortfolioID)
		if err != nil {
			return account.Holdings{}, err
		}
		holdings.Accounts[p] = account.SubAccount{
			AssetType:  asset.Spot,
			ID:         portfolios[p].PortfolioID,
			Currencies: make([]account.Balance, len(balances)),
		}
		for b := range balances {
			holdings.Accounts[p].Currencies[b] = account.Balance{
				Currency:               currency.NewCode(balances[b].AssetName),
				Total:                  balances[b].Quantity.Float64(),
				Hold:                   balances[b].Hold.Float64(),
				Free:                   balances[b].Quantity.Float64() - balances[b].Hold.Float64(),
				AvailableWithoutBorrow: balances[b].MaxWithdrawAmount.Float64(),
				Borrowed:               balances[b].CollateralValue.Float64(), // TODO: amount to be borrowed
			}
		}
	}
	return holdings, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (co *CoinbaseInternational) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	creds, err := co.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	acc, err := account.GetHoldings(co.Name, creds, assetType)
	if err != nil {
		return co.UpdateAccountInfo(ctx, assetType)
	}
	return acc, nil
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (co *CoinbaseInternational) GetAccountFundingHistory(ctx context.Context) ([]exchange.FundingHistory, error) {
	history, err := co.ListMatchingTransfers(ctx, "", "", "", "", 0, 0, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.FundingHistory, len(history.Results))
	for j := range history.Results {
		resp[j] = exchange.FundingHistory{
			ExchangeName: co.Name,
			CryptoTxID:   history.Results[j].TransferUUID,
			CryptoChain:  history.Results[j].NetworkName,
			Timestamp:    history.Results[j].CreatedAt,
			Status:       history.Results[j].Status,
			Currency:     history.Results[j].Asset,
			Amount:       history.Results[j].Amount,
			TransferType: history.Results[j].Type,
		}
	}
	return resp, nil
}

// GetWithdrawalsHistory returns previous withdrawals data
func (co *CoinbaseInternational) GetWithdrawalsHistory(ctx context.Context, c currency.Code, a asset.Item) ([]exchange.WithdrawalHistory, error) {
	if a != asset.Spot {
		return nil, asset.ErrNotSupported
	}
	history, err := co.ListMatchingTransfers(ctx, "", "", "", "WITHDRAW", 0, 0, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.WithdrawalHistory, len(history.Results))
	for j := range history.Results {
		resp[j] = exchange.WithdrawalHistory{
			Status:       history.Results[j].Status,
			Timestamp:    history.Results[j].CreatedAt,
			Currency:     history.Results[j].Asset,
			Amount:       history.Results[j].Amount,
			TransferType: history.Results[j].Type,
			CryptoTxID:   history.Results[j].TransferUUID,
			CryptoChain:  history.Results[j].NetworkName,
		}
	}
	return resp, nil
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (co *CoinbaseInternational) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return nil, common.ErrNotYetImplemented
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (co *CoinbaseInternational) GetHistoricTrades(ctx context.Context, p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	return nil, common.ErrNotYetImplemented
}

// GetServerTime returns the current exchange server time.
func (co *CoinbaseInternational) GetServerTime(ctx context.Context, a asset.Item) (time.Time, error) {
	return time.Time{}, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (co *CoinbaseInternational) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	oType, err := orderTypeString(s.Type)
	if err != nil {
		return nil, err
	}
	response, err := co.CreateOrder(ctx, &OrderRequestParams{
		ClientOrderID: s.ClientOrderID,
		Side:          s.Side.String(),
		BaseSize:      s.Amount,
		Instrument:    s.Pair.String(),
		OrderType:     oType,
		Price:         s.Price,
		StopPrice:     s.TriggerPrice,
		PostOnly:      s.PostOnly,
	})
	if err != nil {
		return nil, err
	}
	oStatus, err := order.StringToOrderStatus(response.OrderStatus)
	if err != nil {
		return nil, err
	}
	return &order.SubmitResponse{
		Exchange:      co.Name,
		Type:          s.Type,
		Side:          s.Side,
		Pair:          s.Pair,
		AssetType:     asset.Spot,
		PostOnly:      s.PostOnly,
		ReduceOnly:    s.ReduceOnly,
		Leverage:      s.Leverage,
		Price:         response.Price,
		Amount:        response.Size,
		TriggerPrice:  response.StopPrice,
		ClientOrderID: response.ClientOrderID,
		Status:        oStatus,
		OrderID:       strconv.FormatInt(response.OrderID, 10),
		Fee:           response.Fee.Float64(),
	}, nil
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (co *CoinbaseInternational) ModifyOrder(ctx context.Context, action *order.Modify) (*order.ModifyResponse, error) {
	if err := action.Validate(); err != nil {
		return nil, err
	}
	var orderID string
	switch {
	case action.OrderID != "":
		orderID = action.OrderID
	case action.ClientOrderID != "":
		orderID = action.ClientOrderID
	}
	response, err := co.ModifyOpenOrder(ctx, orderID, &ModifyOrderParam{
		ClientOrderID: action.ClientOrderID,
		Portfolio:     "",
		Price:         action.Price,
		StopPrice:     action.TriggerPrice,
		Size:          action.Amount,
	})
	if err != nil {
		return nil, err
	}
	resp, err := action.DeriveModifyResponse()
	if err != nil {
		return nil, nil
	}
	resp.OrderID = strconv.FormatInt(response.OrderID, 10)
	return resp, nil
}

// CancelOrder cancels an order by its corresponding ID number
func (co *CoinbaseInternational) CancelOrder(ctx context.Context, ord *order.Cancel) error {
	err := ord.Validate(ord.StandardCancel())
	if err != nil {
		return err
	}
	_, err = co.CancelTradeOrder(ctx, ord.OrderID, ord.ClientOrderID, ord.AccountID, "")
	if err != nil {
		return err
	}
	return nil
}

// CancelBatchOrders cancels orders by their corresponding ID numbers
func (co *CoinbaseInternational) CancelBatchOrders(ctx context.Context, orders []order.Cancel) (*order.CancelBatchResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (co *CoinbaseInternational) CancelAllOrders(ctx context.Context, orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	return order.CancelAllResponse{}, common.ErrNotYetImplemented
}

// GetOrderInfo returns order information based on order ID
func (co *CoinbaseInternational) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (*order.Detail, error) {
	resp, err := co.GetOrderDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}
	oType, err := order.StringToOrderType(resp.Type)
	if err != nil {
		return nil, err
	}
	oSide, err := order.StringToOrderSide(resp.Side)
	if err != nil {
		return nil, err
	}
	oStatus, err := order.StringToOrderStatus(resp.OrderStatus)
	if err != nil {
		return nil, err
	}
	pair, err = currency.NewPairFromString(resp.Symbol)
	if err != nil {
		return nil, err
	}
	return &order.Detail{
		Price:                resp.Price,
		Amount:               resp.Size,
		Exchange:             co.Name,
		TriggerPrice:         resp.StopPrice,
		AverageExecutedPrice: resp.AveragePrice.Float64(),
		QuoteAmount:          resp.Size * resp.AveragePrice.Float64(),
		ExecutedAmount:       resp.ExecQty.Float64(),
		RemainingAmount:      resp.Size - resp.ExecQty.Float64(),
		Fee:                  resp.Fee.Float64(),
		OrderID:              strconv.FormatInt(resp.OrderID, 10),
		ClientOrderID:        resp.ClientOrderID,
		Type:                 oType,
		Side:                 oSide,
		Status:               oStatus,
		AssetType:            asset.Spot,
		CloseTime:            resp.ExpireTime,
		Pair:                 pair,
	}, nil
}

// GetDepositAddress returns a deposit address for a specified currency
func (co *CoinbaseInternational) GetDepositAddress(ctx context.Context, c currency.Code, accountID string, chain string) (*deposit.Address, error) {
	return nil, common.ErrNotYetImplemented
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (co *CoinbaseInternational) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}
	resp, err := co.WithdrawToCryptoAddress(ctx, &WithdrawCryptoParams{
		Portfolio:       withdrawRequest.ClientOrderID,
		AssetIdentifier: withdrawRequest.Currency.String(),
		Amount:          withdrawRequest.Amount,
		Address:         withdrawRequest.Crypto.Address,
	})
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		Name: co.Name,
		ID:   resp.Idem,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (co *CoinbaseInternational) WithdrawFiatFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (co *CoinbaseInternational) WithdrawFiatFundsToInternationalBank(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// GetActiveOrders retrieves any orders that are active/open
func (co *CoinbaseInternational) GetActiveOrders(ctx context.Context, getOrdersRequest *order.MultiOrderRequest) (order.FilteredOrders, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}
	var instrument string
	if len(getOrdersRequest.Pairs) == 1 {
		instrument = getOrdersRequest.Pairs[0].String()
	}
	response, err := co.GetOpenOrders(ctx, "", "", instrument, "", "", getOrdersRequest.StartTime, 0, 0)
	if err != nil {
		return nil, err
	}
	orders := make([]order.Detail, 0, len(response.Results))
	for x := range response.Results {
		oType, err := order.StringToOrderType(response.Results[x].Type)
		if err != nil {
			return nil, err
		}
		oSide, err := order.StringToOrderSide(response.Results[x].Side)
		if err != nil {
			return nil, err
		}
		oStatus, err := order.StringToOrderStatus(response.Results[x].OrderStatus)
		if err != nil {
			return nil, err
		}
		var pair currency.Pair
		pair, err = currency.NewPairFromString(response.Results[x].Symbol)
		if err != nil {
			return nil, err
		}
		if len(getOrdersRequest.Pairs) != 0 && getOrdersRequest.Pairs.Contains(pair, true) {
			continue
		}
		orders = append(orders, order.Detail{
			Amount:               response.Results[x].Size,
			Price:                response.Results[x].Price,
			TriggerPrice:         response.Results[x].StopPrice,
			AverageExecutedPrice: response.Results[x].AveragePrice.Float64(),
			QuoteAmount:          response.Results[x].Size * response.Results[x].AveragePrice.Float64(),
			RemainingAmount:      response.Results[x].Size - response.Results[x].ExecQty.Float64(),
			OrderID:              strconv.FormatInt(response.Results[x].OrderID, 10),
			ExecutedAmount:       response.Results[x].ExecQty.Float64(),
			Fee:                  response.Results[x].Fee.Float64(),
			ClientOrderID:        response.Results[x].ClientOrderID,
			CloseTime:            response.Results[x].ExpireTime,
			Exchange:             co.Name,
			Type:                 oType,
			Side:                 oSide,
			Status:               oStatus,
			AssetType:            asset.Spot,
			Pair:                 pair,
		})
	}
	return orders, nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (co *CoinbaseInternational) GetOrderHistory(ctx context.Context, getOrdersRequest *order.MultiOrderRequest) (order.FilteredOrders, error) {
	return nil, common.ErrNotYetImplemented
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (co *CoinbaseInternational) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	if feeBuilder == nil {
		return 0, fmt.Errorf("%T %w", feeBuilder, common.ErrNilPointer)
	}
	if !co.AreCredentialsValid(ctx) && // TODO check connection status
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return co.GetFee(ctx, feeBuilder)
}

// ValidateAPICredentials validates current credentials used for wrapper
func (co *CoinbaseInternational) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	_, err := co.UpdateAccountInfo(ctx, assetType)
	return co.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (co *CoinbaseInternational) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	return nil, common.ErrNotYetImplemented
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (co *CoinbaseInternational) GetHistoricCandlesExtended(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	return nil, common.ErrNotYetImplemented
}

// GetFuturesContractDetails returns all contracts from the exchange by asset type
func (co *CoinbaseInternational) GetFuturesContractDetails(context.Context, asset.Item) ([]futures.Contract, error) {
	return nil, common.ErrFunctionNotSupported
}
