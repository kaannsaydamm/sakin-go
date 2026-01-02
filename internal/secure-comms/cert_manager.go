package securecomms

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CertConfig, sertifika oluşturma yapılandırmasını içerir.
type CertConfig struct {
	Organization       string
	OrganizationalUnit string
	Country            string
	Province           string
	Locality           string
	CommonName         string
	ValidityDays       int
	KeySize            int
}

// CertManager, sertifika yönetimi için kullanılır.
type CertManager struct {
	certsDir string
}

// NewCertManager, yeni bir CertManager oluşturur.
func NewCertManager(certsDir string) (*CertManager, error) {
	// Sertifika dizinini oluştur
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create certs directory: %w", err)
	}

	return &CertManager{
		certsDir: certsDir,
	}, nil
}

// GenerateCA, yeni bir Certificate Authority (CA) oluşturur.
func (cm *CertManager) GenerateCA(config *CertConfig) error {
	// Private key oluştur
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// CA sertifika template'i
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{config.Organization},
			OrganizationalUnit: []string{config.OrganizationalUnit},
			Country:            []string{config.Country},
			Province:           []string{config.Province},
			Locality:           []string{config.Locality},
			CommonName:         config.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, config.ValidityDays),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	// Self-signed CA sertifikası oluştur
	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// CA certificate'ı dosyaya yaz
	certPath := filepath.Join(cm.certsDir, "ca.crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Private key'i dosyaya yaz
	keyPath := filepath.Join(cm.certsDir, "ca.key")
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	return nil
}

// GenerateServerCert, CA ile imzalanmış server sertifikası oluşturur.
func (cm *CertManager) GenerateServerCert(config *CertConfig, dnsNames []string, ipAddresses []string) error {
	// CA sertifikası ve private key'i oku
	caCert, caKey, err := cm.loadCA()
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	// Server private key oluştur
	serverKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate server key: %w", err)
	}

	// Server sertifika template'i
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	serverCert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{config.Organization},
			OrganizationalUnit: []string{config.OrganizationalUnit},
			Country:            []string{config.Country},
			Province:           []string{config.Province},
			Locality:           []string{config.Locality},
			CommonName:         config.CommonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, config.ValidityDays),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// DNS names ve IP addresses ekle
	for _, dns := range dnsNames {
		serverCert.DNSNames = append(serverCert.DNSNames, dns)
	}

	// CA ile imzala
	certBytes, err := x509.CreateCertificate(rand.Reader, serverCert, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Server certificate'ı dosyaya yaz
	certPath := filepath.Join(cm.certsDir, "server.crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create server cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to encode server certificate: %w", err)
	}

	// Server private key'i dosyaya yaz
	keyPath := filepath.Join(cm.certsDir, "server.key")
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create server key file: %w", err)
	}
	defer keyFile.Close()

	serverKeyBytes := x509.MarshalPKCS1PrivateKey(serverKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverKeyBytes}); err != nil {
		return fmt.Errorf("failed to encode server private key: %w", err)
	}

	return nil
}

// GenerateClientCert, CA ile imzalanmış client sertifikası oluşturur.
func (cm *CertManager) GenerateClientCert(config *CertConfig, clientID string) error {
	// CA sertifikası ve private key'i oku
	caCert, caKey, err := cm.loadCA()
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	// Client private key oluştur
	clientKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate client key: %w", err)
	}

	// Client sertifika template'i
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	clientCert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{config.Organization},
			OrganizationalUnit: []string{config.OrganizationalUnit},
			Country:            []string{config.Country},
			Province:           []string{config.Province},
			Locality:           []string{config.Locality},
			CommonName:         clientID,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, config.ValidityDays),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// CA ile imzala
	certBytes, err := x509.CreateCertificate(rand.Reader, clientCert, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create client certificate: %w", err)
	}

	// Client certificate'ı dosyaya yaz
	certPath := filepath.Join(cm.certsDir, fmt.Sprintf("client-%s.crt", clientID))
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create client cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to encode client certificate: %w", err)
	}

	// Client private key'i dosyaya yaz
	keyPath := filepath.Join(cm.certsDir, fmt.Sprintf("client-%s.key", clientID))
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create client key file: %w", err)
	}
	defer keyFile.Close()

	clientKeyBytes := x509.MarshalPKCS1PrivateKey(clientKey)
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: clientKeyBytes}); err != nil {
		return fmt.Errorf("failed to encode client private key: %w", err)
	}

	return nil
}

// loadCA, CA sertifikası ve private key'i yükler.
func (cm *CertManager) loadCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	// CA certificate oku
	certPath := filepath.Join(cm.certsDir, "ca.crt")
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA cert PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA cert: %w", err)
	}

	// CA private key oku
	keyPath := filepath.Join(cm.certsDir, "ca.key")
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA key: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key PEM")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA key: %w", err)
	}

	return caCert, caKey, nil
}

// LoadTLSConfig, mTLS için TLS yapılandırmasını yükler.
func (cm *CertManager) LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// Server/Client certificate ve key yükle
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate pair: %w", err)
	}

	// CA certificate pool oluştur
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	// TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}

// CheckCertExpiry, sertifikanın süresini kontrol eder.
func (cm *CertManager) CheckCertExpiry(certPath string) (time.Duration, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return 0, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return time.Until(cert.NotAfter), nil
}
