package maintenance

import "testing"

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
		wantErr            bool
	}{
		{"windows", "amd64", "cloudflared-windows-amd64.exe", false},
		{"linux", "amd64", "cloudflared-linux-amd64", false},
		{"linux", "arm64", "cloudflared-linux-arm64", false},
		{"darwin", "arm64", "cloudflared-darwin-arm64.tgz", false},
		{"plan9", "amd64", "", true},
	}
	for _, c := range cases {
		got, err := assetName(c.goos, c.goarch)
		if c.wantErr {
			if err == nil {
				t.Errorf("assetName(%s,%s): expected error", c.goos, c.goarch)
			}
			continue
		}
		if err != nil {
			t.Errorf("assetName(%s,%s): unexpected error: %v", c.goos, c.goarch, err)
		}
		if got != c.want {
			t.Errorf("assetName(%s,%s) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}
