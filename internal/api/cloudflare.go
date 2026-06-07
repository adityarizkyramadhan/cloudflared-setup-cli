package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	token     string
	accountID string
	http      *http.Client
}

func New(token, accountID string) *Client {
	return &Client{
		token:     token,
		accountID: accountID,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

type apiResponse struct {
	Success bool            `json:"success"`
	Errors  []apiError      `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) get(path string, out interface{}) error {
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	var wrapper apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !wrapper.Success {
		if len(wrapper.Errors) > 0 {
			return fmt.Errorf("api error %d: %s", wrapper.Errors[0].Code, wrapper.Errors[0].Message)
		}
		return fmt.Errorf("api request failed (HTTP %d)", resp.StatusCode)
	}
	if out != nil {
		return json.Unmarshal(wrapper.Result, out)
	}
	return nil
}

type TunnelStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) ListTunnels() ([]TunnelStatus, error) {
	var result []TunnelStatus
	if err := c.get(fmt.Sprintf("/accounts/%s/cfd_tunnel", c.accountID), &result); err != nil {
		return nil, err
	}
	return result, nil
}

type TunnelMetrics struct {
	TunnelID     string
	RequestCount int64
	BytesIn      int64
	BytesOut     int64
}

func (c *Client) GetMetrics(tunnelID string) (*TunnelMetrics, error) {
	tunnels, err := c.ListTunnels()
	if err != nil {
		return nil, err
	}
	for _, t := range tunnels {
		if t.ID == tunnelID || t.Name == tunnelID {
			return &TunnelMetrics{TunnelID: t.ID}, nil
		}
	}
	return nil, fmt.Errorf("tunnel %q not found", tunnelID)
}

func (c *Client) ValidateToken() error {
	return c.get("/user/tokens/verify", nil)
}
