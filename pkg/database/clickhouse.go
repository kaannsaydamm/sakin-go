package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"sakin-go/pkg/models"
)

// ClickHouseConfig, ClickHouse bağlantı ayarlarını içerir.
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	UseTLS   bool
	Debug    bool
}

// ClickHouseClient, ClickHouse bağlantı havuzunu yönetir.
type ClickHouseClient struct {
	conn   driver.Conn
	config *ClickHouseConfig
}

// NewClickHouseClient, yeni bir ClickHouse client oluşturur.
func NewClickHouseClient(config *ClickHouseConfig) (*ClickHouseClient, error) {
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Debug: config.Debug,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Second * 10,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	// TLS yapılandırması
	if config.UseTLS {
		options.TLS = &tls.Config{
			InsecureSkipVerify: false, // Production'da false olmalı
		}
	}

	// Bağlantı oluştur
	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("clickhouse connection failed: %w", err)
	}

	// Bağlantı testı
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping failed: %w", err)
	}

	return &ClickHouseClient{
		conn:   conn,
		config: config,
	}, nil
}

// Conn, aktif bağlantıyı döndürür.
func (c *ClickHouseClient) Conn() driver.Conn {
	return c.conn
}

// Ping, bağlantının sağlıklı olup olmadığını kontrol eder.
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// Close, bağlantıyı kapatır.
func (c *ClickHouseClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// InsertEvents, Event batch'ini ClickHouse'a yazar.
func (c *ClickHouseClient) InsertEvents(ctx context.Context, events []*models.Event) error {
	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return fmt.Errorf("prepare batch failed: %w", err)
	}

	for _, event := range events {
		if event == nil {
			continue
		}
		// Metadata serialization (simplified for string column, ideal use JSON)
		metaStr := "" // TODO: proper json marshal if needed

		err := batch.Append(
			event.ID,
			event.Timestamp,
			event.Source,
			event.SourceIP,
			event.DestIP,
			event.EventType,
			string(event.Severity),
			event.Description,
			event.RawLog,
			metaStr, // metadata
		)
		if err != nil {
			return fmt.Errorf("batch append failed: %w", err)
		}
	}

	return batch.Send()
}

// InsertNetworkFlows, NetworkFlow batch'ini ClickHouse'a yazar.
func (c *ClickHouseClient) InsertNetworkFlows(ctx context.Context, flows []map[string]interface{}) error {
	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO network_flows")
	if err != nil {
		return fmt.Errorf("prepare batch failed: %w", err)
	}

	for _, flow := range flows {
		err := batch.Append(
			flow["id"],
			flow["timestamp"],
			flow["source_ip"],
			flow["source_port"],
			flow["dest_ip"],
			flow["dest_port"],
			flow["protocol"],
			flow["l7_protocol"],
			flow["bytes_sent"],
			flow["bytes_received"],
			flow["packets_sent"],
			flow["packets_received"],
			flow["duration"],
			flow["flags"],
			flow["suspicious"],
		)
		if err != nil {
			return fmt.Errorf("batch append failed: %w", err)
		}
	}

	return batch.Send()
}

// Query, genel amaçlı sorgu çalıştırır.
func (c *ClickHouseClient) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}

// Exec, DML komutları çalıştırır.
func (c *ClickHouseClient) Exec(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

// InitializeSchema, gerekli ClickHouse tablolarını oluşturur.
func (c *ClickHouseClient) InitializeSchema(ctx context.Context) error {
	// Events tablosu
	eventsSchema := `
	CREATE TABLE IF NOT EXISTS events (
		id String,
		timestamp DateTime64(3),
		source String,
		source_ip String,
		dest_ip String,
		event_type String,
		severity String,
		description String,
		raw_log String,
		metadata String
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMMDD(timestamp)
	ORDER BY (timestamp, source_ip, event_type)
	TTL timestamp + INTERVAL 90 DAY
	SETTINGS index_granularity = 8192
	`

	if err := c.Exec(ctx, eventsSchema); err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}

	// Network Flows tablosu
	flowsSchema := `
	CREATE TABLE IF NOT EXISTS network_flows (
		id String,
		timestamp DateTime64(3),
		source_ip String,
		source_port UInt16,
		dest_ip String,
		dest_port UInt16,
		protocol String,
		l7_protocol String,
		bytes_sent UInt64,
		bytes_received UInt64,
		packets_sent UInt32,
		packets_received UInt32,
		duration UInt32,
		flags String,
		suspicious UInt8
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMMDD(timestamp)
	ORDER BY (timestamp, source_ip, dest_ip)
	TTL timestamp + INTERVAL 90 DAY
	SETTINGS index_granularity = 8192
	`

	if err := c.Exec(ctx, flowsSchema); err != nil {
		return fmt.Errorf("failed to create network_flows table: %w", err)
	}

	return nil
}
