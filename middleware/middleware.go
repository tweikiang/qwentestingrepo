package middleware

import (
	"net/http"
	"sync"
	"time"

	"brave-search-clone/config"
	"brave-search-clone/logger"
)

// RateLimiter implements token bucket rate limiting per IP
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	tokens   int
}

func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    cfg.RequestsPerMinute,
		window:   time.Minute,
	}

	// Cleanup old visitors periodically
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			lastSeen: time.Now(),
			tokens:   rl.limit - 1,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := time.Since(v.lastSeen)
	tokensToAdd := int(elapsed.Seconds()) * rl.limit / int(rl.window.Seconds())
	v.tokens = min(rl.limit, v.tokens+tokensToAdd)
	v.lastSeen = time.Now()

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// APIKeyAuth middleware validates API key
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}

			if key != apiKey {
				logger.Log.Warn().Str("ip", getClientIP(r)).Msg("Invalid API key")
				http.Error(w, `{"error":"Unauthorized","message":"Invalid or missing API key"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit middleware applies rate limiting
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !rl.Allow(ip) {
				logger.Log.Warn().Str("ip", ip).Msg("Rate limit exceeded")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too Many Requests","message":"Rate limit exceeded. Try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
