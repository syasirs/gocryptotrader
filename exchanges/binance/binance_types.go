package binance

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/types"
)

const wsRateLimitMilliseconds = 250

// withdrawals status codes description
const (
	EmailSent = iota
	Cancelled
	AwaitingApproval
	Rejected
	Processing
	Failure
	Completed

	// Futures channels
	contractInfoAllChan = "!contractInfo"
	forceOrderAllChan   = "!forceOrder@arr"
	bookTickerAllChan   = "!bookTicker"
	tickerAllChan       = "!ticker@arr"
	miniTickerAllChan   = "!miniTicker@arr"
	aggTradeChan        = "@aggTrade"
	depthChan           = "@depth"
	markPriceChan       = "@markPrice"
	tickerChan          = "@ticker"
	klineChan           = "@kline"
	miniTickerChan      = "@miniTicker"
	forceOrderChan      = "@forceOrder"
	continuousKline     = "continuousKline"

	// USDT Marigined futures
	markPriceAllChan   = "!markPrice@arr"
	assetIndexChan     = "@assetIndex"
	bookTickersChan    = "@bookTickers"
	assetIndexAllChan  = "!assetIndex@arr"
	compositeIndexChan = "@compositeIndex"

	// Coin Margined futures
	indexPriceCFuturesChan      = "@indexPrice"
	bookTickerCFuturesChan      = "@bookTicker"
	indexPriceKlineCFuturesChan = "@indexPriceKline"
	markPriceKlineCFuturesChan  = "@markPriceKline"
)

type filterType string

const (
	priceFilter              filterType = "PRICE_FILTER"
	lotSizeFilter            filterType = "LOT_SIZE"
	icebergPartsFilter       filterType = "ICEBERG_PARTS"
	marketLotSizeFilter      filterType = "MARKET_LOT_SIZE"
	trailingDeltaFilter      filterType = "TRAILING_DELTA"
	percentPriceFilter       filterType = "PERCENT_PRICE"
	percentPriceBySizeFilter filterType = "PERCENT_PRICE_BY_SIDE"
	notionalFilter           filterType = "NOTIONAL"
	maxNumOrdersFilter       filterType = "MAX_NUM_ORDERS"
	maxNumAlgoOrdersFilter   filterType = "MAX_NUM_ALGO_ORDERS"
)

