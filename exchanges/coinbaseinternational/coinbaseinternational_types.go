package coinbaseinternational

import (
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/types"
)

// AssetItemInfo represents a single an asset item instance.
type AssetItemInfo struct {
	AssetID                  string  `json:"asset_id"`
	AssetUUID                string  `json:"asset_uuid"`
	AssetName                string  `json:"asset_name"`
	Status                   string  `json:"status"`
	CollateralWeight         float64 `json:"collateral_weight"`
	SupportedNetworksEnabled bool    `json:"supported_networks_enabled"`
}

// AssetInfoWithSupportedNetwork represents network information for a specific asset.
type AssetInfoWithSupportedNetwork struct {
	AssetID             string               `json:"asset_id"`
	AssetUUID           string               `json:"asset_uuid"`
	AssetName           string               `json:"asset_name"`
	NetworkName         string               `json:"network_name"`
	DisplayName         string               `json:"display_name"`
	NetworkArnID        string               `json:"network_arn_id"`
	MinWithdrawalAmount types.Number         `json:"min_withdrawal_amt"`
	MaxWithdrawalAmount types.Number         `json:"max_withdrawal_amt"`
	NetworkConfirms     int64                `json:"network_confirms"`
	ProcessingTime      convert.ExchangeTime `json:"processing_time"`
	IsDefault           bool                 `json:"is_default"`
}

// InstrumentInfo represents an instrument detail for specific instrument id.
type InstrumentInfo struct {
	InstrumentID        string       `json:"instrument_id"`
	InstrumentUUID      string       `json:"instrument_uuid"`
	Symbol              string       `json:"symbol"`
	Type                string       `json:"type"`
	BaseAssetID         string       `json:"base_asset_id"`
	BaseAssetUUID       string       `json:"base_asset_uuid"`
	BaseAssetName       string       `json:"base_asset_name"`
	QuoteAssetID        string       `json:"quote_asset_id"`
	QuoteAssetUUID      string       `json:"quote_asset_uuid"`
	QuoteAssetName      string       `json:"quote_asset_name"`
	BaseIncrement       types.Number `json:"base_increment"`
	QuoteIncrement      types.Number `json:"quote_increment"`
	PriceBandPercent    float64      `json:"price_band_percent"`
	MarketOrderPercent  float64      `json:"market_order_percent"`
	Qty24Hr             types.Number `json:"qty_24hr"`
	Notional24Hr        types.Number `json:"notional_24hr"`
	AvgDailyQty         types.Number `json:"avg_daily_qty"`
	AvgDailyNotional    types.Number `json:"avg_daily_notional"`
	PreviousDayQty      types.Number `json:"previous_day_qty"`
	OpenInterest        types.Number `json:"open_interest"`
	PositionLimitQty    types.Number `json:"position_limit_qty"`
	PositionLimitAdqPct float64      `json:"position_limit_adq_pct"`
	ReplacementCost     types.Number `json:"replacement_cost"`
	BaseImf             float64      `json:"base_imf"`
	MinNotionalValue    string       `json:"min_notional_value"`
	FundingInterval     string       `json:"funding_interval"`
	TradingState        string       `json:"trading_state"`
	PositionLimitAdv    float64      `json:"position_limit_adv"`
	InitialMarginAdv    float64      `json:"initial_margin_adv"`
	Quote               struct {
		BestBidPrice     types.Number `json:"best_bid_price"`
		BestBidSize      types.Number `json:"best_bid_size"`
		BestAskPrice     types.Number `json:"best_ask_price"`
		BestAskSize      types.Number `json:"best_ask_size"`
		TradePrice       types.Number `json:"trade_price"`
		TradeQty         types.Number `json:"trade_qty"`
		IndexPrice       types.Number `json:"index_price"`
		MarkPrice        types.Number `json:"mark_price"`
		SettlementPrice  types.Number `json:"settlement_price"`
		LimitUp          types.Number `json:"limit_up"`
		LimitDown        types.Number `json:"limit_down"`
		PredictedFunding types.Number `json:"predicted_funding"`
		Timestamp        time.Time    `json:"timestamp"`
	} `json:"quote"`
}

