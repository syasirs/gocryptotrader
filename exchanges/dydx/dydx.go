package dydx

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/yanue/starkex"
)

// DYDX is the overarching type across this package
type DYDX struct {
	exchange.Base
}

const (
	dydxOnlySignOnDomainMainnet = "https://trade.dydx.exchange"
	dydxAPIURL                  = "https://api.dydx.exchange/" + dydxAPIVersion
	dydxAPIVersion              = "v3/"
	dydxWSAPIURL                = "wss://api.dydx.exchange/" + dydxAPIVersion + "ws"

	// Public endpoints
	markets                        = "markets"
	marketOrderbook                = "orderbook/%s" // orderbook/:market
	marketTrades                   = "trades/%s"    // trades/:market
	fastWithdrawals                = "fast-withdrawals"
	marketStats                    = "stats/%s"              // stats/:market
	marketHistoricalFunds          = "historical-funding/%s" // historical-funding/:market
	marketCandles                  = "candles/%s"            // candles/:market
	globalConfigurations           = "config"
	usersExists                    = "users/exists"
	usernameExists                 = "usernames"
	apiServerTime                  = "time"
	leaderboardPNL                 = "leaderboard-pnl"
	publicRetroactiveMiningRewards = "rewards/public-retroactive-mining"
	verifyEmailAddress             = "emails/verify-email"
	historicalHedgies              = "hedgies/history"
	currentHedgies                 = "hedgies/current"
	insuranceFundBalance           = "insurance-fund/balance"
	publicProifle                  = "profile/%s" // profile/:publicId

	// Authenticated endpoints
	onboarding                      = "onboarding"
	recovery                        = "recovery"
	registration                    = "registration"
	apiKeys                         = "api-keys"
	users                           = "users"
	userActiveLinks                 = "users/links"
	userPendingLinkRequests         = "users/links/requests"
	accounts                        = "accounts"
	accountIDs                      = "accounts/%s"                             // accounts/:id
	accountLeaderBoardPNL           = "accounts/leaderboard-pnl/%s"             // accounts/leaderboard-pnl/:period
	accountHistoricalLeaderboardPNL = "accounts/historical-leaderboard-pnls/%s" // accounts/historical-leaderboard-pnls/:period
	positions                       = "positions"
	transfers                       = "transfers"
	withdrawals                     = "withdrawals"
	orders                          = "orders"
	orderByID                       = "orders/%s"        // orders/:id
	activeOrders                    = "active-orders"    // active-orders
	orderClientID                   = "orders/client/%s" // orders/client/:id
	fills                           = "fills"
	funding                         = "funding"
	historicalPNL                   = "historical-pnl"
	rewardsWeight                   = "rewards/weight"
	rewardsLiquidityProvider        = "rewards/liquidity-provider"
	liquidityRewards                = "rewards/liquidity"
	rewardsRetroactiveMining        = "rewards/retroactive-mining"
	emailsSendVeroficationEmail     = "emails/send-verification-email"
	testnetTokens                   = "testnet/tokens"
	privateProfile                  = "profile/private"
)

var (
	errMissingMarketInstrument     = errors.New("missing market instrument")
	errInvalidPeriod               = errors.New("invalid period specified")
	errSortByIsRequired            = errors.New("parameter \"sortBy\" is required")
	errMissingPublicID             = errors.New("missing user public id")
	errInvalidSendRequestAction    = errors.New("invalid send request action")
	errInvalidStarkCredentials     = errors.New("invalid stark key credentials")
	errInvalidTransferType         = errors.New("invalid transfer type")
	errInvalidAmount               = errors.New("amount must be greater than zero")
	errInvalidMarket               = errors.New("missing market name")
	errInvalidSide                 = errors.New("invalid order side")
	errInvalidPrice                = errors.New("invalid order price")
	errInvalidExpirationTimeString = errors.New("invalid expiration time string")
	errMissingEthereumAddress      = errors.New("ethereum address is required")
	errMissingPublicKey            = errors.New("missing public key")
)

// GetMarkets retrives one or all markets as well as metadata about each retrieved market.
func (dy *DYDX) GetMarkets(ctx context.Context, instrument string) (*InstrumentDatas, error) {
	params := url.Values{}
	if instrument != "" {
		params.Set("market", instrument)
	}
	var resp *InstrumentDatas
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(markets, params), &resp)
}

// GetOrderbooks retrives  the active orderbook for a market. All bids and asks that are fillable are returned.
func (dy *DYDX) GetOrderbooks(ctx context.Context, instrument string) (*MarketOrderbook, error) {
	if instrument == "" {
		return nil, errMissingMarketInstrument
	}
	var resp *MarketOrderbook
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, fmt.Sprintf(marketOrderbook, instrument), &resp)
}

