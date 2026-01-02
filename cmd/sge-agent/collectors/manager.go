package collectors

import (
	"context"
	"log"

	"sakin-go/cmd/sge-agent/communicator"
)

// Start starts all OS-specific collectors.
func Start(ctx context.Context, comm *communicator.Communicator) {
	log.Println("[Collectors] Starting platform specific collectors...")
	startPlatformCollectors(ctx, comm)
}
