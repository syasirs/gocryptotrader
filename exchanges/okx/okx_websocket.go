package okx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// OkxOrderbookMutex Ensures if two entries arrive at once, only one can be
// processed at a time
var OkxOrderbookMutex sync.Mutex

// responseStream a channel thought which the data coming from the two websocket connection will go through.
var responseStream = make(chan stream.Response)

// defaultSubscribedChannels list of chanels which are subscribed by default
var defaultSubscribedChannels = []string{
	okxChannelTrades,
	okxChannelOrderBooks,
	okxChannelOrderBooks5,
	okxChannelOrderBooks50TBT,
	okxChannelOrderBooksTBT,
	okxChannelCandle5m,
	okxChannelTickers,
}

const (
	// allowableIterations use the first 25 bids and asks in the full load to form a string
	allowableIterations = 25

	// OkxOrderBookSnapshot orderbook push data type 'snapshot'
	OkxOrderBookSnapshot = "snapshot"
	// OkxOrderBookUpdate orderbook push data type 'update'
	OkxOrderBookUpdate = "update"

	// ColonDelimiter to be used in validating checksum
	ColonDelimiter = ":"

	// maxConnByteLen otal length of multiple channels cannot exceed 4096 bytes.
	maxConnByteLen = 4096

	// Candlestick channels

	markPrice        = "mark-price-"
	indexCandlestick = "index-"
	candle           = "candle"

	// Candlesticks

	okxChannelCandle1Y     = candle + "1Y"
	okxChannelCandle6M     = candle + "6M"
	okxChannelCandle3M     = candle + "3M"
	okxChannelCandle1M     = candle + "1M"
	okxChannelCandle1W     = candle + "1W"
	okxChannelCandle1D     = candle + "1D"
	okxChannelCandle2D     = candle + "2D"
	okxChannelCandle3D     = candle + "3D"
	okxChannelCandle5D     = candle + "5D"
	okxChannelCandle12H    = candle + "12H"
	okxChannelCandle6H     = candle + "6H"
	okxChannelCandle4H     = candle + "4H"
	okxChannelCandle2H     = candle + "2H"
	okxChannelCandle1H     = candle + "1H"
	okxChannelCandle30m    = candle + "30m"
	okxChannelCandle15m    = candle + "15m"
	okxChannelCandle5m     = candle + "5m"
	okxChannelCandle3m     = candle + "3m"
	okxChannelCandle1m     = candle + "1m"
	okxChannelCandle1Yutc  = candle + "1Yutc"
	okxChannelCandle3Mutc  = candle + "3Mutc"
	okxChannelCandle1Mutc  = candle + "1Mutc"
	okxChannelCandle1Wutc  = candle + "1Wutc"
	okxChannelCandle1Dutc  = candle + "1Dutc"
	okxChannelCandle2Dutc  = candle + "2Dutc"
	okxChannelCandle3Dutc  = candle + "3Dutc"
	okxChannelCandle5Dutc  = candle + "5Dutc"
	okxChannelCandle12Hutc = candle + "12Hutc"
	okxChannelCandle6Hutc  = candle + "6Hutc"

	// Ticker channel
	okxChannelTickers                = "tickers"
	okxChannelIndexTickers           = "index-tickers"
	okxChannelStatus                 = "status"
	okxChannelPublicStrucBlockTrades = "public-struc-block-trades"
	okxChannelBlockTickers           = "block-tickers"

	// Private Channels
	okxChannelAccount            = "account"
	okxChannelPositions          = "positions"
	okxChannelBalanceAndPosition = "balance_and_position"
	okxChannelOrders             = "orders"
	okxChannelAlgoOrders         = "orders-algo"
	okxChannelAlgoAdvanced       = "algo-advance"
	okxChannelLiquidationWarning = "liquidation-warning"
	okxChannelAccountGreeks      = "account-greeks"
	okxChannelRFQs               = "rfqs"
	okxChannelQuotes             = "quotes"
	okxChannelStruckeBlockTrades = "struc-block-trades"
	okxChannelSpotGridOrder      = "grid-orders-spot"
	okxChannelGridOrdersConstuct = "grid-orders-contract"
	okxChannelGridPositions      = "grid-positions"
	okcChannelGridSubOrders      = "grid-sub-orders"
	okxChannelInstruments        = "instruments"
	okxChannelOpenInterest       = "open-interest"
	okxChannelTrades             = "trades"

	okxChannelEstimatedPrice  = "estimated-price"
	okxChannelMarkPrice       = "mark-price"
	okxChannelPriceLimit      = "price-limit"
	okxChannelOrderBooks      = "books"
	okxChannelOrderBooks5     = "books5"
	okxChannelOrderBooks50TBT = "books50-l2-tbt"
	okxChannelOrderBooksTBT   = "books-l2-tbt"
	okxChannelBBOTBT          = "bbo-tbt"
	okxChannelOptSummary      = "opt-summary"
	okxChannelFundingRate     = "funding-rate"

	// Index Candlesticks Channels
	okxChannelIndexCandle1Y     = indexCandlestick + okxChannelCandle1Y
	okxChannelIndexCandle6M     = indexCandlestick + okxChannelCandle6M
	okxChannelIndexCandle3M     = indexCandlestick + okxChannelCandle3M
	okxChannelIndexCandle1M     = indexCandlestick + okxChannelCandle1M
	okxChannelIndexCandle1W     = indexCandlestick + okxChannelCandle1W
	okxChannelIndexCandle1D     = indexCandlestick + okxChannelCandle1D
	okxChannelIndexCandle2D     = indexCandlestick + okxChannelCandle2D
	okxChannelIndexCandle3D     = indexCandlestick + okxChannelCandle3D
	okxChannelIndexCandle5D     = indexCandlestick + okxChannelCandle5D
	okxChannelIndexCandle12H    = indexCandlestick + okxChannelCandle12H
	okxChannelIndexCandle6H     = indexCandlestick + okxChannelCandle6H
	okxChannelIndexCandle4H     = indexCandlestick + okxChannelCandle4H
	okxChannelIndexCandle2H     = indexCandlestick + okxChannelCandle2H
	okxChannelIndexCandle1H     = indexCandlestick + okxChannelCandle1H
	okxChannelIndexCandle30m    = indexCandlestick + okxChannelCandle30m
	okxChannelIndexCandle15m    = indexCandlestick + okxChannelCandle15m
	okxChannelIndexCandle5m     = indexCandlestick + okxChannelCandle5m
	okxChannelIndexCandle3m     = indexCandlestick + okxChannelCandle3m
	okxChannelIndexCandle1m     = indexCandlestick + okxChannelCandle1m
	okxChannelIndexCandle1Yutc  = indexCandlestick + okxChannelCandle1Yutc
	okxChannelIndexCandle3Mutc  = indexCandlestick + okxChannelCandle3Mutc
	okxChannelIndexCandle1Mutc  = indexCandlestick + okxChannelCandle1Mutc
	okxChannelIndexCandle1Wutc  = indexCandlestick + okxChannelCandle1Wutc
	okxChannelIndexCandle1Dutc  = indexCandlestick + okxChannelCandle1Dutc
	okxChannelIndexCandle2Dutc  = indexCandlestick + okxChannelCandle2Dutc
	okxChannelIndexCandle3Dutc  = indexCandlestick + okxChannelCandle3Dutc
	okxChannelIndexCandle5Dutc  = indexCandlestick + okxChannelCandle5Dutc
	okxChannelIndexCandle12Hutc = indexCandlestick + okxChannelCandle12Hutc
	okxChannelIndexCandle6Hutc  = indexCandlestick + okxChannelCandle6Hutc

	// Mark price candlesticks channel
	okxChannelMarkPriceCandle1Y     = markPrice + okxChannelCandle1Y
	okxChannelMarkPriceCandle6M     = markPrice + okxChannelCandle6M
	okxChannelMarkPriceCandle3M     = markPrice + okxChannelCandle3M
	okxChannelMarkPriceCandle1M     = markPrice + okxChannelCandle1M
	okxChannelMarkPriceCandle1W     = markPrice + okxChannelCandle1W
	okxChannelMarkPriceCandle1D     = markPrice + okxChannelCandle1D
	okxChannelMarkPriceCandle2D     = markPrice + okxChannelCandle2D
	okxChannelMarkPriceCandle3D     = markPrice + okxChannelCandle3D
	okxChannelMarkPriceCandle5D     = markPrice + okxChannelCandle5D
	okxChannelMarkPriceCandle12H    = markPrice + okxChannelCandle12H
	okxChannelMarkPriceCandle6H     = markPrice + okxChannelCandle6H
	okxChannelMarkPriceCandle4H     = markPrice + okxChannelCandle4H
	okxChannelMarkPriceCandle2H     = markPrice + okxChannelCandle2H
	okxChannelMarkPriceCandle1H     = markPrice + okxChannelCandle1H
	okxChannelMarkPriceCandle30m    = markPrice + okxChannelCandle30m
	okxChannelMarkPriceCandle15m    = markPrice + okxChannelCandle15m
	okxChannelMarkPriceCandle5m     = markPrice + okxChannelCandle5m
	okxChannelMarkPriceCandle3m     = markPrice + okxChannelCandle3m
	okxChannelMarkPriceCandle1m     = markPrice + okxChannelCandle1m
	okxChannelMarkPriceCandle1Yutc  = markPrice + okxChannelCandle1Yutc
	okxChannelMarkPriceCandle3Mutc  = markPrice + okxChannelCandle3Mutc
	okxChannelMarkPriceCandle1Mutc  = markPrice + okxChannelCandle1Mutc
	okxChannelMarkPriceCandle1Wutc  = markPrice + okxChannelCandle1Wutc
	okxChannelMarkPriceCandle1Dutc  = markPrice + okxChannelCandle1Dutc
	okxChannelMarkPriceCandle2Dutc  = markPrice + okxChannelCandle2Dutc
	okxChannelMarkPriceCandle3Dutc  = markPrice + okxChannelCandle3Dutc
	okxChannelMarkPriceCandle5Dutc  = markPrice + okxChannelCandle5Dutc
	okxChannelMarkPriceCandle12Hutc = markPrice + okxChannelCandle12Hutc
	okxChannelMarkPriceCandle6Hutc  = markPrice + okxChannelCandle6Hutc
)

