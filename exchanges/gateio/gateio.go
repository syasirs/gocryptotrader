package gateio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

const (
	gateioTradeURL   = "https://api.gateio.io"
	gateioMarketURL  = "https://data.gateio.io"
	gateioAPIVersion = "api2/1"

	gateioSymbol          = "pairs"
	gateioMarketInfo      = "marketinfo"
	gateioKline           = "candlestick2"
	gateioOrder           = "private"
	gateioBalances        = "private/balances"
	gateioCancelOrder     = "private/cancelOrder"
	gateioCancelAllOrders = "private/cancelAllOrders"
	gateioWithdraw        = "private/withdraw"
	gateioOpenOrders      = "private/openOrders"
	gateioTradeHistory    = "private/tradeHistory"
	gateioDepositAddress  = "private/depositAddress"
	gateioTicker          = "ticker"
	gateioTrades          = "tradeHistory"
	gateioTickers         = "tickers"
	gateioOrderbook       = "orderBook"

	gateioGenerateAddress = "New address is being generated for you, please wait a moment and refresh this page. "
)

// Gateio is the overarching type across this package
type Gateio struct {
	exchange.Base
}

// GetSymbols returns all supported symbols
func (g *Gateio) GetSymbols(ctx context.Context) ([]string, error) {
	var result []string
	urlPath := fmt.Sprintf("/%s/%s", gateioAPIVersion, gateioSymbol)
	err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &result)
	return result, err
}

// GetMarketInfo returns information about all trading pairs, including
// transaction fee, minimum order quantity, price accuracy and so on
func (g *Gateio) GetMarketInfo(ctx context.Context) (MarketInfoResponse, error) {
	type response struct {
		Result string        `json:"result"`
		Pairs  []interface{} `json:"pairs"`
	}

	urlPath := fmt.Sprintf("/%s/%s", gateioAPIVersion, gateioMarketInfo)
	var res response
	var result MarketInfoResponse
	err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &res)
	if err != nil {
		return result, err
	}

	result.Result = res.Result
	for _, v := range res.Pairs {
		item, ok := v.(map[string]interface{})
		if !ok {
			return result, errors.New("unable to type assert item")
		}
		for itemk, itemv := range item {
			pairv, ok := itemv.(map[string]interface{})
			if !ok {
				return result, errors.New("unable to type assert pairv")
			}
			decimalPlaces, ok := pairv["decimal_places"].(float64)
			if !ok {
				return result, errors.New("unable to type assert decimal_places")
			}
			minAmount, ok := pairv["min_amount"].(float64)
			if !ok {
				return result, errors.New("unable to type assert min_amount")
			}
			fee, ok := pairv["fee"].(float64)
			if !ok {
				return result, errors.New("unable to type assert fee")
			}
			result.Pairs = append(result.Pairs, MarketInfoPairsResponse{
				Symbol:        itemk,
				DecimalPlaces: decimalPlaces,
				MinAmount:     minAmount,
				Fee:           fee,
			})
		}
	}
	return result, nil
}

// GetLatestSpotPrice returns latest spot price of symbol
// updated every 10 seconds
//
// symbol: string of currency pair
func (g *Gateio) GetLatestSpotPrice(ctx context.Context, symbol string) (float64, error) {
	res, err := g.GetTicker(ctx, symbol)
	if err != nil {
		return 0, err
	}

	return res.Last, nil
}

// GetTicker returns a ticker for the supplied symbol
// updated every 10 seconds
func (g *Gateio) GetTicker(ctx context.Context, symbol string) (TickerResponse, error) {
	urlPath := fmt.Sprintf("/%s/%s/%s", gateioAPIVersion, gateioTicker, symbol)
	var res TickerResponse
	return res, g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &res)
}

// GetTickers returns tickers for all symbols
func (g *Gateio) GetTickers(ctx context.Context) (map[string]TickerResponse, error) {
	urlPath := fmt.Sprintf("/%s/%s", gateioAPIVersion, gateioTickers)
	resp := make(map[string]TickerResponse)
	err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetTrades returns trades for symbols
func (g *Gateio) GetTrades(ctx context.Context, symbol string) (TradeHistory, error) {
	urlPath := fmt.Sprintf("/%s/%s/%s", gateioAPIVersion, gateioTrades, symbol)
	var resp TradeHistory
	err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &resp)
	if err != nil {
		return TradeHistory{}, err
	}
	return resp, nil
}

