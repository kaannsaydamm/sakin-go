# SGE Management Panel API 🖥️

SGE Dashboard için Backend API servisi.

## Özellikler
- **Dashboard Stats:** ClickHouse'dan gerçek zamanlı olay istatistiklerini çeker.
- **Auto Schema Init:** Başlangıçta gerekli ClickHouse tablolarını (`events`, `network_flows`) otomatik oluşturur.
- **Secure Auth:** ClickHouse ve Postgres bağlantılarında güvenli kimlik doğrulama kullanır.
- **CORS:** Frontend geliştirme ortamı (`localhost:3000`) için yapılandırılmıştır.

## Teknoloji
- **Dil:** Go 1.22+
- **Web Framework:** Fiber v2
- **Veri Tabanları:**
    - ClickHouse (OLAP - Loglar)
    - PostgreSQL (OLTP - Meta Veri)
- **Config:** `.env` dosyasından yükleme (`godotenv`).

## Yapılandırma
Aşağıdaki çevre değişkenleri `.env` dosyasında tanımlanmalıdır:

```env
PANEL_PORT=:8080
CLICKHOUSE_ADDR=127.0.0.1
CLICKHOUSE_DB=default
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=sakin123
POSTGRES_ADDR=localhost
POSTGRES_PASSWORD=sakin123
```

## Çalıştırma
```bash
go run cmd/sge-panel-api/main.go
```
