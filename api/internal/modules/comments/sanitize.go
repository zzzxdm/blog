package comments

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxCommentBodyRunes = 4000

var (
	ErrBodyTooLong = errors.New("comment body is too long")
	htmlTagPattern = regexp.MustCompile(`(?is)<[^>]*>`)
)

// SanitizeCommentBody normalizes user comment markdown before storage.
// It strips raw HTML tags and control characters so later client rendering
// cannot rely on untrusted markup that slipped past the markdown renderer.
func SanitizeCommentBody(body string) (string, error) {
	cleaned := stripControlChars(body)
	cleaned = htmlTagPattern.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, "\u0000", "")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "", ErrEmptyBody
	}
	if utf8.RuneCountInString(cleaned) > maxCommentBodyRunes {
		return "", ErrBodyTooLong
	}
	return cleaned, nil
}

func stripControlChars(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
}
