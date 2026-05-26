package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type SearchEvent struct {
	Query     string    `json:"query"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: w}
}

func (p *Producer) Send(ctx context.Context, event SearchEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func main() {
	producer := NewProducer([]string{"localhost:29092"}, "search-events")
	defer producer.Close()

	queries := []string{"айфон", "пылесос", "кроссовки", "наушники", "ноутбук", "спам"}

	for i := 0; i < 100; i++ {
		event := SearchEvent{
			Query:     queries[i%len(queries)],
			UserID:    "user_123",
			Timestamp: time.Now(),
		}

		err := producer.Send(context.Background(), event)
		if err != nil {
			log.Printf("Failed to send: %v", err)
		} else {
			log.Printf("Sent: %s", event.Query)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
