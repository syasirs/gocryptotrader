package bitmex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Bitmex is the overarching type across this package
type Bitmex struct {
	exchange.Base
}

const (
	bitmexAPIVersion    = "v1"
	bitmexAPIURL        = "https://www.bitmex.com/api/v1"
	bitmexAPItestnetURL = "https://testnet.bitmex.com/api/v1"

	// Public endpoints
	bitmexEndpointAnnouncement              = "/announcement"
	bitmexEndpointAnnouncementUrgent        = "/announcement/urgent"
	bitmexEndpointOrderbookL2               = "/orderBook/L2"
	bitmexEndpointTrollbox                  = "/chat"
	bitmexEndpointTrollboxChannels          = "/chat/channels"
	bitmexEndpointTrollboxConnected         = "/chat/connected"
	bitmexEndpointFundingHistory            = "/funding"
	bitmexEndpointInstruments               = "/instrument"
	bitmexEndpointActiveInstruments         = "/instrument/active"
	bitmexEndpointActiveAndIndexInstruments = "/instrument/activeAndIndices"
	bitmexEndpointActiveIntervals           = "/instrument/activeIntervals"
	bitmexEndpointCompositeIndex            = "/instrument/compositeIndex"
	bitmexEndpointIndices                   = "/instrument/indices"
	bitmexEndpointInsuranceHistory          = "/insurance"
	bitmexEndpointLiquidation               = "/liquidation"
	bitmexEndpointLeader                    = "/leaderboard"
	bitmexEndpointAlias                     = "/leaderboard/name"
	bitmexEndpointQuote                     = "/quote"
	bitmexEndpointQuoteBucketed             = "/quote/bucketed"
	bitmexEndpointSettlement                = "/settlement"
	bitmexEndpointStats                     = "/stats"
	bitmexEndpointStatsHistory              = "/stats/history"
	bitmexEndpointStatsSummary              = "/stats/historyUSD"
	bitmexEndpointTrade                     = "/trade"
	bitmexEndpointTradeBucketed             = "/trade/bucketed"

	// Authenticated endpoints
	bitmexEndpointAPIkeys               = "/apiKey"
	bitmexEndpointDisableAPIkey         = "/apiKey/disable"
	bitmexEndpointEnableAPIkey          = "/apiKey/enable"
	bitmexEndpointTrollboxSend          = "/chat"
	bitmexEndpointExecution             = "/execution"
	bitmexEndpointExecutionTradeHistory = "/execution/tradeHistory"
	bitmexEndpointNotifications         = "/notification"
	bitmexEndpointOrder                 = "/order"
	bitmexEndpointCancelAllOrders       = "/order/all"
	bitmexEndpointBulk                  = "/order/bulk"
	bitmexEndpointCancelOrderAfter      = "/order/cancelAllAfter"
	bitmexEndpointClosePosition         = "/order/closePosition"
	bitmexEndpointPosition              = "/position"
	bitmexEndpointIsolatePosition       = "/position/isolate"
	bitmexEndpointLeveragePosition      = "/position/leverage"
	bitmexEndpointAdjustRiskLimit       = "/position/riskLimit"
	bitmexEndpointTransferMargin        = "/position/transferMargin"

	// Rate limits - 150 requests per 5 minutes
	bitmexUnauthRate = 30
	// 300 requests per 5 minutes
	bitmexAuthRate = 40

	// Contract Types
	ContractPerpetual = iota
	ContractFutures
	ContractDownsideProfit
	ContractUpsideProfit
)

