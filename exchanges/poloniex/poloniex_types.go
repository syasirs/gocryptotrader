package poloniex

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/types"
)

// DepositAddressesResponse holds the full address per crypto-currency
type DepositAddressesResponse map[string]string

// WithdrawalFees the large list of predefined withdrawal fees
// Prone to change, using highest value
var WithdrawalFees = map[currency.Code]float64{
	currency.ZRX:   5,
	currency.ARDR:  2,
	currency.REP:   0.1,
	currency.BTC:   0.0005,
	currency.BCH:   0.0001,
	currency.XBC:   0.0001,
	currency.BTCD:  0.01,
	currency.BTM:   0.01,
	currency.BTS:   5,
	currency.BURST: 1,
	currency.BCN:   1,
	currency.CVC:   1,
	currency.CLAM:  0.001,
	currency.XCP:   1,
	currency.DASH:  0.01,
	currency.DCR:   0.1,
	currency.DGB:   0.1,
	currency.DOGE:  5,
	currency.EMC2:  0.01,
	currency.EOS:   0,
	currency.ETH:   0.01,
	currency.ETC:   0.01,
	currency.EXP:   0.01,
	currency.FCT:   0.01,
	currency.GAME:  0.01,
	currency.GAS:   0,
	currency.GNO:   0.015,
	currency.GNT:   1,
	currency.GRC:   0.01,
	currency.HUC:   0.01,
	currency.LBC:   0.05,
	currency.LSK:   0.1,
	currency.LTC:   0.001,
	currency.MAID:  10,
	currency.XMR:   0.015,
	currency.NMC:   0.01,
	currency.NAV:   0.01,
	currency.XEM:   15,
	currency.NEOS:  0.0001,
	currency.NXT:   1,
	currency.OMG:   0.3,
	currency.OMNI:  0.1,
	currency.PASC:  0.01,
	currency.PPC:   0.01,
	currency.POT:   0.01,
	currency.XPM:   0.01,
	currency.XRP:   0.15,
	currency.SC:    10,
	currency.STEEM: 0.01,
	currency.SBD:   0.01,
	currency.XLM:   0.00001,
	currency.STORJ: 1,
	currency.STRAT: 0.01,
	currency.AMP:   5,
	currency.SYS:   0.01,
	currency.USDT:  10,
	currency.VRC:   0.01,
	currency.VTC:   0.001,
	currency.VIA:   0.01,
	currency.ZEC:   0.001,
}

// SymbolDetail represents a symbol(currency pair) detailed information.
type SymbolDetail struct {
	Symbol            string               `json:"symbol"`
	BaseCurrencyName  string               `json:"baseCurrencyName"`
	QuoteCurrencyName string               `json:"quoteCurrencyName"`
	DisplayName       string               `json:"displayName"`
	State             string               `json:"state"`
	VisibleStartTime  convert.ExchangeTime `json:"visibleStartTime"`
	TradableStartTime convert.ExchangeTime `json:"tradableStartTime"`
	SymbolTradeLimit  struct {
		Symbol        string       `json:"symbol"`
		PriceScale    float64      `json:"priceScale"`
		QuantityScale float64      `json:"quantityScale"`
		AmountScale   float64      `json:"amountScale"`
		MinQuantity   types.Number `json:"minQuantity"`
		MinAmount     types.Number `json:"minAmount"`
		HighestBid    types.Number `json:"highestBid"`
		LowestAsk     types.Number `json:"lowestAsk"`
	} `json:"symbolTradeLimit"`
	CrossMargin struct {
		SupportCrossMargin bool    `json:"supportCrossMargin"`
		MaxLeverage        float64 `json:"maxLeverage"`
	} `json:"crossMargin"`
}

// CurrencyDetail represents all supported currencies.
type CurrencyDetail map[string]struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Type                  string   `json:"type"`
	WithdrawalFee         string   `json:"withdrawalFee"`
	MinConf               int64    `json:"minConf"`
	DepositAddress        string   `json:"depositAddress"`
	Blockchain            string   `json:"blockchain"`
	Delisted              bool     `json:"delisted"`
	TradingState          string   `json:"tradingState"`
	WalletState           string   `json:"walletState"`
	WalletDepositState    string   `json:"walletDepositState"`
	WalletWithdrawalState string   `json:"walletWithdrawalState"`
	ParentChain           string   `json:"parentChain"`
	IsMultiChain          bool     `json:"isMultiChain"`
	IsChildChain          bool     `json:"isChildChain"`
	SupportCollateral     bool     `json:"supportCollateral"`
	SupportBorrow         bool     `json:"supportBorrow"`
	ChildChains           []string `json:"childChains"`
}

