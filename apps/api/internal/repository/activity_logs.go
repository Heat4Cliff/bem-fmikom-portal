package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type ActivityEvent struct {
	ActorUserID *int64
	Action      string
	EntityType  string
	EntityID    string
	Details     map[string]any
	IPAddress   string
	UserAgent   string
}

type ActivityLogs struct{ db *sql.DB }

func NewActivityLogs(db *sql.DB) *ActivityLogs { return &ActivityLogs{db: db} }

func (repository *ActivityLogs) Write(ctx context.Context, tx *sql.Tx, event ActivityEvent) error {
	var details []byte
	var err error
	if len(event.Details) > 0 {
		details, err = json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("encode activity details")
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO activity_logs (actor_user_id, action, entity_type, entity_id, details_json, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?)`, event.ActorUserID, event.Action, event.EntityType, nullableString(event.EntityID), nullableJSON(details), nullableString(event.IPAddress), nullableString(limitString(event.UserAgent, 255))); err != nil {
		return fmt.Errorf("write activity event")
	}
	return nil
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
