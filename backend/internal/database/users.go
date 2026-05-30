package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID           int64     `db:"id"              json:"id"`
	Email        string    `db:"email"            json:"email"`
	Username     string    `db:"username"         json:"username"`
	PasswordHash string    `db:"password_hash"    json:"-"`
	XP           int64     `db:"xp"               json:"xp"`
	Level        int       `db:"level"            json:"level"`
	Role         string    `db:"role"             json:"role"`
	Status       string    `db:"status"           json:"status"`
	TokenVersion int       `db:"token_version"    json:"-"`
	CreatedAt    time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"       json:"updated_at"`
}

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) *UserStore { return &UserStore{db: db} }

func (s *UserStore) Create(ctx context.Context, email, username, passwordHash string) (*User, error) {
	var u User
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, password_hash, xp, level, role, token_version, created_at, updated_at
	`, email, username, passwordHash).StructScan(&u)
	return &u, err
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := s.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	return &u, err
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := s.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	return &u, err
}

func (s *UserStore) AddXP(ctx context.Context, userID, xpAmount int64, newLevel int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET xp = xp + $1, level = $2, updated_at = NOW() WHERE id = $3
	`, xpAmount, newLevel, userID)
	return err
}

func (s *UserStore) GetTokenVersion(ctx context.Context, id int64) (int, error) {
	var v int
	err := s.db.QueryRowxContext(ctx, `SELECT token_version FROM users WHERE id = $1`, id).Scan(&v)
	return v, err
}

func (s *UserStore) IncrementTokenVersion(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (s *UserStore) ListAll(ctx context.Context) ([]*User, error) {
	var users []*User
	err := s.db.SelectContext(ctx, &users, `SELECT * FROM users ORDER BY id`)
	return users, err
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *UserStore) GetStatus(ctx context.Context, id int64) (string, error) {
	var status string
	err := s.db.QueryRowxContext(ctx, `SELECT status FROM users WHERE id = $1`, id).Scan(&status)
	return status, err
}

func (s *UserStore) ApproveUser(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE users SET status = 'active', updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ActivateAdmin sets status=active and role=admin for the given email.
// Used by the bootstrap admin flow on startup. Safe to call when the user
// is already active/admin — it is a no-op in that case.
func (s *UserStore) ActivateAdmin(ctx context.Context, email string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET status = 'active', role = 'admin', updated_at = NOW() WHERE email = $1
	`, email)
	return err
}

func (s *UserStore) UpdateRole(ctx context.Context, id int64, role string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2
	`, role, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
