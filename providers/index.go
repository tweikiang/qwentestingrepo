package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"brave-search-clone/strategy"
	"brave-search-clone/config"
)

// IndexProvider implements search using Meilisearch index
type IndexProvider struct {
	httpClient *http.Client
	host       string
	apiKey     string
}

type MeiliSearchRequest struct {
	Q     string `json:"q"`
	Limit int    `json:"limit"`
}

type MeiliSearchResponse struct {
	Hits      []MeiliHit `json:"hits"`
	Query     string     `json:"query"`
	ProcessingTimeMs int `json:"processingTimeMs"`
}

type MeiliHit struct {
	Title       string                 `json:"title"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Rank        int                    `json:"_rankingScore"`
	Additional  map[string]interface{} `json:"-"`
}

func NewIndexProvider(cfg *config.MeilisearchConfig) *IndexProvider {
	return &IndexProvider{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		host:   cfg.Host,
		apiKey: cfg.Key,
	}
}

func (p *IndexProvider) Name() string {
	return "index"
}

func (p *IndexProvider) Search(ctx context.Context, query string, limit int) ([]strategy.SearchResult, error) {
	searchURL := fmt.Sprintf("%s/indexes/search_index/search", p.host)

	reqBody := MeiliSearchRequest{
		Q:     query,
		Limit: limit,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var meiliResp MeiliSearchResponse
	if err := json.Unmarshal(body, &meiliResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]strategy.SearchResult, 0, len(meiliResp.Hits))
	for _, hit := range meiliResp.Hits {
		score := float64(hit.Rank)
		if score == 0 {
			score = 1.0
		}

		results = append(results, strategy.SearchResult{
			Title:       hit.Title,
			URL:         hit.URL,
			Description: hit.Description,
			Score:       score,
		})
	}

	return results, nil
}
