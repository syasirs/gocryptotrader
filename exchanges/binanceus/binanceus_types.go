package binanceus

import (
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

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
)

// ExchangeInfo holds the full exchange information type
type ExchangeInfo struct {
	Code       int       `json:"code"`
	Msg        string    `json:"msg"`
	Timezone   string    `json:"timezone"`
	Servertime time.Time `json:"serverTime"`
	RateLimits []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		Limit         int    `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters interface{} `json:"exchangeFilters"`
	Symbols         []struct {
		Symbol                     string   `json:"symbol"`
		Status                     string   `json:"status"`
		BaseAsset                  string   `json:"baseAsset"`
		BaseAssetPrecision         int      `json:"baseAssetPrecision"`
		QuoteAsset                 string   `json:"quoteAsset"`
		QuotePrecision             int      `json:"quotePrecision"`
		OrderTypes                 []string `json:"orderTypes"`
		IcebergAllowed             bool     `json:"icebergAllowed"`
		OCOAllowed                 bool     `json:"ocoAllowed"`
		QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"`
		IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
		Filters                    []struct {
			FilterType          string  `json:"filterType"`
			MinPrice            float64 `json:"minPrice,string"`
			MaxPrice            float64 `json:"maxPrice,string"`
			TickSize            float64 `json:"tickSize,string"`
			MultiplierUp        float64 `json:"multiplierUp,string"`
			MultiplierDown      float64 `json:"multiplierDown,string"`
			AvgPriceMinutes     int64   `json:"avgPriceMins"`
			MinQty              float64 `json:"minQty,string"`
			MaxQty              float64 `json:"maxQty,string"`
			StepSize            float64 `json:"stepSize,string"`
			MinNotional         float64 `json:"minNotional,string"`
			ApplyToMarket       bool    `json:"applyToMarket"`
			Limit               int64   `json:"limit"`
			MaxNumAlgoOrders    int64   `json:"maxNumAlgoOrders"`
			MaxNumIcebergOrders int64   `json:"maxNumIcebergOrders"`
			MaxNumOrders        int64   `json:"maxNumOrders"`
		} `json:"filters"`
		Permissions []string `json:"permissions"`
	} `json:"symbols"`
}

// RecentTradeRequestParams represents Klines request data.
type RecentTradeRequestParams struct {
	Symbol currency.Pair `json:"symbol"` // Required field. example LTCBTC, BTCUSDT
	Limit  int           `json:"limit"`  // Default 500; max 1000.
}

// RecentTrade holds recent trade data
type RecentTrade struct {
	ID           int64     `json:"id"`
	Price        float64   `json:"price,string"`
	Quantity     float64   `json:"qty,string"`
	Time         time.Time `json:"time"`
	IsBuyerMaker bool      `json:"isBuyerMaker"`
	IsBestMatch  bool      `json:"isBestMatch"`
}

type HistoricalTradeParams struct {
	Symbol string `json:"symbol"`  // Required field. example LTCBTC, BTCUSDT
	Limit  int    `json:"limit"`   // Default 500; max 1000.
	FromID uint64 `json:"from_id"` // Optional Field. Specifies the trade ID to fetch most recent trade histories from
}

// HistoricalTrade holds recent trade data
type HistoricalTrade struct {
	ID            int64     `json:"id"`
	Price         float64   `json:"price,string"`
	Quantity      float64   `json:"qty,string"`
	QuoteQuantity float64   `json:"quoteQty,string"`
	Time          time.Time `json:"time"`
	IsBuyerMaker  bool      `json:"isBuyerMaker"`
	IsBestMatch   bool      `json:"isBestMatch"`
}

// AggregatedTradeRequestParams holds request params
type AggregatedTradeRequestParams struct {
	Symbol currency.Pair // Required field; example LTCBTC, BTCUSDT
	// The first trade to retrieve
	FromID int64
	// The API seems to accept (start and end time) or FromID and no other combinations
	StartTime uint64
	EndTime   uint64
	// Default 500; max 1000.
	Limit int
}

// toTradeData this method converts the AggregatedTrade data into an instance of trade.Data...
func (a *AggregatedTrade) toTradeData(p currency.Pair, exchange string, aType asset.Item) *trade.Data {
	return &trade.Data{
		CurrencyPair: p,
		TID:          strconv.FormatInt(a.ATradeID, 10),
		Amount:       a.Quantity,
		Exchange:     exchange,
		Price:        a.Price,
		Timestamp:    a.TimeStamp,
		AssetType:    aType,
		Side:         order.AnySide,
	}
}

