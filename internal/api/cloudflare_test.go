package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/api"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *api.Client) {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, api.New("test-token", "test-account")
}

func TestListTunnels_parseJSON(t *testing.T) {
	raw := `[{"id":"abc","name":"my-tunnel","status":"active"}]`
	var tunnels []api.TunnelStatus
	if err := json.Unmarshal([]byte(raw), &tunnels); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(tunnels) != 1 || tunnels[0].Name != "my-tunnel" {
		t.Errorf("unexpected result: %+v", tunnels)
	}
}

func TestNew_notNil(t *testing.T) {
	c := api.New("token", "account")
	if c == nil {
		t.Error("New() returned nil")
	}
}
