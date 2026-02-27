package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) PushQueue(ctx context.Context, queueName string, message string) error {
	return s.client.RPush(ctx, queueName, message).Err()
}

func (s *RedisStore) PopQueue(ctx context.Context, queueName string, timeout time.Duration) (string, error) {
	// BLPop expects timeout in seconds (float) or 0 for infinite. 
	// go-redis takes time.Duration.
	res, err := s.client.BLPop(ctx, timeout, queueName).Result()
	if err != nil {
		return "", err
	}
	if len(res) < 2 {
		return "", fmt.Errorf("invalid response from BLPop")
	}
	return res[1], nil
}

func (s *RedisStore) SaveResult(ctx context.Context, hashKey string, field string, value string, ttl time.Duration) error {
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, hashKey, field, value)
	pipe.Expire(ctx, hashKey, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetResults(ctx context.Context, hashKey string) (map[string]string, error) {
	return s.client.HGetAll(ctx, hashKey).Result()
}

func (s *RedisStore) SetMeta(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *RedisStore) GetMeta(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}
