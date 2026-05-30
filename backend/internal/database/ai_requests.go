package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type AIRequest struct {
	ID           int64          `db:"id"            json:"id"`
	UserID       int64          `db:"user_id"       json:"user_id"`
	WorkoutID    *int64         `db:"workout_id"    json:"workout_id"`
	Provider     string         `db:"provider"      json:"provider"`
	InputTokens  int            `db:"input_tokens"  json:"input_tokens"`
	OutputTokens int            `db:"output_tokens" json:"output_tokens"`
	CostUSD      float64        `db:"cost_usd"      json:"cost_usd"`
	LatencyMs    int64          `db:"latency_ms"    json:"latency_ms"`
	ValidJSON    bool           `db:"valid_json"    json:"valid_json"`
	ErrorMessage sql.NullString `db:"error_message" json:"error_message"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"`
}

type AIUsageSummary struct {
	TotalRequests int     `db:"total_requests" json:"total_requests"`
	TotalCostUSD  float64 `db:"total_cost_usd" json:"total_cost_usd"`
	AvgLatencyMs  int64   `db:"avg_latency_ms" json:"avg_latency_ms"`
	ThisMonth     int     `db:"this_month"     json:"this_month"`
}

type AIRequestRow struct {
	AIRequest
	Username string `db:"username" json:"username"`
}

type AIProviderStat struct {
	Provider          string  `db:"provider"           json:"provider"`
	TotalCalls        int     `db:"total_calls"        json:"total_calls"`
	ValidCalls        int     `db:"valid_calls"        json:"valid_calls"`
	AvgLatencyMs      int64   `db:"avg_latency_ms"     json:"avg_latency_ms"`
	AvgCostUSD        float64 `db:"avg_cost_usd"       json:"avg_cost_usd"`
	TotalCostUSD      float64 `db:"total_cost_usd"     json:"total_cost_usd"`
	AvgInputTokens    int64   `db:"avg_input_tokens"   json:"avg_input_tokens"`
	AvgOutputTokens   int64   `db:"avg_output_tokens"  json:"avg_output_tokens"`
	TotalInputTokens  int64   `db:"total_input_tokens" json:"total_input_tokens"`
	TotalOutputTokens int64   `db:"total_output_tokens" json:"total_output_tokens"`
}

type AIRequestStore struct {
	db *sqlx.DB
}

func NewAIRequestStore(db *sqlx.DB) *AIRequestStore { return &AIRequestStore{db: db} }

func (s *AIRequestStore) Log(ctx context.Context, r *AIRequest) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_requests
			(user_id, workout_id, provider, input_tokens, output_tokens, cost_usd, latency_ms, valid_json, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, r.UserID, r.WorkoutID, r.Provider, r.InputTokens, r.OutputTokens, r.CostUSD, r.LatencyMs, r.ValidJSON, r.ErrorMessage)
	return err
}

func (s *AIRequestStore) ListAdmin(ctx context.Context, limit, offset int) ([]*AIRequestRow, error) {
	var rows []*AIRequestRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT r.*, u.username
		FROM ai_requests r
		JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return rows, err
}

func (s *AIRequestStore) CountAll(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM ai_requests`)
	return count, err
}

func (s *AIRequestStore) ProviderStats(ctx context.Context) ([]*AIProviderStat, error) {
	var stats []*AIProviderStat
	err := s.db.SelectContext(ctx, &stats, `
		SELECT
			provider,
			COUNT(*)                                    AS total_calls,
			COUNT(*) FILTER (WHERE valid_json)          AS valid_calls,
			COALESCE(AVG(latency_ms), 0)::BIGINT        AS avg_latency_ms,
			COALESCE(AVG(cost_usd), 0)                  AS avg_cost_usd,
			COALESCE(SUM(cost_usd), 0)                  AS total_cost_usd,
			COALESCE(AVG(input_tokens), 0)::BIGINT      AS avg_input_tokens,
			COALESCE(AVG(output_tokens), 0)::BIGINT     AS avg_output_tokens,
			COALESCE(SUM(input_tokens), 0)::BIGINT      AS total_input_tokens,
			COALESCE(SUM(output_tokens), 0)::BIGINT     AS total_output_tokens
		FROM ai_requests
		GROUP BY provider
		ORDER BY total_calls DESC
	`)
	return stats, err
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
