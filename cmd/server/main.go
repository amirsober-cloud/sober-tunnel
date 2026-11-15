package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/connection"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/mux"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
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

	t, err := transport.NewTransport(
		cfg.Transport,
		cfg.KCP,
		cfg.QUIC,
	)

	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(
		"%s:%d",
		cfg.ListenAddr,
		cfg.ListenPort,
	)

	listener, err := t.Listen(addr)

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	logger.Info.Printf(
		"Server listening on %s via %s",
		addr,
		t.Name(),
	)

	go func() {

		connection.WaitForShutdown()

		logger.Info.Println(
			"Stopping server...",
		)

		listener.Close()

	}()

	for {

		conn, err := listener.Accept()

		if err != nil {

			logger.Info.Println(
				"Server stopped",
			)

			break
		}

		logger.Info.Printf(
			"Client connected: %s",
			conn.RemoteAddr(),
		)

		manager := mux.NewManager()

		handler := mux.NewHandler(
			manager,
		)

		go handler.Handle(conn)
	}
}
