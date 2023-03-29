package engine

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/portfolio/banking"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

func TestLoadConfigWithSettings(t *testing.T) {
	empty := ""
	somePath := "somePath"
	// Clean up after the tests
	defer os.RemoveAll(somePath)
	tests := []struct {
		name     string
		flags    []string
		settings *Settings
		want     *string
		wantErr  bool
	}{
		{
			name: "invalid file",
			settings: &Settings{
				ConfigFile: "nonExistent.json",
			},
			wantErr: true,
		},
		{
			name: "test file",
			settings: &Settings{
				ConfigFile:   config.TestFile,
				EnableDryRun: true,
			},
			want:    &empty,
			wantErr: false,
		},
		{
			name:  "data dir in settings overrides config data dir",
			flags: []string{"datadir"},
			settings: &Settings{
				ConfigFile:   config.TestFile,
				DataDir:      somePath,
				EnableDryRun: true,
			},
			want:    &somePath,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// prepare the 'flags'
			flagSet := make(map[string]bool)
			for _, v := range tt.flags {
				flagSet[v] = true
			}
			// Run the test
			got, err := loadConfigWithSettings(tt.settings, flagSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfigWithSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil || tt.want != nil {
				if (got == nil && tt.want != nil) || (got != nil && tt.want == nil) {
					t.Errorf("loadConfigWithSettings() = is nil %v, want nil %v", got == nil, tt.want == nil)
				} else if got.DataDirectory != *tt.want {
					t.Errorf("loadConfigWithSettings() = %v, want %v", got.DataDirectory, *tt.want)
				}
			}
		})
	}
}

