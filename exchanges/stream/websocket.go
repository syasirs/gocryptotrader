package stream

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	defaultJobBuffer = 5000
	// defaultTrafficPeriod defines a period of pause for the traffic monitor,
	// as there are periods with large incoming traffic alerts which requires a
	// timer reset, this limits work on this routine to a more effective rate
	// of check.
	defaultTrafficPeriod = time.Second
)

var (
	// ErrSubscriptionNotFound defines an error when a subscription is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrSubscribedAlready defines an error when a channel is already subscribed
	ErrSubscribedAlready = errors.New("duplicate subscription")
	// ErrSubscriptionFailure defines an error when a subscription fails
	ErrSubscriptionFailure = errors.New("subscription failure")
	// ErrUnsubscribeFailure defines an error when a unsubscribe fails
	ErrUnsubscribeFailure = errors.New("unsubscribe failure")
	// ErrChannelInStateAlready defines an error when a subscription channel is already in a new state
	ErrChannelInStateAlready = errors.New("channel already in state")
	// ErrAlreadyDisabled is returned when you double-disable the websocket
	ErrAlreadyDisabled = errors.New("websocket already disabled")
	// ErrNotConnected defines an error when websocket is not connected
	ErrNotConnected = errors.New("websocket is not connected")

	errAlreadyRunning                       = errors.New("connection monitor is already running")
	errExchangeConfigIsNil                  = errors.New("exchange config is nil")
	errWebsocketIsNil                       = errors.New("websocket is nil")
	errWebsocketSetupIsNil                  = errors.New("websocket setup is nil")
	errWebsocketAlreadyInitialised          = errors.New("websocket already initialised")
	errWebsocketFeaturesIsUnset             = errors.New("websocket features is unset")
	errConfigFeaturesIsNil                  = errors.New("exchange config features is nil")
	errDefaultURLIsEmpty                    = errors.New("default url is empty")
	errRunningURLIsEmpty                    = errors.New("running url cannot be empty")
	errInvalidWebsocketURL                  = errors.New("invalid websocket url")
	errExchangeConfigNameUnset              = errors.New("exchange config name unset")
	errInvalidTrafficTimeout                = errors.New("invalid traffic timeout")
	errWebsocketSubscriberUnset             = errors.New("websocket subscriber function needs to be set")
	errWebsocketUnsubscriberUnset           = errors.New("websocket unsubscriber functionality allowed but unsubscriber function not set")
	errWebsocketConnectorUnset              = errors.New("websocket connector function not set")
	errWebsocketSubscriptionsGeneratorUnset = errors.New("websocket subscriptions generator function needs to be set")
	errClosedConnection                     = errors.New("use of closed network connection")
	errSubscriptionsExceedsLimit            = errors.New("subscriptions exceeds limit")
	errInvalidMaxSubscriptions              = errors.New("max subscriptions cannot be less than 0")
	errNoSubscriptionsSupplied              = errors.New("no subscriptions supplied")
	errChannelAlreadySubscribed             = errors.New("channel already subscribed")
	errInvalidChannelState                  = errors.New("invalid Channel state")
)

var globalReporter Reporter

// SetupGlobalReporter sets a reporter interface to be used
// for all exchange requests
func SetupGlobalReporter(r Reporter) {
	globalReporter = r
}

// New initialises the websocket struct
func New() *Websocket {
	return &Websocket{
		Init:              true,
		DataHandler:       make(chan interface{}, defaultJobBuffer),
		ToRoutine:         make(chan interface{}, defaultJobBuffer),
		TrafficAlert:      make(chan struct{}),
		ReadMessageErrors: make(chan error),
		Subscribe:         make(chan []subscription.Subscription),
		Unsubscribe:       make(chan []subscription.Subscription),
		Match:             NewMatch(),
	}
}

