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

	if err := database.Migrate(db); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	userStore          := database.NewUserStore(db)
	auth.BootstrapAdmin(context.Background(), userStore, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPass, logger)
	workoutStore        := database.NewWorkoutStore(db)
	exerciseStore       := database.NewExerciseStore(db)
	aiRequestStore      := database.NewAIRequestStore(db)
	compareMetricsStore := database.NewCompareMetricsStore(db)

	// Seed exercise library (ON CONFLICT DO NOTHING — safe to run every boot)
	if err := exerciseStore.Seed(context.Background(), exercise.DefaultExercises()); err != nil {
		logger.Error("seed exercises", "error", err)
	}

	// AI providers — only register providers with keys configured.
	// Groq is excluded from workout generators and used only as the neutral grader.
	providers := make(map[string]ai.Provider)
	if cfg.AnthropicKey != "" {
		providers["claude"] = ai.NewClaude(cfg.AnthropicKey)
	}
	if cfg.OpenAIKey != "" {
		providers["openai"] = ai.NewOpenAI(cfg.OpenAIKey)
	}
	if cfg.GeminiKey != "" {
		geminiModels := []struct{ id, name string }{
			{"gemini-2.5-flash", "gemini-2.5-flash"},
			{"gemini-2.5-pro",   "gemini-2.5-pro"},
		}
		for _, m := range geminiModels {
			g, err := ai.NewGemini(cfg.GeminiKey, m.id, m.name)
			if err != nil {
				logger.Error("init gemini provider", "error", err, "model", m.id)
				os.Exit(1)
			}
			providers[m.name] = g
		}
	}

	// Groq as neutral grader — separate from providers map so it evaluates others without self-grading.
	// GroqGrader also satisfies admin.Narrator for the aggregate session analysis endpoint.
	var grader ai.Grader
	var narrator admin.Narrator
	if cfg.GroqKey != "" {
		g := ai.NewGroqGrader(cfg.GroqKey)
		grader = g
		narrator = g
	}

	hub := ws.NewHub(logger, cfg.AllowedOrigins)
	svc := workout.NewService(workoutStore, userStore, providers, hub, aiRequestStore).
		WithExerciseLister(exerciseStore)
	authHandler     := auth.NewHandler(userStore, cfg.JWTSecret)
	workoutHandler  := workout.NewHandler(svc, workoutStore)
	userHandler     := user.NewHandler(userStore, workoutStore)
	exerciseHandler := exercise.NewHandler(exerciseStore, cfg.ExerciseDBKey)
	compareHandler  := ai.NewCompareHandler(providers, grader, exerciseStore, aiRequestStore, compareMetricsStore)

	var exerciseDBClient exercise.GIFFetcher
	if cfg.ExerciseDBKey != "" {
		exerciseDBClient = exercise.NewExerciseDBClient(cfg.ExerciseDBKey)
	}
	syncSvc := exercise.NewSyncService(exerciseStore, exerciseDBClient)

	adminHandler := admin.NewHandler(userStore, workoutStore, syncSvc, aiRequestStore, compareMetricsStore, narrator)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register",                auth.AuthRateLimit(), authHandler.Register)
	v1.POST("/auth/login",                   auth.AuthRateLimit(), authHandler.Login)
	v1.GET("/exercises/image/:exerciseid",   exerciseHandler.ProxyImage) // public — img tags can't send Bearer

	protected := v1.Group("", auth.Middleware(cfg.JWTSecret))
	{
		// Free routes — JWT required but no status check (no AI token spend).
		protected.GET("/exercises", exerciseHandler.List)

		// Active-only routes — pending users are blocked with 403.
		active := protected.Group("", auth.ActiveMiddleware(userStore))
		{
			active.GET("/users/me/stats",          userHandler.Stats)
			active.POST("/workouts/generate",      workoutHandler.Generate)
			active.GET("/workouts",                workoutHandler.List)
			active.GET("/workouts/:id",            workoutHandler.Get)
			active.POST("/workouts/:id/complete",  workoutHandler.Complete)
			active.POST("/ai/compare",             compareHandler.Compare)
		}
	}

	adminGroup := v1.Group("/admin", auth.Middleware(cfg.JWTSecret), auth.AdminMiddleware(userStore))
	{
		adminGroup.GET("/users",              adminHandler.ListUsers)
		adminGroup.GET("/users/:id",          adminHandler.GetUser)
		adminGroup.PUT("/users/:id",          adminHandler.UpdateUser)
		adminGroup.PUT("/users/:id/approve",  adminHandler.ApproveUser)
		adminGroup.DELETE("/users/:id",       adminHandler.DeleteUser)
		adminGroup.GET("/workouts",            adminHandler.ListWorkouts)
		adminGroup.GET("/workouts/:id",        adminHandler.GetWorkout)
		adminGroup.PUT("/workouts/:id",        adminHandler.UpdateWorkout)
		adminGroup.POST("/exercises/sync",     adminHandler.SyncExercises)
		adminGroup.GET("/ai-requests",          adminHandler.ListAIRequests)
		adminGroup.GET("/ai-compare-stats",     adminHandler.CompareStats)
		adminGroup.GET("/compare/latest",       adminHandler.LatestSession)
		adminGroup.POST("/compare/narrative",   adminHandler.NarrativeAnalysis)
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
