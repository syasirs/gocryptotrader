package okx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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
	okxAPIURL     = "https://www.okx.com/" + okxAPIPath
	okxAPIVersion = "/v5/"

	okxAPIPath      = "api" + okxAPIVersion
	okxWebsocketURL = "wss://ws.okx.com:8443/ws" + okxAPIVersion

	okxAPIWebsocketPublicURL  = okxWebsocketURL + "public"
	okxAPIWebsocketPrivateURL = okxWebsocketURL + "private"

	// tradeEndpoints
	tradeOrder                          = "trade/order"
	placeMultipleOrderURL               = "trade/batch-orders"
	cancelTradeOrder                    = "trade/cancel-order"
	cancelBatchTradeOrders              = "trade/cancel-batch-orders"
	amendOrder                          = "trade/amend-order"
	amendBatchOrders                    = "trade/amend-batch-orders"
	closePositionPath                   = "trade/close-position"
	pendingTradeOrders                  = "trade/orders-pending"
	tradeHistory                        = "trade/orders-history"
	orderHistoryArchive                 = "trade/orders-history-archive"
	tradeFills                          = "trade/fills"
	tradeFillsHistory                   = "trade/fills-history"
	assetBills                          = "asset/bills"
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
	cancelAlgoOrder               = "trade/cancel-algos"
	algoTradeOrder                = "trade/order-algo"
	cancelAdvancedAlgoOrder       = "trade/cancel-advance-algos"
	getAlgoOrders                 = "trade/orders-algo-pending"
	algoOrderHistory              = "trade/orders-algo-history"
	easyConvertCurrencyList       = "trade/easy-convert-currency-list"
	easyConvertHistoryPath        = "trade/easy-convert-history"
	easyConvert                   = "trade/easy-convert"
	oneClickRepayCurrencyListPath = "trade/one-click-repay-currency-list"
	oneClickRepayHistory          = "trade/one-click-repay-history"
	oneClickRepay                 = "trade/one-click-repay"

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

	// Account Endpoints
	accountBalance                   = "account/balance"
	accountPosition                  = "account/positions"
	accountPositionHistory           = "account/positions-history"
	accountAndPositionRisk           = "account/account-position-risk"
	accountBillsDetail               = "account/bills"
	accountBillsDetailArchive        = "account/bills-archive"
	accountConfiguration             = "account/config"
	accountSetPositionMode           = "account/set-position-mode"
	accountSetLeverage               = "account/set-leverage"
	accountMaxSize                   = "account/max-size"
	accountMaxAvailSize              = "account/max-avail-size"
	accountPositionMarginBalance     = "account/position/margin-balance"
	accountLeverageInfo              = "account/leverage-info"
	accountMaxLoan                   = "account/max-loan"
	accountTradeFee                  = "account/trade-fee"
	accountInterestAccrued           = "account/interest-accrued"
	accountInterestRate              = "account/interest-rate"
	accountSetGeeks                  = "account/set-greeks"
	accountSetIsolatedMode           = "account/set-isolated-mode"
	accountMaxWithdrawal             = "account/max-withdrawal"
	accountRiskState                 = "account/risk-state"
	accountBorrowReply               = "account/borrow-repay"
	accountBorrowRepayHistory        = "account/borrow-repay-history"
	accountInterestLimits            = "account/interest-limits"
	accountSimulatedMargin           = "account/simulated_margin"
	accountGeeks                     = "account/greeks"
	accountPortfolioMarginLimitation = "account/position-tiers"

	// Block Trading
	rfqCounterparties       = "rfq/counterparties"
	rfqCreateRFQ            = "rfq/create-rfq"
	rfqCancelRfq            = "rfq/cancel-rfq"
	rfqCancelRfqs           = "rfq/cancel-batch-rfqs"
	rfqCancelAllRfqs        = "rfq/cancel-all-rfqs"
	rfqExecuteQuote         = "rfq/execute-quote"
	makerInstrumentSettings = "rfq/maker-instrument-settings"
	mmpReset                = "rfq/mmp-reset"
	rfqCreateQuote          = "rfq/create-quote"
	rfqCancelQuote          = "rfq/cancel-quote"
	rfqCancelBatchQuotes    = "rfq/cancel-batch-quotes"
	rfqCancelAllQuotes      = "rfq/cancel-all-quotes"
	rfqRfqs                 = "rfq/rfqs"
	rfqQuotes               = "rfq/quotes"
	rfqTrades               = "rfq/trades"
	rfqPublicTrades         = "rfq/public-trades"
	// Subaccount endpoints
	usersSubaccountList          = "users/subaccount/list"
	subAccountModifyAPIKey       = "users/subaccount/modify-apikey"
	accountSubaccountBalances    = "account/subaccount/balances"
	assetSubaccountBalances      = "asset/subaccount/balances"
	assetSubaccountBills         = "asset/subaccount/bills"
	assetSubaccountTransfer      = "asset/subaccount/transfer"
	userSubaccountSetTransferOut = "users/subaccount/set-transfer-out"
	usersEntrustSubaccountList   = "users/entrust-subaccount-list"
	// Grid Trading Endpoints
	gridOrderAlgo            = "tradingBot/grid/order-algo"
	gridAmendOrderAlgo       = "tradingBot/grid/amend-order-algo"
	gridAlgoOrderStop        = "tradingBot/grid/stop-order-algo"
	gridAlgoOrders           = "tradingBot/grid/orders-algo-pending"
	gridAlgoOrdersHistory    = "tradingBot/grid/orders-algo-history"
	gridOrdersAlgoDetails    = "tradingBot/grid/orders-algo-details"
	gridSuborders            = "tradingBot/grid/sub-orders"
	gridPositions            = "tradingBot/grid/positions"
	gridWithdrawalIncome     = "tradingBot/grid/withdraw-income"
	gridComputeMarginBalance = "tradingBot/grid/compute-margin-balance"
	gridMarginBalance        = "tradingBot/grid/margin-balance"
	gridAIParams             = "tradingBot/grid/ai-param"
	// Earn
	financeOffers        = "finance/staking-defi/offers"
	financePurchase      = "finance/staking-defi/purchase"
	financeRedeem        = "finance/staking-defi/redeem"
	financeCacelPurchase = "finance/staking-defi/cancel"
	financeActiveOrders  = "finance/staking-defi/orders-active"
	financeOrdersHistory = "finance/staking-defi/orders-history"
	// Status Endpoints
	systemStatus = "system/status"
)

var (
	// Letters a regular expression for both uppercase and lowercase english characters.
	Letters = regexp.MustCompile(`^[a-zA-Z]+$`)
	// Numbers a regular expression for numbers.
	Numbers = regexp.MustCompile(`^\d+$`)

	errUnableToTypeAssertResponseData                = errors.New("unable to type assert responseData")
	errUnableToTypeAssertKlineData                   = errors.New("unable to type assert kline data")
	errUnexpectedKlineDataLength                     = errors.New("unexpected kline data length")
	errLimitExceedsMaximumResultPerRequest           = errors.New("maximum result per request exceeds the limit")
	errNo24HrTradeVolumeFound                        = errors.New("no trade record found in the 24 trade volume ")
	errOracleInformationNotFound                     = errors.New("oracle informations not found")
	errExchangeInfoNotFound                          = errors.New("exchange information not found")
	errIndexComponentNotFound                        = errors.New("unable to fetch index components")
	errMissingRequiredArgInstType                    = errors.New("invalid required argument instrument type")
	errLimitValueExceedsMaxOf100                     = errors.New("limit value exceeds the maximum value 100")
	errMissingInstrumentID                           = errors.New("missing instrument id")
	errFundingRateHistoryNotFound                    = errors.New("funding rate history not found")
	errMissingRequiredUnderlying                     = errors.New("error missing required parameter underlying")
	errMissingRequiredParamInstID                    = errors.New("missing required parameter instrument id")
	errLiquidationOrderResponseNotFound              = errors.New("liquidation order not found")
	errEitherInstIDOrCcyIsRequired                   = errors.New("either parameter instId or ccy is required")
	errIncorrectRequiredParameterTradeMode           = errors.New("unacceptable required argument, trade mode")
	errInterestRateAndLoanQuotaNotFound              = errors.New("interest rate and loan quota not found")
	errUnderlyingsForSpecifiedInstTypeNofFound       = errors.New("underlyings for the specified instrument id is not found")
	errInsuranceFundInformationNotFound              = errors.New("insurance fund information not found")
	errMissingExpiryTimeParameter                    = errors.New("missing expiry date parameter")
	errInvalidTradeModeValue                         = errors.New("invalid trade mode value")
	errMissingOrderSide                              = errors.New("missing order side")
	errInvalidOrderType                              = errors.New("invalid order type")
	errInvalidQuantityToButOrSell                    = errors.New("unacceptable quantity to buy or sell")
	errMissingClientOrderIDOrOrderID                 = errors.New("client supplier order id or order id is missing")
	errMissingNewSizeOrPriceInformation              = errors.New("missing the new size or price information")
	errMissingNewSize                                = errors.New("missing the order size information")
	errMissingMarginMode                             = errors.New("missing required param margin mode \"mgnMode\"")
	errMissingTradeMode                              = errors.New("missing trade mode")
	errInvalidTriggerPrice                           = errors.New("invalid trigger price value")
	errMssingAlgoOrderID                             = errors.New("missing algo orders id")
	errInvalidPriceLimit                             = errors.New("invalid price limit value")
	errMissingIntervalValue                          = errors.New("missing interval value")
	errMissingTakeProfitTriggerPrice                 = errors.New("missing take profit trigger price")
	errMissingTakeProfitOrderPrice                   = errors.New("missing take profit order price")
	errMissingSizeLimit                              = errors.New("missing required parameter \"size limit\"")
	errMissingEitherAlgoIDOrState                    = errors.New("either algo id or order state is required")
	errUnacceptableAmount                            = errors.New("amount must be greater than 0")
	errInvalidCurrencyValue                          = errors.New("invalid currency value")
	errInvalidDepositAmount                          = errors.New("invalid deposit amount")
	errMissingResponseBody                           = errors.New("error missing response body")
	errMissingValidWithdrawalID                      = errors.New("missing valid withdrawal id")
	errNoValidResponseFromServer                     = errors.New("no valid response from server")
	errInvalidInstrumentType                         = errors.New("invalid instrument type")
	errMissingValidGreeksType                        = errors.New("missing valid greeks type")
	errMissingIsolatedMarginTradingSetting           = errors.New("missing isolated margin trading setting, isolated margin trading settings automatic:Auto transfers autonomy:Manual transfers")
	errInvalidOrderSide                              = errors.New("invalid order side")
	errInvalidCounterParties                         = errors.New("missing counter parties")
	errInvalidLegs                                   = errors.New("no legs are provided")
	errMissingRFQIDANDClientSuppliedRFQID            = errors.New("missing rfq id or client supplied rfq id")
	errMissingRfqIDOrQuoteID                         = errors.New("either RFQ ID or Quote ID is missing")
	errMissingRfqID                                  = errors.New("error missing rfq id")
	errMissingLegs                                   = errors.New("missing legs")
	errMissingSizeOfQuote                            = errors.New("missing size of quote leg")
	errMossingLegsQuotePrice                         = errors.New("error missing quote price")
	errMissingQuoteIDOrClientSuppliedQuoteID         = errors.New("missing quote id or client supplied quote id")
	errMissingEitherQuoteIDAOrClientSuppliedQuoteIDs = errors.New("missing either quote ids or client supplied quote ids")
	errMissingRequiredParameterSubaccountName        = errors.New("missing required parameter subaccount name")
	errInvalidTransferAmount                         = errors.New("unacceptable transfer amount")
	errInvalidInvalidSubaccount                      = errors.New("invalid account type")
	errMissingDestinationSubaccountName              = errors.New("missing destination subaccount name")
	errMissingInitialSubaccountName                  = errors.New("missing initial subaccount name")
	errMissingAlgoOrderType                          = errors.New("missing algo order type \"grid\": Spot grid, \"contract_grid\": Contract grid")
	errInvalidMaximumPrice                           = errors.New("invalid maximum price")
	errInvalidMinimumPrice                           = errors.New("invalid minimum price")
	errInvalidGridQuantity                           = errors.New("invalid grid quantity (grid number)")
	errMissingSize                                   = errors.New("missing required argument, size")
	errMissingRequiredArgumentDirection              = errors.New("missing required argument, direction")
	errRequiredParameterMissingLeverage              = errors.New("missing required parameter, leverage")
	errMissingAlgoOrderID                            = errors.New("missing algo order id")
	errMissingValidStopType                          = errors.New("invalid grid order stop type, only valiues are \"1\" and \"2\" ")
	errMissingSubOrderType                           = errors.New("missing sub order type")
	errMissingQuantity                               = errors.New("invalid quantity to buy or sell")
	errDepositAddressNotFound                        = errors.New("deposit address with the specified currency code and chain not found")
	errMissingAtLeast1CurrencyPair                   = errors.New("at least one currency is required to fetch order history")
	errNoCandlestickDataFound                        = errors.New("no candlesticks data found")
	errInvalidWebsocketEvent                         = errors.New("invalid websocket event")
	errMissingValidChannelInformation                = errors.New("missing channel information")
	errNilArgument                                   = errors.New("nil argument is not acceptable")
	errNoOrderParameterPassed                        = errors.New("no order parameter was passed")
	errEitherPriceVarianceOrPriceSpreadRequired      = errors.New("either \"price ratio\" or \"price variance\" is allowed to be passed")
	errIncompleteCandlestickData                     = errors.New("incomplete candlestick data")
	errMaxRFQOrdersToCancel                          = errors.New("no more than 100 RFQ cancel order parameter is allowed")
	errMalformedData                                 = errors.New("malformed data")
	errInvalidUnderlying                             = errors.New("invalid underlying")
	errMissingRequiredParameter                      = errors.New("missing required parameter")
	errMissingMakerInstrumentSettings                = errors.New("missing maker instrument settings")
	errInvalidSubAccountName                         = errors.New("invalid sub-account name")
	errInvalidAPIKey                                 = errors.New("invalid api key")
	errInvalidAlgoID                                 = errors.New("invalid algo id")
	errInvalidMarginTypeAdjust                       = errors.New("invalid margin type adjust, only 'add' and 'reduce' are allowed")
	errInvalidAlgoOrderType                          = errors.New("invalid algo order type")
	errEmptyArgument                                 = errors.New("empty argument")
	errInvalidCurrencyPair                           = errors.New("invalid currency pair")
	errInvalidIPAddress                              = errors.New("invalid ip address")
	errInvalidAPIKeyPermission                       = errors.New("invalid API Key permission")
	errNoInstrumentFound                             = errors.New("instruments not found")
	errInvalidOrderbookUpdateChecksum                = errors.New("invalid orderbook update checksum")
)

/************************************ MarketData Endpoints *************************************************/

// OrderTypeFromString returns order.Type instance from string
func (ok *Okx) OrderTypeFromString(orderType string) (order.Type, error) {
	switch strings.ToUpper(orderType) {
	case OkxOrderMarket:
		return order.Market, nil
	case OkxOrderLimit:
		return order.Limit, nil
	case OkxOrderPostOnly:
		return order.PostOnly, nil
	case OkxOrderFOK:
		return order.FillOrKill, nil
	case OkxOrderIOC:
		return order.ImmediateOrCancel, nil
	case OkxOrderOptimalLimitIOC:
		return order.OptimalLimitIOC, nil
	default:
		return order.UnknownType, errInvalidOrderType
	}
}

// OrderTypeString returns a string representation of order.Type instance
func (ok *Okx) OrderTypeString(orderType order.Type) (string, error) {
	switch orderType {
	case order.Market:
		return OkxOrderMarket, nil
	case order.Limit:
		return OkxOrderLimit, nil
	case order.PostOnly:
		return OkxOrderPostOnly, nil
	case order.FillOrKill:
		return OkxOrderFOK, nil
	case order.IOS:
		return OkxOrderIOC, nil
	case order.OptimalLimitIOC:
		return OkxOrderOptimalLimitIOC, nil
	default:
		return "", errInvalidOrderType
	}
}

