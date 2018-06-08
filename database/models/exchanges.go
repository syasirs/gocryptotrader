// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/strmangle"
	"gopkg.in/volatiletech/null.v6"
)

// Exchange is an object representing the database table.
type Exchange struct {
	ExchangeID               int64       `boil:"exchange_id" json:"exchange_id" toml:"exchange_id" yaml:"exchange_id"`
	ConfigID                 int64       `boil:"config_id" json:"config_id" toml:"config_id" yaml:"config_id"`
	ExchangeName             string      `boil:"exchange_name" json:"exchange_name" toml:"exchange_name" yaml:"exchange_name"`
	Enabled                  bool        `boil:"enabled" json:"enabled" toml:"enabled" yaml:"enabled"`
	IsVerbose                bool        `boil:"is_verbose" json:"is_verbose" toml:"is_verbose" yaml:"is_verbose"`
	Websocket                bool        `boil:"websocket" json:"websocket" toml:"websocket" yaml:"websocket"`
	UseSandbox               bool        `boil:"use_sandbox" json:"use_sandbox" toml:"use_sandbox" yaml:"use_sandbox"`
	RestPollingDelay         int64       `boil:"rest_polling_delay" json:"rest_polling_delay" toml:"rest_polling_delay" yaml:"rest_polling_delay"`
	HTTPTimeout              int64       `boil:"http_timeout" json:"http_timeout" toml:"http_timeout" yaml:"http_timeout"`
	AuthenticatedAPISupport  bool        `boil:"authenticated_api_support" json:"authenticated_api_support" toml:"authenticated_api_support" yaml:"authenticated_api_support"`
	APIKey                   null.String `boil:"api_key" json:"api_key,omitempty" toml:"api_key" yaml:"api_key,omitempty"`
	APISecret                null.String `boil:"api_secret" json:"api_secret,omitempty" toml:"api_secret" yaml:"api_secret,omitempty"`
	ClientID                 null.String `boil:"client_id" json:"client_id,omitempty" toml:"client_id" yaml:"client_id,omitempty"`
	AvailablePairs           string      `boil:"available_pairs" json:"available_pairs" toml:"available_pairs" yaml:"available_pairs"`
	EnabledPairs             string      `boil:"enabled_pairs" json:"enabled_pairs" toml:"enabled_pairs" yaml:"enabled_pairs"`
	BaseCurrencies           string      `boil:"base_currencies" json:"base_currencies" toml:"base_currencies" yaml:"base_currencies"`
	AssetTypes               null.String `boil:"asset_types" json:"asset_types,omitempty" toml:"asset_types" yaml:"asset_types,omitempty"`
	SupportedAutoPairUpdates bool        `boil:"supported_auto_pair_updates" json:"supported_auto_pair_updates" toml:"supported_auto_pair_updates" yaml:"supported_auto_pair_updates"`
	PairsLastUpdated         time.Time   `boil:"pairs_last_updated" json:"pairs_last_updated" toml:"pairs_last_updated" yaml:"pairs_last_updated"`

	R *exchangeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L exchangeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ExchangeColumns = struct {
	ExchangeID               string
	ConfigID                 string
	ExchangeName             string
	Enabled                  string
	IsVerbose                string
	Websocket                string
	UseSandbox               string
	RestPollingDelay         string
	HTTPTimeout              string
	AuthenticatedAPISupport  string
	APIKey                   string
	APISecret                string
	ClientID                 string
	AvailablePairs           string
	EnabledPairs             string
	BaseCurrencies           string
	AssetTypes               string
	SupportedAutoPairUpdates string
	PairsLastUpdated         string
}{
	ExchangeID:               "exchange_id",
	ConfigID:                 "config_id",
	ExchangeName:             "exchange_name",
	Enabled:                  "enabled",
	IsVerbose:                "is_verbose",
	Websocket:                "websocket",
	UseSandbox:               "use_sandbox",
	RestPollingDelay:         "rest_polling_delay",
	HTTPTimeout:              "http_timeout",
	AuthenticatedAPISupport:  "authenticated_api_support",
	APIKey:                   "api_key",
	APISecret:                "api_secret",
	ClientID:                 "client_id",
	AvailablePairs:           "available_pairs",
	EnabledPairs:             "enabled_pairs",
	BaseCurrencies:           "base_currencies",
	AssetTypes:               "asset_types",
	SupportedAutoPairUpdates: "supported_auto_pair_updates",
	PairsLastUpdated:         "pairs_last_updated",
}

// exchangeR is where relationships are stored.
type exchangeR struct {
	CurrencyPairFormats CurrencyPairFormatSlice
}

// exchangeL is where Load methods for each relationship are stored.
type exchangeL struct{}

var (
	exchangeColumns               = []string{"exchange_id", "config_id", "exchange_name", "enabled", "is_verbose", "websocket", "use_sandbox", "rest_polling_delay", "http_timeout", "authenticated_api_support", "api_key", "api_secret", "client_id", "available_pairs", "enabled_pairs", "base_currencies", "asset_types", "supported_auto_pair_updates", "pairs_last_updated"}
	exchangeColumnsWithoutDefault = []string{"exchange_id", "config_id", "exchange_name", "enabled", "is_verbose", "websocket", "use_sandbox", "rest_polling_delay", "http_timeout", "authenticated_api_support", "api_key", "api_secret", "client_id", "available_pairs", "enabled_pairs", "base_currencies", "asset_types", "supported_auto_pair_updates", "pairs_last_updated"}
	exchangeColumnsWithDefault    = []string{}
	exchangePrimaryKeyColumns     = []string{"exchange_id"}
)

type (
	// ExchangeSlice is an alias for a slice of pointers to Exchange.
	// This should generally be used opposed to []Exchange.
	ExchangeSlice []*Exchange

	exchangeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	exchangeType                 = reflect.TypeOf(&Exchange{})
	exchangeMapping              = queries.MakeStructMapping(exchangeType)
	exchangePrimaryKeyMapping, _ = queries.BindMapping(exchangeType, exchangeMapping, exchangePrimaryKeyColumns)
	exchangeInsertCacheMut       sync.RWMutex
	exchangeInsertCache          = make(map[string]insertCache)
	exchangeUpdateCacheMut       sync.RWMutex
	exchangeUpdateCache          = make(map[string]updateCache)
	exchangeUpsertCacheMut       sync.RWMutex
	exchangeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force bytes in case of primary key column that uses []byte (for relationship compares)
	_ = bytes.MinRead
)

// OneP returns a single exchange record from the query, and panics on error.
func (q exchangeQuery) OneP() *Exchange {
	o, err := q.One()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return o
}

// One returns a single exchange record from the query.
func (q exchangeQuery) One() (*Exchange, error) {
	o := &Exchange{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for exchanges")
	}

	return o, nil
}

// AllP returns all Exchange records from the query, and panics on error.
func (q exchangeQuery) AllP() ExchangeSlice {
	o, err := q.All()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return o
}

// All returns all Exchange records from the query.
func (q exchangeQuery) All() (ExchangeSlice, error) {
	var o []*Exchange

	err := q.Bind(&o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Exchange slice")
	}

	return o, nil
}

// CountP returns the count of all Exchange records in the query, and panics on error.
func (q exchangeQuery) CountP() int64 {
	c, err := q.Count()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return c
}

// Count returns the count of all Exchange records in the query.
func (q exchangeQuery) Count() (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow().Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count exchanges rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table, and panics on error.
func (q exchangeQuery) ExistsP() bool {
	e, err := q.Exists()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}

// Exists checks if the row exists in the table.
func (q exchangeQuery) Exists() (bool, error) {
	var count int64

	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow().Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if exchanges exists")
	}

	return count > 0, nil
}

// CurrencyPairFormatsG retrieves all the currency_pair_format's currency pair format.
func (o *Exchange) CurrencyPairFormatsG(mods ...qm.QueryMod) currencyPairFormatQuery {
	return o.CurrencyPairFormats(boil.GetDB(), mods...)
}

// CurrencyPairFormats retrieves all the currency_pair_format's currency pair format with an executor.
func (o *Exchange) CurrencyPairFormats(exec boil.Executor, mods ...qm.QueryMod) currencyPairFormatQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"currency_pair_format\".\"exchange_id\"=?", o.ExchangeID),
	)

	query := CurrencyPairFormats(exec, queryMods...)
	queries.SetFrom(query.Query, "\"currency_pair_format\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"currency_pair_format\".*"})
	}

	return query
}

