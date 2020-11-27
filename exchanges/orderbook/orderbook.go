package orderbook

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/dispatch"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// Get checks and returns the orderbook given an exchange name and currency pair
// if it exists
func Get(exchange string, p currency.Pair, a asset.Item) (*Base, error) {
	o, err := service.Retrieve(exchange, p, a)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// SubscribeOrderbook subcribes to an orderbook and returns a communication
// channel to stream orderbook data updates
func SubscribeOrderbook(exchange string, p currency.Pair, a asset.Item) (dispatch.Pipe, error) {
	exchange = strings.ToLower(exchange)
	service.RLock()
	defer service.RUnlock()
	book, ok := service.Books[exchange][p.Base.Item][p.Quote.Item][a]
	if !ok {
		return dispatch.Pipe{},
			fmt.Errorf("orderbook item not found for %s %s %s",
				exchange,
				p,
				a)
	}
	return service.mux.Subscribe(book.Main)
}

// SubscribeToExchangeOrderbooks subcribes to all orderbooks on an exchange
func SubscribeToExchangeOrderbooks(exchange string) (dispatch.Pipe, error) {
	exchange = strings.ToLower(exchange)
	service.RLock()
	defer service.RUnlock()
	id, ok := service.Exchange[exchange]
	if !ok {
		return dispatch.Pipe{}, fmt.Errorf("%s exchange orderbooks not found",
			exchange)
	}
	return service.mux.Subscribe(id)
}

// Update stores orderbook data
func (s *Service) Update(b *Base) error {
	name := strings.ToLower(b.ExchangeName)
	s.Lock()
	book, ok := s.Books[name][b.Pair.Base.Item][b.Pair.Quote.Item][b.AssetType]
	if ok {
		book.b.Bids = b.Bids
		book.b.Asks = b.Asks
		book.b.LastUpdated = b.LastUpdated
		ids := append(book.Assoc, book.Main)
		s.Unlock()
		return s.mux.Publish(ids, b)
	}

	switch {
	case s.Books[name] == nil:
		s.Books[name] = make(map[*currency.Item]map[*currency.Item]map[asset.Item]*Book)
		fallthrough
	case s.Books[name][b.Pair.Base.Item] == nil:
		s.Books[name][b.Pair.Base.Item] = make(map[*currency.Item]map[asset.Item]*Book)
		fallthrough
	case s.Books[name][b.Pair.Base.Item][b.Pair.Quote.Item] == nil:
		s.Books[name][b.Pair.Base.Item][b.Pair.Quote.Item] = make(map[asset.Item]*Book)
	}

	err := s.SetNewData(b, name)
	if err != nil {
		s.Unlock()
		return err
	}
	s.Unlock()
	return nil
}

// SetNewData sets new data
func (s *Service) SetNewData(b *Base, fmtName string) error {
	ids, err := s.GetAssociations(b, fmtName)
	if err != nil {
		return err
	}
	singleID, err := s.mux.GetID()
	if err != nil {
		return err
	}

	// Below instigates orderbook item separation so we can ensure, in the event
	// of a simultaneous update via websocket/rest/fix, we don't affect package
	// scoped orderbook data which could result in a potential panic
	cpyBook := *b
	cpyBook.Bids = make([]Item, len(b.Bids))
	copy(cpyBook.Bids, b.Bids)
	cpyBook.Asks = make([]Item, len(b.Asks))
	copy(cpyBook.Asks, b.Asks)

	s.Books[fmtName][b.Pair.Base.Item][b.Pair.Quote.Item][b.AssetType] = &Book{
		b:     &cpyBook,
		Main:  singleID,
		Assoc: ids}
	return nil
}

// GetAssociations links a singular book with it's dispatch associations
func (s *Service) GetAssociations(b *Base, fmtName string) ([]uuid.UUID, error) {
	if b == nil {
		return nil, errors.New("orderbook is nil")
	}

	var ids []uuid.UUID
	exchangeID, ok := s.Exchange[fmtName]
	if !ok {
		var err error
		exchangeID, err = s.mux.GetID()
		if err != nil {
			return nil, err
		}
		s.Exchange[fmtName] = exchangeID
	}

	ids = append(ids, exchangeID)
	return ids, nil
}

// Retrieve gets orderbook data from the slice
func (s *Service) Retrieve(exchange string, p currency.Pair, a asset.Item) (*Base, error) {
	exchange = strings.ToLower(exchange)
	s.RLock()
	defer s.RUnlock()
	if _, ok := s.Books[exchange]; !ok {
		return nil, fmt.Errorf("no orderbooks for %s exchange", exchange)
	}

	if _, ok := s.Books[exchange][p.Base.Item]; !ok {
		return nil, fmt.Errorf("no orderbooks associated with base currency %s",
			p.Base)
	}

	if _, ok := s.Books[exchange][p.Base.Item][p.Quote.Item]; !ok {
		return nil, fmt.Errorf("no orderbooks associated with quote currency %s",
			p.Quote)
	}

	var liveOrderBook *Book
	var ok bool
	if liveOrderBook, ok = s.Books[exchange][p.Base.Item][p.Quote.Item][a]; !ok {
		return nil, fmt.Errorf("no orderbooks associated with asset type %s",
			a)
	}

	localCopyOfAsks := make([]Item, len(s.Books[exchange][p.Base.Item][p.Quote.Item][a].b.Asks))
	localCopyOfBids := make([]Item, len(s.Books[exchange][p.Base.Item][p.Quote.Item][a].b.Bids))
	copy(localCopyOfBids, liveOrderBook.b.Bids)
	copy(localCopyOfAsks, liveOrderBook.b.Asks)

	ob := Base{
		Pair:         liveOrderBook.b.Pair,
		Bids:         localCopyOfBids,
		Asks:         localCopyOfAsks,
		LastUpdated:  liveOrderBook.b.LastUpdated,
		LastUpdateID: liveOrderBook.b.LastUpdateID,
		AssetType:    liveOrderBook.b.AssetType,
		ExchangeName: liveOrderBook.b.ExchangeName,
	}

	return &ob, nil
}

// TotalBidsAmount returns the total amount of bids and the total orderbook
// bids value
func (b *Base) TotalBidsAmount() (amountCollated, total float64) {
	for x := range b.Bids {
		amountCollated += b.Bids[x].Amount
		total += b.Bids[x].Amount * b.Bids[x].Price
	}
	return amountCollated, total
}

// TotalAsksAmount returns the total amount of asks and the total orderbook
// asks value
func (b *Base) TotalAsksAmount() (amountCollated, total float64) {
	for y := range b.Asks {
		amountCollated += b.Asks[y].Amount
		total += b.Asks[y].Amount * b.Asks[y].Price
	}
	return amountCollated, total
}

// Update updates the bids and asks
func (b *Base) Update(bids, asks []Item) {
	b.Bids = bids
	b.Asks = asks
	b.LastUpdated = time.Now()
}

// Verify ensures that the orderbook items are correctly sorted prior to being
// set and will reject any book with incorrect values.
// Bids should always go from a high price to a low price and
// Asks should always go from a low price to a higher price
func (b *Base) Verify() error {
	if len(b.Asks) == 0 && len(b.Bids) == 0 {
		return errNoOrderbook
	}
	for i := range b.Bids {
		if b.Bids[i].Price == 0 {
			return fmt.Errorf(bidLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errPriceNotSet)
		}
		if b.Bids[i].Amount == 0 {
			return fmt.Errorf(bidLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errAmountNotSet)
		}
		if i != 0 {
			if b.Bids[i].Price > b.Bids[i-1].Price {
				return fmt.Errorf(bidLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errOutOfOrder)
			}

			if b.Bids[i].ID != 0 {
				if b.Bids[i].ID == b.Bids[i-1].ID {
					return fmt.Errorf(bidLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errors.New("awwww man"))
				}
				continue
			}

			fmt.Println(b.Bids)
			os.Exit(1)

			if b.Bids[i].ID != 0 && b.Bids[i].Price == b.Bids[i-1].Price {
				return fmt.Errorf(bidLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errDuplication)
			}
		}
	}

	for i := range b.Asks {
		if b.Asks[i].Price == 0 {
			return fmt.Errorf(askLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errPriceNotSet)
		}
		if b.Asks[i].Amount == 0 {
			return fmt.Errorf(askLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errAmountNotSet)
		}
		if i != 0 {
			if b.Asks[i].Price < b.Asks[i-1].Price {
				return fmt.Errorf(askLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errOutOfOrder)
			}

			if b.Asks[i].ID != 0 {
				if b.Asks[i].ID == b.Asks[i-1].ID {
					return fmt.Errorf(askLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errors.New("awwww man"))
				}
				continue
			}

			if b.Asks[i].Price == b.Asks[i-1].Price {
				return fmt.Errorf(askLoadBookFailure, b.ExchangeName, b.Pair, b.AssetType, errDuplication)
			}
		}
	}
	return nil
}

// Process processes incoming orderbooks, creating or updating the orderbook
// list
func (b *Base) Process() error {
	if b.ExchangeName == "" {
		return errExchangeNameUnset
	}

	if b.Pair.IsEmpty() {
		return errPairNotSet
	}

	if b.AssetType.String() == "" {
		return errAssetTypeNotSet
	}

	if b.LastUpdated.IsZero() {
		b.LastUpdated = time.Now()
	}

	err := b.Verify()
	if err != nil {
		return err
	}

	return service.Update(b)
}

// SortAsks sorts ask items to the correct relative order
func SortAsks(d []Item) []Item {
	sort.Sort(byOBPrice(d))
	return d
}

// SortBids sorts bid items to the correct relative order
func SortBids(d []Item) []Item {
	sort.Sort(sort.Reverse(byOBPrice(d)))
	return d
}