// PlaceOrder place an order only if you have sufficient funds.
func (ok *Okx) PlaceOrder(ctx context.Context, arg *PlaceOrderRequestParam) (*PlaceOrderResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(arg.TradeMode == TradeModeCross || arg.TradeMode == TradeModeIsolated || arg.TradeMode == TradeModeCash) {
		return nil, errInvalidTradeModeValue
	}
	arg.Side = strings.ToUpper(arg.Side)
	if !(arg.Side == order.Buy.String() || arg.Side == order.Sell.String()) {
		return nil, errMissingOrderSide
	}
	arg.OrderType = strings.ToUpper(arg.OrderType)
	if !(arg.OrderType == OkxOrderMarket ||
		arg.OrderType == OkxOrderLimit ||
		arg.OrderType == OkxOrderPostOnly ||
		arg.OrderType == OkxOrderFOK ||
		arg.OrderType == OkxOrderIOC ||
		arg.OrderType == OkxOrderOptimalLimitIOC) {
		return nil, errInvalidOrderType
	}
	if arg.QuantityToBuyOrSell <= 0 {
		return nil, errInvalidQuantityToButOrSell
	}
	if arg.OrderPrice <= 0 && (arg.OrderType == OkxOrderLimit ||
		arg.OrderType == OkxOrderPostOnly ||
		arg.OrderType == OkxOrderFOK ||
		arg.OrderType == OkxOrderIOC) {
		return nil, fmt.Errorf("invalid order price for %s order types", arg.OrderType)
	}
	if !(arg.QuantityType == "base_ccy" ||
		arg.QuantityType == "quote_ccy") {
		arg.QuantityType = ""
	}
	var resp []PlaceOrderResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, placeOrderEPL, http.MethodPost, tradeOrder, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// PlaceMultipleOrders  to place orders in batches. Maximum 20 orders can be placed at a time. Request parameters should be passed in the form of an array.
func (ok *Okx) PlaceMultipleOrders(ctx context.Context, args []PlaceOrderRequestParam) ([]PlaceOrderResponse, error) {
	if len(args) == 0 {
		return nil, errNoOrderParameterPassed
	}
	for x := range args {
		if args[x].InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if !(args[x].TradeMode == TradeModeCross ||
			args[x].TradeMode == TradeModeIsolated ||
			args[x].TradeMode == TradeModeCash) {
			return nil, errInvalidTradeModeValue
		}
		args[x].Side = strings.ToUpper(args[x].Side)
		if !(args[x].Side == order.Buy.String() ||
			args[x].Side == order.Sell.String()) {
			return nil, errMissingOrderSide
		}
		args[x].OrderType = strings.ToLower(args[x].OrderType)
		if !(args[x].OrderType == "market" ||
			args[x].OrderType == "limit" ||
			args[x].OrderType == "post_only" ||
			args[x].OrderType == "fok" ||
			args[x].OrderType == "ioc" ||
			args[x].OrderType == "optimal_limit_ioc") {
			return nil, errInvalidOrderType
		}
		if args[x].QuantityToBuyOrSell <= 0 {
			return nil, errInvalidQuantityToButOrSell
		}
		if args[x].OrderPrice <= 0 && (args[x].OrderType == OkxOrderLimit ||
			args[x].OrderType == OkxOrderPostOnly ||
			args[x].OrderType == OkxOrderFOK ||
			args[x].OrderType == OkxOrderIOC) {
			return nil, fmt.Errorf("invalid order price for %s order types", args[x].OrderType)
		}
		if !(args[x].QuantityType == "base_ccy" || args[x].QuantityType == "quote_ccy") {
			args[x].QuantityType = ""
		}
	}
	var resp []PlaceOrderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeMultipleOrdersEPL, http.MethodPost, placeMultipleOrderURL, &args, &resp, true)
}

// CancelSingleOrder cancel an incomplete order.
func (ok *Okx) CancelSingleOrder(ctx context.Context, arg CancelOrderRequestParam) (*CancelOrderResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
		return nil, fmt.Errorf("either order id or client supplier id is required")
	}
	var resp []CancelOrderResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelOrderEPL, http.MethodPost, cancelTradeOrder, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelMultipleOrders cancel incomplete orders in batches. Maximum 20 orders can be canceled at a time.
// Request parameters should be passed in the form of an array.
func (ok *Okx) CancelMultipleOrders(ctx context.Context, args []CancelOrderRequestParam) ([]CancelOrderResponse, error) {
	for x := range args {
		arg := args[x]
		if arg.InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if arg.OrderID == "" && arg.ClientSupplierOrderID == "" {
			return nil, fmt.Errorf("either order id or client supplier id is required")
		}
	}
	var resp []CancelOrderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleOrdersEPL,
		http.MethodPost, cancelBatchTradeOrders, args, &resp, true)
}

// AmendOrder an incomplete order.
func (ok *Okx) AmendOrder(ctx context.Context, arg *AmendOrderRequestParams) (*AmendOrderResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if arg.ClientSuppliedOrderID == "" && arg.OrderID == "" {
		return nil, errMissingClientOrderIDOrOrderID
	}
	if arg.NewQuantity <= 0 && arg.NewPrice <= 0 {
		return nil, errMissingNewSizeOrPriceInformation
	}
	var resp []AmendOrderResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, amendOrderEPL, http.MethodPost, amendOrder, arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// AmendMultipleOrders amend incomplete orders in batches. Maximum 20 orders can be amended at a time. Request parameters should be passed in the form of an array.
func (ok *Okx) AmendMultipleOrders(ctx context.Context, args []AmendOrderRequestParams) ([]AmendOrderResponse, error) {
	for x := range args {
		if args[x].InstrumentID == "" {
			return nil, errMissingInstrumentID
		}
		if args[x].ClientSuppliedOrderID == "" && args[x].OrderID == "" {
			return nil, errMissingClientOrderIDOrOrderID
		}
		if args[x].NewQuantity <= 0 && args[x].NewPrice <= 0 {
			return nil, errMissingNewSizeOrPriceInformation
		}
	}
	var resp []AmendOrderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, amendMultipleOrdersEPL, http.MethodPost, amendBatchOrders, &args, &resp, true)
}

// ClosePositions Close all positions of an instrument via a market order.
func (ok *Okx) ClosePositions(ctx context.Context, arg *ClosePositionsRequestParams) (*ClosePositionResponse, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(arg.MarginMode != "" &&
		(arg.MarginMode == TradeModeCross ||
			arg.MarginMode == TradeModeIsolated)) {
		return nil, errMissingMarginMode
	}
	var resp []ClosePositionResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, closePositionEPL, http.MethodPost, closePositionPath, arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetOrderDetail retrieves order details given instrument id and order identification
func (ok *Okx) GetOrderDetail(ctx context.Context, arg *OrderDetailRequestParam) (*OrderDetail, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	params := url.Values{}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", arg.InstrumentID)
	switch {
	case arg.OrderID == "" && arg.ClientSupplierOrderID == "":
		return nil, errMissingClientOrderIDOrOrderID
	case arg.ClientSupplierOrderID == "":
		params.Set("ordId", arg.OrderID)
	default:
		params.Set("clOrdId", arg.ClientSupplierOrderID)
	}
	var resp []OrderDetail
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderDetEPL, http.MethodGet, common.EncodeURLValues(tradeOrder, params), nil, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetOrderList retrieves all incomplete orders under the current account.
func (ok *Okx) GetOrderList(ctx context.Context, arg *OrderListRequestParams) ([]OrderDetail, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	params := url.Values{}
	if arg.InstrumentType == okxInstTypeSpot ||
		arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeFutures ||
		arg.InstrumentType == okxInstTypeOption {
		params.Set("instType", arg.InstrumentType)
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if arg.OrderType == OkxOrderMarket ||
		arg.OrderType == OkxOrderLimit ||
		arg.OrderType == OkxOrderPostOnly ||
		arg.OrderType == OkxOrderFOK ||
		arg.OrderType == OkxOrderIOC ||
		arg.OrderType == OkxOrderOptimalLimitIOC {
		params.Set("orderType", arg.OrderType)
	}
	arg.State = strings.ToUpper(arg.State)
	if arg.State == order.Cancelled.String() ||
		arg.State == order.Filled.String() {
		params.Set("state", arg.State)
	}
	if !arg.Before.IsZero() {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !arg.After.IsZero() {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	var resp []OrderDetail
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderListEPL, http.MethodGet, common.EncodeURLValues(pendingTradeOrders, params), nil, &resp, true)
}

// Get7DayOrderHistory retrieves the completed order data for the last 7 days, and the incomplete orders that have been cancelled are only reserved for 2 hours.
func (ok *Okx) Get7DayOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams) ([]OrderDetail, error) {
	return ok.getOrderHistory(ctx, arg, tradeHistory)
}

// Get3MonthOrderHistory retrieves the completed order data for the last 7 days, and the incomplete orders that have been cancelled are only reserved for 2 hours.
func (ok *Okx) Get3MonthOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams) ([]OrderDetail, error) {
	return ok.getOrderHistory(ctx, arg, orderHistoryArchive)
}

// getOrderHistory retrieves the order history of the past limited times
func (ok *Okx) getOrderHistory(ctx context.Context, arg *OrderHistoryRequestParams, route string) ([]OrderDetail, error) {
	params := url.Values{}
	if arg.InstrumentType == okxInstTypeSpot ||
		arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeFutures ||
		arg.InstrumentType == okxInstTypeOption {
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
	arg.OrderType = strings.ToUpper(arg.OrderType)
	if arg.OrderType == OkxOrderMarket ||
		arg.OrderType == OkxOrderLimit ||
		arg.OrderType == OkxOrderPostOnly ||
		arg.OrderType == OkxOrderFOK ||
		arg.OrderType == OkxOrderIOC ||
		arg.OrderType == OkxOrderOptimalLimitIOC {
		params.Set("orderType", arg.OrderType)
	}
	if arg.State == order.Cancelled.String() ||
		arg.State == order.Filled.String() {
		params.Set("state", arg.State)
	}
	if !arg.Before.IsZero() {
		params.Set("before", strconv.FormatInt(arg.Before.UnixMilli(), 10))
	}
	if !arg.After.IsZero() {
		params.Set("after", strconv.FormatInt(arg.After.UnixMilli(), 10))
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	if arg.Category == "twap" || arg.Category == "adl" ||
		arg.Category == "full_liquidation" || arg.Category == "partial_liquidation" ||
		arg.Category == "delivery" || arg.Category == "ddh" {
		params.Set("category", strings.ToLower(arg.Category))
	}
	var resp []OrderDetail
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderHistoryEPL, http.MethodGet, common.EncodeURLValues(route, params), nil, &resp, true)
}

// GetTransactionDetailsLast3Days retrieves recently-filled transaction details in the last 3 day.
func (ok *Okx) GetTransactionDetailsLast3Days(ctx context.Context, arg *TransactionDetailRequestParams) ([]TransactionDetail, error) {
	return ok.getTransactionDetails(ctx, arg, tradeFills)
}

// GetTransactionDetailsLast3Months Retrieve recently-filled transaction details in the last 3 months.
func (ok *Okx) GetTransactionDetailsLast3Months(ctx context.Context, arg *TransactionDetailRequestParams) ([]TransactionDetail, error) {
	return ok.getTransactionDetails(ctx, arg, tradeFillsHistory)
}

// GetTransactionDetails retrieves recently-filled transaction details.
func (ok *Okx) getTransactionDetails(ctx context.Context, arg *TransactionDetailRequestParams, route string) ([]TransactionDetail, error) {
	params := url.Values{}
	arg.InstrumentType = strings.ToUpper(arg.InstrumentType)
	if arg.InstrumentType == okxInstTypeSpot ||
		arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeFutures ||
		arg.InstrumentType == okxInstTypeOption {
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
	if !arg.Begin.IsZero() {
		params.Set("begin", strconv.FormatInt(arg.Begin.UnixMilli(), 10))
	}
	if !arg.End.IsZero() {
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
	var resp []TransactionDetail
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTrasactionDetailsEPL, http.MethodGet, common.EncodeURLValues(route, params), nil, &resp, true)
}

// PlaceAlgoOrder order includes trigger order, oco order, conditional order,iceberg order, twap order and trailing order.
func (ok *Okx) PlaceAlgoOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	arg.TradeMode = strings.ToLower(arg.TradeMode)
	if !(arg.TradeMode == TradeModeCross ||
		arg.TradeMode == TradeModeIsolated) {
		return nil, errMissingTradeMode
	}
	if !(arg.Side == order.Buy ||
		arg.Side == order.Sell) {
		return nil, errMissingOrderSide
	}
	if !(arg.OrderType == "conditional" ||
		arg.OrderType == "oco" ||
		arg.OrderType == "trigger" ||
		arg.OrderType == "move_order_stop" ||
		arg.OrderType == "iceberg" ||
		arg.OrderType == "twap") {
		return nil, errInvalidOrderType
	}
	if arg.Size <= 0 {
		return nil, errMissingNewSize
	}
	var resp []AlgoOrder
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, placeAlgoOrderEPL, http.MethodGet, algoTradeOrder, arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// PlaceStopOrder to place stop order
func (ok *Okx) PlaceStopOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.OrderType != "conditional" {
		return nil, errInvalidOrderType
	}
	if arg.TakeProfitTriggerPrice == 0 {
		return nil, errMissingTakeProfitTriggerPrice
	}
	if arg.TakeProfitTriggerPriceType == "" {
		return nil, errMissingTakeProfitOrderPrice
	}
	if !(arg.TakeProfitTriggerPriceType != "" &&
		(arg.TakeProfitTriggerPriceType == "index" ||
			arg.TakeProfitTriggerPriceType == "last" ||
			arg.TakeProfitTriggerPriceType == "mark")) {
		arg.TakeProfitTriggerPriceType = ""
	}
	if !(arg.StopLossTriggerPriceType != "" &&
		(arg.StopLossTriggerPriceType == "index" ||
			arg.StopLossTriggerPriceType == "last" ||
			arg.StopLossTriggerPriceType == "mark")) {
		arg.StopLossTriggerPriceType = ""
	}
	return ok.PlaceAlgoOrder(ctx, arg)
}

// PlaceTrailingStopOrder to place trailing stop order
func (ok *Okx) PlaceTrailingStopOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.OrderType != "move_order_stop" {
		return nil, errInvalidOrderType
	}
	if arg.CallbackRatio == 0 &&
		arg.CallbackSpreadVariance == "" {
		return nil, errors.New("either \"callbackRatio\" or \"callbackSpread\" is allowed to be passed")
	}
	return ok.PlaceAlgoOrder(ctx, arg)
}

// PlaceIcebergOrder to place iceburg algo order
func (ok *Okx) PlaceIcebergOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.OrderType != "iceberg" {
		return nil, errInvalidOrderType
	}
	if arg.PriceVariance == "" && arg.PriceSpread == "" {
		return nil, errEitherPriceVarianceOrPriceSpreadRequired
	}
	if arg.SizeLimit <= 0 {
		return nil, errMissingSizeLimit
	}
	if arg.PriceLimit <= 0 {
		return nil, errInvalidPriceLimit
	}
	return ok.PlaceAlgoOrder(ctx, arg)
}

// PlaceTWAPOrder to place TWAP algo orders
func (ok *Okx) PlaceTWAPOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg.PriceVariance == "" && arg.PriceSpread == "" {
		return nil, errEitherPriceVarianceOrPriceSpreadRequired
	}
	if arg.OrderType != "twap" {
		return nil, errInvalidOrderType
	}
	if arg.SizeLimit <= 0 {
		return nil, errors.New("missing required parameter \"size limit\"")
	}
	if arg.PriceLimit <= 0 {
		return nil, errInvalidPriceLimit
	}
	if ok.GetIntervalEnum(arg.TimeInterval) == "" {
		return nil, errMissingIntervalValue
	}
	return ok.PlaceAlgoOrder(ctx, arg)
}

// TriggerAlgoOrder fetches algo trigger orders for SWAP market types.
func (ok *Okx) TriggerAlgoOrder(ctx context.Context, arg *AlgoOrderParams) (*AlgoOrder, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.OrderType != "trigger" {
		return nil, errInvalidOrderType
	}
	if arg.TriggerPrice <= 0 {
		return nil, errInvalidTriggerPrice
	}
	if !(arg.TriggerPriceType == "last" ||
		arg.TriggerPriceType == "index" ||
		arg.TriggerPriceType == "mark") {
		arg.TriggerPriceType = ""
	}
	return ok.PlaceAlgoOrder(ctx, arg)
}

// CancelAdvanceAlgoOrder Cancel unfilled algo orders
// A maximum of 10 orders can be canceled at a time.
// Request parameters should be passed in the form of an array.
func (ok *Okx) CancelAdvanceAlgoOrder(ctx context.Context, args []AlgoOrderCancelParams) ([]AlgoOrder, error) {
	return ok.cancelAlgoOrder(ctx, args, cancelAdvancedAlgoOrder)
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
		return nil, errors.New("no parameter")
	}
	var resp []AlgoOrder
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAlgoOrderEPL, http.MethodPost, route, args, &resp, true)
}

