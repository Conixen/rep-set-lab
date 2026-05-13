package xp_test

import (
	"testing"

	"github.com/leonj/rep-set-lab/internal/xp"
)

func TestEarned(t *testing.T) {
	tests := []struct {
		duration int
		want     int64
	}{
		{0, 0},
		{1, 10},
		{30, 300},
		{60, 600},
		{90, 900},
	}
	for _, tt := range tests {
		got := xp.Earned(tt.duration)
		if got != tt.want {
			t.Errorf("Earned(%d) = %d, want %d", tt.duration, got, tt.want)
		}
	}
}

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		totalXP int64
		want    int
	}{
		{0, 1},      // below first threshold → level 1
		{499, 1},    // just below level 2 threshold (500)
		{500, 2},    // exactly at level 2 threshold
		{501, 2},
		{1499, 2},   // just below level 3 threshold (1500)
		{1500, 3},
		{34999, 9},
		{35000, 10}, // max level
		{99999, 10}, // beyond max — stays at 10
	}
	for _, tt := range tests {
		got := xp.LevelForXP(tt.totalXP)
		if got != tt.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tt.totalXP, got, tt.want)
		}
	}
}

func TestNextThreshold(t *testing.T) {
	tests := []struct {
		level int
		want  int64
	}{
		{1, 500},    // XP needed to reach level 2
		{2, 1500},
		{9, 35000},
		{10, 0},     // max level — no next threshold
		{99, 0},     // beyond max
	}
	for _, tt := range tests {
		got := xp.NextThreshold(tt.level)
		if got != tt.want {
			t.Errorf("NextThreshold(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}
