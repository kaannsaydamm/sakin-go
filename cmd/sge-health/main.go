package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"sakin-go/pkg/database"
	"sakin-go/pkg/messaging"
)

// ANSI Colors
const (
	Green = "\033[32m"
	Red   = "\033[31m"
	Reset = "\033[0m"
)

func main() {
	fmt.Println("🏥 SGE System Health Check")
	fmt.Println("=========================")

	overallStatus := true

	// 1. Check Redis
	if checkRedis() {
		printStatus("Redis", true)
	} else {
		printStatus("Redis", false)
		overallStatus = false
	}

	// 2. Check Postgres
	if checkPostgres() {
		printStatus("PostgreSQL", true)
	} else {
		printStatus("PostgreSQL", false)
		overallStatus = false
	}

	// 3. Check ClickHouse
	if checkClickHouse() {
		printStatus("ClickHouse", true)
	} else {
		printStatus("ClickHouse", false)
		overallStatus = false
	}

	// 4. Check NATS
	if checkNATS() {
		printStatus("NATS JetStream", true)
	} else {
		printStatus("NATS JetStream", false)
		overallStatus = false
	}

	fmt.Println("=========================")
	if overallStatus {
		fmt.Printf("%s✅ System Ready%s\n", Green, Reset)
		os.Exit(0)
	} else {
		fmt.Printf("%s❌ System Unhealthy%s\n", Red, Reset)
		os.Exit(1)
	}
}

func printStatus(service string, up bool) {
	if up {
		fmt.Printf("[%sOK%s] %s\n", Green, Reset, service)
	} else {
		fmt.Printf("[%sFAIL%s] %s\n", Red, Reset, service)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func checkRedis() bool {
	client, _ := database.NewRedisClient(&database.RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
	})
	if client == nil {
		return false
	}
	defer client.Close()
	return client.Ping(context.Background()) == nil
}

func checkPostgres() bool {
	client, err := database.NewPostgresClient(&database.PostgresConfig{
		Host:     getEnv("POSTGRES_ADDR", "localhost"),
		Port:     5432,
		Username: getEnv("POSTGRES_USER", "postgres"),
		Database: "postgres", // Just check connection to default db
		Password: getEnv("POSTGRES_PASSWORD", "sakin123"),
		SSLMode:  "disable",
	})
	if err != nil {
		return false
	}
	// defer client.Close() // PostgresClient doesn't expose Close easily in current pkg, rely on pool
	_, err = client.Health(context.Background())
	return err == nil
}

func checkClickHouse() bool {
	client, err := database.NewClickHouseClient(&database.ClickHouseConfig{
		Host:     getEnv("CLICKHOUSE_ADDR", "localhost"),
		Port:     9000,
		Database: "default",
		Username: "default",
	})
	if err != nil {
		return false
	}
	// Ping?
	return client.Conn().Ping(context.Background()) == nil
}

func checkNATS() bool {
	nc, err := messaging.NewClient(&messaging.NatsConfig{
		URL:           getEnv("NATS_URL", "nats://localhost:4222"),
		ReconnectWait: 100 * time.Millisecond,
		MaxReconnects: 1,
	})
	if err != nil {
		return false
	}
	defer nc.Close()
	return nc.Connection().IsConnected()
}
