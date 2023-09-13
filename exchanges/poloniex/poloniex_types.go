package poloniex

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
)

// Ticker holds ticker data
type Ticker struct {
	ID            float64 `json:"id"`
	Last          float64 `json:"last,string"`
	LowestAsk     float64 `json:"lowestAsk,string"`
	HighestBid    float64 `json:"highestBid,string"`
	PercentChange float64 `json:"percentChange,string"`
	BaseVolume    float64 `json:"baseVolume,string"`
	QuoteVolume   float64 `json:"quoteVolume,string"`
	High24Hr      float64 `json:"high24hr,string"`
	Low24Hr       float64 `json:"low24hr,string"`
	IsFrozen      uint8   `json:"isFrozen,string"`
	PostOnly      uint8   `json:"postOnly,string"`
}

// OrderbookResponseAll holds the full response type orderbook
type OrderbookResponseAll struct {
	Data map[string]OrderbookResponse
}

// CompleteBalances holds the full balance data
type CompleteBalances map[string]CompleteBalance

// OrderbookResponse is a sub-type for orderbooks
type OrderbookResponse struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen string          `json:"isFrozen"`
	Error    string          `json:"error"`
	Seq      int64           `json:"seq"`
}

// OrderbookItem holds data on an individual item
type OrderbookItem struct {
	Price  float64
	Amount float64
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
	GlobalTradeID string  `json:"globalTradeID"`
	TradeID       string  `json:"tradeID"`
	Date          string  `json:"date"`
	Type          string  `json:"type"`
	Rate          float64 `json:"rate,string"`
	Amount        float64 `json:"amount,string"`
	Total         float64 `json:"total,string"`
}

// OrderStatus holds order status data
type OrderStatus struct {
	Result  json.RawMessage `json:"result"`
	Success int64           `json:"success"`
}

// OrderStatusData defines order status details
type OrderStatusData struct {
	Pair           string  `json:"currencyPair"`
	Rate           float64 `json:"rate,string"`
	Amount         float64 `json:"amount,string"`
	Total          float64 `json:"total,string"`
	StartingAmount float64 `json:"startingAmount,string"`
	Type           string  `json:"type"`
	Status         string  `json:"status"`
	Date           string  `json:"date"`
	Fee            float64 `json:"fee,string"`
}

// OrderTrade holds order trade data
type OrderTrade struct {
	Status        string  `json:"status"`
	GlobalTradeID string  `json:"globalTradeID"`
	TradeID       string  `json:"tradeID"`
	CurrencyPair  string  `json:"currencyPair"`
	Type          string  `json:"type"`
	Rate          float64 `json:"rate,string"`
	Amount        float64 `json:"amount,string"`
	Total         float64 `json:"total,string"`
	Fee           float64 `json:"fee,string"`
	Date          string  `json:"date"`
}

// ChartData holds kline data
type ChartData struct {
	Date            int64   `json:"date,string"`
	High            float64 `json:"high,string"`
	Low             float64 `json:"low,string"`
	Open            float64 `json:"open,string"`
	Close           float64 `json:"close,string"`
	Volume          float64 `json:"volume,string"`
	QuoteVolume     float64 `json:"quoteVolume,string"`
	WeightedAverage float64 `json:"weightedAverage,string"`
	Error           string  `json:"error"`
}

// Currencies contains currency information
type Currencies struct {
	ID                        float64  `json:"id"`
	Name                      string   `json:"name"`
	HumanType                 string   `json:"humanType"`
	CurrencyType              string   `json:"currencyType"`
	TxFee                     float64  `json:"txFee,string"`
	MinConfirmations          int64    `json:"minConf"`
	DepositAddress            string   `json:"depositAddress"`
	WithdrawalDepositDisabled uint8    `json:"disabled"`
	Frozen                    uint8    `json:"frozen"`
	HexColour                 string   `json:"hexColor"`
	Blockchain                string   `json:"blockchain"`
	Delisted                  uint8    `json:"delisted"`
	ParentChain               string   `json:"parentChain"`
	IsMultiChain              uint8    `json:"isMultiChain"`
	IsChildChain              uint8    `json:"isChildChain"`
	ChildChains               []string `json:"childChains"`
	IsGeofenced               uint8    `json:"isGeofenced"`
}

// LoanOrder holds loan order information
type LoanOrder struct {
	Rate     float64 `json:"rate,string"`
	Amount   float64 `json:"amount,string"`
	RangeMin int64   `json:"rangeMin"`
	RangeMax int64   `json:"rangeMax"`
}

// LoanOrders holds loan order information range
type LoanOrders struct {
	Offers  []LoanOrder `json:"offers"`
	Demands []LoanOrder `json:"demands"`
}

