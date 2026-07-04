package feeds

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"blog/api/internal/modules/posts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	postRepo  posts.Repository
	publicURL string
}

func NewHandler(postRepo posts.Repository, publicURL string) *Handler {
	return &Handler{
		postRepo:  postRepo,
		publicURL: strings.TrimRight(defaultString(publicURL, "http://localhost:5173"), "/"),
	}
}

func RegisterRoutes(router gin.IRouter, postRepo posts.Repository, publicURL string) {
	handler := NewHandler(postRepo, publicURL)

	router.GET("/rss.xml", handler.RSS)
	router.GET("/feed.xml", handler.RSS)
	router.GET("/sitemap.xml", handler.Sitemap)
	router.GET("/robots.txt", handler.Robots)
}

func (handler *Handler) RSS(ctx *gin.Context) {
	result, err := handler.postRepo.List(ctx.Request.Context(), posts.ListQuery{Page: 1, PageSize: 50})
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to load rss")
		return
	}

	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       "云间笔记",
			Link:        handler.publicURL + "/",
			Description: "技术、产品、工程实践和长期写作的沉淀。",
			Language:    "zh-CN",
			LastBuild:   time.Now().Format(time.RFC1123Z),
			AtomLink: atomLink{
				Href: handler.publicURL + "/rss.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: make([]rssItem, 0, len(result.Items)),
		},
	}

	for _, post := range result.Items {
		link := handler.publicURL + "/posts/" + post.Slug
		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       post.Title,
			Link:        link,
			GUID:        link,
			Description: post.Summary,
			PubDate:     post.PublishedAt.Format(time.RFC1123Z),
		})
	}

	ctx.Header("Content-Type", "application/rss+xml; charset=utf-8")
	ctx.XML(http.StatusOK, feed)
}

func (handler *Handler) Sitemap(ctx *gin.Context) {
	result, err := handler.postRepo.List(ctx.Request.Context(), posts.ListQuery{Page: 1, PageSize: 50})
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to load sitemap")
		return
	}

	urls := []siteURL{
		{Loc: handler.publicURL + "/", ChangeFreq: "daily", Priority: "1.0"},
		{Loc: handler.publicURL + "/archive", ChangeFreq: "daily", Priority: "0.8"},
	}

	for _, post := range result.Items {
		urls = append(urls, siteURL{
			Loc:        handler.publicURL + "/posts/" + post.Slug,
			LastMod:    post.PublishedAt.Format(time.DateOnly),
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	ctx.Header("Content-Type", "application/xml; charset=utf-8")
	ctx.XML(http.StatusOK, urlSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	})
}

func (handler *Handler) Robots(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", handler.publicURL)
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	AtomLink    atomLink  `xml:"atom:link"`
	Items       []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type urlSet struct {
	XMLName xml.Name  `xml:"urlset"`
	XMLNS   string    `xml:"xmlns,attr"`
	URLs    []siteURL `xml:"url"`
}

type siteURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
