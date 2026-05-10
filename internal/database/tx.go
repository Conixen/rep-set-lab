package database

import (
	"context"
	"fmt"
)

// CompleteAndAwardXP marks a workout as completed and awards XP to the user in a
// single transaction. If either update fails, both are rolled back — the user will
// never have a completed workout with no XP, or XP with no completed workout.
func (s *WorkoutStore) CompleteAndAwardXP(ctx context.Context, workoutID, userID, xpAmount int64, newLevel int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx, `
		UPDATE workouts SET completed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND completed_at IS NULL
	`, workoutID, userID)
	if err != nil {
		return fmt.Errorf("complete workout: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("workout not found or already completed")
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE users SET xp = xp + $1, level = $2, updated_at = NOW() WHERE id = $3
	`, xpAmount, newLevel, userID); err != nil {
		return fmt.Errorf("award xp: %w", err)
	}

	return tx.Commit()
}
