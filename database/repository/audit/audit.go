package audit

import (
	"github.com/thrasher-corp/gocryptotrader/database"
	"github.com/thrasher-corp/gocryptotrader/database/models"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

// Repository that is required for each driver type to implement
type Repository interface {
	AddEvent(event *models.AuditEvent)
	AddEventTx(event []*models.AuditEvent)
}

var (
	Audit Repository // Global Audit repository
	events []*models.AuditEvent
)

// Event allows you to call audit.Event() as long as the audit repository package without the need to include each driver
func Event(msgType, identifier, message string) {
	if database.Conn.SQL == nil {
		return
	}

	if Audit == nil {
		return
	}

	tempEvent := models.AuditEvent{
		Type:       msgType,
		Identifier: identifier,
		Message:    message}

	poolEvents(&tempEvent)
}

func poolEvents(event *models.AuditEvent) {
	database.Conn.Mu.RLock()
	events = append(events, event)

	if !database.Conn.Connected {
		log.Warnln(log.DatabaseMgr, "connection to database interrupted pooling database write", )
		return
	}

	Audit.AddEventTx(events)
	events = nil
	database.Conn.Mu.RUnlock()
}