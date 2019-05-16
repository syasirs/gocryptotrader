package exchange

import (
	"testing"
	"time"
	"strings"

	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
)

var wsTest Base

func TestWebsocketInit(t *testing.T) {
	if wsTest.Websocket != nil {
		t.Error("test failed - WebsocketInit() error")
	}

	wsTest.WebsocketInit()

	if wsTest.Websocket == nil {
		t.Error("test failed - WebsocketInit() error")
	}
}

func TestWebsocket(t *testing.T) {
	if err := wsTest.Websocket.SetProxyAddress("testProxy"); err != nil {
		t.Error("test failed - SetProxyAddress", err)
	}

	wsTest.WebsocketSetup(func() error { return nil },
		func(test WebsocketChannelSubscription) error { return nil },
		func(test WebsocketChannelSubscription) error { return nil },
		"testName",
		true,
		false,
		"testDefaultURL",
		"testRunningURL")

	// Test variable setting and retreival
	if wsTest.Websocket.GetName() != "testName" {
		t.Error("test failed - WebsocketSetup")
	}

	if !wsTest.Websocket.IsEnabled() {
		t.Error("test failed - WebsocketSetup")
	}

	if wsTest.Websocket.GetProxyAddress() != "testProxy" {
		t.Error("test failed - WebsocketSetup")
	}

	if wsTest.Websocket.GetDefaultURL() != "testDefaultURL" {
		t.Error("test failed - WebsocketSetup")
	}

	if wsTest.Websocket.GetWebsocketURL() != "testRunningURL" {
		t.Error("test failed - WebsocketSetup")
	}

	// Test websocket connect and shutdown functions
	comms := make(chan struct{}, 1)
	go func() {
		var count int
		for {
			if count == 4 {
				close(comms)
				return
			}
			select {
			case <-wsTest.Websocket.Connected:
				count++
			case <-wsTest.Websocket.Disconnected:
				count++
			}
		}
	}()
	// -- Not connected shutdown
	err := wsTest.Websocket.Shutdown()
	if err == nil {
		t.Fatal("test failed - should not be connected to able to shut down")
	}
	time.Sleep(800 * time.Millisecond)
	// -- Normal connect
	err = wsTest.Websocket.Connect()
	if err != nil {
		t.Fatal("test failed - WebsocketSetup", err)
	}

	// -- Already connected connect
	err = wsTest.Websocket.Connect()
	if err == nil {
		t.Fatal("test failed - should not connect, already connected")
	}

	wsTest.Websocket.SetWebsocketURL("")

	// -- Set true when already true
	err = wsTest.Websocket.SetWsStatusAndConnection(true)
	if err == nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Set false normal
	err = wsTest.Websocket.SetWsStatusAndConnection(false)
	if err != nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Set true normal
	err = wsTest.Websocket.SetWsStatusAndConnection(true)
	if err != nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Normal shutdown
	err = wsTest.Websocket.Shutdown()
	if err != nil {
		t.Fatal("test failed - WebsocketSetup", err)
	}

	timer := time.NewTimer(5 * time.Second)
	select {
	case <-comms:
	case <-timer.C:
		t.Fatal("test failed - WebsocketSetup - timeout")
	}
}

