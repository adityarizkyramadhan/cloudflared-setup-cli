package ui

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/api"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/monitoring"
)

func (c *Console) monitoringMenu() {
	for {
		c.println("")
		c.println("=== OBSERVABILITY & MONITORING ===")
		c.println("[1] Live logs (Ctrl+C untuk berhenti)")
		c.println("[2] Status tunnel")
		c.println("[3] Metrics via Cloudflare API")
		c.println("[4] Health check endpoint")
		c.println("[0] Kembali")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			c.liveLogs()
		case "2":
			c.tunnelStatus()
		case "3":
			c.apiMetrics()
		case "4":
			c.healthCheck()
		case "0":
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}

func (c *Console) liveLogs() {
	name, err := cloudflared.ActiveTunnel()
	if err != nil || name == "" {
		c.fail("Tidak ada tunnel aktif di config — buat tunnel terlebih dahulu")
		return
	}
	streamer, err := monitoring.NewLogStreamer(name)
	if err != nil {
		c.fail(err.Error())
		return
	}
	c.info(fmt.Sprintf("Streaming logs untuk %q — tekan Ctrl+C untuk berhenti", name))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	lines := make(chan string, 16)
	done := make(chan struct{})
	go func() {
		for {
			line, err := streamer.NextLine()
			if err != nil {
				close(done)
				return
			}
			lines <- line
		}
	}()

	for {
		select {
		case <-sig:
			streamer.Stop()
			c.info("[dihentikan]")
			return
		case <-done:
			c.info("[stream berakhir]")
			return
		case line := <-lines:
			c.println(line)
		}
	}
}

func (c *Console) tunnelStatus() {
	name, err := cloudflared.ActiveTunnel()
	if err != nil || name == "" {
		c.fail("Tidak ada tunnel aktif di config")
		return
	}
	s, err := monitoring.GetStatus(name)
	if err != nil {
		c.fail(err.Error())
		return
	}
	c.ok(fmt.Sprintf("Tunnel %q: %s", s.Name, s.Status))
}

func (c *Console) apiMetrics() {
	token := c.prompt("Cloudflare API token: ")
	if token == "" {
		c.fail("token kosong")
		return
	}
	account := c.prompt("Account ID: ")
	if account == "" {
		c.fail("account ID kosong")
		return
	}
	client := api.New(token, account)
	if err := client.ValidateToken(); err != nil {
		c.fail("token tidak valid: " + err.Error())
		return
	}
	tunnels, err := client.ListTunnels()
	if err != nil {
		c.fail(err.Error())
		return
	}
	if len(tunnels) == 0 {
		c.info("Tidak ada tunnel di account ini")
		return
	}
	for _, t := range tunnels {
		c.printf("• %s (%s) — %s\n", t.Name, t.ID, t.Status)
	}
}

func (c *Console) healthCheck() {
	ep := c.promptDefault("Endpoint (Enter = http://localhost:8080): ", "http://localhost:8080")
	c.println(monitoring.FormatHealth(monitoring.CheckHealth(ep)))
}
