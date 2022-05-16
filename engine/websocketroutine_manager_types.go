package engine

import (
	"errors"
	"sync"

	"github.com/thrasher-corp/gocryptotrader/currency"
)

var (
	errNilOrderManager                 = errors.New("nil order manager received")
	errNilCurrencyPairSyncer           = errors.New("nil currency pair syncer received")
	errNilCurrencyConfig               = errors.New("nil currency config received")
	errNilCurrencyPairFormat           = errors.New("nil currency pair format received")
	errNilWebsocketDataHandlerFunction = errors.New("websocket data handler function is nil")
	errNilWebsocket                    = errors.New("websocket is nil")
	errRoutineManagerNotStarted        = errors.New("websocket routine manager not started")
)

// websocketRoutineManager is used to process websocket updates from a unified location
type websocketRoutineManager struct {
	started         int32
	verbose         bool
	exchangeManager iExchangeManager
	orderManager    iOrderManager
	syncer          iCurrencyPairSyncer
	currencyConfig  *currency.Config
	shutdown        chan struct{}
	dataHandlers    []WebsocketDataHandler
	wg              sync.WaitGroup
	mu              sync.RWMutex
}

// WebsocketDataHandler defines a function signature for a function that handles
// data coming from websocket connections.
type WebsocketDataHandler func(service string, incoming interface{}) error
