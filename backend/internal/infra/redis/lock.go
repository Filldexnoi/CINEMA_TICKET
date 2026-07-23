package redis

import (
	"context"
	"time"

	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/redis/go-redis/v9"
)

var releaseIfMatchScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

type Lock struct {
	client *redis.Client
}

func NewLock(client *redis.Client) *Lock {
	return &Lock{client: client}
}

var _ ports.DistributedLock = (*Lock)(nil)

func (l *Lock) Acquire(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (l *Lock) ReleaseIfMatch(ctx context.Context, key, token string) (bool, error) {
	res, err := releaseIfMatchScript.Run(ctx, l.client, []string{key}, token).Int64()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (l *Lock) IsHeldByToken(ctx context.Context, key, token string) (bool, error) {
	val, err := l.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == token, nil
}
