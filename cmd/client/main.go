package main

import (
	"github.com/amirsober-cloud/sober-tunnel/internal/health"
	"github.com/amirsober-cloud/sober-tunnel/internal/logger"
)

func main() {

	logger.Init()

	checker := health.New()

	checker.SetConnected(true)

	logger.Info.Println("Sober Tunnel started")

	if checker.Connected() {
		logger.Info.Println("Tunnel status: CONNECTED")
	}
}