// AggregatedTrade holds aggregated trade information
type AggregatedTrade struct {
	ATradeID       int64     `json:"a"`
	Price          float64   `json:"p,string"`
	Quantity       float64   `json:"q,string"`
	FirstTradeID   int64     `json:"f"`
	LastTradeID    int64     `json:"l"`
	TimeStamp      time.Time `json:"T"`
	Maker          bool      `json:"m"`
	BestMatchPrice bool      `json:"M"`
}

// OrderBookDataRequestParams represents Klines request data.
type OrderBookDataRequestParams struct {
	Symbol currency.Pair `json:"symbol"` // Required field; example LTCBTC,BTCUSDT
	Limit  int           `json:"limit"`  // Default 100; max 1000. Valid limits:[5, 10, 20, 50, 100, 500, 1000]
}

// OrderbookItem stores an individual orderbook item
type OrderbookItem struct {
	Price    float64
	Quantity float64
}

// OrderBookData is resp data from orderbook endpoint
type OrderBookData struct {
	LastUpdateID int64       `json:"lastUpdateId"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
}

// OrderBook actual structured data that can be used for orderbook
type OrderBook struct {
	Symbol       string
	LastUpdateID int64
	Code         int
	Msg          string
	Bids         []OrderbookItem
	Asks         []OrderbookItem
}

// KlinesRequestParams represents Klines request data.
type KlinesRequestParams struct {
	Symbol    currency.Pair // Required field; example LTCBTC, BTCUSDT
	Interval  string        // Time interval period
	Limit     int           // Default 500; max 500.
	StartTime time.Time
	EndTime   time.Time
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

// SymbolPrice represents a symbol and it's price.
type SymbolPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

// SymbolPrices lis tof Symbol Price
type SymbolPrices []SymbolPrice

// AveragePrice holds current average symbol price
type AveragePrice struct {
	Mins  int64   `json:"mins"`
	Price float64 `json:"price,string"`
}

// BestPrice holds best price data
type BestPrice struct {
	Symbol   string  `json:"symbol"`
	BidPrice float64 `json:"bidPrice,string"`
	BidQty   float64 `json:"bidQty,string"`
	AskPrice float64 `json:"askPrice,string"`
	AskQty   float64 `json:"askQty,string"`
}

// PriceChangeStats contains statistics for the last 24 hours trade
type PriceChangeStats struct {
	Symbol             string    `json:"symbol"`
	PriceChange        float64   `json:"priceChange,string"`
	PriceChangePercent float64   `json:"priceChangePercent,string"`
	WeightedAvgPrice   float64   `json:"weightedAvgPrice,string"`
	PrevClosePrice     float64   `json:"prevClosePrice,string"`
	LastPrice          float64   `json:"lastPrice,string"`
	LastQty            float64   `json:"lastQty,string"`
	BidPrice           float64   `json:"bidPrice,string"`
	AskPrice           float64   `json:"askPrice,string"`
	OpenPrice          float64   `json:"openPrice,string"`
	HighPrice          float64   `json:"highPrice,string"`
	LowPrice           float64   `json:"lowPrice,string"`
	Volume             float64   `json:"volume,string"`
	QuoteVolume        float64   `json:"quoteVolume,string"`
	OpenTime           time.Time `json:"openTime"`
	CloseTime          time.Time `json:"closeTime"`
	FirstID            int64     `json:"firstId"`
	LastID             int64     `json:"lastId"`
	Count              int64     `json:"count"`
}

// Response holds basic binance api response data
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
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
	UpdateTime       time.Time `json:"updateTime"`
	AccounType       string    `json:"spot"`
	Balances         []Balance `json:"balances"`
	Permissions      []string  `json:"permissions"`
}

// Balance holds query order data
type Balance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

// AccountStatusResponse holds informations related to the
// User Account status information request
type AccountStatusResponse struct {
	Msg     string   `json:"msg"`
	Success bool     `json:"success"`
	Objs    []string `json:"objs,omitempty"`
}

// TradeStatus represents trade status and holds list of trade status indicator Item instances.
type TradeStatus struct {
	IsLocked           bool                                  `json:"isLocked"`
	PlannedRecoverTime uint                                  `json:"plannedRecoverTime"`
	TriggerCondition   map[string]uint                       `json:"triggerCondition"`
	Indicators         map[string]TradingStatusIndicatorItem `json:"indicators"`
	UpdateTime         time.Time                             `json:"updateTime"`
}

// TradingStatusIndicatorItem represents Trade Status Indication
type TradingStatusIndicatorItem struct {
	IndicatorSymbol  string  `json:"i"`
	CountOfAllOrders float32 `json:"c"`
	CurrentValue     float32 `json:"v"`
	TriggerValue     float32 `json:"t"`
}

// TradeFee represents the symbol and corresponding maker and taker trading fee value.
type TradeFee struct {
	Symbol string  `json:"symbol"`
	Maker  float64 `json:"maker"`
	Taker  float64 `json:"taker"`
}

// TradeFeeList list of trading fee for different trade symbols.
type TradeFeeList struct {
	TradeFee []TradeFee `json:"tradeFee"`
	Success  bool       `json:"success,omitempty"`
}

// AssetHistory holds the asset type and translation info
type AssetHistory struct {
	Amount  float64 `json:"amount,string"` // Amount
	Asset   string  `json:"asset"`         // Asset Type eg. BHFT
	DivTime uint64  `json:"divTime"`       // DivTime
	EnInfo  string  `json:"enInfo"`        //
	TranID  uint64  `json:"tranId"`        // Transaction ID
}

// AssetDictributionHistories this endpoint to query asset distribution records,
// including for staking, referrals and airdrops etc.
type AssetDistributionHistories struct {
	Rows  []*AssetHistory `json:"rows"`
	Total uint            `json:"total"`
}

// SubAccount  holds a single sub account instance in a Binance US account.
// including the email and related information related to it.
type SubAccount struct {
	Email      string    `json:"email"`
	Status     bool      `json:"status"`
	Activated  bool      `json:"activated"`
	Mobile     string    `json:"mobile"`
	GAuth      bool      `json:"gAuth"`
	CreateTime time.Time `json:"createTime"`
}

// TransferHistory a single asset transfer history between Sub accounts
type TransferHistory struct {
	Fron      string    `json:"from"`
	To        string    `json:"to"`
	Asset     string    `json:"asset"`
	Qty       uint      `json:"qty,string"`
	TimeStamp time.Time `json:"time"`
}

// SubAccountTransferRequestParams a argument varaibles holder used to transfer an asset from one account to another subaccount
// this account has to be present in the sub accounts list information.
type SubaccountTransferRequestParams struct {
	FromEmail  string  // Mandatory
	ToEmail    string  // Mandatory
	Asset      string  // Mandatory
	Amount     float64 // Mandatory
	RecvWindow uint64
}

// SubAccountTransferResponse repsents a suabccount transffer history
// having the transaction id which is to be returned due to the transfer
type SubaccountTransferResponse struct {
	Success bool   `json:"success"`
	TxnID   uint64 `json:"txnId,string"`
}

// SubaccountAsset holds asset informations.
type AssetInfo struct {
	Asset  string `json:"asset"`
	Free   uint64 `json:"free"`
	Locked uint64 `json:"locked"`
}

// SubAccountAssets holds all the balance and email of a subaccount
type SubAccountAssets struct {
	Balances        []AssetInfo `json:"balances"`
	Success         bool        `json:"success"`
	SubaccountEmail string      `json:"email,omitempty"`
}

// OrderRateLimit
type OrderRateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   uint   `json:"intervalNum"`
	Limit         uint   `json:"limit"`
	Count         uint   `json:"count"`
}

// RequestParamsOrderType trade order type
type RequestParamsOrderType string

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

// NewOrderRequest request type
type NewOrderRequest struct {
	Symbol           currency.Pair
	Side             string
	TradeType        RequestParamsOrderType
	TimeInForce      RequestParamsTimeForceType
	Quantity         float64
	QuoteOrderQty    float64
	Price            float64
	NewClientOrderID string
	StopPrice        float64 // Used with STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, and TAKE_PROFIT_LIMIT orders.
	IcebergQty       float64 // Used with LIMIT, STOP_LOSS_LIMIT, and TAKE_PROFIT_LIMIT to create an iceberg order.
	NewOrderRespType string
}

// NewOrderResponse is the return structured response from the exchange
type NewOrderResponse struct {
	Symbol          string    `json:"symbol"`
	OrderID         int64     `json:"orderId"`
	OrderListID     int8      `json:"orderListId"`
	ClientOrderID   string    `json:"clientOrderId"`
	TransactionTime time.Time `json:"transactTime"`
	Price           float64   `json:"price,string"`
	OrigQty         float64   `json:"origQty,string"`
	ExecutedQty     float64   `json:"executedQty,string"`
	// The cumulative amount of the quote that has been spent (with a BUY order) or received (with a SELL order).
	CumulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"`
	Status             string  `json:"status"`
	TimeInForce        string  `json:"timeInForce"`
	Type               string  `json:"type"`
	Side               string  `json:"side"`
	// --
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	// --
	Fills []struct {
		Price           float64 `json:"price,string"`
		Qty             float64 `json:"qty,string"`
		Commission      float64 `json:"commission,string"`
		CommissionAsset string  `json:"commissionAsset"`
	} `json:"fills"`
}

