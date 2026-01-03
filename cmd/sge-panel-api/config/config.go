package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string

	// ClickHouse
	ClickHouseHost string
	ClickHousePort int
	ClickHouseDB   string
	ClickHouseUser string
	ClickHousePass string

	// Postgres
	PostgresHost string
	PostgresPort int
	PostgresUser string
	PostgresPass string
	PostgresDB   string

	RedisAddr string
}

func LoadConfig() *Config {
	return &Config{
		Port:      getEnv("PANEL_PORT", ":8080"),
		JWTSecret: getEnv("JWT_SECRET", "sakin-secret-key-change-me"),

		ClickHouseHost: getEnv("CLICKHOUSE_ADDR", "localhost"),
		ClickHousePort: 9000,
		ClickHouseDB:   getEnv("CLICKHOUSE_DB", "default"),
		ClickHouseUser: getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePass: getEnv("CLICKHOUSE_PASSWORD", ""),

		PostgresHost: getEnv("POSTGRES_ADDR", "localhost"),
		PostgresPort: 5432,
		PostgresUser: getEnv("POSTGRES_USER", "postgres"),
		PostgresPass: getEnv("POSTGRES_PASSWORD", "sakin123"),
		PostgresDB:   getEnv("POSTGRES_DB", "sge_db"),

		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
