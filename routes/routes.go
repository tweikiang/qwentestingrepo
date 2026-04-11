package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"brave-search-clone/handlers"
	"brave-search-clone/middleware"
)

func SetupRoutes(
	searchHandler *handlers.SearchHandler,
	apiKey string,
	rateLimiter *middleware.RateLimiter,
) *mux.Router {
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.APIKeyAuth(apiKey))
	router.Use(middleware.RateLimit(rateLimiter))

	// Search endpoint
	router.HandleFunc("/search", searchHandler.Search).Methods("GET")

	// Health check endpoint (no auth required for health)
	router.HandleFunc("/health", searchHandler.Health).Methods("GET").Headers("X-Health-Check", "true")

	// Root endpoint with API info
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"name": "Brave Search Clone API",
			"version": "1.0.0",
			"endpoints": {
				"/search": "GET - Search the web",
				"/health": "GET - Health check"
			},
			"parameters": {
				"q": "Search query (required)",
				"limit": "Number of results (1-50, default: 10)",
				"api_key": "API key (if configured)"
			}
		}`))
	}).Methods("GET")

	return router
}
