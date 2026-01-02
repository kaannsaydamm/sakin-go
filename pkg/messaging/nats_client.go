package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NatsConfig holds configuration for NATS connection.
type NatsConfig struct {
	URL      string
	Username string
	Password string
	// MaxReconnects sets the number of reconnect attempts
	MaxReconnects int
	// ReconnectWait sets the time to wait between reconnect attempts
	ReconnectWait time.Duration
}

// Client wraps the NATS connection and JetStream context.
type Client struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// NewClient creates a new optimized NATS client with JetStream support.
func NewClient(config *NatsConfig) (*Client, error) {
	opts := []nats.Option{
		nats.Name("SGE-Backend"),
		nats.ReconnectWait(config.ReconnectWait),
		nats.MaxReconnects(config.MaxReconnects),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			// Log disconnection?
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			// Log reconnection?
		}),
	}

	if config.Username != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.Username, config.Password))
	}

	nc, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect failed: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream init failed: %w", err)
	}

	return &Client{
		nc: nc,
		js: js,
	}, nil
}

// Close closes the NATS connection.
func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

// Connection returns the underlying NATS connection.
func (c *Client) Connection() *nats.Conn {
	return c.nc
}

// JetStream returns the JetStream context.
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

// PublishAsync publishes a message asynchronously to JetStream.
// This is non-blocking and highly performant. A future is returned to check status if needed.
func (c *Client) PublishAsync(ctx context.Context, subject string, data []byte) (jetstream.PubAckFuture, error) {
	// PublishAsync is the key for high throughput.
	// It doesn't wait for the server to acknowledge receipt.
	return c.js.PublishAsync(subject, data)
}

// PublishSync publishes a message synchronously.
// Use this only when delivery guarantee is critical before proceeding.
func (c *Client) PublishSync(ctx context.Context, subject string, data []byte) (*jetstream.PubAck, error) {
	return c.js.Publish(ctx, subject, data)
}

// Subscribe is a wrapper for simple Pull Consumer (worker pattern).
// It creates a Durable Consumer with FilterSubject and DeliverGroup (Queue).
func (c *Client) QueueSubscribe(ctx context.Context, stream, subject, queueGroup string, handler func(msg jetstream.Msg)) (jetstream.ConsumeContext, error) {
	// 1. Create/Update Consumer
	// Name must he unique for the queue group
	consumerName := queueGroup

	cfg := jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: subject,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	}

	cons, err := c.js.CreateOrUpdateConsumer(ctx, stream, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// 2. Consume (Pull)
	// This starts a goroutine that pulls messages and calls handler
	cc, err := cons.Consume(handler)
	if err != nil {
		return nil, fmt.Errorf("consume failed: %w", err)
	}

	return cc, nil
}

// InitializeStreams creates the necessary JetStream streams if they don't exist.
// Configured for high performance (File storage, defined retention).
func (c *Client) InitializeStreams(ctx context.Context) error {
	// Events Stream
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        StreamEvents,
		Description: "SGE Security Events Stream",
		Subjects:    []string{"events.>"},
		Retention:   jetstream.WorkQueuePolicy, // Using WorkQueue for processing
		Storage:     jetstream.FileStorage,
		Replicas:    1,              // Increase for HA
		MaxAge:      24 * time.Hour, // Keep events for 24h in stream buffer
	})
	if err != nil {
		return fmt.Errorf("failed to create events stream: %w", err)
	}

	// Alerts Stream
	_, err = c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        StreamAlerts,
		Description: "SGE Generated Alerts",
		Subjects:    []string{"alerts.>"},
		Retention:   jetstream.LimitsPolicy, // Keep alerts available for history
		Storage:     jetstream.FileStorage,
		MaxAge:      7 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("failed to create alerts stream: %w", err)
	}

	return nil
}
