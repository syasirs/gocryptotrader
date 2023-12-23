package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	binanceCFuturesWebsocketURL = "wss://dstream.binance.com"
)

var defaultCFuturesSubscriptions = []string{
	depthChan,
	tickerAllChan,
	continuousKline,
}

func (b *Binance) WsCFutureConnect() error {
	if !b.Websocket.IsEnabled() || !b.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}
	var err error
	var dialer websocket.Dialer
	dialer.HandshakeTimeout = b.Config.HTTPTimeout
	dialer.Proxy = http.ProxyFromEnvironment
	wsURL := binanceCFuturesWebsocketURL + "/stream"
	err = b.Websocket.SetWebsocketURL(wsURL, false, false)
	if err != nil {
		return err
	}
	switch {
	case err != nil:
		b.Websocket.SetCanUseAuthenticatedEndpoints(false)
		log.Errorf(log.ExchangeSys,
			"%v unable to connect to authenticated Websocket. Error: %s",
			b.Name,
			err)
	}
	err = b.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return fmt.Errorf("%v - Unable to connect to Websocket. Error: %s", b.Name, err)
	}
	b.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.PongMessage,
		Delay:             pingDelay,
	})
	b.Websocket.Wg.Add(1)
	go b.wsUFuturesReadData(asset.CoinMarginedFutures)

	subscriptions, err := b.GenerateDefaultCFuturesSubscriptions()
	if err != nil {
		return err
	}
	return b.SubscribeFutures(subscriptions)
	// return nil
}

// GenerateDefaultCFuturesSubscriptions genearates a list of subscription instances.
func (b *Binance) GenerateDefaultCFuturesSubscriptions() ([]stream.ChannelSubscription, error) {
	var channels = defaultCFuturesSubscriptions
	var subscriptions []stream.ChannelSubscription
	pairs, err := b.FetchTradablePairs(context.Background(), asset.CoinMarginedFutures)
	if err != nil {
		return nil, err
	}
	if len(pairs) > 4 {
		pairs = pairs[:3]
	}
	for z := range channels {
		var subscription stream.ChannelSubscription
		switch channels[z] {
		case contractInfoAllChan, forceOrderAllChan,
			bookTickerAllChan, tickerAllChan, miniTickerAllChan:
			subscriptions = append(subscriptions, stream.ChannelSubscription{
				Channel: channels[z],
			})
		case aggTradeChan, depthChan, markPriceChan, tickerChan,
			klineChan, miniTickerChan, forceOrderChan,
			indexPriceCFuturesChan, bookTickerCFuturesChan:
			for y := range pairs {
				lp := pairs[y].Lower()
				lp.Delimiter = ""
				subscription = stream.ChannelSubscription{
					Channel: lp.String() + channels[z],
				}
				switch channels[z] {
				case depthChan:
					subscription.Channel += "@100ms"
				case klineChan, indexPriceKlineCFuturesChan, markPriceKlineCFuturesChan:
					subscription.Channel += "_" + getKlineIntervalString(kline.FiveMin)
				}
				subscriptions = append(subscriptions, subscription)
			}
		case continuousKline:
			for y := range pairs {
				lp := pairs[y].Lower()
				lp.Delimiter = ""
				subscription = stream.ChannelSubscription{
					// Contract types:""perpetual", "current_quarter", "next_quarter""
					Channel: lp.String() + "_PERPETUAL@" + channels[z] + "_" + getKlineIntervalString(kline.FiveMin),
				}
				subscriptions = append(subscriptions, subscription)
			}
		default:
			return nil, errors.New("unsupported subscription")
		}
	}
	return subscriptions, nil
}

// wsCFuturesReadData receives and passes on websocket messages for processing
// for Coin margined instruments.
func (b *Binance) wsCFuturesReadData() {
	defer b.Websocket.Wg.Done()
	for {
		resp := b.Websocket.Conn.ReadMessage()
		if resp.Raw == nil {
			return
		}
		err := b.wsHandleCFuturesData(resp.Raw)
		if err != nil {
			b.Websocket.DataHandler <- err
		}
	}
}