// Balance holds data for a range of currencies
type Balance struct {
	Currency map[string]float64
}

// CompleteBalance contains the complete balance with a btcvalue
type CompleteBalance struct {
	Available float64 `json:"available,string"`
	OnOrders  float64 `json:"onOrders,string"`
	BTCValue  float64 `json:"btcValue,string"`
}

// DepositAddressesResponse holds the full address per crypto-currency
type DepositAddressesResponse map[string]string

// DepositsWithdrawals holds withdrawal information
type DepositsWithdrawals struct {
	Deposits []struct {
		Currency      string  `json:"currency"`
		Address       string  `json:"address"`
		Amount        float64 `json:"amount,string"`
		Confirmations int64   `json:"confirmations"`
		TransactionID string  `json:"txid"`
		Timestamp     int64   `json:"timestamp"`
		Status        string  `json:"status"`
	} `json:"deposits"`
	Withdrawals []struct {
		WithdrawalNumber int64   `json:"withdrawalNumber"`
		Currency         string  `json:"currency"`
		Address          string  `json:"address"`
		Amount           float64 `json:"amount,string"`
		Confirmations    int64   `json:"confirmations"`
		TransactionID    string  `json:"txid"`
		Timestamp        int64   `json:"timestamp"`
		Status           string  `json:"status"`
		IPAddress        string  `json:"ipAddress"`
	} `json:"withdrawals"`
}

// Order hold order information
type Order struct {
	OrderNumber int64   `json:"orderNumber,string"`
	Type        string  `json:"type"`
	Rate        float64 `json:"rate,string"`
	Amount      float64 `json:"amount,string"`
	Total       float64 `json:"total,string"`
	Date        string  `json:"date"`
	Margin      float64 `json:"margin"`
}

// OpenOrdersResponseAll holds all open order responses
type OpenOrdersResponseAll struct {
	Data map[string][]Order
}

// OpenOrdersResponse holds open response orders
type OpenOrdersResponse struct {
	Data []Order
}

// AuthenticatedTradeHistory holds client trade history information
type AuthenticatedTradeHistory struct {
	GlobalTradeID string  `json:"globalTradeID"`
	TradeID       string  `json:"tradeID"`
	Date          string  `json:"date"`
	Rate          float64 `json:"rate,string"`
	Amount        float64 `json:"amount,string"`
	Total         float64 `json:"total,string"`
	Fee           float64 `json:"fee,string"`
	OrderNumber   int64   `json:"orderNumber,string"`
	Type          string  `json:"type"`
	Category      string  `json:"category"`
}

// AuthenticatedTradeHistoryAll holds the full client trade history
type AuthenticatedTradeHistoryAll struct {
	Data map[string][]AuthenticatedTradeHistory
}

// AuthenticatedTradeHistoryResponse is a response type for trade history
type AuthenticatedTradeHistoryResponse struct {
	Data []AuthenticatedTradeHistory
}

// ResultingTrades holds resultant trade information
type ResultingTrades struct {
	Amount  float64 `json:"amount,string"`
	Date    string  `json:"date"`
	Rate    float64 `json:"rate,string"`
	Total   float64 `json:"total,string"`
	TradeID int64   `json:"tradeID,string"`
	Type    string  `json:"type"`
}

// OrderResponse is a response type of trades
type OrderResponse struct {
	OrderNumber int64             `json:"orderNumber,string"`
	Trades      []ResultingTrades `json:"resultingTrades"`
}

// GenericResponse is a response type for exchange generic responses
type GenericResponse struct {
	Success int64  `json:"success"`
	Error   string `json:"error"`
}

// MoveOrderResponse is a response type for move order trades
type MoveOrderResponse struct {
	Success     int64                        `json:"success"`
	Error       string                       `json:"error"`
	OrderNumber int64                        `json:"orderNumber,string"`
	Trades      map[string][]ResultingTrades `json:"resultingTrades"`
}

// Fee holds fees for specific trades
type Fee struct {
	MakerFee        float64 `json:"makerFee,string"`
	TakerFee        float64 `json:"takerFee,string"`
	ThirtyDayVolume float64 `json:"thirtyDayVolume,string"`
}

// Margin holds margin information
type Margin struct {
	TotalValue    float64 `json:"totalValue,string"`
	ProfitLoss    float64 `json:"pl,string"`
	LendingFees   float64 `json:"lendingFees,string"`
	NetValue      float64 `json:"netValue,string"`
	BorrowedValue float64 `json:"totalBorrowedValue,string"`
	CurrentMargin float64 `json:"currentMargin,string"`
}

