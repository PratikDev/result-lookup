package redisClient

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis(logger *slog.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,  // use default DB
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("unable to connect to redis", "error", err)
		os.Exit(1)
	}

	return rdb
}

func Get(rdb *redis.Client, logger *slog.Logger, key string) (string, error) {
	val, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		// log only if it's NOT a
		// redis not found err
		if !errors.Is(err, redis.Nil) {
			logger.Error("failed to fetch data from redis", "key", key)
		}
		return "", err
	}

	return val, nil
}