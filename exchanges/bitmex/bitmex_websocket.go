package bitmex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/common/crypto"
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/asset"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	log "github.com/thrasher-/gocryptotrader/logger"
)

const (
	bitmexWSURL = "wss://www.bitmex.com/realtime"

	// Public Subscription Channels
	bitmexWSAnnouncement        = "announcement"
	bitmexWSChat                = "chat"
	bitmexWSConnected           = "connected"
	bitmexWSFunding             = "funding"
	bitmexWSInstrument          = "instrument"
	bitmexWSInsurance           = "insurance"
	bitmexWSLiquidation         = "liquidation"
	bitmexWSOrderbookL2         = "orderBookL2"
	bitmexWSOrderbookL10        = "orderBook10"
	bitmexWSPublicNotifications = "publicNotifications"
	bitmexWSQuote               = "quote"
	bitmexWSQuote1m             = "quoteBin1m"
	bitmexWSQuote5m             = "quoteBin5m"
	bitmexWSQuote1h             = "quoteBin1h"
	bitmexWSQuote1d             = "quoteBin1d"
	bitmexWSSettlement          = "settlement"
	bitmexWSTrade               = "trade"
	bitmexWSTrade1m             = "tradeBin1m"
	bitmexWSTrade5m             = "tradeBin5m"
	bitmexWSTrade1h             = "tradeBin1h"
	bitmexWSTrade1d             = "tradeBin1d"

	// Authenticated Subscription Channels
	bitmexWSAffiliate            = "affiliate"
	bitmexWSExecution            = "execution"
	bitmexWSOrder                = "order"
	bitmexWSMargin               = "margin"
	bitmexWSPosition             = "position"
	bitmexWSPrivateNotifications = "privateNotifications"
	bitmexWSTransact             = "transact"
	bitmexWSWallet               = "wallet"

	bitmexActionInitialData = "partial"
	bitmexActionInsertData  = "insert"
	bitmexActionDeleteData  = "delete"
	bitmexActionUpdateData  = "update"
)

var (
	pongChan = make(chan int, 1)
)

// WsConnector initiates a new websocket connection
func (b *Bitmex) WsConnector() error {
	if !b.Websocket.IsEnabled() || !b.IsEnabled() {
		return errors.New(exchange.WebsocketNotEnabled)
	}

	var dialer websocket.Dialer
	var err error

	if b.Websocket.GetProxyAddress() != "" {
		var proxy *url.URL
		proxy, err = url.Parse(b.Websocket.GetProxyAddress())
		if err != nil {
			return err
		}

		dialer.Proxy = http.ProxyURL(proxy)
	}

	b.WebsocketConn, _, err = dialer.Dial(b.Websocket.GetWebsocketURL(), nil)
	if err != nil {
		return err
	}

	_, p, err := b.WebsocketConn.ReadMessage()
	if err != nil {
		return err
	}

	var welcomeResp WebsocketWelcome
	err = common.JSONDecode(p, &welcomeResp)
	if err != nil {
		return err
	}

	if b.Verbose {
		log.Debugf(log.ExchangeSys, "Successfully connected to Bitmex %s at time: %s Limit: %d",
			welcomeResp.Info,
			welcomeResp.Timestamp,
			welcomeResp.Limit.Remaining)
	}

	go b.wsHandleIncomingData()
	b.GenerateDefaultSubscriptions()

	err = b.websocketSendAuth()
	if err != nil {
		log.Errorf(log.ExchangeSys, "%v - authentication failed: %v\n", b.Name, err)
	}
	b.GenerateAuthenticatedSubscriptions()
	return nil
}

func (b *Bitmex) wsReadData() (exchange.WebsocketResponse, error) {
	_, resp, err := b.WebsocketConn.ReadMessage()
	if err != nil {
		return exchange.WebsocketResponse{}, err
	}

	b.Websocket.TrafficAlert <- struct{}{}

	return exchange.WebsocketResponse{
		Raw: resp,
	}, nil
}

