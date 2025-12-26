/**
* DatabaseHandler.go - Handler for MySQL
* Written by: atailh4n
 */

package Handlers

import (
	"database/sql"
	"fmt"
	"time"
)

// Initalize DB
func InitDB() (*sql.DB, error) {
	dsn := "root@tcp(127.0.0.1:3306)/network_db"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Create DB
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS network_db")
	if err != nil {
		return nil, fmt.Errorf("database creation error: %v", err)
	}

	// Select it
	_, err = db.Exec("USE network_db")
	if err != nil {
		return nil, fmt.Errorf("database selection error: %v", err)
	}

	// Create table
	createTableQuery := `
	 CREATE TABLE IF NOT EXISTS packets (
		 id INT AUTO_INCREMENT PRIMARY KEY,
		 src_ip VARCHAR(15),
		 dst_ip VARCHAR(15),
		 protocol VARCHAR(10),
		 timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	 )`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("table creation error : %v", err)
	}

	return db, nil
}

// Save packages to DB
func SavePacket(db *sql.DB, srcIP, dstIP, protocol string, timestamp time.Time) error {
	query := "INSERT INTO packets (src_ip, dst_ip, protocol, timestamp) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, srcIP, dstIP, protocol, timestamp)
	return err
}
