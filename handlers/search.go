package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"brave-search-clone/strategy"
	"brave-search-clone/logger"
)

type SearchHandler struct {
	service *strategy.SearchService
}

type SearchResponse struct {
	Query      string                `json:"query"`
	Results    []strategy.SearchResult `json:"results"`
	Count      int                   `json:"count"`
	Timestamp  string                `json:"timestamp"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewSearchHandler(service *strategy.SearchService) *SearchHandler {
	return &SearchHandler{
		service: service,
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendError(w, "Bad Request", "Missing required parameter: q", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	logger.Log.Info().Str("query", query).Int("limit", limit).Msg("Search request")

	results, err := h.service.Search(ctx, query, limit)
	if err != nil {
		logger.Log.Error().Err(err).Str("query", query).Msg("Search failed")
		// Return empty results instead of error on timeout/context cancellation
		results = []strategy.SearchResult{}
	}

	response := SearchResponse{
		Query:     query,
		Results:   results,
		Count:     len(results),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SearchHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *SearchHandler) sendError(w http.ResponseWriter, err string, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   err,
		Message: msg,
	})
}
