// Code generated by SQLBoiler (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
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

// CryptocurrencyPortfolioAddress is an object representing the database table.
type CryptocurrencyPortfolioAddress struct {
	ID          int64  `boil:"id" json:"id" toml:"id" yaml:"id"`
	CoinName    string `boil:"coin_name" json:"coin_name" toml:"coin_name" yaml:"coin_name"`
	CoinAddress string `boil:"coin_address" json:"coin_address" toml:"coin_address" yaml:"coin_address"`
	IsHotWallet bool   `boil:"is_hot_wallet" json:"is_hot_wallet" toml:"is_hot_wallet" yaml:"is_hot_wallet"`
	PortfolioID int64  `boil:"portfolio_id" json:"portfolio_id" toml:"portfolio_id" yaml:"portfolio_id"`

	R *cryptocurrencyPortfolioAddressR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L cryptocurrencyPortfolioAddressL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CryptocurrencyPortfolioAddressColumns = struct {
	ID          string
	CoinName    string
	CoinAddress string
	IsHotWallet string
	PortfolioID string
}{
	ID:          "id",
	CoinName:    "coin_name",
	CoinAddress: "coin_address",
	IsHotWallet: "is_hot_wallet",
	PortfolioID: "portfolio_id",
}

// CryptocurrencyPortfolioAddressRels is where relationship names are stored.
var CryptocurrencyPortfolioAddressRels = struct {
	Portfolio string
}{
	Portfolio: "Portfolio",
}

// cryptocurrencyPortfolioAddressR is where relationships are stored.
type cryptocurrencyPortfolioAddressR struct {
	Portfolio *Portfolio
}

// NewStruct creates a new relationship struct
func (*cryptocurrencyPortfolioAddressR) NewStruct() *cryptocurrencyPortfolioAddressR {
	return &cryptocurrencyPortfolioAddressR{}
}

// cryptocurrencyPortfolioAddressL is where Load methods for each relationship are stored.
type cryptocurrencyPortfolioAddressL struct{}

var (
	cryptocurrencyPortfolioAddressColumns               = []string{"id", "coin_name", "coin_address", "is_hot_wallet", "portfolio_id"}
	cryptocurrencyPortfolioAddressColumnsWithoutDefault = []string{}
	cryptocurrencyPortfolioAddressColumnsWithDefault    = []string{"id", "coin_name", "coin_address", "is_hot_wallet", "portfolio_id"}
	cryptocurrencyPortfolioAddressPrimaryKeyColumns     = []string{"id"}
)

type (
	// CryptocurrencyPortfolioAddressSlice is an alias for a slice of pointers to CryptocurrencyPortfolioAddress.
	// This should generally be used opposed to []CryptocurrencyPortfolioAddress.
	CryptocurrencyPortfolioAddressSlice []*CryptocurrencyPortfolioAddress
	// CryptocurrencyPortfolioAddressHook is the signature for custom CryptocurrencyPortfolioAddress hook methods
	CryptocurrencyPortfolioAddressHook func(context.Context, boil.ContextExecutor, *CryptocurrencyPortfolioAddress) error

	cryptocurrencyPortfolioAddressQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	cryptocurrencyPortfolioAddressType                 = reflect.TypeOf(&CryptocurrencyPortfolioAddress{})
	cryptocurrencyPortfolioAddressMapping              = queries.MakeStructMapping(cryptocurrencyPortfolioAddressType)
	cryptocurrencyPortfolioAddressPrimaryKeyMapping, _ = queries.BindMapping(cryptocurrencyPortfolioAddressType, cryptocurrencyPortfolioAddressMapping, cryptocurrencyPortfolioAddressPrimaryKeyColumns)
	cryptocurrencyPortfolioAddressInsertCacheMut       sync.RWMutex
	cryptocurrencyPortfolioAddressInsertCache          = make(map[string]insertCache)
	cryptocurrencyPortfolioAddressUpdateCacheMut       sync.RWMutex
	cryptocurrencyPortfolioAddressUpdateCache          = make(map[string]updateCache)
	cryptocurrencyPortfolioAddressUpsertCacheMut       sync.RWMutex
	cryptocurrencyPortfolioAddressUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
)

var cryptocurrencyPortfolioAddressBeforeInsertHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressBeforeUpdateHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressBeforeDeleteHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressBeforeUpsertHooks []CryptocurrencyPortfolioAddressHook

var cryptocurrencyPortfolioAddressAfterInsertHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressAfterSelectHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressAfterUpdateHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressAfterDeleteHooks []CryptocurrencyPortfolioAddressHook
var cryptocurrencyPortfolioAddressAfterUpsertHooks []CryptocurrencyPortfolioAddressHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *CryptocurrencyPortfolioAddress) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *CryptocurrencyPortfolioAddress) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *CryptocurrencyPortfolioAddress) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *CryptocurrencyPortfolioAddress) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *CryptocurrencyPortfolioAddress) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *CryptocurrencyPortfolioAddress) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *CryptocurrencyPortfolioAddress) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *CryptocurrencyPortfolioAddress) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *CryptocurrencyPortfolioAddress) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	for _, hook := range cryptocurrencyPortfolioAddressAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddCryptocurrencyPortfolioAddressHook registers your hook function for all future operations.
func AddCryptocurrencyPortfolioAddressHook(hookPoint boil.HookPoint, cryptocurrencyPortfolioAddressHook CryptocurrencyPortfolioAddressHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		cryptocurrencyPortfolioAddressBeforeInsertHooks = append(cryptocurrencyPortfolioAddressBeforeInsertHooks, cryptocurrencyPortfolioAddressHook)
	case boil.BeforeUpdateHook:
		cryptocurrencyPortfolioAddressBeforeUpdateHooks = append(cryptocurrencyPortfolioAddressBeforeUpdateHooks, cryptocurrencyPortfolioAddressHook)
	case boil.BeforeDeleteHook:
		cryptocurrencyPortfolioAddressBeforeDeleteHooks = append(cryptocurrencyPortfolioAddressBeforeDeleteHooks, cryptocurrencyPortfolioAddressHook)
	case boil.BeforeUpsertHook:
		cryptocurrencyPortfolioAddressBeforeUpsertHooks = append(cryptocurrencyPortfolioAddressBeforeUpsertHooks, cryptocurrencyPortfolioAddressHook)
	case boil.AfterInsertHook:
		cryptocurrencyPortfolioAddressAfterInsertHooks = append(cryptocurrencyPortfolioAddressAfterInsertHooks, cryptocurrencyPortfolioAddressHook)
	case boil.AfterSelectHook:
		cryptocurrencyPortfolioAddressAfterSelectHooks = append(cryptocurrencyPortfolioAddressAfterSelectHooks, cryptocurrencyPortfolioAddressHook)
	case boil.AfterUpdateHook:
		cryptocurrencyPortfolioAddressAfterUpdateHooks = append(cryptocurrencyPortfolioAddressAfterUpdateHooks, cryptocurrencyPortfolioAddressHook)
	case boil.AfterDeleteHook:
		cryptocurrencyPortfolioAddressAfterDeleteHooks = append(cryptocurrencyPortfolioAddressAfterDeleteHooks, cryptocurrencyPortfolioAddressHook)
	case boil.AfterUpsertHook:
		cryptocurrencyPortfolioAddressAfterUpsertHooks = append(cryptocurrencyPortfolioAddressAfterUpsertHooks, cryptocurrencyPortfolioAddressHook)
	}
}

