package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream/buffer"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	binanceDefaultWebsocketURL = "wss://stream.binance.com:9443/stream"
	pingDelay                  = time.Minute * 9

	// maxBatchedPayloads defines an upper restriction on outbound batched
	// subscriptions. There seems to be a byte limit that is not documented.
	// Max bytes == 4096
	maxBatchedPayloads = 150
)

var listenKey string

// Job defines a syncro job
type Job struct {
	Pair currency.Pair
	Conn stream.Connection
}

// WsConnect initiates a websocket connection
func (b *Binance) WsConnect(conn stream.Connection) error {
	if !b.Websocket.IsEnabled() || !b.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}

	var dialer websocket.Dialer
	var err error
	if conn.IsAuthenticated() {
		if b.Websocket.CanUseAuthenticatedEndpoints() {
			listenKey, err = b.GetWsAuthStreamKey()
			if err != nil {
				b.Websocket.SetCanUseAuthenticatedEndpoints(false)
				log.Errorf(log.ExchangeSys,
					"%v unable to connect to authenticated Websocket. Error: %s",
					b.Name,
					err)
			} else {
				// cleans on failed connection
				clean := strings.Split(b.Websocket.GetWebsocketURL(), "?streams=")
				authPayload := clean[0] + "?streams=" + listenKey
				err = b.Websocket.SetWebsocketURL(authPayload, false, false)
				if err != nil {
					return err
				}
			}

			fmt.Println("Authconn:", conn)

			err = conn.Dial(&dialer, http.Header{})
			if err != nil {
				return fmt.Errorf("%v - Unable to connect to Websocket. Error: %s",
					b.Name,
					err)
			}

			if b.Websocket.CanUseAuthenticatedEndpoints() {
				b.Websocket.Wg.Add(1)
				go b.KeepAuthKeyAlive()
			}

			conn.SetupPingHandler(stream.PingHandler{
				UseGorillaHandler: true,
				MessageType:       websocket.PongMessage,
				Delay:             pingDelay,
			})

			b.Websocket.Wg.Add(1)
			go b.wsReadData(conn)

			return nil
		}
		return nil
	}

	err = conn.Dial(&dialer, http.Header{})
	if err != nil {
		return fmt.Errorf("%v - Unable to connect to Websocket. Error: %s",
			b.Name,
			err)
	}

	conn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.PongMessage,
		Delay:             pingDelay,
	})

	b.preConnectionSetup(conn)

	b.Websocket.Wg.Add(1)
	go b.wsReadData(conn)
	return nil
}

func (b *Binance) spawnConnection(url string, auth bool) (stream.Connection, error) {
	if url == "" {
		return nil, errors.New("url not specified when generating connection")
	}
	return &stream.WebsocketConnection{
		Verbose:          b.Verbose,
		ExchangeName:     b.Name,
		URL:              url,
		ProxyURL:         b.Websocket.GetProxyAddress(),
		Authenticated:    auth,
		Match:            b.Websocket.Match,
		Wg:               b.Websocket.Wg,
		Traffic:          b.Websocket.TrafficAlert,
		RateLimit:        250,
		ResponseMaxLimit: time.Second * 10,
	}, nil
}

func (b *Binance) preConnectionSetup(conn stream.Connection) {
	b.buffer = make(map[string]map[asset.Item]chan *WebsocketDepthStream)
	b.fetchingbook = make(map[string]map[asset.Item]bool)
	b.initialSync = make(map[string]map[asset.Item]bool)
	b.jobs = make(chan Job, 2000)
	for i := 0; i < 10; i++ {
		// 10 workers for synchronising book
		b.SynchroniseWebsocketOrderbook()
	}
}

// KeepAuthKeyAlive will continuously send messages to
// keep the WS auth key active
func (b *Binance) KeepAuthKeyAlive() {
	defer b.Websocket.Wg.Done()
	ticks := time.NewTicker(time.Minute * 30)
	for {
		select {
		case <-b.Websocket.ShutdownC:
			ticks.Stop()
			return
		case <-ticks.C:
			err := b.MaintainWsAuthStreamKey()
			if err != nil {
				b.Websocket.DataHandler <- err
				log.Warnf(log.ExchangeSys,
					b.Name+" - Unable to renew auth websocket token, may experience shutdown")
			}
		}
	}
}

