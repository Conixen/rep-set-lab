package exercise

import (
	"context"
	"time"

	"github.com/leonj/rep-set-lab/internal/database"
)

// SyncResult summarises the outcome of a media sync run.
type SyncResult struct {
	Total      int `json:"total"`
	Thumbnails int `json:"thumbnails"`
	GIFs       int `json:"gifs"`
	Errors     int `json:"errors"`
}

type exerciseMediaStore interface {
	List(ctx context.Context, muscleGroup string) ([]*database.Exercise, error)
	UpdateMedia(ctx context.Context, id int64, thumbnailURL, gifURL string) error
}

type thumbnailFetcher interface {
	FetchThumbnail(ctx context.Context, name string) (string, error)
}

// GIFFetcher is satisfied by *ExerciseDBClient. Exported so callers can declare
// a nil interface value when no ExerciseDB key is configured, avoiding the
// typed-nil-in-interface pitfall.
type GIFFetcher interface {
	FetchGIF(ctx context.Context, name string) (string, error)
}

type SyncService struct {
	exercises  exerciseMediaStore
	wger       thumbnailFetcher
	exerciseDB GIFFetcher // nil when EXERCISEDB_API_KEY is not set
}

func NewSyncService(exercises exerciseMediaStore, wger thumbnailFetcher, exerciseDB GIFFetcher) *SyncService {
	return &SyncService{exercises: exercises, wger: wger, exerciseDB: exerciseDB}
}

// Sync fetches thumbnail and GIF URLs for every exercise that is missing either.
// A 100ms pause between exercises avoids hammering the external APIs.
func (s *SyncService) Sync(ctx context.Context) (SyncResult, error) {
	all, err := s.exercises.List(ctx, "")
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{Total: len(all)}

	for _, ex := range all {
		if ctx.Err() != nil {
			break
		}

		var thumbnail, gif string
		calledAPI := false

		if !ex.ThumbnailURL.Valid || ex.ThumbnailURL.String == "" {
			calledAPI = true
			t, err := s.wger.FetchThumbnail(ctx, ex.Name)
			if err == nil && t != "" {
				thumbnail = t
				result.Thumbnails++
			} else if err != nil {
				result.Errors++
			}
		}

		if s.exerciseDB != nil && (!ex.GifURL.Valid || ex.GifURL.String == "") {
			calledAPI = true
			g, err := s.exerciseDB.FetchGIF(ctx, ex.Name)
			if err == nil && g != "" {
				gif = g
				result.GIFs++
			} else if err != nil {
				result.Errors++
			}
		}

		if thumbnail != "" || gif != "" {
			if err := s.exercises.UpdateMedia(ctx, ex.ID, thumbnail, gif); err != nil {
				result.Errors++
			}
		}

		if calledAPI {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return result, nil
}