// GetAlgoOrderList retrieves a list of untriggered Algo orders under the current account.
func (ok *Okx) GetAlgoOrderList(ctx context.Context, orderType, algoOrderID, clientSuppliedOrderID, instrumentType, instrumentID string, after, before time.Time, limit uint) ([]AlgoOrderResponse, error) {
	params := url.Values{}
	orderType = strings.ToLower(orderType)
	if !(orderType == "conditional" ||
		orderType == "oco" ||
		orderType == "trigger" ||
		orderType == "move_order_stop" ||
		orderType == "iceberg" ||
		orderType == "twap") {
		return nil, fmt.Errorf("invalid order type value %s,%s,%s,%s,%s,and %s", "conditional", "oco", "trigger", "move_order_stop", "iceberg", "twap")
	}
	params.Set("ordType", orderType)
	var resp []AlgoOrderResponse
	if algoOrderID != "" {
		params.Set("algoId", algoOrderID)
	}
	if clientSuppliedOrderID != "" {
		params.Set("clOrdId", clientSuppliedOrderID)
	}
	if instrumentType == okxInstTypeSpot ||
		instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeFutures ||
		instrumentType == okxInstTypeOption {
		params.Set("instType", instrumentType)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if !before.IsZero() && before.Before(time.Now()) {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() && after.After(time.Now()) {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAlgoOrderListEPL, http.MethodGet, common.EncodeURLValues(getAlgoOrders, params), nil, &resp, true)
}

// GetAlgoOrderHistory load a list of all algo orders under the current account in the last 3 months.
func (ok *Okx) GetAlgoOrderHistory(ctx context.Context, orderType, state, algoOrderID, instrumentType, instrumentID string, after, before time.Time, limit uint) ([]AlgoOrderResponse, error) {
	params := url.Values{}
	orderType = strings.ToLower(strings.Trim(orderType, " "))
	if !(orderType == "conditional" || orderType == "oco" || orderType == "trigger" || orderType == "move_order_stop" ||
		orderType == "iceberg" || orderType == "twap") {
		return nil, fmt.Errorf("invalid order type value %s,%s,%s,%s,%s,and %s", "conditional", "oco", "trigger", "move_order_stop", "iceberg", "twap")
	}
	params.Set("ordType", orderType)
	var resp []AlgoOrderResponse
	if algoOrderID == "" &&
		!(state == "effective" ||
			state == "order_failed" ||
			state == "canceled") {
		return nil, errMissingEitherAlgoIDOrState
	}
	if algoOrderID != "" {
		params.Set("algoId", algoOrderID)
	} else {
		params.Set("state", state)
	}
	if instrumentType == okxInstTypeSpot || instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeFutures || instrumentType == okxInstTypeOption {
		params.Set("instType", instrumentType)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if !before.IsZero() && before.Before(time.Now()) {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() && after.After(time.Now()) {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAlgoOrderHistoryEPL, http.MethodGet, common.EncodeURLValues(algoOrderHistory, params), nil, &resp, true)
}

// GetEasyConvertCurrencyList retrieve list of small convertibles and mainstream currencies. Only applicable to the crypto balance less than $10.
func (ok *Okx) GetEasyConvertCurrencyList(ctx context.Context) (*EasyConvertDetail, error) {
	var resp []EasyConvertDetail
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getEasyConvertCurrencyListRate, http.MethodGet,
		easyConvertCurrencyList, nil, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// PlaceEasyConvert onvert small currencies to mainstream currencies. Only applicable to the crypto balance less than $10.
func (ok *Okx) PlaceEasyConvert(ctx context.Context, arg PlaceEasyConvertParam) ([]EasyConvertItem, error) {
	if len(arg.FromCurrency) == 0 {
		return nil, fmt.Errorf("%v, missing 'fromCcy'", errMissingRequiredParameter)
	}
	if arg.ToCurrency == "" {
		return nil, fmt.Errorf("%v, missing t'toCcy'", errMissingRequiredParameter)
	}
	var resp []EasyConvertItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, placeEasyConvertEPL, http.MethodPost, easyConvert, &arg, &resp, true)
}

// GetEasyConvertHistory retrives the history and status of easy convert trades.
func (ok *Okx) GetEasyConvertHistory(ctx context.Context, after, before time.Time, limit int) ([]EasyConvertItem, error) {
	params := url.Values{}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.Unix(), 10))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.Unix(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var resp []EasyConvertItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEasyConvertHistoryEPL, http.MethodGet, easyConvertHistoryPath, nil, &resp, true)
}

// GetOneClickRepayCurrencyList retrives list of debt currency data and repay currencies. Debt currencies include both cross and isolated debts.
func (ok *Okx) GetOneClickRepayCurrencyList(ctx context.Context, debtType string) ([]CurrencyOneClickRepay, error) {
	params := url.Values{}
	if debtType == "cross" || debtType == "isolated" {
		params.Set("debtType", debtType)
	}
	var resp []CurrencyOneClickRepay
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, oneClickRepayCurrencyListEPL, http.MethodGet, oneClickRepayCurrencyListPath, nil, &resp, true)
}

// TradeOneClickRepay trade one-click repay to repay cross debts. Isolated debts are not applicable. The maximum repayment amount is based on the remaining available balance of funding and trading accounts.
func (ok *Okx) TradeOneClickRepay(ctx context.Context, arg TradeOneClickRepayParam) ([]CurrencyOneClickRepay, error) {
	if len(arg.DebtCurrency) == 0 {
		return nil, fmt.Errorf("%v, missing 'debtCcy'", errMissingRequiredParameter)
	}
	if arg.RepayCurrency == "" {
		return nil, fmt.Errorf("%v, missing 'repayCcy'", errMissingRequiredParameter)
	}
	var resp []CurrencyOneClickRepay
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, tradeOneClickRepayEPL, http.MethodPost, oneClickRepay, &arg, &resp, true)
}

// GetOneClickRepayHistory get the history and status of one-click repay trades.
func (ok *Okx) GetOneClickRepayHistory(ctx context.Context, after, before time.Time, limit int) ([]CurrencyOneClickRepay, error) {
	params := url.Values{}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.Unix(), 10))
	}
	if !after.IsZero() {
		params.Set("after", strconv.FormatInt(after.Unix(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var resp []CurrencyOneClickRepay
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOneClickRepayHistoryEPL, http.MethodGet, common.EncodeURLValues(oneClickRepayHistory, params), nil, &resp, true)
}

/*************************************** Block trading ********************************/

// GetCounterparties retrieves the list of counterparties that the user has permissions to trade with.
func (ok *Okx) GetCounterparties(ctx context.Context) ([]CounterpartiesResponse, error) {
	var resp []CounterpartiesResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCounterpartiesEPL, http.MethodGet, rfqCounterparties, nil, &resp, true)
}

// CreateRFQ Creates a new RFQ
func (ok *Okx) CreateRFQ(ctx context.Context, arg CreateRFQInput) (*RFQResponse, error) {
	if len(arg.CounterParties) == 0 {
		return nil, errInvalidCounterParties
	}
	if len(arg.Legs) == 0 {
		return nil, errInvalidLegs
	}
	var resp []RFQResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, createRfqEPL, http.MethodPost, rfqCreateRFQ, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelRFQ Cancel an existing active RFQ that you has previously created.
func (ok *Okx) CancelRFQ(ctx context.Context, arg CancelRFQRequestParam) (*CancelRFQResponse, error) {
	if arg.RfqID == "" && arg.ClientSuppliedRFQID == "" {
		return nil, errMissingRFQIDANDClientSuppliedRFQID
	}
	var resp []CancelRFQResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelRfqEPL, http.MethodPost, rfqCancelRfq, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelMultipleRFQs cancel multiple active RFQs in a single batch. Maximum 100 RFQ orders can be canceled at a time.
func (ok *Okx) CancelMultipleRFQs(ctx context.Context, arg CancelRFQRequestsParam) ([]CancelRFQResponse, error) {
	if len(arg.RfqID) == 0 && len(arg.ClientSuppliedRFQID) == 0 {
		return nil, errMissingRFQIDANDClientSuppliedRFQID
	} else if len(arg.RfqID)+len(arg.ClientSuppliedRFQID) > 100 {
		return nil, errMaxRFQOrdersToCancel
	}
	var resp []CancelRFQResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleRfqEPL, http.MethodPost, rfqCancelRfqs, &arg, &resp, true)
}

// CancelAllRFQs cancels all active RFQs.
func (ok *Okx) CancelAllRFQs(ctx context.Context) (time.Time, error) {
	var resp []TimestampResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAllRfqsEPL, http.MethodPost, rfqCancelAllRfqs, nil, &resp, true)
	if err != nil {
		return time.Time{}, err
	}
	if len(resp) == 1 {
		return resp[0].Timestamp, nil
	}
	return time.Time{}, errNoValidResponseFromServer
}

// ExecuteQuote executes a Quote. It is only used by the creator of the RFQ
func (ok *Okx) ExecuteQuote(ctx context.Context, arg ExecuteQuoteParams) (*ExecuteQuoteResponse, error) {
	if arg.RfqID == "" || arg.QuoteID == "" {
		return nil, errMissingRfqIDOrQuoteID
	}
	var resp []ExecuteQuoteResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, executeQuoteEPL, http.MethodPost, rfqExecuteQuote, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// SetQuoteProducts customize the products which makers want to quote and receive RFQs for, and the corresponding price and size limit.
func (ok *Okx) SetQuoteProducts(ctx context.Context, args []SetQuoteProductParam) (*SetQuoteProductsResult, error) {
	if len(args) == 0 {
		return nil, errEmptyArgument
	}
	for x := range args {
		if !(args[x].InstrumentType == okxInstTypeSwap ||
			args[x].InstrumentType == okxInstTypeSpot ||
			args[x].InstrumentType == okxInstTypeFutures ||
			args[x].InstrumentType == okxInstTypeOption) {
			return nil, errInvalidInstrumentType
		}
		if len(args) == 0 {
			return nil, errMissingMakerInstrumentSettings
		}
		for y := range args[x].Data {
			if (args[x].InstrumentType == okxInstTypeSwap ||
				args[x].InstrumentType == okxInstTypeFutures ||
				args[x].InstrumentType == okxInstTypeOption) && args[x].Data[y].Underlying == "" {
				return nil, fmt.Errorf("%v, for instrument type %s and %s", errInvalidUnderlying, args[x].InstrumentType, args[x].Data[x].Underlying)
			}
			if (args[x].InstrumentType == okxInstTypeSpot) && args[x].Data[x].InstrumentID == "" {
				return nil, fmt.Errorf("%v, for instrument type %s and %s", errMissingInstrumentID, args[x].InstrumentType, args[x].Data[x].InstrumentID)
			}
		}
	}
	var resp []SetQuoteProductsResult
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, setQuoteProductsEPL, http.MethodPost, makerInstrumentSettings, &args, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// ResetMMPStatus reset the MMP status to be inactive.
func (ok *Okx) ResetMMPStatus(ctx context.Context) (time.Time, error) {
	var resp []TimestampResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, restMMPStatusEPL, http.MethodPost, mmpReset, nil, &resp, true); err != nil {
		return time.Time{}, err
	}
	if len(resp) == 1 {
		return resp[0].Timestamp, nil
	}
	return time.Time{}, errNoValidResponseFromServer
}

// CreateQuote allows the user to Quote an RFQ that they are a counterparty to. The user MUST quote
// the entire RFQ and not part of the legs or part of the quantity. Partial quoting or partial fills are not allowed.
func (ok *Okx) CreateQuote(ctx context.Context, arg CreateQuoteParams) (*QuoteResponse, error) {
	switch {
	case arg.RfqID == "":
		return nil, errMissingRfqID
	case !(arg.QuoteSide == order.Buy || arg.QuoteSide == order.Sell):
		return nil, errMissingOrderSide
	case len(arg.Legs) == 0:
		return nil, errMissingLegs
	}
	for x := range arg.Legs {
		switch {
		case arg.Legs[x].InstrumentID == "":
			return nil, errMissingInstrumentID
		case arg.Legs[x].SizeOfQuoteLeg <= 0:
			return nil, errMissingSizeOfQuote
		case arg.Legs[x].Price <= 0:
			return nil, errMossingLegsQuotePrice
		case arg.Legs[x].Side == order.UnknownSide:
			return nil, errMissingOrderSide
		}
	}
	var resp []QuoteResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, createQuoteEPL, http.MethodPost, rfqCreateQuote, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelQuote cancels an existing active quote you have created in response to an RFQ.
// rfqCancelQuote = "rfq/cancel-quote"
func (ok *Okx) CancelQuote(ctx context.Context, arg CancelQuoteRequestParams) (*CancelQuoteResponse, error) {
	var resp []CancelQuoteResponse
	if arg.ClientSuppliedQuoteID == "" && arg.QuoteID == "" {
		return nil, errMissingQuoteIDOrClientSuppliedQuoteID
	}
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelQuoteEPL, http.MethodPost, rfqCancelQuote, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelMultipleQuote cancel multiple active Quotes in a single batch. Maximum 100 quote orders can be canceled at a time.
func (ok *Okx) CancelMultipleQuote(ctx context.Context, arg CancelQuotesRequestParams) ([]CancelQuoteResponse, error) {
	if len(arg.QuoteIDs) == 0 && len(arg.ClientSuppliedQuoteIDs) == 0 {
		return nil, errMissingEitherQuoteIDAOrClientSuppliedQuoteIDs
	}
	var resp []CancelQuoteResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelMultipleQuotesEPL, http.MethodPost, rfqCancelBatchQuotes, &arg, &resp, true)
}

// CancelAllQuotes cancels all active Quotes.
func (ok *Okx) CancelAllQuotes(ctx context.Context) (time.Time, error) {
	var resp []TimestampResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelAllQuotesEPL, http.MethodPost, rfqCancelAllQuotes, nil, &resp, true)
	if err != nil {
		return time.Time{}, err
	}
	if len(resp) == 1 {
		return resp[0].Timestamp, nil
	}
	return time.Time{}, errMissingResponseBody
}

// GetRfqs retrieves details of RFQs that the user is a counterparty to (either as the creator or the receiver of the RFQ).
func (ok *Okx) GetRfqs(ctx context.Context, arg *RfqRequestParams) ([]RFQResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	params := url.Values{}
	if arg.RfqID != "" {
		params.Set("rfqId", arg.RfqID)
	}
	if arg.ClientSuppliedRfqID != "" {
		params.Set("clRfqId", arg.ClientSuppliedRfqID)
	}
	arg.State = strings.ToUpper(arg.State)
	if arg.State == order.Active.String() ||
		arg.State == order.Cancelled.String() ||
		arg.State == "PENDING_FILL" ||
		arg.State == order.Filled.String() ||
		arg.State == order.Expired.String() ||
		arg.State == "TRADED_AWAY" ||
		arg.State == "FAILED" {
		params.Set("state", strings.ToLower(arg.State))
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
	var resp []RFQResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getRfqsEPL, http.MethodGet, common.EncodeURLValues(rfqRfqs, params), nil, &resp, true)
}

// GetQuotes retrieves all Quotes that the user is a counterparty to (either as the creator or the receiver).
func (ok *Okx) GetQuotes(ctx context.Context, arg *QuoteRequestParams) ([]QuoteResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
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
	arg.State = strings.ToUpper(arg.State)
	if arg.State == order.Active.String() ||
		arg.State == order.Cancelled.String() ||
		arg.State == "PENDING_FILL" ||
		arg.State == order.Filled.String() ||
		arg.State == order.Expired.String() ||
		arg.State == "FAILED" {
		params.Set("state", strings.ToLower(arg.State))
	}
	if arg.BeginID != "" {
		params.Set("beginId", arg.BeginID)
	}
	if arg.EndID != "" {
		params.Set("endId", arg.EndID)
	}
	if arg.Limit > 0 {
		params.Set("limit", strconv.Itoa(arg.Limit))
	}
	var resp []QuoteResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getQuotesEPL, http.MethodGet, common.EncodeURLValues(rfqQuotes, params), nil, &resp, true)
}

// GetRFQTrades retrieves the executed trades that the user is a counterparty to (either as the creator or the receiver).
func (ok *Okx) GetRFQTrades(ctx context.Context, arg *RFQTradesRequestParams) ([]RfqTradeResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
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
	if arg.State == order.Active.String() ||
		arg.State == "CANCELED" ||
		arg.State == "PENDING_FILL" ||
		arg.State == order.Filled.String() ||
		arg.State == order.Expired.String() ||
		arg.State == "TRADED_AWAY" ||
		arg.State == "FAILED" {
		params.Set("state", strings.ToLower(arg.State))
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
	var resp []RfqTradeResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesEPL, http.MethodGet, common.EncodeURLValues(rfqTrades, params), nil, &resp, true)
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
	var resp []PublicTradesResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPublicTradesEPL, http.MethodGet, common.EncodeURLValues(rfqPublicTrades, params), nil, &resp, true)
}

/*************************************** Funding Tradings ********************************/

// GetCurrencies Retrieve a list of all currencies.
func (ok *Okx) GetCurrencies(ctx context.Context) ([]CurrencyResponse, error) {
	var resp []CurrencyResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCurrenciesEPL, http.MethodGet, assetCurrencies, nil, &resp, true)
}

// GetBalance retrieves the balances of all the assets and the amount that is available or on hold.
func (ok *Okx) GetBalance(ctx context.Context, currency string) ([]AssetBalance, error) {
	var resp []AssetBalance
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBalanceEPL, http.MethodGet, common.EncodeURLValues(assetBalance, params), nil, &resp, true)
}

// GetAccountAssetValuation view account asset valuation
func (ok *Okx) GetAccountAssetValuation(ctx context.Context, currency string) ([]AccountAssetValuation, error) {
	params := url.Values{}
	currency = strings.ToUpper(currency)
	if currency == "BTC" || currency == "USDT" ||
		currency == "USD" || currency == "CNY" ||
		currency == "JPY" || currency == "KRW" ||
		currency == "RUB" || currency == "EUR" ||
		currency == "VND" || currency == "IDR" ||
		currency == "INR" || currency == "PHP" ||
		currency == "THB" || currency == "TRY" ||
		currency == "AUD" || currency == "SGD" ||
		currency == "ARS" || currency == "SAR" ||
		currency == "AED" || currency == "IQD" {
		params.Set("ccy", currency)
	}
	var resp []AccountAssetValuation
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountAssetValuationEPL, http.MethodGet, common.EncodeURLValues(assetValuation, params), nil, &resp, true)
}