// GetTrades retrives Trades by specified parameters. Passing in all query parameters to the HTTP endpoint would look like.
func (dy *DYDX) GetTrades(ctx context.Context, instrument string, startingBeforeOrAT time.Time, limit int64) ([]MarketTrade, error) {
	if instrument == "" {
		return nil, errMissingMarketInstrument
	}
	params := url.Values{}
	if !startingBeforeOrAT.IsZero() {
		params.Set("startingBeforeOrAt", startingBeforeOrAT.UTC().Format(timeFormat))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *MarketTrades
	return resp.Trades, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(fmt.Sprintf(marketTrades, instrument), params), &resp)
}

// GetFastWithdrawalLiquidity returns a map of all LP provider accounts that have available funds for fast withdrawals.
// Given a debitAmount and asset the user wants sent to L1, this endpoint also returns the predicted amount of the desired asset the user will be credited on L1.
// Given a creditAmount and asset the user wants sent to L1,
// this endpoint also returns the predicted amount the user will be debited on L2.
func (dy *DYDX) GetFastWithdrawalLiquidity(ctx context.Context, param FastWithdrawalRequestParam) (map[string]LiquidityProvider, error) {
	if (param.CreditAmount != 0 || param.DebitAmount != 0) && param.CreditAsset == "" {
		return nil, errors.New("cannot find quote without creditAsset")
	} else if param.CreditAmount != 0 && param.DebitAmount != 0 {
		return nil, errors.New("creditAmount and debitAmount cannot both be set")
	}
	params := url.Values{}
	if param.CreditAsset != "" {
		params.Set("creditAsset", param.CreditAsset)
	}
	if param.CreditAmount != 0 {
		params.Set("creditAmount", strconv.FormatFloat(param.CreditAmount, 'f', -1, 64))
	}
	if param.DebitAmount != 0 {
		params.Set("debitAmount", strconv.FormatFloat(param.DebitAmount, 'f', -1, 64))
	}
	var resp *WithdrawalLiquidityResponse
	return resp.LiquidityProviders, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(fastWithdrawals, params), &resp)
}

// GetMarketStats retrives an individual market's statistics over a set period of time or all available periods of time.
func (dy *DYDX) GetMarketStats(ctx context.Context, instrument string, days int64) (map[string]TickerData, error) {
	params := url.Values{}
	if days != 0 {
		if days != 1 && days != 7 && days != 30 {
			return nil, errors.New("only 1,7, and 30 days are allowed")
		}
		params.Set("days", strconv.FormatInt(days, 10))
	}
	var resp *TickerDatas
	return resp.Markets, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, fmt.Sprintf(marketStats, instrument), &resp)
}

// GetHistoricalFunding retrives the historical funding rates for a market.
func (dy *DYDX) GetHistoricalFunding(ctx context.Context, instrument string, effectiveBeforeOrAt time.Time) ([]HistoricalFunding, error) {
	if instrument == "" {
		return nil, errMissingMarketInstrument
	}
	params := url.Values{}
	if !effectiveBeforeOrAt.IsZero() {
		params.Set("effectiveBeforeOrAt", effectiveBeforeOrAt.UTC().Format(timeFormat))
	}
	var resp *HistoricFundingResponse
	return resp.HistoricalFundings, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(fmt.Sprintf(marketHistoricalFunds, instrument), params), &resp)
}

// GetResolutionFromInterval returns the resolution(string representation of interval) from interval instance if supported by the exchange.
func (dy *DYDX) GetResolutionFromInterval(interval kline.Interval) (string, error) {
	switch interval {
	case kline.OneMin:
		return "1MIN", nil
	case kline.FiveMin:
		return "5MINS", nil
	case kline.FifteenMin:
		return "15MINS", nil
	case kline.ThirtyMin:
		return "30MINS", nil
	case kline.OneHour:
		return "1HOUR", nil
	case kline.FourHour:
		return "4HOURS", nil
	case kline.OneDay:
		return "1DAY", nil
	default:
		return "", kline.ErrUnsupportedInterval
	}
}

// GetCandlesForMarket retrives the candle statistics for a market.
func (dy *DYDX) GetCandlesForMarket(ctx context.Context, instrument string, interval kline.Interval, fromISO, toISO string, limit int64) ([]MarketCandle, error) {
	if instrument == "" {
		return nil, errMissingMarketInstrument
	}
	resolution, err := dy.GetResolutionFromInterval(interval)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("resolution", resolution)
	if fromISO != "" {
		params.Set("fromISO", fromISO)
	}
	if toISO != "" {
		params.Set("toISO", toISO)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *MarketCandlesResponse
	return resp.Candles, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(fmt.Sprintf(marketCandles, instrument), params), &resp)
}

