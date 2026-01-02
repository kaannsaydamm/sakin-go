package config

import (
	"flag"
	"os"
)

type AgentConfig struct {
	AgentID   string
	ServerURL string // NATS URL (tls://...)
	CertFile  string
	KeyFile   string
	CAFile    string

	// Collection intervals
	HostInfoInterval int
	AuditInterval    int
}

func LoadConfig() *AgentConfig {
	cfg := &AgentConfig{}

	flag.StringVar(&cfg.AgentID, "id", getEnv("AGENT_ID", "agent-unknown"), "Unique Agent ID")
	flag.StringVar(&cfg.ServerURL, "server", getEnv("SGE_SERVER_URL", "tls://localhost:4222"), "SGE Server NATS URL")
	flag.StringVar(&cfg.CertFile, "cert", getEnv("SGE_CERT_FILE", "./certs/client.crt"), "Client Certificate")
	flag.StringVar(&cfg.KeyFile, "key", getEnv("SGE_KEY_FILE", "./certs/client.key"), "Client Key")
	flag.StringVar(&cfg.CAFile, "ca", getEnv("SGE_CA_FILE", "./certs/ca.crt"), "CA Certificate")
	flag.IntVar(&cfg.HostInfoInterval, "host-interval", 60, "Host info collection interval (seconds)")

	flag.Parse()

	// Auto-generate ID if needed (could rely on machine-id)
	if cfg.AgentID == "agent-unknown" {
		hostname, _ := os.Hostname()
		cfg.AgentID = "agent-" + hostname
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