// Setup sets main variables for websocket connection
func (w *Websocket) Setup(s *WebsocketSetup) error {
	if w == nil {
		return errWebsocketIsNil
	}

	if s == nil {
		return errWebsocketSetupIsNil
	}

	if !w.Init {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketAlreadyInitialised)
	}

	if s.ExchangeConfig == nil {
		return errExchangeConfigIsNil
	}

	if s.ExchangeConfig.Name == "" {
		return errExchangeConfigNameUnset
	}
	w.exchangeName = s.ExchangeConfig.Name
	w.verbose = s.ExchangeConfig.Verbose

	if s.Features == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketFeaturesIsUnset)
	}
	w.features = s.Features

	if s.ExchangeConfig.Features == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errConfigFeaturesIsNil)
	}
	w.enabled = s.ExchangeConfig.Features.Enabled.Websocket

	if s.Connector == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketConnectorUnset)
	}
	w.connector = s.Connector

	if s.Subscriber == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketSubscriberUnset)
	}
	w.Subscriber = s.Subscriber

	if w.features.Unsubscribe && s.Unsubscriber == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketUnsubscriberUnset)
	}
	w.connectionMonitorDelay = s.ExchangeConfig.ConnectionMonitorDelay
	if w.connectionMonitorDelay <= 0 {
		w.connectionMonitorDelay = config.DefaultConnectionMonitorDelay
	}
	w.Unsubscriber = s.Unsubscriber

	if s.GenerateSubscriptions == nil {
		return fmt.Errorf("%s %w", w.exchangeName, errWebsocketSubscriptionsGeneratorUnset)
	}
	w.GenerateSubs = s.GenerateSubscriptions

	if s.DefaultURL == "" {
		return fmt.Errorf("%s websocket %w", w.exchangeName, errDefaultURLIsEmpty)
	}
	w.defaultURL = s.DefaultURL
	if s.RunningURL == "" {
		return fmt.Errorf("%s websocket %w", w.exchangeName, errRunningURLIsEmpty)
	}
	err := w.SetWebsocketURL(s.RunningURL, false, false)
	if err != nil {
		return fmt.Errorf("%s %w", w.exchangeName, err)
	}

	if s.RunningURLAuth != "" {
		err = w.SetWebsocketURL(s.RunningURLAuth, true, false)
		if err != nil {
			return fmt.Errorf("%s %w", w.exchangeName, err)
		}
	}

	if s.ExchangeConfig.WebsocketTrafficTimeout < time.Second {
		return fmt.Errorf("%s %w cannot be less than %s",
			w.exchangeName,
			errInvalidTrafficTimeout,
			time.Second)
	}
	w.trafficTimeout = s.ExchangeConfig.WebsocketTrafficTimeout

	w.ShutdownC = make(chan struct{})
	w.Wg = new(sync.WaitGroup)
	w.SetCanUseAuthenticatedEndpoints(s.ExchangeConfig.API.AuthenticatedWebsocketSupport)

	if err := w.Orderbook.Setup(s.ExchangeConfig, &s.OrderbookBufferConfig, w.DataHandler); err != nil {
		return err
	}

	w.Trade.Setup(w.exchangeName, s.TradeFeed, w.DataHandler)
	w.Fills.Setup(s.FillsFeed, w.DataHandler)

	if s.MaxWebsocketSubscriptionsPerConnection < 0 {
		return fmt.Errorf("%s %w", w.exchangeName, errInvalidMaxSubscriptions)
	}
	w.MaxSubscriptionsPerConnection = s.MaxWebsocketSubscriptionsPerConnection
	return nil
}

// SetupNewConnection sets up an auth or unauth streaming connection
func (w *Websocket) SetupNewConnection(c ConnectionSetup) error {
	if w == nil {
		return errors.New("setting up new connection error: websocket is nil")
	}
	if c == (ConnectionSetup{}) {
		return errors.New("setting up new connection error: websocket connection configuration empty")
	}

	if w.exchangeName == "" {
		return errors.New("setting up new connection error: exchange name not set, please call setup first")
	}

	if w.TrafficAlert == nil {
		return errors.New("setting up new connection error: traffic alert is nil, please call setup first")
	}

	if w.ReadMessageErrors == nil {
		return errors.New("setting up new connection error: read message errors is nil, please call setup first")
	}

	connectionURL := w.GetWebsocketURL()
	if c.URL != "" {
		connectionURL = c.URL
	}

	if c.ConnectionLevelReporter == nil {
		c.ConnectionLevelReporter = w.ExchangeLevelReporter
	}

	if c.ConnectionLevelReporter == nil {
		c.ConnectionLevelReporter = globalReporter
	}

	newConn := &WebsocketConnection{
		ExchangeName:      w.exchangeName,
		URL:               connectionURL,
		ProxyURL:          w.GetProxyAddress(),
		Verbose:           w.verbose,
		ResponseMaxLimit:  c.ResponseMaxLimit,
		Traffic:           w.TrafficAlert,
		readMessageErrors: w.ReadMessageErrors,
		ShutdownC:         w.ShutdownC,
		Wg:                w.Wg,
		Match:             w.Match,
		RateLimit:         c.RateLimit,
		Reporter:          c.ConnectionLevelReporter,
	}

	if c.Authenticated {
		w.AuthConn = newConn
	} else {
		w.Conn = newConn
	}

	return nil
}

