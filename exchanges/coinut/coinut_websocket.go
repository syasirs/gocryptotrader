package coinut

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/asset"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	log "github.com/thrasher-/gocryptotrader/logger"
)

const coinutWebsocketURL = "wss://wsapi.coinut.com"
const coinutWebsocketRateLimit = 30 * time.Millisecond

var nNonce map[int64]string
var channels map[string]chan []byte
var instrumentListByString map[string]int64
var instrumentListByCode map[int64]string
var populatedList bool

// NOTE for speed considerations
// wss://wsapi-as.coinut.com
// wss://wsapi-na.coinut.com
// wss://wsapi-eu.coinut.com

// WsReadData reads data from the websocket connection
func (c *COINUT) WsReadData() (exchange.WebsocketResponse, error) {
	_, resp, err := c.WebsocketConn.ReadMessage()
	if err != nil {
		return exchange.WebsocketResponse{}, err
	}

	c.Websocket.TrafficAlert <- struct{}{}
	return exchange.WebsocketResponse{Raw: resp}, nil
}

// WsHandleData handles read data
func (c *COINUT) WsHandleData() {
	c.Websocket.Wg.Add(1)

	defer func() {
		c.Websocket.Wg.Done()
	}()

	for {
		select {
		case <-c.Websocket.ShutdownC:
			return

		default:
			resp, err := c.WsReadData()
			if err != nil {
				c.Websocket.DataHandler <- err
				return
			}

			var incoming wsResponse
			err = common.JSONDecode(resp.Raw, &incoming)
			if err != nil {
				c.Websocket.DataHandler <- err
				continue
			}
			switch incoming.Reply {
			case "hb":
				channels["hb"] <- resp.Raw

			case "inst_tick":
				var ticker WsTicker
				err := common.JSONDecode(resp.Raw, &ticker)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				currencyPair := instrumentListByCode[ticker.InstID]
				c.Websocket.DataHandler <- exchange.TickerData{
					Timestamp:  time.Unix(0, ticker.Timestamp),
					Pair:       currency.NewPairFromString(currencyPair),
					Exchange:   c.GetName(),
					AssetType:  asset.Spot,
					HighPrice:  ticker.HighestBuy,
					LowPrice:   ticker.LowestSell,
					ClosePrice: ticker.Last,
					Quantity:   ticker.Volume,
				}

			case "inst_order_book":
				var orderbooksnapshot WsOrderbookSnapshot
				err := common.JSONDecode(resp.Raw, &orderbooksnapshot)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				err = c.WsProcessOrderbookSnapshot(&orderbooksnapshot)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				currencyPair := instrumentListByCode[orderbooksnapshot.InstID]

				c.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
					Exchange: c.GetName(),
					Asset:    asset.Spot,
					Pair:     currency.NewPairFromString(currencyPair),
				}

			case "inst_order_book_update":
				var orderbookUpdate WsOrderbookUpdate
				err := common.JSONDecode(resp.Raw, &orderbookUpdate)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				err = c.WsProcessOrderbookUpdate(&orderbookUpdate)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				currencyPair := instrumentListByCode[orderbookUpdate.InstID]

				c.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
					Exchange: c.GetName(),
					Asset:    asset.Spot,
					Pair:     currency.NewPairFromString(currencyPair),
				}

			case "inst_trade":
				var tradeSnap WsTradeSnapshot
				err := common.JSONDecode(resp.Raw, &tradeSnap)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

			case "inst_trade_update":
				var tradeUpdate WsTradeUpdate
				err := common.JSONDecode(resp.Raw, &tradeUpdate)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				currencyPair := instrumentListByCode[tradeUpdate.InstID]

				c.Websocket.DataHandler <- exchange.TradeData{
					Timestamp:    time.Unix(tradeUpdate.Timestamp, 0),
					CurrencyPair: currency.NewPairFromString(currencyPair),
					AssetType:    asset.Spot,
					Exchange:     c.GetName(),
					Price:        tradeUpdate.Price,
					Side:         tradeUpdate.Side,
				}
			}
		}
	}
}

// WsConnect intiates a websocket connection
func (c *COINUT) WsConnect() error {
	if !c.Websocket.IsEnabled() || !c.IsEnabled() {
		return errors.New(exchange.WebsocketNotEnabled)
	}

	var Dialer websocket.Dialer

	if c.Websocket.GetProxyAddress() != "" {
		proxy, err := url.Parse(c.Websocket.GetProxyAddress())
		if err != nil {
			return err
		}

		Dialer.Proxy = http.ProxyURL(proxy)
	}

	var err error
	c.WebsocketConn, _, err = Dialer.Dial(c.Websocket.GetWebsocketURL(),
		http.Header{})

	if err != nil {
		return err
	}

	if !populatedList {
		instrumentListByString = make(map[string]int64)
		instrumentListByCode = make(map[int64]string)

		err = c.WsSetInstrumentList()
		if err != nil {
			return err
		}
		populatedList = true
	}

	c.GenerateDefaultSubscriptions()

	// define bi-directional communication
	channels = make(map[string]chan []byte)
	channels["hb"] = make(chan []byte, 1)

	go c.WsHandleData()

	return nil
}

