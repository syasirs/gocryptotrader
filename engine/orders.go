package engine

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/communications/base"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// vars for the fund manager package
var (
	OrderManagerDelay      = time.Second * 10
	ErrOrdersAlreadyExists = errors.New("order already exists")
	ErrOrderFourOhFour     = errors.New("order does not exist")
)

// get returns all orders for all exchanges
// should not be exported as it can have large impact if used improperly
func (o *orderStore) get() map[string][]*order.Detail {
	o.m.Lock()
	defer o.m.Unlock()
	return o.Orders
}

// GetByExchangeAndID returns a specific order
func (o *orderStore) GetByExchangeAndID(exchange, ID string) (*order.Detail, error) {
	o.m.Lock()
	defer o.m.Unlock()
	r, ok := o.Orders[exchange]
	if !ok {
		return nil, ErrOrderFourOhFour
	}

	for x := range r {
		if r[x].ID == ID {
			return r[x], nil
		}
	}
	return nil, ErrOrderFourOhFour
}

// GetByInternalOrderID will search all orders for our internal orderID
// and return the order
func (o *orderStore) GetByInternalOrderID(internalOrderID string) (*order.Detail, error) {
	o.m.Lock()
	defer o.m.Unlock()
	for _, v := range o.Orders {
		for x := range v {
			if v[x].InternalOrderID == internalOrderID {
				return v[x], nil
			}
		}
	}
	return nil, ErrOrderFourOhFour
}

func (o *orderStore) exists(order *order.Detail) bool {
	if order == nil {
		return false
	}
	r, ok := o.Orders[order.Exchange]
	if !ok {
		return false
	}

	for x := range r {
		if r[x].ID == order.ID {
			return true
		}
	}

	return false
}

func (o *orderStore) existsWithLock(order *order.Detail) bool {
	o.m.Lock()
	defer o.m.Unlock()
	return o.exists(order)
}

// Adds an order to the orderStore for tracking the lifecycle
// lock is toggle-able in the event that you have already used
// orderStore's lock. Most cases should be true
func (o *orderStore) Add(order *order.Detail, lock bool) error {
	if lock {
		o.m.Lock()
		defer o.m.Unlock()
	}
	if order == nil {
		return errors.New("Order manager: Order is nil")
	}
	exch := GetExchangeByName(order.Exchange)
	if exch == nil {
		return errors.New("unable to get exchange by name")
	}

	if o.exists(order) {
		return ErrOrdersAlreadyExists
	}

	// Untracked websocket orders will not have internalIDs yet
	if order.InternalOrderID == "" {
		id, err := uuid.NewV4()
		if err != nil {
			log.Warnf(log.OrderMgr,
				"Order manager: Unable to generate UUID. Err: %s\n",
				err)
		}
		order.InternalOrderID = id.String()
	}
	orders := o.Orders[order.Exchange]
	orders = append(orders, order)
	o.Orders[order.Exchange] = orders
	return nil
}

// Started returns the status of the orderManager
func (o *orderManager) Started() bool {
	return atomic.LoadInt32(&o.started) == 1
}

// Start will boot up the orderManager
func (o *orderManager) Start() error {
	if atomic.AddInt32(&o.started, 1) != 1 {
		return errors.New("order manager already started")
	}

	log.Debugln(log.OrderBook, "Order manager starting...")

	o.shutdown = make(chan struct{})
	o.orderStore.Orders = make(map[string][]*order.Detail)
	go o.run()
	return nil
}

// Stop will attempt to shutdown the orderManager
func (o *orderManager) Stop() error {
	if atomic.LoadInt32(&o.started) == 0 {
		return errors.New("order manager not started")
	}

	if atomic.AddInt32(&o.stopped, 1) != 1 {
		return errors.New("order manager is already stopped")
	}
	defer func() {
		atomic.CompareAndSwapInt32(&o.stopped, 1, 0)
		atomic.CompareAndSwapInt32(&o.started, 1, 0)
	}()

	log.Debugln(log.OrderBook, "Order manager shutting down...")
	close(o.shutdown)
	return nil
}

func (o *orderManager) gracefulShutdown() {
	if o.cfg.CancelOrdersOnShutdown {
		log.Debugln(log.OrderMgr, "Order manager: Cancelling any open orders...")

		o.CancelAllOrders(nil)
	}
}

