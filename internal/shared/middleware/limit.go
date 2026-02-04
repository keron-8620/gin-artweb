package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"gin-artweb/internal/shared/errors"
)

// IPRateLimiter IP限流器管理
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

// NewIPRateLimiter 创建IP限流器管理器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// GetLimiter 获取指定IP的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.limiters[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		// 双重检查防止并发创建
		limiter, exists = i.limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(i.r, i.b)
			i.limiters[ip] = limiter
		}
		i.mu.Unlock()
	}

	return limiter
}

// GlobalRateLimiterMiddleware 全局限流中间件
func GlobalRateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			code := http.StatusTooManyRequests
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrRateLimitExceeded))
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
			code := http.StatusTooManyRequests
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrRateLimitExceeded))
			return
		}
		c.Next()
	}
}

// RateLimiterMiddleware 向后兼容的限流中间件
func RateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	return GlobalRateLimiterMiddleware(r, b)
}
