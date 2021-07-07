package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/gctrpc"
	"github.com/urfave/cli/v2"
)

var (
	maxRetryAttempts, requestSizeLimit, batchSize, comparisonDecimalPlaces uint64
	prerequisiteJobSubCommands                                             = []cli.Flag{
		&cli.StringFlag{
			Name:  "nickname",
			Usage: "binance-spot-btc-usdt-2019-trades",
		},
		&cli.StringFlag{
			Name:  "prerequisite",
			Usage: "binance-spot-btc-usdt-2018-trades",
		},
	}
	guidExample            = "deadbeef-dead-beef-dead-beef13371337"
	specificJobSubCommands = []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: guidExample,
		},
		&cli.StringFlag{
			Name:  "nickname",
			Usage: "binance-spot-btc-usdt-2019-trades",
		},
	}
	baseJobSubCommands = []cli.Flag{
		&cli.StringFlag{
			Name:     "nickname",
			Usage:    "binance-spot-btc-usdt-2019-trades",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "exchange",
			Usage:    "binance",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "asset",
			Usage:    "spot",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "pair",
			Usage:    "btc-usdt",
			Required: true,
		},
		&cli.StringFlag{
			Name:        "start_date",
			Usage:       "formatted as: 2006-01-02 15:04:05",
			Value:       time.Now().AddDate(-1, 0, 0).Format(common.SimpleTimeFormat),
			Destination: &startTime,
		},
		&cli.StringFlag{
			Name:        "end_date",
			Usage:       "formatted as: 2006-01-02 15:04:05",
			Value:       time.Now().AddDate(0, -1, 0).Format(common.SimpleTimeFormat),
			Destination: &endTime,
		},
		&cli.Uint64Flag{
			Name:     "interval",
			Usage:    klineMessage,
			Required: true,
		},
		&cli.Uint64Flag{
			Name:        "request_size_limit",
			Usage:       "the number of candles to retrieve per API request",
			Destination: &requestSizeLimit,
			Value:       50,
		},
		&cli.Uint64Flag{
			Name:        "max_retry_attempts",
			Usage:       "the maximum retry attempts for an interval period before giving up",
			Value:       3,
			Destination: &maxRetryAttempts,
		},
		&cli.Uint64Flag{
			Name:        "batch_size",
			Usage:       "the amount of API calls to make per run",
			Destination: &batchSize,
			Value:       3,
		},
		&cli.StringFlag{
			Name:  "prerequisite",
			Usage: "optional - adds or updates the job to have a prerequisite, will only run when prerequisite job is complete - use command 'removeprerequisite' to remove a prerequisite",
		},
		&cli.BoolFlag{
			Name:  "upsert",
			Usage: "if true, will update an existing job if the nickname is shared. if false, will reject a job if the nickname already exists",
		},
	}
	retrievalJobSubCommands = []cli.Flag{
		&cli.BoolFlag{
			Name:  "overwrite_existing_data",
			Usage: "will process and overwrite data if matching data exists at an interval period. if false, will not process or save data",
		},
	}
	conversionJobSubCommands = []cli.Flag{
		&cli.BoolFlag{
			Name:  "overwrite_existing_data",
			Usage: "will process and overwrite data if matching data exists at an interval period. if false, will not process or save data",
		},
		&cli.Uint64Flag{
			Name:     "conversion_interval",
			Usage:    "data will be converted and saved at this interval",
			Required: true,
		},
	}
	validationJobSubCommands = []cli.Flag{
		&cli.Uint64Flag{
			Name:        "comparison_decimal_places",
			Usage:       "the number of decimal places used to compare against API data for accuracy",
			Destination: &comparisonDecimalPlaces,
			Value:       3,
		},
	}
)

