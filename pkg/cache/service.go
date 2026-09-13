package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"basilisk/pkg/helper"

	"github.com/redis/go-redis/v9"
)

var instance Cache

const defaultCacheExp = 1 * time.Hour

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Cache interface {
	Add(ctx context.Context, logger *slog.Logger, key string, data any, ttl ...time.Duration) error
	Get(ctx context.Context, logger *slog.Logger, key string, target any) error
	Delete(ctx context.Context, logger *slog.Logger, key string) error
}

type CacheInst struct {
	RedisInstance *redis.Client
}

func SetInstance(c Cache) {
	instance = c
}

func GetInstance() Cache {
	return instance
}

func Init(ctx context.Context, c Config) error {

	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Password,
		DB:       0,
	})

	// verify if connection is made or not
	_, err := rc.Ping(ctx).Result()
	if err != nil {
		return err
	}

	instance = &CacheInst{
		RedisInstance: rc,
	}

	return nil
}

func (c *CacheInst) Close() error {
	return c.RedisInstance.Close()
}

func (c *CacheInst) Add(ctx context.Context, logger *slog.Logger, key string, data any, ttl ...time.Duration) error {
	var valToStore any
	var ttlV time.Duration

	// check if ttl is being passed or not, if
	// yes then set it to be used.
	if len(ttl) > 0 {
		ttlV = ttl[0]
	} else {
		ttlV = defaultCacheExp
	}

	switch v := data.(type) {
	case string, []byte:
		valToStore = v

	// Signed Integers
	case int, int8, int16, int32, int64:
		valToStore = v

	// Unsigned Integers & uintptr
	case uint, uint8, uint16, uint32, uint64, uintptr:
		valToStore = v

	// Floating-point numbers
	case float32, float64:
		valToStore = v

	// Booleans
	case bool:
		valToStore = v

	default:
		d, err := json.Marshal(data)
		if err != nil {
			logger.Error("error while marshalling data", "error", err)
			return err
		}
		valToStore = d
	}

	rdStatus := c.RedisInstance.Set(ctx, key, valToStore, ttlV)

	if rdStatus.Err() != nil {
		logger.Error("error while inserting data into cache", "error", rdStatus.Err())
		return rdStatus.Err()
	}

	return nil
}

// Get fetches data from cache and stores it in target
func (c *CacheInst) Get(ctx context.Context, logger *slog.Logger, key string, target any) error {
	val, err := c.RedisInstance.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Debug("key does not exist.Cache miss", "key", key)
			return helper.NotFoundError
		}
		logger.Error("error while getting data", "key", key, "error", err)
		return helper.InternalServerError
	}

	switch ptr := target.(type) {
	case *string:
		*ptr = string(val)
	case *[]byte:
		*ptr = val
	default:
		if err := json.Unmarshal(val, target); err != nil {
			logger.Error("error while unmarshalling data", "error", err)
			return helper.InternalServerError
		}
	}

	return nil
}

// Delete removes data from cache
func (c *CacheInst) Delete(ctx context.Context, logger *slog.Logger, key string) error {
	if err := c.RedisInstance.Del(ctx, key).Err(); err != nil {
		logger.Error("error while deleting data from cache", "key", key)
		return helper.InternalServerError
	}

	return nil
}
