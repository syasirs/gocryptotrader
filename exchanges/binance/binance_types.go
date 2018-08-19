package binance

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// Response holds basic binance api response data
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ExchangeInfo holds the full exchange information type
type ExchangeInfo struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Timezone   string `json:"timezone"`
	Servertime int64  `json:"serverTime"`
	RateLimits []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		Limit         int    `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters interface{} `json:"exchangeFilters"`
	Symbols         []struct {
		Symbol             string   `json:"symbol"`
		Status             string   `json:"status"`
		BaseAsset          string   `json:"baseAsset"`
		BaseAssetPrecision int      `json:"baseAssetPrecision"`
		QuoteAsset         string   `json:"quoteAsset"`
		QuotePrecision     int      `json:"quotePrecision"`
		OrderTypes         []string `json:"orderTypes"`
		IcebergAllowed     bool     `json:"icebergAllowed"`
		Filters            []struct {
			FilterType  string          `json:"filterType"`
			MinPrice    decimal.Decimal `json:"minPrice,string"`
			MaxPrice    decimal.Decimal `json:"maxPrice,string"`
			TickSize    decimal.Decimal `json:"tickSize,string"`
			MinQty      decimal.Decimal `json:"minQty,string"`
			MaxQty      decimal.Decimal `json:"maxQty,string"`
			StepSize    decimal.Decimal `json:"stepSize,string"`
			MinNotional decimal.Decimal `json:"minNotional,string"`
		} `json:"filters"`
	} `json:"symbols"`
}

// OrderBookDataRequestParams represents Klines request data.
type OrderBookDataRequestParams struct {
	Symbol string `json:"symbol"` // Required field; example LTCBTC,BTCUSDT
	Limit  int    `json:"limit"`  // Default 100; max 1000. Valid limits:[5, 10, 20, 50, 100, 500, 1000]
}

// OrderBookData is resp data from orderbook endpoint
type OrderBookData struct {
	Code         int           `json:"code"`
	Msg          string        `json:"msg"`
	LastUpdateID int64         `json:"lastUpdateId"`
	Bids         []interface{} `json:"bids"`
	Asks         []interface{} `json:"asks"`
}

// OrderBook actual structured data that can be used for orderbook
type OrderBook struct {
	Code int
	Msg  string
	Bids []struct {
		Price    decimal.Decimal
		Quantity decimal.Decimal
	}
	Asks []struct {
		Price    decimal.Decimal
		Quantity decimal.Decimal
	}
}

// RecentTradeRequestParams represents Klines request data.
type RecentTradeRequestParams struct {
	Symbol string `json:"symbol"` // Required field. example LTCBTC, BTCUSDT
	Limit  int    `json:"limit"`  // Default 500; max 500.
}

// RecentTrade holds recent trade data
type RecentTrade struct {
	Code         int             `json:"code"`
	Msg          string          `json:"msg"`
	ID           int64           `json:"id"`
	Price        decimal.Decimal `json:"price,string"`
	Quantity     decimal.Decimal `json:"qty,string"`
	Time         int64           `json:"time"`
	IsBuyerMaker bool            `json:"isBuyerMaker"`
	IsBestMatch  bool            `json:"isBestMatch"`
}

// MultiStreamData holds stream data
type MultiStreamData struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// TradeStream holds the trade stream data
type TradeStream struct {
	EventType      string `json:"e"`
	EventTime      int64  `json:"E"`
	Symbol         string `json:"s"`
	TradeID        int64  `json:"t"`
	Price          string `json:"p"`
	Quantity       string `json:"q"`
	BuyerOrderID   int64  `json:"b"`
	SellerOrderID  int64  `json:"a"`
	TimeStamp      int64  `json:"T"`
	Maker          bool   `json:"m"`
	BestMatchPrice bool   `json:"M"`
}

// KlineStream holds the kline stream data
type KlineStream struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime                int64  `json:"t"`
		CloseTime                int64  `json:"T"`
		Symbol                   string `json:"s"`
		Interval                 string `json:"i"`
		FirstTradeID             int64  `json:"f"`
		LastTradeID              int64  `json:"L"`
		OpenPrice                string `json:"o"`
		ClosePrice               string `json:"c"`
		HighPrice                string `json:"h"`
		LowPrice                 string `json:"l"`
		Volume                   string `json:"v"`
		NumberOfTrades           int64  `json:"n"`
		KlineClosed              bool   `json:"x"`
		Quote                    string `json:"q"`
		TakerBuyBaseAssetVolume  string `json:"V"`
		TakerBuyQuoteAssetVolume string `json:"Q"`
	} `json:"k"`
}

// TickerStream holds the ticker stream data
type TickerStream struct {
	EventType              string `json:"e"`
	EventTime              int64  `json:"E"`
	Symbol                 string `json:"s"`
	PriceChange            string `json:"p"`
	PriceChangePercent     string `json:"P"`
	WeightedAvgPrice       string `json:"w"`
	PrevDayClose           string `json:"x"`
	CurrDayClose           string `json:"c"`
	CloseTradeQuantity     string `json:"Q"`
	BestBidPrice           string `json:"b"`
	BestBidQuantity        string `json:"B"`
	BestAskPrice           string `json:"a"`
	BestAskQuantity        string `json:"A"`
	OpenPrice              string `json:"o"`
	HighPrice              string `json:"h"`
	LowPrice               string `json:"l"`
	TotalTradedVolume      string `json:"v"`
	TotalTradedQuoteVolume string `json:"q"`
	OpenTime               int64  `json:"O"`
	CloseTime              int64  `json:"C"`
	FirstTradeID           int64  `json:"F"`
	LastTradeID            int64  `json:"L"`
	NumberOfTrades         int64  `json:"n"`
}

// HistoricalTrade holds recent trade data
type HistoricalTrade struct {
	Code         int             `json:"code"`
	Msg          string          `json:"msg"`
	ID           int64           `json:"id"`
	Price        decimal.Decimal `json:"price,string"`
	Quantity     decimal.Decimal `json:"qty,string"`
	Time         int64           `json:"time"`
	IsBuyerMaker bool            `json:"isBuyerMaker"`
	IsBestMatch  bool            `json:"isBestMatch"`
}

// AggregatedTrade holds aggregated trade information
type AggregatedTrade struct {
	ATradeID       int64           `json:"a"`
	Price          decimal.Decimal `json:"p,string"`
	Quantity       decimal.Decimal `json:"q,string"`
	FirstTradeID   int64           `json:"f"`
	LastTradeID    int64           `json:"l"`
	TimeStamp      int64           `json:"T"`
	Maker          bool            `json:"m"`
	BestMatchPrice bool            `json:"M"`
}

// CandleStick holds kline data
type CandleStick struct {
	OpenTime                 decimal.Decimal
	Open                     decimal.Decimal
	High                     decimal.Decimal
	Low                      decimal.Decimal
	Close                    decimal.Decimal
	Volume                   decimal.Decimal
	CloseTime                decimal.Decimal
	QuoteAssetVolume         decimal.Decimal
	TradeCount               decimal.Decimal
	TakerBuyAssetVolume      decimal.Decimal
	TakerBuyQuoteAssetVolume decimal.Decimal
}

// PriceChangeStats contains statistics for the last 24 hours trade
type PriceChangeStats struct {
	Symbol             string          `json:"symbol"`
	PriceChange        decimal.Decimal `json:"priceChange,string"`
	PriceChangePercent decimal.Decimal `json:"priceChangePercent,string"`
	WeightedAvgPrice   decimal.Decimal `json:"weightedAvgPrice,string"`
	PrevClosePrice     decimal.Decimal `json:"prevClosePrice,string"`
	LastPrice          decimal.Decimal `json:"lastPrice,string"`
	LastQty            decimal.Decimal `json:"lastQty,string"`
	BidPrice           decimal.Decimal `json:"bidPrice,string"`
	AskPrice           decimal.Decimal `json:"askPrice,string"`
	OpenPrice          decimal.Decimal `json:"openPrice,string"`
	HighPrice          decimal.Decimal `json:"highPrice,string"`
	LowPrice           decimal.Decimal `json:"lowPrice,string"`
	Volume             decimal.Decimal `json:"volume,string"`
	QuoteVolume        decimal.Decimal `json:"quoteVolume,string"`
	OpenTime           int64           `json:"openTime"`
	CloseTime          int64           `json:"closeTime"`
	FirstID            int64           `json:"fristId"`
	LastID             int64           `json:"lastId"`
	Count              int64           `json:"count"`
}

// SymbolPrice holds basic symbol price
type SymbolPrice struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price,string"`
}

