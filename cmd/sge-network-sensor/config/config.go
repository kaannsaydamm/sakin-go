package config

import (
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	SensorName      string
	Interface       string // comma separated or "all"
	PromiscuousMode bool
	SnapLen         int32
	BufferSize      int           // pcap buffer size in bytes
	ReadTimeout     time.Duration // pcap read timeout
	BPFFilter       string

	NatsURL      string
	NatsUser     string
	NatsPassword string

	ClickHouseAddr     string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string

	DebugMode bool
}

// LoadConfig loads configuration from environment variables (or defaults).
// In a real app, this might use viper or similar, but keeping it zero-alloc/simple here.
func LoadConfig() *AppConfig {
	return &AppConfig{
		SensorName:      getEnv("SENSOR_NAME", "sge-sensor-01"),
		Interface:       getEnv("SENSOR_INTERFACE", "any"),
		PromiscuousMode: getEnv("SENSOR_PROMISCUOUS", "true") == "true",
		SnapLen:         1600,                                         // Optimized: capture headers + some payload (MTU ~1500)
		BufferSize:      getEnvInt("SENSOR_BUFFER_SIZE", 8*1024*1024), // 8MB buffer
		ReadTimeout:     time.Duration(getEnvInt("SENSOR_TIMEOUT_MS", 100)) * time.Millisecond,
		BPFFilter:       getEnv("SENSOR_BPF", ""), // Empty defaults to capturing everything

		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		NatsUser:     getEnv("NATS_USER", "admin"),
		NatsPassword: getEnv("NATS_PASSWORD", "sakin123"),

		ClickHouseAddr:     getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDB:       getEnv("CLICKHOUSE_DB", "sge_logs"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),

		DebugMode: getEnv("DEBUG_MODE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
