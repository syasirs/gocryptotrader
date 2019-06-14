package coinbasepro

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	log "github.com/thrasher-/gocryptotrader/logger"
)

const (
	coinbaseproWebsocketURL = "wss://ws-feed.pro.coinbase.com"
)

// WsConnect initiates a websocket connection
func (c *CoinbasePro) WsConnect() error {
	if !c.Websocket.IsEnabled() || !c.IsEnabled() {
		return errors.New(exchange.WebsocketNotEnabled)
	}

	var dialer websocket.Dialer

	if c.Websocket.GetProxyAddress() != "" {
		proxy, err := url.Parse(c.Websocket.GetProxyAddress())
		if err != nil {
			return fmt.Errorf("coinbasepro_websocket.go error - proxy address %s",
				err)
		}

		dialer.Proxy = http.ProxyURL(proxy)
	}

	var err error
	c.WebsocketConn, _, err = dialer.Dial(c.Websocket.GetWebsocketURL(),
		http.Header{})
	if err != nil {
		return fmt.Errorf("coinbasepro_websocket.go error - unable to connect to websocket %s",
			err)
	}

	c.GenerateDefaultSubscriptions()
	go c.WsHandleData()

	return nil
}

// WsReadData reads data from the websocket connection
func (c *CoinbasePro) WsReadData() (exchange.WebsocketResponse, error) {
	_, resp, err := c.WebsocketConn.ReadMessage()
	if err != nil {
		return exchange.WebsocketResponse{}, err
	}
	c.Websocket.TrafficAlert <- struct{}{}
	return exchange.WebsocketResponse{Raw: resp}, nil
}

// WsHandleData handles read data from websocket connection
func (c *CoinbasePro) WsHandleData() {
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
			type MsgType struct {
				Type      string `json:"type"`
				Sequence  int64  `json:"sequence"`
				ProductID string `json:"product_id"`
			}

			msgType := MsgType{}
			err = common.JSONDecode(resp.Raw, &msgType)
			if err != nil {
				c.Websocket.DataHandler <- err
				continue
			}

			if msgType.Type == "subscriptions" || msgType.Type == "heartbeat" {
				continue
			}

			switch msgType.Type {
			case "error":
				c.Websocket.DataHandler <- errors.New(string(resp.Raw))

			case "ticker":
				ticker := WebsocketTicker{}
				err := common.JSONDecode(resp.Raw, &ticker)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				c.Websocket.DataHandler <- exchange.TickerData{
					Timestamp:  ticker.Time,
					Pair:       currency.NewPairFromString(ticker.ProductID),
					AssetType:  "SPOT",
					Exchange:   c.GetName(),
					OpenPrice:  ticker.Open24H,
					HighPrice:  ticker.High24H,
					LowPrice:   ticker.Low24H,
					ClosePrice: ticker.Price,
					Quantity:   ticker.Volume24H,
				}

			case "snapshot":
				snapshot := WebsocketOrderbookSnapshot{}
				err := common.JSONDecode(resp.Raw, &snapshot)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				err = c.ProcessSnapshot(&snapshot)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

			case "l2update":
				update := WebsocketL2Update{}
				err := common.JSONDecode(resp.Raw, &update)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}

				err = c.ProcessUpdate(update)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
			case "received":
				// We currently use l2update to calculate orderbook changes
				received := WebsocketReceived{}
				err := common.JSONDecode(resp.Raw, &received)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
				c.Websocket.DataHandler <- received
			case "open":
				// We currently use l2update to calculate orderbook changes
				open := WebsocketOpen{}
				err := common.JSONDecode(resp.Raw, &open)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
				c.Websocket.DataHandler <- open
			case "done":
				// We currently use l2update to calculate orderbook changes
				done := WebsocketDone{}
				err := common.JSONDecode(resp.Raw, &done)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
				c.Websocket.DataHandler <- done
			case "change":
				// We currently use l2update to calculate orderbook changes
				change := WebsocketChange{}
				err := common.JSONDecode(resp.Raw, &change)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
				c.Websocket.DataHandler <- change
			case "activate":
				// We currently use l2update to calculate orderbook changes
				activate := WebsocketActivate{}
				err := common.JSONDecode(resp.Raw, &activate)
				if err != nil {
					c.Websocket.DataHandler <- err
					continue
				}
				c.Websocket.DataHandler <- activate
			}
		}
	}
}

