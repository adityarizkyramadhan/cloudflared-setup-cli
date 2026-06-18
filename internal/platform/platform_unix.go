//go:build !windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ServiceManager returns the native service manager identifier for the host.
func ServiceManager() string {
	switch runtime.GOOS {
	case "linux":
		return "systemd"
	case "darwin":
		return "launchd"
	default:
		return "none"
	}
}

// IsAdmin reports whether the process runs as root.
func IsAdmin() bool { return os.Geteuid() == 0 }

// installCandidates lists Unix install directories in preference order.
func installCandidates() []string {
	cands := []string{"/usr/local/bin"}
	if home, err := os.UserHomeDir(); err == nil {
		cands = append(cands,
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".cloudflared", "bin"),
		)
	}
	return cands
}

// RelaunchElevated is unsupported on Unix; callers should re-run with sudo.
func RelaunchElevated() error {
	return fmt.Errorf("auto-elevate tidak didukung di OS ini — jalankan ulang dengan sudo")
}