// ExchangeInfo holds the full exchange information type
type ExchangeInfo struct {
	Code            int                  `json:"code"`
	Msg             string               `json:"msg"`
	Timezone        string               `json:"timezone"`
	ServerTime      convert.ExchangeTime `json:"serverTime"`
	RateLimits      []*RateLimitItem     `json:"rateLimits"`
	ExchangeFilters interface{}          `json:"exchangeFilters"`
	Symbols         []*struct {
		Symbol                     string        `json:"symbol"`
		Status                     string        `json:"status"`
		BaseAsset                  string        `json:"baseAsset"`
		BaseAssetPrecision         int64         `json:"baseAssetPrecision"`
		QuoteAsset                 string        `json:"quoteAsset"`
		QuotePrecision             int64         `json:"quotePrecision"`
		OrderTypes                 []string      `json:"orderTypes"`
		IcebergAllowed             bool          `json:"icebergAllowed"`
		OCOAllowed                 bool          `json:"ocoAllowed"`
		QuoteOrderQtyMarketAllowed bool          `json:"quoteOrderQtyMarketAllowed"`
		IsSpotTradingAllowed       bool          `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool          `json:"isMarginTradingAllowed"`
		Filters                    []*filterData `json:"filters"`
		Permissions                []string      `json:"permissions"`
	} `json:"symbols"`
}

type filterData struct {
	FilterType          filterType `json:"filterType"`
	MinPrice            float64    `json:"minPrice,string"`
	MaxPrice            float64    `json:"maxPrice,string"`
	TickSize            float64    `json:"tickSize,string"`
	MultiplierUp        float64    `json:"multiplierUp,string"`
	MultiplierDown      float64    `json:"multiplierDown,string"`
	AvgPriceMinutes     int64      `json:"avgPriceMins"`
	MinQty              float64    `json:"minQty,string"`
	MaxQty              float64    `json:"maxQty,string"`
	StepSize            float64    `json:"stepSize,string"`
	MinNotional         float64    `json:"minNotional,string"`
	ApplyToMarket       bool       `json:"applyToMarket"`
	Limit               int64      `json:"limit"`
	MaxNumAlgoOrders    int64      `json:"maxNumAlgoOrders"`
	MaxNumIcebergOrders int64      `json:"maxNumIcebergOrders"`
	MaxNumOrders        int64      `json:"maxNumOrders"`
}

// CoinInfo stores information about all supported coins
type CoinInfo struct {
	Coin              string  `json:"coin"`
	DepositAllEnable  bool    `json:"depositAllEnable"`
	WithdrawAllEnable bool    `json:"withdrawAllEnable"`
	Free              float64 `json:"free,string"`
	Freeze            float64 `json:"freeze,string"`
	IPOAble           float64 `json:"ipoable,string"`
	IPOing            float64 `json:"ipoing,string"`
	IsLegalMoney      bool    `json:"isLegalMoney"`
	Locked            float64 `json:"locked,string"`
	Name              string  `json:"name"`
	NetworkList       []struct {
		AddressRegex        string  `json:"addressRegex"`
		Coin                string  `json:"coin"`
		DepositDescription  string  `json:"depositDesc"` // shown only when "depositEnable" is false
		DepositEnable       bool    `json:"depositEnable"`
		IsDefault           bool    `json:"isDefault"`
		MemoRegex           string  `json:"memoRegex"`
		MinimumConfirmation uint16  `json:"minConfirm"`
		Name                string  `json:"name"`
		Network             string  `json:"network"`
		ResetAddressStatus  bool    `json:"resetAddressStatus"`
		SpecialTips         string  `json:"specialTips"`
		UnlockConfirm       uint16  `json:"unLockConfirm"`
		WithdrawDescription string  `json:"withdrawDesc"` // shown only when "withdrawEnable" is false
		WithdrawEnable      bool    `json:"withdrawEnable"`
		WithdrawFee         float64 `json:"withdrawFee,string"`
		WithdrawMinimum     float64 `json:"withdrawMin,string"`
		WithdrawMaximum     float64 `json:"withdrawMax,string"`
	} `json:"networkList"`
	Storage     float64 `json:"storage,string"`
	Trading     bool    `json:"trading"`
	Withdrawing float64 `json:"withdrawing,string"`
}

// OrderBookDataRequestParams represents Klines request data.
type OrderBookDataRequestParams struct {
	Symbol currency.Pair `json:"symbol"` // Required field; example LTCBTC,BTCUSDT
	Limit  int64         `json:"limit"`  // Default 100; max 1000. Valid limits:[5, 10, 20, 50, 100, 500, 1000]
}

// OrderbookItem stores an individual orderbook item
type OrderbookItem struct {
	Price    float64
	Quantity float64
}

// OrderBookData is resp data from orderbook endpoint
type OrderBookData struct {
	Code         int64             `json:"code"`
	Msg          string            `json:"msg"`
	LastUpdateID int64             `json:"lastUpdateId"`
	Bids         [][2]types.Number `json:"bids"`
	Asks         [][2]types.Number `json:"asks"`
}

// OrderBook actual structured data that can be used for orderbook
type OrderBook struct {
	Symbol       string
	LastUpdateID int64
	Code         int64
	Msg          string
	Bids         []OrderbookItem
	Asks         []OrderbookItem
}

// DepthUpdateParams is used as an embedded type for WebsocketDepthStream
type DepthUpdateParams []struct {
	PriceLevel float64
	Quantity   float64
	ignore     []interface{}
}

// WebsocketDepthStream is the difference for the update depth stream
type WebsocketDepthStream struct {
	Event         string               `json:"e"`
	Timestamp     convert.ExchangeTime `json:"E"`
	Pair          string               `json:"s"`
	FirstUpdateID int64                `json:"U"`
	LastUpdateID  int64                `json:"u"`
	UpdateBids    [][2]types.Number    `json:"b"`
	UpdateAsks    [][2]types.Number    `json:"a"`
}

// RecentTradeRequestParams represents Klines request data.
type RecentTradeRequestParams struct {
	Symbol currency.Pair `json:"symbol"` // Required field. example LTCBTC, BTCUSDT
	Limit  int64         `json:"limit"`  // Default 500; max 500.
	FromID int64         `json:"fromId,omitempty"`
}

// RecentTrade holds recent trade data
type RecentTrade struct {
	ID           int64                `json:"id"`
	Price        float64              `json:"price,string"`
	Quantity     float64              `json:"qty,string"`
	Time         convert.ExchangeTime `json:"time"`
	IsBuyerMaker bool                 `json:"isBuyerMaker"`
	IsBestMatch  bool                 `json:"isBestMatch"`
}

// TradeStream holds the trade stream data
type TradeStream struct {
	EventType      string               `json:"e"`
	EventTime      convert.ExchangeTime `json:"E"`
	Symbol         string               `json:"s"`
	TradeID        int64                `json:"t"`
	Price          types.Number         `json:"p"`
	Quantity       types.Number         `json:"q"`
	BuyerOrderID   int64                `json:"b"`
	SellerOrderID  int64                `json:"a"`
	TimeStamp      convert.ExchangeTime `json:"T"`
	Maker          bool                 `json:"m"`
	BestMatchPrice bool                 `json:"M"`
}

// KlineStream holds the kline stream data
type KlineStream struct {
	EventType string               `json:"e"`
	EventTime convert.ExchangeTime `json:"E"`
	Symbol    string               `json:"s"`
	Kline     KlineStreamData      `json:"k"`
}

// KlineStreamData defines kline streaming data
type KlineStreamData struct {
	StartTime                convert.ExchangeTime `json:"t"`
	CloseTime                convert.ExchangeTime `json:"T"`
	Symbol                   string               `json:"s"`
	Interval                 string               `json:"i"`
	FirstTradeID             int64                `json:"f"`
	LastTradeID              int64                `json:"L"`
	OpenPrice                types.Number         `json:"o"`
	ClosePrice               types.Number         `json:"c"`
	HighPrice                types.Number         `json:"h"`
	LowPrice                 types.Number         `json:"l"`
	Volume                   types.Number         `json:"v"`
	NumberOfTrades           int64                `json:"n"`
	KlineClosed              bool                 `json:"x"`
	Quote                    types.Number         `json:"q"`
	TakerBuyBaseAssetVolume  types.Number         `json:"V"`
	TakerBuyQuoteAssetVolume types.Number         `json:"Q"`
}

// TickerStream holds the ticker stream data
type TickerStream struct {
	EventType              string               `json:"e"`
	EventTime              convert.ExchangeTime `json:"E"`
	Symbol                 string               `json:"s"`
	PriceChange            types.Number         `json:"p"`
	PriceChangePercent     types.Number         `json:"P"`
	WeightedAvgPrice       types.Number         `json:"w"`
	ClosePrice             types.Number         `json:"x"`
	LastPrice              types.Number         `json:"c"`
	LastPriceQuantity      types.Number         `json:"Q"`
	BestBidPrice           types.Number         `json:"b"`
	BestBidQuantity        types.Number         `json:"B"`
	BestAskPrice           types.Number         `json:"a"`
	BestAskQuantity        types.Number         `json:"A"`
	OpenPrice              types.Number         `json:"o"`
	HighPrice              types.Number         `json:"h"`
	LowPrice               types.Number         `json:"l"`
	TotalTradedVolume      types.Number         `json:"v"`
	TotalTradedQuoteVolume types.Number         `json:"q"`
	OpenTime               time.Time            `json:"O"`
	CloseTime              time.Time            `json:"C"`
	FirstTradeID           int64                `json:"F"`
	LastTradeID            int64                `json:"L"`
	NumberOfTrades         int64                `json:"n"`
}

// HistoricalTrade holds recent trade data
type HistoricalTrade struct {
	ID            int64                `json:"id"`
	Price         float64              `json:"price,string"`
	Quantity      float64              `json:"qty,string"`
	QuoteQuantity float64              `json:"quoteQty,string"`
	Time          convert.ExchangeTime `json:"time"`
	IsBuyerMaker  bool                 `json:"isBuyerMaker"`
	IsBestMatch   bool                 `json:"isBestMatch"`
}

// AggregatedTradeRequestParams holds request params
type AggregatedTradeRequestParams struct {
	Symbol string // Required field; example LTCBTC, BTCUSDT
	// The first trade to retrieve
	FromID int64
	// The API seems to accept (start and end time) or FromID and no other combinations
	StartTime time.Time
	EndTime   time.Time
	// Default 500; max 1000.
	Limit int
}

// WsAggregateTradeRequestParams holds request parameters for aggregate trades
type WsAggregateTradeRequestParams struct {
	Symbol    string `json:"symbol"`
	FromID    int64  `json:"fromId,omitempty"`
	Limit     int64  `json:"limit,omitempty"`
	StartTime int64  `json:"startTime,omitempty"`
	EndTime   int64  `json:"endTime,omitempty"`
}

// AggregatedTrade holds aggregated trade information
type AggregatedTrade struct {
	ATradeID       int64                `json:"a"`
	Price          float64              `json:"p,string"`
	Quantity       float64              `json:"q,string"`
	FirstTradeID   int64                `json:"f"`
	LastTradeID    int64                `json:"l"`
	TimeStamp      convert.ExchangeTime `json:"T"`
	Maker          bool                 `json:"m"`
	BestMatchPrice bool                 `json:"M"`
}

// UFuturesAggregatedTrade represents usdt futures aggregated trade information
type UFuturesAggregatedTrade struct {
	EventType        string               `json:"e"`
	EventTime        convert.ExchangeTime `json:"E"`
	Symbol           string               `json:"s"`
	AggregateTradeID int64                `json:"a"`
	Price            types.Number         `json:"p"`
	Quantity         types.Number         `json:"q"`
	FirstTradeID     int64                `json:"f"`
	LastTradeID      int64                `json:"l"`
	TradeTime        convert.ExchangeTime `json:"T"`
	MarketMaker      bool                 `json:"m"`
}

// IndexMarkPrice stores data for index and mark prices
type IndexMarkPrice struct {
	Symbol               string               `json:"symbol"`
	Pair                 string               `json:"pair"`
	MarkPrice            types.Number         `json:"markPrice"`
	IndexPrice           types.Number         `json:"indexPrice"`
	EstimatedSettlePrice types.Number         `json:"estimatedSettlePrice"`
	LastFundingRate      types.Number         `json:"lastFundingRate"`
	NextFundingTime      convert.ExchangeTime `json:"nextFundingTime"`
	Time                 convert.ExchangeTime `json:"time"`
}

// CandleStick holds kline data
type CandleStick struct {
	OpenTime                 time.Time
	Open                     float64
	High                     float64
	Low                      float64
	Close                    float64
	Volume                   float64
	CloseTime                time.Time
	QuoteAssetVolume         float64
	TradeCount               float64
	TakerBuyAssetVolume      float64
	TakerBuyQuoteAssetVolume float64
}

// AveragePrice holds current average symbol price
type AveragePrice struct {
	Mins  int64   `json:"mins"`
	Price float64 `json:"price,string"`
}

// PriceChangeStats contains statistics for the last 24 hours trade
type PriceChangeStats struct {
	Symbol             string               `json:"symbol"`
	PriceChange        types.Number         `json:"priceChange"`
	PriceChangePercent types.Number         `json:"priceChangePercent"`
	WeightedAvgPrice   types.Number         `json:"weightedAvgPrice"`
	PrevClosePrice     types.Number         `json:"prevClosePrice"`
	LastPrice          types.Number         `json:"lastPrice"`
	OpenPrice          types.Number         `json:"openPrice"`
	HighPrice          types.Number         `json:"highPrice"`
	LowPrice           types.Number         `json:"lowPrice"`
	Volume             types.Number         `json:"volume"`
	QuoteVolume        types.Number         `json:"quoteVolume"`
	OpenTime           convert.ExchangeTime `json:"openTime"`
	CloseTime          convert.ExchangeTime `json:"closeTime"`
	FirstID            int64                `json:"firstId"`
	LastID             int64                `json:"lastId"`
	Count              int64                `json:"count"`

	LastQty  types.Number `json:"lastQty"`
	BidPrice types.Number `json:"bidPrice"`
	BidQty   types.Number `json:"bidQty"`
	AskPrice types.Number `json:"askPrice"`
	AskQty   types.Number `json:"askQty"`
}

// SymbolPrice holds basic symbol price
type SymbolPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

// BestPrice holds best price data
type BestPrice struct {
	Symbol   string  `json:"symbol"`
	BidPrice float64 `json:"bidPrice,string"`
	BidQty   float64 `json:"bidQty,string"`
	AskPrice float64 `json:"askPrice,string"`
	AskQty   float64 `json:"askQty,string"`
}

// NewOrderRequest request type
type NewOrderRequest struct {
	// Symbol (currency pair to trade)
	Symbol currency.Pair
	// Side Buy or Sell
	Side string
	// TradeType (market or limit order)
	TradeType RequestParamsOrderType
	// TimeInForce specifies how long the order remains in effect.
	// Examples are (Good Till Cancel (GTC), Immediate or Cancel (IOC) and Fill Or Kill (FOK))
	TimeInForce RequestParamsTimeForceType
	// Quantity is the total base qty spent or received in an order.
	Quantity float64
	// QuoteOrderQty is the total quote qty spent or received in a MARKET order.
	QuoteOrderQty    float64
	Price            float64
	NewClientOrderID string
	StopPrice        float64 // Used with STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, and TAKE_PROFIT_LIMIT orders.
	IcebergQty       float64 // Used with LIMIT, STOP_LOSS_LIMIT, and TAKE_PROFIT_LIMIT to create an iceberg order.
	NewOrderRespType string
}

// NewOrderResponse is the return structured response from the exchange
type NewOrderResponse struct {
	Code            int64                `json:"code"`
	Msg             string               `json:"msg"`
	Symbol          string               `json:"symbol"`
	OrderID         int64                `json:"orderId"`
	ClientOrderID   string               `json:"clientOrderId"`
	TransactionTime convert.ExchangeTime `json:"transactTime"`
	Price           float64              `json:"price,string"`
	OrigQty         float64              `json:"origQty,string"`
	ExecutedQty     float64              `json:"executedQty,string"`
	// The cumulative amount of the quote that has been spent (with a BUY order) or received (with a SELL order).
	CumulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"`
	Status             string  `json:"status"`
	TimeInForce        string  `json:"timeInForce"`
	Type               string  `json:"type"`
	Side               string  `json:"side"`
	Fills              []struct {
		Price           float64 `json:"price,string"`
		Quantity        float64 `json:"qty,string"`
		Commission      float64 `json:"commission,string"`
		CommissionAsset string  `json:"commissionAsset"`
	} `json:"fills"`
}

// CancelOrderResponse is the return structured response from the exchange
type CancelOrderResponse struct {
	Symbol            string `json:"symbol"`
	OrigClientOrderID string `json:"origClientOrderId"`
	OrderID           int64  `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
}

// TradeOrder holds query order data
// Note that some fields are optional and included only for orders that set them.
type TradeOrder struct {
	Code                    int64                `json:"code"`
	Msg                     string               `json:"msg"`
	Symbol                  string               `json:"symbol"`
	OrderID                 int64                `json:"orderId"`
	OrderListID             int64                `json:"orderListId"`
	ClientOrderID           string               `json:"clientOrderId"`
	Price                   types.Number         `json:"price"`
	OrigQty                 types.Number         `json:"origQty"`
	ExecutedQty             types.Number         `json:"executedQty"`
	CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
	Status                  string               `json:"status"`
	TimeInForce             string               `json:"timeInForce"`
	Type                    string               `json:"type"`
	Side                    string               `json:"side"`
	IsWorking               bool                 `json:"isWorking"`
	StopPrice               types.Number         `json:"stopPrice"`
	IcebergQty              types.Number         `json:"icebergQty"`
	Time                    convert.ExchangeTime `json:"time"`
	UpdateTime              convert.ExchangeTime `json:"updateTime"`
	WorkingTime             convert.ExchangeTime `json:"workingTime"`
	OrigQuoteOrderQty       types.Number         `json:"origQuoteOrderQty"`
	SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	OrigClientOrderID       string               `json:"origClientOrderId"`
	TransactTime            convert.ExchangeTime `json:"transactTime"`

	PreventedMatchID  int64        `json:"preventedMatchId"`
	PreventedQuantity types.Number `json:"preventedQuantity"`
}

// Balance holds query order data
type Balance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

// Account holds the account data
type Account struct {
	UID              int64        `json:"uid"`
	MakerCommission  types.Number `json:"makerCommission"`
	TakerCommission  types.Number `json:"takerCommission"`
	BuyerCommission  types.Number `json:"buyerCommission"`
	SellerCommission types.Number `json:"sellerCommission"`
	CanTrade         bool         `json:"canTrade"`
	CanWithdraw      bool         `json:"canWithdraw"`
	CanDeposit       bool         `json:"canDeposit"`
	CommissionRates  struct {
		Maker  types.Number `json:"maker"`
		Taker  types.Number `json:"taker"`
		Buyer  types.Number `json:"buyer"`
		Seller types.Number `json:"seller"`
	} `json:"commissionRates"`
	Brokered                   bool                 `json:"brokered"`
	RequireSelfTradePrevention bool                 `json:"requireSelfTradePrevention"`
	PreventSor                 bool                 `json:"preventSor"`
	UpdateTime                 convert.ExchangeTime `json:"updateTime"`
	AccountType                string               `json:"accountType"`
	Balances                   []Balance            `json:"balances"`
	Permissions                []string             `json:"permissions"`
}

// MarginAccount holds the margin account data
type MarginAccount struct {
	BorrowEnabled       bool                 `json:"borrowEnabled"`
	MarginLevel         float64              `json:"marginLevel,string"`
	TotalAssetOfBtc     float64              `json:"totalAssetOfBtc,string"`
	TotalLiabilityOfBtc float64              `json:"totalLiabilityOfBtc,string"`
	TotalNetAssetOfBtc  float64              `json:"totalNetAssetOfBtc,string"`
	TradeEnabled        bool                 `json:"tradeEnabled"`
	TransferEnabled     bool                 `json:"transferEnabled"`
	UserAssets          []MarginAccountAsset `json:"userAssets"`
}

// MarginAccountAsset holds each individual margin account asset
type MarginAccountAsset struct {
	Asset    string  `json:"asset"`
	Borrowed float64 `json:"borrowed,string"`
	Free     float64 `json:"free,string"`
	Interest float64 `json:"interest,string"`
	Locked   float64 `json:"locked,string"`
	NetAsset float64 `json:"netAsset,string"`
}

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
	Symbol         currency.Pair `json:"symbol"`             // Required field; example LTCBTC, BTCUSDT
	Interval       string        `json:"interval,omitempty"` // Time interval period
	Limit          int64         `json:"limit,omitempty"`    // Default 500; max 500.
	StartTime      time.Time     `json:"-"`
	EndTime        time.Time     `json:"-"`
	Timezone       string        `json:"timeZone,omitempty"`
	StartTimestamp int64         `json:"startTime,omitempty"`
	EndTimestamp   int64         `json:"endTime,omitempty"`
}

// WithdrawalFees the large list of predefined withdrawal fees
// Prone to change
var WithdrawalFees = map[currency.Code]float64{
	currency.BNB:     0.13,
	currency.BTC:     0.0005,
	currency.NEO:     0,
	currency.ETH:     0.01,
	currency.LTC:     0.001,
	currency.QTUM:    0.01,
	currency.EOS:     0.1,
	currency.SNT:     35,
	currency.BNT:     1,
	currency.GAS:     0,
	currency.BCC:     0.001,
	currency.BTM:     5,
	currency.USDT:    3.4,
	currency.HCC:     0.0005,
	currency.OAX:     6.5,
	currency.DNT:     54,
	currency.MCO:     0.31,
	currency.ICN:     3.5,
	currency.ZRX:     1.9,
	currency.OMG:     0.4,
	currency.WTC:     0.5,
	currency.LRC:     12.3,
	currency.LLT:     67.8,
	currency.YOYO:    1,
	currency.TRX:     1,
	currency.STRAT:   0.1,
	currency.SNGLS:   54,
	currency.BQX:     3.9,
	currency.KNC:     3.5,
	currency.SNM:     25,
	currency.FUN:     86,
	currency.LINK:    4,
	currency.XVG:     0.1,
	currency.CTR:     35,
	currency.SALT:    2.3,
	currency.MDA:     2.3,
	currency.IOTA:    0.5,
	currency.SUB:     11.4,
	currency.ETC:     0.01,
	currency.MTL:     2,
	currency.MTH:     45,
	currency.ENG:     2.2,
	currency.AST:     14.4,
	currency.DASH:    0.002,
	currency.BTG:     0.001,
	currency.EVX:     2.8,
	currency.REQ:     29.9,
	currency.VIB:     30,
	currency.POWR:    8.2,
	currency.ARK:     0.2,
	currency.XRP:     0.25,
	currency.MOD:     2,
	currency.ENJ:     26,
	currency.STORJ:   5.1,
	currency.KMD:     0.002,
	currency.RCN:     47,
	currency.NULS:    0.01,
	currency.RDN:     2.5,
	currency.XMR:     0.04,
	currency.DLT:     19.8,
	currency.AMB:     8.9,
	currency.BAT:     8,
	currency.ZEC:     0.005,
	currency.BCPT:    14.5,
	currency.ARN:     3,
	currency.GVT:     0.13,
	currency.CDT:     81,
	currency.GXS:     0.3,
	currency.POE:     134,
	currency.QSP:     36,
	currency.BTS:     1,
	currency.XZC:     0.02,
	currency.LSK:     0.1,
	currency.TNT:     47,
	currency.FUEL:    79,
	currency.MANA:    18,
	currency.BCD:     0.01,
	currency.DGD:     0.04,
	currency.ADX:     6.3,
	currency.ADA:     1,
	currency.PPT:     0.41,
	currency.CMT:     12,
	currency.XLM:     0.01,
	currency.CND:     58,
	currency.LEND:    84,
	currency.WABI:    6.6,
	currency.SBTC:    0.0005,
	currency.BCX:     0.5,
	currency.WAVES:   0.002,
	currency.TNB:     139,
	currency.GTO:     20,
	currency.ICX:     0.02,
	currency.OST:     32,
	currency.ELF:     3.9,
	currency.AION:    3.2,
	currency.CVC:     10.9,
	currency.REP:     0.2,
	currency.GNT:     8.9,
	currency.DATA:    37,
	currency.ETF:     1,
	currency.BRD:     3.8,
	currency.NEBL:    0.01,
	currency.VIBE:    17.3,
	currency.LUN:     0.36,
	currency.CHAT:    60.7,
	currency.RLC:     3.4,
	currency.INS:     3.5,
	currency.IOST:    105.6,
	currency.STEEM:   0.01,
	currency.NANO:    0.01,
	currency.AE:      1.3,
	currency.VIA:     0.01,
	currency.BLZ:     10.3,
	currency.SYS:     1,
	currency.NCASH:   247.6,
	currency.POA:     0.01,
	currency.ONT:     1,
	currency.ZIL:     37.2,
	currency.STORM:   152,
	currency.XEM:     4,
	currency.WAN:     0.1,
	currency.WPR:     43.4,
	currency.QLC:     1,
	currency.GRS:     0.2,
	currency.CLOAK:   0.02,
	currency.LOOM:    11.9,
	currency.BCN:     1,
	currency.TUSD:    1.35,
	currency.ZEN:     0.002,
	currency.SKY:     0.01,
	currency.THETA:   24,
	currency.IOTX:    90.5,
	currency.QKC:     24.6,
	currency.AGI:     29.81,
	currency.NXS:     0.02,
	currency.SC:      0.1,
	currency.EON:     10,
	currency.NPXS:    897,
	currency.KEY:     223,
	currency.NAS:     0.1,
	currency.ADD:     100,
	currency.MEETONE: 300,
	currency.ATD:     100,
	currency.MFT:     175,
	currency.EOP:     5,
	currency.DENT:    596,
	currency.IQ:      50,
	currency.ARDR:    2,
	currency.HOT:     1210,
	currency.VET:     100,
	currency.DOCK:    68,
	currency.POLY:    7,
	currency.VTHO:    21,
	currency.ONG:     0.1,
	currency.PHX:     1,
	currency.HC:      0.005,
	currency.GO:      0.01,
	currency.PAX:     1.4,
	currency.EDO:     1.3,
	currency.WINGS:   8.9,
	currency.NAV:     0.2,
	currency.TRIG:    49.1,
	currency.APPC:    12.4,
	currency.PIVX:    0.02,
}

// DepositHistory stores deposit history info
type DepositHistory struct {
	Amount        float64 `json:"amount,string"`
	Coin          string  `json:"coin"`
	Network       string  `json:"network"`
	Status        uint8   `json:"status"`
	Address       string  `json:"address"`
	AddressTag    string  `json:"adressTag"`
	TransactionID string  `json:"txId"`
	InsertTime    float64 `json:"insertTime"`
	TransferType  uint8   `json:"transferType"`
	ConfirmTimes  string  `json:"confirmTimes"`
}

// WithdrawResponse contains status of withdrawal request
type WithdrawResponse struct {
	ID string `json:"id"`
}

// WithdrawStatusResponse defines a withdrawal status response
type WithdrawStatusResponse struct {
	Address         string               `json:"address"`
	Amount          float64              `json:"amount,string"`
	ApplyTime       convert.ExchangeTime `json:"applyTime"`
	Coin            string               `json:"coin"`
	ID              string               `json:"id"`
	WithdrawOrderID string               `json:"withdrawOrderId"`
	Network         string               `json:"network"`
	TransferType    uint8                `json:"transferType"`
	Status          int64                `json:"status"`
	TransactionFee  float64              `json:"transactionFee,string"`
	TransactionID   string               `json:"txId"`
	ConfirmNumber   int64                `json:"confirmNo"`
}

// DepositAddress stores the deposit address info
type DepositAddress struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag"`
	URL     string `json:"url"`
}

// UserAccountStream contains a key to maintain an authorised
// websocket connection
type UserAccountStream struct {
	ListenKey string `json:"listenKey"`
}

type wsAccountInfo struct {
	Stream string            `json:"stream"`
	Data   WsAccountInfoData `json:"data"`
}

// WsAccountInfoData defines websocket account info data
type WsAccountInfoData struct {
	CanDeposit       bool                 `json:"D"`
	CanTrade         bool                 `json:"T"`
	CanWithdraw      bool                 `json:"W"`
	EventTime        convert.ExchangeTime `json:"E"`
	LastUpdated      convert.ExchangeTime `json:"u"`
	BuyerCommission  float64              `json:"b"`
	MakerCommission  float64              `json:"m"`
	SellerCommission float64              `json:"s"`
	TakerCommission  float64              `json:"t"`
	EventType        string               `json:"e"`
	Currencies       []struct {
		Asset     string  `json:"a"`
		Available float64 `json:"f,string"`
		Locked    float64 `json:"l,string"`
	} `json:"B"`
}

type wsAccountPosition struct {
	Stream string                `json:"stream"`
	Data   WsAccountPositionData `json:"data"`
}

// WsAccountPositionData defines websocket account position data
type WsAccountPositionData struct {
	Currencies []struct {
		Asset     string  `json:"a"`
		Available float64 `json:"f,string"`
		Locked    float64 `json:"l,string"`
	} `json:"B"`
	EventTime   convert.ExchangeTime `json:"E"`
	LastUpdated convert.ExchangeTime `json:"u"`
	EventType   string               `json:"e"`
}

type wsBalanceUpdate struct {
	Stream string              `json:"stream"`
	Data   WsBalanceUpdateData `json:"data"`
}

// WsBalanceUpdateData defines websocket account balance data
type WsBalanceUpdateData struct {
	EventTime    convert.ExchangeTime `json:"E"`
	ClearTime    convert.ExchangeTime `json:"T"`
	BalanceDelta float64              `json:"d,string"`
	Asset        string               `json:"a"`
	EventType    string               `json:"e"`
}

type wsOrderUpdate struct {
	Stream string            `json:"stream"`
	Data   WsOrderUpdateData `json:"data"`
}

// WsOrderUpdateData defines websocket account order update data
type WsOrderUpdateData struct {
	EventType                         string               `json:"e"`
	EventTime                         convert.ExchangeTime `json:"E"`
	Symbol                            string               `json:"s"`
	ClientOrderID                     string               `json:"c"`
	Side                              string               `json:"S"`
	OrderType                         string               `json:"o"`
	TimeInForce                       string               `json:"f"`
	Quantity                          float64              `json:"q,string"`
	Price                             float64              `json:"p,string"`
	StopPrice                         float64              `json:"P,string"`
	IcebergQuantity                   float64              `json:"F,string"`
	OrderListID                       int64                `json:"g"`
	CancelledClientOrderID            string               `json:"C"`
	CurrentExecutionType              string               `json:"x"`
	OrderStatus                       string               `json:"X"`
	RejectionReason                   string               `json:"r"`
	OrderID                           int64                `json:"i"`
	LastExecutedQuantity              float64              `json:"l,string"`
	CumulativeFilledQuantity          float64              `json:"z,string"`
	LastExecutedPrice                 float64              `json:"L,string"`
	Commission                        float64              `json:"n,string"`
	CommissionAsset                   string               `json:"N"`
	TransactionTime                   convert.ExchangeTime `json:"T"`
	TradeID                           int64                `json:"t"`
	Ignored                           int64                `json:"I"` // Must be ignored explicitly, otherwise it overwrites 'i'.
	IsOnOrderBook                     bool                 `json:"w"`
	IsMaker                           bool                 `json:"m"`
	Ignored2                          bool                 `json:"M"` // See the comment for "I".
	OrderCreationTime                 convert.ExchangeTime `json:"O"`
	WorkingTime                       convert.ExchangeTime `json:"W"`
	CumulativeQuoteTransactedQuantity float64              `json:"Z,string"`
	LastQuoteAssetTransactedQuantity  float64              `json:"Y,string"`
	QuoteOrderQuantity                float64              `json:"Q,string"`
}

type wsListStatus struct {
	Stream string           `json:"stream"`
	Data   WsListStatusData `json:"data"`
}

// WsListStatusData defines websocket account listing status data
type WsListStatusData struct {
	ListClientOrderID string               `json:"C"`
	EventTime         convert.ExchangeTime `json:"E"`
	ListOrderStatus   string               `json:"L"`
	Orders            []struct {
		ClientOrderID string `json:"c"`
		OrderID       int64  `json:"i"`
		Symbol        string `json:"s"`
	} `json:"O"`
	TransactionTime convert.ExchangeTime `json:"T"`
	ContingencyType string               `json:"c"`
	EventType       string               `json:"e"`
	OrderListID     int64                `json:"g"`
	ListStatusType  string               `json:"l"`
	RejectionReason string               `json:"r"`
	Symbol          string               `json:"s"`
}

// WsPayload defines the payload through the websocket connection
type WsPayload struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}

