package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error)
}

type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{client: client}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, key, window).Err(); err != nil {
			return false, 0, err
		}
	}
	ttl, err := l.client.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	return int(count) <= limit, ttl, nil
}

type MemoryLimiter struct {
	mu    sync.Mutex
	store map[string]memoryEntry
	clock func() time.Time
}

type memoryEntry struct {
	count   int
	expires time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		store: map[string]memoryEntry{},
		clock: time.Now,
	}
}

func (l *MemoryLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock()
	entry, ok := l.store[key]
	if !ok || now.After(entry.expires) {
		entry = memoryEntry{count: 0, expires: now.Add(window)}
	}
	entry.count++
	l.store[key] = entry
	return entry.count <= limit, time.Until(entry.expires), nil
}
