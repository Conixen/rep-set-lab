package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type Workout struct {
	ID              int64           `db:"id"               json:"id"`
	UserID          int64           `db:"user_id"          json:"user_id"`
	Prompt          string          `db:"prompt"           json:"prompt"`
	MuscleGroup     string          `db:"muscle_group"     json:"muscle_group"`
	DurationMinutes int             `db:"duration_minutes" json:"duration_minutes"`
	Injuries        sql.NullString  `db:"injuries"         json:"injuries"`
	Goals           sql.NullString  `db:"goals"            json:"goals"`
	AIProvider      string          `db:"ai_provider"      json:"ai_provider"`
	AIResponse      json.RawMessage `db:"ai_response"      json:"ai_response"`
	XPEarned        int             `db:"xp_earned"        json:"xp_earned"`
	CompletedAt     sql.NullTime    `db:"completed_at"     json:"completed_at"`
	CreatedAt       time.Time       `db:"created_at"       json:"created_at"`
}

type WorkoutStore struct {
	db *sqlx.DB
}

func NewWorkoutStore(db *sqlx.DB) *WorkoutStore { return &WorkoutStore{db: db} }

func (s *WorkoutStore) Create(ctx context.Context, w *Workout) (*Workout, error) {
	var result Workout
	err := s.db.QueryRowxContext(ctx, `
		INSERT INTO workouts
			(user_id, prompt, muscle_group, duration_minutes, injuries, goals, ai_provider, ai_response, xp_earned)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING *
	`, w.UserID, w.Prompt, w.MuscleGroup, w.DurationMinutes,
		w.Injuries, w.Goals, w.AIProvider, w.AIResponse, w.XPEarned,
	).StructScan(&result)
	return &result, err
}

func (s *WorkoutStore) GetByID(ctx context.Context, id, userID int64) (*Workout, error) {
	var w Workout
	err := s.db.GetContext(ctx, &w,
		`SELECT * FROM workouts WHERE id = $1 AND user_id = $2`, id, userID)
	return &w, err
}

func (s *WorkoutStore) ListByUser(ctx context.Context, userID int64) ([]*Workout, error) {
	var workouts []*Workout
	err := s.db.SelectContext(ctx, &workouts,
		`SELECT * FROM workouts WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return workouts, err
}

func (s *WorkoutStore) Complete(ctx context.Context, id, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE workouts SET completed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND completed_at IS NULL
	`, id, userID)
	return err
}

func (s *WorkoutStore) ListAll(ctx context.Context) ([]*Workout, error) {
	var workouts []*Workout
	err := s.db.SelectContext(ctx, &workouts, `SELECT * FROM workouts ORDER BY created_at DESC`)
	return workouts, err
}

func (s *WorkoutStore) GetByIDAdmin(ctx context.Context, id int64) (*Workout, error) {
	var w Workout
	err := s.db.GetContext(ctx, &w, `SELECT * FROM workouts WHERE id = $1`, id)
	return &w, err
}

func (s *WorkoutStore) AdminSetCompleted(ctx context.Context, id int64, completed bool) error {
	var result sql.Result
	var err error
	if completed {
		result, err = s.db.ExecContext(ctx, `UPDATE workouts SET completed_at = NOW() WHERE id = $1`, id)
	} else {
		result, err = s.db.ExecContext(ctx, `UPDATE workouts SET completed_at = NULL WHERE id = $1`, id)
	}
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