// Connect initiates a websocket connection by using a package defined connection
// function
func (w *Websocket) Connect() error {
	if w.connector == nil {
		return errors.New("websocket connect function not set, cannot continue")
	}
	w.m.Lock()
	defer w.m.Unlock()

	if !w.IsEnabled() {
		return errors.New(WebsocketNotEnabled)
	}
	if w.IsConnecting() {
		return fmt.Errorf("%v Websocket already attempting to connect",
			w.exchangeName)
	}
	if w.IsConnected() {
		return fmt.Errorf("%v Websocket already connected",
			w.exchangeName)
	}

	w.dataMonitor()
	w.trafficMonitor()
	w.setConnectingStatus(true)

	err := w.connector()
	if err != nil {
		w.setConnectingStatus(false)
		return fmt.Errorf("%v Error connecting %s",
			w.exchangeName, err)
	}
	w.setConnectedStatus(true)
	w.setConnectingStatus(false)
	w.setInit(true)

	if !w.IsConnectionMonitorRunning() {
		err = w.connectionMonitor()
		if err != nil {
			log.Errorf(log.WebsocketMgr,
				"%s cannot start websocket connection monitor %v",
				w.GetName(),
				err)
		}
	}

	subs, err := w.GenerateSubs() // regenerate state on new connection
	if err != nil {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, common.AppendError(ErrSubscriptionFailure, err))
	}
	if len(subs) == 0 {
		return nil
	}
	err = w.checkSubscriptions(subs)
	if err != nil {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, common.AppendError(ErrSubscriptionFailure, err))
	}
	err = w.Subscriber(subs)
	if err != nil {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, common.AppendError(ErrSubscriptionFailure, err))
	}
	return nil
}

// Disable disables the exchange websocket protocol
func (w *Websocket) Disable() error {
	if !w.IsEnabled() {
		return fmt.Errorf("%w for exchange '%s'", ErrAlreadyDisabled, w.exchangeName)
	}

	w.setEnabled(false)
	return nil
}

// Enable enables the exchange websocket protocol
func (w *Websocket) Enable() error {
	if w.IsConnected() || w.IsEnabled() {
		return fmt.Errorf("websocket is already enabled for exchange %s",
			w.exchangeName)
	}

	w.setEnabled(true)
	return w.Connect()
}

// dataMonitor monitors job throughput and logs if there is a back log of data
func (w *Websocket) dataMonitor() {
	if w.IsDataMonitorRunning() {
		return
	}
	w.setDataMonitorRunning(true)
	w.Wg.Add(1)

	go func() {
		defer func() {
			for {
				// Bleeds data from the websocket connection if needed
				select {
				case <-w.DataHandler:
				default:
					w.setDataMonitorRunning(false)
					w.Wg.Done()
					return
				}
			}
		}()

		for {
			select {
			case <-w.ShutdownC:
				return
			case d := <-w.DataHandler:
				select {
				case w.ToRoutine <- d:
				case <-w.ShutdownC:
					return
				default:
					log.Warnf(log.WebsocketMgr,
						"%s exchange backlog in websocket processing detected",
						w.exchangeName)
					select {
					case w.ToRoutine <- d:
					case <-w.ShutdownC:
						return
					}
				}
			}
		}
	}()
}

