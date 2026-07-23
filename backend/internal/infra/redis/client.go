package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func Connect(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

type Pinger struct {
	client *redis.Client
}

func NewPinger(client *redis.Client) *Pinger {
	return &Pinger{client: client}
}

func (p *Pinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
