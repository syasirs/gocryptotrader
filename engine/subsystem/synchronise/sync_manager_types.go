package synchronise

import (
	"errors"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine/subsystem"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

const (
	// DefaultWorkers limits the number of sync workers
	DefaultWorkers = 15
	// DefaultTimeoutREST the default time to switch from REST to websocket
	// protocols without a response.
	DefaultTimeoutREST = time.Second * 15
	// DefaultTimeoutWebsocket the default time to switch from websocket to REST
	// protocols without a response.
	DefaultTimeoutWebsocket = time.Minute
	// ManagerName defines a string identifier for the subsystem
	ManagerName = "exchange_syncer"

	defaultChannelBuffer = 10000
	book                 = "%s %s %s %s ORDERBOOK: Bids len: %d Amount: %f %s. Total value: %s Asks len: %d Amount: %f %s. Total value: %s"
)

var (
	// ErrNoItemsEnabled is for when there is not at least one sync item enabled
	// e.g. an orderbook or ticker item.
	ErrNoItemsEnabled = errors.New("no sync items enabled")

	errUnknownSyncType   = errors.New("unknown sync type")
	errExchangeNameUnset = errors.New("exchange name unset")
	errProtocolUnset     = errors.New("protocol unset")
)

// Agent stores the sync agent information on exchange, asset type and pair
// and holds the individual item bases.
type Agent struct {
	Exchange            string
	Asset               asset.Item
	Pair                currency.Pair
	SynchronisationType subsystem.SynchronisationType
	IsUsingWebsocket    bool
	IsUsingREST         bool
	IsProcessing        bool
	LastUpdated         time.Time
	NumErrors           int
	mu                  sync.Mutex
}

// ManagerConfig stores the currency pair synchronisation manager config
type ManagerConfig struct {
	SynchroniseTicker       bool
	SynchroniseOrderbook    bool
	SynchroniseContinuously bool
	TimeoutREST             time.Duration
	TimeoutWebsocket        time.Duration
	NumWorkers              int
	FiatDisplayCurrency     currency.Code
	PairFormatDisplay       currency.PairFormat
	Verbose                 bool
	ExchangeManager         subsystem.ExchangeManager
	WebsocketRPCEnabled     bool
	APIServerManager        subsystem.APIServer
}

// Manager defines the main total currency pair synchronisation subsystem that
// fetches and maintains up to date market data.
type Manager struct {
	initSyncCompleted int32
	started           int32
	initSyncStartTime time.Time
	mu                sync.Mutex
	initSyncWG        sync.WaitGroup

	currencyPairs            map[string]map[*currency.Item]map[*currency.Item]map[asset.Item]map[subsystem.SynchronisationType]*Agent
	tickerBatchLastRequested map[string]map[asset.Item]time.Time
	batchMtx                 sync.Mutex

	ManagerConfig

	createdCounter int64
	removedCounter int64

	orderbookJobs chan RESTJob
	tickerJobs    chan RESTJob
	workerWG      sync.WaitGroup
}

// RESTJob defines a potential REST synchronisation job
type RESTJob struct {
	exch  exchange.IBotExchange
	Pair  currency.Pair
	Asset asset.Item
}
