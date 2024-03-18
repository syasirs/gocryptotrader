package poloniex

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

const (
	poloniexFuturesAPIURL = "https://futures-api.poloniex.com/api"
)

// GetOpenContractList retrieves the info of all open contracts.
func (p *Poloniex) GetOpenContractList(ctx context.Context) (*Contracts, error) {
	var resp *Contracts
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/contracts/active", &resp)
}

// GetOrderInfoOfTheContract info of the specified contract.
func (p *Poloniex) GetOrderInfoOfTheContract(ctx context.Context, symbol string) (*ContractItem, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *ContractItem
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/contracts/"+symbol, &resp)
}

// GetRealTimeTicker real-time ticker 1.0 includes the last traded price, the last traded size, transaction ID,
// the side of the liquidity taker, the best bid price and size, the best ask price and size as well as the transaction time of the orders.
func (p *Poloniex) GetRealTimeTicker(ctx context.Context, symbol string) (*TickerDetail, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *TickerDetail
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/ticker?symbol="+symbol, &resp)
}

// TestGetRealTimeTickersOfSymbols retrieves real-time tickers includes tickers of all trading symbols.
func (p *Poloniex) TestGetRealTimeTickersOfSymbols(ctx context.Context) ([]TickerInfo, error) {
	var resp []TickerInfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v2/tickers", &resp)
}

// GetFullOrderbookLevel2 retrieves a snapshot of aggregated open orders for a symbol.
// level 2 order book includes all bids and asks (aggregated by price). This level returns only one aggregated size for each price (as if there was only one single order for that price).
func (p *Poloniex) GetFullOrderbookLevel2(ctx context.Context, symbol string) (*Orderbook, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	var resp *Orderbook
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/level2/snapshot", params), &resp)
}