// BestPrice holds best price data
type BestPrice struct {
	Symbol   string          `json:"symbol"`
	BidPrice decimal.Decimal `json:"bidPrice,string"`
	BidQty   decimal.Decimal `json:"bidQty,string"`
	AskPrice decimal.Decimal `json:"askPrice,string"`
	AskQty   decimal.Decimal `json:"askQty,string"`
}

// NewOrderRequest request type
type NewOrderRequest struct {
	// Symbol (currency pair to trade)
	Symbol string
	// Side Buy or Sell
	Side RequestParamsSideType
	// TradeType (market or limit order)
	TradeType RequestParamsOrderType
	// TimeInForce specifies how long the order remains in effect.
	// Examples are (Good Till Cancel (GTC), Immediate or Cancel (IOC) and Fill Or Kill (FOK))
	TimeInForce RequestParamsTimeForceType
	// Quantity
	Quantity         decimal.Decimal
	Price            decimal.Decimal
	NewClientOrderID string
	StopPrice        decimal.Decimal //Used with STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, and TAKE_PROFIT_LIMIT orders.
	IcebergQty       decimal.Decimal //Used with LIMIT, STOP_LOSS_LIMIT, and TAKE_PROFIT_LIMIT to create an iceberg order.
	NewOrderRespType string
}

