package order

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/validate"
)

const (
	orderSubmissionValidSides = Buy | Sell | Bid | Ask | Long | Short
	shortSide                 = Short | Sell | Ask
	longSide                  = Long | Buy | Bid
	inactiveStatuses          = Filled | Cancelled | InsufficientBalance | MarketUnavailable | Rejected | PartiallyCancelled | Expired | Closed | AnyStatus | Cancelling | Liquidated
	activeStatuses            = Active | Open | PartiallyFilled | New | PendingCancel | Hidden | AutoDeleverage | Pending
	bypassSideFilter          = UnknownSide | AnySide
	bypassTypeFilter          = UnknownType | AnyType
)

var (
	errTimeInForceConflict     = errors.New("multiple time in force options applied")
	errUnrecognisedOrderSide   = errors.New("unrecognised order side")
	errUnrecognisedOrderType   = errors.New("unrecognised order type")
	errUnrecognisedOrderStatus = errors.New("unrecognised order status")
)

// Validate checks the supplied data and returns whether or not it's valid
func (s *Submit) Validate(opt ...validate.Checker) error {
	if s == nil {
		return ErrSubmissionIsNil
	}

	if s.Pair.IsEmpty() {
		return ErrPairIsEmpty
	}

	if s.AssetType == asset.Empty {
		return ErrAssetNotSet
	}

	if s.Side == UnknownSide || orderSubmissionValidSides&s.Side != s.Side {
		return ErrSideIsInvalid
	}

	if s.Type != Market && s.Type != Limit {
		return ErrTypeIsInvalid
	}

	if s.ImmediateOrCancel && s.FillOrKill {
		return errTimeInForceConflict
	}

	if s.Amount == 0 && s.QuoteAmount == 0 {
		return fmt.Errorf("submit validation error %w, amount and quote amount cannot be zero", ErrAmountIsInvalid)
	}

	if s.Amount < 0 {
		return fmt.Errorf("submit validation error base %w, suppled: %v", ErrAmountIsInvalid, s.Amount)
	}

	if s.QuoteAmount < 0 {
		return fmt.Errorf("submit validation error quote %w, suppled: %v", ErrAmountIsInvalid, s.QuoteAmount)
	}

	if s.Type == Limit && s.Price <= 0 {
		return ErrPriceMustBeSetIfLimitOrder
	}

	for _, o := range opt {
		err := o.Check()
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateOrderFromDetail Will update an order detail (used in order management)
// by comparing passed in and existing values
func (d *Detail) UpdateOrderFromDetail(m *Detail) {
	var updated bool
	if d.ImmediateOrCancel != m.ImmediateOrCancel {
		d.ImmediateOrCancel = m.ImmediateOrCancel
		updated = true
	}
	if d.HiddenOrder != m.HiddenOrder {
		d.HiddenOrder = m.HiddenOrder
		updated = true
	}
	if d.FillOrKill != m.FillOrKill {
		d.FillOrKill = m.FillOrKill
		updated = true
	}
	if m.Price > 0 && m.Price != d.Price {
		d.Price = m.Price
		updated = true
	}
	if m.Amount > 0 && m.Amount != d.Amount {
		d.Amount = m.Amount
		updated = true
	}
	if m.LimitPriceUpper > 0 && m.LimitPriceUpper != d.LimitPriceUpper {
		d.LimitPriceUpper = m.LimitPriceUpper
		updated = true
	}
	if m.LimitPriceLower > 0 && m.LimitPriceLower != d.LimitPriceLower {
		d.LimitPriceLower = m.LimitPriceLower
		updated = true
	}
	if m.TriggerPrice > 0 && m.TriggerPrice != d.TriggerPrice {
		d.TriggerPrice = m.TriggerPrice
		updated = true
	}
	if m.QuoteAmount > 0 && m.QuoteAmount != d.QuoteAmount {
		d.QuoteAmount = m.QuoteAmount
		updated = true
	}
	if m.ExecutedAmount > 0 && m.ExecutedAmount != d.ExecutedAmount {
		d.ExecutedAmount = m.ExecutedAmount
		updated = true
	}
	if m.Fee > 0 && m.Fee != d.Fee {
		d.Fee = m.Fee
		updated = true
	}
	if m.AccountID != "" && m.AccountID != d.AccountID {
		d.AccountID = m.AccountID
		updated = true
	}
	if m.PostOnly != d.PostOnly {
		d.PostOnly = m.PostOnly
		updated = true
	}
	if !m.Pair.IsEmpty() && !m.Pair.Equal(d.Pair) {
		// TODO: Add a check to see if the original pair is empty as well, but
		// error if it is changing from BTC-USD -> LTC-USD.
		d.Pair = m.Pair
		updated = true
	}
	if m.Leverage != 0 && m.Leverage != d.Leverage {
		d.Leverage = m.Leverage
		updated = true
	}
	if m.ClientID != "" && m.ClientID != d.ClientID {
		d.ClientID = m.ClientID
		updated = true
	}
	if m.WalletAddress != "" && m.WalletAddress != d.WalletAddress {
		d.WalletAddress = m.WalletAddress
		updated = true
	}
	if m.Type != UnknownType && m.Type != d.Type {
		d.Type = m.Type
		updated = true
	}
	if m.Side != UnknownSide && m.Side != d.Side {
		d.Side = m.Side
		updated = true
	}
	if m.Status != UnknownStatus && m.Status != d.Status {
		d.Status = m.Status
		updated = true
	}
	if m.AssetType != asset.Empty && m.AssetType != d.AssetType {
		d.AssetType = m.AssetType
		updated = true
	}
	for x := range m.Trades {
		var found bool
		for y := range d.Trades {
			if d.Trades[y].TID != m.Trades[x].TID {
				continue
			}
			found = true
			if d.Trades[y].Fee != m.Trades[x].Fee {
				d.Trades[y].Fee = m.Trades[x].Fee
				updated = true
			}
			if m.Trades[x].Price != 0 && d.Trades[y].Price != m.Trades[x].Price {
				d.Trades[y].Price = m.Trades[x].Price
				updated = true
			}
			if d.Trades[y].Side != m.Trades[x].Side {
				d.Trades[y].Side = m.Trades[x].Side
				updated = true
			}
			if d.Trades[y].Type != m.Trades[x].Type {
				d.Trades[y].Type = m.Trades[x].Type
				updated = true
			}
			if d.Trades[y].Description != m.Trades[x].Description {
				d.Trades[y].Description = m.Trades[x].Description
				updated = true
			}
			if m.Trades[x].Amount != 0 && d.Trades[y].Amount != m.Trades[x].Amount {
				d.Trades[y].Amount = m.Trades[x].Amount
				updated = true
			}
			if d.Trades[y].Timestamp != m.Trades[x].Timestamp {
				d.Trades[y].Timestamp = m.Trades[x].Timestamp
				updated = true
			}
			if d.Trades[y].IsMaker != m.Trades[x].IsMaker {
				d.Trades[y].IsMaker = m.Trades[x].IsMaker
				updated = true
			}
		}
		if !found {
			d.Trades = append(d.Trades, m.Trades[x])
			updated = true
		}
		m.RemainingAmount -= m.Trades[x].Amount
	}
	if m.RemainingAmount > 0 && m.RemainingAmount != d.RemainingAmount {
		d.RemainingAmount = m.RemainingAmount
		updated = true
	}
	if updated {
		if d.LastUpdated.Equal(m.LastUpdated) {
			d.LastUpdated = time.Now()
		} else {
			d.LastUpdated = m.LastUpdated
		}
	}
	if d.Exchange == "" {
		d.Exchange = m.Exchange
	}
	if d.ID == "" {
		d.ID = m.ID
	}
	if d.InternalOrderID == "" {
		d.InternalOrderID = m.InternalOrderID
	}
}

// UpdateOrderFromModify Will update an order detail (used in order management)
// by comparing passed in and existing values
func (d *Detail) UpdateOrderFromModify(m *Modify) {
	var updated bool
	if m.ID != "" && d.ID != m.ID {
		d.ID = m.ID
		updated = true
	}
	if d.ImmediateOrCancel != m.ImmediateOrCancel {
		d.ImmediateOrCancel = m.ImmediateOrCancel
		updated = true
	}
	if d.HiddenOrder != m.HiddenOrder {
		d.HiddenOrder = m.HiddenOrder
		updated = true
	}
	if d.FillOrKill != m.FillOrKill {
		d.FillOrKill = m.FillOrKill
		updated = true
	}
	if m.Price > 0 && m.Price != d.Price {
		d.Price = m.Price
		updated = true
	}
	if m.Amount > 0 && m.Amount != d.Amount {
		d.Amount = m.Amount
		updated = true
	}
	if m.LimitPriceUpper > 0 && m.LimitPriceUpper != d.LimitPriceUpper {
		d.LimitPriceUpper = m.LimitPriceUpper
		updated = true
	}
	if m.LimitPriceLower > 0 && m.LimitPriceLower != d.LimitPriceLower {
		d.LimitPriceLower = m.LimitPriceLower
		updated = true
	}
	if m.TriggerPrice > 0 && m.TriggerPrice != d.TriggerPrice {
		d.TriggerPrice = m.TriggerPrice
		updated = true
	}
	if m.QuoteAmount > 0 && m.QuoteAmount != d.QuoteAmount {
		d.QuoteAmount = m.QuoteAmount
		updated = true
	}
	if m.ExecutedAmount > 0 && m.ExecutedAmount != d.ExecutedAmount {
		d.ExecutedAmount = m.ExecutedAmount
		updated = true
	}
	if m.Fee > 0 && m.Fee != d.Fee {
		d.Fee = m.Fee
		updated = true
	}
	if m.AccountID != "" && m.AccountID != d.AccountID {
		d.AccountID = m.AccountID
		updated = true
	}
	if m.PostOnly != d.PostOnly {
		d.PostOnly = m.PostOnly
		updated = true
	}
	if !m.Pair.IsEmpty() && !m.Pair.Equal(d.Pair) {
		// TODO: Add a check to see if the original pair is empty as well, but
		// error if it is changing from BTC-USD -> LTC-USD.
		d.Pair = m.Pair
		updated = true
	}
	if m.Leverage != 0 && m.Leverage != d.Leverage {
		d.Leverage = m.Leverage
		updated = true
	}
	if m.ClientID != "" && m.ClientID != d.ClientID {
		d.ClientID = m.ClientID
		updated = true
	}
	if m.WalletAddress != "" && m.WalletAddress != d.WalletAddress {
		d.WalletAddress = m.WalletAddress
		updated = true
	}
	if m.Type != UnknownType && m.Type != d.Type {
		d.Type = m.Type
		updated = true
	}
	if m.Side != UnknownSide && m.Side != d.Side {
		d.Side = m.Side
		updated = true
	}
	if m.Status != UnknownStatus && m.Status != d.Status {
		d.Status = m.Status
		updated = true
	}
	if m.AssetType != asset.Empty && m.AssetType != d.AssetType {
		d.AssetType = m.AssetType
		updated = true
	}
	for x := range m.Trades {
		var found bool
		for y := range d.Trades {
			if d.Trades[y].TID != m.Trades[x].TID {
				continue
			}
			found = true
			if d.Trades[y].Fee != m.Trades[x].Fee {
				d.Trades[y].Fee = m.Trades[x].Fee
				updated = true
			}
			if m.Trades[x].Price != 0 && d.Trades[y].Price != m.Trades[x].Price {
				d.Trades[y].Price = m.Trades[x].Price
				updated = true
			}
			if d.Trades[y].Side != m.Trades[x].Side {
				d.Trades[y].Side = m.Trades[x].Side
				updated = true
			}
			if d.Trades[y].Type != m.Trades[x].Type {
				d.Trades[y].Type = m.Trades[x].Type
				updated = true
			}
			if d.Trades[y].Description != m.Trades[x].Description {
				d.Trades[y].Description = m.Trades[x].Description
				updated = true
			}
			if m.Trades[x].Amount != 0 && d.Trades[y].Amount != m.Trades[x].Amount {
				d.Trades[y].Amount = m.Trades[x].Amount
				updated = true
			}
			if d.Trades[y].Timestamp != m.Trades[x].Timestamp {
				d.Trades[y].Timestamp = m.Trades[x].Timestamp
				updated = true
			}
			if d.Trades[y].IsMaker != m.Trades[x].IsMaker {
				d.Trades[y].IsMaker = m.Trades[x].IsMaker
				updated = true
			}
		}
		if !found {
			d.Trades = append(d.Trades, m.Trades[x])
			updated = true
		}
		m.RemainingAmount -= m.Trades[x].Amount
	}
	if m.RemainingAmount > 0 && m.RemainingAmount != d.RemainingAmount {
		d.RemainingAmount = m.RemainingAmount
		updated = true
	}
	if updated {
		if d.LastUpdated.Equal(m.LastUpdated) {
			d.LastUpdated = time.Now()
		} else {
			d.LastUpdated = m.LastUpdated
		}
	}
}

// MatchFilter will return true if a detail matches the filter criteria
// empty elements are ignored
func (d *Detail) MatchFilter(f *Filter) bool {
	if f.Exchange != "" && !strings.EqualFold(d.Exchange, f.Exchange) {
		return false
	}
	if f.AssetType != asset.Empty && d.AssetType != f.AssetType {
		return false
	}
	if !f.Pair.IsEmpty() && !d.Pair.Equal(f.Pair) {
		return false
	}
	if f.ID != "" && d.ID != f.ID {
		return false
	}
	if f.Type != UnknownType && f.Type != AnyType && d.Type != f.Type {
		return false
	}
	if f.Side != UnknownSide && f.Side != AnySide && d.Side != f.Side {
		return false
	}
	if f.Status != UnknownStatus && f.Status != AnyStatus && d.Status != f.Status {
		return false
	}
	if f.ClientOrderID != "" && d.ClientOrderID != f.ClientOrderID {
		return false
	}
	if f.ClientID != "" && d.ClientID != f.ClientID {
		return false
	}
	if f.InternalOrderID != "" && d.InternalOrderID != f.InternalOrderID {
		return false
	}
	if f.AccountID != "" && d.AccountID != f.AccountID {
		return false
	}
	if f.WalletAddress != "" && d.WalletAddress != f.WalletAddress {
		return false
	}
	return true
}

// IsActive returns true if an order has a status that indicates it is currently
// available on the exchange
func (d *Detail) IsActive() bool {
	return d.Status != UnknownStatus &&
		d.Amount > 0 &&
		d.Amount > d.ExecutedAmount &&
		activeStatuses&d.Status == d.Status
}

// IsInactive returns true if an order has a status that indicates it is
// currently not available on the exchange
func (d *Detail) IsInactive() bool {
	return d.Amount <= 0 ||
		d.Amount <= d.ExecutedAmount ||
		d.Status.IsInactive()
}

// IsInactive returns true if the status indicates it is
// currently not available on the exchange
func (s Status) IsInactive() bool {
	return inactiveStatuses&s == s
}

// GenerateInternalOrderID sets a new V4 order ID or a V5 order ID if
// the V4 function returns an error
func (d *Detail) GenerateInternalOrderID() {
	if d.InternalOrderID == "" {
		var id uuid.UUID
		id, err := uuid.NewV4()
		if err != nil {
			id = uuid.NewV5(uuid.UUID{}, d.ID)
		}
		d.InternalOrderID = id.String()
	}
}

// CopyToPointer will return the address of a new copy of the order Detail
// WARNING: DO NOT DEREFERENCE USE METHOD Copy().
func (d *Detail) CopyToPointer() *Detail {
	c := d.Copy()
	return &c
}

// Copy makes a full copy of underlying details NOTE: This is Addressable.
func (d *Detail) Copy() Detail {
	c := *d
	if len(d.Trades) > 0 {
		c.Trades = make([]TradeHistory, len(d.Trades))
		copy(c.Trades, d.Trades)
	}
	return c
}

// CopyPointerOrderSlice returns a copy of all order detail and returns a slice
// of pointers.
func CopyPointerOrderSlice(old []*Detail) []*Detail {
	copySlice := make([]*Detail, len(old))
	for x := range old {
		copySlice[x] = old[x].CopyToPointer()
	}
	return copySlice
}

// String implements the stringer interface
func (t Type) String() string {
	switch t {
	case AnyType:
		return "ANY"
	case Limit:
		return "LIMIT"
	case Market:
		return "MARKET"
	case PostOnly:
		return "POST_ONLY"
	case ImmediateOrCancel:
		return "IMMEDIATE_OR_CANCEL"
	case Stop:
		return "STOP"
	case StopLimit:
		return "STOP LIMIT"
	case StopMarket:
		return "STOP MARKET"
	case TakeProfit:
		return "TAKE PROFIT"
	case TakeProfitMarket:
		return "TAKE PROFIT MARKET"
	case TrailingStop:
		return "TRAILING_STOP"
	case FillOrKill:
		return "FOK"
	case IOS:
		return "IOS"
	case Liquidation:
		return "LIQUIDATION"
	case Trigger:
		return "TRIGGER"
	default:
		return "UNKNOWN"
	}
}

// Lower returns the type lower case string
func (t Type) Lower() string {
	return strings.ToLower(t.String())
}

// Title returns the type titleized, eg "Limit"
func (t Type) Title() string {
	return strings.Title(strings.ToLower(t.String())) // nolint:staticcheck // Ignore Title usage warning
}

// String implements the stringer interface
func (s Side) String() string {
	switch s {
	case Buy:
		return "BUY"
	case Sell:
		return "SELL"
	case Bid:
		return "BID"
	case Ask:
		return "ASK"
	case Long:
		return "LONG"
	case Short:
		return "SHORT"
	case AnySide:
		return "ANY"
	case ClosePosition:
		return "CLOSE POSITION"
		// Backtester signal types below.
	case DoNothing:
		return "DO NOTHING"
	case TransferredFunds:
		return "TRANSFERRED FUNDS"
	case CouldNotBuy:
		return "COULD NOT BUY"
	case CouldNotSell:
		return "COULD NOT SELL"
	case CouldNotShort:
		return "COULD NOT SHORT"
	case CouldNotLong:
		return "COULD NOT LONG"
	case CouldNotCloseShort:
		return "COULD NOT CLOSE SHORT"
	case CouldNotCloseLong:
		return "COULD NOT CLOSE LONG"
	case MissingData:
		return "MISSING DATA"
	default:
		return "UNKNOWN"
	}
}

// Lower returns the side lower case string
func (s Side) Lower() string {
	return strings.ToLower(s.String())
}

// Title returns the side titleized, eg "Buy"
func (s Side) Title() string {
	return strings.Title(strings.ToLower(s.String())) // nolint:staticcheck // Ignore Title usage warning
}

// IsShort returns if the side is short
func (s Side) IsShort() bool {
	return s != UnknownSide && shortSide&s == s
}

// IsLong returns if the side is long
func (s Side) IsLong() bool {
	return s != UnknownSide && longSide&s == s
}

// String implements the stringer interface
func (s Status) String() string {
	switch s {
	case AnyStatus:
		return "ANY"
	case New:
		return "NEW"
	case Active:
		return "ACTIVE"
	case PartiallyCancelled:
		return "PARTIALLY_CANCELLED"
	case PartiallyFilled:
		return "PARTIALLY_FILLED"
	case Filled:
		return "FILLED"
	case Cancelled:
		return "CANCELLED"
	case PendingCancel:
		return "PENDING_CANCEL"
	case InsufficientBalance:
		return "INSUFFICIENT_BALANCE"
	case MarketUnavailable:
		return "MARKET_UNAVAILABLE"
	case Rejected:
		return "REJECTED"
	case Expired:
		return "EXPIRED"
	case Hidden:
		return "HIDDEN"
	case Open:
		return "OPEN"
	case AutoDeleverage:
		return "ADL"
	case Closed:
		return "CLOSED"
	case Pending:
		return "PENDING"
	case Cancelling:
		return "CANCELLING"
	default:
		return "UNKNOWN"
	}
}

// InferCostsAndTimes infer order costs using execution information and times
// when available
func (d *Detail) InferCostsAndTimes() {
	if d.CostAsset.IsEmpty() {
		d.CostAsset = d.Pair.Quote
	}

	if d.LastUpdated.IsZero() {
		if d.CloseTime.IsZero() {
			d.LastUpdated = d.Date
		} else {
			d.LastUpdated = d.CloseTime
		}
	}

	if d.ExecutedAmount <= 0 {
		return
	}

	if d.AverageExecutedPrice == 0 {
		if d.Cost != 0 {
			d.AverageExecutedPrice = d.Cost / d.ExecutedAmount
		} else {
			d.AverageExecutedPrice = d.Price
		}
	}
	if d.Cost == 0 {
		d.Cost = d.AverageExecutedPrice * d.ExecutedAmount
	}
}

// FilterOrdersBySide removes any order details that don't match the order
// status provided
func FilterOrdersBySide(orders *[]Detail, side Side) {
	if bypassSideFilter&side == side || len(*orders) == 0 {
		return
	}

	target := 0
	for i := range *orders {
		if (*orders)[i].Side == side {
			(*orders)[target] = (*orders)[i]
			target++
		}
	}
	*orders = (*orders)[:target]
}

// FilterOrdersByType removes any order details that don't match the order type
// provided
func FilterOrdersByType(orders *[]Detail, orderType Type) {
	if bypassTypeFilter&orderType == orderType || len(*orders) == 0 {
		return
	}

	target := 0
	for i := range *orders {
		if (*orders)[i].Type == orderType {
			(*orders)[target] = (*orders)[i]
			target++
		}
	}
	*orders = (*orders)[:target]
}

// FilterOrdersByTimeRange removes any OrderDetails outside of the time range
func FilterOrdersByTimeRange(orders *[]Detail, startTime, endTime time.Time) error {
	if len(*orders) == 0 {
		return nil
	}

	if err := common.StartEndTimeCheck(startTime, endTime); err != nil {
		if errors.Is(err, common.ErrDateUnset) {
			return nil
		}
		return fmt.Errorf("cannot filter orders by time range %w", err)
	}

	target := 0
	for i := range *orders {
		if ((*orders)[i].Date.Unix() >= startTime.Unix() && (*orders)[i].Date.Unix() <= endTime.Unix()) ||
			(*orders)[i].Date.IsZero() {
			(*orders)[target] = (*orders)[i]
			target++
		}
	}
	*orders = (*orders)[:target]
	return nil
}

// FilterOrdersByPairs removes any order details that do not match the
// provided currency pairs list. It is forgiving in that the provided pairs can
// match quote or base pairs
func FilterOrdersByPairs(orders *[]Detail, pairs []currency.Pair) {
	if len(pairs) == 0 ||
		(len(pairs) == 1 && pairs[0].IsEmpty()) ||
		len(*orders) == 0 {
		return
	}

	target := 0
	for x := range *orders {
		for y := range pairs {
			if (*orders)[x].Pair.EqualIncludeReciprocal(pairs[y]) {
				(*orders)[target] = (*orders)[x]
				target++
				break
			}
		}
	}
	*orders = (*orders)[:target]
}

func (b ByPrice) Len() int {
	return len(b)
}

func (b ByPrice) Less(i, j int) bool {
	return b[i].Price < b[j].Price
}

func (b ByPrice) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// SortOrdersByPrice the caller function to sort orders
func SortOrdersByPrice(orders *[]Detail, reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(ByPrice(*orders)))
	} else {
		sort.Sort(ByPrice(*orders))
	}
}