// CrossMarginInterestData stores cross margin data for borrowing
type CrossMarginInterestData struct {
	Code          int64  `json:"code,string"`
	Message       string `json:"message"`
	MessageDetail string `json:"messageDetail"`
	Data          []struct {
		AssetName string `json:"assetName"`
		Specs     []struct {
			VipLevel          string `json:"vipLevel"`
			DailyInterestRate string `json:"dailyInterestRate"`
			BorrowLimit       string `json:"borrowLimit"`
		} `json:"specs"`
	} `json:"data"`
	Success bool `json:"success"`
}

// orderbookManager defines a way of managing and maintaining synchronisation
// across connections and assets.
type orderbookManager struct {
	state map[currency.Code]map[currency.Code]map[asset.Item]*update
	sync.Mutex

	jobs chan job
}

type update struct {
	buffer            chan *WebsocketDepthStream
	fetchingBook      bool
	initialSync       bool
	needsFetchingBook bool
	lastUpdateID      int64
}

// job defines a synchronisation job that tells a go routine to fetch an
// orderbook via the REST protocol
type job struct {
	Pair currency.Pair
}

// UserMarginInterestHistoryResponse user margin interest history response
type UserMarginInterestHistoryResponse struct {
	Rows  []UserMarginInterestHistory `json:"rows"`
	Total int64                       `json:"total"`
}

