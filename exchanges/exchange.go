package exchange

import (
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

const (
	warningBase64DecryptSecretKeyFailed = "WARNING -- Exchange %s unable to base64 decode secret key.. Disabling Authenticated API support."
	// ErrExchangeNotFound is a constant for an error message
	ErrExchangeNotFound = "Exchange not found in dataset."
)

// AccountInfo is a Generic type to hold each exchange's holdings in
// all enabled currencies
type AccountInfo struct {
	ExchangeName string
	Currencies   []AccountCurrencyInfo
}

// AccountCurrencyInfo is a sub type to store currency name and value
type AccountCurrencyInfo struct {
	CurrencyName string
	TotalValue   float64
	Hold         float64
}

// Base stores the individual exchange information
type Base struct {
	Name                        string
	Enabled                     bool
	Verbose                     bool
	Websocket                   bool
	RESTPollingDelay            time.Duration
	AuthenticatedAPISupport     bool
	APISecret, APIKey, ClientID string
	TakerFee, MakerFee, Fee     float64
	BaseCurrencies              []string
	AvailablePairs              []string
	EnabledPairs                []string
	WebsocketURL                string
	APIUrl                      string
}

// IBotExchange enforces standard functions for all exchanges supported in
// GoCryptoTrader
type IBotExchange interface {
	Setup(exch config.ExchangeConfig)
	Start()
	SetDefaults()
	GetName() string
	IsEnabled() bool
	GetTickerPrice(currency pair.CurrencyPair) (ticker.TickerPrice, error)
	GetOrderbookEx(currency pair.CurrencyPair) (orderbook.OrderbookBase, error)
	GetEnabledCurrencies() []string
	GetExchangeAccountInfo() (AccountInfo, error)
}

// GetName is a method that returns the name of the exchange base
func (e *Base) GetName() string {
	return e.Name
}

// GetEnabledCurrencies is a method that returns the enabled currency pairs of
// the exchange base
func (e *Base) GetEnabledCurrencies() []string {
	return e.EnabledPairs
}

// SetEnabled is a method that sets if the exchange is enabled
func (e *Base) SetEnabled(enabled bool) {
	e.Enabled = enabled
}

// IsEnabled is a method that returns if the current exchange is enabled
func (e *Base) IsEnabled() bool {
	return e.Enabled
}

// SetAPIKeys is a method that sets the current API keys for the exchange
func (e *Base) SetAPIKeys(APIKey, APISecret, ClientID string, b64Decode bool) {
	e.APIKey = APIKey
	e.ClientID = ClientID

	if b64Decode {
		result, err := common.Base64Decode(APISecret)
		if err != nil {
			e.AuthenticatedAPISupport = false
			log.Printf(warningBase64DecryptSecretKeyFailed, e.Name)
		}
		e.APISecret = string(result)
	} else {
		e.APISecret = APISecret
	}
}

// UpdateAvailableCurrencies is a method that sets new pairs to the current
// exchange
func (e *Base) UpdateAvailableCurrencies(exchangeProducts []string) error {
	exchangeProducts = common.SplitStrings(common.StringToUpper(common.JoinStrings(exchangeProducts, ",")), ",")
	diff := common.StringSliceDifference(e.AvailablePairs, exchangeProducts)
	if len(diff) > 0 {
		cfg := config.GetConfig()
		exch, err := cfg.GetExchangeConfig(e.Name)
		if err != nil {
			return err
		}
		log.Printf("%s Updating available pairs. Difference: %s.\n", e.Name, diff)
		exch.AvailablePairs = common.JoinStrings(exchangeProducts, ",")
		cfg.UpdateExchangeConfig(exch)
	}
	return nil
}
