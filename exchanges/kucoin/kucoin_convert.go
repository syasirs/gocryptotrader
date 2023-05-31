package kucoin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// kucoinTime provides an internal conversion helper
type kucoinTime time.Time

// UnmarshalJSON is custom type json unmarshaller for kucoinTimeSec
func (k *kucoinTime) UnmarshalJSON(data []byte) error {
	var timestamp interface{}
	err := json.Unmarshal(data, &timestamp)
	if err != nil {
		return err
	}
	var standard uint64
	switch value := timestamp.(type) {
	case string:
		if value == "" {
			// Setting the time to zero value because some timestamp fields could return an empty string while there is no error
			// So, in such cases, kucoinTime returns zero timestamp.
			*k = kucoinTime(time.Time{})
			return nil
		}
		standard, err = strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
	case uint64:
		standard = value
	case float64:
		standard = uint64(value)
	case uint32:
		standard = uint64(value)
	case nil:
		// for some kucoin timestamp fields, if the timestamp information is not specified,
		// the data is 'nil' instead of zero value string or integer value.
	default:
		return fmt.Errorf("unsupported timestamp type %T", timestamp)
	}

	switch {
	case standard == 0:
		*k = kucoinTime(time.Time{})
	case standard >= 1e13:
		*k = kucoinTime(time.Unix(int64(standard/1e9), int64(standard%1e9)))
	case standard > 9999999999:
		*k = kucoinTime(time.UnixMilli(int64(standard)))
	default:
		*k = kucoinTime(time.Unix(int64(standard), 0))
	}
	return nil
}

// Time returns a time.Time instance from kucoinTime instance object.
func (k *kucoinTime) Time() time.Time {
	return time.Time(*k)
}

// UnmarshalJSON valid data to SubAccountsResponse of return nil if the data is empty list.
// this is added to handle the empty list returned when there are no accounts.
func (a *SubAccountsResponse) UnmarshalJSON(data []byte) error {
	var result interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	var ok bool
	if a, ok = result.(*SubAccountsResponse); ok {
		if a == nil {
			return errNoValidResponseFromServer
		}
		return nil
	} else if _, ok := result.([]interface{}); ok {
		return nil
	}
	return fmt.Errorf("%w can not unmarshal to SubAccountsResponse", errMalformedData)
}

// kucoinNumber unmarshals and extract numeric value from a byte slice.
type kucoinNumber float64

// Float64 returns an float64 value from kucoinNumeric instance
func (a *kucoinNumber) Float64() float64 {
	return float64(*a)
}

// UnmarshalJSON decerializes integer and string data having an integer value to int64
func (a *kucoinNumber) UnmarshalJSON(data []byte) error {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	switch val := value.(type) {
	case float64:
		*a = kucoinNumber(val)
	case float32:
		*a = kucoinNumber(val)
	case string:
		if val == "" {
			*a = kucoinNumber(0) // setting empty string value to zero to reset previous value if exist.
			return nil
		}
		value, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*a = kucoinNumber(value)
	case int64:
		*a = kucoinNumber(val)
	case int32:
		*a = kucoinNumber(val)
	default:
		return fmt.Errorf("unsupported input numeric type %T", value)
	}
	return nil
}
