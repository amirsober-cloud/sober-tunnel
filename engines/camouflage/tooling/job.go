// Package tooling runs the original diagnostics in bounded child processes.
package tooling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Job struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Command string    `json:"command"`
	Args    []string  `json:"args,omitempty"`
	Started time.Time `json:"started"`
	Done    bool      `json:"done"`
	Error   string    `json:"error,omitempty"`
	Output  string    `json:"output,omitempty"`
}

// CappedWriter keeps diagnostics from consuming unbounded disk or memory.
type CappedWriter struct {
	W         io.Writer
	Remaining int64
}

func (w *CappedWriter) Write(p []byte) (int, error) {
	n := len(p)
	if w.Remaining > 0 {
		part := p
		if int64(len(part)) > w.Remaining {
			part = part[:w.Remaining]
		}
		written, err := w.W.Write(part)
		w.Remaining -= int64(written)
		if err != nil {
			return written, err
		}
	}
	return n, nil
}

func WriteState(dir string, job Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "state.tmp")
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "state.json"))
}

// RunJob is called only by the native helper's dedicated worker mode.
func RunJob(dir string) error {
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		return err
	}
	var job Job
	if err = json.Unmarshal(b, &job); err != nil {
		return err
	}
	switch job.Command {
	case "probe", "test", "speedtest", "check-ip":
	default:
		return fmt.Errorf("unsupported diagnostic")
	}
	f, err := os.OpenFile(filepath.Join(dir, "output.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	self, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, self, append([]string{job.Command}, job.Args...)...)
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = 5 * time.Second
	cmd.Env = append(os.Environ(), "SOBER_MIMIC_CONFIG="+filepath.Join(dir, "config.json"), "SOBER_MIMIC_PREVIEW=1", "NO_COLOR=1", "TERM=dumb")
	writer := &CappedWriter{W: f, Remaining: 128 << 10}
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err = cmd.Run(); err != nil {
		job.Error = err.Error()
	}
	if ctx.Err() != nil {
		job.Error = "diagnostic timed out"
	}
	job.Done = true
	return WriteState(dir, job)
}
