package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/communications/base"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// SetupOrderManager will boot up the OrderManager
func SetupOrderManager(exchangeManager iExchangeManager, communicationsManager iCommsManager, wg *sync.WaitGroup, enabledFuturesTracking, verbose bool) (*OrderManager, error) {
	if exchangeManager == nil {
		return nil, errNilExchangeManager
	}
	if communicationsManager == nil {
		return nil, errNilCommunicationsManager
	}
	if wg == nil {
		return nil, errNilWaitGroup
	}

	return &OrderManager{
		shutdown: make(chan struct{}),
		orderStore: store{
			Orders:                    make(map[string][]*order.Detail),
			exchangeManager:           exchangeManager,
			commsManager:              communicationsManager,
			wg:                        wg,
			futuresPositionController: order.SetupPositionController(),
			trackFuturesPositions:     enabledFuturesTracking,
		},
		verbose: verbose,
	}, nil
}

// IsRunning safely checks whether the subsystem is running
func (m *OrderManager) IsRunning() bool {
	return m != nil && atomic.LoadInt32(&m.started) == 1
}

// Start runs the subsystem
func (m *OrderManager) Start() error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if !atomic.CompareAndSwapInt32(&m.started, 0, 1) {
		return fmt.Errorf("order manager %w", ErrSubSystemAlreadyStarted)
	}
	log.Debugln(log.OrderMgr, "Order manager starting...")
	m.shutdown = make(chan struct{})
	m.orderStore.wg.Add(1)
	go m.run()
	return nil
}

// Stop attempts to shutdown the subsystem
func (m *OrderManager) Stop() error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	log.Debugln(log.OrderMgr, "Order manager shutting down...")
	close(m.shutdown)
	atomic.CompareAndSwapInt32(&m.started, 1, 0)
	return nil
}

// gracefulShutdown cancels all orders (if enabled) before shutting down
func (m *OrderManager) gracefulShutdown() {
	if !m.cfg.CancelOrdersOnShutdown {
		return
	}
	log.Debugln(log.OrderMgr, "Order manager: Cancelling any open orders...")
	exchanges, err := m.orderStore.exchangeManager.GetExchanges()
	if err != nil {
		log.Errorf(log.OrderMgr, "Order manager cannot get exchanges: %v", err)
		return
	}
	m.CancelAllOrders(context.TODO(), exchanges)
}

// run will periodically process orders
func (m *OrderManager) run() {
	log.Debugln(log.OrderMgr, "Order manager started.")
	timer := time.NewTimer(orderManagerDelay)
	for {
		select {
		case <-m.shutdown:
			m.gracefulShutdown()
			if !timer.Stop() {
				<-timer.C
			}
			m.orderStore.wg.Done()
			log.Debugln(log.OrderMgr, "Order manager shutdown.")
			return
		case <-timer.C:
			// Process orders go routine allows shutdown procedures to continue
			go m.processOrders()
			timer.Reset(orderManagerDelay)
		}
	}
}

// CancelAllOrders iterates and cancels all orders for each exchange provided
func (m *OrderManager) CancelAllOrders(ctx context.Context, exchanges []exchange.IBotExchange) {
	if m == nil || atomic.LoadInt32(&m.started) == 0 {
		return
	}

	allOrders := m.orderStore.get()
	if len(allOrders) == 0 {
		return
	}

	for i := range exchanges {
		orders, ok := allOrders[strings.ToLower(exchanges[i].GetName())]
		if !ok {
			continue
		}
		for j := range orders {
			log.Debugf(log.OrderMgr, "Order manager: Cancelling order(s) for exchange %s.", exchanges[i].GetName())
			cancel, err := orders[j].DeriveCancel()
			if err != nil {
				log.Error(log.OrderMgr, err)
				continue
			}
			err = m.Cancel(ctx, cancel)
			if err != nil {
				log.Error(log.OrderMgr, err)
			}
		}
	}
}

