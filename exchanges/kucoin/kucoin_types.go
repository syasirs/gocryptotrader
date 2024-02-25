package kucoin

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/types"
)

var (
	validPeriods = []string{
		"1min", "3min", "5min", "15min", "30min", "1hour", "2hour", "4hour", "6hour", "8hour", "12hour", "1day", "1week",
	}

	errInvalidResponseReceiver   = errors.New("invalid response receiver")
	errInvalidPrice              = errors.New("invalid price")
	errInvalidStopPriceType      = errors.New("stopPriceType is required")
	errInvalidSize               = errors.New("invalid size")
	errMalformedData             = errors.New("malformed data")
	errNoDepositAddress          = errors.New("no deposit address found")
	errMultipleDepositAddress    = errors.New("multiple deposit addresses")
	errInvalidResultInterface    = errors.New("result interface has to be pointer")
	errInvalidSubAccountName     = errors.New("invalid sub-account name")
	errAPIKeyRequired            = errors.New("account API key is required")
	errInvalidPassPhraseInstance = errors.New("invalid passphrase string")
	errNoValidResponseFromServer = errors.New("no valid response from server")
	errMissingOrderbookSequence  = errors.New("missing orderbook sequence")
	errSizeOrFundIsRequired      = errors.New("at least one required among size and funds")
	errInvalidLeverage           = errors.New("invalid leverage value")
	errInvalidClientOrderID      = errors.New("no client order ID supplied, this endpoint requires a UUID or similar string")
	errAccountTypeMissing        = errors.New("account type is required")
	errTransferTypeMissing       = errors.New("transfer type is required")
	errTradeTypeMissing          = errors.New("trade type is missing")
	errTimeInForceRequired       = errors.New("time in force is required")
	errInvalidMsgType            = errors.New("message type field not valid")
	errSubscriptionPairRequired  = errors.New("pair required for manual subscriptions")

	subAccountRegExp           = regexp.MustCompile("^[a-zA-Z0-9]{7-32}$")
	subAccountPassphraseRegExp = regexp.MustCompile("^[a-zA-Z0-9]{7-24}$")
	errAccountIDMissing        = errors.New("account ID is required")
)

// UnmarshalTo acts as interface to exchange API response
type UnmarshalTo interface {
	GetError() error
}

// Error defines all error information for each request
type Error struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// GetError checks and returns an error if it is supplied.
func (e Error) GetError() error {
	code, err := strconv.ParseInt(e.Code, 10, 64)
	if err != nil {
		return err
	}
	switch code {
	case 200:
		return nil
	case 200000:
		if e.Msg == "" {
			return nil
		}
	}
	return fmt.Errorf("code: %s message: %s", e.Code, e.Msg)
}

// SymbolInfo stores symbol information
type SymbolInfo struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	BaseCurrency    string  `json:"baseCurrency"`
	QuoteCurrency   string  `json:"quoteCurrency"`
	FeeCurrency     string  `json:"feeCurrency"`
	Market          string  `json:"market"`
	BaseMinSize     float64 `json:"baseMinSize,string"`
	QuoteMinSize    float64 `json:"quoteMinSize,string"`
	BaseMaxSize     float64 `json:"baseMaxSize,string"`
	QuoteMaxSize    float64 `json:"quoteMaxSize,string"`
	BaseIncrement   float64 `json:"baseIncrement,string"`
	QuoteIncrement  float64 `json:"quoteIncrement,string"`
	PriceIncrement  float64 `json:"priceIncrement,string"`
	PriceLimitRate  float64 `json:"priceLimitRate,string"`
	MinFunds        float64 `json:"minFunds,string"`
	IsMarginEnabled bool    `json:"isMarginEnabled"`
	EnableTrading   bool    `json:"enableTrading"`
}

// Ticker stores ticker data
type Ticker struct {
	Sequence    string               `json:"sequence"`
	BestAsk     float64              `json:"bestAsk,string"`
	Size        float64              `json:"size,string"`
	Price       float64              `json:"price,string"`
	BestBidSize float64              `json:"bestBidSize,string"`
	BestBid     float64              `json:"bestBid,string"`
	BestAskSize float64              `json:"bestAskSize,string"`
	Time        convert.ExchangeTime `json:"time"`
}

type tickerInfoBase struct {
	Symbol           string  `json:"symbol"`
	Buy              float64 `json:"buy,string"`
	Sell             float64 `json:"sell,string"`
	ChangeRate       float64 `json:"changeRate,string"`
	ChangePrice      float64 `json:"changePrice,string"`
	High             float64 `json:"high,string"`
	Low              float64 `json:"low,string"`
	Volume           float64 `json:"vol,string"`
	VolumeValue      float64 `json:"volValue,string"`
	Last             float64 `json:"last,string"`
	AveragePrice     float64 `json:"averagePrice,string"`
	TakerFeeRate     float64 `json:"takerFeeRate,string"`
	MakerFeeRate     float64 `json:"makerFeeRate,string"`
	TakerCoefficient float64 `json:"takerCoefficient,string"`
	MakerCoefficient float64 `json:"makerCoefficient,string"`
	BestBidSize      float64 `json:"bestBidSize,string"`
	BestAskSize      float64 `json:"bestAskSize,string"`
}

// TickerInfo stores ticker information
type TickerInfo struct {
	tickerInfoBase
	SymbolName string `json:"symbolName"`
}

// Stats24hrs stores 24 hrs statistics
type Stats24hrs struct {
	tickerInfoBase
	Time convert.ExchangeTime `json:"time"`
}

// Orderbook stores the orderbook data
type Orderbook struct {
	Sequence int64
	Bids     []orderbook.Item
	Asks     []orderbook.Item
	Time     time.Time
}

type orderbookResponse struct {
	Asks     [][2]types.Number    `json:"asks"`
	Bids     [][2]types.Number    `json:"bids"`
	Time     convert.ExchangeTime `json:"time"`
	Sequence string               `json:"sequence"`
}

// Trade stores trade data
type Trade struct {
	Sequence string               `json:"sequence"`
	Price    float64              `json:"price,string"`
	Size     float64              `json:"size,string"`
	Side     string               `json:"side"`
	Time     convert.ExchangeTime `json:"time"`
}

// Kline stores kline data
type Kline struct {
	StartTime time.Time
	Open      float64
	Close     float64
	High      float64
	Low       float64
	Volume    float64 // Transaction volume
	Amount    float64 // Transaction amount
}

type currencyBase struct {
	Currency        string `json:"currency"` // a unique currency code that will never change
	Name            string `json:"name"`     // will change after renaming
	FullName        string `json:"fullName"`
	Precision       int64  `json:"precision"`
	Confirms        int64  `json:"confirms"`
	ContractAddress string `json:"contractAddress"`
	IsMarginEnabled bool   `json:"isMarginEnabled"`
	IsDebitEnabled  bool   `json:"isDebitEnabled"`
}

// Currency stores currency data
type Currency struct {
	currencyBase
	WithdrawalMinSize float64 `json:"withdrawalMinSize,string"`
	WithdrawalMinFee  float64 `json:"withdrawalMinFee,string"`
	IsWithdrawEnabled bool    `json:"isWithdrawEnabled"`
	IsDepositEnabled  bool    `json:"isDepositEnabled"`
}

// Chain stores blockchain data
type Chain struct {
	Name              string  `json:"chainName"`
	Confirms          int64   `json:"confirms"`
	ContractAddress   string  `json:"contractAddress"`
	WithdrawalMinSize float64 `json:"withdrawalMinSize,string"`
	WithdrawalMinFee  float64 `json:"withdrawalMinFee,string"`
	IsWithdrawEnabled bool    `json:"isWithdrawEnabled"`
	IsDepositEnabled  bool    `json:"isDepositEnabled"`
}

// CurrencyDetail stores currency details
type CurrencyDetail struct {
	currencyBase
	Chains []Chain `json:"chains"`
}

// LeveragedTokenInfo represents a leveraged token information
type LeveragedTokenInfo struct {
	Currency              string       `json:"currency"`
	NetAsset              float64      `json:"netAsset"`
	TargetLeverage        types.Number `json:"targetLeverage"`
	ActualLeverage        types.Number `json:"actualLeverage"`
	AssetsUnderManagement string       `json:"assetsUnderManagement"`
	Basket                string       `json:"basket"`
}

// MarkPrice stores mark price data
type MarkPrice struct {
	Symbol      string               `json:"symbol"`
	Granularity int64                `json:"granularity"`
	TimePoint   convert.ExchangeTime `json:"timePoint"`
	Value       float64              `json:"value"`
}

// MarginConfiguration stores margin configuration
type MarginConfiguration struct {
	CurrencyList     []string `json:"currencyList"`
	WarningDebtRatio float64  `json:"warningDebtRatio,string"`
	LiqDebtRatio     float64  `json:"liqDebtRatio,string"`
	MaxLeverage      float64  `json:"maxLeverage"`
}

// MarginAccount stores margin account data
type MarginAccount struct {
	AvailableBalance float64 `json:"availableBalance,string"`
	Currency         string  `json:"currency"`
	HoldBalance      float64 `json:"holdBalance,string"`
	Liability        float64 `json:"liability,string"`
	MaxBorrowSize    float64 `json:"maxBorrowSize,string"`
	TotalBalance     float64 `json:"totalBalance,string"`
}

// MarginAccounts stores margin accounts data
type MarginAccounts struct {
	Accounts  []MarginAccount `json:"accounts"`
	DebtRatio float64         `json:"debtRatio,string"`
}

// MarginRiskLimit stores margin risk limit
type MarginRiskLimit struct {
	Currency            string  `json:"currency"`
	MaximumBorrowAmount float64 `json:"borrowMaxAmount,string"`
	MaxumumBuyAmount    float64 `json:"buyMaxAmount,string"`
	MaximumHoldAmount   float64 `json:"holdMaxAmount,string"`
	Precision           int64   `json:"precision"`
}

// MarginBorrowParam represents a margin borrow parameter
type MarginBorrowParam struct {
	Currency    currency.Code `json:"currency"`
	Size        float64       `json:"size"`
	IsIsolated  bool          `json:"isisolated"`
	Symbol      currency.Pair `json:"symbol"`
	TimeInForce string        `json:"timeInForce"`
}

// RepayParam represents a repayment parameter for cross and isolated margin orders.
type RepayParam struct {
	Currency   currency.Code `json:"currency"`
	Size       float64       `json:"size"`
	IsIsolated bool          `json:"isisolated"`
	Symbol     currency.Pair `json:"symbol"`
}

// BorrowAndRepaymentOrderResp stores borrow order response
type BorrowAndRepaymentOrderResp struct {
	OrderNo    string  `json:"orderNo"`
	ActualSize float64 `json:"actualSize"`
}

