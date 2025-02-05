package redis

import (
	"Go-internship-Manifure/internal/monitoring"
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	successStatus = "success"
	errStatus     = "error"
)

type CacheInterface interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Close() error
}

type Cache struct {
	Client *redis.Client
}

var ctx = context.Background()

// Подключение к redis.
func NewCache(address, password string, db int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	res, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to redis: %s, %s", res, err)
	}

	log.Printf("Connected to redis: %s", res)

	return &Cache{Client: client}
}

// Получает ключ и проверяет, существует ли он в Redis.
func (c *Cache) Get(key string) (string, error) {
	log.Printf("Get redis key: %s", key)

	start := time.Now()

	res, err := c.Client.Get(ctx, key).Result()
	duration := time.Since(start).Seconds()

	status := successStatus

	if errors.Is(err, redis.Nil) {
		status = errStatus

		monitoring.RedisRequestsTotal.WithLabelValues("set", status).Inc()
		monitoring.RedisRequestDuration.WithLabelValues("set").Observe(duration)

		return "", nil
	} else if err != nil {
		status = errStatus

		monitoring.RedisRequestsTotal.WithLabelValues("set", status).Inc()
		monitoring.RedisRequestDuration.WithLabelValues("set").Observe(duration)

		return "", err
	}

	monitoring.RedisRequestsTotal.WithLabelValues("get", status).Inc()
	monitoring.RedisRequestDuration.WithLabelValues("get").Observe(duration)

	return res, nil
}

// Устанавливает ключ.
func (c *Cache) Set(key, value string, ttl time.Duration) error {
	log.Printf("Set redis key: %s", key)

	start := time.Now()

	err := c.Client.Set(ctx, key, value, ttl).Err()
	duration := time.Since(start).Seconds()

	status := successStatus
	if err != nil {
		status = errStatus
	}

	// обновление метрик.
	monitoring.RedisRequestsTotal.WithLabelValues("set", status).Inc()
	monitoring.RedisRequestDuration.WithLabelValues("set").Observe(duration)

	return err
}

func (c *Cache) Close() error {
	log.Println("Closing Redis connection...")
	if c.Client == nil {
		return errors.New("redis client is nil")
	}
	return c.Client.Close()
}
