package wshandler

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var ws *Websocket

func TestWebsocketInit(t *testing.T) {
	ws = New()
	if ws == nil {
		t.Error("test failed - Websocket New() error")
	}
}

func TestWebsocket(t *testing.T) {
	if err := ws.SetProxyAddress("testProxy"); err != nil {
		t.Error("test failed - SetProxyAddress", err)
	}

	ws.Setup(func() error { return nil },
		func(test WebsocketChannelSubscription) error { return nil },
		func(test WebsocketChannelSubscription) error { return nil },
		"testName",
		true,
		false,
		"testDefaultURL",
		"testRunningURL",
		false)

	// Test variable setting and retreival
	if ws.GetName() != "testName" {
		t.Error("test failed - WebsocketSetup")
	}

	if !ws.IsEnabled() {
		t.Error("test failed - WebsocketSetup")
	}

	if ws.GetProxyAddress() != "testProxy" {
		t.Error("test failed - WebsocketSetup")
	}

	if ws.GetDefaultURL() != "testDefaultURL" {
		t.Error("test failed - WebsocketSetup")
	}

	if ws.GetWebsocketURL() != "testRunningURL" {
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
			case <-ws.Connected:
				count++
			case <-ws.Disconnected:
				count++
			}
		}
	}()
	// -- Not connected shutdown
	err := ws.Shutdown()
	if err == nil {
		t.Fatal("test failed - should not be connected to able to shut down")
	}
	ws.Wg.Wait()
	// -- Normal connect
	err = ws.Connect()
	if err != nil {
		t.Fatal("test failed - WebsocketSetup", err)
	}

	// -- Already connected connect
	err = ws.Connect()
	if err == nil {
		t.Fatal("test failed - should not connect, already connected")
	}

	ws.SetWebsocketURL("")

	// -- Set true when already true
	err = ws.SetWsStatusAndConnection(true)
	if err == nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Set false normal
	err = ws.SetWsStatusAndConnection(false)
	if err != nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Set true normal
	err = ws.SetWsStatusAndConnection(true)
	if err != nil {
		t.Fatal("test failed - setting enabled should not work")
	}

	// -- Normal shutdown
	err = ws.Shutdown()
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

// placeholderSubscriber basic function to test subscriptions
func placeholderSubscriber(channelToSubscribe WebsocketChannelSubscription) error {
	return nil
}