// wsReadData receives and passes on websocket messages for processing
func (b *Binance) wsReadData(conn stream.Connection) {
	defer b.Websocket.Wg.Done()

	for {
		resp := conn.ReadMessage()
		if resp.Raw == nil {
			return
		}
		go func() {
			err := b.wsHandleData(resp.Raw, conn)
			if err != nil {
				b.Websocket.DataHandler <- err
			}
		}()
	}
}

func (b *Binance) wsHandleData(respRaw []byte, conn stream.Connection) error {
	// fmt.Println("Incoming:", string(respRaw))
	var multiStreamData map[string]interface{}
	err := json.Unmarshal(respRaw, &multiStreamData)
	if err != nil {
		return err
	}

	if err, ok := multiStreamData["error"]; ok {
		hello, ok := err.(map[string]interface{})
		if !ok {
			return errors.New("could not type cast to map[string]interface{}")
		}

		code, ok := hello["code"].(float64)
		if !ok {
			return errors.New("invalid data for error code")
		}

		message, ok := hello["msg"].(string)
		if !ok {
			return errors.New("invalid data for error message")
		}

		return fmt.Errorf("websocket error code [%d], %s",
			int(code),
			message)
	}

	if data, ok := multiStreamData["result"]; ok {
		fmt.Println("RESULT:", data)
	}

	if data, ok := multiStreamData["id"]; ok {
		id, ok := data.(float64)
		if !ok {
			return errors.New("failed to type case to int")
		}

		fmt.Printf("WebsocketIncoming: %d for connection %p\n", int64(id), conn)

		if !b.Websocket.Match.Incoming(int64(id)) {
			log.Warnln(log.WebsocketMgr, "could not match successful subscription")
		}
		return nil
	}

	if method, ok := multiStreamData["method"].(string); ok {
		// TODO handle subscription handling
		if strings.EqualFold(method, "subscribe") {
			return nil
		}
		if strings.EqualFold(method, "unsubscribe") {
			return nil
		}
	}
	if newdata, ok := multiStreamData["data"].(map[string]interface{}); ok {
		if e, ok := newdata["e"].(string); ok {
			switch e {
			case "outboundAccountInfo":
				var data wsAccountInfo
				err := json.Unmarshal(respRaw, &data)
				if err != nil {
					return fmt.Errorf("%v - Could not convert to outboundAccountInfo structure %s",
						b.Name,
						err)
				}
				b.Websocket.DataHandler <- data
			case "outboundAccountPosition":
				var data wsAccountPosition
				err := json.Unmarshal(respRaw, &data)
				if err != nil {
					return fmt.Errorf("%v - Could not convert to outboundAccountPosition structure %s",
						b.Name,
						err)
				}
				b.Websocket.DataHandler <- data
			case "balanceUpdate":
				var data wsBalanceUpdate
				err := json.Unmarshal(respRaw, &data)
				if err != nil {
					return fmt.Errorf("%v - Could not convert to balanceUpdate structure %s",
						b.Name,
						err)
				}
				b.Websocket.DataHandler <- data
			case "executionReport":
				var data wsOrderUpdate
				err := json.Unmarshal(respRaw, &data)
				if err != nil {
					return fmt.Errorf("%v - Could not convert to executionReport structure %s",
						b.Name,
						err)
				}
				var orderID = strconv.FormatInt(data.Data.OrderID, 10)
				oType, err := order.StringToOrderType(data.Data.OrderType)
				if err != nil {
					b.Websocket.DataHandler <- order.ClassificationError{
						Exchange: b.Name,
						OrderID:  orderID,
						Err:      err,
					}
				}
				var oSide order.Side
				oSide, err = order.StringToOrderSide(data.Data.Side)
				if err != nil {
					b.Websocket.DataHandler <- order.ClassificationError{
						Exchange: b.Name,
						OrderID:  orderID,
						Err:      err,
					}
				}
				var oStatus order.Status
				oStatus, err = stringToOrderStatus(data.Data.CurrentExecutionType)
				if err != nil {
					b.Websocket.DataHandler <- order.ClassificationError{
						Exchange: b.Name,
						OrderID:  orderID,
						Err:      err,
					}
				}
				var p currency.Pair
				var a asset.Item
				p, a, err = b.GetRequestFormattedPairAndAssetType(data.Data.Symbol)
				if err != nil {
					return err
				}
				b.Websocket.DataHandler <- &order.Detail{
					Price:           data.Data.Price,
					Amount:          data.Data.Quantity,
					ExecutedAmount:  data.Data.CumulativeFilledQuantity,
					RemainingAmount: data.Data.Quantity - data.Data.CumulativeFilledQuantity,
					Exchange:        b.Name,
					ID:              orderID,
					Type:            oType,
					Side:            oSide,
					Status:          oStatus,
					AssetType:       a,
					Date:            time.Unix(0, data.Data.OrderCreationTime*int64(time.Millisecond)),
					Pair:            p,
				}
			case "listStatus":
				var data wsListStatus
				err := json.Unmarshal(respRaw, &data)
				if err != nil {
					return fmt.Errorf("%v - Could not convert to listStatus structure %s",
						b.Name,
						err)
				}
				b.Websocket.DataHandler <- data
			}
		}
	}
	if wsStream, ok := multiStreamData["stream"].(string); ok {
		streamType := strings.Split(wsStream, "@")
		if len(streamType) > 1 {
			if data, ok := multiStreamData["data"]; ok {
				rawData, err := json.Marshal(data)
				if err != nil {
					return err
				}

				// TODO: Need to infer asset by connection
				pairs, err := b.GetEnabledPairs(asset.Spot)
				if err != nil {
					return err
				}

				format, err := b.GetPairFormat(asset.Spot, true)
				if err != nil {
					return err
				}

				switch streamType[1] {
				case "trade":
					var trade TradeStream
					err := json.Unmarshal(rawData, &trade)
					if err != nil {
						return fmt.Errorf("%v - Could not unmarshal trade data: %s",
							b.Name,
							err)
					}

					price, err := strconv.ParseFloat(trade.Price, 64)
					if err != nil {
						return fmt.Errorf("%v - price conversion error: %s",
							b.Name,
							err)
					}

					amount, err := strconv.ParseFloat(trade.Quantity, 64)
					if err != nil {
						return fmt.Errorf("%v - amount conversion error: %s",
							b.Name,
							err)
					}

					pair, err := currency.NewPairFromFormattedPairs(trade.Symbol, pairs, format)
					if err != nil {
						return err
					}

					assets, err := conn.GetAssetsBySubscriptionType(stream.Trade, pair)
					if err != nil {
						return err
					}

					for i := range assets {
						b.Websocket.DataHandler <- stream.TradeData{
							CurrencyPair: pair,
							Timestamp:    time.Unix(0, trade.TimeStamp*int64(time.Millisecond)),
							Price:        price,
							Amount:       amount,
							Exchange:     b.Name,
							AssetType:    assets[i],
						}
					}
				case "ticker":
					var t TickerStream
					err := json.Unmarshal(rawData, &t)
					if err != nil {
						return fmt.Errorf("%v - Could not convert to a TickerStream structure %s",
							b.Name,
							err.Error())
					}

					pair, err := currency.NewPairFromFormattedPairs(t.Symbol, pairs, format)
					if err != nil {
						return err
					}

					assets, err := conn.GetAssetsBySubscriptionType(stream.Ticker, pair)
					if err != nil {
						return err
					}

					for i := range assets {
						b.Websocket.DataHandler <- &ticker.Price{
							ExchangeName: b.Name,
							Open:         t.OpenPrice,
							Close:        t.ClosePrice,
							Volume:       t.TotalTradedVolume,
							QuoteVolume:  t.TotalTradedQuoteVolume,
							High:         t.HighPrice,
							Low:          t.LowPrice,
							Bid:          t.BestBidPrice,
							Ask:          t.BestAskPrice,
							Last:         t.LastPrice,
							LastUpdated:  time.Unix(0, t.EventTime*int64(time.Millisecond)),
							AssetType:    assets[i],
							Pair:         pair,
						}
					}
				case "kline_1m", "kline_3m", "kline_5m", "kline_15m", "kline_30m", "kline_1h", "kline_2h", "kline_4h",
					"kline_6h", "kline_8h", "kline_12h", "kline_1d", "kline_3d", "kline_1w", "kline_1M":
					var kline KlineStream
					err := json.Unmarshal(rawData, &kline)
					if err != nil {
						return fmt.Errorf("%v - Could not convert to a KlineStream structure %s",
							b.Name,
							err)
					}

					pair, err := currency.NewPairFromFormattedPairs(kline.Symbol, pairs, format)
					if err != nil {
						return err
					}

					assets, err := conn.GetAssetsBySubscriptionType(stream.Kline, pair)
					if err != nil {
						return err
					}

					for i := range assets {
						b.Websocket.DataHandler <- stream.KlineData{
							Timestamp:  time.Unix(0, kline.EventTime*int64(time.Millisecond)),
							Pair:       pair,
							AssetType:  assets[i],
							Exchange:   b.Name,
							StartTime:  time.Unix(0, kline.Kline.StartTime*int64(time.Millisecond)),
							CloseTime:  time.Unix(0, kline.Kline.CloseTime*int64(time.Millisecond)),
							Interval:   kline.Kline.Interval,
							OpenPrice:  kline.Kline.OpenPrice,
							ClosePrice: kline.Kline.ClosePrice,
							HighPrice:  kline.Kline.HighPrice,
							LowPrice:   kline.Kline.LowPrice,
							Volume:     kline.Kline.Volume,
						}
					}
				case "depth":
					var depth WebsocketDepthStream
					err := json.Unmarshal(rawData, &depth)
					if err != nil {
						return fmt.Errorf("%v - Could not convert to depthStream structure %s",
							b.Name,
							err)
					}

					err = b.UpdateLocalBuffer(&depth, conn)
					if err != nil {
						return fmt.Errorf("%v - UpdateLocalCache error: %s",
							b.Name,
							err)
					}
				default:
					b.Websocket.DataHandler <- stream.UnhandledMessageWarning{
						Message: b.Name + stream.UnhandledMessage + string(respRaw),
					}
				}
			}
		}
	}
	return nil
}

