package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig, PostgreSQL bağlantı ayarlarını içerir.
type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string // disable, require, verify-ca, verify-full
}

// PostgresClient, PostgreSQL bağlantı havuzunu yönetir.
type PostgresClient struct {
	db     *sql.DB
	config *PostgresConfig
}

// NewPostgresClient, yeni bir PostgreSQL client oluşturur.
func NewPostgresClient(config *PostgresConfig) (*PostgresClient, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}

	// Connection pool ayarları
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Bağlantı testi
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return &PostgresClient{
		db:     db,
		config: config,
	}, nil
}

// GetDB, *sql.DB instance'ını döndürür.
func (p *PostgresClient) GetDB() *sql.DB {
	return p.db
}

// Ping, bağlantının sağlıklı olup olmadığını kontrol eder.
func (p *PostgresClient) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close, bağlantıyı kapatır.
func (p *PostgresClient) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Query, sorgu çalıştırır ve satırları döndürür.
func (p *PostgresClient) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow, tek satır döndüren sorgu çalıştırır.
func (p *PostgresClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// Exec, DML komutları çalıştırır.
func (p *PostgresClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// BeginTx, yeni bir transaction başlatır.
func (p *PostgresClient) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// InitializeSchema, gerekli PostgreSQL tablolarını oluşturur.
func (p *PostgresClient) InitializeSchema(ctx context.Context) error {
	schema := `
	-- Users tablosu
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'viewer',
		enabled BOOLEAN NOT NULL DEFAULT true,
		permissions TEXT[] DEFAULT '{}',
		last_login TIMESTAMPTZ,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Assets tablosu
	CREATE TABLE IF NOT EXISTS assets (
		id SERIAL PRIMARY KEY,
		type VARCHAR(50) NOT NULL,
		name VARCHAR(255) NOT NULL,
		ip_address INET,
		os VARCHAR(100),
		location VARCHAR(255),
		tags TEXT[] DEFAULT '{}',
		status VARCHAR(50) DEFAULT 'active',
		last_seen TIMESTAMPTZ,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Rules tablosu
	CREATE TABLE IF NOT EXISTS rules (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		enabled BOOLEAN NOT NULL DEFAULT true,
		severity VARCHAR(50) NOT NULL,
		expression TEXT NOT NULL,
		time_window INTEGER DEFAULT 60,
		threshold INTEGER DEFAULT 1,
		actions TEXT[] DEFAULT '{}',
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Alerts tablosu
	CREATE TABLE IF NOT EXISTS alerts (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		rule_id INTEGER REFERENCES rules(id) ON DELETE CASCADE,
		rule_name VARCHAR(255) NOT NULL,
		severity VARCHAR(50) NOT NULL,
		description TEXT,
		event_ids TEXT[] DEFAULT '{}',
		affected_assets TEXT[] DEFAULT '{}',
		status VARCHAR(50) DEFAULT 'open',
		assigned_to VARCHAR(100),
		mitre_techniques TEXT[] DEFAULT '{}',
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Playbooks tablosu
	CREATE TABLE IF NOT EXISTS playbooks (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		trigger VARCHAR(255),
		enabled BOOLEAN NOT NULL DEFAULT true,
		steps JSONB NOT NULL,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	-- Audit Logs tablosu (değiştirilemez)
	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		user_id INTEGER REFERENCES users(id),
		action VARCHAR(50) NOT NULL,
		resource_type VARCHAR(50),
		resource_id VARCHAR(100),
		ip_address INET,
		changes JSONB
	);

	-- Audit logs için trigger (UPDATE ve DELETE engelle)
	CREATE OR REPLACE FUNCTION prevent_audit_log_modifications()
	RETURNS TRIGGER AS $$
	BEGIN
		RAISE EXCEPTION 'Audit logs cannot be modified or deleted';
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS audit_logs_immutable ON audit_logs;
	CREATE TRIGGER audit_logs_immutable
	BEFORE UPDATE OR DELETE ON audit_logs
	FOR EACH ROW
	EXECUTE FUNCTION prevent_audit_log_modifications();

	-- İndeksler
	CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
	CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
	CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_assets_ip_address ON assets(ip_address);
	CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
	CREATE INDEX IF NOT EXISTS idx_rules_enabled ON rules(enabled);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);

	-- GIN indeksler (JSONB ve array kolonları için)
	CREATE INDEX IF NOT EXISTS idx_alerts_metadata ON alerts USING GIN(metadata);
	CREATE INDEX IF NOT EXISTS idx_assets_tags ON assets USING GIN(tags);
	CREATE INDEX IF NOT EXISTS idx_rules_metadata ON rules USING GIN(metadata);
	`

	_, err := p.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Health, database sağlık durumunu döndürür.
func (p *PostgresClient) Health(ctx context.Context) (map[string]string, error) {
	var version string
	err := p.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return nil, err
	}

	stats := p.db.Stats()

	return map[string]string{
		"status":           "healthy",
		"version":          version,
		"open_connections": fmt.Sprintf("%d", stats.OpenConnections),
		"in_use":           fmt.Sprintf("%d", stats.InUse),
		"idle":             fmt.Sprintf("%d", stats.Idle),
	}, nil
}
