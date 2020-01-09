package order

import (
	"errors"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
)

// vars related to orders
var (
	ErrSubmissionIsNil            = errors.New("order submission is nil")
	ErrPairIsEmpty                = errors.New("order pair is empty")
	ErrSideIsInvalid              = errors.New("order side is invalid")
	ErrTypeIsInvalid              = errors.New("order type is invalid")
	ErrAmountIsInvalid            = errors.New("order amount is invalid")
	ErrPriceMustBeSetIfLimitOrder = errors.New("order price must be set if limit order type is desired")
)

// Order struct holds order values
type Order struct {
	OrderID  int
	Exchange string
	Type     int
	Amount   float64
	Price    float64
}

// Properties is the shared holder of all order details
// Not all properties are required by all order structs
// e.g. Modify order may only need an ID and amount for one exchange,
// but may contain all properties on another.
// Depending on the data returned by exchanges, it is best that all order
// types contain the same properties so the orderStore can know all at all times
type Properties struct {
	ImmediateOrCancel bool
	HiddenOrder       bool
	FillOrKill        bool
	PostOnly          bool
	Price             float64
	Amount            float64
	LimitPriceUpper   float64
	LimitPriceLower   float64
	TriggerPrice      float64
	TargetAmount      float64
	ExecutedAmount    float64
	RemainingAmount   float64
	Fee               float64
	Exchange          string
	AccountID         string
	ID                string
	ClientID          string
	OrderID           string
	OrderType         Type
	OrderSide         Side
	OrderStatus       Status
	OrderDate         time.Time
	CurrencyPair      currency.Pair
	Trades            []TradeHistory
}

// Submit contains the order submission data
type Submit struct {
	Properties
}

// SubmitResponse is what is returned after submitting an order to an exchange
type SubmitResponse struct {
	IsOrderPlaced bool
	FullyMatched  bool
	OrderID       string
}

// Modify is an order modifyer
type Modify struct {
	Properties
}

// ModifyResponse is an order modifying return type
type ModifyResponse struct {
	OrderID string
}

// Detail holds order detail data
type Detail struct {
	Properties
}

// Cancel type required when requesting to cancel an order
type Cancel struct {
	Properties
}

// CancelAllResponse returns the status from attempting to cancel all orders on
// an exchagne
type CancelAllResponse struct {
	Status map[string]string
}

// TradeHistory holds exchange history data
type TradeHistory struct {
	Timestamp   time.Time
	TID         string
	Price       float64
	Amount      float64
	Exchange    string
	Type        Type
	Side        Side
	Fee         float64
	Description string
}

// GetOrdersRequest used for GetOrderHistory and GetOpenOrders wrapper functions
type GetOrdersRequest struct {
	OrderType  Type
	OrderSide  Side
	StartTicks time.Time
	EndTicks   time.Time
	// Currencies Empty array = all currencies. Some endpoints only support
	// singular currency enquiries
	Currencies []currency.Pair
}

// Status defines order status types
type Status string

// All order status types
const (
	AnyStatus          Status = "ANY"
	New                Status = "NEW"
	Active             Status = "ACTIVE"
	PartiallyCancelled Status = "PARTIALLY_CANCELLED"
	PartiallyFilled    Status = "PARTIALLY_FILLED"
	Filled             Status = "FILLED"
	Cancelled          Status = "CANCELLED"
	PendingCancel      Status = "PENDING_CANCEL"
	Rejected           Status = "REJECTED"
	Expired            Status = "EXPIRED"
	Hidden             Status = "HIDDEN"
	UnknownStatus      Status = "UNKNOWN"
)

// Type enforces a standard for order types across the code base
type Type string

// Defined package order types
const (
	AnyType           Type = "ANY"
	Limit             Type = "LIMIT"
	Market            Type = "MARKET"
	ImmediateOrCancel Type = "IMMEDIATE_OR_CANCEL"
	Stop              Type = "STOP"
	TrailingStop      Type = "TRAILINGSTOP"
	Unknown           Type = "UNKNOWN"
)

// Side enforces a standard for order sides across the code base
type Side string

// Order side types
const (
	AnySide     Side = "ANY"
	Buy         Side = "BUY"
	Sell        Side = "SELL"
	Bid         Side = "BID"
	Ask         Side = "ASK"
	SideUnknown Side = "SIDEUNKNOWN"
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

// Orders variable holds an array of pointers to order structs
var Orders []*Order
