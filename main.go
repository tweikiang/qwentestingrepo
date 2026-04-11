package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brave-search-clone/config"
	"brave-search-clone/logger"
	"brave-search-clone/cache"
	"brave-search-clone/strategy"
	"brave-search-clone/providers"
	"brave-search-clone/middleware"
	"brave-search-clone/handlers"
	"brave-search-clone/routes"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Log.Info().Msg("Starting Brave Search Clone API")

	// Initialize cache
	redisCache := cache.NewCache(&cfg.Redis)
	defer redisCache.Close()

	// Initialize search provider (ScrapeProvider by default)
	scrapeProvider := providers.NewScrapeProvider()

	// Initialize search service
	searchService := strategy.NewSearchService(scrapeProvider, redisCache)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(searchService)

	// Setup routes
	router := routes.SetupRoutes(searchHandler, cfg.Server.APIKey, rateLimiter)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Log.Info().Msg("Server exited")
}
