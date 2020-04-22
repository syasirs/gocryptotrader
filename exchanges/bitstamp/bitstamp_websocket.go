package bitstamp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	bitstampWSURL = "wss://ws.bitstamp.net"
)

// WsConnect connects to a websocket feed
func (b *Bitstamp) WsConnect() error {
	fmt.Println("MEOW")
	if !b.Websocket.IsEnabled() || !b.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := b.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	if b.Verbose {
		log.Debugf(log.ExchangeSys, "%s Connected to Websocket.\n", b.Name)
	}
	err = b.seedOrderBook()
	if err != nil {
		b.Websocket.DataHandler <- err
	}
	subs, err := b.generateDefaultSubscriptions()
	if err != nil {
		return err
	}
	go b.wsReadData()
	b.Websocket.SubscribeToChannels(subs)
	return nil
}

// wsReadData receives and passes on websocket messages for processing
func (b *Bitstamp) wsReadData() {
	b.Websocket.Wg.Add(1)
	defer func() {
		b.Websocket.Wg.Done()
	}()
	for {
		resp, err := b.Websocket.Conn.ReadMessage()
		if err != nil {
			return
		}
		err = b.wsHandleData(resp.Raw)
		if err != nil {
			b.Websocket.DataHandler <- err
		}
	}
}

func (b *Bitstamp) wsHandleData(respRaw []byte) error {
	var wsResponse websocketResponse
	err := json.Unmarshal(respRaw, &wsResponse)
	if err != nil {
		return err
	}

	switch wsResponse.Event {
	case "bts:subscribe":
		if b.Verbose {
			log.Debugf(log.ExchangeSys, "%v - Websocket subscription acknowledgement", b.Name)
		}
	case "bts:unsubscribe":
		if b.Verbose {
			log.Debugf(log.ExchangeSys, "%v - Websocket unsubscribe acknowledgement", b.Name)
		}
	case "bts:request_reconnect":
		if b.Verbose {
			log.Debugf(log.ExchangeSys, "%v - Websocket reconnection request received", b.Name)
		}
		go b.Websocket.Shutdown() // Connection monitor will reconnect
	case "data":
		wsOrderBookTemp := websocketOrderBookResponse{}
		err := json.Unmarshal(respRaw, &wsOrderBookTemp)
		if err != nil {
			return err
		}
		currencyPair := strings.Split(wsResponse.Channel, "_")
		p, err := currency.NewPairFromString(strings.ToUpper(currencyPair[2]))
		if err != nil {
			return err
		}

		err = b.wsUpdateOrderbook(wsOrderBookTemp.Data, p, asset.Spot)
		if err != nil {
			return err
		}
	case "trade":
		wsTradeTemp := websocketTradeResponse{}
		err := json.Unmarshal(respRaw, &wsTradeTemp)
		if err != nil {
			return err
		}
		currencyPair := strings.Split(wsResponse.Channel, "_")
		p, err := currency.NewPairFromString(strings.ToUpper(currencyPair[2]))
		if err != nil {
			return err
		}

		side := order.Buy
		if wsTradeTemp.Data.Type == -1 {
			side = order.Sell
		}
		var a asset.Item
		a, err = b.GetPairAssetType(p)
		if err != nil {
			return err
		}
		b.Websocket.DataHandler <- stream.TradeData{
			Timestamp:    time.Unix(wsTradeTemp.Data.Timestamp, 0),
			CurrencyPair: p,
			AssetType:    a,
			Exchange:     b.Name,
			EventType:    order.UnknownType,
			Price:        wsTradeTemp.Data.Price,
			Amount:       wsTradeTemp.Data.Amount,
			Side:         side,
		}
	case "order_created", "order_deleted", "order_changed":
		if b.Verbose {
			log.Debugf(log.ExchangeSys, "%v - Websocket order acknowledgement", b.Name)
		}
	default:
		b.Websocket.DataHandler <- stream.UnhandledMessageWarning{Message: b.Name + stream.UnhandledMessage + string(respRaw)}
	}
	return nil
}