// Cancel will find the order in the OrderManager, send a cancel request
// to the exchange and if successful, update the status of the order
func (m *OrderManager) Cancel(ctx context.Context, cancel *order.Cancel) error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	var err error
	defer func() {
		if err != nil {
			m.orderStore.commsManager.PushEvent(base.Event{
				Type:    "order",
				Message: err.Error(),
			})
		}
	}()

	if cancel == nil {
		err = errors.New("order cancel param is nil")
		return err
	}
	if cancel.Exchange == "" {
		err = errors.New("order exchange name is empty")
		return err
	}
	if cancel.OrderID == "" {
		err = errors.New("order id is empty")
		return err
	}

	exch, err := m.orderStore.exchangeManager.GetExchangeByName(cancel.Exchange)
	if err != nil {
		return err
	}

	if cancel.AssetType.String() != "" && !exch.GetAssetTypes(false).Contains(cancel.AssetType) {
		err = errors.New("order asset type not supported by exchange")
		return err
	}

	log.Debugf(log.OrderMgr, "Order manager: Cancelling order ID %v [%+v]",
		cancel.OrderID, cancel)

	err = exch.CancelOrder(ctx, cancel)
	if err != nil {
		err = fmt.Errorf("%v - Failed to cancel order: %w", cancel.Exchange, err)
		return err
	}
	od, err := m.orderStore.getByExchangeAndID(cancel.Exchange, cancel.OrderID)
	if err != nil {
		err = fmt.Errorf("%v - Failed to retrieve order %v to update cancelled status: %w",
			cancel.Exchange, cancel.OrderID, err)
		return err
	}
	od.Status = order.Cancelled
	err = m.orderStore.updateExisting(od)
	if err != nil {
		err = fmt.Errorf("%v - Failed to update existing order when cancelled: %w", cancel.Exchange, err)
		return err
	}

	msg := fmt.Sprintf("Order manager: Exchange %s order ID=%v cancelled.",
		od.Exchange, od.OrderID)
	log.Debugln(log.OrderMgr, msg)
	m.orderStore.commsManager.PushEvent(base.Event{Type: "order", Message: msg})
	return nil
}

// GetFuturesPositionsForExchange returns futures positions stored within
// the order manager's futures position tracker that match the provided params
func (m *OrderManager) GetFuturesPositionsForExchange(exch string, item asset.Item, pair currency.Pair) ([]order.PositionStats, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	if m.orderStore.futuresPositionController == nil {
		return nil, errFuturesTrackerNotSetup
	}
	if !item.IsFutures() {
		return nil, fmt.Errorf("%v %w", item, order.ErrNotFuturesAsset)
	}

	return m.orderStore.futuresPositionController.GetPositionsForExchange(exch, item, pair)
}

// ClearFuturesTracking will clear existing futures positions for a given exchange,
// asset, pair for the event that positions have not been tracked accurately
func (m *OrderManager) ClearFuturesTracking(exch string, item asset.Item, pair currency.Pair) error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	if m.orderStore.futuresPositionController == nil {
		return errFuturesTrackerNotSetup
	}
	if !item.IsFutures() {
		return fmt.Errorf("%v %w", item, order.ErrNotFuturesAsset)
	}

	return m.orderStore.futuresPositionController.ClearPositionsForExchange(exch, item, pair)
}

