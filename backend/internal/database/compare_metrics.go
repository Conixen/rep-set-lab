package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type CompareMetric struct {
	ID                  int64          `db:"id"`
	SessionID           string         `db:"session_id"`
	UserID              int64          `db:"user_id"`
	Provider            string         `db:"provider"`
	MuscleGroup         string         `db:"muscle_group"`
	DurationMinutes     int            `db:"duration_minutes"`
	Environment         string         `db:"environment"`
	HasInjuries         bool           `db:"has_injuries"`
	LibraryMatchRate    float64        `db:"library_match_rate"`
	LibraryMatchCount   int            `db:"library_match_count"`
	LibraryTotalCount   int            `db:"library_total_count"`
	CharCount           int            `db:"char_count"`
	EmojiCount          int            `db:"emoji_count"`
	EquipmentViolations int            `db:"equipment_violations"`
	CompletenessScore   int            `db:"completeness_score"`
	WarmUpCount         int            `db:"warm_up_count"`
	MainCount           int            `db:"main_count"`
	CoolDownCount       int            `db:"cool_down_count"`
	TipsCount           int            `db:"tips_count"`
	AvgNoteLength       float64        `db:"avg_note_length"`
	NotesPresentRate    float64        `db:"notes_present_rate"`
	EstimatedMinutes    float64        `db:"estimated_minutes"`
	GroqInjuryGrade     sql.NullString `db:"groq_injury_grade"`
	GroqEquipmentGrade  sql.NullString `db:"groq_equipment_grade"`
	GroqGoalGrade       sql.NullString `db:"groq_goal_grade"`
	GroqFeedback        sql.NullString `db:"groq_feedback"`
	CreatedAt           time.Time      `db:"created_at"`
}

type ProviderCompareAvg struct {
	Provider              string  `db:"provider"               json:"provider"`
	TotalSessions         int     `db:"total_sessions"         json:"total_sessions"`
	AvgLibraryMatchRate   float64 `db:"avg_library_match_rate" json:"avg_library_match_rate"`
	AvgCharCount          float64 `db:"avg_char_count"         json:"avg_char_count"`
	AvgEmojiCount         float64 `db:"avg_emoji_count"        json:"avg_emoji_count"`
	AvgEquipmentViolations float64 `db:"avg_equipment_violations" json:"avg_equipment_violations"`
	AvgCompletenessScore  float64 `db:"avg_completeness_score" json:"avg_completeness_score"`
	AvgWarmUpCount        float64 `db:"avg_warm_up_count"      json:"avg_warm_up_count"`
	AvgMainCount          float64 `db:"avg_main_count"         json:"avg_main_count"`
	AvgCoolDownCount      float64 `db:"avg_cool_down_count"    json:"avg_cool_down_count"`
	AvgTipsCount          float64 `db:"avg_tips_count"         json:"avg_tips_count"`
	AvgNotesPresentRate   float64 `db:"avg_notes_present_rate" json:"avg_notes_present_rate"`
	AvgEstimatedMinutes   float64 `db:"avg_estimated_minutes"  json:"avg_estimated_minutes"`
	AvgGroqInjuryScore    float64 `db:"avg_groq_injury_score"  json:"avg_groq_injury_score"`
	AvgGroqEquipmentScore float64 `db:"avg_groq_equipment_score" json:"avg_groq_equipment_score"`
	AvgGroqGoalScore      float64 `db:"avg_groq_goal_score"    json:"avg_groq_goal_score"`
}

type CompareMetricsStore struct {
	db *sqlx.DB
}

func NewCompareMetricsStore(db *sqlx.DB) *CompareMetricsStore {
	return &CompareMetricsStore{db: db}
}