func TestInsertingSnapShots(t *testing.T) {
	var snapShot1 orderbook.Base
	asks := []orderbook.Item{
		{Price: 6000, Amount: 1, ID: 1},
		{Price: 6001, Amount: 0.5, ID: 2},
		{Price: 6002, Amount: 2, ID: 3},
		{Price: 6003, Amount: 3, ID: 4},
		{Price: 6004, Amount: 5, ID: 5},
		{Price: 6005, Amount: 2, ID: 6},
		{Price: 6006, Amount: 1.5, ID: 7},
		{Price: 6007, Amount: 0.5, ID: 8},
		{Price: 6008, Amount: 23, ID: 9},
		{Price: 6009, Amount: 9, ID: 10},
		{Price: 6010, Amount: 7, ID: 11},
	}

	bids := []orderbook.Item{
		{Price: 5999, Amount: 1, ID: 12},
		{Price: 5998, Amount: 0.5, ID: 13},
		{Price: 5997, Amount: 2, ID: 14},
		{Price: 5996, Amount: 3, ID: 15},
		{Price: 5995, Amount: 5, ID: 16},
		{Price: 5994, Amount: 2, ID: 17},
		{Price: 5993, Amount: 1.5, ID: 18},
		{Price: 5992, Amount: 0.5, ID: 19},
		{Price: 5991, Amount: 23, ID: 20},
		{Price: 5990, Amount: 9, ID: 21},
		{Price: 5989, Amount: 7, ID: 22},
	}

	snapShot1.Asks = asks
	snapShot1.Bids = bids
	snapShot1.AssetType = "SPOT"
	snapShot1.Pair = currency.NewPairFromString("BTCUSD")

	wsTest.Websocket.Orderbook.LoadSnapshot(&snapShot1, "ExchangeTest", false)

	var snapShot2 orderbook.Base
	asks = []orderbook.Item{
		{Price: 51, Amount: 1, ID: 1},
		{Price: 52, Amount: 0.5, ID: 2},
		{Price: 53, Amount: 2, ID: 3},
		{Price: 54, Amount: 3, ID: 4},
		{Price: 55, Amount: 5, ID: 5},
		{Price: 56, Amount: 2, ID: 6},
		{Price: 57, Amount: 1.5, ID: 7},
		{Price: 58, Amount: 0.5, ID: 8},
		{Price: 59, Amount: 23, ID: 9},
		{Price: 50, Amount: 9, ID: 10},
		{Price: 60, Amount: 7, ID: 11},
	}

	bids = []orderbook.Item{
		{Price: 49, Amount: 1, ID: 12},
		{Price: 48, Amount: 0.5, ID: 13},
		{Price: 47, Amount: 2, ID: 14},
		{Price: 46, Amount: 3, ID: 15},
		{Price: 45, Amount: 5, ID: 16},
		{Price: 44, Amount: 2, ID: 17},
		{Price: 43, Amount: 1.5, ID: 18},
		{Price: 42, Amount: 0.5, ID: 19},
		{Price: 41, Amount: 23, ID: 20},
		{Price: 40, Amount: 9, ID: 21},
		{Price: 39, Amount: 7, ID: 22},
	}

	snapShot2.Asks = asks
	snapShot2.Bids = bids
	snapShot2.AssetType = "SPOT"
	snapShot2.Pair = currency.NewPairFromString("LTCUSD")

	wsTest.Websocket.Orderbook.LoadSnapshot(&snapShot2, "ExchangeTest", false)

	var snapShot3 orderbook.Base
	asks = []orderbook.Item{
		{Price: 51, Amount: 1, ID: 1},
		{Price: 52, Amount: 0.5, ID: 2},
		{Price: 53, Amount: 2, ID: 3},
		{Price: 54, Amount: 3, ID: 4},
		{Price: 55, Amount: 5, ID: 5},
		{Price: 56, Amount: 2, ID: 6},
		{Price: 57, Amount: 1.5, ID: 7},
		{Price: 58, Amount: 0.5, ID: 8},
		{Price: 59, Amount: 23, ID: 9},
		{Price: 50, Amount: 9, ID: 10},
		{Price: 60, Amount: 7, ID: 11},
	}

	bids = []orderbook.Item{
		{Price: 49, Amount: 1, ID: 12},
		{Price: 48, Amount: 0.5, ID: 13},
		{Price: 47, Amount: 2, ID: 14},
		{Price: 46, Amount: 3, ID: 15},
		{Price: 45, Amount: 5, ID: 16},
		{Price: 44, Amount: 2, ID: 17},
		{Price: 43, Amount: 1.5, ID: 18},
		{Price: 42, Amount: 0.5, ID: 19},
		{Price: 41, Amount: 23, ID: 20},
		{Price: 40, Amount: 9, ID: 21},
		{Price: 39, Amount: 7, ID: 22},
	}

	snapShot3.Asks = asks
	snapShot3.Bids = bids
	snapShot3.AssetType = "FUTURES"
	snapShot3.Pair = currency.NewPairFromString("LTCUSD")

	wsTest.Websocket.Orderbook.LoadSnapshot(&snapShot3, "ExchangeTest", false)

	if len(wsTest.Websocket.Orderbook.ob) != 3 {
		t.Error("test failed - inserting orderbook data")
	}
}

func TestUpdate(t *testing.T) {
	LTCUSDPAIR := currency.NewPairFromString("LTCUSD")
	BTCUSDPAIR := currency.NewPairFromString("BTCUSD")

	bidTargets := []orderbook.Item{
		{Price: 49, Amount: 24},    // Amend
		{Price: 48, Amount: 0},     // Delete
		{Price: 1337, Amount: 100}, // Append
		{Price: 1336, Amount: 0},   // Ghost delete
	}

	askTargets := []orderbook.Item{
		{Price: 51, Amount: 24},    // Amend
		{Price: 52, Amount: 0},     // Delete
		{Price: 1337, Amount: 100}, // Append
		{Price: 1336, Amount: 0},   // Ghost delete
	}
	err := wsTest.Websocket.Orderbook.Update(bidTargets,
		askTargets,
		LTCUSDPAIR,
		time.Now(),
		"ExchangeTest",
		"SPOT")

	if err != nil {
		t.Error("test failed - OrderbookUpdate error", err)
	}

	err = wsTest.Websocket.Orderbook.Update(bidTargets,
		askTargets,
		LTCUSDPAIR,
		time.Now(),
		"ExchangeTest",
		"FUTURES")

	if err != nil {
		t.Error("test failed - OrderbookUpdate error", err)
	}

	bidTargets = []orderbook.Item{
		{Price: 5999, Amount: 24},  // Amend
		{Price: 5998, Amount: 0},   // Delete
		{Price: 1337, Amount: 100}, // Append
		{Price: 1336, Amount: 0},   // Ghost delete
	}

	askTargets = []orderbook.Item{
		{Price: 6000, Amount: 24},  // Amend
		{Price: 6001, Amount: 0},   // Delete
		{Price: 1337, Amount: 100}, // Append
		{Price: 1336, Amount: 0},   // Ghost delete
	}

	err = wsTest.Websocket.Orderbook.Update(bidTargets,
		askTargets,
		BTCUSDPAIR,
		time.Now(),
		"ExchangeTest",
		"SPOT")

	if err != nil {
		t.Error("test failed - OrderbookUpdate error", err)
	}
}

