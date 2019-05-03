package exchange

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	log "github.com/thrasher-/gocryptotrader/logger"
)

// WebsocketInit initialises the websocket struct
func (e *Base) WebsocketInit() {
	e.Websocket = &Websocket{
		defaultURL: "",
		enabled:    false,
		proxyAddr:  "",
		runningURL: "",
		init:       true,
	}
}

// WebsocketReset sends the shutdown command, waits for finish and then reconnects
func (w *Websocket) WebsocketReset() {
	err := w.Shutdown()
	if err != nil {
		log.Debugf("shutdown error: %v", err)
	}
	log.Debug("Waiting for wait groups to exit...")
	w.Wg.Wait()
	log.Debug("Reconnecting")
	w.init = true
	err = w.Connect()
	if err != nil {
		log.Debugf("connection error: %v", err)
	}
}

// WebsocketSetup sets main variables for websocket connection
func (e *Base) WebsocketSetup(connector func() error,
	subscriber func(channelToSubscribe WebsocketChannelSubscription) error,
	unsubscriber func(channelToUnsubscribe WebsocketChannelSubscription) error,
	exchangeName string,
	wsEnabled bool,
	defaultURL,
	runningURL string) error {

	e.Websocket.DataHandler = make(chan interface{}, 1)
	e.Websocket.Connected = make(chan struct{}, 1)
	e.Websocket.Disconnected = make(chan struct{}, 1)
	e.Websocket.TrafficAlert = make(chan struct{}, 1)

	err := e.Websocket.SetWsStatusAndConnection(wsEnabled)
	if err != nil {
		return err
	}

	e.Websocket.SetChannelSubscriber(subscriber)
	e.Websocket.SetChannelUnsubscriber(unsubscriber)

	e.Websocket.SetDefaultURL(defaultURL)
	e.Websocket.SetConnector(connector)
	e.Websocket.SetWebsocketURL(runningURL)
	e.Websocket.SetExchangeName(exchangeName)

	e.Websocket.init = false

	return nil
}

// trafficMonitor monitors traffic and switches connection modes for websocket
func (w *Websocket) trafficMonitor(wg *sync.WaitGroup) {
	log.Debug("trafficMonitor HIT")
	w.Wg.Add(1)
	wg.Done() // Makes sure we are unlocking after we add to waitgroup

	defer func() {
		log.Debug("trafficMonitor DEFER FUNC EXIT")
		if w.connected {
			log.Debug("trafficMonitor SENDING DISCONNECT")
			w.Disconnected <- struct{}{}
		}
		w.Wg.Done()
	}()

	// Define an initial traffic timer which will be a delay then fall over to
	// WebsocketTrafficLimitTime after first response
	trafficTimer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-w.ShutdownC: // Returns on shutdown channel close
			log.Debug("trafficMonitor SHUTDOWN RECEIVED")
			return
		case <-w.TrafficAlert: // Resets timer on traffic
			if !w.connected {
				w.Connected <- struct{}{}
				log.Debug("--------------------- Connected True 1--------------------------")
				w.connected = true
			}
			trafficTimer.Reset(WebsocketTrafficLimitTime)
		case <-trafficTimer.C: // Falls through when timer runs out
			log.Debug("trafficMonitor FIRST TIMEOUT HIT")
			newtimer := time.NewTimer(10 * time.Second) // New secondary timer set
			if w.connected {
				//If connected divert traffic to rest
				w.Disconnected <- struct{}{}
				log.Debug("--------------------- Connected False 1--------------------------")
				w.connected = false
			}

			select {
			case <-w.ShutdownC: // Returns on shutdown channel close
				log.Debug("--------------------- Connected False 2--------------------------")
				log.Debug("trafficMonitor SHUTDOWN RECEIVED")
				w.connected = false
				return

			case <-newtimer.C: // If secondary timer runs state timeout is sent to the data handler
				log.Debug("trafficMonitor SEONCD TIMEOUT HIT")
				w.DataHandler <- fmt.Errorf("trafficMonitor %v", WebsocketStateTimeout)
				return

			case <-w.TrafficAlert: // If in this time response traffic comes through
				trafficTimer.Reset(WebsocketTrafficLimitTime)
				if !w.connected {
					// If not connected dive rt traffic from REST to websocket
					w.Connected <- struct{}{}
					log.Debug("--------------------- Connected True 2--------------------------")
					w.connected = true
				}
			}
		}
	}
}