// connectionMonitor ensures that the WS keeps connecting
func (w *Websocket) connectionMonitor() error {
	if w.checkAndSetMonitorRunning() {
		return errAlreadyRunning
	}
	w.fieldMutex.RLock()
	delay := w.connectionMonitorDelay
	w.fieldMutex.RUnlock()

	go func() {
		timer := time.NewTimer(delay)
		for {
			if w.verbose {
				log.Debugf(log.WebsocketMgr,
					"%v websocket: running connection monitor cycle\n",
					w.exchangeName)
			}
			if !w.IsEnabled() {
				if w.verbose {
					log.Debugf(log.WebsocketMgr,
						"%v websocket: connectionMonitor - websocket disabled, shutting down\n",
						w.exchangeName)
				}
				if w.IsConnected() {
					err := w.Shutdown()
					if err != nil {
						log.Errorln(log.WebsocketMgr, err)
					}
				}
				if w.verbose {
					log.Debugf(log.WebsocketMgr,
						"%v websocket: connection monitor exiting\n",
						w.exchangeName)
				}
				timer.Stop()
				w.setConnectionMonitorRunning(false)
				return
			}
			select {
			case err := <-w.ReadMessageErrors:
				if IsDisconnectionError(err) {
					w.setInit(false)
					log.Warnf(log.WebsocketMgr,
						"%v websocket has been disconnected. Reason: %v",
						w.exchangeName, err)
					w.setConnectedStatus(false)
				}

				w.DataHandler <- err
			case <-timer.C:
				if !w.IsConnecting() && !w.IsConnected() {
					err := w.Connect()
					if err != nil {
						log.Errorln(log.WebsocketMgr, err)
					}
				}
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(delay)
			}
		}
	}()
	return nil
}

// Shutdown attempts to shut down a websocket connection and associated routines
// by using a package defined shutdown function
func (w *Websocket) Shutdown() error {
	w.m.Lock()
	defer w.m.Unlock()

	if !w.IsConnected() {
		return fmt.Errorf("%v websocket: cannot shutdown %w",
			w.exchangeName,
			ErrNotConnected)
	}

	// TODO: Interrupt connection and or close connection when it is re-established.
	if w.IsConnecting() {
		return fmt.Errorf("%v websocket: cannot shutdown, in the process of reconnection",
			w.exchangeName)
	}

	if w.verbose {
		log.Debugf(log.WebsocketMgr,
			"%v websocket: shutting down websocket\n",
			w.exchangeName)
	}

	defer w.Orderbook.FlushBuffer()

	if w.Conn != nil {
		if err := w.Conn.Shutdown(); err != nil {
			return err
		}
	}

	if w.AuthConn != nil {
		if err := w.AuthConn.Shutdown(); err != nil {
			return err
		}
	}

	// flush any subscriptions from last connection if needed
	w.subscriptionMutex.Lock()
	w.subscriptions = subscriptionMap{}
	w.subscriptionMutex.Unlock()

	close(w.ShutdownC)
	w.Wg.Wait()
	w.ShutdownC = make(chan struct{})
	w.setConnectedStatus(false)
	w.setConnectingStatus(false)
	if w.verbose {
		log.Debugf(log.WebsocketMgr,
			"%v websocket: completed websocket shutdown\n",
			w.exchangeName)
	}
	return nil
}

// FlushChannels flushes channel subscriptions when there is a pair/asset change
func (w *Websocket) FlushChannels() error {
	if !w.IsEnabled() {
		return fmt.Errorf("%s websocket: service not enabled", w.exchangeName)
	}

	if !w.IsConnected() {
		return fmt.Errorf("%s websocket: service not connected", w.exchangeName)
	}

	if w.features.Subscribe {
		newsubs, err := w.GenerateSubs()
		if err != nil {
			return err
		}

		subs, unsubs := w.GetChannelDifference(newsubs)
		if w.features.Unsubscribe {
			if len(unsubs) != 0 {
				err := w.UnsubscribeChannels(unsubs)
				if err != nil {
					return err
				}
			}
		}

		if len(subs) < 1 {
			return nil
		}
		return w.SubscribeToChannels(subs)
	} else if w.features.FullPayloadSubscribe {
		// FullPayloadSubscribe means that the endpoint requires all
		// subscriptions to be sent via the websocket connection e.g. if you are
		// subscribed to ticker and orderbook but require trades as well, you
		// would need to send ticker, orderbook and trades channel subscription
		// messages.
		newsubs, err := w.GenerateSubs()
		if err != nil {
			return err
		}

		if len(newsubs) != 0 {
			// Purge subscription list as there will be conflicts
			w.subscriptionMutex.Lock()
			w.subscriptions = subscriptionMap{}
			w.subscriptionMutex.Unlock()
			return w.SubscribeToChannels(newsubs)
		}
		return nil
	}

	if err := w.Shutdown(); err != nil {
		return err
	}
	return w.Connect()
}

