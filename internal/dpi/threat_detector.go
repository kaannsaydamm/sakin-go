package dpi

import (
	"container/list"
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atailh4n/sakin/internal/config"
	"github.com/atailh4n/sakin/internal/normalization"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// ThreatType represents the type of detected threat
type ThreatType string

const (
	ThreatTypeAnomaly          ThreatType = "anomaly"
	ThreatTypePortScan         ThreatType = "port_scan"
	ThreatTypeC2Beacon         ThreatType = "c2_beacon"
	ThreatTypeExfiltration     ThreatType = "exfiltration"
	ThreatTypeBruteForce       ThreatType = "brute_force"
	ThreatTypeMalformedPacket  ThreatType = "malformed_packet"
	ThreatTypeSuspiciousPayload ThreatType = "suspicious_payload"
	ThreatTypeLateralMovement  ThreatType = "lateral_movement"
)

// ThreatMatch represents a detected threat
type ThreatMatch struct {
	Type           ThreatType       `json:"type"`
	Severity       Severity         `json:"severity"`
	SourceIP       string           `json:"source_ip"`
	DestIP         string           `json:"dest_ip"`
	SourcePort     uint16           `json:"source_port"`
	DestPort       uint16           `json:"dest_port"`
	Protocol       string           `json:"protocol"`
	Description    string           `json:"description"`
	Score          int              `json:"score"`          // 0-100 confidence score
	Metadata       map[string]any   `json:"metadata"`
	Events         []*normalization.NetworkEvent `json:"events"`
	Timestamp      time.Time        `json:"timestamp"`
}

// Severity represents the severity level of a threat
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// ThreatDetector performs threat detection on network traffic
type ThreatDetector struct {
	cfg             *config.ThreatDetectConfig
	mu              sync.RWMutex
	portScanTracker *PortScanTracker
	beaconTracker   *BeaconTracker
	exfilTracker    *ExfiltrationTracker
	anomalyTracker  *AnomalyTracker
	stats           ThreatStats
}

// ThreatStats holds threat detection statistics
type ThreatStats struct {
	TotalThreats      uint64
	Anomalies         uint64
	PortScans         uint64
	C2Beacons         uint64
	Exfiltrations     uint64
	BruteForces       uint64
	LastThreatTime    time.Time
}

// PortScanTracker tracks potential port scan activity
type PortScanTracker struct {
	mu        sync.RWMutex
	entries   map[string]*PortScanEntry
	window    time.Duration
	threshold int
}

// PortScanEntry tracks port scan detection for a source IP
type PortScanEntry struct {
	SourceIP       string
	PortSet        map[uint16]bool
	UniquePorts    int
	FirstSeen      time.Time
	LastSeen       time.Time
	Packets        uint64
	Score          int
}

// BeaconTracker tracks potential C2 beacon activity
type BeaconTracker struct {
	mu        sync.RWMutex
	entries   map[string]*BeaconEntry
	window    time.Duration
	minScore  int
}

// BeaconEntry tracks beacon-like connection patterns
type BeaconEntry struct {
	SourceIP       string
	DestIP         string
	DestPort       uint16
	Intervals      []time.Duration
	LastInterval   time.Duration
	AvgInterval    time.Duration
	StdDev         float64
	FirstSeen      time.Time
	LastSeen       time.Time
	RequestCount   uint64
	ResponseCount  uint64
	TotalBytes     uint64
	Score          int
	Jitter         float64
}

// ExfiltrationTracker tracks potential data exfiltration
type ExfiltrationTracker struct {
	mu            sync.RWMutex
	entries       map[string]*ExfilEntry
	window        time.Duration
	rateThreshold int64
	volumeThreshold int64
}

// ExfilEntry tracks data transfer patterns
type ExfilEntry struct {
	SourceIP       string
	DestIP         string
	DestPort       uint16
	Protocol       string
	BytesSent      uint64
	BytesReceived  uint64
	StartTime      time.Time
	LastTime       time.Time
	Rate           float64
	Score          int
}

// AnomalyTracker tracks packet anomalies
type AnomalyTracker struct {
	mu           sync.RWMutex
	entries      map[string]*AnomalyEntry
	window       time.Duration
	burstCount   map[string]*list.Element
	burstList    *list.List
	maxBurst     int
}

// AnomalyEntry tracks packet anomalies for an IP
type AnomalyEntry struct {
	IP            string
	PacketSizes   []int
	BurstCount    int
	TotalPackets  uint64
	FirstSeen     time.Time
	LastSeen      time.Time
	Score         int
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector(cfg *config.ThreatDetectConfig) (*ThreatDetector, error) {
	td := &ThreatDetector{
		cfg:            cfg,
		portScanTracker: NewPortScanTracker(cfg.PortScan.Window, cfg.PortScan.PortThreshold),
		beaconTracker:   NewBeaconTracker(),
		exfilTracker:    NewExfiltrationTracker(),
		anomalyTracker:  NewAnomalyTracker(),
	}

	return td, nil
}

// NewPortScanTracker creates a new port scan tracker
func NewPortScanTracker(window time.Duration, threshold int) *PortScanTracker {
	return &PortScanTracker{
		entries:   make(map[string]*PortScanEntry),
		window:    window,
		threshold: threshold,
	}
}

// NewBeaconTracker creates a new beacon tracker
func NewBeaconTracker() *BeaconTracker {
	return &BeaconTracker{
		entries:  make(map[string]*BeaconEntry),
		window:   5 * time.Minute,
		minScore: 70,
	}
}

// NewExfiltrationTracker creates a new exfiltration tracker
func NewExfiltrationTracker() *ExfiltrationTracker {
	return &ExfiltrationTracker{
		entries:   make(map[string]*ExfilEntry),
		window:    time.Hour,
	}
}

// NewAnomalyTracker creates a new anomaly tracker
func NewAnomalyTracker() *AnomalyTracker {
	return &AnomalyTracker{
		entries:   make(map[string]*AnomalyEntry),
		window:    time.Minute,
		burstList: list.New(),
		maxBurst:  1000,
	}
}

// Detect performs threat detection on a packet
func (td *ThreatDetector) Detect(packet gopacket.Packet, events []*normalization.NetworkEvent, ts time.Time) []ThreatMatch {
	threats := make([]ThreatMatch, 0)

	// Extract IP layer
	var srcIP, dstIP string
	var srcPort, dstPort uint16
	var protocol string

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		srcIP = ip.SrcIP.String()
		dstIP = ip.DstIP.String()
		protocol = ip.Protocol.String()
	} else if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv6)
		srcIP = ip.SrcIP.String()
		dstIP = ip.DstIP.String()
		protocol = ip.NextHeader.String()
	}

	// Extract transport layer ports
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		srcPort = tcp.SrcPort
		dstPort = tcp.DstPort
		protocol = "TCP"
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		srcPort = udp.SrcPort
		dstPort = udp.DstPort
		protocol = "UDP"
	}

	// Anomaly detection
	if td.cfg.Anomaly.Enabled {
		if anomaly := td.anomalyTracker.Check(packet, srcIP, ts); anomaly != nil {
			threats = append(threats, ThreatMatch{
				Type:        ThreatTypeAnomaly,
				Severity:    td.calculateSeverity(anomaly.Score),
				SourceIP:    srcIP,
				DestIP:      dstIP,
				SourcePort:  srcPort,
				DestPort:    dstPort,
				Protocol:    protocol,
				Description: anomaly.Description,
				Score:       anomaly.Score,
				Metadata:    anomaly.Metadata,
				Timestamp:   ts,
			})
		}
	}

	// Port scan detection
	if td.cfg.PortScan.Enabled && srcIP != "" && (srcPort > 0 || dstPort > 0) {
		if threat := td.portScanTracker.Check(srcIP, dstPort, ts); threat != nil {
			threats = append(threats, *threat)
		}
	}

	// C2 beacon detection
	if td.cfg.C2Beacon.Enabled && srcIP != "" && dstIP != "" {
		if threat := td.beaconTracker.Check(srcIP, dstIP, dstPort, protocol, ts); threat != nil {
			threats = append(threats, *threat)
		}
	}

	// Data exfiltration detection
	if td.cfg.Exfiltration.Enabled && srcIP != "" && dstIP != "" {
		var payloadSize uint64
		if appLayer := packet.ApplicationLayer(); appLayer != nil {
			payloadSize = uint64(len(appLayer.Payload()))
		}
		if threat := td.exfilTracker.Check(srcIP, dstIP, dstPort, protocol, payloadSize, ts); threat != nil {
			threats = append(threats, *threat)
		}
	}

	// Update stats
	for _, threat := range threats {
		td.mu.Lock()
		td.stats.TotalThreats++
		switch threat.Type {
		case ThreatTypeAnomaly:
			td.stats.Anomalies++
		case ThreatTypePortScan:
			td.stats.PortScans++
		case ThreatTypeC2Beacon:
			td.stats.C2Beacons++
		case ThreatTypeExfiltration:
			td.stats.Exfiltrations++
		case ThreatTypeBruteForce:
			td.stats.BruteForces++
		}
		td.stats.LastThreatTime = ts
		td.mu.Unlock()
	}

	return threats
}

