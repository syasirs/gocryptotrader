package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/communications"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/currency/forexprovider"
	"github.com/thrasher-/gocryptotrader/database"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/portfolio"
)

// Bot contains configuration, portfolio, exchange & ticker data and is the
// overarching type across this code base.
type Bot struct {
	config     *config.Config
	portfolio  *portfolio.Base
	exchanges  []exchange.IBotExchange
	comms      *communications.Communications
	db         *database.ORM
	shutdown   chan bool
	dryRun     bool
	configFile string
}

const (
	dbname     = "gocryptotrader"
	dbhost     = "localhost"
	dbuser     = "gocryptotrader"
	dbpassword = "gocryptotrader"
	banner     = `
   ______        ______                     __        ______                  __
  / ____/____   / ____/_____ __  __ ____   / /_ ____ /_  __/_____ ______ ____/ /___   _____
 / / __ / __ \ / /    / ___// / / // __ \ / __// __ \ / /  / ___// __  // __  // _ \ / ___/
/ /_/ // /_/ // /___ / /   / /_/ // /_/ // /_ / /_/ // /  / /   / /_/ // /_/ //  __// /
\____/ \____/ \____//_/    \__, // .___/ \__/ \____//_/  /_/    \__,_/ \__,_/ \___//_/
                          /____//_/
`
)

var bot Bot

