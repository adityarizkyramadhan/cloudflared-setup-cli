package cloudflared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)

func TestWriteReadConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".cloudflared"), 0700); err != nil {
		t.Fatal(err)
	}

	cfg := &cloudflared.Config{
		Tunnel: "my-tunnel",
		Ingress: []cloudflared.IngressRule{
			{Hostname: "app.example.com", Service: "http://localhost:8080"},
		},
	}

	if err := cloudflared.WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	got, err := cloudflared.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if got.Tunnel != "my-tunnel" {
		t.Errorf("Tunnel: got %q, want %q", got.Tunnel, "my-tunnel")
	}
	if len(got.Ingress) < 2 {
		t.Errorf("expected at least 2 ingress rules, got %d", len(got.Ingress))
	}
}

func TestActiveTunnelEmptyWhenNoConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	got, err := cloudflared.ActiveTunnel()
	if err != nil {
		t.Fatalf("ActiveTunnel: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty tunnel, got %q", got)
	}
}

func TestSetAndGetActiveTunnel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".cloudflared"), 0700); err != nil {
		t.Fatal(err)
	}

	// Pre-existing ingress rule should survive SetTunnel.
	if err := cloudflared.AddIngressRule("app.example.com", "http://localhost:8080"); err != nil {
		t.Fatal(err)
	}
	if err := cloudflared.SetTunnel("my-tunnel"); err != nil {
		t.Fatalf("SetTunnel: %v", err)
	}

	got, err := cloudflared.ActiveTunnel()
	if err != nil {
		t.Fatalf("ActiveTunnel: %v", err)
	}
	if got != "my-tunnel" {
		t.Errorf("Tunnel: got %q, want %q", got, "my-tunnel")
	}

	cfg, err := cloudflared.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range cfg.Ingress {
		if r.Hostname == "app.example.com" {
			found = true
		}
	}
	if !found {
		t.Error("SetTunnel dropped the existing ingress rule")
	}
}

func TestAddIngressRule(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".cloudflared"), 0700); err != nil {
		t.Fatal(err)
	}

	if err := cloudflared.WriteConfig(&cloudflared.Config{Tunnel: "t"}); err != nil {
		t.Fatal(err)
	}

	if err := cloudflared.AddIngressRule("app.example.com", "http://localhost:3000"); err != nil {
		t.Fatalf("AddIngressRule: %v", err)
	}

	cfg, err := cloudflared.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Ingress[0].Hostname != "app.example.com" {
		t.Errorf("got %q, want %q", cfg.Ingress[0].Hostname, "app.example.com")
	}
}
