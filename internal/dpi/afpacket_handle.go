//go:build linux
// +build linux

package dpi

import (
    "errors"
    "fmt"
    "os"
    "syscall"
    "time"
    "unsafe"

    "github.com/atailh4n/sakin/internal/config"
    "github.com/google/gopacket"
    "github.com/google/gopacket/afpacket"
    "github.com/google/gopacket/layers"
)

// afPacketCaptureHandle uses AF_PACKET for zero-copy packet capture on Linux
type afPacketCaptureHandle struct {
    handle   *afpacket.TPacket
    iface    string
    cfg      *config.InterfaceConfig
}

func newAFPacketCaptureHandle() *afPacketCaptureHandle {
    return &afPacketCaptureHandle{}
}

// Open opens a network interface for packet capture using AF_PACKET
func (aph *afPacketCaptureHandle) Open(iface string, cfg *config.InterfaceConfig) error {
    aph.iface = iface
    aph.cfg = cfg

    // Build options for AF_PACKET
    options := []afpacket.Option{
        afpacket.Device(iface),
        afpacket.Snaplen(int(cfg.Snaplen)),
        afpacket.Promiscuous(cfg.Promiscuous),
        afpacket.BufferSize(int(cfg.BufferSize)),
    }

    // Set timeout
    options = append(options, afpacket.Timeout(time.Duration(cfg.Timeout)*time.Second))

    // Open the AF_PACKET socket
    handle, err := afpacket.NewTPacket(
        afpacket.TPacketVersion3,
        options...,
    )
    if err != nil {
        // Fall back to AF_PACKET v1 if v3 is not available
        handle, err = afpacket.NewTPacket(
            afpacket.TPacketVersion1,
            options...,
        )
        if err != nil {
            return fmt.Errorf("failed to open AF_PACKET interface %s: %w", iface, err)
        }
    }

    // Apply BPF filter if specified
    if cfg.BPFFilter != "" {
        if err := handle.SetBPFFilter(cfg.BPFFilter); err != nil {
            handle.Close()
            return fmt.Errorf("failed to set BPF filter: %w", err)
        }
    }

    aph.handle = handle
    return nil
}

// Close closes the capture handle
func (aph *afPacketCaptureHandle) Close() error {
    if aph.handle != nil {
        return aph.handle.Close()
    }
    return nil
}

// ReadPacket reads a single packet from the interface
func (aph *afPacketCaptureHandle) ReadPacket() (gopacket.Packet, error) {
    if aph.handle == nil {
        return nil, errors.New("AF_PACKET handle not opened")
    }

    data, ci, err := aph.handle.ReadPacketData()
    if err != nil {
        return nil, fmt.Errorf("failed to read packet: %w", err)
    }

    return gopacket.NewPacket(data, layers.LinkTypeEthernet, gopacket.DecodeOptions{
        DecodeStreamsAsDatagrams: true,
        NoCopy:                   false,
        SkipDecodeRecovery:       true,
    }).ApplyMetadata(ci), nil
}

// Stats returns capture statistics
func (aph *afPacketCaptureHandle) Stats() (uint64, uint64, error) {
    if aph.handle == nil {
        return 0, 0, errors.New("AF_PACKET handle not opened")
    }

    // Get AF_PACKET stats via socket option
    stats := aph.handle.SocketStats()
    return uint64(stats.Packets), uint64(stats.Drops), nil
}

// SetBPFFilter sets a BPF filter on the capture handle
func (aph *afPacketCaptureHandle) SetBPFFilter(filter string) error {
    if aph.handle == nil {
        return errors.New("AF_PACKET handle not opened")
    }
    return aph.handle.SetBPFFilter(filter)
}

// GetStats returns extended statistics
func (aph *afPacketCaptureHandle) GetStats() (*afpacket.SocketStats, error) {
    if aph.handle == nil {
        return nil, errors.New("AF_PACKET handle not opened")
    }
    stats := aph.handle.SocketStats()
    return stats, nil
}

// SetFanout enables fanout for load balancing across multiple processes
func (aph *afPacketCaptureHandle) SetFanout(id uint16, fanoutType afpacket.FanoutType) error {
    if aph.handle == nil {
        return errors.New("AF_PACKET handle not opened")
    }
    return aph.handle.SetFanout(id, fanoutType)
}

// LinkType returns the link type of the capture interface
func (aph *afPacketCaptureHandle) LinkType() layers.LinkType {
    return layers.LinkTypeEthernet
}

// IsOpen returns whether the handle is open
func (aph *afPacketCaptureHandle) IsOpen() bool {
    return aph.handle != nil
}

// FanoutGroupType represents the type of fanout group
type FanoutGroupType int

const (
    FanoutHash    FanoutGroupType = syscall.PACKET_FANOUT_HASH
    FanoutLB      FanoutGroupType = syscall.PACKET_FANOUT_LB
    FanoutCPU     FanoutGroupType = syscall.PACKET_FANOUT_CPU
    FanoutRollb   FanoutGroupType = syscall.PACKET_FANOUT_ROLLB
    FanoutRandom  FanoutGroupType = syscall.PACKET_FANOUT_RND
)

// GetInterfaceMTU returns the MTU of the interface
func GetInterfaceMTU(iface string) (int, error) {
    ifaceFile, err := os.Open("/sys/class/net/" + iface + "/mtu")
    if err != nil {
        return 0, err
    }
    defer ifaceFile.Close()

    var mtu int
    _, err = fmt.Fscan(ifaceFile, &mtu)
    if err != nil {
        return 0, err
    }

    return mtu, nil
}

// GetInterfaceFlags returns the flags of the interface
func GetInterfaceFlags(iface string) (uint32, error) {
    flagsPath := fmt.Sprintf("/sys/class/net/%s/flags", iface)
    flagsFile, err := os.Open(flagsPath)
    if err != nil {
        return 0, err
    }
    defer flagsFile.Close()

    var flags uint32
    _, err = fmt.Fscan(flagsFile, "%x", &flags)
    if err != nil {
        return 0, err
    }

    return flags, nil
}

// SetPacketLossPriority sets the loss priority for a packet
// This is used with tc to classify packets
func SetPacketLossPriority(fd int, priority uint8) error {
    // Use socket option to set traffic class
    opt := syscall.TC_PRIO_MAX + int32(priority)
    _, _, err := syscall.Syscall6(
        syscall.SYS_SETSOCKOPT,
        uintptr(fd),
        syscall.SOL_PACKET,
        syscall.PACKET_LOSS,
        uintptr(unsafe.Pointer(&opt)),
        unsafe.Sizeof(opt),
        0,
    )
    if err != 0 {
        return err
    }
    return nil
}

// GetHardwareAddr returns the hardware address (MAC) of an interface
func GetHardwareAddr(iface string) ([]byte, error) {
    ifaceFile, err := os.Open("/sys/class/net/" + iface + "/address")
    if err != nil {
        return nil, err
    }
    defer ifaceFile.Close()

    var mac [6]byte
    _, err = fmt.Fscan(ifaceFile, "%02x:%02x:%02x:%02x:%02x:%02x",
        &mac[0], &mac[1], &mac[2], &mac[3], &mac[4], &mac[5])
    if err != nil {
        return nil, err
    }

    return mac[:], nil
}
