package ports

import (
	"context"
	"time"
)

type DistributedLock interface {
	Acquire(ctx context.Context, key, token string, ttl time.Duration) (ok bool, err error)

	ReleaseIfMatch(ctx context.Context, key, token string) (released bool, err error)

	IsHeldByToken(ctx context.Context, key, token string) (bool, error)
}
