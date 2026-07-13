package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lasfh/goapi/examples/api/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Redis: %w", err)
	}

	slog.Debug("Conexão com o Redis estabelecida.")

	return rdb, nil
}
