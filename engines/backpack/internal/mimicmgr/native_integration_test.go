package mimicmgr

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hedioum/Hedioum-Pool-Tunnel/config"
)

// Uses the real compiled native tools when the Linux release artifacts exist.
func TestNativeSetupPairingAndEditingWithoutHostSideEffects(t *testing.T) {
	binary, _ := filepath.Abs("../../../../build/sober-tunnel")
	if !fileExists(binary) || !fileExists(filepath.Join(filepath.Dir(binary), "sober-tunnel-mimic")) {
		t.Skip("build both Linux artifacts first")
	}
	dir := t.TempDir()
	m := New(filepath.Join(dir, "config"), filepath.Join(dir, "units"), binary)
	m.Run = func(ctx context.Context, bin string, args ...string) ([]byte, error) {
		if bin == "systemctl" {
			t.Fatal("preview executed systemctl")
			return nil, errors.New("unexpected host operation")
		}
		return exec.CommandContext(ctx, bin, args...).CombinedOutput()
	}
	stub := filepath.Join(dir, "bin")
	os.Mkdir(stub, 0700)
	calls := filepath.Join(dir, "systemctl-calls")
	os.WriteFile(filepath.Join(stub, "systemctl"), []byte("#!/bin/sh\necho called >> \"$PREVIEW_TEST_CALLS\"\nexit 1\n"), 0755)
	t.Setenv("PATH", stub+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PREVIEW_TEST_CALLS", calls)
	ctx := context.Background()
	result, err := m.Prepare(ctx, ToolRequest{Name: "exit", Command: "setup-foreign", Flags: map[string]string{"mimics": "tls", "tls-port": "18443", "http-decoy-port": "0", "public-ip": "203.0.113.10"}})
	if err != nil {
		t.Fatal(err)
	}
	exitData := result.(map[string]any)["config"].(json.RawMessage)
	if _, err = os.Stat(m.configPath("exit")); !os.IsNotExist(err) {
		t.Fatal("preview wrote the live config")
	}
	if err = os.WriteFile(m.configPath("exit"), exitData, 0600); err != nil {
		t.Fatal(err)
	}
	result, err = m.Prepare(ctx, ToolRequest{Name: "exit", Command: "pairing", Flags: map[string]string{"public-ip": "203.0.113.10"}})
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimSpace(result.(map[string]any)["output"].(string))
	result, err = m.Prepare(ctx, ToolRequest{Name: "hub", Command: "setup-iran", Flags: map[string]string{"alias": "foreign", "socks-port": "11081", "token": token}})
	if err != nil {
		t.Fatal(err)
	}
	hubData := result.(map[string]any)["config"].(json.RawMessage)
	var hub, exit config.AppConfig
	json.Unmarshal(hubData, &hub)
	json.Unmarshal(exitData, &exit)
	if hub.ForeignNodes[0].AuthToken != exit.AuthToken || hub.ForeignNodes[0].Endpoints[0].Target != "203.0.113.10:18443" || hub.ForeignNodes[0].SocksBind != "127.0.0.1" {
		t.Fatal("pairing did not preserve token, endpoint or loopback SOCKS binding")
	}
	if err = os.WriteFile(m.configPath("hub"), hubData, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Prepare(ctx, ToolRequest{Name: "hub", Command: "edit-node", Flags: map[string]string{"alias": "foreign", "min": "10", "max": "2"}}); err == nil {
		t.Fatal("accepted an invalid pool")
	}
	unchanged, _ := os.ReadFile(m.configPath("hub"))
	if string(unchanged) != string(hubData) {
		t.Fatal("failed edit changed live configuration")
	}
	result, err = m.Prepare(ctx, ToolRequest{Name: "exit", Command: "edit-foreign", Flags: map[string]string{"rotate-token": "true"}})
	if err != nil {
		t.Fatal(err)
	}
	var rotated config.AppConfig
	json.Unmarshal(result.(map[string]any)["config"].(json.RawMessage), &rotated)
	if rotated.AuthToken == exit.AuthToken {
		t.Fatal("rotation failed")
	}
	if _, err = os.Stat(calls); !os.IsNotExist(err) {
		t.Fatal("native preview invoked systemctl")
	}
}
