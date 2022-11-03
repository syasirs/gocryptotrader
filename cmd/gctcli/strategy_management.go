package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/gctrpc"
	"github.com/urfave/cli/v2"
)

var (
	stratStartTime   string
	stratEndTime     string
	stratGranularity int64
	stratMaxImpact   float64
	stratMaxSpread   float64
	stratSimulate    bool
)

var strategyManagementCommand = &cli.Command{
	Name:      "strategy",
	Usage:     "execute strategy management command",
	ArgsUsage: "<command> <args>",
	Subcommands: []*cli.Command{
		{
			Name:        "manager",
			Usage:       "interacts with manager layer",
			ArgsUsage:   "<command> <args>",
			Subcommands: []*cli.Command{managerGetAll, managerStopAll},
		},
		{
			Name:        "twap",
			Usage:       "initiates a twap strategy to accumulate or decumulate your position",
			ArgsUsage:   "<command> <args>",
			Subcommands: []*cli.Command{twapStream},
		},
	},
}

var (
	managerGetAll = &cli.Command{
		Name:   "getall",
		Usage:  "gets all strategies",
		Action: getAllStrats,
	}
	managerStopAll = &cli.Command{
		Name:   "stopall",
		Usage:  "stops all strategies",
		Action: stopAllStrats,
	}

	twapStream = &cli.Command{
		Name:      "stream",
		Usage:     "executes strategy while reporting all actions to the client, exiting will stop strategy NOTE: cli flag might need to be used to access underyling funds e.g. --apisubaccount='main' for ftx main sub account",
		ArgsUsage: "<exchange> <pair> <asset> <start> <end>",
		Action:    twapStreamfunc,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "exchange",
				Usage: "the exchange to act on",
			},
			&cli.StringFlag{
				Name:  "pair",
				Usage: "curreny pair",
			},
			&cli.StringFlag{
				Name:  "asset",
				Usage: "asset",
			},
			&cli.BoolFlag{
				Name:        "simulate",
				Usage:       "puts the strategy in simulation mode and will not execute live orders, this is on by default",
				Value:       true,
				Destination: &stratSimulate,
			},
			&cli.StringFlag{
				Name:        "start",
				Usage:       "the start date - can be scheduled for future",
				Value:       time.Now().Format(common.SimpleTimeFormat),
				Destination: &stratStartTime,
			},
			&cli.StringFlag{
				Name:        "end",
				Usage:       "the end date",
				Value:       time.Now().Add(time.Minute * 5).Format(common.SimpleTimeFormat),
				Destination: &stratEndTime,
			},
			&cli.BoolFlag{
				Name:  "continue",
				Usage: "this will continue to deplete the required amount even after the strategy end date",
			},
			&cli.Int64Flag{
				Name:        "granularity",
				Aliases:     []string{"g"},
				Usage:       klineMessage,
				Value:       60,
				Destination: &stratGranularity,
			},
			&cli.Float64Flag{
				Name:  "amount",
				Usage: "if buying is how much quote to use, if selling is how much base to liquidate",
			},
			&cli.BoolFlag{
				Name:  "fullamount",
				Usage: "will use entire funding amount associated with the apikeys",
			},
			&cli.Float64Flag{
				Name:  "pricelimit",
				Usage: "enforces price limits if lifting the asks it will not execute an order above this price. If hitting the bids this will not execute an order below this price",
			},
			&cli.Float64Flag{
				Name:        "maximpact",
				Usage:       "will enforce no orderbook impact slippage beyond this percentage amount",
				Value:       1, // Default 1% slippage catch if not set.
				Destination: &stratMaxImpact,
			},
			&cli.Float64Flag{
				Name:  "maxnominal",
				Usage: "will enforce no orderbook nominal (your average order cost from initial order cost) slippage beyond this percentage amount",
			},
			&cli.BoolFlag{
				Name:  "buy",
				Usage: "whether you are buying base or selling base",
			},
			&cli.Float64Flag{
				Name:        "maxspread",
				Usage:       "will enforce no orderbook spread percentage beyond this amount. If there is massive spread it usually means liquidity issues",
				Value:       1, // Default 1% spread catch if not set.
				Destination: &stratMaxSpread,
			},
		},
	}
)

func getAllStrats(c *cli.Context) error {
	return nil
}

func stopAllStrats(c *cli.Context) error {
	return nil
}

