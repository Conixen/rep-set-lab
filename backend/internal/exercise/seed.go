package exercise

import (
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/lib/pq"
)

func DefaultExercises() []*database.Exercise {
	return []*database.Exercise{
		// Chest
		{Name: "Barbell Bench Press", Description: "Compound horizontal push", MuscleGroup: "chest", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"bench press", "flat bench", "flat bench press", "chest press"}},
		{Name: "Incline Dumbbell Press", Description: "Upper chest emphasis", MuscleGroup: "chest", Difficulty: "intermediate", Equipment: "dumbbells",
			Aliases: pq.StringArray{"incline press", "incline bench press", "incline bench"}},
		{Name: "Dumbbell Flyes", Description: "Chest isolation with stretch", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "dumbbells",
			Aliases: pq.StringArray{"chest fly", "dumbbell fly", "flies", "pec fly"}},
		{Name: "Push-Up", Description: "Bodyweight chest and tricep movement", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "none",
			Aliases: pq.StringArray{"push ups", "pushup", "pushups", "push-ups"}},
		{Name: "Cable Crossover", Description: "Cable chest fly with constant tension", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"cable fly", "cable flyes", "cable chest fly"}},
		// Back
		{Name: "Pull-Up", Description: "Bodyweight vertical pull", MuscleGroup: "back", Difficulty: "intermediate", Equipment: "pull-up bar",
			Aliases: pq.StringArray{"pull-ups", "pullup", "pullups", "chin-up", "chin up", "chinup"}},
		{Name: "Barbell Row", Description: "Compound horizontal pull", MuscleGroup: "back", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"bent over row", "bent-over row", "barbell bent over row", "barbell rowing"}},
		{Name: "Lat Pulldown", Description: "Machine lat pull", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"lat pull-down", "lat pull down", "pulldown", "cable pulldown"}},
		{Name: "Seated Cable Row", Description: "Horizontal back pull", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"cable row", "seated row", "low cable row"}},
		{Name: "Deadlift", Description: "Full posterior chain compound lift", MuscleGroup: "back", Difficulty: "advanced", Equipment: "barbell",
			Aliases: pq.StringArray{"conventional deadlift", "barbell deadlift", "deadlifts"}},
		{Name: "Face Pull", Description: "Rear delt and rotator cuff", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"cable face pull", "face pulls", "rear delt pull"}},
		// Legs
		{Name: "Barbell Squat", Description: "King of compound leg movements", MuscleGroup: "legs", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"squat", "squats", "back squat", "barbell back squat"}},
		{Name: "Leg Press", Description: "Machine quad and glute push", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "leg press machine",
			Aliases: pq.StringArray{"machine leg press", "leg press machine"}},
		{Name: "Romanian Deadlift", Description: "Hamstring-focused hip hinge", MuscleGroup: "legs", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"rdl", "stiff-legged deadlift", "romanian deadlifts", "dumbbell rdl"}},
		{Name: "Leg Curl", Description: "Hamstring isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine",
			Aliases: pq.StringArray{"hamstring curl", "lying leg curl", "seated leg curl", "hamstring curls"}},
		{Name: "Leg Extension", Description: "Quad isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine",
			Aliases: pq.StringArray{"quad extension", "knee extension", "leg extensions"}},
		{Name: "Lunges", Description: "Unilateral quad and glute movement", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "dumbbells",
			Aliases: pq.StringArray{"lunge", "dumbbell lunge", "dumbbell lunges", "walking lunges", "walking lunge"}},
		{Name: "Calf Raise", Description: "Calf isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine or bodyweight",
			Aliases: pq.StringArray{"standing calf raise", "standing calf raises", "calf raises"}},
		{Name: "Hip Thrust", Description: "Glute isolation compound", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "barbell",
			Aliases: pq.StringArray{"glute bridge", "barbell hip thrust", "barbell glute bridge", "hip thrusts"}},
		// Shoulders
		{Name: "Overhead Press", Description: "Compound shoulder press", MuscleGroup: "shoulders", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"ohp", "shoulder press", "military press", "barbell overhead press", "barbell press"}},
		{Name: "Dumbbell Lateral Raise", Description: "Side deltoid isolation", MuscleGroup: "shoulders", Difficulty: "beginner", Equipment: "dumbbells",
			Aliases: pq.StringArray{"lateral raise", "lateral raises", "side raise", "side lateral raise", "side raises"}},
		{Name: "Rear Delt Fly", Description: "Posterior deltoid isolation", MuscleGroup: "shoulders", Difficulty: "beginner", Equipment: "dumbbells",
			Aliases: pq.StringArray{"reverse fly", "reverse delt fly", "bent-over lateral raise", "rear delt raises"}},
		{Name: "Arnold Press", Description: "Dumbbell shoulder press with rotation", MuscleGroup: "shoulders", Difficulty: "intermediate", Equipment: "dumbbells",
			Aliases: pq.StringArray{"arnold dumbbell press", "arnold shoulder press"}},
		// Arms
		{Name: "Barbell Curl", Description: "Compound bicep curl", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "barbell",
			Aliases: pq.StringArray{"bicep curl", "bicep curls", "barbell bicep curl", "ez bar curl", "biceps curl"}},
		{Name: "Hammer Curl", Description: "Bicep and brachialis curl", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "dumbbells",
			Aliases: pq.StringArray{"hammer curls", "neutral grip curl", "dumbbell hammer curl"}},
		{Name: "Tricep Pushdown", Description: "Tricep cable isolation", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"tricep pulldown", "cable pushdown", "cable tricep pushdown", "triceps pushdown"}},
		{Name: "Skull Crusher", Description: "Lying tricep extension", MuscleGroup: "arms", Difficulty: "intermediate", Equipment: "barbell",
			Aliases: pq.StringArray{"lying tricep extension", "ez bar skull crusher", "skull crushers", "french press"}},
		{Name: "Dip", Description: "Bodyweight tricep and chest movement", MuscleGroup: "arms", Difficulty: "intermediate", Equipment: "dip bars",
			Aliases: pq.StringArray{"dips", "tricep dip", "tricep dips", "parallel bar dip", "bench dip"}},
		// Core
		{Name: "Plank", Description: "Isometric core hold", MuscleGroup: "core", Difficulty: "beginner", Equipment: "none",
			Aliases: pq.StringArray{"forearm plank", "plank hold", "plank position"}},
		{Name: "Cable Crunch", Description: "Weighted core flexion", MuscleGroup: "core", Difficulty: "beginner", Equipment: "cable machine",
			Aliases: pq.StringArray{"kneeling cable crunch", "ab crunch", "cable ab crunch"}},
		{Name: "Hanging Leg Raise", Description: "Lower ab and hip flexor movement", MuscleGroup: "core", Difficulty: "intermediate", Equipment: "pull-up bar",
			Aliases: pq.StringArray{"leg raise", "hanging knee raise", "hanging leg raises"}},
		{Name: "Ab Wheel Rollout", Description: "Anti-extension core movement", MuscleGroup: "core", Difficulty: "advanced", Equipment: "ab wheel",
			Aliases: pq.StringArray{"ab wheel", "rollout", "wheel rollout", "ab roller"}},
		{Name: "Russian Twist", Description: "Rotational core movement", MuscleGroup: "core", Difficulty: "beginner", Equipment: "none or plate",
			Aliases: pq.StringArray{"russian twists", "oblique twist", "seated russian twist"}},
		// Lower Legs
		{Name: "Seated Calf Raise", Description: "Soleus-focused calf isolation on machine", MuscleGroup: "lower legs", Difficulty: "beginner", Equipment: "machine",
			Aliases: pq.StringArray{"seated calf raises", "machine calf raise", "seated calf"}},
		{Name: "Single-Leg Calf Raise", Description: "Unilateral calf isolation for balance and symmetry", MuscleGroup: "lower legs", Difficulty: "beginner", Equipment: "none or dumbbell",
			Aliases: pq.StringArray{"single leg calf raise", "one-leg calf raise", "one leg calf raise", "unilateral calf raise"}},
		// Lower Arms
		{Name: "Barbell Wrist Curl", Description: "Forearm flexor isolation", MuscleGroup: "lower arms", Difficulty: "beginner", Equipment: "barbell",
			Aliases: pq.StringArray{"wrist curl", "wrist curls", "forearm curl", "dumbbell wrist curl"}},
		{Name: "Reverse Barbell Wrist Curl", Description: "Forearm extensor isolation", MuscleGroup: "lower arms", Difficulty: "beginner", Equipment: "barbell",
			Aliases: pq.StringArray{"reverse wrist curl", "wrist extension", "wrist extensions", "reverse wrist curls"}},
		{Name: "Barbell Reverse Curl", Description: "Forearm and brachioradialis builder", MuscleGroup: "lower arms", Difficulty: "beginner", Equipment: "barbell",
			Aliases: pq.StringArray{"reverse curl", "reverse curls", "reverse bicep curl", "overhand curl"}},
	}
}
