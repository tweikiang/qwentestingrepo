# Brave Search Clone API

A high-performance web search API inspired by Brave Search, built with Go. Features real-time web scraping, Redis caching, Meilisearch indexing, rate limiting, and API key authentication.

## Features

- **Strategy Pattern**: Pluggable search providers (ScrapeProvider for DuckDuckGo, IndexProvider for Meilisearch)
- **Redis Caching**: Fast response times with automatic result caching
- **Meilisearch Integration**: Optional full-text search index support
- **Rate Limiting**: 60 requests per minute per IP (configurable)
- **API Key Authentication**: Optional Bearer token authentication
- **Concurrent Scraping**: Parallel fetching with rotating User-Agents
- **Smart Ranking**: Results scored by relevance to query terms
- **Graceful Shutdown**: Clean server termination on SIGINT/SIGTERM

## Project Structure

```
brave-search-clone/
├── main.go              # Application entry point
├── config/              # Configuration management (viper + godotenv)
├── logger/              # Structured logging (zerolog)
├── cache/               # Redis cache layer
├── strategy/            # Strategy pattern implementation
├── providers/           # Search providers (scrape, index)
├── middleware/          # Auth & rate limiting middleware
├── handlers/            # HTTP request handlers
├── routes/              # Router setup
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Full stack orchestration
└── .env.example         # Environment template
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Copy environment template
cp .env.example .env

# Start all services (app + Redis + Meilisearch)
docker-compose up -d --build

# Test the API
curl "http://localhost:8080/search?q=golang+tutorial&limit=10"
```

### Local Development

**Prerequisites:**
- Go 1.19+
- Redis server running on localhost:6379
- (Optional) Meilisearch running on localhost:7700

```bash
# Install dependencies
go mod download

# Copy and configure environment
cp .env.example .env

# Run the application
go run main.go

# Test in another terminal
curl "http://localhost:8080/search?q=hello+world"
```

## API Endpoints

### GET /search

Search the web with optional parameters.

**Parameters:**
| Parameter | Type   | Required | Default | Description                    |
|-----------|--------|----------|---------|--------------------------------|
| q         | string | Yes      | -       | Search query                   |
| limit     | int    | No       | 10      | Number of results (1-50)       |
| api_key   | string | No*      | -       | API key (if authentication enabled) |

**Example Request:**
```bash
curl "http://localhost:8080/search?q=artificial+intelligence&limit=5"
```

**Example Response:**
```json
{
  "query": "artificial intelligence",
  "results": [
    {
      "title": "Introduction to AI",
      "url": "https://example.com/ai-intro",
      "description": "Learn the basics of artificial intelligence...",
      "score": 3.5
    }
  ],
  "count": 1,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /health

Health check endpoint.

```bash
curl -H "X-Health-Check: true" http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:00Z"
}
```

### GET /

API information and documentation.

```bash
curl http://localhost:8080/
```

## Configuration

All configuration is done via environment variables or `.env` file:

| Variable              | Default     | Description                          |
|-----------------------|-------------|--------------------------------------|
| SERVER_PORT           | 8080        | HTTP server port                     |
| API_KEY               | (empty)     | API key for authentication           |
| REDIS_HOST            | localhost   | Redis server hostname                |
| REDIS_PORT            | 6379        | Redis server port                    |
| REDIS_PASSWORD        | (empty)     | Redis password                       |
| REDIS_DB              | 0           | Redis database number                |
| MEILISEARCH_HOST      | localhost:7700 | Meilisearch server URL           |
| MEILISEARCH_KEY       | (empty)     | Meilisearch API key                  |
| RATE_LIMIT_REQUESTS   | 60          | Max requests per minute per IP       |

## Architecture

### Strategy Pattern

The search system uses the Strategy pattern for flexible provider switching:

```go
type SearchProvider interface {
    Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
    Name() string
}
```

**Providers:**
- `ScrapeProvider`: Scrapes DuckDuckGo HTML results with rotating User-Agents
- `IndexProvider`: Queries Meilisearch for indexed content

### Scoring Algorithm

Results are ranked using a multi-factor scoring system:
- Base score: 1.0
- Exact query match in title: +2.0
- Query word in title: +0.5 per word
- Query word in description: +0.2 per word

### Caching Strategy

- Cache key format: `search:{query}`
- Cache TTL: Configurable (default: no expiration)
- Cache stored as JSON in Redis

## Error Handling

All errors return JSON responses:

```json
{
  "error": "Error Type",
  "message": "Human-readable description"
}
```

**HTTP Status Codes:**
- `200 OK`: Successful request
- `400 Bad Request`: Missing or invalid parameters
- `401 Unauthorized`: Invalid or missing API key
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server-side error

## License

MIT License