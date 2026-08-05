package connection

import (
	"fmt"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
)

type Manager struct {
	transport transport.Transport
	pool      *Pool
}

func New(t transport.Transport) *Manager {
	return &Manager{
		transport: t,
		pool:      NewPool(),
	}
}

func (m *Manager) Start(cfg *config.Config) {

	address := fmt.Sprintf(
		"%s:%d",
		cfg.RemoteHost,
		cfg.RemotePort,
	)

	logger.Info.Println("Connection manager started")

	for i := 0; i < cfg.MinConnections; i++ {

		logger.Info.Printf(
			"Creating connection %d/%d",
			i+1,
			cfg.MinConnections,
		)

		conn, err := m.transport.Connect(address)

		if err != nil {
			logger.Error.Printf(
				"Connection failed: %v",
				err,
			)
			continue
		}

		m.pool.Add(conn)

		logger.Info.Println(
			"Connection established",
		)
	}
}

func (m *Manager) Stop() error {
	logger.Info.Println("Connection manager stopped")
	return m.transport.Close()
}