// GetGlobalConfigurationVariables retrives any global configuration variables for the exchange as a whole.
func (dy *DYDX) GetGlobalConfigurationVariables(ctx context.Context) (*ConfigurationVariableResponse, error) {
	var resp *ConfigurationVariableResponse
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, globalConfigurations, &resp)
}

// CheckIfUserExists checks if a user exists for a given Ethereum address.
func (dy *DYDX) CheckIfUserExists(ctx context.Context, etheriumAddress string) (bool, error) {
	resp := &struct {
		Exists bool `json:"exists"`
	}{}
	return resp.Exists, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, usersExists+"?ethereumAddress="+etheriumAddress, resp)
}

// CheckIfUsernameExists check if a username has been taken by a user.
func (dy *DYDX) CheckIfUsernameExists(ctx context.Context, username string) (bool, error) {
	resp := &struct {
		Exists bool `json:"exists"`
	}{}
	return resp.Exists, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, usernameExists+"?username="+username, resp)
}

// GetAPIServerTime get the current time of the API server.
func (dy *DYDX) GetAPIServerTime(ctx context.Context) (*APIServerTime, error) {
	var resp *APIServerTime
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, apiServerTime, &resp)
}

// GetPublicLeaderboardPNLs retrives the top PNLs for a specified period and how they rank against each other.
func (dy *DYDX) GetPublicLeaderboardPNLs(ctx context.Context, period, sortBy string, startingBeforeOrAt time.Time, limit int64) (*LeaderboardPNLs, error) {
	if period == "" {
		return nil, fmt.Errorf("%w \"period\" is required", errInvalidPeriod)
	}
	if sortBy == "" {
		return nil, errSortByIsRequired
	}
	params := url.Values{}
	params.Set("period", period)
	if !startingBeforeOrAt.IsZero() {
		params.Set("startingBeforeOrAt", startingBeforeOrAt.Format(timeFormat))
	}
	params.Set("sortBy", sortBy)
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *LeaderboardPNLs
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(leaderboardPNL, params), &resp)
}

// GetPublicRetroactiveMiningReqards retrives the retroactive mining rewards for an ethereum address.
func (dy *DYDX) GetPublicRetroactiveMiningReqards(ctx context.Context, ethereumAddress string) (*RetroactiveMiningReward, error) {
	var resp *RetroactiveMiningReward
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, publicRetroactiveMiningRewards+"?ethereumAddress="+ethereumAddress, &resp)
}

// VerifyEmailAddress verify an email address by providing the verification token sent to the email address.
func (dy *DYDX) VerifyEmailAddress(ctx context.Context, token string) (interface{}, error) {
	var response interface{}
	return response, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, verifyEmailAddress+"?token="+token, response)
}

// GetCurrentlyRevealedHedgies retrives the currently revealed Hedgies for competition distribution.
func (dy *DYDX) GetCurrentlyRevealedHedgies(ctx context.Context, daily, weekly string) (*CurrentRevealedHedgies, error) {
	params := url.Values{}
	if daily != "" {
		params.Set("daily", daily)
	}
	if weekly != "" {
		params.Set("weekly", weekly)
	}
	var resp *CurrentRevealedHedgies
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, currentHedgies, &resp)
}

// GetHistoricallyRevealedHedgies retrives the historically revealed Hedgies from competition distributions.
func (dy *DYDX) GetHistoricallyRevealedHedgies(ctx context.Context, nftRevealType string, start, end int64) (*HistoricalRevealedHedgies, error) {
	params := url.Values{}
	if nftRevealType != "" {
		params.Set("nftRevealType", nftRevealType)
	}
	if start != 0 {
		params.Set("start", strconv.FormatInt(start, 10))
	}
	if end != 0 {
		params.Set("end", strconv.FormatInt(end, 10))
	}
	var resp *HistoricalRevealedHedgies
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, common.EncodeURLValues(historicalHedgies, params), &resp)
}

// GetInsuranceFundBalance retrives the balance of dydx insurance fund.
func (dy *DYDX) GetInsuranceFundBalance(ctx context.Context) (*InsuranceFundBalance, error) {
	var resp *InsuranceFundBalance
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, insuranceFundBalance, &resp)
}

// GetPublicProfile retrives the public profile of a user given their public id.
func (dy *DYDX) GetPublicProfile(ctx context.Context, publicID string) (*PublicProfile, error) {
	if publicID == "" {
		return nil, errMissingPublicID
	}
	var resp *PublicProfile
	return resp, dy.SendHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, fmt.Sprintf(publicProifle, publicID), &resp)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (dy *DYDX) SendHTTPRequest(ctx context.Context, endpoint exchange.URL, epl request.EndpointLimit, path string, result interface{}) error {
	urlPath, err := dy.API.Endpoints.GetURL(endpoint)
	if err != nil {
		return err
	}
	item := &request.Item{
		Method:        http.MethodGet,
		Path:          urlPath + path,
		Result:        result,
		Verbose:       dy.Verbose,
		HTTPDebugging: dy.HTTPDebugging,
		HTTPRecording: dy.HTTPRecording,
	}
	return dy.SendPayload(ctx, epl, func() (*request.Item, error) {
		return item, nil
	})
}

