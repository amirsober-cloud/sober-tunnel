package mimicmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hedioum/Hedioum-Pool-Tunnel/config"
	"github.com/hedioum/Hedioum-Pool-Tunnel/tooling"
)

// ExistingXray checks the core already supplied by x-ui on the selected Iran
// server. It never reads the panel database, installs a core or starts Xray.
func (m *Manager) ExistingXray(ctx context.Context, name string) (any, error) {
	if err := CheckName(name); err != nil {
		return nil, err
	}
	cfg, err := config.LoadFile(m.configPath(name))
	if err != nil {
		return nil, err
	}
	if cfg.Role != "iran" {
		return nil, fmt.Errorf("select the Iran hub: x-ui and its Xray run on the Iran server")
	}
	candidates := []string{"/usr/local/x-ui/bin/xray-linux-" + runtime.GOARCH, "/usr/local/x-ui/bin/xray", "/usr/local/bin/xray", "/usr/bin/xray"}
	if path, e := exec.LookPath("xray"); e == nil {
		candidates = append([]string{path}, candidates...)
	}
	var binary string
	for _, path := range candidates {
		if info, e := os.Stat(path); e == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			binary = path
			break
		}
	}
	if binary == "" {
		return map[string]any{"installed": false, "note": "Run this check on the Iran server where x-ui is installed. For a container installation, run the check in the same container/network namespace or validate the exported JSON with its Xray core."}, nil
	}
	run := func(args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, args...)
		var b bytes.Buffer
		w := &tooling.CappedWriter{W: &b, Remaining: 16 << 10}
		cmd.Stdout = w
		cmd.Stderr = w
		err := cmd.Run()
		return b.String(), err
	}
	version, err := run("version")
	if err != nil {
		return nil, fmt.Errorf("could not query the existing Xray core: %w", err)
	}
	if !strings.Contains(strings.ToLower(version), "xray") {
		return nil, fmt.Errorf("the detected executable did not identify itself as Xray")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err = os.MkdirAll(m.Dir, 0700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(m.Dir, ".xray-check-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	rows, err := m.Integration(name)
	if err != nil {
		return nil, err
	}
	var checks []Check
	for _, shape := range []string{"legacy_outbound", "modern_outbound"} {
		var outbounds []any
		for _, row := range rows.([]any) {
			outbounds = append(outbounds, row.(map[string]any)[shape])
		}
		fragment := map[string]any{"log": map[string]string{"loglevel": "warning"}, "outbounds": outbounds}
		data, _ := json.Marshal(fragment)
		path := filepath.Join(dir, shape+".json")
		if err = os.WriteFile(path, data, 0600); err != nil {
			return nil, err
		}
		output, e := run("run", "-test", "-config="+path)
		checks = append(checks, Check{Name: shape, OK: e == nil, Detail: strings.TrimSpace(output)})
	}
	return map[string]any{"installed": true, "path": binary, "version": strings.Split(strings.TrimSpace(version), "\n")[0], "checks": checks, "note": "Configuration validation only. The existing x-ui/Xray process and configuration were not changed."}, nil
}
