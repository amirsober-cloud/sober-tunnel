package health

import (
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

type Monitor struct {
	checker  *Checker
	interval time.Duration
	stop     chan struct{}
}

func NewMonitor(
	checker *Checker,
	interval time.Duration,
) *Monitor {

	return &Monitor{
		checker:  checker,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (m *Monitor) Start() {

	logger.Info.Println(
		"Health monitor started",
	)

	go func() {

		ticker := time.NewTicker(
			m.interval,
		)

		defer ticker.Stop()

		for {

			select {

			case <-ticker.C:

				if m.checker.Connected() {

					logger.Info.Println(
						"Health check: OK",
					)

				} else {

					logger.Error.Println(
						"Health check: disconnected",
					)
				}

			case <-m.stop:
				logger.Info.Println(
					"Health monitor stopped",
				)
				return
			}
		}

	}()
}

func (m *Monitor) Stop() {

	close(m.stop)
}
