package config

import (
	"os"
)

type Config struct {
	NatsURL      string
	NatsUser     string
	NatsPassword string

	ClickHouseAddr     string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string

	BatchSize     int
	FlushInterval int // Seconds
}

func LoadConfig() *Config {
	return &Config{
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		NatsUser:     getEnv("NATS_USER", "admin"),
		NatsPassword: getEnv("NATS_PASSWORD", "sakin123"),

		ClickHouseAddr:     getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDB:       getEnv("CLICKHOUSE_DB", "sge_logs"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),

		BatchSize:     5000,
		FlushInterval: 5,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
