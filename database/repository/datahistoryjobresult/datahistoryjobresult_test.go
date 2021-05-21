package datahistoryjobresult

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/drivers"
	"github.com/thrasher-corp/gocryptotrader/database/repository/exchange"
	"github.com/thrasher-corp/gocryptotrader/database/testhelpers"
)

var (
	verbose       = false
	testExchanges = []exchange.Details{
		{
			Name: "one",
		},
		{
			Name: "two",
		},
	}
	db = &DBService{}
)

func TestMain(m *testing.M) {
	if verbose {
		testhelpers.EnableVerboseTestOutput()
	}
	//var err error
	testhelpers.PostgresTestDatabase = testhelpers.GetConnectionDetails()
	//testhelpers.TempDir, err = ioutil.TempDir("", "gct-temp")
	//if err != nil {
	//	log.Fatal(err)
	//}
	t := m.Run()
	//err = os.RemoveAll(testhelpers.TempDir)
	//if err != nil {
	//	fmt.Printf("Failed to remove temp db file: %v", err)
	//}

	os.Exit(t)
}

func seedDB() error {
	err := exchange.InsertMany(testExchanges)
	if err != nil {
		return err
	}

	for i := range testExchanges {
		lol, err := exchange.One(testExchanges[i].Name)
		if err != nil {
			return err
		}
		testExchanges[i].UUID = lol.UUID
	}

	return nil
}

func TestDataHistoryJob(t *testing.T) {
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
				err = test.seedDB()
				if err != nil {
					t.Error(err)
				}
			}

			db, err := Setup(dbConn)
			if err != nil {
				log.Fatal(err)
			}

			var resulterinos, resultaroos []*DataHistoryJobResult
			for i := 0; i < 20; i++ {
				uu, _ := uuid.NewV4()
				resulterinos = append(resulterinos, &DataHistoryJobResult{
					ID:                uu.String(),
					JobID:             uu.String(),
					IntervalStartDate: time.Now(),
					IntervalEndDate:   time.Now().Add(time.Second),
					Status:            0,
					Result:            "Yay",
					Date:              time.Now(),
				})
			}
			err = db.Upsert(resulterinos...)
			if err != nil {
				t.Fatal(err)
			}
			jobID, _ := uuid.NewV4()
			// insert the same results to test conflict resolution
			for i := 0; i < 20; i++ {
				uu, _ := uuid.NewV4()
				j := &DataHistoryJobResult{
					ID:                uu.String(),
					JobID:             jobID.String(),
					IntervalStartDate: time.Now(),
					IntervalEndDate:   time.Now().Add(time.Second),
					Status:            0,
					Result:            "Wow",
					Date:              time.Now(),
				}
				if i == 19 {
					j.Status = 1
					j.Date = time.Now().Add(time.Hour * 24)
				}
				resultaroos = append(resultaroos, j)
			}
			err = db.Upsert(resultaroos...)
			if err != nil {
				t.Fatal(err)
			}

			results, err := db.GetByJobID(jobID.String())
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 20 {
				t.Error("expected 20 job results")
			}

			results, err = db.GetJobResultsBetween(jobID.String(), time.Now().Add(time.Hour*23), time.Now().Add(time.Hour*25))
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Errorf("expected 1 job result, received %v", len(results))
			}

			err = testhelpers.CloseDatabase(dbConn)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
