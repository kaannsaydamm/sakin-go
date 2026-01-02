# SGE Ingest Service 📥

Merkezi log toplama ve API Gateway servisi.

## Özellikler
- **High Performance API:** Go Fiber ile saniyede binlerce istek karşılama.
- **Normalization:** Farklı kaynaklardan (Agent, Syslog) gelen veriyi standart `Event` formatına çevirir.
- **Async Streaming:** Veriyi diske yazmak yerine doğrudan NATS JetStream'e basar.

## API Endpoints

### `POST /api/v1/events`
Agent'lardan bulk event alır.

**Örnek İstek:**
```json
{
  "source": "syslog",
  "severity": "info",
  "message": "SSH login successful"
}
```

## Çalıştırma
```bash
go run cmd/sge-ingest/main.go
```