// CurrencyV2Information represents all supported currencies
type CurrencyV2Information struct {
	ID          int64  `json:"id"`
	Coin        string `json:"coin"`
	Delisted    bool   `json:"delisted"`
	TradeEnable bool   `json:"tradeEnable"`
	Name        string `json:"name"`
	NetworkList []struct {
		ID               int64        `json:"id"`
		Coin             string       `json:"coin"`
		Name             string       `json:"name"`
		CurrencyType     string       `json:"currencyType"`
		Blockchain       string       `json:"blockchain"`
		WithdrawalEnable bool         `json:"withdrawalEnable"`
		DepositEnable    bool         `json:"depositEnable"`
		DepositAddress   string       `json:"depositAddress"`
		Decimals         float64      `json:"decimals"`
		MinConfirm       float64      `json:"minConfirm"`
		WithdrawMin      types.Number `json:"withdrawMin"`
		WithdrawFee      types.Number `json:"withdrawFee"`
	} `json:"networkList"`
	SupportCollateral bool `json:"supportCollateral,omitempty"`
	SupportBorrow     bool `json:"supportBorrow,omitempty"`
}

// ServerSystemTime represents a server time.
type ServerSystemTime struct {
	ServerTime convert.ExchangeTime `json:"serverTime"`
}

// MarketPrice represents ticker information.
type MarketPrice struct {
	Symbol        string               `json:"symbol"`
	DailyChange   types.Number         `json:"dailyChange"`
	Price         types.Number         `json:"price"`
	Timestamp     convert.ExchangeTime `json:"time"`
	PushTimestamp convert.ExchangeTime `json:"ts"`
}

// MarkPrice represents latest mark price for all cross margin symbols.
type MarkPrice struct {
	Symbol          string               `json:"symbol"`
	MarkPrice       types.Number         `json:"markPrice"`
	RecordTimestamp convert.ExchangeTime `json:"time"`
}

// MarkPriceComponent represents a mark price component instance.
type MarkPriceComponent struct {
	Symbol     string               `json:"symbol"`
	Timestamp  convert.ExchangeTime `json:"ts"`
	MarkPrice  types.Number         `json:"markPrice"`
	Components []struct {
		Symbol       string       `json:"symbol"`
		Exchange     string       `json:"exchange"`
		SymbolPrice  types.Number `json:"symbolPrice"`
		Weight       types.Number `json:"weight"`
		ConvertPrice types.Number `json:"convertPrice"`
	} `json:"components"`
}

// OrderbookData represents an order book data for a specific symbol.
type OrderbookData struct {
	CreationTime  convert.ExchangeTime `json:"time"`
	Scale         types.Number         `json:"scale"`
	Asks          []types.Number       `json:"asks"`
	Bids          []types.Number       `json:"bids"`
	PushTimestamp convert.ExchangeTime `json:"ts"`
}

// CandlestickArrayData symbol at given timeframe (interval).
type CandlestickArrayData [14]interface{}

// CandlestickData represents a candlestick data for a specific symbol.
type CandlestickData struct {
	Low              float64
	High             float64
	Open             float64
	Close            float64
	Amount           float64
	Quantity         float64
	BuyTakeAmount    float64
	BuyTakerQuantity float64
	TradeCount       float64
	PushTimestamp    time.Time
	WeightedAverage  float64
	Interval         kline.Interval
	StartTime        time.Time
	EndTime          time.Time
}

func processCandlestickData(candlestickData []CandlestickArrayData) ([]CandlestickData, error) {
	candles := make([]CandlestickData, len(candlestickData))
	for i := range candlestickData {
		candle, err := getCandlestickData(&candlestickData[i])
		if err != nil {
			return nil, err
		}
		candles[i] = *candle
	}
	return candles, nil
}

