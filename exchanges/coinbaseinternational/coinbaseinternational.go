package coinbaseinternational

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

// CoinbaseInternational is the overarching type across this package
type CoinbaseInternational struct {
	exchange.Base
}

const (
	coinbaseInternationalAPIURL = "https://api.international.coinbase.com"
	coinbaseAPIVersion          = "/api/v1"
)

var (
	errArgumentMustBeInterface   = errors.New("argument must be an interface")
	errMissingPortfolioID        = errors.New("missing portfolio identification")
	errNetworkArnID              = errors.New("identifies the blockchain network")
	errMissingTransferID         = errors.New("missing transfer ID")
	errAddressIsRequired         = errors.New("err: missing address")
	errAssetIdentifierIsRequired = errors.New("err: asset identified is required")
	errEmptyArgument             = errors.New("err: empty argument")
	errTimeInForceRequired       = errors.New("err: time_in_force is required")
)

// ListAssets returns a list of all supported assets.
func (co *CoinbaseInternational) ListAssets(ctx context.Context) ([]AssetItemInfo, error) {
	var resp []AssetItemInfo
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "assets", nil, nil, &resp, false)
}

// GetAssetDetails retrieves information for a specific asset.
func (co *CoinbaseInternational) GetAssetDetails(ctx context.Context, assetName currency.Code, assetUUID, assetID string) (*AssetItemInfo, error) {
	path := "assets/"
	switch {
	case !assetName.IsEmpty():
		path += assetName.String()
	case assetUUID != "":
		path += assetUUID
	case assetID != "":
		path += assetID
	default:
		return nil, errors.New("missing asset information; ")
	}
	var resp AssetItemInfo
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, false)
}

// GetSupportedNetworkPerAsset returns a list of supported networks and network information for a specific asset.
func (co *CoinbaseInternational) GetSupportedNetworksPerAsset(ctx context.Context, assetName currency.Code, assetUUID, assetID string) ([]AssetInfoWithSupportedNetwork, error) {
	path := "assets/"
	switch {
	case !assetName.IsEmpty():
		path += assetName.String()
	case assetUUID != "":
		path += assetUUID
	case assetID != "":
		path += assetID
	default:
		return nil, errors.New("missing asset information; ")
	}
	path += "/networks"
	var resp []AssetInfoWithSupportedNetwork
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, false)
}

// GetInstruments returns all of the instruments available for trading.
func (co *CoinbaseInternational) GetInstruments(ctx context.Context) ([]InstrumentInfo, error) {
	var resp []InstrumentInfo
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "instruments", nil, nil, &resp, false)
}

// GetInstrumentDetails retrieves market information for a specific instrument.
func (co *CoinbaseInternational) GetInstrumentDetails(ctx context.Context, instrumentName, instrumentUUID, instrumentID string) (*InstrumentInfo, error) {
	path := "instruments/"
	switch {
	case instrumentName != "":
		path += instrumentName
	case instrumentUUID != "":
		path += instrumentUUID
	case instrumentID != "":
		path += instrumentID
	default:
		return nil, errors.New("instrument information is required")
	}
	var resp InstrumentInfo
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, false)
}

// GetQuotePerInstrument retrieves the current quote for a specific instrument.
func (co *CoinbaseInternational) GetQuotePerInstrument(ctx context.Context, instrumentName, instrumentUUID, instrumentID string) (*QuoteInformation, error) {
	path := "instruments/"
	switch {
	case instrumentName != "":
		path += instrumentName
	case instrumentUUID != "":
		path += instrumentUUID
	case instrumentID != "":
		path += instrumentID
	default:
		return nil, errors.New("instrument information is required")
	}
	var resp QuoteInformation
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path+"/quote", nil, nil, &resp, false)
}

// CreateOrder creates a new order.
func (co *CoinbaseInternational) CreateOrder(ctx context.Context, arg *OrderRequestParams) (*TradeOrder, error) {
	if arg == nil || *arg == (OrderRequestParams{}) {
		return nil, common.ErrNilPointer
	}
	if arg.Side == "" {
		return nil, order.ErrSideIsInvalid
	}
	if arg.BaseSize <= 0 {
		return nil, order.ErrAmountBelowMin
	}
	if arg.Price <= 0 {
		return nil, order.ErrPriceBelowMin
	}
	if arg.OrderType == "" {
		return nil, order.ErrUnsupportedOrderType
	}
	if arg.ClientOrderID == "" {
		return nil, fmt.Errorf("%w, client_order_id is required", order.ErrOrderIDNotSet)
	}
	if arg.TimeInForce == "" {
		return nil, errTimeInForceRequired
	}
	var resp TradeOrder
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "orders", nil, arg, &resp, true)
}