// UpdateOpenPositionUnrealisedPNL finds an open position from
// an exchange asset pair, then calculates the unrealisedPNL
// using the latest ticker data
func (m *OrderManager) UpdateOpenPositionUnrealisedPNL(e string, item asset.Item, pair currency.Pair, last float64, updated time.Time) (decimal.Decimal, error) {
	if m == nil {
		return decimal.Zero, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return decimal.Zero, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	if m.orderStore.futuresPositionController == nil {
		return decimal.Zero, errFuturesTrackerNotSetup
	}
	if !item.IsFutures() {
		return decimal.Zero, fmt.Errorf("%v %w", item, order.ErrNotFuturesAsset)
	}

	return m.orderStore.futuresPositionController.UpdateOpenPositionUnrealisedPNL(e, item, pair, last, updated)
}

// GetOrderInfo calls the exchange's wrapper GetOrderInfo function
// and stores the result in the order manager
func (m *OrderManager) GetOrderInfo(ctx context.Context, exchangeName, orderID string, cp currency.Pair, a asset.Item) (order.Detail, error) {
	if m == nil {
		return order.Detail{}, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return order.Detail{}, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}

	if orderID == "" {
		return order.Detail{}, ErrOrderIDCannotBeEmpty
	}

	exch, err := m.orderStore.exchangeManager.GetExchangeByName(exchangeName)
	if err != nil {
		return order.Detail{}, err
	}
	result, err := exch.GetOrderInfo(ctx, orderID, cp, a)
	if err != nil {
		return order.Detail{}, err
	}

	upsertResponse, err := m.orderStore.upsert(&result)
	if err != nil {
		return order.Detail{}, err
	}

	return upsertResponse.OrderDetails, nil
}

// validate ensures a submitted order is valid before adding to the manager
func (m *OrderManager) validate(newOrder *order.Submit) error {
	if newOrder == nil {
		return errNilOrder
	}

	if newOrder.Exchange == "" {
		return ErrExchangeNameIsEmpty
	}

	if err := newOrder.Validate(); err != nil {
		return fmt.Errorf("order manager: %w", err)
	}

	if m.cfg.EnforceLimitConfig {
		if !m.cfg.AllowMarketOrders && newOrder.Type == order.Market {
			return errors.New("order market type is not allowed")
		}

		if m.cfg.LimitAmount > 0 && newOrder.Amount > m.cfg.LimitAmount {
			return errors.New("order limit exceeds allowed limit")
		}

		if len(m.cfg.AllowedExchanges) > 0 &&
			!common.StringDataCompareInsensitive(m.cfg.AllowedExchanges, newOrder.Exchange) {
			return errors.New("order exchange not found in allowed list")
		}

		if len(m.cfg.AllowedPairs) > 0 && !m.cfg.AllowedPairs.Contains(newOrder.Pair, true) {
			return errors.New("order pair not found in allowed list")
		}
	}
	return nil
}

// Modify depends on the order.Modify.ID and order.Modify.Exchange fields to uniquely
// identify an order to modify.
func (m *OrderManager) Modify(ctx context.Context, mod *order.Modify) (*order.ModifyResponse, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}

	// Fetch details from locally managed order store.
	det, err := m.orderStore.getByExchangeAndID(mod.Exchange, mod.OrderID)
	if det == nil || err != nil {
		return nil, fmt.Errorf("order does not exist: %w", err)
	}

	// Populate additional Modify fields as some of them are required by various
	// exchange implementations.
	mod.Pair = det.Pair                           // Used by Bithumb.
	mod.Side = det.Side                           // Used by Bithumb.
	mod.PostOnly = det.PostOnly                   // Used by Poloniex.
	mod.ImmediateOrCancel = det.ImmediateOrCancel // Used by Poloniex.

	// Following is just a precaution to not modify orders by mistake if exchange
	// implementations do not check fields of the Modify struct for zero values.
	if mod.Amount == 0 {
		mod.Amount = det.Amount
	}
	if mod.Price == 0 {
		mod.Price = det.Price
	}

	// Get exchange instance and submit order modification request.
	exch, err := m.orderStore.exchangeManager.GetExchangeByName(mod.Exchange)
	if err != nil {
		return nil, err
	}
	res, err := exch.ModifyOrder(ctx, mod)
	if err != nil {
		message := fmt.Sprintf(
			"Order manager: Exchange %s order ID=%v: failed to modify",
			mod.Exchange,
			mod.OrderID,
		)
		m.orderStore.commsManager.PushEvent(base.Event{
			Type:    "order",
			Message: message,
		})
		return nil, err
	}

	// If modification is successful, apply changes to local order store.
	//
	// XXX: This comes with a race condition, because [request -> changes] are not
	// atomic.
	err = m.orderStore.modifyExisting(mod.OrderID, res)

	// Notify observers.
	var message string
	if err != nil {
		message = "Order manager: Exchange %s order ID=%v: modified on exchange, but failed to modify locally"
	} else {
		message = "Order manager: Exchange %s order ID=%v: modified successfully"
	}
	m.orderStore.commsManager.PushEvent(base.Event{
		Type:    "order",
		Message: fmt.Sprintf(message, mod.Exchange, res.OrderID),
	})
	return &order.ModifyResponse{OrderID: res.OrderID}, err
}

