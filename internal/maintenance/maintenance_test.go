package maintenance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/maintenance"
)

func TestOrphanedCredentials(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "used-tunnel.json"), []byte("{}"), 0600)
	os.WriteFile(filepath.Join(dir, "orphan.json"), []byte("{}"), 0600)
	os.WriteFile(filepath.Join(dir, "cert.pem"), []byte("pem"), 0600)

	inUse := map[string]bool{"used-tunnel": true}
	orphans, err := maintenance.OrphanedCredentials(dir, inUse)
	if err != nil {
		t.Fatalf("OrphanedCredentials: %v", err)
	}
	if len(orphans) != 1 {
		t.Errorf("expected 1 orphan, got %d: %v", len(orphans), orphans)
	}
}

func TestDeleteFiles(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "to-delete.txt")
	os.WriteFile(f, []byte("x"), 0600)

	errs := maintenance.DeleteFiles([]string{f})
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestCurrentVersion_noPanic(t *testing.T) {
	_ = maintenance.CurrentVersion()
}
