package twap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	strategy "github.com/thrasher-corp/gocryptotrader/exchanges/strategy/common"
)

var (
	errNoBalanceFound     = errors.New("no balance found")
	errExceedsFreeBalance = errors.New("amount exceeds current free balance")
	errInvalidAssetType   = errors.New("non spot trading pairs not currently supported")
	errStrategyIsNil      = errors.New("strategy is nil")
)

// GetTWAP returns a TWAP struct to manage allocation or deallocation of
// position(s).
func New(ctx context.Context, c *Config) (*Strategy, error) {
	err := c.Check(ctx)
	if err != nil {
		return nil, err
	}

	if c.Asset != asset.Spot {
		return nil, errInvalidAssetType
	}

	depth, err := orderbook.GetDepth(c.Exchange.GetName(), c.Pair, c.Asset)
	if err != nil {
		return nil, err
	}

	creds, err := c.Exchange.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	buying, err := account.GetBalance(c.Exchange.GetName(),
		creds.SubAccount, creds, c.Asset, c.Pair.Base)
	if err != nil {
		return nil, err
	}

	deployment := c.Pair.Quote
	selling, err := account.GetBalance(c.Exchange.GetName(),
		creds.SubAccount, creds, c.Asset, c.Pair.Quote)
	if err != nil {
		return nil, err
	}

	if !c.Buy {
		buying, selling = selling, buying
		deployment = c.Pair.Base
	}

	balance := selling.GetAvailableWithoutBorrow()
	if balance == 0 {
		return nil, fmt.Errorf("cannot sell %s amount %f to buy base %s %w of %f",
			deployment,
			c.Amount,
			c.Pair.Base,
			errNoBalanceFound,
			balance)
	}

	if !c.FullAmount {
		if c.Amount > balance {
			return nil, fmt.Errorf("cannot sell %s amount %f to buy base %s %w of %f",
				deployment,
				c.Amount,
				c.Pair.Base,
				errExceedsFreeBalance,
				balance)
		}
		balance = c.Amount
	}

	// NOTE: For now this will not allow any amount to deplete the full
	// orderbook, just until a safe, effective and efficient system has been
	// tested and deployed for public use.
	// TODO: Bypass error errBookSmallerThanDeploymentAmount.
	allocation, err := c.GetDistrbutionAmount(balance, depth)
	if err != nil {
		return nil, err
	}

	return &Strategy{
		Config:     c,
		orderbook:  depth,
		Buying:     buying,
		Selling:    selling,
		allocation: allocation,
	}, nil
}

// deploy oversees the deployment of the current strategy adhering to policies,
// limits, signals and timings. TODO: Abstract to common. Implement OnSignal
// interface requirement method.
func (s *Strategy) deploy(ctx context.Context, start time.Duration) {
	defer func() {
		s.wg.Done()
		s.mtx.Lock()
		s.running = false
		s.mtx.Unlock()
	}()

	s.reporter.OnStart(s)
	// NOTE: Zero value start duration will execute immediately then deploy at
	// intervals.
	timer := time.NewTimer(start)
	finished := time.NewTimer(time.Until(s.End))

	for {
		select {
		case sig := <-timer.C:
			err := s.SetTimer(timer)
			if err != nil {
				s.reporter.OnFatalError(err)
				return
			}

			var complete bool
			complete, err = s.OnSignal(ctx, sig)
			if err != nil {
				s.reporter.OnFatalError(err)
				return
			}

			if complete {
				s.reporter.OnComplete()
				return
			}
		case <-finished.C:
			s.reporter.OnTimeout(s.End)
			return
		case <-ctx.Done():
			s.reporter.OnContextDone(ctx.Err())
			return
		case <-s.shutdown:
			s.reporter.OnShutdown()
			return
		}
	}
}

// checkAndSubmit verifies orderbook deployability then executes an order if
// all checks pass.
func (s *Strategy) checkAndSubmit(ctx context.Context) error {
	if s == nil {
		return errStrategyIsNil
	}

	deploymentInBase, details, err := s.VerifyBookDeployment(s.orderbook, s.allocation.Deployment)
	if err != nil {
		return err
	}

	conformed, err := s.VerifyExecutionLimitsReturnConformed(deploymentInBase)
	if err != nil {
		return err
	}

	submit, err := s.deriveOrder(conformed)
	if err != nil {
		return err
	}

	resp, err := s.submitOrder(ctx, submit)
	if err != nil {
		return err
	}

	// Note: For first iteration of strategy this is just easy reconciliation.
	// TODO: Reconcile to adjusted amount.
	s.allocation.Deployed += s.allocation.Deployment
	s.allocation.Deployments++

	s.reporter.OnOrder(submit, resp, details)
	return nil
}

// deriveOrder checks amount and returns an order submission. TODO: Abstract
// futher.
func (s *Strategy) deriveOrder(amountInBase float64) (*order.Submit, error) {
	if amountInBase <= 0 {
		return nil, errInvalidAllocatedAmount
	}
	side := order.Buy
	if !s.Buy {
		side = order.Sell
	}
	return &order.Submit{
		Exchange:  s.Config.Exchange.GetName(),
		Type:      order.Market,
		Pair:      s.Config.Pair,
		AssetType: s.Config.Asset,
		Side:      side,
		Amount:    amountInBase,
	}, nil
}

// submitOrder will submit and retry an order if fail. TODO: Abstract futher
func (s *Strategy) submitOrder(ctx context.Context, submit *order.Submit) (*order.SubmitResponse, error) {
	if submit == nil {
		return nil, errors.New("submit order is invalid")
	}
	var errors common.Errors
	var resp *order.SubmitResponse
	for attempt := 0; attempt < int(s.RetryAttempts); attempt++ {
		// Check context here so we can immediately bypass the retry attempt and
		// release resources.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var err error
		if !s.Simulate {
			resp, err = s.Config.Exchange.SubmitOrder(ctx, submit)
		} else {
			resp, err = submit.DeriveSubmitResponse(strategy.SimulationTag)
		}
		if err == nil {
			errors = nil // These errors prior we don't need to worry about.
			break
		}
		errors = append(errors, err)
		time.Sleep(time.Second)
	}
	var errReturn error
	if errors != nil {
		errReturn = errors
	}
	return resp, errReturn
}
