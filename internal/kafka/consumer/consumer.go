package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"trending-searches/internal/metrics"
	"trending-searches/internal/storageredis"

	"github.com/segmentio/kafka-go"
)

type SearchEvent struct {
	Query     string    `json:"query"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Consumer struct {
	reader      *kafka.Reader
	redisClient *storageredis.RedisClient
}

func NewConsumer(brokers []string, topic string, groupID string, redisClient *storageredis.RedisClient) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	return &Consumer{
		reader:      reader,
		redisClient: redisClient,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Kafka consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer stopping...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Kafka read error: %v", err)
				continue
			}

			var event SearchEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("JSON parse error: %v", err)
				continue
			}

			log.Printf("Received: query=%s, user=%s", event.Query, event.UserID)

			allowed, err := c.redisClient.IncrementUserRequests(ctx, event.UserID, time.Minute, 10)
			if err != nil {
				log.Printf("Rate limit error: %v", err)
				continue
			}
			if !allowed {
				log.Printf("Rate limit exceeded for user %s, skipping", event.UserID)
				continue
			}

			if ok, err := c.redisClient.IsStoplisted(ctx, event.Query); err == nil && ok {
				log.Printf("Stoplisted word '%s' skipped", event.Query)
				continue
			}

			if err := c.redisClient.IncrementWord(ctx, event.Query); err != nil {
				log.Printf("Failed to increment word: %v", err)
			} else {
				metrics.EventsReceived.WithLabelValues(event.Query).Inc()
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
