package kucoin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/types"
)

// Kucoin is the overarching type across this package
type Kucoin struct {
	exchange.Base
	obm *orderbookManager

	// HFSpot whether a high-frequency spot order is enabled or not.
	HFSpot bool
	// HFMargin whether a high-frequency margin order is enabled or not.
	HFMargin bool
}

var locker sync.Mutex

const (
	kucoinAPIURL        = "https://api.kucoin.com/api"
	kucoinAPIKeyVersion = "2"

	symbolQuery = "?symbol="
)

// GetSymbols gets pairs details on the exchange
// For market details see endpoint: https://www.kucoin.com/docs/rest/spot-trading/market-data/get-market-list
func (ku *Kucoin) GetSymbols(ctx context.Context, market string) ([]SymbolInfo, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	var resp []SymbolInfo
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, symbolsEPL, common.EncodeURLValues("/v2/symbols", params), &resp)
}

// GetTicker gets pair ticker information
func (ku *Kucoin) GetTicker(ctx context.Context, pair string) (*Ticker, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var resp *Ticker
	err := ku.SendHTTPRequest(ctx, exchange.RestSpot, tickersEPL, common.EncodeURLValues("/v1/market/orderbook/level1", params), &resp)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, common.ErrNoResponse
	}
	return resp, nil
}

// GetTickers gets all trading pair ticker information including 24h volume
func (ku *Kucoin) GetTickers(ctx context.Context) (*TickersResponse, error) {
	var resp *TickersResponse
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, allTickersEPL, "/v1/market/allTickers", &resp)
}

// Get24hrStats get the statistics of the specified pair in the last 24 hours
func (ku *Kucoin) Get24hrStats(ctx context.Context, pair string) (*Stats24hrs, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var resp *Stats24hrs
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, statistics24HrEPL, common.EncodeURLValues("/v1/market/stats", params), &resp)
}

// GetMarketList get the transaction currency for the entire trading market
func (ku *Kucoin) GetMarketList(ctx context.Context) ([]string, error) {
	var resp []string
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, marketListEPL, "/v1/markets", &resp)
}

func processOB(ob [][2]types.Number) []orderbook.Item {
	o := make([]orderbook.Item, len(ob))
	for x := range ob {
		o[x].Amount = ob[x][1].Float64()
		o[x].Price = ob[x][0].Float64()
	}
	return o
}

func constructOrderbook(o *orderbookResponse) (*Orderbook, error) {
	var (
		s = Orderbook{
			Bids: processOB(o.Bids),
			Asks: processOB(o.Asks),
			Time: o.Time.Time(),
		}
	)
	if o.Sequence != "" {
		var err error
		s.Sequence, err = strconv.ParseInt(o.Sequence, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

// GetPartOrderbook20 gets orderbook for a specified pair with depth 20
func (ku *Kucoin) GetPartOrderbook20(ctx context.Context, pair string) (*Orderbook, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var o *orderbookResponse
	err := ku.SendHTTPRequest(ctx, exchange.RestSpot, partOrderbook20EPL, common.EncodeURLValues("/v1/market/orderbook/level2_20", params), &o)
	if err != nil {
		return nil, err
	}
	return constructOrderbook(o)
}

// GetPartOrderbook100 gets orderbook for a specified pair with depth 100
func (ku *Kucoin) GetPartOrderbook100(ctx context.Context, pair string) (*Orderbook, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var o *orderbookResponse
	err := ku.SendHTTPRequest(ctx, exchange.RestSpot, partOrderbook100EPL, common.EncodeURLValues("/v1/market/orderbook/level2_100", params), &o)
	if err != nil {
		return nil, err
	}
	return constructOrderbook(o)
}

// GetOrderbook gets full orderbook for a specified pair
func (ku *Kucoin) GetOrderbook(ctx context.Context, pair string) (*Orderbook, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var o *orderbookResponse
	err := ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, fullOrderbookEPL, http.MethodGet, common.EncodeURLValues("/v3/market/orderbook/level2", params), nil, &o)
	if err != nil {
		return nil, err
	}
	return constructOrderbook(o)
}

// GetTradeHistory gets trade history of the specified pair
func (ku *Kucoin) GetTradeHistory(ctx context.Context, pair string) ([]Trade, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	var resp []Trade
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, tradeHistoryEPL, common.EncodeURLValues("/v1/market/histories", params), &resp)
}

// GetKlines gets kline of the specified pair
func (ku *Kucoin) GetKlines(ctx context.Context, pair, period string, start, end time.Time) ([]Kline, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", pair)
	if period == "" {
		return nil, errors.New("period can not be empty")
	}
	if !common.StringDataContains(validPeriods, period) {
		return nil, errors.New("invalid period")
	}
	params.Set("type", period)
	if !start.IsZero() {
		params.Set("startAt", strconv.FormatInt(start.Unix(), 10))
	}
	if !end.IsZero() {
		params.Set("endAt", strconv.FormatInt(end.Unix(), 10))
	}
	var resp [][7]types.Number
	err := ku.SendHTTPRequest(ctx, exchange.RestSpot, klinesEPL, common.EncodeURLValues("/v1/market/candles", params), &resp)
	if err != nil {
		return nil, err
	}
	klines := make([]Kline, len(resp))
	for i := range resp {
		klines[i].StartTime = time.Unix(resp[i][0].Int64(), 0)
		klines[i].Open = resp[i][1].Float64()
		klines[i].Close = resp[i][2].Float64()
		klines[i].High = resp[i][3].Float64()
		klines[i].Low = resp[i][4].Float64()
		klines[i].Volume = resp[i][5].Float64()
		klines[i].Amount = resp[i][6].Float64()
	}
	return klines, nil
}

// GetCurrenciesV3 the V3 of retrieving list of currencies
func (ku *Kucoin) GetCurrenciesV3(ctx context.Context) ([]CurrencyDetail, error) {
	var resp []CurrencyDetail
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, spotCurrenciesV3EPL, "/v3/currencies", &resp)
}

// GetCurrencyDetailV3 V3 endpoint to gets currency detail using currency code and chain information.
func (ku *Kucoin) GetCurrencyDetailV3(ctx context.Context, ccy, chain string) (*CurrencyDetail, error) {
	return ku.getCurrencyDetail(ctx, ccy, chain, "/v3/currencies/")
}

func (ku *Kucoin) getCurrencyDetail(ctx context.Context, ccy, chain, path string) (*CurrencyDetail, error) {
	if ccy == "" {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	if chain != "" {
		params.Set("chain", chain)
	}
	var resp *CurrencyDetail
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, spotCurrencyDetailEPL, common.EncodeURLValues(path+strings.ToUpper(ccy), params), &resp)
}

// GetFiatPrice gets fiat prices of currencies, default base currency is USD
func (ku *Kucoin) GetFiatPrice(ctx context.Context, base, currencies string) (map[string]types.Number, error) {
	params := url.Values{}
	if base != "" {
		params.Set("base", base)
	}
	if currencies != "" {
		params.Set("currencies", currencies)
	}
	var resp map[string]types.Number
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, fiatPriceEPL, common.EncodeURLValues("/v1/prices", params), &resp)
}

// GetLeveragedTokenInfo returns leveraged token information
func (ku *Kucoin) GetLeveragedTokenInfo(ctx context.Context, ccy string) ([]LeveragedTokenInfo, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	var resp []LeveragedTokenInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, leveragedTokenInfoEPL, http.MethodGet, common.EncodeURLValues("/v3/etf/info", params), nil, &resp)
}

// GetMarkPrice gets index price of the specified pair
func (ku *Kucoin) GetMarkPrice(ctx context.Context, pair string) (*MarkPrice, error) {
	if pair == "" {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var resp *MarkPrice
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, getMarkPriceEPL, "/v1/mark-price/"+pair+"/current", &resp)
}

// GetMarginConfiguration gets configure info of the margin
func (ku *Kucoin) GetMarginConfiguration(ctx context.Context) (*MarginConfiguration, error) {
	var resp *MarginConfiguration
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, getMarginConfigurationEPL, "/v1/margin/config", &resp)
}

// GetMarginAccount gets configure info of the margin
func (ku *Kucoin) GetMarginAccount(ctx context.Context) (*MarginAccounts, error) {
	var resp *MarginAccounts
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, marginAccountDetailEPL, http.MethodGet, "/v1/margin/account", nil, &resp)
}

// GetCrossIsolatedMarginRiskLimitCurrencyConfig risk limit and currency configuration of cross margin/isolated margin.
// isIsolated: true - isolated, false - cross ; default false
func (ku *Kucoin) GetCrossIsolatedMarginRiskLimitCurrencyConfig(ctx context.Context, isIsolated bool, symbol, ccy string) ([]RiskLimitCurrencyConfig, error) {
	params := url.Values{}
	if isIsolated {
		params.Set("isIsolated", "true")
	} else {
		params.Set("isIsolated", "false")
	}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	var resp []RiskLimitCurrencyConfig
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, crossIsolatedMarginRiskLimitCurrencyConfigEPL, http.MethodGet, common.EncodeURLValues("/v3/margin/currencies", params), nil, &resp)
}

// PostMarginBorrowOrder used to post borrow order
func (ku *Kucoin) PostMarginBorrowOrder(ctx context.Context, arg *MarginBorrowParam) (*BorrowAndRepaymentOrderResp, error) {
	if arg == nil || *arg == (MarginBorrowParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.TimeInForce == "" {
		return nil, errTimeInForceRequired
	}
	if arg.Size <= 0 {
		return nil, fmt.Errorf("%w , size = %f", order.ErrAmountBelowMin, arg.Size)
	}
	var resp *BorrowAndRepaymentOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, postMarginBorrowOrderEPL, http.MethodPost, "/v3/margin/borrow", arg, &resp)
}

// GetMarginBorrowingHistory retrieves the borrowing orders for cross and isolated margin accounts
func (ku *Kucoin) GetMarginBorrowingHistory(ctx context.Context, ccy currency.Code, isIsolated bool,
	symbol currency.Pair, orderNo string,
	startTime, endTime time.Time,
	currentPage, pageSize int64) (*BorrowRepayDetailResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	params.Set("currency", ccy.String())
	if isIsolated {
		params.Set("isIsonalted", "true")
	}
	if !symbol.IsEmpty() {
		params.Set("symbol", symbol.String())
	}
	if orderNo != "" {
		params.Set("orderNo", orderNo)
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if currentPage != 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize != 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *BorrowRepayDetailResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, marginBorrowingHistoryEPL, http.MethodGet, common.EncodeURLValues("/v3/margin/borrow", params), nil, &resp)
}

// PostRepayment used to initiate an application for the repayment of cross or isolated margin borrowing.
func (ku *Kucoin) PostRepayment(ctx context.Context, arg *RepayParam) (*BorrowAndRepaymentOrderResp, error) {
	if arg == nil || *arg == (RepayParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.Size <= 0 {
		return nil, fmt.Errorf("%w , size = %f", order.ErrAmountBelowMin, arg.Size)
	}
	var resp *BorrowAndRepaymentOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, postMarginRepaymentEPL, http.MethodPost, "/v3/margin/repay", arg, &resp)
}

