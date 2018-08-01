package gateio

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

const (
	gateioTradeURL   = "https://api.gateio.io"
	gateioMarketURL  = "https://data.gateio.io"
	gateioAPIVersion = "api2/1"

	gateioSymbol      = "pairs"
	gateioMarketInfo  = "marketinfo"
	gateioKline       = "candlestick2"
	gateioOrder       = "private"
	gateioBalances    = "private/balances"
	gateioCancelOrder = "private/cancelOrder"
	gateioTicker      = "ticker"
	gateioTickers     = "tickers"
	gateioOrderbook   = "orderBook"

	gateioAuthRate   = 100
	gateioUnauthRate = 100
)

// Gateio is the overarching type across this package
type Gateio struct {
	exchange.Base
}

// SetDefaults sets default values for the exchange
func (g *Gateio) SetDefaults() {
	g.Name = "GateIO"
	g.Enabled = false
	g.Verbose = false
	g.Websocket = false
	g.RESTPollingDelay = 10
	g.RequestCurrencyPairFormat.Delimiter = "_"
	g.RequestCurrencyPairFormat.Uppercase = false
	g.ConfigCurrencyPairFormat.Delimiter = "_"
	g.ConfigCurrencyPairFormat.Uppercase = true
	g.AssetTypes = []string{ticker.Spot}
	g.SupportsAutoPairUpdating = true
	g.SupportsRESTTickerBatching = true
	g.Requester = request.New(g.Name, request.NewRateLimit(time.Second*10, gateioAuthRate), request.NewRateLimit(time.Second*10, gateioUnauthRate), common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
}

// Setup sets user configuration
func (g *Gateio) Setup(exch config.ExchangeConfig) {
	if !exch.Enabled {
		g.SetEnabled(false)
	} else {
		g.Enabled = true
		g.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		g.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		g.APIAuthPEMKey = exch.APIAuthPEMKey
		g.SetHTTPClientTimeout(exch.HTTPTimeout)
		g.RESTPollingDelay = exch.RESTPollingDelay
		g.Verbose = exch.Verbose
		g.Websocket = exch.Websocket
		g.BaseCurrencies = common.SplitStrings(exch.BaseCurrencies, ",")
		g.AvailablePairs = common.SplitStrings(exch.AvailablePairs, ",")
		g.EnabledPairs = common.SplitStrings(exch.EnabledPairs, ",")
		err := g.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = g.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = g.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetSymbols 返回所有系统支持的交易对
func (g *Gateio) GetSymbols() ([]string, error) {
	var result []string

	url := fmt.Sprintf("%s/%s/%s", gateioMarketURL, gateioAPIVersion, gateioSymbol)

	err := g.SendHTTPRequest(url, &result)
	if err != nil {
		return nil, nil
	}
	return result, err
}

// GetMarketInfo 返回所有系统支持的交易市场的参数信息，包括交易费，最小下单量，价格精度等。
func (g *Gateio) GetMarketInfo() (MarketInfoResponse, error) {
	type response struct {
		Result string        `json:"result"`
		Pairs  []interface{} `json:"pairs"`
	}

	url := fmt.Sprintf("%s/%s/%s", gateioMarketURL, gateioAPIVersion, gateioMarketInfo)

	var res response
	var result MarketInfoResponse
	err := g.SendHTTPRequest(url, &res)
	if err != nil {
		return result, err
	}

	result.Result = res.Result
	for _, v := range res.Pairs {
		item := v.(map[string]interface{})
		for itemk, itemv := range item {
			pairv := itemv.(map[string]interface{})
			result.Pairs = append(result.Pairs, MarketInfoPairsResponse{
				Symbol:        itemk,
				DecimalPlaces: pairv["decimal_places"].(float64),
				MinAmount:     pairv["min_amount"].(float64),
				Fee:           pairv["fee"].(float64),
			})
		}
	}
	return result, nil
}

// GetLatestSpotPrice returns latest spot price of symbol
//
// symbol: string of currency pair
// 获取最新价格，官方每10秒钟更新一次
func (g *Gateio) GetLatestSpotPrice(symbol string) (float64, error) {
	res, err := g.GetTicker(symbol)
	if err != nil {
		return 0, err
	}

	return res.Last, nil
}

// GetTicker returns 单项交易行情，官方每10秒钟更新一次
func (g *Gateio) GetTicker(symbol string) (TickerResponse, error) {
	url := fmt.Sprintf("%s/%s/%s/%s", gateioMarketURL, gateioAPIVersion, gateioTicker, symbol)

	var res TickerResponse
	err := g.SendHTTPRequest(url, &res)
	if err != nil {
		return res, err
	}
	return res, nil
}

// GetTickers returns tickers for all symbols
func (g *Gateio) GetTickers() (map[string]TickerResponse, error) {
	url := fmt.Sprintf("%s/%s/%s", gateioMarketURL, gateioAPIVersion, gateioTickers)

	resp := make(map[string]TickerResponse)
	err := g.SendHTTPRequest(url, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetOrderbook returns the orderbook data for a suppled symbol
func (g *Gateio) GetOrderbook(symbol string) (Orderbook, error) {
	url := fmt.Sprintf("%s/%s/%s/%s", gateioMarketURL, gateioAPIVersion, gateioOrderbook, symbol)

	var resp OrderbookResponse
	err := g.SendHTTPRequest(url, &resp)
	if err != nil {
		return Orderbook{}, err
	}

	if resp.Result != "true" {
		return Orderbook{}, errors.New("result was not true")
	}

	var ob Orderbook

	// Asks are in reverse order
	for x := len(resp.Asks) - 1; x != 0; x-- {
		data := resp.Asks[x]

		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			continue
		}

		ob.Asks = append(ob.Asks, OrderbookItem{Price: price, Amount: amount})
	}

	for x := range resp.Bids {
		data := resp.Bids[x]

		price, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			continue
		}

		ob.Bids = append(ob.Bids, OrderbookItem{Price: price, Amount: amount})
	}

	ob.Result = resp.Result
	ob.Elapsed = resp.Elapsed
	return ob, nil
}

// GetSpotKline 返回市场最近时间段内的K先数据
func (g *Gateio) GetSpotKline(arg KlinesRequestParams) ([]*KLineResponse, error) {

	url := fmt.Sprintf("%s/%s/%s/%s?group_sec=%d&range_hour=%d", gateioMarketURL, gateioAPIVersion, gateioKline, arg.Symbol, arg.GroupSec, arg.HourSize)

	var rawKlines map[string]interface{}
	err := g.SendHTTPRequest(url, &rawKlines)
	if err != nil {
		return nil, err
	}

	var result []*KLineResponse

	if rawKlines == nil || rawKlines["data"] == nil {
		return nil, fmt.Errorf("rawKlines is nil. Err: %s", err)
	}

	//对于 Data数据，再次解析
	rawKlineDatasString, _ := json.Marshal(rawKlines["data"].([]interface{}))
	rawKlineDatas := [][]interface{}{}
	if err := json.Unmarshal(rawKlineDatasString, &rawKlineDatas); err != nil {
		return nil, fmt.Errorf("rawKlines unmarshal failed. Err: %s", err)
	}
	for _, k := range rawKlineDatas {
		otString, _ := strconv.ParseFloat(k[0].(string), 64)
		ot, err := common.TimeFromUnixTimestampFloat(otString)
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.OpenTime. Err: %s", err)
		}
		_vol, err := common.FloatFromString(k[1])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.Volume. Err: %s", err)
		}
		_id, err := common.FloatFromString(k[0])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.Id. Err: %s", err)
		}
		_close, err := common.FloatFromString(k[2])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.Close. Err: %s", err)
		}
		_high, err := common.FloatFromString(k[3])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.High. Err: %s", err)
		}
		_low, err := common.FloatFromString(k[4])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.Low. Err: %s", err)
		}
		_open, err := common.FloatFromString(k[5])
		if err != nil {
			return nil, fmt.Errorf("cannot parse Kline.Open. Err: %s", err)
		}
		result = append(result, &KLineResponse{
			ID:        _id,
			KlineTime: ot,
			Volume:    _vol,   //成交量
			Close:     _close, //收盘价
			High:      _high,  //最高
			Low:       _low,   //最低
			Open:      _open,  //开盘价
		})
	}

	return result, nil
}

