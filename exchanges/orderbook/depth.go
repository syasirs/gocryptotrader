package orderbook

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// Depth defines a linked list of orderbook items
type Depth struct {
	ask linkedList
	bid linkedList

	// Unexported stack of nodes
	stack Stack

	// Change of state to re-check depth list
	wait    chan struct{}
	waiting uint32
	wMtx    sync.Mutex
	// -----

	Exchange string
	Pair     currency.Pair
	Asset    asset.Item

	// RestSnapshot defines if the depth was applied via the REST protocol thus
	// an update cannot be applied via websocket mechanics and a resubscription
	// would need to take place to maintain book integrity
	restSnapshot bool

	lastUpdated time.Time
	sync.Mutex
}

// GetAskLength returns length of asks
func (d *Depth) GetAskLength() int {
	d.Lock()
	defer d.Unlock()
	return d.ask.length
}

// GetBidLength returns length of bids
func (d *Depth) GetBidLength() int {
	d.Lock()
	defer d.Unlock()
	return d.bid.length
}

// AddBid adds a bid to the list
func (d *Depth) AddBid(i Item) error {
	d.Lock()
	defer d.Unlock()
	return d.bid.Add(func(i Item) bool { return true }, i, &d.stack)
}

// Retrieve gets stuff
func (d *Depth) Retrieve() (bids, asks Items) {
	d.Lock()
	defer d.Unlock()
	return d.bid.Retrieve(), d.ask.Retrieve()
}

// // AddBids adds a collection of bids to the linked list
// func (d *Depth) AddBids(i Item) error {
// 	d.Lock()
// 	defer d.Unlock()
// 	n := d.stack.Pop()
// 	n.value = i
// 	d.bid.Add(func(i Item) bool { return true }, n)
// 	return nil
// }

// RemoveBidByPrice removes a bid
func (d *Depth) RemoveBidByPrice(price float64) error {
	// d.Lock()
	// defer d.Unlock()
	// n, err := d.bid.Remove(func(i Item) bool { return i.Price == price })
	// if err != nil {
	// 	return err
	// }
	// d.stack.Push(n)
	return nil
}

// DisplayBids does a helpful display!!! YAY!
func (d *Depth) DisplayBids() {
	d.Lock()
	defer d.Unlock()
	d.bid.Display()
}

// alert establishes state change for depth to all waiting routines
func (d *Depth) alert() {
	if !atomic.CompareAndSwapUint32(&d.waiting, 1, 0) {
		// return if no waiting routines
		return
	}
	d.wMtx.Lock()
	close(d.wait)
	d.wait = make(chan struct{})
	d.wMtx.Unlock()
}

// kicker defines a channel that allows a system to kick routine away from
// waiting for a change on the linked list
type kicker chan struct{}

// timeInForce allows a kick
func timeInForce(t time.Duration) kicker {
	ch := make(chan struct{})
	go func(ch chan<- struct{}) {
		time.Sleep(t)
		close(ch)
	}(ch)
	return ch
}

// Wait pauses routine until depth change has been established
func (d *Depth) Wait(kick <-chan struct{}) bool {
	d.wMtx.Lock()
	if d.wait == nil {
		d.wait = make(chan struct{})
	}
	atomic.StoreUint32(&d.waiting, 1)
	d.wMtx.Unlock()
	select {
	case <-d.wait:
		return true
	case <-kick:
		return false
	}
}

// TotalBidAmounts returns the total amount of bids and the total orderbook
// bids value
func (d *Depth) TotalBidAmounts() (liquidity, value float64) {
	d.Lock()
	defer d.Unlock()
	return d.bid.Amount()
}

// TotalAskAmounts returns the total amount of asks and the total orderbook
// asks value
func (d *Depth) TotalAskAmounts() (liquidity, value float64) {
	d.Lock()
	defer d.Unlock()
	return d.ask.Amount()
}

// LoadSnapshot flushes the bids and asks with a snapshot
func (d *Depth) LoadSnapshot(bids, asks []Item) (err error) {
	d.Lock()
	defer func() {
		// TODO: Restructure locks as this will alert routines after slip ring actuates
		if err != nil {
			d.flush()
		}
		d.Unlock()
	}()

	d.bid.Load(bids, &d.stack)
	d.ask.Load(asks, &d.stack)

	// Update occurred, alert routines
	d.alert()
	return nil
}

// Flush attempts to flush bid and ask sides
func (d *Depth) Flush() {
	d.Lock()
	d.flush()
	d.Unlock()
}

// Process processes incoming orderbook snapshots
func (d *Depth) Process(bids, asks Items) error {
	d.Lock()
	defer d.Unlock() // TODO: Restructure locks as this will alert routines
	// after slip ring actuates
	d.bid.Load(bids, &d.stack)
	d.ask.Load(asks, &d.stack)
	d.alert()
	return nil
}

// flush will pop entire bid and ask node chain onto stack when invalidated or
// required for full flush when resubscribing
func (d *Depth) flush() {
	d.bid.Load(nil, &d.stack)
	d.ask.Load(nil, &d.stack)
}

type outOfOrder func(float64, float64) bool

// UpdateBidAskByPrice updates the bid and ask spread by supplied updates
func (d *Depth) UpdateBidAskByPrice(bid, ask Items, maxDepth int) error {
	var errs common.Errors
	d.Lock()
	defer d.Unlock()
	err := d.bid.updateInsertBidsByPrice(bid, &d.stack, maxDepth)
	if err != nil {
		errs = append(errs, err)
	}

	err = d.bid.updateInsertBidsByPrice(ask, &d.stack, maxDepth)
	if err != nil {
		errs = append(errs, err)
	}
	d.alert()
	if errs != nil {
		return errs
	}
	return nil
}

// UpdateBidAskByID amends details by ID
func (d *Depth) UpdateBidAskByID(bid, ask Items) error {
	d.Lock()
	defer d.Unlock()
	err := d.bid.updateByID(bid)
	if err != nil {
		return err
	}

	err = d.ask.updateByID(ask)
	if err != nil {
		return err
	}
	d.alert()
	return nil
}

// DeleteBidAskByID deletes a price level by ID
func (d *Depth) DeleteBidAskByID(bid, ask Items, bypassErr bool) error {
	d.Lock()
	defer d.Unlock()

	err := d.bid.deleteByID(bid, &d.stack, bypassErr)
	if err != nil {
		return err
	}

	err = d.ask.deleteByID(ask, &d.stack, bypassErr)
	if err != nil {
		return err
	}

	d.alert()
	return nil
}

// InsertBidAskByID inserts new updates
func (d *Depth) InsertBidAskByID(bid, ask Items) {
	d.Lock()
	d.bid.insertUpdatesBid(bid, &d.stack)
	d.ask.insertUpdatesAsk(ask, &d.stack)
	d.alert()
	d.Unlock()
}

// UpdateInsertByID ...
func (d *Depth) UpdateInsertByID(bidUpdates, askUpdates Items) {
	d.Lock()
	d.bid.updateInsertByIDBid(bidUpdates, &d.stack)
	d.ask.updateInsertByIDAsk(askUpdates, &d.stack)
	d.alert()
	d.Unlock()
}
