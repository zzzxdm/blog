package auth

import (
	"strings"
	"testing"
)

func TestSMTPEmailSenderAuthLinkUsesPublicURLAndToken(t *testing.T) {
	sender, err := NewSMTPEmailSender("smtp.example.com", "587", "user", "pass", "noreply@example.com", "https://blog.example.com/")
	if err != nil {
		t.Fatalf("NewSMTPEmailSender returned error: %v", err)
	}

	link := sender.authLink("reset", "token with spaces")
	if link != "https://blog.example.com/login?mode=reset&token=token+with+spaces" {
		t.Fatalf("authLink = %q", link)
	}
}

func TestRenderEmailTemplateIncludesActionURLAndToken(t *testing.T) {
	html, err := renderEmailTemplate("password-setup.html", emailTemplateData{
		SiteName:        "Blog",
		DisplayName:     "Invite User",
		Email:           "invite@example.com",
		ActionURL:       "https://blog.example.com/login?mode=reset&token=abc",
		LoginURL:        "https://blog.example.com/login",
		Token:           "abc",
		InitialPassword: "TempPass123",
		ExpiresIn:       "30 minutes",
	})
	if err != nil {
		t.Fatalf("renderEmailTemplate returned error: %v", err)
	}
	if !strings.Contains(html, "https://blog.example.com/login?mode=reset&amp;token=abc") {
		t.Fatalf("expected rendered template to include escaped action URL, got %q", html)
	}
	if !strings.Contains(html, "https://blog.example.com/login") {
		t.Fatalf("expected rendered template to include login URL, got %q", html)
	}
	if !strings.Contains(html, "abc") {
		t.Fatalf("expected rendered template to include token, got %q", html)
	}
	if !strings.Contains(html, "TempPass123") {
		t.Fatalf("expected rendered template to include initial password, got %q", html)
	}
}

func TestRenderPasswordResetTemplateDoesNotRequireInitialPassword(t *testing.T) {
	html, err := renderEmailTemplate("password-reset.html", emailTemplateData{
		SiteName:    "Blog",
		DisplayName: "Reset User",
		Email:       "reset@example.com",
		ActionURL:   "https://blog.example.com/login?mode=reset&token=reset-token",
		Token:       "reset-token",
		ExpiresIn:   "30 minutes",
	})
	if err != nil {
		t.Fatalf("renderEmailTemplate returned error: %v", err)
	}
	if !strings.Contains(html, "reset-token") {
		t.Fatalf("expected rendered template to include reset token, got %q", html)
	}
	if strings.Contains(html, "TempPass") {
		t.Fatalf("did not expect password reset template to include invitation password copy, got %q", html)
	}
}
