package base

import (
	"github.com/thrasher-corp/gocryptotrader/backtester/data"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/event"
	"github.com/thrasher-corp/gocryptotrader/backtester/eventtypes/signal"
)

type Strategy struct {
	Why string
}

func (s *Strategy) GetBase(d data.Handler) signal.Signal {
	return signal.Signal{
		Event: event.Event{
			Exchange:     d.Latest().GetExchange(),
			Time:         d.Latest().GetTime(),
			CurrencyPair: d.Latest().Pair(),
			AssetType:    d.Latest().GetAssetType(),
			Interval:     d.Latest().GetInterval(),
		},
	}
}
