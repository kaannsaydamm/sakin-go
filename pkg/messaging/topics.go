package messaging

// Topic constants define the subject names for NATS JetStream.
// Using constants avoids memory allocation for topic strings during runtime.
const (
	// EventsRaw is the topic for raw events coming from agents/ingest.
	// Subject: events.raw.<severity>.<source>
	TopicEventsRaw = "events.raw.>"

	// EventsEnriched is the topic for events enriched with GeoIP/ThreatIntel.
	// Subject: events.enriched.<severity>.<source>
	TopicEventsEnriched = "events.enriched.>"

	// Alerts is the topic for generated alerts from correlation engine.
	// Subject: alerts.<severity>.<rule_id>
	TopicAlerts = "alerts.>"

	// SystemLogs is the topic for internal system logs.
	// Subject: system.logs.<service>.<level>
	TopicSystemLogs = "system.logs.>"

	// Commands is the topic for sending commands to agents.
	// Subject: commands.<agent_id>
	TopicCommands = "commands.>"
)

// Stream names
const (
	StreamEvents   = "SGE_EVENTS"
	StreamAlerts   = "SGE_ALERTS"
	StreamSystem   = "SGE_SYSTEM"
	StreamCommands = "SGE_COMMANDS"
)

// Consumer names (Durable)
const (
	ConsumerEnrichment  = "SGE_ENRICHMENT_PROCESSOR"
	ConsumerCorrelation = "SGE_CORRELATION_ENGINE"
	ConsumerArchival    = "SGE_ARCHIVAL_WORKER"
	ConsumerSOAR        = "SGE_SOAR_EXECUTOR"
)
