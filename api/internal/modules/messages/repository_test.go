package messages

import (
	"testing"
	"time"
)

func TestMessageStatsCountScheduledForAdminOnly(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	repo := &MemoryRepository{
		messages: []Message{
			{
				ID:          "message_now",
				RecipientID: "user_linyi",
				Type:        TypeAdmin,
				CreatedAt:   now,
			},
			{
				ID:          "message_scheduled",
				RecipientID: "user_linyi",
				Type:        TypeAdmin,
				ScheduledAt: &future,
				CreatedAt:   now,
			},
		},
	}

	userStats := repo.statsLocked("user_linyi", now)
	if userStats.Total != 1 || userStats.Scheduled != 0 {
		t.Fatalf("expected scheduled messages hidden from user stats, got %+v", userStats)
	}

	adminStats := repo.adminStatsLocked(now)
	if adminStats.Total != 2 || adminStats.Scheduled != 1 {
		t.Fatalf("expected scheduled messages counted for admin stats, got %+v", adminStats)
	}
}
