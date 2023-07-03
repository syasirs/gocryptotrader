package cryptodotcom

import (
	"context"
	"fmt"
	"sort"
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
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// GetDefaultConfig returns a default exchange config
func (cr *Cryptodotcom) GetDefaultConfig(ctx context.Context) (*config.Exchange, error) {
	cr.SetDefaults()
	exchCfg := new(config.Exchange)
	exchCfg.Name = cr.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = cr.BaseCurrencies
	err := cr.SetupDefaults(exchCfg)
	if err != nil {
		return nil, err
	}
	if cr.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := cr.UpdateTradablePairs(ctx, true)
		if err != nil {
			return nil, err
		}
	}
	return exchCfg, nil
}

// SetDefaults sets the basic defaults for Cryptodotcom
func (cr *Cryptodotcom) SetDefaults() {
	cr.Name = "Cryptodotcom"
	cr.Enabled = true
	cr.Verbose = true
	cr.API.CredentialsValidator.RequiresKey = true
	cr.API.CredentialsValidator.RequiresSecret = true

	requestFmt := &currency.PairFormat{Uppercase: true, Delimiter: currency.UnderscoreDelimiter}
	configFmt := &currency.PairFormat{Uppercase: true, Delimiter: currency.UnderscoreDelimiter}
	err := cr.SetGlobalPairsManager(requestFmt, configFmt, asset.Spot)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	// Fill out the capabilities/features that the exchange supports
	cr.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerBatching:      true,
				TickerFetching:      true,
				TradeFetching:       true,
				KlineFetching:       true,
				OrderbookFetching:   true,
				CryptoWithdrawal:    true,
				AutoPairUpdates:     true,
				AccountInfo:         true,
				GetOrder:            true,
				GetOrders:           true,
				CancelOrder:         true,
				CancelOrders:        true,
				SubmitOrder:         true,
				SubmitOrders:        true,
				UserTradeHistory:    true,
				TradeFee:            true,
				CryptoWithdrawalFee: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerBatching:         true,
				TickerFetching:         true,
				KlineFetching:          true,
				OrderbookFetching:      true,
				AuthenticatedEndpoints: true,
				AccountInfo:            true,
				CryptoWithdrawal:       true,
				TradeFetching:          true,
				AccountBalance:         true,
				SubmitOrder:            true,
				SubmitOrders:           true,
				CancelOrder:            true,
				CancelOrders:           true,
				GetOrder:               true,
				GetOrders:              true,
				Subscribe:              true,
				Unsubscribe:            true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: kline.DeployExchangeIntervals(
					kline.IntervalCapacity{Interval: kline.OneMin},
					kline.IntervalCapacity{Interval: kline.FiveMin},
					kline.IntervalCapacity{Interval: kline.FifteenMin},
					kline.IntervalCapacity{Interval: kline.ThirtyMin},
					kline.IntervalCapacity{Interval: kline.OneHour},
					kline.IntervalCapacity{Interval: kline.FourHour},
					kline.IntervalCapacity{Interval: kline.SixHour},
					kline.IntervalCapacity{Interval: kline.TwelveHour},
					kline.IntervalCapacity{Interval: kline.OneDay},
					kline.IntervalCapacity{Interval: kline.SevenDay},
					kline.IntervalCapacity{Interval: kline.TwoWeek},
					kline.IntervalCapacity{Interval: kline.OneMonth},
				),
				GlobalResultLimit: 300,
			},
		},
	}
	cr.Requester, err = request.New(cr.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()),
	)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	cr.API.Endpoints = cr.NewEndpoints()
	err = cr.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:                   cryptodotcomAPIURL,
		exchange.WebsocketSpot:              cryptodotcomWebsocketMarketAPI,
		exchange.WebsocketSpotSupplementary: cryptodotcomWebsocketUserAPI,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	cr.Websocket = stream.New()
	cr.WebsocketResponseMaxLimit = time.Second * 15
	cr.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	cr.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (cr *Cryptodotcom) Setup(exch *config.Exchange) error {
	err := exch.Validate()
	if err != nil {
		return err
	}
	if !exch.Enabled {
		cr.SetEnabled(false)
		return nil
	}
	err = cr.SetupDefaults(exch)
	if err != nil {
		return err
	}
	wsRunningEndpoint, err := cr.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}
	err = cr.Websocket.Setup(
		&stream.WebsocketSetup{
			ExchangeConfig:         exch,
			DefaultURL:             cryptodotcomWebsocketUserAPI,
			RunningURL:             wsRunningEndpoint,
			Connector:              cr.WsConnect,
			Subscriber:             cr.Subscribe,
			Unsubscriber:           cr.Unsubscribe,
			ConnectionMonitorDelay: exch.ConnectionMonitorDelay,
			GenerateSubscriptions:  cr.GenerateDefaultSubscriptions,
			Features:               &cr.Features.Supports.WebsocketCapabilities,
		})
	if err != nil {
		return err
	}
	err = cr.Websocket.SetupNewConnection(stream.ConnectionSetup{
		URL:                  cryptodotcomWebsocketMarketAPI,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
	if err != nil {
		return err
	}
	return cr.Websocket.SetupNewConnection(stream.ConnectionSetup{
		URL:                  cryptodotcomWebsocketUserAPI,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
		Authenticated:        true,
	})
}

// Start starts the Cryptodotcom go routine
func (cr *Cryptodotcom) Start(_ context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return fmt.Errorf("%T %w", wg, common.ErrNilPointer)
	}
	wg.Add(1)
	go func() {
		cr.Run()
		wg.Done()
	}()
	return nil
}

