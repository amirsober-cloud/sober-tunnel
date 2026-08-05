package connection

import (
	"net"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

type Watcher struct {
	pool   *Pool
	stop   chan struct{}
	period time.Duration
}

func NewWatcher(pool *Pool, period time.Duration) *Watcher {
	return &Watcher{
		pool:   pool,
		stop:   make(chan struct{}),
		period: period,
	}
}

func (w *Watcher) Start() {

	logger.Info.Println(
		"Connection watcher started",
	)

	go func() {

		ticker := time.NewTicker(w.period)
		defer ticker.Stop()

		for {

			select {

			case <-ticker.C:
				w.check()

			case <-w.stop:

				logger.Info.Println(
					"Connection watcher stopped",
				)

				return
			}
		}

	}()
}

func (w *Watcher) check() {

	connections := w.pool.Connections()

	for _, conn := range connections {

		if !isAlive(conn) {

			logger.Error.Printf(
				"Dead connection detected: %s",
				conn.RemoteAddr(),
			)

			w.pool.Remove(conn)
		}
	}
}

func (w *Watcher) Stop() {

	close(w.stop)
}

func isAlive(conn net.Conn) bool {

	err := conn.SetDeadline(
		time.Now().Add(
			2 * time.Second,
		),
	)

	if err != nil {
		return false
	}

	defer conn.SetDeadline(time.Time{})

	return true
}
