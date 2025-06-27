package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexnthnz/url-shortener/internal/config"
	"github.com/alexnthnz/url-shortener/internal/handlers"
	"github.com/alexnthnz/url-shortener/internal/repository"
	"github.com/alexnthnz/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := repository.RunMigrations(db); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis cache
	cache := repository.NewRedisCache(cfg.RedisURL)
	defer cache.Close()

	// Initialize repositories
	urlRepo := repository.NewURLRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Initialize services
	urlService := services.NewURLService(urlRepo, cache, logger)
	analyticsService := services.NewAnalyticsService(analyticsRepo, logger)

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(urlService, analyticsService, logger)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handlers.LoggerMiddleware(logger))
	router.Use(handlers.CORSMiddleware())
	router.Use(handlers.SecurityMiddleware())
	router.Use(handlers.RateLimitMiddleware(cache))

	// Setup routes
	setupRoutes(router, urlHandler)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func setupRoutes(router *gin.Engine, urlHandler *handlers.URLHandler) {
	// Health check
	router.GET("/health", urlHandler.HealthCheck)

	// Metrics endpoint
	router.GET("/metrics", urlHandler.MetricsHandler)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/shorten", urlHandler.ShortenURL)
		api.GET("/urls/:short_code/stats", urlHandler.GetURLStats)
	}

	// Redirect route
	router.GET("/:short_code", urlHandler.RedirectURL)
}
