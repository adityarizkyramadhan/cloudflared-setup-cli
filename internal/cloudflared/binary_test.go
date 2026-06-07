package cloudflared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)

func TestConfigDir(t *testing.T) {
	dir, err := cloudflared.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".cloudflared")
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestDownloadURL(t *testing.T) {
	url, err := cloudflared.DownloadURL()
	if err != nil {
		t.Fatalf("DownloadURL() error: %v", err)
	}
	if url == "" {
		t.Error("DownloadURL() returned empty string")
	}
}

func TestIsInstalled_noPanic(t *testing.T) {
	_ = cloudflared.IsInstalled()
}