// Run implements the Cryptodotcom wrapper
func (cr *Cryptodotcom) Run() {
	if cr.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			cr.Name,
			common.IsEnabled(cr.Websocket.IsEnabled()))
		cr.PrintEnabledPairs()
	}

	if !cr.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := cr.UpdateTradablePairs(context.TODO(), false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			cr.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (cr *Cryptodotcom) FetchTradablePairs(ctx context.Context, a asset.Item) (currency.Pairs, error) {
	if !cr.SupportsAsset(a) {
		return nil, fmt.Errorf("asset type of %s is not supported by %s", a, cr.Name)
	}
	instruments, err := cr.GetInstruments(ctx)
	if err != nil {
		return nil, err
	}
	pairs := make(currency.Pairs, len(instruments.Instruments))
	for x := range instruments.Instruments {
		pairs[x], err = currency.NewPairFromString(instruments.Instruments[x].InstrumentName)
		if err != nil {
			return nil, err
		}
	}
	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (cr *Cryptodotcom) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	pairs, err := cr.FetchTradablePairs(ctx, asset.Spot)
	if err != nil {
		return err
	}
	return cr.UpdatePairs(pairs, asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (cr *Cryptodotcom) UpdateTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w asset type: %v", asset.ErrNotSupported, assetType)
	}
	p, err := cr.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	tick, err := cr.GetTicker(ctx, p.String())
	if err != nil {
		return nil, err
	}
	if len(tick.Data) != 1 {
		return nil, errInvalidResponseFromServer
	}
	tickerPrice := &ticker.Price{
		High:         tick.Data[0].HighestTradePrice,
		Low:          tick.Data[0].LowestTradePrice,
		Bid:          tick.Data[0].BestBidPrice,
		Ask:          tick.Data[0].BestAskPrice,
		Open:         tick.Data[0].OpenInterest,
		Last:         tick.Data[0].LatestTradePrice,
		Volume:       tick.Data[0].TradedVolume,
		LastUpdated:  tick.Data[0].TradeTimestamp.Time(),
		AssetType:    assetType,
		ExchangeName: cr.Name,
		Pair:         p,
	}
	err = ticker.ProcessTicker(tickerPrice)
	if err != nil {
		return tickerPrice, err
	}
	return ticker.GetTicker(cr.Name, p, assetType)
}

// UpdateTickers updates all currency pairs of a given asset type
func (cr *Cryptodotcom) UpdateTickers(ctx context.Context, assetType asset.Item) error {
	if !cr.SupportsAsset(assetType) {
		return fmt.Errorf("%w asset type: %v", asset.ErrNotSupported, assetType)
	}
	tick, err := cr.GetTicker(ctx, "")
	if err != nil {
		return err
	}
	for y := range tick.Data {
		cp, err := currency.NewPairFromString(tick.Data[y].InstrumentName)
		if err != nil {
			return err
		}
		err = ticker.ProcessTicker(&ticker.Price{
			Last:         tick.Data[y].LatestTradePrice,
			High:         tick.Data[y].HighestTradePrice,
			Low:          tick.Data[y].LowestTradePrice,
			Bid:          tick.Data[y].BestBidPrice,
			Ask:          tick.Data[y].BestAskPrice,
			Volume:       tick.Data[y].TradedVolume,
			Open:         tick.Data[y].OpenInterest,
			Pair:         cp,
			ExchangeName: cr.Name,
			AssetType:    assetType,
			QuoteVolume:  tick.Data[y].TradedVolumeInUSD24H,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchTicker returns the ticker for a currency pair
func (cr *Cryptodotcom) FetchTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w, asset type: %v", asset.ErrNotSupported, assetType)
	}
	tickerNew, err := ticker.GetTicker(cr.Name, p, assetType)
	if err != nil {
		return cr.UpdateTicker(ctx, p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (cr *Cryptodotcom) FetchOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w, asset type: %v", asset.ErrNotSupported, assetType)
	}
	ob, err := orderbook.Get(cr.Name, pair, assetType)
	if err != nil {
		return cr.UpdateOrderbook(ctx, pair, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (cr *Cryptodotcom) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w, asset type: %v", asset.ErrNotSupported, assetType)
	}
	pair, err := cr.FormatExchangeCurrency(pair, assetType)
	if err != nil {
		return nil, err
	}
	orderbookNew, err := cr.GetOrderbook(ctx, pair.String(), 0)
	if err != nil {
		return nil, err
	}
	book := &orderbook.Base{
		Exchange:        cr.Name,
		Pair:            pair,
		Asset:           assetType,
		VerifyOrderbook: cr.CanVerifyOrderbook,
	}
	if len(orderbookNew.Data) == 0 {
		return nil, fmt.Errorf("%w, missing orderbook data", orderbook.ErrOrderbookInvalid)
	}
	book.Bids = make([]orderbook.Item, len(orderbookNew.Data[0].Bids))
	for x := range orderbookNew.Data[0].Bids {
		var price float64
		var amount float64
		price, err = strconv.ParseFloat(orderbookNew.Data[0].Bids[x][0], 64)
		if err != nil {
			return nil, err
		}
		amount, err = strconv.ParseFloat(orderbookNew.Data[0].Bids[x][1], 64)
		if err != nil {
			return nil, err
		}
		book.Bids[x] = orderbook.Item{
			Amount: amount,
			Price:  price,
		}
	}
	book.Asks = make([]orderbook.Item, len(orderbookNew.Data[0].Asks))
	for x := range orderbookNew.Data[0].Asks {
		var price float64
		var amount float64
		price, err = strconv.ParseFloat(orderbookNew.Data[0].Asks[x][0], 64)
		if err != nil {
			return nil, err
		}
		amount, err = strconv.ParseFloat(orderbookNew.Data[0].Asks[x][1], 64)
		if err != nil {
			return nil, err
		}
		book.Asks[x] = orderbook.Item{
			Amount: amount,
			Price:  price,
		}
	}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(cr.Name, pair, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (cr *Cryptodotcom) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings
	if !cr.SupportsAsset(assetType) {
		return info, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	var accs *Accounts
	var err error
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		accs, err = cr.WsRetriveAccountSummary(currency.EMPTYCODE)
	} else {
		accs, err = cr.GetAccountSummary(ctx, currency.EMPTYCODE)
	}
	if err != nil {
		return info, err
	}
	balances := make([]account.Balance, len(accs.Accounts))
	for i := range accs.Accounts {
		balances[i] = account.Balance{
			Currency: currency.NewCode(accs.Accounts[i].Currency),
			Total:    accs.Accounts[i].Balance,
			Hold:     accs.Accounts[i].Stake + accs.Accounts[i].Order,
			Free:     accs.Accounts[i].Available,
		}
	}
	acc := account.SubAccount{
		Currencies: balances,
		AssetType:  assetType,
	}
	info.Accounts = []account.SubAccount{acc}
	creds, err := cr.GetCredentials(ctx)
	if err != nil {
		return info, err
	}
	if err := account.Process(&info, creds); err != nil {
		return account.Holdings{}, err
	}
	info.Exchange = cr.Name
	return info, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (cr *Cryptodotcom) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	creds, err := cr.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	acc, err := account.GetHoldings(cr.Name, creds, assetType)
	if err != nil {
		return cr.UpdateAccountInfo(ctx, assetType)
	}
	return acc, nil
}

// GetAccountFundingHistory returns funding history, deposits and
// withdrawals
func (cr *Cryptodotcom) GetAccountFundingHistory(ctx context.Context) ([]exchange.FundingHistory, error) {
	var err error
	var withdrawals *WithdrawalResponse
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		withdrawals, err = cr.WsRetriveWithdrawalHistory()
	} else {
		withdrawals, err = cr.GetWithdrawalHistory(ctx)
	}
	if err != nil {
		return nil, err
	}
	deposits, err := cr.GetDepositHistory(ctx, currency.EMPTYCODE, time.Time{}, time.Time{}, 0, 0, 0)
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.FundingHistory, 0, len(withdrawals.WithdrawalList)+len(deposits.DepositList))
	for x := range withdrawals.WithdrawalList {
		resp = append(resp, exchange.FundingHistory{
			Status:          translateWithdrawalStatus(withdrawals.WithdrawalList[x].Status),
			Timestamp:       withdrawals.WithdrawalList[x].UpdateTime.Time(),
			Currency:        withdrawals.WithdrawalList[x].Currency,
			Amount:          withdrawals.WithdrawalList[x].Amount,
			TransferType:    "withdrawal",
			CryptoToAddress: withdrawals.WithdrawalList[x].Address,
			TransferID:      withdrawals.WithdrawalList[x].TransactionID,
			Fee:             withdrawals.WithdrawalList[x].Fee,
		})
	}
	for x := range deposits.DepositList {
		resp = append(resp, exchange.FundingHistory{
			ExchangeName:    cr.Name,
			Status:          translateDepositStatus(deposits.DepositList[x].Status),
			Timestamp:       deposits.DepositList[x].UpdateTime.Time(),
			Currency:        deposits.DepositList[x].Currency,
			Amount:          deposits.DepositList[x].Amount,
			TransferType:    "deposit",
			CryptoToAddress: deposits.DepositList[x].Address,
			CryptoTxID:      deposits.DepositList[x].ID,
		})
	}
	return resp, nil
}

// GetWithdrawalsHistory returns previous withdrawals data
func (cr *Cryptodotcom) GetWithdrawalsHistory(ctx context.Context, _ currency.Code, _ asset.Item) ([]exchange.WithdrawalHistory, error) {
	withdrawals, err := cr.GetWithdrawalHistory(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.WithdrawalHistory, len(withdrawals.WithdrawalList))
	for x := range withdrawals.WithdrawalList {
		resp[x] = exchange.WithdrawalHistory{
			Status:          translateWithdrawalStatus(withdrawals.WithdrawalList[x].Status),
			Timestamp:       withdrawals.WithdrawalList[x].UpdateTime.Time(),
			Currency:        withdrawals.WithdrawalList[x].Currency,
			Amount:          withdrawals.WithdrawalList[x].Amount,
			TransferType:    "withdrawal",
			CryptoToAddress: withdrawals.WithdrawalList[x].Address,
			TransferID:      withdrawals.WithdrawalList[x].TransactionID,
			Fee:             withdrawals.WithdrawalList[x].Fee,
		}
	}
	return resp, nil
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (cr *Cryptodotcom) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	p, err := cr.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	if !p.IsPopulated() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	trades, err := cr.GetTrades(ctx, p.String())
	if err != nil {
		return nil, err
	}
	resp := make([]trade.Data, len(trades.Data))
	for x := range trades.Data {
		var side order.Side
		side, err = order.StringToOrderSide(trades.Data[x].Side)
		if err != nil {
			return nil, err
		}
		resp[x] = trade.Data{
			TID:          trades.Data[x].TradeID,
			Exchange:     cr.Name,
			CurrencyPair: p,
			AssetType:    assetType,
			Side:         side,
			Price:        trades.Data[x].TradePrice,
			Amount:       trades.Data[x].TradeQuantity,
			Timestamp:    trades.Data[x].DataTime.Time(),
		}
	}
	if cr.IsSaveTradeDataEnabled() {
		err = trade.AddTradesToBuffer(cr.Name, resp...)
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(trade.ByDate(resp))
	return resp, nil
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (cr *Cryptodotcom) GetHistoricTrades(ctx context.Context, p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w, asset type %v not supported", asset.ErrNotSupported, assetType)
	}
	if err := common.StartEndTimeCheck(timestampStart, timestampEnd); err != nil {
		return nil, fmt.Errorf("invalid time range supplied. Start: %v End %v %w", timestampStart, timestampEnd, err)
	}
	var err error
	p, err = cr.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	limit := 1000
	ts := timestampStart
	var resp []trade.Data
allTrades:
	for {
		var tradeData *TradesResponse
		tradeData, err = cr.GetTrades(ctx, p.String())
		if err != nil {
			return nil, err
		}
		for i := range tradeData.Data {
			if tradeData.Data[i].TradeTimestamp.Time().Before(timestampStart) || tradeData.Data[i].TradeTimestamp.Time().After(timestampEnd) {
				break allTrades
			}
			var side order.Side
			side, err = order.StringToOrderSide(tradeData.Data[i].Side)
			if err != nil {
				return nil, err
			}
			if tradeData.Data[i].TradePrice == 0 {
				continue
			}
			resp = append(resp, trade.Data{
				Exchange:     cr.Name,
				CurrencyPair: p,
				AssetType:    assetType,
				Side:         side,
				Price:        tradeData.Data[i].TradePrice,
				Amount:       tradeData.Data[i].TradeQuantity,
				Timestamp:    tradeData.Data[i].TradeTimestamp.Time(),
				TID:          tradeData.Data[i].TradeID,
			})
			if i == len(tradeData.Data)-1 {
				if ts.Equal(tradeData.Data[i].TradeTimestamp.Time()) {
					// reached end of trades to crawl
					break allTrades
				}
				ts = tradeData.Data[i].TradeTimestamp.Time()
			}
		}
		if len(tradeData.Data) != limit {
			break allTrades
		}
	}
	err = cr.AddTradesToBuffer(resp...)
	if err != nil {
		return nil, err
	}

	sort.Sort(trade.ByDate(resp))
	return trade.FilterTradesByTime(resp, timestampStart, timestampEnd), nil
}

// SubmitOrder submits a new order
func (cr *Cryptodotcom) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if !cr.SupportsAsset(s.AssetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, s.AssetType)
	}
	if s.Amount <= 0 {
		return nil, fmt.Errorf("amount, or size (sz) of quantity to buy or sell hast to be greater than zero ")
	}
	format, err := cr.GetPairFormat(s.AssetType, false)
	if err != nil {
		return nil, err
	}
	if !s.Pair.IsPopulated() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var notional float64
	switch s.Type {
	case order.Market, order.StopLoss, order.TakeProfit:
		// For MARKET (BUY), STOP_LOSS (BUY), TAKE_PROFIT (BUY) orders only: Amount to spend
		notional = s.Amount
	}
	var ordersResp *CreateOrderResponse
	arg := &CreateOrderParam{InstrumentName: format.Format(s.Pair), Side: s.Side, OrderType: orderTypeToString(s.Type), Price: s.Price, Quantity: s.Amount, ClientOrderID: s.ClientOrderID, Notional: notional, PostOnly: s.PostOnly, TriggerPrice: s.TriggerPrice}
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		ordersResp, err = cr.WsPlaceOrder(arg)
	} else {
		ordersResp, err = cr.CreateOrder(ctx, arg)
	}
	if err != nil {
		return nil, err
	}
	return s.DeriveSubmitResponse(ordersResp.OrderID)
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (cr *Cryptodotcom) ModifyOrder(_ context.Context, _ *order.Modify) (*order.ModifyResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// CancelOrder cancels an order by its corresponding ID number
func (cr *Cryptodotcom) CancelOrder(ctx context.Context, ord *order.Cancel) error {
	err := ord.Validate(ord.StandardCancel())
	if err != nil {
		return err
	}
	if !cr.SupportsAsset(ord.AssetType) {
		return fmt.Errorf("%w: %v", asset.ErrNotSupported, ord.AssetType)
	}
	format, err := cr.GetPairFormat(ord.AssetType, false)
	if err != nil {
		return err
	}
	if !ord.Pair.IsPopulated() {
		return currency.ErrCurrencyPairEmpty
	}
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		return cr.WsCancelExistingOrder(format.Format(ord.Pair), ord.OrderID)
	}
	return cr.CancelExistingOrder(ctx, format.Format(ord.Pair), ord.OrderID)
}

// CancelBatchOrders cancels orders by their corresponding ID numbers
func (cr *Cryptodotcom) CancelBatchOrders(ctx context.Context, orders []order.Cancel) (*order.CancelBatchResponse, error) {
	cancelBatchResponse := &order.CancelBatchResponse{
		Status: map[string]string{},
	}
	cancelOrderParams := []CancelOrderParam{}
	format, err := cr.GetPairFormat(asset.Spot, true)
	if err != nil {
		return nil, err
	}
	for x := range orders {
		cancelOrderParams = append(cancelOrderParams, CancelOrderParam{
			InstrumentName: format.Format(orders[x].Pair),
			OrderID:        orders[x].OrderID,
		})
	}
	var cancelResp *CancelOrdersResponse
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		cancelResp, err = cr.WsCancelOrderList(cancelOrderParams)
	} else {
		cancelResp, err = cr.CancelOrderList(ctx, cancelOrderParams)
	}
	if err != nil {
		return nil, err
	}
	for x := range cancelResp.ResultList {
		if cancelResp.ResultList[x].Code != 0 {
			cancelBatchResponse.Status[cancelOrderParams[cancelResp.ResultList[x].Index].InstrumentName] = ""
		} else {
			cancelBatchResponse.Status[cancelOrderParams[cancelResp.ResultList[x].Index].InstrumentName] = order.Cancelled.String()
		}
	}
	return cancelBatchResponse, nil
}

// CancelAllOrders cancels all orders associated with a currency pair
func (cr *Cryptodotcom) CancelAllOrders(ctx context.Context, orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	cancelAllResponse := order.CancelAllResponse{
		Status: map[string]string{},
	}
	err := orderCancellation.Validate()
	if err != nil {
		return cancelAllResponse, err
	}
	format, err := cr.GetPairFormat(orderCancellation.AssetType, true)
	if err != nil {
		return cancelAllResponse, err
	}
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		return order.CancelAllResponse{}, cr.WsCancelAllPersonalOrders(orderCancellation.Pair.Format(format).String())
	}
	return order.CancelAllResponse{}, cr.CancelAllPersonalOrders(ctx, orderCancellation.Pair.Format(format).String())
}

// GetOrderInfo returns order information based on order ID
func (cr *Cryptodotcom) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (*order.Detail, error) {
	if !cr.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	if !pair.IsPopulated() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var orderDetail *OrderDetail
	var err error
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		orderDetail, err = cr.WsRetriveOrderDetail(orderID)
	} else {
		orderDetail, err = cr.GetOrderDetail(ctx, orderID)
	}
	if err != nil {
		return nil, err
	}
	status, err := order.StringToOrderStatus(orderDetail.OrderInfo.Status)
	if err != nil {
		return nil, err
	}
	orderType, err := stringToOrderType(orderDetail.OrderInfo.Type)
	if err != nil {
		return nil, err
	}
	side, err := order.StringToOrderSide(orderDetail.OrderInfo.Side)
	if err != nil {
		return nil, err
	}
	pair, err = cr.FormatExchangeCurrency(pair, asset.Spot)
	if err != nil {
		return nil, err
	}
	return &order.Detail{
		Amount:         orderDetail.OrderInfo.Quantity,
		Exchange:       cr.Name,
		OrderID:        orderDetail.OrderInfo.OrderID,
		ClientOrderID:  orderDetail.OrderInfo.ClientOid,
		Side:           side,
		Type:           orderType,
		Pair:           pair,
		Cost:           orderDetail.OrderInfo.CumulativeValue,
		AssetType:      assetType,
		Status:         status,
		Price:          orderDetail.OrderInfo.Price,
		ExecutedAmount: orderDetail.OrderInfo.CumulativeQuantity - orderDetail.OrderInfo.Quantity,
		Date:           orderDetail.OrderInfo.CreateTime.Time(),
		LastUpdated:    orderDetail.OrderInfo.UpdateTime.Time(),
	}, err
}

