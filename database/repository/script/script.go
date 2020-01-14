package script

import (
	"context"
	"time"

	"github.com/thrasher-corp/gocryptotrader/database"
	modelPSQL "github.com/thrasher-corp/gocryptotrader/database/models/postgres"
	modelSQLite "github.com/thrasher-corp/gocryptotrader/database/models/sqlite3"
	"github.com/thrasher-corp/gocryptotrader/database/repository"
	log "github.com/thrasher-corp/gocryptotrader/logger"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/volatiletech/null"
)

// Event inserts a new script event into database with execution details (script name time status hash of script)
func Event(id, name, path string, hash null.String, executionType, status string, time time.Time) {
	if database.DB.SQL == nil {
		return
	}

	ctx := context.Background()
	ctx = boil.SkipTimestamps(ctx)
	tx, err := database.DB.SQL.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf(log.DatabaseMgr, "Event transaction begin failed: %v", err)
		return
	}

	if repository.GetSQLDialect() == database.DBSQLite3 {
		var tempEvent = modelSQLite.ScriptEvent{
			// ScriptID:        id.String(),
			// ScriptName:      name,
			// ScriptPath:      path,
			// ScriptHash:      hash,
			// ExecutionType:   executionType,
			// ExecutionTime:   time.UTC().String(),
			// ExecutionStatus: status,
		}
		err = tempEvent.Insert(ctx, tx, boil.Infer())
	} else {
		var tempEvent = modelPSQL.ScriptEvent{
			ScriptID:        id,
			ScriptName: name,
			ScriptPath: path,
			ScriptHash: hash,
		}
		err = tempEvent.Upsert(ctx, tx, true, []string{"script_id"}, boil.Whitelist("created_at"), boil.Infer())
	}

	if err != nil {
		log.Errorf(log.Global, "Event insert failed: %v", err)
		err = tx.Rollback()
		if err != nil {
			log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
		}
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf(log.Global, "Event Transaction commit failed: %v", err)
		err = tx.Rollback()
		if err != nil {
			log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
		}
	}
}
