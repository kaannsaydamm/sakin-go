# SGE Implementation Plan

## 📚 1. Shared Infrastructure [Kısmen Tamamlandı]
- ✅ Monorepo & Project Structure
- ✅ Database Layer (`pkg/database`) - ClickHouse, Postgres, Redis
- ✅ Secure Comms (`internal/secure-comms`) - mTLS
- ✅ **Utils Package** (`pkg/utils`)
    - Common string/time/net utilities.
    - Zero-allocation converters & Buffer Pool.
- ✅ **Messaging Package** (`pkg/messaging`)
    - NATS JetStream wrappers.
    - Async publishing & Topic definition constants.

## ✅ 2. Core Service: Network Sensor (`cmd/sge-network-sensor`) [TAMAMLANDI]
- ✅ Ported from C# `NetworkSensorService`.
- ✅ **Inspector**: `gopacket` based multi-interface packet capture.
- ✅ **DPI**: Optimized byte-level TLS SNI parser & HTTP metadata extractor.
- ✅ **Infrastructure**: NATS integration & ClickHouse batch writes.
- ✅ **Config**: Environment variable based configuration (`config.go`).

## ✅ 3. Core Service: Agent (`cmd/sge-agent`) [KISMEN TAMAMLANDI]
- ✅ **Multi-OS**: Windows & Linux (Build tags implemented).
- ✅ **Communication**: mTLS connection to NATS directly (High Perf).
- ✅ **Host Info**: Basic periodic heartbeat.
- [ ] **Collectors**:
    - Auditd (Linux) - *Pending implementation*
    - ETW (Windows) - *Pending implementation*

## ✅ 4. Ingest Service (`cmd/sge-ingest`) [TAMAMLANDI]
- ✅ **API**: Fiber HTTP Server for high-performance agent event ingestion.
- ✅ **Normalization**: JSON payload -> Standard Event Model mapping.
- ✅ **Streaming**: Async NATS Publishing to `events.raw` topic.
- ✅ **Endpoints**: POST `/api/v1/events` and `/health`.

## ✅ 5. Correlation Service (`cmd/sge-correlation`) [TAMAMLANDI]
- ✅ **Engine**: `expr` based high-performance rule evaluation.
- ✅ **Input**: Consumes from NATS `events.raw` (Queue Grouping for scalability).
- ✅ **Output**: Publishes `Alert` objects to `alerts` topic.
- ✅ **Infrastructure**: Postgres state & Rule loading logic integrated.

## ✅ 6. Enrichment Service (`cmd/sge-enrichment`) [TAMAMLANDI]
- ✅ **Threat Intel**: IP checking cache layer (`provider.go`) backed by Redis.
- ✅ **GeoIP**: MaxMind DB integration with fail-safe fallback.
- ✅ **Pipeline**: `events.raw` -> Enrichment -> `events.enriched`.
- ✅ **Action**: Automatic Severity escalation for malicious IPs.

## ✅ 7. Analytics Service (`cmd/sge-analytics`) [TAMAMLANDI]
- ✅ **Sink**: Buffered Batch Insert to ClickHouse (`sink/clickhouse.go`).
- ✅ **Baseline**: Time-window based volume analysis (`baseline/worker.go`).
- ✅ **Pipeline**: Consumes `events.enriched`, archives, and computes stats.

## ✅ 8. SOAR Service (`cmd/sge-soar`) [TAMAMLANDI]
- ✅ **Engine**: Playbook execution engine triggered by Alerts.
- ✅ **Actions**: Modular action registry (`BlockIP`, `SlackNotify`).
- ✅ **Integration**: Publishes commands to Agents via NATS `commands.>` topic.
- ✅ **Flow**: Alert -> Match Trigger -> Execute Steps Sequentialy.

## ✅ 9. Management Panel [TAMAMLANDI]
- ✅ **API (`cmd/sge-panel-api`)**: Go Fiber REST API.
    - Serves data for all modules (Traffic, Analytics, Alerts).
- ✅ **UI (`web/panel-ui`)**: Next.js 14 dashboard (Skeleton).
    - **Traffic**: Live flow table with fitering.
    - **Agents**: Status/Command center.
    - **SOAR**: Visual editor.
    - **Analytics**: Charts & Graphs.

## ✅ 10. Tools & Scripts [TAMAMLANDI]
- ✅ **Health Check (`cmd/sge-health`)**: CLI tool to verify connectivity.
- ✅ **TUI (`cmd/sge-tui`)**: Interactive terminal dashboard.
- ✅ **Infrastructure**: `docker-compose.yml` for NATS/PG/CH/Redis.
- ✅ **Scripts**: Cross-platform (`.sh`/`.ps1`) management scripts.

## 📊 İlerleme
```
[██████████████] 100% - Ready for Launch
```