func (b ByOrderType) Len() int {
	return len(b)
}

func (b ByOrderType) Less(i, j int) bool {
	return b[i].Type.String() < b[j].Type.String()
}

func (b ByOrderType) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// SortOrdersByType the caller function to sort orders
func SortOrdersByType(orders *[]Detail, reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(ByOrderType(*orders)))
	} else {
		sort.Sort(ByOrderType(*orders))
	}
}

func (b ByCurrency) Len() int {
	return len(b)
}

func (b ByCurrency) Less(i, j int) bool {
	return b[i].Pair.String() < b[j].Pair.String()
}

func (b ByCurrency) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// SortOrdersByCurrency the caller function to sort orders
func SortOrdersByCurrency(orders *[]Detail, reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(ByCurrency(*orders)))
	} else {
		sort.Sort(ByCurrency(*orders))
	}
}

func (b ByDate) Len() int {
	return len(b)
}

func (b ByDate) Less(i, j int) bool {
	return b[i].Date.Unix() < b[j].Date.Unix()
}

func (b ByDate) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// SortOrdersByDate the caller function to sort orders
func SortOrdersByDate(orders *[]Detail, reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(ByDate(*orders)))
	} else {
		sort.Sort(ByDate(*orders))
	}
}

func (b ByOrderSide) Len() int {
	return len(b)
}

