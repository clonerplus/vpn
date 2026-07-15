package main

import (
	"flag"
	"log"

	"github.com/clonerplus/vpn-manager/internal/api"
	"github.com/clonerplus/vpn-manager/internal/config"
	"github.com/clonerplus/vpn-manager/internal/db"
)

func main() {
	configPath := flag.String("config", "/etc/vpn-manager/config.json", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer store.Close()

	api.Run(cfg, store)
}
