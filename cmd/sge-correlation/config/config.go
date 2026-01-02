package config

import (
	"os"
)

type Config struct {
	NatsURL      string
	NatsUser     string
	NatsPassword string

	RedisAddr     string
	RedisPassword string

	PostgresAddr     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

func LoadConfig() *Config {
	return &Config{
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		NatsUser:     getEnv("NATS_USER", "admin"),
		NatsPassword: getEnv("NATS_PASSWORD", "sakin123"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		PostgresAddr:     getEnv("POSTGRES_ADDR", "localhost:5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "sakin123"),
		PostgresDB:       getEnv("POSTGRES_DB", "sge_db"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