// LoadCurrencyPairFormats allows an eager lookup of values, cached into the
// loaded structs of the objects.
func (exchangeL) LoadCurrencyPairFormats(e boil.Executor, singular bool, maybeExchange interface{}) error {
	var slice []*Exchange
	var object *Exchange

	count := 1
	if singular {
		object = maybeExchange.(*Exchange)
	} else {
		slice = *maybeExchange.(*[]*Exchange)
		count = len(slice)
	}

	args := make([]interface{}, count)
	if singular {
		if object.R == nil {
			object.R = &exchangeR{}
		}
		args[0] = object.ExchangeID
	} else {
		for i, obj := range slice {
			if obj.R == nil {
				obj.R = &exchangeR{}
			}
			args[i] = obj.ExchangeID
		}
	}

	query := fmt.Sprintf(
		"select * from \"currency_pair_format\" where \"exchange_id\" in (%s)",
		strmangle.Placeholders(dialect.IndexPlaceholders, count, 1, 1),
	)
	if boil.DebugMode {
		fmt.Fprintf(boil.DebugWriter, "%s\n%v\n", query, args)
	}

	results, err := e.Query(query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to eager load currency_pair_format")
	}
	defer results.Close()

	var resultSlice []*CurrencyPairFormat
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice currency_pair_format")
	}

	if singular {
		object.R.CurrencyPairFormats = resultSlice
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ExchangeID == foreign.ExchangeID {
				local.R.CurrencyPairFormats = append(local.R.CurrencyPairFormats, foreign)
				break
			}
		}
	}

	return nil
}

