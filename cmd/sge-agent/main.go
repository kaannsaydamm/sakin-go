package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sakin-go/cmd/sge-agent/collectors"
	"sakin-go/cmd/sge-agent/collectors/host"
	"sakin-go/cmd/sge-agent/communicator"
	"sakin-go/cmd/sge-agent/config"
)

func main() {
	// 1. Config
	cfg := config.LoadConfig()
	log.Printf("[Agent] Starting SGE Agent (%s)...", cfg.AgentID)

	// 2. Communicator (mTLS)
	comm, err := communicator.NewCommunicator(cfg)
	if err != nil {
		log.Printf("[Agent] Warning: Failed to connect to server: %v", err)
		// We might want to retry loop here in real life
	} else {
		defer comm.Close()
		log.Println("[Agent] Connected to server securely.")
	}

	// 3. Execution Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Start Host Info Heartbeat
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.HostInfoInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info, err := host.CollectEnvInfo()
				if err != nil {
					log.Printf("[Agent] Error collecting host info: %v", err)
					continue
				}
				info.AgentID = cfg.AgentID

				// Publish to 'events.raw.info.host' or similar
				// Assuming topic structure: events.raw.<severity>.<source>
				if comm != nil {
					topic := "events.raw.info.agent"
					comm.Publish(topic, info.ToJSON())
				}
			}
		}
	}()

	// 5. Start Platform Collectors
	if comm != nil {
		collectors.Start(ctx, comm)
	}

	// 6. Wait for Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[Agent] Shutting down...")
	cancel()
	time.Sleep(1 * time.Second) // Give routines time to stop
	log.Println("[Agent] Goodbye.")
}
