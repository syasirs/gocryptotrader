package coinbene

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// Coinbene is the overarching type across this package
type Coinbene struct {
	exchange.Base
}

const (
	coinbeneAPIURL     = "http://openapi-exchange.coinbene.com/api/exchange/"
	coinbeneAuthPath   = "/api/exchange/v2"
	coinbeneAPIVersion = "v2"

	// Public endpoints
	coinbeneFetchTicker    = "/market/ticker/one"
	coinbeneFetchOrderBook = "/market/orderBook"
	coinbeneGetTrades      = "/market/trades"
	coinbeneGetAllPairs    = "/market/tradePair/list"
	coinbenePairInfo       = "/market/tradePair/one"

	// Authenticated endpoints
	coinbeneGetUserBalance = "/account/list"
	coinbenePlaceOrder     = "/order/place"
	coinbeneOrderInfo      = "/order/info"
	coinbeneRemoveOrder    = "/order/cancel"
	coinbeneOpenOrders     = "/order/openOrders"
)

// SetDefaults sets the basic defaults for Coinbene
func (c *Coinbene) SetDefaults() {
	c.Name = "Coinbene"
	c.Enabled = true
	c.Verbose = false
	c.RESTPollingDelay = 10
	c.RequestCurrencyPairFormat.Delimiter = ""
	c.RequestCurrencyPairFormat.Uppercase = true
	c.ConfigCurrencyPairFormat.Delimiter = "/"
	c.ConfigCurrencyPairFormat.Uppercase = true
	c.AssetTypes = []string{ticker.Spot}
	c.SupportsAutoPairUpdating = false
	c.SupportsRESTTickerBatching = false
	c.Requester = request.New(c.Name,
		request.NewRateLimit(time.Second, 0),
		request.NewRateLimit(time.Second, 0),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	c.APIUrlDefault = coinbeneAPIURL
	c.APIUrl = c.APIUrlDefault
	c.Websocket = wshandler.New()
}

// Setup takes in the supplied exchange configuration details and sets params
func (c *Coinbene) Setup(exch *config.ExchangeConfig) {
	if !exch.Enabled {
		c.SetEnabled(false)
	} else {
		c.Enabled = true
		c.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		c.AuthenticatedWebsocketAPISupport = exch.AuthenticatedWebsocketAPISupport
		c.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		c.SetHTTPClientTimeout(exch.HTTPTimeout)
		c.SetHTTPClientUserAgent(exch.HTTPUserAgent)
		c.RESTPollingDelay = exch.RESTPollingDelay
		c.Verbose = exch.Verbose
		c.Websocket.SetWsStatusAndConnection(exch.Websocket)
		c.BaseCurrencies = exch.BaseCurrencies
		c.AvailablePairs = exch.AvailablePairs
		c.EnabledPairs = exch.EnabledPairs
		err := c.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = c.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = c.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
		err = c.SetAPIURL(exch)
		if err != nil {
			log.Fatal(err)
		}
		err = c.SetClientProxyAddress(exch.ProxyAddress)
		if err != nil {
			log.Fatal(err)
		}

		// If the exchange supports websocket, update the below block
		// err = c.WebsocketSetup(c.WsConnect,
		//	exch.Name,
		//	exch.Websocket,
		//	coinbeneWebsocket,
		//	exch.WebsocketURL)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// c.WebsocketConn = &wshandler.WebsocketConnection{
		// 		ExchangeName:         c.Name,
		// 		URL:                  c.Websocket.GetWebsocketURL(),
		// 		ProxyURL:             c.Websocket.GetProxyAddress(),
		// 		Verbose:              c.Verbose,
		// 		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		// 		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
		// }
	}
}

// FetchTicker gets and stores ticker data for a currency pair
func (c *Coinbene) FetchTicker(symbol string) (TickerResponse, error) {
	var t TickerResponse
	params := url.Values{}
	params.Set("symbol", symbol)
	path := fmt.Sprintf("%s%s%s?%s", c.APIUrl, coinbeneAPIVersion, coinbeneFetchTicker, params.Encode())
	return t, c.SendHTTPRequest(path, &t)
}

// FetchOrderbooks gets and stores orderbook data for given pair
func (c *Coinbene) FetchOrderbooks(symbol, size string) (OrderbookResponse, error) {
	var o OrderbookResponse
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("depth", size)
	path := fmt.Sprintf("%s%s%s?%s", c.APIUrl, coinbeneAPIVersion, coinbeneFetchOrderBook, params.Encode())
	return o, c.SendHTTPRequest(path, &o)
}

// GetTrades gets recent trades from the exchange
func (c *Coinbene) GetTrades(symbol string) (TradeResponse, error) {
	var t TradeResponse
	params := url.Values{}
	params.Set("symbol", symbol)
	path := fmt.Sprintf("%s%s%s?%s", c.APIUrl, coinbeneAPIVersion, coinbeneGetTrades, params.Encode())
	return t, c.SendHTTPRequest(path, &t)
}

// GetPairInfo gets info about a single pair
func (c *Coinbene) GetPairInfo(symbol string) (PairResponse, error) {
	var resp PairResponse
	params := url.Values{}
	params.Set("symbol", symbol)
	path := fmt.Sprintf("%s%s%s?%s", c.APIUrl, coinbeneAPIVersion, coinbenePairInfo, params.Encode())
	return resp, c.SendHTTPRequest(path, &resp)
}

// GetAllPairs gets all pairs on the exchange
func (c *Coinbene) GetAllPairs() (AllPairResponse, error) {
	var a AllPairResponse
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbeneGetAllPairs)
	return a, c.SendHTTPRequest(path, &a)
}

