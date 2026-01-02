package communicator

import (
	"context"
	"fmt"
	"log"
	"time"

	securecomms "sakin-go/internal/secure-comms"

	"github.com/nats-io/nats.go"

	"sakin-go/cmd/sge-agent/config"
)

type Communicator struct {
	config *config.AgentConfig
	nc     *nats.Conn
}

func NewCommunicator(cfg *config.AgentConfig) (*Communicator, error) {
	// 1. Load mTLS Config
	// Use secure-comms/CertManager logic to load specific files
	// We create a temp manager just to use its utility or simple loader
	// Ideally we refactor utils to handle this without manager instance,
	// but for now we can simple use LoadTLSConfig if we export it or use the one in utils if added.
	// Actually, let's look at internal/secure-comms/cert_manager.go -> LoadTLSConfig is a method of CertManager.
	// We should probably make a standalone loader or just instantiate a dummy manager.
	// Better: Use `securecomms.LoadTLSConfig` if I exposed it as a function not method, or just use `tls.LoadX509KeyPair` + `x509.NewCertPool`.

	// Let's implement the TLS loading manually here for simplicity to avoid instantiating a full manager logic which expects a directory structure
	// Or even better, let's add a helper to secure-comms/utils.go if not present.
	// Checked utils: It has `NewMTLSHTTPClient`, `LoadCAPool`.

	// We will accept the limitation and implement TLS loading here using the utils helper.
	certManager, _ := securecomms.NewCertManager(".") // Dummy dir
	tlsConfig, err := certManager.LoadTLSConfig(cfg.CertFile, cfg.KeyFile, cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load mTLS config: %w", err)
	}

	// 2. Connect NATS
	opts := []nats.Option{
		nats.Secure(tlsConfig),
		nats.Name("SGE-Agent-" + cfg.AgentID),
		nats.ReconnectWait(5 * time.Second),
		nats.MaxReconnects(-1), // Infinite reconnects
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("[Communicator] Disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Printf("[Communicator] Reconnected")
		}),
	}

	nc, err := nats.Connect(cfg.ServerURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect failed: %w", err)
	}

	return &Communicator{
		config: cfg,
		nc:     nc,
	}, nil
}

func (c *Communicator) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *Communicator) Publish(subject string, data []byte) error {
	return c.nc.Publish(subject, data)
}

func (c *Communicator) SubscribeCommands(ctx context.Context, handler func(cmd []byte)) error {
	topic := "commands." + c.config.AgentID
	_, err := c.nc.Subscribe(topic, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	return err
}