// GetPositions retrives all current positions for a user by specified query parameters.
func (dy *DYDX) GetPositions(ctx context.Context, market, status, createdBeforeOrAt string, limit int64) (*Position, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if status != "" {
		params.Set("status", status)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if createdBeforeOrAt != "" {
		params.Set("createdBeforeOrAt", createdBeforeOrAt)
	}
	var resp *Position
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(positions, params), nil, &resp)
}

// GetUsers returns the user and user information.
func (dy *DYDX) GetUsers(ctx context.Context) (*UsersResponse, error) {
	var resp *UsersResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, users, nil, &resp)
}

// Updateusers update user information and return the updated user.
func (dy *DYDX) Updateusers(ctx context.Context, params *UpdateUserParams) (*UsersResponse, error) {
	var resp *UsersResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPut, users, &params, &resp)
}

// GetUserActiveLinks return active user links.
func (dy *DYDX) GetUserActiveLinks(ctx context.Context, userType, primaryAddress, linkedAddress string) (*UserActiveLink, error) {
	params := url.Values{}
	if userType != "" {
		params.Set("userType", userType)
	}
	if primaryAddress != "" {
		params.Set("primaryAddress", primaryAddress)
	}
	if linkedAddress != "" {
		params.Set("linkedAddress", linkedAddress)
	}
	var resp *UserActiveLink
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(userActiveLinks, params), nil, &resp)
}

// SendUserLinkRequest end a new request to link users, respond to a pending request, or remove a link.
func (dy *DYDX) SendUserLinkRequest(ctx context.Context, params UserLinkParams) (interface{}, error) {
	if params.Action != "CREATE_SECONDARY_REQUEST" &&
		params.Action != "DELETE_SECONDARY_REQUEST" &&
		params.Action != "ACCEPT_PRIMARY_REQUEST" &&
		params.Action != "REJECT_PRIMARY_REQUEST" &&
		params.Action != "REMOVE" {
		return nil, errInvalidSendRequestAction
	}
	creds, err := dy.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if creds.ClientID == params.Address {
		return nil, errors.New("address should not be your address")
	}
	var resp interface{}
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, userActiveLinks, params, &resp)
}

// GetUserPendingLinkRequest retrieves pending user links.
func (dy *DYDX) GetUserPendingLinkRequest(ctx context.Context, userType, outgoinRequests, incomingRequests string) (*UserPendingLink, error) {
	params := url.Values{}
	if userType != "" {
		params.Set("userType", userType)
	}
	if outgoinRequests != "" {
		params.Set("outgoinRequests", outgoinRequests)
	}
	if incomingRequests != "" {
		params.Set("incomingRequests", incomingRequests)
	}
	var resp *UserPendingLink
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(userPendingLinkRequests, params), nil, &resp)
}

// CreateAccount represents a new account instance created using the provided stark Key credentials.
func (dy *DYDX) CreateAccount(ctx context.Context, starkKey, starkYCoordinate string) (*AccountResponse, error) {
	if starkKey == "" {
		return nil, fmt.Errorf("%w missing \"starkKey\"", errInvalidStarkCredentials)
	}
	if starkYCoordinate == "" {
		return nil, fmt.Errorf("%w missing \"starkYCoordinate\"", errInvalidStarkCredentials)
	}
	param := map[string]string{"starkKey": starkKey, "starkKeyYCoordinate": starkYCoordinate}
	var resp *AccountResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, accounts, param, &resp)
}

// GetAccount retrives an account for a user by id. Using the client, the id will be generated with client information and an Ethereum address.
func (dy *DYDX) GetAccount(ctx context.Context, etheriumAddress string) (*AccountResponse, error) {
	params := url.Values{}
	if etheriumAddress != "" {
		params.Set("etheriumAddress", etheriumAddress)
	}
	var resp *AccountResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(accounts, params), nil, &resp)
}

// GetAccountLeaderboardPNLs represents an account's personal leaderboard pnls.
func (dy *DYDX) GetAccountLeaderboardPNLs(ctx context.Context, period string, startingBeforeOrAt time.Time) (*AccountLeaderboardPNL, error) {
	if period == "" {
		return nil, errInvalidPeriod
	}
	param := url.Values{}
	if !startingBeforeOrAt.IsZero() {
		param.Set("startingBeforeOrAt", startingBeforeOrAt.UTC().Format(timeFormat))
	}
	var resp *AccountLeaderboardPNL
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(fmt.Sprintf(accountLeaderBoardPNL, strings.ToUpper(period)), param), nil, &resp)
}

