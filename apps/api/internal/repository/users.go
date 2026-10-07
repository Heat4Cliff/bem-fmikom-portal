package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type Users struct{ db *sql.DB }

func NewUsers(db *sql.DB) *Users { return &Users{db: db} }

const userSelect = `SELECT u.id, u.name, u.email, u.password_hash, u.role, u.division_id, u.status, u.last_login_at, d.id, d.name
FROM users u LEFT JOIN divisions d ON d.id = u.division_id`

func (repository *Users) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return scanUser(repository.db.QueryRowContext(ctx, userSelect+` WHERE u.email = ?`, email))
}

func (repository *Users) FindByID(ctx context.Context, id int64) (domain.User, error) {
	return scanUser(repository.db.QueryRowContext(ctx, userSelect+` WHERE u.id = ?`, id))
}

func (repository *Users) FindByIDTx(ctx context.Context, tx *sql.Tx, id int64) (domain.User, error) {
	return scanUser(tx.QueryRowContext(ctx, userSelect+` WHERE u.id = ? FOR UPDATE`, id))
}

func (repository *Users) Create(ctx context.Context, tx *sql.Tx, user domain.User) (int64, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO users (name, email, password_hash, role, division_id, status) VALUES (?, ?, ?, ?, ?, ?)`, user.Name, user.Email, user.PasswordHash, user.Role, user.DivisionID, user.Status)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read created user id")
	}
	return id, nil
}

func (repository *Users) SetLastLogin(ctx context.Context, tx *sql.Tx, id int64, at time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE users SET last_login_at = ? WHERE id = ? AND status = 'ACTIVE'`, at.UTC(), id)
	if err != nil {
		return fmt.Errorf("update last login")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row *sql.Row) (domain.User, error) {
	var user domain.User
	var divisionID sql.NullInt64
	var joinedDivisionID sql.NullInt64
	var divisionName sql.NullString
	var lastLogin sql.NullTime
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &divisionID, &user.Status, &lastLogin, &joinedDivisionID, &divisionName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("read user")
	}
	if divisionID.Valid && joinedDivisionID.Valid {
		id := divisionID.Int64
		user.DivisionID = &id
		user.Division = &domain.Division{ID: id, Name: divisionName.String}
	}
	if lastLogin.Valid {
		at := lastLogin.Time.UTC()
		user.LastLoginAt = &at
	}
	return user, nil
}