func TestStartStopDoesNotCausePanic(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	botOne, err := NewFromSettings(&Settings{
		ConfigFile:   config.TestFile,
		EnableDryRun: true,
		DataDir:      tempDir,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	botOne.Settings.EnableGRPCProxy = false
	for i := range botOne.Config.Exchanges {
		if botOne.Config.Exchanges[i].Name != testExchange {
			// there is no need to load all exchanges for this test
			botOne.Config.Exchanges[i].Enabled = false
		}
	}
	if err = botOne.Start(); err != nil {
		t.Error(err)
	}

	botOne.Stop()
}

var enableExperimentalTest = false

func TestStartStopTwoDoesNotCausePanic(t *testing.T) {
	t.Parallel()
	if !enableExperimentalTest {
		t.Skip("test is functional, however does not need to be included in go test runs")
	}
	tempDir := t.TempDir()
	tempDir2 := t.TempDir()
	botOne, err := NewFromSettings(&Settings{
		ConfigFile:   config.TestFile,
		EnableDryRun: true,
		DataDir:      tempDir,
	}, nil)
	if err != nil {
		t.Error(err)
	}
	botOne.Settings.EnableGRPCProxy = false

	botTwo, err := NewFromSettings(&Settings{
		ConfigFile:   config.TestFile,
		EnableDryRun: true,
		DataDir:      tempDir2,
	}, nil)
	if err != nil {
		t.Error(err)
	}
	botTwo.Settings.EnableGRPCProxy = false

	if err = botOne.Start(); err != nil {
		t.Error(err)
	}
	if err = botTwo.Start(); err != nil {
		t.Error(err)
	}

	botOne.Stop()
	botTwo.Stop()
}

func TestGetExchangeByName(t *testing.T) {
	t.Parallel()
	_, err := (*ExchangeManager)(nil).GetExchangeByName("tehehe")
	if !errors.Is(err, ErrNilSubsystem) {
		t.Errorf("received: %v expected: %v", err, ErrNilSubsystem)
	}

	em := SetupExchangeManager()
	exch, err := em.NewExchangeByName(testExchange)
	if !errors.Is(err, nil) {
		t.Fatalf("received '%v' expected '%v'", err, nil)
	}
	exch.SetDefaults()
	exch.SetEnabled(true)
	em.Add(exch)
	e := &Engine{ExchangeManager: em}

	if !exch.IsEnabled() {
		t.Errorf("TestGetExchangeByName: Unexpected result")
	}

	exch.SetEnabled(false)
	bfx, err := e.GetExchangeByName(testExchange)
	if err != nil {
		t.Fatal(err)
	}
	if bfx.IsEnabled() {
		t.Errorf("TestGetExchangeByName: Unexpected result")
	}
	if exch.GetName() != testExchange {
		t.Errorf("TestGetExchangeByName: Unexpected result")
	}

	_, err = e.GetExchangeByName("Asdasd")
	if !errors.Is(err, ErrExchangeNotFound) {
		t.Errorf("received: %v expected: %v", err, ErrExchangeNotFound)
	}
}

func TestUnloadExchange(t *testing.T) {
	t.Parallel()
	em := SetupExchangeManager()
	exch, err := em.NewExchangeByName(testExchange)
	if !errors.Is(err, nil) {
		t.Fatalf("received '%v' expected '%v'", err, nil)
	}
	exch.SetDefaults()
	exch.SetEnabled(true)
	em.Add(exch)
	e := &Engine{ExchangeManager: em,
		Config: &config.Config{Exchanges: []config.Exchange{{Name: testExchange}}},
	}
	err = e.UnloadExchange("asdf")
	if !errors.Is(err, config.ErrExchangeNotFound) {
		t.Errorf("error '%v', expected '%v'", err, config.ErrExchangeNotFound)
	}

	err = e.UnloadExchange(testExchange)
	if err != nil {
		t.Errorf("TestUnloadExchange: Failed to get exchange. %s",
			err)
	}

	err = e.UnloadExchange(testExchange)
	if !errors.Is(err, ErrNoExchangesLoaded) {
		t.Errorf("error '%v', expected '%v'", err, ErrNoExchangesLoaded)
	}
}

func TestDryRunParamInteraction(t *testing.T) {
	t.Parallel()
	bot := &Engine{
		ExchangeManager: SetupExchangeManager(),
		Settings:        Settings{},
		Config: &config.Config{
			Exchanges: []config.Exchange{
				{
					Name:                    testExchange,
					WebsocketTrafficTimeout: time.Second,
				},
			},
		},
	}
	if err := bot.LoadExchange(testExchange, nil); err != nil {
		t.Error(err)
	}
	exchCfg, err := bot.Config.GetExchangeConfig(testExchange)
	if err != nil {
		t.Error(err)
	}
	if exchCfg.Verbose {
		t.Error("verbose should have been disabled")
	}
	if err = bot.UnloadExchange(testExchange); err != nil {
		t.Error(err)
	}

	// Now set dryrun mode to true,
	// enable exchange verbose mode and verify that verbose mode
	// will be set on Bitfinex
	bot.Settings.EnableDryRun = true
	bot.Settings.CheckParamInteraction = true
	bot.Settings.EnableExchangeVerbose = true
	if err = bot.LoadExchange(testExchange, nil); err != nil {
		t.Error(err)
	}

	exchCfg, err = bot.Config.GetExchangeConfig(testExchange)
	if err != nil {
		t.Error(err)
	}
	if !bot.Settings.EnableDryRun ||
		!exchCfg.Verbose {
		t.Error("dryrun should be true and verbose should be true")
	}
}

func TestFlagSetWith(t *testing.T) {
	var isRunning bool
	flags := make(FlagSet)
	// Flag not set default to config
	flags.WithBool("NOT SET", &isRunning, true)
	if !isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, true)
	}
	flags.WithBool("NOT SET", &isRunning, false)
	if isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, false)
	}

	flags["IS SET"] = true
	isRunning = true
	// Flag set true which will override config
	flags.WithBool("IS SET", &isRunning, true)
	if !isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, true)
	}
	flags.WithBool("IS SET", &isRunning, false)
	if !isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, true)
	}

	flags["IS SET"] = true
	isRunning = false
	// Flag set false which will override config
	flags.WithBool("IS SET", &isRunning, true)
	if isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, false)
	}
	flags.WithBool("IS SET", &isRunning, false)
	if isRunning {
		t.Fatalf("received: '%v' but expected: '%v'", isRunning, false)
	}
}

