package connection

import (
	"fmt"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/config"
	"github.com/amirsober-cloud/sober-tunnel/internal/health"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
	"github.com/amirsober-cloud/sober-tunnel/internal/transport"
)

type Manager struct {
	transport transport.Transport
	pool      *Pool

	checker *health.Checker
	monitor *health.Monitor
}

func New(t transport.Transport) *Manager {

	checker := health.New()

	return &Manager{
		transport: t,
		pool:      NewPool(),
		checker:   checker,
		monitor: health.NewMonitor(
			checker,
			10*time.Second,
		),
	}
}

func (m *Manager) Start(cfg *config.Config) {

	address := fmt.Sprintf(
		"%s:%d",
		cfg.RemoteHost,
		cfg.RemotePort,
	)

	logger.Info.Println(
		"Connection manager started",
	)

	m.monitor.Start()

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

		m.checker.SetConnected(true)

		logger.Info.Println(
			"Connection established",
		)
	}
}

func (m *Manager) Stop() error {

	logger.Info.Println(
		"Connection manager stopped",
	)

	m.monitor.Stop()

	m.checker.SetConnected(false)

	m.pool.Close()

	return m.transport.Close()
}