// FundingTransfer transfer of funds between your funding account and trading account,
// and from the master account to sub-accounts.
func (ok *Okx) FundingTransfer(ctx context.Context, arg *FundingTransferRequestInput) ([]FundingTransferResponse, error) {
	var resp []FundingTransferResponse
	if arg == nil {
		return nil, errors.New("argument can not be null")
	}
	if arg.Amount <= 0 {
		return nil, errors.New("invalid funding amount")
	}
	if arg.Currency == "" {
		return nil, errors.New("invalid currency value")
	}
	if !(arg.From == "6" || arg.From == "18") {
		return nil, errors.New("missing funding source field \"From\", only '6' and '18' are supported")
	}
	if arg.To == "" {
		return nil, errors.New("missing funding destination field \"To\", only '6' and '18' are supported")
	}
	if arg.From == arg.To {
		return nil, errors.New("parameter 'from' can not equal to parameter 'to'")
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
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, fundsTransferEPL, http.MethodPost, assetTransfer, arg, &resp, true)
}

// GetFundsTransferState get funding rate response.
func (ok *Okx) GetFundsTransferState(ctx context.Context, transferID, clientID string, transferType int) ([]TransferFundRateResponse, error) {
	params := url.Values{}
	switch {
	case transferID == "" && clientID == "":
		return nil, errors.New("either 'transfer id' or 'client id' is required")
	case transferID != "":
		params.Set("transId", transferID)
	case clientID != "":
		if !(Letters.MatchString(clientID) || Numbers.MatchString(clientID)) {
			return nil, errors.New("invalid client id")
		}
		params.Set("clientId", clientID)
	}
	if transferType > 0 && transferType <= 4 {
		params.Set("type", strconv.Itoa(transferType))
	}
	var resp []TransferFundRateResponse
	println(common.EncodeURLValues(assetTransferState, params))
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundsTransferStateEPL, http.MethodGet, common.EncodeURLValues(assetTransferState, params), nil, &resp, true)
}

// GetAssetBillsDetails Query the billing record, you can get the latest 1 month historical data
func (ok *Okx) GetAssetBillsDetails(ctx context.Context, currency string, billType int, clientID, clientSecret string, after, before time.Time, limit int) ([]AssetBillDetail, error) {
	params := url.Values{}
	if billType == 1 || billType == 2 ||
		billType == 13 || billType == 20 ||
		billType == 21 || billType == 28 ||
		billType == 41 || billType == 42 ||
		(billType >= 47 && billType <= 54) ||
		(billType >= 59 && billType <= 61) ||
		billType == 68 || billType == 69 ||
		(billType >= 72 && billType < 131) ||
		billType == 150 || billType == 151 ||
		billType == 198 || billType == 199 {
		params.Set("type", strconv.Itoa(billType))
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
	var resp []AssetBillDetail
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, assetBillsDetailsEPL, http.MethodGet, common.EncodeURLValues(assetBills, params), nil, &resp, true)
}

// GetLightningDeposits users can create up to 10 thousand different invoices within 24 hours.
// this method fetches list of lightning deposits filtered by a currency and amount.
func (ok *Okx) GetLightningDeposits(ctx context.Context, currency string, amount float64, to int) ([]LightningDepositItem, error) {
	params := url.Values{}
	if currency == "" {
		return nil, errInvalidCurrencyValue
	}
	params.Set("ccy", currency)
	if amount <= 0 {
		return nil, errInvalidDepositAmount
	}
	params.Set("amt", strconv.FormatFloat(amount, 'f', 0, 64))
	if to == 6 || to == 18 {
		params.Set("to", strconv.Itoa(to))
	}
	var resp []LightningDepositItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, ligntningDepositsEPL, http.MethodGet, common.EncodeURLValues(lightningDeposit, params), nil, &resp, true)
}

// GetCurrencyDepositAddress returns the deposit address and related informations for the provided currency information.
func (ok *Okx) GetCurrencyDepositAddress(ctx context.Context, currency string) ([]CurrencyDepositResponseItem, error) {
	params := url.Values{}
	if currency == "" {
		return nil, errInvalidCurrencyValue
	}
	params.Set("ccy", currency)
	var resp []CurrencyDepositResponseItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDepositAddressEPL, http.MethodGet, common.EncodeURLValues(assetDeposits, params), nil, &resp, true)
}

// GetCurrencyDepositHistory retrieves deposit records and withdrawal status information depending on the currency, timestamp, and chronological order.
func (ok *Okx) GetCurrencyDepositHistory(ctx context.Context, currency, depositID, transactionID string, state int, after, before time.Time, limit uint) ([]DepositHistoryResponseItem, error) {
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
	var resp []DepositHistoryResponseItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDepositHistoryEPL, http.MethodGet, common.EncodeURLValues(pathToAssetDepositHistory, params), nil, &resp, true)
}

// Withdrawal to perform a withdrawal action. Sub-account does not support withdrawal.
func (ok *Okx) Withdrawal(ctx context.Context, input *WithdrawalInput) (*WithdrawalResponse, error) {
	if input == nil {
		return nil, errNilArgument
	}
	var resp []WithdrawalResponse
	switch {
	case input.Currency == "":
		return nil, errInvalidCurrencyValue
	case input.Amount <= 0:
		return nil, errors.New("invalid withdrawal amount")
	case input.WithdrawalDestination == "":
		return nil, errors.New("withdrawal destination")
	case input.ToAddress == "":
		return nil, errors.New("missing verified digital currency address \"toAddr\" information")
	}
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, withdrawalEPL, http.MethodPost, assetWithdrawal, &input, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

/*
 This API function service is only open to some users. If you need this function service, please send an email to `liz.jensen@okg.com` to apply
*/

// LightningWithdrawal to withdraw a currency from an invoice.
func (ok *Okx) LightningWithdrawal(ctx context.Context, arg LightningWithdrawalRequestInput) (*LightningWithdrawalResponse, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if arg.Invoice == "" {
		return nil, errors.New("missing invoice text")
	}
	var resp []LightningWithdrawalResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, lightningWithdrawalsEPL, http.MethodPost, assetLightningWithdrawal, &arg, &resp, true)
	if err != nil {
		return nil, err
	} else if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
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
	var output inout
	return output.WithdrawalID, ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelWithdrawalEPL, http.MethodPost, cancelWithdrawal, input, &output, true)
}

// GetWithdrawalHistory retrieves the withdrawal records according to the currency, withdrawal status, and time range in reverse chronological order.
// The 100 most recent records are returned by default.
func (ok *Okx) GetWithdrawalHistory(ctx context.Context, currency, withdrawalID, clientID, transactionID string, state int, after, before time.Time, limit int) ([]WithdrawalHistoryResponse, error) {
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
	var resp []WithdrawalHistoryResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getWithdrawalHistoryEPL, http.MethodGet, common.EncodeURLValues(withdrawalHistory, params), nil, &resp, true)
}

// SmallAssetsConvert Convert small assets in funding account to OKB. Only one convert is allowed within 24 hours.
func (ok *Okx) SmallAssetsConvert(ctx context.Context, currency []string) (*SmallAssetConvertResponse, error) {
	input := map[string][]string{"ccy": currency}
	var resp []SmallAssetConvertResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, smallAssetsConvertEPL, http.MethodPost, smallAssetsConvert, input, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetSavingBalance returns saving balance, and only assets in the funding account can be used for saving.
func (ok *Okx) GetSavingBalance(ctx context.Context, currency string) ([]SavingBalanceResponse, error) {
	var resp []SavingBalanceResponse
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSavingBalanceEPL, http.MethodGet, common.EncodeURLValues(assetSavingBalance, params), nil, &resp, true)
}

// SavingsPurchaseOrRedemption creates a purchase or redemption instance
func (ok *Okx) SavingsPurchaseOrRedemption(ctx context.Context, arg *SavingsPurchaseRedemptionInput) (*SavingsPurchaseRedemptionResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	arg.ActionType = strings.ToLower(arg.ActionType)
	switch {
	case arg.Currency == "":
		return nil, errInvalidCurrencyValue
	case arg.Amount <= 0:
		return nil, errUnacceptableAmount
	case !(arg.ActionType == "purchase" || arg.ActionType == "redempt"):
		return nil, errors.New("invalid side value, side has to be either \"redemption\" or \"purchase\"")
	case arg.ActionType == "purchase" && !(arg.Rate >= 1 && arg.Rate <= 365):
		return nil, errors.New("the rate value range is between 1% and 365%")
	}
	var resp []SavingsPurchaseRedemptionResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, savingsPurchaseRedemptionEPL, http.MethodPost, assetSavingPurchaseOrRedemptionPath, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
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
	var resp []LendingHistory
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, setLendingRateEPL, http.MethodGet, common.EncodeURLValues(assetsLendingHistory, params), nil, &resp, true)
}

// SetLendingRate sets assets Lending Rate
func (ok *Okx) SetLendingRate(ctx context.Context, arg LendingRate) (*LendingRate, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	} else if !(arg.Rate >= 1 && arg.Rate <= 365) {
		return nil, errors.New("invalid lending rate value. the rate value range is between 1% and 365%")
	}
	var resp []LendingRate
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLendingHistoryEPL, http.MethodPost, assetSetLendingRateRoute, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetPublicBorrowInfo return list of publix borrow history.
func (ok *Okx) GetPublicBorrowInfo(ctx context.Context, currency string) ([]PublicBorrowInfo, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []PublicBorrowInfo
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPublicBorrowInfoEPL, http.MethodGet, common.EncodeURLValues(publicBorrowInfo, params), nil, &resp, false)
}

/***********************************Convert Endpoints | Authenticated s*****************************************/

// GetConvertCurrencies retrieves the currency conversion information.
func (ok *Okx) GetConvertCurrencies(ctx context.Context) ([]ConvertCurrency, error) {
	var resp []ConvertCurrency
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertCurrenciesEPL, http.MethodGet, assetConvertCurrencies, nil, &resp, true)
}

// GetConvertCurrencyPair retrieves the currency conversion response detail given the 'currency from' and 'currency to'
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
	var resp []ConvertCurrencyPair
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertCurrencyPairEPL, http.MethodGet, common.EncodeURLValues(convertCurrencyPairsPath, params), nil, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// EstimateQuote retrieves quote estimation detail result given the base and quote currency.
func (ok *Okx) EstimateQuote(ctx context.Context, arg *EstimateQuoteRequestInput) (*EstimateQuoteResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.BaseCurrency == "" {
		return nil, errors.New("missing base currency")
	}
	if arg.QuoteCurrency == "" {
		return nil, errors.New("missing quote currency")
	}
	arg.Side = strings.ToUpper(arg.Side)
	if !(arg.Side == order.Buy.String() || arg.Side == order.Sell.String()) {
		return nil, errors.New("missing  order side")
	}
	if arg.RFQAmount <= 0 {
		return nil, errors.New("missing rfq amount")
	}
	if arg.RFQSzCurrency == "" {
		return nil, errors.New("missing rfq currency")
	}
	var resp []EstimateQuoteResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, estimateQuoteEPL, http.MethodPost, assetEstimateQuote, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// ConvertTrade converts a base currency to quote currency.
func (ok *Okx) ConvertTrade(ctx context.Context, arg *ConvertTradeInput) (*ConvertTradeResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.BaseCurrency == "" {
		return nil, errors.New("missing base currency")
	}
	if arg.QuoteCurrency == "" {
		return nil, errors.New("missing quote currency")
	}
	arg.Side = strings.ToUpper(arg.Side)
	if !(arg.Side == order.Buy.String() ||
		arg.Side == order.Sell.String()) {
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
	var resp []ConvertTradeResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, convertTradeEPL, http.MethodPost, assetConvertTrade, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
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
	var resp []ConvertHistory
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getConvertHistoryEPL, http.MethodGet, common.EncodeURLValues(assetConvertHistory, params), nil, &resp, true)
}

/********************************** Account endpoints ***************************************************/

// GetNonZeroBalances retrieves a list of assets (with non-zero balance), remaining balance, and available amount in the trading account.
// Interest-free quota and discount rates are public data and not displayed on the account interface.
func (ok *Okx) GetNonZeroBalances(ctx context.Context, currency string) ([]Account, error) {
	var resp []Account
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountBalanceEPL, http.MethodGet, common.EncodeURLValues(accountBalance, params), nil, &resp, true)
}

// GetPositions retrieves information on your positions. When the account is in net mode, net positions will be displayed, and when the account is in long/short mode, long or short positions will be displayed.
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
	var resp []AccountPosition
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionsEPL, http.MethodGet, common.EncodeURLValues(accountPosition, params), nil, &resp, true)
}

// GetPositionsHistory retrieves the updated position data for the last 3 months.
func (ok *Okx) GetPositionsHistory(ctx context.Context, instrumentType, instrumentID, marginMode string, closePositionType uint, after, before time.Time, limit uint) ([]AccountPositionHistory, error) {
	params := url.Values{}
	if strings.EqualFold(instrumentType, okxInstTypeMargin) || strings.EqualFold(instrumentType, okxInstTypeSwap) || strings.EqualFold(instrumentType, okxInstTypeFutures) ||
		strings.EqualFold(instrumentType, okxInstTypeOption) {
		params.Set("instType", strings.ToUpper(instrumentType))
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if strings.EqualFold(marginMode, TradeModeCross) || strings.EqualFold(marginMode, TradeModeIsolated) {
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
	var resp []AccountPositionHistory
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionsHistoryEPL, http.MethodGet, common.EncodeURLValues(accountPositionHistory, params), nil, &resp, true)
}

// GetAccountAndPositionRisk  get account and position risks.
func (ok *Okx) GetAccountAndPositionRisk(ctx context.Context, instrumentType string) ([]AccountAndPositionRisk, error) {
	params := url.Values{}
	if instrumentType != "" {
		params.Set("instType", instrumentType)
	}
	var resp []AccountAndPositionRisk
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountAndPositionRiskEPL, http.MethodGet, common.EncodeURLValues(accountAndPositionRisk, params), nil, &resp, true)
}

// GetBillsDetailLast7Days The bill refers to all transaction records that result in changing the balance of an account. Pagination is supported, and the response is sorted with the most recent first. This endpoint can retrieves data from the last 7 days.
func (ok *Okx) GetBillsDetailLast7Days(ctx context.Context, arg *BillsDetailQueryParameter) ([]BillsDetailResponse, error) {
	return ok.GetBillsDetail(ctx, arg, accountBillsDetail)
}

// GetBillsDetail3Months retrieves the account’s bills.
// The bill refers to all transaction records that result in changing the balance of an account.
// Pagination is supported, and the response is sorted with most recent first.
// This endpoint can retrieves data from the last 3 months.
func (ok *Okx) GetBillsDetail3Months(ctx context.Context, arg *BillsDetailQueryParameter) ([]BillsDetailResponse, error) {
	return ok.GetBillsDetail(ctx, arg, accountBillsDetailArchive)
}

// GetBillsDetail retrieves the bills of the account.
func (ok *Okx) GetBillsDetail(ctx context.Context, arg *BillsDetailQueryParameter, route string) ([]BillsDetailResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	params := url.Values{}
	if arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeOption ||
		arg.InstrumentType == okxInstTypeFutures {
		params.Set("instType", strings.ToUpper(arg.InstrumentType))
	}
	if arg.Currency != "" {
		params.Set("ccy", strings.ToUpper(arg.Currency))
	}
	if arg.MarginMode == TradeModeIsolated || arg.MarginMode == TradeModeCross {
		params.Set("mgnMode", arg.MarginMode)
	}
	if arg.ContractType == "linear" || arg.ContractType == "inverse" {
		params.Set("ctType", arg.ContractType)
	}
	if arg.BillType >= 1 && arg.BillType <= 13 {
		params.Set("type", strconv.Itoa(int(arg.BillType)))
	}
	if (arg.BillSubType >= 1 && arg.BillSubType <= 6) || arg.BillSubType == 9 || arg.BillSubType == 11 ||
		arg.BillSubType == 12 || arg.BillSubType == 14 || (arg.BillSubType >= 100 && arg.BillSubType <= 106) ||
		(arg.BillSubType >= 110 && arg.BillSubType <= 113) || (arg.BillSubType >= 117 && arg.BillSubType <= 119) ||
		(arg.BillSubType >= 125 && arg.BillSubType <= 128) || (arg.BillSubType >= 131 && arg.BillSubType <= 132) ||
		(arg.BillSubType >= 160 && arg.BillSubType <= 162) || (arg.BillSubType >= 170 && arg.BillSubType <= 174) ||
		(arg.BillSubType >= 200 && arg.BillSubType <= 203) {
		params.Set("subType", strconv.Itoa(int(arg.BillSubType)))
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
	var resp []BillsDetailResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBillsDetailsEPL, http.MethodGet, common.EncodeURLValues(route, params), nil, &resp, true)
}

// GetAccountConfiguration retrieves current account configuration.
func (ok *Okx) GetAccountConfiguration(ctx context.Context) ([]AccountConfigurationResponse, error) {
	var resp []AccountConfigurationResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountConfigurationEPL, http.MethodGet, accountConfiguration, nil, &resp, true)
}

