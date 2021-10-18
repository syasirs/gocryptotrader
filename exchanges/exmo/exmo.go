package exmo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

const (
	exmoAPIURL        = "https://api.exmo.com"
	exmoAPIVersion    = "1"
	exmoAPIVersion1p1 = "1.1"

	exmoTrades       = "trades"
	exmoOrderbook    = "order_book"
	exmoTicker       = "ticker"
	exmoPairSettings = "pair_settings"
	exmoCurrency     = "currency"

	exmoUserInfo                  = "user_info"
	exmoOrderCreate               = "order_create"
	exmoOrderCancel               = "order_cancel"
	exmoOpenOrders                = "user_open_orders"
	exmoUserTrades                = "user_trades"
	exmoCancelledOrders           = "user_cancelled_orders"
	exmoOrderTrades               = "order_trades"
	exmoRequiredAmount            = "required_amount"
	exmoDepositAddress            = "deposit_address"
	exmoWithdrawCrypt             = "withdraw_crypt"
	exmoGetWithdrawTXID           = "withdraw_get_txid"
	exmoExcodeCreate              = "excode_create"
	exmoExcodeLoad                = "excode_load"
	exmoWalletHistory             = "wallet_history"
	exmoCryptoPaymentProviderList = "payments/providers/crypto/list"
	exmoPairList                  = "margin/pair/list"

	// Rate limit: 180 per/minute
	exmoRateInterval = time.Minute
	exmoRequestRate  = 180
)

// EXMO exchange struct
type EXMO struct {
	exchange.Base
}

// GetTrades returns the trades for a symbol or symbols
func (e *EXMO) GetTrades(ctx context.Context, symbol string) (map[string][]Trades, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string][]Trades)
	urlPath := fmt.Sprintf("/v%s/%s", exmoAPIVersion, exmoTrades)
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, common.EncodeURLValues(urlPath, v), &result)
}

// GetOrderbook returns the orderbook for a symbol or symbols
func (e *EXMO) GetOrderbook(ctx context.Context, symbol string) (map[string]Orderbook, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string]Orderbook)
	urlPath := fmt.Sprintf("/v%s/%s", exmoAPIVersion, exmoOrderbook)
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, common.EncodeURLValues(urlPath, v), &result)
}

// GetTicker returns the ticker for a symbol or symbols
func (e *EXMO) GetTicker(ctx context.Context) (map[string]Ticker, error) {
	v := url.Values{}
	result := make(map[string]Ticker)
	urlPath := fmt.Sprintf("/v%s/%s", exmoAPIVersion, exmoTicker)
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, common.EncodeURLValues(urlPath, v), &result)
}

// GetPairSettings returns the pair settings for a symbol or symbols
func (e *EXMO) GetPairSettings(ctx context.Context) (map[string]PairSettings, error) {
	result := make(map[string]PairSettings)
	urlPath := fmt.Sprintf("/v%s/%s", exmoAPIVersion, exmoPairSettings)
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, urlPath, &result)
}

// GetCurrency returns a list of currencies
func (e *EXMO) GetCurrency(ctx context.Context) ([]string, error) {
	var result []string
	urlPath := fmt.Sprintf("/v%s/%s", exmoAPIVersion, exmoCurrency)
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, urlPath, &result)
}

// GetUserInfo returns the user info
func (e *EXMO) GetUserInfo(ctx context.Context) (UserInfo, error) {
	var result UserInfo
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoUserInfo, url.Values{}, &result)
	return result, err
}

