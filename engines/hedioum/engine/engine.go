// Package engine exposes Hedioum's networking core to the sober-tunnel binary.
// The standalone upstream CLI remains available for advanced setup utilities.
package engine

import (
	"github.com/hedioum/Hedioum-Pool-Tunnel/config"
	"github.com/hedioum/Hedioum-Pool-Tunnel/internal/egress"
	"github.com/hedioum/Hedioum-Pool-Tunnel/internal/ingress"
	"github.com/hedioum/Hedioum-Pool-Tunnel/internal/logging"
	"os"
	"path/filepath"
)

func Check(path string) error { _, err := config.LoadFile(path); return err }

// Run owns this process and blocks until termination, like the other engines.
func Run(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	cfg, err := config.LoadFile(abs)
	if err != nil {
		return err
	}
	if err := os.Setenv("SOBER_MIMIC_CONFIG", abs); err != nil {
		return err
	}
	logging.Init(false)
	if cfg.Role == "foreign" {
		egress.StartForeignDaemon(cfg)
	} else {
		ingress.StartIranHub(cfg)
	}
	return nil
}
