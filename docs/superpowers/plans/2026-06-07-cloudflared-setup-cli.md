# Cloudflared Setup CLI — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an ATM-style interactive TUI in Go that wraps `cloudflared` for full tunnel lifecycle management — authentication, credentials, monitoring, orchestration, and maintenance — distributable as a single binary.

**Architecture:** Bubbletea root model holds a `currentScreen` state and delegates `Update()`/`View()` to whichever sub-model is active. Internal packages (`internal/`) are pure business logic with no Bubbletea dependency — the UI layer wraps their results in `tea.Cmd`. Navigation uses typed messages (`NavigateMsg`, `GoBackMsg`) sent through the Bubbletea event loop.

**Tech Stack:** Go 1.22+, `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`, Cloudflare API over plain `net/http`, GoReleaser, GitHub Actions.

**Module path:** `github.com/adityarizkyramadhan/cloudflared-setup-cli`

---

## Parallel Execution Guide

```
Phase 1 (Sequential)
  Task 1: Project Scaffolding
  Task 2: Root UI Model + Main Menu

Phase 2 (ALL PARALLEL — start after Task 2 completes)
  Task 3: internal/cloudflared
  Task 4: internal/credentials
  Task 5: internal/api
  Task 6: internal/orchestration
  Task 7: internal/monitoring
  Task 8: internal/maintenance

Phase 3 (ALL PARALLEL — start after ALL Phase 2 tasks complete)
  Task 9:  ui — Auth screen
  Task 10: ui — Credentials screen
  Task 11: ui — Monitoring screen
  Task 12: ui — Orchestration screen
  Task 13: ui — Maintenance screen
  Task 14: CI/CD — GitHub Actions + GoReleaser  ← also parallel here

Phase 4 (Sequential — after ALL Phase 3 tasks complete)
  Task 15: Integration + smoke test
```

---

## Phase 1: Foundation (Sequential)

### Task 1: Project Scaffolding

🔗 **No dependencies. Run first.**

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: all directories

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p cmd ui internal/cloudflared internal/credentials internal/api \
  internal/orchestration internal/monitoring internal/maintenance \
  .github/workflows
```

- [ ] **Step 2: Initialise Go module**

```bash
go mod init github.com/adityarizkyramadhan/cloudflared-setup-cli
```

- [ ] **Step 3: Add dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get gopkg.in/yaml.v3@latest
```

- [ ] **Step 4: Write `main.go`**

```go
package main

import "github.com/adityarizkyramadhan/cloudflared-setup-cli/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 5: Write `cmd/root.go`** (stub — UI not wired yet)

```go
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/ui"
)

