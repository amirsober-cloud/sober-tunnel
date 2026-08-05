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
	watcher *Watcher

	backoff   *Backoff
	lastScale time.Time

	stop chan struct{}
}

func New(t transport.Transport) *Manager {

	checker := health.New()

	pool := NewPool()

	return &Manager{
		transport: t,
		pool:      pool,
		checker:   checker,

		monitor: health.NewMonitor(
			checker,
			10*time.Second,
		),

		watcher: NewWatcher(
			pool,
			5*time.Second,
		),

		backoff: NewBackoff(),

		lastScale: time.Now(),

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

	m.watcher.Start()

	for i := 0; i < cfg.MinConnections; i++ {
		m.createConnection(address)
	}

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

	m.backoff.Reset()

	m.checker.SetConnected(true)

	logger.Info.Printf(
		"Connection established. Pool: %d",
		m.pool.Size(),
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
				cfg.MaxConnections,
			)

			// اگر بیشتر از سقف شد کم کن
			if current > cfg.MaxConnections {

				excess := current - cfg.MaxConnections

				logger.Info.Printf(
					"Scaling down. Removing %d connections",
					excess,
				)

				for i := 0; i < excess; i++ {
					m.pool.RemoveLast()
				}

				continue
			}

			// اگر کمتر از حداقل شد جبران کن
			if current < cfg.MinConnections {

				needed := cfg.MinConnections - current

				logger.Info.Printf(
					"Minimum pool refill: %d connections",
					needed,
				)

				for i := 0; i < needed; i++ {
					m.createConnection(address)
				}

				continue
			}

			// Auto scale تا سقف
			if current < cfg.MaxConnections {

				if time.Since(m.lastScale) > 30*time.Second {

					add := 1

					logger.Info.Printf(
						"Auto scaling up +%d connection",
						add,
					)

					for i := 0; i < add; i++ {
						m.createConnection(address)
					}

					m.lastScale = time.Now()
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

	m.watcher.Stop()

	m.monitor.Stop()

	m.checker.SetConnected(false)

	m.pool.Close()

	return m.transport.Close()
}
