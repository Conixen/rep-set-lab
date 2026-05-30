package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/leonj/rep-set-lab/internal/database"
)

type bootstrapStore interface {
	GetByEmail(ctx context.Context, email string) (*database.User, error)
	Create(ctx context.Context, email, username, passwordHash string) (*database.User, error)
	ActivateAdmin(ctx context.Context, email string) error
}

// BootstrapAdmin ensures the configured admin account exists and is active.
// If the account does not exist it is created. If it exists but is pending or
// not yet admin it is upgraded. Safe to call on every boot — idempotent.
func BootstrapAdmin(ctx context.Context, store bootstrapStore, email, password string, logger *slog.Logger) {
	if email == "" || password == "" {
		return
	}

	user, err := store.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("bootstrap admin: failed to look up account", "error", err)
		return
	}
	if err != nil {
		// User doesn't exist yet — create them.
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("bootstrap admin: failed to hash password", "error", err)
			return
		}
		// Username derived from the local part of the email.
		username := email
		for i, c := range email {
			if c == '@' {
				username = email[:i]
				break
			}
		}
		user, err = store.Create(ctx, email, username, string(hash))
		if err != nil {
			logger.Error("bootstrap admin: failed to create user", "error", err)
			return
		}
		logger.Info("bootstrap admin: account created", "email", email, "id", user.ID)
	}

	if err := store.ActivateAdmin(ctx, email); err != nil {
		logger.Error("bootstrap admin: failed to activate admin", "error", err)
		return
	}

	logger.Info("bootstrap admin: account is active admin", "email", email)
}
