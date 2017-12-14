package bitfinex

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

const (
	bitfinexAPIURL             = "https://api.bitfinex.com/v1/"
	bitfinexAPIVersion         = "1"
	bitfinexTicker             = "pubticker/"
	bitfinexStats              = "stats/"
	bitfinexLendbook           = "lendbook/"
	bitfinexOrderbook          = "book/"
	bitfinexTrades             = "trades/"
	bitfinexKeyPermissions     = "key_info"
	bitfinexLends              = "lends/"
	bitfinexSymbols            = "symbols/"
	bitfinexSymbolsDetails     = "symbols_details/"
	bitfinexAccountInfo        = "account_infos"
	bitfinexAccountFees        = "account_fees"
	bitfinexAccountSummary     = "summary"
	bitfinexDeposit            = "deposit/new"
	bitfinexOrderNew           = "order/new"
	bitfinexOrderNewMulti      = "order/new/multi"
	bitfinexOrderCancel        = "order/cancel"
	bitfinexOrderCancelMulti   = "order/cancel/multi"
	bitfinexOrderCancelAll     = "order/cancel/all"
	bitfinexOrderCancelReplace = "order/cancel/replace"
	bitfinexOrderStatus        = "order/status"
	bitfinexOrders             = "orders"
	bitfinexPositions          = "positions"
	bitfinexClaimPosition      = "position/claim"
	bitfinexHistory            = "history"
	bitfinexHistoryMovements   = "history/movements"
	bitfinexTradeHistory       = "mytrades"
	bitfinexOfferNew           = "offer/new"
	bitfinexOfferCancel        = "offer/cancel"
	bitfinexOfferStatus        = "offer/status"
	bitfinexOffers             = "offers"
	bitfinexMarginActiveFunds  = "taken_funds"
	bitfinexMarginTotalFunds   = "total_taken_funds"
	bitfinexMarginUnusedFunds  = "unused_taken_funds"
	bitfinexMarginClose        = "funding/close"
	bitfinexBalances           = "balances"
	bitfinexMarginInfo         = "margin_infos"
	bitfinexTransfer           = "transfer"
	bitfinexWithdrawal         = "withdraw"
	bitfinexActiveCredits      = "credits"

	// bitfinexMaxRequests if exceeded IP address blocked 10-60 sec, JSON response
	// {"error": "ERR_RATE_LIMIT"}
	bitfinexMaxRequests = 90
)

// Bitfinex is the overarching type across the bitfinex package
// Notes: Bitfinex has added a rate limit to the number of REST requests.
// Rate limit policy can vary in a range of 10 to 90 requests per minute
// depending on some factors (e.g. servers load, endpoint, etc.).
type Bitfinex struct {
	exchange.Base
	WebsocketConn         *websocket.Conn
	WebsocketSubdChannels map[int]WebsocketChanInfo
}

// SetDefaults sets the basic defaults for bitfinex
func (b *Bitfinex) SetDefaults() {
	b.Name = "Bitfinex"
	b.Enabled = false
	b.Verbose = false
	b.Websocket = false
	b.RESTPollingDelay = 10
	b.WebsocketSubdChannels = make(map[int]WebsocketChanInfo)
	b.RequestCurrencyPairFormat.Delimiter = ""
	b.RequestCurrencyPairFormat.Uppercase = true
	b.ConfigCurrencyPairFormat.Delimiter = ""
	b.ConfigCurrencyPairFormat.Uppercase = true
	b.AssetTypes = []string{ticker.Spot}
}

