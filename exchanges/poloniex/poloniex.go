package poloniex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

const (
	poloniexAPIURL             = "https://api.poloniex.com"
	poloniexAltAPIUrl          = "https://api.poloniex.com"
	poloniexAPITradingEndpoint = "tradingApi"
	poloniexAPIVersion         = "1"
	poloniexMaxOrderbookDepth  = 100
)

var (
	errIntervalRequired                = errors.New("interval is required")
	errNilArgument                     = errors.New("error: nil argument")
	errAddressRequired                 = errors.New("address is required")
	errInvalidWithdrawalChain          = errors.New("invalid withdrawal chain")
	errClientOrderIDOROrderIDsRequired = errors.New("either client order IDs or order IDs or both are required")
	errInvalidTimeout                  = errors.New("timeout must not be empty")
	errChainsNotFound                  = errors.New("err chains not found")
)

// Poloniex is the overarching type across the poloniex package
type Poloniex struct {
	exchange.Base
	details CurrencyDetails
}

// Reference data endpoints.

// GetSymbolInformation all symbols and their tradeLimit info. priceScale is referring to the max number of decimals allowed for a given symbol.
func (p *Poloniex) GetSymbolInformation(ctx context.Context, symbol currency.Pair) ([]SymbolDetail, error) {
	var resp []SymbolDetail
	path := "/markets"
	if !symbol.IsEmpty() {
		path = fmt.Sprintf("%s/%s", path, symbol)
	}
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
}

// GetCurrencyInformations retrieves list of currencies and theiir detailed information.
func (p *Poloniex) GetCurrencyInformations(ctx context.Context) ([]CurrencyDetail, error) {
	var resp []CurrencyDetail
	path := "/currencies"
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
}

// GetCurrencyInformation retrieves currency and their detailed information.
func (p *Poloniex) GetCurrencyInformation(ctx context.Context, ccy currency.Code) (CurrencyDetail, error) {
	var resp CurrencyDetail
	path := "/currencies"
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	path = fmt.Sprintf("%s/%s", path, ccy.String())
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
}

// GetV2CurrencyInformations retrieves list of currency details for V2 API.
func (p *Poloniex) GetV2CurrencyInformations(ctx context.Context) ([]CurrencyV2Information, error) {
	var resp []CurrencyV2Information
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/v2/currencies", &resp)
}

// GetV2CurrencyInformations retrieves currency details for V2 API.
func (p *Poloniex) GetV2CurrencyInformation(ctx context.Context, ccy currency.Code) (*CurrencyV2Information, error) {
	var resp CurrencyV2Information
	path := "/v2/currencies"
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	path = fmt.Sprintf("%s/%s", path, ccy.String())
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, path, &resp)
}

// GetSystemTimestamp retrieves current server time.
func (p *Poloniex) GetSystemTimestamp(ctx context.Context) (*ServerSystemTime, error) {
	var resp ServerSystemTime
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/timestamp", &resp)
}

// Marker Data endpoints.

// GetMarketPrices retrieves latest trade price for all symbols.
func (p *Poloniex) GetMarketPrices(ctx context.Context) ([]MarketPrice, error) {
	var resp []MarketPrice
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/markets/price", &resp)
}

// GetMarketPrice retrieves latest trade price for all symbols.
func (p *Poloniex) GetMarketPrice(ctx context.Context, symbol currency.Pair) (*MarketPrice, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	var resp MarketPrice
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/price", symbol.String()), &resp)
}

// GetMarkPrices retrieves latest mark price for a single cross margin
func (p *Poloniex) GetMarkPrices(ctx context.Context) ([]MarkPrice, error) {
	var resp []MarkPrice
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/markets/markPrice", &resp)
}

// GetMarkPrice retrieves latest mark price for all cross margin symbol.
func (p *Poloniex) GetMarkPrice(ctx context.Context, symbol currency.Pair) (*MarkPrice, error) {
	var resp MarkPrice
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/markPrice", symbol.String()), &resp)
}

// MarkPriceComponents retrieves components of the mark price for a given symbol.
func (p *Poloniex) MarkPriceComponents(ctx context.Context, symbol currency.Pair) (*MarkPriceComponent, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	var resp MarkPriceComponent
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/markPriceComponents", symbol.String()), &resp)
}

// GetOrderbook retrieves the order book for a given symbol. Scale and limit values are optional.
// For valid scale values, please refer to the scale values defined for each symbol .
// If scale is not supplied, then no grouping/aggregation will be applied.
func (p *Poloniex) GetOrderbook(ctx context.Context, symbol currency.Pair) (*OrderbookData, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var resp OrderbookData
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/orderBook", symbol.String()), &resp)
}

// GetCandlesticks retrieves OHLC for a symbol at given timeframe (interval).
func (p *Poloniex) GetCandlesticks(ctx context.Context, symbol currency.Pair, interval kline.Interval, startTime, endTime time.Time, limit int64) ([]CandlestickData, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	intervalString, err := intervalToString(interval)
	if err != nil {
		return nil, err
	} else if intervalString == "" {
		return nil, errIntervalRequired
	}
	params := url.Values{}
	params.Set("interval", intervalString)
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	var resp []CandlestickArrayData
	err = p.SendHTTPRequest(ctx, exchange.RestSpot, common.EncodeURLValues(fmt.Sprintf("/markets/%s/candles", symbol.String()), params), &resp)
	if err != nil {
		return nil, err
	}
	return processCandlestickData(resp)
}