func (b ByOrderSide) Less(i, j int) bool {
	return b[i].Side.String() < b[j].Side.String()
}

func (b ByOrderSide) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// SortOrdersBySide the caller function to sort orders
func SortOrdersBySide(orders *[]Detail, reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(ByOrderSide(*orders)))
	} else {
		sort.Sort(ByOrderSide(*orders))
	}
}

// StringToOrderSide for converting case insensitive order side
// and returning a real Side
func StringToOrderSide(side string) (Side, error) {
	side = strings.ToUpper(side)
	switch side {
	case Buy.String():
		return Buy, nil
	case Sell.String():
		return Sell, nil
	case Bid.String():
		return Bid, nil
	case Ask.String():
		return Ask, nil
	case Long.String():
		return Long, nil
	case Short.String():
		return Short, nil
	case AnySide.String():
		return AnySide, nil
	default:
		return UnknownSide, fmt.Errorf("'%s' %w", side, errUnrecognisedOrderSide)
	}
}

// StringToOrderType for converting case insensitive order type
// and returning a real Type
func StringToOrderType(oType string) (Type, error) {
	oType = strings.ToUpper(oType)
	switch oType {
	case Limit.String(), "EXCHANGE LIMIT":
		return Limit, nil
	case Market.String(), "EXCHANGE MARKET":
		return Market, nil
	case ImmediateOrCancel.String(), "IMMEDIATE OR CANCEL", "IOC", "EXCHANGE IOC":
		return ImmediateOrCancel, nil
	case Stop.String(), "STOP LOSS", "STOP_LOSS", "EXCHANGE STOP":
		return Stop, nil
	case StopLimit.String(), "EXCHANGE STOP LIMIT":
		return StopLimit, nil
	case TrailingStop.String(), "TRAILING STOP", "EXCHANGE TRAILING STOP":
		return TrailingStop, nil
	case FillOrKill.String(), "EXCHANGE FOK":
		return FillOrKill, nil
	case IOS.String():
		return IOS, nil
	case PostOnly.String():
		return PostOnly, nil
	case AnyType.String():
		return AnyType, nil
	case Trigger.String():
		return Trigger, nil
	default:
		return UnknownType, fmt.Errorf("'%v' %w", oType, errUnrecognisedOrderType)
	}
}

