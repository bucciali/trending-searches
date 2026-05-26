package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"trending-searches/internal/api"
	"trending-searches/internal/config"
	"trending-searches/internal/kafka/consumer"
	"trending-searches/internal/storageredis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	redisClient, err := storageredis.NewRedisClientWithRetry(cfg.RedisAddress, 10, 3*time.Second)
	if err != nil {
		log.Fatalf("redis error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected")

	log.Printf("Kafka config: brokers=%v, topic=%s, group=%s", cfg.KafkaBrokers, cfg.KafkaTopic, cfg.ConsumerGroup)
	kafkaConsumer := consumer.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.ConsumerGroup, redisClient)
	go kafkaConsumer.Start(context.Background())
	defer kafkaConsumer.Close()
	log.Println("Kafka consumer started")

	router := api.NewRouter(redisClient)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
		Handler: router,
	}
	go func() {
		log.Printf("REST server listening on port %d", cfg.HttpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