// UserMarginInterestHistory user margin interest history row
type UserMarginInterestHistory struct {
	TxID                int64                `json:"txId"`
	InterestAccruedTime convert.ExchangeTime `json:"interestAccuredTime"` // typo in docs, cannot verify due to API restrictions
	Asset               string               `json:"asset"`
	RawAsset            string               `json:"rawAsset"`
	Principal           float64              `json:"principal,string"`
	Interest            float64              `json:"interest,string"`
	InterestRate        float64              `json:"interestRate,string"`
	Type                string               `json:"type"`
	IsolatedSymbol      string               `json:"isolatedSymbol"`
}

// CryptoLoansIncomeHistory stores crypto loan income history data
type CryptoLoansIncomeHistory struct {
	Asset         currency.Code `json:"asset"`
	Type          string        `json:"type"`
	Amount        float64       `json:"amount,string"`
	TransactionID int64         `json:"tranId"`
}

// CryptoLoanBorrow stores crypto loan borrow data
type CryptoLoanBorrow struct {
	LoanCoin           currency.Code `json:"loanCoin"`
	Amount             float64       `json:"amount,string"`
	CollateralCoin     currency.Code `json:"collateralCoin"`
	CollateralAmount   float64       `json:"collateralAmount,string"`
	HourlyInterestRate float64       `json:"hourlyInterestRate,string"`
	OrderID            int64         `json:"orderId,string"`
}

// LoanBorrowHistoryItem stores loan borrow history item data
type LoanBorrowHistoryItem struct {
	OrderID                 int64                `json:"orderId"`
	LoanCoin                currency.Code        `json:"loanCoin"`
	InitialLoanAmount       float64              `json:"initialLoanAmount,string"`
	HourlyInterestRate      float64              `json:"hourlyInterestRate,string"`
	LoanTerm                int64                `json:"loanTerm,string"`
	CollateralCoin          currency.Code        `json:"collateralCoin"`
	InitialCollateralAmount float64              `json:"initialCollateralAmount,string"`
	BorrowTime              convert.ExchangeTime `json:"borrowTime"`
	Status                  string               `json:"status"`
}

// LoanBorrowHistory stores loan borrow history data
type LoanBorrowHistory struct {
	Rows  []LoanBorrowHistoryItem `json:"rows"`
	Total int64                   `json:"total"`
}

// CryptoLoanOngoingOrderItem stores crypto loan ongoing order item data
type CryptoLoanOngoingOrderItem struct {
	OrderID          int64                `json:"orderId"`
	LoanCoin         currency.Code        `json:"loanCoin"`
	TotalDebt        float64              `json:"totalDebt,string"`
	ResidualInterest float64              `json:"residualInterest,string"`
	CollateralCoin   currency.Code        `json:"collateralCoin"`
	CollateralAmount float64              `json:"collateralAmount,string"`
	CurrentLTV       float64              `json:"currentLTV,string"`
	ExpirationTime   convert.ExchangeTime `json:"expirationTime"`
}

// CryptoLoanOngoingOrder stores crypto loan ongoing order data
type CryptoLoanOngoingOrder struct {
	Rows  []CryptoLoanOngoingOrderItem `json:"rows"`
	Total int64                        `json:"total"`
}

// CryptoLoanRepay stores crypto loan repayment data
type CryptoLoanRepay struct {
	LoanCoin            currency.Code `json:"loanCoin"`
	RemainingPrincipal  float64       `json:"remainingPrincipal,string"`
	RemainingInterest   float64       `json:"remainingInterest,string"`
	CollateralCoin      currency.Code `json:"collateralCoin"`
	RemainingCollateral float64       `json:"remainingCollateral,string"`
	CurrentLTV          float64       `json:"currentLTV,string"`
	RepayStatus         string        `json:"repayStatus"`
}

// CryptoLoanRepayHistoryItem stores crypto loan repayment history item data
type CryptoLoanRepayHistoryItem struct {
	LoanCoin         currency.Code        `json:"loanCoin"`
	RepayAmount      float64              `json:"repayAmount,string"`
	CollateralCoin   currency.Code        `json:"collateralCoin"`
	CollateralUsed   float64              `json:"collateralUsed,string"`
	CollateralReturn float64              `json:"collateralReturn,string"`
	RepayType        string               `json:"repayType"`
	RepayTime        convert.ExchangeTime `json:"repayTime"`
	OrderID          int64                `json:"orderId"`
}

// CryptoLoanRepayHistory stores crypto loan repayment history data
type CryptoLoanRepayHistory struct {
	Rows  []CryptoLoanRepayHistoryItem `json:"rows"`
	Total int64                        `json:"total"`
}

// CryptoLoanAdjustLTV stores crypto loan LTV adjustment data
type CryptoLoanAdjustLTV struct {
	LoanCoin       currency.Code `json:"loanCoin"`
	CollateralCoin currency.Code `json:"collateralCoin"`
	Direction      string        `json:"direction"`
	Amount         float64       `json:"amount,string"`
	CurrentLTV     float64       `json:"currentLTV,string"`
}