// GetRepaymentHistory retrieves the repayment orders for cross and isolated margin accounts.
func (ku *Kucoin) GetRepaymentHistory(ctx context.Context, ccy currency.Code, isIsolated bool,
	symbol currency.Pair, orderNo string,
	startTime, endTime time.Time,
	currentPage, pageSize int64) (*BorrowRepayDetailResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	params.Set("currency", ccy.String())
	if isIsolated {
		params.Set("isIsonalted", "true")
	}
	if !symbol.IsEmpty() {
		params.Set("symbol", symbol.String())
	}
	if orderNo != "" {
		params.Set("orderNo", orderNo)
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if currentPage != 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize != 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *BorrowRepayDetailResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, marginRepaymentHistoryEPL, http.MethodGet, common.EncodeURLValues("/v3/margin/repay", params), nil, &resp)
}

// GetIsolatedMarginPairConfig get the current isolated margin trading pair configuration
func (ku *Kucoin) GetIsolatedMarginPairConfig(ctx context.Context) ([]IsolatedMarginPairConfig, error) {
	var resp []IsolatedMarginPairConfig
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, isolatedMarginPairConfigEPL, http.MethodGet, "/v1/isolated/symbols", nil, &resp)
}

// GetIsolatedMarginAccountInfo get all isolated margin accounts of the current user
func (ku *Kucoin) GetIsolatedMarginAccountInfo(ctx context.Context, balanceCurrency string) (*IsolatedMarginAccountInfo, error) {
	params := url.Values{}
	if balanceCurrency != "" {
		params.Set("balanceCurrency", balanceCurrency)
	}
	var resp *IsolatedMarginAccountInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, isolatedMarginAccountInfoEPL, http.MethodGet, common.EncodeURLValues("/v1/isolated/accounts", params), nil, &resp)
}

// GetSingleIsolatedMarginAccountInfo get single isolated margin accounts of the current user
func (ku *Kucoin) GetSingleIsolatedMarginAccountInfo(ctx context.Context, symbol string) (*AssetInfo, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *AssetInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, singleIsolatedMarginAccountInfoEPL, http.MethodGet, "/v1/isolated/account/"+symbol, nil, &resp)
}

// GetCurrentServerTime gets the server time
func (ku *Kucoin) GetCurrentServerTime(ctx context.Context) (time.Time, error) {
	resp := struct {
		Timestamp convert.ExchangeTime `json:"data"`
		Error
	}{}
	err := ku.SendHTTPRequest(ctx, exchange.RestSpot, currentServerTimeEPL, "/v1/timestamp", &resp)
	if err != nil {
		return time.Time{}, err
	}
	return resp.Timestamp.Time(), nil
}

// GetServiceStatus gets the service status
func (ku *Kucoin) GetServiceStatus(ctx context.Context) (*ServiceStatus, error) {
	var resp *ServiceStatus
	return resp, ku.SendHTTPRequest(ctx, exchange.RestSpot, serviceStatusEPL, "/v1/status", &resp)
}

// --------------------------------------------- Spot High Frequency(HF) Pro Account ---------------------------

// HFSpotPlaceOrder places a high frequency spot order
// There are two types of orders: (limit) order: set price and quantity for the transaction. (market) order : set amount or quantity for the transaction.
func (ku *Kucoin) HFSpotPlaceOrder(ctx context.Context, arg *PlaceHFParam) (string, error) {
	return ku.spotHFPlaceOrder(ctx, arg, "/v1/hf/orders")
}

// SpotPlaceHFOrderTest order test endpoint, the request parameters and return parameters of this endpoint are exactly the same as the order endpoint,
// and can be used to verify whether the signature is correct and other operations.
func (ku *Kucoin) SpotPlaceHFOrderTest(ctx context.Context, arg *PlaceHFParam) (string, error) {
	return ku.spotHFPlaceOrder(ctx, arg, "/v1/hf/orders/test")
}

func (ku *Kucoin) spotHFPlaceOrder(ctx context.Context, arg *PlaceHFParam, path string) (string, error) {
	if arg == nil || *arg == (PlaceHFParam{}) {
		return "", common.ErrNilPointer
	}
	if arg.Symbol.IsEmpty() {
		return "", currency.ErrSymbolStringEmpty
	}
	if arg.OrderType == "" {
		return "", order.ErrTypeIsInvalid
	}
	if arg.Side == "" {
		return "", order.ErrSideIsInvalid
	}
	if arg.Price <= 0 {
		return "", order.ErrPriceBelowMin
	}
	if arg.Size <= 0 {
		return "", order.ErrAmountBelowMin
	}
	resp := &struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfPlaceOrderEPL, http.MethodPost, path, arg, &resp)
}

// SyncPlaceHFOrder this interface will synchronously return the order information after the order matching is completed.
func (ku *Kucoin) SyncPlaceHFOrder(ctx context.Context, arg *PlaceHFParam) (*SyncPlaceHFOrderResp, error) {
	if arg == nil || *arg == (PlaceHFParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Symbol.IsEmpty() {
		return nil, currency.ErrSymbolStringEmpty
	}
	if arg.OrderType == "" {
		return nil, order.ErrTypeIsInvalid
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	if arg.Price <= 0 {
		return nil, order.ErrPriceBelowMin
	}
	if arg.Size <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	var resp *SyncPlaceHFOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfSyncPlaceOrderEPL, http.MethodPost, "/v1/hf/orders/sync", arg, &resp)
}

// PlaceMultipleOrders endpoint supports sequential batch order placement from a single endpoint. A maximum of 5orders can be placed simultaneously.
func (ku *Kucoin) PlaceMultipleOrders(ctx context.Context, args []PlaceHFParam) ([]PlaceOrderResp, error) {
	if len(args) == 0 {
		return nil, common.ErrNilPointer
	}
	for i := range args {
		if args[i].Symbol.IsEmpty() {
			return nil, currency.ErrSymbolStringEmpty
		}
		if args[i].OrderType == "" {
			return nil, order.ErrTypeIsInvalid
		}
		if args[i].Side == "" {
			return nil, order.ErrSideIsInvalid
		}
		if args[i].Price <= 0 {
			return nil, order.ErrPriceBelowMin
		}
		if args[i].Size <= 0 {
			return nil, order.ErrAmountBelowMin
		}
	}
	var resp []PlaceOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfMultipleOrdersEPL, http.MethodPost, "/v1/hf/orders/multi", &PlaceOrderParams{OrderList: args}, &resp)
}

// SyncPlaceMultipleHFOrders this interface will synchronously return the order information after the order matching is completed
func (ku *Kucoin) SyncPlaceMultipleHFOrders(ctx context.Context, args []PlaceHFParam) ([]SyncPlaceHFOrderResp, error) {
	if len(args) == 0 {
		return nil, common.ErrNilPointer
	}
	for i := range args {
		if args[i].Symbol.IsEmpty() {
			return nil, currency.ErrSymbolStringEmpty
		}
		if args[i].OrderType == "" {
			return nil, order.ErrTypeIsInvalid
		}
		if args[i].Side == "" {
			return nil, order.ErrSideIsInvalid
		}
		if args[i].Price <= 0 {
			return nil, order.ErrPriceBelowMin
		}
		if args[i].Size <= 0 {
			return nil, order.ErrAmountBelowMin
		}
	}
	var resp []SyncPlaceHFOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfSyncPlaceMultipleHFOrdersEPL, http.MethodPost, "/v1/hf/orders/multi/sync", args, &resp)
}

// ModifyHFOrder modifies a high frequency order.
func (ku *Kucoin) ModifyHFOrder(ctx context.Context, arg *ModifyHFOrderParam) (string, error) {
	if arg == nil || *arg == (ModifyHFOrderParam{}) {
		return "", common.ErrNilPointer
	}
	if arg.Symbol.IsEmpty() {
		return "", currency.ErrCurrencyPairEmpty
	}
	resp := &struct {
		NewOrderID string `json:"newOrderId"`
	}{}
	return resp.NewOrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfModifyOrderEPL, http.MethodPost, "/v1/hf/orders/alter", arg, &resp)
}

// CancelHFOrder used to cancel a high-frequency order by orderId.
func (ku *Kucoin) CancelHFOrder(ctx context.Context, orderID, symbol string) (string, error) {
	if orderID == "" {
		return "", order.ErrOrderIDNotSet
	}
	if symbol == "" {
		return "", currency.ErrSymbolStringEmpty
	}
	resp := &struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelHFOrderEPL, http.MethodDelete, "/v1/hf/orders/"+orderID+symbolQuery+symbol, nil, &resp)
}

// SyncCancelHFOrder this interface will synchronously return the order information after the order canceling is completed.
func (ku *Kucoin) SyncCancelHFOrder(ctx context.Context, orderID, symbol string) (*SyncCancelHFOrderResp, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	return ku.syncCancelHFOrder(ctx, orderID, symbol, "/v1/hf/orders/sync/")
}

// SyncCancelHFOrderByClientOrderID this interface will synchronously return the order information after the order canceling is completed.
func (ku *Kucoin) SyncCancelHFOrderByClientOrderID(ctx context.Context, clientOrderID, symbol string) (*SyncCancelHFOrderResp, error) {
	if clientOrderID == "" {
		return nil, errInvalidClientOrderID
	}
	return ku.syncCancelHFOrder(ctx, clientOrderID, symbol, "/v1/hf/orders/sync/client-order/")
}

func (ku *Kucoin) syncCancelHFOrder(ctx context.Context, id, symbol, path string) (*SyncCancelHFOrderResp, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *SyncCancelHFOrderResp
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfSyncCancelOrderEPL, http.MethodDelete, path+id+symbolQuery+symbol, nil, &resp)
}

// CancelHFOrderByClientOrderID sends out a request to cancel a high-frequency order using clientOid.
func (ku *Kucoin) CancelHFOrderByClientOrderID(ctx context.Context, clientOrderID, symbol string) (string, error) {
	if clientOrderID == "" {
		return "", errInvalidClientOrderID
	}
	if symbol == "" {
		return "", currency.ErrSymbolStringEmpty
	}
	resp := &struct {
		ClientOrderID string `json:"clientOid"`
	}{}
	return resp.ClientOrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfCancelOrderByClientOrderIDEPL, http.MethodGet, "/v1/hf/orders/client-order/"+clientOrderID+symbolQuery+symbol, nil, &resp)
}

// CancelSpecifiedNumberHFOrdersByOrderID cancel the specified quantity of the order according to the orderId.
func (ku *Kucoin) CancelSpecifiedNumberHFOrdersByOrderID(ctx context.Context, orderID, symbol string, cancelSize float64) (*CancelOrderByNumberResponse, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if cancelSize == 0 {
		return nil, errors.New("invalid cancel size")
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("cancelSize", strconv.FormatFloat(cancelSize, 'f', -1, 64))
	var resp *CancelOrderByNumberResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelSpecifiedNumberHFOrdersByOrderIDEPL, http.MethodDelete, common.EncodeURLValues("/v1/hf/orders/cancel/"+orderID, params), nil, &resp)
}

// CancelAllHFOrdersBySymbol cancel all open high-frequency orders (orders created through
func (ku *Kucoin) CancelAllHFOrdersBySymbol(ctx context.Context, symbol string) (string, error) {
	if symbol == "" {
		return "", currency.ErrSymbolStringEmpty
	}
	var resp string
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfCancelAllOrdersBySymbolEPL, http.MethodDelete, "/v1/hf/orders?symbol="+symbol, nil, &resp)
}

