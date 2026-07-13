package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type rateBucket struct {
	WindowStart time.Time
	Count       int
}

type rateLimitPolicy struct {
	MaxRequests int
	Window      time.Duration
}

func rateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return rateLimitWithRedis(nil, maxRequests, window)
}

func rateLimitWithRedis(client *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return rateLimitByPolicy(client, func(ctx *gin.Context) (rateLimitPolicy, bool) {
		if !isWriteMethod(ctx.Request.Method) {
			return rateLimitPolicy{}, false
		}

		return rateLimitPolicy{MaxRequests: maxRequests, Window: window}, true
	})
}

func authSensitiveRateLimit() gin.HandlerFunc {
	return authSensitiveRateLimitWithRedis(nil)
}

func authSensitiveRateLimitWithRedis(client *redis.Client) gin.HandlerFunc {
	policies := map[string]rateLimitPolicy{
		"/api/auth/login":              {MaxRequests: 20, Window: time.Minute},
		"/api/auth/register":           {MaxRequests: 5, Window: 10 * time.Minute},
		"/api/auth/forgot-password":    {MaxRequests: 5, Window: 10 * time.Minute},
		"/api/auth/email-verification": {MaxRequests: 3, Window: 10 * time.Minute},
		"/api/auth/verify-email":       {MaxRequests: 20, Window: 10 * time.Minute},
		"/api/auth/reset-password":     {MaxRequests: 10, Window: 10 * time.Minute},
	}

	return rateLimitByPolicy(client, func(ctx *gin.Context) (rateLimitPolicy, bool) {
		if ctx.Request.Method != http.MethodPost {
			return rateLimitPolicy{}, false
		}

		policy, ok := policies[ctx.Request.URL.Path]
		return policy, ok
	})
}

func rateLimitByPolicy(client *redis.Client, policyFor func(*gin.Context) (rateLimitPolicy, bool)) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateBucket{}

	return func(ctx *gin.Context) {
		policy, ok := policyFor(ctx)
		if !ok {
			ctx.Next()
			return
		}

		key := fmt.Sprintf("%s:%s", ctx.ClientIP(), ctx.Request.URL.Path)
		if client != nil {
			allowed, err := redisAllow(ctx.Request.Context(), client, "ratelimit:"+key, policy.MaxRequests, policy.Window)
			if err == nil {
				if !allowed {
					ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
					return
				}
				ctx.Next()
				return
			}
			// fall through to memory on redis errors
		}

		now := time.Now()
		mu.Lock()
		bucket := buckets[key]
		if bucket.WindowStart.IsZero() || now.Sub(bucket.WindowStart) >= policy.Window {
			bucket = rateBucket{WindowStart: now}
		}
		bucket.Count++
		buckets[key] = bucket

		allowed := bucket.Count <= policy.MaxRequests
		if len(buckets) > 4096 {
			for itemKey, item := range buckets {
				if now.Sub(item.WindowStart) >= policy.Window {
					delete(buckets, itemKey)
				}
			}
		}
		mu.Unlock()

		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		ctx.Next()
	}
}

func redisAllow(ctx context.Context, client *redis.Client, key string, maxRequests int, window time.Duration) (bool, error) {
	pipe := client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}

	count, err := incr.Result()
	if err != nil {
		return false, err
	}

	return count <= int64(maxRequests), nil
}

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
