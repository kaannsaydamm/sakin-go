package Handlers

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// Function to handle HTTP connection (simple response for now)
func handleHTTPConnection(conn net.Conn) {
	defer conn.Close()

	// Read the HTTP request from the client
	buffer := make([]byte, 4096)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading HTTP request:", err)
		return
	}

	// Simulate a successful response for the sake of the example
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHTTP Request Handled"))
}

// Function to handle HTTPS (TLS) proxy connections
func handleHTTPSConnection(conn net.Conn) {
	defer conn.Close()

	// Read the CONNECT request from the client
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from client:", err)
		return
	}

	// Log the received data for debugging
	requestLine := string(buffer[:n])
	log.Printf("Received CONNECT request: %s", requestLine)

	// Extract the target host and port from the CONNECT request line
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		log.Println("Invalid CONNECT request, missing target host and port")
		return
	}

	targetHost := parts[1] // Target host should be in the form "hostname:port"
	log.Printf("Target host extracted: %s", targetHost)

	// Generate the certificate for the target host using your MITM CA
	cert, err := GenerateCertificateForHost(targetHost)
	if err != nil {
		log.Println("Error generating certificate for target:", err)
		return
	}

	// Send "200 Connection Established" response to the client
	_, err = fmt.Fprintf(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		log.Println("Error sending response to client:", err)
		return
	}

	// Establish a TLS connection to the target host using the generated certificate
	destConn, err := tls.Dial("tcp", targetHost, &tls.Config{
		InsecureSkipVerify: true,                     // Skipping server cert validation for MITM
		Certificates:       []tls.Certificate{*cert}, // Use the dynamically generated cert
	})
	if err != nil {
		log.Println("Error connecting to target server:", err)
		return
	}
	defer destConn.Close()

	// Proxy traffic between the client and the destination server
	go proxyTraffic(conn, destConn)
	go proxyTraffic(destConn, conn)
}

// Function to proxy traffic between two connections (client <-> server)
func proxyTraffic(src, dest net.Conn) {
	buffer := make([]byte, 4096) // Buffer to hold data

	for {
		// Read data from the source connection
		n, err := src.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client/server")
			} else {
				log.Println("Error reading data:", err)
			}
			return
		}

		// Write data to the destination connection
		_, err = dest.Write(buffer[:n])
		if err != nil {
			log.Println("Error writing data:", err)
			return
		}
	}
}

// Handle incoming TCP connections and decide HTTP or HTTPS
func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// Read the first few bytes from the connection to determine the protocol
	buffer := make([]byte, 512)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from connection:", err)
		return
	}

	log.Printf("Received initial data: %s", string(buffer)) // Debug log

	// Check if the connection is a CONNECT (HTTPS) request
	// CONNECT request starts with "CONNECT"
	if strings.HasPrefix(string(buffer), "CONNECT") {
		log.Println("Handling HTTPS connection")
		handleHTTPSConnection(conn)
	} else if strings.HasPrefix(string(buffer), "GET") || strings.HasPrefix(string(buffer), "POST") || strings.HasPrefix(string(buffer), "HEAD") {
		// Handle HTTP connection
		log.Println("Handling HTTP connection")
		handleHTTPConnection(conn)
	} else {
		log.Println("Unknown connection type, closing.")
	}
}

// Proxy server initialization
func InitProxyServer() {
	// Start TCP listener on port 9009
	listen, err := net.Listen("tcp", ":9009")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listen.Close()

	log.Println("Proxy server listening on port 9009...")

	// Accept incoming TCP connections
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleTCPConnection(conn)
	}
}