// GetAccountHistoricalLeaderboardPNLs retrives  an account's historical leaderboard pnls.
func (dy *DYDX) GetAccountHistoricalLeaderboardPNLs(ctx context.Context, period string, limit int64) (*AccountHistorical, error) {
	period = strings.ToUpper(period)
	if period == "" {
		return nil, errInvalidPeriod
	}
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	var resp *AccountHistorical
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(fmt.Sprintf(accountHistoricalLeaderboardPNL, period), params), nil, &resp)
}

// GetAccounts all accounts for a user.
func (dy *DYDX) GetAccounts(ctx context.Context) (*AccountsResponse, error) {
	var resp *AccountsResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, accounts, nil, &resp)
}

// GetPosition retrives all current positions for a user by specified query parameters.
func (dy *DYDX) GetPosition(ctx context.Context, market, status string, limit int64, createdBeforeOrAt time.Time) (*Position, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if status != "" {
		params.Set("status", status)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !createdBeforeOrAt.IsZero() {
		params.Set("createdBeforeOrAt", createdBeforeOrAt.UTC().Format(timeFormat))
	}
	var resp *Position
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(positions, params), nil, &resp)
}

// GetTransfers retrives transfers for a user, limited by query parameters.
func (dy *DYDX) GetTransfers(ctx context.Context, transferType string, limit int64, createdBeforeOrAt time.Time) (*TransfersResponse, error) {
	transferType = strings.ToUpper(transferType)
	if transferType != "DEPOSIT" && transferType != "WITHDRAWAL" && transferType != "FAST_WITHDRAWAL" {
		return nil, fmt.Errorf("%w %s, %s, or %s are supported", errInvalidTransferType, "DEPOSIT", "WITHDRAWAL", "FAST_WITHDRAWAL")
	}
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !createdBeforeOrAt.IsZero() {
		params.Set("createdBeforeOrAt", createdBeforeOrAt.UTC().Format(timeFormat))
	}
	var resp *TransfersResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(transfers, params), nil, &resp)
}

// CreateTransfer sends a StarkEx L2 transfer.
func (dy *DYDX) CreateTransfer(ctx context.Context, param *TransferParam) (*TransferResponse, error) {
	if param.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if param.ReceiverAccountID == "" {
		return nil, errors.New("invalid receiver account id")
	}
	if param.ReceiverPublicKey == "" {
		return nil, errors.New("invalid stark receiver public key")
	}
	creds, err := dy.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if param.ClientID == "" {
		param.ClientID = strconv.FormatInt(dy.Websocket.Conn.GenerateMessageID(true), 10)
	}
	var signature string
	signature, err = starkex.TransferSign(creds.SubAccount, starkex.TransferSignParam{
		NetworkId: 1,
		ClientId:  param.ClientID,
		// TODO: More information to generate signature.
	})
	if err != nil {
		return nil, err
	}
	resp := &struct {
		Transfer TransferResponse `json:"transfer"`
	}{}
	param.Signature = signature
	return &resp.Transfer, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, transfers, param, &resp)
}

// CreateWithdrawal create a withdrawal from an account.
// An additional L1 transaction has to be sent to the Starkware contract to retrieve funds after a slow withdrawal. This cannot be done until the zero-knowledge proof for the block has been constructed and verified on-chain.
// For the L1 transaction, the Ethereum address that the starkKey is registered to must call either the withdraw or withdrawTo smart-contract functions. The contract ABI is not tied to a particular client but can be accessed via a client. All withdrawable funds are withdrawn at once.
func (dy *DYDX) CreateWithdrawal(ctx context.Context, arg WithdrawalParam) (*TransferResponse, error) {
	if arg.Amount <= 0 {
		return nil, errInvalidAmount
	}
	if arg.Asset == "" {
		return nil, fmt.Errorf("%w parameter: asset", currency.ErrCurrencyCodeEmpty)
	}
	if arg.Expiration == "" {
		return nil, errInvalidExpirationTimeString
	}
	creds, err := dy.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if arg.ClientGeneratedID == "" {
		arg.ClientGeneratedID = strconv.FormatInt(dy.Websocket.Conn.GenerateMessageID(true), 10)
	}
	var signature string
	signature, err = starkex.WithdrawSign(creds.SubAccount, starkex.WithdrawSignParam{
		NetworkId:   1,
		HumanAmount: strconv.FormatFloat(arg.Amount, 'f', -1, 64),
		Expiration:  arg.Expiration,
		ClientId:    arg.ClientGeneratedID,
	})
	if err != nil {
		return nil, err
	}
	arg.Signature = signature
	var resp *TransferResponse
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, withdrawals, &arg, &resp)
}