// GetBalances 获取帐号资金余额
// 通过以下API，用户可以使用程序控制自动进行账号资金查询，下单交易，取消挂单。
// 请注意：请在您的程序中设置的HTTP请求头参数 Content-Type 为 application/x-www-form-urlencoded
// 用户首先要通过这个链接获取API接口身份认证用到的Key和Secret。 然后在程序中用Secret作为密码，通过SHA512加密方式签名需要POST给服务器的数据得到Sign，并在HTTPS请求的Header部分传回Key和Sign。请参考以下接口说明和例子程序进行设置。
func (g *Gateio) GetBalances() (BalancesResponse, error) {

	var result BalancesResponse

	err := g.SendAuthenticatedHTTPRequest("POST", gateioBalances, "", &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// SpotNewOrder 下订单
func (g *Gateio) SpotNewOrder(arg SpotNewOrderRequestParams) (SpotNewOrderResponse, error) {
	var result SpotNewOrderResponse

	//获取交易对的价格精度格式
	params := fmt.Sprintf("currencyPair=%s&rate=%s&amount=%s",
		arg.Symbol,
		strconv.FormatFloat(arg.Price, 'f', -1, 64),
		strconv.FormatFloat(arg.Amount, 'f', -1, 64),
	)

	strRequestURL := fmt.Sprintf("%s/%s", gateioOrder, arg.Type)

	err := g.SendAuthenticatedHTTPRequest("POST", strRequestURL, params, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// CancelOrder 取消订单
// @orderID 下单单号
// @symbol 交易币种对(如 ltc_btc)
func (g *Gateio) CancelOrder(orderID int64, symbol string) (bool, error) {
	type response struct {
		Result  bool   `json:"result"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	var result response
	//获取交易对的价格精度格式
	params := fmt.Sprintf("orderNumber=%d&currencyPair=%s",
		orderID,
		symbol,
	)
	err := g.SendAuthenticatedHTTPRequest("POST", gateioCancelOrder, params, &result)
	if err != nil {
		return false, err
	}
	if !result.Result {
		return false, fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	return true, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (g *Gateio) SendHTTPRequest(path string, result interface{}) error {
	return g.SendPayload("GET", path, nil, nil, result, false, g.Verbose)
}

// SendAuthenticatedHTTPRequest sends authenticated requests to the Gateio API
func (g *Gateio) SendAuthenticatedHTTPRequest(method, endpoint, param string, result interface{}) error {
	if !g.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, g.Name)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["key"] = g.APIKey

	hmac := common.GetHMAC(common.HashSHA512, []byte(param), []byte(g.APISecret))
	headers["sign"] = common.ByteArrayToString(hmac)

	url := fmt.Sprintf("%s/%s/%s", gateioTradeURL, gateioAPIVersion, endpoint)

	return g.SendPayload(method, url, headers, strings.NewReader(param), result, true, g.Verbose)
}
