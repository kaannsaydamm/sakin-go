package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"sakin-go/cmd/sge-ingest/config"
	"sakin-go/cmd/sge-ingest/handlers"
	"sakin-go/pkg/messaging"
)

func main() {
	// 1. Config
	cfg := config.LoadConfig()
	log.Println("[Ingest] Starting SGE Ingest Service...")

	// 2. NATS Connection
	natsConfig := &messaging.NatsConfig{
		URL:           cfg.NatsURL,
		Username:      cfg.NatsUser,
		Password:      cfg.NatsPassword,
		MaxReconnects: 10,
		ReconnectWait: 2 * time.Second,
	}
	nc, err := messaging.NewClient(natsConfig)
	if err != nil {
		log.Fatalf("[Ingest] NATS connect failed: %v", err)
	}
	defer nc.Close()

	// 3. Setup Streams
	// Ensure "SGE_EVENTS" stream exists
	if err := nc.InitializeStreams(context.Background()); err != nil {
		log.Printf("[Ingest] Warning: Stream initialization failed: %v", err)
	}

	// 4. Fiber App
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             10 * 1024 * 1024, // 10MB limit
	})

	// Handlers
	eventHandler := handlers.NewEventHandler(nc)

	// Routes
	api := app.Group("/api/v1")
	api.Post("/events", eventHandler.HandleHTTPEvent)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 5. Start Server
	go func() {
		if err := app.Listen(cfg.HTTPPort); err != nil {
			log.Fatalf("[Ingest] HTTP Listen failed: %v", err)
		}
	}()

	log.Printf("[Ingest] HTTP Server listening on %s", cfg.HTTPPort)

	// 6. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[Ingest] Shutting down...")
	app.Shutdown()
}
