package exercise

import "github.com/leonj/rep-set-lab/internal/database"

func DefaultExercises() []*database.Exercise {
	return []*database.Exercise{
		// Chest
		{Name: "Barbell Bench Press", Description: "Compound horizontal push", MuscleGroup: "chest", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Incline Dumbbell Press", Description: "Upper chest emphasis", MuscleGroup: "chest", Difficulty: "intermediate", Equipment: "dumbbells"},
		{Name: "Dumbbell Flyes", Description: "Chest isolation with stretch", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "dumbbells"},
		{Name: "Push-Up", Description: "Bodyweight chest and tricep movement", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "none"},
		{Name: "Cable Crossover", Description: "Cable chest fly with constant tension", MuscleGroup: "chest", Difficulty: "beginner", Equipment: "cable machine"},
		// Back
		{Name: "Pull-Up", Description: "Bodyweight vertical pull", MuscleGroup: "back", Difficulty: "intermediate", Equipment: "pull-up bar"},
		{Name: "Barbell Row", Description: "Compound horizontal pull", MuscleGroup: "back", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Lat Pulldown", Description: "Machine lat pull", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine"},
		{Name: "Seated Cable Row", Description: "Horizontal back pull", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine"},
		{Name: "Deadlift", Description: "Full posterior chain compound lift", MuscleGroup: "back", Difficulty: "advanced", Equipment: "barbell"},
		{Name: "Face Pull", Description: "Rear delt and rotator cuff", MuscleGroup: "back", Difficulty: "beginner", Equipment: "cable machine"},
		// Legs
		{Name: "Barbell Squat", Description: "King of compound leg movements", MuscleGroup: "legs", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Leg Press", Description: "Machine quad and glute push", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "leg press machine"},
		{Name: "Romanian Deadlift", Description: "Hamstring-focused hip hinge", MuscleGroup: "legs", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Leg Curl", Description: "Hamstring isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine"},
		{Name: "Leg Extension", Description: "Quad isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine"},
		{Name: "Lunges", Description: "Unilateral quad and glute movement", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "dumbbells"},
		{Name: "Calf Raise", Description: "Calf isolation", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "machine or bodyweight"},
		{Name: "Hip Thrust", Description: "Glute isolation compound", MuscleGroup: "legs", Difficulty: "beginner", Equipment: "barbell"},
		// Shoulders
		{Name: "Overhead Press", Description: "Compound shoulder press", MuscleGroup: "shoulders", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Dumbbell Lateral Raise", Description: "Side deltoid isolation", MuscleGroup: "shoulders", Difficulty: "beginner", Equipment: "dumbbells"},
		{Name: "Rear Delt Fly", Description: "Posterior deltoid isolation", MuscleGroup: "shoulders", Difficulty: "beginner", Equipment: "dumbbells"},
		{Name: "Arnold Press", Description: "Dumbbell shoulder press with rotation", MuscleGroup: "shoulders", Difficulty: "intermediate", Equipment: "dumbbells"},
		// Arms
		{Name: "Barbell Curl", Description: "Compound bicep curl", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "barbell"},
		{Name: "Hammer Curl", Description: "Bicep and brachialis curl", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "dumbbells"},
		{Name: "Tricep Pushdown", Description: "Tricep cable isolation", MuscleGroup: "arms", Difficulty: "beginner", Equipment: "cable machine"},
		{Name: "Skull Crusher", Description: "Lying tricep extension", MuscleGroup: "arms", Difficulty: "intermediate", Equipment: "barbell"},
		{Name: "Dip", Description: "Bodyweight tricep and chest movement", MuscleGroup: "arms", Difficulty: "intermediate", Equipment: "dip bars"},
		// Core
		{Name: "Plank", Description: "Isometric core hold", MuscleGroup: "core", Difficulty: "beginner", Equipment: "none"},
		{Name: "Cable Crunch", Description: "Weighted core flexion", MuscleGroup: "core", Difficulty: "beginner", Equipment: "cable machine"},
		{Name: "Hanging Leg Raise", Description: "Lower ab and hip flexor movement", MuscleGroup: "core", Difficulty: "intermediate", Equipment: "pull-up bar"},
		{Name: "Ab Wheel Rollout", Description: "Anti-extension core movement", MuscleGroup: "core", Difficulty: "advanced", Equipment: "ab wheel"},
		{Name: "Russian Twist", Description: "Rotational core movement", MuscleGroup: "core", Difficulty: "beginner", Equipment: "none or plate"},
	}
}
