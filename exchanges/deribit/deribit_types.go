package deribit

import (
	"errors"
	"time"
)

// UnmarshalError is the struct which is used for unmarshalling errors
type UnmarshalError struct {
	Message string `json:"message"`
	Data    struct {
		Reason string `json:"reason"`
	}
	Code int64 `json:"code"`
}

const (
	sideBUY  = "buy"
	sideSELL = "sell"

	// currencies

	currencyBTC  = "BTC"
	currencyETH  = "ETH"
	currencySOL  = "SOL"
	currencyUSDC = "USDC"
)

var (
	errTypeAssert                    = errors.New("type assertion failed")
	errStartTimeCannotBeAfterEndTime = errors.New("start timestamp cannot be after end timestamp")
	errUnsupportedIndexName          = errors.New("unsupported index name")
	errInvalidInstrumentID           = errors.New("invalid instrument ID")
	errInvalidIndexPriceCurrency     = errors.New("invalid currency, only BTC, ETH, SOL and USDC are supported")
	errInvalidInstrumentName         = errors.New("invalid instrument name")
	errInvalidComboID                = errors.New("invalid combo ID")
	errInvalidCurrency               = errors.New("invalid currency")
	errInvalidComboState             = errors.New("invalid combo state")
	errNoArgumentPassed              = errors.New("no argument passed")
	errInvalidAmount                 = errors.New("invalid amount, must be greater than 0")
	errMissingNonce                  = errors.New("missing nonce")
	errInvalidTradeRole              = errors.New("invalid trade role, only 'maker' and 'taker' are allowed")
	errInvalidPrice                  = errors.New("invalid trade price")
	errInvalidCryptoAddress          = errors.New("invalid crypto address")
	errIntervalNotSupported          = errors.New("iterval not supported")
	errInvalidTimestamp              = errors.New("invalid or zero timestamp")
	errInvalidID                     = errors.New("invalid id")
	errInvalidEmailAddress           = errors.New("invalid email address")
)

// BookSummaryData stores summary data
type BookSummaryData struct {
	VolumeUSD              float64 `json:"volume_usd"`
	Volume                 float64 `json:"volume"`
	QuoteCurrency          string  `json:"quote_currency"`
	PriceChange            float64 `json:"price_change"`
	OpenInterest           float64 `json:"open_interest"`
	MidPrice               float64 `json:"mid_price"`
	MarkPlace              float64 `json:"mark_place"`
	Low                    float64 `json:"low"`
	Last                   float64 `json:"last"`
	InstrumentName         string  `json:"instrument_name"`
	High                   float64 `json:"high"`
	EstimatedDeliveryPrice float64 `json:"estimated_delivery_price"`
	CreationTimestamp      int64   `json:"creation_timestamp"`
	BidPrice               float64 `json:"bid_price"`
	BaseCurrency           string  `json:"base_currency"`
	AskPrice               float64 `json:"ask_price"`
}

// ContractSizeData stores contract size for given instrument
type ContractSizeData struct {
	ContractSize float64 `json:"contract_size"`
}

// CurrencyData stores data for currencies
type CurrencyData struct {
	CoinType             string  `json:"coin_type"`
	Currency             string  `json:"currency"`
	CurrencyLong         string  `json:"currency_long"`
	FeePrecision         int64   `json:"fee_precision"`
	MinConfirmations     int64   `json:"min_confirmations"`
	MinWithdrawalFee     float64 `json:"min_withdrawal_fee"`
	WithdrawalFee        float64 `json:"withdrawal_fee"`
	WithdrawalPriorities []struct {
		Value float64 `json:"value"`
		Name  string  `json:"name"`
	} `json:"withdrawal_priorities"`
}

// IndexDeliveryPrice store index delivery prices list.
type IndexDeliveryPrice struct {
	Data         []DeliveryPriceData `json:"data"`
	TotalRecords int64               `json:"records_total"`
}

// DeliveryPriceData stores index delivery_price
type DeliveryPriceData struct {
	Date          string  `json:"date"`
	DeliveryPrice float64 `json:"delivery_price"`
}

// FundingChartData stores futures funding chart data
type FundingChartData struct {
	CurrentInterest float64 `json:"current_interest"`
	Data            []struct {
		IndexPrice float64 `json:"index_price"`
		Interest8H float64 `json:"interest_8h"`
		Timestamp  int64   `json:"timestamp"`
	} `json:"data"`
}

// FundingRateHistoryData stores data for funding rate history
type FundingRateHistoryData struct {
	Timestamp      int64   `json:"timestamp"`
	IndexPrice     float64 `json:"index_price"`
	PrevIndexPrice float64 `json:"prev_index_price"`
	Interest8H     float64 `json:"interest_8h"`
	Interest1H     float64 `json:"interest_1h"`
}

// FundingRateValueData stores funding rate for the requested period
type FundingRateValueData struct {
	Result float64 `json:"result"`
}

// HistoricalVolatilityData stores volatility data for requested symbols
type HistoricalVolatilityData struct {
	Timestamp float64
	Value     float64
}

// IndexPriceData gets index price data
type IndexPriceData struct {
	EstimatedDeliveryPrice float64 `json:"estimated_delivery_price"`
	IndexPrice             float64 `json:"index_price"`
}

