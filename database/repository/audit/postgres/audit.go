package audit

import (
	"github.com/thrasher-/gocryptotrader/database"
	"github.com/thrasher-/gocryptotrader/database/models"
	"github.com/thrasher-/gocryptotrader/database/repository/audit"

	log "github.com/thrasher-/gocryptotrader/logger"
)

type auditRepo struct {
}

// Audit returns a new instance of auditRepo
func Audit() audit.Repository {
	return &auditRepo{}
}

// AddEvent writes event to database
// writes are done using a transaction with a rollback on error
func (pg *auditRepo) AddEvent(event *models.AuditEvent) {
	if pg == nil {
		return
	}
	query := `INSERT INTO audit_event (type, identifier, message) VALUES($1, $2, $3)`
	tx, err := database.Conn.SQL.Begin()
	if err != nil {
		return
	}
	_, err = tx.Exec(query, &event.Type, &event.Identifier, &event.Message)
	if err != nil {
		_ = tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log.Errorf(log.Global, "Failed to write audit event: %v\n", err)
		return
	}
}
