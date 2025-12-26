// Package Utils provides attack vector detection utilities
// This package is deprecated in favor of internal/dpi/threat_detector.go
package Utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// AttackType represents the type of attack vector
type AttackType string

const (
	AttackTypePortScan      AttackType = "port_scan"
	AttackTypeBruteForce    AttackType = "brute_force"
	AttackTypeInjection     AttackType = "injection"
	AttackTypeMalware       AttackType = "malware"
	AttackTypeC2Beacon      AttackType = "c2_beacon"
	AttackTypeExfiltration  AttackType = "exfiltration"
	AttackTypeSQLInjection  AttackType = "sql_injection"
	AttackTypeXSS           AttackType = "xss"
	AttackTypeDirectoryTrav AttackType = "directory_traversal"
)

// AttackSignature represents a known attack pattern
type AttackSignature struct {
	Type        AttackType
	Pattern     string
	Regex       string
	Severity    int
	Description string
}

// DetectedAttack represents a detected attack
type DetectedAttack struct {
	Type        AttackType
	Severity    int
	SourceIP    string
	DestIP      string
	DestPort    uint16
	Description string
	Payload     string
	Timestamp   time.Time
}

// Common attack signatures database
var attackSignatures = []AttackSignature{
	{
		Type:        AttackTypeSQLInjection,
		Pattern:     ".*('|(\")|(--)|(#)|(\\/\\*)).*",
		Severity:    90,
		Description: "SQL Injection attempt detected",
	},
	{
		Type:        AttackTypeDirectoryTrav,
		Pattern:     ".*(\\.\\.\\/|<script|\\.\\.\\\\).*",
		Severity:    80,
		Description: "Directory traversal or file inclusion attempt",
	},
	{
		Type:        AttackTypeXSS,
		Pattern:     ".*(<script|javascript:|on\\w+=).*",
		Severity:    85,
		Description: "Cross-site scripting (XSS) attempt",
	},
	{
		Type:        AttackTypeInjection,
		Pattern:     ".*(eval\\(|system\\(|exec\\(|shell_exec\\().*",
		Severity:    95,
		Description: "Code injection attempt",
	},
}

// DetectPotentialAttack analyzes a potential attack vector
// Deprecated: Use internal/dpi/threat_detector.go instead
func DetectPotentialAttack(ip string) bool {
	// Basic check - expanded in threat_detector.go
	return isSuspiciousIP(ip)
}

// AnalyzePayload analyzes payload for attack signatures
func AnalyzePayload(payload string) []DetectedAttack {
	attacks := make([]DetectedAttack, 0)

	for _, sig := range attackSignatures {
		if strings.Contains(strings.ToLower(payload), strings.ToLower(sig.Pattern)) {
			attacks = append(attacks, DetectedAttack{
				Type:        sig.Type,
				Severity:    sig.Severity,
				Description: sig.Description,
				Payload:     truncateString(payload, 256),
				Timestamp:   time.Now(),
			})
		}
	}

	return attacks
}

// AnalyzePacketPayload analyzes raw packet payload for attacks
func AnalyzePacketPayload(data []byte) []DetectedAttack {
	attacks := make([]DetectedAttack, 0)
	payload := string(data)

	for _, sig := range attackSignatures {
		if strings.Contains(strings.ToLower(payload), strings.ToLower(sig.Pattern)) {
			attacks = append(attacks, DetectedAttack{
				Type:        sig.Type,
				Severity:    sig.Severity,
				Description: sig.Description,
				Payload:     truncateString(payload, 256),
				Timestamp:   time.Now(),
			})
		}
	}

	return attacks
}

// IsPrivateIP checks if an IP is in a private range
func IsPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, r := range privateRanges {
		_, ipNet, err := net.ParseCIDR(r)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// IsSuspiciousPort checks if a port is commonly targeted
func IsSuspiciousPort(port uint16) bool {
	suspiciousPorts := []uint16{
		22,   // SSH
		23,   // Telnet
		3389, // RDP
		5900, // VNC
		1433, // MSSQL
		3306, // MySQL
		5432, // PostgreSQL
		27017, // MongoDB
	}

	for _, p := range suspiciousPorts {
		if port == p {
			return true
		}
	}

	return false
}

// IsBlacklistedIP checks if an IP is in a blacklist
// This is a simplified version - in production, use a proper threat intelligence feed
func IsBlacklistedIP(ip string) bool {
	// Known malicious IP ranges (example)
	// In production, integrate with threat intelligence feeds
	blacklist := []string{
		"1.2.3.4",
		"5.6.7.8",
	}

	for _, badIP := range blacklist {
		if ip == badIP {
			return true
		}
	}

	return false
}

// GetThreatScore calculates a threat score for an IP
func GetThreatScore(ip string, port uint16, payload string) int {
	score := 0

	// Check if private IP
	if IsPrivateIP(ip) {
		score += 10
	}

	// Check if suspicious port
	if IsSuspiciousPort(port) {
		score += 20
	}

	// Check if blacklisted
	if IsBlacklistedIP(ip) {
		score += 100
	}

	// Check payload for attacks
	attacks := AnalyzePayload(payload)
	for _, attack := range attacks {
		score += attack.Severity
	}

	return min(score, 100)
}

// GetAttackDescription returns a description for an attack type
func GetAttackDescription(attackType AttackType) string {
	descriptions := map[AttackType]string{
		AttackTypePortScan:      "Port scan detected - enumeration of open ports",
		AttackTypeBruteForce:    "Brute force attempt - repeated login attempts",
		AttackTypeSQLInjection:  "SQL injection attempt - malicious SQL code injection",
		AttackTypeXSS:           "Cross-site scripting (XSS) - malicious script injection",
		AttackTypeDirectoryTrav: "Directory traversal - unauthorized file access",
		AttackTypeC2Beacon:      "C2 beacon pattern - possible command and control traffic",
		AttackTypeExfiltration:  "Data exfiltration - unauthorized data transfer",
		AttackTypeMalware:       "Malware signature detected",
		AttackTypeInjection:     "Code injection attempt",
	}

	return descriptions[attackType]
}

// Helper functions

func isSuspiciousIP(ip string) bool {
	// Basic suspicious IP check
	if ip == "" {
		return true
	}

	// Check for private IPs (might be internal reconnaissance)
	if IsPrivateIP(ip) {
		return false // Private IPs are generally not suspicious
	}

	// Check for localhost
	if ip == "127.0.0.1" || ip == "::1" {
		return true
	}

	return false
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PrintAttackSignatures prints all known attack signatures
func PrintAttackSignatures() {
	fmt.Println("Attack Signatures:")
	fmt.Println("------------------")
	for _, sig := range attackSignatures {
		fmt.Printf("  Type: %s\n", sig.Type)
		fmt.Printf("    Pattern: %s\n", sig.Pattern)
		fmt.Printf("    Severity: %d\n", sig.Severity)
		fmt.Printf("    Description: %s\n", sig.Description)
		fmt.Println()
	}
}