// QuoteInformation represents a instrument quote information
type QuoteInformation struct {
	BestBidPrice     types.Number `json:"best_bid_price"`
	BestBidSize      types.Number `json:"best_bid_size"`
	BestAskPrice     types.Number `json:"best_ask_price"`
	BestAskSize      types.Number `json:"best_ask_size"`
	TradePrice       types.Number `json:"trade_price"`
	TradeQty         types.Number `json:"trade_qty"`
	IndexPrice       types.Number `json:"index_price"`
	MarkPrice        types.Number `json:"mark_price"`
	SettlementPrice  types.Number `json:"settlement_price"`
	LimitUp          types.Number `json:"limit_up"`
	LimitDown        types.Number `json:"limit_down"`
	PredictedFunding types.Number `json:"predicted_funding"`
	Timestamp        time.Time    `json:"timestamp"`
}

// OrderRequestParams represents a request parameter for creating order.
type OrderRequestParams struct {
	ClientOrderID string  `json:"client_order_id"`
	Side          string  `json:"side,omitempty"`
	BaseSize      float64 `json:"size,omitempty,string"`
	TimeInForce   string  `json:"tif,omitempty"`        // Possible values: [GTC, IOC, GTT]
	Instrument    string  `json:"instrument,omitempty"` // The name, ID, or UUID of the instrument the order wants to transact
	OrderType     string  `json:"type,omitempty"`
	Price         float64 `json:"price,omitempty,string"`
	StopPrice     float64 `json:"stop_price,omitempty,string"`
	ExpireTime    string  `json:"expire_time,omitempty"` // e.g., 2023-03-16T23:59:53Z
	Portfolio     string  `json:"portfolio,omitempty"`
	User          string  `json:"user,omitempty"`     // The ID or UUID of the user the order belongs to (only used and required for brokers)
	STPMode       string  `json:"stp_mode,omitempty"` // Possible values: [NONE, AGGRESSING, BOTH]
	PostOnly      bool    `json:"post_only,omitempty"`
}

// TradeOrder represents a single order
type TradeOrder struct {
	OrderID        int64        `json:"order_id"`
	ClientOrderID  string       `json:"client_order_id"`
	Side           string       `json:"side"`
	InstrumentID   int64        `json:"instrument_id"`
	InstrumentUUID string       `json:"instrument_uuid"`
	Symbol         string       `json:"symbol"`
	PortfolioID    int64        `json:"portfolio_id"`
	PortfolioUUID  string       `json:"portfolio_uuid"`
	Type           string       `json:"type"`
	Price          float64      `json:"price"`
	StopPrice      float64      `json:"stop_price"`
	Size           float64      `json:"size"`
	TimeInForce    string       `json:"tif"`
	ExpireTime     time.Time    `json:"expire_time"`
	StpMode        string       `json:"stp_mode"`
	EventType      string       `json:"event_type"`
	OrderStatus    string       `json:"order_status"`
	LeavesQty      types.Number `json:"leaves_qty"`
	ExecQty        types.Number `json:"exec_qty"`
	AvgPrice       types.Number `json:"avg_price"`
	Message        string       `json:"message"`
	Fee            types.Number `json:"fee"`
}

// OrderItemDetail represents an open order detail.
type OrderItemDetail struct {
	Pagination struct {
		RefDatetime  time.Time `json:"ref_datetime"`
		ResultLimit  int64     `json:"result_limit"`
		ResultOffset int64     `json:"result_offset"`
	} `json:"pagination"`
	Results []OrderItem `json:"results"`
}

// ModifyOrderParam holds update parameters to modify an order.
type ModifyOrderParam struct {
	ClientOrderID string  `json:"client_order_id,omitempty"`
	Portfolio     string  `json:"portfolio,omitempty"`
	Price         float64 `json:"price,omitempty,string"`
	StopPrice     float64 `json:"stop_price,omitempty,string"`
	Size          float64 `json:"size,omitempty,string"`
}

