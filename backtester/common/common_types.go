package common

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	// CandleStr is a config readable data type to tell the backtester to retrieve candle data
	CandleStr = "candle"
	// TradeStr is a config readable data type to tell the backtester to retrieve trade data
	TradeStr = "trade"
)

// custom order side declarations for backtesting processing and
// decision-making
const (
	// DoNothing is an explicit signal for the backtester to not perform an action
	// based upon indicator results
	DoNothing order.Side = "DO NOTHING"
	// TransferredFunds is a status signal to do nothing
	TransferredFunds order.Side = "TRANSFERRED FUNDS"
	// CouldNotBuy is flagged when a BUY  signal is raised in the strategy/signal phase, but the
	// portfolio manager or exchange cannot place an order
	CouldNotBuy order.Side = "COULD NOT BUY"
	// CouldNotSell is flagged when a SELL  signal is raised in the strategy/signal phase, but the
	// portfolio manager or exchange cannot place an order
	CouldNotSell  order.Side = "COULD NOT SELL"
	CouldNotShort order.Side = "COULD NOT SHORT"
	CouldNotLong  order.Side = "COULD NOT LONG"
	// ClosePosition is used to signal a complete closure
	// of any exposure of a position
	// This will handle any amount of exposure, no need to calculate how
	// much to close
	ClosePosition      order.Side = "CLOSE POSITION"
	CouldNotCloseShort order.Side = "COULD NOT CLOSE SHORT"
	CouldNotCloseLong  order.Side = "COULD NOT CLOSE LONG"
	Liquidated         order.Side = "LIQUIDATED"
	// MissingData is signalled during the strategy/signal phase when data has been identified as missing
	// No buy or sell events can occur
	MissingData order.Side = "MISSING DATA"
)

// DataCandle is an int64 representation of a candle data type
const (
	DataCandle = iota
	DataTrade
)

var (
	// ErrNilArguments is a common error response to highlight that nils were passed in
	// when they should not have been
	ErrNilArguments = errors.New("received nil argument(s)")
	// ErrNilEvent is a common error for whenever a nil event occurs when it shouldn't have
	ErrNilEvent = errors.New("nil event received")
	// ErrInvalidDataType occurs when an invalid data type is defined in the config
	ErrInvalidDataType = errors.New("invalid datatype received")
)

// EventHandler interface implements required GetTime() & Pair() return
type EventHandler interface {
	GetOffset() int64
	SetOffset(int64)
	IsEvent() bool
	GetTime() time.Time
	Pair() currency.Pair
	GetExchange() string
	GetInterval() kline.Interval
	GetAssetType() asset.Item
	GetReason() string
	GetClosePrice() decimal.Decimal
	AppendReason(string)
	AppendReasonf(string, ...interface{})
}

// custom subloggers for backtester use
var (
	Backtester         *log.SubLogger
	Setup              *log.SubLogger
	Strategy           *log.SubLogger
	Config             *log.SubLogger
	Portfolio          *log.SubLogger
	Exchange           *log.SubLogger
	Fill               *log.SubLogger
	Report             *log.SubLogger
	Statistics         *log.SubLogger
	CurrencyStatistics *log.SubLogger
	FundingStatistics  *log.SubLogger
	Holdings           *log.SubLogger
	Data               *log.SubLogger
)

// DataEventHandler interface used for loading and interacting with Data
type DataEventHandler interface {
	EventHandler
	GetUnderlyingPair() currency.Pair
	GetClosePrice() decimal.Decimal
	GetHighPrice() decimal.Decimal
	GetLowPrice() decimal.Decimal
	GetOpenPrice() decimal.Decimal
}

// Directioner dictates the side of an order
type Directioner interface {
	SetDirection(side order.Side)
	GetDirection() order.Side
}

// colours to display for the terminal output
var (
	ColourDefault  = "\u001b[0m"
	ColourGreen    = "\033[38;5;157m"
	ColourWhite    = "\033[38;5;255m"
	ColourGrey     = "\033[38;5;240m"
	ColourDarkGrey = "\033[38;5;243m"
	ColourH1       = "\033[38;5;33m"
	ColourH2       = "\033[38;5;39m"
	ColourH3       = "\033[38;5;45m"
	ColourH4       = "\033[38;5;51m"
	ColourSuccess  = "\033[38;5;40m"
	ColourInfo     = "\u001B[32m"
	ColourDebug    = "\u001B[34m"
	ColourWarn     = "\u001B[33m"
	ColourError    = "\033[38;5;196m"
)