// ProcessSnapshot processes the initial orderbook snap shot
func (c *CoinbasePro) ProcessSnapshot(snapshot *WebsocketOrderbookSnapshot) error {
	var base orderbook.Base
	for _, bid := range snapshot.Bids {
		price, err := strconv.ParseFloat(bid[0].(string), 64)
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(bid[1].(string), 64)
		if err != nil {
			return err
		}

		base.Bids = append(base.Bids,
			orderbook.Item{Price: price, Amount: amount})
	}

	for _, ask := range snapshot.Asks {
		price, err := strconv.ParseFloat(ask[0].(string), 64)
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(ask[1].(string), 64)
		if err != nil {
			return err
		}

		base.Asks = append(base.Asks,
			orderbook.Item{Price: price, Amount: amount})
	}

	pair := currency.NewPairFromString(snapshot.ProductID)
	base.AssetType = "SPOT"
	base.Pair = pair

	err := c.Websocket.Orderbook.LoadSnapshot(&base, c.GetName(), false)
	if err != nil {
		return err
	}

	c.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
		Pair:     pair,
		Asset:    "SPOT",
		Exchange: c.GetName(),
	}

	return nil
}

// ProcessUpdate updates the orderbook local cache
func (c *CoinbasePro) ProcessUpdate(update WebsocketL2Update) error {
	var Asks, Bids []orderbook.Item

	for _, data := range update.Changes {
		price, _ := strconv.ParseFloat(data[1].(string), 64)
		volume, _ := strconv.ParseFloat(data[2].(string), 64)

		if data[0].(string) == "buy" {
			Bids = append(Bids, orderbook.Item{Price: price, Amount: volume})
		} else {
			Asks = append(Asks, orderbook.Item{Price: price, Amount: volume})
		}
	}

	if len(Asks) == 0 && len(Bids) == 0 {
		return errors.New("coibasepro_websocket.go error - no data in websocket update")
	}

	p := currency.NewPairFromString(update.ProductID)

	err := c.Websocket.Orderbook.Update(Bids, Asks, p, time.Now(), c.GetName(), "SPOT")
	if err != nil {
		return err
	}

	c.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
		Pair:     p,
		Asset:    "SPOT",
		Exchange: c.GetName(),
	}

	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (c *CoinbasePro) GenerateDefaultSubscriptions() {
	var channels = []string{"heartbeat", "level2", "ticker", "user"}
	enabledCurrencies := c.GetEnabledCurrencies()
	var subscriptions []exchange.WebsocketChannelSubscription
	for i := range channels {
		if (channels[i] == "user" || channels[i] == "full") && !c.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
			continue
		}
		for j := range enabledCurrencies {
			enabledCurrencies[j].Delimiter = "-"
			subscriptions = append(subscriptions, exchange.WebsocketChannelSubscription{
				Channel:  channels[i],
				Currency: enabledCurrencies[j],
			})
		}
	}
	c.Websocket.SubscribeToChannels(subscriptions)
}

// Subscribe sends a websocket message to receive data from the channel
func (c *CoinbasePro) Subscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	subscribe := WebsocketSubscribe{
		Type: "subscribe",
		Channels: []WsChannels{
			{
				Name: channelToSubscribe.Channel,
				ProductIDs: []string{
					channelToSubscribe.Currency.String(),
				},
			},
		},
	}
	if channelToSubscribe.Channel == "user" || channelToSubscribe.Channel == "full" {
		n := fmt.Sprintf("%v", time.Now().Unix())
		message := n + "GET" + "/users/self/verify"
		hmac := common.GetHMAC(common.HashSHA256, []byte(message), []byte(c.APISecret))
		subscribe.Signature = common.Base64Encode(hmac)
		subscribe.Key = c.APIKey
		subscribe.Passphrase = c.ClientID
		subscribe.Timestamp = n
	}
	return c.wsSend(subscribe)
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (c *CoinbasePro) Unsubscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	subscribe := WebsocketSubscribe{
		Type: "unsubscribe",
		Channels: []WsChannels{
			{
				Name: channelToSubscribe.Channel,
				ProductIDs: []string{
					channelToSubscribe.Currency.String(),
				},
			},
		},
	}
	return c.wsSend(subscribe)
}

// WsSend sends data to the websocket server
func (c *CoinbasePro) wsSend(data interface{}) error {
	c.wsRequestMtx.Lock()
	defer c.wsRequestMtx.Unlock()
	if c.Verbose {
		log.Debugf("%v sending message to websocket %v", c.Name, data)
	}
	json, err := common.JSONEncode(data)
	if err != nil {
		return err
	}
	return c.WebsocketConn.WriteMessage(websocket.TextMessage, json)
}
