package xp

const xpPerMinute = 10

// levelThresholds[i] is the total XP required to reach level i+1.
// Extend the slice to add more levels — no other code changes needed.
var levelThresholds = []int64{
	0,     // level 1
	500,   // level 2  (~50 min workout)
	1500,  // level 3
	3000,  // level 4
	5500,  // level 5
	9000,  // level 6
	13500, // level 7
	19000, // level 8
	26000, // level 9
	35000, // level 10
}

// Earned returns XP awarded for a workout of the given duration.
func Earned(durationMinutes int) int64 {
	return int64(durationMinutes) * xpPerMinute
}

// LevelForXP returns what level a user with totalXP should be at.
func LevelForXP(totalXP int64) int {
	level := 1
	for i, threshold := range levelThresholds {
		if totalXP >= threshold {
			level = i + 1
		} else {
			break
		}
	}
	return level
}

// NextThreshold returns the XP needed to reach the next level.
// Returns 0 when the user is already at max level.
func NextThreshold(currentLevel int) int64 {
	if currentLevel >= len(levelThresholds) {
		return 0
	}
	return levelThresholds[currentLevel]
}
