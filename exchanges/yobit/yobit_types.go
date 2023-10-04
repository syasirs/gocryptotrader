package yobit

import "github.com/thrasher-corp/gocryptotrader/currency"

// Response is a generic struct used for exchange API request result
type Response struct {
	Return  interface{} `json:"return"`
	Success int         `json:"success"`
	Error   string      `json:"error"`
}

// Info holds server time and pair information
type Info struct {
	ServerTime int64           `json:"server_time"`
	Pairs      map[string]Pair `json:"pairs"`
}

// Ticker stores the ticker information
type Ticker struct {
	High          float64 // maximal price
	Low           float64 // minimal price
	Avg           float64 // average price
	Vol           float64 // traded volume
	VolumeCurrent float64 `json:"vol_cur"` // traded volume in currency
	Last          float64 // last transaction price
	Buy           float64 // buying price
	Sell          float64 // selling price
	Updated       int64   // last cache upgrade
}

// Orderbook stores the asks and bids orderbook information
type Orderbook struct {
	Asks [][]float64 `json:"asks"` // selling orders
	Bids [][]float64 `json:"bids"` // buying orders
}

// Trade stores trade information
type Trade struct {
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
	TID       int64   `json:"tid"`
	Timestamp int64   `json:"timestamp"`
}

// ActiveOrders stores active order information
type ActiveOrders struct {
	Pair             string  `json:"pair"`
	Type             string  `json:"type"`
	Amount           float64 `json:"amount"`
	Rate             float64 `json:"rate"`
	TimestampCreated float64 `json:"timestamp_created"`
	Status           int     `json:"status"`
}

// Pair holds pair information
type Pair struct {
	DecimalPlaces int     `json:"decimal_places"` // Quantity of permitted numbers after decimal point
	MinPrice      float64 `json:"min_price"`      // Minimal permitted price
	MaxPrice      float64 `json:"max_price"`      // Maximal permitted price
	MinAmount     float64 `json:"min_amount"`     // Minimal permitted buy or sell amount
	Hidden        int     `json:"hidden"`         // Pair is hidden (0 or 1)
	Fee           float64 `json:"fee"`            // Pair commission
}

// AccountInfo stores the account information for a user
type AccountInfo struct {
	Funds           map[string]float64 `json:"funds"`
	FundsInclOrders map[string]float64 `json:"funds_incl_orders"`
	Rights          struct {
		Info     int `json:"info"`
		Trade    int `json:"trade"`
		Withdraw int `json:"withdraw"`
	} `json:"rights"`
	TransactionCount int     `json:"transaction_count"`
	OpenOrders       int     `json:"open_orders"`
	ServerTime       float64 `json:"server_time"`
	Error            string  `json:"error"`
}

// OrderInfo stores order information
type OrderInfo struct {
	Pair             string  `json:"pair"`
	Type             string  `json:"type"`
	StartAmount      float64 `json:"start_amount"`
	Amount           float64 `json:"amount"`
	Rate             float64 `json:"rate"`
	TimestampCreated float64 `json:"timestamp_created"`
	Status           int     `json:"status"`
}

// CancelOrder is used for the CancelOrder API request response
type CancelOrder struct {
	OrderID float64            `json:"order_id"`
	Funds   map[string]float64 `json:"funds"`
	Error   string             `json:"error"`
}

// TradeOrderResponse stores the trade information
type TradeOrderResponse struct {
	Received float64            `json:"received"`
	Remains  float64            `json:"remains"`
	OrderID  float64            `json:"order_id"`
	Funds    map[string]float64 `json:"funds"`
	Error    string             `json:"error"`
}

// TradeHistoryResponse returns all your trade history
type TradeHistoryResponse struct {
	Success int64                   `json:"success"`
	Data    map[string]TradeHistory `json:"return,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// TradeHistory stores trade history
type TradeHistory struct {
	Pair      string  `json:"pair"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	OrderID   float64 `json:"order_id"`
	MyOrder   int     `json:"is_your_order"`
	Timestamp float64 `json:"timestamp"`
}

// DepositAddress stores a currency deposit address
type DepositAddress struct {
	Success int `json:"success"`
	Return  struct {
		Address         string  `json:"address"`
		ProcessedAmount float64 `json:"processed_amount"`
		ServerTime      int64   `json:"server_time"`
	} `json:"return"`
	Error string `json:"error"`
}

// WithdrawCoinsToAddress stores information for a withdrawcoins request
type WithdrawCoinsToAddress struct {
	ServerTime int64  `json:"server_time"`
	Error      string `json:"error"`
}

// CreateCoupon stores information coupon information
type CreateCoupon struct {
	Coupon  string             `json:"coupon"`
	TransID int64              `json:"transID"`
	Funds   map[string]float64 `json:"funds"`
	Error   string             `json:"error"`
}

// RedeemCoupon stores redeem coupon information
type RedeemCoupon struct {
	CouponAmount   float64            `json:"couponAmount,string"`
	CouponCurrency string             `json:"couponCurrency"`
	TransID        int64              `json:"transID"`
	Funds          map[string]float64 `json:"funds"`
	Error          string             `json:"error"`
}

