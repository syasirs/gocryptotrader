package gct

import (
	"fmt"
	"strings"

	"github.com/d5/tengo/objects"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/gctscript/modules"
)

var exchangeModule = map[string]objects.Object{
	"orderbook":      &objects.UserFunction{Name: "orderbook", Value: ExchangeOrderbook},
	"ticker":         &objects.UserFunction{Name: "ticker", Value: ExchangeTicker},
	"exchanges":      &objects.UserFunction{Name: "exchanges", Value: ExchangeExchanges},
	"pairs":          &objects.UserFunction{Name: "pairs", Value: ExchangePairs},
	"accountinfo":    &objects.UserFunction{Name: "accountinfo", Value: ExchangeAccountInfo},
	"depositaddress": &objects.UserFunction{Name: "depositaddress", Value: ExchangeDepositAddress},
	"orderquery":     &objects.UserFunction{Name: "orderquery", Value: ExchangeOrderQuery},
	"ordercancel":    &objects.UserFunction{Name: "ordercancel", Value: ExchangeOrderCancel},
	"ordersubmit":    &objects.UserFunction{Name: "ordersubmit", Value: ExchangeOrderSubmit},
}

// ExchangeOrderbook returns orderbook for requested exchange & currencypair
func ExchangeOrderbook(args ...objects.Object) (ret objects.Object, err error) {
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
		return
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

	return &objects.Map{
		Value: data,
	}, nil
}

// ExchangeTicker returns ticker data for requested exchange and currency pair
func ExchangeTicker(args ...objects.Object) (ret objects.Object, err error) {
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
		return
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

// ExchangeExchanges returns list of exchanges either enabled or all
func ExchangeExchanges(args ...objects.Object) (ret objects.Object, err error) {
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

// ExchangePairs returns currency pairs for requested exchange
func ExchangePairs(args ...objects.Object) (ret objects.Object, err error) {
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
		return
	}

	r := objects.Array{}
	for x := range *rtnValue {
		v := *rtnValue
		fmt.Println(v)
		r.Value = append(r.Value, &objects.String{Value: v[x].String()})
	}

	return &r, nil
}

// ExchangeAccountInfo returns account information for requested exchange
func ExchangeAccountInfo(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 1 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	rtnValue, err := modules.Wrapper.AccountInformation(exchangeName)
	if err != nil {
		return
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

// ExchangeOrderQuery query order on exchange
func ExchangeOrderQuery(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 2 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	orderID, _ := objects.ToString(args[1])

	orderDetails, err := modules.Wrapper.QueryOrder(exchangeName, orderID)
	if err != nil {
		return
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

// ExchangeOrderCancel cancels order on requested exchange
func ExchangeOrderCancel(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 2 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	orderID, _ := objects.ToString(args[1])

	rtn, err := modules.Wrapper.CancelOrder(exchangeName, orderID)
	if err != nil {
		return
	}

	if rtn {
		return objects.TrueValue, nil
	}
	return objects.FalseValue, nil
}

// ExchangeOrderSubmit submit order on exchange
func ExchangeOrderSubmit(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 8 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	currencyPair, _ := objects.ToString(args[1])
	delimiter, _ := objects.ToString(args[2])

	orderType, _ := objects.ToString(args[3])
	orderSide, _ := objects.ToString(args[4])
	orderPrice, _ := objects.ToFloat64(args[5])
	orderAmount, _ := objects.ToFloat64(args[6])
	orderClientID, _ := objects.ToString(args[7])

	pair := currency.NewPairDelimiter(currencyPair, delimiter)

	tempSubmit := &order.Submit{
		Pair:      pair,
		OrderType: order.Type(orderType),
		OrderSide: order.Side(orderSide),
		Price:     orderPrice,
		Amount:    orderAmount,
		ClientID:  orderClientID,
	}

	err = tempSubmit.Validate()
	if err != nil {
		return
	}

	rtn, err := modules.Wrapper.SubmitOrder(exchangeName, tempSubmit)
	if err != nil {
		return
	}

	data := make(map[string]objects.Object, 2)
	data["orderid"] = &objects.String{Value: rtn.OrderID}
	if rtn.IsOrderPlaced {
		data["isorderplaced"] = objects.TrueValue
	} else {
		data["isorderplaced"] = objects.FalseValue
	}

	return &objects.Map{
		Value: data,
	}, nil
}

// ExchangeDepositAddress returns deposit address (if supported by exchange)
func ExchangeDepositAddress(args ...objects.Object) (ret objects.Object, err error) {
	if len(args) != 3 {
		err = objects.ErrWrongNumArguments
		return
	}

	exchangeName, _ := objects.ToString(args[0])
	currencyCode, _ := objects.ToString(args[1])
	accountID, _ := objects.ToString(args[2])

	currCode := currency.NewCode(currencyCode)

	rtn, err := modules.Wrapper.DepositAddress(exchangeName, currCode, accountID)
	if err != nil {
		return
	}

	return &objects.String{Value: rtn}, nil
}
