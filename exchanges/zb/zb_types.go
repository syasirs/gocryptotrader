package zb

import (
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fee"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

// OrderbookResponse holds the orderbook data for a symbol
type OrderbookResponse struct {
	Timestamp int64       `json:"timestamp"`
	Asks      [][]float64 `json:"asks"`
	Bids      [][]float64 `json:"bids"`
}

// AccountsResponseCoin holds the accounts coin details
type AccountsResponseCoin struct {
	Freeze      string `json:"freez"`       // 冻结资产
	EnName      string `json:"enName"`      // 币种英文名
	UnitDecimal int    `json:"unitDecimal"` // 保留小数位
	UnName      string `json:"cnName"`      // 币种中文名
	UnitTag     string `json:"unitTag"`     // 币种符号
	Available   string `json:"available"`   // 可用资产
	Key         string `json:"key"`         // 币种
}

// AccountsBaseResponse holds basic account details
type AccountsBaseResponse struct {
	UserName             string `json:"username"`               // 用户名
	TradePasswordEnabled bool   `json:"trade_password_enabled"` // 是否开通交易密码
	AuthGoogleEnabled    bool   `json:"auth_google_enabled"`    // 是否开通谷歌验证
	AuthMobileEnabled    bool   `json:"auth_mobile_enabled"`    // 是否开通手机验证
}

// Order is the order details for retrieving all orders
type Order struct {
	Currency    string  `json:"currency"`
	ID          int64   `json:"id,string"`
	Price       float64 `json:"price"`
	Status      int     `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	TradeAmount float64 `json:"trade_amount"`
	TradeDate   int     `json:"trade_date"`
	TradeMoney  float64 `json:"trade_money"`
	Type        int64   `json:"type"`
	Fees        float64 `json:"fees,omitempty"`
	TradePrice  float64 `json:"trade_price,omitempty"`
	No          int64   `json:"no,string,omitempty"`
}

// AccountsResponse 用户基本信息
type AccountsResponse struct {
	Result struct {
		Coins []AccountsResponseCoin `json:"coins"`
		Base  AccountsBaseResponse   `json:"base"`
	} `json:"result"` // 用户名
	AssetPerm   bool `json:"assetPerm"`   // 是否开通交易密码
	LeverPerm   bool `json:"leverPerm"`   // 是否开通谷歌验证
	EntrustPerm bool `json:"entrustPerm"` // 是否开通手机验证
	MoneyPerm   bool `json:"moneyPerm"`   // 资产列表
}

// MarketResponseItem stores market data
type MarketResponseItem struct {
	AmountScale float64 `json:"amountScale"`
	PriceScale  float64 `json:"priceScale"`
}

// TickerResponse holds the ticker response data
type TickerResponse struct {
	Date   string              `json:"date"`
	Ticker TickerChildResponse `json:"ticker"`
}

// TickerChildResponse holds the ticker child response data
type TickerChildResponse struct {
	Volume float64 `json:"vol,string"`  // 成交量(最近的24小时)
	Last   float64 `json:"last,string"` // 最新成交价
	Sell   float64 `json:"sell,string"` // 卖一价
	Buy    float64 `json:"buy,string"`  // 买一价
	High   float64 `json:"high,string"` // 最高价
	Low    float64 `json:"low,string"`  // 最低价
}

// SpotNewOrderRequestParamsType ZB 交易类型
type SpotNewOrderRequestParamsType string

var (
	// SpotNewOrderRequestParamsTypeBuy 买
	SpotNewOrderRequestParamsTypeBuy = SpotNewOrderRequestParamsType("1")
	// SpotNewOrderRequestParamsTypeSell 卖
	SpotNewOrderRequestParamsTypeSell = SpotNewOrderRequestParamsType("0")
)

// SpotNewOrderRequestParams is the params used for placing an order
type SpotNewOrderRequestParams struct {
	Amount float64                       `json:"amount"`    // 交易数量
	Price  float64                       `json:"price"`     // 下单价格,
	Symbol string                        `json:"currency"`  // 交易对, btcusdt, bccbtc......
	Type   SpotNewOrderRequestParamsType `json:"tradeType"` // 订单类型, buy-market: 市价买, sell-market: 市价卖, buy-limit: 限价买, sell-limit: 限价卖
}

// SpotNewOrderResponse stores the new order response data
type SpotNewOrderResponse struct {
	Code    int    `json:"code"`    // 返回代码
	Message string `json:"message"` // 提示信息
	ID      string `json:"id"`      // 委托挂单号
}

// //-------------Kline

// KlinesRequestParams represents Klines request data.
type KlinesRequestParams struct {
	Symbol string // 交易对, zb_qc,zb_usdt,zb_btc...
	Type   string // K线类型, 1min, 3min, 15min, 30min, 1hour......
	Since  int64  // 从这个时间戳之后的
	Size   int64  // 返回数据的条数限制(默认为1000，如果返回数据多于1000条，那么只返回1000条)
}

// KLineResponseData Kline Data
type KLineResponseData struct {
	KlineTime time.Time `json:"klineTime"`
	Open      float64   `json:"open"`  // 开盘价
	Close     float64   `json:"close"` // 收盘价, 当K线为最晚的一根时, 时最新成交价
	Low       float64   `json:"low"`   // 最低价
	High      float64   `json:"high"`  // 最高价
	Volume    float64   `json:"vol"`   // 成交量
}

// KLineResponse K线返回类型
type KLineResponse struct {
	// Data      string                `json:"data"`      // 买入货币
	MoneyType string               `json:"moneyType"` // 卖出货币
	Symbol    string               `json:"symbol"`    // 内容说明
	Data      []*KLineResponseData `json:"data"`      // KLine数据
}

// UserAddress defines Users Address for depositing funds
type UserAddress struct {
	Code    int64 `json:"code"`
	Message struct {
		Description  string `json:"des"`
		IsSuccessful bool   `json:"isSuc"`
		Data         struct {
			Address string `json:"key"`
			Tag     string // custom field we populate
		} `json:"datas"`
	} `json:"message"`
}

// MultiChainDepositAddress stores an individual multichain deposit item
type MultiChainDepositAddress struct {
	Blockchain  string `json:"blockChain"`
	IsUseMemo   bool   `json:"isUseMemo"`
	Account     string `json:"account"`
	Address     string `json:"address"`
	Memo        string `json:"memo"`
	CanDeposit  bool   `json:"canDeposit"`
	CanWithdraw bool   `json:"canWithdraw"`
}

// MultiChainDepositAddressResponse stores the multichain deposit address response
type MultiChainDepositAddressResponse struct {
	Code    int64 `json:"code"`
	Message struct {
		Description  string                     `json:"des"`
		IsSuccessful bool                       `json:"isSuc"`
		Data         []MultiChainDepositAddress `json:"datas"`
	} `json:"message"`
}

// WithdrawalFees the large list of predefined withdrawal fees
// Prone to change, using highest value
var WithdrawalFees = map[asset.Item]map[currency.Code]fee.Transfer{
	asset.Spot: {
		currency.ZB:     {Withdrawal: fee.Convert(5)},
		currency.BTC:    {Withdrawal: fee.Convert(0.001)},
		currency.BCH:    {Withdrawal: fee.Convert(0.0006)},
		currency.LTC:    {Withdrawal: fee.Convert(0.005)},
		currency.ETH:    {Withdrawal: fee.Convert(0.01)},
		currency.ETC:    {Withdrawal: fee.Convert(0.01)},
		currency.BTS:    {Withdrawal: fee.Convert(3)},
		currency.EOS:    {Withdrawal: fee.Convert(0.1)},
		currency.QTUM:   {Withdrawal: fee.Convert(0.01)},
		currency.HC:     {Withdrawal: fee.Convert(0.001)},
		currency.XRP:    {Withdrawal: fee.Convert(0.1)},
		currency.QC:     {Withdrawal: fee.Convert(5)},
		currency.DASH:   {Withdrawal: fee.Convert(0.002)},
		currency.BCD:    {Withdrawal: fee.Convert(0)},
		currency.UBTC:   {Withdrawal: fee.Convert(0.001)},
		currency.SBTC:   {Withdrawal: fee.Convert(0)},
		currency.INK:    {Withdrawal: fee.Convert(60)},
		currency.BTH:    {Withdrawal: fee.Convert(0.01)},
		currency.LBTC:   {Withdrawal: fee.Convert(0.01)},
		currency.CHAT:   {Withdrawal: fee.Convert(20)},
		currency.BITCNY: {Withdrawal: fee.Convert(20)},
		currency.HLC:    {Withdrawal: fee.Convert(100)},
		currency.BTP:    {Withdrawal: fee.Convert(0.001)},
		currency.TOPC:   {Withdrawal: fee.Convert(200)},
		currency.ENT:    {Withdrawal: fee.Convert(50)},
		currency.BAT:    {Withdrawal: fee.Convert(40)},
		currency.FIRST:  {Withdrawal: fee.Convert(30)},
		currency.SAFE:   {Withdrawal: fee.Convert(0.001)},
		currency.QUN:    {Withdrawal: fee.Convert(200)},
		currency.BTN:    {Withdrawal: fee.Convert(0.005)},
		currency.TRUE:   {Withdrawal: fee.Convert(5)},
		currency.CDC:    {Withdrawal: fee.Convert(1)},
		currency.DDM:    {Withdrawal: fee.Convert(1)},
		currency.HOTC:   {Withdrawal: fee.Convert(150)},
		currency.USDT:   {Withdrawal: fee.Convert(5)},
		currency.XUC:    {Withdrawal: fee.Convert(1)},
		currency.EPC:    {Withdrawal: fee.Convert(40)},
		currency.BDS:    {Withdrawal: fee.Convert(3)},
		currency.GRAM:   {Withdrawal: fee.Convert(5)},
		currency.DOGE:   {Withdrawal: fee.Convert(20)},
		currency.NEO:    {Withdrawal: fee.Convert(0)},
		currency.OMG:    {Withdrawal: fee.Convert(0.5)},
		currency.BTM:    {Withdrawal: fee.Convert(4)},
		currency.SNT:    {Withdrawal: fee.Convert(60)},
		currency.AE:     {Withdrawal: fee.Convert(3)},
		currency.ICX:    {Withdrawal: fee.Convert(3)},
		currency.ZRX:    {Withdrawal: fee.Convert(10)},
		currency.EDO:    {Withdrawal: fee.Convert(4)},
		currency.FUN:    {Withdrawal: fee.Convert(250)},
		currency.MANA:   {Withdrawal: fee.Convert(70)},
		currency.RCN:    {Withdrawal: fee.Convert(70)},
		currency.MCO:    {Withdrawal: fee.Convert(0.6)},
		currency.MITH:   {Withdrawal: fee.Convert(10)},
		currency.KNC:    {Withdrawal: fee.Convert(5)},
		currency.XLM:    {Withdrawal: fee.Convert(0.1)},
		currency.GNT:    {Withdrawal: fee.Convert(20)},
		currency.MTL:    {Withdrawal: fee.Convert(3)},
		currency.SUB:    {Withdrawal: fee.Convert(20)},
		currency.XEM:    {Withdrawal: fee.Convert(4)},
		currency.EOSDAC: {Withdrawal: fee.Convert(0)},
		currency.KAN:    {Withdrawal: fee.Convert(350)},
		currency.AAA:    {Withdrawal: fee.Convert(1)},
		currency.XWC:    {Withdrawal: fee.Convert(1)},
		currency.PDX:    {Withdrawal: fee.Convert(1)},
		currency.SLT:    {Withdrawal: fee.Convert(100)},
		currency.ADA:    {Withdrawal: fee.Convert(1)},
		currency.HPY:    {Withdrawal: fee.Convert(100)},
		currency.PAX:    {Withdrawal: fee.Convert(5)},
		currency.XTZ:    {Withdrawal: fee.Convert(0.1)},
	},
}

// orderSideMap holds order type info based on Alphapoint data
var orderSideMap = map[int64]order.Side{
	0: order.Buy,
	1: order.Sell,
}

// TradeHistory defines a slice of historic trades
type TradeHistory []struct {
	Amount    float64 `json:"amount,string"`
	Date      int64   `json:"date"`
	Price     float64 `json:"price,string"`
	Tid       int64   `json:"tid"`
	TradeType string  `json:"trade_type"`
	Type      string  `json:"type"`
}
