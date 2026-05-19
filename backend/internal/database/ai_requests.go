package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type AIRequest struct {
	ID           int64     `db:"id"            json:"id"`
	UserID       int64     `db:"user_id"       json:"user_id"`
	WorkoutID    *int64    `db:"workout_id"    json:"workout_id"`
	Provider     string    `db:"provider"      json:"provider"`
	InputTokens  int       `db:"input_tokens"  json:"input_tokens"`
	OutputTokens int       `db:"output_tokens" json:"output_tokens"`
	CostUSD      float64   `db:"cost_usd"      json:"cost_usd"`
	LatencyMs    int64     `db:"latency_ms"    json:"latency_ms"`
	ValidJSON    bool      `db:"valid_json"    json:"valid_json"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type AIUsageSummary struct {
	TotalRequests int     `db:"total_requests" json:"total_requests"`
	TotalCostUSD  float64 `db:"total_cost_usd" json:"total_cost_usd"`
	AvgLatencyMs  int64   `db:"avg_latency_ms" json:"avg_latency_ms"`
	ThisMonth     int     `db:"this_month"     json:"this_month"`
}

type AIRequestStore struct {
	db *sqlx.DB
}

func NewAIRequestStore(db *sqlx.DB) *AIRequestStore { return &AIRequestStore{db: db} }

func (s *AIRequestStore) Log(ctx context.Context, r *AIRequest) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_requests
			(user_id, workout_id, provider, input_tokens, output_tokens, cost_usd, latency_ms, valid_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, r.UserID, r.WorkoutID, r.Provider, r.InputTokens, r.OutputTokens, r.CostUSD, r.LatencyMs, r.ValidJSON)
	return err
}

func (s *AIRequestStore) Summary(ctx context.Context, userID int64) (*AIUsageSummary, error) {
	var summary AIUsageSummary
	err := s.db.GetContext(ctx, &summary, `
		SELECT
			COUNT(*)                                                          AS total_requests,
			COALESCE(SUM(cost_usd), 0)                                       AS total_cost_usd,
			COALESCE(AVG(latency_ms), 0)::BIGINT                             AS avg_latency_ms,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('month', NOW())) AS this_month
		FROM ai_requests
		WHERE user_id = $1
	`, userID)
	return &summary, err
}
