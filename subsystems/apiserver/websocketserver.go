package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/database/repository/exchange"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// StartWebsocketServer starts a Websocket handler
func (m *Manager) StartWebsocketServer() error {
	if !atomic.CompareAndSwapInt32(&m.websocketStarted, 0, 1) {
		return fmt.Errorf("websocket server %w", errAlreadyRunning)
	}
	if !m.remoteConfig.WebsocketRPC.Enabled {
		atomic.StoreInt32(&m.websocketStarted, 0)
		return fmt.Errorf("websocket %w", errServerDisabled)
	}
	atomic.StoreInt32(&m.started, 1)
	log.Debugf(log.CommunicationMgr,
		"Websocket RPC support enabled. Listen URL: ws://%s:%d/ws\n",
		common.ExtractHost(m.websocketListenAddress), common.ExtractPort(m.websocketListenAddress))
	m.websocketRouter = m.newRouter(false)
	if m.websocketListenAddress == "localhost:-1" {
		atomic.StoreInt32(&m.websocketStarted, 0)
		return errInvalidListenAddress
	}

	m.websocketHttpServer = &http.Server{
		Addr:    m.websocketListenAddress,
		Handler: m.websocketRouter,
	}

	go func() {
		err := m.websocketHttpServer.ListenAndServe()
		if err != nil {
			atomic.StoreInt32(&m.websocketStarted, 0)
			log.Error(log.GRPCSys, err)
		}
	}()
	return nil
}

// NewWebsocketHub Creates a new websocket hub
func NewWebsocketHub() *WebsocketHub {
	return &WebsocketHub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *WebsocketClient),
		Unregister: make(chan *WebsocketClient),
		Clients:    make(map[*WebsocketClient]bool),
	}
}

func (h *WebsocketHub) run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				log.Debugln(log.WebsocketMgr, "websocket: disconnected client")
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					log.Debugln(log.WebsocketMgr, "websocket: disconnected client")
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// SendWebsocketMessage sends a websocket event to the client
func (c *WebsocketClient) SendWebsocketMessage(evt interface{}) error {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Errorf(log.WebsocketMgr, "websocket: failed to send message: %s\n", err)
		return err
	}

	c.Send <- data
	return nil
}

func (c *WebsocketClient) read() {
	defer func() {
		c.Hub.Unregister <- c
		conErr := c.Conn.Close()
		if conErr != nil {
			log.Error(log.WebsocketMgr, conErr)
		}
	}()

	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf(log.WebsocketMgr, "websocket: client disconnected, err: %s\n", err)
			}
			break
		}

		if msgType == websocket.TextMessage {
			var evt WebsocketEvent
			err := json.Unmarshal(message, &evt)
			if err != nil {
				log.Errorf(log.WebsocketMgr, "websocket: failed to decode JSON sent from client %s\n", err)
				continue
			}

			if evt.Event == "" {
				log.Warnln(log.WebsocketMgr, "websocket: client sent a blank event, disconnecting")
				continue
			}

			dataJSON, err := json.Marshal(evt.Data)
			if err != nil {
				log.Errorln(log.WebsocketMgr, "websocket: client sent data we couldn't JSON decode")
				break
			}

			req := strings.ToLower(evt.Event)
			log.Debugf(log.WebsocketMgr, "websocket: request received: %s\n", req)

			result, ok := wsHandlers[req]
			if !ok {
				log.Debugln(log.WebsocketMgr, "websocket: unsupported event")
				continue
			}

			if result.authRequired && !c.Authenticated {
				log.Warnf(log.WebsocketMgr, "Websocket: request %s failed due to unauthenticated request on an authenticated API\n", evt.Event)
				err = c.SendWebsocketMessage(WebsocketEventResponse{Event: evt.Event, Error: "unauthorised request on authenticated API"})
				if err != nil {
					log.Error(log.WebsocketMgr, err)
				}
				continue
			}

			err = result.handler(c, dataJSON)
			if err != nil {
				log.Errorf(log.WebsocketMgr, "websocket: request %s failed. Error %s\n", evt.Event, err)
				continue
			}
		}
	}
}

func (c *WebsocketClient) write() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			log.Error(log.WebsocketMgr, err)
		}
	}()
	for { // nolint // ws client write routine loop
		select {
		case message, ok := <-c.Send:
			if !ok {
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Error(log.WebsocketMgr, err)
				}
				log.Debugln(log.WebsocketMgr, "websocket: hub closed the channel")
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Errorf(log.WebsocketMgr, "websocket: failed to create new io.writeCloser: %s\n", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log.Error(log.WebsocketMgr, err)
			}

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-c.Send)
				if err != nil {
					log.Error(log.WebsocketMgr, err)
				}
			}

			if err := w.Close(); err != nil {
				log.Errorf(log.WebsocketMgr, "websocket: failed to close io.WriteCloser: %s\n", err)
				return
			}
		}
	}
}

// StartWebsocketHandler starts the websocket hub and routine which
// handles clients
func StartWebsocketHandler() {
	if !wsHubStarted {
		wsHubStarted = true
		wsHub = NewWebsocketHub()
		go wsHub.run()
	}
}

