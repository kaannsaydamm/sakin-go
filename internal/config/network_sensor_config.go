package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// NetworkSensorConfig represents the complete configuration for the network sensor
type NetworkSensorConfig struct {
	// General settings
	InstanceID   string            `mapstructure:"instance_id"`
	LogLevel     string            `mapstructure:"log_level"`
	Environment  string            `mapstructure:"environment"` // production, development

	// Interface settings
	Interfaces   InterfaceConfig   `mapstructure:"interfaces"`

	// Capture settings
	Capture      CaptureConfig     `mapstructure:"capture"`

	// DPI settings
	DPI          DPIConfig         `mapstructure:"dpi"`

	// Threat detection settings
	ThreatDetect ThreatDetectConfig `mapstructure:"threat_detection"`

	// Output settings
	Output       OutputConfig      `mapstructure:"output"`

	// Resource limits
	Resources    ResourceConfig    `mapstructure:"resources"`
}

// InterfaceConfig defines network interface settings
type InterfaceConfig struct {
	// Interface names to capture from (empty = all interfaces)
	Interfaces []string `mapstructure:"names"`
	// Promiscuous mode
	Promiscuous bool `mapstructure:"promiscuous"`
	// Snaplen for packet capture
	Snaplen int `mapstructure:"snaplen"`
	// BPF filter expression
	BPFFilter string `mapstructure:"bpf_filter"`
	// Buffer size in bytes
	BufferSize int32 `mapstructure:"buffer_size"`
}

// CaptureConfig defines packet capture settings
type CaptureConfig struct {
	// Capture timeout in seconds
	Timeout int `mapstructure:"timeout"`
	// Immediate mode (return immediately after packet receipt)
	Immediate bool `mapstructure:"immediate"`
	// Fanout group ID for load balancing
	FanoutID int `mapstructure:"fanout_id"`
	// Fanout type: hash, lb, cpu, queue
	FanoutType string `mapstructure:"fanout_type"`
}

// DPIConfig defines Deep Packet Inspection settings
type DPIConfig struct {
	// Enable HTTP parsing
	HTTPEnabled bool `mapstructure:"http_enabled"`
	// Enable DNS parsing
	DNSEnabled bool `mapstructure:"dns_enabled"`
	// Enable TLS parsing
	TLSEnabled bool `mapstructure:"tls_enabled"`
	// Enable SMB parsing
	SMBEnabled bool `mapstructure:"smb_enabled"`
	// Enable custom protocol detection
	CustomProtocols bool `mapstructure:"custom_protocols"`
	// Maximum payload bytes to inspect
	MaxPayloadBytes int `mapstructure:"max_payload_bytes"`
	// Reassembly buffer size
	ReassemblyBufferSize int `mapstructure:"reassembly_buffer_size"`
	// Enable stream reassembly
	StreamReassembly bool `mapstructure:"stream_reassembly"`
}

// ThreatDetectConfig defines threat detection settings
type ThreatDetectConfig struct {
	// Enable threat detection
	Enabled bool `mapstructure:"enabled"`
		// Anomaly detection settings
	Anomaly AnomalyConfig `mapstructure:"anomaly"`
		// Port scan detection settings
	PortScan PortScanConfig `mapstructure:"port_scan"`
		// C2 beacon detection settings
		C2Beacon C2BeaconConfig `mapstructure:"c2_beacon"`
		// Data exfiltration detection settings
	Exfiltration ExfiltrationConfig `mapstructure:"exfiltration"`
}

// AnomalyConfig defines anomaly detection settings
type AnomalyConfig struct {
	// Enable anomaly detection
	Enabled bool `mapstructure:"enabled"`
		// Packet size threshold (bytes)
		MaxPacketSize int `mapstructure:"max_packet_size"`
		// Burst threshold (packets per second)
		BurstThreshold int `mapstructure:"burst_threshold"`
		// Time window for burst detection (seconds)
		BurstWindow int `mapstructure:"burst_window"`
}

// PortScanConfig defines port scan detection settings
type PortScanConfig struct {
	Enabled bool `mapstructure:"enabled"`
		// Unique ports threshold
		PortThreshold int `mapstructure:"port_threshold"`
		// Time window (seconds)
		Window int `mapstructure:"window"`
		// Sensitivity level: low, medium, high
		Sensitivity string `mapstructure:"sensitivity"`
}

// C2BeaconConfig defines C2 beacon detection settings
type C2BeaconConfig struct {
	Enabled bool `mapstructure:"enabled"`
		// Minimum beacon interval (seconds)
		MinInterval int `mapstructure:"min_interval"`
		// Maximum beacon interval (seconds)
		MaxInterval int `mapstructure:"max_interval"`
		// Jitter threshold (percentage)
		JitterThreshold float64 `mapstructure:"jitter_threshold"`
		// Score threshold for detection
		ScoreThreshold int `mapstructure:"score_threshold"`
}