// MarginPosition holds margin positional information
type MarginPosition struct {
	Amount           float64 `json:"amount,string"`
	Total            float64 `json:"total,string"`
	BasePrice        float64 `json:"basePrice,string"`
	LiquidationPrice float64 `json:"liquidationPrice"`
	ProfitLoss       float64 `json:"pl,string"`
	LendingFees      float64 `json:"lendingFees,string"`
	Type             string  `json:"type"`
}

// LoanOffer holds loan offer information
type LoanOffer struct {
	ID        int64   `json:"id"`
	Rate      float64 `json:"rate,string"`
	Amount    float64 `json:"amount,string"`
	Duration  int64   `json:"duration"`
	AutoRenew bool    `json:"autoRenew"`
	Date      string  `json:"date"`
}

// ActiveLoans shows the full active loans on the exchange
type ActiveLoans struct {
	Provided []LoanOffer `json:"provided"`
	Used     []LoanOffer `json:"used"`
}

// LendingHistory holds the full lending history data
type LendingHistory struct {
	ID       int64   `json:"id"`
	Currency string  `json:"currency"`
	Rate     float64 `json:"rate,string"`
	Amount   float64 `json:"amount,string"`
	Duration float64 `json:"duration,string"`
	Interest float64 `json:"interest,string"`
	Fee      float64 `json:"fee,string"`
	Earned   float64 `json:"earned,string"`
	Open     string  `json:"open"`
	Close    string  `json:"close"`
}

// WebsocketTicker holds ticker data for the websocket
type WebsocketTicker struct {
	CurrencyPair  string
	Last          float64
	LowestAsk     float64
	HighestBid    float64
	PercentChange float64
	BaseVolume    float64
	QuoteVolume   float64
	IsFrozen      bool
	High          float64
	Low           float64
}

// WebsocketTrollboxMessage holds trollbox messages and information for
// websocket
type WebsocketTrollboxMessage struct {
	MessageNumber float64
	Username      string
	Message       string
	Reputation    float64
}

// WsCommand defines the request params after a websocket connection has been
// established
type WsCommand struct {
	Command string      `json:"command"`
	Channel interface{} `json:"channel"`
	APIKey  string      `json:"key,omitempty"`
	Payload string      `json:"payload,omitempty"`
	Sign    string      `json:"sign,omitempty"`
}

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

// WsOrderUpdateResponse Authenticated Ws Account data
type WsOrderUpdateResponse struct {
	OrderNumber float64
	NewAmount   string
}

// WsTradeNotificationResponse Authenticated Ws Account data
type WsTradeNotificationResponse struct {
	TradeID       float64
	Rate          float64
	Amount        float64
	FeeMultiplier float64
	FundingType   float64
	OrderNumber   float64
	TotalFee      float64
	Date          time.Time
}

// WsAuthorisationRequest Authenticated Ws Account data request
type WsAuthorisationRequest struct {
	Command string `json:"command"`
	Channel int64  `json:"channel"`
	Sign    string `json:"sign"`
	Key     string `json:"key"`
	Payload string `json:"payload"`
}

