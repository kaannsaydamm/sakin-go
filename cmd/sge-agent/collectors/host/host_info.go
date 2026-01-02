package host

import (
	"encoding/json"
	"net"
	"os"
	"runtime"
	"time"
)

type HostInfo struct {
	AgentID   string    `json:"agent_id"` // Fill in main
	Hostname  string    `json:"hostname"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	NumCPU    int       `json:"num_cpu"`
	IPs       []string  `json:"ips"`
	Timestamp time.Time `json:"timestamp"`
}

func CollectEnvInfo() (*HostInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	info := &HostInfo{
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		Timestamp: time.Now().UTC(),
	}

	// IPs
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						info.IPs = append(info.IPs, ipnet.IP.String())
					}
				}
			}
		}
	}

	return info, nil
}

func (h *HostInfo) ToJSON() []byte {
	b, _ := json.Marshal(h)
	return b
}
