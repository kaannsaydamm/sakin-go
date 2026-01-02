package engine

import (
	"context"
	"log"

	"sakin-go/cmd/sge-soar/actions"
	"sakin-go/pkg/messaging"
	"sakin-go/pkg/models"
)

// Playbook definition
type Playbook struct {
	ID      string
	Name    string
	Trigger string // e.g., "Critical Severity", "RuleID=xyz"
	Steps   []PlaybookStep
}

type PlaybookStep struct {
	ActionName string
	Params     map[string]interface{}
}

// Engine executes playbooks.
type Engine struct {
	playbooks  []*Playbook
	natsClient *messaging.Client
}

func NewEngine(nc *messaging.Client) *Engine {
	e := &Engine{
		natsClient: nc,
	}
	e.loadDummyPlaybooks()
	return e
}

func (e *Engine) loadDummyPlaybooks() {
	// Demo Playbook: Auto-Block Critical Threats
	e.playbooks = append(e.playbooks, &Playbook{
		ID:      "pb-001",
		Name:    "Auto Block Critical IPs",
		Trigger: "critical",
		Steps: []PlaybookStep{
			{
				ActionName: "slack_notify",
				Params:     map[string]interface{}{"message": "Critical Alert Detected! Initiating Block."},
			},
			{
				ActionName: "block_ip",
				Params:     nil,
			},
		},
	})
}

// Execute checks if alert triggers any playbook and runs it.
func (e *Engine) Execute(ctx context.Context, alert *models.Alert) {
	// Simple matching logic for demo
	// In real world: Use 'expr' engine again for complex triggers

	for _, pb := range e.playbooks {
		shouldRun := false
		if pb.Trigger == string(alert.Severity) {
			shouldRun = true
		}
		// Check specific Rule ID
		// if pb.Trigger == "RuleID=" + alert.RuleID ...

		if shouldRun {
			log.Printf("[SOAR] Triggered Playbook: %s for Alert %s", pb.Name, alert.ID)
			e.runPlaybook(ctx, pb, alert)
		}
	}
}

func (e *Engine) runPlaybook(ctx context.Context, pb *Playbook, alert *models.Alert) {
	// Create context
	// Need to extract Target IP from alert (which usually comes from Event)
	// For MVP we assume we fetch event or Alert has it.
	// Let's assume Alert ID is enough for lookup or passed in context.
	// Simulating target IP extraction:
	targetIP := "1.2.3.4" // Placeholder

	execCtx := &actions.ExecutionContext{
		AlertID:    alert.ID,
		TargetIP:   targetIP,
		NatsClient: e.natsClient,
	}

	for _, step := range pb.Steps {
		action, exists := actions.Registry[step.ActionName]
		if !exists {
			log.Printf("[SOAR] Error: Action %s not found", step.ActionName)
			continue
		}

		if err := action.Execute(ctx, execCtx, step.Params); err != nil {
			log.Printf("[SOAR] Action Failed: %v", err)
			break // Stop playbook on failure?
		}
	}
}