// CancelOrdersResponse holds cancelled order info
type CancelOrdersResponse struct {
	OrderID       string `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	State         string `json:"state"`
	Code          int64  `json:"code"`
	Message       string `json:"message"`
}

// TimeStampResponse returns the time
type TimeStampResponse struct {
	ServerTime int64 `json:"serverTime"`
}

//  ---------------------------------------------------------------- New ----------------------------------------------------------------

type SymbolDetail struct {
	Symbol            string               `json:"symbol"`
	BaseCurrencyName  string               `json:"baseCurrencyName"`
	QuoteCurrencyName string               `json:"quoteCurrencyName"`
	DisplayName       string               `json:"displayName"`
	State             string               `json:"state"`
	VisibleStartTime  convert.ExchangeTime `json:"visibleStartTime"`
	TradableStartTime convert.ExchangeTime `json:"tradableStartTime"`
	SymbolTradeLimit  struct {
		Symbol        string                  `json:"symbol"`
		PriceScale    float64                 `json:"priceScale"`
		QuantityScale float64                 `json:"quantityScale"`
		AmountScale   float64                 `json:"amountScale"`
		MinQuantity   convert.StringToFloat64 `json:"minQuantity"`
		MinAmount     convert.StringToFloat64 `json:"minAmount"`
		HighestBid    convert.StringToFloat64 `json:"highestBid"`
		LowestAsk     convert.StringToFloat64 `json:"lowestAsk"`
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
		ID               int64                   `json:"id"`
		Coin             string                  `json:"coin"`
		Name             string                  `json:"name"`
		CurrencyType     string                  `json:"currencyType"`
		Blockchain       string                  `json:"blockchain"`
		WithdrawalEnable bool                    `json:"withdrawalEnable"`
		DepositEnable    bool                    `json:"depositEnable"`
		DepositAddress   string                  `json:"depositAddress"`
		Decimals         float64                 `json:"decimals"`
		MinConfirm       float64                 `json:"minConfirm"`
		WithdrawMin      convert.StringToFloat64 `json:"withdrawMin"`
		WithdrawFee      convert.StringToFloat64 `json:"withdrawFee"`
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
	Symbol        string                  `json:"symbol"`
	DailyChange   convert.StringToFloat64 `json:"dailyChange"`
	Price         convert.StringToFloat64 `json:"price"`
	Timestamp     convert.ExchangeTime    `json:"time"`
	PushTimestamp convert.ExchangeTime    `json:"ts"`
}

// MarkPrice represents latest mark price for all cross margin symbols.
type MarkPrice struct {
	Symbol          string                  `json:"symbol"`
	MarkPrice       convert.StringToFloat64 `json:"markPrice"`
	RecordTimestamp convert.ExchangeTime    `json:"time"`
}

// MarkPriceComponent represents a mark price component instance.
type MarkPriceComponent struct {
	Symbol     string                  `json:"symbol"`
	Timestamp  convert.ExchangeTime    `json:"ts"`
	MarkPrice  convert.StringToFloat64 `json:"markPrice"`
	Components []struct {
		Symbol       string                  `json:"symbol"`
		Exchange     string                  `json:"exchange"`
		SymbolPrice  convert.StringToFloat64 `json:"symbolPrice"`
		Weight       convert.StringToFloat64 `json:"weight"`
		ConvertPrice convert.StringToFloat64 `json:"convertPrice"`
	} `json:"components"`
}

// OrderbookData represents an order book data for a specific symbol.
type OrderbookData struct {
	CreationTime  convert.ExchangeTime `json:"time"`
	Scale         string               `json:"scale"`
	Asks          []string             `json:"asks"`
	Bids          []string             `json:"bids"`
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
	var err error
	var candle *CandlestickData
	for i := range candlestickData {
		candle, err = getCandlestickData(candlestickData[i])
		if err != nil {
			return nil, err
		}
		candles[i] = *candle
	}
	return candles, nil
}

func getCandlestickData(candlestickData CandlestickArrayData) (*CandlestickData, error) {
	candle := &CandlestickData{}
	var err error
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
	candle.BuyTakerQuantity, err = strconv.ParseFloat(candlestickData[7].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.TradeCount = candlestickData[8].(float64)
	candle.PushTimestamp = time.UnixMilli(int64(candlestickData[9].(float64)))
	candle.WeightedAverage, err = strconv.ParseFloat(candlestickData[10].(string), 64)
	if err != nil {
		return nil, err
	}
	candle.Interval, err = stringToInterval(candlestickData[11].(string))
	if err != nil {
		return nil, err
	}
	candle.StartTime = time.UnixMilli(int64(candlestickData[12].(float64)))
	candle.EndTime = time.UnixMilli(int64(candlestickData[13].(float64)))
	return candle, nil
}

// Trade represents a trade instance.
type Trade struct {
	ID         string                  `json:"id"`
	Price      convert.StringToFloat64 `json:"price"`
	Quantity   convert.StringToFloat64 `json:"quantity"`
	Amount     convert.StringToFloat64 `json:"amount"`
	TakerSide  string                  `json:"takerSide"`
	Timestamp  convert.ExchangeTime    `json:"ts"`
	CreateTime convert.ExchangeTime    `json:"createTime"`
}

// TickerData represents a price ticker information.
type TickerData struct {
	Symbol      string                  `json:"symbol"`
	Open        convert.StringToFloat64 `json:"open"`
	Low         convert.StringToFloat64 `json:"low"`
	High        convert.StringToFloat64 `json:"high"`
	Close       convert.StringToFloat64 `json:"close"`
	Quantity    convert.StringToFloat64 `json:"quantity"`
	Amount      convert.StringToFloat64 `json:"amount"`
	TradeCount  int64                   `json:"tradeCount"`
	StartTime   convert.ExchangeTime    `json:"startTime"`
	CloseTime   convert.ExchangeTime    `json:"closeTime"`
	DisplayName string                  `json:"displayName"`
	DailyChange string                  `json:"dailyChange"`
	Bid         convert.StringToFloat64 `json:"bid"`
	BidQuantity convert.StringToFloat64 `json:"bidQuantity"`
	Ask         convert.StringToFloat64 `json:"ask"`
	AskQuantity convert.StringToFloat64 `json:"askQuantity"`
	Timestamp   convert.ExchangeTime    `json:"ts"`
	MarkPrice   convert.StringToFloat64 `json:"markPrice"`
}

// CollateralInfo represents collateral information.
type CollateralInfo struct {
	Currency              string                  `json:"currency"`
	CollateralRate        convert.StringToFloat64 `json:"collateralRate"`
	InitialMarginRate     convert.StringToFloat64 `json:"initialMarginRate"`
	MaintenanceMarginRate convert.StringToFloat64 `json:"maintenanceMarginRate"`
}

// BorrowRateinfo represents borrow rates information
type BorrowRateinfo struct {
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
		CurrencyID string                  `json:"currencyId"`
		Currency   string                  `json:"currency"`
		Available  convert.StringToFloat64 `json:"available"`
		Hold       convert.StringToFloat64 `json:"hold"`
	} `json:"balances"`
}

// AccountActivity represents activities such as airdrop, rebates, staking,
// credit/debit adjustments, and other (historical adjustments).
type AccountActivity struct {
	ID           string                  `json:"id"`
	Currency     string                  `json:"currency"`
	Amount       convert.StringToFloat64 `json:"amount"`
	State        string                  `json:"state"`
	CreateTime   convert.ExchangeTime    `json:"createTime"`
	Description  string                  `json:"description"`
	ActivityType int64                   `json:"activityType"`
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
	ID          string                  `json:"id"`
	FromAccount string                  `json:"fromAccount"`
	ToAccount   string                  `json:"toAccount"`
	Currency    string                  `json:"currency"`
	State       string                  `json:"state"`
	CreateTime  convert.ExchangeTime    `json:"createTime"`
	Amount      convert.StringToFloat64 `json:"amount"`
}

// FeeInfo represents an account transfer information.
type FeeInfo struct {
	TransactionDiscount bool                    `json:"trxDiscount"`
	MakerRate           convert.StringToFloat64 `json:"makerRate"`
	TakerRate           convert.StringToFloat64 `json:"takerRate"`
	Volume30D           convert.StringToFloat64 `json:"volume30D"`
	SpecialFeeRates     []struct {
		Symbol    string                  `json:"symbol"`
		MakerRate convert.StringToFloat64 `json:"makerRate"`
		TakerRate convert.StringToFloat64 `json:"takerRate"`
	} `json:"specialFeeRates"`
}

// InterestHistory represents an interest history.
type InterestHistory struct {
	ID                  string                  `json:"id"`
	CurrencyName        string                  `json:"currencyName"`
	Principal           convert.StringToFloat64 `json:"principal"`
	Interest            convert.StringToFloat64 `json:"interest"`
	InterestRate        convert.StringToFloat64 `json:"interestRate"`
	InterestAccuredTime convert.ExchangeTime    `json:"interestAccuredTime"`
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
		Currency              string                  `json:"currency"`
		Available             convert.StringToFloat64 `json:"available"`
		Hold                  convert.StringToFloat64 `json:"hold"`
		MaxAvailable          convert.StringToFloat64 `json:"maxAvailable"`
		AccountEquity         string                  `json:"accountEquity,omitempty"`
		UnrealisedPNL         string                  `json:"unrealisedPNL,omitempty"`
		MarginBalance         string                  `json:"marginBalance,omitempty"`
		PositionMargin        string                  `json:"positionMargin,omitempty"`
		OrderMargin           string                  `json:"orderMargin,omitempty"`
		FrozenFunds           string                  `json:"frozenFunds,omitempty"`
		AvailableBalance      convert.StringToFloat64 `json:"availableBalance,omitempty"`
		RealizedProfitAndLoss string                  `json:"pnl,omitempty"`
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
	ID              string                  `json:"id"`
	FromAccountID   string                  `json:"fromAccountId"`
	FromAccountName string                  `json:"fromAccountName"`
	FromAccountType string                  `json:"fromAccountType"`
	ToAccountID     string                  `json:"toAccountId"`
	ToAccountName   string                  `json:"toAccountName"`
	ToAccountType   string                  `json:"toAccountType"`
	Currency        string                  `json:"currency"`
	Amount          convert.StringToFloat64 `json:"amount"`
	State           string                  `json:"state"`
	CreateTime      convert.ExchangeTime    `json:"createTime"`
}

// WalletActivityResponse holds wallet activity info
type WalletActivityResponse struct {
	Deposits    []WalletDeposits    `json:"deposits"`
	Withdrawals []WalletWithdrawals `json:"withdrawals"`
}

// WalletDeposits holds wallet deposit info
type WalletDeposits struct {
	DepositNumber int64                   `json:"depositNumber"`
	Currency      string                  `json:"currency"`
	Address       string                  `json:"address"`
	Amount        convert.StringToFloat64 `json:"amount"`
	Confirmations int64                   `json:"confirmations"`
	TransactionID string                  `json:"txid"`
	Timestamp     convert.ExchangeTime    `json:"timestamp"`
	Status        string                  `json:"status"`
}

// WalletWithdrawals holds wallet withdrawal info
type WalletWithdrawals struct {
	WithdrawalRequestsID int64                   `json:"withdrawalRequestsId"`
	Currency             string                  `json:"currency"`
	Address              string                  `json:"address"`
	Amount               convert.StringToFloat64 `json:"amount"`
	Fee                  convert.StringToFloat64 `json:"fee"`
	Timestamp            convert.ExchangeTime    `json:"timestamp"`
	Status               string                  `json:"status"`
	TransactionID        string                  `json:"txid"`
	IPAddress            string                  `json:"ipAddress"`
	PaymentID            string                  `json:"paymentID"`
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
	TotalAccountValue convert.StringToFloat64 `json:"totalAccountValue"`
	TotalMargin       convert.StringToFloat64 `json:"totalMargin"`
	UsedMargin        convert.StringToFloat64 `json:"usedMargin"`
	FreeMargin        convert.StringToFloat64 `json:"freeMargin"`
	MaintenanceMargin convert.StringToFloat64 `json:"maintenanceMargin"`
	CreationTime      convert.ExchangeTime    `json:"time"`
	MarginRatio       string                  `json:"marginRatio"`
}

// BorroweStatus represents currency borrow status.
type BorroweStatus struct {
	Currency         string                  `json:"currency"`
	Available        convert.StringToFloat64 `json:"available"`
	Borrowed         convert.StringToFloat64 `json:"borrowed"`
	Hold             convert.StringToFloat64 `json:"hold"`
	MaxAvailable     convert.StringToFloat64 `json:"maxAvailable"`
	HourlyBorrowRate convert.StringToFloat64 `json:"hourlyBorrowRate"`
	Version          string                  `json:"version"`
}

// MaxBuySellAmount represents a maximum buy and sell amount.
type MaxBuySellAmount struct {
	Symbol           string                  `json:"symbol"`
	MaxLeverage      int64                   `json:"maxLeverage"`
	AvailableBuy     convert.StringToFloat64 `json:"availableBuy"`
	MaxAvailableBuy  convert.StringToFloat64 `json:"maxAvailableBuy"`
	AvailableSell    convert.StringToFloat64 `json:"availableSell"`
	MaxAvailableSell convert.StringToFloat64 `json:"maxAvailableSell"`
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
	Code          int    `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
}