// CommonOrder instance holds the order informations common to both
// for Order and OrderReportItem
type CommonOrder struct {
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	OrderListID   int8   `json:"orderListId"`
	ClientOrderID string `json:"clientOrderId"`

	Price               float64 `json:"price,string"`
	OrigQty             float64 `json:"origQty,string"`
	ExecutedQty         float64 `json:"executedQty,string"`
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	StopPrice           float64 `json:"stopPrice,string"`
}

// Order struct represents an ordinary order response.
type Order struct {
	CommonOrder
	IcebergQty        float64   `json:"icebergQty,string"`
	Time              time.Time `json:"time"`
	UpdateTime        time.Time `json:"updateTime"`
	IsWorking         bool      `json:"isWorking"`
	OrigQuoteOrderQty float64   `json:"origQuoteOrderQty"`
}

// OrderReportItem this is used by the OCO order creating response
type OCOOrderReportItem struct {
	CommonOrder
	TransactionTime time.Time `json:"transactionTime"`
}

// GetOrderRequestParams this struct will be used to get a
// order and its related information
type OrderRequestParams struct {
	Symbol            string `json:"symbol"` // REQUIRED
	OrderID           uint64 `json:"orderId"`
	OrigClientOrderId string `json:"origClientOrderId"`
	RecvWindow        uint
}