// InstrumentData gets data for instruments
type InstrumentData struct {
	BaseCurrency         string  `json:"base_currency"`
	BlockTradeCommission float64 `json:"block_trade_commission"`
	ContractSize         float64 `json:"contract_size"`
	CreationTimestamp    int64   `json:"creation_timestamp"`
	ExpirationTimestamp  int64   `json:"expiration_timestamp"`
	InstrumentName       string  `json:"instrument_name"`
	IsActive             bool    `json:"is_active"`
	Kind                 string  `json:"kind"`
	Leverage             float64 `json:"leverage"`
	MakerCommission      float64 `json:"maker_commission"`
	MinimumTradeAmount   float64 `json:"min_trade_amount"`
	OptionType           string  `json:"option_type"`
	QuoteCurrency        string  `json:"quote_currency"`
	TickSize             float64 `json:"tick_size"`
	TakerCommission      float64 `json:"taker_commission"`
	Strike               float64 `json:"strike"`
	SettlementPeriod     string  `json:"settlement_period"`
}

// SettlementsData stores data for settlement futures
type SettlementsData struct {
	Settlements []struct {
		Funded            float64 `json:"funded"`
		Funding           float64 `json:"funding"`
		IndexPrice        float64 `json:"index_price"`
		SessionBankruptcy float64 `json:"session_bankrupcy"`
		SessionTax        float64 `json:"session_tax"`
		SessionTaxRate    float64 `json:"session_tax_rate"`
		Socialized        float64 `json:"socialized"`
		SettlementType    string  `json:"type"`
		Timestamp         int64   `json:"timestamp"`
		SessionProfitLoss float64 `json:"session_profit_loss"`
		ProfitLoss        float64 `json:"profit_loss"`
		Position          float64 `json:"position"`
		MarkPrice         float64 `json:"mark_price"`
		InstrumentName    string  `json:"instrument_name"`
	} `json:"settlements"`
	Continuation string `json:"continuation"`
}

// PublicTradesData stores data for public trades
type PublicTradesData struct {
	Trades []struct {
		TradeSeq       float64 `json:"trade_seq"`
		TradeID        string  `json:"trade_id"`
		Timestamp      int64   `json:"timestamp"`
		TickDirection  int64   `json:"tick_direction"`
		Price          float64 `json:"price"`
		MarkPrice      float64 `json:"mark_price"`
		Liquidation    string  `json:"liquidation"`
		IV             float64 `json:"iv"`
		InstrumentName string  `json:"instrument_name"`
		IndexPrice     float64 `json:"index_price"`
		Direction      string  `json:"direction"`
		BlockTradeID   string  `json:"block_trade_id"`
		Amount         float64 `json:"amount"`
	} `json:"trades"`
	HasMore bool `json:"has_more"`
}

// MarkPriceHistory stores data for mark price history
type MarkPriceHistory struct {
	Timestamp      int64
	MarkPriceValue float64
}

// Orderbook stores orderbook data
type Orderbook struct {
	UnderlyingPrice float64   `json:"underlying_price"`
	UnderlyingIndex string    `json:"underlying_index"`
	Timestamp       time.Time `json:"timestamp"`
	Stats           struct {
		Volume      float64 `json:"volume"`
		PriceChange float64 `json:"price_change"`
		Low         float64 `json:"low"`
		High        float64 `json:"high"`
	} `json:"stats"`
	State           string  `json:"state"`
	SettlementPrice float64 `json:"settlement_price"`
	OpenInterest    float64 `json:"open_interest"`
	MinPrice        float64 `json:"min_price"`
	MaxPrice        float64 `json:"max_price"`
	MarkIV          float64 `json:"mark_iv"`
	MarkPrice       float64 `json:"mark_price"`
	LastPrice       float64 `json:"last_price"`
	InterestRate    float64 `json:"interest_rate"`
	InstrumentName  string  `json:"instrument_name"`
	IndexPrice      float64 `json:"index_price"`
	GreeksData      struct {
		Delta float64 `json:"delta"`
		Gamma float64 `json:"gamma"`
		Rho   float64 `json:"rho"`
		Theta float64 `json:"theta"`
		Vega  float64 `json:"vega"`
	} `json:"greeks"`
	Funding8H      float64      `json:"funding_8h"`
	CurrentFunding float64      `json:"current_funding"`
	ChangeID       int64        `json:"change_id"`
	Bids           [][2]float64 `json:"bids"`
	Asks           [][2]float64 `json:"asks"`
	BidIV          float64      `json:"bid_iv"`
	BestBidPrice   float64      `json:"best_bid_price"`
	BestBidAmount  float64      `json:"best_bid_amount"`
	BestAskAmount  float64      `json:"best_ask_amount"`
	AskIV          float64      `json:"ask_iv"`
}