// CreateOrder creates an order
// Params: pair, quantity, price and type
// Type can be buy, sell, market_buy, market_sell, market_buy_total and market_sell_total
func (e *EXMO) CreateOrder(ctx context.Context, pair, orderType string, price, amount float64) (int64, error) {
	type response struct {
		OrderID int64  `json:"order_id"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
	}

	v := url.Values{}
	v.Set("pair", pair)
	v.Set("type", orderType)
	v.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))

	var resp response
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoOrderCreate, v, &resp)
	if !resp.Result {
		return -1, errors.New(resp.Error)
	}
	return resp.OrderID, err
}

// CancelExistingOrder cancels an order by the orderID
func (e *EXMO) CancelExistingOrder(ctx context.Context, orderID int64) error {
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))
	type response struct {
		Result bool   `json:"result"`
		Error  string `json:"error"`
	}
	var resp response
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoOrderCancel, v, &resp)
	if !resp.Result {
		return errors.New(resp.Error)
	}
	return err
}

// GetOpenOrders returns the users open orders
func (e *EXMO) GetOpenOrders(ctx context.Context) (map[string]OpenOrders, error) {
	result := make(map[string]OpenOrders)
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoOpenOrders, url.Values{}, &result)
	return result, err
}

// GetUserTrades returns the user trades
func (e *EXMO) GetUserTrades(ctx context.Context, pair, offset, limit string) (map[string][]UserTrades, error) {
	result := make(map[string][]UserTrades)
	v := url.Values{}
	v.Set("pair", pair)

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoUserTrades, v, &result)
	return result, err
}

// GetCancelledOrders returns a list of cancelled orders
func (e *EXMO) GetCancelledOrders(ctx context.Context, offset, limit string) ([]CancelledOrder, error) {
	var result []CancelledOrder
	v := url.Values{}

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoCancelledOrders, v, &result)
	return result, err
}

// GetOrderTrades returns a history of order trade details for the specific orderID
func (e *EXMO) GetOrderTrades(ctx context.Context, orderID int64) (OrderTrades, error) {
	var result OrderTrades
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))

	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoOrderTrades, v, &result)
	return result, err
}

// GetRequiredAmount calculates the sum of buying a certain amount of currency
// for the particular currency pair
func (e *EXMO) GetRequiredAmount(ctx context.Context, pair string, amount float64) (RequiredAmount, error) {
	v := url.Values{}
	v.Set("pair", pair)
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))
	var result RequiredAmount
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoRequiredAmount, v, &result)
	return result, err
}

// GetCryptoDepositAddress returns a list of addresses for cryptocurrency deposits
func (e *EXMO) GetCryptoDepositAddress(ctx context.Context) (map[string]string, error) {
	var result interface{}
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoDepositAddress, url.Values{}, &result)
	if err != nil {
		return nil, err
	}

	switch r := result.(type) {
	case map[string]interface{}:
		mapString := make(map[string]string)
		for key, value := range r {
			v, ok := value.(string)
			if !ok {
				return nil, errors.New("unable to type assert value data")
			}
			mapString[key] = v
		}
		return mapString, nil
	default:
		return nil, errors.New("no addresses found, generate required addresses via site")
	}
}

// WithdrawCryptocurrency withdraws a cryptocurrency from the exchange to the desired address
// NOTE: This API function is available only after request to their tech support team
func (e *EXMO) WithdrawCryptocurrency(ctx context.Context, currency, address, invoice, transport string, amount float64) (int64, error) {
	type response struct {
		TaskID  int64  `json:"task_id,string"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
		Success int64  `json:"success"`
	}

	v := url.Values{}
	v.Set("currency", currency)
	v.Set("address", address)

	if invoice != "" {
		v.Set("invoice", invoice)
	}

	if transport != "" {
		v.Set("transport", strings.ToUpper(transport))
	}

	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	var resp response
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoWithdrawCrypt, v, &resp)
	if err != nil {
		return -1, err
	}
	if resp.Success == 0 || !resp.Result {
		return -1, errors.New(resp.Error)
	}
	return resp.TaskID, err
}

// GetWithdrawTXID gets the result of a withdrawal request
func (e *EXMO) GetWithdrawTXID(ctx context.Context, taskID int64) (string, error) {
	type response struct {
		Status bool   `json:"status"`
		TXID   string `json:"txid"`
	}

	v := url.Values{}
	v.Set("task_id", strconv.FormatInt(taskID, 10))

	var result response
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoGetWithdrawTXID, v, &result)
	return result.TXID, err
}

