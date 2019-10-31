package gct

import (
	"strings"

	"github.com/d5/tengo/objects"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/gctscript/modules"
)

var exchangeModule = map[string]objects.Object{
	"orderbook":   &objects.UserFunction{Name: "orderbook", Value: exchangeOrderbook},
	"ticker":      &objects.UserFunction{Name: "ticker", Value: exchangeTicker},
	"exchanges":   &objects.UserFunction{Name: "exchanges", Value: exchangeExchanges},
	"pairs":       &objects.UserFunction{Name: "pairs", Value: exchangePairs},
	"accountinfo": &objects.UserFunction{Name: "accountinfo", Value: exchangeAccountInfo},
	"orderquery":  &objects.UserFunction{Name: "order", Value: exchangeOrderQuery},
	"ordercancel": &objects.UserFunction{Name: "order", Value: exchangeOrderCancel},
	"ordersubmit": &objects.UserFunction{Name: "order", Value: exchangeOrderSubmit},
}

func exchangeOrderbook(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 4 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	currencyPair, _ := objects.ToString(args[1])
	delimiter, _ := objects.ToString(args[2])
	assetTypeParam, _ := objects.ToString(args[3])

	pairs := currency.NewPairDelimiter(currencyPair, delimiter)
	assetType := asset.Item(assetTypeParam)

	ob, err := modules.Wrapper.Orderbook(exchangeName, pairs, assetType)
	if err != nil {
		return nil, err
	}

	var asks, bids objects.Array

	for x := range ob.Asks {
		temp := make(map[string]objects.Object, 2)
		temp["amount"] = &objects.Float{Value: ob.Asks[x].Amount}
		temp["price"] = &objects.Float{Value: ob.Asks[x].Price}
		asks.Value = append(asks.Value, &objects.Map{Value: temp})
	}

	for x := range ob.Bids {
		temp := make(map[string]objects.Object, 2)
		temp["amount"] = &objects.Float{Value: ob.Bids[x].Amount}
		temp["price"] = &objects.Float{Value: ob.Bids[x].Price}
		bids.Value = append(bids.Value, &objects.Map{Value: temp})
	}

	data := make(map[string]objects.Object, 13)
	data["exchange"] = &objects.String{Value: ob.ExchangeName}
	data["pair"] = &objects.String{Value: ob.Pair.String()}
	data["asks"] = &asks
	data["bids"] = &bids
	data["asset"] = &objects.String{Value: ob.AssetType.String()}

	r := objects.Map{
		Value: data,
	}

	return &r, nil
}

func exchangeTicker(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 4 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	currencyPair, _ := objects.ToString(args[1])
	delimiter, _ := objects.ToString(args[2])
	assetTypeParam, _ := objects.ToString(args[3])

	pairs := currency.NewPairDelimiter(currencyPair, delimiter)
	assetType := asset.Item(assetTypeParam)

	tx, err := modules.Wrapper.Ticker(exchangeName, pairs, assetType)
	if err != nil {
		return nil, err
	}

	data := make(map[string]objects.Object, 13)
	data["exchange"] = &objects.String{Value: tx.ExchangeName}
	data["last"] = &objects.Float{Value: tx.Last}
	data["High"] = &objects.Float{Value: tx.High}
	data["Low"] = &objects.Float{Value: tx.Low}
	data["bid"] = &objects.Float{Value: tx.Bid}
	data["ask"] = &objects.Float{Value: tx.Ask}
	data["volume"] = &objects.Float{Value: tx.Volume}
	data["quotevolume"] = &objects.Float{Value: tx.QuoteVolume}
	data["priceath"] = &objects.Float{Value: tx.PriceATH}
	data["open"] = &objects.Float{Value: tx.Open}
	data["close"] = &objects.Float{Value: tx.Close}
	data["pair"] = &objects.String{Value: tx.Pair.String()}
	data["asset"] = &objects.String{Value: tx.AssetType.String()}
	data["updated"] = &objects.Time{Value: tx.LastUpdated}

	return &objects.Map{
		Value: data,
	}, nil
}

func exchangeExchanges(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		err = objects.ErrWrongNumArguments
		return
	}

	enabledOnly, _ := objects.ToBool(args[0])
	rtnValue := modules.Wrapper.Exchanges(enabledOnly)

	r := objects.Array{}
	for x := range rtnValue {
		r.Value = append(r.Value, &objects.String{Value: rtnValue[x]})
	}

	return &r, nil
}

