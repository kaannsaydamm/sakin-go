package dpi

import (
    "bytes"
    "encoding/hex"
    "fmt"
    "net"
    "strings"
    "time"

    "github.com/atailh4n/sakin/internal/config"
    "github.com/atailh4n/sakin/internal/normalization"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

// protocolParser handles protocol-specific packet parsing
type protocolParser struct {
    cfg       *config.DPIConfig
    ipParser  *IPParser
    tcpParser *TCPParser
    udpParser *UDPParser
    httpParser *HTTPParser
    dnsParser  *DNSParser
    tlsParser  *TLSParser
    smbParser  *SMBParser
}

// IPParser parses IP layer information
type IPParser struct{}

type TCPParser struct{}

type UDPParser struct{}

type HTTPParser struct{}

type DNSParser struct{}

type TLSParser struct{}

type SMBParser struct{}

// HTTPPacket represents parsed HTTP information
type HTTPPacket struct {
    Method    string
    URL       string
    Host      string
    UserAgent string
    ContentType string
    ContentLength int
    Headers    map[string]string
    Body       []byte
    IsTLS      bool
}

// DNSPacket represents parsed DNS information
type DNSPacket struct {
    TransactionID uint16
    Query         bool
    Domain        string
    RecordType    string
    RecordClass   string
    ResponseCode  uint8
    Answers       []DNSAnswer
}

// DNSAnswer represents a DNS answer record
type DNSAnswer struct {
    Domain string
    Type   string
    Class  string
    TTL    uint32
    Data   string
}

// TLSPacket represents parsed TLS information
type TLSPacket struct {
    Handshake       bool
    Version         string
    CipherSuite     string
    ServerName      string
    CertificateInfo *CertificateInfo
    Extensions      []TLSExtension
}

// CertificateInfo represents TLS certificate information
type CertificateInfo struct {
    Subject      string
    Issuer       string
    SerialNumber string
    NotBefore    time.Time
    NotAfter     time.Time
    SubjectAltNames []string
}

// TLSExtension represents a TLS extension
type TLSExtension struct {
    Type   uint16
    Name   string
    Data   []byte
}

// SMBPacket represents parsed SMB information
type SMBPacket struct {
    Command     string
    TreeID      uint16
    MessageID   uint16
    SessionID   uint64
    Action      string
    Path        string
    Operation   string
}

// newProtocolParser creates a new protocol parser
func newProtocolParser(cfg *config.DPIConfig) (*protocolParser, error) {
    return &protocolParser{
        cfg:       cfg,
        ipParser:  &IPParser{},
        tcpParser: &TCPParser{},
        udpParser: &UDPParser{},
        httpParser: &HTTPParser{},
        dnsParser:  &DNSParser{},
        tlsParser:  &TLSParser{},
        smbParser:  &SMBParser{},
    }, nil
}

// parsePacket parses a packet and extracts normalized events
func (pp *protocolParser) parsePacket(packet gopacket.Packet, iface string, ts time.Time) []*normalization.NetworkEvent {
    events := make([]*normalization.NetworkEvent, 0)

    // Parse IP layer
    ipEvent := pp.ipParser.parse(packet, iface, ts)
    if ipEvent != nil {
        events = append(events, ipEvent)
    }

    // Parse transport layer
    var srcPort, dstPort uint16
    var transportProtocol string

    if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
        tcp, _ := tcpLayer.(*layers.TCP)
        srcPort = uint16(tcp.SrcPort)
        dstPort = uint16(tcp.DstPort)
        transportProtocol = "TCP"
        transportEvent := pp.tcpParser.parse(packet, tcp, ipEvent, ts)
        if transportEvent != nil {
            events = append(events, transportEvent)
        }
    } else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
        udp, _ := udpLayer.(*layers.UDP)
        srcPort = uint16(udp.SrcPort)
        dstPort = uint16(udp.DstPort)
        transportProtocol = "UDP"
        transportEvent := pp.udpParser.parse(packet, udp, ipEvent, ts)
        if transportEvent != nil {
            events = append(events, transportEvent)
        }
    }

    // Update IP event with port info
    if ipEvent != nil {
        ipEvent.SourcePort = srcPort
        ipEvent.DestPort = dstPort
        ipEvent.TransportProtocol = transportProtocol
    }

    // Parse application layer
    if appLayer := packet.ApplicationLayer(); appLayer != nil {
        payload := appLayer.Payload()

        // Detect protocol based on ports and payload
        switch {
        case pp.cfg.HTTPEnabled && isHTTPPort(dstPort):
            if httpEvent := pp.httpParser.parse(payload, ipEvent, ts); httpEvent != nil {
                events = append(events, httpEvent)
            }
        case pp.cfg.DNSEnabled && isDNSPort(dstPort):
            if dnsEvent := pp.dnsParser.parse(payload, ipEvent, ts); dnsEvent != nil {
                events = append(events, dnsEvent)
            }
        case pp.cfg.TLSEnabled && isTLSPort(dstPort):
            if tlsEvent := pp.tlsParser.parse(payload, ipEvent, ts); tlsEvent != nil {
                events = append(events, tlsEvent)
            }
        case pp.cfg.SMBEnabled && isSMBPort(dstPort):
            if smbEvent := pp.smbParser.parse(payload, ipEvent, ts); smbEvent != nil {
                events = append(events, smbEvent)
            }
        }

        // Generic application layer event
        if len(payload) > 0 {
            appEvent := &normalization.NetworkEvent{
                Timestamp:      ts,
                EventType:      normalization.EventTypeApplication,
                Protocol:       "application",
                Severity:       normalization.SeverityInfo,
                SourceIP:       ipEvent.SourceIP,
                DestIP:         ipEvent.DestIP,
                SourcePort:     srcPort,
                DestPort:       dstPort,
                TransportProtocol: transportProtocol,
                PayloadPreview: truncatePreview(payload),
            }
            events = append(events, appEvent)
        }
    }

    return events
}

