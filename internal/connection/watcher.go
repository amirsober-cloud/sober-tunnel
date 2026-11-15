package connection

import (
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
	// The session's read loop detects disconnects. Setting deadlines here
	// interrupts live streams and cannot determine whether a peer is alive.
}

func (w *Watcher) Stop() {

	close(w.stop)
}
