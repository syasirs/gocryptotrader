package portfolio

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/exchange"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/portfolio/risk"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/fill"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/order"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/signal"
	"github.com/thrasher-corp/gocryptotrader/backtester/funding"
	"github.com/thrasher-corp/gocryptotrader/currency"
	gctexchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	gctorder "github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

const notEnoughFundsTo = "not enough funds to"

var (
	errInvalidDirection     = errors.New("invalid direction")
	errRiskManagerUnset     = errors.New("risk manager unset")
	errSizeManagerUnset     = errors.New("size manager unset")
	errAssetUnset           = errors.New("asset unset")
	errCurrencyPairUnset    = errors.New("currency pair unset")
	errExchangeUnset        = errors.New("exchange unset")
	errNegativeRiskFreeRate = errors.New("received negative risk free rate")
	errNoPortfolioSettings  = errors.New("no portfolio settings")
	errNoHoldings           = errors.New("no holdings found")
	errHoldingsNoTimestamp  = errors.New("holding with unset timestamp received")
	errHoldingsAlreadySet   = errors.New("holding already set")
	errUnsetFuturesTracker  = errors.New("portfolio settings futures tracker unset")
)

// Portfolio stores all holdings and rules to assess orders, allowing the portfolio manager to
// modify, accept or reject strategy signals
type Portfolio struct {
	riskFreeRate              decimal.Decimal
	sizeManager               SizeHandler
	riskManager               risk.Handler
	exchangeAssetPairSettings map[string]map[asset.Item]map[currency.Pair]*Settings
}

// Handler contains all functions expected to operate a portfolio manager
type Handler interface {
	OnSignal(signal.Event, *exchange.Settings, funding.IFundReserver) (*order.Order, error)
	OnFill(fill.Event, funding.IFundReleaser) (fill.Event, error)
	GetLatestOrderSnapshotForEvent(common.EventHandler) (compliance.Snapshot, error)
	GetLatestOrderSnapshots() ([]compliance.Snapshot, error)
	ViewHoldingAtTimePeriod(common.EventHandler) (*holdings.Holding, error)
	setHoldingsForOffset(*holdings.Holding, bool) error
	UpdateHoldings(common.DataEventHandler, funding.IFundReleaser) error
	GetComplianceManager(string, asset.Item, currency.Pair) (*compliance.Manager, error)
	GetPositions(common.EventHandler) ([]gctorder.PositionStats, error)
	TrackFuturesOrder(fill.Event, funding.IFundReleaser) (*PNLSummary, error)
	UpdatePNL(common.EventHandler, decimal.Decimal) error
	GetLatestPNLForEvent(common.EventHandler) (*PNLSummary, error)
	GetLatestPNLs() []PNLSummary
	CheckLiquidationStatus(common.DataEventHandler, funding.ICollateralReader, *PNLSummary) error
	CreateLiquidationOrdersForExchange(common.DataEventHandler, funding.IFundingManager) ([]order.Event, error)
	Reset()
}

// SizeHandler is the interface to help size orders
type SizeHandler interface {
	SizeOrder(order.Event, decimal.Decimal, *exchange.Settings) (*order.Order, decimal.Decimal, error)
}

// Settings holds all important information for the portfolio manager
// to assess purchasing decisions
type Settings struct {
	BuySideSizing     exchange.MinMax
	SellSideSizing    exchange.MinMax
	Leverage          exchange.Leverage
	HoldingsSnapshots []holdings.Holding
	ComplianceManager compliance.Manager
	Exchange          gctexchange.IBotExchange
	FuturesTracker    *gctorder.MultiPositionTracker
}

// PNLSummary holds a PNL result along with
// exchange details
type PNLSummary struct {
	Exchange           string
	Item               asset.Item
	Pair               currency.Pair
	CollateralCurrency currency.Code
	Offset             int64
	Result             gctorder.PNLResult
}

// IPNL defines an interface for an implementation
// to retrieve PNL from a position
type IPNL interface {
	GetUnrealisedPNL() BasicPNLResult
	GetRealisedPNL() BasicPNLResult
	GetCollateralCurrency() currency.Code
	GetDirection() gctorder.Side
	GetPositionStatus() gctorder.Status
}

// BasicPNLResult holds the time and the pnl
// of a position
type BasicPNLResult struct {
	Currency currency.Code
	Time     time.Time
	PNL      decimal.Decimal
}
