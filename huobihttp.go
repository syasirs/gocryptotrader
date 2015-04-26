package main

import (
	"net/url"
	"strings"
	"strconv"
	"time"
	"fmt"
	"log"
)

const (
	HUOBI_API_URL = "https://api.huobi.com/apiv2.php"
	HUOBI_API_VERSION = "2"
)

type HUOBI struct {
	Name string
	Enabled bool
	Verbose bool
	Websocket bool
	RESTPollingDelay time.Duration
	AccessKey, SecretKey string
	Fee float64
}

type HuobiTicker struct {
	High float64
	Low float64
	Last float64 
	Vol float64 
	Buy float64 
	Sell float64 
}

type HuobiTickerResponse struct {
	Time string
	Ticker HuobiTicker
}

func (h *HUOBI) SetDefaults() {
	h.Name = "Huobi"
	h.Enabled = true
	h.Fee = 0
	h.Verbose = false
	h.Websocket = false
	h.RESTPollingDelay = 10
}

func (h *HUOBI) GetName() (string) {
	return h.Name
}

func (h *HUOBI) SetEnabled(enabled bool) {
	h.Enabled = enabled
}

func (h *HUOBI) IsEnabled() (bool) {
	return h.Enabled
}

func (h *HUOBI) SetAPIKeys(apiKey, apiSecret string) {
	h.AccessKey = apiKey
	h.SecretKey = apiSecret
}

func (h *HUOBI) GetFee() (float64) {
	return h.Fee
}

func (h *HUOBI) Run() {
	if h.Verbose {
		log.Printf("%s Websocket: %s (url: %s).\n", h.GetName(), IsEnabled(h.Websocket), HUOBI_SOCKETIO_ADDRESS)
		log.Printf("%s polling delay: %ds.\n", h.GetName(), h.RESTPollingDelay)
	}

	if h.Websocket {
		go h.WebsocketClient()
	}

	for h.Enabled {
		go func() {
			HuobiBTC := h.GetTicker("btc")
			HuobiBTCLastUSD, _ := ConvertCurrency(HuobiBTC.Last, "CNY", "USD")
			HuobiBTCHighUSD, _ := ConvertCurrency(HuobiBTC.High, "CNY", "USD")
			HuobiBTCLowUSD, _ := ConvertCurrency(HuobiBTC.Low, "CNY", "USD")
			log.Printf("Huobi BTC: Last %f (%f) High %f (%f) Low %f (%f) Volume %f\n", HuobiBTCLastUSD, HuobiBTC.Last, HuobiBTCHighUSD, HuobiBTC.High, HuobiBTCLowUSD, HuobiBTC.Low, HuobiBTC.Vol)
			AddExchangeInfo(h.GetName(), "BTC", HuobiBTCLastUSD, HuobiBTC.Vol)
		}()

		go func() {
			HuobiLTC := h.GetTicker("ltc")
			HuobiLTCLastUSD, _ := ConvertCurrency(HuobiLTC.Last, "CNY", "USD")
			HuobiLTCHighUSD, _ := ConvertCurrency(HuobiLTC.High, "CNY", "USD")
			HuobiLTCLowUSD, _ := ConvertCurrency(HuobiLTC.Low, "CNY", "USD")
			log.Printf("Huobi LTC: Last %f (%f) High %f (%f) Low %f (%f) Volume %f\n", HuobiLTCLastUSD, HuobiLTC.Last, HuobiLTCHighUSD, HuobiLTC.High, HuobiLTCLowUSD, HuobiLTC.Low, HuobiLTC.Vol)
			AddExchangeInfo(h.GetName(), "LTC", HuobiLTCLastUSD, HuobiLTC.Vol)
		}()
		time.Sleep(time.Second * h.RESTPollingDelay)
	}
}

func (h *HUOBI) GetTicker(symbol string) (HuobiTicker) {
	resp := HuobiTickerResponse{}
	path := fmt.Sprintf("http://market.huobi.com/staticmarket/ticker_%s_json.js", symbol)
	err := SendHTTPGetRequest(path, true, &resp)

	if err != nil {
		log.Println(err)
		return HuobiTicker{}
	}
	return resp.Ticker
}