// Setup takes in the supplied exchange configuration details and sets params
func (b *Bitfinex) Setup(exch config.ExchangeConfig) {
	if !exch.Enabled {
		b.SetEnabled(false)
	} else {
		b.Enabled = true
		b.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		b.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		b.RESTPollingDelay = exch.RESTPollingDelay
		b.Verbose = exch.Verbose
		b.Websocket = exch.Websocket
		b.BaseCurrencies = common.SplitStrings(exch.BaseCurrencies, ",")
		b.AvailablePairs = common.SplitStrings(exch.AvailablePairs, ",")
		b.EnabledPairs = common.SplitStrings(exch.EnabledPairs, ",")
		err := b.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = b.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetTicker returns ticker information
func (b *Bitfinex) GetTicker(symbol string, values url.Values) (Ticker, error) {
	response := Ticker{}
	path := common.EncodeURLValues(bitfinexAPIURL+bitfinexTicker+symbol, values)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetStats returns various statistics about the requested pair
func (b *Bitfinex) GetStats(symbol string) ([]Stat, error) {
	response := []Stat{}
	path := fmt.Sprint(bitfinexAPIURL + bitfinexStats + symbol)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetFundingBook the entire margin funding book for both bids and asks sides
// per currency string
// symbol - example "USD"
func (b *Bitfinex) GetFundingBook(symbol string) (FundingBook, error) {
	response := FundingBook{}
	path := fmt.Sprint(bitfinexAPIURL + bitfinexLendbook + symbol)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetOrderbook retieves the orderbook bid and ask price points for a currency
// pair - By default the response will return 25 bid and 25 ask price points.
// CurrencyPair - Example "BTCUSD"
// Values can contain limit amounts for both the asks and bids - Example
// "limit_bids" = 1000
func (b *Bitfinex) GetOrderbook(currencyPair string, values url.Values) (Orderbook, error) {
	response := Orderbook{}
	path := common.EncodeURLValues(
		bitfinexAPIURL+bitfinexOrderbook+currencyPair,
		values,
	)
	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetTrades returns a list of the most recent trades for the given curencyPair
// By default the response will return 100 trades
// CurrencyPair - Example "BTCUSD"
// Values can contain limit amounts for the number of trades returned - Example
// "limit_trades" = 1000
func (b *Bitfinex) GetTrades(currencyPair string, values url.Values) ([]TradeStructure, error) {
	response := []TradeStructure{}
	path := common.EncodeURLValues(
		bitfinexAPIURL+bitfinexTrades+currencyPair,
		values,
	)
	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetLendbook returns a list of the most recent funding data for the given
// currency: total amount provided and Flash Return Rate (in % by 365 days) over
// time
// Symbol - example "USD"
func (b *Bitfinex) GetLendbook(symbol string, values url.Values) (Lendbook, error) {
	response := Lendbook{}
	if len(symbol) == 6 {
		symbol = symbol[:3]
	}
	path := common.EncodeURLValues(bitfinexAPIURL+bitfinexLendbook+symbol, values)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetLends returns a list of the most recent funding data for the given
// currency: total amount provided and Flash Return Rate (in % by 365 days)
// over time
// Symbol - example "USD"
func (b *Bitfinex) GetLends(symbol string, values url.Values) ([]Lends, error) {
	response := []Lends{}
	path := common.EncodeURLValues(bitfinexAPIURL+bitfinexLends+symbol, values)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetSymbols returns the available currency pairs on the exchange
func (b *Bitfinex) GetSymbols() ([]string, error) {
	products := []string{}
	path := fmt.Sprint(bitfinexAPIURL + bitfinexSymbols)

	return products, common.SendHTTPGetRequest(path, true, b.Verbose, &products)
}

// GetSymbolsDetails a list of valid symbol IDs and the pair details
func (b *Bitfinex) GetSymbolsDetails() ([]SymbolDetails, error) {
	response := []SymbolDetails{}
	path := fmt.Sprint(bitfinexAPIURL + bitfinexSymbolsDetails)

	return response, common.SendHTTPGetRequest(path, true, b.Verbose, &response)
}

// GetAccountInfo returns information about your account incl. trading fees
func (b *Bitfinex) GetAccountInfo() ([]AccountInfo, error) {
	response := []AccountInfo{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexAccountInfo, nil, &response)
}

// GetAccountFees - NOT YET IMPLEMENTED
func (b *Bitfinex) GetAccountFees() (AccountFees, error) {
	response := AccountFees{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexAccountFees, nil, &response)
}

// GetAccountSummary returns a 30-day summary of your trading volume and return
// on margin funding
func (b *Bitfinex) GetAccountSummary() (AccountSummary, error) {
	response := AccountSummary{}

	return response,
		b.SendAuthenticatedHTTPRequest(
			"POST", bitfinexAccountSummary, nil, &response,
		)
}

// NewDeposit returns a new deposit address
// Method - Example methods accepted: “bitcoin”, “litecoin”, “ethereum”,
//“tethers", "ethereumc", "zcash", "monero", "iota", "bcash"
// WalletName - accepted: “trading”, “exchange”, “deposit”
// renew - Default is 0. If set to 1, will return a new unused deposit address
func (b *Bitfinex) NewDeposit(method, walletName string, renew int) (DepositResponse, error) {
	response := DepositResponse{}
	request := make(map[string]interface{})
	request["method"] = method
	request["wallet_name"] = walletName
	request["renew"] = renew

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexDeposit, request, &response)
}

// GetKeyPermissions checks the permissions of the key being used to generate
// this request.
func (b *Bitfinex) GetKeyPermissions() (KeyPermissions, error) {
	response := KeyPermissions{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexKeyPermissions, nil, &response)
}

// GetMarginInfo shows your trading wallet information for margin trading
func (b *Bitfinex) GetMarginInfo() ([]MarginInfo, error) {
	response := []MarginInfo{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexMarginInfo, nil, &response)
}

// GetAccountBalance returns full wallet balance information
func (b *Bitfinex) GetAccountBalance() ([]Balance, error) {
	response := []Balance{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexBalances, nil, &response)
}

// WalletTransfer move available balances between your wallets
// Amount - Amount to move
// Currency -  example "BTC"
// WalletFrom - example "exchange"
// WalletTo -  example "deposit"
func (b *Bitfinex) WalletTransfer(amount float64, currency, walletFrom, walletTo string) ([]WalletTransfer, error) {
	response := []WalletTransfer{}
	request := make(map[string]interface{})
	request["amount"] = amount
	request["currency"] = currency
	request["walletfrom"] = walletFrom
	request["walletTo"] = walletTo

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexTransfer, request, &response)
}

// Withdrawal requests a withdrawal from one of your wallets.
// Major Upgrade needed on this function to include all query params
func (b *Bitfinex) Withdrawal(withdrawType, wallet, address string, amount float64) ([]Withdrawal, error) {
	response := []Withdrawal{}
	request := make(map[string]interface{})
	request["withdrawal_type"] = withdrawType
	request["walletselected"] = wallet
	request["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	request["address"] = address

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexWithdrawal, request, &response)
}

// NewOrder submits a new order and returns a order information
// Major Upgrade needed on this function to include all query params
func (b *Bitfinex) NewOrder(currencyPair string, amount float64, price float64, buy bool, Type string, hidden bool) (Order, error) {
	response := Order{}
	request := make(map[string]interface{})
	request["symbol"] = currencyPair
	request["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	request["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	request["exchange"] = "bitfinex"
	request["type"] = Type
	request["is_hidden"] = hidden

	if buy {
		request["side"] = "buy"
	} else {
		request["side"] = "sell"
	}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderNew, request, &response)
}

// NewOrderMulti allows several new orders at once
func (b *Bitfinex) NewOrderMulti(orders []PlaceOrder) (OrderMultiResponse, error) {
	response := OrderMultiResponse{}
	request := make(map[string]interface{})
	request["orders"] = orders

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderNewMulti, request, &response)
}

// CancelOrder cancels a single order
func (b *Bitfinex) CancelOrder(OrderID int64) (Order, error) {
	response := Order{}
	request := make(map[string]interface{})
	request["order_id"] = OrderID

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderCancel, request, &response)
}

// CancelMultipleOrders cancels multiple orders
func (b *Bitfinex) CancelMultipleOrders(OrderIDs []int64) (string, error) {
	response := GenericResponse{}
	request := make(map[string]interface{})
	request["order_ids"] = OrderIDs

	return response.Result,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderCancelMulti, request, nil)
}

// CancelAllOrders cancels all active and open orders
func (b *Bitfinex) CancelAllOrders() (string, error) {
	response := GenericResponse{}

	return response.Result,
		b.SendAuthenticatedHTTPRequest("GET", bitfinexOrderCancelAll, nil, nil)
}

// ReplaceOrder replaces an older order with a new order
func (b *Bitfinex) ReplaceOrder(OrderID int64, Symbol string, Amount float64, Price float64, Buy bool, Type string, Hidden bool) (Order, error) {
	response := Order{}
	request := make(map[string]interface{})
	request["order_id"] = OrderID
	request["symbol"] = Symbol
	request["amount"] = strconv.FormatFloat(Amount, 'f', -1, 64)
	request["price"] = strconv.FormatFloat(Price, 'f', -1, 64)
	request["exchange"] = "bitfinex"
	request["type"] = Type
	request["is_hidden"] = Hidden

	if Buy {
		request["side"] = "buy"
	} else {
		request["side"] = "sell"
	}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderCancelReplace, request, &response)
}

// GetOrderStatus returns order status information
func (b *Bitfinex) GetOrderStatus(OrderID int64) (Order, error) {
	orderStatus := Order{}
	request := make(map[string]interface{})
	request["order_id"] = OrderID

	return orderStatus,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderStatus, request, &orderStatus)
}

// GetActiveOrders returns all active orders and statuses
func (b *Bitfinex) GetActiveOrders() ([]Order, error) {
	response := []Order{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrders, nil, &response)
}

// GetActivePositions returns an array of active positions
func (b *Bitfinex) GetActivePositions() ([]Position, error) {
	response := []Position{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexPositions, nil, &response)
}

// ClaimPosition allows positions to be claimed
func (b *Bitfinex) ClaimPosition(PositionID int) (Position, error) {
	response := Position{}
	request := make(map[string]interface{})
	request["position_id"] = PositionID

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexClaimPosition, nil, nil)
}

// GetBalanceHistory returns balance history for the account
func (b *Bitfinex) GetBalanceHistory(symbol string, timeSince, timeUntil time.Time, limit int, wallet string) ([]BalanceHistory, error) {
	response := []BalanceHistory{}
	request := make(map[string]interface{})
	request["currency"] = symbol

	if !timeSince.IsZero() {
		request["since"] = timeSince
	}
	if !timeUntil.IsZero() {
		request["until"] = timeUntil
	}
	if limit > 0 {
		request["limit"] = limit
	}
	if len(wallet) > 0 {
		request["wallet"] = wallet
	}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexHistory, request, &response)
}

