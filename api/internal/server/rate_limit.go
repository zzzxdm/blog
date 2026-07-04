package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	WindowStart time.Time
	Count       int
}

func rateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateBucket{}

	return func(ctx *gin.Context) {
		if !isWriteMethod(ctx.Request.Method) {
			ctx.Next()
			return
		}

		now := time.Now()
		key := fmt.Sprintf("%s:%s", ctx.ClientIP(), ctx.Request.URL.Path)

		mu.Lock()
		bucket := buckets[key]
		if bucket.WindowStart.IsZero() || now.Sub(bucket.WindowStart) >= window {
			bucket = rateBucket{WindowStart: now}
		}
		bucket.Count++
		buckets[key] = bucket

		allowed := bucket.Count <= maxRequests
		if len(buckets) > 4096 {
			for itemKey, item := range buckets {
				if now.Sub(item.WindowStart) >= window {
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

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
