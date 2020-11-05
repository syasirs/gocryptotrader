package stream

import (
	"errors"
	"fmt"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

// NewSubscriptionManager sets up a new subscription manager for use in the
// connection manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{m: make(map[Subscription]*[]ChannelSubscription)}
}

// AddSuccessfulSubscriptions adds a successful subscription
func (s *SubscriptionManager) AddSuccessfulSubscriptions(subscriptions []ChannelSubscription) error {
	if subscriptions == nil {
		return errors.New("subscriptions cannot be nil")
	}

	s.Lock()
	defer s.Unlock()

	for x := range subscriptions {
		ptr, ok := s.m[subscriptions[x].SubscriptionType]
		if !ok {
			cont := []ChannelSubscription{subscriptions[x]}
			s.m[subscriptions[x].SubscriptionType] = &cont
			continue
		}
		*ptr = append(*ptr, subscriptions[x])
	}
	return nil
}

// RemoveSuccessfulUnsubscriptions removes a subscription that was successfully
// unsubscribed
func (s *SubscriptionManager) RemoveSuccessfulUnsubscriptions(subscriptions []ChannelSubscription) error {
	s.Lock()
	defer s.Unlock()

removals:
	for x := range subscriptions {
		slice, ok := s.m[subscriptions[x].SubscriptionType]
		if !ok {
			return errors.New("cannot remove subscription type not found to be associated with connection")
		}

		for y := range *slice {
			if subscriptions[x].Channel == (*slice)[y].Channel {
				(*slice)[y] = (*slice)[len(*slice)-1]
				(*slice)[len((*slice))-1] = ChannelSubscription{}
				(*slice) = (*slice)[:len((*slice))-1]
				continue removals
			}
		}
		return errors.New("cannot remove subscription, not found in subscribed list")
	}
	return nil
}

// GetAllSubscriptions returns current subscriptions for our streaming
// connections
func (s *SubscriptionManager) GetAllSubscriptions() []ChannelSubscription {
	s.Lock()
	defer s.Unlock()

	var subscriptions []ChannelSubscription
	for _, subs := range s.m {
		subscriptions = append(subscriptions, *subs...)
	}
	return subscriptions
}

// GetAssetsBySubscriptionType returns assets associated with the same channel
// subscription type. This is used for when margin and spot which collectively
// are the same thing but have different functionality segregated by individual
// connection
func (s *SubscriptionManager) GetAssetsBySubscriptionType(t Subscription, pair currency.Pair) (asset.Items, error) {
	s.Lock()
	defer s.Unlock()

	subscriptions, ok := s.m[t]
	if !ok {
		return nil,
			fmt.Errorf("subscription type %v not found in individual connection subscription list",
				t)
	}

	var assets asset.Items
	for i := range *subscriptions {
		if !(*subscriptions)[i].Currency.Equal(pair) {
			continue
		}
		if assets.Contains((*subscriptions)[i].Asset) {
			continue
		}
		assets = append(assets, (*subscriptions)[i].Asset)
	}

	if len(assets) == 0 {
		return nil,
			fmt.Errorf("no asset associations found for subscription type %v and pair %s",
				t, pair)
	}

	return assets, nil
}

// FlushSubscriptions removes all subscriptions associated with the connection
func (s *SubscriptionManager) FlushSubscriptions() {
	s.Lock()
	for _, p := range s.m {
		// TODO: Check leakage
		*p = []ChannelSubscription{}
	}
	s.Unlock()
}