var dataHistoryJobCommands = &cli.Command{
	Name:      "addjob",
	Usage:     "add or update data history jobs",
	ArgsUsage: "<command> <args>",
	Subcommands: []*cli.Command{
		{
			Name:   "savecandles",
			Usage:  "will fetch candle data from an exchange and save it to the database",
			Flags:  append(baseJobSubCommands, retrievalJobSubCommands...),
			Action: upsertDataHistoryJob,
		},
		{
			Name:   "convertcandles",
			Usage:  "convert candles saved to the database to a new resolution eg 1min -> 5min",
			Flags:  append(baseJobSubCommands, retrievalJobSubCommands...),
			Action: upsertDataHistoryJob,
		},
		{
			Name:   "savetrades",
			Usage:  "will fetch trade data from an exchange and save it to the database",
			Flags:  append(baseJobSubCommands, conversionJobSubCommands...),
			Action: upsertDataHistoryJob,
		},
		{
			Name:   "converttrades",
			Usage:  "convert trades saved to the database to any candle resolution eg 30min",
			Flags:  append(baseJobSubCommands, conversionJobSubCommands...),
			Action: upsertDataHistoryJob,
		},
		{
			Name:   "validatecandles",
			Usage:  "will compare database candle data with API candle data - useful for validating converted trades and candles",
			Flags:  append(baseJobSubCommands, validationJobSubCommands...),
			Action: upsertDataHistoryJob,
		},
	},
}

var dataHistoryCommands = &cli.Command{
	Name:      "datahistory",
	Usage:     "manage data history jobs to retrieve historic trade or candle data over time",
	ArgsUsage: "<command> <args>",
	Subcommands: []*cli.Command{
		{
			Name:   "getactivejobs",
			Usage:  "returns all jobs that are currently active",
			Flags:  []cli.Flag{},
			Action: getActiveDataHistoryJobs,
		},
		{
			Name:  "getjobsbetweendates",
			Usage: "returns all jobs with creation dates between the two provided dates",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "start_date",
					Usage: "formatted as: 2006-01-02 15:04:05",
				},
				&cli.StringFlag{
					Name:  "end_date",
					Usage: "formatted as: 2006-01-02 15:04:05",
				},
			},
			Action: getDataHistoryJobsBetween,
		},
		{
			Name:        "getajob",
			Usage:       "returns a job by either its id or nickname",
			Description: "na-na, why don't you get a job?",
			ArgsUsage:   "<id> or <nickname>",
			Action:      getDataHistoryJob,
			Flags:       specificJobSubCommands,
		},
		{
			Name:        "getjobwithdetailedresults",
			Usage:       "returns a job by either its nickname along with all its data retrieval results",
			Description: "results may be large",
			ArgsUsage:   "<nickname>",
			Action:      getDataHistoryJob,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "nickname",
					Usage: "binance-spot-btc-usdt-2019-trades",
				},
			},
		},
		{
			Name:      "getjobstatussummary",
			Usage:     "returns a job with human readable summary of its status",
			ArgsUsage: "<nickname>",
			Action:    getDataHistoryJobSummary,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "nickname",
					Usage: "binance-spot-btc-usdt-2019-trades",
				},
			},
		},
		dataHistoryJobCommands,
		{
			Name:      "deletejob",
			Usage:     "sets a jobs status to deleted so it no longer is processed",
			ArgsUsage: "<id> or <nickname>",
			Flags:     specificJobSubCommands,
			Action:    setDataHistoryJobStatus,
		},
		{
			Name:      "pausejob",
			Usage:     "sets a jobs status to paused so it no longer is processed",
			ArgsUsage: "<id> or <nickname>",
			Flags:     specificJobSubCommands,
			Action:    setDataHistoryJobStatus,
		},
		{
			Name:      "unpausejob",
			Usage:     "sets a jobs status to active so it can be processed",
			ArgsUsage: "<id> or <nickname>",
			Flags:     specificJobSubCommands,
			Action:    setDataHistoryJobStatus,
		},
		{
			Name:      "updateprerequisite",
			Usage:     "adds or updates a prerequisite job to the job referenced - if the job is active, it will be set as 'paused'",
			ArgsUsage: "<prerequisite> <nickname>",
			Flags:     prerequisiteJobSubCommands,
			Action:    setPrerequisiteJob,
		},
		{
			Name:      "removeprerequisite",
			Usage:     "removes a prerequisite job from the job referenced - if the job is 'paused', it will be set as 'active'",
			ArgsUsage: "<nickname>",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "nickname",
					Usage: "binance-spot-btc-usdt-2019-trades",
				},
			},
			Action: setPrerequisiteJob,
		},
	},
}

