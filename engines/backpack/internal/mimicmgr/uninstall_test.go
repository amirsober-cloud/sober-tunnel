package mimicmgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallOnlyRemovesOwnedCamouflageUnits(t *testing.T) {
	m := testManager(t)
	for _, name := range []string{"sober-mimic-hub.service", "sober-tunnel-mimic.service", "x-ui.service", "hedioum.service", "sober-mimic--invalid.service"} {
		if err := os.WriteFile(filepath.Join(m.UnitDir, name), []byte("unit"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.RemoveOwnedUnits(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(m.UnitDir, "sober-mimic-hub.service")); !os.IsNotExist(err) {
		t.Fatal("owned unit still present")
	}
	if _, err := os.Stat(filepath.Join(m.UnitDir, "sober-tunnel-mimic.service")); !os.IsNotExist(err) {
		t.Fatal("standalone native unit still present")
	}
	for _, name := range []string{"x-ui.service", "hedioum.service", "sober-mimic--invalid.service"} {
		if _, err := os.Stat(filepath.Join(m.UnitDir, name)); err != nil {
			t.Fatalf("unrelated unit %s was changed", name)
		}
	}
}
