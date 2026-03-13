package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type UserDeletedEvent struct {
	UserID int    `json:"user_id"`
	Action string `json:"action"`
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "user-events",
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) SendUserDeleted(ctx context.Context, userID int) error {
	event := UserDeletedEvent{
		UserID: userID,
		Action: "deleted",
	}

	bytes, _ := json.Marshal(event)

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprint(userID)),
		Value: bytes,
	})
	if err != nil {
		slog.Error("Failed to send user deleted event", "userID", userID, "error", err)
		return err
	}

	return nil
}
