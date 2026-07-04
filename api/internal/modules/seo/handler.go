package seo

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	postRepo  posts.Repository
	publicURL string
	template  *template.Template
}

const siteName = "云间笔记"

func NewHandler(postRepo posts.Repository, publicURL string) *Handler {
	return &Handler{
		postRepo:  postRepo,
		publicURL: strings.TrimRight(defaultString(publicURL, "http://localhost:5173"), "/"),
		template:  template.Must(template.New("article").Parse(articleTemplate)),
	}
}

func RegisterRoutes(router gin.IRouter, postRepo posts.Repository, publicURL string) {
	handler := NewHandler(postRepo, publicURL)

	router.GET("/posts/:slug", handler.Article)
}

func (handler *Handler) Article(ctx *gin.Context) {
	post, err := handler.postRepo.GetBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		if errors.Is(err, posts.ErrNotFound) {
			ctx.String(http.StatusNotFound, "post not found")
			return
		}

		ctx.String(http.StatusInternalServerError, "failed to load post")
		return
	}

	canonical := handler.publicURL + "/posts/" + post.Slug
	description := defaultString(post.Summary, post.Title)
	jsonLD, err := json.Marshal(articleStructuredData{
		Context:       "https://schema.org",
		Type:          "BlogPosting",
		Headline:      post.Title,
		Description:   description,
		Image:         post.CoverImage,
		URL:           canonical,
		DatePublished: post.PublishedAt.Format("2006-01-02"),
		Author: personStructuredData{
			Type: "Person",
			Name: post.AuthorName,
		},
		Publisher: organizationStructuredData{
			Type: "Organization",
			Name: siteName,
		},
		Keywords: strings.Join(post.Tags, ","),
	})
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to render post metadata")
		return
	}

	var body bytes.Buffer
	if err := handler.template.Execute(&body, articlePageData{
		SiteName:    siteName,
		Title:       post.Title,
		Description: description,
		Canonical:   canonical,
		Image:       post.CoverImage,
		Category:    post.Category,
		AuthorName:  post.AuthorName,
		PublishedAt: post.PublishedAt.Format("2006-01-02"),
		ReadingTime: post.ReadingTime,
		Content:     post.Content,
		JSONLD:      template.JS(jsonLD),
	}); err != nil {
		ctx.String(http.StatusInternalServerError, "failed to render post")
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", body.Bytes())
}

type articlePageData struct {
	SiteName    string
	Title       string
	Description string
	Canonical   string
	Image       string
	Category    string
	AuthorName  string
	PublishedAt string
	ReadingTime int
	Content     string
	JSONLD      template.JS
}

type articleStructuredData struct {
	Context       string                     `json:"@context"`
	Type          string                     `json:"@type"`
	Headline      string                     `json:"headline"`
	Description   string                     `json:"description"`
	Image         string                     `json:"image,omitempty"`
	URL           string                     `json:"url"`
	DatePublished string                     `json:"datePublished"`
	Author        personStructuredData       `json:"author"`
	Publisher     organizationStructuredData `json:"publisher"`
	Keywords      string                     `json:"keywords,omitempty"`
}

type personStructuredData struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type organizationStructuredData struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

const articleTemplate = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }} - {{ .SiteName }}</title>
    <meta name="description" content="{{ .Description }}">
    <link rel="canonical" href="{{ .Canonical }}">
    <meta property="og:type" content="article">
    <meta property="og:site_name" content="{{ .SiteName }}">
    <meta property="og:title" content="{{ .Title }}">
    <meta property="og:description" content="{{ .Description }}">
    <meta property="og:url" content="{{ .Canonical }}">
    <meta property="article:published_time" content="{{ .PublishedAt }}">
    <meta property="article:section" content="{{ .Category }}">
    {{ if .Image }}
    <meta property="og:image" content="{{ .Image }}">
    {{ end }}
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="{{ .Title }}">
    <meta name="twitter:description" content="{{ .Description }}">
    {{ if .Image }}
    <meta name="twitter:image" content="{{ .Image }}">
    {{ end }}
    <script type="application/ld+json">{{ .JSONLD }}</script>
    <link rel="stylesheet" href="/assets/index.css">
  </head>
  <body>
    <div id="app">
      <main class="article-shell">
        <article>
          <header class="article-hero">
            <div class="meta-row">
              <span class="tag">{{ .Category }}</span>
              <span>{{ .ReadingTime }} 分钟阅读</span>
              <span>{{ .PublishedAt }}</span>
            </div>
            <h1>{{ .Title }}</h1>
            <p class="dek">{{ .Description }}</p>
            <div class="author-row">
              <span class="avatar">管</span>
              <div>
                <strong>{{ .AuthorName }}</strong>
              </div>
            </div>
          </header>
          {{ if .Image }}
            <figure class="article-cover">
              <img src="{{ .Image }}" alt="{{ .Title }}">
            </figure>
          {{ end }}
          <section class="article-body">
            <p>{{ .Content }}</p>
          </section>
        </article>
      </main>
    </div>
    <script type="module" src="/assets/index.js"></script>
  </body>
</html>
`

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
