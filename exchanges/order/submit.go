package order

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/validate"
)

// Validate checks the supplied data and returns whether it's valid
func (s *Submit) Validate(opt ...validate.Checker) error {
	if s == nil {
		return ErrSubmissionIsNil
	}

	if s.Exchange == "" {
		return errExchangeNameUnset
	}

	if s.Pair.IsEmpty() {
		return ErrPairIsEmpty
	}

	if s.AssetType == asset.Empty {
		return ErrAssetNotSet
	}

	if !s.AssetType.IsValid() {
		return fmt.Errorf("'%s' %w", s.AssetType, asset.ErrNotSupported)
	}

	if !IsValidOrderSubmissionSide(s.Side) {
		return fmt.Errorf("%w %v", ErrSideIsInvalid, s.Side)
	}

	if s.Type != Market && s.Type != Limit {
		return ErrTypeIsInvalid
	}

	if s.ImmediateOrCancel && s.FillOrKill {
		return errTimeInForceConflict
	}

	if s.Amount == 0 && s.QuoteAmount == 0 {
		return fmt.Errorf("submit validation error %w, amount and quote amount cannot be zero", ErrAmountIsInvalid)
	}

	if s.Amount < 0 {
		return fmt.Errorf("submit validation error base %w, supplied: %v", ErrAmountIsInvalid, s.Amount)
	}

	if s.QuoteAmount < 0 {
		return fmt.Errorf("submit validation error quote %w, supplied: %v", ErrAmountIsInvalid, s.QuoteAmount)
	}

	if s.Type == Limit && s.Price <= 0 {
		return ErrPriceMustBeSetIfLimitOrder
	}

	for _, o := range opt {
		err := o.Check()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeriveSubmitResponse will construct an order SubmitResponse when a successful
// submission has occurred. NOTE: order status is populated as order.Filled for a
// market order else order.New if an order is accepted as default, date and
// lastupdated fields have been populated as time.Now(). All fields can be
// customized in caller scope if needed.
func (s *Submit) DeriveSubmitResponse(orderID string) (*SubmitResponse, error) {
	if s == nil {
		return nil, errOrderSubmitIsNil
	}

	if orderID == "" {
		return nil, ErrOrderIDNotSet
	}

	status := New
	if s.Type == Market { // NOTE: This will need to be scrutinized.
		status = Filled
	}

	return &SubmitResponse{
		Exchange:  s.Exchange,
		Type:      s.Type,
		Side:      s.Side,
		Pair:      s.Pair,
		AssetType: s.AssetType,

		ImmediateOrCancel: s.ImmediateOrCancel,
		FillOrKill:        s.FillOrKill,
		PostOnly:          s.PostOnly,
		ReduceOnly:        s.ReduceOnly,
		Leverage:          s.Leverage,
		Price:             s.Price,
		Amount:            s.Amount,
		QuoteAmount:       s.QuoteAmount,
		TriggerPrice:      s.TriggerPrice,
		ClientID:          s.ClientID,
		ClientOrderID:     s.ClientOrderID,

		LastUpdated: time.Now(),
		Date:        time.Now(),
		Status:      status,
		OrderID:     orderID,
	}, nil
}

// DeriveDetail will construct an order detail when a successful submission
// has occurred. Has an optional parameter field internal uuid for internal
// management.
func (s *SubmitResponse) DeriveDetail(internal uuid.UUID) (*Detail, error) {
	if s == nil {
		return nil, errOrderSubmitResponseIsNil
	}

	return &Detail{
		Exchange:  s.Exchange,
		Type:      s.Type,
		Side:      s.Side,
		Pair:      s.Pair,
		AssetType: s.AssetType,

		ImmediateOrCancel: s.ImmediateOrCancel,
		FillOrKill:        s.FillOrKill,
		PostOnly:          s.PostOnly,
		ReduceOnly:        s.ReduceOnly,
		Leverage:          s.Leverage,
		Price:             s.Price,
		Amount:            s.Amount,
		QuoteAmount:       s.QuoteAmount,
		TriggerPrice:      s.TriggerPrice,
		ClientID:          s.ClientID,
		ClientOrderID:     s.ClientOrderID,

		InternalOrderID: internal,

		LastUpdated: s.LastUpdated,
		Date:        s.Date,
		Status:      s.Status,
		OrderID:     s.OrderID,
		Trades:      s.Trades,
		Fee:         s.Fee,
		Cost:        s.Cost,
	}, nil
}
