package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig, Redis bağlantı ayarlarını içerir.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// RedisClient, Redis bağlantı havuzunu yönetir.
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient, yeni bir Redis client oluşturur.
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Bağlantı testi
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisClient{
		client: client,
		config: config,
	}, nil
}

// GetClient, *redis.Client instance'ını döndürür.
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Ping, bağlantının sağlıklı olup olmadığını kontrol eder.
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close, bağlantıyı kapatır.
func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Set, key-value çiftini belirtilen TTL ile saklar.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get, key'e karşılık gelen değeri getirir.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete, key'i siler.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists, key'in var olup olmadığını kontrol eder.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Increment, key'in değerini 1 artırır.
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrementBy, key'in değerini belirtilen miktarda artırır.
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// SetExpire, var olan key'e TTL ekler.
func (r *RedisClient) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// GetWithTTL, key'in değerini ve kalan TTL'ini getirir.
func (r *RedisClient) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return "", 0, err
	}

	value, err := getCmd.Result()
	if err != nil {
		return "", 0, err
	}

	ttl, err := ttlCmd.Result()
	if err != nil {
		return "", 0, err
	}

	return value, ttl, nil
}

// --- Correlation State Management ---

// IncrementCorrelationCounter, korelasyon sayacını artırır.
// Sliding window için kullanılır.
func (r *RedisClient) IncrementCorrelationCounter(ctx context.Context, ruleID string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("correlation:counter:%s", ruleID)
	pipe := r.client.Pipeline()

	// Counter'ı artır
	incrCmd := pipe.Incr(ctx, key)

	// TTL set et (sliding window)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incrCmd.Val(), nil
}

// GetCorrelationCounter, korelasyon sayacını okur.
func (r *RedisClient) GetCorrelationCounter(ctx context.Context, ruleID string) (int64, error) {
	key := fmt.Sprintf("correlation:counter:%s", ruleID)
	result, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

// ResetCorrelationCounter, korelasyon sayacını sıfırlar.
func (r *RedisClient) ResetCorrelationCounter(ctx context.Context, ruleID string) error {
	key := fmt.Sprintf("correlation:counter:%s", ruleID)
	return r.client.Del(ctx, key).Err()
}

// --- Cache Management (Threat Intel, GeoIP) ---

// SetThreatIntel, threat intel sonucunu cache'ler.
func (r *RedisClient) SetThreatIntel(ctx context.Context, ip string, data string, ttl time.Duration) error {
	key := fmt.Sprintf("threat:intel:%s", ip)
	return r.Set(ctx, key, data, ttl)
}

// GetThreatIntel, cache'lenmiş threat intel verisini getirir.
func (r *RedisClient) GetThreatIntel(ctx context.Context, ip string) (string, error) {
	key := fmt.Sprintf("threat:intel:%s", ip)
	result, err := r.Get(ctx, key)
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return result, err
}

// CacheGeoIP, GeoIP sonucunu cache'ler.
func (r *RedisClient) CacheGeoIP(ctx context.Context, ip string, data string, ttl time.Duration) error {
	key := fmt.Sprintf("geoip:%s", ip)
	return r.Set(ctx, key, data, ttl)
}

// GetCachedGeoIP, cache'lenmiş GeoIP verisini getirir.
func (r *RedisClient) GetCachedGeoIP(ctx context.Context, ip string) (string, error) {
	key := fmt.Sprintf("geoip:%s", ip)
	result, err := r.Get(ctx, key)
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return result, err
}

// --- Session Management ---

// SetSession, kullanıcı oturumunu saklar.
func (r *RedisClient) SetSession(ctx context.Context, sessionID string, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(ctx, key, userID, ttl)
}

// GetSession, oturum bilgisini getirir.
func (r *RedisClient) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	result, err := r.Get(ctx, key)
	if err == redis.Nil {
		return "", fmt.Errorf("session not found")
	}
	return result, err
}

// DeleteSession, oturumu siler (logout).
func (r *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// --- Rate Limiting ---

// CheckRateLimit, rate limit kontrolü yapar.
// Dönen değer: (mevcut request sayısı, izin verilip verilmediği, error)
func (r *RedisClient) CheckRateLimit(ctx context.Context, identifier string, limit int64, window time.Duration) (int64, bool, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)

	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, false, err
	}

	current := incrCmd.Val()
	allowed := current <= limit

	return current, allowed, nil
}

// --- Health Check ---

// Health, Redis sağlık durumunu döndürür.
func (r *RedisClient) Health(ctx context.Context) (map[string]string, error) {
	_, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	stats := r.client.PoolStats()

	return map[string]string{
		"status":      "healthy",
		"hits":        fmt.Sprintf("%d", stats.Hits),
		"misses":      fmt.Sprintf("%d", stats.Misses),
		"total_conns": fmt.Sprintf("%d", stats.TotalConns),
		"idle_conns":  fmt.Sprintf("%d", stats.IdleConns),
		"stale_conns": fmt.Sprintf("%d", stats.StaleConns),
	}, nil
}

// FlushDB, tüm database'i temizler (DIKKAT: Sadece test için kullan!).
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
