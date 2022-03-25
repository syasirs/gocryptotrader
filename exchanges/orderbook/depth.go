package orderbook

import (
	"errors"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/dispatch"
	"github.com/thrasher-corp/gocryptotrader/exchanges/alert"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// ErrOrderbookInvalid defines an error for when the orderbook is invalid and
// should not be trusted
var ErrOrderbookInvalid = errors.New("orderbook data integrity compromised")

// Depth defines a linked list of orderbook items
type Depth struct {
	asks
	bids

	// unexported stack of nodes
	stack *stack

	alert.Notice

	mux *dispatch.Mux
	id  uuid.UUID

	options
	m sync.Mutex
}

// NewDepth returns a new depth item
func NewDepth(id uuid.UUID) *Depth {
	return &Depth{
		stack: newStack(),
		id:    id,
		mux:   service.Mux,
	}
}

// Publish alerts any subscribed routines using a dispatch mux
func (d *Depth) Publish() {
	ob, err := d.Retrieve()
	if err != nil {
		log.Errorf(log.ExchangeSys, "Cannot publish orderbook update to mux %v", err)
	}
	err = d.mux.Publish([]uuid.UUID{d.id}, ob)
	if err != nil {
		log.Errorf(log.ExchangeSys, "Cannot publish orderbook update to mux %v", err)
	}
}

// GetAskLength returns length of asks
func (d *Depth) GetAskLength() (int, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return 0, ErrOrderbookInvalid
	}
	return d.asks.length, nil
}

// GetBidLength returns length of bids
func (d *Depth) GetBidLength() (int, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return 0, ErrOrderbookInvalid
	}
	return d.bids.length, nil
}

// Retrieve returns the orderbook base a copy of the underlying linked list
// spread
func (d *Depth) Retrieve() (*Base, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return nil, ErrOrderbookInvalid
	}
	return &Base{
		Bids:             d.bids.retrieve(),
		Asks:             d.asks.retrieve(),
		Exchange:         d.exchange,
		Asset:            d.asset,
		Pair:             d.pair,
		LastUpdated:      d.lastUpdated,
		LastUpdateID:     d.lastUpdateID,
		PriceDuplication: d.priceDuplication,
		IsFundingRate:    d.isFundingRate,
		VerifyOrderbook:  d.VerifyOrderbook,
	}, nil
}

// TotalBidAmounts returns the total amount of bids and the total orderbook
// bids value
func (d *Depth) TotalBidAmounts() (liquidity, value float64, err error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return 0, 0, ErrOrderbookInvalid
	}
	liquidity, value = d.bids.amount()
	return liquidity, value, nil
}

// TotalAskAmounts returns the total amount of asks and the total orderbook
// asks value
func (d *Depth) TotalAskAmounts() (liquidity, value float64, err error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return 0, 0, ErrOrderbookInvalid
	}
	liquidity, value = d.asks.amount()
	return liquidity, value, nil
}

// LoadSnapshot flushes the bids and asks with a snapshot
func (d *Depth) LoadSnapshot(bids, asks []Item, lastUpdateID int64, lastUpdated time.Time, updateByREST bool) {
	d.m.Lock()
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	d.restSnapshot = updateByREST
	d.bids.load(bids, d.stack)
	d.asks.load(asks, d.stack)
	d.isValid = true
	d.Alert()
	d.m.Unlock()
}

// invalidate flushes all values back to zero so as to not allow strategy
// traversal on compromised data.
func (d *Depth) invalidate() {
	d.lastUpdateID = 0
	d.lastUpdated = time.Time{}
	d.bids.load(nil, d.stack)
	d.asks.load(nil, d.stack)
	d.isValid = false
	d.Alert()
}

// Invalidate flushes all values back to zero so as to not allow strategy
// traversal on compromised data.
func (d *Depth) Invalidate() {
	d.m.Lock()
	d.invalidate()
	d.m.Unlock()
}

// IsValid returns if the underlying book is valid.
func (d *Depth) IsValid() bool {
	d.m.Lock()
	valid := d.isValid
	d.m.Unlock()
	return valid
}