// ExfiltrationConfig defines data exfiltration detection settings
type ExfiltrationConfig struct {
	Enabled bool `mapstructure:"enabled"`
		// Data transfer threshold (bytes per second)
		RateThreshold int64 `mapstructure:"rate_threshold"`
		// Volume threshold (bytes)
		VolumeThreshold int64 `mapstructure:"volume_threshold"`
		// Unusual hours multiplier
		UnusualHoursMultiplier float64 `mapstructure:"unusual_hours_multiplier"`
}

// OutputConfig defines output settings
type OutputConfig struct {
	// NATS configuration
	NATS NATSConfig `mapstructure:"nats"`
	// Local file output (fallback)
	File FileOutputConfig `mapstructure:"file"`
}

// NATSConfig defines NATS JetStream settings
type NATSConfig struct {
	Enabled bool `mapstructure:"enabled"`
		// NATS server URLs (comma separated)
		URLs []string `mapstructure:"urls"`
		// Subject for events
		Subject string `mapstructure:"subject"`
		// Stream name
		Stream string `mapstructure:"stream"`
		// Consumer name
		Consumer string `mapstructure:"consumer"`
		// Connection timeout
		ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
		// Reconnect wait time
		ReconnectWait time.Duration `mapstructure:"reconnect_wait"`
		// Max reconnect attempts
		MaxReconnectAttempts int `mapstructure:"max_reconnect_attempts"`
		// TLS certificate file
		CertFile string `mapstructure:"cert_file"`
		// TLS key file
		KeyFile string `mapstructure:"key_file"`
		// CA certificate file
		CACertFile string `mapstructure:"ca_cert_file"`
}

// FileOutputConfig defines file output settings
type FileOutputConfig struct {
	Enabled bool `mapstructure:"enabled"`
		// Output directory
		Directory string `mapstructure:"directory"`
		// Max file size (bytes)
		MaxSize int64 `mapstructure:"max_size"`
		// Max age (hours)
		MaxAge int `mapstructure:"max_age"`
		// Compress old files
		Compress bool `mapstructure:"compress"`
}

