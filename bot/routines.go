package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/stats"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

func (b *Bot) printCurrencyFormat(price float64) string {
	displaySymbol, err := symbol.GetSymbolByCurrencyName(b.Config.FiatDisplayCurrency)
	if err != nil {
		log.Printf("Failed to get display symbol: %s", err)
	}

	return fmt.Sprintf("%s%.8f", displaySymbol, price)
}

func (b *Bot) printConvertCurrencyFormat(origCurrency string, origPrice float64) string {
	displayCurrency := b.Config.FiatDisplayCurrency
	conv, err := currency.ConvertCurrency(origPrice, origCurrency, displayCurrency)
	if err != nil {
		log.Printf("Failed to convert currency: %s", err)
	}

	displaySymbol, err := symbol.GetSymbolByCurrencyName(displayCurrency)
	if err != nil {
		log.Printf("Failed to get display symbol: %s", err)
	}

	origSymbol, err := symbol.GetSymbolByCurrencyName(origCurrency)
	if err != nil {
		log.Printf("Failed to get original currency symbol: %s", err)
	}

	return fmt.Sprintf("%s%.2f %s (%s%.2f %s)",
		displaySymbol,
		conv,
		displayCurrency,
		origSymbol,
		origPrice,
		origCurrency,
	)
}

func (b *Bot) printSummary(result ticker.Price, p pair.CurrencyPair, assetType, exchangeName string, err error) {
	if err != nil {
		log.Printf("Failed to get %s %s ticker. Error: %s",
			p.Pair().String(),
			exchangeName,
			err)
		return
	}

	stats.Add(exchangeName, p, assetType, result.Last, result.Volume)
	if currency.IsFiatCurrency(p.SecondCurrency.String()) && p.SecondCurrency.String() != b.Config.FiatDisplayCurrency {
		origCurrency := p.SecondCurrency.Upper().String()
		log.Printf("%s %s %s: Last %s Ask %s Bid %s High %s Low %s Volume %.8f",
			exchangeName,
			exchange.FormatCurrency(p).String(),
			assetType,
			b.printConvertCurrencyFormat(origCurrency, result.Last),
			b.printConvertCurrencyFormat(origCurrency, result.Ask),
			b.printConvertCurrencyFormat(origCurrency, result.Bid),
			b.printConvertCurrencyFormat(origCurrency, result.High),
			b.printConvertCurrencyFormat(origCurrency, result.Low),
			result.Volume)
	} else {
		if currency.IsFiatCurrency(p.SecondCurrency.String()) && p.SecondCurrency.Upper().String() == b.Config.FiatDisplayCurrency {
			log.Printf("%s %s %s: Last %s Ask %s Bid %s High %s Low %s Volume %.8f",
				exchangeName,
				exchange.FormatCurrency(p).String(),
				assetType,
				b.printCurrencyFormat(result.Last),
				b.printCurrencyFormat(result.Ask),
				b.printCurrencyFormat(result.Bid),
				b.printCurrencyFormat(result.High),
				b.printCurrencyFormat(result.Low),
				result.Volume)
		} else {
			log.Printf("%s %s %s: Last %.8f Ask %.8f Bid %.8f High %.8f Low %.8f Volume %.8f",
				exchangeName,
				exchange.FormatCurrency(p).String(),
				assetType,
				result.Last,
				result.Ask,
				result.Bid,
				result.High,
				result.Low,
				result.Volume)
		}
	}
}