func getCandlestickData(candlestickData *CandlestickArrayData) (*CandlestickData, error) {
	var err error
	candle := &CandlestickData{}
	candle.Low, err = strconv.ParseFloat(candlestickData[0].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.High, err = strconv.ParseFloat(candlestickData[1].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.Open, err = strconv.ParseFloat(candlestickData[2].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.Close, err = strconv.ParseFloat(candlestickData[3].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.Amount, err = strconv.ParseFloat(candlestickData[4].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.Quantity, err = strconv.ParseFloat(candlestickData[5].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.BuyTakeAmount, err = strconv.ParseFloat(candlestickData[6].(string), 64)
	if err != nil {
		return nil, err
	}
	buyTakerQuantity, okay := candlestickData[7].(string)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.BuyTakerQuantity, err = strconv.ParseFloat(buyTakerQuantity, 64)
	if err != nil {
		return nil, err
	}
	candle.TradeCount, okay = candlestickData[8].(float64)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	puchTime, okay := candlestickData[9].(float64)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.PushTimestamp = time.UnixMilli(int64(puchTime))
	weightedAverage, okay := candlestickData[10].(string)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.WeightedAverage, err = strconv.ParseFloat(weightedAverage, 64)
	if err != nil {
		return nil, err
	}
	intervalString, okay := candlestickData[11].(string)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.Interval, err = stringToInterval(intervalString)
	if err != nil {
		return nil, err
	}
	timestamp, okay := candlestickData[12].(float64)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.StartTime = time.UnixMilli(int64(timestamp))
	timestamp, okay = candlestickData[13].(float64)
	if !okay {
		return nil, errUnexpectedIncomingDataType
	}
	candle.EndTime = time.UnixMilli(int64(timestamp))
	return candle, nil
}

// Trade represents a trade instance.
type Trade struct {
	ID         string               `json:"id"`
	Price      types.Number         `json:"price"`
	Quantity   types.Number         `json:"quantity"`
	Amount     types.Number         `json:"amount"`
	TakerSide  string               `json:"takerSide"`
	Timestamp  convert.ExchangeTime `json:"ts"`
	CreateTime convert.ExchangeTime `json:"createTime"`
}

// TickerData represents a price ticker information.
type TickerData struct {
	Symbol      string               `json:"symbol"`
	Open        types.Number         `json:"open"`
	Low         types.Number         `json:"low"`
	High        types.Number         `json:"high"`
	Close       types.Number         `json:"close"`
	Quantity    types.Number         `json:"quantity"`
	Amount      types.Number         `json:"amount"`
	TradeCount  int64                `json:"tradeCount"`
	StartTime   convert.ExchangeTime `json:"startTime"`
	CloseTime   convert.ExchangeTime `json:"closeTime"`
	DisplayName string               `json:"displayName"`
	DailyChange string               `json:"dailyChange"`
	Bid         types.Number         `json:"bid"`
	BidQuantity types.Number         `json:"bidQuantity"`
	Ask         types.Number         `json:"ask"`
	AskQuantity types.Number         `json:"askQuantity"`
	Timestamp   convert.ExchangeTime `json:"ts"`
	MarkPrice   types.Number         `json:"markPrice"`
}

// CollateralInfo represents collateral information.
type CollateralInfo struct {
	Currency              string       `json:"currency"`
	CollateralRate        types.Number `json:"collateralRate"`
	InitialMarginRate     types.Number `json:"initialMarginRate"`
	MaintenanceMarginRate types.Number `json:"maintenanceMarginRate"`
}

// BorrowRateInfo represents borrow rates information
type BorrowRateInfo struct {
	Tier  string `json:"tier"`
	Rates []struct {
		Currency         string `json:"currency"`
		DailyBorrowRate  string `json:"dailyBorrowRate"`
		HourlyBorrowRate string `json:"hourlyBorrowRate"`
		BorrowLimit      string `json:"borrowLimit"`
	} `json:"rates"`
}

// AccountInformation represents a user account information.
type AccountInformation struct {
	AccountID    string `json:"accountId"`
	AccountType  string `json:"accountType"`
	AccountState string `json:"accountState"`
}

// AccountBalance represents each account’s id, type and balances (assets).
type AccountBalance struct {
	AccountID   string `json:"accountId"`
	AccountType string `json:"accountType"`
	Balances    []struct {
		CurrencyID string       `json:"currencyId"`
		Currency   string       `json:"currency"`
		Available  types.Number `json:"available"`
		Hold       types.Number `json:"hold"`
	} `json:"balances"`
}

// AccountActivity represents activities such as airdrop, rebates, staking,
// credit/debit adjustments, and other (historical adjustments).
type AccountActivity struct {
	ID           string               `json:"id"`
	Currency     string               `json:"currency"`
	Amount       types.Number         `json:"amount"`
	State        string               `json:"state"`
	CreateTime   convert.ExchangeTime `json:"createTime"`
	Description  string               `json:"description"`
	ActivityType int64                `json:"activityType"`
}

// AccountTransferParams request parameter for account fund transfer.
type AccountTransferParams struct {
	Ccy         currency.Code `json:"currency"`
	Amount      float64       `json:"amount,string"`
	FromAccount string        `json:"fromAccount"`
	ToAccount   string        `json:"toAccount"`
}

// AccountTransferResponse represents an account transfer response.
type AccountTransferResponse struct {
	TransferID string `json:"transferId"`
}

type errorResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// AccountTransferRecord represents an account transfer record.
type AccountTransferRecord struct {
	ID          string               `json:"id"`
	FromAccount string               `json:"fromAccount"`
	ToAccount   string               `json:"toAccount"`
	Currency    string               `json:"currency"`
	State       string               `json:"state"`
	CreateTime  convert.ExchangeTime `json:"createTime"`
	Amount      types.Number         `json:"amount"`
}

// FeeInfo represents an account transfer information.
type FeeInfo struct {
	TransactionDiscount bool         `json:"trxDiscount"`
	MakerRate           types.Number `json:"makerRate"`
	TakerRate           types.Number `json:"takerRate"`
	Volume30D           types.Number `json:"volume30D"`
	SpecialFeeRates     []struct {
		Symbol    string       `json:"symbol"`
		MakerRate types.Number `json:"makerRate"`
		TakerRate types.Number `json:"takerRate"`
	} `json:"specialFeeRates"`
}

// InterestHistory represents an interest history.
type InterestHistory struct {
	ID                  string               `json:"id"`
	CurrencyName        string               `json:"currencyName"`
	Principal           types.Number         `json:"principal"`
	Interest            types.Number         `json:"interest"`
	InterestRate        types.Number         `json:"interestRate"`
	InterestAccuredTime convert.ExchangeTime `json:"interestAccuredTime"`
}

// SubAccount represents a users account.
type SubAccount struct {
	AccountID    string `json:"accountId"`
	AccountName  string `json:"accountName"`
	AccountState string `json:"accountState"`
	IsPrimary    string `json:"isPrimary"`
}

// SubAccountBalance represents a users account balance.
type SubAccountBalance struct {
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName"`
	AccountType string `json:"accountType"`
	IsPrimary   string `json:"isPrimary"`
	Balances    []struct {
		Currency              string       `json:"currency"`
		Available             types.Number `json:"available"`
		Hold                  types.Number `json:"hold"`
		MaxAvailable          types.Number `json:"maxAvailable"`
		AccountEquity         string       `json:"accountEquity,omitempty"`
		UnrealisedPNL         string       `json:"unrealisedPNL,omitempty"`
		MarginBalance         string       `json:"marginBalance,omitempty"`
		PositionMargin        string       `json:"positionMargin,omitempty"`
		OrderMargin           string       `json:"orderMargin,omitempty"`
		FrozenFunds           string       `json:"frozenFunds,omitempty"`
		AvailableBalance      types.Number `json:"availableBalance,omitempty"`
		RealizedProfitAndLoss string       `json:"pnl,omitempty"`
	} `json:"balances"`
}

// SubAccountTransferParam represents a sub-account transfer request parameters.
type SubAccountTransferParam struct {
	Currency        currency.Code `json:"currency"`
	Amount          float64       `json:"amount,string"`
	FromAccountID   string        `json:"fromAccountId"`
	FromAccountType string        `json:"fromAccountType"`
	ToAccountID     string        `json:"toAccountId"`
	ToAccountType   string        `json:"toAccountType"`
}

// SubAccountTransfer represents a sub-account transfer record.
type SubAccountTransfer struct {
	ID              string               `json:"id"`
	FromAccountID   string               `json:"fromAccountId"`
	FromAccountName string               `json:"fromAccountName"`
	FromAccountType string               `json:"fromAccountType"`
	ToAccountID     string               `json:"toAccountId"`
	ToAccountName   string               `json:"toAccountName"`
	ToAccountType   string               `json:"toAccountType"`
	Currency        string               `json:"currency"`
	Amount          types.Number         `json:"amount"`
	State           string               `json:"state"`
	CreateTime      convert.ExchangeTime `json:"createTime"`
}

// WalletActivityResponse holds wallet activity info
type WalletActivityResponse struct {
	Deposits    []WalletDeposits    `json:"deposits"`
	Withdrawals []WalletWithdrawals `json:"withdrawals"`
}

// WalletDeposits holds wallet deposit info
type WalletDeposits struct {
	DepositNumber int64                `json:"depositNumber"`
	Currency      string               `json:"currency"`
	Address       string               `json:"address"`
	Amount        types.Number         `json:"amount"`
	Confirmations int64                `json:"confirmations"`
	TransactionID string               `json:"txid"`
	Timestamp     convert.ExchangeTime `json:"timestamp"`
	Status        string               `json:"status"`
}

// WalletWithdrawals holds wallet withdrawal info
type WalletWithdrawals struct {
	WithdrawalRequestsID int64                `json:"withdrawalRequestsId"`
	Currency             string               `json:"currency"`
	Address              string               `json:"address"`
	Amount               types.Number         `json:"amount"`
	Fee                  types.Number         `json:"fee"`
	Timestamp            convert.ExchangeTime `json:"timestamp"`
	Status               string               `json:"status"`
	TransactionID        string               `json:"txid"`
	IPAddress            string               `json:"ipAddress"`
	PaymentID            string               `json:"paymentID"`
}

// Withdraw holds withdraw information
type Withdraw struct {
	WithdrawRequestID int64 `json:"withdrawalRequestsId"`
}

// WithdrawCurrencyParam represents a currency withdrawal parameter.
type WithdrawCurrencyParam struct {
	Currency    currency.Code `json:"currency"`
	Amount      float64       `json:"amount,string"`
	Address     string        `json:"address"`
	PaymentID   string        `json:"paymentId,omitempty"`
	AllowBorrow bool          `json:"allowBorrow,omitempty"`
}

// WithdrawCurrencyV2Param represents a V2 currency withdrawal parameter.
type WithdrawCurrencyV2Param struct {
	Coin        currency.Code `json:"coin"`
	Network     string        `json:"network"`
	Amount      float64       `json:"amount,string"`
	Address     string        `json:"address"`
	AddressTag  string        `json:"addressTag,omitempty"`
	AllowBorrow bool          `json:"allowBorrow,omitempty"`
}

// AccountMargin represents an account margin response
type AccountMargin struct {
	TotalAccountValue types.Number         `json:"totalAccountValue"`
	TotalMargin       types.Number         `json:"totalMargin"`
	UsedMargin        types.Number         `json:"usedMargin"`
	FreeMargin        types.Number         `json:"freeMargin"`
	MaintenanceMargin types.Number         `json:"maintenanceMargin"`
	CreationTime      convert.ExchangeTime `json:"time"`
	MarginRatio       string               `json:"marginRatio"`
}

// BorroweStatus represents currency borrow status.
type BorroweStatus struct {
	Currency         string       `json:"currency"`
	Available        types.Number `json:"available"`
	Borrowed         types.Number `json:"borrowed"`
	Hold             types.Number `json:"hold"`
	MaxAvailable     types.Number `json:"maxAvailable"`
	HourlyBorrowRate types.Number `json:"hourlyBorrowRate"`
	Version          string       `json:"version"`
}

// MaxBuySellAmount represents a maximum buy and sell amount.
type MaxBuySellAmount struct {
	Symbol           string       `json:"symbol"`
	MaxLeverage      int64        `json:"maxLeverage"`
	AvailableBuy     types.Number `json:"availableBuy"`
	MaxAvailableBuy  types.Number `json:"maxAvailableBuy"`
	AvailableSell    types.Number `json:"availableSell"`
	MaxAvailableSell types.Number `json:"maxAvailableSell"`
}

// PlaceOrderParams represents place order parameters.
type PlaceOrderParams struct {
	Symbol      currency.Pair `json:"symbol"`
	Side        string        `json:"side"`
	Type        string        `json:"type,omitempty"`
	AccountType string        `json:"accountType,omitempty"`

	// Quantity Base units for the order. Quantity is required for MARKET SELL or any LIMIT orders
	Quantity float64 `json:"quantity,omitempty,string"`

	// Amount Quote units for the order. Amount is required for MARKET BUY order
	Amount float64 `json:"amount,omitempty,string"`

	// Price is required for non-market orders
	Price float64 `json:"price,omitempty,string"`

	TimeInForce   string `json:"timeInForce,omitempty"` // GTC, IOC, FOK (Default: GTC)
	ClientOrderID string `json:"clientOrderId,omitempty"`

	AllowBorrow bool   `json:"allowBorrow,omitempty"`
	STPMode     string `json:"stpMode,omitempty"` // self-trade prevention. Defaults to EXPIRE_TAKER. None: enable self-trade; EXPIRE_TAKER: Taker order will be canceled when self-trade happens

	SlippageTolerance string `json:"slippageTolerance,omitempty"` // Used to control the maximum slippage ratio, the value range is greater than 0 and less than 1
}

// PlaceOrderResponse represents a response structure for placing order.
type PlaceOrderResponse struct {
	ID            string `json:"id"`
	ClientOrderID string `json:"clientOrderId"`

	// Websocket response status code and message.
	Message string `json:"message,omitempty"`
	Code    int64  `json:"code,omitempty"`
}

// PlaceBatchOrderRespItem represents a single batch order response item.
type PlaceBatchOrderRespItem struct {
	ID            string `json:"id,omitempty"`
	ClientOrderID string `json:"clientOrderId"`
	Code          int64  `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
}

// CancelReplaceOrderParam represents a cancellation and order replacement request parameter.
type CancelReplaceOrderParam struct {
	orderID           string  // orderID: used in order path parameter.
	ClientOrderID     string  `json:"clientOrderId"`
	Price             float64 `json:"price,omitempty,string"`
	Quantity          float64 `json:"quantity,omitempty,string"`
	Amount            float64 `json:"amount,omitempty,string"`
	AmendedType       string  `json:"type,omitempty,string"`
	TimeInForce       string  `json:"timeInForce"`
	AllowBorrow       bool    `json:"allowBorrow"`
	ProceedOnFailure  bool    `json:"proceedOnFailure,omitempty,string"`
	SlippageTolerance float64 `json:"slippageTolerance,omitempty,string"`
}

// CancelReplaceOrderResponse represents a response parameter for order cancellation and replacement operation.
type CancelReplaceOrderResponse struct {
	ID            string               `json:"id"`
	ClientOrderID string               `json:"clientOrderId"`
	Price         convert.ExchangeTime `json:"price"`
	Quantity      convert.ExchangeTime `json:"quantity"`
	Code          int64                `json:"code"`
	Message       string               `json:"message"`
}

// TradeOrder represents a trade order instance.
type TradeOrder struct {
	ID             string               `json:"id"`
	ClientOrderID  string               `json:"clientOrderId"`
	Symbol         string               `json:"symbol"`
	State          string               `json:"state"`
	AccountType    string               `json:"accountType"`
	Side           string               `json:"side"`
	Type           string               `json:"type"`
	TimeInForce    string               `json:"timeInForce"`
	Quantity       types.Number         `json:"quantity"`
	Price          types.Number         `json:"price"`
	AvgPrice       types.Number         `json:"avgPrice"`
	Amount         types.Number         `json:"amount"`
	FilledQuantity types.Number         `json:"filledQuantity"`
	FilledAmount   types.Number         `json:"filledAmount"`
	CreateTime     convert.ExchangeTime `json:"createTime"`
	UpdateTime     convert.ExchangeTime `json:"updateTime"`
	OrderSource    string               `json:"orderSource"`
	Loan           bool                 `json:"loan"`
	CancelReason   int64                `json:"cancelReason"`
}

// SmartOrderItem represents a smart order detail.
type SmartOrderItem struct {
	ID            string               `json:"id"`
	ClientOrderID string               `json:"clientOrderId"`
	Symbol        string               `json:"symbol"`
	State         string               `json:"state"`
	AccountType   string               `json:"accountType"`
	Side          string               `json:"side"`
	Type          string               `json:"type"`
	TimeInForce   string               `json:"timeInForce"`
	Quantity      types.Number         `json:"quantity"`
	Price         types.Number         `json:"price"`
	Amount        types.Number         `json:"amount"`
	StopPrice     types.Number         `json:"stopPrice"`
	CreateTime    convert.ExchangeTime `json:"createTime"`
	UpdateTime    convert.ExchangeTime `json:"updateTime"`
}

// CancelOrderResponse represents a cancel order response instance.
type CancelOrderResponse struct {
	OrderID       string `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	State         string `json:"state"`
	Code          int64  `json:"code"`
	Message       string `json:"message"`
}

// WsCancelOrderResponse represents a websocket cancel orders instance.
type WsCancelOrderResponse struct {
	OrderID int64 `json:"orderId"`
	CancelOrderResponse
}

// OrderCancellationParams represents an order cancellation parameters.
type OrderCancellationParams struct {
	OrderIds       []string `json:"orderIds"`
	ClientOrderIds []string `json:"clientOrderIds"`
}

// KillSwitchStatus represents a kill switch response
type KillSwitchStatus struct {
	StartTime        convert.ExchangeTime `json:"startTime"`
	CancellationTime convert.ExchangeTime `json:"cancellationTime"`
}

// SmartOrderRequestParam represents a smart trade order parameters
type SmartOrderRequestParam struct {
	Symbol        currency.Pair `json:"symbol"`
	Side          string        `json:"side"`
	TimeInForce   string        `json:"timeInForce,omitempty"`
	Type          string        `json:"type,omitempty"`
	AccountType   string        `json:"accountType,omitempty"`
	Price         float64       `json:"price,omitempty,string"`
	StopPrice     float64       `json:"stopPrice,omitempty,string"`
	Quantity      float64       `json:"quantity,omitempty,string"`
	Amount        float64       `json:"amount,omitempty,string"`
	ClientOrderID string        `json:"clientOrderId,omitempty"`
}

// CancelReplaceSmartOrderParam represents a cancellation and order replacement request parameter for smart orders.
type CancelReplaceSmartOrderParam struct {
	orderID          string  // orderID: will be used in request path.
	ClientOrderID    string  `json:"clientOrderId"`
	Price            float64 `json:"price,omitempty,string"`
	StopPrice        float64 `json:"stopPrice,omitempty,string"`
	Quantity         float64 `json:"quantity,omitempty,string"`
	Amount           float64 `json:"amount,omitempty,string"`
	AmendedType      string  `json:"type,omitempty,string"`
	TimeInForce      string  `json:"timeInForce,omitempty"`
	ProceedOnFailure bool    `json:"proceedOnFailure,omitempty,string"`
}

// CancelReplaceSmartOrderResponse represents a response parameter for order cancellation and replacement operation.
type CancelReplaceSmartOrderResponse struct {
	ID        string               `json:"id"`
	StopPrice types.Number         `json:"stopPrice"`
	Price     convert.ExchangeTime `json:"price"`
	Quantity  convert.ExchangeTime `json:"quantity"`
	Code      int64                `json:"code"`
	Message   string               `json:"message"`
}

// SmartOrderDetail represents a smart order information and trigger detailed information.
type SmartOrderDetail struct {
	ID             string               `json:"id"`
	ClientOrderID  string               `json:"clientOrderId"`
	Symbol         string               `json:"symbol"`
	State          string               `json:"state"`
	AccountType    string               `json:"accountType"`
	Side           string               `json:"side"`
	Type           string               `json:"type"`
	TimeInForce    string               `json:"timeInForce"`
	Quantity       types.Number         `json:"quantity"`
	Price          types.Number         `json:"price"`
	Amount         types.Number         `json:"amount"`
	StopPrice      types.Number         `json:"stopPrice"`
	CreateTime     convert.ExchangeTime `json:"createTime"`
	UpdateTime     convert.ExchangeTime `json:"updateTime"`
	TriggeredOrder struct {
		ID             string               `json:"id"`
		ClientOrderID  string               `json:"clientOrderId"`
		Symbol         string               `json:"symbol"`
		State          string               `json:"state"`
		AccountType    string               `json:"accountType"`
		Side           string               `json:"side"`
		Type           string               `json:"type"`
		TimeInForce    string               `json:"timeInForce"`
		Quantity       types.Number         `json:"quantity"`
		Price          types.Number         `json:"price"`
		AvgPrice       types.Number         `json:"avgPrice"`
		Amount         types.Number         `json:"amount"`
		FilledQuantity types.Number         `json:"filledQuantity"`
		FilledAmount   types.Number         `json:"filledAmount"`
		CreateTime     convert.ExchangeTime `json:"createTime"`
		UpdateTime     convert.ExchangeTime `json:"updateTime"`
	} `json:"triggeredOrder"`
}

// CancelSmartOrderResponse represents a close cancel smart order response instance.
type CancelSmartOrderResponse struct {
	OrderID       string `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	State         string `json:"state"`
	Code          int64  `json:"code"`
	Message       string `json:"message"`
}

// TradeHistoryItem represents an order trade history instance.
type TradeHistoryItem struct {
	ID            string               `json:"id"`
	ClientOrderID string               `json:"clientOrderId"`
	Symbol        string               `json:"symbol"`
	AccountType   string               `json:"accountType"`
	OrderID       string               `json:"orderId"`
	Side          string               `json:"side"`
	Type          string               `json:"type"`
	MatchRole     string               `json:"matchRole"`
	Price         types.Number         `json:"price"`
	Quantity      types.Number         `json:"quantity"`
	Amount        types.Number         `json:"amount"`
	FeeCurrency   string               `json:"feeCurrency"`
	FeeAmount     types.Number         `json:"feeAmount"`
	PageID        string               `json:"pageId"`
	CreateTime    convert.ExchangeTime `json:"createTime"`
}

// SubscriptionPayload represents a subscriptions request instance structure.
type SubscriptionPayload struct {
	Event      string   `json:"event"`
	Channel    []string `json:"channel"`
	Symbols    []string `json:"symbols,omitempty"`
	Currencies []string `json:"currencies,omitempty"`
	Depth      int64    `json:"depth,omitempty"`
}

// SubscriptionResponse represents a subscription response instance.
type SubscriptionResponse struct {
	ID string `json:"id"`

	Event   string          `json:"event"`
	Channel string          `json:"channel"`
	Action  string          `json:"action"`
	Data    json.RawMessage `json:"data"`

	Currencies []string `json:"currencies"`
	Symbols    []string `json:"symbols"`
}

// GetWsResponse returns a *WsResponse instance from *SubscriptionResponse
func (a *SubscriptionResponse) GetWsResponse() *WsResponse {
	return &WsResponse{
		Event:   a.Event,
		Channel: a.Channel,
		Action:  a.Action,
		Data:    a.Data,
	}
}

// WsResponse represents a websocket push data instance.
type WsResponse struct {
	Event   string      `json:"event"`
	Channel string      `json:"channel"`
	Action  string      `json:"action"`
	Data    interface{} `json:"data"`
}

// WsSymbol represents a subscription
type WsSymbol struct {
	Symbol            string               `json:"symbol"`
	BaseCurrencyName  string               `json:"baseCurrencyName"`
	QuoteCurrencyName string               `json:"quoteCurrencyName"`
	DisplayName       string               `json:"displayName"`
	State             string               `json:"state"`
	VisibleStartTime  convert.ExchangeTime `json:"visibleStartTime"`
	TradableStartTime convert.ExchangeTime `json:"tradableStartTime"`
	CrossMargin       struct {
		SupportCrossMargin bool   `json:"supportCrossMargin"`
		MaxLeverage        string `json:"maxLeverage"`
	} `json:"crossMargin"`
	SymbolTradeLimit struct {
		Symbol        string       `json:"symbol"`
		PriceScale    float64      `json:"priceScale"`
		QuantityScale float64      `json:"quantityScale"`
		AmountScale   float64      `json:"amountScale"`
		MinQuantity   types.Number `json:"minQuantity"`
		MinAmount     types.Number `json:"minAmount"`
		HighestBid    types.Number `json:"highestBid"`
		LowestAsk     types.Number `json:"lowestAsk"`
	} `json:"symbolTradeLimit"`
}

// WsCurrency represents a currency instance from websocket stream.
type WsCurrency struct {
	Currency          string       `json:"currency"`
	ID                int64        `json:"id"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Type              string       `json:"type"`
	WithdrawalFee     types.Number `json:"withdrawalFee"`
	MinConf           int64        `json:"minConf"`
	DepositAddress    any          `json:"depositAddress"`
	Blockchain        string       `json:"blockchain"`
	Delisted          bool         `json:"delisted"`
	TradingState      string       `json:"tradingState"`
	WalletState       string       `json:"walletState"`
	ParentChain       any          `json:"parentChain"`
	IsMultiChain      bool         `json:"isMultiChain"`
	IsChildChain      bool         `json:"isChildChain"`
	SupportCollateral bool         `json:"supportCollateral"`
	SupportBorrow     bool         `json:"supportBorrow"`
	ChildChains       []string     `json:"childChains"`
}

// WsExchangeStatus represents websocket exchange status.
// the values for MM and POM are ON and OFF
type WsExchangeStatus []struct {
	MaintenanceMode string `json:"MM"`
	PostOnlyMode    string `json:"POM"`
}

// WsCandles represents a candlestick data instance.
type WsCandles struct {
	Symbol     string               `json:"symbol"`
	Open       types.Number         `json:"open"`
	High       types.Number         `json:"high"`
	Low        types.Number         `json:"low"`
	Close      types.Number         `json:"close"`
	Quantity   types.Number         `json:"quantity"`
	Amount     types.Number         `json:"amount"`
	TradeCount int64                `json:"tradeCount"`
	StartTime  convert.ExchangeTime `json:"startTime"`
	CloseTime  convert.ExchangeTime `json:"closeTime"`
	Timestamp  convert.ExchangeTime `json:"ts"`
}

// WsTrade represents websocket trade data
type WsTrade struct {
	ID         string               `json:"id"`
	Symbol     string               `json:"symbol"`
	Amount     types.Number         `json:"amount"`
	Quantity   types.Number         `json:"quantity"`
	TakerSide  string               `json:"takerSide"`
	Price      types.Number         `json:"price"`
	CreateTime convert.ExchangeTime `json:"createTime"`
	Timestamp  convert.ExchangeTime `json:"ts"`
}

// WsTicker represents a websocket ticker information.
type WsTicker struct {
	TradeCount  int64                `json:"tradeCount"`
	Symbol      string               `json:"symbol"`
	StartTime   convert.ExchangeTime `json:"startTime"`
	Open        types.Number         `json:"open"`
	High        types.Number         `json:"high"`
	Low         types.Number         `json:"low"`
	Close       types.Number         `json:"close"`
	Quantity    types.Number         `json:"quantity"`
	Amount      types.Number         `json:"amount"`
	DailyChange types.Number         `json:"dailyChange"`
	MarkPrice   types.Number         `json:"markPrice"`
	CloseTime   convert.ExchangeTime `json:"closeTime"`
	Timestamp   convert.ExchangeTime `json:"ts"`
}

// WsBook represents an orderbook.
type WsBook struct {
	Symbol     string               `json:"symbol"`
	Asks       [][]types.Number     `json:"asks"`
	Bids       [][]types.Number     `json:"bids"`
	ID         int64                `json:"id"`
	Timestamp  convert.ExchangeTime `json:"ts"`
	CreateTime convert.ExchangeTime `json:"createTime"`

	LastID int64 `json:"lastId"`
}

// AuthParams represents websocket authenticaten parameters
type AuthParams struct {
	Key              string `json:"key"`
	SignTimestamp    int64  `json:"signTimestamp"`
	SignatureMethod  string `json:"signatureMethod,omitempty"`
	SignatureVersion string `json:"signatureVersion,omitempty"`
	Signature        string `json:"signature"`
}

// WebsocketAuthenticationResponse represents websocket authentication response.
type WebsocketAuthenticationResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	Timestamp convert.ExchangeTime `json:"ts"`
}

// WebsocketTradeOrder represents a websocket trade order.
type WebsocketTradeOrder struct {
	Symbol         string               `json:"symbol"`
	Type           string               `json:"type"`
	Quantity       types.Number         `json:"quantity"`
	OrderID        string               `json:"orderId"`
	TradeFee       types.Number         `json:"tradeFee"`
	ClientOrderID  string               `json:"clientOrderId"`
	AccountType    string               `json:"accountType"`
	FeeCurrency    string               `json:"feeCurrency"`
	EventType      string               `json:"eventType"`
	Source         string               `json:"source"`
	Side           string               `json:"side"`
	FilledQuantity types.Number         `json:"filledQuantity"`
	FilledAmount   types.Number         `json:"filledAmount"`
	MatchRole      string               `json:"matchRole"`
	State          string               `json:"state"`
	TradeTime      convert.ExchangeTime `json:"tradeTime"`
	TradeAmount    types.Number         `json:"tradeAmount"`
	OrderAmount    types.Number         `json:"orderAmount"`
	CreateTime     convert.ExchangeTime `json:"createTime"`
	Price          types.Number         `json:"price"`
	TradeQty       types.Number         `json:"tradeQty"`
	TradePrice     types.Number         `json:"tradePrice"`
	TradeID        string               `json:"tradeId"`
	Timestamp      convert.ExchangeTime `json:"ts"`
}

// WsTradeBalance represents a balance information through the websocket channel
type WsTradeBalance []struct {
	ID          int64                `json:"id"`
	UserID      int64                `json:"userId"`
	ChangeTime  convert.ExchangeTime `json:"changeTime"`
	AccountID   string               `json:"accountId"`
	AccountType string               `json:"accountType"`
	EventType   string               `json:"eventType"`
	Available   types.Number         `json:"available"`
	Currency    string               `json:"currency"`
	Hold        types.Number         `json:"hold"`
	Timestamp   convert.ExchangeTime `json:"ts"`
}

// WebsocketResponse represents a websocket responses.
type WebsocketResponse struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}
