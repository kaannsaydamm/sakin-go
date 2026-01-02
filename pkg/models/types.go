package models

import "time"

// Enums
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type EventStatus string

const (
	EventStatusNew        EventStatus = "new"
	EventStatusProcessing EventStatus = "processing"
	EventStatusEnriched   EventStatus = "enriched"
	EventStatusArchived   EventStatus = "archived"
)

type AlertStatus string

const (
	AlertStatusNew           AlertStatus = "new"
	AlertStatusInvestigating AlertStatus = "investigating"
	AlertStatusClosed        AlertStatus = "closed"
)

// Event, sistemdeki tüm olayların temel veri yapısıdır.
type Event struct {
	ID          string                 `json:"id" db:"id"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Source      string                 `json:"source" db:"source"`
	SourceIP    string                 `json:"source_ip" db:"source_ip"`
	DestIP      string                 `json:"dest_ip" db:"dest_ip"`
	EventType   string                 `json:"event_type" db:"event_type"`
	Severity    Severity               `json:"severity" db:"severity"`
	Status      EventStatus            `json:"status" db:"status"`
	Description string                 `json:"description" db:"description"`
	RawLog      string                 `json:"raw_log" db:"raw_log"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string               `json:"tags,omitempty"`
	Enrichment  map[string]interface{} `json:"enrichment,omitempty"`

	// Flattened Enrichment (for simple DB mapping if needed, but Map is better for JSONB)
	// Keeping flexible map above.
}

// Alert, korelasyon motoru tarafından üretilen uyarı yapısıdır.
type Alert struct {
	ID          string                 `json:"id" db:"id"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	RuleID      string                 `json:"rule_id" db:"rule_id"`
	Title       string                 `json:"title" db:"title"` // Renamed from RuleName to Title for generic usage
	Severity    Severity               `json:"severity" db:"severity"`
	Description string                 `json:"description" db:"description"`
	EventIDs    []string               `json:"event_ids" db:"event_ids"`
	Status      AlertStatus            `json:"status" db:"status"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// Rule, korelasyon kurallarını temsil eder.
type Rule struct {
	ID        string   `json:"id" db:"id"`
	Name      string   `json:"name" db:"name"`
	Condition string   `json:"condition" db:"condition"` // Renamed from Expression
	Severity  Severity `json:"severity" db:"severity"`
	Enabled   bool     `json:"enabled" db:"enabled"`
	Actions   []string `json:"actions" db:"actions"`
}

// Asset, izlenen varlıkları temsil eder.
type Asset struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	IPAddress string `json:"ip_address" db:"ip_address"`
}