// One returns a single cryptocurrencyPortfolioAddress record from the query.
func (q cryptocurrencyPortfolioAddressQuery) One(ctx context.Context, exec boil.ContextExecutor) (*CryptocurrencyPortfolioAddress, error) {
	o := &CryptocurrencyPortfolioAddress{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for cryptocurrency_portfolio_address")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all CryptocurrencyPortfolioAddress records from the query.
func (q cryptocurrencyPortfolioAddressQuery) All(ctx context.Context, exec boil.ContextExecutor) (CryptocurrencyPortfolioAddressSlice, error) {
	var o []*CryptocurrencyPortfolioAddress

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to CryptocurrencyPortfolioAddress slice")
	}

	if len(cryptocurrencyPortfolioAddressAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all CryptocurrencyPortfolioAddress records in the query.
func (q cryptocurrencyPortfolioAddressQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count cryptocurrency_portfolio_address rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q cryptocurrencyPortfolioAddressQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if cryptocurrency_portfolio_address exists")
	}

	return count > 0, nil
}

// Portfolio pointed to by the foreign key.
func (o *CryptocurrencyPortfolioAddress) Portfolio(mods ...qm.QueryMod) portfolioQuery {
	queryMods := []qm.QueryMod{
		qm.Where("id=?", o.PortfolioID),
	}

	queryMods = append(queryMods, mods...)

	query := Portfolios(queryMods...)
	queries.SetFrom(query.Query, "\"portfolio\"")

	return query
}

// LoadPortfolio allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (cryptocurrencyPortfolioAddressL) LoadPortfolio(ctx context.Context, e boil.ContextExecutor, singular bool, maybeCryptocurrencyPortfolioAddress interface{}, mods queries.Applicator) error {
	var slice []*CryptocurrencyPortfolioAddress
	var object *CryptocurrencyPortfolioAddress

	if singular {
		object = maybeCryptocurrencyPortfolioAddress.(*CryptocurrencyPortfolioAddress)
	} else {
		slice = *maybeCryptocurrencyPortfolioAddress.(*[]*CryptocurrencyPortfolioAddress)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &cryptocurrencyPortfolioAddressR{}
		}
		args = append(args, object.PortfolioID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &cryptocurrencyPortfolioAddressR{}
			}

			for _, a := range args {
				if a == obj.PortfolioID {
					continue Outer
				}
			}

			args = append(args, obj.PortfolioID)
		}
	}

	query := NewQuery(qm.From(`portfolio`), qm.WhereIn(`id in ?`, args...))
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Portfolio")
	}

	var resultSlice []*Portfolio
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Portfolio")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for portfolio")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for portfolio")
	}

	if len(cryptocurrencyPortfolioAddressAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Portfolio = foreign
		if foreign.R == nil {
			foreign.R = &portfolioR{}
		}
		foreign.R.CryptocurrencyPortfolioAddresses = append(foreign.R.CryptocurrencyPortfolioAddresses, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PortfolioID == foreign.ID {
				local.R.Portfolio = foreign
				if foreign.R == nil {
					foreign.R = &portfolioR{}
				}
				foreign.R.CryptocurrencyPortfolioAddresses = append(foreign.R.CryptocurrencyPortfolioAddresses, local)
				break
			}
		}
	}

	return nil
}

