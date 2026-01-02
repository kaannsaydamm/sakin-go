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

	"sakin-go/cmd/sge-enrichment/config"
	"sakin-go/cmd/sge-enrichment/geoip"
	"sakin-go/cmd/sge-enrichment/intel"
	"sakin-go/pkg/database"
	"sakin-go/pkg/messaging"
	"sakin-go/pkg/models"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("[Enrichment] Starting SGE Enrichment Service...")

	// 1. Infrastructure
	// NATS
	natsCfg := &messaging.NatsConfig{
		URL: cfg.NatsURL, Username: cfg.NatsUser, Password: cfg.NatsPassword,
		ReconnectWait: 2 * time.Second,
	}
	nc, err := messaging.NewClient(natsCfg)
	if err != nil {
		log.Fatalf("[Enrichment] NATS Error: %v", err)
	}
	defer nc.Close()

	// Redis
	rdb, _ := database.NewRedisClient(&database.RedisConfig{
		Addr: cfg.RedisAddr, Password: cfg.RedisPassword,
	})
	if rdb != nil {
		defer rdb.Close()
	}

	// 2. Providers
	intelProvider := intel.NewCachingProvider(rdb)
	geoProvider, _ := geoip.NewProvider(cfg.MaxMindPath)
	defer geoProvider.Close()

	// 3. Process Loop
	// Subscribe to RAW events
	// Subscribe to RAW events
	// Stream name is messaging.StreamEvents ("EVENTS")
	_, err = nc.QueueSubscribe(context.Background(), messaging.StreamEvents, messaging.TopicEventsRaw, messaging.ConsumerEnrichment, func(msg jetstream.Msg) {
		msg.Ack()

		var evt models.Event
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			return
		}

		// ENRICHMENT LOGIC

		// 3.1 Host Enrichment (GeoIP)
		if evt.SourceIP != "" {
			if loc := geoProvider.Lookup(evt.SourceIP); loc != nil {
				if evt.Enrichment == nil {
					evt.Enrichment = make(map[string]interface{})
				}
				evt.Enrichment["src_geo_country"] = loc.Country
				evt.Enrichment["src_geo_city"] = loc.City
				evt.Enrichment["src_geo_iso"] = loc.ISO

			}

			// 3.2 Intel Enrichment
			rep, _ := intelProvider.CheckIP(context.Background(), evt.SourceIP)
			if rep != nil && rep.IsMalicious {
				if evt.Enrichment == nil {
					evt.Enrichment = make(map[string]interface{})
				}
				evt.Enrichment["threat_intel_score"] = rep.Score
				evt.Enrichment["threat_intel_source"] = rep.Source

				// Escalate severity if malicious
				evt.Severity = models.SeverityCritical
				evt.Tags = append(evt.Tags, "malicious_ip")

			}
		}

		// 4. Republish if enriched (or simply passthrough all to enriched stream?
		// Usually passthrough is better for unified downstream)
		// Subject: events.enriched.<severity>.<source>
		subject := messaging.TopicEventsEnriched + string(evt.Severity) + "." + evt.Source

		outBytes, _ := json.Marshal(evt)
		nc.PublishAsync(context.Background(), subject, outBytes)

	})

	if err != nil {
		log.Fatalf("[Enrichment] Subscribe failed: %v", err)
	}

	// Wait
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[Enrichment] Shutting down...")
}