// GetNonce returns a nonce for a required request
func (c *COINUT) GetNonce() int64 {
	if c.Nonce.Get() == 0 {
		c.Nonce.Set(time.Now().Unix())
	} else {
		c.Nonce.Inc()
	}

	return int64(c.Nonce.Get())
}

// WsSetInstrumentList fetches instrument list and propagates a local cache
func (c *COINUT) WsSetInstrumentList() error {
	err := c.wsSend(wsRequest{
		Request: "inst_list",
		SecType: "SPOT",
		Nonce:   c.GetNonce(),
	})
	if err != nil {
		return err
	}

	_, resp, err := c.WebsocketConn.ReadMessage()
	if err != nil {
		return err
	}

	c.Websocket.TrafficAlert <- struct{}{}

	var list WsInstrumentList
	err = common.JSONDecode(resp, &list)
	if err != nil {
		return err
	}

	for currency, data := range list.Spot {
		instrumentListByString[currency] = data[0].InstID
		instrumentListByCode[data[0].InstID] = currency
	}

	if len(instrumentListByString) == 0 || len(instrumentListByCode) == 0 {
		return errors.New("instrument lists failed to populate")
	}

	return nil
}

// WsProcessOrderbookSnapshot processes the orderbook snapshot
func (c *COINUT) WsProcessOrderbookSnapshot(ob *WsOrderbookSnapshot) error {
	var bids []orderbook.Item
	for _, bid := range ob.Buy {
		bids = append(bids, orderbook.Item{
			Amount: bid.Volume,
			Price:  bid.Price,
		})
	}

	var asks []orderbook.Item
	for _, ask := range ob.Sell {
		asks = append(asks, orderbook.Item{
			Amount: ask.Volume,
			Price:  ask.Price,
		})
	}

	var newOrderBook orderbook.Base
	newOrderBook.Asks = asks
	newOrderBook.Bids = bids
	newOrderBook.Pair = currency.NewPairFromString(instrumentListByCode[ob.InstID])
	newOrderBook.AssetType = asset.Spot
	newOrderBook.LastUpdated = time.Now()

	return c.Websocket.Orderbook.LoadSnapshot(&newOrderBook, c.GetName(), false)
}

// WsProcessOrderbookUpdate process an orderbook update
func (c *COINUT) WsProcessOrderbookUpdate(ob *WsOrderbookUpdate) error {
	p := currency.NewPairFromString(instrumentListByCode[ob.InstID])

	if ob.Side == exchange.BuyOrderSide.ToLower().ToString() {
		return c.Websocket.Orderbook.Update([]orderbook.Item{
			{Price: ob.Price, Amount: ob.Volume}},
			nil,
			p,
			time.Now(),
			c.GetName(),
			asset.Spot)
	}

	return c.Websocket.Orderbook.Update([]orderbook.Item{
		{Price: ob.Price, Amount: ob.Volume}},
		nil,
		p,
		time.Now(),
		c.GetName(),
		asset.Spot)
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (c *COINUT) GenerateDefaultSubscriptions() {
	var channels = []string{"inst_tick", "inst_order_book"}
	subscriptions := []exchange.WebsocketChannelSubscription{}
	enabledCurrencies := c.GetEnabledPairs(asset.Spot)
	for i := range channels {
		for j := range enabledCurrencies {
			subscriptions = append(subscriptions, exchange.WebsocketChannelSubscription{
				Channel:  channels[i],
				Currency: enabledCurrencies[j],
			})
		}
	}
	c.Websocket.SubscribeToChannels(subscriptions)
}

// Subscribe sends a websocket message to receive data from the channel
func (c *COINUT) Subscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	subscribe := wsRequest{
		Request:   channelToSubscribe.Channel,
		InstID:    instrumentListByString[channelToSubscribe.Currency.String()],
		Subscribe: true,
		Nonce:     c.GetNonce(),
	}
	return c.wsSend(subscribe)
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (c *COINUT) Unsubscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	subscribe := wsRequest{
		Request:   channelToSubscribe.Channel,
		InstID:    instrumentListByString[channelToSubscribe.Currency.String()],
		Subscribe: false,
		Nonce:     c.GetNonce(),
	}
	return c.wsSend(subscribe)
}

// WsSend sends data to the websocket server
func (c *COINUT) wsSend(data interface{}) error {
	c.wsRequestMtx.Lock()
	defer c.wsRequestMtx.Unlock()

	json, err := common.JSONEncode(data)
	if err != nil {
		return err
	}
	if c.Verbose {
		log.Debugf("%v sending message to websocket %v", c.Name, string(json))
	}
	// Basic rate limiter
	time.Sleep(coinutWebsocketRateLimit)
	return c.WebsocketConn.WriteMessage(websocket.TextMessage, json)
}