// GetDepositAddress returns a deposit address for a specified currency
func (cr *Cryptodotcom) GetDepositAddress(ctx context.Context, c currency.Code, accountID, chain string) (*deposit.Address, error) {
	dAddresses, err := cr.GetPersonalDepositAddress(ctx, c)
	if err != nil {
		return nil, err
	}
	for x := range dAddresses.DepositAddressList {
		if dAddresses.DepositAddressList[x].Currency == c.String() &&
			(accountID == "" || accountID == dAddresses.DepositAddressList[x].ID) &&
			(chain == "" || chain == dAddresses.DepositAddressList[x].Network) {
			return &deposit.Address{
				Address: dAddresses.DepositAddressList[x].Address,
				Chain:   dAddresses.DepositAddressList[x].Network,
			}, nil
		}
	}
	return nil, fmt.Errorf("deposit address not found for currency: %s chain: %s", c, chain)
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (cr *Cryptodotcom) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	err := withdrawRequest.Validate()
	if err != nil {
		return nil, err
	}
	var withdrawalResp *WithdrawalItem
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		withdrawalResp, err = cr.WsCreateWithdrawal(withdrawRequest.Currency, withdrawRequest.Amount, withdrawRequest.Crypto.Address, withdrawRequest.Crypto.AddressTag, withdrawRequest.Crypto.Chain, withdrawRequest.ClientOrderID)
	} else {
		withdrawalResp, err = cr.WithdrawFunds(ctx, withdrawRequest.Currency, withdrawRequest.Amount, withdrawRequest.Crypto.Address, withdrawRequest.Crypto.AddressTag, withdrawRequest.Crypto.Chain, withdrawRequest.ClientOrderID)
	}
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		ID:     withdrawalResp.ID,
		Name:   cr.Name,
		Status: withdrawalResp.Status,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (cr *Cryptodotcom) WithdrawFiatFunds(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (cr *Cryptodotcom) WithdrawFiatFundsToInternationalBank(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetActiveOrders retrieves any orders that are active/open
func (cr *Cryptodotcom) GetActiveOrders(ctx context.Context, getOrdersRequest *order.MultiOrderRequest) (order.FilteredOrders, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}
	var orders *PersonalOrdersResponse
	var err error
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		orders, err = cr.WsRetrivePersonalOpenOrders("", 0, 0)
	} else {
		orders, err = cr.GetPersonalOpenOrders(ctx, "", 0, 0)
	}
	if err != nil {
		return nil, err
	}
	pairFormat, err := cr.GetPairFormat(getOrdersRequest.AssetType, false)
	if err != nil {
		return nil, err
	}
	resp := []order.Detail{}
	for x := range orders.OrderList {
		cp, err := currency.NewPairFromString(orders.OrderList[x].InstrumentName)
		if err != nil {
			return nil, err
		}
		if len(orders.OrderList) != 0 {
			found := false
			for b := range getOrdersRequest.Pairs {
				if cp.Equal(getOrdersRequest.Pairs[b].Format(pairFormat)) {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		orderType, err := stringToOrderType(orders.OrderList[x].Type)
		if err != nil {
			return nil, err
		}
		side, err := order.StringToOrderSide(orders.OrderList[x].Side)
		if err != nil {
			return nil, err
		}
		status, err := order.StringToOrderStatus(orders.OrderList[x].Status)
		if err != nil {
			return nil, err
		}
		resp = append(resp, order.Detail{
			Price:                orders.OrderList[x].Price,
			AverageExecutedPrice: orders.OrderList[x].AvgPrice,
			Amount:               orders.OrderList[x].CumulativeQuantity,
			ExecutedAmount:       orders.OrderList[x].Quantity,
			RemainingAmount:      orders.OrderList[x].CumulativeQuantity - orders.OrderList[x].Quantity,
			Exchange:             cr.Name,
			OrderID:              orders.OrderList[x].OrderID,
			ClientOrderID:        orders.OrderList[x].ClientOid,
			Status:               status,
			Side:                 side,
			Type:                 orderType,
			AssetType:            getOrdersRequest.AssetType,
			Date:                 orders.OrderList[x].CreateTime.Time(),
			LastUpdated:          orders.OrderList[x].UpdateTime.Time(),
			Pair:                 cp,
		})
	}
	return getOrdersRequest.Filter(cr.Name, resp), nil
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (cr *Cryptodotcom) GetOrderHistory(ctx context.Context, getOrdersRequest *order.MultiOrderRequest) (order.FilteredOrders, error) {
	if err := getOrdersRequest.Validate(); err != nil {
		return nil, err
	}
	pairFormat, err := cr.GetPairFormat(getOrdersRequest.AssetType, false)
	if err != nil {
		return nil, err
	}
	var orders *PersonalOrdersResponse
	if cr.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		orders, err = cr.WsRetrivePersonalOrderHistory("", getOrdersRequest.StartTime, getOrdersRequest.EndTime, 0, 0)
	} else {
		orders, err = cr.GetPersonalOrderHistory(ctx, "", getOrdersRequest.StartTime, getOrdersRequest.EndTime, 0, 0)
	}
	if err != nil {
		return nil, err
	}
	resp := []order.Detail{}
	for x := range orders.OrderList {
		cp, err := currency.NewPairFromString(orders.OrderList[x].InstrumentName)
		if err != nil {
			return nil, err
		}
		if len(orders.OrderList) != 0 {
			found := false
			for b := range getOrdersRequest.Pairs {
				if cp.Equal(getOrdersRequest.Pairs[b].Format(pairFormat)) {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		orderType, err := stringToOrderType(orders.OrderList[x].Type)
		if err != nil {
			return nil, err
		}
		side, err := order.StringToOrderSide(orders.OrderList[x].Side)
		if err != nil {
			return nil, err
		}
		status, err := order.StringToOrderStatus(orders.OrderList[x].Status)
		if err != nil {
			return nil, err
		}
		resp = append(resp, order.Detail{
			Price:                orders.OrderList[x].Price,
			AverageExecutedPrice: orders.OrderList[x].AvgPrice,
			Amount:               orders.OrderList[x].CumulativeQuantity,
			ExecutedAmount:       orders.OrderList[x].Quantity,
			RemainingAmount:      orders.OrderList[x].CumulativeQuantity - orders.OrderList[x].Quantity,
			Exchange:             cr.Name,
			OrderID:              orders.OrderList[x].OrderID,
			ClientOrderID:        orders.OrderList[x].ClientOid,
			Status:               status,
			Side:                 side,
			Type:                 orderType,
			AssetType:            getOrdersRequest.AssetType,
			Date:                 orders.OrderList[x].CreateTime.Time(),
			LastUpdated:          orders.OrderList[x].UpdateTime.Time(),
			Pair:                 cp,
		})
	}
	return getOrdersRequest.Filter(cr.Name, resp), nil
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (cr *Cryptodotcom) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	if feeBuilder == nil {
		return 0, fmt.Errorf("%T %w", feeBuilder, common.ErrNilPointer)
	}
	if !cr.AreCredentialsValid(ctx) &&
		feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = calculateTradingFee(feeBuilder) * feeBuilder.Amount * feeBuilder.PurchasePrice
	case exchange.CryptocurrencyWithdrawalFee:
		fee = 0.5 * feeBuilder.PurchasePrice * feeBuilder.Amount
	case exchange.OfflineTradeFee:
		fee = getOfflineTradeFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}
	return fee, nil
}

// getOfflineTradeFee calculates the worst case-scenario trading fee
func getOfflineTradeFee(price, amount float64) float64 {
	return 0.0750 * price * amount
}

// calculateTradingFee return fee based on users current fee tier or default values
func calculateTradingFee(feeBuilder *exchange.FeeBuilder) float64 {
	switch {
	case feeBuilder.Amount*feeBuilder.PurchasePrice <= 250:
		return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.075
	case feeBuilder.Amount*feeBuilder.PurchasePrice < 1000000:
		if feeBuilder.IsMaker {
			return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.07
		}
		return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.072
	case feeBuilder.Amount*feeBuilder.PurchasePrice < 5000000:
		if feeBuilder.IsMaker {
			return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.065
		}
		return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.069
	case feeBuilder.Amount*feeBuilder.PurchasePrice <= 10000000:
		if feeBuilder.IsMaker {
			return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.06
		}
		return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.065
	default:
		if !feeBuilder.IsMaker {
			return feeBuilder.PurchasePrice * feeBuilder.Amount * 0.05
		}
		return 0
	}
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (cr *Cryptodotcom) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := cr.GetKlineRequest(pair, a, interval, start, end, false)
	if err != nil {
		return nil, err
	}
	candles, err := cr.GetCandlestickDetail(ctx, req.RequestFormatted.String(), interval)
	if err != nil {
		return nil, err
	}
	candleElements := make([]kline.Candle, len(candles.Data))
	for x := range candles.Data {
		candleElements[x] = kline.Candle{
			Time:   candles.Data[x].EndTime.Time(),
			Open:   candles.Data[x].Open,
			High:   candles.Data[x].High,
			Low:    candles.Data[x].Low,
			Close:  candles.Data[x].Close,
			Volume: candles.Data[x].Volume,
		}
	}
	return req.ProcessResponse(candleElements)
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (cr *Cryptodotcom) GetHistoricCandlesExtended(_ context.Context, _ currency.Pair, _ asset.Item, _ kline.Interval, _, _ time.Time) (*kline.Item, error) {
	return nil, common.ErrFunctionNotSupported
}

// ValidateAPICredentials validates current credentials used for wrapper
func (cr *Cryptodotcom) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	_, err := cr.UpdateAccountInfo(ctx, assetType)
	return cr.CheckTransientError(err)
}

// GetServerTime returns the current exchange server time.
func (cr *Cryptodotcom) GetServerTime(_ context.Context, _ asset.Item) (time.Time, error) {
	return time.Time{}, common.ErrFunctionNotSupported
}