// GetOpenOrders returns a list of active orders resting on the order book matching the requested criteria. Does not return any rejected, cancelled, or fully filled orders as they are not active.
func (co *CoinbaseInternational) GetOpenOrders(ctx context.Context, portfolioUUID, portfolioID, instrument, clientOrderID, eventType string, RefDateTime time.Time, resultOffset, resultLimit int64) (*OrderItemDetail, error) {
	params := url.Values{}
	switch {
	case portfolioID != "":
		params.Set("portfolio", portfolioID)
	case portfolioUUID != "":
		params.Set("portfolio", portfolioUUID)
	}
	if instrument != "" {
		params.Set("instrument", instrument)
	}
	if clientOrderID != "" {
		params.Set("client_order_id", clientOrderID)
	}
	if eventType != "" {
		params.Set("event_type", eventType)
	}
	if !RefDateTime.IsZero() {
		params.Set("ref_datetime", RefDateTime.String())
	}
	if resultOffset > 0 {
		params.Set("result_offset", strconv.FormatInt(resultOffset, 10))
	}
	if resultLimit > 0 {
		params.Set("result_limit", strconv.FormatInt(resultLimit, 10))
	}
	var resp OrderItemDetail
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "orders", params, nil, &resp, true)
}

// CancelOrders cancels all orders matching the requested criteria.
func (co *CoinbaseInternational) CancelOrders(ctx context.Context, portfolioID, portfolioUUID, instrument string) ([]OrderItem, error) {
	params := url.Values{}
	switch {
	case portfolioID != "":
		params.Set("portfolio", portfolioID)
	case portfolioUUID != "":
		params.Set("portfolio", portfolioUUID)
	default:
		return nil, errMissingPortfolioID
	}
	if instrument != "" {
		params.Set("instrument", instrument)
	}
	var resp []OrderItem
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "orders", params, nil, &resp, true)
}

// ModifyOpenOrder modifies an open order.
func (co *CoinbaseInternational) ModifyOpenOrder(ctx context.Context, orderID string, arg *ModifyOrderParam) (*OrderItem, error) {
	if orderID == "" {
		return nil, order.ErrOrderIDNotSet
	}
	if arg == (&ModifyOrderParam{}) {
		return nil, fmt.Errorf("%w, empty modification parameter", common.ErrNilPointer)
	}
	var resp OrderItem
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodPut, "orders/"+orderID, nil, arg, &resp, true)
}

// GetOrderDetails retrieves a single order. The order retrieved can be either active or inactive.
func (co *CoinbaseInternational) GetOrderDetails(ctx context.Context, orderID string) (*OrderItem, error) {
	var resp OrderItem
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "orders/"+orderID, nil, nil, &resp, true)
}

// CancelTradeOrder cancels a single open order.
func (co *CoinbaseInternational) CancelTradeOrder(ctx context.Context, orderID, clientOrderID, portfolioID, portfolioUUID string) (*OrderItem, error) {
	switch {
	case orderID != "":
	case clientOrderID != "":
		orderID = clientOrderID
	default:
		return nil, order.ErrOrderIDNotSet
	}
	params := url.Values{}
	switch {
	case portfolioID != "":
		params.Set("portfolio", portfolioID)
	case portfolioUUID != "":
		params.Set("portfolio", portfolioUUID)
	default:
		return nil, errMissingPortfolioID
	}
	var resp OrderItem
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodDelete, "orders/"+orderID, params, nil, &resp, true)
}

// GetAllUserPortfolios returns all of the user's portfolios.
func (co *CoinbaseInternational) GetAllUserPortfolios(ctx context.Context) ([]PortfolioItem, error) {
	var resp []PortfolioItem
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "portfolios", nil, nil, &resp, true)
}

// GetPortfolioDetails retrieves the summary, positions, and balances of a portfolio.
func (co *CoinbaseInternational) GetPortfolioDetails(ctx context.Context, portfolioID, portfolioUUID string) (*PortfolioDetail, error) {
	if portfolioID == "" {
		portfolioID = portfolioUUID
	} else if portfolioUUID == "" {
		return nil, errMissingPortfolioID
	}
	var resp PortfolioDetail
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "portfolios/"+portfolioID+"/detail", nil, nil, &resp, true)
}

