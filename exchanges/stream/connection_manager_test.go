package stream

import (
	"errors"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
)

var (
	errGenerateSubs  = errors.New("failed to generate subs")
	errGenerateConns = errors.New("failed to generate connections")

	passConnectionFunc        = func(c Connection) error { return nil }
	failConnectionFunc        = func(c Connection) error { return errors.New("fail") }
	passGenerateConnection    = func(_ string, _ bool) (Connection, error) { return &WebsocketConnection{}, nil }
	failGenerateConnection    = func(_ string, _ bool) (Connection, error) { return nil, errGenerateConns }
	passGenerateSubscriptions = func(_ SubscriptionOptions) ([]ChannelSubscription, error) {
		return []ChannelSubscription{
			{
				Channel: "Test Subscription",
			},
		}, nil
	}
	passGenerateSubscriptionsTwoSubs = func(_ SubscriptionOptions) ([]ChannelSubscription, error) {
		return []ChannelSubscription{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}}, nil
	}
	passSubscription          = func(_ SubscriptionParameters) error { return nil }
	failSubscription          = func(_ SubscriptionParameters) error { return errors.New("failed") }
	failGenerateSubscriptions = func(_ SubscriptionOptions) ([]ChannelSubscription, error) { return nil, errGenerateSubs }
	blankConfig               = []ConnectionSetup{{}}
	OneConfigNoMax            = []ConnectionSetup{{URL: "TEST URL"}}
	OneConfigMaxSub           = []ConnectionSetup{{URL: "TEST URL", MaxSubscriptions: 3}}
)

func TestNewConnectionManager(t *testing.T) {
	tests := []struct {
		Name            string
		ConnectionSetup *ConnectionManagerConfig
		Error           error
	}{
		{
			Name:            "Nil Connection Manager Config",
			ConnectionSetup: nil,
			Error:           errNoMainConfiguration,
		},
		{
			Name:            "No Connection Function",
			ConnectionSetup: &ConnectionManagerConfig{},
			Error:           errNoExchangeConnectionFunction,
		},
		{
			Name: "No Generate Subscription Function",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector: passConnectionFunc,
			},
			Error: errNoGenerateSubsFunc,
		},
		{
			Name: "No Subscriber Function",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
			},
			Error: errNoSubscribeFunction,
		},
		{
			Name: "No Unsubscriber Function",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
			},
			Error: errNoUnsubscribeFunction,
		},
		{
			Name: "No Generate Connection Function",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
			},
			Error: errNoGenerateConnFunc,
		},
		{
			Name: "No Feature Set",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
				ExchangeGenerateConnection:    passGenerateConnection,
			},
			Error: errNoFeatures,
		},
		{
			Name: "No General Configurations Set",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
				ExchangeGenerateConnection:    passGenerateConnection,
				Features:                      &protocol.Features{},
			},
			Error: errNoConfigurations,
		},
		{
			Name: "Invalid URL in Configurations Set",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
				ExchangeGenerateConnection:    passGenerateConnection,
				Features:                      &protocol.Features{},
				Configurations:                []ConnectionSetup{{}},
			},
			Error: errMissingURLInConfig,
		},
		{
			Name: "All set",
			ConnectionSetup: &ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: passGenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
				ExchangeGenerateConnection:    passGenerateConnection,
				Features:                      &protocol.Features{},
				Configurations: []ConnectionSetup{{
					URL: "test",
				}},
			},
			Error: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			man, err := NewConnectionManager(tt.ConnectionSetup)
			if err != nil {
				if !errors.Is(err, tt.Error) {
					t.Fatalf("expecting error [%v] but received [%v]", tt.Error, err)
				}
				return
			}
			if man == nil {
				t.Fatal("manager is nil")
			}
		})
	}
}

func TestGenerateConnections(t *testing.T) {
	tests := []struct {
		Name                  string
		GenerateSubscriptions func(SubscriptionOptions) ([]ChannelSubscription, error)
		GenerateConnection    func(url string, auth bool) (Connection, error)
		ConfigurationSet      []ConnectionSetup
		ConnectionCount       int
		SubscriptionCount     int
		Error                 error
	}{
		{
			Name:                  "Dodgy Generate Subscriptions",
			GenerateSubscriptions: failGenerateSubscriptions,
			GenerateConnection:    passGenerateConnection,
			ConfigurationSet:      OneConfigNoMax,
			Error:                 errGenerateSubs,
		},
		{
			Name:                  "Dodgy Generate Connections",
			GenerateSubscriptions: passGenerateSubscriptions,
			GenerateConnection:    failGenerateConnection,
			ConfigurationSet:      OneConfigNoMax,
			Error:                 errGenerateConns,
		},
		{
			Name:                  "Dodgy Generate Connections Subscriptions Exceed Max",
			GenerateSubscriptions: passGenerateSubscriptionsTwoSubs,
			GenerateConnection:    failGenerateConnection,
			ConfigurationSet:      OneConfigMaxSub,
			Error:                 errGenerateConns,
		},
		{
			Name:                  "Generate Connection based on subscriptions with no max",
			GenerateSubscriptions: passGenerateSubscriptionsTwoSubs,
			GenerateConnection:    passGenerateConnection,
			ConfigurationSet:      OneConfigNoMax,
			ConnectionCount:       1,
			SubscriptionCount:     10,
			Error:                 nil,
		},
		{
			Name:                  "Generate Connection based on subscriptions with max",
			GenerateSubscriptions: passGenerateSubscriptionsTwoSubs,
			GenerateConnection:    passGenerateConnection,
			ConfigurationSet:      OneConfigMaxSub,
			ConnectionCount:       4,
			SubscriptionCount:     10,
			Error:                 nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			man, err := NewConnectionManager(&ConnectionManagerConfig{
				ExchangeConnector:             passConnectionFunc,
				ExchangeGenerateSubscriptions: tt.GenerateSubscriptions,
				ExchangeSubscriber:            passSubscription,
				ExchangeUnsubscriber:          passSubscription,
				ExchangeGenerateConnection:    tt.GenerateConnection,
				Configurations:                tt.ConfigurationSet,
				Features:                      &protocol.Features{},
			})
			if err != nil {
				if !errors.Is(err, tt.Error) {
					t.Fatalf("expecting error [%v] but received [%v]", tt.Error, err)
				}
				return
			}
			subs, err := man.GenerateSubscriptions()
			if err != nil {
				if !errors.Is(err, tt.Error) {
					t.Fatalf("expecting error [%v] but received [%v]", tt.Error, err)
				}
				return
			}

			m, err := man.GenerateConnections(subs)
			if err != nil {
				if !errors.Is(err, tt.Error) {
					t.Fatalf("expecting error [%v] but received [%v]", tt.Error, err)
				}
				return
			}

			if !errors.Is(err, tt.Error) {
				if !errors.Is(err, tt.Error) {
					t.Fatalf("expecting error [%v] but received [%v]", tt.Error, err)
				}
				return
			}

			if len(m) != tt.ConnectionCount {
				t.Fatalf("expecting [%d] connections but received [%d]", tt.ConnectionCount, len(m))
			}

			var subCount int
			for _, v := range m {
				subCount += len(v)
			}

			if subCount != tt.SubscriptionCount {
				t.Fatalf("expecting [%d] subscriptions but received [%d]", tt.SubscriptionCount, subCount)
			}
		})
	}
}