func (b *Bitstamp) generateDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	fmt.Println("WOW")
	var channels = []string{"live_trades_", "order_book_"}
	enabledCurrencies, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		return nil, err
	}
	var subscriptions []stream.ChannelSubscription
	for i := range channels {
		for j := range enabledCurrencies {
			subscriptions = append(subscriptions, stream.ChannelSubscription{
				Channel: channels[i] + enabledCurrencies[j].Lower().String(),
			})
		}
	}
	return subscriptions, nil
}

// Subscribe sends a websocket message to receive data from the channel
func (b *Bitstamp) Subscribe(channelsToSubscribe []stream.ChannelSubscription) error {
	fmt.Println("SUB:", channelsToSubscribe)
	for i := range channelsToSubscribe {
		req := websocketEventRequest{
			Event: "bts:subscribe",
			Data: websocketData{
				Channel: channelsToSubscribe[i].Channel,
			},
		}
		err := b.Websocket.Conn.SendJSONMessage(req)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (b *Bitstamp) Unsubscribe(channelsToUnsubscribe []stream.ChannelSubscription) error {
	fmt.Println("UNSUB:", channelsToUnsubscribe)
	for i := range channelsToUnsubscribe {
		req := websocketEventRequest{
			Event: "bts:unsubscribe",
			Data: websocketData{
				Channel: channelsToUnsubscribe[i].Channel,
			},
		}
		err := b.Websocket.Conn.SendJSONMessage(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bitstamp) wsUpdateOrderbook(update websocketOrderBook, p currency.Pair, assetType asset.Item) error {
	if len(update.Asks) == 0 && len(update.Bids) == 0 {
		return errors.New("bitstamp_websocket.go error - no orderbook data")
	}
	var asks, bids []orderbook.Item
	for i := range update.Asks {
		target, err := strconv.ParseFloat(update.Asks[i][0], 64)
		if err != nil {
			b.Websocket.DataHandler <- err
			continue
		}
		amount, err := strconv.ParseFloat(update.Asks[i][1], 64)
		if err != nil {
			b.Websocket.DataHandler <- err
			continue
		}
		asks = append(asks, orderbook.Item{Price: target, Amount: amount})
	}
	for i := range update.Bids {
		target, err := strconv.ParseFloat(update.Bids[i][0], 64)
		if err != nil {
			b.Websocket.DataHandler <- err
			continue
		}
		amount, err := strconv.ParseFloat(update.Bids[i][1], 64)
		if err != nil {
			b.Websocket.DataHandler <- err
			continue
		}

		bids = append(bids, orderbook.Item{Price: target, Amount: amount})
	}
	return b.Websocket.Orderbook.LoadSnapshot(&orderbook.Base{
		Bids:         bids,
		Asks:         asks,
		Pair:         p,
		LastUpdated:  time.Unix(update.Timestamp, 0),
		AssetType:    asset.Spot,
		ExchangeName: b.Name,
	})
}

func (b *Bitstamp) seedOrderBook() error {
	p, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		return err
	}

	for x := range p {
		orderbookSeed, err := b.GetOrderbook(p[x].String())
		if err != nil {
			return err
		}

		var newOrderBook orderbook.Base
		for i := range orderbookSeed.Asks {
			newOrderBook.Asks = append(newOrderBook.Asks, orderbook.Item{
				Price:  orderbookSeed.Asks[i].Price,
				Amount: orderbookSeed.Asks[i].Amount,
			})
		}
		for i := range orderbookSeed.Bids {
			newOrderBook.Bids = append(newOrderBook.Bids, orderbook.Item{
				Price:  orderbookSeed.Bids[i].Price,
				Amount: orderbookSeed.Bids[i].Amount,
			})
		}
		newOrderBook.Pair = p[x]
		newOrderBook.AssetType = asset.Spot
		newOrderBook.ExchangeName = b.Name

		err = b.Websocket.Orderbook.LoadSnapshot(&newOrderBook)
		if err != nil {
			return err
		}
	}
	return nil
}
