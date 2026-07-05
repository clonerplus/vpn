package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/outi-ops/outi-vpn/internal/config"
	"github.com/outi-ops/outi-vpn/internal/vless"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "/etc/vpn-server/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := cfg.Logger()
	defer logger.Sync()

	// Build user map
	users := make(map[string]string)
	for _, u := range cfg.Users {
		users[u.UUID] = u.Name
	}

	// Start VLESS server
	if cfg.VLESS.Enabled {
		addr := fmt.Sprintf(":%d", cfg.VLESS.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Fatal("failed to start vless listener", zap.Error(err), zap.String("addr", addr))
		}
		defer listener.Close()

		server := vless.NewServer(logger, users)
		logger.Info("vless server started", zap.String("addr", addr))

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					logger.Error("accept error", zap.Error(err))
					continue
				}
				go server.Handle(conn)
			}
		}()
	}

	// Start WireGuard server
	if cfg.WireGuard.Enabled {
		logger.Info("wireguard enabled", zap.String("addr", cfg.WireGuard.Address))
		// WireGuard setup handled by kernel module + wg-quick
	}

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down")
}
