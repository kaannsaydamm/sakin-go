# SGE Network Sensor 🕸️

Yüksek performanslı, `gopacket` tabanlı ağ dinleme ve analiz servisi.

## Özellikler
- **Zero-Copy Capture:** Çekirdek seviyesinde paket yakalama.
- **DPI (Deep Packet Inspection):**
    - TLS Handshake analizi ile SNI (Server Name) tespiti.
    - HTTP Header analizi.
- **Multithread:** Her ağ arayüzü (NIC) için ayrı goroutine.
- **Batched Write:** Yakalanan paketleri tamponlayıp ClickHouse'a toplu yazar.

## Konfigürasyon
Çevresel değişkenler ile yönetilir:

| Değişken | Varsayılan | Açıklama |
|----------|------------|-----------|
| `SENSOR_INTERFACE` | `eth0` | Dinlenecek ağ kartı. |
| `SENSOR_BPF` | (Boş) | BPF Filtresi (örn: `tcp port 80`). |
| `SENSOR_PROMISCUOUS` | `true` | Promiscuous modunu açar. |

## Çalıştırma

```bash
# Root yetkisi gerekebilir
sudo -E go run cmd/sge-network-sensor/main.go
```