// trafficMonitor uses a timer of WebsocketTrafficLimitTime and once it expires,
// it will reconnect if the TrafficAlert channel has not received any data. The
// trafficTimer will reset on each traffic alert
func (w *Websocket) trafficMonitor() {
	if w.IsTrafficMonitorRunning() {
		return
	}
	w.setTrafficMonitorRunning(true)
	w.Wg.Add(1)

	go func() {
		var trafficTimer = time.NewTimer(w.trafficTimeout)
		pause := make(chan struct{})
		for {
			select {
			case <-w.ShutdownC:
				if w.verbose {
					log.Debugf(log.WebsocketMgr,
						"%v websocket: trafficMonitor shutdown message received\n",
						w.exchangeName)
				}
				trafficTimer.Stop()
				w.setTrafficMonitorRunning(false)
				w.Wg.Done()
				return
			case <-w.TrafficAlert:
				if !trafficTimer.Stop() {
					select {
					case <-trafficTimer.C:
					default:
					}
				}
				w.setConnectedStatus(true)
				trafficTimer.Reset(w.trafficTimeout)
			case <-trafficTimer.C: // Falls through when timer runs out
				if w.verbose {
					log.Warnf(log.WebsocketMgr,
						"%v websocket: has not received a traffic alert in %v. Reconnecting",
						w.exchangeName,
						w.trafficTimeout)
				}
				trafficTimer.Stop()
				w.setTrafficMonitorRunning(false)
				w.Wg.Done() // without this the w.Shutdown() call below will deadlock
				if !w.IsConnecting() && w.IsConnected() {
					err := w.Shutdown()
					if err != nil {
						log.Errorf(log.WebsocketMgr,
							"%v websocket: trafficMonitor shutdown err: %s",
							w.exchangeName, err)
					}
				}

				return
			}

			if w.IsConnected() {
				// Routine pausing mechanism
				go func(p chan<- struct{}) {
					time.Sleep(defaultTrafficPeriod)
					select {
					case p <- struct{}{}:
					default:
					}
				}(pause)
				select {
				case <-w.ShutdownC:
					trafficTimer.Stop()
					w.setTrafficMonitorRunning(false)
					w.Wg.Done()
					return
				case <-pause:
				}
			}
		}
	}()
}

func (w *Websocket) setConnectedStatus(b bool) {
	w.fieldMutex.Lock()
	w.connected = b
	w.fieldMutex.Unlock()
}

// IsConnected returns status of connection
func (w *Websocket) IsConnected() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.connected
}

func (w *Websocket) setConnectingStatus(b bool) {
	w.fieldMutex.Lock()
	w.connecting = b
	w.fieldMutex.Unlock()
}

// IsConnecting returns status of connecting
func (w *Websocket) IsConnecting() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.connecting
}

func (w *Websocket) setEnabled(b bool) {
	w.fieldMutex.Lock()
	w.enabled = b
	w.fieldMutex.Unlock()
}

// IsEnabled returns status of enabled
func (w *Websocket) IsEnabled() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.enabled
}

func (w *Websocket) setInit(b bool) {
	w.fieldMutex.Lock()
	w.Init = b
	w.fieldMutex.Unlock()
}

// IsInit returns status of init
func (w *Websocket) IsInit() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.Init
}

func (w *Websocket) setTrafficMonitorRunning(b bool) {
	w.fieldMutex.Lock()
	w.trafficMonitorRunning = b
	w.fieldMutex.Unlock()
}

// IsTrafficMonitorRunning returns status of the traffic monitor
func (w *Websocket) IsTrafficMonitorRunning() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.trafficMonitorRunning
}

