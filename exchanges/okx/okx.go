package okx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"

	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"

	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
)

// Okx is the overarching type across this package
type Okx struct {
	exchange.Base
}

const (
	okxRateInterval = time.Second
	okxAPIURL       = "https://aws.okex.com/" + okxAPIPath
	okxAPIVersion   = "/v5/"

	okxStandardRequestRate = 6
	okxAPIPath             = "api" + okxAPIVersion
	okxExchangeName        = "Okx"
	okxWebsocketURL        = "wss://ws.okx.com:8443/ws" + okxAPIVersion

	okxAPIWebsocketPublicURL  = okxWebsocketURL + "public"
	okxAPIWebsocketPrivateURL = okxWebsocketURL + "private"

	// tradeEndpoints
	tradeOrder                          = "trade/order"
	placeMultipleOrderUrl               = "trade/batch-orders"
	cancelTradeOrder                    = "trade/cancel-order"
	cancelBatchTradeOrders              = "trade/cancel-batch-orders"
	amendOrder                          = "trade/amend-order"
	amendBatchOrders                    = "trade/amend-batch-orders"
	closePositionPath                   = "trade/close-position"
	pandingTradeOrders                  = "trade/orders-pending"
	tradeHistory                        = "trade/orders-history"
	orderHistoryArchive                 = "trade/orders-history-archive"
	tradeFills                          = "trade/fills"
	tradeFillsHistory                   = "trade/fills-history"
	assetbills                          = "asset/bills"
	lightningDeposit                    = "asset/deposit-lightning"
	assetDeposits                       = "asset/deposit-address"
	pathToAssetDepositHistory           = "asset/deposit-history"
	assetWithdrawal                     = "asset/withdrawal"
	assetLightningWithdrawal            = "asset/withdrawal-lightning"
	cancelWithdrawal                    = "asset/cancel-withdrawal"
	withdrawalHistory                   = "asset/withdrawal-history"
	smallAssetsConvert                  = "asset/convert-dust-assets"
	assetSavingBalance                  = "asset/saving-balance"
	assetSavingPurchaseOrRedemptionPath = "asset/purchase_redempt"
	assetsLendingHistory                = "asset/lending-history"
	assetSetLendingRateRoute            = "asset/set-lending-rate"
	publicBorrowInfo                    = "asset/lending-rate-history"

	// Algo order routes
	cancelAlgoOrder        = "trade/cancel-algos"
	algoTradeOrder         = "trade/order-algo"
	canceAdvancedAlgoOrder = "trade/cancel-advance-algos"
	getAlgoOrders          = "trade/orders-algo-pending"
	algoOrderHistory       = "trade/orders-algo-history"

	// Funding orders routes
	assetCurrencies    = "asset/currencies"
	assetBalance       = "asset/balances"
	assetValuation     = "asset/asset-valuation"
	assetTransfer      = "asset/transfer"
	assetTransferState = "asset/transfer-state"

	// Market Data
	marketTickers                = "market/tickers"
	marketTicker                 = "market/ticker"
	indexTickers                 = "market/index-tickers"
	marketBooks                  = "market/books"
	marketCandles                = "market/candles"
	marketCandlesHistory         = "market/history-candles"
	marketCandlesIndex           = "market/index-candles"
	marketPriceCandles           = "market/mark-price-candles"
	marketTrades                 = "market/trades"
	marketTradesHistory          = "market/history-trades"
	marketPlatformVolumeIn24Hour = "market/platform-24-volume"
	marketOpenOracles            = "market/open-oracle"
	marketExchangeRate           = "market/exchange-rate"
	marketIndexComponents        = "market/index-components"
	marketBlockTickers           = "market/block-tickers"
	marketBlockTicker            = "market/block-ticker"
	marketBlockTrades            = "market/block-trades"

	// Public endpoints
	publicInstruments                 = "public/instruments"
	publicDeliveryExerciseHistory     = "public/delivery-exercise-history"
	publicOpenInterestValues          = "public/open-interest"
	publicFundingRate                 = "public/funding-rate"
	publicFundingRateHistory          = "public/funding-rate-history"
	publicLimitPath                   = "public/price-limit"
	publicOptionalData                = "public/opt-summary"
	publicEstimatedPrice              = "public/estimated-price"
	publicDiscountRate                = "public/discount-rate-interest-free-quota"
	publicTime                        = "public/time"
	publicLiquidationOrders           = "public/liquidation-orders"
	publicMarkPrice                   = "public/mark-price"
	publicPositionTiers               = "public/position-tiers"
	publicInterestRateAndLoanQuota    = "public/interest-rate-loan-quota"
	publicVIPInterestRateAndLoanQuota = "public/vip-interest-rate-loan-quota"
	publicUnderlyings                 = "public/underlying"
	publicInsuranceFunds              = "public/insurance-fund"
	publicCurrencyConvertContract     = "public/convert-contract-coin"

	// Trading Endpoints
	tradingDataSupportedCoins      = "rubik/stat/trading-data/support-coin"
	tradingTakerVolume             = "rubik/stat/taker-volume"
	tradingMarginLoanRatio         = "rubik/stat/margin/loan-ratio"
	longShortAccountRatio          = "rubik/stat/contracts/long-short-account-ratio"
	contractOpenInterestVolume     = "rubik/stat/contracts/open-interest-volume"
	optionOpenInterestVolume       = "rubik/stat/option/open-interest-volume"
	optionOpenInterestVolumeRatio  = "rubik/stat/option/open-interest-volume-ratio"
	optionOpenInterestVolumeExpiry = "rubik/stat/option/open-interest-volume-expiry"
	optionOpenInterestVolumeStrike = "rubik/stat/option/open-interest-volume-strike"
	takerBlockVolume               = "rubik/stat/option/taker-block-volume"

	// Convert Currencies end points
	assetConvertCurrencies   = "asset/convert/currencies"
	convertCurrencyPairsPath = "asset/convert/currency-pair"
	assetEstimateQuote       = "asset/convert/estimate-quote"
	assetConvertTrade        = "asset/convert/trade"
	assetConvertHistory      = "asset/convert/history"

	// Account Endpints
	accountBalance               = "account/balance"
	accountPosition              = "account/positions"
	accountPositionHistory       = "account/positions-history"
	accountAndPositionRisk       = "account/account-position-risk"
	accountBillsDetail           = "account/bills"
	accountBillsDetailArchive    = "account/bills-archive"
	accountConfiguration         = "account/config"
	accountSetPositionMode       = "account/set-position-mode"
	accountSetLeverage           = "account/set-leverage"
	accountMaxSize               = "account/max-size"
	accountMaxAvailSize          = "account/max-avail-size"
	accountPositionMarginBalance = "account/position/margin-balance"
	accountLeverageInfo          = "account/leverage-info"
	accountMaxLoan               = "account/max-loan"
	accountTradeFee              = "account/trade-fee"
	accountInterestAccured       = "account/interest-accrued"
	accountInterestRate          = "account/interest-rate"
	accountSetGeeks              = "account/set-greeks"
	accountSetIsolatedMode       = "account/set-isolated-mode"
	accountMaxWithdrawal         = "account/max-withdrawal"
	accountRiskState             = "account/risk-state"
	accountBorrowReply           = "account/borrow-repay"
	accountBorrowRepayHistory    = "account/borrow-repay-history"
	accountInterestLimits        = "account/interest-limits"
	accountSimulatedMargin       = "account/simulated_margin"
	accountGeeks                 = "account/greeks"

	// Block Trading
	rfqCounterparties    = "rfq/counterparties"
	rfqCreateRFQ         = "rfq/create-rfq"
	rfqCancelRfq         = "rfq/cancel-rfq"
	rfqCancelRfqs        = "rfq/cancel-batch-rfqs"
	rfqCancelAllRfqs     = "rfq/cancel-all-rfqs"
	rfqExecuteQuote      = "rfq/execute-quote"
	rfqCreateQuote       = "rfq/create-quote"
	rfqCancelQuote       = "rfq/cancel-quote"
	rfqCancelBatchQuotes = "rfq/cancel-batch-quotes"
	rfqCancelAllQuotes   = "rfq/cancel-all-quotes"
	rfqRfqs              = "rfq/rfqs"
	rfqQuotes            = "rfq/quotes"
	rfqTrades            = "rfq/trades"
	rfqPublicTrades      = "rfq/public-trades"
	// Subaccount endpoints
	usersSubaccountList          = "users/subaccount/list"
	accountSubaccountBalances    = "account/subaccount/balances"
	assetSubaccountBalances      = "asset/subaccount/balances"
	assetSubaccountBills         = "asset/subaccount/bills"
	assetSubaccountTransfer      = "asset/subaccount/transfer"
	userSubaccountSetTransferOut = "users/subaccount/set-transfer-out"
	usersEntrustSubaccountList   = "users/entrust-subaccount-list"

	// Grid Trading Endpoints
	gridOrderAlgo         = "tradingBot/grid/order-algo"
	gridAmendOrderAlgo    = "tradingBot/grid/amend-order-algo"
	gridAlgoOrderStop     = "tradingBot/grid/stop-order-algo"
	gridAlgoOrders        = "tradingBot/grid/orders-algo-pending"
	gridAlgoOrdersHistory = "tradingBot/grid/orders-algo-history"
	gridOrdersAlgoDetails = "tradingBot/grid/orders-algo-details"
	gridSuborders         = "tradingBot/grid/sub-orders"
	gridPositions         = "tradingBot/grid/positions"
	gridWithdrawalIncome  = "tradingBot/grid/withdraw-income"

	// Status Endpoints

	systemStatus = "system/status"
)

var (
	errUnableToTypeAssertResponseData               = errors.New("unable to type assert responseData")
	errUnableToTypeAssertKlineData                  = errors.New("unable to type assert kline data")
	errUnexpectedKlineDataLength                    = errors.New("unexpected kline data length")
	errLimitExceedsMaximumResultPerRequest          = errors.New("maximum result per request exeeds the limit")
	errNo24HrTradeVolumeFound                       = errors.New("no trade record found in the 24 trade volume ")
	errOracleInformationNotFound                    = errors.New("oracle informations not found")
	errExchangeInfoNotFound                         = errors.New("exchange information not found")
	errIndexComponentNotFound                       = errors.New("unable to fetch index components")
	errMissingRequiredArgInstType                   = errors.New("invalid required argument instrument type")
	errLimitValueExceedsMaxof100                    = fmt.Errorf("limit value exceeds the maximum value %d", 100)
	errParameterUnderlyingCanNotBeEmpty             = errors.New("parameter uly can not be empty")
	errMissingInstrumentID                          = errors.New("missing instrument id")
	errFundingRateHistoryNotFound                   = errors.New("funding rate history not found")
	errMissingRequiredUnderlying                    = errors.New("error missing required parameter underlying")
	errMissingRequiredParamInstID                   = errors.New("missing required parameter instrument id")
	errLiquidationOrderResponseNotFound             = errors.New("liquidation order not found")
	errEitherInstIDOrCcyIsRequired                  = errors.New("either parameter instId or ccy is required")
	errIncorrectRequiredParameterTradeMode          = errors.New("unacceptable required argument, trade mode")
	errInterestRateAndLoanQuotaNotFound             = errors.New("interest rate and loan quota not found")
	errUnderlyingsForSpecifiedInstTypeNofFound      = errors.New("underlyings for the specified instrument id is not found")
	errInsuranceFundInformationNotFound             = errors.New("insurance fund information not found")
	errMissingExpiryTimeParameter                   = errors.New("missing expiry date parameter")
	errParsingResponseError                         = errors.New("error while parsing the response")
	errInvalidTradeModeValue                        = errors.New("invalid trade mode value")
	errMissingOrderSide                             = errors.New("missing order side")
	errInvalidOrderType                             = errors.New("invalid order type")
	errInvalidQuantityToButOrSell                   = errors.New("unacceptable quantity to buy or sell")
	errMissingClientOrderIDOrOrderID                = errors.New("client supplier order id or order id is missing")
	errMissingNewSizeOrPriceInformation             = errors.New("missing the new size or price information")
	errMissingNewSize                               = errors.New("missing the order size information")
	errMissingMarginMode                            = errors.New("missing required param margin mode \"mgnMode\"")
	errMissingTradeMode                             = errors.New("missing trade mode")
	errInvalidTriggerPrice                          = errors.New("invalid trigger price value")
	errMssingAlgoOrderID                            = errors.New("missing algo orders id")
	errInvalidAverageAmountParameterValue           = errors.New("invalid average amount parameter value")
	errInvalidPriceLimit                            = errors.New("invalid price limit value")
	errMissingIntervalValue                         = errors.New("missing interval value")
	errMissingEitherAlgoIDOrState                   = errors.New("either algo id or order state is required")
	errUnacceptableAmount                           = errors.New("amount must be greater than 0")
	errInvalidCurrencyValue                         = errors.New("invalid currency value")
	errInvalidDepositAmount                         = errors.New("invalid deposit amount")
	errMissingResponseBody                          = errors.New("error missing response body")
	errMissingValidWithdrawalID                     = errors.New("missing valid withdrawal id")
	errNoValidResponseFromServer                    = errors.New("no valid response from server")
	errInvalidInstrumentType                        = errors.New("invlaid instrument type")
	errMissingValidGeeskType                        = errors.New("missing valid geeks type")
	errMissingIsolatedMarginTradingSetting          = errors.New("missing isolated margin trading setting, isolated margin trading settings automatic:Auto transfers autonomy:Manual transfers")
	errInvalidOrderSide                             = errors.New(" invalid order side")
	errInvalidCounterParties                        = errors.New("missing counter parties")
	errInvalidLegs                                  = errors.New("no legs are provided")
	errMissingRFQIDANDClientSuppliedRFQID           = errors.New("missing rfq id or client supplied rfq id")
	errMissingRfqIDOrQuoteID                        = errors.New("either RFQ ID or Quote ID is missing")
	errMissingRfqID                                 = errors.New("error missing rfq id")
	errMissingLegs                                  = errors.New("missing legs")
	errMissingSizeOfQuote                           = errors.New("missing size of quote leg")
	errMossingLegsQuotePrice                        = errors.New("error missing quote price")
	errMissingQuoteIDOrClientSuppliedQuoteID        = errors.New("missing quote id or client supplied quote id")
	errMissingEitherQuoteIDAOrlientSuppliedQuoteIDs = errors.New("missing either quote ids or client supplied quote ids")
	errMissingRequiredParameterSubaccountName       = errors.New("missing required parameter subaccount name")
	errInvalidTransferAmount                        = errors.New("unacceptable transfer amount")
	errInvalidInvalidSubaccount                     = errors.New("invalid account type")
	errMissingDestinationSubaccountName             = errors.New("missing destination subaccount name")
	errMissingInitialSubaccountName                 = errors.New("missing initial subaccount name")
	errMissingAlgoOrderType                         = errors.New("missing algo order type \"grid\": Spot grid, \"contract_grid\": Contract grid")
	errInvalidMaximumPrice                          = errors.New("invalid maximum price")
	errInvalidMinimumPrice                          = errors.New("invalid minimum price")
	errInvalidGridQuantity                          = errors.New("invalid grid quantity (grid number)")
	errMissingSize                                  = errors.New("missing size")
	errMissingRequiredArgumentDirection             = errors.New("missing required argument, direction")
	errRequiredParameterMissingLeverage             = errors.New("missing required parameter, leverage")
	errMissingAlgoOrderID                           = errors.New("missing algo order id")
	errMissingValidStopType                         = errors.New("invalid grid order stop type, only valiues are \"1\" and \"2\" ")
	errMissingSubOrderType                          = errors.New("missing sub order type")
	errMissingQuantity                              = errors.New("invalid quantity to buy or sell")
	errDepositAddressNotFound                       = errors.New("deposit address with the specified currency code and chain not found")
	errMissingAtLeast1CurrencyPair                  = errors.New("at least one currenct is required to fetch order history")
	errNoCandlestickDataFound                       = errors.New("no candlesticks data found")
	errInvalidWebsocketEvent                        = errors.New("invalid websocket event")
	errMissingValidChannelInformation               = errors.New("missing channel information")
)

/************************************ MarketData Endpoints *************************************************/

// PlaceOrder place an order only if you have sufficient funds.
func (ok *Okx) PlaceOrder(ctx context.Context, arg PlaceOrderRequestParam) (*PlaceOrderResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	arg.TradeMode = strings.Trim(arg.TradeMode, " ")
	if !(strings.EqualFold("cross", arg.TradeMode) || strings.EqualFold("isolated", arg.TradeMode) || strings.EqualFold("cash", arg.TradeMode)) {
		return nil, errInvalidTradeModeValue
	}
	if !(strings.EqualFold(arg.Side, "buy") || strings.EqualFold(arg.Side, "sell")) {
		return nil, errMissingOrderSide
	}
	arg.Side = strings.Trim(strings.ToLower(arg.Side), " ")
	if !(strings.EqualFold(arg.OrderType, OkxOrderMarket) ||
		strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
		strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
		strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
		strings.EqualFold(arg.OrderType, OkxOrderIOC) ||
		strings.EqualFold(arg.OrderType, OkxOrderOptimalLimitIOC)) {
		return nil, errInvalidOrderType
	}
	if arg.QuantityToBuyOrSell <= 0 {
		return nil, errInvalidQuantityToButOrSell
	}
	if arg.OrderPrice <= 0 && (strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
		strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
		strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
		strings.EqualFold(arg.OrderType, OkxOrderIOC)) {
		return nil, fmt.Errorf("invalid order price for %s order types", arg.OrderType)
	}
	if !(strings.EqualFold(arg.QuantityType, "base_ccy") || strings.EqualFold(arg.QuantityType, "quote_ccy")) {
		arg.QuantityType = ""
	}
	type response struct {
		Msg  string               `json:"msg"`
		Code string               `json:"code"`
		Data []PlaceOrderResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, placeOrderEPL, http.MethodPost, tradeOrder, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("error parsing the response")
	}
	return &resp.Data[0], nil
}

