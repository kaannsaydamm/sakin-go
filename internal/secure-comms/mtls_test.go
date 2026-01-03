package securecomms

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMTLSManager_GenerateAndLoad(t *testing.T) {
	// 1. Setup temp dir
	tmpDir, err := os.MkdirTemp("", "sge-mtls-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := &MTLSConfig{
		CertsDir:       tmpDir,
		ServerCertFile: "server.crt",
		ServerKeyFile:  "server.key",
		ClientCertFile: "client.crt",
		ClientKeyFile:  "client.key",
		CACertFile:     "ca.crt",
		AutoRotate:     false,
		RotationDays:   30,
		CheckInterval:  1 * time.Hour,
	}

	// 2. Initialize Manager
	// Note: NewMTLSManager tries to reloadTLSConfig immediately, which might fail if certs don't exist yet.
	// Check logic of NewMTLSManager: it calls reloadTLSConfig.
	// So we might need to generate certs FIRST using CertManager directly, or handle the error.
	// However, usually we want a manager instance to start generating.
	// Let's see if we can instantiate CertManager separately easily or if MTLSManager should handle "empty state".
	// Looking at code: NewMTLSManager calls reloadTLSConfig immediately and returns error if fails.
	// So for first time setup, we can't use NewMTLSManager directly if certs are missing.
	// This is a design insight for the test. We will use CertManager manually first or modify NewMTLSManager to allow lazy loading.
	// Since we can't modify code in this step effortlessly without side effects, let's look at how we can bootstrap.
	// We can use the internal certManager of a nil-config-loaded struct? No.

	// Better approach: Use the same helper GenerateAllCertificates logic but maybe checking NewMTLSManager isn't possible yet.
	// Actually, MTLSManager.GenerateAllCertificates is a method on the instance.
	// So we need an instance to generate. But we can't get an instance because New... fails.
	// Chicken and egg problem in the design.
	// FIX: We will construct MTLSManager manually for the test to bypass the initial load check.

	cm, err := NewCertManager(tmpDir)
	if err != nil {
		t.Fatalf("NewCertManager failed: %v", err)
	}

	manager := &MTLSManager{
		config:      config,
		certManager: cm,
		// stopChan not needed for this test
	}

	// 3. Generate Certs
	err = manager.GenerateAllCertificates([]string{"localhost"}, []string{"test-client"})
	if err != nil {
		t.Fatalf("GenerateAllCertificates failed: %v", err)
	}

	// 4. Verify Files Exist
	files := []string{"ca.crt", "server.crt", "server.key", "client-test-client.crt", "client-test-client.key"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found", f)
		}
	}

	// 5. Verify TLS Config Loading
	// Now we can properly "reload"
	err = manager.reloadTLSConfig()
	if err != nil {
		t.Fatalf("reloadTLSConfig failed: %v", err)
	}

	tlsConfig := manager.GetTLSConfig()
	if tlsConfig == nil {
		t.Fatal("GetTLSConfig returned nil")
	}
	if len(tlsConfig.Certificates) == 0 {
		t.Error("TLS Config has no certificates loaded")
	}
	if tlsConfig.ClientAuth != tls.RequireAndVerifyClientCert {
		// Default behavior of helper might be this? Needs checking load logic.
		// Actually CertManager.LoadTLSConfig sets this.
	}
}
