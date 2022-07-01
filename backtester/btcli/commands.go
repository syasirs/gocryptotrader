package main

import (
	"fmt"
	"path/filepath"

	"github.com/thrasher-corp/gocryptotrader/backtester/btrpc"
	"github.com/thrasher-corp/gocryptotrader/backtester/config"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var executeStrategyFromFileCommand = &cli.Command{
	Name:      "executestrategyfromfile",
	Usage:     "runs the strategy from a config file",
	ArgsUsage: "<path>",
	Action:    executeStrategyFromFile,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "the filepath to a strategy to execute",
		},
	},
}

func executeStrategyFromFile(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)

	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowCommandHelp(c, "executestrategyfromfile")
	}

	var path string
	if c.IsSet("path") {
		path = c.String("path")
	} else {
		path = c.Args().First()
	}

	client := btrpc.NewBacktesterClient(conn)
	result, err := client.ExecuteStrategyFromFile(
		c.Context,
		&btrpc.ExecuteStrategyFromFileRequest{
			StrategyFilePath: path,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

var executeStrategyFromConfigCommand = &cli.Command{
	Name:        "executestrategyfromconfig",
	Usage:       "runs the default strategy config but via passing in as a struct instead of a filepath - this is a proof-of-concept implementation",
	Description: "the cli is not a good place to manage this type of command with n variables to pass in from a command line",
	Action:      executeStrategyFromConfig,
}

// executeStrategyFromConfig this is a proof of concept command
// it demonstrates that a user can send complex strategies via GRPC
// and have them execute. The ultimate goal is to allow a user to continuously
// tweak values and send them via GRPC and determine the best returns then test them across
// large ranges of data to avoid over fitting
func executeStrategyFromConfig(c *cli.Context) error {
	conn, cancel, err := setupClient(c)
	if err != nil {
		return err
	}
	defer closeConn(conn, cancel)
	defaultPath := filepath.Join(
		"..",
		"config",
		"strategyexamples",
		"ftx-cash-carry.strat")
	defaultConfig, err := config.ReadStrategyConfigFromFile(defaultPath)
	if err != nil {
		return err
	}
	customSettings := make([]*btrpc.CustomSettings, len(defaultConfig.StrategySettings.CustomSettings))
	x := 0
	for k, v := range defaultConfig.StrategySettings.CustomSettings {
		customSettings[x] = &btrpc.CustomSettings{
			KeyField: k,
			KeyValue: fmt.Sprintf("%v", v),
		}
		x++
	}

	currencySettings := make([]*btrpc.CurrencySettings, len(defaultConfig.CurrencySettings))
	for i := range defaultConfig.CurrencySettings {
		var sd *btrpc.SpotDetails
		if defaultConfig.CurrencySettings[i].SpotDetails != nil {
			sd.InitialBaseFunds = defaultConfig.CurrencySettings[i].SpotDetails.InitialBaseFunds.String()
			sd.InitialQuoteFunds = defaultConfig.CurrencySettings[i].SpotDetails.InitialQuoteFunds.String()
		}
		var fd *btrpc.FuturesDetails
		if defaultConfig.CurrencySettings[i].FuturesDetails != nil {
			fd.Leverage = &btrpc.Leverage{
				CanUseLeverage:                 defaultConfig.CurrencySettings[i].FuturesDetails.Leverage.CanUseLeverage,
				MaximumOrdersWithLeverageRatio: defaultConfig.CurrencySettings[i].FuturesDetails.Leverage.MaximumOrdersWithLeverageRatio.String(),
				MaximumLeverageRate:            defaultConfig.CurrencySettings[i].FuturesDetails.Leverage.MaximumOrderLeverageRate.String(),
				MaximumCollateralLeverageRate:  defaultConfig.CurrencySettings[i].FuturesDetails.Leverage.MaximumCollateralLeverageRate.String(),
			}
		}
		currencySettings[i] = &btrpc.CurrencySettings{
			ExchangeName: defaultConfig.CurrencySettings[i].ExchangeName,
			Asset:        defaultConfig.CurrencySettings[i].Asset.String(),
			Base:         defaultConfig.CurrencySettings[i].Base.String(),
			Quote:        defaultConfig.CurrencySettings[i].Quote.String(),
			BuySide: &btrpc.PurchaseSide{
				MinimumSize:  defaultConfig.CurrencySettings[i].BuySide.MinimumSize.String(),
				MaximumSize:  defaultConfig.CurrencySettings[i].BuySide.MaximumSize.String(),
				MaximumTotal: defaultConfig.CurrencySettings[i].BuySide.MaximumTotal.String(),
			},
			SellSide: &btrpc.PurchaseSide{
				MinimumSize:  defaultConfig.CurrencySettings[i].SellSide.MinimumSize.String(),
				MaximumSize:  defaultConfig.CurrencySettings[i].SellSide.MaximumSize.String(),
				MaximumTotal: defaultConfig.CurrencySettings[i].SellSide.MaximumTotal.String(),
			},
			MinSlippagePercent:         defaultConfig.CurrencySettings[i].MinimumSlippagePercent.String(),
			MaxSlippagePercent:         defaultConfig.CurrencySettings[i].MaximumSlippagePercent.String(),
			MakerFeeOverride:           defaultConfig.CurrencySettings[i].MakerFee.String(),
			TakerFeeOverride:           defaultConfig.CurrencySettings[i].TakerFee.String(),
			MaximumHoldingsRatio:       defaultConfig.CurrencySettings[i].MaximumHoldingsRatio.String(),
			SkipCandleVolumeFitting:    defaultConfig.CurrencySettings[i].SkipCandleVolumeFitting,
			UseExchangeOrderLimits:     defaultConfig.CurrencySettings[i].CanUseExchangeLimits,
			UseExchange_PNLCalculation: defaultConfig.CurrencySettings[i].UseExchangePNLCalculation,
			SpotDetails:                sd,
			FuturesDetails:             fd,
		}
	}

	exchangeLevelFunding := make([]*btrpc.ExchangeLevelFunding, len(defaultConfig.FundingSettings.ExchangeLevelFunding))
	for i := range defaultConfig.FundingSettings.ExchangeLevelFunding {
		exchangeLevelFunding[i] = &btrpc.ExchangeLevelFunding{
			ExchangeName: defaultConfig.FundingSettings.ExchangeLevelFunding[i].ExchangeName,
			Asset:        defaultConfig.FundingSettings.ExchangeLevelFunding[i].Asset.String(),
			Currency:     defaultConfig.FundingSettings.ExchangeLevelFunding[i].Currency.String(),
			InitialFunds: defaultConfig.FundingSettings.ExchangeLevelFunding[i].InitialFunds.String(),
			TransferFee:  defaultConfig.FundingSettings.ExchangeLevelFunding[i].TransferFee.String(),
		}
	}

	dataSettings := &btrpc.DataSettings{
		Interval: uint64(defaultConfig.DataSettings.Interval.Duration().Nanoseconds()),
		Datatype: defaultConfig.DataSettings.DataType,
	}
	if defaultConfig.DataSettings.APIData != nil {
		dataSettings.ApiData = &btrpc.ApiData{
			StartDate:        timestamppb.New(defaultConfig.DataSettings.APIData.StartDate),
			EndDate:          timestamppb.New(defaultConfig.DataSettings.APIData.EndDate),
			InclusiveEndDate: defaultConfig.DataSettings.APIData.InclusiveEndDate,
		}
	}
	if defaultConfig.DataSettings.LiveData != nil {
		dataSettings.LiveData = &btrpc.LiveData{
			ApiKeyOverride:        defaultConfig.DataSettings.LiveData.APIKeyOverride,
			ApiSecretOverride:     defaultConfig.DataSettings.LiveData.APISecretOverride,
			ApiClientIdOverride:   defaultConfig.DataSettings.LiveData.APIClientIDOverride,
			Api_2FaOverride:       defaultConfig.DataSettings.LiveData.API2FAOverride,
			ApiSubAccountOverride: defaultConfig.DataSettings.LiveData.APISubAccountOverride,
			UseRealOrders:         defaultConfig.DataSettings.LiveData.RealOrders,
		}
	}
	if defaultConfig.DataSettings.CSVData != nil {
		dataSettings.CsvData = &btrpc.CSVData{
			Path: defaultConfig.DataSettings.CSVData.FullPath,
		}
	}
	if defaultConfig.DataSettings.DatabaseData != nil {
		dbConnectionDetails := &btrpc.DatabaseConnectionDetails{
			Host:     defaultConfig.DataSettings.DatabaseData.Config.Host,
			Port:     int64(defaultConfig.DataSettings.DatabaseData.Config.Port),
			Password: defaultConfig.DataSettings.DatabaseData.Config.Password,
			Database: defaultConfig.DataSettings.DatabaseData.Config.Database,
			SslMode:  defaultConfig.DataSettings.DatabaseData.Config.SSLMode,
			UserName: defaultConfig.DataSettings.DatabaseData.Config.Username,
		}
		dbConfig := &btrpc.DatabaseConfig{
			Enabled: false,
			Verbose: false,
			Driver:  "",
			Config:  dbConnectionDetails,
		}
		dataSettings.DatabaseData = &btrpc.DatabaseData{
			StartDate:        timestamppb.New(defaultConfig.DataSettings.DatabaseData.StartDate),
			EndDate:          timestamppb.New(defaultConfig.DataSettings.DatabaseData.EndDate),
			Config:           dbConfig,
			Path:             defaultConfig.DataSettings.DatabaseData.Path,
			InclusiveEndDate: defaultConfig.DataSettings.DatabaseData.InclusiveEndDate,
		}
	}

	cfg := &btrpc.Config{
		Nickname: defaultConfig.Nickname,
		Goal:     defaultConfig.Goal,
		StrategySettings: &btrpc.StrategySettings{
			Name:                            defaultConfig.StrategySettings.Name,
			UseSimultaneousSignalProcessing: defaultConfig.StrategySettings.SimultaneousSignalProcessing,
			Disable_USDTracking:             defaultConfig.StrategySettings.DisableUSDTracking,
			CustomSettings:                  customSettings,
		},
		FundingSettings: &btrpc.FundingSettings{
			UseExchangeLevelFunding: defaultConfig.FundingSettings.UseExchangeLevelFunding,
			ExchangeLevelFunding:    exchangeLevelFunding,
		},
		CurrencySettings: currencySettings,
		DataSettings:     dataSettings,
		PortfolioSettings: &btrpc.PortfolioSettings{
			Leverage: &btrpc.Leverage{
				CanUseLeverage:                 defaultConfig.PortfolioSettings.Leverage.CanUseLeverage,
				MaximumOrdersWithLeverageRatio: defaultConfig.PortfolioSettings.Leverage.MaximumOrdersWithLeverageRatio.String(),
				MaximumLeverageRate:            defaultConfig.PortfolioSettings.Leverage.MaximumOrderLeverageRate.String(),
				MaximumCollateralLeverageRate:  defaultConfig.PortfolioSettings.Leverage.MaximumCollateralLeverageRate.String(),
			},
			BuySide: &btrpc.PurchaseSide{
				MinimumSize:  defaultConfig.PortfolioSettings.BuySide.MinimumSize.String(),
				MaximumSize:  defaultConfig.PortfolioSettings.BuySide.MaximumSize.String(),
				MaximumTotal: defaultConfig.PortfolioSettings.BuySide.MaximumTotal.String(),
			},
			SellSide: &btrpc.PurchaseSide{
				MinimumSize:  defaultConfig.PortfolioSettings.SellSide.MinimumSize.String(),
				MaximumSize:  defaultConfig.PortfolioSettings.SellSide.MaximumSize.String(),
				MaximumTotal: defaultConfig.PortfolioSettings.SellSide.MaximumTotal.String(),
			},
		},
		StatisticSettings: &btrpc.StatisticSettings{
			RiskFreeRate: defaultConfig.StatisticSettings.RiskFreeRate.String(),
		},
	}

	client := btrpc.NewBacktesterClient(conn)
	result, err := client.ExecuteStrategyFromConfig(
		c.Context,
		&btrpc.ExecuteStrategyFromConfigRequest{
			Config: cfg,
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}