// TradeVolumesData stores data for trade volumes
type TradeVolumesData struct {
	PutsVolume       float64 `json:"puts_volume"`
	PutsVolume7D     float64 `json:"puts_volume_7d"`
	PutsVolume30D    float64 `json:"puts_volume_30d"`
	FuturesVolume7D  float64 `json:"futures_volume_7d"`
	FuturesVolume30D float64 `json:"futures_volume_30d"`
	FuturesVolume    float64 `json:"futures_volume"`
	CurrencyPair     string  `json:"currency_pair"`
	CallsVolume7D    float64 `json:"calls_volume_7d"`
	CallsVolume30D   float64 `json:"calls_volume_30d"`
	CallsVolume      float64 `json:"calls_volume"`
}

// TVChartData stores trading view chart data
type TVChartData struct {
	Volume []float64 `json:"volume"`
	Cost   []float64 `json:"cost"`
	Ticks  []float64 `json:"ticks"`
	Status string    `json:"status"`
	Open   []float64 `json:"open"`
	Low    []float64 `json:"low"`
	High   []float64 `json:"high"`
	Close  []float64 `json:"close"`
}

// VolatilityIndexData stores index data for volatility
type VolatilityIndexData struct {
	Data [][]float64 `json:"data"`
}

// TickerData stores data for ticker
type TickerData struct {
	AskIV          float64 `json:"ask_iv"`
	BestAskAmount  float64 `json:"best_ask_amount"`
	BestAskPrice   float64 `json:"best_ask_price"`
	BestBidAmount  float64 `json:"best_bid_amount"`
	BestBidPrice   float64 `json:"best_bid_price"`
	BidIV          float64 `json:"bid_iv"`
	CurrentFunding float64 `json:"current_funding"`
	DeliveryPrice  float64 `json:"delivery_price"`
	Funding8H      float64 `json:"funding_8h"`
	GreeksData     struct {
		Delta float64 `json:"delta"`
		Gamma float64 `json:"gamma"`
		Rho   float64 `json:"rho"`
		Theta float64 `json:"theta"`
		Vega  float64 `json:"vega"`
	} `json:"greeks"`
	IndexPrice      float64 `json:"index_price"`
	InstrumentName  string  `json:"instrument_name"`
	LastPrice       float64 `json:"last_price"`
	MarkIV          float64 `json:"mark_iv"`
	MarkPrice       float64 `json:"mark_price"`
	MaxPrice        float64 `json:"max_price"`
	MinPrice        float64 `json:"min_price"`
	OpenInterest    float64 `json:"open_interest"`
	SettlementPrice float64 `json:"settlement_price"`
	State           string  `json:"state"`
	Stats           struct {
		Volume      float64 `json:"volume"`
		PriceChange float64 `json:"price_change"`
		Low         float64 `json:"low"`
		High        float64 `json:"high"`
	} `json:"stats"`
	Timestamp       int64   `json:"timestamp"`
	UnderlyingIndex string  `json:"underlying_index"`
	UnderlyingPrice float64 `json:"underlying_price"`
}

// CancelTransferData stores data for a cancel transfer
type CancelTransferData struct {
	Amount           float64 `json:"amount"`
	CreatedTimestamp int64   `json:"created_timestamp"`
	Currency         string  `json:"currency"`
	Direction        string  `json:"direction"`
	ID               int64   `json:"id"`
	OtherSide        string  `json:"other_side"`
	State            string  `json:"state"`
	Type             string  `json:"type"`
	UpdatedTimestamp int64   `json:"updated_timestamp"`
}

// CancelWithdrawalData stores cancel request data for a withdrawal
type CancelWithdrawalData struct {
	Address            string  `json:"address"`
	Amount             float64 `json:"amount"`
	ConfirmedTimestamp int64   `json:"confirmed_timestamp"`
	CreatedTimestamp   int64   `json:"created_timestamp"`
	Currency           string  `json:"currency"`
	Fee                float64 `json:"fee"`
	ID                 int64   `json:"id"`
	Priority           float64 `json:"priority"`
	Status             string  `json:"status"`
	TransactionID      int64   `json:"transaction_id"`
	UpdatedTimestamp   int64   `json:"updated_timestamp"`
}

// DepositAddressData stores data of a deposit address
type DepositAddressData struct {
	Address           string `json:"address"`
	CreationTimestamp int64  `json:"creation_timestamp"`
	Currency          string `json:"currency"`
	Type              string `json:"type"`
}

// DepositsData stores data of deposits
type DepositsData struct {
	Count int64 `json:"count"`
	Data  []struct {
		Address           string  `json:"address"`
		Amount            float64 `json:"amount"`
		Currency          string  `json:"currency"`
		ReceivedTimestamp int64   `json:"receivedTimestamp"`
		State             string  `json:"state"`
		TransactionID     string  `json:"transaction_id"`
		UpdatedTimestamp  int64   `json:"updated_timestamp"`
	} `json:"data"`
}

// TransferData stores data for a transfer
type TransferData struct {
	Amount           float64 `json:"amount"`
	CreatedTimestamp int64   `json:"created_timestamp"`
	Currency         string  `json:"currency"`
	Direction        string  `json:"direction"`
	ID               int64   `json:"id"`
	OtherSide        string  `json:"other_side"`
	State            string  `json:"state"`
	Type             string  `json:"type"`
	UpdatedTimestamp int64   `json:"updated_timestamp"`
}

// TransfersData stores data of transfers
type TransfersData struct {
	Count int64          `json:"count"`
	Data  []TransferData `json:"data"`
}

