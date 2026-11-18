package mimicmgr

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewInstanceStartsButEditStaysStopped(t *testing.T) {
	m := testManager(t)
	previous := m.Run
	starts := 0
	m.Run = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		if bin == "systemctl" && args[0] == "enable" {
			if strings.Join(args, " ") != "enable --now sober-mimic-exit.service" {
				t.Fatal(args)
			}
			starts++
		}
		return previous(ctx, bin, args...)
	}
	for i := 0; i < 2; i++ {
		if err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`)); err != nil {
			t.Fatal(err)
		}
	}
	if starts != 1 {
		t.Fatalf("starts = %d, want 1", starts)
	}
}

func TestAutomaticStartFailureRetainsRecoverableConfiguration(t *testing.T) {
	m := testManager(t)
	previous := m.Run
	m.Run = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		if bin == "systemctl" && args[0] == "enable" {
			return nil, errors.New("start failed")
		}
		return previous(ctx, bin, args...)
	}
	err := m.Save(context.Background(), "exit", []byte(`{"role":"foreign"}`))
	if err == nil || !strings.Contains(err.Error(), "configuration saved") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := m.Read("exit"); err != nil {
		t.Fatal("lost recoverable configuration", err)
	}
}
