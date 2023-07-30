package bybit

import (
	"context"
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
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fill"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

const (
	bybitWSBaseURL      = "wss://stream.bybit.com/"
	wsSpotPublicTopicV2 = "v5/public/spot"
	wsSpotPrivate       = "spot/ws"
	bybitWebsocketTimer = 20 * time.Second

	wsOrderbook = "depth"
	wsTicker    = "bookTicker"
	wsTrades    = "trade"
	wsKlines    = "kline"

	// Public v5 channels
	chanOrderbook           = "orderbook"
	chanPublicTrade         = "publicTrade"
	chanPublicTicker        = "tickers"
	chanKline               = "kline"
	chanLiquidation         = "liquidation"
	chanLeverageTokenKline  = "kline_lt"
	chanLeverageTokenTicker = "tickers_lt"
	chanLeverageTokenNav    = "lt"

	// Private v5 channels
	chanPositions = "position"
	chanExecution = "execution"
	chanOrder     = "order"
	chanWallet    = "wallet"
	chanGreeks    = "greeks"
	chanDCP       = "dcp"

	wsAccountInfo    = "outboundAccountInfo"
	wsOrderExecution = "executionReport"
	wsTickerInfo     = "ticketInfo"

	sub    = "sub"    // event for subscribe
	cancel = "cancel" // event for unsubscribe

	spotPublic    = "wss://stream.bybit.com/v5/public/spot"
	linearPublic  = "wss://stream.bybit.com/v5/public/linear"  // USDT, USDC perpetual & USDC Futures
	inversePublic = "wss://stream.bybit.com/v5/public/inverse" // Inverse contract
	optionPublic  = "wss://stream.bybit.com/v5/public/option"  // USDC Option

	// Testnet:
	spotTestnet = "wss://stream-testnet.bybit.com/v5/public/spot"
	linearTest  = "wss://stream-testnet.bybit.com/v5/public/linear"  // USDT and USDC perpetual
	inverseTest = "wss://stream-testnet.bybit.com/v5/public/inverse" // Inverse contract
	optionsTest = "wss://stream-testnet.bybit.com/v5/public/option"  // USDC Option

	// Main-net private
	websocketPrivate = "wss://stream.bybit.com/v5/private"

	// Test-net private
	websocketPrivateTest = "wss://stream-testnet.bybit.com/v5/private"
)

// WsConnect connects to a websocket feed
func (by *Bybit) WsConnect() error {
	if !by.Websocket.IsEnabled() || !by.IsEnabled() || !by.IsAssetWebsocketSupported(asset.Spot) {
		return errWebsocketNotEnabled
	}
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
	go by.wsReadData(asset.Spot, by.Websocket.Conn)
	if by.IsWebsocketAuthenticationSupported() {
		err = by.WsAuth(context.TODO())
		if err != nil {
			by.Websocket.DataHandler <- err
			by.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
	}
	return nil
}

// WsAuth sends an authentication message to receive auth data
func (by *Bybit) WsAuth(ctx context.Context) error {
	var dialer websocket.Dialer
	err := by.Websocket.AuthConn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}

	by.Websocket.AuthConn.SetupPingHandler(stream.PingHandler{
		MessageType: websocket.TextMessage,
		Message:     []byte(`{"op":"ping"}`),
		Delay:       bybitWebsocketTimer,
	})

	by.Websocket.Wg.Add(1)
	go by.wsReadData(asset.Spot, by.Websocket.AuthConn)
	creds, err := by.GetCredentials(ctx)
	if err != nil {
		return err
	}
	intNonce := (time.Now().Unix() + 1) * 1000
	strNonce := strconv.FormatInt(intNonce, 10)
	hmac, err := crypto.GetHMAC(
		crypto.HashSHA256,
		[]byte("GET/realtime"+strNonce),
		[]byte(creds.Secret),
	)
	if err != nil {
		return err
	}
	sign := crypto.HexEncodeToString(hmac)
	req := Authenticate{
		RequestID: strconv.FormatInt(by.Websocket.AuthConn.GenerateMessageID(false), 10),
		Operation: "auth",
		Args:      []interface{}{creds.Key, intNonce, sign},
	}
	resp, err := by.Websocket.AuthConn.SendMessageReturnResponse(req.RequestID, req)
	if err != nil {
		return err
	}
	var response SubscriptionResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return err
	}
	if !response.Success {
		return fmt.Errorf("%s with request ID %s msg: %s", response.Operation, response.RequestID, response.RetMsg)
	}
	return nil
}

