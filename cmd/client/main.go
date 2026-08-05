package main

import (
	"flag"
	"log"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
)

func main() {

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

	log.Printf("Role: %s", cfg.Role)
	log.Printf("Remote: %s:%d", cfg.RemoteHost, cfg.RemotePort)
	log.Printf("Transport: %s", cfg.Transport)
}
