package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type mainMenuModel struct{}

func newMainMenuModel() mainMenuModel { return mainMenuModel{} }

func (m mainMenuModel) Init() tea.Cmd { return nil }

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			return m, NavigateTo(ScreenAuth)
		case "2":
			return m, NavigateTo(ScreenCredentials)
		case "3":
			return m, NavigateTo(ScreenMonitoring)
		case "4":
			return m, NavigateTo(ScreenOrchestration)
		case "5":
			return m, NavigateTo(ScreenMaintenance)
		case "0", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mainMenuModel) View() string {
	title := TitleStyle.Render("CLOUDFLARED SETUP CLI")
	menu := MenuStyle.Render(fmt.Sprintf(
		"[1] Autentikasi & Setup\n" +
			"[2] Manajemen Kredensial\n" +
			"[3] Observability & Monitoring\n" +
			"[4] Orkestrasi\n" +
			"[5] Pemeliharaan\n\n" +
			"[0] Keluar",
	))
	return title + "\n" + menu
}