// parse parses IP layer information
func (p *IPParser) parse(packet gopacket.Packet, iface string, ts time.Time) *normalization.NetworkEvent {
    var srcIP, dstIP string
    var protocol string
    var length uint16

    switch ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer.(type) {
    case *layers.IPv4:
        ip, _ := ipLayer.(*layers.IPv4)
        srcIP = ip.SrcIP.String()
        dstIP = ip.DstIP.String()
        protocol = ip.Protocol.String()
        length = ip.Length
    case *layers.IPv6:
        ip, _ := packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
        srcIP = ip.SrcIP.String()
        dstIP = ip.DstIP.String()
        protocol = ip.NextHeader.String()
        length = ip.Length
    default:
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:      ts,
        EventType:      normalization.EventTypeNetwork,
        Protocol:       protocol,
        Severity:       normalization.SeverityInfo,
        SourceIP:       srcIP,
        DestIP:         dstIP,
        Interface:      iface,
        PayloadSize:    int(length),
    }
}

// parse parses TCP layer information
func (p *TCPParser) parse(packet gopacket.Packet, tcp *layers.TCP, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil {
        return nil
    }

    syn := tcp.SYN
    ack := tcp.ACK
    fin := tcp.FIN
    rst := tcp.RST

    var tcpFlags string
    if syn {
        tcpFlags += "S"
    }
    if ack {
        tcpFlags += "A"
    }
    if fin {
        tcpFlags += "F"
    }
    if rst {
        tcpFlags += "R"
    }
    if tcp.PSH {
        tcpFlags += "P"
    }
    if tcp.URG {
        tcpFlags += "U"
    }
    if tcp.ECE {
        tcpFlags += "E"
    }
    if tcp.CWR {
        tcpFlags += "C"
    }

    return &normalization.NetworkEvent{
        Timestamp:        ts,
        EventType:        normalization.EventTypeTransport,
        Protocol:         "TCP",
        Severity:         normalization.SeverityInfo,
        SourceIP:         ipEvent.SourceIP,
        DestIP:           ipEvent.DestIP,
        SourcePort:       uint16(tcp.SrcPort),
        DestPort:         uint16(tcp.DstPort),
        TransportProtocol: "TCP",
        TCPFlags:         tcpFlags,
        SequenceNumber:   tcp.Seq,
        Acknowledgment:   tcp.AckSeq,
        WindowSize:       int(tcp.Window),
        PayloadSize:      len(tcp.Payload),
    }
}

