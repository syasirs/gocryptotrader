package wsorderbook

import (
	"errors"
	"fmt"
	"sort"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
)

// Setup sets private variables
func (w *WebsocketOrderbookLocal) Setup(obBufferLimit int, bufferEnabled, sortBuffer, sortBufferByUpdateIDs, updateEntriesByID bool, exchangeName string) {
	w.obBufferLimit = obBufferLimit
	w.bufferEnabled = bufferEnabled
	w.sortBuffer = sortBuffer
	w.sortBufferByUpdateIDs = sortBufferByUpdateIDs
	w.updateEntriesByID = updateEntriesByID
	w.exchangeName = exchangeName
}

// Update updates a local cache using bid targets and ask targets then updates
// main orderbook
// Volume == 0; deletion at price target
// Price target not found; append of price target
// Price target found; amend volume of price target
func (w *WebsocketOrderbookLocal) Update(orderbookUpdate *WebsocketOrderbookUpdate) error {
	if (orderbookUpdate.Bids == nil && orderbookUpdate.Asks == nil) ||
		(len(orderbookUpdate.Bids) == 0 && len(orderbookUpdate.Asks) == 0) {
		return fmt.Errorf("%v cannot have bids and ask targets both nil", w.exchangeName)
	}
	w.m.Lock()
	defer w.m.Unlock()
	obLookup, ok := w.ob[orderbookUpdate.CurrencyPair][orderbookUpdate.AssetType]
	if !ok {
		return fmt.Errorf("ob.Base could not be found for Exchange %s CurrencyPair: %s AssetType: %s",
			w.exchangeName,
			orderbookUpdate.CurrencyPair.String(),
			orderbookUpdate.AssetType)
	}
	if w.bufferEnabled {
		overBufferLimit := w.processBufferUpdate(obLookup, orderbookUpdate)
		if !overBufferLimit {
			return nil
		}
	} else {
		w.processObUpdate(obLookup, orderbookUpdate)
	}
	err := obLookup.Process()
	if err != nil {
		return err
	}
	if w.bufferEnabled {
		// Reset the buffer
		w.buffer[orderbookUpdate.CurrencyPair][orderbookUpdate.AssetType] = nil
	}
	return nil
}

func (w *WebsocketOrderbookLocal) processBufferUpdate(o *orderbook.Base, orderbookUpdate *WebsocketOrderbookUpdate) bool {
	if w.buffer == nil {
		w.buffer = make(map[currency.Pair]map[asset.Item][]*WebsocketOrderbookUpdate)
	}
	if w.buffer[orderbookUpdate.CurrencyPair] == nil {
		w.buffer[orderbookUpdate.CurrencyPair] = make(map[asset.Item][]*WebsocketOrderbookUpdate)
	}
	bufferLookup := w.buffer[orderbookUpdate.CurrencyPair][orderbookUpdate.AssetType]
	if len(bufferLookup) <= w.obBufferLimit {
		bufferLookup = append(bufferLookup, orderbookUpdate)
		if len(bufferLookup) < w.obBufferLimit {
			w.buffer[orderbookUpdate.CurrencyPair][orderbookUpdate.AssetType] = bufferLookup
			return false
		}
	}
	if w.sortBuffer {
		// sort by last updated to ensure each update is in order
		if w.sortBufferByUpdateIDs {
			sort.Slice(bufferLookup, func(i, j int) bool {
				return bufferLookup[i].UpdateID < bufferLookup[j].UpdateID
			})
		} else {
			sort.Slice(bufferLookup, func(i, j int) bool {
				return bufferLookup[i].UpdateTime.Before(bufferLookup[j].UpdateTime)
			})
		}
	}
	for i := 0; i < len(bufferLookup); i++ {
		w.processObUpdate(o, bufferLookup[i])
	}
	w.buffer[orderbookUpdate.CurrencyPair][orderbookUpdate.AssetType] = bufferLookup
	return true
}

func (w *WebsocketOrderbookLocal) processObUpdate(o *orderbook.Base, orderbookUpdate *WebsocketOrderbookUpdate) {
	if w.updateEntriesByID {
		w.updateByIDAndAction(o, orderbookUpdate)
	} else {
		w.updateAsksByPrice(o, orderbookUpdate)
		w.updateBidsByPrice(o, orderbookUpdate)
	}
}

func (w *WebsocketOrderbookLocal) updateAsksByPrice(o *orderbook.Base, base *WebsocketOrderbookUpdate) {
	for j := 0; j < len(base.Asks); j++ {
		found := false
		for k := 0; k < len(o.Asks); k++ {
			if o.Asks[k].Price == base.Asks[j].Price {
				found = true
				if base.Asks[j].Amount == 0 {
					o.Asks = append(o.Asks[:k], o.Asks[k+1:]...)
					break
				}
				o.Asks[k].Amount = base.Asks[j].Amount
				break
			}
		}
		if !found {
			if base.Asks[j].Amount == 0 {
				continue
			}
			o.Asks = append(o.Asks, base.Asks[j])
		}
	}
	sort.Slice(o.Asks, func(i, j int) bool {
		return o.Asks[i].Price < o.Asks[j].Price
	})
}

