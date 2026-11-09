package mimicmgr

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hedioum/Hedioum-Pool-Tunnel/tooling"
)

//go:embed options.json
var optionsJSON []byte

type ToolFlag struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default"`
	Help    string `json:"help"`
}

func ToolOptions() map[string][]ToolFlag {
	var result map[string][]ToolFlag
	if err := json.Unmarshal(optionsJSON, &result); err != nil {
		panic(err)
	}
	return result
}

type ToolRequest struct {
	Name    string            `json:"name"`
	Action  string            `json:"action"`
	Command string            `json:"command,omitempty"`
	Flags   map[string]string `json:"flags,omitempty"`
	Config  json.RawMessage   `json:"config,omitempty"`
	JobID   string            `json:"job_id,omitempty"`
	Confirm string            `json:"confirm,omitempty"`
}

func toolArgs(command string, values map[string]string) ([]string, error) {
	flags, ok := ToolOptions()[command]
	if !ok {
		return nil, errors.New("unknown native tool")
	}
	known := map[string]string{}
	for _, f := range flags {
		known[f.Name] = f.Type
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var args []string
	size := 0
	for _, key := range keys {
		value := values[key]
		kind, ok := known[key]
		if !ok {
			return nil, fmt.Errorf("unsupported option: %s", key)
		}
		size += len(value)
		if len(value) > 8192 || size > 32768 || strings.ContainsRune(value, 0) {
			return nil, errors.New("options too large or contain NUL")
		}
		switch kind {
		case "int":
			if _, e := strconv.Atoi(value); e != nil {
				return nil, fmt.Errorf("%s must be an integer", key)
			}
		case "bool":
			if value != "true" && value != "false" {
				return nil, fmt.Errorf("%s must be true or false", key)
			}
		}
		args = append(args, "--"+key+"="+value)
	}
	return args, nil
}
func (m *Manager) nativeBinary() string {
	return filepath.Join(filepath.Dir(m.Binary), "sober-tunnel-mimic")
}

// Prepare uses the original setup/edit logic on a private temporary config.
// The caller reviews the result before Save installs it; preview never restarts
// units, moves SSH, opens a firewall or activates routes.
func (m *Manager) Prepare(ctx context.Context, r ToolRequest) (any, error) {
	if err := CheckName(r.Name); err != nil {
		return nil, err
	}
	switch r.Command {
	case "setup-foreign", "setup-iran", "add-node", "edit-foreign", "edit-node", "remove-node", "pairing":
	default:
		return nil, errors.New("not a configuration tool")
	}
	args, err := toolArgs(r.Command, r.Flags)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err = os.MkdirAll(m.Dir, 0700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(m.Dir, ".preview-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "config.json")
	data := r.Config
	if len(data) == 0 && !strings.HasPrefix(r.Command, "setup-") {
		data, err = os.ReadFile(m.configPath(r.Name))
		if err != nil {
			return nil, err
		}
	}
	if len(data) > 0 {
		if len(data) > 64<<10 || !json.Valid(data) {
			return nil, errors.New("invalid starting configuration")
		}
		if err = os.WriteFile(path, data, 0600); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.nativeBinary(), append([]string{r.Command}, args...)...)
	cmd.Env = append(os.Environ(), "SOBER_MIMIC_CONFIG="+path, "SOBER_MIMIC_PREVIEW=1", "NO_COLOR=1", "TERM=dumb")
	var output bytes.Buffer
	w := &tooling.CappedWriter{W: &output, Remaining: 64 << 10}
	cmd.Stdout = w
	cmd.Stderr = w
	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("native setup rejected settings: %s", output.String())
	}
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if out, e := m.command(ctx, m.Binary, "--check-mimic", "--mimic-config", path); e != nil {
		return nil, fmt.Errorf("configuration validation: %s", out)
	}
	return map[string]any{"config": json.RawMessage(data), "output": output.String()}, nil
}

func (m *Manager) StartDiagnostic(r ToolRequest) (tooling.Job, error) {
	var job tooling.Job
	if err := CheckName(r.Name); err != nil {
		return job, err
	}
	switch r.Command {
	case "probe", "test", "speedtest", "check-ip":
	default:
		return job, errors.New("not a diagnostic tool")
	}
	args, err := toolArgs(r.Command, r.Flags)
	if err != nil {
		return job, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.configPath(r.Name))
	if err != nil {
		return job, err
	}
	jobs := filepath.Join(m.Dir, ".jobs")
	if err = os.MkdirAll(jobs, 0700); err != nil {
		return job, err
	}
	unlock, err := lockJobLaunch(jobs)
	if err != nil {
		return job, err
	}
	defer unlock()
	entries, _ := os.ReadDir(jobs)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			continue
		}
		if time.Since(info.ModTime()) > time.Hour {
			_ = os.RemoveAll(filepath.Join(jobs, entry.Name()))
			continue
		}
		b, e := os.ReadFile(filepath.Join(jobs, entry.Name(), "state.json"))
		var old tooling.Job
		if e == nil && json.Unmarshal(b, &old) == nil && !old.Done && time.Since(old.Started) < 6*time.Minute {
			return job, errors.New("a diagnostic is already running on this server; wait for it to finish")
		}
	}
	id := make([]byte, 16)
	if _, err = rand.Read(id); err != nil {
		return job, err
	}
	job = tooling.Job{ID: hex.EncodeToString(id), Name: r.Name, Command: r.Command, Args: args, Started: time.Now().UTC()}
	dir := filepath.Join(jobs, job.ID)
	if err = os.Mkdir(dir, 0700); err != nil {
		return job, err
	}
	if err = os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		os.RemoveAll(dir)
		return job, err
	}
	if err = tooling.WriteState(dir, job); err != nil {
		os.RemoveAll(dir)
		return job, err
	}
	cmd := exec.Command(m.nativeBinary(), "panel-job", dir)
	detachDiagnostic(cmd)
	if err = cmd.Start(); err != nil {
		os.RemoveAll(dir)
		return job, fmt.Errorf("native tools unavailable: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	job.Args = nil
	return job, nil
}
func (m *Manager) Diagnostic(name, id string) (tooling.Job, error) {
	var job tooling.Job
	if CheckName(name) != nil || len(id) != 32 {
		return job, errors.New("invalid diagnostic")
	}
	if _, err := hex.DecodeString(id); err != nil {
		return job, errors.New("invalid diagnostic")
	}
	dir := filepath.Join(m.Dir, ".jobs", id)
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		return job, err
	}
	if err = json.Unmarshal(b, &job); err != nil || job.Name != name {
		return tooling.Job{}, errors.New("diagnostic does not belong to this instance")
	}
	b, _ = os.ReadFile(filepath.Join(dir, "output.log"))
	if len(b) > 128<<10 {
		b = b[:128<<10]
	}
	job.Output = string(b)
	job.Args = nil
	if !job.Done && time.Since(job.Started) > 6*time.Minute {
		job.Done = true
		job.Error = "diagnostic worker stopped or timed out"
	}
	return job, nil
}
