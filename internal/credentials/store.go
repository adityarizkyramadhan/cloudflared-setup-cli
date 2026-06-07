package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	dir string
}

func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

func NewAt(dir string) *Store { return &Store{dir: dir} }

func (s *Store) Read(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.dir, name))
}

func (s *Store) Write(name string, data []byte) error {
	return os.WriteFile(filepath.Join(s.dir, name), data, 0600)
}

func (s *Store) Exists(name string) bool {
	_, err := os.Stat(filepath.Join(s.dir, name))
	return err == nil
}

func (s *Store) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name))
}

func (s *Store) ListCredentialFiles() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

func (s *Store) BackupTo(destDir string) error {
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destDir, e.Name()), data, 0600); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) RestoreFrom(srcDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(s.dir, e.Name()), data, 0600); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Nuke() error {
	return os.RemoveAll(s.dir)
}
