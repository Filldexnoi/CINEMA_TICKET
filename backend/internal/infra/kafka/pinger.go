package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Pinger struct {
	brokers []string
}

func NewPinger(brokers []string) *Pinger {
	return &Pinger{brokers: brokers}
}

func (p *Pinger) Ping(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}
