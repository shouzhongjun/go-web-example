package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"goWebExample/internal/configs"
	"goWebExample/internal/infra/cache"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// IPRateLimiter IP限流器
type IPRateLimiter struct {
	ips       map[string]*rate.Limiter
	lastUsed  map[string]time.Time
	mu        *sync.RWMutex
	r         rate.Limit
	b         int
	expire    time.Duration
	redisConn *cache.RedisConnector // Redis连接器
	redisKey  string                // Redis中存储限流器数据的键前缀
}

// NewIPRateLimiter 创建一个新的IP限流器
func NewIPRateLimiter(r rate.Limit, b int, redisConn *cache.RedisConnector) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:       make(map[string]*rate.Limiter),
		lastUsed:  make(map[string]time.Time),
		mu:        &sync.RWMutex{},
		r:         r,
		b:         b,
		expire:    time.Hour, // 默认过期时间
		redisConn: redisConn,
		redisKey:  "rate_limiter:",
	}

	// 启动清理过期限流器的goroutine
	go limiter.CleanupLimiters()

	return limiter
}

// 将限流器数据保存到Redis
func (i *IPRateLimiter) saveToRedis(ctx context.Context, ip string, lastUsed time.Time) error {
	if i.redisConn == nil || !i.redisConn.IsConnected() {
		return nil // Redis未连接，跳过保存
	}

	data := map[string]interface{}{
		"last_used": lastUsed.Unix(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化限流器数据失败: %w", err)
	}

	key := i.redisKey + ip
	err = i.redisConn.GetClient().Set(ctx, key, jsonData, i.expire).Err()
	if err != nil {
		return fmt.Errorf("保存限流器数据到Redis失败: %w", err)
	}

	return nil
}

// 从Redis加载限流器数据
func (i *IPRateLimiter) loadFromRedis(ctx context.Context, ip string) (time.Time, bool, error) {
	if i.redisConn == nil || !i.redisConn.IsConnected() {
		return time.Time{}, false, nil // Redis未连接，返回空数据
	}

	key := i.redisKey + ip
	jsonData, err := i.redisConn.GetClient().Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return time.Time{}, false, nil // 键不存在
		}
		return time.Time{}, false, fmt.Errorf("从Redis加载限流器数据失败: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return time.Time{}, false, fmt.Errorf("解析限流器数据失败: %w", err)
	}

	lastUsedUnix, ok := data["last_used"].(float64)
	if !ok {
		return time.Time{}, false, fmt.Errorf("限流器数据格式错误")
	}

	lastUsed := time.Unix(int64(lastUsedUnix), 0)
	return lastUsed, true, nil
}

// GetLimiter 获取指定IP的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	now := time.Now()
	ctx := context.Background()

	if !exists {
		// 尝试从Redis加载
		if i.redisConn != nil && i.redisConn.IsConnected() {
			lastUsed, found, err := i.loadFromRedis(ctx, ip)
			if err == nil && found {
				i.mu.Lock()
				limiter = rate.NewLimiter(i.r, i.b)
				i.ips[ip] = limiter
				i.lastUsed[ip] = lastUsed
				i.mu.Unlock()

				// 更新最后使用时间
				i.mu.Lock()
				i.lastUsed[ip] = now
				i.mu.Unlock()

				// 保存到Redis
				_ = i.saveToRedis(ctx, ip, now)

				return limiter
			}
		}

		// Redis中不存在或Redis未连接，创建新的限流器
		i.mu.Lock()
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
		i.lastUsed[ip] = now
		i.mu.Unlock()

		// 保存到Redis
		if i.redisConn != nil && i.redisConn.IsConnected() {
			_ = i.saveToRedis(ctx, ip, now)
		}
	} else {
		// 更新最后使用时间
		i.mu.Lock()
		i.lastUsed[ip] = now
		i.mu.Unlock()

		// 保存到Redis
		if i.redisConn != nil && i.redisConn.IsConnected() {
			_ = i.saveToRedis(ctx, ip, now)
		}
	}

	return limiter
}

// CleanupLimiters 清理过期的限流器
func (i *IPRateLimiter) CleanupLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		i.mu.Lock()
		for ip, lastAccess := range i.lastUsed {
			// 只删除过期的限流器
			if now.Sub(lastAccess) > i.expire {
				delete(i.ips, ip)
				delete(i.lastUsed, ip)

				// 从Redis中删除
				if i.redisConn != nil && i.redisConn.IsConnected() {
					ctx := context.Background()
					key := i.redisKey + ip
					_ = i.redisConn.GetClient().Del(ctx, key).Err()
				}
			}
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config *configs.AllConfig, redisConn *cache.RedisConnector) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(config.RateLimiter.Rate), config.RateLimiter.Burst, redisConn)

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
