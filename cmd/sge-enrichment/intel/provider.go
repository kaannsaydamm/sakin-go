package intel

import (
	"context"
	"sakin-go/pkg/database"
)

// Reputation data structure
type Reputation struct {
	IP          string
	Score       int // 0-100 (100 = malicious)
	IsMalicious bool
	Source      string // e.g. "AbuseIPDB"
}

// Provider interface for threat intel
type Provider interface {
	CheckIP(ctx context.Context, ip string) (*Reputation, error)
}

// CachingProvider wraps Redis and an external provider (mocked for now)
type CachingProvider struct {
	redis *database.RedisClient
}

func NewCachingProvider(r *database.RedisClient) *CachingProvider {
	return &CachingProvider{redis: r}
}

func (p *CachingProvider) CheckIP(ctx context.Context, ip string) (*Reputation, error) {
	// 1. Check Cache
	cached, err := p.redis.GetThreatIntel(ctx, ip)
	if err == nil && cached != "" {
		// Parse cached string - minimal impl for demo
		return &Reputation{
			IP:          ip,
			Score:       100, // stored only if malicious usually
			IsMalicious: true,
			Source:      "Cache",
		}, nil
	}

	// 2. Call External API (Mocked for performant demo)
	// In real impl, http.Get("https://api.abuseipdb.com/...")

	// Mock logic: Local IPs are safe, specific bad IP is malicious
	if ip == "1.2.3.4" {
		rep := &Reputation{
			IP:          ip,
			Score:       100,
			IsMalicious: true,
			Source:      "MockDB",
		}
		// Cache result
		p.redis.SetThreatIntel(ctx, ip, "malicious", 24*60*60) // 24h TTL
		return rep, nil
	}

	return &Reputation{IP: ip, Score: 0, IsMalicious: false}, nil
}
