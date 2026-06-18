//go:build windows

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// ServiceManager returns the native service manager identifier for the host.
func ServiceManager() string { return "windows" }

// IsAdmin reports whether the current process is running elevated.
func IsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// installCandidates lists Windows install directories in preference order.
func installCandidates() []string {
	var cands []string
	if p, err := exec.LookPath("cloudflared"); err == nil {
		cands = append(cands, filepath.Dir(p))
	}
	if pd := os.Getenv("ProgramData"); pd != "" {
		cands = append(cands, filepath.Join(pd, "cloudflared"))
	}
	if la := os.Getenv("LOCALAPPDATA"); la != "" {
		cands = append(cands, filepath.Join(la, "cloudflared", "bin"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		cands = append(cands, filepath.Join(home, ".cloudflared", "bin"))
	}
	return cands
}

// RelaunchElevated re-launches the current executable with admin rights via
// the UAC "runas" verb. On success the caller should exit so the elevated
// copy takes over. It returns an error if the user cancels UAC.
func RelaunchElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	cwd, err := windows.UTF16PtrFromString(filepath.Dir(exe))
	if err != nil {
		return err
	}
	var argPtr *uint16
	if len(os.Args) > 1 {
		argPtr, err = windows.UTF16PtrFromString(strings.Join(os.Args[1:], " "))
		if err != nil {
			return err
		}
	}
	return windows.ShellExecute(0, verb, file, argPtr, cwd, windows.SW_SHOWNORMAL)
}