// Subscribe sends a websocket message to receive data from the channel
func (by *Bybit) Subscribe(channelsToSubscribe []stream.ChannelSubscription) error {
	payloads, err := by.handleSubscriptions(asset.Spot, "subscribe", channelsToSubscribe)
	if err != nil {
		return err
	}
	for a := range payloads {
		var response []byte
		if payloads[a].auth {
			response, err = by.Websocket.AuthConn.SendMessageReturnResponse(payloads[a].RequestID, payloads[a])
			if err != nil {
				return err
			}
		} else {
			response, err = by.Websocket.Conn.SendMessageReturnResponse(payloads[a].RequestID, payloads[a])
			if err != nil {
				return err
			}
		}
		var resp SubscriptionResponse
		err = json.Unmarshal(response, &resp)
		if err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf("%s with request ID %s msg: %s", resp.Operation, resp.RequestID, resp.RetMsg)
		}
	}
	return nil
}

func (by *Bybit) handleSubscriptions(assetType asset.Item, operation string, channelsToSubscribe []stream.ChannelSubscription) ([]SubscriptionArgument, error) {
	var args []SubscriptionArgument
	arg := SubscriptionArgument{
		Operation: operation,
		RequestID: by.Websocket.Conn.GenerateMessageID(false),
		Arguments: []string{},
	}
	authArg := SubscriptionArgument{
		auth:      true,
		Operation: operation,
		RequestID: by.Websocket.Conn.GenerateMessageID(false),
		Arguments: []string{},
	}

	var selectedChannels, positions, execution, order, wallet, greeks, dCP = 0, 1, 2, 3, 4, 5, 6
	chanMap := map[string]int{
		chanPositions: positions,
		chanExecution: execution,
		chanOrder:     order,
		chanWallet:    wallet,
		chanGreeks:    greeks,
		chanDCP:       dCP}

	pairFormat, err := by.GetPairFormat(assetType, true)
	if err != nil {
		return nil, err
	}
	for i := range channelsToSubscribe {
		switch channelsToSubscribe[i].Channel {
		case chanOrderbook:
			arg.Arguments = append(arg.Arguments, fmt.Sprintf("%s.%d.%s", channelsToSubscribe[i].Channel, 50, channelsToSubscribe[i].Currency.Format(pairFormat).String()))
		case chanPublicTrade, chanPublicTicker, chanLiquidation, chanLeverageTokenTicker, chanLeverageTokenNav:
			arg.Arguments = append(arg.Arguments, channelsToSubscribe[i].Channel+"."+channelsToSubscribe[i].Currency.Format(pairFormat).String())
		case chanKline, chanLeverageTokenKline:
			interval, err := intervalToString(kline.FiveMin)
			if err != nil {
				return nil, err
			}
			arg.Arguments = append(arg.Arguments, channelsToSubscribe[i].Channel+"."+interval+"."+channelsToSubscribe[i].Currency.Format(pairFormat).String())
		case chanPositions, chanExecution, chanOrder, chanWallet, chanGreeks, chanDCP:
			if chanMap[channelsToSubscribe[i].Channel]|selectedChannels > 0 {
				continue
			}
			authArg.Arguments = append(authArg.Arguments, channelsToSubscribe[i].Channel)
			selectedChannels |= chanMap[channelsToSubscribe[i].Channel]
		}
		if assetType == asset.Spot && len(arg.Arguments) >= 10 ||
			assetType == asset.Options && len(arg.Arguments) >= 2000 {
			args = append(args, arg)
			arg = SubscriptionArgument{
				Operation: operation,
				RequestID: by.Websocket.Conn.GenerateMessageID(false),
				Arguments: []string{},
			}
		}
	}
	if len(arg.Arguments) != 0 {
		args = append(args, arg)
	}
	if len(authArg.Arguments) != 0 {
		args = append(args, authArg)
	}
	return args, nil
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (by *Bybit) Unsubscribe(channelsToUnsubscribe []stream.ChannelSubscription) error {
	payloads, err := by.handleSubscriptions(asset.Spot, "unsubscribe", channelsToUnsubscribe)
	if err != nil {
		return err
	}
	for a := range payloads {
		var response []byte
		if payloads[a].auth {
			response, err = by.Websocket.AuthConn.SendMessageReturnResponse(payloads[a].RequestID, payloads[a])
			if err != nil {
				return err
			}
		} else {
			response, err = by.Websocket.Conn.SendMessageReturnResponse(payloads[a].RequestID, payloads[a])
			if err != nil {
				return err
			}
		}
		var resp SubscriptionResponse
		err = json.Unmarshal(response, &resp)
		if err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf("%s with request ID %s msg: %s", resp.Operation, resp.RequestID, resp.RetMsg)
		}
	}
	return nil
}

