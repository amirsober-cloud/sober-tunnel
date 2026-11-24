package tunnelspec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/backpack/backpack/config"
	"github.com/backpack/backpack/internal/app"
	"github.com/backpack/backpack/internal/enginectl"
	"github.com/backpack/backpack/internal/manage/core"
)

// Apply writes a changed tunnel config, restarts the service and verifies
// it actually came back up. If it does not, Apply attempts to restore the
// previous configuration and service state.
func Apply(s Spec) error {
	// Helper refusals happen before any mutation; do not enter rollback and
	// replace the current file for an edit that was never written.
	if err := s.validateHelpers(); err != nil {
		return err
	}
	return applyWithUnit(s.Name, []byte(s.Render()), func() error {
		_, err := s.installService()
		return err
	})
}

// ApplyBodyWithUnit applies a rendered config and refreshes its service unit.
func ApplyBodyWithUnit(name, body string) error {
	return applyWithUnit(name, []byte(body), func() error {
		if err := core.WriteUnit(name); err != nil {
			return err
		}
		return core.DaemonReload()
	})
}

func applyWithUnit(name string, body []byte, installUnit func() error) error {
	if err := validateConfigBody(body); err != nil {
		return err
	}
	path := configPath(name)
	service := app.ServiceName(name)

	prev, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read the current config: %w", err)
	}
	wasActive := core.IsActive(service)
	unitPath := filepath.Join(app.ServiceDir, service)
	prevUnit, unitErr := os.ReadFile(unitPath)
	if unitErr != nil && !os.IsNotExist(unitErr) {
		return fmt.Errorf("could not read the current unit: %w", unitErr)
	}
	restoreUnit := func() error {
		if unitErr == nil {
			if err := app.WriteFileAtomic(unitPath, prevUnit, 0644); err != nil {
				return fmt.Errorf("restore previous unit: %w", err)
			}
		}
		// An absent unit had no service to restore. Disable before removing
		// the new unit so systemd can remove any enablement links it created.
		if os.IsNotExist(unitErr) {
			if _, err := os.Stat(unitPath); os.IsNotExist(err) {
				return nil
			}
			_, disableErr := core.Systemctl("disable", "--now", service)
			return errors.Join(disableErr, os.Remove(unitPath), core.DaemonReload())
		}
		return core.DaemonReload()
	}

	if err := app.WriteFileAtomic(path, body, app.TunnelConfigMode); err != nil {
		return err
	}
	if err := installUnit(); err != nil {
		return revertAfter(err, path, prev, service, wasActive, restoreUnit)
	}
	// Save alone won't reload an already-running unit — restart explicitly.
	if err := core.RestartService(service); err != nil {
		return revertAfter(fmt.Errorf("the tunnel failed to restart with the new settings: %w", err), path, prev, service, wasActive, restoreUnit)
	}

	// The unit can report "activating" for a moment; give it a chance, then
	// confirm it is really running.
	if !waitTunnelActive(name, body) {
		return revertAfter(startupFailure(service), path, prev, service, wasActive, restoreUnit)
	}
	// Filed only now. A change that was reverted above replaced nothing, and
	// recording it would put the configuration that is still running into the
	// list of ones to go back to. See confhist.go.
	RecordChange(name, prev, "")
	return nil
}

// applyRaw writes a configuration verbatim and restarts the tunnel on it,
// with the same revert-if-it-will-not-start guarantee Apply gives.
//
// Verbatim is the point: restoring by re-rendering a spec would silently drop
// any key the current spec cannot hold, which is the trap DirectSpec's own
// comment warns about. What was kept is what goes back.
// ApplyBody is applyRaw for a caller that has rendered the configuration
// itself: written verbatim, restarted, put back if it does not come up.
func ApplyBody(name, body, note string) error { return applyRaw(name, []byte(body), note) }

