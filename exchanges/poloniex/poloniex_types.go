package poloniex

import "github.com/shopspring/decimal"

// Ticker holds ticker data
type Ticker struct {
	Last          decimal.Decimal `json:"last,string"`
	LowestAsk     decimal.Decimal `json:"lowestAsk,string"`
	HighestBid    decimal.Decimal `json:"highestBid,string"`
	PercentChange decimal.Decimal `json:"percentChange,string"`
	BaseVolume    decimal.Decimal `json:"baseVolume,string"`
	QuoteVolume   decimal.Decimal `json:"quoteVolume,string"`
	IsFrozen      int             `json:"isFrozen,string"`
	High24Hr      decimal.Decimal `json:"high24hr,string"`
	Low24Hr       decimal.Decimal `json:"low24hr,string"`
}

// OrderbookResponseAll holds the full response type orderbook
type OrderbookResponseAll struct {
	Data map[string]OrderbookResponse
}

// CompleteBalances holds the full balance data
type CompleteBalances struct {
	Currency map[string]CompleteBalance
}

// OrderbookResponse is a sub-type for orderbooks
type OrderbookResponse struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen string          `json:"isFrozen"`
	Error    string          `json:"error"`
}

// OrderbookItem holds data on an individual item
type OrderbookItem struct {
	Price  decimal.Decimal
	Amount decimal.Decimal
}

// OrderbookAll contains the full range of orderbooks
type OrderbookAll struct {
	Data map[string]Orderbook
}

// Orderbook is a generic type golding orderbook information
type Orderbook struct {
	Asks []OrderbookItem `json:"asks"`
	Bids []OrderbookItem `json:"bids"`
}

// TradeHistory holds trade history data
type TradeHistory struct {
	GlobalTradeID int64           `json:"globalTradeID"`
	TradeID       int64           `json:"tradeID"`
	Date          string          `json:"date"`
	Type          string          `json:"type"`
	Rate          decimal.Decimal `json:"rate,string"`
	Amount        decimal.Decimal `json:"amount,string"`
	Total         decimal.Decimal `json:"total,string"`
}

// ChartData holds kline data
type ChartData struct {
	Date            int             `json:"date"`
	High            decimal.Decimal `json:"high"`
	Low             decimal.Decimal `json:"low"`
	Open            decimal.Decimal `json:"open"`
	Close           decimal.Decimal `json:"close"`
	Volume          decimal.Decimal `json:"volume"`
	QuoteVolume     decimal.Decimal `json:"quoteVolume"`
	WeightedAverage decimal.Decimal `json:"weightedAverage"`
	Error           string          `json:"error"`
}

// Currencies contains currency information
type Currencies struct {
	Name               string          `json:"name"`
	MaxDailyWithdrawal string          `json:"maxDailyWithdrawal"`
	TxFee              decimal.Decimal `json:"txFee,string"`
	MinConfirmations   int             `json:"minConf"`
	DepositAddresses   interface{}     `json:"depositAddress"`
	Disabled           int             `json:"disabled"`
	Delisted           int             `json:"delisted"`
	Frozen             int             `json:"frozen"`
}

// LoanOrder holds loan order information
type LoanOrder struct {
	Rate     decimal.Decimal `json:"rate,string"`
	Amount   decimal.Decimal `json:"amount,string"`
	RangeMin int             `json:"rangeMin"`
	RangeMax int             `json:"rangeMax"`
}

// LoanOrders holds loan order information range
type LoanOrders struct {
	Offers  []LoanOrder `json:"offers"`
	Demands []LoanOrder `json:"demands"`
}

// Balance holds data for a range of currencies
type Balance struct {
	Currency map[string]decimal.Decimal
}

// CompleteBalance contains the complete balance with a btcvalue
type CompleteBalance struct {
	Available decimal.Decimal
	OnOrders  decimal.Decimal
	BTCValue  decimal.Decimal
}

// DepositAddresses holds the full address per crypto-currency
type DepositAddresses struct {
	Addresses map[string]string
}

// DepositsWithdrawals holds withdrawal information
type DepositsWithdrawals struct {
	Deposits []struct {
		Currency      string          `json:"currency"`
		Address       string          `json:"address"`
		Amount        decimal.Decimal `json:"amount,string"`
		Confirmations int             `json:"confirmations"`
		TransactionID string          `json:"txid"`
		Timestamp     int64           `json:"timestamp"`
		Status        string          `json:"status"`
	} `json:"deposits"`
	Withdrawals []struct {
		WithdrawalNumber int64           `json:"withdrawalNumber"`
		Currency         string          `json:"currency"`
		Address          string          `json:"address"`
		Amount           decimal.Decimal `json:"amount,string"`
		Confirmations    int             `json:"confirmations"`
		TransactionID    string          `json:"txid"`
		Timestamp        int64           `json:"timestamp"`
		Status           string          `json:"status"`
		IPAddress        string          `json:"ipAddress"`
	} `json:"withdrawals"`
}

// Order hold order information
type Order struct {
	OrderNumber int64           `json:"orderNumber,string"`
	Type        string          `json:"type"`
	Rate        decimal.Decimal `json:"rate,string"`
	Amount      decimal.Decimal `json:"amount,string"`
	Total       decimal.Decimal `json:"total,string"`
	Date        string          `json:"date"`
	Margin      decimal.Decimal `json:"margin"`
}

