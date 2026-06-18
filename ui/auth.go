package ui

import (
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/cloudflared"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/platform"
)

func (c *Console) authMenu() {
	for {
		c.println("")
		c.println("=== AUTENTIKASI & SETUP ===")
		c.println("[1] Cek instalasi cloudflared")
		c.println("[2] Install / download cloudflared")
		c.println("[3] Login ke Cloudflare")
		c.println("[4] Verifikasi koneksi")
		c.println("[0] Kembali")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			if cloudflared.IsInstalled() {
				v, _ := cloudflared.GetVersion()
				c.ok("cloudflared terinstall: " + v)
			} else {
				c.fail("cloudflared tidak ditemukan di PATH")
			}
		case "2":
			dir, err := platform.InstallDir()
			if err != nil {
				c.fail(err.Error())
				break
			}
			c.info("Mengunduh cloudflared ke " + dir + " ...")
			if err := cloudflared.Install(dir); err != nil {
				c.fail(err.Error())
				break
			}
			c.ok("cloudflared terinstall ke " + dir)
			c.info("Pastikan folder ini ada di PATH agar perintah 'cloudflared' bisa dipanggil.")
		case "3":
			c.info("Membuka browser untuk login Cloudflare...")
			if err := cloudflared.Login(); err != nil {
				c.fail(err.Error())
				break
			}
			c.ok("Login berhasil")
		case "4":
			ok, err := cloudflared.VerifyConnection()
			if err != nil {
				c.fail(err.Error())
				break
			}
			if !ok {
				c.fail("cert.pem tidak ditemukan — jalankan Login terlebih dahulu")
			} else {
				c.ok("Koneksi terverifikasi — cert.pem ditemukan")
			}
		case "0":
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}
