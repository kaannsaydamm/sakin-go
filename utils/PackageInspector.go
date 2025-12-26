// Package Utils provides utility functions for network traffic monitoring
// This package is deprecated in favor of internal/dpi package
package Utils

import (
    "database/sql"
    "fmt"
    "log"
    "sync"
    "time"

    Handlers "github.com/atailh4n/sakin/handlers"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pcap"
)

// MonitorTraffic monitors network traffic on specified interfaces
// Deprecated: Use internal/dpi package instead
func MonitorTraffic(ifaces []pcap.Interface, db *sql.DB, wg *sync.WaitGroup) {
    for _, iface := range ifaces {
        log.Printf("Found network interface: %s\n", iface.Name)
        wg.Add(1)

        go func(ifaceName string) {
            defer wg.Done()

            // Open network interface
            handle, err := pcap.OpenLive(ifaceName, 1600, true, pcap.BlockForever)
            if err != nil {
                log.Printf("Error opening device %s: %v", ifaceName, err)
                return
            }
            log.Printf("Successfully opened device %s\n", ifaceName)
            defer handle.Close()

            packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

            // Process each packet
            for packet := range packetSource.Packets() {
                log.Printf("Processing packet\n")
                if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
                    ip, _ := ipLayer.(*layers.IPv4)
                    log.Printf("Captured IP Packet: %s -> %s\n", ip.SrcIP, ip.DstIP)
                    timestamp := time.Now()

                    // Save to database
                    payloadSize := 0
                    if appLayer := packet.ApplicationLayer(); appLayer != nil {
                        payloadSize = len(appLayer.Payload())
                    }
                    err := Handlers.SavePacket(db, ip.SrcIP.String(), ip.DstIP.String(), ip.Protocol.String(), payloadSize, timestamp)
                    if err != nil {
                        log.Printf("Error saving packet to DB: %v", err)
                    }

                    // Log application layer (HTTP/HTTPS)
                    if packet.ApplicationLayer() != nil {
                        payload := packet.ApplicationLayer().Payload()
                        if len(payload) > 0 {
                            log.Printf("Captured Payload: %s\n", string(payload))
                        } else {
                            log.Printf("Captured Encrypted HTTPS Traffic: %s -> %s\n", ip.SrcIP, ip.DstIP)
                        }
                    }
                } else {
                    log.Printf("Non-IP packet captured\n")
                }
            }
        }(iface.Name)
    }
}

// MonitorTrafficWithDPI monitors traffic with Deep Packet Inspection
// This is the recommended function for new implementations
func MonitorTrafficWithDPI(ifaces []pcap.Interface, db *sql.DB, wg *sync.WaitGroup, dpiEnabled bool) {
    for _, iface := range ifaces {
        log.Printf("Found network interface: %s (DPI: %v)\n", iface.Name, dpiEnabled)
        wg.Add(1)

        go func(ifaceName string) {
            defer wg.Done()

            handle, err := pcap.OpenLive(ifaceName, 1600, true, pcap.BlockForever)
            if err != nil {
                log.Printf("Error opening device %s: %v", ifaceName, err)
                return
            }
            defer handle.Close()

            log.Printf("Successfully opened device %s with DPI support\n", ifaceName)

            packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

            for packet := range packetSource.Packets() {
                processPacketWithDPI(packet, db, dpiEnabled)
            }
        }(iface.Name)
    }
}

// processPacketWithDPI processes a packet with optional DPI
func processPacketWithDPI(packet gopacket.Packet, db *sql.DB, dpiEnabled bool) {
    if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
        ip, _ := ipLayer.(*layers.IPv4)
        timestamp := time.Now()

        payloadSize := 0
        protocol := ip.Protocol.String()

        if packet.ApplicationLayer() != nil {
            payload := packet.ApplicationLayer().Payload()
            payloadSize = len(payload)

            if dpiEnabled {
                protocol = detectProtocol(payload, 80)
            }

            if len(payload) > 0 && dpiEnabled {
                log.Printf("DPI: Captured %s payload: %s\n", protocol, truncate(payload, 256))
            }
        }

        err := Handlers.SavePacket(db, ip.SrcIP.String(), ip.DstIP.String(), protocol, payloadSize, timestamp)
        if err != nil {
            log.Printf("Error saving packet to DB: %v", err)
        }
    }
}

// detectProtocol detects the application protocol from payload
func detectProtocol(payload []byte, port uint16) string {
    if len(payload) == 0 {
        return "unknown"
    }

    // Check for HTTP
    if string(payload[:4]) == "GET " || string(payload[:4]) == "POST" ||
        string(payload[:4]) == "HEAD" || string(payload[:4]) == "PUT " ||
        string(payload[:4]) == "DELE" || string(payload[:4]) == "HTTP" {
        return "HTTP"
    }

    // Check for DNS (typically UDP port 53)
    if port == 53 && len(payload) >= 12 {
        return "DNS"
    }

    // Check for TLS/SSL handshake
    if len(payload) >= 5 && payload[0] == 0x16 && payload[1] == 0x03 {
        return "TLS"
    }

    // Check for SSH
    if len(payload) >= 4 && string(payload[:4]) == "SSH-" {
        return "SSH"
    }

    // Check for SMB
    if len(payload) >= 4 && payload[0] == 0xff && payload[1] == 'S' && payload[2] == 'M' && payload[3] == 'B' {
        return "SMB"
    }

    return "TCP"
}

// truncate truncates a string to the specified length
func truncate(s []byte, maxLen int) string {
    if len(s) <= maxLen {
        return string(s)
    }
    return string(s[:maxLen]) + "..."
}

// GetNetworkInterfaces returns available network interfaces
func GetNetworkInterfaces() ([]pcap.Interface, error) {
    return pcap.FindAllDevs()
}

// PrintInterfaces prints available network interfaces
func PrintInterfaces() {
    ifaces, err := pcap.FindAllDevs()
    if err != nil {
        log.Printf("Error finding interfaces: %v", err)
        return
    }

    fmt.Println("Available Network Interfaces:")
    fmt.Println("----------------------------")
    for _, iface := range ifaces {
        fmt.Printf("  Name: %s\n", iface.Name)
        fmt.Printf("    Description: %s\n", iface.Description)
        fmt.Printf("    Flags: %v\n", iface.Flags)
        for _, addr := range iface.Addresses {
            fmt.Printf("    Address: %s\n", addr.IP.String())
        }
        fmt.Println()
    }
}

// IsLoopback checks if an interface is a loopback interface
func IsLoopback(iface pcap.Interface) bool {
    for _, flag := range iface.Flags {
        if flag&pcap.InterfaceFlagLoopback != 0 {
            return true
        }
    }
    return iface.Name == "lo" || iface.Name == "lo0"
}

// FilterInterfaces returns non-loopback interfaces
func FilterInterfaces(ifaces []pcap.Interface) []pcap.Interface {
    filtered := make([]pcap.Interface, 0)
    for _, iface := range ifaces {
        if !IsLoopback(iface) {
            filtered = append(filtered, iface)
        }
    }
    return filtered
}
