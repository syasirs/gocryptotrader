package coinut

import (
	"github.com/thrasher-/gocryptotrader/currency"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
)

// GenericResponse is the generic response you will get from coinut
type GenericResponse struct {
	Nonce   int64    `json:"nonce"`
	Reply   string   `json:"reply"`
	Status  []string `json:"status"`
	TransID int64    `json:"trans_id"`
}

// InstrumentBase holds information on base currency
type InstrumentBase struct {
	Base          string `json:"base"`
	DecimalPlaces int    `json:"decimal_places"`
	InstID        int    `json:"inst_id"`
	Quote         string `json:"quote"`
}

// Instruments holds the full information on base currencies
type Instruments struct {
	Instruments map[string][]InstrumentBase `json:"SPOT"`
}

// Ticker holds ticker information
type Ticker struct {
	HighestBuy   float64 `json:"highest_buy,string"`
	InstrumentID int     `json:"inst_id"`
	Last         float64 `json:"last,string"`
	LowestSell   float64 `json:"lowest_sell,string"`
	OpenInterest float64 `json:"open_interest,string"`
	Timestamp    float64 `json:"timestamp"`
	TransID      int64   `json:"trans_id"`
	Volume       float64 `json:"volume,string"`
	Volume24     float64 `json:"volume24,string"`
}

// OrderbookBase is a sub-type holding price and quantity
type OrderbookBase struct {
	Count    int     `json:"count"`
	Price    float64 `json:"price,string"`
	Quantity float64 `json:"qty,string"`
}

// Orderbook is the full order book
type Orderbook struct {
	Buy          []OrderbookBase `json:"buy"`
	Sell         []OrderbookBase `json:"sell"`
	InstrumentID int             `json:"inst_id"`
	TotalBuy     float64         `json:"total_buy,string"`
	TotalSell    float64         `json:"total_sell,string"`
	TransID      int64           `json:"trans_id"`
}

// TradeBase is a sub-type holding information on trades
type TradeBase struct {
	Price     float64 `json:"price,string"`
	Quantity  float64 `json:"quantity,string"`
	Side      string  `json:"side"`
	Timestamp float64 `json:"timestamp"`
	TransID   int64   `json:"trans_id"`
}

// Trades holds the full amount of trades associated with API keys
type Trades struct {
	Trades []TradeBase `json:"trades"`
}

// UserBalance holds user balances on the exchange
type UserBalance struct {
	BCH     float64  `json:"BCH,string"`
	BTC     float64  `json:"BTC,string"`
	BTG     float64  `json:"BTG,string"`
	CAD     float64  `json:"CAD,string"`
	ETC     float64  `json:"ETC,string"`
	ETH     float64  `json:"ETH,string"`
	LCH     float64  `json:"LCH,string"`
	LTC     float64  `json:"LTC,string"`
	MYR     float64  `json:"MYR,string"`
	SGD     float64  `json:"SGD,string"`
	USD     float64  `json:"USD,string"`
	USDT    float64  `json:"USDT,string"`
	XMR     float64  `json:"XMR,string"`
	ZEC     float64  `json:"ZEC,string"`
	Nonce   int64    `json:"nonce"`
	Reply   string   `json:"reply"`
	Status  []string `json:"status"`
	TransID int64    `json:"trans_id"`
}

// Order holds order information
type Order struct {
	InstrumentID  int64   `json:"inst_id"`
	Price         float64 `json:"price,string"`
	Quantity      float64 `json:"qty,string"`
	ClientOrderID int     `json:"client_ord_id"`
	Side          string  `json:"side,string"`
}

// OrderResponse is a response for orders
type OrderResponse struct {
	OrderID       int64   `json:"order_id"`
	OpenQuantity  float64 `json:"open_qty,string"`
	Price         float64 `json:"price,string"`
	Quantity      float64 `json:"qty,string"`
	InstrumentID  int64   `json:"inst_id"`
	ClientOrderID int64   `json:"client_ord_id"`
	Timestamp     int64   `json:"timestamp"`
	OrderPrice    float64 `json:"order_price,string"`
	Side          string  `json:"side"`
}

// Commission holds trade commission structure
type Commission struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount,string"`
}

// OrderFilledResponse contains order filled response
type OrderFilledResponse struct {
	GenericResponse
	Commission   Commission    `json:"commission"`
	FillPrice    float64       `json:"fill_price,string"`
	FillQuantity float64       `json:"fill_qty,string"`
	Order        OrderResponse `json:"order"`
}