// AddCurrencyPairFormatsG adds the given related objects to the existing relationships
// of the exchange, optionally inserting them as new records.
// Appends related to o.R.CurrencyPairFormats.
// Sets related.R.Exchange appropriately.
// Uses the global database handle.
func (o *Exchange) AddCurrencyPairFormatsG(insert bool, related ...*CurrencyPairFormat) error {
	return o.AddCurrencyPairFormats(boil.GetDB(), insert, related...)
}

// AddCurrencyPairFormatsP adds the given related objects to the existing relationships
// of the exchange, optionally inserting them as new records.
// Appends related to o.R.CurrencyPairFormats.
// Sets related.R.Exchange appropriately.
// Panics on error.
func (o *Exchange) AddCurrencyPairFormatsP(exec boil.Executor, insert bool, related ...*CurrencyPairFormat) {
	if err := o.AddCurrencyPairFormats(exec, insert, related...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// AddCurrencyPairFormatsGP adds the given related objects to the existing relationships
// of the exchange, optionally inserting them as new records.
// Appends related to o.R.CurrencyPairFormats.
// Sets related.R.Exchange appropriately.
// Uses the global database handle and panics on error.
func (o *Exchange) AddCurrencyPairFormatsGP(insert bool, related ...*CurrencyPairFormat) {
	if err := o.AddCurrencyPairFormats(boil.GetDB(), insert, related...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// AddCurrencyPairFormats adds the given related objects to the existing relationships
// of the exchange, optionally inserting them as new records.
// Appends related to o.R.CurrencyPairFormats.
// Sets related.R.Exchange appropriately.
func (o *Exchange) AddCurrencyPairFormats(exec boil.Executor, insert bool, related ...*CurrencyPairFormat) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.ExchangeID = o.ExchangeID
			if err = rel.Insert(exec); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"currency_pair_format\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"exchange_id"}),
				strmangle.WhereClause("\"", "\"", 2, currencyPairFormatPrimaryKeyColumns),
			)
			values := []interface{}{o.ExchangeID, rel.CurrencyPairFormatID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}

			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.ExchangeID = o.ExchangeID
		}
	}

	if o.R == nil {
		o.R = &exchangeR{
			CurrencyPairFormats: related,
		}
	} else {
		o.R.CurrencyPairFormats = append(o.R.CurrencyPairFormats, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &currencyPairFormatR{
				Exchange: o,
			}
		} else {
			rel.R.Exchange = o
		}
	}
	return nil
}

// ExchangesG retrieves all records.
func ExchangesG(mods ...qm.QueryMod) exchangeQuery {
	return Exchanges(boil.GetDB(), mods...)
}

// Exchanges retrieves all the records using an executor.
func Exchanges(exec boil.Executor, mods ...qm.QueryMod) exchangeQuery {
	mods = append(mods, qm.From("\"exchanges\""))
	return exchangeQuery{NewQuery(exec, mods...)}
}

