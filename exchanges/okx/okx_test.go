package okx

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
)

// Please supply your own keys here to do authenticated endpoint testing
const (
	apiKey                  = ""
	apiSecret               = ""
	passphrase              = ""
	canManipulateRealOrders = false
)

var ok Okx

func TestMain(m *testing.M) {
	ok.SetDefaults()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal(err)
	}
	exchCfg, err := cfg.GetExchangeConfig("okx")
	if err != nil {
		log.Fatal(err)
	}
	exchCfg.API.AuthenticatedSupport = true
	exchCfg.API.AuthenticatedWebsocketSupport = true
	exchCfg.API.Credentials.Key = apiKey
	exchCfg.API.Credentials.Secret = apiSecret
	exchCfg.API.Credentials.ClientID = passphrase

	err = ok.Setup(exchCfg)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func areTestAPIKeysSet() bool {
	return ok.ValidateAPICredentials(ok.GetDefaultCredentials()) == nil
}

func TestGetTickers(t *testing.T) {
	t.Parallel()
	_, er := ok.GetTickers(context.Background(), "SPOT", "", "BTC-USD-SWAP")
	if er != nil {
		t.Error("Okx GetTickers() error", er)
	}
}

func TestGetIndexTickers(t *testing.T) {
	t.Parallel()
	_, er := ok.GetIndexTickers(context.Background(), "USDT", "")
	if er != nil {
		t.Error("OKX GetIndexTickers() error", er)
	}
}

func TestGetOrderBookDepth(t *testing.T) {
	t.Parallel()
	_, er := ok.GetOrderBookDepth(context.Background(), currency.NewPair(currency.BTC, currency.USDT), 10)
	if er != nil {
		t.Error("OKX GetOrderBookDepth() error", er)
	}
}

func TestGetCandlesticks(t *testing.T) {
	t.Parallel()
	_, er := ok.GetCandlesticks(context.Background(), "BTC-USDT", kline.OneHour, time.Unix(time.Now().Unix()-3600, 0), time.Now(), 30)
	if er != nil {
		t.Error("Okx GetCandlesticks() error", er)
	}
}

