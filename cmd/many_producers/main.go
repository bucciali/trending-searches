package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type SearchEvent struct {
	Query     string    `json:"query"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	log.Println("Starting many_producers...")

	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:29092"),
		Topic:    "search-events",
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}
	defer writer.Close()

	queries := []string{"айфон", "пылесос", "кроссовки", "headphones", "ноутбук", "стиралка", "холодильник", "чайник", "телевизор", "rolls"}

	var wg sync.WaitGroup
	workers := 10

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				userID := fmt.Sprintf("user_%d", rand.Intn(1000))
				query := queries[rand.Intn(len(queries))]

				event := SearchEvent{
					Query:     query,
					UserID:    userID,
					Timestamp: time.Now(),
				}

				data, _ := json.Marshal(event)
				err := writer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(event.UserID),
					Value: data,
				})
				if err != nil {
					log.Printf("Worker %d error: %v", workerID, err)
				} else {
					log.Printf("Worker %d sent: %s from %s", workerID, query, userID)
				}

				time.Sleep(10 * time.Millisecond)
			}
		}(w)
	}

	wg.Wait()
}
