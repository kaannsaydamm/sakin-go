SGE (Sakin Go Edition) Teknik Mimari Dokümantasyonu

Bu belge, .NET tabanlı mevcut S.A.K.I.N. projesinin, Go (Golang) dili temel alınarak yeniden yazılması (rewrite) sürecindeki teknik mimariyi, bileşenleri ve tasarım kararlarını açıklar. Proje, çalışma zamanı bağımlılığı olmayan (AOT), yüksek performanslı ve dağıtık bir SIEM/SOAR mimarisini hedefler.

1. Mimari Genel Bakış
Sistem, mikroservis mimarisi üzerine kurgulanmış olup, bileşenler arası iletişim asenkron mesajlaşma kuyrukları üzerinden sağlanır. Tüm backend servisleri statik binary olarak derlenir ve hedef sistemlerde (Linux, Windows, macOS) ek bir runtime gerektirmeden çalışır.

Temel Tasarım Prensipleri
- Runtime-less Execution: .NET Runtime veya JVM bağımlılığı yoktur.
- High Throughput: Saniyede 100k+ log işleme kapasitesi.
- Zero-Copy Networking: Ağ izleme modüllerinde çekirdek seviyesinde paket işleme.
- Event-Driven: Servisler arası gevşek bağlılık (loose coupling).

2. Teknoloji Yığını (Tech Stack)
Katman | Teknoloji | Açıklama
Dil | Go (Golang) 1.22+ | Backend servisleri ve ajanlar.
Web Framework | Fiber | Yüksek performanslı HTTP API ve Middleware yönetimi.
Mesajlaşma | NATS JetStream | Düşük gecikmeli, "at-least-once" teslim garantili kuyruk yapısı.
Analitik DB | ClickHouse | Log ve olay verilerinin sıkıştırılarak saklanması (OLAP).
Metadata DB | PostgreSQL | Varlık, kullanıcı, kural ve konfigürasyon verileri (OLTP).
Cache/State | Redis | Oturum yönetimi, korelasyon durum takibi ve Threat Intel önbelleği.
Frontend | Next.js 16 (App Router) | SSR destekli yönetim paneli.
UI Library | shadcn/ui + Tailwind | Modüler arayüz bileşenleri.
Görselleştirme | Recharts | Analitik grafikler.

3. Proje Yapısı (Monorepo)
Proje monorepo yapısında yönetilecek olup, dizin hiyerarşisi aşağıdaki gibidir:

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

4. Servis Detayları ve Sorumluluklar

4.1. SGE Agent (Uç Nokta Veri Toplayıcı)
İşletim sistemi seviyesinde çalışan, kaynak tüketimi optimize edilmiş binary.
- Linux: auditd, syslog soket dinleme, eBPF entegrasyonu ve FIM (File Integrity Monitoring).
- Windows: ETW (Event Tracing for Windows) üzerinden gerçek zamanlı loglama ve Registry izleme.
- Güvenlik: Binary obfuscation (garble) ile derleme ve mTLS ile sunucu iletişimi.

4.2. SGE Ingest (Veri Giriş ve İşleme)
Ajanlardan ve ağ cihazlarından gelen verinin karşılandığı nokta.
- Protokoller: HTTP/JSON (Ajanlar için), UDP/TCP 514 (Syslog cihazları için).
- Zenginleştirme (Enrichment):
  * GeoIP: Yerel maxminddb veritabanı ile IP konumlandırma.
  * Threat Intel: Redis önbellekli AbuseIPDB ve OTX sorguları ile IP/Hash itibar skorlaması.
- Yönlendirme: İşlenen veriyi NATS JetStream üzerindeki ilgili topic'lere basar.

4.3. SGE Network Sensor & DPI (Derin Paket Analizi)
Ağ trafiğinin içeriğini analiz eden sensör modülüdür.
- Teknoloji: gopacket kütüphanesi ve AF_PACKET (Linux) / Npcap (Windows).
- Zero-Copy Capture: Kernel'den user-space'e veri kopyalamayı minimize eden paket yakalama teknikleri.
- L7 Analizi: HTTP, DNS, SMB, TLS Handshake gibi uygulama katmanı protokollerinin ayrıştırılması (Decoding).
- Tespit Kapsamı: Veri sızıntısı (Data Exfiltration), C2 beaconing, şüpheli TLS sertifikaları ve yanal hareketler (Lateral Movement).

4.4. SGE Correlation (Korelasyon Motoru)
- Kural Motoru: Go expr kütüphanesi kullanılarak derlenmiş hızında çalışan mantıksal analiz.
- Stateful Analysis: Redis tabanlı "Sliding Window" (Kayan Pencere) yapısı ile zamana dayalı kural işleme (Örn: "X dakikada Y adet başarısız işlem").

4.5. SGE Analytics (İstatistiksel Analiz)
Statik kuralların ötesinde davranışsal analiz yapan servis.
- Baseline: Geçmiş verileri (ClickHouse) analiz ederek kullanıcı/cihaz bazlı normal davranış profilleri oluşturur.
- Anomaly Detection: Profil dışı aktiviteleri (Mesai dışı erişim, anormal veri boyutu vb.) tespit eder.

4.6. SGE SOAR (Otomasyon ve Müdahale)
- Playbook Execution: Tanımlı YAML senaryolarını çalıştırır.
- Active Response: Ajanlara komut göndererek (Process kill, IP ban, User disable) olaylara otomatik müdahale eder.

5. Veri Stratejisi
- Hot Data (ClickHouse): Analiz ve log verileri. Veriler zamana göre partisyonlara ayrılır (Partitioning) ve yüksek oranda sıkıştırılır. Retention politikasına göre yönetilir.
- Metadata (PostgreSQL): Varlık envanteri, kullanıcı yetkileri, kural setleri ve playbook tanımları.
- State Store (Redis): Dağıtık önbellek, kural durumları ve oturum bilgileri.
- Audit Trails: Sistem üzerinde yapılan tüm yönetimsel işlemler (Kural ekleme/silme, yetki değişimi) değiştirilemez şekilde postgres.audit_logs tablosunda saklanır.

6. Güvenlik ve Dağıtım (SecOps)
- İletişim Güvenliği: Servisler ve ajanlar arası tüm iletişim mTLS (Mutual TLS) ile şifrelenir ve kimlik doğrulaması yapılır.
- CI/CD: GitHub Actions üzerinden otomatik test ve cross-platform derleme süreçleri.
- Deploy: Linux için tek satırlık bash script, Windows için PowerShell tabanlı sessiz yükleyici (Silent Installer).