// BroadcastWebsocketMessage meow
func BroadcastWebsocketMessage(evt WebsocketEvent) error {
	if !wsHubStarted {
		return ErrWebsocketServiceNotRunning
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	wsHub.Broadcast <- data
	return nil
}

// WebsocketClientHandler upgrades the HTTP connection to a websocket
// compatible one
func (m *Manager) WebsocketClientHandler(w http.ResponseWriter, r *http.Request) {
	if !wsHubStarted {
		StartWebsocketHandler()
	}

	connectionLimit := m.remoteConfig.WebsocketRPC.ConnectionLimit
	numClients := len(wsHub.Clients)

	if numClients >= connectionLimit {
		log.Warnf(log.WebsocketMgr,
			"websocket: client rejected due to websocket client limit reached. Number of clients %d. Limit %d.\n",
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
	if m.remoteConfig.WebsocketRPC.AllowInsecureOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(log.WebsocketMgr, err)
		return
	}

	client := &WebsocketClient{
		Hub:              wsHub,
		Conn:             conn,
		Send:             make(chan []byte, 1024),
		maxAuthFailures:  m.remoteConfig.WebsocketRPC.MaxAuthFailures,
		username:         m.remoteConfig.Username,
		password:         m.remoteConfig.Password,
		configPath:       m.gctConfigPath,
		exchangeManager:  m.exchangeManager,
		bot:              m.bot,
		portfolioManager: m.portfolioManager,
	}

	client.Hub.Register <- client
	log.Debugf(log.WebsocketMgr,
		"websocket: client connected. Connected clients: %d. Limit %d.\n",
		numClients+1, connectionLimit)

	go client.read()
	go client.write()
}

func wsAuth(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "auth",
	}

	var auth WebsocketAuth
	err := json.Unmarshal(data.([]byte), &auth)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}

	hashPW := crypto.HexEncodeToString(crypto.GetSHA256([]byte(client.password)))
	if auth.Username == client.username && auth.Password == hashPW {
		client.Authenticated = true
		wsResp.Data = WebsocketResponseSuccess
		log.Debugln(log.WebsocketMgr,
			"websocket: client authenticated successfully")
		return client.SendWebsocketMessage(wsResp)
	}

	wsResp.Error = "invalid username/password"
	client.authFailures++
	sendErr := client.SendWebsocketMessage(wsResp)
	if sendErr != nil {
		log.Error(log.WebsocketMgr, sendErr)
	}
	if client.authFailures >= client.maxAuthFailures {
		log.Debugf(log.WebsocketMgr,
			"websocket: disconnecting client, maximum auth failures threshold reached (failures: %d limit: %d)\n",
			client.authFailures, client.maxAuthFailures)
		wsHub.Unregister <- client
		return nil
	}

	log.Debugf(log.WebsocketMgr,
		"websocket: client sent wrong username/password (failures: %d limit: %d)\n",
		client.authFailures, client.maxAuthFailures)
	return nil
}

func wsGetConfig(client *WebsocketClient, _ interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetConfig",
		Data:  config.GetConfig(),
	}
	return client.SendWebsocketMessage(wsResp)
}

func wsSaveConfig(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "SaveConfig",
	}
	var respCfg config.Config
	err := json.Unmarshal(data.([]byte), &respCfg)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}

	cfg := config.GetConfig()
	err = cfg.UpdateConfig(client.configPath, &respCfg, false)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}

	err = client.bot.SetupExchanges()
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}
	wsResp.Data = WebsocketResponseSuccess
	return client.SendWebsocketMessage(wsResp)
}

func wsGetAccountInfo(client *WebsocketClient, data interface{}) error {
	accountInfo := getAllActiveAccounts(client.exchangeManager)
	wsResp := WebsocketEventResponse{
		Event: "GetAccountInfo",
		Data:  accountInfo,
	}
	return client.SendWebsocketMessage(wsResp)
}

func wsGetTickers(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetTickers",
	}
	wsResp.Data = getAllActiveTickers(client.exchangeManager)
	return client.SendWebsocketMessage(wsResp)
}

func wsGetTicker(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetTicker",
	}
	var tickerReq WebsocketOrderbookTickerRequest
	err := json.Unmarshal(data.([]byte), &tickerReq)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}

	p, err := currency.NewPairFromString(tickerReq.Currency)
	if err != nil {
		return err
	}

	a, err := asset.New(tickerReq.AssetType)
	if err != nil {
		return err
	}

	exch := client.exchangeManager.GetExchangeByName(tickerReq.Exchange)
	if exch == nil {
		wsResp.Error = exchange.ErrNoExchangeFound.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}
	tick, err := exch.FetchTicker(p, a)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}
	wsResp.Data = tick
	return client.SendWebsocketMessage(wsResp)
}

func wsGetOrderbooks(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetOrderbooks",
	}
	wsResp.Data = getAllActiveOrderbooks(client.exchangeManager)
	return client.SendWebsocketMessage(wsResp)
}

func wsGetOrderbook(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetOrderbook",
	}
	var orderbookReq WebsocketOrderbookTickerRequest
	err := json.Unmarshal(data.([]byte), &orderbookReq)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}

	p, err := currency.NewPairFromString(orderbookReq.Currency)
	if err != nil {
		return err
	}

	a, err := asset.New(orderbookReq.AssetType)
	if err != nil {
		return err
	}

	exch := client.exchangeManager.GetExchangeByName(orderbookReq.Exchange)
	if exch == nil {
		wsResp.Error = exchange.ErrNoExchangeFound.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}
	ob, err := exch.FetchOrderbook(p, a)
	if err != nil {
		wsResp.Error = err.Error()
		sendErr := client.SendWebsocketMessage(wsResp)
		if sendErr != nil {
			log.Error(log.WebsocketMgr, sendErr)
		}
		return err
	}
	wsResp.Data = ob
	return nil
}

func wsGetExchangeRates(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetExchangeRates",
	}

	var err error
	wsResp.Data, err = currency.GetExchangeRates()
	if err != nil {
		return err
	}

	return client.SendWebsocketMessage(wsResp)
}

func wsGetPortfolio(client *WebsocketClient, data interface{}) error {
	wsResp := WebsocketEventResponse{
		Event: "GetPortfolio",
	}

	wsResp.Data = client.portfolioManager.GetPortfolioSummary()
	return client.SendWebsocketMessage(wsResp)
}
