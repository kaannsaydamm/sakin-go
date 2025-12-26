// Package storage provides data storage and messaging capabilities for the SGE Network Sensor.
// This package implements NATS JetStream producer for event streaming.
package storage

import (
    "context"
    "errors"
    "fmt"
    "math"
    "os"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/atailh4n/sakin/internal/config"
    "github.com/atailh4n/sakin/internal/normalization"
    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// NATSProducer sends network events to NATS JetStream
type NATSProducer struct {
    cfg            *config.NATSConfig
    conn           *nats.Conn
    js             jetstream.JetStream
    producer       jetstream.Producer
    metrics        *ProducerMetrics
    batchCh        chan *normalization.NetworkEvent
    batchBuf       []*normalization.NetworkEvent
    mu             sync.Mutex
    wg             sync.WaitGroup
    ctx            context.Context
    cancel         context.CancelFunc
    stopCh         chan struct{}
    retryMu        sync.RWMutex
    lastRetry      time.Time
    circuitBreaker *CircuitBreaker
}

// ProducerMetrics holds producer metrics
type ProducerMetrics struct {
    TotalEvents      uint64
    TotalBatches     uint64
    TotalBytes       uint64
    SuccessCount     uint64
    FailureCount     uint64
    RetryCount       uint64
    LastSuccessTime  time.Time
    LastFailureTime  time.Time
    AverageBatchSize float64
    Errors           []ProducerError
    mu               sync.RWMutex
}

// ProducerError represents an error that occurred during production
type ProducerError struct {
    Timestamp   time.Time
    Error       error
    BatchSize   int
    WillRetry   bool
}

// Batch represents a batch of events to be sent
type Batch struct {
    Events   []*normalization.NetworkEvent
    Metadata BatchMetadata
}

// BatchMetadata contains metadata about a batch
type BatchMetadata struct {
    CreatedAt   time.Time
    Attempts    int
    FirstEvent  string
    LastEvent   string
}

// circuitState represents the state of the circuit breaker
type circuitState int

const (
    circuitClosed circuitState = iota
    circuitHalfOpen
    circuitOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
    state             circuitState
    failureCount      int
    successCount      int
    lastFailure       time.Time
    mu                sync.RWMutex
    threshold         int
    timeout           time.Duration
    recoveryThreshold int
}

// NewNATSProducer creates a new NATS producer
func NewNATSProducer(cfg *config.NATSConfig) (*NATSProducer, error) {
    ctx, cancel := context.WithCancel(context.Background())

    p := &NATSProducer{
        cfg:      cfg,
        batchCh:  make(chan *normalization.NetworkEvent, 10000),
        batchBuf: make([]*normalization.NetworkEvent, 0),
        metrics:  &ProducerMetrics{},
        ctx:      ctx,
        cancel:   cancel,
        stopCh:   make(chan struct{}),
    }

    // Set up circuit breaker
    p.circuitBreaker = &CircuitBreaker{
        threshold:         5,
        timeout:           30 * time.Second,
        recoveryThreshold: 3,
    }

    // Connect to NATS
    if err := p.connect(); err != nil {
        cancel()
        return nil, err
    }

    return p, nil
}

// connect establishes connection to NATS
func (np *NATSProducer) connect() error {
    // Configure NATS options
    opts := []nats.Option{
        nats.Name("sge-network-sensor"),
        nats.ReconnectWait(np.cfg.ReconnectWait),
        nats.MaxReconnects(np.cfg.MaxReconnectAttempts),
        nats.ErrorHandler(np.natsErrorHandler),
    }

    // Add TLS if configured
    if np.cfg.CertFile != "" && np.cfg.KeyFile != "" {
        opts = append(opts, nats.ClientCert(np.cfg.CertFile, np.cfg.KeyFile))
    }
    if np.cfg.CACertFile != "" {
        opts = append(opts, nats.RootCAs(np.cfg.CACertFile))
    }

    // Connect
    conn, err := nats.Connect(np.cfg.URLs[0], opts...)
    if err != nil {
        return fmt.Errorf("failed to connect to NATS: %w", err)
    }
    np.conn = conn

    // Get JetStream context
    np.js, err = jetstream.New(np.conn)
    if err != nil {
        conn.Close()
        return fmt.Errorf("failed to create JetStream context: %w", err)
    }

    return nil
}

// natsErrorHandler handles NATS connection errors
func (np *NATSProducer) natsErrorHandler(conn *nats.Conn, sub *nats.Subscription, err error) {
    np.recordFailure(err, 0)
}

// Start starts the producer
func (np *NATSProducer) Start(batchSize int, flushInterval time.Duration) {
    np.wg.Add(1)
    go np.batchProcessor(batchSize, flushInterval)

    np.wg.Add(1)
    go np.flushWorker(flushInterval)
}

// Stop gracefully stops the producer
func (np *NATSProducer) Stop() error {
    np.cancel()
    close(np.stopCh)

    // Flush remaining events
    np.mu.Lock()
    if len(np.batchBuf) > 0 {
        if err := np.flush(); err != nil {
            np.recordFailure(err, len(np.batchBuf))
        }
    }
    np.mu.Unlock()

    np.wg.Wait()

    // Close connection
    if np.conn != nil {
        return np.conn.Close()
    }

    return nil
}

// Publish publishes a single event
func (np *NATSProducer) Publish(event *normalization.NetworkEvent) error {
    select {
    case np.batchCh <- event:
        return nil
    default:
        // Buffer full, record drop
        np.recordFailure(errors.New("batch buffer full"), 1)
        return errors.New("batch buffer full, event dropped")
    }
}

// PublishBatch publishes a batch of events
func (np *NATSProducer) PublishBatch(events []*normalization.NetworkEvent) error {
    for _, event := range events {
        if err := np.Publish(event); err != nil {
            return err
        }
    }
    return nil
}

// batchProcessor processes events from the batch channel
func (np *NATSProducer) batchProcessor(batchSize int, flushInterval time.Duration) {
    defer np.wg.Done()

    for {
        select {
        case <-np.ctx.Done():
            return
        case event := <-np.batchCh:
            np.mu.Lock()
            np.batchBuf = append(np.batchBuf, event)
            if len(np.batchBuf) >= batchSize {
                if err := np.flush(); err != nil {
                    np.recordFailure(err, len(np.batchBuf))
                }
            }
            np.mu.Unlock()
        }
    }
}

// flushWorker periodically flushes the batch buffer
func (np *NATSProducer) flushWorker(interval time.Duration) {
    defer np.wg.Done()

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-np.ctx.Done():
            return
        case <-ticker.C:
            np.mu.Lock()
            if len(np.batchBuf) > 0 {
                if err := np.flush(); err != nil {
                    np.recordFailure(err, len(np.batchBuf))
                }
            }
            np.mu.Unlock()
        }
    }
}

