/**
* DatabaseHandler.go - Handler for MySQL
* Written by: atailh4n
 */

package Handlers

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// InitializeDB creates a connection to the MySQL database
func InitializeDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}

// CreateNetworkDB creates the network database and required tables
func CreateNetworkDB(db *sql.DB) error {
	// Create database
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS network_db")
	if err != nil {
		return fmt.Errorf("database creation error: %w", err)
	}

	// Select it
	_, err = db.Exec("USE network_db")
	if err != nil {
		return fmt.Errorf("database selection error: %w", err)
	}

	// Create packets table
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS packets (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			src_ip VARCHAR(45),
			dst_ip VARCHAR(45),
			protocol VARCHAR(10),
			payload_size INT,
			timestamp TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6),
			INDEX idx_src_ip (src_ip),
			INDEX idx_dst_ip (dst_ip),
			INDEX idx_protocol (protocol),
			INDEX idx_timestamp (timestamp)
		)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("table creation error: %w", err)
	}

	// Create threats table for detected threats
	createThreatsTable := `
		CREATE TABLE IF NOT EXISTS threats (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			threat_type VARCHAR(50),
			severity VARCHAR(20),
			src_ip VARCHAR(45),
			dst_ip VARCHAR(45),
			src_port INT,
			dst_port INT,
			protocol VARCHAR(10),
			description TEXT,
			score INT,
			timestamp TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6),
			INDEX idx_threat_type (threat_type),
			INDEX idx_severity (severity),
			INDEX idx_src_ip (src_ip),
			INDEX idx_timestamp (timestamp)
		)`
	_, err = db.Exec(createThreatsTable)
	if err != nil {
		return fmt.Errorf("threats table creation error: %w", err)
	}

	return nil
}

// SavePacket saves a packet to the database
func SavePacket(db *sql.DB, srcIP, dstIP, protocol string, payloadSize int, timestamp time.Time) error {
	query := "INSERT INTO packets (src_ip, dst_ip, protocol, payload_size, timestamp) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, srcIP, dstIP, protocol, payloadSize, timestamp)
	return err
}

// SaveThreat saves a detected threat to the database
func SaveThreat(db *sql.DB, threatType, severity, srcIP, dstIP, protocol, description string, srcPort, dstPort, score int, timestamp time.Time) error {
	query := "INSERT INTO threats (threat_type, severity, src_ip, dst_ip, src_port, dst_port, protocol, description, score, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, threatType, severity, srcIP, dstIP, srcPort, dstPort, protocol, description, score, timestamp)
	return err
}

// GetPacketCount returns the total number of packets
func GetPacketCount(db *sql.DB) (int64, error) {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM packets").Scan(&count)
	return count, err
}

// GetThreatCount returns the total number of threats by type
func GetThreatCount(db *sql.DB, threatType string) (int64, error) {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM threats WHERE threat_type = ?", threatType).Scan(&count)
	return count, err
}

// CloseDB closes the database connection
func CloseDB(db *sql.DB) error {
	return db.Close()
}
