package order

import (
	"errors"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// var error definitions
var (
	ErrSubmissionIsNil            = errors.New("order submission is nil")
	ErrCancelOrderIsNil           = errors.New("cancel order is nil")
	ErrGetOrdersRequestIsNil      = errors.New("get order request is nil")
	ErrModifyOrderIsNil           = errors.New("modify order request is nil")
	ErrPairIsEmpty                = errors.New("order pair is empty")
	ErrAssetNotSet                = errors.New("order asset type is not set")
	ErrSideIsInvalid              = errors.New("order side is invalid")
	ErrTypeIsInvalid              = errors.New("order type is invalid")
	ErrAmountIsInvalid            = errors.New("order amount is equal or less than zero")
	ErrPriceMustBeSetIfLimitOrder = errors.New("order price must be set if limit order type is desired")
	ErrOrderIDNotSet              = errors.New("order id or client order id is not set")
	errCannotLiquidate            = errors.New("cannot liquidate position")
)

// Submit contains all properties of an order that may be required
// for an order to be created on an exchange
// Each exchange has their own requirements, so not all fields
// are required to be populated
type Submit struct {
	ImmediateOrCancel bool
	HiddenOrder       bool
	FillOrKill        bool
	PostOnly          bool
	ReduceOnly        bool
	Leverage          float64
	Price             float64

	// Amount in base terms
	Amount float64
	// QuoteAmount is the max amount in quote currency when purchasing base.
	// This is only used in Market orders.
	QuoteAmount float64

	StopPrice       float64
	LimitPriceUpper float64
	LimitPriceLower float64
	TriggerPrice    float64
	ExecutedAmount  float64
	RemainingAmount float64
	Fee             float64
	Exchange        string
	InternalOrderID string
	ID              string
	AccountID       string
	ClientID        string
	ClientOrderID   string
	WalletAddress   string
	Offset          string
	Type            Type
	Side            Side
	Status          Status
	AssetType       asset.Item
	Date            time.Time
	LastUpdated     time.Time
	Pair            currency.Pair
	Trades          []TradeHistory
}

// SubmitResponse is what is returned after submitting an order to an exchange
type SubmitResponse struct {
	IsOrderPlaced bool
	FullyMatched  bool
	OrderID       string
	Rate          float64
	Fee           float64
	Cost          float64
	Trades        []TradeHistory
}

// Modify contains all properties of an order
// that may be updated after it has been created
// Each exchange has their own requirements, so not all fields
// are required to be populated
type Modify struct {
	ImmediateOrCancel bool
	HiddenOrder       bool
	FillOrKill        bool
	PostOnly          bool
	Leverage          float64
	Price             float64
	Amount            float64
	LimitPriceUpper   float64
	LimitPriceLower   float64
	TriggerPrice      float64
	QuoteAmount       float64
	ExecutedAmount    float64
	RemainingAmount   float64
	Fee               float64
	Exchange          string
	InternalOrderID   string
	ID                string
	ClientOrderID     string
	AccountID         string
	ClientID          string
	WalletAddress     string
	Type              Type
	Side              Side
	Status            Status
	AssetType         asset.Item
	Date              time.Time
	LastUpdated       time.Time
	Pair              currency.Pair
	Trades            []TradeHistory
}

// ModifyResponse is an order modifying return type
type ModifyResponse struct {
	OrderID string
}

// Detail contains all properties of an order
// Each exchange has their own requirements, so not all fields
// are required to be populated
type Detail struct {
	ImmediateOrCancel    bool
	HiddenOrder          bool
	FillOrKill           bool
	PostOnly             bool
	Leverage             float64
	Price                float64
	Amount               float64
	LimitPriceUpper      float64
	LimitPriceLower      float64
	TriggerPrice         float64
	AverageExecutedPrice float64
	QuoteAmount          float64
	ExecutedAmount       float64
	RemainingAmount      float64
	Cost                 float64
	CostAsset            currency.Code
	Fee                  float64
	FeeAsset             currency.Code
	Exchange             string
	InternalOrderID      string
	ID                   string
	ClientOrderID        string
	AccountID            string
	ClientID             string
	WalletAddress        string
	Type                 Type
	Side                 Side
	Status               Status
	AssetType            asset.Item
	Date                 time.Time
	CloseTime            time.Time
	LastUpdated          time.Time
	Pair                 currency.Pair
	Trades               []TradeHistory
}

// Filter contains all properties an order can be filtered for
// empty strings indicate to ignore the property otherwise all need to match
type Filter struct {
	Exchange        string
	InternalOrderID string
	ID              string
	ClientOrderID   string
	AccountID       string
	ClientID        string
	WalletAddress   string
	Type            Type
	Side            Side
	Status          Status
	AssetType       asset.Item
	Pair            currency.Pair
}

// Cancel contains all properties that may be required
// to cancel an order on an exchange
// Each exchange has their own requirements, so not all fields
// are required to be populated
type Cancel struct {
	Price         float64
	Amount        float64
	Exchange      string
	ID            string
	ClientOrderID string
	AccountID     string
	ClientID      string
	WalletAddress string
	Type          Type
	Side          Side
	Status        Status
	AssetType     asset.Item
	Date          time.Time
	Pair          currency.Pair
	Symbol        string
	Trades        []TradeHistory
}

// CancelAllResponse returns the status from attempting to
// cancel all orders on an exchange
type CancelAllResponse struct {
	Status map[string]string
	Count  int64
}

// CancelBatchResponse returns the status of orders
// that have been requested for cancellation
type CancelBatchResponse struct {
	Status map[string]string
}

// TradeHistory holds exchange history data
type TradeHistory struct {
	Price       float64
	Amount      float64
	Fee         float64
	Exchange    string
	TID         string
	Description string
	Type        Type
	Side        Side
	Timestamp   time.Time
	IsMaker     bool
	FeeAsset    string
	Total       float64
}

// GetOrdersRequest used for GetOrderHistory and GetOpenOrders wrapper functions
type GetOrdersRequest struct {
	Type      Type
	Side      Side
	StartTime time.Time
	EndTime   time.Time
	OrderID   string
	// Currencies Empty array = all currencies. Some endpoints only support
	// singular currency enquiries
	Pairs     currency.Pairs
	AssetType asset.Item
}

// Status defines order status types
type Status uint32

// All order status types
const (
	UnknownStatus Status = 0
	AnyStatus     Status = 1 << iota
	New
	Active
	PartiallyCancelled
	PartiallyFilled
	Filled
	Cancelled
	PendingCancel
	InsufficientBalance
	MarketUnavailable
	Rejected
	Expired
	Hidden
	Open
	AutoDeleverage
	Closed
	Pending
	Cancelling
	Liquidated          Status = "LIQUIDATED"
)

// Type enforces a standard for order types across the code base
type Type uint16

// Defined package order types
const (
	UnknownType Type = 0
	Limit       Type = 1 << iota
	Market
	PostOnly
	ImmediateOrCancel
	Stop
	StopLimit
	StopMarket
	TakeProfit
	TakeProfitMarket
	TrailingStop
	FillOrKill
	IOS
	AnyType
	Liquidation
	Trigger
)

// Side enforces a standard for order sides across the code base
type Side uint16

// Order side types
const (
	UnknownSide Side = 0
	Buy         Side = 1 << iota
	Sell
	Bid
	Ask
	AnySide
	Long
	Short
	// Backtester signal types
	DoNothing
	TransferredFunds
	CouldNotBuy
	CouldNotSell
	MissingData
	SideNA      Side = "N/A"
)

// ByPrice used for sorting orders by price
type ByPrice []Detail

// ByOrderType used for sorting orders by order type
type ByOrderType []Detail

// ByCurrency used for sorting orders by order currency
type ByCurrency []Detail

// ByDate used for sorting orders by order date
type ByDate []Detail

// ByOrderSide used for sorting orders by order side (buy sell)
type ByOrderSide []Detail

// ClassificationError returned when an order status
// side or type cannot be recognised
type ClassificationError struct {
	Exchange string
	OrderID  string
	Err      error
}
