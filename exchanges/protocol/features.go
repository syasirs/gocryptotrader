package protocol

// Features holds all variables for the exchanges supported features
// for a protocol (e.g REST or Websocket)
type Features struct {
	TickerBatching      bool `json:"tickerBatching,omitempty"`
	AutoPairUpdates     bool `json:"autoPairUpdates,omitempty"`
	AccountBalance      bool `json:"accountBalance,omitempty"`
	CryptoDeposit       bool `json:"cryptoDeposit,omitempty"`
	CryptoWithdrawal    bool `json:"cryptoWithdrawal,omitempty"`
	FiatWithdraw        bool `json:"fiatWithdraw,omitempty"`
	GetOrder            bool `json:"getOrder,omitempty"`
	GetOrders           bool `json:"getOrders,omitempty"`
	CancelOrders        bool `json:"cancelOrders,omitempty"`
	CancelOrder         bool `json:"cancelOrder,omitempty"`
	SubmitOrder         bool `json:"submitOrder,omitempty"`
	SubmitOrders        bool `json:"submitOrders,omitempty"`
	ModifyOrder         bool `json:"modifyOrder,omitempty"`
	DepositHistory      bool `json:"depositHistory,omitempty"`
	WithdrawalHistory   bool `json:"withdrawalHistory,omitempty"`
	TradeHistory        bool `json:"tradeHistory,omitempty"`
	UserTradeHistory    bool `json:"userTradeHistory,omitempty"`
	TradeFee            bool `json:"tradeFee,omitempty"`
	FiatDepositFee      bool `json:"fiatDepositFee,omitempty"`
	FiatWithdrawalFee   bool `json:"fiatWithdrawalFee,omitempty"`
	CryptoDepositFee    bool `json:"cryptoDepositFee,omitempty"`
	CryptoWithdrawalFee bool `json:"cryptoWithdrawalFee,omitempty"`
	TickerFetching      bool `json:"tickerFetching,omitempty"`
	KlineFetching       bool `json:"klineFetching,omitempty"`
	TradeFetching       bool `json:"tradeFetching,omitempty"`
	OrderbookFetching   bool `json:"orderbookFetching,omitempty"`
	AccountInfo         bool `json:"accountInfo,omitempty"`
	FiatDeposit         bool `json:"fiatDeposit,omitempty"`
	DeadMansSwitch      bool `json:"deadMansSwitch,omitempty"`
	// FullPayloadSubscribe flushes and changes full subscription on websocket
	// connection by subscribing with full default stream channel list
	FullPayloadSubscribe   bool `json:"fullPayloadSubscribe,omitempty"`
	Subscribe              bool `json:"subscribe,omitempty"`
	Unsubscribe            bool `json:"unsubscribe,omitempty"`
	AuthenticatedEndpoints bool `json:"authenticatedEndpoints,omitempty"`
	MessageCorrelation     bool `json:"messageCorrelation,omitempty"`
	MessageSequenceNumbers bool `json:"messageSequenceNumbers,omitempty"`
	CandleHistory          bool `json:"candlehistory,omitempty"`
	MultiChainDeposits     bool `json:"multiChainDeposits,omitempty"`
	MultiChainWithdrawals  bool `json:"multiChainWithdrawals,omitempty"`
}