// wsHandleIncomingData services incoming data from the websocket connection
func (b *Bitmex) wsHandleIncomingData() {
	b.Websocket.Wg.Add(1)

	defer func() {
		b.Websocket.Wg.Done()
	}()

	for {
		select {
		case <-b.Websocket.ShutdownC:
			return

		default:
			resp, err := b.wsReadData()
			if err != nil {
				b.Websocket.DataHandler <- err
				return
			}

			message := string(resp.Raw)
			if strings.Contains(message, "pong") {
				pongChan <- 1
				continue
			}

			if strings.Contains(message, "ping") {
				err = b.wsSend("pong")
				if err != nil {
					b.Websocket.DataHandler <- err
					continue
				}
			}

			quickCapture := make(map[string]interface{})
			err = common.JSONDecode(resp.Raw, &quickCapture)
			if err != nil {
				b.Websocket.DataHandler <- err
				continue
			}

			var respError WebsocketErrorResponse
			if _, ok := quickCapture["status"]; ok {
				err = common.JSONDecode(resp.Raw, &respError)
				if err != nil {
					b.Websocket.DataHandler <- err
					continue
				}
				b.Websocket.DataHandler <- errors.New(respError.Error)
				continue
			}

			if _, ok := quickCapture["success"]; ok {
				var decodedResp WebsocketSubscribeResp
				err := common.JSONDecode(resp.Raw, &decodedResp)
				if err != nil {
					b.Websocket.DataHandler <- err
					continue
				}

				if decodedResp.Success {
					b.Websocket.DataHandler <- decodedResp
					if len(quickCapture) == 3 {
						if b.Verbose {
							log.Debugf(log.ExchangeSys, "%s websocket: Successfully subscribed to %s",
								b.Name, decodedResp.Subscribe)
						}
					} else {
						b.Websocket.SetCanUseAuthenticatedEndpoints(true)
						if b.Verbose {
							log.Debugf(log.ExchangeSys, "%s websocket: Successfully authenticated websocket connection",
								b.Name)
						}
					}
					continue
				}

				b.Websocket.DataHandler <- fmt.Errorf("%s websocket error: Unable to subscribe %s",
					b.Name, decodedResp.Subscribe)

			} else if _, ok := quickCapture["table"]; ok {
				var decodedResp WebsocketMainResponse
				err := common.JSONDecode(resp.Raw, &decodedResp)
				if err != nil {
					b.Websocket.DataHandler <- err
					continue
				}

				switch decodedResp.Table {
				case bitmexWSOrderbookL2:
					var orderbooks OrderBookData
					err = common.JSONDecode(resp.Raw, &orderbooks)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}

					p := currency.NewPairFromString(orderbooks.Data[0].Symbol)
					// TODO: update this to support multiple asset types
					err = b.processOrderbook(orderbooks.Data, orderbooks.Action, p, "CONTRACT")
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}

				case bitmexWSTrade:
					var trades TradeData
					err = common.JSONDecode(resp.Raw, &trades)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}

					if trades.Action == bitmexActionInitialData {
						continue
					}

					for _, trade := range trades.Data {
						var timestamp time.Time
						timestamp, err = time.Parse(time.RFC3339, trade.Timestamp)
						if err != nil {
							b.Websocket.DataHandler <- err
							continue
						}

						// TODO: update this to support multiple asset types
						b.Websocket.DataHandler <- exchange.TradeData{
							Timestamp:    timestamp,
							Price:        trade.Price,
							Amount:       float64(trade.Size),
							CurrencyPair: currency.NewPairFromString(trade.Symbol),
							Exchange:     b.GetName(),
							AssetType:    "CONTRACT",
							Side:         trade.Side,
						}
					}

				case bitmexWSAnnouncement:
					var announcement AnnouncementData
					err = common.JSONDecode(resp.Raw, &announcement)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}

					if announcement.Action == bitmexActionInitialData {
						continue
					}

					b.Websocket.DataHandler <- announcement.Data
				case bitmexWSAffiliate:
					var response WsAffiliateResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSExecution:
					var response WsExecutionResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSOrder:
					var response WsOrderResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSMargin:
					var response WsMarginResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSPosition:
					var response WsPositionResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSPrivateNotifications:
					var response WsPrivateNotificationsResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSTransact:
					var response WsTransactResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				case bitmexWSWallet:
					var response WsWalletResponse
					err = common.JSONDecode(resp.Raw, &response)
					if err != nil {
						b.Websocket.DataHandler <- err
						continue
					}
					b.Websocket.DataHandler <- response
				default:
					b.Websocket.DataHandler <- fmt.Errorf("%s websocket error: Table unknown - %s",
						b.Name, decodedResp.Table)
				}
			}
		}
	}
}

