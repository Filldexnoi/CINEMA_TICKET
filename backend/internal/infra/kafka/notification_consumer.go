package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"cinema-ticket/backend/internal/domain"

	"github.com/segmentio/kafka-go"
)

func RunNotificationConsumer(ctx context.Context, brokers []string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       SeatEventsTopic,
		GroupID:     ConsumerGroupNotifications,
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("notification consumer: read error: %v", err)
			continue
		}

		var event domain.SeatEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("notification consumer: bad message: %v", err)
			continue
		}
		if event.EventType != domain.EventBookingConfirmed {
			continue
		}

		log.Printf(
			"notification (mock): booking confirmed - user=%s booking=%s showtime=%s seats=%v",
			event.UserID, event.BookingID, event.ShowtimeID, event.SeatLabels,
		)
	}
}
