package kraken

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// List of all websocket channels to subscribe to
const (
	krakenWSURL              = "wss://ws.kraken.com"
	krakenWSSandboxURL       = "wss://sandbox.kraken.com"
	krakenWSSupportedVersion = "0.1.1"
	// If a checksum fails, then resubscribing to the channel fails, fatal after these attempts
	krakenWsResubscribeFailureLimit   = 3
	krakenWsResubscribeDelayInSeconds = 3

	krakenWsHeartbeat          = "heartbeat"
	krakenWsPing               = "ping"
	krakenWsPong               = "pong"
	krakenWsSystemStatus       = "systemStatus"
	krakenWsSubscribe          = "subscribe"
	krakenWsSubscriptionStatus = "subscriptionStatus"
	krakenWsUnsubscribe        = "unsubscribe"
	krakenWsTicker             = "ticker"
	krakenWsOHLC               = "ohlc"
	krakenWsTrade              = "trade"
	krakenWsSpread             = "spread"
	krakenWsOrderbook          = "book"
	// Spot endpoints

)

// orderbookMutex Ensures if two entries arrive at once, only one can be processed at a time
var orderbookMutex sync.Mutex
var subscriptionChannelPair []WebsocketChannelData

// writeToWebsocket sends a message to the websocket endpoint
func (k *Kraken) writeToWebsocket(message string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.Verbose {
		log.Debugf("Sending message to WS: %v", message)
	}
	return k.WebsocketConn.WriteMessage(websocket.TextMessage, []byte(message))
}

// WsConnect initiates a websocket connection
func (k *Kraken) WsConnect() error {
	if !k.Websocket.IsEnabled() || !k.IsEnabled() {
		return errors.New(exchange.WebsocketNotEnabled)
	}

	var dialer websocket.Dialer
	if k.Websocket.GetProxyAddress() != "" {
		proxy, err := url.Parse(k.Websocket.GetProxyAddress())
		if err != nil {
			return err
		}

		dialer.Proxy = http.ProxyURL(proxy)
	}

	var err error
	if k.Verbose {
		log.Debugf("Attempting to connect to %v", k.Websocket.GetWebsocketURL())
	}
	k.WebsocketConn, _, err = dialer.Dial(k.Websocket.GetWebsocketURL(),
		http.Header{})
	if err != nil {
		return fmt.Errorf("%s Unable to connect to Websocket. Error: %s",
			k.Name,
			err)
	}
	if k.Verbose {
		log.Debugf("Successful connection to %v", k.Websocket.GetWebsocketURL())
	}

	go k.WsHandleData()
	go k.wsPingHandler()

	err = k.WsSubscribeToDefaults()
	if err != nil {
		return fmt.Errorf("error: Could not subscribe to the %v websocket %s",
			k.GetName(), err)
	}
	return nil
}

// WsSubscribeToDefaults subscribes to the websocket channels
func (k *Kraken) WsSubscribeToDefaults() (err error) {
	channelsToSubscribe := []string{krakenWsTicker, krakenWsTrade, krakenWsOrderbook, krakenWsOHLC, krakenWsSpread}
	for _, pair := range k.EnabledPairs {
		formattedPair := strings.ToUpper(strings.Replace(pair.String(), "-", "/", 1))
		for _, channel := range channelsToSubscribe {
			err = k.WsSubscribeToChannel(channel, []string{formattedPair}, 0)
			if err != nil {
				return
			}
		}
	}
	return nil
}

// WsReadData reads data from the websocket connection
func (k *Kraken) WsReadData() (exchange.WebsocketResponse, error) {
	mType, resp, err := k.WebsocketConn.ReadMessage()
	if err != nil {
		return exchange.WebsocketResponse{}, err
	}

	k.Websocket.TrafficAlert <- struct{}{}
	var standardMessage []byte
	switch mType {
	case websocket.TextMessage:
		standardMessage = resp

	case websocket.BinaryMessage:
		reader := flate.NewReader(bytes.NewReader(resp))
		standardMessage, err = ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			return exchange.WebsocketResponse{}, err
		}
	}
	if k.Verbose {
		log.Debugf("%v Websocket message received: %v", k.Name, string(standardMessage))
	}

	return exchange.WebsocketResponse{Raw: standardMessage}, nil
}

