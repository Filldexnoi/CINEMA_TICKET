package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/segmentio/kafka-go"
)

type WSPayload struct {
	Type       domain.EventType `json:"type"`
	ShowtimeID string           `json:"showtime_id"`
	SeatLabels []string         `json:"seat_labels"`
}

func RunConsumer(ctx context.Context, brokers []string, broadcaster ports.Broadcaster) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       SeatEventsTopic,
		Partition:   0,
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("kafka consumer: read error: %v", err)
			continue
		}

		var event domain.SeatEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("kafka consumer: bad message: %v", err)
			continue
		}

		payload, err := json.Marshal(WSPayload{
			Type:       event.EventType,
			ShowtimeID: event.ShowtimeID,
			SeatLabels: event.SeatLabels,
		})
		if err != nil {
			log.Printf("kafka consumer: marshal ws payload: %v", err)
			continue
		}

		broadcaster.BroadcastToShowtime(event.ShowtimeID, payload)
	}
}
