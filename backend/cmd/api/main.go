package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/leonj/rep-set-lab/internal/admin"
	"github.com/leonj/rep-set-lab/internal/ai"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/config"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/exercise"
	"github.com/leonj/rep-set-lab/internal/user"
	"github.com/leonj/rep-set-lab/internal/workout"
	"github.com/leonj/rep-set-lab/internal/ws"
)

const shutdownTimeout = 30 * time.Second

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

	userStore        := database.NewUserStore(db)
	workoutStore     := database.NewWorkoutStore(db)
	exerciseStore    := database.NewExerciseStore(db)
	aiRequestStore   := database.NewAIRequestStore(db)

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
		g, err := ai.NewGemini(cfg.GeminiKey)
		if err != nil {
			logger.Error("init gemini provider", "error", err)
			os.Exit(1)
		}
		providers["gemini"] = g
	}

	hub             := ws.NewHub(logger)
	svc             := workout.NewService(workoutStore, userStore, providers, hub, aiRequestStore)
	authHandler     := auth.NewHandler(userStore, cfg.JWTSecret)
	workoutHandler  := workout.NewHandler(svc, workoutStore)
	userHandler     := user.NewHandler(userStore, workoutStore)
	exerciseHandler := exercise.NewHandler(exerciseStore)
	compareHandler  := ai.NewCompareHandler(providers, aiRequestStore)

	wgerClient := exercise.NewWgerClient()
	var exerciseDBClient *exercise.ExerciseDBClient
	if cfg.ExerciseDBKey != "" {
		exerciseDBClient = exercise.NewExerciseDBClient(cfg.ExerciseDBKey)
	}
	syncSvc := exercise.NewSyncService(exerciseStore, wgerClient, exerciseDBClient)

	adminHandler := admin.NewHandler(userStore, workoutStore, syncSvc, aiRequestStore)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login",    authHandler.Login)

	protected := v1.Group("", auth.Middleware(cfg.JWTSecret))
	{
		protected.GET("/users/me/stats",           userHandler.Stats)
		protected.POST("/workouts/generate",       workoutHandler.Generate)
		protected.GET("/workouts",                 workoutHandler.List)
		protected.GET("/workouts/:id",             workoutHandler.Get)
		protected.POST("/workouts/:id/complete",   workoutHandler.Complete)
		protected.GET("/exercises",                exerciseHandler.List)
		protected.POST("/ai/compare",              compareHandler.Compare)
	}

	adminGroup := v1.Group("/admin", auth.Middleware(cfg.JWTSecret), auth.AdminMiddleware(userStore))
	{
		adminGroup.GET("/users",              adminHandler.ListUsers)
		adminGroup.GET("/users/:id",          adminHandler.GetUser)
		adminGroup.PUT("/users/:id",          adminHandler.UpdateUser)
		adminGroup.DELETE("/users/:id",       adminHandler.DeleteUser)
		adminGroup.GET("/workouts",            adminHandler.ListWorkouts)
		adminGroup.GET("/workouts/:id",        adminHandler.GetWorkout)
		adminGroup.PUT("/workouts/:id",        adminHandler.UpdateWorkout)
		adminGroup.POST("/exercises/sync",     adminHandler.SyncExercises)
		adminGroup.GET("/ai-requests",         adminHandler.ListAIRequests)
	}

	r.GET("/ws", auth.WSMiddleware(cfg.JWTSecret), hub.Handler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