// CryptoLoanLTVAdjustmentItem stores crypto loan LTV adjustment item data
type CryptoLoanLTVAdjustmentItem struct {
	LoanCoin       currency.Code        `json:"loanCoin"`
	CollateralCoin currency.Code        `json:"collateralCoin"`
	Direction      string               `json:"direction"`
	Amount         float64              `json:"amount,string"`
	PreviousLTV    float64              `json:"preLTV,string"`
	AfterLTV       float64              `json:"afterLTV,string"`
	AdjustTime     convert.ExchangeTime `json:"adjustTime"`
	OrderID        int64                `json:"orderId"`
}

// CryptoLoanLTVAdjustmentHistory stores crypto loan LTV adjustment history data
type CryptoLoanLTVAdjustmentHistory struct {
	Rows  []CryptoLoanLTVAdjustmentItem `json:"rows"`
	Total int64                         `json:"total"`
}

// LoanableAssetItem stores loanable asset item data
type LoanableAssetItem struct {
	LoanCoin                             currency.Code `json:"loanCoin"`
	SevenDayHourlyInterestRate           float64       `json:"_7dHourlyInterestRate,string"`
	SevenDayDailyInterestRate            float64       `json:"_7dDailyInterestRate,string"`
	FourteenDayHourlyInterest            float64       `json:"_14dHourlyInterestRate,string"`
	FourteenDayDailyInterest             float64       `json:"_14dDailyInterestRate,string"`
	ThirtyDayHourlyInterest              float64       `json:"_30dHourlyInterestRate,string"`
	ThirtyDayDailyInterest               float64       `json:"_30dDailyInterestRate,string"`
	NinetyDayHourlyInterest              float64       `json:"_90dHourlyInterestRate,string"`
	NinetyDayDailyInterest               float64       `json:"_90dDailyInterestRate,string"`
	OneHundredAndEightyDayHourlyInterest float64       `json:"_180dHourlyInterestRate,string"`
	OneHundredAndEightyDayDailyInterest  float64       `json:"_180dDailyInterestRate,string"`
	MinimumLimit                         float64       `json:"minLimit,string"`
	MaximumLimit                         float64       `json:"maxLimit,string"`
	VIPLevel                             int64         `json:"vipLevel"`
}

// LoanableAssetsData stores loanable assets data
type LoanableAssetsData struct {
	Rows  []LoanableAssetItem `json:"rows"`
	Total int64               `json:"total"`
}

// CollateralAssetItem stores collateral asset item data
type CollateralAssetItem struct {
	CollateralCoin currency.Code `json:"collateralCoin"`
	InitialLTV     float64       `json:"initialLTV,string"`
	MarginCallLTV  float64       `json:"marginCallLTV,string"`
	LiquidationLTV float64       `json:"liquidationLTV,string"`
	MaxLimit       float64       `json:"maxLimit,string"`
	VIPLevel       int64         `json:"vipLevel"`
}

// CollateralAssetData stores collateral asset data
type CollateralAssetData struct {
	Rows  []CollateralAssetItem `json:"rows"`
	Total int64                 `json:"total"`
}

// CollateralRepayRate stores collateral repayment rate data
type CollateralRepayRate struct {
	LoanCoin       currency.Code `json:"loanCoin"`
	CollateralCoin currency.Code `json:"collateralCoin"`
	RepayAmount    float64       `json:"repayAmount,string"`
	Rate           float64       `json:"rate,string"`
}

// CustomiseMarginCallItem stores customise margin call item data
type CustomiseMarginCallItem struct {
	OrderID         int64                `json:"orderId"`
	CollateralCoin  currency.Code        `json:"collateralCoin"`
	PreMarginCall   float64              `json:"preMarginCall,string"`
	AfterMarginCall float64              `json:"afterMarginCall,string"`
	CustomiseTime   convert.ExchangeTime `json:"customizeTime"`
}

// CustomiseMarginCall stores customise margin call data
type CustomiseMarginCall struct {
	Rows  []CustomiseMarginCallItem `json:"rows"`
	Total int64                     `json:"total"`
}

// FlexibleLoanBorrow stores a flexible loan borrow
type FlexibleLoanBorrow struct {
	LoanCoin         currency.Code `json:"loanCoin"`
	LoanAmount       float64       `json:"loanAmount,string"`
	CollateralCoin   currency.Code `json:"collateralCoin"`
	CollateralAmount float64       `json:"collateralAmount,string"`
	Status           string        `json:"status"`
}

// FlexibleLoanOngoingOrderItem stores a flexible loan ongoing order item
type FlexibleLoanOngoingOrderItem struct {
	LoanCoin         currency.Code `json:"loanCoin"`
	TotalDebt        float64       `json:"totalDebt,string"`
	CollateralCoin   currency.Code `json:"collateralCoin"`
	CollateralAmount float64       `json:"collateralAmount,string"`
	CurrentLTV       float64       `json:"currentLTV,string"`
}

// FlexibleLoanOngoingOrder stores flexible loan ongoing orders
type FlexibleLoanOngoingOrder struct {
	Rows  []FlexibleLoanOngoingOrderItem `json:"rows"`
	Total int64                          `json:"total"`
}

// FlexibleLoanBorrowHistoryItem stores a flexible loan borrow history item
type FlexibleLoanBorrowHistoryItem struct {
	LoanCoin                currency.Code        `json:"loanCoin"`
	InitialLoanAmount       float64              `json:"initialLoanAmount,string"`
	CollateralCoin          currency.Code        `json:"collateralCoin"`
	InitialCollateralAmount float64              `json:"initialCollateralAmount,string"`
	BorrowTime              convert.ExchangeTime `json:"borrowTime"`
	Status                  string               `json:"status"`
}

// FlexibleLoanBorrowHistory stores flexible loan borrow history
type FlexibleLoanBorrowHistory struct {
	Rows  []FlexibleLoanBorrowHistoryItem `json:"rows"`
	Total int64                           `json:"total"`
}

// FlexibleLoanRepay stores a flexible loan repayment
type FlexibleLoanRepay struct {
	LoanCoin            currency.Code `json:"loanCoin"`
	CollateralCoin      currency.Code `json:"collateralCoin"`
	RemainingDebt       float64       `json:"remainingDebt,string"`
	RemainingCollateral float64       `json:"remainingCollateral,string"`
	FullRepayment       bool          `json:"fullRepayment"`
	CurrentLTV          float64       `json:"currentLTV,string"`
	RepayStatus         string        `json:"repayStatus"`
}

// FlexibleLoanRepayHistoryItem stores a flexible loan repayment history item
type FlexibleLoanRepayHistoryItem struct {
	LoanCoin         currency.Code        `json:"loanCoin"`
	RepayAmount      float64              `json:"repayAmount,string"`
	CollateralCoin   currency.Code        `json:"collateralCoin"`
	CollateralReturn float64              `json:"collateralReturn,string"`
	RepayStatus      string               `json:"repayStatus"`
	RepayTime        convert.ExchangeTime `json:"repayTime"`
}

// FlexibleLoanRepayHistory stores flexible loan repayment history
type FlexibleLoanRepayHistory struct {
	Rows  []FlexibleLoanRepayHistoryItem `json:"rows"`
	Total int64                          `json:"total"`
}

// FlexibleLoanAdjustLTV stores a flexible loan LTV adjustment
type FlexibleLoanAdjustLTV struct {
	LoanCoin       currency.Code `json:"loanCoin"`
	CollateralCoin currency.Code `json:"collateralCoin"`
	Direction      string        `json:"direction"`
	Amount         float64       `json:"amount,string"` // docs error: API actually returns "amount" instead of "adjustedAmount"
	CurrentLTV     float64       `json:"currentLTV,string"`
	Status         string        `json:"status"`
}

// FlexibleLoanLTVAdjustmentHistoryItem stores a flexible loan LTV adjustment history item
type FlexibleLoanLTVAdjustmentHistoryItem struct {
	LoanCoin         currency.Code        `json:"loanCoin"`
	CollateralCoin   currency.Code        `json:"collateralCoin"`
	Direction        string               `json:"direction"`
	CollateralAmount float64              `json:"collateralAmount,string"`
	PreviousLTV      float64              `json:"preLTV,string"`
	AfterLTV         float64              `json:"afterLTV,string"`
	AdjustTime       convert.ExchangeTime `json:"adjustTime"`
}

// FlexibleLoanLTVAdjustmentHistory stores flexible loan LTV adjustment history
type FlexibleLoanLTVAdjustmentHistory struct {
	Rows  []FlexibleLoanLTVAdjustmentHistoryItem `json:"rows"`
	Total int64                                  `json:"total"`
}

// FlexibleLoanAssetsDataItem stores a flexible loan asset data item
type FlexibleLoanAssetsDataItem struct {
	LoanCoin             currency.Code `json:"loanCoin"`
	FlexibleInterestRate float64       `json:"flexibleInterestRate,string"`
	FlexibleMinLimit     float64       `json:"flexibleMinLimit,string"`
	FlexibleMaxLimit     float64       `json:"flexibleMaxLimit,string"`
}

// FlexibleLoanAssetsData stores flexible loan asset data
type FlexibleLoanAssetsData struct {
	Rows  []FlexibleLoanAssetsDataItem `json:"rows"`
	Total int64                        `json:"total"`
}

// FlexibleCollateralAssetsDataItem stores a flexible collateral asset data item
type FlexibleCollateralAssetsDataItem struct {
	CollateralCoin currency.Code `json:"collateralCoin"`
	InitialLTV     float64       `json:"initialLTV,string"`
	MarginCallLTV  float64       `json:"marginCallLTV,string"`
	LiquidationLTV float64       `json:"liquidationLTV,string"`
	MaxLimit       float64       `json:"maxLimit,string"`
}

// FlexibleCollateralAssetsData stores flexible collateral asset data
type FlexibleCollateralAssetsData struct {
	Rows  []FlexibleCollateralAssetsDataItem `json:"rows"`
	Total int64                              `json:"total"`
}

// UFuturesOrderbook holds orderbook data for usdt assets
type UFuturesOrderbook struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType               string               `json:"e"`
		EventTime               convert.ExchangeTime `json:"E"`
		TransactionTime         convert.ExchangeTime `json:"T"`
		Symbol                  string               `json:"s"`
		FirstUpdateID           int64                `json:"U"`
		FinalUpdateID           int64                `json:"u"`
		FinalUpdateIDLastStream int64                `json:"pu"`
		Bids                    [][2]types.Number    `json:"b"`
		Asks                    [][2]types.Number    `json:"a"`
	} `json:"data"`
}

