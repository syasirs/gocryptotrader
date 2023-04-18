package okx

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
)

type okxNumericalValue float64

// UnmarshalJSON is custom type json unmarshaller for okxNumericalValue
func (a *okxNumericalValue) UnmarshalJSON(data []byte) error {
	var num string
	err := json.Unmarshal(data, &num)
	if err != nil {
		return err
	}

	if num == "" {
		return nil
	}

	v, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return err
	}

	*a = okxNumericalValue(v)
	return nil
}

// Float64 returns a float64 value for okxNumericalValue
func (a *okxNumericalValue) Float64() float64 { return float64(*a) }

type okxUnixMilliTime int64

// UnmarshalJSON deserializes byte data to okxunixMilliTime instance.
func (a *okxUnixMilliTime) UnmarshalJSON(data []byte) error {
	var num string
	err := json.Unmarshal(data, &num)
	if err != nil {
		return err
	}
	if num == "" {
		return nil
	}
	value, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return err
	}
	*a = okxUnixMilliTime(value)
	return nil
}

// Time returns the time instance from unix value of integer.
func (a *okxUnixMilliTime) Time() time.Time {
	return time.UnixMilli(int64(*a))
}

// numbersOnlyRegexp for checking the value is numerics only
var numbersOnlyRegexp = regexp.MustCompile(`^\d*$`)

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *Instrument) UnmarshalJSON(data []byte) error {
	type Alias Instrument
	chil := &struct {
		*Alias
		ListTime string `json:"listTime"`
		ExpTime  string `json:"expTime"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	if numbersOnlyRegexp.MatchString(chil.ListTime) {
		var val int
		if val, err = strconv.Atoi(chil.ListTime); err == nil {
			a.ListTime = time.UnixMilli(int64(val))
		}
	}
	if numbersOnlyRegexp.MatchString(chil.ExpTime) {
		var val int
		if val, err = strconv.Atoi(chil.ExpTime); err == nil {
			a.ExpTime = time.UnixMilli(int64(val))
		}
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *FundingRateResponse) UnmarshalJSON(data []byte) error {
	type Alias FundingRateResponse
	chil := &struct {
		*Alias
		FundingRate string `json:"fundingRate"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *LimitPriceResponse) UnmarshalJSON(data []byte) error {
	type Alias LimitPriceResponse
	chil := &struct {
		*Alias
		Timestamp int64 `json:"ts,string"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *OrderDetail) UnmarshalJSON(data []byte) error {
	type Alias OrderDetail
	chil := &struct {
		*Alias
		Side         string `json:"side"`
		UpdateTime   int64  `json:"uTime,string"`
		CreationTime int64  `json:"cTime,string"`
		FillTime     string `json:"fillTime"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	var err error
	a.UpdateTime = time.UnixMilli(chil.UpdateTime)
	a.CreationTime = time.UnixMilli(chil.CreationTime)
	a.Side, err = order.StringToOrderSide(chil.Side)
	if chil.FillTime == "" {
		a.FillTime = time.Time{}
	} else {
		var value int64
		value, err = strconv.ParseInt(chil.FillTime, 10, 64)
		if err != nil {
			return err
		}
		a.FillTime = time.UnixMilli(value)
	}
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *PendingOrderItem) UnmarshalJSON(data []byte) error {
	type Alias PendingOrderItem
	chil := &struct {
		*Alias
		Side         string `json:"side"`
		UpdateTime   string `json:"uTime"`
		CreationTime string `json:"cTime"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	uTime, err := strconv.ParseInt(chil.UpdateTime, 10, 64)
	if err != nil {
		return err
	}
	cTime, err := strconv.ParseInt(chil.CreationTime, 10, 64)
	if err != nil {
		return err
	}
	a.Side, err = order.StringToOrderSide(chil.Side)
	if err != nil {
		return err
	}
	a.CreationTime = time.UnixMilli(cTime)
	a.UpdateTime = time.UnixMilli(uTime)
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *RfqTradeResponse) UnmarshalJSON(data []byte) error {
	type Alias RfqTradeResponse
	chil := &struct {
		*Alias
		CreationTime int64 `json:"cTime,string"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	a.CreationTime = time.UnixMilli(chil.CreationTime)
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *BlockTicker) UnmarshalJSON(data []byte) error {
	type Alias BlockTicker
	chil := &struct {
		*Alias
		Timestamp int64 `json:"ts,string"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *BlockTrade) UnmarshalJSON(data []byte) error {
	type Alias BlockTrade
	chil := &struct {
		*Alias
		Side string `json:"side"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	switch {
	case strings.EqualFold(chil.Side, "buy"):
		a.Side = order.Buy
	case strings.EqualFold(chil.Side, "sell"):
		a.Side = order.Sell
	default:
		a.Side = order.UnknownSide
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *UnitConvertResponse) UnmarshalJSON(data []byte) error {
	type Alias UnitConvertResponse
	chil := &struct {
		*Alias
		ConvertType int `json:"type,string"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	switch chil.ConvertType {
	case 1:
		a.ConvertType = 1
	case 2:
		a.ConvertType = 2
	}
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *QuoteLeg) UnmarshalJSON(data []byte) error {
	type Alias QuoteLeg
	chil := &struct {
		*Alias
		Side string `json:"side"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	chil.Side = strings.ToLower(chil.Side)
	if chil.Side == "buy" {
		a.Side = order.Buy
	} else {
		a.Side = order.Sell
	}
	return nil
}

// MarshalJSON serialized QuoteLeg instance into bytes
func (a *QuoteLeg) MarshalJSON() ([]byte, error) {
	type Alias QuoteLeg
	chil := &struct {
		*Alias
		Side string `json:"side"`
	}{
		Alias: (*Alias)(a),
	}
	if a.Side == order.Buy {
		chil.Side = "buy"
	} else {
		chil.Side = "sell"
	}
	return json.Marshal(chil)
}

// MarshalJSON serialized CreateQuoteParams instance into bytes
func (a *CreateQuoteParams) MarshalJSON() ([]byte, error) {
	type Alias CreateQuoteParams
	chil := &struct {
		*Alias
		QuoteSide string `json:"quoteSide"`
	}{
		Alias: (*Alias)(a),
	}
	if a.QuoteSide == order.Buy {
		chil.QuoteSide = "buy"
	} else {
		chil.QuoteSide = "sell"
	}
	return json.Marshal(chil)
}

// MarshalJSON serializes the WebsocketLoginData object
func (a *WebsocketLoginData) MarshalJSON() ([]byte, error) {
	type Alias WebsocketLoginData
	return json.Marshal(struct {
		Timestamp int64 `json:"timestamp"`
		*Alias
	}{
		Timestamp: a.Timestamp.UTC().Unix(),
		Alias:     (*Alias)(a),
	})
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *WebsocketLoginData) UnmarshalJSON(data []byte) error {
	type Alias WebsocketLoginData
	chil := &struct {
		*Alias
		Timestamp int64 `json:"timestamp"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, chil); err != nil {
		return err
	}
	a.Timestamp = time.UnixMilli(chil.Timestamp)
	return nil
}

// UnmarshalJSON deserializes JSON, and timestamp information.
func (a *CurrencyOneClickRepay) UnmarshalJSON(data []byte) error {
	type Alias CurrencyOneClickRepay
	chil := &struct {
		*Alias
		UpdateTime   int64  `json:"uTime,string"`
		FillToSize   string `json:"fillToSz"`
		FillFromSize string `json:"fillFromSz"`
	}{
		Alias: (*Alias)(a),
	}
	err := json.Unmarshal(data, chil)
	if err != nil {
		return err
	}
	a.UpdateTime = time.Unix(chil.UpdateTime, 0)
	return nil
}