func TestGetCandlesticksHistory(t *testing.T) {
	t.Parallel()
	_, er := ok.GetCandlesticksHistory(context.Background(), "BTC-USDT", kline.OneHour, time.Unix(time.Now().Unix()-3600, 0), time.Now(), 30)
	if er != nil {
		t.Error("Okx GetCandlesticksHistory() error", er)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()
	_, er := ok.GetTrades(context.Background(), "BTC-USDT", 30)
	if er != nil {
		t.Error("Okx GetTrades() error", er)
	}
}

func TestGet24HTotalVolume(t *testing.T) {
	t.Parallel()
	_, er := ok.Get24HTotalVolume(context.Background())
	if er != nil {
		t.Error("Okx Get24HTotalVolume() error", er)
	}
}

func TestGetOracle(t *testing.T) {
	t.Parallel()
	_, er := ok.GetOracle(context.Background())
	if er != nil {
		t.Error("Okx GetOracle() error", er)
	}
}

func TestGetExchangeRate(t *testing.T) {
	t.Parallel()
	_, er := ok.GetExchangeRate(context.Background())
	if er != nil {
		t.Error("Okx GetExchangeRate() error", er)
	}
}

func TestGetIndexComponents(t *testing.T) {
	t.Parallel()
	_, er := ok.GetIndexComponents(context.Background(), currency.NewPair(currency.BTC, currency.USDT))
	if er != nil {
		t.Error("Okx GetIndexComponents() error", er)
	}
}

func TestGetinstrument(t *testing.T) {
	t.Parallel()
	_, er := ok.GetInstruments(context.Background(), &InstrumentsFetchParams{
		InstrumentType: "MARGIN",
	})
	if er != nil {
		t.Error("Okx GetInstruments() error", er)
	}
}

// TODO: this handler function has  to be amended
var deliveryHistoryData = `{
    "code":"0",
    "msg":"",
    "data":[
        {
            "ts":"1597026383085",
            "details":[
                {
                    "type":"delivery",
                    "instId":"ZIL-BTC",
                    "px":"0.016"
                }
            ]
        },
        {
            "ts":"1597026383085",
            "details":[
                {
                    "instId":"BTC-USD-200529-6000-C",
                    "type":"exercised",
                    "px":"0.016"
                },
                {
                    "instId":"BTC-USD-200529-8000-C",
                    "type":"exercised",
                    "px":"0.016"
                }
            ]
        }
    ]
}`

func TestGetDeliveryHistory(t *testing.T) {
	t.Parallel()
	t.Skip()
	var repo DeliveryHistoryResponse
	if err := json.Unmarshal([]byte(deliveryHistoryData), &repo); err != nil {
		t.Error("Okx error", err)
	}
	_, er := ok.GetDeliveryHistory(context.Background(), "FUTURES", "FUTURES", time.Time{}, time.Time{}, 100)
	if er != nil {
		t.Error("okx GetDeliveryHistory() error", er)
	}
}

func TestGetOpenInterest(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetOpenInterest(context.Background(), "FUTURES", "BTC-USDT", ""); er != nil {
		t.Error("Okx GetOpenInterest() error", er)
	}
}

func TestGetFundingRate(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetFundingRate(context.Background(), "BTC-USD-SWAP"); er != nil {
		t.Error("okx GetFundingRate() error", er)
	}
}

func TestGetFundingRateHistory(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetFundingRateHistory(context.Background(), "BTC-USD-SWAP", time.Time{}, time.Time{}, 10); er != nil {
		t.Error("Okx GetFundingRateHistory() error", er)
	}
}

func TestGetLimitPrice(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetLimitPrice(context.Background(), "BTC-USD-SWAP"); er != nil {
		t.Error("okx GetLimitPrice() error", er)
	}
}

func TestGetOptionMarketData(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetOptionMarketData(context.Background(), "BTC-USD", time.Time{}); er != nil {
		t.Error("Okx GetOptionMarketData() error", er)
	}
}

var estimatedDeliveryResponseString = `{
    "code":"0",
    "msg":"",
    "data":[
    {
        "instType":"FUTURES",
        "instId":"BTC-USDT-201227",
        "settlePx":"200",
        "ts":"1597026383085"
    }
  ]
}`

func TestGetEstimatedDeliveryPrice(t *testing.T) {
	t.Parallel()
	var result DeliveryEstimatedPriceResponse
	er := json.Unmarshal([]byte(estimatedDeliveryResponseString), (&result))
	if er != nil {
		t.Error("Okx GetEstimatedDeliveryPrice() error", er)
	}
	if _, er := ok.GetEstimatedDeliveryPrice(context.Background(), "BTC-USD"); er != nil && !(strings.Contains(er.Error(), "Instrument ID does not exist.")) {
		t.Error("Okx GetEstimatedDeliveryPrice() error", er)
	}
}

func TestGetDiscountRateAndInterestFreeQuota(t *testing.T) {
	t.Parallel()
	_, er := ok.GetDiscountRateAndInterestFreeQuota(context.Background(), "BTC", 0)
	if er != nil {
		t.Error("Okx GetDiscountRateAndInterestFreeQuota() error", er)
	}
}

func TestGetSystemTime(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetSystemTime(context.Background()); er != nil {
		t.Error("Okx GetSystemTime() error", er)
	}
}

func TestGetLiquidationOrders(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetLiquidationOrders(context.Background(), &LiquidationOrderRequestParams{
		InstrumentType: "MARGIN",
		Underlying:     "BTC-USD",
		Currency:       currency.BTC,
	}); er != nil {
		t.Error("Okx GetLiquidationOrders() error", er)
	}
}

func TestGetMarkPrice(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetMarkPrice(context.Background(), "MARGIN", "", ""); er != nil {
		t.Error("Okx GetMarkPrice() error", er)
	}
}

func TestGetPositionTiers(t *testing.T) {
	t.Parallel()
	t.Skip()
	if _, er := ok.GetPositionTiers(context.Background(), "FUTURES", "cross", "BTC-USDT", "", ""); er != nil {
		t.Error("Okx GetPositionTiers() error", er)
	}
}

func TestGetInterestRateAndLoanQuota(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetInterestRateAndLoanQuota(context.Background()); er != nil {
		t.Error("Okx GetInterestRateAndLoanQuota() error", er)
	}
}

func TestGetInterestRateAndLoanQuotaForVIPLoans(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetInterestRateAndLoanQuotaForVIPLoans(context.Background()); er != nil {
		t.Error("Okx GetInterestRateAndLoanQuotaForVIPLoans() error", er)
	}
}

func TestGetPublicUnderlyings(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetPublicUnderlyings(context.Background(), "swap"); er != nil {
		t.Error("Okx GetPublicUnderlyings() error", er)
	}
}

func TestGetInsuranceFundInformations(t *testing.T) {
	t.Parallel()
	// getting the Underlyings usig the Get public Underlyinggs method for specific instrument type.
	var underlyings []string
	var er error
	if underlyings, er = ok.GetPublicUnderlyings(context.Background(), "futures"); er != nil {
		t.Error("Okx GetPublicUnderlyings() error", er)
		t.SkipNow()
	}
	if _, er := ok.GetInsuranceFundInformations(context.Background(), InsuranceFundInformationRequestParams{
		InstrumentType: "FUTURES",
		Underlying:     underlyings[0],
	}); er != nil {
		t.Error("Okx GetInsuranceFundInformations() error", er)
	}
}

// Trading related enndpoints test functions.

func TestGetSupportCoins(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetSupportCoins(context.Background()); er != nil {
		t.Error("Okx GetSupportCoins() error", er)
	}
}

func TestGetTakerVolume(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetTakerVolume(context.Background(), "BTC", "SPOT", time.Time{}, time.Time{}, kline.OneDay); er != nil {
		t.Error("Okx GetTakerVolume() error", er)
	}
}
func TestGetMarginLendingRatio(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetMarginLendingRatio(context.Background(), "BTC", time.Time{}, time.Time{}, kline.OneDay); er != nil {
		t.Error("Okx GetMarginLendingRatio() error", er)
	}
}

func TestGetLongShortRatio(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetLongShortRatio(context.Background(), "BTC", time.Time{}, time.Time{}, kline.OneDay); er != nil {
		t.Error("Okx GetLongShortRatio() error", er)
	}
}

func TestGetContractsOpenInterestAndVolume(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetContractsOpenInterestAndVolume(context.Background(), "BTC", time.Time{}, time.Time{}, kline.OneDay); er != nil {
		t.Error("Okx GetContractsOpenInterestAndVolume() error", er)
	}
}

func TestGetOptionsOpenInterestAndVolume(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetOptionsOpenInterestAndVolume(context.Background(), "BTC", kline.OneDay); er != nil {
		t.Error("Okx GetOptionsOpenInterestAndVolume() error", er)
	}
}

func TestGetPutCallRatio(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetPutCallRatio(context.Background(), "BTC", kline.OneDay); er != nil {
		t.Error("Okx GetPutCallRatio() error", er)
	}
}

func TestGetOpenInterestAndVolumeExpiry(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetOpenInterestAndVolumeExpiry(context.Background(), "BTC", kline.OneDay); er != nil {
		t.Error("Okx GetOpenInterestAndVolume() error", er)
	}
}

func TestGetOpenInterestAndVolumeStrike(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetOpenInterestAndVolumeStrike(context.Background(), "BTC", time.Now(), kline.OneDay); er != nil {
		t.Error("Okx GetOpenInterestAndVolumeStrike() error", er)
	}
}

func TestGetTakerFlow(t *testing.T) {
	t.Parallel()
	if _, er := ok.GetTakerFlow(context.Background(), "BTC", kline.OneDay); er != nil {
		t.Error("Okx GetTakerFlow() error", er)
	}
}

func TestPlaceOrder(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if resp, er := ok.PlaceOrder(context.Background(), PlaceOrderRequestParam{
		InstrumentID:        "REN-USDT",
		TradeMode:           "cross",
		Side:                "sell",
		OrderType:           "limit",
		QuantityToBuyOrSell: 1,
		OrderPrice:          1,
	}); er != nil {
		t.Error("Okx PlaceOrder() error", er)
	} else {
		println(string(MarshalThis(resp)))
	}
}

func MarshalThis(vv interface{}) []byte {
	v, _ := json.Marshal(vv)
	return v
}

func TestPlaceMultipleOrders(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if _, er := ok.PlaceMultipleOrders(context.Background(),
		[]PlaceOrderRequestParam{
			{
				InstrumentID:        "FIL-BTC",
				TradeMode:           "cross",
				Side:                "sell",
				OrderType:           "limit",
				QuantityToBuyOrSell: 1,
				OrderPrice:          1,
			},
		}); er != nil {
		t.Error("Okx PlaceOrderRequestParam() error", er)
	}
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if _, er := ok.CancelOrder(context.Background(),
		CancelOrderRequestParam{
			InstrumentID: "BTC-USD-190927",
			OrderID:      "2510789768709120",
		}); er != nil {
		t.Error("Okx CancelOrder() error", er)
	}
}

func TestCancelMultipleOrders(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if _, er := ok.CancelMultipleOrders(context.Background(), []CancelOrderRequestParam{{
		InstrumentID: "DCR-BTC",
		OrderID:      "2510789768709120",
	}}); er != nil {
		t.Error("Okx CancelMultipleOrders() error", er)
	}
}

func TestAmendOrder(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if _, er := ok.AmendOrder(context.Background(), &AmendOrderRequestParams{
		InstrumentID: "DCR-BTC",
		OrderID:      "2510789768709120",
		NewPrice:     1233324.332,
	}); er != nil {
		t.Error("Okx AmendOrder() error", er)
	}
}
func TestAmendMultipleOrders(t *testing.T) {
	t.Parallel()
	if !areTestAPIKeysSet() {
		t.SkipNow()
	}
	if responses, er := ok.AmendMultipleOrders(context.Background(), []AmendOrderRequestParams{{
		InstrumentID: "DCR-BTC",
		OrderID:      "2510789768709120",
		NewPrice:     1233324.332,
	}}); er != nil {
		t.Error("Okx AmendMultipleOrders() error", er)
	} else {
		println(string(MarshalThis(responses)))
	}
}
