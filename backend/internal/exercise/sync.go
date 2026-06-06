package exercise

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/leonj/rep-set-lab/internal/database"
)

// SyncResult summarises the outcome of a media sync run.
type SyncResult struct {
	Total   int      `json:"total"`
	Skipped int      `json:"skipped"`  // already had a GIF
	GIFs    int      `json:"gifs"`     // newly updated
	NoMatch []string `json:"no_match"` // no ExerciseDB result found
	Failed  []string `json:"failed"`   // API or DB error
}

type exerciseMediaStore interface {
	List(ctx context.Context, muscleGroup string) ([]*database.Exercise, error)
	UpdateMedia(ctx context.Context, id int64, thumbnailURL, gifURL string) error
}

// GIFFetcher is satisfied by *ExerciseDBClient. Exported so callers can declare
// a nil interface value when no ExerciseDB key is configured, avoiding the
// typed-nil-in-interface pitfall.
type GIFFetcher interface {
	FetchGIF(ctx context.Context, name string) (string, error)
}

// BulkExerciseFetcher is satisfied by *ExerciseDBClient.
type BulkExerciseFetcher interface {
	FetchByBodyPart(ctx context.Context, bodyPart string) ([]ExerciseDBExercise, error)
}

// exerciseBulkStore handles bulk upsert for the import.
type exerciseBulkStore interface {
	BulkUpsert(ctx context.Context, exercises []*database.Exercise) error
}

// BulkImportResult summarises a bulk import run.
type BulkImportResult struct {
	Imported int `json:"imported"`
	Failed   int `json:"failed"`
}

// ErrBulkImportNotConfigured is returned when no ExerciseDB key is set.
var ErrBulkImportNotConfigured = errors.New("bulk import not configured: EXERCISEDB_API_KEY not set")

// bodyPartToMuscleGroup maps ExerciseDB body parts to our muscle group labels.
var bodyPartToMuscleGroup = map[string]string{
	"back":       "back",
	"chest":      "chest",
	"shoulders":  "shoulders",
	"upper arms": "arms",
	"upper legs": "legs",
	"waist":      "core",
	"lower legs": "lower legs",
	"lower arms": "lower arms",
}

type SyncService struct {
	exercises   exerciseMediaStore
	exerciseDB  GIFFetcher // nil when EXERCISEDB_API_KEY is not set
	bulkStore   exerciseBulkStore
	bulkFetcher BulkExerciseFetcher
}

func NewSyncService(exercises exerciseMediaStore, exerciseDB GIFFetcher) *SyncService {
	return &SyncService{exercises: exercises, exerciseDB: exerciseDB}
}

// WithBulkImport wires in the extra dependencies needed for BulkImport.
func (s *SyncService) WithBulkImport(store exerciseBulkStore, fetcher BulkExerciseFetcher) *SyncService {
	s.bulkStore = store
	s.bulkFetcher = fetcher
	return s
}

// Sync fetches GIF URLs for every exercise that is missing one.
// A 100ms pause between exercises avoids hammering the ExerciseDB API.
//
// Note: wger.de thumbnail fetching was removed because the
// /api/v2/exercise/search/ endpoint no longer exists. The thumbnail_url
// DB column is retained as a future-use fallback (see LibraryView.vue).
func (s *SyncService) Sync(ctx context.Context) (SyncResult, error) {
	all, err := s.exercises.List(ctx, "")
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{
		Total:   len(all),
		NoMatch: []string{},
		Failed:  []string{},
	}

	for _, ex := range all {
		if ctx.Err() != nil {
			break
		}

		if s.exerciseDB == nil || (ex.GifURL.Valid && ex.GifURL.String != "") {
			result.Skipped++
			continue
		}

		g, err := s.exerciseDB.FetchGIF(ctx, ex.Name)
		if err != nil {
			result.Failed = append(result.Failed, ex.Name)
		} else if g == "" {
			result.NoMatch = append(result.NoMatch, ex.Name)
		} else {
			if err := s.exercises.UpdateMedia(ctx, ex.ID, "", g); err != nil {
				result.Failed = append(result.Failed, ex.Name)
			} else {
				result.GIFs++
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return result, nil
}

// BulkImport fetches all exercises from ExerciseDB by body part and upserts them.
// One-time admin operation — safe to re-run, existing aliases are preserved.
func (s *SyncService) BulkImport(ctx context.Context) (BulkImportResult, error) {
	if s.bulkFetcher == nil || s.bulkStore == nil {
		return BulkImportResult{}, ErrBulkImportNotConfigured
	}

	var result BulkImportResult

	for bodyPart, muscleGroup := range bodyPartToMuscleGroup {
		if ctx.Err() != nil {
			break
		}

		entries, err := s.bulkFetcher.FetchByBodyPart(ctx, bodyPart)
		if err != nil {
			slog.Default().Error("bulk import: fetch failed", "body_part", bodyPart, "error", err)
			result.Failed++
			continue
		}

		batch := make([]*database.Exercise, 0, len(entries))
		for _, e := range entries {
			ex := &database.Exercise{
				Name:        titleCase(e.Name),
				Description: e.Description,
				MuscleGroup: muscleGroup,
				Difficulty:  e.Difficulty,
				Equipment:   e.Equipment,
				Aliases:     pq.StringArray{},
			}
			ex.GifURL.String = "/api/v1/exercises/image/" + e.ID
			ex.GifURL.Valid = true
			batch = append(batch, ex)
		}

		if err := s.bulkStore.BulkUpsert(ctx, batch); err != nil {
			slog.Default().Error("bulk import: upsert failed", "body_part", bodyPart, "error", err)
			result.Failed += len(batch)
			continue
		}
		result.Imported += len(batch)
	}

	return result, nil
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