// SetDefaults sets the basic defaults for Bitmex
func (b *Bitmex) SetDefaults() {
	b.Name = "Bitmex"
	b.Enabled = false
	b.Verbose = false
	b.Websocket = false
	b.RESTPollingDelay = 10
	b.RequestCurrencyPairFormat.Delimiter = ""
	b.RequestCurrencyPairFormat.Uppercase = true
	b.ConfigCurrencyPairFormat.Delimiter = ""
	b.ConfigCurrencyPairFormat.Uppercase = true
	b.AssetTypes = []string{ticker.Spot}
	b.Requester = request.New(b.Name,
		request.NewRateLimit(time.Second, bitmexAuthRate),
		request.NewRateLimit(time.Second, bitmexUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
}

// Setup takes in the supplied exchange configuration details and sets params
func (b *Bitmex) Setup(exch config.ExchangeConfig) {
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

// GetAnnouncement returns the general announcements from Bitmex
func (b *Bitmex) GetAnnouncement() ([]Announcement, error) {
	var announcement []Announcement

	return announcement, b.SendHTTPRequest(bitmexEndpointAnnouncement,
		nil,
		&announcement)
}

// GetUrgentAnnouncement returns an urgent announcement for your account
func (b *Bitmex) GetUrgentAnnouncement() ([]Announcement, error) {
	var announcement []Announcement

	return announcement, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointAnnouncementUrgent,
		nil,
		&announcement)
}

// GetAPIKeys returns the APIkeys from bitmex
func (b *Bitmex) GetAPIKeys() ([]APIKey, error) {
	var keys []APIKey

	return keys, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointAPIkeys,
		nil,
		&keys)
}

// RemoveAPIKey removes an Apikey from the bitmex trading engine
func (b *Bitmex) RemoveAPIKey(params APIKeyParams) (bool, error) {
	var keyDeleted bool

	return keyDeleted, b.SendAuthenticatedHTTPRequest("DELETE",
		bitmexEndpointAPIkeys,
		params,
		&keyDeleted)
}

// DisableAPIKey disables an Apikey from the bitmex trading engine
func (b *Bitmex) DisableAPIKey(params APIKeyParams) (APIKey, error) {
	var keyInfo APIKey

	return keyInfo, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointDisableAPIkey,
		params,
		&keyInfo)
}

// EnableAPIKey enables an Apikey from the bitmex trading engine
func (b *Bitmex) EnableAPIKey(params APIKeyParams) (APIKey, error) {
	var keyInfo APIKey

	return keyInfo, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointEnableAPIkey,
		params,
		&keyInfo)
}

// GetTrollboxMessages returns messages from the bitmex trollbox
func (b *Bitmex) GetTrollboxMessages(params ChatGetParams) ([]Chat, error) {
	var messages []Chat

	return messages, b.SendHTTPRequest(bitmexEndpointTrollbox, params, &messages)
}

// SendTrollboxMessage sends a message to the bitmex trollbox
func (b *Bitmex) SendTrollboxMessage(params ChatSendParams) ([]Chat, error) {
	var messages []Chat

	return messages, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointTrollboxSend,
		params,
		&messages)
}

// GetTrollboxChannels the channels from the the bitmex trollbox
func (b *Bitmex) GetTrollboxChannels() ([]ChatChannel, error) {
	var channels []ChatChannel

	return channels, b.SendHTTPRequest(bitmexEndpointTrollboxChannels,
		nil,
		&channels)
}

// GetTrollboxConnectedUsers the channels from the the bitmex trollbox
func (b *Bitmex) GetTrollboxConnectedUsers() (ConnectedUsers, error) {
	var users ConnectedUsers

	return users, b.SendHTTPRequest(bitmexEndpointTrollboxConnected, nil, &users)
}

// GetAccountExecutions returns all raw transactions, which includes order
// opening and cancelation, and order status changes. It can be quite noisy.
// More focused information is available at /execution/tradeHistory.
func (b *Bitmex) GetAccountExecutions(params GenericRequestParams) ([]Execution, error) {
	var executionList []Execution

	return executionList, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointExecution,
		params,
		&executionList)
}

// GetAccountExecutionTradeHistory returns all balance-affecting executions.
// This includes each trade, insurance charge, and settlement.
func (b *Bitmex) GetAccountExecutionTradeHistory(params GenericRequestParams) ([]Execution, error) {
	var tradeHistory []Execution

	return tradeHistory, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointExecutionTradeHistory,
		params,
		&tradeHistory)
}