// parse parses UDP layer information
func (p *UDPParser) parse(packet gopacket.Packet, udp *layers.UDP, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil {
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:      ts,
        EventType:      normalization.EventTypeTransport,
        Protocol:       "UDP",
        Severity:       normalization.SeverityInfo,
        SourceIP:       ipEvent.SourceIP,
        DestIP:         ipEvent.DestIP,
        SourcePort:     uint16(udp.SrcPort),
        DestPort:       uint16(udp.DstPort),
        TransportProtocol: "UDP",
        PayloadSize:    int(udp.Length),
    }
}

// parse parses HTTP payload
func (p *HTTPParser) parse(payload []byte, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil {
        return nil
    }

    http := parseHTTP(payload)
    if http == nil {
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:        ts,
        EventType:        normalization.EventTypeApplication,
        Protocol:         "HTTP",
        Severity:         normalization.SeverityInfo,
        SourceIP:         ipEvent.SourceIP,
        DestIP:           ipEvent.DestIP,
        SourcePort:       ipEvent.SourcePort,
        DestPort:         ipEvent.DestPort,
        TransportProtocol: ipEvent.TransportProtocol,
        ApplicationData: map[string]interface{}{
            "method":       http.Method,
            "url":          http.URL,
            "host":         http.Host,
            "user_agent":   http.UserAgent,
            "content_type": http.ContentType,
            "headers":      http.Headers,
        },
        PayloadPreview: truncatePreview(http.Body),
    }
}

// parse parses DNS payload
func (p *DNSParser) parse(payload []byte, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil || len(payload) < 12 {
        return nil
    }

    dns := parseDNS(payload)
    if dns == nil {
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:        ts,
        EventType:        normalization.EventTypeApplication,
        Protocol:         "DNS",
        Severity:         normalization.SeverityInfo,
        SourceIP:         ipEvent.SourceIP,
        DestIP:           ipEvent.DestIP,
        SourcePort:       ipEvent.SourcePort,
        DestPort:         ipEvent.DestPort,
        TransportProtocol: ipEvent.TransportProtocol,
        ApplicationData: map[string]interface{}{
            "transaction_id": dns.TransactionID,
            "query":          dns.Query,
            "domain":         dns.Domain,
            "record_type":    dns.RecordType,
            "response_code":  dns.ResponseCode,
            "answers":        dns.Answers,
        },
    }
}

// parse parses TLS payload
func (p *TLSParser) parse(payload []byte, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil || len(payload) < 5 {
        return nil
    }

    tls := parseTLS(payload)
    if tls == nil {
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:        ts,
        EventType:        normalization.EventTypeApplication,
        Protocol:         "TLS",
        Severity:         normalization.SeverityInfo,
        SourceIP:         ipEvent.SourceIP,
        DestIP:           ipEvent.DestIP,
        SourcePort:       ipEvent.SourcePort,
        DestPort:         ipEvent.DestPort,
        TransportProtocol: ipEvent.TransportProtocol,
        ApplicationData: map[string]interface{}{
            "handshake":      tls.Handshake,
            "version":        tls.Version,
            "cipher_suite":   tls.CipherSuite,
            "server_name":    tls.ServerName,
            "certificate":    tls.CertificateInfo,
            "extensions":     tls.Extensions,
        },
    }
}

// parse parses SMB payload
func (p *SMBParser) parse(payload []byte, ipEvent *normalization.NetworkEvent, ts time.Time) *normalization.NetworkEvent {
    if ipEvent == nil {
        return nil
    }

    smb := parseSMB(payload)
    if smb == nil {
        return nil
    }

    return &normalization.NetworkEvent{
        Timestamp:        ts,
        EventType:        normalization.EventTypeApplication,
        Protocol:         "SMB",
        Severity:         normalization.SeverityInfo,
        SourceIP:         ipEvent.SourceIP,
        DestIP:           ipEvent.DestIP,
        SourcePort:       ipEvent.SourcePort,
        DestPort:         ipEvent.DestPort,
        TransportProtocol: ipEvent.TransportProtocol,
        ApplicationData: map[string]interface{}{
            "command":     smb.Command,
            "tree_id":     smb.TreeID,
            "message_id":  smb.MessageID,
            "session_id":  smb.SessionID,
            "action":      smb.Action,
            "path":        smb.Path,
            "operation":   smb.Operation,
        },
    }
}

