package normalizer

import (
	"encoding/json"
	"time"

	"sakin-go/pkg/models"
	"sakin-go/pkg/utils"
)

// NormalizeAgentEvent converts agent payload to standard Event model.
func NormalizeAgentEvent(data []byte) (*models.Event, error) {
	// Simple unmarshal for now, but in reality mapping huge raw data to schema
	// For demo, we assume Agent sends "Event" like structure or we Map it.

	// Let's assume Agent sends a struct capable of direct mapping or we construct it.
	// Since we haven't defined Agent's exact wire format beyond HostInfo, let's make a generic map.
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, err
	}

	evt := &models.Event{
		ID:        utils.GenerateID(),
		Timestamp: time.Now().UTC(),
		Source:    "agent", // Default
		Status:    models.EventStatusNew,
	}

	if val, ok := rawMap["source"].(string); ok {
		evt.Source = val
	}
	if val, ok := rawMap["event_type"].(string); ok {
		evt.EventType = val
	}
	if val, ok := rawMap["severity"].(string); ok {
		evt.Severity = models.Severity(val)
	}

	// ... Map other fields ...

	return evt, nil
}

// NormalizeSyslog converts syslog message to Event.
func NormalizeSyslog(msg string, remoteAddr string) *models.Event {
	// Syslog parsing logic (RFC3164/5424) would go here.
	// Simplified for MVP.
	return &models.Event{
		ID:        utils.GenerateID(),
		Timestamp: time.Now().UTC(),
		Source:    "syslog",
		SourceIP:  remoteAddr,
		EventType: "system.log",
		Severity:  models.SeverityInfo,
		RawLog:    msg,
		Status:    models.EventStatusNew,
	}
}