// UFuturesKline updates to the current klines/candlestick
type UFuturesKline struct {
	EventType string               `json:"e"`
	EventTime convert.ExchangeTime `json:"E"`
	Symbol    string               `json:"s"`
	KlineData struct {
		StartTime                convert.ExchangeTime `json:"t"`
		CloseTime                convert.ExchangeTime `json:"T"`
		Symbol                   string               `json:"s"`
		Interval                 string               `json:"i"`
		FirstTradeID             int64                `json:"f"`
		LastTradeID              int64                `json:"L"`
		OpenPrice                types.Number         `json:"o"`
		ClosePrice               types.Number         `json:"c"`
		HighPrice                types.Number         `json:"h"`
		LowPrice                 types.Number         `json:"l"`
		BaseVolume               types.Number         `json:"v"`
		NumberOfTrades           int64                `json:"n"`
		IsKlineClosed            bool                 `json:"x"`
		QuoteVolume              types.Number         `json:"q"`
		TakerBuyBaseAssetVolume  types.Number         `json:"V"`
		TakerBuyQuoteAssetVolume types.Number         `json:"Q"`
		B                        string               `json:"B"`
	} `json:"k"`
}

// FuturesMarkPrice represents usdt futures mark price and funding rate for a single symbol pushed every 3
type FuturesMarkPrice struct {
	EventType            string               `json:"e"`
	EventTime            convert.ExchangeTime `json:"E"`
	Symbol               string               `json:"s"`
	MarkPrice            types.Number         `json:"p"`
	IndexPrice           types.Number         `json:"i"`
	EstimatedSettlePrice types.Number         `json:"P"` // Estimated Settle Price, only useful in the last hour before the settlement starts
	FundingRate          types.Number         `json:"r"`
	NextFundingTime      convert.ExchangeTime `json:"T"`
}

// UFuturesAssetIndexUpdate holds asset index for multi-assets mode user
type UFuturesAssetIndexUpdate struct {
	EventType             string               `json:"e"`
	EventTime             convert.ExchangeTime `json:"E"`
	Symbol                string               `json:"s"`
	IndexPrice            types.Number         `json:"i"`
	BidBuffer             types.Number         `json:"b"`
	AskBuffer             types.Number         `json:"a"`
	BidRate               types.Number         `json:"B"`
	AskRate               types.Number         `json:"A"`
	AutoExchangeBidBuffer types.Number         `json:"q"`
	AutoExchangeAskbuffer types.Number         `json:"g"`
	AutoExchangeBidRate   types.Number         `json:"Q"`
	AutoExchangeAskRate   types.Number         `json:"G"`
}

// FuturesContractInfo contract info updates. bks field only shows up when bracket gets updated.
type FuturesContractInfo struct {
	EventType        string               `json:"e"`
	EventTime        convert.ExchangeTime `json:"E"`
	Symbol           string               `json:"s"`
	Pair             string               `json:"ps"`
	ContractType     string               `json:"ct"`
	DeliveryDateTime convert.ExchangeTime `json:"dt"`
	OnboardDateTime  convert.ExchangeTime `json:"ot"`
	ContractStatus   string               `json:"cs"`
	Brackets         []struct {
		NationalBracket      float64 `json:"bs"`
		BracketFloorNotional float64 `json:"bnf"`
		BracketNotionalCap   float64 `json:"bnc"`
		MaintenanceRatio     float64 `json:"mmr"`
		Cf                   float64 `json:"cf"`
		MinLeverage          float64 `json:"mi"`
		MaxLeverage          float64 `json:"ma"`
	} `json:"bks"`
}

// MarketLiquidationOrder all Liquidation Order Snapshot Streams push force liquidation order information for all symbols in the market.
type MarketLiquidationOrder struct {
	EventType string               `json:"e"`
	EventTime convert.ExchangeTime `json:"E"`
	Order     struct {
		Symbol                         string               `json:"s"`
		Side                           string               `json:"S"`
		OrderType                      string               `json:"o"`
		TimeInForce                    string               `json:"f"`
		OriginalQuantity               types.Number         `json:"q"`
		Price                          types.Number         `json:"p"`
		AveragePrice                   types.Number         `json:"ap"`
		OrderStatus                    string               `json:"X"`
		OrderLastFieldQuantity         types.Number         `json:"l"`
		OrderFilledAccumulatedQuantity types.Number         `json:"z"`
		OrderTradeTime                 convert.ExchangeTime `json:"T"`
	} `json:"o"`
}

// FuturesBookTicker update to the best bid or ask's price or quantity in real-time for a specified symbol.
type FuturesBookTicker struct {
	EventType         string               `json:"e"`
	OrderbookUpdateID int64                `json:"u"`
	EventTime         convert.ExchangeTime `json:"E"`
	TransactionTime   convert.ExchangeTime `json:"T"`
	Symbol            string               `json:"s"`
	BestBidPrice      types.Number         `json:"b"`
	BestBidQty        types.Number         `json:"B"`
	BestAskPrice      types.Number         `json:"a"`
	BestAskQty        types.Number         `json:"A"`

	// Pair added to coin marigined futures
	Pair string `json:"ps"`
}

// UFutureMarketTicker 24hr rolling window ticker statistics for all symbols.
type UFutureMarketTicker struct {
	EventType             string               `json:"e"`
	EventTime             convert.ExchangeTime `json:"E"`
	Symbol                string               `json:"s"`
	PriceChange           types.Number         `json:"p"`
	PriceChangePercent    types.Number         `json:"P"`
	WeightedAveragePrice  types.Number         `json:"w"`
	LastPrice             types.Number         `json:"c"`
	LastQuantity          types.Number         `json:"Q"`
	OpenPrice             types.Number         `json:"o"`
	HighPrice             types.Number         `json:"h"`
	LowPrice              types.Number         `json:"l"`
	TotalTradeBaseVolume  types.Number         `json:"v"`
	TotalQuoteAssetVolume types.Number         `json:"q"`
	OpenTime              convert.ExchangeTime `json:"O"`
	CloseTIme             convert.ExchangeTime `json:"C"`
	FirstTradeID          int64                `json:"F"`
	LastTradeID           int64                `json:"L"`
	TotalNumberOfTrades   int64                `json:"n"`
}

// FutureMiniTickerPrice holds market mini tickers stream
type FutureMiniTickerPrice struct {
	EventType  string               `json:"e"`
	EventTime  convert.ExchangeTime `json:"E"`
	Symbol     string               `json:"s"`
	ClosePrice types.Number         `json:"c"`
	OpenPrice  types.Number         `json:"o"`
	HighPrice  types.Number         `json:"h"`
	LowPrice   types.Number         `json:"l"`
	Volume     types.Number         `json:"v"`

	QuoteVolume types.Number `json:"q"` // Total traded base asset volume for Coin Margined Futures

	Pair string `json:"ps"`
}

// FuturesAggTrade aggregate trade streams push market trade
type FuturesAggTrade struct {
	EventType        string               `json:"e"`
	EventTime        convert.ExchangeTime `json:"E"`
	Symbol           string               `json:"s"`
	AggregateTradeID int64                `json:"a"`
	Price            types.Number         `json:"p"`
	Quantity         types.Number         `json:"q"`
	FirstTradeID     int64                `json:"f"`
	LastTradeID      int64                `json:"l"`
	TradeTime        convert.ExchangeTime `json:"T"`
	IsMaker          bool                 `json:"m"`
}

// FuturesDepthOrderbook represents bids and asks
type FuturesDepthOrderbook struct {
	EventType               string               `json:"e"`
	EventTime               convert.ExchangeTime `json:"E"`
	TransactionTime         convert.ExchangeTime `json:"T"`
	Symbol                  string               `json:"s"`
	FirstUpdateID           int64                `json:"U"`
	LastUpdateID            int64                `json:"u"`
	FinalUpdateIDLastStream int64                `json:"pu"`
	Bids                    [][]string           `json:"b"`
	Asks                    [][]string           `json:"a"`

	// Added for coin margined futures
	Pair string `json:"ps"`
}

// UFutureCompositeIndex represents symbols a composite index
type UFutureCompositeIndex struct {
	EventType   string               `json:"e"`
	EventTime   convert.ExchangeTime `json:"E"`
	Symbol      string               `json:"s"`
	Price       types.Number         `json:"p"`
	C           string               `json:"C"`
	Composition []struct {
		BaseAsset          string       `json:"b"`
		QuoteAsset         string       `json:"q"`
		WeightQuantity     types.Number `json:"w"`
		WeightInPercentage types.Number `json:"W"`
		IndexPrice         types.Number `json:"i"`
	} `json:"c"`
}

// FutureContinuousKline represents continuous kline data.
type FutureContinuousKline struct {
	EventType    string               `json:"e"`
	EventTime    convert.ExchangeTime `json:"E"`
	Pair         string               `json:"ps"`
	ContractType string               `json:"ct"`
	KlineData    struct {
		StartTime                convert.ExchangeTime `json:"t"`
		EndTime                  convert.ExchangeTime `json:"T"`
		Interval                 string               `json:"i"`
		FirstUpdateID            int64                `json:"f"`
		LastupdateID             int64                `json:"L"`
		OpenPrice                types.Number         `json:"o"`
		ClosePrice               types.Number         `json:"c"`
		HighPrice                types.Number         `json:"h"`
		LowPrice                 types.Number         `json:"l"`
		Volume                   types.Number         `json:"v"`
		NumberOfTrades           int64                `json:"n"`
		IsKlineClosed            bool                 `json:"x"`
		QuoteAssetVolume         types.Number         `json:"q"`
		TakerBuyVolume           types.Number         `json:"V"`
		TakerBuyQuoteAssetVolume types.Number         `json:"Q"`
		B                        string               `json:"B"`
	} `json:"k"`
}

// WebsocketActionResponse represents a response for websocket actions like "SET_PROPERTY", "LIST_SUBSCRIPTIONS" and others
type WebsocketActionResponse struct {
	Result []string `json:"result"`
	ID     int64    `json:"id"`
}

// RateLimitItem holds ratelimit information for endpoint calls.
type RateLimitItem struct {
	RateLimitType  string `json:"rateLimitType"`
	Interval       string `json:"interval"`
	IntervalNumber int64  `json:"intervalNum"`
	Limit          int64  `json:"limit"`
	Count          int64  `json:"count"`
}

