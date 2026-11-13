package config

import "os"

type Config struct {
	Port         string
	DB_DSN       string
	RedisAddr    string
	KafkaBrokers string
	JWTSecret    string
}

func LoadConfig() *Config {
	// 1. Baca konfigurasi sistem dengan nilai default fallback
	return &Config{
		Port:         getEnv("PORT", "8082"),
		DB_DSN:       getEnv("DB_DSN", "root:admin@tcp(localhost:3306)/wallet_db?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:    getEnv("JWT_SECRET", "supersecretjwtkeyforauthservice123!"),
	}
}

func getEnv(key, fallback string) string {
	// 1. Periksa ketersediaan key pada environment OS
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	// 2. Gunakan fallback jika key tidak ditemukan
	return fallback
}