// CancelOrderRequestParams this struct will be used as a parameter for
// cancel order method.
type CancelOrderRequestParams struct {
	Symbol            currency.Pair
	OrderID           uint64
	OrigClientOrderID string
	NewClientOrderID  string
	RecvWindow        uint
}

// GetTradesParams  request param to get the trade history
type GetTradesParams struct {
	Symbol     string     `json:"symbol"`
	OrderID    uint64     `json:"orderId"`
	StartTime  *time.Time `json:"startTime"`
	EndTime    *time.Time `json:"endTime"`
	FromID     uint64     `json:"fromId"`
	Limit      uint       `json:"limit"`
	RecvWindow uint       `json:"recvWindow"`
}

// Trade this struct represents a trade response.
type Trade struct {
	Symbol          string    `json:"symbol"`
	ID              uint64    `json:"id"`
	OrderID         uint64    `json:"orderId"`
	OrderListId     int       `json:"orderListId"`
	Price           float64   `json:"price"`
	Qty             float64   `json:"qty"`
	QuoteQty        float64   `json:"quoteQty"`
	Commission      float64   `json:"commission"`
	CommissionAsset float64   `json:"commissionAsset"`
	Time            time.Time `json:"time"`
	IsBuyer         bool      `json:"isBuyer"`
	IsMaker         bool      `json:"isMaker"`
	IsBestMatch     bool      `json:"isBestMatch"`
}

// OCOOrderParams
// One -cancel-the-other order creation input Parameter
type OCOOrderInputParams struct {
	Symbol               string  `json:"symbol"`    // Required
	StopPrice            float64 `json:"stopPrice"` // Required
	Side                 string  `json:"side"`      // Required
	Quantity             float64 `json:"quantity"`  // Required
	Price                float64 `json:"price"`     // Required
	ListClientOrderID    string  `json:"listClientOrderId"`
	LimitClientOrderID   string  `json:"limitClientOrderId"`
	LimitIcebergQty      float64 `json:"limitIcebergQty"`
	StopClientOrderID    string  `json:"stopClientOrderId"`
	StopLimitPrice       float64 `json:"stopLimitPrice"`
	StopIcebergQty       float64 `json:"stopIcebergQty"`
	StopLimitTimeInForce string  `json:"stopLimitTimeInForce"`
	NewOrderRespType     string  `json:"newOrderRespType"`
	RecvWindow           uint64  `json:"recvWindow"`
}

// GetOCOPrderRequestParams
type GetOCOPrderRequestParams struct {
	OrderListID       uint64
	OrigClientOrderID string
}

type OrderShortResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       uint64 `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
}

// OCONewOrderResponse this model is to be used to fetch the respons of create new OCO order response
type OCOOrderResponse struct {
	OrderListId       int64                 `json:"orderListId"`
	ContingencyType   string                `json:"contingencyType"`
	ListStatusType    string                `json:"listStatusType"`
	ListOrderStatus   string                `json:"listOrderStatus"`
	ListClientOrderId string                `json:"listClientOrderId"`
	TransactionTime   time.Time             `json:"transactionTime"`
	Symbol            string                `json:"symbol"`
	Orders            []*OrderShortResponse `json:"orders"`
}

// CreateNewOrderResponse
type OCOFullOrderResponse struct {
	*OCOOrderResponse
	OrderReports []*OCOOrderReportItem `json:"orderReports"`
}

// OCOOrdersRequestParams
type OCOOrdersRequestParams struct {
	FromID     uint64
	StartTime  time.Time
	EndTime    time.Time
	Limit      uint
	RecvWindow uint
}

// OCOOrdersDeleteRequestParams
// holds the params to delete a new order
type OCOOrdersDeleteRequestParams struct {
	Symbol            string
	OrderListID       uint64
	ListClientOrderID string
	NewClientOrderID  string
	RecvWindow        uint
}

// OTC endpoinsts

// CoinPairInfo
type CoinPairInfo struct {
	FromCoin          string  `json:"fromCoin"`
	ToCoin            string  `json:"toCoin"`
	FromCoinMinAmount float64 `json:"fromCoinMinAmount,string"`
	FromCoinMaxAmount float64 `json:"fromCoinMaxAmount,string"`
	ToCoinMinAmount   float64 `json:"toCoinMinAmount,string"`
	ToCoinMaxAmount   float64 `json:"toCoinMaxAmount,string"`
}

// RequestQuoteParams
type RequestQuoteParams struct {
	FromCoin      string  `json:"fronCoin"`
	ToCoin        string  `json:"toCoin"`
	RequestCoin   string  `json:"requestCoin"`
	RequestAmount float64 `json:"requestAmount"`
}

// RequestQuote
type RequestQuote struct {
	QuoteId        string  `json:"quoteId"`
	Symbol         string  `json:"symbol"`
	Ratio          float64 `json:"ratio"`
	InverseRatio   float64 `json:"inverseRatio"`
	ValidTimestamp float64 `json:"validTimestamp"`
	ToAmount       float64 `json:"toAmount"`
	FromAmount     uint64  `json:"fromAmount"`
}

// OTCTradeOrderResponse
type OTCTradeOrderResponse struct {
	OrderID     uint64    `json:"orderId,string"`
	CreateTime  time.Time `json:"createTime"`
	OrderStatus string    `json:"orderStatus"`
}

// OTCTradeOrder
type OTCTradeOrder struct {
	QuoteID      string    `json:"quoteId"`
	OrderID      uint64    `json:"orderId,string"`
	OrderStatus  string    `json:"orderStatus"`
	FromCoin     string    `json:"fromCoin"`
	FromAmount   float64   `json:"fromAmount"`
	ToCoin       string    `json:"toCoin"`
	ToAmount     float64   `json:"toAmount"`
	Ratio        float64   `json:"ratio"`
	InverseRatio float64   `json:"inverseRatio"`
	CreateTime   time.Time `json:"createTime"`
}

// OTCTradeOrderParams request param for Over-the-Counter trade order params.
type OTCTradeOrderRequestParams struct {
	OrderId   string
	FromCoin  string
	ToCoin    string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int8
}

// Wallet Endpoints
//

// AssetWalletDetail represents the wallet asset information.
type AssetWalletDetail struct {
	Coin              string `json:"coin"`
	DepositAllEnable  bool   `json:"depositAllEnable"`
	WithdrawAllEnable bool   `json:"withdrawAllEnable"`
	Name              string `json:"name"`
	Free              string `json:"free"`
	Locked            string `json:"locked"`
	Freeze            string `json:"freeze"`
	Withdrawing       string `json:"withdrawing"`
	Ipoing            string `json:"ipoing"`
	Ipoable           string `json:"ipoable"`
	Storage           string `json:"storage"`
	IsLegalMoney      bool   `json:"isLegalMoney"`
	Trading           bool   `json:"trading"`
	NetworkList       []struct {
		Network                 string  `json:"network"`
		Coin                    string  `json:"coin"`
		WithdrawIntegerMultiple string  `json:"withdrawIntegerMultiple"`
		IsDefault               bool    `json:"isDefault"`
		DepositEnable           bool    `json:"depositEnable"`
		WithdrawEnable          bool    `json:"withdrawEnable"`
		DepositDesc             string  `json:"depositDesc"`
		WithdrawDesc            string  `json:"withdrawDesc"`
		Name                    string  `json:"name"`
		ResetAddressStatus      bool    `json:"resetAddressStatus"`
		WithdrawFee             float64 `json:"withdrawFee,string"`
		WithdrawMin             float64 `json:"withdrawMin,string"`
		WithdrawMax             float64 `json:"withdrawMax,string"`
		AddressRegex            string  `json:"addressRegex,omitempty"`
		MemoRegex               string  `json:"memoRegex,omitempty"`
		MinConfirm              int     `json:"minConfirm,omitempty"`
		UnLockConfirm           int     `json:"unLockConfirm,omitempty"`
	} `json:"networkList"`
}

// AssetWalletList list of asset wallet details
type AssetWalletList []AssetWalletDetail

// WithdrawalRequestParam represents the params for the
// input parameters of Withdraw Crypto
type WithdrawalRequestParam struct {
	Coin            string  `json:"coin"`
	Network         string  `json:"network"`
	WithdrawOrderId string  `json:"withdrawOrderId"` // Client ID for withdraw
	Address         string  `json:"address"`
	AddressTag      string  `json:"addressTag"`
	Amount          float64 `json:"amount"`
	RecvWindow      uint64  `json:"recvWindow"`
}

// WithdrawalResponse holds the transaction id for a withdrawal action.
type WithdrawalResponse struct {
	ID string `json:"id"`
}

// WithdrawStatusResponse defines a withdrawal status response
type WithdrawStatusResponse struct {
	ID             string  `json:"id"`
	Amount         float64 `json:"amount,string"`
	TransactionFee float64 `json:"transactionFee,string"`
	Coin           string  `json:"coin"`
	Status         int     `json:"status"`
	Address        string  `json:"address"`
	ApplyTime      string  `json:"applyTime"`
	Network        string  `json:"network"`
	TransferType   int     `json:"transferType"`
}

// FiatAssetRecord asset information for fiat.
type FiatAssetRecord struct {
	OrderID        string `json:"orderId"`
	PaymentAccount string `json:"paymentAccount"`
	PaymentChannel string `json:"paymentChannel"`
	PaymentMethod  string `json:"paymentMethod"`
	OrderStatus    string `json:"orderStatus"`
	Amount         string `json:"amount"`
	TransactionFee string `json:"transactionFee"`
	PlatformFee    string `json:"platformFee"`
}

// FiatWithdrawalHistory holds list of availabel fiat asset records.
type FiatAssetsHistory struct {
	AssetLogRecordList []FiatAssetRecord `json:"assetLogRecordList"`
}

// WithdrawFiatRequestParams repsents the fiat withdrawal request params.
type WithdrawFiatRequestParams struct {
	PaymentChannel string
	PaymentMethod  string
	PaymentAccount string
	FiatCurrency   string
	Amount         float64
	RecvWindow     uint64
}

// FiatWithdrawalRequestParams to fetch your fiat (USD) withdrawal history.
type FiatWithdrawalRequestParams struct {
	FiatCurrency   string
	OrderId        string
	Offset         int
	PaymentChannel string
	PaymentMethod  string
	StartTime      time.Time
	EndTime        time.Time
}

// DepositAddress stores the deposit address info
type DepositAddress struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag"`
	URL     string `json:"url"`
}