var snapshotloaded = make(map[currency.Pair]map[asset.Item]bool)

// ProcessOrderbook processes orderbook updates
func (b *Bitmex) processOrderbook(data []OrderBookL2, action string, currencyPair currency.Pair, assetType asset.Item) error { // nolint: unparam
	if len(data) < 1 {
		return errors.New("bitmex_websocket.go error - no orderbook data")
	}

	_, ok := snapshotloaded[currencyPair]
	if !ok {
		snapshotloaded[currencyPair] = make(map[asset.Item]bool)
	}

	_, ok = snapshotloaded[currencyPair][assetType]
	if !ok {
		snapshotloaded[currencyPair][assetType] = false
	}

	switch action {
	case bitmexActionInitialData:
		if !snapshotloaded[currencyPair][assetType] {
			var newOrderBook orderbook.Base
			var bids, asks []orderbook.Item

			for _, orderbookItem := range data {
				if orderbookItem.Side == exchange.SellOrderSide.ToString() {
					asks = append(asks, orderbook.Item{
						Price:  orderbookItem.Price,
						Amount: float64(orderbookItem.Size),
					})
					continue
				}
				bids = append(bids, orderbook.Item{
					Price:  orderbookItem.Price,
					Amount: float64(orderbookItem.Size),
				})
			}

			if len(bids) == 0 || len(asks) == 0 {
				return errors.New("bitmex_websocket.go error - snapshot not initialised correctly")
			}

			newOrderBook.Asks = asks
			newOrderBook.Bids = bids
			newOrderBook.AssetType = assetType
			newOrderBook.Pair = currencyPair

			err := b.Websocket.Orderbook.LoadSnapshot(&newOrderBook, b.GetName(), false)
			if err != nil {
				return fmt.Errorf("bitmex_websocket.go process orderbook error -  %s",
					err)
			}
			snapshotloaded[currencyPair][assetType] = true
			b.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
				Pair:     currencyPair,
				Asset:    assetType,
				Exchange: b.GetName(),
			}
		}

	default:
		if snapshotloaded[currencyPair][assetType] {
			var asks, bids []orderbook.Item
			for _, orderbookItem := range data {
				if orderbookItem.Side == "Sell" {
					asks = append(asks, orderbook.Item{
						Price:  orderbookItem.Price,
						Amount: float64(orderbookItem.Size),
					})
					continue
				}
				bids = append(bids, orderbook.Item{
					Price:  orderbookItem.Price,
					Amount: float64(orderbookItem.Size),
				})
			}

			err := b.Websocket.Orderbook.UpdateUsingID(bids,
				asks,
				currencyPair,
				b.GetName(),
				assetType,
				action)

			if err != nil {
				return err
			}

			b.Websocket.DataHandler <- exchange.WebsocketOrderbookUpdate{
				Pair:     currencyPair,
				Asset:    assetType,
				Exchange: b.GetName(),
			}
		}
	}
	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (b *Bitmex) GenerateDefaultSubscriptions() {
	contracts := b.GetEnabledPairs(asset.PerpetualContract)
	channels := []string{bitmexWSOrderbookL2, bitmexWSTrade}
	subscriptions := []exchange.WebsocketChannelSubscription{
		{
			Channel: bitmexWSAnnouncement,
		},
	}

	for i := range channels {
		for j := range contracts {
			subscriptions = append(subscriptions, exchange.WebsocketChannelSubscription{
				Channel:  fmt.Sprintf("%v:%v", channels[i], contracts[j].String()),
				Currency: contracts[j],
			})
		}
	}
	b.Websocket.SubscribeToChannels(subscriptions)
}