// GetTrades returns a list of recent trades, request param limit is optional, its default value is 500, and max value is 1000.
func (p *Poloniex) GetTrades(ctx context.Context, symbol currency.Pair, limit int64) ([]Trade, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []Trade
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, common.EncodeURLValues(fmt.Sprintf("/markets/%s/trades", symbol.String()), params), &resp)
}

// GetTicker retrieve ticker in last 24 hours for all symbols.
func (p *Poloniex) GetTickers(ctx context.Context) ([]TickerData, error) {
	var resp []TickerData
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/markets/ticker24h", &resp)
}

// GetTicker retrieve ticker in last 24 hours for provided symbols.
func (p *Poloniex) GetTicker(ctx context.Context, symbol currency.Pair) (*TickerData, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var resp TickerData
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/ticker24h", symbol.String()), &resp)
}

// Margin endpoints.

// GetCollateralInfos retrieves collateral information for all currencies.
func (p *Poloniex) GetCollateralInfos(ctx context.Context) ([]CollateralInfo, error) {
	var resp []CollateralInfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/markets/collateralInfo", &resp)
}

// GetCollateralInfo retrieves collateral information for all currencies.
func (p *Poloniex) GetCollateralInfo(ctx context.Context, ccy currency.Code) (*CollateralInfo, error) {
	var resp CollateralInfo
	return &resp, p.SendHTTPRequest(ctx, exchange.RestSpot, fmt.Sprintf("/markets/%s/collateralInfo", ccy.String()), &resp)
}

// GetBorrowRateInfo retrieves borrow rates information for all tiers and currencies.
func (p *Poloniex) GetBorrowRateInfo(ctx context.Context) ([]BorrowRateinfo, error) {
	var resp []BorrowRateinfo
	return resp, p.SendHTTPRequest(ctx, exchange.RestSpot, "/markets/borrowRatesInfo", &resp)
}

// ---------------- Authenticated endpoints ----------------------------------

// GetAccountInformation retrieves all accounts of a user.
func (p *Poloniex) GetAccountInformation(ctx context.Context) ([]AccountInformation, error) {
	var resp []AccountInformation
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/accounts", nil, nil, &resp)
}

// GetAllBalances get a list of all accounts of a user with each account’s id, type and balances (assets).
func (p *Poloniex) GetAllBalances(ctx context.Context, accountType string) ([]AccountBalance, error) {
	params := url.Values{}
	if accountType != "" {
		params.Set("accountType", accountType)
	}
	var resp []AccountBalance
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/accounts/balances", params, nil, &resp)
}

// GetAllBalance get an accounts of a user with each account’s id, type and balances (assets).
func (p *Poloniex) GetAllBalance(ctx context.Context, accountID, accountType string) ([]AccountBalance, error) {
	if accountID == "" {
		return nil, errAccountIDRequired
	}
	params := url.Values{}
	if accountType != "" {
		params.Set("accountType", accountType)
	}
	var resp []AccountBalance
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, fmt.Sprintf("/accounts/%s/balances", accountID), params, nil, &resp)
}

// GetAllAccountActivities retrieves a list of activities such as airdrop, rebates, staking, credit/debit adjustments, and other (historical adjustments).
// Type of activity: ALL: 200, AIRDROP: 201, COMMISSION_REBATE: 202, STAKING: 203, REFERAL_REBATE: 204, CREDIT_ADJUSTMENT: 104, DEBIT_ADJUSTMENT: 105, OTHER: 199
func (p *Poloniex) GetAllAccountActivities(ctx context.Context, startTime, endTime time.Time,
	activityType, limit, from int64, direction string, ccy currency.Code) ([]AccountActivity, error) {
	params := url.Values{}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if activityType != 0 {
		params.Set("activityType", strconv.FormatInt(activityType, 10))
	}
	if limit != 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if from != 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if !ccy.IsEmpty() {
		params.Set("currency", ccy.String())
	}
	var resp []AccountActivity
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/accounts/activity", params, nil, &resp)
}

// AccountsTransfer transfer amount of currency from an account to another account for a user.
func (p *Poloniex) AccountsTransfer(ctx context.Context, arg *AccountTransferParams) (*AccountTransferResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.Ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.Amount <= 0 {
		return nil, fmt.Errorf("%w, amount has to be greater than zero", order.ErrAmountIsInvalid)
	}
	if arg.FromAccount == "" {
		return nil, fmt.Errorf("%w, fromAccount=''", errAddressRequired)
	}
	if arg.ToAccount == "" {
		return nil, fmt.Errorf("%w, toAccount=''", errAddressRequired)
	}
	var resp AccountTransferResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/accounts/transfer", nil, arg, &resp)
}

