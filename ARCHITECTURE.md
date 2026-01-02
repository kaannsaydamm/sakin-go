# SGE (Sakin Go Edition) Teknik Mimari Dokümantasyonu

**Version:** 1.0.0  
**Status:** RFC  
**Last Updated:** 2025-12-26

---

Bu belge, .NET tabanlı mevcut S.A.K.I.N. projesinin, Go (Golang) dili temel alınarak yeniden yazılması (rewrite) sürecindeki teknik mimariyi, bileşenleri ve tasarım kararlarını açıklar. Proje, çalışma zamanı bağımlılığı olmayan (AOT), yüksek performanslı ve dağıtık bir SIEM/SOAR mimarisini hedefler.

## 1. Mimari Genel Bakış

Sistem, mikroservis mimarisi üzerine kurgulanmış olup, bileşenler arası iletişim asenkron mesajlaşma kuyrukları üzerinden sağlanır. Tüm backend servisleri statik binary olarak derlenir ve hedef sistemlerde (Linux, Windows, macOS) ek bir runtime gerektirmeden çalışır.

### Temel Tasarım Prensipleri

- **Runtime-less Execution:** .NET Runtime veya JVM bağımlılığı yoktur.
- **High Throughput:** Saniyede 100k+ log işleme kapasitesi.
- **Zero-Copy Networking:** Ağ izleme modüllerinde çekirdek seviyesinde paket işleme.
- **Event-Driven:** Servisler arası gevşek bağlılık (loose coupling).


## 2. Teknoloji Yığını (Tech Stack)

| Katman | Teknoloji | Açıklama |
|--------|-----------|----------|
| Dil | Go (Golang) 1.22+ | Backend servisleri ve ajanlar. |
| Web Framework | Fiber | Yüksek performanslı HTTP API ve Middleware yönetimi. |
| Mesajlaşma | NATS JetStream | Düşük gecikmeli, "at-least-once" teslim garantili kuyruk yapısı. |
| Analitik DB | ClickHouse | Log ve olay verilerinin sıkıştırılarak saklanması (OLAP). |
| Metadata DB | PostgreSQL | Varlık, kullanıcı, kural ve konfigürasyon verileri (OLTP). |
| Cache/State | Redis | Oturum yönetimi, korelasyon durum takibi ve Threat Intel önbelleği. |
| Frontend | Next.js 16 (App Router) | SSR destekli yönetim paneli. |
| UI Library | shadcn/ui + Tailwind | Modüler arayüz bileşenleri. |
| Görselleştirme | Recharts | Analitik grafikler. |
| Obfuscation | garble | Anti-reverse engineering (production binaries). |


## 3. Proje Yapısı (Monorepo)

Proje monorepo yapısında yönetilecek olup, dizin hiyerarşisi aşağıdaki gibidir:

```
sge-core/
├── cmd/                          # Çalıştırılabilir Servisler (Main Packages)
│   ├── sge-agent/                # Cross-platform Veri Toplayıcı Ajan
│   ├── sge-ingest/               # Log Alım, Normalizasyon ve Zenginleştirme API'si
│   ├── sge-network-sensor/       # DPI ve Trafik Analiz Modülü
│   ├── sge-correlation/          # Kural Motoru ve Olay İlişkilendirme
│   ├── sge-analytics/            # Davranışsal Analiz (Baseline/Anomaly)
│   └── sge-soar/                 # Otomasyon ve Müdahale Servisi
├── internal/                     # Dahili Paketler (Private Library)
│   ├── parser/                   # Regex, Grok, EVTX ayrıştırıcıları
│   ├── rules/                    # Kural değerlendirme mantığı (Expr engine)
│   ├── secure-comms/             # mTLS ve Kriptografi işlemleri
│   └── middleware/               # Audit Logging, Auth, Rate Limiter
├── pkg/                          # Dışa Açık Kütüphaneler (Public Library)
│   ├── models/                   # Ortak Veri Modelleri (DTOs, Structs)
│   ├── database/                 # DB Bağlantı sürücüleri (ClickHouse, PG, Redis)
│   └── enrichment/               # Threat Intel ve GeoIP entegrasyonları
├── deployments/                  # Docker, K8s, Systemd konfigürasyonları
└── scripts/                      # Build, Test ve Deploy scriptleri
```

## 4. Servis Detayları ve Sorumluluklar

### 4.1. SGE Agent (Uç Nokta Veri Toplayıcı)

İşletim sistemi seviyesinde çalışan, kaynak tüketimi optimize edilmiş binary.

**Platform-Specific Implementations:**

- **Linux:** 
  - `auditd` socket listener for syscall monitoring
  - `syslog` integration
  - eBPF hooks for kernel-level event capture
  - File Integrity Monitoring (FIM) via inotify
  - Self-protection: `chattr +i` on configuration files