// GetMovementHistory returns an array of past deposits and withdrawals
func (b *Bitfinex) GetMovementHistory(symbol, method string, timeSince, timeUntil time.Time, limit int) ([]MovementHistory, error) {
	response := []MovementHistory{}
	request := make(map[string]interface{})
	request["currency"] = symbol

	if len(method) > 0 {
		request["method"] = method
	}
	if !timeSince.IsZero() {
		request["since"] = timeSince
	}
	if !timeUntil.IsZero() {
		request["until"] = timeUntil
	}
	if limit > 0 {
		request["limit"] = limit
	}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexHistoryMovements, request, &response)
}

// GetTradeHistory returns past executed trades
func (b *Bitfinex) GetTradeHistory(currencyPair string, timestamp, until time.Time, limit, reverse int) ([]TradeHistory, error) {
	response := []TradeHistory{}
	request := make(map[string]interface{})
	request["currency"] = currencyPair
	request["timestamp"] = timestamp

	if !until.IsZero() {
		request["until"] = until
	}
	if limit > 0 {
		request["limit"] = limit
	}
	if reverse > 0 {
		request["reverse"] = reverse
	}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexTradeHistory, request, &response)
}

// NewOffer submits a new offer
func (b *Bitfinex) NewOffer(symbol string, amount, rate float64, period int64, direction string) (Offer, error) {
	response := Offer{}
	request := make(map[string]interface{})
	request["currency"] = symbol
	request["amount"] = amount
	request["rate"] = rate
	request["period"] = period
	request["direction"] = direction

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOfferNew, request, &response)
}