// Submit will take in an order struct, send it to the exchange and
// populate it in the OrderManager if successful
func (m *OrderManager) Submit(ctx context.Context, newOrder *order.Submit) (*OrderSubmitResponse, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}

	err := m.validate(newOrder)
	if err != nil {
		return nil, err
	}
	exch, err := m.orderStore.exchangeManager.GetExchangeByName(newOrder.Exchange)
	if err != nil {
		return nil, err
	}

	// Checks for exchange min max limits for order amounts before order
	// execution can occur
	err = exch.CheckOrderExecutionLimits(newOrder.AssetType,
		newOrder.Pair,
		newOrder.Price,
		newOrder.Amount,
		newOrder.Type)
	if err != nil {
		return nil, fmt.Errorf("order manager: exchange %s unable to place order: %w",
			newOrder.Exchange,
			err)
	}

	// Determines if current trading activity is turned off by the exchange for
	// the currency pair
	err = exch.CanTradePair(newOrder.Pair, newOrder.AssetType)
	if err != nil {
		return nil, fmt.Errorf("order manager: exchange %s cannot trade pair %s %s: %w",
			newOrder.Exchange,
			newOrder.Pair,
			newOrder.AssetType,
			err)
	}

	result, err := exch.SubmitOrder(ctx, newOrder)
	if err != nil {
		return nil, err
	}

	return m.processSubmittedOrder(result)
}

// SubmitFakeOrder runs through the same process as order submission
// but does not touch live endpoints
func (m *OrderManager) SubmitFakeOrder(newOrder *order.Submit, resultingOrder *order.SubmitResponse, checkExchangeLimits bool) (*OrderSubmitResponse, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}

	err := m.validate(newOrder)
	if err != nil {
		return nil, err
	}
	exch, err := m.orderStore.exchangeManager.GetExchangeByName(newOrder.Exchange)
	if err != nil {
		return nil, err
	}

	if checkExchangeLimits {
		// Checks for exchange min max limits for order amounts before order
		// execution can occur
		err = exch.CheckOrderExecutionLimits(newOrder.AssetType,
			newOrder.Pair,
			newOrder.Price,
			newOrder.Amount,
			newOrder.Type)
		if err != nil {
			return nil, fmt.Errorf("order manager: exchange %s unable to place order: %w",
				newOrder.Exchange,
				err)
		}
	}
	return m.processSubmittedOrder(resultingOrder)
}

// GetOrdersSnapshot returns a snapshot of all orders in the orderstore. It optionally filters any orders that do not match the status
// but a status of "" or ANY will include all
// the time adds contexts for when the snapshot is relevant for
func (m *OrderManager) GetOrdersSnapshot(s order.Status) []order.Detail {
	if m == nil || atomic.LoadInt32(&m.started) == 0 {
		return nil
	}
	var os []order.Detail
	for _, v := range m.orderStore.Orders {
		for i := range v {
			if s != v[i].Status && s != order.AnyStatus && s != order.UnknownStatus {
				continue
			}
			os = append(os, *v[i])
		}
	}

	return os
}

// GetOrdersFiltered returns a snapshot of all orders in the order store.
// Filtering is applied based on the order.Filter unless entries are empty
func (m *OrderManager) GetOrdersFiltered(f *order.Filter) ([]order.Detail, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	return m.orderStore.getFilteredOrders(f)
}

