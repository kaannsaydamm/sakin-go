//go:build windows

package collectors

import (
	"context"
	"log"

	"sakin-go/cmd/sge-agent/communicator"
)

func startPlatformCollectors(ctx context.Context, comm *communicator.Communicator) {
	log.Println("[Collectors] Windows: Starting ETW and EventLog collectors...")
	// TODO: Implement ETW / EventLog runners here
	// go runETW(ctx, comm)
}
