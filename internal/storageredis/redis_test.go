package storageredis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestEmptyTop(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisClient{Client: client}
	ctx := context.Background()

	top, err := rc.GetTop(ctx, 10)
	if err != nil {
		t.Fatalf("GetTop failed: %v", err)
	}
	if len(top) != 0 {
		t.Fatalf("expected empty top, got %d items", len(top))
	}
}

func TestSlidingWindow(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisClient{Client: client}
	ctx := context.Background()

	// Добавляем слово в "старую" минуту (6 минут назад)
	oldMinuteKey := fmt.Sprintf("stats:%s", time.Now().Add(-6*time.Minute).UTC().Format("2006-01-02T15:04"))
	err := client.ZIncrBy(ctx, oldMinuteKey, 10, "old_word").Err()
	if err != nil {
		t.Fatalf("failed to add old word: %v", err)
	}

	// Добавляем слово в текущую минуту
	err = rc.IncrementWord(ctx, "new_word")
	if err != nil {
		t.Fatalf("IncrementWord failed: %v", err)
	}

	// Получаем топ
	top, err := rc.GetTop(ctx, 10)
	if err != nil {
		t.Fatalf("GetTop failed: %v", err)
	}

	// old_word не должен быть в топе (старше 5 минут)
	for _, w := range top {
		if w.Word == "old_word" {
			t.Fatalf("old_word should not be in top (older than 5 minutes)")
		}
	}

	// new_word должен быть в топе
	found := false
	for _, w := range top {
		if w.Word == "new_word" && w.Count == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("new_word not found in top")
	}
}

func TestRateLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisClient{Client: client}
	ctx := context.Background()

	userID := "user_123"
	limit := 5

	// Отправляем 10 запросов
	for i := 1; i <= 10; i++ {
		allowed, err := rc.IncrementUserRequests(ctx, userID, time.Minute, limit)
		if err != nil {
			t.Fatalf("IncrementUserRequests failed: %v", err)
		}

		// Первые 5 запросов должны быть разрешены, остальные — нет
		if i <= limit && !allowed {
			t.Fatalf("request %d should be allowed, but got false", i)
		}
		if i > limit && allowed {
			t.Fatalf("request %d should be blocked, but got true", i)
		}
	}
}

func TestAddToStoplist(t *testing.T) {
	// Запускаем мини-Redis
	mr := miniredis.RunT(t)
	defer mr.Close()

	// Подключаемся к нему
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisClient{Client: client}
	ctx := context.Background()

	// Добавляем слово
	err := rc.AddToStoplist(ctx, "spam")
	if err != nil {
		t.Fatalf("AddToStoplist failed: %v", err)
	}

	// Проверяем, что слово в стоп-листе
	ok, err := rc.IsStoplisted(ctx, "spam")
	if err != nil {
		t.Fatalf("IsStoplisted failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected stoplisted, got false")
	}

	// Удаляем слово
	err = rc.RemoveFromStoplist(ctx, "spam")
	if err != nil {
		t.Fatalf("RemoveFromStoplist failed: %v", err)
	}

	// Проверяем, что слова больше нет
	ok, err = rc.IsStoplisted(ctx, "spam")
	if err != nil {
		t.Fatalf("IsStoplisted failed: %v", err)
	}
	if ok {
		t.Fatalf("expected not stoplisted, got true")
	}
}

func TestIncrementWordAndGetTop(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &RedisClient{Client: client}
	ctx := context.Background()

	// Добавляем 5 запросов для "айфон"
	for i := 0; i < 5; i++ {
		err := rc.IncrementWord(ctx, "айфон")
		if err != nil {
			t.Fatalf("IncrementWord failed: %v", err)
		}
	}

	// Добавляем 3 запроса для "пылесос"
	for i := 0; i < 3; i++ {
		err := rc.IncrementWord(ctx, "пылесос")
		if err != nil {
			t.Fatalf("IncrementWord failed: %v", err)
		}
	}

	// Добавляем 1 запрос для "телевизор"
	err := rc.IncrementWord(ctx, "телевизор")
	if err != nil {
		t.Fatalf("IncrementWord failed: %v", err)
	}

	// Получаем топ-2
	top, err := rc.GetTop(ctx, 2)
	if err != nil {
		t.Fatalf("GetTop failed: %v", err)
	}

	if len(top) != 2 {
		t.Fatalf("expected 2 words, got %d", len(top))
	}

	if top[0].Word != "айфон" && top[0].Count != 5 {
		t.Fatalf("expected 'айфон' with count 5, got %s with count %d", top[0].Word, top[0].Count)
	}

	if top[1].Word != "пылесос" && top[1].Count != 3 {
		t.Fatalf("expected 'пылесос' with count 3, got %s with count %d", top[1].Word, top[1].Count)
	}
}
