package connection

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

func WaitForShutdown() {

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(
		signalChannel,
		os.Interrupt,
		syscall.SIGTERM,
	)

	sig := <-signalChannel

	logger.Info.Printf("Shutdown signal received: %s", sig)
}
