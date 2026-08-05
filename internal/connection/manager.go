package connection

import (
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

func (m *Manager) Start() {
	logger.Info.Println("Connection manager started")
}

func (m *Manager) Stop() error {
	logger.Info.Println("Connection manager stopped")
	return m.transport.Close()
}
