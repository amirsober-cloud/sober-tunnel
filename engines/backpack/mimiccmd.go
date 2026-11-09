package main

import (
	"flag"
	"fmt"
	"github.com/hedioum/Hedioum-Pool-Tunnel/engine"
	"os"
	"strings"
)

// A separate engine mode of the same binary; never launches a second panel.
func mimicCommand(args []string) (bool, error) {
	if len(args) == 0 || (args[0] != "--mimic-config" && args[0] != "--check-mimic" && !strings.HasPrefix(args[0], "--mimic-config=")) {
		return false, nil
	}
	fs := flag.NewFlagSet("sober-tunnel camouflage", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	path := fs.String("mimic-config", "", "Camouflage instance JSON configuration")
	check := fs.Bool("check-mimic", false, "Validate configuration without starting a tunnel")
	if err := fs.Parse(args); err != nil {
		return true, err
	}
	if *path == "" || len(fs.Args()) != 0 {
		return true, fmt.Errorf("--mimic-config <file> is required")
	}
	if *check {
		return true, engine.Check(*path)
	}
	return true, engine.Run(*path)
}
