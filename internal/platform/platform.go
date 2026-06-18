// Package platform centralizes OS-specific decisions (install location,
// service manager, privilege elevation) so the rest of the codebase stays
// platform-agnostic. OS-specific behavior lives in the build-tagged files
// platform_windows.go and platform_unix.go.
package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallDir returns the first writable directory among the OS-specific
// candidates, creating it if necessary. It auto-detects the best location:
// preferring an existing cloudflared install, then well-known system/user
// directories, falling back to a user-owned directory that needs no admin.
func InstallDir() (string, error) {
	var lastErr error
	for _, dir := range installCandidates() {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			lastErr = err
			continue
		}
		if writable(dir) {
			return dir, nil
		}
		lastErr = fmt.Errorf("not writable: %s", dir)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no install candidates available")
	}
	return "", fmt.Errorf("no writable install directory found: %w", lastErr)
}

// writable reports whether a temp file can be created in dir.
func writable(dir string) bool {
	probe := filepath.Join(dir, ".cfd-write-test")
	f, err := os.OpenFile(probe, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(probe)
	return true
}