// SetPositionMode FUTURES and SWAP support both long/short mode and net mode. In net mode, users can only have positions in one direction; In long/short mode, users can hold positions in long and short directions.
func (ok *Okx) SetPositionMode(ctx context.Context, positionMode string) (string, error) {
	if !(positionMode == "long_short_mode" || positionMode == "net_mode") {
		return "", errors.New("missing position mode, \"long_short_mode\": long/short, only applicable to FUTURES/SWAP \"net_mode\": net")
	}
	input := &PositionMode{
		PositionMode: positionMode,
	}
	var resp []PositionMode
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, setPositionModeEPL, http.MethodPost, accountSetPositionMode, input, &resp, true)
	if err != nil {
		return "", err
	}
	if len(resp) == 1 {
		return resp[0].PositionMode, nil
	}
	return "", errNoValidResponseFromServer
}

// SetLeverage sets a leverage setting for instrument id.
func (ok *Okx) SetLeverage(ctx context.Context, arg SetLeverageInput) (*SetLeverageResponse, error) {
	if arg.InstrumentID == "" && arg.Currency == "" {
		return nil, errors.New("either instrument id or currency is missing")
	}
	if arg.Leverage < 0 {
		return nil, errors.New("missing leverage")
	}
	if !(arg.MarginMode == TradeModeIsolated || arg.MarginMode == TradeModeCross) {
		return nil, errors.New("invalid margin mode")
	} else if arg.MarginMode == TradeModeCross {
		arg.PositionSide = ""
	}
	if arg.InstrumentID == "" && arg.MarginMode == TradeModeIsolated {
		return nil, errors.New("only can be cross if ccy is passed")
	}
	if !(arg.MarginMode == TradeModeCross || arg.MarginMode == TradeModeIsolated) {
		return nil, errors.New("only applicable to \"isolated\" margin mode of FUTURES/SWAP")
	}
	arg.PositionSide = strings.ToLower(arg.PositionSide)
	if !(arg.PositionSide == "long" ||
		arg.PositionSide == "short") &&
		arg.MarginMode == "isolated" {
		return nil, errors.New("\"long\" \"short\" Only applicable to isolated margin mode of FUTURES/SWAP")
	} else if arg.MarginMode != "isolated" {
		arg.MarginMode = ""
	}
	var resp []SetLeverageResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, setLeverateEPL, http.MethodPost, accountSetLeverage, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetMaximumBuySellAmountOROpenAmount retrieves the maximum buy or sell amount for a specific instrument id
func (ok *Okx) GetMaximumBuySellAmountOROpenAmount(ctx context.Context, instrumentID, tradeMode, currency, leverage string, price float64) ([]MaximumBuyAndSell, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errors.New("missing instrument id")
	}
	params.Set("instId", instrumentID)
	if !(tradeMode == TradeModeCross || tradeMode == TradeModeIsolated || tradeMode == TradeModeCash) {
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
	var resp []MaximumBuyAndSell
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumBuyOrSellAmountEPL, http.MethodGet, common.EncodeURLValues(accountMaxSize, params), nil, &resp, true)
}

// GetMaximumAvailableTradableAmount retrieves the maximum tradable amount for specific instrument id, and/or currency
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
	if !(tradeMode == TradeModeIsolated ||
		tradeMode == TradeModeCross ||
		tradeMode == TradeModeCash) {
		return nil, errors.New("missing trade mode")
	}
	params.Set("tdMode", tradeMode)
	params.Set("px", strconv.FormatFloat(price, 'f', 0, 64))
	var resp []MaximumTradableAmount
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumAvailableTradableAmountEPL, http.MethodGet, common.EncodeURLValues(accountMaxAvailSize, params), nil, &resp, true)
}

// IncreaseDecreaseMargin Increase or decrease the margin of the isolated position. Margin reduction may result in the change of the actual leverage.
func (ok *Okx) IncreaseDecreaseMargin(ctx context.Context, arg IncreaseDecreaseMarginInput) (*IncreaseDecreaseMargin, error) {
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if !(arg.PositionSide == "long" ||
		arg.PositionSide == "short" ||
		arg.PositionSide == "net") {
		return nil, errors.New("missing valid position side")
	}
	if !(arg.Type == "add" || arg.Type == "reduce") {
		return nil, errors.New("missing valid 'type', use 'add': add margin 'reduce': reduce margin")
	}
	if arg.Amount <= 0 {
		return nil, errors.New("missing valid amount")
	}
	var resp []IncreaseDecreaseMargin
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, increaseOrDecreaseMarginEPL, http.MethodGet, accountPositionMarginBalance, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetLeverage retrieves leverage data for different instrument id or margin mode.
func (ok *Okx) GetLeverage(ctx context.Context, instrumentID, marginMode string) ([]LeverageResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("invalid instrument id \"instId\" ")
	}
	if marginMode == TradeModeCross || marginMode == TradeModeIsolated {
		params.Set("mgnMode", marginMode)
	} else {
		return nil, errors.New("missing margin mode \"mgnMode\"")
	}
	var resp []LeverageResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getLeverateEPL, http.MethodGet, common.EncodeURLValues(accountLeverageInfo, params), nil, &resp, true)
}

// GetMaximumLoanOfInstrument returns list of maximum loan of instruments.
func (ok *Okx) GetMaximumLoanOfInstrument(ctx context.Context, instrumentID, marginMode, mgnCurrency string) ([]MaximumLoanInstrument, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("invalid instrument id \"instId\"")
	}
	if marginMode == TradeModeCross || marginMode == TradeModeIsolated {
		params.Set("mgnMode", marginMode)
	} else {
		return nil, errors.New("missing margin mode \"mgnMode\"")
	}
	if mgnCurrency != "" {
		params.Set("mgnCcy", mgnCurrency)
	}
	var resp []MaximumLoanInstrument
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTheMaximumLoanOfInstrumentEPL, http.MethodGet, common.EncodeURLValues(accountMaxLoan, params), nil, &resp, true)
}

// GetFee returns Cryptocurrency trade fee, and offline trade fee
func (ok *Okx) GetFee(ctx context.Context, feeBuilder *exchange.FeeBuilder) (float64, error) {
	// Here the Asset Type for the instrument Type is needed for getting the CryptocurrencyTrading Fee.
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		var responses []TradeFeeRate
		uly, err := ok.GetUnderlying(feeBuilder.Pair, asset.Spot)
		if err != nil {
			return 0, err
		}
		responses, err = ok.GetTradeFee(ctx, okxInstTypeSpot, uly, "")
		if err != nil {
			return 0, err
		} else if len(responses) == 0 {
			return 0, errors.New("no trade fee response found")
		}
		if feeBuilder.IsMaker {
			if fee, err = strconv.ParseFloat(responses[0].FeeRateMaker, 64); err != nil || fee == 0 {
				fee, err = strconv.ParseFloat(responses[0].FeeRateMakerUSDT, 64)
				if err != nil {
					return fee, err
				}
			}
		} else {
			if fee, err = strconv.ParseFloat(responses[0].FeeRateTaker, 64); err != nil || fee == 0 {
				fee, err = strconv.ParseFloat(responses[0].FeeRateTakerUSDT, 64)
				if err != nil {
					return fee, err
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

// GetTradeFee query trade fee rate of various instrument types and instrument ids.
func (ok *Okx) GetTradeFee(ctx context.Context, instrumentType, instrumentID, underlying string) ([]TradeFeeRate, error) {
	params := url.Values{}
	instrumentType = strings.ToUpper(instrumentType)
	if !(instrumentType == okxInstTypeSpot || instrumentType == okxInstTypeMargin || instrumentType == okxInstTypeSwap || instrumentType == okxInstTypeFutures || instrumentType == okxInstTypeOption) {
		return nil, errInvalidInstrumentType
	}
	params.Set("instType", strings.ToUpper(instrumentType))
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if underlying != "" {
		params.Set("uly", underlying)
	}
	var resp []TradeFeeRate
	println(common.EncodeURLValues(accountTradeFee, params))
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFeeRatesEPL, http.MethodGet, common.EncodeURLValues(accountTradeFee, params), nil, &resp, true)
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
	marginMode = strings.ToLower(marginMode)
	if marginMode == TradeModeCross ||
		marginMode == TradeModeIsolated {
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
	var resp []InterestAccruedData
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestAccruedDataEPL, http.MethodGet, common.EncodeURLValues(accountInterestAccrued, params), nil, &resp, true)
}

// GetInterestRate get the user's current leveraged currency borrowing interest rate
func (ok *Okx) GetInterestRate(ctx context.Context, currency string) ([]InterestRateResponse, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []InterestRateResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateEPL, http.MethodGet, common.EncodeURLValues(accountInterestRate, params), nil, &resp, true)
}

// SetGreeks set the display type of Greeks. PA: Greeks in coins BS: Black-Scholes Greeks in dollars
func (ok *Okx) SetGreeks(ctx context.Context, greeksType string) (*GreeksType, error) {
	greeksType = strings.ToUpper(greeksType)
	if !(greeksType == "PA" || greeksType == "BS") {
		return nil, errMissingValidGreeksType
	}
	input := &GreeksType{
		GreeksType: greeksType,
	}
	var resp []GreeksType
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, setGeeksEPL, http.MethodPost, accountSetGeeks, input, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// IsolatedMarginTradingSettings to set the currency margin and futures/perpetual Isolated margin trading mode.
func (ok *Okx) IsolatedMarginTradingSettings(ctx context.Context, arg IsolatedMode) (*IsolatedMode, error) {
	arg.IsoMode = strings.ToLower(arg.IsoMode)
	if !(arg.IsoMode == "automatic" ||
		arg.IsoMode == "autonomy") {
		return nil, errMissingIsolatedMarginTradingSetting
	}
	if !(arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeContract) {
		return nil, errMissingInstrumentID
	}
	var resp []IsolatedMode
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, isolatedMarginTradingSettingsEPL, http.MethodPost, accountSetIsolatedMode, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetMaximumWithdrawals retrieves the maximum transferable amount from trading account to funding account.
func (ok *Okx) GetMaximumWithdrawals(ctx context.Context, currency string) ([]MaximumWithdrawal, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []MaximumWithdrawal
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMaximumWithdrawalsEPL, http.MethodGet, common.EncodeURLValues(accountMaxWithdrawal, params), nil, &resp, true)
}

// GetAccountRiskState gets the account risk status.
// only applicable to Portfolio margin account
func (ok *Okx) GetAccountRiskState(ctx context.Context) ([]AccountRiskState, error) {
	var resp []AccountRiskState
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getAccountRiskStateEPL, http.MethodGet, accountRiskState, nil, &resp, true)
}

// VIPLoansBorrowAndRepay creates VIP borrow or repay for a currency.
func (ok *Okx) VIPLoansBorrowAndRepay(ctx context.Context, arg LoanBorrowAndReplayInput) (*LoanBorrowAndReplay, error) {
	if arg.Currency == "" {
		return nil, errInvalidCurrencyValue
	}
	if arg.Side == "" {
		return nil, errInvalidOrderSide
	}
	if arg.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	var resp []LoanBorrowAndReplay
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, vipLoansBorrowAnsRepayEPL, http.MethodPost, accountBorrowReply, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetBorrowAndRepayHistoryForVIPLoans retrieves borrow and repay history for VIP loans.
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
	var resp []BorrowRepayHistory
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBorrowAnsRepayHistoryHistoryEPL, http.MethodGet, common.EncodeURLValues(accountBorrowRepayHistory, params), nil, &resp, true)
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
	var resp []BorrowInterestAndLimitResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBorrowInterestAndLimitEPL, http.MethodGet, common.EncodeURLValues(accountInterestLimits, params), nil, &resp, true)
}

// PositionBuilder calculates portfolio margin information for simulated position or current position of the user. You can add up to 200 simulated positions in one request.
func (ok *Okx) PositionBuilder(ctx context.Context, arg PositionBuilderInput) ([]PositionBuilderResponse, error) {
	arg.InstrumentType = strings.ToUpper(arg.InstrumentType)
	if !(arg.InstrumentType == okxInstTypeSwap || arg.InstrumentType == okxInstTypeFutures || arg.InstrumentType == okxInstTypeOption) {
		arg.InstrumentType = ""
	}
	var resp []PositionBuilderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, positionBuilderEPL, http.MethodPost, accountSimulatedMargin, &arg, &resp, true)
}

// GetGreeks retrieves a greeks list of all assets in the account.
func (ok *Okx) GetGreeks(ctx context.Context, currency string) ([]GreeksItem, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []GreeksItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGeeksEPL, http.MethodGet, common.EncodeURLValues(accountGeeks, params), nil, &resp, true)
}

// GetPMLimitation retrieve cross position limitation of SWAP/FUTURES/OPTION under Portfolio margin mode.
func (ok *Okx) GetPMLimitation(ctx context.Context, instrumentType, underlying string) ([]PMLimitationResponse, error) {
	params := url.Values{}
	if instrumentType == "" {
		return nil, errInvalidInstrumentType
	}
	if underlying == "" {
		return nil, errInvalidUnderlying
	}
	params.Set("instType", instrumentType)
	params.Set("uly", underlying)
	var resp []PMLimitationResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPMLimitationEPL, http.MethodGet, common.EncodeURLValues(accountPortfolioMarginLimitation, params), nil, &resp, true)
}

/********************************** Subaccount Endpoints ***************************************************/

// ViewSubAccountList applies to master accounts only
func (ok *Okx) ViewSubAccountList(ctx context.Context, enable bool, subaccountName string, after, before time.Time, limit int) ([]SubaccountInfo, error) {
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
		params.Set("limit", strconv.Itoa(limit))
	}
	var resp []SubaccountInfo
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, viewSubaccountListEPL, http.MethodGet, common.EncodeURLValues(usersSubaccountList, params), nil, &resp, true)
}

// ResetSubAccountAPIKey applies to master accounts only and master accounts APIKey must be linked to IP addresses.
func (ok *Okx) ResetSubAccountAPIKey(ctx context.Context, arg *SubAccountAPIKeyParam) (*SubAccountAPIKeyResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.SubAccountName == "" {
		return nil, errInvalidSubAccountName
	}
	if arg.APIKey == "" {
		return nil, errInvalidAPIKey
	}
	var resp []SubAccountAPIKeyResponse
	if arg.IP != "" && !common.MatchesIPV4Address(arg.IP) {
		return nil, errInvalidIPAddress
	}
	if !(arg.APIKeyPermission == "read" || arg.APIKeyPermission == "withdraw" || arg.APIKeyPermission == "trade" || arg.APIKeyPermission == "read_only") {
		return nil, errInvalidAPIKeyPermission
	}
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, resetSubAccountAPIKeyEPL, http.MethodPost, subAccountModifyAPIKey, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetSubaccountTradingBalance query detailed balance info of Trading Account of a sub-account via the master account (applies to master accounts only)
func (ok *Okx) GetSubaccountTradingBalance(ctx context.Context, subaccountName string) ([]SubaccountBalanceResponse, error) {
	params := url.Values{}
	if subaccountName == "" {
		return nil, errMissingRequiredParameterSubaccountName
	}
	params.Set("subAcct", subaccountName)
	var resp []SubaccountBalanceResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSubaccountTradingBalanceEPL, http.MethodGet, common.EncodeURLValues(accountSubaccountBalances, params), nil, &resp, true)
}

// GetSubaccountFundingBalance query detailed balance info of Funding Account of a sub-account via the master account (applies to master accounts only)
func (ok *Okx) GetSubaccountFundingBalance(ctx context.Context, subaccountName, currency string) ([]FundingBalance, error) {
	params := url.Values{}
	if subaccountName == "" {
		return nil, errMissingRequiredParameterSubaccountName
	}
	params.Set("subAcct", subaccountName)
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []FundingBalance
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSubaccountFundingBalanceEPL, http.MethodGet, common.EncodeURLValues(assetSubaccountBalances, params), nil, &resp, true)
}

// HistoryOfSubaccountTransfer retrieves subaccount transfer histories; applies to master accounts only.
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
	var resp []SubaccountBillItem
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, historyOfSubaccountTransferEPL, http.MethodGet, common.EncodeURLValues(assetSubaccountBills, params), nil, &resp, true)
}

// MasterAccountsManageTransfersBetweenSubaccounts master accounts manage the transfers between sub-accounts applies to master accounts only
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
	var resp []TransferIDInfo
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, masterAccountsManageTransfersBetweenSubaccountEPL, http.MethodGet, common.EncodeURLValues(assetSubaccountTransfer, params), nil, &resp, true)
}

// SetPermissionOfTransferOut set permission of transfer out for sub-account(only applicable to master account). Sub-account can transfer out to master account by default.
func (ok *Okx) SetPermissionOfTransferOut(ctx context.Context, arg PermissingOfTransfer) ([]PermissingOfTransfer, error) {
	if arg.SubAcct == "" {
		return nil, errMissingRequiredParameterSubaccountName
	}
	var resp []PermissingOfTransfer
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, setPermissingOfTransferOutEPL, http.MethodPost, userSubaccountSetTransferOut, &arg, &resp, true)
}

