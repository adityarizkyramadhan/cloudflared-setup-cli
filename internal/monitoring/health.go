package monitoring

import (
	"fmt"
	"net/http"
	"time"
)

type HealthResult struct {
	Endpoint   string
	StatusCode int
	Latency    time.Duration
	Healthy    bool
	Error      string
}

func CheckHealth(endpoint string) HealthResult {
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Get(endpoint)
	latency := time.Since(start)

	if err != nil {
		return HealthResult{
			Endpoint: endpoint,
			Latency:  latency,
			Healthy:  false,
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()

	return HealthResult{
		Endpoint:   endpoint,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Healthy:    resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
}

func CheckHealthMulti(endpoints []string) []HealthResult {
	results := make([]HealthResult, len(endpoints))
	ch := make(chan struct {
		i int
		r HealthResult
	}, len(endpoints))
	for i, ep := range endpoints {
		go func(i int, ep string) {
			ch <- struct {
				i int
				r HealthResult
			}{i, CheckHealth(ep)}
		}(i, ep)
	}
	for range endpoints {
		res := <-ch
		results[res.i] = res.r
	}
	return results
}

func FormatHealth(r HealthResult) string {
	if r.Healthy {
		return fmt.Sprintf("✓ %s — %d (%s)", r.Endpoint, r.StatusCode, r.Latency.Round(time.Millisecond))
	}
	if r.Error != "" {
		return fmt.Sprintf("✗ %s — error: %s", r.Endpoint, r.Error)
	}
	return fmt.Sprintf("✗ %s — HTTP %d (%s)", r.Endpoint, r.StatusCode, r.Latency.Round(time.Millisecond))
}