// parseHTTP parses HTTP request/response
func parseHTTP(data []byte) *HTTPPacket {
    http := &HTTPPacket{
        Headers: make(map[string]string),
    }

    // Split headers and body
    parts := bytes.SplitN(data, []byte("\r\n\r\n"), 2)
    headerData := parts[0]
    var body []byte
    if len(parts) > 1 {
        body = parts[1]
    }

    // Parse request line or status line
    headerLines := bytes.Split(headerData, []byte("\r\n"))
    if len(headerLines) == 0 {
        return nil
    }

    firstLine := string(headerLines[0])
    requestParts := strings.Fields(firstLine)

    if len(requestParts) >= 3 && (requestParts[0] == "GET" || requestParts[0] == "POST" ||
        requestParts[0] == "PUT" || requestParts[0] == "DELETE" || requestParts[0] == "HEAD" ||
        requestParts[0] == "OPTIONS" || requestParts[0] == "PATCH") {
        // HTTP Request
        http.Method = requestParts[0]
        http.URL = requestParts[1]
    } else if len(requestParts) >= 2 && strings.HasPrefix(requestParts[0], "HTTP/") {
        // HTTP Response - extract status code
        http.Method = "RESPONSE"
    }

    // Parse headers
    for i := 1; i < len(headerLines); i++ {
        line := string(headerLines[i])
        colonIdx := strings.Index(line, ":")
        if colonIdx > 0 {
            key := strings.TrimSpace(line[:colonIdx])
            value := strings.TrimSpace(line[colonIdx+1:])
            http.Headers[key] = value

            // Extract important headers
            switch strings.ToLower(key) {
            case "host":
                http.Host = value
            case "user-agent":
                http.UserAgent = value
            case "content-type":
                http.ContentType = value
            case "content-length":
                fmt.Sscanf(value, "%d", &http.ContentLength)
            }
        }
    }

    http.Body = body
    return http
}

// parseDNS parses DNS packet
func parseDNS(data []byte) *DNSPacket {
    if len(data) < 12 {
        return nil
    }

    dns := &DNSPacket{
        TransactionID: uint16(data[0])<<8 | uint16(data[1]),
        Answers:       make([]DNSAnswer, 0),
    }

    flags := uint16(data[2])<<8 | uint16(data[3])
    dns.Query = (flags & 0x8000) == 0

    // Count sections
    numQuestions := int(data[4])<<8 | int(data[5])
    numAnswers := int(data[6])<<8 | int(data[7])
    numAuthority := int(data[8])<<8 | int(data[9])
    numAdditional := int(data[10])<<8 | int(data[11])

    responseCode := flags & 0x0F
    dns.ResponseCode = uint8(responseCode)

    // Parse questions
    offset := 12
    for i := 0; i < numQuestions && offset < len(data); i++ {
        name, newOffset := parseDNSName(data, offset)
        if newOffset <= offset {
            break
        }
        offset = newOffset

        if offset+4 > len(data) {
            break
        }
        qType := uint16(data[offset])<<8 | uint16(data[offset+1])
        qClass := uint16(data[offset+2])<<8 | uint16(data[offset+3])
        offset += 4

        dns.Domain = name
        dns.RecordType = dnsTypeToString(qType)
        dns.RecordClass = dnsClassToString(qClass)
    }

    // Parse answers
    for i := 0; i < numAnswers && offset < len(data); i++ {
        name, newOffset := parseDNSName(data, offset)
        if newOffset <= offset {
            break
        }
        offset = newOffset

        if offset+10 > len(data) {
            break
        }

        aType := uint16(data[offset])<<8 | uint16(data[offset+1])
        aClass := uint16(data[offset+2])<<8 | uint16(data[offset+3])
        ttl := uint32(data[offset+4])<<24 | uint32(data[offset+5])<<16 | uint32(data[offset+6])<<8 | uint32(data[offset+7])
        rdLength := int(data[offset+8])<<8 | int(data[offset+9])
        offset += 10

        if offset+rdLength > len(data) {
            break
        }
        rData := data[offset : offset+rdLength]
        offset += rdLength

        var dataStr string
        switch aType {
        case 1: // A record
            dataStr = net.IP(rData).String()
        case 5: // CNAME
            name, _ := parseDNSName(rData, 0)
            dataStr = name
        case 16: // TXT
            dataStr = string(rData[1:])
        default:
            dataStr = hex.EncodeToString(rData)
        }

        dns.Answers = append(dns.Answers, DNSAnswer{
            Domain: name,
            Type:   dnsTypeToString(aType),
            Class:  dnsClassToString(aClass),
            TTL:    ttl,
            Data:   dataStr,
        })
    }

    return dns
}

