package stream

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fill"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream/buffer"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
)

// Websocket functionality list and state consts
const (
	// WebsocketNotEnabled alerts of a disabled websocket
	WebsocketNotEnabled                = "exchange_websocket_not_enabled"
	WebsocketNotAuthenticatedUsingRest = "%v - Websocket not authenticated, using REST\n"
	Ping                               = "ping"
	Pong                               = "pong"
	UnhandledMessage                   = " - Unhandled websocket message: "

	// AutoSubscribe defines if the websocket should automatically subscribe
	// to channels on connection.
	AutoSubscribe SubscriptionAllowed = 1
	// DeferSubscribe defines if the websocket should defer subscription to
	// channels until after connection.
	DeferSubscribe SubscriptionAllowed = 0
)

// SubscriptionAllowed defines if the websocket should automatically subscribe
type SubscriptionAllowed uint8
type subscriptionMap map[any]*ChannelSubscription

// Websocket defines a return type for websocket connections via the interface
// wrapper for routine processing
type Websocket struct {
	canUseAuthenticatedEndpoints bool
	enabled                      bool
	Init                         bool
	connected                    bool
	connecting                   bool
	verbose                      bool
	connectionMonitorRunning     bool
	trafficMonitorRunning        bool
	dataMonitorRunning           bool
	trafficTimeout               time.Duration
	connectionMonitorDelay       time.Duration
	proxyAddr                    string
	defaultURL                   string
	defaultURLAuth               string
	runningURL                   string
	runningURLAuth               string
	exchangeName                 string
	m                            sync.Mutex
	fieldMutex                   sync.RWMutex
	connector                    func(ctx context.Context) error

	subscriptionMutex sync.RWMutex
	subscriptions     subscriptionMap
	Subscribe         chan []ChannelSubscription
	Unsubscribe       chan []ChannelSubscription

	// Subscriber function for package defined websocket subscriber
	// functionality
	Subscriber func(context.Context, []ChannelSubscription) error
	// Unsubscriber function for packaged defined websocket unsubscriber
	// functionality
	Unsubscriber func(context.Context, []ChannelSubscription) error
	// GenerateSubs function for package defined websocket generate
	// subscriptions functionality
	GenerateSubs func() ([]ChannelSubscription, error)

	DataHandler chan interface{}
	ToRoutine   chan interface{}

	Match *Match

	// shutdown synchronises shutdown event across routines
	ShutdownC chan struct{}
	Wg        *sync.WaitGroup

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

	// Standard stream connection
	Conn Connection
	// Authenticated stream connection
	AuthConn Connection

	// Latency reporter
	ExchangeLevelReporter Reporter

	// MaxSubScriptionsPerConnection defines the maximum number of
	// subscriptions per connection that is allowed by the exchange.
	MaxSubscriptionsPerConnection int
}

// WebsocketSetup defines variables for setting up a websocket connection
type WebsocketSetup struct {
	ExchangeConfig        *config.Exchange
	DefaultURL            string
	RunningURL            string
	RunningURLAuth        string
	Connector             func(context.Context) error
	Subscriber            func(context.Context, []ChannelSubscription) error
	Unsubscriber          func(context.Context, []ChannelSubscription) error
	GenerateSubscriptions func() ([]ChannelSubscription, error)
	Features              *protocol.Features

	// Local orderbook buffer config values
	OrderbookBufferConfig buffer.Config

	TradeFeed bool

	// Fill data config values
	FillsFeed bool

	// MaxWebsocketSubscriptionsPerConnection defines the maximum number of
	// subscriptions per connection that is allowed by the exchange.
	MaxWebsocketSubscriptionsPerConnection int
}

// WebsocketConnection contains all the data needed to send a message to a WS
// connection
type WebsocketConnection struct {
	Verbose   bool
	connected int32

	// Gorilla websocket does not allow more than one goroutine to utilise
	// writes methods
	writeControl sync.Mutex

	RateLimit    int64
	ExchangeName string
	URL          string
	ProxyURL     string
	Wg           *sync.WaitGroup
	Connection   *websocket.Conn
	ShutdownC    chan struct{}

	Match             *Match
	ResponseMaxLimit  time.Duration
	Traffic           chan struct{}
	readMessageErrors chan error

	Reporter Reporter
}