// GetFundingHistory returns funding history
func (b *Bitmex) GetFundingHistory() ([]Funding, error) {
	var fundingHistory []Funding

	return fundingHistory, b.SendHTTPRequest(bitmexEndpointFundingHistory,
		nil,
		&fundingHistory)
}

// GetInstruments returns instrument data
func (b *Bitmex) GetInstruments(params GenericRequestParams) ([]Instrument, error) {
	var instruments []Instrument

	return instruments, b.SendHTTPRequest(bitmexEndpointInstruments,
		params,
		&instruments)
}

// GetActiveInstruments returns active instruments
func (b *Bitmex) GetActiveInstruments(params GenericRequestParams) ([]Instrument, error) {
	var activeInstruments []Instrument

	return activeInstruments, b.SendHTTPRequest(bitmexEndpointActiveInstruments,
		params,
		&activeInstruments)
}

// GetActiveAndIndexInstruments returns all active instruments and all indices
func (b *Bitmex) GetActiveAndIndexInstruments() ([]Instrument, error) {
	var activeAndIndices []Instrument

	return activeAndIndices,
		b.SendHTTPRequest(bitmexEndpointActiveAndIndexInstruments,
			nil,
			&activeAndIndices)
}

// GetActiveIntervals returns funding history
func (b *Bitmex) GetActiveIntervals() (InstrumentInterval, error) {
	var interval InstrumentInterval

	return interval, b.SendHTTPRequest(bitmexEndpointActiveIntervals,
		nil,
		&interval)
}

// GetCompositeIndex returns composite index
func (b *Bitmex) GetCompositeIndex(params GenericRequestParams) ([]IndexComposite, error) {
	var compositeIndices []IndexComposite

	return compositeIndices, b.SendHTTPRequest(bitmexEndpointCompositeIndex,
		params,
		&compositeIndices)
}

// GetIndices returns all price indices
func (b *Bitmex) GetIndices() ([]Instrument, error) {
	var indices []Instrument

	return indices, b.SendHTTPRequest(bitmexEndpointIndices, nil, &indices)
}

// GetInsuranceFundHistory returns insurance fund history
func (b *Bitmex) GetInsuranceFundHistory(params GenericRequestParams) ([]Insurance, error) {
	var history []Insurance

	return history, b.SendHTTPRequest(bitmexEndpointIndices, params, &history)
}

// GetLeaderboard returns leaderboard information
func (b *Bitmex) GetLeaderboard(params LeaderboardGetParams) ([]Leaderboard, error) {
	var leader []Leaderboard

	return leader, b.SendHTTPRequest(bitmexEndpointLeader, params, &leader)
}

// GetAliasOnLeaderboard returns your alias on the leaderboard
func (b *Bitmex) GetAliasOnLeaderboard() (Alias, error) {
	var alias Alias

	return alias, b.SendHTTPRequest(bitmexEndpointAlias, nil, &alias)
}

// GetLiquidationOrders returns liquidation orders
func (b *Bitmex) GetLiquidationOrders(params GenericRequestParams) ([]Liquidation, error) {
	var orders []Liquidation

	return orders, b.SendHTTPRequest(bitmexEndpointLiquidation,
		params,
		&orders)
}

// GetCurrentNotifications returns your current notifications
func (b *Bitmex) GetCurrentNotifications() ([]Notification, error) {
	var notifications []Notification

	return notifications, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointNotifications,
		nil,
		&notifications)
}

// GetOrders returns all the orders, open and closed
func (b *Bitmex) GetOrders(params GenericRequestParams) ([]Order, error) {
	var orders []Order

	return orders, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointOrder,
		params,
		&orders)
}

// AmendOrder amends the quantity or price of an open order
func (b *Bitmex) AmendOrder(params OrderAmendParams) ([]Order, error) {
	var orders []Order

	return orders, b.SendAuthenticatedHTTPRequest("PUT",
		bitmexEndpointOrder,
		params,
		&orders)
}

// CreateOrder creates a new order
func (b *Bitmex) CreateOrder(params OrderNewParams) (Order, error) {
	var orderInfo Order

	return orderInfo, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointOrder,
		params,
		&orderInfo)
}