// wsPingHandler sends a message "ping" every 27 to maintain the connection to the websocket
func (k *Kraken) wsPingHandler() {
	k.Websocket.Wg.Add(1)
	defer k.Websocket.Wg.Done()
	ticker := time.NewTicker(time.Second * 27)
	for {
		select {
		case <-k.Websocket.ShutdownC:
			return

		case <-ticker.C:
			pingEvent := fmt.Sprintf("{\"event\":\"%v\"}", krakenWsPing)
			err := k.writeToWebsocket(pingEvent)
			if k.Verbose {
				log.Debugf("%v sending ping", k.GetName())
			}
			if err != nil {
				k.Websocket.DataHandler <- err
				return
			}
		}
	}
}

// WsHandleData handles the read data from the websocket connection
func (k *Kraken) WsHandleData() {
	k.Websocket.Wg.Add(1)
	defer func() {
		err := k.WebsocketConn.Close()
		if err != nil {
			k.Websocket.DataHandler <- fmt.Errorf("okex_websocket.go - Unable to to close Websocket connection. Error: %s",
				err)
		}
		k.Websocket.Wg.Done()
	}()

	for {
		select {
		case <-k.Websocket.ShutdownC:
			return
		default:
			resp, err := k.WsReadData()
			if err != nil {
				k.Websocket.DataHandler <- err
				return
			}
			// event response handling
			var eventResponse WebsocketEventResponse
			err = common.JSONDecode(resp.Raw, &eventResponse)
			if err == nil && eventResponse.Event != "" {
				k.WsHandleEventResponse(&eventResponse)
				continue
			}
			// data handling
			var dataResponse WebsocketDataResponse
			err = common.JSONDecode(resp.Raw, &dataResponse)
			log.Debug(dataResponse[0])
			if err == nil && dataResponse[0].(float64) >= 0 {
				k.WsHandleDataResponse(dataResponse)
				continue
			}

			// Unknown data handling
			k.Websocket.DataHandler <- fmt.Errorf("unrecognised response: %v", string(resp.Raw))
			continue
		}
	}
}

// WsHandleDataResponse classifies the WS response and sends to appropriate handler
func (k *Kraken) WsHandleDataResponse(response WebsocketDataResponse) {
	channelID := response[0].(float64)
	channelData := getSubscriptionChannelData(channelID)
	switch channelData.Subscription {
	case krakenWsTicker:
		if k.Verbose {
			log.Debugf("%v Websocket ticker data received", k.GetName())
		}
		k.wsProcessTickers(channelData, response[1])
	case krakenWsOHLC:
		if k.Verbose {
			log.Debugf("%v Websocket OHLC data received", k.GetName())
		}
		k.wsProcessCandles(channelData, response[1])
	case krakenWsOrderbook:
		if k.Verbose {
			log.Debugf("%v Websocket Orderbook data received", k.GetName())
		}
	case krakenWsSpread:
		if k.Verbose {
			log.Debugf("%v Websocket Spread data received", k.GetName())
		}
	case krakenWsTrade:
		if k.Verbose {
			log.Debugf("%v Websocket Trade data received", k.GetName())
		}
		k.wsProcessTrades(channelData, response[1])
	default:
		log.Errorf("%v Unidentified websocket data received: %v", k.GetName(), response)
	}
}

// WsHandleDataResponse classifies the WS response and sends to appropriate handler
func (k *Kraken) WsHandleEventResponse(response *WebsocketEventResponse) {
	switch response.Event {
	case krakenWsHeartbeat:
		if k.Verbose {
			log.Debugf("%v Websocket heartbeat data received", k.GetName())
		}
	case krakenWsPong:
		if k.Verbose {
			log.Debugf("%v Websocket pong data received", k.GetName())
		}
	case krakenWsSystemStatus:
		if k.Verbose {
			log.Debugf("%v Websocket status data received", k.GetName())
		}
		if response.Status != "online" {
			k.Websocket.DataHandler <- fmt.Errorf("%v Websocket status '%v'", k.GetName(), response.Status)
		}
		if response.WebsocketStatusResponse.Version != krakenWSSupportedVersion {
			log.Warnf("%v New version of Websocket API released. Was %v Now %v", k.GetName(), krakenWSSupportedVersion, response.WebsocketStatusResponse.Version)
		}
	case krakenWsSubscriptionStatus:
		if k.Verbose {
			log.Debugf("%v Websocket subscription status data received", k.GetName())
		}
		if response.Status != "subscribed" {
			k.Websocket.DataHandler <- fmt.Errorf(response.WebsocketErrorResponse.ErrorMessage)
			k.ResubscribeToChannel(response.Subscription.Name, response.Pair)
			return
		}
		addNewSubscriptionChannelData(response)
	default:
		log.Errorf("%v Unidentified websocket data received: %v", k.GetName(), response)
	}
}

