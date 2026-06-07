package credentials_test

import (
	"path/filepath"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/credentials"
)

func TestWriteReadExists(t *testing.T) {
	dir := t.TempDir()
	s := credentials.NewAt(dir)

	if err := s.Write("test.json", []byte(`{"id":"abc"}`)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !s.Exists("test.json") {
		t.Error("Exists: expected true")
	}
	data, err := s.Read("test.json")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(data) != `{"id":"abc"}` {
		t.Errorf("Read: got %q", string(data))
	}
}

func TestListCredentialFiles(t *testing.T) {
	dir := t.TempDir()
	s := credentials.NewAt(dir)
	s.Write("tunnel1.json", []byte("{}"))
	s.Write("tunnel2.json", []byte("{}"))
	s.Write("cert.pem", []byte("pem"))

	files, err := s.ListCredentialFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 json files, got %d", len(files))
	}
}

func TestBackupRestore(t *testing.T) {
	src := t.TempDir()
	s := credentials.NewAt(src)
	s.Write("cert.pem", []byte("pem-content"))

	backup := t.TempDir()
	if err := s.BackupTo(backup); err != nil {
		t.Fatalf("BackupTo: %v", err)
	}

	dst := t.TempDir()
	s2 := credentials.NewAt(dst)
	if err := s2.RestoreFrom(backup); err != nil {
		t.Fatalf("RestoreFrom: %v", err)
	}
	data, _ := s2.Read("cert.pem")
	if string(data) != "pem-content" {
		t.Errorf("restore: got %q", string(data))
	}
	_ = filepath.Join
}
