package maintenance

import (
	"os"
	"path/filepath"
	"strings"
)

func OrphanedCredentials(cloudflaredDir string, tunnelNamesInUse map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(cloudflaredDir)
	if err != nil {
		return nil, err
	}
	var orphans []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".json")
		if !tunnelNamesInUse[base] {
			orphans = append(orphans, filepath.Join(cloudflaredDir, e.Name()))
		}
	}
	return orphans, nil
}

func DeleteFiles(paths []string) []error {
	var errs []error
	for _, p := range paths {
		if err := os.Remove(p); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
