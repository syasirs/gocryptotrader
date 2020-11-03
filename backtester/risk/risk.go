package risk

import (
	"github.com/thrasher-corp/gocryptotrader/backtester/eventhandlers/order"
	portfolio "github.com/thrasher-corp/gocryptotrader/backtester/interfaces"
	"github.com/thrasher-corp/gocryptotrader/backtester/orders"
	"github.com/thrasher-corp/gocryptotrader/backtester/positions"
	"github.com/thrasher-corp/gocryptotrader/currency"
)

// TODO implement risk manager
func (r *Risk) EvaluateOrder(o orders.OrderEvent, _ portfolio.DataEventHandler, _ map[currency.Pair]positions.Positions) (*order.Order, error) {
	retOrder := o.(*order.Order)

	if o.IsLeveraged() {
		// handle risk
	}
	return retOrder, nil
}
