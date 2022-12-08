package twap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	strategy "github.com/thrasher-corp/gocryptotrader/exchanges/strategy/common"
)

// OnSignal processing signals that have been generated by the strategy.
func (s *Strategy) OnSignal(ctx context.Context, sig interface{}) (bool, error) {
	if s == nil {
		return false, strategy.ErrIsNil
	}

	if sig == nil {
		return false, errors.New("signal is non-existant")
	}

	if _, ok := sig.(time.Time); !ok {
		return false, errors.New("unhandled signal")
	}

	err := s.checkAndSubmit(ctx)
	if err != nil {
		return false, err
	}

	if s.allocation.Deployed == s.allocation.Total {
		return true, nil
	}

	return false, nil
}

// GetDescription returns the strategy description
func (s *Strategy) GetDescription() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("Start:%s End:%s Exchange:%s Pair:%s Asset:%s Interval:%s Window:%s Simulation:%v Amount:%v Deployent:%v",
		s.Start.Format(common.SimpleTimeFormat),
		s.End.Format(common.SimpleTimeFormat),
		s.Config.Exchange.GetName(),
		s.Config.Pair,
		s.Config.Asset,
		s.Config.Interval,
		s.allocation.Window,
		s.Config.Simulate,
		s.allocation.Total,
		s.allocation.Deployment)
}