func (b *Bot) printOrderbookSummary(result orderbook.Base, p pair.CurrencyPair, assetType, exchangeName string, err error) {
	if err != nil {
		log.Printf("Failed to get %s %s orderbook. Error: %s",
			p.Pair().String(),
			exchangeName,
			err)
		return
	}
	bidsAmount, bidsValue := result.CalculateTotalBids()
	asksAmount, asksValue := result.CalculateTotalAsks()

	if currency.IsFiatCurrency(p.SecondCurrency.String()) && p.SecondCurrency.String() != b.Config.FiatDisplayCurrency {
		origCurrency := p.SecondCurrency.Upper().String()
		log.Printf("%s %s %s: Orderbook Bids len: %d Amount: %f %s. Total value: %s Asks len: %d Amount: %f %s. Total value: %s",
			exchangeName,
			exchange.FormatCurrency(p).String(),
			assetType,
			len(result.Bids),
			bidsAmount,
			p.FirstCurrency.String(),
			b.printConvertCurrencyFormat(origCurrency, bidsValue),
			len(result.Asks),
			asksAmount,
			p.FirstCurrency.String(),
			b.printConvertCurrencyFormat(origCurrency, asksValue),
		)
	} else {
		if currency.IsFiatCurrency(p.SecondCurrency.String()) && p.SecondCurrency.Upper().String() == b.Config.FiatDisplayCurrency {
			log.Printf("%s %s %s: Orderbook Bids len: %d Amount: %f %s. Total value: %s Asks len: %d Amount: %f %s. Total value: %s",
				exchangeName,
				exchange.FormatCurrency(p).String(),
				assetType,
				len(result.Bids),
				bidsAmount,
				p.FirstCurrency.String(),
				b.printCurrencyFormat(bidsValue),
				len(result.Asks),
				asksAmount,
				p.FirstCurrency.String(),
				b.printCurrencyFormat(asksValue),
			)
		} else {
			log.Printf("%s %s %s: Orderbook Bids len: %d Amount: %f %s. Total value: %f Asks len: %d Amount: %f %s. Total value: %f",
				exchangeName,
				exchange.FormatCurrency(p).String(),
				assetType,
				len(result.Bids),
				bidsAmount,
				p.FirstCurrency.String(),
				bidsValue,
				len(result.Asks),
				asksAmount,
				p.FirstCurrency.String(),
				asksValue,
			)
		}
	}

}

func relayWebsocketEvent(result interface{}, event, assetType, exchangeName string) {
	evt := WebsocketEvent{
		Data:      result,
		Event:     event,
		AssetType: assetType,
		Exchange:  exchangeName,
	}
	err := BroadcastWebsocketMessage(evt)
	if err != nil {
		log.Println(fmt.Errorf("Failed to broadcast websocket event. Error: %s",
			err))
	}
}

// TickerUpdaterRoutine updates ticker information in a segregated routine
func (b *Bot) TickerUpdaterRoutine() {
	log.Println("Starting ticker updater routine")
	for {
		for x := range b.Exchanges {
			if b.Exchanges[x].IsEnabled() {
				exchangeName := b.Exchanges[x].GetName()
				enabledCurrencies := b.Exchanges[x].GetEnabledCurrencies()

				var result ticker.Price
				var err error
				var assetTypes []string

				assetTypes, err = exchange.GetExchangeAssetTypes(exchangeName)
				if err != nil {
					log.Printf("failed to get %s exchange asset types. Error: %s",
						exchangeName, err)
				}

				for y := range enabledCurrencies {
					currency := enabledCurrencies[y]

					if len(assetTypes) > 1 {
						for z := range assetTypes {
							result, err = b.Exchanges[x].UpdateTicker(currency,
								assetTypes[z])
							b.printSummary(result, currency, assetTypes[z], exchangeName, err)
							if err == nil {
								relayWebsocketEvent(result, "ticker_update", assetTypes[z], exchangeName)
							}
						}
					} else {
						result, err = b.Exchanges[x].UpdateTicker(currency,
							assetTypes[0])
						b.printSummary(result, currency, assetTypes[0], exchangeName, err)
						if err == nil {
							relayWebsocketEvent(result, "ticker_update", assetTypes[0], exchangeName)
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 10)
	}
}

// OrderbookUpdaterRoutine updates orderbook
func (b *Bot) OrderbookUpdaterRoutine() {
	log.Println("Starting orderbook updater routine")
	for {
		for x := range b.Exchanges {
			if b.Exchanges[x].IsEnabled() {
				if b.Exchanges[x].GetName() == "ANX" {
					continue
				}

				exchangeName := b.Exchanges[x].GetName()
				enabledCurrencies := b.Exchanges[x].GetEnabledCurrencies()
				var result orderbook.Base
				var err error
				var assetTypes []string

				assetTypes, err = exchange.GetExchangeAssetTypes(exchangeName)
				if err != nil {
					log.Printf("failed to get %s exchange asset types. Error: %s",
						exchangeName, err)
				}

				for y := range enabledCurrencies {
					currency := enabledCurrencies[y]

					if len(assetTypes) > 1 {
						for z := range assetTypes {
							result, err = b.Exchanges[x].UpdateOrderbook(currency,
								assetTypes[z])
							b.printOrderbookSummary(result, currency, assetTypes[z], exchangeName, err)
							if err == nil {
								relayWebsocketEvent(result, "orderbook_update", assetTypes[z], exchangeName)
							}
						}
					} else {
						result, err = b.Exchanges[x].UpdateOrderbook(currency,
							assetTypes[0])
						b.printOrderbookSummary(result, currency, assetTypes[0], exchangeName, err)
						if err == nil {
							relayWebsocketEvent(result, "orderbook_update", assetTypes[0], exchangeName)
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 10)
	}
}
