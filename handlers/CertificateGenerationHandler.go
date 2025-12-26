package Handlers

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// This is where your MITM root certificate (certificate authority) is located
var mitmRootCert *x509.Certificate
var mitmRootKey *ecdsa.PrivateKey

// Function to create a dynamically generated certificate for the target host
func GenerateCertificateForHost(hostname string) (*tls.Certificate, error) {
	// Validate the host (should not be empty)
	if len(hostname) == 0 {
		return nil, errors.New("hostname cannot be empty")
	}

	// Set up the certificate template
	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: hostname},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname},
	}

	// Create the certificate using the MITM root certificate and private key
	cert, err := x509.CreateCertificate(rand.Reader, &template, mitmRootCert, mitmRootKey.Public(), mitmRootKey)
	if err != nil {
		return nil, fmt.Errorf("error creating certificate: %v", err)
	}

	// Create a tls.Certificate
	certificate := tls.Certificate{
		Certificate: [][]byte{cert},
		PrivateKey:  mitmRootKey,
	}

	return &certificate, nil
}
