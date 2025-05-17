package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

type RedisClient struct {
    Client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "", // если без пароля
        DB:       0,
    })

    return &RedisClient{Client: rdb}
}

func (r *RedisClient) Set(key string, value string, ttl time.Duration) error {
    return r.Client.Set(Ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
    return r.Client.Get(Ctx, key).Result()
}

func (r *RedisClient) Delete(key string) error {
    return r.Client.Del(Ctx, key).Err()
}