func TestRegisterWebsocketDataHandler(t *testing.T) {
	t.Parallel()
	var e *Engine
	err := e.RegisterWebsocketDataHandler(nil, false)
	if !errors.Is(err, errNilBot) {
		t.Fatalf("received: '%v' but expected: '%v'", err, errNilBot)
	}

	e = &Engine{websocketRoutineManager: &websocketRoutineManager{}}
	err = e.RegisterWebsocketDataHandler(func(_ string, _ interface{}) error { return nil }, false)
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v' but expected: '%v'", err, nil)
	}
}

func TestSetDefaultWebsocketDataHandler(t *testing.T) {
	t.Parallel()
	var e *Engine
	err := e.SetDefaultWebsocketDataHandler()
	if !errors.Is(err, errNilBot) {
		t.Fatalf("received: '%v' but expected: '%v'", err, errNilBot)
	}

	e = &Engine{websocketRoutineManager: &websocketRoutineManager{}}
	err = e.SetDefaultWebsocketDataHandler()
	if !errors.Is(err, nil) {
		t.Fatalf("received: '%v' but expected: '%v'", err, nil)
	}
}

func TestAllExchangeWrappers(t *testing.T) {
	t.Parallel()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../testdata/configtest.json", true)
	if err != nil {
		t.Fatal("load config error", err)
	}
	isCITest := os.Getenv("CI_TEST")
	for i := range cfg.Exchanges {
		name := strings.ToLower(cfg.Exchanges[i].Name)
		t.Run(name+" wrapper tests", func(t *testing.T) {
			t.Parallel()
			if common.StringDataContains(unsupportedExchangeNames, name) {
				t.Skipf("skipping unsupported exchange %v", name)
			}
			if isCITest == "true" && common.StringDataContains(blockedCIExchanges, name) {
				t.Skipf("cannot execute tests for %v on via continuous integration tests, skipping", name)
			}
			exch, assetPairs := setupExchange(t, name, cfg)
			executeExchangeWrapperTests(t, exch, assetPairs)
		})
	}
}