// CancelOffer cancels offer by offerID
func (b *Bitfinex) CancelOffer(OfferID int64) (Offer, error) {
	response := Offer{}
	request := make(map[string]interface{})
	request["offer_id"] = OfferID

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOfferCancel, request, &response)
}

// GetOfferStatus checks offer status whether it has been cancelled, execute or
// is still active
func (b *Bitfinex) GetOfferStatus(OfferID int64) (Offer, error) {
	response := Offer{}
	request := make(map[string]interface{})
	request["offer_id"] = OfferID

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOrderStatus, request, &response)
}

// GetActiveCredits returns all available credits
func (b *Bitfinex) GetActiveCredits() ([]Offer, error) {
	response := []Offer{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexActiveCredits, nil, &response)
}

// GetActiveOffers returns all current active offers
func (b *Bitfinex) GetActiveOffers() ([]Offer, error) {
	response := []Offer{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexOffers, nil, &response)
}

// GetActiveMarginFunding returns an array of active margin funds
func (b *Bitfinex) GetActiveMarginFunding() ([]MarginFunds, error) {
	response := []MarginFunds{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexMarginActiveFunds, nil, &response)
}

// GetUnusedMarginFunds returns an array of funding borrowed but not currently
// used
func (b *Bitfinex) GetUnusedMarginFunds() ([]MarginFunds, error) {
	response := []MarginFunds{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexMarginUnusedFunds, nil, &response)
}