// StringToOrderStatus for converting case insensitive order status
// and returning a real Status
func StringToOrderStatus(status string) (Status, error) {
	status = strings.ToUpper(status)
	switch status {
	case AnyStatus.String():
		return AnyStatus, nil
	case New.String(), "PLACED", "ACCEPTED":
		return New, nil
	case Active.String(), "STATUS_ACTIVE":
		return Active, nil
	case PartiallyFilled.String(), "PARTIALLY MATCHED", "PARTIALLY FILLED":
		return PartiallyFilled, nil
	case Filled.String(), "FULLY MATCHED", "FULLY FILLED", "ORDER_FULLY_TRANSACTED":
		return Filled, nil
	case PartiallyCancelled.String(), "PARTIALLY CANCELLED", "ORDER_PARTIALLY_TRANSACTED":
		return PartiallyCancelled, nil
	case Open.String():
		return Open, nil
	case Closed.String():
		return Closed, nil
	case Cancelled.String(), "CANCELED", "ORDER_CANCELLED":
		return Cancelled, nil
	case PendingCancel.String(), "PENDING CANCEL", "PENDING CANCELLATION":
		return PendingCancel, nil
	case Rejected.String(), "FAILED":
		return Rejected, nil
	case Expired.String():
		return Expired, nil
	case Hidden.String():
		return Hidden, nil
	case InsufficientBalance.String():
		return InsufficientBalance, nil
	case MarketUnavailable.String():
		return MarketUnavailable, nil
	case Cancelling.String():
		return Cancelling, nil
	default:
		return UnknownStatus, fmt.Errorf("'%s' %w", status, errUnrecognisedOrderStatus)
	}
}

