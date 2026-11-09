package main

import (
	"fmt"
	"github.com/backpack/backpack/internal/app"
	"github.com/backpack/backpack/internal/bundle"
	"os"
	"path/filepath"
)

// Used by the source installer; never dispatched by the node RPC protocol.
func installBundleCommand(args []string) (bool, error) {
	if len(args) == 0 || args[0] != "--install-bundle" {
		return false, nil
	}
	if len(args) != 2 {
		return true, fmt.Errorf("--install-bundle requires the local build directory")
	}
	if os.Geteuid() != 0 {
		return true, fmt.Errorf("bundle installation requires root")
	}
	return true, bundle.InstallPair(filepath.Join(args[1], "sober-tunnel"), filepath.Join(args[1], "sober-tunnel-mimic"), app.BinPath, app.NativeBinPath)
}