// parseDNSName parses a DNS name from packet data
func parseDNSName(data []byte, offset int) (string, int) {
    var name bytes.Buffer
    start := offset

    for offset < len(data) {
        length := int(data[offset])
        if length == 0 {
            offset++
            break
        }
        if length&0xC0 == 0xC0 {
            // Pointer
            if offset+1 >= len(data) {
                break
            }
            pointer := ((length & 0x3F) << 8) | int(data[offset+1])
            pointedName, _ := parseDNSName(data, pointer)
            if name.Len() > 0 {
                name.WriteByte('.')
            }
            name.WriteString(pointedName)
            offset += 2
            break
        }
        offset++
        if offset+length > len(data) {
            break
        }
        if name.Len() > 0 {
            name.WriteByte('.')
        }
        name.Write(data[offset : offset+length])
        offset += length
    }

    return name.String(), offset
}

// parseTLS parses TLS record
func parseTLS(data []byte) *TLSPacket {
    if len(data) < 5 {
        return nil
    }

    tls := &TLSPacket{}

    contentType := data[0]
    version := (uint16(data[1]) << 8) | uint16(data[2])
    tls.Version = tlsVersionToString(version)

    switch contentType {
    case 22: // Handshake
        tls.Handshake = true
        tls.parseHandshake(data[5:])
    case 20, 21, 23, 24: // ChangeCipherSpec, Alert, Application Data, Heartbeat
        // Not a handshake
    }

    return tls
}

// parseHandshake parses TLS handshake
func (t *TLSParser) parseHandshake(data []byte) {
    if len(data) < 4 {
        return
    }

    msgType := data[0]
    length := (uint32(data[1]) << 16) | (uint32(data[2]) << 8) | uint32(data[3])
    msgData := data[4 : 4+length]

    switch msgType {
    case 1: // Client Hello
        t.parseClientHello(msgData)
    case 2: // Server Hello
        t.parseServerHello(msgData)
    }
}

// parseClientHello parses TLS Client Hello
func (t *TLSParser) parseClientHello(data []byte) *TLSPacket {
    tls := &TLSPacket{Handshake: true}

    offset := 0

    // Version
    if offset+2 > len(data) {
        return tls
    }
    tls.Version = tlsVersionToString((uint16(data[0]) << 8) | uint16(data[1]))
    offset += 2

    // Random (32 bytes)
    offset += 32

    // Session ID
    if offset >= len(data) {
        return tls
    }
    sessionLen := int(data[offset])
    offset += 1 + sessionLen

    // Cipher Suites
    if offset+2 > len(data) {
        return tls
    }
    cipherLen := (int(data[offset]) << 8) | int(data[offset+1])
    offset += 2 + cipherLen

    // Compression Methods
    if offset >= len(data) {
        return tls
    }
    compLen := int(data[offset])
    offset += 1 + compLen

    // Extensions
    if offset+2 > len(data) {
        return tls
    }
    extLen := (int(data[offset]) << 8) | int(data[offset+1])
    offset += 2

    for offset+4 <= len(data) && offset < extLen+4 {
        extType := (uint16(data[offset]) << 8) | uint16(data[offset+1])
        extLen := (int(data[offset+2]) << 8) | int(data[offset+3])
        offset += 4

        if extType == 0 { // Server Name (SNI)
            if offset+2 > len(data) {
                break
            }
            nameType := data[offset]
            nameLen := (int(data[offset+1]) << 8) | int(data[offset+2])
            if nameType == 0 && offset+3+nameLen <= len(data) {
                tls.ServerName = string(data[offset+3 : offset+3+nameLen])
            }
        }

        offset += extLen
    }

    return tls
}