// CheckPortScan checks for port scan activity
func (pst *PortScanTracker) Check(srcIP string, port uint16, ts time.Time) *ThreatMatch {
	pst.mu.Lock()
	defer pst.mu.Unlock()

	entry, exists := pst.entries[srcIP]
	if !exists {
		entry = &PortScanEntry{
			SourceIP:   srcIP,
			PortSet:    make(map[uint16]bool),
			FirstSeen:  ts,
			LastSeen:   ts,
		}
		pst.entries[srcIP] = entry
	}

	// Add port to the set
	if !entry.PortSet[port] {
		entry.PortSet[port] = true
		entry.UniquePorts++
	}

	entry.LastSeen = ts
	entry.Packets++

	// Check if threshold exceeded
	sensitivityMultiplier := 1.0
	switch pst.threshold {
	case "low":
		sensitivityMultiplier = 2.0
	case "high":
		sensitivityMultiplier = 0.5
	}

	threshold := float64(pst.threshold) * sensitivityMultiplier

	if float64(entry.UniquePorts) > threshold {
		// Calculate score based on port count and speed
		elapsed := entry.LastSeen.Sub(entry.FirstSeen).Seconds()
		speed := float64(entry.UniquePorts) / (elapsed + 1)

		entry.Score = int(math.Min(100, float64(entry.UniquePorts)*2+speed*10))

		return &ThreatMatch{
			Type:        ThreatTypePortScan,
			Severity:    SeverityHigh,
			SourceIP:    srcIP,
			Description: fmt.Sprintf("Port scan detected: %d unique ports in %.1f seconds", entry.UniquePorts, elapsed),
			Score:       entry.Score,
			Metadata: map[string]any{
				"unique_ports": entry.UniquePorts,
				"duration_sec": elapsed,
				"port_list":    pst.getPortList(entry.PortSet),
			},
			Timestamp: ts,
		}
	}

	// Clean old entries
	pst.cleanOldEntries(ts)

	return nil
}

