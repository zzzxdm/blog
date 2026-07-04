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
	if settings.SubmissionLimit != "每天最多 3 篇" {
		t.Fatalf("SubmissionLimit = %q, want default", settings.SubmissionLimit)
	}
	if settings.SubmissionGuide != "原创优先" {
		t.Fatalf("SubmissionGuide = %q, want trimmed value", settings.SubmissionGuide)
	}
	if settings.MailProvider != "Resend" || settings.FromEmail != "noreply@example.com" {
		t.Fatalf("mail settings = %q/%q, want defaults", settings.MailProvider, settings.FromEmail)
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