// GetOrderbook returns the orderbook data for a suppled symbol
func (g *Gateio) GetOrderbook(ctx context.Context, symbol string) (*Orderbook, error) {
	urlPath := fmt.Sprintf("/%s/%s/%s", gateioAPIVersion, gateioOrderbook, symbol)
	var resp OrderbookResponse
	err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &resp)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.Result != "true":
		return nil, errors.New("result was not true")
	case len(resp.Asks) == 0:
		return nil, errors.New("asks are empty")
	case len(resp.Bids) == 0:
		return nil, errors.New("bids are empty")
	}

	// Asks are in reverse order
	ob := Orderbook{
		Result:  resp.Result,
		Elapsed: resp.Elapsed,
		Bids:    make([]OrderbookItem, len(resp.Bids)),
		Asks:    make([]OrderbookItem, 0, len(resp.Asks)),
	}

	for x := len(resp.Asks) - 1; x != 0; x-- {
		price, err := strconv.ParseFloat(resp.Asks[x][0], 64)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseFloat(resp.Asks[x][1], 64)
		if err != nil {
			return nil, err
		}

		ob.Asks = append(ob.Asks, OrderbookItem{Price: price, Amount: amount})
	}

	for x := range resp.Bids {
		price, err := strconv.ParseFloat(resp.Bids[x][0], 64)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseFloat(resp.Bids[x][1], 64)
		if err != nil {
			return nil, err
		}

		ob.Bids[x] = OrderbookItem{Price: price, Amount: amount}
	}
	return &ob, nil
}

// GetSpotKline returns kline data for the most recent time period
func (g *Gateio) GetSpotKline(ctx context.Context, arg KlinesRequestParams) (kline.Item, error) {
	urlPath := fmt.Sprintf("/%s/%s/%s?group_sec=%s&range_hour=%d",
		gateioAPIVersion,
		gateioKline,
		arg.Symbol,
		arg.GroupSec,
		arg.HourSize)

	resp := struct {
		Data   [][]string `json:"data"`
		Result string     `json:"result"`
	}{}

	if err := g.SendHTTPRequest(ctx, exchange.RestSpotSupplementary, urlPath, &resp); err != nil {
		return kline.Item{}, err
	}
	if resp.Result != "true" || len(resp.Data) == 0 {
		return kline.Item{}, errors.New("rawKlines unexpected data returned")
	}

	result := kline.Item{
		Exchange: g.Name,
	}

	for x := range resp.Data {
		if len(resp.Data[x]) < 6 {
			return kline.Item{}, fmt.Errorf("unexpected kline data length")
		}
		otString, err := strconv.ParseFloat(resp.Data[x][0], 64)
		if err != nil {
			return kline.Item{}, err
		}
		ot, err := convert.TimeFromUnixTimestampFloat(otString)
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.OpenTime. Err: %s", err)
		}
		_vol, err := convert.FloatFromString(resp.Data[x][1])
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.Volume. Err: %s", err)
		}
		_close, err := convert.FloatFromString(resp.Data[x][2])
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.Close. Err: %s", err)
		}
		_high, err := convert.FloatFromString(resp.Data[x][3])
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.High. Err: %s", err)
		}
		_low, err := convert.FloatFromString(resp.Data[x][4])
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.Low. Err: %s", err)
		}
		_open, err := convert.FloatFromString(resp.Data[x][5])
		if err != nil {
			return kline.Item{}, fmt.Errorf("cannot parse Kline.Open. Err: %s", err)
		}
		result.Candles = append(result.Candles, kline.Candle{
			Time:   ot,
			Volume: _vol,
			Close:  _close,
			High:   _high,
			Low:    _low,
			Open:   _open,
		})
	}
	return result, nil
}

