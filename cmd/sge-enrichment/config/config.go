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

	AbuseIPDBKey string
	OTXKey       string
	MaxMindPath  string
}

func LoadConfig() *Config {
	return &Config{
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		NatsUser:     getEnv("NATS_USER", "admin"),
		NatsPassword: getEnv("NATS_PASSWORD", "sakin123"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		AbuseIPDBKey: getEnv("ABUSEIPDB_KEY", ""),
		OTXKey:       getEnv("OTX_KEY", ""),
		MaxMindPath:  getEnv("MAXMIND_DB_PATH", "./GeoLite2-City.mmdb"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