// flush sends the current batch to NATS
func (np *NATSProducer) flush() error {
    if len(np.batchBuf) == 0 {
        return nil
    }

    // Check circuit breaker
    if !np.circuitBreaker.allowRequest() {
        return errors.New("circuit breaker open")
    }

    events := np.batchBuf
    np.batchBuf = make([]*normalization.NetworkEvent, 0)

    // Create batch
    batch := Batch{
        Events: events,
        Metadata: BatchMetadata{
            CreatedAt:  time.Now(),
            Attempts:   1,
            FirstEvent: events[0].ID,
            LastEvent:  events[len(events)-1].ID,
        },
    }

    // Send to NATS
    if err := np.sendBatch(batch); err != nil {
        np.circuitBreaker.recordFailure()
        return err
    }

    np.circuitBreaker.recordSuccess()

    // Update metrics
    atomic.AddUint64(&np.metrics.TotalEvents, uint64(len(events)))
    atomic.AddUint64(&np.metrics.TotalBatches, 1)
    np.metrics.LastSuccessTime = time.Now()

    // Calculate bytes
    for _, event := range events {
        data, _ := event.ConvertToJSON()
        atomic.AddUint64(&np.metrics.TotalBytes, uint64(len(data)))
    }

    return nil
}

// sendBatch sends a batch to NATS JetStream
func (np *NATSProducer) sendBatch(batch Batch) error {
    // Marshal events to JSON lines format
    var sb strings.Builder
    for _, event := range batch.Events {
        data, err := event.ConvertToJSON()
        if err != nil {
            continue
        }
        sb.Write(data)
        sb.WriteString("\n")
    }

    data := []byte(sb.String())

    // Publish to subject
    _, err := np.js.Publish(np.ctx, np.cfg.Subject, data)
    if err != nil {
        // Check if we should retry
        if shouldRetry(err) && batch.Metadata.Attempts < 3 {
            np.retryMu.Lock()
            np.lastRetry = time.Now()
            np.retryMu.Unlock()

            batch.Metadata.Attempts++
            atomic.AddUint64(&np.metrics.RetryCount, 1)

            // Wait before retry with exponential backoff
            backoff := time.Duration(math.Pow(2, float64(batch.Metadata.Attempts))) * time.Second
            time.Sleep(backoff)

            return np.sendBatch(batch)
        }
        return err
    }

    atomic.AddUint64(&np.metrics.SuccessCount, uint64(len(batch.Events)))
    return nil
}

