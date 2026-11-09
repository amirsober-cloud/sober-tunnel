package mimicmgr

import (
	"context"
	"github.com/backpack/backpack/internal/app"
	"os"
	"path/filepath"
	"strings"
)

var Default = New(filepath.Join(app.ConfigDir, "mimic"), app.ServiceDir, app.BinPath)

func (m *Manager) Names() []string {
	entries, _ := os.ReadDir(m.Dir)
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			name := strings.TrimSuffix(e.Name(), ".json")
			if CheckName(name) == nil {
				names = append(names, name)
			}
		}
	}
	return names
}

// Reconcile recreates units after restore, without implicitly starting routing.
func (m *Manager) Reconcile(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := m.Names()
	for _, name := range names {
		if _, err := m.command(ctx, m.Binary, "--check-mimic", "--mimic-config", m.configPath(name)); err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(m.UnitDir, m.unit(name)), []byte(m.UnitFor(name)), 0644); err != nil {
			return err
		}
	}
	if len(names) == 0 {
		return nil
	}
	_, err := m.command(ctx, "systemctl", "daemon-reload")
	return err
}

// RestartActive preserves the operator's choice to leave an instance stopped.
func (m *Manager) RestartActive(ctx context.Context) (ok, failed int) {
	for _, name := range m.Names() {
		out, err := m.command(ctx, "systemctl", "is-active", m.unit(name))
		if err != nil || strings.TrimSpace(string(out)) != "active" {
			continue
		}
		if err := m.Action(ctx, name, "restart"); err != nil {
			failed++
		} else {
			ok++
		}
	}
	return
}
