package coinbasepro

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wsorderbook"
)

const (
	coinbaseproWebsocketURL = "wss://ws-feed.pro.coinbase.com"
)

// WsConnect initiates a websocket connection
func (c *CoinbasePro) WsConnect() error {
	if !c.Websocket.IsEnabled() || !c.IsEnabled() {
		return errors.New(wshandler.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := c.WebsocketConn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}

	c.GenerateDefaultSubscriptions()
	go c.wsReadData()

	return nil
}

// wsReadData handles read data from websocket connection
func (c *CoinbasePro) wsReadData() {
	c.Websocket.Wg.Add(1)

	defer func() {
		c.Websocket.Wg.Done()
	}()

	for {
		select {
		case <-c.Websocket.ShutdownC:
			return
		default:
			resp, err := c.WebsocketConn.ReadMessage()
			if err != nil {
				c.Websocket.ReadMessageErrors <- err
				return
			}
			c.Websocket.TrafficAlert <- struct{}{}
			err = c.wsHandleData(resp.Raw)
			if err != nil {
				c.Websocket.DataHandler <- err
			}

		}
	}
}

func (c *CoinbasePro) wsHandleData(respRaw []byte) error {
	type MsgType struct {
		Type      string `json:"type"`
		Sequence  int64  `json:"sequence"`
		ProductID string `json:"product_id"`
	}

	msgType := MsgType{}
	err := json.Unmarshal(respRaw, &msgType)
	if err != nil {
		return err
	}

	if msgType.Type == "subscriptions" || msgType.Type == "heartbeat" {
		return nil
	}

	switch msgType.Type {
	case "error":
		c.Websocket.DataHandler <- errors.New(string(respRaw))

	case "ticker":
		wsTicker := WebsocketTicker{}
		err := json.Unmarshal(respRaw, &wsTicker)
		if err != nil {
			return err
		}

		c.Websocket.DataHandler <- &ticker.Price{
			LastUpdated:  wsTicker.Time,
			Pair:         wsTicker.ProductID,
			AssetType:    asset.Spot,
			ExchangeName: c.Name,
			Open:         wsTicker.Open24H,
			High:         wsTicker.High24H,
			Low:          wsTicker.Low24H,
			Last:         wsTicker.Price,
			Volume:       wsTicker.Volume24H,
			Bid:          wsTicker.BestBid,
			Ask:          wsTicker.BestAsk,
		}

	case "snapshot":
		snapshot := WebsocketOrderbookSnapshot{}
		err := json.Unmarshal(respRaw, &snapshot)
		if err != nil {
			return err
		}

		err = c.ProcessSnapshot(&snapshot)
		if err != nil {
			return err
		}

	case "l2update":
		update := WebsocketL2Update{}
		err := json.Unmarshal(respRaw, &update)
		if err != nil {
			return err
		}

		err = c.ProcessUpdate(update)
		if err != nil {
			return err
		}
		// the following cases contains data to synchronise authenticated orders
		// subscribing to the "full" channel will consider ALL cbp orders as
		// personal orders
		// remove sending &order.Detail to the datahandler if you wish to subscribe to the
		// "full" channel
	case "received", "open", "done", "change", "activate":
		var wsOrder wsOrderReceived
		err := json.Unmarshal(respRaw, &wsOrder)
		if err != nil {
			return err
		}
		createdDate, err := time.Parse(wsOrder.Time, time.RFC3339)
		if err != nil {
			c.Websocket.DataHandler <- err
			createdDate = time.Now()
		}
		oType, err := order.StringToOrderType(wsOrder.OrderType)
		if err != nil {
			c.Websocket.DataHandler <- err
		}
		oSide, err := order.StringToOrderSide(wsOrder.Side)
		if err != nil {
			c.Websocket.DataHandler <- err
		}
		oStatus := statusToStandardStatus(wsOrder.Type)
		if wsOrder.Reason == "canceled" {
			oStatus = order.Cancelled
		}
		c.Websocket.DataHandler <- &order.Detail{
			HiddenOrder:     wsOrder.Private,
			Price:           wsOrder.Price,
			Amount:          wsOrder.Size,
			LimitPriceUpper: 0,
			LimitPriceLower: 0,
			TriggerPrice:    wsOrder.StopPrice,
			TargetAmount:    0,
			ExecutedAmount:  wsOrder.Size - wsOrder.RemainingSize,
			RemainingAmount: wsOrder.RemainingSize,
			Fee:             wsOrder.TakerFeeRate,
			Exchange:        c.Name,
			ID:              wsOrder.OrderID,
			AccountID:       wsOrder.ProfileID,
			ClientID:        c.API.Credentials.ClientID,
			Type:            oType,
			Side:            oSide,
			Status:          oStatus,
			AssetType:       asset.Spot,
			Date:            createdDate,
			Pair:            currency.NewPairFromString(wsOrder.ProductID),
		}
	}
	return fmt.Errorf("%v Unhandled websocket message %s", c.Name, respRaw)
}

func statusToStandardStatus(stat string) order.Status {
	switch stat {
	case "received":
		return order.New
	case "open":
		return order.Active
	case "done":
		return order.Filled
	case "match":
		return order.PartiallyFilled
	case "change":
		return order.Active
	case "activate":
		return order.Rejected
	default:
		return order.UnknownStatus
	}
}

// ProcessSnapshot processes the initial orderbook snap shot
func (c *CoinbasePro) ProcessSnapshot(snapshot *WebsocketOrderbookSnapshot) error {
	var base orderbook.Base
	for i := range snapshot.Bids {
		price, err := strconv.ParseFloat(snapshot.Bids[i][0].(string), 64)
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(snapshot.Bids[i][1].(string), 64)
		if err != nil {
			return err
		}

		base.Bids = append(base.Bids,
			orderbook.Item{Price: price, Amount: amount})
	}

	for i := range snapshot.Asks {
		price, err := strconv.ParseFloat(snapshot.Asks[i][0].(string), 64)
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(snapshot.Asks[i][1].(string), 64)
		if err != nil {
			return err
		}

		base.Asks = append(base.Asks,
			orderbook.Item{Price: price, Amount: amount})
	}

	pair := currency.NewPairFromString(snapshot.ProductID)
	base.AssetType = asset.Spot
	base.Pair = pair
	base.ExchangeName = c.Name

	err := c.Websocket.Orderbook.LoadSnapshot(&base)
	if err != nil {
		return err
	}

	c.Websocket.DataHandler <- wshandler.WebsocketOrderbookUpdate{
		Pair:     pair,
		Asset:    asset.Spot,
		Exchange: c.Name,
	}

	return nil
}

// ProcessUpdate updates the orderbook local cache
func (c *CoinbasePro) ProcessUpdate(update WebsocketL2Update) error {
	var asks, bids []orderbook.Item

	for i := range update.Changes {
		price, _ := strconv.ParseFloat(update.Changes[i][1].(string), 64)
		volume, _ := strconv.ParseFloat(update.Changes[i][2].(string), 64)

		if update.Changes[i][0].(string) == order.Buy.Lower() {
			bids = append(bids, orderbook.Item{Price: price, Amount: volume})
		} else {
			asks = append(asks, orderbook.Item{Price: price, Amount: volume})
		}
	}

	if len(asks) == 0 && len(bids) == 0 {
		return errors.New("coinbasepro_websocket.go error - no data in websocket update")
	}

	p := currency.NewPairFromString(update.ProductID)
	timestamp, err := time.Parse(time.RFC3339, update.Time)
	if err != nil {
		return err
	}
	err = c.Websocket.Orderbook.Update(&wsorderbook.WebsocketOrderbookUpdate{
		Bids:       bids,
		Asks:       asks,
		Pair:       p,
		UpdateTime: timestamp,
		Asset:      asset.Spot,
	})
	if err != nil {
		return err
	}

	c.Websocket.DataHandler <- wshandler.WebsocketOrderbookUpdate{
		Pair:     p,
		Asset:    asset.Spot,
		Exchange: c.Name,
	}

	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (c *CoinbasePro) GenerateDefaultSubscriptions() {
	var channels = []string{"heartbeat", "level2", "ticker", "user"}
	enabledCurrencies := c.GetEnabledPairs(asset.Spot)
	var subscriptions []wshandler.WebsocketChannelSubscription
	for i := range channels {
		if (channels[i] == "user" || channels[i] == "full") && !c.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
			continue
		}
		for j := range enabledCurrencies {
			subscriptions = append(subscriptions, wshandler.WebsocketChannelSubscription{
				Channel: channels[i],
				Currency: c.FormatExchangeCurrency(enabledCurrencies[j],
					asset.Spot),
			})
		}
	}
	c.Websocket.SubscribeToChannels(subscriptions)
}

// Subscribe sends a websocket message to receive data from the channel
func (c *CoinbasePro) Subscribe(channelToSubscribe wshandler.WebsocketChannelSubscription) error {
	subscribe := WebsocketSubscribe{
		Type: "subscribe",
		Channels: []WsChannels{
			{
				Name: channelToSubscribe.Channel,
				ProductIDs: []string{
					c.FormatExchangeCurrency(channelToSubscribe.Currency,
						asset.Spot).String(),
				},
			},
		},
	}
	if channelToSubscribe.Channel == "user" || channelToSubscribe.Channel == "full" {
		n := strconv.FormatInt(time.Now().Unix(), 10)
		message := n + "GET" + "/users/self/verify"
		hmac := crypto.GetHMAC(crypto.HashSHA256, []byte(message),
			[]byte(c.API.Credentials.Secret))
		subscribe.Signature = crypto.Base64Encode(hmac)
		subscribe.Key = c.API.Credentials.Key
		subscribe.Passphrase = c.API.Credentials.ClientID
		subscribe.Timestamp = n
	}
	return c.WebsocketConn.SendJSONMessage(subscribe)
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (c *CoinbasePro) Unsubscribe(channelToSubscribe wshandler.WebsocketChannelSubscription) error {
	subscribe := WebsocketSubscribe{
		Type: "unsubscribe",
		Channels: []WsChannels{
			{
				Name: channelToSubscribe.Channel,
				ProductIDs: []string{
					c.FormatExchangeCurrency(channelToSubscribe.Currency,
						asset.Spot).String(),
				},
			},
		},
	}
	return c.WebsocketConn.SendJSONMessage(subscribe)
}