// DepositHistory stores deposit history info.
type DepositHistory struct {
	Amount       string `json:"amount"`
	Coin         string `json:"coin"`
	Network      string `json:"network"`
	Status       int    `json:"status"`
	Address      string `json:"address"`
	AddressTag   string `json:"addressTag"`
	TxID         string `json:"txId"`
	InsertTime   int64  `json:"insertTime"`
	TransferType int    `json:"transferType"`
	ConfirmTimes string `json:"confirmTimes"`
}

// UserAccountStream represents the response for getting the listen key for the websocket
type UserAccountStream struct {
	ListenKey string `json:"listenKey"`
}

// WebsocketPayload defines the payload through the websocket connection
type WebsocketPayload struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int64         `json:"id"`
}

// orderbookManager defines a way of managing and maintaining synchronisation
// across connections and assets.
type orderbookManager struct {
	state map[currency.Code]map[currency.Code]map[asset.Item]*update
	sync.Mutex

	jobs chan job
}

// job defines a synchonisation job that tells a go routine to fetch an
// orderbook via the REST protocol
type job struct {
	Pair currency.Pair
}

// update holds websocker depth stream response data and update informations.
type update struct {
	buffer            chan *WebsocketDepthStream
	fetchingBook      bool
	initialSync       bool
	needsFetchingBook bool
	lastUpdateID      int64
}

