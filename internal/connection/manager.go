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

	stop chan struct{}
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
		stop: make(chan struct{}),
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

	// ایجاد اتصال اولیه
	for i := 0; i < cfg.MinConnections; i++ {
		m.createConnection(address)
	}

	// شروع نگهداری Pool
	go m.maintain(cfg, address)
}

func (m *Manager) createConnection(address string) {

	conn, err := m.transport.Connect(address)

	if err != nil {
		logger.Error.Printf(
			"Connection failed: %v",
			err,
		)
		return
	}

	m.pool.Add(conn)

	m.checker.SetConnected(true)

	logger.Info.Println(
		"Connection established",
	)
}

func (m *Manager) maintain(
	cfg *config.Config,
	address string,
) {

	ticker := time.NewTicker(
		5 * time.Second,
	)

	defer ticker.Stop()

	logger.Info.Println(
		"Connection maintenance started",
	)

	for {

		select {

		case <-ticker.C:

			current := m.pool.Size()

			logger.Info.Printf(
				"Pool status: %d/%d",
				current,
				cfg.MinConnections,
			)

			if current < cfg.MinConnections {

				needed := cfg.MinConnections - current

				logger.Info.Printf(
					"Need %d more connections",
					needed,
				)

				for i := 0; i < needed; i++ {
					m.createConnection(address)
				}
			}

		case <-m.stop:

			logger.Info.Println(
				"Connection maintenance stopped",
			)

			return
		}
	}
}

func (m *Manager) Stop() error {

	logger.Info.Println(
		"Connection manager stopped",
	)

	close(m.stop)

	m.monitor.Stop()

	m.checker.SetConnected(false)

	m.pool.Close()

	return m.transport.Close()
}
