package main

import (
	"flag"
	"fmt"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/config"
	"os"
	"os/exec"
	"path/filepath"
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
		_, err := config.LoadFile(*path)
		return true, err
	}
	binary, err := os.Executable()
	if err != nil {
		return true, err
	}
	command := exec.Command(filepath.Join(filepath.Dir(binary), "sober-tunnel-mimic"), "--foreground", "--config", *path)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return true, command.Run()
}