// ResourceConfig defines resource limits
type ResourceConfig struct {
	// Maximum CPU usage percentage
	MaxCPU int `mapstructure:"max_cpu"`
	// Maximum memory usage in MB
	MaxMemory int `mapstructure:"max_memory"`
	// Maximum goroutines
	MaxGoroutines int `mapstructure:"max_goroutines"`
	// Worker pool size
	WorkerPoolSize int `mapstructure:"worker_pool_size"`
	// Queue size for worker pool
	QueueSize int `mapstructure:"queue_size"`
	// Batch size for output
	BatchSize int `mapstructure:"batch_size"`
	// Batch flush interval
	BatchFlushInterval time.Duration `mapstructure:"batch_flush_interval"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*NetworkSensorConfig, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("sge-network-sensor")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/sge/")
		v.AddConfigPath("$HOME/.sge")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/sge")
	}

	// Enable environment variable override
	v.SetEnvPrefix("SGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	var config NetworkSensorConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Post-process configuration
	if err := postProcessConfig(&config); err != nil {
		return nil, fmt.Errorf("error post-processing config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// General defaults
	v.SetDefault("instance_id", generateInstanceID())
	v.SetDefault("log_level", "info")
	v.SetDefault("environment", "production")

	// Interface defaults
	v.SetDefault("interfaces.promiscuous", true)
	v.SetDefault("interfaces.snaplen", 1600)
	v.SetDefault("interfaces.buffer_size", 1024*1024*10) // 10MB

	// Capture defaults
	v.SetDefault("capture.timeout", 30)
	v.SetDefault("capture.immediate", false)
	v.SetDefault("capture.fanout_type", "hash")

	// DPI defaults
	v.SetDefault("dpi.http_enabled", true)
	v.SetDefault("dpi.dns_enabled", true)
	v.SetDefault("dpi.tls_enabled", true)
	v.SetDefault("dpi.smb_enabled", false)
	v.SetDefault("dpi.max_payload_bytes", 4096)
	v.SetDefault("dpi.stream_reassembly", true)

	// Threat detection defaults
	v.SetDefault("threat_detection.enabled", true)
	v.SetDefault("threat_detection.anomaly.enabled", true)
	v.SetDefault("threat_detection.anomaly.max_packet_size", 65535)
	v.SetDefault("threat_detection.anomaly.burst_threshold", 1000)
	v.SetDefault("threat_detection.anomaly.burst_window", 1)
	v.SetDefault("threat_detection.port_scan.enabled", true)
	v.SetDefault("threat_detection.port_scan.port_threshold", 100)
	v.SetDefault("threat_detection.port_scan.window", 60)
	v.SetDefault("threat_detection.port_scan.sensitivity", "medium")
	v.SetDefault("threat_detection.c2_beacon.enabled", true)
	v.SetDefault("threat_detection.c2_beacon.min_interval", 10)
	v.SetDefault("threat_detection.c2_beacon.max_interval", 300)
	v.SetDefault("threat_detection.c2_beacon.jitter_threshold", 0.2)
	v.SetDefault("threat_detection.c2_beacon.score_threshold", 70)
	v.SetDefault("threat_detection.exfiltration.enabled", true)
	v.SetDefault("threat_detection.exfiltration.rate_threshold", 10485760) // 10MB/s
	v.SetDefault("threat_detection.exfiltration.volume_threshold", 1073741824) // 1GB

	// Output defaults
	v.SetDefault("output.nats.enabled", true)
	v.SetDefault("output.nats.subject", "sge.events.network")
	v.SetDefault("output.nats.stream", "sge-network-events")
	v.SetDefault("output.nats.consumer", "sge-network-sensor")
	v.SetDefault("output.nats.connect_timeout", 10*time.Second)
	v.SetDefault("output.nats.reconnect_wait", 5*time.Second)
	v.SetDefault("output.nats.max_reconnect_attempts", 10)

	// Resource defaults
	v.SetDefault("resources.max_cpu", 2)
	v.SetDefault("resources.max_memory", 512)
	v.SetDefault("resources.max_goroutines", 100)
	v.SetDefault("resources.worker_pool_size", 10)
	v.SetDefault("resources.queue_size", 10000)
	v.SetDefault("resources.batch_size", 1000)
	v.SetDefault("resources.batch_flush_interval", 5*time.Second)
}

// postProcessConfig performs post-processing on the configuration
func postProcessConfig(config *NetworkSensorConfig) error {
	// Validate instance ID
	if config.InstanceID == "" {
		config.InstanceID = generateInstanceID()
	}

	// Parse NATS URLs from comma-separated string if needed
	if len(config.Output.NATS.URLs) == 0 {
		config.Output.NATS.URLs = []string{"nats://localhost:4222"}
	}

	// Validate resource limits
	if config.Resources.MaxCPU < 1 {
		config.Resources.MaxCPU = 1
	}
	if config.Resources.MaxCPU > 100 {
		config.Resources.MaxCPU = 100
	}
	if config.Resources.WorkerPoolSize < 1 {
		config.Resources.WorkerPoolSize = 10
	}
	if config.Resources.QueueSize < 100 {
		config.Resources.QueueSize = 10000
	}
	if config.Resources.BatchSize < 10 {
		config.Resources.BatchSize = 1000
	}

	return nil
}

// generateInstanceID generates a unique instance ID
func generateInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("sge-sensor-%s", hostname)
}

// Preset returns a configuration preset
func Preset(name string) (*NetworkSensorConfig, error) {
	v := viper.New()
	setDefaults(v)

	switch name {
	case "light":
		// Light preset for resource-constrained environments
		v.Set("resources.worker_pool_size", 4)
		v.Set("resources.queue_size", 5000)
		v.Set("resources.batch_size", 100)
		v.Set("dpi.http_enabled", true)
		v.Set("dpi.dns_enabled", true)
		v.Set("dpi.tls_enabled", false)
		v.Set("dpi.smb_enabled", false)
		v.Set("threat_detection.c2_beacon.enabled", false)
		v.Set("threat_detection.exfiltration.enabled", false)

	case "standard":
		// Standard preset (default)
		// Already configured via setDefaults

	case "aggressive":
		// Aggressive preset for high-security environments
		v.Set("resources.worker_pool_size", 20)
		v.Set("resources.queue_size", 50000)
		v.Set("resources.batch_size", 5000)
		v.Set("dpi.http_enabled", true)
		v.Set("dpi.dns_enabled", true)
		v.Set("dpi.tls_enabled", true)
		v.Set("dpi.smb_enabled", true)
		v.Set("dpi.stream_reassembly", true)
		v.Set("threat_detection.port_scan.sensitivity", "high")
		v.Set("threat_detection.c2_beacon.enabled", true)
		v.Set("threat_detection.exfiltration.enabled", true)
		v.Set("threat_detection.exfiltration.unusual_hours_multiplier", 2.0)

	default:
		return nil, fmt.Errorf("unknown preset: %s", name)
	}

	var config NetworkSensorConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling preset config: %w", err)
	}

	config.InstanceID = generateInstanceID()

	return &config, nil
}

// Save saves the configuration to a file
func (c *NetworkSensorConfig) Save(path string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("instance_id", c.InstanceID)
	v.Set("log_level", c.LogLevel)
	v.Set("environment", c.Environment)
	v.Set("interfaces", c.Interfaces)
	v.Set("capture", c.Capture)
	v.Set("dpi", c.DPI)
	v.Set("threat_detection", c.ThreatDetect)
	v.Set("output", c.Output)
	v.Set("resources", c.Resources)

	return v.SafeWriteConfigAs(path)
}