// CancelAllHFOrders cancels all HF orders for all symbol.
func (ku *Kucoin) CancelAllHFOrders(ctx context.Context) (*CancelAllHFOrdersResponse, error) {
	var resp *CancelAllHFOrdersResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfCancelAllOrdersEPL, http.MethodDelete, "/v1/hf/orders/cancelAll", nil, &resp)
}

// GetActiveHFOrders obtain all active order lists, and the return value of the active order interface is the paged data of all uncompleted order lists.
func (ku *Kucoin) GetActiveHFOrders(ctx context.Context, symbol string) ([]HFOrderDetail, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp []HFOrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfGetAllActiveOrdersEPL, http.MethodGet, "/v1/hf/orders/active?symbol="+symbol, nil, &resp)
}

// GetSymbolsWithActiveHFOrderList retrieves all trading pairs that the user has active orders
func (ku *Kucoin) GetSymbolsWithActiveHFOrderList(ctx context.Context) ([]string, error) {
	resp := &struct {
		Symbols []string `json:"symbols"`
	}{}
	return resp.Symbols, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfSymbolsWithActiveOrdersEPL, http.MethodGet, "/v1/hf/orders/active/symbols", nil, &resp)
}

// GetHFCompletedOrderList obtains a list of filled HF orders and returns paginated data. The returned data is sorted in descending order based on the latest order update times.
func (ku *Kucoin) GetHFCompletedOrderList(ctx context.Context, symbol, side, orderType, lastID string, startAt, endAt time.Time, limit int64) (*CompletedHFOrder, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	if side != "" {
		params.Set("side", side)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if lastID == "" {
		params.Set("lastId", lastID)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *CompletedHFOrder
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfCompletedOrderListEPL, http.MethodGet, common.EncodeURLValues("/v1/hf/orders/done", params), nil, &resp)
}

// GetHFOrderDetailsByOrderID obtain information for a single HF order using the order id.
// If the order is not an active order, you can only get data within the time range of 3 _ 24 hours (ie: from the current time to 3 _ 24 hours ago).
func (ku *Kucoin) GetHFOrderDetailsByOrderID(ctx context.Context, orderID, symbol string) (*HFOrderDetail, error) {
	return ku.getHFOrderDetailsByID(ctx, orderID, symbol, "/v1/hf/orders/")
}

// GetHFOrderDetailsByClientOrderID used to obtain information about a single order using clientOid. If the order does not exist, then there will be a prompt saying that the order does not exist.
func (ku *Kucoin) GetHFOrderDetailsByClientOrderID(ctx context.Context, clientOrderID, symbol string) (*HFOrderDetail, error) {
	return ku.getHFOrderDetailsByID(ctx, clientOrderID, symbol, "/v1/hf/orders/client-order/")
}

func (ku *Kucoin) getHFOrderDetailsByID(ctx context.Context, orderID, symbol, path string) (*HFOrderDetail, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *HFOrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfOrderDetailByOrderIDEPL, http.MethodGet, path+orderID+symbolQuery+symbol, nil, &resp)
}

// AutoCancelHFOrderSetting automatically cancel all orders of the set trading pair after the specified time.
// If this interface is not called again for renewal or cancellation before the set time,
// the system will help the user to cancel the order of the corresponding trading pair. Otherwise it will not.
func (ku *Kucoin) AutoCancelHFOrderSetting(ctx context.Context, timeout int64, symbols []string) (*AutoCancelHFOrderResponse, error) {
	if timeout == 0 {
		return nil, errors.New("timeout values required")
	}
	arg := make(map[string]interface{})
	arg["timeout"] = timeout
	if len(symbols) != 0 {
		arg["symbols"] = symbols
	}
	var resp *AutoCancelHFOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, autoCancelHFOrderSettingEPL, http.MethodPost, "/v1/hf/orders/dead-cancel-all", arg, &resp)
}

// AutoCancelHFOrderSettingQuery query the settings of automatic order cancellation
func (ku *Kucoin) AutoCancelHFOrderSettingQuery(ctx context.Context) (*AutoCancelHFOrderResponse, error) {
	var resp *AutoCancelHFOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, autoCancelHFOrderSettingQueryEPL, http.MethodGet, "/v1/hf/orders/dead-cancel-all/query", nil, &resp)
}

// GetHFFilledList retrievesa list of the latest HF transaction details. The returned results are paginated. The data is sorted in descending order according to time.
func (ku *Kucoin) GetHFFilledList(ctx context.Context, orderID, symbol, side, orderType, lastID string, startAt, endAt time.Time, limit int64) (*HFOrderFills, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	if orderID != "" {
		params.Set("orderId", orderID)
	}
	if side != "" {
		params.Set("side", side)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if lastID != "" {
		params.Set("lastId", lastID)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *HFOrderFills
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfFilledListEPL, http.MethodGet, common.EncodeURLValues("/v1/hf/fills", params), nil, &resp)
}

// PostOrder used to place two types of orders: limit and market
// Note: use this only for SPOT trades
func (ku *Kucoin) PostOrder(ctx context.Context, arg *SpotOrderParam) (string, error) {
	return ku.handlePostOrder(ctx, arg, "/v1/orders")
}

// PostOrderTest used to verify whether the signature is correct and other operations.
// After placing an order, the order will not enter the matching system, and the order cannot be queried.
func (ku *Kucoin) PostOrderTest(ctx context.Context, arg *SpotOrderParam) (string, error) {
	return ku.handlePostOrder(ctx, arg, "/v1/orders/test")
}

func (ku *Kucoin) handlePostOrder(ctx context.Context, arg *SpotOrderParam, path string) (string, error) {
	if arg.ClientOrderID == "" {
		// NOTE: 128 bit max length character string. UUID recommended.
		return "", errInvalidClientOrderID
	}
	if arg.Side == "" {
		return "", order.ErrSideIsInvalid
	}
	if arg.Symbol.IsEmpty() {
		return "", fmt.Errorf("%w, empty symbol", currency.ErrCurrencyPairEmpty)
	}
	switch arg.OrderType {
	case "limit", "":
		if arg.Price <= 0 {
			return "", fmt.Errorf("%w, price =%.3f", errInvalidPrice, arg.Price)
		}
		if arg.Size <= 0 {
			return "", errInvalidSize
		}
		if arg.VisibleSize < 0 {
			return "", fmt.Errorf("%w, visible size must be non-zero positive value", errInvalidSize)
		}
	case "market":
		if arg.Size == 0 && arg.Funds == 0 {
			return "", errSizeOrFundIsRequired
		}
	default:
		return "", fmt.Errorf("%w %s", order.ErrTypeIsInvalid, arg.OrderType)
	}
	var resp struct {
		Data struct {
			OrderID string `json:"orderId"`
		} `json:"data"`
		Error
	}
	return resp.Data.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeOrderEPL, http.MethodPost, path, &arg, &resp)
}

// PostMarginOrderTest a test endpoint used to place two types of margin orders: limit and margin.
func (ku *Kucoin) PostMarginOrderTest(ctx context.Context, arg *MarginOrderParam) (*PostMarginOrderResp, error) {
	return ku.postMarginOrder(ctx, arg, "/v1/margin/order/test")
}

// PostMarginOrder used to place two types of margin orders: limit and market
func (ku *Kucoin) PostMarginOrder(ctx context.Context, arg *MarginOrderParam) (*PostMarginOrderResp, error) {
	return ku.postMarginOrder(ctx, arg, "/v1/margin/order")
}

func (ku *Kucoin) postMarginOrder(ctx context.Context, arg *MarginOrderParam, path string) (*PostMarginOrderResp, error) {
	if arg.ClientOrderID == "" {
		return nil, errInvalidClientOrderID
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	if arg.Symbol.IsEmpty() {
		return nil, fmt.Errorf("%w, empty symbol", currency.ErrCurrencyPairEmpty)
	}
	arg.OrderType = strings.ToLower(arg.OrderType)
	switch arg.OrderType {
	case "limit", "":
		if arg.Price <= 0 {
			return nil, fmt.Errorf("%w, price=%.3f", errInvalidPrice, arg.Price)
		}
		if arg.Size <= 0 {
			return nil, errInvalidSize
		}
		if arg.VisibleSize < 0 {
			return nil, fmt.Errorf("%w, visible size must be non-zero positive value", errInvalidSize)
		}
	case "market":
		sum := arg.Size + arg.Funds
		if sum <= 0 || (sum != arg.Size && sum != arg.Funds) {
			return nil, fmt.Errorf("%w, either 'size' or 'funds' has to be set, but not both", errSizeOrFundIsRequired)
		}
	default:
		return nil, fmt.Errorf("%w %s", order.ErrTypeIsInvalid, arg.OrderType)
	}
	resp := struct {
		PostMarginOrderResp
		Error
	}{}
	return &resp.PostMarginOrderResp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeMarginOrdersEPL, http.MethodPost, path, &arg, &resp)
}

// PostBulkOrder used to place 5 orders at the same time. The order type must be a limit order of the same symbol
// Note: it supports only SPOT trades
// Note: To check if order was posted successfully, check status field in response
func (ku *Kucoin) PostBulkOrder(ctx context.Context, symbol string, orderList []OrderRequest) ([]PostBulkOrderResp, error) {
	if symbol == "" {
		return nil, errors.New("symbol can not be empty")
	}
	for i := range orderList {
		if orderList[i].ClientOID == "" {
			return nil, errors.New("clientOid can not be empty")
		}
		if orderList[i].Side == "" {
			return nil, errors.New("side can not be empty")
		}
		if orderList[i].Price <= 0 {
			return nil, errors.New("price must be positive")
		}
		if orderList[i].Size <= 0 {
			return nil, errors.New("size must be positive")
		}
	}
	arg := &struct {
		Symbol    string         `json:"symbol"`
		OrderList []OrderRequest `json:"orderList"`
	}{
		Symbol:    symbol,
		OrderList: orderList,
	}
	resp := &struct {
		Data []PostBulkOrderResp `json:"data"`
	}{}
	return resp.Data, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeBulkOrdersEPL, http.MethodPost, "/v1/orders/multi", arg, &resp)
}

// CancelSingleOrder used to cancel single order previously placed
func (ku *Kucoin) CancelSingleOrder(ctx context.Context, orderID string) ([]string, error) {
	if orderID == "" {
		return nil, errors.New("orderID can not be empty")
	}
	resp := struct {
		CancelledOrderIDs []string `json:"cancelledOrderIds"`
		Error
	}{}
	return resp.CancelledOrderIDs, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelOrderEPL, http.MethodDelete, "/v1/orders/"+orderID, nil, &resp)
}

// CancelOrderByClientOID used to cancel order via the clientOid
func (ku *Kucoin) CancelOrderByClientOID(ctx context.Context, orderID string) (*CancelOrderResponse, error) {
	var resp *CancelOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelOrderByClientOrderIDEPL, http.MethodDelete, "/v1/order/client-order/"+orderID, nil, &resp)
}

// CancelAllOpenOrders used to cancel all order based upon the parameters passed
func (ku *Kucoin) CancelAllOpenOrders(ctx context.Context, symbol, tradeType string) ([]string, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if tradeType != "" {
		params.Set("tradeType", tradeType)
	}
	resp := struct {
		CancelledOrderIDs []string `json:"cancelledOrderIds"`
		Error
	}{}
	return resp.CancelledOrderIDs, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelAllOrdersEPL, http.MethodDelete, common.EncodeURLValues("/v1/orders", params), nil, &resp)
}