func twapStreamfunc(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchangeName string
	if c.IsSet("exchange") {
		exchangeName = c.String("exchange")
	} else {
		exchangeName = c.Args().First()
	}

	var pair string
	if c.IsSet("pair") {
		pair = c.String("pair")
	} else {
		pair = c.Args().Get(1)
	}

	cp, err := currency.NewPairDelimiter(pair, pairDelimiter)
	if err != nil {
		return err
	}

	var assetType string
	if c.IsSet("asset") {
		assetType = c.String("asset")
	} else {
		assetType = c.Args().Get(2)
	}

	if !validAsset(assetType) {
		return errInvalidAsset
	}

	if c.IsSet("simulate") {
		stratSimulate = c.Bool("simulate")
	} else {
		var arg bool
		arg, err = strconv.ParseBool(c.Args().Get(3))
		if err == nil {
			stratSimulate = arg
		}
	}

	if !c.IsSet("start") {
		if c.Args().Get(4) != "" {
			stratStartTime = c.Args().Get(4)
		}
	} else {
		stratStartTime, _ = c.Value("start").(string)
	}

	s, err := time.Parse(common.SimpleTimeFormat, stratStartTime)
	if err != nil {
		return fmt.Errorf("invalid time format for start: %v", err)
	}

	if !c.IsSet("end") {
		if c.Args().Get(5) != "" {
			stratEndTime = c.Args().Get(5)
		}
	} else {
		stratEndTime, _ = c.Value("end").(string)
	}

	e, err := time.Parse(common.SimpleTimeFormat, stratEndTime)
	if err != nil {
		return fmt.Errorf("invalid time format for end: %v", err)
	}

	err = common.StartEndTimeCheck(s, e)
	if err != nil && !errors.Is(err, common.ErrStartAfterTimeNow) {
		return err
	}

	var continueAfterEnd bool
	if c.IsSet("continue") {
		continueAfterEnd = c.Bool("continue")
	} else {
		continueAfterEnd, _ = strconv.ParseBool(c.Args().Get(6))
	}

	if c.IsSet("granularity") {
		stratGranularity = c.Int64("granularity")
	} else if c.Args().Get(6) != "" {
		stratGranularity, err = strconv.ParseInt(c.Args().Get(7), 10, 64)
		if err != nil {
			return err
		}
	}

	var amount float64
	if c.IsSet("amount") {
		amount = c.Float64("amount")
	} else if c.Args().Get(7) != "" {
		amount, err = strconv.ParseFloat(c.Args().Get(8), 64)
		if err != nil {
			return err
		}
	}

	var fullAmount bool
	if c.IsSet("fullamount") {
		fullAmount = c.Bool("fullamount")
	} else {
		fullAmount, _ = strconv.ParseBool(c.Args().Get(9))
	}

	var priceLimit float64
	if c.IsSet("pricelimit") {
		priceLimit = c.Float64("pricelimit")
	} else if c.Args().Get(7) != "" {
		priceLimit, err = strconv.ParseFloat(c.Args().Get(10), 64)
		if err != nil {
			return err
		}
	}

	if c.IsSet("maximpact") {
		stratMaxImpact = c.Float64("maximpact")
	} else if c.Args().Get(7) != "" {
		stratMaxImpact, err = strconv.ParseFloat(c.Args().Get(11), 64)
		if err != nil {
			return err
		}
	}

	var maxNominal float64
	if c.IsSet("maxnominal") {
		maxNominal = c.Float64("maxnominal")
	} else if c.Args().Get(7) != "" {
		maxNominal, err = strconv.ParseFloat(c.Args().Get(12), 64)
		if err != nil {
			return err
		}
	}

	if stratMaxImpact <= 0 && maxNominal <= 0 {
		// Protection for user without any slippage protection if a large amount
		// on a non-liquid book was to be deployed.
		log.Println("Warning: No slippage protection on strategy run, this can have dire consequences. Continue (y/n)?")
		input := ""
		if _, err := fmt.Scanln(&input); err != nil {
			return err
		}
		if !common.YesOrNo(input) {
			return nil
		}
	}

	var buy bool
	if c.IsSet("buy") {
		buy = c.Bool("buy")
	} else {
		buy, _ = strconv.ParseBool(c.Args().Get(13))
	}

	if c.IsSet("maxspread") {
		stratMaxSpread = c.Float64("maxspread")
	} else if c.Args().Get(7) != "" {
		stratMaxSpread, err = strconv.ParseFloat(c.Args().Get(14), 64)
		if err != nil {
			return err
		}
	}

	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	client := gctrpc.NewGoCryptoTraderServiceClient(conn)
	result, err := client.TWAPStream(c.Context, &gctrpc.TWAPRequest{
		Exchange: exchangeName,
		Pair: &gctrpc.CurrencyPair{
			Base:  cp.Base.String(),
			Quote: cp.Quote.String(),
		},
		Simulate:            stratSimulate,
		Asset:               assetType,
		Start:               negateLocalOffsetTS(s),
		End:                 negateLocalOffsetTS(e),
		AllowTradingPastEnd: continueAfterEnd,
		Interval:            stratGranularity * int64(time.Second),
		Amount:              amount,
		FullAmount:          fullAmount,
		PriceLimit:          priceLimit,
		MaxImpactSlippage:   stratMaxImpact,
		MaxNominalSlippage:  maxNominal,
		Buy:                 buy,
		MaxSpreadPercentage: stratMaxSpread,
	})
	if err != nil {
		return err
	}

	for {
		resp, err := result.Recv()
		if err != nil {
			return err
		}

		jsonOutput(resp)
		if resp.Finished {
			return nil
		}
	}
}
