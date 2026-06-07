package cloudflared_test

import (
	"bytes"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)

func TestListTunnels_emptyOutput(t *testing.T) {
	if cloudflared.IsInstalled() {
		t.Skip("cloudflared is installed; skipping stub test")
	}
	_, err := cloudflared.ListTunnels()
	if err == nil {
		t.Error("expected error when cloudflared not installed")
	}
}

func TestStopTunnel_nilCmd(t *testing.T) {
	if err := cloudflared.StopTunnel(nil); err != nil {
		t.Errorf("StopTunnel(nil) should not error: %v", err)
	}
}

func TestCleanupTunnel_emptyList(t *testing.T) {
	errs := cloudflared.CleanupTunnel([]string{})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

var _ = bytes.NewBuffer
