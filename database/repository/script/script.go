package script

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/thrasher-corp/gocryptotrader/database"
	modelPSQL "github.com/thrasher-corp/gocryptotrader/database/models/postgres"
	modelSQLite "github.com/thrasher-corp/gocryptotrader/database/models/sqlite3"
	"github.com/thrasher-corp/gocryptotrader/database/repository"
	log "github.com/thrasher-corp/gocryptotrader/logger"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/volatiletech/null"
)

// Event inserts a new script event into database with execution details (script name time status hash of script)
func Event(id, name, path string, hash null.String, data null.Bytes, executionType, status string, time time.Time) {
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
		query := modelSQLite.ScriptWhere.ScriptID.EQ(id)
		f, errQry := modelSQLite.Scripts(query).Exists(ctx, tx)
		if errQry != nil {
			log.Errorf(log.Global, "query failed: %v", err)
			err = tx.Rollback()
			if err != nil {
				log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
			}
			return
		}
		if !f {
			newUUID, errUUID := uuid.NewV4()
			if errUUID != nil {
				log.Errorf(log.DatabaseMgr, "Failed to generate UUID: %v", err)
				_ = tx.Rollback()
				return
			}

			var tempEvent = modelSQLite.Script{
				ID:         newUUID.String(),
				ScriptID:   id,
				ScriptName: name,
				ScriptPath: path,
				ScriptHash: hash,
			}
			err = tempEvent.Insert(ctx, tx, boil.Infer())
			if err != nil {
				log.Errorf(log.Global, "Event insert failed: %v", err)
				err = tx.Rollback()
				if err != nil {
					log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
				}
				return
			}
		} else {
			var tempEvent = modelSQLite.Script{
				ID: id,
			}
			tempScriptExecution := &modelSQLite.ScriptExecution{
				ScriptID:        id,
				ExecutionTime:   time.UTC().String(),
				ExecutionStatus: status,
				ExecutionType:   executionType,
			}
			fmt.Println(id)
			err = tempEvent.AddScriptExecutions(ctx, tx, true, tempScriptExecution)
			if err != nil {
				log.Errorf(log.Global, "Event insert failed: %v", err)
				err = tx.Rollback()
				if err != nil {
					log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
				}
				return
			}
		}
	} else {
		var tempEvent = modelPSQL.Script{
			ScriptID:   id,
			ScriptName: name,
			ScriptPath: path,
			ScriptHash: hash,
		}
		err = tempEvent.Upsert(ctx, tx, true, []string{"script_id"}, boil.Whitelist("last_executed_at"), boil.Infer())
		if err != nil {
			log.Errorf(log.Global, "Event insert failed: %v", err)
			err = tx.Rollback()
			if err != nil {
				log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
			}
			return
		}

		tempScriptExecution := &modelPSQL.ScriptExecution{
			ExecutionTime:   time.UTC(),
			ExecutionStatus: status,
			ExecutionType:   executionType,
		}

		err = tempEvent.AddScriptExecutions(ctx, tx, true, tempScriptExecution)
		if err != nil {
			log.Errorf(log.Global, "Event insert failed: %v", err)
			err = tx.Rollback()
			if err != nil {
				log.Errorf(log.DatabaseMgr, "Event Transaction rollback failed: %v", err)
			}
			return
		}
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