// CancelReplaceOrderParam represents a cancellation and order replacement request parameter.
type CancelReplaceOrderParam struct {
	ID                string  `json:"-"`
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
	ID             string                  `json:"id"`
	ClientOrderID  string                  `json:"clientOrderId"`
	Symbol         string                  `json:"symbol"`
	State          string                  `json:"state"`
	AccountType    string                  `json:"accountType"`
	Side           string                  `json:"side"`
	Type           string                  `json:"type"`
	TimeInForce    string                  `json:"timeInForce"`
	Quantity       convert.StringToFloat64 `json:"quantity"`
	Price          convert.StringToFloat64 `json:"price"`
	AvgPrice       convert.StringToFloat64 `json:"avgPrice"`
	Amount         convert.StringToFloat64 `json:"amount"`
	FilledQuantity convert.StringToFloat64 `json:"filledQuantity"`
	FilledAmount   convert.StringToFloat64 `json:"filledAmount"`
	CreateTime     convert.ExchangeTime    `json:"createTime"`
	UpdateTime     convert.ExchangeTime    `json:"updateTime"`
	OrderSource    string                  `json:"orderSource"`
	Loan           bool                    `json:"loan"`
	CancelReason   int64                   `json:"cancelReason"`
}

// SmartOrderItem represents a smart order detail.
type SmartOrderItem struct {
	ID            string                  `json:"id"`
	ClientOrderID string                  `json:"clientOrderId"`
	Symbol        string                  `json:"symbol"`
	State         string                  `json:"state"`
	AccountType   string                  `json:"accountType"`
	Side          string                  `json:"side"`
	Type          string                  `json:"type"`
	TimeInForce   string                  `json:"timeInForce"`
	Quantity      convert.StringToFloat64 `json:"quantity"`
	Price         convert.StringToFloat64 `json:"price"`
	Amount        convert.StringToFloat64 `json:"amount"`
	StopPrice     convert.StringToFloat64 `json:"stopPrice"`
	CreateTime    convert.ExchangeTime    `json:"createTime"`
	UpdateTime    convert.ExchangeTime    `json:"updateTime"`
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
	ID               string  `json:"-"`
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
	ID        string                  `json:"id"`
	StopPrice convert.StringToFloat64 `json:"stopPrice"`
	Price     convert.ExchangeTime    `json:"price"`
	Quantity  convert.ExchangeTime    `json:"quantity"`
	Code      int64                   `json:"code"`
	Message   string                  `json:"message"`
}