// WebsocketDepthStream is the difference for the update depth stream
type WebsocketDepthStream struct {
	Event         string           `json:"e"`
	Timestamp     time.Time        `json:"E"`
	Pair          string           `json:"s"`
	FirstUpdateID int64            `json:"U"`
	LastUpdateID  int64            `json:"u"`
	UpdateBids    [][2]interface{} `json:"b"`
	UpdateAsks    [][2]interface{} `json:"a"`
}

// WebsocketDepthDiffStream websocket response of depth diff stream
type WebsocketDepthDiffStream struct {
	LastUpdateID int         `json:"lastUpdateId"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
}

// wsAccountInfo websocekt response of account information.
type wsAccountInfo struct {
	Stream string            `json:"stream"`
	Data   WsAccountInfoData `json:"data"`
}

// WsAccountInfoData defines websocket account info data
type WsAccountInfoData struct {
	CanDeposit       bool      `json:"D"`
	CanTrade         bool      `json:"T"`
	CanWithdraw      bool      `json:"W"`
	EventTime        time.Time `json:"E"`
	LastUpdated      time.Time `json:"u"`
	BuyerCommission  float64   `json:"b"`
	MakerCommission  float64   `json:"m"`
	SellerCommission float64   `json:"s"`
	TakerCommission  float64   `json:"t"`
	EventType        string    `json:"e"`
	Currencies       []struct {
		Asset     string  `json:"a"`
		Available float64 `json:"f,string"`
		Locked    float64 `json:"l,string"`
	} `json:"B"`
}

// wsAccountPosition websocke response of account position.
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
	EventTime   time.Time `json:"E"`
	LastUpdated time.Time `json:"u"`
	EventType   string    `json:"e"`
}

// wsBalanceUpdate respresents the websocket response of update balance.
type wsBalanceUpdate struct {
	Stream string              `json:"stream"`
	Data   WsBalanceUpdateData `json:"data"`
}

// WsBalanceUpdateData defines websocket account balance data.
type WsBalanceUpdateData struct {
	EventTime    time.Time `json:"E"`
	ClearTime    time.Time `json:"T"`
	BalanceDelta float64   `json:"d,string"`
	Asset        string    `json:"a"`
	EventType    string    `json:"e"`
}

type wsOrderUpdate struct {
	Stream string            `json:"stream"`
	Data   WsOrderUpdateData `json:"data"`
}

// WsOrderUpdateData defines websocket account order update data
type WsOrderUpdateData struct {
	EventType                         string    `json:"e"`
	EventTime                         time.Time `json:"E"`
	Symbol                            string    `json:"s"`
	ClientOrderID                     string    `json:"c"`
	Side                              string    `json:"S"`
	OrderType                         string    `json:"o"`
	TimeInForce                       string    `json:"f"`
	Quantity                          float64   `json:"q,string"`
	Price                             float64   `json:"p,string"`
	StopPrice                         float64   `json:"P,string"`
	IcebergQuantity                   float64   `json:"F,string"`
	OrderListID                       int64     `json:"g"`
	CancelledClientOrderID            string    `json:"C"`
	CurrentExecutionType              string    `json:"x"`
	OrderStatus                       string    `json:"X"`
	RejectionReason                   string    `json:"r"`
	OrderID                           int64     `json:"i"`
	LastExecutedQuantity              float64   `json:"l,string"`
	CumulativeFilledQuantity          float64   `json:"z,string"`
	LastExecutedPrice                 float64   `json:"L,string"`
	Commission                        float64   `json:"n,string"`
	CommissionAsset                   string    `json:"N"`
	TransactionTime                   time.Time `json:"T"`
	TradeID                           int64     `json:"t"`
	Ignored                           int64     `json:"I"` // Must be ignored explicitly, otherwise it overwrites 'i'.
	IsOnOrderBook                     bool      `json:"w"`
	IsMaker                           bool      `json:"m"`
	Ignored2                          bool      `json:"M"` // See the comment for "I".
	OrderCreationTime                 time.Time `json:"O"`
	CumulativeQuoteTransactedQuantity float64   `json:"Z,string"`
	LastQuoteAssetTransactedQuantity  float64   `json:"Y,string"`
	QuoteOrderQuantity                float64   `json:"Q,string"`
}

// wsListStatus holder for websocker account listing status response.
type wsListStatus struct {
	Stream string           `json:"stream"`
	Data   WsListStatusData `json:"data"`
}

// WsListStatusData defines websocket account listing status data
type WsListStatusData struct {
	ListClientOrderID string    `json:"C"`
	EventTime         time.Time `json:"E"`
	ListOrderStatus   string    `json:"L"`
	Orders            []struct {
		ClientOrderID string `json:"c"`
		OrderID       int64  `json:"i"`
		Symbol        string `json:"s"`
	} `json:"O"`
	TransactionTime time.Time `json:"T"`
	ContingencyType string    `json:"c"`
	EventType       string    `json:"e"`
	OrderListID     int64     `json:"g"`
	ListStatusType  string    `json:"l"`
	RejectionReason string    `json:"r"`
	Symbol          string    `json:"s"`
}

// TradeStream holds the trade stream data
type TradeStream struct {
	EventType      string    `json:"e"`
	EventTime      time.Time `json:"E"`
	Symbol         string    `json:"s"`
	TradeID        int64     `json:"t"`
	Price          string    `json:"p"`
	Quantity       string    `json:"q"`
	BuyerOrderID   int64     `json:"b"`
	SellerOrderID  int64     `json:"a"`
	TimeStamp      time.Time `json:"T"`
	Maker          bool      `json:"m"`
	BestMatchPrice bool      `json:"M"`
}

// KlineStream holds the kline stream data
type KlineStream struct {
	EventType string          `json:"e"`
	EventTime time.Time       `json:"E"`
	Symbol    string          `json:"s"`
	Kline     KlineStreamData `json:"k"`
}

// KlineStreamData defines kline streaming data
type KlineStreamData struct {
	StartTime                time.Time `json:"t"`
	CloseTime                time.Time `json:"T"`
	Symbol                   string    `json:"s"`
	Interval                 string    `json:"i"`
	FirstTradeID             int64     `json:"f"`
	LastTradeID              int64     `json:"L"`
	OpenPrice                float64   `json:"o,string"`
	ClosePrice               float64   `json:"c,string"`
	HighPrice                float64   `json:"h,string"`
	LowPrice                 float64   `json:"l,string"`
	Volume                   float64   `json:"v,string"`
	NumberOfTrades           int64     `json:"n"`
	KlineClosed              bool      `json:"x"`
	Quote                    float64   `json:"q,string"`
	TakerBuyBaseAssetVolume  float64   `json:"V,string"`
	TakerBuyQuoteAssetVolume float64   `json:"Q,string"`
}

// TickerStream holds the ticker stream data
type TickerStream struct {
	EventType              string    `json:"e"`
	EventTime              time.Time `json:"E"`
	Symbol                 string    `json:"s"`
	PriceChange            float64   `json:"p,string"`
	PriceChangePercent     float64   `json:"P,string"`
	WeightedAvgPrice       float64   `json:"w,string"`
	ClosePrice             float64   `json:"x,string"`
	LastPrice              float64   `json:"c,string"`
	LastPriceQuantity      float64   `json:"Q,string"`
	BestBidPrice           float64   `json:"b,string"`
	BestBidQuantity        float64   `json:"B,string"`
	BestAskPrice           float64   `json:"a,string"`
	BestAskQuantity        float64   `json:"A,string"`
	OpenPrice              float64   `json:"o,string"`
	HighPrice              float64   `json:"h,string"`
	LowPrice               float64   `json:"l,string"`
	TotalTradedVolume      float64   `json:"v,string"`
	TotalTradedQuoteVolume float64   `json:"q,string"`
	OpenTime               time.Time `json:"O"`
	CloseTime              time.Time `json:"C"`
	FirstTradeID           int64     `json:"F"`
	LastTradeID            int64     `json:"L"`
	NumberOfTrades         int64     `json:"n"`
}

// Additional Data MOdels when adding the Market Data Streams
type OrderBookTickerStream struct {
	LastUpdateID int    `json:"u"`
	S            string `json:"s"`
	Symbol       currency.Pair
	BestBidPrice float64 `json:"b,string"`
	BestBidQty   float64 `json:"B,string"`
	BestAskPrice float64 `json:"a,string"`
	BestAskQty   float64 `json:"A,string"`
}

// Websocket stream aggregate trade
type WebsocketAggregateTradeStream struct {
	EventType        string    `json:"e"`
	EventTime        time.Time `json:"E"`
	Symbol           string    `json:"s"`
	AggregateTradeID int       `json:"a"`
	Price            float64   `json:"p,string"`
	Quantity         float64   `json:"q,string"`
	FirstTradeID     int       `json:"f"`
	LastTradeID      int       `json:"l"`
	TradeTime        time.Time `json:"T"`
	IsMaker          bool      `json:"m"`
}