// GetCustodyTradingSubaccountList the trading team uses this interface to view the list of sub-accounts currently under escrow
// usersEntrustSubaccountList ="users/entrust-subaccount-list"
func (ok *Okx) GetCustodyTradingSubaccountList(ctx context.Context, subaccountName string) ([]SubaccountName, error) {
	params := url.Values{}
	if subaccountName != "" {
		params.Set("setAcct", subaccountName)
	}
	var resp []SubaccountName
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getCustoryTradingSubaccountListEPL, http.MethodGet, common.EncodeURLValues(usersEntrustSubaccountList, params), nil, &resp, true)
}

/*************************************** Grid Trading Endpoints ***************************************************/

// PlaceGridAlgoOrder place spot grid algo order.
func (ok *Okx) PlaceGridAlgoOrder(ctx context.Context, arg *GridAlgoOrder) (*GridAlgoOrderIDResponse, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	arg.AlgoOrdType = strings.ToLower(arg.AlgoOrdType)
	if !(arg.AlgoOrdType == AlgoOrdTypeGrid || arg.AlgoOrdType == AlgoOrdTypeContractGrid) {
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
		arg.Direction = strings.ToLower(arg.Direction)
		if !(arg.Direction == "long" || arg.Direction == "short" || arg.Direction == "neutral") {
			return nil, errMissingRequiredArgumentDirection
		}
		if arg.Lever == "" {
			return nil, errRequiredParameterMissingLeverage
		}
	}
	var resp []GridAlgoOrderIDResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, gridTradingEPL, http.MethodPost, gridOrderAlgo, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// AmendGridAlgoOrder supported contract grid algo order amendment.
func (ok *Okx) AmendGridAlgoOrder(ctx context.Context, arg GridAlgoOrderAmend) (*GridAlgoOrderIDResponse, error) {
	if arg.AlgoID == "" {
		return nil, errMissingAlgoOrderID
	}
	if arg.InstrumentID == "" {
		return nil, errMissingInstrumentID
	}
	var resp []GridAlgoOrderIDResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, amendGridAlgoOrderEPL, http.MethodPost, gridAmendOrderAlgo, &arg, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
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
		arg[x].AlgoOrderType = strings.ToLower(arg[x].AlgoOrderType)
		if !(arg[x].AlgoOrderType == AlgoOrdTypeGrid || arg[x].AlgoOrderType == AlgoOrdTypeContractGrid) {
			return nil, errMissingAlgoOrderType
		}
		if !(arg[x].StopType == 1 || arg[x].StopType == 2) {
			return nil, errMissingValidStopType
		}
	}
	var resp []GridAlgoOrderIDResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, stopGridAlgoOrderEPL, http.MethodPost, gridAlgoOrderStop, arg, &resp, true)
}

// GetGridAlgoOrdersList retrieves list of pending grid algo orders with the complete data.
func (ok *Okx) GetGridAlgoOrdersList(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	return ok.getGridAlgoOrders(ctx, algoOrderType, algoID,
		instrumentID, instrumentType,
		after, before, limit, gridAlgoOrders)
}

// GetGridAlgoOrderHistory retrieves list of grid algo orders with the complete data including the stoped orders.
func (ok *Okx) GetGridAlgoOrderHistory(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	return ok.getGridAlgoOrders(ctx, algoOrderType, algoID,
		instrumentID, instrumentType,
		after, before, limit, gridAlgoOrdersHistory)
}

// getGridAlgoOrderList retrieves list of grid algo orders with the complete data.
func (ok *Okx) getGridAlgoOrders(ctx context.Context, algoOrderType, algoID,
	instrumentID, instrumentType,
	after, before string, limit uint, route string) ([]GridAlgoOrderResponse, error) {
	params := url.Values{}
	algoOrderType = strings.ToLower(algoOrderType)
	if !(algoOrderType == AlgoOrdTypeGrid || algoOrderType == AlgoOrdTypeContractGrid) {
		return nil, errMissingAlgoOrderType
	}
	params.Set("algoOrdType", algoOrderType)
	if algoID != "" {
		params.Set("algoId", algoID)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	if instrumentType == okxInstTypeSpot || instrumentType == okxInstTypeMargin ||
		instrumentType == okxInstTypeFutures || instrumentType == okxInstTypeSwap {
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
	epl := getGridAlgoOrderListEPL
	if route == gridAlgoOrdersHistory {
		epl = getGridAlgoOrderHistoryEPL
	}
	var resp []GridAlgoOrderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, epl, http.MethodGet, common.EncodeURLValues(route, params), nil, &resp, true)
}

// GetGridAlgoOrderDetails retrieves grid algo order details
func (ok *Okx) GetGridAlgoOrderDetails(ctx context.Context, algoOrderType, algoID string) (*GridAlgoOrderResponse, error) {
	params := url.Values{}
	if !(algoOrderType == AlgoOrdTypeGrid ||
		algoOrderType == AlgoOrdTypeContractGrid) {
		return nil, errMissingAlgoOrderType
	}
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	params.Set("algoOrdType", algoOrderType)
	params.Set("algoId", algoID)
	var resp []GridAlgoOrderResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoOrderDetailsEPL, http.MethodGet, common.EncodeURLValues(gridOrdersAlgoDetails, params), nil, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetGridAlgoSubOrders retrieves grid algo sub orders
func (ok *Okx) GetGridAlgoSubOrders(ctx context.Context, algoOrderType, algoID, subOrderType, groupID, after, before string, limit uint) ([]GridAlgoOrderResponse, error) {
	params := url.Values{}
	if !(algoOrderType == AlgoOrdTypeGrid ||
		algoOrderType == AlgoOrdTypeContractGrid) {
		return nil, errMissingAlgoOrderType
	}
	params.Set("algoOrdType", algoOrderType)
	if algoID != "" {
		params.Set("algoId", algoID)
	} else {
		return nil, errMissingAlgoOrderID
	}
	subOrderType = strings.ToUpper(subOrderType)
	if subOrderType == "LIVE" || subOrderType == order.Filled.String() {
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
	var resp []GridAlgoOrderResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoSubOrdersEPL, http.MethodGet, common.EncodeURLValues(gridSuborders, params), nil, &resp, true)
}

// GetGridAlgoOrderPositions retrieves grid algo order positions.
func (ok *Okx) GetGridAlgoOrderPositions(ctx context.Context, algoOrderType, algoID string) ([]AlgoOrderPosition, error) {
	params := url.Values{}
	if !(algoOrderType == AlgoOrdTypeGrid || algoOrderType == AlgoOrdTypeContractGrid) {
		return nil, errMissingAlgoOrderType
	}
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	params.Set("algoOrdType", algoOrderType)
	params.Set("algoId", algoID)
	var resp []AlgoOrderPosition
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAlgoOrderPositionsEPL, http.MethodGet, common.EncodeURLValues(gridPositions, params), nil, &resp, true)
}

// SpotGridWithdrawProfit returns the spot grid orders withdrawal profit given an instrument id.
func (ok *Okx) SpotGridWithdrawProfit(ctx context.Context, algoID string) (*AlgoOrderWithdrawalProfit, error) {
	if algoID == "" {
		return nil, errMissingAlgoOrderID
	}
	input := &struct {
		AlgoID string `json:"algoId"`
	}{
		AlgoID: algoID,
	}
	var resp []AlgoOrderWithdrawalProfit
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, spotGridWithdrawIncomeEPL, http.MethodPost, gridWithdrawalIncome, input, &resp, true)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// ComputeMarginBalance computes margin balance with 'add' and 'reduce' balance type
func (ok *Okx) ComputeMarginBalance(ctx context.Context, arg MarginBalanceParam) (*ComputeMarginBalance, error) {
	if arg.AlgoID == "" {
		return nil, errInvalidAlgoID
	}
	if !(arg.Type == "add" || arg.Type == "reduce") {
		return nil, errInvalidMarginTypeAdjust
	}
	var resp []ComputeMarginBalance
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, computeMarginBalanceEPL, http.MethodPost, gridComputeMarginBalance, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// AdjustMarginBalance retrives adjust margin balance with 'add' and 'reduce' balance type
func (ok *Okx) AdjustMarginBalance(ctx context.Context, arg MarginBalanceParam) (*AdjustMarginBalanceResponse, error) {
	if arg.AlgoID == "" {
		return nil, errInvalidAlgoID
	}
	if !(arg.Type == "add" || arg.Type == "reduce") {
		return nil, errInvalidMarginTypeAdjust
	}
	if arg.Percentage <= 0 && arg.Amount < 0 {
		return nil, errors.New("either percentage or amount is required")
	}
	var resp []AdjustMarginBalanceResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, adjustMarginBalanceEPL, http.MethodPost, gridMarginBalance, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetGridAIParameter retrives grid AI parameter
func (ok *Okx) GetGridAIParameter(ctx context.Context, algoOrderType, instrumentID, direction, duration string) ([]GridAIParameterResponse, error) {
	params := url.Values{}
	if !(algoOrderType == "moon_grid" || algoOrderType == "contract_grid" || algoOrderType == "grid") {
		return nil, errInvalidAlgoOrderType
	}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	if algoOrderType == "contract_grid" && !(direction == "long" || direction == "short" || direction == "neutral") {
		return nil, errors.New("parameter 'direction' is required for 'contract_grid' algo order type")
	}
	params.Set("direction", direction)
	params.Set("algoOrdType", algoOrderType)
	params.Set("instId", instrumentID)
	if duration == "7D" || duration == "30D" || duration == "180D" {
		params.Set("duration", duration)
	}
	var resp []GridAIParameterResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getGridAIParameterEPL, http.MethodGet, common.EncodeURLValues(gridAIParams, params), nil, &resp, true)
}

// ****************************************** Earn **************************************************

// GetOffers retrives list of offers for different protocols.
func (ok *Okx) GetOffers(ctx context.Context, productID, protocolType, currency string) ([]Offer, error) {
	params := url.Values{}
	if productID != "" {
		params.Set("productId", productID)
	}
	if protocolType != "" {
		params.Set("protocolType", protocolType)
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	var resp []Offer
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOfferEPL, http.MethodGet, common.EncodeURLValues(financeOffers, params), nil, &resp, true)
}

// Purchase invest on specific product
func (ok *Okx) Purchase(ctx context.Context, arg PurchaseRequestParam) (*OrderIDResponse, error) {
	if arg.ProductID == "" {
		return nil, fmt.Errorf("%v, missing product id", errMissingRequiredParameter)
	}
	for x := range arg.InvestData {
		if arg.InvestData[x].Currency == "" {
			return nil, fmt.Errorf("%v, currency information for investment is required", errMissingRequiredParameter)
		}
		if arg.InvestData[x].Amount <= 0 {
			return nil, errUnacceptableAmount
		}
	}
	var resp []OrderIDResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, purchaseEPL, http.MethodPost, financePurchase, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// Redeem redemption of investment
func (ok *Okx) Redeem(ctx context.Context, arg RedeemRequestParam) (*OrderIDResponse, error) {
	if arg.OrderID == "" {
		return nil, fmt.Errorf("%v, missing 'orderId'", errMissingRequiredParameter)
	}
	if !(arg.ProtocolType == "staking" || arg.ProtocolType == "defi") {
		return nil, fmt.Errorf("%v, invalid protocol type, only %s and %s allowed", errMissingRequiredParameter, "staking", "defi")
	}
	var resp []OrderIDResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, redeemEPL, http.MethodPost, financeRedeem, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// CancelPurchaseOrRedemption canceld Purchase or Redemptions
// after cancelling, returning funds will go to the funding account.
func (ok *Okx) CancelPurchaseOrRedemption(ctx context.Context, arg CancelFundingParam) (*OrderIDResponse, error) {
	if arg.OrderID == "" {
		return nil, fmt.Errorf("%v, missing 'orderId'", errMissingRequiredParameter)
	}
	if !(arg.ProtocolType == "staking" || arg.ProtocolType == "defi") {
		return nil, fmt.Errorf("%v, invalid protocol type, only %s and %s allowed", errMissingRequiredParameter, "staking", "defi")
	}
	var resp []OrderIDResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, cancelPurchaseOrRedemptionEPL, http.MethodPost, financeCacelPurchase, &arg, &resp, true); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetEarnActiveOrders retrives active orders.
func (ok *Okx) GetEarnActiveOrders(ctx context.Context, productID, protocolType, currency, state string) ([]ActiveFundingOrder, error) {
	params := url.Values{}
	if productID != "" {
		params.Set("productId", productID)
	}
	if protocolType == "staking" || protocolType == "defi" {
		params.Set("protocolType", protocolType)
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if state == "1" || state == "2" || state == "8" || state == "9" || state == "13" {
		params.Set("state", state)
	}
	var resp []ActiveFundingOrder
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEarnActiveOrdersEPL, http.MethodGet, common.EncodeURLValues(financeActiveOrders, params), nil, &resp, true)
}

// GetFundingOrderHistory retrives funding order history
func (ok *Okx) GetFundingOrderHistory(ctx context.Context, productID, protocolType, currency string, after, before time.Time, limit int) ([]ActiveFundingOrder, error) {
	params := url.Values{}
	if productID != "" {
		params.Set("productId", productID)
	}
	if protocolType == "staking" || protocolType == "defi" {
		params.Set("protocolType", protocolType)
	}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if !before.IsZero() {
		params.Set("before", strconv.FormatInt(before.UnixMilli(), 10))
	}
	if !after.IsZero() && after.Before(before) {
		params.Set("after", strconv.FormatInt(after.UnixMilli(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var resp []ActiveFundingOrder
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEarnActiveOrdersEPL, http.MethodGet, common.EncodeURLValues(financeOrdersHistory, params), nil, &resp, true)
}

// GetTickers retrieves the latest price snopshots best bid/ ask price, and tranding volume in the last 34 hours.
func (ok *Okx) GetTickers(ctx context.Context, instType, uly, instID string) ([]TickerResponse, error) {
	params := url.Values{}
	instType = strings.ToUpper(instType)
	switch {
	case instType == okxInstTypeSpot || instType == okxInstTypeSwap || instType == okxInstTypeFutures || instType == okxInstTypeOption:
		params.Set("instType", instType)
		if (instType == okxInstTypeSwap || instType == okxInstTypeFutures || instType == okxInstTypeOption) && uly != "" {
			params.Set("uly", uly)
		}
	case instID != "":
		params.Set("instId", instID)
	default:
		return nil, errors.New("missing required variable instType (instrument type) or insId( Instrument ID )")
	}
	var response []TickerResponse
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTickersEPL, http.MethodGet, common.EncodeURLValues(marketTickers, params), nil, &response, false)
}

// GetTicker retrieves the latest price snapshot, best bid/ask price, and trading volume in the last 24 hours.
func (ok *Okx) GetTicker(ctx context.Context, instrumentID string) (*TickerResponse, error) {
	params := url.Values{}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	} else {
		return nil, errors.New("missing required variable instType(instruction type) or insId( Instrument ID )")
	}
	var response []TickerResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTickersEPL, http.MethodGet, common.EncodeURLValues(marketTicker, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	if len(response) == 1 {
		return &response[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetIndexTickers Retrieves index tickers.
func (ok *Okx) GetIndexTickers(ctx context.Context, quoteCurrency, instID string) ([]IndexTicker, error) {
	response := []IndexTicker{}
	if instID == "" && quoteCurrency == "" {
		return nil, errors.New("missing required variable! param quoteCcy or instId has to be set")
	}
	params := url.Values{}
	if quoteCurrency != "" {
		params.Set("quoteCcy", quoteCurrency)
	} else if instID != "" {
		params.Set("instId", instID)
	}
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getIndexTickersEPL, http.MethodGet, common.EncodeURLValues(indexTickers, params), nil, response, false)
}

// getInstrumentIDFromPair returns the instrument ID for the corresponding asset pairs and asset type( Instrument Type )
func (ok *Okx) getInstrumentIDFromPair(ctx context.Context, pair currency.Pair, a asset.Item) (string, error) {
	format, err := ok.GetPairFormat(a, false)
	if err != nil {
		return "", err
	}
	if pair.Base.String() == "" || pair.Quote.String() == "" {
		return "", errors.New("incomplete currency pair")
	}
	switch a {
	case asset.PerpetualSwap:
		return pair.Base.String() + format.Delimiter + pair.Quote.String() + format.Delimiter + okxInstTypeSwap, nil
	case asset.Option:
		instruments, err := ok.GetInstruments(ctx, &InstrumentsFetchParams{
			InstrumentType: okxInstTypeOption,
			Underlying:     pair.Base.String() + format.Delimiter + pair.Quote.String(),
		})
		if err != nil {
			return "", err
		}
		for x := range instruments {
			p, err := currency.NewPairFromString(instruments[x].Underlying)
			if err != nil {
				continue
			}
			if p.Equal(pair) {
				return instruments[x].InstrumentID, nil
			}
		}
	case asset.Futures:
		instruments, err := ok.GetInstruments(ctx, &InstrumentsFetchParams{
			InstrumentType: okxInstTypeFutures,
		})
		if err != nil {
			return "", err
		}
		for x := range instruments {
			p, err := currency.NewPairFromString(instruments[x].Underlying)
			if err != nil {
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
		return okxInstTypeSwap
	}
	return assetType.String()
}

// GetUnderlying returns the instrument ID for the corresponding asset pairs and asset type( Instrument Type )
func (ok *Okx) GetUnderlying(pair currency.Pair, a asset.Item) (string, error) {
	format, err := ok.GetPairFormat(a, false)
	if err != nil {
		return "", err
	}
	if pair.Base.String() == "" || pair.Quote.String() == "" {
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
	pair, err := currency.NewPairFromString(instrumentID)
	return pair, err
}

// GetOrderBookDepth returns the recent order asks and bids before specified timestamp.
func (ok *Okx) GetOrderBookDepth(ctx context.Context, instrumentID string, depth uint) (*OrderBookResponse, error) {
	params := url.Values{}
	params.Set("instId", instrumentID)
	if depth > 0 {
		params.Set("sz", strconv.Itoa(int(depth)))
	}
	var resp []OrderBookResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOrderBookEPL, http.MethodGet, common.EncodeURLValues(marketBooks, params), nil, &resp, false)
	if err != nil {
		return nil, err
	} else if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
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

// GetCandlesticks Retrieve the candlestick charts. This endpoint can retrieves the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
func (ok *Okx) GetCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandles)
}

// GetCandlesticksHistory Retrieve history candlestick charts from recent years.
func (ok *Okx) GetCandlesticksHistory(ctx context.Context, instrumentID string, interval kline.Interval, before, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandlesHistory)
}

// GetIndexCandlesticks Retrieve the candlestick charts of the index. This endpoint can retrieves the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
// the respos is a lis of Candlestick data.
func (ok *Okx) GetIndexCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketCandlesIndex)
}

// GetMarkPriceCandlesticks Retrieve the candlestick charts of mark price. This endpoint can retrieves the latest 1,440 data entries. Charts are returned in groups based on the requested bar.
func (ok *Okx) GetMarkPriceCandlesticks(ctx context.Context, instrumentID string, interval kline.Interval, before, after time.Time, limit uint64) ([]CandleStick, error) {
	return ok.GetCandlestickData(ctx, instrumentID, interval, before, after, limit, marketPriceCandles)
}

// GetCandlestickData handles fetching the data for both the default GetCandlesticks, GetCandlesticksHistory, and GetIndexCandlesticks() methods.
func (ok *Okx) GetCandlestickData(ctx context.Context, instrumentID string, interval kline.Interval, before, after time.Time, limit uint64, route string) ([]CandleStick, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	var resp interface{}
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
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getCandlesticksEPL, http.MethodGet, common.EncodeURLValues(marketCandles, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	responseData, okk := (resp).([]interface{})
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
		var err error
		timestamp, err := strconv.Atoi(individualData[0].(string))
		if err != nil {
			return nil, err
		}
		candle.OpenTime = time.UnixMilli(int64(timestamp))
		if candle.OpenPrice, err = convert.FloatFromString(individualData[1]); err != nil {
			return nil, err
		}
		if candle.HighestPrice, err = convert.FloatFromString(individualData[2]); err != nil {
			return nil, err
		}
		if candle.LowestPrice, err = convert.FloatFromString(individualData[3]); err != nil {
			return nil, err
		}
		if candle.ClosePrice, err = convert.FloatFromString(individualData[4]); err != nil {
			return nil, err
		}
		if candle.Volume, err = convert.FloatFromString(individualData[5]); err != nil {
			return nil, err
		}
		if candle.QuoteAssetVolume, err = convert.FloatFromString(individualData[6]); err != nil {
			return nil, err
		}
		klineData[x] = candle
	}
	return klineData, nil
}

// GetTrades Retrieve the recent transactions of an instrument.
func (ok *Okx) GetTrades(ctx context.Context, instrumentID string, limit uint) ([]TradeResponse, error) {
	var resp []TradeResponse
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	if limit > 0 && limit <= 500 {
		params.Set("limit", strconv.Itoa(int(limit)))
	}
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesRequestEPL, http.MethodGet, common.EncodeURLValues(marketTrades, params), nil, &resp, false)
}

// GetTradesHistory retrieves the recent transactions of an instrument from the last 3 months with pagination.
func (ok *Okx) GetTradesHistory(ctx context.Context, instrumentID, before, after string, limit uint) ([]TradeResponse, error) {
	var resp []TradeResponse
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
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getTradesRequestEPL, http.MethodGet, common.EncodeURLValues(marketTradesHistory, params), nil, &resp, false)
}

// Get24HTotalVolume The 24-hour trading volume is calculated on a rolling basis, using USD as the pricing unit.
func (ok *Okx) Get24HTotalVolume(ctx context.Context) (*TradingVolumdIn24HR, error) {
	var resp []TradingVolumdIn24HR
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, get24HTotalVolumeEPL, http.MethodGet, marketPlatformVolumeIn24Hour, nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNo24HrTradeVolumeFound
}

// GetOracle Get the crypto price of signing using Open Oracle smart contract.
func (ok *Okx) GetOracle(ctx context.Context) (*OracleSmartContractResponse, error) {
	var resp []OracleSmartContractResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOracleEPL, http.MethodGet, marketOpenOracles, nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errOracleInformationNotFound
}

// GetExchangeRate this interface provides the average exchange rate data for 2 weeks
// from USD to CNY
func (ok *Okx) GetExchangeRate(ctx context.Context) (*UsdCnyExchangeRate, error) {
	var resp []UsdCnyExchangeRate
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getExchangeRateRequestEPL, http.MethodGet, marketExchangeRate, nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errExchangeInfoNotFound
}

// GetIndexComponents returns the index component information data on the market
func (ok *Okx) GetIndexComponents(ctx context.Context, index string) (*IndexComponent, error) {
	params := url.Values{}
	params.Set("index", index)
	var resp *IndexComponent
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getIndexComponentsEPL, http.MethodGet, common.EncodeURLValues(marketIndexComponents, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errIndexComponentNotFound
	}
	return resp, nil
}

// GetBlockTickers retrieves the latest block trading volume in the last 24 hours.
// Instrument Type Is Mendatory, and Underlying is Optional.
func (ok *Okx) GetBlockTickers(ctx context.Context, instrumentType, underlying string) ([]BlockTicker, error) {
	params := url.Values{}
	if !(instrumentType == okxInstTypeSpot || instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeFutures || instrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instrumentType)
	if underlying != "" {
		params.Set("uly", underlying)
	}
	var resp []BlockTicker
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTickersEPL, http.MethodGet, common.EncodeURLValues(marketBlockTickers, params), nil, &resp, false)
}

// GetBlockTicker retrieves the latest block trading volume in the last 24 hours.
func (ok *Okx) GetBlockTicker(ctx context.Context, instrumentID string) (*BlockTicker, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	var resp []BlockTicker
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTickersEPL, http.MethodGet, common.EncodeURLValues(marketBlockTicker, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetBlockTrades retrieves the recent block trading transactions of an instrument. Descending order by tradeId.
func (ok *Okx) GetBlockTrades(ctx context.Context, instrumentID string) ([]BlockTrade, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	var resp []BlockTrade
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getBlockTradesEPL, http.MethodGet, common.EncodeURLValues(marketBlockTrades, params), nil, &resp, false)
}

/************************************ Public Data Endpoinst *************************************************/

// GetInstruments Retrieve a list of instruments with open contracts.
func (ok *Okx) GetInstruments(ctx context.Context, arg *InstrumentsFetchParams) ([]Instrument, error) {
	params := url.Values{}
	if !(arg.InstrumentType == okxInstTypeSpot || arg.InstrumentType == okxInstTypeMargin || arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeFutures || arg.InstrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", arg.InstrumentType)
	if arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if arg.InstrumentID != "" {
		params.Set("instId", arg.InstrumentID)
	}
	var resp []Instrument
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInstrumentsEPL, http.MethodGet, common.EncodeURLValues(publicInstruments, params), nil, &resp, false)
}

// GetDeliveryHistory retrieves the estimated delivery price of the last 3 months, which will only have a return value one hour before the delivery/exercise.
func (ok *Okx) GetDeliveryHistory(ctx context.Context, instrumentType, underlying string, after, before time.Time, limit int) ([]DeliveryHistory, error) {
	params := url.Values{}
	if instrumentType != "" && !(instrumentType == okxInstTypeFutures || instrumentType == okxInstTypeOption) {
		return nil, fmt.Errorf("unacceptable instrument Type! Only %s and %s are allowed", "FUTURE", okxInstTypeOption)
	} else if instrumentType == "" {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instrumentType)
	if underlying == "" {
		return nil, errMissingRequiredUnderlying
	}
	params.Set("Underlying", underlying)
	params.Set("uly", underlying)
	if !after.IsZero() {
		params.Set("after", strconv.Itoa(int(after.UnixMilli())))
	}
	if !before.IsZero() {
		params.Set("before", strconv.Itoa(int(before.UnixMilli())))
	}
	if !(limit > 0 && limit <= 100) {
		return nil, errLimitValueExceedsMaxOf100
	}
	params.Set("limit", strconv.Itoa(limit))
	var resp []DeliveryHistory
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDeliveryExerciseHistoryEPL, http.MethodGet, common.EncodeURLValues(publicDeliveryExerciseHistory, params), nil, &resp, false)
}

// GetOpenInterest retrieves the total open interest for contracts on OKX
func (ok *Okx) GetOpenInterest(ctx context.Context, instType, uly, instID string) ([]OpenInterest, error) {
	params := url.Values{}
	instType = strings.ToUpper(instType)
	if !(instType == okxInstTypeSpot || instType == okxInstTypeFutures || instType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instType)
	if uly != "" {
		params.Set("uly", uly)
	}
	if instID != "" {
		params.Set("instId", instID)
	}
	var resp []OpenInterest
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOpenInterestEPL, http.MethodGet, common.EncodeURLValues(publicOpenInterestValues, params), nil, &resp, false)
}

// GetFundingRate  Retrieve funding rate.
func (ok *Okx) GetFundingRate(ctx context.Context, instrumentID string) (*FundingRateResponse, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	var resp []FundingRateResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundingEPL, http.MethodGet, common.EncodeURLValues(publicFundingRate, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// GetFundingRateHistory retrieves funding rate history. This endpoint can retrieves data from the last 3 months.
func (ok *Okx) GetFundingRateHistory(ctx context.Context, instrumentID string, before, after time.Time, limit uint) ([]FundingRateResponse, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	if !before.IsZero() {
		params.Set("before", strconv.Itoa(int(before.UnixMilli())))
	}
	if !after.IsZero() {
		params.Set("after", strconv.Itoa(int(after.UnixMilli())))
	}
	if limit > 0 && limit < 100 {
		params.Set("limit", strconv.Itoa(int(limit)))
	} else if limit > 0 {
		return nil, errLimitValueExceedsMaxOf100
	}
	var resp []FundingRateResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getFundingRateHistoryEPL, http.MethodGet, common.EncodeURLValues(publicFundingRateHistory, params), nil, &resp, false)
}

// GetLimitPrice retrieves the highest buy limit and lowest sell limit of the instrument.
func (ok *Okx) GetLimitPrice(ctx context.Context, instrumentID string) (*LimitPriceResponse, error) {
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingInstrumentID
	}
	params.Set("instId", instrumentID)
	var resp []LimitPriceResponse
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLimitPriceEPL, http.MethodGet, common.EncodeURLValues(publicLimitPath, params), nil, &resp, false); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errFundingRateHistoryNotFound
}

// GetOptionMarketData retrieves option market data.
func (ok *Okx) GetOptionMarketData(ctx context.Context, underlying string, expTime time.Time) ([]OptionMarketDataResponse, error) {
	params := url.Values{}
	if underlying == "" {
		return nil, errMissingRequiredUnderlying
	}
	params.Set("uly", underlying)
	if !expTime.IsZero() {
		params.Set("expTime", fmt.Sprintf("%d%d%d", expTime.Year(), expTime.Month(), expTime.Day()))
	}
	var resp []OptionMarketDataResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getOptionMarketDateEPL, http.MethodGet, common.EncodeURLValues(publicOptionalData, params), nil, &resp, false)
}

