# SGE Enrichment Service 💎

Olay zenginleştirme servisi.

## Özellikler
- **GeoIP:** IP adreslerinin coğrafi konumunu (Ülke, Şehir, Koordinat) ekler.
- **Threat Intel:** IP adreslerini AbuseIPDB vb. veritabanlarında sorgular (Redis Cache destekli).
- **Severity Escalation:** Zararlı IP tespit edilirse olayın seviyesini otomatik `Critical` yapar.

## Gereksinimler
- MaxMind `GeoLite2-City.mmdb` dosyası (Opsiyonel, yoksa GeoIP devre dışı kalır).

## Çalıştırma
```bash
go run cmd/sge-enrichment/main.go
```