// SetPortfolio of the cryptocurrencyPortfolioAddress to the related item.
// Sets o.R.Portfolio to related.
// Adds o to related.R.CryptocurrencyPortfolioAddresses.
func (o *CryptocurrencyPortfolioAddress) SetPortfolio(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Portfolio) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"cryptocurrency_portfolio_address\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 0, []string{"portfolio_id"}),
		strmangle.WhereClause("\"", "\"", 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PortfolioID = related.ID
	if o.R == nil {
		o.R = &cryptocurrencyPortfolioAddressR{
			Portfolio: related,
		}
	} else {
		o.R.Portfolio = related
	}

	if related.R == nil {
		related.R = &portfolioR{
			CryptocurrencyPortfolioAddresses: CryptocurrencyPortfolioAddressSlice{o},
		}
	} else {
		related.R.CryptocurrencyPortfolioAddresses = append(related.R.CryptocurrencyPortfolioAddresses, o)
	}

	return nil
}

// CryptocurrencyPortfolioAddresses retrieves all the records using an executor.
func CryptocurrencyPortfolioAddresses(mods ...qm.QueryMod) cryptocurrencyPortfolioAddressQuery {
	mods = append(mods, qm.From("\"cryptocurrency_portfolio_address\""))
	return cryptocurrencyPortfolioAddressQuery{NewQuery(mods...)}
}