// CreateFastWithdrawal creates a fast withdrawal. Fast withdrawals utilize a withdrawal liquidity provider to send funds immediately and do not require users to wait for a Layer 2 block to be mined.
// Users do not need to send any Transactions to perform a fast withdrawal.
// Behind the scenes, the withdrawal liquidity provider will immediately send a transaction to Ethereum which, once mined, will send the user their funds.
// Users must pay a fee to the liquidity provider for fast withdrawals equal to the greater of the gas fee the provider must pay and 0.1% of the amount of the withdraw.
func (dy *DYDX) CreateFastWithdrawal(ctx context.Context, param *FastWithdrawalParam) (*TransferResponse, error) {
	if param.CreditAsset == "" {
		return nil, fmt.Errorf("%w parameter: creditAsset", currency.ErrCurrencyCodeEmpty)
	}
	if param.CreditAmount <= 0 {
		return nil, fmt.Errorf("%w parameter: creditAmount", errInvalidAmount)
	}
	if param.DebitAmount <= 0 {
		return nil, fmt.Errorf("%w parameter: debitAmount", errInvalidAmount)
	}
	if param.SlippageTolerance < 0 || param.SlippageTolerance > 1 {
		return nil, fmt.Errorf("slippageTolerance has to be less than 1 and grater than 0 but passed %f", param.SlippageTolerance)
	}
	if param.ToAddress == "" {
		// Address to be credited
		return nil, fmt.Errorf("address to be credited must not be empty")
	}
	creds, err := dy.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	// Here SubAccount represents the starkx private account
	signature, err := starkex.WithdrawSign(creds.SubAccount, starkex.WithdrawSignParam{
		NetworkId:   1,
		ClientId:    param.ClientID,
		PositionId:  int64(param.LPPositionID),
		HumanAmount: strconv.FormatFloat(param.CreditAmount, 'f', -1, 64),
		Expiration:  param.Expiration,
	})
	if err != nil {
		return nil, err
	}
	param.Signature = signature
	var resp *WithdrawalResponse
	return &resp.Withdrawal, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, fastWithdrawals, param, &resp)
}

// CreateNewOrder creates a new order.
func (dy *DYDX) CreateNewOrder(ctx context.Context, arg *CreateOrderRequestParams) (*Order, error) {
	if arg.Market == "" {
		return nil, errInvalidMarket
	}
	if arg.Side == "" {
		return nil, errInvalidSide
	}
	if arg.Type == "" {
		return nil, order.ErrTypeIsInvalid
	}
	if arg.Size <= 0 {
		return nil, fmt.Errorf("%w order size have to be greater than zero", errInvalidAmount)
	}
	if arg.Price <= 0 {
		return nil, fmt.Errorf("%w order price have to be greater than zero", errInvalidPrice)
	}
	if arg.ClientID == "" {
		arg.ClientID = strconv.FormatInt(dy.Websocket.Conn.GenerateMessageID(true), 10)
	}
	if arg.PostOnly && arg.TimeInForce == "FOK" {
		return nil, errors.New("Order cannot be postOnly and have timeInForce: FOK")
	}
	creds, err := dy.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	orderSignParam := starkex.OrderSignParam{
		NetworkId:  1,
		Market:     arg.Market,
		Side:       arg.Side,
		HumanSize:  strconv.FormatFloat(arg.Size, 'f', -1, 64),
		HumanPrice: strconv.FormatFloat(arg.Price, 'f', -1, 64),
		LimitFee:   strconv.FormatFloat(arg.LimitFee, 'f', -1, 64),
		ClientId:   arg.ClientID,
		Expiration: arg.Expiration,
	}
	signature, err := starkex.OrderSign(creds.SubAccount, orderSignParam)
	if err != nil {
		return nil, err
	}
	arg.Signature = signature
	var resp *Order
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, orders, &arg, &resp)
}

// CancelOrderByID cancel an order by its unique id.
func (dy *DYDX) CancelOrderByID(ctx context.Context, orderID string) (*Order, error) {
	var resp *Order
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, cancelSingleOrderEPL, http.MethodDelete, fmt.Sprintf(orderByID, orderID), nil, &resp)
}

// CancelMultipleOrders either bulk cancel all orders or just all orders for a specific market.
func (dy *DYDX) CancelMultipleOrders(ctx context.Context, market string) ([]Order, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	resp := struct {
		CancelOrders []Order `json:"cancelOrders"`
	}{}
	return resp.CancelOrders, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, cancelOrdersEPL, http.MethodDelete, common.EncodeURLValues(orders, params), nil, &resp)
}

