package main

import (
	"flag"
	"log"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/connection"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
)

func main() {

	logger.Init()

	configPath := flag.String(
		"config",
		"configs/client.json",
		"config file",
	)

	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	logger.Info.Printf("Role: %s", cfg.Role)
	logger.Info.Printf("Remote: %s:%d", cfg.RemoteHost, cfg.RemotePort)
	logger.Info.Printf("Transport: %s", cfg.Transport)

	tcp := transport.NewTCP()

	manager := connection.New(tcp)

	manager.Start(cfg)

	connection.WaitForShutdown()

	manager.Stop()
}
