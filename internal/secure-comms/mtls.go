package securecomms

import (
	"crypto/tls"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"
)

// MTLSConfig, mTLS yapılandırmasını içerir.
type MTLSConfig struct {
	CertsDir       string
	ServerCertFile string
	ServerKeyFile  string
	ClientCertFile string
	ClientKeyFile  string
	CACertFile     string
	AutoRotate     bool
	RotationDays   int
	CheckInterval  time.Duration
}

// MTLSManager, mTLS sertifikalarını yönetir ve otomatik rotation yapar.
type MTLSManager struct {
	config      *MTLSConfig
	certManager *CertManager
	tlsConfig   *tls.Config
	mu          sync.RWMutex
	stopChan    chan struct{}
}

// NewMTLSManager, yeni bir MTLSManager oluşturur.
func NewMTLSManager(config *MTLSConfig) (*MTLSManager, error) {
	certManager, err := NewCertManager(config.CertsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert manager: %w", err)
	}

	manager := &MTLSManager{
		config:      config,
		certManager: certManager,
		stopChan:    make(chan struct{}),
	}

	// İlk TLS config'i yükle
	if err := manager.reloadTLSConfig(); err != nil {
		return nil, fmt.Errorf("failed to load initial TLS config: %w", err)
	}

	// Auto-rotation başlat
	if config.AutoRotate {
		go manager.startAutoRotation()
	}

	return manager, nil
}

// GetTLSConfig, güncel TLS yapılandırmasını döndürür.
func (m *MTLSManager) GetTLSConfig() *tls.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tlsConfig.Clone()
}

// reloadTLSConfig, TLS yapılandırmasını yeniden yükler.
func (m *MTLSManager) reloadTLSConfig() error {
	certFile := filepath.Join(m.config.CertsDir, m.config.ServerCertFile)
	keyFile := filepath.Join(m.config.CertsDir, m.config.ServerKeyFile)
	caFile := filepath.Join(m.config.CertsDir, m.config.CACertFile)

	tlsConfig, err := m.certManager.LoadTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.tlsConfig = tlsConfig
	m.mu.Unlock()

	return nil
}

// startAutoRotation, otomatik sertifika rotation'ı başlatır.
func (m *MTLSManager) startAutoRotation() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	log.Printf("[mTLS] Auto-rotation started (check interval: %v)", m.config.CheckInterval)

	for {
		select {
		case <-ticker.C:
			if err := m.checkAndRotate(); err != nil {
				log.Printf("[mTLS] Rotation check failed: %v", err)
			}
		case <-m.stopChan:
			log.Println("[mTLS] Auto-rotation stopped")
			return
		}
	}
}

// checkAndRotate, sertifikaların süresini kontrol eder ve gerekirse rotate eder.
func (m *MTLSManager) checkAndRotate() error {
	certPath := filepath.Join(m.config.CertsDir, m.config.ServerCertFile)

	timeUntilExpiry, err := m.certManager.CheckCertExpiry(certPath)
	if err != nil {
		return fmt.Errorf("failed to check cert expiry: %w", err)
	}

	daysUntilExpiry := int(timeUntilExpiry.Hours() / 24)

	log.Printf("[mTLS] Certificate expires in %d days", daysUntilExpiry)

	// Eğer rotation threshold'una ulaşıldıysa yenile
	if daysUntilExpiry <= m.config.RotationDays {
		log.Printf("[mTLS] Certificate rotation triggered (threshold: %d days)", m.config.RotationDays)

		// Yeni sertifika oluştur
		certConfig := &CertConfig{
			Organization:       "SGE",
			OrganizationalUnit: "Security",
			Country:            "TR",
			Province:           "Istanbul",
			Locality:           "Istanbul",
			CommonName:         "SGE Server",
			ValidityDays:       365,
			KeySize:            2048,
		}

		// Server sertifikasını yenile
		if err := m.certManager.GenerateServerCert(certConfig, []string{"localhost"}, nil); err != nil {
			return fmt.Errorf("failed to rotate server cert: %w", err)
		}

		// TLS config'i yeniden yükle
		if err := m.reloadTLSConfig(); err != nil {
			return fmt.Errorf("failed to reload TLS config: %w", err)
		}

		log.Println("[mTLS] Certificate rotation completed successfully")
	}

	return nil
}

// Stop, otomatik rotation'ı durdurur.
func (m *MTLSManager) Stop() {
	close(m.stopChan)
}

// GenerateAllCertificates, CA, server ve client sertifikalarını oluşturur.
func (m *MTLSManager) GenerateAllCertificates(serverDNS []string, clientIDs []string) error {
	certConfig := &CertConfig{
		Organization:       "SGE",
		OrganizationalUnit: "Security",
		Country:            "TR",
		Province:           "Istanbul",
		Locality:           "Istanbul",
		CommonName:         "SGE CA",
		ValidityDays:       3650, // CA 10 yıl geçerli
		KeySize:            4096,
	}

	// 1. CA oluştur
	log.Println("[mTLS] Generating CA certificate...")
	if err := m.certManager.GenerateCA(certConfig); err != nil {
		return fmt.Errorf("failed to generate CA: %w", err)
	}

	// 2. Server certificate oluştur
	log.Println("[mTLS] Generating server certificate...")
	serverConfig := &CertConfig{
		Organization:       "SGE",
		OrganizationalUnit: "Services",
		Country:            "TR",
		Province:           "Istanbul",
		Locality:           "Istanbul",
		CommonName:         "SGE Server",
		ValidityDays:       365,
		KeySize:            2048,
	}

	if err := m.certManager.GenerateServerCert(serverConfig, serverDNS, nil); err != nil {
		return fmt.Errorf("failed to generate server cert: %w", err)
	}

	// 3. Client certificates oluştur
	for _, clientID := range clientIDs {
		log.Printf("[mTLS] Generating client certificate for: %s", clientID)
		clientConfig := &CertConfig{
			Organization:       "SGE",
			OrganizationalUnit: "Agents",
			Country:            "TR",
			Province:           "Istanbul",
			Locality:           "Istanbul",
			CommonName:         clientID,
			ValidityDays:       365,
			KeySize:            2048,
		}

		if err := m.certManager.GenerateClientCert(clientConfig, clientID); err != nil {
			return fmt.Errorf("failed to generate client cert for %s: %w", clientID, err)
		}
	}

	// TLS config'i yükle
	if err := m.reloadTLSConfig(); err != nil {
		return fmt.Errorf("failed to load TLS config: %w", err)
	}

	log.Println("[mTLS] All certificates generated successfully")
	return nil
}

// GetCertificateInfo, sertifika bilgilerini döndürür.
func (m *MTLSManager) GetCertificateInfo(certPath string) (map[string]interface{}, error) {
	fullPath := filepath.Join(m.config.CertsDir, certPath)

	timeUntilExpiry, err := m.certManager.CheckCertExpiry(fullPath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"path":              certPath,
		"days_until_expiry": int(timeUntilExpiry.Hours() / 24),
		"expires_at":        time.Now().Add(timeUntilExpiry).Format(time.RFC3339),
		"valid":             timeUntilExpiry > 0,
	}, nil
}
