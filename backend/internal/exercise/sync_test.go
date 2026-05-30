package exercise_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/exercise"
)

// --- stubs ---

type stubExerciseStore struct {
	exercises    []*database.Exercise
	mediaUpdates map[int64][2]string
	updateErr    error
}

func (s *stubExerciseStore) List(_ context.Context, _ string) ([]*database.Exercise, error) {
	return s.exercises, nil
}

func (s *stubExerciseStore) UpdateMedia(_ context.Context, id int64, thumbnail, gif string) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	if s.mediaUpdates == nil {
		s.mediaUpdates = make(map[int64][2]string)
	}
	s.mediaUpdates[id] = [2]string{thumbnail, gif}
	return nil
}

type stubGIFFetcher struct {
	url string
	err error
}

func (s *stubGIFFetcher) FetchGIF(_ context.Context, _ string) (string, error) {
	return s.url, s.err
}

// --- helpers ---

func nullStr() database.NullString { return database.NullString{} }
func validStr(v string) database.NullString {
	return database.NullString{NullString: sql.NullString{String: v, Valid: true}}
}

// --- tests ---

func TestSync_GIFAlreadyPopulated_NoAPICalls(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 1, Name: "Squat", GifURL: validStr("/api/v1/exercises/image/0001")},
	}}
	svc := exercise.NewSyncService(store, &stubGIFFetcher{})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if result.GIFs != 0 || len(result.Failed) != 0 {
		t.Errorf("want no increments, got GIFs=%d Failed=%v", result.GIFs, result.Failed)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if len(store.mediaUpdates) != 0 {
		t.Errorf("expected no UpdateMedia calls, got %v", store.mediaUpdates)
	}
}

func TestSync_MissingGIF(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 3, Name: "Deadlift", GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store, &stubGIFFetcher{url: "/api/v1/exercises/image/0002"})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GIFs != 1 || len(result.Failed) != 0 {
		t.Errorf("want GIFs=1 Failed=[], got %+v", result)
	}
	if store.mediaUpdates[3][1] != "/api/v1/exercises/image/0002" {
		t.Errorf("gif not saved, updates=%v", store.mediaUpdates)
	}
}

func TestSync_NoExerciseDBKey_SkipsAll(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 4, Name: "Curl", GifURL: nullStr()},
	}}
	// nil exerciseDB simulates EXERCISEDB_API_KEY not configured
	svc := exercise.NewSyncService(store, nil)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GIFs != 0 || len(result.Failed) != 0 {
		t.Errorf("expected no updates with nil exerciseDB, got %+v", result)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
}

func TestSync_GIFFetchError_IncrementsFailed(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 6, Name: "Row", GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store, &stubGIFFetcher{err: errors.New("rapidapi unavailable")})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Failed) != 1 || result.Failed[0] != "Row" {
		t.Errorf("Failed = %v, want [Row]", result.Failed)
	}
}

func TestSync_NoMatch_PopulatesNoMatch(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 5, Name: "Obscure Exercise", GifURL: nullStr()},
	}}
	// empty url = no ExerciseDB match
	svc := exercise.NewSyncService(store, &stubGIFFetcher{url: ""})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.NoMatch) != 1 || result.NoMatch[0] != "Obscure Exercise" {
		t.Errorf("NoMatch = %v, want [Obscure Exercise]", result.NoMatch)
	}
	if result.GIFs != 0 {
		t.Errorf("GIFs = %d, want 0", result.GIFs)
	}
}

func TestSync_UpdateMediaError_IncrementsFailed(t *testing.T) {
	store := &stubExerciseStore{
		exercises: []*database.Exercise{
			{ID: 7, Name: "Lunge", GifURL: nullStr()},
		},
		updateErr: errors.New("db write failed"),
	}
	svc := exercise.NewSyncService(store, &stubGIFFetcher{url: "/api/v1/exercises/image/0003"})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Failed) != 1 || result.Failed[0] != "Lunge" {
		t.Errorf("Failed = %v, want [Lunge]", result.Failed)
	}
}