// WithdrawData stores data of withdrawal
type WithdrawData struct {
	Address            string  `json:"address"`
	Amount             float64 `json:"amount"`
	ConfirmedTimestamp int64   `json:"confirmed_timestamp"`
	CreatedTimestamp   int64   `json:"created_timestamp"`
	Currency           string  `json:"currency"`
	Fee                float64 `json:"fee"`
	ID                 int64   `json:"id"`
	Priority           float64 `json:"priority"`
	State              string  `json:"state"`
	TransactionID      string  `json:"transaction_id"`
	UpdatedTimestamp   int64   `json:"updated_timestamp"`
}

// WithdrawalsData stores data of withdrawals
type WithdrawalsData struct {
	Count int64          `json:"count"`
	Data  []WithdrawData `json:"data"`
}

// TradeData stores a data for a private trade
type TradeData struct {
	TradeSequence  int64   `json:"trade_seq"`
	TradeID        int64   `json:"trade_id"`
	Timestamp      int64   `json:"timestamp"`
	TickDirection  int64   `json:"tick_direction"`
	State          string  `json:"state"`
	SelfTrade      bool    `json:"self_trade"`
	ReduceOnly     bool    `json:"reduce_only"`
	Price          float64 `json:"price"`
	PostOnly       bool    `json:"post_only"`
	OrderType      string  `json:"order_type"`
	OrderID        string  `json:"order_id"`
	MatchingID     int64   `json:"matching_id"`
	MarkPrice      float64 `json:"mark_price"`
	Liquidity      string  `json:"liquidity"`
	Label          string  `json:"label"`
	InstrumentName string  `json:"instrument_name"`
	IndexPrice     float64 `json:"index_price"`
	FeeCurrency    string  `json:"fee_currency"`
	Fee            float64 `json:"fee"`
	Direction      string  `json:"direction"`
	Amount         float64 `json:"amount"`
}

// OrderData stores order data
type OrderData struct {
	Web                 bool    `json:"web"`
	TimeInForce         string  `json:"time_in_force"`
	Replaced            bool    `json:"replaced"`
	ReduceOnly          bool    `json:"reduce_only"`
	ProfitLoss          float64 `json:"profit_loss"`
	Price               float64 `json:"price"`
	PostOnly            bool    `json:"post_only"`
	OrderType           string  `json:"order_type"`
	OrderState          string  `json:"order_state"`
	OrderID             string  `json:"order_id"`
	MaxShow             float64 `json:"max_show"`
	LastUpdateTimestamp int64   `json:"last_update_timestamp"`
	Label               string  `json:"label"`
	IsLiquidation       bool    `json:"is_liquidation"`
	InstrumentName      string  `json:"instrument_name"`
	FilledAmount        float64 `json:"filled_amount"`
	Direction           string  `json:"direction"`
	CreationTimestamp   int64   `json:"creation_timestamp"`
	Commission          float64 `json:"commission"`
	AveragePrice        float64 `json:"average_price"`
	API                 bool    `json:"api"`
	Amount              float64 `json:"amount"`
}

// PrivateTradeData stores data of a private buy, sell or edit
type PrivateTradeData struct {
	Trades []TradeData `json:"trades"`
	Order  OrderData   `json:"order"`
}

// PrivateCancelData stores data of a private cancel
type PrivateCancelData struct {
	Triggered           bool    `json:"triggered"`
	Trigger             string  `json:"trigger"`
	TimeInForce         string  `json:"time_in_force"`
	TriggerPrice        float64 `json:"trigger_price"`
	ReduceOnly          bool    `json:"reduce_only"`
	ProfitLoss          float64 `json:"profit_loss"`
	Price               string  `json:"price"`
	PostOnly            bool    `json:"post_only"`
	OrderType           string  `json:"order_type"`
	OrderState          string  `json:"order_state"`
	OrderID             string  `json:"order_id"`
	MaxShow             float64 `json:"max_show"`
	LastUpdateTimestamp int64   `json:"last_update_timestamp"`
	Label               string  `json:"label"`
	IsLiquidation       bool    `json:"is_liquidation"`
	InstrumentName      string  `json:"instrument_name"`
	Direction           string  `json:"direction"`
	CreationTimestamp   int64   `json:"creation_timestamp"`
	API                 bool    `json:"api"`
	Amount              float64 `json:"amount"`
}

// MarginsData stores data for margin
type MarginsData struct {
	Buy      float64 `json:"buy"`
	MaxPrice float64 `json:"max_price"`
	MinPrice float64 `json:"min_price"`
	Sell     float64 `json:"sell"`
}

// MMPConfigData gets the current configuration data for MMP
type MMPConfigData struct {
	Currency      string  `json:"currency"`
	Interval      int64   `json:"interval"`
	FrozenTime    int64   `json:"frozen_time"`
	QuantityLimit float64 `json:"quantity_limit"`
}

// OrderMarginData stores data for order margins
type OrderMarginData struct {
	OrderID       string  `json:"order_id"`
	InitialMargin float64 `json:"initial_margin"`
}