// Connect intiates a websocket connection by using a package defined connection
// function
func (w *Websocket) Connect() error {
	w.m.Lock()
	defer w.m.Unlock()
	w.Connecting = true
	if !w.IsEnabled() {
		return errors.New(WebsocketNotEnabled)
	}

	if w.connected {
		w.Connecting = false
		return errors.New("exchange_websocket.go error - already connected, cannot connect again")
	}

	w.ShutdownC = make(chan struct{}, 1)

	err := w.connector()
	if err != nil {
		w.Connecting = false
		return fmt.Errorf("exchange_websocket.go connection error %s",
			err)
	}

	if !w.connected {
		w.Connected <- struct{}{}
		log.Debug("--------------------- Connected True 3--------------------------")
		w.connected = true
		w.Connecting = false
	}

	var anotherWG sync.WaitGroup
	anotherWG.Add(1)
	go w.trafficMonitor(&anotherWG)
	anotherWG.Wait()
	go w.wsConnectionMonitor()
	go w.manageSubscriptions()

	return nil
}

// WsConnectionMonitor ensures that the WS keeps connecting
func (w *Websocket) wsConnectionMonitor() {
	log.Debug("STARTING WsConnectionMonitor")
	noConnectionTolerance := 0
	defer func() {
		log.Debug("WsConnectionMonitor EXITING")
	}()
	for {
		if !w.enabled {
			log.Debug("WsConnectionMonitor: websocket disabled, shutting down")
			w.Shutdown()
			return
		}
		time.Sleep(2 * time.Second)
		if !w.IsConnected() && !w.Connecting {
			log.Debugf("No connection %v/5", noConnectionTolerance)
			if noConnectionTolerance >= 5 {
				log.Debug("Resetting connection")
				w.Connecting = true
				go w.WebsocketReset()
				noConnectionTolerance = 0
			}
			noConnectionTolerance++
		} else if w.Connecting {
			log.Debug("Busy reconnecting")
		} else {
			noConnectionTolerance = 0
		}
	}
}

// IsConnected exposes websocket connection status
func (w *Websocket) IsConnected() bool {
	return w.connected
}

// Shutdown attempts to shut down a websocket connection and associated routines
// by using a package defined shutdown function
func (w *Websocket) Shutdown() error {
	w.m.Lock()
	defer func() {
		w.Orderbook.FlushCache()
		w.m.Unlock()
	}()
	log.Debug("Shutting down channels")
	timer := time.NewTimer(15 * time.Second)
	c := make(chan struct{}, 1)

	go func(c chan struct{}) {
		close(w.ShutdownC)
		log.Debug("Shutting down ShutdownC Channels")
		w.Wg.Wait()
		log.Debug("completed waiting for wg")
		c <- struct{}{}
	}(c)

	select {
	case <-c:
		w.connected = false
		return nil
	case <-timer.C:
		log.Fatalf("%s - Websocket routines failed to shutdown after 15 seconds",
			w.GetName())
	}
	return nil
}

// SetWebsocketURL sets websocket URL
func (w *Websocket) SetWebsocketURL(websocketURL string) {
	if websocketURL == "" || websocketURL == config.WebsocketURLNonDefaultMessage {
		w.runningURL = w.defaultURL
		return
	}
	w.runningURL = websocketURL
}

// GetWebsocketURL returns the running websocket URL
func (w *Websocket) GetWebsocketURL() string {
	return w.runningURL
}

