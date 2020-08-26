package btse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

const (
	btseWebsocket      = "wss://ws.btse.com/spotWS"
	btseWebsocketTimer = time.Second * 57
)

// WsConnect connects the websocket client
func (b *BTSE) WsConnect() error {
	if !b.Websocket.IsEnabled() || !b.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := b.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	b.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		MessageType: websocket.PingMessage,
		Delay:       btseWebsocketTimer,
	})

	go b.wsReadData()
	if b.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
		err = b.WsAuthenticate()
		if err != nil {
			b.Websocket.DataHandler <- err
			b.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
	}

	subs, err := b.GenerateDefaultSubscriptions()
	if err != nil {
		return err
	}
	return b.Websocket.SubscribeToChannels(subs)
}

// WsAuthenticate Send an authentication message to receive auth data
func (b *BTSE) WsAuthenticate() error {
	nonce := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	path := "/spotWS" + nonce
	hmac := crypto.GetHMAC(crypto.HashSHA512_384,
		[]byte((path)),
		[]byte(b.API.Credentials.Secret),
	)
	sign := crypto.HexEncodeToString(hmac)
	req := wsSub{
		Operation: "authKeyExpires",
		Arguments: []string{b.API.Credentials.Key, nonce, sign},
	}
	return b.Websocket.Conn.SendJSONMessage(req)
}

func stringToOrderStatus(status string) (order.Status, error) {
	switch status {
	case "ORDER_INSERTED", "TRIGGER_INSERTED":
		return order.New, nil
	case "ORDER_CANCELLED":
		return order.Cancelled, nil
	case "ORDER_FULL_TRANSACTED":
		return order.Filled, nil
	case "ORDER_PARTIALLY_TRANSACTED":
		return order.PartiallyFilled, nil
	case "TRIGGER_ACTIVATED":
		return order.Active, nil
	case "INSUFFICIENT_BALANCE":
		return order.InsufficientBalance, nil
	case "MARKET_UNAVAILABLE":
		return order.MarketUnavailable, nil
	default:
		return order.UnknownStatus, errors.New(status + " not recognised as order status")
	}
}

// wsReadData receives and passes on websocket messages for processing
func (b *BTSE) wsReadData() {
	b.Websocket.Wg.Add(1)
	defer b.Websocket.Wg.Done()

	for {
		resp := b.Websocket.Conn.ReadMessage()
		if resp.Raw == nil {
			return
		}
		err := b.wsHandleData(resp.Raw)
		if err != nil {
			b.Websocket.DataHandler <- err
		}
	}
}

