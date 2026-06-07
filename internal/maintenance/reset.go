package maintenance

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func UninstallCloudflared() error {
	path, err := exec.LookPath("cloudflared")
	if err != nil {
		return fmt.Errorf("cloudflared not found in PATH")
	}
	return os.Remove(path)
}

func RemoveConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(home, ".cloudflared"))
}

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

func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

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
