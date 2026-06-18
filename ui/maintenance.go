package ui

import (
	"fmt"
	"strings"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/maintenance"
)

func (c *Console) maintenanceMenu() {
	for {
		c.println("")
		c.println("=== PEMELIHARAAN ===")
		c.println("[1] Update cloudflared")
		c.println("[2] Cleanup credential orphan")
		c.println("[3] Backup config")
		c.println("[4] Reset / Uninstall semua")
		c.println("[0] Kembali")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			c.updateCloudflared()
		case "2":
			c.cleanupOrphans()
		case "3":
			c.backupConfig()
		case "4":
			c.resetAll()
		case "0":
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}

func (c *Console) updateCloudflared() {
	latest, err := maintenance.LatestVersion()
	if err != nil {
		c.fail(err.Error())
		return
	}
	current := maintenance.CurrentVersion()
	if current == "" {
		c.fail(fmt.Sprintf("cloudflared belum terinstall (versi terbaru: %s)", latest))
		return
	}
	if current == latest {
		c.ok("Sudah versi terbaru: " + current)
		return
	}
	c.info(fmt.Sprintf("Update tersedia: %s → %s", current, latest))
	if !c.confirm("Update sekarang?") {
		c.info("Dibatalkan")
		return
	}
	c.info("Mengupdate cloudflared...")
	if err := maintenance.UpdateBinary(); err != nil {
		c.fail("update gagal: " + err.Error())
		return
	}
	if v := maintenance.CurrentVersion(); v != "" {
		c.ok("cloudflared berhasil diupdate → " + v)
	} else {
		c.ok("cloudflared berhasil diupdate")
	}
}

func (c *Console) cleanupOrphans() {
	dir, err := cloudflared.ConfigDir()
	if err != nil {
		c.fail(err.Error())
		return
	}
	tunnels, err := cloudflared.ListTunnels()
	if err != nil {
		c.fail("gagal mengambil daftar tunnel: " + err.Error())
		return
	}
	inUse := make(map[string]bool, len(tunnels)*2)
	for _, t := range tunnels {
		inUse[t.Name] = true
		inUse[t.ID] = true // credential files are named by tunnel UUID
	}
	orphans, err := maintenance.OrphanedCredentials(dir, inUse)
	if err != nil {
		c.fail(err.Error())
		return
	}
	if len(orphans) == 0 {
		c.ok("Tidak ada credential orphan")
		return
	}
	c.info("Credential orphan ditemukan:")
	for _, o := range orphans {
		c.println("  • " + o)
	}
	if !c.confirm(fmt.Sprintf("Hapus %d file ini?", len(orphans))) {
		c.info("Dibatalkan")
		return
	}
	if errs := maintenance.DeleteFiles(orphans); len(errs) > 0 {
		c.fail(fmt.Sprintf("%d file gagal dihapus", len(errs)))
		return
	}
	c.ok(fmt.Sprintf("%d credential orphan dihapus", len(orphans)))
}

func (c *Console) backupConfig() {
	path := c.promptDefault("Path backup (Enter = ./cloudflared-backup): ", "./cloudflared-backup")
	from, err := cloudflared.ConfigDir()
	if err != nil {
		c.fail(err.Error())
		return
	}
	if err := maintenance.CopyDir(from, path); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok("Backup selesai → " + path)
}

func (c *Console) resetAll() {
	c.warn("Ini akan uninstall cloudflared dan menghapus folder config (~/.cloudflared)")
	if c.prompt("Ketik 'yes' untuk konfirmasi: ") != "yes" {
		c.info("Reset dibatalkan")
		return
	}
	var errs []string
	if err := maintenance.UninstallCloudflared(); err != nil {
		errs = append(errs, "binary: "+err.Error())
	}
	if err := maintenance.RemoveConfigDir(); err != nil {
		errs = append(errs, "config: "+err.Error())
	}
	if len(errs) > 0 {
		c.fail("Reset sebagian gagal: " + strings.Join(errs, "; "))
		return
	}
	c.ok("Reset selesai — cloudflared dan config dihapus")
}
