package engine

import (
	"fmt"
	"sync/atomic"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine/subsystem"
	"github.com/thrasher-corp/gocryptotrader/engine/subsystem/synchronize"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fill"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// setupWebsocketRoutineManager creates a new websocket routine manager
func setupWebsocketRoutineManager(exchangeManager subsystem.ExchangeManager, orderManager subsystem.OrderManager, syncer subsystem.CurrencyPairSyncer, cfg *currency.Config, verbose bool) (*websocketRoutineManager, error) {
	if exchangeManager == nil {
		return nil, subsystem.ErrNilExchangeManager
	}
	if orderManager == nil {
		return nil, errNilOrderManager
	}
	if syncer == nil {
		return nil, errNilCurrencyPairSyncer
	}
	if cfg == nil {
		return nil, errNilCurrencyConfig
	}
	if cfg.CurrencyPairFormat == nil {
		return nil, errNilCurrencyPairFormat
	}
	man := &websocketRoutineManager{
		verbose:         verbose,
		exchangeManager: exchangeManager,
		orderManager:    orderManager,
		syncer:          syncer,
		currencyConfig:  cfg,
		shutdown:        make(chan struct{}),
	}
	return man, man.registerWebsocketDataHandler(man.websocketDataHandler, false)
}

// Start runs the subsystem
func (m *websocketRoutineManager) Start() error {
	if m == nil {
		return fmt.Errorf("websocket routine manager %w", subsystem.ErrNil)
	}

	if m.currencyConfig == nil {
		return errNilCurrencyConfig
	}

	if m.currencyConfig.CurrencyPairFormat == nil {
		return errNilCurrencyPairFormat
	}

	if !atomic.CompareAndSwapInt32(&m.started, 0, 1) {
		return subsystem.ErrAlreadyStarted
	}
	m.shutdown = make(chan struct{})
	m.websocketRoutine()
	return nil
}

// IsRunning safely checks whether the subsystem is running
func (m *websocketRoutineManager) IsRunning() bool {
	if m == nil {
		return false
	}
	return atomic.LoadInt32(&m.started) == 1
}

// Stop attempts to shutdown the subsystem
func (m *websocketRoutineManager) Stop() error {
	if m == nil {
		return fmt.Errorf("websocket routine manager %w", subsystem.ErrNil)
	}
	if !atomic.CompareAndSwapInt32(&m.started, 1, 0) {
		return fmt.Errorf("websocket routine manager %w", subsystem.ErrNotStarted)
	}
	close(m.shutdown)
	m.wg.Wait()
	return nil
}

// websocketRoutine Initial routine management system for websocket
func (m *websocketRoutineManager) websocketRoutine() {
	if m.verbose {
		log.Debugln(log.WebsocketMgr, "Connecting exchange websocket services...")
	}
	exchanges, err := m.exchangeManager.GetExchanges()
	if err != nil {
		log.Errorf(log.WebsocketMgr, "websocket routine manager cannot get exchanges: %v", err)
	}
	for i := range exchanges {
		go func(i int) {
			if exchanges[i].SupportsWebsocket() {
				if m.verbose {
					log.Debugf(log.WebsocketMgr,
						"Exchange %s websocket support: Yes Enabled: %v",
						exchanges[i].GetName(),
						common.IsEnabled(exchanges[i].IsWebsocketEnabled()),
					)
				}

				ws, err := exchanges[i].GetWebsocket()
				if err != nil {
					log.Errorf(
						log.WebsocketMgr,
						"Exchange %s GetWebsocket error: %s",
						exchanges[i].GetName(),
						err,
					)
					return
				}

				if ws.IsEnabled() {
					err = ws.Connect()
					if err != nil {
						log.Errorf(log.WebsocketMgr, "%v", err)
					}

					err = m.websocketDataReceiver(ws)
					if err != nil {
						log.Errorf(log.WebsocketMgr, "%v", err)
					}

					err = ws.FlushChannels()
					if err != nil {
						log.Errorf(log.WebsocketMgr, "Failed to subscribe: %v", err)
					}
				}
			} else if m.verbose {
				log.Debugf(log.WebsocketMgr,
					"Exchange %s websocket support: No",
					exchanges[i].GetName(),
				)
			}
		}(i)
	}
}

// WebsocketDataReceiver handles websocket data coming from a websocket feed
// associated with an exchange
func (m *websocketRoutineManager) websocketDataReceiver(ws *stream.Websocket) error {
	if m == nil {
		return fmt.Errorf("websocket routine manager %w", subsystem.ErrNil)
	}

	if ws == nil {
		return errNilWebsocket
	}

	if atomic.LoadInt32(&m.started) == 0 {
		return errRoutineManagerNotStarted
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.shutdown:
				return
			case data := <-ws.ToRoutine:
				if data == nil {
					log.Errorf(log.WebsocketMgr, "exchange %s nil data sent to websocket", ws.GetName())
				}
				m.mu.RLock()
				for x := range m.dataHandlers {
					err := m.dataHandlers[x](ws.GetName(), data)
					if err != nil {
						log.Error(log.WebsocketMgr, err)
					}
				}
				m.mu.RUnlock()
			}
		}
	}()
	return nil
}

