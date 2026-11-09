// Package mimicmgr manages isolated Hedioum engines beneath the sober-tunnel
// control plane. It never changes SSH, firewall rules, or the host default route.
package mimicmgr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,47}$`)

type Runner func(context.Context, string, ...string) ([]byte, error)

type Manager struct {
	Dir, UnitDir, Binary string
	Run                  Runner
	mu                   sync.Mutex
}

type Instance struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

func New(dir, units, binary string) *Manager {
	return &Manager{Dir: dir, UnitDir: units, Binary: binary, Run: func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, bin, args...).CombinedOutput()
	}}
}

func CheckName(name string) error {
	if !validName.MatchString(name) {
		return errors.New("name must be 1-48 letters, digits, underscores or hyphens, starting with a letter or digit")
	}
	return nil
}

func (m *Manager) configPath(name string) string { return filepath.Join(m.Dir, name+".json") }
func (m *Manager) unit(name string) string       { return "sober-mimic-" + name + ".service" }
func (m *Manager) command(ctx context.Context, bin string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return m.Run(ctx, bin, args...)
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".sober-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(mode); err != nil {
		f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func (m *Manager) List(ctx context.Context) ([]Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := []Instance{}
	entries, err := os.ReadDir(m.Dir)
	if os.IsNotExist(err) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		if CheckName(name) != nil {
			continue
		}
		data, err := os.ReadFile(m.configPath(name))
		if err != nil {
			return nil, err
		}
		var cfg struct {
			Role string `json:"role"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid saved instance %s", name)
		}
		out, err := m.command(ctx, "systemctl", "is-active", m.unit(name))
		items = append(items, Instance{Name: name, Role: cfg.Role, Active: err == nil && strings.TrimSpace(string(out)) == "active"})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

// Read returns credentials and must only be exposed to panel administrators.
func (m *Manager) Read(name string) (json.RawMessage, error) {
	if err := CheckName(name); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return os.ReadFile(m.configPath(name))
}

// Save validates using the actual engine, then installs config and unit together.
// A running instance must be stopped before editing, so failed edits cannot
// leave live state and persisted state disagreeing.
func (m *Manager) Save(ctx context.Context, name string, data json.RawMessage) error {
	if err := CheckName(name); err != nil {
		return err
	}
	if len(data) == 0 || len(data) > 64<<10 || !json.Valid(data) {
		return errors.New("configuration must be a JSON object under 64 KiB")
	}
	var cfg struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || (cfg.Role != "iran" && cfg.Role != "foreign") {
		return errors.New("role must be iran or foreign")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if out, err := m.command(ctx, "systemctl", "is-active", m.unit(name)); err == nil && strings.TrimSpace(string(out)) == "active" {
		return errors.New("stop this instance before editing its configuration")
	}
	if err := os.MkdirAll(m.Dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(m.Dir, ".validate-*.json")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if out, err := m.command(ctx, m.Binary, "--check-mimic", "--mimic-config", tmp); err != nil {
		// Engine validation messages contain field names, never config values.
		msg := strings.TrimSpace(string(out))
		if len(msg) > 512 {
			msg = msg[:512]
		}
		return fmt.Errorf("engine rejected configuration: %s (%w)", msg, err)
	}
	unitPath := filepath.Join(m.UnitDir, m.unit(name))
	oldConfig, cfgErr := os.ReadFile(m.configPath(name))
	oldUnit, unitErr := os.ReadFile(unitPath)
	if cfgErr != nil && !os.IsNotExist(cfgErr) {
		return cfgErr
	}
	if unitErr != nil && !os.IsNotExist(unitErr) {
		return unitErr
	}
	restore := func() {
		if cfgErr == nil {
			_ = atomicWrite(m.configPath(name), oldConfig, 0600)
		} else {
			_ = os.Remove(m.configPath(name))
		}
		if unitErr == nil {
			_ = atomicWrite(unitPath, oldUnit, 0644)
		} else {
			_ = os.Remove(unitPath)
		}
		_, _ = m.command(context.Background(), "systemctl", "daemon-reload")
	}
	if err := atomicWrite(m.configPath(name), data, 0600); err != nil {
		return err
	}
	if err := atomicWrite(unitPath, []byte(m.UnitFor(name)), 0644); err != nil {
		restore()
		return err
	}
	if _, err := m.command(ctx, "systemctl", "daemon-reload"); err != nil {
		restore()
		return fmt.Errorf("systemd reload failed: %w", err)
	}
	return nil
}

func (m *Manager) UnitFor(name string) string {
	// Paths are controlled by the installation. Quote systemd arguments and
	// escape specifier/environment expansion, independently of shell quoting.
	quote := func(s string) string {
		s = strings.ReplaceAll(s, "%", "%%")
		s = strings.ReplaceAll(s, "$", "$$")
		return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
	}
	return fmt.Sprintf("[Unit]\nDescription=sober-tunnel camouflage (%s)\nAfter=network-online.target\nWants=network-online.target\n\n[Service]\nType=simple\nExecStart=%s --mimic-config %s\nRestart=on-failure\nRestartSec=3\nTimeoutStopSec=30\nLimitNOFILE=1048576\nUMask=0077\n\n[Install]\nWantedBy=multi-user.target\n", name, quote(m.Binary), quote(m.configPath(name)))
}

func (m *Manager) Action(ctx context.Context, name, action string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := os.Stat(m.configPath(name)); err != nil {
		return err
	}
	var args []string
	switch action {
	case "start":
		args = []string{"enable", "--now", m.unit(name)}
	case "stop":
		args = []string{"disable", "--now", m.unit(name)}
	case "restart":
		args = []string{"restart", m.unit(name)}
	case "delete":
		if _, err := m.command(ctx, "systemctl", "disable", "--now", m.unit(name)); err != nil {
			return fmt.Errorf("cannot stop instance: %w", err)
		}
		if err := os.Remove(filepath.Join(m.UnitDir, m.unit(name))); err != nil && !os.IsNotExist(err) {
			return err
		}
		if _, err := m.command(ctx, "systemctl", "daemon-reload"); err != nil {
			return fmt.Errorf("systemd reload failed: %w", err)
		}
		return os.Remove(m.configPath(name))
	default:
		return errors.New("unknown action")
	}
	if _, err := m.command(ctx, "systemctl", args...); err != nil {
		return fmt.Errorf("service %s failed: %w", action, err)
	}
	return nil
}

func (m *Manager) Logs(ctx context.Context, name string) (string, error) {
	if err := CheckName(name); err != nil {
		return "", err
	}
	if _, err := os.Stat(m.configPath(name)); err != nil {
		return "", err
	}
	out, err := m.command(ctx, "journalctl", "--unit", m.unit(name), "--lines", "150", "--no-pager", "--output", "short-iso")
	return string(out), err
}