// CheckBeacon checks for beacon-like connection patterns
func (bt *BeaconTracker) Check(srcIP, dstIP string, dstPort uint16, protocol string, ts time.Time) *ThreatMatch {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	key := fmt.Sprintf("%s->%s:%d", srcIP, dstIP, dstPort)
	now := ts

	entry, exists := bt.entries[key]
	if !exists {
		entry = &BeaconEntry{
			SourceIP:    srcIP,
			DestIP:      dstIP,
			DestPort:    dstPort,
			Protocol:    protocol,
			FirstSeen:   now,
			LastSeen:    now,
		}
		bt.entries[key] = entry
	}

	// Calculate interval from last connection
	if !entry.FirstSeen.Equal(entry.LastSeen) {
		interval := now.Sub(entry.LastSeen)
		entry.Intervals = append(entry.Intervals, interval)
		entry.LastInterval = interval

		// Keep only last 50 intervals
		if len(entry.Intervals) > 50 {
			entry.Intervals = entry.Intervals[len(entry.Intervals)-50:]
		}
	}

	entry.LastSeen = now
	entry.RequestCount++

	// Analyze beacon pattern if we have enough intervals
	if len(entry.Intervals) >= 10 {
		avgInterval, stdDev, jitter := bt.analyzeIntervals(entry.Intervals)

		entry.AvgInterval = avgInterval
		entry.StdDev = stdDev
		entry.Jitter = jitter

		// Calculate beacon score
		entry.Score = bt.calculateBeaconScore(entry, jitter, stdDev)

		// Check if beacon-like
		if entry.Score >= bt.minScore {
			return &ThreatMatch{
				Type:        ThreatTypeC2Beacon,
				Severity:    SeverityCritical,
				SourceIP:    srcIP,
				DestIP:      dstIP,
				DestPort:    dstPort,
				Protocol:    protocol,
				Description: fmt.Sprintf("C2 beacon pattern detected: avg interval %.2fs, jitter %.1f%%", avgInterval.Seconds(), jitter*100),
				Score:       entry.Score,
				Metadata: map[string]any{
					"avg_interval_sec": avgInterval.Seconds(),
					"std_dev_sec":      stdDev.Seconds(),
					"jitter_percent":   jitter * 100,
					"request_count":    entry.RequestCount,
				},
				Timestamp: now,
			}
		}
	}

	// Clean old entries
	bt.cleanOldEntries(now)

	return nil
}

