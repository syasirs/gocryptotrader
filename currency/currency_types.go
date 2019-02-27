package currency

import (
	"time"

	"github.com/thrasher-/gocryptotrader/currency/coinmarketcap"
)

// MainConfiguration is the main configuration from the config.json file
type MainConfiguration struct {
	ForexProviders         []FXSettings
	CryptocurrencyProvider coinmarketcap.Settings
	Cryptocurrencies       Currencies
	CurrencyPairFormat     interface{}
	FiatDisplayCurrency    Code
}

// BotOverrides defines a bot overriding factor for quick running currency
// subsystems
type BotOverrides struct {
	Coinmarketcap       bool
	FxCurrencyConverter bool
	FxCurrencyLayer     bool
	FxFixer             bool
	FxOpenExchangeRates bool
}

// AnalysisData defines information pertaining to exchange, cryptocurrency or
// fiat from coinmarketcap
type AnalysisData struct {
	ID          int
	Name        string
	Symbol      string `json:",omitempty"`
	Slug        string
	Active      bool
	LastUpdated time.Time
}

// CoinmarketcapSettings refers to settings
type CoinmarketcapSettings coinmarketcap.Settings

// SystemsSettings defines incoming system settings
type SystemsSettings struct {
	Coinmarketcap     coinmarketcap.Settings
	Currencyconverter FXSettings
	Currencylayer     FXSettings
	Fixer             FXSettings
	Openexchangerates FXSettings
}

// FXSettings defines foreign exchange requester settings
type FXSettings struct {
	Name             string        `json:"name"`
	Enabled          bool          `json:"enabled"`
	Verbose          bool          `json:"verbose"`
	RESTPollingDelay time.Duration `json:"restPollingDelay"`
	APIKey           string        `json:"apiKey"`
	APIKeyLvl        int           `json:"apiKeyLvl"`
	PrimaryProvider  bool          `json:"primaryProvider"`
}