// TriggerOrderData stores data for trigger orders
type TriggerOrderData struct {
	Trigger        string  `json:"trigger"`
	Timestamp      int64   `json:"timestamp"`
	TriggerPrice   float64 `json:"trigger_price"`
	TriggerOrderID string  `json:"trigger_order_id"`
	OrderState     string  `json:"order_state"`
	Request        string  `json:"request"`
	Price          float64 `json:"price"`
	OrderID        int64   `json:"order_id"`
	Offset         int64   `json:"offset"`
	InstrumentName string  `json:"instrument_name"`
	Amount         float64 `json:"amount"`
	Direction      string  `json:"direction"`
}

// UserTradesData stores data of user trades
type UserTradesData struct {
	Trades  []UserTradeData `json:"trades"`
	HasMore bool            `json:"has_more"`
}

// UserTradeData stores data of user trades
type UserTradeData struct {
	UnderlyingPrice float64 `json:"underlying_price"`
	TradeSequence   int64   `json:"trade_sequence"`
	TradeID         int64   `json:"trade_id"`
	Timestamp       int64   `json:"timestamp"`
	TickDirection   int64   `json:"tick_direction"`
	State           string  `json:"state"`
	SelfTrade       bool    `json:"self_trade"`
	ReduceOnly      bool    `json:"reduce_only"`
	Price           float64 `json:"price"`
	PostOnly        bool    `json:"post_only"`
	OrderType       string  `json:"order_type"`
	OrderID         int64   `json:"order_id"`
	MatchingID      int64   `json:"matching_id"`
	MarkPrice       float64 `json:"mark_price"`
	Liquidity       string  `json:"liquidity"`
	IV              float64 `json:"iv"`
	InstrumentName  string  `json:"instrument_name"`
	IndexPrice      float64 `json:"index_price"`
	FeeCurrency     string  `json:"fee_currency"`
	Fee             float64 `json:"fee"`
	Direction       string  `json:"direction"`
	Amount          float64 `json:"amount"`
}

// PrivateSettlementsHistoryData stores data for private settlement history
type PrivateSettlementsHistoryData struct {
	Settlements  []PrivateSettlementData `json:"settlements"`
	Continuation string                  `json:"continuation"`
}

// PrivateSettlementData stores private settlement data
type PrivateSettlementData struct {
	Type              string  `json:"type"`
	Timestamp         int64   `json:"timestamp"`
	SessionProfitLoss float64 `json:"session_profit_loss"`
	ProfitLoss        float64 `json:"profit_loss"`
	Position          float64 `json:"position"`
	MarkPrice         float64 `json:"mark_price"`
	InstrumentName    string  `json:"instrument_name"`
	IndexPrice        float64 `json:"index_price"`
}

// AccountSummaryData stores data of account summary for a given currency
type AccountSummaryData struct {
	Balance                  float64 `json:"balance"`
	OptionsSessionUPL        float64 `json:"options_session_upl"`
	DepositAddress           string  `json:"deposit_address"`
	OptionsGamma             float64 `json:"options_gamma"`
	OptionsTheta             float64 `json:"options_theta"`
	Username                 string  `json:"username"`
	Equity                   float64 `json:"equity"`
	Type                     string  `json:"type"`
	Currency                 string  `json:"currency"`
	DeltaTotal               float64 `json:"delta_total"`
	FuturesSessionRPL        float64 `json:"futures_session_rpl"`
	PortfolioManagingEnabled bool    `json:"portfolio_managing_enabled"`
	TotalPL                  float64 `json:"total_pl"`
	MarginBalance            float64 `json:"margin_balance"`
	TFAEnabled               bool    `json:"tfa_enabled"`
	OptionsSessionRPL        float64 `json:"options_session_rpl"`
	OptionsDelta             float64 `json:"options_delta"`
	FuturesPL                float64 `json:"futures_pl"`
	ReferrerID               string  `json:"referrer_id"`
	ID                       int64   `json:"id"`
	SessionUPL               float64 `json:"session_upl"`
	AvailableWithdrawalFunds float64 `json:"available_withdrawal_funds"`
	CreationTimestamp        int64   `json:"creation_timestamp"`
	OptionsPL                float64 `json:"options_pl"`
	SystemName               string  `json:"system_name"`
	Limits                   struct {
		NonMatchingEngine struct {
			Rate  int64 `json:"rate"`
			Burst int64 `json:"burst"`
		} `json:"non_matching_engine"`
		MatchingEngine struct {
			Rate  int64 `json:"rate"`
			Burst int64 `json:"burst"`
		} `json:"matching_engine"`
	} `json:"limits"`
	InitialMargin             float64 `json:"initial_margin"`
	ProjectedInitialMargin    float64 `json:"projected_initial_margin"`
	MaintenanceMargin         float64 `json:"maintenance_margin"`
	SessionRPL                float64 `json:"session_rpl"`
	InteruserTransfersEnabled bool    `json:"interuser_transfers_enabled"`
	OptionsVega               float64 `json:"options_vega"`
	ProjectedDeltaTotal       float64 `json:"projected_delta_total"`
	Email                     string  `json:"email"`
	FuturesSessionUPL         float64 `json:"futures_session_upl"`
	AvailableFunds            float64 `json:"available_funds"`
	OptionsValue              float64 `json:"options_value"`
}