// SymbolAveragePrice represents the average symbol price
type SymbolAveragePrice struct {
	PriceIntervalMins int64                `json:"mins"`
	Price             types.Number         `json:"price"`
	CloseTime         convert.ExchangeTime `json:"closeTime"`
}

// PriceChangeRequestParam holds request parameters for price change request parameters
type PriceChangeRequestParam struct {
	Symbol     string          `json:"symbol,omitempty"`
	Symbols    []currency.Pair `json:"symbols,omitempty"`
	Timezone   string          `json:"timeZone,omitempty"`
	TickerType string          `json:"type,omitempty"`
}

// PriceChanges holds a single or slice of WsTickerPriceChange instance into a new type.
type PriceChanges []PriceChangeStats

// WsRollingWindowPriceParams rolling window price change statistics request params
type WsRollingWindowPriceParams struct {
	Symbols            []currency.Pair `json:"symbols,omitempty"`
	WindowSizeDuration time.Duration   `json:"-"`
	WindowSize         string          `json:"windowSize,omitempty"`
	TickerType         string          `json:"type,omitempty"`
	Symbol             string          `json:"symbol,omitempty"`
}

// SymbolTickerItem holds symbol and price information
type SymbolTickerItem struct {
	Symbol string       `json:"symbol"`
	Price  types.Number `json:"price" `
}

// SymbolTickers holds symbol and price ticker information.
type SymbolTickers []SymbolTickerItem

// WsOrderbookTicker holds orderbook ticker information
type WsOrderbookTicker struct {
	Symbol   string       `json:"symbol"`
	BidPrice types.Number `json:"bidPrice"`
	BidQty   types.Number `json:"bidQty"`
	AskPrice types.Number `json:"askPrice"`
	AskQty   types.Number `json:"askQty"`
}

// WsOrderbookTickers represents an orderbook ticker information
type WsOrderbookTickers []WsOrderbookTicker

// APISignatureInfo holds API key and signature information
type APISignatureInfo struct {
	APIKey    string `json:"apiKey,omitempty"`
	Signature string `json:"signature,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// TradeOrderRequestParam new order request parameter
type TradeOrderRequestParam struct {
	APISignatureInfo
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	OrderType   string  `json:"type"`
	TimeInForce string  `json:"timeInForce"`
	Price       float64 `json:"price,omitempty,string"`
	Quantity    float64 `json:"quantity,omitempty,string"`
}

// QueryOrderParam represents an order querying parameters
type QueryOrderParam struct {
	APISignatureInfo
	Symbol            string `json:"symbol,omitempty"`
	OrderID           int64  `json:"orderId,omitempty"`
	OrigClientOrderID string `json:"origClientOrderId,omitempty"`
	RecvWindow        int64  `json:"recvWindow,omitempty"`

	NewClientOrderID   string `json:"newClientOrderId,omitempty"`
	CancelRestrictions string `json:"cancelRestrictions,omitempty"`
}

// WsCancelAndReplaceParam represents a cancel and replace request parameters
type WsCancelAndReplaceParam struct {
	APISignatureInfo
	Symbol        string `json:"symbol,omitempty"`
	CancelOrderID string `json:"cancelOrderId,omitempty"`

	// CancelReplaceMode possible values are 'STOP_ON_FAILURE', 'ALLOW_FAILURE'
	CancelReplaceMode         string  `json:"cancelReplaceMode,omitempty"`
	CancelNewClientOrderID    string  `json:"cancelNewClientOrderId,omitempty"`
	CancelOriginClientOrderID string  `json:"cancelOrigClientOrderId,omitempty"`
	Side                      string  `json:"side,omitempty"` // BUY or SELL
	Price                     float64 `json:"price,omitempty"`
	Quantity                  float64 `json:"quantity,omitempty"`
	OrderType                 string  `json:"type,omitempty"`
	TimeInForce               string  `json:"timeInForce,omitempty"`
	QuoteOrderQty             float64 `json:"quoteOrderQty,omitempty"`
	NewClientOrderID          string  `json:"newClientOrderId,omitempty"`

	// Select response format: ACK, RESULT, FULL.
	NewOrderRespType string  `json:"newOrderRespType,omitempty"` // Select response format: ACK, RESULT, FULL. MARKET and LIMIT orders produce FULL response by default, other order types default to ACK.
	StopPrice        float64 `json:"stopPrice,omitempty"`
	TrailingDelta    float64 `json:"trailingDelta,omitempty"`
	IcebergQty       float64 `json:"icebergQty,omitempty"`
	StrategyID       int64   `json:"strategyId,omitempty"`

	// Values smaller than 1000000 are reserved and cannot be used.
	StrategyType int64 `json:"strategyType,omitempty"`

	// The possible supported values are EXPIRE_TAKER, EXPIRE_MAKER, EXPIRE_BOTH, NONE.
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty"`

	// Supported values:
	// ONLY_NEW - Cancel will succeed if the order status is NEW.
	// ONLY_PARTIALLY_FILLED - Cancel will succeed if order status is PARTIALLY_FILLED. For more information please refer to Regarding cancelRestrictions.
	CancelRestrictions string `json:"cancelRestrictions,omitempty"`
	RecvWindow         int64  `json:"recvWindow,omitempty"`
}

// PlaceOCOOrderParam holds a request parameters for one-cancel-other orders
type PlaceOCOOrderParam struct {
	APISignatureInfo
	Symbol               string  `json:"symbol,omitempty"`
	Side                 string  `json:"side,omitempty"`
	Price                float64 `json:"price,omitempty"`
	Quantity             float64 `json:"quantity,omitempty"`
	ListClientOrderID    string  `json:"listClientOrderId,omitempty"`
	LimitClientOrderID   string  `json:"limitClientOrderId,omitempty"`
	LimitIcebergQty      float64 `json:"limitIcebergQty,omitempty"`
	LimitStrategyID      string  `json:"limitStrategyId,omitempty"`
	LimitStrategyType    string  `json:"limitStrategyType,omitempty"`
	StopPrice            float64 `json:"stopPrice,omitempty"`
	TrailingDelta        int64   `json:"trailingDelta,omitempty"`
	StopClientOrderID    string  `json:"stopClientOrderId,omitempty"`
	StopLimitPrice       float64 `json:"stopLimitPrice,omitempty"`
	StopLimitTimeInForce string  `json:"stopLimitTimeInForce,omitempty"`
	StopIcebergQty       float64 `json:"stopIcebergQty,omitempty"`
	StopStrategyID       string  `json:"stopStrategyId,omitempty"`
	StopStrategyType     string  `json:"stopStrategyType,omitempty"`
	NewOrderRespType     string  `json:"newOrderRespType,omitempty"`

	// The allowed enums is dependent on what is configured on the symbol. The possible supported values are 'EXPIRE_TAKER', 'EXPIRE_MAKER', 'EXPIRE_BOTH', 'NONE'.
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty"`
	RecvWindow              int64  `json:"recvWindow,omitempty"`
}

// TradeOrderResponse holds response for trade order.
type TradeOrderResponse struct {
	Symbol        string               `json:"symbol"`
	OrderID       int64                `json:"orderId"`
	OrderListID   int64                `json:"orderListId"`
	ClientOrderID string               `json:"clientOrderId"`
	TransactTime  convert.ExchangeTime `json:"transactTime"`
}

// FuturesAuthenticationResp holds authentication.
type FuturesAuthenticationResp struct {
	APIKey           string               `json:"apiKey"`
	AuthorizedSince  int64                `json:"authorizedSince"`
	ConnectedSince   int64                `json:"connectedSince"`
	ReturnRateLimits bool                 `json:"returnRateLimits"`
	ServerTime       convert.ExchangeTime `json:"serverTime"`
}

// WsCancelAndReplaceTradeOrderResponse holds a response from cancel and replacing an existing trade order
type WsCancelAndReplaceTradeOrderResponse struct {
	CancelResult   string `json:"cancelResult"`
	NewOrderResult string `json:"newOrderResult"`
	CancelResponse struct {
		Symbol                  string               `json:"symbol"`
		OrigClientOrderID       string               `json:"origClientOrderId"`
		OrderID                 int64                `json:"orderId"`
		OrderListID             int64                `json:"orderListId"`
		ClientOrderID           string               `json:"clientOrderId"`
		TransactTime            convert.ExchangeTime `json:"transactTime"`
		Price                   types.Number         `json:"price"`
		OrigQty                 types.Number         `json:"origQty"`
		ExecutedQty             types.Number         `json:"executedQty"`
		CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
		Status                  string               `json:"status"`
		TimeInForce             string               `json:"timeInForce"`
		Type                    string               `json:"type"`
		Side                    string               `json:"side"`
		SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	} `json:"cancelResponse"`
	NewOrderResponse struct {
		Symbol                  string               `json:"symbol"`
		OrderID                 int64                `json:"orderId"`
		OrderListID             int64                `json:"orderListId"`
		ClientOrderID           string               `json:"clientOrderId"`
		TransactTime            convert.ExchangeTime `json:"transactTime"`
		Price                   types.Number         `json:"price"`
		OrigQty                 types.Number         `json:"origQty"`
		ExecutedQty             types.Number         `json:"executedQty"`
		CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
		Status                  string               `json:"status"`
		TimeInForce             string               `json:"timeInForce"`
		Type                    string               `json:"type"`
		Side                    string               `json:"side"`
		WorkingTime             convert.ExchangeTime `json:"workingTime"`
		Fills                   []any                `json:"fills"`
		SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	} `json:"newOrderResponse"`
}

// WsCancelOrder holds a response data for canceling an open order.
type WsCancelOrder struct {
	Symbol                  string               `json:"symbol"`
	OrigClientOrderID       string               `json:"origClientOrderId,omitempty"`
	OrderID                 int64                `json:"orderId,omitempty"`
	OrderListID             int64                `json:"orderListId"`
	ClientOrderID           string               `json:"clientOrderId,omitempty"`
	TransactTime            convert.ExchangeTime `json:"transactTime,omitempty"`
	Price                   types.Number         `json:"price,omitempty"`
	OrigQty                 types.Number         `json:"origQty,omitempty"`
	ExecutedQty             types.Number         `json:"executedQty,omitempty"`
	CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty,omitempty"`
	Status                  string               `json:"status,omitempty"`
	TimeInForce             string               `json:"timeInForce,omitempty"`
	Type                    string               `json:"type,omitempty"`
	Side                    string               `json:"side,omitempty"`
	StopPrice               types.Number         `json:"stopPrice,omitempty"`
	IcebergQty              types.Number         `json:"icebergQty,omitempty"`
	StrategyID              int64                `json:"strategyId,omitempty"`
	StrategyType            int64                `json:"strategyType,omitempty"`
	SelfTradePreventionMode string               `json:"selfTradePreventionMode,omitempty"`
	ContingencyType         string               `json:"contingencyType,omitempty"`
	ListStatusType          string               `json:"listStatusType,omitempty"`
	ListOrderStatus         string               `json:"listOrderStatus,omitempty"`
	ListClientOrderID       string               `json:"listClientOrderId,omitempty"`
	TransactionTime         convert.ExchangeTime `json:"transactionTime,omitempty"`
	Orders                  []struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
	} `json:"orders,omitempty"`
	OrderReports []OrderReportItem `json:"orderReports,omitempty"`
}