var (
	// LogoLines contains the lovely GCT logo
	LogoLines = []string{
		"                                                                                ",
		"                               " + ColourWhite + "@@@@@@@@@@@@@@@@@                                ",
		"                            " + ColourWhite + "@@@@@@@@@@@@@@@@@@@@@@@    " + ColourGrey + ",,,,,," + ColourWhite + "                   ",
		"                           " + ColourWhite + "@@@@@@@@" + ColourGrey + ",,,,,    " + ColourWhite + "@@@@@@@@@" + ColourGrey + ",,,,,,,," + ColourWhite + "                   ",
		"                         " + ColourWhite + "@@@@@@@@" + ColourGrey + ",,,,,,,       " + ColourWhite + "@@@@@@@" + ColourGrey + ",,,,,,," + ColourWhite + "                   ",
		"                         " + ColourWhite + "@@@@@@" + ColourGrey + "(,,,,,,,,      " + ColourGrey + ",," + ColourWhite + "@@@@@@@" + ColourGrey + ",,,,,," + ColourWhite + "                   ",
		"                      " + ColourGrey + ",," + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,   #,,,,,,,,,,,,,,,,,," + ColourWhite + "                   ",
		"                   " + ColourGrey + ",,,,*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,,,,,,,,,,,,,,,,,," + ColourGreen + "%%%%%%%" + ColourWhite + "                ",
		"                " + ColourGrey + ",,,,,,,*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,,,,,," + ColourGreen + "%%%%%" + ColourGrey + " ,,,,,," + ColourGrey + "%" + ColourGreen + "%%%%%%" + ColourWhite + "                 ",
		"               " + ColourGrey + ",,,,,,,,*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,,," + ColourGreen + "%%%%%%%%%%%%%%%%%%" + ColourGrey + "#" + ColourGreen + "%%" + ColourGrey + "                  ",
		"                 " + ColourGrey + ",,,,,,*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,," + ColourGreen + "%%%" + ColourGrey + " ,,,,," + ColourGreen + "%%%%%%%%" + ColourGrey + ",,,,,                   ",
		"                    " + ColourGrey + ",,,*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,," + ColourGreen + "%%" + ColourGrey + ",,  ,,,,,,," + ColourWhite + "@" + ColourGreen + "*%%," + ColourWhite + "@" + ColourGrey + ",,,,,,                   ",
		"                       " + ColourGrey + "*" + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,     " + ColourGrey + ",,,,," + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,," + ColourWhite + "       ,            ",
		"                         " + ColourWhite + "@@@@@@" + ColourGrey + ",,,,,,,,,        " + ColourWhite + "@@@@@@@" + ColourGrey + ",,,,,," + ColourWhite + "                   ",
		"                         " + ColourWhite + "@@@@@@@@" + ColourGrey + ",,,,,,,       " + ColourWhite + "@@@@@@@" + ColourGrey + ",,,,,,," + ColourWhite + "                   ",
		"                           " + ColourWhite + "@@@@@@@@@" + ColourGrey + ",,,,    " + ColourWhite + "@@@@@@@@@" + ColourGrey + "#,,,,,,," + ColourWhite + "                   ",
		"                            " + ColourWhite + "@@@@@@@@@@@@@@@@@@@@@@@     " + ColourGrey + "*,,,," + ColourWhite + "                   ",
		"                                " + ColourWhite + "@@@@@@@@@@@@@@@@" + ColourDefault + "                                ",
	}
)

// ASCIILogo is a sweet logo that is optionally printed to the command line window
const ASCIILogo = `
   ______      ______                 __      ______               __         
  / ____/___  / ____/______  ______  / /_____/_  __/________ _____/ /__  _____
 / / __/ __ \/ /   / ___/ / / / __ \/ __/ __ \/ / / ___/ __  / __  / _ \/ ___/
/ /_/ / /_/ / /___/ /  / /_/ / /_/ / /_/ /_/ / / / /  / /_/ / /_/ /  __/ /
\____/\____/\____/_/   \__, / .___/\__/\____/_/ /_/   \__,_/\__,_/\___/_/
                       /___/
                 ____             __   __            __           
                / __ )____ ______/ /__/ /____  _____/ /____  _____
               / __  / __  / ___/ //_/ __/ _ \/ ___/ __/ _ \/ ___/
              / /_/ / /_/ / /__/ ,< / /_/  __(__  ) /_/  __/ /
             /_____/\__,_/\___/_/|_|\__/\___/____/\__/\___/_/

`