func setupExchange(t *testing.T, name string, cfg *config.Config) (exchange.IBotExchange, []assetPair) {
	t.Helper()
	em := SetupExchangeManager()
	exch, err := em.NewExchangeByName(name)
	if err != nil {
		t.Fatalf("%v %v", name, err)
	}
	var exchCfg *config.Exchange
	exchCfg, err = cfg.GetExchangeConfig(name)
	if err != nil {
		t.Fatalf("%v %v", name, err)
	}
	exch.SetDefaults()
	exchCfg.API.AuthenticatedSupport = true
	exchCfg.API.Credentials.Key = "realKey"
	exchCfg.API.Credentials.Secret = "realSecret"
	exchCfg.API.Credentials.ClientID = "realClientID"
	err = exch.Setup(exchCfg)
	if err != nil {
		t.Fatalf("%v %v", name, err)
	}

	err = exch.UpdateTradablePairs(context.Background(), true)
	if err != nil {
		t.Fatalf("%v %v", name, err)
	}
	b := exch.GetBase()
	assets := b.CurrencyPairs.GetAssetTypes(false)
	if len(assets) == 0 {
		t.Fatalf("exchange '%v' has not assets", name)
	}
	for j := range assets {
		err = b.CurrencyPairs.SetAssetEnabled(assets[j], true)
		if err != nil && !errors.Is(err, currency.ErrAssetAlreadyEnabled) {
			t.Fatalf("%v %v", name, err)
		}
	}

	// Add +1 to len to verify that exchanges can handle requests with unset pairs and assets
	assetPairs := make([]assetPair, len(assets)+1)
	for j := range assets {
		var pairs currency.Pairs
		pairs, err = b.CurrencyPairs.GetPairs(assets[j], true)
		if err != nil {
			t.Fatalf("%v %v", name, err)
		}
		var p currency.Pair
		if len(pairs) == 0 {
			pairs, err = b.CurrencyPairs.GetPairs(assets[j], false)
			if err != nil {
				t.Fatalf("%v GetPairs %v %v", name, err, assets[j])
			}
			p, err = getPairFromPairs(t, pairs)
			if err != nil {
				t.Fatalf("%v getPairFromPairs %v %v", name, err, assets[j])
			}
			p, err = b.FormatExchangeCurrency(p, assets[j])
			if err != nil {
				t.Fatalf("%v FormatExchangeCurrency %v %v", name, err, assets[j])
			}
			err = b.CurrencyPairs.EnablePair(assets[j], p)
			if err != nil {
				t.Fatalf("%v EnablePair %v %v", name, err, assets[j])
			}
		} else {
			p, err = getPairFromPairs(t, pairs)
			if err != nil {
				t.Fatalf("%v getPairFromPairs %v %v", name, err, assets[j])
			}
		}
		p, err = b.FormatExchangeCurrency(p, assets[j])
		if err != nil {
			t.Fatalf("%v %v", name, err)
		}
		p, err = disruptFormatting(t, p)
		if err != nil {
			t.Fatalf("%v %v", name, err)
		}
		assetPairs[j] = assetPair{
			Pair:  p,
			Asset: assets[j],
		}
	}

	return exch, assetPairs
}

// isUnacceptableError sentences errs to 10 years dungeon if unacceptable
func isUnacceptableError(t *testing.T, err error) error {
	t.Helper()
	for i := range acceptableErrors {
		if errors.Is(err, acceptableErrors[i]) {
			return nil
		}
	}
	for i := range warningErrors {
		if errors.Is(err, warningErrors[i]) {
			t.Log(err)
			return nil
		}
	}
	return err
}

func executeExchangeWrapperTests(t *testing.T, exch exchange.IBotExchange, assetParams []assetPair) {
	t.Helper()
	iExchange := reflect.TypeOf(&exch).Elem()
	actualExchange := reflect.ValueOf(exch)

	e := time.Now().Add(-time.Hour * 24)
	for x := 0; x < iExchange.NumMethod(); x++ {
		methodName := iExchange.Method(x).Name
		if _, ok := excludedMethodNames[methodName]; ok {
			continue
		}
		method := actualExchange.MethodByName(methodName)

		var assetLen int
		for y := 0; y < method.Type().NumIn(); y++ {
			input := method.Type().In(y)
			if input.AssignableTo(assetParam) {
				assetLen = len(assetParams) - 1
			}
		}

		s := time.Now().Add(-time.Hour * 24 * 7).Truncate(time.Hour)
		if methodName == "GetHistoricTrades" {
			// limit trade history
			s = time.Now().Add(-time.Minute * 5)
		}
		for y := 0; y <= assetLen; y++ {
			inputs := make([]reflect.Value, method.Type().NumIn())
			argGenerator := &MethodArgumentGenerator{
				Exchange:    exch,
				AssetParams: &assetParams[y],
				MethodName:  methodName,
				Start:       s,
				End:         e,
			}
			for z := 0; z < method.Type().NumIn(); z++ {
				argGenerator.MethodInputType = method.Type().In(z)
				generatedArg := generateMethodArg(t, argGenerator)
				inputs[z] = *generatedArg
			}
			t.Run(methodName+"-"+assetParams[y].Asset.String()+"-"+assetParams[y].Pair.String(), func(t *testing.T) {
				t.Parallel()
				CallExchangeMethod(t, method, inputs, methodName, exch)
			})
		}
	}
}