// WsSubscribeToChannel sends a request to WS to subscribe to supplied channel name and pairs
func (k *Kraken) WsSubscribeToChannel(topic string, currencies []string, requestID int64) error {
	resp := WebsocketSubscriptionEventRequest{
		Event: krakenWsSubscribe,
		Pairs: currencies,
		Subscription: WebsocketSubscriptionData{
			Name: topic,
		},
	}
	if requestID > 0 {
		resp.RequestID = requestID
	}
	json, err := common.JSONEncode(resp)
	if err != nil {
		return err
	}
	err = k.writeToWebsocket(string(json))
	if err != nil {
		return err
	}
	return nil
}

// WsUnsubscribeToChannel sends a request to WS to unsubscribe to supplied channel name and pairs
func (k *Kraken) WsUnsubscribeToChannel(topic string, currencies []string, requestID int64) error {
	resp := WebsocketUnsubscribeEventRequest{
		Event: krakenWsUnsubscribe,
		Pairs: currencies,
		Subscription: WebsocketSubscriptionData{
			Name: topic,
		},
	}
	if requestID > 0 {
		resp.RequestID = requestID
	}
	json, err := common.JSONEncode(resp)
	if err != nil {
		return err
	}
	err = k.writeToWebsocket(string(json))
	if err != nil {
		return err
	}
	return nil
}

// WsUnsubscribeToChannelByChannelID sends a request to WS to unsubscribe to supplied channel ID
func (k *Kraken) WsUnsubscribeToChannelByChannelID(channelID int64) error {
	resp := WebsocketUnsubscribeEventRequest{
		Event:     krakenWsUnsubscribe,
		ChannelID: channelID,
	}
	json, err := common.JSONEncode(resp)
	if err != nil {
		return err
	}
	err = k.writeToWebsocket(string(json))
	if err != nil {
		return err
	}
	return nil
}

func addNewSubscriptionChannelData(response *WebsocketEventResponse) {
	for i := range subscriptionChannelPair {
		if response.ChannelID == subscriptionChannelPair[i].ChannelID {
			return
		}
	}
	subscriptionChannelPair = append(subscriptionChannelPair, WebsocketChannelData{
		Subscription: response.Subscription.Name,
		Pair:         response.Pair,
		ChannelID:    response.ChannelID,
	})
}

func getSubscriptionChannelData(id float64) WebsocketChannelData {
	for i := range subscriptionChannelPair {
		if id == subscriptionChannelPair[i].ChannelID {
			return subscriptionChannelPair[i]
		}
	}
	return WebsocketChannelData{}
}

// resubscribeToChannel will attempt to unsubscribe and resubscribe to a channel
func (k *Kraken) ResubscribeToChannel(channel string, pair currency.Pair) {
	if krakenWsResubscribeFailureLimit > 0 {
		var successfulUnsubscribe bool
		for i := 0; i < krakenWsResubscribeFailureLimit; i++ {
			err := k.WsUnsubscribeToChannel(channel, []string{pair.String()}, 0)
			if err != nil {
				log.Error(err)
				time.Sleep(krakenWsResubscribeDelayInSeconds * time.Second)
				continue
			}
			successfulUnsubscribe = true
			break
		}
		if !successfulUnsubscribe {
			log.Fatalf("%v websocket channel %v failed to unsubscribe after %v attempts", k.GetName(), channel, krakenWsResubscribeFailureLimit)
		}
		successfulSubscribe := true
		for i := 0; i < krakenWsResubscribeFailureLimit; i++ {
			err := k.WsSubscribeToChannel(channel, []string{pair.String()}, 0)
			if err != nil {
				log.Error(err)
				time.Sleep(krakenWsResubscribeDelayInSeconds * time.Second)
				continue
			}
			successfulSubscribe = true
			break
		}
		if !successfulSubscribe {
			log.Fatalf("%v websocket channel %v failed to resubscribe after %v attempts", k.GetName(), channel, krakenWsResubscribeFailureLimit)
		}
	} else {
		log.Fatalf("%v websocket channel %v cannot resubscribe. Limit: %v", k.GetName(), channel, krakenWsResubscribeFailureLimit)
	}
}