// PlaceMultipleOrders  to place orders in batches. Maximum 20 orders can be placed at a time. Request parameters should be passed in the form of an array.
func (ok *Okx) PlaceMultipleOrders(ctx context.Context, args []PlaceOrderRequestParam) ([]PlaceOrderResponse, error) {
	for x := range args {
		arg := args[x]
		if arg.InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		arg.TradeMode = strings.Trim(arg.TradeMode, " ")
		if !(strings.EqualFold("cross", arg.TradeMode) || strings.EqualFold("isolated", arg.TradeMode) || strings.EqualFold("cash", arg.TradeMode)) {
			return nil, errInvalidTradeModeValue
		}
		if !(strings.EqualFold(arg.Side, "buy") || strings.EqualFold(arg.Side, "sell")) {
			return nil, errMissingOrderSide
		}
		if !(strings.EqualFold(arg.OrderType, "market") || strings.EqualFold(arg.OrderType, "limit") || strings.EqualFold(arg.OrderType, "post_only") ||
			strings.EqualFold(arg.OrderType, "fok") || strings.EqualFold(arg.OrderType, "ioc") || strings.EqualFold(arg.OrderType, "optimal_limit_ioc")) {
			return nil, errInvalidOrderType
		}
		if arg.QuantityToBuyOrSell <= 0 {
			return nil, errInvalidQuantityToButOrSell
		}
		if arg.OrderPrice <= 0 && (strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
			strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
			strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
			strings.EqualFold(arg.OrderType, OkxOrderIOC)) {
			return nil, fmt.Errorf("invalid order price for %s order types", arg.OrderType)
		}
		if !(strings.EqualFold(arg.QuantityType, "base_ccy") || strings.EqualFold(arg.QuantityType, "quote_ccy")) {
			arg.QuantityType = ""
		}
	}
	type response struct {
		Msg  string               `json:"msg"`
		Code string               `json:"code"`
		Data []PlaceOrderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeMultipleOrdersEPL, http.MethodPost, placeMultipleOrderUrl, &args, &resp, true)
}

// CancelSingleOrder cancel an incomplete order.
func (ok *Okx) CancelSingleOrder(ctx context.Context, arg CancelOrderRequestParam) (*CancelOrderResponse, error) {
	if strings.Trim(arg.InstrumentID, " ") == "" {
		return nil, errMissingInstrumentID
	}
	if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
		return nil, fmt.Errorf("either order id or client supplier id is required")
	}
	type response struct {
		Msg  string                `json:"msg"`
		Code string                `json:"code"`
		Data []CancelOrderResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelOrderEPL, http.MethodPost, cancelTradeOrder, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("error parsing the response data")
	}
	return &resp.Data[0], nil
}

// CancelMultipleOrders cancel incomplete orders in batches. Maximum 20 orders can be canceled at a time.
// Request parameters should be passed in the form of an array.
func (ok *Okx) CancelMultipleOrders(ctx context.Context, args []CancelOrderRequestParam) ([]CancelOrderResponse, error) {
	for x := range args {
		arg := args[x]
		if strings.Trim(arg.InstrumentID, " ") == "" {
			return nil, errMissingInstrumentID
		}
		if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
			return nil, fmt.Errorf("either order id or client supplier id is required")
		}
	}
	type response struct {
		Msg  string                `json:"msg"`
		Code string                `json:"code"`
		Data []CancelOrderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleOrdersEPL,
		http.MethodPost, cancelBatchTradeOrders, args, &resp, true)
}

// AmendOrder an incomplete order.
func (ok *Okx) AmendOrder(ctx context.Context, arg *AmendOrderRequestParams) (*AmendOrderResponse, error) {
	if strings.Trim(arg.InstrumentID, " ") == "" {
		return nil, errMissingInstrumentID
	}
	if arg.ClientSuppliedOrderID == "" && arg.OrderID == "" {
		return nil, errMissingClientOrderIDOrOrderID
	}
	if arg.NewQuantity <= 0 && arg.NewPrice <= 0 {
		return nil, errMissingNewSizeOrPriceInformation
	}
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []AmendOrderResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, amendOrderEPL, http.MethodPost, amendOrder, arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errParsingResponseError
	}
	return &resp.Data[0], nil
}

// AmendMultipleOrders amend incomplete orders in batches. Maximum 20 orders can be amended at a time. Request parameters should be passed in the form of an array.
func (ok *Okx) AmendMultipleOrders(ctx context.Context, args []AmendOrderRequestParams) ([]AmendOrderResponse, error) {
	for x := range args {
		if strings.Trim(args[x].InstrumentID, " ") == "" {
			return nil, errMissingInstrumentID
		}
		if args[x].ClientSuppliedOrderID == "" && args[x].OrderID == "" {
			return nil, errMissingClientOrderIDOrOrderID
		}
		if args[x].NewQuantity <= 0 && args[x].NewPrice <= 0 {
			return nil, errMissingNewSizeOrPriceInformation
		}
	}
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []AmendOrderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, amendMultipleOrdersEPL, http.MethodPost, amendBatchOrders, &args, &resp, true)
}

// ClosePositions Close all positions of an instrument via a market order.
func (ok *Okx) ClosePositions(ctx context.Context, arg *ClosePositionsRequestParams) (*ClosePositionResponse, error) {
	if strings.Trim(arg.InstrumentID, " ") == "" {
		return nil, errMissingInstrumentID
	}
	if !(arg.MarginMode != "" && (strings.EqualFold(arg.MarginMode, "cross") || strings.EqualFold(arg.MarginMode, "isolated"))) {
		return nil, errMissingMarginMode
	}
	type response struct {
		Msg  string                  `json:"msg"`
		Code string                  `json:"code"`
		Data []ClosePositionResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, closePositionEPL, http.MethodPost, closePositionPath, arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetOrderDetails retrieve order details.
func (ok *Okx) GetOrderDetail(ctx context.Context, arg *OrderDetailRequestParam) (*OrderDetail, error) {
	params := url.Values{}
	if strings.Trim(arg.InstrumentID, " ") == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", arg.InstrumentID)
	if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
		return nil, errMissingClientOrderIDOrOrderID
	} else if arg.ClientSupplierOrderID == "" {
		params.Set("ordId", arg.OrderID)
	} else {
		params.Set("clOrdId", arg.ClientSupplierOrderID)
	}
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []OrderDetail `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(tradeOrder, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderDetEPL, http.MethodGet, path, nil, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetOrderList retrieve all incomplete orders under the current account.
func (ok *Okx) GetOrderList(ctx context.Context, arg *OrderListRequestParams) ([]PendingOrderItem, error) {
	params := url.Values{}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeSpot) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeOption) {
		params.Set("instType", arg.InstrumentType)
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if strings.EqualFold(arg.OrderType, OkxOrderMarket) ||
		strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
		strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
		strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
		strings.EqualFold(arg.OrderType, OkxOrderIOC) ||
		strings.EqualFold(arg.OrderType, OkxOrderOptimalLimitIOC) {
		params.Set("orderType", arg.OrderType)
	}
	if strings.EqualFold(arg.State, "canceled") ||
		strings.EqualFold(arg.State, "filled") {
		params.Set("state", arg.State)
	}
	if !(arg.Before.IsZero()) {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !(arg.After.IsZero()) {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	path := common.EncodeURLValues(pandingTradeOrders, params)
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []PendingOrderItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderListEPL, http.MethodGet, path, nil, &resp, true)
}

// Get7DayOrderHistory retrieve the completed order data for the last 7 days, and the incomplete orders that have been cancelled are only reserved for 2 hours.
func (ok *Okx) Get7DayOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams) ([]PendingOrderItem, error) {
	return ok.getOrderHistory(ctx, arg, tradeHistory)
}

// Get3MonthOrderHistory retrieve the completed order data for the last 7 days, and the incomplete orders that have been cancelled are only reserved for 2 hours.
func (ok *Okx) Get3MonthOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams) ([]PendingOrderItem, error) {
	return ok.getOrderHistory(ctx, arg, orderHistoryArchive)
}

// getOrderHistory retrives the order history of the past limited times
func (ok *Okx) getOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams, route string) ([]PendingOrderItem, error) {
	params := url.Values{}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeSpot) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeOption) {
		params.Set("instType", arg.InstrumentType)
	} else {
		return nil, errMissingRequiredArgInstType
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if strings.EqualFold(arg.OrderType, OkxOrderMarket) ||
		strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
		strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
		strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
		strings.EqualFold(arg.OrderType, OkxOrderIOC) ||
		strings.EqualFold(arg.OrderType, OkxOrderOptimalLimitIOC) {
		params.Set("orderType", arg.OrderType)
	}
	if strings.EqualFold(arg.State, "canceled") ||
		strings.EqualFold(arg.State, "filled") {
		params.Set("state", arg.State)
	}
	if !(arg.Before.IsZero()) {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !(arg.After.IsZero()) {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	if strings.EqualFold("twap", arg.Category) || strings.EqualFold("adl", arg.Category) || strings.EqualFold("full_liquidation", arg.Category) || strings.EqualFold("partial_liquidation", arg.Category) || strings.EqualFold("delivery", arg.Category) || strings.EqualFold("ddh", arg.Category) {
		params.Set("category", strings.ToLower(arg.Category))
	}
	path := common.EncodeURLValues(route, params)
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []PendingOrderItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// GetTransactionDetailsLast3Days retrieve recently-filled transaction details in the last 3 day.
func (ok *Okx) GetTransactionDetailsLast3Days(ctx context.Context, arg *TransactionDetailRequestParams) ([]TransactionDetail, error) {
	return ok.getTransactionDetails(ctx, arg, tradeFills)
}

// GetTransactionDetailsLast3Months Retrieve recently-filled transaction details in the last 3 months.
func (ok *Okx) GetTransactionDetailsLast3Months(ctx context.Context, arg *TransactionDetailRequestParams) ([]TransactionDetail, error) {
	return ok.getTransactionDetails(ctx, arg, tradeFillsHistory)
}

// GetTransactionDetails retrieve recently-filled transaction details.
func (ok *Okx) getTransactionDetails(ctx context.Context, arg *TransactionDetailRequestParams, route string) ([]TransactionDetail, error) {
	params := url.Values{}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeSpot) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeOption) {
		params.Set("instType", arg.InstrumentType)
	} else {
		return nil, errMissingRequiredArgInstType
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if strings.EqualFold(arg.OrderType, OkxOrderMarket) ||
		strings.EqualFold(arg.OrderType, OkxOrderLimit) ||
		strings.EqualFold(arg.OrderType, OkxOrderPostOnly) ||
		strings.EqualFold(arg.OrderType, OkxOrderFOK) ||
		strings.EqualFold(arg.OrderType, OkxOrderIOC) ||
		strings.EqualFold(arg.OrderType, OkxOrderOptimalLimitIOC) {
		params.Set("orderType", arg.OrderType)
	}
	if !(arg.Begin.IsZero()) {
		params.Set("begin", strconv.FormatInt(arg.Begin.UnixMilli(), 10))
	}
	if !(arg.End.IsZero()) {
		params.Set("end", strconv.FormatInt(arg.End.UnixMilli(), 10))
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	if arg.After != "" {
		params.Set("after", arg.After)
	}
	if arg.Before != "" {
		params.Set("before", arg.Before)
	}
	path := common.EncodeURLValues(route, params)
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []TransactionDetail `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTrasactionDetailsEPL, http.MethodGet, path, nil, &resp, true)
}

// PlaceAlgoOrder order includes trigger order, oco order, conditional order,iceberg order, twap order and trailing order.
func (ok *Okx) PlaceAlgoOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if !(arg.InstrumentID != "") {
		return nil, errMissingInstrumentID
	}
	if !(arg.TradeMode != "") {
		return nil, errMissingTradeMode
	}
	if arg.Side == order.UnknownSide {
		return nil, errMissingOrderSide
	}
	if arg.OrderType == "" {
		return nil, errInvalidOrderType
	}
	if arg.Size <= 0 {
		return nil, errMissingNewSize
	}
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodGet, algoTradeOrder, arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}
	return nil, errors.New(resp.Msg)
}

// StopOrderParams
func (ok *Okx) StopOrderParams(ctx context.Context, arg *StopOrderParams) ([]AlgoOrder, error) {
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodPost, algoTradeOrder, arg, &resp, true)
}

// PlaceTrailingStopOrder
func (ok *Okx) PlaceTrailingStopOrder(ctx context.Context, arg *TrailingStopOrderRequestParam) ([]AlgoOrder, error) {
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodPost, algoTradeOrder, arg, &resp, true)
}

// IceburgOrder
func (ok *Okx) PlaceIceburgOrder(ctx context.Context, arg *IceburgOrder) ([]AlgoOrder, error) {
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodPost, algoTradeOrder, arg, &resp, true)
}

// PlaceTWAPOrder
func (ok *Okx) PlaceTWAPOrder(ctx context.Context, arg *TWAPOrderRequestParams) ([]AlgoOrder, error) {
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	if arg.PriceRatio == "" || arg.PriceVariance == "" {
		return nil, fmt.Errorf("missing price ratio or price variance parameter")
	} else if arg.AverageAmount <= 0 {
		return nil, errInvalidAverageAmountParameterValue
	} else if arg.PriceLimit <= 0 {
		return nil, errInvalidPriceLimit
	} else if ok.GetIntervalEnum(arg.Timeinterval) == "" {
		return nil, errMissingIntervalValue
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodPost, algoTradeOrder, arg, &resp, true)
}

// TriggerAlogOrder  fetches algo trigger orders for SWAP market types.
func (ok *Okx) TriggerAlogOrder(ctx context.Context, arg *TriggerAlogOrderParams) (*AlgoOrder, error) {
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	if arg.TriggerPrice <= 0 {
		return nil, errInvalidTriggerPrice
	}
	if !(strings.EqualFold(arg.TriggerPriceType, "last") ||
		strings.EqualFold(arg.TriggerPriceType, "index") ||
		strings.EqualFold(arg.TriggerPriceType, "mark")) {
		arg.TriggerPriceType = ""
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodPost, algoTradeOrder, arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}
	return nil, errors.New(resp.Msg)
}

// CancelAdvanceAlgoOrder Cancel unfilled algo orders
// A maximum of 10 orders can be canceled at a time.
// Request parameters should be passed in the form of an array.
func (ok *Okx) CancelAdvanceAlgoOrder(ctx context.Context, args []AlgoOrderCancelParams) ([]AlgoOrder, error) {
	return ok.cancelAlgoOrder(ctx, args, canceAdvancedAlgoOrder)
}

// CancelAlgoOrder to cancel unfilled algo orders (not including Iceberg order, TWAP order, Trailing Stop order).
// A maximum of 10 orders can be canceled at a time.
// Request parameters should be passed in the form of an array.
func (ok *Okx) CancelAlgoOrder(ctx context.Context, args []AlgoOrderCancelParams) ([]AlgoOrder, error) {
	return ok.cancelAlgoOrder(ctx, args, cancelAlgoOrder)
}