// MethodArgumentGenerator is used to create arguments for
// an IBotExchange method
type MethodArgumentGenerator struct {
	Exchange        exchange.IBotExchange
	AssetParams     *assetPair
	MethodInputType reflect.Type
	MethodName      string
	Start           time.Time
	End             time.Time
	StartTimeSet    bool
	argNum          int64
}

var (
	currencyPairParam     = reflect.TypeOf((*currency.Pair)(nil)).Elem()
	klineParam            = reflect.TypeOf((*kline.Interval)(nil)).Elem()
	contextParam          = reflect.TypeOf((*context.Context)(nil)).Elem()
	timeParam             = reflect.TypeOf((*time.Time)(nil)).Elem()
	codeParam             = reflect.TypeOf((*currency.Code)(nil)).Elem()
	assetParam            = reflect.TypeOf((*asset.Item)(nil)).Elem()
	currencyPairsParam    = reflect.TypeOf((*currency.Pairs)(nil)).Elem()
	withdrawRequestParam  = reflect.TypeOf((**withdraw.Request)(nil)).Elem()
	stringParam           = reflect.TypeOf((*string)(nil)).Elem()
	orderSubmitParam      = reflect.TypeOf((**order.Submit)(nil)).Elem()
	orderModifyParam      = reflect.TypeOf((**order.Modify)(nil)).Elem()
	orderCancelParam      = reflect.TypeOf((**order.Cancel)(nil)).Elem()
	orderCancelsParam     = reflect.TypeOf((*[]order.Cancel)(nil)).Elem()
	getOrdersRequestParam = reflect.TypeOf((**order.MultiOrderRequest)(nil)).Elem()
)

