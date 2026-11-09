package mimicmgr

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	units := filepath.Join(dir, "units")
	if err := os.Mkdir(units, 0700); err != nil {
		t.Fatal(err)
	}
	m := New(filepath.Join(dir, "config"), units, "/usr/local/bin/sober-tunnel")
	m.Run = func(_ context.Context, bin string, args ...string) ([]byte, error) {
		if bin == "systemctl" && args[0] == "is-active" {
			return []byte("inactive"), errors.New("inactive")
		}
		return nil, nil
	}
	return m
}

func TestSaveRejectsBadNameWithoutCommands(t *testing.T) {
	m := testManager(t)
	m.Run = func(context.Context, string, ...string) ([]byte, error) {
		t.Fatal("command executed for invalid name")
		return nil, nil
	}
	for _, name := range []string{"../escape", "-option", "x.service/a", "a\nb", ""} {
		if err := m.Save(context.Background(), name, []byte(`{"role":"iran"}`)); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
}

func TestValidationFailurePreservesSavedConfiguration(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	old := []byte(`{"role":"foreign","auth_token":"old"}`)
	if err := m.Save(ctx, "exit", old); err != nil {
		t.Fatal(err)
	}
	unit, _ := os.ReadFile(filepath.Join(m.UnitDir, m.unit("exit")))
	previous := m.Run
	m.Run = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		if bin == m.Binary {
			return []byte("invalid endpoint"), errors.New("invalid")
		}
		return previous(ctx, bin, args...)
	}
	if err := m.Save(ctx, "exit", []byte(`{"role":"iran"}`)); err == nil {
		t.Fatal("accepted engine refusal")
	}
	got, _ := m.Read("exit")
	if string(got) != string(old) {
		t.Fatal("overwrote config before validation")
	}
	gotUnit, _ := os.ReadFile(filepath.Join(m.UnitDir, m.unit("exit")))
	if string(gotUnit) != string(unit) {
		t.Fatal("overwrote unit")
	}
}

func TestReloadFailureRollsBackConfigAndUnit(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	old := []byte(`{"role":"foreign"}`)
	if err := m.Save(ctx, "exit", old); err != nil {
		t.Fatal(err)
	}
	previous := m.Run
	failed := false
	m.Run = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		if bin == "systemctl" && args[0] == "daemon-reload" && !failed {
			failed = true
			return nil, errors.New("reload failed")
		}
		return previous(ctx, bin, args...)
	}
	if err := m.Save(ctx, "exit", []byte(`{"role":"iran"}`)); err == nil {
		t.Fatal("reload failure ignored")
	}
	got, _ := m.Read("exit")
	if string(got) != string(old) {
		t.Fatal("rollback did not restore previous config")
	}
}

func TestActiveInstanceCannotBeEdited(t *testing.T) {
	m := testManager(t)
	m.Run = func(context.Context, string, ...string) ([]byte, error) { return []byte("active"), nil }
	if err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`)); err == nil {
		t.Fatal("edited live instance")
	}
	if _, err := os.Stat(m.configPath("exit")); !os.IsNotExist(err) {
		t.Fatal("wrote live instance config")
	}
}

func TestLifecycleUsesOnlySelectedUnit(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	if err := m.Save(ctx, "exit", []byte(`{"role":"foreign"}`)); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m.Run = func(_ context.Context, bin string, args ...string) ([]byte, error) {
		calls = append(calls, bin+" "+strings.Join(args, " "))
		return nil, nil
	}
	if err := m.Action(ctx, "exit", "stop"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "systemctl disable --now sober-mimic-exit.service" {
		t.Fatalf("unexpected commands %v", calls)
	}
	if err := m.Action(ctx, "exit", "shell"); err == nil {
		t.Fatal("accepted arbitrary action")
	}
	if err := m.Action(ctx, "exit", "delete"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Read("exit"); !os.IsNotExist(err) {
		t.Fatal("config survived delete")
	}
}

func TestSingleBinaryUnitAndTemporaryFileCleanup(t *testing.T) {
	m := testManager(t)
	if err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`)); err != nil {
		t.Fatal(err)
	}
	unit, _ := os.ReadFile(filepath.Join(m.UnitDir, m.unit("exit")))
	if !strings.Contains(string(unit), `"/usr/local/bin/sober-tunnel" --mimic-config`) {
		t.Fatalf("wrong executable: %s", unit)
	}
	entries, _ := os.ReadDir(m.Dir)
	if len(entries) != 1 || entries[0].Name() != "exit.json" {
		t.Fatalf("temporary secrets left behind: %v", entries)
	}
}
