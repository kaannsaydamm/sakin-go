package actions

import (
	"context"
	"fmt"
	"log"
	"time"

	"sakin-go/pkg/messaging"
)

// ExecutionContext holds data available to an action (e.g. the Alert).
type ExecutionContext struct {
	AlertID    string
	TargetIP   string
	NatsClient *messaging.Client
}

// Action interface
type Action interface {
	Name() string
	Execute(ctx context.Context, execCtx *ExecutionContext, params map[string]interface{}) error
}

// Registry
var Registry = make(map[string]Action)

func Register(a Action) {
	Registry[a.Name()] = a
}

// --- Implementation: Block IP ---

type BlockIPAction struct{}

func (a *BlockIPAction) Name() string { return "block_ip" }

func (a *BlockIPAction) Execute(ctx context.Context, execCtx *ExecutionContext, params map[string]interface{}) error {
	log.Printf("[SOAR] Executing BlockIP on %s (Alert: %s)", execCtx.TargetIP, execCtx.AlertID)

	// In real system: Send command to Firewall Agent or API
	// Simulation: Publish a command to NATS
	// Subject: commands.<agent_id> (We assume we know which agent controls the firewall, or broadcast)
	// For MVP: Just log success

	// Command payload
	cmd := fmt.Sprintf(`{"action": "firewall_block", "ip": "%s"}`, execCtx.TargetIP)

	if execCtx.NatsClient != nil {
		subject := messaging.TopicCommands + "firewall-agent" // simplified
		execCtx.NatsClient.PublishAsync(ctx, subject, []byte(cmd))
	}

	time.Sleep(100 * time.Millisecond) // Simulate work
	return nil
}

// --- Implementation: Send Slack Notification ---

type SlackNotifyAction struct{}

func (a *SlackNotifyAction) Name() string { return "slack_notify" }

func (a *SlackNotifyAction) Execute(ctx context.Context, execCtx *ExecutionContext, params map[string]interface{}) error {
	msg, _ := params["message"].(string)
	log.Printf("[SOAR] Sending Slack Notification: %s", msg)
	return nil
}

func init() {
	Register(&BlockIPAction{})
	Register(&SlackNotifyAction{})
}