// SetWsStatusAndConnection sets if websocket is enabled
// it will also connect/disconnect the websocket connection
func (w *Websocket) SetWsStatusAndConnection(enabled bool) error {
	if w.enabled == enabled {
		if w.init {
			return nil
		}
		return fmt.Errorf("exchange_websocket.go error - already set as %t",
			enabled)
	}

	w.enabled = enabled

	if !w.init {
		if enabled {
			if w.connected {
				return nil
			}
			return w.Connect()
		}

		if !w.connected {
			return nil
		}
		return w.Shutdown()
	}
	return nil
}

// IsEnabled returns bool
func (w *Websocket) IsEnabled() bool {
	return w.enabled
}

// SetProxyAddress sets websocket proxy address
func (w *Websocket) SetProxyAddress(proxyAddr string) error {
	if w.proxyAddr == proxyAddr {
		return errors.New("exchange_websocket.go error - Setting proxy address - same address")
	}

	w.proxyAddr = proxyAddr

	if !w.init && w.enabled {
		if w.connected {
			err := w.Shutdown()
			if err != nil {
				return err
			}
			return w.Connect()
		}
		return w.Connect()
	}
	return nil
}

// GetProxyAddress returns the current websocket proxy
func (w *Websocket) GetProxyAddress() string {
	return w.proxyAddr
}

// SetDefaultURL sets default websocket URL
func (w *Websocket) SetDefaultURL(defaultURL string) {
	w.defaultURL = defaultURL
}

// GetDefaultURL returns the default websocket URL
func (w *Websocket) GetDefaultURL() string {
	return w.defaultURL
}

// SetConnector sets connection function
func (w *Websocket) SetConnector(connector func() error) {
	w.connector = connector
}

// SetExchangeName sets exchange name
func (w *Websocket) SetExchangeName(exchName string) {
	w.exchangeName = exchName
}

// GetName returns exchange name
func (w *Websocket) GetName() string {
	return w.exchangeName
}

// Update updates a local cache using bid targets and ask targets then updates
// main cache in orderbook.go
// Volume == 0; deletion at price target
// Price target not found; append of price target
// Price target found; amend volume of price target
func (w *WebsocketOrderbookLocal) Update(bidTargets, askTargets []orderbook.Item,
	p currency.Pair,
	updated time.Time,
	exchName, assetType string) error {
	if bidTargets == nil && askTargets == nil {
		return errors.New("exchange.go websocket orderbook cache Update() error - cannot have bids and ask targets both nil")
	}

	if w.lastUpdated.After(updated) {
		return errors.New("exchange.go WebsocketOrderbookLocal Update() - update is before last update time")
	}

	w.m.Lock()
	defer w.m.Unlock()

	var orderbookAddress *orderbook.Base
	for i := range w.ob {
		if w.ob[i].Pair == p && w.ob[i].AssetType == assetType {
			orderbookAddress = w.ob[i]
		}
	}

	if orderbookAddress == nil {
		return fmt.Errorf("exchange.go WebsocketOrderbookLocal Update() - orderbook.Base could not be found for Exchange %s CurrencyPair: %s AssetType: %s",
			exchName,
			p.String(),
			assetType)
	}

	if len(orderbookAddress.Asks) == 0 || len(orderbookAddress.Bids) == 0 {
		return errors.New("exchange.go websocket orderbook cache Update() error - snapshot incorrectly loaded")
	}

	if orderbookAddress.Pair == (currency.Pair{}) {
		return fmt.Errorf("exchange.go websocket orderbook cache Update() error - snapshot not found %v",
			p)
	}

	for x := range bidTargets {
		// bid targets
		func() {
			for y := range orderbookAddress.Bids {
				if orderbookAddress.Bids[y].Price == bidTargets[x].Price {
					if bidTargets[x].Amount == 0 {
						// Delete
						orderbookAddress.Bids = append(orderbookAddress.Bids[:y],
							orderbookAddress.Bids[y+1:]...)
						return
					}
					// Amend
					orderbookAddress.Bids[y].Amount = bidTargets[x].Amount
					return
				}
			}

			if bidTargets[x].Amount == 0 {
				// Makes sure we dont append things we missed
				return
			}

			// Append
			orderbookAddress.Bids = append(orderbookAddress.Bids, orderbook.Item{
				Price:  bidTargets[x].Price,
				Amount: bidTargets[x].Amount,
			})
		}()
		// bid targets
	}

	for x := range askTargets {
		func() {
			for y := range orderbookAddress.Asks {
				if orderbookAddress.Asks[y].Price == askTargets[x].Price {
					if askTargets[x].Amount == 0 {
						// Delete
						orderbookAddress.Asks = append(orderbookAddress.Asks[:y],
							orderbookAddress.Asks[y+1:]...)
						return
					}
					// Amend
					orderbookAddress.Asks[y].Amount = askTargets[x].Amount
					return
				}
			}

			if askTargets[x].Amount == 0 {
				// Makes sure we dont append things we missed
				return
			}

			// Append
			orderbookAddress.Asks = append(orderbookAddress.Asks, orderbook.Item{
				Price:  askTargets[x].Price,
				Amount: askTargets[x].Amount,
			})
		}()
	}

	return orderbookAddress.Process()

}

