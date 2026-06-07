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