// WsConnect initiates a websocket connection
func (ok *Okx) WsConnect() error {
	if !ok.Websocket.IsEnabled() || !ok.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	dialer.ReadBufferSize = 8192
	dialer.WriteBufferSize = 8192

	er := ok.Websocket.Conn.Dial(&dialer, http.Header{})
	if er != nil {
		return er
	}
	if ok.Verbose {
		log.Debugf(log.ExchangeSys, "Successful connection to %v\n",
			ok.Websocket.GetWebsocketURL())
	}
	ok.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.PingMessage,
		Delay:             time.Second * 5,
	})

	if ok.IsWebsocketAuthenticationSupported() {
		var authDialer websocket.Dialer
		authDialer.ReadBufferSize = 8192
		authDialer.WriteBufferSize = 8192
		go func() {
			er = ok.WsAuth(context.Background(), &authDialer)
			if er != nil {
				ok.Websocket.SetCanUseAuthenticatedEndpoints(false)
				return
			}
			ok.Websocket.Wg.Add(1)
			go ok.wsFunnelConnectionData(ok.Websocket.AuthConn)
		}()
	}
	ok.Websocket.Wg.Add(2)
	go ok.wsFunnelConnectionData(ok.Websocket.Conn)
	go ok.WsReadData()
	return nil
}

// WsAuth will connect to Okx's Private websocket connection and Authenticate with a login payload.
func (ok *Okx) WsAuth(ctx context.Context, dialer *websocket.Dialer) error {
	if !ok.Websocket.CanUseAuthenticatedEndpoints() {
		return fmt.Errorf("%v AuthenticatedWebsocketAPISupport not enabled", ok.Name)
	}
	var creds *account.Credentials
	err := ok.Websocket.AuthConn.Dial(dialer, http.Header{})
	if err != nil {
		return fmt.Errorf("%v Websocket connection %v error. Error %v", ok.Name, okxAPIWebsocketPrivateURL, err)
	}
	ok.Websocket.AuthConn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.PingMessage,
		Delay:             time.Second * 5,
	})
	creds, err = ok.GetCredentials(ctx)
	if err != nil {
		return err
	}
	ok.Websocket.SetCanUseAuthenticatedEndpoints(true)
	timeUnix := time.Now()
	signPath := "/users/self/verify"
	hmac, err := crypto.GetHMAC(crypto.HashSHA256,
		[]byte(strconv.FormatInt(timeUnix.UTC().Unix(), 10)+http.MethodGet+signPath),
		[]byte(creds.Secret),
	)
	if err != nil {
		return err
	}
	base64Sign := crypto.Base64Encode(hmac)
	request := WebsocketEventRequest{
		Operation: "login",
		Arguments: []WebsocketLoginData{
			{
				APIKey:     creds.Key,
				Passphrase: creds.ClientID,
				Timestamp:  timeUnix,
				Sign:       base64Sign,
			},
		},
	}
	go func() {
		var response []byte
		response, err = ok.Websocket.AuthConn.SendMessageReturnResponse("login", request)
		var resp WSLoginResponse
		err = json.Unmarshal(response, &resp)
		if err == nil && (strings.EqualFold(resp.Event, "login") && resp.Code == 0) {
			ok.Websocket.SetCanUseAuthenticatedEndpoints(true)
		}
	}()
	return nil
}

// wsFunnelConnectionData receives data from multiple connection and pass the data
// to wsRead through a channel responseStream
func (ok *Okx) wsFunnelConnectionData(ws stream.Connection) {
	defer ok.Websocket.Wg.Done()
	for {
		resp := ws.ReadMessage()
		if resp.Raw == nil {
			return
		}
		responseStream <- stream.Response{Raw: resp.Raw}
	}
}

// Subscribe sends a websocket subscription request to several channels to receive data.
func (ok *Okx) Subscribe(channelsToSubscribe []stream.ChannelSubscription) error {
	return ok.handleSubscription("subscribe", channelsToSubscribe)
}

// Unsubscribe sends a websocket unsubscription request to several channels to receive data.
func (ok *Okx) Unsubscribe(channelsToUnsubscribe []stream.ChannelSubscription) error {
	return ok.handleSubscription("unsubscribe", channelsToUnsubscribe)
}

// handleSubscription sends a subscription and unsubscription information thought the websocket endpoint.
// as of the okex, exchange this endpoint sends subscription and unsubscription messages but with a list of json objects.
func (ok *Okx) handleSubscription(operation string, subscriptions []stream.ChannelSubscription) error {
	request := WSSubscriptionInformations{
		Operation: operation,
		Arguments: []SubscriptionInfo{},
	}

	authRequests := WSSubscriptionInformations{
		Operation: operation,
		Arguments: []SubscriptionInfo{},
	}

	var channels []stream.ChannelSubscription
	var authChannels []stream.ChannelSubscription
	var er error
	for i := 0; i < len(subscriptions); i++ {
		arg := SubscriptionInfo{
			Channel: subscriptions[i].Channel,
		}
		var instrumentID string
		var underlying string
		var okay bool
		var instrumentType string
		var authSubscription bool
		var algoID string
		var uid string

		if strings.EqualFold(arg.Channel, okxChannelAccount) ||
			strings.EqualFold(arg.Channel, okxChannelOrders) {
			authSubscription = true
		}
		if strings.EqualFold(arg.Channel, "grid-positions") {
			algoID, _ = subscriptions[i].Params["algoId"].(string)
		}

		if strings.EqualFold(arg.Channel, "grid-sub-orders") || strings.EqualFold(arg.Channel, "grid-positions") {
			uid, _ = subscriptions[i].Params["uid"].(string)
		}

		if strings.HasPrefix(arg.Channel, "candle") ||
			strings.EqualFold(arg.Channel, okxChannelTickers) ||
			strings.EqualFold(arg.Channel, okxChannelOrderBooks) ||
			strings.EqualFold(arg.Channel, okxChannelOrderBooks5) ||
			strings.EqualFold(arg.Channel, okxChannelOrderBooks50TBT) ||
			strings.EqualFold(arg.Channel, okxChannelOrderBooksTBT) ||
			strings.EqualFold(arg.Channel, okxChannelTrades) {
			if subscriptions[i].Params["instId"] != "" {
				instrumentID, okay = subscriptions[i].Params["instId"].(string)
				if !okay {
					instrumentID = ""
				}
			} else if subscriptions[i].Params["instrumentID"] != "" {
				instrumentID, okay = subscriptions[i].Params["instrumentID"].(string)
				if !okay {
					instrumentID = ""
				}
			}
			if instrumentID == "" {
				instrumentID, er = ok.getInstrumentIDFromPair(subscriptions[i].Currency, subscriptions[i].Asset)
				if er != nil {
					instrumentID = ""
				}
			}
		}
		if strings.EqualFold(arg.Channel, "instruments") ||
			strings.EqualFold(arg.Channel, "positions") ||
			strings.EqualFold(arg.Channel, "orders") ||
			strings.EqualFold(arg.Channel, "orders-algo") ||
			strings.EqualFold(arg.Channel, "algo-advance") ||
			strings.EqualFold(arg.Channel, "liquidation-warning") ||
			strings.EqualFold(arg.Channel, "grid-orders-spot") ||
			strings.EqualFold(arg.Channel, "grid-orders-spot") ||
			strings.EqualFold(arg.Channel, "grid-orders-contract") ||
			strings.EqualFold(arg.Channel, "estimated-price") {
			instrumentType = ok.GetInstrumentTypeFromAssetItem(subscriptions[i].Asset)
		}

		if strings.EqualFold(arg.Channel, "positions") ||
			strings.EqualFold(arg.Channel, "orders") ||
			strings.EqualFold(arg.Channel, "orders-algo") ||
			strings.EqualFold(arg.Channel, "estimated-price") ||
			strings.EqualFold(arg.Channel, "opt-summary") {
			underlying, _ = ok.GetUnderlying(subscriptions[i].Currency, subscriptions[i].Asset)
		}
		arg.InstrumentID = instrumentID
		arg.Underlying = underlying
		arg.InstrumentType = instrumentType
		arg.UID = uid
		arg.AlgoID = algoID

		if authSubscription {
			var authChunk []byte
			authChannels = append(authChannels, subscriptions[i])
			authRequests.Arguments = append(authRequests.Arguments, arg)
			authChunk, er = json.Marshal(authRequests)
			if er != nil {
				return er
			}
			if len(authChunk) > maxConnByteLen {
				authRequests.Arguments = authRequests.Arguments[:len(authRequests.Arguments)-1]
				i--
				er = ok.Websocket.AuthConn.SendJSONMessage(authRequests)
				if er != nil {
					return er
				}
				if operation == "unsubscribe" {
					ok.Websocket.RemoveSuccessfulUnsubscriptions(channels...)
				} else {
					ok.Websocket.AddSuccessfulSubscriptions(channels...)
				}
				authChannels = []stream.ChannelSubscription{}
				authRequests.Arguments = []SubscriptionInfo{}
			}
		} else {
			var chunk []byte
			channels = append(channels, subscriptions[i])
			request.Arguments = append(request.Arguments, arg)
			chunk, er = json.Marshal(request)
			if er != nil {
				return er
			}
			if len(chunk) > maxConnByteLen {
				i--
				er = ok.Websocket.Conn.SendJSONMessage(request)
				if er != nil {
					return er
				}
				if operation == "unsubscribe" {
					ok.Websocket.RemoveSuccessfulUnsubscriptions(channels...)
				} else {
					ok.Websocket.AddSuccessfulSubscriptions(channels...)
				}
				channels = []stream.ChannelSubscription{}
				request.Arguments = []SubscriptionInfo{}
				continue
			}
		}
	}
	if len(request.Arguments) > 0 {
		er = ok.Websocket.Conn.SendJSONMessage(request)
		if er != nil {
			return er
		}
	}
	if len(authRequests.Arguments) > 0 && ok.Websocket.CanUseAuthenticatedEndpoints() {
		er = ok.Websocket.AuthConn.SendJSONMessage(authRequests)
		if er != nil {
			return er
		}
	}
	if er != nil {
		return er
	}

	if operation == "unsubscribe" {
		channels = append(channels, authChannels...)
		ok.Websocket.RemoveSuccessfulUnsubscriptions(channels...)
	} else {
		channels = append(channels, authChannels...)
		ok.Websocket.AddSuccessfulSubscriptions(channels...)
	}
	return nil
}