// GetPortfolioSummary retrieves the high level overview of a portfolio.
func (co *CoinbaseInternational) GetPortfolioSummary(ctx context.Context, portfolioUUID, portfolioID string) (*PortfolioSummary, error) {
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/summary"
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/summary"
	default:
		return nil, errMissingPortfolioID
	}
	var resp PortfolioSummary
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// ListPortfolioBalances returns all of the balances for a given portfolio.
func (co *CoinbaseInternational) ListPortfolioBalances(ctx context.Context, portfolioUUID, portfolioID string) ([]PortfolioBalance, error) {
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/balances"
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/balances"
	default:
		return nil, errMissingPortfolioID
	}
	var resp []PortfolioBalance
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// GetPortfolioAssetBalance retrieves the balance for a given portfolio and asset.
func (co *CoinbaseInternational) GetPortfolioAssetBalance(ctx context.Context, portfolioUUID, portfolioID string, ccy currency.Code) (*PortfolioBalance, error) {
	if ccy.IsEmpty() {
		return nil, currency.ErrCurrencyCodeEmpty
	}
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/balances/" + ccy.String()
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/balances/" + ccy.String()
	default:
		return nil, errMissingPortfolioID
	}
	var resp PortfolioBalance
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// ListPortfolioPositions returns all of the positions for a given portfolio.
func (co *CoinbaseInternational) ListPortfolioPositions(ctx context.Context, portfolioUUID, portfolioID string) ([]PortfolioPosition, error) {
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/positions"
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/positions"
	default:
		return nil, errMissingPortfolioID
	}
	var resp []PortfolioPosition
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// GetPortfolioInstrumentPosition retrieves the position for a given portfolio and symbol.
func (co *CoinbaseInternational) GetPortfolioInstrumentPosition(ctx context.Context, portfolioUUID, portfolioID string, instrument currency.Pair) (*PortfolioPosition, error) {
	if instrument.IsEmpty() {
		return nil, currency.ErrCurrencyPairEmpty
	}
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/positions/" + instrument.String()
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/positions/" + instrument.String()
	default:
		return nil, errMissingPortfolioID
	}
	var resp PortfolioPosition
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// ListPortfolioFills returns all of the fills for a given portfolio.
func (co *CoinbaseInternational) ListPortfolioFills(ctx context.Context, portfolioUUID, portfolioID string) ([]PortfolioFill, error) {
	var path string
	switch {
	case portfolioUUID != "":
		path = "portfolios/" + portfolioUUID + "/fills"
	case portfolioID != "":
		path = "portfolios/" + portfolioID + "/fills"
	default:
		return nil, errMissingPortfolioID
	}
	var resp []PortfolioFill
	return resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, path, nil, nil, &resp, true)
}

// ListMatchingTransfers represents a list of transfer based on the query
// type: possible values DEPOSIT, WITHDRAW, REBATE, STIPEND
// status: possible value PROCESSED, NEW, FAILED, STARTED

func (co *CoinbaseInternational) ListMatchingTransfers(ctx context.Context, portfolioUUID, portfolioID, status, transferType string, resultLimit, resultOffset int64, timeFrom, timeTo time.Time) (*Transfers, error) {
	params := url.Values{}
	switch {
	case portfolioUUID != "":
		params.Set("portfolio", portfolioUUID)
	case portfolioID != "":
		params.Set("portfolio", portfolioID)
	}
	if resultOffset > 0 {
		params.Set("result_offset", strconv.FormatInt(resultOffset, 10))
	}
	if resultLimit > 0 {
		params.Set("result_limit", strconv.FormatInt(resultLimit, 10))
	}
	if status != "" {
		params.Set("status", status)
	}
	if transferType != "" {
		params.Set("type", status)
	}
	if !timeFrom.IsZero() {
		params.Set("time_from", timeFrom.String())
	}
	if !timeTo.IsZero() {
		params.Set("time_to", timeTo.String())
	}
	var resp Transfers
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "transfers", params, nil, &resp, true)
}

// GetTransfer returns a single transfer instance
func (co *CoinbaseInternational) GetTransfer(ctx context.Context, transferID string) (*FundTransfer, error) {
	if transferID == "" {
		return nil, errMissingTransferID
	}
	var resp FundTransfer
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodGet, "transfers/"+transferID, nil, nil, &resp, true)
}

// WithdrawToCryptoAddress withdraws a crypto fund to crypto address
func (co *CoinbaseInternational) WithdrawToCryptoAddress(ctx context.Context, arg *WithdrawCryptoParams) (*WithdrawalResponse, error) {
	if arg == nil {
		return nil, common.ErrNilPointer
	}
	if arg.Address == "" {
		return nil, errAddressIsRequired
	}
	if arg.Amount < 0 {
		return nil, order.ErrAmountIsInvalid
	}
	if arg.AssetIdentifier == "" {
		return nil, errAssetIdentifierIsRequired
	}
	var resp WithdrawalResponse
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "transfers/withdraw", nil, arg, &resp, true)
}

