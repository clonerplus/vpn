package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	ProxyLog struct {
		Path string `json:"path"`
	} `json:"proxy_log"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "/data/vpn-manager.db"
	}

	return cfg, nil
}