// GetAccountTransferRecords gets a list of transfer records of a user. Max interval for start and end time is 6 months. If no start/end time params are specified then records for last 7 days will be returned.
func (p *Poloniex) GetAccountTransferRecords(ctx context.Context, startTime, endTime time.Time, direction string, ccy currency.Code, from, limit int64) ([]AccountTransferRecord, error) {
	params := url.Values{}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if !ccy.IsEmpty() {
		params.Set("currency", ccy.String())
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if from != 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []AccountTransferRecord
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/accounts/transfer", params, nil, &resp)
}

// GetAccountTransferRecord gets a transfer records of a user.
func (p *Poloniex) GetAccountTransferRecord(ctx context.Context, accountID string) ([]AccountTransferRecord, error) {
	var resp []AccountTransferRecord
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, fmt.Sprintf("/accounts/transfer/%s", accountID), nil, nil, &resp)
}

// GetFeeInfo retrieves fee rate for an account
func (p *Poloniex) GetFeeInfo(ctx context.Context) (*FeeInfo, error) {
	var resp FeeInfo
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/feeinfo", nil, nil, &resp)
}

// GetInterestHistory get a list of interest collection records of a user.
// Max interval for start and end time is 90 days.
// If no start/end time params are specified then records for last 7 days will be returned.
func (p *Poloniex) GetInterestHistory(ctx context.Context, startTime, endTime time.Time, direction string, from, limit int64) ([]InterestHistory, error) {
	params := url.Values{}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if from != 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []InterestHistory
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/accounts/interest/history", params, nil, &resp)
}

// ------------------------------------  Sub-Accounts endpoints ----------------------------

// GetSubAccountInformations get a list of all the accounts within an Account Group for a user.
func (p *Poloniex) GetSubAccountInformations(ctx context.Context) ([]SubAccount, error) {
	var resp []SubAccount
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/subaccounts", nil, nil, &resp)
}

// GetSubAccountBalances get balances information by currency and account type (SPOT and FUTURES) for each account in the account group.
// This is only functional for a primary user.
// A subaccount user can call /accounts/balances for SPOT account type and the futures API overview for its FUTURES balances.
func (p *Poloniex) GetSubAccountBalances(ctx context.Context) ([]SubAccountBalance, error) {
	var resp []SubAccountBalance
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/subaccounts/balances", nil, nil, &resp)
}

// GetSubAccountBalance get balances information by currency and account type (SPOT and FUTURES) for each account in the account group.
func (p *Poloniex) GetSubAccountBalance(ctx context.Context, id string) ([]SubAccountBalance, error) {
	var resp []SubAccountBalance
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, fmt.Sprintf("/subaccounts/%s/balances", id), nil, nil, &resp)
}

// SubAccountTransfer transfer amount of currency from an account and account type to another account and account type among the accounts in the account group.
// Primary account can transfer to and from any subaccounts as well as transfer between 2 subaccounts across account types.
// Subaccount can only transfer to the primary account across account types.
func (p *Poloniex) SubAccountTransfer(ctx context.Context, arg *SubAccountTransferParam) (*AccountTransferResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.Amount <= 0 {
		return nil, order.ErrAmountIsInvalid
	}
	if arg.FromAccountID == "" {
		return nil, fmt.Errorf("%w, fromAccountID=''", errAccountIDRequired)
	}
	if arg.ToAccountID == "" {
		return nil, fmt.Errorf("%w, toAccountID=''", errAccountIDRequired)
	}
	if arg.FromAccountType == "" {
		return nil, fmt.Errorf("%w, fromAccountType=''", errAccountTypeRequired)
	}
	if arg.ToAccountType == "" {
		return nil, fmt.Errorf("%w, toAccountType=''", errAccountTypeRequired)
	}
	var resp AccountTransferResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/subaccounts/transfer", nil, arg, &resp)
}

// GetSubAccountTransferRecords get a list of transfer records of a user. Max interval for start and end time is 6 months. If no start/end time params are specified then records for last 7 days will be returned.
func (p *Poloniex) GetSubAccountTransferRecords(ctx context.Context, ccy currency.Code, startTime,
	endTime time.Time, fromAccountID, toAccountID, fromAccountType,
	toAccountType, direction string, from, limit int64) ([]SubAccountTransfer, error) {
	params := url.Values{}
	if !ccy.IsEmpty() {
		params.Set("currency", ccy.String())
	}
	if !startTime.IsZero() {
		params.Set("startTime", startTime.String())
	}
	if !endTime.IsZero() {
		params.Set("endTime", endTime.String())
	}
	if fromAccountID != "" {
		params.Set("fromAccountID", fromAccountID)
	}
	if toAccountID != "" {
		params.Set("toAccountID", toAccountID)
	}
	if fromAccountType != "" {
		params.Set("fromAccountType", fromAccountType)
	}
	if toAccountType != "" {
		params.Set("toAccountType", toAccountType)
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if from > 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []SubAccountTransfer
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/subaccounts/transfer", params, nil, &resp)
}

// GetSubAccountTransferRecord retrieves a subaccount transfer record.
func (p *Poloniex) GetSubAccountTransferRecord(ctx context.Context, id string) ([]SubAccountTransfer, error) {
	var resp []SubAccountTransfer
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, fmt.Sprintf("/subaccounts/transfer/%s", id), nil, nil, &resp)
}