func main() {
	bot.shutdown = make(chan bool)
	HandleInterrupt()

	defaultPath, err := config.GetFilePath("")
	if err != nil {
		log.Fatal(err)
	}

	// Handle flags
	flag.StringVar(&bot.configFile, "config", defaultPath, "-config sets filepath to load configuration")
	dryrun := flag.Bool("dryrun", false, "-dryrun does not save config.json file")
	version := flag.Bool("version", false, "-version retrieves current GoCryptoTrader version")
	verbosity := flag.Bool("verbose", false, "-verbose increases verbosity for GoCryptoTrader")

	dbStart := flag.Bool("db", false, "-db connects to a postgres database")
	dbSeedHistory := flag.Bool("seeddb", false, "-seeddb aggregates historic exchange trade data into the database")
	dbName := flag.String("dbname", dbname, "-dbname changes database name")
	dbHost := flag.String("dbhost", dbhost, "-dbhost changes database host details")
	dbUser := flag.String("dbuser", dbuser, "-dbuser changes database username")
	dbPassword := flag.String("dbpassword", dbpassword, "-dbpassword changes database password")

	flag.Parse()

	if *version {
		fmt.Printf(BuildVersion(true))
		os.Exit(0)
	}

	if *dryrun {
		bot.dryRun = true
	}

	bot.config = &config.Cfg
	fmt.Println(banner)
	fmt.Println(BuildVersion(false))
	log.Printf("Loading config file %s..\n", bot.configFile)

	err = bot.config.LoadConfig(bot.configFile)
	if err != nil {
		log.Fatalf("Failed to load config. Err: %s", err)
	}

	AdjustGoMaxProcs()
	log.Printf("Bot '%s' started.\n", bot.config.Name)
	log.Printf("Bot dry run mode: %v.\n", common.IsEnabled(bot.dryRun))

	log.Printf("Available Exchanges: %d. Enabled Exchanges: %d.\n",
		len(bot.config.Exchanges),
		bot.config.CountEnabledExchanges())

	common.HTTPClient = common.NewHTTPClientWithTimeout(bot.config.GlobalHTTPTimeout)
	log.Printf("Global HTTP request timeout: %v.\n", common.HTTPClient.Timeout)

	SetupExchanges()
	if len(bot.exchanges) == 0 {
		log.Fatalf("No exchanges were able to be loaded. Exiting")
	}

	log.Println("Starting communication mediums..")
	bot.comms = communications.NewComm(bot.config.GetCommunicationsConfig())
	bot.comms.GetEnabledCommunicationMediums()

	log.Printf("Fiat display currency: %s.", bot.config.Currency.FiatDisplayCurrency)
	currency.BaseCurrency = bot.config.Currency.FiatDisplayCurrency
	currency.FXProviders = forexprovider.StartFXService(bot.config.GetCurrencyConfig().ForexProviders)
	log.Printf("Primary forex conversion provider: %s.\n", bot.config.GetPrimaryForexProvider())
	err = bot.config.RetrieveConfigCurrencyPairs(true)
	if err != nil {
		log.Fatalf("Failed to retrieve config currency pairs. Error: %s", err)
	}
	log.Println("Successfully retrieved config currencies.")
	log.Println("Fetching currency data from forex provider..")
	err = currency.SeedCurrencyData(common.JoinStrings(currency.FiatCurrencies, ","))
	if err != nil {
		log.Fatalf("Unable to fetch forex data. Error: %s", err)
	}

	bot.portfolio = &portfolio.Portfolio
	bot.portfolio.SeedPortfolio(bot.config.Portfolio)
	SeedExchangeAccountInfo(GetAllEnabledExchangeAccountInfo().Data)

	go portfolio.StartPortfolioWatcher()
	go TickerUpdaterRoutine()
	go OrderbookUpdaterRoutine()

	if *dbStart {
		log.Println("Opening connection to database....")
		bot.db, err = database.NewORMConnection(*dbName, *dbHost, *dbUser, *dbPassword, *verbosity)
		if err != nil {
			log.Fatal(err)
		}
		err = bot.db.LoadConfiguration(bot.config.Name)
		if err != nil {
			err = bot.db.InsertNewConfiguration(bot.config, "newPassword")
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Println("Database connection established")
		if *dbSeedHistory {
			go HistoricExchangeDataUpdaterRoutine()
		}
	}

	if bot.config.Webserver.Enabled {
		listenAddr := bot.config.Webserver.ListenAddress
		log.Printf(
			"HTTP Webserver support enabled. Listen URL: http://%s:%d/\n",
			common.ExtractHost(listenAddr), common.ExtractPort(listenAddr),
		)

		router := NewRouter(bot.exchanges)
		go func() {
			err = http.ListenAndServe(listenAddr, router)
			if err != nil {
				log.Fatal(err)
			}
		}()

		log.Println("HTTP Webserver started successfully.")
		log.Println("Starting websocket handler.")
		StartWebsocketHandler()
	} else {
		log.Println("HTTP RESTful Webserver support disabled.")
	}

	<-bot.shutdown
	Shutdown()
}

// AdjustGoMaxProcs adjusts the maximum processes that the CPU can handle.
func AdjustGoMaxProcs() {
	log.Println("Adjusting bot runtime performance..")
	maxProcsEnv := os.Getenv("GOMAXPROCS")
	maxProcs := runtime.NumCPU()
	log.Println("Number of CPU's detected:", maxProcs)

	if maxProcsEnv != "" {
		log.Println("GOMAXPROCS env =", maxProcsEnv)
		env, err := strconv.Atoi(maxProcsEnv)
		if err != nil {
			log.Println("Unable to convert GOMAXPROCS to int, using", maxProcs)
		} else {
			maxProcs = env
		}
	}
	if i := runtime.GOMAXPROCS(maxProcs); i != maxProcs {
		log.Fatal("Go Max Procs were not set correctly.")
	}
	log.Println("Set GOMAXPROCS to:", maxProcs)
}

// HandleInterrupt monitors and captures the SIGTERM in a new goroutine then
// shuts down bot
func HandleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Printf("Captured %v, shutdown requested.", sig)
		bot.shutdown <- true
	}()
}

// Shutdown correctly shuts down bot saving configuration files
func Shutdown() {
	log.Println("Bot shutting down..")

	if len(portfolio.Portfolio.Addresses) != 0 {
		bot.config.Portfolio = portfolio.Portfolio
	}

	if !bot.dryRun {
		err := bot.config.SaveConfig(bot.configFile)

		if err != nil {
			log.Println("Unable to save config.")
		} else {
			log.Println("Config file saved successfully.")
		}
	}

	log.Println("Exiting.")
	os.Exit(0)
}