// WsReadData read coming messages thought the websocket connection and process the data.
func (ok *Okx) WsReadData() {
	defer ok.Websocket.Wg.Done()
	for {
		select {
		case <-ok.Websocket.ShutdownC:
			select {
			case resp := <-responseStream:
				err := ok.WsHandleData(resp.Raw)
				if err != nil {
					select {
					case ok.Websocket.DataHandler <- err:
					default:
						log.Errorf(log.WebsocketMgr, "%s websocket handle data error: %v", ok.Name, err)
					}
				}
			default:
			}
			return
		case resp := <-responseStream:
			err := ok.WsHandleData(resp.Raw)
			if err != nil {
				ok.Websocket.DataHandler <- err
			}
		}
	}
}

// WsHandleData will read websocket raw data and pass to appropriate handler
func (ok *Okx) WsHandleData(respRaw []byte) error {
	var dataResponse WebsocketDataResponse
	er := json.Unmarshal(respRaw, &dataResponse)
	if er != nil {
		var resp WSLoginResponse
		er = json.Unmarshal(respRaw, &resp)
		if er == nil && (strings.EqualFold(resp.Event, "login") && resp.Code == 0) {
			ok.Websocket.SetCanUseAuthenticatedEndpoints(true)
		} else if er == nil && (strings.EqualFold(resp.Event, "error") || resp.Code == 60006 || resp.Code == 60007 || resp.Code == 60009 || resp.Code == 60026) {
			ok.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
		return er
	}
	if len(dataResponse.Data) > 0 {
		switch strings.ToLower(dataResponse.Argument.Channel) {
		case okxChannelCandle1Y, okxChannelCandle6M, okxChannelCandle3M, okxChannelCandle1M, okxChannelCandle1W,
			okxChannelCandle1D, okxChannelCandle2D, okxChannelCandle3D, okxChannelCandle5D, okxChannelCandle12H,
			okxChannelCandle6H, okxChannelCandle4H, okxChannelCandle2H, okxChannelCandle1H, okxChannelCandle30m,
			okxChannelCandle15m, okxChannelCandle5m, okxChannelCandle3m, okxChannelCandle1m, okxChannelCandle1Yutc,
			okxChannelCandle3Mutc, okxChannelCandle1Mutc, okxChannelCandle1Wutc, okxChannelCandle1Dutc,
			okxChannelCandle2Dutc, okxChannelCandle3Dutc, okxChannelCandle5Dutc, okxChannelCandle12Hutc,
			okxChannelCandle6Hutc:
			return ok.wsProcessCandles(&dataResponse)
		case okxChannelIndexCandle1Y, okxChannelIndexCandle6M, okxChannelIndexCandle3M, okxChannelIndexCandle1M,
			okxChannelIndexCandle1W, okxChannelIndexCandle1D, okxChannelIndexCandle2D, okxChannelIndexCandle3D,
			okxChannelIndexCandle5D, okxChannelIndexCandle12H, okxChannelIndexCandle6H, okxChannelIndexCandle4H,
			okxChannelIndexCandle2H, okxChannelIndexCandle1H, okxChannelIndexCandle30m, okxChannelIndexCandle15m,
			okxChannelIndexCandle5m, okxChannelIndexCandle3m, okxChannelIndexCandle1m, okxChannelIndexCandle1Yutc,
			okxChannelIndexCandle3Mutc, okxChannelIndexCandle1Mutc, okxChannelIndexCandle1Wutc,
			okxChannelIndexCandle1Dutc, okxChannelIndexCandle2Dutc, okxChannelIndexCandle3Dutc, okxChannelIndexCandle5Dutc,
			okxChannelIndexCandle12Hutc, okxChannelIndexCandle6Hutc:
			return ok.wsProcessIndexCandles(&dataResponse)
		case okxChannelTickers:
			return ok.wsProcessTickers(respRaw)
		case okxChannelIndexTickers:
			var response WsIndexTicker
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelStatus:
			var response WsSystemStatusResponse
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelPublicStrucBlockTrades:
			var response WsPublicTradesResponse
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelBlockTickers:
			var response WsBlockTicker
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelAccountGreeks:
			var response WsGreeks
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelAccount:
			var response WsAccountChannelPushData
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelPositions,
			okxChannelLiquidationWarning:
			var response WsPositionResponse
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelBalanceAndPosition:
			var response WsBalanceAndPosition
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelOrders:
			return ok.wsProcessOrders(respRaw)
		case okxChannelAlgoOrders:
			var response WsAlgoOrder
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelAlgoAdvanced:
			var response WsAdvancedAlgoOrder
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelRFQs:
			var response WsRFQ
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelQuotes:
			var response WsQuote
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelStruckeBlockTrades:
			var response WsStructureBlocTrade
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelSpotGridOrder:
			var response WsSpotGridAlgoOrder
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelGridOrdersConstuct:
			var response WsContractGridAlgoOrder
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelGridPositions:
			var response WsContractGridAlgoOrder
			return ok.wsProcessPushData(respRaw, &response)
		case okcChannelGridSubOrders:
			var response WsGridSubOrderData
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelInstruments:
			var response WSInstrumentResponse
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelOpenInterest:
			var response WSOpenInterestResponse
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelTrades:
			return ok.wsProcessTrades(respRaw)
		case okxChannelEstimatedPrice:
			var response WsDeliveryEstimatedPrice
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelMarkPrice,
			okxChannelPriceLimit:
			var response WsMarkPrice
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelOrderBooks,
			okxChannelOrderBooks5,
			okxChannelOrderBooks50TBT,
			okxChannelBBOTBT,
			okxChannelOrderBooksTBT:
			return ok.wsProcessOrderBooks(respRaw)
		case okxChannelOptSummary:
			var response WsOptionSummary
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelFundingRate:
			var response WsFundingRate
			return ok.wsProcessPushData(respRaw, &response)
		case okxChannelMarkPriceCandle1Y, okxChannelMarkPriceCandle6M, okxChannelMarkPriceCandle3M, okxChannelMarkPriceCandle1M,
			okxChannelMarkPriceCandle1W, okxChannelMarkPriceCandle1D, okxChannelMarkPriceCandle2D, okxChannelMarkPriceCandle3D,
			okxChannelMarkPriceCandle5D, okxChannelMarkPriceCandle12H, okxChannelMarkPriceCandle6H, okxChannelMarkPriceCandle4H,
			okxChannelMarkPriceCandle2H, okxChannelMarkPriceCandle1H, okxChannelMarkPriceCandle30m, okxChannelMarkPriceCandle15m,
			okxChannelMarkPriceCandle5m, okxChannelMarkPriceCandle3m, okxChannelMarkPriceCandle1m, okxChannelMarkPriceCandle1Yutc,
			okxChannelMarkPriceCandle3Mutc, okxChannelMarkPriceCandle1Mutc, okxChannelMarkPriceCandle1Wutc, okxChannelMarkPriceCandle1Dutc,
			okxChannelMarkPriceCandle2Dutc, okxChannelMarkPriceCandle3Dutc, okxChannelMarkPriceCandle5Dutc, okxChannelMarkPriceCandle12Hutc,
			okxChannelMarkPriceCandle6Hutc:
			return ok.wsHandleMarkPriceCandles(respRaw)
		default:
			ok.Websocket.DataHandler <- stream.UnhandledMessageWarning{Message: ok.Name + stream.UnhandledMessage + string(respRaw)}
			return nil
		}
	}
	return nil
}

// wsProcessIndexCandles processes index candlestic data
func (ok *Okx) wsProcessIndexCandles(intermediate *WebsocketDataResponse) error {
	if intermediate == nil {
		return errNilArgument
	}
	var response WSCandlestickResponse
	if len(intermediate.Data) == 0 {
		return errNoCandlestickDataFound
	}
	pair, er := ok.GetPairFromInstrumentID(intermediate.Argument.InstrumentID)
	if er != nil {
		return er
	}
	var a asset.Item
	a, _ = ok.GetAssetTypeFromInstrumentType(intermediate.Argument.InstrumentType)
	candleInterval := strings.TrimPrefix(intermediate.Argument.Channel, candle)
	for i := range response.Data {
		candles, okay := (intermediate.Data[i]).([5]string)
		if !okay {
			return errIncompleteCandlestickData
		}
		timestamp, er := strconv.Atoi(candles[0])
		if er != nil {
			return er
		}
		candle := stream.KlineData{
			Pair:      pair,
			Exchange:  ok.Name,
			Timestamp: time.UnixMilli(int64(timestamp)),
			Interval:  candleInterval,
			AssetType: a,
		}
		candle.OpenPrice, er = strconv.ParseFloat(candles[1], 64)
		if er != nil {
			return er
		}
		candle.HighPrice, er = strconv.ParseFloat(candles[2], 64)
		if er != nil {
			return er
		}
		candle.LowPrice, er = strconv.ParseFloat(candles[3], 64)
		if er != nil {
			return er
		}
		candle.ClosePrice, er = strconv.ParseFloat(candles[4], 64)
		if er != nil {
			return er
		}
		ok.Websocket.DataHandler <- candle
	}
	return nil
}

// wsProcessOrderBooks processes "snapshot" and "update" order book
func (ok *Okx) wsProcessOrderBooks(data []byte) error {
	var response WsOrderBook
	var er error
	er = json.Unmarshal(data, &response)
	if er != nil {
		return er
	}
	if !(strings.EqualFold(response.Action, OkxOrderBookUpdate) ||
		strings.EqualFold(response.Action, OkxOrderBookSnapshot) ||
		strings.EqualFold(response.Argument.Channel, okxChannelOrderBooks5) ||
		strings.EqualFold(response.Argument.Channel, okxChannelBBOTBT) ||
		strings.EqualFold(response.Argument.Channel, okxChannelOrderBooks50TBT) ||
		strings.EqualFold(response.Argument.Channel, okxChannelOrderBooksTBT)) {
		return errors.New("invalid order book action ")
	}
	OkxOrderbookMutex.Lock()
	defer OkxOrderbookMutex.Unlock()
	var pair currency.Pair
	var a asset.Item
	a, _ = ok.GetAssetTypeFromInstrumentType(response.Argument.InstrumentType)
	if a == asset.Empty {
		a = ok.GuessAssetTypeFromInstrumentID(response.Argument.InstrumentID)
	}
	pair, er = ok.GetPairFromInstrumentID(response.Argument.InstrumentID)
	if er != nil {
		pair.Delimiter = currency.DashDelimiter
	}
	for i := range response.Data {
		if strings.EqualFold(response.Action, OkxOrderBookSnapshot) ||
			strings.EqualFold(response.Argument.Channel, okxChannelOrderBooks5) ||
			strings.EqualFold(response.Argument.Channel, okxChannelBBOTBT) {
			er = ok.WsProcessFullOrderBook(response.Data[i], pair, a)
			if er != nil {
				_, err2 := ok.OrderBooksSubscription("subscribe", response.Argument.Channel, a, pair)
				if err2 != nil {
					ok.Websocket.DataHandler <- err2
				}
				return er
			}
		} else {
			if len(response.Data[i].Asks) == 0 && len(response.Data[i].Bids) == 0 {
				return nil
			}
			er := ok.WsProcessUpdateOrderbook(response.Argument.Channel, response.Data[i], pair, a)
			if er != nil {
				_, err2 := ok.OrderBooksSubscription("subscribe", response.Argument.Channel, a, pair)
				if err2 != nil {
					ok.Websocket.DataHandler <- err2
				}
				return er
			}
		}
	}
	return nil
}

// WsProcessFullOrderBook processes snapshot order books
func (ok *Okx) WsProcessFullOrderBook(data WsOrderBookData, pair currency.Pair, a asset.Item) error {
	var er error
	if data.Checksum != 0 {
		var signedChecksum int32
		signedChecksum, er = ok.CalculateOrderbookChecksum(data)
		if er != nil {
			return fmt.Errorf("%s channel: Orderbook unable to calculate orderbook checksum: %s", ok.Name, er)
		}
		if signedChecksum != data.Checksum {
			return fmt.Errorf("%s channel: Orderbook for %v checksum invalid",
				ok.Name,
				pair)
		}
	}

	if ok.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s passed checksum for pair %v",
			ok.Name, pair,
		)
	}
	var asks []orderbook.Item
	var bids []orderbook.Item
	asks, er = ok.AppendWsOrderbookItems(data.Asks)
	if er != nil {
		return er
	}
	bids, er = ok.AppendWsOrderbookItems(data.Bids)
	if er != nil {
		return er
	}
	newOrderBook := orderbook.Base{
		Asset:           a,
		Asks:            asks,
		Bids:            bids,
		LastUpdated:     data.Timestamp,
		Pair:            pair,
		Exchange:        ok.Name,
		VerifyOrderbook: ok.CanVerifyOrderbook,
	}
	return ok.Websocket.Orderbook.LoadSnapshot(&newOrderBook)
}