// OrderItem represents a single order item.
type OrderItem struct {
	OrderID        string       `json:"order_id"`
	ClientOrderID  string       `json:"client_order_id"`
	Side           string       `json:"side"`
	InstrumentID   string       `json:"instrument_id"`
	InstrumentUUID string       `json:"instrument_uuid"`
	Symbol         string       `json:"symbol"`
	PortfolioID    int64        `json:"portfolio_id"`
	PortfolioUUID  string       `json:"portfolio_uuid"`
	Type           string       `json:"type"`
	Price          float64      `json:"price"`
	StopPrice      float64      `json:"stop_price"`
	Size           float64      `json:"size"`
	TimeInForce    string       `json:"tif"`
	ExpireTime     time.Time    `json:"expire_time"`
	StpMode        string       `json:"stp_mode"`
	EventType      string       `json:"event_type"`
	OrderStatus    string       `json:"order_status"`
	LeavesQuantity types.Number `json:"leaves_qty"`
	ExecQty        types.Number `json:"exec_qty"`
	AveragePrice   types.Number `json:"avg_price"`
	Message        string       `json:"message"`
	Fee            types.Number `json:"fee"`
}

// PortfolioItem represents a user portfolio item
// and transaction fee information.
type PortfolioItem struct {
	PortfolioID    string       `json:"portfolio_id"`
	PortfolioUUID  string       `json:"portfolio_uuid"`
	Name           string       `json:"name"`
	UserUUID       string       `json:"user_uuid"`
	MakerFeeRate   types.Number `json:"maker_fee_rate"`
	TakerFeeRate   types.Number `json:"taker_fee_rate"`
	TradingLock    bool         `json:"trading_lock"`
	BorrowDisabled bool         `json:"borrow_disabled"`
	IsLSP          bool         `json:"is_lsp"` // Indicates if the portfolio is setup to take liquidation assignments
}

// PortfolioDetail represents a portfolio detail.
type PortfolioDetail struct {
	Summary struct {
		Collateral             float64 `json:"collateral"`
		UnrealizedPnl          float64 `json:"unrealized_pnl"`
		PositionNotional       float64 `json:"position_notional"`
		OpenPositionNotional   float64 `json:"open_position_notional"`
		PendingFees            float64 `json:"pending_fees"`
		Borrow                 float64 `json:"borrow"`
		AccruedInterest        float64 `json:"accrued_interest"`
		RollingDebt            float64 `json:"rolling_debt"`
		Balance                float64 `json:"balance"`
		BuyingPower            float64 `json:"buying_power"`
		PortfolioCurrentMargin float64 `json:"portfolio_current_margin"`
		PortfolioInitialMargin float64 `json:"portfolio_initial_margin"`
		InLiquidation          string  `json:"in_liquidation"`
	} `json:"summary"`
	Balances []struct {
		AssetID           string  `json:"asset_id"`
		AssetUUID         string  `json:"asset_uuid"`
		AssetName         string  `json:"asset_name"`
		Quantity          float64 `json:"quantity"`
		Hold              float64 `json:"hold"`
		TransferHold      float64 `json:"transfer_hold"`
		CollateralValue   float64 `json:"collateral_value"`
		MaxWithdrawAmount float64 `json:"max_withdraw_amount"`
	} `json:"balances"`
	Positions []struct {
		InstrumentID   string  `json:"instrument_id"`
		InstrumentUUID string  `json:"instrument_uuid"`
		Symbol         string  `json:"symbol"`
		Vwap           float64 `json:"vwap"`
		NetSize        float64 `json:"net_size"`
		BuyOrderSize   float64 `json:"buy_order_size"`
		SellOrderSize  float64 `json:"sell_order_size"`
		ImContribution float64 `json:"im_contribution"`
		UnrealizedPnl  float64 `json:"unrealized_pnl"`
		MarkPrice      float64 `json:"mark_price"`
	} `json:"positions"`
}