// cancelAlgoOrder to cancel unfilled algo orders.
func (ok *Okx) cancelAlgoOrder(ctx context.Context, args []AlgoOrderCancelParams, route string) ([]AlgoOrder, error) {
	for x := range args {
		arg := args[x]
		if arg.AlgoOrderID == "" {
			return nil, errMssingAlgoOrderID
		} else if arg.InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
	}
	if len(args) == 0 {
		return nil, errors.New("missing important arguments")
	}
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []AlgoOrder `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAlgoOrderEPL, http.MethodPost, route, args, &resp, true)
}

// GetAlgoOrderList retrieve a list of untriggered Algo orders under the current account.
func (ok *Okx) GetAlgoOrderList(ctx context.Context, orderType, algoOrderID, instrumentType, instrumentID string, after, before time.Time, limit uint) ([]AlgoOrderResponse, error) {
	params := url.Values{}
	if !(strings.EqualFold(orderType, "conditional") || strings.EqualFold(orderType, "oco") || strings.EqualFold(orderType, "trigger") || strings.EqualFold(orderType, "move_order_stop") ||
		strings.EqualFold(orderType, "iceberg") || strings.EqualFold(orderType, "twap")) {
		return nil, fmt.Errorf("invalid order type value %s,%s,%s,%s,%s,and %s", "conditional", "oco", "trigger", "move_order_stop", "iceberg", "twap")
	} else {
		params.Set("ordType", orderType)
	}
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []AlgoOrderResponse `json:"data"`
	}
	var resp response
	if algoOrderID != "" {
		params.Set("algoId", algoOrderID)
	}
	if strings.EqualFold(instrumentType, OkxInstTypeSpot) || strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeFutures) || strings.EqualFold(instrumentType, OkxInstTypeOption) {
		params.Set("instType", instrumentType)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if !before.IsZero() && before.Before(time.Now()) {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() && after.After(time.Now()) {
		params.
			Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(getAlgoOrders, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAlgoOrderListEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAlgoOrderHistory load a list of all algo orders under the current account in the last 3 months.
func (ok *Okx) GetAlgoOrderHistory(ctx context.Context, orderType, state, algoOrderID, instrumentType, instrumentID string, after, before time.Time, limit uint) ([]AlgoOrderResponse, error) {
	params := url.Values{}
	if !(strings.EqualFold(orderType, "conditional") || strings.EqualFold(orderType, "oco") || strings.EqualFold(orderType, "trigger") || strings.EqualFold(orderType, "move_order_stop") ||
		strings.EqualFold(orderType, "iceberg") || strings.EqualFold(orderType, "twap")) {
		return nil, fmt.Errorf("invalid order type value %s,%s,%s,%s,%s,and %s", "conditional", "oco", "trigger", "move_order_stop", "iceberg", "twap")
	} else {
		params.Set("ordType", orderType)
	}
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []AlgoOrderResponse `json:"data"`
	}
	var resp response
	if algoOrderID == "" &&
		!(strings.EqualFold(state, "effective") ||
			strings.EqualFold(state, "order_failed") ||
			strings.EqualFold(state, "canceled")) {
		return nil, errMissingEitherAlgoIDOrState
	} else {
		if algoOrderID != "" {
			params.Set("algoId", algoOrderID)
		} else {
			params.Set("state", state)
		}
	}
	if strings.EqualFold(instrumentType, OkxInstTypeSpot) || strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeFutures) || strings.EqualFold(instrumentType, OkxInstTypeOption) {
		params.Set("instType", instrumentType)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if !before.IsZero() && before.Before(time.Now()) {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() && after.After(time.Now()) {
		params.
			Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(algoOrderHistory, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAlgoOrderhistoryEPL, http.MethodGet, path, nil, &resp, true)
}

/*************************************** Block trading ********************************/

// GetCounterparties retrieves the list of counterparties that the user is permissioned to trade with.
func (ok *Okx) GetCounterparties(ctx context.Context) ([]CounterpartiesResponse, error) {
	type response struct {
		Code string                   `json:"code"`
		Msg  string                   `json:"msg"`
		Data []CounterpartiesResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCounterpartiesEPL, http.MethodGet, rfqCounterparties, nil, &resp, true)
}

// CreateRFQ Creates a new RFQ
func (ok *Okx) CreateRFQ(ctx context.Context, arg CreateRFQInput) (*RFQResponse, error) {
	if len(arg.CounterParties) == 0 {
		return nil, errInvalidCounterParties
	}
	if len(arg.Legs) == 0 {
		return nil, errInvalidLegs
	}
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []RFQResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, createRfqEPL, http.MethodPost, rfqCreateRFQ, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// CancelRFQ Cancel an existing active RFQ that you has previously created.
func (ok *Okx) CancelRFQ(ctx context.Context, arg CancelRFQRequestParam) (*CancelRFQResponse, error) {
	if arg.RfqID == "" && arg.ClientSuppliedRFQID == "" {
		return nil, errMissingRFQIDANDClientSuppliedRFQID
	}
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []CancelRFQResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelRfqEPL, http.MethodPost, rfqCancelRfq, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// CancelMultipleRFQs cancel multiple active RFQs in a single batch. Maximum 100 RFQ orders can be canceled at a time.
func (ok *Okx) CancelMultipleRFQs(ctx context.Context, arg CancelRFQRequestsParam) ([]CancelRFQResponse, error) {
	if len(arg.RfqID) == 0 && len(arg.ClientSuppliedRFQID) == 0 {
		return nil, errMissingRFQIDANDClientSuppliedRFQID
	}
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []CancelRFQResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleRfqEPL, http.MethodPost, rfqCancelRfqs, &arg, &resp, true)
}

// CancelAllRFQs cancels all active RFQs.
func (ok *Okx) CancelAllRFQs(ctx context.Context) (time.Time, error) {
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []TimestampResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAllRfqsEPL, http.MethodPost, rfqCancelAllRfqs, nil, &resp, true)
	if er != nil {
		return time.Time{}, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return time.Time{}, errNoValidResponseFromServer
		}
		return time.Time{}, errors.New(resp.Msg)
	}
	return resp.Data[0].Timestamp, nil
}

// ExecuteQuote ecxecutes a Quote. It is only used by the creator of the RFQ
func (ok *Okx) ExecuteQuote(ctx context.Context, arg ExecuteQuoteParams) (*ExecuteQuoteResponse, error) {
	if arg.RfqID == "" || arg.QuoteID == "" {
		return nil, errMissingRfqIDOrQuoteID
	}
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []ExecuteQuoteResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, executeQuoteEPL, http.MethodPost, rfqExecuteQuote, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// CreateQuote llows the user to Quote an RFQ that they are a counterparty to. The user MUST quote
// the entire RFQ and not part of the legs or part of the quantity. Partial quoting or partial fills are not allowed.
func (ok *Okx) CreateQuote(ctx context.Context, arg CreateQuoteParams) (*QuoteResponse, error) {
	if arg.RfqID == "" {
		return nil, errMissingRfqID
	} else if !(strings.EqualFold(arg.QuoteSide.String(), order.Buy.String()) || strings.EqualFold(arg.QuoteSide.String(), order.Sell.String())) {
		return nil, errMissingOrderSide
	} else if len(arg.Legs) == 0 {
		return nil, errMissingLegs
	}
	for x := range arg.Legs {
		if arg.Legs[x].InstrumentID == "" {
			return nil, errMissingInstrumentID
		} else if arg.Legs[x].SizeOfQuoteLeg <= 0 {
			return nil, errMissingSizeOfQuote
		} else if arg.Legs[x].Price <= 0 {
			return nil, errMossingLegsQuotePrice
		} else if arg.Legs[x].Side == order.UnknownSide {
			return nil, errMissingOrderSide
		}
	}
	type response struct {
		Code string          `json:"code"`
		Msg  string          `json:"msg"`
		Data []QuoteResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, createQuoteEPL, http.MethodPost, rfqCreateQuote, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// CancelQuote cancels an existing active quote you have created in response to an RFQ.
// rfqCancelQuote = "rfq/cancel-quote"
func (ok *Okx) CancelQuote(ctx context.Context, arg CancelQuoteRequestParams) (*CancelQuoteResponse, error) {
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []CancelQuoteResponse `json:"data"`
	}
	var resp response
	if arg.ClientSuppliedQuoteID == "" && arg.QuoteID == "" {
		return nil, errMissingQuoteIDOrClientSuppliedQuoteID
	}
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelQuoteEPL, http.MethodPost, rfqCancelQuote, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// CancelMultipleQuote cancel multiple active Quotes in a single batch. Maximum 100 quote orders can be canceled at a time.
func (ok *Okx) CancelMultipleQuote(ctx context.Context, arg CancelQuotesRequestParams) ([]CancelQuoteResponse, error) {
	if len(arg.QuoteIDs) == 0 && len(arg.ClientSuppliedQuoteIDs) == 0 {
		return nil, errMissingEitherQuoteIDAOrlientSuppliedQuoteIDs
	}
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []CancelQuoteResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleQuotesEPL, http.MethodPost, rfqCancelBatchQuotes, &arg, &resp, true)
}

// CancelAllQuotes cancels all active Quotes.
func (ok *Okx) CancelAllQuotes(ctx context.Context) (time.Time, error) {
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []TimestampResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAllQuotesEPL, http.MethodPost, rfqCancelAllQuotes, nil, &resp, true)
	if er != nil {
		return time.Time{}, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return time.Time{}, errNoValidResponseFromServer
		}
		return time.Time{}, errors.New(resp.Msg)
	}
	return resp.Data[0].Timestamp, nil
}

// GetRfqs retrieves details of RFQs that the user is a counterparty to (either as the creator or the receiver of the RFQ).
func (ok *Okx) GetRfqs(ctx context.Context, arg RfqRequestParams) ([]RFQResponse, error) {
	params := url.Values{}
	if arg.RfqID != "" {
		params.Set("rfqId", arg.RfqID)
	}
	if arg.ClientSuppliedRfqID != "" {
		params.Set("clRfqId", arg.ClientSuppliedRfqID)
	}
	if strings.EqualFold(arg.State, "active") ||
		strings.EqualFold(arg.State, "canceled") ||
		strings.EqualFold(arg.State, "pending_fill") ||
		strings.EqualFold(arg.State, "filled") ||
		strings.EqualFold(arg.State, "expired") ||
		strings.EqualFold(arg.State, "traded_away") ||
		strings.EqualFold(arg.State, "failed") ||
		strings.EqualFold(arg.State, "traded_away") ||
		strings.EqualFold(arg.State, "traded_away") {
		params.Set("state", strings.ToLower(strings.Trim(arg.State, " ")))
	}
	if arg.BeginingID != "" {
		params.Set("beginId", arg.BeginingID)
	}
	if arg.EndID != "" {
		params.Set("endId", arg.EndID)
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(int(arg.Limit)))
	}

	path := common.EncodeURLValues(rfqRfqs, params)
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []RFQResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getRfqsEPL, http.MethodGet, path, nil, &resp, true)
}

// GetQuotes retrieve all Quotes that the user is a counterparty to (either as the creator or the receiver).
func (ok *Okx) GetQuotes(ctx context.Context, arg QuoteRequestParams) ([]QuoteResponse, error) {
	params := url.Values{}
	if arg.RfqID != "" {
		params.Set("rfqId", arg.RfqID)
	}
	if arg.ClientSuppliedRfqID != "" {
		params.Set("clRfqId", arg.ClientSuppliedRfqID)
	}
	if arg.QuoteID != "" {
		params.Set("quoteId", arg.QuoteID)
	}
	if arg.ClientSuppliedQuoteID != "" {
		params.Set("clQuoteId", arg.ClientSuppliedQuoteID)
	}
	if strings.EqualFold(arg.State, "active") ||
		strings.EqualFold(arg.State, "canceled") ||
		strings.EqualFold(arg.State, "pending_fill") ||
		strings.EqualFold(arg.State, "filled") ||
		strings.EqualFold(arg.State, "expired") ||
		strings.EqualFold(arg.State, "failed") {
		params.Set("state", strings.ToLower(strings.Trim(arg.State, " ")))
	}
	if arg.BeginID != "" {
		params.Set("beginId", arg.BeginID)
	}
	if arg.EndID != "" {
		params.Set("endId", arg.EndID)
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(int(arg.Limit)))
	}
	type response struct {
		Code string          `json:"code"`
		Msg  string          `json:"msg"`
		Data []QuoteResponse `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(rfqQuotes, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getQuotesEPL, http.MethodGet, path, nil, &resp, true)
}

// GetTrades retrieves the executed trades that the user is a counterparty to (either as the creator or the receiver).
func (ok *Okx) GetRFQTrades(ctx context.Context, arg RFQTradesRequestParams) ([]RfqTradeResponse, error) {
	params := url.Values{}
	if arg.RfqID != "" {
		params.Set("rfqId", arg.RfqID)
	}
	if arg.ClientSuppliedRfqID != "" {
		params.Set("clRfqId", arg.ClientSuppliedRfqID)
	}
	if arg.QuoteID != "" {
		params.Set("quoteId", arg.QuoteID)
	}
	if arg.ClientSuppliedQuoteID != "" {
		params.Set("clQuoteId", arg.ClientSuppliedQuoteID)
	}
	if strings.EqualFold(arg.State, "active") ||
		strings.EqualFold(arg.State, "canceled") ||
		strings.EqualFold(arg.State, "pending_fill") ||
		strings.EqualFold(arg.State, "filled") ||
		strings.EqualFold(arg.State, "expired") ||
		strings.EqualFold(arg.State, "traded_away") ||
		strings.EqualFold(arg.State, "failed") {
		params.Set("state", strings.ToLower(strings.Trim(arg.State, " ")))
	}
	if arg.BlockTradeID != "" {
		params.Set("blockTdId", arg.BlockTradeID)
	}
	if arg.BeginID != "" {
		params.Set("beginId", arg.BeginID)
	}
	if arg.EndID != "" {
		params.Set("endId", arg.EndID)
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(int(arg.Limit)))
	}
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []RfqTradeResponse `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(rfqTrades, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesEPL, http.MethodGet, path, nil, &resp, true)
}

// GetPublicTrades retrieves the recent executed block trades.
func (ok *Okx) GetPublicTrades(ctx context.Context, beginID, endID string, limit int) ([]PublicTradesResponse, error) {
	params := url.Values{}
	if beginID != "" {
		params.Set("beginId", beginID)
	}
	if endID != "" {
		params.Set("endId", endID)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	path := common.EncodeURLValues(rfqPublicTrades, params)
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []PublicTradesResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPublicTradesEPL, http.MethodGet, path, nil, &resp, true)
}

/*************************************** Funding Tradings ********************************/

// GetCurrencies Retrieve a list of all currencies.
func (ok *Okx) GetCurrencies(ctx context.Context) ([]CurrencyResponse, error) {
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []CurrencyResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCurrenciesEPL, http.MethodGet, assetCurrencies, nil, &resp, true)
}

// GetBalance retrieve the balances of all the assets and the amount that is available or on hold.
func (ok *Okx) GetBalance(ctx context.Context, currency string) ([]AssetBalance, error) {
	type response struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data []AssetBalance `json:"data"`
	}
	var resp response
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(assetBalance, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBalanceEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAccountAssetValuation view account asset valuation
func (ok *Okx) GetAccountAssetValuation(ctx context.Context, currency string) ([]AccountAssetValuation, error) {
	params := url.Values{}
	if strings.EqualFold(currency, "BTC") || strings.EqualFold(currency, "USDT") ||
		strings.EqualFold(currency, "USD") || strings.EqualFold(currency, "CNY") ||
		strings.EqualFold(currency, "JPY") || strings.EqualFold(currency, "KRW") ||
		strings.EqualFold(currency, "RUB") || strings.EqualFold(currency, "EUR") ||
		strings.EqualFold(currency, "VND") || strings.EqualFold(currency, "IDR") ||
		strings.EqualFold(currency, "INR") || strings.EqualFold(currency, "PHP") ||
		strings.EqualFold(currency, "THB") || strings.EqualFold(currency, "TRY") ||
		strings.EqualFold(currency, "AUD") || strings.EqualFold(currency, "SGD") ||
		strings.EqualFold(currency, "ARS") || strings.EqualFold(currency, "SAR") ||
		strings.EqualFold(currency, "AED") || strings.EqualFold(currency, "IQD") {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(assetValuation, params)
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []AccountAssetValuation `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountAssetValuationEPL, http.MethodGet, path, nil, &resp, true)
}

// FundingTransfer transfer of funds between your funding account and trading account,
// and from the master account to sub-accounts.
func (ok *Okx) FundingTransfer(ctx context.Context, arg *FundingTransferRequestInput) ([]FundingTransferResponse, error) {
	type response struct {
		Msg  string                    `json:"msg"`
		Code string                    `json:"code"`
		Data []FundingTransferResponse `json:"data"`
	}
	var resp response
	if arg == nil {
		return nil, errors.New("argument can not be null")
	}
	if arg.Amount <= 0 {
		return nil, errors.New("invalid funding amount")
	}
	if arg.Currency == "" {
		return nil, errors.New("invalid currency value")
	}
	if !(strings.EqualFold(arg.From, "6") || strings.EqualFold(arg.From, "18")) {
		return nil, errors.New("missing funding source field \"From\"")
	}
	if arg.To == "" {
		return nil, errors.New("missing funding destination field \"To\"")
	}
	if arg.Type >= 0 && arg.Type <= 4 {
		if arg.Type == 1 || arg.Type == 2 {
			if arg.SubAccount == "" {
				return nil, errors.New("subaccount name required")
			}
		}
	} else {
		return nil, errors.New("invalid reqest type")
	}
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, fundsTransferEPL, http.MethodPost, assetTransfer, arg, &resp, true)
}

// GetFundsTransferState get funding rate response.
func (ok *Okx) GetFundsTransferState(ctx context.Context, transferID, clientID string, transferType int) ([]TransferFundRateResponse, error) {
	params := url.Values{}
	if transferID == "" && clientID == "" {
		return nil, errors.New("either 'transfer id' or 'client id' is required")
	} else if transferID != "" {
		params.Set("transId", transferID)
	} else if clientID == "" {
		params.Set("clientId", clientID)
	}
	if transferType > 0 && transferType <= 2 {
		params.Set("type", strconv.Itoa(transferType))
	}
	path := common.EncodeURLValues(assetTransferState, params)
	type response struct {
		Code string                     `json:"code"`
		Msg  string                     `json:"msg"`
		Data []TransferFundRateResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundsTransferStateEPL, http.MethodGet, path, nil, &resp, true)
}

