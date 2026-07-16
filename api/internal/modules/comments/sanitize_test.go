package comments

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeCommentBodyStripsHTMLAndKeepsMarkdown(t *testing.T) {
	got, err := SanitizeCommentBody("  你好 <script>alert(1)</script> **粗体**\n\n第二段  ")
	if err != nil {
		t.Fatalf("SanitizeCommentBody returned error: %v", err)
	}
	if strings.Contains(got, "<script>") || strings.Contains(got, "</script>") {
		t.Fatalf("expected HTML tags stripped, got %q", got)
	}
	if !strings.Contains(got, "**粗体**") {
		t.Fatalf("expected markdown to remain, got %q", got)
	}
	if !strings.Contains(got, "第二段") {
		t.Fatalf("expected paragraph text to remain, got %q", got)
	}
}

func TestSanitizeCommentBodyRejectsEmptyAndTooLong(t *testing.T) {
	if _, err := SanitizeCommentBody("   <b></b>  "); !errors.Is(err, ErrEmptyBody) {
		t.Fatalf("expected ErrEmptyBody, got %v", err)
	}

	long := strings.Repeat("啊", maxCommentBodyRunes+1)
	if _, err := SanitizeCommentBody(long); !errors.Is(err, ErrBodyTooLong) {
		t.Fatalf("expected ErrBodyTooLong, got %v", err)
	}
}

func TestSanitizeCommentBodyRemovesControlChars(t *testing.T) {
	got, err := SanitizeCommentBody("hello\x00world\x07!")
	if err != nil {
		t.Fatalf("SanitizeCommentBody returned error: %v", err)
	}
	if got != "helloworld!" {
		t.Fatalf("expected control chars removed, got %q", got)
	}
}