// SmartOrderDetail represents a smart order information and trigger detailed information.
type SmartOrderDetail struct {
	ID             string                  `json:"id"`
	ClientOrderID  string                  `json:"clientOrderId"`
	Symbol         string                  `json:"symbol"`
	State          string                  `json:"state"`
	AccountType    string                  `json:"accountType"`
	Side           string                  `json:"side"`
	Type           string                  `json:"type"`
	TimeInForce    string                  `json:"timeInForce"`
	Quantity       convert.StringToFloat64 `json:"quantity"`
	Price          convert.StringToFloat64 `json:"price"`
	Amount         convert.StringToFloat64 `json:"amount"`
	StopPrice      convert.StringToFloat64 `json:"stopPrice"`
	CreateTime     convert.ExchangeTime    `json:"createTime"`
	UpdateTime     convert.ExchangeTime    `json:"updateTime"`
	TriggeredOrder struct {
		ID             string                  `json:"id"`
		ClientOrderID  string                  `json:"clientOrderId"`
		Symbol         string                  `json:"symbol"`
		State          string                  `json:"state"`
		AccountType    string                  `json:"accountType"`
		Side           string                  `json:"side"`
		Type           string                  `json:"type"`
		TimeInForce    string                  `json:"timeInForce"`
		Quantity       convert.StringToFloat64 `json:"quantity"`
		Price          convert.StringToFloat64 `json:"price"`
		AvgPrice       convert.StringToFloat64 `json:"avgPrice"`
		Amount         convert.StringToFloat64 `json:"amount"`
		FilledQuantity convert.StringToFloat64 `json:"filledQuantity"`
		FilledAmount   convert.StringToFloat64 `json:"filledAmount"`
		CreateTime     convert.ExchangeTime    `json:"createTime"`
		UpdateTime     convert.ExchangeTime    `json:"updateTime"`
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
	ID            string                  `json:"id"`
	ClientOrderID string                  `json:"clientOrderId"`
	Symbol        string                  `json:"symbol"`
	AccountType   string                  `json:"accountType"`
	OrderID       string                  `json:"orderId"`
	Side          string                  `json:"side"`
	Type          string                  `json:"type"`
	MatchRole     string                  `json:"matchRole"`
	Price         convert.StringToFloat64 `json:"price"`
	Quantity      convert.StringToFloat64 `json:"quantity"`
	Amount        convert.StringToFloat64 `json:"amount"`
	FeeCurrency   string                  `json:"feeCurrency"`
	FeeAmount     convert.StringToFloat64 `json:"feeAmount"`
	PageID        string                  `json:"pageId"`
	CreateTime    convert.ExchangeTime    `json:"createTime"`
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
		Symbol        string                  `json:"symbol"`
		PriceScale    float64                 `json:"priceScale"`
		QuantityScale float64                 `json:"quantityScale"`
		AmountScale   float64                 `json:"amountScale"`
		MinQuantity   convert.StringToFloat64 `json:"minQuantity"`
		MinAmount     convert.StringToFloat64 `json:"minAmount"`
		HighestBid    convert.StringToFloat64 `json:"highestBid"`
		LowestAsk     convert.StringToFloat64 `json:"lowestAsk"`
	} `json:"symbolTradeLimit"`
}

