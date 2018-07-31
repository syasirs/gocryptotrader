package main

import (
	"fmt"
	"time"

	"github.com/thrasher-/gocryptotrader/communications"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/huobi"
	"github.com/thrasher-/gocryptotrader/portfolio"
)

// Bot contains configuration, portfolio, exchange & ticker data and is the
// overarching type across this code base.
type Bot struct {
	config     *config.Config
	portfolio  *portfolio.Base
	exchanges  []exchange.IBotExchange
	comms      *communications.Communications
	shutdown   chan bool
	dryRun     bool
	configFile string
}

const banner = `
   ______        ______                     __        ______                  __
  / ____/____   / ____/_____ __  __ ____   / /_ ____ /_  __/_____ ______ ____/ /___   _____
 / / __ / __ \ / /    / ___// / / // __ \ / __// __ \ / /  / ___// __  // __  // _ \ / ___/
/ /_/ // /_/ // /___ / /   / /_/ // /_/ // /_ / /_/ // /  / /   / /_/ // /_/ //  __// /
\____/ \____/ \____//_/    \__, // .___/ \__/ \____//_/  /_/    \__,_/ \__,_/ \___//_/
                          /____//_/
`

var bot Bot

// getDefaultConfig 获取默认配置
func getDefaultConfig() config.ExchangeConfig {
	return config.ExchangeConfig{
		Name:                    "huobi",
		Enabled:                 true,
		Verbose:                 true,
		Websocket:               false,
		BaseAsset:               "btc",
		QuoteAsset:              "usdt",
		RESTPollingDelay:        10,
		HTTPTimeout:             15000000000,
		AuthenticatedAPISupport: true,
		APIKey:                  "",
		APISecret:               "",
	}
}

func main() {
	fmt.Println(time.Now())
	// exchange := gateio.Gateio{}
	// exchange := bitfinex.Bitfinex{}
	// exchange := okex.OKEX{}
	exchange := huobi.HUOBI{}
	defaultConfig := getDefaultConfig()
	exchange.SetDefaults()
	fmt.Println("----------setup-------")
	exchange.Setup(defaultConfig)

	arg := huobi.SpotNewOrderRequestParams{
		Symbol:    exchange.GetSymbol(),
		AccountID: 3838465,
		Amount:    0.01,
		Price:     10.1,
		Type:      huobi.SpotNewOrderRequestTypeBuyLimit,
	}
	fmt.Println(exchange.SpotNewOrder(arg))

	// res, err := exchange.SpotNewOrder(okex.SpotNewOrderRequestParams{
	// 	Symbol: exchange.GetSymbol(),
	// 	Amount: 1.1,
	// 	Price:  10.1,
	// 	Type:   okex.SpotNewOrderRequestTypeBuy,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(res)
	// }

	// fmt.Println(exchange.GetKline("btcusdt", "1min", ""))
	// fmt.Println(exchange.GetKline("btcusdt", "1min", ""))
	// fmt.Println(exchange.GetKline("btcusdt", "15min", ""))
	// fmt.Println(exchange.GetKline("btcusdt", "1hour", ""))
	// fmt.Println(exchange.GetKline("btcusdt", "1day", ""))

	// list, err := exchange.GetAccountInfo()
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	for k, v := range list {
	// 		// b, _ := json.Marshal(v)
	// 		fmt.Printf("%s:%v \n", k, v)
	// 	}
	// }

	// fmt.Println(exchange.CancelOrder(917591554, exchange.GetSymbol()))

	//获取交易所的规则和交易对信息
	// getExchangeInfo(exchange)

	//获取交易深度
	// getOrderBook(exchange)

	//获取最近的交易记录
	// getRecentTrades(exchange)

	//获取 k 线数据
	// getCandleStickData(exchange)

	//获取最新价格
	// getLatestSpotPrice(exchange)

	//新订单
	// newOrder(exchange)

	//取消订单
	// cancelOrder(exchange, 82584683)

	// fmt.Println(exchange.GetAccount())

	// fmt.Println(exchange.GetSymbol())

}