// LoadSnapshot loads initial snapshot of orderbook data, overite allows full
// orderbook to be completely rewritten because the exchange is a doing a full
// update not an incremental one
func (w *WebsocketOrderbookLocal) LoadSnapshot(newOrderbook *orderbook.Base, exchName string, overwrite bool) error {
	if len(newOrderbook.Asks) == 0 || len(newOrderbook.Bids) == 0 {
		return errors.New("exchange.go websocket orderbook cache LoadSnapshot() error - snapshot ask and bids are nil")
	}

	w.m.Lock()
	defer w.m.Unlock()

	for i := range w.ob {
		if w.ob[i].Pair.Equal(newOrderbook.Pair) && w.ob[i].AssetType == newOrderbook.AssetType {
			if overwrite {
				w.ob[i] = newOrderbook
				return newOrderbook.Process()
			}
			return errors.New("exchange.go websocket orderbook cache LoadSnapshot() error - Snapshot instance already found")
		}
	}

	w.ob = append(w.ob, newOrderbook)
	return newOrderbook.Process()
}

// UpdateUsingID updates orderbooks using specified ID
func (w *WebsocketOrderbookLocal) UpdateUsingID(bidTargets, askTargets []orderbook.Item,
	p currency.Pair,
	exchName, assetType, action string) error {
	w.m.Lock()
	defer w.m.Unlock()

	var orderbookAddress *orderbook.Base
	for i := range w.ob {
		if w.ob[i].Pair == p && w.ob[i].AssetType == assetType {
			orderbookAddress = w.ob[i]
		}
	}

	if orderbookAddress == nil {
		return fmt.Errorf("exchange.go WebsocketOrderbookLocal Update() - orderbook.Base could not be found for Exchange %s CurrencyPair: %s AssetType: %s",
			exchName,
			assetType,
			p.String())
	}

	switch action {
	case "update":
		for _, target := range bidTargets {
			for i := range orderbookAddress.Bids {
				if orderbookAddress.Bids[i].ID == target.ID {
					orderbookAddress.Bids[i].Amount = target.Amount
					break
				}
			}
		}

		for _, target := range askTargets {
			for i := range orderbookAddress.Asks {
				if orderbookAddress.Asks[i].ID == target.ID {
					orderbookAddress.Asks[i].Amount = target.Amount
					break
				}
			}
		}

	case "delete":
		for _, target := range bidTargets {
			for i := range orderbookAddress.Bids {
				if orderbookAddress.Bids[i].ID == target.ID {
					orderbookAddress.Bids = append(orderbookAddress.Bids[:i],
						orderbookAddress.Bids[i+1:]...)
					break
				}
			}
		}

		for _, target := range askTargets {
			for i := range orderbookAddress.Asks {
				if orderbookAddress.Asks[i].ID == target.ID {
					orderbookAddress.Asks = append(orderbookAddress.Asks[:i],
						orderbookAddress.Asks[i+1:]...)
					break
				}
			}
		}

	case "insert":
		orderbookAddress.Bids = append(orderbookAddress.Bids, bidTargets...)
		orderbookAddress.Asks = append(orderbookAddress.Asks, askTargets...)
	}

	return orderbookAddress.Process()
}

