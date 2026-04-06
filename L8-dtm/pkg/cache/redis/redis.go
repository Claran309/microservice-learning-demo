package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type RedisCluster struct {
	client *redis.ClusterClient
}

type RedisConfig struct {
	Addrs        []string
	Password     string
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewRedisCluster(config RedisConfig) (*RedisCluster, error) {
	zap.L().Info("开始初始化Redis集群连接",
		zap.Strings("addrs", config.Addrs),
		zap.String("component", "redis"),
	)

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Password:     config.Password,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		zap.L().Error("× Redis集群连接失败",
			zap.Error(err),
			zap.String("component", "redis"),
		)
		return nil, errors.New("Redis集群连接失败: " + err.Error())
	}

	zap.L().Info("√ Redis集群连接成功",
		zap.String("component", "redis"),
	)

	return &RedisCluster{client: client}, nil
}

func (r *RedisCluster) Get(ctx context.Context, key string) (string, error) {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.operation", "GET"),
		attribute.String("redis.key", key),
	)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			span.SetAttributes(attribute.Bool("cache.hit", false))
			return "", nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× Redis Get失败",
			zap.Error(err),
			zap.String("key", key),
			zap.String("component", "redis"),
		)
		return "", err
	}

	span.SetAttributes(attribute.Bool("cache.hit", true))
	zap.L().Debug("√ Redis Get成功",
		zap.String("key", key),
		zap.String("component", "redis"),
	)
	return result, nil
}

func (r *RedisCluster) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Set")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.operation", "SET"),
		attribute.String("redis.key", key),
		attribute.String("redis.ttl", expiration.String()),
	)

	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× Redis Set失败",
			zap.Error(err),
			zap.String("key", key),
			zap.String("component", "redis"),
		)
		return err
	}

	zap.L().Debug("√ Redis Set成功",
		zap.String("key", key),
		zap.String("component", "redis"),
	)
	return nil
}

func (r *RedisCluster) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.SetJSON")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.operation", "SET_JSON"),
		attribute.String("redis.key", key),
	)

	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× Redis SetJSON失败",
			zap.Error(err),
			zap.String("key", key),
			zap.String("component", "redis"),
		)
		return err
	}

	zap.L().Debug("√ Redis SetJSON成功",
		zap.String("key", key),
		zap.String("component", "redis"),
	)
	return nil
}

func (r *RedisCluster) Del(ctx context.Context, keys ...string) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Del")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.operation", "DEL"),
		attribute.String("redis.keys", fmt.Sprintf("%v", keys)),
	)

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× Redis Del失败",
			zap.Error(err),
			zap.Strings("keys", keys),
			zap.String("component", "redis"),
		)
		return err
	}

	zap.L().Debug("√ Redis Del成功",
		zap.Strings("keys", keys),
		zap.String("component", "redis"),
	)
	return nil
}

func (r *RedisCluster) Exists(ctx context.Context, keys ...string) (int64, error) {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Exists")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.operation", "EXISTS"),
	)

	result, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	return result, nil
}

func (r *RedisCluster) Expire(ctx context.Context, key string, expiration time.Duration) error {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Expire")
	defer span.End()

	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (r *RedisCluster) Incr(ctx context.Context, key string) (int64, error) {
	tracer := otel.Tracer("redis")
	ctx, span := tracer.Start(ctx, "redis.Incr")
	defer span.End()

	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	return result, nil
}

func (r *RedisCluster) Close() error {
	return r.client.Close()
}

func (r *RedisCluster) GetClient() *redis.ClusterClient {
	return r.client
}