// ExcodeCreate creates an EXMO coupon
func (e *EXMO) ExcodeCreate(ctx context.Context, currency string, amount float64) (ExcodeCreate, error) {
	v := url.Values{}
	v.Set("currency", currency)
	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	var result ExcodeCreate
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoExcodeCreate, v, &result)
	return result, err
}

// ExcodeLoad loads an EXMO coupon
func (e *EXMO) ExcodeLoad(ctx context.Context, excode string) (ExcodeLoad, error) {
	v := url.Values{}
	v.Set("code", excode)

	var result ExcodeLoad
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoExcodeLoad, v, &result)
	return result, err
}

// GetWalletHistory returns the users deposit/withdrawal history
func (e *EXMO) GetWalletHistory(ctx context.Context, date int64) (WalletHistory, error) {
	v := url.Values{}
	v.Set("date", strconv.FormatInt(date, 10))

	var result WalletHistory
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion, exmoWalletHistory, v, &result)
	return result, err
}

// GetPairInfo returns pair info for margin pairs
func (e *EXMO) GetPairInfo(ctx context.Context) ([]PairInformation, error) {
	var result struct {
		Pairs []PairInformation `json:"pairs"`
	}
	v := url.Values{}
	err := e.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, exmoAPIVersion1p1, exmoPairList, v, &result)
	return result.Pairs, err
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (e *EXMO) SendHTTPRequest(ctx context.Context, endpoint exchange.URL, path string, result interface{}) error {
	urlPath, err := e.API.Endpoints.GetURL(endpoint)
	if err != nil {
		return err
	}

	item := &request.Item{
		Method:        http.MethodGet,
		Path:          urlPath + path,
		Result:        result,
		Verbose:       e.Verbose,
		HTTPDebugging: e.HTTPDebugging,
		HTTPRecording: e.HTTPRecording,
	}
	return e.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		return item, nil
	})
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (e *EXMO) SendAuthenticatedHTTPRequest(ctx context.Context, epath exchange.URL, method, version, endpoint string, vals url.Values, result interface{}) error {
	if !e.AllowAuthenticatedRequest() {
		return fmt.Errorf("%s %w", e.Name, exchange.ErrAuthenticatedRequestWithoutCredentialsSet)
	}

	urlPath, err := e.API.Endpoints.GetURL(epath)
	if err != nil {
		return err
	}

	path := urlPath + fmt.Sprintf("/v%s/%s", version, endpoint)

	return e.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		n := e.Requester.GetNonce(true).String()
		vals.Set("nonce", n)

		payload := vals.Encode()
		hash, err := crypto.GetHMAC(crypto.HashSHA512,
			[]byte(payload),
			[]byte(e.API.Credentials.Secret))
		if err != nil {
			return nil, err
		}

		headers := make(map[string]string)
		headers["Key"] = e.API.Credentials.Key
		headers["Sign"] = crypto.HexEncodeToString(hash)
		headers["Content-Type"] = "application/x-www-form-urlencoded"

		return &request.Item{
			Method:        method,
			Path:          path,
			Headers:       headers,
			Body:          strings.NewReader(payload),
			Result:        result,
			AuthRequest:   true,
			NonceEnabled:  true,
			Verbose:       e.Verbose,
			HTTPDebugging: e.HTTPDebugging,
			HTTPRecording: e.HTTPRecording,
		}, nil
	})
}

// GetCryptoPaymentProvidersList returns a map of all the supported cryptocurrency transfer settings
func (e *EXMO) GetCryptoPaymentProvidersList(ctx context.Context) (map[string][]CryptoPaymentProvider, error) {
	var result map[string][]CryptoPaymentProvider
	path := "/v" + exmoAPIVersion + "/" + exmoCryptoPaymentProviderList
	return result, e.SendHTTPRequest(ctx, exchange.RestSpot, path, &result)
}