// OrderReportItem represents a single order report instance.
type OrderReportItem struct {
	Symbol                  string               `json:"symbol"`
	OrigClientOrderID       string               `json:"origClientOrderId"`
	OrderID                 int64                `json:"orderId"`
	OrderListID             int64                `json:"orderListId"`
	ClientOrderID           string               `json:"clientOrderId"`
	TransactTime            convert.ExchangeTime `json:"transactTime"`
	Price                   types.Number         `json:"price"`
	OrigQty                 types.Number         `json:"origQty"`
	ExecutedQty             types.Number         `json:"executedQty"`
	CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
	Status                  string               `json:"status"`
	TimeInForce             string               `json:"timeInForce"`
	Type                    string               `json:"type"`
	Side                    string               `json:"side"`
	StopPrice               types.Number         `json:"stopPrice,omitempty"`
	SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
}

// OCOOrder represents a one-close-other order type.
type OCOOrder struct {
	OrderListID       int64                `json:"orderListId"`
	ContingencyType   string               `json:"contingencyType"`
	ListStatusType    string               `json:"listStatusType"`
	ListOrderStatus   string               `json:"listOrderStatus"`
	ListClientOrderID string               `json:"listClientOrderId"`
	TransactionTime   convert.ExchangeTime `json:"transactionTime"`
	Symbol            string               `json:"symbol"`
	Orders            []struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
	} `json:"orders"`
	OrderReports []struct {
		Symbol                  string               `json:"symbol"`
		OrderID                 int64                `json:"orderId"`
		OrderListID             int64                `json:"orderListId"`
		ClientOrderID           string               `json:"clientOrderId"`
		TransactTime            convert.ExchangeTime `json:"transactTime"`
		Price                   types.Number         `json:"price"`
		OrigQty                 types.Number         `json:"origQty"`
		ExecutedQty             types.Number         `json:"executedQty"`
		CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
		Status                  string               `json:"status"`
		TimeInForce             string               `json:"timeInForce"`
		Type                    string               `json:"type"`
		Side                    string               `json:"side"`
		StopPrice               types.Number         `json:"stopPrice,omitempty"`
		WorkingTime             convert.ExchangeTime `json:"workingTime"`
		SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	} `json:"orderReports"`
}

// OCOOrderInfo represents OCO order information.
type OCOOrderInfo struct {
	OrderListID       int64                `json:"orderListId"`
	ContingencyType   string               `json:"contingencyType"`
	ListStatusType    string               `json:"listStatusType"`
	ListOrderStatus   string               `json:"listOrderStatus"`
	ListClientOrderID string               `json:"listClientOrderId"`
	TransactionTime   convert.ExchangeTime `json:"transactionTime"`
	Symbol            string               `json:"symbol"`
	Orders            []struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
	} `json:"orders"`

	// returned when cancelling the order
	OrderReports []struct {
		Symbol                  string               `json:"symbol"`
		OrderID                 int64                `json:"orderId"`
		OrderListID             int64                `json:"orderListId"`
		ClientOrderID           string               `json:"clientOrderId"`
		TransactTime            convert.ExchangeTime `json:"transactTime"`
		Price                   types.Number         `json:"price"`
		OrigQty                 types.Number         `json:"origQty"`
		ExecutedQty             types.Number         `json:"executedQty"`
		CummulativeQuoteQty     types.Number         `json:"cummulativeQuoteQty"`
		Status                  string               `json:"status"`
		TimeInForce             string               `json:"timeInForce"`
		Type                    string               `json:"type"`
		Side                    string               `json:"side"`
		StopPrice               types.Number         `json:"stopPrice,omitempty"`
		SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	} `json:"orderReports"`
}

// WsOSRPlaceOrderParams holds request parameters for placing OSR orders.
type WsOSRPlaceOrderParams struct {
	APISignatureInfo
	Symbol           string  `json:"symbol,omitempty"`
	Side             string  `json:"side,omitempty"`
	OrderType        string  `json:"type,omitempty"`
	TimeInForce      string  `json:"timeInForce,omitempty"`
	Price            float64 `json:"price,omitempty"`
	Quantity         float64 `json:"quantity,omitempty"`
	NewClientOrderID string  `json:"newClientOrderId,omitempty"`

	// Select response format: ACK, RESULT, FULL.
	// MARKET and LIMIT orders use FULL by default.
	NewOrderRespType string  `json:"newOrderRespType,omitempty"`
	IcebergQty       float64 `json:"icebergQty,omitempty"`
	StrategyID       int64   `json:"strategyId,omitempty"`
	StrategyType     string  `json:"strategyType,omitempty"`

	// The allowed enums is dependent on what is configured on the symbol. The possible supported values are EXPIRE_TAKER, EXPIRE_MAKER, EXPIRE_BOTH, NONE.
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty"`
	RecvWindow              string `json:"recvWindow,omitempty"`
}

// OSROrder represents a request parameters for Smart Order Routing (SOR)
type OSROrder struct {
	Symbol              string               `json:"symbol"`
	OrderID             int64                `json:"orderId"`
	OrderListID         int64                `json:"orderListId"`
	ClientOrderID       string               `json:"clientOrderId"`
	TransactTime        int64                `json:"transactTime"`
	Price               types.Number         `json:"price"`
	OrigQty             types.Number         `json:"origQty"`
	ExecutedQty         types.Number         `json:"executedQty"`
	CummulativeQuoteQty types.Number         `json:"cummulativeQuoteQty"`
	Status              string               `json:"status"`
	TimeInForce         string               `json:"timeInForce"`
	Type                string               `json:"type"`
	Side                string               `json:"side"`
	WorkingTime         convert.ExchangeTime `json:"workingTime"`
	Fills               []struct {
		MatchType       string       `json:"matchType"`
		Price           types.Number `json:"price"`
		Qty             types.Number `json:"qty"`
		Commission      string       `json:"commission"`
		CommissionAsset string       `json:"commissionAsset"`
		TradeID         int64        `json:"tradeId"`
		AllocID         int64        `json:"allocId"`
	} `json:"fills"`
	WorkingFloor            string `json:"workingFloor"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	UsedSor                 bool   `json:"usedSor"`
}

// AccountOrderRequestParam retrieves an account order history parameters
type AccountOrderRequestParam struct {
	APISignatureInfo
	EndTime    int64  `json:"endTime,omitempty"`
	Limit      int64  `json:"limit,omitempty"`
	OrderID    int64  `json:"orderId,omitempty"` // Order ID to begin at
	RecvWindow int64  `json:"recvWindow,omitempty"`
	StartTime  int64  `json:"startTime,omitempty"`
	Symbol     string `json:"symbol"`

	// for requesting trades
	FromID int64 `json:"fromId,omitempty"`
}

// TradeHistory holds trade history information.
type TradeHistory struct {
	Symbol          string               `json:"symbol"`
	ID              int                  `json:"id"`
	OrderID         int64                `json:"orderId"`
	OrderListID     int                  `json:"orderListId"`
	Price           types.Number         `json:"price"`
	Qty             types.Number         `json:"qty"`
	QuoteQty        types.Number         `json:"quoteQty"`
	Commission      string               `json:"commission"`
	CommissionAsset string               `json:"commissionAsset"`
	Time            convert.ExchangeTime `json:"time"`
	IsBuyer         bool                 `json:"isBuyer"`
	IsMaker         bool                 `json:"isMaker"`
	IsBestMatch     bool                 `json:"isBestMatch"`
}

// SelfTradePrevention represents a self-trade prevention instance.
type SelfTradePrevention struct {
	Symbol                  string               `json:"symbol"`
	PreventedMatchID        int64                `json:"preventedMatchId"`
	TakerOrderID            int64                `json:"takerOrderId"`
	MakerOrderID            int64                `json:"makerOrderId"`
	TradeGroupID            int64                `json:"tradeGroupId"`
	SelfTradePreventionMode string               `json:"selfTradePreventionMode"`
	Price                   types.Number         `json:"price"`
	MakerPreventedQuantity  types.Number         `json:"makerPreventedQuantity"`
	TransactTime            convert.ExchangeTime `json:"transactTime"`
}

// SORReplacements represents response instance after for Smart Order Routing(SOR) order placement.
type SORReplacements struct {
	Symbol          string               `json:"symbol"`
	AllocationID    int64                `json:"allocationId"`
	AllocationType  string               `json:"allocationType"`
	OrderID         int64                `json:"orderId"`
	OrderListID     int64                `json:"orderListId"`
	Price           types.Number         `json:"price"`
	Quantity        types.Number         `json:"qty"`
	QuoteQty        types.Number         `json:"quoteQty"`
	Commission      string               `json:"commission"`
	CommissionAsset string               `json:"commissionAsset"`
	Time            convert.ExchangeTime `json:"time"`
	IsBuyer         bool                 `json:"isBuyer"`
	IsMaker         bool                 `json:"isMaker"`
	IsAllocator     bool                 `json:"isAllocator"`
}

// CommissionRateInto represents commission rate info.
type CommissionRateInto struct {
	Symbol             string          `json:"symbol"`
	StandardCommission *CommissionInfo `json:"standardCommission"`
	TaxCommission      *CommissionInfo `json:"taxCommission"`
	Discount           struct {
		EnabledForAccount bool   `json:"enabledForAccount"`
		EnabledForSymbol  bool   `json:"enabledForSymbol"`
		DiscountAsset     string `json:"discountAsset"`
		Discount          string `json:"discount"`
	} `json:"discount"`
}

// CommissionInfo holds tax and standard
type CommissionInfo struct {
	Maker  string `json:"maker"`
	Taker  string `json:"taker"`
	Buyer  string `json:"buyer"`
	Seller string `json:"seller"`
}
