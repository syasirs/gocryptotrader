//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package anx

import (
	"os"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/exchanges/mock"
	"github.com/thrasher-corp/gocryptotrader/exchanges/sharedtestvalues"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

var mockTests = true

func TestMain(m *testing.M) {
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	anxConfig, err := cfg.GetExchangeConfig("ANX")
	if err != nil {
		log.Error("Test Failed - Mock server error", err)
		os.Exit(1)
	}
	anxConfig.AuthenticatedAPISupport = true
	anxConfig.APIKey = apiKey
	anxConfig.APISecret = apiSecret
	a.SetDefaults()
	a.Setup(&anxConfig)

	serverDetails, err := mock.NewVCRServer("../../testdata/http_mock/anx/anx.json")
	if err != nil {
		log.Warn("Test Failed - Mock server error", err)
	} else {
		a.APIUrl = serverDetails + "/"
	}

	log.Debugf(sharedtestvalues.MockTesting, a.GetName(), a.APIUrl)
	os.Exit(m.Run())
}