// GetMarginTotalTakenFunds returns an array of active funding used in a
// position
func (b *Bitfinex) GetMarginTotalTakenFunds() ([]MarginTotalTakenFunds, error) {
	response := []MarginTotalTakenFunds{}

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexMarginTotalFunds, nil, &response)
}

// CloseMarginFunding closes an unused or used taken fund
func (b *Bitfinex) CloseMarginFunding(SwapID int64) (Offer, error) {
	response := Offer{}
	request := make(map[string]interface{})
	request["swap_id"] = SwapID

	return response,
		b.SendAuthenticatedHTTPRequest("POST", bitfinexMarginClose, request, &response)
}

// SendAuthenticatedHTTPRequest sends an autheticated http request and json
// unmarshals result to a supplied variable
func (b *Bitfinex) SendAuthenticatedHTTPRequest(method, path string, params map[string]interface{}, result interface{}) error {
	if !b.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, b.Name)
	}

	if b.Nonce.Get() == 0 {
		b.Nonce.Set(time.Now().UnixNano())
	} else {
		b.Nonce.Inc()
	}

	respErr := ErrorCapture{}
	request := make(map[string]interface{})
	request["request"] = fmt.Sprintf("/v%s/%s", bitfinexAPIVersion, path)
	request["nonce"] = b.Nonce.String()

	if params != nil {
		for key, value := range params {
			request[key] = value
		}
	}

	PayloadJSON, err := common.JSONEncode(request)
	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON request")
	}

	if b.Verbose {
		log.Printf("Request JSON: %s\n", PayloadJSON)
	}

	PayloadBase64 := common.Base64Encode(PayloadJSON)
	hmac := common.GetHMAC(common.HashSHA512_384, []byte(PayloadBase64), []byte(b.APISecret))
	headers := make(map[string]string)
	headers["X-BFX-APIKEY"] = b.APIKey
	headers["X-BFX-PAYLOAD"] = PayloadBase64
	headers["X-BFX-SIGNATURE"] = common.HexEncodeToString(hmac)

	resp, err := common.SendHTTPRequest(
		method, bitfinexAPIURL+path, headers, strings.NewReader(""),
	)
	if err != nil {
		return err
	}

	if b.Verbose {
		log.Printf("Received raw: \n%s\n", resp)
	}

	if err = common.JSONDecode([]byte(resp), &respErr); err == nil {
		if len(respErr.Message) != 0 {
			return errors.New("Responded Error Issue: " + respErr.Message)
		}
	}

	if err = common.JSONDecode([]byte(resp), &result); err != nil {
		return errors.New("sendAuthenticatedHTTPRequest: Unable to JSON Unmarshal response")
	}
	return nil
}
