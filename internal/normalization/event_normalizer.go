// Package normalization provides network event normalization and enrichment
// for the SGE Network Sensor. It converts protocol-specific events to a
// standardized format suitable for storage and analysis.
package normalization

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EventType represents the type of network event
type EventType string

const (
	EventTypeNetwork      EventType = "network"
	EventTypeTransport    EventType = "transport"
	EventTypeApplication  EventType = "application"
	EventTypeThreat       EventType = "threat"
	EventTypeConnection   EventType = "connection"
	EventTypeFlow         EventType = "flow"
)

// Severity represents the severity level of an event
type Severity string

const (
	SeverityDebug   Severity = "debug"
	SeverityInfo    Severity = "info"
	SeverityLow     Severity = "low"
	SeverityMedium  Severity = "medium"
	SeverityHigh    Severity = "high"
	SeverityCritical Severity = "critical"
)

// NetworkEvent represents a normalized network event
type NetworkEvent struct {
	// Core identification
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     EventType              `json:"event_type"`
	Protocol      string                 `json:"protocol"`
	Severity      Severity               `json:"severity"`
	InstanceID    string                 `json:"instance_id"`

	// Network layer information
	SourceIP      string                 `json:"source_ip"`
	DestIP        string                 `json:"dest_ip"`
	Interface     string                 `json:"interface,omitempty"`

	// Transport layer information
	SourcePort    uint16                 `json:"source_port"`
	DestPort      uint16                 `json:"dest_port"`
	TransportProtocol string             `json:"transport_protocol"`
	TCPFlags      string                 `json:"tcp_flags,omitempty"`
	SequenceNumber uint32                `json:"sequence_number,omitempty"`
	Acknowledgment uint32                `json:"acknowledgment,omitempty"`
	WindowSize    int                    `json:"window_size,omitempty"`

	// Application layer information
	ApplicationData map[string]interface{} `json:"application_data,omitempty"`
	PayloadPreview  string                `json:"payload_preview,omitempty"`
	PayloadSize     int                   `json:"payload_size"`

	// Enrichment data
	GeoLocation    *GeoLocation           `json:"geo_location,omitempty"`
	AssetInfo      *AssetInfo             `json:"asset_info,omitempty"`
	ThreatIntel    *ThreatIntel           `json:"threat_intel,omitempty"`
	Tags           []string               `json:"tags,omitempty"`

	// Raw data for debugging
	RawHeaders     map[string]string      `json:"raw_headers,omitempty"`
	PCAPReference  string                 `json:"pcap_reference,omitempty"`
}

// GeoLocation contains geographical location information for an IP
type GeoLocation struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	ISP         string  `json:"isp,omitempty"`
	ASN          int    `json:"asn,omitempty"`
	Org         string  `json:"org,omitempty"`
}

// AssetInfo contains asset identification information
type AssetInfo struct {
	Hostname     string `json:"hostname,omitempty"`
	OS           string `json:"os,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	DeviceType   string `json:"device_type,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Department   string `json:"department,omitempty"`
	Criticality  string `json:"criticality,omitempty"`
}