// CancelOrders cancels one or a batch of orders on the exchange and returns
// a cancelled order list
func (b *Bitmex) CancelOrders(params OrderCancelParams) ([]Order, error) {
	var cancelledOrders []Order

	return cancelledOrders, b.SendAuthenticatedHTTPRequest("DELETE",
		bitmexEndpointOrder,
		params,
		&cancelledOrders)
}

// CancelAllOrders cancels all open orders on the exchange
func (b *Bitmex) CancelAllOrders(params OrderCancelAllParams) ([]Order, error) {
	var cancelledOrders []Order

	return cancelledOrders, b.SendAuthenticatedHTTPRequest("DELETE",
		bitmexEndpointCancelAllOrders,
		params,
		&cancelledOrders)
}

// AmendBulkOrders amends multiple orders for the same symbol
func (b *Bitmex) AmendBulkOrders(params OrderAmendBulkParams) ([]Order, error) {
	var amendedOrders []Order

	return amendedOrders, b.SendAuthenticatedHTTPRequest("PUT",
		bitmexEndpointBulk,
		params,
		&amendedOrders)
}

// CreateBulkOrders creates multiple orders for the same symbol
func (b *Bitmex) CreateBulkOrders(params OrderNewBulkParams) ([]Order, error) {
	var orders []Order

	return orders, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointBulk,
		params,
		&orders)
}

// CancelAllOrdersAfterTime closes all positions after a certain time period
func (b *Bitmex) CancelAllOrdersAfterTime(params OrderCancelAllAfterParams) ([]Order, error) {
	var cancelledOrder []Order

	return cancelledOrder, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointCancelOrderAfter,
		params,
		&cancelledOrder)
}

// ClosePosition closes a position WARNING deprecated use /order endpoint
func (b *Bitmex) ClosePosition(params OrderClosePositionParams) ([]Order, error) {
	var closedPositions []Order

	return closedPositions, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointOrder,
		params,
		&closedPositions)
}

// GetOrderbook returns layer two orderbook data
func (b *Bitmex) GetOrderbook(params OrderBookGetL2Params) ([]OrderBookL2, error) {
	var orderBooks []OrderBookL2

	return orderBooks, b.SendHTTPRequest(bitmexEndpointOrderbookL2,
		params,
		&orderBooks)
}

// GetPositions returns positions
func (b *Bitmex) GetPositions(params PositionGetParams) ([]Position, error) {
	var positions []Position

	return positions, b.SendAuthenticatedHTTPRequest("GET",
		bitmexEndpointPosition,
		params,
		&positions)
}

// IsolatePosition enables isolated margin or cross margin per-position
func (b *Bitmex) IsolatePosition(params PositionIsolateMarginParams) (Position, error) {
	var position Position

	return position, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointIsolatePosition,
		params,
		&position)
}

// LeveragePosition chooses leverage for a position
func (b *Bitmex) LeveragePosition(params PositionUpdateLeverageParams) (Position, error) {
	var position Position

	return position, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointLeveragePosition,
		params,
		&position)
}

// UpdateRiskLimit updates risk limit on a position
func (b *Bitmex) UpdateRiskLimit(params PositionUpdateRiskLimitParams) (Position, error) {
	var position Position

	return position, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointAdjustRiskLimit,
		params,
		&position)
}

// TransferMargin transfers equity in or out of a position
func (b *Bitmex) TransferMargin(params PositionTransferIsolatedMarginParams) (Position, error) {
	var position Position

	return position, b.SendAuthenticatedHTTPRequest("POST",
		bitmexEndpointTransferMargin,
		params,
		&position)
}

// GetQuotes returns quotations
func (b *Bitmex) GetQuotes(params GenericRequestParams) ([]Quote, error) {
	var quotations []Quote

	return quotations, b.SendHTTPRequest(bitmexEndpointQuote,
		params,
		&quotations)
}

// GetQuotesByBuckets returns previous quotes in time buckets
func (b *Bitmex) GetQuotesByBuckets(params QuoteGetBucketedParams) ([]Quote, error) {
	var quotations []Quote

	return quotations, b.SendHTTPRequest(bitmexEndpointQuoteBucketed,
		params,
		&quotations)
}

