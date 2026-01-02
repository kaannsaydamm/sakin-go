package baseline

import (
	"log"
	"sync"
	"time"

	"sakin-go/pkg/models"
)

type Worker struct {
	counts map[string]int
	mu     sync.Mutex
}

func NewWorker() *Worker {
	w := &Worker{
		counts: make(map[string]int),
	}
	go w.run()
	return w
}

func (w *Worker) Process(evt *models.Event) {
	w.mu.Lock()
	w.counts[evt.Source]++
	w.mu.Unlock()
}

func (w *Worker) run() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		w.mu.Lock()
		for src, count := range w.counts {
			// Simple anomaly detection: > 1000 events/min is "high"
			if count > 1000 {
				log.Printf("[Analytics] 📈 High Volume Detected: %s (%d events/min)", src, count)
				// TODO: Publish a 'system.alert' or verify against historical Z-Score
			}
			delete(w.counts, src) // Reset
		}
		w.mu.Unlock()
	}
}
