// Package main is the entry point for the SGE Network Sensor
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/atailh4n/sakin/internal/config"
	"github.com/atailh4n/sakin/internal/dpi"
	"github.com/atailh4n/sakin/internal/normalization"
	"github.com/atailh4n/sakin/internal/storage"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	preset := flag.String("preset", "", "Configuration preset (light, standard, aggressive)")
	workers := flag.Int("workers", 0, "Number of worker threads (0 = auto)")
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("SGE Network Sensor v%s (commit: %s, date: %s)\n", version, commit, date)
		fmt.Printf("Go version: %s\n", runtime.Version())
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Starting SGE Network Sensor v%s", version)

	// Load configuration
	cfg, err := loadConfiguration(*configPath, *preset)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded (preset: %s, instance: %s)", *preset, cfg.InstanceID)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create event normalizer
	normalizer := normalization.NewEventNormalizer(normalization.DefaultNormalizerConfig())
	normalizer.SetInstanceID(cfg.InstanceID)

	// Create packet inspector
	dpiConfig := &config.DPIConfig{
		HTTPEnabled:         cfg.DPI.HTTPEnabled,
		DNSEnabled:          cfg.DPI.DNSEnabled,
		TLSEnabled:          cfg.DPI.TLSEnabled,
		SMBEnabled:          cfg.DPI.SMBEnabled,
		MaxPayloadBytes:     cfg.DPI.MaxPayloadBytes,
		StreamReassembly:    cfg.DPI.StreamReassembly,
	}

	inspector, err := dpi.NewPacketInspector(dpiConfig, &cfg.ThreatDetect)
	if err != nil {
		log.Fatalf("Failed to create packet inspector: %v", err)
	}

	// Determine worker count
	workerCount := *workers
	if workerCount <= 0 {
		workerCount = cfg.Resources.WorkerPoolSize
		if workerCount <= 0 {
			workerCount = runtime.NumCPU()
		}
	}
	log.Printf("Starting %d worker threads", workerCount)
	inspector.Start(workerCount)

	// Create NATS producer
	var producer *storage.NATSProducer
	if cfg.Output.NATS.Enabled {
		producer, err = storage.NewNATSProducer(&cfg.Output.NATS)
		if err != nil {
			log.Printf("Warning: Failed to create NATS producer: %v", err)
			log.Printf("Continuing without NATS - events will not be sent to central storage")
		} else {
			defer producer.Stop()
			producer.Start(cfg.Resources.BatchSize, cfg.Resources.BatchFlushInterval)
			log.Printf("NATS producer started (subject: %s)", cfg.Output.NATS.Subject)
		}
	}

	// Set up packet processing
	eventCh := make(chan *normalization.NetworkEvent, 10000)
	var wg sync.WaitGroup

	// Start event normalizer and publisher
	wg.Add(1)
	go func() {
		defer wg.Done()
		processEvents(ctx, normalizer, producer, eventCh)
	}()

	// Start capture on configured interfaces
	captureWg := sync.WaitGroup{}
	interfaces, err := getInterfaces(cfg.Interfaces.Interfaces)
	if err != nil {
		log.Printf("Warning: Failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		captureWg.Add(1)
		go func(ifaceName string) {
			defer captureWg.Done()
			captureTraffic(ctx, ifaceName, cfg.Interfaces, inspector, eventCh)
		}(iface)
	}

	log.Printf("Capturing traffic on %d interface(s)", len(interfaces))

	// Wait for shutdown signal
	sig := <-sigCh
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Cancel context to stop workers
	cancel()

	// Stop inspector
	inspector.Stop()

	// Wait for capture to finish
	captureWg.Wait()

	// Close event channel
	close(eventCh)

	// Wait for event processing to complete
	wg.Wait()

	// Print final stats
	stats := inspector.GetStats()
	log.Printf("Final statistics:")
	log.Printf("  Processed packets: %d", stats.ProcessedPackets)
	log.Printf("  Dropped packets: %d", stats.DroppedPackets)
	log.Printf("  Bytes processed: %d", stats.BytesProcessed)

	log.Printf("SGE Network Sensor shutdown complete")
}

// loadConfiguration loads configuration from file or preset
func loadConfiguration(configPath, preset string) (*config.NetworkSensorConfig, error) {
	var cfg *config.NetworkSensorConfig
	var err error

	if preset != "" {
		cfg, err = config.Preset(preset)
		if err != nil {
			return nil, fmt.Errorf("failed to load preset %s: %w", preset, err)
		}
		log.Printf("Loaded configuration preset: %s", preset)
	} else if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration from %s: %w", configPath, err)
		}
		log.Printf("Loaded configuration from: %s", configPath)
	} else {
		// Try to load default config
		cfg, err = config.Load("")
		if err != nil {
			// Use standard preset as default
			cfg, err = config.Preset("standard")
			if err != nil {
				return nil, fmt.Errorf("failed to load default configuration: %w", err)
			}
		}
		log.Printf("Using default configuration")
	}

	return cfg, nil
}

// getInterfaces returns the list of network interfaces to capture from
func getInterfaces(configuredInterfaces []string) ([]string, error) {
	if len(configuredInterfaces) > 0 {
		return configuredInterfaces, nil
	}

	// Get all interfaces
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	interfaces := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		// Skip loopback
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}
		interfaces = append(interfaces, iface.Name)
	}

	return interfaces, nil
}

// captureTraffic captures traffic on a specific interface
func captureTraffic(ctx context.Context, ifaceName string, ifaceConfig config.InterfaceConfig, inspector *dpi.PacketInspector, eventCh chan<- *normalization.NetworkEvent) {
	log.Printf("Starting capture on interface: %s", ifaceName)

	// Set defaults
	snaplen := ifaceConfig.Snaplen
	if snaplen <= 0 {
		snaplen = 1600
	}

	timeout := time.Duration(ifaceConfig.Timeout)
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	// Open interface
	handle, err := pcap.OpenLive(ifaceName, int32(snaplen), ifaceConfig.Promiscuous, timeout)
	if err != nil {
		log.Printf("Error opening interface %s: %v", ifaceName, err)
		return
	}
	defer handle.Close()

	// Set BPF filter if configured
	if ifaceConfig.BPFFilter != "" {
		if err := handle.SetBPFFilter(ifaceConfig.BPFFilter); err != nil {
			log.Printf("Warning: Failed to set BPF filter on %s: %v", ifaceName, err)
		} else {
			log.Printf("BPF filter set on %s: %s", ifaceName, ifaceConfig.BPFFilter)
		}
	}

	// Create packet source
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// Capture loop
	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping capture on interface: %s", ifaceName)
			return
		case packet, ok := <-packetSource.Packets():
			if !ok {
				log.Printf("Packet source closed for interface: %s", ifaceName)
				return
			}

			// Process packet
			inspector.ProcessPacket(packet, ifaceName)
		}
	}
}

// processEvents processes normalized events and sends them to NATS
func processEvents(ctx context.Context, normalizer *normalization.EventNormalizer, producer *storage.NATSProducer, eventCh <-chan *normalization.NetworkEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventCh:
			if !ok {
				return
			}

			// Normalize event
			normalized, err := normalizer.Normalize(event)
			if err != nil {
				log.Printf("Warning: Failed to normalize event: %v", err)
				continue
			}

			// Send to NATS if available
			if producer != nil {
				if err := producer.Publish(normalized); err != nil {
					log.Printf("Warning: Failed to publish event: %v", err)
				}
			}
		}
	}
}
