package operations

import (
	"testing"
	"time"
)

func TestNewStatsRangeSpec(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		rangeKey  string
		wantKey   string
		wantLabel string
		wantStart time.Time
	}{
		{
			name:      "seven days",
			rangeKey:  "7d",
			wantKey:   "7d",
			wantLabel: "最近 7 天",
			wantStart: now.AddDate(0, 0, -7),
		},
		{
			name:      "year to date",
			rangeKey:  "ytd",
			wantKey:   "ytd",
			wantLabel: "今年",
			wantStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "default",
			rangeKey:  "",
			wantKey:   "30d",
			wantLabel: "最近 30 天",
			wantStart: now.AddDate(0, 0, -30),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			spec := newStatsRangeSpec(tt.rangeKey, now)
			if spec.Key != tt.wantKey || spec.Label != tt.wantLabel || !spec.Start.Equal(tt.wantStart) || !spec.End.Equal(now) {
				t.Fatalf("unexpected spec: %+v", spec)
			}
		})
	}
}

func TestStatsFormattingHelpers(t *testing.T) {
	if got := formatStatNumber(1234567); got != "1,234,567" {
		t.Fatalf("formatStatNumber() = %q", got)
	}
	if got := formatReadingMinutes(4.25); got != "4.2 分钟" {
		t.Fatalf("formatReadingMinutes() = %q", got)
	}
	if got := formatRate(48, 12420); got != "0.4%" {
		t.Fatalf("formatRate() = %q", got)
	}
	if got := barPercent(3, 10); got != 30 {
		t.Fatalf("barPercent() = %d", got)
	}
}