func exchangePairs(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 3 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	enabledOnly, _ := objects.ToBool(args[1])
	assetTypeParam, _ := objects.ToString(args[2])
	assetType := asset.Item(strings.ToLower(assetTypeParam))

	rtnValue, err := modules.Wrapper.Pairs(exchangeName, enabledOnly, assetType)
	if err != nil {
		return nil, err
	}

	r := objects.Array{}
	for x := range rtnValue {
		r.Value = append(r.Value, &objects.String{Value: rtnValue[x].String()})
	}

	return &r, nil
}

func exchangeAccountInfo(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	rtnValue, err := modules.Wrapper.AccountInformation(exchangeName)
	if err != nil {
		return nil, err
	}

	var funds objects.Array
	for x := range rtnValue.Accounts {
		for y := range rtnValue.Accounts[x].Currencies {
			temp := make(map[string]objects.Object, 3)
			temp["name"] = &objects.String{Value: rtnValue.Accounts[x].Currencies[y].CurrencyName.String()}
			temp["total"] = &objects.Float{Value: rtnValue.Accounts[x].Currencies[y].TotalValue}
			temp["hold"] = &objects.Float{Value: rtnValue.Accounts[x].Currencies[y].Hold}
			funds.Value = append(funds.Value, &objects.Map{Value: temp})
		}
	}

	data := make(map[string]objects.Object, 2)
	data["exchange"] = &objects.String{Value: rtnValue.Exchange}
	data["currencies"] = &funds

	return &objects.Map{
		Value: data,
	}, nil
}

func exchangeOrderQuery(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 2 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	orderID, _ := objects.ToString(args[1])

	orderDetails, err := modules.Wrapper.QueryOrder(exchangeName, orderID)
	if err != nil {
		return nil, err
	}

	var tradeHistory objects.Array
	for x := range orderDetails.Trades {
		temp := make(map[string]objects.Object, 3)
		temp["timestamp"] = &objects.Time{Value: orderDetails.Trades[x].Timestamp}
		temp["price"] = &objects.Float{Value: orderDetails.Trades[x].Price}
		temp["fee"] = &objects.Float{Value: orderDetails.Trades[x].Fee}
		temp["amount"] = &objects.Float{Value: orderDetails.Trades[x].Amount}
		temp["type"] = &objects.String{Value: orderDetails.Trades[x].Type.String()}
		temp["side"] = &objects.String{Value: orderDetails.Trades[x].Side.String()}
		temp["description"] = &objects.String{Value: orderDetails.Trades[x].Description}
		tradeHistory.Value = append(tradeHistory.Value, &objects.Map{Value: temp})
	}

	data := make(map[string]objects.Object, 13)
	data["exchange"] = &objects.String{Value: orderDetails.Exchange}
	data["id"] = &objects.String{Value: orderDetails.ID}
	data["accountid"] = &objects.String{Value: orderDetails.AccountID}
	data["currencypair"] = &objects.String{Value: orderDetails.CurrencyPair.String()}
	data["price"] = &objects.Float{Value: orderDetails.Price}
	data["amount"] = &objects.Float{Value: orderDetails.Amount}
	data["amountexecuted"] = &objects.Float{Value: orderDetails.ExecutedAmount}
	data["amountremaining"] = &objects.Float{Value: orderDetails.RemainingAmount}
	data["fee"] = &objects.Float{Value: orderDetails.Fee}
	data["side"] = &objects.String{Value: orderDetails.OrderSide.String()}
	data["type"] = &objects.String{Value: orderDetails.OrderType.String()}
	data["date"] = &objects.String{Value: orderDetails.OrderDate.String()}
	data["status"] = &objects.String{Value: orderDetails.Status.String()}
	data["trades"] = &tradeHistory

	return &objects.Map{
		Value: data,
	}, nil
}

func exchangeOrderCancel(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 2 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	orderID, _ := objects.ToString(args[1])

	_, err = modules.Wrapper.CancelOrder(exchangeName, orderID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func exchangeOrderSubmit(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 2 {
		err = objects.ErrWrongNumArguments
		return
	}
	return nil, nil
}
