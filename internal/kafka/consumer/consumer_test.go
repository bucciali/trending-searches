package consumer

import (
	"context"
	"testing"
	"time"

	"trending-searches/internal/storageredis"

	"github.com/redis/go-redis/v9"
)

func TestConsumerWithRealRedis(t *testing.T) {
	// Подключаемся к реальному Redis (должен быть запущен)
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	redisClient := &storageredis.RedisClient{
		Client: client,
	}

	ctx := context.Background()

	// Очищаем тестовые данные
	client.Del(ctx, "stoplist")
	client.Del(ctx, "stats:*")

	// Тест: добавление слова в Redis через consumer логику
	word := "телефон"

	// Добавляем слово несколько раз, чтобы гарантировать попадание в топ
	for i := 0; i < 5; i++ {
		err := redisClient.IncrementWord(ctx, word)
		if err != nil {
			t.Fatalf("IncrementWord failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Проверяем, что слово появилось в топе
	top, err := redisClient.GetTop(ctx, 10)
	if err != nil {
		t.Fatalf("GetTop failed: %v", err)
	}

	found := false
	for _, w := range top {
		if w.Word == word && w.Count >= 1 {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("word %s not found in top or count wrong", word)
	}
}
