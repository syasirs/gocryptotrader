package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/gctrpc"
	"github.com/urfave/cli"
)

var exchangePairManagerCommand = cli.Command{
	Name:      "pair",
	Usage:     "execute exchange pair management command",
	ArgsUsage: "<command> <args>",
	Subcommands: []cli.Command{
		{
			Name:      "get",
			Usage:     "returns all enabled and available pairs by asset type",
			ArgsUsage: "<asset>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.StringFlag{
					Name:  "asset",
					Usage: "asset",
				},
			},
			Action: getExchangePairs,
		},
		{
			Name:  "disableasset",
			Usage: "disables asset type",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.StringFlag{
					Name:  "asset",
					Usage: "asset",
				},
			},
			Action: enableDisableExchangeAsset,
		},
		{
			Name:  "enableasset",
			Usage: "enables asset type",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.StringFlag{
					Name:  "asset",
					Usage: "asset",
				},
				cli.BoolTFlag{
					Name:   "enable",
					Hidden: true,
				},
			},
			Action: enableDisableExchangeAsset,
		},
		{
			Name:  "disable",
			Usage: "disable pairs by asset type",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.StringFlag{
					Name:  "pairs",
					Usage: "either a single currency pair string or comma delimiter string of pairs e.g. \"BTC-USD,XRP-USD\"",
				},
				cli.StringFlag{
					Name:  "asset",
					Usage: "asset",
				},
			},
			Action: enableDisableExchangePair,
		},
		{
			Name:  "enable",
			Usage: "enable pairs by asset type",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.StringFlag{
					Name:  "pairs",
					Usage: "either a single currency pair string or comma delimiter string of pairs e.g. \"BTC-USD,XRP-USD\"",
				},
				cli.StringFlag{
					Name:  "asset",
					Usage: "asset",
				},
				cli.BoolTFlag{
					Name:   "enable",
					Hidden: true,
				},
			},
			Action: enableDisableExchangePair,
		},
		{
			Name:  "enableall",
			Usage: "enable all pairs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				cli.BoolTFlag{
					Name:   "enable",
					Hidden: true,
				},
			},
			Action: enableDisableAllExchangePairs,
		},
		{
			Name:  "disableall",
			Usage: "dissable all pairs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: enableDisableAllExchangePairs,
		},
		{
			Name:  "update",
			Usage: "fetches supported pairs from the exchange and updates available pairs and removes unsupported enable pairs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: updateExchangeSupportedPairs,
		},
		{
			Name:  "getassets",
			Usage: "fetches supported assets",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: getExchangeAssets,
		},
	},
}

func enableDisableExchangePair(c *cli.Context) error {
	enable := c.BoolT("enable")
	if c.NArg() == 0 && c.NumFlags() == 0 {
		if enable {
			return cli.ShowCommandHelp(c, "enable")
		}

		return cli.ShowCommandHelp(c, "disable")
	}

	var exchange string
	var pairs string
	var asset string

	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	if c.IsSet("pair") {
		pairs = c.String("pairs")
	} else {
		pairs = c.Args().Get(1)
	}

	if c.IsSet("asset") {
		asset = c.String("asset")
	} else {
		asset = c.Args().Get(2)
	}

	asset = strings.ToLower(asset)
	if !validAsset(asset) {
		return errInvalidAsset
	}

	pairList := strings.Split(pairs, ",")

	var validPairs []*gctrpc.CurrencyPair
	for i := range pairList {
		if !validPair(pairList[i]) {
			return errInvalidPair
		}

		var p currency.Pair
		if strings.EqualFold(exchange, "okex") && strings.Contains(pairList[i], "SWAP") {
			p = currency.NewPairDelimiter(pairList[i], "_")
		} else {
			p = currency.NewPairDelimiter(pairList[i], pairDelimiter)
		}
		fmt.Println(p)

		fmt.Println(p.Base.String())
		fmt.Println(p.Quote.String())
		fmt.Println(p.Delimiter)

		validPairs = append(validPairs, &gctrpc.CurrencyPair{
			Delimiter: p.Delimiter,
			Base:      p.Base.String(),
			Quote:     p.Quote.String(),
		})
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)

	var result *gctrpc.GenericExchangeNameResponse
	if enable {
		result, err = client.EnableExchangePair(context.Background(),
			&gctrpc.ExchangePairRequest{
				Exchange:  exchange,
				Pairs:     validPairs,
				AssetType: asset,
			},
		)
	} else {
		result, err = client.DisableExchangePair(context.Background(),
			&gctrpc.ExchangePairRequest{
				Exchange:  exchange,
				Pairs:     validPairs,
				AssetType: asset,
			},
		)
	}
	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func getExchangePairs(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	var asset string

	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	if c.IsSet("asset") {
		asset = c.String("asset")
	} else {
		asset = c.Args().Get(1)
	}

	asset = strings.ToLower(asset)
	if !validAsset(asset) {
		return errInvalidAsset
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.GetExchangePairs(context.Background(),
		&gctrpc.GetExchangePairsRequest{
			Exchange: exchange,
			Asset:    asset,
		},
	)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func enableDisableExchangeAsset(c *cli.Context) error {
	enable := c.BoolT("enable")
	if c.NArg() == 0 && c.NumFlags() == 0 {
		if enable {
			return cli.ShowCommandHelp(c, "enableasset")
		}
		return cli.ShowCommandHelp(c, "disableasset")
	}

	var exchange string
	var asset string

	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	if c.IsSet("asset") {
		asset = c.String("asset")
	} else {
		asset = c.Args().Get(1)
	}

	asset = strings.ToLower(asset)
	if !validAsset(asset) {
		return errInvalidAsset
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.EnableDisableExchangeAsset(context.Background(),
		&gctrpc.ExchangeDisableEnableAssetRequest{
			Exchange: exchange,
			Asset:    asset,
			Enable:   enable,
		},
	)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func enableDisableAllExchangePairs(c *cli.Context) error {
	enable := c.BoolT("enable")
	if c.NArg() == 0 && c.NumFlags() == 0 {
		if enable {
			return cli.ShowCommandHelp(c, "enableall")
		}
		return cli.ShowCommandHelp(c, "disableall")
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.EnableDisableAllExchangePairs(context.Background(),
		&gctrpc.ExchangeDisableEnableAllPairsRequest{
			Exchange: exchange,
			Enable:   enable,
		},
	)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func updateExchangeSupportedPairs(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.UpdateExchangeSupportedPairs(context.Background(),
		&gctrpc.UpdateExchangeSupportedPairsRequest{
			Exchange: exchange,
		},
	)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func getExchangeAssets(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return errInvalidExchange
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.GetExchangeAssets(context.Background(),
		&gctrpc.GetExchangeAssetsRequest{
			Exchange: exchange,
		},
	)
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}