- **Windows:**
  - Event Tracing for Windows (ETW) real-time log collection
  - Registry change monitoring
  - Service watchdog for automatic restart on failure
  - Silent MSI installer support

- **macOS:**
  - Unified Logging System integration
  - Endpoint Security Framework (ESF) integration

**Güvenlik:**
- Binary obfuscation (`garble`) ile derleme ve mTLS ile sunucu iletişimi.

**Communication Protocol:**
- gRPC over mTLS
- Batch compression (LZ4/ZSTD)
- Configurable retry logic with exponential backoff

### 4.2. SGE Ingest (Veri Giriş ve İşleme)

Ajanlardan ve ağ cihazlarından gelen verinin karşılandığı nokta.

**Protokoller:**
- HTTP/JSON (Ajanlar için)
- UDP/TCP 514 (Syslog cihazları için)
- Optional: CEF, LEEF

**Zenginleştirme (Enrichment):**

- **GeoIP:** Yerel maxminddb veritabanı ile IP konumlandırma.
- **Threat Intel:** Redis önbellekli AbuseIPDB ve OTX sorguları ile IP/Hash itibar skorlaması.

**Yönlendirme:**
- Batch writes to ClickHouse (configurable batch size: 1000-10000 events)
- Connection pooling for all database connections
- Fiber framework with zero-allocation routing

**Routing:**
- İşlenen veriyi NATS JetStream üzerindeki ilgili topic'lere basar.

### 4.3. SGE Network Sensor & DPI (Derin Paket Analizi)

Ağ trafiğinin içeriğini analiz eden sensör modülüdür.

**Teknoloji:**
- **gopacket** kütüphanesi ve AF_PACKET (Linux) / Npcap (Windows).
- **Zero-Copy Capture:** Kernel'den user-space'e veri kopyalamayı minimize eden paket yakalama teknikleri.

**L7 Analizi:**

- HTTP, DNS, SMB, TLS Handshake gibi uygulama katmanı protokollerinin ayrıştırılması (Decoding).

**Tespit Kapsamı:**
- **Veri Sızıntısı (Data Exfiltration):** Anormal data boyutları
- **Yanal Hareket (Lateral Movement):** SMB, RDP aktiviteleri
- **C2 Beaconing:** Periyodik trafik paternleri
- **TLS Anomalileri:** Şüpheli sertifikalar

### 4.4. SGE Correlation (Korelasyon Motoru)

**Kural Motoru:**
- Go `expr` kütüphanesi kullanılarak derlenmiş hızında çalışan mantıksal analiz.

**Stateful Analysis:**
- Redis tabanlı "Sliding Window" (Kayan Pencere) yapısı ile zamana dayalı kural işleme.
- Örn: "X dakikada Y adet başarısız işlem"

### 4.5. SGE Analytics (İstatistiksel Analiz)

Statik kuralların ötesinde davranışsal analiz yapan servis.

**Baseline:**
- Geçmiş verileri (ClickHouse) analiz ederek kullanıcı/cihaz bazlı normal davranış profilleri oluşturur.

**Anomaly Detection:**
- Profil dışı aktiviteleri (Mesai dışı erişim, anormal veri boyutu vb.) tespit eder.

### 4.6. SGE SOAR (Otomasyon ve Müdahale)

**Playbook Execution:**
- Tanımlı YAML senaryolarını çalıştırır.

**Active Response:**
- Ajanlara komut göndererek (Process kill, IP ban, User disable) olaylara otomatik müdahale eder.

## 5. Veri Stratejisi

### 5.1. Hot Data (ClickHouse)


**Schema Design:**
- Veriler zamana göre partisyonlara ayrılır (Partitioning) ve yüksek oranda sıkıştırılır.
- Retention politikasına göre yönetilir.

### 5.2. Metadata (PostgreSQL)

Varlık envanteri, kullanıcı yetkileri, kural setleri ve playbook tanımları.

**Tables:**
- `users`, `assets`, `rules`, `playbooks`, `audit_logs`, `alerts`

### 5.3. State Store (Redis)

Dağıtık önbellek, kural durumları ve oturum bilgileri.

**Use Cases:**
1. Session Management (JWT)
2. Correlation State (Sliding window)
3. Threat Intel Cache (24h TTL)
4. Baseline Profiles
5. Rate Limiting

### 5.4. Audit Trails

Sistem üzerinde yapılan tüm yönetimsel işlemler (Kural ekleme/silme, yetki değişimi) değiştirilemez şekilde `postgres.audit_logs` tablosunda saklanır.

### 5.4. Audit Trails

**Storage:** PostgreSQL `audit_logs` table