func applyRaw(name string, cfg []byte, note string) error {
	if err := validateConfigBody(cfg); err != nil {
		return err
	}
	path := configPath(name)
	service := app.ServiceName(name)

	prev, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read the current config: %w", err)
	}
	wasActive := core.IsActive(service)

	if err := app.WriteFileAtomic(path, cfg, app.TunnelConfigMode); err != nil {
		return err
	}
	if err := core.RestartService(service); err != nil {
		return revertAfter(fmt.Errorf("the tunnel failed to restart on that configuration: %w", err), path, prev, service, wasActive, nil)
	}
	if !waitTunnelActive(name, cfg) {
		return revertAfter(startupFailure(service), path, prev, service, wasActive, nil)
	}
	RecordChange(name, prev, note)
	return nil
}

func validateConfigBody(body []byte) error {
	var cfg config.Config
	if _, err := toml.Decode(string(body), &cfg); err != nil {
		return fmt.Errorf("invalid tunnel configuration: %w", err)
	}
	if cfg.Server.BindAddr == "" && cfg.Client.RemoteAddr == "" && !cfg.L3.Enabled() && !cfg.Direct.Present() {
		return errors.New("configuration describes no tunnel")
	}
	return nil
}

func startupFailure(service string) error {
	if detail := LastLogLine(service); detail != "" {
		return fmt.Errorf("the tunnel did not come up with the new settings: %s", detail)
	}
	return errors.New("the tunnel did not come up with the new settings")
}

func revertAfter(cause error, path string, prev []byte, service string, wasActive bool, restoreUnit func() error) error {
	if err := revert(path, prev, service, wasActive, restoreUnit); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback failed: %w", err))
	}
	return fmt.Errorf("%w (previous configuration restored)", cause)
}

// Type=simple becomes active before the engine starts its helper. Managed
// carriers also need an affirmative readiness response from the new process;
// neither peer connectivity nor an older engine's missing field can replace it.
func waitTunnelActive(name string, body []byte) bool {
	var cfg config.Config
	if _, err := toml.Decode(string(body), &cfg); err != nil {
		return false
	}
	managed := cfg.Server.Naive.Enabled() || cfg.Client.Naive.Enabled() || cfg.Server.Xray.Enabled() || cfg.Client.Xray.Enabled()
	timeout := 10 * time.Second
	if managed {
		// Preflight and listener readiness each have a ten-second bound.
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if !core.WaitServiceActive(app.ServiceName(name), timeout) {
		return false
	}
	return !managed || waitHelperReady(ctx, name)
}

func waitHelperReady(ctx context.Context, name string) bool {
	for ctx.Err() == nil {
		probeCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		var status enginectl.Status
		err := enginectl.AskContext(probeCtx, name, enginectl.OpStatus, &status)
		cancel()
		if err == nil && status.Name == name && status.HelperReady != nil && *status.HelperReady {
			return true
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
	}
	return false
}

func revert(path string, prev []byte, service string, wasActive bool, restoreUnit func() error) error {
	if err := app.WriteFileAtomic(path, prev, app.TunnelConfigMode); err != nil {
		return fmt.Errorf("restore previous configuration: %w", err)
	}
	if restoreUnit != nil {
		if err := restoreUnit(); err != nil {
			return err
		}
	}
	if wasActive {
		return core.RestartService(service)
	}
	if restoreUnit != nil {
		if _, err := os.Stat(filepath.Join(app.ServiceDir, service)); os.IsNotExist(err) {
			return nil
		}
	}
	return core.StopService(service)
}

// LastLogLine returns the most recent meaningful journal line for a service,
// used to explain why an edit was reverted.
func LastLogLine(service string) string {
	out, err := exec.Command("journalctl", "-u", service, "-n", "12", "--no-pager", "-o", "cat").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		low := strings.ToLower(l)
		if strings.Contains(low, "error") || strings.Contains(low, "fatal") ||
			strings.Contains(low, "failed") || strings.Contains(low, "in use") {
			return l
		}
	}
	return ""
}