// PortfolioSummary represents a portfolio summary detailed instance.
type PortfolioSummary struct {
	Collateral             float64 `json:"collateral"`
	UnrealizedPnl          float64 `json:"unrealized_pnl"`
	PositionNotional       float64 `json:"position_notional"`
	OpenPositionNotional   float64 `json:"open_position_notional"`
	PendingFees            float64 `json:"pending_fees"`
	Borrow                 float64 `json:"borrow"`
	AccruedInterest        float64 `json:"accrued_interest"`
	RollingDebt            float64 `json:"rolling_debt"`
	Balance                float64 `json:"balance"`
	BuyingPower            float64 `json:"buying_power"`
	PortfolioCurrentMargin float64 `json:"portfolio_current_margin"`
	PortfolioInitialMargin float64 `json:"portfolio_initial_margin"`
	InLiquidation          string  `json:"in_liquidation"`
}

// PortfolioBalance represents a portfolio balance instance.
type PortfolioBalance struct {
	AssetID           string       `json:"asset_id"`
	AssetUUID         string       `json:"asset_uuid"`
	AssetName         string       `json:"asset_name"`
	Quantity          types.Number `json:"quantity"`
	Hold              types.Number `json:"hold"`
	TransferHold      types.Number `json:"transfer_hold"`
	CollateralValue   types.Number `json:"collateral_value"`
	MaxWithdrawAmount types.Number `json:"max_withdraw_amount"`
}

// PortfolioPosition represents a portfolio positions instance.
type PortfolioPosition struct {
	InstrumentID              string  `json:"instrument_id"`
	InstrumentUUID            string  `json:"instrument_uuid"`
	Symbol                    string  `json:"symbol"`
	Vwap                      float64 `json:"vwap"`
	NetSize                   float64 `json:"net_size"`
	BuyOrderSize              float64 `json:"buy_order_size"`
	SellOrderSize             float64 `json:"sell_order_size"`
	InitialMarginContribution float64 `json:"im_contribution"`
	UnrealizedPnl             float64 `json:"unrealized_pnl"`
	MarkPrice                 float64 `json:"mark_price"`
}

// PortfolioFill represents a portfolio fill information.
type PortfolioFill struct {
	Pagination struct {
		RefDatetime  time.Time `json:"ref_datetime"`
		ResultLimit  int64     `json:"result_limit"`
		ResultOffset int64     `json:"result_offset"`
	} `json:"pagination"`
	Results []struct {
		FillID         string    `json:"fill_id"`
		OrderID        string    `json:"order_id"`
		InstrumentID   string    `json:"instrument_id"`
		InstrumentUUID string    `json:"instrument_uuid"`
		Symbol         string    `json:"symbol"`
		MatchID        string    `json:"match_id"`
		FillPrice      float64   `json:"fill_price"`
		FillQty        float64   `json:"fill_qty"`
		ClientID       string    `json:"client_id"`
		ClientOrderID  string    `json:"client_order_id"`
		OrderQty       float64   `json:"order_qty"`
		LimitPrice     float64   `json:"limit_price"`
		TotalFilled    float64   `json:"total_filled"`
		FilledVwap     float64   `json:"filled_vwap"`
		ExpireTime     time.Time `json:"expire_time"`
		StopPrice      float64   `json:"stop_price"`
		Side           string    `json:"side"`
		TimeInForce    string    `json:"tif"`
		StpMode        string    `json:"stp_mode"`
		Flags          string    `json:"flags"`
		Fee            float64   `json:"fee"`
		FeeAsset       string    `json:"fee_asset"`
		OrderStatus    string    `json:"order_status"`
		EventTime      time.Time `json:"event_time"`
	} `json:"results"`
}

// Transfers returns a list of fund transfers.
type Transfers struct {
	Pagination struct {
		ResultLimit  int64 `json:"result_limit"`
		ResultOffset int64 `json:"result_offset"`
	} `json:"pagination"`
	Results []FundTransfer `json:"results"`
}

