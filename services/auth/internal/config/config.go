package config

import "os"

type Config struct {
	Port      string
	DB_DSN    string
	RedisAddr string
	JWTSecret string
}

// LoadConfig memuat konfigurasi dari environment variable dengan nilai default
func LoadConfig() *Config {
	return &Config{
		Port:      getEnv("PORT", "8081"),
		DB_DSN:    getEnv("DB_DSN", "root:admin@tcp(localhost:3306)/auth_db?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "supersecretjwtkeyforauthservice123!"),
	}
}

// getEnv adalah fungsi pembantu untuk membaca env atau mengembalikan default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