// FindExchangeG retrieves a single record by ID.
func FindExchangeG(exchangeID int64, selectCols ...string) (*Exchange, error) {
	return FindExchange(boil.GetDB(), exchangeID, selectCols...)
}

// FindExchangeGP retrieves a single record by ID, and panics on error.
func FindExchangeGP(exchangeID int64, selectCols ...string) *Exchange {
	retobj, err := FindExchange(boil.GetDB(), exchangeID, selectCols...)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return retobj
}

// FindExchange retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindExchange(exec boil.Executor, exchangeID int64, selectCols ...string) (*Exchange, error) {
	exchangeObj := &Exchange{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"exchanges\" where \"exchange_id\"=$1", sel,
	)

	q := queries.Raw(exec, query, exchangeID)

	err := q.Bind(exchangeObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from exchanges")
	}

	return exchangeObj, nil
}

// FindExchangeP retrieves a single record by ID with an executor, and panics on error.
func FindExchangeP(exec boil.Executor, exchangeID int64, selectCols ...string) *Exchange {
	retobj, err := FindExchange(exec, exchangeID, selectCols...)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return retobj
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *Exchange) InsertG(whitelist ...string) error {
	return o.Insert(boil.GetDB(), whitelist...)
}

// InsertGP a single record, and panics on error. See Insert for whitelist
// behavior description.
func (o *Exchange) InsertGP(whitelist ...string) {
	if err := o.Insert(boil.GetDB(), whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// InsertP a single record using an executor, and panics on error. See Insert
// for whitelist behavior description.
func (o *Exchange) InsertP(exec boil.Executor, whitelist ...string) {
	if err := o.Insert(exec, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Insert a single record using an executor.
// Whitelist behavior: If a whitelist is provided, only those columns supplied are inserted
// No whitelist behavior: Without a whitelist, columns are inferred by the following rules:
// - All columns without a default value are included (i.e. name, age)
// - All columns with a default, but non-zero are included (i.e. health = 75)
func (o *Exchange) Insert(exec boil.Executor, whitelist ...string) error {
	if o == nil {
		return errors.New("models: no exchanges provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(exchangeColumnsWithDefault, o)

	key := makeCacheKey(whitelist, nzDefaults)
	exchangeInsertCacheMut.RLock()
	cache, cached := exchangeInsertCache[key]
	exchangeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := strmangle.InsertColumnSet(
			exchangeColumns,
			exchangeColumnsWithDefault,
			exchangeColumnsWithoutDefault,
			nzDefaults,
			whitelist,
		)

		cache.valueMapping, err = queries.BindMapping(exchangeType, exchangeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(exchangeType, exchangeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"exchanges\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.IndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"exchanges\" DEFAULT VALUES"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		if len(wl) != 0 {
			cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into exchanges")
	}

	if !cached {
		exchangeInsertCacheMut.Lock()
		exchangeInsertCache[key] = cache
		exchangeInsertCacheMut.Unlock()
	}

	return nil
}

// UpdateG a single Exchange record. See Update for
// whitelist behavior description.
func (o *Exchange) UpdateG(whitelist ...string) error {
	return o.Update(boil.GetDB(), whitelist...)
}

// UpdateGP a single Exchange record.
// UpdateGP takes a whitelist of column names that should be updated.
// Panics on error. See Update for whitelist behavior description.
func (o *Exchange) UpdateGP(whitelist ...string) {
	if err := o.Update(boil.GetDB(), whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateP uses an executor to update the Exchange, and panics on error.
// See Update for whitelist behavior description.
func (o *Exchange) UpdateP(exec boil.Executor, whitelist ...string) {
	err := o.Update(exec, whitelist...)
	if err != nil {
		panic(boil.WrapErr(err))
	}
}

// Update uses an executor to update the Exchange.
// Whitelist behavior: If a whitelist is provided, only the columns given are updated.
// No whitelist behavior: Without a whitelist, columns are inferred by the following rules:
// - All columns are inferred to start with
// - All primary keys are subtracted from this set
// Update does not automatically update the record in case of default values. Use .Reload()
// to refresh the records.
func (o *Exchange) Update(exec boil.Executor, whitelist ...string) error {
	var err error
	key := makeCacheKey(whitelist, nil)
	exchangeUpdateCacheMut.RLock()
	cache, cached := exchangeUpdateCache[key]
	exchangeUpdateCacheMut.RUnlock()

	if !cached {
		wl := strmangle.UpdateColumnSet(
			exchangeColumns,
			exchangePrimaryKeyColumns,
			whitelist,
		)

		if len(whitelist) == 0 {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return errors.New("models: unable to update exchanges, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"exchanges\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, exchangePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(exchangeType, exchangeMapping, append(wl, exchangePrimaryKeyColumns...))
		if err != nil {
			return err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	_, err = exec.Exec(cache.query, values...)
	if err != nil {
		return errors.Wrap(err, "models: unable to update exchanges row")
	}

	if !cached {
		exchangeUpdateCacheMut.Lock()
		exchangeUpdateCache[key] = cache
		exchangeUpdateCacheMut.Unlock()
	}

	return nil
}

// UpdateAllP updates all rows with matching column names, and panics on error.
func (q exchangeQuery) UpdateAllP(cols M) {
	if err := q.UpdateAll(cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAll updates all rows with the specified column values.
func (q exchangeQuery) UpdateAll(cols M) error {
	queries.SetUpdate(q.Query, cols)

	_, err := q.Query.Exec()
	if err != nil {
		return errors.Wrap(err, "models: unable to update all for exchanges")
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (o ExchangeSlice) UpdateAllG(cols M) error {
	return o.UpdateAll(boil.GetDB(), cols)
}

// UpdateAllGP updates all rows with the specified column values, and panics on error.
func (o ExchangeSlice) UpdateAllGP(cols M) {
	if err := o.UpdateAll(boil.GetDB(), cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAllP updates all rows with the specified column values, and panics on error.
func (o ExchangeSlice) UpdateAllP(exec boil.Executor, cols M) {
	if err := o.UpdateAll(exec, cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ExchangeSlice) UpdateAll(exec boil.Executor, cols M) error {
	ln := int64(len(o))
	if ln == 0 {
		return nil
	}

	if len(cols) == 0 {
		return errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), exchangePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"exchanges\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, exchangePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to update all in exchange slice")
	}

	return nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *Exchange) UpsertG(updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) error {
	return o.Upsert(boil.GetDB(), updateOnConflict, conflictColumns, updateColumns, whitelist...)
}

// UpsertGP attempts an insert, and does an update or ignore on conflict. Panics on error.
func (o *Exchange) UpsertGP(updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) {
	if err := o.Upsert(boil.GetDB(), updateOnConflict, conflictColumns, updateColumns, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpsertP attempts an insert using an executor, and does an update or ignore on conflict.
// UpsertP panics on error.
func (o *Exchange) UpsertP(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) {
	if err := o.Upsert(exec, updateOnConflict, conflictColumns, updateColumns, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
func (o *Exchange) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) error {
	if o == nil {
		return errors.New("models: no exchanges provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(exchangeColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs postgres problems
	buf := strmangle.GetBuffer()

	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range updateColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range whitelist {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	exchangeUpsertCacheMut.RLock()
	cache, cached := exchangeUpsertCache[key]
	exchangeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := strmangle.InsertColumnSet(
			exchangeColumns,
			exchangeColumnsWithDefault,
			exchangeColumnsWithoutDefault,
			nzDefaults,
			whitelist,
		)

		update := strmangle.UpdateColumnSet(
			exchangeColumns,
			exchangePrimaryKeyColumns,
			updateColumns,
		)
		if len(update) == 0 {
			return errors.New("models: unable to upsert exchanges, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(exchangePrimaryKeyColumns))
			copy(conflict, exchangePrimaryKeyColumns)
		}
		cache.query = queries.BuildUpsertQueryPostgres(dialect, "\"exchanges\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(exchangeType, exchangeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(exchangeType, exchangeMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert exchanges")
	}

	if !cached {
		exchangeUpsertCacheMut.Lock()
		exchangeUpsertCache[key] = cache
		exchangeUpsertCacheMut.Unlock()
	}

	return nil
}

// DeleteP deletes a single Exchange record with an executor.
// DeleteP will match against the primary key column to find the record to delete.
// Panics on error.
func (o *Exchange) DeleteP(exec boil.Executor) {
	if err := o.Delete(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteG deletes a single Exchange record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *Exchange) DeleteG() error {
	if o == nil {
		return errors.New("models: no Exchange provided for deletion")
	}

	return o.Delete(boil.GetDB())
}

// DeleteGP deletes a single Exchange record.
// DeleteGP will match against the primary key column to find the record to delete.
// Panics on error.
func (o *Exchange) DeleteGP() {
	if err := o.DeleteG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Delete deletes a single Exchange record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Exchange) Delete(exec boil.Executor) error {
	if o == nil {
		return errors.New("models: no Exchange provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), exchangePrimaryKeyMapping)
	sql := "DELETE FROM \"exchanges\" WHERE \"exchange_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to delete from exchanges")
	}

	return nil
}

// DeleteAllP deletes all rows, and panics on error.
func (q exchangeQuery) DeleteAllP() {
	if err := q.DeleteAll(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAll deletes all matching rows.
func (q exchangeQuery) DeleteAll() error {
	if q.Query == nil {
		return errors.New("models: no exchangeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	_, err := q.Query.Exec()
	if err != nil {
		return errors.Wrap(err, "models: unable to delete all from exchanges")
	}

	return nil
}

// DeleteAllGP deletes all rows in the slice, and panics on error.
func (o ExchangeSlice) DeleteAllGP() {
	if err := o.DeleteAllG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAllG deletes all rows in the slice.
func (o ExchangeSlice) DeleteAllG() error {
	if o == nil {
		return errors.New("models: no Exchange slice provided for delete all")
	}
	return o.DeleteAll(boil.GetDB())
}

// DeleteAllP deletes all rows in the slice, using an executor, and panics on error.
func (o ExchangeSlice) DeleteAllP(exec boil.Executor) {
	if err := o.DeleteAll(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ExchangeSlice) DeleteAll(exec boil.Executor) error {
	if o == nil {
		return errors.New("models: no Exchange slice provided for delete all")
	}

	if len(o) == 0 {
		return nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), exchangePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"exchanges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, exchangePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to delete all from exchange slice")
	}

	return nil
}

// ReloadGP refetches the object from the database and panics on error.
func (o *Exchange) ReloadGP() {
	if err := o.ReloadG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadP refetches the object from the database with an executor. Panics on error.
func (o *Exchange) ReloadP(exec boil.Executor) {
	if err := o.Reload(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadG refetches the object from the database using the primary keys.
func (o *Exchange) ReloadG() error {
	if o == nil {
		return errors.New("models: no Exchange provided for reload")
	}

	return o.Reload(boil.GetDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Exchange) Reload(exec boil.Executor) error {
	ret, err := FindExchange(exec, o.ExchangeID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllGP refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
// Panics on error.
func (o *ExchangeSlice) ReloadAllGP() {
	if err := o.ReloadAllG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadAllP refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
// Panics on error.
func (o *ExchangeSlice) ReloadAllP(exec boil.Executor) {
	if err := o.ReloadAll(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ExchangeSlice) ReloadAllG() error {
	if o == nil {
		return errors.New("models: empty ExchangeSlice provided for reload all")
	}

	return o.ReloadAll(boil.GetDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ExchangeSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	exchanges := ExchangeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), exchangePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"exchanges\".* FROM \"exchanges\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, exchangePrimaryKeyColumns, len(*o))

	q := queries.Raw(exec, sql, args...)

	err := q.Bind(&exchanges)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in ExchangeSlice")
	}

	*o = exchanges

	return nil
}

// ExchangeExists checks if the Exchange row exists.
func ExchangeExists(exec boil.Executor, exchangeID int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"exchanges\" where \"exchange_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, exchangeID)
	}

	row := exec.QueryRow(sql, exchangeID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if exchanges exists")
	}

	return exists, nil
}

// ExchangeExistsG checks if the Exchange row exists.
func ExchangeExistsG(exchangeID int64) (bool, error) {
	return ExchangeExists(boil.GetDB(), exchangeID)
}

// ExchangeExistsGP checks if the Exchange row exists. Panics on error.
func ExchangeExistsGP(exchangeID int64) bool {
	e, err := ExchangeExists(boil.GetDB(), exchangeID)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}

// ExchangeExistsP checks if the Exchange row exists. Panics on error.
func ExchangeExistsP(exec boil.Executor, exchangeID int64) bool {
	e, err := ExchangeExists(exec, exchangeID)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}
