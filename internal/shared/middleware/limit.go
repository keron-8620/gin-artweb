package middleware

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"gin-artweb/internal/shared/errors"
)

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

// IPRateLimiter IP限流器管理
type IPRateLimiter struct {
	limiters map[string]*ipLimiterEntry
	mu       sync.RWMutex
	r        rate.Limit
	b        int
	ttl      time.Duration
}

// NewIPRateLimiter 创建IP限流器管理器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters: make(map[string]*ipLimiterEntry),
		r:        r,
		b:        b,
		ttl:      10 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (i *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		i.mu.Lock()
		now := time.Now().UnixNano()
		for ip, entry := range i.limiters {
			if now-entry.lastSeen.Load() > int64(i.ttl) {
				delete(i.limiters, ip)
			}
		}
		i.mu.Unlock()
	}
}

// GetLimiter 获取指定IP的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.limiters[ip]
	i.mu.RUnlock()

	if exists {
		entry.lastSeen.Store(time.Now().UnixNano())
		return entry.limiter
	}

	i.mu.Lock()
	entry, exists = i.limiters[ip]
	if !exists {
		entry = &ipLimiterEntry{
			limiter: rate.NewLimiter(i.r, i.b),
		}
		entry.lastSeen.Store(time.Now().UnixNano())
		i.limiters[ip] = entry
	}
	i.mu.Unlock()

	return entry.limiter
}

// GlobalRateLimiterMiddleware 全局限流中间件
func GlobalRateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			errors.RespondWithError(c, errors.ErrRateLimitExceeded)
			return
		}
		c.Next()
	}
}

// IPBasedRateLimiterMiddleware IP限流中间件
func IPBasedRateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	ipLimiter := NewIPRateLimiter(r, b)

	return func(c *gin.Context) {
		limiter := ipLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			errors.RespondWithError(c, errors.ErrRateLimitExceeded)
			return
		}
		c.Next()
	}
}

// RateLimiterMiddleware 向后兼容的限流中间件
func RateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	return GlobalRateLimiterMiddleware(r, b)
}
