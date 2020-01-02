// Code generated by SQLBoiler 3.5.0-gct (https://github.com/thrasher-corp/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package sqlite3

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

func testScriptEvents(t *testing.T) {
	t.Parallel()

	query := ScriptEvents()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testScriptEventsDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
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

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testScriptEventsQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := ScriptEvents().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testScriptEventsSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ScriptEventSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testScriptEventsExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := ScriptEventExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if ScriptEvent exists: %s", err)
	}
	if !e {
		t.Errorf("Expected ScriptEventExists to return true, but got false.")
	}
}

func testScriptEventsFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	scriptEventFound, err := FindScriptEvent(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if scriptEventFound == nil {
		t.Error("want a record, got nil")
	}
}

func testScriptEventsBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = ScriptEvents().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testScriptEventsOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := ScriptEvents().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testScriptEventsAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	scriptEventOne := &ScriptEvent{}
	scriptEventTwo := &ScriptEvent{}
	if err = randomize.Struct(seed, scriptEventOne, scriptEventDBTypes, false, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}
	if err = randomize.Struct(seed, scriptEventTwo, scriptEventDBTypes, false, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = scriptEventOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = scriptEventTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ScriptEvents().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testScriptEventsCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	scriptEventOne := &ScriptEvent{}
	scriptEventTwo := &ScriptEvent{}
	if err = randomize.Struct(seed, scriptEventOne, scriptEventDBTypes, false, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}
	if err = randomize.Struct(seed, scriptEventTwo, scriptEventDBTypes, false, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = scriptEventOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = scriptEventTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func scriptEventBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func scriptEventAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ScriptEvent) error {
	*o = ScriptEvent{}
	return nil
}

func testScriptEventsHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &ScriptEvent{}
	o := &ScriptEvent{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, scriptEventDBTypes, false); err != nil {
		t.Errorf("Unable to randomize ScriptEvent object: %s", err)
	}

	AddScriptEventHook(boil.BeforeInsertHook, scriptEventBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	scriptEventBeforeInsertHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.AfterInsertHook, scriptEventAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	scriptEventAfterInsertHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.AfterSelectHook, scriptEventAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	scriptEventAfterSelectHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.BeforeUpdateHook, scriptEventBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	scriptEventBeforeUpdateHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.AfterUpdateHook, scriptEventAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	scriptEventAfterUpdateHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.BeforeDeleteHook, scriptEventBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	scriptEventBeforeDeleteHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.AfterDeleteHook, scriptEventAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	scriptEventAfterDeleteHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.BeforeUpsertHook, scriptEventBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	scriptEventBeforeUpsertHooks = []ScriptEventHook{}

	AddScriptEventHook(boil.AfterUpsertHook, scriptEventAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	scriptEventAfterUpsertHooks = []ScriptEventHook{}
}

func testScriptEventsInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testScriptEventsInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(scriptEventColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testScriptEventsReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
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

func testScriptEventsReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ScriptEventSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testScriptEventsSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ScriptEvents().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	scriptEventDBTypes = map[string]string{`ID`: `INTEGER`, `ScriptID`: `TEXT`, `ScriptName`: `TEXT`, `ScriptPath`: `TEXT`, `ScriptHash`: `TEXT`, `ExecutionType`: `TEXT`, `ExecutionTime`: `TIMESTAMP`, `ExecutionStatus`: `TEXT`}
	_                  = bytes.MinRead
)

func testScriptEventsUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(scriptEventPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(scriptEventAllColumns) == len(scriptEventPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testScriptEventsSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(scriptEventAllColumns) == len(scriptEventPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ScriptEvent{}
	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ScriptEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, scriptEventDBTypes, true, scriptEventPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ScriptEvent struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(scriptEventAllColumns, scriptEventPrimaryKeyColumns) {
		fields = scriptEventAllColumns
	} else {
		fields = strmangle.SetComplement(
			scriptEventAllColumns,
			scriptEventPrimaryKeyColumns,
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

	slice := ScriptEventSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}
