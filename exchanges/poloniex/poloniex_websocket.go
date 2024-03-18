package poloniex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	poloniexWebsocketAddress        = "wss://ws.poloniex.com/ws/public"
	poloniexPrivateWebsocketAddress = "wss://ws.poloniex.com/ws/private"

	cnlExchange = "exchange"
	cnlAuth     = "auth"

	// Public channels
	cnlSymbols    = "symbols"
	cnlCurrencies = "currencies"
	cnlCandles    = "candles"
	cnlTrades     = "trades"
	cnlTicker     = "ticker"
	cnlBooks      = "book"
	cnlBookLevel2 = "book_lv2"

	// Authenticated channels
	cnlOrders   = "orders"
	cnlBalances = "balances"
)

var defaultSubscriptions = []string{
	cnlCandles,
	cnlTrades,
	cnlTicker,
	cnlBooks,
}

var onceOrderbook map[string]struct{}

// WsConnect initiates a websocket connection
func (p *Poloniex) WsConnect() error {
	if !p.Websocket.IsEnabled() || !p.IsEnabled() {
		return stream.ErrWebsocketNotEnabled
	}
	var dialer websocket.Dialer
	err := p.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	pingMessage := &struct {
		Event string `json:"event"`
	}{
		Event: "ping",
	}
	var pingPayload []byte
	pingPayload, err = json.Marshal(pingMessage)
	if err != nil {
		return err
	}
	p.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.TextMessage,
		Message:           pingPayload,
		Delay:             30,
	})
	p.Websocket.SetCanUseAuthenticatedEndpoints(true)
	if p.Websocket.CanUseAuthenticatedEndpoints() {
		err := p.wsAuthConn()
		if err != nil {
			p.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
	}
	onceOrderbook = make(map[string]struct{})
	p.Websocket.Wg.Add(1)
	go p.wsReadData(p.Websocket.Conn)
	return nil
}

func (p *Poloniex) wsAuthConn() error {
	creds, err := p.GetCredentials(context.Background())
	if err != nil {
		return err
	}

	var dialer websocket.Dialer
	err = p.Websocket.AuthConn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	pingMessage := &struct {
		Event string `json:"event"`
	}{
		Event: "ping",
	}
	var pingPayload []byte
	pingPayload, err = json.Marshal(pingMessage)
	if err != nil {
		return err
	}
	p.Websocket.AuthConn.SetupPingHandler(stream.PingHandler{
		UseGorillaHandler: true,
		MessageType:       websocket.TextMessage,
		Message:           pingPayload,
		Delay:             30,
	})
	p.Websocket.Wg.Add(1)
	go p.wsReadData(p.Websocket.AuthConn)
	timestamp := time.Now()
	hmac, err := crypto.GetHMAC(crypto.HashSHA256,
		[]byte(fmt.Sprintf("GET\n/ws\nsignTimestamp=%d", timestamp.UnixMilli())),
		[]byte(creds.Secret))
	if err != nil {
		return err
	}
	auth := &struct {
		Event   string     `json:"event"`
		Channel []string   `json:"channel"`
		Params  AuthParams `json:"params"`
	}{
		Event:   "subscribe",
		Channel: []string{cnlAuth},
		Params: AuthParams{
			Key:             creds.Key,
			SignatureMethod: "hmacSHA256",
			SignTimestamp:   timestamp.UnixMilli(),
			Signature:       crypto.Base64Encode(hmac),
		},
	}
	return p.Websocket.AuthConn.SendJSONMessage(auth)
}

// wsReadData handles data from the websocket connection
func (p *Poloniex) wsReadData(conn stream.Connection) {
	defer p.Websocket.Wg.Done()
	for {
		resp := conn.ReadMessage()
		if resp.Raw == nil {
			return
		}
		err := p.wsHandleData(resp.Raw)
		if err != nil {
			p.Websocket.DataHandler <- fmt.Errorf("%s: %w", p.Name, err)
		}
	}
}

