package server

import (
	"net/http"
	"time"

	"blog/api/internal/config"
	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	return NewRouterWithPostsRepository(cfg, posts.NewMemoryRepository())
}

func NewRouterWithPostsRepository(cfg config.Config, postRepo posts.Repository) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors(cfg.WebOrigin))

	api := router.Group("/api")

	api.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	posts.RegisterPublicRoutes(api, postRepo)

	return router
}

func cors(origin string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