// generateMethodArg determines the argument type and returns a pre-made
// response, else an empty version of the type
func generateMethodArg(t *testing.T, argGenerator *MethodArgumentGenerator) *reflect.Value {
	t.Helper()
	exchName := argGenerator.Exchange.GetName()
	var input reflect.Value
	switch {
	case argGenerator.MethodInputType.AssignableTo(stringParam):
		switch argGenerator.MethodName {
		case "GetDepositAddress":
			if argGenerator.argNum == 2 {
				// account type
				input = reflect.ValueOf("trading")
			} else {
				// Crypto Chain
				input = reflect.ValueOf("")
			}
		default:
			// OrderID
			input = reflect.ValueOf("1337")
		}
	case argGenerator.MethodInputType.Implements(contextParam):
		// Need to deploy a context.Context value as nil value is not
		// checked throughout codebase.
		input = reflect.ValueOf(context.Background())
	case argGenerator.MethodInputType.AssignableTo(currencyPairParam):
		input = reflect.ValueOf(argGenerator.AssetParams.Pair)
	case argGenerator.MethodInputType.AssignableTo(assetParam):
		input = reflect.ValueOf(argGenerator.AssetParams.Asset)
	case argGenerator.MethodInputType.AssignableTo(klineParam):
		input = reflect.ValueOf(kline.OneDay)
	case argGenerator.MethodInputType.AssignableTo(codeParam):
		if argGenerator.MethodName == "GetAvailableTransferChains" {
			input = reflect.ValueOf(currency.ETH)
		} else {
			input = reflect.ValueOf(argGenerator.AssetParams.Pair.Quote)
		}
	case argGenerator.MethodInputType.AssignableTo(timeParam):
		if !argGenerator.StartTimeSet {
			input = reflect.ValueOf(argGenerator.Start)
			argGenerator.StartTimeSet = true
		} else {
			input = reflect.ValueOf(argGenerator.End)
		}
	case argGenerator.MethodInputType.AssignableTo(currencyPairsParam):
		input = reflect.ValueOf(currency.Pairs{
			argGenerator.AssetParams.Pair,
		})
	case argGenerator.MethodInputType.AssignableTo(withdrawRequestParam):
		req := &withdraw.Request{
			Exchange:      exchName,
			Description:   "1337",
			Amount:        1,
			ClientOrderID: "1337",
		}
		if argGenerator.MethodName == "WithdrawCryptocurrencyFunds" {
			req.Type = withdraw.Crypto
			switch {
			case !isFiat(t, argGenerator.AssetParams.Pair.Base.Item.Lower):
				req.Currency = argGenerator.AssetParams.Pair.Base
			case !isFiat(t, argGenerator.AssetParams.Pair.Quote.Item.Lower):
				req.Currency = argGenerator.AssetParams.Pair.Quote
			default:
				req.Currency = currency.ETH
			}

			req.Crypto = withdraw.CryptoRequest{
				Address:    "1337",
				AddressTag: "1337",
				Chain:      "ERC20",
			}
		} else {
			req.Type = withdraw.Fiat
			b := argGenerator.Exchange.GetBase()
			if len(b.Config.BaseCurrencies) > 0 {
				req.Currency = b.Config.BaseCurrencies[0]
			} else {
				req.Currency = currency.USD
			}
			req.Fiat = withdraw.FiatRequest{
				Bank: banking.Account{
					Enabled:             true,
					ID:                  "1337",
					BankName:            "1337",
					BankAddress:         "1337",
					BankPostalCode:      "1337",
					BankPostalCity:      "1337",
					BankCountry:         "1337",
					AccountName:         "1337",
					AccountNumber:       "1337",
					SWIFTCode:           "1337",
					IBAN:                "1337",
					BSBNumber:           "1337",
					BankCode:            1337,
					SupportedCurrencies: req.Currency.String(),
					SupportedExchanges:  exchName,
				},
				IsExpressWire:                 false,
				RequiresIntermediaryBank:      false,
				IntermediaryBankAccountNumber: 1338,
				IntermediaryBankName:          "1338",
				IntermediaryBankAddress:       "1338",
				IntermediaryBankCity:          "1338",
				IntermediaryBankCountry:       "1338",
				IntermediaryBankPostalCode:    "1338",
				IntermediarySwiftCode:         "1338",
				IntermediaryBankCode:          1338,
				IntermediaryIBAN:              "1338",
				WireCurrency:                  "1338",
			}
		}
		input = reflect.ValueOf(req)
	case argGenerator.MethodInputType.AssignableTo(orderSubmitParam):
		input = reflect.ValueOf(&order.Submit{
			Exchange:          exchName,
			Type:              order.Limit,
			Side:              order.Buy,
			Pair:              argGenerator.AssetParams.Pair,
			AssetType:         argGenerator.AssetParams.Asset,
			Price:             1337,
			Amount:            1,
			ClientID:          "1337",
			ClientOrderID:     "13371337",
			ImmediateOrCancel: true,
		})
	case argGenerator.MethodInputType.AssignableTo(orderModifyParam):
		input = reflect.ValueOf(&order.Modify{
			Exchange:          exchName,
			Type:              order.Limit,
			Side:              order.Buy,
			Pair:              argGenerator.AssetParams.Pair,
			AssetType:         argGenerator.AssetParams.Asset,
			Price:             1337,
			Amount:            1,
			ClientOrderID:     "13371337",
			OrderID:           "1337",
			ImmediateOrCancel: true,
		})
	case argGenerator.MethodInputType.AssignableTo(orderCancelParam):
		input = reflect.ValueOf(&order.Cancel{
			Exchange:  exchName,
			Type:      order.Limit,
			Side:      order.Buy,
			Pair:      argGenerator.AssetParams.Pair,
			AssetType: argGenerator.AssetParams.Asset,
			OrderID:   "1337",
		})
	case argGenerator.MethodInputType.AssignableTo(orderCancelsParam):
		input = reflect.ValueOf([]order.Cancel{
			{
				Exchange:  exchName,
				Type:      order.Market,
				Side:      order.Buy,
				Pair:      argGenerator.AssetParams.Pair,
				AssetType: argGenerator.AssetParams.Asset,
				OrderID:   "1337",
			},
		})
	case argGenerator.MethodInputType.AssignableTo(getOrdersRequestParam):
		input = reflect.ValueOf(&order.MultiOrderRequest{
			Type:        order.AnyType,
			Side:        order.AnySide,
			FromOrderID: "1337",
			AssetType:   argGenerator.AssetParams.Asset,
			Pairs:       currency.Pairs{argGenerator.AssetParams.Pair},
		})
	default:
		input = reflect.Zero(argGenerator.MethodInputType)
	}
	argGenerator.argNum++

	return &input
}

