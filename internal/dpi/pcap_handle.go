package dpi

import (
	"errors"
	"fmt"
	"time"

	"github.com/atailh4n/sakin/internal/config"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type pcapCaptureHandle struct {
	handle   *pcap.Handle
	iface    string
	stats    pcap.Stats
	cfg      *config.InterfaceConfig
}

// pcap-specific statistics
type pcapStats struct {
	PacketsReceived  uint64
	PacketsDropped   uint64
	InterfaceDropped uint64
}

func newPcapCaptureHandle() *pcapCaptureHandle {
	return &pcapCaptureHandle{}
}

// Open opens a network interface for packet capture using pcap
func (ph *pcapCaptureHandle) Open(iface string, cfg *config.InterfaceConfig) error {
	ph.iface = iface
	ph.cfg = cfg

	// Set snaplen with default
	snaplen := cfg.Snaplen
	if snaplen <= 0 {
		snaplen = 1600
	}

	// Build pcap flags
	flags := pcap.BlockForever
	if cfg.Promiscuous {
		flags |= pcap.Promisc
	}

	// Open the interface
	handle, err := pcap.OpenLive(iface, int32(snaplen), cfg.Promiscuous, time.Duration(cfg.Timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("failed to open interface %s: %w", iface, err)
	}

	// Apply BPF filter if specified
	if cfg.BPFFilter != "" {
		if err := handle.SetBPFFilter(cfg.BPFFilter); err != nil {
			handle.Close()
			return fmt.Errorf("failed to set BPF filter: %w", err)
		}
	}

	ph.handle = handle
	return nil
}

// Close closes the capture handle
func (ph *pcapCaptureHandle) Close() error {
	if ph.handle != nil {
		return ph.handle.Close()
	}
	return nil
}

// ReadPacket reads a single packet from the interface
func (ph *pcapCaptureHandle) ReadPacket() (gopacket.Packet, error) {
	if ph.handle == nil {
		return nil, errors.New("capture handle not opened")
	}

	data, ci, err := ph.handle.ReadPacketData()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	return gopacket.NewPacket(data, ph.handle.LinkType(), gopacket.DecodeOptions{
		DecodeStreamsAsDatagrams: true,
		NoCopy:                   false,
		SkipDecodeRecovery:       true,
	}).ApplyMetadata(ci), nil
}

// Stats returns capture statistics
func (ph *pcapCaptureHandle) Stats() (uint64, uint64, error) {
	if ph.handle == nil {
		return 0, 0, errors.New("capture handle not opened")
	}

	stats, err := ph.handle.Stats()
	if err != nil {
		return 0, 0, err
	}

	return uint64(stats.PacketsReceived), uint64(stats.PacketsDropped), nil
}

// SetBPFFilter sets a BPF filter on the capture handle
func (ph *pcapCaptureHandle) SetBPFFilter(filter string) error {
	if ph.handle == nil {
		return errors.New("capture handle not opened")
	}
	return ph.handle.SetBPFFilter(filter)
}

// GetPcapHandle returns the underlying pcap handle for advanced usage
func (ph *pcapCaptureHandle) GetPcapHandle() *pcap.Handle {
	return ph.handle
}

// LinkType returns the link type of the capture interface
func (ph *pcapCaptureHandle) LinkType() layers.LinkType {
	if ph.handle == nil {
		return layers.LinkTypeNull
	}
	return ph.handle.LinkType()
}

// Direction determines which direction(s) to capture
type pcapDirection int

const (
	DirectionIn    pcapDirection = 0x01
	DirectionOut   pcapDirection = 0x02
	DirectionInOut pcapDirection = DirectionIn | DirectionOut
)

// SetDirection sets the capture direction
func (ph *pcapCaptureHandle) SetDirection(dir pcapDirection) error {
	if ph.handle == nil {
		return errors.New("capture handle not opened")
	}

	var direction pcap.Direction
	switch dir {
	case DirectionIn:
		direction = pcap.DirectionIn
	case DirectionOut:
		direction = pcap.DirectionOut
	case DirectionInOut:
		direction = pcap.DirectionInOut
	default:
		direction = pcap.DirectionInOut
	}

	return ph.handle.SetDirection(direction)
}

// SetBufferSize sets the capture buffer size in bytes
func (ph *pcapCaptureHandle) SetBufferSize(size int32) error {
	if ph.handle == nil {
		return errors.New("capture handle not opened")
	}
	return ph.handle.SetBufferSize(size)
}

// SetTimeout sets the read timeout
func (ph *pcapCaptureHandle) SetTimeout(timeout time.Duration) error {
	if ph.handle == nil {
		return errors.New("capture handle not opened")
	}
	return ph.handle.SetTimeout(timeout)
}

// IsOpen returns whether the handle is open
func (ph *pcapCaptureHandle) IsOpen() bool {
	return ph.handle != nil
}