// CheckExfiltration checks for data exfiltration patterns
func (et *ExfiltrationTracker) Check(srcIP, dstIP string, dstPort uint16, protocol string, bytes uint64, ts time.Time) *ThreatMatch {
	et.mu.Lock()
	defer et.mu.Unlock()

	key := fmt.Sprintf("%s->%s:%d", srcIP, dstIP, dstPort)
	now := ts

	entry, exists := et.entries[key]
	if !exists {
		entry = &ExfilEntry{
			SourceIP:   srcIP,
			DestIP:     dstIP,
			DestPort:   dstPort,
			Protocol:   protocol,
			StartTime:  now,
			LastTime:   now,
		}
		et.entries[key] = entry
	}

	entry.BytesSent += bytes
	entry.LastTime = now

	// Calculate rate
	elapsed := entry.LastTime.Sub(entry.StartTime).Seconds()
	if elapsed > 0 {
		entry.Rate = float64(entry.BytesSent) / elapsed
	}

	// Check thresholds
	rateExceeded := entry.Rate > float64(et.rateThreshold)
	volumeExceeded := entry.BytesSent > uint64(et.volumeThreshold)

	if rateExceeded || volumeExceeded {
		entry.Score = int(math.Min(100, (entry.Rate/float64(et.rateThreshold))*50 + (float64(entry.BytesSent)/float64(et.volumeThreshold))*50))

		return &ThreatMatch{
			Type:        ThreatTypeExfiltration,
			Severity:    SeverityHigh,
			SourceIP:    srcIP,
			DestIP:      dstIP,
			DestPort:    dstPort,
			Protocol:    protocol,
			Description: fmt.Sprintf("Data exfiltration detected: %.2f MB at %.2f KB/s", float64(entry.BytesSent)/(1024*1024), entry.Rate/1024),
			Score:       entry.Score,
			Metadata: map[string]any{
				"total_bytes":     entry.BytesSent,
				"rate_bps":        entry.Rate,
				"duration_sec":    elapsed,
			},
			Timestamp: now,
		}
	}

	// Clean old entries
	et.cleanOldEntries(now)

	return nil
}

// Check checks for packet anomalies
func (at *AnomalyTracker) Check(packet gopacket.Packet, srcIP string, ts time.Time) *AnomalyEntry {
	at.mu.Lock()
	defer at.mu.Unlock()

	entry, exists := at.entries[srcIP]
	if !exists {
		entry = &AnomalyEntry{
			IP:        srcIP,
			FirstSeen: ts,
		}
		at.entries[srcIP] = entry
	}

	packetSize := len(packet.Data())
	entry.PacketSizes = append(entry.PacketSizes, packetSize)
	entry.TotalPackets++
	entry.LastSeen = ts

	// Keep only last 100 packet sizes
	if len(entry.PacketSizes) > 100 {
		entry.PacketSizes = entry.PacketSizes[len(entry.PacketSizes)-100:]
	}

	// Check for anomalies
	anomaly := at.detectAnomaly(entry, packetSize)
	if anomaly != nil {
		entry.Score = int(math.Min(100, float64(entry.TotalPackets)/10+float64(anomaly.Score)))
	}

	// Clean old entries
	at.cleanOldEntries(ts)

	return anomaly
}

// analyzeIntervals analyzes connection intervals for beacon patterns
func (bt *BeaconTracker) analyzeIntervals(intervals []time.Duration) (avg, stdDev, jitter time.Duration) {
	if len(intervals) == 0 {
		return 0, 0, 0
	}

	var sum time.Duration
	for _, interval := range intervals {
		sum += interval
	}
	avg = sum / time.Duration(len(intervals))

	var varianceSum float64
	for _, interval := range intervals {
		diff := float64(interval - avg)
		varianceSum += diff * diff
	}
	stdDev = time.Duration(math.Sqrt(varianceSum / float64(len(intervals))))

	// Calculate jitter (deviation from average)
	if avg > 0 {
		jitter = time.Duration(math.Abs(float64(stdDev) / float64(avg)))
	}

	return avg, stdDev, jitter
}

// calculateBeaconScore calculates the beacon likelihood score
func (bt *BeaconTracker) calculateBeaconScore(entry *BeaconEntry, jitter, stdDev time.Duration) int {
	score := 0

	// Low jitter is suspicious
	if entry.Jitter < 0.1 {
		score += 40
	} else if entry.Jitter < 0.2 {
		score += 30
	} else if entry.Jitter < 0.3 {
		score += 20
	}

	// Regular intervals are suspicious
	if entry.StdDev < 100*time.Millisecond {
		score += 30
	} else if entry.StdDev < 500*time.Millisecond {
		score += 20
	}

	// High request count with regular intervals
	if entry.RequestCount > 100 {
		score += 20
	} else if entry.RequestCount > 50 {
		score += 10
	}

	return score
}