// GenerateAuthenticatedSubscriptions Adds authenticated subscriptions to websocket to be handled by ManageSubscriptions()
func (b *Bitmex) GenerateAuthenticatedSubscriptions() {
	if !b.Websocket.CanUseAuthenticatedEndpoints() {
		return
	}
	contracts := b.GetEnabledPairs(asset.PerpetualContract)
	channels := []string{bitmexWSExecution,
		bitmexWSPosition,
	}
	subscriptions := []exchange.WebsocketChannelSubscription{
		{
			Channel: bitmexWSAffiliate,
		},
		{
			Channel: bitmexWSOrder,
		},
		{
			Channel: bitmexWSMargin,
		},
		{
			Channel: bitmexWSPrivateNotifications,
		},
		{
			Channel: bitmexWSTransact,
		},
		{
			Channel: bitmexWSWallet,
		},
	}
	for i := range channels {
		for j := range contracts {
			subscriptions = append(subscriptions, exchange.WebsocketChannelSubscription{
				Channel:  fmt.Sprintf("%v:%v", channels[i], contracts[j].String()),
				Currency: contracts[j],
			})
		}
	}
	b.Websocket.SubscribeToChannels(subscriptions)
}

// Subscribe subscribes to a websocket channel
func (b *Bitmex) Subscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	var subscriber WebsocketRequest
	subscriber.Command = "subscribe"
	subscriber.Arguments = append(subscriber.Arguments, channelToSubscribe.Channel)
	return b.wsSend(subscriber)
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (b *Bitmex) Unsubscribe(channelToSubscribe exchange.WebsocketChannelSubscription) error {
	var subscriber WebsocketRequest
	subscriber.Command = "unsubscribe"
	subscriber.Arguments = append(subscriber.Arguments,
		channelToSubscribe.Params["args"],
		channelToSubscribe.Channel+":"+channelToSubscribe.Currency.String())
	return b.wsSend(subscriber)
}

// WebsocketSendAuth sends an authenticated subscription
func (b *Bitmex) websocketSendAuth() error {
	if !b.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
		return fmt.Errorf("%v AuthenticatedWebsocketAPISupport not enabled", b.Name)
	}
	b.Websocket.SetCanUseAuthenticatedEndpoints(true)
	timestamp := time.Now().Add(time.Hour * 1).Unix()
	newTimestamp := strconv.FormatInt(timestamp, 10)
	hmac := crypto.GetHMAC(crypto.HashSHA256,
		[]byte("GET/realtime"+newTimestamp),
		[]byte(b.API.Credentials.Secret))
	signature := crypto.HexEncodeToString(hmac)

	var sendAuth WebsocketRequest
	sendAuth.Command = "authKeyExpires"
	sendAuth.Arguments = append(sendAuth.Arguments, b.API.Credentials.Key, timestamp,
		signature)
	err := b.wsSend(sendAuth)
	if err != nil {
		b.Websocket.SetCanUseAuthenticatedEndpoints(false)
		return err
	}
	return nil
}

// WsSend sends data to the websocket server
func (b *Bitmex) wsSend(data interface{}) error {
	b.wsRequestMtx.Lock()
	defer b.wsRequestMtx.Unlock()
	if b.Verbose {
		log.Debugf(log.ExchangeSys, "%v sending message to websocket %v", b.Name, data)
	}
	return b.WebsocketConn.WriteJSON(data)
}