// ListOrders gets the user order list
func (ku *Kucoin) ListOrders(ctx context.Context, status, symbol, side, orderType, tradeType string, startAt, endAt time.Time) (*OrdersListResponse, error) {
	params := fillParams(symbol, side, orderType, tradeType, startAt, endAt)
	if status != "" {
		params.Set("status", status)
	}
	var resp *OrdersListResponse
	err := ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, listOrdersEPL, http.MethodGet, common.EncodeURLValues("/v1/orders", params), nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func fillParams(symbol, side, orderType, tradeType string, startAt, endAt time.Time) url.Values {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if side != "" {
		params.Set("side", side)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if tradeType != "" {
		params.Set("tradeType", tradeType)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	return params
}

// GetRecentOrders get orders in the last 24 hours.
func (ku *Kucoin) GetRecentOrders(ctx context.Context) ([]OrderDetail, error) {
	var resp []OrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, recentOrdersEPL, http.MethodGet, "/v1/limit/orders", nil, &resp)
}

// GetOrderByID get a single order info by order ID
func (ku *Kucoin) GetOrderByID(ctx context.Context, orderID string) (*OrderDetail, error) {
	if orderID == "" {
		return nil, errors.New("orderID can not be empty")
	}
	var resp *OrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, orderDetailByIDEPL, http.MethodGet, "/v1/orders/"+orderID, nil, &resp)
}

// GetOrderByClientSuppliedOrderID get a single order info by client order ID
func (ku *Kucoin) GetOrderByClientSuppliedOrderID(ctx context.Context, clientOID string) (*OrderDetail, error) {
	if clientOID == "" {
		return nil, errors.New("client order ID can not be empty")
	}
	var resp *OrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getOrderByClientSuppliedOrderIDEPL, http.MethodGet, "/v1/order/client-order/"+clientOID, nil, &resp)
}

// GetFills get fills
func (ku *Kucoin) GetFills(ctx context.Context, orderID, symbol, side, orderType, tradeType string, startAt, endAt time.Time) (*ListFills, error) {
	params := fillParams(symbol, side, orderType, tradeType, startAt, endAt)
	if orderID != "" {
		params.Set("orderId", orderID)
	}
	var resp *ListFills
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, listFillsEPL, http.MethodGet, common.EncodeURLValues("/v1/fills", params), nil, &resp)
}

// GetRecentFills get a list of 1000 fills in last 24 hours
func (ku *Kucoin) GetRecentFills(ctx context.Context) ([]Fill, error) {
	var resp []Fill
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getRecentFillsEPL, http.MethodGet, "/v1/limit/fills", nil, &resp)
}

// PostStopOrder used to place two types of stop orders: limit and market
func (ku *Kucoin) PostStopOrder(ctx context.Context, clientOID, side, symbol, orderType, remark, stop, stp,
	tradeType, timeInForce string, size, price, stopPrice, cancelAfter, visibleSize,
	funds float64, postOnly, hidden, iceberg bool) (string, error) {
	if clientOID == "" {
		return "", errors.New("clientOid can not be empty")
	}
	if side == "" {
		return "", errors.New("side can not be empty")
	}
	if symbol == "" {
		return "", fmt.Errorf("%w, empty symbol", currency.ErrCurrencyPairEmpty)
	}
	arg := make(map[string]interface{})
	arg["clientOid"] = clientOID
	arg["side"] = side
	arg["symbol"] = symbol
	if remark != "" {
		arg["remark"] = remark
	}
	if stop != "" {
		arg["stop"] = stop
		if stopPrice <= 0 {
			return "", errors.New("stopPrice is required")
		}
		arg["stopPrice"] = strconv.FormatFloat(stopPrice, 'f', -1, 64)
	}
	if stp != "" {
		arg["stp"] = stp
	}
	if tradeType != "" {
		arg["tradeType"] = tradeType
	}
	orderType = strings.ToLower(orderType)
	switch orderType {
	case "limit", "":
		if price <= 0 {
			return "", errors.New("price is required")
		}
		arg["price"] = strconv.FormatFloat(price, 'f', -1, 64)
		if size <= 0 {
			return "", errors.New("size can not be zero or negative")
		}
		arg["size"] = strconv.FormatFloat(size, 'f', -1, 64)
		if timeInForce != "" {
			arg["timeInForce"] = timeInForce
		}
		if cancelAfter > 0 && timeInForce == "GTT" {
			arg["cancelAfter"] = strconv.FormatFloat(cancelAfter, 'f', -1, 64)
		}
		arg["postOnly"] = postOnly
		arg["hidden"] = hidden
		arg["iceberg"] = iceberg
		if visibleSize > 0 {
			arg["visibleSize"] = strconv.FormatFloat(visibleSize, 'f', -1, 64)
		}
	case "market":
		switch {
		case size > 0:
			arg["size"] = strconv.FormatFloat(size, 'f', -1, 64)
		case funds > 0:
			arg["funds"] = strconv.FormatFloat(funds, 'f', -1, 64)
		default:
			return "", errSizeOrFundIsRequired
		}
	default:
		return "", fmt.Errorf("%w, order type: %s", order.ErrTypeIsInvalid, orderType)
	}
	if orderType != "" {
		arg["type"] = orderType
	}
	resp := struct {
		OrderID string `json:"orderId"`
		Error
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeStopOrderEPL, http.MethodPost, "/v1/stop-order", arg, &resp)
}

// CancelStopOrder used to cancel single stop order previously placed
func (ku *Kucoin) CancelStopOrder(ctx context.Context, orderID string) ([]string, error) {
	if orderID == "" {
		return nil, errors.New("orderID can not be empty")
	}
	resp := struct {
		Data []string `json:"cancelledOrderIds"`
		Error
	}{}
	return resp.Data, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelStopOrderEPL, http.MethodDelete, "/v1/stop-order/"+orderID, nil, &resp)
}

// CancelStopOrders used to cancel all order based upon the parameters passed
func (ku *Kucoin) CancelStopOrders(ctx context.Context, symbol, tradeType, orderIDs string) ([]string, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if tradeType != "" {
		params.Set("tradeType", tradeType)
	}
	if orderIDs != "" {
		params.Set("orderIds", orderIDs)
	}
	resp := struct {
		CancelledOrderIDs []string `json:"cancelledOrderIds"`
		Error
	}{}
	return resp.CancelledOrderIDs, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelStopOrdersEPL, http.MethodDelete, common.EncodeURLValues("/v1/stop-order/cancel", params), nil, &resp)
}

// GetStopOrder used to cancel single stop order previously placed
func (ku *Kucoin) GetStopOrder(ctx context.Context, orderID string) (*StopOrder, error) {
	if orderID == "" {
		return nil, errors.New("orderID can not be empty")
	}
	resp := struct {
		StopOrder
		Error
	}{}
	return &resp.StopOrder, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getStopOrderDetailEPL, http.MethodGet, "/v1/stop-order/"+orderID, nil, &resp)
}

// ListStopOrders get all current untriggered stop orders
func (ku *Kucoin) ListStopOrders(ctx context.Context, symbol, side, orderType, tradeType, orderIDs string, startAt, endAt time.Time, currentPage, pageSize int64) (*StopOrderListResponse, error) {
	params := fillParams(symbol, side, orderType, tradeType, startAt, endAt)
	if orderIDs != "" {
		params.Set("orderIds", orderIDs)
	}
	if currentPage != 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize != 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *StopOrderListResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, listStopOrdersEPL, http.MethodGet, common.EncodeURLValues("/v1/stop-order", params), nil, &resp)
}

// GetStopOrderByClientID get a stop order information via the clientOID
func (ku *Kucoin) GetStopOrderByClientID(ctx context.Context, symbol, clientOID string) ([]StopOrder, error) {
	if clientOID == "" {
		return nil, errors.New("clientOID can not be empty")
	}
	params := url.Values{}
	params.Set("clientOid", clientOID)
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	var resp []StopOrder
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getStopOrderByClientIDEPL, http.MethodGet, common.EncodeURLValues("/v1/stop-order/queryOrderByClientOid", params), nil, &resp)
}

// CancelStopOrderByClientID used to cancel a stop order via the clientOID.
func (ku *Kucoin) CancelStopOrderByClientID(ctx context.Context, symbol, clientOID string) (*CancelOrderResponse, error) {
	if clientOID == "" {
		return nil, errors.New("clientOID can not be empty")
	}
	params := url.Values{}
	params.Set("clientOid", clientOID)
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	var resp *CancelOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelStopOrderByClientIDEPL, http.MethodDelete, common.EncodeURLValues("/v1/stop-order/cancelOrderByClientOid", params), nil, &resp)
}

// ------------------------------------------------ OCO Order -----------------------------------------------------------------

// PlaceOCOOrder creates a new One cancel other(OCO) order.
func (ku *Kucoin) PlaceOCOOrder(ctx context.Context, arg *OCOOrderParams) (string, error) {
	if arg == nil || *arg == (OCOOrderParams{}) {
		return "", common.ErrNilPointer
	}
	if arg.Symbol.IsEmpty() {
		return "", currency.ErrCurrencyPairEmpty
	}
	if arg.Side == "" {
		return "", order.ErrSideIsInvalid
	}
	if arg.Price <= 0 {
		return "", order.ErrPriceBelowMin
	}
	if arg.Size <= 0 {
		return "", errInvalidSize
	}
	if arg.StopPrice <= 0 {
		return "", fmt.Errorf("%w stop price = %f", order.ErrPriceBelowMin, arg.StopPrice)
	}
	if arg.LimitPrice <= 0 {
		return "", fmt.Errorf("%w limit price = %f", order.ErrPriceBelowMin, arg.LimitPrice)
	}
	if arg.ClientOrderID == "" {
		return "", errInvalidClientOrderID
	}
	resp := &struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeOCOOrderEPL, http.MethodPost, "/v3/oco/order", &arg, &resp)
}

// CancelOCOOrderByOrderID cancels a single oco order previously placed by order ID.
func (ku *Kucoin) CancelOCOOrderByOrderID(ctx context.Context, orderID string) (*OCOOrderCancellationResponse, error) {
	return ku.cancelOrderByID(ctx, "/v3/oco/order/", orderID)
}

// CancelOCOOrderByClientOrderID cancels a single oco order previously placed by client order ID.
func (ku *Kucoin) CancelOCOOrderByClientOrderID(ctx context.Context, clientOrderID string) (*OCOOrderCancellationResponse, error) {
	return ku.cancelOrderByID(ctx, "/v3/oco/client-order/", clientOrderID)
}

func (ku *Kucoin) cancelOrderByID(ctx context.Context, path, id string) (*OCOOrderCancellationResponse, error) {
	var resp *OCOOrderCancellationResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelOCOOrderByIDEPL, http.MethodDelete, path+id, nil, &resp)
}

// CancelOCOMultipleOrders batch cancel OCO orders through orderIds.
func (ku *Kucoin) CancelOCOMultipleOrders(ctx context.Context, orderIDs []string, symbol string) (*OCOOrderCancellationResponse, error) {
	params := url.Values{}
	if len(orderIDs) > 0 {
		params.Set("orderIds", strings.Join(orderIDs, ","))
	}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	var resp *OCOOrderCancellationResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelMultipleOCOOrdersEPL, http.MethodDelete, common.EncodeURLValues("/v3/oco/orders", params), nil, &resp)
}