// CreateCryptoAddress created a new crypto address
func (co *CoinbaseInternational) CreateCryptoAddress(ctx context.Context, arg *CryptoAddressParam) (*CryptoAddressInfo, error) {
	if arg == nil {
		return nil, common.ErrNilPointer
	}
	if arg.AssetIdentifier == "" {
		return nil, errAssetIdentifierIsRequired
	}
	if arg.Portfolio == "" {
		return nil, errMissingPortfolioID
	}
	if arg.NetworkArnID == "" {
		return nil, errNetworkArnID
	}
	var resp CryptoAddressInfo
	return &resp, co.SendHTTPRequest(ctx, exchange.RestSpot, http.MethodPost, "transfers/address", nil, arg, &resp, true)

}

// SendHTTPRequest sends a public HTTP request.
func (co *CoinbaseInternational) SendHTTPRequest(ctx context.Context, ep exchange.URL, method, path string, params url.Values, data, result interface{}, authenticated bool) error {
	endpoint, err := co.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	urlPath := endpoint + coinbaseAPIVersion + "/" + path
	if params != nil {
		urlPath = common.EncodeURLValues(urlPath, params)
	}
	requestType := request.AuthType(request.UnauthenticatedRequest)
	var creds *account.Credentials
	if authenticated {
		creds, err = co.GetCredentials(ctx)
		if err != nil {
			return err
		}
		requestType = request.AuthenticatedRequest
	}
	value := reflect.ValueOf(data)
	var payload []byte
	if value != (reflect.Value{}) && !value.IsNil() && value.Kind() != reflect.Ptr {
		return errArgumentMustBeInterface
	} else if value != (reflect.Value{}) && !value.IsNil() {
		payload, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	return co.SendPayload(ctx, request.Unset, func() (*request.Item, error) {
		timestamp := time.Now()
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"
		headers["Accept"] = "application/json"

		if authenticated {
			headers["CB-ACCESS-KEY"] = creds.Key
			headers["CB-ACCESS-PASSPHRASE"] = creds.ClientID
			headers["CB-ACCESS-TIMESTAMP"] = strconv.FormatInt(timestamp.Unix(), 10)
			signatureString := headers["CB-ACCESS-TIMESTAMP"] + method + coinbaseAPIVersion + "/" + path + string(payload)
			var hmac []byte
			secretBytes, err := crypto.Base64Decode(creds.Secret)
			if err != nil {
				return nil, err
			}
			hmac, err = crypto.GetHMAC(crypto.HashSHA256,
				[]byte(signatureString),
				secretBytes)
			if err != nil {
				return nil, err
			}
			headers["CB-ACCESS-SIGN"] = crypto.Base64Encode(hmac)
		}

		return &request.Item{
			Method:        method,
			Path:          urlPath,
			Headers:       headers,
			Result:        result,
			Body:          bytes.NewBuffer(payload),
			Verbose:       co.Verbose,
			HTTPDebugging: co.HTTPDebugging,
			HTTPRecording: co.HTTPRecording,
		}, nil
	}, requestType)
}

func orderTypeString(oType order.Type) (string, error) {
	switch oType {
	case order.Limit, order.Market, order.Stop:
		return oType.String(), nil
	case order.StopLimit:
		return "STOP_LIMIT", nil
	default:
		return "", order.ErrUnsupportedOrderType
	}
}

// GetFee returns an estimate of fee based on type of transaction
func (co *CoinbaseInternational) GetFee(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	var err error
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee, err = co.calculateTradingFee(
			ctx,
			feeBuilder.Pair.Base,
			feeBuilder.Pair.Quote,
			feeBuilder.PurchasePrice,
			feeBuilder.Amount,
			feeBuilder.IsMaker)
		if err != nil {
			return 0, err
		}
	case exchange.OfflineTradeFee:
		fee = getOfflineTradeFee(feeBuilder.Pair, feeBuilder.PurchasePrice, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

func (co *CoinbaseInternational) calculateTradingFee(ctx context.Context, base, quote currency.Code, purchasePrice, amount float64, isMaker bool) (float64, error) {
	fees, err := co.GetAllUserPortfolios(ctx)
	if err != nil {
		return 0, err
	}
	for x := range fees {
		if strings.EqualFold(fees[x].Name, currency.Pair{Base: base, Delimiter: "-", Quote: quote}.String()) {
			if isMaker {
				return fees[x].MakerFeeRate.Float64() * amount * purchasePrice, nil
			}
			return fees[x].TakerFeeRate.Float64() * amount * purchasePrice, nil
		}
	}
	if isMaker {
		return 0.018 * amount * purchasePrice, nil
	}
	return 0.02 * amount * purchasePrice, nil
}

// getOfflineTradeFee calculates the worst case-scenario trading fee
func getOfflineTradeFee(c currency.Pair, price, amount float64) float64 {
	return 0.02 * price * amount
}