func (o *ClassificationError) Error() string {
	if o.OrderID != "" {
		return fmt.Sprintf("%s - OrderID: %s classification error: %v",
			o.Exchange,
			o.OrderID,
			o.Err)
	}
	return fmt.Sprintf("%s - classification error: %v",
		o.Exchange,
		o.Err)
}

// StandardCancel defines an option in the validator to make sure an ID is set
// for a standard cancel
func (c *Cancel) StandardCancel() validate.Checker {
	return validate.Check(func() error {
		if c.ID == "" {
			return errors.New("ID not set")
		}
		return nil
	})
}

// PairAssetRequired is a validation check for when a cancel request
// requires an asset type and currency pair to be present
func (c *Cancel) PairAssetRequired() validate.Checker {
	return validate.Check(func() error {
		if c.Pair.IsEmpty() {
			return ErrPairIsEmpty
		}

		if c.AssetType == asset.Empty {
			return ErrAssetNotSet
		}
		return nil
	})
}

// Validate checks internal struct requirements
func (c *Cancel) Validate(opt ...validate.Checker) error {
	if c == nil {
		return ErrCancelOrderIsNil
	}

	var errs common.Errors
	for _, o := range opt {
		err := o.Check()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return errs
	}
	return nil
}

// Validate checks internal struct requirements
func (g *GetOrdersRequest) Validate(opt ...validate.Checker) error {
	if g == nil {
		return ErrGetOrdersRequestIsNil
	}
	if !g.AssetType.IsValid() {
		return fmt.Errorf("%v %w", g.AssetType, asset.ErrNotSupported)
	}
	var errs common.Errors
	for _, o := range opt {
		err := o.Check()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return errs
	}
	return nil
}

// Validate checks internal struct requirements
func (m *Modify) Validate(opt ...validate.Checker) error {
	if m == nil {
		return ErrModifyOrderIsNil
	}

	if m.Pair.IsEmpty() {
		return ErrPairIsEmpty
	}

	if m.AssetType == asset.Empty {
		return ErrAssetNotSet
	}

	var errs common.Errors
	for _, o := range opt {
		err := o.Check()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return errs
	}
	if m.ClientOrderID == "" && m.ID == "" {
		return ErrOrderIDNotSet
	}
	return nil
}