// GetOCOOrderInfoByOrderID to get a oco order information via the order ID.
func (ku *Kucoin) GetOCOOrderInfoByOrderID(ctx context.Context, orderID string) (*OCOOrderInfo, error) {
	return ku.getOrderInfoByID(ctx, orderID, "/v3/oco/order/")
}

// GetOCOOrderInfoByClientOrderID to get a oco order information via the client order ID.
func (ku *Kucoin) GetOCOOrderInfoByClientOrderID(ctx context.Context, clientOrderID string) (*OCOOrderInfo, error) {
	return ku.getOrderInfoByID(ctx, clientOrderID, "/v3/oco/client-order/")
}

func (ku *Kucoin) getOrderInfoByID(ctx context.Context, id, path string) (*OCOOrderInfo, error) {
	if id == "" {
		return nil, order.ErrOrderIDNotSet
	}
	var resp *OCOOrderInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getOCOOrderByIDEPL, http.MethodGet, path+id, nil, &resp)
}

// GetOCOOrderDetailsByOrderID get a oco order detail via the order ID.
func (ku *Kucoin) GetOCOOrderDetailsByOrderID(ctx context.Context, orderID string) (*OCOOrderDetail, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	var resp *OCOOrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getOCOOrderDetailsByOrderIDEPL, http.MethodGet, "/v3/oco/order/details/"+orderID, nil, &resp)
}

// GetOCOOrderList retrieves list of OCO orders.
func (ku *Kucoin) GetOCOOrderList(ctx context.Context, pageSize, currentPage, symbol string, startAt, endAt time.Time, orderIDs []string) (*OCOOrders, error) {
	if pageSize == "" {
		return nil, errors.New("pageSize cannot be empty")
	}
	if currentPage == "" {
		return nil, errors.New("currentPage cannot be empty")
	}
	params := url.Values{}
	params.Set("pageSize", pageSize)
	params.Set("currentPage", currentPage)
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if len(orderIDs) == 0 {
		params.Set("orderIds", strings.Join(orderIDs, ","))
	}
	var resp *OCOOrders
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getOCOOrdersEPL, http.MethodGet, common.EncodeURLValues("/v3/oco/orders", params), nil, &resp)
}

// ----------------------------------------------------------- Margin HF Trade -------------------------------------------------------------

// PlaceMarginHFOrder used to place cross-margin or isolated-margin high-frequency margin trading
func (ku *Kucoin) PlaceMarginHFOrder(ctx context.Context, arg *PlaceMarginHFOrderParam) (*MarginHFOrderResponse, error) {
	return ku.placeMarginHFOrder(ctx, arg, "/v3/hf/margin/order")
}

// PlaceMarginHFOrderTest used to verify whether the signature is correct and other operations. After placing an order,
// the order will not enter the matching system, and the order cannot be queried.
func (ku *Kucoin) PlaceMarginHFOrderTest(ctx context.Context, arg *PlaceMarginHFOrderParam) (*MarginHFOrderResponse, error) {
	return ku.placeMarginHFOrder(ctx, arg, "/v3/hf/margin/order/test")
}

func (ku *Kucoin) placeMarginHFOrder(ctx context.Context, arg *PlaceMarginHFOrderParam, path string) (*MarginHFOrderResponse, error) {
	if arg == nil || *arg == (PlaceMarginHFOrderParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.ClientOrderID == "" {
		return nil, order.ErrClientOrderIDNotSupported
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	if arg.Symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if arg.Price <= 0 {
		return nil, order.ErrPriceBelowMin
	}
	if arg.Size <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	var resp *MarginHFOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, placeMarginOrderEPL, http.MethodPost, path, arg, &resp)
}

// CancelMarginHFOrderByOrderID cancels a single order by orderId. If the order cannot be canceled (sold or canceled),
// an error message will be returned, and the reason can be obtained according to the returned msg.
func (ku *Kucoin) CancelMarginHFOrderByOrderID(ctx context.Context, orderID, symbol string) (string, error) {
	return ku.cancelMarginHFOrderByID(ctx, orderID, symbol, "/v3/hf/margin/orders/")
}

// CancelMarginHFOrderByClientOrderID to cancel a single order by clientOid.
func (ku *Kucoin) CancelMarginHFOrderByClientOrderID(ctx context.Context, clientOrderID, symbol string) (string, error) {
	return ku.cancelMarginHFOrderByID(ctx, clientOrderID, symbol, "/v3/hf/margin/orders/client-order/")
}

func (ku *Kucoin) cancelMarginHFOrderByID(ctx context.Context, id, symbol, path string) (string, error) {
	if id == "" {
		return "", order.ErrOrderIDNotSet
	}
	if symbol == "" {
		return "", currency.ErrSymbolStringEmpty
	}
	resp := &struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelMarginHFOrderByIDEPL, http.MethodDelete, path+id+symbolQuery+symbol, nil, &resp)
}

// CancelAllMarginHFOrdersBySymbol cancel all open high-frequency Margin orders(orders created through POST /api/v3/hf/margin/order).
// Transaction type: MARGIN_TRADE - cross margin trade, MARGIN_ISOLATED_TRADE - isolated margin trade
func (ku *Kucoin) CancelAllMarginHFOrdersBySymbol(ctx context.Context, symbol, tradeType string) (string, error) {
	if symbol == "" {
		return "", currency.ErrSymbolStringEmpty
	}
	if tradeType == "" {
		return "", errTradeTypeMissing
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("tradeType", tradeType)
	var resp string
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelAllMarginHFOrdersBySymbolEPL, http.MethodDelete, common.EncodeURLValues("/v3/hf/margin/orders", params), nil, &resp)
}

// GetActiveMarginHFOrders retrieves list if active high-frequency margin orders
func (ku *Kucoin) GetActiveMarginHFOrders(ctx context.Context, symbol, tradeType string) ([]HFOrderDetail, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if tradeType != "" {
		params.Set("tradeType", tradeType)
	}
	var resp []HFOrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getActiveMarginHFOrdersEPL, http.MethodGet, common.EncodeURLValues("/v3/hf/margin/orders/active", params), nil, &resp)
}

// GetFilledHFMarginOrders list of filled margin HF orders and returns paginated data.
// The returned data is sorted in descending order based on the latest order update times.
func (ku *Kucoin) GetFilledHFMarginOrders(ctx context.Context, symbol, tradeType, side, orderType string, startAt, endAt time.Time, lastID, limit int64) (*FilledMarginHFOrdersResponse, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if tradeType == "" {
		return nil, errTradeTypeMissing
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("tradeType", tradeType)
	if side != "" {
		params.Set("side", side)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if lastID > 0 {
		params.Set("lastId", strconv.FormatInt(lastID, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *FilledMarginHFOrdersResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getFilledHFMarginOrdersEPL, http.MethodGet, common.EncodeURLValues("/v3/hf/margin/orders/done", params), nil, &resp)
}

// GetMarginHFOrderDetailByOrderID retrieves the detail of a HF margin order by order ID.
func (ku *Kucoin) GetMarginHFOrderDetailByOrderID(ctx context.Context, orderID, symbol string) (*HFOrderDetail, error) {
	return ku.getMarginHFOrderDetailByID(ctx, orderID, symbol, "/v3/hf/margin/orders/")
}

// GetMarginHFOrderDetailByClientOrderID retrieves the detaul of a HF margin order by client order ID.
func (ku *Kucoin) GetMarginHFOrderDetailByClientOrderID(ctx context.Context, clientOrderID, symbol string) (*HFOrderDetail, error) {
	return ku.getMarginHFOrderDetailByID(ctx, clientOrderID, symbol, "/v3/hf/margin/orders/client-order/")
}

func (ku *Kucoin) getMarginHFOrderDetailByID(ctx context.Context, orderID, symbol, path string) (*HFOrderDetail, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)
	var resp *HFOrderDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getMarginHFOrderDetailByOrderIDEPL, http.MethodGet, path+orderID+symbolQuery+symbol, nil, &resp)
}

// GetMarginHFTradeFills to obtain a list of the latest margin HF transaction details. The returned results are paginated. The data is sorted in descending order according to time.
func (ku *Kucoin) GetMarginHFTradeFills(ctx context.Context, orderID, symbol, tradeType, side, orderType string, startAt, endAt time.Time, lastID, limit int64) (*HFMarginOrderTransaction, error) {
	if tradeType == "" {
		return nil, errTransferTypeMissing
	}
	params := url.Values{}
	params.Set("tradeType", tradeType)
	if orderID != "" {
		params.Set("orderId", orderID)
	}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if side != "" {
		params.Set("side", side)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if lastID > 0 {
		params.Set("lastId", strconv.FormatInt(lastID, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *HFMarginOrderTransaction
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getMarginHFTradeFillsEPL, http.MethodGet, common.EncodeURLValues("/v3/hf/margin/fills", params), nil, &resp)
}

// CreateSubUser creates a new sub-user for the account.
func (ku *Kucoin) CreateSubUser(ctx context.Context, subAccountName, password, remarks, access string) (*SubAccount, error) {
	if regexp.MustCompile("^[a-zA-Z0-9]{7-32}$").MatchString(subAccountName) {
		return nil, errors.New("invalid sub-account name")
	}
	if regexp.MustCompile("^[a-zA-Z0-9]{7-24}$").MatchString(password) {
		return nil, errInvalidPassPhraseInstance
	}
	params := make(map[string]interface{})
	params["subName"] = subAccountName
	params["password"] = password
	if remarks != "" {
		params["remarks"] = remarks
	}
	if access != "" {
		params["access"] = access
	}
	var resp *SubAccount
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, createSubUserEPL, http.MethodPost, "/v2/sub/user/created", params, &resp)
}

// GetSubAccountSpotAPIList used to obtain a list of Spot APIs pertaining to a sub-account.
func (ku *Kucoin) GetSubAccountSpotAPIList(ctx context.Context, subAccountName, apiKeys string) ([]SpotAPISubAccount, error) {
	if subAccountRegExp.MatchString(subAccountName) {
		return nil, errInvalidSubAccountName
	}
	params := url.Values{}
	params.Set("subName", subAccountName)
	if apiKeys != "" {
		params.Set("apiKey", apiKeys)
	}
	var resp []SpotAPISubAccount
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, subAccountSpotAPIListEPL, http.MethodGet, common.EncodeURLValues("/v1/sub/api-key", params), nil, &resp)
}

// CreateSpotAPIsForSubAccount can be used to create Spot APIs for sub-accounts.
func (ku *Kucoin) CreateSpotAPIsForSubAccount(ctx context.Context, arg *SpotAPISubAccountParams) (*SpotAPISubAccount, error) {
	if subAccountRegExp.MatchString(arg.SubAccountName) {
		return nil, errInvalidSubAccountName
	}
	if subAccountPassphraseRegExp.MatchString(arg.Passphrase) {
		return nil, fmt.Errorf("%w, must contain 7-32 characters. cannot contain any spaces", errInvalidPassPhraseInstance)
	}
	if arg.Remark == "" {
		return nil, errors.New("a remarks with a 24 characters max length is required")
	}
	var resp *SpotAPISubAccount
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, createSpotAPIForSubAccountEPL, http.MethodPost, "/v1/sub/api-key", &arg, &resp)
}

// ModifySubAccountSpotAPIs modifies sub-account Spot APIs.
func (ku *Kucoin) ModifySubAccountSpotAPIs(ctx context.Context, arg *SpotAPISubAccountParams) (*SpotAPISubAccount, error) {
	if subAccountRegExp.MatchString(arg.SubAccountName) {
		return nil, errInvalidSubAccountName
	}
	if arg.APIKey == "" {
		return nil, errAPIKeyRequired
	}
	if subAccountPassphraseRegExp.MatchString(arg.Passphrase) {
		return nil, fmt.Errorf("%w, must contain 7-32 characters. cannot contain any spaces", errInvalidPassPhraseInstance)
	}
	var resp *SpotAPISubAccount
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, modifySubAccountSpotAPIEPL, http.MethodPut, "/v1/sub/api-key/update", &arg, &resp)
}

// DeleteSubAccountSpotAPI delete sub-account Spot APIs.
func (ku *Kucoin) DeleteSubAccountSpotAPI(ctx context.Context, apiKey, subAccountName, passphrase string) (*DeleteSubAccountResponse, error) {
	if subAccountRegExp.MatchString(subAccountName) {
		return nil, errInvalidSubAccountName
	}
	if apiKey == "" {
		return nil, errors.New("apiKey is required")
	}
	if passphrase == "" {
		return nil, errInvalidPassPhraseInstance
	}
	params := url.Values{}
	params.Set("apiKey", apiKey)
	params.Set("subName", subAccountName)
	params.Set("passphrase", passphrase)
	var resp *DeleteSubAccountResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, deleteSubAccountSpotAPIEPL, http.MethodDelete, common.EncodeURLValues("/v1/sub/api-key", params), nil, &resp)
}

