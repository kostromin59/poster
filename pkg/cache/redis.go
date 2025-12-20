package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) (*Redis, error) {
	status := client.Ping(context.Background())
	if status.Err() != nil {
		return nil, status.Err()
	}

	return &Redis{
		client: client,
	}, nil
}

func (r *Redis) Set(ctx context.Context, key string, data []byte) error {
	return r.SetWithExpiration(ctx, key, data, 0)
}

func (r *Redis) SetWithExpiration(ctx context.Context, key string, data []byte, exp time.Duration) error {
	status := r.client.Set(ctx, key, data, exp)
	if status.Err() != nil {
		if errors.Is(status.Err(), redis.Nil) {
			return nil
		}

		return status.Err()
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	return []byte(val), nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	status := r.client.Del(ctx, key)
	if status.Err() != nil {
		if errors.Is(status.Err(), redis.Nil) {
			return nil
		}

		return status.Err()
	}

	return nil
}
