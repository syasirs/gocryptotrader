package lbank

import "encoding/json"

// Ticker stores the ticker price data for a currency pair
type Ticker struct {
	Change   float64 `json:"change"`
	High     float64 `json:"high"`
	Latest   float64 `json:"latest"`
	Low      float64 `json:"low"`
	Turnover float64 `json:"turnover"`
	Volume   float64 `json:"vol"`
}

// TickerResponse stores the ticker price data and timestamp for a currency pair
type TickerResponse struct {
	Symbol    string `json:"symbol"`
	Timestamp int64  `json:"timestamp"`
	Ticker    Ticker `json:"ticker"`
}

// MarketDepthResponse stores arrays for asks, bids and a timestamp for a currecy pair
type MarketDepthResponse struct {
	Asks      [][]float64 `json:"asks"`
	Bids      [][]float64 `json:"bids"`
	Timestamp int64       `json:"timestamp"`
}

// TradeResponse stores date_ms, amount, price, type, tid for a currency pair
type TradeResponse struct {
	DateMS int64   `json:"date_ms"`
	Amount float64 `json:"amount"`
	Price  float64 `json:"price"`
	Type   string  `json:"type"`
	TID    string  `json:"tid"`
}

// KlineResponse stores kline info for given currency exchange
type KlineResponse struct {
	TimeStamp     int64   `json:"timestamp"`
	OpenPrice     float64 `json:"openprice"`
	HigestPrice   float64 `json:"highestprice"`
	LowestPrice   float64 `json:"lowestprice"`
	ClosePrice    float64 `json:"closeprice"`
	TradingVolume float64 `json:"tradingvolume"`
}

// InfoResponse stores info
type InfoResponse struct {
	Freeze map[string]float64 `json:"freeze,string"`
	Asset  map[string]float64 `json:"asset,string"`
	Free   map[string]float64 `json:"free,string"`
}

// CreateOrderResponse stores the result of the Order and
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
}

// RemoveOrderResponse stores the result when an order is cancelled
type RemoveOrderResponse struct {
	OrderID string `json:"order_id"`
	Success string `json:"success"`
	Error   string `json:"error"`
}

// OrderResponse stores the data related to the given OrderIDs
type OrderResponse struct {
	Symbol     string  `json:"symbol"`
	Amount     float64 `json:"amount"`
	CreateTime int64   `json:"create_time"`
	Price      float64 `json:"price"`
	AvgPrice   float64 `json:"avg_price"`
	Type       string  `json:"type"`
	OrderID    string  `json:"order_id"`
	DealAmount float64 `json:"deal_amount"`
	Status     int64   `json:"status"`
}

// QueryOrderResponse stores the data from queries
type QueryOrderResponse struct {
	Result    bool            `json:"result,string"`
	Orders    []OrderResponse `json:"orders"`
	ErrorCode int64           `json:"error_code"`
}

// OrderHistory is bla
type OrderHistory struct {
	Result      bool            `json:"result,string"`
	Total       string          `json:"total"`
	PageLength  int64           `json:"page_length"`
	Orders      json.RawMessage `json:"orders"`
	CurrentPage int64           `json:"current_page"`
	ErrorCode   int64           `json:"error_code"`
}

// OrderHistoryResponse stores past orders
type OrderHistoryResponse struct {
	Result      bool            `json:"result,string"`
	Total       string          `json:"total"`
	PageLength  int64           `json:"page_length"`
	Orders      []OrderResponse `json:"orders"`
	CurrentPage int64           `json:"current_page"`
	ErrorCode   int64           `json:"error_code"`
}

// PairInfoResponse stores information about trading pairs
type PairInfoResponse struct {
	MinimumQuantity  string `json:"minTranQua"`
	PriceAccuracy    string `json:"priceAccuracy"`
	QuantityAccuracy string `json:"quantityAccuracy"`
	Symbol           string `json:"symbol"`
	ErrorCode        int64  `json:"error_code"`
}

// OpenOrderResponse stores information about the opening orders
type OpenOrderResponse struct {
	PageLength int64           `json:"page_length"`
	PageNumber int64           `json:"page_number"`
	Total      string          `json:"total"`
	Result     bool            `json:"result,string"`
	Orders     []OrderResponse `json:"orders"`
	ErrorCode  int64           `json:"error_code"`
}

// ExchangeRateResponse stores information about USD-RMB rate
type ExchangeRateResponse struct {
	USD2CNY string `json:"USD2CNY"`
}

// WithdrawConfigResponse stores info about withdrawal configurations
type WithdrawConfigResponse struct {
	AssetCode   string `json:"assetCode"`
	Minimum     string `json:"min"`
	CanWithDraw bool   `json:"canWithDraw"`
	Fee         string `json:"fee"`
}

// WithdrawResponse stores info about something
type WithdrawResponse struct {
	Result     bool    `json:"result,string"`
	WithdrawID string  `json:"withdrawId"`
	Fee        float64 `json:"fee"`
	ErrorCode  int64   `json:"error_code"`
}

// RevokeWithdrawResponse stores info about something
type RevokeWithdrawResponse struct {
	Result     bool   `json:"result,string"`
	WithdrawID string `json:"string"`
	ErrorCode  int64  `json:"error_code"`
}

// ListDataResponse contains some of withdrawal data
type ListDataResponse struct {
	Amount    float64 `json:"amount"`
	AssetCode string  `json:"assetCode"`
	Address   string  `json:"address"`
	Fee       float64 `json:"fee"`
	ID        int64   `json:"id"`
	Time      int64   `json:"time"`
	TXHash    string  `json:"txhash"`
	Status    string  `json:"status"`
}

// WithdrawalResponse stores data for withdrawals
type WithdrawalResponse struct {
	TotalPages int64              `json:"totalPages"`
	PageSize   int64              `json:"pageSize"`
	PageNo     int64              `json:"pageNo"`
	List       []ListDataResponse `json:"list"`
	ErrorCode  int64              `json:"error_code"`
}

// WsPing sends a ping
type WsPing struct {
	Action string `json:"action"`
	Ping   string `json:"ping"`
}

// Generated by https://quicktype.io
