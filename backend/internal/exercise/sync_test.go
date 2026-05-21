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

type stubThumbFetcher struct {
	url string
	err error
}

func (s *stubThumbFetcher) FetchThumbnail(_ context.Context, _ string) (string, error) {
	return s.url, s.err
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

func TestSync_AllPopulated_NoAPICalls(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 1, Name: "Squat", ThumbnailURL: validStr("https://t.co/sq.png"), GifURL: validStr("https://g.co/sq.gif")},
	}}
	svc := exercise.NewSyncService(store, &stubThumbFetcher{}, &stubGIFFetcher{})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if result.Thumbnails != 0 || result.GIFs != 0 || result.Errors != 0 {
		t.Errorf("want no increments, got Thumbnails=%d GIFs=%d Errors=%d", result.Thumbnails, result.GIFs, result.Errors)
	}
	if len(store.mediaUpdates) != 0 {
		t.Errorf("expected no UpdateMedia calls, got %v", store.mediaUpdates)
	}
}

func TestSync_MissingThumbnail(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 2, Name: "Bench Press", ThumbnailURL: nullStr(), GifURL: validStr("https://g.co/bp.gif")},
	}}
	svc := exercise.NewSyncService(store, &stubThumbFetcher{url: "https://t.co/bp.png"}, nil)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Thumbnails != 1 || result.GIFs != 0 || result.Errors != 0 {
		t.Errorf("want Thumbnails=1, got %+v", result)
	}
	if store.mediaUpdates[2][0] != "https://t.co/bp.png" {
		t.Errorf("thumbnail not saved, updates=%v", store.mediaUpdates)
	}
}

func TestSync_MissingGIF(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 3, Name: "Deadlift", ThumbnailURL: validStr("https://t.co/dl.png"), GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store, &stubThumbFetcher{}, &stubGIFFetcher{url: "https://g.co/dl.gif"})

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GIFs != 1 || result.Thumbnails != 0 || result.Errors != 0 {
		t.Errorf("want GIFs=1, got %+v", result)
	}
	if store.mediaUpdates[3][1] != "https://g.co/dl.gif" {
		t.Errorf("gif not saved, updates=%v", store.mediaUpdates)
	}
}

func TestSync_BothMissing(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 4, Name: "Curl", ThumbnailURL: nullStr(), GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store,
		&stubThumbFetcher{url: "https://t.co/curl.png"},
		&stubGIFFetcher{url: "https://g.co/curl.gif"},
	)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Thumbnails != 1 || result.GIFs != 1 || result.Errors != 0 {
		t.Errorf("want Thumbnails=1 GIFs=1 Errors=0, got %+v", result)
	}
}

func TestSync_ThumbnailFetchError_IncrementsErrors(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 5, Name: "Pullup", ThumbnailURL: nullStr(), GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store,
		&stubThumbFetcher{err: errors.New("wger unavailable")},
		&stubGIFFetcher{url: "https://g.co/pullup.gif"},
	)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Errors != 1 {
		t.Errorf("Errors = %d, want 1", result.Errors)
	}
	if result.GIFs != 1 {
		t.Errorf("GIFs = %d, want 1 (GIF should still succeed despite thumb error)", result.GIFs)
	}
}

func TestSync_GIFFetchError_IncrementsErrors(t *testing.T) {
	store := &stubExerciseStore{exercises: []*database.Exercise{
		{ID: 6, Name: "Row", ThumbnailURL: validStr("https://t.co/row.png"), GifURL: nullStr()},
	}}
	svc := exercise.NewSyncService(store,
		&stubThumbFetcher{},
		&stubGIFFetcher{err: errors.New("rapidapi unavailable")},
	)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Errors != 1 {
		t.Errorf("Errors = %d, want 1", result.Errors)
	}
}

func TestSync_UpdateMediaError_IncrementsErrors(t *testing.T) {
	store := &stubExerciseStore{
		exercises: []*database.Exercise{
			{ID: 7, Name: "Lunge", ThumbnailURL: nullStr(), GifURL: nullStr()},
		},
		updateErr: errors.New("db write failed"),
	}
	svc := exercise.NewSyncService(store,
		&stubThumbFetcher{url: "https://t.co/lunge.png"},
		&stubGIFFetcher{url: "https://g.co/lunge.gif"},
	)

	result, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Errors != 1 {
		t.Errorf("Errors = %d, want 1", result.Errors)
	}
}