func (w *Websocket) checkAndSetMonitorRunning() (alreadyRunning bool) {
	w.fieldMutex.Lock()
	defer w.fieldMutex.Unlock()
	if w.connectionMonitorRunning {
		return true
	}
	w.connectionMonitorRunning = true
	return false
}

func (w *Websocket) setConnectionMonitorRunning(b bool) {
	w.fieldMutex.Lock()
	w.connectionMonitorRunning = b
	w.fieldMutex.Unlock()
}

// IsConnectionMonitorRunning returns status of connection monitor
func (w *Websocket) IsConnectionMonitorRunning() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.connectionMonitorRunning
}

func (w *Websocket) setDataMonitorRunning(b bool) {
	w.fieldMutex.Lock()
	w.dataMonitorRunning = b
	w.fieldMutex.Unlock()
}

// IsDataMonitorRunning returns status of data monitor
func (w *Websocket) IsDataMonitorRunning() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.dataMonitorRunning
}

// CanUseAuthenticatedWebsocketForWrapper Handles a common check to
// verify whether a wrapper can use an authenticated websocket endpoint
func (w *Websocket) CanUseAuthenticatedWebsocketForWrapper() bool {
	if w.IsConnected() && w.CanUseAuthenticatedEndpoints() {
		return true
	} else if w.IsConnected() && !w.CanUseAuthenticatedEndpoints() {
		log.Infof(log.WebsocketMgr,
			WebsocketNotAuthenticatedUsingRest,
			w.exchangeName)
	}
	return false
}

// SetWebsocketURL sets websocket URL and can refresh underlying connections
func (w *Websocket) SetWebsocketURL(url string, auth, reconnect bool) error {
	defaultVals := url == "" || url == config.WebsocketURLNonDefaultMessage
	if auth {
		if defaultVals {
			url = w.defaultURLAuth
		}

		err := checkWebsocketURL(url)
		if err != nil {
			return err
		}
		w.runningURLAuth = url

		if w.verbose {
			log.Debugf(log.WebsocketMgr,
				"%s websocket: setting authenticated websocket URL: %s\n",
				w.exchangeName,
				url)
		}

		if w.AuthConn != nil {
			w.AuthConn.SetURL(url)
		}
	} else {
		if defaultVals {
			url = w.defaultURL
		}
		err := checkWebsocketURL(url)
		if err != nil {
			return err
		}
		w.runningURL = url

		if w.verbose {
			log.Debugf(log.WebsocketMgr,
				"%s websocket: setting unauthenticated websocket URL: %s\n",
				w.exchangeName,
				url)
		}

		if w.Conn != nil {
			w.Conn.SetURL(url)
		}
	}

	if w.IsConnected() && reconnect {
		log.Debugf(log.WebsocketMgr,
			"%s websocket: flushing websocket connection to %s\n",
			w.exchangeName,
			url)
		return w.Shutdown()
	}
	return nil
}

// GetWebsocketURL returns the running websocket URL
func (w *Websocket) GetWebsocketURL() string {
	return w.runningURL
}

// SetProxyAddress sets websocket proxy address
func (w *Websocket) SetProxyAddress(proxyAddr string) error {
	if proxyAddr != "" {
		_, err := url.ParseRequestURI(proxyAddr)
		if err != nil {
			return fmt.Errorf("%v websocket: cannot set proxy address error '%v'",
				w.exchangeName,
				err)
		}

		if w.proxyAddr == proxyAddr {
			return fmt.Errorf("%v websocket: cannot set proxy address to the same address '%v'",
				w.exchangeName,
				w.proxyAddr)
		}

		log.Debugf(log.ExchangeSys,
			"%s websocket: setting websocket proxy: %s\n",
			w.exchangeName,
			proxyAddr)
	} else {
		log.Debugf(log.ExchangeSys,
			"%s websocket: removing websocket proxy\n",
			w.exchangeName)
	}

	if w.Conn != nil {
		w.Conn.SetProxy(proxyAddr)
	}
	if w.AuthConn != nil {
		w.AuthConn.SetProxy(proxyAddr)
	}

	w.proxyAddr = proxyAddr
	if w.IsInit() && w.IsEnabled() {
		if w.IsConnected() {
			err := w.Shutdown()
			if err != nil {
				return err
			}
		}
		return w.Connect()
	}
	return nil
}

