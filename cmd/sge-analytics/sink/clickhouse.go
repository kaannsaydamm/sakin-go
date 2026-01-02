package sink

import (
	"context"
	"log"
	"sync"
	"time"

	"sakin-go/cmd/sge-analytics/config"
	"sakin-go/pkg/database"
	"sakin-go/pkg/models"
)

type ClickHouseSink struct {
	client *database.ClickHouseClient
	config *config.Config
	buffer []*models.Event
	mu     sync.Mutex
	done   chan struct{}
}

func NewClickHouseSink(cfg *config.Config, client *database.ClickHouseClient) *ClickHouseSink {
	s := &ClickHouseSink{
		client: client,
		config: cfg,
		buffer: make([]*models.Event, 0, cfg.BatchSize),
		done:   make(chan struct{}),
	}
	go s.flushLoop()
	return s
}

// Write adds an event to the buffer.
func (s *ClickHouseSink) Write(evt *models.Event) {
	s.mu.Lock()
	s.buffer = append(s.buffer, evt)
	shouldFlush := len(s.buffer) >= s.config.BatchSize
	s.mu.Unlock()

	if shouldFlush {
		s.Flush() // Async or sync? Sync here to apply backpressure if DB is slow
	}
}

// Flush forces a database write.
func (s *ClickHouseSink) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.buffer) == 0 {
		return
	}

	// Copy buffer to free it up for incoming writes while previous batch is writing
	// For simplicity in this non-concurrent write design (single buffer), we just write directly.
	// To optimize further, we could swap buffers.

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.client.InsertEvents(ctx, s.buffer); err != nil {
		log.Printf("[Sink] ClickHouse Insert Error: %v", err)
		// Retry logic would go here
	}

	// efficient clear
	// s.buffer = s.buffer[:0] // keep capacity
	// But we passed slice reference to InsertEvents, if it's async inside client it's dangerous.
	// database.InsertEvents copies or streams? It iterates.
	// To be safe and let GC reclaim references:
	clear(s.buffer) // Go 1.21+
	s.buffer = s.buffer[:0]
}

func (s *ClickHouseSink) flushLoop() {
	ticker := time.NewTicker(time.Duration(s.config.FlushInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			s.Flush()
			return
		case <-ticker.C:
			s.Flush()
		}
	}
}

func (s *ClickHouseSink) Close() {
	close(s.done)
}

func clear(s []*models.Event) {
	for i := range s {
		s[i] = nil
	}
}