// FundTransfer represents a fund transfer instance.
type FundTransfer struct {
	TransferUUID   string    `json:"transfer_uuid"`
	TransferType   string    `json:"type"`
	Amount         float64   `json:"amount"`
	Asset          string    `json:"asset"`
	TransferStatus string    `json:"status"`
	NetworkName    string    `json:"network_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	FromPortfolio  struct {
		ID   string `json:"id"`
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"from_portfolio"`
	ToPortfolio struct {
		ID   string `json:"id"`
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"to_portfolio"`
	FromAddress         int64  `json:"from_address"`
	ToAddress           int64  `json:"to_address"`
	FromCoinbaseAccount string `json:"from_cb_account"`
	ToCoinbaseAccount   string `json:"to_cb_account"`
}

// WithdrawToCoinbaseINTXParam holds withdraw funds parameters.
type WithdrawToCoinbaseINTXParam struct {
	ProfileID         string        `json:"profile_id"`
	Amount            float64       `json:"amount,omitempty,string"`
	CoinbaseAccountID string        `json:"coinbase_account_id,omitempty"`
	Currency          currency.Code `json:"currency"`
}

// WithdrawToCoinbaseResponse represents a response after withdrawing to coinbase account.
type WithdrawToCoinbaseResponse struct {
	ID       string       `json:"id"`
	Amount   types.Number `json:"amount"`
	Fee      types.Number `json:"fee"`
	Currency string       `json:"currency"`
	PayoutAt string       `json:"payout_at"`
	Subtotal string       `json:"subtotal"`
}

// WithdrawCryptoParams holds crypto fund withdrawal information.
type WithdrawCryptoParams struct {
	Portfolio            string  `json:"portfolio,omitempty"` // Identifies the portfolio by UUID
	AssetIdentifier      string  `json:"asset"`               // Identifies the asset by name
	Amount               float64 `json:"amount,string"`
	AddNetworkFeeToTotal bool    `json:"add_network_fee_to_total,omitempty"` // if true, deducts network fee from the portfolio, otherwise deduct fee from the withdrawal
	NetworkArnID         string  `json:"network_arn_id,omitempty"`           // Identifies the blockchain network
	Address              string  `json:"address"`
	Nonce                string  `json:"nonce,omitempty"`
}

// WithdrawalResponse holds crypto withdrawal ID information
type WithdrawalResponse struct {
	Idem string `json:"idem"` // Idempotent uuid representing the successful withdraw
}

// CryptoAddressParam holds crypto address creation parameters.
type CryptoAddressParam struct {
	Portfolio       string `json:"portfolio"`      // Identifies the portfolio by UUID
	AssetIdentifier string `json:"asset"`          // Identifies the asset by name (e.g., BTC), UUID (e.g., 291efb0f-2396-4d41-ad03-db3b2311cb2c), or asset ID (e.g., 1482439423963469)
	NetworkArnID    string `json:"network_arn_id"` // Identifies the blockchain network
}

// CryptoAddressInfo holds crypto address information after creation
type CryptoAddressInfo struct {
	Address      string `json:"address"`
	NetworkArnID string `json:"network_arn_id"`
}

// SubscriptionInput holds channel subscription information
type SubscriptionInput struct {
	Type           string         `json:"type"` // SUBSCRIBE or UNSUBSCRIBE
	ProductIDPairs currency.Pairs `json:"-"`
	ProductIDs     []string       `json:"product_ids"`
	Channels       []string       `json:"channels"`
	Time           string         `json:"time,omitempty"`
	Key            string         `json:"key,omitempty"`
	Passphrase     string         `json:"passphrase,omitempty"`
	Signature      string         `json:"signature,omitempty"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	Channels      []SubscribedChannel `json:"channels,omitempty"`
	Authenticated bool                `json:"authenticated,omitempty"`
	Channel       string              `json:"channel,omitempty"`
	Type          string              `json:"type,omitempty"`
	Time          time.Time           `json:"time,omitempty"`

	// Error message and failure reason information.
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// SubscribedChannel represents a subscribed channel name and product ID(instrument) list.
type SubscribedChannel struct {
	Name       string   `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

// WsInstrument holds response information to websocket
type WsInstrument struct {
	Sequence            int64        `json:"sequence"`
	ProductID           string       `json:"product_id"`
	InstrumentType      string       `json:"instrument_type"`
	BaseAssetName       string       `json:"base_asset_name"`
	QuoteAssetName      string       `json:"quote_asset_name"`
	BaseIncrement       types.Number `json:"base_increment"`
	QuoteIncrement      types.Number `json:"quote_increment"`
	AvgDailyQuantity    types.Number `json:"avg_daily_quantity"`
	AvgDailyVolume      types.Number `json:"avg_daily_volume"`
	Total30DayQuantity  types.Number `json:"total_30_day_quantity"`
	Total30DayVolume    types.Number `json:"total_30_day_volume"`
	Total24HourQuantity types.Number `json:"total_24_hour_quantity"`
	Total24HourVolume   types.Number `json:"total_24_hour_volume"`
	BaseImf             string       `json:"base_imf"`
	MinQuantity         types.Number `json:"min_quantity"`
	PositionSizeLimit   types.Number `json:"position_size_limit"`
	FundingInterval     string       `json:"funding_interval"`
	TradingState        string       `json:"trading_state"`
	LastUpdateTime      time.Time    `json:"last_update_time"`
	GatewayTime         time.Time    `json:"time"`
	Channel             string       `json:"channel"`
	Type                string       `json:"type"`
}

// WsMatch holds push data information through the channel MATCH.
type WsMatch struct {
	Sequence   int64        `json:"sequence"`
	ProductID  string       `json:"product_id"`
	Time       time.Time    `json:"time"`
	MatchID    string       `json:"match_id"`
	TradeQty   types.Number `json:"trade_qty"`
	TradePrice types.Number `json:"trade_price"`
	Channel    string       `json:"channel"`
	Type       string       `json:"type"`
}

// WsFunding holds push data information through the FUNDING channel.
type WsFunding struct {
	Sequence    int64        `json:"sequence"`
	ProductID   string       `json:"product_id"`
	Time        time.Time    `json:"time"`
	FundingRate types.Number `json:"funding_rate"`
	IsFinal     bool         `json:"is_final"`
	Channel     string       `json:"channel"`
	Type        string       `json:"type"`
}

// WsRisk holds push data information through the RISK channel.
type WsRisk struct {
	Sequence        int64        `json:"sequence"`
	ProductID       string       `json:"product_id"`
	Time            time.Time    `json:"time"`
	LimitUp         string       `json:"limit_up"`
	LimitDown       string       `json:"limit_down"`
	IndexPrice      types.Number `json:"index_price"`
	MarkPrice       types.Number `json:"mark_price"`
	SettlementPrice types.Number `json:"settlement_price"`
	Channel         string       `json:"channel"`
	Type            string       `json:"type"`
}

// WsOrderbookLevel1 holds Level-1 orderbook information
type WsOrderbookLevel1 struct {
	Sequence  int64        `json:"sequence"`
	ProductID string       `json:"product_id"`
	Time      time.Time    `json:"time"`
	BidPrice  types.Number `json:"bid_price"`
	BidQty    types.Number `json:"bid_qty"`
	Channel   string       `json:"channel"`
	Type      string       `json:"type"`
	AskPrice  types.Number `json:"ask_price,omitempty"`
	AskQty    types.Number `json:"ask_qty,omitempty"`
}

// WsOrderbookLevel2 holds Level-2 orderbook information.
type WsOrderbookLevel2 struct {
	Sequence  int64             `json:"sequence"`
	ProductID string            `json:"product_id"`
	Time      time.Time         `json:"time"`
	Asks      [][2]types.Number `json:"asks"`
	Bids      [][2]types.Number `json:"bids"`
	Channel   string            `json:"channel"`
	Type      string            `json:"type"` // Possible values: UPDATE and SNAPSHOT

	// Changes when the data is UPDATE
	Changes [][3]string `json:"changes"`
}