//  AssetBillsDeta Query the billing record, you can get the latest 1 month historical data
func (ok *Okx) GetAssetBillsDetails(ctx context.Context, currency string, bill_type int, clientID string, clientSecret string, after time.Time, before time.Time, limit int) ([]AssetBillDetail, error) {
	params := url.Values{}
	billTypeIDS := []int{1, 2, 13, 20, 21, 28, 41, 42, 47, 48, 49, 50, 51, 52, 53, 54, 59, 60, 61, 68, 69, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 150, 151, 198, 199}
	for x := range billTypeIDS {
		if billTypeIDS[x] == bill_type {
			params.Set("type", strconv.Itoa(bill_type))
		}
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if clientID != "" {
		params.Set("clientId", clientID)
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	path := common.EncodeURLValues(assetbills, params)
	type response struct {
		Code string            `json:"code"`
		Msg  string            `json:"msg"`
		Data []AssetBillDetail `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, assetBillsDetailsEPL,
		http.MethodGet, path, nil, &resp, true)
}

// GetLightningDeposits users can create up to 10 thousand different invoices within 24 hours.
// this method fetches list of lightning deposits filtered by a currency and amount.
func (ok *Okx) GetLightningDeposits(ctx context.Context, currency string, amount float64, to int) ([]LightningDepositItem, error) {
	params := url.Values{}
	if currency == "" {
		return nil, errInvalidCurrencyValue
	} else {
		params.Set("ccy", currency)
	}
	if amount <= 0 {
		return nil, errInvalidDepositAmount
	} else {
		params.Set("amt", strconv.FormatFloat(amount, 'f', 0, 64))
	}
	if to == 6 || to == 18 {
		params.Set("to", strconv.Itoa(to))
	}
	path := common.EncodeURLValues(lightningDeposit, params)
	type response struct {
		Code string                 `json:"code"`
		Data []LightningDepositItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, ligntningDepositsEPL, http.MethodGet, path, nil, &resp, true)
}

// GetCurrencyDepositAddress returns the deposit address and related informations for the provided currency information.
func (ok *Okx) GetCurrencyDepositAddress(ctx context.Context, currency string) ([]CurrencyDepositResponseItem, error) {
	params := url.Values{}
	if currency == "" {
		return nil, errInvalidCurrencyValue
	} else {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(assetDeposits, params)
	type response struct {
		Code string                        `json:"code"`
		Msg  string                        `json:"msg"`
		Data []CurrencyDepositResponseItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDepositAddressEPL, http.MethodGet, path, nil, &resp, true)
}

// GetCurrencyDepositHistory retrives deposit records and withdrawal status information depending on the currency, timestamp, and chronogical order.
func (ok *Okx) GetCurrencyDepositHistory(ctx context.Context, currency string, depositID string, transactionID string, state int, after, before time.Time, limit uint) ([]DepositHistoryResponseItem, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if depositID != "" {
		params.Set("depId", depositID)
	}
	if transactionID != "" {
		params.Set("txId", transactionID)
	}
	if state == 0 ||
		state == 1 ||
		state == 2 {
		params.Set("state", strconv.Itoa(state))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(pathToAssetDepositHistory, params)
	type response struct {
		Code string                       `json:"code"`
		Msg  string                       `json:"msg"`
		Data []DepositHistoryResponseItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDepositHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// Withdrawal to perform a withdrawal action. Sub-account does not support withdrawal.
func (ok *Okx) Withdrawal(ctx context.Context, input WithdrawalInput) (*WithdrawalResponse, error) {
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []WithdrawalResponse `json:"data"`
	}
	var resp response
	if input.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if input.Amount <= 0 {
		return nil, errors.New("invalid withdrawal amount")
	} else if input.WithdrawalDestination == "" {
		return nil, errors.New("withdrawal destination")
	} else if input.ToAddress == "" {
		return nil, errors.New("missing verified digital currency address \"toAddr\" information")
	}
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, withdrawalEPL, http.MethodPost, assetWithdrawal, &input, &resp, true)
	if er != nil {
		return nil, er
	}
	return &resp.Data[0], nil
}

/*
 This API function service is only open to some users. If you need this function service, please send an email to `liz.jensen@okg.com` to apply
*/
func (ok *Okx) LightningWithdrawal(ctx context.Context, arg LightningWithdrawalRequestInput) (*LightningWithdrawalResponse, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if arg.Invoice == "" {
		return nil, errors.New("error missing invoice text")
	}
	type response struct {
		Code string                        `json:"code"`
		Msg  string                        `json:"msg"`
		Data []LightningWithdrawalResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, lightningWithdrawalsEPL, http.MethodPost, assetLightningWithdrawal, &arg, &resp, true)
	if er != nil {
		return nil, er
	} else if er == nil && len(resp.Data) == 0 {
		return nil, errMissingResponseBody
	}
	return &resp.Data[0], nil
}

// CancelWithdrawal You can cancel normal withdrawal, but can not cancel the withdrawal on Lightning.
func (ok *Okx) CancelWithdrawal(ctx context.Context, withdrawalID string) (string, error) {
	if withdrawalID == "" {
		return "", errMissingValidWithdrawalID
	}
	type inout struct {
		WithdrawalID string `json:"wdId"` // Required
	}
	input := &inout{
		WithdrawalID: withdrawalID,
	}
	type response struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
		Data   inout  `json:"data"`
	}
	var output response
	return output.Data.WithdrawalID, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelWithdrawalEPL, http.MethodPost, cancelWithdrawal, input, &output, true)
}

// Retrieve the withdrawal records according to the currency, withdrawal status, and time range in reverse chronological order.
// The 100 most recent records are returned by default.
func (ok *Okx) GetWithdrawalHistory(ctx context.Context, currency, withdrawalID, clientID, transactionID string, state int, after, before time.Time, limit int) ([]WithdrawalHistoryResponse, error) {
	type response struct {
		Code string                      `json:"code"`
		Msg  string                      `json:"msg"`
		Data []WithdrawalHistoryResponse `json:"data"`
	}
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if withdrawalID != "" {
		params.Set("wdId", withdrawalID)
	}
	if clientID != "" {
		params.Set("clientId", clientID)
	}
	if transactionID != "" {
		params.Set("txId", transactionID)
	}
	if state == -3 || state == -2 || state == -1 || state == 0 || state == 1 || state == 2 || state == 3 || state == 4 || state == 5 {
		params.Set("state", strconv.Itoa(state))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var resp response
	path := common.EncodeURLValues(withdrawalHistory, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getWithdrawalHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// SmallAssetsConvert Convert small assets in funding account to OKB. Only one convert is allowed within 24 hours.
func (ok *Okx) SmallAssetsConvert(ctx context.Context, currency []string) (*SmallAssetConvertResponse, error) {
	type response struct {
		Code string                      `json:"code"`
		Msg  string                      `json:"msg"`
		Data []SmallAssetConvertResponse `json:"data"`
	}
	input := map[string][]string{"ccy": currency}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, smallAssetsConvertEPL, http.MethodPost, smallAssetsConvert, input, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("no response data found")
	}
	return &resp.Data[0], nil
}

// GetSavingBalance returns saving balance, and only assets in the funding account can be used for saving.
func (ok *Okx) GetSavingBalance(ctx context.Context, currency string) ([]SavingBalanceResponse, error) {
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []SavingBalanceResponse `json:"data"`
	}
	var resp response
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(assetSavingBalance, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSavingBalanceEPL, http.MethodGet, path, nil, &resp, true)
}

// SavingsPurchase posts a purchase
func (ok *Okx) SavingsPurchase(ctx context.Context, arg *SavingsPurchaseRedemptionInput) (*SavingsPurchaseRedemptionResponse, error) {
	if arg == nil {
		return nil, errors.New("invalid argument, null argument")
	}
	arg.ActionType = "purchase" // purchase: purchase saving shares
	return ok.savingsPurchaseOrRedemption(ctx, arg)
}

// SavingsRedemption posts a redemption
func (ok *Okx) SavingsRedemption(ctx context.Context, arg *SavingsPurchaseRedemptionInput) (*SavingsPurchaseRedemptionResponse, error) {
	if arg == nil {
		return nil, errors.New("invalid argument, null argument")
	}
	arg.ActionType = "redempt" // redempt: redeem saving shares
	return ok.savingsPurchaseOrRedemption(ctx, arg)
}

func (ok *Okx) savingsPurchaseOrRedemption(ctx context.Context, arg *SavingsPurchaseRedemptionInput) (*SavingsPurchaseRedemptionResponse, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if arg.Amount <= 0 {
		return nil, errUnacceptableAmount
	} else if !(strings.EqualFold(arg.ActionType, "purchase") || strings.EqualFold(arg.ActionType, "redempt")) {
		return nil, errors.New("invalid side value, side has to be either \"redemption\" or \"purchase\"")
	} else if strings.EqualFold(arg.ActionType, "purchase") && !(arg.Rate >= 1 && arg.Rate <= 365) {
		return nil, errors.New("the rate value range is between 1% and 365%")
	}
	type response struct {
		Code string                              `json:"code"`
		Msg  string                              `json:"msg"`
		Data []SavingsPurchaseRedemptionResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, savingsPurchaseRedemptionEPL, http.MethodPost, assetSavingPurchaseOrRedemptionPath, &arg, &resp, true)
	if er != nil {
		return nil, er
	} else if len(resp.Data) == 0 {
		return nil, errors.New("no response from the server")
	}
	return &resp.Data[0], nil
}

// GetLendingHistory lending history
func (ok *Okx) GetLendingHistory(ctx context.Context, currency string, before, after time.Time, limit uint) ([]LendingHistory, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(assetsLendingHistory, params)
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []LendingHistory `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, setLendingRateEPL, http.MethodGet, path, nil, &resp, true)
}

// SetLendingRate sets assets Lending Rate
func (ok *Okx) SetLendingRate(ctx context.Context, arg LendingRate) (*LendingRate, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if !(arg.Rate >= 1 && arg.Rate <= 365) {
		return nil, errors.New("invalid lending rate value. the rate value range is between 1% and 365%")
	}
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []LendingRate `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLendinghistoryEPL, http.MethodPost, assetSetLendingRateRoute, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		val, _ := json.Marshal(resp)
		return nil, errors.New(string(val))
	}
	return &resp.Data[0], nil
}

// GetPublicBorrowInfo return list of publix borrow history.
func (ok *Okx) GetPublicBorrowInfo(ctx context.Context, currency string) ([]PublicBorrowInfo, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(publicBorrowInfo, params)
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []PublicBorrowInfo `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPublicBorrowInfoEPL, http.MethodGet, path, nil, &resp, false)
}

/***********************************Convert Endpoints | Authenticated s*****************************************/

func (ok *Okx) GetConvertCurrencies(ctx context.Context) ([]ConvertCurrency, error) {
	type response struct {
		Code string            `json:"code"`
		Msg  string            `json:"msg"`
		Data []ConvertCurrency `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertCurrenciesEPL, http.MethodGet, assetConvertCurrencies, nil, &resp, true)
}

// GetConvertCurrencyPair
func (ok *Okx) GetConvertCurrencyPair(ctx context.Context, fromCurrency, toCurrency string) (*ConvertCurrencyPair, error) {
	params := url.Values{}
	if fromCurrency == "" {
		return nil, errors.New("missing reference currency name \"fromCcy\"")
	}
	if toCurrency == "" {
		return nil, errors.New("missing destination currency name \"toCcy\"")
	}
	params.Set("fromCcy", fromCurrency)
	params.Set("toCcy", toCurrency)
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []ConvertCurrencyPair `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(convertCurrencyPairsPath, params)
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertCurrencyPairEPL, http.MethodGet, path, nil, &resp, true); er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("invalid response from server")
	}
	return &resp.Data[0], nil
}

// EstimateQuote
func (ok *Okx) EstimateQuote(ctx context.Context, arg EstimateQuoteRequestInput) (*EstimateQuoteResponse, error) {
	if arg.BaseCurrency == "" {
		return nil, errors.New("missing base currency")
	}
	if arg.QuoteCurrency == "" {
		return nil, errors.New("missing quote currency")
	}
	if !(strings.EqualFold(arg.Side, "buy") || strings.EqualFold(arg.Side, "sell")) {
		return nil, errors.New("missing  order side")
	}
	arg.Side = strings.ToLower(arg.Side)
	if arg.RFQAmount <= 0 {
		return nil, errors.New("missing rfq amount")
	}
	if arg.RFQSzCurrency == "" {
		return nil, errors.New("missing rfq currency")
	}
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []EstimateQuoteResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, estimateQuoteEPL, http.MethodPost, assetEstimateQuote, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// ConvertTrade converts a base currency to quote currency.
func (ok *Okx) ConvertTrade(ctx context.Context, arg ConvertTradeInput) (*ConvertTradeResponse, error) {
	if arg.BaseCurrency == "" {
		return nil, errors.New("missing base currency")
	}
	if arg.QuoteCurrency == "" {
		return nil, errors.New("missing quote currency")
	}
	if !(strings.EqualFold(arg.Side, "buy") ||
		strings.EqualFold(arg.Side, "sell")) {
		return nil, errMissingOrderSide
	}
	if arg.Size <= 0 {
		return nil, errors.New("quote amount should be more than 0 and RFQ amount")
	}
	if arg.SizeCurrency == "" {
		return nil, errors.New("missing size currency")
	}
	if arg.QuoteID == "" {
		return nil, errors.New("missing quote id")
	}
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []ConvertTradeResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, convertTradeEPL, http.MethodPost, assetConvertTrade, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// GetConvertHistory gets the recent history.
func (ok *Okx) GetConvertHistory(ctx context.Context, before, after time.Time, limit uint, tag string) ([]ConvertHistory, error) {
	params := url.Values{}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	if tag != "" {
		params.Set("tag", tag)
	}
	path := common.EncodeURLValues(assetConvertHistory, params)
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []ConvertHistory `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

/********************************** Account endpoints ***************************************************/

// GetBalance retrieve a list of assets (with non-zero balance), remaining balance, and available amount in the trading account.
// Interest-free quota and discount rates are public data and not displayed on the account interface.
func (ok *Okx) GetNonZeroBalances(ctx context.Context, currency string) ([]Account, error) {
	type response struct {
		Code string    `json:"code"`
		Msg  string    `json:"msg"`
		Data []Account `json:"data"`
	}
	var resp response
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(accountBalance, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountBalanceEPL, http.MethodGet, path, nil, &resp, true)
}

// GetPositions retrieve information on your positions. When the account is in net mode, net positions will be displayed, and when the account is in long/short mode, long or short positions will be displayed.
func (ok *Okx) GetPositions(ctx context.Context, instrumentType, instrumentID, positionID string) ([]AccountPosition, error) {
	params := url.Values{}
	if instrumentType != "" {
		params.Set("instType", instrumentType)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if positionID != "" {
		params.Set("posId", positionID)
	}
	path := common.EncodeURLValues(accountPosition, params)
	type response struct {
		Code string            `json:"code"`
		Msg  string            `json:"msg"`
		Data []AccountPosition `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionsEPL, http.MethodGet, path, nil, &resp, true)
}

// GetPositionsHistory  retrieve the updated position data for the last 3 months.
func (ok *Okx) GetPositionsHistory(ctx context.Context, instrumentType, instrumentID, marginMode string, closePositionType uint, after, before time.Time, limit uint) ([]AccountPositionHistory, error) {
	params := url.Values{}
	if strings.EqualFold(instrumentType, OkxInstTypeMargin) || strings.EqualFold(instrumentType, OkxInstTypeSwap) || strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeOption) {
		params.Set("instType", strings.ToUpper(instrumentType))
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if strings.EqualFold(marginMode, "cross") || strings.EqualFold(marginMode, "isolated") {
		params.Set("mgnMode", marginMode)
	}
	// The type of closing position
	// 1：Close position partially;2：Close all;3：Liquidation;4：Partial liquidation; 5：ADL;
	// It is the latest type if there are several types for the same position.
	if closePositionType >= 1 && closePositionType <= 5 {
		params.Set("type", strconv.Itoa(int(closePositionType)))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(accountPositionHistory, params)
	type response struct {
		Code string                   `json:"code"`
		Msg  string                   `json:"msg"`
		Data []AccountPositionHistory `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionsHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAccountAndPositionRisk  get account and position risks.
func (ok *Okx) GetAccountAndPositionRisk(ctx context.Context, instrumentType string) ([]AccountAndPositionRisk, error) {
	params := url.Values{}
	if instrumentType != "" {
		params.Set("instType", instrumentType)
	}
	path := common.EncodeURLValues(accountAndPositionRisk, params)
	type response struct {
		Code string                   `json:"code"`
		Msg  string                   `json:"msg"`
		Data []AccountAndPositionRisk `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountAndPositionRiskEPL, http.MethodGet, path, nil, &resp, true)
}

// GetBillsDetailLast7Days The bill refers to all transaction records that result in changing the balance of an account. Pagination is supported, and the response is sorted with the most recent first. This endpoint can retrieve data from the last 7 days.
func (ok *Okx) GetBillsDetailLast7Days(ctx context.Context, arg BillsDetailQueryParameter) ([]BillsDetailResponse, error) {
	return ok.GetBillsDetail(ctx, arg, accountBillsDetail)
}

// GetBillsDetail3Months retrieve the account’s bills.
// The bill refers to all transaction records that result in changing the balance of an account.
// Pagination is supported, and the response is sorted with most recent first.
// This endpoint can retrieve data from the last 3 months.
func (ok *Okx) GetBillsDetail3Months(ctx context.Context, arg BillsDetailQueryParameter) ([]BillsDetailResponse, error) {
	return ok.GetBillsDetail(ctx, arg, accountBillsDetailArchive)
}

// GetBillsDetail retrieve the bills of the account.
func (ok *Okx) GetBillsDetail(ctx context.Context, arg BillsDetailQueryParameter, route string) ([]BillsDetailResponse, error) {
	params := url.Values{}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) || strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) || strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeOption) || strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) {
		params.Set("instType", strings.ToUpper(arg.InstrumentType))
	}
	if arg.Currency != "" {
		params.Set("ccy", strings.ToUpper(arg.Currency))
	}
	if strings.EqualFold(arg.MarginMode, "isolated") || strings.EqualFold(arg.MarginMode, "cross") {
		params.Set("mgnMode", strings.ToLower(arg.MarginMode))
	}
	if strings.EqualFold(arg.ContractType, "linear") || strings.EqualFold(arg.ContractType, "inverse") {
		params.Set("ctType", arg.ContractType)
	}
	if arg.BillType >= 1 && arg.BillType <= 13 {
		params.Set("type", strconv.Itoa(int(arg.BillType)))
	}
	billSubtypes := []int{1, 2, 3, 4, 5, 6, 9, 11, 12, 14, 160, 161, 162, 110, 111, 118, 119, 100, 101, 102, 103, 104, 105, 106, 110, 125, 126, 127, 128, 131, 132, 170, 171, 172, 112, 113, 117, 173, 174, 200, 201, 202, 203}
	for x := range billSubtypes {
		if int(arg.BillSubType) == billSubtypes[x] {
			params.Set("subType", strconv.Itoa(int(arg.BillSubType)))
		}
	}
	if arg.After != "" {
		params.Set("after", arg.After)
	}
	if arg.Before != "" {
		params.Set("before", arg.Before)
	}
	if !arg.BeginTime.IsZero() {
		params.Set("begin", strconv.FormatInt(arg.BeginTime.UnixMilli(), 10))
	}
	if !arg.EndTime.IsZero() {
		params.Set("end", strconv.FormatInt(arg.EndTime.UnixMilli(), 10))
	}
	if int(arg.Limit) > 0 {
		params.Set("limit", strconv.Itoa(int(arg.Limit)))
	}
	path := common.EncodeURLValues(route, params)
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []BillsDetailResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBillsDetailsEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAccountConfiguration retrieve current account configuration.
func (ok *Okx) GetAccountConfiguration(ctx context.Context) ([]AccountConfigurationResponse, error) {
	type response struct {
		Code string                         `json:"code"`
		Msg  string                         `json:"msg"`
		Data []AccountConfigurationResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountConfigurationEPL, http.MethodGet, accountConfiguration, nil, &resp, true)
}

// SetPositionMode FUTURES and SWAP support both long/short mode and net mode. In net mode, users can only have positions in one direction; In long/short mode, users can hold positions in long and short directions.
func (ok *Okx) SetPositionMode(ctx context.Context, positionMode string) (string, error) {
	if !(strings.EqualFold(positionMode, "long_short_mode") || strings.EqualFold(positionMode, "net_mode")) {
		return "", errors.New("missing position mode, \"long_short_mode\": long/short, only applicable to FUTURES/SWAP \"net_mode\": net")
	}
	input := &PositionMode{
		PositionMode: positionMode,
	}
	type response struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data []PositionMode `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, setPositionModeEPL, http.MethodPost, accountSetPositionMode, input, &resp, true)
	if er != nil {
		return "", er
	}
	if len(resp.Data) == 0 {
		return "", errors.New(resp.Msg)
	}
	return resp.Data[0].PositionMode, nil
}

// SetLeverage
func (ok *Okx) SetLeverage(ctx context.Context, arg SetLeverageInput) (*SetLeverageResponse, error) {
	if arg.InstrumentID == "" && arg.Currency == "" {
		return nil, errors.New("either instrument id or currency is missing")
	}
	if arg.Leverage < 0 {
		return nil, errors.New("missing leverage")
	}
	if !(strings.EqualFold(arg.MarginMode, "isolated") || strings.EqualFold(arg.MarginMode, "cross")) {
		return nil, errors.New("invalid margin mode")
	}
	if arg.InstrumentID == "" && strings.EqualFold(arg.MarginMode, "isolated") {
		return nil, errors.New("only can be cross if ccy is passed")
	}
	if !(strings.EqualFold(arg.MarginMode, "cross") || strings.EqualFold(arg.MarginMode, "isolated")) {
		return nil, errors.New("only applicable to \"isolated\" margin mode of FUTURES/SWAP")
	}
	if !(strings.EqualFold(arg.PositionSide, "long") ||
		strings.EqualFold(arg.PositionSide, "short")) && strings.EqualFold(arg.MarginMode, "isolated") {
		return nil, errors.New("\"long\" \"short\" Only applicable to isolated margin mode of FUTURES/SWAP")
	} else if !strings.EqualFold(arg.MarginMode, "isolated") {
		arg.MarginMode = ""
	}
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []SetLeverageResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, setLeverateEPL, http.MethodPost, accountSetLeverage, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetMaximumBuySellAmountOROpenAmount
func (ok *Okx) GetMaximumBuySellAmountOROpenAmount(ctx context.Context, instrumentID, tradeMode, currency, leverage string, price float64) ([]MaximumBuyAndSell, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errors.New("missing instrument id")
	}
	params.Set("instId", instrumentID)
	if !(strings.EqualFold(tradeMode, "cross") || strings.EqualFold(tradeMode, "isolated") || strings.EqualFold(tradeMode, "cash")) {
		return nil, errors.New("missing valid trade mode")
	}
	params.Set("tdMode", tradeMode)
	if currency != "" {
		params.Set("ccy", currency)
	}
	if price > 0 {
		params.Set("px", strconv.Itoa(int(price)))
	}
	if leverage != "" {
		params.Set("leverage", leverage)
	}
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []MaximumBuyAndSell `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(accountMaxSize, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumBuyOrSellAmountEPL, http.MethodGet, path, nil, &resp, true)
}