// OrderRejectResponse holds information on a rejected order
type OrderRejectResponse struct {
	OrderResponse
	Reasons []string `json:"reasons"`
}

// OrdersBase contains generic response and order responses
type OrdersBase struct {
	GenericResponse
	OrderResponse
}

// GetOpenOrdersResponse holds all order data from GetOpenOrders request
type GetOpenOrdersResponse struct {
	Nonce   int             `json:"nonce"`
	Orders  []OrderResponse `json:"orders"`
	Reply   string          `json:"reply"`
	Status  []string        `json:"status"`
	TransID int             `json:"trans_id"`
}

// OrdersResponse holds the full data range on orders
type OrdersResponse struct {
	Data []OrdersBase
}

// CancelOrders holds information about a cancelled order
type CancelOrders struct {
	InstrumentID int64 `json:"inst_id"`
	OrderID      int64 `json:"order_id"`
}

// CancelOrdersResponse is response for a cancelled order
type CancelOrdersResponse struct {
	GenericResponse
	Results []struct {
		OrderID      int64  `json:"order_id"`
		Status       string `json:"status"`
		InstrumentID int    `json:"inst_id"`
	} `json:"results"`
}

// TradeHistory holds trade history information
type TradeHistory struct {
	TotalNumber int64                 `json:"total_number"`
	Trades      []OrderFilledResponse `json:"trades"`
}

// IndexTicker holds indexed ticker inforamtion
type IndexTicker struct {
	Asset string  `json:"asset"`
	Price float64 `json:"price,string"`
}

// Option holds options information
type Option struct {
	HighestBuy   float64 `json:"highest_buy,string"`
	InstrumentID int     `json:"inst_id"`
	Last         float64 `json:"last,string"`
	LowestSell   float64 `json:"lowest_sell,string"`
	OpenInterest float64 `json:"open_interest,string"`
}

// OptionChainResponse is the response type for options
type OptionChainResponse struct {
	ExpiryTime   int64  `json:"expiry_time"`
	SecurityType string `json:"sec_type"`
	Asset        string `json:"asset"`
	Entries      []struct {
		Call   Option  `json:"call"`
		Put    Option  `json:"put"`
		Strike float64 `json:"strike,string"`
	}
}

// OptionChainUpdate contains information on the chain update options
type OptionChainUpdate struct {
	Option
	GenericResponse
	Asset        string  `json:"asset"`
	ExpiryTime   int64   `json:"expiry_time"`
	SecurityType string  `json:"sec_type"`
	Volume       float64 `json:"volume,string"`
}

// PositionHistory holds the complete position history
type PositionHistory struct {
	Positions []struct {
		PositionID int `json:"position_id"`
		Records    []struct {
			Commission    Commission `json:"commission"`
			FillPrice     float64    `json:"fill_price,string,omitempty"`
			TransactionID int        `json:"trans_id"`
			FillQuantity  float64    `json:"fill_qty,omitempty"`
			Position      struct {
				Commission Commission `json:"commission"`
				Timestamp  int64      `json:"timestamp"`
				OpenPrice  float64    `json:"open_price,string"`
				RealizedPL float64    `json:"realized_pl,string"`
				Quantity   float64    `json:"qty,string"`
			} `json:"position"`
			AssetAtExpiry float64 `json:"asset_at_expiry,string,omitempty"`
		} `json:"records"`
		Instrument struct {
			ExpiryTime     int64   `json:"expiry_time"`
			ContractSize   float64 `json:"contract_size,string"`
			ConversionRate float64 `json:"conversion_rate,string"`
			OptionType     string  `json:"option_type"`
			InstrumentID   int     `json:"inst_id"`
			SecType        string  `json:"sec_type"`
			Asset          string  `json:"asset"`
			Strike         float64 `json:"strike,string"`
		} `json:"inst"`
		OpenTimestamp int64 `json:"open_timestamp"`
	} `json:"positions"`
	TotalNumber int `json:"total_number"`
}

// OpenPosition holds information on an open position
type OpenPosition struct {
	PositionID    int        `json:"position_id"`
	Commission    Commission `json:"commission"`
	OpenPrice     float64    `json:"open_price,string"`
	RealizedPL    float64    `json:"realized_pl,string"`
	Quantity      float64    `json:"qty,string"`
	OpenTimestamp int64      `json:"open_timestamp"`
	InstrumentID  int        `json:"inst_id"`
}