// -------------------------------------  Wallet sub-accounts  ---------------------------------------

// GetDepositAddresses get all deposit addresses for a user.
func (p *Poloniex) GetDepositAddresses(ctx context.Context, ccy currency.Code) (*DepositAddressesResponse, error) {
	addresses := &DepositAddressesResponse{}
	params := url.Values{}
	if !ccy.IsEmpty() {
		params.Set("currency", ccy.String())
	}
	return addresses, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/wallets/addresses", params, nil, addresses)
}

// WalletActivity returns the wallet activity between set start and end time
func (p *Poloniex) WalletActivity(ctx context.Context, start, end time.Time, activityType string) (*WalletActivityResponse, error) {
	values := url.Values{}
	err := common.StartEndTimeCheck(start, end)
	if err != nil {
		return nil, err
	}
	values.Set("start", strconv.FormatInt(start.Unix(), 10))
	values.Set("end", strconv.FormatInt(end.Unix(), 10))
	if activityType != "" {
		values.Set("activityType", activityType)
	}
	var resp WalletActivityResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/wallets/activity", values, nil, &resp)
}

// NewCurrencyDepoditAddress create a new address for a currency.
// Some currencies use a common deposit address for everyone on the exchange and designate the account
// for which this payment is destined by populating paymentID field.
// In these cases, use /currencies to look up the mainAccount for the currency to find
// the deposit address and use the address returned here as the paymentID.
// Note: currencies will only include a mainAccount property for currencies which require a paymentID.
func (p *Poloniex) NewCurrencyDepositAddress(ctx context.Context, ccy currency.Code) (string, error) {
	if ccy.IsEmpty() {
		return "", currency.ErrCurrencyCodeEmpty
	}
	resp := &struct {
		Address string `json:"address"`
	}{}
	return resp.Address, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/wallets/address", nil, map[string]string{"currency": ccy.String()}, &resp)
}

// func (p *Poloniex) WithdrawCurrency(ctx context.Context, )

// ---------------------------------------------- End ------------------------------------------------

func intervalToString(interval kline.Interval) (string, error) {
	intervalMap := map[kline.Interval]string{
		kline.OneMin:     "MINUTE_1",
		kline.FiveMin:    "MINUTE_5",
		kline.TenMin:     "MINUTE_10",
		kline.FifteenMin: "MINUTE_15",
		kline.ThirtyMin:  "MINUTE_30",
		kline.OneHour:    "HOUR_1",
		kline.TwoHour:    "HOUR_2",
		kline.FourHour:   "HOUR_4",
		kline.SixHour:    "HOUR_6",
		kline.TwelveHour: "HOUR_12",
		kline.OneDay:     "DAY_1",
		kline.ThreeDay:   "DAY_3",
		kline.SevenDay:   "WEEK_1",
		kline.OneMonth:   "MONTH_1",
	}
	intervalString, okay := intervalMap[interval]
	if okay {
		return intervalString, nil
	}
	return "", kline.ErrUnsupportedInterval
}

func stringToInterval(interval string) (kline.Interval, error) {
	intervalMap := map[string]kline.Interval{
		"MINUTE_1":  kline.OneMin,
		"MINUTE_5":  kline.FiveMin,
		"MINUTE_10": kline.TenMin,
		"MINUTE_15": kline.FifteenMin,
		"MINUTE_30": kline.ThirtyMin,
		"HOUR_1":    kline.OneHour,
		"HOUR_2":    kline.TwoHour,
		"HOUR_4":    kline.FourHour,
		"HOUR_6":    kline.SixHour,
		"HOUR_12":   kline.TwelveHour,
		"DAY_1":     kline.OneDay,
		"DAY_3":     kline.ThreeDay,
		"WEEK_1":    kline.SevenDay,
		"MONTH_1":   kline.OneMonth,
	}
	intervalInstance, okay := intervalMap[interval]
	if okay {
		return intervalInstance, nil
	}
	return kline.Interval(0), kline.ErrUnsupportedInterval
}

// WithdrawCurrency withdraws a currency to a specific delegated address.
// Immediately places a withdrawal for a given currency, with no email confirmation.
// In order to use this method, withdrawal privilege must be enabled for your API key.
// Some currencies use a common deposit address for everyone on the exchange and designate the account for
// which this payment is destined by populating paymentID field.
// In these cases, use /currencies to look up the mainAccount for the currency to find the deposit address and
// use the address returned by /wallets/addresses or generate one using /wallets/address as the paymentId.
// Note: currencies will only include a mainAccount property for currencies which require a paymentID.
// For currencies where there are multiple networks to choose from (like USDT or BTC), you can specify the chain by setting the "currency" parameter
//
//	to be a multiChain currency name, like USDTTRON, USDTETH, or BTCTRON. You can get information on these currencies,
//
// like fees or if they"re disabled, by adding the "includeMultiChainCurrencies" optional parameter to the /currencies endpoint.
func (p *Poloniex) WithdrawCurrency(ctx context.Context, arg *WithdrawCurrencyParam) (*Withdraw, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.Currency.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.Amount <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if arg.Address == "" {
		return nil, errAddressRequired
	}
	result := &Withdraw{}
	return result, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/wallets/withdraw", nil, arg, &result)
}