// CancelActiveOrders cancel active orders that match request parameters.
func (dy *DYDX) CancelActiveOrders(ctx context.Context, market, side, id string) ([]Order, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if side != "" {
		params.Set("side", strings.ToUpper(side))
	}
	if id != "" {
		params.Set("id", id)
	}
	resp := struct {
		CancelOrders []Order `json:"cancelOrders"`
	}{}
	return resp.CancelOrders, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, cancelActiveOrdersEPL, http.MethodDelete, common.EncodeURLValues(activeOrders, params), nil, &resp)
}

// GetOrders retrives active (not filled or canceled) orders for a user by specified parameters.
func (dy *DYDX) GetOrders(ctx context.Context, market, status, side, orderType string, limit int64, createdBeforeOrAt time.Time, returnLatestOrders bool) ([]Order, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if status != "" {
		params.Set("status", status)
	}
	if side != "" {
		params.Set("side", strings.ToUpper(side))
	}
	if orderType != "" {
		params.Set("type", strings.ToUpper(orderType))
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !createdBeforeOrAt.IsZero() {
		params.Set("createdBeforeOrAt", createdBeforeOrAt.UTC().Format(timeFormat))
	}
	if returnLatestOrders {
		params.Set("returnLatestOrders", "true")
	}
	resp := &struct {
		Orders []Order `json:"orders"`
	}{}
	return resp.Orders, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, getActiveOrdersEPL, http.MethodGet, common.EncodeURLValues(orders, params), nil, &resp)
}

// GetOpenOrders retrives active (not filled or canceled) orders for a user by specified parameters.
func (dy *DYDX) GetOpenOrders(ctx context.Context, market, side, id string) ([]Order, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if side != "" {
		params.Set("side", side)
	}
	if id != "" {
		params.Set("id", id)
	}
	resp := &struct {
		Orders []Order `json:"orders"`
	}{}
	return resp.Orders, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, getActiveOrdersEPL, http.MethodGet, common.EncodeURLValues(activeOrders, params), nil, &resp)
}

// GetOrderByID represents an order by id from the active orderbook and order history.
func (dy *DYDX) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	resp := struct {
		Order Order `json:"order"`
	}{}
	return &resp.Order, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, fmt.Sprintf(orderByID, id), nil, &resp)
}

// GetOrderByClientID retrives an order by clientId from the active orderbook and order history.
// Only the latest 1 hour of orders can be fetched from this endpoint.
func (dy *DYDX) GetOrderByClientID(ctx context.Context, clientID string) (*Order, error) {
	resp := struct {
		Order Order `json:"order"`
	}{}
	return &resp.Order, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, fmt.Sprintf(orderClientID, clientID), nil, &resp)
}

// GetFills represents fills for a user by specified parameters.
func (dy *DYDX) GetFills(ctx context.Context, market, orderID string, limit int64, createdBeforeOrAt time.Time) ([]OrderFill, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if orderID != "" {
		params.Set("orderId", orderID)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !createdBeforeOrAt.IsZero() {
		params.Set("createdBeforeOrAt", createdBeforeOrAt.UTC().Format(timeFormat))
	}
	resp := &OrderFills{}
	return resp.Fills, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(fills, params), nil, &resp)
}

// GetFundingPayment retrives funding Payments made to an account.
func (dy *DYDX) GetFundingPayment(ctx context.Context, market string, limit int64, effectiveBeforeOrAt time.Time) ([]FundingPayment, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if limit > 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !effectiveBeforeOrAt.IsZero() {
		params.Set("effectiveBeforeOrAt", effectiveBeforeOrAt.UTC().Format(timeFormat))
	}
	var resp *FundingPayments
	return resp.FundingPayments, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(funding, params), nil, &resp)
}

// GetHistoricPNLTicks retrives historical PNL for an account during an interval.
func (dy *DYDX) GetHistoricPNLTicks(ctx context.Context, effectiveBeforeOrAt, effectiveAtOrAfter time.Time) ([]HistoricPNL, error) {
	params := url.Values{}
	if effectiveBeforeOrAt.IsZero() {
		params.Set("effectiveBeforeOrAt", effectiveBeforeOrAt.UTC().Format(timeFormat))
	}
	if effectiveAtOrAfter.IsZero() {
		params.Set("effectiveAtOrAfter", effectiveAtOrAfter.UTC().Format(timeFormat))
	}
	var resp *HistoricPNLResponse
	return resp.HistoricalPNL, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(historicalPNL, params), nil, &resp)
}

