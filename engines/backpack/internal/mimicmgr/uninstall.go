package mimicmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RemoveOwnedUnits is used only after the full uninstaller's confirmation.
// Unit filenames outside the validated camouflage namespace are untouched.
func (m *Manager) RemoveOwnedUnits(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	files, err := filepath.Glob(filepath.Join(m.UnitDir, "sober-mimic-*.service"))
	if err != nil {
		return err
	}
	if _, e := os.Lstat(filepath.Join(m.UnitDir, "sober-tunnel-mimic.service")); e == nil {
		files = append(files, filepath.Join(m.UnitDir, "sober-tunnel-mimic.service"))
	}
	for _, path := range files {
		if filepath.Base(path) == "sober-tunnel-mimic.service" {
			if _, err = m.command(ctx, "systemctl", "disable", "--now", "sober-tunnel-mimic.service"); err != nil {
				return err
			}
			if err = os.Remove(path); err != nil {
				return err
			}
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "sober-mimic-"), ".service")
		if CheckName(name) != nil {
			continue
		}
		if _, err = m.command(ctx, "systemctl", "disable", "--now", m.unit(name)); err != nil {
			return fmt.Errorf("could not stop camouflage instance %s: %w", name, err)
		}
		if err = os.Remove(path); err != nil {
			return err
		}
	}
	if len(files) > 0 {
		_, err = m.command(ctx, "systemctl", "daemon-reload")
	}
	return err
}