// APIKeyData stores data regarding the api key
type APIKeyData struct {
	Timestamp    int64  `json:"timestamp"`
	MaxScope     string `json:"max_scope"`
	ID           int64  `json:"id"`
	Enabled      bool   `json:"enabled"`
	Default      bool   `json:"default"`
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	Name         string `json:"name"`
}

// SubAccountData stores stores subaccount data
type SubAccountData struct {
	Email        string `json:"email"`
	ID           int64  `json:"id"`
	IsPassword   bool   `json:"is_password"`
	LoginEnabled bool   `json:"login_enabled"`
	Portfolio    struct {
		Eth struct {
			AvailableFunds           float64 `json:"available_funds"`
			AvailableWithdrawalFunds float64 `json:"available_withdrawal_funds"`
			Balance                  float64 `json:"balance"`
			Currency                 string  `json:"currency"`
			Equity                   float64 `json:"equity"`
			InitialMargin            float64 `json:"initial_margin"`
			MaintenanceMargin        float64 `json:"maintenance_margin"`
			MarginBalance            float64 `json:"margin_balance"`
		} `json:"eth"`
		Btc struct {
			AvailableFunds           float64 `json:"available_funds"`
			AvailableWithdrawalFunds float64 `json:"available_withdrawal_funds"`
			Balance                  float64 `json:"balance"`
			Currency                 string  `json:"currency"`
			Equity                   float64 `json:"equity"`
			InitialMargin            float64 `json:"initial_margin"`
			MaintenanceMargin        float64 `json:"maintenance_margin"`
			MarginBalance            float64 `json:"margin_balance"`
		} `json:"btc"`
	}
	ReceiveNotifications bool   `json:"receive_notifications"`
	SystemName           string `json:"system_name"`
	TFAEnabled           bool   `json:"tfa_enabled"`
	Type                 string `json:"type"`
	Username             string `json:"username"`
}

// AffiliateProgramInfo stores info of affiliate program
type AffiliateProgramInfo struct {
	Received struct {
		Eth float64 `json:"eth"`
		Btc float64 `json:"btc"`
	} `json:"received"`
	NumberOfAffiliates int64  `json:"number_of_affiliates"`
	Link               string `json:"link"`
	IsEnabled          bool   `json:"is_enabled"`
}

// PrivateAnnouncementsData stores data of private announcements
type PrivateAnnouncementsData struct {
	Title                string `json:"title"`
	PublicationTimestamp int64  `json:"publication_timestamp"`
	Important            bool   `json:"important"`
	ID                   int64  `json:"id"`
	Body                 string `json:"body"`
}

// PositionData stores data for account's position
type PositionData struct {
	AveragePrice              float64 `json:"average_price"`
	Delta                     float64 `json:"delta"`
	Direction                 string  `json:"direction"`
	EstimatedLiquidationPrice float64 `json:"estimated_liquidation_price"`
	FloatingProfitLoss        float64 `json:"floating_profit_loss"`
	IndexPrice                float64 `json:"index_price"`
	InitialMargin             float64 `json:"initial_margin"`
	InstrumentName            string  `json:"instrument_name"`
	Leverage                  float64 `json:"leverage"`
	Kind                      string  `json:"kind"`
	MaintenanceMargin         float64 `json:"maintenance_margin"`
	MarkPrice                 float64 `json:"mark_price"`
	OpenOrdersMargin          float64 `json:"open_orders_margin"`
	RealizedProfitLoss        float64 `json:"realized_profit_loss"`
	SettlementPrice           float64 `json:"settlement_price"`
	Size                      float64 `json:"size"`
	SizeCurrency              float64 `json:"size_currency"`
	TotalProfitLoss           float64 `json:"total_profit_loss"`
}

// TransactionLogData stores information regarding an account transaction
type TransactionLogData struct {
	Username        string  `json:"username"`
	UserSeq         int64   `json:"user_seq"`
	UserID          int64   `json:"user_id"`
	TransactionType string  `json:"transaction_type"`
	TradeID         int64   `json:"trade_id"`
	Timestamp       int64   `json:"timestamp"`
	Side            string  `json:"side"`
	Price           float64 `json:"price"`
	Position        float64 `json:"position"`
	OrderID         int64   `json:"order_id"`
	InterestPL      float64 `json:"interest_pl"`
	InstrumentName  string  `json:"instrument_name"`
	Info            struct {
		TransferType string `json:"transfer_type"`
		OtherUserID  int64  `json:"other_user_id"`
		OtherUser    string `json:"other_user"`
	} `json:"info"`
	ID         int64   `json:"id"`
	Equity     float64 `json:"equity"`
	Currency   string  `json:"currency"`
	Commission float64 `json:"commission"`
	Change     float64 `json:"change"`
	Cashflow   float64 `json:"cashflow"`
	Balance    float64 `json:"balance"`
}

// TransactionsData stores multiple transaction logs
type TransactionsData struct {
	Logs         []TransactionLogData `json:"logs"`
	Continuation int64                `json:"continuation"`
}

// PlaceTradeData stores data of a private trade/order
type PlaceTradeData struct {
	Trades []PrivateTradeData `json:"trades"`
	Order  OrderData          `json:"order"`
}

