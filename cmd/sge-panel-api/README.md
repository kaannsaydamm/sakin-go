# SGE Management Panel API 🖥️

SGE Dashboard için Backend API servisi.

## Özellikler
- **Dashboard Stats:** ClickHouse'dan gerçek zamanlı olay istatistiklerini çeker.
- **Alert Management:** Postgres üzerindeki alarmları listeler ve yönetir.
- **Agent Command:** Agent'lara komut göndermek için NATS ile konuşur.

## Teknoloji
- Loglama ve İstatistikler için **ClickHouse**.
- İlişkisel veriler (Kullanıcılar, Kurallar) için **PostgreSQL**.
- HTTP Sunucusu için **Go Fiber**.

## Çalıştırma
```bash
go run cmd/sge-panel-api/main.go
```