// GetMaximumAvailableTradableAmount
func (ok *Okx) GetMaximumAvailableTradableAmount(ctx context.Context, instrumentID, currency, tradeMode string, reduceOnly bool, price float64) ([]MaximumTradableAmount, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("missing instrument id")
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !(strings.EqualFold(tradeMode, "isolated") ||
		strings.EqualFold(tradeMode, "cross") ||
		strings.EqualFold(tradeMode, "cash")) {
		return nil, errors.New("missing trade mode")
	}
	params.Set("tdMode", tradeMode)
	params.Set("px", strconv.FormatFloat(price, 'f', 0, 64))
	path := common.EncodeURLValues(accountMaxAvailSize, params)
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []MaximumTradableAmount `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumAvailableTradableAmountEPL, http.MethodGet, path, nil, &resp, true)
}

// IncreaseDecreaseMargin Increase or decrease the margin of the isolated position. Margin reduction may result in the change of the actual leverage.
func (ok *Okx) IncreaseDecreaseMargin(ctx context.Context, arg IncreaseDecreaseMarginInput) (*IncreaseDecreaseMargin, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(strings.EqualFold(arg.PositionSide, "long") || strings.EqualFold(arg.PositionSide, "short") || strings.EqualFold(arg.PositionSide, "net")) {
		return nil, errors.New("missing valid position side")
	}
	if !(strings.EqualFold(arg.Type, "add") || strings.EqualFold(arg.Type, "reduce")) {
		return nil, errors.New("missing valid 'type', use 'add': add margin 'reduce': reduce margin")
	}
	if arg.Amount <= 0 {
		return nil, errors.New("missing valid amount")
	}
	type response struct {
		Code string                   `json:"code"`
		Msg  string                   `json:"msg"`
		Data []IncreaseDecreaseMargin `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, increaseOrDecreaseMarginEPL, http.MethodGet, accountPositionMarginBalance, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetLeverage
func (ok *Okx) GetLeverage(ctx context.Context, instrumentID, marginMode string) ([]LeverageResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("invalid instrument id \"instId\" ")
	}
	if strings.EqualFold(marginMode, "cross") || strings.EqualFold(marginMode, "isolated") {
		params.Set("mgnMode", marginMode)
	} else {
		return nil, errors.New("missing margin mode \"mgnMode\"")
	}
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []LeverageResponse `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(accountLeverageInfo, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getLeverateEPL, http.MethodGet, path, nil, &resp, true)
}

// GetMaximumLoanOfInstrument returns list of maximum loan of instruments.
func (ok *Okx) GetMaximumLoanOfInstrument(ctx context.Context, instrumentID, marginMode, mgnCurrency string) ([]MaximumLoanInstrument, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("invalid instrument id \"instId\"")
	}
	if strings.EqualFold(marginMode, "cross") || strings.EqualFold(marginMode, "isolated") {
		params.Set("mgnMode", marginMode)
	} else {
		return nil, errors.New("missing margin mode \"mgnMode\"")
	}
	if mgnCurrency != "" {
		params.Set("mgnCcy", mgnCurrency)
	}
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []MaximumLoanInstrument `json:"data"`
	}
	path := common.EncodeURLValues(accountMaxLoan, params)
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTheMaximumLoanOfInstrumentEPL, http.MethodGet, path, nil, &resp, true)
}

// GetFee returns Cryptocurrency trade fee, and offline trade fee
func (ok *Okx) GetFee(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	// Here the Asset Type for the instrument Type is needed for getting the CryptocurrencyTrading Fee.
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		var responses []TradeFeeRate
		uly, er := ok.GetUnderlying(feeBuilder.Pair, asset.Spot)
		if er != nil {
			return 0, er
		}
		responses, er = ok.GetTradeFee(ctx, OkxInstTypeSpot, "", uly)
		if er != nil {
			return 0, er
		} else if len(responses) == 0 {
			return 0, errors.New("no trade fee response found")
		}
		var fee float64
		if feeBuilder.IsMaker {
			if fee, er = strconv.ParseFloat(responses[0].FeeRateMaker, 64); er != nil || fee == 0 {
				fee, er = strconv.ParseFloat(responses[0].FeeRateMakerUSDT, 64)
				if er != nil {
					return fee, er
				}
			}
		} else {
			if fee, er = strconv.ParseFloat(responses[0].FeeRateTaker, 64); er != nil || fee == 0 {
				fee, er = strconv.ParseFloat(responses[0].FeeRateTakerUSDT, 64)
				if er != nil {
					return fee, er
				}
			}
		}
		if fee < 0 {
			fee = -fee
		}
		return fee * feeBuilder.Amount * feeBuilder.PurchasePrice, nil
	case exchange.OfflineTradeFee:
		return 0.0015 * feeBuilder.PurchasePrice * feeBuilder.Amount, nil
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

// GetTradeFeeRate query trade fee rate of various instrument types and instrument ids.
func (ok *Okx) GetTradeFee(ctx context.Context, instrumentType, instrumentID, underlying string) ([]TradeFeeRate, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, OkxInstTypeSpot) || strings.EqualFold(instrumentType, OkxInstTypeMargin) || strings.EqualFold(instrumentType, OkxInstTypeSwap) || strings.EqualFold(instrumentType, OkxInstTypeFutures) || strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, errInvalidInstrumentType
	}
	params.Set("instType", strings.ToUpper(instrumentType))
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if underlying != "" {
		params.Set("uly", underlying)
	}
	path := common.EncodeURLValues(accountTradeFee, params)
	type response struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data []TradeFeeRate `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFeeRatesEPL, http.MethodGet, path, nil, &resp, true)
}