// WithdrawCurrencyV2 withdraws a currency to a specific delegated address.
// Immediately places a withdrawal for a given currency, with no email confirmation.
// In order to use this method, withdrawal privilege must be enabled for your API key.
func (p *Poloniex) WithdrawCurrencyV2(ctx context.Context, arg *WithdrawCurrencyV2Param) (*Withdraw, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.Coin.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	if arg.Amount <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if arg.Network == "" {
		return nil, errInvalidWithdrawalChain
	}
	if arg.Address == "" {
		return nil, errAddressRequired
	}
	var resp Withdraw
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/v2/wallets/withdraw", nil, arg, &resp)
}

// ---------------------------------------------------------------- Margin endpoints ------------------------------------------------

// GetAccountMarginInformation retrieves account margin information
func (p *Poloniex) GetAccountMarginInformation(ctx context.Context, accountType string) (*AccountMargin, error) {
	params := url.Values{}
	if accountType != "" {
		params.Set("accountType", accountType)
	}
	var resp AccountMargin
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/margin/accountMargin", params, nil, &resp)
}

// GetBorrowStatus retrieves borrow status of currencies
func (p *Poloniex) GetBorrowStatus(ctx context.Context, ccy currency.Code) ([]BorroweStatus, error) {
	params := url.Values{}
	if !ccy.IsEmpty() {
		params.Set("currency", ccy.String())
	}
	var resp []BorroweStatus
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/margin/borrowStatus", params, nil, &resp)
}

// MaximumBuySellAmount get maximum and available buy/sell amount for a given symbol.
func (p *Poloniex) MaximumBuySellAmount(ctx context.Context, symbol currency.Pair) (*MaxBuySellAmount, error) {
	if symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	params := url.Values{}
	params.Set("symbol", symbol.String())
	var resp MaxBuySellAmount
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/margin/maxSize", params, nil, &resp)
}

// ---------------------------------------------- Orders endpoint ------------------------------------------------------------

// PlaceOrder places an order for an account.
func (p *Poloniex) PlaceOrder(ctx context.Context, arg *PlaceOrderParams) (*PlaceOrderResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.Symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	var resp PlaceOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/orders", nil, arg, &resp)
}

// PlaceBatchOrders places a batch of order for an account.
func (p *Poloniex) PlaceBatchOrders(ctx context.Context, args []PlaceOrderParams) ([]PlaceBatchOrderRespItem, error) {
	if len(args) == 0 {
		return nil, errNilArgument
	}
	for x := range args {
		if args[x].Symbol.IsEmpty() {
			return nil, currency.ErrCurrencyPairEmpty
		}
		if args[x].Side == "" {
			return nil, order.ErrSideIsInvalid
		}
	}
	var resp []PlaceBatchOrderRespItem
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/orders/batch", nil, args, &resp)
}

// CancelReplaceOrder cancel an existing active order, new or partially filled, and place a new order on the same
// symbol with details from existing order unless amended by new parameters.
// The replacement order can amend price, quantity, amount, type, timeInForce, and allowBorrow fields.
// Specify the existing order id in the path; if id is a clientOrderId, prefix with cid: e.g. cid:myId-1.
// The proceedOnFailure flag is intended to specify whether to continue with new order placement in case cancelation of the existing order fails.
// Please note that since the new order placement does not wait for funds to clear from the existing order cancelation,
// it is possible that the new order will fail due to low available balance.
func (p *Poloniex) CancelReplaceOrder(ctx context.Context, arg *CancelReplaceOrderParam) (*CancelReplaceOrderResponse, error) {
	if arg == nil || (*arg) == (CancelReplaceOrderParam{}) {
		return nil, errNilArgument
	}
	if arg.ID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	var resp CancelReplaceOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPut, fmt.Sprintf("/orders/%s", arg.ID), nil, arg, &resp)
}

// GetOpenOrders retrieves a list of active orders for an account.
func (p *Poloniex) GetOpenOrders(ctx context.Context, symbol currency.Pair, side, direction, fromOrderID string, limit int64) ([]TradeOrder, error) {
	params := url.Values{}
	if !symbol.IsEmpty() {
		params.Set("symbol", symbol.String())
	}
	if side != "" {
		params.Set("side", side)
	}
	if fromOrderID != "" {
		params.Set("from", fromOrderID)
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []TradeOrder
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/orders", nil, nil, &resp)
}

// GetOrderDetail get an order’s status. Either by specifying orderId or clientOrderId. If id is a clientOrderId, prefix with cid: e.g. cid:myId-1
func (p *Poloniex) GetOrderDetail(ctx context.Context, id, clientOrderID string) (*TradeOrder, error) {
	var path string
	if id != "" {
		path = fmt.Sprintf("/orders/%s", id)
	} else if clientOrderID != "" {
		path = fmt.Sprintf("/orders/cid:%s", id)
	} else {
		return nil, fmt.Errorf("%w, orderid or client order id is required", order.ErrOrderIDNotSet)
	}
	var resp TradeOrder
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp)
}