// GenerateDefaultSubscriptions generates default subscription
func (by *Bybit) GenerateDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	var subscriptions []stream.ChannelSubscription
	var channels = []string{chanOrderbook, chanPublicTrade, chanPublicTicker}
	if by.IsWebsocketAuthenticationSupported() {
		channels = append(channels, []string{
			chanPositions,
			chanExecution,
			chanOrder,
			chanWallet,
		}...)
	}
	pairs, err := by.GetEnabledPairs(asset.Spot)
	if err != nil {
		return nil, err
	}
	for z := range pairs {
		for x := range channels {
			subscriptions = append(subscriptions,
				stream.ChannelSubscription{
					Channel:  channels[x],
					Currency: pairs[z],
					Asset:    asset.Spot,
				})
		}
	}
	return subscriptions, nil
}

func stringToOrderStatus(status string) (order.Status, error) {
	switch status {
	case "NEW":
		return order.New, nil
	case "CANCELED":
		return order.Cancelled, nil
	case "REJECTED":
		return order.Rejected, nil
	case "TRADE":
		return order.PartiallyFilled, nil
	case "EXPIRED":
		return order.Expired, nil
	default:
		return order.UnknownStatus, errors.New(status + " not recognised as order status")
	}
}

// wsReadData receives and passes on websocket messages for processing
func (by *Bybit) wsReadData(assetType asset.Item, ws stream.Connection) {
	defer by.Websocket.Wg.Done()
	for {
		select {
		case <-by.Websocket.ShutdownC:
			return
		default:
			resp := ws.ReadMessage()
			if resp.Raw == nil {
				return
			}
			err := by.wsHandleData(assetType, resp.Raw)
			if err != nil {
				by.Websocket.DataHandler <- err
			}
		}
	}
}

func (by *Bybit) wsHandleData(assetType asset.Item, respRaw []byte) error {
	var result WebsocketResponse
	err := json.Unmarshal(respRaw, &result)
	if err != nil {
		return err
	}
	if result.Topic == "" {
		var subscriptionResult SubscriptionResponse
		err := json.Unmarshal(respRaw, &subscriptionResult)
		if err != nil {
			return err
		}
		if !by.Websocket.Match.IncomingWithData(subscriptionResult.RequestID, respRaw) {
			return fmt.Errorf("could not match subscription with id %s data %s", subscriptionResult.RequestID, respRaw)
		}
		return nil
	}
	topicSplit := strings.Split(result.Topic, ".")
	if len(topicSplit) == 0 {
		return errInvalidPushData
	}
	switch topicSplit[0] {
	case chanOrderbook:
		return by.wsProcessOrderbook(assetType, &result)
	case chanPublicTrade:
		return by.wsProcessPublicTrade(assetType, &result)
	case chanPublicTicker:
		return by.wsProcessPublicTicker(assetType, &result)
	case chanKline:
		return by.wsProcessKline(assetType, &result, topicSplit)
	case chanLiquidation:
		return by.wsProcessLiquidation(assetType, &result)
	case chanLeverageTokenKline:
		return by.wsProcessLeverageTokenKline(assetType, &result, topicSplit)
	case chanLeverageTokenTicker:
		return by.wsProcessLeverageTokenTicker(assetType, &result)
	case chanLeverageTokenNav:
		return by.wsLeverageTokenNav(assetType, &result)
	case chanPositions:
		return by.wsProcessPosition(asset.Spot, &result)
	case chanExecution:
		return by.wsProcessExecution(asset.Spot, &result)
	case chanOrder:
		return by.wsProcessOrder(asset.Spot, &result)
	case chanWallet:
		return by.wsProcessWalletPushData(asset.Spot, respRaw)
	case chanGreeks:
		return by.wsProcessGreeks(respRaw)
	case chanDCP:
		// TODO:
		return nil
	}
	return fmt.Errorf("unhandled stream data %s", string(respRaw))
}

