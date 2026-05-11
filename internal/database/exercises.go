package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Exercise struct {
	ID           int64      `db:"id"            json:"id"`
	Name         string     `db:"name"          json:"name"`
	Description  string     `db:"description"   json:"description"`
	MuscleGroup  string     `db:"muscle_group"  json:"muscle_group"`
	Difficulty   string     `db:"difficulty"    json:"difficulty"`
	Equipment    string     `db:"equipment"     json:"equipment"`
	ThumbnailURL NullString `db:"thumbnail_url" json:"thumbnail_url"`
	GifURL       NullString `db:"gif_url"       json:"gif_url"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

type ExerciseStore struct {
	db *sqlx.DB
}

func NewExerciseStore(db *sqlx.DB) *ExerciseStore { return &ExerciseStore{db: db} }

func (s *ExerciseStore) List(ctx context.Context, muscleGroup string) ([]*Exercise, error) {
	var exercises []*Exercise
	if muscleGroup != "" {
		err := s.db.SelectContext(ctx, &exercises,
			`SELECT * FROM exercises WHERE muscle_group = $1 ORDER BY name`, muscleGroup)
		return exercises, err
	}
	err := s.db.SelectContext(ctx, &exercises,
		`SELECT * FROM exercises ORDER BY muscle_group, name`)
	return exercises, err
}

func (s *ExerciseStore) Seed(ctx context.Context, exercises []*Exercise) error {
	for _, e := range exercises {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO exercises (name, description, muscle_group, difficulty, equipment)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name) DO NOTHING
		`, e.Name, e.Description, e.MuscleGroup, e.Difficulty, e.Equipment)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateMedia sets thumbnail_url and/or gif_url for an exercise.
// An empty string leaves the existing value unchanged.
func (s *ExerciseStore) UpdateMedia(ctx context.Context, id int64, thumbnailURL, gifURL string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE exercises
		SET
			thumbnail_url = CASE WHEN $1 <> '' THEN $1 ELSE thumbnail_url END,
			gif_url       = CASE WHEN $2 <> '' THEN $2 ELSE gif_url END
		WHERE id = $3
	`, thumbnailURL, gifURL, id)
	return err
}
