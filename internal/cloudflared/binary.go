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

func IsInstalled() bool {
	_, err := exec.LookPath("cloudflared")
	return err == nil
}

func GetVersion() (string, error) {
	out, err := exec.Command("cloudflared", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("cloudflared --version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

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

func Install(destDir string) error {
	if runtime.GOOS == "darwin" {
		return fmt.Errorf("auto-install belum didukung di macOS (rilis berupa .tgz) — install manual via brew/installer")
	}
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

func Login() error {
	cmd := exec.Command("cloudflared", "tunnel", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

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

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cloudflared"), nil
}
