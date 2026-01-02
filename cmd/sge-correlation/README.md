# SGE Correlation Engine 🧠

Gerçek zamanlı olay korelasyon ve alarm üretme motoru.

## Nasıl Çalışır?
1. NATS üzerinden `events.raw` akışını dinler.
2. Belleğe yüklenen kuralları (Rules) her gelen olay için değerlendirir.
3. Kural eşleşirse `Alert` üretir ve `alerts` kanalına basar.

## Kural Mantığı
Kurallar `expr` dili ile yazılır. C# LINQ benzeri esnek bir sözdizimi vardır.

**Örnek:**
```javascript
Event.Severity == 'critical' && Event.Source in ['firewall', 'ips']
```

## Çalıştırma
```bash
go run cmd/sge-correlation/main.go
```