// CancelOrderByID cancel an active order.
func (p *Poloniex) CancelOrderByID(ctx context.Context, id string) (*CancelOrderResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("%w; order 'id' is required", order.ErrOrderIDNotSet)
	}
	var resp CancelOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, fmt.Sprintf("/orders/%s", id), nil, nil, &resp)
}

// CancelMultipleOrdersByIDs batch cancel one or many active orders in an account by IDs.
func (p *Poloniex) CancelMultipleOrdersByIDs(ctx context.Context, args *OrderCancellationParams) ([]CancelOrderResponse, error) {
	if args == nil {
		return nil, errNilArgument
	}
	if len(args.ClientOrderIds) == 0 && len(args.OrderIds) == 0 {
		return nil, errClientOrderIDOROrderIDsRequired
	}
	var resp []CancelOrderResponse
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "/orders/cancelByIds", nil, args, &resp)
}

// CancelAllTradeOrders batch cancel all orders in an account.
func (p *Poloniex) CancelAllTradeOrders(ctx context.Context, symbols, accountTypes []string) ([]CancelOrderResponse, error) {
	args := make(map[string][]string)
	if len(symbols) > 0 {
		args["symbols"] = symbols
	}
	if len(accountTypes) > 0 {
		args["accountTypes"] = accountTypes
	}
	var resp []CancelOrderResponse
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "/orders", nil, args, &resp)
}

// KillSwitch set a timer that cancels all regular and smartorders after the timeout has expired.
// Timeout can be reset by calling this command again with a new timeout value.
// A timeout value of -1 disables the timer. Timeout is defined in seconds.
// timeout value in seconds; range is -1 and 10 to 600
func (p *Poloniex) KillSwitch(ctx context.Context, timeout string) (*KillSwitchStatus, error) {
	if timeout == "" {
		return nil, errInvalidTimeout
	}
	var resp *KillSwitchStatus
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/orders/killSwitch", nil, map[string]string{"timeout": timeout}, &resp)
}

// GetKillSwitchStatus get status of kill switch.
// If there is an active kill switch then the start and cancellation time is returned.
// If no active kill switch then an error message with code is returned
func (p *Poloniex) GetKillSwitchStatus(ctx context.Context) (*KillSwitchStatus, error) {
	var resp KillSwitchStatus
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/orders/killSwitchStatus", nil, nil, &resp)
}

// ---------------------------------------------- Smart Orders -----------------------------------------

// CreateSmartOrder create a smart order for an account. Funds will only be frozen when the smart order triggers, not upon smart order creation.
func (p *Poloniex) CreateSmartOrder(ctx context.Context, arg *SmartOrderRequestParam) (*PlaceOrderResponse, error) {
	if arg == nil || (*arg) == (SmartOrderRequestParam{}) {
		return nil, errNilArgument
	}
	if arg.Symbol.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	var resp PlaceOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "/smartorders", nil, arg, &resp)
}

func orderTypeString(oType order.Type) string {
	switch oType {
	case order.StopLimit:
		return "STOP_LIMIT"
	case order.AnyType, order.UnknownType:
		return ""
	default:
		return oType.String()
	}
}

// CancelReplaceSmartOrder cancel an existing untriggered smart order and place a new smart order on the same symbol with details from existing smart order unless amended by new parameters.
// The replacement smart order can amend price, stopPrice, quantity, amount, type, and timeInForce fields. Specify the existing smart order id in the path;
// if id is a clientOrderId, prefix with cid: e.g. cid:myId-1.
// The proceedOnFailure flag is intended to specify whether to continue with new smart order placement in case cancelation of the existing smart order fails.
func (p *Poloniex) CancelReplaceSmartOrder(ctx context.Context, arg *CancelReplaceSmartOrderParam) (*CancelReplaceSmartOrderResponse, error) {
	if arg == nil || (*arg) == (CancelReplaceSmartOrderParam{}) {
		return nil, errNilArgument
	}
	var path string
	if arg.ID != "" {
		path = fmt.Sprintf("/smartorders/%s", arg.ID)
	} else if arg.ClientOrderID != "" {
		path = fmt.Sprintf("/smartorders/cid:%s", arg.ClientOrderID)
	} else {
		return nil, errClientOrderIDOROrderIDsRequired
	}
	var resp CancelReplaceSmartOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodPut, path, nil, arg, &resp)
}

// GetSmartOpenOrders get a list of (pending) smart orders for an account.
func (p *Poloniex) GetSmartOpenOrders(ctx context.Context, limit int64) ([]SmartOrderItem, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp []SmartOrderItem
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/smartorders", params, nil, &resp)
}