// WithdrawalFees the large list of predefined withdrawal fees
// Prone to change, using highest value
var WithdrawalFees = map[currency.Code]float64{
	currency.ZERO07:     0.002,
	currency.BIT16:      0.002,
	currency.TWO015:     0.002,
	currency.TWO56:      0.0002,
	currency.TWOBACCO:   0.002,
	currency.TWOGIVE:    0.01,
	currency.THIRTY2BIT: 0.002,
	currency.THREE65:    0.01,
	currency.FOUR04:     0.01,
	currency.SEVEN00:    0.01,
	currency.EIGHTBIT:   0.002,
	currency.ACLR:       0.002,
	currency.ACES:       0.01,
	currency.ACPR:       0.01,
	currency.ACID:       0.01,
	currency.ACOIN:      0.01,
	currency.ACRN:       0.01,
	currency.ADAM:       0.01,
	currency.ADT:        0.05,
	currency.AIB:        0.01,
	currency.ADZ:        0.002,
	currency.AECC:       0.002,
	currency.AM:         0.002,
	currency.AE:         10,
	currency.DLT:        0.05,
	currency.AGRI:       0.01,
	currency.AGT:        0.01,
	currency.AST:        0.002,
	currency.AIR:        0.01,
	currency.ALEX:       0.01,
	currency.AUM:        0.002,
	currency.ALIEN:      0.01,
	currency.ALIS:       0.05,
	currency.ALL:        0.01,
	currency.ASAFE:      0.01,
	currency.AMBER:      0.002,
	currency.AMS:        0.002,
	currency.ANAL:       0.002,
	currency.ACP:        0.002,
	currency.ANI:        0.01,
	currency.ANTI:       0.002,
	currency.ALTC:       0.01,
	currency.APT:        0.01,
	currency.ARCO:       0.002,
	currency.ALC:        0.01,
	currency.ANT:        0.01,
	currency.ARB:        0.002,
	currency.ARCT:       10,
	currency.ARCX:       0.01,
	currency.ARGUS:      0.01,
	currency.ARH:        0.01,
	currency.ARM:        0.01,
	currency.ARNA:       10,
	currency.ARPA:       0.002,
	currency.ARTA:       0.01,
	currency.ABY:        0.01,
	currency.ARTC:       0.01,
	currency.AL:         0.01,
	currency.ASN:        0.002,
	currency.ADCN:       0.01,
	currency.ATB:        0.01,
	currency.ATL:        0.1,
	currency.ATM:        0.002,
	currency.ATMCHA:     0.05,
	currency.ATOM:       0.01,
	currency.ADC:        0.002,
	currency.REP:        0.002,
	currency.ARE:        0.002,
	currency.AUR:        0.01,
	currency.AV:         0.002,
	currency.AXIOM:      0.002,
	currency.B2B:        10,
	currency.B2:         0.01,
	currency.B3:         0.1,
	currency.BAB:        0.01,
	currency.BAN:        0.002,
	currency.BamitCoin:  0.002,
	currency.NANAS:      0.002,
	currency.BNT:        0.05,
	currency.BBCC:       0.002,
	currency.BAT:        0.05,
	currency.BTA:        0.002,
	currency.BSTK:       0.002,
	currency.BATL:       0.01,
	currency.BBH:        0.01,
	currency.BITB:       0.002,
	currency.BRDD:       0.002,
	currency.XBTS:       0.01,
	currency.BVC:        0.01,
	currency.CHATX:      10,
	currency.BEEP:       0.01,
	currency.BEEZ:       0.002,
	currency.BENJI:      0.01,
	currency.BERN:       0.002,
	currency.PROFIT:     0.01,
	currency.BEST:       0.01,
	currency.BGF:        0.01,
	currency.BIGUP:      0.002,
	currency.BLRY:       0.01,
	currency.BILL:       0.01,
	currency.BNB:        0.002,
	currency.BIOB:       0.01,
	currency.BIO:        0.1,
	currency.BIOS:       0.002,
	currency.BPTN:       10,
	currency.BTCA:       10,
	currency.BA:         0.002,
	currency.BAC:        0.002,
	currency.BBT:        10,
	currency.BOSS:       0.01,
	currency.BRONZ:      0.002,
	currency.CAT:        0.01,
	currency.BTD:        0.01,
	currency.BTC:        0.0012,
	currency.XBTC21:     0.01,
	currency.BCA:        0.01,
	currency.BCH:        0.01,
	currency.BCP:        0.01,
	currency.BCD:        0.01,
	currency.BTDOLL:     0.01,
	currency.GOD:        0.01,
	currency.BTG:        0.01,
	currency.LIZA:       0.01,
	currency.BTCRED:     10,
	currency.BTCS:       0.01,
	currency.BTU:        0.01,
	currency.BUM:        0.01,
	currency.LITE:       0.01,
	currency.BCM:        0.01,
	currency.BCS:        0.01,
	currency.BTCU:       0.002,
	currency.BM:         10,
	currency.BTCRY:      0.002,
	currency.BTCR:       0.002,
	currency.HIRE:       0.002,
	currency.STU:        10,
	currency.BITOK:      0.0001,
	currency.BITON:      0.002,
	currency.BPC:        0.01,
	currency.BPOK:       0.01,
	currency.BTP:        0.002,
	currency.RNTB:       10,
	currency.BSH:        0.002,
	currency.BTS:        5,
	currency.XBS:        0.002,
	currency.BITS:       0.01,
	currency.BST:        0.002,
	currency.BXT:        0.01,
	currency.VEG:        0.002,
	currency.VOLT:       0.01,
	currency.BTV:        0.01,
	currency.BITZ:       0.002,
	currency.BTZ:        0.002,
	currency.BHC:        0.01,
	currency.BDC:        0.002,
	currency.JACK:       0.01,
	currency.BS:         0.01,
	currency.BSTAR:      0.01,
	currency.BLAZR:      0.01,
	currency.BOD:        0.002,
	currency.BLUE:       10,
	currency.BLU:        0.002,
	currency.BLUS:       0.002,
	currency.BMT:        10,
	currency.BOT:        0.002,
	currency.BOLI:       0.002,
	currency.BOMB:       0.01,
	currency.BON:        0.01,
	currency.BOOM:       0.002,
	currency.BOSON:      0.01,
	currency.BSC:        0.002,
	currency.BRH:        10,
	currency.BRAIN:      0.01,
	currency.BRE:        0.002,
	currency.BTCM:       0.1,
	currency.BTCO:       0.01,
	currency.TALK:       0.01,
	currency.BUB:        0.002,
	currency.BUY:        0.01,
	currency.BUZZ:       0.002,
	currency.BTH:        0.1,
	currency.C0C0:       0.002,
	currency.CAB:        0.01,
	currency.CF:         0.002,
	currency.CLO:        10,
	currency.CAM:        0.2,
	currency.CD:         0.002,
	currency.CANN:       0.2,
	currency.CNNC:       0.01,
	currency.CPC:        0.002,
	currency.CST:        0.01,
	currency.CAPT:       0.002,
	currency.CARBON:     0.01,
	currency.CME:        0.002,
	currency.CTK:        0.002,
	currency.CBD:        0.01,
	currency.CCC:        0.01,
	currency.CNT:        0.01,
	currency.XCE:        0.002,
	currency.CAG:        1,
	currency.CHRG:       0.01,
	currency.CHAT:       0.01,
	currency.CHEMX:      0.01,
	currency.CHESS:      0.01,
	currency.CKS:        0.01,
	currency.CHILL:      0.01,
	currency.CHIP:       0.002,
	currency.CHOOF:      0.01,
	currency.TIME:       0.05,
	currency.CRX:        0.01,
	currency.CIN:        0.01,
	currency.CLAM:       0.002,
	currency.POLL:       10,
	currency.CLICK:      0.002,
	currency.CLINT:      0.01,
	currency.CLOAK:      0.002,
	currency.CLUB:       0.002,
	currency.CLUD:       0.01,
	currency.COX:        0.01,
	currency.COXST:      0.01,
	currency.CFC:        0.002,
	currency.CTIC2:      0.01,
	currency.COIN:       0.01,
	currency.BTTF:       0.002,
	currency.C2:         0.01,
	currency.CAID:       0.002,
	currency.CL:         10,
	currency.CTIC:       0.01,
	currency.CXT:        0.01,
	currency.CHP:        10,
	currency.CV2:        0.002,
	currency.CMT:        0.01,
	currency.COC:        0.01,
	currency.COMP:       0.01,
	currency.CMS:        10,
	currency.CONX:       0.01,
	currency.CCX:        0.01,
	currency.CLR:        10,
	currency.CORAL:      0.01,
	currency.CORG:       0.01,
	currency.CSMIC:      0.01,
	currency.CMC:        0.01,
	currency.COV:        0.002,
	currency.COVX:       10,
	currency.CRAB:       0.01,
	currency.CRAFT:      0.01,
	currency.CRNK:       0.01,
	currency.CRAVE:      0.002,
	currency.CRM:        0.01,
	currency.XCRE:       0.01,
	currency.CREDIT:     0.002,
	currency.CREVA:      0.002,
	currency.CRIME:      0.002,
	currency.CROC:       0.01,
	currency.CRC:        10,
	currency.CRW:        0.002,
	currency.CRY:        0.002,
	currency.CBX:        0.002,
	currency.TKTX:       10,
	currency.CB:         0.02,
	currency.CIRC:       0.002,
	currency.CCB:        0.002,
	currency.CDO:        0.01,
	currency.CG:         0.01,
	currency.CJ:         0.01,
	currency.CJC:        0.01,
	currency.CYT:        0.002,
	currency.CNX:        0.01,
	currency.CRPS:       0.002,
	currency.PING:       0.05,
	currency.CS:         0.002,
	currency.CWXT:       0.01,
	currency.CCT:        0.05,
	currency.CTL:        0.01,
	currency.CURVES:     0.002,
	currency.CC:         0.002,
	currency.CYC:        0.002,
	currency.CYG:        0.002,
	currency.CYP:        0.002,
	currency.FUNK:       0.01,
	currency.CZECO:      0.01,
	currency.DALC:       0.1,
	currency.DLISK:      0.2,
	currency.MOOND:      0.002,
	currency.DB:         0.002,
	currency.DCC:        0.002,
	currency.DCYP:       0.002,
	currency.DETH:       0.002,
	currency.DKC:        0.01,
	currency.DISK:       0.01,
	currency.DRKT:       0.002,
	currency.DTT:        0.002,
	currency.DASH:       0.002,
	currency.DASHS:      0.01,
	currency.DBTC:       0.01,
	currency.DCT:        0.002,
	currency.DBET:       10,
	currency.DEC:        0.002,
	currency.DCR:        0.05,
	currency.DECR:       0.002,
	currency.DEA:        0.01,
	currency.DPAY:       0.01,
	currency.DCRE:       0.002,
	currency.DC:         0.002,
	currency.DES:        0.01,
	currency.DEM:        0.002,
	currency.DXC:        0.01,
	currency.DCK:        0.01,
	currency.DGB:        0.002,
	currency.CUBE:       0.002,
	currency.DGMS:       0.002,
	currency.DBG:        0.01,
	currency.DGCS:       0.002,
	currency.DBLK:       0.002,
	currency.DGD:        0.002,
	currency.DIME:       0.002,
	currency.DIRT:       0.002,
	currency.DVD:        10,
	currency.DMT:        10,
	currency.NOTE:       0.002,
	currency.DOGE:       100,
	currency.DGORE:      0.002,
	currency.DLC:        0.01,
	currency.DRT:        0.1,
	currency.DOTA:       0.01,
	currency.DOX:        0.002,
	currency.DRA:        0.002,
	currency.DFT:        0.002,
	currency.XDB:        0.002,
	currency.DRM:        0.002,
	currency.DRZ:        0.002,
	currency.DRACO:      0.002,
	currency.DBIC:       0.002,
	currency.DUB:        0.002,
	currency.GUM:        0.002,
	currency.DUR:        0.01,
	currency.DUST:       0.002,
	currency.DUX:        0.01,
	currency.DXO:        0.01,
	currency.ECN:        0.01,
	currency.EDR2:       0.002,
	currency.EA:         0.002,
	currency.EAGS:       0.002,
	currency.EMT:        10,
	currency.EBONUS:     0.001,
	currency.ECCHI:      0.01,
	currency.EKO:        0.01,
	currency.ECLI:       0.002,
	currency.ECOB:       3,
	currency.ECO:        0.01,
	currency.EDIT:       0.01,
	currency.EDRC:       0.01,
	currency.EDC:        0.01,
	currency.EGAME:      0.01,
	currency.EGG:        0.002,
	currency.EGO:        0.01,
	currency.ELC:        0.01,
	currency.ELCO:       0.01,
	currency.ECA:        0.01,
	currency.EPC:        0.002,
	currency.ELE:        0.005,
	currency.ONE337:     0.002,
	currency.EMB:        0.01,
	currency.EMC:        0.02,
	currency.EPY:        0.002,
	currency.EMPC:       0.01,
	currency.EMP:        0.002,
	currency.ENE:        0.01,
	currency.EET:        10,
	currency.XNG:        0.01,
	currency.EGMA:       0.002,
	currency.ENTER:      0.01,
	currency.ETRUST:     0.002,
	currency.EOS:        10,
	currency.EQL:        0.01,
	currency.EQM:        0.002,
	currency.EQT:        0.01,
	currency.ERR:        0.002,
	currency.ESC:        0.002,
	currency.ESP:        0.01,
	currency.ENT:        0.01,
	currency.ETCO:       0.2,
	currency.DOGETH:     0.002,
	currency.ETH:        0.005,
	currency.ECASH:      0.1,
	currency.ETC:        0.005,
	currency.ELITE:      0.05,
	currency.ETHS:       0.01,
	currency.ETL:        1,
	currency.ETZ:        10,
	currency.EUC:        0.01,
	currency.EURC:       0.002,
	currency.EUROPE:     0.01,
	currency.EVA:        0.01,
	currency.EGC:        0.002,
	currency.EOC:        0.002,
	currency.EVIL:       0.002,
	currency.EVO:        0.002,
	currency.EXB:        0.002,
	currency.EXIT:       0.01,
	currency.EXP:        0.01,
	currency.XT:         0.01,
	currency.F16:        0.01,
	currency.FADE:       0.002,
	currency.DROP:       0.002,
	currency.FAZZ:       0.01,
	currency.FX:         0.01,
	currency.FIDEL:      0.01,
	currency.FIDGT:      0.01,
	currency.FIND:       0.002,
	currency.FPC:        0.01,
	currency.FIRE:       0.002,
	currency.FFC:        0.002,
	currency.FRST:       0.01,
	currency.FIST:       0.002,
	currency.FIT:        0.05,
	currency.FLX:        0.01,
	currency.FLVR:       0.01,
	currency.FLY:        0.002,
	currency.FONZ:       0.002,
	currency.XFCX:       0.002,
	currency.FOREX:      0.01,
	currency.FRN:        0.002,
	currency.FRK:        0.002,
	currency.FRWC:       0.01,
	currency.FGZ:        0.01,
	currency.FRE:        0.002,
	currency.FRDC:       0.002,
	currency.FJC:        0.01,
	currency.FURY:       0.002,
	currency.FSN:        0.002,
	currency.FCASH:      0.002,
	currency.FTO:        0.01,
	currency.FUZZ:       0.002,
	currency.GAKH:       0.01,
	currency.GBT:        0.01,
	currency.GAME:       0.01,
	currency.GML:        0.2,
	currency.UNITS:      0.01,
	currency.FOUR20G:    0.01,
	currency.GSY:        0.002,
	currency.GENIUS:     0.002,
	currency.GEN:        0.002,
	currency.GEO:        0.002,
	currency.GER:        0.01,
	currency.GSR:        0.01,
	currency.SPKTR:      0.002,
	currency.GIFT:       0.002,
	currency.WTT:        10,
	currency.GHS:        0.01,
	currency.GIG:        0.002,
	currency.GOT:        0.01,
	currency.XGTC:       0.01,
	currency.GIZ:        0.002,
	currency.GLO:        0.01,
	currency.GCR:        0.002,
	currency.BSTY:       0.002,
	currency.GLC:        0.01,
	currency.GSX:        0.02,
	currency.GNO:        0.05,
	currency.GOAT:       0.002,
	currency.GO:         0.01,
	currency.GB:         0.01,
	currency.GFL:        0.01,
	currency.MNTP:       10,
	currency.GP:         0.002,
	currency.GNT:        0.05,
	currency.GLUCK:      0.002,
	currency.GOON:       0.01,
	currency.GTFO:       0.002,
	currency.GOTX:       0.01,
	currency.GPU:        0.01,
	currency.GRF:        0.002,
	currency.GRAM:       0.002,
	currency.GRAV:       0.002,
	currency.GBIT:       0.002,
	currency.GREED:      0.002,
	currency.GE:         0.002,
	currency.GREENF:     0.01,
	currency.GRE:        0.01,
	currency.GREXIT:     0.002,
	currency.GMCX:       0.002,
	currency.GROW:       0.01,
	currency.GSM:        0.002,
	currency.GT:         0.01,
	currency.NLG:        0.01,
	currency.HKN:        10,
	currency.HAC:        0.05,
	currency.HALLO:      0.01,
	currency.HAMS:       0.01,
	currency.HCC:        0.01,
	currency.HPC:        0.01,
	currency.HMC:        0.01,
	currency.HAWK:       0.01,
	currency.HAZE:       0.002,
	currency.HZT:        0.002,
	currency.HDG:        0.1,
	currency.HEDG:       0.002,
	currency.HEEL:       0.002,
	currency.HMP:        0.01,
	currency.PLAY:       0.01,
	currency.HXX:        0.002,
	currency.XHI:        0.01,
	currency.HVCO:       0.01,
	currency.HTC:        0.01,
	currency.MINH:       0.01,
	currency.HODL:       0.01,
	currency.HON:        0.01,
	currency.HOPE:       0.01,
	currency.HQX:        10,
	currency.HSP:        0.002,
	currency.HTML5:      0.002,
	currency.HMQ:        0.01,
	currency.HYPERX:     0.01,
	currency.HPS:        10,
	currency.IOC:        0.002,
	currency.IBANK:      0.01,
	currency.IBITS:      0.002,
	currency.ICASH:      0.002,
	currency.ICOB:       0.01,
	currency.ICN:        0.002,
	currency.ICON:       0.01,
	currency.IETH:       0.1,
	currency.ILM:        0.002,
	currency.IMPS:       0.01,
	currency.NKA:        0.002,
	currency.INCP:       0.01,
	currency.IN:         0.01,
	currency.INC:        0.002,
	currency.IMS:        0.01,
	currency.IND:        0.01,
	currency.XIN:        0.01,
	currency.IFLT:       0.01,
	currency.INFX:       0.002,
	currency.INGT:       0.01,
	currency.INPAY:      0.01,
	currency.INSANE:     0.01,
	currency.INXT:       0.01,
	currency.IFT:        0.05,
	currency.INV:        0.01,
	currency.IVZ:        0.002,
	currency.ILT:        0.002,
	currency.IONX:       0.01,
	currency.ISL:        0.002,
	currency.ITI:        0.01,
	currency.ING:        10,
	currency.IEC:        0.002,
	currency.IW:         0.01,
	currency.IXC:        0.01,
	currency.IXT:        0.05,
	currency.JPC:        0.002,
	currency.JANE:       0.01,
	currency.JWL:        0.01,
	currency.JNT:        0.01,
	currency.JIF:        0.002,
	currency.JOBS:       0.01,
	currency.JOCKER:     0.01,
	currency.JW:         0.01,
	currency.JOK:        0.01,
	currency.XJO:        0.002,
	currency.KGB:        0.01,
	currency.KARMC:      0.01,
	currency.KARMA:      0.002,
	currency.KASHH:      0.01,
	currency.KAT:        0.002,
	currency.KC:         0.002,
	currency.KICK:       0.05,
	currency.KIDS:       0.01,
	currency.KIN:        10,
	currency.KNC:        0.01,
	currency.KISS:       0.01,
	currency.KOBO:       0.002,
	currency.TP1:        0.002,
	currency.KRAK:       0.002,
	currency.KGC:        0.002,
	currency.KTK:        0.002,
	currency.KR:         0.005,
	currency.KUBO:       0.01,
	currency.KURT:       0.01,
	currency.KUSH:       0.01,
	currency.LANA:       0.01,
	currency.LTH:        0.01,
	currency.LAZ:        0.2,
	currency.LEA:        0.002,
	currency.LEAF:       0.002,
	currency.LENIN:      0.01,
	currency.LEPEN:      0.01,
	currency.LIR:        0.01,
	currency.LVG:        0.002,
	currency.LGBTQ:      0.002,
	currency.LHC:        10,
	currency.EXT:        0.002,
	currency.LBTC:       0.1,
	currency.LSD:        0.01,
	currency.LIMX:       0.002,
	currency.LTD:        0.0000002,
	currency.LINDA:      0.01,
	currency.LKC:        0.002,
	currency.LSK:        0.2,
	currency.LBTCX:      0.01,
	currency.LTC:        0.002,
	currency.LCC:        1,
	currency.LTCU:       0.01,
	currency.LTCR:       0.002,
	currency.LDOGE:      0.002,
	currency.LTS:        0.002,
	currency.LIV:        0.01,
	currency.LIZI:       0.01,
	currency.LOC:        0.01,
	currency.LOCX:       10,
	currency.LOOK:       0.01,
	currency.LRC:        0.05,
	currency.LOOT:       0.01,
	currency.XLTCG:      0.002,
	currency.BASH:       0.01,
	currency.LUCKY:      0.002,
	currency.L7S:        0.002,
	currency.LDM:        0.05,
	currency.LUMI:       0.01,
	currency.LUNA:       0.01,
	currency.LUN:        0.002,
	currency.LC:         0.01,
	currency.LUX:        0.002,
	currency.MCRN:       0.01,
	currency.XMG:        0.01,
	currency.MMXIV:      0.02,
	currency.MAT:        0.01,
	currency.MAO:        0.01,
	currency.MAPC:       0.002,
	currency.MRB:        0.002,
	currency.MXT:        0.01,
	currency.MARV:       0.01,
	currency.MARX:       0.01,
	currency.MCAR:       0.002,
	currency.MM:         0.002,
	currency.GUP:        0.05,
	currency.MVC:        0.05,
	currency.MAVRO:      0.01,
	currency.MAX:        0.01,
	currency.MAZE:       0.002,
	currency.MBIT:       0.01,
	currency.MCOIN:      0.01,
	currency.MPRO:       0.01,
	currency.XMS:        0.002,
	currency.MLITE:      0.01,
	currency.MLNC:       0.01,
	currency.MENTAL:     0.01,
	currency.MERGEC:     0.01,
	currency.MTLMC3:     0.002,
	currency.METAL:      0.002,
	currency.AMM:        0.01,
	currency.MDT:        0.002,
	currency.MUU:        0.01,
	currency.MILO:       0.01,
	currency.MND:        0.002,
	currency.XMINE:      0.002,
	currency.MNM:        0.01,
	currency.XNM:        0.01,
	currency.MIRO:       10,
	currency.MIS:        0.002,
	currency.MMXVI:      0.01,
	currency.MGO:        0.01,
	currency.MOIN:       0.002,
	currency.MOJO:       0.01,
	currency.TAB:        0.002,
	currency.MCO:        0.005,
	currency.MONETA:     0.002,
	currency.MUE:        0.002,
	currency.MONEY:      0.01,
	currency.MRP:        0.002,
	currency.MOTO:       0.002,
	currency.MULTI:      0.01,
	currency.MST:        0.01,
	currency.MVR:        0.01,
	currency.MYSTIC:     0.002,
	currency.WISH:       10,
	currency.NKT:        0.002,
	currency.NMC:        0.002,
	currency.NAT:        0.002,
	currency.ENAU:       10,
	currency.NAV:        0.002,
	currency.NEBU:       0.002,
	currency.NEF:        0.01,
	currency.XEM:        20,
	currency.NBIT:       0.01,
	currency.NETKO:      0.01,
	currency.NTM:        0.01,
	currency.NETC:       0.002,
	currency.NEU:        10,
	currency.NRC:        1,
	currency.NTK:        10,
	currency.NTRN:       0.002,
	currency.NEVA:       0.01,
	currency.NIC:        0.01,
	currency.NKC:        0.002,
	currency.NYC:        0.01,
	currency.NZC:        0.01,
	currency.NICE:       0.002,
	currency.NET:        0.01,
	currency.NDOGE:      0.002,
	currency.XTR:        0.01,
	currency.N2O:        0.002,
	currency.NIXON:      0.01,
	currency.NOC:        0.002,
	currency.NODC:       0.01,
	currency.NODES:      0.002,
	currency.NODX:       0.002,
	currency.NLC:        0.01,
	currency.NLC2:       0.01,
	currency.NOO:        0.002,
	currency.NVC:        0.002,
	currency.NPC:        0.002,
	currency.NUBIS:      0.002,
	currency.NUKE:       0.002,
	currency.N7:         0.01,
	currency.NUM:        0.01,
	currency.NMR:        0.05,
	currency.NXE:        0.002,
	currency.OBS:        0.002,
	currency.OCEAN:      0.01,
	currency.OCOW:       0.01,
	currency.EIGHT88:    0.02,
	currency.OCC:        0.02,
	currency.OK:         0.002,
	currency.ODNT:       0.002,
	currency.FLAV:       0.002,
	currency.OLIT:       0.01,
	currency.OLYMP:      0.01,
	currency.OMA:        0.002,
	currency.OMC:        0.01,
	currency.OMG:        0.01,
	currency.ONEK:       0.05,
	currency.ONX:        0.01,
	currency.XPO:        0.01,
	currency.OPAL:       0.2,
	currency.OTN:        0.1,
	currency.OP:         0.01,
	currency.OPES:       0.002,
	currency.OPTION:     0.002,
	currency.ORLY:       0.01,
	currency.OS76:       0.002,
	currency.OZC:        0.002,
	currency.P7C:        0.002,
	currency.PAC:        1,
	currency.PAK:        0.002,
	currency.PAL:        0.01,
	currency.PND:        0.002,
	currency.PINKX:      0.01,
	currency.POPPY:      0.01,
	currency.DUO:        0.002,
	currency.PARA:       0.002,
	currency.PKB:        0.002,
	currency.GENE:       0.002,
	currency.PARTY:      0.01,
	currency.PYN:        10,
	currency.XPY:        0.002,
	currency.CON:        0.002,
	currency.PAYP:       0.01,
	currency.PPC:        0.2,
	currency.GUESS:      10,
	currency.PEN:        0.002,
	currency.PTA:        0.002,
	currency.PEO:        0.002,
	currency.PSB:        0.01,
	currency.XPD:        0.01,
	currency.PXL:        0.002,
	currency.PHR:        0.002,
	currency.PIE:        0.01,
	currency.PIO:        0.01,
	currency.PIPR:       0.01,
	currency.SKULL:      0.01,
	currency.PIVX:       0.002,
	currency.PLANET:     0.002,
	currency.PNC:        0.002,
	currency.XPTX:       0.01,
	currency.PLNC:       0.002,
	currency.PLU:        0.01,
	currency.XPS:        0.01,
	currency.POKE:       0.01,
	currency.PLBT:       0.01,
	currency.POLY:       0.002,
	currency.POM:        0.001,
	currency.PONZ2:      0.01,
	currency.PONZI:      0.01,
	currency.XSP:        0.002,
	currency.PPT:        10,
	currency.XPC:        0.002,
	currency.PEX:        0.002,
	currency.TRON:       0.002,
	currency.POST:       0.01,
	currency.POSW:       0.01,
	currency.PWR:        0.01,
	currency.POWER:      0.002,
	currency.PRE:        0.002,
	currency.PRS:        10,
	currency.PXI:        0.002,
	currency.PEXT:       10,
	currency.PRIMU:      0.01,
	currency.PRX:        0.01,
	currency.PRM:        0.01,
	currency.PRIX:       10,
	currency.XPRO:       0.002,
	currency.PCM:        0.01,
	currency.PROC:       0.01,
	currency.NANOX:      0.01,
	currency.VRP:        0.01,
	currency.PTY:        0.002,
	currency.PSI:        0.002,
	currency.PSY:        0.002,
	currency.PULSE:      0.01,
	currency.PUPA:       0.01,
	currency.PURE:       0.002,
	currency.VIDZ:       0.01,
	currency.PUTIN:      0.01,
	currency.PX:         0.1,
	currency.QTM:        0.01,
	currency.QTZ:        0.002,
	currency.QBC:        0.01,
	currency.XQN:        0.02,
	currency.RBBT:       0.01,
	currency.RAC:        10,
	currency.RADI:       0.01,
	currency.RAD:        0.002,
	currency.RAI:        0.01,
	currency.XRA:        0.002,
	currency.RATIO:      0.002,
	currency.RCN:        0.01,
	currency.REA:        10,
	currency.RCX:        0.01,
	currency.RDD:        0.002,
	currency.REE:        0.01,
	currency.REC:        0.01,
	currency.REQ:        10,
	currency.RMS:        0.002,
	currency.RBIT:       0.01,
	currency.RNC:        0.002,
	currency.R:          1,
	currency.REV:        0.01,
	currency.RH:         0.01,
	currency.XRL:        1,
	currency.RICE:       0.002,
	currency.RICHX:      0.002,
	currency.RID:        0.01,
	currency.RIDE:       0.01,
	currency.RBT:        0.002,
	currency.RING:       0.01,
	currency.RIO:        0.01,
	currency.RISE:       0.2,
	currency.ROCKET:     0.01,
	currency.RPC:        0.01,
	currency.ROS:        0.002,
	currency.ROYAL:      0.01,
	currency.RSGP:       0.01,
	currency.RBIES:      0.002,
	currency.RUBIT:      0.002,
	currency.RBY:        0.01,
	currency.RUC:        0.01,
	currency.RUPX:       0.01,
	currency.RUP:        0.01,
	currency.RUST:       0.01,
	currency.SFE:        0.01,
	currency.SLS:        0.002,
	currency.SMSR:       0.002,
	currency.RONIN:      0.002,
	currency.SAN:        0.05,
	currency.STV:        0.002,
	currency.HIFUN:      0.002,
	currency.MAD:        0.002,
	currency.SANDG:      0.002,
	currency.STO:        0.01,
	currency.SCAN:       0.01,
	currency.SCITW:      0.002,
	currency.SCRPT:      0.01,
	currency.SCRT:       0.002,
	currency.SED:        0.002,
	currency.SEEDS:      0.002,
	currency.B2X:        0.1,
	currency.SEL:        0.01,
	currency.SLFI:       0.002,
	currency.SSC:        0.002,
	currency.SMBR:       0.2,
	currency.SEN:        0.002,
	currency.SENT:       10,
	currency.SRNT:       10,
	currency.SEV:        0.01,
	currency.SP:         0.01,
	currency.SXC:        0.002,
	currency.GELD:       0.05,
	currency.SHDW:       0.01,
	currency.SDC:        0.02,
	currency.SAK:        0.002,
	currency.SHRP:       0.01,
	currency.SHELL:      0.002,
	currency.SH:         0.01,
	currency.SHORTY:     0.01,
	currency.SHREK:      0.01,
	currency.SHRM:       0.002,
	currency.SIB:        0.002,
	currency.SIGT:       0.01,
	currency.SLCO:       0.01,
	currency.SIGU:       0.002,
	currency.SIX:        0.002,
	currency.SJW:        0.002,
	currency.SKB:        0.002,
	currency.SW:         0.01,
	currency.SLEEP:      0.01,
	currency.SLING:      0.01,
	currency.SMART:      0.01,
	currency.SMC:        0.002,
	currency.SMT:        10,
	currency.SMF:        0.01,
	currency.SOCC:       0.01,
	currency.SCL:        0.05,
	currency.SDAO:       10,
	currency.SOLAR:      0.01,
	currency.SOLO:       0.002,
	currency.SCT:        0.01,
	currency.SONG:       0.01,
	currency.SNM:        0.05,
	currency.ALTCOM:     0.01,
	currency.SPHTX:      10,
	currency.SOUL:       0.01,
	currency.SPC:        0.002,
	currency.SPACE:      0.002,
	currency.SBT:        0.01,
	currency.SPEC:       0.002,
	currency.SPX:        0.002,
	currency.SCS:        0.01,
	currency.SPORT:      0.01,
	currency.SPT:        0.002,
	currency.SPR:        0.002,
	currency.SPEX:       0.002,
	currency.SQL:        0.002,
	currency.SBIT:       0.002,
	currency.STHR:       0.002,
	currency.STALIN:     0.01,
	currency.STAR:       0.01,
	currency.STA:        0.01,
	currency.START:      0.02,
	currency.STP:        0.002,
	currency.SNT:        10,
	currency.PNK:        0.002,
	currency.STEPS:      0.002,
	currency.STK:        0.002,
	currency.STONK:      0.01,
	currency.STORJ:      0.05,
	currency.STORM:      10,
	currency.STS:        0.002,
	currency.STRP:       0.002,
	currency.STY:        10,
	currency.SUB:        0.01,
	currency.XMT:        0.002,
	currency.SNC:        1,
	currency.SSTC:       0.01,
	currency.SBTC:       0.1,
	currency.SUPER:      0.002,
	currency.SRND:       0.002,
	currency.STRB:       0.02,
	currency.M1:         0.002,
	currency.SPM:        0.01,
	currency.BUCKS:      0.002,
	currency.TOKEN:      0.01,
	currency.SWT:        0.05,
	currency.SWEET:      0.002,
	currency.SWING:      0.002,
	currency.CHSB:       10,
	currency.SIC:        0.002,
	currency.SDP:        0.002,
	currency.XSY:        0.002,
	currency.SYNX:       0.01,
	currency.SNRG:       0.002,
	currency.SYS:        0.002,
	currency.TAG:        0.01,
	currency.TAGR:       0.002,
	currency.TAJ:        0.01,
	currency.TAK:        0.01,
	currency.TAKE:       0.01,
	currency.TAM:        0.002,
	currency.XTO:        0.01,
	currency.TAP:        0.01,
	currency.TLE:        0.01,
	currency.TSE:        0.01,
	currency.TLEX:       0.01,
	currency.TAXI:       10,
	currency.TCN:        0.01,
	currency.TDFB:       0.002,
	currency.TEAM:       0.01,
	currency.TECH:       0.01,
	currency.TEC:        0.002,
	currency.TEK:        0.002,
	currency.TB:         0.002,
	currency.TLX:        10,
	currency.TELL:       0.01,
	currency.TENNET:     0.002,
	currency.PAY:        0.002,
	currency.TES:        0.002,
	currency.TRA:        0.01,
	currency.TGS:        10,
	currency.XVE:        0.01,
	currency.TCR:        0.01,
	currency.GCC:        0.002,
	currency.MAY:        0.01,
	currency.THOM:       0.01,
	currency.TIA:        0.002,
	currency.TIDE:       0.01,
	currency.TNT:        0.05,
	currency.TIE:        10,
	currency.TIT:        0.002,
	currency.TTC:        0.002,
	currency.TODAY:      0.01,
	currency.TBX:        10,
	currency.TKN:        0.01,
	currency.TDS:        10,
	currency.TLOSH:      0.01,
	currency.TOKC:       0.01,
	currency.TMRW:       0.01,
	currency.TOOL:       0.01,
	currency.TCX:        0.002,
	currency.TOT:        0.01,
	currency.TX:         0.002,
	currency.TRANSF:     0.002,
	currency.TRAP:       0.01,
	currency.TBCX:       0.01,
	currency.TRICK:      0.002,
	currency.TPG:        0.01,
	currency.TRX:        300,
	currency.TFL:        10,
	currency.TRUMP:      0.002,
	currency.TNG:        0.002,
	currency.TUR:        0.002,
	currency.TWERK:      0.002,
	currency.TWIST:      0.002,
	currency.TWO:        0.002,
	currency.UCASH:      10,
	currency.UAE:        0.01,
	currency.XBU:        0.01,
	currency.UBQ:        0.01,
	currency.U:          0.002,
	currency.UDOWN:      0.01,
	currency.GAIN:       0.01,
	currency.USC:        0.01,
	currency.UMC:        0.1,
	currency.UNF:        0.01,
	currency.UNIFY:      0.01,
	currency.UKG:        10,
	currency.USDE:       0.01,
	currency.UBTC:       0.1,
	currency.UIS:        0.01,
	currency.UNIT:       0.002,
	currency.UNI:        0.01,
	currency.UXC:        0.01,
	currency.URC:        0.002,
	currency.XUP:        0.002,
	currency.UFR:        10,
	currency.URO:        0.002,
	currency.UTLE:       0.002,
	currency.VAL:        0.02,
	currency.VPRC:       0.01,
	currency.VAPOR:      0.002,
	currency.VCOIN:      0.002,
	currency.VEC:        0.002,
	currency.VEC2:       0.01,
	currency.VLT:        0.01,
	currency.VENE:       0.01,
	currency.VNTX:       0.01,
	currency.VTN:        0.002,
	currency.XVG:        0.002,
	currency.CRED:       10,
	currency.VERS:       0.01,
	currency.VTC:        0.2,
	currency.VTX:        0.002,
	currency.VIA:        0.002,
	currency.VTY:        0.01,
	currency.VIP:        0.002,
	currency.VISIO:      0.01,
	currency.VK:         2,
	currency.VOL:        0.002,
	currency.VOYA:       0.002,
	currency.VPN:        0.002,
	currency.VSL:        0.01,
	currency.XVS:        0.01,
	currency.VTL:        0.01,
	currency.VULC:       0.01,
	currency.VVI:        10,
	currency.WGR:        0.01,
	currency.WAM:        0.01,
	currency.WARP:       0.002,
	currency.WASH:       0.01,
	currency.WAVES:      0.002,
	currency.WGO:        0.01,
	currency.WAY:        0.01,
	currency.WCASH:      0.01,
	currency.WEALTH:     0.002,
	currency.WEEK:       0.01,
	currency.WHO:        0.5,
	currency.WIC:        0.05,
	currency.WBB:        0.002,
	currency.WINE:       0.01,
	currency.WINK:       0.01,
	currency.WISC:       0.01,
	currency.WITCH:      0.01,
	currency.WMC:        0.01,
	currency.WOMEN:      0.01,
	currency.WOK:        0.01,
	currency.WRC:        10,
	currency.WRT:        0.01,
	currency.XCO:        0.002,
	currency.X2:         0.002,
	currency.XNX:        0.002,
	currency.XAU:        0.002,
	currency.XAV:        0.01,
	currency.XDE2:       0.002,
	currency.XDE:        0.002,
	currency.XIOS:       0.01,
	currency.XOC:        0.01,
	currency.XSSX:       0.002,
	currency.XBY:        0.01,
	currency.YAC:        0.01,
	currency.YMC:        0.01,
	currency.YAY:        0.01,
	currency.YBC:        0.002,
	currency.YES:        0.01,
	currency.YOB2X:      0.01,
	currency.YOVI:       0.002,
	currency.ZYD:        0.01,
	currency.ZEC:        0.02,
	currency.ZECD:       0.01,
	currency.ZEIT:       0.002,
	currency.ZENI:       0.01,
	currency.ZET2:       0.002,
	currency.ZET:        0.002,
	currency.ZMC:        0.002,
	currency.ZIRK:       0.002,
	currency.ZLQ:        0.01,
	currency.ZNE:        0.01,
	currency.ZONTO:      0.05,
	currency.ZOOM:       0.002,
	currency.ZRC:        0.01,
	currency.ZUR:        0.002,
}