func TestFunctionality(t *testing.T) {
	var w Websocket

	if w.FormatFunctionality() != NoWebsocketSupportText {
		t.Fatalf("Test Failed - FormatFunctionality error expected %s but received %s",
			NoWebsocketSupportText, w.FormatFunctionality())
	}

	w.Functionality = 1 << 31

	if w.FormatFunctionality() != UnknownWebsocketFunctionality+"[1<<31]" {
		t.Fatal("Test Failed - GetFunctionality error incorrect error returned")
	}

	w.Functionality = WebsocketOrderbookSupported

	if w.GetFunctionality() != WebsocketOrderbookSupported {
		t.Fatal("Test Failed - GetFunctionality error incorrect bitmask returned")
	}

	if !w.SupportsFunctionality(WebsocketOrderbookSupported) {
		t.Fatal("Test Failed - SupportsFunctionality error should be true")
	}
}


func placeholderSubscriber (channelToSubscribe WebsocketChannelSubscription) error {
	return nil
}
func TestSubscribe(t *testing.T) {
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{},
	}
	w.SetChannelSubscriber(placeholderSubscriber)
	w.subscribeToChannels()
	if len(w.subscribedChannels) != 1 {
		t.Errorf("Subscription did not occur")
	}
}

func TestUnsubscribe(t *testing.T) {
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{},
		subscribedChannels: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.unsubscribeToChannels()
	if len(w.subscribedChannels) != 0 {
		t.Errorf("Unsubscription did not occur")
	}
}

func TestSubscriptionWithExistingEntry(t *testing.T) {
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
	}
	w.SetChannelSubscriber(placeholderSubscriber)
	w.subscribeToChannels()
	if len(w.subscribedChannels) != 1 {
		t.Errorf("Subscription should not have occured")
	}
}

func TestUnsubscriptionWithExistingEntry(t *testing.T) {
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{
			WebsocketChannelSubscription{
				Channel: "hello",
			},
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.unsubscribeToChannels()
	if len(w.subscribedChannels) != 1 {
		t.Errorf("Unsubscription should not have occured")
	}
}

func TestManageSubscriptionsWithoutFunctionality(t *testing.T) {
	w := Websocket{
		ShutdownC: make(chan struct{}, 1),
	}

	err := w.manageSubscriptions()
	if err == nil {
		t.Error("Requires functionality to work")
	}
}

func TestManageSubscriptionsStartStop(t *testing.T) {
	w := Websocket{
		ShutdownC: make(chan struct{}, 1),
		Functionality: WebsocketSubscribeSupported | WebsocketUnsubscribeSupported,
	}

	go w.manageSubscriptions()
	time.Sleep(time.Second)
	close(w.ShutdownC)
}


func TestWsConnectionMonitorNoConnection(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	go w.wsConnectionMonitor()
	err := <- w.DataHandler
	if !strings.EqualFold(err.(error).Error(), "WsConnectionMonitor: websocket disabled, shutting down") {
		t.Error("expecting error 'WsConnectionMonitor: websocket disabled, shutting down'")
	}
}


func TestWsNoConnectionTolerance(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.checkConnection()
	if w.noConnectionChecks == 0 {
		t.Error("Expected noConnectionTolerance to increment")
	}
}

func TestConnecting(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.Connecting = true
	w.checkConnection()
	if w.reconnectionChecks != 1 {
		t.Error("Expected reconnectionLimit to increment")
	}
}

func TestReconnectionLimit(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.Connecting = true
	w.reconnectionLimit = 99
	err := w.checkConnection()
	if err == nil {
		t.Error("Expected error")
	}
}

func TestRemoveChannelToSubscribe (t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello", 
	}
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{
			subscription,
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.RemoveChannelToSubscribe(subscription)
	if len(w.subscribedChannels) != 0 {
		t.Errorf("Unsubscription did not occur")
	} 
}

func TestRemoveChannelToSubscribeWithNoSubscription (t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello", 
	}
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{},
	}
	w.DataHandler = make(chan interface{}, 1)
	w.SetChannelUnsubscriber(placeholderSubscriber)
	go w.RemoveChannelToSubscribe(subscription)
	err := <- w.DataHandler
	if !strings.Contains(err.(error).Error(), "could not be removed because it was not found") {
		t.Error("Expected not found error")
	}
}

func TestResubscribeToChannel(t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello", 
	}
	w := Websocket{
		ChannelsToSubscribe: []WebsocketChannelSubscription{},
	}
	w.DataHandler = make(chan interface{}, 1)
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.SetChannelSubscriber(placeholderSubscriber)
	w.ResubscribeToChannel(subscription)
}