// TestSubscribe logic test
func TestSubscribe(t *testing.T) {
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{
			{
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

// TestUnsubscribe logic test
func TestUnsubscribe(t *testing.T) {
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{},
		subscribedChannels: []WebsocketChannelSubscription{
			{
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

// TestSubscriptionWithExistingEntry logic test
func TestSubscriptionWithExistingEntry(t *testing.T) {
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{
			{
				Channel: "hello",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{
			{
				Channel: "hello",
			},
		},
	}
	w.SetChannelSubscriber(placeholderSubscriber)
	w.subscribeToChannels()
	if len(w.subscribedChannels) != 1 {
		t.Errorf("Subscription should not have occurred")
	}
}

// TestUnsubscriptionWithExistingEntry logic test
func TestUnsubscriptionWithExistingEntry(t *testing.T) {
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{
			{
				Channel: "hello",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{
			{
				Channel: "hello",
			},
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.unsubscribeToChannels()
	if len(w.subscribedChannels) != 1 {
		t.Errorf("Unsubscription should not have occurred")
	}
}

// TestManageSubscriptionsStartStop logic test
func TestManageSubscriptionsStartStop(t *testing.T) {
	w := Websocket{
		ShutdownC:     make(chan struct{}, 1),
		Functionality: WebsocketSubscribeSupported | WebsocketUnsubscribeSupported,
	}
	go w.manageSubscriptions()
	time.Sleep(time.Second)
	close(w.ShutdownC)
}

// TestConnectionMonitorNoConnection logic test
func TestConnectionMonitorNoConnection(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.exchangeName = "hello"
	go w.connectionMonitor()
	err := <-w.DataHandler
	if !strings.EqualFold(err.(error).Error(),
		fmt.Sprintf("%v connectionMonitor: websocket disabled, shutting down", w.exchangeName)) {
		t.Errorf("expecting error 'connectionMonitor: websocket disabled, shutting down', received '%v'", err)
	}
}

// TestWsNoConnectionTolerance logic test
func TestWsNoConnectionTolerance(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.noConnectionCheckLimit = 500
	w.checkConnection()
	if w.noConnectionChecks == 0 {
		t.Errorf("Expected noConnectionTolerance to increment, received '%v'", w.noConnectionChecks)
	}
}

// TestConnecting logic test
func TestConnecting(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.connecting = true
	w.reconnectionLimit = 500
	w.checkConnection()
	if w.reconnectionChecks != 1 {
		t.Errorf("Expected reconnectionLimit to increment, received '%v'", w.reconnectionChecks)
	}
}

// TestReconnectionLimit logic test
func TestReconnectionLimit(t *testing.T) {
	w := Websocket{}
	w.DataHandler = make(chan interface{}, 1)
	w.ShutdownC = make(chan struct{}, 1)
	w.enabled = true
	w.connecting = true
	w.reconnectionChecks = 99
	w.reconnectionLimit = 1
	err := w.checkConnection()
	if err == nil {
		t.Error("Expected error")
	}
}

// TestRemoveChannelToSubscribe logic test
func TestRemoveChannelToSubscribe(t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello",
	}
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{
			subscription,
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.removeChannelToSubscribe(subscription)
	if len(w.subscribedChannels) != 0 {
		t.Errorf("Unsubscription did not occur")
	}
}

// TestRemoveChannelToSubscribeWithNoSubscription logic test
func TestRemoveChannelToSubscribeWithNoSubscription(t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello",
	}
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{},
	}
	w.DataHandler = make(chan interface{}, 1)
	w.SetChannelUnsubscriber(placeholderSubscriber)
	go w.removeChannelToSubscribe(subscription)
	err := <-w.DataHandler
	if !strings.Contains(err.(error).Error(), "could not be removed because it was not found") {
		t.Error("Expected not found error")
	}
}

// TestResubscribeToChannel logic test
func TestResubscribeToChannel(t *testing.T) {
	subscription := WebsocketChannelSubscription{
		Channel: "hello",
	}
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{},
	}
	w.DataHandler = make(chan interface{}, 1)
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.SetChannelSubscriber(placeholderSubscriber)
	w.ResubscribeToChannel(subscription)
}

// TestSliceCopyDoesntImpactBoth logic test
func TestSliceCopyDoesntImpactBoth(t *testing.T) {
	w := Websocket{
		channelsToSubscribe: []WebsocketChannelSubscription{
			{
				Channel: "hello1",
			},
			{
				Channel: "hello2",
			},
		},
		subscribedChannels: []WebsocketChannelSubscription{
			{
				Channel: "hello3",
			},
		},
	}
	w.SetChannelUnsubscriber(placeholderSubscriber)
	w.unsubscribeToChannels()
	if len(w.subscribedChannels) != 2 {
		t.Errorf("Unsubscription did not occur")
	}
	w.subscribedChannels[0].Channel = "test"
	if strings.EqualFold(w.subscribedChannels[0].Channel, w.channelsToSubscribe[0].Channel) {
		t.Errorf("Slice has not been copies appropriately")
	}
}

// TestSetCanUseAuthenticatedEndpoints logic test
func TestSetCanUseAuthenticatedEndpoints(t *testing.T) {
	w := Websocket{}
	result := w.CanUseAuthenticatedEndpoints()
	if result {
		t.Error("expected `canUseAuthenticatedEndpoints` to be false")
	}
	w.SetCanUseAuthenticatedEndpoints(true)
	result = w.CanUseAuthenticatedEndpoints()
	if !result {
		t.Error("expected `canUseAuthenticatedEndpoints` to be true")
	}
}
