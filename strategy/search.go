package strategy

import (
	"context"
	"brave-search-clone/cache"
)

// SearchResult represents a single search result
type SearchResult struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// SearchProvider defines the interface for different search strategies
type SearchProvider interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	Name() string
}

// SearchService orchestrates search using providers
type SearchService struct {
	provider SearchProvider
	cache    *cache.Cache
}

func NewSearchService(provider SearchProvider, cache *cache.Cache) *SearchService {
	return &SearchService{
		provider: provider,
		cache:    cache,
	}
}

func (s *SearchService) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Try cache first
	if s.cache != nil {
		if cachedResults, found := s.cache.Get(query); found {
			// Convert CachedResult to SearchResult
			results := make([]SearchResult, len(cachedResults))
			for i, r := range cachedResults {
				results[i] = SearchResult{
					Title:       r.Title,
					URL:         r.URL,
					Description: r.Description,
					Score:       r.Score,
				}
			}
			return results, nil
		}
	}

	// Perform search
	results, err := s.provider.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Cache results
	if s.cache != nil && len(results) > 0 {
		cachedResults := make([]cache.CachedResult, len(results))
		for i, r := range results {
			cachedResults[i] = cache.CachedResult{
				Title:       r.Title,
				URL:         r.URL,
				Description: r.Description,
				Score:       r.Score,
			}
		}
		s.cache.Set(query, cachedResults, 0)
	}

	return results, nil
}

func (s *SearchService) SetProvider(provider SearchProvider) {
	s.provider = provider
}
