package twap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	strategy "github.com/thrasher-corp/gocryptotrader/exchanges/strategy/common"
)

// OnSignal processing signals that have been generated by the strategy.
func (s *Strategy) OnSignal(ctx context.Context, sig interface{}) (bool, error) {
	if s == nil {
		return false, strategy.ErrIsNil
	}

	if sig == nil {
		return false, errors.New("signal is non-existent")
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
func (s *Strategy) GetDescription() strategy.Descriptor {
	if s == nil {
		return nil
	}

	selling := s.Pair.Base
	if s.Buy {
		selling = s.Pair.Quote
	}

	sched := s.Scheduler.GetSchedule()
	untilStart := "immediately"
	if until := time.Until(sched.Next); until > 0 {
		untilStart = until.String()
	}

	sinceStart := "not yet started"
	if since := time.Since(sched.Start); since > 0 {
		sinceStart = since.String()
	}

	return &Description{
		Exchange:           s.Exchange.GetName(),
		Pair:               s.Pair,
		Asset:              s.Asset,
		Start:              sched.Start.UTC().Format(common.SimpleTimeFormat),
		End:                sched.End.UTC().Format(common.SimpleTimeFormat),
		UntilStart:         untilStart,
		SinceStart:         sinceStart,
		Aligned:            s.CandleStickAligned,
		DeploymentInterval: s.Interval,
		OperatingWindow:    sched.Window.String(),
		Overtrade:          s.AllowTradingPastEndTime,
		Simulation:         s.Simulate,
		Total:              Deployment{Amount: s.allocation.Total, Currency: selling},
		Individual:         Deployment{Amount: s.allocation.Deployment, Currency: selling},
		TWAPInterval:       s.TWAP,
		TWAPPeriod:         30, // TODO: Make this configurable.
	}
}

// Deployment defines an amount and its corresponding currency code.
type Deployment struct {
	Amount   float64       `json:"amount"`
	Currency currency.Code `json:"currency"`
}

// Description defines the full operating description of the strategy with its
// configuration parameters.
type Description struct {
	Exchange           string         `json:"exchange"`
	Pair               currency.Pair  `json:"pair"`
	Asset              asset.Item     `json:"asset"`
	Start              string         `json:"start"`
	End                string         `json:"end"`
	UntilStart         string         `json:"untilStart"`
	SinceStart         string         `json:"sinceStart"`
	Aligned            bool           `json:"aligned"`
	DeploymentInterval kline.Interval `json:"deploymentInterval"`
	OperatingWindow    string         `json:"operatingWindow"`
	Overtrade          bool           `json:"overtrade"`
	Simulation         bool           `json:"simulation"`
	Total              Deployment     `json:"total"`
	Individual         Deployment     `json:"individual"`
	TWAPInterval       kline.Interval `json:"twapInterval"`
	TWAPPeriod         int            `json:"twapPeriod"`
}

// String implements stringer interface for a short description
func (d *Description) String() string {
	if d == nil {
		return ""
	}

	sim := "[STRATEGY IS LIVE]"
	if d.Simulation {
		sim = "[STRATEGY IS IN SIMULATION]"
	}

	return fmt.Sprintf("Exchange:%s Pair:%s Asset:%s Interval:%s Total:%v[%s] individual:%v[%s] StartingIn:%s TWAP:%s TWAP-PERIOD:%d %s",
		d.Exchange,
		d.Pair,
		d.Asset,
		d.DeploymentInterval,
		d.Total.Amount,
		d.Total.Currency,
		d.Individual.Amount,
		d.Individual.Currency,
		d.UntilStart,
		d.TWAPInterval,
		d.TWAPPeriod,
		sim)
}