// GetInterestAccruedData account accred data.
func (ok *Okx) GetInterestAccruedData(ctx context.Context, loanType int, currency, instrumentID, marginMode string, after, before time.Time, limit int) ([]InterestAccruedData, error) {
	params := url.Values{}
	if loanType == 1 || loanType == 2 {
		params.Set("type", strconv.Itoa(loanType))
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if strings.EqualFold(marginMode, "cross") ||
		strings.EqualFold(marginMode, "isolated") {
		params.Set("mgnMode", marginMode)
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []InterestAccruedData `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(accountInterestAccured, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestAccruedDataEPL, http.MethodGet, path, nil, &resp, true)
}

// GetInterestRate get the user's current leveraged currency borrowing interest rate
func (ok *Okx) GetInterestRate(ctx context.Context, currency string) ([]InterestRateResponse, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(accountInterestRate, params)
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []InterestRateResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateEPL, http.MethodGet, path, nil, &resp, true)
}

// SetGeeks set the display type of Greeks. PA: Greeks in coins BS: Black-Scholes Greeks in dollars
func (ok *Okx) SetGeeks(ctx context.Context, greeksType string) (*GreeksType, error) {
	if !(strings.EqualFold(greeksType, "PA") || strings.EqualFold(greeksType, "BS")) {
		return nil, errMissingValidGeeskType
	}
	input := &GreeksType{
		GreeksType: greeksType,
	}
	type response struct {
		Code string       `json:"code"`
		Msg  string       `json:"msg"`
		Data []GreeksType `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, setGeeksEPL, http.MethodPost, accountSetGeeks, input, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// IsolatedMarginTradingSettings to set the currency margin and futures/perpetual Isolated margin trading mode.
func (ok *Okx) IsolatedMarginTradingSettings(ctx context.Context, arg IsolatedMode) (*IsolatedMode, error) {
	if !(strings.EqualFold(arg.IsoMode, "automatic") || strings.EqualFold(arg.IsoMode, "autonomy")) {
		return nil, errMissingIsolatedMarginTradingSetting
	}
	if !(strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) || strings.EqualFold(arg.InstrumentType, OkxInstTypeContract)) {
		return nil, errMissingInstrumentID
	}

	type response struct {
		Code string         `json:"code"`
		Msg  string         `json:"msg"`
		Data []IsolatedMode `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, isolatedMarginTradingSettingsEPL, http.MethodPost, accountSetIsolatedMode, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetMaximumWithdrawals retrieve the maximum transferable amount from trading account to funding account.
func (ok *Okx) GetMaximumWithdrawals(ctx context.Context, currency string) ([]MaximumWithdrawal, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(accountMaxWithdrawal, params)
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []MaximumWithdrawal `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumWithdrawalsEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAccountRiskState
func (ok *Okx) GetAccountRiskState(ctx context.Context) ([]AccountRiskState, error) {
	type response struct {
		Code string             `json:"code"`
		Msg  string             `json:"msg"`
		Data []AccountRiskState `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountRiskStateEPL, http.MethodGet, accountRiskState, nil, &resp, true)
}

// VIPLoansBorrowAndRepay
func (ok *Okx) VIPLoansBorrowAndRepay(ctx context.Context, arg LoanBorrowAndReplayInput) (*LoanBorrowAndReplay, error) {
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []LoanBorrowAndReplay `json:"data"`
	}
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	}
	if arg.Side == "" {
		return nil, errInvalidOrderSide
	}
	if arg.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, vipLoansBorrowAnsRepayEPL, http.MethodPost, accountBorrowReply, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetBorrowAndRepayHistoryForVIPLoans retrives borrow and repay history for VIP loans.
func (ok *Okx) GetBorrowAndRepayHistoryForVIPLoans(ctx context.Context, currency string, after, before time.Time, limit uint) ([]BorrowRepayHistory, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(accountBorrowRepayHistory, params)
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []BorrowRepayHistory `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBorrowAnsRepayHistoryHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// GetBorrowInterestAndLimit borrow interest and limit
func (ok *Okx) GetBorrowInterestAndLimit(ctx context.Context, loanType int, currency string) ([]BorrowInterestAndLimitResponse, error) {
	params := url.Values{}
	if loanType == 1 || loanType == 2 {
		params.Set("type", strconv.Itoa(loanType))
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	type response struct {
		Code string                           `json:"code"`
		Msg  string                           `json:"msg"`
		Data []BorrowInterestAndLimitResponse `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(accountInterestLimits, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBorrowInterestAndLimitEPL, http.MethodGet, path, nil, &resp, true)
}

// PositionBuilder calculates portfolio margin information for simulated position or current position of the user. You can add up to 200 simulated positions in one request.
func (ok *Okx) PositionBuilder(ctx context.Context, arg PositionBuilderInput) ([]PositionBuilderResponse, error) {
	if !(strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) || strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) || strings.EqualFold(arg.InstrumentType, "OPTIONS")) {
		arg.InstrumentType = ""
	}
	type response struct {
		Code string                    `json:"code"`
		Msg  string                    `json:"msg"`
		Data []PositionBuilderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, positionBuilderEPL, http.MethodPost, accountSimulatedMargin, &arg, &resp, true)
}

// GetGreeks retrieve a greeks list of all assets in the account.
func (ok *Okx) GetGreeks(ctx context.Context, currency string) ([]GreeksItem, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	path := common.EncodeURLValues(accountGeeks, params)
	type response struct {
		Code string       `json:"code"`
		Msg  string       `json:"msg"`
		Data []GreeksItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGeeksEPL, http.MethodGet, path, nil, &resp, true)
}

/********************************** Subaccount Endpoints ***************************************************/

// ViewSubaccountList applies to master accounts only
func (ok *Okx) ViewSubaccountList(ctx context.Context, enable bool, subaccountName string, after, before time.Time, limit int) ([]SubaccountInfo, error) {
	params := url.Values{}
	params.Set("enable", strconv.FormatBool(enable))
	if subaccountName != "" {
		params.Set("subAcct", subaccountName)
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []SubaccountInfo `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(usersSubaccountList, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, viewSubaccountListEPL, http.MethodGet, path, nil, &resp, true)
}

// GetSubaccountTradingBalance query detailed balance info of Trading Account of a sub-account via the master account (applies to master accounts only)
func (ok *Okx) GetSubaccountTradingBalance(ctx context.Context, subaccountName string) ([]SubaccountBalanceResponse, error) {
	params := url.Values{}
	if subaccountName != "" {
		params.Set("subAcct", subaccountName)
	} else {
		return nil, errMissingRequiredParameterSubaccountName
	}
	path := common.EncodeURLValues(accountSubaccountBalances, params)
	type response struct {
		Code string                      `json:"code"`
		Msg  string                      `json:"msg"`
		Data []SubaccountBalanceResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSubaccountTradingBalanceEPL, http.MethodGet, path, nil, &resp, true)
}

// GetSubaccountFundingBalance query detailed balance info of Funding Account of a sub-account via the master account (applies to master accounts only)
func (ok *Okx) GetSubaccountFundingBalance(ctx context.Context, subaccountName, currency string) ([]FundingBalance, error) {
	params := url.Values{}
	if subaccountName == "" {
		return nil, errMissingRequiredParameterSubaccountName
	} else {
		params.Set("subAcct", subaccountName)
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []FundingBalance `json:"data"`
	}
	path := common.EncodeURLValues(assetSubaccountBalances, params)
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSubaccountFundingBalanceEPL, http.MethodGet, path, nil, &resp, true)
}

// HistoryOfSubaccountTransfer retrives subaccount transfer histories; applies to master accounts only.
// Retrieve the transfer data for the last 3 months.
func (ok *Okx) HistoryOfSubaccountTransfer(ctx context.Context, currency string, subaccountType uint8, subaccountName string, before, after time.Time, limit int) ([]SubaccountBillItem, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if subaccountType == 0 || subaccountType == 1 {
		params.Set("type", strconv.Itoa(int(subaccountType)))
	}
	if subaccountName != "" {
		params.Set("subacct", subaccountName)
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	path := common.EncodeURLValues(assetSubaccountBills, params)
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []SubaccountBillItem `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, historyOfSubaccountTransferEPL, http.MethodGet, path, nil, &resp, true)
}

// ManageAccountManageTransfersBetweenSubaccounts master accounts manage the transfers between sub-accounts applies to master accounts only
func (ok *Okx) MasterAccountsManageTransfersBetweenSubaccounts(ctx context.Context, currency string, amount float64, from, to uint, fromSubaccount, toSubaccount string, loanTransfer bool) ([]TransferIDInfo, error) {
	params := url.Values{}
	if currency == "" {
		return nil, errInvalidCurrencyValue
	}
	params.Set("ccy", currency)
	if amount <= 0 {
		return nil, errInvalidTransferAmount
	}
	params.Set("amt", strconv.FormatFloat(amount, 'f', 2, 64))
	if !(from == 6 || from == 18) {
		return nil, errInvalidInvalidSubaccount
	}
	params.Set("from", strconv.Itoa(int(from)))
	if !(to == 6 || to == 18) {
		return nil, errInvalidInvalidSubaccount
	}
	params.Set("to", strconv.Itoa(int(to)))
	if fromSubaccount == "" {
		return nil, errMissingInitialSubaccountName
	}
	params.Set("fromSubAccount", fromSubaccount)
	if toSubaccount == "" {
		return nil, errMissingDestinationSubaccountName
	}
	params.Set("toSubAccount", fromSubaccount)
	params.Set("loanTrans", strconv.FormatBool(loanTransfer))
	path := common.EncodeURLValues(assetSubaccountTransfer, params)
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []TransferIDInfo `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, masterAccountsManageTransfersBetweenSubaccountEPL, http.MethodGet, path, nil, &resp, true)
}

// SetPermissingOfTransfer set permission of transfer out for sub-account(only applicable to master account). Sub-account can transfer out to master account by default.
func (ok *Okx) SetPermissionOfTransferOut(ctx context.Context, arg PermissingOfTransfer) ([]PermissingOfTransfer, error) {
	if arg.SubAcct == "" {
		return nil, errMissingRequiredParameterSubaccountName
	}
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []PermissingOfTransfer `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, setPermissingOfTransferOutEPL, http.MethodPost, userSubaccountSetTransferOut, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg != "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return resp.Data, nil
}

// GetCustodyTradingSubaccountList the trading team uses this interface to view the list of sub-accounts currently under escrow
// usersEntrustSubaccountList ="users/entrust-subaccount-list"
func (ok *Okx) GetCustodyTradingSubaccountList(ctx context.Context, subaccountName string) ([]SubaccountName, error) {
	params := url.Values{}
	if subaccountName != "" {
		params.Set("setAcct", subaccountName)
	}
	type response struct {
		Code string           `json:"code"`
		Msg  string           `json:"msg"`
		Data []SubaccountName `json:"data"`
	}
	path := common.EncodeURLValues(usersEntrustSubaccountList, params)
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCustoryTradingSubaccountListEPL, http.MethodGet, path, nil, &resp, true)
}

/*************************************** Grid Trading Endpoints ***************************************************/

// PlaceGridAlgoOrder place spot grid algo order.
func (ok *Okx) PlaceGridAlgoOrder(ctx context.Context, arg GridAlgoOrder) (*GridAlgoOrderIDResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(strings.EqualFold(arg.AlgoOrdType, "grid") || strings.EqualFold(arg.AlgoOrdType, "contract_grid")) {
		return nil, errMissingAlgoOrderType
	}
	if arg.MaxPrice <= 0 {
		return nil, errInvalidMaximumPrice
	}
	if arg.MinPrice < 0 {
		return nil, errInvalidMinimumPrice
	}
	if arg.GridQuantity < 0 {
		return nil, errInvalidGridQuantity
	}
	isSpotGridOrder := false
	if arg.QuoteSize > 0 || arg.BaseSize > 0 {
		isSpotGridOrder = true
	}
	if !isSpotGridOrder {
		if arg.Size <= 0 {
			return nil, errMissingSize
		}
		if !(strings.EqualFold(arg.Direction, "long") || strings.EqualFold(arg.Direction, "short") || strings.EqualFold(arg.Direction, "neutral")) {
			return nil, errMissingRequiredArgumentDirection
		}
		if arg.Lever == "" {
			return nil, errRequiredParameterMissingLeverage
		}
	}
	type response struct {
		Code string                    `json:"code"`
		Msg  string                    `json:"msg"`
		Data []GridAlgoOrderIDResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, gridTradingEPL, http.MethodPost, gridOrderAlgo, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg != "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// AmendGridAlgoOrder supported contract grid algo order amendment.
func (ok *Okx) AmendGridAlgoOrder(ctx context.Context, arg GridAlgoOrderAmend) (*GridAlgoOrderIDResponse, error) {
	if arg.AlgoID == "" {
		return nil, errMissingAlgoOrderID
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	type response struct {
		Code string                    `json:"code"`
		Msg  string                    `json:"msg"`
		Data []GridAlgoOrderIDResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, amendGridAlgoOrderEPL, http.MethodPost, gridAmendOrderAlgo, &arg, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg != "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// StopGridAlgoOrder stop a batch of grid algo orders.
func (ok *Okx) StopGridAlgoOrder(ctx context.Context, arg []StopGridAlgoOrderRequest) ([]GridAlgoOrderIDResponse, error) {
	for x := range arg {
		if arg[x].AlgoID == "" {
			return nil, errMissingAlgoOrderID
		}
		if arg[x].InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if !(strings.EqualFold(arg[x].AlgoOrderType, "grid") || strings.EqualFold(arg[x].AlgoOrderType, "contract_grid")) {
			return nil, errMissingAlgoOrderType
		}
		if !(arg[x].StopType == 1 || arg[x].StopType == 2) {
			return nil, errMissingValidStopType
		}
	}
	type response struct {
		Code string                    `json:"code"`
		Msg  string                    `json:"msg"`
		Data []GridAlgoOrderIDResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, stopGridAlgoOrderEPL, http.MethodPost, gridAlgoOrderStop, arg, &resp, true)
}

// GetGridAlgoOrdersList retrives list of pending grid algo orders with the complete data.
func (ok *Okx) GetGridAlgoOrdersList(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	return ok.getGridAlgoOrders(ctx, algoOrderType, algoID,
		instrumentID, instrumentType,
		after, before, limit, gridAlgoOrders)
}

// GetGridAlgoOrderHistory retrives list of grid algo orders with the complete data including the stoped orders.
func (ok *Okx) GetGridAlgoOrderHistory(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	return ok.getGridAlgoOrders(ctx, algoOrderType, algoID,
		instrumentID, instrumentType,
		after, before, limit, gridAlgoOrdersHistory)
}

// getGridAlgoOrderList retrives list of grid algo orders with the complete data.
func (ok *Okx) getGridAlgoOrders(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint, route string) ([]GridAlgoOrderResponse, error) {
	params := url.Values{}
	if !(strings.EqualFold(algoOrderType, "grid") || strings.EqualFold(algoOrderType, "contract_grid")) {
		return nil, errMissingAlgoOrderType
	}
	params.Set("algoOrdType", algoOrderType)
	if algoID != "" {
		params.Set("algoId", algoID)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if strings.EqualFold(instrumentType, OkxInstTypeSpot) || strings.EqualFold(instrumentType, OkxInstTypeMargin) || strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeSwap) {
		params.Set("instType", strings.ToUpper(instrumentType))
	}
	if after != "" {
		params.Set("after", after)
	}
	if before != "" {
		params.Set("before", before)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(route, params)
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []GridAlgoOrderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoOrderHistoryEPL, http.MethodGet, path, nil, &resp, true)
}

// GetGridAlgoOrderDetails retrives grid algo order details
func (ok *Okx) GetGridAlgoOrderDetails(ctx context.Context, algoOrderType, algoID string) (*GridAlgoOrderResponse, error) {
	params := url.Values{}
	if !(strings.EqualFold(algoOrderType, "grid") ||
		strings.EqualFold(algoOrderType, "contract_grid")) {
		return nil, errMissingAlgoOrderType
	}
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	params.Set("algoOrdType", algoOrderType)
	params.Set("algoId", algoID)
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []GridAlgoOrderResponse `json:"data"`
	}
	path := common.EncodeURLValues(gridOrdersAlgoDetails, params)
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoOrderDetailsEPL, http.MethodGet, path, nil, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// GetGridAlgoSubOrders retrives grid algo sub orders
func (ok *Okx) GetGridAlgoSubOrders(ctx context.Context, algoOrderType, algoID, subOrderType, groupID, after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	params := url.Values{}
	if !(strings.EqualFold(algoOrderType, "grid") || strings.EqualFold(algoOrderType, "contract_grid")) {
		return nil, errMissingAlgoOrderType
	}
	params.Set("algoOrdType", algoOrderType)
	if algoID != "" {
		params.Set("algoId", algoID)
	} else {
		return nil, errMissingAlgoOrderID
	}
	if strings.EqualFold(subOrderType, "live") || strings.EqualFold(subOrderType, "filled") {
		params.Set("type", subOrderType)
	} else {
		return nil, errMissingSubOrderType
	}
	if groupID != "" {
		params.Set("groupId", groupID)
	}
	if after != "" {
		params.Set("after", after)
	}
	if before != "" {
		params.Set("before", before)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(gridSuborders, params)
	type response struct {
		Code string                  `json:"code"`
		Msg  string                  `json:"msg"`
		Data []GridAlgoOrderResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoSubOrdersEPL, http.MethodGet, path, nil, &resp, true)
}

// GetGridAlgoOrderPositions retrives grid algo order positions.
func (ok *Okx) GetGridAlgoOrderPositions(ctx context.Context, algoOrderType, algoID string) ([]AlgoOrderPosition, error) {
	params := url.Values{}
	if !(strings.EqualFold(algoOrderType, "grid") || strings.EqualFold(algoOrderType, "contract_grid")) {
		return nil, errMissingAlgoOrderType
	}
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	params.Set("algoOrdType", algoOrderType)
	params.Set("algoId", algoID)
	type response struct {
		Code string              `json:"code"`
		Msg  string              `json:"msg"`
		Data []AlgoOrderPosition `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(gridPositions, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoOrderPositionsEPL, http.MethodGet, path, nil, &resp, true)
}

// SpotGridWithdrawalncome returns the spot grid orders withdrawal profit given an instrument id.
func (ok *Okx) SpotGridWithdrawProfit(ctx context.Context, algoID string) (*AlgoOrderWithdrawaProfit, error) {
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	input := &struct {
		AlgoID string `json:"algoId"`
	}{
		AlgoID: algoID,
	}
	type response struct {
		Code string                     `json:"code"`
		Msg  string                     `json:"msg"`
		Data []AlgoOrderWithdrawaProfit `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, spotGridWithdrawIncomeEPL, http.MethodPost, gridWithdrawalIncome, input, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetTickers retrives the latest price snopshots best bid/ ask price, and tranding volume in the last 34 hours.
func (ok *Okx) GetTickers(ctx context.Context, instType, uly, instID string) ([]TickerResponse, error) {
	params := url.Values{}
	if strings.EqualFold(instType, OkxInstTypeSpot) || strings.EqualFold(instType, OkxInstTypeSwap) || strings.EqualFold(instType, OkxInstTypeFutures) || strings.EqualFold(instType, OkxInstTypeOption) {
		params.Set("instType", instType)
		if (strings.EqualFold(instType, OkxInstTypeSwap) || strings.EqualFold(instType, OkxInstTypeFutures) || strings.EqualFold(instType, OkxInstTypeOption)) && uly != "" {
			params.Set("uly", uly)
		}
	} else if instID != "" {
		params.Set("instId", instID)
	} else {
		return nil, errors.New("missing required variable instType (instrument type) or insId( Instrument ID )")
	}
	path := common.EncodeURLValues(marketTickers, params)
	var response OkxMarketDataResponse
	return response.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTickersEPL, http.MethodGet, path, nil, &response, false)
}

// GetTicker retrieve the latest price snapshot, best bid/ask price, and trading volume in the last 24 hours.
func (ok *Okx) GetTicker(ctx context.Context, instrumentID string) (*TickerResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("missing required variable instType(instruction type) or insId( Instrument ID )")
	}
	path := common.EncodeURLValues(marketTicker, params)
	var response OkxMarketDataResponse
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTickersEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	if len(response.Data) == 0 {
		if response.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(response.Msg)
	}
	return &response.Data[0], nil
}

// GetIndexTickers Retrieves index tickers.
func (ok *Okx) GetIndexTickers(ctx context.Context, quoteCurrency, instId string) ([]IndexTicker, error) {
	response := &struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []IndexTicker `json:"data"`
	}{}
	if instId == "" && quoteCurrency == "" {
		return nil, errors.New("missing required variable! param quoteCcy or instId has to be set")
	}
	params := url.Values{}
	if quoteCurrency != "" {
		params.Set("quoteCcy", quoteCurrency)
	} else if instId != "" {
		params.Set("instId", instId)
	}
	path := common.EncodeURLValues(indexTickers, params)
	return response.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getIndexTickersEPL, http.MethodGet, path, nil, response, false)
}

// GetInstrumentIDFromPair returns the instrument ID for the corresponding asset pairs and asset type( Instrument Type )
func (ok *Okx) GetInstrumentIDFromPair(pair currency.Pair, a asset.Item) (string, error) {
	format, err := ok.GetPairFormat(a, false)
	if err != nil {
		return "", err
	}
	if pair.IsEmpty() || pair.Base.String() == "" || pair.Quote.String() == "" {
		return "", errors.New("incomplete currency pair")
	}
	if a == asset.PerpetualSwap {
		return pair.Base.String() + format.Delimiter + pair.Quote.String() + format.Delimiter + OkxInstTypeSwap, nil
	} else if a == asset.Option {
		instruments, er := ok.GetInstruments(context.Background(), &InstrumentsFetchParams{
			InstrumentType: OkxInstTypeOption,
			Underlying:     pair.Base.String() + format.Delimiter + pair.Quote.String(),
		})
		if er != nil {
			return "", er
		}
		for x := range instruments {
			p, er := currency.NewPairFromString(instruments[x].Underlying)
			if er != nil {
				continue
			}
			if p.Equal(pair) {
				return instruments[x].InstrumentID, nil
			}
		}
	} else if a == asset.Futures {
		instruments, er := ok.GetInstruments(context.Background(), &InstrumentsFetchParams{
			InstrumentType: OkxInstTypeFutures,
		})
		if er != nil {
			return "", er
		}
		for x := range instruments {
			p, er := currency.NewPairFromString(instruments[x].Underlying)
			if er != nil {
				continue
			}
			if p.Equal(pair) {
				return instruments[x].InstrumentID, nil
			}
		}
		return "", errors.New("instrument id with this asset pairs not found")
	}
	return pair.Base.String() + format.Delimiter + pair.Quote.String(), nil
}

// GetInstrumentTypeFromAssetItem returns a string representation of asset.Item; which is an equivalent term for InstrumentType in Okx exchange.
func (ok *Okx) GetInstrumentTypeFromAssetItem(assetType asset.Item) string {
	if assetType == asset.PerpetualSwap {
		return OkxInstTypeSwap
	}
	return assetType.String()
}

// GetUnderlying returns the instrument ID for the corresponding asset pairs and asset type( Instrument Type )
func (ok *Okx) GetUnderlying(pair currency.Pair, a asset.Item) (string, error) {
	format, err := ok.GetPairFormat(a, false)
	if err != nil {
		return "", err
	}
	if pair.IsEmpty() || pair.Base.String() == "" || pair.Quote.String() == "" {
		return "", errors.New("incomplete currency pair")
	}
	return pair.Base.String() + format.Delimiter + pair.Quote.String(), nil
}

// GetPairFromInstrumentID returns a currency pair give an instrument ID and asset Item, which represents the instrument type.
func (ok *Okx) GetPairFromInstrumentID(instrumentID string) (currency.Pair, error) {
	codes := strings.Split(instrumentID, "-")
	if len(codes) >= 2 {
		instrumentID = codes[0] + "-" + codes[1]
	}
	pair, er := currency.NewPairFromString(instrumentID)
	return pair, er
}

// GetOrderBook returns the recent order asks and bids before specified timestamp.
func (ok *Okx) GetOrderBookDepth(ctx context.Context, instrumentID string, depth uint) (*OrderBookResponse, error) {
	params := url.Values{}
	params.Set("instId", instrumentID)
	if depth > 0 {
		params.Set("sz", strconv.Itoa(int(depth)))
	}
	type response struct {
		Code int                 `json:"code,string"`
		Msg  string              `json:"msg"`
		Data []OrderBookResponse `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(marketBooks, params)
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderBookEPL, http.MethodGet, path, nil, &resp, false)
	if err != nil {
		return nil, err
	} else if len(resp.Data) == 0 {
		if resp.Msg != "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// GetIntervalEnum allowed interval params by Okx Exchange
func (ok *Okx) GetIntervalEnum(interval kline.Interval) string {
	switch interval {
	case kline.OneMin:
		return "1m"
	case kline.ThreeMin:
		return "3m"
	case kline.FiveMin:
		return "5m"
	case kline.FifteenMin:
		return "15m"
	case kline.ThirtyMin:
		return "30m"
	case kline.OneHour:
		return "1H"
	case kline.TwoHour:
		return "2H"
	case kline.FourHour:
		return "4H"
	case kline.SixHour:
		return "6H"
	case kline.EightHour:
		return "8H"
	case kline.TwelveHour:
		return "12H"
	case kline.OneDay:
		return "1D"
	case kline.TwoDay:
		return "2D"
	case kline.ThreeDay:
		return "3D"
	case kline.OneWeek:
		return "1W"
	case kline.OneMonth:
		return "1M"
	case kline.ThreeMonth:
		return "3M"
	case kline.SixMonth:
		return "6M"
	case kline.OneYear:
		return "1Y"
	default:
		return ""
	}
}

// GetCandlesticks Retrieve the candlestick charts. This endpoint can retrieve the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
func (ok *Okx) GetCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before time.Time, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandles)
}

// GetCandlesticksHistory Retrieve history candlestick charts from recent years.
func (ok *Okx) GetCandlesticksHistory(ctx context.Context, instrumentID string, interval kline.Interval, before time.Time, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandlesHistory)
}

// GetIndexCandlesticks Retrieve the candlestick charts of the index. This endpoint can retrieve the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
// the respos is a lis of Candlestick data.
func (ok *Okx) GetIndexCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before time.Time, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandlesIndex)
}

// GetMarkPriceCandlesticks Retrieve the candlestick charts of mark price. This endpoint can retrieve the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
func (ok *Okx) GetMarkPriceCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before time.Time, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketPriceCandles)
}

// GetCandlestickData handles fetching the data for both the default GetCandlesticks, GetCandlesticksHistory, and GetIndexCandlesticks() methods.
func (ok *Okx) GetCandlestickData(ctx context.Context, instrumentID string, interval kline.Interval, before time.Time, after time.Time, limit uint64, route string) ([]CandleStick, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	} else {
		params.Set("instId", instrumentID)
	}
	type response struct {
		Msg  string      `json:"msg"`
		Code int         `json:"code,string"`
		Data interface{} `json:"data"`
	}
	var resp response
	if limit > 0 && limit <= 100 {
		params.Set("limit", strconv.Itoa(int(limit)))
	} else if limit > 100 {
		return nil, errLimitExceedsMaximumResultPerRequest
	}
	if !before.IsZero() {
		params.Set("before", strconv.Itoa(int(before.UnixMilli())))
	}
	if !after.IsZero() {
		params.Set("after", strconv.Itoa(int(after.UnixMilli())))
	}
	bar := ok.GetIntervalEnum(interval)
	if bar != "" {
		params.Set("bar", bar)
	}
	path := common.EncodeURLValues(marketCandles, params)
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getCandlesticksEPL, http.MethodGet, path, nil, &resp, false)
	if err != nil {
		return nil, err
	}
	responseData, okk := (resp.Data).([]interface{})
	if !okk {
		return nil, errUnableToTypeAssertResponseData
	}
	klineData := make([]CandleStick, len(responseData))
	for x := range responseData {
		individualData, ok := responseData[x].([]interface{})
		if !ok {
			return nil, errUnableToTypeAssertKlineData
		}
		if len(individualData) != 7 {
			return nil, errUnexpectedKlineDataLength
		}
		var candle CandleStick
		var er error
		timestamp, er := strconv.Atoi(individualData[0].(string))
		if er != nil {
			return nil, er
		}
		candle.OpenTime = time.UnixMilli(int64(timestamp))
		if candle.OpenPrice, er = convert.FloatFromString(individualData[1]); er != nil {
			return nil, er
		}
		if candle.HighestPrice, er = convert.FloatFromString(individualData[2]); er != nil {
			return nil, er
		}
		if candle.LowestPrice, er = convert.FloatFromString(individualData[3]); er != nil {
			return nil, er
		}
		if candle.ClosePrice, er = convert.FloatFromString(individualData[4]); er != nil {
			return nil, er
		}
		if candle.Volume, er = convert.FloatFromString(individualData[5]); er != nil {
			return nil, er
		}
		if candle.QuoteAssetVolume, er = convert.FloatFromString(individualData[6]); er != nil {
			return nil, er
		}
		klineData[x] = candle
	}
	return klineData, nil
}

// GetTrades Retrieve the recent transactions of an instrument.
func (ok *Okx) GetTrades(ctx context.Context, instrumentId string, limit uint) ([]TradeResponse, error) {
	type response struct {
		Msg  string          `json:"msg"`
		Code int             `json:"code,string"`
		Data []TradeResponse `json:"data"`
	}
	var resp response
	params := url.Values{}
	if instrumentId == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentId)
	if limit > 0 && limit <= 500 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(marketTrades, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesRequestEPL, http.MethodGet, path, nil, &resp, false)
}

// GetTradesHistory retrieve the recent transactions of an instrument from the last 3 months with pagination.
func (ok *Okx) GetTradesHistory(ctx context.Context, instrumentID, before, after string, limit uint) ([]TradeResponse, error) {
	type response struct {
		Msg  string          `json:"msg"`
		Code int             `json:"code,string"`
		Data []TradeResponse `json:"data"`
	}
	var resp response
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	if before != "" {
		params.Set("before", before)
	}
	if after != "" {
		params.Set("after", after)
	}
	if limit > 0 && limit <= 100 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	path := common.EncodeURLValues(marketTradesHistory, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesRequestEPL, http.MethodGet, path, nil, &resp, false)
}

// Get24HTotalVolume The 24-hour trading volume is calculated on a rolling basis, using USD as the pricing unit.
func (ok *Okx) Get24HTotalVolume(ctx context.Context) (*TradingVolumdIn24HR, error) {
	type response struct {
		Msg  string                `json:"msg"`
		Code int                   `json:"code,string"`
		Data []TradingVolumdIn24HR `json:"data"`
	}
	var resp response
	era := ok.SendHTTPRequest(ctx, exchange.RestSpot, get24HTotalVolumeEPL, http.MethodGet, marketPlatformVolumeIn24Hour, nil, &resp, false)
	if era != nil {
		return nil, era
	}
	if len(resp.Data) == 0 {
		return nil, errNo24HrTradeVolumeFound
	}
	return &resp.Data[0], nil
}

// GetOracle Get the crypto price of signing using Open Oracle smart contract.
func (ok *Okx) GetOracle(ctx context.Context) (*OracleSmartContractResponse, error) {
	type response struct {
		Msg  string                        `json:"msg"`
		Code int                           `json:"code,string"`
		Data []OracleSmartContractResponse `json:"data"`
	}
	var resp response
	era := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOracleEPL, http.MethodGet, marketOpenOracles, nil, &resp, false)
	if era != nil {
		return nil, era
	}
	if len(resp.Data) == 0 {
		return nil, errOracleInformationNotFound
	}
	return &resp.Data[0], nil
}