func (by *Bybit) wsProcessGreeks(resp []byte) error {
	var result GreeksResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return err
	}
	by.Websocket.DataHandler <- &result
	return nil
}

func (by *Bybit) wsProcessWalletPushData(assetType asset.Item, resp []byte) error {
	var result WebsocketWallet
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return err
	}
	accounts := []account.Change{}
	for x := range result.Data {
		for y := range result.Data[x].Coin {
			accounts = append(accounts, account.Change{
				Exchange: by.Name,
				Currency: currency.NewCode(result.Data[x].Coin[y].Coin),
				Asset:    assetType,
				Amount:   result.Data[x].Coin[y].WalletBalance.Float64(),
			})
		}
	}
	by.Websocket.DataHandler <- accounts
	return nil
}

// wsProcessOrder the order stream to see changes to your orders in real-time.
func (by *Bybit) wsProcessOrder(assetType asset.Item, resp *WebsocketResponse) error {
	var result WsOrders
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	execution := make([]order.Detail, len(result))
	for x := range result {
		cp, err := currency.NewPairFromString(result[x].Symbol)
		if err != nil {
			return err
		}
		orderType, err := order.StringToOrderType(result[x].OrderType)
		if err != nil {
			return err
		}
		side, err := order.StringToOrderSide(result[x].Side)
		if err != nil {
			return err
		}
		execution[x] = order.Detail{
			Amount:         result[x].Qty.Float64(),
			Exchange:       by.Name,
			OrderID:        result[x].OrderID,
			ClientOrderID:  result[x].OrderLinkID,
			Side:           side,
			Type:           orderType,
			Pair:           cp,
			Cost:           result[x].CumExecQty.Float64() * result[x].AvgPrice.Float64(),
			AssetType:      assetType,
			Status:         getOrderStatus(result[x].OrderStatus),
			Price:          result[x].Price.Float64(),
			ExecutedAmount: result[x].CumExecQty.Float64(),
			Date:           result[x].CreatedTime.Time(),
			LastUpdated:    result[x].UpdatedTime.Time(),
		}
	}
	by.Websocket.DataHandler <- execution
	return nil
}

func (by *Bybit) wsProcessExecution(assetType asset.Item, resp *WebsocketResponse) error {
	var result WsExecutions
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	executions := make([]fill.Data, len(result))
	for x := range result {
		cp, err := currency.NewPairFromString(result[x].Symbol)
		if err != nil {
			return err
		}
		side, err := order.StringToOrderSide(result[x].Side)
		if err != nil {
			return err
		}
		executions[x] = fill.Data{
			ID:            result[x].ExecID,
			Timestamp:     result[x].ExecTime.Time(),
			Exchange:      by.Name,
			AssetType:     assetType,
			CurrencyPair:  cp,
			Side:          side,
			OrderID:       result[x].OrderID,
			ClientOrderID: result[x].OrderLinkID,
			Price:         result[x].ExecPrice.Float64(),
			Amount:        result[x].ExecQty.Float64(),
		}
	}
	by.Websocket.DataHandler <- executions
	return nil
}

func (by *Bybit) wsProcessPosition(assetType asset.Item, resp *WebsocketResponse) error {
	var result WsPositions
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	by.Websocket.DataHandler <- result
	return nil
}

func (by *Bybit) wsLeverageTokenNav(assetType asset.Item, resp *WebsocketResponse) error {
	var result LTNav
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	by.Websocket.DataHandler <- result
	return nil
}