// GetUserBalance gets user balanace info
func (c *Coinbene) GetUserBalance() (UserBalanceResponse, error) {
	var resp UserBalanceResponse
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbeneGetUserBalance)
	return resp, c.SendAuthHTTPRequest(http.MethodGet, path, nil, &resp, coinbeneGetUserBalance)
}

// PlaceOrder creates an order
func (c *Coinbene) PlaceOrder(price, quantity float64, symbol, direction, clientID string) (PlaceOrderResponse, error) {
	var resp PlaceOrderResponse
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbenePlaceOrder)
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("direction", direction)
	params.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	params.Set("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
	params.Set("clientId", clientID)
	return resp, c.SendAuthHTTPRequest(http.MethodPost, path, params, &resp, coinbenePlaceOrder)
}

// FetchOrderInfo gets order info
func (c *Coinbene) FetchOrderInfo(orderID string) (OrderInfoResponse, error) {
	var resp OrderInfoResponse
	params := url.Values{}
	params.Set("orderid", orderID)
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbeneOrderInfo)
	return resp, c.SendAuthHTTPRequest(http.MethodGet, path, params, &resp, coinbeneOrderInfo)
}

// RemoveOrder removes a given order
func (c *Coinbene) RemoveOrder(orderID string) (RemoveOrderResponse, error) {
	var resp RemoveOrderResponse
	params := url.Values{}
	params.Set("orderid", orderID)
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbeneRemoveOrder)
	return resp, c.SendAuthHTTPRequest(http.MethodPost, path, params, &resp, coinbeneRemoveOrder)
}

// FetchOpenOrders finds open orders
func (c *Coinbene) FetchOpenOrders(symbol string) (OpenOrderResponse, error) {
	var resp OpenOrderResponse
	params := url.Values{}
	params.Set("symbol", symbol)
	path := fmt.Sprintf("%s%s%s", c.APIUrl, coinbeneAPIVersion, coinbeneOpenOrders)
	return resp, c.SendAuthHTTPRequest(http.MethodGet, path, params, &resp, coinbeneOpenOrders)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (c *Coinbene) SendHTTPRequest(path string, result interface{}) error {
	return c.SendPayload(http.MethodGet,
		path,
		nil,
		nil,
		&result,
		false,
		false,
		c.Verbose,
		c.HTTPDebugging,
		c.HTTPRecording)
}

// SendAuthHTTPRequest sends an authenticated HTTP request
func (c *Coinbene) SendAuthHTTPRequest(method, path string, params url.Values, result interface{}, epPath string) error {
	if params == nil {
		params = url.Values{}
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	var finalBody io.Reader
	var preSign string
	if len(params) != 0 {

		m := make(map[string]string)
		for k, v := range params {
			m[k] = strings.Join(v, "")
		}
		tempBody, err := json.Marshal(m)
		if err != nil {
			return err
		}
		finalBody = bytes.NewBufferString(string(tempBody))
		cats, _ := json.Marshal(m)
		preSign = fmt.Sprintf("%s%s%s%s%s", timestamp, method, coinbeneAuthPath, epPath, cats)
	}
	if len(params) == 0 {
		preSign = fmt.Sprintf("%s%s%s%s%s", timestamp, method, coinbeneAuthPath, epPath, params.Encode())
	}
	tempSign := common.GetHMAC(common.HashSHA256, []byte(preSign), []byte(c.APISecret))
	hexEncodedd := common.HexEncodeToString(tempSign)
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["ACCESS-KEY"] = c.APIKey
	headers["ACCESS-SIGN"] = hexEncodedd
	headers["ACCESS-TIMESTAMP"] = timestamp
	log.Println(preSign)
	return c.SendPayload(method,
		path,
		headers,
		finalBody,
		&result,
		true,
		false,
		c.Verbose,
		c.HTTPDebugging,
		c.HTTPRecording)
}