// GetPartialOrderbookLevel2 represents partial snapshot of aggregated open orders for a symbol.
// depth: depth5, depth10, depth20 , depth30 , depth50 or depth100
func (p *Poloniex) GetPartialOrderbookLevel2(ctx context.Context, symbol, depth string) (*Orderbook, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if depth == "" {
		return nil, errors.New("depth is required")
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("depth", depth)
	var resp *Orderbook
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/level2/depth", params), &resp)
}

// Level2PullingMessages if the messages pushed by Websocket are not continuous, you can submit the following request and re-pull the data to ensure that the sequence is not missing.
func (p *Poloniex) Level2PullingMessages(ctx context.Context, symbol string, startSequence, endSequence int64) (*OrderbookChanges, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if startSequence <= 0 {
		return nil, errors.New("start sequence is required")
	}
	if endSequence <= 0 {
		return nil, errors.New("end sequence is required")
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("start", strconv.FormatInt(startSequence, 10))
	params.Set("end", strconv.FormatInt(endSequence, 10))
	var resp *OrderbookChanges
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/level2/message/query", params), &resp)
}

// GetFullOrderBookLevel3 a snapshot of all the open orders for a symbol. The Level 3 order book includes all bids and asks (the data is non-aggregated, and each item means a single order).
// To ensure your local orderbook data is the latest one, please use Websocket incremental feed after retrieving the level 3 snapshot.
func (p *Poloniex) GetFullOrderBookLevel3(ctx context.Context, symbol string) (*Orderbook, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	var resp *Orderbook
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v2/level3/snapshot", &resp)
}

// Level3PullingMessages If the messages pushed by the Websocket is not continuous, you can submit the following request and re-pull the data to ensure that the sequence is not missing.
func (p *Poloniex) Level3PullingMessages(ctx context.Context) (*Level3PullingMessageResponse, error) {
	var resp *Level3PullingMessageResponse
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v2/level3/snapshot", &resp)
}

// ----------------------------------------------------   Historical Data  ---------------------------------------------------------------

// GetTransactionHistory list the last 100 trades for a symbol.
func (p *Poloniex) GetTransactionHistory(ctx context.Context, symbol string) (*TransactionHistory, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *TransactionHistory
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/trade/history?symbol="+symbol, &resp)
}

func (p *Poloniex) populateIndexParams(symbol string, startAt, endAt time.Time, reverse, forward bool, maxCount int64) url.Values {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if !startAt.IsZero() {
		params.Set("startAt", strconv.FormatInt(startAt.UnixMilli(), 10))
	}
	if !endAt.IsZero() {
		params.Set("endAt", strconv.FormatInt(endAt.UnixMilli(), 10))
	}
	if reverse {
		params.Set("reverse", "true")
	}
	if forward {
		params.Set("forward", "true")
	}
	if maxCount > 0 {
		params.Set("maxCount", strconv.FormatInt(maxCount, 10))
	}
	return params
}

// GetInterestRateList retrieves interest rate list.
func (p *Poloniex) GetInterestRateList(ctx context.Context, symbol string, startAt, endAt time.Time, reverse, forward bool, maxCount int64) (*IndexInfo, error) {
	params := p.populateIndexParams(symbol, startAt, endAt, reverse, forward, maxCount)
	var resp *IndexInfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/interest/query", params), &resp)
}

// GetIndexList check index list
func (p *Poloniex) GetIndexList(ctx context.Context, symbol string, startAt, endAt time.Time, reverse, forward bool, maxCount int64) (*IndexInfo, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := p.populateIndexParams(symbol, startAt, endAt, reverse, forward, maxCount)
	var resp *IndexInfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/index/query", params), &resp)
}

// GetCurrentMarkPrice retrieves the current mark price.
func (p *Poloniex) GetCurrentMarkPrice(ctx context.Context, symbol string) (*MarkPriceDetail, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *MarkPriceDetail
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/mark-price/"+symbol+"/current", &resp)
}

// GetPremiumIndex request to get premium index.
func (p *Poloniex) GetPremiumIndex(ctx context.Context, symbol string, startAt, endAt time.Time, reverse, forward bool, maxCount int64) (*IndexInfo, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	params := p.populateIndexParams(symbol, startAt, endAt, reverse, forward, maxCount)
	var resp *IndexInfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, unauthEPL, common.EncodeURLValues("/v1/premium/query", params), &resp)
}

// GetCurrentFundingRate request to check the current mark price.
func (p *Poloniex) GetCurrentFundingRate(ctx context.Context, symbol string) (*FundingRate, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	var resp *FundingRate
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/funding-rate/"+symbol+"/current", &resp)
}

// GetFuturesServerTime get the API server time. This is the Unix timestamp.
func (p *Poloniex) GetFuturesServerTime(ctx context.Context) (*ServerTimeResponse, error) {
	var resp *ServerTimeResponse
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/timestamp", &resp)
}

// GetServiceStatus the service status.
func (p *Poloniex) GetServiceStatus(ctx context.Context) (*ServiceStatus, error) {
	var resp *ServiceStatus
	return resp, p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, "/v1/status", &resp)
}

// GetKlineDataOfContract retrieves candlestick information
func (p *Poloniex) GetKlineDataOfContract(ctx context.Context, symbol string, granularity int64, from, to time.Time) ([]KlineChartData, error) {
	if symbol == "" {
		return nil, currency.ErrSymbolStringEmpty
	}
	if granularity == 0 {
		return nil, errors.New("granularity is required")
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("granularity", strconv.FormatInt(granularity, 10))
	if !from.IsZero() {
		params.Set("from", strconv.FormatInt(from.UnixMilli(), 10))
	}
	if !to.IsZero() {
		params.Set("to", strconv.FormatInt(to.UnixMilli(), 10))
	}
	var resp *KlineChartResponse
	err := p.SendHTTPRequest(ctx, exchange.RestFutures, unauthEPL, common.EncodeURLValues("/v1/kline/query", params), &resp)
	if err != nil {
		return nil, err
	}
	return resp.ExtractKlineChart(), nil
}