// FlushCache flushes w.ob data to be garbage collected and refreshed when a
// connection is lost and reconnected
func (w *WebsocketOrderbookLocal) FlushCache() {
	w.m.Lock()
	w.ob = nil
	w.m.Unlock()
}

// GetFunctionality returns a functionality bitmask for the websocket
// connection
func (w *Websocket) GetFunctionality() uint32 {
	return w.Functionality
}

// SupportsFunctionality returns if the functionality is supported as a boolean
func (w *Websocket) SupportsFunctionality(f uint32) bool {
	return w.GetFunctionality()&f == f
}

// FormatFunctionality will return each of the websocket connection compatible
// stream methods as a string
func (w *Websocket) FormatFunctionality() string {
	var functionality []string
	for i := 0; i < 32; i++ {
		var check uint32 = 1 << uint32(i)
		if w.GetFunctionality()&check != 0 {
			switch check {
			case WebsocketTickerSupported:
				functionality = append(functionality, WebsocketTickerSupportedText)

			case WebsocketOrderbookSupported:
				functionality = append(functionality, WebsocketOrderbookSupportedText)

			case WebsocketKlineSupported:
				functionality = append(functionality, WebsocketKlineSupportedText)

			case WebsocketTradeDataSupported:
				functionality = append(functionality, WebsocketTradeDataSupportedText)

			case WebsocketAccountSupported:
				functionality = append(functionality, WebsocketAccountSupportedText)

			case WebsocketAllowsRequests:
				functionality = append(functionality, WebsocketAllowsRequestsText)

			default:
				functionality = append(functionality,
					fmt.Sprintf("%s[1<<%v]", UnknownWebsocketFunctionality, i))
			}
		}
	}

	if len(functionality) > 0 {
		return strings.Join(functionality, " & ")
	}

	return NoWebsocketSupportText
}

// SetChannelSubscriber sets the function to use the base subscribe func
func (w *Websocket) SetChannelSubscriber(subscriber func(channelToSubscribe WebsocketChannelSubscription) error) {
	w.channelSubscriber = subscriber
}

// SetChannelUnsubscriber sets the function to use the base unsubscribe func
func (w *Websocket) SetChannelUnsubscriber(unsubscriber func(channelToUnsubscribe WebsocketChannelSubscription) error) {
	w.channelUnsubscriber = unsubscriber
} 

// ManageSubscriptions ensures the subscriptions specified continue to be subscribed to
func (w *Websocket) manageSubscriptions() {
	w.Wg.Add(1)
	defer func() {
		log.Debug("ManageSubscriptions EXITING")
		w.Wg.Done()
	}()
	for {
		select {
		case <-w.ShutdownC:
			w.subscribedChannels = []WebsocketChannelSubscription{}
			log.Debug("SHUTDOWN ManageSubscriptions")
			return
		default:
			time.Sleep(800 * time.Millisecond)
			log.Debug("Checking subscriptions")
			// Subscribe to channels Pending a subscription
			err := w.subscribeToChannels()
			if err != nil {
				w.DataHandler <- err
				if err == common.ErrFunctionNotSupported {
					log.Infof("Exchange %v does not support channel subscriptions, exiting ManageSubscriptions()", w.exchangeName)
					return
				}
			}
			// Unsubscribe from any channels removed from ChannelsToSubscribe
			err = w.unsubscribeToChannels()
			if err != nil {
				w.DataHandler <- err
				if err == common.ErrFunctionNotSupported {
					log.Infof("Exchange %v does not support channel subscriptions, exiting ManageSubscriptions()", w.exchangeName)
					return
				}
			}
		}
	}
}

