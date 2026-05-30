package exercise

import (
	"context"
	"time"

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

type SyncService struct {
	exercises  exerciseMediaStore
	exerciseDB GIFFetcher // nil when EXERCISEDB_API_KEY is not set
}

func NewSyncService(exercises exerciseMediaStore, exerciseDB GIFFetcher) *SyncService {
	return &SyncService{exercises: exercises, exerciseDB: exerciseDB}
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
