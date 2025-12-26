// Package dpi provides Deep Packet Inspection capabilities for network traffic analysis.
// This package implements packet parsing, protocol detection, and threat identification
// for the SGE Network Sensor.
package dpi

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atailh4n/sakin/internal/config"
	"github.com/atailh4n/sakin/internal/normalization"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// PacketInspector performs deep packet inspection on network traffic
type PacketInspector struct {
	cfg              *config.DPIConfig
	parser           *protocolParser
	threatDetector   *ThreatDetector
	workerPool       chan *packetJob
	wg               sync.WaitGroup
	processedCount   atomic.Uint64
	droppedCount     atomic.Uint64
	startedAt        time.Time
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.RWMutex
	stats            InspectorStats
}

// InspectorStats holds statistics for the packet inspector
type InspectorStats struct {
	ProcessedPackets uint64            `json:"processed_packets"`
	DroppedPackets   uint64            `json:"dropped_packets"`
	BytesProcessed   uint64            `json:"bytes_processed"`
	ProtocolStats    map[string]uint64 `json:"protocol_stats"`
	ThreatStats      map[string]uint64 `json:"threat_stats"`
}

// packetJob represents a packet to be processed
type packetJob struct {
	packet   gopacket.Packet
	iface    string
	timestamp time.Time
}

// InspectionResult contains the results of packet inspection
type InspectionResult struct {
	Events         []*normalization.NetworkEvent
	Threats        []ThreatMatch
	Metadata       InspectionMetadata
	ProcessingTime time.Duration
}

// InspectionMetadata contains metadata about the inspection
type InspectionMetadata struct {
	PacketSize    int
	NumLayers     int
	LinkType      layers.LinkType
	Timestamp     time.Time
	Interface     string
}

// NewPacketInspector creates a new packet inspector with the given configuration
func NewPacketInspector(cfg *config.DPIConfig, threatCfg *config.ThreatDetectConfig) (*PacketInspector, error) {
	ctx, cancel := context.WithCancel(context.Background())

	parser, err := newProtocolParser(cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	threatDetector, err := NewThreatDetector(threatCfg)
	if err != nil {
		cancel()
		return nil, err
	}

	return &PacketInspector{
		cfg:            cfg,
		parser:         parser,
		threatDetector: threatDetector,
		workerPool:     make(chan *packetJob, 10000),
		ctx:            ctx,
		cancel:         cancel,
		stats: InspectorStats{
			ProtocolStats: make(map[string]uint64),
			ThreatStats:   make(map[string]uint64),
		},
		startedAt: time.Now(),
	}, nil
}

// Start starts the packet inspector worker pool
func (pi *PacketInspector) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		pi.wg.Add(1)
		go pi.worker(i)
	}
}

// Stop gracefully stops the packet inspector
func (pi *PacketInspector) Stop() {
	pi.cancel()
	close(pi.workerPool)
	pi.wg.Wait()
}

// ProcessPacket submits a packet for processing
func (pi *PacketInspector) ProcessPacket(packet gopacket.Packet, iface string) {
	select {
	case pi.workerPool <- &packetJob{
		packet:   packet,
		iface:    iface,
		timestamp: time.Now(),
	}:
	default:
		pi.droppedCount.Add(1)
	}
}

// ProcessPacketAsync submits a packet for processing asynchronously
func (pi *PacketInspector) ProcessPacketAsync(packet gopacket.Packet, iface string) {
	go func() {
		select {
		case pi.workerPool <- &packetJob{
			packet:   packet,
			iface:    iface,
			timestamp: time.Now(),
		}:
		default:
			pi.droppedCount.Add(1)
		}
	}()
}

// worker is the worker function for processing packets
func (pi *PacketInspector) worker(id int) {
	defer pi.wg.Done()

	for {
		select {
		case <-pi.ctx.Done():
			return
		case job, ok := <-pi.workerPool:
			if !ok {
				return
			}
			pi.processPacket(job)
		}
	}
}