type wsRequest struct {
	Request   string `json:"request"`
	SecType   string `json:"sec_type,omitempty"`
	InstID    int64  `json:"inst_id,omitempty"`
	TopN      int64  `json:"top_n,omitempty"`
	Subscribe bool   `json:"subscribe"`
	Nonce     int64  `json:"nonce"`
}

type wsResponse struct {
	Reply string `json:"reply"`
}

type wsHeartbeatResp struct {
	Nonce  int64         `json:"nonce"`
	Reply  string        `json:"reply"`
	Status []interface{} `json:"status"`
}

// WsTicker defines the resp for ticker updates from the websocket connection
type WsTicker struct {
	HighestBuy   float64 `json:"highest_buy,string"`
	InstID       int64   `json:"inst_id"`
	Last         float64 `json:"last,string"`
	LowestSell   float64 `json:"lowest_sell,string"`
	OpenInterest float64 `json:"open_interest,string"`
	Reply        string  `json:"reply"`
	Timestamp    int64   `json:"timestamp"`
	TransID      int64   `json:"trans_id"`
	Volume       float64 `json:"volume,string"`
	Volume24H    float64 `json:"volume24,string"`
}

// WsOrderbookSnapshot defines the resp for orderbook snapshot updates from
// the websocket connection
type WsOrderbookSnapshot struct {
	Buy       []WsOrderbookData `json:"buy"`
	Sell      []WsOrderbookData `json:"sell"`
	InstID    int64             `json:"inst_id"`
	Nonce     int64             `json:"nonce"`
	TotalBuy  float64           `json:"total_buy,string"`
	TotalSell float64           `json:"total_sell,string"`
	Reply     string            `json:"reply"`
	Status    []interface{}     `json:"status"`
}

// WsOrderbookData defines singular orderbook data
type WsOrderbookData struct {
	Count  int64   `json:"count"`
	Price  float64 `json:"price,string"`
	Volume float64 `json:"qty,string"`
}

// WsOrderbookUpdate defines orderbook update response from the websocket
// connection
type WsOrderbookUpdate struct {
	Count    int64   `json:"count"`
	InstID   int64   `json:"inst_id"`
	Price    float64 `json:"price,string"`
	Volume   float64 `json:"qty,string"`
	TotalBuy float64 `json:"total_buy,string"`
	Reply    string  `json:"reply"`
	Side     string  `json:"side"`
	TransID  int64   `json:"trans_id"`
}

// WsTradeSnapshot defines Market trade response from the websocket
// connection
type WsTradeSnapshot struct {
	Nonce  int64         `json:"nonce"`
	Reply  string        `json:"reply"`
	Status []interface{} `json:"status"`
	Trades []WsTradeData `json:"trades"`
}

// WsTradeData defines market trade data
type WsTradeData struct {
	Price     float64 `json:"price,string"`
	Volume    float64 `json:"qty,string"`
	Side      string  `json:"side"`
	Timestamp int64   `json:"timestamp"`
	TransID   int64   `json:"trans_id"`
}

// WsTradeUpdate defines trade update response from the websocket connection
type WsTradeUpdate struct {
	InstID    int64   `json:"inst_id"`
	Price     float64 `json:"price,string"`
	Reply     string  `json:"reply"`
	Side      string  `json:"side"`
	Timestamp int64   `json:"timestamp"`
	TransID   int64   `json:"trans_id"`
}

// WsInstrumentList defines instrument list
type WsInstrumentList struct {
	Spot   map[string][]WsSupportedCurrency `json:"SPOT"`
	Nonce  int64                            `json:"nonce"`
	Reply  string                           `json:"inst_list"`
	Status []interface{}                    `json:"status"`
}

// WsSupportedCurrency defines supported currency on the exchange
type WsSupportedCurrency struct {
	Base          string `json:"base"`
	InstID        int64  `json:"inst_id"`
	DecimalPlaces int64  `json:"decimal_places"`
	Quote         string `json:"quote"`
}

// WsRequest base request
type WsRequest struct {
	Request string `json:"request"`
	Nonce   int64  `json:"nonce"`
}