// recordFailure records a failure
func (np *NATSProducer) recordFailure(err error, batchSize int) {
    np.metrics.mu.Lock()
    np.metrics.FailureCount++
    np.metrics.LastFailureTime = time.Now()
    np.metrics.Errors = append(np.metrics.Errors, ProducerError{
        Timestamp: time.Now(),
        Error:     err,
        BatchSize: batchSize,
        WillRetry: shouldRetry(err),
    })
    // Keep only last 100 errors
    if len(np.metrics.Errors) > 100 {
        np.metrics.Errors = np.metrics.Errors[len(np.metrics.Errors)-100:]
    }
    np.metrics.mu.Unlock()
}

// GetMetrics returns current producer metrics
func (np *NATSProducer) GetMetrics() ProducerMetrics {
    np.metrics.mu.RLock()
    defer np.metrics.mu.RUnlock()

    return ProducerMetrics{
        TotalEvents:     np.metrics.TotalEvents,
        TotalBatches:    np.metrics.TotalBatches,
        TotalBytes:      np.metrics.TotalBytes,
        SuccessCount:    np.metrics.SuccessCount,
        FailureCount:    np.metrics.FailureCount,
        RetryCount:      np.metrics.RetryCount,
        LastSuccessTime: np.metrics.LastSuccessTime,
        LastFailureTime: np.metrics.LastFailureTime,
        Errors:          np.metrics.Errors,
    }
}

// GetQueueDepth returns the current queue depth
func (np *NATSProducer) GetQueueDepth() int {
    return len(np.batchCh)
}

// IsConnected returns true if the producer is connected
func (np *NATSProducer) IsConnected() bool {
    return np.conn != nil && np.conn.IsConnected()
}

// CreateStream creates a JetStream stream for the events
func (np *NATSProducer) CreateStream(streamName string) error {
    stream, err := np.js.CreateStream(np.ctx, jetstream.StreamConfig{
        Name:     streamName,
        Subjects: []string{np.cfg.Subject},
        Storage:  jetstream.FileStorage,
        MaxAge:   24 * time.Hour,
    })
    if err != nil {
        return fmt.Errorf("failed to create stream: %w", err)
    }

    np.producer = stream
    return nil
}

// shouldRetry determines if an error should trigger a retry
func shouldRetry(err error) bool {
    errStr := err.Error()
    retryableErrors := []string{
        "connection refused",
        "connection closed",
        "timeout",
        "no responders",
        "temporary failure",
    }

    for _, pattern := range retryableErrors {
        if strings.Contains(errStr, pattern) {
            return true
        }
    }

    return false
}

// CircuitBreaker methods

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker *CircuitBreaker

func (cb *CircuitBreaker) allowRequest() bool {
    cb.mu.RWMutex.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case circuitClosed:
        return true
    case circuitHalfOpen:
        return cb.successCount < cb.recoveryThreshold
    case circuitOpen:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = circuitHalfOpen
            cb.successCount = 0
            cb.failureCount = 0
            return true
        }
        return false
    }
    return true
}

func (cb *CircuitBreaker) recordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failureCount++
    cb.lastFailure = time.Now()

    if cb.failureCount >= cb.threshold {
        cb.state = circuitOpen
    }
}

func (cb *CircuitBreaker) recordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.successCount++
    cb.failureCount = 0

    if cb.state == circuitHalfOpen && cb.successCount >= cb.recoveryThreshold {
        cb.state = circuitClosed
    }
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case circuitClosed:
        return "closed"
    case circuitHalfOpen:
        return "half-open"
    case circuitOpen:
        return "open"
    default:
        return "unknown"
    }
}

// FileProducer is a simple file-based producer for fallback
type FileProducer struct {
    cfg     *config.FileOutputConfig
    file    *FileWriter
    metrics *ProducerMetrics
    mu      sync.Mutex
}

// FileWriter wraps file operations
type FileWriter struct {
    file    *os.File
    path    string
    mu      sync.Mutex
    size    int64
    maxSize int64
}

// NewFileProducer creates a new file-based producer
func NewFileProducer(cfg *config.FileOutputConfig) (*FileProducer, error) {
    fp := &FileProducer{
        cfg:     cfg,
        metrics: &ProducerMetrics{},
    }

    // Create output directory
    if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
        return nil, err
    }

    return fp, nil
}

// Start starts the file producer
func (fp *FileProducer) Start() error {
    return nil
}

// Stop stops the file producer
func (fp *FileProducer) Stop() error {
    return nil
}

// Publish writes an event to file
func (fp *FileProducer) Publish(event *normalization.NetworkEvent) error {
    data, err := event.ConvertToJSON()
    if err != nil {
        return err
    }

    fp.mu.Lock()
    defer fp.mu.Unlock()

    fp.metrics.TotalEvents++
    fp.metrics.TotalBytes += uint64(len(data))

    _, err = fp.file.Write(append(data, '\n'))
    return err
}

// GetMetrics returns producer metrics
func (fp *FileProducer) GetMetrics() ProducerMetrics {
    return *fp.metrics
}
