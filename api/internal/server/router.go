package server

import (
	"net/http"
	"time"

	"blog/api/internal/config"
	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/comments"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/posts"
	"blog/api/internal/modules/reactions"
	"blog/api/internal/modules/submissions"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	return NewRouterWithRepositories(cfg, Repositories{})
}

func NewRouterWithPostsRepository(cfg config.Config, postRepo posts.Repository) *gin.Engine {
	return NewRouterWithRepositories(cfg, Repositories{PostRepo: postRepo})
}

type Repositories struct {
	AuthStore      auth.Store
	PostRepo       posts.Repository
	CommentRepo    comments.Repository
	ReactionRepo   reactions.Repository
	MessageRepo    messages.Repository
	SubmissionRepo submissions.Repository
}

func NewRouterWithRepositories(cfg config.Config, repos Repositories) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if repos.AuthStore == nil {
		repos.AuthStore = auth.NewMemoryStore()
	}
	if repos.PostRepo == nil {
		repos.PostRepo = posts.NewMemoryRepository()
	}
	if repos.CommentRepo == nil {
		repos.CommentRepo = comments.NewMemoryRepository()
	}
	if repos.ReactionRepo == nil {
		repos.ReactionRepo = reactions.NewMemoryRepository()
	}
	if repos.MessageRepo == nil {
		repos.MessageRepo = messages.NewMemoryRepository()
	}
	if repos.SubmissionRepo == nil {
		repos.SubmissionRepo = submissions.NewMemoryRepository()
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors(cfg.WebOrigin))

	router.Use(auth.Middleware(repos.AuthStore))

	api := router.Group("/api")

	api.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	auth.RegisterRoutes(api, repos.AuthStore)
	posts.RegisterPublicRoutes(api, repos.PostRepo)
	comments.RegisterRoutes(api, repos.CommentRepo)
	reactions.RegisterRoutes(api, repos.ReactionRepo, repos.PostRepo)
	messages.RegisterRoutes(api, repos.MessageRepo)

	var publisher posts.Publisher
	if item, ok := repos.PostRepo.(posts.Publisher); ok {
		publisher = item
	}
	submissions.RegisterRoutes(api, repos.SubmissionRepo, repos.MessageRepo, publisher)

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
