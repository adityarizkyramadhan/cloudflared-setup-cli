package cloudflared

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type IngressRule struct {
	Hostname string `yaml:"hostname"`
	Service  string `yaml:"service"`
}

type Config struct {
	Tunnel  string        `yaml:"tunnel"`
	Token   string        `yaml:"token,omitempty"`
	Ingress []IngressRule `yaml:"ingress"`
}

func ReadConfig() (*Config, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func WriteConfig(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	hasHTTPStatus := false
	for _, r := range cfg.Ingress {
		if r.Service == "http_status:404" {
			hasHTTPStatus = true
		}
	}
	if !hasHTTPStatus {
		cfg.Ingress = append(cfg.Ingress, IngressRule{Service: "http_status:404"})
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "config.yml"), data, 0600)
}

// ActiveTunnel returns the tunnel name recorded in config.yml, or "" if none
// is set. A missing config file is not an error.
func ActiveTunnel() (string, error) {
	cfg, err := ReadConfig()
	if err != nil {
		return "", err
	}
	return cfg.Tunnel, nil
}

// SetTunnel records name as the active tunnel in config.yml, preserving any
// existing ingress rules.
func SetTunnel(name string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	cfg.Tunnel = name
	return WriteConfig(cfg)
}

func AddIngressRule(hostname, service string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	filtered := cfg.Ingress[:0]
	for _, r := range cfg.Ingress {
		if r.Service != "http_status:404" {
			filtered = append(filtered, r)
		}
	}
	cfg.Ingress = append(filtered, IngressRule{Hostname: hostname, Service: service})
	return WriteConfig(cfg)
}
