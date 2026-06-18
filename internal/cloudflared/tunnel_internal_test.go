package cloudflared

import (
	"errors"
	"testing"
)

func TestIsActiveConnectionsErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"code 1022", errors.New("Failed to delete tunnel: code: 1022, reason: ..."), true},
		{"active connections text", errors.New("This tunnel has active connections. Please stop all replicas"), true},
		{"unrelated", errors.New("tunnel not found"), false},
	}
	for _, c := range cases {
		if got := isActiveConnectionsErr(c.err); got != c.want {
			t.Errorf("%s: isActiveConnectionsErr = %v, want %v", c.name, got, c.want)
		}
	}
}