// GetProxyAddress returns the current websocket proxy
func (w *Websocket) GetProxyAddress() string {
	return w.proxyAddr
}

// GetName returns exchange name
func (w *Websocket) GetName() string {
	return w.exchangeName
}

// GetChannelDifference finds the difference between the subscribed channels
// and the new subscription list when pairs are disabled or enabled.
func (w *Websocket) GetChannelDifference(genSubs []subscription.Subscription) (sub, unsub []subscription.Subscription) {
	w.subscriptionMutex.RLock()
	unsubMap := make(map[any]subscription.Subscription, len(w.subscriptions))
	for k, c := range w.subscriptions {
		unsubMap[k] = *c
	}
	w.subscriptionMutex.RUnlock()

	for i := range genSubs {
		key := genSubs[i].EnsureKeyed()
		if _, ok := unsubMap[key]; ok {
			delete(unsubMap, key) // If it's in both then we remove it from the unsubscribe list
		} else {
			sub = append(sub, genSubs[i]) // If it's in genSubs but not existing subs we want to subscribe
		}
	}

	for _, c := range unsubMap {
		unsub = append(unsub, c)
	}

	return
}

// UnsubscribeChannels unsubscribes from a websocket channel
func (w *Websocket) UnsubscribeChannels(channels []subscription.Subscription) error {
	if len(channels) == 0 {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, errNoSubscriptionsSupplied)
	}
	w.subscriptionMutex.RLock()

	for i := range channels {
		key := channels[i].EnsureKeyed()
		if _, ok := w.subscriptions[key]; !ok {
			w.subscriptionMutex.RUnlock()
			return fmt.Errorf("%s websocket: %w: %+v", w.exchangeName, ErrSubscriptionNotFound, channels[i])
		}
	}
	w.subscriptionMutex.RUnlock()
	return w.Unsubscriber(channels)
}

// ResubscribeToChannel resubscribes to channel
func (w *Websocket) ResubscribeToChannel(subscribedChannel *subscription.Subscription) error {
	err := w.UnsubscribeChannels([]subscription.Subscription{*subscribedChannel})
	if err != nil {
		return err
	}
	return w.SubscribeToChannels([]subscription.Subscription{*subscribedChannel})
}

// SubscribeToChannels appends supplied channels to channelsToSubscribe
func (w *Websocket) SubscribeToChannels(channels []subscription.Subscription) error {
	if err := w.checkSubscriptions(channels); err != nil {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, common.AppendError(ErrSubscriptionFailure, err))
	}
	if err := w.Subscriber(channels); err != nil {
		return fmt.Errorf("%s websocket: %w", w.exchangeName, common.AppendError(ErrSubscriptionFailure, err))
	}
	return nil
}

// AddSubscription adds a subscription to the subscription lists
// Unlike AddSubscriptions this method will error if the subscription already exists
func (w *Websocket) AddSubscription(c *subscription.Subscription) error {
	w.subscriptionMutex.Lock()
	defer w.subscriptionMutex.Unlock()
	if w.subscriptions == nil {
		w.subscriptions = subscriptionMap{}
	}
	key := c.EnsureKeyed()
	if _, ok := w.subscriptions[key]; ok {
		return ErrSubscribedAlready
	}

	n := *c // Fresh copy; we don't want to use the pointer we were given and allow encapsulation/locks to be bypassed
	w.subscriptions[key] = &n

	return nil
}

// SetSubscriptionState sets an existing subscription state
// returns an error if the subscription is not found, or the new state is already set
func (w *Websocket) SetSubscriptionState(c *subscription.Subscription, state subscription.State) error {
	w.subscriptionMutex.Lock()
	defer w.subscriptionMutex.Unlock()
	if w.subscriptions == nil {
		w.subscriptions = subscriptionMap{}
	}
	key := c.EnsureKeyed()
	p, ok := w.subscriptions[key]
	if !ok {
		return ErrSubscriptionNotFound
	}
	if state == p.State {
		return ErrChannelInStateAlready
	}
	if state > subscription.UnsubscribingState {
		return errInvalidChannelState
	}
	p.State = state
	return nil
}

