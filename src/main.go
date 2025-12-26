package main

import (
	"log"
	"sync"

	Handlers "github.com/atailh4n/sakin/handlers"
	Utils "github.com/atailh4n/sakin/utils"
	"github.com/google/gopacket/pcap"
)

func main() {
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// Initalize Database
	db, err := Handlers.InitDB()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// Run proxy server for HTTPS listen.
	Handlers.InitProxyServer()

	// Listen network traffic and log.
	var wg sync.WaitGroup
	Utils.MonitorTraffic(ifaces, db, &wg)
	wg.Wait()
}
