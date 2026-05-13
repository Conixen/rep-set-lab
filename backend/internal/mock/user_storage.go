package mock

import (
	"context"

	"github.com/leonj/rep-set-lab/internal/database"
)

// UserStorage is a handwritten mock for workout.UserStorage.
type UserStorage struct {
	GetByIDFunc func(ctx context.Context, id int64) (*database.User, error)
}

func (m *UserStorage) GetByID(ctx context.Context, id int64) (*database.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &database.User{ID: id, XP: 0, Level: 1}, nil
}
