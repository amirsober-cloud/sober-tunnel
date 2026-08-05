package main

import "github.com/amirsober-cloud/sober-tunnel/internal/logger"

func main() {
	logger.Init()

	logger.Info.Println("Sober Tunnel started")
	logger.Error.Println("Logger test error")
}
