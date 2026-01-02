package inspector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"sakin-go/cmd/sge-network-sensor/config"
	"sakin-go/cmd/sge-network-sensor/dpi"
)

// Inspector manages packet capture across interfaces.
type Inspector struct {
	config    *config.AppConfig
	eventChan chan<- interface{} // Channel to send detected events
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NetworkEvent represents a captured network event (simplified).
type NetworkEvent struct {
	Timestamp   time.Time
	SrcIP       string
	DstIP       string
	SrcPort     uint16
	DstPort     uint16
	Protocol    string
	PayloadSize int
	SNI         string // HTTPS
	HTTPHost    string // HTTP
}

// NewInspector creates a new inspector instance.
func NewInspector(cfg *config.AppConfig, eventChan chan<- interface{}) *Inspector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Inspector{
		config:    cfg,
		eventChan: eventChan,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins capturing on configured interfaces.
func (i *Inspector) Start() error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	for _, device := range devices {
		// Filter interfaces logic
		if i.shouldIgnoreInterface(device.Name) {
			continue
		}

		// Only capture on specific interface if config demands
		if i.config.Interface != "all" && i.config.Interface != "any" {
			if !strings.Contains(i.config.Interface, device.Name) {
				continue
			}
		}

		i.wg.Add(1)
		go i.captureLoop(device.Name)
	}

	return nil
}

// Stop halts all capture routines.
func (i *Inspector) Stop() {
	i.cancel()
	i.wg.Wait()
}

func (i *Inspector) shouldIgnoreInterface(name string) bool {
	// Simple filters for cleaner logs (can be expanded)
	lower := strings.ToLower(name)
	if strings.Contains(lower, "loopback") || strings.Contains(lower, "docker") {
		return true
	}
	return false
}

func (i *Inspector) captureLoop(iface string) {
	defer i.wg.Done()
	log.Printf("[Inspector] Starting capture on %s", iface)

	handle, err := pcap.OpenLive(iface, i.config.SnapLen, i.config.PromiscuousMode, pcap.BlockForever)
	if err != nil {
		log.Printf("[Inspector] Error opening %s: %v", iface, err)
		return
	}
	defer handle.Close()

	if i.config.BPFFilter != "" {
		if err := handle.SetBPFFilter(i.config.BPFFilter); err != nil {
			log.Printf("[Inspector] Failed to set BPF on %s: %v", iface, err)
		}
	}

	// Create layer parsers once to reuse
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP
	var payload gopacket.Payload

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth, &ip4, &ip6, &tcp, &udp, &payload,
	)

	decoded := []gopacket.LayerType{}

	for {
		select {
		case <-i.ctx.Done():
			return
		default:
			// Read packet
			data, _, err := handle.ReadPacketData()
			if err != nil {
				continue
			}

			err = parser.DecodeLayers(data, &decoded)
			if err != nil {
				// Continue even if full decode fails, as long as we got some layers
			}

			// Process Decoded Layers
			evt := NetworkEvent{Timestamp: time.Now()}
			hasIP := false

			for _, layerType := range decoded {
				switch layerType {
				case layers.LayerTypeIPv4:
					evt.SrcIP = ip4.SrcIP.String()
					evt.DstIP = ip4.DstIP.String()
					evt.Protocol = ip4.Protocol.String()
					hasIP = true
				case layers.LayerTypeIPv6:
					evt.SrcIP = ip6.SrcIP.String()
					evt.DstIP = ip6.DstIP.String()
					evt.Protocol = ip6.NextHeader.String()
					hasIP = true
				case layers.LayerTypeTCP:
					evt.SrcPort = uint16(tcp.SrcPort)
					evt.DstPort = uint16(tcp.DstPort)
					evt.PayloadSize = len(tcp.Payload)

					// DPI Checks
					if len(tcp.Payload) > 0 {
						if sni, ok := dpi.ParseTLSClientHello(tcp.Payload); ok {
							evt.SNI = sni.ServerName
						} else if http, ok := dpi.ParseHTTPRequest(tcp.Payload); ok {
							evt.HTTPHost = http.Host
						}
					}
				case layers.LayerTypeUDP:
					evt.SrcPort = uint16(udp.SrcPort)
					evt.DstPort = uint16(udp.DstPort)
					evt.PayloadSize = len(udp.Payload)
				}
			}

			if hasIP {
				// If ports are 0 (e.g. ICMP), they stay 0 which is fine
				// Non-blocking send to avoid stalling capture loop
				select {
				case i.eventChan <- evt:
				default:
					// Drop if channel full
				}
			}
		}
	}
}