// processPacket processes a single packet
func (pi *PacketInspector) processPacket(job *packetJob) {
	start := time.Now()

	result := pi.inspectPacket(job.packet, job.iface, job.timestamp)

	// Update stats
	pi.processedCount.Add(1)
	pi.stats.ProcessedPackets++
	pi.stats.BytesProcessed += uint64(job.packet.Metadata().Length)

	// Update protocol stats
	for _, event := range result.Events {
		pi.mu.Lock()
		pi.stats.ProtocolStats[event.Protocol]++
		pi.mu.Unlock()
	}

	// Update threat stats
	for _, threat := range result.Threats {
		pi.mu.Lock()
		pi.stats.ThreatStats[threat.Type]++
		pi.mu.Unlock()
	}

	_ = result.ProcessingTime // Can be used for latency tracking
	_ = start
}

// inspectPacket performs deep inspection on a single packet
func (pi *PacketInspector) inspectPacket(packet gopacket.Packet, iface string, ts time.Time) *InspectionResult {
	result := &InspectionResult{
		Events:    make([]*normalization.NetworkEvent, 0),
		Threats:   make([]ThreatMatch, 0),
		Metadata: InspectionMetadata{
			PacketSize: len(packet.Data()),
			NumLayers:  len(packet.Layers()),
			Timestamp:  ts,
			Interface:  iface,
		},
	}

	// Set link type if available
	if linkLayer := packet.LinkLayer(); linkLayer != nil {
		result.Metadata.LinkType = linkLayer.LayerType()
	}

	// Parse layers and extract events
	events := pi.parser.parsePacket(packet, iface, ts)
	result.Events = append(result.Events, events...)

	// Perform threat detection if enabled
	if pi.threatDetector != nil && pi.threatDetector.cfg.Enabled {
		threats := pi.threatDetector.Detect(packet, events, ts)
		result.Threats = threats
	}

	// Calculate processing time
	result.ProcessingTime = time.Since(ts)

	return result
}

// GetStats returns current inspector statistics
func (pi *PacketInspector) GetStats() InspectorStats {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	stats := pi.stats
	stats.ProcessedPackets = pi.processedCount.Load()
	stats.DroppedPackets = pi.droppedCount.Load()
	return stats
}

// ResetStats resets the inspector statistics
func (pi *PacketInspector) ResetStats() {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.processedCount.Store(0)
	pi.droppedCount.Store(0)
	pi.stats.ProcessedPackets = 0
	pi.stats.DroppedPackets = 0
	pi.stats.BytesProcessed = 0
	pi.stats.ProtocolStats = make(map[string]uint64)
	pi.stats.ThreatStats = make(map[string]uint64)
}

// GetThreatDetector returns the threat detector instance
func (pi *PacketInspector) GetThreatDetector() *ThreatDetector {
	return pi.threatDetector
}

// GetParser returns the protocol parser instance
func (pi *PacketInspector) GetParser() *protocolParser {
	return pi.parser
}

// CaptureType represents the type of packet capture interface
type CaptureType string

const (
	CaptureTypePCAP    CaptureType = "pcap"
	CaptureTypeAFPacket CaptureType = "af_packet"
	CaptureTypeDpdk    CaptureType = "dpdk"
)

// CaptureHandle represents an interface to capture packets
type CaptureHandle interface {
	// Open opens a capture interface
	Open(iface string, cfg *config.InterfaceConfig) error
	// Close closes the capture handle
	Close() error
	// ReadPacket reads a single packet
	ReadPacket() (gopacket.Packet, error)
	// Stats returns capture statistics
	Stats() (uint64, uint64, error)
	// SetBPFFilter sets a BPF filter
	SetBPFFilter(filter string) error
}

// NewCaptureHandle creates a capture handle for the given type
func NewCaptureHandle(captureType CaptureType) (CaptureHandle, error) {
	switch captureType {
	case CaptureTypePCAP:
		return newPcapCaptureHandle(), nil
	case CaptureTypeAFPacket:
		return newAFPacketCaptureHandle(), nil
	default:
		return nil, &UnsupportedCaptureTypeError{CaptureType: captureType}
	}
}

// UnsupportedCaptureTypeError is returned when an unsupported capture type is requested
type UnsupportedCaptureTypeError struct {
	CaptureType CaptureType
}

func (e *UnsupportedCaptureTypeError) Error() string {
	return string(e.CaptureType) + " is not a supported capture type"
}