// WsProcessUpdateOrderbook updates an existing orderbook using websocket data
// After merging WS data, it will sort, validate and finally update the existing
// orderbook
func (ok *Okx) WsProcessUpdateOrderbook(channel string, data WsOrderBookData, pair currency.Pair, a asset.Item) error {
	update := &orderbook.Update{
		Asset: a,
		Pair:  pair,
	}
	switch channel {
	case okxChannelOrderBooks,
		okxChannelOrderBooksTBT:
		update.MaxDepth = 400
	case okxChannelOrderBooks5:
		update.MaxDepth = 5
	case okxChannelBBOTBT:
		update.MaxDepth = 1
		update.MaxDepth = 400
	case okxChannelOrderBooks50TBT:
		update.MaxDepth = 50
	}
	var err error
	update.Asks, err = ok.AppendWsOrderbookItems(data.Asks)
	if err != nil {
		return err
	}
	update.Bids, err = ok.AppendWsOrderbookItems(data.Bids)
	if err != nil {
		return err
	}
	err = ok.Websocket.Orderbook.Update(update)
	if err != nil {
		return err
	}

	if data.Checksum != 0 {
		var checksum int32
		checksum, err = ok.CalculateOrderbookChecksum(data)
		if err != nil {
			return err
		}
		if checksum != data.Checksum {
			log.Warnf(log.ExchangeSys, "%s checksum failure for pair %v",
				ok.Name,
				pair)
			return errors.New("checksum failed")
		}
	}
	return nil
}

// AppendWsOrderbookItems adds websocket orderbook data bid/asks into an
// orderbook item array
func (ok *Okx) AppendWsOrderbookItems(entries [][4]string) ([]orderbook.Item, error) {
	items := make([]orderbook.Item, len(entries))
	for j := range entries {
		amount, err := strconv.ParseFloat(entries[j][1], 64)
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(entries[j][0], 64)
		if err != nil {
			return nil, err
		}
		items[j] = orderbook.Item{Amount: amount, Price: price}
	}
	return items, nil
}

// CalculateOrderbookChecksum alternates over the first 25 bid and ask entries from websocket data.
func (ok *Okx) CalculateOrderbookChecksum(orderbookData WsOrderBookData) (int32, error) {
	var checksum strings.Builder
	for i := 0; i < allowableIterations; i++ {
		if len(orderbookData.Bids)-1 >= i {
			bidPrice := orderbookData.Bids[i][0]
			bidAmount := orderbookData.Bids[i][1]
			checksum.WriteString(
				bidPrice +
					ColonDelimiter +
					bidAmount +
					ColonDelimiter)
		}
		if len(orderbookData.Asks)-1 >= i {
			askPrice := orderbookData.Asks[i][0]
			askAmount := orderbookData.Asks[i][1]
			checksum.WriteString(askPrice +
				ColonDelimiter +
				askAmount +
				ColonDelimiter)
		}
	}
	checksumStr := strings.TrimSuffix(checksum.String(), ColonDelimiter)
	return int32(crc32.ChecksumIEEE([]byte(checksumStr))), nil
}

// CalculateUpdateOrderbookChecksum alternates over the first 25 bid and ask
// entries of a merged orderbook. The checksum is made up of the price and the
// quantity with a semicolon (:) deliminating them. This will also work when
// there are less than 25 entries (for whatever reason)
// eg Bid:Ask:Bid:Ask:Ask:Ask
func (ok *Okx) CalculateUpdateOrderbookChecksum(orderbookData *orderbook.Base) int32 {
	var checksum strings.Builder
	for i := 0; i < allowableIterations; i++ {
		if len(orderbookData.Bids)-1 >= i {
			price := strconv.FormatFloat(orderbookData.Bids[i].Price, 'f', 0, 64)
			amount := strconv.FormatFloat(orderbookData.Bids[i].Amount, 'f', 0, 64)
			checksum.WriteString(price + ColonDelimiter + amount + ColonDelimiter)
		}
		if len(orderbookData.Asks)-1 >= i {
			price := strconv.FormatFloat(orderbookData.Asks[i].Price, 'f', 0, 64)
			amount := strconv.FormatFloat(orderbookData.Asks[i].Amount, 'f', 0, 64)
			checksum.WriteString(price + ColonDelimiter + amount + ColonDelimiter)
		}
	}
	checksumStr := strings.TrimSuffix(checksum.String(), ColonDelimiter)
	return int32(crc32.ChecksumIEEE([]byte(checksumStr)))
}

// wsHandleMarkPriceCandles processes candlestick mark price push data as a result of  subscription to "mark-price-candle*" channel.
func (ok *Okx) wsHandleMarkPriceCandles(data []byte) error {
	tempo := &struct {
		Argument SubscriptionInfo `json:"arg"`
		Data     [][5]string      `json:"data"`
	}{}
	var er error
	er = json.Unmarshal(data, tempo)
	if er != nil {
		return er
	}
	var tsInt int64
	var ts time.Time
	var op float64
	var hp float64
	var lp float64
	var cp float64
	candles := []CandlestickMarkPrice{}
	for x := range tempo.Data {
		tsInt, er = strconv.ParseInt(tempo.Data[x][0], 10, 64)
		if er != nil {
			return er
		}
		ts = time.UnixMilli(tsInt)
		op, er = strconv.ParseFloat(tempo.Data[x][1], 64)
		if er != nil {
			return er
		}
		hp, er = strconv.ParseFloat(tempo.Data[x][2], 64)
		if er != nil {
			return er
		}
		lp, er = strconv.ParseFloat(tempo.Data[x][3], 64)
		if er != nil {
			return er
		}
		cp, er = strconv.ParseFloat(tempo.Data[x][4], 64)
		if er != nil {
			return er
		}
		candles = append(candles, CandlestickMarkPrice{
			Timestamp:    ts,
			OpenPrice:    op,
			HighestPrice: hp,
			LowestPrice:  lp,
			ClosePrice:   cp,
		})
	}
	ok.Websocket.DataHandler <- candles
	return nil
}

