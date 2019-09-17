package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"

	psqlConn "github.com/thrasher-corp/gocryptotrader/database/drivers/postgres"
	sqliteConn "github.com/thrasher-corp/gocryptotrader/database/drivers/sqlite"
)

var (
	tempDir string

	postgresTestDatabase = database.Config{
		Enabled:           true,
		Driver:            "postgres",
		ConnectionDetails: drivers.ConnectionDetails{
			//Host:     "",
			//Port:     5432,
			//Username: "",
			//Password: "",
			//Database: "",
			//SSLMode:  "",
		},
	}
)

func TestMain(m *testing.M) {
	var err error
	tempDir, err = ioutil.TempDir("", "gct-temp")
	if err != nil {
		fmt.Printf("failed to create temp file: %v", err)
		os.Exit(1)
	}

	t := m.Run()

	err = os.RemoveAll(tempDir)
	if err != nil {
		fmt.Printf("Failed to remove temp db file: %v", err)
	}

	os.Exit(t)
}

func TestDatabaseConnect(t *testing.T) {
	testCases := []struct {
		name   string
		config database.Config
		closer func(t *testing.T, dbConn *database.Db) error
		output interface{}
	}{
		{
			"SQLite",
			database.Config{
				Driver:            "sqlite",
				ConnectionDetails: drivers.ConnectionDetails{Database: path.Join(tempDir, "./testdb.db")},
			},
			closeDatabase,
			nil,
		},
		{
			"SQliteNoDatabase",
			database.Config{
				Driver: "sqlite",
				ConnectionDetails: drivers.ConnectionDetails{
					Host: "localhost",
				},
			},
			nil,
			database.ErrNoDatabaseProvided,
		},
		{
			name:   "Postgres",
			config: postgresTestDatabase,
			output: nil,
		},
	}

	for _, tests := range testCases {
		test := tests
		t.Run(test.name, func(t *testing.T) {
			if !checkValidConfig(t, &test.config.ConnectionDetails) {
				t.Skip("database not configured skipping test")
			}

			dbConn, err := connectToDatabase(t, &test.config)
			if err != nil {
				switch v := test.output.(type) {
				case error:
					if v.Error() != err.Error() {
						t.Fatal(err)
					}
					return
				default:
					break
				}
			}

			if test.closer != nil {
				err = test.closer(t, dbConn)
				if err != nil {
					t.Log(err)
				}
			}
		})
	}
}

func connectToDatabase(t *testing.T, conn *database.Config) (dbConn *database.Db, err error) {
	t.Helper()
	database.DB.Config = conn

	if conn.Driver == "postgres" {
		dbConn, err = psqlConn.Connect()
		if err != nil {
			return
		}
	} else if conn.Driver == "sqlite" {
		dbConn, err = sqliteConn.Connect()
		if err != nil {
			return
		}
	}
	database.DB.Connected = true
	return
}

func closeDatabase(t *testing.T, conn *database.Db) (err error) {
	t.Helper()

	if conn != nil {
		return conn.SQL.Close()
	}
	return nil
}

func checkValidConfig(t *testing.T, config *drivers.ConnectionDetails) bool {
	t.Helper()

	return !reflect.DeepEqual(drivers.ConnectionDetails{}, *config)
}
