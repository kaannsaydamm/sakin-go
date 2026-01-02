package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sakin-go/cmd/sge-network-sensor/config"
	"sakin-go/cmd/sge-network-sensor/handlers"
	"sakin-go/cmd/sge-network-sensor/inspector"
	"sakin-go/pkg/database"
	"sakin-go/pkg/messaging"
)

func main() {
	// 1. Config
	cfg := config.LoadConfig()
	log.Println("[Main] Starting SGE Network Sensor:", cfg.SensorName)

	// 2. Database Clients
	// PostgreSQL (Optional if only network flows)
	// pg, _ := database.NewPostgresClient(...)

	// ClickHouse (Required for flows)
	chCfg := &database.ClickHouseConfig{
		Host:     cfg.ClickHouseAddr, // Needs splitting host/port in real impl
		Port:     9000,
		Database: cfg.ClickHouseDB,
		Username: cfg.ClickHouseUser,
		Password: cfg.ClickHousePassword,
	}
	// Warning: Error handling simplified for snippet
	ch, err := database.NewClickHouseClient(chCfg)
	if err != nil {
		log.Printf("[Main] Warning: ClickHouse not connected: %v", err)
	}

	// 3. NATS Client
	natsConfig := &messaging.NatsConfig{
		URL:           cfg.NatsURL,
		Username:      cfg.NatsUser,
		Password:      cfg.NatsPassword,
		MaxReconnects: 5,
		ReconnectWait: 2 * time.Second,
	}
	nc, err := messaging.NewClient(natsConfig)
	if err != nil {
		log.Fatalf("[Main] NATS connection failed: %v", err)
	}
	defer nc.Close()

	// 4. Setup Pipeline
	// Buffered channel for events
	eventChan := make(chan interface{}, 10000)

	// Inspector (Producer)
	insp := inspector.NewInspector(cfg, eventChan)

	// DB Handler (Consumer 1)
	if ch != nil {
		dbHandler := handlers.NewDBHandler(nil, ch)
		go dbHandler.ProcessEvents(context.Background(), eventChan)
	}

	// NATS Publisher (Consumer 2 - Logic needed inside handler or separate)
	// For now, simpler implementation: let's assume DBHandler handles flows
	// and we might want a separate routine for alerting events.

	// 5. Start Capture
	if err := insp.Start(); err != nil {
		log.Fatalf("[Main] Failed to start inspector: %v", err)
	}

	// 6. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("[Main] Shutting down...")

	insp.Stop()
	// Drain channel logic here...
	log.Println("[Main] Shutdown complete.")
}