// websocketDataHandler is the default central point for exchange websocket
// implementations to send processed data which will then pass that to an
// appropriate handler.
func (m *websocketRoutineManager) websocketDataHandler(exchName string, data interface{}) error {
	switch d := data.(type) {
	case string:
		log.Info(log.WebsocketMgr, d)
	case error:
		return fmt.Errorf("exchange %s websocket error - %s", exchName, data)
	case stream.FundingData:
		if m.verbose {
			log.Infof(log.WebsocketMgr, "%s websocket %s %s funding updated %+v",
				exchName,
				m.FormatCurrency(d.CurrencyPair),
				d.AssetType,
				d)
		}
	case *ticker.Price:
		if m.syncer.IsRunning() {
			err := m.syncer.Update(exchName,
				synchronize.WebsocketUpdate,
				d.Pair,
				d.AssetType,
				int(synchronize.Ticker),
				nil)
			if err != nil {
				return err
			}
		}
		err := ticker.ProcessTicker(d)
		if err != nil {
			return err
		}
		m.syncer.PrintTickerSummary(d, synchronize.WebsocketUpdate, err)
	case stream.KlineData:
		if m.verbose {
			log.Infof(log.WebsocketMgr, "%s websocket %s %s kline updated %+v",
				exchName,
				m.FormatCurrency(d.Pair),
				d.AssetType,
				d)
		}
	case *orderbook.Depth:
		base, err := d.Retrieve()
		if err != nil {
			return err
		}
		if m.syncer.IsRunning() {
			err := m.syncer.Update(exchName,
				synchronize.WebsocketUpdate,
				base.Pair,
				base.Asset,
				int(synchronize.Orderbook),
				nil)
			if err != nil {
				return err
			}
		}
		m.syncer.PrintOrderbookSummary(base, synchronize.WebsocketUpdate, nil)
	case *order.Detail:
		if !m.orderManager.Exists(d) {
			err := m.orderManager.Add(d)
			if err != nil {
				return err
			}
			m.printOrderSummary(d, false)
		} else {
			od, err := m.orderManager.GetByExchangeAndID(d.Exchange, d.OrderID)
			if err != nil {
				return err
			}
			err = od.UpdateOrderFromDetail(d)
			if err != nil {
				return err
			}

			err = m.orderManager.UpdateExistingOrder(od)
			if err != nil {
				return err
			}
			m.printOrderSummary(d, true)
		}
	case order.ClassificationError:
		return fmt.Errorf("%w %s", d.Err, d.Error())
	case stream.UnhandledMessageWarning:
		log.Warn(log.WebsocketMgr, d.Message)
	case account.Change:
		if m.verbose {
			m.printAccountHoldingsChangeSummary(d)
		}
	case []trade.Data:
		if m.verbose {
			log.Infof(log.Trade, "%+v", d)
		}
	case []fill.Data:
		if m.verbose {
			log.Infof(log.Fill, "%+v", d)
		}
	default:
		if m.verbose {
			log.Warnf(log.WebsocketMgr,
				"%s websocket Unknown type: %+v",
				exchName,
				d)
		}
	}
	return nil
}

// FormatCurrency is a method that formats and returns a currency pair
// based on the user currency display preferences
func (m *websocketRoutineManager) FormatCurrency(p currency.Pair) currency.Pair {
	if m == nil || atomic.LoadInt32(&m.started) == 0 {
		return p
	}
	return p.Format(*m.currencyConfig.CurrencyPairFormat)
}

// printOrderSummary this function will be deprecated when a order manager
// update is done.
func (m *websocketRoutineManager) printOrderSummary(o *order.Detail, isUpdate bool) {
	if m == nil || atomic.LoadInt32(&m.started) == 0 || o == nil {
		return
	}

	orderNotif := "New Order:"
	if isUpdate {
		orderNotif = "Order Change:"
	}

	log.Debugf(log.WebsocketMgr,
		"%s %s %s %s %s %s %s OrderID:%s ClientOrderID:%s Price:%f Amount:%f Executed Amount:%f Remaining Amount:%f",
		orderNotif,
		o.Exchange,
		o.AssetType,
		o.Pair,
		o.Status,
		o.Type,
		o.Side,
		o.OrderID,
		o.ClientOrderID,
		o.Price,
		o.Amount,
		o.ExecutedAmount,
		o.RemainingAmount)
}

// printAccountHoldingsChangeSummary this function will be deprecated when a
// account holdings update is done.
func (m *websocketRoutineManager) printAccountHoldingsChangeSummary(o account.Change) {
	if m == nil || atomic.LoadInt32(&m.started) == 0 {
		return
	}
	log.Debugf(log.WebsocketMgr,
		"Account Holdings Balance Changed: %s %s %s has changed balance by %f for account: %s",
		o.Exchange,
		o.Asset,
		o.Currency,
		o.Amount,
		o.Account)
}

// registerWebsocketDataHandler registers an externally (GCT Library) defined
// dedicated filter specific data types for internal & external strategy use.
// InterceptorOnly as true will purge all other registered handlers
// (including default) bypassing all other handling.
func (m *websocketRoutineManager) registerWebsocketDataHandler(fn WebsocketDataHandler, interceptorOnly bool) error {
	if m == nil {
		return fmt.Errorf("%T %w", m, subsystem.ErrNil)
	}

	if fn == nil {
		return errNilWebsocketDataHandlerFunction
	}

	if interceptorOnly {
		return m.setWebsocketDataHandler(fn)
	}

	m.mu.Lock()
	// Push front so that any registered data handler has first preference
	// over the gct default handler.
	m.dataHandlers = append([]WebsocketDataHandler{fn}, m.dataHandlers...)
	m.mu.Unlock()
	return nil
}

// setWebsocketDataHandler sets a single websocket data handler, removing all
// pre-existing handlers.
func (m *websocketRoutineManager) setWebsocketDataHandler(fn WebsocketDataHandler) error {
	if m == nil {
		return fmt.Errorf("%T %w", m, subsystem.ErrNil)
	}
	if fn == nil {
		return errNilWebsocketDataHandlerFunction
	}
	m.mu.Lock()
	m.dataHandlers = []WebsocketDataHandler{fn}
	m.mu.Unlock()
	return nil
}
