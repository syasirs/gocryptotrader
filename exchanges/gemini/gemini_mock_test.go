//+build !mock_test_off

// This will build if build tag mock_test_off is not parsed and will try to mock
// all tests in _test.go
package gemini

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
	geminiConfig, err := cfg.GetExchangeConfig("Gemini")
	if err != nil {
		log.Error("Test Failed - Mock server error", err)
		os.Exit(1)
	}
	geminiConfig.AuthenticatedAPISupport = true
	geminiConfig.APIKey = apiKey
	geminiConfig.APISecret = apiSecret
	g.SetDefaults()
	g.Setup(&geminiConfig)

	serverDetails, err := mock.NewVCRServer("../../testdata/http_mock/gemini/gemini.json")
	if err != nil {
		log.Warnf("Test Failed - Mock server error %s", err)
	} else {
		g.APIUrl = serverDetails
	}

	log.Debugf(sharedtestvalues.MockTesting, g.GetName(), g.APIUrl)
	os.Exit(m.Run())
}
