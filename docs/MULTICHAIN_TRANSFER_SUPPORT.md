# Multichain transfer support

Several exchanges support deposits and withdrawals by other blockchain networks. An example would be Tether (USDT) which supports the ERC20 (Ethereum), TRC20 (Tron) and Omni (BTC) networks.

GoCryptoTrader contains a `GetAvailableTransferChains` exchange method for supported exchanges which returns a list of the supported transfer chains specified by a cryptocurrency.

A simple demonstration using `gctcli` is as follows:

## Obtaining a list of supported transfer chains

```sh
$ ./gctcli getavailabletransferchains --exchange=ftx --cryptocurrency=usdt
{
 "chains": [
  "erc20",
  "trx",
  "sol",
  "omni"
 ]
}
```

## Obtaining a deposit address based on a specific cryptocurrency and chain

```sh
$ ./gctcli getcryptocurrencydepositaddress --exchange=ftx --cryptocurrency=usdt --chain=sol
{
 "address": "GW3oT9JpFyTkCWPnt6Yw9ugppSQwDv4ZMG1vabC8WmHS"
}
```

## Withdrawing

```sh
$ ./gctcli withdrawcryptofunds --exchange=ftx --currency=USDT --address=TJU9piX2WA8WTvxVKMqpvTzZGhvXQAZKSY --amount=10 --chain=trx
{
 "id": "01234567-0000-0000-0000-000000000000",
}
```

## Exchange multichain transfer support table

| Exchange | Deposits | Withdrawals | Notes|
|----------|----------|-------------|------|
| Alphapoint | No | No | |
| Binance | Yes | Yes | |
| Bitfinex | Yes | Yes | Only supports USDT |
| Bitflyer | No | No | |
| Bithumb | No | No | |
| BitMEX | No | No | Supports BTC only |
| Bitstamp | No | No | |
| Bittrex | No | No | NA |
| BTCMarkets | No | No| NA  |
| BTSE | No | No | Only through website |
| CoinbasePro | No | No | No|
| COINUT | No | No | NA |
| Exmo | Yes | Yes | Addresses must be created via their website first |
| FTX | Yes | Yes | |
| GateIO | Yes | Yes | |
| Gemini | No | No | |
| HitBTC | No | No | |
| Huobi.Pro | Yes | Yes | |
| ItBit | No | No | |
| Kraken | Yes | Yes | Front-end and API don't match total available transfer chains |
| Lbank | No | No | |
| LocalBitcoins | No | No | Supports BTC only |
| OKCoin International | No | No | Requires API update to version 5 |
| OKEX | No | No | Same as above |
| Poloniex | Yes | Yes | |
| Yobit | No | No | |
| ZB.COM | Yes | No | Addresses must be created via their website first |