func (s *CompareMetricsStore) Log(ctx context.Context, m *CompareMetric) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO compare_metrics (
			session_id, user_id, provider, muscle_group, duration_minutes, environment, has_injuries,
			library_match_rate, library_match_count, library_total_count,
			char_count, emoji_count, equipment_violations, completeness_score,
			warm_up_count, main_count, cool_down_count, tips_count,
			avg_note_length, notes_present_rate, estimated_minutes,
			groq_injury_grade, groq_equipment_grade, groq_goal_grade, groq_feedback
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21,
			$22, $23, $24, $25
		)`,
		m.SessionID, m.UserID, m.Provider, m.MuscleGroup, m.DurationMinutes, m.Environment, m.HasInjuries,
		m.LibraryMatchRate, m.LibraryMatchCount, m.LibraryTotalCount,
		m.CharCount, m.EmojiCount, m.EquipmentViolations, m.CompletenessScore,
		m.WarmUpCount, m.MainCount, m.CoolDownCount, m.TipsCount,
		m.AvgNoteLength, m.NotesPresentRate, m.EstimatedMinutes,
		m.GroqInjuryGrade, m.GroqEquipmentGrade, m.GroqGoalGrade, m.GroqFeedback,
	)
	return err
}

func (s *CompareMetricsStore) LatestSession(ctx context.Context) ([]*CompareMetric, error) {
	var rows []*CompareMetric
	err := s.db.SelectContext(ctx, &rows, `
		SELECT * FROM compare_metrics
		WHERE session_id = (
			SELECT session_id FROM compare_metrics ORDER BY created_at DESC LIMIT 1
		)
		ORDER BY provider
	`)
	return rows, err
}

func (s *CompareMetricsStore) SessionCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(DISTINCT session_id) FROM compare_metrics`)
	return count, err
}

func (s *CompareMetricsStore) ProviderAverages(ctx context.Context) ([]*ProviderCompareAvg, error) {
	var avgs []*ProviderCompareAvg
	err := s.db.SelectContext(ctx, &avgs, `
		SELECT
			provider,
			COUNT(DISTINCT session_id)                   AS total_sessions,
			COALESCE(AVG(library_match_rate),    0)      AS avg_library_match_rate,
			COALESCE(AVG(char_count),            0)      AS avg_char_count,
			COALESCE(AVG(emoji_count),           0)      AS avg_emoji_count,
			COALESCE(AVG(equipment_violations),  0)      AS avg_equipment_violations,
			COALESCE(AVG(completeness_score),    0)      AS avg_completeness_score,
			COALESCE(AVG(warm_up_count),         0)      AS avg_warm_up_count,
			COALESCE(AVG(main_count),            0)      AS avg_main_count,
			COALESCE(AVG(cool_down_count),       0)      AS avg_cool_down_count,
			COALESCE(AVG(tips_count),            0)      AS avg_tips_count,
			COALESCE(AVG(notes_present_rate),    0)      AS avg_notes_present_rate,
			COALESCE(AVG(estimated_minutes),     0)      AS avg_estimated_minutes,
			COALESCE(AVG(CASE groq_injury_grade
				WHEN 'A' THEN 5 WHEN 'B' THEN 4 WHEN 'C' THEN 3 WHEN 'D' THEN 2 WHEN 'F' THEN 1
				ELSE NULL END), 0) AS avg_groq_injury_score,
			COALESCE(AVG(CASE groq_equipment_grade
				WHEN 'A' THEN 5 WHEN 'B' THEN 4 WHEN 'C' THEN 3 WHEN 'D' THEN 2 WHEN 'F' THEN 1
				ELSE NULL END), 0) AS avg_groq_equipment_score,
			COALESCE(AVG(CASE groq_goal_grade
				WHEN 'A' THEN 5 WHEN 'B' THEN 4 WHEN 'C' THEN 3 WHEN 'D' THEN 2 WHEN 'F' THEN 1
				ELSE NULL END), 0) AS avg_groq_goal_score
		FROM compare_metrics
		GROUP BY provider
		ORDER BY total_sessions DESC
	`)
	return avgs, err
}