// GetUserInfoOfAllSubAccounts get the user info of all sub-users via this interface.
func (ku *Kucoin) GetUserInfoOfAllSubAccounts(ctx context.Context) (*SubAccountResponse, error) {
	var resp *SubAccountResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allUserSubAccountsV2EPL, http.MethodGet, "/v2/sub/user", nil, &resp)
}

// GetPaginatedListOfSubAccounts to retrieve a paginated list of sub-accounts. Pagination is required.
func (ku *Kucoin) GetPaginatedListOfSubAccounts(ctx context.Context, currentPage, pageSize int64) (*SubAccountResponse, error) {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	if currentPage > 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	var resp *SubAccountResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, modifySubAccountAPIEPL, http.MethodGet, common.EncodeURLValues("/v1/sub/api-key/update", params), nil, &resp)
}

// GetAllAccounts get all accounts
// accountType possible values are main、trade、margin、trade_hf
func (ku *Kucoin) GetAllAccounts(ctx context.Context, ccy, accountType string) ([]AccountInfo, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if accountType != "" {
		params.Set("type", accountType)
	}
	var resp []AccountInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allAccountEPL, http.MethodGet, common.EncodeURLValues("/v1/accounts", params), nil, &resp)
}

// GetAccountDetail get information of single account
func (ku *Kucoin) GetAccountDetail(ctx context.Context, accountID string) (*AccountInfo, error) {
	if accountID == "" {
		return nil, errAccountIDMissing
	}
	var resp *AccountInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, accountDetailEPL, http.MethodGet, "/v1/accounts/"+accountID, nil, &resp)
}

// GetCrossMarginAccountsDetail retrieves the info of the cross margin account.
func (ku *Kucoin) GetCrossMarginAccountsDetail(ctx context.Context, quoteCurrency, queryType string) (*CrossMarginAccountDetail, error) {
	params := url.Values{}
	if quoteCurrency != "" {
		params.Set("quoteCurrency", quoteCurrency)
	}
	if queryType != "" {
		params.Set("queryType", queryType)
	}
	var resp *CrossMarginAccountDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, crossMarginAccountsDetailEPL, http.MethodGet, common.EncodeURLValues("/v3/margin/accounts", params), nil, &resp)
}

// GetIsolatedMarginAccountDetail to get the info of the isolated margin account.
func (ku *Kucoin) GetIsolatedMarginAccountDetail(ctx context.Context, symbol, queryCurrency, queryType string) (*IsolatedMarginAccountDetail, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if queryCurrency != "" {
		params.Set("quoteCurrency", queryCurrency)
	}
	if queryType != "" {
		params.Set("queryType", queryType)
	}
	var resp *IsolatedMarginAccountDetail
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, isolatedMarginAccountDetailEPL, http.MethodGet, common.EncodeURLValues("/v3/isolated/accounts", params), nil, &resp)
}

// GetFuturesAccountDetail retrieves futures account detail information
func (ku *Kucoin) GetFuturesAccountDetail(ctx context.Context, ccy string) (*FuturesAccountOverview, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	var resp *FuturesAccountOverview
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, futuresAccountsDetailEPL, http.MethodGet, common.EncodeURLValues("/v1/account-overview", params), nil, &resp)
}

// GetSubAccounts retrieves all sub-account information
func (ku *Kucoin) GetSubAccounts(ctx context.Context, subUserID string, includeBaseAmount bool) (*SubAccounts, error) {
	if subUserID == "" {
		return nil, errors.New("sub users ID is required")
	}
	params := url.Values{}
	if includeBaseAmount {
		params.Set("includeBaseAmount", "true")
	} else {
		params.Set("includeBaseAmount", "false")
	}
	var resp *SubAccounts
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, subAccountsEPL, http.MethodGet, common.EncodeURLValues("/v1/sub-accounts/"+subUserID, params), nil, &resp)
}

// GetAllFuturesSubAccountBalances retrieves all futures subaccount balances
func (ku *Kucoin) GetAllFuturesSubAccountBalances(ctx context.Context, ccy string) (*FuturesSubAccountBalance, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	var resp *FuturesSubAccountBalance
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allFuturesSubAccountBalancesEPL, http.MethodGet, common.EncodeURLValues("/v1/account-overview-all", params), nil, &resp)
}

func populateParams(ccy, direction, bizType string, lastID, limit int64, startTime, endTime time.Time) url.Values {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if bizType != "" {
		params.Set("bizType", bizType)
	}
	if lastID != 0 {
		params.Set("lastId", strconv.FormatInt(lastID, 10))
	}
	if limit != 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	return params
}

// GetAccountLedgers retrieves the transaction records from all types of your accounts, supporting inquiry of various currencies.
// bizType possible values: 'DEPOSIT' -deposit, 'WITHDRAW' -withdraw, 'TRANSFER' -transfer, 'SUB_TRANSFER' -subaccount transfer,'TRADE_EXCHANGE' -trade, 'MARGIN_EXCHANGE' -margin trade, 'KUCOIN_BONUS' -bonus
func (ku *Kucoin) GetAccountLedgers(ctx context.Context, ccy, direction, bizType string, startAt, endAt time.Time) (*AccountLedgerResponse, error) {
	params := populateParams(ccy, direction, bizType, 0, 0, time.Time{}, time.Time{})
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	var resp *AccountLedgerResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, accountLedgersEPL, http.MethodGet, common.EncodeURLValues("/v1/accounts/ledgers", params), nil, &resp)
}

// GetAccountLedgersHFTrade returns all transfer (in and out) records in high-frequency trading account and supports multi-coin queries.
// The query results are sorted in descending order by createdAt and id.
func (ku *Kucoin) GetAccountLedgersHFTrade(ctx context.Context, ccy, direction, bizType string, lastID, limit int64, startTime, endTime time.Time) ([]LedgerInfo, error) {
	params := populateParams(ccy, direction, bizType, lastID, limit, startTime, endTime)
	var resp []LedgerInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfAccountLedgersEPL, http.MethodGet, common.EncodeURLValues("/v1/hf/accounts/ledgers", params), nil, &resp)
}

// GetAccountLedgerHFMargin returns all transfer (in and out) records in high-frequency margin trading account and supports multi-coin queries.
func (ku *Kucoin) GetAccountLedgerHFMargin(ctx context.Context, ccy, direction, bizType string, lastID, limit int64, startTime, endTime time.Time) ([]LedgerInfo, error) {
	params := populateParams(ccy, direction, bizType, lastID, limit, startTime, endTime)
	var resp []LedgerInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, hfAccountLedgersMarginEPL, http.MethodGet, common.EncodeURLValues("/v3/hf/margin/account/ledgers", params), nil, &resp)
}

// GetFuturesAccountLedgers If there are open positions, the status of the first page returned will be Pending,
// indicating the realised profit and loss in the current 8-hour settlement period.
// Type RealisedPNL-Realised profit and loss, Deposit-Deposit, Withdrawal-withdraw, Transferin-Transfer in, TransferOut-Transfer out
func (ku *Kucoin) GetFuturesAccountLedgers(ctx context.Context, ccy string, forward bool, startAt, endAt time.Time, offset, maxCount int64) (*FuturesLedgerInfo, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if forward {
		params.Set("forward", "true")
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if offset != 0 {
		params.Set("offset", strconv.FormatInt(offset, 10))
	}
	if maxCount != 0 {
		params.Set("maxCount", strconv.FormatInt(maxCount, 10))
	}
	var resp *FuturesLedgerInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, futuresAccountLedgersEPL, http.MethodGet, common.EncodeURLValues("/v1/transaction-history", params), nil, &resp)
}

// GetAllSubAccountsInfoV1 retrieves the user info of all sub-account via this interface.
func (ku *Kucoin) GetAllSubAccountsInfoV1(ctx context.Context) ([]SubAccount, error) {
	var resp []SubAccount
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, subAccountInfoV1EPL, http.MethodGet, "/v1/sub/user", nil, &resp)
}

// GetAllSubAccountsInfoV2 retrieves list of sub-accounts.
func (ku *Kucoin) GetAllSubAccountsInfoV2(ctx context.Context, currentPage, pageSize int64) (*SubAccountV2Response, error) {
	params := url.Values{}
	if currentPage > 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize > 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *SubAccountV2Response
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allSubAccountsInfoV2EPL, http.MethodGet, "/v2/sub/user", nil, &resp)
}

// GetAccountSummaryInformation this can be used to obtain account summary information.
func (ku *Kucoin) GetAccountSummaryInformation(ctx context.Context) (*AccountSummaryInformation, error) {
	var resp *AccountSummaryInformation
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, accountSummaryInfoEPL, http.MethodGet, "/v2/user-info", nil, &resp)
}

// GetAggregatedSubAccountBalance get the account info of all sub-users
func (ku *Kucoin) GetAggregatedSubAccountBalance(ctx context.Context) ([]SubAccountInfo, error) {
	var resp []SubAccountInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, subAccountBalancesEPL, http.MethodGet, "/v1/sub-accounts", nil, &resp)
}

// GetAllSubAccountsBalanceV2 retrieves sub-account balance information through the V2 API
func (ku *Kucoin) GetAllSubAccountsBalanceV2(ctx context.Context) (*SubAccountBalanceV2, error) {
	var resp *SubAccountBalanceV2
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allSubAccountBalancesV2EPL, http.MethodGet, "/v2/sub-accounts", nil, &resp)
}

// GetPaginatedSubAccountInformation this endpoint can be used to get paginated sub-account information. Pagination is required.
func (ku *Kucoin) GetPaginatedSubAccountInformation(ctx context.Context, currentPage, pageSize int64) ([]SubAccountInfo, error) {
	params := url.Values{}
	if currentPage != 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize != 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp []SubAccountInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, allSubAccountsBalanceEPL, http.MethodGet, common.EncodeURLValues("/v1/sub-accounts", params), nil, &resp)
}

