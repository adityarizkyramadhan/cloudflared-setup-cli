package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/orchestration"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/platform"
)

func (c *Console) orchestrationMenu() {
	for {
		c.println("")
		c.println("=== ORKESTRASI ===")
		c.printf("[1] Install service native (auto-detect: %s)\n", platform.ServiceManager())
		c.println("[2] Docker / Docker Compose")
		c.println("[3] Kubernetes manifest")
		c.println("[0] Kembali")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			c.installNativeService()
		case "2":
			c.generateDockerCompose()
		case "3":
			c.generateKubernetes()
		case "0":
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}

func (c *Console) installNativeService() {
	name := c.prompt("Nama tunnel: ")
	if name == "" {
		c.fail("nama kosong")
		return
	}
	switch platform.ServiceManager() {
	case "windows":
		if !platform.IsAdmin() {
			c.info("Butuh hak admin — meminta elevasi (UAC)...")
			if err := platform.RelaunchElevated(); err != nil {
				c.fail("elevasi dibatalkan: " + err.Error())
				return
			}
			c.info("Instance admin dijalankan di jendela baru — lanjutkan di sana.")
			os.Exit(0)
		}
		cfPath, err := exec.LookPath("cloudflared")
		if err != nil {
			dir, derr := platform.InstallDir()
			if derr != nil {
				c.fail("cloudflared tidak ditemukan di PATH — install dulu lewat menu Autentikasi")
				return
			}
			cfPath = filepath.Join(dir, "cloudflared.exe")
		}
		if err := orchestration.InstallWindowsService(name, cfPath); err != nil {
			c.fail(err.Error())
			return
		}
		c.ok(fmt.Sprintf("Windows Service cloudflared-%s terpasang & berjalan", name))
	case "systemd":
		path, err := orchestration.InstallSystemd(name)
		if err != nil {
			c.fail("gagal pasang systemd (perlu root? jalankan ulang dengan sudo): " + err.Error())
			return
		}
		c.ok("systemd service terpasang & aktif: " + path)
	case "launchd":
		c.fail("macOS (launchd) belum didukung — gunakan Docker [2]")
	default:
		c.fail("OS ini tidak didukung untuk service native — gunakan Docker [2]")
	}
}

func (c *Console) generateDockerCompose() {
	name := c.prompt("Nama tunnel: ")
	if name == "" {
		c.fail("nama kosong")
		return
	}
	token := c.prompt("Tunnel token (dari Cloudflare dashboard): ")
	content, err := orchestration.DockerCompose(name, token)
	if err != nil {
		c.fail(err.Error())
		return
	}
	if err := os.WriteFile("docker-compose.yml", []byte(content), 0644); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok("docker-compose.yml disimpan — jalankan: docker compose up -d")
}

func (c *Console) generateKubernetes() {
	name := c.prompt("Nama tunnel: ")
	if name == "" {
		c.fail("nama kosong")
		return
	}
	content, err := orchestration.KubernetesManifest(name)
	if err != nil {
		c.fail(err.Error())
		return
	}
	path := fmt.Sprintf("cloudflared-%s-deployment.yaml", name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok(fmt.Sprintf("Manifest disimpan: %s — jalankan: kubectl apply -f %s", path, path))
}
