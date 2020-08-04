package candle

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"
	"github.com/thrasher-corp/gocryptotrader/database/repository/exchange"
	"github.com/thrasher-corp/gocryptotrader/database/testhelpers"
)

var (
	verbose       = false
	testExchanges = []exchange.Details{
		{
			Name: "One",
		},
		{
			Name: "Two",
		},
	}
)

func TestMain(m *testing.M) {
	if verbose {
		testhelpers.EnableVerboseTestOutput()
	}

	var err error
	testhelpers.PostgresTestDatabase = testhelpers.GetConnectionDetails()
	testhelpers.TempDir, err = ioutil.TempDir("", "gct-temp")
	if err != nil {
		fmt.Printf("failed to create temp file: %v", err)
		os.Exit(1)
	}

	t := m.Run()

	err = os.RemoveAll(testhelpers.TempDir)
	if err != nil {
		fmt.Printf("Failed to remove temp db file: %v", err)
	}

	os.Exit(t)
}

func TestInsert(t *testing.T) {
	testCases := []struct {
		name   string
		config *database.Config
		seedDB func(includeOHLCVData bool) error
		runner func(t *testing.T)
		closer func(dbConn *database.Instance) error
	}{
		{
			name:   "postgresql",
			config: testhelpers.PostgresTestDatabase,
			seedDB: seedDB,
		},
		{
			name: "SQLite",
			config: &database.Config{
				Driver:            database.DBSQLite3,
				ConnectionDetails: drivers.ConnectionDetails{Database: "./testdb"},
			},
			seedDB: seedDB,
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			if !testhelpers.CheckValidConfig(&test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := testhelpers.ConnectToDatabase(test.config)
			if err != nil {
				t.Fatal(err)
			}

			if test.seedDB != nil {
				err = test.seedDB(false)
				if err != nil {
					t.Fatal(err)
				}
			}

			data, err := genOHCLVData()
			if err != nil {
				t.Fatal(err)
			}

			r, err := Insert(&data)
			if err != nil {
				t.Fatal(err)
			}

			if r != 365 {
				t.Fatalf("unexpected number inserted: %v", r)
			}
			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestInsertFromCSV(t *testing.T) {
	testCases := []struct {
		name   string
		config *database.Config
		seedDB func(includeOHLCVData bool) error
		runner func(t *testing.T)
		closer func(dbConn *database.Instance) error
	}{
		{
			name:   "postgresql",
			config: testhelpers.PostgresTestDatabase,
			seedDB: seedDB,
		},
		{
			name: "SQLite",
			config: &database.Config{
				Driver:            database.DBSQLite3,
				ConnectionDetails: drivers.ConnectionDetails{Database: "./testdb"},
			},
			seedDB: seedDB,
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			if !testhelpers.CheckValidConfig(&test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := testhelpers.ConnectToDatabase(test.config)
			if err != nil {
				t.Fatal(err)
			}

			if test.seedDB != nil {
				err = test.seedDB(false)
				if err != nil {
					t.Fatal(err)
				}
			}

			exchange.ResetExchangeCache()
			testFile := filepath.Join("..", "..", "..", "testdata", "binance_BTCUSDT_24h_2019_01_01_2020_01_01.csv")
			count, err := InsertFromCSV(testExchanges[0].Name, "BTC", "USDT", "24h", "spot", testFile)
			if err != nil {
				t.Fatal(err)
			}
			if count != 365 {
				t.Fatalf("expected 365 results to be inserted received: %v", count)
			}

			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSeries(t *testing.T) {
	testCases := []struct {
		name   string
		config *database.Config
		seedDB func(includeOHLCVData bool) error
		runner func(t *testing.T)
		closer func(dbConn *database.Instance) error
	}{
		{
			name:   "postgresql",
			config: testhelpers.PostgresTestDatabase,
			seedDB: seedDB,
		},
		{
			name: "SQLite",
			config: &database.Config{
				Driver:            database.DBSQLite3,
				ConnectionDetails: drivers.ConnectionDetails{Database: "./testdb"},
			},
			seedDB: seedDB,
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			if !testhelpers.CheckValidConfig(&test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := testhelpers.ConnectToDatabase(test.config)
			if err != nil {
				t.Fatal(err)
			}

			if test.seedDB != nil {
				err = test.seedDB(true)
				if err != nil {
					t.Fatal(err)
				}
			}

			start := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
			end := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			ret, err := Series(testExchanges[0].Name, "BTC", "USDT", "24h", "spot", start, end)
			if err != nil {
				t.Fatal(err)
			}
			if len(ret.Candles) != 365 {
				t.Errorf("unexpected number of results received:  %v", len(ret.Candles))
			}

			ret, err = Series("", "", "", "", "", start, end)
			if err != nil {
				if !errors.Is(err, errInvalidInput) {
					t.Fatal(err)
				}
			}

			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func seedDB(includeOHLCVData bool) error {
	err := exchange.InsertMany(testExchanges)
	if err != nil {
		return err
	}

	if includeOHLCVData {
		exchange.ResetExchangeCache()
		data, err := genOHCLVData()
		if err != nil {
			return err
		}
		_, err = Insert(&data)
		return err
	}
	return nil
}

func genOHCLVData() (out Item, err error) {
	exchangeUUID, err := exchange.UUIDByName(testExchanges[0].Name)
	if err != nil {
		return
	}
	out.ExchangeID = exchangeUUID.String()
	out.Base = currency.BTC.String()
	out.Quote = currency.USDT.String()
	out.Interval = "24h"
	out.Asset = "spot"

	start := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	for x := 0; x < 365; x++ {
		out.Candles = append(out.Candles, Candle{
			Timestamp: start.Add(time.Hour * 24 * time.Duration(x)),
			Open:      1000,
			High:      1000,
			Low:       1000,
			Close:     1000,
			Volume:    1000,
		})
	}

	return out, nil
}
