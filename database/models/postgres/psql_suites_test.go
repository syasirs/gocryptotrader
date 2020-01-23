// Code generated by SQLBoiler 3.5.0-gct (https://github.com/thrasher-corp/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package postgres

import "testing"

func TestUpsert(t *testing.T) {
	t.Run("AuditEvents", testAuditEventsUpsert)

	t.Run("Scripts", testScriptsUpsert)

	t.Run("ScriptExecutions", testScriptExecutionsUpsert)

	t.Run("WithdrawalCryptos", testWithdrawalCryptosUpsert)

	t.Run("WithdrawalFiats", testWithdrawalFiatsUpsert)

	t.Run("WithdrawalHistories", testWithdrawalHistoriesUpsert)
}
