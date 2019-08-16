# GoCryptoTrader package Database

<img src="https://github.com/thrasher-corp/gocryptotrader/blob/master/web/src/assets/page-logo.png?raw=true" width="350px" height="350px" hspace="70">


[![Build Status](https://travis-ci.org/thrasher-corp/gocryptotrader.svg?branch=master)](https://travis-ci.org/thrasher-corp/gocryptotrader)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/thrasher-corp/gocryptotrader/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/thrasher-corp/gocryptotrader?status.svg)](https://godoc.org/github.com/thrasher-corp/gocryptotrader/portfolio)
[![Coverage Status](http://codecov.io/github/thrasher-corp/gocryptotrader/coverage.svg?branch=master)](http://codecov.io/github/thrasher-corp/gocryptotrader?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thrasher-corp/gocryptotrader)](https://goreportcard.com/report/github.com/thrasher-corp/gocryptotrader)


This database package is part of the GoCryptoTrader codebase.

## This is still in active development

You can track ideas, planned features and what's in progresss on this Trello board: [https://trello.com/b/ZAhMhpOy/gocryptotrader](https://trello.com/b/ZAhMhpOy/gocryptotrader).

Join our slack to discuss all things related to GoCryptoTrader! [GoCryptoTrader Slack](https://join.slack.com/t/gocryptotrader/shared_invite/enQtNTQ5NDAxMjA2Mjc5LTQyYjIxNGVhMWU5MDZlOGYzMmE0NTJmM2MzYWY5NGMzMmM4MzUwNTBjZTEzNjIwODM5NDcxODQwZDljMGQyNGY)

## Current Features for database package

+ Establishes & Maintains database connection across program life cycle
+ Multiple database support via simple repository model
+ Run migration on connection to assure database is at correct version

## How to use

#####To Manually migrate to the latest database you can run the "dbmigrate" helper in the cmd folder

This will parse and run all migration files in your $GoCryptoTrader/database/migrations

_This is also run from the bot when a connection is established to the database_

```sh
go run ./cmd/dbmigrate
```
A Makefile command has also been added for this
```sh
make db_migrate
```

#####To create a new migrate file you can also run the same command with the -create "migration name" flag

```sh
go run ./cmd/dbmigrate -create "alter some table"
```

##### Adding a new model

+ Create Model in github.com/thrasher-corp/gocryptotrader/database/models directory

##### Adding a Repository
+ Create Repository directory in github.com/thrasher-corp/gocryptotrader/database/repository/
+ Create a base Repository interface with any required Methods
+ Create a per driver implementation of the Repository that implement all required methods to match the interface

## Contribution

Please feel free to submit any pull requests or suggest any desired features to be added.

When submitting a PR, please abide by our coding guidelines:

+ Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
+ Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
+ Code must adhere to our [coding style](https://github.com/thrasher-corp/gocryptotrader/blob/master/doc/coding_style.md).
+ Pull requests need to be based on and opened against the `master` branch.

## Donations

<img src="https://github.com/thrasher-corp/gocryptotrader/blob/master/web/src/assets/donate.png?raw=true" hspace="70">

If this framework helped you in any way, or you would like to support the developers working on it, please donate Bitcoin to:

***1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB***