// CallExchangeMethod will call an exchange's method using generated arguments
// and determine if the error is friendly
func CallExchangeMethod(t *testing.T, methodToCall reflect.Value, methodValues []reflect.Value, methodName string, exch exchange.IBotExchange) {
	t.Helper()
	errType := reflect.TypeOf(common.ErrNotYetImplemented)
	outputs := methodToCall.Call(methodValues)
	for i := range outputs {
		incoming := outputs[i].Interface()
		if reflect.TypeOf(incoming) == errType {
			err, ok := incoming.(error)
			if !ok {
				t.Errorf("%s type assertion failure for %v", methodName, incoming)
				continue
			}
			if isUnacceptableError(t, err) != nil {
				literalInputs := make([]interface{}, len(methodValues))
				for j := range methodValues {
					literalInputs[j] = methodValues[j].Interface()
				}
				t.Errorf("%v Func '%v' Error: '%v'. Inputs: %v.", exch.GetName(), methodName, err, literalInputs)
			}
			break
		}
	}
}

// assetPair holds a currency pair associated with an asset
type assetPair struct {
	Pair  currency.Pair
	Asset asset.Item
}

// excludedMethodNames represent the functions that are not
// currently tested under this suite due to irrelevance
// or not worth checking yet
var excludedMethodNames = map[string]struct{}{
	"Setup":                            {}, // Is run via test setup
	"Start":                            {}, // Is run via test setup
	"SetDefaults":                      {}, // Is run via test setup
	"UpdateTradablePairs":              {}, // Is run via test setup
	"GetDefaultConfig":                 {}, // Is run via test setup
	"FetchTradablePairs":               {}, // Is run via test setup
	"GetCollateralCurrencyForContract": {}, // Not widely supported/implemented futures endpoint
	"GetCurrencyForRealisedPNL":        {}, // Not widely supported/implemented futures endpoint
	"GetFuturesPositions":              {}, // Not widely supported/implemented futures endpoint
	"GetFundingRates":                  {}, // Not widely supported/implemented futures endpoint
	"IsPerpetualFutureCurrency":        {}, // Not widely supported/implemented futures endpoint
	"GetMarginRatesHistory":            {}, // Not widely supported/implemented futures endpoint
	"CalculatePNL":                     {}, // Not widely supported/implemented futures endpoint
	"CalculateTotalCollateral":         {}, // Not widely supported/implemented futures endpoint
	"ScaleCollateral":                  {}, // Not widely supported/implemented futures endpoint
	"GetPositionSummary":               {}, // Not widely supported/implemented futures endpoint
	"AuthenticateWebsocket":            {}, // Unnecessary websocket test
	"FlushWebsocketChannels":           {}, // Unnecessary websocket test
	"UnsubscribeToWebsocketChannels":   {}, // Unnecessary websocket test
	"SubscribeToWebsocketChannels":     {}, // Unnecessary websocket test
	"GetOrderExecutionLimits":          {}, // Not widely supported/implemented feature
	"UpdateCurrencyStates":             {}, // Not widely supported/implemented feature
	"UpdateOrderExecutionLimits":       {}, // Not widely supported/implemented feature
	"CanTradePair":                     {}, // Not widely supported/implemented feature
	"CanTrade":                         {}, // Not widely supported/implemented feature
	"CanWithdraw":                      {}, // Not widely supported/implemented feature
	"CanDeposit":                       {}, // Not widely supported/implemented feature
	"GetCurrencyStateSnapshot":         {}, // Not widely supported/implemented feature
	"SetHTTPClientUserAgent":           {}, // standard base implementation
	"SetClientProxyAddress":            {}, // standard base implementation
}

