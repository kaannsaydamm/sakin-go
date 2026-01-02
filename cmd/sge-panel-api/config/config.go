package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string

	ClickHouseAddr string
	PostgresAddr   string
	RedisAddr      string
}

func LoadConfig() *Config {
	return &Config{
		Port:      getEnv("PANEL_PORT", ":8080"),
		JWTSecret: getEnv("JWT_SECRET", "sakin-secret-key-change-me"),

		ClickHouseAddr: getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		PostgresAddr:   getEnv("POSTGRES_ADDR", "localhost:5432"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