// WsCurrency represents a currency instance from websocket stream.
type WsCurrency struct {
	Currency          string   `json:"currency"`
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Type              string   `json:"type"`
	WithdrawalFee     string   `json:"withdrawalFee"`
	MinConf           int      `json:"minConf"`
	DepositAddress    any      `json:"depositAddress"`
	Blockchain        string   `json:"blockchain"`
	Delisted          bool     `json:"delisted"`
	TradingState      string   `json:"tradingState"`
	WalletState       string   `json:"walletState"`
	ParentChain       any      `json:"parentChain"`
	IsMultiChain      bool     `json:"isMultiChain"`
	IsChildChain      bool     `json:"isChildChain"`
	SupportCollateral bool     `json:"supportCollateral"`
	SupportBorrow     bool     `json:"supportBorrow"`
	ChildChains       []string `json:"childChains"`
}

type WsExchangeStatus []struct {
	Mm  string `json:"MM"`
	Pom string `json:"POM"`
}

// WsCandles represents a candlestick data instance.
type WsCandles struct {
	Symbol     string                  `json:"symbol"`
	Open       convert.StringToFloat64 `json:"open"`
	High       convert.StringToFloat64 `json:"high"`
	Low        convert.StringToFloat64 `json:"low"`
	Close      convert.StringToFloat64 `json:"close"`
	Quantity   convert.StringToFloat64 `json:"quantity"`
	Amount     convert.StringToFloat64 `json:"amount"`
	TradeCount int64                   `json:"tradeCount"`
	StartTime  convert.ExchangeTime    `json:"startTime"`
	CloseTime  convert.ExchangeTime    `json:"closeTime"`
	Timestamp  convert.ExchangeTime    `json:"ts"`
}