func (p *Poloniex) wsHandleData(respRaw []byte) error {
	var result SubscriptionResponse
	err := json.Unmarshal(respRaw, &result)
	if err != nil {
		return err
	}
	if result.ID != "" {
		if !p.Websocket.Match.IncomingWithData(result.ID, respRaw) {
			return fmt.Errorf("could not match trade response with ID: %s Event: %s ", result.ID, result.Event)
		}
		return nil
	}
	if result.Event != "" {
		log.Debugf(log.ExchangeSys, string(respRaw))
		return nil
	}
	switch result.Channel {
	case cnlSymbols:
		var response [][]WsSymbol
		return p.processResponse(&result, &response)
	case cnlCurrencies:
		var response [][]WsCurrency
		return p.processResponse(&result, &response)
	case cnlExchange:
		var response WsExchangeStatus
		return p.processResponse(&result, &response)
	case cnlTrades:
		return p.processTrades(&result)
	case cnlTicker:
		return p.processTicker(&result)
	case cnlBooks,
		cnlBookLevel2:
		return p.processBooks(&result)
	case cnlOrders:
		return p.processOrders(&result)
	case cnlBalances:
		return p.processBalance(&result)
	case cnlAuth:
		resp := &WebsocketAuthenticationResponse{}
		err = json.Unmarshal(result.Data, &resp)
		if err != nil {
			return err
		}
		if !resp.Success {
			log.Errorf(log.ExchangeSys, "%s Websocket: %s", p.Name, resp.Message)
			return nil
		}
		if p.Verbose {
			log.Debugf(log.ExchangeSys, "%s Websocket: connection authenticated\n", p.Name)
		}
	default:
		if strings.HasPrefix(result.Channel, cnlCandles) {
			return p.processCandlestickData(&result)
		}
		p.Websocket.DataHandler <- stream.UnhandledMessageWarning{Message: p.Name + stream.UnhandledMessage + string(respRaw)}
		return fmt.Errorf("%s unhandled message: %s", p.Name, string(respRaw))
	}
	return nil
}

func (p *Poloniex) processBalance(result *SubscriptionResponse) error {
	var resp WsTradeBalance
	err := json.Unmarshal(result.Data, &resp)
	if err != nil {
		return err
	}
	accountChanges := make([]account.Change, len(resp))
	for x := range resp {
		accountChanges[x] = account.Change{
			Exchange: p.Name,
			Currency: currency.NewCode(resp[x].Currency),
			Asset:    stringToAccountType(resp[x].AccountType),
			Amount:   resp[x].Available.Float64(),
			Account:  resp[x].AccountType,
		}
	}
	p.Websocket.DataHandler <- accountChanges
	return nil
}

func (p *Poloniex) processOrders(result *SubscriptionResponse) error {
	response := []WebsocketTradeOrder{}
	err := json.Unmarshal(result.Data, &response)
	if err != nil {
		return err
	}
	orderDetails := make([]order.Detail, len(response))
	for x := range response {
		pair, err := currency.NewPairFromString(response[x].Symbol)
		if err != nil {
			return err
		}
		oType, err := order.StringToOrderType(response[x].Type)
		if err != nil {
			return err
		}
		oSide, err := order.StringToOrderSide(response[x].Side)
		if err != nil {
			return err
		}
		oStatus, err := order.StringToOrderStatus(response[x].State)
		if err != nil {
			return err
		}
		orderDetails[x] = order.Detail{
			Price:           response[x].Price.Float64(),
			Amount:          response[x].Quantity.Float64(),
			QuoteAmount:     response[x].OrderAmount.Float64(),
			ExecutedAmount:  response[x].FilledAmount.Float64(),
			RemainingAmount: response[x].OrderAmount.Float64() - response[x].FilledAmount.Float64(),
			Fee:             response[x].TradeFee.Float64(),
			FeeAsset:        currency.NewCode(response[x].FeeCurrency),
			Exchange:        p.Name,
			OrderID:         response[x].OrderID,
			ClientOrderID:   response[x].ClientOrderID,
			Type:            oType,
			Side:            oSide,
			Status:          oStatus,
			AssetType:       stringToAccountType(response[x].AccountType),
			Date:            response[x].CreateTime.Time(),
			LastUpdated:     response[x].TradeTime.Time(),
			Pair:            pair,
			Trades: []order.TradeHistory{
				{
					Price:     response[x].TradePrice.Float64(),
					Amount:    response[x].TradeQty.Float64(),
					Fee:       response[x].TradeFee.Float64(),
					Exchange:  p.Name,
					TID:       response[x].TradeID,
					Type:      oType,
					Side:      oSide,
					Timestamp: response[x].Timestamp.Time(),
					FeeAsset:  response[x].FeeCurrency,
					Total:     response[x].Quantity.Float64(),
				},
			},
		}
	}
	p.Websocket.DataHandler <- orderDetails
	return nil
}

