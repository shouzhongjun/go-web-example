package middleware

import (
	"goWebExample/internal/configs"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter IP限流器
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     *sync.RWMutex
	r      rate.Limit
	b      int
	expire time.Duration
}

// NewIPRateLimiter 创建一个新的IP限流器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mu:     &sync.RWMutex{},
		r:      r,
		b:      b,
		expire: time.Hour, // 默认过期时间
	}
}

// GetLimiter 获取指定IP的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
		i.mu.Unlock()
	}

	return limiter
}

// CleanupLimiters 清理过期的限流器
func (i *IPRateLimiter) CleanupLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip := range i.ips {
			delete(i.ips, ip)
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config *configs.AllConfig) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(config.RateLimiter.Rate), config.RateLimiter.Burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
