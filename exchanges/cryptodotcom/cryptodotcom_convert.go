package cryptodotcom

import (
	"encoding/json"
	"strconv"
	"time"
)

type cryptoDotComMilliSec int64
type cryptoDotComMilliSecString int64

// UnmarshalJSON decerializes timestamp information into a cryptoDotComMilliSec instance.
func (d *cryptoDotComMilliSec) UnmarshalJSON(data []byte) error {
	var value int64
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	*d = cryptoDotComMilliSec(value)
	return nil
}

// Time returns a time.Time instance from the timestamp information.
func (d *cryptoDotComMilliSec) Time() time.Time {
	return time.UnixMilli(int64(*d))
}

// UnmarshalJSON decerializes timestamp information into a cryptoDotComMilliSec instance.
func (d *cryptoDotComMilliSecString) UnmarshalJSON(data []byte) error {
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	*d = cryptoDotComMilliSecString(val)
	return nil
}

// Time returns a time.Time instance from the timestamp information.
func (d *cryptoDotComMilliSecString) Time() time.Time {
	return time.UnixMilli(int64(*d))
}
