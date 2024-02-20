package coinbasepro

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/fundingrate"
	"github.com/thrasher-corp/gocryptotrader/exchanges/futures"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
	gctlog "github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

var (
	c          = &CoinbasePro{}
	testCrypto = currency.BTC
	testFiat   = currency.USD
	testPair   = currency.NewPairWithDelimiter(testCrypto.String(), testFiat.String(), "-")
)

// Please supply your APIKeys here for better testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
	testingInSandbox        = false
)

// Constants used within tests
const (
	// Donation address
	testAddress = "bc1qk0jareu4jytc0cfrhr5wgshsq8282awpavfahc"

	skipPayMethodNotFound          = "no payment methods found, skipping"
	skipInsufSuitableAccs          = "insufficient suitable accounts for test, skipping"
	skipInsufficientFunds          = "insufficient funds for test, skipping"
	skipInsufficientOrders         = "insufficient orders for test, skipping"
	skipInsufficientPortfolios     = "insufficient portfolios for test, skipping"
	skipInsufficientWallets        = "insufficient wallets for test, skipping"
	skipInsufficientFundsOrWallets = "insufficient funds or wallets for test, skipping"
	skipInsufficientTransactions   = "insufficient transactions for test, skipping"

	errExpectMismatch            = "received: '%v' but expected: '%v'"
	errExpectedNonEmpty          = "expected non-empty response"
	errOrder0CancelFail          = "order 0 failed to cancel"
	errIDNotSet                  = "ID not set"
	errx7f                       = "setting proxy address error parse \"\\x7f\": net/url: invalid control character in URL"
	errPortfolioNameDuplicate    = `CoinbasePro unsuccessful HTTP status code: 409 raw response: {"error":"CONFLICT","error_details":"[PORTFOLIO_ERROR_CODE_ALREADY_EXISTS] the requested portfolio name already exists","message":"[PORTFOLIO_ERROR_CODE_ALREADY_EXISTS] the requested portfolio name already exists"}, authenticated request failed`
	errPortTransferInsufFunds    = `CoinbasePro unsuccessful HTTP status code: 429 raw response: {"error":"unknown","error_details":"[PORTFOLIO_ERROR_CODE_INSUFFICIENT_FUNDS] insufficient funds in source account","message":"[PORTFOLIO_ERROR_CODE_INSUFFICIENT_FUNDS] insufficient funds in source account"}, authenticated request failed`
	errInvalidProductID          = `CoinbasePro unsuccessful HTTP status code: 400 raw response: {"error":"INVALID_ARGUMENT","error_details":"valid product_id is required","message":"valid product_id is required"}, authenticated request failed`
	errFeeBuilderNil             = "*exchange.FeeBuilder nil pointer"
	errUnsupportedAssetType      = " unsupported asset type"
	errUpsideUnsupported         = "unsupported asset type upsideprofitcontract"
	errBlorboGranularity         = "invalid granularity blorbo, allowed granularities are: [ONE_MINUTE FIVE_MINUTE FIFTEEN_MINUTE THIRTY_MINUTE ONE_HOUR TWO_HOUR SIX_HOUR ONE_DAY]"
	errNoEndpointPathEdgeCase3   = "no endpoint path found for the given key: EdgeCase3URL"
	errJSONUnsupportedChan       = "json: unsupported type: chan struct {}, authenticated request failed"
	errExpectedFeeRange          = "expected fee range of %v and %v, received %v"
	errJSONNumberIntoString      = "json: cannot unmarshal number into Go value of type string"
	errParseIntValueOutOfRange   = `strconv.ParseInt: parsing "922337203685477580700": value out of range`
	errParseUintInvalidSyntax    = `strconv.ParseUint: parsing "l": invalid syntax`
	errJSONInvalidCharacter      = `invalid character ':' after array element`
	errL2DataMoo                 = "unknown l2update data type moo"
	errUnrecognisedOrderType     = `'' unrecognised order type`
	errOrderSideInvalid          = `'' order side is invalid`
	errUnrecognisedStatusType    = " not recognised as status type"
	errFakeSide                  = "unknown side fakeside"
	errCoinbaseWSAlreadyDisabled = "websocket already disabled for exchange 'CoinbasePro'"

	expectedTimestamp = "1970-01-01 00:20:34 +0000 UTC"

	testAmount = 1e-08
	testPrice  = 1e+09
)

func TestSetup(t *testing.T) {
	cfg, err := c.GetStandardConfig()
	assert.NoError(t, err)
	cfg.API.AuthenticatedSupport = true
	cfg.API.Credentials.Key = apiKey
	cfg.API.Credentials.Secret = apiSecret
	cfg.Enabled = false
	cfg.Enabled = true
	cfg.ProxyAddress = string(rune(0x7f))
	err = c.Setup(cfg)
	if err.Error() != errx7f {
		t.Errorf(errExpectMismatch, err, errx7f)
	}
}

