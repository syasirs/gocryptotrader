package bot

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency"
)

// WebsocketEvent is the struct used for websocket events
type WebsocketEvent struct {
	Exchange  string `json:"exchange,omitempty"`
	AssetType string `json:"assetType,omitempty"`
	Event     string
	Data      interface{}
}

// WebsocketAuth is the struct used for a websocket auth request
type WebsocketAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// WebsocketEventResponse is the struct used for websocket event responses
type WebsocketEventResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

// WebsocketOrderbookTickerRequest is a struct used for ticker and orderbook
// requests
type WebsocketOrderbookTickerRequest struct {
	Exchange  string `json:"exchangeName"`
	Currency  string `json:"currency"`
	AssetType string `json:"assetType"`
}

// Const vars for websocket
const (
	WebsocketResponseSuccess = "OK"
)

// WebsocketClient stores information related to the websocket client
type WebsocketClient struct {
	ID            int
	Conn          *websocket.Conn
	LastRecv      time.Time
	Authenticated bool
}

// WebsocketClientHub stores an array of websocket clients
var WebsocketClientHub []WebsocketClient

// WebsocketClientHandler upgrades the HTTP connection to a websocket
// compatible one
func (b *Bot) WebsocketClientHandler(w http.ResponseWriter, r *http.Request) {
	connectionLimit := b.Config.Webserver.WebsocketConnectionLimit
	numClients := len(WebsocketClientHub)

	if numClients >= connectionLimit {
		log.Printf("Websocket client rejected due to websocket client limit reached. Number of clients %d. Limit %d.",
			numClients, connectionLimit)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	upgrader := websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}

	// Allow insecure origin if the Origin request header is present and not
	// equal to the Host request header. Default to false
	if b.Config.Webserver.WebsocketAllowInsecureOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	newClient := WebsocketClient{
		ID: len(WebsocketClientHub),
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	newClient.Conn = conn
	WebsocketClientHub = append(WebsocketClientHub, newClient)
	numClients++
	log.Printf("New websocket client connected. Connected clients: %d. Limit %d.",
		numClients, connectionLimit)
}

// DisconnectWebsocketClient disconnects a websocket client
func DisconnectWebsocketClient(id int, err error) {
	for i := range WebsocketClientHub {
		if WebsocketClientHub[i].ID == id {
			WebsocketClientHub[i].Conn.Close()
			WebsocketClientHub = append(WebsocketClientHub[:i], WebsocketClientHub[i+1:]...)
			log.Printf("Disconnected Websocket client, error: %s", err)
			return
		}
	}
}

// SendWebsocketMessage sends a websocket message to a specific client
func SendWebsocketMessage(id int, data interface{}) error {
	for _, x := range WebsocketClientHub {
		if x.ID == id {
			return x.Conn.WriteJSON(data)
		}
	}
	return nil
}

// BroadcastWebsocketMessage broadcasts a websocket event message to all
// websocket clients
func BroadcastWebsocketMessage(evt WebsocketEvent) error {
	for _, x := range WebsocketClientHub {
		x.Conn.WriteJSON(evt)
	}
	return nil
}

// WsCommandHandler is a function for websocket commands
type WsCommandHandler func(wsClient *websocket.Conn, data interface{}) error

// GetWSHandlers instantiates a hash table of websocket commands by its string
// name
func (b *Bot) GetWSHandlers() map[string]WsCommandHandler {
	var wsHandlers = map[string]WsCommandHandler{
		"getconfig":        b.WsGetConfig,
		"saveconfig":       b.WsSaveConfig,
		"getaccountinfo":   b.WsGetAccountInfo,
		"gettickers":       b.WsGetTickers,
		"getticker":        b.WsGetTicker,
		"getorderbooks":    b.WsGetOrderbooks,
		"getorderbook":     b.WsGetOrderbook,
		"getexchangerates": b.WsGetExchangeRates,
		"getportfolio":     b.WsGetPortfolio,
	}

	return wsHandlers
}

// WsGetConfig writes configuration request
func (b *Bot) WsGetConfig(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetConfig",
		Data:  b.Config,
	}
	return wsClient.WriteJSON(wsResp)
}

// WsSaveConfig saves configuration
func (b *Bot) WsSaveConfig(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "SaveConfig",
	}
	var cfg config.Config
	err := common.JSONDecode(data.([]byte), &cfg)
	if err != nil {
		wsResp.Error = err.Error()
		err = wsClient.WriteJSON(wsResp)
		if err != nil {
			return err
		}
	}

	err = b.Config.UpdateConfig(b.ConfigFile, cfg)
	if err != nil {
		wsResp.Error = err.Error()
		err = wsClient.WriteJSON(wsResp)
		if err != nil {
			return err
		}
	}

	b.SetupBotExchanges()
	wsResp.Data = WebsocketResponseSuccess
	return wsClient.WriteJSON(wsResp)
}

// WsGetAccountInfo sends account information
func (b *Bot) WsGetAccountInfo(wsClient *websocket.Conn, data interface{}) error {
	accountInfo := b.GetAllEnabledExchangeAccountInfo()
	wsResp := WebsocketEventResponse{
		Event: "GetAccountInfo",
		Data:  accountInfo,
	}
	return wsClient.WriteJSON(wsResp)
}

