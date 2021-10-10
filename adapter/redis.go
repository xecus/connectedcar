package adapter

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/xecus/connectedcar/config"
)

type RedisClient struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisClient() *RedisClient {
	rc := &RedisClient{}
	return rc
}

func (rc *RedisClient) Init(globalConfig *config.GlobalConfig) error {
	rc.ctx = context.Background()

	rc.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return nil
}

func (rc *RedisClient) Write(key, value string) error {
	err := rc.rdb.Set(rc.ctx, "key", "value", 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisClient) Read(key string) (string, error) {
	val, err := rc.rdb.Get(rc.ctx, "key").Result()
	if err != nil {
		return "", err
	}
	return val, err
}
