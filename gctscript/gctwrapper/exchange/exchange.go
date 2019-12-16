package exchange

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/engine"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/withdraw"
	"github.com/thrasher-corp/gocryptotrader/gctscript/modules"
)

// Exchange implements all required methods for Wrapper
type Exchange struct{}

// Exchanges returns slice of all current exchanges
func (e Exchange) Exchanges(enabledOnly bool) []string {
	return engine.GetExchanges(enabledOnly)
}

// GetExchange returns IBotExchange for exchange or error if exchange is not found
func (e Exchange) GetExchange(exch string) (exchange.IBotExchange, error) {
	ex := engine.GetExchangeByName(exch)
	if ex == nil {
		return nil, fmt.Errorf("%v exchange not found", exch)
	}

	return ex, nil
}

// IsEnabled returns if requested exchange is enabled or disabled
func (e Exchange) IsEnabled(exch string) bool {
	ex, err := e.GetExchange(exch)
	if err != nil {
		return false
	}

	return ex.IsEnabled()
}

// Orderbook returns current orderbook requested exchange, pair and asset
func (e Exchange) Orderbook(exch string, pair currency.Pair, item asset.Item) (*orderbook.Base, error) {
	return engine.GetSpecificOrderbook(pair, exch, item)
}

// Ticker returns ticker for provided currency pair & asset type
func (e Exchange) Ticker(exch string, pair currency.Pair, item asset.Item) (*ticker.Price, error) {
	ex, err := e.GetExchange(exch)
	if err != nil {
		return nil, err
	}

	tx, err := ex.FetchTicker(pair, item)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// Pairs returns either all or enabled currency pairs
func (e Exchange) Pairs(exch string, enabledOnly bool, item asset.Item) (*currency.Pairs, error) {
	x, err := engine.Bot.Config.GetExchangeConfig(exch)
	if err != nil {
		return nil, err
	}

	if enabledOnly {
		return &x.CurrencyPairs.Get(item).Enabled, nil
	}
	return &x.CurrencyPairs.Get(item).Available, nil
}

// QueryOrder returns details of a valid exchange order
func (e Exchange) QueryOrder(exch, orderID string) (*order.Detail, error) {
	ex, err := e.GetExchange(exch)
	if err != nil {
		return nil, err
	}

	r, err := ex.GetOrderInfo(orderID)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// SubmitOrder submit new order on exchange
func (e Exchange) SubmitOrder(exch string, submit *order.Submit) (*order.SubmitResponse, error) {
	r, err := engine.Bot.OrderManager.Submit(exch, submit)
	if err != nil {
		return nil, err
	}

	return &r.SubmitResponse, nil
}

// CancelOrder wrapper to cancel order on exchange
func (e Exchange) CancelOrder(exch, orderID string) (bool, error) {
	orderDetails, err := e.QueryOrder(exch, orderID)
	if err != nil {
		return false, err
	}

	cancel := &order.Cancel{
		AccountID:    orderDetails.AccountID,
		OrderID:      orderDetails.ID,
		CurrencyPair: orderDetails.CurrencyPair,
		Side:         orderDetails.OrderSide,
	}

	err = engine.Bot.OrderManager.Cancel(exch, cancel)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AccountInformation returns account information (balance etc) for requested exchange
func (e Exchange) AccountInformation(exch string) (*modules.AccountInfo, error) {
	ex, err := e.GetExchange(exch)
	if err != nil {
		return nil, err
	}

	r, err := ex.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	temp, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	accountInfo := modules.AccountInfo{}
	err = json.Unmarshal(temp, &accountInfo)
	if err != nil {
		return nil, err
	}

	return &accountInfo, nil
}

// DepositAddress gets the address required to deposit funds for currency type
func (e Exchange) DepositAddress(exch string, currencyCode currency.Code, accountID string) (out string, err error) {
	if currencyCode.IsEmpty() {
		err = errors.New("currency code is empty")
		return
	}
	return engine.Bot.DepositAddressManager.GetDepositAddressByExchange(exch, currencyCode)
}

// WithdrawalFiatFunds withdraw funds from exchange to requested fiat source
func (e Exchange) WithdrawalFiatFunds(exch string, request *withdraw.FiatRequest) (out string, err error) {
	return "", common.ErrNotYetImplemented
}

// WithdrawalCryptoFunds withdraw funds from exchange to requested Crypto source
func (e Exchange) WithdrawalCryptoFunds(exch string,request  *withdraw.CryptoRequest) (out string, err error) {
	return "", common.ErrNotYetImplemented
}