// AddSuccessfulSubscriptions adds subscriptions to the subscription lists that
// has been successfully subscribed
func (w *Websocket) AddSuccessfulSubscriptions(channels ...subscription.Subscription) {
	w.subscriptionMutex.Lock()
	defer w.subscriptionMutex.Unlock()
	if w.subscriptions == nil {
		w.subscriptions = subscriptionMap{}
	}
	for _, cN := range channels {
		c := cN // cN is an iteration var; Not safe to make a pointer to
		key := c.EnsureKeyed()
		c.State = subscription.SubscribedState
		w.subscriptions[key] = &c
	}
}

// RemoveSubscriptions removes subscriptions from the subscription list
func (w *Websocket) RemoveSubscriptions(channels ...subscription.Subscription) {
	w.subscriptionMutex.Lock()
	defer w.subscriptionMutex.Unlock()
	if w.subscriptions == nil {
		w.subscriptions = subscriptionMap{}
	}
	for i := range channels {
		key := channels[i].EnsureKeyed()
		delete(w.subscriptions, key)
	}
}

// GetSubscription returns a pointer to a copy of the subscription at the key provided
// returns nil if no subscription is at that key or the key is nil
func (w *Websocket) GetSubscription(key any) *subscription.Subscription {
	if key == nil || w == nil || w.subscriptions == nil {
		return nil
	}
	w.subscriptionMutex.RLock()
	defer w.subscriptionMutex.RUnlock()
	if s, ok := w.subscriptions[key]; ok {
		c := *s
		return &c
	}
	return nil
}

// GetSubscriptions returns a new slice of the subscriptions
func (w *Websocket) GetSubscriptions() []subscription.Subscription {
	w.subscriptionMutex.RLock()
	defer w.subscriptionMutex.RUnlock()
	subs := make([]subscription.Subscription, 0, len(w.subscriptions))
	for _, c := range w.subscriptions {
		subs = append(subs, *c)
	}
	return subs
}

// SetCanUseAuthenticatedEndpoints sets canUseAuthenticatedEndpoints val in
// a thread safe manner
func (w *Websocket) SetCanUseAuthenticatedEndpoints(val bool) {
	w.fieldMutex.Lock()
	defer w.fieldMutex.Unlock()
	w.canUseAuthenticatedEndpoints = val
}

// CanUseAuthenticatedEndpoints gets canUseAuthenticatedEndpoints val in
// a thread safe manner
func (w *Websocket) CanUseAuthenticatedEndpoints() bool {
	w.fieldMutex.RLock()
	defer w.fieldMutex.RUnlock()
	return w.canUseAuthenticatedEndpoints
}

// IsDisconnectionError Determines if the error sent over chan ReadMessageErrors is a disconnection error
func IsDisconnectionError(err error) bool {
	if websocket.IsUnexpectedCloseError(err) {
		return true
	}
	if _, ok := err.(*net.OpError); ok {
		return !errors.Is(err, errClosedConnection)
	}
	return false
}

// checkWebsocketURL checks for a valid websocket url
func checkWebsocketURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("cannot set %w %s", errInvalidWebsocketURL, s)
	}
	return nil
}

// checkSubscriptions checks subscriptions against the max subscription limit
// and if the subscription already exists.
func (w *Websocket) checkSubscriptions(subs []subscription.Subscription) error {
	if len(subs) == 0 {
		return errNoSubscriptionsSupplied
	}

	w.subscriptionMutex.RLock()
	defer w.subscriptionMutex.RUnlock()

	if w.MaxSubscriptionsPerConnection > 0 && len(w.subscriptions)+len(subs) > w.MaxSubscriptionsPerConnection {
		return fmt.Errorf("%w: current subscriptions: %v, incoming subscriptions: %v, max subscriptions per connection: %v - please reduce enabled pairs",
			errSubscriptionsExceedsLimit,
			len(w.subscriptions),
			len(subs),
			w.MaxSubscriptionsPerConnection)
	}

	for i := range subs {
		key := subs[i].EnsureKeyed()
		if _, ok := w.subscriptions[key]; ok {
			return fmt.Errorf("%w for %+v", errChannelAlreadySubscribed, subs[i])
		}
	}

	return nil
}