// GetExchangeRate this interface provides the average exchange rate data for 2 weeks
// from USD to CNY
func (ok *Okx) GetExchangeRate(ctx context.Context) (*UsdCnyExchangeRate, error) {
	type response struct {
		Msg  string               `json:"msg"`
		Code int                  `json:"code,string"`
		Data []UsdCnyExchangeRate `json:"data"`
	}
	var resp response
	era := ok.SendHTTPRequest(ctx, exchange.RestSpot, getExchangeRateRequestEPL, http.MethodGet, marketExchangeRate, nil, &resp, false)
	if era != nil {
		return nil, era
	}
	if len(resp.Data) == 0 {
		return nil, errExchangeInfoNotFound
	}
	return &resp.Data[0], nil
}

// GetIndexComponents returns the index component information data on the market
func (ok *Okx) GetIndexComponents(ctx context.Context, index string) (*IndexComponent, error) {
	type response struct {
		Msg  string          `json:"msg"`
		Code int             `json:"code,string"`
		Data *IndexComponent `json:"data"`
	}
	params := url.Values{}
	params.Set("index", index)
	var resp response
	path := common.EncodeURLValues(marketIndexComponents, params)
	era := ok.SendHTTPRequest(ctx, exchange.RestSpot, getINdexComponentsEPL, http.MethodGet, path, nil, &resp, false)
	if era != nil {
		return nil, era
	}
	if resp.Data == nil {
		return nil, errIndexComponentNotFound
	}
	return resp.Data, nil
}

// GetblockTickers retrieve the latest block trading volume in the last 24 hours.
// Instrument Type Is Mendatory, and Underlying is Optional.
func (ok *Okx) GetBlockTickers(ctx context.Context, instrumentType, underlying string) ([]BlockTicker, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, OkxInstTypeSpot) || strings.EqualFold(instrumentType, OkxInstTypeSwap) || strings.EqualFold(instrumentType, OkxInstTypeFutures) || strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instrumentType)
	if underlying != "" {
		params.Set("uly", underlying)
	}
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []BlockTicker `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(marketBlockTickers, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTickersEPL, http.MethodGet, path, nil, &resp, true)
}

// GetBlockTicker retrieve the latest block trading volume in the last 24 hours.
func (ok *Okx) GetBlockTicker(ctx context.Context, instrumentID string) (*BlockTicker, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	path := common.EncodeURLValues(marketBlockTicker, params)
	type response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []BlockTicker `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTickersEPL, http.MethodGet, path, nil, &resp, true)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errNoValidResponseFromServer
		}
		return nil, errors.New(resp.Msg)
	}
	return &resp.Data[0], nil
}