func (by *Bybit) wsProcessLeverageTokenTicker(assetType asset.Item, resp *WebsocketResponse) error {
	var result LTTicker
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(result.Symbol)
	if err != nil {
		return err
	}
	by.Websocket.DataHandler <- ticker.Price{
		Last:         result.LastPrice.Float64(),
		High:         result.HighPrice24H.Float64(),
		Low:          result.LowPrice24H.Float64(),
		Pair:         cp,
		ExchangeName: by.Name,
		AssetType:    assetType,
		LastUpdated:  resp.Timestamp.Time(),
	}
	return nil
}

func (by *Bybit) wsProcessLeverageTokenKline(assetType asset.Item, resp *WebsocketResponse, topicSplit []string) error {
	var result LTKlines
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(topicSplit[2])
	if err != nil {
		return err
	}
	ltKline := make([]stream.KlineData, len(result))
	for x := range result {
		interval, err := stringToInterval(result[x].Interval)
		if err != nil {
			return err
		}
		ltKline[x] = stream.KlineData{
			Timestamp:  result[x].Timestamp.Time(),
			Pair:       cp,
			AssetType:  assetType,
			Exchange:   by.Name,
			StartTime:  result[x].Start.Time(),
			CloseTime:  result[x].End.Time(),
			Interval:   interval.String(),
			OpenPrice:  result[x].Open.Float64(),
			ClosePrice: result[x].Close.Float64(),
			HighPrice:  result[x].High.Float64(),
			LowPrice:   result[x].Low.Float64(),
		}
	}
	by.Websocket.DataHandler <- result
	return nil
}

func (by *Bybit) wsProcessLiquidation(assetType asset.Item, resp *WebsocketResponse) error {
	var result WebsocketLiquidiation
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	by.Websocket.DataHandler <- result
	return nil
}

func (by *Bybit) wsProcessKline(assetType asset.Item, resp *WebsocketResponse, topicSplit []string) error {
	var result WsKlines
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(topicSplit[2])
	if err != nil {
		return err
	}
	spotCandlesticks := make([]stream.KlineData, len(result))
	for x := range result {
		interval, err := stringToInterval(result[x].Interval)
		if err != nil {
			return err
		}
		spotCandlesticks[x] = stream.KlineData{
			Timestamp:  result[x].Timestamp.Time(),
			Pair:       cp,
			AssetType:  assetType,
			Exchange:   by.Name,
			StartTime:  result[x].Start.Time(),
			CloseTime:  result[x].End.Time(),
			Interval:   interval.String(),
			OpenPrice:  result[x].Open.Float64(),
			ClosePrice: result[x].Close.Float64(),
			HighPrice:  result[x].High.Float64(),
			LowPrice:   result[x].Low.Float64(),
			Volume:     result[x].Volume.Float64(),
		}
	}
	by.Websocket.DataHandler <- spotCandlesticks
	return nil
}