// ThreatIntel contains threat intelligence enrichment data
type ThreatIntel struct {
	ReputationScore int       `json:"reputation_score"`
	ThreatCategory  string    `json:"threat_category,omitempty"`
	ThreatFamily    string    `json:"threat_family,omitempty"`
	Campaign        string    `json:"campaign,omitempty"`
	Attribution     string    `json:"attribution,omitempty"`
	FirstSeen       time.Time `json:"first_seen,omitempty"`
	LastSeen        time.Time `json:"last_seen,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	References      []string  `json:"references,omitempty"`
}

// EventNormalizer normalizes network events to a standard format
type EventNormalizer struct {
	cfg              *NormalizerConfig
	geoEnricher      GeoEnricher
	assetEnricher    AssetEnricher
	threatEnricher   ThreatEnricher
	taggers          []Tagger
}

// NormalizerConfig holds configuration for the event normalizer
type NormalizerConfig struct {
	// Enable GeoIP enrichment
	GeoIPEnabled bool `mapstructure:"geoip_enabled"`
	// Enable asset enrichment
	AssetEnabled bool `mapstructure:"asset_enabled"`
	// Enable threat intelligence enrichment
	ThreatIntelEnabled bool `mapstructure:"threat_intel_enabled"`
	// Add default tags
	DefaultTags []string `mapstructure:"default_tags"`
	// Timestamp format
	TimestampFormat string `mapstructure:"timestamp_format"`
	// Enable payload truncation
	MaxPayloadPreview int `mapstructure:"max_payload_preview"`
}

// GeoEnricher interface for GeoIP enrichment
type GeoEnricher interface {
	Enrich(ip string) (*GeoLocation, error)
	Close() error
}

// AssetEnricher interface for asset enrichment
type AssetEnricher interface {
	Enrich(ip string) (*AssetInfo, error)
	Close() error
}

// ThreatEnricher interface for threat intelligence enrichment
type ThreatEnricher interface {
	Enrich(ip string) (*ThreatIntel, error)
	Close() error
}

// Tagger interface for custom tagging
type Tagger interface {
	Tag(event *NetworkEvent) []string
}

// NewEventNormalizer creates a new event normalizer
func NewEventNormalizer(cfg *NormalizerConfig) *EventNormalizer {
	return &EventNormalizer{
		cfg:      cfg,
		taggers:  make([]Tagger, 0),
	}
}

// SetGeoEnricher sets the GeoIP enricher
func (en *EventNormalizer) SetGeoEnricher(e GeoEnricher) {
	en.geoEnricher = e
}

// SetAssetEnricher sets the asset enricher
func (en *EventNormalizer) SetAssetEnricher(e AssetEnricher) {
	en.assetEnricher = e
}

// SetThreatEnricher sets the threat intelligence enricher
func (en *EventNormalizer) SetThreatEnricher(e ThreatEnricher) {
	en.threatEnricher = e
}

// AddTagger adds a custom tagger
func (en *EventNormalizer) AddTagger(t Tagger) {
	en.taggers = append(en.taggers, t)
}

// Normalize normalizes a network event
func (en *EventNormalizer) Normalize(event *NetworkEvent) (*NetworkEvent, error) {
	// Generate event ID if not set
	if event.ID == "" {
		event.ID = generateEventID(event.Timestamp)
	}

	// Normalize timestamp
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Normalize IP addresses
	event.SourceIP = normalizeIP(event.SourceIP)
	event.DestIP = normalizeIP(event.DestIP)

	// Normalize protocol names
	event.Protocol = normalizeProtocol(event.Protocol)
	event.TransportProtocol = normalizeProtocol(event.TransportProtocol)

	// Truncate payload preview
	if event.PayloadPreview != "" && en.cfg.MaxPayloadPreview > 0 {
		if len(event.PayloadPreview) > en.cfg.MaxPayloadPreview {
			event.PayloadPreview = event.PayloadPreview[:en.cfg.MaxPayloadPreview]
		}
	}

	// Add default tags
	event.Tags = append(event.Tags, en.cfg.DefaultTags...)

	// Apply custom taggers
	for _, tagger := range en.taggers {
		event.Tags = append(event.Tags, tagger.Tag(event)...)
	}

	// Remove duplicate tags
	event.Tags = uniqueTags(event.Tags)

	// Apply enrichments
	if en.geoEnricher != nil && en.cfg.GeoIPEnabled {
		if geo := en.geoEnricher.Enrich(event.SourceIP); geo != nil {
			event.GeoLocation = geo
		}
		if geo := en.geoEnricher.Enrich(event.DestIP); geo != nil {
			if event.GeoLocation == nil {
				event.GeoLocation = geo
			} else if geo.CountryCode != event.GeoLocation.CountryCode {
				// Cross-border communication - add tag
				event.Tags = append(event.Tags, "cross-border")
			}
		}
	}

	if en.assetEnricher != nil && en.cfg.AssetEnabled {
		if asset := en.assetEnricher.Enrich(event.SourceIP); asset != nil {
			event.AssetInfo = asset
		}
		if asset := en.assetEnricher.Enrich(event.DestIP); asset != nil {
			if event.AssetInfo == nil {
				event.AssetInfo = asset
			}
		}
	}

	if en.threatEnricher != nil && en.cfg.ThreatIntelEnabled {
		if threat := en.threatEnricher.Enrich(event.SourceIP); threat != nil {
			event.ThreatIntel = threat
			// Adjust severity based on reputation score
			event.Severity = adjustSeverity(event.Severity, threat.ReputationScore)
		}
	}

	return event, nil
}

// NormalizeBatch normalizes a batch of events
func (en *EventNormalizer) NormalizeBatch(events []*NetworkEvent) ([]*NetworkEvent, error) {
	normalized := make([]*NetworkEvent, 0, len(events))

	for _, event := range events {
		normEvent, err := en.Normalize(event)
		if err != nil {
			continue // Skip invalid events
		}
		normalized = append(normalized, normEvent)
	}

	return normalized, nil
}

// ConvertToJSON converts a network event to JSON bytes
func (e *NetworkEvent) ConvertToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ConvertToJSONString converts a network event to JSON string
func (e *NetworkEvent) ConvertToJSONString() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses a network event from JSON bytes
func (e *NetworkEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// Clone creates a deep copy of the event
func (e *NetworkEvent) Clone() *NetworkEvent {
	if e == nil {
		return nil
	}

	clone := *e

	// Deep copy maps
	if e.ApplicationData != nil {
		clone.ApplicationData = make(map[string]interface{})
		for k, v := range e.ApplicationData {
			clone.ApplicationData[k] = v
		}
	}

	if e.Tags != nil {
		clone.Tags = make([]string, len(e.Tags))
		copy(clone.Tags, e.Tags)
	}

	if e.RawHeaders != nil {
		clone.RawHeaders = make(map[string]string)
		for k, v := range e.RawHeaders {
			clone.RawHeaders[k] = v
		}
	}

	if e.GeoLocation != nil {
		loc := *e.GeoLocation
		clone.GeoLocation = &loc
	}

	if e.AssetInfo != nil {
		info := *e.AssetInfo
		clone.AssetInfo = &info
	}

	if e.ThreatIntel != nil {
		intel := *e.ThreatIntel
		clone.ThreatIntel = &intel
	}

	return &clone
}

// IsExternal returns true if the destination IP is external
func (e *NetworkEvent) IsExternal() bool {
	if e.DestIP == "" {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
	}

	for _, r := range privateRanges {
		if strings.HasPrefix(e.DestIP, r) {
			return false
		}
	}

	return true
}

// IsIncoming returns true if the traffic is incoming
func (e *NetworkEvent) IsIncoming() bool {
	// This would typically be determined based on network topology
	// For now, use a simple heuristic
	return e.DestPort == 80 || e.DestPort == 443 || e.DestPort == 22
}

// Helper functions

// generateEventID generates a unique event ID
func generateEventID(ts time.Time) string {
	return fmt.Sprintf("%d-%s", ts.UnixNano(), randomString(8))
}

// randomString generates a random string of specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// normalizeIP normalizes an IP address
func normalizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}

	// Handle IPv6 mapped IPv4
	if strings.HasPrefix(ip, "::ffff:") {
		ip = ip[7:]
	}

	return ip
}

// normalizeProtocol normalizes protocol names
func normalizeProtocol(protocol string) string {
	protocol = strings.ToUpper(strings.TrimSpace(protocol))

	// Protocol aliases
	aliases := map[string]string{
		"HTTP":       "HTTP",
		"HTTPS":      "HTTPS",
		"SSL":        "TLS",
		"TCP":        "TCP",
		"UDP":        "UDP",
		"ICMP":       "ICMP",
		"ICMPV6":     "ICMPv6",
		"DNS":        "DNS",
		"SMB":        "SMB",
		"SSH":        "SSH",
		"FTP":        "FTP",
		"TELNET":     "TELNET",
		"SMTP":       "SMTP",
		"POP3":       "POP3",
		"IMAP":       "IMAP",
		"Ldap":       "LDAP",
		"SNMP":       "SNMP",
		"NTP":        "NTP",
		"DHCP":       "DHCP",
	}

	if p, ok := aliases[protocol]; ok {
		return p
	}

	return protocol
}

// adjustSeverity adjusts severity based on threat reputation score
func adjustSeverity(current Severity, score int) Severity {
	if score >= 90 {
		return SeverityCritical
	}
	if score >= 70 {
		if current == SeverityInfo || current == SeverityLow {
			return SeverityHigh
		}
		return current
	}
	if score >= 50 {
		if current == SeverityInfo {
			return SeverityMedium
		}
		return current
	}

	return current
}

// uniqueTags removes duplicate tags
func uniqueTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}

// DefaultNormalizerConfig returns the default normalizer configuration
func DefaultNormalizerConfig() *NormalizerConfig {
	return &NormalizerConfig{
		GeoIPEnabled:      true,
		AssetEnabled:      true,
		ThreatIntelEnabled: true,
		DefaultTags:       []string{"network-sensor", "normalized"},
		TimestampFormat:   "rfc3339",
		MaxPayloadPreview: 256,
	}
}
