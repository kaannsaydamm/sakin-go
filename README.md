# SGE (Sakin Go Edition) 🛡️

**S.A.K.I.N.** (Siber Analiz ve Karar İstihbarat Ağı) - Yeni Nesil Go Tabanlı SIEM/SOAR Platformu.

Bu proje, orijinal C# `sakin-core` mimarisinin **Go (Golang)** diline port edilmiş, yüksek performanslı, dağıtık ve bulut tabanlı (Cloud Native) versiyonudur. "Milyonlarca veriyi en düşük donanımda bile işleme" felsefesiyle tasarlanmıştır.

## 🚀 Özellikler

- **Yüksek Performans:** Go'nun concurrency modeli ve Zero-Allocation teknikleri ile minimum RAM kullanımı.
- **Modern Mimari:** NATS JetStream tabanlı Event-Driven mikroservis yapısı.
- **Cross-Platform:** Hem **Linux** hem **Windows** üzerinde çalışabilen Agent ve Server bileşenleri.
- **Tam Kapsamlı Güvenlik:**
    - **Network Sensor:** `gopacket` ile DPI (Deep Packet Inspection), TLS SNI yakalama.
    - **Correlation:** `expr` tabanlı dinamik kural motoru.
    - **Enrichment:** GeoIP ve Threat Intel zenginleştirme cache katmanı.
    - **Analytics:** ClickHouse üzerinde Big Data analitiği ve Baseline tespiti.
    - **SOAR:** Otomatik aksiyon ve olay müdahale (Playbooks).
- **Gelişmiş Yönetim:**
    - **Panel:** Next.js ve Go Fiber tabanlı modern Web Arayüzü.
    - **TUI:** Terminal üzerinden anlık sistem izleme aracı.
    - **Health Check:** CLI tabanlı sağlık kontrolü.

## 🏗️ Mimari Bileşenler

| Servis | Tanım | Teknoloji |
|--------|-------|-----------|
| `sge-network-sensor` | Ağ trafiğini dinler, analiz eder ve loglar. | gopacket, pcap |
| `sge-agent` | Uç noktalardan (Linux/Windows) log toplar. | mTLS, Auditd, ETW |
| `sge-ingest` | Agent ve Syslog verilerini karşılayan API Gateway. | Fiber, NATS |
| `sge-correlation` | Gerçek zamanlı kural eşleştirme ve alarm üretme. | expr-lang |
| `sge-enrichment` | Eventleri GeoIP ve İstihbarat verisiyle zenginleştirir. | Redis, MaxMind |
| `sge-analytics` | Verileri ClickHouse'a yazar ve istatistik çıkarır. | ClickHouse |
| `sge-soar` | Alarmlara otomatik tepki verir (IP Bloklama vb.). | Command Pattern |
| `sge-panel-api` | UI için Backend API. | Fiber, JWT |

## 🛠️ Kurulum ve Çalıştırma

### Gereksinimler
- **Go** 1.22+
- **Docker** & **Docker Compose**
- **Linux:** `libpcap-dev` | **Windows:** `Npcap`

### 1. Hazırlık
Geliştirme ortamını hazırlamak için ilgili scripti çalıştırın:

**Linux / macOS:**
```bash
./scripts/setup.sh
```

**Windows (PowerShell):**
```powershell
.\scripts\setup.ps1
```

### ❗ Önemli: Konfigürasyon
Proje kök dizinindeki `.env` dosyasını kendi ortamınıza göre düzenleyin.
Manuel çalıştırmalarda çevresel değişkenlerin yüklü olduğundan emin olun (örn: `set -a; source .env; set +a` veya IDE ayarları).

### 2. Başlatma
Tüm altyapıyı (NATS, Veritabanları) ve servisleri başlatmak için:

**Linux:**
```bash
./scripts/manage.sh start
```

**Windows:**
```powershell
.\scripts\manage.ps1 start
```

### 3. Erişim
Sistem açıldığında aşağıdaki adreslerden erişebilirsiniz:

- **Web Panel:** `http://localhost:3000`
- **Panel API:** `http://localhost:8080`
- **ClickHouse:** `http://localhost:8123`

### 4. Araçlar
Terminal arayüzü ile sistemi izlemek için:
```bash
go run cmd/sge-tui/main.go
```

Sistem sağlığını kontrol etmek için:
```bash
go run cmd/sge-health/main.go
```

## 📂 Dizin Yapısı

```
sakin-go/
├── cmd/                # Servislerin kaynak kodları (main entry points)
│   ├── sge-agent/
│   ├── sge-network-sensor/
│   └── ...
├── pkg/                # Paylaşılan kütüphaneler (DB, Messaging, Models)
├── internal/           # Dahili paketler (Secure Comms)
├── web/                # Frontend (Next.js) projeleri
├── scripts/            # Yönetim scriptleri (.sh, .ps1)
└── docker-compose.yml  # Altyapı tanımları
```

## 🤝 Katkıda Bulunma
Bu proje açık kaynaklıdır ve topluluk katkılarına açıktır. Lütfen `IMPLEMENTATION.md` dosyasındaki yol haritasını inceleyin.

---
*Deepmind Agentic Coding Team tarafından tasarlanmıştır.*