// WsTrade represents websocket trade data
type WsTrade struct {
	ID         string                  `json:"id"`
	Symbol     string                  `json:"symbol"`
	Amount     convert.StringToFloat64 `json:"amount"`
	Quantity   convert.StringToFloat64 `json:"quantity"`
	TakerSide  string                  `json:"takerSide"`
	Price      convert.StringToFloat64 `json:"price"`
	CreateTime convert.ExchangeTime    `json:"createTime"`
	Timestamp  convert.ExchangeTime    `json:"ts"`
}

// WsTicker represents a websocket ticker information.
type WsTicker struct {
	Symbol      string                  `json:"symbol"`
	StartTime   convert.ExchangeTime    `json:"startTime"`
	Open        convert.StringToFloat64 `json:"open"`
	High        convert.StringToFloat64 `json:"high"`
	Low         convert.StringToFloat64 `json:"low"`
	Close       convert.StringToFloat64 `json:"close"`
	Quantity    convert.StringToFloat64 `json:"quantity"`
	Amount      convert.StringToFloat64 `json:"amount"`
	TradeCount  int64                   `json:"tradeCount"`
	DailyChange convert.StringToFloat64 `json:"dailyChange"`
	MarkPrice   convert.StringToFloat64 `json:"markPrice"`
	CloseTime   convert.ExchangeTime    `json:"closeTime"`
	Timestamp   convert.ExchangeTime    `json:"ts"`
}

// WsBook represents an orderbook.
type WsBook struct {
	Symbol     string               `json:"symbol"`
	Asks       [][]string           `json:"asks"`
	Bids       [][]string           `json:"bids"`
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
	Symbol         string                  `json:"symbol"`
	Type           string                  `json:"type"`
	Quantity       convert.StringToFloat64 `json:"quantity"`
	OrderID        string                  `json:"orderId"`
	TradeFee       convert.StringToFloat64 `json:"tradeFee"`
	ClientOrderID  string                  `json:"clientOrderId"`
	AccountType    string                  `json:"accountType"`
	FeeCurrency    string                  `json:"feeCurrency"`
	EventType      string                  `json:"eventType"`
	Source         string                  `json:"source"`
	Side           string                  `json:"side"`
	FilledQuantity convert.StringToFloat64 `json:"filledQuantity"`
	FilledAmount   convert.StringToFloat64 `json:"filledAmount"`
	MatchRole      string                  `json:"matchRole"`
	State          string                  `json:"state"`
	TradeTime      convert.ExchangeTime    `json:"tradeTime"`
	TradeAmount    convert.StringToFloat64 `json:"tradeAmount"`
	OrderAmount    convert.StringToFloat64 `json:"orderAmount"`
	CreateTime     convert.ExchangeTime    `json:"createTime"`
	Price          convert.StringToFloat64 `json:"price"`
	TradeQty       convert.StringToFloat64 `json:"tradeQty"`
	TradePrice     convert.StringToFloat64 `json:"tradePrice"`
	TradeID        string                  `json:"tradeId"`
	Timestamp      convert.ExchangeTime    `json:"ts"`
}

// WsTradeBalance represents a balance information through the websocket channel
type WsTradeBalance []struct {
	ID          int64                   `json:"id"`
	UserID      int64                   `json:"userId"`
	ChangeTime  convert.ExchangeTime    `json:"changeTime"`
	AccountID   string                  `json:"accountId"`
	AccountType string                  `json:"accountType"`
	EventType   string                  `json:"eventType"`
	Available   convert.StringToFloat64 `json:"available"`
	Currency    string                  `json:"currency"`
	Hold        convert.StringToFloat64 `json:"hold"`
	Timestamp   convert.ExchangeTime    `json:"ts"`
}

// WebsocketResponse represents a websocket responses.
type WebsocketResponse struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}