func (by *Bybit) wsProcessPublicTicker(assetType asset.Item, resp *WebsocketResponse) error {
	switch assetType {
	case asset.Spot:
		var result WsSpotTicker
		err := json.Unmarshal(resp.Data, &result)
		if err != nil {
			return err
		}
		cp, err := currency.NewPairFromString(result.Symbol)
		if err != nil {
			return err
		}
		by.Websocket.DataHandler <- ticker.Price{
			Last:         result.LastPrice.Float64(),
			High:         result.HighPrice24H.Float64(),
			Low:          result.LowPrice24H.Float64(),
			Volume:       result.Volume24H.Float64(),
			Pair:         cp,
			ExchangeName: by.Name,
			AssetType:    assetType,
			LastUpdated:  resp.Timestamp.Time(),
		}
		return nil
	case asset.Linear:
		var result WsLinearTicker
		err := json.Unmarshal(resp.Data, &result)
		if err != nil {
			return err
		}
		cp, err := currency.NewPairFromString(result.Symbol)
		if err != nil {
			return err
		}
		by.Websocket.DataHandler <- ticker.Price{
			Last:         result.LastPrice.Float64(),
			High:         result.HighPrice24H.Float64(),
			Low:          result.LastPrice.Float64(),
			Bid:          result.Bid1Price.Float64(),
			BidSize:      result.Bid1Size.Float64(),
			Ask:          result.Ask1Size.Float64(),
			AskSize:      result.Ask1Size.Float64(),
			Volume:       result.Volume24H.Float64(),
			Pair:         cp,
			ExchangeName: by.Name,
			AssetType:    assetType,
		}
		return nil
	case asset.Options:
		var result WsOptionTicker
		err := json.Unmarshal(resp.Data, &result)
		if err != nil {
			return err
		}
		cp, err := currency.NewPairFromString(result.Symbol)
		if err != nil {
			return err
		}
		by.Websocket.DataHandler <- ticker.Price{
			Last:         result.LastPrice.Float64(),
			High:         result.HighPrice24H.Float64(),
			Low:          result.LastPrice.Float64(),
			Bid:          result.BidPrice.Float64(),
			BidSize:      result.BidSize.Float64(),
			Ask:          result.AskSize.Float64(),
			AskSize:      result.AskSize.Float64(),
			Volume:       result.Volume24H.Float64(),
			Pair:         cp,
			ExchangeName: by.Name,
			AssetType:    assetType,
		}
		return nil
	default:
		return fmt.Errorf("ticker data unsupported for asset type %v", assetType)
	}
}

func (by *Bybit) wsProcessPublicTrade(assetType asset.Item, resp *WebsocketResponse) error {
	var result WebsocketPublicTrades
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	tradeDatas := make([]trade.Data, len(result))
	for x := range result {
		cp, err := currency.NewPairFromString(result[x].Symbol)
		if err != nil {
			return err
		}
		side, err := order.StringToOrderSide(result[x].Side)
		if err != nil {
			return err
		}
		tradeDatas[x] = trade.Data{
			Timestamp:    result[x].OrderFillTimestamp.Time(),
			CurrencyPair: cp,
			AssetType:    assetType,
			Exchange:     by.Name,
			Price:        result[x].Price.Float64(),
			Amount:       result[x].Size.Float64(),
			Side:         side,
			TID:          result[x].TradeID,
		}
	}
	return trade.AddTradesToBuffer(by.Name, tradeDatas...)
}

func (by *Bybit) wsProcessOrderbook(assetType asset.Item, resp *WebsocketResponse) error {
	var result WsOrderbookDetail
	err := json.Unmarshal(resp.Data, &result)
	if err != nil {
		return err
	}
	cp, err := currency.NewPairFromString(result.Symbol)
	if err != nil {
		return err
	}
	asks := make([]orderbook.Item, len(result.Asks))
	for i := range result.Asks {
		var price float64
		price, err = strconv.ParseFloat(result.Asks[i][0], 64)
		if err != nil {
			return err
		}
		size, err := strconv.ParseFloat(result.Asks[i][1], 64)
		if err != nil {
			return err
		}
		asks[i] = orderbook.Item{Price: price, Amount: size}
	}
	bids := make([]orderbook.Item, len(result.Bids))
	for i := range result.Bids {
		var price float64
		price, err = strconv.ParseFloat(result.Bids[i][0], 64)
		if err != nil {
			return err
		}
		size, err := strconv.ParseFloat(result.Bids[i][1], 64)
		if err != nil {
			return err
		}
		bids[i] = orderbook.Item{Price: price, Amount: size}
	}
	if len(asks) == 0 && len(bids) == 0 {
		return nil
	}
	if resp.Type == "snapshot" {
		err = by.Websocket.Orderbook.LoadSnapshot(&orderbook.Base{
			Pair:         cp,
			Exchange:     by.Name,
			Asset:        assetType,
			LastUpdated:  result.UpdateTime.Time(),
			LastUpdateID: result.Sequence,
			Asks:         asks,
			Bids:         bids,
		})
		if err != nil {
			return err
		}
	} else {
		err = by.Websocket.Orderbook.Update(&orderbook.Update{
			Pair:       cp,
			Asks:       asks,
			Bids:       bids,
			Asset:      assetType,
			UpdateID:   result.Sequence,
			UpdateTime: result.UpdateTime.Time(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
