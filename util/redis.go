package util

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"swap-wallet/config"
)

func ConnectRedis(cfg config.Config) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.REDIS_HOST, cfg.REDIS_PORT),
		Password: cfg.REDIS_PASSWORD,
		DB:       0,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to Redis!")
	return redisClient, nil
}
