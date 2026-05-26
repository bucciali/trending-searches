package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	KafkaBrokers      []string      `mapstructure:"kafka_brokers"`
	KafkaTopic        string        `mapstructure:"kafka_topic"`
	ConsumerGroup     string        `mapstructure:"consumer_group"`
	RedisAddress      string        `mapstructure:"redis_addr"`
	RedisPassword     string        `mapstructure:"redis_password"`
	RedisDb           int           `mapstructure:"redis_db"`
	HttpPort          int           `mapstructure:"http_port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	AggregateInterval time.Duration `mapstructure:"aggregate_interval"`
	RateLimitPerUser  int           `mapstructure:"rate_limit_per_user"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("TRENDING")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetDefault("kafka_brokers", []string{"localhost:9092"})
	v.SetDefault("kafka_topic", "search-events")
	v.SetDefault("consumer_group", "trending-group")
	v.SetDefault("redis_addr", "localhost:6379")
	v.SetDefault("redis_password", "")
	v.SetDefault("redis_db", 0)
	v.SetDefault("http_port", 8080)
	v.SetDefault("read_timeout", 10*time.Second)
	v.SetDefault("write_timeout", 10*time.Second)
	v.SetDefault("aggregate_interval", 10*time.Second)
	v.SetDefault("rate_limit_per_user", 10)

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka_brokers is required")
	}
	if cfg.KafkaTopic == "" {
		return nil, fmt.Errorf("kafka_topic is required")
	}
	if cfg.RedisAddress == "" {
		return nil, fmt.Errorf("redis_addr is required")
	}
	return &cfg, nil
}
