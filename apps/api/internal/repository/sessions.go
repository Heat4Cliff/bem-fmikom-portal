package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RefreshSession struct {
	ID          int64
	UserID      int64
	FamilyID    string
	ExpiresAt   time.Time
	RevokedAt   sql.NullTime
	RotatedFrom sql.NullInt64
}

type Sessions struct{ db *sql.DB }

func NewSessions(db *sql.DB) *Sessions { return &Sessions{db: db} }

func (repository *Sessions) FindByHashForUpdate(ctx context.Context, tx *sql.Tx, hash string) (RefreshSession, error) {
	var session RefreshSession
	err := tx.QueryRowContext(ctx, `SELECT id, user_id, family_id, expires_at, revoked_at, rotated_from_id FROM refresh_sessions WHERE token_hash = ? FOR UPDATE`, hash).
		Scan(&session.ID, &session.UserID, &session.FamilyID, &session.ExpiresAt, &session.RevokedAt, &session.RotatedFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			return RefreshSession{}, ErrNotFound
		}
		return RefreshSession{}, fmt.Errorf("read refresh session")
	}
	return session, nil
}

func (repository *Sessions) Create(ctx context.Context, tx *sql.Tx, session RefreshSession, tokenHash, ipAddress, userAgent string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO refresh_sessions (user_id, token_hash, family_id, expires_at, rotated_from_id, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?)`, session.UserID, tokenHash, session.FamilyID, session.ExpiresAt.UTC(), nullableInt64(session.RotatedFrom), nullableString(ipAddress), nullableString(limitString(userAgent, 255))); err != nil {
		return fmt.Errorf("insert refresh session")
	}
	return nil
}

func (repository *Sessions) Revoke(ctx context.Context, tx *sql.Tx, id int64, at time.Time) error {
	if _, err := tx.ExecContext(ctx, `UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, ?) WHERE id = ?`, at.UTC(), id); err != nil {
		return fmt.Errorf("revoke refresh session")
	}
	return nil
}

func (repository *Sessions) RevokeFamily(ctx context.Context, tx *sql.Tx, familyID string, at time.Time) error {
	if _, err := tx.ExecContext(ctx, `UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, ?) WHERE family_id = ?`, at.UTC(), familyID); err != nil {
		return fmt.Errorf("revoke refresh session family")
	}
	return nil
}

func (repository *Sessions) RevokeHash(ctx context.Context, tx *sql.Tx, tokenHash string, at time.Time) (int64, error) {
	var session RefreshSession
	err := tx.QueryRowContext(ctx, `SELECT id, family_id FROM refresh_sessions WHERE token_hash = ? FOR UPDATE`, tokenHash).Scan(&session.ID, &session.FamilyID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read logout refresh session")
	}
	if err := repository.RevokeFamily(ctx, tx, session.FamilyID, at); err != nil {
		return 0, err
	}
	var userID int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id FROM refresh_sessions WHERE id = ?`, session.ID).Scan(&userID); err != nil {
		return 0, fmt.Errorf("read logout actor")
	}
	return userID, nil
}

func nullableInt64(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func limitString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
