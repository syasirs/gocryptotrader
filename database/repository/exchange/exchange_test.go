package exchange

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"
	"github.com/thrasher-corp/gocryptotrader/database/testhelpers"
)

var (
	verbose = false
)

func TestMain(m *testing.M) {
	var err error
	testhelpers.PostgresTestDatabase = testhelpers.GetConnectionDetails()
	testhelpers.TempDir, err = ioutil.TempDir("", "gct-temp")
	if err != nil {
		fmt.Printf("failed to create temp file: %v", err)
		os.Exit(1)
	}

	if verbose {
		testhelpers.EnableVerboseTestOutput()
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
		seedDB func() error
		runner func(t *testing.T)
		closer func(dbConn *database.Instance) error
	}{
		{
			name:   "postgresql",
			config: testhelpers.PostgresTestDatabase,
			seedDB: nil,
		},
		{
			name: "SQLite",
			config: &database.Config{
				Driver:            database.DBSQLite3,
				ConnectionDetails: drivers.ConnectionDetails{Database: "./testdb"},
			},
			seedDB: nil,
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			if !testhelpers.CheckValidConfig(&test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := testhelpers.ConnectToDatabase(test.config, true)
			if err != nil {
				t.Fatal(err)
			}

			if test.seedDB != nil {
				err = test.seedDB()
				if err != nil {
					t.Error(err)
				}
			}

			err = InsertMany(allExchanges)
			if err != nil {
				t.Fatal(err)
			}

			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOneAndOneByUUID(t *testing.T) {
	testCases := []struct {
		name   string
		config *database.Config
		seedDB func() error
		runner func(t *testing.T)
		closer func(dbConn *database.Instance) error
	}{
		{
			name:   "postgresql",
			config: testhelpers.PostgresTestDatabase,
			seedDB: Seed,
		},
		{
			name: "SQLite",
			config: &database.Config{
				Driver:            database.DBSQLite3,
				ConnectionDetails: drivers.ConnectionDetails{Database: "./testdb"},
			},
			seedDB: Seed,
		},
	}

	for x := range testCases {
		test := testCases[x]

		t.Run(test.name, func(t *testing.T) {
			if !testhelpers.CheckValidConfig(&test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := testhelpers.ConnectToDatabase(test.config, true)
			if err != nil {
				t.Fatal(err)
			}

			if test.seedDB != nil {
				err = test.seedDB()
				if err != nil {
					t.Error(err)
				}
			}

			ret, err := One("Binance")
			if err != nil {
				t.Fatal(err)
			}

			ret2, err := OneByUUID(ret.UUID)
			if err != nil {
				t.Fatal(err)
			}

			if ret.Name != ret2.Name {
				t.Fatalf("unexpected value received: %v", ret2.Name)
			}
			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
