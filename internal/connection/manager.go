package connection

import (
	"fmt"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
)

type Manager struct {
	transport transport.Transport
}

func New(t transport.Transport) *Manager {
	return &Manager{
		transport: t,
	}
}

func (m *Manager) Start(cfg *config.Config) {

	address := fmt.Sprintf(
		"%s:%d",
		cfg.RemoteHost,
		cfg.RemotePort,
	)

	logger.Info.Printf("Connecting to %s", address)

	_, err := m.transport.Connect(address)
	if err != nil {
		logger.Error.Printf("Connection failed: %v", err)
		return
	}

	logger.Info.Println("Connection established")
}

func (m *Manager) Stop() error {
	logger.Info.Println("Connection manager stopped")
	return m.transport.Close()
}
