package server

import (
	"context"
	"net/http"
	"strings"
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

	secureCookies := cookieSecure(cfg.AppEnv, cfg.PublicURL, cfg.WebOrigin)
	auth.ConfigureCookieSecurity(secureCookies)
	auth.ConfigureDevAuthTokens(!strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "production"))
	requireRepositories(repos)
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
	router.Use(csrfProtection(cfg.WebOrigin, cfg.PublicURL, secureCookies))
	router.Use(rateLimit(120, time.Minute))
	router.Use(authSensitiveRateLimit())
	router.Static("/uploads", uploadDir(cfg.UploadDir))

	router.Use(navigationRedirects(repos.OperationsRepo))
	router.Use(auth.Middleware(repos.AuthStore))
	router.Use(auditAdminOperations(repos.OperationsRepo))
	mediaStorage := newMediaStorage(cfg)
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
	operations.RegisterRoutes(api, repos.OperationsRepo, mediaStorage)
	users.RegisterRoutesWithEmailSender(api, repos.UserRepo, repos.AuthStore, repos.AuthEmailSender)

	var publisher posts.SubmissionPublisher
	if item, ok := repos.PostRepo.(posts.SubmissionPublisher); ok {
		publisher = item
	}
	var adminPublisher posts.AdminPublisher
	if item, ok := repos.PostRepo.(posts.AdminPublisher); ok {
		adminPublisher = item
	}
	var postArchiver posts.Archiver
	if item, ok := repos.PostRepo.(posts.Archiver); ok {
		postArchiver = item
	}
	var postRestorer posts.Restorer
	if item, ok := repos.PostRepo.(posts.Restorer); ok {
		postRestorer = item
	}
	adminposts.RegisterRoutes(api, repos.AdminPostRepo, adminPublisher, postArchiver)
	submissions.RegisterRoutesWithTurnstile(api, repos.SubmissionRepo, repos.MessageRepo, publisher, postArchiver, postRestorer, repos.OperationsRepo, repos.TurnstileVerifier)

	return router
}

func requireRepositories(repos Repositories) {
	missing := make([]string, 0)
	if repos.AuthStore == nil {
		missing = append(missing, "AuthStore")
	}
	if repos.PostRepo == nil {
		missing = append(missing, "PostRepo")
	}
	if repos.CommentRepo == nil {
		missing = append(missing, "CommentRepo")
	}
	if repos.ReactionRepo == nil {
		missing = append(missing, "ReactionRepo")
	}
	if repos.MessageRepo == nil {
		missing = append(missing, "MessageRepo")
	}
	if repos.SubmissionRepo == nil {
		missing = append(missing, "SubmissionRepo")
	}
	if repos.OperationsRepo == nil {
		missing = append(missing, "OperationsRepo")
	}
	if repos.UserRepo == nil {
		missing = append(missing, "UserRepo")
	}
	if repos.AdminPostRepo == nil {
		missing = append(missing, "AdminPostRepo")
	}
	if repos.TaxonomyRepo == nil {
		missing = append(missing, "TaxonomyRepo")
	}
	if repos.TopicRepo == nil {
		missing = append(missing, "TopicRepo")
	}
	if len(missing) > 0 {
		panic("server repositories are required: " + strings.Join(missing, ", "))
	}
}

func uploadDir(value string) string {
	if value == "" {
		return "uploads"
	}

	return value
}

func newMediaStorage(cfg config.Config) operations.MediaStorage {
	if strings.EqualFold(cfg.MediaStorage, "minio") {
		storage, err := operations.NewMinIOMediaStorage(operations.MinIOStorageConfig{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Bucket:    cfg.MinIOBucket,
			UseSSL:    cfg.MinIOUseSSL,
			PublicURL: cfg.MinIOPublicURL,
		})
		if err != nil {
			panic("failed to configure minio media storage: " + err.Error())
		}

		return storage
	}

	return operations.NewLocalMediaStorage(uploadDir(cfg.UploadDir))
}

func cookieSecure(appEnv string, urls ...string) bool {
	if strings.EqualFold(strings.TrimSpace(appEnv), "production") {
		return true
	}

	for _, value := range urls {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "https://") {
			return true
		}
	}

	return false
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


