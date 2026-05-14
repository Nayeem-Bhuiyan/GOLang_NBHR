package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"nbhr/internal/constants"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Strategy interface — makes the limiter backend swappable (in-memory / Redis)
// ---------------------------------------------------------------------------

// LimiterStore defines the backend contract for rate limiting.
type LimiterStore interface {
	// Allow checks whether the key is within its allowed quota.
	// Returns (allowed, remaining, resetAt, err).
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error)
	// Reset clears the counter for the given key.
	Reset(ctx context.Context, key string) error
}

// ---------------------------------------------------------------------------
// RateLimitConfig
// ---------------------------------------------------------------------------

// RateLimitConfig holds rate-limiter options.
type RateLimitConfig struct {
	// Requests is the maximum number of requests allowed per Window.
	Requests int64
	// Window is the sliding/fixed window duration.
	Window time.Duration
	// KeyFunc derives the rate-limit bucket key from the request.
	// Defaults to ClientIP if nil.
	KeyFunc func(c *gin.Context) string
	// Store is the backing LimiterStore. Defaults to in-process memory store.
	Store LimiterStore
	// Logger is optional; nil disables limit-hit logging.
	Logger *zap.Logger
	// TrustProxy controls whether X-Forwarded-For is trusted for IP extraction.
	TrustProxy bool
	// SkipFunc — return true to bypass rate limiting for a specific request.
	SkipFunc func(c *gin.Context) bool
}

func (cfg *RateLimitConfig) applyDefaults() {
	if cfg.Requests <= 0 {
		cfg.Requests = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Store == nil {
		cfg.Store = NewMemoryStore()
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = defaultKeyFunc(cfg.TrustProxy)
	}
}

// ---------------------------------------------------------------------------
// Middleware constructors
// ---------------------------------------------------------------------------

// RateLimit returns a general-purpose rate-limiting middleware.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	cfg.applyDefaults()

	return func(c *gin.Context) {
		if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)
		allowed, remaining, resetAt, err := cfg.Store.Allow(
			c.Request.Context(), key, cfg.Requests, cfg.Window,
		)
		if err != nil {
			// On store errors, fail open to avoid blocking legitimate traffic.
			if cfg.Logger != nil {
				cfg.Logger.Error("rate limit store error", zap.String("key", key), zap.Error(err))
			}
			c.Next()
			return
		}

		retryAfter := time.Until(resetAt).Seconds()
		writeRateLimitHeaders(c, cfg.Requests, remaining, resetAt, retryAfter)

		if !allowed {
			if cfg.Logger != nil {
				requestID, _ := c.Get(constants.ContextKeyRequestID)
				cfg.Logger.Warn("rate limit exceeded",
					zap.String("key", key),
					zap.String("request_id", toString(requestID)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"message":     "too many requests — please slow down",
				"retry_after": retryAfter,
			})
			return
		}

		c.Next()
	}
}

// RateLimitByIP is a convenience constructor keyed purely on client IP.
func RateLimitByIP(requests int64, window time.Duration, log *zap.Logger) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Requests:   requests,
		Window:     window,
		TrustProxy: false,
		Logger:     log,
	})
}

// RateLimitByUser is a convenience constructor keyed on the authenticated
// user ID injected by the Authenticate middleware. Falls back to IP when no
// user is present (public endpoints).
func RateLimitByUser(requests int64, window time.Duration, log *zap.Logger) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Requests: requests,
		Window:   window,
		Logger:   log,
		KeyFunc: func(c *gin.Context) string {
			if uid, exists := c.Get(constants.ContextKeyUserID); exists {
				return fmt.Sprintf("user:%v", uid)
			}
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
	})
}

// RateLimitByRoute keys on method + path, useful for protecting specific
// endpoints (e.g. login, register) with tighter limits.
func RateLimitByRoute(requests int64, window time.Duration, log *zap.Logger) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Requests: requests,
		Window:   window,
		Logger:   log,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("route:%s:%s:%s", c.Request.Method, c.FullPath(), c.ClientIP())
		},
	})
}

// ---------------------------------------------------------------------------
// Header helpers
// ---------------------------------------------------------------------------

// writeRateLimitHeaders sets standard rate-limit response headers.
func writeRateLimitHeaders(c *gin.Context, limit, remaining int64, resetAt time.Time, retryAfter float64) {
	c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(max64(remaining, 0), 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
	if remaining <= 0 {
		c.Header("Retry-After", strconv.FormatInt(int64(retryAfter)+1, 10))
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Default key function
// ---------------------------------------------------------------------------

func defaultKeyFunc(trustProxy bool) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		if trustProxy {
			if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
				return fmt.Sprintf("ip:%s", xff)
			}
		}
		return fmt.Sprintf("ip:%s", c.ClientIP())
	}
}

// ---------------------------------------------------------------------------
// In-process memory store (zero external dependencies)
// ---------------------------------------------------------------------------

// bucket is a single sliding-window counter.
type bucket struct {
	mu        sync.Mutex
	count     int64
	windowEnd time.Time
}

// MemoryStore is a thread-safe, in-process rate-limit store using fixed windows.
// For multi-instance deployments replace with a Redis-backed store.
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	// cleanupInterval controls how often expired buckets are purged.
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewMemoryStore constructs an in-memory LimiterStore and starts a background
// cleanup goroutine to prevent unbounded memory growth.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		buckets:         make(map[string]*bucket),
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Allow implements LimiterStore.
func (s *MemoryStore) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	s.mu.Lock()
	b, ok := s.buckets[key]
	if !ok {
		b = &bucket{}
		s.buckets[key] = b
	}
	s.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// Start a new window when the previous one expired.
	if now.After(b.windowEnd) {
		b.count = 0
		b.windowEnd = now.Add(window)
	}

	b.count++
	remaining := limit - b.count
	allowed := b.count <= limit

	return allowed, remaining, b.windowEnd, nil
}

// Reset implements LimiterStore.
func (s *MemoryStore) Reset(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.buckets, key)
	return nil
}

// Stop halts the background cleanup goroutine.
func (s *MemoryStore) Stop() {
	close(s.stopCleanup)
}

// cleanup periodically removes expired buckets.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.purgeExpired()
		case <-s.stopCleanup:
			return
		}
	}
}

func (s *MemoryStore) purgeExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, b := range s.buckets {
		b.mu.Lock()
		expired := now.After(b.windowEnd)
		b.mu.Unlock()
		if expired {
			delete(s.buckets, key)
		}
	}
}