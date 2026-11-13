package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedis membuat client Redis dan memverifikasi koneksi dengan PING
func InitRedis(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Kosong jika default Redis tanpa password
		DB:       0,  // Gunakan database index 0
	})

	// Buat timeout context untuk memastikan koneksi tidak menggantung jika Redis mati
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("gagal terkoneksi ke Redis: %w", err)
	}

	return rdb, nil
}
