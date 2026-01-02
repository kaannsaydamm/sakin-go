//go:build linux

package collectors

import (
	"context"
	"log"

	"sakin-go/cmd/sge-agent/communicator"
)

func startPlatformCollectors(ctx context.Context, comm *communicator.Communicator) {
	log.Println("[Collectors] Linux: Starting Auditd and Syslog collectors...")
	// TODO: Implement Auditd / Syslog runners here
	_, _ = ctx, comm
	// go runAuditd(ctx, comm)
}