// wsProcessTrades handles a list of trade information.
func (ok *Okx) wsProcessTrades(data []byte) error {
	var response WsTradeOrder
	var er error
	var assetType asset.Item
	er = json.Unmarshal(data, &response)
	if er != nil {
		return er
	}
	assetType, _ = ok.GetAssetTypeFromInstrumentType(response.Argument.InstrumentType)
	if assetType == asset.Empty {
		assetType = ok.GuessAssetTypeFromInstrumentID(response.Argument.InstrumentID)
	}
	trades := make([]trade.Data, len(response.Data))
	for i := range response.Data {
		var pair currency.Pair
		pair, er = ok.GetPairFromInstrumentID(response.Data[i].InstrumentID)
		if er != nil {
			ok.Websocket.DataHandler <- order.ClassificationError{
				Exchange: ok.Name,
				Err:      er,
			}
			return er
		}
		amount := response.Data[i].Quantity
		side := order.ParseOrderSideString(response.Data[i].Side)
		trades[i] = trade.Data{
			Amount:       amount,
			AssetType:    assetType,
			CurrencyPair: pair,
			Exchange:     ok.Name,
			Side:         side,
			Timestamp:    response.Data[i].Timestamp,
			TID:          response.Data[i].TradeID,
			Price:        response.Data[i].Price,
		}
	}
	return trade.AddTradesToBuffer(ok.Name, trades...)
}

// wsProcessOrders handles websocket order push data responses.
func (ok *Okx) wsProcessOrders(respRaw []byte) error {
	var response WsOrderResponse
	var pair currency.Pair
	var assetType asset.Item
	var er error
	er = json.Unmarshal(respRaw, &response)
	if er != nil {
		return er
	}
	assetType, er = ok.GetAssetTypeFromInstrumentType(response.Argument.InstrumentType)
	if er != nil {
		return er
	}
	for x := range response.Data {
		var orderType order.Type
		var orderStatus order.Status
		side := response.Data[x].Side
		orderType, er = order.StringToOrderType(response.Data[x].OrderType)
		if er != nil {
			ok.Websocket.DataHandler <- order.ClassificationError{
				Exchange: ok.Name,
				OrderID:  response.Data[x].OrderID,
				Err:      er,
			}
		}
		orderStatus, er = order.StringToOrderStatus(response.Data[x].State)
		if er != nil {
			ok.Websocket.DataHandler <- order.ClassificationError{
				Exchange: ok.Name,
				OrderID:  response.Data[x].OrderID,
				Err:      er,
			}
		}
		var a asset.Item
		a, er = ok.GetAssetTypeFromInstrumentType(response.Argument.InstrumentType)
		if er != nil {
			ok.Websocket.DataHandler <- order.ClassificationError{
				Exchange: ok.Name,
				OrderID:  response.Data[x].OrderID,
				Err:      er,
			}
			a = assetType
		}
		pair, er = ok.GetPairFromInstrumentID(response.Data[x].InstrumentID)
		if er != nil {
			return er
		}
		ok.Websocket.DataHandler <- &order.Detail{
			Price:           response.Data[x].Price,
			Amount:          response.Data[x].Size,
			ExecutedAmount:  response.Data[x].LastFilledSize,
			RemainingAmount: response.Data[x].AccumulatedFillSize - response.Data[x].LastFilledSize,
			Exchange:        ok.Name,
			OrderID:         response.Data[x].OrderID,
			Type:            orderType,
			Side:            side,
			Status:          orderStatus,
			AssetType:       a,
			Date:            response.Data[x].CreationTime,
			Pair:            pair,
		}
	}
	return nil
}

// wsProcessCandles handler to get a list of candlestick messages.
func (ok *Okx) wsProcessCandles(intermediate *WebsocketDataResponse) error {
	if intermediate == nil {
		return errNilArgument
	}
	var response WSCandlestickResponse
	if len(intermediate.Data) == 0 {
		return errNoCandlestickDataFound
	}
	pair, er := ok.GetPairFromInstrumentID(intermediate.Argument.InstrumentID)
	if er != nil {
		return er
	}
	var a asset.Item
	a, er = ok.GetAssetTypeFromInstrumentType(intermediate.Argument.InstrumentType)
	if er != nil {
		a = ok.GuessAssetTypeFromInstrumentID(intermediate.Argument.InstrumentID)
	}
	candleInterval := strings.TrimPrefix(intermediate.Argument.Channel, candle)
	for i := range response.Data {
		candles, okay := (intermediate.Data[i]).([7]string)
		if !okay {
			return errIncompleteCandlestickData
		}
		timestamp, er := strconv.Atoi(candles[0])
		if er != nil {
			return er
		}
		candle := stream.KlineData{
			Pair:      pair,
			Exchange:  ok.Name,
			Timestamp: time.UnixMilli(int64(timestamp)),
			Interval:  candleInterval,
			AssetType: a,
		}
		candle.OpenPrice, er = strconv.ParseFloat(candles[1], 64)
		if er != nil {
			return er
		}
		candle.HighPrice, er = strconv.ParseFloat(candles[2], 64)
		if er != nil {
			return er
		}
		candle.LowPrice, er = strconv.ParseFloat(candles[3], 64)
		if er != nil {
			return er
		}
		candle.ClosePrice, er = strconv.ParseFloat(candles[4], 64)
		if er != nil {
			return er
		}
		candle.Volume, er = strconv.ParseFloat(candles[5], 64)
		if er != nil {
			return er
		}
		ok.Websocket.DataHandler <- candle
	}
	return nil
}

// wsProcessTickers handles the trade ticker information.
func (ok *Okx) wsProcessTickers(data []byte) error {
	var response WSTickerResponse
	if er := json.Unmarshal(data, &response); er != nil {
		return er
	}
	for i := range response.Data {
		a := response.Data[i].InstrumentType
		if a == asset.Empty {
			a = ok.GuessAssetTypeFromInstrumentID(response.Data[i].InstrumentID)
		}
		if !(ok.SupportsAsset(a)) {
			return errInvalidInstrumentType
		}
		var c currency.Pair
		var er error
		c, er = ok.GetPairFromInstrumentID(response.Data[i].InstrumentID)
		if er != nil {
			return er
		}
		var baseVolume float64
		var quoteVolume float64
		if a == asset.Spot || a == asset.Margin {
			baseVolume = response.Data[i].Vol24H
			quoteVolume = response.Data[i].VolCcy24H
		} else {
			baseVolume = response.Data[i].VolCcy24H
			quoteVolume = response.Data[i].Vol24H
		}
		ok.Websocket.DataHandler <- &ticker.Price{
			ExchangeName: ok.Name,
			Open:         response.Data[i].Open24H,
			Volume:       baseVolume,
			QuoteVolume:  quoteVolume,
			High:         response.Data[i].High24H,
			Low:          response.Data[i].Low24H,
			Bid:          response.Data[i].BidPrice,
			Ask:          response.Data[i].BestAskPrice,
			BidSize:      response.Data[i].BidSize,
			AskSize:      response.Data[i].BestAskSize,
			Last:         response.Data[i].LastTradePrice,
			AssetType:    a,
			Pair:         c,
			LastUpdated:  response.Data[i].TickerDataGenerationTime,
		}
	}
	return nil
}

// GenerateDefaultSubscriptions returns a list of default subscription message.
func (ok *Okx) GenerateDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	var subscriptions []stream.ChannelSubscription
	assets := ok.GetAssetTypes(true)
	if ok.Websocket.CanUseAuthenticatedEndpoints() {
		defaultSubscribedChannels = append(defaultSubscribedChannels,
			okxChannelAccount,
			okxChannelOrders,
		)
	}
	for x := range assets {
		pairs, err := ok.GetEnabledPairs(assets[x])
		if err != nil {
			return nil, err
		}
		for y := range defaultSubscribedChannels {
			if defaultSubscribedChannels[y] == okxChannelCandle5m ||
				defaultSubscribedChannels[y] == okxChannelTickers ||
				defaultSubscribedChannels[y] == okxChannelOrders ||
				defaultSubscribedChannels[y] == okxChannelOrderBooks ||
				defaultSubscribedChannels[y] == okxChannelOrderBooks5 ||
				defaultSubscribedChannels[y] == okxChannelOrderBooks50TBT ||
				defaultSubscribedChannels[y] == okxChannelOrderBooksTBT ||
				defaultSubscribedChannels[y] == okxChannelTrades {
				for p := range pairs {
					subscriptions = append(subscriptions, stream.ChannelSubscription{
						Channel:  defaultSubscribedChannels[y],
						Asset:    assets[x],
						Currency: pairs[p],
					})
				}
			} else {
				subscriptions = append(subscriptions, stream.ChannelSubscription{
					Channel: defaultSubscribedChannels[y],
				})
			}
		}
	}
	return subscriptions, nil
}

// wsProcessPushData processes push data coming through the websocket channel
func (ok *Okx) wsProcessPushData(data []byte, resp interface{}) error {
	if er := json.Unmarshal(data, resp); er != nil {
		return er
	}
	ok.Websocket.DataHandler <- resp
	return nil
}

// Websocket Trade methods

