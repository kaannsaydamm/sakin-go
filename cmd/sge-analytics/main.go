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

	"sakin-go/cmd/sge-analytics/baseline"
	"sakin-go/cmd/sge-analytics/config"
	"sakin-go/cmd/sge-analytics/sink"
	"sakin-go/pkg/database"
	"sakin-go/pkg/messaging"
	"sakin-go/pkg/models"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("[Analytics] Starting SGE Analytics & Archival Service...")

	// 1. ClickHouse
	chCfg := &database.ClickHouseConfig{
		Host: cfg.ClickHouseAddr, Port: 9000,
		Database: cfg.ClickHouseDB, Username: cfg.ClickHouseUser, Password: cfg.ClickHousePassword,
	} // Need robust error handling here
	// Assuming NewClickHouseClient handles "Host" parsing correctly or we split it properly.
	// For now assume config matches library expectation.

	chClient, err := database.NewClickHouseClient(chCfg)
	if err != nil {
		log.Printf("[Analytics] Warning: ClickHouse connect failed: %v", err)
	}

	// 2. NATS
	nc, err := messaging.NewClient(&messaging.NatsConfig{
		URL: cfg.NatsURL, Username: cfg.NatsUser, Password: cfg.NatsPassword,
		ReconnectWait: 2 * time.Second,
	})
	if err != nil {
		log.Fatalf("[Analytics] NATS Error: %v", err)
	}
	defer nc.Close()

	// 3. Components
	var eventSink *sink.ClickHouseSink
	if chClient != nil {
		eventSink = sink.NewClickHouseSink(cfg, chClient)
		defer eventSink.Close()
	}

	baWorker := baseline.NewWorker()

	// 4. Consume
	// We listen to Enriched events to store the final state of the event
	// 4. Consume
	// We listen to Enriched events to store the final state of the event
	_, err = nc.QueueSubscribe(context.Background(), messaging.StreamEvents, messaging.TopicEventsEnriched, messaging.ConsumerArchival, func(msg jetstream.Msg) {
		msg.Ack()

		var evt models.Event
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			return
		}

		// Parallel Processing
		if eventSink != nil {
			eventSink.Write(&evt)
		}

		baWorker.Process(&evt)

	})

	if err != nil {
		log.Fatalf("[Analytics] Subscribe failed: %v", err)
	}

	log.Println("[Analytics] Consuming events...")

	// Wait
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[Analytics] Shutting down...")
}