// GetSmartOrderDetail get a smart order’s status. {id} can be smart order’s id or its clientOrderId (prefix with cid: ).
// If smart order’s state is TRIGGERED, the response will include the triggered order’s data
func (p *Poloniex) GetSmartOrderDetail(ctx context.Context, id, clientSuppliedID string) ([]SmartOrderDetail, error) {
	var path string
	if id != "" {
		path = fmt.Sprintf("/smartorders/%s", id)
	} else if clientSuppliedID != "" {
		path = fmt.Sprintf("/smartorders/cid:%s", clientSuppliedID)
	} else {
		return nil, errClientOrderIDOROrderIDsRequired
	}
	var resp []SmartOrderDetail
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp)
}

// CancelSmartOrderByID cancel a smart order by its id.
func (p *Poloniex) CancelSmartOrderByID(ctx context.Context, id, clientSuppliedID string) (*CancelSmartOrderResponse, error) {
	var path string
	if id != "" {
		path = fmt.Sprintf("/smartorders/%s", id)
	} else if clientSuppliedID != "" {
		path = fmt.Sprintf("/smartorders/cid:%s", clientSuppliedID)
	} else {
		return nil, errClientOrderIDOROrderIDsRequired
	}
	var resp CancelSmartOrderResponse
	return &resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, path, nil, nil, &resp)
}

// CancelMultipleSmartOrders performs a batch cancel one or many smart orders in an account by IDs.
func (p *Poloniex) CancelMultipleSmartOrders(ctx context.Context, args *OrderCancellationParams) ([]CancelOrderResponse, error) {
	if args == nil {
		return nil, errNilArgument
	}
	if len(args.ClientOrderIds) == 0 && len(args.OrderIds) == 0 {
		return nil, errClientOrderIDOROrderIDsRequired
	}
	var resp []CancelOrderResponse
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "/smartorders/cancelByIds", nil, args, &resp)
}

// CancelAllSmartOrders batch cancel all smart orders in an account.
func (p *Poloniex) CancelAllSmartOrders(ctx context.Context, symbols, accountTypes []string) ([]CancelOrderResponse, error) {
	args := make(map[string][]string)
	if len(symbols) > 0 {
		args["symbols"] = symbols
	}
	if len(accountTypes) > 0 {
		args["accountTypes"] = accountTypes
	}
	var resp []CancelOrderResponse
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "/smartorders", nil, args, &resp)
}

// ---------------------------------------------------------------- Order History ----------------------------------------------------------------

