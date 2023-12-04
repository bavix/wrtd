package scheduler

import (
	"context"
	"time"
)

type Task func(ctx context.Context, td time.Time)

func RunTask(
	ctx context.Context,
	dur time.Duration,
	callback Task,
) {
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			callback(context.WithoutCancel(ctx), now)
		}
	}
}