// GetTransferableBalance get the transferable balance of a specified account
// The account type:MAIN、TRADE、TRADE_HF、MARGIN、ISOLATED
func (ku *Kucoin) GetTransferableBalance(ctx context.Context, ccy, accountType, tag string) (*TransferableBalanceInfo, error) {
	if ccy == "" {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if accountType == "" {
		return nil, errors.New("accountType can not be empty")
	}
	params := url.Values{}
	params.Set("currency", ccy)
	params.Set("type", accountType)
	if tag != "" {
		params.Set("tag", tag)
	}
	var resp *TransferableBalanceInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getTransferablesEPL, http.MethodGet, common.EncodeURLValues("/v1/accounts/transferable", params), nil, &resp)
}

// GetUniversalTransfer support transfer between master and sub accounts (only applicable to master account APIKey).
func (ku *Kucoin) GetUniversalTransfer(ctx context.Context, arg *UniversalTransferParam) (string, error) {
	if arg == nil || *arg == (UniversalTransferParam{}) {
		return "", common.ErrNilPointer
	}
	if arg.ClientSuppliedOrderID == "" {
		return "", errInvalidClientOrderID
	}
	if arg.Amount <= 0 {
		return "", order.ErrAmountBelowMin
	}
	if arg.FromAccountType == "" {
		return "", fmt.Errorf("%w, empty fromAccountType", errAccountTypeMissing)
	}
	if arg.TransferType == "" {
		return "", fmt.Errorf("%w, transfer type is empty", errTransferTypeMissing)
	}
	if arg.ToAccountType == "" {
		return "", fmt.Errorf("%w, toAccountType is empty", errAccountTypeMissing)
	}
	var resp string
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, flexiTransferEPL, http.MethodPost, "/v3/accounts/universal-transfer", arg, &resp)
}

// TransferMainToSubAccount used to transfer funds from main account to sub-account
func (ku *Kucoin) TransferMainToSubAccount(ctx context.Context, clientOID, ccy, amount, direction, accountType, subAccountType, subUserID string) (string, error) {
	if clientOID == "" {
		return "", errors.New("clientOID can not be empty")
	}
	if ccy == "" {
		return "", currency.ErrCurrencyPairEmpty
	}
	if amount == "" {
		return "", errors.New("amount can not be empty")
	}
	if direction == "" {
		return "", errors.New("direction can not be empty")
	}
	if subUserID == "" {
		return "", errors.New("subUserID can not be empty")
	}
	params := make(map[string]interface{})
	params["clientOid"] = clientOID
	params["currency"] = ccy
	params["amount"] = amount
	params["direction"] = direction
	if accountType != "" {
		params["accountType"] = accountType
	}
	if subAccountType != "" {
		params["subAccountType"] = subAccountType
	}
	params["subUserId"] = subUserID
	resp := struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, masterSubUserTransferEPL, http.MethodPost, "/v2/accounts/sub-transfer", params, &resp)
}

// MakeInnerTransfer used to transfer funds between accounts internally
func (ku *Kucoin) MakeInnerTransfer(ctx context.Context, clientOID, ccy, from, to, amount, fromTag, toTag string) (string, error) {
	if clientOID == "" {
		return "", errors.New("clientOID can not be empty")
	}
	if ccy == "" {
		return "", currency.ErrCurrencyPairEmpty
	}
	if amount == "" {
		return "", errors.New("amount can not be empty")
	}
	if from == "" {
		return "", errors.New("from can not be empty")
	}
	if to == "" {
		return "", errors.New("to can not be empty")
	}
	params := make(map[string]interface{})
	params["clientOid"] = clientOID
	params["currency"] = ccy
	params["amount"] = amount
	params["from"] = from
	params["to"] = to
	if fromTag != "" {
		params["fromTag"] = fromTag
	}
	if toTag != "" {
		params["toTag"] = toTag
	}
	resp := struct {
		OrderID string `json:"orderId"`
	}{}
	return resp.OrderID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, innerTransferEPL, http.MethodPost, "/v2/accounts/inner-transfer", params, &resp)
}

// TransferToMainOrTradeAccount transfers fund from KuCoin Futures account to Main or Trade accounts.
func (ku *Kucoin) TransferToMainOrTradeAccount(ctx context.Context, arg *FundTransferFuturesParam) (*InnerTransferToMainAndTradeResponse, error) {
	if arg == nil || *arg == (FundTransferFuturesParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Amount <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if arg.RecieveAccountType != "MAIN" && arg.RecieveAccountType != "TRADE" {
		return nil, fmt.Errorf("invalid receive account type %s, only TRADE and MAIN are supported", arg.RecieveAccountType)
	}
	var resp *InnerTransferToMainAndTradeResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, toMainOrTradeAccountEPL, http.MethodPost, "/v3/transfer-out", arg, &resp)
}

// TransferToFuturesAccount transfers fund from KuCoin Futures account to Main or Trade accounts.
func (ku *Kucoin) TransferToFuturesAccount(ctx context.Context, arg *FundTransferToFuturesParam) (*FundTransferToFuturesResponse, error) {
	if arg == nil || *arg == (FundTransferToFuturesParam{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Amount <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if arg.PaymentAccountType != "MAIN" && arg.PaymentAccountType != "TRADE" {
		return nil, fmt.Errorf("invalid receive account type %s, only TRADE and MAIN are supported", arg.PaymentAccountType)
	}
	var resp *FundTransferToFuturesResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, toFuturesAccountEPL, http.MethodPost, "/v1/transfer-in", arg, &resp)
}

// GetFuturesTransferOutRequestRecords retrieves futures transfers out requests.
func (ku *Kucoin) GetFuturesTransferOutRequestRecords(ctx context.Context, startAt, endAt time.Time, status, queryStatus, ccy string, currentPage, pageSize int64) (*FuturesTransferOutResponse, error) {
	params := url.Values{}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if status != "" {
		params.Set("status", status)
	}
	if queryStatus != "" {
		params.Set("queryStatus", queryStatus)
	}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if currentPage != 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize != 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *FuturesTransferOutResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, futuresTransferOutRequestRecordsEPL, http.MethodGet, common.EncodeURLValues("/v1/transfer-list", params), nil, &resp)
}

// CreateDepositAddress create a deposit address for a currency you intend to deposit
func (ku *Kucoin) CreateDepositAddress(ctx context.Context, arg *DepositAddressParams) (*DepositAddress, error) {
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	var resp *DepositAddress
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, createDepositAddressEPL, http.MethodPost, "/v1/deposit-addresses", arg, &resp)
}

// GetDepositAddressesV2 get all deposit addresses for the currency you intend to deposit
func (ku *Kucoin) GetDepositAddressesV2(ctx context.Context, ccy string) ([]DepositAddress, error) {
	if ccy == "" {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	params.Set("currency", ccy)
	var resp []DepositAddress
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, depositAddressesV2EPL, http.MethodGet, common.EncodeURLValues("/v2/deposit-addresses", params), nil, &resp)
}

// GetDepositAddressV1 get a deposit address for the currency you intend to deposit
func (ku *Kucoin) GetDepositAddressV1(ctx context.Context, ccy, chain string) (*DepositAddress, error) {
	if ccy == "" {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	params.Set("currency", ccy)
	if chain != "" {
		params.Set("chain", chain)
	}
	var resp *DepositAddress
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, depositAddressesV1EPL, http.MethodGet, common.EncodeURLValues("/v1/deposit-addresses", params), nil, &resp)
}

// GetDepositList get deposit list items and sorted to show the latest first
// Status. Available value: PROCESSING, SUCCESS, and FAILURE
func (ku *Kucoin) GetDepositList(ctx context.Context, ccy, status string, startAt, endAt time.Time) (*DepositResponse, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if status != "" {
		params.Set("status", status)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	var resp *DepositResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, depositListEPL, http.MethodGet, common.EncodeURLValues("/v1/deposits", params), nil, &resp)
}

// GetHistoricalDepositList get historical deposit list items
func (ku *Kucoin) GetHistoricalDepositList(ctx context.Context, ccy, status string, startAt, endAt time.Time) (*HistoricalDepositWithdrawalResponse, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if status != "" {
		params.Set("status", status)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	var resp *HistoricalDepositWithdrawalResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, historicDepositListEPL, http.MethodGet, common.EncodeURLValues("/v1/hist-deposits", params), nil, &resp)
}

// GetWithdrawalList get withdrawal list items
func (ku *Kucoin) GetWithdrawalList(ctx context.Context, ccy, status string, startAt, endAt time.Time) (*WithdrawalsResponse, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if status != "" {
		params.Set("status", status)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	var resp *WithdrawalsResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, withdrawalListEPL, http.MethodGet, common.EncodeURLValues("/v1/withdrawals", params), nil, &resp)
}

// GetHistoricalWithdrawalList get historical withdrawal list items
func (ku *Kucoin) GetHistoricalWithdrawalList(ctx context.Context, ccy, status string, startAt, endAt time.Time) (*HistoricalDepositWithdrawalResponse, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	if status != "" {
		params.Set("status", status)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	var resp *HistoricalDepositWithdrawalResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, retrieveV1HistoricalWithdrawalListEPL, http.MethodGet, common.EncodeURLValues("/v1/hist-withdrawals", params), nil, &resp)
}

// GetWithdrawalQuotas get withdrawal quota details
func (ku *Kucoin) GetWithdrawalQuotas(ctx context.Context, ccy, chain string) (*WithdrawalQuota, error) {
	if ccy == "" {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	params := url.Values{}
	params.Set("currency", ccy)
	if chain != "" {
		params.Set("chain", chain)
	}
	var resp *WithdrawalQuota
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, withdrawalQuotaEPL, http.MethodGet, common.EncodeURLValues("/v1/withdrawals/quotas", params), nil, &resp)
}

// ApplyWithdrawal create a withdrawal request
// The endpoint was deprecated for futures, please transfer assets from the FUTURES account to the MAIN account first, and then withdraw from the MAIN account
// Withdrawal fee deduct types are: INTERNAL and EXTERNAL
func (ku *Kucoin) ApplyWithdrawal(ctx context.Context, ccy, address, memo, remark, chain, feeDeductType string, isInner bool, amount float64) (string, error) {
	if ccy == "" {
		return "", currency.ErrCurrencyPairEmpty
	}
	if address == "" {
		return "", errors.New("address can not be empty")
	}
	if amount == 0 {
		return "", errors.New("amount can not be empty")
	}
	params := make(map[string]interface{})
	params["currency"] = ccy
	params["address"] = address
	params["amount"] = amount
	if memo != "" {
		params["memo"] = memo
	}
	params["isInner"] = isInner
	if remark != "" {
		params["remark"] = remark
	}
	if chain != "" {
		params["chain"] = chain
	}
	if feeDeductType != "" {
		params["feeDeductType"] = feeDeductType
	}
	resp := struct {
		WithdrawalID string `json:"withdrawalId"`
		Error
	}{}
	return resp.WithdrawalID, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, applyWithdrawalEPL, http.MethodPost, "/v1/withdrawals", params, &resp)
}

// CancelWithdrawal used to cancel a withdrawal request
func (ku *Kucoin) CancelWithdrawal(ctx context.Context, withdrawalID string) error {
	if withdrawalID == "" {
		return errors.New("withdrawal ID is required")
	}
	return ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, cancelWithdrawalsEPL, http.MethodDelete, "/v1/withdrawals/"+withdrawalID, nil, &struct{}{})
}