// WsTradeHistoryRequest ws request
type WsTradeHistoryRequest struct {
	InstID int64 `json:"inst_id"`
	Start  int64 `json:"start,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
	WsRequest
}

// WsCancelOrdersRequest ws request
type WsCancelOrdersRequest struct {
	Entries []WsCancelOrdersRequestEntry `json:"entries"`
	WsRequest
}

// WsCancelOrdersRequestEntry ws request entry
type WsCancelOrdersRequestEntry struct {
	InstID  int64 `json:"inst_id"`
	OrderID int64 `json:"order_id"`
}

// WsCancelOrderParameters ws request parameters
type WsCancelOrderParameters struct {
	Currency currency.Pair
	OrderID  int64
}

// WsCancelOrderRequest ws request
type WsCancelOrderRequest struct {
	InstID  int64 `json:"inst_id"`
	OrderID int64 `json:"order_id"`
	WsRequest
}

// WsCancelOrderResponse ws response
type WsCancelOrderResponse struct {
	Nonce       int64    `json:"nonce"`
	Reply       string   `json:"reply"`
	OrderID     int64    `json:"order_id"`
	ClientOrdID int64    `json:"client_ord_id"`
	Status      []string `json:"status"`
}

// WsCancelOrdersResponse ws response
type WsCancelOrdersResponse struct {
	Request string                        `json:"request"`
	Entries []WsCancelOrdersResponseEntry `json:"entries"`
	Nonce   int64                         `json:"nonce"`
}

// WsCancelOrdersResponseEntry ws response entry
type WsCancelOrdersResponseEntry struct {
	InstID  int64 `json:"inst_id"`
	OrderID int64 `json:"order_id"`
}

// WsGetOpenOrdersRequest ws request
type WsGetOpenOrdersRequest struct {
	InstID int64 `json:"inst_id"`
	WsRequest
}

// WsSubmitOrdersRequest ws request
type WsSubmitOrdersRequest struct {
	Orders []WsSubmitOrdersRequestData `json:"orders"`
	WsRequest
}

// WsSubmitOrdersRequestData ws request data
type WsSubmitOrdersRequestData struct {
	InstID      int64   `json:"inst_id"`
	Price       float64 `json:"price,string"`
	Qty         float64 `json:"qty,string"`
	ClientOrdID int     `json:"client_ord_id"`
	Side        string  `json:"side"`
}

// WsSubmitOrderRequest ws request
type WsSubmitOrderRequest struct {
	InstID  int64   `json:"inst_id"`
	Price   float64 `json:"price,string"`
	Qty     float64 `json:"qty,string"`
	OrderID int64   `json:"client_ord_id"`
	Side    string  `json:"side"`
	WsRequest
}

// WsSubmitOrderParameters ws request parameters
type WsSubmitOrderParameters struct {
	Currency      currency.Pair
	Side          exchange.OrderSide
	Amount, Price float64
	OrderID       int64
}

// WsUserBalanceResponse ws response
type WsUserBalanceResponse struct {
	Nonce             int64    `json:"nonce"`
	Status            []string `json:"status"`
	Btc               string   `json:"BTC"`
	Ltc               string   `json:"LTC"`
	Etc               string   `json:"ETC"`
	Eth               string   `json:"ETH"`
	FloatingPl        string   `json:"floating_pl"`
	InitialMargin     string   `json:"initial_margin"`
	RealizedPl        string   `json:"realized_pl"`
	MaintenanceMargin string   `json:"maintenance_margin"`
	Equity            string   `json:"equity"`
	Reply             string   `json:"reply"`
	TransID           int64    `json:"trans_id"`
}

// WsOrderAcceptedResponse ws response
type WsOrderAcceptedResponse struct {
	Nonce       int64    `json:"nonce"`
	Status      []string `json:"status"`
	OrderID     int64    `json:"order_id"`
	OpenQty     string   `json:"open_qty"`
	InstID      int64    `json:"inst_id"`
	Qty         string   `json:"qty"`
	ClientOrdID int64    `json:"client_ord_id"`
	OrderPrice  string   `json:"order_price"`
	Reply       string   `json:"reply"`
	Side        string   `json:"side"`
	TransID     int64    `json:"trans_id"`
}

// WsOrderFilledResponse ws response
type WsOrderFilledResponse struct {
	Commission WsOrderFilledCommissionData `json:"commission"`
	FillPrice  string                      `json:"fill_price"`
	FillQty    string                      `json:"fill_qty"`
	Nonce      int64                       `json:"nonce"`
	Order      WsOrderFilledOrderData      `json:"order"`
	Reply      string                      `json:"reply"`
	Status     []string                    `json:"status"`
	Timestamp  int64                       `json:"timestamp"`
	TransID    int64                       `json:"trans_id"`
}

// WsOrderFilledOrderData ws response data
type WsOrderFilledOrderData struct {
	ClientOrdID int64  `json:"client_ord_id"`
	InstID      int64  `json:"inst_id"`
	OpenQty     string `json:"open_qty"`
	OrderID     int64  `json:"order_id"`
	Price       string `json:"price"`
	Qty         string `json:"qty"`
	Side        string `json:"side"`
	Timestamp   int64  `json:"timestamp"`
}

// WsOrderFilledCommissionData ws response data
type WsOrderFilledCommissionData struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// WsOrderRejectedResponse ws response
type WsOrderRejectedResponse struct {
	Nonce       int64    `json:"nonce"`
	Status      []string `json:"status"`
	OrderID     int64    `json:"order_id"`
	OpenQty     string   `json:"open_qty"`
	Price       string   `json:"price"`
	InstID      int64    `json:"inst_id"`
	Reasons     []string `json:"reasons"`
	ClientOrdID int64    `json:"client_ord_id"`
	Timestamp   int64    `json:"timestamp"`
	Reply       string   `json:"reply"`
	Qty         string   `json:"qty"`
	Side        string   `json:"side"`
	TransID     int64    `json:"trans_id"`
}

// WsUserOpenOrdersResponse ws response
type WsUserOpenOrdersResponse struct {
	Nonce  int64                       `json:"nonce"`
	Reply  string                      `json:"reply"`
	Status []string                    `json:"status"`
	Orders []WsUserOpenOrdersOrderData `json:"orders"`
}

// WsUserOpenOrdersOrderData ws response data
type WsUserOpenOrdersOrderData struct {
	OrderID     int64  `json:"order_id"`
	OpenQty     string `json:"open_qty"`
	Price       string `json:"price"`
	InstID      int64  `json:"inst_id"`
	ClientOrdID int64  `json:"client_ord_id"`
	Timestamp   int64  `json:"timestamp"`
	Qty         string `json:"qty"`
	Side        string `json:"side"`
}

// WsTradeHistoryResponse ws response
type WsTradeHistoryResponse struct {
	Nonce       int64                     `json:"nonce"`
	Reply       string                    `json:"reply"`
	Status      []string                  `json:"status"`
	TotalNumber int64                     `json:"total_number"`
	Trades      []WsTradeHistoryTradeData `json:"trades"`
}

// WsTradeHistoryOrderData ws response data
type WsTradeHistoryOrderData struct {
	ClientOrdID int64  `json:"client_ord_id"`
	InstID      int64  `json:"inst_id"`
	OpenQty     string `json:"open_qty"`
	OrderID     int64  `json:"order_id"`
	Price       string `json:"price"`
	Qty         string `json:"qty"`
	Side        string `json:"side"`
	Timestamp   int64  `json:"timestamp"`
}

// WsTradeHistoryCommissionData ws response data
type WsTradeHistoryCommissionData struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// WsTradeHistoryTradeData ws response data
type WsTradeHistoryTradeData struct {
	Commission WsTradeHistoryCommissionData `json:"commission"`
	Order      WsTradeHistoryOrderData      `json:"order"`
	FillPrice  string                       `json:"fill_price"`
	FillQty    string                       `json:"fill_qty"`
	Timestamp  int64                        `json:"timestamp"`
	TransID    int64                        `json:"trans_id"`
}

// WsLoginResponse ws response data
type WsLoginResponse struct {
	APIKey          string   `json:"api_key"`
	Country         string   `json:"country"`
	DepositEnabled  bool     `json:"deposit_enabled"`
	Deposited       bool     `json:"deposited"`
	Email           string   `json:"email"`
	FailedTimes     int64    `json:"failed_times"`
	KycPassed       bool     `json:"kyc_passed"`
	Lang            string   `json:"lang"`
	Nonce           int64    `json:"nonce"`
	OtpEnabled      bool     `json:"otp_enabled"`
	PhoneNumber     string   `json:"phone_number"`
	ProductsEnabled []string `json:"products_enabled"`
	Referred        bool     `json:"referred"`
	Reply           string   `json:"reply"`
	SessionID       string   `json:"session_id"`
	Status          []string `json:"status"`
	Timezone        string   `json:"timezone"`
	Traded          bool     `json:"traded"`
	UnverifiedEmail string   `json:"unverified_email"`
	Username        string   `json:"username"`
	WithdrawEnabled bool     `json:"withdraw_enabled"`
}

// WsNewOrderResponse returns if new_order response failes
type WsNewOrderResponse struct {
	Msg    string   `json:"msg"`
	Nonce  int64    `json:"nonce"`
	Reply  string   `json:"reply"`
	Status []string `json:"status"`
}
