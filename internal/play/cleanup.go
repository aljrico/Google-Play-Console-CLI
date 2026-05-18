package play

import (
	"context"
	"time"
)

const cleanupTimeout = 30 * time.Second

func newCleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(context.Background()), cleanupTimeout)
}