func (w *WebsocketOrderbookLocal) updateBidsByPrice(o *orderbook.Base, base *WebsocketOrderbookUpdate) {
	for j := 0; j < len(base.Bids); j++ {
		found := false
		for k := 0; k < len(o.Bids); k++ {
			if o.Bids[k].Price == base.Bids[j].Price {
				found = true
				if base.Bids[j].Amount == 0 {
					o.Bids = append(o.Bids[:k], o.Bids[k+1:]...)
					break
				}
				o.Bids[k].Amount = base.Bids[j].Amount
				break
			}
		}
		if !found {
			if base.Bids[j].Amount == 0 {
				continue
			}
			o.Bids = append(o.Bids, base.Bids[j])
		}
	}
	sort.Slice(o.Bids, func(i, j int) bool {
		return o.Bids[i].Price > o.Bids[j].Price
	})
}

// updateByIDAndAction will receive an action to execute against the orderbook
// it will then match by IDs instead of price to perform the action
func (w *WebsocketOrderbookLocal) updateByIDAndAction(o *orderbook.Base, orderbookUpdate *WebsocketOrderbookUpdate) {
	switch orderbookUpdate.Action {
	case "update":
		for x := range orderbookUpdate.Bids {
			for y := range o.Bids {
				if o.Bids[y].ID == orderbookUpdate.Bids[x].ID {
					o.Bids[y].Amount = orderbookUpdate.Bids[x].Amount
					break
				}
			}
		}
		for x := range orderbookUpdate.Asks {
			for y := range o.Asks {
				if o.Asks[y].ID == orderbookUpdate.Asks[x].ID {
					o.Asks[y].Amount = orderbookUpdate.Asks[x].Amount
					break
				}
			}
		}
	case "delete":
		for x := range orderbookUpdate.Bids {
			for y := 0; y < len(o.Bids); y++ {
				if o.Bids[y].ID == orderbookUpdate.Bids[x].ID {
					o.Bids = append(o.Bids[:y], o.Bids[y+1:]...)
					break
				}
			}
		}
		for x := range orderbookUpdate.Asks {
			for y := 0; y < len(o.Asks); y++ {
				if o.Asks[y].ID == orderbookUpdate.Asks[x].ID {
					o.Asks = append(o.Asks[:y], o.Asks[y+1:]...)
					break
				}
			}
		}
	case "insert":
		o.Bids = append(o.Bids, orderbookUpdate.Bids...)
		sort.Slice(o.Bids, func(i, j int) bool {
			return o.Bids[i].Price > o.Bids[j].Price
		})

		o.Asks = append(o.Asks, orderbookUpdate.Asks...)
		sort.Slice(o.Asks, func(i, j int) bool {
			return o.Asks[i].Price < o.Asks[j].Price
		})
	}
}

// LoadSnapshot loads initial snapshot of ob data, overwrite allows full
// ob to be completely rewritten because the exchange is a doing a full
// update not an incremental one
func (w *WebsocketOrderbookLocal) LoadSnapshot(newOrderbook *orderbook.Base) error {
	if len(newOrderbook.Asks) == 0 || len(newOrderbook.Bids) == 0 {
		return fmt.Errorf("%v snapshot ask and bids are nil", w.exchangeName)
	}

	if newOrderbook.Pair.IsEmpty() {
		return errors.New("websocket orderbook pair unset")
	}

	if newOrderbook.AssetType.String() == "" {
		return errors.New("websocket orderbook asset type unset")
	}

	if newOrderbook.ExchangeName == "" {
		return errors.New("websocket orderbook exchange name unset")
	}

	w.m.Lock()
	defer w.m.Unlock()
	if w.ob == nil {
		w.ob = make(map[currency.Pair]map[asset.Item]*orderbook.Base)
	}
	if w.ob[newOrderbook.Pair] == nil {
		w.ob[newOrderbook.Pair] = make(map[asset.Item]*orderbook.Base)
	}
	fullObLookup := w.ob[newOrderbook.Pair][newOrderbook.AssetType]
	if fullObLookup != nil &&
		(len(fullObLookup.Asks) > 0 ||
			len(fullObLookup.Bids) > 0) {
		fullObLookup = newOrderbook
		return newOrderbook.Process()
	}
	w.ob[newOrderbook.Pair][newOrderbook.AssetType] = newOrderbook
	return newOrderbook.Process()
}

// GetOrderbook use sparingly. Modifying anything here will ruin hash calculation and cause problems
func (w *WebsocketOrderbookLocal) GetOrderbook(p currency.Pair, assetType asset.Item) *orderbook.Base {
	w.m.Lock()
	defer w.m.Unlock()
	return w.ob[p][assetType]
}

// FlushCache flushes w.ob data to be garbage collected and refreshed when a
// connection is lost and reconnected
func (w *WebsocketOrderbookLocal) FlushCache() {
	w.m.Lock()
	w.ob = nil
	w.buffer = nil
	w.m.Unlock()
}