// OpenOrdersResponseAll holds all open order responses
type OpenOrdersResponseAll struct {
	Data map[string][]Order
}

// OpenOrdersResponse holds open response orders
type OpenOrdersResponse struct {
	Data []Order
}

// AuthentictedTradeHistory holds client trade history information
type AuthentictedTradeHistory struct {
	GlobalTradeID int64           `json:"globalTradeID"`
	TradeID       int64           `json:"tradeID,string"`
	Date          string          `json:"date"`
	Rate          decimal.Decimal `json:"rate,string"`
	Amount        decimal.Decimal `json:"amount,string"`
	Total         decimal.Decimal `json:"total,string"`
	Fee           decimal.Decimal `json:"fee,string"`
	OrderNumber   int64           `json:"orderNumber,string"`
	Type          string          `json:"type"`
	Category      string          `json:"category"`
}

// AuthenticatedTradeHistoryAll holds the full client trade history
type AuthenticatedTradeHistoryAll struct {
	Data map[string][]AuthentictedTradeHistory
}

// AuthenticatedTradeHistoryResponse is a response type for trade history
type AuthenticatedTradeHistoryResponse struct {
	Data []AuthentictedTradeHistory
}

// ResultingTrades holds resultant trade information
type ResultingTrades struct {
	Amount  decimal.Decimal `json:"amount,string"`
	Date    string          `json:"date"`
	Rate    decimal.Decimal `json:"rate,string"`
	Total   decimal.Decimal `json:"total,string"`
	TradeID int64           `json:"tradeID,string"`
	Type    string          `json:"type"`
}

// OrderResponse is a response type of trades
type OrderResponse struct {
	OrderNumber int64             `json:"orderNumber,string"`
	Trades      []ResultingTrades `json:"resultingTrades"`
}

// GenericResponse is a response type for exchange generic responses
type GenericResponse struct {
	Success int    `json:"success"`
	Error   string `json:"error"`
}

// MoveOrderResponse is a response type for move order trades
type MoveOrderResponse struct {
	Success     int                          `json:"success"`
	Error       string                       `json:"error"`
	OrderNumber int64                        `json:"orderNumber,string"`
	Trades      map[string][]ResultingTrades `json:"resultingTrades"`
}

// Withdraw holds withdraw information
type Withdraw struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Fee holds fees for specific trades
type Fee struct {
	MakerFee        decimal.Decimal `json:"makerFee,string"`
	TakerFee        decimal.Decimal `json:"takerFee,string"`
	ThirtyDayVolume decimal.Decimal `json:"thirtyDayVolume,string"`
	NextTier        decimal.Decimal `json:"nextTier,string"`
}

// Margin holds margin information
type Margin struct {
	TotalValue    decimal.Decimal `json:"totalValue,string"`
	ProfitLoss    decimal.Decimal `json:"pl,string"`
	LendingFees   decimal.Decimal `json:"lendingFees,string"`
	NetValue      decimal.Decimal `json:"netValue,string"`
	BorrowedValue decimal.Decimal `json:"totalBorrowedValue,string"`
	CurrentMargin decimal.Decimal `json:"currentMargin,string"`
}

// MarginPosition holds margin positional information
type MarginPosition struct {
	Amount            decimal.Decimal `json:"amount,string"`
	Total             decimal.Decimal `json:"total,string"`
	BasePrice         decimal.Decimal `json:"basePrice,string"`
	LiquidiationPrice decimal.Decimal `json:"liquidiationPrice"`
	ProfitLoss        decimal.Decimal `json:"pl,string"`
	LendingFees       decimal.Decimal `json:"lendingFees,string"`
	Type              string          `json:"type"`
}

// LoanOffer holds loan offer information
type LoanOffer struct {
	ID        int64           `json:"id"`
	Rate      decimal.Decimal `json:"rate,string"`
	Amount    decimal.Decimal `json:"amount,string"`
	Duration  int             `json:"duration"`
	AutoRenew bool            `json:"autoRenew,int"`
	Date      string          `json:"date"`
}

// ActiveLoans shows the full active loans on the exchange
type ActiveLoans struct {
	Provided []LoanOffer `json:"provided"`
	Used     []LoanOffer `json:"used"`
}

// LendingHistory holds the full lending history data
type LendingHistory struct {
	ID       int64           `json:"id"`
	Currency string          `json:"currency"`
	Rate     decimal.Decimal `json:"rate,string"`
	Amount   decimal.Decimal `json:"amount,string"`
	Duration decimal.Decimal `json:"duration,string"`
	Interest decimal.Decimal `json:"interest,string"`
	Fee      decimal.Decimal `json:"fee,string"`
	Earned   decimal.Decimal `json:"earned,string"`
	Open     string          `json:"open"`
	Close    string          `json:"close"`
}

// WebsocketTicker holds ticker data for the websocket
type WebsocketTicker struct {
	CurrencyPair  string
	Last          decimal.Decimal
	LowestAsk     decimal.Decimal
	HighestBid    decimal.Decimal
	PercentChange decimal.Decimal
	BaseVolume    decimal.Decimal
	QuoteVolume   decimal.Decimal
	IsFrozen      bool
	High          decimal.Decimal
	Low           decimal.Decimal
}

// WebsocketTrollboxMessage holds trollbox messages and information for
// websocket
type WebsocketTrollboxMessage struct {
	MessageNumber decimal.Decimal
	Username      string
	Message       string
	Reputation    decimal.Decimal
}
