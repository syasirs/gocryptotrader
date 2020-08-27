// Code generated by SQLBoiler 3.5.1-gct (https://github.com/thrasher-corp/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package postgres

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/randomize"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testTrades(t *testing.T) {
	t.Parallel()

	query := Trades()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testTradesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testTradesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Trades().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testTradesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := TradeSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testTradesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := TradeExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if Trade exists: %s", err)
	}
	if !e {
		t.Errorf("Expected TradeExists to return true, but got false.")
	}
}

func testTradesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	tradeFound, err := FindTrade(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if tradeFound == nil {
		t.Error("want a record, got nil")
	}
}

func testTradesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Trades().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testTradesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Trades().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testTradesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	tradeOne := &Trade{}
	tradeTwo := &Trade{}
	if err = randomize.Struct(seed, tradeOne, tradeDBTypes, false, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}
	if err = randomize.Struct(seed, tradeTwo, tradeDBTypes, false, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = tradeOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = tradeTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Trades().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testTradesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	tradeOne := &Trade{}
	tradeTwo := &Trade{}
	if err = randomize.Struct(seed, tradeOne, tradeDBTypes, false, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}
	if err = randomize.Struct(seed, tradeTwo, tradeDBTypes, false, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = tradeOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = tradeTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func tradeBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func tradeAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Trade) error {
	*o = Trade{}
	return nil
}

func testTradesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Trade{}
	o := &Trade{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, tradeDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Trade object: %s", err)
	}

	AddTradeHook(boil.BeforeInsertHook, tradeBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	tradeBeforeInsertHooks = []TradeHook{}

	AddTradeHook(boil.AfterInsertHook, tradeAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	tradeAfterInsertHooks = []TradeHook{}

	AddTradeHook(boil.AfterSelectHook, tradeAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	tradeAfterSelectHooks = []TradeHook{}

	AddTradeHook(boil.BeforeUpdateHook, tradeBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	tradeBeforeUpdateHooks = []TradeHook{}

	AddTradeHook(boil.AfterUpdateHook, tradeAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	tradeAfterUpdateHooks = []TradeHook{}

	AddTradeHook(boil.BeforeDeleteHook, tradeBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	tradeBeforeDeleteHooks = []TradeHook{}

	AddTradeHook(boil.AfterDeleteHook, tradeAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	tradeAfterDeleteHooks = []TradeHook{}

	AddTradeHook(boil.BeforeUpsertHook, tradeBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	tradeBeforeUpsertHooks = []TradeHook{}

	AddTradeHook(boil.AfterUpsertHook, tradeAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	tradeAfterUpsertHooks = []TradeHook{}
}

func testTradesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testTradesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(tradeColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testTradesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testTradesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := TradeSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testTradesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Trades().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	tradeDBTypes = map[string]string{`ID`: `uuid`, `Tid`: `character varying`, `ExchangeID`: `uuid`, `Currency`: `character varying`, `Asset`: `character varying`, `Price`: `double precision`, `Amount`: `double precision`, `Side`: `character varying`, `Timestamp`: `bigint`}
	_            = bytes.MinRead
)

func testTradesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(tradePrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(tradeAllColumns) == len(tradePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testTradesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(tradeAllColumns) == len(tradePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Trade{}
	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, tradeDBTypes, true, tradePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(tradeAllColumns, tradePrimaryKeyColumns) {
		fields = tradeAllColumns
	} else {
		fields = strmangle.SetComplement(
			tradeAllColumns,
			tradePrimaryKeyColumns,
		)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := TradeSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testTradesUpsert(t *testing.T) {
	t.Parallel()

	if len(tradeAllColumns) == len(tradePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Trade{}
	if err = randomize.Struct(seed, &o, tradeDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Trade: %s", err)
	}

	count, err := Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, tradeDBTypes, false, tradePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Trade struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Trade: %s", err)
	}

	count, err = Trades().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
