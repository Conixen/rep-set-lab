package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/leonj/rep-set-lab/internal/ai"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/config"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/exercise"
	"github.com/leonj/rep-set-lab/internal/user"
	"github.com/leonj/rep-set-lab/internal/workout"
	"github.com/leonj/rep-set-lab/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db, "migrations"); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	// Stores
	userStore     := database.NewUserStore(db)
	workoutStore  := database.NewWorkoutStore(db)
	exerciseStore := database.NewExerciseStore(db)

	// Seed exercise library (ON CONFLICT DO NOTHING — safe to run every boot)
	if err := exerciseStore.Seed(context.Background(), exercise.DefaultExercises()); err != nil {
		logger.Error("seed exercises", "error", err)
	}

	// AI providers — only register providers with keys configured
	providers := make(map[string]ai.Provider)
	if cfg.AnthropicKey != "" {
		providers["claude"] = ai.NewClaude(cfg.AnthropicKey)
	}
	if cfg.OpenAIKey != "" {
		providers["openai"] = ai.NewOpenAI(cfg.OpenAIKey)
	}
	if cfg.GeminiKey != "" {
		providers["gemini"] = ai.NewGemini(cfg.GeminiKey)
	}

	hub     := ws.NewHub(logger)
	svc     := workout.NewService(workoutStore, userStore, providers, hub)

	authHandler     := auth.NewHandler(userStore, cfg.JWTSecret)
	workoutHandler  := workout.NewHandler(svc, workoutStore)
	userHandler     := user.NewHandler(userStore, workoutStore)
	exerciseHandler := exercise.NewHandler(exerciseStore)
	compareHandler  := ai.NewCompareHandler(providers)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	v1 := r.Group("/api/v1")

	// Public routes
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login",    authHandler.Login)

	// Protected routes
	protected := v1.Group("", auth.Middleware(cfg.JWTSecret))
	{
		protected.GET("/users/me/stats", userHandler.Stats)

		protected.POST("/workouts/generate",       workoutHandler.Generate)
		protected.GET("/workouts",                 workoutHandler.List)
		protected.GET("/workouts/:id",             workoutHandler.Get)
		protected.POST("/workouts/:id/complete",   workoutHandler.Complete)

		protected.GET("/exercises", exerciseHandler.List)

		protected.POST("/ai/compare", compareHandler.Compare)
	}

	// WebSocket — JWT via ?token= query param
	r.GET("/ws", auth.WSMiddleware(cfg.JWTSecret), hub.Handler)

	logger.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