func (o *orderManager) run() {
	log.Debugln(log.OrderBook, "Order manager started.")
	tick := time.NewTicker(OrderManagerDelay)
	Bot.ServicesWG.Add(1)
	defer func() {
		log.Debugln(log.OrderMgr, "Order manager shutdown.")
		tick.Stop()
		Bot.ServicesWG.Done()
	}()

	for {
		select {
		case <-o.shutdown:
			o.gracefulShutdown()
			return
		case <-tick.C:
			o.processOrders()
		}
	}
}

// CancelAllOrders iterates and cancels all orders
// for all exchanges if exchangeNames nil
func (o *orderManager) CancelAllOrders(exchangeNames []string) {
	orders := o.orderStore.get()
	if orders == nil {
		return
	}
orderLoop:
	for k, v := range orders {
		log.Debugf(log.OrderMgr, "Order manager: Cancelling order(s) for exchange %s.\n", k)
		if exchangeNames != nil && len(exchangeNames) > 0 {
			var found bool
			for i := range exchangeNames {
				if strings.EqualFold(k, exchangeNames[i]) {
					found = true
				}
			}
			if !found {
				continue orderLoop
			}
		}

		for y := range v {
			log.Debugf(log.OrderMgr, "order manager: Cancelling order ID %v [%v]",
				v[y].ID, v[y])
			err := o.Cancel(&order.Cancel{
				Exchange:      k,
				ID:            v[y].ID,
				AccountID:     v[y].AccountID,
				ClientID:      v[y].ClientID,
				WalletAddress: v[y].WalletAddress,
				Type:          v[y].Type,
				Side:          v[y].Side,
				Pair:          v[y].Pair,
			})
			if err != nil {
				log.Debugln(log.OrderBook, err)
				Bot.CommsManager.PushEvent(base.Event{
					Type:    "order",
					Message: err.Error(),
				})
				continue
			}

			msg := fmt.Sprintf("Order manager: Exchange %s order ID=%v cancelled.",
				k, v[y].ID)
			log.Debugln(log.OrderBook, msg)
			Bot.CommsManager.PushEvent(base.Event{
				Type:    "order",
				Message: msg,
			})
		}
	}
}

// Cancel will find the order in the orderManager, send a cancel request
// to the exchange and if successful, update the status of the order
func (o *orderManager) Cancel(cancel *order.Cancel) error {
	if cancel == nil {
		return errors.New("order cancel param is nil")
	}

	if cancel.Exchange == "" {
		return errors.New("order exchange name is empty")
	}

	if cancel.ID == "" {
		return errors.New("order id is empty")
	}

	exch := GetExchangeByName(cancel.Exchange)
	if exch == nil {
		return errors.New("unable to get exchange by name")
	}

	if cancel.AssetType.String() != "" && !exch.GetAssetTypes().Contains(cancel.AssetType) {
		return errors.New("order asset type not supported by exchange")
	}

	err := exch.CancelOrder(cancel)
	if err != nil {
		return fmt.Errorf("%v - Failed to cancel order: %v", cancel.Exchange, err)
	}
	var od *order.Detail
	od, err = o.orderStore.GetByExchangeAndID(cancel.Exchange, cancel.ID)
	if err != nil {
		return fmt.Errorf("%v - Failed to retrieve order %v to update cancelled status: %v", cancel.Exchange, cancel.ID, err)
	}

	od.Status = order.Cancelled
	return nil
}