func TestMain(m *testing.M) {
	c.SetDefaults()
	if testingInSandbox {
		err := c.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
			exchange.RestSpot: coinbaseproSandboxAPIURL,
		})
		log.Fatal("failed to set sandbox endpoint", err)
	}
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal("load config error", err)
	}
	gdxConfig, err := cfg.GetExchangeConfig("CoinbasePro")
	if err != nil {
		log.Fatal("init error")
	}
	if apiKey != "" {
		gdxConfig.API.Credentials.Key = apiKey
		gdxConfig.API.Credentials.Secret = apiSecret
		gdxConfig.API.AuthenticatedSupport = true
		gdxConfig.API.AuthenticatedWebsocketSupport = true
	}
	c.Websocket = sharedtestvalues.NewTestWebsocket()
	err = c.Setup(gdxConfig)
	if err != nil {
		log.Fatal("CoinbasePro setup error", err)
	}
	if apiKey != "" {
		c.GetBase().API.AuthenticatedSupport = true
		c.GetBase().API.AuthenticatedWebsocketSupport = true
	}
	err = gctlog.SetGlobalLogConfig(gctlog.GenDefaultSettings())
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func TestGetAllAccounts(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetAllAccounts(context.Background(), 50, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAccountByID(t *testing.T) {
	_, err := c.GetAccountByID(context.Background(), "")
	assert.ErrorIs(t, err, errAccountIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	longResp, err := c.GetAllAccounts(context.Background(), 49, "")
	assert.NoError(t, err)
	if len(longResp.Accounts) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	shortResp, err := c.GetAccountByID(context.Background(), longResp.Accounts[0].UUID)
	assert.NoError(t, err)
	if *shortResp != longResp.Accounts[0] {
		t.Errorf(errExpectMismatch, shortResp, longResp.Accounts[0])
	}
}

func TestGetBestBidAsk(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testPairs := []string{testPair.String(), "ETH-USD"}
	resp, err := c.GetBestBidAsk(context.Background(), testPairs)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetProductBook(t *testing.T) {
	_, err := c.GetProductBook(context.Background(), "", 0)
	assert.ErrorIs(t, err, errProductIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetProductBook(context.Background(), testPair.String(), 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllProducts(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testPairs := []string{testPair.String(), "ETH-USD"}
	resp, err := c.GetAllProducts(context.Background(), 30000, 0, "SPOT", "PERPETUAL", "STATUS_ALL",
		testPairs)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetProductByID(t *testing.T) {
	_, err := c.GetProductByID(context.Background(), "")
	assert.ErrorIs(t, err, errProductIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetProductByID(context.Background(), testPair.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetHistoricRates(t *testing.T) {
	_, err := c.GetHistoricRates(context.Background(), "", granUnknown, time.Time{}, time.Time{})
	assert.ErrorIs(t, err, errProductIDEmpty)
	_, err = c.GetHistoricRates(context.Background(), testPair.String(), "blorbo", time.Time{}, time.Time{})
	if err.Error() != errBlorboGranularity {
		t.Errorf(errExpectMismatch, err, errBlorboGranularity)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetHistoricRates(context.Background(), testPair.String(), granOneMin,
		time.Now().Add(-5*time.Minute), time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetTicker(t *testing.T) {
	_, err := c.GetTicker(context.Background(), "", 1, time.Time{}, time.Time{})
	assert.ErrorIs(t, err, errProductIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetTicker(context.Background(), testPair.String(), 5, time.Time{}, time.Time{})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestPlaceOrder(t *testing.T) {
	_, err := c.PlaceOrder(context.Background(), "", "", "", "", "", "", "", "", 0, 0, 0, 0, false, time.Time{})
	assert.ErrorIs(t, err, errClientOrderIDEmpty)
	_, err = c.PlaceOrder(context.Background(), "meow", "", "", "", "", "", "", "", 0, 0, 0, 0, false, time.Time{})
	assert.ErrorIs(t, err, errProductIDEmpty)
	_, err = c.PlaceOrder(context.Background(), "meow", testPair.String(), order.Sell.String(), "", "", "", "", "", 0,
		0, 0, 0, false, time.Time{})
	assert.ErrorIs(t, err, errAmountEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	id, err := uuid.NewV4()
	assert.NoError(t, err)
	resp, err := c.PlaceOrder(context.Background(), id.String(), testPair.String(), order.Sell.String(), "",
		order.Limit.String(), "", "CROSS", "", testAmount, testPrice, 0, 9999, false, time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	id, err = uuid.NewV4()
	assert.NoError(t, err)
	resp, err = c.PlaceOrder(context.Background(), id.String(), testPair.String(), order.Sell.String(), "",
		order.Limit.String(), "", "MULTI", "", testAmount, testPrice, 0, 9999, false, time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCancelOrders(t *testing.T) {
	var OrderSlice []string
	_, err := c.CancelOrders(context.Background(), OrderSlice)
	assert.ErrorIs(t, err, errOrderIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	ordID, err := c.PlaceOrder(context.Background(), "meow", testPair.String(), order.Sell.String(), "",
		order.Limit.String(), "", "", "", testPrice, testAmount, 0, 9999, false, time.Time{})
	assert.NoError(t, err)
	OrderSlice = append(OrderSlice, ordID.OrderID)
	resp, err := c.CancelOrders(context.Background(), OrderSlice)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestEditOrder(t *testing.T) {
	_, err := c.EditOrder(context.Background(), "", 0, 0)
	assert.ErrorIs(t, err, errOrderIDEmpty)
	_, err = c.EditOrder(context.Background(), "meow", 0, 0)
	assert.ErrorIs(t, err, errSizeAndPriceZero)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	id, err := uuid.NewV4()
	assert.NoError(t, err)
	ordID, err := c.PlaceOrder(context.Background(), id.String(), testPair.String(), order.Sell.String(), "",
		order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999, false, time.Time{})
	assert.NoError(t, err)
	resp, err := c.EditOrder(context.Background(), ordID.OrderID, testAmount, testPrice*10)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestEditOrderPreview(t *testing.T) {
	_, err := c.EditOrderPreview(context.Background(), "", 0, 0)
	assert.ErrorIs(t, err, errOrderIDEmpty)
	_, err = c.EditOrderPreview(context.Background(), "meow", 0, 0)
	assert.ErrorIs(t, err, errSizeAndPriceZero)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	skipTestIfLowOnFunds(t)
	id, err := uuid.NewV4()
	assert.NoError(t, err)
	ordID, err := c.PlaceOrder(context.Background(), id.String(), testPair.String(), order.Sell.String(), "",
		order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999, false, time.Time{})
	assert.NoError(t, err)
	resp, err := c.EditOrderPreview(context.Background(), ordID.OrderID, testAmount, testPrice*10)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllOrders(t *testing.T) {
	assets := []string{"USD"}
	status := make([]string, 2)
	_, err := c.GetAllOrders(context.Background(), "", "", "", "", "", "", "", "", "", status, assets, 0,
		time.Unix(2, 2), time.Unix(1, 1))
	assert.ErrorIs(t, err, common.ErrStartAfterEnd)
	status[0] = "CANCELLED"
	status[1] = "OPEN"
	_, err = c.GetAllOrders(context.Background(), "", "", "", "", "", "", "", "", "", status, assets, 0, time.Time{},
		time.Time{})
	assert.ErrorIs(t, err, errOpenPairWithOtherTypes)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	status = make([]string, 0)
	assets = make([]string, 1)
	assets[0] = testCrypto.String()
	_, err = c.GetAllOrders(context.Background(), "", "USD", "LIMIT", "SELL", "", "SPOT", "RETAIL_ADVANCED",
		"UNKNOWN_CONTRACT_EXPIRY_TYPE", "2", status, assets, 10, time.Time{}, time.Time{})
	assert.NoError(t, err)
}

func TestGetFills(t *testing.T) {
	_, err := c.GetFills(context.Background(), "", "", "", time.Unix(2, 2), time.Unix(1, 1), 0)
	assert.ErrorIs(t, err, common.ErrStartAfterEnd)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err = c.GetFills(context.Background(), "", testPair.String(), "", time.Unix(1, 1), time.Now(), 5)
	assert.NoError(t, err)
	status := []string{"OPEN"}
	ordID, err := c.GetAllOrders(context.Background(), "", "", "", "", "", "", "", "", "", status, nil, 3, time.Time{},
		time.Time{})
	assert.NoError(t, err)
	if len(ordID.Orders) == 0 {
		t.Skip(skipInsufficientOrders)
	}
	_, err = c.GetFills(context.Background(), ordID.Orders[0].OrderID, "", "", time.Time{}, time.Time{}, 5)
	assert.NoError(t, err)
}

func TestGetOrderByID(t *testing.T) {
	_, err := c.GetOrderByID(context.Background(), "", "", "")
	assert.ErrorIs(t, err, errOrderIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	ordID, err := c.GetAllOrders(context.Background(), "", "", "", "", "", "", "", "", "", nil, nil, 10,
		time.Time{}, time.Time{})
	assert.NoError(t, err)
	if len(ordID.Orders) == 0 {
		t.Skip(skipInsufficientOrders)
	}
	resp, err := c.GetOrderByID(context.Background(), ordID.Orders[0].OrderID, ordID.Orders[0].ClientOID, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestPreviewOrder(t *testing.T) {
	_, err := c.PreviewOrder(context.Background(), "", "", "", "", "", 0, 0, 0, 0, 0, 0, false, false, false, time.Time{})
	assert.ErrorIs(t, err, errAmountEmpty)
	_, err = c.PreviewOrder(context.Background(), "", "", "", "", "", 0, 1, 0, 0, 0, 0, false, false, false, time.Time{})
	assert.ErrorIs(t, err, errInvalidOrderType)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	skipTestIfLowOnFunds(t)
	resp, err := c.PreviewOrder(context.Background(), testPair.String(), "BUY", "MARKET", "", "", 0, 1, 0, 0, 0, 0, false,
		false, false, time.Time{})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllPortfolios(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetAllPortfolios(context.Background(), "DEFAULT")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCreatePortfolio(t *testing.T) {
	_, err := c.CreatePortfolio(context.Background(), "")
	assert.ErrorIs(t, err, errNameEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	_, err = c.CreatePortfolio(context.Background(), "GCT Test Portfolio")
	if err != nil && err.Error() != errPortfolioNameDuplicate {
		t.Error(err)
	}
}

func TestMovePortfolioFunds(t *testing.T) {
	_, err := c.MovePortfolioFunds(context.Background(), "", "", "", 0)
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	_, err = c.MovePortfolioFunds(context.Background(), "", "meowPort", "woofPort", 0)
	assert.ErrorIs(t, err, errCurrencyEmpty)
	_, err = c.MovePortfolioFunds(context.Background(), testCrypto.String(), "meowPort", "woofPort", 0)
	assert.ErrorIs(t, err, errAmountEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	portID, err := c.GetAllPortfolios(context.Background(), "")
	assert.NoError(t, err)
	if len(portID.Portfolios) < 2 {
		t.Skip(skipInsufficientPortfolios)
	}
	_, err = c.MovePortfolioFunds(context.Background(), testCrypto.String(), portID.Portfolios[0].UUID, portID.Portfolios[1].UUID,
		testAmount)
	if err != nil && err.Error() != errPortTransferInsufFunds {
		t.Error(err)
	}
}

func TestGetPortfolioByID(t *testing.T) {
	_, err := c.GetPortfolioByID(context.Background(), "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	portID, err := c.GetAllPortfolios(context.Background(), "")
	assert.NoError(t, err)
	if len(portID.Portfolios) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	resp, err := c.GetPortfolioByID(context.Background(), portID.Portfolios[0].UUID)
	assert.NoError(t, err)
	if resp.Breakdown.Portfolio != portID.Portfolios[0] {
		t.Errorf(errExpectMismatch, resp.Breakdown.Portfolio, portID.Portfolios[0])
	}
}

func TestDeletePortfolio(t *testing.T) {
	err := c.DeletePortfolio(context.Background(), "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	pID := portfolioIDFromName(t, "GCT Test Portfolio To-Delete")
	err = c.DeletePortfolio(context.Background(), pID)
	assert.NoError(t, err)
}

func TestEditPortfolio(t *testing.T) {
	_, err := c.EditPortfolio(context.Background(), "", "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	_, err = c.EditPortfolio(context.Background(), "meow", "")
	assert.ErrorIs(t, err, errNameEmpty)
	pID := portfolioIDFromName(t, "GCT Test Portfolio To-Edit")
	_, err = c.EditPortfolio(context.Background(), pID, "GCT Test Portfolio Edited")
	if err != nil && err.Error() != errPortfolioNameDuplicate {
		t.Error(err)
	}
}

func TestGetFuturesBalanceSummary(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.GetFuturesBalanceSummary(context.Background())
	assert.NoError(t, err)
}

func TestGetAllFuturesPositions(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testGetNoArgs(t, c.GetAllFuturesPositions)
}

func TestGetFuturesPositionByID(t *testing.T) {
	_, err := c.GetFuturesPositionByID(context.Background(), "")
	assert.ErrorIs(t, err, errProductIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err = c.GetFuturesPositionByID(context.Background(), "meow")
	assert.NoError(t, err)
}

func TestScheduleFuturesSweep(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	curSweeps, err := c.ListFuturesSweeps(context.Background())
	assert.NoError(t, err)
	preCancel := false
	if len(curSweeps.Sweeps) > 0 {
		for i := range curSweeps.Sweeps {
			if curSweeps.Sweeps[i].Status == "PENDING" {
				preCancel = true

			}
		}
	}
	if preCancel {
		_, err = c.CancelPendingFuturesSweep(context.Background())
		if err != nil {
			t.Error(err)
		}
	}
	_, err = c.ScheduleFuturesSweep(context.Background(), 0.001337)
	assert.NoError(t, err)
}

func TestListFuturesSweeps(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.ListFuturesSweeps(context.Background())
	assert.NoError(t, err)
}

func TestCancelPendingFuturesSweep(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	curSweeps, err := c.ListFuturesSweeps(context.Background())
	assert.NoError(t, err)
	partialSkip := false
	if len(curSweeps.Sweeps) > 0 {
		for i := range curSweeps.Sweeps {
			if curSweeps.Sweeps[i].Status == "PENDING" {
				partialSkip = true
			}
		}
	}
	if !partialSkip {
		_, err = c.ScheduleFuturesSweep(context.Background(), 0.001337)
		if err != nil {
			t.Error(err)
		}

	}
	_, err = c.CancelPendingFuturesSweep(context.Background())
	assert.NoError(t, err)
}

func TestAllocatePortfolio(t *testing.T) {
	err := c.AllocatePortfolio(context.Background(), "", "", "", 0)
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	err = c.AllocatePortfolio(context.Background(), "meow", "", "", 0)
	assert.ErrorIs(t, err, errProductIDEmpty)
	err = c.AllocatePortfolio(context.Background(), "meow", "bark", "", 0)
	assert.ErrorIs(t, err, errCurrencyEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	pID := getINTXPortfolio(t)
	err = c.AllocatePortfolio(context.Background(), pID, testCrypto.String(), "USD", 0.001337)
	assert.NoError(t, err)
}

func TestGetPerpetualsPortfolioSummary(t *testing.T) {
	_, err := c.GetPerpetualsPortfolioSummary(context.Background(), "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	pID := getINTXPortfolio(t)
	resp, err := c.GetPerpetualsPortfolioSummary(context.Background(), pID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllPerpetualsPositions(t *testing.T) {
	_, err := c.GetAllPerpetualsPositions(context.Background(), "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	pID := getINTXPortfolio(t)
	_, err = c.GetAllPerpetualsPositions(context.Background(), pID)
	assert.NoError(t, err)
}

func TestGetPerpetualsPositionByID(t *testing.T) {
	_, err := c.GetPerpetualsPositionByID(context.Background(), "", "")
	assert.ErrorIs(t, err, errPortfolioIDEmpty)
	_, err = c.GetPerpetualsPositionByID(context.Background(), "meow", "")
	assert.ErrorIs(t, err, errProductIDEmpty)
	pID := getINTXPortfolio(t)
	_, err = c.GetPerpetualsPositionByID(context.Background(), pID, testPair.String())
	assert.NoError(t, err)
}

func TestGetTransactionSummary(t *testing.T) {
	_, err := c.GetTransactionSummary(context.Background(), time.Unix(2, 2), time.Unix(1, 1), "", "", "")
	assert.ErrorIs(t, err, common.ErrStartAfterEnd)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetTransactionSummary(context.Background(), time.Unix(1, 1), time.Now(), "", asset.Spot.Upper(),
		"UNKNOWN_CONTRACT_EXPIRY_TYPE")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCreateConvertQuote(t *testing.T) {
	_, err := c.CreateConvertQuote(context.Background(), "", "", "", "", 0)
	assert.ErrorIs(t, err, errAccountIDEmpty)
	_, err = c.CreateConvertQuote(context.Background(), "meow", "123", "", "", 0)
	assert.ErrorIs(t, err, errAmountEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	fromAccID, toAccID := convertTestHelper(t)
	resp, err := c.CreateConvertQuote(context.Background(), fromAccID, toAccID, "", "", 0.01)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCommitConvertTrade(t *testing.T) {
	convertTestShared(t, c.CommitConvertTrade)
}

func TestGetConvertTradeByID(t *testing.T) {
	convertTestShared(t, c.GetConvertTradeByID)
}

func TestGetV3Time(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testGetNoArgs(t, c.GetV3Time)
}

func TestListNotifications(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.ListNotifications(context.Background(), PaginationInp{})
	assert.NoError(t, err)
}

func TestGetUserByID(t *testing.T) {
	_, err := c.GetUserByID(context.Background(), "")
	assert.ErrorIs(t, err, errUserIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetCurrentUser(context.Background())
	assert.NoError(t, err)
	if resp == nil {
		t.Fatal(errExpectedNonEmpty)
	}
	resp, err = c.GetUserByID(context.Background(), resp.Data.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetCurrentUser(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testGetNoArgs(t, c.GetCurrentUser)
}

func TestGetAuthInfo(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	testGetNoArgs(t, c.GetAuthInfo)
}

func TestGetAllWallets(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	pagIn := PaginationInp{Limit: 2}
	resp, err := c.GetAllWallets(context.Background(), pagIn)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	if resp.Pagination.NextStartingAfter == "" {
		t.Skip(skipInsufficientWallets)
	}
	pagIn.StartingAfter = resp.Pagination.NextStartingAfter
	resp, err = c.GetAllWallets(context.Background(), pagIn)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetWalletByID(t *testing.T) {
	_, err := c.GetWalletByID(context.Background(), "", "")
	assert.ErrorIs(t, err, errCurrWalletConflict)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.GetWalletByID(context.Background(), resp.Data.ID, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCreateAddress(t *testing.T) {
	_, err := c.CreateAddress(context.Background(), "", "")
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	resp, err := c.CreateAddress(context.Background(), wID.Data.ID, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllAddresses(t *testing.T) {
	var pag PaginationInp
	_, err := c.GetAllAddresses(context.Background(), "", pag)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	resp, err := c.GetAllAddresses(context.Background(), wID.Data.ID, pag)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAddressByID(t *testing.T) {
	_, err := c.GetAddressByID(context.Background(), "", "")
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.GetAddressByID(context.Background(), "123", "")
	assert.ErrorIs(t, err, errAddressIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	addID, err := c.GetAllAddresses(context.Background(), wID.Data.ID, PaginationInp{})
	assert.NoError(t, err)
	assert.NotEmpty(t, addID, errExpectedNonEmpty)
	resp, err := c.GetAddressByID(context.Background(), wID.Data.ID, addID.Data[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAddressTransactions(t *testing.T) {
	_, err := c.GetAddressTransactions(context.Background(), "", "", PaginationInp{})
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.GetAddressTransactions(context.Background(), "123", "", PaginationInp{})
	assert.ErrorIs(t, err, errAddressIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	addID, err := c.GetAllAddresses(context.Background(), wID.Data.ID, PaginationInp{})
	assert.NoError(t, err)
	assert.NotEmpty(t, addID, errExpectedNonEmpty)
	_, err = c.GetAddressTransactions(context.Background(), wID.Data.ID, addID.Data[0].ID, PaginationInp{})
	assert.NoError(t, err)
}

func TestSendMoney(t *testing.T) {
	_, err := c.SendMoney(context.Background(), "", "", "", "", "", "", "", "", 0, false, false)
	assert.ErrorIs(t, err, errTransactionTypeEmpty)
	_, err = c.SendMoney(context.Background(), "123", "", "", "", "", "", "", "", 0, false, false)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.SendMoney(context.Background(), "123", "123", "", "", "", "", "", "", 0, false, false)
	assert.ErrorIs(t, err, errToEmpty)
	_, err = c.SendMoney(context.Background(), "123", "123", "123", "", "", "", "", "", 0, false, false)
	assert.ErrorIs(t, err, errAmountEmpty)
	_, err = c.SendMoney(context.Background(), "123", "123", "123", "", "", "", "", "", 1, false, false)
	assert.ErrorIs(t, err, errCurrencyEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wID, err := c.GetAllWallets(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	if len(wID.Data) < 2 {
		t.Skip(skipInsufficientWallets)
	}
	var (
		fromID string
		toID   string
	)
	for i := range wID.Data {
		if wID.Data[i].Currency.Name == testCrypto.String() {
			if wID.Data[i].Balance.Amount > testAmount*100 {
				fromID = wID.Data[i].ID
			} else {
				toID = wID.Data[i].ID
			}
		}
		if fromID != "" && toID != "" {
			break
		}
	}
	if fromID == "" || toID == "" {
		t.Skip(skipInsufficientFundsOrWallets)
	}
	resp, err := c.SendMoney(context.Background(), "transfer", wID.Data[0].ID, wID.Data[1].ID,
		testCrypto.String(), "GCT Test", "123", "", "", testAmount, false, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllTransactions(t *testing.T) {
	var pag PaginationInp
	_, err := c.GetAllTransactions(context.Background(), "", pag)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	_, err = c.GetAllTransactions(context.Background(), wID.Data.ID, pag)
	assert.NoError(t, err)
}

func TestGetTransactionByID(t *testing.T) {
	_, err := c.GetTransactionByID(context.Background(), "", "")
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.GetTransactionByID(context.Background(), "123", "")
	assert.ErrorIs(t, err, errTransactionIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", testCrypto.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	tID, err := c.GetAllTransactions(context.Background(), wID.Data.ID, PaginationInp{})
	assert.NoError(t, err)
	if len(tID.Data) == 0 {
		t.Skip(skipInsufficientTransactions)
	}
	resp, err := c.GetTransactionByID(context.Background(), wID.Data.ID, tID.Data[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestFiatTransfer(t *testing.T) {
	_, err := c.FiatTransfer(context.Background(), "", "", "", 0, false, FiatDeposit)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.FiatTransfer(context.Background(), "123", "", "", 0, false, FiatDeposit)
	assert.ErrorIs(t, err, errAmountEmpty)
	_, err = c.FiatTransfer(context.Background(), "123", "", "", 1, false, FiatDeposit)
	assert.ErrorIs(t, err, errCurrencyEmpty)
	_, err = c.FiatTransfer(context.Background(), "123", "123", "", 1, false, FiatDeposit)
	assert.ErrorIs(t, err, errPaymentMethodEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wallets, err := c.GetAllWallets(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	assert.NotEmpty(t, wallets, errExpectedNonEmpty)
	wID, pmID := transferTestHelper(t, wallets)
	resp, err := c.FiatTransfer(context.Background(), wID, testFiat.String(), pmID, testAmount, false, FiatDeposit)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.FiatTransfer(context.Background(), wID, testFiat.String(), pmID, testAmount, false, FiatWithdrawal)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestCommitTransfer(t *testing.T) {
	_, err := c.CommitTransfer(context.Background(), "", "", FiatDeposit)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.CommitTransfer(context.Background(), "123", "", FiatDeposit)
	assert.ErrorIs(t, err, errDepositIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wallets, err := c.GetAllWallets(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	assert.NotEmpty(t, wallets, errExpectedNonEmpty)
	wID, pmID := transferTestHelper(t, wallets)
	depID, err := c.FiatTransfer(context.Background(), wID, testFiat.String(), pmID, testAmount,
		false, FiatDeposit)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.CommitTransfer(context.Background(), wID, depID.Data.ID, FiatDeposit)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	depID, err = c.FiatTransfer(context.Background(), wID, testFiat.String(), pmID, testAmount,
		false, FiatWithdrawal)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = c.CommitTransfer(context.Background(), wID, depID.Data.ID, FiatWithdrawal)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllFiatTransfers(t *testing.T) {
	var pag PaginationInp
	_, err := c.GetAllFiatTransfers(context.Background(), "", pag, FiatDeposit)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", "AUD")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	_, err = c.GetAllFiatTransfers(context.Background(), wID.Data.ID, pag, FiatDeposit)
	assert.NoError(t, err)
	_, err = c.GetAllFiatTransfers(context.Background(), wID.Data.ID, pag, FiatWithdrawal)
	assert.NoError(t, err)
}

func TestGetFiatTransferByID(t *testing.T) {
	_, err := c.GetFiatTransferByID(context.Background(), "", "", FiatDeposit)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	_, err = c.GetFiatTransferByID(context.Background(), "123", "", FiatDeposit)
	assert.ErrorIs(t, err, errDepositIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	wID, err := c.GetWalletByID(context.Background(), "", "AUD")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, wID, errExpectedNonEmpty)
	dID, err := c.GetAllFiatTransfers(context.Background(), wID.Data.ID, PaginationInp{}, FiatDeposit)
	assert.NoError(t, err)
	if len(dID.Data) == 0 {
		t.Skip(skipInsufficientTransactions)
	}
	resp, err := c.GetFiatTransferByID(context.Background(), wID.Data.ID, dID.Data[0].ID, FiatDeposit)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.GetFiatTransferByID(context.Background(), wID.Data.ID, dID.Data[0].ID, FiatWithdrawal)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAllPaymentMethods(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetAllPaymentMethods(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetPaymentMethodByID(t *testing.T) {
	_, err := c.GetPaymentMethodByID(context.Background(), "")
	assert.ErrorIs(t, err, errPaymentMethodEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	pmID, err := c.GetAllPaymentMethods(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	if len(pmID.Data) == 0 {
		t.Skip(skipPayMethodNotFound)
	}
	resp, err := c.GetPaymentMethodByID(context.Background(), pmID.Data[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetFiatCurrencies(t *testing.T) {
	testGetNoArgs(t, c.GetFiatCurrencies)
}

func TestGetCryptocurrencies(t *testing.T) {
	testGetNoArgs(t, c.GetCryptocurrencies)
}

func TestGetExchangeRates(t *testing.T) {
	resp, err := c.GetExchangeRates(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetPrice(t *testing.T) {
	_, err := c.GetPrice(context.Background(), "", "")
	assert.ErrorIs(t, err, errInvalidPriceType)
	resp, err := c.GetPrice(context.Background(), testPair.String(), asset.Spot.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.GetPrice(context.Background(), testPair.String(), "buy")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.GetPrice(context.Background(), testPair.String(), "sell")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetV2Time(t *testing.T) {
	testGetNoArgs(t, c.GetV2Time)
}

func TestSendHTTPRequest(t *testing.T) {
	err := c.SendHTTPRequest(context.Background(), exchange.EdgeCase3, "", nil)
	if err.Error() != errNoEndpointPathEdgeCase3 {
		t.Errorf(errExpectMismatch, err, errNoEndpointPathEdgeCase3)
	}
}

func TestSendAuthenticatedHTTPRequest(t *testing.T) {
	fc := &CoinbasePro{}
	err := fc.SendAuthenticatedHTTPRequest(context.Background(), exchange.EdgeCase3, "", "", "", nil, false, nil, nil)
	assert.ErrorIs(t, err, exchange.ErrCredentialsAreEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err = c.SendAuthenticatedHTTPRequest(context.Background(), exchange.EdgeCase3, "", "", "", nil, false, nil, nil)
	if err.Error() != errNoEndpointPathEdgeCase3 {
		t.Errorf(errExpectMismatch, err, errNoEndpointPathEdgeCase3)
	}
	ch := make(chan struct{})
	body := map[string]interface{}{"Unmarshalable": ch}
	err = c.SendAuthenticatedHTTPRequest(context.Background(), exchange.RestSpot, "", "", "", body, false, nil, nil)
	if err.Error() != errJSONUnsupportedChan {
		t.Errorf(errExpectMismatch, err, errJSONUnsupportedChan)
	}
}

func TestGetFee(t *testing.T) {
	_, err := c.GetFee(context.Background(), nil)
	if err.Error() != errFeeBuilderNil {
		t.Errorf(errExpectMismatch, errFeeBuilderNil, err)
	}
	feeBuilder := exchange.FeeBuilder{
		FeeType:       exchange.OfflineTradeFee,
		Amount:        1,
		PurchasePrice: 1,
	}
	resp, err := c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if resp != WorstCaseTakerFee {
		t.Errorf(errExpectMismatch, resp, WorstCaseTakerFee)
	}
	feeBuilder.IsMaker = true
	resp, err = c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if resp != WorstCaseMakerFee {
		t.Errorf(errExpectMismatch, resp, WorstCaseMakerFee)
	}
	feeBuilder.Pair = currency.NewPair(currency.USDT, currency.USD)
	resp, err = c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if resp != 0 {
		t.Errorf(errExpectMismatch, resp, StablePairMakerFee)
	}
	feeBuilder.IsMaker = false
	resp, err = c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if resp != WorstCaseStablePairTakerFee {
		t.Errorf(errExpectMismatch, resp, WorstCaseStablePairTakerFee)
	}
	feeBuilder.FeeType = exchange.CryptocurrencyDepositFee
	_, err = c.GetFee(context.Background(), &feeBuilder)
	assert.ErrorIs(t, err, errFeeTypeNotSupported)
	feeBuilder.Pair = currency.Pair{}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	feeBuilder.FeeType = exchange.CryptocurrencyTradeFee
	resp, err = c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if !(resp <= WorstCaseTakerFee && resp >= BestCaseTakerFee) {
		t.Errorf(errExpectedFeeRange, BestCaseTakerFee, WorstCaseTakerFee, resp)
	}
	feeBuilder.IsMaker = true
	resp, err = c.GetFee(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if !(resp <= WorstCaseMakerFee && resp >= BestCaseMakerFee) {
		t.Errorf(errExpectedFeeRange, BestCaseMakerFee, WorstCaseMakerFee, resp)
	}
}

func TestFetchTradablePairs(t *testing.T) {
	_, err := c.FetchTradablePairs(context.Background(), asset.Empty)
	assert.ErrorIs(t, err, asset.ErrNotSupported)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.FetchTradablePairs(context.Background(), asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	resp, err = c.FetchTradablePairs(context.Background(), asset.Futures)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestUpdateTradablePairs(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err := c.UpdateTradablePairs(context.Background(), false)
	assert.NoError(t, err)
}

func TestUpdateAccountInfo(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.UpdateAccountInfo(context.Background(), asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestFetchAccountInfo(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.FetchAccountInfo(context.Background(), asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestUpdateTickers(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err := c.UpdateTickers(context.Background(), asset.Futures)
	assert.NoError(t, err)
	err = c.UpdateTickers(context.Background(), asset.Spot)
	assert.NoError(t, err)
}

func TestUpdateTicker(t *testing.T) {
	_, err := c.UpdateTicker(context.Background(), currency.Pair{}, asset.Empty)
	assert.ErrorIs(t, err, currency.ErrCurrencyPairEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.UpdateTicker(context.Background(), testPair, asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestFetchTicker(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.FetchTicker(context.Background(), testPair, asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestFetchOrderbook(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.FetchOrderbook(context.Background(), testPair, asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestUpdateOrderbook(t *testing.T) {
	_, err := c.UpdateOrderbook(context.Background(), currency.Pair{}, asset.Empty)
	assert.ErrorIs(t, err, currency.ErrCurrencyPairEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err = c.UpdateOrderbook(context.Background(), currency.NewPairWithDelimiter("meow", "woof", "-"), asset.Spot)
	if err.Error() != errInvalidProductID {
		t.Errorf(errExpectMismatch, err, errInvalidProductID)
	}
	resp, err := c.UpdateOrderbook(context.Background(), testPair, asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetAccountFundingHistory(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.GetAccountFundingHistory(context.Background())
	assert.NoError(t, err)
}

func TestGetWithdrawalsHistory(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.GetWithdrawalsHistory(context.Background(), currency.NewCode("meow"), asset.Spot)
	assert.ErrorIs(t, err, errNoMatchingWallets)
	_, err = c.GetWithdrawalsHistory(context.Background(), testCrypto, asset.Spot)
	assert.NoError(t, err)
}

func TestSubmitOrder(t *testing.T) {
	_, err := c.SubmitOrder(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	var ord order.Submit
	_, err = c.SubmitOrder(context.Background(), &ord)
	assert.ErrorIs(t, err, common.ErrExchangeNameUnset)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	ord.Exchange = c.Name
	ord.Pair = testPair
	ord.AssetType = asset.Spot
	ord.Side = order.Sell
	ord.Type = order.StopLimit
	ord.StopDirection = order.StopUp
	ord.Amount = testAmount
	ord.Price = testPrice
	ord.RetrieveFees = true
	ord.ClientOrderID = strconv.FormatInt(time.Now().UnixMilli(), 18) + "GCTSubmitOrderTest"
	resp, err := c.SubmitOrder(context.Background(), &ord)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	ord.StopDirection = order.StopDown
	ord.Side = order.Buy
	resp, err = c.SubmitOrder(context.Background(), &ord)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
	ord.Type = order.Market
	ord.QuoteAmount = testAmount
	resp, err = c.SubmitOrder(context.Background(), &ord)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestModifyOrder(t *testing.T) {
	_, err := c.ModifyOrder(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	var ord order.Modify
	_, err = c.ModifyOrder(context.Background(), &ord)
	assert.ErrorIs(t, err, order.ErrPairIsEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	resp, err := c.PlaceOrder(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 18)+"GCTModifyOrderTest",
		testPair.String(), order.Sell.String(), "", order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999,
		false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	ord.OrderID = resp.OrderID
	ord.Price = testPrice + 1
	resp2, err := c.ModifyOrder(context.Background(), &ord)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp2, errExpectedNonEmpty)
}

func TestCancelOrder(t *testing.T) {
	err := c.CancelOrder(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	var can order.Cancel
	err = c.CancelOrder(context.Background(), &can)
	if err.Error() != errIDNotSet {
		t.Errorf(errExpectMismatch, err, errIDNotSet)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	can.OrderID = "0"
	err = c.CancelOrder(context.Background(), &can)
	if err.Error() != errOrder0CancelFail {
		t.Errorf(errExpectMismatch, err, errOrder0CancelFail)
	}
	skipTestIfLowOnFunds(t)
	resp, err := c.PlaceOrder(context.Background(), strconv.FormatInt(time.Now().UnixMilli(), 18)+"GCTCancelOrderTest",
		testPair.String(), order.Sell.String(), "", order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999,
		false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	can.OrderID = resp.OrderID
	err = c.CancelOrder(context.Background(), &can)
	assert.NoError(t, err)
}

func TestCancelBatchOrders(t *testing.T) {
	_, err := c.CancelBatchOrders(context.Background(), nil)
	assert.ErrorIs(t, err, errOrderIDEmpty)
	can := make([]order.Cancel, 1)
	_, err = c.CancelBatchOrders(context.Background(), can)
	if err.Error() != errIDNotSet {
		t.Errorf(errExpectMismatch, err, errIDNotSet)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	resp, err := c.PlaceOrder(context.Background(),
		strconv.FormatInt(time.Now().UnixMilli(), 18)+"GCTCancelBatchOrdersTest", testPair.String(),
		order.Sell.String(), "", order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	can[0].OrderID = resp.OrderID
	resp2, err := c.CancelBatchOrders(context.Background(), can)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp2, errExpectedNonEmpty)
}

func TestCancelAllOrders(t *testing.T) {
	_, err := c.CancelAllOrders(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	var can order.Cancel
	_, err = c.CancelAllOrders(context.Background(), &can)
	assert.ErrorIs(t, err, order.ErrPairIsEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	skipTestIfLowOnFunds(t)
	_, err = c.PlaceOrder(context.Background(),
		strconv.FormatInt(time.Now().UnixMilli(), 18)+"GCTCancelAllOrdersTest", testPair.String(),
		order.Sell.String(), "", order.Limit.String(), "", "", "", testAmount, testPrice, 0, 9999, false, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	can.Pair = testPair
	can.AssetType = asset.Spot
	resp, err := c.CancelAllOrders(context.Background(), &can)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetOrderInfo(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	ordID, err := c.GetAllOrders(context.Background(), testPair.String(), "", "", "", "",
		asset.Spot.Upper(), "", "", "", nil, nil, 2, time.Time{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(ordID.Orders) == 0 {
		t.Skip(skipInsufficientOrders)
	}
	resp, err := c.GetOrderInfo(context.Background(), ordID.Orders[0].OrderID, testPair, asset.Spot)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetDepositAddress(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	_, err := c.GetDepositAddress(context.Background(), currency.NewCode("fake currency that doesn't exist"), "", "")
	assert.ErrorIs(t, err, errNoWalletForCurrency)
	resp, err := c.GetDepositAddress(context.Background(), testCrypto, "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestWithdrawCryptocurrencyFunds(t *testing.T) {
	req := withdraw.Request{}
	_, err := c.WithdrawCryptocurrencyFunds(context.Background(), &req)
	assert.ErrorIs(t, err, common.ErrExchangeNameUnset)
	req.Exchange = c.Name
	req.Currency = testCrypto
	req.Amount = testAmount
	req.Type = withdraw.Crypto
	req.Crypto.Address = testAddress
	_, err = c.WithdrawCryptocurrencyFunds(context.Background(), &req)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wallets, err := c.GetAllWallets(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	if len(wallets.Data) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	for i := range wallets.Data {
		if wallets.Data[i].Currency.Name == testCrypto.String() && wallets.Data[i].Balance.Amount > testAmount*100 {
			req.WalletID = wallets.Data[i].ID
			break
		}
	}
	if req.WalletID == "" {
		t.Skip(skipInsufficientFunds)
	}
	resp, err := c.WithdrawCryptocurrencyFunds(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestWithdrawFiatFunds(t *testing.T) {
	withdrawFiatFundsHelper(t, c.WithdrawFiatFunds)
}

func TestWithdrawFiatFundsToInternationalBank(t *testing.T) {
	withdrawFiatFundsHelper(t, c.WithdrawFiatFundsToInternationalBank)
}

func TestGetFeeByType(t *testing.T) {
	_, err := c.GetFeeByType(context.Background(), nil)
	if err.Error() != errFeeBuilderNil {
		t.Errorf(errExpectMismatch, err, errFeeBuilderNil)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	var feeBuilder exchange.FeeBuilder
	feeBuilder.FeeType = exchange.OfflineTradeFee
	feeBuilder.Amount = 1
	feeBuilder.PurchasePrice = 1
	resp, err := c.GetFeeByType(context.Background(), &feeBuilder)
	assert.NoError(t, err)
	if resp != WorstCaseTakerFee {
		t.Errorf(errExpectMismatch, resp, WorstCaseTakerFee)
	}
}

func TestGetActiveOrders(t *testing.T) {
	_, err := c.GetActiveOrders(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	var req order.MultiOrderRequest
	_, err = c.GetActiveOrders(context.Background(), &req)
	if err.Error() != errUnsupportedAssetType {
		t.Errorf(errExpectMismatch, err, errUnsupportedAssetType)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	req.AssetType = asset.Spot
	req.Side = order.AnySide
	req.Type = order.AnyType
	_, err = c.GetActiveOrders(context.Background(), &req)
	assert.NoError(t, err)
	req.Pairs = req.Pairs.Add(currency.NewPair(testCrypto, testFiat))
	_, err = c.GetActiveOrders(context.Background(), &req)
	assert.NoError(t, err)
}

func TestGetOrderHistory(t *testing.T) {
	_, err := c.GetOrderHistory(context.Background(), nil)
	assert.ErrorIs(t, err, order.ErrGetOrdersRequestIsNil)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	var req order.MultiOrderRequest
	req.AssetType = asset.Spot
	req.Side = order.AnySide
	req.Type = order.AnyType
	_, err = c.GetOrderHistory(context.Background(), &req)
	assert.NoError(t, err)
	req.Pairs = req.Pairs.Add(testPair)
	_, err = c.GetOrderHistory(context.Background(), &req)
	assert.NoError(t, err)

}

func TestGetHistoricCandles(t *testing.T) {
	_, err := c.GetHistoricCandles(context.Background(), currency.Pair{}, asset.Empty, kline.OneYear, time.Time{},
		time.Time{})
	assert.ErrorIs(t, err, currency.ErrCurrencyPairEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetHistoricCandles(context.Background(), testPair, asset.Spot, kline.ThreeHour,
		time.Now().Add(-time.Hour*30), time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetHistoricCandlesExtended(t *testing.T) {
	_, err := c.GetHistoricCandlesExtended(context.Background(), currency.Pair{}, asset.Empty, kline.OneYear,
		time.Time{}, time.Time{})
	assert.ErrorIs(t, err, currency.ErrCurrencyPairEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetHistoricCandlesExtended(context.Background(), testPair, asset.Spot, kline.OneMin,
		time.Now().Add(-time.Hour*9), time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestValidateAPICredentials(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err := c.ValidateAPICredentials(context.Background(), asset.Spot)
	assert.NoError(t, err)
}

func TestGetServerTime(t *testing.T) {
	_, err := c.GetServerTime(context.Background(), 0)
	assert.NoError(t, err)
}

func TestGetLatestFundingRates(t *testing.T) {
	_, err := c.GetLatestFundingRates(context.Background(), nil)
	assert.ErrorIs(t, err, common.ErrNilPointer)
	req := fundingrate.LatestRateRequest{Asset: asset.UpsideProfitContract}
	_, err = c.GetLatestFundingRates(context.Background(), &req)
	if err.Error() != errUpsideUnsupported {
		t.Errorf(errExpectMismatch, err, errUpsideUnsupported)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	req.Asset = asset.Futures
	resp, err := c.GetLatestFundingRates(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestGetFuturesContractDetails(t *testing.T) {
	_, err := c.GetFuturesContractDetails(context.Background(), asset.Empty)
	assert.ErrorIs(t, err, futures.ErrNotFuturesAsset)
	_, err = c.GetFuturesContractDetails(context.Background(), asset.UpsideProfitContract)
	if err.Error() != errUpsideUnsupported {
		t.Errorf(errExpectMismatch, err, errUpsideUnsupported)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetFuturesContractDetails(context.Background(), asset.Futures)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

func TestUpdateOrderExecutionLimits(t *testing.T) {
	err := c.UpdateOrderExecutionLimits(context.Background(), asset.UpsideProfitContract)
	if err.Error() != errUpsideUnsupported {
		t.Errorf(errExpectMismatch, err, errUpsideUnsupported)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err = c.UpdateOrderExecutionLimits(context.Background(), asset.Futures)
	assert.NoError(t, err)
}

func TestFiatTransferTypeString(t *testing.T) {
	t.Parallel()
	var f FiatTransferType
	if f.String() != "deposit" {
		t.Errorf(errExpectMismatch, f.String(), "deposit")
	}
	f = FiatWithdrawal
	if f.String() != "withdrawal" {
		t.Errorf(errExpectMismatch, f.String(), "withdrawal")
	}
}

func TestUnixTimestampUnmarshalJSON(t *testing.T) {
	t.Parallel()
	var u UnixTimestamp
	err := u.UnmarshalJSON([]byte("0"))
	if err.Error() != errJSONNumberIntoString {
		t.Errorf(errExpectMismatch, err, errJSONNumberIntoString)
	}
	err = u.UnmarshalJSON([]byte("\"922337203685477580700\""))
	if err.Error() != errParseIntValueOutOfRange {
		t.Errorf(errExpectMismatch, err, errParseIntValueOutOfRange)
	}
	err = u.UnmarshalJSON([]byte("\"1234\""))
	assert.NoError(t, err)
}

func TestUnixTimestampString(t *testing.T) {
	t.Parallel()
	var u UnixTimestamp
	err := u.UnmarshalJSON([]byte("\"1234\""))
	assert.NoError(t, err)
	s := u.String()
	if s != expectedTimestamp {
		t.Errorf(errExpectMismatch, s, expectedTimestamp)
	}
}

func TestFormatExchangeKlineInterval(t *testing.T) {
	t.Parallel()
	testSequence := map[kline.Interval]string{
		kline.FiveMin:    granFiveMin,
		kline.FifteenMin: granFifteenMin,
		kline.ThirtyMin:  granThirtyMin,
		kline.TwoHour:    granTwoHour,
		kline.SixHour:    granSixHour,
		kline.OneDay:     granOneDay,
		kline.OneWeek:    errIntervalNotSupported}
	for k := range testSequence {
		resp := formatExchangeKlineInterval(k)
		if resp != testSequence[k] {
			t.Errorf(errExpectMismatch, resp, testSequence[k])
		}
	}
}

// TestWsAuth dials websocket, sends login request.
func TestWsAuth(t *testing.T) {
	if !c.Websocket.IsEnabled() && !c.API.AuthenticatedWebsocketSupport || !sharedtestvalues.AreAPICredentialsSet(c) {
		t.Skip(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := c.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		t.Fatal(err)
	}
	go c.wsReadData()

	err = c.Subscribe([]subscription.Subscription{
		{
			Channel: "user",
			Pair:    testPair,
		},
	})
	if err != nil {
		t.Error(err)
	}
	timer := time.NewTimer(sharedtestvalues.WebsocketResponseDefaultTimeout)
	select {
	case badResponse := <-c.Websocket.DataHandler:
		t.Error(badResponse)
	case <-timer.C:
	}
	timer.Stop()
}

func TestStatusToStandardStatus(t *testing.T) {
	type TestCases struct {
		Case   string
		Result order.Status
	}
	testCases := []TestCases{
		{Case: "received", Result: order.New},
		{Case: "open", Result: order.Active},
		{Case: "done", Result: order.Filled},
		{Case: "match", Result: order.PartiallyFilled},
		{Case: "change", Result: order.Active},
		{Case: "activate", Result: order.Active},
		{Case: "LOL", Result: order.UnknownStatus},
	}
	for i := range testCases {
		result, _ := statusToStandardStatus(testCases[i].Case)
		if result != testCases[i].Result {
			t.Errorf("Expected: %v, received: %v", testCases[i].Result, result)
		}
	}
}

func TestStringToFloatPtr(t *testing.T) {
	t.Parallel()
	err := stringToFloatPtr(nil, "")
	assert.ErrorIs(t, err, errPointerNil)
	var fl float64
	err = stringToFloatPtr(&fl, "")
	assert.NoError(t, err)
	err = stringToFloatPtr(&fl, "1.1")
	assert.NoError(t, err)
}

func TestWsConnect(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	err := c.Websocket.Disable()
	if err != nil && err.Error() != errCoinbaseWSAlreadyDisabled {
		t.Error(err)
	}
	err = c.WsConnect()
	if err.Error() != stream.WebsocketNotEnabled {
		t.Errorf(errExpectMismatch, err, stream.WebsocketNotEnabled)
	}
	err = c.Websocket.Enable()
	assert.NoError(t, err)
}

func TestWsHandleData(t *testing.T) {
	go func() {
		for range c.Websocket.DataHandler {
			continue
		}
	}()
	_, err := c.wsHandleData(nil, 0)
	assert.ErrorIs(t, err, jsonparser.KeyPathNotFoundError)
	mockJSON := []byte(`{"sequence_num": "l"}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errParseUintInvalidSyntax {
		t.Errorf(errExpectMismatch, err, errParseUintInvalidSyntax)
	}
	mockJSON = []byte(`{"sequence_num": 1, /\\/"""}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, jsonparser.KeyPathNotFoundError)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "subscriptions"}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "", "events":}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, jsonparser.UnknownValueTypeError)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "status", "events": ["type": 1234]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "status", "events": [{"type": "moo"}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "error", "events": [{"type": "moo"}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "ticker", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "ticker", "events": [{"type": "moo", "tickers": [{"price": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, jsonparser.KeyPathNotFoundError)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "ticker", "timestamp": "2006-01-02T15:04:05Z", "events": [{"type": "moo", "tickers": [{"price": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "candles", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "candles", "events": [{"type": "moo", "candles": [{"low": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, jsonparser.KeyPathNotFoundError)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "candles", "timestamp": "2006-01-02T15:04:05Z", "events": [{"type": "moo", "candles": [{"low": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "market_trades", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "market_trades", "events": [{"type": "moo", "trades": [{"price": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "l2_data", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "l2_data", "events": [{"type": "moo", "updates": [{"price_level": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, jsonparser.KeyPathNotFoundError)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "l2_data", "timestamp": "2006-01-02T15:04:05Z", "events": [{"type": "moo", "updates": [{"price_level": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errL2DataMoo {
		t.Errorf(errExpectMismatch, err, errL2DataMoo)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "l2_data", "timestamp": "2006-01-02T15:04:05Z", "events": [{"type": "snapshot", "product_id": "BTC-USD", "updates": [{"side": "bid", "price_level": "1.1", "new_quantity": "2.2"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "l2_data", "timestamp": "2006-01-02T15:04:05Z", "events": [{"type": "update", "product_id": "BTC-USD", "updates": [{"side": "bid", "price_level": "1.1", "new_quantity": "2.2"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "user", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errJSONInvalidCharacter {
		t.Errorf(errExpectMismatch, err, errJSONInvalidCharacter)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "user", "events": [{"type": "moo", "orders": [{"total_fees": "1.1"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errUnrecognisedOrderType {
		t.Errorf(errExpectMismatch, err, errUnrecognisedOrderType)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "user", "events": [{"type": "moo", "orders": [{"total_fees": "1.1", "order_type": "ioc"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errOrderSideInvalid {
		t.Errorf(errExpectMismatch, err, errOrderSideInvalid)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "user", "events": [{"type": "moo", "orders": [{"total_fees": "1.1", "order_type": "ioc", "order_side": "buy"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	if err.Error() != errUnrecognisedStatusType {
		t.Errorf(errExpectMismatch, err, errUnrecognisedStatusType)
	}
	mockJSON = []byte(`{"sequence_num": 0, "channel": "user", "events": [{"type": "moo", "orders": [{"total_fees": "1.1", "order_type": "ioc", "order_side": "buy", "status": "done"}]}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.NoError(t, err)
	mockJSON = []byte(`{"sequence_num": 0, "channel": "fakechan", "events": ["type": ""}]}`)
	_, err = c.wsHandleData(mockJSON, 0)
	assert.ErrorIs(t, err, errChannelNameUnknown)
}

func TestProcessSnapshotUpdate(t *testing.T) {
	req := WebsocketOrderbookDataHolder{Changes: []WebsocketOrderbookData{{Side: "fakeside", PriceLevel: 1.1,
		NewQuantity: 2.2}}, ProductID: currency.NewBTCUSD()}
	err := c.ProcessSnapshot(&req, time.Time{})
	if err.Error() != errFakeSide {
		t.Errorf(errExpectMismatch, err, errFakeSide)
	}
	err = c.ProcessUpdate(&req, time.Time{})
	if err.Error() != errFakeSide {
		t.Errorf(errExpectMismatch, err, errFakeSide)
	}
	req.Changes[0].Side = "offer"
	err = c.ProcessSnapshot(&req, time.Now())
	assert.NoError(t, err)
	err = c.ProcessUpdate(&req, time.Now())
	assert.NoError(t, err)
}

func TestGenerateDefaultSubscriptions(t *testing.T) {
	comparison := []subscription.Subscription{{Channel: "heartbeats"}, {Channel: "status"}, {Channel: "ticker"},
		{Channel: "ticker_batch"}, {Channel: "candles"}, {Channel: "market_trades"}, {Channel: "level2"},
		{Channel: "user"}}
	for i := range comparison {
		comparison[i].Pair = currency.NewPairWithDelimiter(testCrypto.String(), testFiat.String(), "-")
		comparison[i].Asset = asset.Spot
	}
	resp, err := c.GenerateDefaultSubscriptions()
	if err != nil {
		t.Fatal(err)
	}
	assert.ElementsMatch(t, comparison, resp)
}

func TestSubscribeUnsubscribe(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	req := []subscription.Subscription{{Channel: "heartbeats", Asset: asset.Spot,
		Pair: currency.NewPairWithDelimiter(testCrypto.String(), testFiat.String(), "-")}}
	err := c.Subscribe(req)
	assert.NoError(t, err)
	err = c.Unsubscribe(req)
	assert.NoError(t, err)
}

func TestGetJWT(t *testing.T) {
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	creds, err := c.GetCredentials(context.Background())
	assert.NoError(t, err)
	_, err = c.GetJWT(context.Background(), "")
	if strings.HasPrefix(creds.Secret, "-----BEGIN EC PRIVATE KEY-----\n") {
		assert.NoError(t, err)
	} else {
		assert.ErrorIs(t, err, errCantDecodePrivKey)
	}
}

func skipTestIfLowOnFunds(t *testing.T) {
	t.Helper()
	accounts, err := c.GetAllAccounts(context.Background(), 250, "")
	assert.NoError(t, err)
	if len(accounts.Accounts) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	var hasValidFunds bool
	for i := range accounts.Accounts {
		if accounts.Accounts[i].Currency == testCrypto.String() && accounts.Accounts[i].AvailableBalance.Value > testAmount*100 {
			hasValidFunds = true
		}
	}
	if !hasValidFunds {
		t.Skip(skipInsufficientFunds)
	}
}

func portfolioIDFromName(t *testing.T, targetName string) string {
	t.Helper()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	createResp, err := c.CreatePortfolio(context.Background(), targetName)
	var targetID string
	if err != nil {
		if err.Error() != errPortfolioNameDuplicate {
			t.Error(err)
		}
		getResp, err := c.GetAllPortfolios(context.Background(), "")
		if err != nil {
			t.Error(err)
		}
		if len(getResp.Portfolios) == 0 {
			t.Fatal(errExpectedNonEmpty)
		}
		for i := range getResp.Portfolios {
			if getResp.Portfolios[i].Name == targetName {
				targetID = getResp.Portfolios[i].UUID
				break
			}
		}
	} else {
		targetID = createResp.Portfolio.UUID
	}
	return targetID
}

func getINTXPortfolio(t *testing.T) string {
	t.Helper()
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	resp, err := c.GetAllPortfolios(context.Background(), "")
	assert.NoError(t, err)
	if len(resp.Portfolios) == 0 {
		t.Skip(skipInsufficientPortfolios)
	}
	var targetID string
	for i := range resp.Portfolios {
		if resp.Portfolios[i].Type == "INTX" {
			targetID = resp.Portfolios[i].UUID
			break
		}
	}
	if targetID == "" {
		t.Skip(skipInsufficientPortfolios)
	}
	return targetID
}

func convertTestHelper(t *testing.T) (fromAccID, toAccID string) {
	t.Helper()
	accIDs, err := c.GetAllAccounts(context.Background(), 250, "")
	assert.NoError(t, err)
	if len(accIDs.Accounts) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	for x := range accIDs.Accounts {
		if accIDs.Accounts[x].Currency == "USDC" {
			fromAccID = accIDs.Accounts[x].UUID
		}
		if accIDs.Accounts[x].Currency == "USD" {
			toAccID = accIDs.Accounts[x].UUID
		}
		if fromAccID != "" && toAccID != "" {
			break
		}
	}
	if fromAccID == "" || toAccID == "" {
		t.Skip(skipInsufSuitableAccs)
	}
	return fromAccID, toAccID
}

func transferTestHelper(t *testing.T, wallets GetAllWalletsResponse) (srcWalletID, tarWalletID string) {
	t.Helper()
	var hasValidFunds bool
	for i := range wallets.Data {
		if wallets.Data[i].Currency.Code == testFiat.String() && wallets.Data[i].Balance.Amount > 10 {
			hasValidFunds = true
			srcWalletID = wallets.Data[i].ID
		}
	}
	if !hasValidFunds {
		t.Skip(skipInsufficientFunds)
	}
	pmID, err := c.GetAllPaymentMethods(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	if len(pmID.Data) == 0 {
		t.Skip(skipPayMethodNotFound)
	}
	return srcWalletID, pmID.Data[0].FiatAccount.ID
}

type withdrawFiatFunc func(context.Context, *withdraw.Request) (*withdraw.ExchangeResponse, error)

func withdrawFiatFundsHelper(t *testing.T, fn withdrawFiatFunc) {
	t.Helper()
	req := withdraw.Request{}
	_, err := fn(context.Background(), &req)
	assert.ErrorIs(t, err, common.ErrExchangeNameUnset)
	req.Exchange = c.Name
	req.Currency = testFiat
	req.Amount = 1
	req.Type = withdraw.Fiat
	req.Fiat.Bank.Enabled = true
	req.Fiat.Bank.SupportedExchanges = "CoinbasePro"
	req.Fiat.Bank.SupportedCurrencies = testFiat.String()
	req.Fiat.Bank.AccountNumber = "123"
	req.Fiat.Bank.SWIFTCode = "456"
	req.Fiat.Bank.BSBNumber = "789"
	_, err = fn(context.Background(), &req)
	assert.ErrorIs(t, err, errWalletIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c)
	req.WalletID = "meow"
	req.Fiat.Bank.BankName = "GCT's Fake and Not Real Test Bank Meow Meow Meow"
	expectedError := fmt.Sprintf(errPayMethodNotFound, req.Fiat.Bank.BankName)
	_, err = fn(context.Background(), &req)
	if err.Error() != expectedError {
		t.Errorf(errExpectMismatch, err, expectedError)
	}
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	wallets, err := c.GetAllWallets(context.Background(), PaginationInp{})
	assert.NoError(t, err)
	if len(wallets.Data) == 0 {
		t.Fatal(errExpectedNonEmpty)
	}
	req.WalletID = ""
	for i := range wallets.Data {
		if wallets.Data[i].Currency.Name == testFiat.String() && wallets.Data[i].Balance.Amount > testAmount*100 {
			req.WalletID = wallets.Data[i].ID
			break
		}
	}
	if req.WalletID == "" {
		t.Skip(skipInsufficientFunds)
	}
	req.Fiat.Bank.BankName = "AUD Wallet"
	resp, err := fn(context.Background(), &req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

type getNoArgsResp interface {
	AllFuturesPositions | ServerTimeV3 | *UserResponse | AuthResponse | GetFiatCurrenciesResp |
		GetCryptocurrenciesResp | ServerTimeV2
}

type getNoArgsAssertNotEmpty[G getNoArgsResp] func(context.Context) (G, error)

func testGetNoArgs[G getNoArgsResp](t *testing.T, f getNoArgsAssertNotEmpty[G]) {
	t.Helper()
	resp, err := f(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}

type genConvertTestFunc func(context.Context, string, string, string) (ConvertResponse, error)

func convertTestShared(t *testing.T, f genConvertTestFunc) {
	t.Helper()
	_, err := f(context.Background(), "", "", "")
	assert.ErrorIs(t, err, errTransactionIDEmpty)
	_, err = f(context.Background(), "meow", "", "")
	assert.ErrorIs(t, err, errAccountIDEmpty)
	sharedtestvalues.SkipTestIfCredentialsUnset(t, c, canManipulateRealOrders)
	fromAccID, toAccID := convertTestHelper(t)
	resp, err := c.CreateConvertQuote(context.Background(), fromAccID, toAccID, "", "", 0.01)
	assert.NoError(t, err)
	resp, err = f(context.Background(), resp.Trade.ID, fromAccID, toAccID)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp, errExpectedNonEmpty)
}