// GetBlockTrades retrieve the recent block trading transactions of an instrument. Descending order by tradeId.
func (ok *Okx) GetBlockTrades(ctx context.Context, instrumentID string) ([]BlockTrade, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	path := common.EncodeURLValues(marketBlockTrades, params)
	type response struct {
		Code string       `json:"code"`
		Msg  string       `json:"msg"`
		Data []BlockTrade `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTradesEPL, http.MethodGet, path, nil, &resp, true)
}

/************************************ Public Data Endpoinst *************************************************/

// GetInstruments Retrieve a list of instruments with open contracts.
func (ok *Okx) GetInstruments(ctx context.Context, arg *InstrumentsFetchParams) ([]Instrument, error) {
	params := url.Values{}
	if !(strings.EqualFold(arg.InstrumentType, OkxInstTypeSpot) || strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) || strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) || strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) || strings.EqualFold(arg.InstrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", arg.InstrumentType)
	}
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	type response struct {
		Code string       `json:"code"`
		Msg  string       `json:"msg"`
		Data []Instrument `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(publicInstruments, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInstrumentsEPL, http.MethodGet, path, nil, &resp, false)
}

// GetDeliveryHistory retrieve the estimated delivery price of the last 3 months, which will only have a return value one hour before the delivery/exercise.
func (ok *Okx) GetDeliveryHistory(ctx context.Context, instrumentType, underlying string, after, before time.Time, limit int) ([]DeliveryHistory, error) {
	params := url.Values{}
	if instrumentType != "" && !(strings.EqualFold(instrumentType, OkxInstTypeFutures) || strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, fmt.Errorf("unacceptable instrument Type! Only %s and %s are allowed", "FUTURE", OkxInstTypeOption)
	} else if instrumentType == "" {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", instrumentType)
	}
	if underlying != "" {
		params.Set("Underlying", underlying)
		params.Set("uly", underlying)
	} else {
		return nil, errParameterUnderlyingCanNotBeEmpty
	}
	if !(after.IsZero()) {
		params.Set("after", strconv.Itoa(int(after.UnixMilli())))
	}
	if !(before.IsZero()) {
		params.Set("before", strconv.Itoa(int(before.UnixMilli())))
	}
	if limit > 0 && limit <= 100 {
		params.Set("limit", strconv.Itoa(limit))
	} else {
		return nil, errLimitValueExceedsMaxof100
	}
	var resp DeliveryHistoryResponse
	path := common.EncodeURLValues(publicDeliveryExerciseHistory, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDeliveryExerciseHistoryEPL, http.MethodGet, path, nil, &resp, false)
}

// GetOpenInterest retrieve the total open interest for contracts on OKX
func (ok *Okx) GetOpenInterest(ctx context.Context, instType, uly, instId string) ([]OpenInterest, error) {
	params := url.Values{}
	if !(strings.EqualFold(instType, OkxInstTypeSpot) || strings.EqualFold(instType, OkxInstTypeFutures) || strings.EqualFold(instType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", instType)
	}
	if uly != "" {
		params.Set("uly", uly)
	}
	if instId != "" {
		params.Set("instId", instId)
	}
	type response struct {
		Code int            `json:"code,string"`
		Msg  string         `json:"msg"`
		Data []OpenInterest `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(publicOpenInterestValues, params)
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOpenInterestEPL, http.MethodGet, path, nil, &resp, false)
}

// GetFundingRate  Retrieve funding rate.
func (ok *Okx) GetFundingRate(ctx context.Context, instrumentID string) (*FundingRateResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errMissingInstrumentID
	}
	path := common.EncodeURLValues(publicFundingRate, params)
	type response struct {
		Code string                `json:"code"`
		Data []FundingRateResponse `json:"data"`
		Msg  string                `json:"msg"`
	}
	var resp response
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundingEPL, http.MethodGet, path, nil, &resp, false)
	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}
	return nil, err
}

// GetFundingRateHistory retrieve funding rate history. This endpoint can retrieve data from the last 3 months.
func (ok *Okx) GetFundingRateHistory(ctx context.Context, instrumentID string, before time.Time, after time.Time, limit uint) ([]FundingRateResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errMissingInstrumentID
	}
	if !before.IsZero() {
		params.Set("before", strconv.Itoa(int(before.UnixMilli())))
	}
	if !after.IsZero() {
		params.Set("after", strconv.Itoa(int(after.UnixMilli())))
	}
	if limit > 0 && limit < 100 {
		params.Set("limit", strconv.Itoa(int(limit)))
	} else if limit > 0 {
		return nil, errLimitValueExceedsMaxof100
	}
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []FundingRateResponse `json:"data"`
	}
	path := common.EncodeURLValues(publicFundingRateHistory, params)
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundingRateHistoryEPL, http.MethodGet, path, nil, &resp, false)
}

// GetLimitPrice retrieve the highest buy limit and lowest sell limit of the instrument.
func (ok *Okx) GetLimitPrice(ctx context.Context, instrumentID string) (*LimitPriceResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errMissingInstrumentID
	}
	path := common.EncodeURLValues(publicLimitPath, params)
	type response struct {
		Code string               `json:"code"`
		Msg  string               `json:"msg"`
		Data []LimitPriceResponse `json:"data"`
	}
	var resp response
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLimitPriceEPL, http.MethodGet, path, nil, &resp, false); er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}
	return nil, errFundingRateHistoryNotFound
}

// GetOptionMarketData retrieve option market data.
func (ok *Okx) GetOptionMarketData(ctx context.Context, underlying string, expTime time.Time) ([]OptionMarketDataResponse, error) {
	params := url.Values{}
	if underlying != "" {
		params.Set("uly", underlying)
	} else {
		return nil, errMissingRequiredUnderlying
	}
	if !expTime.IsZero() {
		params.Set("expTime", fmt.Sprintf("%d%d%d", expTime.Year(), expTime.Month(), expTime.Day()))
	}
	path := common.EncodeURLValues(publicOptionalData, params)
	type response struct {
		Code string                     `json:"code"`
		Msg  string                     `json:"msg"`
		Data []OptionMarketDataResponse `json:"data"`
	}
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOptionMarketDateEPL, http.MethodGet, path, nil, &resp, false)
}

// GetEstimatedDeliveryPrice retrieve the estimated delivery price which will only have a return value one hour before the delivery/exercise.
func (ok *Okx) GetEstimatedDeliveryPrice(ctx context.Context, instrumentID string) (*DeliveryEstimatedPrice, error) {
	var resp DeliveryEstimatedPriceResponse
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errMissingRequiredParamInstID
	}
	path := common.EncodeURLValues(publicEstimatedPrice, params)
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getEstimatedDeliveryPriceEPL, http.MethodGet, path, nil, &resp, false); er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}
	return nil, errors.New(resp.Msg)
}

// GetDiscountRateAndInterestFreeQuota retrieve discount rate level and interest-free quota.
func (ok *Okx) GetDiscountRateAndInterestFreeQuota(ctx context.Context, currency string, discountLevel int8) (*DiscountRate, error) {
	var response DiscountRateResponse
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if discountLevel > 0 && discountLevel < 5 {
		params.Set("discountLv", strconv.Itoa(int(discountLevel)))
	}
	path := common.EncodeURLValues(publicDiscountRate, params)
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getDiscountRateAndInterestFreeQuotaEPL, http.MethodGet, path, nil, &response, false); er != nil {
		return nil, er
	}
	if len(response.Data) > 0 {
		return &response.Data[0], nil
	}
	return nil, errors.New(response.Msg)
}

// GetSystemTime Retrieve API server time.
func (ok *Okx) GetSystemTime(ctx context.Context) (*time.Time, error) {
	type response struct {
		Code string       `json:"code"`
		Msg  string       `json:"msg"`
		Data []ServerTime `json:"data"`
	}
	var resp response
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getSystemTimeEPL, http.MethodGet, publicTime, nil, &resp, false); er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return &(resp.Data[0].Timestamp), nil
	}
	return nil, errors.New(resp.Msg)
}

// GetLiquidationOrders retrieve information on liquidation orders in the last day.
func (ok *Okx) GetLiquidationOrders(ctx context.Context, arg *LiquidationOrderRequestParams) (*LiquidationOrder, error) {
	params := url.Values{}
	if !(strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) || strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) || strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) || strings.EqualFold(arg.InstrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", arg.InstrumentType)
	}
	if strings.EqualFold(arg.MarginMode, "isolated") || strings.EqualFold(arg.MarginMode, "cross") {
		params.Set("mgnMode", arg.MarginMode)
	}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) && arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	} else if strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) && arg.Currency.String() != "" {
		params.Set("ccy", arg.Currency.String())
	} else {
		return nil, errEitherInstIDOrCcyIsRequired
	}
	if (strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) || strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) || strings.EqualFold(arg.InstrumentType, OkxInstTypeOption)) && arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) && (strings.EqualFold(arg.Alias, "this_week") || strings.EqualFold(arg.Alias, "next_week ") || strings.EqualFold(arg.Alias, "quarter") || strings.EqualFold(arg.Alias, "next_quarter")) {
		params.Set("alias", arg.Alias)
	}
	if ((strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) || strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap)) &&
		strings.EqualFold(arg.Alias, "unfilled")) || strings.EqualFold(arg.Alias, "filled ") {
		params.Set("alias", arg.Underlying)
	}
	if !arg.Before.IsZero() {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !arg.After.IsZero() {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 && arg.Limit < 100 {
		params.Set("limit", strconv.FormatInt(arg.Limit, 10))
	}
	path := common.EncodeURLValues(publicLiquidationOrders, params)
	var response LiquidationOrderResponse
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLiquidationOrdersEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	if len(response.Data) > 0 {
		return &response.Data[0], nil
	}
	return nil, errLiquidationOrderResponseNotFound
}

// GetMarkPrice  Retrieve mark price.
func (ok *Okx) GetMarkPrice(ctx context.Context, instrumentType, underlying, instrumentID string) ([]MarkPrice, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", instrumentType)
	}
	if underlying != "" {
		params.Set("uly", underlying)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	var response MarkPriceResponse
	path := common.EncodeURLValues(publicMarkPrice, params)
	return response.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMarkPriceEPL, http.MethodGet, path, nil, &response, false)
}

// GetPositionTiers retrieve position tiers information，maximum leverage depends on your borrowings and margin ratio.
func (ok *Okx) GetPositionTiers(ctx context.Context, instrumentType, tradeMode, underlying, instrumentID, tiers string) ([]PositionTiers, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", instrumentType)
	}
	if !(strings.EqualFold(tradeMode, "cross") || strings.EqualFold(tradeMode, "isolated")) {
		return nil, errIncorrectRequiredParameterTradeMode
	} else {
		params.Set("tdMode", tradeMode)
	}
	if (!strings.EqualFold(instrumentType, OkxInstTypeMargin)) && underlying != "" {
		params.Set("uly", underlying)
	}
	if strings.EqualFold(instrumentType, OkxInstTypeMargin) && instrumentID != "" {
		params.Set("instId", instrumentID)
	} else if strings.EqualFold(instrumentType, OkxInstTypeMargin) {
		return nil, errMissingInstrumentID
	}
	if tiers != "" {
		params.Set("tiers", tiers)
	}
	var response PositionTiersResponse
	path := common.EncodeURLValues(publicPositionTiers, params)
	return response.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionTiersEPL, http.MethodGet, path, nil, &response, false)
}

// GetInterestRate retrives an interest rate and loan quota information for various currencies.
func (ok *Okx) GetInterestRateAndLoanQuota(ctx context.Context) (map[string][]InterestRateLoanQuotaItem, error) {
	var response InterestRateLoanQuotaResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateAndLoanQuotaEPL, http.MethodGet, publicInterestRateAndLoanQuota, nil, &response, false)
	if err != nil {
		return nil, err
	} else if len(response.Data) == 0 {
		return nil, errInterestRateAndLoanQuotaNotFound
	}
	return response.Data[0], nil
}

// GetInterestRate retrives an interest rate and loan quota information for VIP users of various currencies.
func (ok *Okx) GetInterestRateAndLoanQuotaForVIPLoans(ctx context.Context) (*VIPInterestRateAndLoanQuotaInformation, error) {
	var response VIPInterestRateAndLoanQuotaInformationResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateAndLoanQuoteForVIPLoansEPL, http.MethodGet, publicVIPInterestRateAndLoanQuota, nil, &response, false)
	if err != nil {
		return nil, err
	} else if len(response.Data) == 0 {
		return nil, errInterestRateAndLoanQuotaNotFound
	}
	return &response.Data[0], nil
}

// GetPublicUnderlyings returns list of underlyings for various instrument types.
func (ok *Okx) GetPublicUnderlyings(ctx context.Context, instrumentType string) ([]string, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", strings.ToUpper(instrumentType))
	}
	path := common.EncodeURLValues(publicUnderlyings, params)
	type response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	var resp response
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getUnderlyingEPL, http.MethodGet, path, nil, &resp, false); er != nil {
		return nil, er
	}
	if len(resp.Data) > 0 {
		return resp.Data[0], nil
	}
	return nil, errUnderlyingsForSpecifiedInstTypeNofFound
}

// GetInsuranceFund returns insurance fund balance informations.
func (ok *Okx) GetInsuranceFundInformations(ctx context.Context, arg InsuranceFundInformationRequestParams) (*InsuranceFundInformation, error) {
	params := url.Values{}
	if !(strings.EqualFold(arg.InstrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(arg.InstrumentType, OkxInstTypeOption)) {
		return nil, errMissingRequiredArgInstType
	} else {
		params.Set("instType", strings.ToUpper(arg.InstrumentType))
	}
	if strings.EqualFold(arg.Type, "liquidation_balance_deposit") ||
		strings.EqualFold(arg.Type, "bankruptcy_loss") ||
		strings.EqualFold(arg.Type, "platform_revenue ") {
		params.Set("type", arg.Type)
	}
	if (!strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin)) && arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	} else if !strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin) {
		return nil, errParameterUnderlyingCanNotBeEmpty
	}
	if (strings.EqualFold(arg.InstrumentType, OkxInstTypeMargin)) && arg.Currency != "" {
		params.Set("ccy", arg.Currency)
	}
	if !arg.Before.IsZero() {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !arg.After.IsZero() {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 && arg.Limit < 100 {
		params.Set("limit", strconv.Itoa(int(arg.Limit)))
	}
	var response InsuranceFundInformationResponse
	path := common.EncodeURLValues(publicInsuranceFunds, params)
	if er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getInsuranceFundEPL, http.MethodGet, path, nil, &response, false); er != nil {
		return nil, er
	}
	if len(response.Data) > 0 {
		return &response.Data[0], nil
	}
	return nil, errInsuranceFundInformationNotFound
}

// CurrencyUnitConvert convert currency to contract, or contract to currency.
func (ok *Okx) CurrencyUnitConvert(ctx context.Context, instrumentID string, quantity, orderPrice float64, convertType CurrencyConvertType, unitOfCurrency string) (*UnitConvertResponse, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if quantity <= 0 {
		return nil, errMissingQuantity
	}
	params.Set("instId", instrumentID)
	params.Set("sz", strconv.FormatFloat(quantity, 'f', 0, 64))
	if orderPrice > 0 {
		params.Set("px", strconv.FormatFloat(orderPrice, 'f', 0, 64))
	}
	if convertType == CurrencyToContract || convertType == ContractToCurrency {
		params.Set("type", strconv.Itoa(int(convertType)))
	}
	if unitOfCurrency != "" {
		params.Set("unit", unitOfCurrency)
	}
	path := common.EncodeURLValues(publicCurrencyConvertContract, params)
	type response struct {
		Code string                `json:"code"`
		Msg  string                `json:"msg"`
		Data []UnitConvertResponse `json:"data"`
	}
	var resp response
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, unitConvertEPL, http.MethodGet, path, nil, &resp, false)
	if er != nil {
		return nil, er
	}
	if len(resp.Data) == 0 {
		if resp.Msg == "" {
			return nil, errors.New(resp.Msg)
		}
		return nil, errNoValidResponseFromServer
	}
	return &resp.Data[0], nil
}

// Trading Data Endpoints

// GetSupportCoins retrieve the currencies supported by the trading data endpoints
func (ok *Okx) GetSupportCoins(ctx context.Context) (*SupportedCoinsData, error) {
	var response SupportedCoinsResponse
	return response.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSupportCoinEPL, http.MethodGet, tradingDataSupportedCoins, nil, &response, false)
}

// GetTakerVolume retrieve the taker volume for both buyers and sellers.
func (ok *Okx) GetTakerVolume(ctx context.Context, currency, instrumentType string, begin, end time.Time, period kline.Interval) ([]TakerVolume, error) {
	params := url.Values{}
	if !(strings.EqualFold(instrumentType, "CONTRACTS") ||
		strings.EqualFold(instrumentType, OkxInstTypeSpot)) {
		return nil, errMissingRequiredArgInstType
	} else if strings.EqualFold(instrumentType, OkxInstTypeFutures) ||
		strings.EqualFold(instrumentType, OkxInstTypeMargin) ||
		strings.EqualFold(instrumentType, OkxInstTypeSwap) ||
		strings.EqualFold(instrumentType, OkxInstTypeOption) {
		return nil, fmt.Errorf("instrument type %s is not allowed for this query", instrumentType)
	} else {
		params.Set("instType", strings.ToUpper(instrumentType))
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !begin.IsZero() {
		params.Set("begin", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	if !end.IsZero() {
		params.Set("end", strconv.FormatInt(end.UnixMilli(), 10))
	}
	path := common.EncodeURLValues(tradingTakerVolume, params)
	var response TakerVolumeResponse
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTakerVolumeEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	takerVolumes := []TakerVolume{}
	for x := range response.Data {
		if len(response.Data[x]) != 3 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil {
			continue
		}
		sellVolume, er := strconv.ParseFloat(response.Data[x][1], 64)
		if er != nil {
			continue
		}
		buyVolume, er := strconv.ParseFloat(response.Data[x][2], 64)
		if er != nil {
			continue
		}
		takerVolume := TakerVolume{
			Timestamp:  time.UnixMilli(int64(timestamp)),
			SellVolume: sellVolume,
			BuyVolume:  buyVolume,
		}
		takerVolumes = append(takerVolumes, takerVolume)
	}
	return takerVolumes, nil
}

// GetMarginLendingRatio retrieve the ratio of cumulative amount between currency margin quote currency and base currency.
func (ok *Okx) GetMarginLendingRatio(ctx context.Context, currency string, begin, end time.Time, period kline.Interval) ([]MarginLendRatioItem, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !begin.IsZero() {
		params.Set("begin", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	if !end.IsZero() {
		params.Set("end", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	var response MarginLendRatioResponse
	path := common.EncodeURLValues(tradingMarginLoanRatio, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getMarginLendingRatioEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	lendingRatios := []MarginLendRatioItem{}
	for x := range response.Data {
		if len(response.Data[x]) != 2 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil || timestamp <= 0 {
			continue
		}
		ratio, er := strconv.ParseFloat(response.Data[x][0], 64)
		if er != nil || ratio <= 0 {
			continue
		}
		lendRatio := MarginLendRatioItem{
			Timestamp:       time.UnixMilli(int64(timestamp)),
			MarginLendRatio: ratio,
		}
		lendingRatios = append(lendingRatios, lendRatio)
	}
	return lendingRatios, nil
}

// GetLongShortRatio retrieve the ratio of users with net long vs net short positions for futures and perpetual swaps.
func (ok *Okx) GetLongShortRatio(ctx context.Context, currency string, begin, end time.Time, period kline.Interval) ([]LongShortRatio, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !begin.IsZero() {
		params.Set("begin", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	if !end.IsZero() {
		params.Set("end", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	var response LongShortRatioResponse
	path := common.EncodeURLValues(longShortAccountRatio, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLongShortRatioEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	ratios := []LongShortRatio{}
	for x := range response.Data {
		if len(response.Data[x]) != 2 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil || timestamp <= 0 {
			continue
		}
		ratio, er := strconv.ParseFloat(response.Data[x][0], 64)
		if er != nil || ratio <= 0 {
			continue
		}
		dratio := LongShortRatio{
			Timestamp:       time.UnixMilli(int64(timestamp)),
			MarginLendRatio: ratio,
		}
		ratios = append(ratios, dratio)
	}
	return ratios, nil
}

// GetContractsOpenInterestAndVolume retrieve the open interest and trading volume for futures and perpetual swaps.
func (ok *Okx) GetContractsOpenInterestAndVolume(
	ctx context.Context, currency string,
	begin, end time.Time, period kline.Interval) ([]OpenInterestVolume, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !begin.IsZero() {
		params.Set("begin", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	if !end.IsZero() {
		params.Set("end", strconv.FormatInt(begin.UnixMilli(), 10))
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	openInterestVolumes := []OpenInterestVolume{}
	var response OpenInterestVolumeResponse
	path := common.EncodeURLValues(contractOpenInterestVolume, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getContractsOpeninterestAndVolumeEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	for x := range response.Data {
		if len(response.Data[x]) != 3 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil || timestamp <= 0 {
			continue
		}
		openInterest, er := strconv.Atoi(response.Data[x][1])
		if er != nil || openInterest <= 0 {
			continue
		}
		volumen, err := strconv.Atoi(response.Data[x][2])
		if err != nil {
			continue
		}
		openInterestVolume := OpenInterestVolume{
			Timestamp:    time.UnixMilli(int64(timestamp)),
			Volume:       float64(volumen),
			OpenInterest: float64(openInterest),
		}
		openInterestVolumes = append(openInterestVolumes, openInterestVolume)
	}
	return openInterestVolumes, nil
}

// GetOptionsOpenInterestAndVolume retrieve the open interest and trading volume for options.
func (ok *Okx) GetOptionsOpenInterestAndVolume(ctx context.Context, currency string,
	period kline.Interval) ([]OpenInterestVolume, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	openInterestVolumes := []OpenInterestVolume{}
	var response OpenInterestVolumeResponse
	path := common.EncodeURLValues(optionOpenInterestVolume, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOptionsOpenInterestAndVolumeEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	for x := range response.Data {
		if len(response.Data[x]) != 3 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil || timestamp <= 0 {
			continue
		}
		openInterest, er := strconv.Atoi(response.Data[x][1])
		if er != nil || openInterest <= 0 {
			continue
		}
		volumen, err := strconv.Atoi(response.Data[x][2])
		if err != nil {
			continue
		}
		openInterestVolume := OpenInterestVolume{
			Timestamp:    time.UnixMilli(int64(timestamp)),
			Volume:       float64(volumen),
			OpenInterest: float64(openInterest),
		}
		openInterestVolumes = append(openInterestVolumes, openInterestVolume)
	}
	return openInterestVolumes, nil
}

// GetPutCallRatio retrieve the open interest ration and trading volume ratio of calls vs puts.
func (ok *Okx) GetPutCallRatio(ctx context.Context, currency string,
	period kline.Interval) ([]OpenInterestVolumeRatio, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	openInterestVolumeRatios := []OpenInterestVolumeRatio{}
	var response OpenInterestVolumeResponse
	path := common.EncodeURLValues(optionOpenInterestVolumeRatio, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getPutCallRatioEPL, http.MethodGet, path, nil, &response, false)
	if er != nil {
		return nil, er
	}
	for x := range response.Data {
		if len(response.Data[x]) != 3 {
			continue
		}
		timestamp, er := strconv.Atoi(response.Data[x][0])
		if er != nil || timestamp <= 0 {
			continue
		}
		openInterest, er := strconv.Atoi(response.Data[x][1])
		if er != nil || openInterest <= 0 {
			continue
		}
		volumen, err := strconv.Atoi(response.Data[x][2])
		if err != nil {
			continue
		}
		openInterestVolume := OpenInterestVolumeRatio{
			Timestamp:         time.UnixMilli(int64(timestamp)),
			VolumeRatio:       float64(volumen),
			OpenInterestRatio: float64(openInterest),
		}
		openInterestVolumeRatios = append(openInterestVolumeRatios, openInterestVolume)
	}
	return openInterestVolumeRatios, nil
}

// GetOpenInterestAndVolume retrieve the open interest and trading volume of calls and puts for each upcoming expiration.
func (ok *Okx) GetOpenInterestAndVolumeExpiry(ctx context.Context, currency string, period kline.Interval) ([]ExpiryOpenInterestAndVolume, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data [][6]string `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(optionOpenInterestVolumeExpiry, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOpenInterestAndVolumeEPL, http.MethodGet, path, nil, &resp, false)
	if er != nil {
		return nil, er
	}
	var volumes []ExpiryOpenInterestAndVolume
	for x := range resp.Data {
		if len(resp.Data[x]) != 6 {
			continue
		}
		timestamp, er := strconv.Atoi(resp.Data[x][0])
		if er != nil {
			continue
		}
		var expiryTime time.Time
		expTime := resp.Data[x][1]
		if expTime != "" && len(expTime) == 8 {
			year, er := strconv.Atoi(expTime[0:4])
			if er != nil {
				continue
			}
			month, er := strconv.Atoi(expTime[4:6])
			var months string
			var days string
			if month <= 9 {
				months = fmt.Sprintf("0%d", month)
			} else {
				months = strconv.Itoa(month)
			}
			if er != nil {
				continue
			}
			day, er := strconv.Atoi(expTime[6:])
			if day <= 9 {
				days = fmt.Sprintf("0%d", day)
			} else {
				days = strconv.Itoa(day)
			}
			if er != nil {
				continue
			}
			expiryTime, er = time.Parse("2006-01-02", fmt.Sprintf("%d-%s-%s", year, months, days))
			if er != nil {
				continue
			}
		}
		calloi, er := strconv.ParseFloat(resp.Data[x][2], 64)
		if er != nil {
			continue
		}
		putoi, er := strconv.ParseFloat(resp.Data[x][3], 64)
		if er != nil {
			continue
		}
		callvol, er := strconv.ParseFloat(resp.Data[x][4], 64)
		if er != nil {
			continue
		}
		putvol, er := strconv.ParseFloat(resp.Data[x][5], 64)
		if er != nil {
			continue
		}
		volume := ExpiryOpenInterestAndVolume{
			Timestamp:        time.UnixMilli(int64(timestamp)),
			ExpiryTime:       expiryTime,
			CallOpenInterest: calloi,
			PutOpenInterest:  putoi,
			CallVolume:       callvol,
			PutVolume:        putvol,
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

// GetOpenInterestAndVolumeStrike retrieve the taker volume for both buyers and sellers of calls and puts.
func (ok *Okx) GetOpenInterestAndVolumeStrike(ctx context.Context, currency string,
	expTime time.Time, period kline.Interval) ([]StrikeOpenInterestAndVolume, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	if !expTime.IsZero() {
		var months string
		var days string
		if expTime.Month() <= 9 {
			months = fmt.Sprintf("0%d", expTime.Month())
		} else {
			months = strconv.Itoa(int(expTime.Month()))
		}
		if expTime.Day() <= 9 {
			days = fmt.Sprintf("0%d", expTime.Day())
		} else {
			days = strconv.Itoa(int(expTime.Day()))
		}
		params.Set("expTime", fmt.Sprintf("%d%s%s", expTime.Year(), months, days))
	} else {
		return nil, errMissingExpiryTimeParameter
	}
	type response struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data [][6]string `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(optionOpenInterestVolumeStrike, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getopenInterestAndVolumeEPL, http.MethodGet, path, nil, &resp, false)
	if er != nil {
		return nil, er
	}
	var volumes []StrikeOpenInterestAndVolume
	for x := range resp.Data {
		if len(resp.Data[x]) != 6 {
			continue
		}
		timestamp, er := strconv.Atoi(resp.Data[x][0])
		if er != nil {
			continue
		}
		strike, er := strconv.ParseInt(resp.Data[x][1], 10, 64)
		if er != nil {
			continue
		}
		calloi, er := strconv.ParseFloat(resp.Data[x][2], 64)
		if er != nil {
			continue
		}
		putoi, er := strconv.ParseFloat(resp.Data[x][3], 64)
		if er != nil {
			continue
		}
		callvol, er := strconv.ParseFloat(resp.Data[x][4], 64)
		if er != nil {
			continue
		}
		putvol, er := strconv.ParseFloat(resp.Data[x][5], 64)
		if er != nil {
			continue
		}
		volume := StrikeOpenInterestAndVolume{
			Timestamp:        time.UnixMilli(int64(timestamp)),
			Strike:           strike,
			CallOpenInterest: calloi,
			PutOpenInterest:  putoi,
			CallVolume:       callvol,
			PutVolume:        putvol,
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

// GetTakerFlow shows the relative buy/sell volume for calls and puts.
// It shows whether traders are bullish or bearish on price and volatility
func (ok *Okx) GetTakerFlow(ctx context.Context, currency string, period kline.Interval) (*CurrencyTakerFlow, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	type response struct {
		Msg  string    `json:"msg"`
		Code string    `json:"code"`
		Data [7]string `json:"data"`
	}
	var resp response
	path := common.EncodeURLValues(takerBlockVolume, params)
	er := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTakerFlowEPL, http.MethodGet, path, nil, &resp, false)
	if er != nil {
		return nil, er
	}
	timestamp, era := strconv.ParseInt(resp.Data[0], 10, 64)
	callbuyvol, erb := strconv.ParseFloat(resp.Data[1], 64)
	callselvol, erc := strconv.ParseFloat(resp.Data[2], 64)
	putbutvol, erd := strconv.ParseFloat(resp.Data[3], 64)
	putsellvol, ere := strconv.ParseFloat(resp.Data[4], 64)
	callblockvol, erf := strconv.ParseFloat(resp.Data[5], 64)
	putblockvol, erg := strconv.ParseFloat(resp.Data[6], 64)
	if era != nil || erb != nil || erc != nil || erd != nil || ere != nil || erf != nil || erg != nil {
		return nil, errParsingResponseError
	}
	return &CurrencyTakerFlow{
		Timestamp:       time.UnixMilli(timestamp),
		CallBuyVolume:   callbuyvol,
		CallSellVolume:  callselvol,
		PutBuyVolume:    putbutvol,
		PutSellVolume:   putsellvol,
		CallBlockVolume: callblockvol,
		PutBlockVolume:  putblockvol,
	}, nil
}

// Stauts Endpoints
// https://www.okx.com/docs-v5/en/#rest-api-status
// The response out of this endpoint is an XML file. So, for the timebeing, it is not implemented.

// SendHTTPRequest sends an authenticated http request to a desired
// path with a JSON payload (of present)
// URL arguments must be in the request path and not as url.URL values
func (ok *Okx) SendHTTPRequest(ctx context.Context, ep exchange.URL, f request.EndpointLimit, httpMethod, requestPath string, data, result interface{}, authenticated bool) (err error) {
	endpoint, err := ok.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	var intermediary json.RawMessage
	newRequest := func() (*request.Item, error) {
		utcTime := time.Now().UTC().Format(time.RFC3339)
		payload := []byte("")

		if data != nil {
			payload, err = json.Marshal(data)
			if err != nil {
				return nil, err
			}
		}
		path := endpoint + requestPath
		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"
		if authenticated {
			var creds *account.Credentials
			creds, err = ok.GetCredentials(ctx)
			if err != nil {
				return nil, err
			}
			signPath := fmt.Sprintf("/%v%v", okxAPIPath, requestPath)
			var hmac []byte
			hmac, err = crypto.GetHMAC(crypto.HashSHA256,
				[]byte(utcTime+httpMethod+signPath+string(payload)),
				[]byte(creds.Secret))
			if err != nil {
				return nil, err
			}
			headers["OK-ACCESS-KEY"] = creds.Key
			headers["OK-ACCESS-SIGN"] = crypto.Base64Encode(hmac)
			headers["OK-ACCESS-TIMESTAMP"] = utcTime
			headers["OK-ACCESS-PASSPHRASE"] = creds.ClientID
		}
		return &request.Item{
			Method:        strings.ToUpper(httpMethod),
			Path:          path,
			Headers:       headers,
			Body:          bytes.NewBuffer(payload),
			Result:        &intermediary,
			AuthRequest:   authenticated,
			Verbose:       ok.Verbose,
			HTTPDebugging: ok.HTTPDebugging,
			HTTPRecording: ok.HTTPRecording,
		}, nil
	}
	err = ok.SendPayload(ctx, f, newRequest)
	if err != nil {
		return err
	}
	type errCapFormat struct {
		Error        int64  `json:"error_code,omitempty"`
		ErrorMessage string `json:"error_message,omitempty"`
		Result       bool   `json:"result,string,omitempty"`
	}
	errCap := errCapFormat{Result: true}
	err = json.Unmarshal(intermediary, &errCap)
	if err == nil {
		if errCap.ErrorMessage != "" {
			return fmt.Errorf("error: %v", errCap.ErrorMessage)
		}
		if errCap.Error > 0 {
			return fmt.Errorf("sendHTTPRequest error - %s",
				ErrorCodes[strconv.FormatInt(errCap.Error, 10)])
		}
		if !errCap.Result {
			return errors.New("unspecified error occurred")
		}
	}
	return json.Unmarshal(intermediary, result)
}

// Status

// SystemStatusResponse retrives the system status.
func (ok *Okx) SystemStatusResponse(ctx context.Context, state string) ([]SystemStatusResponse, error) {
	params := url.Values{}
	if strings.EqualFold(state, "scheduled") || strings.EqualFold(state, "ongoing") || strings.EqualFold(state, "pre_open") || strings.EqualFold(state, "completed") || strings.EqualFold(state, "canceled") {
		params.Set("state", state)
	}
	type response struct {
		Code string                 `json:"code"`
		Msg  string                 `json:"msg"`
		Data []SystemStatusResponse `json:"data"`
	}
	path := common.EncodeURLValues(systemStatus, params)
	var resp response
	return resp.Data, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEventStatusEPL, http.MethodGet, path, nil, &resp, true)
}

// GetAssetTypeFromInstrumentType returns an asset Item instance given and Instrument Type string.
func (ok *Okx) GetAssetTypeFromInstrumentType(instrumentType string) (asset.Item, error) {
	switch strings.ToUpper(instrumentType) {
	case OkxInstTypeContract:
		return asset.PerpetualContract, nil
	case OkxInstTypeSwap:
		return asset.PerpetualSwap, nil
	case OkxInstTypeANY:
		return asset.Empty, nil
	default:
		return asset.New(strings.ToLower(instrumentType))
	}
}

// GuessAssetTypeFromInstrumentID returns or guesses the instrument id.
func (ok *Okx) GuessAssetTypeFromInstrumentID(instrumentID string) asset.Item {
	if strings.HasSuffix(instrumentID, OkxInstTypeSwap) {
		return asset.PerpetualSwap
	}
	filter := strings.Split(instrumentID, "-")
	if len(filter) >= 4 {
		return asset.Option
	} else if len(filter) == 3 {
		return asset.Futures
	}
	return asset.Spot
}
