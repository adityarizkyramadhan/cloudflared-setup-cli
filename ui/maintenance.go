package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/maintenance"
)

type maintMsg struct {
	text  string
	isErr bool
}

type maintInputState int

const (
	maintIdle maintInputState = iota
	maintWaitingBackupPath
	maintWaitingResetConfirm
)

type maintenanceModel struct {
	status     string
	isErr      bool
	inputState maintInputState
	input      string
}

func newMaintenanceModel() maintenanceModel { return maintenanceModel{} }

func (m maintenanceModel) Init() tea.Cmd { return nil }

func (m maintenanceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputState != maintIdle {
			switch msg.String() {
			case "enter":
				return m.handleInput()
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "esc":
				m.inputState = maintIdle
				m.input = ""
				m.status = "Dibatalkan"
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "1":
			m.status = "Memeriksa update..."
			return m, checkUpdate
		case "2":
			return m, runCleanup
		case "3":
			m.inputState = maintWaitingBackupPath
			m.input = ""
			m.status = "Path backup (Enter = ./cloudflared-backup): "
		case "4":
			m.inputState = maintWaitingResetConfirm
			m.input = ""
			m.status = "KONFIRMASI RESET — ketik 'yes' untuk melanjutkan: "
		case "0":
			return m, GoBack()
		}
	case maintMsg:
		m.status = msg.text
		m.isErr = msg.isErr
		m.inputState = maintIdle
		m.input = ""
	}
	return m, nil
}

func (m maintenanceModel) handleInput() (maintenanceModel, tea.Cmd) {
	val := strings.TrimSpace(m.input)
	switch m.inputState {
	case maintWaitingBackupPath:
		if val == "" {
			val = "./cloudflared-backup"
		}
		path := val
		m.inputState = maintIdle
		m.input = ""
		return m, func() tea.Msg { return doBackup(path) }
	case maintWaitingResetConfirm:
		m.inputState = maintIdle
		m.input = ""
		if val != "yes" {
			return m, func() tea.Msg { return maintMsg{text: "Reset dibatalkan"} }
		}
		return m, doReset
	}
	m.inputState = maintIdle
	return m, nil
}

func (m maintenanceModel) View() string {
	title := TitleStyle.Render("PEMELIHARAAN")
	menu := MenuStyle.Render(
		"[1] Update cloudflared\n" +
			"[2] Cleanup (hapus yang tidak terpakai)\n" +
			"[3] Backup & Restore config\n" +
			"[4] Reset / Uninstall semua\n\n" +
			"[0] Kembali",
	)
	var bottom string
	if m.inputState != maintIdle {
		bottom = "\n" + DimStyle.Render(m.status) + m.input + "█"
	} else if m.status != "" {
		if m.isErr {
			bottom = "\n" + ErrorStyle.Render("✗ "+m.status)
		} else {
			bottom = "\n" + SuccessStyle.Render(m.status)
		}
	}
	return title + "\n" + menu + bottom
}

func checkUpdate() tea.Msg {
	latest, err := maintenance.LatestVersion()
	if err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	current := maintenance.CurrentVersion()
	if current == "" {
		return maintMsg{text: fmt.Sprintf("Versi terbaru: %s (cloudflared belum terinstall)", latest), isErr: true}
	}
	if current == latest {
		return maintMsg{text: fmt.Sprintf("Sudah versi terbaru: %s", current)}
	}
	return maintMsg{text: fmt.Sprintf("Update tersedia: %s → %s\nTekan [1] lagi untuk update", current, latest)}
}

func runCleanup() tea.Msg {
	return maintMsg{text: "Cleanup: gunakan menu Kredensial > Lihat tunnel untuk identifikasi tunnel tidak terpakai"}
}

func doBackup(path string) tea.Msg {
	home, err := maintenance.GetHomeDir()
	if err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	from := home + "/.cloudflared"
	if err := maintenance.CopyDir(from, path); err != nil {
		return maintMsg{text: err.Error(), isErr: true}
	}
	return maintMsg{text: fmt.Sprintf("Backup selesai → %s", path)}
}

func doReset() tea.Msg {
	var errs []string
	if err := maintenance.UninstallCloudflared(); err != nil {
		errs = append(errs, "binary: "+err.Error())
	}
	if err := maintenance.RemoveConfigDir(); err != nil {
		errs = append(errs, "config: "+err.Error())
	}
	if len(errs) > 0 {
		return maintMsg{text: "Reset sebagian gagal: " + strings.Join(errs, "; "), isErr: true}
	}
	return maintMsg{text: "Reset selesai — cloudflared dan config dihapus"}
}