func orderFillParams(symbol currency.Pair,
	accountType, orderType, side, direction, states string,
	from, limit int64, startTime, endTime time.Time, hideCancel bool) url.Values {
	params := url.Values{}
	if accountType != "" {
		params.Set("accountType", accountType)
	}
	if orderType != "" {
		params.Set("type", orderType)
	}
	if side != "" {
		params.Set("side", side)
	}
	if !symbol.IsEmpty() {
		params.Set("symbol", symbol.String())
	}
	if from > 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	if states != "" {
		params.Set("states", states)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if hideCancel {
		params.Set("hideCancel", "true")
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	return params
}

// GetOrdersHistory get a list of historical orders in an account.
// Interval between startTime and endTime cannot exceed 30 days.
// If only endTime is populated then startTime will default to 7 days before endTime.
// If only startTime is populated then endTime will be defaulted to 7 days after startTime.
// Please note that canceled orders that are before 7 days from current time will be archived.
// Rest of the orders will be archived before 90 days from current time.
func (p *Poloniex) GetOrdersHistory(ctx context.Context, symbol currency.Pair,
	accountType, orderType, side, direction, states string,
	from, limit int64, startTime, endTime time.Time, hideCancel bool) ([]TradeOrder, error) {
	params := orderFillParams(symbol, accountType, orderType, side, direction, states, from, limit, startTime, endTime, hideCancel)
	var resp []TradeOrder
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/orders/history", params, nil, &resp)
}

// GetSmartOrderHistory get a list of historical smart orders in an account.
// Interval between startTime and endTime cannot exceed 30 days.
// If only endTime is populated then startTime will default to 7 days before endTime.
// If only startTime is populated then endTime will be defaulted to 7 days after startTime.
func (p *Poloniex) GetSmartOrderHistory(ctx context.Context, symbol currency.Pair,
	accountType, orderType, side, direction, states string,
	from, limit int64, startTime, endTime time.Time, hideCancel bool) ([]SmartOrderItem, error) {
	params := orderFillParams(symbol, accountType, orderType, side, direction, states, from, limit, startTime, endTime, hideCancel)
	var resp []SmartOrderItem
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/smartorders/history", params, nil, &resp)
}

// ------------------------------------------- Trade endpoints --------------------------------

// GetTradeHistory get a list of all trades for an account. Currently, trade history is supported since 07/30/2021.
// Interval between startTime and endTime cannot exceed 90 days.
// If only endTime is populated then startTime will default to 7 days before endTime.
// If only startTime is populated then endTime will be defaulted to 7 days after startTime.
func (p *Poloniex) GetTradeHistory(ctx context.Context, symbols currency.Pairs, direction string, from, limit int64, startTime, endTime time.Time) ([]TradeHistoryItem, error) {
	params := url.Values{}
	if len(symbols) == 0 {
		params.Set("symbols", symbols.Join())
	}
	if !startTime.IsZero() {
		params.Set("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		params.Set("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if from > 0 {
		params.Set("from", strconv.FormatInt(from, 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if direction != "" {
		params.Set("direction", direction)
	}
	var resp []TradeHistoryItem
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "/trades", params, nil, &resp)
}

// GetTradesByOrderID get a list of all trades for an order specified by its orderId.
func (p *Poloniex) GetTradesByOrderID(ctx context.Context, orderID string) ([]TradeHistoryItem, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	var resp []TradeHistoryItem
	return resp, p.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, fmt.Sprintf("/orders/%s/trades", orderID), nil, nil, &resp)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (p *Poloniex) SendHTTPRequest(ctx context.Context, ep exchange.URL, path string, result interface{}) error {
	endpoint, err := p.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}

	item := &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        result,
		Verbose:       p.Verbose,
		HTTPDebugging: p.HTTPDebugging,
		HTTPRecording: p.HTTPRecording,
	}

	return p.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		return item, nil
	}, request.UnauthenticatedRequest)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (p *Poloniex) SendAuthenticatedHTTPRequest(ctx context.Context, ep exchange.URL, method, endpoint string, values url.Values, body interface{}, result interface{}) error {
	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return err
	}
	ePoint, err := p.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	var rawResponse = &json.RawMessage{}
	requestFunc := func() (*request.Item, error) {
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"
		headers["key"] = creds.Key
		headers["recvWindow"] = strconv.FormatInt(1500, 10)
		if values == nil {
			values = url.Values{}
		}
		var bodyPayload []byte
		var signatureStrings string
		bodyPayload, err = json.Marshal(&struct{}{})
		timestamp := time.Now()
		values.Set("signTimestamp", strconv.FormatInt(timestamp.UnixMilli(), 10))
		switch method {
		case http.MethodGet, "get":
			signatureStrings = fmt.Sprintf("%s\n%s\n%s", http.MethodGet, endpoint, values.Encode())
		default:
			if body != nil {
				bodyPayload, err = json.Marshal(body)
				if err != nil {
					return nil, err
				}
			}
			if string(bodyPayload) != "{}" {
				signatureStrings = fmt.Sprintf("%s\n%s\n%s&%s", method, endpoint, "requestBody="+string(bodyPayload), values.Encode())
			} else {
				signatureStrings = fmt.Sprintf("%s\n%s\n%s", method, endpoint, values.Encode())
			}
		}
		hmac, err := crypto.GetHMAC(crypto.HashSHA256,
			[]byte(signatureStrings),
			[]byte(creds.Secret))
		if err != nil {
			return nil, err
		}

		headers["signatureMethod"] = "hmacSHA256"
		headers["signature"] = crypto.Base64Encode(hmac)
		headers["signTimestamp"] = strconv.FormatInt(timestamp.UnixMilli(), 10)
		values.Del("signTimestamp")
		path := common.EncodeURLValues(fmt.Sprintf("%s%s", ePoint, endpoint), values)
		req := &request.Item{
			Method:        method,
			Path:          path,
			Result:        rawResponse,
			Headers:       headers,
			Verbose:       p.Verbose,
			HTTPDebugging: p.HTTPDebugging,
			HTTPRecording: p.HTTPRecording,
		}
		if method != http.MethodGet && len(bodyPayload) > 0 {
			req.Body = bytes.NewBuffer(bodyPayload)
		}
		return req, nil
	}
	err = p.SendPayload(ctx, request.Unset, requestFunc, request.AuthenticatedRequest)
	if err != nil {
		return err
	}
	errorResp := &errorResponse{}
	err = json.Unmarshal(*rawResponse, errorResp)
	if err == nil && errorResp.Code != 0 {
		return fmt.Errorf("error code: %d message: %s", errorResp.Code, errorResp.Message)
	}
	return json.Unmarshal(*rawResponse, result)
}

// GetFee returns an estimate of fee based on type of transaction
func (p *Poloniex) GetFee(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		feeInfo, err := p.GetFeeInfo(ctx)
		if err != nil {
			return 0, err
		}
		fee = calculateTradingFee(feeInfo,
			feeBuilder.PurchasePrice,
			feeBuilder.Amount,
			feeBuilder.IsMaker)

	case exchange.CryptocurrencyWithdrawalFee:
		fee = getWithdrawalFee(feeBuilder.Pair.Base)
	case exchange.OfflineTradeFee:
		fee = getOfflineTradeFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

// getOfflineTradeFee calculates the worst case-scenario trading fee
func getOfflineTradeFee(price, amount float64) float64 {
	return 0.002 * price * amount
}

func calculateTradingFee(feeInfo *FeeInfo, purchasePrice, amount float64, isMaker bool) (fee float64) {
	if isMaker {
		fee = feeInfo.MakerRate.Float64()
	} else {
		fee = feeInfo.TakerRate.Float64()
	}
	return fee * amount * purchasePrice
}

func getWithdrawalFee(c currency.Code) float64 {
	return WithdrawalFees[c]
}