func getDataHistoryJob(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	var id string
	if c.IsSet("id") {
		id = c.String("id")
	} else {
		id = c.Args().First()
	}
	var nickname string
	if c.IsSet("nickname") {
		nickname = c.String("nickname")
	}

	if nickname != "" && id != "" {
		return errors.New("can only set 'id' OR 'nickname'")
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()
	client := gctrpc.NewGoCryptoTraderClient(conn)
	request := &gctrpc.GetDataHistoryJobDetailsRequest{
		Id:       id,
		Nickname: nickname,
	}
	if strings.EqualFold(c.Command.Name, "getjobwithdetailedresults") {
		request.FullDetails = true
	}

	result, err := client.GetDataHistoryJobDetails(context.Background(), request)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func getActiveDataHistoryJobs(_ *cli.Context) error {
	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.GetActiveDataHistoryJobs(context.Background(),
		&gctrpc.GetInfoRequest{})
	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func upsertDataHistoryJob(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	var (
		err                                 error
		nickname, exchange, assetType, pair string
		interval, dataType                  int64
	)
	if c.IsSet("nickname") {
		nickname = c.String("nickname")
	}

	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	}
	if !validExchange(exchange) {
		return errInvalidExchange
	}

	if c.IsSet("asset") {
		assetType = c.String("asset")
	}
	if !validAsset(assetType) {
		return errInvalidAsset
	}

	if c.IsSet("pair") {
		pair = c.String("pair")
	}
	if !validPair(pair) {
		return errInvalidPair
	}
	p, err := currency.NewPairDelimiter(pair, pairDelimiter)
	if err != nil {
		return fmt.Errorf("cannot process pair: %w", err)
	}

	if c.IsSet("start_date") {
		startTime = c.String("start_date")
	}
	if c.IsSet("end_date") {
		endTime = c.String("end_date")
	}

	var s, e time.Time
	s, err = time.Parse(common.SimpleTimeFormat, startTime)
	if err != nil {
		return fmt.Errorf("invalid time format for start: %v", err)
	}
	e, err = time.Parse(common.SimpleTimeFormat, endTime)
	if err != nil {
		return fmt.Errorf("invalid time format for end: %v", err)
	}

	if c.IsSet("interval") {
		interval = c.Int64("interval")
	}
	candleInterval := time.Duration(interval) * time.Second
	if c.IsSet("request_size_limit") {
		requestSizeLimit = c.Uint64("request_size_limit")
	}

	if c.IsSet("max_retry_attempts") {
		maxRetryAttempts = c.Uint64("max_retry_attempts")
	}

	if c.IsSet("batch_size") {
		batchSize = c.Uint64("batch_size")
	}
	var upsert bool
	if c.IsSet("upsert") {
		upsert = c.Bool("upsert")
	}

	switch c.Command.Name {
	case "savecandles":
		dataType = 0
	case "savetrades":
		dataType = 1
	case "convertcandles":
		dataType = 2
	case "converttrades":
		dataType = 3
	case "validatecandles":
		dataType = 4
	default:
		return errors.New("unrecognised command, cannot set data type")
	}

	var conversionInterval time.Duration
	var overwriteExistingData bool

	switch dataType {
	case 0, 1:
		if c.IsSet("overwrite_existing_data") {
			overwriteExistingData = c.Bool("overwrite_existing_data")
		}
	case 2, 3:
		var cInterval int64
		if c.IsSet("conversion_interval") {
			cInterval = c.Int64("conversion_interval")
		}
		conversionInterval = time.Duration(cInterval) * time.Second
		if c.IsSet("overwrite_existing_data") {
			overwriteExistingData = c.Bool("overwrite_existing_data")
		}
	case 4:
		if c.IsSet("comparison_decimal_places") {
			comparisonDecimalPlaces = c.Uint64("comparison_decimal_places")
		}
	}

	var prerequisiteJobNickname string
	if c.IsSet("prerequisite") {
		prerequisiteJobNickname = c.String("prerequisite")
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()
	client := gctrpc.NewGoCryptoTraderClient(conn)
	request := &gctrpc.UpsertDataHistoryJobRequest{
		Nickname: nickname,
		Exchange: exchange,
		Asset:    assetType,
		Pair: &gctrpc.CurrencyPair{
			Delimiter: p.Delimiter,
			Base:      p.Base.String(),
			Quote:     p.Quote.String(),
		},
		StartDate:               negateLocalOffset(s),
		EndDate:                 negateLocalOffset(e),
		Interval:                int64(candleInterval),
		RequestSizeLimit:        int64(requestSizeLimit),
		DataType:                dataType,
		MaxRetryAttempts:        int64(maxRetryAttempts),
		BatchSize:               int64(batchSize),
		ConversionInterval:      int64(conversionInterval),
		OverwriteExistingData:   overwriteExistingData,
		PrerequisiteJobNickname: prerequisiteJobNickname,
		InsertOnly:              !upsert,
		DecimalPlaceComparison:  int64(comparisonDecimalPlaces),
	}

	result, err := client.UpsertDataHistoryJob(context.Background(), request)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func getDataHistoryJobsBetween(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	if c.IsSet("start_date") {
		startTime = c.String("start_date")
	} else {
		startTime = c.Args().First()
	}
	if c.IsSet("end_date") {
		endTime = c.String("end_date")
	} else {
		endTime = c.Args().Get(1)
	}
	s, err := time.Parse(common.SimpleTimeFormat, startTime)
	if err != nil {
		return fmt.Errorf("invalid time format for start: %v", err)
	}
	e, err := time.Parse(common.SimpleTimeFormat, endTime)
	if err != nil {
		return fmt.Errorf("invalid time format for end: %v", err)
	}

	if e.Before(s) {
		return errors.New("start cannot be after end")
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.GetDataHistoryJobsBetween(context.Background(),
		&gctrpc.GetDataHistoryJobsBetweenRequest{
			StartDate: negateLocalOffset(s),
			EndDate:   negateLocalOffset(e),
		})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func setDataHistoryJobStatus(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	var id string
	if c.IsSet("id") {
		id = c.String("id")
	} else {
		id = c.Args().First()
	}

	var nickname string
	if c.IsSet("nickname") {
		nickname = c.String("nickname")
	}

	if nickname != "" && id != "" {
		return errors.New("can only set 'id' OR 'nickname'")
	}

	var status int64
	switch c.Command.Name {
	case "deletejob":
		status = 3
	case "pausejob":
		status = 4
	case "unpausejob":
		status = 0
	default:
		return fmt.Errorf("unrecognised data history job status type")
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()
	client := gctrpc.NewGoCryptoTraderClient(conn)
	request := &gctrpc.SetDataHistoryJobStatusRequest{
		Id:       id,
		Nickname: nickname,
		Status:   status,
	}

	result, err := client.SetDataHistoryJobStatus(context.Background(), request)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func getDataHistoryJobSummary(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	var nickname string
	if c.IsSet("nickname") {
		nickname = c.String("nickname")
	} else {
		nickname = c.Args().First()
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()
	client := gctrpc.NewGoCryptoTraderClient(conn)
	request := &gctrpc.GetDataHistoryJobDetailsRequest{
		Nickname: nickname,
	}

	result, err := client.GetDataHistoryJobSummary(context.Background(), request)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func setPrerequisiteJob(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}

	var nickname string
	if c.IsSet("nickname") {
		nickname = c.String("nickname")
	} else {
		nickname = c.Args().First()
	}

	var prerequisite string
	if c.IsSet("prerequisite") {
		prerequisite = c.String("prerequisite")
	} else {
		prerequisite = c.Args().Get(1)
	}

	if c.Command.Name == "updateprerequisite" && prerequisite == "" {
		return errors.New("prerequisite required")
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}()
	client := gctrpc.NewGoCryptoTraderClient(conn)
	request := &gctrpc.UpdateDataHistoryJobPrerequisiteRequest{
		PrerequisiteJobNickname: prerequisite,
		Nickname:                nickname,
	}

	result, err := client.UpdateDataHistoryJobPrerequisite(context.Background(), request)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}