func (h *HUOBI) GetOrderBook(symbol string) (bool) {
	path := fmt.Sprintf("http://market.huobi.com/staticmarket/depth_%s_json.js", symbol)
	err := SendHTTPGetRequest(path, true, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (h *HUOBI) GetAccountInfo() {
	err := h.SendAuthenticatedRequest("get_account_info", url.Values{})

	if err != nil {
		log.Println(err)
	}
}


func (h *HUOBI) GetOrders(coinType int) {
	values := url.Values{}
	values.Set("coin_type", strconv.Itoa(coinType))
	err := h.SendAuthenticatedRequest("get_orders", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) GetOrderInfo(orderID, coinType int) {
	values := url.Values{}
	values.Set("id", strconv.Itoa(orderID))
	values.Set("coin_type", strconv.Itoa(coinType))
	err := h.SendAuthenticatedRequest("order_info", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) Trade(orderType string, coinType int, price, amount float64) {
	values := url.Values{}
	if orderType != "buy" {
		orderType = "sell"
	}
	values.Set("coin_type", strconv.Itoa(coinType))
	values.Set("amount", strconv.FormatFloat(amount, 'f', 8, 64))
	values.Set("price",  strconv.FormatFloat(price, 'f', 8, 64))
	err := h.SendAuthenticatedRequest(orderType, values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) MarketTrade(orderType string, coinType int, price, amount float64) {
	values := url.Values{}
	if orderType != "buy_market" {
		orderType = "sell_market"
	}
	values.Set("coin_type", strconv.Itoa(coinType))
	values.Set("amount", strconv.FormatFloat(amount, 'f', 8, 64))
	values.Set("price",  strconv.FormatFloat(price, 'f', 8, 64))
	err := h.SendAuthenticatedRequest(orderType, values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) CancelOrder(orderID, coinType int) {
	values := url.Values{}
	values.Set("coin_type", strconv.Itoa(coinType))
	values.Set("id", strconv.Itoa(orderID))
	err := h.SendAuthenticatedRequest("cancel_order", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) ModifyOrder(orderType string, coinType, orderID int, price, amount float64) {
	values := url.Values{}
	values.Set("coin_type", strconv.Itoa(coinType))
	values.Set("id", strconv.Itoa(orderID))
	values.Set("amount", strconv.FormatFloat(amount, 'f', 8, 64))
	values.Set("price",  strconv.FormatFloat(price, 'f', 8, 64))
	err := h.SendAuthenticatedRequest("modify_order", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) GetNewDealOrders(coinType int) {
	values := url.Values{}
	values.Set("coin_type", strconv.Itoa(coinType))
	err := h.SendAuthenticatedRequest("get_new_deal_orders", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) GetOrderIDByTradeID(coinType, orderID int) {
	values := url.Values{}
	values.Set("coin_type", strconv.Itoa(coinType))
	values.Set("trade_id", strconv.Itoa(orderID))
	err := h.SendAuthenticatedRequest("get_order_id_by_trade_id", values)

	if err != nil {
		log.Println(err)
	}
}

func (h *HUOBI) SendAuthenticatedRequest(method string, v url.Values) (error) {
	v.Set("access_key", h.AccessKey)
	v.Set("created", strconv.FormatInt(time.Now().Unix(), 10))
	v.Set("method", method)
	hash := GetMD5([]byte(v.Encode() + "&secret_key=" + h.SecretKey))
	v.Set("sign", strings.ToLower(HexEncodeToString(hash)))
	encoded := v.Encode()

	if h.Verbose {
		log.Printf("Sending POST request to %s with params %s\n", HUOBI_API_URL, encoded)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	resp, err := SendHTTPRequest("POST", HUOBI_API_URL, headers, strings.NewReader(encoded))

	if err != nil {
		return err
	}

	if h.Verbose {
		log.Printf("Recieved raw: %s\n", resp)
	}
	
	return nil
}