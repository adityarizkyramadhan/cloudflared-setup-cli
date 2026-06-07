package monitoring_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/internal/monitoring"
)

func TestCheckHealth_healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result := monitoring.CheckHealth(srv.URL)
	if !result.Healthy {
		t.Errorf("expected healthy, got error: %s", result.Error)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
}

func TestCheckHealth_unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	result := monitoring.CheckHealth(srv.URL)
	if result.Healthy {
		t.Error("expected unhealthy for 503")
	}
}

func TestCheckHealth_unreachable(t *testing.T) {
	result := monitoring.CheckHealth("http://127.0.0.1:1")
	if result.Healthy {
		t.Error("expected unhealthy for unreachable host")
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestCheckHealthMulti(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	results := monitoring.CheckHealthMulti([]string{srv.URL, srv.URL})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestFormatHealth_healthy(t *testing.T) {
	r := monitoring.HealthResult{Endpoint: "http://localhost:8080", StatusCode: 200, Healthy: true}
	out := monitoring.FormatHealth(r)
	if out == "" {
		t.Error("FormatHealth returned empty string")
	}
}

func TestGetAllStatuses_emptyList(t *testing.T) {
	results := monitoring.GetAllStatuses([]string{})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
