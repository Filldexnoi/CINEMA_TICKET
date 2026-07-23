package redis

import (
	"context"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

type ExpiryHandler func(showtimeID, seatLabel string)

func ListenForExpiry(ctx context.Context, client *redis.Client, handler ExpiryHandler) {
	sub := client.PSubscribe(ctx, "__keyevent@0__:expired")
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			showtimeID, seatLabel, ok := parseSeatLockKey(msg.Payload)
			if !ok {
				continue
			}
			log.Printf("redis: lock expired for showtime=%s seat=%s", showtimeID, seatLabel)
			handler(showtimeID, seatLabel)
		}
	}
}

func parseSeatLockKey(key string) (showtimeID, seatLabel string, ok bool) {
	const prefix = "lock:showtime:"
	const seatMarker = ":seat:"
	if !strings.HasPrefix(key, prefix) {
		return "", "", false
	}
	rest := key[len(prefix):]
	idx := strings.Index(rest, seatMarker)
	if idx == -1 {
		return "", "", false
	}
	return rest[:idx], rest[idx+len(seatMarker):], true
}
