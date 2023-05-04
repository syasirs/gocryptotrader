package stream

import (
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fill"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream/buffer"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

// WrapperWebsocket defines a return type for websocket connections via the interface
// wrapper for routine processing
type WrapperWebsocket struct {
	canUseAuthenticatedEndpoints bool
	enabled                      bool
	Init                         bool
	connected                    bool
	connecting                   bool
	verbose                      bool
	dataMonitorRunning           bool
	trafficTimeout               time.Duration
	connectionMonitorDelay       time.Duration
	proxyAddr                    string
	runningURL                   string
	exchangeName                 string
	m                            sync.Mutex
	connectionMutex              sync.RWMutex
	subscriptionMutex            sync.Mutex
	DataHandler                  chan interface{}
	ToRoutine                    chan interface{}
	Match                        *Match

	// connectedAssetTypesFlag holds a list of asset type connections
	connectedAssetTypesFlag asset.Item
	// shutdown synchronises shutdown event across routines
	ShutdownC chan asset.Item

	Wg *sync.WaitGroup
	// Orderbook is a local buffer of orderbooks
	Orderbook buffer.Orderbook
	// Trade is a notifier of occurring trades
	Trade trade.Trade

	// Fills is a notifier of occurring fills
	Fills fill.Fills

	// trafficAlert monitors if there is a halt in traffic throughput
	TrafficAlert chan struct{}

	// ReadMessageErrors will received all errors from ws.ReadMessage() and
	// verify if its a disconnection
	ReadMessageErrors chan error
	features          *protocol.Features

	// Latency reporter
	ExchangeLevelReporter Reporter

	// AssetTypeWebsockets defines a map of asset type item to corresponding websocket class
	AssetTypeWebsockets map[asset.Item]*Websocket
}

// NewWrapper creates a new websocket wrapper instance
func NewWrapper() *WrapperWebsocket {
	return &WrapperWebsocket{
		Init:                true,
		DataHandler:         make(chan interface{}),
		ToRoutine:           make(chan interface{}, defaultJobBuffer),
		TrafficAlert:        make(chan struct{}),
		ReadMessageErrors:   make(chan error),
		AssetTypeWebsockets: make(map[asset.Item]*Websocket),
		ShutdownC:           make(chan asset.Item),
		Match:               NewMatch(),
	}
}