// GetEstimatedDeliveryPrice retrieves the estimated delivery price which will only have a return value one hour before the delivery/exercise.
func (ok *Okx) GetEstimatedDeliveryPrice(ctx context.Context, instrumentID string) ([]DeliveryEstimatedPrice, error) {
	var resp []DeliveryEstimatedPrice
	params := url.Values{}
	if instrumentID == "" {
		return nil, errMissingRequiredParamInstID
	}
	params.Set("instId", instrumentID)
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEstimatedDeliveryPriceEPL, http.MethodGet, common.EncodeURLValues(publicEstimatedPrice, params), nil, &resp, false)
}

// GetDiscountRateAndInterestFreeQuota retrieves discount rate level and interest-free quota.
func (ok *Okx) GetDiscountRateAndInterestFreeQuota(ctx context.Context, currency string, discountLevel int8) ([]DiscountRate, error) {
	var response []DiscountRate
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	if discountLevel > 0 && discountLevel < 5 {
		params.Set("discountLv", strconv.Itoa(int(discountLevel)))
	}
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getDiscountRateAndInterestFreeQuotaEPL, http.MethodGet, common.EncodeURLValues(publicDiscountRate, params), nil, &response, false)
}

// GetSystemTime Retrieve API server time.
func (ok *Okx) GetSystemTime(ctx context.Context) (*time.Time, error) {
	var resp []ServerTime
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getSystemTimeEPL, http.MethodGet, publicTime, nil, &resp, false); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &(resp[0].Timestamp), nil
	}
	return nil, errNoValidResponseFromServer
}

// GetLiquidationOrders retrieves information on liquidation orders in the last day.
func (ok *Okx) GetLiquidationOrders(ctx context.Context, arg *LiquidationOrderRequestParams) (*LiquidationOrder, error) {
	params := url.Values{}
	if !(arg.InstrumentType == okxInstTypeMargin || arg.InstrumentType == okxInstTypeFutures ||
		arg.InstrumentType == okxInstTypeSwap || arg.InstrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", arg.InstrumentType)
	arg.MarginMode = strings.ToLower(arg.MarginMode)
	if arg.MarginMode == TradeModeIsolated || arg.MarginMode == TradeModeCross {
		params.Set("mgnMode", arg.MarginMode)
	}
	switch {
	case arg.InstrumentType == okxInstTypeMargin && arg.InstrumentID != "":
		params.Set("instId", arg.InstrumentID)
	case arg.InstrumentType == okxInstTypeMargin && arg.Currency.String() != "":
		params.Set("ccy", arg.Currency.String())
	default:
		return nil, errEitherInstIDOrCcyIsRequired
	}
	if (arg.InstrumentType == okxInstTypeFutures || arg.InstrumentType == okxInstTypeSwap || arg.InstrumentType == okxInstTypeOption) && arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	}
	if arg.InstrumentType == okxInstTypeFutures && (arg.Alias == "this_week" || arg.Alias == "next_week" || arg.Alias == "quarter" || arg.Alias == "next_quarter") {
		params.Set("alias", arg.Alias)
	}
	if ((arg.InstrumentType == okxInstTypeFutures || arg.InstrumentType == okxInstTypeSwap) && arg.Alias == "unfilled") || arg.Alias == "filled" {
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
	var response []LiquidationOrder
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLiquidationOrdersEPL, http.MethodGet, common.EncodeURLValues(publicLiquidationOrders, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	if len(response) == 1 {
		return &response[0], nil
	}
	return nil, errLiquidationOrderResponseNotFound
}

// GetMarkPrice  Retrieve mark price.
func (ok *Okx) GetMarkPrice(ctx context.Context, instrumentType, underlying, instrumentID string) ([]MarkPrice, error) {
	params := url.Values{}
	instrumentType = strings.ToUpper(instrumentType)
	if !(instrumentType == okxInstTypeMargin ||
		instrumentType == okxInstTypeFutures ||
		instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instrumentType)
	if underlying != "" {
		params.Set("uly", underlying)
	}
	if instrumentID != "" {
		params.Set("instId", instrumentID)
	}
	var response []MarkPrice
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getMarkPriceEPL, http.MethodGet, common.EncodeURLValues(publicMarkPrice, params), nil, &response, false)
}

// GetPositionTiers retrieves position tiers information，maximum leverage depends on your borrowings and margin ratio.
func (ok *Okx) GetPositionTiers(ctx context.Context, instrumentType, tradeMode, underlying, instrumentID, tiers string) ([]PositionTiers, error) {
	params := url.Values{}
	instrumentType = strings.ToUpper(instrumentType)
	if !(instrumentType == okxInstTypeMargin ||
		instrumentType == okxInstTypeFutures ||
		instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", instrumentType)
	tradeMode = strings.ToLower(tradeMode)
	if !(tradeMode == TradeModeCross || tradeMode == TradeModeIsolated) {
		return nil, errIncorrectRequiredParameterTradeMode
	}
	params.Set("tdMode", tradeMode)
	if instrumentType != okxInstTypeMargin && underlying != "" {
		params.Set("uly", underlying)
	}
	if instrumentType == okxInstTypeMargin && instrumentID != "" {
		params.Set("instId", instrumentID)
	} else if instrumentType == okxInstTypeMargin {
		return nil, errMissingInstrumentID
	}
	if tiers != "" {
		params.Set("tiers", tiers)
	}
	var response []PositionTiers
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getPositionTiersEPL, http.MethodGet, common.EncodeURLValues(publicPositionTiers, params), nil, &response, false)
}

// GetInterestRateAndLoanQuota retrieves an interest rate and loan quota information for various currencies.
func (ok *Okx) GetInterestRateAndLoanQuota(ctx context.Context) (map[string][]InterestRateLoanQuotaItem, error) {
	var response []map[string][]InterestRateLoanQuotaItem
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateAndLoanQuotaEPL, http.MethodGet, publicInterestRateAndLoanQuota, nil, &response, false)
	if err != nil {
		return nil, err
	} else if len(response) == 1 {
		return response[0], nil
	}
	return nil, errInterestRateAndLoanQuotaNotFound
}

