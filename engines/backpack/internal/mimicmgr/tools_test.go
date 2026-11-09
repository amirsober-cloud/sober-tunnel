package mimicmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hedioum/Hedioum-Pool-Tunnel/tooling"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNativeToolArgumentsAreBoundedAndNotShellCommands(t *testing.T) {
	for _, r := range []ToolRequest{
		{Command: "setup-foreign", Flags: map[string]string{"move-ssh": "true"}},
		{Command: "uninstall"}, {Command: "probe", Flags: map[string]string{"config": "/etc/passwd"}},
		{Command: "add-node", Flags: map[string]string{"tun": "yes"}},
		{Command: "speedtest", Flags: map[string]string{"seconds": "1; reboot"}},
	} {
		if _, err := toolArgs(r.Command, r.Flags); err == nil {
			t.Fatalf("accepted unsupported request: %+v", r)
		}
	}
	args, err := toolArgs("probe", map[string]string{"node": "a; touch /tmp/should-not-exist"})
	if err != nil || len(args) != 1 || args[0] != "--node=a; touch /tmp/should-not-exist" {
		t.Fatalf("argument not preserved literally: %v %v", args, err)
	}
}

func writeHub(t *testing.T, m *Manager, port int) {
	t.Helper()
	os.MkdirAll(m.Dir, 0700)
	cfg := fmt.Sprintf(`{"role":"iran","foreign_nodes":[{"alias":"exit","auth_token":"0123456789abcdef0123456789abcdef","local_socks_port":%d,"socks_bind":"127.0.0.1","endpoints":[{"target":"203.0.113.1:8443","mimic":"tls"}],"min_connections":3,"max_connections":10,"bandwidth_limit_mbps":8,"bandwidth_jitter_mbps":2}]}`, port)
	if err := os.WriteFile(m.configPath("hub"), []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestPreflightDetectsOccupiedSocksPortAndIntegrationUsesSamePort(t *testing.T) {
	c, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	port := c.Addr().(*net.TCPAddr).Port
	m := testManager(t)
	writeHub(t, m, port)
	checks, err := m.Preflight(context.Background(), "hub")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, check := range checks {
		if strings.Contains(check.Name, "SOCKS5") {
			found = true
			if check.OK {
				t.Fatal("occupied port was accepted")
			}
		}
	}
	if !found {
		t.Fatal("missing SOCKS check")
	}
	r, err := m.Integration("hub")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(r)
	if !strings.Contains(string(b), fmt.Sprint(port)) || !strings.Contains(string(b), "127.0.0.1") || !strings.Contains(string(b), "legacy_outbound") || !strings.Contains(string(b), "modern_outbound") {
		t.Fatalf("incorrect x-ui integration: %s", b)
	}
}
func TestDiagnosticCannotReadAnotherInstanceOrTraverse(t *testing.T) {
	m := testManager(t)
	id := strings.Repeat("a", 32)
	dir := filepath.Join(m.Dir, ".jobs", id)
	os.MkdirAll(dir, 0700)
	if err := tooling.WriteState(dir, tooling.Job{ID: id, Name: "first", Started: time.Now(), Done: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Diagnostic("second", id); err == nil {
		t.Fatal("returned another instance's diagnostics")
	}
	if _, err := m.Diagnostic("first", "../../state.json"); err == nil {
		t.Fatal("accepted path traversal")
	}
	if _, err := m.Diagnostic("first", id); err != nil {
		t.Fatal(err)
	}
}
func TestHostActionRequiresExplicitConfirmationBeforeExecution(t *testing.T) {
	m := testManager(t)
	writeHub(t, m, 1081)
	if _, err := m.HostAction(context.Background(), ToolRequest{Name: "hub", Command: "bbr"}); err == nil || !strings.Contains(err.Error(), "ENABLE BBR") {
		t.Fatalf("missing confirmation guard: %v", err)
	}
}