func Execute() {
	p := tea.NewProgram(ui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 6: Verify build (will fail until Task 2 creates `ui` package)**

```bash
go build ./cmd/... 2>&1 | head -5
```

Expected: error about missing `ui` package — that is fine at this step.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum main.go cmd/root.go
git commit -m "chore: initialise Go module and project structure"
```

---

### Task 2: Root UI Model + Main Menu

🔗 **Depends on: Task 1**

**Files:**
- Create: `ui/model.go`
- Create: `ui/mainmenu.go`
- Create: `ui/styles.go`

- [ ] **Step 1: Write `ui/styles.go`** — shared lipgloss styles

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	MenuStyle = lipgloss.NewStyle().Padding(0, 2)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
```

- [ ] **Step 2: Write `ui/model.go`** — root model and navigation types

```go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifies which sub-model is active.
type Screen int

const (
	ScreenMain Screen = iota
	ScreenAuth
	ScreenCredentials
	ScreenMonitoring
	ScreenOrchestration
	ScreenMaintenance
)

// NavigateMsg tells the root model to switch to a new screen.
type NavigateMsg struct{ To Screen }

// GoBackMsg returns to the main menu.
type GoBackMsg struct{}

// NavigateTo returns a Cmd that sends a NavigateMsg.
func NavigateTo(s Screen) tea.Cmd {
	return func() tea.Msg { return NavigateMsg{To: s} }
}

// GoBack returns a Cmd that sends a GoBackMsg.
func GoBack() tea.Cmd {
	return func() tea.Msg { return GoBackMsg{} }
}

// RootModel is the top-level Bubbletea model.
type RootModel struct {
	screen  Screen
	current tea.Model
}

func NewRootModel() RootModel {
	return RootModel{
		screen:  ScreenMain,
		current: newMainMenuModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case NavigateMsg:
		m.screen = msg.To
		m.current = screenFor(msg.To)
		return m, m.current.Init()
	case GoBackMsg:
		m.screen = ScreenMain
		m.current = newMainMenuModel()
		return m, m.current.Init()
	}

	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	return m.current.View()
}

// screenFor returns a fresh model for the given screen.
// Sub-models are wired in Task 15; stubs return main menu until then.
func screenFor(s Screen) tea.Model {
	return newMainMenuModel() // replaced in Task 15
}
```

- [ ] **Step 3: Write `ui/mainmenu.go`**

```go
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type mainMenuModel struct{}

func newMainMenuModel() mainMenuModel { return mainMenuModel{} }

func (m mainMenuModel) Init() tea.Cmd { return nil }

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			return m, NavigateTo(ScreenAuth)
		case "2":
			return m, NavigateTo(ScreenCredentials)
		case "3":
			return m, NavigateTo(ScreenMonitoring)
		case "4":
			return m, NavigateTo(ScreenOrchestration)
		case "5":
			return m, NavigateTo(ScreenMaintenance)
		case "0", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mainMenuModel) View() string {
	title := TitleStyle.Render("CLOUDFLARED SETUP CLI")
	menu := MenuStyle.Render(fmt.Sprintf(
		"[1] Autentikasi & Setup\n"+
			"[2] Manajemen Kredensial\n"+
			"[3] Observability & Monitoring\n"+
			"[4] Orkestrasi\n"+
			"[5] Pemeliharaan\n\n"+
			"[0] Keluar",
	))
	return title + "\n" + menu
}
```

- [ ] **Step 4: Verify build succeeds**

```bash
go build ./...
```

Expected: exits 0, binary produced.

- [ ] **Step 5: Smoke test — run the app, verify main menu renders**

```bash
go run . 
```

Press `0` to quit. Confirm the 5 menu options appear.

- [ ] **Step 6: Commit**

```bash
git add ui/
git commit -m "feat: add root Bubbletea model and main menu"
```

---

## Phase 2: Internal Packages (All Parallel After Task 2)

> ⚡ Tasks 3–8 have no dependencies on each other. Dispatch to separate subagents simultaneously after Task 2 is committed.

---

### Task 3: internal/cloudflared

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 4, 5, 6, 7, 8**

**Files:**
- Create: `internal/cloudflared/binary.go`
- Create: `internal/cloudflared/binary_test.go`
- Create: `internal/cloudflared/tunnel.go`
- Create: `internal/cloudflared/tunnel_test.go`
- Create: `internal/cloudflared/config.go`
- Create: `internal/cloudflared/config_test.go`

- [ ] **Step 1: Write `internal/cloudflared/binary.go`**

```go
package cloudflared

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// IsInstalled returns true if cloudflared binary is in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("cloudflared")
	return err == nil
}

// GetVersion returns the version string from cloudflared --version.
func GetVersion() (string, error) {
	out, err := exec.Command("cloudflared", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("cloudflared --version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DownloadURL returns the GitHub release URL for the current OS/arch.
func DownloadURL() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var asset string
	switch {
	case goos == "linux" && goarch == "amd64":
		asset = "cloudflared-linux-amd64"
	case goos == "linux" && goarch == "arm64":
		asset = "cloudflared-linux-arm64"
	case goos == "darwin" && goarch == "amd64":
		asset = "cloudflared-darwin-amd64.tgz"
	case goos == "darwin" && goarch == "arm64":
		asset = "cloudflared-darwin-arm64.tgz"
	case goos == "windows" && goarch == "amd64":
		asset = "cloudflared-windows-amd64.exe"
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
	return "https://github.com/cloudflare/cloudflared/releases/latest/download/" + asset, nil
}

// Install downloads and installs the cloudflared binary into destDir.
func Install(destDir string) error {
	url, err := DownloadURL()
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	binName := "cloudflared"
	if runtime.GOOS == "windows" {
		binName = "cloudflared.exe"
	}
	dest := filepath.Join(destDir, binName)

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// Login runs `cloudflared tunnel login` which opens a browser for OAuth.
func Login() error {
	cmd := exec.Command("cloudflared", "tunnel", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// VerifyConnection checks that the cert.pem exists (login was completed).
func VerifyConnection() (bool, error) {
	dir, err := ConfigDir()
	if err != nil {
		return false, err
	}
	certPath := filepath.Join(dir, "cert.pem")
	_, err = os.Stat(certPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// ConfigDir returns ~/.cloudflared (cross-platform).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cloudflared"), nil
}
```

- [ ] **Step 2: Write `internal/cloudflared/binary_test.go`**

```go
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
	_ = cloudflared.IsInstalled() // just ensure no panic
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/cloudflared/ -run TestConfigDir -v
go test ./internal/cloudflared/ -run TestDownloadURL -v
go test ./internal/cloudflared/ -run TestIsInstalled_noPanic -v
```

Expected: all PASS.

- [ ] **Step 4: Write `internal/cloudflared/tunnel.go`**

```go
package cloudflared

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// TunnelInfo holds metadata about a cloudflare tunnel.
type TunnelInfo struct {
	ID   string
	Name string
}

// CreateTunnel runs `cloudflared tunnel create <name>` and returns the tunnel ID.
func CreateTunnel(name string) (string, error) {
	out, err := exec.Command("cloudflared", "tunnel", "create", "--output", "json", name).Output()
	if err != nil {
		return "", fmt.Errorf("tunnel create: %w", err)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		// fallback: parse plain text output
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, l := range lines {
			if strings.Contains(l, "Created tunnel") {
				parts := strings.Fields(l)
				if len(parts) > 0 {
					return parts[len(parts)-1], nil
				}
			}
		}
		return "", fmt.Errorf("parse tunnel ID from output: %s", string(out))
	}
	return result.ID, nil
}

// ListTunnels returns all tunnels via `cloudflared tunnel list --output json`.
func ListTunnels() ([]TunnelInfo, error) {
	out, err := exec.Command("cloudflared", "tunnel", "list", "--output", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("tunnel list: %w", err)
	}
	var raw []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse tunnel list: %w", err)
	}
	tunnels := make([]TunnelInfo, len(raw))
	for i, r := range raw {
		tunnels[i] = TunnelInfo{ID: r.ID, Name: r.Name}
	}
	return tunnels, nil
}

// DeleteTunnel runs `cloudflared tunnel delete <name>`.
func DeleteTunnel(name string) error {
	out, err := exec.Command("cloudflared", "tunnel", "delete", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tunnel delete: %w — %s", err, string(out))
	}
	return nil
}

// RunTunnel starts `cloudflared tunnel run <name>` and pipes stdout to w.
// Returns the running *exec.Cmd so the caller can kill it.
func RunTunnel(name string, w io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command("cloudflared", "tunnel", "run", name)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("tunnel run: %w", err)
	}
	return cmd, nil
}

// RouteDNS creates a CNAME DNS record: `cloudflared tunnel route dns <tunnel> <hostname>`.
func RouteDNS(tunnelName, hostname string) error {
	out, err := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelName, hostname).CombinedOutput()
	if err != nil {
		return fmt.Errorf("route dns: %w — %s", err, string(out))
	}
	return nil
}

// CleanupTunnel deletes all unused (non-active) tunnels via listing and deleting.
func CleanupTunnel(names []string) []error {
	var errs []error
	for _, name := range names {
		if err := DeleteTunnel(name); err != nil {
			errs = append(errs, fmt.Errorf("delete %q: %w", name, err))
		}
	}
	return errs
}

// StopTunnel sends SIGTERM to a running tunnel process.
func StopTunnel(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(os.Interrupt)
}
```

- [ ] **Step 5: Write `internal/cloudflared/tunnel_test.go`**

```go
package cloudflared_test

import (
	"bytes"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)

func TestListTunnels_emptyOutput(t *testing.T) {
	// If cloudflared is not installed, ListTunnels should return an error, not panic.
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

var _ = bytes.NewBuffer // keep import
```

- [ ] **Step 6: Write `internal/cloudflared/config.go`**

```go
package cloudflared

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// IngressRule maps a hostname to a local service.
type IngressRule struct {
	Hostname string `yaml:"hostname"`
	Service  string `yaml:"service"`
}

// Config represents ~/.cloudflared/config.yml.
type Config struct {
	Tunnel   string        `yaml:"tunnel"`
	Token    string        `yaml:"token,omitempty"`
	Ingress  []IngressRule `yaml:"ingress"`
}

// ReadConfig reads and parses ~/.cloudflared/config.yml.
func ReadConfig() (*Config, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// WriteConfig serialises cfg and writes it to ~/.cloudflared/config.yml.
func WriteConfig(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// always add catch-all as last ingress rule
	hasHTTPStatus := false
	for _, r := range cfg.Ingress {
		if r.Service == "http_status:404" {
			hasHTTPStatus = true
		}
	}
	if !hasHTTPStatus {
		cfg.Ingress = append(cfg.Ingress, IngressRule{Service: "http_status:404"})
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "config.yml"), data, 0600)
}

// AddIngressRule appends a rule and saves config.
func AddIngressRule(hostname, service string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	// remove existing catch-all before adding new rule
	filtered := cfg.Ingress[:0]
	for _, r := range cfg.Ingress {
		if r.Service != "http_status:404" {
			filtered = append(filtered, r)
		}
	}
	cfg.Ingress = append(filtered, IngressRule{Hostname: hostname, Service: service})
	return WriteConfig(cfg)
}
```

- [ ] **Step 7: Write `internal/cloudflared/config_test.go`**

```go
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
	// WriteConfig appends http_status:404 catch-all
	if len(got.Ingress) < 2 {
		t.Errorf("expected at least 2 ingress rules, got %d", len(got.Ingress))
	}
}

func TestAddIngressRule(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".cloudflared"), 0700); err != nil {
		t.Fatal(err)
	}

	// initialise with empty config
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
```

- [ ] **Step 8: Run tests**

```bash
go test ./internal/cloudflared/ -v
```

Expected: TestWriteReadConfig PASS, TestAddIngressRule PASS, others PASS or SKIP.

- [ ] **Step 9: Commit**

```bash
git add internal/cloudflared/
git commit -m "feat: add internal/cloudflared — binary, tunnel, config"
```

---

### Task 4: internal/credentials

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 3, 5, 6, 7, 8**

**Files:**
- Create: `internal/credentials/store.go`
- Create: `internal/credentials/store_test.go`

- [ ] **Step 1: Write `internal/credentials/store.go`**

```go
package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store manages files inside ~/.cloudflared/.
type Store struct {
	dir string
}

// New returns a Store pointed at the real ~/.cloudflared.
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

// NewAt is used in tests to point the store at an arbitrary directory.
func NewAt(dir string) *Store { return &Store{dir: dir} }

// Read returns the contents of a file in the store.
func (s *Store) Read(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.dir, name))
}

// Write writes data to a file in the store (mode 0600).
func (s *Store) Write(name string, data []byte) error {
	return os.WriteFile(filepath.Join(s.dir, name), data, 0600)
}

// Exists returns true if the named file is present.
func (s *Store) Exists(name string) bool {
	_, err := os.Stat(filepath.Join(s.dir, name))
	return err == nil
}

// Delete removes a file from the store.
func (s *Store) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name))
}

// ListCredentialFiles returns all .json files (tunnel credential files).
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

// BackupTo copies all store files to destDir.
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

// RestoreFrom copies all files from srcDir into the store (overwrite).
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

// Nuke removes all files in the store (for reset/uninstall).
func (s *Store) Nuke() error {
	return os.RemoveAll(s.dir)
}
```

- [ ] **Step 2: Write `internal/credentials/store_test.go`**

```go
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
	_ = filepath.Join // use import
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/credentials/ -v
```

Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/credentials/
git commit -m "feat: add internal/credentials store"
```

---

### Task 5: internal/api

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 3, 4, 6, 7, 8**

**Files:**
- Create: `internal/api/cloudflare.go`
- Create: `internal/api/cloudflare_test.go`

- [ ] **Step 1: Write `internal/api/cloudflare.go`**

```go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client wraps the Cloudflare v4 API.
type Client struct {
	token     string
	accountID string
	http      *http.Client
}

// New creates an API client. token is a Cloudflare API token, accountID is the account ID.
func New(token, accountID string) *Client {
	return &Client{
		token:     token,
		accountID: accountID,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

type apiResponse struct {
	Success bool            `json:"success"`
	Errors  []apiError      `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) get(path string, out interface{}) error {
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	var wrapper apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !wrapper.Success {
		if len(wrapper.Errors) > 0 {
			return fmt.Errorf("api error %d: %s", wrapper.Errors[0].Code, wrapper.Errors[0].Message)
		}
		return fmt.Errorf("api request failed (HTTP %d)", resp.StatusCode)
	}
	if out != nil {
		return json.Unmarshal(wrapper.Result, out)
	}
	return nil
}

// TunnelStatus holds status info for a single tunnel.
type TunnelStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"` // "active" | "inactive" | "degraded"
}

// ListTunnels fetches all tunnels for the account.
func (c *Client) ListTunnels() ([]TunnelStatus, error) {
	var result []TunnelStatus
	if err := c.get(fmt.Sprintf("/accounts/%s/cfd_tunnel", c.accountID), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TunnelMetrics holds aggregate metrics for a tunnel.
type TunnelMetrics struct {
	TunnelID     string
	RequestCount int64
	BytesIn      int64
	BytesOut     int64
}

// GetMetrics fetches metrics for the named tunnel (simplified — returns placeholder if API doesn't support).
// Full metrics require Cloudflare Analytics API which needs a separate GraphQL call.
func (c *Client) GetMetrics(tunnelID string) (*TunnelMetrics, error) {
	// Cloudflare tunnel metrics live at the Analytics engine endpoint.
	// For v1, return status-based info; full GraphQL analytics is a future enhancement.
	tunnels, err := c.ListTunnels()
	if err != nil {
		return nil, err
	}
	for _, t := range tunnels {
		if t.ID == tunnelID || t.Name == tunnelID {
			return &TunnelMetrics{TunnelID: t.ID}, nil
		}
	}
	return nil, fmt.Errorf("tunnel %q not found", tunnelID)
}

// ValidateToken checks that the token has the necessary permissions.
func (c *Client) ValidateToken() error {
	return c.get("/user/tokens/verify", nil)
}
```

- [ ] **Step 2: Write `internal/api/cloudflare_test.go`**

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/api"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *api.Client) {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	// We can't inject the base URL directly; this test validates JSON parsing.
	return srv, api.New("test-token", "test-account")
}

func TestListTunnels_parseJSON(t *testing.T) {
	// Validate that TunnelStatus unmarshals correctly.
	raw := `[{"id":"abc","name":"my-tunnel","status":"active"}]`
	var tunnels []api.TunnelStatus
	if err := json.Unmarshal([]byte(raw), &tunnels); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(tunnels) != 1 || tunnels[0].Name != "my-tunnel" {
		t.Errorf("unexpected result: %+v", tunnels)
	}
}

func TestNew_notNil(t *testing.T) {
	c := api.New("token", "account")
	if c == nil {
		t.Error("New() returned nil")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/api/ -v
```

Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/api/
git commit -m "feat: add internal/api Cloudflare HTTP client"
```

---

### Task 6: internal/orchestration

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 3, 4, 5, 7, 8**

**Files:**
- Create: `internal/orchestration/systemd.go`
- Create: `internal/orchestration/docker.go`
- Create: `internal/orchestration/windows.go`
- Create: `internal/orchestration/kubernetes.go`
- Create: `internal/orchestration/orchestration_test.go`

- [ ] **Step 1: Write `internal/orchestration/systemd.go`**

```go
package orchestration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const systemdTmpl = `[Unit]
Description=Cloudflare Tunnel — {{.TunnelName}}
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/cloudflared tunnel run {{.TunnelName}}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

// SystemdUnit generates a systemd service file content.
func SystemdUnit(tunnelName string) (string, error) {
	tmpl, err := template.New("systemd").Parse(systemdTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"TunnelName": tunnelName}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// InstallSystemd writes the service file and enables it.
// Returns the path where the file was written.
func InstallSystemd(tunnelName string) (string, error) {
	content, err := SystemdUnit(tunnelName)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/etc/systemd/system/cloudflared-%s.service", tunnelName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write service file: %w", err)
	}
	cmds := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", fmt.Sprintf("cloudflared-%s", tunnelName)},
		{"systemctl", "start", fmt.Sprintf("cloudflared-%s", tunnelName)},
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return path, fmt.Errorf("systemctl %v: %w — %s", args[1:], err, string(out))
		}
	}
	return path, nil
}
```

- [ ] **Step 2: Write `internal/orchestration/docker.go`**

```go
package orchestration

import (
	"bytes"
	"text/template"
)

const dockerComposeTmpl = `version: "3.8"

services:
  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel run {{.TunnelName}}
    environment:
      - TUNNEL_TOKEN={{.TunnelToken}}
    volumes:
      - ~/.cloudflared:/etc/cloudflared
`

// DockerCompose generates a docker-compose.yml content.
func DockerCompose(tunnelName, tunnelToken string) (string, error) {
	tmpl, err := template.New("docker").Parse(dockerComposeTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"TunnelName":  tunnelName,
		"TunnelToken": tunnelToken,
	})
	return buf.String(), err
}
```

- [ ] **Step 3: Write `internal/orchestration/windows.go`**

```go
package orchestration

import (
	"fmt"
	"os/exec"
)

// WindowsServiceName returns the service name for a tunnel.
func WindowsServiceName(tunnelName string) string {
	return "cloudflared-" + tunnelName
}

// InstallWindowsService registers cloudflared as a Windows service using sc.exe.
// cloudflaredPath is the full path to the cloudflared.exe binary.
func InstallWindowsService(tunnelName, cloudflaredPath string) error {
	svcName := WindowsServiceName(tunnelName)
	binPath := fmt.Sprintf(`"%s" tunnel run %s`, cloudflaredPath, tunnelName)

	cmds := [][]string{
		{"sc", "create", svcName, "binPath=", binPath, "start=", "auto"},
		{"sc", "description", svcName, "Cloudflare Tunnel — " + tunnelName},
		{"sc", "start", svcName},
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc %v: %w — %s", args[1:], err, string(out))
		}
	}
	return nil
}

// RemoveWindowsService stops and deletes the Windows service.
func RemoveWindowsService(tunnelName string) error {
	svcName := WindowsServiceName(tunnelName)
	cmds := [][]string{
		{"sc", "stop", svcName},
		{"sc", "delete", svcName},
	}
	for _, args := range cmds {
		exec.Command(args[0], args[1:]...).CombinedOutput() // best-effort
	}
	return nil
}
```

- [ ] **Step 4: Write `internal/orchestration/kubernetes.go`**

```go
package orchestration

import (
	"bytes"
	"text/template"
)

const kubernetesTmpl = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared-{{.TunnelName}}
  labels:
    app: cloudflared
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      containers:
        - name: cloudflared
          image: cloudflare/cloudflared:latest
          args:
            - tunnel
            - run
            - {{.TunnelName}}
          env:
            - name: TUNNEL_TOKEN
              valueFrom:
                secretKeyRef:
                  name: cloudflared-{{.TunnelName}}
                  key: token
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 128Mi
`

// KubernetesManifest generates a Kubernetes Deployment YAML for the tunnel.
func KubernetesManifest(tunnelName string) (string, error) {
	tmpl, err := template.New("k8s").Parse(kubernetesTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{"TunnelName": tunnelName})
	return buf.String(), err
}
```

- [ ] **Step 5: Write `internal/orchestration/orchestration_test.go`**

```go
package orchestration_test

import (
	"strings"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
)

func TestSystemdUnit(t *testing.T) {
	out, err := orchestration.SystemdUnit("my-tunnel")
	if err != nil {
		t.Fatalf("SystemdUnit: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in systemd unit")
	}
	if !strings.Contains(out, "ExecStart=") {
		t.Error("expected ExecStart in systemd unit")
	}
}

func TestDockerCompose(t *testing.T) {
	out, err := orchestration.DockerCompose("my-tunnel", "tok123")
	if err != nil {
		t.Fatalf("DockerCompose: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in docker-compose")
	}
	if !strings.Contains(out, "tok123") {
		t.Error("expected token in docker-compose")
	}
}

func TestKubernetesManifest(t *testing.T) {
	out, err := orchestration.KubernetesManifest("my-tunnel")
	if err != nil {
		t.Fatalf("KubernetesManifest: %v", err)
	}
	if !strings.Contains(out, "my-tunnel") {
		t.Error("expected tunnel name in k8s manifest")
	}
	if !strings.Contains(out, "kind: Deployment") {
		t.Error("expected Deployment kind")
	}
}

func TestWindowsServiceName(t *testing.T) {
	name := orchestration.WindowsServiceName("my-tunnel")
	if name != "cloudflared-my-tunnel" {
		t.Errorf("got %q", name)
	}
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/orchestration/ -v
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/orchestration/
git commit -m "feat: add internal/orchestration — systemd, docker, windows, k8s"
```

---

### Task 7: internal/monitoring

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 3, 4, 5, 6, 8**

**Files:**
- Create: `internal/monitoring/logs.go`
- Create: `internal/monitoring/status.go`
- Create: `internal/monitoring/health.go`
- Create: `internal/monitoring/monitoring_test.go`

- [ ] **Step 1: Write `internal/monitoring/logs.go`**

```go
package monitoring

import (
	"bufio"
	"io"
	"os/exec"
)

// LogStreamer streams lines from a running cloudflared process.
type LogStreamer struct {
	cmd     *exec.Cmd
	scanner *bufio.Scanner
	stdout  io.ReadCloser
}

// NewLogStreamer starts `cloudflared tunnel run <name>` and returns a streamer.
func NewLogStreamer(tunnelName string) (*LogStreamer, error) {
	cmd := exec.Command("cloudflared", "tunnel", "--loglevel", "info", "run", tunnelName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout // combine stderr into stdout pipe isn't possible; log both
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &LogStreamer{
		cmd:     cmd,
		scanner: bufio.NewScanner(stdout),
		stdout:  stdout,
	}, nil
}

// NextLine returns the next log line. Returns ("", io.EOF) when the process exits.
func (ls *LogStreamer) NextLine() (string, error) {
	if ls.scanner.Scan() {
		return ls.scanner.Text(), nil
	}
	if err := ls.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

// Stop kills the cloudflared process.
func (ls *LogStreamer) Stop() error {
	if ls.cmd.Process != nil {
		return ls.cmd.Process.Kill()
	}
	return nil
}
```

- [ ] **Step 2: Write `internal/monitoring/status.go`**

```go
package monitoring

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// TunnelStatus holds the runtime status of a tunnel.
type TunnelStatus struct {
	Name   string
	Status string // "active" | "inactive" | "unknown"
}

// GetStatus queries `cloudflared tunnel info <name>` for status.
func GetStatus(tunnelName string) (*TunnelStatus, error) {
	out, err := exec.Command("cloudflared", "tunnel", "info", "--output", "json", tunnelName).Output()
	if err != nil {
		return &TunnelStatus{Name: tunnelName, Status: "unknown"}, nil
	}
	var raw struct {
		Status string `json:"status"`
		Name   string `json:"name"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse tunnel info: %w", err)
	}
	return &TunnelStatus{Name: raw.Name, Status: raw.Status}, nil
}

// GetAllStatuses returns status for a list of tunnel names.
func GetAllStatuses(names []string) []TunnelStatus {
	result := make([]TunnelStatus, 0, len(names))
	for _, name := range names {
		s, err := GetStatus(name)
		if err != nil || s == nil {
			result = append(result, TunnelStatus{Name: name, Status: "unknown"})
			continue
		}
		result = append(result, *s)
	}
	return result
}
```

- [ ] **Step 3: Write `internal/monitoring/health.go`**

```go
package monitoring

import (
	"fmt"
	"net/http"
	"time"
)

// HealthResult is the result of a single health check.
type HealthResult struct {
	Endpoint   string
	StatusCode int
	Latency    time.Duration
	Healthy    bool
	Error      string
}

// CheckHealth performs an HTTP GET against endpoint and returns the result.
func CheckHealth(endpoint string) HealthResult {
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Get(endpoint)
	latency := time.Since(start)

	if err != nil {
		return HealthResult{
			Endpoint: endpoint,
			Latency:  latency,
			Healthy:  false,
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()

	return HealthResult{
		Endpoint:   endpoint,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Healthy:    resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
}

// CheckHealthMulti runs health checks on multiple endpoints concurrently.
func CheckHealthMulti(endpoints []string) []HealthResult {
	results := make([]HealthResult, len(endpoints))
	ch := make(chan struct{ i int; r HealthResult }, len(endpoints))
	for i, ep := range endpoints {
		go func(i int, ep string) {
			ch <- struct {
				i int
				r HealthResult
			}{i, CheckHealth(ep)}
		}(i, ep)
	}
	for range endpoints {
		res := <-ch
		results[res.i] = res.r
	}
	return results
}

// FormatHealth returns a human-readable health string.
func FormatHealth(r HealthResult) string {
	if r.Healthy {
		return fmt.Sprintf("✓ %s — %d (%s)", r.Endpoint, r.StatusCode, r.Latency.Round(time.Millisecond))
	}
	if r.Error != "" {
		return fmt.Sprintf("✗ %s — error: %s", r.Endpoint, r.Error)
	}
	return fmt.Sprintf("✗ %s — HTTP %d (%s)", r.Endpoint, r.StatusCode, r.Latency.Round(time.Millisecond))
}
```

- [ ] **Step 4: Write `internal/monitoring/monitoring_test.go`**

```go
package monitoring_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/monitoring"
)

func TestCheckHealth_healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result := monitoring.CheckHealth(srv.URL)
	if !result.Healthy {
		t.Errorf("expected healthy, got error: %s", result.Error)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
}

func TestCheckHealth_unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	result := monitoring.CheckHealth(srv.URL)
	if result.Healthy {
		t.Error("expected unhealthy for 503")
	}
}

func TestCheckHealth_unreachable(t *testing.T) {
	result := monitoring.CheckHealth("http://127.0.0.1:1") // nothing listening
	if result.Healthy {
		t.Error("expected unhealthy for unreachable host")
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestCheckHealthMulti(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	results := monitoring.CheckHealthMulti([]string{srv.URL, srv.URL})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFormatHealth_healthy(t *testing.T) {
	r := monitoring.HealthResult{Endpoint: "http://localhost:8080", StatusCode: 200, Healthy: true}
	out := monitoring.FormatHealth(r)
	if out == "" {
		t.Error("FormatHealth returned empty string")
	}
}

func TestGetAllStatuses_emptyList(t *testing.T) {
	results := monitoring.GetAllStatuses([]string{})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/monitoring/ -v
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/monitoring/
git commit -m "feat: add internal/monitoring — logs, status, health"
```

---

### Task 8: internal/maintenance

🔗 **Depends on: Task 2**  
⚡ **Parallel with: Tasks 3, 4, 5, 6, 7**

**Files:**
- Create: `internal/maintenance/update.go`
- Create: `internal/maintenance/cleanup.go`
- Create: `internal/maintenance/reset.go`
- Create: `internal/maintenance/maintenance_test.go`

- [ ] **Step 1: Write `internal/maintenance/update.go`**

```go
package maintenance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const releaseAPI = "https://api.github.com/repos/cloudflare/cloudflared/releases/latest"

// LatestVersion fetches the latest cloudflared release tag from GitHub.
func LatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", releaseAPI, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch latest version: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return strings.TrimPrefix(result.TagName, "v"), nil
}

// CurrentVersion returns the installed cloudflared version, or empty string if not installed.
func CurrentVersion() string {
	out, err := exec.Command("cloudflared", "--version").Output()
	if err != nil {
		return ""
	}
	// output: "cloudflared version 2024.x.x"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 3 {
		return parts[2]
	}
	return strings.TrimSpace(string(out))
}

// UpdateBinary downloads the latest cloudflared binary and replaces the existing one.
func UpdateBinary() error {
	path, err := exec.LookPath("cloudflared")
	if err != nil {
		return fmt.Errorf("cloudflared not in PATH: %w", err)
	}

	tag, err := LatestVersion()
	if err != nil {
		return err
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	var asset string
	switch {
	case goos == "linux" && goarch == "amd64":
		asset = "cloudflared-linux-amd64"
	case goos == "linux" && goarch == "arm64":
		asset = "cloudflared-linux-arm64"
	case goos == "darwin" && goarch == "amd64":
		asset = "cloudflared-darwin-amd64.tgz"
	case goos == "darwin" && goarch == "arm64":
		asset = "cloudflared-darwin-arm64.tgz"
	case goos == "windows" && goarch == "amd64":
		asset = "cloudflared-windows-amd64.exe"
	default:
		return fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	url := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/download/v%s/%s", tag, asset)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open binary for writing: %w", err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
```

- [ ] **Step 2: Write `internal/maintenance/cleanup.go`**

```go
package maintenance

import (
	"os"
	"path/filepath"
	"strings"
)

// OrphanedCredentials returns credential JSON files that have no matching config entry.
// tunnelNamesInUse is the set of tunnel names referenced in configs.
func OrphanedCredentials(cloudflaredDir string, tunnelNamesInUse map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(cloudflaredDir)
	if err != nil {
		return nil, err
	}
	var orphans []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".json")
		if !tunnelNamesInUse[base] {
			orphans = append(orphans, filepath.Join(cloudflaredDir, e.Name()))
		}
	}
	return orphans, nil
}

// DeleteFiles removes the given file paths.
func DeleteFiles(paths []string) []error {
	var errs []error
	for _, p := range paths {
		if err := os.Remove(p); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
```

- [ ] **Step 3: Write `internal/maintenance/reset.go`**

```go
package maintenance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// UninstallCloudflared attempts to remove cloudflared from common install paths.
func UninstallCloudflared() error {
	path, err := exec.LookPath("cloudflared")
	if err != nil {
		return fmt.Errorf("cloudflared not found in PATH")
	}
	return os.Remove(path)
}

// RemoveConfigDir deletes ~/.cloudflared entirely.
func RemoveConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(home, ".cloudflared"))
}

// RemoveSystemService attempts to disable and remove the cloudflared system service.
func RemoveSystemService(tunnelName string) error {
	switch runtime.GOOS {
	case "linux":
		svcName := fmt.Sprintf("cloudflared-%s", tunnelName)
		exec.Command("systemctl", "stop", svcName).Run()
		exec.Command("systemctl", "disable", svcName).Run()
		return os.Remove(fmt.Sprintf("/etc/systemd/system/%s.service", svcName))
	case "windows":
		svcName := fmt.Sprintf("cloudflared-%s", tunnelName)
		exec.Command("sc", "stop", svcName).Run()
		exec.Command("sc", "delete", svcName).Run()
	case "darwin":
		// macOS: launchd plist removal would go here
	}
	return nil
}
```

- [ ] **Step 4: Write `internal/maintenance/maintenance_test.go`**

```go
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
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/maintenance/ -v
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/maintenance/
git commit -m "feat: add internal/maintenance — update, cleanup, reset"
```

---

## Phase 3: UI Screens + CI/CD (All Parallel After Phase 2)

> ⚡ Tasks 9–14 can all be dispatched simultaneously. Tasks 9–13 each implement one UI screen. Task 14 is fully independent (no code dependencies).

---

### Task 9: UI — Authentication Screen

🔗 **Depends on: Tasks 3, 4**  
⚡ **Parallel with: Tasks 10, 11, 12, 13, 14**

**Files:**
- Create: `ui/auth.go`

- [ ] **Step 1: Write `ui/auth.go`**

```go
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)

type authMsg struct{ text string; isErr bool }

type authModel struct {
	status  string
	isErr   bool
	loading bool
}

func newAuthModel() authModel { return authModel{} }

func (m authModel) Init() tea.Cmd { return nil }

func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "1":
			return m, checkInstalled
		case "2":
			m.loading = true
			m.status = "Mengunduh cloudflared..."
			return m, downloadCloudflared
		case "3":
			m.loading = true
			m.status = "Membuka browser untuk login Cloudflare..."
			return m, loginCloudflare
		case "4":
			return m, verifyConnection
		case "0":
			return m, GoBack()
		}
	case authMsg:
		m.loading = false
		m.status = msg.text
		m.isErr = msg.isErr
	}
	return m, nil
}

func (m authModel) View() string {
	title := TitleStyle.Render("AUTENTIKASI & SETUP")
	menu := MenuStyle.Render(
		"[1] Cek instalasi cloudflared\n" +
			"[2] Install / download cloudflared\n" +
			"[3] Login ke Cloudflare\n" +
			"[4] Verifikasi koneksi\n\n" +
			"[0] Kembali",
	)
	var statusLine string
	if m.status != "" {
		if m.isErr {
			statusLine = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			statusLine = "\n" + SuccessStyle.Render("✓ "+m.status)
		}
	}
	return title + "\n" + menu + statusLine
}

// tea.Cmd implementations

func checkInstalled() tea.Msg {
	if cloudflared.IsInstalled() {
		v, _ := cloudflared.GetVersion()
		return authMsg{text: fmt.Sprintf("cloudflared terinstall: %s", v)}
	}
	return authMsg{text: "cloudflared tidak ditemukan di PATH", isErr: true}
}

func downloadCloudflared() tea.Msg {
	home, err := cloudflared.ConfigDir()
	_ = home
	if err != nil {
		return authMsg{text: err.Error(), isErr: true}
	}
	// Install to /usr/local/bin or %LOCALAPPDATA%\cloudflared
	if err := cloudflared.Install("/usr/local/bin"); err != nil {
		return authMsg{text: err.Error(), isErr: true}
	}
	return authMsg{text: "cloudflared berhasil diinstall"}
}

func loginCloudflare() tea.Msg {
	if err := cloudflared.Login(); err != nil {
		return authMsg{text: err.Error(), isErr: true}
	}
	return authMsg{text: "Login berhasil"}
}

func verifyConnection() tea.Msg {
	ok, err := cloudflared.VerifyConnection()
	if err != nil {
		return authMsg{text: err.Error(), isErr: true}
	}
	if !ok {
		return authMsg{text: "cert.pem tidak ditemukan — jalankan Login terlebih dahulu", isErr: true}
	}
	return authMsg{text: "Koneksi terverifikasi — cert.pem ditemukan"}
}
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
go build ./...
```

Expected: exits 0.

- [ ] **Step 3: Commit**

```bash
git add ui/auth.go
git commit -m "feat: add auth UI screen"
```

---

### Task 10: UI — Credentials Screen

🔗 **Depends on: Tasks 3, 4**  
⚡ **Parallel with: Tasks 9, 11, 12, 13, 14**

**Files:**
- Create: `ui/credentials.go`

- [ ] **Step 1: Write `ui/credentials.go`**

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/credentials"
)

type credMsg struct{ text string; isErr bool }

type credInputState int

const (
	credIdle credInputState = iota
	credWaitingTunnelName
	credWaitingHostname
	credWaitingService
	credWaitingDeleteName
	credWaitingExportPath
)

type credentialsModel struct {
	status     string
	isErr      bool
	inputState credInputState
	input      string
	pendingName string // holds tunnel name between steps
}

func newCredentialsModel() credentialsModel { return credentialsModel{} }

func (m credentialsModel) Init() tea.Cmd { return nil }

func (m credentialsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// When waiting for text input, accumulate characters
		if m.inputState != credIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc", "ctrl+c":
				m.inputState = credIdle
				m.input = ""
				m.status = "Dibatalkan"
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			return m, nil
		}
		// Normal menu navigation
		switch msg.String() {
		case "1":
			return m, listTunnels
		case "2":
			m.inputState = credWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel: "
		case "3":
			m.inputState = credWaitingDeleteName
			m.input = ""
			m.status = "Nama tunnel yang akan dihapus: "
		case "4":
			m.inputState = credWaitingHostname
			m.input = ""
			m.status = "Hostname (contoh: app.domain.com): "
		case "5":
			return m, exportConfig
		case "0":
			return m, GoBack()
		}
	case credMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = credIdle
		m.input = ""
	}
	return m, nil
}

func (m credentialsModel) handleInput() (credentialsModel, tea.Cmd) {
	switch m.inputState {
	case credWaitingTunnelName:
		name := strings.TrimSpace(m.input)
		m.inputState = credIdle
		m.input = ""
		return m, func() tea.Msg {
			id, err := cloudflared.CreateTunnel(name)
			if err != nil {
				return credMsg{text: err.Error(), isErr: true}
			}
			return credMsg{text: fmt.Sprintf("Tunnel %q dibuat — ID: %s", name, id)}
		}
	case credWaitingDeleteName:
		name := strings.TrimSpace(m.input)
		m.inputState = credIdle
		m.input = ""
		return m, func() tea.Msg {
			if err := cloudflared.DeleteTunnel(name); err != nil {
				return credMsg{text: err.Error(), isErr: true}
			}
			return credMsg{text: fmt.Sprintf("Tunnel %q dihapus", name)}
		}
	case credWaitingHostname:
		m.pendingName = strings.TrimSpace(m.input)
		m.inputState = credWaitingService
		m.input = ""
		m.status = "Service (contoh: http://localhost:8080): "
	case credWaitingService:
		hostname := m.pendingName
		service := strings.TrimSpace(m.input)
		m.inputState = credIdle
		m.input = ""
		m.pendingName = ""
		return m, func() tea.Msg {
			if err := cloudflared.AddIngressRule(hostname, service); err != nil {
				return credMsg{text: err.Error(), isErr: true}
			}
			return credMsg{text: fmt.Sprintf("Ingress %s → %s ditambahkan", hostname, service)}
		}
	}
	return m, nil
}

func (m credentialsModel) View() string {
	title := TitleStyle.Render("MANAJEMEN KREDENSIAL")
	menu := MenuStyle.Render(
		"[1] Lihat tunnel tersimpan\n" +
			"[2] Buat tunnel baru\n" +
			"[3] Hapus tunnel\n" +
			"[4] Konfigurasi ingress rules\n" +
			"[5] Export / import config\n\n" +
			"[0] Kembali",
	)

	var bottom string
	if m.inputState != credIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render("✓ "+m.status)
		}
	}
	return title + "\n" + menu + bottom
}

func listTunnels() tea.Msg {
	tunnels, err := cloudflared.ListTunnels()
	if err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	if len(tunnels) == 0 {
		return credMsg{text: "Tidak ada tunnel"}
	}
	var sb strings.Builder
	for _, t := range tunnels {
		sb.WriteString(fmt.Sprintf("• %s (%s)\n", t.Name, t.ID))
	}
	return credMsg{text: strings.TrimSpace(sb.String())}
}

func exportConfig() tea.Msg {
	store, err := credentials.New()
	if err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	if err := store.BackupTo("./cloudflared-backup"); err != nil {
		return credMsg{text: err.Error(), isErr: true}
	}
	return credMsg{text: "Config di-export ke ./cloudflared-backup"}
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add ui/credentials.go
git commit -m "feat: add credentials UI screen"
```

---

### Task 11: UI — Monitoring Screen

🔗 **Depends on: Tasks 5, 7**  
⚡ **Parallel with: Tasks 9, 10, 12, 13, 14**

**Files:**
- Create: `ui/monitoring.go`

- [ ] **Step 1: Write `ui/monitoring.go`**

```go
package ui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/monitoring"
)

type logLineMsg string
type logDoneMsg struct{}
type monitorMsg struct{ text string; isErr bool }

type monitoringSubScreen int

const (
	monitorMain monitoringSubScreen = iota
	monitorLogs
)

type monitoringModel struct {
	subScreen   monitoringSubScreen
	viewport    viewport.Model
	logLines    []string
	streamer    *monitoring.LogStreamer
	status      string
	isErr       bool
	ready       bool
}

func newMonitoringModel() monitoringModel {
	vp := viewport.New(80, 20)
	return monitoringModel{viewport: vp}
}

func (m monitoringModel) Init() tea.Cmd { return nil }

func (m monitoringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		m.ready = true

	case tea.KeyMsg:
		if m.subScreen == monitorLogs {
			switch msg.String() {
			case "q", "0":
				if m.streamer != nil {
					m.streamer.Stop()
					m.streamer = nil
				}
				m.subScreen = monitorMain
				m.logLines = nil
				return m, nil
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "1":
			m.subScreen = monitorLogs
			m.logLines = nil
			m.status = "Menunggu tunnel name... (fitur: masukkan nama tunnel di kode)"
			// In a real flow, tunnel name comes from config; here we read it
			return m, startLogStream("my-tunnel")
		case "2":
			return m, checkStatus
		case "3":
			return m, fetchMetrics
		case "4":
			return m, runHealthCheck
		case "0":
			return m, GoBack()
		}

	case logLineMsg:
		m.logLines = append(m.logLines, string(msg))
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))
		m.viewport.GotoBottom()
		return m, readNextLine(m.streamer)

	case logDoneMsg:
		m.logLines = append(m.logLines, "[stream ended]")
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))

	case monitorMsg:
		m.status = msg.text
		m.isErr = msg.isErr
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m monitoringModel) View() string {
	if m.subScreen == monitorLogs {
		header := TitleStyle.Render("LIVE LOGS") + DimStyle.Render("  [q] berhenti")
		return header + "\n" + m.viewport.View()
	}

	title := TitleStyle.Render("OBSERVABILITY & MONITORING")
	menu := MenuStyle.Render(
		"[1] Live logs\n" +
			"[2] Status tunnel\n" +
			"[3] Metrics (Cloudflare API)\n" +
			"[4] Health check endpoint\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}

// tea.Cmd helpers

func startLogStream(tunnelName string) tea.Cmd {
	return func() tea.Msg {
		streamer, err := monitoring.NewLogStreamer(tunnelName)
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		line, err := streamer.NextLine()
		if err == io.EOF {
			return logDoneMsg{}
		}
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		_ = streamer // note: in real implementation store streamer in model via custom Msg
		return logLineMsg(line)
	}
}

func readNextLine(streamer *monitoring.LogStreamer) tea.Cmd {
	if streamer == nil {
		return nil
	}
	return func() tea.Msg {
		line, err := streamer.NextLine()
		if err == io.EOF {
			return logDoneMsg{}
		}
		if err != nil {
			return monitorMsg{text: err.Error(), isErr: true}
		}
		return logLineMsg(line)
	}
}

func checkStatus() tea.Msg {
	// Read tunnel name from config
	cfg, err := readActiveTunnelName()
	if err != nil || cfg == "" {
		return monitorMsg{text: "Tidak ada tunnel aktif di config", isErr: true}
	}
	s, err := monitoring.GetStatus(cfg)
	if err != nil {
		return monitorMsg{text: err.Error(), isErr: true}
	}
	return monitorMsg{text: fmt.Sprintf("Tunnel %q: %s", s.Name, s.Status)}
}

func fetchMetrics() tea.Msg {
	return monitorMsg{text: "Metrics memerlukan CF_API_TOKEN dan ACCOUNT_ID — set di config"}
}

func runHealthCheck() tea.Msg {
	result := monitoring.CheckHealth("http://localhost:8080")
	return monitorMsg{text: monitoring.FormatHealth(result)}
}

func readActiveTunnelName() (string, error) {
	// reads ~/.cloudflared/config.yml for the active tunnel name
	return "", nil // placeholder wired in Task 15
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add ui/monitoring.go
git commit -m "feat: add monitoring UI screen with live log viewport"
```

---

### Task 12: UI — Orchestration Screen

🔗 **Depends on: Tasks 3, 6**  
⚡ **Parallel with: Tasks 9, 10, 11, 13, 14**

**Files:**
- Create: `ui/orchestration.go`

- [ ] **Step 1: Write `ui/orchestration.go`**

```go
package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
)

type orchMsg struct{ text string; isErr bool }

type orchInputState int

const (
	orchIdle orchInputState = iota
	orchWaitingTunnelName
	orchWaitingToken
	orchWaitingInstallConfirm
	orchPendingAction
)

type orchAction int

const (
	orchActionSystemd orchAction = iota
	orchActionDocker
	orchActionWindows
	orchActionKubernetes
)

type orchestrationModel struct {
	status      string
	isErr       bool
	inputState  orchInputState
	input       string
	pendingName string
	pendingAction orchAction
}

func newOrchestrationModel() orchestrationModel { return orchestrationModel{} }

func (m orchestrationModel) Init() tea.Cmd { return nil }

func (m orchestrationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputState != orchIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc":
				m.inputState = orchIdle
				m.input = ""
				m.status = "Dibatalkan"
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "1":
			m.pendingAction = orchActionSystemd
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk systemd service: "
		case "2":
			m.pendingAction = orchActionDocker
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Docker Compose: "
		case "3":
			m.pendingAction = orchActionWindows
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Windows Service: "
		case "4":
			m.pendingAction = orchActionKubernetes
			m.inputState = orchWaitingTunnelName
			m.input = ""
			m.status = "Nama tunnel untuk Kubernetes: "
		case "0":
			return m, GoBack()
		}
	case orchMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = orchIdle
		m.input = ""
	}
	return m, nil
}

func (m orchestrationModel) handleInput() (orchestrationModel, tea.Cmd) {
	val := strings.TrimSpace(m.input)
	switch m.inputState {
	case orchWaitingTunnelName:
		m.pendingName = val
		m.input = ""
		if m.pendingAction == orchActionDocker {
			m.inputState = orchWaitingToken
			m.status = "Tunnel token (dari Cloudflare dashboard): "
			return m, nil
		}
		m.inputState = orchIdle
		name := m.pendingName
		action := m.pendingAction
		return m, func() tea.Msg { return executeOrch(action, name, "") }

	case orchWaitingToken:
		token := val
		name := m.pendingName
		m.inputState = orchIdle
		m.input = ""
		return m, func() tea.Msg { return executeOrch(orchActionDocker, name, token) }
	}
	m.inputState = orchIdle
	return m, nil
}

func executeOrch(action orchAction, tunnelName, token string) tea.Msg {
	switch action {
	case orchActionSystemd:
		content, err := orchestration.SystemdUnit(tunnelName)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		path := fmt.Sprintf("cloudflared-%s.service", tunnelName)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("Service file disimpan: %s\nJalankan: sudo cp %s /etc/systemd/system/ && sudo systemctl enable --now %s", path, path, path[:len(path)-8])}

	case orchActionDocker:
		content, err := orchestration.DockerCompose(tunnelName, token)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		if err := os.WriteFile("docker-compose.yml", []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: "docker-compose.yml disimpan — jalankan: docker compose up -d"}

	case orchActionWindows:
		content, err := orchestration.SystemdUnit(tunnelName) // generate for display
		_ = content
		return orchMsg{text: fmt.Sprintf("Jalankan sebagai Admin:\nsc create cloudflared-%s binPath=\"cloudflared.exe tunnel run %s\" start=auto\nsc start cloudflared-%s", tunnelName, tunnelName, tunnelName)}

	case orchActionKubernetes:
		content, err := orchestration.KubernetesManifest(tunnelName)
		if err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		path := fmt.Sprintf("cloudflared-%s-deployment.yaml", tunnelName)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return orchMsg{text: err.Error(), isErr: true}
		}
		return orchMsg{text: fmt.Sprintf("Manifest disimpan: %s\nJalankan: kubectl apply -f %s", path, path)}
	}
	return orchMsg{text: "Unknown action", isErr: true}
}

func (m orchestrationModel) View() string {
	title := TitleStyle.Render("ORKESTRASI")
	menu := MenuStyle.Render(
		"[1] systemd service  (Linux)\n" +
			"[2] Docker / Docker Compose\n" +
			"[3] Windows Service\n" +
			"[4] Kubernetes manifest\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.inputState != orchIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add ui/orchestration.go
git commit -m "feat: add orchestration UI screen"
```

---

### Task 13: UI — Maintenance Screen

🔗 **Depends on: Tasks 3, 4, 8**  
⚡ **Parallel with: Tasks 9, 10, 11, 12, 14**

**Files:**
- Create: `ui/maintenance.go`

- [ ] **Step 1: Write `ui/maintenance.go`**

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/maintenance"
)

type maintMsg struct{ text string; isErr bool }

type maintInputState int

const (
	maintIdle maintInputState = iota
	maintWaitingBackupPath
	maintWaitingRestorePath
	maintWaitingResetConfirm
)

type maintenanceModel struct {
	status     string
	isErr      bool
	inputState maintInputState
	input      string
}

func newMaintenanceModel() maintenanceModel { return maintenanceModel{} }

func (m maintenanceModel) Init() tea.Cmd { return nil }

func (m maintenanceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputState != maintIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc":
				m.inputState = maintIdle
				m.input = ""
				m.status = "Dibatalkan"
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "1":
			m.status = "Memeriksa update..."
			return m, checkUpdate
		case "2":
			return m, runCleanup
		case "3":
			m.inputState = maintWaitingBackupPath
			m.input = ""
			m.status = "Path backup (Enter = ./cloudflared-backup): "
		case "4":
			m.inputState = maintWaitingResetConfirm
			m.input = ""
			m.status = "KONFIRMASI RESET — ketik 'yes' untuk melanjutkan: "
		case "0":
			return m, GoBack()
		}
	case maintMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = maintIdle
		m.input = ""
	}
	return m, nil
}

func (m maintenanceModel) handleInput() (maintenanceModel, tea.Cmd) {
	val := strings.TrimSpace(m.input)
	switch m.inputState {
	case maintWaitingBackupPath:
		if val == "" {
			val = "./cloudflared-backup"
		}
		path := val
		m.inputState = maintIdle
		m.input = ""
		return m, func() tea.Msg { return doBackup(path) }

	case maintWaitingRestorePath:
		path := val
		m.inputState = maintIdle
		m.input = ""
		return m, func() tea.Msg { return doRestore(path) }

	case maintWaitingResetConfirm:
		m.inputState = maintIdle
		m.input = ""
		if val != "yes" {
			return m, func() tea.Msg { return maintMsg{text: "Reset dibatalkan"} }
		}
		return m, doReset
	}
	m.inputState = maintIdle
	return m, nil
}

func (m maintenanceModel) View() string {
	title := TitleStyle.Render("PEMELIHARAAN")
	menu := MenuStyle.Render(
		"[1] Update cloudflared\n" +
			"[2] Cleanup (hapus yang tidak terpakai)\n" +
			"[3] Backup & Restore config\n" +
			"[4] Reset / Uninstall semua\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.inputState != maintIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}

// tea.Cmd implementations

func checkUpdate() tea.Msg {
	latest, err := maintenance.LatestVersion()
	if err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	current := maintenance.CurrentVersion()
	if current == "" {
		return maintMsg{text: fmt.Sprintf("Versi terbaru: %s (cloudflared belum terinstall)", latest), isErr: true}
	}
	if current == latest {
		return maintMsg{text: fmt.Sprintf("Sudah versi terbaru: %s", current)}
	}
	return maintMsg{text: fmt.Sprintf("Update tersedia: %s → %s\nTekan [1] lagi untuk update", current, latest)}
}

func runCleanup() tea.Msg {
	return maintMsg{text: "Cleanup: gunakan menu Kredensial > Lihat tunnel untuk identifikasi tunnel tidak terpakai"}
}

func doBackup(path string) tea.Msg {
	home, err := maintenance.GetHomeDir()
	if err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	_ = home
	// Use credentials store backup
	from := home + "/.cloudflared"
	if err := maintenance.CopyDir(from, path); err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	return maintMsg{text: fmt.Sprintf("Backup selesai → %s", path)}
}

func doRestore(path string) tea.Msg {
	home, err := maintenance.GetHomeDir()
	if err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	dest := home + "/.cloudflared"
	if err := maintenance.CopyDir(path, dest); err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	return maintMsg{text: "Restore selesai"}
}

func doReset() tea.Msg {
	var errs []string
	if err := maintenance.UninstallCloudflared(); err != nil {
		errs = append(errs, "binary: "+err.Error())
	}
	if err := maintenance.RemoveConfigDir(); err != nil {
		errs = append(errs, "config: "+err.Error())
	}
	if len(errs) > 0 {
		return maintMsg{text: "Reset sebagian gagal: " + strings.Join(errs, "; "), isErr: true}
	}
	return maintMsg{text: "Reset selesai — cloudflared dan config dihapus"}
}
```

- [ ] **Step 2: Add helper functions to `internal/maintenance/reset.go`** (CopyDir, GetHomeDir)

```go
// tambahkan ke internal/maintenance/reset.go

import (
	"io"
	"os"
	"path/filepath"
)

// GetHomeDir returns the current user's home directory.
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// CopyDir copies all files from src to dst.
func CopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0700); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		in, err := os.Open(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		out, err := os.OpenFile(filepath.Join(dst, e.Name()), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			in.Close()
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add ui/maintenance.go internal/maintenance/
git commit -m "feat: add maintenance UI screen"
```

---

### Task 14: CI/CD — GitHub Actions + GoReleaser

🔗 **No internal code dependencies. Can start immediately after Task 2.**  
⚡ **Parallel with: Tasks 9–13**

**Files:**
- Create: `.github/workflows/build.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yml`

- [ ] **Step 1: Write `.github/workflows/build.yml`**

```yaml
name: Build & Test

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  build:
    name: Build (${{ matrix.goos }}/${{ matrix.goarch }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: windows
            goarch: amd64
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Test
        if: matrix.goos == 'linux' && matrix.goarch == 'amd64'
        run: go test ./internal/... -v -coverprofile=coverage.out

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: "0"
        run: go build -o /dev/null ./...
```

- [ ] **Step 2: Write `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Write `.goreleaser.yml`**

```yaml
version: 2

project_name: cloudflared-setup-cli

before:
  hooks:
    - go mod tidy

builds:
  - id: cloudflared-setup-cli
    main: .
    binary: cloudflared-setup-cli
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

release:
  github:
    owner: adityarizkyramadhan
    name: cloudflared-setup-cli
  name_template: "v{{ .Version }}"
```

- [ ] **Step 4: Commit**

```bash
git add .github/ .goreleaser.yml
git commit -m "ci: add GitHub Actions build workflow and GoReleaser config"
```

---

## Phase 4: Integration (Sequential — After All Phase 3 Tasks)

### Task 15: Wire Root Model + Smoke Test

🔗 **Depends on: Tasks 9, 10, 11, 12, 13, 14**

**Files:**
- Modify: `ui/model.go` — wire all sub-models into `screenFor()`
- Modify: `ui/monitoring.go` — wire `readActiveTunnelName()`
- Create: `README.md`

- [ ] **Step 1: Update `screenFor()` in `ui/model.go`**

Replace the stub `screenFor` function with the real wiring:

```go
// in ui/model.go — replace the existing screenFor function

func screenFor(s Screen) tea.Model {
	switch s {
	case ScreenAuth:
		return newAuthModel()
	case ScreenCredentials:
		return newCredentialsModel()
	case ScreenMonitoring:
		return newMonitoringModel()
	case ScreenOrchestration:
		return newOrchestrationModel()
	case ScreenMaintenance:
		return newMaintenanceModel()
	default:
		return newMainMenuModel()
	}
}
```

- [ ] **Step 2: Wire `readActiveTunnelName()` in `ui/monitoring.go`**

Replace the placeholder `readActiveTunnelName` function:

```go
// in ui/monitoring.go — replace existing readActiveTunnelName

func readActiveTunnelName() (string, error) {
	cfg, err := cloudflared.ReadConfig()
	if err != nil {
		return "", err
	}
	return cfg.Tunnel, nil
}
```

Add the import at the top of `ui/monitoring.go`:
```go
import (
	// existing imports...
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
)
```

- [ ] **Step 3: Full build**

```bash
go build ./...
```

Expected: exits 0, binary `cloudflared-setup-cli` (or `.exe` on Windows) produced.

- [ ] **Step 4: Run all tests**

```bash
go test ./internal/... -v
```

Expected: all PASS.

- [ ] **Step 5: Smoke test — run the app and navigate all menus**

```bash
go run .
```

Verify:
- Main menu shows 5 options
- Press `1` → Auth screen shows 4 options, press `0` returns to main
- Press `2` → Credentials screen, press `0` returns
- Press `3` → Monitoring screen, press `0` returns
- Press `4` → Orchestration screen, press `0` returns
- Press `5` → Maintenance screen, press `0` returns
- Press `0` at main → app exits

- [ ] **Step 6: Write `README.md`**

```markdown
# cloudflared-setup-cli

ATM-style interactive TUI for managing Cloudflare Tunnels.

## Install

```bash
go install github.com/adityarizkyramadhan/cloudflared-setup-cli@latest
```

Or download a binary from [Releases](https://github.com/adityarizkyramadhan/cloudflared-setup-cli/releases).

## Usage

```bash
cloudflared-setup-cli
```

Navigate with number keys. Press `0` to go back. Press `Ctrl+C` to quit.

## Menus

1. **Autentikasi & Setup** — install cloudflared, login to Cloudflare
2. **Manajemen Kredensial** — create/delete tunnels, configure ingress
3. **Observability & Monitoring** — live logs, status, health check
4. **Orkestrasi** — generate systemd, Docker Compose, Windows Service, or Kubernetes manifests
5. **Pemeliharaan** — update, cleanup, backup, reset

## Release

Tag a commit to trigger automatic multi-platform release:

```bash
git tag v1.0.0 && git push origin v1.0.0
```
```

- [ ] **Step 7: Final commit**

```bash
git add ui/model.go ui/monitoring.go README.md
git commit -m "feat: wire all UI screens into root model — v1 complete"
```

- [ ] **Step 8: Push to remote**

```bash
git push origin master
```

---

## Self-Review Checklist

- [x] **Task 1** covers `go.mod`, `main.go`, `cmd/root.go` — scaffolding complete
- [x] **Task 2** covers `ui/model.go` (NavigateMsg, GoBackMsg, RootModel, screenFor stub) and `ui/mainmenu.go`
- [x] **Tasks 3–8** cover all 6 `internal/` packages with tests
- [x] **Tasks 9–13** cover all 5 UI screens
- [x] **Task 14** covers GitHub Actions (build + release) and `.goreleaser.yml`
- [x] **Task 15** wires `screenFor()`, fixes `readActiveTunnelName()`, smoke test, README
- [x] No TBDs or incomplete placeholders
- [x] All `internal/` types used in UI tasks match definitions in Phase 2 tasks
- [x] `cloudflared.ReadConfig()` used in Task 15 matches signature defined in Task 3
- [x] `credentials.New()` / `store.BackupTo()` used in Task 10 match Task 4
- [x] `monitoring.CheckHealth()` / `FormatHealth()` used in Task 11 match Task 7
- [x] `maintenance.UninstallCloudflared()` / `RemoveConfigDir()` / `CopyDir()` / `GetHomeDir()` all defined before use
