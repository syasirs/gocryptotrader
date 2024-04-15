package buffer

import (
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/key"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
)

// Config defines the configuration variables for the websocket buffer; snapshot
// and incremental update orderbook processing.
type Config struct {
	// SortBuffer enables a websocket to sort incoming updates before processing.
	SortBuffer bool
	// SortBufferByUpdateIDs allows the sorting of the buffered updates by their
	// corresponding update IDs.
	SortBufferByUpdateIDs bool
	// UpdateEntriesByID will match by IDs instead of price to perform the an
	// action. e.g. update, delete, insert.
	UpdateEntriesByID bool
	// UpdateIDProgression requires that the new update ID be greater than the
	// prior ID. This will skip processing and not error.
	UpdateIDProgression bool
	// Checksum is a package defined checksum calculation for updated books.
	Checksum func(state *orderbook.Base, checksum uint32) error
}

// Orderbook defines a local cache of orderbooks for amending, appending
// and deleting changes and updates the main store for a stream
type Orderbook struct {
	ob                    map[key.PairAsset]*orderbookHolder
	obBufferLimit         int
	bufferEnabled         bool
	sortBuffer            bool
	sortBufferByUpdateIDs bool // When timestamps aren't provided, an id can help sort
	updateEntriesByID     bool // Use the update IDs to match ob entries
	exchangeName          string
	dataHandler           chan<- interface{}
	verbose               bool

	// updateIDProgression requires that the new update ID be greater than the
	// prior ID. This will skip processing and not error.
	updateIDProgression bool
	// checksum is a package defined checksum calculation for updated books.
	checksum func(state *orderbook.Base, checksum uint32) error

	publishPeriod time.Duration

	mtx sync.RWMutex
}

// orderbookHolder defines a store of pending updates and a pointer to the
// orderbook depth
type orderbookHolder struct {
	ob     *orderbook.Depth
	buffer *[]orderbook.Update
	// Reduces the amount of outbound alerts to the data handler for example
	// coinbasepro can have up too 100 updates per second introducing overhead.
	// The sync agent only requires an alert every 15 seconds for a specific
	// currency.
	ticker   *time.Ticker
	updateID int64

	// initialUpdateCompleted determines initial sync state, the very first
	// update from any pair is critical for initial sync.
	initialUpdateCompleted bool

	// Protect indv holder. TODO: Expose orderbook.Depth mutex to use that
	// instead. This is temporary.
	mtx sync.Mutex
}
