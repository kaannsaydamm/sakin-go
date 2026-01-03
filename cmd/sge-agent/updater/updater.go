package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

// Updater manages the self-update process.
type Updater struct {
	CurrentVersion string
	UpdateURL      string // URL to check for updates (returns JSON or plain version)
	BinaryURL      string // URL to download the new binary
}

// CheckUpdate checks if a newer version is available.
// In a real scenario, this would hit an endpoint.
// For now, it mocks a check.
func (u *Updater) CheckUpdate() (string, bool, error) {
	// Mock implementation
	// Real world: GET u.UpdateURL -> parse JSON -> compare semver
	return "", false, nil
}

// PerformUpdate downloads the new binary and replaces the current executable.
func (u *Updater) PerformUpdate() error {
	log.Printf("[Updater] Starting self-update process...")

	// 1. Download new binary
	resp, err := http.Get(u.BinaryURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// 2. Prepare temporary file
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	tmpPath := exePath + ".new"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp binary: %w", err)
	}
	defer out.Close()

	// 3. Write data
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	// 4. Make executable (Linux/Mac)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			return fmt.Errorf("failed to chmod: %w", err)
		}
	}

	out.Close() // Close before move

	// 5. Replace binary (Atomic move on Linux)
	// On Windows, you can't rename over a running executable easily.
	// Windows strategy often involves a wrapper script or renaming the *current* exe to .old first.
	if runtime.GOOS == "windows" {
		oldPath := exePath + ".old"
		os.Remove(oldPath) // Setup
		if err := os.Rename(exePath, oldPath); err != nil {
			return fmt.Errorf("windows rename failed: %w", err)
		}
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	log.Printf("[Updater] Update successful! Restarting...")

	// 6. Restart
	// Simple strategy: Exit and let Systemd/Supervisor restart us.
	os.Exit(0)
	return nil
}

// SelfRestart attempts to restart the process directly (alternative to os.Exit)
func SelfRestart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := os.Args
	env := os.Environ()

	return syscallExec(exe, args, env)
}

func syscallExec(_ string, _ []string, _ []string) error {
	// Wrapper to handle OS differences if needed, but syscall.Exec is Unix only normally.
	// Go's syscall.Exec replaces the process.
	// On Windows this is not supported directly in the same way.
	// We will rely on os.Exit(0) for service managers.
	return fmt.Errorf("syscall.Exec not implemented fully for cross-platform here")
}
