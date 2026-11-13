package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(addr string) (*redis.Client, error) {
	// 1. Buat instance client Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	// 2. Siapkan timeout context 5 detik (Fail-Fast)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3. Uji koneksi dengan perintah PING
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("gagal menghubungi server Redis: %w", err)
	}

	return rdb, nil
}
