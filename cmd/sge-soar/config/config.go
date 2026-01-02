package config

import (
	"os"
)

type Config struct {
	NatsURL      string
	NatsUser     string
	NatsPassword string
}

func LoadConfig() *Config {
	return &Config{
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		NatsUser:     getEnv("NATS_USER", "admin"),
		NatsPassword: getEnv("NATS_PASSWORD", "sakin123"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
