package lakebtc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ws/monitor"
	log "github.com/thrasher-/gocryptotrader/logger"
	"github.com/toorop/go-pusher"
)

const (
	lakeBTCWSURL         = "ws.lakebtc.com:8085"
	marketGlobalEndpoint = "market-global"
	marketSubstring      = "market-"
	globalSubstring      = "-global"
	volumeString         = "volume"
	highString           = "high"
	lowString            = "low"
	wssSchem             = "wss"
)

// WsConnect initiates a new websocket connection
func (l *LakeBTC) WsConnect() error {
	if !l.Websocket.IsEnabled() || !l.IsEnabled() {
		return errors.New(monitor.WebsocketNotEnabled)
	}
	var err error
	l.WebsocketConn.Client, err = pusher.NewCustomClient(strings.ToLower(l.Name), lakeBTCWSURL, wssSchem)
	if err != nil {
		return err
	}
	err = l.WebsocketConn.Client.Subscribe(marketGlobalEndpoint)
	if err != nil {
		return err
	}
	l.GenerateDefaultSubscriptions()
	err = l.listenToEndpoints()
	if err != nil {
		return err
	}
	go l.wsHandleIncomingData()
	return nil
}

func (l *LakeBTC) listenToEndpoints() error {
	var err error
	l.WebsocketConn.Ticker, err = l.WebsocketConn.Client.Bind("tickers")
	if err != nil {
		return fmt.Errorf("%s Websocket Bind error: %s", l.GetName(), err)
	}
	l.WebsocketConn.Orderbook, err = l.WebsocketConn.Client.Bind("update")
	if err != nil {
		return fmt.Errorf("%s Websocket Bind error: %s", l.GetName(), err)
	}
	l.WebsocketConn.Trade, err = l.WebsocketConn.Client.Bind("trades")
	if err != nil {
		return fmt.Errorf("%s Websocket Bind error: %s", l.GetName(), err)
	}
	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (l *LakeBTC) GenerateDefaultSubscriptions() {
	var subscriptions []monitor.WebsocketChannelSubscription
	enabledCurrencies := l.GetEnabledCurrencies()
	for j := range enabledCurrencies {
		enabledCurrencies[j].Delimiter = ""
		channel := fmt.Sprintf("%v%v%v", marketSubstring, enabledCurrencies[j].Lower(), globalSubstring)
		subscriptions = append(subscriptions, monitor.WebsocketChannelSubscription{
			Channel:  channel,
			Currency: enabledCurrencies[j],
		})
	}
	l.Websocket.SubscribeToChannels(subscriptions)
}

// Subscribe sends a websocket message to receive data from the channel
func (l *LakeBTC) Subscribe(channelToSubscribe monitor.WebsocketChannelSubscription) error {
	return l.WebsocketConn.Client.Subscribe(channelToSubscribe.Channel)
}

// wsHandleIncomingData services incoming data from the websocket connection
func (l *LakeBTC) wsHandleIncomingData() {
	l.Websocket.Wg.Add(1)

	defer func() {
		l.Websocket.Wg.Done()
	}()

	for {
		select {
		case <-l.Websocket.ShutdownC:
			return
		case data := <-l.WebsocketConn.Ticker:
			if l.Verbose {
				log.Debugf("%v Websocket message received: %v", l.Name, data)
			}
			err := l.processTicker(data.Data)
			if err != nil {
				l.Websocket.DataHandler <- err
				return
			}
		case data := <-l.WebsocketConn.Trade:
			if l.Verbose {
				log.Debugf("%v Websocket message received: %v", l.Name, data)
			}
			err := l.processTrades(data.Data, data.Channel)
			if err != nil {
				l.Websocket.DataHandler <- err
				return
			}
		case data := <-l.WebsocketConn.Orderbook:
			if l.Verbose {
				log.Debugf("%v Websocket message received: %v", l.Name, data)
			}
			err := l.processOrderbook(data.Data, data.Channel)
			if err != nil {
				l.Websocket.DataHandler <- err
				return
			}
		}
	}
}

func (l *LakeBTC) processTrades(data, channel string) error {
	var tradeData WsTrades
	err := common.JSONDecode([]byte(data), &tradeData)
	if err != nil {
		return err
	}
	curr := l.getCurrencyFromChannel(channel)
	for i := 0; i < len(tradeData.Trades); i++ {
		l.Websocket.DataHandler <- monitor.TradeData{
			Timestamp:    time.Unix(tradeData.Trades[i].Date, 0),
			CurrencyPair: curr,
			AssetType:    orderbook.Spot,
			Exchange:     l.GetName(),
			EventType:    orderbook.Spot,
			EventTime:    tradeData.Trades[i].Date,
			Price:        tradeData.Trades[i].Price,
			Amount:       tradeData.Trades[i].Amount,
			Side:         tradeData.Trades[i].Type,
		}
	}
	return nil
}

func (l *LakeBTC) processOrderbook(obUpdate, channel string) error {
	var update WsOrderbookUpdate
	err := common.JSONDecode([]byte(obUpdate), &update)
	if err != nil {
		return err
	}
	l.Websocket.TrafficAlert <- struct{}{}
	book := orderbook.Base{
		Pair:         l.getCurrencyFromChannel(channel),
		LastUpdated:  time.Now(),
		AssetType:    orderbook.Spot,
		ExchangeName: l.Name,
	}

	for i := 0; i < len(update.Asks); i++ {
		var amount, price float64
		amount, err = strconv.ParseFloat(update.Asks[i][1], 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing ticker data 'low' %v", l.Name, update.Asks[i])
			continue
		}
		price, err = strconv.ParseFloat(update.Asks[i][0], 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing orderbook price %v", l.Name, update.Asks[i])
			continue
		}
		book.Asks = append(book.Asks, orderbook.Item{
			Amount: amount,
			Price:  price,
		})
	}
	for i := 0; i < len(update.Bids); i++ {
		var amount, price float64
		amount, err = strconv.ParseFloat(update.Bids[i][1], 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing ticker data 'low' %v", l.Name, update.Bids[i])
			continue
		}
		price, err = strconv.ParseFloat(update.Bids[i][0], 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing orderbook price %v", l.Name, update.Bids[i])
			continue
		}
		book.Bids = append(book.Bids, orderbook.Item{
			Amount: amount,
			Price:  price,
		})
	}
	log.Debug(book)
	err = l.Websocket.Orderbook.LoadSnapshot(&book, book.ExchangeName, true)
	if err != nil {
		return err
	}
	return nil
}

func (l *LakeBTC) getCurrencyFromChannel(channel string) currency.Pair {
	curr := strings.Replace(channel, marketSubstring, "", 1)
	curr = strings.Replace(curr, globalSubstring, "", 1)
	return currency.NewPairFromString(curr)
}

func (l *LakeBTC) processTicker(ticker string) error {
	var tUpdate map[string]interface{}
	err := common.JSONDecode([]byte(ticker), &tUpdate)
	if err != nil {
		l.Websocket.DataHandler <- err
		return err
	}
	l.Websocket.TrafficAlert <- struct{}{}
	for k, v := range tUpdate {
		tickerData := v.(map[string]interface{})
		if tickerData[highString] == nil || tickerData[lowString] == nil || tickerData[volumeString] == nil {
			continue
		}
		high, err := strconv.ParseFloat(tickerData[highString].(string), 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing ticker data 'high' %v", l.Name, tickerData)
			continue
		}
		low, err := strconv.ParseFloat(tickerData[lowString].(string), 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing ticker data 'low' %v", l.Name, tickerData)
			continue
		}
		vol, err := strconv.ParseFloat(tickerData[volumeString].(string), 64)
		if err != nil {
			l.Websocket.DataHandler <- fmt.Errorf("%v error parsing ticker data 'volume' %v", l.Name, tickerData)
			continue
		}
		l.Websocket.DataHandler <- monitor.TickerData{
			Timestamp: time.Now(),
			Pair:      currency.NewPairFromString(k),
			AssetType: orderbook.Spot,
			Exchange:  l.GetName(),
			Quantity:  vol,
			HighPrice: high,
			LowPrice:  low,
		}
	}
	return nil
}