// GetSettlementHistory returns settlement history
func (b *Bitmex) GetSettlementHistory(params GenericRequestParams) ([]Settlement, error) {
	var history []Settlement

	return history, b.SendHTTPRequest(bitmexEndpointSettlement,
		params,
		&history)
}

// GetStats returns exchange wide per series turnover and volume statistics
func (b *Bitmex) GetStats() ([]Stats, error) {
	var stats []Stats

	return stats, b.SendHTTPRequest(bitmexEndpointStats, nil, &stats)
}

// GetStatsHistorical historic stats
func (b *Bitmex) GetStatsHistorical() ([]StatsHistory, error) {
	var history []StatsHistory

	return history, b.SendHTTPRequest(bitmexEndpointStatsHistory, nil, &history)
}

// GetStatSummary returns the stats summary in USD terms
func (b *Bitmex) GetStatSummary() ([]StatsUSD, error) {
	var summary []StatsUSD

	return summary, b.SendHTTPRequest(bitmexEndpointStatsSummary, nil, &summary)
}

// GetTrade returns executed trades on the desk
func (b *Bitmex) GetTrade(params GenericRequestParams) ([]Trade, error) {
	var trade []Trade

	return trade, b.SendHTTPRequest(bitmexEndpointTrade, params, &trade)
}

// GetPreviousTrades previous trade history in time buckets
func (b *Bitmex) GetPreviousTrades(params TradeGetBucketedParams) ([]Trade, error) {
	var trade []Trade

	return trade, b.SendHTTPRequest(bitmexEndpointTradeBucketed,
		params,
		&trade)
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (b *Bitmex) SendHTTPRequest(path string, params Parameter, result interface{}) error {
	var respCheck interface{}
	path = bitmexAPIURL + path
	if params != nil {
		if !params.IsNil() {
			encodedPath, err := params.ToURLVals(path)
			if err != nil {
				return err
			}
			err = b.SendPayload("GET", encodedPath, nil, nil, &respCheck, false, b.Verbose)
			if err != nil {
				return err
			}
			return b.CaptureError(respCheck, result)
		}
	}
	err := b.SendPayload("GET", path, nil, nil, &respCheck, false, b.Verbose)
	if err != nil {
		return err
	}
	return b.CaptureError(respCheck, result)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request to bitmex
func (b *Bitmex) SendAuthenticatedHTTPRequest(verb, path string, params Parameter, result interface{}) error {
	if !b.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet,
			b.Name)
	}

	timestamp := time.Now().Add(time.Second * 10).UnixNano()
	timestampStr := strconv.FormatInt(timestamp, 10)
	timestampNew := timestampStr[:13]

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["api-expires"] = timestampNew
	headers["api-key"] = b.APIKey

	var payload string
	if params != nil {
		err := params.VerifyData()
		if err != nil {
			return err
		}
		data, err := common.JSONEncode(params)
		if err != nil {
			return err
		}
		payload = string(data)
	}

	hmac := common.GetHMAC(common.HashSHA256,
		[]byte(verb+"/api/v1"+path+timestampNew+payload),
		[]byte(b.APISecret))

	headers["api-signature"] = common.HexEncodeToString(hmac)

	var respCheck interface{}

	err := b.SendPayload(verb,
		bitmexAPIURL+path,
		headers,
		bytes.NewBuffer([]byte(payload)),
		&respCheck,
		true,
		b.Verbose)
	if err != nil {
		return err
	}

	return b.CaptureError(respCheck, result)
}

// CaptureError little hack that captures an error
func (b *Bitmex) CaptureError(resp, reType interface{}) error {
	var Error RequestError

	marshalled, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	err = common.JSONDecode(marshalled, &Error)
	if err == nil {
		return fmt.Errorf("bitmex error %s: %s",
			Error.Error.Name,
			Error.Error.Message)
	}

	err = common.JSONDecode(marshalled, reType)
	if err != nil {
		return err
	}
	return nil
}