// WsRequest defines a request obj for the JSON-RPC and gets a websocket
// response
type WsRequest struct {
	JSONRPCVersion string                 `json:"jsonrpc,omitempty"`
	ID             int64                  `json:"id,omitempty"`
	Method         string                 `json:"method"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

type wsResponse struct {
	JSONRPCVersion string `json:"jsonrpc,omitempty"`
	ID             int64  `json:"id,omitempty"`
}

type wsSubmitOrderResponse struct {
	JSONRPCVersion string            `json:"jsonrpc"`
	ID             int64             `json:"id"`
	Method         string            `json:"method"`
	Result         *PrivateTradeData `json:"result"`
	Error          *UnmarshalError   `json:"error"`
}

type wsLoginResponse struct {
	JSONRPCVersion string                 `json:"jsonrpc"`
	ID             int64                  `json:"id"`
	Method         string                 `json:"method"`
	Result         map[string]interface{} `json:"result"`
	Error          *UnmarshalError        `json:"error"`
}

// RFQ RFQs for instruments in given currency.
type RFQ struct {
	TradedVolume     float64   `json:"traded_volume"`
	Amount           float64   `json:"amount"`
	Side             string    `json:"side"`
	LastRfqTimestamp time.Time `json:"last_rfq_tstamp"`
	InstrumentName   string    `json:"instrument_name"`
}

// ComboDetail retrieves information about a combo
type ComboDetail struct {
	ID                string    `json:"id"`
	InstrumentID      int64     `json:"instrument_id"`
	CreationTimestamp time.Time `json:"creation_timestamp"`
	StateTimestamp    int64     `json:"state_timestamp"`
	State             string    `json:"state"`
	Legs              []struct {
		InstrumentName string  `json:"instrument_name"`
		Amount         float64 `json:"amount"`
	} `json:"legs"`
}

// ComboParam represents a parameter to sell and buy combo.
type ComboParam struct {
	InstrumentName string  `json:"instrument_name"`
	Direction      string  `json:"direction"`
	Amount         float64 `json:"amount,string"`
}

// BlockTradeParam represents a block trade parameter.
type BlockTradeParam struct {
	Price          float64 `json:"price"`
	InstrumentName string  `json:"instrument_name"`
	Direction      string  `json:"direction,omitempty"`
	Amount         float64 `json:"amount"`
}

// BlockTradeData represents a user's block trade data.
type BlockTradeData struct {
	TradeSeq               int64       `json:"trade_seq"`
	TradeID                string      `json:"trade_id"`
	Timestamp              int64       `json:"timestamp"`
	TickDirection          int64       `json:"tick_direction"`
	State                  string      `json:"state"`
	SelfTrade              bool        `json:"self_trade"`
	Price                  float64     `json:"price"`
	OrderType              string      `json:"order_type"`
	OrderID                string      `json:"order_id"`
	MatchingID             interface{} `json:"matching_id"`
	Liquidity              string      `json:"liquidity"`
	OptionmpliedVolatility float64     `json:"iv,omitempty"`
	InstrumentName         string      `json:"instrument_name"`
	IndexPrice             float64     `json:"index_price"`
	FeeCurrency            string      `json:"fee_currency"`
	Fee                    float64     `json:"fee"`
	Direction              string      `json:"direction"`
	BlockTradeID           string      `json:"block_trade_id"`
	Amount                 float64     `json:"amount"`
}

// Announcement represents public announcements.
type Announcement struct {
	Title                string    `json:"title"`
	PublicationTimestamp time.Time `json:"publication_timestamp"`
	Important            bool      `json:"important"`
	ID                   int64     `json:"id"`
	Body                 string    `json:"body"`
}

// PortfolioMargin represents public portfolio margins.
type PortfolioMargin struct {
	VolumeRange         []float64          `json:"vol_range"`
	VegaPow2            float64            `json:"vega_pow2"`
	VegaPow1            float64            `json:"vega_pow1"`
	Skew                float64            `json:"skew"`
	PriceRange          float64            `json:"price_range"`
	OptSumContinguency  float64            `json:"opt_sum_continguency"`
	OptContinguency     float64            `json:"opt_continguency"`
	Kurtosis            float64            `json:"kurtosis"`
	IntRate             float64            `json:"int_rate"`
	InitialMarginFactor float64            `json:"initial_margin_factor"`
	FtuContinguency     float64            `json:"ftu_continguency"`
	AtmRange            float64            `json:"atm_range"`
	ProjectedMarginPos  float64            `json:"projected_margin_pos"`
	ProjectedMargin     float64            `json:"projected_margin"`
	PositionSizes       map[string]float64 `json:"position_sizes"`
	Pls                 []float64          `json:"pls"`
	PcoOpt              float64            `json:"pco_opt"`
	PcoFtu              float64            `json:"pco_ftu"`
	OptSummary          []interface{}      `json:"opt_summary"`
	OptPls              []float64          `json:"opt_pls"`
	OptEntries          []interface{}      `json:"opt_entries"`
	MarginPos           float64            `json:"margin_pos"`
	Margin              float64            `json:"margin"`
	FtuSummary          []struct {
		ShortTotalCost  float64   `json:"short_total_cost"`
		PlVec           []float64 `json:"pl_vec"`
		LongTotalCost   float64   `json:"long_total_cost"`
		ExpiryTimestamp int64     `json:"exp_tstamp"`
	} `json:"ftu_summary"`
	FtuPls     []float64 `json:"ftu_pls"`
	FtuEntries []struct {
		TotalCost       float64   `json:"total_cost"`
		Size            float64   `json:"size"`
		PlVec           []float64 `json:"pl_vec"`
		MarkPrice       float64   `json:"mark_price"`
		InstrumentName  string    `json:"instrument_name"`
		ExpiryTimestamp int64     `json:"exp_tstamp"`
	} `json:"ftu_entries"`
	CoOpt                float64 `json:"co_opt"`
	CoFtu                float64 `json:"co_ftu"`
	CalculationTimestamp int64   `json:"calculation_timestamp"`
}

// AccessLog represents access log information.
type AccessLog struct {
	RecordsTotal int               `json:"records_total"`
	Data         []AccessLogDetail `json:"data"`
}

// AccessLogDetail represents detailed access log information.
type AccessLogDetail struct {
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"result"`
	IP        string    `json:"ip"`
	ID        int64     `json:"id"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
}

// SubAccountDetail represents subaccount positions detail.
type SubAccountDetail struct {
	UID       int `json:"uid"`
	Positions []struct {
		TotalProfitLoss           float64 `json:"total_profit_loss"`
		SizeCurrency              float64 `json:"size_currency"`
		Size                      float64 `json:"size"`
		SettlementPrice           float64 `json:"settlement_price"`
		RealizedProfitLoss        float64 `json:"realized_profit_loss"`
		RealizedFunding           float64 `json:"realized_funding"`
		OpenOrdersMargin          float64 `json:"open_orders_margin"`
		MarkPrice                 float64 `json:"mark_price"`
		MaintenanceMargin         float64 `json:"maintenance_margin"`
		Leverage                  float64 `json:"leverage"`
		Kind                      string  `json:"kind"`
		InstrumentName            string  `json:"instrument_name"`
		InitialMargin             float64 `json:"initial_margin"`
		IndexPrice                float64 `json:"index_price"`
		FloatingProfitLoss        float64 `json:"floating_profit_loss"`
		EstimatedLiquidationPrice float64 `json:"estimated_liquidation_price"`
		Direction                 string  `json:"direction"`
		Delta                     float64 `json:"delta"`
		AveragePrice              float64 `json:"average_price"`
	} `json:"positions"`
}

// UserLock represents a user lock information for currency.
type UserLock struct {
	Message  string `json:"message"`
	Locked   bool   `json:"locked"`
	Currency string `json:"currency"`
}

// TogglePortfolioMarginResponse represents a response from toggling portfolio margin for currency.
type TogglePortfolioMarginResponse struct {
	OldState struct {
		MaintenanceMarginRate float64 `json:"maintenance_margin_rate"`
		InitialMarginRate     float64 `json:"initial_margin_rate"`
		AvailableBalance      float64 `json:"available_balance"`
	} `json:"old_state"`
	NewState struct {
		MaintenanceMarginRate float64 `json:"maintenance_margin_rate"`
		InitialMarginRate     float64 `json:"initial_margin_rate"`
		AvailableBalance      float64 `json:"available_balance"`
	} `json:"new_state"`
	Currency string `json:"currency"`
}

// BlockTradeResponse represents a block trade response.
type BlockTradeResponse struct {
	TradeSeq       int64     `json:"trade_seq"`
	TradeID        string    `json:"trade_id"`
	Timestamp      time.Time `json:"timestamp"`
	TickDirection  int64     `json:"tick_direction"`
	State          string    `json:"state"`
	SelfTrade      bool      `json:"self_trade"`
	ReduceOnly     bool      `json:"reduce_only"`
	Price          float64   `json:"price"`
	PostOnly       bool      `json:"post_only"`
	OrderType      string    `json:"order_type"`
	OrderID        string    `json:"order_id"`
	MatchingID     string    `json:"matching_id"`
	MarkPrice      float64   `json:"mark_price"`
	Liquidity      string    `json:"liquidity"`
	InstrumentName string    `json:"instrument_name"`
	IndexPrice     float64   `json:"index_price"`
	FeeCurrency    string    `json:"fee_currency"`
	Fee            float64   `json:"fee"`
	Direction      string    `json:"direction"`
	BlockTradeID   string    `json:"block_trade_id"`
	Amount         float64   `json:"amount"`
}

// BlockTradeMoveResponse represents block trade move response.
type BlockTradeMoveResponse struct {
	TargetSubAccountUID int64   `json:"target_uid"`
	SourceSubAccountUID int64   `json:"source_uid"`
	Price               float64 `json:"price"`
	InstrumentName      string  `json:"instrument_name"`
	Direction           string  `json:"direction"`
	Amount              float64 `json:"amount"`
}

// TFAChallenge represents response to Remove API Key.
type TFAChallenge struct {
	Challenge                        string `json:"challenge"`
	RpID                             string `json:"rp_id"`
	SecurityKeyAuthorizationRequired bool   `json:"security_key_authorization_required"`
	SecurityKeys                     []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"security_keys"`
}