// UpdateBidAskByPrice updates the bid and ask spread by supplied updates, this
// will trim total length of depth level to a specified supplied number
func (d *Depth) UpdateBidAskByPrice(bidUpdts, askUpdts Items, maxDepth int, lastUpdateID int64, lastUpdated time.Time) {
	if len(bidUpdts) == 0 && len(askUpdts) == 0 {
		return
	}
	d.m.Lock()
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	tn := getNow()
	if len(bidUpdts) != 0 {
		d.bids.updateInsertByPrice(bidUpdts, d.stack, maxDepth, tn)
	}
	if len(askUpdts) != 0 {
		d.asks.updateInsertByPrice(askUpdts, d.stack, maxDepth, tn)
	}
	d.Alert()
	d.m.Unlock()
}

// UpdateBidAskByID amends details by ID
func (d *Depth) UpdateBidAskByID(bidUpdts, askUpdts Items, lastUpdateID int64, lastUpdated time.Time) error {
	if len(bidUpdts) == 0 && len(askUpdts) == 0 {
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if len(bidUpdts) != 0 {
		err := d.bids.updateByID(bidUpdts)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	if len(askUpdts) != 0 {
		err := d.asks.updateByID(askUpdts)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	d.Alert()
	return nil
}

// DeleteBidAskByID deletes a price level by ID
func (d *Depth) DeleteBidAskByID(bidUpdts, askUpdts Items, bypassErr bool, lastUpdateID int64, lastUpdated time.Time) error {
	if len(bidUpdts) == 0 && len(askUpdts) == 0 {
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if len(bidUpdts) != 0 {
		err := d.bids.deleteByID(bidUpdts, d.stack, bypassErr)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	if len(askUpdts) != 0 {
		err := d.asks.deleteByID(askUpdts, d.stack, bypassErr)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	d.Alert()
	return nil
}

// InsertBidAskByID inserts new updates
func (d *Depth) InsertBidAskByID(bidUpdts, askUpdts Items, lastUpdateID int64, lastUpdated time.Time) error {
	if len(bidUpdts) == 0 && len(askUpdts) == 0 {
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if len(bidUpdts) != 0 {
		err := d.bids.insertUpdates(bidUpdts, d.stack)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	if len(askUpdts) != 0 {
		err := d.asks.insertUpdates(askUpdts, d.stack)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	d.Alert()
	return nil
}

// UpdateInsertByID updates or inserts by ID at current price level.
func (d *Depth) UpdateInsertByID(bidUpdts, askUpdts Items, lastUpdateID int64, lastUpdated time.Time) error {
	if len(bidUpdts) == 0 && len(askUpdts) == 0 {
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if len(bidUpdts) != 0 {
		err := d.bids.updateInsertByID(bidUpdts, d.stack)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	if len(askUpdts) != 0 {
		err := d.asks.updateInsertByID(askUpdts, d.stack)
		if err != nil {
			d.invalidate()
			return err
		}
	}
	d.Alert()
	d.lastUpdateID = lastUpdateID
	d.lastUpdated = lastUpdated
	return nil
}

// AssignOptions assigns the initial options for the depth instance
func (d *Depth) AssignOptions(b *Base) {
	d.m.Lock()
	d.options = options{
		exchange:         b.Exchange,
		pair:             b.Pair,
		asset:            b.Asset,
		lastUpdated:      b.LastUpdated,
		lastUpdateID:     b.LastUpdateID,
		priceDuplication: b.PriceDuplication,
		isFundingRate:    b.IsFundingRate,
		VerifyOrderbook:  b.VerifyOrderbook,
		restSnapshot:     b.RestSnapshot,
		idAligned:        b.IDAlignment,
	}
	d.m.Unlock()
}

// GetName returns name of exchange
func (d *Depth) GetName() string {
	d.m.Lock()
	defer d.m.Unlock()
	return d.exchange
}

// IsRestSnapshot returns if the depth item was updated via REST
func (d *Depth) IsRestSnapshot() (bool, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return false, ErrOrderbookInvalid
	}
	return d.restSnapshot, nil
}

// LastUpdateID returns the last Update ID
func (d *Depth) LastUpdateID() (int64, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if !d.isValid {
		return 0, ErrOrderbookInvalid
	}
	return d.lastUpdateID, nil
}

// IsFundingRate returns if the depth is a funding rate
func (d *Depth) IsFundingRate() bool {
	d.m.Lock()
	defer d.m.Unlock()
	return d.isFundingRate
}