// GetBasicFee get basic fee rate of users
// Currency type: '0'-crypto currency, '1'-fiat currency. default is '0'-crypto currency
func (ku *Kucoin) GetBasicFee(ctx context.Context, currencyType string) (*Fees, error) {
	params := url.Values{}
	if currencyType != "" {
		params.Set("currencyType", currencyType)
	}
	var resp *Fees
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, basicFeesEPL, http.MethodGet, common.EncodeURLValues("/v1/base-fee", params), nil, &resp)
}

// GetTradingFee get fee rate of trading pairs
// WARNING: There is a limit of 10 currency pairs allowed to be requested per call.
func (ku *Kucoin) GetTradingFee(ctx context.Context, pairs currency.Pairs) ([]Fees, error) {
	if len(pairs) == 0 {
		return nil, currency.ErrCurrencyPairsEmpty
	}
	var resp []Fees
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, tradeFeesEPL, http.MethodGet, "/v1/trade-fees?symbols="+pairs.Upper().Join(), nil, &resp)
}

// ----------------------------------------------------------  Lending Market ----------------------------------------------------------------------------

// GetLendingCurrencyInformation retrieves a lending currency information.
func (ku *Kucoin) GetLendingCurrencyInformation(ctx context.Context, ccy string) ([]LendingCurrencyInfo, error) {
	params := url.Values{}
	if ccy != "" {
		params.Set("currency", ccy)
	}
	var resp []LendingCurrencyInfo
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, lendingCurrencyInfoEPL, http.MethodGet, common.EncodeURLValues("/v3/project/list", params), nil, &resp)
}

// GetInterestRate retrieves the interest rates of the margin lending market over the past 7 days.
func (ku *Kucoin) GetInterestRate(ctx context.Context, ccy currency.Code) ([]InterestRate, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	var resp []InterestRate
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, interestRateEPL, http.MethodGet, "/v3/project/marketInterestRate?currency="+ccy.String(), nil, &resp)
}

// MarginLendingSubscription retrieves margin lending subscription information.
func (ku *Kucoin) MarginLendingSubscription(ctx context.Context, ccy currency.Code, size, interestRate float64) (*OrderNumberResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if size <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if interestRate <= 0 {
		return nil, errMissingInterestRate
	}
	arg := map[string]interface{}{
		"currency":     ccy.String(),
		"size":         size,
		"interestRate": interestRate,
	}
	var resp *OrderNumberResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, marginLendingSubscriptionEPL, http.MethodPost, "/v3/purchase", arg, &resp)
}

// Redemption initiate redemptions of margin lending.
func (ku *Kucoin) Redemption(ctx context.Context, ccy currency.Code, size float64, purchaseOrderNo string) (*OrderNumberResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if size <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if purchaseOrderNo == "" {
		return nil, errMissingPurchaseOrderNumber
	}
	arg := map[string]interface{}{
		"currency":        ccy.String(),
		"size":            size,
		"purchaseOrderNo": purchaseOrderNo,
	}
	var resp *OrderNumberResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, redemptionEPL, http.MethodPost, "/v3/redeem", arg, &resp)
}

// ModifySubscriptionOrder is used to update the interest rates of subscription orders, which will take effect at the beginning of the next hour.
func (ku *Kucoin) ModifySubscriptionOrder(ctx context.Context, ccy currency.Code, purchaseOrderNo string, interestRate float64) (*ModifySubscriptionOrderResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if interestRate <= 0 {
		return nil, errMissingInterestRate
	}
	if purchaseOrderNo == "" {
		return nil, errMissingPurchaseOrderNumber
	}
	arg := map[string]interface{}{
		"currency":        ccy.String(),
		"interestRate":    interestRate,
		"purchaseOrderNo": purchaseOrderNo,
	}
	var resp *ModifySubscriptionOrderResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, modifySubscriptionEPL, http.MethodPost, "/v3/lend/purchase/update", arg, &resp)
}

// GetRedemptionOrders query for the redemption orders.
// Status: DONE-completed; PENDING-settling
func (ku *Kucoin) GetRedemptionOrders(ctx context.Context, ccy currency.Code, redeemOrderNo, status string, currentPage, pageSize int64) (*RedemptionOrdersResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if status == "" {
		return nil, errors.New("status is missing")
	}
	params := url.Values{}
	params.Set("currency", ccy.String())
	params.Set("status", status)
	if redeemOrderNo != "" {
		params.Set("redeemOrderNo", redeemOrderNo)
	}
	if currentPage > 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize > 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *RedemptionOrdersResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getRedemptionOrdersEPL, http.MethodGet, common.EncodeURLValues("/v3/redeem/orders", params), nil, &resp)
}

// GetSubscriptionOrders provides pagination query for the subscription orders.
func (ku *Kucoin) GetSubscriptionOrders(ctx context.Context, ccy currency.Code, purchaseOrderNo, status string, currentPage, pageSize int64) (*SubscriptionOrdersResponse, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if status == "" {
		return nil, errors.New("status is missing")
	}
	params := url.Values{}
	params.Set("currency", ccy.String())
	params.Set("status", status)
	if purchaseOrderNo != "" {
		params.Set("purchaseOrderNo", purchaseOrderNo)
	}
	if currentPage > 0 {
		params.Set("currentPage", strconv.FormatInt(currentPage, 10))
	}
	if pageSize > 0 {
		params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	}
	var resp *SubscriptionOrdersResponse
	return resp, ku.SendAuthHTTPRequest(ctx, exchange.RestSpot, getSubscriptionOrdersEPL, http.MethodGet, common.EncodeURLValues("/v3/purchase/orders", params), nil, &resp)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (ku *Kucoin) SendHTTPRequest(ctx context.Context, ePath exchange.URL, epl request.EndpointLimit, path string, result interface{}) error {
	value := reflect.ValueOf(result)
	if value.Kind() != reflect.Pointer {
		return errInvalidResultInterface
	}
	resp, okay := result.(UnmarshalTo)
	if !okay {
		resp = &Response{Data: result}
	}
	endpointPath, err := ku.API.Endpoints.GetURL(ePath)
	if err != nil {
		return err
	}
	err = ku.SendPayload(ctx, epl, func() (*request.Item, error) {
		return &request.Item{
			Method:        http.MethodGet,
			Path:          endpointPath + path,
			Result:        resp,
			Verbose:       ku.Verbose,
			HTTPDebugging: ku.HTTPDebugging,
			HTTPRecording: ku.HTTPRecording}, nil
	}, request.UnauthenticatedRequest)
	if err != nil {
		return err
	}
	if result == nil {
		return errNoValidResponseFromServer
	}
	return resp.GetError()
}

// SendAuthHTTPRequest sends an authenticated HTTP request
// Request parameters are added to path variable for GET and DELETE request and for other requests its passed in params variable
func (ku *Kucoin) SendAuthHTTPRequest(ctx context.Context, ePath exchange.URL, epl request.EndpointLimit, method, path string, arg, result interface{}) error {
	value := reflect.ValueOf(result)
	if value.Kind() != reflect.Pointer {
		return errInvalidResultInterface
	}
	creds, err := ku.GetCredentials(ctx)
	if err != nil {
		return err
	}
	resp, okay := result.(UnmarshalTo)
	if !okay {
		resp = &Response{Data: result}
	}
	endpointPath, err := ku.API.Endpoints.GetURL(ePath)
	if err != nil {
		return err
	}
	if value.IsNil() || value.Kind() != reflect.Pointer {
		return fmt.Errorf("%w receiver has to be non-nil pointer", errInvalidResponseReceiver)
	}
	err = ku.SendPayload(ctx, epl, func() (*request.Item, error) {
		var (
			body    io.Reader
			payload []byte
		)
		if arg != nil {
			payload, err = json.Marshal(arg)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(payload)
		}
		timeStamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		var signHash, passPhraseHash []byte
		signHash, err = crypto.GetHMAC(crypto.HashSHA256, []byte(timeStamp+method+"/api"+path+string(payload)), []byte(creds.Secret))
		if err != nil {
			return nil, err
		}
		passPhraseHash, err = crypto.GetHMAC(crypto.HashSHA256, []byte(creds.ClientID), []byte(creds.Secret))
		if err != nil {
			return nil, err
		}
		headers := map[string]string{
			"KC-API-KEY":         creds.Key,
			"KC-API-SIGN":        crypto.Base64Encode(signHash),
			"KC-API-TIMESTAMP":   timeStamp,
			"KC-API-PASSPHRASE":  crypto.Base64Encode(passPhraseHash),
			"KC-API-KEY-VERSION": kucoinAPIKeyVersion,
			"Content-Type":       "application/json",
		}
		return &request.Item{
			Method:        method,
			Path:          endpointPath + path,
			Headers:       headers,
			Body:          body,
			Result:        &resp,
			Verbose:       ku.Verbose,
			HTTPDebugging: ku.HTTPDebugging,
			HTTPRecording: ku.HTTPRecording}, nil
	}, request.AuthenticatedRequest)
	if err != nil {
		return err
	}
	if result == nil {
		return errNoValidResponseFromServer
	}
	return resp.GetError()
}

var intervalMap = map[kline.Interval]string{
	kline.OneMin: "1min", kline.ThreeMin: "3min", kline.FiveMin: "5min", kline.FifteenMin: "15min", kline.ThirtyMin: "30min", kline.OneHour: "1hour", kline.TwoHour: "2hour", kline.FourHour: "4hour", kline.SixHour: "6hour", kline.EightHour: "8hour", kline.TwelveHour: "12hour", kline.OneDay: "1day", kline.OneWeek: "1week",
}

func (ku *Kucoin) intervalToString(interval kline.Interval) (string, error) {
	intervalString, okay := intervalMap[interval]
	if okay {
		return intervalString, nil
	}
	return "", kline.ErrUnsupportedInterval
}

func (ku *Kucoin) stringToOrderStatus(status string) (order.Status, error) {
	switch status {
	case "match":
		return order.Filled, nil
	case "open":
		return order.Open, nil
	case "done":
		return order.Closed, nil
	default:
		return order.StringToOrderStatus(status)
	}
}

func (ku *Kucoin) accountTypeToString(a asset.Item) string {
	switch a {
	case asset.Spot:
		return "trade"
	case asset.Margin:
		return "margin"
	case asset.Empty:
		return ""
	default:
		return "main"
	}
}

func (ku *Kucoin) accountToTradeTypeString(a asset.Item, marginMode string) string {
	switch a {
	case asset.Spot:
		return "TRADE"
	case asset.Margin:
		if strings.EqualFold(marginMode, "isolated") {
			return "MARGIN_ISOLATED_TRADE"
		}
		return "MARGIN_TRADE"
	default:
		return ""
	}
}

func (ku *Kucoin) orderTypeToString(orderType order.Type) (string, error) {
	switch orderType {
	case order.AnyType, order.UnknownType:
		return "", nil
	case order.Market, order.Limit:
		return orderType.Lower(), nil
	default:
		return "", order.ErrUnsupportedOrderType
	}
}

func (ku *Kucoin) orderSideString(side order.Side) (string, error) {
	switch {
	case side.IsLong():
		return order.Buy.Lower(), nil
	case side.IsShort():
		return order.Sell.Lower(), nil
	case side == order.AnySide:
		return "", nil
	default:
		return "", fmt.Errorf("%w, side:%s", order.ErrSideIsInvalid, side.String())
	}
}