// WSPlaceOrder places an order thought the websocket connection stream, and returns a SubmitResponse and error message.
func (ok *Okx) WSPlaceOrder(arg *PlaceOrderRequestParam) (*PlaceOrderResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(strings.EqualFold(TradeModeCross, arg.TradeMode) || strings.EqualFold(TradeModeIsolated, arg.TradeMode) || strings.EqualFold(TradeModeCash, arg.TradeMode)) {
		return nil, errInvalidTradeModeValue
	}
	if !(strings.EqualFold(arg.Side, "buy") || strings.EqualFold(arg.Side, "sell")) {
		return nil, errMissingOrderSide
	}
	if !(strings.EqualFold(arg.OrderType, "market") || strings.EqualFold(arg.OrderType, "limit") || strings.EqualFold(arg.OrderType, "post_only") ||
		strings.EqualFold(arg.OrderType, "fok") || strings.EqualFold(arg.OrderType, "ioc") || strings.EqualFold(arg.OrderType, "optimal_limit_ioc")) {
		return nil, errInvalidOrderType
	}
	if arg.QuantityToBuyOrSell <= 0 {
		return nil, errInvalidQuantityToButOrSell
	}
	if arg.OrderPrice <= 0 && (strings.EqualFold(arg.OrderType, "limit") || strings.EqualFold(arg.OrderType, "post_only") ||
		strings.EqualFold(arg.OrderType, "fok") || strings.EqualFold(arg.OrderType, "ioc")) {
		return nil, fmt.Errorf("invalid order price for %s order types", arg.OrderType)
	}
	if !(strings.EqualFold(arg.QuantityType, "base_ccy") || strings.EqualFold(arg.QuantityType, "quote_ccy")) {
		arg.QuantityType = ""
	}
	randomID := common.GenerateRandomString(32, common.SmallLetters, common.CapitalLetters, common.NumberCharacters)
	input := WsPlaceOrderInput{
		ID:        randomID,
		Arguments: []PlaceOrderRequestParam{*arg},
		Operation: "batch-orders",
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("order", input)
	if er != nil {
		return nil, er
	}
	var placeOrderResponse WSPlaceOrderResponse
	er = json.Unmarshal(respData, &placeOrderResponse)
	if er != nil {
		return nil, er
	}
	if !(placeOrderResponse.Code == "0" ||
		placeOrderResponse.Code == "2") {
		if placeOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(placeOrderResponse.Msg)
	} else if len(placeOrderResponse.Data) == 0 {
		if placeOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(placeOrderResponse.Msg)
	}
	return &(placeOrderResponse.Data[0]), nil
}

// WsPlaceMultipleOrder creates an order through the websocket stream.
func (ok *Okx) WsPlaceMultipleOrder(args []PlaceOrderRequestParam) ([]PlaceOrderResponse, error) {
	for x := range args {
		arg := args[x]
		if arg.InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if !(strings.EqualFold(TradeModeCross, arg.TradeMode) || strings.EqualFold(TradeModeIsolated, arg.TradeMode) || strings.EqualFold(TradeModeCash, arg.TradeMode)) {
			return nil, errInvalidTradeModeValue
		}
		if !(strings.EqualFold(arg.Side, "buy") || strings.EqualFold(arg.Side, "sell")) {
			return nil, errMissingOrderSide
		}
		if !(strings.EqualFold(arg.OrderType, "market") || strings.EqualFold(arg.OrderType, "limit") || strings.EqualFold(arg.OrderType, "post_only") ||
			strings.EqualFold(arg.OrderType, "fok") || strings.EqualFold(arg.OrderType, "ioc") || strings.EqualFold(arg.OrderType, "optimal_limit_ioc")) {
			return nil, errInvalidOrderType
		}
		if arg.QuantityToBuyOrSell <= 0 {
			return nil, errInvalidQuantityToButOrSell
		}
		if arg.OrderPrice <= 0 && (strings.EqualFold(arg.OrderType, "limit") || strings.EqualFold(arg.OrderType, "post_only") ||
			strings.EqualFold(arg.OrderType, "fok") || strings.EqualFold(arg.OrderType, "ioc")) {
			return nil, fmt.Errorf("invalid order price for %s order types", arg.OrderType)
		}
		if !(strings.EqualFold(arg.QuantityType, "base_ccy") || strings.EqualFold(arg.QuantityType, "quote_ccy")) {
			arg.QuantityType = ""
		}
	}
	randomID := common.GenerateRandomString(4, common.NumberCharacters)
	input := WsPlaceOrderInput{
		ID:        randomID,
		Arguments: args,
		Operation: "batch-orders",
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("orders", input)
	if er != nil {
		return nil, er
	}
	var placeOrderResponse WSPlaceOrderResponse
	er = json.Unmarshal(respData, &placeOrderResponse)
	if er != nil {
		return nil, er
	}
	if !(placeOrderResponse.Code == "0" ||
		placeOrderResponse.Code == "2") {
		if placeOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(placeOrderResponse.Msg)
	}
	return placeOrderResponse.Data, nil
}

// WsCancelOrder websocket function to cancel a trade order
func (ok *Okx) WsCancelOrder(arg CancelOrderRequestParam) (*PlaceOrderResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
		return nil, fmt.Errorf("either order id or client supplier id is required")
	}
	randomID := common.GenerateRandomString(4, common.NumberCharacters)
	input := WsCancelOrderInput{
		ID:        randomID,
		Arguments: []CancelOrderRequestParam{arg},
		Operation: "cancel-order",
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("cancel-orders", input)
	if er != nil {
		return nil, er
	}
	var cancelOrderResponse WSPlaceOrderResponse
	er = json.Unmarshal(respData, &cancelOrderResponse)
	if er != nil {
		return nil, er
	}
	if cancelOrderResponse.Code != "1" || strings.EqualFold(cancelOrderResponse.Code, "60013") {
		if cancelOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(cancelOrderResponse.Msg)
	} else if len(cancelOrderResponse.Data) == 0 {
		if cancelOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(cancelOrderResponse.Msg)
	}
	return &(cancelOrderResponse.Data[0]), nil
}

// WsCancleMultipleOrder cancel multiple order through the websocket channel.
func (ok *Okx) WsCancleMultipleOrder(args []CancelOrderRequestParam) ([]PlaceOrderResponse, error) {
	for x := range args {
		arg := args[x]
		if arg.InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
			return nil, fmt.Errorf("either order id or client supplier id is required")
		}
	}
	randomID := common.GenerateRandomString(4, common.NumberCharacters)
	input := WsCancelOrderInput{
		ID:        randomID,
		Arguments: args,
		Operation: "batch-cancel-orders",
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("cancel-orders", input)
	if er != nil {
		return nil, er
	}
	var cancelOrderResponse WSPlaceOrderResponse
	er = json.Unmarshal(respData, &cancelOrderResponse)
	if er != nil {
		return nil, er
	}
	if cancelOrderResponse.Code != "1" ||
		strings.EqualFold(cancelOrderResponse.Code, "60013") {
		if cancelOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(cancelOrderResponse.Msg)
	}
	return cancelOrderResponse.Data, nil
}

// WsAmendOrder method to amend trade order using a request thought the websocket channel.
func (ok *Okx) WsAmendOrder(arg *AmendOrderRequestParams) (*AmendOrderResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if arg.ClientSuppliedOrderID == "" && arg.OrderID == "" {
		return nil, errMissingClientOrderIDOrOrderID
	}
	if arg.NewQuantity <= 0 && arg.NewPrice <= 0 {
		return nil, errMissingNewSizeOrPriceInformation
	}
	randomID := common.GenerateRandomString(4, common.NumberCharacters)
	input := WsAmendOrderInput{
		ID:        randomID,
		Operation: "amend-order",
		Arguments: []AmendOrderRequestParams{*arg},
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("amend-order", input)
	if er != nil {
		return nil, er
	}
	var amendOrderResponse WsAmendOrderResponse
	er = json.Unmarshal(respData, &amendOrderResponse)
	if er != nil {
		return nil, er
	}

	if !strings.EqualFold(amendOrderResponse.Code, "0") ||
		strings.EqualFold(amendOrderResponse.Code, "1") ||
		strings.EqualFold(amendOrderResponse.Code, "60013") {
		if amendOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(amendOrderResponse.Msg)
	} else if len(amendOrderResponse.Data) == 0 {
		if amendOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(amendOrderResponse.Msg)
	}
	return &amendOrderResponse.Data[0], nil
}

// WsAmendMultipleOrders a request through the websocket connection to amend multiple trade orders.
func (ok *Okx) WsAmendMultipleOrders(args []AmendOrderRequestParams) ([]AmendOrderResponse, error) {
	for x := range args {
		if args[x].InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if args[x].ClientSuppliedOrderID == "" && args[x].OrderID == "" {
			return nil, errMissingClientOrderIDOrOrderID
		}
		if args[x].NewQuantity <= 0 && args[x].NewPrice <= 0 {
			return nil, errMissingNewSizeOrPriceInformation
		}
	}
	randomID := common.GenerateRandomString(4, common.NumberCharacters)
	input := &WsAmendOrderInput{
		ID:        randomID,
		Operation: "batch-amend-orders",
		Arguments: args,
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("amend-orders", input)
	if er != nil {
		return nil, er
	}
	var amendOrderResponse WsAmendOrderResponse
	er = json.Unmarshal(respData, &amendOrderResponse)
	if er != nil {
		return nil, er
	}
	if !strings.EqualFold(amendOrderResponse.Code, "0") ||
		!strings.EqualFold(amendOrderResponse.Code, "2") ||
		strings.EqualFold(amendOrderResponse.Code, "1") ||
		strings.EqualFold(amendOrderResponse.Code, "60013") {
		if amendOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(amendOrderResponse.Msg)
	} else if len(amendOrderResponse.Data) == 0 {
		if amendOrderResponse.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(amendOrderResponse.Msg)
	}
	return amendOrderResponse.Data, nil
}

// WsChannelSubscription send a subscription or unsubscription request for different channels through the websocket stream.
func (ok *Okx) WsChannelSubscription(operation, channel string, assetType asset.Item, pair currency.Pair, tooglers ...bool) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(operation, "subscribe") || strings.EqualFold(operation, "unsubscribe")) {
		return nil, errInvalidWebsocketEvent
	}
	var underlying string
	var instrumentID string
	var instrumentType string
	var er error
	if len(tooglers) > 0 && tooglers[0] {
		instrumentType = strings.ToUpper(assetType.String())
		if !(strings.EqualFold(instrumentType, okxInstTypeSpot) ||
			strings.EqualFold(instrumentType, okxInstTypeMargin) ||
			strings.EqualFold(instrumentType, okxInstTypeSwap) ||
			strings.EqualFold(instrumentType, okxInstTypeFutures) ||
			strings.EqualFold(instrumentType, okxInstTypeOption)) {
			instrumentType = okxInstTypeANY
		}
	}
	if len(tooglers) > 2 && tooglers[2] {
		if !pair.IsEmpty() {
			underlying, _ = ok.GetUnderlying(pair, assetType)
		}
	}
	if len(tooglers) > 1 && tooglers[1] {
		instrumentID, er = ok.getInstrumentIDFromPair(pair, assetType)
		if er != nil {
			instrumentID = ""
		}
	}
	if channel == "" {
		return nil, errMissingValidChannelInformation
	}
	input := &SubscriptionOperationInput{
		Operation: operation,
		Arguments: []SubscriptionInfo{
			{
				Channel:        channel,
				InstrumentType: instrumentType,
				Underlying:     underlying,
				InstrumentID:   instrumentID,
			},
		},
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse(channel, input)
	if er != nil {
		return nil, er
	}
	var resp SubscriptionOperationResponse
	er = json.Unmarshal(respData, &resp)
	if er != nil {
		return nil, er
	}
	if strings.EqualFold(resp.Event, "error") || strings.EqualFold(resp.Code, "60012") {
		if resp.Msg == "" {
			return nil, fmt.Errorf("%s %s error %s", channel, operation, string(respData))
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp, nil
}

// Private Channel Websocket methods

// WsAuthChannelSubscription send a subscription or unsubscription request for different channels through the websocket stream.
func (ok *Okx) WsAuthChannelSubscription(operation, channel string, assetType asset.Item, pair currency.Pair, uid, algoID string, tooglers ...bool) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(operation, "subscribe") || strings.EqualFold(operation, "unsubscribe")) {
		return nil, errInvalidWebsocketEvent
	}
	var underlying string
	var instrumentID string
	var instrumentType string
	var ccy string
	var er error
	if len(tooglers) > 0 && tooglers[0] {
		instrumentType = strings.ToUpper(assetType.String())
		if !(strings.EqualFold(instrumentType, okxInstTypeSpot) ||
			strings.EqualFold(instrumentType, okxInstTypeMargin) ||
			strings.EqualFold(instrumentType, okxInstTypeSwap) ||
			strings.EqualFold(instrumentType, okxInstTypeFutures) ||
			strings.EqualFold(instrumentType, okxInstTypeOption)) {
			instrumentType = okxInstTypeANY
		}
	}
	if len(tooglers) > 2 && tooglers[2] {
		if !pair.IsEmpty() {
			underlying, _ = ok.GetUnderlying(pair, assetType)
		}
	}
	if len(tooglers) > 1 && tooglers[1] {
		instrumentID, er = ok.getInstrumentIDFromPair(pair, assetType)
		if er != nil {
			instrumentID = ""
		}
	}
	if len(tooglers) > 3 && tooglers[3] {
		if !(pair.IsEmpty()) {
			if !(pair.Base.IsEmpty()) {
				ccy = strings.ToUpper(pair.Base.String())
			} else {
				ccy = strings.ToUpper(pair.Quote.String())
			}
		}
	}
	if channel == "" {
		return nil, errMissingValidChannelInformation
	}
	input := &SubscriptionOperationInput{
		Operation: operation,
		Arguments: []SubscriptionInfo{
			{
				Channel:        channel,
				InstrumentType: instrumentType,
				Underlying:     underlying,
				InstrumentID:   instrumentID,
				Currency:       ccy,
				UID:            uid,
			},
		},
	}
	respData, er := ok.Websocket.AuthConn.SendMessageReturnResponse(channel, input)
	if er != nil {
		return nil, er
	}
	var resp SubscriptionOperationResponse
	er = json.Unmarshal(respData, &resp)
	if er != nil {
		return nil, er
	}
	if strings.EqualFold(resp.Event, "error") || strings.EqualFold(resp.Code, "60012") {
		if resp.Msg == "" {
			return nil, fmt.Errorf("%s %s error %s", channel, operation, string(respData))
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp, nil
}

// WsAccountSubscription retrieve account information. Data will be pushed when triggered by
// events such as placing order, canceling order, transaction execution, etc.
// It will also be pushed in regular interval according to subscription granularity.
func (ok *Okx) WsAccountSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "account", assetType, pair, "", "", false, false, false, true)
}

// WsPositionChannel retrieve the position data. The first snapshot will be sent in accordance with the granularity of the subscription. Data will be pushed when certain actions, such placing or canceling an order, trigger it. It will also be pushed periodically based on the granularity of the subscription.
func (ok *Okx) WsPositionChannel(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "positions", assetType, pair, "", "", true)
}

// BalanceAndPositionSubscription retrieve account balance and position information. Data will be pushed when triggered by events such as filled order, funding transfer.
func (ok *Okx) BalanceAndPositionSubscription(operation, uid string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "balance_and_position", asset.Empty, currency.EMPTYPAIR, "", "")
}

// WsOrderChannel for subscribing for orders.
func (ok *Okx) WsOrderChannel(operation string, assetType asset.Item, pair currency.Pair, instrumentID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "orders", assetType, pair, "", "", true, true, true)
}