func (b *BTSE) wsHandleData(respRaw []byte) error {
	type Result map[string]interface{}
	var result Result
	err := json.Unmarshal(respRaw, &result)
	if err != nil {
		if strings.Contains(string(respRaw), "UNLOGIN_USER connect success") ||
			strings.Contains(string(respRaw), "authenticated successfully") {
			return nil
		} else if strings.Contains(string(respRaw), "AUTHENTICATE ERROR") {
			b.Websocket.SetCanUseAuthenticatedEndpoints(false)
			return errors.New("authentication failure")
		}
		return err
	}
	switch {
	case result["topic"] == "notificationApi":
		var notification wsNotification
		err = json.Unmarshal(respRaw, &notification)
		if err != nil {
			return err
		}
		for i := range notification.Data {
			var oType order.Type
			var oSide order.Side
			var oStatus order.Status
			oType, err = order.StringToOrderType(notification.Data[i].Type)
			if err != nil {
				b.Websocket.DataHandler <- order.ClassificationError{
					Exchange: b.Name,
					OrderID:  notification.Data[i].OrderID,
					Err:      err,
				}
			}
			oSide, err = order.StringToOrderSide(notification.Data[i].OrderMode)
			if err != nil {
				b.Websocket.DataHandler <- order.ClassificationError{
					Exchange: b.Name,
					OrderID:  notification.Data[i].OrderID,
					Err:      err,
				}
			}
			oStatus, err = stringToOrderStatus(notification.Data[i].Status)
			if err != nil {
				b.Websocket.DataHandler <- order.ClassificationError{
					Exchange: b.Name,
					OrderID:  notification.Data[i].OrderID,
					Err:      err,
				}
			}

			var p currency.Pair
			p, err = currency.NewPairFromString(notification.Data[i].Symbol)
			if err != nil {
				return err
			}

			var a asset.Item
			a, err = b.GetPairAssetType(p)
			if err != nil {
				return err
			}

			b.Websocket.DataHandler <- &order.Detail{
				Price:        notification.Data[i].Price,
				Amount:       notification.Data[i].Size,
				TriggerPrice: notification.Data[i].TriggerPrice,
				Exchange:     b.Name,
				ID:           notification.Data[i].OrderID,
				Type:         oType,
				Side:         oSide,
				Status:       oStatus,
				AssetType:    a,
				Date:         time.Unix(0, notification.Data[i].Timestamp*int64(time.Millisecond)),
				Pair:         p,
			}
		}
	case strings.Contains(result["topic"].(string), "tradeHistory"):
		// auth trade history endpoint
		var tradeHistory wsTradeHistory
		err = json.Unmarshal(respRaw, &tradeHistory)
		if err != nil {
			return err
		}
		var trades []trade.Data
		for x := range tradeHistory.Data {
			side := order.Buy
			if tradeHistory.Data[x].Gain == -1 {
				side = order.Sell
			}

			var p currency.Pair
			p, err = currency.NewPairFromString(strings.Replace(tradeHistory.Topic,
				"tradeHistory:",
				"",
				1))
			if err != nil {
				return err
			}
			var a asset.Item
			a, err = b.GetPairAssetType(p)
			if err != nil {
				return err
			}
			trades = append(trades, trade.Data{
				Timestamp:    time.Unix(0, tradeHistory.Data[x].TransactionTime*int64(time.Millisecond)),
				CurrencyPair: p,
				AssetType:    a,
				Exchange:     b.Name,
				Price:        tradeHistory.Data[x].Price,
				Amount:       tradeHistory.Data[x].Amount,
				Side:         side,
			})
		}
		return trade.AddTradesToBuffer(b.Name, trades...)
	case strings.Contains(result["topic"].(string), "orderBookApi"):
		var t wsOrderBook
		err = json.Unmarshal(respRaw, &t)
		if err != nil {
			return err
		}
		var newOB orderbook.Base
		var price, amount float64
		for i := range t.Data.SellQuote {
			p := strings.Replace(t.Data.SellQuote[i].Price, ",", "", -1)
			price, err = strconv.ParseFloat(p, 64)
			if err != nil {
				return err
			}
			a := strings.Replace(t.Data.SellQuote[i].Size, ",", "", -1)
			amount, err = strconv.ParseFloat(a, 64)
			if err != nil {
				return err
			}
			newOB.Asks = append(newOB.Asks, orderbook.Item{
				Price:  price,
				Amount: amount,
			})
		}
		for j := range t.Data.BuyQuote {
			p := strings.Replace(t.Data.BuyQuote[j].Price, ",", "", -1)
			price, err = strconv.ParseFloat(p, 64)
			if err != nil {
				return err
			}
			a := strings.Replace(t.Data.BuyQuote[j].Size, ",", "", -1)
			amount, err = strconv.ParseFloat(a, 64)
			if err != nil {
				return err
			}
			newOB.Bids = append(newOB.Bids, orderbook.Item{
				Price:  price,
				Amount: amount,
			})
		}
		p, err := currency.NewPairFromString(t.Topic[strings.Index(t.Topic, ":")+1 : strings.Index(t.Topic, currency.UnderscoreDelimiter)])
		if err != nil {
			return err
		}
		var a asset.Item
		a, err = b.GetPairAssetType(p)
		if err != nil {
			return err
		}
		newOB.Pair = p
		newOB.AssetType = a
		newOB.ExchangeName = b.Name
		err = b.Websocket.Orderbook.LoadSnapshot(&newOB)
		if err != nil {
			return err
		}
	default:
		b.Websocket.DataHandler <- stream.UnhandledMessageWarning{Message: b.Name + stream.UnhandledMessage + string(respRaw)}
		return nil
	}
	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (b *BTSE) GenerateDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	var channels = []string{"orderBookApi:%s_0", "tradeHistory:%s"}
	pairs, err := b.GetEnabledPairs(asset.Spot)
	if err != nil {
		return nil, err
	}
	var subscriptions []stream.ChannelSubscription
	if b.Websocket.CanUseAuthenticatedEndpoints() {
		subscriptions = append(subscriptions, stream.ChannelSubscription{
			Channel: "notificationApi",
		})
	}
	for i := range channels {
		for j := range pairs {
			subscriptions = append(subscriptions, stream.ChannelSubscription{
				Channel:  fmt.Sprintf(channels[i], pairs[j]),
				Currency: pairs[j],
				Asset:    asset.Spot,
			})
		}
	}
	return subscriptions, nil
}

// Subscribe sends a websocket message to receive data from the channel
func (b *BTSE) Subscribe(channelsToSubscribe []stream.ChannelSubscription) error {
	var sub wsSub
	sub.Operation = "subscribe"
	for i := range channelsToSubscribe {
		sub.Arguments = append(sub.Arguments, channelsToSubscribe[i].Channel)
	}
	err := b.Websocket.Conn.SendJSONMessage(sub)
	if err != nil {
		return err
	}
	b.Websocket.AddSuccessfulSubscriptions(channelsToSubscribe...)
	return nil
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (b *BTSE) Unsubscribe(channelsToUnsubscribe []stream.ChannelSubscription) error {
	var unSub wsSub
	unSub.Operation = "unsubscribe"
	for i := range channelsToUnsubscribe {
		unSub.Arguments = append(unSub.Arguments,
			channelsToUnsubscribe[i].Channel)
	}
	err := b.Websocket.Conn.SendJSONMessage(unSub)
	if err != nil {
		return err
	}
	b.Websocket.RemoveSuccessfulUnsubscriptions(channelsToUnsubscribe...)
	return nil
}
