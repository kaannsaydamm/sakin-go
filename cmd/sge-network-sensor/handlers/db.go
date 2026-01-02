package handlers

import (
	"context"
	"log"
	"time"

	"sakin-go/cmd/sge-network-sensor/inspector"
	"sakin-go/pkg/database"
)

// DBHandler manages database persistence.
type DBHandler struct {
	pg *database.PostgresClient
	ch *database.ClickHouseClient
}

// NewDBHandler creates a new DB persistence handler.
func NewDBHandler(pg *database.PostgresClient, ch *database.ClickHouseClient) *DBHandler {
	return &DBHandler{
		pg: pg,
		ch: ch,
	}
}

// ProcessEvents consumes network events and writes them to ClickHouse in batches.
func (h *DBHandler) ProcessEvents(ctx context.Context, envChan <-chan interface{}) {
	batchSize := 1000
	flushInterval := 2 * time.Second

	buffer := make([]map[string]interface{}, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.flush(buffer)
			return

		case e := <-envChan:
			event, ok := e.(inspector.NetworkEvent)
			if !ok {
				continue
			}

			// Map to ClickHouse schema structure (NetworkFlows)
			flow := map[string]interface{}{
				"id":          timestampToID(event.Timestamp), // optimize UUID gen?
				"timestamp":   event.Timestamp,
				"source_ip":   event.SrcIP,
				"source_port": event.SrcPort,
				"dest_ip":     event.DstIP,
				"dest_port":   event.DstPort,
				"protocol":    event.Protocol,
				"bytes_sent":  uint64(event.PayloadSize), // Estimate
				// Add SNI/HTTP info to flags or extended fields if needed
			}

			buffer = append(buffer, flow)
			if len(buffer) >= batchSize {
				h.flush(buffer)
				buffer = buffer[:0] // Reset keep cap
			}

		case <-ticker.C:
			if len(buffer) > 0 {
				h.flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (h *DBHandler) flush(flows []map[string]interface{}) {
	if len(flows) == 0 {
		return
	}

	// Create context with timeout for flush
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.ch.InsertNetworkFlows(ctx, flows); err != nil {
		log.Printf("[DB] ClickHouse insert failed: %v", err)
	}
}

func timestampToID(t time.Time) string {
	// Simple ID gen for high throughput (better than UUID v4 for DB locality)
	// In real impl, use ULID or Snowflake
	return t.Format("20060102150405.000000")
}