// AlgoOrdersSubscription for subscribing to algo - order channels
func (ok *Okx) AlgoOrdersSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "orders-algo", assetType, pair, "", "", true, true, true)
}

// AdvanceAlgoOrdersSubscription algo order subscription to retrieve advance algo orders (including Iceberg order, TWAP order, Trailing order). Data will be pushed when first subscribed. Data will be pushed when triggered by events such as placing/canceling order.
func (ok *Okx) AdvanceAlgoOrdersSubscription(operation string, assetType asset.Item, pair currency.Pair, algoID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "algo-advance", assetType, pair, "", algoID, true, true)
}

// PositionRiskWarningSubscription this push channel is only used as a risk warning, and is not recommended as a risk judgment for strategic trading
// In the case that the market is not moving violently, there may be the possibility that the position has been liquidated at the same time that this message is pushed.
func (ok *Okx) PositionRiskWarningSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "liquidation-warning", assetType, pair, "", "", true, true, true)
}

// AccountGreeksSubscription algo order subscription to retrieve account greeks information. Data will be pushed when triggered by events such as increase/decrease positions or cash balance in account, and will also be pushed in regular interval according to subscription granularity.
func (ok *Okx) AccountGreeksSubscription(operation string, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "account-greeks", asset.Empty, pair, "", "", false, false, false, true)
}

// RfqSubscription subscription to retrieve Rfq updates on RFQ orders.
func (ok *Okx) RfqSubscription(operation, uid string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "rfqs", asset.Empty, currency.EMPTYPAIR, uid, "")
}

// QuotesSubscription subscription to retrieve Quote subscription
func (ok *Okx) QuotesSubscription(operation string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "quotes", asset.Empty, currency.EMPTYPAIR, "", "")
}

// StructureBlockTradesSubscription to retrieve Structural block subscription
func (ok *Okx) StructureBlockTradesSubscription(operation string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "struc-block-trades", asset.Empty, currency.EMPTYPAIR, "", "")
}

// SpotGridAlgoOrdersSubscription to retrieve spot grid algo orders. Data will be pushed when first subscribed. Data will be pushed when triggered by events such as placing/canceling order.
func (ok *Okx) SpotGridAlgoOrdersSubscription(operation string, assetType asset.Item, pair currency.Pair, algoID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "grid-orders-spot", assetType, pair, "", algoID, true, true, true)
}

// ContractGridAlgoOrders to retrieve contract grid algo orders. Data will be pushed when first subscribed. Data will be pushed when triggered by events such as placing/canceling order.
func (ok *Okx) ContractGridAlgoOrders(operation string, assetType asset.Item, pair currency.Pair, algoID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "grid-orders-contract", assetType, pair, "", algoID, true, true, true)
}

// GridPositionsSubscription to retrieve grid positions. Data will be pushed when first subscribed. Data will be pushed when triggered by events such as placing/canceling order.
func (ok *Okx) GridPositionsSubscription(operation, algoID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "grid-positions", asset.Empty, currency.EMPTYPAIR, "", algoID)
}

// GridSubOrders to retrieve grid sub orders. Data will be pushed when first subscribed. Data will be pushed when triggered by events such as placing order.
func (ok *Okx) GridSubOrders(operation, algoID string) (*SubscriptionOperationResponse, error) {
	return ok.WsAuthChannelSubscription(operation, "grid-sub-orders", asset.Empty, currency.EMPTYPAIR, "", algoID)
}

// Public Websocket stream subscription

// InstrumentsSubscription to subscribe for instruments. The full instrument list will be pushed
// for the first time after subscription. Subsequently, the instruments will be pushed if there is any change to the instrument’s state (such as delivery of FUTURES,
// exercise of OPTION, listing of new contracts / trading pairs, trading suspension, etc.).
func (ok *Okx) InstrumentsSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "instruments", assetType, pair, true)
}

// TickersSubscription subscribing to "ticker" channel to retrieve the last traded price, bid price, ask price and 24-hour trading volume of instruments. Data will be pushed every 100 ms.
func (ok *Okx) TickersSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "tickers", assetType, pair, false, true)
}

// OpenInterestSubscription to subscribe or unsubscribe to "open-interest" channel to retrieve the open interest. Data will by pushed every 3 seconds.
func (ok *Okx) OpenInterestSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "open-interest", assetType, pair, false, true)
}