// detectAnomaly detects packet anomalies
func (at *AnomalyTracker) detectAnomaly(entry *AnomalyEntry, packetSize int) *AnomalyEntry {
	if len(entry.PacketSizes) < 10 {
		return nil
	}

	// Calculate statistics
	var sum int
	for _, size := range entry.PacketSizes {
		sum += size
	}
	avg := float64(sum) / float64(len(entry.PacketSizes))

	// Check for oversized packets
	if packetSize > 65535 {
		return entry
	}

	// Check for unusual variance
	var variance float64
	for _, size := range entry.PacketSizes {
		diff := float64(size) - avg
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(entry.PacketSizes)))

	// Detect burst
	if packetSize > avg+3*stdDev {
		entry.BurstCount++
		if entry.BurstCount > 5 {
			entry.Score = 80
			return entry
		}
	}

	return nil
}

// Helper functions
func (pst *PortScanTracker) getPortList(ports map[uint16]bool) []uint16 {
	portList := make([]uint16, 0, len(ports))
	for port := range ports {
		portList = append(portList, port)
	}
	return portList
}

func (pst *PortScanTracker) cleanOldEntries(ts time.Time) {
	cutoff := ts.Add(-pst.window * 2)
	for ip, entry := range pst.entries {
		if entry.LastSeen.Before(cutoff) {
			delete(pst.entries, ip)
		}
	}
}

func (bt *BeaconTracker) cleanOldEntries(ts time.Time) {
	cutoff := ts.Add(-bt.window * 2)
	for key, entry := range bt.entries {
		if entry.LastSeen.Before(cutoff) {
			delete(bt.entries, key)
		}
	}
}

func (et *ExfiltrationTracker) cleanOldEntries(ts time.Time) {
	cutoff := ts.Add(-et.window * 2)
	for key, entry := range et.entries {
		if entry.LastTime.Before(cutoff) {
			delete(et.entries, key)
		}
	}
}

func (at *AnomalyTracker) cleanOldEntries(ts time.Time) {
	cutoff := ts.Add(-at.window * 2)
	for ip, entry := range at.entries {
		if entry.LastSeen.Before(cutoff) {
			delete(at.entries, ip)
		}
	}
}

func (td *ThreatDetector) calculateSeverity(score int) Severity {
	switch {
	case score >= 90:
		return SeverityCritical
	case score >= 70:
		return SeverityHigh
	case score >= 50:
		return SeverityMedium
	case score >= 30:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

// GetStats returns threat detection statistics
func (td *ThreatDetector) GetStats() ThreatStats {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.stats
}

// GetPortScanStats returns port scan tracker statistics
func (td *ThreatDetector) GetPortScanStats() map[string]int {
	td.portScanTracker.mu.RLock()
	defer td.portScanTracker.mu.RUnlock()

	stats := make(map[string]int)
	for ip, entry := range td.portScanTracker.entries {
		stats[ip] = entry.UniquePorts
	}
	return stats
}

// GetBeaconStats returns beacon tracker statistics
func (td *ThreatDetector) GetBeaconStats() map[string]int {
	td.beaconTracker.mu.RLock()
	defer td.beaconTracker.mu.RUnlock()

	stats := make(map[string]int)
	for key, entry := range td.beaconTracker.entries {
		stats[key] = entry.Score
	}
	return stats
}

// IsPrivateIP checks if an IP is a private/internal IP
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check RFC 1918 private ranges
	privateRanges := []net.IPNet{
		*mustParseCIDR("10.0.0.0/8"),
		*mustParseCIDR("172.16.0.0/12"),
		*mustParseCIDR("192.168.0.0/16"),
	}

	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}

	// Check localhost
	return ip.IsLoopback()
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}

// Reset resets all trackers
func (td *ThreatDetector) Reset() {
	td.portScanTracker.mu.Lock()
	td.portScanTracker.entries = make(map[string]*PortScanEntry)
	td.portScanTracker.mu.Unlock()

	td.beaconTracker.mu.Lock()
	td.beaconTracker.entries = make(map[string]*BeaconEntry)
	td.beaconTracker.mu.Unlock()

	td.exfilTracker.mu.Lock()
	td.exfilTracker.entries = make(map[string]*ExfilEntry)
	td.exfilTracker.mu.Unlock()

	td.anomalyTracker.mu.Lock()
	td.anomalyTracker.entries = make(map[string]*AnomalyEntry)
	td.anomalyTracker.mu.Unlock()

	atomic.StoreUint64(&td.stats.TotalThreats, 0)
}