func (b *Binance) wsHandleCFuturesData(respRaw []byte) error {
	result := struct {
		Result json.RawMessage `json:"result"`
		ID     int64           `json:"id"`
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}{}
	err := json.Unmarshal(respRaw, &result)
	if err != nil {
		return err
	}
	if result.Stream == "" || (result.ID != 0 && result.Result != nil) {
		if !b.Websocket.Match.IncomingWithData(result.ID, respRaw) {
			return errors.New("Unhandled data: " + string(respRaw))
		}
		return nil
	}
	var stream string
	switch result.Stream {
	case assetIndexAllChan, forceOrderAllChan, bookTickerAllChan, tickerAllChan, miniTickerAllChan:
		stream = result.Stream
	default:
		stream = extractStreamInfo(result.Stream)
	}
	switch stream {
	case contractInfoAllChan:
		return b.processContractInfoStream(result.Data)
	case forceOrderAllChan, "forceOrder":
		return b.processCFuturesForceOrder(result.Data)
	// case bookTickerAllChan, "bookTicker":
	// 	return b.processBookTicker(result.Data, assetType)
	// case tickerAllChan:
	// 	return b.processMarketTicker(result.Data, true, assetType)
	// case "ticker":
	// 	return b.processMarketTicker(result.Data, false, assetType)
	// case miniTickerAllChan:
	// 	return b.processMiniTickers(result.Data, true, assetType)
	// case "miniTicker":
	// 	return b.processMiniTickers(result.Data, false, assetType)
	// case "aggTrade":
	// 	return b.processAggregateTrade(result.Data, assetType)
	// case "markPrice":
	// 	return b.processMarkPriceUpdate(result.Data, false)
	// case "depth":
	// 	return b.processOrderbookDepthUpdate(result.Data, assetType)
	// case "compositeIndex":
	// 	return b.processCompositeIndex(result.Data)
	// case continuousKline:
	// 	return b.processContinuousKlineUpdate(result.Data, assetType)
	case klineChan:
	case indexPriceCFuturesChan:
	case indexPriceKlineCFuturesChan:
	case markPriceKlineCFuturesChan:
		return b.processMarkPriceKline(result.Data)
	}
	return fmt.Errorf("unhandled stream data %s", string(respRaw))
}

func (b *Binance) processCFuturesForceOrder(respRaw []byte) error {
	var resp MarketLiquidationOrder
	err := json.Unmarshal(respRaw, &resp)
	if err != nil {
		return err
	}
	oType, err := order.StringToOrderType(resp.Order.OrderType)
	if err != nil {
		return err
	}
	oSide, err := order.StringToOrderSide(resp.Order.Side)
	if err != nil {
		return err
	}
	oStatus, err := order.StringToOrderStatus(resp.Order.OrderStatus)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(resp.Order.Symbol)
	if err != nil {
		return err
	}
	b.Websocket.DataHandler <- order.Detail{
		Price:                resp.Order.Price.Float64(),
		Amount:               resp.Order.OriginalQuantity.Float64(),
		AverageExecutedPrice: resp.Order.AveragePrice.Float64(),
		ExecutedAmount:       resp.Order.OrderFilledAccumulatedQuantity.Float64(),
		RemainingAmount:      resp.Order.OriginalQuantity.Float64() - resp.Order.OrderFilledAccumulatedQuantity.Float64(),
		Exchange:             b.Name,
		Type:                 oType,
		Side:                 oSide,
		Status:               oStatus,
		AssetType:            asset.CoinMarginedFutures,
		LastUpdated:          resp.Order.OrderTradeTime.Time(),
		Pair:                 cp,
	}
	return nil
}

func (b *Binance) processMarkPriceKline(respRaw []byte) error {
	var resp CFutureMarkPriceKline
	err := json.Unmarshal(respRaw, &resp)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(resp.Kline.Symbol)
	if err != nil {
		return err
	}
	b.Websocket.DataHandler <- &stream.KlineData{
		Pair:       cp,
		AssetType:  asset.CoinMarginedFutures,
		Interval:   resp.Kline.Interval,
		StartTime:  resp.Kline.StartTime.Time(),
		CloseTime:  resp.Kline.CloseTime.Time(),
		OpenPrice:  resp.Kline.OpenPrice.Float64(),
		ClosePrice: resp.Kline.ClosePrice.Float64(),
		HighPrice:  resp.Kline.HighPrice.Float64(),
		LowPrice:   resp.Kline.LowPrice.Float64(),
	}
	return nil
}
