package db_redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	rdInstance *redis.Client
	rdOnce     sync.Once
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

type RedisData struct {
	Key   string
	Value interface{}
	TTL   time.Duration
}

func init() {
	msg := fmt.Sprintf("init redis client %v", time.Now())
	fmt.Println(msg)
}

func RedisConnection(ctx context.Context) (*RedisService, error) {

	rdOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:           "localhost:6379",
			Password:       "",
			DB:             0,
			PoolSize:       16,
			MinIdleConns:   8,
			MaxActiveConns: 16,
		})

		err := client.Ping(ctx).Err()
		if err != nil {
			return
		}
		rdInstance = client
	})

	return &RedisService{
		client: rdInstance,
		ctx:    ctx,
	}, nil
}

func (r *RedisService) Close() {
	_ = r.client.Close().Error()
}

func (r *RedisService) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisService) Set(ctx context.Context, data RedisData) error {
	err := r.client.Set(ctx, data.Key, data.Value, data.TTL).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (r *RedisService) Del(ctx context.Context, data RedisData) error {
	err := r.client.Del(ctx, data.Key).Err()
	if err != nil {
		return fmt.Errorf("redis del err: %v", err)
	}
	return nil
}

func (r *RedisService) Get(ctx context.Context, data RedisData) (string, error) {
	val, err := r.client.Get(ctx, data.Key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get err: %v", err)
	}

	return val, nil
}
