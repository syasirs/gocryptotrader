package audit

import (
	"context"
	"errors"

	"github.com/thrasher-corp/gocryptotrader/database"
	modelPSQL "github.com/thrasher-corp/gocryptotrader/database/models/postgres"
	modelSQLite "github.com/thrasher-corp/gocryptotrader/database/models/sqlite"
	"github.com/thrasher-corp/gocryptotrader/database/repository"
	log "github.com/thrasher-corp/gocryptotrader/logger"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// TableTimeFormat Go Time format conversion
const TableTimeFormat = "2006-01-02 15:04:05"

// Event inserts a new audit event to database
func Event(id, msgtype, message string) {
	if database.DB.SQL == nil {
		return
	}

	ctx := context.Background()
	ctx = boil.SkipTimestamps(ctx)

	tx, err := database.DB.SQL.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf(log.Global, "transaction begin failed: %v", err)
		return
	}

	if repository.GetSQLDialect() == database.DBSQLite3 {
		var tempEvent = modelSQLite.AuditEvent{
			Type:       msgtype,
			Identifier: id,
			Message:    message,
		}
		err = tempEvent.Insert(ctx, tx, boil.Blacklist("created_at"))
	} else {
		var tempEvent = modelPSQL.AuditEvent{
			Type:       msgtype,
			Identifier: id,
			Message:    message,
		}
		err = tempEvent.Insert(ctx, tx, boil.Blacklist("created_at"))
	}

	if err != nil {
		log.Errorf(log.Global, "insert failed: %v", err)
		err = tx.Rollback()
		if err != nil {
			log.Errorf(log.Global, "transaction rollback failed: %v", err)
		}
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf(log.Global, "transaction commit failed: %v", err)
		err = tx.Rollback()
		if err != nil {
			log.Errorf(log.Global, "transaction rollback failed: %v", err)
		}
		return
	}
}

// GetEvent () returns list of order events matching query
func GetEvent(starTime, endTime, order string, limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, errors.New("database is nil")
	}

	query := qm.Where("created_at BETWEEN ? AND ?", starTime, endTime)

	orderByQueryString := "id"
	if order == "desc" {
		orderByQueryString += " desc"
	}

	orderByQuery := qm.OrderBy(orderByQueryString)
	limitQuery := qm.Limit(limit)

	ctx := context.Background()
	if repository.GetSQLDialect() == database.DBSQLite3 {
		return modelSQLite.AuditEvents(query, orderByQuery, limitQuery).All(ctx, database.DB.SQL)
	}

	return modelPSQL.AuditEvents(query, orderByQuery, limitQuery).All(ctx, database.DB.SQL)
}
