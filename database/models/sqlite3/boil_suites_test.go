// Code generated by SQLBoiler 3.5.0-gct (https://github.com/thrasher-corp/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package sqlite3

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("AuditEvents", testAuditEvents)
	t.Run("ScriptEvents", testScriptEvents)
}

func TestDelete(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsDelete)
	t.Run("ScriptEvents", testScriptEventsDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsQueryDeleteAll)
	t.Run("ScriptEvents", testScriptEventsQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsSliceDeleteAll)
	t.Run("ScriptEvents", testScriptEventsSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsExists)
	t.Run("ScriptEvents", testScriptEventsExists)
}

func TestFind(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsFind)
	t.Run("ScriptEvents", testScriptEventsFind)
}

func TestBind(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsBind)
	t.Run("ScriptEvents", testScriptEventsBind)
}

func TestOne(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsOne)
	t.Run("ScriptEvents", testScriptEventsOne)
}

func TestAll(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsAll)
	t.Run("ScriptEvents", testScriptEventsAll)
}

func TestCount(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsCount)
	t.Run("ScriptEvents", testScriptEventsCount)
}

func TestHooks(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsHooks)
	t.Run("ScriptEvents", testScriptEventsHooks)
}

func TestInsert(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsInsert)
	t.Run("AuditEvents", testAuditEventsInsertWhitelist)
	t.Run("ScriptEvents", testScriptEventsInsert)
	t.Run("ScriptEvents", testScriptEventsInsertWhitelist)
}

// TestToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestToOne(t *testing.T) {}

// TestOneToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOne(t *testing.T) {}

// TestToMany tests cannot be run in parallel
// or deadlocks can occur.
func TestToMany(t *testing.T) {}

// TestToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneSet(t *testing.T) {}

// TestToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneRemove(t *testing.T) {}

// TestOneToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneSet(t *testing.T) {}

// TestOneToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneRemove(t *testing.T) {}

// TestToManyAdd tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyAdd(t *testing.T) {}

// TestToManySet tests cannot be run in parallel
// or deadlocks can occur.
func TestToManySet(t *testing.T) {}

// TestToManyRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyRemove(t *testing.T) {}

func TestReload(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsReload)
	t.Run("ScriptEvents", testScriptEventsReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsReloadAll)
	t.Run("ScriptEvents", testScriptEventsReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsSelect)
	t.Run("ScriptEvents", testScriptEventsSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsUpdate)
	t.Run("ScriptEvents", testScriptEventsUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsSliceUpdateAll)
	t.Run("ScriptEvents", testScriptEventsSliceUpdateAll)
}
