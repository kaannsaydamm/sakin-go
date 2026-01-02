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

	"sakin-go/cmd/sge-correlation/config"
	"sakin-go/cmd/sge-correlation/engine"
	"sakin-go/pkg/database"
	"sakin-go/pkg/messaging"
	"sakin-go/pkg/models"
	"sakin-go/pkg/utils"
)

func main() {
	// 1. Config
	cfg := config.LoadConfig()
	log.Println("[Correlation] Starting SGE Correlation Engine...")

	// 2. Database Clients
	pgCfg := &database.PostgresConfig{
		Host:     cfg.PostgresAddr, // Needs port splitting logic
		Port:     5432,
		Username: cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		Database: cfg.PostgresDB,
		SSLMode:  "disable",
	}
	_, _ = database.NewPostgresClient(pgCfg)
	// assuming connected for now

	// 3. NATS
	natsConfig := &messaging.NatsConfig{
		URL:           cfg.NatsURL,
		Username:      cfg.NatsUser,
		Password:      cfg.NatsPassword,
		ReconnectWait: 2 * time.Second,
	}
	nc, err := messaging.NewClient(natsConfig)
	if err != nil {
		log.Fatalf("[Correlation] NATS Error: %v", err)
	}
	defer nc.Close()

	// 4. Rule Engine
	eng := engine.NewEngine()
	// TODO: Load rules from Postgres using pg client
	// For demo, load a dummy rule
	dummyRule := &models.Rule{
		ID:        "rule-001",
		Name:      "Critical Severity Event",
		Condition: "Event.Severity == 'critical'",
		Severity:  models.SeverityCritical,
	}
	eng.LoadRules([]*models.Rule{dummyRule})

	// 5. Consumption Loop
	// Queue Subscribe ensures load balancing if multiple correlation instances run
	_, err = nc.QueueSubscribe(context.Background(), messaging.StreamEvents, messaging.TopicEventsRaw, messaging.ConsumerCorrelation, func(msg jetstream.Msg) {
		// Ack immediately or manual? Manual is safer.
		msg.Ack()

		var evt models.Event
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			log.Printf("[Correlation] Unmarshal error: %v", err)
			return
		}

		// Evaluate
		matchedRules := eng.Evaluate(&evt)
		if len(matchedRules) > 0 {
			for _, r := range matchedRules {
				// Raise Alert
				alert := models.Alert{
					ID:        utils.GenerateID(),
					RuleID:    r.ID,
					Title:     r.Name,
					Severity:  r.Severity,
					Status:    models.AlertStatusNew,
					CreatedAt: time.Now().UTC(),
					EventIDs:  []string{evt.ID},
				}

				// Publish Alert
				alertBytes, _ := json.Marshal(alert)
				subject := messaging.TopicAlerts + string(alert.Severity) + "." + r.ID
				nc.PublishAsync(context.Background(), subject, alertBytes)

				// Save to DB (Async optimized)
				go func(a models.Alert) {
					// pg.CreateAlert(context.Background(), &a) // Implement in pkg/database
				}(alert)

				log.Printf("[Correlation] 🚨 ALERT Generated: %s (Rule: %s)", alert.Title, r.Name)
			}
		}

	})

	if err != nil {
		log.Fatalf("[Correlation] Subscribe failed: %v", err)
	}

	// 6. Wait
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[Correlation] Shutting down...")
}
