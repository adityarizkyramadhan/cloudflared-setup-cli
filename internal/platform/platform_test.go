package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestServiceManager(t *testing.T) {
	got := ServiceManager()
	var want string
	switch runtime.GOOS {
	case "windows":
		want = "windows"
	case "linux":
		want = "systemd"
	case "darwin":
		want = "launchd"
	default:
		want = "none"
	}
	if got != want {
		t.Errorf("ServiceManager() = %q, want %q for GOOS %q", got, want, runtime.GOOS)
	}
}

func TestInstallDirIsWritable(t *testing.T) {
	dir, err := InstallDir()
	if err != nil {
		t.Fatalf("InstallDir() returned error: %v", err)
	}
	if dir == "" {
		t.Fatal("InstallDir() returned empty path")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("install dir %q not stat-able: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("install dir %q is not a directory", dir)
	}
	// Confirm we can actually write there.
	probe := filepath.Join(dir, ".cfd-test-marker")
	if err := os.WriteFile(probe, []byte("x"), 0o600); err != nil {
		t.Fatalf("install dir %q not writable: %v", dir, err)
	}
	os.Remove(probe)
}

func TestWritableFalseForMissingDir(t *testing.T) {
	if writable(filepath.Join(t.TempDir(), "does-not-exist")) {
		t.Error("writable() returned true for a non-existent directory")
	}
}