// BorrowRepayDetailItem represents a borrow and repay order detail
type BorrowRepayDetailItem struct {
	OrderNo     string               `json:"orderNo"`
	Symbol      string               `json:"symbol"`
	Currency    string               `json:"currency"`
	Size        float64              `json:"size"`
	ActualSize  float64              `json:"actualSize"`
	Status      string               `json:"status"`
	CreatedTime convert.ExchangeTime `json:"createdTime"`
}

// BorrowOrder stores borrow order
type BorrowOrder struct {
	OrderID   string  `json:"orderId"`
	Currency  string  `json:"currency"`
	Size      float64 `json:"size,string"`
	Filled    float64 `json:"filled"`
	MatchList []struct {
		TradeID      string               `json:"tradeId"`
		Currency     string               `json:"currency"`
		DailyIntRate float64              `json:"dailyIntRate,string"`
		Size         float64              `json:"size,string"`
		Term         int64                `json:"term"`
		Timestamp    convert.ExchangeTime `json:"timestamp"`
	} `json:"matchList"`
	Status string `json:"status"`
}

type baseRecord struct {
	TradeID      string  `json:"tradeId"`
	Currency     string  `json:"currency"`
	DailyIntRate float64 `json:"dailyIntRate,string"`
	Principal    float64 `json:"principal,string"`
	RepaidSize   float64 `json:"repaidSize,string"`
	Term         int64   `json:"term"`
}

// RepaidRecordsResponse stores list of repaid record details.
type RepaidRecordsResponse struct {
	CurrentPage int64          `json:"currentPage"`
	PageSize    int64          `json:"pageSize"`
	TotalNumber int64          `json:"totalNum"`
	TotalPage   int64          `json:"totalPage"`
	Items       []RepaidRecord `json:"items"`
}

// RepaidRecord stores repaid record
type RepaidRecord struct {
	baseRecord
	Interest  float64              `json:"interest,string"`
	RepayTime convert.ExchangeTime `json:"repayTime"`
}

// LendOrder stores lend order
type LendOrder struct {
	OrderID      string               `json:"orderId"`
	Currency     string               `json:"currency"`
	Size         float64              `json:"size,string"`
	FilledSize   float64              `json:"filledSize,string"`
	DailyIntRate float64              `json:"dailyIntRate,string"`
	Term         int64                `json:"term"`
	CreatedAt    convert.ExchangeTime `json:"createdAt"`
}

// IsolatedMarginPairConfig current isolated margin trading pair configuration
type IsolatedMarginPairConfig struct {
	Symbol                string  `json:"symbol"`
	SymbolName            string  `json:"symbolName"`
	BaseCurrency          string  `json:"baseCurrency"`
	QuoteCurrency         string  `json:"quoteCurrency"`
	MaxLeverage           int64   `json:"maxLeverage"`
	LiquidationDebtRatio  float64 `json:"flDebtRatio,string"`
	TradeEnable           bool    `json:"tradeEnable"`
	AutoRenewMaxDebtRatio float64 `json:"autoRenewMaxDebtRatio,string"`
	BaseBorrowEnable      bool    `json:"baseBorrowEnable"`
	QuoteBorrowEnable     bool    `json:"quoteBorrowEnable"`
	BaseTransferInEnable  bool    `json:"baseTransferInEnable"`
	QuoteTransferInEnable bool    `json:"quoteTransferInEnable"`
}

type baseAsset struct {
	Currency         string  `json:"currency"`
	TotalBalance     float64 `json:"totalBalance,string"`
	HoldBalance      float64 `json:"holdBalance,string"`
	AvailableBalance float64 `json:"availableBalance,string"`
	Liability        float64 `json:"liability,string"`
	Interest         float64 `json:"interest,string"`
	BorrowableAmount float64 `json:"borrowableAmount,string"`
}

// AssetInfo holds asset information for an instrument.
type AssetInfo struct {
	Symbol     string    `json:"symbol"`
	Status     string    `json:"status"`
	DebtRatio  float64   `json:"debtRatio,string"`
	BaseAsset  baseAsset `json:"baseAsset"`
	QuoteAsset baseAsset `json:"quoteAsset"`
}

// IsolatedMarginAccountInfo holds isolated margin accounts of the current user
type IsolatedMarginAccountInfo struct {
	TotalConversionBalance     float64     `json:"totalConversionBalance,string"`
	LiabilityConversionBalance float64     `json:"liabilityConversionBalance,string"`
	Assets                     []AssetInfo `json:"assets"`
}

type baseRepaymentRecord struct {
	LoanID            string               `json:"loanId"`
	Symbol            string               `json:"symbol"`
	Currency          string               `json:"currency"`
	PrincipalTotal    float64              `json:"principalTotal,string"`
	InterestBalance   float64              `json:"interestBalance,string"`
	CreatedAt         convert.ExchangeTime `json:"createdAt"`
	Period            int64                `json:"period"`
	RepaidSize        float64              `json:"repaidSize,string"`
	DailyInterestRate float64              `json:"dailyInterestRate,string"`
}

// ServiceStatus represents a service status message.
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

// PlaceHFParam represents a place HF order parameters.
type PlaceHFParam struct {
	ClientOrderID       string        `json:"clientOid,omitempty"`
	Symbol              currency.Pair `json:"symbol"`
	OrderType           string        `json:"type"`
	Side                string        `json:"side"`
	SelfTradePrevention string        `json:"stp,omitempty"`
	OrderTags           string        `json:"tags,omitempty"`
	Remark              string        `json:"remark,omitempty"`

	// Additional 'limit' order parameters
	Price       float64 `json:"price,string,omitempty"`
	Size        float64 `json:"size,string,omitempty"`
	TimeInForce string  `json:"timeInForce"`
	CancelAfter int64   `json:"cancelAfter"`
	PostOnly    bool    `json:"postOnly"`
	Hidden      bool    `json:"hidden"`
	Iceberg     bool    `json:"iceberg"`
	VisibleSize float64 `json:"visibleSize"`

	// Additional 'market' parameters
	Funds string `json:"funds"`
}

// ModifyHFOrderParam represents a modify high frequency order parameter
type ModifyHFOrderParam struct {
	Symbol        currency.Pair `json:"symbol"`
	ClientOrderID string        `json:"clientOid,omitempty"`
	OrderID       string        `json:"orderId,omitempty"`
	NewPrice      float64       `json:"newPrice,omitempty,string"`
	NewSize       float64       `json:"newSize,omitempty,string"`
}

// SyncCancelHFOrderResp represents a cancel sync high frequency order.
type SyncCancelHFOrderResp struct {
	OrderID      string       `json:"orderId"`
	OriginSize   types.Number `json:"originSize"`
	OriginFunds  string       `json:"originFunds"`
	DealSize     types.Number `json:"dealSize"`
	RemainSize   types.Number `json:"remainSize"`
	CanceledSize types.Number `json:"canceledSize"`
	Status       string       `json:"status"`
}

// PlaceOrderResp represents a place order response
type PlaceOrderResp struct {
	OrderID string `json:"orderId"`
	Success bool   `json:"success"`
}

// SyncPlaceHFOrderResp represents a request parameter for high frequency sync orders
type SyncPlaceHFOrderResp struct {
	OrderID       string               `json:"orderId"`
	OrderTime     convert.ExchangeTime `json:"orderTime"`
	OriginSize    types.Number         `json:"originSize"`
	OriginFunds   string               `json:"originFunds"`
	DealSize      types.Number         `json:"dealSize"`
	DealFunds     string               `json:"dealFunds"`
	RemainSize    types.Number         `json:"remainSize"`
	RemainFunds   types.Number         `json:"remainFunds"`
	CanceledSize  string               `json:"canceledSize"`
	CanceledFunds string               `json:"canceledFunds"`
	Status        string               `json:"status"`
	MatchTime     convert.ExchangeTime `json:"matchTime"`
}

// CancelOrderByNumberResponse represents response for canceling an order by number
type CancelOrderByNumberResponse struct {
	OrderID    string `json:"orderId"`
	CancelSize string `json:"cancelSize"`
}

// CancelAllHFOrdersResponse represents a response for cancelling all high-frequency orders.
type CancelAllHFOrdersResponse struct {
	SucceedSymbols []string `json:"succeedSymbols"`
	FailedSymbols  []struct {
		Symbol string `json:"symbol"`
		Error  string `json:"error"`
	} `json:"failedSymbols"`
}

// HFOrder represents a high-frequency order instance.
type HFOrder struct {
	ID                  string               `json:"id"`
	Symbol              string               `json:"symbol"`
	OpType              string               `json:"opType"`
	Type                string               `json:"type"`
	Side                string               `json:"side"`
	Price               types.Number         `json:"price"`
	Size                types.Number         `json:"size"`
	Funds               string               `json:"funds"`
	DealFunds           string               `json:"dealFunds"`
	DealSize            types.Number         `json:"dealSize"`
	Fee                 types.Number         `json:"fee"`
	FeeCurrency         string               `json:"feeCurrency"`
	SelfTradePrevention string               `json:"stp"`
	TimeInForce         string               `json:"timeInForce"`
	PostOnly            bool                 `json:"postOnly"`
	Hidden              bool                 `json:"hidden"`
	Iceberg             bool                 `json:"iceberg"`
	VisibleSize         types.Number         `json:"visibleSize"`
	CancelAfter         int64                `json:"cancelAfter"`
	Channel             string               `json:"channel"`
	ClientOid           string               `json:"clientOid"`
	Remark              string               `json:"remark"`
	Tags                string               `json:"tags"`
	Active              bool                 `json:"active"`
	InOrderBook         bool                 `json:"inOrderBook"`
	CancelExist         bool                 `json:"cancelExist"`
	CreatedAt           convert.ExchangeTime `json:"createdAt"`
	LastUpdatedAt       convert.ExchangeTime `json:"lastUpdatedAt"`
	TradeType           string               `json:"tradeType"`
	CancelledSize       types.Number         `json:"cancelledSize"`
	CancelledFunds      string               `json:"cancelledFunds"`
	RemainSize          types.Number         `json:"remainSize"`
	RemainFunds         string               `json:"remainFunds"`
}

// CompletedHFOrder represents a completed HF orders list
type CompletedHFOrder struct {
	LastID int64     `json:"lastId"`
	Items  []HFOrder `json:"items"`
}

// AutoCancelHFOrderResponse represents an auto cancel HF order response
type AutoCancelHFOrderResponse struct {
	Timeout     int64                `json:"timeout"`
	Symbols     string               `json:"symbols"`
	CurrentTime convert.ExchangeTime `json:"currentTime"`
	TriggerTime convert.ExchangeTime `json:"triggerTime"`
}

