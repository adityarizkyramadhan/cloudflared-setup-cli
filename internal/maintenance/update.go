package maintenance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/platform"
)

const releaseAPI = "https://api.github.com/repos/cloudflare/cloudflared/releases/latest"

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

func CurrentVersion() string {
	out, err := exec.Command("cloudflared", "--version").Output()
	if err != nil {
		return ""
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 3 {
		return parts[2]
	}
	return strings.TrimSpace(string(out))
}

// assetName returns the cloudflared release asset for the current platform.
func assetName(goos, goarch string) (string, error) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return "cloudflared-linux-amd64", nil
	case goos == "linux" && goarch == "arm64":
		return "cloudflared-linux-arm64", nil
	case goos == "darwin" && goarch == "amd64":
		return "cloudflared-darwin-amd64.tgz", nil
	case goos == "darwin" && goarch == "arm64":
		return "cloudflared-darwin-arm64.tgz", nil
	case goos == "windows" && goarch == "amd64":
		return "cloudflared-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
}

// resolveBinaryPath finds the installed cloudflared binary: first on PATH, then
// in the platform's auto-detected install directory.
func resolveBinaryPath() (string, error) {
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p, nil
	}
	dir, err := platform.InstallDir()
	if err != nil {
		return "", fmt.Errorf("cloudflared tidak ditemukan di PATH: %w", err)
	}
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name = "cloudflared.exe"
	}
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("cloudflared tidak ditemukan di PATH maupun %s — install dulu lewat menu Autentikasi", p)
	}
	return p, nil
}

func UpdateBinary() error {
	if runtime.GOOS == "darwin" {
		return fmt.Errorf("auto-update belum didukung di macOS (rilis berupa .tgz) — update manual via brew/installer")
	}

	path, err := resolveBinaryPath()
	if err != nil {
		return err
	}

	tag, err := LatestVersion()
	if err != nil {
		return err
	}

	asset, err := assetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/download/v%s/%s", tag, asset)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	// Download to a sidecar file first so a failed/partial download never
	// corrupts the working binary.
	tmp := path + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("download: %w", err)
	}
	f.Close()

	// Swap into place. On Windows the running/locked .exe cannot be overwritten
	// directly, but it can be renamed aside, freeing the original name.
	if runtime.GOOS == "windows" {
		old := path + ".old"
		os.Remove(old)
		if err := os.Rename(path, old); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("rename current binary: %w", err)
		}
		if err := os.Rename(tmp, path); err != nil {
			os.Rename(old, path) // rollback
			return fmt.Errorf("install new binary: %w", err)
		}
		os.Remove(old) // best effort; may be locked if a service is running it
		return nil
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("install new binary: %w", err)
	}
	return nil
}