// parseServerHello parses TLS Server Hello
func (t *TLSParser) parseServerHello(data []byte) {
    // Similar parsing logic for Server Hello
}

// parseSMB parses SMB packet
func parseSMB(data []byte) *SMBPacket {
    if len(data) < 32 {
        return nil
    }

    // Check SMB signature
    if !bytes.HasPrefix(data, []byte{0xFF, 'S', 'M', 'B'}) {
        return nil
    }

    smb := &SMBPacket{}

    // Parse SMB header
    command := data[4]
    ntStatus := uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24

    smb.Command = smbCommandToString(command)

    // Extract common fields
    if len(data) > 24 {
        smb.TreeID = uint16(data[20]) | uint16(data[21])<<8
        smb.MessageID = uint16(data[22]) | uint16(data[23])<<8
        smb.SessionID = uint64(data[24]) | uint64(data[25])<<8 | uint64(data[26])<<16 | uint64(data[27])<<24 |
            uint64(data[28])<<32 | uint64(data[29])<<40 | uint64(data[30])<<48 | uint64(data[31])<<56
    }

    // Set action based on NT status
    if ntStatus == 0 {
        smb.Action = "success"
    } else {
        smb.Action = fmt.Sprintf("0x%08X", ntStatus)
    }

    // Parse additional fields based on command
    switch command {
    case 0x75: // SMB2 CREATE
        if len(data) > 56 {
            smb.Path = parseUTF16String(data[56:])
        }
    case 0x31: // SMB READ
        smb.Operation = "read"
    case 0x37: // SMB WRITE
        smb.Operation = "write"
    }

    return smb
}

// parseUTF16String parses UTF-16 string from packet data
func parseUTF16String(data []byte) string {
    var str strings.Builder

    for i := 0; i+1 < len(data); i += 2 {
        if data[i] == 0 && data[i+1] == 0 {
            break
        }
        if data[i+1] == 0 {
            str.WriteByte(data[i])
        }
    }

    return str.String()
}

// Helper functions for protocol detection
func isHTTPPort(port uint16) bool {
    return port == 80 || port == 8080 || port == 8008 || port == 3000 || port == 5000 || port == 8000
}

func isDNSPort(port uint16) bool {
    return port == 53
}

func isTLSPort(port uint16) bool {
    return port == 443 || port == 8443 || port == 993 || port == 995 || port == 5223
}

func isSMBPort(port uint16) bool {
    return port == 139 || port == 445
}

func dnsTypeToString(t uint16) string {
    types := map[uint16]string{
        1:  "A",
        2:  "NS",
        5:  "CNAME",
        6:  "SOA",
        12: "PTR",
        15: "MX",
        16: "TXT",
        28: "AAAA",
        33: "SRV",
    }
    if s, ok := types[t]; ok {
        return s
    }
    return fmt.Sprintf("TYPE%d", t)
}

func dnsClassToString(c uint16) string {
    if c == 1 {
        return "IN"
    }
    return fmt.Sprintf("CLASS%d", c)
}

func tlsVersionToString(v uint16) string {
    versions := map[uint16]string{
        0x0300: "SSL30",
        0x0301: "TLS10",
        0x0302: "TLS11",
        0x0303: "TLS12",
        0x0304: "TLS13",
    }
    if s, ok := versions[v]; ok {
        return s
    }
    return fmt.Sprintf("0x%04X", v)
}

func smbCommandToString(c uint8) string {
    commands := map[uint8]string{
        0x72: "SMB2_NEGOTIATE",
        0x73: "SMB2_SESSION_SETUP",
        0x75: "SMB2_CREATE",
        0x76: "SMB2_CLOSE",
        0x77: "SMB2_FLUSH",
        0x78: "SMB2_READ",
        0x79: "SMB2_WRITE",
        0x7A: "SMB2_LOCK",
        0x7B: "SMB2_IOCTL",
        0x7C: "SMB2_CANCEL",
        0x7D: "SMB2_PUSH",
        0x7E: "SMB2_PULL",
    }
    if s, ok := commands[c]; ok {
        return s
    }
    return fmt.Sprintf("0x%02X", c)
}

func truncatePreview(data []byte) string {
    maxLen := 256
    if len(data) <= maxLen {
        return string(data)
    }
    return string(data[:maxLen]) + "..."
}
