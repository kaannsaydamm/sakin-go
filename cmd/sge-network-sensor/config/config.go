package config

import (
	"os"
)

// AppConfig represents the global configuration for the Network Sensor.
type AppConfig struct {
	SensorName      string
	Interface       string // comma separated or "all"
	PromiscuousMode bool
	SnapLen         int32
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
		SnapLen:         65535,                    // Capture full packet
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