// GetOrdersActive returns a snapshot of all orders in the order store
// that have a status that indicates it's currently tradable
func (m *OrderManager) GetOrdersActive(f *order.Filter) ([]order.Detail, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	return m.orderStore.getActiveOrders(f), nil
}

// processSubmittedOrder adds a new order to the manager
func (m *OrderManager) processSubmittedOrder(newOrderResp *order.SubmitResponse) (*OrderSubmitResponse, error) {
	if newOrderResp == nil {
		return nil, order.ErrOrderDetailIsNil
	}

	id, err := uuid.NewV4()
	if err != nil {
		log.Warnf(log.OrderMgr, "Order manager: Unable to generate UUID. Err: %s", err)
	}

	detail, err := newOrderResp.DeriveDetail(id)
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Order manager: Exchange %s submitted order ID=%v [Ours: %v] pair=%v price=%v amount=%v quoteAmount=%v side=%v type=%v for time %v.",
		detail.Exchange,
		detail.OrderID,
		detail.InternalOrderID.String(),
		detail.Pair,
		detail.Price,
		detail.Amount,
		detail.QuoteAmount,
		detail.Side,
		detail.Type,
		detail.Date)

	log.Debugln(log.OrderMgr, msg)
	if m.orderStore.commsManager != nil {
		m.orderStore.commsManager.PushEvent(base.Event{Type: "order", Message: msg})
	}

	err = m.orderStore.add(detail.CopyToPointer())
	if err != nil {
		return nil, fmt.Errorf("unable to add %v order %v to orderStore: %s",
			detail.Exchange, detail.OrderID, err)
	}

	return &OrderSubmitResponse{Detail: detail, InternalOrderID: id.String()}, nil
}

// processOrders iterates over all exchange orders via API
// and adds them to the internal order store
func (m *OrderManager) processOrders() {
	if !atomic.CompareAndSwapInt32(&m.processingOrders, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&m.processingOrders, 0)

	exchanges, err := m.orderStore.exchangeManager.GetExchanges()
	if err != nil {
		log.Errorf(log.OrderMgr, "Order manager cannot get exchanges: %v", err)
		return
	}
	var wg sync.WaitGroup
	for i := range exchanges {
		if !exchanges[i].IsRESTAuthenticationSupported() {
			continue
		}
		log.Debugf(log.OrderMgr,
			"Order manager: Processing orders for exchange %v.",
			exchanges[i].GetName())

		enabledAssets := exchanges[i].GetAssetTypes(true)
		for y := range enabledAssets {
			pairs, err := exchanges[i].GetEnabledPairs(enabledAssets[y])
			if err != nil {
				log.Errorf(log.OrderMgr,
					"Order manager: Unable to get enabled pairs for %s and asset type %s: %s",
					exchanges[i].GetName(),
					enabledAssets[y],
					err)
				continue
			}

			if len(pairs) == 0 {
				if m.verbose {
					log.Debugf(log.OrderMgr,
						"Order manager: No pairs enabled for %s and asset type %s, skipping...",
						exchanges[i].GetName(),
						enabledAssets[y])
				}
				continue
			}

			filter := &order.Filter{Exchange: exchanges[i].GetName()}
			orders := m.orderStore.getActiveOrders(filter)
			order.FilterOrdersByPairs(&orders, pairs)

			result, err := exchanges[i].GetActiveOrders(context.TODO(), &order.GetOrdersRequest{
				Side:      order.AnySide,
				Type:      order.AnyType,
				Pairs:     pairs,
				AssetType: enabledAssets[y],
			})
			if err != nil {
				log.Errorf(log.OrderMgr,
					"Order manager: Unable to get active orders for %s and asset type %s: %s",
					exchanges[i].GetName(),
					enabledAssets[y],
					err)
				continue
			}
			if len(orders) == 0 && len(result) == 0 {
				continue
			}

			for z := range result {
				upsertResponse, err := m.UpsertOrder(&result[z])
				if err != nil {
					log.Error(log.OrderMgr, err)
				} else {
					for i := range orders {
						if orders[i].InternalOrderID != upsertResponse.OrderDetails.InternalOrderID {
							continue
						}
						orders[i] = orders[len(orders)-1]
						orders = orders[:len(orders)-1]
					}
				}
			}
			if !exchanges[i].GetBase().GetSupportedFeatures().RESTCapabilities.GetOrder {
				continue
			}
			wg.Add(1)
			go m.processMatchingOrders(exchanges[i], orders, &wg)
		}
	}
	wg.Wait()
}

