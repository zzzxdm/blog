package operations

import "time"

type Settings struct {
	SiteName                string    `json:"siteName"`
	SiteDescription         string    `json:"siteDescription"`
	SiteURL                 string    `json:"siteUrl"`
	Beian                   string    `json:"beian"`
	ThemePrimary            string    `json:"themePrimary"`
	HomepageLayout          string    `json:"homepageLayout"`
	DarkModeEnabled         bool      `json:"darkModeEnabled"`
	ReadingProgressEnabled  bool      `json:"readingProgressEnabled"`
	CommentsEnabled         bool      `json:"commentsEnabled"`
	LoginRequiredForComment bool      `json:"loginRequiredForComment"`
	AutoApproveComments     bool      `json:"autoApproveComments"`
	BlockedWords            []string  `json:"blockedWords"`
	SubmissionsEnabled      bool      `json:"submissionsEnabled"`
	SubmissionManualReview  bool      `json:"submissionManualReview"`
	SubmissionLimit         string    `json:"submissionLimit"`
	SubmissionGuide         string    `json:"submissionGuide"`
	MailEnabled             bool      `json:"mailEnabled"`
	MailProvider            string    `json:"mailProvider"`
	FromEmail               string    `json:"fromEmail"`
	AdminTwoFactorRequired  bool      `json:"adminTwoFactorRequired"`
	LoginFailureLock        bool      `json:"loginFailureLock"`
	SessionDays             int       `json:"sessionDays"`
	BackupCycle             string    `json:"backupCycle"`
	LastBackupAt            time.Time `json:"lastBackupAt"`
	BackupRetentionDays     int       `json:"backupRetentionDays"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type NavItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Order int    `json:"order"`
}

type RedirectRule struct {
	From string `json:"from"`
	To   string `json:"to"`
	Code int    `json:"code"`
}

type Navigation struct {
	TopItems               []NavItem      `json:"topItems"`
	FooterItems            []NavItem      `json:"footerItems"`
	MobileCollapse         bool           `json:"mobileCollapse"`
	ExternalLinksNewWindow bool           `json:"externalLinksNewWindow"`
	ShowLoginEntry         bool           `json:"showLoginEntry"`
	GitHubURL              string         `json:"githubUrl"`
	ContactEmail           string         `json:"contactEmail"`
	RSSURL                 string         `json:"rssUrl"`
	Redirects              []RedirectRule `json:"redirects"`
	UpdatedAt              time.Time      `json:"updatedAt"`
}

type MediaAsset struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	URL        string    `json:"url"`
	Alt        string    `json:"alt"`
	Type       string    `json:"type"`
	Category   string    `json:"category"`
	SizeLabel  string    `json:"sizeLabel"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	UsageCount int       `json:"usageCount"`
	UploadedBy string    `json:"uploadedBy"`
	UploadedAt time.Time `json:"uploadedAt"`
}

type MediaListResult struct {
	Items []MediaAsset `json:"items"`
	Total int          `json:"total"`
}

type MediaUpdateRequest struct {
	Alt      string `json:"alt"`
	Category string `json:"category"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
}

type BarPoint struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Percent int    `json:"percent"`
	Tone    string `json:"tone"`
}

type TopPost struct {
	Title     string `json:"title"`
	Views     string `json:"views"`
	Bookmarks int    `json:"bookmarks"`
	Comments  int    `json:"comments"`
	RSSRate   string `json:"rssRate"`
}

type SearchTerm struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

type ContentSuggestion struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Stats struct {
	Metrics     []Metric            `json:"metrics"`
	Trend       []BarPoint          `json:"trend"`
	TopPosts    []TopPost           `json:"topPosts"`
	Sources     []BarPoint          `json:"sources"`
	SearchTerms []SearchTerm        `json:"searchTerms"`
	Suggestions []ContentSuggestion `json:"suggestions"`
}