// GetTradingRewards retrives the rewards weight of a given epoch.
func (dy *DYDX) GetTradingRewards(ctx context.Context, epoch int64, secondaryAddress string) (*TradingRewards, error) {
	params := url.Values{}
	if epoch != 0 {
		params.Set("epoch", strconv.FormatInt(epoch, 10))
	}
	if secondaryAddress != "" {
		params.Set("secondaryAddress", secondaryAddress)
	}
	var resp *TradingRewards
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(rewardsWeight, params), nil, &resp)
}

// GetLiquidityProviderRewards the liquidity provider rewards of a given epoch (epochs 13+).
func (dy *DYDX) GetLiquidityProviderRewards(ctx context.Context, epoch int64) (*LiquidityProviderRewards, error) {
	params := url.Values{}
	if epoch != 0 {
		params.Set("epoch", strconv.FormatInt(epoch, 10))
	}
	var resp *LiquidityProviderRewards
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(rewardsLiquidityProvider, params), nil, &resp)
}

// GetLiquidityRewards retrives the liquidity rewards of a given epoch.
func (dy *DYDX) GetLiquidityRewards(ctx context.Context, epoch int64, secondaryAddress string) (*LiquidityProviderRewards, error) {
	params := url.Values{}
	if epoch != 0 {
		params.Set("epoch", strconv.FormatInt(epoch, 10))
	}
	if secondaryAddress != "" {
		params.Set("secondaryAddress", secondaryAddress)
	}
	var resp *LiquidityProviderRewards
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, common.EncodeURLValues(liquidityRewards, params), nil, &resp)
}

// GetRetroactiveMiningRewards retrives the retroactive mining rewards of a given epoch.
func (dy *DYDX) GetRetroactiveMiningRewards(ctx context.Context) (*RetroactiveMining, error) {
	var resp *RetroactiveMining
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, rewardsRetroactiveMining, nil, &resp)
}

// SendVerificationEmail send an email to the email address associated with the user, requesting that they click on a link to verify their email address.
func (dy *DYDX) SendVerificationEmail(ctx context.Context) (interface{}, error) {
	var resp interface{}
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, sendVerificationEmailEPL, http.MethodGet, emailsSendVeroficationEmail, nil, &resp)
}

// RequestTestnetTokens requests tokens on dYdX's staging server.
// a fixed number of tokens will be transferred to the account. Please take note of rate limits.
func (dy *DYDX) RequestTestnetTokens(ctx context.Context) (*TestnetToken, error) {
	var resp *TestnetToken
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodPost, testnetTokens, nil, &resp)
}

// GetPrivateProfile retrives private profile data for the user. This is a superset of the /v3/profile/:publicId endpoint.
func (dy *DYDX) GetPrivateProfile(ctx context.Context) (*PrivateProfile, error) {
	var resp *PrivateProfile
	return resp, dy.SendAuthenticatedHTTPRequest(ctx, exchange.RestSpot, defaultV3EPL, http.MethodGet, privateProfile, nil, &resp)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (dy *DYDX) SendAuthenticatedHTTPRequest(ctx context.Context, endpoint exchange.URL, epl request.EndpointLimit, method, path string, data, result interface{}) error {
	urlPath, err := dy.API.Endpoints.GetURL(endpoint)
	if err != nil {
		return err
	}
	var dataString string
	if data == nil {
		dataString = ""
	} else {
		var value []byte
		value, err = json.Marshal(data)
		if err != nil {
			return err
		}
		dataString = string(value)
	}
	var creds *account.Credentials
	creds, err = dy.GetCredentials(ctx)
	if err != nil {
		return err
	}
	newRequest := func() (*request.Item, error) {
		var body io.Reader
		var payload []byte
		if data != nil {
			payload, err = json.Marshal(data)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(payload)
		}
		if err != nil {
			return nil, err
		}
		timestamp := time.Now().UTC().Format(timeFormat)
		message := fmt.Sprintf("%s%s%s%s", timestamp, strings.ToUpper(method), "/"+dydxAPIVersion+path, dataString)
		secret, _ := base64.URLEncoding.DecodeString(creds.Secret)
		h := hmac.New(sha256.New, secret)
		h.Write([]byte(message))
		headers := make(map[string]string)
		headers["DYDX-SIGNATURE"] = base64.URLEncoding.EncodeToString(h.Sum(nil))
		headers["DYDX-PASSPHRASE"] = creds.PEMKey
		headers["DYDX-API-KEY"] = creds.Key
		headers["DYDX-TIMESTAMP"] = timestamp
		headers["Content-Type"] = "application/json"
		return &request.Item{
			Method:        method,
			Path:          urlPath + path,
			Headers:       headers,
			Body:          body,
			Result:        result,
			AuthRequest:   true,
			Verbose:       dy.Verbose,
			HTTPDebugging: dy.HTTPDebugging,
			HTTPRecording: dy.HTTPRecording,
		}, nil
	}
	return dy.SendPayload(ctx, epl, newRequest)
}