// FindCryptocurrencyPortfolioAddress retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCryptocurrencyPortfolioAddress(ctx context.Context, exec boil.ContextExecutor, iD int64, selectCols ...string) (*CryptocurrencyPortfolioAddress, error) {
	cryptocurrencyPortfolioAddressObj := &CryptocurrencyPortfolioAddress{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"cryptocurrency_portfolio_address\" where \"id\"=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, cryptocurrencyPortfolioAddressObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from cryptocurrency_portfolio_address")
	}

	return cryptocurrencyPortfolioAddressObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CryptocurrencyPortfolioAddress) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no cryptocurrency_portfolio_address provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(cryptocurrencyPortfolioAddressColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	cryptocurrencyPortfolioAddressInsertCacheMut.RLock()
	cache, cached := cryptocurrencyPortfolioAddressInsertCache[key]
	cryptocurrencyPortfolioAddressInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			cryptocurrencyPortfolioAddressColumns,
			cryptocurrencyPortfolioAddressColumnsWithDefault,
			cryptocurrencyPortfolioAddressColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(cryptocurrencyPortfolioAddressType, cryptocurrencyPortfolioAddressMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(cryptocurrencyPortfolioAddressType, cryptocurrencyPortfolioAddressMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"cryptocurrency_portfolio_address\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"cryptocurrency_portfolio_address\" () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT \"%s\" FROM \"cryptocurrency_portfolio_address\" WHERE %s", strings.Join(returnColumns, "\",\""), strmangle.WhereClause("\"", "\"", 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into cryptocurrency_portfolio_address")
	}

	var lastID int64
	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = int64(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == cryptocurrencyPortfolioAddressMapping["ID"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ID,
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.retQuery)
		fmt.Fprintln(boil.DebugWriter, identifierCols...)
	}

	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for cryptocurrency_portfolio_address")
	}

CacheNoHooks:
	if !cached {
		cryptocurrencyPortfolioAddressInsertCacheMut.Lock()
		cryptocurrencyPortfolioAddressInsertCache[key] = cache
		cryptocurrencyPortfolioAddressInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the CryptocurrencyPortfolioAddress.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CryptocurrencyPortfolioAddress) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	cryptocurrencyPortfolioAddressUpdateCacheMut.RLock()
	cache, cached := cryptocurrencyPortfolioAddressUpdateCache[key]
	cryptocurrencyPortfolioAddressUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			cryptocurrencyPortfolioAddressColumns,
			cryptocurrencyPortfolioAddressPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update cryptocurrency_portfolio_address, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"cryptocurrency_portfolio_address\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 0, wl),
			strmangle.WhereClause("\"", "\"", 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(cryptocurrencyPortfolioAddressType, cryptocurrencyPortfolioAddressMapping, append(wl, cryptocurrencyPortfolioAddressPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}

	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update cryptocurrency_portfolio_address row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for cryptocurrency_portfolio_address")
	}

	if !cached {
		cryptocurrencyPortfolioAddressUpdateCacheMut.Lock()
		cryptocurrencyPortfolioAddressUpdateCache[key] = cache
		cryptocurrencyPortfolioAddressUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q cryptocurrencyPortfolioAddressQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for cryptocurrency_portfolio_address")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for cryptocurrency_portfolio_address")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CryptocurrencyPortfolioAddressSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptocurrencyPortfolioAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"cryptocurrency_portfolio_address\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in cryptocurrencyPortfolioAddress slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all cryptocurrencyPortfolioAddress")
	}
	return rowsAff, nil
}

// Delete deletes a single CryptocurrencyPortfolioAddress record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CryptocurrencyPortfolioAddress) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no CryptocurrencyPortfolioAddress provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cryptocurrencyPortfolioAddressPrimaryKeyMapping)
	sql := "DELETE FROM \"cryptocurrency_portfolio_address\" WHERE \"id\"=?"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from cryptocurrency_portfolio_address")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for cryptocurrency_portfolio_address")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q cryptocurrencyPortfolioAddressQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no cryptocurrencyPortfolioAddressQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from cryptocurrency_portfolio_address")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for cryptocurrency_portfolio_address")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CryptocurrencyPortfolioAddressSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no CryptocurrencyPortfolioAddress slice provided for delete all")
	}

	if len(o) == 0 {
		return 0, nil
	}

	if len(cryptocurrencyPortfolioAddressBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptocurrencyPortfolioAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"cryptocurrency_portfolio_address\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}

	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from cryptocurrencyPortfolioAddress slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for cryptocurrency_portfolio_address")
	}

	if len(cryptocurrencyPortfolioAddressAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *CryptocurrencyPortfolioAddress) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindCryptocurrencyPortfolioAddress(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CryptocurrencyPortfolioAddressSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CryptocurrencyPortfolioAddressSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptocurrencyPortfolioAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"cryptocurrency_portfolio_address\".* FROM \"cryptocurrency_portfolio_address\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptocurrencyPortfolioAddressPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in CryptocurrencyPortfolioAddressSlice")
	}

	*o = slice

	return nil
}

// CryptocurrencyPortfolioAddressExists checks if the CryptocurrencyPortfolioAddress row exists.
func CryptocurrencyPortfolioAddressExists(ctx context.Context, exec boil.ContextExecutor, iD int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"cryptocurrency_portfolio_address\" where \"id\"=? limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}

	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if cryptocurrency_portfolio_address exists")
	}

	return exists, nil
}
