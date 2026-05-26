package storageredis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

type WordCount struct {
	Word  string `json:"word"`
	Count int64  `json:"count"`
}

func (r *RedisClient) AddToStoplist(ctx context.Context, word string) error {
	return r.Client.SAdd(ctx, "stoplist", word).Err()
}

func (r *RedisClient) RemoveFromStoplist(ctx context.Context, word string) error {
	return r.Client.SRem(ctx, "stoplist", word).Err()
}

func (r *RedisClient) IsStoplisted(ctx context.Context, word string) (bool, error) {
	return r.Client.SIsMember(ctx, "stoplist", word).Result()
}

func (r *RedisClient) GetTop(ctx context.Context, limit int) ([]WordCount, error) {
	keys := make([]string, 0, 5)
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		minute := now.Add(-time.Duration(i) * time.Minute)
		keys = append(keys, fmt.Sprintf("stats:%s", minute.Format("2006-01-02T15:04")))
	}

	tmpKey := fmt.Sprintf("tmp_top:%d", now.UnixNano())
	err := r.Client.ZUnionStore(ctx, tmpKey, &redis.ZStore{
		Keys:      keys,
		Weights:   []float64{1, 1, 1, 1, 1},
		Aggregate: "SUM",
	}).Err()
	if err != nil {
		return nil, err
	}

	defer r.Client.Del(ctx, tmpKey)

	scores, err := r.Client.ZRevRangeWithScores(ctx, tmpKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	result := make([]WordCount, 0, len(scores))
	for _, z := range scores {
		result = append(result, WordCount{
			Word:  z.Member.(string),
			Count: int64(z.Score),
		})
	}

	return result, nil
}

func (r *RedisClient) IncrementUserRequests(ctx context.Context, userID string, window time.Duration, limit int) (bool, error) {
	key := fmt.Sprintf("rate:%s", userID)
	pipe := r.Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return incr.Val() <= int64(limit), nil
}

func (r *RedisClient) IncrementWord(ctx context.Context, word string) error {
	minuteKey := fmt.Sprintf("stats:%s", time.Now().UTC().Format("2006-01-02T15:04"))
	return r.Client.ZIncrBy(ctx, minuteKey, 1, word).Err()
}

func NewRedisClientWithRetry(addr string, maxRetries int, delay time.Duration) (*RedisClient, error) {
	for i := 0; i < maxRetries; i++ {
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := rdb.Ping(ctx).Err()
		cancel()
		if err == nil {
			log.Printf("Redis connected (attempt %d/%d)", i+1, maxRetries)
			return &RedisClient{Client: rdb}, nil
		}
		log.Printf("Redis not ready (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
}

func (r *RedisClient) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