// NewOrderResponse is the return structured response from the exchange
type NewOrderResponse struct {
	Code            int             `json:"code"`
	Msg             string          `json:"msg"`
	Symbol          string          `json:"symbol"`
	OrderID         int64           `json:"orderId"`
	ClientOrderID   string          `json:"clientOrderId"`
	TransactionTime int64           `json:"transactTime"`
	Price           decimal.Decimal `json:"price,string"`
	OrigQty         decimal.Decimal `json:"origQty,string"`
	ExecutedQty     decimal.Decimal `json:"executedQty,string"`
	Status          string          `json:"status"`
	TimeInForce     string          `json:"timeInForce"`
	Type            string          `json:"type"`
	Side            string          `json:"side"`
	Fills           []struct {
		Price           decimal.Decimal `json:"price,string"`
		Qty             decimal.Decimal `json:"qty,string"`
		Commission      decimal.Decimal `json:"commission,string"`
		CommissionAsset decimal.Decimal `json:"commissionAsset,string"`
	} `json:"fills"`
}

// CancelOrderResponse is the return structured response from the exchange
type CancelOrderResponse struct {
	Symbol            string `json:"symbol"`
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           int64  `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
}

// QueryOrderData holds query order data
type QueryOrderData struct {
	Code          int             `json:"code"`
	Msg           string          `json:"msg"`
	Symbol        string          `json:"symbol"`
	OrderID       int64           `json:"orderId"`
	ClientOrderID string          `json:"clientOrderId"`
	Price         decimal.Decimal `json:"price,string"`
	OrigQty       decimal.Decimal `json:"origQty,string"`
	ExecutedQty   decimal.Decimal `json:"executedQty,string"`
	Status        string          `json:"status"`
	TimeInForce   string          `json:"timeInForce"`
	Type          string          `json:"type"`
	Side          string          `json:"side"`
	StopPrice     decimal.Decimal `json:"stopPrice,string"`
	IcebergQty    decimal.Decimal `json:"icebergQty,string"`
	Time          int64           `json:"time"`
	IsWorking     bool            `json:"isWorking"`
}

// Balance holds query order data
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// Account holds the account data
type Account struct {
	MakerCommission  int       `json:"makerCommission"`
	TakerCommission  int       `json:"takerCommission"`
	BuyerCommission  int       `json:"buyerCommission"`
	SellerCommission int       `json:"sellerCommission"`
	CanTrade         bool      `json:"canTrade"`
	CanWithdraw      bool      `json:"canWithdraw"`
	CanDeposit       bool      `json:"canDeposit"`
	UpdateTime       int64     `json:"updateTime"`
	Balances         []Balance `json:"balances"`
}

// RequestParamsSideType trade order side (buy or sell)
type RequestParamsSideType string

var (
	// BinanceRequestParamsSideBuy buy order type
	BinanceRequestParamsSideBuy = RequestParamsSideType("BUY")

	// BinanceRequestParamsSideSell sell order type
	BinanceRequestParamsSideSell = RequestParamsSideType("SELL")
)

// RequestParamsTimeForceType Time in force
type RequestParamsTimeForceType string

var (
	// BinanceRequestParamsTimeGTC GTC
	BinanceRequestParamsTimeGTC = RequestParamsTimeForceType("GTC")

	// BinanceRequestParamsTimeIOC IOC
	BinanceRequestParamsTimeIOC = RequestParamsTimeForceType("IOC")

	// BinanceRequestParamsTimeFOK FOK
	BinanceRequestParamsTimeFOK = RequestParamsTimeForceType("FOK")
)

// RequestParamsOrderType trade order type
type RequestParamsOrderType string

var (
	// BinanceRequestParamsOrderLimit Limit order
	BinanceRequestParamsOrderLimit = RequestParamsOrderType("LIMIT")

	// BinanceRequestParamsOrderMarket Market order
	BinanceRequestParamsOrderMarket = RequestParamsOrderType("MARKET")

	// BinanceRequestParamsOrderStopLoss STOP_LOSS
	BinanceRequestParamsOrderStopLoss = RequestParamsOrderType("STOP_LOSS")

	// BinanceRequestParamsOrderStopLossLimit STOP_LOSS_LIMIT
	BinanceRequestParamsOrderStopLossLimit = RequestParamsOrderType("STOP_LOSS_LIMIT")

	// BinanceRequestParamsOrderTakeProfit TAKE_PROFIT
	BinanceRequestParamsOrderTakeProfit = RequestParamsOrderType("TAKE_PROFIT")

	// BinanceRequestParamsOrderTakeProfitLimit TAKE_PROFIT_LIMIT
	BinanceRequestParamsOrderTakeProfitLimit = RequestParamsOrderType("TAKE_PROFIT_LIMIT")

	// BinanceRequestParamsOrderLimitMarker LIMIT_MAKER
	BinanceRequestParamsOrderLimitMarker = RequestParamsOrderType("LIMIT_MAKER")
)

// KlinesRequestParams represents Klines request data.
type KlinesRequestParams struct {
	Symbol    string       // Required field; example LTCBTC, BTCUSDT
	Interval  TimeInterval // Time interval period
	Limit     int          // Default 500; max 500.
	StartTime int64
	EndTime   int64
}

// TimeInterval represents interval enum.
type TimeInterval string

// Vars related to time intervals
var (
	TimeIntervalMinute         = TimeInterval("1m")
	TimeIntervalThreeMinutes   = TimeInterval("3m")
	TimeIntervalFiveMinutes    = TimeInterval("5m")
	TimeIntervalFifteenMinutes = TimeInterval("15m")
	TimeIntervalThirtyMinutes  = TimeInterval("30m")
	TimeIntervalHour           = TimeInterval("1h")
	TimeIntervalTwoHours       = TimeInterval("2h")
	TimeIntervalFourHours      = TimeInterval("4h")
	TimeIntervalSixHours       = TimeInterval("6h")
	TimeIntervalEightHours     = TimeInterval("8h")
	TimeIntervalTwelveHours    = TimeInterval("12h")
	TimeIntervalDay            = TimeInterval("1d")
	TimeIntervalThreeDays      = TimeInterval("3d")
	TimeIntervalWeek           = TimeInterval("1w")
	TimeIntervalMonth          = TimeInterval("1M")
)