// HFOrderFill represents an HF order fill
type HFOrderFill struct {
	ID             int64                `json:"id"`
	Symbol         string               `json:"symbol"`
	TradeID        int64                `json:"tradeId"`
	OrderID        string               `json:"orderId"`
	CounterOrderID string               `json:"counterOrderId"`
	Side           string               `json:"side"`
	Liquidity      string               `json:"liquidity"`
	ForceTaker     bool                 `json:"forceTaker"`
	Price          types.Number         `json:"price"`
	Size           types.Number         `json:"size"`
	Funds          string               `json:"funds"`
	Fee            types.Number         `json:"fee"`
	FeeRate        types.Number         `json:"feeRate"`
	FeeCurrency    string               `json:"feeCurrency"`
	Stop           string               `json:"stop"`
	TradeType      string               `json:"tradeType"`
	OrderType      string               `json:"type"`
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
}

// HFOrderFills represents an HF order list
type HFOrderFills struct {
	Items  []HFOrderFill `json:"items"`
	LastID int64         `json:"lastId"`
}

// PlaceOrderParams represents a batch place order parameters.
type PlaceOrderParams struct {
	OrderList []PlaceHFParam `json:"orderList"`
}

// CompletedRepaymentRecord represents repayment records of isolated margin positions
type CompletedRepaymentRecord struct {
	baseRepaymentRecord
	RepayFinishAt convert.ExchangeTime `json:"repayFinishAt"`
}

// PostMarginOrderResp represents response data for placing margin orders
type PostMarginOrderResp struct {
	OrderID     string  `json:"orderId"`
	BorrowSize  float64 `json:"borrowSize"`
	LoanApplyID string  `json:"loanApplyId"`
}

