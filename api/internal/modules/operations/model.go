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

type PublicSettings struct {
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
	SubmissionsEnabled      bool      `json:"submissionsEnabled"`
	SubmissionLimit         string    `json:"submissionLimit"`
	SubmissionGuide         string    `json:"submissionGuide"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type TestMailResult struct {
	OK        bool      `json:"ok"`
	Provider  string    `json:"provider"`
	FromEmail string    `json:"fromEmail"`
	Delivery  string    `json:"delivery"`
	Message   string    `json:"message"`
	TestedAt  time.Time `json:"testedAt"`
}

type BackupResult struct {
	OK        bool      `json:"ok"`
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	FileName  string    `json:"fileName"`
	SizeLabel string    `json:"sizeLabel"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	Settings  Settings  `json:"settings"`
}

type AdminJob struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Scope       string         `json:"scope"`
	Status      string         `json:"status"`
	Progress    int            `json:"progress"`
	Message     string         `json:"message"`
	FileName    string         `json:"fileName,omitempty"`
	DownloadURL string         `json:"downloadUrl,omitempty"`
	Result      map[string]any `json:"result,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type AdminJobRequest struct {
	Scope    string `json:"scope"`
	FileName string `json:"fileName"`
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
	Title          string `json:"title"`
	Views          string `json:"views"`
	Bookmarks      int    `json:"bookmarks"`
	Comments       int    `json:"comments"`
	EngagementRate string `json:"engagementRate"`
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
	Range       string              `json:"range"`
	RangeLabel  string              `json:"rangeLabel"`
	Metrics     []Metric            `json:"metrics"`
	Trend       []BarPoint          `json:"trend"`
	TopPosts    []TopPost           `json:"topPosts"`
	Sources     []BarPoint          `json:"sources"`
	SearchTerms []SearchTerm        `json:"searchTerms"`
	Suggestions []ContentSuggestion `json:"suggestions"`
}

type AuditLog struct {
	ID            string    `json:"id"`
	ActorID       string    `json:"actorId"`
	ActorName     string    `json:"actorName"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resourceType"`
	ResourceID    string    `json:"resourceId"`
	ResourceTitle string    `json:"resourceTitle"`
	Status        string    `json:"status"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"userAgent"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AuditLogQuery struct {
	Action       string
	ResourceType string
	Page         int
	PageSize     int
}

type AuditLogListResult struct {
	Items    []AuditLog `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int        `json:"total"`
}
