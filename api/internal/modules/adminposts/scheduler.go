package adminposts

import (
	"context"
	"log/slog"
	"time"

	"blog/api/internal/modules/posts"
)

func StartScheduledPublisher(ctx context.Context, repo Repository, publisher posts.AdminPublisher, interval time.Duration) {
	if repo == nil || publisher == nil {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}

	go func() {
		publishDue(ctx, repo, publisher)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				publishDue(ctx, repo, publisher)
			}
		}
	}()
}

func publishDue(ctx context.Context, repo Repository, publisher posts.AdminPublisher) {
	publishCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	count, err := repo.PublishDue(publishCtx, publisher, time.Now())
	if err != nil {
		slog.Warn("scheduled admin posts publish failed", "error", err)
	}
	if count > 0 {
		slog.Info("scheduled admin posts published", "count", count)
	}
}