// wsProcessTickers converts ticker data and sends it to the datahandler
func (k *Kraken) wsProcessTickers(channelData WebsocketChannelData, data interface{}) {
	tickerData := data.(map[string]interface{})
	closeData := tickerData["c"].([]interface{})
	openData := tickerData["o"].([]interface{})
	lowData := tickerData["l"].([]interface{})
	highData := tickerData["h"].([]interface{})
	volumeData := tickerData["v"].([]interface{})
	closePrice, _ := strconv.ParseFloat(closeData[0].(string), 64)
	openPrice, _ := strconv.ParseFloat(openData[0].(string), 64)
	highPrice, _ := strconv.ParseFloat(highData[0].(string), 64)
	lowPrice, _ := strconv.ParseFloat(lowData[0].(string), 64)
	quantity, _ := strconv.ParseFloat(volumeData[0].(string), 64)

	k.Websocket.DataHandler <- exchange.TickerData{
		Timestamp:  time.Now(),
		Exchange:   k.GetName(),
		AssetType:  "SPOT",
		Pair:       channelData.Pair,
		ClosePrice: closePrice,
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		Quantity:   quantity,
	}
}

// wsProcessTrades converts trade data and sends it to the datahandler
func (k *Kraken) wsProcessTrades(channelData WebsocketChannelData, data interface{}) {
	tradeData := data.([]interface{})
	for i := range tradeData {
		trade := tradeData[i].([]interface{})
		timeData, _ := strconv.ParseInt(trade[2].(string), 10, 64)
		timeUnix := time.Unix(timeData, 0)
		price, _ := strconv.ParseFloat(trade[0].(string), 64)
		amount, _ := strconv.ParseFloat(trade[1].(string), 64)

		k.Websocket.DataHandler <- exchange.TradeData{
			AssetType:    "SPOT",
			CurrencyPair: channelData.Pair,
			EventTime:    time.Now().Unix(),
			Exchange:     k.GetName(),
			Price:        price,
			Amount:       amount,
			Timestamp:    timeUnix,
			Side:         trade[3].(string),
		}
	}
}

// wsProcessCandles converts candle data and sends it to the data handler
func (k *Kraken) wsProcessCandles(channelData WebsocketChannelData, data interface{}) {
	candleData := data.([]interface{})
	startTimeData, _ := strconv.ParseInt(candleData[0].(string), 10, 64)
	startTimeUnix := time.Unix(startTimeData, 0)
	endTimeData, _ := strconv.ParseInt(candleData[1].(string), 10, 64)
	endTimeUnix := time.Unix(endTimeData, 0)
	high, _ := strconv.ParseFloat(candleData[3].(string), 64)
	low, _ := strconv.ParseFloat(candleData[4].(string), 64)
	open, _ := strconv.ParseFloat(candleData[2].(string), 64)
	close, _ := strconv.ParseFloat(candleData[5].(string), 64)
	volume, _ := strconv.ParseFloat(candleData[7].(string), 64)

	k.Websocket.DataHandler <- exchange.KlineData{
		AssetType:  "SPOT",
		Pair:       channelData.Pair,
		Timestamp:  time.Now(),
		Exchange:   k.GetName(),
		StartTime:  startTimeUnix,
		CloseTime:  endTimeUnix,
		Interval:   "60",
		HighPrice:  high,
		LowPrice:   low,
		OpenPrice:  open,
		ClosePrice: close,
		Volume:     volume,
	}
}