// WsGetTickers sends current tickers
func (b *Bot) WsGetTickers(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetTickers",
	}
	wsResp.Data = b.GetAllActiveTickers()
	return wsClient.WriteJSON(wsResp)
}

// WsGetTicker sends ticker
func (b *Bot) WsGetTicker(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetTicker",
	}
	var tickerReq WebsocketOrderbookTickerRequest
	err := common.JSONDecode(data.([]byte), &tickerReq)
	if err != nil {
		wsResp.Error = err.Error()
		wsClient.WriteJSON(wsResp)
		return err
	}

	result, err := b.GetSpecificTicker(tickerReq.Currency,
		tickerReq.Exchange, tickerReq.AssetType)

	if err != nil {
		wsResp.Error = err.Error()
		wsClient.WriteJSON(wsResp)
		return err
	}
	wsResp.Data = result
	return wsClient.WriteJSON(wsResp)
}

// WsGetOrderbooks sends orderbook
func (b *Bot) WsGetOrderbooks(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetOrderbooks",
	}
	wsResp.Data = b.GetAllActiveOrderbooks()
	return wsClient.WriteJSON(wsResp)
}

// WsGetOrderbook sends individual orderbook
func (b *Bot) WsGetOrderbook(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetOrderbook",
	}
	var orderbookReq WebsocketOrderbookTickerRequest
	err := common.JSONDecode(data.([]byte), &orderbookReq)
	if err != nil {
		wsResp.Error = err.Error()
		wsClient.WriteJSON(wsResp)
		return err
	}

	result, err := b.GetSpecificOrderbook(orderbookReq.Currency,
		orderbookReq.Exchange, orderbookReq.AssetType)

	if err != nil {
		wsResp.Error = err.Error()
		wsClient.WriteJSON(wsResp)
		return err
	}
	wsResp.Data = result
	return wsClient.WriteJSON(wsResp)
}

// WsGetExchangeRates sends exchange rates
func (b *Bot) WsGetExchangeRates(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetExchangeRates",
	}
	if currency.YahooEnabled {
		wsResp.Data = currency.CurrencyStore
	} else {
		wsResp.Data = currency.CurrencyStoreFixer
	}
	return wsClient.WriteJSON(wsResp)
}

// WsGetPortfolio returns portfolio summary
func (b *Bot) WsGetPortfolio(wsClient *websocket.Conn, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetPortfolio",
	}
	wsResp.Data = b.Portfolio.GetPortfolioSummary()
	return wsClient.WriteJSON(wsResp)
}

// WebsocketHandler Handles websocket client requests
func (b *Bot) WebsocketHandler() {
	wsHandlers := b.GetWSHandlers()
	for {
		for x := range WebsocketClientHub {
			var evt WebsocketEvent
			err := WebsocketClientHub[x].Conn.ReadJSON(&evt)
			if err != nil {
				DisconnectWebsocketClient(x, err)
				continue
			}

			if evt.Event == "" {
				DisconnectWebsocketClient(x, errors.New("Websocket client sent data we did not understand"))
				continue
			}

			dataJSON, err := common.JSONEncode(evt.Data)
			if err != nil {
				log.Println(err)
				continue
			}

			req := common.StringToLower(evt.Event)
			log.Printf("Websocket req: %s", req)

			if !WebsocketClientHub[x].Authenticated && evt.Event != "auth" {
				wsResp := WebsocketEventResponse{
					Event: "auth",
					Error: "you must authenticate first",
				}
				SendWebsocketMessage(x, wsResp)
				DisconnectWebsocketClient(x, errors.New("Websocket client did not auth"))
				continue
			} else if !WebsocketClientHub[x].Authenticated && evt.Event == "auth" {
				var auth WebsocketAuth
				err = common.JSONDecode(dataJSON, &auth)
				if err != nil {
					log.Println(err)
					continue
				}
				hashPW := common.HexEncodeToString(common.GetSHA256([]byte(b.Config.Webserver.AdminPassword)))
				if auth.Username == b.Config.Webserver.AdminUsername && auth.Password == hashPW {
					WebsocketClientHub[x].Authenticated = true
					wsResp := WebsocketEventResponse{
						Event: "auth",
						Data:  WebsocketResponseSuccess,
					}
					SendWebsocketMessage(x, wsResp)
					log.Println("Websocket client authenticated successfully")
					continue
				} else {
					wsResp := WebsocketEventResponse{
						Event: "auth",
						Error: "invalid username/password",
					}
					SendWebsocketMessage(x, wsResp)
					DisconnectWebsocketClient(x, errors.New("Websocket client sent wrong username/password"))
					continue
				}
			}
			result, ok := wsHandlers[req]
			if !ok {
				log.Printf("Websocket unsupported event")
				continue
			}

			err = result(WebsocketClientHub[x].Conn, dataJSON)
			if err != nil {
				log.Printf("Websocket request %s failed. Error %s", evt.Event, err)
				continue
			}
		}
		time.Sleep(time.Millisecond)
	}
}
