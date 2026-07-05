package config

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type ServerConfig struct {
	ListenAddr string       `json:"listen_addr"`
	VLESS      VLESSConfig  `json:"vless"`
	WireGuard  WGConfig     `json:"wireguard"`
	Log        LogConfig    `json:"log"`
	Users      []UserConfig `json:"users"`
}

type VLESSConfig struct {
	Enabled  bool   `json:"enabled"`
	Port     int    `json:"port"`
	UUID     string `json:"uuid"`
	Flow     string `json:"flow"`
	Security string `json:"security"`
}

type WGConfig struct {
	Enabled   bool   `json:"enabled"`
	Port      int    `json:"port"`
	Address   string `json:"address"`
	DNS       string `json:"dns"`
	PrivateKey string `json:"private_key"`
}

type LogConfig struct {
	Level  string `json:"level"`
	Output string `json:"output"`
}

type UserConfig struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func Load(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *ServerConfig) Logger() *zap.Logger {
	level := zap.InfoLevel
	if c.Log.Level == "debug" {
		level = zap.DebugLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	if c.Log.Output == "stdout" {
		cfg.OutputPaths = []string{"stdout"}
	}

	return zap.Must(cfg.Build())
}