var unsupportedExchangeNames = []string{
	"alphapoint",
	"bitflyer", // Bitflyer has many "ErrNotYetImplemented, which is true, but not what we care to test for here
	"bittrex",  // the api is about to expire in March, and we haven't updated it yet
	"itbit",    // itbit has no way of retrieving pair data
}

var blockedCIExchanges = []string{
	"binance", // binance API is banned from executing within the US where github Actions is ran
}

// acceptable errors do not throw test errors, see below for why
var acceptableErrors = []error{
	common.ErrFunctionNotSupported,   // Shows API cannot perform function and developer has recognised this
	asset.ErrNotSupported,            // Shows that valid and invalid asset types are handled
	request.ErrAuthRequestFailed,     // We must set authenticated requests properly in order to understand and better handle auth failures
	order.ErrUnsupportedOrderType,    // Should be returned if an ordertype like ANY is requested and the implementation knows to throw this specific error
	currency.ErrCurrencyPairEmpty,    // Demonstrates handling of EMPTYPAIR scenario and returns the correct error
	currency.ErrCurrencyNotSupported, // Ensures a standard error is used for when a particular currency/pair is not supported by an exchange
	currency.ErrCurrencyNotFound,     // Semi-randomly selected currency pairs may not be found at an endpoint, so long as this is returned it is okay
	asset.ErrNotEnabled,              // Allows distinction when checking for supported versus enabled
}

// warningErrors will t.Log(err) when thrown to diagnose things, but not necessarily suggest
// that the implementation is in error
var warningErrors = []error{
	kline.ErrNoTimeSeriesDataToConvert, // No data returned for a candle isn't worth failing the test suite over necessarily
}

// getPairFromPairs prioritises more normal pairs for an increased
// likelihood of returning data from API endpoints
func getPairFromPairs(t *testing.T, p currency.Pairs) (currency.Pair, error) {
	t.Helper()
	for i := range p {
		if p[i].Base.Equal(currency.BTC) {
			return p[i], nil
		}
	}
	for i := range p {
		if p[i].Base.Equal(currency.ETH) {
			return p[i], nil
		}
	}
	return p.GetRandomPair()
}

// isFiat helps determine fiat currency without using currency.storage
func isFiat(t *testing.T, c string) bool {
	t.Helper()
	var fiats = []string{
		currency.USD.Item.Lower,
		currency.AUD.Item.Lower,
		currency.EUR.Item.Lower,
		currency.CAD.Item.Lower,
		currency.TRY.Item.Lower,
		currency.UAH.Item.Lower,
		currency.RUB.Item.Lower,
		currency.RUR.Item.Lower,
		currency.JPY.Item.Lower,
		currency.HKD.Item.Lower,
		currency.SGD.Item.Lower,
		currency.ZUSD.Item.Lower,
		currency.ZEUR.Item.Lower,
		currency.ZCAD.Item.Lower,
		currency.ZJPY.Item.Lower,
	}
	for i := range fiats {
		if fiats[i] == c {
			return true
		}
	}
	return false
}

// disruptFormatting adds in an unused delimiter and strange casing features to
// ensure format currency pair is used throughout the code base.
func disruptFormatting(t *testing.T, p currency.Pair) (currency.Pair, error) {
	t.Helper()
	if p.Base.IsEmpty() {
		return currency.EMPTYPAIR, errors.New("cannot disrupt formatting as base is not populated")
	}
	// NOTE: Quote can be empty for margin funding
	return currency.Pair{
		Base:      p.Base.Upper(),
		Quote:     p.Quote.Lower(),
		Delimiter: "-TEST-DELIM-",
	}, nil
}
