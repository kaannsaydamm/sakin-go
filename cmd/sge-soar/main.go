package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"sakin-go/cmd/sge-soar/config"
	"sakin-go/cmd/sge-soar/engine"
	"sakin-go/pkg/messaging"
	"sakin-go/pkg/models"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("[SOAR] Starting SGE Automation Response...")

	// 1. NATS
	nc, err := messaging.NewClient(&messaging.NatsConfig{
		URL:           cfg.NatsURL,
		Username:      cfg.NatsUser,
		Password:      cfg.NatsPassword,
		ReconnectWait: 2 * time.Second,
	})
	if err != nil {
		log.Fatalf("[SOAR] NATS Error: %v", err)
	}
	defer nc.Close()

	// 2. Engine
	eng := engine.NewEngine(nc)

	// 3. Consume Alerts
	_, err = nc.QueueSubscribe(context.Background(), messaging.StreamAlerts, messaging.TopicAlerts, messaging.ConsumerSOAR, func(msg jetstream.Msg) {
		msg.Ack()

		var alert models.Alert
		if err := json.Unmarshal(msg.Data(), &alert); err != nil {
			return
		}

		// Parallel execution of playbooks
		go eng.Execute(context.Background(), &alert)

	})

	if err != nil {
		log.Fatalf("[SOAR] Subscribe failed: %v", err)
	}

	// Wait
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[SOAR] Shutting down...")
}