func stringToOrderStatus(status string) (order.Status, error) {
	switch status {
	case "NEW":
		return order.New, nil
	case "CANCELLED":
		return order.Cancelled, nil
	case "REJECTED":
		return order.Rejected, nil
	case "TRADE":
		return order.PartiallyFilled, nil
	case "EXPIRED":
		return order.Expired, nil
	default:
		return order.UnknownStatus, errors.New(status + " not recognised as order status")
	}
}

// SeedLocalCache seeds depth data
func (b *Binance) SeedLocalCache(p currency.Pair, a asset.Items) error {
	fPair, err := b.FormatExchangeCurrency(p, a[0])
	if err != nil {
		return err
	}
	ob, err := b.GetOrderBook(OrderBookDataRequestParams{
		Symbol: fPair.String(),
		Limit:  1000,
	})
	if err != nil {
		return err
	}
	for i := range a {
		err = b.SeedLocalCacheWithBook(fPair, &ob, a[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// SeedLocalCacheWithBook seeds the local orderbook cache
func (b *Binance) SeedLocalCacheWithBook(p currency.Pair, orderbookNew *OrderBook, a asset.Item) error {
	var newOrderBook orderbook.Base
	for i := range orderbookNew.Bids {
		newOrderBook.Bids = append(newOrderBook.Bids, orderbook.Item{
			Amount: orderbookNew.Bids[i].Quantity,
			Price:  orderbookNew.Bids[i].Price,
		})
	}
	for i := range orderbookNew.Asks {
		newOrderBook.Asks = append(newOrderBook.Asks, orderbook.Item{
			Amount: orderbookNew.Asks[i].Quantity,
			Price:  orderbookNew.Asks[i].Price,
		})
	}

	newOrderBook.Pair = p
	newOrderBook.AssetType = a
	newOrderBook.ExchangeName = b.Name
	newOrderBook.LastUpdateID = orderbookNew.LastUpdateID

	return b.Websocket.Orderbook.LoadSnapshot(&newOrderBook)
}

// UpdateLocalBuffer stages update to a related asset type associated with a
// connection.
func (b *Binance) UpdateLocalBuffer(u *WebsocketDepthStream, conn stream.Connection) error {
	// TODO: Infer the asset type from the connection
	pairs, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		return err
	}

	pair, err := currency.NewPairFromFormattedPairs(u.Pair, pairs, currency.PairFormat{Uppercase: true})
	if err != nil {
		return err
	}

	assets, err := conn.GetAssetsBySubscriptionType(stream.Orderbook, pair)
	if err != nil {
		return err
	}

	b.mtx.Lock() // protects fetching book
	defer b.mtx.Unlock()

	var errs common.Errors
	for i := range assets {
		// Stage websocket update to buffer
		ch, ok := b.buffer[pair.String()][assets[i]]
		if !ok {
			ch = make(chan *WebsocketDepthStream, 1000)
			b.buffer[pair.String()] = make(map[asset.Item]chan *WebsocketDepthStream)
			b.buffer[pair.String()][assets[i]] = ch
		}
		select {
		case ch <- u:
		default:
			return fmt.Errorf("channel blockage %s", pair)
		}

		if b.fetchingbook[pair.String()][assets[i]] {
			return nil
		}

		err := b.applyBufferUpdate(pair, assets[i], conn)
		if err != nil {
			flushError := b.Websocket.Orderbook.FlushOrderbook(pair, assets[i])
			if flushError != nil {
				log.Errorln(log.WebsocketMgr, "flushing websocket error:", flushError)
			}
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return errs
	}
	return nil
}

// applyBufferUpdate applies the buffer to the orderbook or initiates a new
// orderbook sync by the REST protocol which is off handed to go routine.
func (b *Binance) applyBufferUpdate(cp currency.Pair, a asset.Item, conn stream.Connection) error {
	currentBook := b.Websocket.Orderbook.GetOrderbook(cp, a)
	if currentBook == nil {
		_, ok := b.initialSync[cp.String()]
		if !ok {
			b.initialSync[cp.String()] = make(map[asset.Item]bool)
		}

		b.initialSync[cp.String()][a] = true

		_, ok = b.fetchingbook[cp.String()]
		if !ok {
			b.fetchingbook[cp.String()] = make(map[asset.Item]bool)
		}
		b.fetchingbook[cp.String()][a] = true

		select {
		case b.jobs <- Job{cp, conn}:
		default:
			return errors.New("book synchronisation channel blocked up")
		}
		return nil
	}

loop: // This will continuously remove updates from the buffered channel and
	// apply them to the current orderbook.
	for {
		select {
		case d := <-b.buffer[cp.String()][a]:
			if d.LastUpdateID <= currentBook.LastUpdateID {
				// Drop any event where u is <= lastUpdateId in the snapshot.
				continue
			}
			id := currentBook.LastUpdateID + 1
			if b.initialSync[cp.String()][a] {
				// The first processed event should have U <= lastUpdateId+1 AND
				// u >= lastUpdateId+1.
				if d.FirstUpdateID > id && d.LastUpdateID < id {
					return errors.New("initial sync failure")
				}
				b.initialSync[cp.String()][a] = false
			} else {
				// While listening to the stream, each new event's U should be
				// equal to the previous event's u+1.
				if d.FirstUpdateID != id {
					return fmt.Errorf("synchronisation failure for %s %s", cp, a)
				}
			}
			err := b.ProcessUpdate(cp, a, d)
			if err != nil {
				return err
			}
		default:
			break loop
		}
	}
	return nil
}

// ProcessUpdate processes the websocket orderbook update
func (b *Binance) ProcessUpdate(cp currency.Pair, a asset.Item, ws *WebsocketDepthStream) error {
	var updateBid []orderbook.Item
	for i := range ws.UpdateBids {
		p, err := strconv.ParseFloat(ws.UpdateBids[i][0].(string), 64)
		if err != nil {
			return err
		}
		a, err := strconv.ParseFloat(ws.UpdateBids[i][1].(string), 64)
		if err != nil {
			return err
		}
		updateBid = append(updateBid, orderbook.Item{Price: p, Amount: a})
	}

	var updateAsk []orderbook.Item
	for i := range ws.UpdateAsks {
		p, err := strconv.ParseFloat(ws.UpdateAsks[i][0].(string), 64)
		if err != nil {
			return err
		}
		a, err := strconv.ParseFloat(ws.UpdateAsks[i][1].(string), 64)
		if err != nil {
			return err
		}
		updateAsk = append(updateAsk, orderbook.Item{Price: p, Amount: a})
	}

	return b.Websocket.Orderbook.Update(&buffer.Update{
		Bids:     updateBid,
		Asks:     updateAsk,
		Pair:     cp,
		UpdateID: ws.LastUpdateID,
		Asset:    a,
	})
}

// SynchroniseWebsocketOrderbook synchronises full orderbook for currency pair
// asset
func (b *Binance) SynchroniseWebsocketOrderbook() {
	b.Websocket.Wg.Add(1)
	go func() {
		defer b.Websocket.Wg.Done()
		for {
			select {
			case job := <-b.jobs:
				assets, err := job.Conn.GetAssetsBySubscriptionType(stream.Orderbook, job.Pair)
				if err != nil {
					log.Errorln(log.WebsocketMgr, "cannot fetch asssociated asset types", err)
					continue
				}

				err = b.SeedLocalCache(job.Pair, assets)
				if err != nil {
					log.Errorln(log.WebsocketMgr, "seeding local cache for orderbook error", err)
					continue
				}

				b.mtx.Lock() // immediatly apply the buffer updates so we don't
				// wait for a new update to initiate this.
				for i := range assets {
					err = b.applyBufferUpdate(job.Pair, assets[i], job.Conn)
					if err != nil {
						log.Errorln(log.WebsocketMgr, "applying orderbook updates error", err)
						err = b.Websocket.Orderbook.FlushOrderbook(job.Pair, assets[i])
						if err != nil {
							log.Errorln(log.WebsocketMgr, "flushing websocket error:", err)
						}
						continue
					}

					b.fetchingbook[job.Pair.String()][assets[i]] = false
				}
				b.mtx.Unlock()

			case <-b.Websocket.ShutdownC:
				return
			}
		}
	}()
}

// GenerateSubscriptions generates the default subscription set
func (b *Binance) GenerateSubscriptions(options stream.SubscriptionOptions) ([]stream.ChannelSubscription, error) {
	var channels []WsChannel
	if options.Features.TickerFetching {
		channels = append(channels, WsChannel{
			Definition: "@ticker",
			Type:       stream.Ticker,
		})
	}

	if options.Features.TradeFetching {
		channels = append(channels, WsChannel{
			Definition: "@trade",
			Type:       stream.Trade,
		})
	}

	if options.Features.KlineFetching {
		channels = append(channels, WsChannel{
			Definition: "@kline_1m",
			Type:       stream.Kline,
		})
	}

	if options.Features.OrderbookFetching {
		channels = append(channels, WsChannel{
			Definition: "@depth@100ms",
			Type:       stream.Orderbook,
		})
	}

	var subscriptions []stream.ChannelSubscription
	assets := b.GetAssetTypes()
	for x := range assets {
		pairs, err := b.GetEnabledPairs(assets[x])
		if err != nil {
			return nil, err
		}

		for y := range pairs {
			for z := range channels {
				lp := pairs[y].Lower()
				lp.Delimiter = ""
				channel := lp.String() + channels[z].Definition
				subscriptions = append(subscriptions,
					stream.ChannelSubscription{
						Channel:          channel,
						Currency:         pairs[y],
						Asset:            assets[x],
						SubscriptionType: channels[z].Type,
					})
			}
		}
	}
	return subscriptions, nil
}

// Subscribe subscribes to a set of channels
func (b *Binance) Subscribe(sub stream.SubscriptionParamaters) error {
	fmt.Printf("Subscribing connection: %p, with how many subs %d\n", sub.Conn, len(sub.Items))
	payload := WsPayload{
		Method: "SUBSCRIBE",
	}

	var subbed []stream.ChannelSubscription
	for i := range sub.Items {
		payload.Params = append(payload.Params, sub.Items[i].Channel)
		subbed = append(subbed, sub.Items[i])
		if (i+1)%maxBatchedPayloads != 0 {
			continue
		}

		// payload.ID = time.Now().Unix()
		wow, _ := json.Marshal(payload)
		fmt.Printf("OUTBOUND PAYLOAD: %d ID: %d connect: %p\n", len(wow), payload.ID, sub.Conn)

		err := sub.Conn.SendJSONMessage(payload)
		if err != nil {
			return err
		}

		err = sub.Conn.AddSuccessfulSubscriptions(subbed)
		if err != nil {
			return err
		}
		payload.Params = nil
		subbed = nil
		time.Sleep(250 * time.Millisecond)
	}

	if payload.Params != nil {
		payload.ID = time.Now().Unix()
		err := sub.Conn.SendJSONMessage(payload)
		if err != nil {
			return err
		}

		wow, _ := json.Marshal(payload)
		fmt.Printf("OUTBOUND PAYLOAD: %d ID: %d connect: %p\n", len(wow), payload.ID, sub.Conn)

		err = sub.Conn.AddSuccessfulSubscriptions(subbed)
		if err != nil {
			return err
		}
	}

	return nil
}

// Unsubscribe unsubscribes from a set of channels
func (b *Binance) Unsubscribe(unsub stream.SubscriptionParamaters) error {
	payload := WsPayload{
		Method: "UNSUBSCRIBE",
	}

	var unsubbed []stream.ChannelSubscription
	for i := range unsub.Items {
		payload.Params = append(payload.Params, unsub.Items[i].Channel)
		unsubbed = append(unsubbed, unsub.Items[i])
		if (i+1)%maxBatchedPayloads != 0 {
			continue
		}
		err := unsub.Conn.SendJSONMessage(payload)
		if err != nil {
			return err
		}

		err = unsub.Conn.RemoveSuccessfulUnsubscriptions(unsubbed)
		if err != nil {
			return err
		}
		payload.Params = nil
		unsubbed = nil
	}

	return nil
}
