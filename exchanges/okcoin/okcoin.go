package okcoin

import (
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/okgroup"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

const (
	okCoinAuthRate     = 0
	okCoinUnauthRate   = 0
	okCoinAPIPath      = "api/"
	okCoinAPIURL       = "https://www.okcoin.com/" + okCoinAPIPath
	okCoinAPIVersion   = "/v3/"
	okCoinExchangeName = "OKCOIN International"
	okCoinWebsocketURL = "wss://real.okcoin.com:10440/websocket/okcoinapi"
)

// OKCoin bases all methods off okgroup implementation
type OKCoin struct {
	okgroup.OKGroup
}

// SetDefaults method assignes the default values for OKEX
func (o *OKCoin) SetDefaults() {
	o.SetErrorDefaults()
	o.SetCheckVarDefaults()
	o.Name = okCoinExchangeName
	o.Enabled = false
	o.Verbose = false
	o.RESTPollingDelay = 10
	o.APIWithdrawPermissions = exchange.AutoWithdrawCrypto |
		exchange.NoFiatWithdrawals
	o.RequestCurrencyPairFormat.Delimiter = "_"
	o.RequestCurrencyPairFormat.Uppercase = false
	o.ConfigCurrencyPairFormat.Delimiter = "_"
	o.ConfigCurrencyPairFormat.Uppercase = true
	o.SupportsAutoPairUpdating = true
	o.SupportsRESTTickerBatching = false
	o.Requester = request.New(o.Name,
		request.NewRateLimit(time.Second, okCoinAuthRate),
		request.NewRateLimit(time.Second, okCoinUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	o.APIUrlDefault = okCoinAPIURL
	o.APIUrl = okCoinAPIURL
	o.AssetTypes = []string{ticker.Spot}
	o.WebsocketInit()
	o.APIVersion = okCoinAPIVersion
	o.Websocket.Functionality = exchange.WebsocketTickerSupported |
		exchange.WebsocketTradeDataSupported |
		exchange.WebsocketKlineSupported |
		exchange.WebsocketOrderbookSupported
}