// GetBalances obtains the users account balance
func (g *Gateio) GetBalances(ctx context.Context) (BalancesResponse, error) {
	var result BalancesResponse
	return result,
		g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioBalances, "", &result)
}

// SpotNewOrder places a new order
func (g *Gateio) SpotNewOrder(ctx context.Context, arg SpotNewOrderRequestParams) (SpotNewOrderResponse, error) {
	var result SpotNewOrderResponse

	// Be sure to use the correct price precision before calling this
	params := fmt.Sprintf("currencyPair=%s&rate=%s&amount=%s",
		arg.Symbol,
		strconv.FormatFloat(arg.Price, 'f', -1, 64),
		strconv.FormatFloat(arg.Amount, 'f', -1, 64),
	)

	urlPath := fmt.Sprintf("%s/%s", gateioOrder, arg.Type)
	return result, g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, urlPath, params, &result)
}

// CancelExistingOrder cancels an order given the supplied orderID and symbol
// orderID order ID number
// symbol trade pair (ltc_btc)
func (g *Gateio) CancelExistingOrder(ctx context.Context, orderID int64, symbol string) (bool, error) {
	type response struct {
		Result  bool   `json:"result"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	var result response
	// Be sure to use the correct price precision before calling this
	params := fmt.Sprintf("orderNumber=%d&currencyPair=%s",
		orderID,
		symbol,
	)
	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioCancelOrder, params, &result)
	if err != nil {
		return false, err
	}
	if !result.Result {
		return false, fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	return true, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (g *Gateio) SendHTTPRequest(ctx context.Context, ep exchange.URL, path string, result interface{}) error {
	endpoint, err := g.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	item := &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       g.Verbose,
		HTTPDebugging: g.HTTPDebugging,
		HTTPRecording: g.HTTPRecording,
	}
	return g.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		return item, nil
	})
}

// CancelAllExistingOrders all orders for a given symbol and side
// orderType (0: sell,1: buy,-1: unlimited)
func (g *Gateio) CancelAllExistingOrders(ctx context.Context, orderType int64, symbol string) error {
	type response struct {
		Result  bool   `json:"result"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	var result response
	params := fmt.Sprintf("type=%d&currencyPair=%s",
		orderType,
		symbol,
	)
	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioCancelAllOrders, params, &result)
	if err != nil {
		return err
	}

	if !result.Result {
		return fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	return nil
}

// GetOpenOrders retrieves all open orders with an optional symbol filter
func (g *Gateio) GetOpenOrders(ctx context.Context, symbol string) (OpenOrdersResponse, error) {
	var params string
	var result OpenOrdersResponse

	if symbol != "" {
		params = fmt.Sprintf("currencyPair=%s", symbol)
	}

	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioOpenOrders, params, &result)
	if err != nil {
		return result, err
	}

	if result.Code > 0 {
		return result, fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	return result, nil
}

// GetTradeHistory retrieves all orders with an optional symbol filter
func (g *Gateio) GetTradeHistory(ctx context.Context, symbol string) (TradHistoryResponse, error) {
	var params string
	var result TradHistoryResponse
	params = fmt.Sprintf("currencyPair=%s", symbol)

	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioTradeHistory, params, &result)
	if err != nil {
		return result, err
	}

	if result.Code > 0 {
		return result, fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	return result, nil
}

// GenerateSignature returns hash for authenticated requests
func (g *Gateio) GenerateSignature(secret, message string) ([]byte, error) {
	return crypto.GetHMAC(crypto.HashSHA512, []byte(message), []byte(secret))
}

// SendAuthenticatedHTTPRequest sends authenticated requests to the Gateio API
// To use this you must setup an APIKey and APISecret from the exchange
func (g *Gateio) SendAuthenticatedHTTPRequest(ctx context.Context, ep exchange.URL, method, endpoint, param string, result interface{}) error {
	creds, err := g.GetCredentials(ctx)
	if err != nil {
		return err
	}
	ePoint, err := g.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["key"] = creds.Key

	hmac, err := g.GenerateSignature(creds.Secret, param)
	if err != nil {
		return err
	}

	headers["sign"] = crypto.HexEncodeToString(hmac)

	urlPath := fmt.Sprintf("%s/%s/%s", ePoint, gateioAPIVersion, endpoint)

	var intermidiary json.RawMessage
	item := &request.Item{
		Method:        method,
		Path:          urlPath,
		Headers:       headers,
		Result:        &intermidiary,
		AuthRequest:   true,
		Verbose:       g.Verbose,
		HTTPDebugging: g.HTTPDebugging,
		HTTPRecording: g.HTTPRecording,
	}
	err = g.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		item.Body = strings.NewReader(param)
		return item, nil
	})
	if err != nil {
		return err
	}

	errCap := struct {
		Result  bool   `json:"result,string"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	if err := json.Unmarshal(intermidiary, &errCap); err == nil {
		if !errCap.Result {
			return fmt.Errorf("%s auth request error, code: %d message: %s",
				g.Name,
				errCap.Code,
				errCap.Message)
		}
	}

	return json.Unmarshal(intermidiary, result)
}

// GetFee returns an estimate of fee based on type of transaction
func (g *Gateio) GetFee(ctx context.Context, feeBuilder *exchange.FeeBuilder) (fee float64, err error) {
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		feePairs, err := g.GetMarketInfo(ctx)
		if err != nil {
			return 0, err
		}

		currencyPair := feeBuilder.Pair.Base.String() +
			feeBuilder.Pair.Delimiter +
			feeBuilder.Pair.Quote.String()

		var feeForPair float64
		for _, i := range feePairs.Pairs {
			if strings.EqualFold(currencyPair, i.Symbol) {
				feeForPair = i.Fee
			}
		}

		if feeForPair == 0 {
			return 0, fmt.Errorf("currency '%s' failed to find fee data",
				currencyPair)
		}

		fee = calculateTradingFee(feeForPair,
			feeBuilder.PurchasePrice,
			feeBuilder.Amount)

	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.Pair.Base)
	case exchange.OfflineTradeFee:
		fee = getOfflineTradeFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

// getOfflineTradeFee calculates the worst case-scenario trading fee
func getOfflineTradeFee(price, amount float64) float64 {
	return 0.002 * price * amount
}

func calculateTradingFee(feeForPair, purchasePrice, amount float64) float64 {
	return (feeForPair / 100) * purchasePrice * amount
}

func getCryptocurrencyWithdrawalFee(c currency.Code) float64 {
	return WithdrawalFees[c]
}

// WithdrawCrypto withdraws cryptocurrency to your selected wallet
func (g *Gateio) WithdrawCrypto(ctx context.Context, curr, address, memo, chain string, amount float64) (*withdraw.ExchangeResponse, error) {
	if curr == "" || address == "" || amount <= 0 {
		return nil, errors.New("currency, address and amount must be set")
	}

	resp := struct {
		Result  bool   `json:"result"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{}

	vals := url.Values{}
	vals.Set("currency", strings.ToUpper(curr))
	vals.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	// Transaction MEMO has to be entered after the address separated by a space
	if memo != "" {
		address += " " + memo
	}
	vals.Set("address", address)

	if chain != "" {
		vals.Set("chain", strings.ToUpper(chain))
	}

	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioWithdraw, vals.Encode(), &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, fmt.Errorf("code:%d message:%s", resp.Code, resp.Message)
	}

	return &withdraw.ExchangeResponse{
		Status: resp.Message,
	}, nil
}

// GetCryptoDepositAddress returns a deposit address for a cryptocurrency
func (g *Gateio) GetCryptoDepositAddress(ctx context.Context, curr string) (*DepositAddr, error) {
	var result DepositAddr
	params := fmt.Sprintf("currency=%s",
		curr)

	err := g.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, gateioDepositAddress, params, &result)
	if err != nil {
		return nil, err
	}

	if !result.Result {
		return nil, fmt.Errorf("code:%d message:%s", result.Code, result.Message)
	}

	// For memo/payment ID currencies
	if strings.Contains(result.Address, " ") {
		split := strings.Split(result.Address, " ")
		result.Address = split[0]
		result.Tag = split[1]
	}
	return &result, nil
}
