package indicators

import (
	"errors"
	"fmt"
	"math"
	"strings"

	objects "github.com/d5/tengo/v2"
	"github.com/thrasher-corp/go-talib/indicators"
	"github.com/thrasher-corp/gocryptotrader/gctscript/modules"
	"github.com/thrasher-corp/gocryptotrader/gctscript/wrappers/validator"
)

// RsiModule relative strength index indicator commands
var RsiModule = map[string]objects.Object{
	"calculate": &objects.UserFunction{Name: "calculate", Value: rsi},
}

func rsi(args ...objects.Object) (objects.Object, error) {
	if len(args) != 2 {
		return nil, objects.ErrWrongNumArguments
	}

	r := &objects.Array{}
	if validator.IsTestExecution.Load() == true {
		return r, nil
	}

	ohlcvInput := objects.ToInterface(args[0])
	ohlcvInputData, valid := ohlcvInput.([]interface{})
	if !valid {
		return nil, fmt.Errorf(modules.ErrParameterConvertFailed, OHLCV)
	}

	var ohlcvClose []float64
	var allErrors []string
	for x := range ohlcvInputData {
		t := ohlcvInputData[x].([]interface{})

		value, err := toFloat64(t[4])
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
		ohlcvClose = append(ohlcvClose, value)
	}

	inTimePeriod, ok := objects.ToInt(args[1])
	if !ok {
		return nil, fmt.Errorf(modules.ErrParameterConvertFailed, inTimePeriod)
	}

	if len(allErrors) > 0 {
		return nil, errors.New(strings.Join(allErrors, ", "))
	}

	ret := indicators.Rsi(ohlcvClose, inTimePeriod)
	for x := range ret {
		r.Value = append(r.Value, &objects.Float{Value: math.Round(ret[x]*100) / 100})
	}

	return r, nil
}
