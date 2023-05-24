package okx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/margin"
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

const (
	okxWebsocketResponseMaxLimit = time.Second * 3
)

// GetDefaultConfig returns a default exchange config
func (ok *Okx) GetDefaultConfig(ctx context.Context) (*config.Exchange, error) {
	ok.SetDefaults()
	exchCfg := new(config.Exchange)
	exchCfg.Name = ok.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = ok.BaseCurrencies

	err := ok.Setup(exchCfg)
	if err != nil {
		return nil, err
	}

	if ok.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := ok.UpdateTradablePairs(ctx, false)
		if err != nil {
			return nil, err
		}
	}
	return exchCfg, nil
}

// SetDefaults sets the basic defaults for Okx
func (ok *Okx) SetDefaults() {
	ok.Name = "Okx"
	ok.Enabled = true
	ok.Verbose = true

	ok.WsRequestSemaphore = make(chan int, 20)
	ok.API.CredentialsValidator.RequiresKey = true
	ok.API.CredentialsValidator.RequiresSecret = true
	ok.API.CredentialsValidator.RequiresClientID = true

	cpf := &currency.PairFormat{
		Delimiter: currency.DashDelimiter,
		Uppercase: true,
	}

	fmt1 := currency.PairStore{
		RequestFormat: cpf,
		ConfigFormat:  cpf,
	}

	err := ok.SetGlobalPairsManager(fmt1.RequestFormat, fmt1.ConfigFormat, asset.Spot, asset.Futures, asset.PerpetualSwap, asset.Options, asset.Margin)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	// Fill out the capabilities/features that the exchange supports
	ok.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:        true,
				OrderbookFetching:     true,
				AutoPairUpdates:       true,
				AccountInfo:           true,
				CryptoDeposit:         true,
				CryptoWithdrawalFee:   true,
				CryptoWithdrawal:      true,
				TradeFee:              true,
				SubmitOrder:           true,
				GetOrder:              true,
				GetOrders:             true,
				CancelOrder:           true,
				CancelOrders:          true,
				TradeFetching:         true,
				UserTradeHistory:      true,
				MultiChainDeposits:    true,
				MultiChainWithdrawals: true,
				KlineFetching:         true,
				DepositHistory:        true,
				WithdrawalHistory:     true,
				ModifyOrder:           true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:         true,
				OrderbookFetching:      true,
				Subscribe:              true,
				Unsubscribe:            true,
				AuthenticatedEndpoints: true,
				AccountInfo:            true,
				GetOrders:              true,
				TradeFetching:          true,
				KlineFetching:          true,
				GetOrder:               true,
				SubmitOrder:            true,
				CancelOrder:            true,
				CancelOrders:           true,
				ModifyOrder:            true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: kline.DeployExchangeIntervals(
					kline.IntervalCapacity{Interval: kline.OneMin},
					kline.IntervalCapacity{Interval: kline.ThreeMin},
					kline.IntervalCapacity{Interval: kline.FiveMin},
					kline.IntervalCapacity{Interval: kline.FifteenMin},
					kline.IntervalCapacity{Interval: kline.ThirtyMin},
					kline.IntervalCapacity{Interval: kline.OneHour},
					kline.IntervalCapacity{Interval: kline.TwoHour},
					kline.IntervalCapacity{Interval: kline.FourHour},
					kline.IntervalCapacity{Interval: kline.SixHour},
					kline.IntervalCapacity{Interval: kline.TwelveHour},
					kline.IntervalCapacity{Interval: kline.OneDay},
					kline.IntervalCapacity{Interval: kline.TwoDay},
					kline.IntervalCapacity{Interval: kline.ThreeDay},
					kline.IntervalCapacity{Interval: kline.FiveDay},
					kline.IntervalCapacity{Interval: kline.OneWeek},
					kline.IntervalCapacity{Interval: kline.OneMonth},
					kline.IntervalCapacity{Interval: kline.ThreeMonth},
					kline.IntervalCapacity{Interval: kline.SixMonth},
					kline.IntervalCapacity{Interval: kline.OneYear},
				),
				GlobalResultLimit: 100, // Reference: https://www.okx.com/docs-v5/en/#rest-api-market-data-get-candlesticks-history
			},
		},
	}
	ok.Requester, err = request.New(ok.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout),
		request.WithLimiter(SetRateLimit()))
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	ok.API.Endpoints = ok.NewEndpoints()
	err = ok.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot:      okxAPIURL,
		exchange.WebsocketSpot: okxAPIWebsocketPublicURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	ok.Websocket = stream.New()
	ok.WebsocketResponseMaxLimit = okxWebsocketResponseMaxLimit
	ok.WebsocketResponseCheckTimeout = okxWebsocketResponseMaxLimit
	ok.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (ok *Okx) Setup(exch *config.Exchange) error {
	err := exch.Validate()
	if err != nil {
		return err
	}
	if !exch.Enabled {
		ok.SetEnabled(false)
		return nil
	}
	err = ok.SetupDefaults(exch)
	if err != nil {
		return err
	}

	ok.WsResponseMultiplexer = wsRequestDataChannelsMultiplexer{
		WsResponseChannelsMap: make(map[string]*wsRequestInfo),
		Register:              make(chan *wsRequestInfo),
		Unregister:            make(chan string),
		Message:               make(chan *wsIncomingData),
	}

	wsRunningEndpoint, err := ok.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}
	err = ok.Websocket.Setup(&stream.WebsocketSetup{
		ExchangeConfig:         exch,
		DefaultURL:             okxAPIWebsocketPublicURL,
		RunningURL:             wsRunningEndpoint,
		Connector:              ok.WsConnect,
		Subscriber:             ok.Subscribe,
		Unsubscriber:           ok.Unsubscribe,
		GenerateSubscriptions:  ok.GenerateDefaultSubscriptions,
		ConnectionMonitorDelay: exch.ConnectionMonitorDelay,
		Features:               &ok.Features.Supports.WebsocketCapabilities,
		OrderbookBufferConfig: buffer.Config{
			Checksum: ok.CalculateUpdateOrderbookChecksum,
		},
	})

	if err != nil {
		return err
	}
	err = ok.Websocket.SetupNewConnection(stream.ConnectionSetup{
		URL:                  okxAPIWebsocketPublicURL,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     okxWebsocketResponseMaxLimit,
		RateLimit:            500,
	})
	if err != nil {
		return err
	}
	return ok.Websocket.SetupNewConnection(stream.ConnectionSetup{
		URL:                  okxAPIWebsocketPrivateURL,
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     okxWebsocketResponseMaxLimit,
		Authenticated:        true,
		RateLimit:            500,
	})
}

// Start starts the Okx go routine
func (ok *Okx) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return fmt.Errorf("%T %w", wg, common.ErrNilPointer)
	}
	wg.Add(1)
	go func() {
		ok.Run(ctx)
		wg.Done()
	}()
	return nil
}

