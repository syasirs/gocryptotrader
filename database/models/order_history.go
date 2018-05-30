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
)

// OrderHistory is an object representing the database table.
type OrderHistory struct {
	OrderHistoryID int64     `boil:"order_history_id" json:"order_history_id" toml:"order_history_id" yaml:"order_history_id"`
	ConfigID       int64     `boil:"config_id" json:"config_id" toml:"config_id" yaml:"config_id"`
	ExchangeID     string    `boil:"exchange_id" json:"exchange_id" toml:"exchange_id" yaml:"exchange_id"`
	FulfilledOn    time.Time `boil:"fulfilled_on" json:"fulfilled_on" toml:"fulfilled_on" yaml:"fulfilled_on"`
	CurrencyPair   string    `boil:"currency_pair" json:"currency_pair" toml:"currency_pair" yaml:"currency_pair"`
	AssetType      string    `boil:"asset_type" json:"asset_type" toml:"asset_type" yaml:"asset_type"`
	OrderType      string    `boil:"order_type" json:"order_type" toml:"order_type" yaml:"order_type"`
	Amount         float64   `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	Rate           float64   `boil:"rate" json:"rate" toml:"rate" yaml:"rate"`

	R *orderHistoryR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L orderHistoryL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var OrderHistoryColumns = struct {
	OrderHistoryID string
	ConfigID       string
	ExchangeID     string
	FulfilledOn    string
	CurrencyPair   string
	AssetType      string
	OrderType      string
	Amount         string
	Rate           string
}{
	OrderHistoryID: "order_history_id",
	ConfigID:       "config_id",
	ExchangeID:     "exchange_id",
	FulfilledOn:    "fulfilled_on",
	CurrencyPair:   "currency_pair",
	AssetType:      "asset_type",
	OrderType:      "order_type",
	Amount:         "amount",
	Rate:           "rate",
}

// orderHistoryR is where relationships are stored.
type orderHistoryR struct {
	Config *Config
}

// orderHistoryL is where Load methods for each relationship are stored.
type orderHistoryL struct{}

var (
	orderHistoryColumns               = []string{"order_history_id", "config_id", "exchange_id", "fulfilled_on", "currency_pair", "asset_type", "order_type", "amount", "rate"}
	orderHistoryColumnsWithoutDefault = []string{"order_history_id", "config_id", "exchange_id", "fulfilled_on", "currency_pair", "asset_type", "order_type", "amount", "rate"}
	orderHistoryColumnsWithDefault    = []string{}
	orderHistoryPrimaryKeyColumns     = []string{"order_history_id"}
)

type (
	// OrderHistorySlice is an alias for a slice of pointers to OrderHistory.
	// This should generally be used opposed to []OrderHistory.
	OrderHistorySlice []*OrderHistory

	orderHistoryQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	orderHistoryType                 = reflect.TypeOf(&OrderHistory{})
	orderHistoryMapping              = queries.MakeStructMapping(orderHistoryType)
	orderHistoryPrimaryKeyMapping, _ = queries.BindMapping(orderHistoryType, orderHistoryMapping, orderHistoryPrimaryKeyColumns)
	orderHistoryInsertCacheMut       sync.RWMutex
	orderHistoryInsertCache          = make(map[string]insertCache)
	orderHistoryUpdateCacheMut       sync.RWMutex
	orderHistoryUpdateCache          = make(map[string]updateCache)
	orderHistoryUpsertCacheMut       sync.RWMutex
	orderHistoryUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force bytes in case of primary key column that uses []byte (for relationship compares)
	_ = bytes.MinRead
)

// OneP returns a single orderHistory record from the query, and panics on error.
func (q orderHistoryQuery) OneP() *OrderHistory {
	o, err := q.One()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return o
}

// One returns a single orderHistory record from the query.
func (q orderHistoryQuery) One() (*OrderHistory, error) {
	o := &OrderHistory{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for order_history")
	}

	return o, nil
}

// AllP returns all OrderHistory records from the query, and panics on error.
func (q orderHistoryQuery) AllP() OrderHistorySlice {
	o, err := q.All()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return o
}

// All returns all OrderHistory records from the query.
func (q orderHistoryQuery) All() (OrderHistorySlice, error) {
	var o []*OrderHistory

	err := q.Bind(&o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to OrderHistory slice")
	}

	return o, nil
}

// CountP returns the count of all OrderHistory records in the query, and panics on error.
func (q orderHistoryQuery) CountP() int64 {
	c, err := q.Count()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return c
}

// Count returns the count of all OrderHistory records in the query.
func (q orderHistoryQuery) Count() (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow().Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count order_history rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table, and panics on error.
func (q orderHistoryQuery) ExistsP() bool {
	e, err := q.Exists()
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}

// Exists checks if the row exists in the table.
func (q orderHistoryQuery) Exists() (bool, error) {
	var count int64

	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow().Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if order_history exists")
	}

	return count > 0, nil
}

// ConfigG pointed to by the foreign key.
func (o *OrderHistory) ConfigG(mods ...qm.QueryMod) configQuery {
	return o.Config(boil.GetDB(), mods...)
}

// Config pointed to by the foreign key.
func (o *OrderHistory) Config(exec boil.Executor, mods ...qm.QueryMod) configQuery {
	queryMods := []qm.QueryMod{
		qm.Where("config_id=?", o.ConfigID),
	}

	queryMods = append(queryMods, mods...)

	query := Configs(exec, queryMods...)
	queries.SetFrom(query.Query, "\"config\"")

	return query
} // LoadConfig allows an eager lookup of values, cached into the
// loaded structs of the objects.
func (orderHistoryL) LoadConfig(e boil.Executor, singular bool, maybeOrderHistory interface{}) error {
	var slice []*OrderHistory
	var object *OrderHistory

	count := 1
	if singular {
		object = maybeOrderHistory.(*OrderHistory)
	} else {
		slice = *maybeOrderHistory.(*[]*OrderHistory)
		count = len(slice)
	}

	args := make([]interface{}, count)
	if singular {
		if object.R == nil {
			object.R = &orderHistoryR{}
		}
		args[0] = object.ConfigID
	} else {
		for i, obj := range slice {
			if obj.R == nil {
				obj.R = &orderHistoryR{}
			}
			args[i] = obj.ConfigID
		}
	}

	query := fmt.Sprintf(
		"select * from \"config\" where \"config_id\" in (%s)",
		strmangle.Placeholders(dialect.IndexPlaceholders, count, 1, 1),
	)

	if boil.DebugMode {
		fmt.Fprintf(boil.DebugWriter, "%s\n%v\n", query, args)
	}

	results, err := e.Query(query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Config")
	}
	defer results.Close()

	var resultSlice []*Config
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Config")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		object.R.Config = resultSlice[0]
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ConfigID == foreign.ConfigID {
				local.R.Config = foreign
				break
			}
		}
	}

	return nil
}

// SetConfigG of the order_history to the related item.
// Sets o.R.Config to related.
// Adds o to related.R.OrderHistories.
// Uses the global database handle.
func (o *OrderHistory) SetConfigG(insert bool, related *Config) error {
	return o.SetConfig(boil.GetDB(), insert, related)
}

// SetConfigP of the order_history to the related item.
// Sets o.R.Config to related.
// Adds o to related.R.OrderHistories.
// Panics on error.
func (o *OrderHistory) SetConfigP(exec boil.Executor, insert bool, related *Config) {
	if err := o.SetConfig(exec, insert, related); err != nil {
		panic(boil.WrapErr(err))
	}
}

// SetConfigGP of the order_history to the related item.
// Sets o.R.Config to related.
// Adds o to related.R.OrderHistories.
// Uses the global database handle and panics on error.
func (o *OrderHistory) SetConfigGP(insert bool, related *Config) {
	if err := o.SetConfig(boil.GetDB(), insert, related); err != nil {
		panic(boil.WrapErr(err))
	}
}

// SetConfig of the order_history to the related item.
// Sets o.R.Config to related.
// Adds o to related.R.OrderHistories.
func (o *OrderHistory) SetConfig(exec boil.Executor, insert bool, related *Config) error {
	var err error
	if insert {
		if err = related.Insert(exec); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"order_history\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"config_id"}),
		strmangle.WhereClause("\"", "\"", 2, orderHistoryPrimaryKeyColumns),
	)
	values := []interface{}{related.ConfigID, o.OrderHistoryID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ConfigID = related.ConfigID

	if o.R == nil {
		o.R = &orderHistoryR{
			Config: related,
		}
	} else {
		o.R.Config = related
	}

	if related.R == nil {
		related.R = &configR{
			OrderHistories: OrderHistorySlice{o},
		}
	} else {
		related.R.OrderHistories = append(related.R.OrderHistories, o)
	}

	return nil
}

// OrderHistoriesG retrieves all records.
func OrderHistoriesG(mods ...qm.QueryMod) orderHistoryQuery {
	return OrderHistories(boil.GetDB(), mods...)
}

// OrderHistories retrieves all the records using an executor.
func OrderHistories(exec boil.Executor, mods ...qm.QueryMod) orderHistoryQuery {
	mods = append(mods, qm.From("\"order_history\""))
	return orderHistoryQuery{NewQuery(exec, mods...)}
}

// FindOrderHistoryG retrieves a single record by ID.
func FindOrderHistoryG(orderHistoryID int64, selectCols ...string) (*OrderHistory, error) {
	return FindOrderHistory(boil.GetDB(), orderHistoryID, selectCols...)
}

// FindOrderHistoryGP retrieves a single record by ID, and panics on error.
func FindOrderHistoryGP(orderHistoryID int64, selectCols ...string) *OrderHistory {
	retobj, err := FindOrderHistory(boil.GetDB(), orderHistoryID, selectCols...)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return retobj
}

// FindOrderHistory retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindOrderHistory(exec boil.Executor, orderHistoryID int64, selectCols ...string) (*OrderHistory, error) {
	orderHistoryObj := &OrderHistory{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"order_history\" where \"order_history_id\"=$1", sel,
	)

	q := queries.Raw(exec, query, orderHistoryID)

	err := q.Bind(orderHistoryObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from order_history")
	}

	return orderHistoryObj, nil
}

// FindOrderHistoryP retrieves a single record by ID with an executor, and panics on error.
func FindOrderHistoryP(exec boil.Executor, orderHistoryID int64, selectCols ...string) *OrderHistory {
	retobj, err := FindOrderHistory(exec, orderHistoryID, selectCols...)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return retobj
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *OrderHistory) InsertG(whitelist ...string) error {
	return o.Insert(boil.GetDB(), whitelist...)
}

// InsertGP a single record, and panics on error. See Insert for whitelist
// behavior description.
func (o *OrderHistory) InsertGP(whitelist ...string) {
	if err := o.Insert(boil.GetDB(), whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// InsertP a single record using an executor, and panics on error. See Insert
// for whitelist behavior description.
func (o *OrderHistory) InsertP(exec boil.Executor, whitelist ...string) {
	if err := o.Insert(exec, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Insert a single record using an executor.
// Whitelist behavior: If a whitelist is provided, only those columns supplied are inserted
// No whitelist behavior: Without a whitelist, columns are inferred by the following rules:
// - All columns without a default value are included (i.e. name, age)
// - All columns with a default, but non-zero are included (i.e. health = 75)
func (o *OrderHistory) Insert(exec boil.Executor, whitelist ...string) error {
	if o == nil {
		return errors.New("models: no order_history provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(orderHistoryColumnsWithDefault, o)

	key := makeCacheKey(whitelist, nzDefaults)
	orderHistoryInsertCacheMut.RLock()
	cache, cached := orderHistoryInsertCache[key]
	orderHistoryInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := strmangle.InsertColumnSet(
			orderHistoryColumns,
			orderHistoryColumnsWithDefault,
			orderHistoryColumnsWithoutDefault,
			nzDefaults,
			whitelist,
		)

		cache.valueMapping, err = queries.BindMapping(orderHistoryType, orderHistoryMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(orderHistoryType, orderHistoryMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"order_history\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.IndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"order_history\" DEFAULT VALUES"
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
		return errors.Wrap(err, "models: unable to insert into order_history")
	}

	if !cached {
		orderHistoryInsertCacheMut.Lock()
		orderHistoryInsertCache[key] = cache
		orderHistoryInsertCacheMut.Unlock()
	}

	return nil
}

// UpdateG a single OrderHistory record. See Update for
// whitelist behavior description.
func (o *OrderHistory) UpdateG(whitelist ...string) error {
	return o.Update(boil.GetDB(), whitelist...)
}

// UpdateGP a single OrderHistory record.
// UpdateGP takes a whitelist of column names that should be updated.
// Panics on error. See Update for whitelist behavior description.
func (o *OrderHistory) UpdateGP(whitelist ...string) {
	if err := o.Update(boil.GetDB(), whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateP uses an executor to update the OrderHistory, and panics on error.
// See Update for whitelist behavior description.
func (o *OrderHistory) UpdateP(exec boil.Executor, whitelist ...string) {
	err := o.Update(exec, whitelist...)
	if err != nil {
		panic(boil.WrapErr(err))
	}
}

// Update uses an executor to update the OrderHistory.
// Whitelist behavior: If a whitelist is provided, only the columns given are updated.
// No whitelist behavior: Without a whitelist, columns are inferred by the following rules:
// - All columns are inferred to start with
// - All primary keys are subtracted from this set
// Update does not automatically update the record in case of default values. Use .Reload()
// to refresh the records.
func (o *OrderHistory) Update(exec boil.Executor, whitelist ...string) error {
	var err error
	key := makeCacheKey(whitelist, nil)
	orderHistoryUpdateCacheMut.RLock()
	cache, cached := orderHistoryUpdateCache[key]
	orderHistoryUpdateCacheMut.RUnlock()

	if !cached {
		wl := strmangle.UpdateColumnSet(
			orderHistoryColumns,
			orderHistoryPrimaryKeyColumns,
			whitelist,
		)

		if len(whitelist) == 0 {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return errors.New("models: unable to update order_history, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"order_history\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, orderHistoryPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(orderHistoryType, orderHistoryMapping, append(wl, orderHistoryPrimaryKeyColumns...))
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
		return errors.Wrap(err, "models: unable to update order_history row")
	}

	if !cached {
		orderHistoryUpdateCacheMut.Lock()
		orderHistoryUpdateCache[key] = cache
		orderHistoryUpdateCacheMut.Unlock()
	}

	return nil
}

// UpdateAllP updates all rows with matching column names, and panics on error.
func (q orderHistoryQuery) UpdateAllP(cols M) {
	if err := q.UpdateAll(cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAll updates all rows with the specified column values.
func (q orderHistoryQuery) UpdateAll(cols M) error {
	queries.SetUpdate(q.Query, cols)

	_, err := q.Query.Exec()
	if err != nil {
		return errors.Wrap(err, "models: unable to update all for order_history")
	}

	return nil
}

// UpdateAllG updates all rows with the specified column values.
func (o OrderHistorySlice) UpdateAllG(cols M) error {
	return o.UpdateAll(boil.GetDB(), cols)
}

// UpdateAllGP updates all rows with the specified column values, and panics on error.
func (o OrderHistorySlice) UpdateAllGP(cols M) {
	if err := o.UpdateAll(boil.GetDB(), cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAllP updates all rows with the specified column values, and panics on error.
func (o OrderHistorySlice) UpdateAllP(exec boil.Executor, cols M) {
	if err := o.UpdateAll(exec, cols); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o OrderHistorySlice) UpdateAll(exec boil.Executor, cols M) error {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), orderHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"order_history\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, orderHistoryPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to update all in orderHistory slice")
	}

	return nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *OrderHistory) UpsertG(updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) error {
	return o.Upsert(boil.GetDB(), updateOnConflict, conflictColumns, updateColumns, whitelist...)
}

// UpsertGP attempts an insert, and does an update or ignore on conflict. Panics on error.
func (o *OrderHistory) UpsertGP(updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) {
	if err := o.Upsert(boil.GetDB(), updateOnConflict, conflictColumns, updateColumns, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// UpsertP attempts an insert using an executor, and does an update or ignore on conflict.
// UpsertP panics on error.
func (o *OrderHistory) UpsertP(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) {
	if err := o.Upsert(exec, updateOnConflict, conflictColumns, updateColumns, whitelist...); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
func (o *OrderHistory) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns []string, whitelist ...string) error {
	if o == nil {
		return errors.New("models: no order_history provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(orderHistoryColumnsWithDefault, o)

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

	orderHistoryUpsertCacheMut.RLock()
	cache, cached := orderHistoryUpsertCache[key]
	orderHistoryUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := strmangle.InsertColumnSet(
			orderHistoryColumns,
			orderHistoryColumnsWithDefault,
			orderHistoryColumnsWithoutDefault,
			nzDefaults,
			whitelist,
		)

		update := strmangle.UpdateColumnSet(
			orderHistoryColumns,
			orderHistoryPrimaryKeyColumns,
			updateColumns,
		)
		if len(update) == 0 {
			return errors.New("models: unable to upsert order_history, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(orderHistoryPrimaryKeyColumns))
			copy(conflict, orderHistoryPrimaryKeyColumns)
		}
		cache.query = queries.BuildUpsertQueryPostgres(dialect, "\"order_history\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(orderHistoryType, orderHistoryMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(orderHistoryType, orderHistoryMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert order_history")
	}

	if !cached {
		orderHistoryUpsertCacheMut.Lock()
		orderHistoryUpsertCache[key] = cache
		orderHistoryUpsertCacheMut.Unlock()
	}

	return nil
}

// DeleteP deletes a single OrderHistory record with an executor.
// DeleteP will match against the primary key column to find the record to delete.
// Panics on error.
func (o *OrderHistory) DeleteP(exec boil.Executor) {
	if err := o.Delete(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteG deletes a single OrderHistory record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *OrderHistory) DeleteG() error {
	if o == nil {
		return errors.New("models: no OrderHistory provided for deletion")
	}

	return o.Delete(boil.GetDB())
}

// DeleteGP deletes a single OrderHistory record.
// DeleteGP will match against the primary key column to find the record to delete.
// Panics on error.
func (o *OrderHistory) DeleteGP() {
	if err := o.DeleteG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// Delete deletes a single OrderHistory record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *OrderHistory) Delete(exec boil.Executor) error {
	if o == nil {
		return errors.New("models: no OrderHistory provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), orderHistoryPrimaryKeyMapping)
	sql := "DELETE FROM \"order_history\" WHERE \"order_history_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to delete from order_history")
	}

	return nil
}

// DeleteAllP deletes all rows, and panics on error.
func (q orderHistoryQuery) DeleteAllP() {
	if err := q.DeleteAll(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAll deletes all matching rows.
func (q orderHistoryQuery) DeleteAll() error {
	if q.Query == nil {
		return errors.New("models: no orderHistoryQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	_, err := q.Query.Exec()
	if err != nil {
		return errors.Wrap(err, "models: unable to delete all from order_history")
	}

	return nil
}

// DeleteAllGP deletes all rows in the slice, and panics on error.
func (o OrderHistorySlice) DeleteAllGP() {
	if err := o.DeleteAllG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAllG deletes all rows in the slice.
func (o OrderHistorySlice) DeleteAllG() error {
	if o == nil {
		return errors.New("models: no OrderHistory slice provided for delete all")
	}
	return o.DeleteAll(boil.GetDB())
}

// DeleteAllP deletes all rows in the slice, using an executor, and panics on error.
func (o OrderHistorySlice) DeleteAllP(exec boil.Executor) {
	if err := o.DeleteAll(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o OrderHistorySlice) DeleteAll(exec boil.Executor) error {
	if o == nil {
		return errors.New("models: no OrderHistory slice provided for delete all")
	}

	if len(o) == 0 {
		return nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), orderHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"order_history\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, orderHistoryPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	_, err := exec.Exec(sql, args...)
	if err != nil {
		return errors.Wrap(err, "models: unable to delete all from orderHistory slice")
	}

	return nil
}

// ReloadGP refetches the object from the database and panics on error.
func (o *OrderHistory) ReloadGP() {
	if err := o.ReloadG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadP refetches the object from the database with an executor. Panics on error.
func (o *OrderHistory) ReloadP(exec boil.Executor) {
	if err := o.Reload(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadG refetches the object from the database using the primary keys.
func (o *OrderHistory) ReloadG() error {
	if o == nil {
		return errors.New("models: no OrderHistory provided for reload")
	}

	return o.Reload(boil.GetDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *OrderHistory) Reload(exec boil.Executor) error {
	ret, err := FindOrderHistory(exec, o.OrderHistoryID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllGP refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
// Panics on error.
func (o *OrderHistorySlice) ReloadAllGP() {
	if err := o.ReloadAllG(); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadAllP refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
// Panics on error.
func (o *OrderHistorySlice) ReloadAllP(exec boil.Executor) {
	if err := o.ReloadAll(exec); err != nil {
		panic(boil.WrapErr(err))
	}
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *OrderHistorySlice) ReloadAllG() error {
	if o == nil {
		return errors.New("models: empty OrderHistorySlice provided for reload all")
	}

	return o.ReloadAll(boil.GetDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *OrderHistorySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	orderHistories := OrderHistorySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), orderHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"order_history\".* FROM \"order_history\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, orderHistoryPrimaryKeyColumns, len(*o))

	q := queries.Raw(exec, sql, args...)

	err := q.Bind(&orderHistories)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in OrderHistorySlice")
	}

	*o = orderHistories

	return nil
}

// OrderHistoryExists checks if the OrderHistory row exists.
func OrderHistoryExists(exec boil.Executor, orderHistoryID int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"order_history\" where \"order_history_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, orderHistoryID)
	}

	row := exec.QueryRow(sql, orderHistoryID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if order_history exists")
	}

	return exists, nil
}

// OrderHistoryExistsG checks if the OrderHistory row exists.
func OrderHistoryExistsG(orderHistoryID int64) (bool, error) {
	return OrderHistoryExists(boil.GetDB(), orderHistoryID)
}

// OrderHistoryExistsGP checks if the OrderHistory row exists. Panics on error.
func OrderHistoryExistsGP(orderHistoryID int64) bool {
	e, err := OrderHistoryExists(boil.GetDB(), orderHistoryID)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}

// OrderHistoryExistsP checks if the OrderHistory row exists. Panics on error.
func OrderHistoryExistsP(exec boil.Executor, orderHistoryID int64) bool {
	e, err := OrderHistoryExists(exec, orderHistoryID)
	if err != nil {
		panic(boil.WrapErr(err))
	}

	return e
}