**Schema:**
```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,  -- 'CREATE', 'UPDATE', 'DELETE'
    resource_type VARCHAR(50),    -- 'rule', 'user', 'playbook'
    resource_id INTEGER,
    ip_address INET,
    payload JSONB,                -- Full request body
    CONSTRAINT immutable_constraint CHECK (false) -- Prevents updates
);
```

**Compliance:**
- SOC 2 Type II
- ISO 27001
- GDPR (data retention policies)

## 6. Güvenlik ve Dağıtım (SecOps)

### 6.1. İletişim Güvenliği

Servisler ve ajanlar arası tüm iletişim mTLS (Mutual TLS) ile şifrelenir ve kimlik doğrulaması yapılır.

### 6.2. CI/CD Pipeline

GitHub Actions üzerinden otomatik test ve cross-platform derleme süreçleri.

**Deploy:**
- Linux için tek satırlık bash script, Windows için PowerShell tabanlı sessiz yükleyici (Silent Installer).


**Tool:** `garble` - Reverse engineering zorluğunu artırmak ve fikri mülkiyet koruması için kullanılır.


## 7. Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Event Ingestion | 100,000+ events/sec | `wrk` benchmark against `/api/v1/events` |
| End-to-End Latency | <10ms (p99) | Agent send → Dashboard display |
| DPI Throughput | 1M+ packets/sec | `tcpreplay` with PCAP files |
| DPI Accuracy | >95% protocol detection | Manual validation against known traffic |
| Correlation Latency | <100ms (p99) | Rule evaluation time |
| Anomaly False Positives | <5% | Validation against labeled dataset |
| System Uptime | 99.9% (MTBF) | Prometheus metrics |

## 8. Migration Strategy (From .NET to Go)

### 8.1. Scope

**Migrate:**
- Asset inventory (PostgreSQL `assets` table)
- User accounts and RBAC roles
- Correlation rule definitions (convert to JSON format)
- SOAR playbook definitions (convert to YAML)

**Do NOT Migrate:**
- Historical log data (archive existing ClickHouse as read-only)
- Temporary/session data

### 8.2. Migration Tool

**Implementation:** `scripts/migrate-from-dotnet.go`

**Requirements:**
- Access to source .NET PostgreSQL database
- Schema mapping configuration file
- Dry-run mode for validation

**Process:**
1. Connect to source and target databases
2. Extract data with schema validation
3. Transform data to new format (rule JSON conversion, etc.)
4. Load into target database with transaction rollback on error
5. Generate validation report (row counts, checksum verification)

**Validation:**
- Compare row counts between source and target
- Sample random records for field-by-field comparison
- Functional testing of migrated rules/playbooks

### 8.3. Rollback Plan

- Maintain source .NET system in read-only mode for 90 days
- Snapshot target database before cutover
- Document rollback procedures

**Note:** Detailed migration implementation requires access to existing .NET database schema and entity models. Current specification is framework-only pending schema review.

## 9. Operational Runbook

### 9.1. Monitoring

**Metrics Collection:**
- Prometheus exporters for all Go services
- Custom metrics: event throughput, rule latency, queue depth
- ClickHouse query performance metrics

**Alerting:**
- Alertmanager for Prometheus alerts
- PagerDuty/Opsgenie integration for on-call

**Dashboards:**
- Grafana for visualization
- Pre-built dashboards for each service

### 9.2. Backup and Recovery

**ClickHouse:**
- Incremental backups via clickhouse-backup tool
- Retention: 7 daily, 4 weekly, 12 monthly

**PostgreSQL:**
- Continuous archiving with WAL (Write-Ahead Logging)
- Point-in-time recovery (PITR) capability

**Redis:**
- RDB snapshots every 6 hours
- AOF for transaction replay

**Recovery Time Objective (RTO):** 1 hour
**Recovery Point Objective (RPO):** 15 minutes

### 9.3. Scaling Guidelines

**Horizontal Scaling:**
- `sge-ingest`: Add replicas behind load balancer
- `sge-correlation`: Partition by rule groups
- `sge-network-sensor`: Deploy per network segment

**Vertical Scaling:**
- ClickHouse: Add RAM for in-memory queries
- Redis: Increase max memory for larger caches

## 10. Future Considerations

- **Machine Learning:** Integration of anomaly detection models (TensorFlow/PyTorch)
- **Blockchain:** Immutable audit log storage via private blockchain
- **Federation:** Multi-tenant support for MSSP use cases
- **STIX/TAXII:** Threat intelligence sharing protocol support


---

**Doküman Durumu:** Bu mimari, implementasyon geri bildirimleri ve operasyonel gereksinimlere göre revize edilebilir.

**Revizyon Döngüsü:** Üç ayda bir veya major versiyon güncellemelerinde.

**Katkıda Bulunanlar:** Engineering Team, Security Operations Team, Infrastructure Team.
