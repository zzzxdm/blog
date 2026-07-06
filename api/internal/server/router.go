package server

import (
	"context"
	"net/http"
	"time"

	"blog/api/internal/config"
	"blog/api/internal/modules/adminposts"
	"blog/api/internal/modules/auth"
	"blog/api/internal/modules/comments"
	"blog/api/internal/modules/feeds"
	"blog/api/internal/modules/messages"
	"blog/api/internal/modules/operations"
	"blog/api/internal/modules/posts"
	"blog/api/internal/modules/reactions"
	"blog/api/internal/modules/seo"
	"blog/api/internal/modules/submissions"
	"blog/api/internal/modules/taxonomies"
	"blog/api/internal/modules/topics"
	"blog/api/internal/modules/users"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	return NewRouterWithRepositories(cfg, Repositories{})
}

func NewRouterWithPostsRepository(cfg config.Config, postRepo posts.Repository) *gin.Engine {
	return NewRouterWithRepositories(cfg, Repositories{PostRepo: postRepo})
}

type Repositories struct {
	AuthStore         auth.Store
	PostRepo          posts.Repository
	CommentRepo       comments.Repository
	ReactionRepo      reactions.Repository
	MessageRepo       messages.Repository
	SubmissionRepo    submissions.Repository
	OperationsRepo    operations.Repository
	UserRepo          users.Repository
	AdminPostRepo     adminposts.Repository
	TaxonomyRepo      taxonomies.Repository
	TopicRepo         topics.Repository
	AuthEmailSender   auth.EmailSender
	TurnstileVerifier auth.TurnstileVerifier
}

type authSecuritySettingsReader struct {
	repo operations.Repository
}

func (reader authSecuritySettingsReader) SecuritySettings(ctx context.Context) (auth.SecuritySettings, error) {
	settings, err := reader.repo.GetSettings(ctx)
	if err != nil {
		return auth.SecuritySettings{}, err
	}

	return auth.SecuritySettings{
		SessionDays:         settings.SessionDays,
		LoginFailureLock:    settings.LoginFailureLock,
		TurnstileEnabled:    settings.TurnstileEnabled,
		TurnstileSiteKey:    settings.TurnstileSiteKey,
		TurnstileSecretKey:  settings.TurnstileSecretKey,
		TurnstileRegister:   settings.TurnstileRegister,
		TurnstileLogin:      settings.TurnstileLogin,
		TurnstileSubmission: settings.TurnstileSubmission,
	}, nil
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
	if repos.OperationsRepo == nil {
		repos.OperationsRepo = operations.NewMemoryRepository()
	}
	if repos.UserRepo == nil {
		repos.UserRepo = users.NewMemoryRepository()
	}
	if repos.AdminPostRepo == nil {
		repos.AdminPostRepo = adminposts.NewMemoryRepository()
	}
	if repos.TaxonomyRepo == nil {
		repos.TaxonomyRepo = taxonomies.NewMemoryRepository()
	}
	if repos.TopicRepo == nil {
		repos.TopicRepo = topics.NewMemoryRepository()
	}
	if repos.AuthEmailSender == nil {
		emailSender, err := auth.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom, cfg.PublicURL)
		if err == nil && emailSender != nil {
			repos.AuthEmailSender = emailSender
		}
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors(cfg.WebOrigin))
	router.Use(csrfProtection(cfg.WebOrigin, cfg.PublicURL))
	router.Use(rateLimit(120, time.Minute))
	router.Use(authSensitiveRateLimit())
	router.Static("/uploads", uploadDir(cfg.UploadDir))

	router.Use(navigationRedirects(repos.OperationsRepo))
	router.Use(auth.Middleware(repos.AuthStore))
	router.Use(auditAdminOperations(repos.OperationsRepo))
	feeds.RegisterRoutes(router, repos.PostRepo, repos.OperationsRepo, cfg.PublicURL)
	seo.RegisterRoutes(router, repos.PostRepo, repos.OperationsRepo, cfg.PublicURL)

	api := router.Group("/api")

	api.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	auth.RegisterRoutesWithDependencies(api, repos.AuthStore, authSecuritySettingsReader{repo: repos.OperationsRepo}, repos.AuthEmailSender, repos.TurnstileVerifier)
	taxonomies.RegisterRoutes(api, repos.TaxonomyRepo)
	topics.RegisterRoutes(api, repos.TopicRepo, repos.PostRepo)
	posts.RegisterPublicRoutes(api, repos.PostRepo)
	comments.RegisterRoutes(api, repos.CommentRepo, repos.OperationsRepo)
	reactions.RegisterRoutes(api, repos.ReactionRepo, repos.PostRepo)
	messages.RegisterRoutes(api, repos.MessageRepo)
	operations.RegisterRoutes(api, repos.OperationsRepo, uploadDir(cfg.UploadDir))
	users.RegisterRoutesWithEmailSender(api, repos.UserRepo, repos.AuthStore, repos.AuthEmailSender)

	var publisher posts.Publisher
	if item, ok := repos.PostRepo.(posts.Publisher); ok {
		publisher = item
	}
	adminposts.RegisterRoutes(api, repos.AdminPostRepo, publisher)
	submissions.RegisterRoutesWithTurnstile(api, repos.SubmissionRepo, repos.MessageRepo, publisher, repos.OperationsRepo, repos.TurnstileVerifier)

	return router
}

func uploadDir(value string) string {
	if value == "" {
		return "uploads"
	}

	return value
}

func cors(origin string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Requested-With")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
