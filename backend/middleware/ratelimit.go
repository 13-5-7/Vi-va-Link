package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// ipLimiter はIPアドレスごとのレートリミッターを保持する
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore はIPごとのリミッターを管理するストア
type RateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	b        int
}

// NewRateLimiterStore は新しいRateLimiterStoreを作成する
// r: 1秒あたりのリクエスト数, b: バースト上限
func NewRateLimiterStore(r rate.Limit, b int) *RateLimiterStore {
	store := &RateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		b:        b,
	}
	// 古いエントリを定期的にクリーンアップ（10分ごと）
	go store.cleanupLoop()
	return store
}

func (s *RateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(s.r, s.b)
		s.limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (s *RateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 30*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit はIPベースのレートリミットミドルウェアを返す
// 制限超過時は 429 Too Many Requests を返す
func RateLimit(store *RateLimiterStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := store.getLimiter(ip)
			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "リクエストが多すぎます。しばらく待ってから再試行してください",
					},
				})
			}
			return next(c)
		}
	}
}