// subscribeToChannels compares ChannelsToSubscribe to subscribedChannels
// and subscribes to any channels not present in  subscribedChannels
func (w *Websocket) subscribeToChannels() error {
	for i := range w.ChannelsToSubscribe {
		channelIsSubscribed := false
		for j := range w.subscribedChannels {
			if strings.EqualFold(w.ChannelsToSubscribe[i].Channel, w.subscribedChannels[j].Channel) &&
				strings.EqualFold(w.ChannelsToSubscribe[i].Currency.String(), w.subscribedChannels[j].Currency.String()) {
				channelIsSubscribed = true
			}
		}
		if !channelIsSubscribed {
			err := w.channelSubscriber(w.ChannelsToSubscribe[i])
			if err != nil {
				return err
			}
			w.subscribedChannels = append(w.subscribedChannels, w.ChannelsToSubscribe[i])
		}
	}
	return nil
}

// unsubscribeToChannels compares subscribedChannels to ChannelsToSubscribe
// and unsubscribes to any channels not present in  ChannelsToSubscribe
func (w *Websocket) unsubscribeToChannels() error {
	for i := range w.subscribedChannels {
		subscriptionFound := false
		for j := range w.ChannelsToSubscribe {
			if strings.EqualFold(w.ChannelsToSubscribe[i].Channel, w.subscribedChannels[j].Channel) &&
				strings.EqualFold(w.ChannelsToSubscribe[i].Currency.String(), w.subscribedChannels[j].Currency.String()) {
					subscriptionFound = true
			}
		}
		if !subscriptionFound {
			err := w.channelUnsubscriber(w.subscribedChannels[i])
			if err != nil {
				return err
			}
			w.subscribedChannels = append(w.subscribedChannels[:i], w.subscribedChannels[i+1:]...)
		}
	}
	return nil
}

// RemoveChannelToSubscribe removes an entry from w.ChannelsToSubscribe 
// so an unsubscribe event can be triggered
func (w *Websocket) RemoveChannelToSubscribe(subscribedChannel WebsocketChannelSubscription) {
	channelRemoved := false
	for i := range w.ChannelsToSubscribe {
		if strings.EqualFold(w.ChannelsToSubscribe[i].Channel, subscribedChannel.Channel) &&
			strings.EqualFold(w.ChannelsToSubscribe[i].Currency.String(), subscribedChannel.Currency.String()) {
			w.ChannelsToSubscribe = append(w.ChannelsToSubscribe[:i], w.ChannelsToSubscribe[i+1:]...)
			channelRemoved = true
			break
		}
	}
	if !channelRemoved {
		w.DataHandler <- fmt.Errorf("%v RemoveChannelToSubscribe() Channel %v Currency %v could not be removed because it was not found",
		w.exchangeName, 
		subscribedChannel.Channel,
		subscribedChannel.Currency)
		return
	}
	w.channelUnsubscriber(subscribedChannel)
}

// ResubscribeToChannel calls unsubscribe func and 
// removes it from subscribedChannels to trigger a subscribe evbent
func (w *Websocket) ResubscribeToChannel(subscribedChannel WebsocketChannelSubscription) {
	err := w.channelUnsubscriber(subscribedChannel)
	if err != nil {
		w.DataHandler <- err
	}
	// Remove the channel from the list of subscribed channels
	// ManageSubscriptions will automatically resubscribe
	for i := range w.subscribedChannels {
		if strings.EqualFold(w.ChannelsToSubscribe[i].Channel, subscribedChannel.Channel) &&
		strings.EqualFold(w.ChannelsToSubscribe[i].Currency.String(), subscribedChannel.Currency.String()) {
			w.subscribedChannels = append(w.ChannelsToSubscribe[:i], w.ChannelsToSubscribe[i+1:]...)
			break
		}
	}
}
 