func (m *OrderManager) processMatchingOrders(exch exchange.IBotExchange, orders []order.Detail, wg *sync.WaitGroup) {
	for x := range orders {
		if time.Since(orders[x].LastUpdated) < time.Minute {
			continue
		}
		err := m.FetchAndUpdateExchangeOrder(exch, &orders[x], orders[x].AssetType)
		if err != nil {
			log.Error(log.OrderMgr, err)
		}
	}
	if wg != nil {
		wg.Done()
	}
}

// FetchAndUpdateExchangeOrder calls the exchange to upsert an order to the order store
func (m *OrderManager) FetchAndUpdateExchangeOrder(exch exchange.IBotExchange, ord *order.Detail, assetType asset.Item) error {
	if ord == nil {
		return errors.New("order manager: Order is nil")
	}
	fetchedOrder, err := exch.GetOrderInfo(context.TODO(), ord.OrderID, ord.Pair, assetType)
	if err != nil {
		ord.Status = order.UnknownStatus
		return err
	}
	fetchedOrder.LastUpdated = time.Now()
	_, err = m.UpsertOrder(&fetchedOrder)
	return err
}

// Exists checks whether an order exists in the order store
func (m *OrderManager) Exists(o *order.Detail) bool {
	return m != nil && atomic.LoadInt32(&m.started) != 0 && m.orderStore.exists(o)
}

// Add adds an order to the orderstore
func (m *OrderManager) Add(o *order.Detail) error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}

	return m.orderStore.add(o)
}

// GetByExchangeAndID returns a copy of an order from an exchange if it matches the ID
func (m *OrderManager) GetByExchangeAndID(exchangeName, id string) (*order.Detail, error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	return m.orderStore.getByExchangeAndID(exchangeName, id)
}

