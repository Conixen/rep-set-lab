package mock

import (
	"context"

	"github.com/leonj/rep-set-lab/internal/database"
)

// WorkoutStorage is a handwritten mock for workout.Storage.
type WorkoutStorage struct {
	CreateFunc      func(ctx context.Context, w *database.Workout) (*database.Workout, error)
	GetByIDFunc     func(ctx context.Context, id, userID int64) (*database.Workout, error)
	ListByUserFunc  func(ctx context.Context, userID int64) ([]*database.Workout, error)
	CompleteFunc    func(ctx context.Context, id, userID int64) error
}

func (m *WorkoutStorage) Create(ctx context.Context, w *database.Workout) (*database.Workout, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, w)
	}
	w.ID = 1
	return w, nil
}

func (m *WorkoutStorage) GetByID(ctx context.Context, id, userID int64) (*database.Workout, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id, userID)
	}
	return &database.Workout{ID: id, UserID: userID, XPEarned: 600}, nil
}

func (m *WorkoutStorage) ListByUser(ctx context.Context, userID int64) ([]*database.Workout, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *WorkoutStorage) Complete(ctx context.Context, id, userID int64) error {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, id, userID)
	}
	return nil
}