// CandlesticksSubscription to subscribe or unsubscribe to "candle" channels to retrieve the candlesticks data of an instrument. the push frequency is the fastest interval 500ms push the data.
func (ok *Okx) CandlesticksSubscription(operation, channel string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(channel, "candle1Y") || strings.EqualFold(channel, "candle6M") || strings.EqualFold(channel, "candle3M") || strings.EqualFold(channel, "candle1M") || strings.EqualFold(channel, "candle1W") || strings.EqualFold(channel, "candle1D") || strings.EqualFold(channel, "candle2D") || strings.EqualFold(channel, "candle3D") || strings.EqualFold(channel, "candle5D") || strings.EqualFold(channel, "candle12H") || strings.EqualFold(channel, "candle6H") || strings.EqualFold(channel, "candle4H") || strings.EqualFold(channel, "candle2H") || strings.EqualFold(channel, "candle1H") || strings.EqualFold(channel, "candle30m") || strings.EqualFold(channel, "candle15m") || strings.EqualFold(channel, "candle5m") || strings.EqualFold(channel, "candle3m") || strings.EqualFold(channel, "candle1m") || strings.EqualFold(channel, "candle1Yutc") || strings.EqualFold(channel, "candle3Mutc") || strings.EqualFold(channel, "candle1Mutc") || strings.EqualFold(channel, "candle1Wutc") || strings.EqualFold(channel, "candle1Dutc") || strings.EqualFold(channel, "candle2Dutc") || strings.EqualFold(channel, "candle3Dutc") || strings.EqualFold(channel, "candle5Dutc") || strings.EqualFold(channel, "candle12Hutc") || strings.EqualFold(channel, "candle6Hutc")) {
		return nil, errMissingValidChannelInformation
	}
	return ok.WsChannelSubscription(operation, channel, assetType, pair, false, true)
}

// TradesSubscription to subscribe or unsubscribe to "trades" channel to retrieve the recent trades data. Data will be pushed whenever there is a trade. Every update contain only one trade.
func (ok *Okx) TradesSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "trades", assetType, pair, false, true)
}

// EstimatedDeliveryExercisePriceSubscription to subscribe or unsubscribe to "estimated-price" channel to retrieve the estimated delivery/exercise price of FUTURES contracts and OPTION.
func (ok *Okx) EstimatedDeliveryExercisePriceSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "estimated-price", assetType, pair, true, false, true)
}

// MarkPriceSubscription to subscribe or unsubscribe to to "mark-price" to retrieve the mark price. Data will be pushed every 200 ms when the mark price changes, and will be pushed every 10 seconds when the mark price does not change.
func (ok *Okx) MarkPriceSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "mark-price", assetType, pair, false, true)
}

// MarkPriceCandlesticksSubscription to subscribe or unsubscribe to "mark-price-candles" channels to retrieve the candlesticks data of the mark price. Data will be pushed every 500 ms.
func (ok *Okx) MarkPriceCandlesticksSubscription(operation, channel string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(channel, "mark-price-candle1Y") || strings.EqualFold(channel, "mark-price-candle6M") || strings.EqualFold(channel, "mark-price-candle3M") || strings.EqualFold(channel, "mark-price-candle1M") || strings.EqualFold(channel, "mark-price-candle1W") || strings.EqualFold(channel, "mark-price-candle1D") || strings.EqualFold(channel, "mark-price-candle2D") || strings.EqualFold(channel, "mark-price-candle3D") || strings.EqualFold(channel, "mark-price-candle5D") || strings.EqualFold(channel, "mark-price-candle12H") || strings.EqualFold(channel, "mark-price-candle6H") || strings.EqualFold(channel, "mark-price-candle4H") || strings.EqualFold(channel, "mark-price-candle2H") || strings.EqualFold(channel, "mark-price-candle1H") || strings.EqualFold(channel, "mark-price-candle30m") || strings.EqualFold(channel, "mark-price-candle15m") || strings.EqualFold(channel, "mark-price-candle5m") || strings.EqualFold(channel, "mark-price-candle3m") || strings.EqualFold(channel, "mark-price-candle1m") || strings.EqualFold(channel, "mark-price-candle1Yutc") || strings.EqualFold(channel, "mark-price-candle3Mutc") || strings.EqualFold(channel, "mark-price-candle1Mutc") || strings.EqualFold(channel, "mark-price-candle1Wutc") || strings.EqualFold(channel, "mark-price-candle1Dutc") || strings.EqualFold(channel, "mark-price-candle2Dutc") || strings.EqualFold(channel, "mark-price-candle3Dutc") || strings.EqualFold(channel, "mark-price-candle5Dutc") || strings.EqualFold(channel, "mark-price-candle12Hutc") || strings.EqualFold(channel, "mark-price-candle6Hutc")) {
		return nil, errMissingValidChannelInformation
	}
	return ok.WsChannelSubscription(operation, channel, assetType, pair, false, true)
}

// PriceLimitSubscription subscribe or unsubscribe to "price-limit" channel to retrieve the maximum buy price and minimum sell price of the instrument. Data will be pushed every 5 seconds when there are changes in limits, and will not be pushed when there is no changes on limit.
func (ok *Okx) PriceLimitSubscription(operation, instrumentID string) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(operation, "subscribe") || strings.EqualFold(operation, "unsubscribe")) {
		return nil, errInvalidWebsocketEvent
	}
	var er error
	input := &SubscriptionOperationInput{
		Operation: operation,
		Arguments: []SubscriptionInfo{
			{
				Channel:      "price-limit",
				InstrumentID: instrumentID,
			},
		},
	}
	respData, er := ok.Websocket.Conn.SendMessageReturnResponse("price-limit", input)
	if er != nil {
		return nil, er
	}
	var resp SubscriptionOperationResponse
	er = json.Unmarshal(respData, &resp)
	if er != nil {
		return nil, er
	}
	if strings.EqualFold(resp.Event, "error") || strings.EqualFold(resp.Code, "60012") {
		if resp.Msg == "" {
			return nil, fmt.Errorf("%s %s error %s", "price-limit", operation, string(respData))
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp, nil
}

// OrderBooksSubscription subscribe or unsubscribe to "books*" channel to retrieve order book data.
func (ok *Okx) OrderBooksSubscription(operation, channel string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(channel, "books") || strings.EqualFold(channel, "books5") || strings.EqualFold(channel, "books50-l2-tbt") || strings.EqualFold(channel, "books-l2-tbt")) {
		return nil, errMissingValidChannelInformation
	}
	return ok.WsChannelSubscription(operation, channel, assetType, pair, false, true)
}

// OptionSummarySubscription a method to subscribe or unsubscribe to "opt-summary" channel
// to retrieve detailed pricing information of all OPTION contracts. Data will be pushed at once.
func (ok *Okx) OptionSummarySubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "opt-summary", assetType, pair, false, false, true)
}

// FundingRateSubscription a methos to subscribe and unsubscribe to "funding-rate" channel.
// retrieve funding rate. Data will be pushed in 30s to 90s.
func (ok *Okx) FundingRateSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "funding-rate", assetType, pair, false, true)
}

// IndexCandlesticksSubscription a method to subscribe and unsubscribe to "index-candle*" channel
// to retrieve the candlesticks data of the index. Data will be pushed every 500 ms.
func (ok *Okx) IndexCandlesticksSubscription(operation, channel string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	if !(strings.EqualFold(channel, "index-candle1Y") || strings.EqualFold(channel, "index-candle6M") || strings.EqualFold(channel, "index-candle3M") || strings.EqualFold(channel, "index-candle1M") || strings.EqualFold(channel, "index-candle1W") || strings.EqualFold(channel, "index-candle1D") || strings.EqualFold(channel, "index-candle2D") || strings.EqualFold(channel, "index-candle3D") || strings.EqualFold(channel, "index-candle5D") || strings.EqualFold(channel, "index-candle12H") ||
		strings.EqualFold(channel, "index-candle6H") || strings.EqualFold(channel, "index-candle4H") || strings.EqualFold(channel, "index -candle2H") || strings.EqualFold(channel, "index-candle1H") || strings.EqualFold(channel, "index-candle30m") || strings.EqualFold(channel, "index-candle15m") || strings.EqualFold(channel, "index-candle5m") || strings.EqualFold(channel, "index-candle3m") || strings.EqualFold(channel, "index-candle1m") || strings.EqualFold(channel, "index-candle1Yutc") || strings.EqualFold(channel, "index-candle3Mutc") || strings.EqualFold(channel, "index-candle1Mutc") || strings.EqualFold(channel, "index-candle1Wutc") || strings.EqualFold(channel, "index-candle1Dutc") || strings.EqualFold(channel, "index-candle2Dutc") || strings.EqualFold(channel, "index-candle3Dutc") || strings.EqualFold(channel, "index-candle5Dutc") || strings.EqualFold(channel, "index-candle12Hutc") || strings.EqualFold(channel, "index-candle6Hutc")) {
		return nil, errMissingValidChannelInformation
	}
	return ok.WsChannelSubscription(operation, channel, assetType, pair, false, true)
}

// IndexTickerChannel a method to subscribe and unsubscribe to "index-tickers" channel
func (ok *Okx) IndexTickerChannel(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "index-tickers", assetType, pair, false, true)
}

// StatusSubscription get the status of system maintenance and push when the system maintenance status changes.
// First subscription: "Push the latest change data"; every time there is a state change, push the changed content
func (ok *Okx) StatusSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "status", assetType, pair)
}

// PublicStructureBlockTradesSubscription a method to subscribe or unsubscribe to "public-struc-block-trades" channel
func (ok *Okx) PublicStructureBlockTradesSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "public-struc-block-trades", assetType, pair)
}

// BlockTickerSubscription a method to subscribe and unsubscribe to a "block-tickers" channel to retrieve the latest block trading volume in the last 24 hours.
// The data will be pushed when triggered by transaction execution event. In addition, it will also be pushed in 5 minutes interval according to subscription granularity.
func (ok *Okx) BlockTickerSubscription(operation string, assetType asset.Item, pair currency.Pair) (*SubscriptionOperationResponse, error) {
	return ok.WsChannelSubscription(operation, "block-tickers", assetType, pair, false, true)
}