// Run implements the Okx wrapper
func (ok *Okx) Run(ctx context.Context) {
	if ok.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			ok.Name,
			common.IsEnabled(ok.Websocket.IsEnabled()))
		ok.PrintEnabledPairs()
	}

	assetTypes := ok.GetAssetTypes(false)
	for i := range assetTypes {
		if err := ok.UpdateOrderExecutionLimits(ctx, assetTypes[i]); err != nil {
			log.Errorf(log.ExchangeSys,
				"%s failed to set exchange order execution limits. Err: %v",
				ok.Name,
				err)
		}
	}

	if ok.GetEnabledFeatures().AutoPairUpdates {
		if err := ok.UpdateTradablePairs(ctx, false); err != nil {
			log.Errorf(log.ExchangeSys,
				"%s failed to update tradable pairs. Err: %s",
				ok.Name,
				err)
		}
	}
}

// GetServerTime returns the current exchange server time.
func (ok *Okx) GetServerTime(ctx context.Context, _ asset.Item) (time.Time, error) {
	return ok.GetSystemTime(ctx)
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (ok *Okx) FetchTradablePairs(ctx context.Context, a asset.Item) (currency.Pairs, error) {
	insts, err := ok.getInstrumentsForAsset(ctx, a)
	if err != nil {
		return nil, err
	}
	if len(insts) == 0 {
		return nil, errNoInstrumentFound
	}
	pairs := make([]currency.Pair, len(insts))
	for x := range insts {
		pairs[x], err = currency.NewPairFromString(insts[x].InstrumentID)
		if err != nil {
			return nil, err
		}
	}
	return pairs, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores them in the exchanges config
func (ok *Okx) UpdateTradablePairs(ctx context.Context, forceUpdate bool) error {
	assetTypes := ok.GetAssetTypes(false)
	for i := range assetTypes {
		pairs, err := ok.FetchTradablePairs(ctx, assetTypes[i])
		if err != nil {
			return err
		}
		err = ok.UpdatePairs(pairs, assetTypes[i], false, forceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateOrderExecutionLimits sets exchange execution order limits for an asset type
func (ok *Okx) UpdateOrderExecutionLimits(ctx context.Context, a asset.Item) error {
	insts, err := ok.getInstrumentsForAsset(ctx, a)
	if err != nil {
		return err
	}
	if len(insts) == 0 {
		return errNoInstrumentFound
	}
	limits := make([]order.MinMaxLevel, len(insts))
	for x := range insts {
		pair, err := currency.NewPairFromString(insts[x].InstrumentID)
		if err != nil {
			return err
		}

		limits[x] = order.MinMaxLevel{
			Pair:                   pair,
			Asset:                  a,
			PriceStepIncrementSize: insts[x].TickSize.Float64(),
			MinAmount:              insts[x].MinimumOrderSize.Float64(),
		}
	}

	return ok.LoadLimits(limits)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (ok *Okx) UpdateTicker(ctx context.Context, p currency.Pair, a asset.Item) (*ticker.Price, error) {
	format, err := ok.GetPairFormat(a, true)
	if err != nil {
		return nil, err
	}
	if !p.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	instrumentID := format.Format(p)
	if !ok.SupportsAsset(a) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, a)
	}
	mdata, err := ok.GetTicker(ctx, instrumentID)
	if err != nil {
		return nil, err
	}
	var baseVolume, quoteVolume float64
	switch a {
	case asset.Spot, asset.Margin:
		baseVolume = mdata.Vol24H.Float64()
		quoteVolume = mdata.VolCcy24H.Float64()
	case asset.PerpetualSwap, asset.Futures, asset.Options:
		baseVolume = mdata.VolCcy24H.Float64()
		quoteVolume = mdata.Vol24H.Float64()
	default:
		return nil, fmt.Errorf("%w, asset type %s is not supported", errInvalidInstrumentType, a.String())
	}
	err = ticker.ProcessTicker(&ticker.Price{
		Last:         mdata.LastTradePrice.Float64(),
		High:         mdata.High24H.Float64(),
		Low:          mdata.Low24H.Float64(),
		Bid:          mdata.BidPrice.Float64(),
		Ask:          mdata.BestAskPrice.Float64(),
		Volume:       baseVolume,
		QuoteVolume:  quoteVolume,
		Open:         mdata.Open24H.Float64(),
		Pair:         p,
		ExchangeName: ok.Name,
		AssetType:    a,
	})
	if err != nil {
		return nil, err
	}
	return ticker.GetTicker(ok.Name, p, a)
}

// UpdateTickers updates all currency pairs of a given asset type
func (ok *Okx) UpdateTickers(ctx context.Context, assetType asset.Item) error {
	pairs, err := ok.GetEnabledPairs(assetType)
	if err != nil {
		return err
	}
	instrumentType := ok.GetInstrumentTypeFromAssetItem(assetType)
	ticks, err := ok.GetTickers(ctx, strings.ToUpper(instrumentType), "", "")
	if err != nil {
		return err
	}
	for y := range ticks {
		pair, err := ok.GetPairFromInstrumentID(ticks[y].InstrumentID)
		if err != nil {
			return err
		}
		for i := range pairs {
			pairFmt, err := ok.FormatExchangeCurrency(pairs[i], assetType)
			if err != nil {
				return err
			}
			if !pair.Equal(pairFmt) {
				continue
			}
			err = ticker.ProcessTicker(&ticker.Price{
				Last:         ticks[y].LastTradePrice.Float64(),
				High:         ticks[y].High24H.Float64(),
				Low:          ticks[y].Low24H.Float64(),
				Bid:          ticks[y].BidPrice.Float64(),
				Ask:          ticks[y].BestAskPrice.Float64(),
				Volume:       ticks[y].Vol24H.Float64(),
				QuoteVolume:  ticks[y].VolCcy24H.Float64(),
				Open:         ticks[y].Open24H.Float64(),
				Pair:         pairFmt,
				ExchangeName: ok.Name,
				AssetType:    assetType,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchTicker returns the ticker for a currency pair
func (ok *Okx) FetchTicker(ctx context.Context, p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	formattedPair, err := ok.FormatExchangeCurrency(p, assetType)
	if err != nil {
		return nil, err
	}
	tickerNew, err := ticker.GetTicker(ok.Name, formattedPair, assetType)
	if err != nil {
		return ok.UpdateTicker(ctx, p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (ok *Okx) FetchOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(ok.Name, pair, assetType)
	if err != nil {
		return ok.UpdateOrderbook(ctx, pair, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (ok *Okx) UpdateOrderbook(ctx context.Context, pair currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        ok.Name,
		Pair:            pair,
		Asset:           assetType,
		VerifyOrderbook: ok.CanVerifyOrderbook,
	}
	var orderbookNew *OrderBookResponse
	var err error
	if !ok.SupportsAsset(assetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	var instrumentID string
	format, err := ok.GetPairFormat(assetType, true)
	if err != nil {
		return nil, err
	}
	if !pair.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	instrumentID = format.Format(pair)
	orderbookNew, err = ok.GetOrderBookDepth(ctx, instrumentID, 0)
	if err != nil {
		return book, err
	}

	orderBookD, err := orderbookNew.GetOrderBookResponseDetail()
	if err != nil {
		return nil, err
	}
	book.Bids = make(orderbook.Items, len(orderBookD.Bids))
	for x := range orderBookD.Bids {
		book.Bids[x] = orderbook.Item{
			Amount: orderBookD.Bids[x].BaseCurrencies,
			Price:  orderBookD.Bids[x].DepthPrice,
		}
	}
	book.Asks = make(orderbook.Items, len(orderBookD.Asks))
	for x := range orderBookD.Asks {
		book.Asks[x] = orderbook.Item{
			Amount: orderBookD.Asks[x].NumberOfContracts,
			Price:  orderBookD.Asks[x].DepthPrice,
		}
	}
	err = book.Process()
	if err != nil {
		return book, err
	}
	return orderbook.Get(ok.Name, pair, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies.
func (ok *Okx) UpdateAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	var info account.Holdings
	var acc account.SubAccount
	info.Exchange = ok.Name
	if !ok.SupportsAsset(assetType) {
		return info, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	accountBalances, err := ok.GetNonZeroBalances(ctx, "")
	if err != nil {
		return info, err
	}
	currencyBalances := []account.Balance{}
	for i := range accountBalances {
		for j := range accountBalances[i].Details {
			currencyBalances = append(currencyBalances, account.Balance{
				Currency: currency.NewCode(accountBalances[i].Details[j].Currency),
				Total:    accountBalances[i].Details[j].EquityOfCurrency.Float64(),
				Hold:     accountBalances[i].Details[j].FrozenBalance.Float64(),
				Free:     accountBalances[i].Details[j].AvailableBalance.Float64(),
			})
		}
	}
	acc.Currencies = currencyBalances
	acc.AssetType = assetType
	info.Accounts = append(info.Accounts, acc)
	creds, err := ok.GetCredentials(ctx)
	if err != nil {
		return info, err
	}
	if err := account.Process(&info, creds); err != nil {
		return account.Holdings{}, err
	}
	return info, nil
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (ok *Okx) FetchAccountInfo(ctx context.Context, assetType asset.Item) (account.Holdings, error) {
	creds, err := ok.GetCredentials(ctx)
	if err != nil {
		return account.Holdings{}, err
	}
	acc, err := account.GetHoldings(ok.Name, creds, assetType)
	if err != nil {
		return ok.UpdateAccountInfo(ctx, assetType)
	}
	return acc, nil
}

// GetFundingHistory returns funding history, deposits and withdrawals
func (ok *Okx) GetFundingHistory(ctx context.Context) ([]exchange.FundHistory, error) {
	depositHistories, err := ok.GetCurrencyDepositHistory(ctx, "", "", "", time.Time{}, time.Time{}, -1, 0)
	if err != nil {
		return nil, err
	}
	withdrawalHistories, err := ok.GetWithdrawalHistory(ctx, "", "", "", "", "", time.Time{}, time.Time{}, -5)
	if err != nil {
		return nil, err
	}
	resp := make([]exchange.FundHistory, 0, len(depositHistories)+len(withdrawalHistories))
	for x := range depositHistories {
		resp = append(resp, exchange.FundHistory{
			ExchangeName:    ok.Name,
			Status:          strconv.Itoa(depositHistories[x].State),
			Timestamp:       depositHistories[x].Timestamp.Time(),
			Currency:        depositHistories[x].Currency,
			Amount:          depositHistories[x].Amount,
			TransferType:    "deposit",
			CryptoToAddress: depositHistories[x].ToDepositAddress,
			CryptoTxID:      depositHistories[x].TransactionID,
		})
	}
	for x := range withdrawalHistories {
		resp = append(resp, exchange.FundHistory{
			ExchangeName:    ok.Name,
			Status:          withdrawalHistories[x].StateOfWithdrawal,
			Timestamp:       withdrawalHistories[x].Timestamp.Time(),
			Currency:        withdrawalHistories[x].Currency,
			Amount:          withdrawalHistories[x].Amount,
			TransferType:    "withdrawal",
			CryptoToAddress: withdrawalHistories[x].ToReceivingAddress,
			CryptoTxID:      withdrawalHistories[x].TransactionID,
			TransferID:      withdrawalHistories[x].WithdrawalID,
			Fee:             withdrawalHistories[x].WithdrawalFee,
			CryptoChain:     withdrawalHistories[x].ChainName,
		})
	}
	return resp, nil
}

// GetWithdrawalsHistory returns previous withdrawals data
func (ok *Okx) GetWithdrawalsHistory(ctx context.Context, c currency.Code, _ asset.Item) (resp []exchange.WithdrawalHistory, err error) {
	withdrawals, err := ok.GetWithdrawalHistory(ctx, c.String(), "", "", "", "", time.Time{}, time.Time{}, -5)
	if err != nil {
		return nil, err
	}
	resp = make([]exchange.WithdrawalHistory, 0, len(withdrawals))
	for x := range withdrawals {
		resp = append(resp, exchange.WithdrawalHistory{
			Status:          withdrawals[x].StateOfWithdrawal,
			Timestamp:       withdrawals[x].Timestamp.Time(),
			Currency:        withdrawals[x].Currency,
			Amount:          withdrawals[x].Amount,
			TransferType:    "withdrawal",
			CryptoToAddress: withdrawals[x].ToReceivingAddress,
			CryptoTxID:      withdrawals[x].TransactionID,
			CryptoChain:     withdrawals[x].ChainName,
			TransferID:      withdrawals[x].WithdrawalID,
			Fee:             withdrawals[x].WithdrawalFee,
		})
	}
	return resp, nil
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (ok *Okx) GetRecentTrades(ctx context.Context, p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	format, err := ok.GetPairFormat(assetType, true)
	if err != nil {
		return nil, err
	}
	if !p.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	instrumentID := format.Format(p)
	tradeData, err := ok.GetTrades(ctx, instrumentID, 1000)
	if err != nil {
		return nil, err
	}

	resp := make([]trade.Data, len(tradeData))
	var side order.Side
	for x := range tradeData {
		side, err = order.StringToOrderSide(tradeData[x].Side)
		if err != nil {
			return nil, err
		}
		resp[x] = trade.Data{
			TID:          tradeData[x].TradeID,
			Exchange:     ok.Name,
			CurrencyPair: p,
			AssetType:    assetType,
			Side:         side,
			Price:        tradeData[x].Price,
			Amount:       tradeData[x].Quantity,
			Timestamp:    tradeData[x].Timestamp.Time(),
		}
	}
	if ok.IsSaveTradeDataEnabled() {
		err = trade.AddTradesToBuffer(ok.Name, resp...)
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(trade.ByDate(resp))
	return resp, nil
}

// GetHistoricTrades retrieves historic trade data within the timeframe provided
func (ok *Okx) GetHistoricTrades(ctx context.Context, p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	if timestampStart.Before(time.Now().Add(-kline.ThreeMonth.Duration())) {
		return nil, errOnlyThreeMonthsSupported
	}
	const limit = 100
	format, err := ok.GetPairFormat(assetType, true)
	if err != nil {
		return nil, err
	}
	if !p.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	var resp []trade.Data
	instrumentID := format.Format(p)
	tradeIDEnd := ""
allTrades:
	for {
		var trades []TradeResponse
		trades, err = ok.GetTradesHistory(ctx, instrumentID, "", tradeIDEnd, limit)
		if err != nil {
			return nil, err
		}
		if len(trades) == 0 {
			break
		}
		for i := 0; i < len(trades); i++ {
			if timestampStart.Equal(trades[i].Timestamp.Time()) ||
				trades[i].Timestamp.Time().Before(timestampStart) ||
				tradeIDEnd == trades[len(trades)-1].TradeID {
				// reached end of trades to crawl
				break allTrades
			}
			var tradeSide order.Side
			tradeSide, err = order.StringToOrderSide(trades[i].Side)
			if err != nil {
				return nil, err
			}
			resp = append(resp, trade.Data{
				TID:          trades[i].TradeID,
				Exchange:     ok.Name,
				CurrencyPair: p,
				AssetType:    assetType,
				Price:        trades[i].Price,
				Amount:       trades[i].Quantity,
				Timestamp:    trades[i].Timestamp.Time(),
				Side:         tradeSide,
			})
		}
		tradeIDEnd = trades[len(trades)-1].TradeID
	}
	if ok.IsSaveTradeDataEnabled() {
		err = trade.AddTradesToBuffer(ok.Name, resp...)
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(trade.ByDate(resp))
	return trade.FilterTradesByTime(resp, timestampStart, timestampEnd), nil
}

// SubmitOrder submits a new order
func (ok *Okx) SubmitOrder(ctx context.Context, s *order.Submit) (*order.SubmitResponse, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if !ok.SupportsAsset(s.AssetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, s.AssetType)
	}
	if s.Amount <= 0 {
		return nil, fmt.Errorf("amount, or size (sz) of quantity to buy or sell hast to be greater than zero ")
	}
	format, err := ok.GetPairFormat(s.AssetType, true)
	if err != nil {
		return nil, err
	}
	if !s.Pair.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	instrumentID := format.Format(s.Pair)
	var tradeMode string
	if s.AssetType != asset.Margin {
		tradeMode = "cash"
	}
	var sideType string
	if s.Side.IsLong() {
		sideType = order.Buy.Lower()
	} else {
		sideType = order.Sell.Lower()
	}
	var orderRequest = &PlaceOrderRequestParam{
		InstrumentID:          instrumentID,
		TradeMode:             tradeMode,
		Side:                  sideType,
		OrderType:             s.Type.Lower(),
		Amount:                s.Amount,
		ClientSupplierOrderID: s.ClientOrderID,
		Price:                 s.Price,
	}
	switch s.Type.Lower() {
	case OkxOrderLimit, OkxOrderPostOnly, OkxOrderFOK, OkxOrderIOC:
		orderRequest.Price = s.Price
	}
	var placeOrderResponse *OrderData
	if s.AssetType == asset.PerpetualSwap || s.AssetType == asset.Futures {
		if s.Type.Lower() == "" {
			orderRequest.OrderType = OkxOrderOptimalLimitIOC
		}
		// TODO: handle positionSideLong while side is Short and positionSideShort while side is Long
		if s.Side.IsLong() {
			orderRequest.PositionSide = positionSideLong
		} else {
			orderRequest.PositionSide = positionSideShort
		}
	}
	if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		placeOrderResponse, err = ok.WsPlaceOrder(orderRequest)
		if err != nil {
			return nil, err
		}
	} else {
		placeOrderResponse, err = ok.PlaceOrder(ctx, orderRequest, s.AssetType)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return s.DeriveSubmitResponse(placeOrderResponse.OrderID)
}

// ModifyOrder will allow of changing orderbook placement and limit to market conversion
func (ok *Okx) ModifyOrder(ctx context.Context, action *order.Modify) (*order.ModifyResponse, error) {
	if err := action.Validate(); err != nil {
		return nil, err
	}
	var err error
	if math.Mod(action.Amount, 1) != 0 {
		return nil, errors.New("okx contract amount can not be decimal")
	}
	format, err := ok.GetPairFormat(action.AssetType, true)
	if err != nil {
		return nil, err
	}
	if !action.Pair.IsPopulated() {
		return nil, errIncompleteCurrencyPair
	}
	instrumentID := format.Format(action.Pair)
	if err != nil {
		return nil, err
	}
	amendRequest := AmendOrderRequestParams{
		InstrumentID:          instrumentID,
		NewQuantity:           action.Amount,
		OrderID:               action.OrderID,
		ClientSuppliedOrderID: action.ClientOrderID,
	}
	if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		_, err = ok.WsAmendOrder(&amendRequest)
	} else {
		_, err = ok.AmendOrder(ctx, &amendRequest)
	}
	if err != nil {
		return nil, err
	}
	return action.DeriveModifyResponse()
}

// CancelOrder cancels an order by its corresponding ID number
func (ok *Okx) CancelOrder(ctx context.Context, ord *order.Cancel) error {
	if err := ord.Validate(ord.StandardCancel()); err != nil {
		return err
	}
	if !ok.SupportsAsset(ord.AssetType) {
		return fmt.Errorf("%w: %v", asset.ErrNotSupported, ord.AssetType)
	}
	format, err := ok.GetPairFormat(ord.AssetType, true)
	if err != nil {
		return err
	}
	if !ord.Pair.IsPopulated() {
		return errIncompleteCurrencyPair
	}
	instrumentID := format.Format(ord.Pair)
	req := CancelOrderRequestParam{
		InstrumentID:          instrumentID,
		OrderID:               ord.OrderID,
		ClientSupplierOrderID: ord.ClientOrderID,
	}
	if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		_, err = ok.WsCancelOrder(req)
	} else {
		_, err = ok.CancelSingleOrder(ctx, req)
	}
	return err
}

// CancelBatchOrders cancels orders by their corresponding ID numbers
func (ok *Okx) CancelBatchOrders(ctx context.Context, orders []order.Cancel) (order.CancelBatchResponse, error) {
	var cancelBatchResponse order.CancelBatchResponse
	if len(orders) > 20 {
		return cancelBatchResponse, fmt.Errorf("%w, cannot cancel more than 20 orders", errExceedLimit)
	} else if len(orders) == 0 {
		return cancelBatchResponse, fmt.Errorf("%w, must have at least 1 cancel order", errNoOrderParameterPassed)
	}
	cancelOrderParams := []CancelOrderRequestParam{}
	var err error
	for x := range orders {
		ord := orders[x]
		err = ord.Validate(ord.StandardCancel())
		if err != nil {
			return cancelBatchResponse, err
		}
		if !ok.SupportsAsset(ord.AssetType) {
			return cancelBatchResponse, fmt.Errorf("%w: %v", asset.ErrNotSupported, ord.AssetType)
		}
		var instrumentID string
		var format currency.PairFormat
		format, err = ok.GetPairFormat(ord.AssetType, true)
		if err != nil {
			return cancelBatchResponse, err
		}
		if !ord.Pair.IsPopulated() {
			return cancelBatchResponse, errIncompleteCurrencyPair
		}
		instrumentID = format.Format(ord.Pair)
		if err != nil {
			return cancelBatchResponse, err
		}
		req := CancelOrderRequestParam{
			InstrumentID:          instrumentID,
			OrderID:               ord.OrderID,
			ClientSupplierOrderID: ord.ClientOrderID,
		}
		cancelOrderParams = append(cancelOrderParams, req)
	}
	var canceledOrders []OrderData
	if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
		canceledOrders, err = ok.WsCancelMultipleOrder(cancelOrderParams)
	} else {
		canceledOrders, err = ok.CancelMultipleOrders(ctx, cancelOrderParams)
	}
	if err != nil {
		return cancelBatchResponse, err
	}
	for x := range canceledOrders {
		cancelBatchResponse.Status[canceledOrders[x].OrderID] = func() string {
			if canceledOrders[x].SCode != "0" && canceledOrders[x].SCode != "2" {
				return ""
			}
			return order.Cancelled.String()
		}()
	}
	return cancelBatchResponse, nil
}

// CancelAllOrders cancels all orders associated with a currency pair
func (ok *Okx) CancelAllOrders(ctx context.Context, orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	err := orderCancellation.Validate()
	if err != nil {
		return order.CancelAllResponse{}, err
	}
	cancelAllResponse := order.CancelAllResponse{
		Status: map[string]string{},
	}
	var instrumentType string
	if orderCancellation.AssetType.IsValid() {
		err = ok.CurrencyPairs.IsAssetEnabled(orderCancellation.AssetType)
		if err != nil {
			return order.CancelAllResponse{}, err
		}
		instrumentType = ok.GetInstrumentTypeFromAssetItem(orderCancellation.AssetType)
	}
	var oType string
	if orderCancellation.Type != order.UnknownType && orderCancellation.Type != order.AnyType {
		oType, err = ok.OrderTypeString(orderCancellation.Type)
		if err != nil {
			return order.CancelAllResponse{}, err
		}
	}
	var curr string
	if orderCancellation.Pair.IsPopulated() {
		curr = orderCancellation.Pair.Upper().String()
	}
	myOrders, err := ok.GetOrderList(ctx, &OrderListRequestParams{
		InstrumentType: instrumentType,
		OrderType:      oType,
		InstrumentID:   curr,
	})
	if err != nil {
		return cancelAllResponse, err
	}
	cancelAllOrdersRequestParams := make([]CancelOrderRequestParam, len(myOrders))
ordersLoop:
	for x := range myOrders {
		switch {
		case orderCancellation.OrderID != "" || orderCancellation.ClientOrderID != "":
			if myOrders[x].OrderID == orderCancellation.OrderID ||
				myOrders[x].ClientSupplierOrderID == orderCancellation.ClientOrderID {
				cancelAllOrdersRequestParams[x] = CancelOrderRequestParam{
					OrderID:               myOrders[x].OrderID,
					ClientSupplierOrderID: myOrders[x].ClientSupplierOrderID,
				}
				break ordersLoop
			}
		case orderCancellation.Side == order.Buy || orderCancellation.Side == order.Sell:
			if myOrders[x].Side == order.Buy || myOrders[x].Side == order.Sell {
				cancelAllOrdersRequestParams[x] = CancelOrderRequestParam{
					OrderID:               myOrders[x].OrderID,
					ClientSupplierOrderID: myOrders[x].ClientSupplierOrderID,
				}
				continue
			}
		default:
			cancelAllOrdersRequestParams[x] = CancelOrderRequestParam{
				OrderID:               myOrders[x].OrderID,
				ClientSupplierOrderID: myOrders[x].ClientSupplierOrderID,
			}
		}
	}
	remaining := cancelAllOrdersRequestParams
	loop := int(math.Ceil(float64(len(remaining)) / 20.0))
	for b := 0; b < loop; b++ {
		var response []OrderData
		if len(remaining) > 20 {
			if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
				response, err = ok.WsCancelMultipleOrder(remaining[:20])
			} else {
				response, err = ok.CancelMultipleOrders(ctx, remaining[:20])
			}
			remaining = remaining[20:]
		} else {
			if ok.Websocket.CanUseAuthenticatedWebsocketForWrapper() {
				response, err = ok.WsCancelMultipleOrder(remaining)
			} else {
				response, err = ok.CancelMultipleOrders(ctx, remaining)
			}
		}
		if err != nil {
			if len(cancelAllResponse.Status) == 0 {
				return cancelAllResponse, err
			}
		}
		for y := range response {
			if response[y].SCode == "0" {
				cancelAllResponse.Status[response[y].OrderID] = order.Cancelled.String()
			} else {
				cancelAllResponse.Status[response[y].OrderID] = response[y].SMessage
			}
		}
	}
	return cancelAllResponse, nil
}

// GetOrderInfo returns order information based on order ID
func (ok *Okx) GetOrderInfo(ctx context.Context, orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	var respData order.Detail
	format, err := ok.GetPairFormat(assetType, true)
	if err != nil {
		return respData, err
	}
	if !pair.IsPopulated() {
		return respData, errIncompleteCurrencyPair
	}
	instrumentID := format.Format(pair)
	if !ok.SupportsAsset(assetType) {
		return respData, fmt.Errorf("%w: %v", asset.ErrNotSupported, assetType)
	}
	orderDetail, err := ok.GetOrderDetail(ctx, &OrderDetailRequestParam{
		InstrumentID: instrumentID,
		OrderID:      orderID,
	})
	if err != nil {
		return respData, err
	}
	status, err := order.StringToOrderStatus(orderDetail.State)
	if err != nil {
		return respData, err
	}
	orderType, err := ok.OrderTypeFromString(orderDetail.OrderType)
	if err != nil {
		return respData, err
	}

	return order.Detail{
		Amount:         orderDetail.Size.Float64(),
		Exchange:       ok.Name,
		OrderID:        orderDetail.OrderID,
		ClientOrderID:  orderDetail.ClientSupplierOrderID,
		Side:           orderDetail.Side,
		Type:           orderType,
		Pair:           pair,
		Cost:           orderDetail.Price.Float64(),
		AssetType:      assetType,
		Status:         status,
		Price:          orderDetail.Price.Float64(),
		ExecutedAmount: orderDetail.RebateAmount.Float64(),
		Date:           orderDetail.CreationTime,
		LastUpdated:    orderDetail.UpdateTime,
	}, err
}

// GetDepositAddress returns a deposit address for a specified currency
func (ok *Okx) GetDepositAddress(ctx context.Context, c currency.Code, _, chain string) (*deposit.Address, error) {
	response, err := ok.GetCurrencyDepositAddress(ctx, c.String())
	if err != nil {
		return nil, err
	}

	// Check if a specific chain was requested
	if chain != "" {
		for x := range response {
			if !strings.EqualFold(response[x].Chain, chain) {
				continue
			}
			return &deposit.Address{
				Address: response[x].Address,
				Tag:     response[x].Tag,
				Chain:   response[x].Chain,
			}, nil
		}
		return nil, fmt.Errorf("specified chain %s not found", chain)
	}

	// If no specific chain was requested, return the first selected address (mainnet addresses are returned first by default)
	for x := range response {
		if !response[x].Selected {
			continue
		}

		return &deposit.Address{
			Address: response[x].Address,
			Tag:     response[x].Tag,
			Chain:   response[x].Chain,
		}, nil
	}
	return nil, errDepositAddressNotFound
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is submitted
func (ok *Okx) WithdrawCryptocurrencyFunds(ctx context.Context, withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	if err := withdrawRequest.Validate(); err != nil {
		return nil, err
	}
	input := WithdrawalInput{
		ChainName:             withdrawRequest.Crypto.Chain,
		Amount:                withdrawRequest.Amount,
		Currency:              withdrawRequest.Currency.String(),
		ToAddress:             withdrawRequest.Crypto.Address,
		TransactionFee:        withdrawRequest.Crypto.FeeAmount,
		WithdrawalDestination: "3",
	}
	resp, err := ok.Withdrawal(ctx, &input)
	if err != nil {
		return nil, err
	}
	return &withdraw.ExchangeResponse{
		ID: resp.WithdrawalID,
	}, nil
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (ok *Okx) WithdrawFiatFunds(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is submitted
func (ok *Okx) WithdrawFiatFundsToInternationalBank(_ context.Context, _ *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrFunctionNotSupported
}

// GetActiveOrders retrieves any orders that are active/open
func (ok *Okx) GetActiveOrders(ctx context.Context, req *order.GetOrdersRequest) (order.FilteredOrders, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	if !req.StartTime.IsZero() && req.StartTime.Before(time.Now().Add(-kline.ThreeMonth.Duration())) {
		return nil, errOnlyThreeMonthsSupported
	}
	if !ok.SupportsAsset(req.AssetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, req.AssetType)
	}
	instrumentType := ok.GetInstrumentTypeFromAssetItem(req.AssetType)
	var orderType string
	if req.Type != order.UnknownType && req.Type != order.AnyType {
		orderType, err = ok.OrderTypeString(req.Type)
		if err != nil {
			return nil, err
		}
	}
	endTime := req.EndTime
	var resp []order.Detail
allOrders:
	for {
		requestParam := &OrderListRequestParams{
			OrderType:      orderType,
			End:            endTime,
			InstrumentType: instrumentType,
		}
		var orderList []OrderDetail
		orderList, err = ok.GetOrderList(ctx, requestParam)
		if err != nil {
			return nil, err
		}
		if len(orderList) == 0 {
			break
		}
		for i := range orderList {
			if req.StartTime.Equal(orderList[i].CreationTime) ||
				orderList[i].CreationTime.Before(req.StartTime) ||
				endTime == orderList[i].CreationTime {
				// reached end of orders to crawl
				break allOrders
			}
			orderSide := orderList[i].Side
			var pair currency.Pair
			pair, err = ok.GetPairFromInstrumentID(orderList[i].InstrumentID)
			if err != nil {
				return nil, err
			}
			if len(req.Pairs) > 0 {
				x := 0
				for x = range req.Pairs {
					if req.Pairs[x].Equal(pair) {
						break
					}
				}
				if !req.Pairs[x].Equal(pair) {
					continue
				}
			}
			var orderStatus order.Status
			orderStatus, err = order.StringToOrderStatus(strings.ToUpper(orderList[i].State))
			if err != nil {
				return nil, err
			}
			var oType order.Type
			oType, err = ok.OrderTypeFromString(orderList[i].OrderType)
			if err != nil {
				return nil, err
			}
			resp = append(resp, order.Detail{
				Amount:          orderList[i].Size.Float64(),
				Pair:            pair,
				Price:           orderList[i].Price.Float64(),
				ExecutedAmount:  orderList[i].FillSize.Float64(),
				RemainingAmount: orderList[i].Size.Float64() - orderList[i].FillSize.Float64(),
				Fee:             orderList[i].TransactionFee.Float64(),
				FeeAsset:        currency.NewCode(orderList[i].FeeCurrency),
				Exchange:        ok.Name,
				OrderID:         orderList[i].OrderID,
				ClientOrderID:   orderList[i].ClientSupplierOrderID,
				Type:            oType,
				Side:            orderSide,
				Status:          orderStatus,
				AssetType:       req.AssetType,
				Date:            orderList[i].CreationTime,
				LastUpdated:     orderList[i].UpdateTime,
			})
		}
		if len(orderList) < 100 {
			// Since the we passed a limit of 0 to the method GetOrderList,
			// we expect 100 orders to be retrieved if the number of orders are more that 100.
			// If not, break out of the loop to not send another request.
			break
		}
		endTime = orderList[len(orderList)-1].CreationTime
	}
	return req.Filter(ok.Name, resp), nil
}

// GetOrderHistory retrieves account order information Can Limit response to specific order status
func (ok *Okx) GetOrderHistory(ctx context.Context, req *order.GetOrdersRequest) (order.FilteredOrders, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if !req.StartTime.IsZero() && req.StartTime.Before(time.Now().Add(-kline.ThreeMonth.Duration())) {
		return nil, errOnlyThreeMonthsSupported
	}
	if len(req.Pairs) == 0 {
		return nil, errMissingAtLeast1CurrencyPair
	}
	if !ok.SupportsAsset(req.AssetType) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, req.AssetType)
	}
	instrumentType := strings.ToUpper(ok.GetInstrumentTypeFromAssetItem(req.AssetType))
	endTime := req.EndTime
	var resp []order.Detail
allOrders:
	for {
		orderList, err := ok.Get3MonthOrderHistory(ctx, &OrderHistoryRequestParams{
			OrderListRequestParams: OrderListRequestParams{
				InstrumentType: instrumentType,
				End:            endTime,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(orderList) == 0 {
			break
		}
		for i := range orderList {
			if req.StartTime.Equal(orderList[i].CreationTime) ||
				orderList[i].CreationTime.Before(req.StartTime) ||
				endTime == orderList[i].CreationTime {
				// reached end of orders to crawl
				break allOrders
			}
			var pair currency.Pair
			pair, err = ok.GetPairFromInstrumentID(orderList[i].InstrumentID)
			if err != nil {
				return nil, err
			}
			for j := range req.Pairs {
				if !req.Pairs[j].Equal(pair) {
					continue
				}
				var orderStatus order.Status
				orderStatus, err = order.StringToOrderStatus(strings.ToUpper(orderList[i].State))
				if err != nil {
					log.Errorf(log.ExchangeSys, "%s %v", ok.Name, err)
				}
				if orderStatus == order.Active {
					continue
				}
				orderSide := orderList[i].Side
				var oType order.Type
				oType, err = ok.OrderTypeFromString(orderList[i].OrderType)
				if err != nil {
					return nil, err
				}
				orderAmount := orderList[i].Size
				if orderList[i].QuantityType == "quote_ccy" {
					// Size is quote amount.
					orderAmount /= orderList[i].AveragePrice
				}

				remainingAmount := float64(0)
				if orderStatus != order.Filled {
					remainingAmount = orderAmount.Float64() - orderList[i].AccumulatedFillSize.Float64()
				}
				resp = append(resp, order.Detail{
					Price:                orderList[i].Price.Float64(),
					AverageExecutedPrice: orderList[i].AveragePrice.Float64(),
					Amount:               orderAmount.Float64(),
					ExecutedAmount:       orderList[i].AccumulatedFillSize.Float64(),
					RemainingAmount:      remainingAmount,
					Fee:                  orderList[i].TransactionFee.Float64(),
					FeeAsset:             currency.NewCode(orderList[i].FeeCurrency),
					Exchange:             ok.Name,
					OrderID:              orderList[i].OrderID,
					ClientOrderID:        orderList[i].ClientSupplierOrderID,
					Type:                 oType,
					Side:                 orderSide,
					Status:               orderStatus,
					AssetType:            req.AssetType,
					Date:                 orderList[i].CreationTime,
					LastUpdated:          orderList[i].UpdateTime,
					Pair:                 pair,
					Cost:                 orderList[i].AveragePrice.Float64() * orderList[i].AccumulatedFillSize.Float64(),
					CostAsset:            currency.NewCode(orderList[i].RebateCurrency),
				})
			}
		}
		if len(orderList) < 100 {
			break
		}
		endTime = orderList[len(orderList)-1].CreationTime
	}
	return req.Filter(ok.Name, resp), nil
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (ok *Okx) GetFeeByType(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	if feeBuilder == nil {
		return 0, fmt.Errorf("%T %w", feeBuilder, common.ErrNilPointer)
	}
	if !ok.AreCredentialsValid(ctx) && feeBuilder.FeeType == exchange.CryptocurrencyTradeFee {
		feeBuilder.FeeType = exchange.OfflineTradeFee
	}
	return ok.GetFee(ctx, feeBuilder)
}

// ValidateAPICredentials validates current credentials used for wrapper
func (ok *Okx) ValidateAPICredentials(ctx context.Context, assetType asset.Item) error {
	_, err := ok.UpdateAccountInfo(ctx, assetType)
	return ok.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (ok *Okx) GetHistoricCandles(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := ok.GetKlineRequest(pair, a, interval, start, end, false)
	if err != nil {
		return nil, err
	}

	candles, err := ok.GetCandlesticksHistory(ctx,
		req.RequestFormatted.Base.String()+
			currency.DashDelimiter+
			req.RequestFormatted.Quote.String(),
		req.ExchangeInterval,
		start.Add(-time.Nanosecond), // Start time not inclusive of candle.
		end,
		300)
	if err != nil {
		return nil, err
	}

	timeSeries := make([]kline.Candle, len(candles))
	for x := range candles {
		timeSeries[x] = kline.Candle{
			Time:   candles[x].OpenTime,
			Open:   candles[x].OpenPrice,
			High:   candles[x].HighestPrice,
			Low:    candles[x].LowestPrice,
			Close:  candles[x].ClosePrice,
			Volume: candles[x].Volume,
		}
	}
	return req.ProcessResponse(timeSeries)
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (ok *Okx) GetHistoricCandlesExtended(ctx context.Context, pair currency.Pair, a asset.Item, interval kline.Interval, start, end time.Time) (*kline.Item, error) {
	req, err := ok.GetKlineExtendedRequest(pair, a, interval, start, end)
	if err != nil {
		return nil, err
	}

	count := kline.TotalCandlesPerInterval(req.Start, req.End, req.ExchangeInterval)
	if count > 1440 {
		return nil,
			fmt.Errorf("candles count: %d max lookback: %d, %w",
				count, 1440, kline.ErrRequestExceedsMaxLookback)
	}

	timeSeries := make([]kline.Candle, 0, req.Size())
	for y := range req.RangeHolder.Ranges {
		var candles []CandleStick
		candles, err = ok.GetCandlesticksHistory(ctx,
			req.RequestFormatted.Base.String()+
				currency.DashDelimiter+
				req.RequestFormatted.Quote.String(),
			req.ExchangeInterval,
			req.RangeHolder.Ranges[y].Start.Time.Add(-time.Nanosecond), // Start time not inclusive of candle.
			req.RangeHolder.Ranges[y].End.Time,
			300)
		if err != nil {
			return nil, err
		}
		for x := range candles {
			timeSeries = append(timeSeries, kline.Candle{
				Time:   candles[x].OpenTime,
				Open:   candles[x].OpenPrice,
				High:   candles[x].HighestPrice,
				Low:    candles[x].LowestPrice,
				Close:  candles[x].ClosePrice,
				Volume: candles[x].Volume,
			})
		}
	}
	return req.ProcessResponse(timeSeries)
}

// GetAvailableTransferChains returns the available transfer blockchains for the specific
// cryptocurrency
func (ok *Okx) GetAvailableTransferChains(ctx context.Context, cryptocurrency currency.Code) ([]string, error) {
	currencyChains, err := ok.GetFundingCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	chains := make([]string, 0, len(currencyChains))
	for x := range currencyChains {
		if (!cryptocurrency.IsEmpty() && !strings.EqualFold(cryptocurrency.String(), currencyChains[x].Currency)) ||
			(!currencyChains[x].CanDeposit && !currencyChains[x].CanWithdraw) ||
			// Lightning network is currently not supported by transfer chains
			// as it is an invoice string which is generated per request and is
			// not a static address. TODO: Add a hook to generate a new invoice
			// string per request.
			(currencyChains[x].Chain != "" && currencyChains[x].Chain == "BTC-Lightning") {
			continue
		}
		chains = append(chains, currencyChains[x].Chain)
	}
	return chains, nil
}

// getInstrumentsForOptions returns the instruments for options asset type
func (ok *Okx) getInstrumentsForOptions(ctx context.Context) ([]Instrument, error) {
	underlyings, err := ok.GetPublicUnderlyings(context.Background(), okxInstTypeOption)
	if err != nil {
		return nil, err
	}
	var insts []Instrument
	for x := range underlyings {
		var instruments []Instrument
		instruments, err = ok.GetInstruments(ctx, &InstrumentsFetchParams{
			InstrumentType: okxInstTypeOption,
			Underlying:     underlyings[x],
		})
		if err != nil {
			return nil, err
		}
		insts = append(insts, instruments...)
	}
	return insts, nil
}

// getInstrumentsForAsset returns the instruments for an asset type
func (ok *Okx) getInstrumentsForAsset(ctx context.Context, a asset.Item) ([]Instrument, error) {
	if !ok.SupportsAsset(a) {
		return nil, fmt.Errorf("asset type of %s is not supported by %s", a, ok.Name)
	}

	var instType string
	switch a {
	case asset.Options:
		return ok.getInstrumentsForOptions(ctx)
	case asset.Spot:
		instType = okxInstTypeSpot
	case asset.Futures:
		instType = okxInstTypeFutures
	case asset.PerpetualSwap:
		instType = okxInstTypeSwap
	case asset.Margin:
		instType = okxInstTypeMargin
	}

	return ok.GetInstruments(ctx, &InstrumentsFetchParams{
		InstrumentType: instType,
	})
}

// SetMarginType sets the default margin type for when opening a new position
// okx allows this to be set with an order, however this sets a default
func (ok *Okx) SetMarginType(_ context.Context, _ asset.Item, _ currency.Pair, _ margin.Type) error {
	return fmt.Errorf("%w margin type is set per order", common.ErrFunctionNotSupported)
}

func (ok *Okx) SetCollateralMode(_ context.Context, _ asset.Item, _ order.CollateralType) error {
	return fmt.Errorf("%w must be set via website", common.ErrFunctionNotSupported)
}

func (ok *Okx) GetCollateralMode(ctx context.Context, item asset.Item) (order.CollateralType, error) {
	cfg, err := ok.GetAccountConfiguration(ctx)
	if err != nil {
		return 0, err
	}
	switch cfg[0].AccountLevel {
	case 1:
		if item != asset.Spot {
			return 0, fmt.Errorf("%w %v", asset.ErrNotSupported, item)
		}
		fallthrough
	case 2:
		return order.SingleCollateral, nil
	case 3:
		return order.MultiCollateral, nil
	case 4:
		return order.GlobalCollateral, nil
	default:
		return order.UnknownCollateral, fmt.Errorf("%w %v", order.ErrCollateralInvalid, cfg[0].AccountLevel)
	}
}

// ChangePositionMargin will modify a position/currencies margin parameters
func (ok *Okx) ChangePositionMargin(ctx context.Context, req *margin.PositionChangeRequest) (*margin.PositionChangeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w PositionChangeRequest", common.ErrNilPointer)
	}
	if !ok.SupportsAsset(req.Asset) {
		return nil, fmt.Errorf("%w: %v", asset.ErrNotSupported, req.Asset)
	}
	if req.NewAllocatedMargin == 0 {
		return nil, fmt.Errorf("%w %v %v", margin.ErrNewAllocatedMarginRequired, req.Asset, req.Pair)
	}
	if req.OriginalAllocatedMargin == 0 {
		return nil, margin.ErrOriginalPositionMarginRequired
	}
	if req.MarginType != margin.Isolated {
		return nil, fmt.Errorf("%w %v", margin.ErrMarginTypeUnsupported, req.MarginType)
	}
	format, err := ok.GetPairFormat(req.Asset, true)
	fPair := req.Pair.Format(format)
	marginType := "add"
	amt := req.NewAllocatedMargin - req.OriginalAllocatedMargin
	if req.NewAllocatedMargin < req.OriginalAllocatedMargin {
		marginType = "reduce"
		amt = req.OriginalAllocatedMargin - req.NewAllocatedMargin
	}
	if req.MarginSide == "" {
		req.MarginSide = "net"
	}
	r := IncreaseDecreaseMarginInput{
		InstrumentID: fPair.String(),
		PositionSide: req.MarginSide,
		Type:         marginType,
		Amount:       amt,
	}

	if req.Asset == asset.Margin {
		r.Currency = req.Pair.Base.Item.Symbol
	}

	resp, err := ok.IncreaseDecreaseMargin(ctx, r)
	if err != nil {
		return nil, err
	}

	return &margin.PositionChangeResponse{
		Exchange:        ok.Name,
		Pair:            req.Pair,
		Asset:           req.Asset,
		AllocatedMargin: resp.Amount.Float64(),
		MarginType:      req.MarginType,
	}, nil
}

func (ok *Okx) GetPositionSummary(ctx context.Context, req *order.PositionSummaryRequest) (*order.PositionSummary, error) {
	if req == nil {
		return nil, fmt.Errorf("%w PositionSummaryRequest", common.ErrNilPointer)
	}
	if req.CalculateOffline {
		return nil, common.ErrCannotCalculateOffline
	}
	if !ok.SupportsAsset(req.Asset) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, req.Asset)
	}
	fPair, err := ok.FormatExchangeCurrency(req.Pair, req.Asset)
	if err != nil {
		return nil, err
	}
	_, err := ok.GetBalance(ctx, fPair.String())
	if err != nil {
		return nil, err
	}
	return nil, nil
	resp := &order.PositionSummary{
		Pair:                         req.Pair,
		Asset:                        req.Asset,
		CollateralType:               order.UnknownCollateral,
		IsolatedMargin:               decimal.Decimal{},
		NotionalSize:                 decimal.Decimal{},
		Leverage:                     decimal.Decimal{},
		MaintenanceMarginRequirement: decimal.Decimal{},
		InitialMarginRequirement:     decimal.Decimal{},
		EstimatedLiquidationPrice:    decimal.Decimal{},
		CollateralUsed:               decimal.Decimal{},
		MarkPrice:                    decimal.Decimal{},
		CurrentSize:                  decimal.Decimal{},
		AverageOpenPrice:             decimal.Decimal{},
		PositionPNL:                  decimal.Decimal{},
		MaintenanceMarginFraction:    decimal.Decimal{},
		FreeCollateral:               decimal.Decimal{},
		TotalCollateral:              decimal.Decimal{},
	}
	return resp, nil
}

// GetFuturesPositionOrders returns the orders for futures positions
func (ok *Okx) GetFuturesPositionOrders(ctx context.Context, req *order.PositionsRequest) ([]order.PositionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w PositionSummaryRequest", common.ErrNilPointer)
	}
	if !ok.SupportsAsset(req.Asset) {
		return nil, fmt.Errorf("%w %v", asset.ErrNotSupported, req.Asset)
	}
	fPair, err := ok.FormatExchangeCurrency(req.Pair, req.Asset)
	if err != nil {
		return nil, err
	}
	instrumentType := ok.GetInstrumentTypeFromAssetItem(req.Asset)

	positions, err := ok.GetPositions(ctx, instrumentType, fPair.String(), "")
	if err != nil {
		return nil, err
	}
	resp := &order.PositionSummary{
		Pair:                         currency.Pair{},
		Asset:                        0,
		MarginType:                   0,
		CollateralType:               0,
		Currency:                     currency.Code{},
		IsolatedMargin:               decimal.Decimal{},
		NotionalSize:                 decimal.Decimal{},
		Leverage:                     decimal.Decimal{},
		MaintenanceMarginRequirement: decimal.Decimal{},
		InitialMarginRequirement:     decimal.Decimal{},
		EstimatedLiquidationPrice:    decimal.Decimal{},
		CollateralUsed:               decimal.Decimal{},
		MarkPrice:                    decimal.Decimal{},
		CurrentSize:                  decimal.Decimal{},
		AverageOpenPrice:             decimal.Decimal{},
		PositionPNL:                  decimal.Decimal{},
		MaintenanceMarginFraction:    decimal.Decimal{},
		FreeCollateral:               decimal.Decimal{},
		TotalCollateral:              decimal.Decimal{},
	}
	for i := range positions {

	}
	return nil, nil
}