// UpdateExistingOrder will update an existing order in the orderstore
func (m *OrderManager) UpdateExistingOrder(od *order.Detail) error {
	if m == nil {
		return fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	return m.orderStore.updateExisting(od)
}

// UpsertOrder updates an existing order or adds a new one to the orderstore
func (m *OrderManager) UpsertOrder(od *order.Detail) (resp *OrderUpsertResponse, err error) {
	if m == nil {
		return nil, fmt.Errorf("order manager %w", ErrNilSubsystem)
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return nil, fmt.Errorf("order manager %w", ErrSubSystemNotStarted)
	}
	if od == nil {
		return nil, errNilOrder
	}
	var msg string
	defer func(message *string) {
		if message == nil {
			log.Errorf(log.OrderMgr, "UpsertOrder: produced nil order event message\n")
			return
		}
		m.orderStore.commsManager.PushEvent(base.Event{
			Type:    "order",
			Message: *message,
		})
	}(&msg)

	upsertResponse, err := m.orderStore.upsert(od)
	if err != nil {
		msg = fmt.Sprintf(
			"Order manager: Exchange %s unable to upsert order ID=%v internal ID=%v pair=%v price=%.8f amount=%.8f side=%v type=%v status=%v: %s",
			od.Exchange, od.OrderID, od.InternalOrderID, od.Pair, od.Price, od.Amount, od.Side, od.Type, od.Status, err)
		return nil, err
	}

	status := "updated"
	if upsertResponse.IsNewOrder {
		status = "added"
	}
	msg = fmt.Sprintf("Order manager: Exchange %s %s order ID=%v internal ID=%v pair=%v price=%.8f amount=%.8f side=%v type=%v status=%v.",
		upsertResponse.OrderDetails.Exchange, status, upsertResponse.OrderDetails.OrderID, upsertResponse.OrderDetails.InternalOrderID,
		upsertResponse.OrderDetails.Pair, upsertResponse.OrderDetails.Price, upsertResponse.OrderDetails.Amount,
		upsertResponse.OrderDetails.Side, upsertResponse.OrderDetails.Type, upsertResponse.OrderDetails.Status)
	if upsertResponse.IsNewOrder {
		log.Info(log.OrderMgr, msg)
		return upsertResponse, nil
	}
	log.Debug(log.OrderMgr, msg)
	return upsertResponse, nil
}

// get returns a copy of all orders for all exchanges.
func (s *store) get() map[string][]*order.Detail {
	orders := make(map[string][]*order.Detail)
	s.m.Lock()
	for k, val := range s.Orders {
		orders[k] = order.CopyPointerOrderSlice(val)
	}
	s.m.Unlock()
	return orders
}

// getByExchangeAndID returns a specific order by exchange and id
func (s *store) getByExchangeAndID(exchange, id string) (*order.Detail, error) {
	s.m.Lock()
	defer s.m.Unlock()
	r, ok := s.Orders[strings.ToLower(exchange)]
	if !ok {
		return nil, ErrExchangeNotFound
	}

	for x := range r {
		if r[x].OrderID == id {
			return r[x].CopyToPointer(), nil
		}
	}
	return nil, ErrOrderNotFound
}

// updateExisting checks if an order exists in the orderstore
// and then updates it
func (s *store) updateExisting(od *order.Detail) error {
	if od == nil {
		return errNilOrder
	}
	s.m.Lock()
	defer s.m.Unlock()
	r, ok := s.Orders[strings.ToLower(od.Exchange)]
	if !ok {
		return ErrExchangeNotFound
	}
	for x := range r {
		if r[x].OrderID != od.OrderID {
			continue
		}
		err := r[x].UpdateOrderFromDetail(od)
		if err != nil {
			return err
		}
		if !r[x].AssetType.IsFutures() {
			return nil
		}
		err = s.futuresPositionController.TrackNewOrder(r[x])
		if err != nil && !errors.Is(err, order.ErrPositionClosed) {
			return err
		}
		return nil
	}
	return ErrOrderNotFound
}

// modifyExisting depends on mod.Exchange and given ID to uniquely identify an order and
// modify it.
func (s *store) modifyExisting(id string, mod *order.ModifyResponse) error {
	s.m.Lock()
	defer s.m.Unlock()
	r, ok := s.Orders[strings.ToLower(mod.Exchange)]
	if !ok {
		return ErrExchangeNotFound
	}
	for x := range r {
		if r[x].OrderID != id {
			continue
		}
		r[x].UpdateOrderFromModifyResponse(mod)
		if !r[x].AssetType.IsFutures() {
			return nil
		}
		err := s.futuresPositionController.TrackNewOrder(r[x])
		if err != nil && !errors.Is(err, order.ErrPositionClosed) {
			return err
		}
		return nil
	}
	return ErrOrderNotFound
}

// upsert (1) checks if such an exchange exists in the exchangeManager, (2) checks if
// order exists and updates/creates it.
func (s *store) upsert(od *order.Detail) (*OrderUpsertResponse, error) {
	if od == nil {
		return nil, errNilOrder
	}
	lName := strings.ToLower(od.Exchange)
	_, err := s.exchangeManager.GetExchangeByName(lName)
	if err != nil {
		return nil, err
	}
	s.m.Lock()
	defer s.m.Unlock()
	if s.trackFuturesPositions && od.AssetType.IsFutures() {
		err = s.futuresPositionController.TrackNewOrder(od)
		if err != nil && !errors.Is(err, order.ErrPositionClosed) {
			return nil, err
		}
	}
	// TODO: Return pointer to slice because new orders we are accessing map
	// twice for lookup.
	exchangeOrders := s.Orders[lName]
	for x := range exchangeOrders {
		if exchangeOrders[x].OrderID != od.OrderID {
			continue
		}
		err := exchangeOrders[x].UpdateOrderFromDetail(od)
		if err != nil {
			return nil, err
		}
		return &OrderUpsertResponse{
			OrderDetails: exchangeOrders[x].Copy(),
			IsNewOrder:   false,
		}, nil
	}
	// Untracked websocket orders will not have internalIDs yet
	od.GenerateInternalOrderID()
	s.Orders[lName] = append(s.Orders[lName], od)
	return &OrderUpsertResponse{OrderDetails: od.Copy(), IsNewOrder: true}, nil
}

// exists verifies if the orderstore contains the provided order
func (s *store) exists(det *order.Detail) bool {
	if det == nil {
		return false
	}
	s.m.RLock()
	defer s.m.RUnlock()
	exchangeOrders := s.Orders[strings.ToLower(det.Exchange)]
	for x := range exchangeOrders {
		if exchangeOrders[x].OrderID == det.OrderID {
			return true
		}
	}
	return false
}

// Add Adds an order to the orderStore for tracking the lifecycle
func (s *store) add(det *order.Detail) error {
	if det == nil {
		return errNilOrder
	}
	name := strings.ToLower(det.Exchange)
	_, err := s.exchangeManager.GetExchangeByName(name)
	if err != nil {
		return err
	}
	if s.exists(det) { // TODO: Error on conflict; remove unnecessary locking.
		return ErrOrdersAlreadyExists
	}

	// Untracked websocket orders will not have internalIDs yet
	det.GenerateInternalOrderID()
	s.m.Lock()
	defer s.m.Unlock()
	s.Orders[name] = append(s.Orders[name], det)
	if !det.AssetType.IsFutures() {
		return nil
	}
	return s.futuresPositionController.TrackNewOrder(det)
}

// getFilteredOrders returns a filtered copy of the orders
func (s *store) getFilteredOrders(f *order.Filter) ([]order.Detail, error) {
	if f == nil {
		return nil, errors.New("filter is nil")
	}
	s.m.RLock()
	defer s.m.RUnlock()

	var os []order.Detail
	// optimization if Exchange is filtered
	if f.Exchange != "" {
		if e, ok := s.Orders[strings.ToLower(f.Exchange)]; ok {
			for i := range e {
				if !e[i].MatchFilter(f) {
					continue
				}
				os = append(os, e[i].Copy())
			}
		}
	} else {
		for _, e := range s.Orders {
			for i := range e {
				if !e[i].MatchFilter(f) {
					continue
				}
				os = append(os, e[i].Copy())
			}
		}
	}
	return os, nil
}

// getActiveOrders returns copy of the orders that are active
func (s *store) getActiveOrders(f *order.Filter) []order.Detail {
	s.m.RLock()
	defer s.m.RUnlock()

	var orders []order.Detail
	switch {
	case f == nil:
		for _, e := range s.Orders {
			for i := range e {
				if e[i].Status != order.UnknownStatus && !e[i].IsActive() {
					continue
				}
				orders = append(orders, e[i].Copy())
			}
		}
	case f.Exchange != "":
		// optimization if Exchange is filtered
		if e, ok := s.Orders[strings.ToLower(f.Exchange)]; ok {
			for i := range e {
				if e[i].Status != order.UnknownStatus && (!e[i].IsActive() || !e[i].MatchFilter(f)) {
					continue
				}
				orders = append(orders, e[i].Copy())
			}
		}
	default:
		for _, e := range s.Orders {
			for i := range e {
				if e[i].Status != order.UnknownStatus && (!e[i].IsActive() || !e[i].MatchFilter(f)) {
					continue
				}
				orders = append(orders, e[i].Copy())
			}
		}
	}

	return orders
}
