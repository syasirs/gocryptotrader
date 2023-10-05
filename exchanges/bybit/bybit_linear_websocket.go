package bybit

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
)

// WsLinearConnect connects to linear a websocket feed
func (by *Bybit) WsLinearConnect() error {
	if !by.Websocket.IsEnabled() || !by.IsEnabled() || !by.IsAssetWebsocketSupported(asset.LinearContract) {
		return errWebsocketNotEnabled
	}
	by.Websocket.Conn.SetURL(linearPublic)
	var dialer websocket.Dialer
	err := by.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	by.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		MessageType: websocket.TextMessage,
		Message:     []byte(`{"op": "ping"}`),
		Delay:       bybitWebsocketTimer,
	})

	by.Websocket.Wg.Add(1)
	go by.wsReadData(asset.LinearContract, by.Websocket.Conn)
	if by.IsWebsocketAuthenticationSupported() {
		err = by.WsAuth(context.TODO())
		if err != nil {
			by.Websocket.DataHandler <- err
			by.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
	}
	return nil
}

// GenerateLinearDefaultSubscriptions generates default subscription
func (by *Bybit) GenerateLinearDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	var subscriptions []stream.ChannelSubscription
	var channels = []string{chanOrderbook, chanPublicTrade, chanPublicTicker}
	pairs, err := by.GetEnabledPairs(asset.LinearContract)
	if err != nil {
		return nil, err
	}
	for z := range pairs {
		for x := range channels {
			subscriptions = append(subscriptions,
				stream.ChannelSubscription{
					Channel:  channels[x],
					Currency: pairs[z],
					Asset:    asset.LinearContract,
				})
		}
	}
	return subscriptions, nil
}

// LinearSubscribe sends a subscription message to linear public channels.
func (by *Bybit) LinearSubscribe(channelSubscriptions []stream.ChannelSubscription) error {
	return by.handleLinearPayloadSubscription("subscribe", channelSubscriptions)
}

// LinearUnsubscribe sends an unsubscription messages through linear public channels.
func (by *Bybit) LinearUnsubscribe(channelSubscriptions []stream.ChannelSubscription) error {
	return by.handleLinearPayloadSubscription("unsubscribe", channelSubscriptions)
}

func (by *Bybit) handleLinearPayloadSubscription(operation string, channelSubscriptions []stream.ChannelSubscription) error {
	payloads, err := by.handleSubscriptions(asset.LinearContract, operation, channelSubscriptions)
	if err != nil {
		return err
	}
	for a := range payloads {
		// The options connection does not send the subscription request id back with the subscription notification payload
		// therefore the code doesn't wait for the response to check whether the subscription is successful or not.
		err = by.Websocket.Conn.SendJSONMessage(payloads[a])
		if err != nil {
			return err
		}
	}
	return nil
}