// OrderRequest represents place order request parameters
type OrderRequest struct {
	ClientOID   string  `json:"clientOid"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Type        string  `json:"type,omitempty"`             // optional
	Remark      string  `json:"remark,omitempty"`           // optional
	Stop        string  `json:"stop,omitempty"`             // optional
	StopPrice   float64 `json:"stopPrice,string,omitempty"` // optional
	STP         string  `json:"stp,omitempty"`              // optional
	Price       float64 `json:"price,string,omitempty"`
	Size        float64 `json:"size,string,omitempty"`
	TimeInForce string  `json:"timeInForce,omitempty"` // optional
	CancelAfter int64   `json:"cancelAfter,omitempty"` // optional
	PostOnly    bool    `json:"postOnly,omitempty"`    // optional
	Hidden      bool    `json:"hidden,omitempty"`      // optional
	Iceberg     bool    `json:"iceberg,omitempty"`     // optional
	VisibleSize string  `json:"visibleSize,omitempty"` // optional
}

// PostBulkOrderResp response data for submitting a bulk order
type PostBulkOrderResp struct {
	OrderRequest
	Channel string `json:"channel"`
	ID      string `json:"id"`
	Status  string `json:"status"`
	FailMsg string `json:"failMsg"`
}

// OrdersListResponse represents an order list response.
type OrdersListResponse struct {
	CurrentPage int64         `json:"currentPage"`
	PageSize    int64         `json:"pageSize"`
	TotalNum    int64         `json:"totalNum"`
	TotalPage   int64         `json:"totalPage"`
	Items       []OrderDetail `json:"items"`
}

// OrderDetail represents order detail
type OrderDetail struct {
	OrderRequest
	Channel       string               `json:"channel"`
	ID            string               `json:"id"`
	OperationType string               `json:"opType"` // operation type: DEAL
	Funds         string               `json:"funds"`
	DealFunds     string               `json:"dealFunds"`
	DealSize      float64              `json:"dealSize,string"`
	Fee           float64              `json:"fee,string"`
	FeeCurrency   string               `json:"feeCurrency"`
	StopTriggered bool                 `json:"stopTriggered"`
	Tags          string               `json:"tags"`
	IsActive      bool                 `json:"isActive"`
	CancelExist   bool                 `json:"cancelExist"`
	CreatedAt     convert.ExchangeTime `json:"createdAt"`
	TradeType     string               `json:"tradeType"`
}

// ListFills represents fills response list detail.
type ListFills struct {
	CurrentPage int64  `json:"currentPage"`
	PageSize    int64  `json:"pageSize"`
	TotalNumber int64  `json:"totalNum"`
	TotalPage   int64  `json:"totalPage"`
	Items       []Fill `json:"items"`
}

// Fill represents order fills for margin and spot orders.
type Fill struct {
	Symbol         string               `json:"symbol"`
	TradeID        string               `json:"tradeId"`
	OrderID        string               `json:"orderId"`
	CounterOrderID string               `json:"counterOrderId"`
	Side           string               `json:"side"`
	Liquidity      string               `json:"liquidity"`
	ForceTaker     bool                 `json:"forceTaker"`
	Price          float64              `json:"price,string"`
	Size           float64              `json:"size,string"`
	Funds          float64              `json:"funds,string"`
	Fee            float64              `json:"fee,string"`
	FeeRate        float64              `json:"feeRate,string"`
	FeeCurrency    string               `json:"feeCurrency"`
	Stop           string               `json:"stop"`
	OrderType      string               `json:"type"`
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
	TradeType      string               `json:"tradeType"`
}

// StopOrderListResponse represents a list of spot orders details.
type StopOrderListResponse struct {
	CurrentPage int64       `json:"currentPage"`
	PageSize    int64       `json:"pageSize"`
	TotalNumber int64       `json:"totalNum"`
	TotalPage   int64       `json:"totalPage"`
	Items       []StopOrder `json:"items"`
}

// StopOrder holds a stop order detail
type StopOrder struct {
	OrderRequest
	ID              string               `json:"id"`
	UserID          string               `json:"userId"`
	Status          string               `json:"status"`
	Funds           float64              `json:"funds,string"`
	Channel         string               `json:"channel"`
	Tags            string               `json:"tags"`
	DomainID        string               `json:"domainId"`
	TradeSource     string               `json:"tradeSource"`
	TradeType       string               `json:"tradeType"`
	FeeCurrency     string               `json:"feeCurrency"`
	TakerFeeRate    string               `json:"takerFeeRate"`
	MakerFeeRate    string               `json:"makerFeeRate"`
	CreatedAt       convert.ExchangeTime `json:"createdAt"`
	OrderTime       convert.ExchangeTime `json:"orderTime"`
	StopTriggerTime convert.ExchangeTime `json:"stopTriggerTime"`
}

// AccountInfo represents account information
type AccountInfo struct {
	ID          string       `json:"id"`
	Currency    string       `json:"currency"`
	AccountType string       `json:"type"` // Account type:，main、trade、trade_hf、margin
	Balance     types.Number `json:"balance"`
	Available   types.Number `json:"available"`
	Holds       types.Number `json:"holds"`
}

// CrossMarginAccountDetail represents a cross-margin account details.
type CrossMarginAccountDetail struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Retry   bool   `json:"retry"`
	Data    struct {
		Timestamp   convert.ExchangeTime `json:"timestamp"`
		CurrentPage int64                `json:"currentPage"`
		PageSize    int64                `json:"pageSize"`
		TotalNum    int64                `json:"totalNum"`
		TotalPage   int64                `json:"totalPage"`
		Items       []struct {
			TotalLiabilityOfQuoteCurrency string `json:"totalLiabilityOfQuoteCurrency"`
			TotalAssetOfQuoteCurrency     string `json:"totalAssetOfQuoteCurrency"`
			DebtRatio                     string `json:"debtRatio"`
			Status                        string `json:"status"`
			Assets                        []struct {
				Currency        string       `json:"currency"`
				BorrowEnabled   bool         `json:"borrowEnabled"`
				RepayEnabled    bool         `json:"repayEnabled"`
				TransferEnabled bool         `json:"transferEnabled"`
				Borrowed        string       `json:"borrowed"`
				TotalAsset      types.Number `json:"totalAsset"`
				Available       types.Number `json:"available"`
				Hold            types.Number `json:"hold"`
				MaxBorrowSize   types.Number `json:"maxBorrowSize"`
			} `json:"assets"`
		} `json:"items"`
	} `json:"data"`
}

// IsolatedMarginAccountDetail represents an isolated-margin account detail.
type IsolatedMarginAccountDetail struct {
	Code string `json:"code"`
	Data []struct {
		TotalAssetOfQuoteCurrency     string               `json:"totalAssetOfQuoteCurrency"`
		TotalLiabilityOfQuoteCurrency string               `json:"totalLiabilityOfQuoteCurrency"`
		Timestamp                     convert.ExchangeTime `json:"timestamp"`
		Assets                        []struct {
			Symbol    string `json:"symbol"`
			DebtRatio string `json:"debtRatio"`
			Status    string `json:"status"`
			BaseAsset struct {
				Currency        string       `json:"currency"`
				BorrowEnabled   bool         `json:"borrowEnabled"`
				RepayEnabled    bool         `json:"repayEnabled"`
				TransferEnabled bool         `json:"transferEnabled"`
				Borrowed        string       `json:"borrowed"`
				TotalAsset      types.Number `json:"totalAsset"`
				Available       types.Number `json:"available"`
				Hold            types.Number `json:"hold"`
				MaxBorrowSize   types.Number `json:"maxBorrowSize"`
			} `json:"baseAsset"`
			QuoteAsset struct {
				Currency        string       `json:"currency"`
				BorrowEnabled   bool         `json:"borrowEnabled"`
				RepayEnabled    bool         `json:"repayEnabled"`
				TransferEnabled bool         `json:"transferEnabled"`
				Borrowed        string       `json:"borrowed"`
				TotalAsset      types.Number `json:"totalAsset"`
				Available       types.Number `json:"available"`
				Hold            types.Number `json:"hold"`
				MaxBorrowSize   types.Number `json:"maxBorrowSize"`
			} `json:"quoteAsset"`
		} `json:"assets"`
	} `json:"data"`
}

// FuturesAccountOverview represents a futures account detail.
type FuturesAccountOverview struct {
	Code string `json:"code"`
	Data struct {
		AccountEquity    float64 `json:"accountEquity"`
		UnrealisedPNL    int     `json:"unrealisedPNL"`
		MarginBalance    float64 `json:"marginBalance"`
		PositionMargin   int     `json:"positionMargin"`
		OrderMargin      int     `json:"orderMargin"`
		FrozenFunds      int     `json:"frozenFunds"`
		AvailableBalance float64 `json:"availableBalance"`
		Currency         string  `json:"currency"`
	} `json:"data"`
}

// AccountBalance represents an account balance detail.
type AccountBalance struct {
	Currency          string       `json:"currency"`
	Balance           types.Number `json:"balance"`
	Available         types.Number `json:"available"`
	Holds             types.Number `json:"holds"`
	BaseCurrency      string       `json:"baseCurrency"`
	BaseCurrencyPrice types.Number `json:"baseCurrencyPrice"`
	BaseAmount        types.Number `json:"baseAmount"`
}

// SubAccounts represents subacounts detail and balances
type SubAccounts struct {
	SubUserID      string           `json:"subUserId"`
	SubName        string           `json:"subName"`
	MainAccounts   []AccountBalance `json:"mainAccounts"`
	TradeAccounts  []AccountBalance `json:"tradeAccounts"`
	MarginAccounts []AccountBalance `json:"marginAccounts"`
}

// FuturesSubAccountBalance represents a subaccount balances for futures trading
type FuturesSubAccountBalance struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Retry   bool   `json:"retry"`
	Data    struct {
		Summary struct {
			AccountEquityTotal    float64 `json:"accountEquityTotal"`
			UnrealisedPNLTotal    float64 `json:"unrealisedPNLTotal"`
			MarginBalanceTotal    float64 `json:"marginBalanceTotal"`
			PositionMarginTotal   float64 `json:"positionMarginTotal"`
			OrderMarginTotal      float64 `json:"orderMarginTotal"`
			FrozenFundsTotal      float64 `json:"frozenFundsTotal"`
			AvailableBalanceTotal float64 `json:"availableBalanceTotal"`
			Currency              string  `json:"currency"`
		} `json:"summary"`
		Accounts []struct {
			AccountName      float64 `json:"accountName"`
			AccountEquity    float64 `json:"accountEquity"`
			UnrealisedPNL    float64 `json:"unrealisedPNL"`
			MarginBalance    float64 `json:"marginBalance"`
			PositionMargin   float64 `json:"positionMargin"`
			OrderMargin      float64 `json:"orderMargin"`
			FrozenFunds      float64 `json:"frozenFunds"`
			AvailableBalance float64 `json:"availableBalance"`
			Currency         string  `json:"currency"`
		} `json:"accounts"`
	} `json:"data"`
}

// LedgerInfo represents account ledger information.
type LedgerInfo struct {
	ID          string               `json:"id"`
	Currency    string               `json:"currency"`
	Amount      types.Number         `json:"amount"`
	Fee         types.Number         `json:"fee"`
	Balance     types.Number         `json:"balance"`
	AccountType string               `json:"accountType"`
	BizType     string               `json:"bizType"`
	Direction   string               `json:"direction"`
	CreatedAt   convert.ExchangeTime `json:"createdAt"`
	Context     string               `json:"context"`
}

// FuturesLedgerInfo represents account ledger information for futures trading.
type FuturesLedgerInfo struct {
	Code string `json:"code"`
	Data struct {
		HasMore  bool `json:"hasMore"`
		DataList []struct {
			Time          convert.ExchangeTime `json:"time"`
			Type          string               `json:"type"`
			Amount        float64              `json:"amount"`
			Fee           float64              `json:"fee"`
			AccountEquity float64              `json:"accountEquity"`
			Status        string               `json:"status"`
			Remark        string               `json:"remark"`
			Offset        int64                `json:"offset"`
			Currency      string               `json:"currency"`
		} `json:"dataList"`
	} `json:"data"`
}

// MainAccountInfo represents main account detailed information.
type MainAccountInfo struct {
	AccountInfo
	BaseCurrency      string  `json:"baseCurrency"`
	BaseCurrencyPrice float64 `json:"baseCurrencyPrice,string"`
	BaseAmount        float64 `json:"baseAmount,string"`
}

// AccountSummaryInformation represents account summary information detail.
type AccountSummaryInformation struct {
	Data struct {
		Level                 float64 `json:"level"`
		SubQuantity           float64 `json:"subQuantity"`
		MaxDefaultSubQuantity float64 `json:"maxDefaultSubQuantity"`
		MaxSubQuantity        float64 `json:"maxSubQuantity"`
		SpotSubQuantity       float64 `json:"spotSubQuantity"`
		MarginSubQuantity     float64 `json:"marginSubQuantity"`
		FuturesSubQuantity    float64 `json:"futuresSubQuantity"`
		MaxSpotSubQuantity    float64 `json:"maxSpotSubQuantity"`
		MaxMarginSubQuantity  float64 `json:"maxMarginSubQuantity"`
		MaxFuturesSubQuantity float64 `json:"maxFuturesSubQuantity"`
	} `json:"data"`
	Code string `json:"code"`
}

// SubAccountsResponse represents a sub-accounts items response instance.
type SubAccountsResponse struct {
	CurrentPage int64            `json:"currentPage"`
	PageSize    int64            `json:"pageSize"`
	TotalNumber int64            `json:"totalNum"`
	TotalPage   int64            `json:"totalPage"`
	Items       []SubAccountInfo `json:"items"`
}

// SubAccountInfo holds subaccount data for main, spot(trade), and margin accounts.
type SubAccountInfo struct {
	SubUserID      string            `json:"subUserId"`
	SubName        string            `json:"subName"`
	MainAccounts   []MainAccountInfo `json:"mainAccounts"`
	TradeAccounts  []MainAccountInfo `json:"tradeAccounts"`
	MarginAccounts []MainAccountInfo `json:"marginAccounts"`
}

// SubAccountBalanceV2 represents a sub-account balance detail through the V2 API.
type SubAccountBalanceV2 struct {
	Code string `json:"code"`
	Data struct {
		CurrentPage int64 `json:"currentPage"`
		PageSize    int64 `json:"pageSize"`
		TotalNum    int64 `json:"totalNum"`
		TotalPage   int64 `json:"totalPage"`
		Items       []struct {
			SubUserID    string `json:"subUserId"`
			SubName      string `json:"subName"`
			MainAccounts []struct {
				Currency          string       `json:"currency"`
				Balance           types.Number `json:"balance"`
				Available         types.Number `json:"available"`
				Holds             types.Number `json:"holds"`
				BaseCurrency      string       `json:"baseCurrency"`
				BaseCurrencyPrice types.Number `json:"baseCurrencyPrice"`
				BaseAmount        types.Number `json:"baseAmount"`
			} `json:"mainAccounts"`
		} `json:"items"`
	} `json:"data"`
}

// TransferableBalanceInfo represents transferable balance information
type TransferableBalanceInfo struct {
	AccountInfo
	Transferable float64 `json:"transferable,string"`
}

// FundTransferFuturesParam holds parameter values for internal transfer
type FundTransferFuturesParam struct {
	Amount             float64       `json:"amount"`
	Currency           currency.Code `json:"currency"`
	RecieveAccountType string        `json:"recAccountType"` // possible values are: MAIN and TRADE
}

// FundTransferToFuturesParam holds request parameters to transfer funds to futures account.
type FundTransferToFuturesParam struct {
	Amount             float64       `json:"amount"`
	Currency           currency.Code `json:"currency"`
	PaymentAccountType string        `json:"payAccountType"` // Payment account type, including MAIN,TRADE
}

// InnerTransferToMainAndTradeResponse represents a detailed response after transferring fund to main and trade accounts.
type InnerTransferToMainAndTradeResponse struct {
	ApplyID        string               `json:"applyId"`
	BizNo          string               `json:"bizNo"`
	PayAccountType string               `json:"payAccountType"`
	PayTag         string               `json:"payTag"`
	Remark         string               `json:"remark"`
	RecAccountType string               `json:"recAccountType"`
	RecTag         string               `json:"recTag"`
	RecRemark      string               `json:"recRemark"`
	RecSystem      string               `json:"recSystem"`
	Status         string               `json:"status"`
	Currency       string               `json:"currency"`
	Amount         types.Number         `json:"amount"`
	Fee            types.Number         `json:"fee"`
	SerialNumber   int64                `json:"sn"`
	Reason         string               `json:"reason"`
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
	UpdatedAt      convert.ExchangeTime `json:"updatedAt"`
}

// FundTransferToFuturesResponse response struct after transferring fund to Futures account
type FundTransferToFuturesResponse struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
	Retry   bool   `json:"retry"`
	Success bool   `json:"success"`
}

// FuturesTransferOutResponse represents a list of transfer out instance.
type FuturesTransferOutResponse struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalNum    int `json:"totalNum"`
	TotalPage   int `json:"totalPage"`
	Items       []struct {
		ApplyID   string               `json:"applyId"`
		Currency  string               `json:"currency"`
		RecRemark string               `json:"recRemark"`
		RecSystem string               `json:"recSystem"`
		Status    string               `json:"status"`
		Amount    types.Number         `json:"amount"`
		Reason    string               `json:"reason"`
		Offset    int64                `json:"offset"`
		CreatedAt convert.ExchangeTime `json:"createdAt"`
		Remark    string               `json:"remark"`
	} `json:"items"`
}

// DepositAddressParams represents a deposit address creation parameters
type DepositAddressParams struct {
	Currency currency.Code `json:"currency"`
	Chain    string        `json:"chain,omitempty"`
}

// DepositAddress represents deposit address information for Spot and Margin trading.
type DepositAddress struct {
	Address string `json:"address"`
	Memo    string `json:"memo"`
	Chain   string `json:"chain"`

	// TODO: to be removed if not used by other endpoints
	ContractAddress string `json:"contractAddress"` // missing in case of futures
}

type baseDeposit struct {
	Currency   string `json:"currency"`
	WalletTxID string `json:"walletTxId"`
	IsInner    bool   `json:"isInner"`
	Status     string `json:"status"`
}

// DepositResponse represents a detailed response for list of deposit.
type DepositResponse struct {
	CurrentPage int64     `json:"currentPage"`
	PageSize    int64     `json:"pageSize"`
	TotalNum    int64     `json:"totalNum"`
	TotalPage   int64     `json:"totalPage"`
	Items       []Deposit `json:"items"`
}

// Deposit represents deposit address and detail and timestamp information.
type Deposit struct {
	baseDeposit
	Amount    float64              `json:"amount,string"`
	Address   string               `json:"address"`
	Memo      string               `json:"memo"`
	Fee       float64              `json:"fee,string"`
	Remark    string               `json:"remark"`
	CreatedAt convert.ExchangeTime `json:"createdAt"`
	UpdatedAt convert.ExchangeTime `json:"updatedAt"`
	Chain     string               `json:"chain"`
}

// HistoricalDepositWithdrawalResponse represents deposit and withdrawal funding items details.
type HistoricalDepositWithdrawalResponse struct {
	CurrentPage int64                         `json:"currentPage"`
	PageSize    int64                         `json:"pageSize"`
	TotalNum    int64                         `json:"totalNum"`
	TotalPage   int64                         `json:"totalPage"`
	Items       []HistoricalDepositWithdrawal `json:"items"`
}

// HistoricalDepositWithdrawal represents deposit and withdrawal funding item.
type HistoricalDepositWithdrawal struct {
	baseDeposit
	Amount    float64              `json:"amount,string"`
	CreatedAt convert.ExchangeTime `json:"createAt"`
}

// WithdrawalsResponse represents a withdrawals list of items details.
type WithdrawalsResponse struct {
	CurrentPage int64        `json:"currentPage"`
	PageSize    int64        `json:"pageSize"`
	TotalNum    int64        `json:"totalNum"`
	TotalPage   int64        `json:"totalPage"`
	Items       []Withdrawal `json:"items"`
}

// Withdrawal represents withdrawal funding information.
type Withdrawal struct {
	Deposit
	ID string `json:"id"`
}

// WithdrawalQuota represents withdrawal quota detail information.
type WithdrawalQuota struct {
	Currency                 string       `json:"currency"`
	LimitBTCAmount           float64      `json:"limitBTCAmount,string"`
	UsedBTCAmount            float64      `json:"usedBTCAmount,string"`
	RemainAmount             float64      `json:"remainAmount,string"`
	AvailableAmount          float64      `json:"availableAmount,string"`
	WithdrawMinFee           float64      `json:"withdrawMinFee,string"`
	InnerWithdrawMinFee      float64      `json:"innerWithdrawMinFee,string"`
	WithdrawMinSize          float64      `json:"withdrawMinSize,string"`
	IsWithdrawEnabled        bool         `json:"isWithdrawEnabled"`
	Precision                int64        `json:"precision"`
	Chain                    string       `json:"chain"`
	QuotaCurrency            string       `json:"quotaCurrency"`
	LimitQuotaCurrencyAmount types.Number `json:"limitQuotaCurrencyAmount"`
	Reason                   string       `json:"reason"`
	UsedQuotaCurrencyAmount  string       `json:"usedQuotaCurrencyAmount"`
}

// Fees represents taker and maker fee information a symbol.
type Fees struct {
	Symbol       string  `json:"symbol"`
	TakerFeeRate float64 `json:"takerFeeRate,string"`
	MakerFeeRate float64 `json:"makerFeeRate,string"`
}

// WSInstanceServers response connection token and websocket instance server information.
type WSInstanceServers struct {
	Token           string           `json:"token"`
	InstanceServers []InstanceServer `json:"instanceServers"`
}

// InstanceServer represents a single websocket instance server information.
type InstanceServer struct {
	Endpoint     string `json:"endpoint"`
	Encrypt      bool   `json:"encrypt"`
	Protocol     string `json:"protocol"`
	PingInterval int64  `json:"pingInterval"`
	PingTimeout  int64  `json:"pingTimeout"`
}

// WSConnMessages represents response messages ping, pong, and welcome message structures.
type WSConnMessages struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// WsSubscriptionInput represents a subscription information structure.
type WsSubscriptionInput struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Topic          string `json:"topic"`
	PrivateChannel bool   `json:"privateChannel"`
	Response       bool   `json:"response,omitempty"`
}

// WsPushData represents a push data from a server.
type WsPushData struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Topic       string          `json:"topic"`
	UserID      string          `json:"userId"`
	Subject     string          `json:"subject"`
	ChannelType string          `json:"channelType"`
	Data        json.RawMessage `json:"data"`
}

// WsTicker represents a ticker push data from server.
type WsTicker struct {
	Sequence    string               `json:"sequence"`
	BestAsk     float64              `json:"bestAsk,string"`
	Size        float64              `json:"size,string"`
	BestBidSize float64              `json:"bestBidSize,string"`
	Price       float64              `json:"price,string"`
	BestAskSize float64              `json:"bestAskSize,string"`
	BestBid     float64              `json:"bestBid,string"`
	Timestamp   convert.ExchangeTime `json:"time"`
}

// WsSnapshot represents a spot ticker push data.
type WsSnapshot struct {
	Sequence types.Number     `json:"sequence"`
	Data     WsSnapshotDetail `json:"data"`
}

// WsSnapshotDetail represents the detail of a spot ticker data.
// This represents all websocket ticker information pushed as a result of subscription to /market/snapshot:{symbol}, and /market/snapshot:{currency,market}
type WsSnapshotDetail struct {
	AveragePrice     float64              `json:"averagePrice"`
	BaseCurrency     string               `json:"baseCurrency"`
	Board            int64                `json:"board"`
	Buy              float64              `json:"buy"`
	ChangePrice      float64              `json:"changePrice"`
	ChangeRate       float64              `json:"changeRate"`
	Close            float64              `json:"close"`
	Datetime         convert.ExchangeTime `json:"datetime"`
	High             float64              `json:"high"`
	LastTradedPrice  float64              `json:"lastTradedPrice"`
	Low              float64              `json:"low"`
	MakerCoefficient float64              `json:"makerCoefficient"`
	MakerFeeRate     float64              `json:"makerFeeRate"`
	MarginTrade      bool                 `json:"marginTrade"`
	Mark             float64              `json:"mark"`
	Market           string               `json:"market"`
	Markets          []string             `json:"markets"`
	Open             float64              `json:"open"`
	QuoteCurrency    string               `json:"quoteCurrency"`
	Sell             float64              `json:"sell"`
	Sort             int64                `json:"sort"`
	Symbol           string               `json:"symbol"`
	SymbolCode       string               `json:"symbolCode"`
	TakerCoefficient float64              `json:"takerCoefficient"`
	TakerFeeRate     float64              `json:"takerFeeRate"`
	Trading          bool                 `json:"trading"`
	Vol              float64              `json:"vol"`
	VolValue         float64              `json:"volValue"`
}

// WsOrderbook represents orderbook information.
type WsOrderbook struct {
	Changes       OrderbookChanges     `json:"changes"`
	SequenceEnd   int64                `json:"sequenceEnd"`
	SequenceStart int64                `json:"sequenceStart"`
	Symbol        string               `json:"symbol"`
	TimeMS        convert.ExchangeTime `json:"time"`
}

// OrderbookChanges represents orderbook ask and bid changes.
type OrderbookChanges struct {
	Asks [][]types.Number `json:"asks"`
	Bids [][]types.Number `json:"bids"`
}

// WsCandlestickData represents candlestick information push data for a symbol.
type WsCandlestickData struct {
	Symbol  string               `json:"symbol"`
	Candles [7]string            `json:"candles"`
	Time    convert.ExchangeTime `json:"time"`
}

// WsCandlestick represents candlestick information push data for a symbol.
type WsCandlestick struct {
	Symbol  string `json:"symbol"`
	Candles struct {
		StartTime         time.Time
		OpenPrice         float64
		ClosePrice        float64
		HighPrice         float64
		LowPrice          float64
		TransactionVolume float64
		TransactionAmount float64
	} `json:"candles"`
	Time time.Time `json:"time"`
}

func (a *WsCandlestickData) getCandlestickData() (*WsCandlestick, error) {
	cand := &WsCandlestick{
		Symbol: a.Symbol,
		Time:   a.Time.Time(),
	}
	timeStamp, err := strconv.ParseInt(a.Candles[0], 10, 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.StartTime = time.UnixMilli(timeStamp)
	cand.Candles.OpenPrice, err = strconv.ParseFloat(a.Candles[1], 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.ClosePrice, err = strconv.ParseFloat(a.Candles[2], 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.HighPrice, err = strconv.ParseFloat(a.Candles[3], 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.LowPrice, err = strconv.ParseFloat(a.Candles[4], 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.TransactionVolume, err = strconv.ParseFloat(a.Candles[5], 64)
	if err != nil {
		return nil, err
	}
	cand.Candles.TransactionAmount, err = strconv.ParseFloat(a.Candles[6], 64)
	if err != nil {
		return nil, err
	}
	return cand, nil
}

// WsTrade represents a trade push data.
type WsTrade struct {
	Sequence     string               `json:"sequence"`
	Type         string               `json:"type"`
	Symbol       string               `json:"symbol"`
	Side         string               `json:"side"`
	Price        float64              `json:"price,string"`
	Size         float64              `json:"size,string"`
	TradeID      string               `json:"tradeId"`
	TakerOrderID string               `json:"takerOrderId"`
	MakerOrderID string               `json:"makerOrderId"`
	Time         convert.ExchangeTime `json:"time"`
}

// WsPriceIndicator represents index price or mark price indicator push data.
type WsPriceIndicator struct {
	Symbol      string               `json:"symbol"`
	Granularity float64              `json:"granularity"`
	Timestamp   convert.ExchangeTime `json:"timestamp"`
	Value       float64              `json:"value"`
}

// WsMarginFundingBook represents order book changes on margin.
type WsMarginFundingBook struct {
	Sequence           int64                `json:"sequence"`
	Currency           string               `json:"currency"`
	DailyInterestRate  float64              `json:"dailyIntRate"`
	AnnualInterestRate float64              `json:"annualIntRate"`
	Term               int64                `json:"term"`
	Size               float64              `json:"size"`
	Side               string               `json:"side"`
	Timestamp          convert.ExchangeTime `json:"ts"` // In Nanosecond

}

// WsTradeOrder represents a private trade order push data.
type WsTradeOrder struct {
	Symbol     string               `json:"symbol"`
	OrderType  string               `json:"orderType"`
	Side       string               `json:"side"`
	OrderID    string               `json:"orderId"`
	Type       string               `json:"type"`
	OrderTime  convert.ExchangeTime `json:"orderTime"`
	Size       float64              `json:"size,string"`
	FilledSize float64              `json:"filledSize,string"`
	Price      float64              `json:"price,string"`
	ClientOid  string               `json:"clientOid"`
	RemainSize float64              `json:"remainSize,string"`
	Status     string               `json:"status"`
	Timestamp  convert.ExchangeTime `json:"ts"`
	Liquidity  string               `json:"liquidity"`
	MatchPrice string               `json:"matchPrice"`
	MatchSize  string               `json:"matchSize"`
	TradeID    string               `json:"tradeId"`
	OldSize    string               `json:"oldSize"`
}

// WsAccountBalance represents a Account Balance push data.
type WsAccountBalance struct {
	Total           float64 `json:"total,string"`
	Available       float64 `json:"available,string"`
	AvailableChange float64 `json:"availableChange,string"`
	Currency        string  `json:"currency"`
	Hold            float64 `json:"hold,string"`
	HoldChange      float64 `json:"holdChange,string"`
	RelationEvent   string  `json:"relationEvent"`
	RelationEventID string  `json:"relationEventId"`
	RelationContext struct {
		Symbol  string `json:"symbol"`
		TradeID string `json:"tradeId"`
		OrderID string `json:"orderId"`
	} `json:"relationContext"`
	Time convert.ExchangeTime `json:"time"`
}

// WsDebtRatioChange represents a push data
type WsDebtRatioChange struct {
	DebtRatio float64              `json:"debtRatio"`
	TotalDebt float64              `json:"totalDebt,string"`
	DebtList  map[string]string    `json:"debtList"`
	Timestamp convert.ExchangeTime `json:"timestamp"`
}

// WsPositionStatus represents a position status push data.
type WsPositionStatus struct {
	Type        string               `json:"type"`
	TimestampMS convert.ExchangeTime `json:"timestamp"`
}

// WsMarginTradeOrderEntersEvent represents a push data to the lenders
// when the order enters the order book or when the order is executed.
type WsMarginTradeOrderEntersEvent struct {
	Currency     string               `json:"currency"`
	OrderID      string               `json:"orderId"`      // Trade ID
	DailyIntRate float64              `json:"dailyIntRate"` // Daily interest rate.
	Term         int64                `json:"term"`         // Term (Unit: Day)
	Size         float64              `json:"size"`         // Size
	LentSize     float64              `json:"lentSize"`     // Size executed -- filled when the subject is order.update
	Side         string               `json:"side"`         // Lend or borrow. Currently, only "Lend" is available
	Timestamp    convert.ExchangeTime `json:"ts"`           // Timestamp (nanosecond)
}

// WsMarginTradeOrderDoneEvent represents a push message to the lenders when the order is completed.
type WsMarginTradeOrderDoneEvent struct {
	Currency  string               `json:"currency"`
	OrderID   string               `json:"orderId"`
	Reason    string               `json:"reason"`
	Side      string               `json:"side"`
	Timestamp convert.ExchangeTime `json:"ts"`
}

// WsStopOrder represents a stop order.
// When a stop order is received by the system, you will receive a message with "open" type.
// It means that this order entered the system and waited to be triggered.
type WsStopOrder struct {
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
	OrderID        string               `json:"orderId"`
	OrderPrice     float64              `json:"orderPrice,string"`
	OrderType      string               `json:"orderType"`
	Side           string               `json:"side"`
	Size           float64              `json:"size,string"`
	Stop           string               `json:"stop"`
	StopPrice      float64              `json:"stopPrice,string"`
	Symbol         string               `json:"symbol"`
	TradeType      string               `json:"tradeType"`
	TriggerSuccess bool                 `json:"triggerSuccess"`
	Timestamp      convert.ExchangeTime `json:"ts"`
	Type           string               `json:"type"`
}

// WsFuturesTicker represents a futures ticker push data.
type WsFuturesTicker struct {
	Symbol       string               `json:"symbol"`
	Sequence     int64                `json:"sequence"`
	Side         string               `json:"side"`
	FilledPrice  float64              `json:"price"`
	FilledSize   float64              `json:"size"`
	TradeID      string               `json:"tradeId"`
	BestBidSize  float64              `json:"bestBidSize"`
	BestBidPrice types.Number         `json:"bestBidPrice"`
	BestAskPrice types.Number         `json:"bestAskPrice"`
	BestAskSize  float64              `json:"bestAskSize"`
	FilledTime   convert.ExchangeTime `json:"ts"`
}

// WsFuturesOrderbokInfo represents Level 2 order book information.
type WsFuturesOrderbokInfo struct {
	Sequence  int64                `json:"sequence"`
	Change    string               `json:"change"`
	Timestamp convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesExecutionData represents execution data for symbol.
type WsFuturesExecutionData struct {
	Sequence         int64                `json:"sequence"`
	FilledQuantity   float64              `json:"matchSize"` // Filled quantity
	UnfilledQuantity float64              `json:"size"`
	FilledPrice      float64              `json:"price"`
	TradeID          string               `json:"tradeId"`
	MakerUserID      string               `json:"makerUserId"`
	Symbol           string               `json:"symbol"`
	Side             string               `json:"side"`
	TakerOrderID     string               `json:"takerOrderId"`
	MakerOrderID     string               `json:"makerOrderId"`
	TakerUserID      string               `json:"takerUserId"`
	Timestamp        convert.ExchangeTime `json:"ts"`
}

// WsOrderbookLevel5 represents an orderbook push data with depth level 5.
type WsOrderbookLevel5 struct {
	Sequence      int64                `json:"sequence"`
	Asks          []orderbook.Item     `json:"asks"`
	Bids          []orderbook.Item     `json:"bids"`
	PushTimestamp convert.ExchangeTime `json:"ts"`
	Timestamp     convert.ExchangeTime `json:"timestamp"`
}

// WsOrderbookLevel5Response represents a response data for an orderbook push data with depth level 5.
type WsOrderbookLevel5Response struct {
	Timestamp     convert.ExchangeTime `json:"timestamp"`
	Sequence      int64                `json:"sequence"`
	Bids          [][2]types.Number    `json:"bids"`
	Asks          [][2]types.Number    `json:"asks"`
	PushTimestamp convert.ExchangeTime `json:"ts"`
}

// ExtractOrderbookItems returns WsOrderbookLevel5 instance from WsOrderbookLevel5Response
func (a *WsOrderbookLevel5Response) ExtractOrderbookItems() *WsOrderbookLevel5 {
	resp := WsOrderbookLevel5{
		Timestamp:     a.Timestamp,
		Sequence:      a.Sequence,
		PushTimestamp: a.PushTimestamp,
	}
	resp.Asks = make([]orderbook.Item, len(a.Asks))
	for x := range a.Asks {
		resp.Asks[x] = orderbook.Item{
			Price:  a.Asks[x][0].Float64(),
			Amount: a.Asks[x][1].Float64(),
		}
	}
	resp.Bids = make([]orderbook.Item, len(a.Bids))
	for x := range a.Bids {
		resp.Bids[x] = orderbook.Item{
			Price:  a.Bids[x][0].Float64(),
			Amount: a.Bids[x][1].Float64(),
		}
	}
	return &resp
}

// WsFundingRate represents the funding rate push data information through the websocket channel.
type WsFundingRate struct {
	Symbol      string               `json:"symbol"`
	Granularity int64                `json:"granularity"`
	FundingRate float64              `json:"fundingRate"`
	Timestamp   convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesMarkPriceAndIndexPrice represents mark price and index price information.
type WsFuturesMarkPriceAndIndexPrice struct {
	Symbol      string               `json:"symbol"`
	Granularity int64                `json:"granularity"`
	IndexPrice  float64              `json:"indexPrice"`
	MarkPrice   float64              `json:"markPrice"`
	Timestamp   convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesFundingBegin represents the Start Funding Fee Settlement.
type WsFuturesFundingBegin struct {
	Subject     string               `json:"subject"`
	Symbol      string               `json:"symbol"`
	FundingTime convert.ExchangeTime `json:"fundingTime"`
	FundingRate float64              `json:"fundingRate"`
	Timestamp   convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesTransactionStatisticsTimeEvent represents transaction statistics data.
type WsFuturesTransactionStatisticsTimeEvent struct {
	Symbol                   string               `json:"symbol"`
	Volume24H                float64              `json:"volume"`
	Turnover24H              float64              `json:"turnover"`
	LastPrice                int64                `json:"lastPrice"`
	PriceChangePercentage24H float64              `json:"priceChgPct"`
	SnapshotTime             convert.ExchangeTime `json:"ts"`
}

// WsFuturesTradeOrder represents trade order information according to the market.
type WsFuturesTradeOrder struct {
	OrderID          string               `json:"orderId"`
	Symbol           string               `json:"symbol"`
	Type             string               `json:"type"`       // Message Type: "open", "match", "filled", "canceled", "update"
	Status           string               `json:"status"`     // Order Status: "match", "open", "done"
	MatchSize        string               `json:"matchSize"`  // Match Size (when the type is "match")
	MatchPrice       string               `json:"matchPrice"` // Match Price (when the type is "match")
	OrderType        string               `json:"orderType"`  // Order Type, "market" indicates market order, "limit" indicates limit order
	Side             string               `json:"side"`       // Trading direction,include buy and sell
	OrderPrice       float64              `json:"price,string"`
	OrderSize        float64              `json:"size,string"`
	RemainSize       float64              `json:"remainSize,string"`
	FilledSize       float64              `json:"filledSize,string"`   // Remaining Size for Trading
	CanceledSize     float64              `json:"canceledSize,string"` // In the update message, the Size of order reduced
	TradeID          string               `json:"tradeId"`             // Trade ID (when the type is "match")
	ClientOid        string               `json:"clientOid"`           // Client supplied order id.
	OrderTime        convert.ExchangeTime `json:"orderTime"`
	OldSize          string               `json:"oldSize "`  // Size Before Update (when the type is "update")
	TradingDirection string               `json:"liquidity"` // Liquidity, Trading direction, buy or sell in taker
	Timestamp        convert.ExchangeTime `json:"ts"`
}

// WsStopOrderLifecycleEvent represents futures stop order lifecycle event.
type WsStopOrderLifecycleEvent struct {
	OrderID        string               `json:"orderId"`
	Symbol         string               `json:"symbol"`
	Type           string               `json:"type"`
	OrderType      string               `json:"orderType"`
	Side           string               `json:"side"`
	Size           float64              `json:"size,string"`
	OrderPrice     float64              `json:"orderPrice,string"`
	Stop           string               `json:"stop"`
	StopPrice      float64              `json:"stopPrice,string"`
	StopPriceType  string               `json:"stopPriceType"`
	TriggerSuccess bool                 `json:"triggerSuccess"`
	Error          string               `json:"error"`
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
	Timestamp      convert.ExchangeTime `json:"ts"`
}

// WsFuturesOrderMarginEvent represents an order margin account balance event.
type WsFuturesOrderMarginEvent struct {
	OrderMargin float64              `json:"orderMargin"`
	Currency    string               `json:"currency"`
	Timestamp   convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesAvailableBalance represents an available balance push data for futures account.
type WsFuturesAvailableBalance struct {
	AvailableBalance float64              `json:"availableBalance"`
	HoldBalance      float64              `json:"holdBalance"`
	Currency         string               `json:"currency"`
	Timestamp        convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesWithdrawalAmountAndTransferOutAmountEvent represents Withdrawal Amount & Transfer-Out Amount Event push data.
type WsFuturesWithdrawalAmountAndTransferOutAmountEvent struct {
	WithdrawHold float64              `json:"withdrawHold"` // Current frozen amount for withdrawal
	Currency     string               `json:"currency"`
	Timestamp    convert.ExchangeTime `json:"timestamp"`
}

// WsFuturesPosition represents futures account position change event.
type WsFuturesPosition struct {
	RealisedGrossPnl  float64              `json:"realisedGrossPnl"` // Accumulated realised profit and loss
	Symbol            string               `json:"symbol"`
	CrossMode         bool                 `json:"crossMode"`        // Cross mode or not
	LiquidationPrice  float64              `json:"liquidationPrice"` // Liquidation price
	PosLoss           float64              `json:"posLoss"`          // Manually added margin amount
	AvgEntryPrice     float64              `json:"avgEntryPrice"`    // Average entry price
	UnrealisedPnl     float64              `json:"unrealisedPnl"`    // Unrealised profit and loss
	MarkPrice         float64              `json:"markPrice"`        // Mark price
	PosMargin         float64              `json:"posMargin"`        // Position margin
	AutoDeposit       bool                 `json:"autoDeposit"`      // Auto deposit margin or not
	RiskLimit         float64              `json:"riskLimit"`
	UnrealisedCost    float64              `json:"unrealisedCost"`    // Unrealised value
	PosComm           float64              `json:"posComm"`           // Bankruptcy cost
	PosMaint          float64              `json:"posMaint"`          // Maintenance margin
	PosCost           float64              `json:"posCost"`           // Position value
	MaintMarginReq    float64              `json:"maintMarginReq"`    // Maintenance margin rate
	BankruptPrice     float64              `json:"bankruptPrice"`     // Bankruptcy price
	RealisedCost      float64              `json:"realisedCost"`      // Currently accumulated realised position value
	MarkValue         float64              `json:"markValue"`         // Mark value
	PosInit           float64              `json:"posInit"`           // Position margin
	RealisedPnl       float64              `json:"realisedPnl"`       // Realised profit and loss
	MaintMargin       float64              `json:"maintMargin"`       // Position margin
	RealLeverage      float64              `json:"realLeverage"`      // Leverage of the order
	ChangeReason      string               `json:"changeReason"`      // changeReason:marginChange、positionChange、liquidation、autoAppendMarginStatusChange、adl
	CurrentCost       float64              `json:"currentCost"`       // Current position value
	OpeningTimestamp  convert.ExchangeTime `json:"openingTimestamp"`  // Open time
	CurrentQty        float64              `json:"currentQty"`        // Current position
	DelevPercentage   float64              `json:"delevPercentage"`   // ADL ranking percentile
	CurrentComm       float64              `json:"currentComm"`       // Current commission
	RealisedGrossCost float64              `json:"realisedGrossCost"` // Accumulated realised gross profit value
	IsOpen            bool                 `json:"isOpen"`            // Opened position or not
	PosCross          float64              `json:"posCross"`          // Manually added margin
	CurrentTimestamp  convert.ExchangeTime `json:"currentTimestamp"`  // Current timestamp
	UnrealisedRoePcnt float64              `json:"unrealisedRoePcnt"` // Rate of return on investment
	UnrealisedPnlPcnt float64              `json:"unrealisedPnlPcnt"` // Position profit and loss ratio
	SettleCurrency    string               `json:"settleCurrency"`    // Currency used to clear and settle the trades
}

// WsFuturesMarkPricePositionChanges represents futures account position change caused by mark price.
type WsFuturesMarkPricePositionChanges struct {
	MarkPrice         float64              `json:"markPrice"`         // Mark price
	MarkValue         float64              `json:"markValue"`         // Mark value
	MaintMargin       float64              `json:"maintMargin"`       // Position margin
	RealLeverage      float64              `json:"realLeverage"`      // Leverage of the order
	UnrealisedPnl     float64              `json:"unrealisedPnl"`     // Unrealised profit and lost
	UnrealisedRoePcnt float64              `json:"unrealisedRoePcnt"` // Rate of return on investment
	UnrealisedPnlPcnt float64              `json:"unrealisedPnlPcnt"` // Position profit and loss ratio
	DelevPercentage   float64              `json:"delevPercentage"`   // ADL ranking percentile
	CurrentTimestamp  convert.ExchangeTime `json:"currentTimestamp"`  // Current timestamp
	SettleCurrency    string               `json:"settleCurrency"`    // Currency used to clear and settle the trades
}

// WsFuturesPositionFundingSettlement represents futures account position funding settlement push data.
type WsFuturesPositionFundingSettlement struct {
	PositionSize     float64              `json:"qty"`
	MarkPrice        float64              `json:"markPrice"`
	FundingRate      float64              `json:"fundingRate"`
	FundingFee       float64              `json:"fundingFee"`
	FundingTime      convert.ExchangeTime `json:"fundingTime"`
	CurrentTimestamp convert.ExchangeTime `json:"ts"`
	SettleCurrency   string               `json:"settleCurrency"`
}

// IsolatedMarginBorrowing represents response data for initiating isolated margin borrowing.
type IsolatedMarginBorrowing struct {
	OrderID    string  `json:"orderId"`
	Currency   string  `json:"currency"`
	ActualSize float64 `json:"actualSize,string"`
}

// Response represents response model and implements UnmarshalTo interface.
type Response struct {
	Data interface{} `json:"data"`
	Error
}

// CancelOrderResponse represents cancel order response model.
type CancelOrderResponse struct {
	CancelledOrderID string `json:"cancelledOrderId"`
	ClientOID        string `json:"clientOid"`
	Error
}

// AccountLedgerResponse represents the account ledger response detailed information
type AccountLedgerResponse struct {
	CurrentPage int64        `json:"currentPage"`
	PageSize    int64        `json:"pageSize"`
	TotalNum    int64        `json:"totalNum"`
	TotalPage   int64        `json:"totalPage"`
	Items       []LedgerInfo `json:"items"`
}

// SpotAPISubAccountParams parameters for Spot APIs for sub-accounts
type SpotAPISubAccountParams struct {
	SubAccountName string `json:"subName"`
	Passphrase     string `json:"passphrase"`
	Remark         string `json:"remark"`
	Permission     string `json:"permission,omitempty"`    // Permissions(Only General、Spot、Futures、Margin permissions can be set, such as "General, Trade". The default is "General")
	IPWhitelist    string `json:"ipWhitelist,omitempty"`   // IP whitelist(You may add up to 20 IPs. Use a halfwidth comma to each IP)
	Expire         int64  `json:"expire,string,omitempty"` // API expiration time; Never expire(default)-1，30Day30，90Day90，180Day180，360Day360

	// used when modifying sub-account API key
	APIKey string `json:"apiKey"`
}

// SubAccountResponse represents the sub-user detail.
type SubAccountResponse struct {
	CurrentPage int64        `json:"currentPage"`
	PageSize    int64        `json:"pageSize"`
	TotalNum    int64        `json:"totalNum"`
	TotalPage   int64        `json:"totalPage"`
	Items       []SubAccount `json:"items"`
}

// SubAccount represents sub-user
type SubAccount struct {
	UserID    string               `json:"userId"`
	SubName   string               `json:"subName"`
	Type      int64                `json:"type"` //type:1-rebot  or type:0-nomal
	Remarks   string               `json:"remarks"`
	UID       int64                `json:"uid"`
	Status    int64                `json:"status"`
	Access    string               `json:"access"`
	CreatedAt convert.ExchangeTime `json:"createdAt"`
}

// SubAccountV2Response represents a sub-account detailed response of V2
type SubAccountV2Response struct {
	Code string `json:"code"`
	Data struct {
		CurrentPage int64        `json:"currentPage"`
		PageSize    int64        `json:"pageSize"`
		TotalNum    int64        `json:"totalNum"`
		TotalPage   int64        `json:"totalPage"`
		Items       []SubAccount `json:"items"`
	} `json:"data"`
}

// SubAccountCreatedResponse represents the sub-account response.
type SubAccountCreatedResponse struct {
	UID     int64  `json:"uid"`
	SubName string `json:"subName"`
	Remarks string `json:"remarks"`
	Access  string `json:"access"`
}

// SpotAPISubAccount represents a Spot APIs for sub-accounts.
type SpotAPISubAccount struct {
	APIKey      string `json:"apiKey"`
	IPWhitelist string `json:"ipWhitelist"`
	Permission  string `json:"permission"`
	SubName     string `json:"subName"`

	Remark    string               `json:"remark"`
	CreatedAt convert.ExchangeTime `json:"createdAt"`

	// TODO: to be removed if not used any other place
	APISecret  string `json:"apiSecret"`
	Passphrase string `json:"passphrase"`
}

// DeleteSubAccountResponse represents delete sub-account response.
type DeleteSubAccountResponse struct {
	SubAccountName string `json:"subName"`
	APIKey         string `json:"apiKey"`
}

// ConnectionMessage represents a connection and subscription status message.
type ConnectionMessage struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// TickersResponse represents list of tickers and update timestamp information.
type TickersResponse struct {
	Time    convert.ExchangeTime `json:"time"`
	Tickers []TickerInfo         `json:"ticker"`
}

// FundingInterestRateResponse represents a funding interest rate list response information.
type FundingInterestRateResponse struct {
	List    []FuturesInterestRate `json:"dataList"`
	HasMore bool                  `json:"hasMore"`
}

// FuturesIndexResponse represents a response data for futures indexes.
type FuturesIndexResponse struct {
	List    []FuturesIndex `json:"dataList"`
	HasMore bool           `json:"hasMore"`
}

// FuturesInterestRateResponse represents a futures interest rate list response.
type FuturesInterestRateResponse struct {
	List    []FuturesInterestRate `json:"dataList"`
	HasMore bool                  `json:"hasMore"`
}

// FuturesTransactionHistoryResponse represents a futures transaction history response.
type FuturesTransactionHistoryResponse struct {
	List    []FuturesTransactionHistory `json:"dataList"`
	HasMore bool                        `json:"hasMore"`
}

// FuturesFundingHistoryResponse represents funding history response for futures account.
type FuturesFundingHistoryResponse struct {
	DataList []FuturesFundingHistory `json:"dataList"`
	HasMore  bool                    `json:"hasMore"`
}

// FuturesOrderParam represents a query parameter for placing future oorder
type FuturesOrderParam struct {
	ClientOrderID string        `json:"clientOid"`
	Side          string        `json:"side"`
	Symbol        currency.Pair `json:"symbol,omitempty"`
	OrderType     string        `json:"type"`
	Remark        string        `json:"remark,omitempty"`
	Stop          string        `json:"stp,omitempty"`           // [optional] Either `down` or `up`. Requires stopPrice and stopPriceType to be defined
	StopPriceType string        `json:"stopPriceType,omitempty"` // [optional] Either TP, IP or MP, Need to be defined if stop is specified. `TP` for trade price, `MP` for Mark price, and "IP" for index price.
	TimeInForce   string        `json:"timeInForce,omitempty"`
	Size          float64       `json:"size,omitempty,string"`
	Price         float64       `json:"price,string,omitempty"`
	StopPrice     float64       `json:"stopPrice,omitempty,string"`
	Leverage      float64       `json:"leverage,omitempty,string"`
	VisibleSize   float64       `json:"visibleSize,omitempty,string"`
	ReduceOnly    bool          `json:"reduceOnly,omitempty"`
	CloseOrder    bool          `json:"closeOrder,omitempty"`
	ForceHold     bool          `json:"forceHold,omitempty"`
	PostOnly      bool          `json:"postOnly,omitempty"`
	Hidden        bool          `json:"hidden,omitempty"`
	Iceberg       bool          `json:"iceberg,omitempty"`
}

// SpotOrderParam represents the spot place order request parameters.
type SpotOrderParam struct {
	ClientOrderID       string        `json:"clientOid"`
	Side                string        `json:"side"`
	Symbol              currency.Pair `json:"symbol"`
	OrderType           string        `json:"type,omitempty"`
	TradeType           string        `json:"tradeType,omitempty"` // [Optional] The type of trading : TRADE（Spot Trade）, MARGIN_TRADE (Margin Trade). Default is TRADE.
	Remark              string        `json:"remark,omitempty"`
	SelfTradePrevention string        `json:"stp,omitempty"`         // [Optional] self trade prevention , CN, CO, CB or DC. `CN` for Cancel newest, `DC` for Decrease and Cancel, `CO` for cancel oldest, and `CB` for Cancel both
	TimeInForce         string        `json:"timeInForce,omitempty"` // [Optional] GTC, GTT, IOC, or FOK (default is GTC)
	PostOnly            bool          `json:"postOnly,omitempty"`
	Hidden              bool          `json:"hidden,omitempty"`
	Iceberg             bool          `json:"iceberg,omitempty"`
	ReduceOnly          bool          `json:"reduceOnly,omitempty"`
	CancelAfter         int64         `json:"cancelAfter,omitempty"`
	Size                float64       `json:"size,omitempty,string"`
	Price               float64       `json:"price,string,omitempty"`
	VisibleSize         float64       `json:"visibleSize,omitempty,string"`
	Funds               float64       `json:"funds,string,omitempty"`
}

// MarginOrderParam represents the margin place order request parameters.
type MarginOrderParam struct {
	ClientOrderID       string        `json:"clientOid"`
	Side                string        `json:"side"`
	Symbol              currency.Pair `json:"symbol"`
	OrderType           string        `json:"type,omitempty"`
	Remark              string        `json:"remark,omitempty"`
	SelfTradePrevention string        `json:"stp,omitempty"`         // [Optional] self trade prevention , CN, CO, CB or DC. `CN` for Cancel newest, `DC` for Decrease and Cancel, `CO` for cancel oldest, and `CB` for Cancel both
	MarginModel         string        `json:"marginModel,omitempty"` // [Optional] The type of trading, including cross (cross mode) and isolated (isolated mode). It is set at cross by default.
	AutoBorrow          bool          `json:"autoBorrow,omitempty"`  // [Optional] Auto-borrow to place order. The system will first borrow you funds at the optimal interest rate and then place an order for you. Currently autoBorrow parameter only supports cross mode, not isolated mode. When add this param, stop profit and stop loss are not supported
	AutoRepay           bool          `json:"autoRepay,omitempty"`
	Price               float64       `json:"price,string,omitempty"`
	Size                float64       `json:"size,omitempty,string"`
	TimeInForce         string        `json:"timeInForce,omitempty"` // [Optional] GTC, GTT, IOC, or FOK (default is GTC)
	CancelAfter         int64         `json:"cancelAfter,omitempty"` // [Optional] cancel after n seconds, requires timeInForce to be GTT
	PostOnly            bool          `json:"postOnly,omitempty"`
	Hidden              bool          `json:"hidden,omitempty"`
	Iceberg             bool          `json:"iceberg,omitempty"`
	VisibleSize         float64       `json:"visibleSize,omitempty,string"`
	Funds               float64       `json:"funds,string,omitempty"`
}

// UniversalTransferParam represents a universal transfer parameter.
type UniversalTransferParam struct {
	ClientSuppliedOrderID string        `json:"clientOid"`
	Currency              currency.Code `json:"currency"`
	Amount                float64       `json:"amount"`
	FromUserID            string        `json:"fromUserId"`
	FromAccountType       string        `json:"fromAccountType"` // Account type：MAIN、TRADE、CONTRACT、MARGIN、ISOLATED、TRADE_HF、MARGIN_V2、ISOLATED_V2
	FromAccountTag        string        `json:"fromAccountTag"`  // Symbol, required when the account type is ISOLATED or ISOLATED_V2, for example: BTC-USDT
	TransferType          string        `json:"type"`            // Transfer Type: Transfer type：INTERNAL、PARENT_TO_SUB，SUB_TO_PARENT
	ToUserID              string        `json:"toUserId"`
	ToAccountType         string        `json:"toAccountType"`
	ToAccountTag          string        `json:"toAccountTag"`
}

// OCOOrderParams represents a an OCO order creation parameter
type OCOOrderParams struct {
	Symbol        currency.Pair `json:"symbol"`
	Side          string        `json:"side"`
	Price         float64       `json:"price,string"`
	Size          float64       `json:"size,string"`
	StopPrice     float64       `json:"stopPrice,string"`
	LimitPrice    float64       `json:"limitPrice,string"`
	TradeType     string        `json:"tradeType,omitempty"`
	ClientOrderID string        `json:"clientOid"`
	Remark        string        `json:"remark,omitempty"`
}

// OCOOrderCancellationResponse represents an order cancellation response.
type OCOOrderCancellationResponse struct {
	CancelledOrderIds []string `json:"cancelledOrderIds"`
}

// OCOOrderInfo represents an order info
type OCOOrderInfo struct {
	OrderID       string               `json:"orderId"`
	Symbol        string               `json:"symbol"`
	ClientOrderID string               `json:"clientOid"`
	OrderTime     convert.ExchangeTime `json:"orderTime"`
	Status        string               `json:"status"`
}

// OCOOrderDetail represents an oco order detail via the order ID.
type OCOOrderDetail struct {
	OrderID   string               `json:"orderId"`
	Symbol    string               `json:"symbol"`
	ClientOid string               `json:"clientOid"`
	OrderTime convert.ExchangeTime `json:"orderTime"`
	Status    string               `json:"status"`
	Orders    []struct {
		ID        string       `json:"id"`
		Symbol    string       `json:"symbol"`
		Side      string       `json:"side"`
		Price     types.Number `json:"price"`
		StopPrice types.Number `json:"stopPrice"`
		Size      types.Number `json:"size"`
		Status    string       `json:"status"`
	} `json:"orders"`
}

// OCOOrders represents an OCO orders list
type OCOOrders struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalNum    int `json:"totalNum"`
	TotalPage   int `json:"totalPage"`
	Items       []struct {
		OrderID   string               `json:"orderId"`
		Symbol    string               `json:"symbol"`
		ClientOid string               `json:"clientOid"`
		OrderTime convert.ExchangeTime `json:"orderTime"`
		Status    string               `json:"status"`
	} `json:"items"`
}

// PlaceMarginHFOrderParam represents a margin HF order parameters.
type PlaceMarginHFOrderParam struct {
	ClientOrderID       string        `json:"clientOid"`
	Side                string        `json:"side"`
	Symbol              currency.Pair `json:"symbol"`
	OrderType           string        `json:"type,omitempty"`
	SelfTradePrevention string        `json:"stp,omitempty"`
	IsIsolated          bool          `json:"isIsolated,omitempty"`
	AutoBorrow          bool          `json:"autoBorrow,omitempty"`
	AutoRepay           bool          `json:"autoRepay,omitempty"`
	Price               float64       `json:"price,string"`
	Size                float64       `json:"size,string"`
	TimeInForce         string        `json:"timeInForce,omitempty,string"`
	CancelAfter         int64         `json:"cancelAfter,omitempty,string"`
	PostOnly            bool          `json:"postOnly,omitempty,string"`
	Hidden              bool          `json:"hidden,omitempty,string"`
	Iceberg             bool          `json:"iceberg,omitempty,string"`
	VisibleSize         float64       `json:"visibleSize,omitempty,string"`
	Funds               string        `json:"funds,omitempty"`
}

// MarginHFOrderResponse  represents a margin HF order creation response.
type MarginHFOrderResponse struct {
	OrderID     string  `json:"orderId"`
	BorrowSize  float64 `json:"borrowSize"`
	LoanApplyID string  `json:"loanApplyId"`
}

// HFMarginOrderDetail represents a high-frequency margin order detail
type HFMarginOrderDetail struct {
	ID                  string               `json:"id"`
	Symbol              string               `json:"symbol"`
	OpType              string               `json:"opType"`
	Type                string               `json:"type"`
	Side                string               `json:"side"`
	Price               types.Number         `json:"price"`
	Size                types.Number         `json:"size"`
	Funds               string               `json:"funds"`
	DealFunds           string               `json:"dealFunds"`
	DealSize            types.Number         `json:"dealSize"`
	Fee                 types.Number         `json:"fee"`
	FeeCurrency         string               `json:"feeCurrency"`
	SelfTradePrevention string               `json:"stp"`
	TimeInForce         string               `json:"timeInForce"`
	PostOnly            bool                 `json:"postOnly"`
	Hidden              bool                 `json:"hidden"`
	Iceberg             bool                 `json:"iceberg"`
	VisibleSize         types.Number         `json:"visibleSize"`
	CancelAfter         int64                `json:"cancelAfter"`
	Channel             string               `json:"channel"`
	ClientOrderID       string               `json:"clientOid"`
	Remark              string               `json:"remark"`
	Tags                string               `json:"tags"`
	Active              bool                 `json:"active"`
	InOrderBook         bool                 `json:"inOrderBook"`
	CancelExist         bool                 `json:"cancelExist"`
	CreatedAt           convert.ExchangeTime `json:"createdAt"`
	LastUpdatedAt       convert.ExchangeTime `json:"lastUpdatedAt"`
	TradeType           string               `json:"tradeType"`
}

// FilledMarginHFOrdersResponse represents a filled HF margin orders
type FilledMarginHFOrdersResponse struct {
	LastID int64                 `json:"lastId"`
	Items  []HFMarginOrderDetail `json:"items"`
}

// HFMarginOrderTrade represents a HF margin order trade item.
type HFMarginOrderTrade struct {
	ID             int64                `json:"id"`
	Symbol         string               `json:"symbol"`
	TradeID        int64                `json:"tradeId"`
	OrderID        string               `json:"orderId"`
	CounterOrderID string               `json:"counterOrderId"`
	Side           string               `json:"side"`
	Liquidity      string               `json:"liquidity"`
	ForceTaker     bool                 `json:"forceTaker"`
	Price          types.Number         `json:"price"`
	Size           types.Number         `json:"size"`
	Funds          types.Number         `json:"funds"`
	Fee            types.Number         `json:"fee"`
	FeeRate        types.Number         `json:"feeRate"`
	FeeCurrency    string               `json:"feeCurrency"`
	Stop           string               `json:"stop"`
	TradeType      string               `json:"tradeType"`
	Type           string               `json:"type"`
	CreatedAt      convert.ExchangeTime `json:"createdAt"`
}

// HFMarginOrderTransaction represents a HF margin order transaction detail.
type HFMarginOrderTransaction struct {
	Items  []HFMarginOrderTrade `json:"items"`
	LastID int64                `json:"lastId"`
}
