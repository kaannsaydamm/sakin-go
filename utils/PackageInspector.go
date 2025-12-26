package Utils

import (
	"database/sql"
	"log"
	"sync"
	"time"

	Handlers "github.com/atailh4n/sakin/handlers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Listen network traffic and monit it.
func MonitorTraffic(ifaces []pcap.Interface, db *sql.DB, wg *sync.WaitGroup) {
	for _, iface := range ifaces {
		log.Printf("Found network interface: %s\n", iface.Name)
		wg.Add(1)

		go func(ifaceName string) {
			defer wg.Done()

			// Ağ arayüzünü aç
			handle, err := pcap.OpenLive(ifaceName, 1600, true, pcap.BlockForever)
			if err != nil {
				log.Printf("Error opening device %s: %v", ifaceName, err)
				return
			}
			log.Printf("Successfully opened device %s\n", ifaceName)
			defer handle.Close()

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

			// Paketlerin her biri işlenirken log yazalım
			for packet := range packetSource.Packets() {
				log.Printf("Processing packet\n")
				if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
					ip, _ := ipLayer.(*layers.IPv4)
					log.Printf("Captured IP Packet: %s -> %s\n", ip.SrcIP, ip.DstIP)
					timestamp := time.Now()

					// Veritabanına kaydetme işlemi
					err := Handlers.SavePacket(db, ip.SrcIP.String(), ip.DstIP.String(), ip.Protocol.String(), timestamp)
					if err != nil {
						log.Printf("Error saving packet to DB: %v", err)
					}

					// Uygulama katmanı (HTTP/HTTPS)
					if packet.ApplicationLayer() != nil {
						payload := packet.ApplicationLayer().Payload()
						if len(payload) > 0 {
							log.Printf("Captured Payload: %s\n", string(payload))
						} else {
							log.Printf("Captured Encrypted HTTPS Traffic: %s -> %s\n", ip.SrcIP, ip.DstIP)
						}
					}
				} else {
					log.Printf("Non-IP packet captured\n")
				}
			}
		}(iface.Name)
	}
}
