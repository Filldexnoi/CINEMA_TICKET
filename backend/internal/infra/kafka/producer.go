package kafka

import (
	"context"
	"encoding/json"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: SeatEventsTopic,

			Balancer: &kafka.Hash{},
		},
	}
}

var _ ports.EventPublisher = (*Producer)(nil)

func (p *Producer) Publish(ctx context.Context, event domain.SeatEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.ShowtimeID),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
