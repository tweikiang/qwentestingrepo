# TODO: Production-Ready Brave Search API Clone in Golang

## Phase 1: Project Setup & Configuration
- [ ] Create project directory structure
- [ ] Initialize Go module (go.mod)
- [ ] Create `.env.example` with all required environment variables
- [ ] Implement `/internal/config/config.go` - Config loading with viper + godotenv
- [ ] Implement `/pkg/logger/logger.go` - zerolog structured JSON logger

## Phase 2: Infrastructure Layer
- [ ] Implement `/internal/cache/redis.go` - Redis connection and caching utilities
- [ ] Create `docker-compose.yml` - Meilisearch, Redis, and Go app services
- [ ] Create `Dockerfile` - Multi-stage build for Go application

## Phase 3: Search Provider Implementation (Core Logic)
- [ ] Implement `/internal/search/provider.go` - SearchProvider interface definition
- [ ] Implement `/internal/search/ranking.go` - Keyword matching scoring algorithm
- [ ] Implement `/internal/search/scrape.go` - Mode A: DuckDuckGo scraping with goquery
  - [ ] Rotating User-Agents (5+ modern browser UA strings)
  - [ ] HTTP 429/403 handling
  - [ ] Concurrent scraping with goroutines + sync.WaitGroup
  - [ ] URL deduplication
- [ ] Implement `/internal/search/index.go` - Mode B: Meilisearch query implementation

## Phase 4: Middleware Layer
- [ ] Implement `/internal/middleware/auth.go` - API Key authentication middleware
- [ ] Implement `/internal/middleware/ratelimit.go` - Rate limiting (60 req/min per IP)

## Phase 5: API Layer
- [ ] Implement `/internal/api/handler.go` - Search endpoint handler
- [ ] Implement `/internal/api/routes.go` - Route registration with middleware
- [ ] Implement `/cmd/server/main.go` - Entry point with graceful shutdown

## Phase 6: Documentation
- [ ] Create `README.md` with:
  - [ ] Docker Compose run instructions
  - [ ] API key setup guide
  - [ ] 3 curl examples (valid search, missing API key, rate limit)

## Phase 7: Testing & Verification
- [ ] Verify all files are created and code is compilable
- [ ] Test docker-compose up builds successfully
- [ ] Verify all imports are correct and no pseudo-code remains

---

## File Checklist

### Configuration & Setup
- [ ] go.mod
- [ ] .env.example
- [ ] docker-compose.yml
- [ ] Dockerfile

### Internal Packages
- [ ] /internal/config/config.go
- [ ] /internal/middleware/auth.go
- [ ] /internal/middleware/ratelimit.go
- [ ] /internal/api/handler.go
- [ ] /internal/api/routes.go
- [ ] /internal/search/provider.go
- [ ] /internal/search/scrape.go
- [ ] /internal/search/index.go
- [ ] /internal/search/ranking.go
- [ ] /internal/cache/redis.go

### Package
- [ ] /pkg/logger/logger.go

### Entry Point
- [ ] /cmd/server/main.go

### Documentation
- [ ] README.md

---

## Key Implementation Notes

### Strategy Pattern
- Define `SearchProvider` interface with `Search(ctx context.Context, query string, page, limit int) (*SearchResponse, error)`
- `ScrapeProvider` implements interface for Mode A
- `IndexProvider` implements interface for Mode B
- Factory function `NewSearchProvider(mode string) SearchProvider` based on `SEARCH_MODE` env

### Scoring Algorithm (Mode A)
```
Score = (0.5 * TitleMatchRatio) + (0.3 * DescMatchRatio) + (0.2 * DomainAuthority)
```

### Cache Key Format
```
search:{md5(query + page + limit)}
```

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

### Search Response Format
```json
{
  "query": "keyword",
  "results": [...],
  "total": 12345,
  "page": 1,
  "limit": 10
}
```