// GetInterestRateAndLoanQuotaForVIPLoans retrieves an interest rate and loan quota information for VIP users of various currencies.
func (ok *Okx) GetInterestRateAndLoanQuotaForVIPLoans(ctx context.Context) ([]VIPInterestRateAndLoanQuotaInformation, error) {
	var response []VIPInterestRateAndLoanQuotaInformation
	return response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getInterestRateAndLoanQuoteForVIPLoansEPL, http.MethodGet, publicVIPInterestRateAndLoanQuota, nil, &response, false)
}

// GetPublicUnderlyings returns list of underlyings for various instrument types.
func (ok *Okx) GetPublicUnderlyings(ctx context.Context, instrumentType string) ([]string, error) {
	params := url.Values{}
	instrumentType = strings.ToUpper(instrumentType)
	if !(instrumentType == okxInstTypeFutures ||
		instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", strings.ToUpper(instrumentType))
	var resp [][]string
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getUnderlyingEPL, http.MethodGet, common.EncodeURLValues(publicUnderlyings, params), nil, &resp, false); err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return resp[0], nil
	}
	return nil, errUnderlyingsForSpecifiedInstTypeNofFound
}

// GetInsuranceFundInformations returns insurance fund balance informations.
func (ok *Okx) GetInsuranceFundInformations(ctx context.Context, arg *InsuranceFundInformationRequestParams) (*InsuranceFundInformation, error) {
	if arg == nil {
		return nil, errNilArgument
	}
	params := url.Values{}
	arg.InstrumentType = strings.ToUpper(arg.InstrumentType)
	if !(arg.InstrumentType == okxInstTypeFutures ||
		arg.InstrumentType == okxInstTypeMargin ||
		arg.InstrumentType == okxInstTypeSwap ||
		arg.InstrumentType == okxInstTypeOption) {
		return nil, errMissingRequiredArgInstType
	}
	params.Set("instType", strings.ToUpper(arg.InstrumentType))
	arg.Type = strings.ToLower(arg.Type)
	if arg.Type == "liquidation_balance_deposit" ||
		arg.Type == "bankruptcy_loss" ||
		arg.Type == "platform_revenue" {
		params.Set("type", arg.Type)
	}
	if (arg.InstrumentType != okxInstTypeMargin) && arg.Underlying != "" {
		params.Set("uly", arg.Underlying)
	} else if arg.InstrumentType != okxInstTypeMargin {
		return nil, errMissingRequiredUnderlying
	}
	if (arg.InstrumentType == okxInstTypeMargin) && arg.Currency != "" {
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
	var response []InsuranceFundInformation
	if err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getInsuranceFundEPL, http.MethodGet, common.EncodeURLValues(publicInsuranceFunds, params), nil, &response, false); err != nil {
		return nil, err
	}
	if len(response) == 1 {
		return &response[0], nil
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
	var resp []UnitConvertResponse
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, unitConvertEPL, http.MethodGet, common.EncodeURLValues(publicCurrencyConvertContract, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	if len(resp) == 1 {
		return &resp[0], nil
	}
	return nil, errNoValidResponseFromServer
}

// Trading Data Endpoints

// GetSupportCoins retrieves the currencies supported by the trading data endpoints
func (ok *Okx) GetSupportCoins(ctx context.Context) (*SupportedCoinsData, error) {
	var response SupportedCoinsData
	return &response, ok.SendHTTPRequest(ctx, exchange.RestSpot, getSupportCoinEPL, http.MethodGet, tradingDataSupportedCoins, nil, &response, false)
}

// GetTakerVolume retrieves the taker volume for both buyers and sellers.
func (ok *Okx) GetTakerVolume(ctx context.Context, currency, instrumentType string, begin, end time.Time, period kline.Interval) ([]TakerVolume, error) {
	params := url.Values{}
	instrumentType = strings.ToUpper(instrumentType)
	if !(instrumentType == "CONTRACTS" ||
		instrumentType == okxInstTypeSpot) {
		return nil, errMissingRequiredArgInstType
	} else if instrumentType == okxInstTypeFutures ||
		instrumentType == okxInstTypeMargin ||
		instrumentType == okxInstTypeSwap ||
		instrumentType == okxInstTypeOption {
		return nil, fmt.Errorf("instrument type %s is not allowed for this query", instrumentType)
	}
	params.Set("instType", strings.ToUpper(instrumentType))
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
	var response [][3]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTakerVolumeEPL, http.MethodGet, common.EncodeURLValues(tradingTakerVolume, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	takerVolumes := []TakerVolume{}
	for x := range response {
		timestamp, err := strconv.Atoi(response[x][0])
		if err != nil {
			return nil, err
		}
		sellVolume, err := strconv.ParseFloat(response[x][1], 64)
		if err != nil {
			return nil, err
		}
		buyVolume, err := strconv.ParseFloat(response[x][2], 64)
		if err != nil {
			return nil, err
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

// GetMarginLendingRatio retrieves the ratio of cumulative amount between currency margin quote currency and base currency.
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
	var response [][2]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getMarginLendingRatioEPL, http.MethodGet, common.EncodeURLValues(tradingMarginLoanRatio, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	lendingRatios := []MarginLendRatioItem{}
	for x := range response {
		if len(response[x]) != 2 {
			return nil, errMalformedData
		}
		timestamp, err := strconv.Atoi(response[x][0])
		if err != nil {
			return nil, err
		}
		ratio, err := strconv.ParseFloat(response[x][0], 64)
		if err != nil || ratio <= 0 {
			return nil, err
		}
		lendRatio := MarginLendRatioItem{
			Timestamp:       time.UnixMilli(int64(timestamp)),
			MarginLendRatio: ratio,
		}
		lendingRatios = append(lendingRatios, lendRatio)
	}
	return lendingRatios, nil
}

// GetLongShortRatio retrieves the ratio of users with net long vs net short positions for futures and perpetual swaps.
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
	var response [][2]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getLongShortRatioEPL, http.MethodGet, common.EncodeURLValues(longShortAccountRatio, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	ratios := []LongShortRatio{}
	for x := range response {
		if len(response[x]) != 2 {
			return nil, fmt.Errorf("%v, expecting length 2 but found %d", errMalformedData, len(response[x]))
		}
		timestamp, err := strconv.Atoi(response[x][0])
		if err != nil || timestamp <= 0 {
			return nil, err
		}
		ratio, err := strconv.ParseFloat(response[x][0], 64)
		if err != nil || ratio <= 0 {
			return nil, err
		}
		dratio := LongShortRatio{
			Timestamp:       time.UnixMilli(int64(timestamp)),
			MarginLendRatio: ratio,
		}
		ratios = append(ratios, dratio)
	}
	return ratios, nil
}

// GetContractsOpenInterestAndVolume retrieves the open interest and trading volume for futures and perpetual swaps.
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
	var response [][3]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getContractsOpeninterestAndVolumeEPL, http.MethodGet, common.EncodeURLValues(contractOpenInterestVolume, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	for x := range response {
		if len(response[x]) != 3 {
			return nil, errMalformedData
		}
		timestamp, err := strconv.ParseFloat(response[x][0], 64)
		if err != nil || timestamp <= 0 {
			return nil, err
		}
		openInterest, err := strconv.ParseFloat(response[x][1], 64)
		if err != nil || openInterest <= 0 {
			return nil, err
		}
		volume, err := strconv.ParseFloat(response[x][2], 64)
		if err != nil {
			return nil, err
		}
		openInterestVolume := OpenInterestVolume{
			Timestamp:    time.UnixMilli(int64(timestamp)),
			Volume:       volume,
			OpenInterest: openInterest,
		}
		openInterestVolumes = append(openInterestVolumes, openInterestVolume)
	}
	return openInterestVolumes, nil
}

// GetOptionsOpenInterestAndVolume retrieves the open interest and trading volume for options.
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
	var response [][3]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOptionsOpenInterestAndVolumeEPL, http.MethodGet, common.EncodeURLValues(optionOpenInterestVolume, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	for x := range response {
		if len(response[x]) != 3 {
			return nil, errors.New("invalid data length")
		}
		timestamp, err := strconv.Atoi(response[x][0])
		if err != nil || timestamp <= 0 {
			return nil, errors.New("invalid timestamp information")
		}
		openInterest, err := strconv.ParseFloat(response[x][1], 64)
		if err != nil || openInterest <= 0 {
			return nil, err
		}
		volumen, err := strconv.ParseFloat(response[x][2], 64)
		if err != nil {
			return nil, err
		}
		openInterestVolume := OpenInterestVolume{
			Timestamp:    time.UnixMilli(int64(timestamp)),
			Volume:       volumen,
			OpenInterest: openInterest,
		}
		openInterestVolumes = append(openInterestVolumes, openInterestVolume)
	}
	return openInterestVolumes, nil
}

// GetPutCallRatio retrieves the open interest ration and trading volume ratio of calls vs puts.
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
	var response [][3]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getPutCallRatioEPL, http.MethodGet, common.EncodeURLValues(optionOpenInterestVolumeRatio, params), nil, &response, false)
	if err != nil {
		return nil, err
	}
	for x := range response {
		if len(response[x]) != 3 {
			return nil, fmt.Errorf("%v, expecting row length 3 but found %d", errMalformedData, len(response[x]))
		}
		timestamp, err := strconv.Atoi(response[x][0])
		if err != nil {
			return nil, err
		}
		openInterest, err := strconv.ParseFloat(response[x][1], 64)
		if err != nil {
			return nil, err
		}
		volumen, err := strconv.ParseFloat(response[x][2], 64)
		if err != nil {
			return nil, err
		}
		openInterestVolume := OpenInterestVolumeRatio{
			Timestamp:         time.UnixMilli(int64(timestamp)),
			VolumeRatio:       volumen,
			OpenInterestRatio: openInterest,
		}
		openInterestVolumeRatios = append(openInterestVolumeRatios, openInterestVolume)
	}
	return openInterestVolumeRatios, nil
}

// GetOpenInterestAndVolumeExpiry retrieves the open interest and trading volume of calls and puts for each upcoming expiration.
func (ok *Okx) GetOpenInterestAndVolumeExpiry(ctx context.Context, currency string, period kline.Interval) ([]ExpiryOpenInterestAndVolume, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("ccy", currency)
	}
	interval := ok.GetIntervalEnum(period)
	if interval != "" {
		params.Set("period", interval)
	}
	var resp [][6]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOpenInterestAndVolumeEPL, http.MethodGet, common.EncodeURLValues(optionOpenInterestVolumeExpiry, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	volumes := []ExpiryOpenInterestAndVolume{}
	for x := range resp {
		var timestamp int
		if len(resp[x]) != 6 {
			return nil, errMalformedData
		}
		timestamp, err = strconv.Atoi(resp[x][0])
		if err != nil {
			return nil, err
		}
		var expiryTime time.Time
		expTime := resp[x][1]
		if expTime != "" && len(expTime) == 8 {
			year, err := strconv.Atoi(expTime[0:4])
			if err != nil {
				continue
			}
			month, err := strconv.Atoi(expTime[4:6])
			var months string
			var days string
			if month <= 9 {
				months = fmt.Sprintf("0%d", month)
			} else {
				months = strconv.Itoa(month)
			}
			if err != nil {
				continue
			}
			day, err := strconv.Atoi(expTime[6:])
			if day <= 9 {
				days = fmt.Sprintf("0%d", day)
			} else {
				days = strconv.Itoa(day)
			}
			if err != nil {
				return nil, err
			}
			expiryTime, err = time.Parse("2006-01-02", fmt.Sprintf("%d-%s-%s", year, months, days))
			if err != nil {
				return nil, err
			}
		}
		calloi, err := strconv.ParseFloat(resp[x][2], 64)
		if err != nil {
			return nil, err
		}
		putoi, err := strconv.ParseFloat(resp[x][3], 64)
		if err != nil {
			return nil, err
		}
		callvol, err := strconv.ParseFloat(resp[x][4], 64)
		if err != nil {
			return nil, err
		}
		putvol, err := strconv.ParseFloat(resp[x][5], 64)
		if err != nil {
			return nil, err
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

// GetOpenInterestAndVolumeStrike retrieves the taker volume for both buyers and sellers of calls and puts.
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
			days = strconv.Itoa(expTime.Day())
		}
		params.Set("expTime", fmt.Sprintf("%d%s%s", expTime.Year(), months, days))
	} else {
		return nil, errMissingExpiryTimeParameter
	}
	var resp [][6]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getOpenInterestAndVolumeEPL, http.MethodGet, common.EncodeURLValues(optionOpenInterestVolumeStrike, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	volumes := []StrikeOpenInterestAndVolume{}
	for x := range resp {
		if len(resp[x]) != 6 {
			return nil, fmt.Errorf("%v, expecting row length of 6 but found %d", errMalformedData, len(resp[x]))
		}
		timestamp, err := strconv.Atoi(resp[x][0])
		if err != nil {
			return nil, err
		}
		strike, err := strconv.ParseInt(resp[x][1], 10, 64)
		if err != nil {
			return nil, err
		}
		calloi, err := strconv.ParseFloat(resp[x][2], 64)
		if err != nil {
			return nil, err
		}
		putoi, err := strconv.ParseFloat(resp[x][3], 64)
		if err != nil {
			return nil, err
		}
		callvol, err := strconv.ParseFloat(resp[x][4], 64)
		if err != nil {
			return nil, err
		}
		putvol, err := strconv.ParseFloat(resp[x][5], 64)
		if err != nil {
			return nil, err
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
	var resp [7]string
	err := ok.SendHTTPRequest(ctx, exchange.RestSpot, getTakerFlowEPL, http.MethodGet, common.EncodeURLValues(takerBlockVolume, params), nil, &resp, false)
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(resp[0], 10, 64)
	if err != nil {
		return nil, err
	}
	callbuyvol, err := strconv.ParseFloat(resp[1], 64)
	if err != nil {
		return nil, err
	}
	callselvol, err := strconv.ParseFloat(resp[2], 64)
	if err != nil {
		return nil, err
	}
	putbutvol, err := strconv.ParseFloat(resp[3], 64)
	if err != nil {
		return nil, err
	}
	putsellvol, err := strconv.ParseFloat(resp[4], 64)
	if err != nil {
		return nil, err
	}
	callblockvol, err := strconv.ParseFloat(resp[5], 64)
	if err != nil {
		return nil, err
	}
	putblockvol, err := strconv.ParseFloat(resp[6], 64)
	if err != nil {
		return nil, err
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
	type errCap struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}
	var errMessage errCap
	errMessage.Data = result
	err = json.Unmarshal(intermediary, &errMessage)
	if err != nil {
		return err
	}
	code, err := strconv.ParseInt(errMessage.Code, 10, 64)
	if err == nil && code != 0 {
		if errMessage.Msg != "" {
			return fmt.Errorf(" error code:%d message: %s", code, errMessage.Msg)
		}
		err, okay := ErrorCodes[strconv.FormatInt(code, 10)]
		if okay {
			return err
		}
		return fmt.Errorf("error code: %d", code)
	}
	return nil
}

// Status

// SystemStatusResponse retrieves the system status.
func (ok *Okx) SystemStatusResponse(ctx context.Context, state string) ([]SystemStatusResponse, error) {
	params := url.Values{}
	if state == "scheduled" || state == "ongoing" || state == "pre_open" || state == "completed" || state == "canceled" {
		params.Set("state", state)
	}
	var resp []SystemStatusResponse
	return resp, ok.SendHTTPRequest(ctx, exchange.RestSpot, getEventStatusEPL, http.MethodGet, common.EncodeURLValues(systemStatus, params), nil, &resp, true)
}

// GetAssetTypeFromInstrumentType returns an asset Item instance given and Instrument Type string.
func (ok *Okx) GetAssetTypeFromInstrumentType(instrumentType string) (asset.Item, error) {
	switch strings.ToUpper(instrumentType) {
	case okxInstTypeContract:
		return asset.PerpetualContract, nil
	case okxInstTypeSwap:
		return asset.PerpetualSwap, nil
	case okxInstTypeANY:
		return asset.Empty, nil
	default:
		return asset.New(strings.ToLower(instrumentType))
	}
}

// GuessAssetTypeFromInstrumentID returns or guesses the instrument id.
func (ok *Okx) GuessAssetTypeFromInstrumentID(instrumentID string) asset.Item {
	if strings.HasSuffix(instrumentID, okxInstTypeSwap) {
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
