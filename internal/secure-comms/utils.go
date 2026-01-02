package securecomms

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// VerifyClientCertificate, client sertifikasını doğrular.
func VerifyClientCertificate(tlsConfig *tls.Config, clientCert *x509.Certificate) error {
	// Sertifika süresini kontrol et
	now := time.Now()
	if now.Before(clientCert.NotBefore) {
		return fmt.Errorf("certificate not yet valid")
	}
	if now.After(clientCert.NotAfter) {
		return fmt.Errorf("certificate expired")
	}

	// CA ile doğrulama
	opts := x509.VerifyOptions{
		Roots:     tlsConfig.RootCAs,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := clientCert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// NewMTLSHTTPClient, mTLS destekli HTTP client oluşturur.
func NewMTLSHTTPClient(certFile, keyFile, caFile string) (*http.Client, error) {
	// Client certificate yükle
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// CA certificate pool oluştur
	caCertPool, err := LoadCAPool(caFile)
	if err != nil {
		return nil, err
	}

	// TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS13,
	}

	// HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return client, nil
}

// readCertFile, sertifika dosyasını okur.
func readCertFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// LoadCAPool, CA certificate pool'unu yükler.
func LoadCAPool(caFile string) (*x509.CertPool, error) {
	// CA certificate oku
	caCert, err := readCertFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	// Pool oluştur
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	return caCertPool, nil
}

// GetCertCommonName, TLS connection'dan client'ın CN'sini çıkarır.
func GetCertCommonName(tlsConn *tls.Conn) (string, error) {
	if err := tlsConn.Handshake(); err != nil {
		return "", fmt.Errorf("TLS handshake failed: %w", err)
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return "", fmt.Errorf("no peer certificates")
	}

	return state.PeerCertificates[0].Subject.CommonName, nil
}

// ValidateTLSVersion, TLS versiyonunun minimum gereksinimi karşıladığını kontrol eder.
func ValidateTLSVersion(tlsConn *tls.Conn, minVersion uint16) error {
	state := tlsConn.ConnectionState()
	if state.Version < minVersion {
		return fmt.Errorf("TLS version %d is below minimum required version %d", state.Version, minVersion)
	}
	return nil
}

// GetTLSConnectionInfo, TLS bağlantı bilgilerini döndürür.
func GetTLSConnectionInfo(tlsConn *tls.Conn) map[string]interface{} {
	state := tlsConn.ConnectionState()

	info := map[string]interface{}{
		"version":             getTLSVersionString(state.Version),
		"cipher_suite":        tls.CipherSuiteName(state.CipherSuite),
		"handshake_complete":  state.HandshakeComplete,
		"server_name":         state.ServerName,
		"negotiated_protocol": state.NegotiatedProtocol,
		"peer_certificates":   len(state.PeerCertificates),
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		info["peer_common_name"] = cert.Subject.CommonName
		info["peer_organization"] = cert.Subject.Organization
		info["peer_not_before"] = cert.NotBefore.Format(time.RFC3339)
		info["peer_not_after"] = cert.NotAfter.Format(time.RFC3339)
	}

	return info
}

// getTLSVersionString, TLS version numarasını string'e çevirir.
func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// ExtractClientIDFromCert, sertifikadan client ID'sini çıkarır.
func ExtractClientIDFromCert(cert *x509.Certificate) string {
	// CommonName'i client ID olarak kullan
	return cert.Subject.CommonName
}

// IsCertExpiringSoon, sertifikanın yakında süresinin doleceğini kontrol eder.
func IsCertExpiringSoon(cert *x509.Certificate, daysThreshold int) bool {
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)
	return daysUntilExpiry <= daysThreshold
}
