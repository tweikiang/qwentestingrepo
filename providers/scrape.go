package providers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"brave-search-clone/strategy"
	"brave-search-clone/logger"
)

// ScrapeProvider implements search by scraping DuckDuckGo
type ScrapeProvider struct {
	httpClient *http.Client
	userAgents []string
	mu         sync.Mutex
	seenURLs   map[string]bool
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
}

func NewScrapeProvider() *ScrapeProvider {
	return &ScrapeProvider{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgents: userAgents,
		seenURLs:   make(map[string]bool),
	}
}

func (p *ScrapeProvider) Name() string {
	return "scrape"
}

func (p *ScrapeProvider) Search(ctx context.Context, query string, limit int) ([]strategy.SearchResult, error) {
	p.mu.Lock()
	p.seenURLs = make(map[string]bool)
	p.mu.Unlock()

	resultsChan := make(chan strategy.SearchResult, limit*2)
	errorChan := make(chan error, 3)

	// Launch concurrent searches
	go p.searchDuckDuckGo(ctx, query, limit, resultsChan, errorChan)

	var results []strategy.SearchResult
	timeout := time.After(15 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-timeout:
			return p.rankAndLimit(results, limit), nil
		case result := <-resultsChan:
			if !p.isDuplicate(result.URL) {
				p.markSeen(result.URL)
				results = append(results, result)
				if len(results) >= limit*2 {
					return p.rankAndLimit(results, limit), nil
				}
			}
		case err := <-errorChan:
			logger.Log.Warn().Err(err).Msg("Search error")
		}
	}
}

func (p *ScrapeProvider) searchDuckDuckGo(ctx context.Context, query string, limit int, resultsChan chan<- strategy.SearchResult, errorChan chan<- error) {
	baseURL := "https://html.duckduckgo.com/html/"
	params := url.Values{}
	params.Set("q", query)

	reqURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		errorChan <- err
		return
	}

	ua := p.getRandomUA()
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		errorChan <- err
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorChan <- err
		return
	}

	parsedResults := p.parseHTML(string(body), query)
	for _, r := range parsedResults {
		resultsChan <- r
	}
}

func (p *ScrapeProvider) parseHTML(html string, query string) []strategy.SearchResult {
	var results []strategy.SearchResult

	// Simple regex-based parsing for DuckDuckGo HTML results
	resultRe := regexp.MustCompile(`(?s)<a[^>]*class="result__a"[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	snippetRe := regexp.MustCompile(`(?s)<a[^>]*class="result__snippet"[^>]*>([^<]*(?:<[^>]*>[^<]*)*)</a>`)

	matches := resultRe.FindAllStringSubmatch(html, -1)
	snippets := snippetRe.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := match[1]
		title := cleanHTML(match[2])

		// Skip internal DDG URLs
		if strings.Contains(url, "duckduckgo.com") && !strings.Contains(url, "adclick") {
			continue
		}

		// Clean URL if it's a redirect
		if strings.HasPrefix(url, "//") {
			url = "https:" + url
		}

		description := ""
		if i < len(snippets) && len(snippets[i]) > 1 {
			description = cleanHTML(snippets[i][1])
		}

		score := calculateScore(title, description, query)

		results = append(results, strategy.SearchResult{
			Title:       title,
			URL:         url,
			Description: description,
			Score:       score,
		})
	}

	return results
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.TrimSpace(s)
	return s
}

func (p *ScrapeProvider) getRandomUA() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := time.Now().UnixNano() % int64(len(p.userAgents))
	return p.userAgents[idx]
}

func (p *ScrapeProvider) isDuplicate(urlStr string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seenURLs[urlStr]
}

func (p *ScrapeProvider) markSeen(urlStr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seenURLs[urlStr] = true
}

func calculateScore(title, description, query string) float64 {
	score := 1.0
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)
	descLower := strings.ToLower(description)

	// Exact title match
	if strings.Contains(titleLower, queryLower) {
		score += 2.0
	}

	// Words in title
	words := strings.Fields(queryLower)
	for _, word := range words {
		if len(word) > 2 && strings.Contains(titleLower, word) {
			score += 0.5
		}
		if len(word) > 2 && strings.Contains(descLower, word) {
			score += 0.2
		}
	}

	return score
}

func (p *ScrapeProvider) rankAndLimit(results []strategy.SearchResult, limit int) []strategy.SearchResult {
	// Sort by score (simple bubble sort for small datasets)
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// Ensure ScrapeProvider implements json.Marshaler for proper serialization
func (p *ScrapeProvider) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"type": "scrape"})
}
