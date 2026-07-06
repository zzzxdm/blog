package operations

import (
	"context"
	"testing"
)

func TestUpdateSettingsNormalizesInput(t *testing.T) {
	repo := NewMemoryRepository()

	settings, err := repo.UpdateSettings(context.Background(), Settings{
		SiteName:            " ",
		SiteDescription:     "  内容站  ",
		SiteURL:             " ",
		ThemePrimary:        "295b4b",
		HomepageLayout:      "  专题 + 最新  ",
		BlockedWords:        []string{" 推广 ", "推广", "Spam", "spam", ""},
		SubmissionLimit:     " ",
		SubmissionGuide:     "  原创优先  ",
		MailProvider:        " ",
		FromEmail:           " ",
		TurnstileSiteKey:    "  site-key  ",
		TurnstileSecretKey:  "  secret-key  ",
		TurnstileRegister:   true,
		TurnstileLogin:      true,
		TurnstileSubmission: true,
		SessionDays:         -3,
		BackupCycle:         " ",
		BackupRetentionDays: 999,
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}

	if settings.SiteName != "云间笔记" {
		t.Fatalf("SiteName = %q, want default", settings.SiteName)
	}
	if settings.SiteDescription != "内容站" {
		t.Fatalf("SiteDescription = %q, want trimmed value", settings.SiteDescription)
	}
	if settings.SiteURL != "https://blog.example.com" {
		t.Fatalf("SiteURL = %q, want default", settings.SiteURL)
	}
	if settings.ThemePrimary != "#295b4b" {
		t.Fatalf("ThemePrimary = %q, want default valid color", settings.ThemePrimary)
	}
	if settings.HomepageLayout != "专题 + 最新" {
		t.Fatalf("HomepageLayout = %q, want trimmed value", settings.HomepageLayout)
	}
	if got := settings.BlockedWords; len(got) != 2 || got[0] != "推广" || got[1] != "Spam" {
		t.Fatalf("BlockedWords = %#v, want trimmed case-insensitive unique values", got)
	}
	if !settings.SubmissionManualReview {
		t.Fatal("SubmissionManualReview should stay enabled for the fixed review workflow")
	}
	if settings.SubmissionLimit != "每天最多 3 篇" {
		t.Fatalf("SubmissionLimit = %q, want default", settings.SubmissionLimit)
	}
	if settings.SubmissionGuide != "原创优先" {
		t.Fatalf("SubmissionGuide = %q, want trimmed value", settings.SubmissionGuide)
	}
	if settings.MailProvider != "Resend" || settings.FromEmail != "noreply@example.com" {
		t.Fatalf("mail settings = %q/%q, want defaults", settings.MailProvider, settings.FromEmail)
	}
	if settings.TurnstileSiteKey != "site-key" || settings.TurnstileSecretKey != "secret-key" || !settings.TurnstileRegister || !settings.TurnstileLogin || !settings.TurnstileSubmission {
		t.Fatalf("turnstile settings not normalized: %+v", settings)
	}
	if settings.SessionDays != 7 {
		t.Fatalf("SessionDays = %d, want default", settings.SessionDays)
	}
	if settings.BackupCycle != "每日全量备份" || settings.BackupRetentionDays != 365 {
		t.Fatalf("backup settings = %q/%d, want default cycle and max retention", settings.BackupCycle, settings.BackupRetentionDays)
	}
	if settings.UpdatedAt.IsZero() || settings.LastBackupAt.IsZero() {
		t.Fatalf("expected UpdatedAt and LastBackupAt to be populated, got %+v", settings)
	}
}

func TestNormalizeThemeColorKeepsValidHex(t *testing.T) {
	if got := normalizeThemeColor(" #A0B1C2 ", "#295b4b"); got != "#a0b1c2" {
		t.Fatalf("normalizeThemeColor returned %q, want #a0b1c2", got)
	}
}

func TestUpdateNavigationNormalizesInput(t *testing.T) {
	repo := NewMemoryRepository()

	navigation, err := repo.UpdateNavigation(context.Background(), Navigation{
		TopItems: []NavItem{
			{ID: " custom ", Label: " 首页 ", URL: " / "},
			{Label: " ", URL: "/empty"},
			{Label: "归档", URL: " /archive "},
		},
		FooterItems: []NavItem{
			{Label: "", URL: ""},
		},
		GitHubURL:    " https://github.com/example ",
		ContactEmail: " hello@example.com ",
		RSSURL:       " /rss.xml ",
		Redirects: []RedirectRule{
			{From: " /old ", To: " /new ", Code: 302},
			{From: "/same", To: "/same", Code: 301},
			{From: "/legacy", To: "/posts/new", Code: 308},
			{From: " ", To: "/missing", Code: 301},
		},
	})
	if err != nil {
		t.Fatalf("UpdateNavigation returned error: %v", err)
	}

	if len(navigation.TopItems) != 2 {
		t.Fatalf("TopItems = %#v, want two valid items", navigation.TopItems)
	}
	if navigation.TopItems[0].ID != "custom" || navigation.TopItems[0].Label != "首页" || navigation.TopItems[0].URL != "/" || navigation.TopItems[0].Order != 1 {
		t.Fatalf("first top item = %+v, want trimmed order 1", navigation.TopItems[0])
	}
	if navigation.TopItems[1].ID != "nav_top_2" || navigation.TopItems[1].Order != 2 {
		t.Fatalf("second top item = %+v, want generated id and order 2", navigation.TopItems[1])
	}
	if len(navigation.FooterItems) != 0 {
		t.Fatalf("FooterItems = %#v, want no default footer items when all provided items are invalid", navigation.FooterItems)
	}
	if navigation.GitHubURL != "https://github.com/example" || navigation.ContactEmail != "hello@example.com" || navigation.RSSURL != "/rss.xml" {
		t.Fatalf("link settings not trimmed: %+v", navigation)
	}
	if len(navigation.Redirects) != 2 {
		t.Fatalf("Redirects = %#v, want two valid rules", navigation.Redirects)
	}
	if navigation.Redirects[0].From != "/old" || navigation.Redirects[0].To != "/new" || navigation.Redirects[0].Code != 302 {
		t.Fatalf("first redirect = %+v, want trimmed 302 rule", navigation.Redirects[0])
	}
	if navigation.Redirects[1].From != "/legacy" || navigation.Redirects[1].Code != 308 {
		t.Fatalf("second redirect = %+v, want permanent redirect code 308", navigation.Redirects[1])
	}
	if navigation.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt should be populated")
	}
}