// Submit will take in an order struct, send it to the exchange and
// populate it in the orderManager if successful
func (o *orderManager) Submit(newOrder *order.Submit) (*orderSubmitResponse, error) {
	if newOrder == nil {
		return nil, errors.New("order cannot be nil")
	}

	if newOrder.Exchange == "" {
		return nil, errors.New("order exchange name must be specified")
	}

	if err := newOrder.Validate(); err != nil {
		return nil, err
	}

	if o.cfg.EnforceLimitConfig {
		if !o.cfg.AllowMarketOrders && newOrder.Type == order.Market {
			return nil, errors.New("order market type is not allowed")
		}

		if o.cfg.LimitAmount > 0 && newOrder.Amount > o.cfg.LimitAmount {
			return nil, errors.New("order limit exceeds allowed limit")
		}

		if len(o.cfg.AllowedExchanges) > 0 &&
			!common.StringDataCompareInsensitive(o.cfg.AllowedExchanges, newOrder.Exchange) {
			return nil, errors.New("order exchange not found in allowed list")
		}

		if len(o.cfg.AllowedPairs) > 0 && !o.cfg.AllowedPairs.Contains(newOrder.Pair, true) {
			return nil, errors.New("order pair not found in allowed list")
		}
	}

	exch := GetExchangeByName(newOrder.Exchange)
	if exch == nil {
		return nil, errors.New("unable to get exchange by name")
	}
	// Lock at this point to ensure wrapper + websocket order submissions return
	// before any routines.go order submission competes and this returns an error
	o.orderStore.m.Lock()
	defer o.orderStore.m.Unlock()
	result, err := exch.SubmitOrder(newOrder)
	if err != nil {
		return nil, err
	}

	if !result.IsOrderPlaced {
		return nil, errors.New("order unable to be placed")
	}

	var id uuid.UUID
	id, err = uuid.NewV4()
	if err != nil {
		log.Warnf(log.OrderMgr,
			"Order manager: Unable to generate UUID. Err: %s\n",
			err)
	}
	msg := fmt.Sprintf("Order manager: Exchange %s submitted order ID=%v [Ours: %v] pair=%v price=%v amount=%v side=%v type=%v.",
		newOrder.Exchange,
		result.OrderID,
		id.String(),
		newOrder.Pair,
		newOrder.Price,
		newOrder.Amount,
		newOrder.Side,
		newOrder.Type)

	log.Debugln(log.OrderMgr, msg)
	Bot.CommsManager.PushEvent(base.Event{
		Type:    "order",
		Message: msg,
	})
	status := order.New
	if result.FullyMatched {
		status = order.Filled
	}
	err = o.orderStore.Add(&order.Detail{
		ImmediateOrCancel: newOrder.ImmediateOrCancel,
		HiddenOrder:       newOrder.HiddenOrder,
		FillOrKill:        newOrder.FillOrKill,
		PostOnly:          newOrder.PostOnly,
		Price:             newOrder.Price,
		Amount:            newOrder.Amount,
		LimitPriceUpper:   newOrder.LimitPriceUpper,
		LimitPriceLower:   newOrder.LimitPriceLower,
		TriggerPrice:      newOrder.TriggerPrice,
		TargetAmount:      newOrder.TargetAmount,
		ExecutedAmount:    newOrder.ExecutedAmount,
		RemainingAmount:   newOrder.RemainingAmount,
		Fee:               newOrder.Fee,
		Exchange:          newOrder.Exchange,
		InternalOrderID:   id.String(),
		ID:                result.OrderID,
		AccountID:         newOrder.AccountID,
		ClientID:          newOrder.ClientID,
		WalletAddress:     newOrder.WalletAddress,
		Type:              newOrder.Type,
		Side:              newOrder.Side,
		Status:            status,
		AssetType:         newOrder.AssetType,
		Date:              time.Now(),
		LastUpdated:       time.Now(),
		Pair:              newOrder.Pair,
	}, false)
	if err != nil {
		return nil, fmt.Errorf("Order manager: Unable to add %v order %v to orderStore: %s\n", newOrder.Exchange, result.OrderID, err)
	}

	return &orderSubmitResponse{
		SubmitResponse: order.SubmitResponse{
			OrderID: result.OrderID,
		},
		InternalOrderID: id.String(),
	}, nil
}

func (o *orderManager) processOrders() {
	authExchanges := GetAuthAPISupportedExchanges()
	for x := range authExchanges {
		log.Debugf(log.OrderMgr, "Order manager: Procesing orders for exchange %v.\n", authExchanges[x])
		exch := GetExchangeByName(authExchanges[x])
		req := order.GetOrdersRequest{
			Side: order.AnySide,
			Type: order.AnyType,
		}
		result, err := exch.GetActiveOrders(&req)
		if err != nil {
			log.Warnf(log.OrderMgr, "Order manager: Unable to get active orders: %s\n", err)
			continue
		}

		for x := range result {
			ord := &result[x]
			result := o.orderStore.Add(ord, true)
			if result != ErrOrdersAlreadyExists {
				msg := fmt.Sprintf("Order manager: Exchange %s added order ID=%v pair=%v price=%v amount=%v side=%v type=%v.",
					ord.Exchange, ord.ID, ord.Pair, ord.Price, ord.Amount, ord.Side, ord.Type)
				log.Debugf(log.OrderMgr, "%v\n", msg)
				Bot.CommsManager.PushEvent(base.Event{
					Type:    "order",
					Message: msg,
				})
				continue
			}
		}
	}
}
