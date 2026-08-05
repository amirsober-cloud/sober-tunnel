package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/connection"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

func main() {

	logger.Init()

	configPath := flag.String(
		"config",
		"configs/server.json",
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

	addr := fmt.Sprintf(
		"%s:%d",
		cfg.ListenAddr,
		cfg.ListenPort,
	)

	listener, err := net.Listen(
		cfg.Transport,
		addr,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	logger.Info.Printf(
		"Server listening on %s",
		addr,
	)

	go func() {
		connection.WaitForShutdown()
		listener.Close()
	}()

	for {

		conn, err := listener.Accept()

		if err != nil {
			logger.Info.Println("Server stopped")
			break
		}

		logger.Info.Printf(
			"Client connected: %s",
			conn.RemoteAddr(),
		)

		conn.Close()
	}
}