func (p *Poloniex) processBooks(result *SubscriptionResponse) error {
	var resp []WsBook
	err := json.Unmarshal(result.Data, &resp)
	if err != nil {
		return err
	}
	for x := range resp {
		pair, err := currency.NewPairFromString(resp[x].Symbol)
		if err != nil {
			return err
		}
		_, okay := onceOrderbook[resp[x].Symbol]
		if !okay {
			if onceOrderbook == nil {
				onceOrderbook = make(map[string]struct{})
			}
			var orderbooks *orderbook.Base
			orderbooks, err = p.UpdateOrderbook(context.Background(), pair, asset.Spot)
			if err != nil {
				return err
			}
			err = p.Websocket.Orderbook.LoadSnapshot(orderbooks)
			if err != nil {
				return err
			}
			onceOrderbook[resp[x].Symbol] = struct{}{}
		}
		update := orderbook.Update{
			Pair:       pair,
			UpdateTime: resp[x].Timestamp.Time(),
			UpdateID:   resp[x].ID,
			Action:     orderbook.UpdateInsert,
			Asset:      asset.Spot,
		}
		update.Asks = make([]orderbook.Item, len(resp[x].Asks))
		for i := range resp[x].Asks {
			update.Asks[i] = orderbook.Item{
				Price:  resp[x].Asks[i][0].Float64(),
				Amount: resp[x].Asks[i][1].Float64(),
			}
		}
		update.Bids = make([]orderbook.Item, len(resp[x].Bids))
		for i := range resp[x].Bids {
			update.Bids[i] = orderbook.Item{
				Price:  resp[x].Bids[i][0].Float64(),
				Amount: resp[x].Bids[i][1].Float64(),
			}
		}
		update.UpdateID = resp[x].LastID
		err = p.Websocket.Orderbook.Update(&update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Poloniex) processTicker(result *SubscriptionResponse) error {
	var resp []WsTicker
	err := json.Unmarshal(result.Data, &resp)
	if err != nil {
		return err
	}
	tickerData := make([]ticker.Price, len(resp))
	for x := range resp {
		pair, err := currency.NewPairFromString(resp[x].Symbol)
		if err != nil {
			return err
		}
		tickerData[x] = ticker.Price{
			Last:         resp[x].MarkPrice.Float64(),
			High:         resp[x].High.Float64(),
			Low:          resp[x].Low.Float64(),
			Volume:       resp[x].Quantity.Float64(),
			QuoteVolume:  resp[x].Amount.Float64(),
			Open:         resp[x].Open.Float64(),
			Close:        resp[x].Close.Float64(),
			Pair:         pair,
			AssetType:    asset.Spot,
			ExchangeName: p.Name,
			LastUpdated:  resp[x].Timestamp.Time(),
		}
	}
	p.Websocket.DataHandler <- tickerData
	return nil
}

func (p *Poloniex) processTrades(result *SubscriptionResponse) error {
	var resp []WsTrade
	err := json.Unmarshal(result.Data, &resp)
	if err != nil {
		return err
	}
	trades := make([]trade.Data, len(resp))
	for x := range resp {
		pair, err := currency.NewPairFromString(resp[x].Symbol)
		if err != nil {
			return err
		}
		trades[x] = trade.Data{
			TID:          resp[x].ID,
			Exchange:     p.Name,
			CurrencyPair: pair,
			Price:        resp[x].Price.Float64(),
			Amount:       resp[x].Amount.Float64(),
			Timestamp:    resp[x].Timestamp.Time(),
		}
	}
	return trade.AddTradesToBuffer(p.Name, trades...)
}

func (p *Poloniex) processCandlestickData(result *SubscriptionResponse) error {
	var resp []WsCandles
	err := json.Unmarshal(result.Data, &resp)
	if err != nil {
		return err
	}
	var pair currency.Pair
	candles := make([]stream.KlineData, len(resp))
	for x := range resp {
		pair, err = currency.NewPairFromString(resp[x].Symbol)
		if err != nil {
			return err
		}
		candles[x] = stream.KlineData{
			Pair:       pair,
			Exchange:   p.Name,
			Timestamp:  resp[x].Timestamp.Time(),
			StartTime:  resp[x].StartTime.Time(),
			CloseTime:  resp[x].CloseTime.Time(),
			OpenPrice:  resp[x].Open.Float64(),
			ClosePrice: resp[x].Close.Float64(),
			HighPrice:  resp[x].High.Float64(),
			LowPrice:   resp[x].Low.Float64(),
			Volume:     resp[x].Quantity.Float64(),
		}
	}
	p.Websocket.DataHandler <- candles
	return nil
}

func (p *Poloniex) processResponse(result *SubscriptionResponse, instance interface{}) error {
	err := json.Unmarshal(result.Data, instance)
	if err != nil {
		return err
	}
	fullResp := result.GetWsResponse()
	fullResp.Data = instance
	p.Websocket.DataHandler <- fullResp
	return nil
}

// GenerateDefaultSubscriptions Adds default subscriptions to websocket to be handled by ManageSubscriptions()
func (p *Poloniex) GenerateDefaultSubscriptions() ([]subscription.Subscription, error) {
	enabledCurrencies, err := p.GetEnabledPairs(asset.Spot)
	if err != nil {
		return nil, err
	}
	channels := defaultSubscriptions
	if p.Websocket.CanUseAuthenticatedEndpoints() {
		channels = append(channels, []string{
			cnlOrders,
			cnlBalances}...)
	}
	subscriptions := make([]subscription.Subscription, 0, 6*len(enabledCurrencies))
	for i := range channels {
		switch channels[i] {
		case cnlSymbols, cnlTrades, cnlTicker, cnlBooks, cnlBookLevel2:
			for x := range enabledCurrencies {
				var params map[string]interface{}
				if channels[i] == cnlBooks {
					params = map[string]interface{}{
						"depth": 20,
					}
				}
				subscriptions = append(subscriptions, subscription.Subscription{
					Pair:    enabledCurrencies[x],
					Channel: channels[i],
					Params:  params,
				})
			}
		case cnlCurrencies:
			currencyMaps := make(map[currency.Code]struct{})
			for x := range enabledCurrencies {
				_, okay := currencyMaps[enabledCurrencies[x].Base]
				if !okay {
					subscriptions = append(subscriptions, subscription.Subscription{
						Channel: channels[i],
						Pair:    currency.Pair{Base: enabledCurrencies[x].Base},
					})
					currencyMaps[enabledCurrencies[x].Base] = struct{}{}
				}
				_, okay = currencyMaps[enabledCurrencies[x].Quote]
				if !okay {
					subscriptions = append(subscriptions, subscription.Subscription{
						Channel: channels[i],
						Pair:    currency.Pair{Base: enabledCurrencies[x].Quote},
					})
					currencyMaps[enabledCurrencies[x].Quote] = struct{}{}
				}
			}
		case cnlCandles:
			for x := range enabledCurrencies {
				subscriptions = append(subscriptions, subscription.Subscription{
					Channel: channels[i],
					Pair:    enabledCurrencies[x],
					Params: map[string]interface{}{
						"interval": kline.FiveMin,
					},
				})
			}
		case cnlOrders, cnlBalances, cnlExchange:
			subscriptions = append(subscriptions, subscription.Subscription{
				Channel: channels[i],
			})
		}
	}
	return subscriptions, nil
}

func (p *Poloniex) handleSubscriptions(operation string, subscs []subscription.Subscription) ([]SubscriptionPayload, error) {
	pairsMap := map[string]*SubscriptionPayload{}
	currencies := make(map[string]struct{})
	payloads := []SubscriptionPayload{}
	for x := range subscs {
		switch subscs[x].Channel {
		case cnlSymbols, cnlTrades, cnlTicker, cnlBooks, cnlBookLevel2:
			sp, okay := pairsMap[subscs[x].Channel]
			if !okay {
				sp = &SubscriptionPayload{
					Event:   operation,
					Channel: []string{subscs[x].Channel},
					Symbols: []string{},
				}
				pairsMap[subscs[x].Channel] = sp
			}
			sp.Symbols = append(sp.Symbols, subscs[x].Pair.String())
			if subscs[x].Channel == cnlBooks {
				depth, okay := subscs[x].Params["depth"]
				if okay {
					sp.Depth, _ = depth.(int64)
				}
			}
		case cnlCurrencies:
			sp, okay := pairsMap[subscs[x].Channel]
			if !okay {
				sp = &SubscriptionPayload{
					Event:      operation,
					Channel:    []string{subscs[x].Channel},
					Currencies: []string{},
				}
				pairsMap[subscs[x].Channel] = sp
			}
			_, okay = currencies[subscs[x].Pair.Base.Upper().String()]
			if !okay {
				sp.Currencies = append(sp.Currencies, subscs[x].Pair.Base.String())
				currencies[subscs[x].Pair.Base.Upper().String()] = struct{}{}
			}
			currencies[subscs[x].Pair.Base.String()] = struct{}{}
		case cnlCandles:
			interval, okay := subscs[x].Params["interval"].(kline.Interval)
			if !okay {
				interval = kline.FiveMin
			}
			intervalString, err := intervalToString(interval)
			if err != nil {
				return nil, err
			}
			channelName := fmt.Sprintf("%s_%s", subscs[x].Channel, strings.ToLower(intervalString))
			sp, okay := pairsMap[channelName]
			if !okay {
				sp = &SubscriptionPayload{
					Event:   operation,
					Channel: []string{channelName},
					Symbols: []string{},
				}
				pairsMap[channelName] = sp
			}
			sp.Symbols = append(sp.Symbols, subscs[x].Pair.String())
		case cnlOrders:
			payloads = append(payloads, SubscriptionPayload{
				Event:   operation,
				Channel: []string{subscs[x].Channel},
				Symbols: []string{"all"},
			})
		case cnlBalances, cnlExchange:
			payloads = append(payloads, SubscriptionPayload{
				Event:   operation,
				Channel: []string{subscs[x].Channel},
			})
		default:
			return nil, errChannelNotSupported
		}
	}
	for _, val := range pairsMap {
		payloads = append(payloads, *val)
	}
	return payloads, nil
}

// Subscribe sends a websocket message to receive data from the channel
func (p *Poloniex) Subscribe(subs []subscription.Subscription) error {
	var canUseAuthenticate bool
	if p.Websocket.CanUseAuthenticatedEndpoints() {
		canUseAuthenticate = true
	}
	payloads, err := p.handleSubscriptions("subscribe", subs)
	if err != nil {
		return err
	}
	for i := range payloads {
		switch payloads[i].Channel[0] {
		case cnlBalances, cnlOrders:
			if canUseAuthenticate {
				err = p.Websocket.AuthConn.SendJSONMessage(payloads[i])
				if err != nil {
					return err
				}
			}
		default:
			err = p.Websocket.Conn.SendJSONMessage(payloads[i])
			if err != nil {
				return err
			}
		}
	}
	p.Websocket.AddSuccessfulSubscriptions(subs...)
	return nil
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (p *Poloniex) Unsubscribe(unsub []subscription.Subscription) error {
	var canUseAuthenticate bool
	if p.IsWebsocketAuthenticationSupported() && p.Websocket.CanUseAuthenticatedEndpoints() {
		canUseAuthenticate = true
	}
	payloads, err := p.handleSubscriptions("unsubscribe", unsub)
	if err != nil {
		return err
	}
	for i := range payloads {
		switch payloads[i].Channel[0] {
		case cnlBalances, cnlOrders:
			if canUseAuthenticate {
				err = p.Websocket.AuthConn.SendJSONMessage(payloads[i])
				if err != nil {
					return err
				}
			}
		default:
			err = p.Websocket.Conn.SendJSONMessage(payloads[i])
			if err != nil {
				return err
			}
		}
	}
	p.Websocket.RemoveSubscriptions(unsub...)
	return nil
}
