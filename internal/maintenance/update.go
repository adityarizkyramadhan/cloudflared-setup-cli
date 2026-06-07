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
