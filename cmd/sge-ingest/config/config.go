package config

import (
	"os"
)

type IngestConfig struct {
	HTTPPort   string
	SyslogPort string
	DebugMode  bool

	NatsURL      string
	NatsUser     string
	NatsPassword string
}

func LoadConfig() *IngestConfig {
	return &IngestConfig{
		HTTPPort:   getEnv("INGEST_HTTP_PORT", ":8080"),
		SyslogPort: getEnv("INGEST_SYSLOG_PORT", "514"),
		DebugMode:  getEnv("DEBUG_MODE", "false") == "true",

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
