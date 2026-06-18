package ui

import (
	"fmt"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/credentials"
)

func (c *Console) credentialsMenu() {
	for {
		c.println("")
		c.println("=== MANAJEMEN KREDENSIAL ===")
		c.println("[1] Lihat tunnel tersimpan")
		c.println("[2] Buat tunnel baru")
		c.println("[3] Hapus tunnel")
		c.println("[4] Konfigurasi ingress rule")
		c.println("[5] Route DNS (CNAME ke tunnel)")
		c.println("[6] Export config")
		c.println("[7] Import config")
		c.println("[0] Kembali")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			c.listTunnels()
		case "2":
			c.createTunnel()
		case "3":
			c.deleteTunnel()
		case "4":
			c.configureIngress()
		case "5":
			c.routeDNSManual()
		case "6":
			c.exportConfig()
		case "7":
			c.importConfig()
		case "0":
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}

func (c *Console) listTunnels() {
	tunnels, err := cloudflared.ListTunnels()
	if err != nil {
		c.fail(err.Error())
		return
	}
	if len(tunnels) == 0 {
		c.info("Tidak ada tunnel")
		return
	}
	for _, t := range tunnels {
		c.printf("• %s (%s)\n", t.Name, t.ID)
	}
}

func (c *Console) createTunnel() {
	name := c.prompt("Nama tunnel: ")
	if name == "" {
		c.fail("nama kosong")
		return
	}
	id, err := cloudflared.CreateTunnel(name)
	if err != nil {
		c.fail(err.Error())
		return
	}
	if err := cloudflared.SetTunnel(name); err != nil {
		c.warn(fmt.Sprintf("Tunnel %q dibuat (ID: %s), tapi gagal diset aktif: %v", name, id, err))
		return
	}
	c.ok(fmt.Sprintf("Tunnel %q dibuat & diset aktif — ID: %s", name, id))
}

func (c *Console) deleteTunnel() {
	name := c.prompt("Nama tunnel yang akan dihapus: ")
	if name == "" {
		c.fail("nama kosong")
		return
	}
	if err := cloudflared.DeleteTunnelWithCleanup(name); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok(fmt.Sprintf("Tunnel %q dihapus", name))
}

func (c *Console) configureIngress() {
	hostname := c.prompt("Hostname (contoh: app.domain.com): ")
	if hostname == "" {
		c.fail("hostname kosong")
		return
	}
	service := c.prompt("Service (contoh: http://localhost:8080): ")
	if service == "" {
		c.fail("service kosong")
		return
	}
	if err := cloudflared.AddIngressRule(hostname, service); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok(fmt.Sprintf("Ingress %s → %s ditambahkan", hostname, service))

	tunnel, _ := cloudflared.ActiveTunnel()
	if tunnel == "" {
		tunnel = c.prompt("Nama tunnel untuk DNS route (Enter = lewati): ")
	}
	if tunnel == "" {
		c.warn("DNS route dilewati (tunnel tidak diketahui)")
		return
	}
	if err := cloudflared.RouteDNS(tunnel, hostname); err != nil {
		c.warn("DNS route gagal: " + err.Error())
		return
	}
	c.ok(fmt.Sprintf("DNS route dibuat: %s → %s", hostname, tunnel))
}

func (c *Console) routeDNSManual() {
	hostname := c.prompt("Hostname: ")
	if hostname == "" {
		c.fail("hostname kosong")
		return
	}
	tunnel, _ := cloudflared.ActiveTunnel()
	if tunnel == "" {
		tunnel = c.prompt("Nama tunnel: ")
	}
	if tunnel == "" {
		c.fail("nama tunnel kosong")
		return
	}
	if err := cloudflared.RouteDNS(tunnel, hostname); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok(fmt.Sprintf("DNS route dibuat: %s → %s", hostname, tunnel))
}

func (c *Console) exportConfig() {
	path := c.promptDefault("Path export (Enter = ./cloudflared-backup): ", "./cloudflared-backup")
	store, err := credentials.New()
	if err != nil {
		c.fail(err.Error())
		return
	}
	if err := store.BackupTo(path); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok("Config di-export ke " + path)
}

func (c *Console) importConfig() {
	path := c.promptDefault("Path import (Enter = ./cloudflared-backup): ", "./cloudflared-backup")
	store, err := credentials.New()
	if err != nil {
		c.fail(err.Error())
		return
	}
	if err := store.RestoreFrom(path); err != nil {
		c.fail(err.Error())
		return
	}
	c.ok("Config di-import dari " + path)
}
